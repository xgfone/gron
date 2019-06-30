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

package main

import (
	"bytes"
	"context"
	"errors"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"github.com/xgfone/klog"

	"github.com/xgfone/gron"
	"github.com/xgfone/ship"
)

func newWorkerHandler(jobScheduleHooks, jobCancelHooks, jobResultHooks []string,
	timeout time.Duration) workerHandler {
	exe := gron.NewExecutor()

	var jobSchedules []func(gron.Job)
	for _, u := range jobScheduleHooks {
		if _, err := url.Parse(u); err != nil {
			panic(err)
		}
		jobSchedules = append(jobSchedules, newHTTPJobScheduleFunc(u))
	}

	var jobCacnels []func(gron.Job)
	for _, u := range jobCancelHooks {
		if _, err := url.Parse(u); err != nil {
			panic(err)
		}
		jobCacnels = append(jobCacnels, newHTTPJobCancelFunc(u))
	}

	var jobResults []func(gron.Task, []byte, error)
	for _, u := range jobResultHooks {
		if _, err := url.Parse(u); err != nil {
			panic(err)
		}
		jobResults = append(jobResults, newHTTPJobResultFunc(u))
	}

	exe.AppendJobResultHook(jobResults...)
	exe.AppendJobCancelHook(jobCacnels...)
	exe.AppendJobScheduleHook(jobSchedules...)

	exe.AppendJobScheduleHook(func(job gron.Job) { klog.K("job", job.Name).Infof("add the job") })
	exe.AppendJobCancelHook(func(job gron.Job) { klog.K("job", job.Name).Infof("remove the job") })
	exe.AppendJobResultHook(func(task gron.Task, data []byte, err error) {
		log := klog.K("job", task.Job.Name).K("time", task.Prev).K("cost", task.Cost)
		if err == nil {
			log.Infof("run the job")
		} else {
			log.E(err).Errorf("failed to run the job")
		}
	})
	return workerHandler{timeout: timeout, executor: exe}
}

type workerHandler struct {
	timeout  time.Duration
	executor *gron.Executor
}

func (h workerHandler) AddJob(ctx *ship.Context) (err error) {
	type retryT struct {
		Number   int           `json:"number"`
		Interval time.Duration `json:"interval"`
	}
	type taskT struct {
		Name     string        `json:"name"`
		When     string        `json:"when"`
		Retry    retryT        `json:"retry"`
		Timeout  time.Duration `json:"timeout"`
		Callback string        `json:"callback"`
		Runner   string        `json:"runner"`
	}

	var task taskT
	var when gron.When
	var runner func(context.Context) ([]byte, error)
	var callback func(gron.Task, []byte, error)

	if err = ctx.Bind(&task); err != nil {
		return ctx.String(400, err.Error())
	} else if task.Name == "" {
		return ctx.String(400, "missing name")
	} else if task.Runner == "" {
		return ctx.String(400, "missing runner")
	} else if task.When == "" {
		return ctx.String(400, "missing when")
	} else if when, err = gron.ParseWhen(task.When); err != nil {
		return ctx.String(400, err.Error())
	} else if task.Callback != "" {
		if _, err := url.Parse(task.Callback); err != nil {
			return ctx.String(400, "invalid callback: %s", err.Error())
		}
		callback = newHTTPJobResultFunc(task.Callback)
	} else if strings.HasPrefix(task.Runner, "shell ") {
		if cmd := strings.TrimSpace(task.Runner[6:]); cmd != "" {
			runner = newShellJobRunner(cmd)
		} else {
			return ctx.String(400, "invalid runner")
		}
	}

	if task.Timeout <= 0 && h.timeout > 0 {
		task.Timeout = h.timeout
	}

	job := gron.Job{
		Name:     task.Name,
		When:     when,
		Run:      runner,
		Retry:    gron.Retry{Number: task.Retry.Number, Interval: task.Retry.Interval},
		Timeout:  task.Timeout,
		Callback: callback,
	}

	if ok := h.executor.ScheduleJob(job); !ok {
		return ctx.NoContent(409)
	}

	return nil
}

func (h workerHandler) DeleteJob(ctx *ship.Context) error {
	name := ctx.Param("name")
	if _, ok := h.executor.CancelJob(name); !ok {
		return ctx.NoContent(404)
	}

	return nil
}

func (h workerHandler) GetJob(ctx *ship.Context) error {
	name := ctx.Param("name")
	if task, ok := h.executor.GetTask(name); ok {
		return ctx.JSON(200, task)
	}
	return ctx.NoContent(404)
}

func (h workerHandler) GetJobs(ctx *ship.Context) error {
	return ctx.JSON(200, map[string]interface{}{"tasks": h.executor.GetAllTasks()})
}

//////

func newHTTPJobScheduleFunc(url string) func(gron.Job) {
	return nil
}

func newHTTPJobCancelFunc(url string) func(gron.Job) {
	return nil
}

func newHTTPJobResultFunc(url string) func(gron.Task, []byte, error) {
	return nil
}

func newShellJobRunner(cmd string) func(context.Context) ([]byte, error) {
	return func(ctx context.Context) (data []byte, err error) {
		cmd := exec.CommandContext(ctx, "/bin/sh", "-c", cmd)
		var output bytes.Buffer
		var errput bytes.Buffer
		cmd.Stdout = &output
		cmd.Stderr = &errput

		if err = cmd.Run(); err != nil {
			if stderr := errput.Bytes(); len(stderr) > 0 {
				err = errors.New(string(stderr))
			} else if stdout := output.Bytes(); len(stdout) > 0 {
				err = errors.New(string(stdout))
			}
		} else {
			data = output.Bytes()
		}

		return
	}
}
