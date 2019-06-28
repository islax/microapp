package util

import (
	log "github.com/sirupsen/logrus"
)

var logx *log.Logger

func init() {
	logx = log.New()
}

// NewLogger returns preconfigured logger
func NewLogger() *log.Logger {
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	logx.SetReportCaller(true)
	logx.SetFormatter(formatter)
	return logx
}
