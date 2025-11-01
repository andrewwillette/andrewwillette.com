package config

import (
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var C Config

const defaultConfigDir = "/.config/andrewwillette.com"

func init() {
	res, err := LoadConfig(".")
	if err != nil {
		log.Error().Msgf("Error loading config: %v", err)
	}
	C = res
}

type Config struct {
	PProfEnabled        bool   `mapstructure:"PPROF_ENABLED"`
	LogLevel            string `mapstructure:"LOG_LEVEL"`
	LogConsole          bool   `mapstructure:"LOG_CONSOLE"`
	LogFile             bool   `mapstructure:"LOG_FILE"`
	LogJSON             bool   `mapstructure:"LOG_JSON"`
	LogDir              string `mapstructure:"LOG_DIR"`
	LogFileName         string `mapstructure:"LOG_FILE_NAME"`
	LogFileMaxMB        int    `mapstructure:"LOG_FILE_MAX_MB"`
	LogFileMaxBacks     int    `mapstructure:"LOG_FILE_MAX_BACKUPS"`
	LogFileMaxAge       int    `mapstructure:"LOG_FILE_MAX_AGE"`
	AudioS3BucketName   string `mapstructure:"AUDIO_S3_BUCKET_NAME"`
	AudioS3BucketPrefix string `mapstructure:"AUDIO_S3_BUCKET_PREFIX"`
	AudioS3Region       string `mapstructure:"AUDIO_S3_REGION"`
	AudioS3URL          string `mapstructure:"AUDIO_S3_URL"`
	AudioSQSURL         string `mapstructure:"AUDIO_SQS_URL"`
	HomePageImageS3URL  string `mapstructure:"HOME_PAGE_IMAGE_S3_URL"`
}

func LoadConfig(path string) (config Config, err error) {
	configName := os.Getenv("ENV")
	if configName != "PROD" {
		configName = "app"
	} else {
		configName = "prod"
	}

	viper.SetConfigName(configName)
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	viper.AddConfigPath(path)
	readErr := viper.ReadInConfig()

	if readErr != nil {
		home, herr := os.UserHomeDir()
		if herr == nil {
			viper.AddConfigPath(home + defaultConfigDir)
			readErr = viper.ReadInConfig()
		}
	}

	if readErr != nil {
		return config, readErr
	}

	if err = viper.Unmarshal(&config); err != nil {
		return config, err
	}

	return config, nil
}
