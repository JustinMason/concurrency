package jobpool

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

type JobPool[T any] struct {
	jobFunc func(T)
	wg      sync.WaitGroup
	jobChan chan *T
	closed  *atomic.Bool
	ctx     context.Context
}

func NewJobPool[T any](ctx context.Context, jobFunc func(T), concurrency int) *JobPool[T] {
	job := &JobPool[T]{
		jobFunc: jobFunc,
		wg:      sync.WaitGroup{},
		jobChan: make(chan *T, concurrency*10),
		closed:  &atomic.Bool{},
		ctx:     ctx,
	}

	for i := 0; i < concurrency; i++ {
		go func() {
			for j := range job.jobChan {
				job.jobFunc(*j)
				job.wg.Done()
			}
		}()
	}

	return job
}

func (j *JobPool[T]) Wait() {
	j.wg.Wait()
}

func (j *JobPool[T]) Process(jobFunc *T) error {
	j.wg.Add(1)

	if j.closed.Load() {
		return errors.New("job pool is closed")
	}

	j.jobChan <- jobFunc

	return nil
}

func (j *JobPool[T]) Close() {
	j.closed.Store(true)
	close(j.jobChan)
	for len(j.jobChan) > 0 {
		<-j.jobChan
	}
	j.wg.Wait()
}
