package scheduler

// Task is a port of pocketmine\scheduler\Task.
//
// Task requires an unexported method (setHandler), which can only be satisfied by embedding
// TaskBase from this package — the Go equivalent of PHP's `abstract class Task`, which must be
// extended rather than just duck-typed against.
type Task interface {
	// OnRun runs the task. Returning a *CancelTaskError cancels it (the equivalent of PHP's
	// CancelTaskException, thrown from onRun() to cancel the task from within its own logic).
	// Any other non-nil error is treated as an uncaught failure (see TaskHandler.Run).
	OnRun() error
	OnCancel()
	Handler() *TaskHandler
	setHandler(*TaskHandler)
}

// CancelTaskError is a port of pocketmine\scheduler\CancelTaskException.
type CancelTaskError struct{}

func (e *CancelTaskError) Error() string { return "task was cancelled" }

// TaskBase provides Task's shared concrete behavior: the handler back-reference PHP's Task class
// manages, and a default no-op OnCancel(). Embed this in a concrete task type.
type TaskBase struct {
	handler *TaskHandler
}

func (t *TaskBase) Handler() *TaskHandler { return t.handler }

func (t *TaskBase) setHandler(h *TaskHandler) {
	if t.handler == nil || h == nil {
		t.handler = h
	}
}

func (t *TaskBase) OnCancel() {
	//NOOP
}

// namedTask is implemented by tasks that override the default name (see taskName) — the
// equivalent of a Task subclass overriding getName() (e.g. ClosureTask, which describes the
// closure instead of using the generic class name).
type namedTask interface {
	Name() string
}
