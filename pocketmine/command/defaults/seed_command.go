package defaults

import (
	"strconv"

	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/permission"
	"pocketmine-go/pocketmine/player"
)

// SeedCommand is a port of pocketmine\command\defaults\SeedCommand.
type SeedCommand struct{ VanillaCommand }

func NewSeedCommand() *SeedCommand {
	return &SeedCommand{newVanillaCommand("seed", lang.KnownTranslationFactory.PocketmineCommandSeedDescription(), nil, nil, permission.CommandSeed)}
}

func (c *SeedCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	worldManager := srv(sender).GetWorldManager()
	w := worldManager.GetDefaultWorld()
	if p, ok := sender.(*player.Player); ok {
		w = p.GetWorld()
	}
	sender.SendMessage(lang.KnownTranslationFactory.CommandsSeedSuccess(strconv.FormatInt(worldManager.GetSeed(w), 10)))
	return true, nil
}
