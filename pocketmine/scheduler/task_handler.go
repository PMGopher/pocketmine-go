package scheduler

import (
	"fmt"
	"reflect"
)

// taskName mirrors Task::getName()'s default (Utils::getNiceClassName($this)) using Go's own
// reflection, unless the task implements namedTask to override it.
func taskName(task Task) string {
	if n, ok := task.(namedTask); ok {
		return n.Name()
	}
	t := reflect.TypeOf(task)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.String()
}

// TaskHandler is a port of pocketmine\scheduler\TaskHandler.
type TaskHandler struct {
	task      Task
	delay     int
	period    int
	nextRun   int
	cancelled bool
	taskName  string
	ownerName string
}

// NewTaskHandler mirrors the TaskHandler constructor. delay/period of -1 mean "not delayed" /
// "not repeating", matching TaskScheduler's own convention.
func NewTaskHandler(task Task, delay int, period int, ownerName string) (*TaskHandler, error) {
	if task.Handler() != nil {
		return nil, fmt.Errorf("cannot assign multiple handlers to the same task")
	}
	if ownerName == "" {
		ownerName = "Unknown"
	}
	h := &TaskHandler{task: task, delay: delay, period: period, taskName: taskName(task), ownerName: ownerName}
	task.setHandler(h)
	return h, nil
}

func (h *TaskHandler) IsCancelled() bool    { return h.cancelled }
func (h *TaskHandler) GetNextRun() int      { return h.nextRun }
func (h *TaskHandler) SetNextRun(ticks int) { h.nextRun = ticks }
func (h *TaskHandler) GetTask() Task        { return h.task }
func (h *TaskHandler) GetDelay() int        { return h.delay }
func (h *TaskHandler) IsDelayed() bool      { return h.delay > 0 }
func (h *TaskHandler) IsRepeating() bool    { return h.period > 0 }
func (h *TaskHandler) GetPeriod() int       { return h.period }
func (h *TaskHandler) GetTaskName() string  { return h.taskName }
func (h *TaskHandler) GetOwnerName() string { return h.ownerName }

func (h *TaskHandler) Cancel() {
	if !h.IsCancelled() {
		h.task.OnCancel()
	}
	h.remove()
}

func (h *TaskHandler) remove() {
	h.cancelled = true
	h.task.setHandler(nil)
}

// Run executes the task. A returned *CancelTaskError cancels it; any other non-nil error is
// treated as an uncaught failure (mirroring an exception propagating out of onRun() uncaught in
// PHP) and panics, rather than being silently swallowed.
func (h *TaskHandler) Run() {
	err := h.task.OnRun()
	if err == nil {
		return
	}
	if _, ok := err.(*CancelTaskError); ok {
		h.Cancel()
		return
	}
	panic(err)
}
