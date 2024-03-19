package log

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	glogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

const defaultSlowSqlMs = 200

type GormLog struct {
	SlowThreshold time.Duration
	logger        *zap.SugaredLogger
}

// 入参是慢SQL的时间，默认是200ms
func NewGormLog(slowSqlMs ...int) *GormLog {
	slowMs := defaultSlowSqlMs
	if len(slowSqlMs) > 0 && slowSqlMs[0] > 0 {
		slowMs = slowSqlMs[0]
	}

	return &GormLog{
		SlowThreshold: time.Duration(slowMs) * time.Millisecond,
		logger:        With("gormLog", "gorm log"),
	}
}

func (l *GormLog) LogMode(level glogger.LogLevel) glogger.Interface {
	return l
}

func (l *GormLog) Info(ctx context.Context, fmt string, args ...interface{}) {
	l.logger.Infof(fmt, args...)
}

func (l *GormLog) Warn(ctx context.Context, fmt string, args ...interface{}) {
	l.logger.Warnf(fmt, args...)
}

func (l *GormLog) Error(ctx context.Context, fmt string, args ...interface{}) {
	l.logger.Errorf(fmt, args...)
}

func (l *GormLog) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	rowStr := ""
	if rows == -1 {
		rowStr = fmt.Sprintf("[rows:%v] ", rows)
	}

	if err != nil {
		l.logger.Errorw(fmt.Sprintf("%s err(%v)", sql, err), "file", utils.FileWithLineNum(), "time", fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6), "row", rowStr, "sql", sql)
		return
	}

	if elapsed > l.SlowThreshold && l.SlowThreshold != 0 {
		slowLog := fmt.Sprintf("SLOW SQL >= %v ", l.SlowThreshold)
		l.logger.Errorw(slowLog+sql, "is_show", 1, "file", utils.FileWithLineNum(), "time", fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6), "row", rowStr, "sql", sql)
		return
	}

	l.logger.Infow(sql, "file", utils.FileWithLineNum(), "time", fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6), "row", rowStr, "sql", sql)
}
