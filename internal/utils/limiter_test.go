// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package utils_test

import (
	"runtime"
	"sync/atomic"
	"testing"

	"poly.red/internal/utils"
)

var (
	f = func() {
		p := 0
		for i := 0; i < 100000; i++ {
			p += i
		}
	}
)

func TestLimiter(t *testing.T) {
	l := utils.NewLimiter(2)
	sum := uint32(0)
	for i := 0; i < 10; i++ {
		ii := uint32(i)
		l.Execute(func() {
			atomic.AddUint32(&sum, ii)
		})
	}
	l.Wait()
	if sum != 45 {
		t.Fatalf("wrong sum, expect: %d, want %d", 45, sum)
	}

	sum = uint32(0)
	for i := 0; i < 10; i++ {
		ii := uint32(i)
		l.Execute(func() {
			atomic.AddUint32(&sum, ii)
		})
	}
	l.Wait()
	if sum != 45 {
		t.Fatalf("wrong sum, expect: %d, want %d", 45, sum)
	}
}

func BenchmarkLimiter(b *testing.B) {
	np := runtime.GOMAXPROCS(0)
	l := utils.NewLimiter(np)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Execute(f)
	}
	l.Wait()
}
