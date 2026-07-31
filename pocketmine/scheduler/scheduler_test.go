package scheduler

import "testing"

type countingTask struct {
	TaskBase
	runs     int
	cancels  int
	failWith error
}

func (t *countingTask) OnRun() error {
	t.runs++
	return t.failWith
}

func (t *countingTask) OnCancel() { t.cancels++ }

func TestScheduleTaskRunsOnNextHeartbeat(t *testing.T) {
	s := NewTaskScheduler("test")
	task := &countingTask{}
	if _, err := s.ScheduleTask(task); err != nil {
		t.Fatalf("ScheduleTask() error = %v", err)
	}

	s.MainThreadHeartbeat(0)
	if task.runs != 1 {
		t.Fatalf("runs = %d, want 1", task.runs)
	}
	s.MainThreadHeartbeat(1)
	if task.runs != 1 {
		t.Fatalf("a non-repeating task ran again: runs = %d, want 1", task.runs)
	}
}

func TestScheduleDelayedTask(t *testing.T) {
	s := NewTaskScheduler("test")
	task := &countingTask{}
	if _, err := s.ScheduleDelayedTask(task, 5); err != nil {
		t.Fatalf("ScheduleDelayedTask() error = %v", err)
	}

	s.MainThreadHeartbeat(4)
	if task.runs != 0 {
		t.Fatalf("runs = %d before the delay elapsed, want 0", task.runs)
	}
	s.MainThreadHeartbeat(5)
	if task.runs != 1 {
		t.Fatalf("runs = %d after the delay elapsed, want 1", task.runs)
	}
}

func TestScheduleRepeatingTask(t *testing.T) {
	s := NewTaskScheduler("test")
	task := &countingTask{}
	if _, err := s.ScheduleRepeatingTask(task, 10); err != nil {
		t.Fatalf("ScheduleRepeatingTask() error = %v", err)
	}

	s.MainThreadHeartbeat(0)
	s.MainThreadHeartbeat(9)
	if task.runs != 1 {
		t.Fatalf("runs = %d before the period elapsed, want 1", task.runs)
	}
	s.MainThreadHeartbeat(10)
	if task.runs != 2 {
		t.Fatalf("runs = %d after one period, want 2", task.runs)
	}
	s.MainThreadHeartbeat(20)
	if task.runs != 3 {
		t.Fatalf("runs = %d after two periods, want 3", task.runs)
	}
}

func TestTaskSelfCancelViaCancelTaskError(t *testing.T) {
	s := NewTaskScheduler("test")
	task := &countingTask{failWith: &CancelTaskError{}}
	handle, _ := s.ScheduleRepeatingTask(task, 1)

	s.MainThreadHeartbeat(0)
	if task.runs != 1 || task.cancels != 1 {
		t.Fatalf("runs=%d cancels=%d, want runs=1 cancels=1", task.runs, task.cancels)
	}
	if !handle.IsCancelled() {
		t.Fatalf("expected the handler to be cancelled")
	}

	s.MainThreadHeartbeat(1)
	if task.runs != 1 {
		t.Fatalf("a self-cancelled task ran again: runs = %d, want 1", task.runs)
	}
}

func TestCancelAllTasks(t *testing.T) {
	s := NewTaskScheduler("test")
	task := &countingTask{}
	handle, _ := s.ScheduleRepeatingTask(task, 1)

	s.CancelAllTasks()
	if !handle.IsCancelled() {
		t.Fatalf("expected the task to be cancelled")
	}
	if task.cancels != 1 {
		t.Fatalf("cancels = %d, want 1", task.cancels)
	}

	s.MainThreadHeartbeat(100)
	if task.runs != 0 {
		t.Fatalf("a cancelled task ran: runs = %d, want 0", task.runs)
	}
}

func TestSchedulingOnDisabledSchedulerFails(t *testing.T) {
	s := NewTaskScheduler("test")
	s.Shutdown()

	if _, err := s.ScheduleTask(&countingTask{}); err == nil {
		t.Fatalf("expected an error scheduling on a disabled scheduler")
	}
}

func TestTaskCannotBeAssignedTwoHandlers(t *testing.T) {
	s1 := NewTaskScheduler("a")
	s2 := NewTaskScheduler("b")
	task := &countingTask{}

	if _, err := s1.ScheduleTask(task); err != nil {
		t.Fatalf("first ScheduleTask() error = %v", err)
	}
	if _, err := s2.ScheduleTask(task); err == nil {
		t.Fatalf("expected an error assigning a second handler to the same task")
	}
}

func TestClosureTaskRuns(t *testing.T) {
	var ran bool
	task := NewClosureTask(func() { ran = true })
	s := NewTaskScheduler("test")
	s.ScheduleTask(task)
	s.MainThreadHeartbeat(0)
	if !ran {
		t.Fatalf("expected the closure to have run")
	}
	if task.Name() == "" {
		t.Fatalf("expected a non-empty Name()")
	}
}
