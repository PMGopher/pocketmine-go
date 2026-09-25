package defaults

import (
	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/permission"
	"pocketmine-go/pocketmine/player"
	"pocketmine-go/pocketmine/utils"
)

// OpCommand is a port of pocketmine\command\defaults\OpCommand.
type OpCommand struct{ VanillaCommand }

func NewOpCommand() *OpCommand {
	return &OpCommand{newVanillaCommand("op", lang.KnownTranslationFactory.PocketmineCommandOpDescription(), lang.KnownTranslationFactory.CommandsOpUsage(), nil, permission.CommandOpGive)}
}

func (c *OpCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	if len(args) == 0 {
		return syntaxError()
	}
	name := args[0]
	if !player.IsValidUserName(name) {
		return syntaxError()
	}
	srv(sender).AddOp(name)
	if p := srv(sender).GetPlayerExact(name); p != nil {
		p.SendMessage(lang.KnownTranslationFactory.CommandsOpMessage().Prefix(utils.Gray))
	}
	command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.CommandsOpSuccess(name), true)
	return true, nil
}

// DeopCommand is a port of pocketmine\command\defaults\DeopCommand.
type DeopCommand struct{ VanillaCommand }

func NewDeopCommand() *DeopCommand {
	return &DeopCommand{newVanillaCommand("deop", lang.KnownTranslationFactory.PocketmineCommandDeopDescription(), lang.KnownTranslationFactory.CommandsDeopUsage(), nil, permission.CommandOpTake)}
}

func (c *DeopCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	if len(args) == 0 {
		return syntaxError()
	}
	name := args[0]
	if !player.IsValidUserName(name) {
		return syntaxError()
	}
	srv(sender).RemoveOp(name)
	if p := srv(sender).GetPlayerExact(name); p != nil {
		p.SendMessage(lang.KnownTranslationFactory.CommandsDeopMessage().Prefix(utils.Gray))
	}
	command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.CommandsDeopSuccess(name), true)
	return true, nil
}
