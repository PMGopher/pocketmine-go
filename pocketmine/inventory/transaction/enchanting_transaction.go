package transaction

import (
	"fmt"

	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/event"
	playerevent "pocketmine-go/pocketmine/event/player"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/item/enchantment"
)

// enchantingSource is the part of Player EnchantingTransaction needs beyond Player.
type enchantingSource interface {
	GetXpManager() *entity.ExperienceManager
	RegenerateEnchantmentSeed()
}

// EnchantingTransaction is a port of pocketmine\inventory\transaction\EnchantingTransaction.
type EnchantingTransaction struct {
	*InventoryTransaction

	inputItem  item.Item
	outputItem item.Item

	option *enchantment.EnchantingOption
	cost   int
}

// NewEnchantingTransaction is a port of EnchantingTransaction::__construct.
func NewEnchantingTransaction(source Player, option *enchantment.EnchantingOption, cost int) *EnchantingTransaction {
	t := &EnchantingTransaction{InventoryTransaction: &InventoryTransaction{}, option: option, cost: cost}
	t.Init(t, source)
	return t
}

// validateOutput is a port of EnchantingTransaction::validateOutput.
func (t *EnchantingTransaction) validateOutput() error {
	if t.inputItem == nil || t.outputItem == nil {
		panic("Expected that inputItem and outputItem are not null before validating output")
	}

	enchantedInput := item.EnchantItem(t.inputItem, t.option.GetEnchantments())
	if !t.outputItem.EqualsExact(enchantedInput) {
		return validationError("Invalid output item")
	}
	return nil
}

// validateFiniteResources is a port of EnchantingTransaction::validateFiniteResources.
func (t *EnchantingTransaction) validateFiniteResources(lapisSpent int) error {
	if lapisSpent != t.cost {
		return validationError(fmt.Sprintf("Expected the amount of lapis lazuli spent to be %d, but received %d", t.cost, lapisSpent))
	}

	xpLevel := t.xpManager().GetXpLevel()
	requiredXpLevel := t.option.GetRequiredXpLevel()

	if xpLevel < requiredXpLevel {
		return validationError(fmt.Sprintf("Player's XP level %d is less than the required XP level %d", xpLevel, requiredXpLevel))
	}
	//XP level cost is intentionally not checked here, as the required level may be lower than the cost, allowing
	//the option to be used with less XP than the cost - in this case, as much XP as possible will be deducted.
	return nil
}

func (t *EnchantingTransaction) xpManager() *entity.ExperienceManager {
	return t.source.(enchantingSource).GetXpManager()
}

// Validate is a port of EnchantingTransaction::validate.
func (t *EnchantingTransaction) Validate() error {
	if len(t.actions) < 1 {
		return validationError("Transaction must have at least one action to be executable")
	}

	outputs, inputs, err := t.MatchItems()
	if err != nil {
		return err
	}

	lapisSpent := 0
	for _, input := range inputs {
		if input.GetTypeId() == item.LAPIS_LAZULI {
			lapisSpent = input.GetCount()
		} else {
			if t.inputItem != nil {
				return validationError("Received more than 1 items to enchant")
			}
			t.inputItem = input
		}
	}

	if t.inputItem == nil {
		return validationError("No item to enchant received")
	}

	if outputCount := len(outputs); outputCount != 1 {
		return validationError(fmt.Sprintf("Expected 1 output item, but received %d", outputCount))
	}
	t.outputItem = outputs[0]

	if err := t.validateOutput(); err != nil {
		return err
	}

	if t.source.HasFiniteResources() {
		return t.validateFiniteResources(lapisSpent)
	}
	return nil
}

// Execute is a port of EnchantingTransaction::execute.
func (t *EnchantingTransaction) Execute() error {
	if err := t.InventoryTransaction.Execute(); err != nil {
		return err
	}

	if t.source.HasFiniteResources() {
		//If the required XP level is less than the XP cost, the option can be selected with less XP than the cost.
		//In this case, as much XP as possible will be taken.
		xpManager := t.xpManager()
		xpManager.SubtractXpLevels(min(t.cost, xpManager.GetXpLevel()))
	}
	t.source.(enchantingSource).RegenerateEnchantmentSeed()
	return nil
}

// CallExecuteEvent is a port of EnchantingTransaction::callExecuteEvent.
func (t *EnchantingTransaction) CallExecuteEvent() bool {
	if t.inputItem == nil || t.outputItem == nil {
		panic("Expected that inputItem and outputItem are not null before executing the event")
	}

	ev := playerevent.NewPlayerItemEnchantEvent(t.source, t, t.option, t.inputItem, t.outputItem, t.cost)
	event.Call(ev)
	return !ev.IsCancelled()
}
