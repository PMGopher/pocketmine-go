package scheduler

import (
	"testing"
	"time"

	"pocketmine-go/pocketmine/log"
)

func TestTimingsTasksUseTheHooks(t *testing.T) {
	var enabled *bool
	reloaded := false
	TimingsSetEnabledFunc = func(v bool) { enabled = &v }
	TimingsReloadFunc = func() { reloaded = true }
	TimingsPrintCurrentThreadRecordsFunc = func() []string { return []string{"a", "b"} }

	pool := NewAsyncPool(1, log.NewSimpleLogger())
	defer pool.Shutdown()
	pool.SubmitTask(NewTimingsControlTaskSetEnabled(true))
	pool.SubmitTask(NewTimingsControlTaskReload())
	pool.SubmitTask(NewGarbageCollectionTask())
	var records []string
	pool.SubmitTask(NewTimingsCollectionTask(func(r []string) { records = r }))

	deadline := time.Now().Add(5 * time.Second)
	for records == nil && time.Now().Before(deadline) {
		if _, err := pool.CollectTasks(); err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Millisecond)
	}
	if enabled == nil || !*enabled || !reloaded {
		t.Error("the control tasks didn't call the timings hooks")
	}
	if len(records) != 2 {
		t.Errorf("collected records = %v", records)
	}
}
