package command

// ClosureExecute is the handler signature for a ClosureCommand.
type ClosureExecute func(sender Sender, cmd CommandLike, commandLabel string, args []string) (any, error)

// ClosureCommand is a port of pocketmine\command\ClosureCommand.
//
// PHP validates the closure's signature at construction time via Utils::validateCallableSignature
// (reflection over the closure's declared parameter types). Go's compiler already enforces
// ClosureExecute's exact signature on whatever function is passed in, so that runtime check has
// nothing left to do here.
type ClosureCommand struct {
	Command
	execute ClosureExecute
}

func NewClosureCommand(name string, execute ClosureExecute, perms []string, description any, usageMessage any, aliases []string) (*ClosureCommand, error) {
	c := &ClosureCommand{Command: InitCommand(name, description, usageMessage, aliases), execute: execute}
	if err := c.SetPermissions(perms); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *ClosureCommand) Execute(sender Sender, commandLabel string, args []string) (any, error) {
	return c.execute(sender, c, commandLabel, args)
}
