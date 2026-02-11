package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	ch := in
	for _, stage := range stages {
		ch = stage(orDone(ch, done))
	}
	return orDone(ch, done)
}

func orDone(in In, done In) Out {
	if done == nil {
		return in
	}
	if in == nil {
		return nil
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
	for v := range in {
		_ = v // avoid revive empty-block warning; intentionally discard
	}
}
