package plugin

import (
	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/scheduler"
)

// Plugin is a port of pocketmine\plugin\Plugin.
//
// PHP plugins are PHP classes loaded at runtime, which a Go server can't do. Plugins for this
// server are Go packages compiled into the binary: they register themselves with
// RegisterGoPlugin (see go_plugin_loader.go) and are then loaded, enabled and disabled by the
// PluginManager exactly like PHP plugins. PHP's constructor
// (`__construct(PluginLoader, Server, PluginDescription, string $dataFolder, string $file, ResourceProvider)`)
// is Initializer.InitPlugin, which PluginBase implements.
type Plugin interface {
	IsEnabled() bool
	// OnEnableStateChange is called by the plugin manager when the plugin is enabled or disabled
	// to inform the plugin of its enabled state. A *DisablePluginException returned while
	// enabling makes the manager disable the plugin again (PHP: `throw new
	// DisablePluginException` in onEnable()).
	OnEnableStateChange(enabled bool) error
	// GetDataFolder is the plugin's data folder to save files and configuration, with a
	// trailing slash.
	GetDataFolder() string
	GetDescription() *Description
	GetName() string
	// Name is GetName, for permission.Plugin.
	Name() string
	GetLogger() log.AttachableLogger
	GetPluginLoader() PluginLoader
	GetScheduler() *scheduler.TaskScheduler
}

// Initializer is Plugin::__construct: it's called once, right after the plugin's main type is
// created. self is the plugin itself (the concrete type embedding PluginBase), which PluginBase
// needs to call the plugin's OnLoad/OnEnable/OnDisable/OnCommand.
type Initializer interface {
	InitPlugin(self Plugin, loader PluginLoader, server Server, description *Description, dataFolder string, file string, resourceProvider ResourceProvider)
}

// Server is what this package needs from pocketmine\Server. The server package imports this one,
// so it can't be imported here; *server.Server satisfies this interface, and plugins that need
// more type-assert GetServer() to *server.Server.
type Server interface {
	GetLogger() log.Logger
	GetLanguage() *lang.Language
	GetSimpleCommandMap() *command.SimpleCommandMap
	GetApiVersion() string
}
