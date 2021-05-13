package log

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/rs/zerolog"
	glogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

type Config struct {
	SlowThreshold time.Duration
}

type gormlogger struct {
	Config Config
	logger zerolog.Logger
}

func NewGormLogger(appName, logLevel string, writer io.Writer, config Config) glogger.Interface {
	return &gormlogger{
		Config: config,
		logger: *New(appName, logLevel, writer),
	}
}

// LogMode log mode
func (l *gormlogger) LogMode(level glogger.LogLevel) glogger.Interface {
	return l
}

// Info print info
func (l *gormlogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.logger.Info().Msgf("%s %s", msg, append([]interface{}{utils.FileWithLineNum()}, data...))
}

// Warn print warn messages
func (l *gormlogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.logger.Warn().Msgf("%s %s", msg, append([]interface{}{utils.FileWithLineNum()}, data...))
}

// Error print error messages
func (l *gormlogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.logger.Error().Msgf("%s %s", msg, append([]interface{}{utils.FileWithLineNum()}, data...))
}

// Trace print sql message
func (l *gormlogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	switch {
	case err != nil:
		sql, rows := fc()
		if rows == -1 {
			l.logger.Trace().Msgf("%s %s %f - %s", err, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, sql)
		} else {
			l.logger.Trace().Msgf("%s %s %f rows:%d %s", err, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > l.Config.SlowThreshold && l.Config.SlowThreshold != 0:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.Config.SlowThreshold)
		if rows == -1 {
			l.logger.Trace().Msgf("%s %s %f - %s", slowLog, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, sql)
		} else {
			l.logger.Trace().Msgf("%s %s %f rows:%d %s", slowLog, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	default:
		sql, rows := fc()
		if rows == -1 {
			l.logger.Trace().Msgf("%s %f - %s", utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, sql)
		} else {
			l.logger.Trace().Msgf("%s %f rows:%d %s", utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}
