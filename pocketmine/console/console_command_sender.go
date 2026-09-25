// Package console is a port of pocketmine\console: the console command sender and the reader
// that turns stdin lines into commands.
package console

import (
	"fmt"
	stdmath "math"
	"strings"

	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/permission"
	"pocketmine-go/pocketmine/utils"
)

// ConsoleCommandSender is a port of pocketmine\console\ConsoleCommandSender: the command sender
// commands typed in the server console run as. It has the console root permission.
type ConsoleCommandSender struct {
	*permission.Permissible

	server     command.Server
	language   *lang.Language
	lineHeight *int
}

func NewConsoleCommandSender(server command.Server, language *lang.Language) *ConsoleCommandSender {
	return &ConsoleCommandSender{
		Permissible: permission.NewPermissible(map[string]bool{permission.RootConsole: true}),
		server:      server,
		language:    language,
	}
}

func (s *ConsoleCommandSender) GetServer() command.Server { return s.server }

func (s *ConsoleCommandSender) GetLanguage() *lang.Language { return s.language }

// SendMessage is a port of ConsoleCommandSender::sendMessage: every line of the message is
// written to the terminal as command output.
func (s *ConsoleCommandSender) SendMessage(message any) {
	var text string
	switch m := message.(type) {
	case *lang.Translatable:
		text = s.language.Translate(m)
	case string:
		text = m
	default:
		text = fmt.Sprint(m)
	}
	for _, line := range strings.Split(strings.TrimSpace(text), "\n") {
		utils.WriteTerminalLine(utils.Green + "Command output | " + utils.AddBase(utils.White, line))
	}
}

func (s *ConsoleCommandSender) GetName() string { return "CONSOLE" }

func (s *ConsoleCommandSender) GetScreenLineHeight() int {
	if s.lineHeight != nil {
		return *s.lineHeight
	}
	return stdmath.MaxInt
}

// SetScreenLineHeight is a port of ConsoleCommandSender::setScreenLineHeight (panicking on a
// height below 1, like PHP's InvalidArgumentException).
func (s *ConsoleCommandSender) SetScreenLineHeight(height *int) {
	if height != nil && *height < 1 {
		panic("Line height must be at least 1")
	}
	s.lineHeight = height
}

var _ command.Sender = (*ConsoleCommandSender)(nil)
