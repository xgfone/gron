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

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Every is equal to ParseWhen("@every d").
func Every(d time.Duration) When {
	return &whenT{orig: "@every " + d.String(), every: d}
}

// MustParseWhen is the same as ParseWhen, but panic if there is an error.
func MustParseWhen(s string) When {
	w, err := ParseWhen(s)
	if err != nil {
		panic(err)
	}
	return w
}

// ParseWhen parses s to a when implementation.
//
// s is made of directives, which follow behind a character '@' and maybe have
// a argument. And the directives and its arguments are separated by one or more
// spaces.
//
// The supported directives have
//   @base  now/prev // The next time is based on the now or previous(default) time to be calculated.
//   @every Duration // The next time is calculated by adding the interval duration.
//   @delay Duration // For the first time, it will be delayed back for the duration.
//   @end   RFC3339  // When the deadline is reached, the job will end.
//   @total int(>0)  // When the job is run for total times, it will end.
//   @once           // It is equal to "@total 1".
//   @at    crontab  // Crontab format, which has not been implemented yet.
//
// Here are the few quick references about crontab simple but powerful syntax.
//
//    *     *     *     *     *
//
//    ^     ^     ^     ^     ^
//    |     |     |     |     |
//    |     |     |     |     +----- day of week (0-6) (Sunday=0)
//    |     |     |     +------- month (1-12)
//    |     |     +--------- day of month (1-31)
//    |     +----------- hour (0-23)
//    +------------- min (0-59)
//
// Examples
//
//   *  *  * * *  run on every minute
//   10 *  * * *  run at 0:10, 1:10 etc
//   10 15 * * *  run at 15:10 every day
//   *  *  1 * *  run on every minute on 1st day of month
//   0  0  1 1 *  Happy new year schedule
//   0  0  * * 1  Run at midnight on every Monday
//
// Lists
//
//   1-15 *        * * *  run at 1, 2, 3...15 minute of each hour
//   0    0-5,10   * * *  run on every hour from 0-5 and in 10 oclock
//   *    10,15,19 * * *  run at 10:00, 15:00 and 19:00
//
// Steps
//
//    */2    *   *   * *  run every two minutes
//    10     */3 *   * *  run every 3 hours on 10th min
//    0      12  */2 * *  run at noon on every two days
//    1-59/2 *   *   * *  run every two minutes, but on odd minutes
//
func ParseWhen(s string) (When, error) {
	if s = strings.TrimSpace(s); len(s) < 2 {
		return nil, fmt.Errorf("When: missing directive")
	} else if s[0] != '@' {
		return nil, fmt.Errorf("When: directive must start with @")
	}

	directives := make(map[string]string, 32)
	for _, item := range strings.Split(s[1:], "@") {
		if item = strings.TrimSpace(item); item == "" {
			return nil, fmt.Errorf("When: missing directive")
		}

		var arg string
		directive := item
		if index := strings.IndexByte(item, ' '); index > 0 {
			directive = item[:index]
			arg = strings.TrimSpace(item[index+1:])
		}
		if _, ok := directives[directive]; ok {
			return nil, fmt.Errorf("When: reduplicative '%s' directive", directive)
		}
		directives[directive] = arg
	}

	var basenow bool
	var total int
	var end time.Time
	var every time.Duration
	var delay time.Duration

	for directive, arg := range directives {
		// TODO: check the compatibility between directives
		switch directive {
		case "base":
			switch arg {
			case "now":
				basenow = true
			case "prev":
			default:
				return nil, fmt.Errorf("When: invalid 'base' directive argument: '%s'", arg)
			}
		case "total":
			n, err := strconv.ParseUint(arg, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("When: invalid 'total' directive argument: '%s'", arg)
			}
			total = int(n)
		case "once":
			if arg != "" {
				return nil, fmt.Errorf("When: 'once' directive must not have the argument")
			}
			total = 1
		case "end":
			t, err := time.Parse(time.RFC3339, arg)
			if err != nil {
				return nil, fmt.Errorf("When: invalid 'end' directive argument: '%s'", arg)
			}
			end = t
		case "every":
			d, err := time.ParseDuration(arg)
			if err != nil {
				return nil, fmt.Errorf("When: invalid 'every' directive argument: '%s'", arg)
			}
			every = d
		case "delay":
			d, err := time.ParseDuration(arg)
			if err != nil {
				return nil, fmt.Errorf("When: invalid 'delay' directive argument: '%s'", arg)
			}
			delay = d

		// case "at": // for cron format
		// case "yearly":
		// case "monthly":
		// case "weekly":
		// case "daily", "midnight":
		// case "hourly":
		// case "minutely":
		default:
			return nil, fmt.Errorf("invalid when directive '%s'", directive)
		}
	}

	return &whenT{
		orig:    s,
		basenow: basenow,
		total:   total,
		every:   every,
		delay:   delay,
		end:     end,
	}, nil
}

type whenT struct {
	orig string

	count int
	total int
	end   time.Time
	delay time.Duration
	every time.Duration

	basenow bool
}

func (w *whenT) String() string {
	return w.orig
}

func (w *whenT) validateNext(next time.Time) time.Time {
	if !w.end.IsZero() && next.After(w.end) {
		return time.Time{}
	} else if !w.basenow {
		if now := time.Now(); now.After(next) {
			return now
		}
	}
	return next
}

func (w *whenT) Next(prev time.Time) (next time.Time) {
	// End
	if w.total > 0 {
		if w.count >= w.total {
			return
		}
		w.count++
	}

	// First time
	if prev.IsZero() {
		next = time.Now()
		if w.delay > 0 {
			next = next.Add(w.delay)
		}
		if !w.end.IsZero() && next.After(w.end) {
			return time.Time{}
		}
		return
	} else if w.basenow {
		prev = time.Now()
	}

	if w.delay > 0 {
		prev = prev.Add(w.delay)
	}

	if w.every > 0 {
		return w.validateNext(prev.Add(w.every))
	}

	return w.validateNext(next)
}
