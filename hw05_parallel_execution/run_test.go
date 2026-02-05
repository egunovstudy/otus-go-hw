package hw05parallelexecution

import (
	"errors"
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	t.Run("if were errors in first M tasks, than finished not more N+M tasks", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32

		for i := 0; i < tasksCount; i++ {
			err := fmt.Errorf("error from task %d", i)
			tasks = append(tasks, func() error {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				atomic.AddInt32(&runTasksCount, 1)
				return err
			})
		}

		workersCount := 10
		maxErrorsCount := 23
		err := Run(tasks, workersCount, maxErrorsCount)

		require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
		require.LessOrEqual(t, runTasksCount, int32(workersCount+maxErrorsCount), "extra tasks were started")
	})

	t.Run("tasks without errors", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		var sumTime time.Duration

		for i := 0; i < tasksCount; i++ {
			taskSleep := time.Millisecond * time.Duration(rand.Intn(100))
			sumTime += taskSleep

			tasks = append(tasks, func() error {
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})
		}

		workersCount := 5
		maxErrorsCount := 1

		start := time.Now()
		err := Run(tasks, workersCount, maxErrorsCount)
		elapsedTime := time.Since(start)
		require.NoError(t, err)

		require.Equal(t, int32(tasksCount), runTasksCount, "not all tasks were completed")
		require.LessOrEqual(t, int64(elapsedTime), int64(sumTime/2), "tasks were run sequentially?")
	})
}

func TestRun_AllTasksExecuted_NoErrors(t *testing.T) {
	t.Parallel()

	const total = 50
	var started int32

	tasks := make([]Task, 0, total)
	for i := 0; i < total; i++ {
		tasks = append(tasks, func() error {
			atomic.AddInt32(&started, 1)
			time.Sleep(2 * time.Millisecond)
			return nil
		})
	}

	err := Run(tasks, 4, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := int(atomic.LoadInt32(&started)); got != total {
		t.Fatalf("expected %d tasks executed, got %d", total, got)
	}
}

func TestRun_StopOnErrors_ExecutedNotMoreThanNPlusM_WhenFirstMFail(t *testing.T) {
	t.Parallel()

	const (
		n     = 5
		m     = 3
		total = 100
	)

	var started int32
	someErr := errors.New("boom")

	tasks := make([]Task, 0, total)
	for i := 0; i < total; i++ {
		tasks = append(tasks, func() error {
			atomic.AddInt32(&started, 1)

			time.Sleep(5 * time.Millisecond)

			if i < m {
				return someErr
			}
			return nil
		})
	}

	err := Run(tasks, n, m)
	if !errors.Is(err, ErrErrorsLimitExceeded) {
		t.Fatalf("expected ErrErrorsLimitExceeded, got %v", err)
	}

	gotStarted := int(atomic.LoadInt32(&started))
	if gotStarted > n+m {
		t.Fatalf("expected executed tasks <= %d (n+m), got %d", n+m, gotStarted)
	}
}

func TestRun_MLessOrEqualZero_IgnoreErrorsAndRunAll(t *testing.T) {
	t.Parallel()

	const total = 30
	var started int32
	someErr := errors.New("fail")

	tasks := make([]Task, 0, total)
	for i := 0; i < total; i++ {
		tasks = append(tasks, func() error {
			atomic.AddInt32(&started, 1)
			return someErr
		})
	}

	err := Run(tasks, 3, 0)
	if err != nil {
		t.Fatalf("expected nil (errors ignored), got %v", err)
	}
	if got := int(atomic.LoadInt32(&started)); got != total {
		t.Fatalf("expected %d tasks executed, got %d", total, got)
	}
}
