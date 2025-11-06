package localcache

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestNew 测试创建新的 Cache 实例
func TestNew(t *testing.T) {
	tests := []struct {
		name string
		opts []Option
	}{
		{
			name: "默认配置",
			opts: nil,
		},
		{
			name: "单分片",
			opts: []Option{
				WithLocalSlotNum(1),
				WithLocalSlotSize(100),
			},
		},
		{
			name: "多分片",
			opts: []Option{
				WithLocalSlotNum(10),
				WithLocalSlotSize(100),
			},
		},
		{
			name: "禁用本地缓存",
			opts: []Option{
				WithLocalDisable(),
			},
		},
		{
			name: "禁用关联功能",
			opts: []Option{
				WithLinkDisable(),
			},
		},
		{
			name: "Lazy 策略",
			opts: []Option{
				WithLazy(),
				WithLocalSlotNum(1),
				WithLocalSlotSize(100),
			},
		},
		{
			name: "Expiration 策略",
			opts: []Option{
				WithExpirationEvict(),
				WithLocalSlotNum(1),
				WithLocalSlotSize(100),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := New[string](tt.opts...)
			if cache == nil {
				t.Error("New() returned nil")
			}
			defer cache.Stop()
		})
	}
}

// TestCache_Get 测试基本的 Get 功能
func TestCache_Get(t *testing.T) {
	cache := New[string](
		WithLocalSlotNum(1),
		WithLocalSlotSize(10),
	)
	defer cache.Stop()

	ctx := context.Background()

	// 测试缓存未命中，调用 fetch
	fetchCount := 0
	value, err := cache.Get(ctx, "key1", func(ctx context.Context) (string, error) {
		fetchCount++
		return "value1", nil
	})

	if err != nil {
		t.Errorf("Get() error = %v, want nil", err)
	}
	if value != "value1" {
		t.Errorf("Get() value = %v, want value1", value)
	}
	if fetchCount != 1 {
		t.Errorf("fetch called %d times, want 1", fetchCount)
	}

	// 测试缓存命中，不调用 fetch
	value2, err := cache.Get(ctx, "key1", func(ctx context.Context) (string, error) {
		fetchCount++
		return "should not be called", nil
	})

	if err != nil {
		t.Errorf("Get() error = %v, want nil", err)
	}
	if value2 != "value1" {
		t.Errorf("Get() value = %v, want value1", value2)
	}
	if fetchCount != 1 {
		t.Errorf("fetch called %d times, want 1", fetchCount)
	}
}

// TestCache_Get_Error 测试 Get 错误处理
func TestCache_Get_Error(t *testing.T) {
	cache := New[string](
		WithLocalSlotNum(1),
		WithLocalSlotSize(10),
	)
	defer cache.Stop()

	ctx := context.Background()

	// 测试 fetch 返回错误
	testErr := errors.New("fetch error")
	value, err := cache.Get(ctx, "key1", func(ctx context.Context) (string, error) {
		return "", testErr
	})

	if !errors.Is(err, testErr) {
		t.Errorf("Get() error = %v, want %v", err, testErr)
	}
	if value != "" {
		t.Errorf("Get() value = %v, want empty string", value)
	}
}

