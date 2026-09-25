// Package plugin is a port of pocketmine\event\plugin: plugin lifecycle events.
//
// Importers conventionally alias this package as pluginevent.
package plugin

// Plugin is the surface these events need from pocketmine\plugin\Plugin.
type Plugin interface {
	GetName() string
}

// PluginEvent is a port of pocketmine\event\plugin\PluginEvent.
type PluginEvent struct {
	plugin Plugin
}

func (e *PluginEvent) GetPlugin() Plugin { return e.plugin }

// PluginEnableEvent is a port of pocketmine\event\plugin\PluginEnableEvent.
type PluginEnableEvent struct {
	PluginEvent
}

func NewPluginEnableEvent(plugin Plugin) *PluginEnableEvent {
	return &PluginEnableEvent{PluginEvent: PluginEvent{plugin: plugin}}
}

// PluginDisableEvent is a port of pocketmine\event\plugin\PluginDisableEvent.
type PluginDisableEvent struct {
	PluginEvent
}

func NewPluginDisableEvent(plugin Plugin) *PluginDisableEvent {
	return &PluginDisableEvent{PluginEvent: PluginEvent{plugin: plugin}}
}
