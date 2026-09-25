package defaults

import (
	"pocketmine-go/pocketmine/command"
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/permission"
)

// KillCommand is a port of pocketmine\command\defaults\KillCommand.
type KillCommand struct{ VanillaCommand }

func NewKillCommand() *KillCommand {
	c := &KillCommand{VanillaCommand{Command: command.InitCommand("kill", lang.KnownTranslationFactory.PocketmineCommandKillDescription(), lang.KnownTranslationFactory.PocketmineCommandKillUsage(), []string{"suicide"})}}
	_ = c.SetPermissions([]string{permission.CommandKillSelf, permission.CommandKillOther})
	return c
}

func (c *KillCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	if len(args) >= 2 {
		return syntaxError()
	}
	var targetName *string
	if len(args) > 0 {
		targetName = &args[0]
	}
	p, err := c.fetchPermittedPlayerTarget(sender, targetName, permission.CommandKillSelf, permission.CommandKillOther)
	if err != nil || p == nil {
		return true, err
	}

	p.Attack(entityevent.NewEntityDamageEvent(p, entityevent.CauseSuicide, p.GetHealth(), nil))
	if command.Sender(p) == sender {
		sender.SendMessage(lang.KnownTranslationFactory.CommandsKillSuccessful(sender.GetName()))
	} else {
		command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.CommandsKillSuccessful(p.GetName()), true)
	}
	return true, nil
}
