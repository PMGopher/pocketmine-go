package event

import "testing"

type testEvent struct {
	CancellableTrait
	Message string
}

func TestHandlersCalledInPriorityOrder(t *testing.T) {
	m := NewManager()
	var order []string

	RegisterListener[testEvent](m, "pluginA", Monitor, false, func(e *testEvent) { order = append(order, "monitor") })
	RegisterListener[testEvent](m, "pluginA", Lowest, false, func(e *testEvent) { order = append(order, "lowest") })
	RegisterListener[testEvent](m, "pluginA", Highest, false, func(e *testEvent) { order = append(order, "highest") })
	RegisterListener[testEvent](m, "pluginA", Normal, false, func(e *testEvent) { order = append(order, "normal") })

	CallOn(m, &testEvent{Message: "hi"})

	want := []string{"lowest", "normal", "highest", "monitor"}
	if len(order) != len(want) {
		t.Fatalf("order = %v, want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("order = %v, want %v", order, want)
		}
	}
}

func TestCancelledEventSkipsNonHandleCancelledListeners(t *testing.T) {
	m := NewManager()
	var calledNormal, calledForced bool

	RegisterListener[testEvent](m, "p", Lowest, false, func(e *testEvent) { e.Cancel() })
	RegisterListener[testEvent](m, "p", Normal, false, func(e *testEvent) { calledNormal = true })
	RegisterListener[testEvent](m, "p", High, true, func(e *testEvent) { calledForced = true })

	CallOn(m, &testEvent{})

	if calledNormal {
		t.Fatalf("a non-handleCancelled listener ran after the event was cancelled")
	}
	if !calledForced {
		t.Fatalf("a handleCancelled listener did not run for a cancelled event")
	}
}

func TestUnregisterRemovesOnlyThatListener(t *testing.T) {
	m := NewManager()
	var aCalled, bCalled bool

	handleA := RegisterListener[testEvent](m, "p", Normal, false, func(e *testEvent) { aCalled = true })
	RegisterListener[testEvent](m, "p", Normal, false, func(e *testEvent) { bCalled = true })

	handleA.Unregister()
	CallOn(m, &testEvent{})

	if aCalled {
		t.Fatalf("unregistered listener A still ran")
	}
	if !bCalled {
		t.Fatalf("listener B should still have run")
	}
}

func TestUnregisterAllForPlugin(t *testing.T) {
	m := NewManager()
	var pluginACalled, pluginBCalled bool

	RegisterListener[testEvent](m, "pluginA", Normal, false, func(e *testEvent) { pluginACalled = true })
	RegisterListener[testEvent](m, "pluginB", Normal, false, func(e *testEvent) { pluginBCalled = true })

	m.UnregisterAllForPlugin("pluginA")
	CallOn(m, &testEvent{})

	if pluginACalled {
		t.Fatalf("pluginA's listener still ran after UnregisterAllForPlugin")
	}
	if !pluginBCalled {
		t.Fatalf("pluginB's listener should still run")
	}
}

func TestHasHandlersOn(t *testing.T) {
	m := NewManager()
	if HasHandlersOn[testEvent](m) {
		t.Fatalf("expected no handlers before registration")
	}
	RegisterListener[testEvent](m, "p", Normal, false, func(e *testEvent) {})
	if !HasHandlersOn[testEvent](m) {
		t.Fatalf("expected handlers after registration")
	}
}

func TestRecursiveEventCallPanics(t *testing.T) {
	m := NewManager()
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected a panic from excessive recursive event calls")
		}
		eventCallDepth = 0 // reset shared state for other tests
	}()

	RegisterListener[testEvent](m, "p", Normal, false, func(e *testEvent) {
		CallOn(m, &testEvent{})
	})
	CallOn(m, &testEvent{})
}
