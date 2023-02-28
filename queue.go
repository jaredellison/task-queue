package main

import (
	"fmt"
	"sync"
	"time"
)

type Queue struct {
	tasks      []task
	tasksLock  sync.Mutex
	Interval   time.Duration
	Retries    int
	status     map[string]bool
	statusLock sync.Mutex
}

type task struct {
	Id   string
	Run  func() error
	Try  int
	Done bool
}

func NewQueue(ts []func() error, interval time.Duration, retries int) *Queue {
	q := Queue{
		Interval: interval,
		status:   map[string]bool{},
		Retries:  retries,
	}

	for i, t := range ts {
		tsk := task{
			Run: t,
			Id:  fmt.Sprintf("%d", i), // Use UUID here?
		}
		q.tasks = append(q.tasks, tsk)
		q.status[tsk.Id] = false
	}

	return &q
}

func (q *Queue) GetTask() *task {
	q.tasksLock.Lock()
	defer q.tasksLock.Unlock()
	if len(q.tasks) == 0 {
		return nil
	}
	tsk := q.tasks[0]
	q.tasks = q.tasks[1:]
	return &tsk
}

func (q *Queue) RetryTask(t *task) {
	q.tasksLock.Lock()
	defer q.tasksLock.Unlock()
	q.tasks = append(q.tasks, *t)
}

func (q *Queue) MarkDone(id string) {
	q.statusLock.Lock()
	defer q.statusLock.Unlock()
	q.status[id] = true
}

func (q *Queue) CheckDone() bool {
	q.statusLock.Lock()
	defer q.statusLock.Unlock()
	for _, done := range q.status {
		if !done {
			return false
		}
	}
	return true
}

func (q *Queue) Run() {
	// Start Wait Group
	wg := sync.WaitGroup{}

	// Start ticker
	ticker := time.NewTicker(q.Interval)
	done := make(chan struct{})

	// Start processing tasks
	wg.Add(1)
	go func() {
		for {
			select {
			case <-done:
				wg.Done()
				return
			case <-ticker.C:
				// Pull task from queue
				tsk := q.GetTask()
				if tsk == nil {
					continue
				}
				// Increment try count
				tsk.Try++
				// Try to run it in a go routine
				go func() {
					err := tsk.Run()
					// Log error
					if err != nil {
						fmt.Printf("Error handling task %s: %v\n", tsk.Id, err)
					}

					// If success or reached retry limit we're done
					if err == nil || tsk.Try >= q.Retries {
						// Mark done in map
						q.MarkDone(tsk.Id)
						// Check if this is the last running task
						if q.CheckDone() {
							done <- struct{}{}
						}
						return
					}

					// Try again
					q.RetryTask(tsk)
				}()
			}
		}
	}()

	wg.Wait()
}
