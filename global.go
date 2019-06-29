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

// DefaultExecutor is the default global executor.
var DefaultExecutor = NewExecutor()

// AppendJobCancelHook is equal to DefaultExecutor.AppendJobCancelHook(hook).
func AppendJobCancelHook(hook func(Job)) {
	DefaultExecutor.AppendJobCancelHook(hook)
}

// AppendJobResultHook is equal to DefaultExecutor.AppendJobResultHook(hook).
func AppendJobResultHook(hook func(Task, []byte, error)) {
	DefaultExecutor.AppendJobResultHook(hook)
}

// AppendJobScheduleHook is equal to DefaultExecutor.AppendJobScheduleHook(hook).
func AppendJobScheduleHook(hook func(Job)) {
	DefaultExecutor.AppendJobScheduleHook(hook)
}

// CancelJob is equal to DefaultExecutor.CancelJob(name).
func CancelJob(name string) (task Task, ok bool) {
	return DefaultExecutor.CancelJob(name)
}

// Close is equal to DefaultExecutor.Close().
func Close() {
	DefaultExecutor.Close()
}

// GetAllTasks is equal to DefaultExecutor.GetAllTasks().
func GetAllTasks() []Task {
	return DefaultExecutor.GetAllTasks()
}

// GetTask is equal to DefaultExecutor.GetTask(name).
func GetTask(name string) (task Task, ok bool) {
	return DefaultExecutor.GetTask(name)
}

// Schedule is equal to DefaultExecutor.Schedule(name, when, run).
func Schedule(name string, when When, run Runner) (ok bool) {
	return DefaultExecutor.Schedule(name, when, run)
}

// ScheduleJob is equal to DefaultExecutor.ScheduleJob(job).
func ScheduleJob(job Job) (ok bool) {
	return DefaultExecutor.ScheduleJob(job)
}

// Wait is equal to DefaultExecutor.Wait().
func Wait() {
	DefaultExecutor.Wait()
}
