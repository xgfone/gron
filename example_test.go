package gron_test

import (
	"context"
	"fmt"
	"time"

	"github.com/xgfone/gron"
)

func jobRunner(name string) gron.Runner {
	return func(c context.Context, now time.Time) (result interface{}, err error) {
		fmt.Printf("Starting to run job '%s'\n", name)
		return
	}
}

func jobResultHook(result gron.JobResult) {
	fmt.Printf("End to run job '%s'\n", result.Job.Name())
}

func ExampleExecutor() {
	exe := gron.NewExecutor()
	exe.AppendResultHooks(jobResultHook) // Add the job result hook
	exe.Start()                          // Start the executor in the background goroutine.

	// Add jobs
	exe.Schedule("job1", gron.Every(time.Second*4), jobRunner("job1"))
	exe.Schedule("job2", gron.MustParseWhen("@every 4s"), jobRunner("job2"))
	everyMinuteScheduler := gron.MustParseWhen("*/4 * * * * *")
	exe.ScheduleJob(gron.NewJob("job3", everyMinuteScheduler, jobRunner("job3")))

	go func() {
		time.Sleep(time.Second * 10)
		// exe.CancelJobs("job1", "job2", "job3")
		exe.Stop()
	}()

	// Wait until the executor is stopped.
	exe.Wait()

	// Example Output:
	// Starting to run job 'job3'
	// End to run job 'job3'
	// Starting to run job 'job1'
	// End to run job 'job1'
	// Starting to run job 'job3'
	// End to run job 'job3'
	// Starting to run job 'job2'
	// End to run job 'job2'
	// Starting to run job 'job3'
	// End to run job 'job3'
	// Starting to run job 'job1'
	// End to run job 'job1'
	// Starting to run job 'job2'
	// End to run job 'job2'
}
