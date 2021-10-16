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

import (
	"fmt"
	"strings"
	"time"

	"github.com/xgfone/gron/crontab"
)

// CronSpecParser is the parser to parse the cron specification expression.
var CronSpecParser = crontab.NewParser(crontab.SecondOptional |
	crontab.Minute |
	crontab.Hour |
	crontab.Dom |
	crontab.Month |
	crontab.Dow |
	crontab.Descriptor)

// Every is equal to ParseWhen("@every interval").
func Every(interval time.Duration) When {
	return &when{spec: "@every " + interval.String(), every: interval}
}

// EveryWithDelay is the same as Every, but delay to run the job for delay duration.
func EveryWithDelay(interval, delay time.Duration) When {
	return &when{spec: "@every " + interval.String(), every: interval, delay: delay}
}

// MustParseWhen is the same as ParseWhen, but panic if there is an error.
func MustParseWhen(s string) When {
	w, err := ParseWhen(s)
	if err != nil {
		panic(err)
	}
	return w
}

// ParseWhen parses the spec to a when implementation, which uses and returns
// the UTC time.
//
// spec supports CRON Expression Format, see https://en.wikipedia.org/wiki/Cron,
// which also supports one of several pre-defined schedules like this:
//
//     Entry                  | Description                                | Equivalent To
//     -----                  | -----------                                | -------------
//     @yearly (or @annually) | Run once a year, midnight, Jan. 1st        | 0 0 1 1 *
//     @monthly               | Run once a month, midnight, first of month | 0 0 1 * *
//     @weekly                | Run once a week, midnight between Sat/Sun  | 0 0 * * 0
//     @daily (or @midnight)  | Run once a day, midnight                   | 0 0 * * *
//     @hourly                | Run once an hour, beginning of hour        | 0 * * * *
//     @every Duration        | Run once every interval duration           |
//
func ParseWhen(spec string) (When, error) {
	spec = strings.TrimSpace(spec)
	if strings.HasPrefix(spec, "@every ") {
		interval, err := time.ParseDuration(spec[7:])
		if err != nil {
			return nil, fmt.Errorf("When: invalid 'every' argument: '%s'", spec[7:])
		}
		return &when{spec: spec, every: interval}, nil
	}

	scheduler, err := CronSpecParser.Parse(spec)
	if err != nil {
		return nil, err
	}

	scheduler.Location = time.UTC
	return &when{spec: spec, sched: scheduler}, nil
}

type when struct {
	spec  string
	delay time.Duration
	every time.Duration
	sched *crontab.SpecSchedule
}

func (w *when) String() string { return w.spec }

func (w *when) Next(prev time.Time) (next time.Time) {
	if prev.IsZero() { // First time
		prev = time.Now().UTC()
	}

	if w.every > 0 {
		return prev.Add(w.every)
	}

	return w.sched.Next(prev)
}
