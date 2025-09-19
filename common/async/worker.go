package async

import "sync"

func NewWorker[T any](workersCount int, exec func(T)) (chan<- T, <-chan struct{}) {
	ch := make(chan T, workersCount)
	done := make(chan struct{}, 1)
	go func() {
		var wg sync.WaitGroup
		for i := 0; i < workersCount; i++ {
			wg.Add(1)
			go func() {
				for dat := range ch {
					exec(dat)
				}
				wg.Done()
			}()
		}
		wg.Wait()
		done <- struct{}{}
		close(done)
	}()
	return ch, done
}
