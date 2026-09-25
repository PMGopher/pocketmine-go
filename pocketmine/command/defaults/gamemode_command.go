package defaults

import (
	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/permission"
	"pocketmine-go/pocketmine/player"
	"pocketmine-go/pocketmine/server"
)

// DefaultGamemodeCommand is a port of pocketmine\command\defaults\DefaultGamemodeCommand.
type DefaultGamemodeCommand struct{ VanillaCommand }

func NewDefaultGamemodeCommand() *DefaultGamemodeCommand {
	return &DefaultGamemodeCommand{newVanillaCommand("defaultgamemode", lang.KnownTranslationFactory.PocketmineCommandDefaultgamemodeDescription(), lang.KnownTranslationFactory.CommandsDefaultgamemodeUsage(), nil, permission.CommandDefaultGamemode)}
}

func (c *DefaultGamemodeCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	if len(args) == 0 {
		return syntaxError()
	}
	gameMode, ok := player.GameModeFromString(args[0])
	if !ok {
		sender.SendMessage(lang.KnownTranslationFactory.PocketmineCommandGamemodeUnknown(args[0]))
		return true, nil
	}
	//TODO: this probably shouldn't use the enum name directly
	srv(sender).GetConfigGroup().SetConfigString(server.PropertyGameMode, gameMode.Name())
	sender.SendMessage(lang.KnownTranslationFactory.CommandsDefaultgamemodeSuccess(gameMode.GetTranslatableName()))
	return true, nil
}

// GamemodeCommand is a port of pocketmine\command\defaults\GamemodeCommand.
type GamemodeCommand struct{ VanillaCommand }

func NewGamemodeCommand() *GamemodeCommand {
	c := &GamemodeCommand{VanillaCommand{Command: command.InitCommand("gamemode", lang.KnownTranslationFactory.PocketmineCommandGamemodeDescription(), lang.KnownTranslationFactory.CommandsGamemodeUsage(), nil)}}
	_ = c.SetPermissions([]string{permission.CommandGamemodeSelf, permission.CommandGamemodeOther})
	return c
}

func (c *GamemodeCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	if len(args) == 0 {
		return syntaxError()
	}
	gameMode, ok := player.GameModeFromString(args[0])
	if !ok {
		sender.SendMessage(lang.KnownTranslationFactory.PocketmineCommandGamemodeUnknown(args[0]))
		return true, nil
	}

	var targetName *string
	if len(args) > 1 {
		targetName = &args[1]
	}
	target, err := c.fetchPermittedPlayerTarget(sender, targetName, permission.CommandGamemodeSelf, permission.CommandGamemodeOther)
	if err != nil || target == nil {
		return true, err
	}

	if target.GetGamemode() == gameMode {
		sender.SendMessage(lang.KnownTranslationFactory.PocketmineCommandGamemodeFailure(target.GetName()))
		return true, nil
	}

	target.SetGamemode(gameMode)
	if gameMode != target.GetGamemode() {
		sender.SendMessage(lang.KnownTranslationFactory.PocketmineCommandGamemodeFailure(target.GetName()))
	} else if command.Sender(target) == sender {
		command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.CommandsGamemodeSuccessSelf(gameMode.GetTranslatableName()), true)
	} else {
		target.SendMessage(lang.KnownTranslationFactory.GameModeChanged(gameMode.GetTranslatableName()))
		command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.CommandsGamemodeSuccessOther(gameMode.GetTranslatableName(), target.GetName()), true)
	}
	return true, nil
}
