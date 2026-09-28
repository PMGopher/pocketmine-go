package plugin

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"pocketmine-go/pocketmine/event"
	pluginevent "pocketmine-go/pocketmine/event/plugin"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/permission"
)

// PluginManager is a port of pocketmine\plugin\PluginManager.
//
// Maps that PHP keeps as ordered arrays have an order slice next to them.
type PluginManager struct {
	server              Server
	pluginDataDirectory string // "" for PHP's null (legacy data folders next to the plugins)
	graylist            *PluginGraylist

	plugins     map[string]Plugin
	pluginOrder []string

	enabledPlugins map[string]Plugin
	enabledOrder   []string

	pluginDependents map[string]map[string]bool

	loadPluginsGuard bool

	fileAssociations map[string]PluginLoader
	loaderOrder      []string
}

// NewPluginManager is a port of PluginManager::__construct. pluginDataDirectory "" is PHP's null.
func NewPluginManager(server Server, pluginDataDirectory string, graylist *PluginGraylist) (*PluginManager, error) {
	if pluginDataDirectory != "" {
		if info, err := os.Stat(pluginDataDirectory); errors.Is(err, os.ErrNotExist) {
			_ = os.MkdirAll(pluginDataDirectory, 0o777)
		} else if err == nil && !info.IsDir() {
			return nil, fmt.Errorf("Plugin data path %s exists and is not a directory", pluginDataDirectory)
		}
	}
	return &PluginManager{
		server:              server,
		pluginDataDirectory: pluginDataDirectory,
		graylist:            graylist,
		plugins:             map[string]Plugin{},
		enabledPlugins:      map[string]Plugin{},
		pluginDependents:    map[string]map[string]bool{},
		fileAssociations:    map[string]PluginLoader{},
	}, nil
}

// GetPlugin is a port of PluginManager::getPlugin (nil if there's no such plugin).
func (m *PluginManager) GetPlugin(name string) Plugin { return m.plugins[name] }

// RegisterInterface is a port of PluginManager::registerInterface.
func (m *PluginManager) RegisterInterface(loader PluginLoader) {
	key := reflect.TypeOf(loader).String()
	if _, ok := m.fileAssociations[key]; !ok {
		m.loaderOrder = append(m.loaderOrder, key)
	}
	m.fileAssociations[key] = loader
}

// GetPlugins is a port of PluginManager::getPlugins, in load order.
func (m *PluginManager) GetPlugins() []Plugin {
	result := make([]Plugin, 0, len(m.pluginOrder))
	for _, name := range m.pluginOrder {
		result = append(result, m.plugins[name])
	}
	return result
}

func (m *PluginManager) getDataDirectory(pluginPath, pluginName string) string {
	if m.pluginDataDirectory != "" {
		return filepath.Join(m.pluginDataDirectory, pluginName)
	}
	return filepath.Join(filepath.Dir(pluginPath), pluginName)
}

func (m *PluginManager) translate(t *lang.Translatable) string {
	return m.server.GetLanguage().Translate(t)
}

