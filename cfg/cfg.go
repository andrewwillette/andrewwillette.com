package cfg

import "github.com/spf13/viper"
import "github.com/rs/zerolog/log"

var C Config

func init() {
	res, err := LoadConfig(".")
	if err != nil {
		log.Error().Msgf("Error loading config: %v", err)
	}
	C = res
}

type Config struct {
	PProfEnabled bool   `mapstructure:"PPROF_ENABLED"`
	LogLevel     string `mapstructure:"LOG_LEVEL"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
