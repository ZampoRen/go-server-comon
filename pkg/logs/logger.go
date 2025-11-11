package logger

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	hertzzap "github.com/hertz-contrib/logger/zap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	// defaultLogger 默认的 logger 实例
	defaultLogger *Logger
)

// RotateConfig 日志切割配置
type RotateConfig struct {
	// Filename 日志文件路径
	Filename string
	// MaxSize 单个日志文件最大大小（MB），默认 20MB
	MaxSize int
	// MaxBackups 保留的旧日志文件最大数量，默认 5
	MaxBackups int
	// MaxAge 保留的旧日志文件最大天数，默认 10 天
	MaxAge int
	// Compress 是否压缩旧日志文件，默认 true
	Compress bool
	// AlsoStdout 是否同时输出到 stdout，默认 false
	AlsoStdout bool
}

// Logger wraps logging functionality using hertz hlog with zap
type Logger struct {
	zapLogger *zap.Logger
	hlog      hlog.FullLogger
}

// Init 初始化 logger，使用 zap 作为底层实现
// level: 日志级别，可选值: debug, info, warn, error
// outputPaths: 日志输出路径，如 []string{"stdout", "/var/log/app.log"}
func Init(level string, outputPaths []string) error {
	// 解析日志级别
	var hlogLevel hlog.Level
	switch level {
	case "debug":
		hlogLevel = hlog.LevelDebug
	case "info":
		hlogLevel = hlog.LevelInfo
	case "warn":
		hlogLevel = hlog.LevelWarn
	case "error":
		hlogLevel = hlog.LevelError
	default:
		hlogLevel = hlog.LevelInfo
	}

	// 配置 zap encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// 解析日志级别用于 zap
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	// 创建 zap config
	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(zapLevel),
		Development:      false,
		Encoding:         "console", // 或 "json"
		EncoderConfig:    encoderConfig,
		OutputPaths:      outputPaths,
		ErrorOutputPaths: outputPaths,
	}

	// 构建 zap logger
	zapLogger, err := config.Build()
	if err != nil {
		return err
	}

	// 使用 hertz-contrib/logger/zap 创建 logger
	// 参考示例代码，添加 caller skip 以正确显示调用位置
	hertzLogger := hertzzap.NewLogger(
		hertzzap.WithZapOptions(
			zap.AddCaller(),
			zap.AddCallerSkip(3),
			zap.WithFatalHook(zapcore.WriteThenPanic),
		),
	)
	hertzLogger.SetLevel(hlogLevel)

	// 使用 hlog 设置 zap logger
	hlog.SetLogger(hertzLogger)

	// 创建默认 logger 实例
	defaultLogger = &Logger{
		zapLogger: zapLogger,
		hlog:      hertzLogger,
	}

	return nil
}

// InitWithZap 使用自定义的 zap logger 初始化
func InitWithZap(zapLogger *zap.Logger) {
	hertzLogger := hertzzap.NewLogger(
		hertzzap.WithZapOptions(
			zap.AddCaller(),
			zap.AddCallerSkip(3),
			zap.WithFatalHook(zapcore.WriteThenPanic),
		),
	)
	hertzLogger.SetLevel(hlog.LevelDebug)
	hlog.SetLogger(hertzLogger)
	defaultLogger = &Logger{
		zapLogger: zapLogger,
		hlog:      hertzLogger,
	}
}

// InitWithOptions 使用自定义选项初始化 logger
// output: 日志输出，可以是文件或 stdout
func InitWithOptions(level string, output io.Writer) error {
	// 解析日志级别
	var hlogLevel hlog.Level
	switch level {
	case "debug":
		hlogLevel = hlog.LevelDebug
	case "info":
		hlogLevel = hlog.LevelInfo
	case "warn":
		hlogLevel = hlog.LevelWarn
	case "error":
		hlogLevel = hlog.LevelError
	default:
		hlogLevel = hlog.LevelInfo
	}

	// 使用 hertz-contrib/logger/zap 创建 logger
	// 参考示例代码，添加 caller skip 以正确显示调用位置
	hertzLogger := hertzzap.NewLogger(
		hertzzap.WithZapOptions(
			zap.AddCaller(),
			zap.AddCallerSkip(3),
			zap.WithFatalHook(zapcore.WriteThenPanic),
		),
	)
	hertzLogger.SetLevel(hlogLevel)
	hertzLogger.SetOutput(output)

	// 使用 hlog 设置 zap logger
	hlog.SetLogger(hertzLogger)

	// 创建默认 logger 实例
	defaultLogger = &Logger{
		zapLogger: nil, // 使用 hertz logger 时不需要直接访问 zap logger
		hlog:      hertzLogger,
	}

	return nil
}

