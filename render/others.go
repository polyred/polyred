// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render

import (
	"runtime"
	"sync/atomic"
)

func (r *Renderer) wait() {
	backoff := 1
	atomic.StoreUint32(&r.stop, 1)
	for atomic.LoadUint32(&r.running) == 1 {
		for i := 0; i < backoff; i++ {
			runtime.Gosched()
		}
		if backoff < 128 {
			backoff <<= 1
		}
	}
	atomic.StoreUint32(&r.stop, 0)
}

func (r *Renderer) startRunning() {
	atomic.StoreUint32(&r.running, 1)
}

func (r *Renderer) stopRunning() {
	atomic.StoreUint32(&r.running, 0)
}

func (r *Renderer) shouldStop() bool {
	return atomic.LoadUint32(&r.stop) == 1
}
