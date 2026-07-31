package scheduler

import (
	"fmt"

	"pocketmine-go/pocketmine/utils"
)

// TaskScheduler is a port of pocketmine\scheduler\TaskScheduler: a tick-driven scheduler for
// delayed and repeating tasks running on the main thread (as opposed to AsyncPool's worker
// threads — see the package doc comment for why that side isn't ported).
type TaskScheduler struct {
	enabled     bool
	queue       *utils.ReversePriorityQueue[int, *TaskHandler]
	tasks       *utils.ObjectSet[*TaskHandler]
	currentTick int
	owner       string
}

func NewTaskScheduler(owner string) *TaskScheduler {
	return &TaskScheduler{
		enabled: true,
		queue:   utils.NewReversePriorityQueue[int, *TaskHandler](),
		tasks:   utils.NewObjectSet[*TaskHandler](),
		owner:   owner,
	}
}

func (s *TaskScheduler) ScheduleTask(task Task) (*TaskHandler, error) {
	return s.addTask(task, -1, -1)
}

func (s *TaskScheduler) ScheduleDelayedTask(task Task, delay int) (*TaskHandler, error) {
	return s.addTask(task, delay, -1)
}

func (s *TaskScheduler) ScheduleRepeatingTask(task Task, period int) (*TaskHandler, error) {
	return s.addTask(task, -1, period)
}

func (s *TaskScheduler) ScheduleDelayedRepeatingTask(task Task, delay int, period int) (*TaskHandler, error) {
	return s.addTask(task, delay, period)
}

func (s *TaskScheduler) CancelAllTasks() {
	for h := range s.tasks.All() {
		h.Cancel()
	}
	s.tasks.Clear()
	for !s.queue.IsEmpty() {
		s.queue.Extract()
	}
}

func (s *TaskScheduler) IsQueued(h *TaskHandler) bool { return s.tasks.Contains(h) }

func (s *TaskScheduler) addTask(task Task, delay int, period int) (*TaskHandler, error) {
	if !s.enabled {
		return nil, fmt.Errorf("tried to schedule task to disabled scheduler")
	}

	if delay <= 0 {
		delay = -1
	}
	if period <= -1 {
		period = -1
	} else if period < 1 {
		period = 1
	}

	h, err := NewTaskHandler(task, delay, period, s.owner)
	if err != nil {
		return nil, err
	}
	return s.handle(h), nil
}

func (s *TaskScheduler) handle(h *TaskHandler) *TaskHandler {
	var nextRun int
	if h.IsDelayed() {
		nextRun = s.currentTick + h.GetDelay()
	} else {
		nextRun = s.currentTick
	}
	h.SetNextRun(nextRun)
	s.tasks.Add(h)
	s.queue.Insert(h, nextRun)
	return h
}

func (s *TaskScheduler) Shutdown() {
	s.enabled = false
	s.CancelAllTasks()
}

func (s *TaskScheduler) SetEnabled(enabled bool) { s.enabled = enabled }

// MainThreadHeartbeat is a port of TaskScheduler::mainThreadHeartbeat(): runs every task whose
// scheduled tick has arrived, called once per server tick.
func (s *TaskScheduler) MainThreadHeartbeat(currentTick int) error {
	if !s.enabled {
		return fmt.Errorf("cannot run heartbeat on a disabled scheduler")
	}
	s.currentTick = currentTick

	for s.isReady(currentTick) {
		task := s.queue.Extract()
		if task.IsCancelled() {
			s.tasks.Remove(task)
			continue
		}
		task.Run()
		if !task.IsCancelled() && task.IsRepeating() {
			task.SetNextRun(s.currentTick + task.GetPeriod())
			s.queue.Insert(task, s.currentTick+task.GetPeriod())
		} else {
			task.remove()
			s.tasks.Remove(task)
		}
	}
	return nil
}

func (s *TaskScheduler) isReady(currentTick int) bool {
	return !s.queue.IsEmpty() && s.queue.Current().GetNextRun() <= currentTick
}
