// Copyright 2019 xgfone
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
	"bytes"
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

type contextType struct{}

var ctxJobName = contextType{}

// GetJobName returns the job name from the context, which is used by the runner
// in general.
func GetJobName(ctx context.Context) string {
	name, _ := ctx.Value(ctxJobName).(string)
	return name
}

// ErrCanceled is returned when the job is running but canceled.
var ErrCanceled = fmt.Errorf("the job is canceled")

// Runner is used to represent the runner of the job.
type Runner func(context.Context) (data []byte, err error)

// When represents when a job is executed.
type When interface {
	// String returns the string representation.
	String() string

	// Next returns the next time when the job is executed.
	Next(prev time.Time) (next time.Time)
}

// Retry represents a retry policy.
type Retry struct {
	Number   int
	Interval time.Duration
}

// Job represents a task job.
type Job struct {
	// Required
	Name string
	When When
	Run  Runner

	// optional
	Retry    Retry
	Timeout  time.Duration
	Callback func(task Task, data []byte, err error)
}

// Task represents a job task.
type Task struct {
	Job  Job
	Prev time.Time     // the last time to be executed
	Next time.Time     // the next time to be executed
	Cost time.Duration // The time cost to execute the job last

	Running bool // Indicate the job is running
}

// MarshalJSON implements json.Marshaler.
func (t Task) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.Grow(256)
	buf.WriteByte('{')
	fmt.Fprintf(buf, `"prev":"%s",`, t.Prev.Format(time.RFC3339))
	fmt.Fprintf(buf, `"next":"%s",`, t.Next.Format(time.RFC3339))
	fmt.Fprintf(buf, `"cost":"%s",`, t.Cost.String())
	fmt.Fprintf(buf, `"running":%t,`, t.Running)

	buf.WriteString(`"job":{`)
	fmt.Fprintf(buf, `"name":"%s",`, t.Job.Name)
	fmt.Fprintf(buf, `"when":"%s",`, t.Job.When.String())
	fmt.Fprintf(buf, `"timeout":"%s",`, t.Job.Timeout.String())

	buf.WriteString(`"retry":{`)
	fmt.Fprintf(buf, `"number":%d,`, t.Job.Retry.Number)
	fmt.Fprintf(buf, `"interval":"%s"`, t.Job.Retry.Interval)
	buf.WriteString("}}}")
	return buf.Bytes(), nil
}

type taskResult struct {
	Job  Job
	Prev time.Time
	Next time.Time
	Cost time.Duration
	Data []byte
	Err  error
}

type jobTasks []Task

func (ts jobTasks) Len() int {
	return len(ts)
}

func (ts jobTasks) Less(i, j int) bool {
	return ts[j].Next.After(ts[i].Next)
}

func (ts jobTasks) Swap(i, j int) {
	t := ts[i]
	ts[i] = ts[j]
	ts[j] = t
}

// Executor represents a task executor.
type Executor struct {
	lock  *sync.RWMutex
	stop  chan struct{}
	tasks []Task

	addHooks    []func(Job)
	deleteHooks []func(Job)
	resultHooks []func(Task, []byte, error)

	adds    chan Job
	deletes chan Job
	results chan taskResult
}

// NewExecutor returns a new job executor.
func NewExecutor() *Executor {
	exe := &Executor{
		lock:    new(sync.RWMutex),
		stop:    make(chan struct{}, 1),
		tasks:   make([]Task, 0, 64),
		adds:    make(chan Job, 128),
		deletes: make(chan Job, 128),
		results: make(chan taskResult, 1024),
	}

	go exe.loop()
	go exe.watchJob()

	return exe
}

// AppendJobResultHook appends the hook to handle the result of the job.
func (exe *Executor) AppendJobResultHook(hook func(Task, []byte, error)) {
	if hook == nil {
		panic("the hook must not be nil")
	}

	exe.lock.Lock()
	exe.resultHooks = append(exe.resultHooks, hook)
	exe.lock.Unlock()
}

// AppendJobScheduleHook appends the hook to watch which job is scheduled.
func (exe *Executor) AppendJobScheduleHook(hook func(Job)) {
	if hook == nil {
		panic("the hook must not be nil")
	}

	exe.lock.Lock()
	exe.addHooks = append(exe.addHooks, hook)
	exe.lock.Unlock()
}

// AppendJobCancelHook appends the hook to watch which job is canceled.
func (exe *Executor) AppendJobCancelHook(hook func(Job)) {
	if hook == nil {
		panic("the hook must not be nil")
	}

	exe.lock.Lock()
	exe.deleteHooks = append(exe.deleteHooks, hook)
	exe.lock.Unlock()
}

// Schedule is equal to
//   ScheduleJob(Job{Name: name, When: when, Run: run})
func (exe *Executor) Schedule(name string, when When, run Runner) (ok bool) {
	return exe.ScheduleJob(Job{Name: name, When: when, Run: run})
}

// ScheduleJob adds a job and returns true, but it will do nothing and return
// false if the job name has existed.
//
// Notice:
//   1. the job won't be run immediately until the next clock tick comes.
//      And a clock tick is one second, so the interval duration to run the job
//      periodically must not be less than one second.
//   2. If the next time to run the job reaches, but the last job has been
//      running, it will wait the next clock tick to be run.
func (exe *Executor) ScheduleJob(job Job) (ok bool) {
	if job.Name == "" {
		panic("the job name must not be empty")
	} else if job.When == nil {
		panic("the job when must not be nil")
	} else if job.Run == nil {
		panic("the job runner must not be nil")
	}

	next := job.When.Next(time.Time{})
	if next.IsZero() {
		return false
	}

	exe.lock.Lock()
	for i := range exe.tasks {
		if exe.tasks[i].Job.Name == job.Name {
			ok = true
			break
		}
	}
	if !ok {
		exe.tasks = append(exe.tasks, Task{Job: job, Next: next})
		sort.Sort(jobTasks(exe.tasks))
	}
	exe.lock.Unlock()

	if !ok {
		exe.adds <- job
	}
	return !ok
}

