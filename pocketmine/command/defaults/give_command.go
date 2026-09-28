package defaults

import (
	"errors"
	"strconv"
	"strings"

	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/permission"
	"pocketmine-go/pocketmine/utils"
)

// GiveCommand is a port of pocketmine\command\defaults\GiveCommand.
type GiveCommand struct{ VanillaCommand }

func NewGiveCommand() *GiveCommand {
	c := &GiveCommand{VanillaCommand{Command: command.InitCommand("give", lang.KnownTranslationFactory.PocketmineCommandGiveDescription(), lang.KnownTranslationFactory.PocketmineCommandGiveUsage(), nil)}}
	_ = c.SetPermissions([]string{permission.CommandGiveSelf, permission.CommandGiveOther})
	return c
}

// parseItem is `StringToItemParser::getInstance()->parse($s) ?? LegacyStringToItemParser::getInstance()->parse($s)`.
func parseItem(s string) (item.Item, error) {
	if it, ok := item.GetStringToItemParser().Parse(s); ok {
		return it, nil
	}
	return item.GetLegacyStringToItemParser().Parse(s)
}

func (c *GiveCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	if len(args) < 2 {
		return syntaxError()
	}

	p, err := c.fetchPermittedPlayerTarget(sender, &args[0], permission.CommandGiveSelf, permission.CommandGiveOther)
	if err != nil || p == nil {
		return true, err
	}

	it, err := parseItem(args[1])
	if err != nil {
		var notFound *item.LegacyStringToItemParserError
		if errors.As(err, &notFound) {
			sender.SendMessage(lang.KnownTranslationFactory.CommandsGiveItemNotFound(args[1]).Prefix(utils.Red))
			return true, nil
		}
		return nil, err
	}

	if len(args) < 3 {
		it.SetCount(it.GetMaxStackSize())
	} else {
		count, err := getBoundedInt(sender, args[2], 1, 32767)
		if err != nil || count == nil {
			return true, err
		}
		it.SetCount(*count)
	}

	if len(args) > 3 {
		data := strings.Join(args[3:], " ")
		tags, err := nbt.ParseJson(data)
		if err != nil {
			sender.SendMessage(lang.KnownTranslationFactory.CommandsGiveTagError(err.Error()))
			return true, nil
		}
		// PHP also reports an NbtException from setNamedTag here (a known tag with the wrong type);
		// Item.SetNamedTag ignores such tags instead of failing.
		it.SetNamedTag(tags)
	}

	//TODO: overflow
	p.GetInventory().AddItem(it)

	command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.CommandsGiveSuccess(
		it.GetName()+" ("+args[1]+")",
		strconv.Itoa(it.GetCount()),
		p.GetName(),
	), true)
	return true, nil
}
