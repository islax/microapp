package log

import (
	"github.com/sirupsen/logrus"
)

// Logger is custom formatted text logger
var Logger *logrus.Logger

func init() {
	customFormattedLogger := CustomLogger{}
	Logger = customFormattedLogger.formatIt()
}

// CustomLogger is custom logger
type CustomLogger struct {
	customlogger *logrus.Logger
}

func (cl *CustomLogger) formatIt() *logrus.Logger {
	cl.customlogger = logrus.New()
	cl.customlogger.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
	filenameHook := NewHook()
	filenameHook.Field = "line"
	cl.customlogger.AddHook(filenameHook)
	return cl.customlogger
}
