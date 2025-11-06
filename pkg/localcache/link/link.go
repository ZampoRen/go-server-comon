// Package link 提供基于分片锁的键关联缓存实现
package link

import (
	"hash/fnv"
	"sync"
)

// Link 定义了键关联缓存的接口
type Link interface {
	// Link 建立 key 与 link 中所有键的双向关联关系
	Link(key string, link ...string)
	// Del 删除指定的 key 及其所有关联的键（级联删除）
	Del(key string) map[string]struct{}
}

func newLinkKey() *linkKey {
	return &linkKey{
		data: make(map[string]map[string]struct{}),
	}
}

type linkKey struct {
	lock sync.Mutex
	data map[string]map[string]struct{}
}

func (x *linkKey) link(key string, link ...string) {
	x.lock.Lock()
	defer x.lock.Unlock()

	v, ok := x.data[key]
	if !ok {
		v = make(map[string]struct{})
		x.data[key] = v
	}

	for _, k := range link {
		v[k] = struct{}{}
	}
}

func (x *linkKey) del(key string) map[string]struct{} {
	x.lock.Lock()
	defer x.lock.Unlock()

	ks, ok := x.data[key]
	if !ok {
		return nil
	}

	delete(x.data, key)
	return ks
}

// New 创建一个新的分片键关联缓存实例
func New(n int) Link {
	if n <= 0 {
		panic("slot count must be greater than 0")
	}

	slots := make([]*linkKey, n)
	for i := 0; i < n; i++ {
		slots[i] = newLinkKey()
	}

	return &slot{
		n:     uint64(n),
		slots: slots,
	}
}

type slot struct {
	n     uint64
	slots []*linkKey
}

func (x *slot) index(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64() % x.n
}

func (x *slot) Link(key string, link ...string) {
	if len(link) == 0 {
		return
	}

	x.slots[x.index(key)].link(key, link...)

	for _, lk := range link {
		x.slots[x.index(lk)].link(lk, key)
	}
}

func (x *slot) Del(key string) map[string]struct{} {
	return x.delKey(key)
}

func (x *slot) delKey(k string) map[string]struct{} {
	del := make(map[string]struct{})
	stack := []string{k}

	for len(stack) > 0 {
		curr := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if _, ok := del[curr]; ok {
			continue
		}

		del[curr] = struct{}{}
		childKeys := x.slots[x.index(curr)].del(curr)

		for ck := range childKeys {
			stack = append(stack, ck)
		}
	}

	return del
}
