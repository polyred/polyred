// Copyright 2021 Changkun Ou. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

package utils

// DefaultLimit is the default Conccurrent limit
const DefaultLimit = 100

// ConccurLimiter object
type ConccurLimiter struct {
	limit   int
	tickets chan int
}

// NewConccurLimiter allocates a new ConccurLimiter
func NewConccurLimiter(limit int) *ConccurLimiter {
	if limit <= 0 {
		limit = DefaultLimit
	}

	// allocate a limiter instance
	c := &ConccurLimiter{
		limit:   limit,
		tickets: make(chan int, limit),
	}

	// allocate the tickets:
	for i := 0; i < c.limit; i++ {
		c.tickets <- i
	}

	return c
}

// Execute adds a function to the execution queue.
// if num of go routines allocated by this instance is < limit
// launch a new go routine to execute job
// else wait until a go routine becomes available
func (c *ConccurLimiter) Execute(job func()) int {
	ticket := <-c.tickets
	go func() {
		defer func() {
			c.tickets <- ticket
		}()
		job()
	}()
	return ticket
}

// Wait will block all the previously Executed jobs completed running.
// Note that calling the Wait function while keep calling Execute leads
// to un-desired race conditions
func (c *ConccurLimiter) Wait() {
	for i := 0; i < c.limit; i++ {
		<-c.tickets
	}
}
