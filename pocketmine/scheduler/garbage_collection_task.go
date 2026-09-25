package scheduler

import (
	"runtime"
	"runtime/debug"
)

// GarbageCollectionTask is a port of pocketmine\scheduler\GarbageCollectionTask: collects garbage
// from a worker. Go has one garbage collector for the whole process (gc_enable/gc_collect_cycles/
// gc_mem_caches become runtime.GC and debug.FreeOSMemory).
type GarbageCollectionTask struct {
	AsyncTaskBase
}

func NewGarbageCollectionTask() *GarbageCollectionTask { return &GarbageCollectionTask{} }

func (t *GarbageCollectionTask) OnRun() {
	runtime.GC()
	debug.FreeOSMemory()
}
