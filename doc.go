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
//   package main
//
//   import (
//   	"context"
//   	"fmt"
//   	"time"
//
//   	"github.com/xgfone/gron"
//   )
//
//   func job(ctx context.Context) (data []byte, err error) {
//   	// TODO:)
//   	fmt.Printf("Running job '%s'\n", gron.GetJobName(ctx))
//   	return
//   }
//
//   func jcb(task gron.Task, data []byte, err error) {
//   	// TODO:)
//   	fmt.Printf("Job callback: job '%s' end\n", task.Job.Name)
//   }
//
//   func gcb(task gron.Task, data []byte, err error) {
//   	// TODO:)
//   	fmt.Printf("Global callback: job '%s' end\n", task.Job.Name)
//   }
//
//   func main() {
//   	exe := gron.NewExecutor()
//   	exe.AppendJobResultHook(gcb)
//   	exe.AppendJobScheduleHook(func(j gron.Job) { fmt.Printf("Schedule job '%s'\n", j.Name) })
//   	exe.AppendJobCancelHook(func(j gron.Job) { fmt.Printf("Cancel job '%s'\n", j.Name) })
//
//   	// Add jobs
//   	exe.Schedule("job1", gron.Every(time.Minute), job)
//   	exe.Schedule("job2", gron.MustParseWhen("@every 3s"), job)
//
//   	job3 := gron.NewJob("job3").When(gron.MustParseWhen("@every 5s")).Runner(job)
//   	exe.ScheduleJob(job3)
//
//   	when := gron.MustParseWhen("@every 10s @total 10") // Only run the job for ten times
//   	job4 := gron.NewJob("job4").When(when).Timeout(time.Minute).Callback(jcb).Runner(job)
//   	exe.ScheduleJob(job4)
//
//   	// crontab syntax has not been implemented.
//   	//exe.Schedule("cronjob", gron.MustParseWhen("@at * * * * *"), job)
//   	//
//   	// It is equal to
//   	//exe.Schedule("cronjob", gron.MustParseWhen("@every 1m"), job)
//
//   	// Cancel jobs
//   	go func() {
//   		time.Sleep(time.Minute)
//   		exe.CancelJob("job1")
//   		exe.CancelJob("job2")
//   		exe.CancelJob("job3")
//   	}()
//
//   	// Wait until all the jobs end.
//   	exe.Wait()
//   }
//
package gron
