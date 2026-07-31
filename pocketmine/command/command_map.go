package command

// CommandMap is a port of pocketmine\command\CommandMap.
type CommandMap interface {
	RegisterAll(fallbackPrefix string, commands []CommandLike)
	Register(fallbackPrefix string, cmd CommandLike, label string) bool
	Dispatch(sender Sender, cmdLine string) bool
	ClearCommands()
	GetCommand(name string) CommandLike
}
