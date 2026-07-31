package command

import (
	"fmt"
	"math"

	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/permission"
)

// BroadcastLoggerForwarder is a port of pocketmine\utils\BroadcastLoggerForwarder.
//
// Moved into the command package rather than utils (where PHP places it): it implements Sender,
// which is declared in this package, and utils already doesn't (and shouldn't) import command —
// command imports utils (for TextFormat), so the reverse would be an import cycle.
//
// Forwards any messages it receives via SendMessage() to the given logger. Used for forwarding
// chat messages and command audit log messages to the server log file. Broadcast subscribers are
// required to implement Sender, hence the handful of otherwise-unused methods below.
type BroadcastLoggerForwarder struct {
	*permission.Permissible
	server   Server
	logger   log.Logger
	language *lang.Language
}

func NewBroadcastLoggerForwarder(server Server, logger log.Logger, language *lang.Language) *BroadcastLoggerForwarder {
	return &BroadcastLoggerForwarder{
		Permissible: permission.NewPermissible(nil),
		server:      server,
		logger:      logger,
		language:    language,
	}
}

func (f *BroadcastLoggerForwarder) GetLanguage() *lang.Language { return f.language }

func (f *BroadcastLoggerForwarder) SendMessage(message any) {
	if t, ok := message.(*lang.Translatable); ok {
		f.logger.Info(f.language.Translate(t))
		return
	}
	if s, ok := message.(string); ok {
		f.logger.Info(s)
		return
	}
	f.logger.Info(fmt.Sprintf("%v", message))
}

func (f *BroadcastLoggerForwarder) GetServer() Server { return f.server }

func (f *BroadcastLoggerForwarder) GetName() string { return "Broadcast Logger Forwarder" }

func (f *BroadcastLoggerForwarder) GetScreenLineHeight() int { return math.MaxInt }

func (f *BroadcastLoggerForwarder) SetScreenLineHeight(height *int) {
	//NOOP
}

var _ Sender = (*BroadcastLoggerForwarder)(nil)