// TestCache_GetLink 测试 GetLink 功能
func TestCache_GetLink(t *testing.T) {
	cache := New[string](
		WithLocalSlotNum(1),
		WithLocalSlotSize(10),
		WithLinkSlotNum(10),
	)
	defer cache.Stop()

	ctx := context.Background()

	// 先单独缓存关联键（在建立关联之前）
	cache.Get(ctx, "user:123:profile", func(ctx context.Context) (string, error) {
		return "profile123", nil
	})
	cache.Get(ctx, "user:123:settings", func(ctx context.Context) (string, error) {
		return "settings123", nil
	})

	// 测试建立关联（此时 user:123 不在缓存中，会调用 fetch，建立关联关系）
	value, err := cache.GetLink(ctx, "user:123", func(ctx context.Context) (string, error) {
		return "user123", nil
	}, "user:123:profile", "user:123:settings")

	if err != nil {
		t.Errorf("GetLink() error = %v, want nil", err)
	}
	if value != "user123" {
		t.Errorf("GetLink() value = %v, want user123", value)
	}

	// 验证关联键仍然存在（此时还未删除主键）
	fetchCount := 0
	profileValue, _ := cache.Get(ctx, "user:123:profile", func(ctx context.Context) (string, error) {
		fetchCount++
		return "new profile", nil
	})
	if fetchCount != 0 {
		t.Error("关联键 user:123:profile 应该仍然存在")
	}
	if profileValue != "profile123" {
		t.Errorf("profile value = %v, want profile123", profileValue)
	}

	// 删除主键，应该级联删除关联键
	cache.Del(ctx, "user:123")

	// 验证关联键也被删除（缓存未命中）
	fetchCount = 0
	_, _ = cache.Get(ctx, "user:123:profile", func(ctx context.Context) (string, error) {
		fetchCount++
		return "new profile", nil
	})
	if fetchCount == 0 {
		t.Error("关联键 user:123:profile 应该被删除，但缓存命中")
	}

	fetchCount = 0
	_, _ = cache.Get(ctx, "user:123:settings", func(ctx context.Context) (string, error) {
		fetchCount++
		return "new settings", nil
	})
	if fetchCount == 0 {
		t.Error("关联键 user:123:settings 应该被删除，但缓存命中")
	}
}

// TestCache_Del 测试删除功能
func TestCache_Del(t *testing.T) {
	cache := New[string](
		WithLocalSlotNum(1),
		WithLocalSlotSize(10),
	)
	defer cache.Stop()

	ctx := context.Background()

	// 添加数据
	cache.Get(ctx, "key1", func(ctx context.Context) (string, error) {
		return "value1", nil
	})
	cache.Get(ctx, "key2", func(ctx context.Context) (string, error) {
		return "value2", nil
	})

	// 删除 key1
	cache.Del(ctx, "key1")

	// 验证 key1 被删除
	fetchCount := 0
	value, err := cache.Get(ctx, "key1", func(ctx context.Context) (string, error) {
		fetchCount++
		return "new value1", nil
	})
	if err != nil {
		t.Errorf("Get() error = %v, want nil", err)
	}
	if value != "new value1" {
		t.Errorf("Get() value = %v, want new value1", value)
	}
	if fetchCount != 1 {
		t.Error("key1 应该被删除，需要重新 fetch")
	}

	// 验证 key2 仍然存在
	fetchCount = 0
	value2, err := cache.Get(ctx, "key2", func(ctx context.Context) (string, error) {
		fetchCount++
		return "should not be called", nil
	})
	if err != nil {
		t.Errorf("Get() error = %v, want nil", err)
	}
	if value2 != "value2" {
		t.Errorf("Get() value = %v, want value2", value2)
	}
	if fetchCount != 0 {
		t.Error("key2 应该仍然存在，不应该调用 fetch")
	}
}

// TestCache_Del_Multiple 测试批量删除
func TestCache_Del_Multiple(t *testing.T) {
	cache := New[string](
		WithLocalSlotNum(1),
		WithLocalSlotSize(10),
	)
	defer cache.Stop()

	ctx := context.Background()

	// 添加多个数据
	for i := 0; i < 5; i++ {
		key := "key" + strconv.Itoa(i)
		cache.Get(ctx, key, func(ctx context.Context) (string, error) {
			return "value" + strconv.Itoa(i), nil
		})
	}

	// 批量删除
	cache.Del(ctx, "key0", "key1", "key2")

	// 验证删除的键
	for i := 0; i < 3; i++ {
		key := "key" + strconv.Itoa(i)
		fetchCount := 0
		_, _ = cache.Get(ctx, key, func(ctx context.Context) (string, error) {
			fetchCount++
			return "new", nil
		})
		if fetchCount == 0 {
			t.Errorf("key %s 应该被删除", key)
		}
	}

	// 验证未删除的键
	for i := 3; i < 5; i++ {
		key := "key" + strconv.Itoa(i)
		fetchCount := 0
		_, _ = cache.Get(ctx, key, func(ctx context.Context) (string, error) {
			fetchCount++
			return "should not be called", nil
		})
		if fetchCount != 0 {
			t.Errorf("key %s 不应该被删除", key)
		}
	}
}

