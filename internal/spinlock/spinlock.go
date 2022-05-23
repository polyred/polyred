// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package spinlock

import (
	"runtime"
	"sync"
	"sync/atomic"
)

var _ sync.Locker = &SpinLock{}

// SpinLock represents a spin lock.
type SpinLock struct {
	val uint32
}

const maxBackoff = 128 // heuristic

func (l *SpinLock) Lock() {
	backoff := 1
	for !atomic.CompareAndSwapUint32(&l.val, 0, 1) {
		// Implements exponential backoff spinlock.
		// See:
		// Herlihy, Maurice, et al. The art of multiprocessor programming. Newnes, 2020.
		// Secion 7.4 Exponential Backoff
		for i := 0; i < backoff; i++ {
			runtime.Gosched()
		}
		if backoff < maxBackoff {
			backoff = backoff << 1
		}
	}
}

func (l *SpinLock) Unlock() { atomic.StoreUint32(&l.val, 0) }
