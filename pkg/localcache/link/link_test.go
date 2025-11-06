package link

import (
	"sync"
	"testing"
)

// TestNew 测试创建新的 Link 实例
func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		n         int
		wantPanic bool
	}{
		{
			name:      "正常创建",
			n:         10,
			wantPanic: false,
		},
		{
			name:      "单个分片",
			n:         1,
			wantPanic: false,
		},
		{
			name:      "大量分片",
			n:         100,
			wantPanic: false,
		},
		{
			name:      "零分片应该panic",
			n:         0,
			wantPanic: true,
		},
		{
			name:      "负数分片应该panic",
			n:         -1,
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); (r != nil) != tt.wantPanic {
					t.Errorf("New() panic = %v, wantPanic %v", r != nil, tt.wantPanic)
				}
			}()
			l := New(tt.n)
			if l == nil {
				t.Error("New() returned nil")
			}
		})
	}
}

// TestLink_BasicLink 测试基本的关联功能
func TestLink_BasicLink(t *testing.T) {
	l := New(10)

	// 测试建立关联
	l.Link("key1", "link1", "link2")
	l.Link("key2", "link2", "link3")

	// 测试删除 key1，应该返回 key1, link1, link2
	// 因为 key1 关联了 link1 和 link2，link2 又关联了 key2
	del := l.Del("key1")

	expected := map[string]struct{}{
		"key1":  {},
		"link1": {},
		"link2": {},
		"key2":  {}, // 因为 link2 关联了 key2
		"link3": {}, // 因为 key2 关联了 link3
	}

	if len(del) != len(expected) {
		t.Errorf("Del() returned %d keys, want %d", len(del), len(expected))
	}

	for k := range expected {
		if _, ok := del[k]; !ok {
			t.Errorf("Del() missing key: %s", k)
		}
	}
}

// TestLink_EmptyLink 测试空关联列表
func TestLink_EmptyLink(t *testing.T) {
	l := New(10)

	// 应该不会panic
	l.Link("key1")

	// 删除应该只返回 key1
	del := l.Del("key1")
	if len(del) != 1 {
		t.Errorf("Del() returned %d keys, want 1", len(del))
	}
	if _, ok := del["key1"]; !ok {
		t.Error("Del() missing key1")
	}
}

// TestLink_MultipleLinks 测试多次关联同一组键
func TestLink_MultipleLinks(t *testing.T) {
	l := New(10)

	// 多次关联，应该不会重复
	l.Link("key1", "link1", "link2")
	l.Link("key1", "link1", "link3") // link1 重复关联

	del := l.Del("key1")

	// 应该包含所有关联的键
	expectedKeys := []string{"key1", "link1", "link2", "link3"}
	for _, k := range expectedKeys {
		if _, ok := del[k]; !ok {
			t.Errorf("Del() missing key: %s", k)
		}
	}
}

// TestLink_CascadeDelete 测试级联删除
func TestLink_CascadeDelete(t *testing.T) {
	l := New(10)

	// 建立复杂的关联关系
	// key1 -> link1, link2
	// link1 -> link3
	// link2 -> link4
	// link3 -> link5
	l.Link("key1", "link1", "link2")
	l.Link("link1", "link3")
	l.Link("link2", "link4")
	l.Link("link3", "link5")

	// 删除 key1 应该级联删除所有关联的键
	del := l.Del("key1")

	expectedKeys := []string{"key1", "link1", "link2", "link3", "link4", "link5"}
	if len(del) != len(expectedKeys) {
		t.Errorf("Del() returned %d keys, want %d", len(del), len(expectedKeys))
	}

	for _, k := range expectedKeys {
		if _, ok := del[k]; !ok {
			t.Errorf("Del() missing key: %s", k)
		}
	}
}

