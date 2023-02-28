package main

import (
	"errors"
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
		runCountA := 0
		runCountB := 0

		q := NewQueue(
			[]func() error{
				func() error {
					runCountA += 1
					return errors.New("example error")
				},
				func() error {
					runCountB += 1
					return errors.New("example error")
				},
			},
			time.Millisecond,
			3,
		)

		q.Run()

		if runCountA != 3 {
			t.Errorf("want 3 retries, got %d", runCountA)
		}

		if runCountB != 3 {
			t.Errorf("want 3 retries, got %d", runCountB)
		}
	})
}
