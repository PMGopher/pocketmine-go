package utils

import "sync"

// Singleton is a generic port of pocketmine\utils\SingletonTrait.
//
// PHP's trait is mixed into many classes and is not itself thread-safe, but PocketMine runs
// some singletons (e.g. the global logger) across worker threads, so this version adds a
// mutex around the lazily-created instance to make that safe under Go's goroutines too.
type Singleton[T any] struct {
	mu       sync.Mutex
	instance *T
	make     func() *T
}

// NewSingleton wraps a factory function, mirroring the trait's overridable make() method.
func NewSingleton[T any](make func() *T) *Singleton[T] {
	return &Singleton[T]{make: make}
}

func (s *Singleton[T]) GetInstance() *T {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.instance == nil {
		s.instance = s.make()
	}
	return s.instance
}

func (s *Singleton[T]) SetInstance(instance *T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.instance = instance
}

func (s *Singleton[T]) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.instance = nil
}
