package pipeline

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProcessJobPool(t *testing.T) {

	s1 := func(in string) string {
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		out := in + " processed 1"

		return out
	}

	s2 := func(in string) string {
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		out := in + " processed 2"

		return out
	}

	s3 := func(in string) string {
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		out := in + " processed 3"

		return out
	}

	s := NewStage(4, s1, s2, s3)

	message := []string{"test0", "test1", "test2"}
	expected := []string{"test0 processed 1 processed 2 processed 3", "test1 processed 1 processed 2 processed 3", "test2 processed 1 processed 2 processed 3"}

	s.Process(message...)

	for i := 0; i < 3; i++ {
		out := <-s.outCh
		assert.Contains(t, expected, out)
	}
}
