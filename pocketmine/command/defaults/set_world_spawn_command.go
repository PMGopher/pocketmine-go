package defaults

import (
	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/permission"
	"pocketmine-go/pocketmine/player"
	"pocketmine-go/pocketmine/utils"
	"pocketmine-go/pocketmine/world"
)

// SetWorldSpawnCommand is a port of pocketmine\command\defaults\SetWorldSpawnCommand.
type SetWorldSpawnCommand struct{ VanillaCommand }

func NewSetWorldSpawnCommand() *SetWorldSpawnCommand {
	return &SetWorldSpawnCommand{newVanillaCommand("setworldspawn", lang.KnownTranslationFactory.PocketmineCommandSetworldspawnDescription(), lang.KnownTranslationFactory.CommandsSetworldspawnUsage(), nil, permission.CommandSetWorldSpawn)}
}

func (c *SetWorldSpawnCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	var w *world.World
	var pos math.Vector3
	p, isPlayer := sender.(*player.Player)
	switch len(args) {
	case 0:
		if !isPlayer {
			sender.SendMessage(lang.KnownTranslationFactory.PocketmineCommandErrorPlayerUserOnly().Prefix(utils.Red))
			return true, nil
		}
		w = p.GetWorld()
		pos = p.GetPosition().Floor()
	case 3:
		base := math.NewVector3(0, 0, 0)
		w = srv(sender).GetWorldManager().GetDefaultWorld()
		if isPlayer {
			base = p.GetPosition()
			w = p.GetWorld()
		}
		pos = math.NewVector3(
			getRelativeDouble(base.X, args[0], MinCoord, MaxCoord),
			getRelativeDouble(base.Y, args[1], world.YMin, world.YMax),
			getRelativeDouble(base.Z, args[2], MinCoord, MaxCoord),
		).Floor()
	default:
		return syntaxError()
	}

	w.SetSpawnLocation(pos)
	command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.CommandsSetworldspawnSuccess(strval(pos.X), strval(pos.Y), strval(pos.Z)), true)
	return true, nil
}

// SpawnpointCommand is a port of pocketmine\command\defaults\SpawnpointCommand.
type SpawnpointCommand struct{ VanillaCommand }

func NewSpawnpointCommand() *SpawnpointCommand {
	c := &SpawnpointCommand{VanillaCommand{Command: command.InitCommand("spawnpoint", lang.KnownTranslationFactory.PocketmineCommandSpawnpointDescription(), lang.KnownTranslationFactory.CommandsSpawnpointUsage(), nil)}}
	_ = c.SetPermissions([]string{permission.CommandSpawnpointSelf, permission.CommandSpawnpointOther})
	return c
}

func (c *SpawnpointCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	var targetName *string
	if len(args) > 0 {
		targetName = &args[0]
	}
	target, err := c.fetchPermittedPlayerTarget(sender, targetName, permission.CommandSpawnpointSelf, permission.CommandSpawnpointOther)
	if err != nil || target == nil {
		return true, err
	}

	self, isPlayer := sender.(*player.Player)
	if len(args) == 4 {
		w := target.GetWorld()
		pos := w.GetSpawnLocation()
		if isPlayer {
			pos = self.GetPosition()
		}
		x := getRelativeDouble(pos.X, args[1], MinCoord, MaxCoord)
		y := getRelativeDouble(pos.Y, args[2], world.YMin, world.YMax)
		z := getRelativeDouble(pos.Z, args[3], MinCoord, MaxCoord)
		spawn := math.NewVector3(x, y, z)
		target.SetSpawn(&spawn, w)

		command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.PocketmineCommandSpawnpointSuccess(target.GetName(), strval(round2(x)), strval(round2(y)), strval(round2(z))), true)
		return true, nil
	} else if len(args) <= 1 && isPlayer {
		pos := self.GetPosition().Floor()
		target.SetSpawn(&pos, self.GetWorld())

		command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.PocketmineCommandSpawnpointSuccess(target.GetName(), strval(pos.X), strval(pos.Y), strval(pos.Z)), true)
		return true, nil
	}
	return syntaxError()
}
