package main

import (
	"testing"
	"time"
)

func TestQueue(t *testing.T) {
	t.Run("Queue.AddTasks adds tasks to the queue to be run", func(t *testing.T) {
		q := Queue{
			Interval: time.Millisecond,
			Retries:  0,
		}

		q.AddTasks([]func() bool{
			func() bool { return true },
			func() bool { return false },
			func() bool { return true },
		})

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
}
