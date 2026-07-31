package promise

// All is a port of Promise::all(): resolves once every promise in `promises` has resolved, with
// a result map carrying each input's own key — matching PHP's version, which preserves the input
// array's (int|string) keys. Rejects as soon as any one input promise rejects.
func All[K comparable, T any](promises map[K]*Promise[T]) *Promise[map[K]T] {
	resolver := NewResolver[map[K]T]()
	if len(promises) == 0 {
		resolver.Resolve(map[K]T{})
		return resolver.GetPromise()
	}

	values := map[K]T{}
	toResolve := len(promises)
	cont := true

	for key, p := range promises {
		key := key
		p.OnCompletion(
			func(value T) {
				values[key] = value
				if len(values) == toResolve {
					resolver.Resolve(values)
				}
			},
			func() {
				if cont {
					cont = false
					resolver.Reject()
				}
			},
		)
		if !cont {
			break
		}
	}

	return resolver.GetPromise()
}
