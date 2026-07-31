package command

import (
	"fmt"
	"strings"
	"sync"

	"pocketmine-go/pocketmine/log"
)

// SimpleCommandMap is a port of pocketmine\command\SimpleCommandMap.
//
// PHP's constructor auto-registers 42 built-in commands (ban, kick, gamemode, etc.) via
// setDefaultCommands(); none of those are ported yet (they depend on Player/World/Server, which
// don't exist in this port yet), so NewSimpleCommandMap starts empty. Once those commands are
// ported, register them explicitly via RegisterAll — there's no hidden auto-registration to
// replicate here.
type SimpleCommandMap struct {
	mu            sync.Mutex
	knownCommands map[string]CommandLike
	server        Server
}

func NewSimpleCommandMap(server Server) *SimpleCommandMap {
	return &SimpleCommandMap{knownCommands: map[string]CommandLike{}, server: server}
}

var _ CommandMap = (*SimpleCommandMap)(nil)

func (m *SimpleCommandMap) RegisterAll(fallbackPrefix string, commands []CommandLike) {
	for _, cmd := range commands {
		m.Register(fallbackPrefix, cmd, "")
	}
}

// Register mirrors SimpleCommandMap::register(). Panics if cmd has no permissions set — PHP
// throws \InvalidArgumentException here, and since this is a programmer error (a plugin author
// forgot to configure required permissions) rather than a runtime condition callers are expected
// to recover from, this follows the same "propagates up -> panic" convention used elsewhere for
// that class of PHP exception in this port.
func (m *SimpleCommandMap) Register(fallbackPrefix string, cmd CommandLike, label string) bool {
	if len(cmd.Permissions()) == 0 {
		panic("Commands must have a permission set")
	}

	if label == "" {
		label = cmd.Label()
	}
	label = strings.TrimSpace(label)
	fallbackPrefix = strings.ToLower(strings.TrimSpace(fallbackPrefix))

	m.mu.Lock()
	defer m.mu.Unlock()

	registered := m.registerAliasLocked(cmd, false, fallbackPrefix, label)

	aliases := append([]string{}, cmd.Aliases()...)
	var kept []string
	for _, alias := range aliases {
		if m.registerAliasLocked(cmd, true, fallbackPrefix, alias) {
			kept = append(kept, alias)
		}
	}
	cmd.SetAliases(kept)

	if !registered {
		cmd.SetLabel(fallbackPrefix + ":" + label)
	}
	cmd.Register(m)

	return registered
}

// registerAliasLocked must be called with m.mu held.
//
// PHP also special-cases `$command instanceof VanillaCommand` in the isAlias branch below (the
// defaults/ commands aren't ported yet — see the type doc comment — so there's no VanillaCommand
// type to check against here; revisit this once they are).
func (m *SimpleCommandMap) registerAliasLocked(cmd CommandLike, isAlias bool, fallbackPrefix string, label string) bool {
	m.knownCommands[fallbackPrefix+":"+label] = cmd

	if isAlias {
		if _, exists := m.knownCommands[label]; exists {
			return false
		}
	}

	if existing, exists := m.knownCommands[label]; exists && existing.Label() == label {
		return false
	}

	if !isAlias {
		cmd.SetLabel(label)
	}
	m.knownCommands[label] = cmd
	return true
}

func (m *SimpleCommandMap) Unregister(cmd CommandLike) bool {
	m.mu.Lock()
	for label, c := range m.knownCommands {
		if c == cmd {
			delete(m.knownCommands, label)
		}
	}
	m.mu.Unlock()

	cmd.Unregister(m)
	return true
}

// Dispatch is a port of SimpleCommandMap::dispatch().
//
// PHP formats the "command not found" and usage messages via KnownTranslationFactory; deferred
// the same way as Command.TestPermission, pending the generated translation factory.
func (m *SimpleCommandMap) Dispatch(sender Sender, cmdLine string) bool {
	args := ParseQuoteAware(cmdLine)
	if len(args) == 0 {
		sender.SendMessage(`Unknown command. Type "/help" for help.`)
		return false
	}

	sentLabel, args := args[0], args[1:]
	target := m.GetCommand(sentLabel)
	if target == nil {
		sender.SendMessage(fmt.Sprintf("Unknown command: %q. Type \"/help\" for help.", sentLabel))
		return false
	}

	if target.TestPermission(sender, nil) {
		_, err := target.Execute(sender, sentLabel, args)
		if _, ok := err.(*InvalidCommandSyntaxException); ok {
			sender.SendMessage(fmt.Sprintf("Usage: %s", stringifyMessage(target.Usage())))
		}
	}
	return true
}

func (m *SimpleCommandMap) ClearCommands() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, cmd := range m.knownCommands {
		cmd.Unregister(m)
	}
	m.knownCommands = map[string]CommandLike{}
}

func (m *SimpleCommandMap) GetCommand(name string) CommandLike {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.knownCommands[name]
}

func (m *SimpleCommandMap) GetCommands() map[string]CommandLike {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make(map[string]CommandLike, len(m.knownCommands))
	for k, v := range m.knownCommands {
		result[k] = v
	}
	return result
}

// RegisterServerAliases is a port of SimpleCommandMap::registerServerAliases(): registers the
// `aliases` section of pocketmine.yml as FormattedCommandAlias commands.
func (m *SimpleCommandMap) RegisterServerAliases() {
	for alias, commandStrings := range m.server.GetCommandAliases() {
		if strings.Contains(alias, ":") {
			log.Global().Warning(fmt.Sprintf("Alias %q cannot contain \":\"", alias))
			continue
		}

		var targets, bad, recursive []string
		for _, commandString := range commandStrings {
			args := ParseQuoteAware(commandString)
			if len(args) == 0 {
				bad = append(bad, commandString)
				continue
			}
			commandName := args[0]
			switch {
			case m.GetCommand(commandName) == nil:
				bad = append(bad, commandString)
			case strings.EqualFold(commandName, alias):
				recursive = append(recursive, commandString)
			default:
				targets = append(targets, commandString)
			}
		}

		if len(recursive) > 0 {
			log.Global().Warning(fmt.Sprintf("Recursive alias %q ignoring commands: %s", alias, strings.Join(recursive, ", ")))
			continue
		}
		if len(bad) > 0 {
			log.Global().Warning(fmt.Sprintf("Alias %q contains unknown commands: %s", alias, strings.Join(bad, ", ")))
			continue
		}

		lowerAlias := strings.ToLower(alias)
		m.mu.Lock()
		if len(targets) > 0 {
			m.knownCommands[lowerAlias] = NewFormattedCommandAlias(lowerAlias, targets)
		} else {
			delete(m.knownCommands, lowerAlias)
		}
		m.mu.Unlock()
	}
}
