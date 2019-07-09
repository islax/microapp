package log

import (
	log "github.com/sirupsen/logrus"
)

var logx *log.Logger

func init() {
	logx = log.New()
}

// New returns preconfigured logger
func formatted() *log.Logger {
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	logx.SetReportCaller(true)
	logx.SetFormatter(formatter)
	return logx
}