// InitWithRotate 使用日志切割功能初始化 logger
// level: 日志级别，可选值: debug, info, warn, error
// config: 日志切割配置
func InitWithRotate(level string, config *RotateConfig) error {
	// 设置默认值
	if config.MaxSize == 0 {
		config.MaxSize = 20 // 默认 20MB
	}
	if config.MaxBackups == 0 {
		config.MaxBackups = 5 // 默认保留 5 个文件
	}
	if config.MaxAge == 0 {
		config.MaxAge = 10 // 默认保留 10 天
	}
	if !config.Compress {
		config.Compress = true // 默认压缩
	}

	// 确保日志目录存在
	if config.Filename != "" {
		dir := filepath.Dir(config.Filename)
		if err := os.MkdirAll(dir, 0o777); err != nil {
			return err
		}
	}

	// 解析日志级别
	var hlogLevel hlog.Level
	switch level {
	case "debug":
		hlogLevel = hlog.LevelDebug
	case "info":
		hlogLevel = hlog.LevelInfo
	case "warn":
		hlogLevel = hlog.LevelWarn
	case "error":
		hlogLevel = hlog.LevelError
	default:
		hlogLevel = hlog.LevelInfo
	}

	// 创建 lumberjack logger 用于日志切割
	var lumberjackLogger *lumberjack.Logger
	if config.Filename != "" {
		lumberjackLogger = &lumberjack.Logger{
			Filename:   config.Filename,
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   config.Compress,
		}
	}

	// 确定输出目标
	var output io.Writer
	if lumberjackLogger != nil && config.AlsoStdout {
		// 同时输出到文件和 stdout
		output = io.MultiWriter(lumberjackLogger, os.Stdout)
	} else if lumberjackLogger != nil {
		// 只输出到文件
		output = lumberjackLogger
	} else {
		// 只输出到 stdout
		output = os.Stdout
	}

	// 使用 hertz-contrib/logger/zap 创建 logger
	// 参考示例代码，添加 caller skip 以正确显示调用位置
	hertzLogger := hertzzap.NewLogger(
		hertzzap.WithZapOptions(
			zap.AddCaller(),
			zap.AddCallerSkip(3),
			zap.WithFatalHook(zapcore.WriteThenPanic),
		),
	)
	hertzLogger.SetLevel(hlogLevel)
	hertzLogger.SetOutput(output)

	// 使用 hlog 设置 zap logger
	hlog.SetLogger(hertzLogger)

	// 创建默认 logger 实例
	defaultLogger = &Logger{
		zapLogger: nil,
		hlog:      hertzLogger,
	}

	return nil
}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	if defaultLogger == nil {
		// 如果没有初始化，使用默认配置
		hertzLogger := hertzzap.NewLogger(
			hertzzap.WithZapOptions(
				zap.AddCaller(),
				zap.AddCallerSkip(3),
			),
		)
		hertzLogger.SetLevel(hlog.LevelInfo)
		hlog.SetLogger(hertzLogger)
		defaultLogger = &Logger{
			zapLogger: nil,
			hlog:      hertzLogger,
		}
	}
	return defaultLogger
}

// Default 返回默认的 logger 实例
func Default() *Logger {
	return NewLogger()
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	hlog.Info(msg)
}

// Infof logs an info message with format
func (l *Logger) Infof(format string, args ...interface{}) {
	hlog.Infof(format, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string, err error) {
	if err != nil {
		hlog.Errorf("%s: %v", msg, err)
	} else {
		hlog.Error(msg)
	}
}

// Errorf logs an error message with format
func (l *Logger) Errorf(format string, args ...interface{}) {
	hlog.Errorf(format, args...)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	hlog.Debug(msg)
}

// Debugf logs a debug message with format
func (l *Logger) Debugf(format string, args ...interface{}) {
	hlog.Debugf(format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	hlog.Warn(msg)
}

// Warnf logs a warning message with format
func (l *Logger) Warnf(format string, args ...interface{}) {
	hlog.Warnf(format, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string) {
	hlog.Fatal(msg)
}

// Fatalf logs a fatal message with format and exits
func (l *Logger) Fatalf(format string, args ...interface{}) {
	hlog.Fatalf(format, args...)
}

// WithContext 返回带上下文的 logger
func (l *Logger) WithContext(ctx context.Context) *Logger {
	return l
}

// Sync 同步日志缓冲区
func (l *Logger) Sync() error {
	if l.zapLogger != nil {
		return l.zapLogger.Sync()
	}
	return nil
}
