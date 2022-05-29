// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package sched_test

import (
	"runtime"
	"sync/atomic"
	"testing"

	"poly.red/internal/sched"
)

func TestSched(t *testing.T) {
	s := sched.New(sched.Workers(4))
	sum := uint32(0)
	for i := 0; i < 10; i++ {
		ii := uint32(i)
		s.Run(func() {
			atomic.AddUint32(&sum, ii)
		})
	}
	s.Wait()
	if sum != 45 {
		t.Fatalf("wrong sum, expect: %d, want %d", 45, sum)
	}
	sum = uint32(0)
	for i := 0; i < 10; i++ {
		ii := uint32(i)
		s.Run(func() {
			atomic.AddUint32(&sum, ii)
		})
	}
	s.Wait()
	if sum != 45 {
		t.Fatalf("wrong sum, expect: %d, want %d", 45, sum)
	}
}

var (
	f = func() {
		p := 0
		for i := 0; i < 0; i++ {
			p += i
		}
	}
)

func BenchmarkSched(b *testing.B) {
	l := sched.New(sched.Workers(runtime.GOMAXPROCS(0)))
	b.Run("no-args", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l.Run(f)
		}
		l.Wait()
	})
	l.Release()
}
