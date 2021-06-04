// Copyright 2020 The golang.design Initiative Authors.
// All rights reserved. Use of this source code is governed
// by a MIT license that can be found in the LICENSE file.

// Package thread provides threading facilities, such as scheduling
// calls on a specific thread, local storage, etc.
package thread // import "golang.design/x/thread"

import (
	"runtime"
	"sync"
	"sync/atomic"
)

// Thread represents a thread instance.
type Thread interface {
	// ID returns the ID of the thread.
	ID() uint64

	// Call calls fn from the given thread. It blocks until fn returns.
	Call(fn func())

	// CallNonBlock call fn from the given thread without waiting
	// fn to complete.
	CallNonBlock(fn func())

	// CallV call fn from the given thread and returns the returned
	// value from fn.
	//
	// The purpose of this function is to avoid value escaping.
	// In particular:
	//
	//   th := thread.New()
	//   var ret interface{}
	//   th.Call(func() {
	//      ret = 1
	//   })
	//
	// will cause variable ret be allocated on the heap, whereas
	//
	//   th := thread.New()
	//   ret := th.CallV(func() interface{} {
	//     return 1
	//   }).(int)
	//
	// will offer zero allocation benefits.
	CallV(fn func() interface{}) interface{}

	// SetTLS stores a given value to the local storage of the given
	// thread. This method must be accessed in Call, or CallV, or
	// CallNonBlock. For instance:
	//
	//   th := thread.New()
	//   th.Call(func() {
	//      th.SetTLS("store in thread local storage")
	//   })
	SetTLS(x interface{})

	// GetTLS returns the locally stored value from local storage of
	// the given thread. This method must be access in Call, or CallV,
	// or CallNonBlock. For instance:
	//
	//   th := thread.New()
	//   th.Call(func() {
	//      tls := th.GetTLS()
	//      // ... do what ever you want to do with tls value ...
	//   })
	//
	GetTLS() interface{}

	// Terminate terminates the given thread gracefully.
	// Scheduled but unexecuted calls will be discarded.
	Terminate()
}

// New creates a new thread instance.
func New() Thread {
	th := thread{
		id:     atomic.AddUint64(&globalID, 1),
		fdCh:   make(chan funcData, runtime.GOMAXPROCS(0)),
		doneCh: make(chan struct{}),
	}
	runtime.SetFinalizer(&th, func(th interface{}) {
		th.(*thread).Terminate()
	})
	go func() {
		runtime.LockOSThread()
		for {
			select {
			case fd := <-th.fdCh:
				func() {
					if fd.fn != nil {
						defer func() {
							if fd.done != nil {
								fd.done <- struct{}{}
							}
						}()
						fd.fn()
					} else if fd.fnv != nil {
						var ret interface{}
						defer func() {
							if fd.ret != nil {
								fd.ret <- ret
							}
						}()
						ret = fd.fnv()
					}
				}()
			case <-th.doneCh:
				close(th.doneCh)
				return
			}
		}
	}()
	return &th
}

var (
	donePool = sync.Pool{
		New: func() interface{} {
			return make(chan struct{})
		},
	}
	varPool = sync.Pool{
		New: func() interface{} {
			return make(chan interface{})
		},
	}
	globalID uint64 // atomic
	_        Thread = &thread{}
)

type funcData struct {
	fn   func()
	done chan struct{}

	fnv func() interface{}
	ret chan interface{}
}

type thread struct {
	id  uint64
	tls interface{}

	fdCh   chan funcData
	doneCh chan struct{}
}

func (th thread) ID() uint64 {
	return th.id
}

func (th *thread) Call(fn func()) {
	if fn == nil {
		return
	}

	select {
	case <-th.doneCh:
		return
	default:
		done := donePool.Get().(chan struct{})
		defer donePool.Put(done)
		defer func() { <-done }()

		th.fdCh <- funcData{fn: fn, done: done}
	}
	return
}

func (th *thread) CallNonBlock(fn func()) {
	if fn == nil {
		return
	}
	select {
	case <-th.doneCh:
		return
	default:
		th.fdCh <- funcData{fn: fn}
	}
}

func (th *thread) CallV(fn func() interface{}) (ret interface{}) {
	if fn == nil {
		return nil
	}

	select {
	case <-th.doneCh:
		return nil
	default:
		done := varPool.Get().(chan interface{})
		defer varPool.Put(done)
		defer func() { ret = <-done }()

		th.fdCh <- funcData{fnv: fn, ret: done}
		return
	}
}

func (th *thread) GetTLS() interface{} {
	return th.tls
}

func (th *thread) SetTLS(x interface{}) {
	th.tls = x
}

func (th *thread) Terminate() {
	select {
	case <-th.doneCh:
		return
	default:
		th.doneCh <- struct{}{}
		select {
		case <-th.doneCh:
			return
		}
	}
}
