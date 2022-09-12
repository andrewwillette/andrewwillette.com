package logging

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path"
)

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

type Logger struct {
	*zerolog.Logger
}

var globalLogConfig = logConfig{
	consoleLoggingEnabled: true,
	encodeLogsAsJson:      true,
	fileLoggingEnabled:    true,
	directory:             "./logging",
	filename:              "server.log",
	maxSizeMB:             200,
	maxBackups:            2,
	maxAge:                31,
	logLevel:              zerolog.DebugLevel,
}
var GlobalLogger = configure(globalLogConfig)

// configure return a zerolog logger with provided behavior based off
// the provided LogConfig
func configure(config logConfig) *Logger {
	var writers []io.Writer

	if config.consoleLoggingEnabled {
		writers = append(writers, zerolog.ConsoleWriter{Out: os.Stderr})
	}
	if config.fileLoggingEnabled {
		writers = append(writers, newRollingFile(config))
	}
	mw := io.MultiWriter(writers...)
	zerolog.SetGlobalLevel(config.logLevel)
	logger := zerolog.New(mw).With().Timestamp().Logger()
	logger.Info().
		Bool("fileLogging", config.fileLoggingEnabled).
		Bool("jsonLogOutput", config.encodeLogsAsJson).
		Str("logDirectory", config.directory).
		Str("fileName", config.filename).
		Int("maxSizeMB", config.maxSizeMB).
		Int("maxBackups", config.maxBackups).
		Int("maxAgeInDays", config.maxAge)

	return &Logger{
		Logger: &logger,
	}
}

// newRollingFile return new Writer value for use with zerolog logging writers
func newRollingFile(config logConfig) io.Writer {
	if err := os.MkdirAll(config.directory, 0744); err != nil {
		log.Error().Err(err).Str("path", config.directory).Msg("can't create log directory")
		return nil
	}

	return &lumberjack.Logger{
		Filename:   path.Join(config.directory, config.filename),
		MaxBackups: config.maxBackups, // files
		MaxSize:    config.maxSizeMB,  // megabytes
		MaxAge:     config.maxAge,     // days
	}
}