// TestCache_DelLocal 测试 DelLocal 功能
func TestCache_DelLocal(t *testing.T) {
	cache := New[string](
		WithLocalSlotNum(1),
		WithLocalSlotSize(10),
	)
	defer cache.Stop()

	ctx := context.Background()

	// 添加数据
	cache.Get(ctx, "key1", func(ctx context.Context) (string, error) {
		return "value1", nil
	})

	// 使用 DelLocal 删除
	cache.DelLocal(ctx, "key1")

	// 验证被删除
	fetchCount := 0
	_, _ = cache.Get(ctx, "key1", func(ctx context.Context) (string, error) {
		fetchCount++
		return "new value1", nil
	})
	if fetchCount == 0 {
		t.Error("key1 应该被删除")
	}
}

// TestCache_Del_WithCallback 测试删除回调
func TestCache_Del_WithCallback(t *testing.T) {
	var deletedKeys []string
	var mu sync.Mutex

	cache := New[string](
		WithLocalSlotNum(1),
		WithLocalSlotSize(10),
		WithDeleteKeyBefore(func(ctx context.Context, key ...string) {
			mu.Lock()
			deletedKeys = append(deletedKeys, key...)
			mu.Unlock()
		}),
	)
	defer cache.Stop()

	ctx := context.Background()

	// 添加数据
	cache.Get(ctx, "key1", func(ctx context.Context) (string, error) {
		return "value1", nil
	})

	// 删除
	cache.Del(ctx, "key1")

	// 验证回调被调用
	mu.Lock()
	if len(deletedKeys) != 1 || deletedKeys[0] != "key1" {
		t.Errorf("删除回调应该被调用，deletedKeys = %v", deletedKeys)
	}
	mu.Unlock()
}

// TestCache_GetLink_CascadeDelete 测试级联删除
func TestCache_GetLink_CascadeDelete(t *testing.T) {
	cache := New[string](
		WithLocalSlotNum(1),
		WithLocalSlotSize(10),
		WithLinkSlotNum(10),
	)
	defer cache.Stop()

	ctx := context.Background()

	// 建立关联关系
	cache.GetLink(ctx, "user:123", func(ctx context.Context) (string, error) {
		return "user123", nil
	}, "user:123:profile", "user:123:settings")

	// 单独缓存关联键
	cache.Get(ctx, "user:123:profile", func(ctx context.Context) (string, error) {
		return "profile123", nil
	})
	cache.Get(ctx, "user:123:settings", func(ctx context.Context) (string, error) {
		return "settings123", nil
	})

	// 删除主键
	cache.Del(ctx, "user:123")

	// 验证关联键也被删除
	keys := []string{"user:123:profile", "user:123:settings"}
	for _, key := range keys {
		fetchCount := 0
		_, _ = cache.Get(ctx, key, func(ctx context.Context) (string, error) {
			fetchCount++
			return "new", nil
		})
		if fetchCount == 0 {
			t.Errorf("关联键 %s 应该被级联删除", key)
		}
	}
}

