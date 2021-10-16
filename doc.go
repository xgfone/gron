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

// Package gron supplies another job periodic runner.
//
// Example
//
//      package main
//
//      import (
//      	"context"
//      	"fmt"
//      	"time"
//
//      	"github.com/xgfone/gron"
//      )
//
//      func jobRunner(name string) gron.Runner {
//      	return func(c context.Context, now time.Time) (result interface{}, err error) {
//      		fmt.Printf("Starting to run job '%s' at '%s'\n", name, now.Format(time.RFC3339Nano))
//      		return
//      	}
//      }
//
//      func jobResultHook(result gron.JobResult) {
//      	fmt.Printf("End to run job '%s', cost '%s'.\n", result.Job.Name(), result.Cost)
//      }
//
//      func main() {
//      	exe := gron.NewExecutor()
//      	exe.AppendResultHooks(jobResultHook) // Add the job result hook
//      	exe.Start()                          // Start the executor in the background goroutine.
//
//      	// Add jobs
//      	exe.Schedule("job1", gron.Every(time.Minute), jobRunner("job1"))
//      	exe.Schedule("job2", gron.MustParseWhen("@every 2m"), jobRunner("job2"))
//      	everyMinuteScheduler := gron.MustParseWhen("*/1 * * * *")
//      	exe.ScheduleJob(gron.NewJob("job3", everyMinuteScheduler, jobRunner("job3")))
//
//      	go func() {
//      		time.Sleep(time.Minute * 4)
//      		// exe.CancelJobs("job1", "job2", "job3")
//      		exe.Stop()
//      	}()
//
//      	// Wait until the executor is stopped.
//      	exe.Wait()
//      }
//
package gron
