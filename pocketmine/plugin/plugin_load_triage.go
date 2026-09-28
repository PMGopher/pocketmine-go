package plugin

// PluginLoadTriageEntry is a port of pocketmine\plugin\PluginLoadTriageEntry.
type PluginLoadTriageEntry struct {
	file        string
	loader      PluginLoader
	description *Description
}

func (e *PluginLoadTriageEntry) GetFile() string              { return e.file }
func (e *PluginLoadTriageEntry) GetLoader() PluginLoader      { return e.loader }
func (e *PluginLoadTriageEntry) GetDescription() *Description { return e.description }

// PluginLoadTriage is a port of pocketmine\plugin\PluginLoadTriage. order keeps the insertion
// order of plugins (PHP arrays are ordered).
type PluginLoadTriage struct {
	plugins          map[string]*PluginLoadTriageEntry
	order            []string
	dependencies     map[string][]string
	softDependencies map[string][]string
}

func newPluginLoadTriage() *PluginLoadTriage {
	return &PluginLoadTriage{plugins: map[string]*PluginLoadTriageEntry{}, dependencies: map[string][]string{}, softDependencies: map[string][]string{}}
}

func (t *PluginLoadTriage) add(name string, entry *PluginLoadTriageEntry) {
	if _, ok := t.plugins[name]; !ok {
		t.order = append(t.order, name)
	}
	t.plugins[name] = entry
}

func (t *PluginLoadTriage) remove(name string) {
	delete(t.plugins, name)
	for i, n := range t.order {
		if n == name {
			t.order = append(t.order[:i], t.order[i+1:]...)
			break
		}
	}
}

// names is a snapshot of the plugin names in order (PHP's foreach over a copy of the array).
func (t *PluginLoadTriage) names() []string { return append([]string(nil), t.order...) }
