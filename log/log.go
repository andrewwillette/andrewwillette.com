package log

import (
	"io"
	"os"
	"path"

	"github.com/andrewwillette/andrewwillettedotcom/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

func Configure() {
	var writers []io.Writer

	if config.C.LogConsole {
		if config.C.LogJSON {
			writers = append(writers, os.Stderr)
		} else {
			writers = append(writers, zerolog.ConsoleWriter{Out: os.Stderr})
		}
	}

	if config.C.LogFile && os.Getenv("ENV") == "PROD" {
		logWriter, err := newRollingFile()
		if err != nil {
			panic(err)
		}
		writers = append(writers, logWriter)
	}

	level, err := zerolog.ParseLevel(config.C.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	multi := io.MultiWriter(writers...)
	log.Logger = zerolog.New(multi).With().Timestamp().Logger()
}

func newRollingFile() (io.Writer, error) {
	if err := os.MkdirAll(config.C.LogDir, 0744); err != nil {
		log.Error().Err(err).Str("path", config.C.LogDir).Msg("Failed to create log directory")
		return nil, err
	}

	return &lumberjack.Logger{
		Filename:   path.Join(config.C.LogDir, config.C.LogFileName),
		MaxBackups: config.C.LogFileMaxBacks,
		MaxSize:    config.C.LogFileMaxMB,
		MaxAge:     config.C.LogFileMaxAge,
	}, nil
}
