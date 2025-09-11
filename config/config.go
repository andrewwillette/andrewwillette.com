package config

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
	PProfEnabled    bool   `mapstructure:"PPROF_ENABLED"`
	LogLevel        string `mapstructure:"LOG_LEVEL"`
	LogConsole      bool   `mapstructure:"LOG_CONSOLE"`
	LogFile         bool   `mapstructure:"LOG_FILE"`
	LogJSON         bool   `mapstructure:"LOG_JSON"`
	LogDir          string `mapstructure:"LOG_DIR"`
	LogFileName     string `mapstructure:"LOG_FILE_NAME"`
	LogFileMaxMB    int    `mapstructure:"LOG_FILE_MAX_MB"`
	LogFileMaxBacks int    `mapstructure:"LOG_FILE_MAX_BACKUPS"`
	LogFileMaxAge   int    `mapstructure:"LOG_FILE_MAX_AGE"`
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
