package main

// Plugins are Go packages compiled into the server (see plugin.RegisterGoPlugin). Import each
// plugin package here, so its init() registers it:
//
//	import _ "example.com/myplugin"
//
// The PluginManager then loads it on startup like PocketMine-MP loads a plugin from the plugins
// folder (plugin_list.yml, API version, dependencies, load order, commands and permissions from
// its plugin.yml, data folder in plugin_data/<name>/).
