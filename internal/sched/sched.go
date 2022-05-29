// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package sched

import (
	"math/rand"
	"sync"
)

// Pool is a worker pool.
type Pool struct {
	numWorkers int
	wg         sync.WaitGroup
	randomizer func(int, int) int
	workers    []chan funcdata
}

type funcdata struct {
	fn func()
}

type Option func(w *Pool)

func Workers(limit int) Option {
	return func(w *Pool) {
		w.numWorkers = limit
	}
}

func Randomizer(f func(min, max int) int) Option {
	return func(w *Pool) {
		w.randomizer = f
	}
}

// TODO: figure out if we can optimize cache misses?

func New(opts ...Option) *Pool {
	p := &Pool{
		randomizer: func(min, max int) int {
			return rand.Intn(max)
		},
	}

	for _, opt := range opts {
		opt(p)
	}

	p.workers = make([]chan funcdata, p.numWorkers)
	for i := 0; i < len(p.workers); i++ {
		p.workers[i] = make(chan funcdata, 128)
	}

	// Start workers
	for i := 0; i < len(p.workers); i++ {
		go func(workerId int) {
			for d := range p.workers[workerId] {
				if d.fn != nil {
					d.fn()
				}
				p.wg.Done()
			}
		}(i)
	}

	return p
}

// Run runs f in the current pool.
func (p *Pool) Run(f ...func()) {
	p.wg.Add(len(f))
	for i := range f {
		ii := p.randomizer(0, p.numWorkers)
		p.workers[ii] <- funcdata{fn: f[i]}
	}
}

func (p *Pool) Wait() {
	p.wg.Wait()
}

func (p *Pool) Release() {
	for i := range p.workers {
		close(p.workers[i])
	}
}