// internalLoadPlugin is a port of PluginManager::internalLoadPlugin.
func (m *PluginManager) internalLoadPlugin(path string, loader PluginLoader, description *Description) Plugin {
	logger := m.server.GetLogger()
	logger.Info(m.translate(lang.KnownTranslationFactory.PocketminePluginLoad(description.GetFullName())))

	dataFolder := m.getDataDirectory(path, description.GetName())
	if info, err := os.Stat(dataFolder); err == nil && !info.IsDir() {
		logger.Critical(m.translate(lang.KnownTranslationFactory.PocketminePluginLoadError(
			description.GetName(),
			lang.KnownTranslationFactory.PocketminePluginBadDataFolder(dataFolder),
		)))
		return nil
	}
	_ = os.MkdirAll(dataFolder, 0o777)

	prefixed := loader.GetAccessProtocol() + path
	loader.LoadPlugin(prefixed)

	factory, _ := loader.(pluginFactory)
	var plugin Plugin
	if factory != nil {
		plugin, _ = factory.NewPlugin(prefixed, description)
	}
	if plugin == nil {
		logger.Critical(m.translate(lang.KnownTranslationFactory.PocketminePluginLoadError(
			description.GetName(),
			lang.KnownTranslationFactory.PocketminePluginMainClassNotFound(),
		)))
		return nil
	}
	initializer, ok := plugin.(Initializer)
	if !ok {
		logger.Critical(m.translate(lang.KnownTranslationFactory.PocketminePluginLoadError(
			description.GetName(),
			lang.KnownTranslationFactory.PocketminePluginMainClassWrongType("pocketmine-go/pocketmine/plugin.Initializer"),
		)))
		return nil
	}

	permManager := permission.GetManager()
	permissions := description.GetPermissions()
	for _, permsGroup := range permissions {
		for _, perm := range permsGroup {
			if permManager.GetPermission(perm.Name()) != nil {
				logger.Critical(m.translate(lang.KnownTranslationFactory.PocketminePluginLoadError(
					description.GetName(),
					lang.KnownTranslationFactory.PocketminePluginDuplicatePermissionError(perm.Name()),
				)))
				return nil
			}
		}
	}
	opRoot := permManager.GetPermission(permission.RootOperator)
	everyoneRoot := permManager.GetPermission(permission.RootUser)
	for _, def := range sortedKeys(permissions) {
		for _, perm := range permissions[def] {
			permManager.AddPermission(perm)
			switch def {
			case permission.DefaultTrue:
				everyoneRoot.AddChild(perm.Name(), true)
			case permission.DefaultOp:
				opRoot.AddChild(perm.Name(), true)
			case permission.DefaultNotOp:
				//TODO: I don't think anyone uses this, and it currently relies on some magic inside PermissibleBase
				//to ensure that the operator override actually applies.
				//Explore getting rid of this.
				//The following grants this permission to anyone who has the "everyone" root permission.
				//However, if the operator root node (which has higher priority) is present, the
				//permission will be denied instead.
				everyoneRoot.AddChild(perm.Name(), true)
				opRoot.AddChild(perm.Name(), false)
			}
		}
	}

	initializer.InitPlugin(plugin, loader, m.server, description, dataFolder, prefixed, factory.GetResourceProvider(prefixed))
	name := plugin.GetDescription().GetName()
	if _, ok := m.plugins[name]; !ok {
		m.pluginOrder = append(m.pluginOrder, name)
	}
	m.plugins[name] = plugin

	return plugin
}

// triagePlugins is a port of PluginManager::triagePlugins. Besides the files in path, it looks at
// the virtual files of loaders that implement pluginSource (compiled-in Go plugins).
func (m *PluginManager) triagePlugins(path string, triage *PluginLoadTriage, loadErrorCount *int, newLoaders []string) {
	var loaders []PluginLoader
	if newLoaders != nil {
		for _, key := range newLoaders {
			if l, ok := m.fileAssociations[key]; ok {
				loaders = append(loaders, l)
			}
		}
	} else {
		for _, key := range m.loaderOrder {
			loaders = append(loaders, m.fileAssociations[key])
		}
	}

	var files []string
	dir := path
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		entries, _ := os.ReadDir(path)
		for _, e := range entries {
			files = append(files, filepath.Join(path, e.Name()))
		}
	} else if err == nil {
		realPath, err := filepath.Abs(path)
		if err != nil {
			realPath = path
		}
		files = []string{realPath}
		dir = filepath.Dir(realPath)
	} else {
		return
	}
	for _, l := range loaders {
		if source, ok := l.(pluginSource); ok {
			files = append(files, source.GetPluginFiles(dir)...)
		}
	}
	rand.Shuffle(len(files), func(i, j int) { files[i], files[j] = files[j], files[i] }) //this prevents plugins implicitly relying on the filesystem name order when they should be using dependency properties

	logger := m.server.GetLogger()
	loadabilityChecker := NewPluginLoadabilityChecker(m.server.GetApiVersion())
	for _, loader := range loaders {
		for _, file := range files {
			if !loader.CanLoadPlugin(file) {
				continue
			}
			description, err := loader.GetPluginDescription(file)
			if err != nil {
				var parseErr *PluginDescriptionParseException
				if errors.As(err, &parseErr) {
					logger.Critical(m.translate(lang.KnownTranslationFactory.PocketminePluginLoadError(
						file,
						lang.KnownTranslationFactory.PocketminePluginInvalidManifest(parseErr.Message),
					)))
				} else {
					logger.Critical(m.translate(lang.KnownTranslationFactory.PocketminePluginLoadError(file, err.Error())))
					logger.LogException(err, "")
				}
				*loadErrorCount++
				continue
			}
			if description == nil {
				continue
			}

			name := description.GetName()

			if m.graylist != nil && !m.graylist.IsAllowed(name) {
				reason := lang.KnownTranslationFactory.PocketminePluginDisallowedByBlacklist()
				if m.graylist.IsWhitelist() {
					reason = lang.KnownTranslationFactory.PocketminePluginDisallowedByWhitelist()
				}
				logger.Notice(m.translate(lang.KnownTranslationFactory.PocketminePluginLoadError(name, reason)))
				//this does NOT increment loadErrorCount, because using the graylist to prevent a plugin from
				//loading is not considered accidental; this is the same as if the plugin were manually removed
				//this means that the server will continue to boot even if some plugins were blocked by graylist
				continue
			}

			if loadabilityError := loadabilityChecker.Check(description); loadabilityError != nil {
				logger.Critical(m.translate(lang.KnownTranslationFactory.PocketminePluginLoadError(name, loadabilityError)))
				*loadErrorCount++
				continue
			}

			if _, exists := triage.plugins[name]; exists || m.GetPlugin(name) != nil {
				logger.Critical(m.translate(lang.KnownTranslationFactory.PocketminePluginDuplicateError(name)))
				*loadErrorCount++
				continue
			}

			if strings.Contains(name, " ") {
				logger.Warning(m.translate(lang.KnownTranslationFactory.PocketminePluginSpacesDiscouraged(name)))
			}

			triage.add(name, &PluginLoadTriageEntry{file: file, loader: loader, description: description})

			triage.softDependencies[name] = append(triage.softDependencies[name], description.GetSoftDepend()...)
			triage.dependencies[name] = append([]string(nil), description.GetDepend()...)

			for _, before := range description.GetLoadBefore() {
				triage.softDependencies[before] = append(triage.softDependencies[before], name)
			}
		}
	}
}

