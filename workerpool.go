// Copyright 2024 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Worker pool package create and use a worker pool to limit the number of
// Goroutines that run at the same time in Go application.
package workerpool

import (
	"sync"
)

// WorkerPool contains data and methods to work with worker pool.
type WokerPool[T any] struct {
	Job   chan T          // Jobs channel to execute task in workers pool
	Wcm   map[int]int     // Map with number of executed jobs by worker
	cnt   chan int        // Count jobs by worker channel
	count bool            // Count flag
	close chan any        // Close channel to close worker pool process
	wg    *sync.WaitGroup // Wait worker pool closed wait group
}

// New creates a new WorkerPool object.
//
// It creates a new worker pool with the specified number of workers and a
// buffered job channel. The job function is used to process the jobs sent to
// the worker pool. The count flag, if true, enables job count by worker.
func New[T any](numWorkers, channelBuffer int, job func(w int, data T),
	counts ...bool) *WokerPool[T] {

	// Detect count flag
	var count bool
	if len(counts) > 0 {
		count = counts[0]
	}

	// Create worker pool object
	wp := &WokerPool[T]{
		make(chan T, channelBuffer), // Jobs channel
		make(map[int]int),           // Worker count map
		make(chan int),              // Count jobs channel
		count,                       // Count flag
		make(chan any),              // Close channel
		&sync.WaitGroup{},           // Closed wait group
	}

	// Start worker count process if count flag is true
	if wp.count {
		go wp.workerCount()
	}

	// Close worker pool process
	wgc := new(sync.WaitGroup)
	wgc.Add(numWorkers)
	go func() {
		// Wait in close channel
		<-wp.close

		// Close jobs channel and wait all worker stopped
		close(wp.Job)
		wgc.Wait()
		wp.wg.Done()

		// If count flag is true, close count channel
		if wp.count {
			close(wp.cnt)
		}
	}()

	// Create and start workers
	wg := new(sync.WaitGroup)
	wg.Add(numWorkers)
	for w := range numWorkers {
		go func(w int) {
			defer wgc.Done() // Worker stopped
			wg.Done()        // Worker started

			// Loop to process jobs in the worker
			for data := range wp.Job {
				job(w, data)

				// Count job by worker if count flag is true
				if wp.count {
					wp.cnt <- w
				}
			}
		}(w)
	}
	wg.Wait() // Wait all workers started

	return wp
}

// Close worker pool.
func (wp *WokerPool[T]) Close() {
	wp.wg.Add(1) // Wait all workers stoped
	if wp.count {
		wp.wg.Add(1) // Wait worker count process stoped
	}
	wp.close <- struct{}{}
	wp.wg.Wait()
}

// workerCount counts executed jobs by worker.
func (wp *WokerPool[T]) workerCount() {
	for w := range wp.cnt {
		wp.Wcm[w]++
	}
	wp.wg.Done()
}
