package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/viper"

	gormLogger "gorm.io/gorm/logger"
)

func contains(s []string, e string) int {
	for index, a := range s {
		if a == e {
			return index
		}
	}
	return -1
}

var AllowedLogLevels = []string{"debug", "info", "warn", "error", "fatal", "panic"}

// get gorm log level based on global zerolog level
func GetGormLogLevel() gormLogger.LogLevel {
	zerologLevel := zerolog.GlobalLevel()
	if zerologLevel >= 3 {
		return gormLogger.Error
	} else if zerologLevel == 2 {
		return gormLogger.Warn
	}

	return gormLogger.Info
}

func NewLogger(serviceName string) zerolog.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	loggerLevel := viper.GetString("logger.level")

	logLevelIndex := contains(AllowedLogLevels, loggerLevel)

	if logLevelIndex != -1 {
		calculatedLogLevel := logLevelIndex
		zerolog.SetGlobalLevel(zerolog.Level(calculatedLogLevel))
	} else {
		panic(fmt.Sprintf("invalid log level provided: %s, expected one of: %s", loggerLevel, strings.Join(AllowedLogLevels[:], ",")))
	}

	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	// add defaults and return instance
	loggerInstance := log.Logger
	if serviceName != "" {
		loggerInstance = log.With().Str("service", serviceName).Logger()
	}

	// enable pretty print based on config
	jsonFormat := viper.GetBool("logger.output.json")
	if !jsonFormat {
		loggerInstance = loggerInstance.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	return loggerInstance
}
