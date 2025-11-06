package localcache

import (
	"context"
	"hash/fnv"
	"sync"

	"github.com/ZampoRen/go-server-comon/pkg/localcache/link"
	"github.com/ZampoRen/go-server-comon/pkg/localcache/lru"
)

type Cache[V any] interface {
	Get(ctx context.Context, key string, fetch func(ctx context.Context) (V, error)) (V, error)
	GetLink(ctx context.Context, key string, fetch func(ctx context.Context) (V, error), link ...string) (V, error)
	Del(ctx context.Context, key ...string)
	DelLocal(ctx context.Context, key ...string)
	Stop()
}

func LRUStringHash(key string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(key))
	return h.Sum64()
}

func New[V any](opts ...Option) Cache[V] {
	opt := defaultOption()
	for _, o := range opts {
		o(opt)
	}

	c := cache[V]{
		opt:        opt,
		pendingDel: make(chan []string, 100), // 缓冲队列，避免阻塞
	}
	if opt.localSlotNum > 0 && opt.localSlotSize > 0 {
		createSimpleLRU := func() lru.LRU[string, V] {
			if opt.expirationEvict {
				return lru.NewExpirationLRU(opt.localSlotSize, opt.localSuccessTTL, opt.localFailedTTL, opt.target, c.onEvict)
			} else {
				return lru.NewLazyLRU(opt.localSlotSize, opt.localSuccessTTL, opt.localFailedTTL, opt.target, c.onEvict)
			}
		}
		if opt.localSlotNum == 1 {
			c.local = createSimpleLRU()
		} else {
			c.local = lru.NewSlotLRU(opt.localSlotNum, LRUStringHash, createSimpleLRU)
		}
		if opt.linkSlotNum > 0 {
			c.link = link.New(opt.linkSlotNum)
		}
	}
	return &c
}

type cache[V any] struct {
	opt        *option
	link       link.Link
	local      lru.LRU[string, V]
	pendingDel chan []string // 待删除的键队列
	once       sync.Once     // 确保只启动一次清理 goroutine
	stopOnce   sync.Once     // 确保只关闭一次 channel
}

func (c *cache[V]) onEvict(key string, value V) {
	_ = value

	// onEvict 在 LRU 自动淘汰时被调用，此时 LRU 的锁已经被持有
	// 我们不能在这里直接调用 c.local.Del，否则会导致死锁
	// 所以先获取关联键，然后异步删除
	if c.link != nil {
		linkedKeys := c.link.Del(key)
		if len(linkedKeys) > 0 {
			// 将关联键转换为切片
			keys := make([]string, 0, len(linkedKeys))
			for k := range linkedKeys {
				if k != key { // 避免删除自己
					keys = append(keys, k)
				}
			}
			if len(keys) > 0 {
				// 启动清理 goroutine（只启动一次）
				c.once.Do(func() {
					go c.processPendingDeletes()
				})
				// 异步发送待删除的键（非阻塞）
				select {
				case c.pendingDel <- keys:
				default:
					// 如果队列满了，忽略（避免阻塞）
				}
			}
		}
	}
}

// processPendingDeletes 处理待删除的键
func (c *cache[V]) processPendingDeletes() {
	for keys := range c.pendingDel {
		for _, key := range keys {
			c.local.Del(key)
		}
	}
}

func (c *cache[V]) del(key ...string) {
	if c.local == nil {
		return
	}
	// 使用 map 记录已删除的键，避免重复删除
	deleted := make(map[string]struct{})

	// 待删除的键队列
	toDelete := make([]string, 0, len(key))
	toDelete = append(toDelete, key...)

	for len(toDelete) > 0 {
		curr := toDelete[0]
		toDelete = toDelete[1:]

		// 如果已经删除过，跳过
		if _, ok := deleted[curr]; ok {
			continue
		}

		// 标记为已删除
		deleted[curr] = struct{}{}

		// 获取关联键（在删除之前）
		var linkedKeys map[string]struct{}
		if c.link != nil {
			linkedKeys = c.link.Del(curr)
			// 将关联键加入待删除队列
			for k := range linkedKeys {
				if _, ok := deleted[k]; !ok {
					toDelete = append(toDelete, k)
				}
			}
		}

		// 删除本地缓存
		// 注意：此时关联关系已经被清理，onEvict 不会获取到关联键
		// 但这是手动删除，我们已经在上面处理了关联键，所以不需要依赖 onEvict
		c.local.Del(curr)
	}
}

func (c *cache[V]) Get(ctx context.Context, key string, fetch func(ctx context.Context) (V, error)) (V, error) {
	return c.GetLink(ctx, key, fetch)
}

func (c *cache[V]) GetLink(ctx context.Context, key string, fetch func(ctx context.Context) (V, error), link ...string) (V, error) {
	if c.local != nil {
		return c.local.Get(key, func() (V, error) {
			if len(link) > 0 && c.link != nil {
				c.link.Link(key, link...)
			}
			return fetch(ctx)
		})
	} else {
		return fetch(ctx)
	}
}

func (c *cache[V]) Del(ctx context.Context, key ...string) {
	for _, fn := range c.opt.delFn {
		fn(ctx, key...)
	}
	c.del(key...)
}

func (c *cache[V]) DelLocal(ctx context.Context, key ...string) {
	c.del(key...)
}

func (c *cache[V]) Stop() {
	if c.local != nil {
		c.local.Stop()
	}
	// 关闭待删除队列（只关闭一次）
	if c.pendingDel != nil {
		c.stopOnce.Do(func() {
			close(c.pendingDel)
		})
	}
}
