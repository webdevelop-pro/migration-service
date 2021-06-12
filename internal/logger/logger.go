package logger

import (
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/webdevelop-pro/migration-service/internal/config"
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
}

// Logger is wrapper struct around zerolog.Logger that adds some custom functionality
type Logger struct {
	zerolog.Logger
}

// Printf is implementation of fx.Printer
func (l Logger) Printf(s string, args ...interface{}) {
	l.Info().Msgf(s, args...)
}

// NewLogger return logger instance
func NewLogger(component string, output io.Writer, cfg *config.Config) Logger {
	logLevel := cfg.LogLevel
	if logLevel == "" {
		logLevel = os.Getenv("LOG_LEVEL")
	}

	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		panic(err)
	}

	// Beautiful output
	if output == os.Stdout && cfg.LogConsole {
		output = zerolog.NewConsoleWriter()
	} else if output == nil {
		output = os.Stdout
	}
	return Logger{
		zerolog.
			New(output).
			Level(level).
			Hook(severityHook{}).
			With().Caller().
			Timestamp().
			Str("version", cfg.GitHash).
			Str("component", component).
			Logger(),
	}
}
