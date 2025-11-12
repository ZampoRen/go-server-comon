package redis

import (
	"context"
	"os"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/ZampoRen/go-server-comon/internal/infra/cache"
	"github.com/ZampoRen/go-server-comon/pkg/envkey"
)

// Cmdable 命令接口类型别名
type Cmdable = cache.Cmdable

// New 创建新的 Redis 客户端，从环境变量读取配置
// 环境变量：
//   - REDIS_ADDR: Redis 地址（必需）
//   - REDIS_PASSWORD: Redis 密码
//   - REDIS_DB: Redis 数据库编号（默认 0）
//   - REDIS_POOL_SIZE: 最大连接数（默认 100）
//   - REDIS_MIN_IDLE_CONNS: 最小空闲连接数（默认 10）
//   - REDIS_MAX_IDLE_CONNS: 最大空闲连接数（默认 30）
//   - REDIS_CONN_MAX_IDLE_TIME: 空闲连接超时时间（默认 5m，格式如 "5m", "10m"）
//   - REDIS_DIAL_TIMEOUT: 连接建立超时（默认 5s，格式如 "5s", "10s"）
//   - REDIS_READ_TIMEOUT: 读操作超时（默认 3s，格式如 "3s", "5s"）
//   - REDIS_WRITE_TIMEOUT: 写操作超时（默认 3s，格式如 "3s", "5s"）
func New() cache.Cmdable {
	addr := os.Getenv("REDIS_ADDR")
	password := os.Getenv("REDIS_PASSWORD")

	return NewWithAddrAndPassword(addr, password)
}

// NewWithAddrAndPassword 使用指定的地址和密码创建 Redis 客户端
// 连接池和超时配置从环境变量读取，如果没有设置则使用默认值
func NewWithAddrAndPassword(addr, password string) cache.Cmdable {
	cache.SetDefaultNilError(redis.Nil)

	// 从环境变量读取数据库编号（默认 0）
	db := envkey.GetIntD("REDIS_DB", 0)

	// 从环境变量读取连接池配置
	poolSize := envkey.GetIntD("REDIS_POOL_SIZE", 100)
	minIdleConns := envkey.GetIntD("REDIS_MIN_IDLE_CONNS", 10)
	maxIdleConns := envkey.GetIntD("REDIS_MAX_IDLE_CONNS", 30)

	// 从环境变量读取连接最大空闲时间（默认 5 分钟）
	connMaxIdleTimeStr := envkey.GetStringD("REDIS_CONN_MAX_IDLE_TIME", "5m")
	connMaxIdleTime, err := time.ParseDuration(connMaxIdleTimeStr)
	if err != nil {
		// 如果解析失败，使用默认值 5 分钟
		connMaxIdleTime = 5 * time.Minute
	}

	// 从环境变量读取超时配置
	dialTimeoutStr := envkey.GetStringD("REDIS_DIAL_TIMEOUT", "5s")
	dialTimeout, err := time.ParseDuration(dialTimeoutStr)
	if err != nil {
		dialTimeout = 5 * time.Second
	}

	readTimeoutStr := envkey.GetStringD("REDIS_READ_TIMEOUT", "3s")
	readTimeout, err := time.ParseDuration(readTimeoutStr)
	if err != nil {
		readTimeout = 3 * time.Second
	}

	writeTimeoutStr := envkey.GetStringD("REDIS_WRITE_TIMEOUT", "3s")
	writeTimeout, err := time.ParseDuration(writeTimeoutStr)
	if err != nil {
		writeTimeout = 3 * time.Second
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,     // Redis 地址
		DB:       db,       // 数据库编号
		Password: password, // Redis 密码
		// 连接池配置
		PoolSize:        poolSize,        // 最大连接数（建议设置为 CPU 核心数 * 10）
		MinIdleConns:    minIdleConns,    // 最小空闲连接数
		MaxIdleConns:    maxIdleConns,    // 最大空闲连接数
		ConnMaxIdleTime: connMaxIdleTime, // 空闲连接超时时间

		// 超时配置
		DialTimeout:  dialTimeout,  // 连接建立超时
		ReadTimeout:  readTimeout,  // 读操作超时
		WriteTimeout: writeTimeout, // 写操作超时
	})

	return &redisImpl{client: rdb}
}

// redisImpl Redis 实现
type redisImpl struct {
	client *redis.Client
}

// Del 删除指定的键
func (r *redisImpl) Del(ctx context.Context, keys ...string) cache.IntCmd {
	return r.client.Del(ctx, keys...)
}

// Exists 检查指定的键是否存在
func (r *redisImpl) Exists(ctx context.Context, keys ...string) cache.IntCmd {
	return r.client.Exists(ctx, keys...)
}

// Expire 设置键的过期时间
func (r *redisImpl) Expire(ctx context.Context, key string, expiration time.Duration) cache.BoolCmd {
	return r.client.Expire(ctx, key, expiration)
}

// Get 获取指定键的值
func (r *redisImpl) Get(ctx context.Context, key string) cache.StringCmd {
	return r.client.Get(ctx, key)
}

// HGetAll 获取哈希表的所有字段和值
func (r *redisImpl) HGetAll(ctx context.Context, key string) cache.MapStringStringCmd {
	return r.client.HGetAll(ctx, key)
}

// HSet 设置哈希表的字段值
func (r *redisImpl) HSet(ctx context.Context, key string, values ...interface{}) cache.IntCmd {
	return r.client.HSet(ctx, key, values...)
}

// Incr 将键的值增加 1
func (r *redisImpl) Incr(ctx context.Context, key string) cache.IntCmd {
	return r.client.Incr(ctx, key)
}

