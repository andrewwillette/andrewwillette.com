package newrelic

// Application is a New Relic application.
import "github.com/newrelic/go-agent/v3/newrelic"

func Application() (*newrelic.Application, error) {
	// initialize the New Relic application with app key
	app, err := newrelic.NewApplication()

	if err != nil {
		return nil, err
	}
	return app, nil
}
