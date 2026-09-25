package scheduler

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// AsyncTask is a port of pocketmine\scheduler\AsyncTask: work that runs on an AsyncPool worker
// goroutine (OnRun) and then reports back on the main thread (OnCompletion, from
// AsyncPool.CollectTasks), so the result can safely touch the world and players.
//
// Concrete tasks embed AsyncTaskBase and implement OnRun; OnCompletion and OnProgressUpdate have
// no-op defaults. Like in PHP, OnRun must not touch main-thread state (worlds, players, ...).
type AsyncTask interface {
	// OnRun is the actions to execute when run, on a worker.
	OnRun()
	// OnCompletion is called from the main thread after OnRun is executed.
	OnCompletion()
	// OnProgressUpdate is called from the main thread with progress published by
	// PublishProgress on the worker.
	OnProgressUpdate(progress any)
	// AsyncTaskState returns the embedded AsyncTaskBase.
	AsyncTaskState() *AsyncTaskBase
}

// AsyncTaskBase holds the state PHP's AsyncTask keeps (result, progress updates, submitted and
// finished flags). Embed it in a concrete task.
type AsyncTaskBase struct {
	mu        sync.Mutex
	result    any
	hasResult bool
	progress  []any

	submitted atomic.Bool
	finished  atomic.Bool
	// crash is the panic value if OnRun panicked (PHP: the worker thread terminated).
	crash any

	local map[string]any

	// worker is AsyncTask::$worker: the worker running the task (set just before OnRun).
	worker *AsyncWorker
}

// GetWorker is AsyncTask's $this->worker: the worker the task is running on (nil before it runs).
func (b *AsyncTaskBase) GetWorker() *AsyncWorker { return b.worker }

func (b *AsyncTaskBase) AsyncTaskState() *AsyncTaskBase { return b }

// OnCompletion is AsyncTask::onCompletion's default (NOOP).
func (b *AsyncTaskBase) OnCompletion() {}

// OnProgressUpdate is AsyncTask::onProgressUpdate's default (NOOP).
func (b *AsyncTaskBase) OnProgressUpdate(progress any) {}

// IsCrashed is a port of AsyncTask::isCrashed.
func (b *AsyncTaskBase) IsCrashed() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.crash != nil
}

// IsFinished is a port of AsyncTask::isFinished: true once OnRun has returned (or crashed).
func (b *AsyncTaskBase) IsFinished() bool { return b.finished.Load() }

// HasResult is a port of AsyncTask::hasResult.
func (b *AsyncTaskBase) HasResult() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.hasResult && b.result != nil
}

// GetResult is a port of AsyncTask::getResult.
func (b *AsyncTaskBase) GetResult() any {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.result
}

// SetResult is a port of AsyncTask::setResult.
func (b *AsyncTaskBase) SetResult(result any) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.result, b.hasResult = result, true
}

// IsSubmitted is a port of AsyncTask::isSubmitted.
func (b *AsyncTaskBase) IsSubmitted() bool { return b.submitted.Load() }

// PublishProgress is a port of AsyncTask::publishProgress: call it from OnRun to have
// OnProgressUpdate called with progress on the main thread.
func (b *AsyncTaskBase) PublishProgress(progress any) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.progress = append(b.progress, progress)
}

// checkProgressUpdates is a port of AsyncTask::checkProgressUpdates.
func checkProgressUpdates(task AsyncTask) {
	b := task.AsyncTaskState()
	for {
		b.mu.Lock()
		if len(b.progress) == 0 {
			b.mu.Unlock()
			return
		}
		progress := b.progress[0]
		b.progress = b.progress[1:]
		b.mu.Unlock()
		task.OnProgressUpdate(progress)
	}
}

// StoreLocal is a port of AsyncTask::storeLocal: stores main-thread-only data (callbacks,
// objects) the task needs again in OnCompletion; it's never handed to the worker.
func (b *AsyncTaskBase) StoreLocal(key string, data any) {
	if b.local == nil {
		b.local = map[string]any{}
	}
	b.local[key] = data
}

// FetchLocal is a port of AsyncTask::fetchLocal; it panics if nothing was stored for key, like
// PHP's InvalidArgumentException.
func (b *AsyncTaskBase) FetchLocal(key string) any {
	v, ok := b.local[key]
	if !ok {
		panic(fmt.Sprintf("No matching thread-local data found on this thread (%s)", key))
	}
	return v
}

// run is a port of AsyncTask::run, on the worker.
func runAsyncTask(task AsyncTask, worker *AsyncWorker) {
	b := task.AsyncTaskState()
	b.worker = worker
	defer func() {
		if r := recover(); r != nil {
			b.mu.Lock()
			b.crash = r
			b.mu.Unlock()
		}
		b.finished.Store(true)
	}()
	task.OnRun()
}
