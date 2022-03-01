package sched_test

import (
	"runtime"
	"sync/atomic"
	"testing"

	"poly.red/internal/sched"
)

func TestSched(t *testing.T) {
	s := sched.New(sched.Workers(4))
	s.Add(10)
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

	s.Add(10)
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

	if s.Running() != 0 {
		t.Fatalf("wrong counter inside the pool")
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
		l.Add(uint64(b.N))
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l.Run(f)
		}
		l.Wait()
	})
	b.Run("with-args", func(b *testing.B) {
		l.Add(uint64(b.N))
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l.RunWithArgs(func(x any) {
				_ = x
			}, 42)
		}
		l.Wait()
	})
	l.Release()
}
