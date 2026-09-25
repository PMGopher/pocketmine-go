package defaults

import (
	"slices"
	"strconv"
	"strings"

	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/permission"
	"pocketmine-go/pocketmine/player"
)

// ListCommand is a port of pocketmine\command\defaults\ListCommand.
type ListCommand struct{ VanillaCommand }

func NewListCommand() *ListCommand {
	return &ListCommand{newVanillaCommand("list", lang.KnownTranslationFactory.PocketmineCommandListDescription(), nil, nil, permission.CommandList)}
}

func (c *ListCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	var playerNames []string
	self, isPlayer := sender.(*player.Player)
	for _, p := range srv(sender).GetOnlinePlayers() {
		if !isPlayer || self.CanSee(p) {
			playerNames = append(playerNames, p.GetName())
		}
	}
	slices.Sort(playerNames)

	sender.SendMessage(lang.KnownTranslationFactory.CommandsPlayersList(strconv.Itoa(len(playerNames)), strconv.Itoa(srv(sender).GetMaxPlayers())))
	sender.SendMessage(strings.Join(playerNames, ", "))
	return true, nil
}
