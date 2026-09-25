package defaults

import (
	"strings"

	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/console"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/permission"
	"pocketmine-go/pocketmine/player"
	"pocketmine-go/pocketmine/utils"
)

// SayCommand is a port of pocketmine\command\defaults\SayCommand.
type SayCommand struct{ VanillaCommand }

func NewSayCommand() *SayCommand {
	return &SayCommand{newVanillaCommand("say", lang.KnownTranslationFactory.PocketmineCommandSayDescription(), lang.KnownTranslationFactory.CommandsSayUsage(), nil, permission.CommandSay)}
}

func (c *SayCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	if len(args) == 0 {
		return syntaxError()
	}
	name := sender.GetName()
	if p, ok := sender.(*player.Player); ok {
		name = p.GetDisplayName()
	} else if _, ok := sender.(*console.ConsoleCommandSender); ok {
		name = "Server"
	}
	srv(sender).BroadcastMessage(lang.KnownTranslationFactory.ChatTypeAnnouncement(name, strings.Join(args, " ")).Prefix(utils.LightPurple), nil)
	return true, nil
}
