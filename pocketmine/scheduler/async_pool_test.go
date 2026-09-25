package scheduler

import (
	"testing"
	"time"

	"pocketmine-go/pocketmine/log"
)

type sumTask struct {
	AsyncTaskBase
	n        int
	progress []any
	done     *int
}

func (t *sumTask) OnRun() {
	total := 0
	for i := 1; i <= t.n; i++ {
		total += i
	}
	t.PublishProgress("half")
	t.SetResult(total)
}

func (t *sumTask) OnProgressUpdate(p any) { t.progress = append(t.progress, p) }

func (t *sumTask) OnCompletion() { *t.done = t.GetResult().(int) }

type panicTask struct{ AsyncTaskBase }

func (t *panicTask) OnRun() { panic("boom") }

func collectAll(t *testing.T, p *AsyncPool) error {
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		more, err := p.CollectTasks()
		if err != nil || !more {
			return err
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("tasks never finished")
	return nil
}

func TestAsyncPoolRunsTasksAndCompletesOnCollect(t *testing.T) {
	p := NewAsyncPool(2, log.NewSimpleLogger())
	defer p.Shutdown()
	done := 0
	task := &sumTask{n: 100, done: &done}
	p.SubmitTask(task)
	if err := collectAll(t, p); err != nil {
		t.Fatal(err)
	}
	if done != 5050 || len(task.progress) != 1 {
		t.Fatalf("done=%d progress=%v", done, task.progress)
	}
}

func TestAsyncPoolReportsCrashes(t *testing.T) {
	p := NewAsyncPool(1, log.NewSimpleLogger())
	p.SubmitTask(&panicTask{})
	err := collectAll(t, p)
	if _, ok := err.(*AsyncWorkerCrashError); !ok {
		t.Fatalf("expected a crash error, got %v", err)
	}
}

func TestAsyncPoolRejectsDoubleSubmit(t *testing.T) {
	p := NewAsyncPool(1, log.NewSimpleLogger())
	defer p.Shutdown()
	done := 0
	task := &sumTask{n: 1, done: &done}
	p.SubmitTask(task)
	defer func() {
		if recover() == nil {
			t.Fatal("resubmitting a task must panic")
		}
	}()
	p.SubmitTask(task)
}
