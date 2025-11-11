package logger

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"gorm.io/gorm/logger"
)

// GormLogger GORM SQL 日志记录器，使用 hlog 记录
type GormLogger struct {
	// LogLevel 日志级别
	LogLevel logger.LogLevel
	// SlowThreshold 慢查询阈值，默认 200ms
	SlowThreshold time.Duration
	// IgnoreRecordNotFoundError 是否忽略记录未找到错误，默认 true
	IgnoreRecordNotFoundError bool
}

// NewGormLogger 创建新的 GORM logger
// level: 日志级别，可选值: silent, error, warn, info
// slowThreshold: 慢查询阈值，默认 200ms
func NewGormLogger(level logger.LogLevel, slowThreshold time.Duration) *GormLogger {
	if slowThreshold == 0 {
		slowThreshold = 200 * time.Millisecond
	}
	return &GormLogger{
		LogLevel:                  level,
		SlowThreshold:             slowThreshold,
		IgnoreRecordNotFoundError: true,
	}
}

// LogMode 设置日志级别
func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

// Info 记录信息日志
func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		if len(data) > 0 {
			hlog.CtxInfof(ctx, msg, data...)
		} else {
			hlog.CtxInfof(ctx, "%s", msg)
		}
	}
}

// Warn 记录警告日志
func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		if len(data) > 0 {
			hlog.CtxWarnf(ctx, msg, data...)
		} else {
			hlog.CtxWarnf(ctx, "%s", msg)
		}
	}
}

// Error 记录错误日志
func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		if len(data) > 0 {
			hlog.CtxErrorf(ctx, msg, data...)
		} else {
			hlog.CtxErrorf(ctx, "%s", msg)
		}
	}
}

// Trace 记录 SQL 执行日志
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	switch {
	case err != nil && l.LogLevel >= logger.Error && (!l.IgnoreRecordNotFoundError || err != logger.ErrRecordNotFound):
		// 记录错误日志
		hlog.CtxErrorf(ctx, "[GORM] SQL: %s | Rows: %d | Error: %v | Elapsed: %v", sql, rows, err, elapsed)
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn:
		// 记录慢查询日志
		hlog.CtxWarnf(ctx, "[GORM] Slow SQL: %s | Rows: %d | Elapsed: %v", sql, rows, elapsed)
	case l.LogLevel >= logger.Info:
		// 记录普通 SQL 日志
		hlog.CtxInfof(ctx, "[GORM] SQL: %s | Rows: %d | Elapsed: %v", sql, rows, elapsed)
	}
}

// DefaultGormLogger 返回默认的 GORM logger（Info 级别）
func DefaultGormLogger() *GormLogger {
	return NewGormLogger(logger.Info, 200*time.Millisecond)
}

// SilentGormLogger 返回静默的 GORM logger（不记录日志）
func SilentGormLogger() *GormLogger {
	return NewGormLogger(logger.Silent, 0)
}

// ErrorGormLogger 返回只记录错误的 GORM logger
func ErrorGormLogger() *GormLogger {
	return NewGormLogger(logger.Error, 0)
}

// WarnGormLogger 返回记录警告和错误的 GORM logger
func WarnGormLogger() *GormLogger {
	return NewGormLogger(logger.Warn, 200*time.Millisecond)
}

// InfoGormLogger 返回记录所有日志的 GORM logger
func InfoGormLogger() *GormLogger {
	return NewGormLogger(logger.Info, 200*time.Millisecond)
}
