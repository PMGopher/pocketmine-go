// Package form is a port of pocketmine\form: API for Minecraft: Bedrock custom UI (forms).
package form

import "encoding/json"

// Form is a port of pocketmine\form\Form: a form sent to a player with Player.SendForm. It's
// marshalled to JSON (PHP's JsonSerializable) to be sent to the client.
type Form interface {
	json.Marshaler
	// HandleResponse handles a form response from a player (a *player.Player). data is the
	// decoded JSON response (nil when the form was closed). It returns a *FormValidationError if
	// the data could not be processed.
	HandleResponse(player any, data any) error
}

// FormValidationError is a port of pocketmine\form\FormValidationException.
type FormValidationError struct{ Message string }

func (e *FormValidationError) Error() string { return e.Message }
