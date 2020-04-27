package log

import (
	"io"
	"time"

	"github.com/rs/zerolog"
)

// InitializeGlobalSettings initializes global settings
func InitializeGlobalSettings() {
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.DurationFieldUnit = time.Millisecond
	zerolog.DurationFieldInteger = true
}

// New creates new isla logger
func New(serviceName string, logLevel string, writers io.Writer) *zerolog.Logger {
	defaultLogLevel := zerolog.ErrorLevel
	if requiredLogLevel, err := zerolog.ParseLevel(logLevel); err == nil {
		defaultLogLevel = requiredLogLevel
	}
	logger := zerolog.New(writers).Level(defaultLogLevel).With().Timestamp().Str("service", serviceName).Logger()
	return &logger
}
