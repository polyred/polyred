package cache

import "sync"

var cache sync.Map

func Set(id uint64, obj any) { cache.Store(id, obj) }

func Get[T any, _ *T](id uint64) (x T) {
	v, _ := cache.Load(id)
	if v == nil {
		return
	}
	return v.(T)
}
