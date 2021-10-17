package gron

import (
	"sort"
	"testing"
	"time"
)

func TestTasks(t *testing.T) {
	ts := tasks{
		nil,
		&task{Job: NewJob("job4", nil, nil), Next: time.Unix(4, 0)},
		&task{Job: NewJob("job2", nil, nil), Next: time.Unix(2, 0)},
		nil,
		&task{Job: NewJob("job1", nil, nil), Next: time.Unix(1, 0)},
		&task{Job: NewJob("job3", nil, nil), Next: time.Unix(3, 0)},
		nil,
	}
	sort.Stable(ts)

	for i, task := range ts {
		if i >= 4 {
			if task != nil {
				t.Errorf("expected job 'nil', but got job '%s'", task.Job.Name())
			}
		} else if v := task.Next.Unix(); v != int64(i+1) {
			t.Errorf("expected job%d, but got job%d", i+1, v)
		}
	}
}

func TestCancelJobs(t *testing.T) {
	exe := NewExecutor()
	exe.Schedule("job1", MustParseWhen("@every 1m"), nil)
	exe.Schedule("job2", MustParseWhen("@every 1m"), nil)
	exe.Schedule("job3", MustParseWhen("@every 1m"), nil)
	exe.Schedule("job4", MustParseWhen("@every 1m"), nil)
	exe.CancelJobs("job2", "job4")

	tasks := exe.GetAllTasks()
	if len(tasks) != 2 {
		t.Fatal(tasks)
	}

	for _, task := range tasks {
		switch name := task.Job.Name(); name {
		case "job1", "job3":
		default:
			t.Errorf("unexpected job '%s'", name)
		}
	}
}
