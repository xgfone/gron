package gron

import (
	"sort"
	"testing"
	"time"
)

func TestTasks(t *testing.T) {
	ts := tasks{
		&task{Job: NewJob("job4", nil, nil), Next: time.Unix(4, 0)},
		&task{Job: NewJob("job2", nil, nil), Next: time.Unix(2, 0)},
		&task{Job: NewJob("job1", nil, nil), Next: time.Unix(1, 0)},
		&task{Job: NewJob("job3", nil, nil), Next: time.Unix(3, 0)},
	}
	sort.Stable(ts)

	for i, task := range ts {
		if v := task.Next.Unix(); v != int64(i+1) {
			t.Errorf("expected job%d, but got job%d", i+1, v)
		}
	}
}
