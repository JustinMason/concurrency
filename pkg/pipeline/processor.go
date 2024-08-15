package pipeline

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

type Stage[T any] struct {
	stageFuncs  []StageFunc[T]
	wg          *sync.WaitGroup
	jobChan     chan T
	closed      *atomic.Bool
	closeOnce   sync.Once
	outCh       chan T
	concurrency int
}

func NewStage[T any](concurrency int, stageFuncs ...StageFunc[T]) *Stage[T] {
	job := &Stage[T]{
		stageFuncs:  stageFuncs,
		wg:          &sync.WaitGroup{},
		jobChan:     make(chan T, concurrency*10),
		outCh:       make(chan T),
		closed:      &atomic.Bool{},
		concurrency: concurrency,
	}

	for i := 0; i < concurrency; i++ {
		go job.setupPipeline(job.jobChan, job.outCh, job.stageFuncs)
	}

	return job
}

// Define the type for stage functions
type StageFunc[T any] func(T) T

// setupPipeline sets up the pipeline by chaining the stages
func (j *Stage[T]) setupPipeline(input <-chan T, out chan<- T, stages []StageFunc[T]) {
	for msg := range input {
		for _, stage := range stages {
			msg = stage(msg)
		}
		out <- msg
		j.wg.Done()
	}
}

func (j *Stage[T]) Wait(ctx context.Context) {
	j.wg.Wait()
}

func (j *Stage[T]) Process(jobData ...T) error {

	if j.closed.Load() {
		return errors.New("job pool is closed")
	}

	for _, job := range jobData {
		j.wg.Add(1)
		j.jobChan <- job
	}

	return nil
}

func (j *Stage[T]) Close(ctx context.Context) {
	j.closeOnce.Do(func() {
		j.closed.Store(true)
		close(j.jobChan)
		for len(j.jobChan) > 0 {
			<-j.jobChan
		}
		j.Wait(ctx)
	})
}
