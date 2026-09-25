// Package transaction is a port of pocketmine\inventory\transaction and its action sub-namespace:
// validated, atomic groups of inventory changes made by players.
//
// The actions (pocketmine\inventory\transaction\action) live in this package too: an action is
// told when it's added to a transaction (onAddToTransaction), and a transaction holds actions, so
// separate packages would import each other.
package transaction

import "pocketmine-go/pocketmine/inventory"

// TransactionValidationError is a port of TransactionValidationException (shared with the
// inventory package's slot validators).
type TransactionValidationError = inventory.TransactionValidationError

// TransactionCancelledError is a port of TransactionCancelledException: a plugin cancelled the
// transaction, or one of its actions.
type TransactionCancelledError struct{ Message string }

func (e *TransactionCancelledError) Error() string { return e.Message }

func validationError(format string) *TransactionValidationError {
	return &TransactionValidationError{Message: format}
}
