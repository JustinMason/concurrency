package jobpool

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

type JobPool[T any] struct {
	jobFunc func(T)
	wg      sync.WaitGroup
	jobChan chan *T
	closed  *atomic.Bool
}

func NewJobPool[T any](jobFunc func(T), concurrency int) *JobPool[T] {
	job := &JobPool[T]{
		jobFunc: jobFunc,
		wg:      sync.WaitGroup{},
		jobChan: make(chan *T, concurrency*10),
		closed:  &atomic.Bool{},
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

func (j *JobPool[T]) WaitTimeout(timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		j.wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
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
