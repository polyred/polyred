package sched

import (
	"math/rand"
	"runtime"
	"sync/atomic"
)

// Pool is a worker pool.
type Pool struct {
	running    uint64
	taskQcap   int
	numWorkers int
	randomizer func(int, int) int
	done       chan struct{}
	workers    []chan funcdata
}

type funcdata struct {
	fn func()
	fg func(interface{})
	ar interface{}
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
		taskQcap:   runtime.GOMAXPROCS(0),
		numWorkers: runtime.GOMAXPROCS(0),
		done:       make(chan struct{}),
	}

	for _, opt := range opts {
		opt(p)
	}

	p.workers = make([]chan funcdata, p.numWorkers)
	for i := 0; i < p.numWorkers; i++ {
		p.workers[i] = make(chan funcdata, 1000)
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

func (p *Pool) RunWithArgs(f func(args interface{}), args interface{}) {
	ii := p.randomizer(0, p.numWorkers)
	p.workers[ii] <- funcdata{fg: f, ar: args}
}

func (p *Pool) Add(numTasks uint64) uint64 {
	return atomic.AddUint64(&p.running, numTasks)
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
