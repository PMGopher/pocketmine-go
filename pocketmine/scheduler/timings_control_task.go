package scheduler

import (
	"fmt"

	"pocketmine-go/pocketmine/log"
)

// Timings hooks, set by the timings package's init() (timings imports this package, so the tasks
// below can't import it): TimingsHandler::setEnabled, ::reload and ::printCurrentThreadRecords.
var (
	TimingsSetEnabledFunc                func(enable bool)
	TimingsReloadFunc                    func()
	TimingsPrintCurrentThreadRecordsFunc func() []string
)

// Timings control operations, TimingsControlTask::ENABLE/DISABLE/RELOAD.
const (
	timingsControlEnable  = 1
	timingsControlDisable = 2
	timingsControlReload  = 3
)

// TimingsControlTask is a port of pocketmine\scheduler\TimingsControlTask: enables, disables or
// resets timings on a worker. Timings are process-wide in this port (see the timings package), so
// it acts on the same state as the main thread's TimingsHandler.
type TimingsControlTask struct {
	AsyncTaskBase
	operation int
}

// NewTimingsControlTaskSetEnabled is TimingsControlTask::setEnabled.
func NewTimingsControlTaskSetEnabled(enable bool) *TimingsControlTask {
	if enable {
		return &TimingsControlTask{operation: timingsControlEnable}
	}
	return &TimingsControlTask{operation: timingsControlDisable}
}

// NewTimingsControlTaskReload is TimingsControlTask::reload.
func NewTimingsControlTaskReload() *TimingsControlTask {
	return &TimingsControlTask{operation: timingsControlReload}
}

func (t *TimingsControlTask) OnRun() {
	switch t.operation {
	case timingsControlEnable:
		TimingsSetEnabledFunc(true)
		log.Global().Debug("Enabled timings")
	case timingsControlDisable:
		TimingsSetEnabledFunc(false)
		log.Global().Debug("Disabled timings")
	case timingsControlReload:
		TimingsReloadFunc()
		log.Global().Debug("Reset timings")
	default:
		panic(fmt.Sprintf("Invalid operation %d", t.operation))
	}
}

// TimingsCollectionTask is a port of pocketmine\scheduler\TimingsCollectionTask: collects a
// worker's timings records and hands them to onResult on the main thread (PHP resolves a
// PromiseResolver stored in the task's thread-local storage).
type TimingsCollectionTask struct {
	AsyncTaskBase
	onResult func(records []string)
}

func NewTimingsCollectionTask(onResult func(records []string)) *TimingsCollectionTask {
	return &TimingsCollectionTask{onResult: onResult}
}

func (t *TimingsCollectionTask) OnRun() {
	t.SetResult(TimingsPrintCurrentThreadRecordsFunc())
}

func (t *TimingsCollectionTask) OnCompletion() {
	records, _ := t.GetResult().([]string)
	t.onResult(records)
}