// checkDepsForTriage is a port of PluginManager::checkDepsForTriage.
func (m *PluginManager) checkDepsForTriage(pluginName, dependencyType string, dependencyLists map[string][]string, loadedPlugins map[string]Plugin, triage *PluginLoadTriage) {
	list, ok := dependencyLists[pluginName]
	if !ok {
		return
	}
	var remaining []string
	for _, dependency := range list {
		if _, loaded := loadedPlugins[dependency]; loaded || m.GetPlugin(dependency) != nil {
			m.server.GetLogger().Debug(fmt.Sprintf("Successfully resolved %s dependency \"%s\" for plugin \"%s\"", dependencyType, dependency, pluginName))
			continue
		}
		if _, found := triage.plugins[dependency]; found {
			m.server.GetLogger().Debug(fmt.Sprintf("Deferring resolution of %s dependency \"%s\" for plugin \"%s\" (found but not loaded yet)", dependencyType, dependency, pluginName))
		}
		remaining = append(remaining, dependency)
	}

	if len(remaining) == 0 {
		delete(dependencyLists, pluginName)
	} else {
		dependencyLists[pluginName] = remaining
	}
}

// LoadPlugins is a port of PluginManager::loadPlugins: loads the plugins in path (a folder, or a
// single plugin file), resolving load order from dependencies. loadErrorCount is incremented for
// each plugin that failed to load.
func (m *PluginManager) LoadPlugins(path string, loadErrorCount *int) map[string]Plugin {
	if m.loadPluginsGuard {
		panic("PluginManager.LoadPlugins() cannot be called from within itself")
	}
	m.loadPluginsGuard = true
	defer func() { m.loadPluginsGuard = false }()
	if loadErrorCount == nil {
		loadErrorCount = new(int)
	}

	triage := newPluginLoadTriage()
	m.triagePlugins(path, triage, loadErrorCount, nil)

	loadedPlugins := map[string]Plugin{}

outer:
	for len(triage.plugins) > 0 {
		loadedThisLoop := 0
		for _, name := range triage.names() {
			entry, ok := triage.plugins[name]
			if !ok {
				continue
			}
			m.checkDepsForTriage(name, "hard", triage.dependencies, loadedPlugins, triage)
			m.checkDepsForTriage(name, "soft", triage.softDependencies, loadedPlugins, triage)

			_, hasDeps := triage.dependencies[name]
			_, hasSoftDeps := triage.softDependencies[name]
			if hasDeps || hasSoftDeps {
				continue
			}
			triage.remove(name)
			loadedThisLoop++

			oldRegisteredLoaders := map[string]bool{}
			for k := range m.fileAssociations {
				oldRegisteredLoaders[k] = true
			}
			if plugin := m.internalLoadPlugin(entry.GetFile(), entry.GetLoader(), entry.GetDescription()); plugin != nil {
				loadedPlugins[name] = plugin
				var diffLoaders []string
				for _, k := range m.loaderOrder {
					if !oldRegisteredLoaders[k] {
						diffLoaders = append(diffLoaders, k)
					}
				}
				if len(diffLoaders) != 0 {
					m.server.GetLogger().Debug("Plugin " + name + " registered a new plugin loader during load, scanning for new plugins")
					before := map[string]bool{}
					for n := range triage.plugins {
						before[n] = true
					}
					m.triagePlugins(path, triage, loadErrorCount, diffLoaders)
					var diffPlugins []string
					for _, n := range triage.order {
						if !before[n] {
							diffPlugins = append(diffPlugins, n)
						}
					}
					m.server.GetLogger().Debug("Re-triage found plugins: " + strings.Join(diffPlugins, ", "))
				}
			} else {
				*loadErrorCount++
			}
		}

		if loadedThisLoop == 0 {
			//No plugins loaded :(

			//check for skippable soft dependencies first, in case the dependents could resolve hard dependencies
			for _, name := range triage.names() {
				softDeps, hasSoft := triage.softDependencies[name]
				if _, hasHard := triage.dependencies[name]; hasSoft && !hasHard {
					var remaining []string
					for _, dependency := range softDeps {
						if _, inTriage := triage.plugins[dependency]; m.GetPlugin(dependency) == nil && !inTriage {
							m.server.GetLogger().Debug(fmt.Sprintf("Skipping resolution of missing soft dependency \"%s\" for plugin \"%s\"", dependency, name))
							continue
						}
						remaining = append(remaining, dependency)
					}
					if len(remaining) == 0 {
						delete(triage.softDependencies, name)
						continue outer //go back to the top and try again
					}
					triage.softDependencies[name] = remaining
				}
			}

			for _, name := range triage.names() {
				deps, ok := triage.dependencies[name]
				if !ok {
					continue
				}
				var unknownDependencies []string
				for _, dependency := range deps {
					if _, inTriage := triage.plugins[dependency]; m.GetPlugin(dependency) == nil && !inTriage {
						//assume that the plugin is never going to be loaded
						//by this point all soft dependencies have been ignored if they were able to be, so
						//there's no chance of this dependency ever being resolved
						unknownDependencies = append(unknownDependencies, dependency)
					}
				}

				if len(unknownDependencies) > 0 {
					m.server.GetLogger().Critical(m.translate(lang.KnownTranslationFactory.PocketminePluginLoadError(
						name,
						lang.KnownTranslationFactory.PocketminePluginUnknownDependency(strings.Join(unknownDependencies, ", ")),
					)))
					triage.remove(name)
					*loadErrorCount++
				}
			}

			for _, name := range triage.names() {
				m.server.GetLogger().Critical(m.translate(lang.KnownTranslationFactory.PocketminePluginLoadError(name, lang.KnownTranslationFactory.PocketminePluginCircularDependency())))
				*loadErrorCount++
			}
			break
		}
	}

	return loadedPlugins
}

