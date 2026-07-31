package promise

// Resolver is a port of pocketmine\promise\PromiseResolver.
type Resolver[T any] struct {
	shared  *sharedData[T]
	promise *Promise[T]
}

func NewResolver[T any]() *Resolver[T] {
	s := &sharedData[T]{}
	return &Resolver[T]{shared: s, promise: &Promise[T]{shared: s}}
}

// Resolve settles the promise successfully. Panics if it was already resolved/rejected — PHP
// throws \LogicException here, and calling Resolve/Reject twice is a programmer error in the
// caller, not a recoverable runtime condition (matching this port's convention of panicking for
// that class of PHP exception rather than threading an error return through every caller).
func (r *Resolver[T]) Resolve(value T) {
	if r.shared.state != nil {
		panic("promise has already been resolved/rejected")
	}
	resolved := true
	r.shared.state = &resolved
	r.shared.result = value

	callbacks := r.shared.onSuccess
	r.shared.onSuccess = nil
	r.shared.onFailure = nil
	for _, cb := range callbacks {
		cb(value)
	}
}

func (r *Resolver[T]) Reject() {
	if r.shared.state != nil {
		panic("promise has already been resolved/rejected")
	}
	rejected := false
	r.shared.state = &rejected

	callbacks := r.shared.onFailure
	r.shared.onSuccess = nil
	r.shared.onFailure = nil
	for _, cb := range callbacks {
		cb()
	}
}

func (r *Resolver[T]) GetPromise() *Promise[T] { return r.promise }
