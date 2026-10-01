package plugin

// PluginLoader is a port of pocketmine\plugin\PluginLoader: handles different types of plugins.
type PluginLoader interface {
	// CanLoadPlugin returns whether this PluginLoader can load the plugin in the given path.
	CanLoadPlugin(path string) bool
	// LoadPlugin loads the plugin contained in file.
	LoadPlugin(file string)
	// GetPluginDescription gets the Description from the file (nil if it has none). Errors are
	// *PluginDescriptionParseException for an invalid plugin manifest.
	GetPluginDescription(file string) (*Description, error)
	// GetAccessProtocol returns the protocol prefix used to access files in this plugin, e.g.
	// file://, phar://.
	GetAccessProtocol() string
}

// pluginFactory is what PluginManager needs from a loader to create the plugin's main object.
// PHP creates it with `new $mainClass(...)` after loadPlugin() made the class available; a Go
// loader returns the constructor its plugin registered instead. ok is false when there's no such
// main (PHP: pocketmine.plugin.mainClassNotFound).
type pluginFactory interface {
	NewPlugin(file string, description *Description) (plugin Plugin, ok bool)
	GetResourceProvider(file string) ResourceProvider
}

// pluginSource is a loader whose plugins aren't files in the plugins folder (compiled-in Go
// plugins): triagePlugins also looks at the paths it returns for that folder.
type pluginSource interface {
	GetPluginFiles(dir string) []string
}
