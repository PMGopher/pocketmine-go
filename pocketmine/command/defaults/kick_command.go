package defaults

import (
	"strings"

	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/permission"
	"pocketmine-go/pocketmine/utils"
)

// KickCommand is a port of pocketmine\command\defaults\KickCommand.
type KickCommand struct{ VanillaCommand }

func NewKickCommand() *KickCommand {
	return &KickCommand{newVanillaCommand("kick", lang.KnownTranslationFactory.PocketmineCommandKickDescription(), lang.KnownTranslationFactory.CommandsKickUsage(), nil, permission.CommandKick)}
}

func (c *KickCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	if len(args) == 0 {
		return syntaxError()
	}
	name := args[0]
	reason := strings.TrimSpace(strings.Join(args[1:], " "))

	if p := srv(sender).GetPlayerByPrefix(name); p != nil {
		if reason != "" {
			p.Kick(lang.KnownTranslationFactory.PocketmineDisconnectKick(reason), nil, nil)
			command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.CommandsKickSuccessReason(p.GetName(), reason), true)
		} else {
			p.Kick(lang.KnownTranslationFactory.PocketmineDisconnectKickNoReason(), nil, nil)
			command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.CommandsKickSuccess(p.GetName()), true)
		}
	} else {
		sender.SendMessage(lang.KnownTranslationFactory.CommandsGenericPlayerNotFound().Prefix(utils.Red))
	}
	return true, nil
}
