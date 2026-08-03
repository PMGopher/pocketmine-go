package generator

import "fmt"

// Factory constructs a Generator from a seed and the raw generatorOptions string level.dat stores
// alongside it - a port of the relevant slice of GeneratorManager's registration contract
// (GeneratorManagerEntry::getGeneratorClass()-plus-validateGeneratorOptions, folded into one call).
//
// Not every ported Generator can be constructed this way yet: Flat's real PHP constructor parses a
// preset string ("2;7,3,3,2;1") via FlatGeneratorOptions - this port's own NewFlat deliberately
// takes a pre-parsed []FlatLayer/[]Populator instead, since it has no such string parser (see
// Flat's own doc comment) - so "flat" is simply never registered here, the same documented gap.
type Factory func(seed int64, options string) (Generator, error)

// registry backs RegisterGenerator/GetFactory - a package-level map mirrors GeneratorManager's own
// real singleton nature (one shared registry for the whole running process), not something that
// needs a per-caller instance.
var registry = map[string]Factory{}

// RegisterGenerator is a port of GeneratorManager::addGenerator.
func RegisterGenerator(name string, factory Factory) { registry[name] = factory }

// GetFactory is a port of GeneratorManager::getGenerator.
func GetFactory(name string) (Factory, bool) {
	f, ok := registry[name]
	return f, ok
}

func init() {
	normalFactory := func(seed int64, options string) (Generator, error) { return NewNormal(int(seed)), nil }
	// "default" is real PocketMine-MP's own long-standing alias for "normal" (GeneratorManager
	// registers both names against the same Normal generator class).
	RegisterGenerator("normal", normalFactory)
	RegisterGenerator("default", normalFactory)
}

// UnknownGeneratorError is returned by anything resolving a generator by name (see world.WorldManager)
// when level.dat names one that was never registered.
type UnknownGeneratorError struct{ Name string }

func (e *UnknownGeneratorError) Error() string {
	return fmt.Sprintf("generator: unknown generator %q", e.Name)
}
