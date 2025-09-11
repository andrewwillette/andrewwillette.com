package logging

import (
	"io"
	"os"
	"path"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

func init() {
	configure()
}

type logConfig struct {
	consoleLoggingEnabled bool
	encodeLogsAsJson      bool
	fileLoggingEnabled    bool
	directory             string
	filename              string
	maxSizeMB             int
	maxBackups            int
	maxAge                int
	logLevel              zerolog.Level
}

var globalLogConfig = logConfig{
	consoleLoggingEnabled: true,
	logLevel:              zerolog.InfoLevel,
	encodeLogsAsJson:      true,
	fileLoggingEnabled:    false,
	directory:             "./logging",
	filename:              "server.log",
	maxSizeMB:             200,
	maxBackups:            2,
	maxAge:                31,
}

// configure return a zerolog logger with provided behavior based off
// the provided LogConfig
func configure() {
	var writers []io.Writer
	if globalLogConfig.consoleLoggingEnabled {
		writers = append(writers, zerolog.ConsoleWriter{Out: os.Stderr})
	}
	if globalLogConfig.fileLoggingEnabled {
		logWriter, err := newRollingFile(globalLogConfig)
		if err != nil {
			panic(err)
		}
		writers = append(writers, logWriter)
	}
	mw := io.MultiWriter(writers...)
	zerolog.SetGlobalLevel(globalLogConfig.logLevel)
	logger := zerolog.New(mw).With().Timestamp().Logger()
	log.Logger = logger
}

// newRollingFile return new Writer value for use with zerolog logging writers
func newRollingFile(config logConfig) (io.Writer, error) {
	if err := os.MkdirAll(config.directory, 0744); err != nil {
		log.Error().Err(err).Str("path", config.directory).Msgf("Failed to create log directory: %s", config.directory)
		return nil, err
	}

	return &lumberjack.Logger{
		Filename:   path.Join(config.directory, config.filename),
		MaxBackups: config.maxBackups, // files
		MaxSize:    config.maxSizeMB,  // megabytes
		MaxAge:     config.maxAge,     // days
	}, nil
}
