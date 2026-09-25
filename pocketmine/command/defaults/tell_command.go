package defaults

import (
	"strings"

	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/permission"
	"pocketmine-go/pocketmine/utils"
)

// TellCommand is a port of pocketmine\command\defaults\TellCommand.
type TellCommand struct{ VanillaCommand }

func NewTellCommand() *TellCommand {
	return &TellCommand{newVanillaCommand("tell", lang.KnownTranslationFactory.PocketmineCommandTellDescription(), lang.KnownTranslationFactory.CommandsMessageUsage(), []string{"w", "msg"}, permission.CommandTell)}
}

func (c *TellCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	if len(args) < 2 {
		return syntaxError()
	}
	p := srv(sender).GetPlayerByPrefix(args[0])
	args = args[1:]

	if p != nil && command.Sender(p) == sender {
		sender.SendMessage(lang.KnownTranslationFactory.CommandsMessageSameTarget().Prefix(utils.Red))
		return true, nil
	}

	if p != nil {
		message := strings.Join(args, " ")
		sender.SendMessage(lang.KnownTranslationFactory.CommandsMessageDisplayOutgoing(p.GetDisplayName(), message).Prefix(utils.Gray + utils.Italic))
		p.SendMessage(lang.KnownTranslationFactory.CommandsMessageDisplayIncoming(senderName(sender), message).Prefix(utils.Gray + utils.Italic))
		command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.CommandsMessageDisplayOutgoing(p.GetDisplayName(), message), false)
	} else {
		sender.SendMessage(lang.KnownTranslationFactory.CommandsGenericPlayerNotFound())
	}
	return true, nil
}
