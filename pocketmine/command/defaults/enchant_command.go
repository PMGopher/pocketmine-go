package defaults

import (
	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/item/enchantment"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/permission"
)

// EnchantCommand is a port of pocketmine\command\defaults\EnchantCommand.
type EnchantCommand struct{ VanillaCommand }

func NewEnchantCommand() *EnchantCommand {
	c := &EnchantCommand{VanillaCommand{Command: command.InitCommand("enchant", lang.KnownTranslationFactory.PocketmineCommandEnchantDescription(), lang.KnownTranslationFactory.CommandsEnchantUsage(), nil)}}
	_ = c.SetPermissions([]string{permission.CommandEnchantSelf, permission.CommandEnchantOther})
	return c
}

func (c *EnchantCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	if len(args) < 2 {
		return syntaxError()
	}

	p, err := c.fetchPermittedPlayerTarget(sender, &args[0], permission.CommandEnchantSelf, permission.CommandEnchantOther)
	if err != nil || p == nil {
		return true, err
	}

	it := p.GetInventory().GetItemInHand()

	if it.IsNull() {
		sender.SendMessage(lang.KnownTranslationFactory.CommandsEnchantNoItem())
		return true, nil
	}

	ench, ok := enchantment.GetStringToEnchantmentParser().Parse(args[1])
	if !ok {
		sender.SendMessage(lang.KnownTranslationFactory.CommandsEnchantNotFound(args[1]))
		return true, nil
	}

	level := 1
	if len(args) > 2 {
		l, err := getBoundedInt(sender, args[2], 1, ench.GetMaxLevel())
		if err != nil || l == nil {
			return false, err
		}
		level = *l
	}

	//this is necessary to deal with enchanted books, which are a different item type than regular books
	enchantedItem := item.EnchantItem(it, []*enchantment.EnchantmentInstance{enchantment.NewEnchantmentInstance(ench, level)})
	p.GetInventory().SetItemInHand(enchantedItem)

	command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.CommandsEnchantSuccess(p.GetName()), true)
	return true, nil
}