// TestLink_ConcurrentAccess 测试并发访问安全性
func TestLink_ConcurrentAccess(t *testing.T) {
	l := New(100) // 使用更多分片提高并发性能
	var wg sync.WaitGroup
	goroutines := 100
	opsPerGoroutine := 100

	// 并发写入
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				key := "key" + string(rune(id*100+j))
				link1 := "link1" + string(rune(id*100+j))
				link2 := "link2" + string(rune(id*100+j))
				l.Link(key, link1, link2)
			}
		}(i)
	}
	wg.Wait()

	// 并发删除
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			key := "key" + string(rune(id*100))
			del := l.Del(key)
			if del == nil {
				t.Errorf("Del() returned nil for key: %s", key)
			}
		}(i)
	}
	wg.Wait()
}

// TestLink_DeleteNonExistent 测试删除不存在的键
func TestLink_DeleteNonExistent(t *testing.T) {
	l := New(10)

	del := l.Del("non_existent_key")

	// 如果键不存在，应该返回空集合或只包含该键
	if len(del) > 0 {
		if len(del) == 1 {
			if _, ok := del["non_existent_key"]; !ok {
				t.Error("Del() should return the key itself even if it doesn't exist")
			}
		}
	}
}

// TestLink_BidirectionalLink 测试双向关联
func TestLink_BidirectionalLink(t *testing.T) {
	l := New(10)

	// 建立双向关联：key1 <-> link1
	l.Link("key1", "link1")

	// 删除 key1 应该也删除 link1
	del1 := l.Del("key1")
	if _, ok := del1["link1"]; !ok {
		t.Error("Del() should cascade delete bidirectionally linked key")
	}
	if _, ok := del1["key1"]; !ok {
		t.Error("Del() should include the deleted key itself")
	}

	// 重新建立关联
	l = New(10)
	l.Link("key1", "link1")

	// 删除 link1 应该也删除 key1
	del2 := l.Del("link1")
	if _, ok := del2["key1"]; !ok {
		t.Error("Del() should cascade delete bidirectionally linked key")
	}
	if _, ok := del2["link1"]; !ok {
		t.Error("Del() should include the deleted key itself")
	}
}

// TestLink_IsolatedKeys 测试独立键的删除
func TestLink_IsolatedKeys(t *testing.T) {
	l := New(10)

	// 创建两个独立的键组
	l.Link("key1", "link1")
	l.Link("key2", "link2")

	// 删除 key1 不应该影响 key2 组
	del := l.Del("key1")

	// 应该只包含 key1 组的键
	if _, ok := del["key2"]; ok {
		t.Error("Del() should not delete keys from unrelated groups")
	}
	if _, ok := del["link2"]; ok {
		t.Error("Del() should not delete keys from unrelated groups")
	}

	// key1 和 link1 应该在删除列表中
	if _, ok := del["key1"]; !ok {
		t.Error("Del() should include key1")
	}
	if _, ok := del["link1"]; !ok {
		t.Error("Del() should include link1")
	}
}

// BenchmarkLink 基准测试 Link 操作
func BenchmarkLink(b *testing.B) {
	l := New(100)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := "key" + string(rune(i))
		link1 := "link1" + string(rune(i))
		link2 := "link2" + string(rune(i))
		l.Link(key, link1, link2)
	}
}

// BenchmarkDel 基准测试 Del 操作
func BenchmarkDel(b *testing.B) {
	l := New(100)

	// 预先创建一些关联
	for i := 0; i < 1000; i++ {
		key := "key" + string(rune(i))
		link1 := "link1" + string(rune(i))
		link2 := "link2" + string(rune(i))
		l.Link(key, link1, link2)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "key" + string(rune(i%1000))
		l.Del(key)
	}
}

// BenchmarkConcurrentLink 并发 Link 操作基准测试
func BenchmarkConcurrentLink(b *testing.B) {
	l := New(100)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := "key" + string(rune(i))
			link1 := "link1" + string(rune(i))
			link2 := "link2" + string(rune(i))
			l.Link(key, link1, link2)
			i++
		}
	})
}
