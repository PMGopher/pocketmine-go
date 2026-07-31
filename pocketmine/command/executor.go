package command

import "pocketmine-go/pocketmine/permission"

// Executor is a port of pocketmine\command\CommandExecutor (used to route PluginCommand calls to
// a plugin's OnCommand method).
type Executor interface {
	OnCommand(sender Sender, cmd CommandLike, label string, args []string) bool
}

// PluginOwned is a port of pocketmine\plugin\PluginOwned, declared locally for the same reason
// as permission.Plugin — it's structurally satisfied by the future plugin package's types with
// no import needed here.
type PluginOwned interface {
	GetOwningPlugin() permission.Plugin
}
