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
	"os"

	"github.com/urfave/cli/v2"
	"github.com/xgfone/go-tools/v7/lifecycle"
	"github.com/xgfone/klog/v3"
)

func main() {
	app := cli.NewApp()
	app.Version = version
	app.Usage = "A crontab service"
	app.Commands = []*cli.Command{getWorkerCommand()}
	if err := app.Run(os.Args); err != nil {
		klog.Ef(err, "The program exits.")
	}
	lifecycle.Stop()
}
