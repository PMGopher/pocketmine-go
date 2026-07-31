package command

import (
	"fmt"
	"strings"

	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/permission"
)

// CommandLike is the interface CommandMap deals in: the concrete fields/methods every Command
// provides, plus Execute (which each concrete command type supplies itself).
//
// PHP's abstract class never calls its own abstract execute() internally — only external callers
// (SimpleCommandMap.dispatch, FormattedCommandAlias.execute) call $target->execute(...) on
// whatever concrete instance the map handed them. So unlike the SimpleLogger/PrefixedLogger
// situation elsewhere in this port, there's no virtual-dispatch pitfall here: a concrete command
// can freely embed *Command by value to get all of its methods promoted, and add its own Execute
// method, and the combination satisfies this interface with no extra plumbing.
type CommandLike interface {
	Name() string
	Label() string
	SetLabel(name string) bool
	Aliases() []string
	SetAliases(aliases []string)
	Permissions() []string
	SetPermissions(perms []string) error
	SetPermission(perm *string)
	TestPermission(target Sender, perm *string) bool
	TestPermissionSilent(target Sender, perm *string) bool
	Description() any
	SetDescription(d any)
	Usage() any
	SetUsage(u any)
	PermissionMessage() *string
	SetPermissionMessage(msg string)
	IsRegistered() bool
	Register(m CommandMap) bool
	Unregister(m CommandMap) bool
	String() string
	Execute(sender Sender, commandLabel string, args []string) (any, error)
}

// Command is a port of pocketmine\command\Command: the concrete, non-abstract fields/methods
// every command shares. Embed this by value in a concrete command type, which then only needs
// to provide its own Execute method to satisfy CommandLike.
type Command struct {
	name              string
	nextLabel         string
	label             string
	aliases           []string
	activeAliases     []string
	commandMap        CommandMap
	description       any // string or *lang.Translatable
	usageMessage      any
	perms             []string
	permissionMessage *string
}

// InitCommand mirrors the Command constructor. Call it from a concrete command's own
// constructor: `c.Command = command.InitCommand(name, description, usage, aliases)`.
func InitCommand(name string, description any, usageMessage any, aliases []string) Command {
	if description == nil {
		description = ""
	}
	if usageMessage == nil {
		usageMessage = "/" + name
	}
	c := Command{name: name, label: name, nextLabel: name, description: description, usageMessage: usageMessage}
	c.SetAliases(aliases)
	return c
}

func (c *Command) Name() string { return c.name }

func (c *Command) Permissions() []string { return c.perms }

func (c *Command) SetPermissions(perms []string) error {
	mgr := permission.GetManager()
	for _, p := range perms {
		if mgr.GetPermission(p) == nil {
			return fmt.Errorf("cannot use non-existing permission %q", p)
		}
	}
	c.perms = perms
	return nil
}

// SetPermission mirrors Command::setPermission(): a nil pointer clears every permission;
// otherwise it's a ";"-separated list, matching how PHP encodes multiple required permissions
// in plugin.yml as a single string.
func (c *Command) SetPermission(perm *string) {
	if perm == nil {
		c.perms = nil
		return
	}
	c.perms = strings.Split(*perm, ";")
}

// TestPermission mirrors Command::testPermission(): checks the permission and, if denied, sends
// the configured permission-denial message.
//
// PHP's version formats the default denial message via KnownTranslationFactory — the generated
// translation-factory package isn't ported yet (see the lang package's doc comment), so this
// falls back to a plain English message until that's wired up.
func (c *Command) TestPermission(target Sender, perm *string) bool {
	if c.TestPermissionSilent(target, perm) {
		return true
	}

	if c.permissionMessage == nil {
		target.SendMessage(fmt.Sprintf("You do not have permission to use this command (%s)", c.name))
	} else if *c.permissionMessage != "" {
		permStr := ""
		if perm != nil {
			permStr = *perm
		} else {
			permStr = strings.Join(c.perms, ";")
		}
		target.SendMessage(strings.ReplaceAll(*c.permissionMessage, "<permission>", permStr))
	}

	return false
}

func (c *Command) TestPermissionSilent(target Sender, perm *string) bool {
	list := c.perms
	if perm != nil {
		list = []string{*perm}
	}
	for _, p := range list {
		if target.HasPermission(p) {
			return true
		}
	}
	return false
}

func (c *Command) Label() string { return c.label }

// SetLabel mirrors Command::setLabel(): only takes effect immediately if the command isn't
// currently registered; otherwise it's queued as nextLabel for when it's next unregistered.
func (c *Command) SetLabel(name string) bool {
	c.nextLabel = name
	if !c.IsRegistered() {
		c.label = name
		return true
	}
	return false
}

func (c *Command) Register(m CommandMap) bool {
	if c.allowChangesFrom(m) {
		c.commandMap = m
		return true
	}
	return false
}

func (c *Command) Unregister(m CommandMap) bool {
	if c.allowChangesFrom(m) {
		c.commandMap = nil
		c.activeAliases = c.aliases
		c.label = c.nextLabel
		return true
	}
	return false
}

func (c *Command) allowChangesFrom(m CommandMap) bool {
	return c.commandMap == nil || c.commandMap == m
}

func (c *Command) IsRegistered() bool { return c.commandMap != nil }

func (c *Command) Aliases() []string { return c.activeAliases }

func (c *Command) PermissionMessage() *string { return c.permissionMessage }

func (c *Command) Description() any { return c.description }

func (c *Command) Usage() any { return c.usageMessage }

func (c *Command) SetAliases(aliases []string) {
	c.aliases = aliases
	if !c.IsRegistered() {
		c.activeAliases = aliases
	}
}

func (c *Command) SetDescription(description any) { c.description = description }

func (c *Command) SetPermissionMessage(permissionMessage string) {
	c.permissionMessage = &permissionMessage
}

func (c *Command) SetUsage(usage any) { c.usageMessage = usage }

// BroadcastCommandMessage is a port of Command::broadcastCommandMessage().
//
// PHP formats the broadcast via KnownTranslationFactory::chat_type_admin(); deferred the same way
// as TestPermission's denial message, pending the generated translation factory.
func BroadcastCommandMessage(source Sender, message any, sendToSource bool) {
	users := source.GetServer().GetBroadcastChannelSubscribers(BroadcastChannelAdministrative)

	messageStr := stringifyMessage(message)
	adminBroadcast := fmt.Sprintf("[%s: %s]", source.GetName(), messageStr)

	if sendToSource {
		source.SendMessage(message)
	}

	for _, user := range users {
		if user != source {
			user.SendMessage(adminBroadcast)
		}
	}
}

func stringifyMessage(message any) string {
	if t, ok := message.(*lang.Translatable); ok {
		return t.Text()
	}
	if s, ok := message.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", message)
}

func (c *Command) String() string { return c.name }
