package scheduler

import (
	"fmt"
	stdmath "math"
	"reflect"
	"sort"
	"sync"
	"time"

	"pocketmine-go/pocketmine/log"
)

// AsyncWorker is a port of pocketmine\scheduler\AsyncWorker: one worker goroutine of an AsyncPool,
// running its tasks in submission order.
type AsyncWorker struct {
	id     int
	logger log.Logger
	queue  chan AsyncTask
	quit   chan struct{}

	storeMu sync.Mutex
	store   map[string]any
}

// GetAsyncWorkerId is a port of AsyncWorker::getAsyncWorkerId.
func (w *AsyncWorker) GetAsyncWorkerId() int { return w.id }

// GetThreadName is a port of AsyncWorker::getThreadName.
func (w *AsyncWorker) GetThreadName() string { return fmt.Sprintf("AsyncWorker#%d", w.id) }

func (w *AsyncWorker) GetLogger() log.Logger { return w.logger }

// SaveToThreadStore is a port of AsyncWorker::saveToThreadStore: data kept on this worker between
// tasks (e.g. a cache a series of tasks share).
func (w *AsyncWorker) SaveToThreadStore(identifier string, value any) {
	w.storeMu.Lock()
	defer w.storeMu.Unlock()
	w.store[identifier] = value
}

// GetFromThreadStore is a port of AsyncWorker::getFromThreadStore.
func (w *AsyncWorker) GetFromThreadStore(identifier string) any {
	w.storeMu.Lock()
	defer w.storeMu.Unlock()
	return w.store[identifier]
}

// RemoveFromThreadStore is a port of AsyncWorker::removeFromThreadStore.
func (w *AsyncWorker) RemoveFromThreadStore(identifier string) {
	w.storeMu.Lock()
	defer w.storeMu.Unlock()
	delete(w.store, identifier)
}

func (w *AsyncWorker) run() {
	for {
		select {
		case task := <-w.queue:
			runAsyncTask(task, w)
		case <-w.quit:
			return
		}
	}
}

// asyncPoolWorkerEntry is a port of pocketmine\scheduler\AsyncPoolWorkerEntry.
type asyncPoolWorkerEntry struct {
	worker   *AsyncWorker
	tasks    []AsyncTask
	lastUsed time.Time
}

// AsyncWorkerCrashError is PHP's ThreadCrashException for a task whose OnRun panicked.
type AsyncWorkerCrashError struct {
	Message string
	Panic   any
}

func (e *AsyncWorkerCrashError) Error() string { return fmt.Sprintf("%s: %v", e.Message, e.Panic) }

// AsyncPool is a port of pocketmine\scheduler\AsyncPool: manages general-purpose worker goroutines
// used for processing asynchronous tasks, and the tasks submitted to them. Tasks are collected
// (OnCompletion called) on the main thread by CollectTasks, which the server calls every tick.
type AsyncPool struct {
	size   int
	logger log.Logger

	workers          map[int]*asyncPoolWorkerEntry
	workerStartHooks []*func(worker int)
}

// NewAsyncPool is a port of AsyncPool::__construct. The worker memory limit and class loader have
// no Go counterpart.
func NewAsyncPool(size int, logger log.Logger) *AsyncPool {
	return &AsyncPool{size: size, logger: logger, workers: map[int]*asyncPoolWorkerEntry{}}
}

// GetSize is a port of AsyncPool::getSize: the maximum size of the pool.
func (p *AsyncPool) GetSize() int { return p.size }

// IncreaseSize is a port of AsyncPool::increaseSize: increases the maximum size of the pool to
// the specified amount. This does not immediately start new workers.
func (p *AsyncPool) IncreaseSize(newSize int) {
	if newSize > p.size {
		p.size = newSize
	}
}

// AddWorkerStartHook is a port of AsyncPool::addWorkerStartHook: registers a function called
// whenever a new worker is started, with the worker ID. It's also called for already-running
// workers.
func (p *AsyncPool) AddWorkerStartHook(hook *func(worker int)) {
	p.workerStartHooks = append(p.workerStartHooks, hook)
	for _, id := range p.GetRunningWorkers() {
		(*hook)(id)
	}
}

// RemoveWorkerStartHook is a port of AsyncPool::removeWorkerStartHook.
func (p *AsyncPool) RemoveWorkerStartHook(hook *func(worker int)) {
	for i, h := range p.workerStartHooks {
		if h == hook {
			p.workerStartHooks = append(p.workerStartHooks[:i], p.workerStartHooks[i+1:]...)
			return
		}
	}
}