// CancelJob cancels the job task by its name.
func (exe *Executor) CancelJob(name string) (task Task, ok bool) {
	exe.lock.Lock()
	task, ok = exe.cancelJob(name)
	exe.lock.Unlock()
	return
}

func (exe *Executor) cancelJob(name string) (task Task, ok bool) {
	for i := range exe.tasks {
		if exe.tasks[i].Job.Name == name {
			ok = true
			task = exe.tasks[i]

			_len := len(exe.tasks) - 1
			copy(exe.tasks[i:_len], exe.tasks[i+1:])
			exe.tasks = exe.tasks[:_len]

			break
		}
	}

	if ok {
		exe.deletes <- task.Job
	}
	return
}

// GetTask returns the job task by its name.
func (exe *Executor) GetTask(name string) (task Task, ok bool) {
	exe.lock.Lock()
	for i := range exe.tasks {
		if exe.tasks[i].Job.Name == name {
			ok = true
			task = exe.tasks[i]
			break
		}
	}
	exe.lock.Unlock()
	return
}

// GetAllTasks returns the tasks of all the jobs.
func (exe *Executor) GetAllTasks() []Task {
	exe.lock.RLock()
	tasks := append([]Task(nil), exe.tasks...)
	exe.lock.RUnlock()
	return tasks
}

// Wait waits until all the jobs end.
func (exe *Executor) Wait() {
	for {
		exe.lock.RLock()
		n := len(exe.tasks)
		exe.lock.RUnlock()
		if n == 0 {
			break
		}
		time.Sleep(time.Second)
	}
}

// Close stops all the jobs in the executor.
func (exe *Executor) Close() {
	select {
	case <-exe.stop:
	default:
		close(exe.stop)
	}
}

func (exe *Executor) loop() {
	names := make([]string, 0, 8)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-exe.stop:
			return
		case now := <-ticker.C:
			var ok bool
			exe.lock.RLock()
			for i, task := range exe.tasks {
				if task.Next.After(now) {
					break
				} else if !task.Running {
					ok = true
					next := task.Job.When.Next(task.Next)
					if next.IsZero() {
						names = append(names, task.Job.Name)
					} else {
						exe.tasks[i].Running = true
						exe.tasks[i].Prev = now
						exe.tasks[i].Next = next
					}
					go exe.runJob(task.Job, now, next)
				}
			}

			if len(names) > 0 {
				for _, name := range names {
					exe.cancelJob(name)
				}
				names = names[:0]
			}

			if ok {
				sort.Sort(jobTasks(exe.tasks))
			}
			exe.lock.RUnlock()
		}
	}
}

func (exe *Executor) setTaskStatus(name string, running bool) {
	exe.lock.Lock()
	for i := range exe.tasks {
		if exe.tasks[i].Job.Name == name {
			exe.tasks[i].Running = running
			break
		}
	}
	exe.lock.Unlock()
}

func (exe *Executor) runJob(job Job, prev, next time.Time) {
	var cancel func()
	ctx := context.WithValue(context.Background(), ctxJobName, job.Name)
	if job.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, job.Timeout)
		defer cancel()
	}

	data, err := exe.retryToRunJob(ctx, job)
	exe.setTaskStatus(job.Name, false)
	exe.results <- taskResult{
		Job:  job,
		Prev: prev,
		Next: next,
		Cost: time.Now().Sub(prev),
		Data: data,
		Err:  err,
	}
}

func (exe *Executor) retryToRunJob(ctx context.Context, job Job) (data []byte, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("job '%s' panic: %v", job.Name, err)
		}
	}()

	for number := job.Retry.Number; number > -1; number-- {
		if data, err = job.Run(ctx); err == nil {
			return
		}
	}
	return
}

func (exe *Executor) watchJob() {
	rcbs := make([]func(Task, []byte, error), 0, 64)
	jcbs := make([]func(Job), 0, 8)
	for {
		select {
		case <-exe.stop:
			return
		case job := <-exe.adds:
			exe.lock.RLock()
			jcbs = append(jcbs[:0], exe.addHooks...)
			exe.lock.RUnlock()
			for _, cb := range jcbs {
				exe.runJobHook(cb, job)
			}
		case job := <-exe.deletes:
			exe.lock.RLock()
			jcbs = append(jcbs[:0], exe.deleteHooks...)
			exe.lock.RUnlock()
			for _, cb := range jcbs {
				exe.runJobHook(cb, job)
			}
		case result := <-exe.results:
			task := Task{
				Job:  result.Job,
				Prev: result.Prev,
				Next: result.Next,
				Cost: result.Cost,
			}

			// Run the job callback
			if result.Job.Callback != nil {
				exe.runJobCallback(result.Job.Callback, task, result.Data, result.Err)
			}

			// Run the global job callback
			exe.lock.RLock()
			rcbs = append(rcbs[:0], exe.resultHooks...)
			exe.lock.RUnlock()
			for _, cb := range rcbs {
				exe.runJobCallback(cb, task, result.Data, result.Err)
			}
		}
	}
}

func (exe *Executor) runJobHook(cb func(Job), job Job) {
	defer recover()
	cb(job)
}

func (exe *Executor) runJobCallback(cb func(Task, []byte, error), t Task, d []byte, e error) {
	defer recover()
	cb(t, d, e)
}
