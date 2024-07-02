package jobpool

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

type JobPool[T any] struct {
	jobFunc   func(T)
	wg        *sync.WaitGroup
	jobChan   chan *T
	closed    *atomic.Bool
	closeOnce sync.Once
}

func NewJobPool[T any](jobFunc func(T), concurrency int) *JobPool[T] {
	job := &JobPool[T]{
		jobFunc: jobFunc,
		wg:      &sync.WaitGroup{},
		jobChan: make(chan *T, concurrency*10),
		closed:  &atomic.Bool{},
	}

	for i := 0; i < concurrency; i++ {
		go job.processJobs()
	}

	return job
}

func (j *JobPool[T]) processJobs() {
	for job := range j.jobChan {
		j.jobFunc(*job)
		j.wg.Done()
	}
}

func (j *JobPool[T]) Wait(ctx context.Context) {
	j.wg.Wait()
}

func (j *JobPool[T]) Process(jobFunc *T) error {

	if j.closed.Load() {
		return errors.New("job pool is closed")
	}

	j.wg.Add(1)

	j.jobChan <- jobFunc

	return nil
}

func (j *JobPool[T]) Close(ctx context.Context) {
	j.closeOnce.Do(func() {
		j.closed.Store(true)
		close(j.jobChan)
		for len(j.jobChan) > 0 {
			<-j.jobChan
		}
		j.Wait(ctx)
	})
}