// TestCache_GetLink_NoLink 测试 GetLink 不建立关联的情况
func TestCache_GetLink_NoLink(t *testing.T) {
	cache := New[string](
		WithLocalSlotNum(1),
		WithLocalSlotSize(10),
		WithLinkDisable(), // 禁用关联功能
	)
	defer cache.Stop()

	ctx := context.Background()

	// 使用 GetLink 但不应该建立关联
	cache.GetLink(ctx, "user:123", func(ctx context.Context) (string, error) {
		return "user123", nil
	}, "user:123:profile")

	// 缓存关联键
	cache.Get(ctx, "user:123:profile", func(ctx context.Context) (string, error) {
		return "profile123", nil
	})

	// 删除主键，关联键不应该被删除（因为关联功能被禁用）
	cache.Del(ctx, "user:123")

	// 验证关联键仍然存在
	fetchCount := 0
	value, err := cache.Get(ctx, "user:123:profile", func(ctx context.Context) (string, error) {
		fetchCount++
		return "new profile", nil
	})
	if err != nil {
		t.Errorf("Get() error = %v, want nil", err)
	}
	if value != "profile123" {
		t.Errorf("Get() value = %v, want profile123", value)
	}
	if fetchCount != 0 {
		t.Error("关联键应该仍然存在，因为关联功能被禁用")
	}
}

// TestCache_LocalDisable 测试禁用本地缓存
func TestCache_LocalDisable(t *testing.T) {
	cache := New[string](
		WithLocalDisable(),
	)
	defer cache.Stop()

	ctx := context.Background()

	// 每次都应该调用 fetch
	fetchCount := 0
	value, err := cache.Get(ctx, "key1", func(ctx context.Context) (string, error) {
		fetchCount++
		return "value1", nil
	})

	if err != nil {
		t.Errorf("Get() error = %v, want nil", err)
	}
	if value != "value1" {
		t.Errorf("Get() value = %v, want value1", value)
	}
	if fetchCount != 1 {
		t.Errorf("fetch called %d times, want 1", fetchCount)
	}

	// 再次获取，应该再次调用 fetch（因为没有缓存）
	value2, err := cache.Get(ctx, "key1", func(ctx context.Context) (string, error) {
		fetchCount++
		return "value1", nil
	})

	if err != nil {
		t.Errorf("Get() error = %v, want nil", err)
	}
	if value2 != "value1" {
		t.Errorf("Get() value = %v, want value1", value2)
	}
	if fetchCount != 2 {
		t.Errorf("fetch called %d times, want 2", fetchCount)
	}
}

// TestCache_Expiration 测试过期策略
func TestCache_Expiration(t *testing.T) {
	cache := New[string](
		WithLocalSlotNum(1),
		WithLocalSlotSize(10),
		WithLocalSuccessTTL(100*time.Millisecond),
		WithExpirationEvict(),
	)
	defer cache.Stop()

	ctx := context.Background()

	// 添加数据
	cache.Get(ctx, "key1", func(ctx context.Context) (string, error) {
		return "value1", nil
	})

	// 立即获取，应该命中缓存
	fetchCount := 0
	value, err := cache.Get(ctx, "key1", func(ctx context.Context) (string, error) {
		fetchCount++
		return "should not be called", nil
	})
	if err != nil {
		t.Errorf("Get() error = %v, want nil", err)
	}
	if value != "value1" {
		t.Errorf("Get() value = %v, want value1", value)
	}
	if fetchCount != 0 {
		t.Error("应该命中缓存")
	}

	// 等待过期
	time.Sleep(150 * time.Millisecond)

	// 再次获取，应该重新 fetch（ExpirationLRU 会自动清理过期项）
	value2, err := cache.Get(ctx, "key1", func(ctx context.Context) (string, error) {
		fetchCount++
		return "new value1", nil
	})
	if err != nil {
		t.Errorf("Get() error = %v, want nil", err)
	}
	if value2 != "new value1" {
		t.Errorf("Get() value = %v, want new value1", value2)
	}
	if fetchCount == 0 {
		t.Error("应该重新 fetch，因为已过期")
	}
}

