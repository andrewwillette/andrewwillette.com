package config

import (
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var C Config

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
		configName = "app" // default outside of prod
	}

	viper.AddConfigPath(path)
	viper.SetConfigName(configName)
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
