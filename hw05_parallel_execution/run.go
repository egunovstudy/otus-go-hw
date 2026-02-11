package hw05parallelexecution

import (
	"errors"
	"math"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
//
// Assumption for m <= 0: ignore errors (run all tasks, return nil).
func Run(tasks []Task, n, m int) error {
	if n <= 0 {
		n = 1
	}

	// m <= 0 => ignore errors: run everything, always nil
	if m <= 0 {
		m = math.MaxInt // effectively "no limit"
	}

	// Safe limit conversion for comparisons with atomic int64.
	limit := int64(m)
	if limit < 0 { // just in case of weird overflow scenarios
		limit = math.MaxInt64
	}

	jobs := make(chan Task)     // unbuffered => bounded in-flight
	done := make(chan struct{}) // closed when error limit reached
	var once sync.Once

	var errCount int64
	var wg sync.WaitGroup
	wg.Add(n)

	worker := func() {
		defer wg.Done()
		for task := range jobs {
			if task == nil {
				continue
			}
			if err := task(); err != nil {
				if atomic.AddInt64(&errCount, 1) >= limit {
					once.Do(func() { close(done) })
				}
			}
		}
	}

	for i := 0; i < n; i++ {
		go worker()
	}

sendLoop:
	for _, task := range tasks {
		select {
		case <-done:
			break sendLoop
		case jobs <- task:
		}
	}

	close(jobs)
	wg.Wait()

	if atomic.LoadInt64(&errCount) >= limit && limit != int64(math.MaxInt) {
		// limit==MaxInt corresponds to m<=0 "ignore errors" mode; return nil there.
		return ErrErrorsLimitExceeded
	}
	return nil
}
