package defaults

import (
	"strings"

	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/permission"
)

// TitleCommand is a port of pocketmine\command\defaults\TitleCommand.
type TitleCommand struct{ VanillaCommand }

func NewTitleCommand() *TitleCommand {
	c := &TitleCommand{VanillaCommand{Command: command.InitCommand("title", lang.KnownTranslationFactory.PocketmineCommandTitleDescription(), lang.KnownTranslationFactory.CommandsTitleUsage(), nil)}}
	_ = c.SetPermissions([]string{permission.CommandTitleSelf, permission.CommandTitleOther})
	return c
}

func (c *TitleCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	if len(args) < 2 {
		return syntaxError()
	}
	p, err := c.fetchPermittedPlayerTarget(sender, &args[0], permission.CommandTitleSelf, permission.CommandTitleOther)
	if err != nil || p == nil {
		return true, err
	}

	switch args[1] {
	case "clear":
		p.RemoveTitles()
	case "reset":
		p.ResetTitles()
	case "title":
		if len(args) < 3 {
			return syntaxError()
		}
		p.SendTitle(strings.Join(args[2:], " "), "", -1, -1, -1)
	case "subtitle":
		if len(args) < 3 {
			return syntaxError()
		}
		p.SendSubTitle(strings.Join(args[2:], " "))
	case "actionbar":
		if len(args) < 3 {
			return syntaxError()
		}
		p.SendActionBarMessage(strings.Join(args[2:], " "))
	case "times":
		if len(args) < 5 {
			return syntaxError()
		}
		p.SetTitleDuration(getInteger(args[2], MinCoord, MaxCoord), getInteger(args[3], MinCoord, MaxCoord), getInteger(args[4], MinCoord, MaxCoord))
	default:
		return syntaxError()
	}

	sender.SendMessage(lang.KnownTranslationFactory.CommandsTitleSuccess())
	return true, nil
}
