package io

import (
	"fmt"
	"sort"
	"strings"
)

// builtinProvider is one of the providers PHP's WorldProviderManager constructor registers.
type builtinProvider struct {
	name      string
	entry     WorldProviderManagerEntry
	isDefault bool
	order     int
}

var builtinProviders []builtinProvider

// RegisterBuiltinProvider registers a provider every new WorldProviderManager starts with. PHP's
// constructor references LevelDB, Anvil, McRegion and PMAnvil directly; those packages import this
// one, so they register themselves from init() instead. order keeps PHP's registration order
// (leveldb, anvil, mcregion, pmanvil).
func RegisterBuiltinProvider(name string, entry WorldProviderManagerEntry, isDefault bool, order int) {
	builtinProviders = append(builtinProviders, builtinProvider{name: name, entry: entry, isDefault: isDefault, order: order})
	sort.SliceStable(builtinProviders, func(i, j int) bool { return builtinProviders[i].order < builtinProviders[j].order })
}

// WorldProviderManager is a port of pocketmine\world\format\io\WorldProviderManager.
type WorldProviderManager struct {
	providers map[string]WorldProviderManagerEntry
	order     []string
	def       *WritableWorldProviderManagerEntry
}

// NewWorldProviderManager is a port of WorldProviderManager::__construct: LevelDB (the default)
// and the read-only Anvil, McRegion and PMAnvil providers.
func NewWorldProviderManager() *WorldProviderManager {
	m := &WorldProviderManager{providers: map[string]WorldProviderManagerEntry{}}
	for _, p := range builtinProviders {
		if p.isDefault {
			m.def = p.entry.(*WritableWorldProviderManagerEntry)
		}
		if err := m.AddProvider(p.entry, p.name, false); err != nil {
			panic(err)
		}
	}
	return m
}

// GetDefault is a port of WorldProviderManager::getDefault.
func (m *WorldProviderManager) GetDefault() *WritableWorldProviderManagerEntry { return m.def }

// SetDefault is a port of WorldProviderManager::setDefault.
func (m *WorldProviderManager) SetDefault(entry *WritableWorldProviderManagerEntry) { m.def = entry }

// AddProvider is a port of WorldProviderManager::addProvider.
func (m *WorldProviderManager) AddProvider(entry WorldProviderManagerEntry, name string, overwrite bool) error {
	name = strings.ToLower(name)
	if _, ok := m.providers[name]; ok {
		if !overwrite {
			return fmt.Errorf("Alias \"%s\" is already assigned", name)
		}
	} else {
		m.order = append(m.order, name)
	}
	m.providers[name] = entry
	return nil
}

// GetMatchingProviders is a port of WorldProviderManager::getMatchingProviders: the providers
// (by alias, in registration order) that can load the world at path.
func (m *WorldProviderManager) GetMatchingProviders(path string) []NamedWorldProviderManagerEntry {
	var result []NamedWorldProviderManagerEntry
	for _, alias := range m.order {
		if entry := m.providers[alias]; entry.IsValid(path) {
			result = append(result, NamedWorldProviderManagerEntry{Alias: alias, Entry: entry})
		}
	}
	return result
}

// NamedWorldProviderManagerEntry is one element of getMatchingProviders' alias => entry array.
type NamedWorldProviderManagerEntry struct {
	Alias string
	Entry WorldProviderManagerEntry
}

// GetAvailableProviders is a port of WorldProviderManager::getAvailableProviders.
func (m *WorldProviderManager) GetAvailableProviders() map[string]WorldProviderManagerEntry {
	return m.providers
}

// GetProviderByName is a port of WorldProviderManager::getProviderByName.
func (m *WorldProviderManager) GetProviderByName(name string) (WorldProviderManagerEntry, bool) {
	entry, ok := m.providers[strings.TrimSpace(strings.ToLower(name))]
	return entry, ok
}
