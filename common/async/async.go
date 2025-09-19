package async

type rerr[T any] struct {
	r   T
	err error
}

type TaskAwaiter[T any] interface {
	Result() (T, error)
}

type taskAwaiterImpl[T any] struct {
	ch chan rerr[T]
}

func RunAsync[T any](exec func() (T, error)) TaskAwaiter[T] {
	ch := make(chan rerr[T], 1)
	go func() {
		r, err := exec()
		ch <- rerr[T]{
			r:   r,
			err: err,
		}
	}()

	return &taskAwaiterImpl[T]{
		ch: ch,
	}
}
func (tw *taskAwaiterImpl[T]) Result() (T, error) {
	r := <-tw.ch
	return r.r, r.err
}
