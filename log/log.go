package log

import (
	log "github.com/sirupsen/logrus"
)

var logx *log.Logger

func init() {
	logx = log.New()
}

// Formatted returns preconfigured logger
func Formatted() *log.Logger {

	logx.SetFormatter(&log.JSONFormatter{})
	logx.SetReportCaller(true)

	return logx
}