// TestCache_LazyExpiration 测试懒删除策略
func TestCache_LazyExpiration(t *testing.T) {
	cache := New[string](
		WithLocalSlotNum(1),
		WithLocalSlotSize(10),
		WithLocalSuccessTTL(100*time.Millisecond),
		WithLazy(),
	)
	defer cache.Stop()

	ctx := context.Background()

	// 添加数据
	cache.Get(ctx, "key1", func(ctx context.Context) (string, error) {
		return "value1", nil
	})

	// 等待过期
	time.Sleep(150 * time.Millisecond)

	// 访问过期项，应该重新 fetch（懒删除策略）
	fetchCount := 0
	value, err := cache.Get(ctx, "key1", func(ctx context.Context) (string, error) {
		fetchCount++
		return "new value1", nil
	})
	if err != nil {
		t.Errorf("Get() error = %v, want nil", err)
	}
	if value != "new value1" {
		t.Errorf("Get() value = %v, want new value1", value)
	}
	if fetchCount == 0 {
		t.Error("应该重新 fetch，因为已过期")
	}
}

// TestCache_Concurrent 测试并发安全
func TestCache_Concurrent(t *testing.T) {
	cache := New[string](
		WithLocalSlotNum(10), // 多分片提高并发性能
		WithLocalSlotSize(100),
		WithLinkSlotNum(10),
	)
	defer cache.Stop()

	ctx := context.Background()
	var wg sync.WaitGroup
	concurrency := 100

	// 并发写入
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func(id int) {
			defer wg.Done()
			key := "key" + strconv.Itoa(id%10)
			_, _ = cache.Get(ctx, key, func(ctx context.Context) (string, error) {
				return "value" + strconv.Itoa(id), nil
			})
		}(i)
	}
	wg.Wait()

	// 并发读取
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func(id int) {
			defer wg.Done()
			key := "key" + strconv.Itoa(id%10)
			_, _ = cache.Get(ctx, key, func(ctx context.Context) (string, error) {
				return "new value", nil
			})
		}(i)
	}
	wg.Wait()

	// 并发删除
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func(id int) {
			defer wg.Done()
			key := "key" + strconv.Itoa(id%10)
			cache.Del(ctx, key)
		}(i)
	}
	wg.Wait()
}

// TestCache_GetLink_Concurrent 测试并发 GetLink
func TestCache_GetLink_Concurrent(t *testing.T) {
	cache := New[string](
		WithLocalSlotNum(10),
		WithLocalSlotSize(100),
		WithLinkSlotNum(10),
	)
	defer cache.Stop()

	ctx := context.Background()
	var wg sync.WaitGroup
	concurrency := 50

	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func(id int) {
			defer wg.Done()
			key := "user:" + strconv.Itoa(id)
			_, _ = cache.GetLink(ctx, key, func(ctx context.Context) (string, error) {
				return "user" + strconv.Itoa(id), nil
			}, key+":profile", key+":settings")
		}(i)
	}
	wg.Wait()
}

// TestCache_Target 测试统计功能
func TestCache_Target(t *testing.T) {
	var hits, successes, failures, delHits, delNotFound int64

	target := &testTarget{
		incrGetHit:      func() { atomic.AddInt64(&hits, 1) },
		incrGetSuccess:  func() { atomic.AddInt64(&successes, 1) },
		incrGetFailed:   func() { atomic.AddInt64(&failures, 1) },
		incrDelHit:      func() { atomic.AddInt64(&delHits, 1) },
		incrDelNotFound: func() { atomic.AddInt64(&delNotFound, 1) },
	}

	cache := New[string](
		WithLocalSlotNum(1),
		WithLocalSlotSize(10),
		WithTarget(target),
	)
	defer cache.Stop()

	ctx := context.Background()

	// 成功获取
	_, _ = cache.Get(ctx, "key1", func(ctx context.Context) (string, error) {
		return "value1", nil
	})

	// 再次获取（应该命中缓存）
	_, _ = cache.Get(ctx, "key1", func(ctx context.Context) (string, error) {
		return "should not be called", nil
	})

	// 失败获取
	_, _ = cache.Get(ctx, "key2", func(ctx context.Context) (string, error) {
		return "", errors.New("fetch error")
	})

	// 删除存在的键
	cache.Del(ctx, "key1")

	// 删除不存在的键
	cache.Del(ctx, "key999")

	// 验证统计
	if atomic.LoadInt64(&hits) == 0 {
		t.Error("应该记录缓存命中")
	}
	if atomic.LoadInt64(&successes) == 0 {
		t.Error("应该记录成功获取")
	}
	if atomic.LoadInt64(&failures) == 0 {
		t.Error("应该记录失败获取")
	}
	if atomic.LoadInt64(&delHits) == 0 {
		t.Error("应该记录删除命中")
	}
	if atomic.LoadInt64(&delNotFound) == 0 {
		t.Error("应该记录删除未找到")
	}
}

