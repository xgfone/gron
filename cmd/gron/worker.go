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
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"github.com/urfave/cli"
	"github.com/xgfone/gconf"
	"github.com/xgfone/go-tools/lifecycle"
	"github.com/xgfone/gron"
	"github.com/xgfone/gron/cmd/internal/logging"
	"github.com/xgfone/klog"
	"github.com/xgfone/ship"
)

var shellCMD = "/bin/sh"

var workerOpts = []gconf.Opt{
	gconf.StrOpt("addr", "The address to listen to.").D(":8002"),
	gconf.StrOpt("log-level", "The log level, such as debug, info, etc.").D("info"),
	gconf.StrOpt("log-file", "The log file path, and the log is output to os.Stdout by default."),
	gconf.StrOpt("shell", "The shell path.").D("/bin/sh"),
	gconf.DurationOpt("timeout", "The default global timeout.").D(time.Minute),
	gconf.StrSliceOpt("job-schedule-hooks", "The comma-separated url lists to be called when a job is scheduled."),
	gconf.StrSliceOpt("job-cancel-hooks", "The comma-separated url lists to be called when a job is canceled."),
	gconf.StrSliceOpt("job-result-hooks", "The comma-separated url lists to be called when a job is called."),
}

func init() {
	gconf.NewGroup("worker").RegisterOpts(workerOpts)
}

func getWorkerCommand() cli.Command {
	conf := gconf.Group("worker")

	return cli.Command{
		Name:  "worker",
		Usage: "An gron runner worker like crontab, but more powerful",
		Flags: gconf.ConvertOptsToCliFlags(conf),
		Action: func(ctx *cli.Context) error {
			gconf.LoadSource(gconf.NewCliSource(ctx, "worker"))

			err := logging.Init(conf.GetString("log-level"), conf.GetString("log-file"))
			if err != nil {
				return err
			}

			shellCMD = conf.GetString("shell")
			handler := newWorkerHandler(conf.GetStringSlice("job-schedule-hooks"),
				conf.GetStringSlice("job-cancel-hooks"),
				conf.GetStringSlice("job-result-hooks"),
				conf.GetDuration("timeout"))

			router := ship.New(ship.SetLogger(klog.ToFmtLoggerError(klog.Std)))
			router.R("/v1/job").POST(handler.AddJob).GET(handler.GetJobs)
			router.R("/v1/job/:name").GET(handler.GetJob).DELETE(handler.DeleteJob)
			router.Start(conf.GetString("addr")).Wait()
			return nil
		},
	}
}

func newWorkerHandler(jobScheduleHooks, jobCancelHooks, jobResultHooks []string,
	timeout time.Duration) workerHandler {
	exe := gron.NewExecutor()
	lifecycle.Register(exe.Close)

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

	var jobResults []func(gron.Task, interface{}, error)
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
	exe.AppendJobResultHook(func(task gron.Task, data interface{}, err error) {
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
	var runner func(context.Context) (interface{}, error)
	var callback func(gron.Task, interface{}, error)

	if err = ctx.Bind(&task); err != nil {
		return ctx.String(400, err.Error())
	} else if task.Name == "" {
		return ctx.String(400, "missing name")
	} else if task.When == "" {
		return ctx.String(400, "missing when")
	} else if when, err = gron.ParseWhen(task.When); err != nil {
		return ctx.String(400, err.Error())
	} else if task.Callback != "" {
		if _, err := url.Parse(task.Callback); err != nil {
			return ctx.String(400, "invalid callback: %s", err.Error())
		}
		callback = newHTTPJobResultFunc(task.Callback)
	} else if task.Retry.Number < 0 {
		return ctx.String(400, "the retry number must not be negative")
	}

	if task.Runner == "" {
		return ctx.String(400, "missing runner")
	} else if strings.HasPrefix(task.Runner, "shell ") { // Shell Runner
		if cmd := strings.TrimSpace(task.Runner[6:]); cmd != "" {
			runner = newShellJobRunner(cmd)
		} else {
			return ctx.String(400, "invalid shell runner")
		}
	} else if strings.HasPrefix(task.Runner, "bin ") { // Binary Runner
		if cmd := strings.TrimSpace(task.Runner[4:]); cmd != "" {
			runner = newBinJobRunner(cmd)
		} else {
			return ctx.String(400, "invalid bin runner")
		}
	} else {
		return ctx.String(400, "invalid runner")
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

//////////////////////////////////////////////////////////////////////////////

func newHTTPJobScheduleFunc(url string) func(gron.Job) {
	timeout := time.Second * 5
	return func(job gron.Job) {
		bs, _ := job.MarshalJSON()
		req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bs))
		req.Header.Set(ship.HeaderContentType, ship.MIMEApplicationJSONCharsetUTF8)
		req.Header.Set("X-Gron-Action", "schedule")
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		if resp, err := http.DefaultClient.Do(req.WithContext(ctx)); err != nil {
			klog.K("job", job.Name).K("url", url).E(err).Errorf("faild to send job schedule notice")
		} else {
			resp.Body.Close()
		}
		cancel()
	}
}

func newHTTPJobCancelFunc(url string) func(gron.Job) {
	timeout := time.Second * 5
	return func(job gron.Job) {
		bs, _ := job.MarshalJSON()
		req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bs))
		req.Header.Set(ship.HeaderContentType, ship.MIMEApplicationJSONCharsetUTF8)
		req.Header.Set("X-Gron-Action", "cancel")
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		if resp, err := http.DefaultClient.Do(req.WithContext(ctx)); err != nil {
			klog.K("job", job.Name).K("url", url).E(err).Errorf("faild to send job cancel notice")
		} else {
			resp.Body.Close()
		}
		cancel()
	}
}

type result struct {
	Result interface{} `json:"result,omitempty"`
	Error  error       `json:"err,omitempty"`
}

func newHTTPJobResultFunc(url string) func(gron.Task, interface{}, error) {
	return func(task gron.Task, data interface{}, err error) {
		bs, _ := json.Marshal(result{Result: data, Error: err})
		resp, err := http.Post(url, ship.MIMEApplicationJSONCharsetUTF8, bytes.NewBuffer(bs))
		if err != nil {
			klog.K("job", task.Job.Name).K("url", url).E(err).Errorf("faild to send job result")
		} else {
			resp.Body.Close()
		}
	}
}

func newShellJobRunner(cmd string) func(context.Context) (interface{}, error) {
	return newJobRunner(shellCMD, "-c", cmd)
}

func newBinJobRunner(cmd string) func(context.Context) (interface{}, error) {
	cmds := strings.Fields(cmd)
	return newJobRunner(cmds[0], cmds[1:]...)
}

func newJobRunner(name string, args ...string) func(context.Context) (interface{}, error) {
	return func(ctx context.Context) (data interface{}, err error) {
		cmd := exec.CommandContext(ctx, name, args...)
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
			data = base64.StdEncoding.EncodeToString(output.Bytes())
		}

		return
	}
}
