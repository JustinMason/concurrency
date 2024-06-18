package jobpool

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type ProcessorMock struct {
	mock.Mock
	timeout time.Duration
}

type testJob struct{}

type Processor interface {
	ProcessResults(job testJob)
}

func NewProcessorMock(timeout time.Duration) *ProcessorMock {
	return &ProcessorMock{
		timeout: timeout}
}

func (m *ProcessorMock) ProcessResults(job testJob) {
	m.Called(job)
}

func TestProcessJobPool(t *testing.T) {
	processorMock := new(ProcessorMock)
	processJobPool := NewJobPool(context.Background(), processorMock.ProcessResults, 2)

	testCases := []string{"test1", "test2", "test3"}
	processJobs := []*testJob{}

	for range testCases {
		process := &testJob{}
		processorMock.On("ProcessResults", mock.Anything).Return()
		processJobPool.Process(process)
		processJobs = append(processJobs, process)
	}

	processJobPool.Wait()

	for range processJobs {
		processorMock.AssertCalled(t, "ProcessResults", mock.Anything)
	}
}

func (m *ProcessorMock) ProcessResultsTimeout(processJob testJob) {
	m.Called(processJob)
	time.Sleep(time.Duration(500 * time.Millisecond))
}

func TestProcessJobPoolWithTimeout(t *testing.T) {
	processorMock := NewProcessorMock(time.Duration(200 * time.Millisecond))

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()
	processJobPool := NewJobPool(ctx, processorMock.ProcessResultsTimeout, 2)

	testCases := []string{"test1"}

	for range testCases {
		processJob := &testJob{}
		processorMock.On("ProcessResultsTimeout", mock.Anything).Return()
		processJobPool.Process(processJob)
	}

	processJobPool.Wait()
	// Check if the context was timed out
	err := ctx.Err()
	assert.Error(t, err, "Expected an error but got none")
	if err != nil {
		assert.True(t, errors.Is(err, context.DeadlineExceeded), "Expected a DeadlineExceeded error")
	}

}

func TestProcessJobPoolWithClose(t *testing.T) {
	processorMock := NewProcessorMock(time.Duration(200 * time.Millisecond))
	processJobPool := NewJobPool(context.Background(), processorMock.ProcessResultsTimeout, 2)

	processJob := &testJob{}
	processorMock.On("ProcessResultsTimeout", mock.Anything).Return()
	processJobPool.Process(processJob)
	processJobPool.Close()

	err := processJobPool.Process(processJob)
	if err == nil {
		t.Error("Expected error when adding job after close, got nil")
	}

}
