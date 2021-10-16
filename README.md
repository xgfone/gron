# gron [![Build Status](https://github.com/xgfone/gron/actions/workflows/go.yml/badge.svg)](https://github.com/xgfone/gron/actions/workflows/go.yml) [![GoDoc](https://pkg.go.dev/badge/github.com/xgfone/gron)](https://pkg.go.dev/github.com/xgfone/gron) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square)](https://raw.githubusercontent.com/xgfone/gron/master/LICENSE)

Another job periodic runner like `crontab` supporting `Go1.11+`

## Example

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/xgfone/gron"
)

func jobRunner(name string) gron.Runner {
	return func(c context.Context, now time.Time) (result interface{}, err error) {
		fmt.Printf("Starting to run job '%s' at '%s'\n", name, now.Format(time.RFC3339Nano))
		return
	}
}

func jobResultHook(result gron.JobResult) {
	fmt.Printf("End to run job '%s', cost '%s'.\n", result.Job.Name(), result.Cost)
}

func main() {
	exe := gron.NewExecutor()
	exe.AppendResultHooks(jobResultHook) // Add the job result hook
	exe.Start()                          // Start the executor in the background goroutine.

	// Add jobs
	exe.Schedule("job1", gron.Every(time.Minute), jobRunner("job1"))
	exe.Schedule("job2", gron.MustParseWhen("@every 2m"), jobRunner("job2"))
	everyMinuteScheduler := gron.MustParseWhen("*/1 * * * *")
	exe.ScheduleJob(gron.NewJob("job3", everyMinuteScheduler, jobRunner("job3")))

	go func() {
		time.Sleep(time.Minute * 4)
		// exe.CancelJobs("job1", "job2", "job3")
		exe.Stop()
	}()

	// Wait until the executor is stopped.
	exe.Wait()
}
```
