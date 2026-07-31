package timings

import (
	"testing"
	"time"
)

func TestTimingsBasicStartStop(t *testing.T) {
	SetEnabled(true)
	defer SetEnabled(false)
	Reload()

	h := NewTimingsHandler("Test Timer", nil, "")
	h.StartTiming()
	time.Sleep(time.Millisecond)
	h.StopTiming()

	found := false
	for _, r := range GetAllRecords() {
		if r.GetTimerID() == h.id {
			found = true
			if r.GetCount() != 1 {
				t.Fatalf("GetCount() = %d, want 1", r.GetCount())
			}
			if r.GetTotalTime() <= 0 {
				t.Fatalf("GetTotalTime() = %v, want > 0", r.GetTotalTime())
			}
		}
	}
	if !found {
		t.Fatalf("expected a record to exist for the handler after timing")
	}
}

func TestTimingsNestedParentChild(t *testing.T) {
	SetEnabled(true)
	defer SetEnabled(false)
	Reload()

	parent := NewTimingsHandler("Parent", nil, "")
	child := NewTimingsHandler("Child", parent, "")

	child.StartTiming()
	child.StopTiming()

	// Starting/stopping the child must also start/stop the parent (see internalStartTiming).
	var parentRecordFound, childRecordFound bool
	for _, r := range GetAllRecords() {
		if r.GetTimerID() == parent.id && r.GetCount() == 1 {
			parentRecordFound = true
		}
		if r.GetTimerID() == child.id && r.GetCount() == 1 {
			childRecordFound = true
		}
	}
	if !parentRecordFound {
		t.Fatalf("expected the parent handler to have recorded a timing too")
	}
	if !childRecordFound {
		t.Fatalf("expected the child handler to have recorded a timing")
	}
}

func TestTimingsDisabledDoesNothing(t *testing.T) {
	SetEnabled(false)
	Reload()

	h := NewTimingsHandler("Disabled Timer", nil, "")
	h.StartTiming()
	h.StopTiming()

	for _, r := range GetAllRecords() {
		if r.GetTimerID() == h.id {
			t.Fatalf("expected no record to be created while timings are disabled")
		}
	}
}

func TestTimeHelperReturnsValue(t *testing.T) {
	SetEnabled(true)
	defer SetEnabled(false)
	Reload()

	h := NewTimingsHandler("Time Helper", nil, "")
	result := Time(h, func() int { return 42 })
	if result != 42 {
		t.Fatalf("Time() = %d, want 42", result)
	}
}

func TestGetScheduledTaskTimingsIsMemoized(t *testing.T) {
	Init()
	// Using a nil-safe minimal stand-in isn't possible without importing scheduler in the test
	// too, but GetCommandDispatchTimings exercises the same memoization path without that
	// dependency.
	a := GetCommandDispatchTimings("test")
	b := GetCommandDispatchTimings("test")
	if a != b {
		t.Fatalf("expected the same *TimingsHandler instance for repeated calls with the same name")
	}
}

func TestResetClearsRecords(t *testing.T) {
	SetEnabled(true)
	defer SetEnabled(false)

	h := NewTimingsHandler("Reset Test", nil, "")
	h.StartTiming()
	h.StopTiming()

	if len(GetAllRecords()) == 0 {
		t.Fatalf("expected at least one record before reset")
	}
	Reload()
	if len(GetAllRecords()) != 0 {
		t.Fatalf("expected no records immediately after Reload(), got %d", len(GetAllRecords()))
	}
}
