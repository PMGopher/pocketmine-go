package defaults

import (
	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/permission"
)

// StopCommand is a port of pocketmine\command\defaults\StopCommand.
type StopCommand struct{ VanillaCommand }

func NewStopCommand() *StopCommand {
	return &StopCommand{newVanillaCommand("stop", lang.KnownTranslationFactory.PocketmineCommandStopDescription(), nil, nil, permission.CommandStop)}
}

func (c *StopCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.CommandsStopStart(), true)
	srv(sender).Shutdown()
	return true, nil
}