// IsPluginEnabled is a port of PluginManager::isPluginEnabled.
func (m *PluginManager) IsPluginEnabled(plugin Plugin) bool {
	_, ok := m.plugins[plugin.GetDescription().GetName()]
	return ok && plugin.IsEnabled()
}

// EnablePlugin is a port of PluginManager::enablePlugin.
func (m *PluginManager) EnablePlugin(plugin Plugin) bool {
	if plugin.IsEnabled() {
		return true //TODO: maybe this should be an error?
	}
	m.server.GetLogger().Info(m.translate(lang.KnownTranslationFactory.PocketminePluginEnable(plugin.GetDescription().GetFullName())))

	plugin.GetScheduler().SetEnabled(true)
	if err := plugin.OnEnableStateChange(true); err != nil {
		var disable *DisablePluginException
		if !errors.As(err, &disable) {
			// PHP lets any other exception from onEnable() crash the server; here the plugin is
			// disabled and the error logged.
			plugin.GetLogger().LogException(err, "")
		}
		m.DisablePlugin(plugin)
	}

	if !plugin.IsEnabled() { //the plugin may have disabled itself during onEnable()
		m.server.GetLogger().Critical(m.translate(lang.KnownTranslationFactory.PocketminePluginEnableError(
			plugin.GetName(),
			lang.KnownTranslationFactory.PocketminePluginSuicide(),
		)))
		return false
	}

	name := plugin.GetDescription().GetName()
	if _, ok := m.enabledPlugins[name]; !ok {
		m.enabledOrder = append(m.enabledOrder, name)
	}
	m.enabledPlugins[name] = plugin

	for _, dependency := range plugin.GetDescription().GetDepend() {
		m.addDependent(dependency, name)
	}
	for _, dependency := range plugin.GetDescription().GetSoftDepend() {
		if _, ok := m.plugins[dependency]; ok {
			m.addDependent(dependency, name)
		}
	}

	event.Call(pluginevent.NewPluginEnableEvent(plugin))

	return true
}

