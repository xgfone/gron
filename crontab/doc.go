// Copyright (C) 2012 Rob Figueiredo
// All Rights Reserved.
//
// MIT LICENSE
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

/*
Package crontab implements the function of crontab, which is ported from
github.com/robfig/cron@v3.0.1.

CRON Expression Format

A cron expression represents a set of times, using 5 space-separated fields.

	Field name   | Mandatory? | Allowed values  | Allowed special characters
	----------   | ---------- | --------------  | --------------------------
	Minutes      | Yes        | 0-59            | * / , -
	Hours        | Yes        | 0-23            | * / , -
	Day of month | Yes        | 1-31            | * / , - ?
	Month        | Yes        | 1-12 or JAN-DEC | * / , -
	Day of week  | Yes        | 0-6 or SUN-SAT  | * / , - ?

Month and Day-of-week field values are case insensitive.  "SUN", "Sun", and
"sun" are equally accepted.

The specific interpretation of the format is based on the Cron Wikipedia page:
https://en.wikipedia.org/wiki/Cron

Alternative Formats

Alternative crontab expression formats support other fields like seconds.
You can implement that by creating a custom Parser as follows.

	NewParser(SecondOptional | Minute | Hour | Dom | Month | Dow | Descriptor)

Special Characters

Asterisk ( * )

The asterisk indicates that the cron expression will match for all values of the
field; e.g., using an asterisk in the 5th field (month) would indicate every
month.

Slash ( / )

Slashes are used to describe increments of ranges. For example 3-59/15 in the
1st field (minutes) would indicate the 3rd minute of the hour and every 15
minutes thereafter. The form "*\/..." is equivalent to the form "first-last/...",
that is, an increment over the largest possible range of the field.  The form
"N/..." is accepted as meaning "N-MAX/...", that is, starting at N, use the
increment until the end of that specific range.  It does not wrap around.

Comma ( , )

Commas are used to separate items of a list. For example, using "MON,WED,FRI" in
the 5th field (day of week) would mean Mondays, Wednesdays and Fridays.

Hyphen ( - )

Hyphens are used to define ranges. For example, 9-17 would indicate every
hour between 9am and 5pm inclusive.

Question mark ( ? )

Question mark may be used instead of '*' for leaving either day-of-month or
day-of-week blank.

Predefined schedules

You may use one of several pre-defined schedules in place of a cron expression.

	Entry                  | Description                                | Equivalent To
	-----                  | -----------                                | -------------
	@yearly (or @annually) | Run once a year, midnight, Jan. 1st        | 0 0 1 1 *
	@monthly               | Run once a month, midnight, first of month | 0 0 1 * *
	@weekly                | Run once a week, midnight between Sat/Sun  | 0 0 * * 0
	@daily (or @midnight)  | Run once a day, midnight                   | 0 0 * * *
	@hourly                | Run once an hour, beginning of hour        | 0 * * * *

Time zones

By default, all interpretation and scheduling is done in the machine's local
time zone (time.Local). You can specify a different time zone on construction:

	p := NewParser(SecondOptional | Minute | Hour | Dom | Month | Dow | Descriptor)
	p.Location = time.UTC

Individual cron schedules may also override the time zone they are to be
interpreted in by providing an additional space-separated field at the beginning
of the cron spec, of the form "CRON_TZ=Asia/Tokyo".

For example:

	# Runs at 6am in time.Local
	ParseStandard("0 6 * * ?")

	# Runs at 6am in America/New_York
	ParseStandard("CRON_TZ=America/New_York 0 6 * * ?")

The prefix "TZ=(TIME ZONE)" is also supported for legacy compatibility.
*/
package crontab
