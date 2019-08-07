package log

import (
	log "github.com/sirupsen/logrus"
)

// Logx is custom initialized logger
type Logx struct {
	logger *log.Logger
}

// Formatted returns preconfigured logger
func (logx *Logx) Formatted() *log.Logger {
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	logx.logger.SetReportCaller(true)
	logx.logger.SetFormatter(formatter)
	return logx.logger
}
