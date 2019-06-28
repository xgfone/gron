# gron [![Build Status](https://travis-ci.org/xgfone/gron.svg?branch=master)](https://travis-ci.org/xgfone/gron) [![GoDoc](https://godoc.org/github.com/xgfone/gron?status.svg)](http://godoc.org/github.com/xgfone/gron) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square)](https://raw.githubusercontent.com/xgfone/gron/master/LICENSE)

Another job periodic runner like `crontab`. See [GoDoc](https://godoc.org/github.com/xgfone/gron).

The supported Go version: `1.x`.

## Example

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/xgfone/gron"
)

func job(ctx context.Context) (data []byte, err error) {
	// TODO:)
	fmt.Printf("Running job '%s'\n", gron.GetJobName(ctx))
	return
}

func jcb(task gron.Task, data []byte, err error) {
	// TODO:)
	fmt.Printf("Job callback: job '%s' end\n", task.Job.Name)
}

func gcb(task gron.Task, data []byte, err error) {
	// TODO:)
	fmt.Printf("Global callback: job '%s' end\n", task.Job.Name)
}

func main() {
	exe := gron.NewExecutor()
	exe.AppendJobResultHook(gcb)
	exe.AppendJobScheduleHook(func(j gron.Job) { fmt.Printf("Schedule job '%s'\n", j.Name) })
	exe.AppendJobCancelHook(func(j gron.Job) { fmt.Printf("Cancel job '%s'\n", j.Name) })

	// Add jobs
	exe.Schedule("job1", gron.Every(time.Minute), job)
	exe.Schedule("job2", gron.MustParseWhen("@every 3s"), job)

	job3 := gron.NewJob("job3").When(gron.MustParseWhen("@every 5s")).Runner(job)
	exe.ScheduleJob(job3)

	when := gron.MustParseWhen("@every 10s @total 10") // Only run the job for ten times
	job4 := gron.NewJob("job4").When(when).Timeout(time.Minute).Callback(jcb).Runner(job)
	exe.ScheduleJob(job4)

	// crontab syntax has not been implemented.
	//exe.Schedule("cronjob", gron.MustParseWhen("@at * * * * *"), job)
	//
	// It is equal to
	//exe.Schedule("cronjob", gron.MustParseWhen("@every 1m"), job)

	// Cancel jobs
	go func() {
		time.Sleep(time.Minute)
		exe.CancelJob("job1")
		exe.CancelJob("job2")
		exe.CancelJob("job3")
	}()

	// Wait until all the jobs end.
	exe.Wait()
}
```
