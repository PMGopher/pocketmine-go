package defaults

import (
	"errors"
	"sort"
	"strconv"

	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/permission"
	"pocketmine-go/pocketmine/utils"
)

// ClearCommand is a port of pocketmine\command\defaults\ClearCommand.
type ClearCommand struct{ VanillaCommand }

func NewClearCommand() *ClearCommand {
	c := &ClearCommand{VanillaCommand{Command: command.InitCommand("clear", lang.KnownTranslationFactory.PocketmineCommandClearDescription(), lang.KnownTranslationFactory.PocketmineCommandClearUsage(), nil)}}
	_ = c.SetPermissions([]string{permission.CommandClearSelf, permission.CommandClearOther})
	return c
}

func (c *ClearCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	if len(args) > 3 {
		return syntaxError()
	}

	var targetName *string
	if len(args) > 0 {
		targetName = &args[0]
	}
	target, err := c.fetchPermittedPlayerTarget(sender, targetName, permission.CommandClearSelf, permission.CommandClearOther)
	if err != nil || target == nil {
		return true, err
	}

	var targetItem item.Item
	maxCount := -1
	if len(args) > 1 {
		it, err := parseItem(args[1])
		if err != nil {
			var notFound *item.LegacyStringToItemParserError
			if errors.As(err, &notFound) {
				//vanilla checks this at argument parsing layer, can't come up with a better alternative
				sender.SendMessage(lang.KnownTranslationFactory.CommandsGiveItemNotFound(args[1]).Prefix(utils.Red))
				return true, nil
			}
			return nil, err
		}
		targetItem = it

		if len(args) > 2 {
			maxCount = getInteger(args[2], -1, MaxCoord)
			targetItem.SetCount(maxCount)
		}
	}

	// This is the order that vanilla would clear items in.
	inventories := []inventory.Inventory{
		target.GetInventory(),
		target.GetCursorInventory(),
		target.GetArmorInventory(),
		target.GetOffHandInventory(),
	}

	// Checking player's inventory for all the items matching the criteria
	if targetItem != nil && maxCount == 0 {
		count := countItems(inventories, targetItem)
		if count > 0 {
			sender.SendMessage(lang.KnownTranslationFactory.CommandsClearTesting(target.GetName(), strconv.Itoa(count)))
		} else {
			sender.SendMessage(lang.KnownTranslationFactory.CommandsClearFailureNoItems(target.GetName()).Prefix(utils.Red))
		}
		return true, nil
	}

	clearedCount := 0
	if targetItem == nil {
		// Clear all items from the inventories
		clearedCount += countItems(inventories, nil)
		for _, inv := range inventories {
			inv.ClearAll()
		}
	} else if maxCount == -1 {
		// Clear the item from target's inventory irrelevant of the count
		clearedCount += countItems(inventories, targetItem)
		for _, inv := range inventories {
			inv.Remove(targetItem)
		}
	} else {
		// Clear the item from target's inventory up to maxCount
	clearLoop:
		for _, inv := range inventories {
			for _, index := range sortedSlots(inv.All(targetItem)) {
				it := inv.GetItem(index)
				// The count to reduce from the item and max count
				reductionCount := min(it.GetCount(), maxCount)
				it.PopCount(reductionCount)
				clearedCount += reductionCount
				inv.SetItem(index, it)

				maxCount -= reductionCount
				if maxCount <= 0 {
					break clearLoop
				}
			}
		}
	}

	if clearedCount > 0 {
		command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.CommandsClearSuccess(target.GetName(), strconv.Itoa(clearedCount)), true)
	} else {
		sender.SendMessage(lang.KnownTranslationFactory.CommandsClearFailureNoItems(target.GetName()).Prefix(utils.Red))
	}
	return true, nil
}

// countItems is a port of ClearCommand::countItems.
func countItems(inventories []inventory.Inventory, target item.Item) int {
	count := 0
	for _, inv := range inventories {
		var contents map[int]item.Item
		if target != nil {
			contents = inv.All(target)
		} else {
			contents = inv.GetContents(false)
		}
		for _, it := range contents {
			count += it.GetCount()
		}
	}
	return count
}

// sortedSlots is the slot order of the array Inventory::all returns (ascending slot).
func sortedSlots(items map[int]item.Item) []int {
	slots := make([]int, 0, len(items))
	for slot := range items {
		slots = append(slots, slot)
	}
	sort.Ints(slots)
	return slots
}
