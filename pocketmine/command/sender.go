package command

import (
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/permission"
)

// Sender is a port of pocketmine\command\CommandSender.
//
// PHP's version extends the Permissible interface; this port's permission package collapsed
// PermissibleInternal/PermissibleBase into one concrete *permission.Permissible type rather than
// keeping an interface (see permission.Permissible's doc comment for why), so a concrete Sender
// implementation (a future Player or console sender) is expected to embed *permission.Permissible
// by value to get HasPermission/AddAttachment/etc. for free via method promotion, and add these
// remaining methods itself.
type Sender interface {
	HasPermission(name string) bool
	IsPermissionSet(name string) bool
	AddAttachment(plugin permission.Plugin, name string, value *bool) (*permission.PermissionAttachment, error)
	RemoveAttachment(attachment *permission.PermissionAttachment)
	RecalculatePermissions() map[string]bool
	GetEffectivePermissions() map[string]*permission.AttachmentInfo

	GetLanguage() *lang.Language
	// SendMessage accepts a string or *lang.Translatable, matching PHP's Translatable|string union.
	SendMessage(message any)
	GetServer() Server
	GetName() string

	// GetScreenLineHeight returns the line height of the sender's screen, for command output
	// pagination (e.g. /help). SetScreenLineHeight(nil) resets it to the default.
	GetScreenLineHeight() int
	SetScreenLineHeight(height *int)
}

// Server is the minimal surface command dispatch needs from the server. Declared locally (like
// permission.Plugin) so the not-yet-ported Server type satisfies it automatically once it exists.
type Server interface {
	GetBroadcastChannelSubscribers(channel string) []Sender
	GetCommandMap() CommandMap
	GetCommandAliases() map[string][]string
	GetLanguage() *lang.Language
}

// BroadcastChannelAdministrative mirrors Server::BROADCAST_CHANNEL_ADMINISTRATIVE.
const BroadcastChannelAdministrative = "pocketmine.broadcast.admin"
