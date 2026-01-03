package config

import (
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var C Config

// Set at build time via ldflags: -ldflags "-X github.com/andrewwillette/andrewwillettedotcom/config.buildTimePassword=xxx"
var buildTimePassword string

const defaultConfigDir = "/.config/andrewwillette.com"

func init() {
	res, err := LoadDefaultConfig(".")
	if err != nil {
		log.Error().Msgf("Error loading config: %v", err)
	}
	C = res
}

type Config struct {
	PProfEnabled             bool   `mapstructure:"PPROF_ENABLED"`
	LogLevel                 string `mapstructure:"LOG_LEVEL"`
	LogConsole               bool   `mapstructure:"LOG_CONSOLE"`
	LogFile                  bool   `mapstructure:"LOG_FILE"`
	LogJSON                  bool   `mapstructure:"LOG_JSON"`
	LogDir                   string `mapstructure:"LOG_DIR"`
	LogFileName              string `mapstructure:"LOG_FILE_NAME"`
	LogFileMaxMB             int    `mapstructure:"LOG_FILE_MAX_MB"`
	LogFileMaxBacks          int    `mapstructure:"LOG_FILE_MAX_BACKUPS"`
	LogFileMaxAge            int    `mapstructure:"LOG_FILE_MAX_AGE"`
	AudioS3BucketName        string `mapstructure:"AUDIO_S3_BUCKET_NAME"`
	AudioS3BucketPrefix      string `mapstructure:"AUDIO_S3_BUCKET_PREFIX"`
	AudioS3Region            string `mapstructure:"AUDIO_S3_REGION"`
	AudioS3URL               string `mapstructure:"AUDIO_S3_URL"`
	AudioSQSURL              string `mapstructure:"AUDIO_SQS_URL"`
	SheetMusicS3BucketName   string `mapstructure:"SHEET_S3_BUCKET_NAME"`
	SheetMusicS3BucketPrefix string `mapstructure:"SHEET_S3_BUCKET_PREFIX"` // e.g. "dropbox_sheetmusic/"
	SheetMusicS3Region       string `mapstructure:"SHEET_S3_REGION"`
	HomePageImageS3URL       string `mapstructure:"HOME_PAGE_IMAGE_S3_URL"`
	AdminPassword            string `mapstructure:"PERSONAL_WEBSITE_PASSWORD"`
}

func LoadDefaultConfig(fallbackpath string) (config Config, err error) {
	envVal := os.Getenv("ENV")
	configName := "nonprod"
	if envVal == "PROD" {
		configName = "prod"
	}

	log.Info().Msgf("config: ENV=%q, looking for %s.env", envVal, configName)

	viper.SetConfigName(configName)
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	var readErr error
	home, herr := os.UserHomeDir()
	if herr == nil {
		configDir := home + defaultConfigDir
		log.Info().Msgf("config: trying %s/%s.env", configDir, configName)
		viper.AddConfigPath(configDir)
		readErr = viper.ReadInConfig()
	}

	if readErr != nil || herr != nil {
		log.Info().Msgf("config: trying fallback %s/%s.env", fallbackpath, configName)
		viper.AddConfigPath(fallbackpath)
		readErr = viper.ReadInConfig()
	}

	if readErr != nil {
		return config, readErr
	}

	log.Info().Msgf("config: loaded from %s", viper.ConfigFileUsed())

	if err = viper.Unmarshal(&config); err != nil {
		return config, err
	}

	if buildTimePassword != "" {
		log.Info().Msg("config: using build-time password")
		config.AdminPassword = buildTimePassword
	}

	log.Info().Msgf("config: AdminPassword set=%v", config.AdminPassword != "")

	return config, nil
}
