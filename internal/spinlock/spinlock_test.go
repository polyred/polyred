// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package spinlock_test

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"

	"poly.red/internal/spinlock"
)

type naiveSpinlock struct {
	val uint32
}

func (l *naiveSpinlock) Lock() {
	for !atomic.CompareAndSwapUint32(&l.val, 0, 1) {
		runtime.Gosched()
	}
}
func (l *naiveSpinlock) Unlock() { atomic.StoreUint32(&l.val, 0) }

func TestLock(t *testing.T) {
	l := spinlock.SpinLock{}
	l.Lock()
	_ = 42
	l.Unlock()
}

func BenchmarkLocks(b *testing.B) {
	locks := map[string]sync.Locker{
		"mutex":   &sync.Mutex{},
		"naive":   &naiveSpinlock{},
		"backoff": &spinlock.SpinLock{},
	}
	for name, lock := range locks {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					lock.Lock()
					_ = 1
					lock.Unlock()
				}
			})
		})
	}
}
