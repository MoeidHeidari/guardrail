package analysis

import (
	"context"
	"runtime"
	"sync"
)

func parallelFor(ctx context.Context, count int, fn func(context.Context, int)) {
	if count <= 0 {
		return
	}
	workers := runtime.NumCPU()
	if workers > count {
		workers = count
	}
	jobs := make(chan int)
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for index := range jobs {
				fn(ctx, index)
			}
		}()
	}
	for i := 0; i < count; i++ {
		if ctx.Err() != nil {
			break
		}
		select {
		case jobs <- i:
		case <-ctx.Done():
		}
	}
	close(jobs)
	wg.Wait()
}
