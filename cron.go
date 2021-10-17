// Copyright 2021 xgfone
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gron

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// When represents when a job is executed.
type When interface {
	// String returns the string representation.
	String() string

	// Next returns the next UTC time when the job is executed.
	//
	// If prev is ZERO, it indicates that it's the first time.
	Next(prev time.Time) (next time.Time)
}

// Job represents a task job.
type Job interface {
	Name() string
	When() When
	Run(c context.Context, now time.Time) (result interface{}, err error)
}

// JobResultHook is the hook to handle the result of the job task.
type JobResultHook func(JobResult)

// JobResult is the result of the job.
type JobResult struct {
	Job    Job
	Next   time.Time
	Start  time.Time
	Cost   time.Duration
	Result interface{}
	Error  error
}

// Task represents a job task.
type Task struct {
	Job       Job
	Prev      time.Time     // The last time to be executed
	Next      time.Time     // The next time to be executed
	Cost      time.Duration // The cost duration to execute the job last time
	IsRunning bool          // Indicate whether the job is running
}

type task struct {
	Job     Job
	Prev    time.Time
	Next    time.Time
	Cost    int64
	Running uint32
}

func (t *task) IsRunning() bool    { return atomic.LoadUint32(&t.Running) == 1 }
func (t *task) SetCost(cost int64) { atomic.StoreInt64(&t.Cost, cost) }
func (t *task) SetRunning(running bool) {
	if running {
		atomic.StoreUint32(&t.Running, 1)
	} else {
		atomic.StoreUint32(&t.Running, 0)
	}
}

func (t *task) Task() Task {
	return Task{
		Job:       t.Job,
		Prev:      t.Prev,
		Next:      t.Next,
		Cost:      time.Duration(atomic.LoadInt64(&t.Cost)),
		IsRunning: t.IsRunning(),
	}
}

type tasks []*task

func (ts tasks) Len() int           { return len(ts) }
func (ts tasks) Less(i, j int) bool { return ts[j].Next.After(ts[i].Next) }
func (ts tasks) Swap(i, j int)      { ts[i], ts[j] = ts[j], ts[i] }

// Executor represents a task executor.
type Executor struct {
	exit  chan struct{}
	stop  chan struct{}
	hooks []JobResultHook

	lock  sync.RWMutex
	timeo time.Duration
	tasks tasks
}

// NewExecutor returns a new job executor.
func NewExecutor() *Executor {
	return &Executor{
		exit:  make(chan struct{}),
		stop:  make(chan struct{}),
		tasks: make([]*task, 0, 64),
	}
}

// SetTimeout resets the timeout of all the jobs to execute the job.
func (e *Executor) SetTimeout(timeout time.Duration) {
	e.lock.Lock()
	e.timeo = timeout
	e.lock.Unlock()
}

// AppendResultHooks appends the some result hooks.
func (e *Executor) AppendResultHooks(hooks ...JobResultHook) {
	e.hooks = append(e.hooks, hooks...)
}

// Schedule is equal to e.ScheduleJob(NewJob(name, when, runner)).
func (e *Executor) Schedule(name string, when When, runner Runner) error {
	return e.ScheduleJob(NewJob(name, when, runner))
}

// ScheduleJob adds a job to wait to be scheduled.
func (e *Executor) ScheduleJob(job Job) (err error) {
	name := job.Name()
	next := job.When().Next(time.Time{}).UTC()

	var ok bool
	e.lock.Lock()
	for _len := len(e.tasks) - 1; _len >= 0; _len-- {
		if e.tasks[_len].Job.Name() == name {
			ok = true
			break
		}
	}
	if !ok {
		e.tasks = append(e.tasks, &task{Job: job, Next: next})
		sort.Sort(e.tasks)
	}
	e.lock.Unlock()

	if ok {
		err = fmt.Errorf("the job named '%s' has been added", name)
	}

	return
}

// CancelJobs cancels the jobs by the given names.
func (e *Executor) CancelJobs(names ...string) {
	if len(names) > 0 {
		e.lock.Lock()
		e.cancelJobs(names...)
		e.lock.Unlock()
	}
	return
}

func (e *Executor) cancelJobs(names ...string) {
	for i := len(e.tasks) - 1; i >= 0; i-- {
		if inStrings(e.tasks[i].Job.Name(), names) {
			_len := len(e.tasks)
			copy(e.tasks[i:_len], e.tasks[i+1:_len])
			e.tasks = e.tasks[:_len-1]
		}
	}
	return
}

func inStrings(s string, ss []string) bool {
	for _, _s := range ss {
		if _s == s {
			return true
		}
	}
	return false
}

// GetTask returns the job task by the name.
func (e *Executor) GetTask(name string) (task Task, ok bool) {
	e.lock.RLock()
	for i := len(e.tasks) - 1; i >= 0; i-- {
		if e.tasks[i].Job.Name() == name {
			task = e.tasks[i].Task()
			ok = true
			break
		}
	}
	e.lock.RUnlock()
	return
}

// GetAllTasks returns the tasks of all the jobs.
func (e *Executor) GetAllTasks() (tasks []Task) {
	e.lock.RLock()
	_len := len(e.tasks)
	tasks = make([]Task, _len)
	for _len--; _len >= 0; _len-- {
		tasks[_len] = e.tasks[_len].Task()
	}
	e.lock.RUnlock()
	return
}

// Wait waits until the executor is stopped.
func (e *Executor) Wait() { <-e.exit }

// Run starts the executor in the current goroutine, which does not return
// until the executor is stopped.
func (e *Executor) Run() { e.loop() }

// Start starts the executor in the background goroutine.
func (e *Executor) Start() { go e.loop() }

// Stop stops all the jobs in the executor.
func (e *Executor) Stop() {
	select {
	case <-e.stop:
	default:
		close(e.stop)
	}
}

func (e *Executor) loop() {
	names := make([]string, 0, 8)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-e.stop:
			close(e.exit)
			return

		case now := <-ticker.C:
			now = now.UTC()
			var ok bool
			e.lock.Lock()
			for _, task := range e.tasks {
				if task.Next.After(now) {
					break
				}

				if task.IsRunning() {
					continue
				}

				next := task.Job.When().Next(now).UTC()
				if next.IsZero() {
					names = append(names, task.Job.Name())
					continue
				}

				ok = true
				task.Cost = 0
				task.Prev = now
				task.Next = next
				task.SetRunning(true)
				go e.runTask(task, now, next, e.timeo)
			}

			if len(names) > 0 {
				e.cancelJobs(names...)
				names = names[:0]
			}

			if ok {
				sort.Sort(e.tasks)
			}
			e.lock.Unlock()
		}
	}
}

func (e *Executor) runTask(task *task, now, next time.Time, timeout time.Duration) {
	defer task.SetRunning(false)
	result, err := e.runJob(task.Job, now, timeout)
	cost := time.Since(now)
	task.SetCost(int64(cost))

	if len(e.hooks) > 0 {
		jobResult := JobResult{
			Job:    task.Job,
			Next:   next,
			Start:  now,
			Cost:   cost,
			Result: result,
			Error:  err,
		}

		for _, hook := range e.hooks {
			hook(jobResult)
		}
	}
}

func (e *Executor) runJob(job Job, now time.Time, timeout time.Duration) (
	data interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("panic: %v", e)
		}
	}()

	var cancel func()
	ctx := context.Background()
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	data, err = job.Run(ctx, now)
	return
}
