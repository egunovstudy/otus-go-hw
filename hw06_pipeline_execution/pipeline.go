package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	stream := wrapWithDone(in, done)
	for _, stage := range stages {
		stream = wrapWithDone(stage(stream), done)
	}
	return stream
}

func wrapWithDone(in In, done In) Out {
	if in == nil {
		return nil
	}

	out := make(Bi)

	if done == nil {
		go func() {
			defer close(out)
			for v := range in {
				out <- v
			}
		}()
		return out
	}

	go func() {
		closed := false
		defer func() {
			if !closed {
				close(out)
			}
		}()

		for {
			select {
			case <-done:
				if !closed {
					close(out)
					closed = true
				}

				for range in {
					// выкидываем
				}
				return

			case v, ok := <-in:
				if !ok {
					return
				}
				if closed {
					continue
				}

				select {
				case <-done:
					close(out)
					closed = true
					for range in {
					}
					return
				case out <- v:
				}
			}
		}
	}()

	return out
}
