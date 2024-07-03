// Copyright 2024 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Worker pool create same application.
package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/kirill-scherba/workerpool"
)

func main() {

	const (
		numWorkers = 4
		numTasks   = 20
	)

	// Wait group for wait all jobs
	wg := new(sync.WaitGroup)

	// Create worker pool and jobs function
	wp := workerpool.New(numWorkers, numWorkers*10, func(w int, data int) {
		fmt.Printf("worker %d, job: %d\n", w, data)
		time.Sleep(10 * time.Millisecond)
		wg.Done()
	}, true)

	// Execute jobs in worker pool
	fmt.Printf("execute %d jobs on %d workers\n", numTasks, numWorkers)
	for i := range numTasks {
		wg.Add(1)
		wp.Job <- i
	}

	// Wait all jobs done and close worker pool
	wg.Wait()
	fmt.Printf("wokers counts: %v\n", wp.Wcm)
	wp.Close()
}
