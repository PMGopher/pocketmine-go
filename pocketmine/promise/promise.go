package promise

// Package promise is a port of pocketmine\promise: a single-resolution future, used to compose
// eventual results (e.g. an async report that different subsystems each contribute a piece of).
//
// Like event.CallOn's call-depth counter and the timings package's current-record pointer, this
// assumes callbacks are registered and resolved sequentially on one execution context — resolve/
// reject fire their registered callbacks synchronously and immediately, matching PHP's
// single-threaded model. It is not a concurrency primitive; if a value is genuinely produced on
// another goroutine, that goroutine must hand off to the main one before calling Resolve/Reject.
//
// PHP's version fakes generics via phpstan-only template annotations (real PHP has no generics);
// this port uses actual Go generics, so Promise[string], Promise[int], etc. are properly
// distinct, statically-checked types.
type sharedData[T any] struct {
	state     *bool // nil = pending, true = resolved, false = rejected
	result    T
	onSuccess []func(T)
	onFailure []func()
}

// Promise is a port of pocketmine\promise\Promise. Construct one via NewResolver, not directly
// (mirroring the PHP original's "do NOT call this directly" constructor note).
type Promise[T any] struct {
	shared *sharedData[T]
}

// OnCompletion registers onSuccess/onFailure, calling whichever already applies immediately if
// the promise has already settled.
func (p *Promise[T]) OnCompletion(onSuccess func(T), onFailure func()) {
	s := p.shared
	if s.state != nil {
		if *s.state {
			onSuccess(s.result)
		} else {
			onFailure()
		}
		return
	}
	s.onSuccess = append(s.onSuccess, onSuccess)
	s.onFailure = append(s.onFailure, onFailure)
}

// IsResolved mirrors Promise::isResolved() exactly, including its own TODO: this returns false
// for both "still pending" and "rejected" — there's no way to distinguish them from here.
func (p *Promise[T]) IsResolved() bool {
	return p.shared.state != nil && *p.shared.state
}
