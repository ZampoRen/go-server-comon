package lru

import (
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
)

type expirationLruItem[V any] struct {
	lock  sync.RWMutex
	err   error
	value V
}

type ExpirationLRU[K comparable, V any] struct {
	lock       sync.Mutex
	core       *expirable.LRU[K, *expirationLruItem[V]]
	successTTL time.Duration
	failedTTL  time.Duration
	target     Target
}

func NewExpirationLRU[K comparable, V any](size int, successTTL, failedTTL time.Duration, target Target, onEvict EvictCallback[K, V]) LRU[K, V] {
	var cb expirable.EvictCallback[K, *expirationLruItem[V]]
	if onEvict != nil {
		cb = func(key K, value *expirationLruItem[V]) {
			onEvict(key, value.value)
		}
	}
	core := expirable.NewLRU(size, cb, successTTL)
	return &ExpirationLRU[K, V]{
		core:       core,
		successTTL: successTTL,
		failedTTL:  failedTTL,
		target:     target,
	}
}

func (x *ExpirationLRU[K, V]) GetBatch(keys []K, fetch func(keys []K) (map[K]V, error)) (map[K]V, error) {
	var (
		err  error
		once sync.Once
	)

	res := make(map[K]V)
	queries := make([]K, 0, len(keys))

	// 第一遍：检查缓存中已有的 key
	for _, key := range keys {
		x.lock.Lock()
		v, ok := x.core.Get(key)
		x.lock.Unlock()
		if ok {
			// 如果 key 存在，说明未过期（expirable.LRU 会自动清理过期项）
			v.lock.RLock()
			value, err1 := v.value, v.err
			v.lock.RUnlock()

			x.target.IncrGetHit()
			res[key] = value

			// 如果有错误，记录第一个错误
			if err1 != nil {
				once.Do(func() {
					err = err1
				})
			}
			continue
		}
		// 缓存未命中，需要查询
		queries = append(queries, key)
	}

	// 如果所有 key 都命中缓存，直接返回
	if len(queries) == 0 {
		return res, err
	}

	// 批量获取缺失的 key
	values, fetchErr := fetch(queries)
	if fetchErr != nil {
		once.Do(func() {
			err = fetchErr
		})
	}

	// 将获取到的值添加到缓存
	x.lock.Lock()
	defer x.lock.Unlock()

	for _, key := range queries {
		val, exists := values[key]
		if exists {
			// 成功获取到值
			v := &expirationLruItem[V]{
				value: val,
				err:   nil,
			}
			x.core.Add(key, v)
			res[key] = val
			x.target.IncrGetSuccess()
		} else {
			// 如果 fetch 返回了错误，或者某个 key 不在结果中
			// 对于失败的项，不缓存（与 Get 方法保持一致）
			if err == nil {
				// 如果没有全局错误，但某个 key 不存在，记录为失败
				x.target.IncrGetFailed()
			}
		}
	}

	// 如果 fetch 整体失败，记录失败统计
	if fetchErr != nil {
		// 已经在上面用 once.Do 记录了错误
		// 但这里需要统计失败的次数
		for range queries {
			x.target.IncrGetFailed()
		}
	}

	return res, err
}

func (x *ExpirationLRU[K, V]) Get(key K, fetch func() (V, error)) (V, error) {
	x.lock.Lock()
	v, ok := x.core.Get(key)
	if ok {
		x.lock.Unlock()
		x.target.IncrGetHit()
		v.lock.RLock()
		defer v.lock.RUnlock()
		return v.value, v.err
	} else {
		v = &expirationLruItem[V]{}
		x.core.Add(key, v)
		v.lock.Lock()
		x.lock.Unlock()
		defer v.lock.Unlock()
		v.value, v.err = fetch()
		if v.err == nil {
			x.target.IncrGetSuccess()
		} else {
			x.target.IncrGetFailed()
			x.core.Remove(key)
		}
		return v.value, v.err
	}
}

func (x *ExpirationLRU[K, V]) Del(key K) bool {
	x.lock.Lock()
	ok := x.core.Remove(key)
	x.lock.Unlock()
	if ok {
		x.target.IncrDelHit()
	} else {
		x.target.IncrDelNotFound()
	}
	return ok
}

func (x *ExpirationLRU[K, V]) SetHas(key K, value V) bool {
	x.lock.Lock()
	defer x.lock.Unlock()
	if x.core.Contains(key) {
		x.core.Add(key, &expirationLruItem[V]{value: value})
		return true
	}
	return false
}

func (x *ExpirationLRU[K, V]) Set(key K, value V) {
	x.lock.Lock()
	defer x.lock.Unlock()
	x.core.Add(key, &expirationLruItem[V]{value: value})
}

func (x *ExpirationLRU[K, V]) Stop() {
}