func (m *PluginManager) addDependent(dependency, dependent string) {
	if m.pluginDependents[dependency] == nil {
		m.pluginDependents[dependency] = map[string]bool{}
	}
	m.pluginDependents[dependency][dependent] = true
}

// DisablePlugins is a port of PluginManager::disablePlugins: disables every plugin, dependents
// before their dependencies.
func (m *PluginManager) DisablePlugins() {
	for len(m.enabledPlugins) > 0 {
		progressed := false
		for _, name := range append([]string(nil), m.enabledOrder...) {
			plugin, ok := m.enabledPlugins[name]
			if !ok || !plugin.IsEnabled() {
				continue //in case a plugin disabled another plugin
			}
			if dependents := m.pluginDependents[name]; len(dependents) > 0 {
				names := make([]string, 0, len(dependents))
				for d := range dependents {
					names = append(names, d)
				}
				sort.Strings(names)
				m.server.GetLogger().Debug("Deferring disable of plugin " + name + " due to dependent plugins still enabled: " + strings.Join(names, ", "))
				continue
			}

			m.DisablePlugin(plugin)
			progressed = true
		}
		if !progressed {
			// PHP loops forever here if a plugin is enabled but not marked as such; stop instead.
			for _, name := range append([]string(nil), m.enabledOrder...) {
				if plugin, ok := m.enabledPlugins[name]; ok {
					if !plugin.IsEnabled() {
						m.removeEnabled(name)
						continue
					}
					m.DisablePlugin(plugin)
				}
			}
		}
	}
}

func (m *PluginManager) removeEnabled(name string) {
	delete(m.enabledPlugins, name)
	for i, n := range m.enabledOrder {
		if n == name {
			m.enabledOrder = append(m.enabledOrder[:i], m.enabledOrder[i+1:]...)
			break
		}
	}
}

// DisablePlugin is a port of PluginManager::disablePlugin.
func (m *PluginManager) DisablePlugin(plugin Plugin) {
	if !plugin.IsEnabled() {
		return
	}
	name := plugin.GetDescription().GetName()
	m.server.GetLogger().Info(m.translate(lang.KnownTranslationFactory.PocketminePluginDisable(plugin.GetDescription().GetFullName())))
	event.Call(pluginevent.NewPluginDisableEvent(plugin))

	m.removeEnabled(name)
	for dependency, dependents := range m.pluginDependents {
		if dependents[name] {
			if len(dependents) == 1 {
				delete(m.pluginDependents, dependency)
			} else {
				delete(dependents, name)
			}
		}
	}

	_ = plugin.OnEnableStateChange(false)
	plugin.GetScheduler().Shutdown()
	event.Global().UnregisterAllForPlugin(plugin)
}

// TickSchedulers is a port of PluginManager::tickSchedulers.
func (m *PluginManager) TickSchedulers(currentTick int) {
	for _, name := range append([]string(nil), m.enabledOrder...) {
		if p, ok := m.enabledPlugins[name]; ok {
			//the plugin may have been disabled as a result of updating other plugins' schedulers, and therefore
			//removed from enabledPlugins
			_ = p.GetScheduler().MainThreadHeartbeat(currentTick)
		}
	}
}

