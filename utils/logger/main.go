package logger

import (
	"os"
	"sync"

	"github.com/rs/zerolog"
)

var (
	Logger zerolog.Logger
	once   sync.Once
)

func init() {
	once.Do(func() {
		output := zerolog.ConsoleWriter{Out: os.Stdout}
		Logger = zerolog.New(output).With().Timestamp().Logger()
	})
}

func Log(message string) {
	Logger.Info().Msg(message)
}

func LogDebug(message string) {
	Logger.Debug().Msg(message)
}

func LogWarn(message string) {
	Logger.Warn().Msg(message)
}

func LogError(message string) {
	Logger.Error().Msg(message)
}

func LogFatal(message string) {
	Logger.Fatal().Msg(message)
}
