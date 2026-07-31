package event

// PluginRef identifies whichever plugin registered a listener, used only for bulk
// unregistration when a plugin is disabled.
//
// This is `any` rather than a concrete *plugin.Plugin because the plugin package doesn't exist
// yet in this port; every caller here only ever treats it as an opaque, comparable identity
// token, so it will narrow to the real plugin type with no other changes once that package lands.
type PluginRef = any

// Listener is a port of pocketmine\event\Listener.
//
// PHP's version is an empty marker interface: PluginManager::registerEvents() reflects over an
// arbitrary object's methods to auto-discover handlers by parameter type, using doc-comment tags
// (@priority, @handleCancelled, @notHandler) to configure each one. Go strips comments before
// compilation — they aren't retained for runtime reflection — so that auto-discovery approach has
// no way to exist here. Registration is explicit and type-safe instead, via RegisterListener,
// which takes priority/handleCancelled as ordinary arguments instead of parsed annotations.
type Listener interface{}
