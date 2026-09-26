package io

import "pocketmine-go/pocketmine/log"

// WorldProviderManagerEntry is a port of pocketmine\world\format\io\WorldProviderManagerEntry.
type WorldProviderManagerEntry interface {
	IsValid(path string) bool
	FromPath(path string, logger log.Logger) (WorldProvider, error)
}

// ReadOnlyWorldProviderManagerEntry is a port of ReadOnlyWorldProviderManagerEntry.
type ReadOnlyWorldProviderManagerEntry struct {
	isValid  func(path string) bool
	fromPath func(path string, logger log.Logger) (WorldProvider, error)
}

// NewReadOnlyWorldProviderManagerEntry is a port of ReadOnlyWorldProviderManagerEntry::__construct.
func NewReadOnlyWorldProviderManagerEntry(isValid func(path string) bool, fromPath func(path string, logger log.Logger) (WorldProvider, error)) *ReadOnlyWorldProviderManagerEntry {
	return &ReadOnlyWorldProviderManagerEntry{isValid: isValid, fromPath: fromPath}
}

func (e *ReadOnlyWorldProviderManagerEntry) IsValid(path string) bool { return e.isValid(path) }

func (e *ReadOnlyWorldProviderManagerEntry) FromPath(path string, logger log.Logger) (WorldProvider, error) {
	return e.fromPath(path, logger)
}

// WritableWorldProviderManagerEntry is a port of WritableWorldProviderManagerEntry.
type WritableWorldProviderManagerEntry struct {
	isValid  func(path string) bool
	fromPath func(path string, logger log.Logger) (WritableWorldProvider, error)
	generate func(path, name string, options WorldCreationOptions) error
}

// NewWritableWorldProviderManagerEntry is a port of WritableWorldProviderManagerEntry::__construct.
func NewWritableWorldProviderManagerEntry(
	isValid func(path string) bool,
	fromPath func(path string, logger log.Logger) (WritableWorldProvider, error),
	generate func(path, name string, options WorldCreationOptions) error,
) *WritableWorldProviderManagerEntry {
	return &WritableWorldProviderManagerEntry{isValid: isValid, fromPath: fromPath, generate: generate}
}

func (e *WritableWorldProviderManagerEntry) IsValid(path string) bool { return e.isValid(path) }

func (e *WritableWorldProviderManagerEntry) FromPath(path string, logger log.Logger) (WorldProvider, error) {
	return e.fromPath(path, logger)
}

// FromPathWritable is WritableWorldProviderManagerEntry::fromPath's WritableWorldProvider result.
func (e *WritableWorldProviderManagerEntry) FromPathWritable(path string, logger log.Logger) (WritableWorldProvider, error) {
	return e.fromPath(path, logger)
}

// Generate is a port of WritableWorldProviderManagerEntry::generate: creates a new, empty world
// at path.
func (e *WritableWorldProviderManagerEntry) Generate(path, name string, options WorldCreationOptions) error {
	return e.generate(path, name, options)
}
