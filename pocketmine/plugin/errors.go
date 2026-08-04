package plugin

// Exception is a port of pocketmine\plugin\PluginException - the base error type this package's
// plugin-loading machinery returns.
type Exception struct{ Message string }

func (e *Exception) Error() string { return e.Message }

// DisablePluginException is a port of pocketmine\plugin\DisablePluginException - a bare marker
// error (extends \RuntimeException with an empty body in real PHP, no fields/message of its own)
// that plugin code can raise from any lifecycle method to request the plugin manager catch it and
// disable that plugin instead of letting it propagate further (see PluginManager::disablePlugin's
// own try/catch on this exact type).
type DisablePluginException struct{}

func (e *DisablePluginException) Error() string { return "plugin requested to be disabled" }

// PluginDescriptionParseException is a port of pocketmine\plugin\PluginDescriptionParseException -
// thrown while parsing a plugin.yml manifest.
type PluginDescriptionParseException struct{ Message string }

func (e *PluginDescriptionParseException) Error() string { return e.Message }
