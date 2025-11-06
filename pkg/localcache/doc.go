// Package localcache 提供了一个高性能的本地缓存实现
//
// 特性：
//   - 基于 LRU（Least Recently Used）算法的缓存淘汰策略
//   - 支持分片（Slot）机制，降低锁竞争，提高并发性能
//   - 支持键关联（Link）功能，可以建立键之间的关联关系，支持级联删除
//   - 支持两种过期策略：主动过期（Expiration）和懒删除（Lazy）
//   - 支持批量操作（GetBatch）
//   - 内置统计功能（Target），可以监控缓存命中率等指标
//
// 基本使用：
//
//	// 创建缓存实例
//	cache := localcache.New[string](
//		localcache.WithLocalSlotNum(500),
//		localcache.WithLocalSlotSize(20000),
//		localcache.WithLocalSuccessTTL(time.Minute),
//	)
//
//	// 获取值（如果缓存未命中，会调用 fetch 函数）
//	value, err := cache.Get(ctx, "key", func(ctx context.Context) (string, error) {
//		// 从数据库或其他数据源获取数据
//		return "value", nil
//	})
//
//	// 建立键关联并获取值
//	value, err := cache.GetLink(ctx, "user:123", fetch, "user:123:profile", "user:123:settings")
//
//	// 删除键（会级联删除关联的键）
//	cache.Del(ctx, "user:123")
//
//	// 仅删除本地缓存
//	cache.DelLocal(ctx, "user:123")
//
//	// 停止缓存
//	cache.Stop()
//
// 配置选项：
//
//	WithLocalSlotNum(n)      - 设置本地缓存分片数量（默认：500）
//	WithLocalSlotSize(n)     - 设置每个分片的容量（默认：20000）
//	WithLinkSlotNum(n)       - 设置键关联分片数量（默认：500）
//	WithLocalSuccessTTL(d)   - 设置成功获取的数据的 TTL（默认：1分钟）
//	WithLocalFailedTTL(d)    - 设置获取失败的数据的 TTL（默认：5秒）
//	WithExpirationEvict()    - 使用主动过期策略
//	WithLazy()               - 使用懒删除策略（默认）
//	WithLocalDisable()       - 禁用本地缓存
//	WithLinkDisable()        - 禁用键关联功能
//	WithTarget(target)       - 设置统计目标
//	WithDeleteKeyBefore(fn)  - 设置删除前的回调函数
//
// LRU 实现：
//
// 包提供了两种 LRU 实现：
//   - ExpirationLRU: 基于 expirable.LRU，支持主动过期清理
//   - LazyLRU: 基于 simplelru.LRU，使用懒删除策略
//
// 键关联（Link）功能：
//
// 键关联功能允许建立键之间的双向关联关系。当删除一个键时，会自动删除所有关联的键。
// 这对于缓存相关的数据非常有用，例如用户信息和用户配置。
//
//	// 建立关联：user:123 <-> user:123:profile, user:123:settings
//	cache.GetLink(ctx, "user:123", fetch, "user:123:profile", "user:123:settings")
//
//	// 删除 user:123 时，会自动删除 user:123:profile 和 user:123:settings
//	cache.Del(ctx, "user:123")
//
// 统计功能：
//
// 通过实现 lru.Target 接口，可以监控缓存的性能指标：
//   - IncrGetHit(): 缓存命中
//   - IncrGetSuccess(): 获取成功（包括缓存未命中但 fetch 成功）
//   - IncrGetFailed(): 获取失败
//   - IncrDelHit(): 删除命中
//   - IncrDelNotFound(): 删除未找到
//
// 示例：
//
//	type StatsTarget struct {
//		hits, misses, errors int64
//	}
//
//	func (s *StatsTarget) IncrGetHit() { atomic.AddInt64(&s.hits, 1) }
//	func (s *StatsTarget) IncrGetSuccess() { atomic.AddInt64(&s.misses, 1) }
//	func (s *StatsTarget) IncrGetFailed() { atomic.AddInt64(&s.errors, 1) }
//	func (s *StatsTarget) IncrDelHit() {}
//	func (s *StatsTarget) IncrDelNotFound() {}
//
//	cache := localcache.New[string](
//		localcache.WithTarget(&StatsTarget{}),
//	)
package localcache
