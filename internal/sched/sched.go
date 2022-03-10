// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package sched

import (
	"math/rand"
	"sync/atomic"
)

// Pool is a worker pool.
type Pool struct {
	running    uint64
	numWorkers int
	randomizer func(int, int) int
	done       chan struct{}
	workers    []chan funcdata
}

type funcdata struct {
	fn func()
	fg func(any)
	ar any
}

// Opt is a scheduler option.
type Opt func(w *Pool)

func Workers(limit int) Opt {
	return func(w *Pool) {
		w.numWorkers = limit
	}
}

func Randomizer(f func(min, max int) int) Opt {
	return func(w *Pool) {
		w.randomizer = f
	}
}

// TODO: figure out if we can optimize cache misses?

// New creates a new task scheduler and returns a pool of workers.
//
func New(opts ...Opt) *Pool {
	p := &Pool{
		randomizer: func(min, max int) int {
			return rand.Intn(max)
		},
		running:    0,
		numWorkers: 4,
		done:       make(chan struct{}),
	}

	for _, opt := range opts {
		opt(p)
	}

	p.workers = make([]chan funcdata, p.numWorkers)
	for i := 0; i < p.numWorkers; i++ {
		p.workers[i] = make(chan funcdata, 128)
	}

	// Start workers
	for i := 0; i < p.numWorkers; i++ {
		go func(workerId int) {
			for d := range p.workers[workerId] {
				if d.fn != nil {
					d.fn()
				} else {
					d.fg(d.ar)
				}
				p.complete()
			}
		}(i)
	}

	return p
}

// Run runs f in the current pool.
func (p *Pool) Run(f ...func()) {
	for i := range f {
		ii := p.randomizer(0, p.numWorkers)
		p.workers[ii] <- funcdata{fn: f[i]}
	}
}

func (p *Pool) RunWithArgs(f func(args any), args any) {
	ii := p.randomizer(0, p.numWorkers)
	p.workers[ii] <- funcdata{fg: f, ar: args}
}

func (p *Pool) Add(numTasks int) int {
	return int(atomic.AddUint64(&p.running, uint64(numTasks)))
}

func (p *Pool) Running() uint64 {
	return atomic.LoadUint64(&p.running)
}

func (p *Pool) Wait() {
	<-p.done
}

func (p *Pool) Release() {
	for i := range p.workers {
		close(p.workers[i])
	}
}

func (p *Pool) complete() {
	ret := atomic.AddUint64(&p.running, ^uint64(0))
	if ret == 0 {
		p.done <- struct{}{}
	}
}
