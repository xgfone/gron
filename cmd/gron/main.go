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
	"time"

	"github.com/urfave/cli"
	"github.com/xgfone/gconf"
	"github.com/xgfone/gron/cmd/internal/logging"
	"github.com/xgfone/klog"
	"github.com/xgfone/ship"
)

var version = "1.0.0"

var workerOpts = []gconf.Opt{
	gconf.StrOpt("addr", "The address to listen to.").D(":8002"),
	gconf.StrOpt("log-level", "The log level, such as debug, info, etc.").D("info"),
	gconf.StrOpt("log-file", "The log file path."),
	gconf.DurationOpt("timeout", "The default global timeout.").D(time.Minute),
	gconf.StrSliceOpt("job-schedule-hooks", "The comma-separated url lists to be called when a job is scheduled."),
	gconf.StrSliceOpt("job-cancel-hooks", "The comma-separated url lists to be called when a job is canceled."),
	gconf.StrSliceOpt("job-result-hooks", "The comma-separated url lists to be called after a job is called."),
}

func main() {
	worker := gconf.NewGroup("worker")
	worker.RegisterOpts(workerOpts)

	app := cli.NewApp()
	app.Version = version
	app.Usage = ""
	app.Commands = []cli.Command{
		cli.Command{
			Name:  "worker",
			Flags: gconf.ConvertOptsToCliFlags(worker),
			Action: func(ctx *cli.Context) error {
				if err := gconf.LoadSource(gconf.NewCliSource(ctx, "worker")); err != nil {
					return err
				}

				err := logging.Init(worker.GetString("log-level"), worker.GetString("log-file"))
				if err != nil {
					return err
				}

				handler := newWorkerHandler(worker.GetStringSlice("job-schedule-hooks"),
					worker.GetStringSlice("job-cancel-hooks"),
					worker.GetStringSlice("job-result-hooks"),
					worker.GetDuration("timeout"))

				router := ship.New(ship.SetLogger(klog.ToFmtLoggerError(klog.Std)))
				router.R("/job").POST(handler.AddJob).GET(handler.GetJobs)
				router.R("/job/:name").GET(handler.GetJob).DELETE(handler.DeleteJob)
				router.Start(worker.GetString("addr")).Wait()
				return nil
			},
		},
	}
	app.RunAndExitOnError()
}
