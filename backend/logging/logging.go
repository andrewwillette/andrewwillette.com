package logging

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path"
)

var logConfig = LogConfig{
	ConsoleLoggingEnabled: true,
	EncodeLogsAsJson:      true,
	FileLoggingEnabled:    true,
	Directory:             "./logging",
	Filename:              "server.log",
	MaxSizeMB:             200,
	MaxBackups:            2,
	MaxAge:                31,
	LogLevel:              zerolog.DebugLevel,
}
var GlobalLogger = Configure(logConfig)
var testlogConfig = LogConfig{
	ConsoleLoggingEnabled: true,
	EncodeLogsAsJson:      true,
	FileLoggingEnabled:    true,
	Directory:             "./logging",
	Filename:              "test.log",
	MaxSizeMB:             200,
	MaxBackups:            2,
	MaxAge:                31,
	LogLevel:              zerolog.DebugLevel,
}
var TestLogger = Configure(testlogConfig)

type LogConfig struct {
	ConsoleLoggingEnabled bool
	EncodeLogsAsJson      bool
	FileLoggingEnabled    bool
	Directory             string
	Filename              string
	MaxSizeMB             int
	MaxBackups            int
	MaxAge                int
	LogLevel              zerolog.Level
}

type Logger struct {
	*zerolog.Logger
}

func Configure(config LogConfig) *Logger {
	var writers []io.Writer

	if config.ConsoleLoggingEnabled {
		writers = append(writers, zerolog.ConsoleWriter{Out: os.Stderr})
	}
	if config.FileLoggingEnabled {
		writers = append(writers, newRollingFile(config))
	}
	mw := io.MultiWriter(writers...)
	zerolog.SetGlobalLevel(config.LogLevel)
	logger := zerolog.New(mw).With().Timestamp().Logger()
	logger.Info().
		Bool("fileLogging", config.FileLoggingEnabled).
		Bool("jsonLogOutput", config.EncodeLogsAsJson).
		Str("logDirectory", config.Directory).
		Str("fileName", config.Filename).
		Int("maxSizeMB", config.MaxSizeMB).
		Int("maxBackups", config.MaxBackups).
		Int("maxAgeInDays", config.MaxAge)

	return &Logger{
		Logger: &logger,
	}
}

func newRollingFile(config LogConfig) io.Writer {
	if err := os.MkdirAll(config.Directory, 0744); err != nil {
		log.Error().Err(err).Str("path", config.Directory).Msg("can't create log directory")
		return nil
	}

	return &lumberjack.Logger{
		Filename:   path.Join(config.Directory, config.Filename),
		MaxBackups: config.MaxBackups, // files
		MaxSize:    config.MaxSizeMB,  // megabytes
		MaxAge:     config.MaxAge,     // days
	}
}