// IncrBy 将键的值增加指定的整数
func (r *redisImpl) IncrBy(ctx context.Context, key string, value int64) cache.IntCmd {
	return r.client.IncrBy(ctx, key, value)
}

// LIndex 获取列表中指定索引的元素
func (r *redisImpl) LIndex(ctx context.Context, key string, index int64) cache.StringCmd {
	return r.client.LIndex(ctx, key, index)
}

// LPop 从列表左侧弹出元素
func (r *redisImpl) LPop(ctx context.Context, key string) cache.StringCmd {
	return r.client.LPop(ctx, key)
}

// LPush 从列表左侧推入元素
func (r *redisImpl) LPush(ctx context.Context, key string, values ...interface{}) cache.IntCmd {
	return r.client.LPush(ctx, key, values...)
}

// LRange 获取列表中指定范围的元素
func (r *redisImpl) LRange(ctx context.Context, key string, start int64, stop int64) cache.StringSliceCmd {
	return r.client.LRange(ctx, key, start, stop)
}

// LSet 设置列表中指定索引的元素值
func (r *redisImpl) LSet(ctx context.Context, key string, index int64, value interface{}) cache.StatusCmd {
	return r.client.LSet(ctx, key, index, value)
}

// Pipeline 创建管道
func (r *redisImpl) Pipeline() cache.Pipeliner {
	p := r.client.Pipeline()
	return &pipelineImpl{p: p}
}

// RPush 从列表右侧推入元素
func (r *redisImpl) RPush(ctx context.Context, key string, values ...interface{}) cache.IntCmd {
	return r.client.RPush(ctx, key, values...)
}

// Set 设置键的值
func (r *redisImpl) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) cache.StatusCmd {
	return r.client.Set(ctx, key, value, expiration)
}

// pipelineImpl 管道实现
type pipelineImpl struct {
	p redis.Pipeliner
}

// Del 删除指定的键
func (p *pipelineImpl) Del(ctx context.Context, keys ...string) cache.IntCmd {
	return p.p.Del(ctx, keys...)
}

// Exec 执行管道中的所有命令
func (p *pipelineImpl) Exec(ctx context.Context) ([]cache.Cmder, error) {
	cmders, err := p.p.Exec(ctx)
	if err != nil {
		return nil, err
	}
	return convertCmders(cmders), nil
}

// convertCmders 转换命令列表
func convertCmders(cmders []redis.Cmder) []cache.Cmder {
	res := make([]cache.Cmder, 0, len(cmders))
	for _, cmder := range cmders {
		res = append(res, &cmderImpl{cmder: cmder})
	}
	return res
}

// cmderImpl 命令实现
type cmderImpl struct {
	cmder redis.Cmder
}

// Err 返回命令的错误
func (c *cmderImpl) Err() error {
	return c.cmder.Err()
}

// Exists 检查指定的键是否存在
func (p *pipelineImpl) Exists(ctx context.Context, keys ...string) cache.IntCmd {
	return p.p.Exists(ctx, keys...)
}

// Expire 设置键的过期时间
func (p *pipelineImpl) Expire(ctx context.Context, key string, expiration time.Duration) cache.BoolCmd {
	return p.p.Expire(ctx, key, expiration)
}

// Get 获取指定键的值
func (p *pipelineImpl) Get(ctx context.Context, key string) cache.StringCmd {
	return p.p.Get(ctx, key)
}

// HGetAll 获取哈希表的所有字段和值
func (p *pipelineImpl) HGetAll(ctx context.Context, key string) cache.MapStringStringCmd {
	return p.p.HGetAll(ctx, key)
}

// HSet 设置哈希表的字段值
func (p *pipelineImpl) HSet(ctx context.Context, key string, values ...interface{}) cache.IntCmd {
	return p.p.HSet(ctx, key, values...)
}

// Incr 将键的值增加 1
func (p *pipelineImpl) Incr(ctx context.Context, key string) cache.IntCmd {
	return p.p.Incr(ctx, key)
}

// IncrBy 将键的值增加指定的整数
func (p *pipelineImpl) IncrBy(ctx context.Context, key string, value int64) cache.IntCmd {
	return p.p.IncrBy(ctx, key, value)
}

// LIndex 获取列表中指定索引的元素
func (p *pipelineImpl) LIndex(ctx context.Context, key string, index int64) cache.StringCmd {
	return p.p.LIndex(ctx, key, index)
}

// LPop 从列表左侧弹出元素
func (p *pipelineImpl) LPop(ctx context.Context, key string) cache.StringCmd {
	return p.p.LPop(ctx, key)
}

// LPush 从列表左侧推入元素
func (p *pipelineImpl) LPush(ctx context.Context, key string, values ...interface{}) cache.IntCmd {
	return p.p.LPush(ctx, key, values...)
}

// LRange 获取列表中指定范围的元素
func (p *pipelineImpl) LRange(ctx context.Context, key string, start int64, stop int64) cache.StringSliceCmd {
	return p.p.LRange(ctx, key, start, stop)
}

// LSet 设置列表中指定索引的元素值
func (p *pipelineImpl) LSet(ctx context.Context, key string, index int64, value interface{}) cache.StatusCmd {
	return p.p.LSet(ctx, key, index, value)
}

// Pipeline 创建管道
func (p *pipelineImpl) Pipeline() cache.Pipeliner {
	return p
}

// RPush 从列表右侧推入元素
func (p *pipelineImpl) RPush(ctx context.Context, key string, values ...interface{}) cache.IntCmd {
	return p.p.RPush(ctx, key, values...)
}

// Set 设置键的值
func (p *pipelineImpl) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) cache.StatusCmd {
	return p.p.Set(ctx, key, value, expiration)
}