// ClearPlugins is a port of PluginManager::clearPlugins.
func (m *PluginManager) ClearPlugins() {
	m.DisablePlugins()
	m.plugins = map[string]Plugin{}
	m.pluginOrder = nil
	m.enabledPlugins = map[string]Plugin{}
	m.enabledOrder = nil
	m.fileAssociations = map[string]PluginLoader{}
	m.loaderOrder = nil
}

// ListenerTags is how a Listener gives RegisterEvents the doc-comment tags PHP reads from each
// handler method (@priority, @handleCancelled, @notHandler; see event.ListenerTag*): Go keeps no
// comments at runtime. The result is keyed by method name, then by tag name.
type ListenerTags interface {
	EventHandlerTags() map[string]map[string]string
}

// cancellable is event\Cancellable.
type cancellable interface{ IsCancelled() bool }

// RegisterEvents is a port of PluginManager::registerEvents: every exported method of listener
// that takes a single event (a pointer to a struct) and returns nothing is registered as a
// handler of that event.
func (m *PluginManager) RegisterEvents(listener event.Listener, plugin Plugin) error {
	if !plugin.IsEnabled() {
		return &Exception{Message: fmt.Sprintf("Plugin attempted to register %T while not enabled", listener)}
	}

	var tags map[string]map[string]string
	if t, ok := listener.(ListenerTags); ok {
		tags = t.EventHandlerTags()
	}

	value := reflect.ValueOf(listener)
	listenerType := value.Type()
	for i := 0; i < listenerType.NumMethod(); i++ {
		method := listenerType.Method(i)
		methodTags := tags[method.Name]
		if _, notHandler := methodTags[event.ListenerTagNotHandler]; notHandler {
			continue
		}
		fn := value.Method(i)
		if fn.Type().NumIn() != 1 || fn.Type().NumOut() != 0 {
			continue
		}
		eventType := fn.Type().In(0)
		if eventType.Kind() != reflect.Pointer || eventType.Elem().Kind() != reflect.Struct {
			continue
		}
		handlerName := fmt.Sprintf("%s.%s", listenerType, method.Name)

		priority := event.Normal
		if raw, ok := methodTags[event.ListenerTagPriority]; ok {
			p, err := event.PriorityFromString(raw)
			if err != nil {
				return &Exception{Message: fmt.Sprintf("Event handler %s() declares invalid/unknown priority \"%s\"", handlerName, raw)}
			}
			priority = p
		}

		handleCancelled := false
		if raw, ok := methodTags[event.ListenerTagHandleCancelled]; ok {
			if !eventType.Implements(reflect.TypeOf((*cancellable)(nil)).Elem()) {
				return &Exception{Message: fmt.Sprintf("Event handler %s() declares @%s for non-cancellable event of type %s", handlerName, event.ListenerTagHandleCancelled, eventType.Elem())}
			}
			switch strings.ToLower(raw) {
			case "true", "":
				handleCancelled = true
			case "false":
			default:
				return &Exception{Message: fmt.Sprintf("Event handler %s() declares invalid @%s value \"%s\"", handlerName, event.ListenerTagHandleCancelled, raw)}
			}
		}

		event.RegisterListenerOfType(event.Global(), eventType, plugin, priority, handleCancelled, func(e any) {
			fn.Call([]reflect.Value{reflect.ValueOf(e)})
		})
	}
	return nil
}

// RegisterEvent is a port of PluginManager::registerEvent: registers handler for events of type
// *E, owned by plugin (so it's unregistered when the plugin is disabled). A function because Go
// methods can't be generic.
func RegisterEvent[E any](m *PluginManager, handler func(e *E), priority event.Priority, plugin Plugin, handleCancelled bool) (event.ListenerHandle, error) {
	if !plugin.IsEnabled() {
		return event.ListenerHandle{}, &Exception{Message: fmt.Sprintf("Plugin attempted to register event handler to event %s while not enabled", reflect.TypeOf((*E)(nil)).Elem())}
	}
	return event.RegisterListener(event.Global(), plugin, priority, handleCancelled, handler), nil
}
