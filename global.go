// Copyright 2021 xgfone
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

// AppendResultHooks is equal to DefaultExecutor.AppendResultHooks(hooks...).
func AppendResultHooks(hooks ...JobResultHook) {
	DefaultExecutor.AppendResultHooks(hooks...)
}

// Schedule is equal to DefaultExecutor.Schedule(name, when, runner).
func Schedule(name string, when When, runner Runner) error {
	return DefaultExecutor.Schedule(name, when, runner)
}

// ScheduleJob is equal to DefaultExecutor.ScheduleJob(job).
func ScheduleJob(job Job) error { return DefaultExecutor.ScheduleJob(job) }

// CancelJobs is equal to DefaultExecutor.CancelJobs(names...).
func CancelJobs(names ...string) { DefaultExecutor.CancelJobs(names...) }

// GetTask is equal to DefaultExecutor.GetTask(name).
func GetTask(name string) (Task, bool) { return DefaultExecutor.GetTask(name) }

// GetAllTasks is equal to DefaultExecutor.GetAllTasks().
func GetAllTasks() []Task { return DefaultExecutor.GetAllTasks() }

// Start is equal to DefaultExecutor.Start().
func Start() { DefaultExecutor.Start() }

// Stop is equal to DefaultExecutor.Stop().
func Stop() { DefaultExecutor.Stop() }

// Wait is equal to DefaultExecutor.Wait().
func Wait() { DefaultExecutor.Wait() }
