package logger

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// RedisLogger Redis 日志记录器，使用 hlog 记录
type RedisLogger struct {
	// LogLevel 日志级别
	// 0: Silent (不记录)
	// 1: Error (只记录错误)
	// 2: Warn (记录警告和错误)
	// 3: Info (记录所有日志)
	LogLevel int
	// SlowThreshold 慢操作阈值，默认 100ms
	SlowThreshold time.Duration
	// LogCommands 是否记录命令，默认 true
	LogCommands bool
	// LogErrors 是否记录错误，默认 true
	LogErrors bool
}

// NewRedisLogger 创建新的 Redis logger
// level: 日志级别，0=Silent, 1=Error, 2=Warn, 3=Info
// slowThreshold: 慢操作阈值，默认 100ms
func NewRedisLogger(level int, slowThreshold time.Duration) *RedisLogger {
	if slowThreshold == 0 {
		slowThreshold = 100 * time.Millisecond
	}
	return &RedisLogger{
		LogLevel:      level,
		SlowThreshold: slowThreshold,
		LogCommands:   true,
		LogErrors:     true,
	}
}

// LogLevel constants
const (
	RedisLogLevelSilent = 0
	RedisLogLevelError  = 1
	RedisLogLevelWarn   = 2
	RedisLogLevelInfo   = 3
)

// LogCommand 记录 Redis 命令执行
// cmd: 命令名称
// args: 命令参数
// duration: 执行耗时
// err: 错误信息（如果有）
func (l *RedisLogger) LogCommand(ctx context.Context, cmd string, args []interface{}, duration time.Duration, err error) {
	if l.LogLevel <= RedisLogLevelSilent {
		return
	}

	// 构建日志消息
	var msg string
	if len(args) > 0 {
		msg = formatRedisCommand(cmd, args)
	} else {
		msg = cmd
	}

	switch {
	case err != nil && l.LogErrors && l.LogLevel >= RedisLogLevelError:
		// 记录错误日志
		hlog.CtxErrorf(ctx, "[Redis] %s | Error: %v | Elapsed: %v", msg, err, duration)
	case duration > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= RedisLogLevelWarn:
		// 记录慢操作警告
		hlog.CtxWarnf(ctx, "[Redis] Slow %s | Elapsed: %v", msg, duration)
	case l.LogCommands && l.LogLevel >= RedisLogLevelInfo:
		// 记录普通操作日志
		hlog.CtxInfof(ctx, "[Redis] %s | Elapsed: %v", msg, duration)
	}
}

// LogPipeline 记录 Redis Pipeline 执行
// cmds: 命令列表
// duration: 执行耗时
// err: 错误信息（如果有）
func (l *RedisLogger) LogPipeline(ctx context.Context, cmds []string, duration time.Duration, err error) {
	if l.LogLevel <= RedisLogLevelSilent {
		return
	}

	msg := formatRedisPipeline(cmds)

	switch {
	case err != nil && l.LogErrors && l.LogLevel >= RedisLogLevelError:
		// 记录错误日志
		hlog.CtxErrorf(ctx, "[Redis] Pipeline: %s | Error: %v | Elapsed: %v", msg, err, duration)
	case duration > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= RedisLogLevelWarn:
		// 记录慢操作警告
		hlog.CtxWarnf(ctx, "[Redis] Slow Pipeline: %s | Elapsed: %v", msg, duration)
	case l.LogCommands && l.LogLevel >= RedisLogLevelInfo:
		// 记录普通操作日志
		hlog.CtxInfof(ctx, "[Redis] Pipeline: %s | Elapsed: %v", msg, duration)
	}
}

// LogError 记录 Redis 错误
func (l *RedisLogger) LogError(ctx context.Context, err error) {
	if l.LogLevel >= RedisLogLevelError && l.LogErrors {
		hlog.CtxErrorf(ctx, "[Redis] Error: %v", err)
	}
}

// LogInfo 记录 Redis 信息日志
func (l *RedisLogger) LogInfo(ctx context.Context, msg string, args ...interface{}) {
	if l.LogLevel >= RedisLogLevelInfo {
		if len(args) > 0 {
			hlog.CtxInfof(ctx, "[Redis] "+msg, args...)
		} else {
			hlog.CtxInfof(ctx, "[Redis] %s", msg)
		}
	}
}

// LogWarn 记录 Redis 警告日志
func (l *RedisLogger) LogWarn(ctx context.Context, msg string, args ...interface{}) {
	if l.LogLevel >= RedisLogLevelWarn {
		if len(args) > 0 {
			hlog.CtxWarnf(ctx, "[Redis] "+msg, args...)
		} else {
			hlog.CtxWarnf(ctx, "[Redis] %s", msg)
		}
	}
}

// formatRedisCommand 格式化 Redis 命令
func formatRedisCommand(cmd string, args []interface{}) string {
	if len(args) == 0 {
		return cmd
	}

	// 限制参数长度，避免日志过长
	maxArgs := 5
	if len(args) > maxArgs {
		args = args[:maxArgs]
	}

	// 简单格式化，实际使用时可以根据需要优化
	result := cmd
	for _, arg := range args {
		result += " " + formatArg(arg)
	}
	if len(args) > maxArgs {
		result += " ..."
	}
	return result
}

// formatRedisPipeline 格式化 Redis Pipeline 命令
func formatRedisPipeline(cmds []string) string {
	if len(cmds) == 0 {
		return "empty"
	}

	// 限制命令数量，避免日志过长
	maxCmds := 5
	if len(cmds) > maxCmds {
		cmds = cmds[:maxCmds]
	}

	result := ""
	for i, cmd := range cmds {
		if i > 0 {
			result += " | "
		}
		result += cmd
	}
	if len(cmds) > maxCmds {
		result += " ..."
	}
	return result
}

// formatArg 格式化参数
func formatArg(arg interface{}) string {
	switch v := arg.(type) {
	case string:
		// 限制字符串长度
		if len(v) > 50 {
			return v[:50] + "..."
		}
		return v
	case []byte:
		if len(v) > 50 {
			return string(v[:50]) + "..."
		}
		return string(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// DefaultRedisLogger 返回默认的 Redis logger（Info 级别）
func DefaultRedisLogger() *RedisLogger {
	return NewRedisLogger(RedisLogLevelInfo, 100*time.Millisecond)
}

// SilentRedisLogger 返回静默的 Redis logger（不记录日志）
func SilentRedisLogger() *RedisLogger {
	return NewRedisLogger(RedisLogLevelSilent, 0)
}

// ErrorRedisLogger 返回只记录错误的 Redis logger
func ErrorRedisLogger() *RedisLogger {
	return NewRedisLogger(RedisLogLevelError, 0)
}

// WarnRedisLogger 返回记录警告和错误的 Redis logger
func WarnRedisLogger() *RedisLogger {
	return NewRedisLogger(RedisLogLevelWarn, 100*time.Millisecond)
}

// InfoRedisLogger 返回记录所有日志的 Redis logger
func InfoRedisLogger() *RedisLogger {
	return NewRedisLogger(RedisLogLevelInfo, 100*time.Millisecond)
}
