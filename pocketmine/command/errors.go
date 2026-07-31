package command

// CommandException is a port of pocketmine\command\utils\CommandException.
type CommandException struct{ Message string }

func (e *CommandException) Error() string { return e.Message }

// InvalidCommandSyntaxException is a port of pocketmine\command\utils\InvalidCommandSyntaxException.
// Thrown by a Command's Execute to signal "show the usage message", mirroring how PHP's version
// carries no message of its own — SimpleCommandMap.Dispatch catches it and prints the command's
// usage instead of the exception's (empty) text.
type InvalidCommandSyntaxException struct{}

func (e *InvalidCommandSyntaxException) Error() string { return "Invalid command syntax" }
