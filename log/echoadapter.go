package log

import (
	"fmt"
	"io"

	gommonlog "github.com/labstack/gommon/log"
	"github.com/rs/zerolog"
)

// EchoAdapter adapts a zerolog.Logger to Echo's Logger interface so the Echo
// instance uses the same structured logger as the rest of the application.
type EchoAdapter struct {
	logger zerolog.Logger
	prefix string
	level  gommonlog.Lvl
}

// NewEchoAdapter wraps logger in an EchoAdapter ready to assign to echo.Echo.Logger.
func NewEchoAdapter(logger zerolog.Logger) *EchoAdapter {
	return &EchoAdapter{logger: logger, level: gommonlog.INFO}
}

func (z *EchoAdapter) Output() io.Writer                      { return z.logger }
func (z *EchoAdapter) SetOutput(w io.Writer)                  { z.logger = z.logger.Output(w) }
func (z *EchoAdapter) Prefix() string                         { return z.prefix }
func (z *EchoAdapter) SetPrefix(p string)                     { z.prefix = p }
func (z *EchoAdapter) Level() gommonlog.Lvl                   { return z.level }
func (z *EchoAdapter) SetLevel(l gommonlog.Lvl)               { z.level = l }
func (z *EchoAdapter) SetHeader(h string)                     {}
func (z *EchoAdapter) Print(i ...interface{})                 { z.logger.Info().Msg(fmt.Sprint(i...)) }
func (z *EchoAdapter) Printf(format string, i ...interface{}) { z.logger.Info().Msgf(format, i...) }
func (z *EchoAdapter) Printj(j gommonlog.JSON)                { z.logger.Info().Fields(j).Msg("") }
func (z *EchoAdapter) Debug(i ...interface{})                 { z.logger.Debug().Msg(fmt.Sprint(i...)) }
func (z *EchoAdapter) Debugf(format string, i ...interface{}) { z.logger.Debug().Msgf(format, i...) }
func (z *EchoAdapter) Debugj(j gommonlog.JSON)                { z.logger.Debug().Fields(j).Msg("") }
func (z *EchoAdapter) Info(i ...interface{})                  { z.logger.Info().Msg(fmt.Sprint(i...)) }
func (z *EchoAdapter) Infof(format string, i ...interface{})  { z.logger.Info().Msgf(format, i...) }
func (z *EchoAdapter) Infoj(j gommonlog.JSON)                 { z.logger.Info().Fields(j).Msg("") }
func (z *EchoAdapter) Warn(i ...interface{})                  { z.logger.Warn().Msg(fmt.Sprint(i...)) }
func (z *EchoAdapter) Warnf(format string, i ...interface{})  { z.logger.Warn().Msgf(format, i...) }
func (z *EchoAdapter) Warnj(j gommonlog.JSON)                 { z.logger.Warn().Fields(j).Msg("") }
func (z *EchoAdapter) Error(i ...interface{})                 { z.logger.Error().Msg(fmt.Sprint(i...)) }
func (z *EchoAdapter) Errorf(format string, i ...interface{}) { z.logger.Error().Msgf(format, i...) }
func (z *EchoAdapter) Errorj(j gommonlog.JSON)                { z.logger.Error().Fields(j).Msg("") }
func (z *EchoAdapter) Fatal(i ...interface{})                 { z.logger.Fatal().Msg(fmt.Sprint(i...)) }
func (z *EchoAdapter) Fatalf(format string, i ...interface{}) { z.logger.Fatal().Msgf(format, i...) }
func (z *EchoAdapter) Fatalj(j gommonlog.JSON)                { z.logger.Fatal().Fields(j).Msg("") }
func (z *EchoAdapter) Panic(i ...interface{})                 { z.logger.Panic().Msg(fmt.Sprint(i...)) }
func (z *EchoAdapter) Panicf(format string, i ...interface{}) { z.logger.Panic().Msgf(format, i...) }
func (z *EchoAdapter) Panicj(j gommonlog.JSON)                { z.logger.Panic().Fields(j).Msg("") }
