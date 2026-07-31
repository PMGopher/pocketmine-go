package command

import "pocketmine-go/pocketmine/permission"

// PluginCommand is a port of pocketmine\command\PluginCommand: routes commands defined in
// plugin.yml to a plugin's Executor.OnCommand.
type PluginCommand struct {
	Command
	owner    permission.Plugin
	executor Executor
}

func NewPluginCommand(name string, owner permission.Plugin, executor Executor) *PluginCommand {
	c := &PluginCommand{Command: InitCommand(name, "", nil, nil), owner: owner, executor: executor}
	c.Command.usageMessage = ""
	return c
}

func (c *PluginCommand) Execute(sender Sender, commandLabel string, args []string) (any, error) {
	if !c.owner.IsEnabled() {
		return false, nil
	}

	success := c.executor.OnCommand(sender, c, commandLabel, args)
	if !success {
		if usage, ok := c.Command.usageMessage.(string); ok && usage != "" {
			return nil, &InvalidCommandSyntaxException{}
		}
	}

	return success, nil
}

func (c *PluginCommand) GetOwningPlugin() permission.Plugin { return c.owner }
func (c *PluginCommand) GetExecutor() Executor              { return c.executor }
func (c *PluginCommand) SetExecutor(executor Executor)      { c.executor = executor }

var _ PluginOwned = (*PluginCommand)(nil)
