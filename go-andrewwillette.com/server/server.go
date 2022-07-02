package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/andrewwillette/keyOfDay/key"
	"github.com/andrewwillette/willette_api/config"
	"github.com/andrewwillette/willette_api/logging"
	"github.com/andrewwillette/willette_api/persistence"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/newrelic/go-agent/v3/integrations/nrecho-v4"
	"github.com/newrelic/go-agent/v3/newrelic"
)

const (
	getSoundcloudAllEndpoint     = "/get-soundcloud-urls"
	addSoundcloudEndpoint        = "/add-soundcloud-url"
	deleteSoundcloudEndpoint     = "/delete-soundcloud-url"
	keyOfDayEndpoint             = "/keyOfDay"
	loginEndpoint                = "/login"
	updateSoundcloudUrlsEndpoint = "/update-soundcloud-urls"
)

func StartServer() {
	databaseFile := config.GetDatabaseFile()
	persistence.InitDatabaseIdempotent(databaseFile)
	userService := &persistence.UserService{SqliteDbFile: databaseFile}
	soundcloudUrlService := &persistence.SoundcloudUrlService{SqliteFile: databaseFile}
	websiteServices := newWebServices(userService, soundcloudUrlService)

	e := newServer(*websiteServices)

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", config.Port)))
}

// authService manages logging users in and authenticating tokens
type authService interface {
	Login(username, password string) (success bool, authToken string)
	IsAuthorized(authToken string) bool
}

type musicService interface {
	GetAllSoundcloudUrls() ([]persistence.SoundcloudUrl, error)
	AddSoundcloudUrl(string) error
	DeleteSoundcloudUrl(string) error
	UpdateSoundcloudUiOrders([]persistence.SoundcloudUrl) error
}

type keyOfDayService interface {
	GetKeyOfDay() string
}

type webServices struct {
	userService          authService
	soundcloudUrlService musicService
	keyOfDayService      keyOfDayService
}

func newWebServices(userService authService, soundcloudUrlService musicService) *webServices {
	return &webServices{
		userService:          userService,
		soundcloudUrlService: soundcloudUrlService,
	}
}

func getNewRelicApp() *newrelic.Application {
	newrelicLicense := os.Getenv("NEW_RELIC_LICENSE")
	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName("go-andrewwillette"),
		newrelic.ConfigLicense(newrelicLicense),
		// newrelic.ConfigDebugLogger(os.Stdout),
	)
	if err != nil {
		logging.GlobalLogger.Error().Msgf("Failed to start new relic app, newrelic key: %s", newrelicLicense)
	}
	return app
}

// newServer setup server with endpoints
func newServer(services webServices) *echo.Echo {
	e := echo.New()
	e.Use(nrecho.Middleware(getNewRelicApp()))
	e.Use(middleware.CORSWithConfig(middleware.DefaultCORSConfig))
	e.GET(getSoundcloudAllEndpoint, services.getAllSoundcloudUrls)
	e.GET(keyOfDayEndpoint, services.getKeyOfDay)
	e.POST(loginEndpoint, services.loginHandler)
	e.PUT(addSoundcloudEndpoint, services.addSoundcloudUrl)
	e.DELETE(deleteSoundcloudEndpoint, services.deleteSoundcloudUrl)
	e.PUT(updateSoundcloudUrlsEndpoint, services.updateSoundcloudUrlUiOrders)
	return e
}

func (u *webServices) getKeyOfDay(c echo.Context) error {
	logging.GlobalLogger.Info().Msg("Calling key of day.")
	c.Response().Header().Set("Content-Type", "application-json")
	c.Response().Header().Set("Access-Control-Allow-Origin", "*")
	if err := json.NewEncoder(c.Response()).Encode(key.GetKeyOfDay()); err != nil {
		const errMsg = "Failed to encode keyOfDay string"
		logging.GlobalLogger.Err(err).Msg(errMsg)
		return c.String(http.StatusInternalServerError, errMsg)
	}
	return c.String(http.StatusOK, "")
}

func (u *webServices) getAllSoundcloudUrls(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "application-json")
	c.Response().Header().Set("Access-Control-Allow-Origin", "*")
	urls, err := u.soundcloudUrlService.GetAllSoundcloudUrls()
	if err != nil {
		const errMsg = "Failed to get soundcloud urls from service."
		logging.GlobalLogger.Err(err).Msg(errMsg)
		return c.String(http.StatusInternalServerError, errMsg)
	}
	var soundcloudUrls = []SoundcloudUrlUiOrderJson{}
	for _, url := range urls {
		soundcloudUrls = append(soundcloudUrls, SoundcloudUrlUiOrderJson{Url: url.Url, UiOrder: url.UiOrder})
	}

	c.Response().WriteHeader(http.StatusOK)
	return json.NewEncoder(c.Response()).Encode(soundcloudUrls)
}

