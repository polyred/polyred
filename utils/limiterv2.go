// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package utils

import (
	"math/rand"
	"runtime"
	"sync/atomic"
)

type WorkerPool struct {
	num        uint64
	running    uint64
	done       chan struct{}
	taskQueues chan funcdata
	workers    []chan funcdata
}

type funcdata struct {
	fn func()
}

func NewWorkerPool(limit uint64) *WorkerPool {
	taskQueue := make(chan funcdata, runtime.GOMAXPROCS(0))
	workers := make([]chan funcdata, limit)
	for i := uint64(0); i < limit; i++ {
		workers[i] = make(chan funcdata, 256)
	}
	p := &WorkerPool{
		num:        limit,
		running:    0,
		taskQueues: taskQueue,
		done:       make(chan struct{}, 1),
		workers:    workers,
	}
	go func() {
		for i := 0; i < len(workers); i++ {
			go func(workerId int) {
				for fd := range workers[workerId] {
					fd.fn()
					p.Done()
				}
			}(i)
		}
	}()
	go func() {
		fanout(func(m int) int { return rand.Intn(m) }, taskQueue, workers...)
	}()
	return p
}

func (p *WorkerPool) Execute(fs ...func()) {
	for i := range fs {
		p.taskQueues <- funcdata{fn: fs[i]}
	}
}

func (p *WorkerPool) Add(numTasks uint64) uint64 {
	return atomic.AddUint64(&p.running, numTasks)
}

func (p *WorkerPool) Done() {
	ret := atomic.AddUint64(&p.running, ^uint64(0))
	if ret == 0 {
		p.done <- struct{}{}
	}
}

func (p *WorkerPool) Wait() {
	<-p.done
}

func (p *WorkerPool) Running() uint64 {
	return atomic.LoadUint64(&p.running)
}

// fanout implements a generic fan-out for variadic channels
func fanout(randomizer func(max int) int, in <-chan funcdata, outs ...chan funcdata) {
	l := len(outs)
	for v := range in {
		i := randomizer(l)
		if i < 0 || i > l {
			i = rand.Intn(l)
		}
		outs[i] <- v
	}
}
