package main

// Plugins are Go packages compiled into the server (see plugin.RegisterGoPlugin). Fetch each
// plugin module, import it here so its init() registers it, then rebuild the server:
//
//	go get github.com/PMGopher/example@latest
//
//	import (
//		_ "github.com/PMGopher/example" // the example plugin
//		_ "github.com/you/yourplugin"   // any other plugin module
//	)
//
// The PluginManager then loads it on startup like a plugin from the plugins folder
// (plugin_list.toml, API version, dependencies, load order, commands and permissions from its
// plugin.toml, data folder in plugin_data/<name>/). See
// https://github.com/PMGopher/example and the README's "Writing plugins" section.
