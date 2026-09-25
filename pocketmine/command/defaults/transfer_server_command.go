package defaults

import (
	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/permission"
	"pocketmine-go/pocketmine/player"
	"pocketmine-go/pocketmine/utils"
)

// TransferServerCommand is a port of pocketmine\command\defaults\TransferServerCommand.
type TransferServerCommand struct{ VanillaCommand }

func NewTransferServerCommand() *TransferServerCommand {
	return &TransferServerCommand{newVanillaCommand("transferserver", lang.KnownTranslationFactory.PocketmineCommandTransferserverDescription(), lang.KnownTranslationFactory.PocketmineCommandTransferserverUsage(), nil, permission.CommandTransferServer)}
}

func (c *TransferServerCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	if len(args) < 1 {
		return syntaxError()
	}
	p, ok := sender.(*player.Player)
	if !ok {
		sender.SendMessage(lang.KnownTranslationFactory.PocketmineCommandErrorPlayerUserOnly().Prefix(utils.Red))
		return false, nil
	}
	port := 19132
	if len(args) > 1 {
		port = phpInt(args[1])
	}
	p.Transfer(args[0], port, nil)
	return true, nil
}
