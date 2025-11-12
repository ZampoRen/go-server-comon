package cache

import (
	"context"
	"time"
)

// Nil 默认的 nil 错误
var Nil error

// SetDefaultNilError 设置默认的 nil 错误
func SetDefaultNilError(err error) {
	Nil = err
}

// Cmdable 可执行命令的接口
type Cmdable interface {
	StringCmdable
	HashCmdable
	GenericCmdable
	ListCmdable
	Pipeline() Pipeliner
}

// StringCmdable 字符串命令接口
type StringCmdable interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) StatusCmd
	Get(ctx context.Context, key string) StringCmd
	IncrBy(ctx context.Context, key string, value int64) IntCmd
	Incr(ctx context.Context, key string) IntCmd
}

// HashCmdable 哈希命令接口
type HashCmdable interface {
	HSet(ctx context.Context, key string, values ...interface{}) IntCmd
	HGetAll(ctx context.Context, key string) MapStringStringCmd
}

// GenericCmdable 通用命令接口
type GenericCmdable interface {
	Del(ctx context.Context, keys ...string) IntCmd
	Exists(ctx context.Context, keys ...string) IntCmd
	Expire(ctx context.Context, key string, expiration time.Duration) BoolCmd
}

// Pipeliner 管道接口
type Pipeliner interface {
	StringCmdable
	HashCmdable
	GenericCmdable
	ListCmdable
	Exec(ctx context.Context) ([]Cmder, error)
}

// ListCmdable 列表命令接口
type ListCmdable interface {
	LIndex(ctx context.Context, key string, index int64) StringCmd
	LPush(ctx context.Context, key string, values ...interface{}) IntCmd
	RPush(ctx context.Context, key string, values ...interface{}) IntCmd
	LSet(ctx context.Context, key string, index int64, value interface{}) StatusCmd
	LPop(ctx context.Context, key string) StringCmd
	LRange(ctx context.Context, key string, start, stop int64) StringSliceCmd
}

// Cmder 命令接口
type Cmder interface {
	Err() error
}

// baseCmd 基础命令接口
type baseCmd interface {
	Err() error
}

// IntCmd 整数命令接口
type IntCmd interface {
	baseCmd
	Result() (int64, error)
}

// MapStringStringCmd 字符串映射命令接口
type MapStringStringCmd interface {
	baseCmd
	Result() (map[string]string, error)
}

// BoolCmd 布尔命令接口
type BoolCmd interface {
	baseCmd
	Result() (bool, error)
}

// StatusCmd 状态命令接口
type StatusCmd interface {
	baseCmd
	Result() (string, error)
}

// StringCmd 字符串命令接口
type StringCmd interface {
	baseCmd
	Result() (string, error)
	Val() string
	Int64() (int64, error)
	Bytes() ([]byte, error)
}

// StringSliceCmd 字符串切片命令接口
type StringSliceCmd interface {
	baseCmd
	Result() ([]string, error)
}
