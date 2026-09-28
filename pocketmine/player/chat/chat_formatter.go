// Package chat is a port of pocketmine\player\chat.
package chat

import (
	"strings"

	"pocketmine-go/pocketmine/lang"
)

// Formatter is a port of pocketmine\player\chat\ChatFormatter. Format returns either a plain
// string (used as-is) or a *lang.Translatable (translated per-recipient) - Go has no union return
// type, so both are folded into `any`; callers type-switch on the result, matching how PHP callers
// already have to handle `Translatable|string` themselves.
type Formatter interface {
	Format(username, message string) any
}

// StandardChatFormatter is a port of pocketmine\player\chat\StandardChatFormatter.
type StandardChatFormatter struct{}

// Format is a port of StandardChatFormatter::format.
func (StandardChatFormatter) Format(username, message string) any {
	return lang.KnownTranslationFactory.ChatTypeText(username, message)
}

// LegacyRawChatFormatter is a port of pocketmine\player\chat\LegacyRawChatFormatter: a template
// string containing {%0}/{%1} placeholders for the username/message respectively.
type LegacyRawChatFormatter struct {
	format string
}

func NewLegacyRawChatFormatter(format string) *LegacyRawChatFormatter {
	return &LegacyRawChatFormatter{format: format}
}

func (f *LegacyRawChatFormatter) Format(username, message string) any {
	r := strings.NewReplacer("{%0}", username, "{%1}", message)
	return r.Replace(f.format)
}
