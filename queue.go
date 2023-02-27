package main

import (
	"time"
)

type Queue struct {
	tasks    []task
	Interval time.Duration
	Retries  int
}

type task struct {
	Run func() bool
	Try int
}

func (q *Queue) AddTasks(ts []func() bool) {
	for _, t := range ts {
		q.tasks = append(q.tasks, task{
			Run: t,
		})
	}
}

func (q *Queue) Run() {

}
