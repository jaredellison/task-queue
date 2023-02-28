package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"testing"
	"time"
)

func TestQueue(t *testing.T) {
	t.Run("NewQueue adds tasks to the queue to be run", func(t *testing.T) {
		q := NewQueue(
			[]func() error{
				func() error { return nil },
				func() error { return nil },
				func() error { return nil },
			},
			time.Millisecond,
			3,
		)

		taskCount := len(q.tasks)
		if taskCount != 3 {
			t.Errorf("want 3 tasks in queue, got %d", taskCount)
		}

		for _, tsk := range q.tasks {
			if tsk.Try != 0 {
				t.Errorf("expect tasks to start on try 0, got %d", tsk.Try)
			}
		}
	})

	t.Run("Queue.Run should retry funcs multiple times", func(t *testing.T) {
		totalFuncs := 50
		funcs, resultsChan := makeUnreliableFuncs(totalFuncs)

		retryCount := 3
		q := NewQueue(
			funcs,
			time.Millisecond,
			retryCount,
		)

		q.Run()

		results := []int{}

		for len(resultsChan) > 0 {
			r := <-resultsChan
			results = append(results, r)
		}

		sort.Ints(results)

		if len(results) != totalFuncs {
			t.Error("Expected one result from each function")
		}
		for i, r := range results {
			if i != r {
				t.Errorf("Expected each result to match index, got %v but want %v", r, i)
			}
		}
	})
}

func makeVariableTimeFunc(low time.Duration, high time.Duration, f func() error) func() error {
	return func() error {
		timeRange := high - low
		randDuration := time.Duration(rand.Int63n(int64(timeRange)))
		time.Sleep(low + randDuration)
		return f()
	}
}

func makeRetryableFunc(retryCount int, f func() error) func() error {
	attempt := 0
	return func() error {
		attempt += 1
		if attempt < retryCount {
			return errors.New(fmt.Sprintf("Example error, failed on try: %d", attempt))
		}
		f()
		return nil
	}
}

func makeVariableSuccessFunc(errRate float64, f func() error) func() error {
	return func() error {
		if rand.Float64() <= errRate {
			return errors.New("Test error")
		}
		f()
		return nil
	}
}

func makeUnreliableFuncs(count int) ([]func() error, chan int) {
	funcs := make([]func() error, count)
	results := make(chan int, count)
	minDuration := time.Microsecond * 1
	maxDuration := time.Millisecond * 1000

	for i := 0; i < count; i++ {
		result := i
		f := func() error {
			results <- result
			return nil
		}

		f = makeRetryableFunc(3, f)
		f = makeVariableTimeFunc(minDuration, maxDuration, f)
		funcs[i] = f
	}

	return funcs, results
}
