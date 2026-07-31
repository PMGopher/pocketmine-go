package utils

import (
	"fmt"
	"strings"
)

// StringToTParser is a generic port of pocketmine\utils\StringToTParser: parses a
// user-friendly alias string (e.g. from the /give command) into a value of type T.
type StringToTParser[T any] struct {
	callbackMap map[string]func(input string) T
}

func NewStringToTParser[T any]() *StringToTParser[T] {
	return &StringToTParser[T]{callbackMap: map[string]func(input string) T{}}
}

func (p *StringToTParser[T]) Register(alias string, callback func(input string) T) error {
	key := p.reprocess(alias)
	if _, ok := p.callbackMap[key]; ok {
		return fmt.Errorf("alias %q is already registered", key)
	}
	p.callbackMap[key] = callback
	return nil
}

func (p *StringToTParser[T]) Override(alias string, callback func(input string) T) {
	p.callbackMap[p.reprocess(alias)] = callback
}

// RegisterAlias registers a new alias for an existing known alias.
func (p *StringToTParser[T]) RegisterAlias(existing, alias string) error {
	existingKey := p.reprocess(existing)
	callback, ok := p.callbackMap[existingKey]
	if !ok {
		return fmt.Errorf("cannot register new alias for unknown existing alias %q", existing)
	}
	newKey := p.reprocess(alias)
	if _, ok := p.callbackMap[newKey]; ok {
		return fmt.Errorf("alias %q is already registered", newKey)
	}
	p.callbackMap[newKey] = callback
	return nil
}

// Parse tries to parse the given string into a T. ok is false if no alias matched.
func (p *StringToTParser[T]) Parse(input string) (result T, ok bool) {
	callback, found := p.callbackMap[p.reprocess(input)]
	if !found {
		return result, false
	}
	return callback(input), true
}

func (p *StringToTParser[T]) reprocess(input string) string {
	return strings.ToLower(strings.NewReplacer(" ", "_", "minecraft:", "").Replace(strings.TrimSpace(input)))
}

func (p *StringToTParser[T]) GetKnownAliases() []string {
	result := make([]string, 0, len(p.callbackMap))
	for k := range p.callbackMap {
		result = append(result, k)
	}
	return result
}
