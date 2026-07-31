package utils

import "iter"

// ObjectSet is a port of pocketmine\utils\ObjectSet: an insertion-ordered set of objects.
//
// PHP arrays preserve insertion order even when keyed by spl_object_id(), and PocketMine
// relies on that (e.g. event listener call order), so this keeps an explicit slice alongside
// the lookup map rather than using a plain Go map, which has no defined iteration order.
// T is expected to be a pointer or other identity-comparable type, mirroring PHP's `object`.
type ObjectSet[T comparable] struct {
	order []T
	seen  map[T]bool
}

func NewObjectSet[T comparable]() *ObjectSet[T] {
	return &ObjectSet[T]{seen: make(map[T]bool)}
}

func (s *ObjectSet[T]) Add(objects ...T) {
	for _, o := range objects {
		if !s.seen[o] {
			s.seen[o] = true
			s.order = append(s.order, o)
		}
	}
}

func (s *ObjectSet[T]) Remove(objects ...T) {
	for _, o := range objects {
		if s.seen[o] {
			delete(s.seen, o)
			for i, existing := range s.order {
				if existing == o {
					s.order = append(s.order[:i], s.order[i+1:]...)
					break
				}
			}
		}
	}
}

func (s *ObjectSet[T]) Clear() {
	s.order = nil
	s.seen = make(map[T]bool)
}

func (s *ObjectSet[T]) Contains(o T) bool {
	return s.seen[o]
}

// ToSlice returns a snapshot of the set's contents, in insertion order.
func (s *ObjectSet[T]) ToSlice() []T {
	result := make([]T, len(s.order))
	copy(result, s.order)
	return result
}

// All returns a range-over-func iterator, the idiomatic Go replacement for PHP's IteratorAggregate.
func (s *ObjectSet[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, o := range s.order {
			if !yield(o) {
				return
			}
		}
	}
}
