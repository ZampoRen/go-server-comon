package mysql

import (
	"fmt"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/ZampoRen/go-server-comon/pkg/envkey"
	logger "github.com/ZampoRen/go-server-comon/pkg/logs"
)

// Config GORM 配置选项
type Config struct {
	// DSN 数据库连接字符串，如果为空则从环境变量 MYSQL_DSN 读取
	DSN string
	// LogLevel 日志级别，可选值: silent, error, warn, info
	// 如果为空，默认使用 info
	LogLevel string
	// SlowThreshold 慢查询阈值，默认 200ms
	SlowThreshold time.Duration
	// IgnoreRecordNotFoundError 是否忽略记录未找到错误，默认 true
	IgnoreRecordNotFoundError bool
	// GormConfig 自定义 GORM 配置，如果提供则优先使用此配置
	GormConfig *gorm.Config
}

// New 创建新的 MySQL 数据库连接，使用默认配置和 sql_logger
func New() (*gorm.DB, error) {
	return NewWithOptions(nil)
}

// NewWithDSN 使用指定的 DSN 创建数据库连接
func NewWithDSN(dsn string) (*gorm.DB, error) {
	config := &Config{
		DSN: dsn,
	}
	return NewWithOptions(config)
}

// NewWithOptions 使用配置选项创建数据库连接
func NewWithOptions(config *Config) (*gorm.DB, error) {
	// 设置默认值
	if config == nil {
		config = &Config{}
	}

	// 获取 DSN
	dsn := config.DSN
	if dsn == "" {
		dsn = os.Getenv("MYSQL_DSN")
	}
	if dsn == "" {
		return nil, fmt.Errorf("mysql dsn is required, set MYSQL_DSN environment variable or provide DSN in config")
	}

	// 构建 GORM 配置
	var gormConfig *gorm.Config
	if config.GormConfig != nil {
		// 使用用户提供的配置
		gormConfig = config.GormConfig
		// 如果用户没有设置 Logger，则使用我们的 sql_logger
		if gormConfig.Logger == nil {
			gormConfig.Logger = buildGormLogger(config)
		}
	} else {
		// 使用默认配置，并设置 sql_logger
		gormConfig = &gorm.Config{
			Logger: buildGormLogger(config),
		}
	}

	// 打开数据库连接
	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("mysql open, dsn: %s, err: %w", dsn, err)
	}

	// 配置连接池和超时设置
	if err := configureConnectionPool(db); err != nil {
		return nil, fmt.Errorf("configure connection pool failed: %w", err)
	}

	return db, nil
}

// NewWithConfig 使用自定义 GORM 配置创建数据库连接
// 如果 config.Logger 为空，则使用默认的 sql_logger
func NewWithConfig(dsn string, gormConfig *gorm.Config) (*gorm.DB, error) {
	if dsn == "" {
		dsn = os.Getenv("MYSQL_DSN")
	}
	if dsn == "" {
		return nil, fmt.Errorf("mysql dsn is required, set MYSQL_DSN environment variable or provide DSN parameter")
	}

	// 如果用户没有设置 Logger，则使用默认的 sql_logger
	if gormConfig == nil {
		gormConfig = &gorm.Config{}
	}
	if gormConfig.Logger == nil {
		// 使用 pkg/logs 包（包名是 logger）的默认 GORM logger
		gormConfig.Logger = logger.DefaultGormLogger()
	}

	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("mysql open, dsn: %s, err: %w", dsn, err)
	}

	// 配置连接池和超时设置
	if err := configureConnectionPool(db); err != nil {
		return nil, fmt.Errorf("configure connection pool failed: %w", err)
	}

	return db, nil
}

// buildGormLogger 根据配置构建 GORM logger
func buildGormLogger(config *Config) gormlogger.Interface {
	// 解析日志级别
	var logLevel gormlogger.LogLevel
	switch config.LogLevel {
	case "silent":
		logLevel = gormlogger.Silent
	case "error":
		logLevel = gormlogger.Error
	case "warn":
		logLevel = gormlogger.Warn
	case "info":
		logLevel = gormlogger.Info
	default:
		logLevel = gormlogger.Info
	}

	// 设置慢查询阈值
	slowThreshold := config.SlowThreshold
	if slowThreshold == 0 {
		slowThreshold = 200 * time.Millisecond
	}

	// 创建 logger（使用 pkg/logs 包，包名是 logger）
	gormLogger := logger.NewGormLogger(logLevel, slowThreshold)
	// 如果用户明确设置了 IgnoreRecordNotFoundError，则使用该值
	// 否则使用默认值 true（NewGormLogger 已设置）
	if !config.IgnoreRecordNotFoundError {
		gormLogger.IgnoreRecordNotFoundError = false
	}

	return gormLogger
}

// configureConnectionPool 配置数据库连接池和超时设置
// 从环境变量读取配置，如果没有设置则使用默认值
func configureConnectionPool(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	// 最大打开连接数（默认 100）
	maxOpenConns := envkey.GetIntD("MYSQL_MAX_OPEN_CONNS", 100)
	sqlDB.SetMaxOpenConns(maxOpenConns)

	// 最大空闲连接数（默认 10）
	maxIdleConns := envkey.GetIntD("MYSQL_MAX_IDLE_CONNS", 10)
	sqlDB.SetMaxIdleConns(maxIdleConns)

	// 连接最大生存时间（默认 1 小时）
	connMaxLifetimeStr := envkey.GetStringD("MYSQL_CONN_MAX_LIFETIME", "1h")
	connMaxLifetime, err := time.ParseDuration(connMaxLifetimeStr)
	if err != nil {
		// 如果解析失败，使用默认值 1 小时
		connMaxLifetime = time.Hour
	}
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	// 连接最大空闲时间（默认 10 分钟）
	connMaxIdleTimeStr := envkey.GetStringD("MYSQL_CONN_MAX_IDLE_TIME", "10m")
	connMaxIdleTime, err := time.ParseDuration(connMaxIdleTimeStr)
	if err != nil {
		// 如果解析失败，使用默认值 10 分钟
		connMaxIdleTime = 10 * time.Minute
	}
	sqlDB.SetConnMaxIdleTime(connMaxIdleTime)

	return nil
}
