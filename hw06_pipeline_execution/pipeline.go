package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	stream := in
	for _, stage := range stages {
		stream = stage(stream)
	}
	return wrapWithDone(stream, done)
}

func wrapWithDone(in In, done In) Out {
	if done == nil {
		return in
	}

	out := make(Bi)

	go func() {
		defer close(out)

		for {
			select {
			case <-done:
				drain(in)
				return

			case v, ok := <-in:
				if !ok {
					return
				}

				select {
				case <-done:
					drain(in)
					return
				case out <- v:
				}
			}
		}
	}()

	return out
}

func drain(in In) {
	for range in {
		// intentionally drained
	}
}