func (u *webServices) addSoundcloudUrl(c echo.Context) error {
	logging.GlobalLogger.Debug().Msg("addSoundcloudUrl called.")
	if c.Request().Method == "OPTIONS" {
		return c.String(http.StatusOK, "Allowing OPTIONS because of prior failed handshaking.")
	}

	var soundcloudData SoundcloudUrlJson
	if err := json.NewDecoder(c.Request().Body).Decode(&soundcloudData); err != nil {
		const errMsg = "Error decoding soundcloud url from request body."
		logging.GlobalLogger.Info().Msg(errMsg)
		return c.String(http.StatusInternalServerError, errMsg)
	}
	if u.userService.IsAuthorized(c.Request().Header.Get("Authorization")) {
		logging.GlobalLogger.Debug().Msg("WilletteToken is valid.")
		err := u.soundcloudUrlService.AddSoundcloudUrl(soundcloudData.Url)
		if err != nil {
			const errMsg = "Error when adding soundcloud url to service layer."
			logging.GlobalLogger.Err(err).Msg(errMsg)
			return c.String(http.StatusInternalServerError, errMsg)
		}
		logging.GlobalLogger.Info().Msg(fmt.Sprintf("Success adding soundcloud url. url: %s", soundcloudData.Url))
		return c.String(http.StatusOK, "Successfuly added soundcloud URL")
	} else {
		const errMsg = "Invalid auth token is invalid."
		logging.GlobalLogger.Info().Msg(errMsg)
		return c.String(http.StatusUnauthorized, errMsg)
	}
}

func (u *webServices) deleteSoundcloudUrl(c echo.Context) error {
	if c.Request().Method == "OPTIONS" {
		return c.String(http.StatusOK, "Allowing OPTIONS because of prior failed handshaking.")
	}
	var soundcloudData SoundcloudUrlJson
	if err := json.NewDecoder(c.Request().Body).Decode(&soundcloudData); err != nil {
		const errMsg = "Error decoding soundcloud url from request body."
		logging.GlobalLogger.Info().Msg(errMsg)
		return c.String(http.StatusInternalServerError, errMsg)
	}
	if u.userService.IsAuthorized(c.Request().Header.Get("Authorization")) {
		err := u.soundcloudUrlService.DeleteSoundcloudUrl(soundcloudData.Url)
		if err != nil {
			switch err.(type) {
			case *persistence.SoundcloudUrlMissingError:
				const errMsg = "Provided url does not exist to delete."
				logging.GlobalLogger.Err(err).Msg(errMsg)
				return c.String(http.StatusBadRequest, errMsg)
			default:
				const errMsg = "Error deleting soundcloudUrl."
				logging.GlobalLogger.Err(err).Msg(errMsg)
				return c.String(http.StatusInternalServerError, errMsg)
			}
		}
		logging.GlobalLogger.Info().Msg(fmt.Sprintf("deleteSoundcloudUrl called successfully for item: %s", soundcloudData.Url))
		return c.String(http.StatusOK, "Successfully deleted soundcloud url")
	} else {
		var errMsg = fmt.Sprintf("deleteSoundcloudUrl called unauthorized for item: %s, authToken: %s", soundcloudData.Url, c.Request().Header.Get("Authorization"))
		logging.GlobalLogger.Info().Msg(errMsg)
		return c.String(http.StatusUnauthorized, errMsg)
	}
}

func (u *webServices) updateSoundcloudUrlUiOrders(c echo.Context) error {
	if c.Request().Method == "OPTIONS" {
		return c.String(http.StatusOK, "Allowing OPTIONS because of prior failed handshaking.")
	}
	var urls []SoundcloudUrlUiOrderJson
	if err := json.NewDecoder(c.Request().Body).Decode(&urls); err != nil {
		const errMsg = "Error decoding soundcloud urls in update soundcloud urls."
		logging.GlobalLogger.Info().Msg(errMsg)
		return c.String(http.StatusBadRequest, errMsg)
	}
	var persistenceUrls []persistence.SoundcloudUrl
	for _, v := range urls {
		persistenceUrls = append(persistenceUrls, persistence.SoundcloudUrl{Url: v.Url, UiOrder: v.UiOrder})
	}
	if err := u.soundcloudUrlService.UpdateSoundcloudUiOrders(persistenceUrls); err != nil {
		const errMsg = "Error updating soundcloud urls."
		logging.GlobalLogger.Err(err).Msg(errMsg)
		return c.String(http.StatusInternalServerError, errMsg)
	}
	return c.String(http.StatusOK, "Sucessfully updated soundcloud url values")
}

func (u *webServices) loginHandler(c echo.Context) error {
	logging.GlobalLogger.Info().Msg("hitting loginHandler")
	if c.Request().Method == "OPTIONS" {
		return c.String(http.StatusOK, "Allowing OPTIONS because of prior failed handshaking.")
	}
	var userCredentials UserJson
	if err := json.NewDecoder(c.Request().Body).Decode(&userCredentials); err != nil {
		const errMsg = "Error decoding user credentials from request body."
		logging.GlobalLogger.Info().Msg(errMsg)

		return c.String(http.StatusInternalServerError, errMsg)
	}
	c.Response().Header().Set("Content-Type", "application-json")
	c.Response().Header().Set("Access-Control-Allow-Origin", "*")

	user := UserJson{Username: userCredentials.Username, Password: userCredentials.Password}
	loginSuccessful, authToken := u.userService.Login(user.Username, user.Password)
	if loginSuccessful {
		if err := json.NewEncoder(c.Response()).Encode(authToken); err != nil {
			const errMsg = "Failed to encode authToken after successful authentication."
			logging.GlobalLogger.Err(err).Msg(errMsg)
			return c.String(http.StatusUnauthorized, errMsg)
		}
		logging.GlobalLogger.Info().Msgf("Login Successful. username: %s", userCredentials.Username)
		return c.String(http.StatusOK, "")
	} else {
		logging.GlobalLogger.Info().Msg(fmt.Sprintf("Login failed with username: %s, password: %s", user.Username, user.Password))
		return c.String(http.StatusUnauthorized, "Login Failed.")
	}
}