type testTarget struct {
	incrGetHit      func()
	incrGetSuccess  func()
	incrGetFailed   func()
	incrDelHit      func()
	incrDelNotFound func()
}

func (t *testTarget) IncrGetHit() {
	if t.incrGetHit != nil {
		t.incrGetHit()
	}
}

func (t *testTarget) IncrGetSuccess() {
	if t.incrGetSuccess != nil {
		t.incrGetSuccess()
	}
}

func (t *testTarget) IncrGetFailed() {
	if t.incrGetFailed != nil {
		t.incrGetFailed()
	}
}

func (t *testTarget) IncrDelHit() {
	if t.incrDelHit != nil {
		t.incrDelHit()
	}
}

func (t *testTarget) IncrDelNotFound() {
	if t.incrDelNotFound != nil {
		t.incrDelNotFound()
	}
}

// TestCache_Stop 测试 Stop 方法
func TestCache_Stop(t *testing.T) {
	cache := New[string](
		WithLocalSlotNum(1),
		WithLocalSlotSize(10),
	)

	// Stop 不应该 panic
	cache.Stop()

	// 再次 Stop 也不应该 panic
	cache.Stop()
}

// TestCache_LRUStringHash 测试哈希函数
func TestCache_LRUStringHash(t *testing.T) {
	hash1 := LRUStringHash("key1")
	hash2 := LRUStringHash("key2")
	hash3 := LRUStringHash("key1")

	// 相同 key 应该产生相同 hash
	if hash1 != hash3 {
		t.Error("相同 key 应该产生相同 hash")
	}

	// 不同 key 应该产生不同 hash（大概率）
	if hash1 == hash2 {
		t.Log("警告：不同 key 产生了相同 hash（虽然可能，但概率很低）")
	}
}

// TestCache_MultiSlot 测试多分片功能
func TestCache_MultiSlot(t *testing.T) {
	cache := New[string](
		WithLocalSlotNum(10), // 10 个分片
		WithLocalSlotSize(10),
	)
	defer cache.Stop()

	ctx := context.Background()

	// 添加多个键，它们应该分布到不同的分片
	for i := 0; i < 20; i++ {
		key := "key" + strconv.Itoa(i)
		_, _ = cache.Get(ctx, key, func(ctx context.Context) (string, error) {
			return "value" + strconv.Itoa(i), nil
		})
	}

	// 验证所有键都可以正常获取
	for i := 0; i < 20; i++ {
		key := "key" + strconv.Itoa(i)
		fetchCount := 0
		value, err := cache.Get(ctx, key, func(ctx context.Context) (string, error) {
			fetchCount++
			return "should not be called", nil
		})
		if err != nil {
			t.Errorf("Get() error = %v, want nil", err)
		}
		expected := "value" + strconv.Itoa(i)
		if value != expected {
			t.Errorf("Get() value = %v, want %v", value, expected)
		}
		if fetchCount != 0 {
			t.Errorf("key %s 应该命中缓存", key)
		}
	}
}
