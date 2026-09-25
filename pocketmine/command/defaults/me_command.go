package defaults

import (
	"strings"

	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/permission"
	"pocketmine-go/pocketmine/utils"
)

// MeCommand is a port of pocketmine\command\defaults\MeCommand.
type MeCommand struct{ VanillaCommand }

func NewMeCommand() *MeCommand {
	return &MeCommand{newVanillaCommand("me", lang.KnownTranslationFactory.PocketmineCommandMeDescription(), lang.KnownTranslationFactory.CommandsMeUsage(), nil, permission.CommandMe)}
}

func (c *MeCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	if len(args) == 0 {
		return syntaxError()
	}
	srv(sender).BroadcastMessage(lang.KnownTranslationFactory.ChatTypeEmote(senderName(sender), utils.Reset+strings.Join(args, " ")), nil)
	return true, nil
}
