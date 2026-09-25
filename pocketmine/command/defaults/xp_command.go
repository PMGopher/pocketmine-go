package defaults

import (
	stdmath "math"
	"strconv"
	"strings"

	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/permission"
	"pocketmine-go/pocketmine/utils"
)

// XpCommand is a port of pocketmine\command\defaults\XpCommand.
type XpCommand struct{ VanillaCommand }

func NewXpCommand() *XpCommand {
	c := &XpCommand{VanillaCommand{Command: command.InitCommand("xp", lang.KnownTranslationFactory.PocketmineCommandXpDescription(), lang.KnownTranslationFactory.PocketmineCommandXpUsage(), nil)}}
	_ = c.SetPermissions([]string{permission.CommandXPSelf, permission.CommandXPOther})
	return c
}

func (c *XpCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	if len(args) < 1 {
		return syntaxError()
	}
	var targetName *string
	if len(args) > 1 {
		targetName = &args[1]
	}
	p, err := c.fetchPermittedPlayerTarget(sender, targetName, permission.CommandXPSelf, permission.CommandXPOther)
	if err != nil || p == nil {
		return true, err
	}

	xpManager := p.GetXpManager()
	if strings.HasSuffix(args[0], "L") {
		xpLevelAttr := p.GetAttributeMap().Get(entity.AttributeExperienceLevel)
		if xpLevelAttr == nil {
			panic("experience level attribute should exist")
		}
		maxXpLevel := int(xpLevelAttr.GetMaxValue())
		currentXpLevel := xpManager.GetXpLevel()
		xpLevels := getInteger(args[0][:len(args[0])-1], -currentXpLevel, maxXpLevel-currentXpLevel)
		if xpLevels >= 0 {
			xpManager.AddXpLevels(xpLevels, false)
			sender.SendMessage(lang.KnownTranslationFactory.CommandsXpSuccessLevels(strconv.Itoa(xpLevels), p.GetName()))
		} else {
			xpLevels = -xpLevels
			xpManager.SubtractXpLevels(xpLevels)
			sender.SendMessage(lang.KnownTranslationFactory.CommandsXpSuccessNegativeLevels(strconv.Itoa(xpLevels), p.GetName()))
		}
	} else {
		xp := getInteger(args[0], MinCoord, stdmath.MaxInt32)
		if xp < 0 {
			sender.SendMessage(lang.KnownTranslationFactory.CommandsXpFailureWidthdrawXp().Prefix(utils.Red))
		} else {
			xpManager.AddXp(xp, false)
			sender.SendMessage(lang.KnownTranslationFactory.CommandsXpSuccess(strconv.Itoa(xp), p.GetName()))
		}
	}
	return true, nil
}
