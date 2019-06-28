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

func ExampleParseWhen() {
	// From now on, run the job once every one minute.
	ParseWhen("@every 1m")

	// From ten seconds later on, run the job once every one minute.
	ParseWhen("@every 1m @delay 10s")

	// Only run the job once immediately.
	ParseWhen("@once")

	// Only run the job once after ten seconds.
	ParseWhen("@delay 10s @once")

	// From now on, run the job once every one minute, but end with 2030-06-26T17:01:11+08:00.
	ParseWhen("@every 1m @end 2030-06-26T17:01:11+08:00")

	// From now on, run the job once every one minute, but calculate the next time by the current time, not the previous time.
	ParseWhen("@every 1m @base now")
}
