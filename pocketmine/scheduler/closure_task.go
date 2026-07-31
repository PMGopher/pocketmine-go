package scheduler

import (
	"fmt"
	"reflect"
	"runtime"
)

// ClosureTask is a port of pocketmine\scheduler\ClosureTask: a Task implementation that just
// calls a plain function.
//
// PHP validates the closure's signature at construction time via reflection (Utils's
// CallbackType machinery); Go's compiler already enforces `func()`'s exact signature on whatever
// function is passed in, so there's nothing left to check at construction time here.
type ClosureTask struct {
	TaskBase
	fn func()
}

func NewClosureTask(fn func()) *ClosureTask {
	return &ClosureTask{fn: fn}
}

func (t *ClosureTask) OnRun() error {
	t.fn()
	return nil
}

// Name mirrors ClosureTask::getName() (Utils::getNiceClosureName()): runtime.FuncForPC gives a
// closure's synthetic name plus file/line, the closest Go equivalent of PHP's "closure@file#Lline".
func (t *ClosureTask) Name() string {
	if fn := runtime.FuncForPC(reflect.ValueOf(t.fn).Pointer()); fn != nil {
		file, line := fn.FileLine(fn.Entry())
		return fmt.Sprintf("closure@%s:%d (%s)", file, line, fn.Name())
	}
	return "closure@unknown"
}
