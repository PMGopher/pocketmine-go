package defaults

import (
	"strconv"

	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/permission"
	"pocketmine-go/pocketmine/server"
	"pocketmine-go/pocketmine/world"
)

// DifficultyCommand is a port of pocketmine\command\defaults\DifficultyCommand.
type DifficultyCommand struct{ VanillaCommand }

func NewDifficultyCommand() *DifficultyCommand {
	return &DifficultyCommand{newVanillaCommand("difficulty", lang.KnownTranslationFactory.PocketmineCommandDifficultyDescription(), lang.KnownTranslationFactory.CommandsDifficultyUsage(), nil, permission.CommandDifficulty)}
}

func (c *DifficultyCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	if len(args) != 1 {
		return syntaxError()
	}
	difficulty := world.GetDifficultyFromString(args[0])
	if srv(sender).IsHardcore() {
		difficulty = world.DifficultyHard
	}
	if difficulty == -1 {
		return syntaxError()
	}
	srv(sender).GetConfigGroup().SetConfigInt(server.PropertyDifficulty, difficulty)

	//TODO: add per-world support
	for _, w := range srv(sender).GetWorldManager().GetWorlds() {
		w.SetDifficulty(difficulty)
	}
	command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.CommandsDifficultySuccess(strconv.Itoa(difficulty)), true)
	return true, nil
}
