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
	"time"
)

// JobBuilder is used to build a job.
type JobBuilder struct {
	name string
	when When

	retry    Retry
	timeout  time.Duration
	callback func(Task, []byte, error)
}

// NewJob returns a new job builder.
func NewJob(name string) JobBuilder {
	return JobBuilder{}.Name(name)
}

// Name returns a new job builder with the name.
func (jb JobBuilder) Name(name string) JobBuilder {
	if name == "" {
		panic("the job name must be empty")
	}
	jb.name = name
	return jb
}

// When returns a new job builder with the when.
func (jb JobBuilder) When(when When) JobBuilder {
	if when == nil {
		panic("the when must not be nil")
	}
	jb.when = when
	return jb
}

// Retry returns a new job builder with the retry number and the retry interval.
func (jb JobBuilder) Retry(number int, interval time.Duration) JobBuilder {
	if number < 0 {
		panic("the retry number must not be a negetive integer.")
	}
	jb.retry = Retry{Number: number, Interval: interval}
	return jb
}

// Timeout returns a new job builder with the timeout.
func (jb JobBuilder) Timeout(timeout time.Duration) JobBuilder {
	jb.timeout = timeout
	return jb
}

// Callback returns a new job builder with the callback function.
func (jb JobBuilder) Callback(cb func(Task, []byte, error)) JobBuilder {
	if cb == nil {
		panic("the job callback must not be nil")
	}
	jb.callback = cb
	return jb
}

// Runner returns a new job.
func (jb JobBuilder) Runner(run Runner) Job {
	if run == nil {
		panic("the runner must not be nil")
	}
	return Job{
		Name: jb.name,
		When: jb.when,
		Run:  run,

		Retry:    jb.retry,
		Timeout:  jb.timeout,
		Callback: jb.callback,
	}
}
