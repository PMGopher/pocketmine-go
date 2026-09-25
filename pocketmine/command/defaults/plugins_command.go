package defaults

import (
	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/permission"
)

// PluginsCommand is a port of pocketmine\command\defaults\PluginsCommand. Plugins aren't ported
// (see AGENTS.md §6 Phase 4), so the list is always empty.
type PluginsCommand struct{ VanillaCommand }

func NewPluginsCommand() *PluginsCommand {
	return &PluginsCommand{newVanillaCommand("plugins", lang.KnownTranslationFactory.PocketmineCommandPluginsDescription(), nil, []string{"pl"}, permission.CommandPlugins)}
}

func (c *PluginsCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	sender.SendMessage(lang.KnownTranslationFactory.PocketmineCommandPluginsSuccess("0", ""))
	return true, nil
}
