package main

// Plugins are Go packages compiled into the server (see plugin.RegisterGoPlugin). Import each
// plugin package here, so its init() registers it, then rebuild the server:
//
//	import (
//		_ "pocketmine-go/plugins/example" // the example plugin in this repository
//		_ "github.com/you/yourplugin"      // any other Go module (go get it first)
//	)
//
// The PluginManager then loads it on startup like a plugin from the plugins folder
// (plugin_list.yml, API version, dependencies, load order, commands and permissions from its
// plugin.yml, data folder in plugin_data/<name>/). See plugins/example and the README's
// "Writing plugins" section.