// GetRunningWorkers is a port of AsyncPool::getRunningWorkers: the IDs of the running workers.
func (p *AsyncPool) GetRunningWorkers() []int {
	ids := make([]int, 0, len(p.workers))
	for id := range p.workers {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	return ids
}

// getWorker is a port of AsyncPool::getWorker: fetches the worker with the given ID, starting it
// if it doesn't exist.
func (p *AsyncPool) getWorker(workerID int) *asyncPoolWorkerEntry {
	entry, ok := p.workers[workerID]
	if !ok {
		w := &AsyncWorker{id: workerID, logger: p.logger, queue: make(chan AsyncTask, 1<<16), quit: make(chan struct{}), store: map[string]any{}}
		entry = &asyncPoolWorkerEntry{worker: w, lastUsed: time.Now()}
		p.workers[workerID] = entry
		go w.run()
		for _, hook := range p.workerStartHooks {
			(*hook)(workerID)
		}
	}
	return entry
}

// SubmitTaskToWorker is a port of AsyncPool::submitTaskToWorker: submits an AsyncTask to an
// arbitrary worker.
func (p *AsyncPool) SubmitTaskToWorker(task AsyncTask, worker int) {
	if worker < 0 || worker >= p.size {
		panic(fmt.Sprintf("Invalid worker %d", worker))
	}
	b := task.AsyncTaskState()
	if !b.submitted.CompareAndSwap(false, true) {
		panic("Cannot submit the same AsyncTask instance more than once")
	}
	entry := p.getWorker(worker)
	entry.tasks = append(entry.tasks, task)
	entry.lastUsed = time.Now()
	entry.worker.queue <- task
}

// SelectWorker is a port of AsyncPool::selectWorker: selects a worker ID to run a task. If
// there are any idle workers, the first found will be returned; otherwise, if there are free
// slots, a new worker will be started and its ID returned; otherwise the worker with the smallest
// backlog is chosen.
func (p *AsyncPool) SelectWorker() int {
	worker := -1
	minUsage := stdmath.MaxInt
	for _, i := range p.GetRunningWorkers() {
		if usage := len(p.workers[i].tasks); usage < minUsage {
			worker, minUsage = i, usage
			if usage == 0 {
				break
			}
		}
	}
	if worker == -1 || (minUsage > 0 && len(p.workers) < p.size) {
		//select a worker to start on the fly
		for i := 0; i < p.size; i++ {
			if _, ok := p.workers[i]; !ok {
				worker = i
				break
			}
		}
	}
	return worker
}

// SubmitTask is a port of AsyncPool::submitTask: submits an AsyncTask to the worker with the
// least load and returns its ID.
func (p *AsyncPool) SubmitTask(task AsyncTask) int {
	if task.AsyncTaskState().IsSubmitted() {
		panic("Cannot submit the same AsyncTask instance more than once")
	}
	worker := p.SelectWorker()
	p.SubmitTaskToWorker(task, worker)
	return worker
}

// CollectTasks is a port of AsyncPool::collectTasks: collects finished and/or crashed tasks from
// the workers, firing their on-completion hooks where appropriate. It reports whether there are
// still tasks running. A task whose OnRun panicked is returned as an *AsyncWorkerCrashError (PHP
// throws a ThreadCrashException, which crashes the server).
func (p *AsyncPool) CollectTasks() (bool, error) {
	for _, id := range p.GetRunningWorkers() {
		if _, err := p.CollectTasksFromWorker(id); err != nil {
			return false, err
		}
	}
	//we check this in a second loop, because task collection could have caused new tasks to be added to the queues
	for _, entry := range p.workers {
		if len(entry.tasks) > 0 {
			return true, nil
		}
	}
	return false, nil
}

// CollectTasksFromWorker is a port of AsyncPool::collectTasksFromWorker.
func (p *AsyncPool) CollectTasksFromWorker(worker int) (bool, error) {
	entry, ok := p.workers[worker]
	if !ok {
		return false, fmt.Errorf("no such worker %d", worker)
	}
	for len(entry.tasks) > 0 {
		task := entry.tasks[0]
		b := task.AsyncTaskState()
		if !b.IsFinished() {
			checkProgressUpdates(task)
			return true, nil //current task is still running, skip to next worker
		}
		entry.tasks = entry.tasks[1:]
		if b.IsCrashed() {
			return false, &AsyncWorkerCrashError{
				Message: fmt.Sprintf("Worker %d crashed while running task %s", worker, reflect.TypeOf(task)),
				Panic:   b.crash,
			}
		}
		/*
		 * It's possible for a task to submit a progress update and then finish before the progress
		 * update is detected by the parent thread, so here we consume any missed updates.
		 */
		checkProgressUpdates(task)
		task.OnCompletion()
	}
	return false, nil
}

// GetTaskQueueSizes is a port of AsyncPool::getTaskQueueSizes: worker ID => number of tasks.
func (p *AsyncPool) GetTaskQueueSizes() map[int]int {
	sizes := make(map[int]int, len(p.workers))
	for id, entry := range p.workers {
		sizes[id] = len(entry.tasks)
	}
	return sizes
}

// ShutdownUnusedWorkers is a port of AsyncPool::shutdownUnusedWorkers: stops workers idle for
// more than 5 minutes and returns how many were stopped.
func (p *AsyncPool) ShutdownUnusedWorkers() int {
	ret := 0
	for id, entry := range p.workers {
		if entry.lastUsed.Add(300*time.Second).Before(time.Now()) && len(entry.tasks) == 0 {
			close(entry.worker.quit)
			delete(p.workers, id)
			ret++
		}
	}
	return ret
}

// Shutdown is a port of AsyncPool::shutdown: cancels all pending tasks and shuts down all the
// workers in the pool, after collecting the ones still running.
func (p *AsyncPool) Shutdown() {
	for {
		more, err := p.CollectTasks()
		if err != nil {
			p.logger.Error(err.Error())
			continue
		}
		if !more {
			break
		}
		time.Sleep(time.Millisecond)
	}
	for _, entry := range p.workers {
		close(entry.worker.quit)
	}
	p.workers = map[int]*asyncPoolWorkerEntry{}
}
