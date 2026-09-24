package inventory

import "pocketmine-go/pocketmine/item"

// TransactionValidationError is a port of
// pocketmine\inventory\transaction\TransactionValidationException, as returned by slot validators.
type TransactionValidationError struct{ Message string }

func (e *TransactionValidationError) Error() string { return e.Message }

// SlotValidator is a port of pocketmine\inventory\transaction\action\validator\SlotValidator:
// validates a slot placement for inventory transactions. It lives in this package (rather than a
// transaction/action/validator sub-package like PHP) because its signature needs Inventory, and a
// sub-package importing this one couldn't be imported back by BaseInventory.
type SlotValidator interface {
	// Validate returns nil if the item can be placed in the slot, or the reason it can't.
	Validate(inv Inventory, it item.Item, slot int) *TransactionValidationError
}

// CallbackSlotValidator is a port of
// pocketmine\inventory\transaction\action\validator\CallbackSlotValidator.
type CallbackSlotValidator struct {
	validate func(inv Inventory, it item.Item, slot int) *TransactionValidationError
}

func NewCallbackSlotValidator(validate func(inv Inventory, it item.Item, slot int) *TransactionValidationError) *CallbackSlotValidator {
	return &CallbackSlotValidator{validate: validate}
}

func (c *CallbackSlotValidator) Validate(inv Inventory, it item.Item, slot int) *TransactionValidationError {
	return c.validate(inv, it, slot)
}
