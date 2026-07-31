package promise

import "testing"

func TestResolveBeforeOnCompletion(t *testing.T) {
	r := NewResolver[int]()
	r.Resolve(42)

	var got int
	var failed bool
	r.GetPromise().OnCompletion(func(v int) { got = v }, func() { failed = true })

	if failed {
		t.Fatalf("onFailure called for a resolved promise")
	}
	if got != 42 {
		t.Fatalf("got = %d, want 42", got)
	}
}

func TestOnCompletionBeforeResolve(t *testing.T) {
	r := NewResolver[string]()
	var got string
	r.GetPromise().OnCompletion(func(v string) { got = v }, func() {})

	r.Resolve("hello")
	if got != "hello" {
		t.Fatalf("got = %q, want %q", got, "hello")
	}
}

func TestRejectCallsOnFailure(t *testing.T) {
	r := NewResolver[int]()
	var failed bool
	r.GetPromise().OnCompletion(func(v int) {}, func() { failed = true })

	r.Reject()
	if !failed {
		t.Fatalf("expected onFailure to be called after Reject()")
	}
}

func TestIsResolved(t *testing.T) {
	r := NewResolver[int]()
	if r.GetPromise().IsResolved() {
		t.Fatalf("expected IsResolved() = false before resolution")
	}
	r.Resolve(1)
	if !r.GetPromise().IsResolved() {
		t.Fatalf("expected IsResolved() = true after Resolve()")
	}
}

func TestDoubleResolvePanics(t *testing.T) {
	r := NewResolver[int]()
	r.Resolve(1)

	defer func() {
		if recover() == nil {
			t.Fatalf("expected a panic resolving an already-settled promise")
		}
	}()
	r.Resolve(2)
}

func TestAllResolvesWhenEveryPromiseResolves(t *testing.T) {
	a := NewResolver[int]()
	b := NewResolver[int]()

	combined := All(map[string]*Promise[int]{"a": a.GetPromise(), "b": b.GetPromise()})

	var result map[string]int
	combined.OnCompletion(func(v map[string]int) { result = v }, func() { t.Fatalf("unexpected rejection") })

	a.Resolve(1)
	if result != nil {
		t.Fatalf("All() resolved before every input settled")
	}
	b.Resolve(2)

	if result["a"] != 1 || result["b"] != 2 {
		t.Fatalf("result = %v, want a=1 b=2", result)
	}
}

func TestAllRejectsIfAnyPromiseRejects(t *testing.T) {
	a := NewResolver[int]()
	b := NewResolver[int]()

	combined := All(map[string]*Promise[int]{"a": a.GetPromise(), "b": b.GetPromise()})

	var rejected bool
	combined.OnCompletion(func(v map[string]int) { t.Fatalf("unexpected resolution") }, func() { rejected = true })

	a.Reject()
	if !rejected {
		t.Fatalf("expected All() to reject when one input rejects")
	}
	b.Resolve(2) // should be a no-op at this point
}

func TestAllWithNoPromisesResolvesImmediately(t *testing.T) {
	combined := All(map[string]*Promise[int]{})
	if !combined.IsResolved() {
		t.Fatalf("expected All({}) to resolve immediately")
	}
}
