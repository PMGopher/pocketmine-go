package defaults

import (
	"sort"
	"strconv"
	"strings"

	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/permission"
	"pocketmine-go/pocketmine/utils"
)

// PluginsCommand is a port of pocketmine\command\defaults\PluginsCommand.
type PluginsCommand struct{ VanillaCommand }

func NewPluginsCommand() *PluginsCommand {
	return &PluginsCommand{newVanillaCommand("plugins", lang.KnownTranslationFactory.PocketmineCommandPluginsDescription(), nil, []string{"pl"}, permission.CommandPlugins)}
}

func (c *PluginsCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	var list []string
	for _, p := range srv(sender).GetPluginManager().GetPlugins() {
		color := utils.Red
		if p.IsEnabled() {
			color = utils.Green
		}
		list = append(list, color+p.GetDescription().GetFullName())
	}
	sort.Strings(list)

	sender.SendMessage(lang.KnownTranslationFactory.PocketmineCommandPluginsSuccess(strconv.Itoa(len(list)), strings.Join(list, utils.Reset+", ")))
	return true, nil
}
