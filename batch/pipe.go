package batch

type (
	WorkFn[T any] func(T) T
	PipeFn[T any] func(chan T) chan T
)

func NewPipeFn[T any](fn WorkFn[T]) PipeFn[T] {
	return func(input chan T) chan T {
		output := make(chan T)
		go func() {
			defer close(output)
			for v := range input {
				output <- fn(v)
			}
		}()

		return output
	}
}

func Pipe[T any](input chan T, fns ...WorkFn[T]) chan T {
	output := input
	for _, fn := range fns {
		output = NewPipeFn(fn)(output)
	}
	return output
}
