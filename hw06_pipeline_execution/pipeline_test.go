package hw06pipelineexecution

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	sleepPerStage = time.Millisecond * 100
	fault         = sleepPerStage / 2
)

func TestPipeline(t *testing.T) {
	// Stage generator
	g := func(_ string, f func(v interface{}) interface{}) Stage {
		return func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					time.Sleep(sleepPerStage)
					out <- f(v)
				}
			}()
			return out
		}
	}

	stages := []Stage{
		g("Dummy", func(v interface{}) interface{} { return v }),
		g("Multiplier (* 2)", func(v interface{}) interface{} { return v.(int) * 2 }),
		g("Adder (+ 100)", func(v interface{}) interface{} { return v.(int) + 100 }),
		g("Stringifier", func(v interface{}) interface{} { return strconv.Itoa(v.(int)) }),
	}

	t.Run("simple case", func(t *testing.T) {
		in := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		start := time.Now()
		for s := range ExecutePipeline(in, nil, stages...) {
			result = append(result, s.(string))
		}
		elapsed := time.Since(start)

		require.Equal(t, []string{"102", "104", "106", "108", "110"}, result)
		require.Less(t,
			int64(elapsed),
			// ~0.8s for processing 5 values in 4 stages (100ms every) concurrently
			int64(sleepPerStage)*int64(len(stages)+len(data)-1)+int64(fault))
	})

	t.Run("done case", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		// Abort after 200ms
		abortDur := sleepPerStage * 2
		go func() {
			<-time.After(abortDur)
			close(done)
		}()

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		start := time.Now()
		for s := range ExecutePipeline(in, done, stages...) {
			result = append(result, s.(string))
		}
		elapsed := time.Since(start)

		require.Len(t, result, 0)
		require.Less(t, int64(elapsed), int64(abortDur)+int64(fault))
	})
}

func TestAllStageStop(t *testing.T) {
	wg := sync.WaitGroup{}
	// Stage generator
	g := func(_ string, f func(v interface{}) interface{}) Stage {
		return func(in In) Out {
			out := make(Bi)
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer close(out)
				for v := range in {
					time.Sleep(sleepPerStage)
					out <- f(v)
				}
			}()
			return out
		}
	}

	stages := []Stage{
		g("Dummy", func(v interface{}) interface{} { return v }),
		g("Multiplier (* 2)", func(v interface{}) interface{} { return v.(int) * 2 }),
		g("Adder (+ 100)", func(v interface{}) interface{} { return v.(int) + 100 }),
		g("Stringifier", func(v interface{}) interface{} { return strconv.Itoa(v.(int)) }),
	}

	t.Run("done case", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		// Abort after 200ms
		abortDur := sleepPerStage * 2
		go func() {
			<-time.After(abortDur)
			close(done)
		}()

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		for s := range ExecutePipeline(in, done, stages...) {
			result = append(result, s.(string))
		}
		wg.Wait()

		require.Len(t, result, 0)

	})
}

func TestPipelineNoStages(t *testing.T) {
	t.Run("pass through", func(t *testing.T) {
		in := make(Bi)
		go func() {
			defer close(in)
			for i := 0; i < 5; i++ {
				in <- i
			}
		}()

		got := make([]int, 0, 5)
		for v := range ExecutePipeline(in, nil /* done */) {
			got = append(got, v.(int))
		}
		require.Equal(t, []int{0, 1, 2, 3, 4}, got)
	})

	t.Run("done cancels", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)
		close(done)

		prodDone := make(chan struct{})
		go func() {
			defer close(prodDone)
			defer close(in)
			for i := 0; i < 100; i++ {
				in <- i
			}
		}()

		select {
		case _, ok := <-ExecutePipeline(in, done):
			require.False(t, ok)
		case <-time.After(200 * time.Millisecond):
			t.Fatal("pipeline did not close on done")
		}

		select {
		case <-prodDone:
			// ok
		case <-time.After(200 * time.Millisecond):
			t.Fatal("producer got stuck; pipeline probably didn't drain input")
		}
	})
}

func TestPipelineDrainsUpstreamOnDone(t *testing.T) {
	in := make(Bi)
	done := make(Bi)

	producerFinished := make(chan struct{})
	go func() {
		defer close(producerFinished)
		defer close(in)
		for i := 0; i < 10_000; i++ {
			in <- i
		}
	}()

	out := ExecutePipeline(in, done,
		func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for range in {
					time.Sleep(time.Millisecond)
					out <- 1
				}
			}()
			return out
		},
	)

	close(done)

	select {
	case _, ok := <-out:
		require.False(t, ok)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("output did not close after done")
	}

	select {
	case <-producerFinished:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("producer stuck; pipeline likely didn't drain upstream")
	}
}
