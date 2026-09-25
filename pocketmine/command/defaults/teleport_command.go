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

// TeleportCommand is a port of pocketmine\command\defaults\TeleportCommand.
type TeleportCommand struct{ VanillaCommand }

func NewTeleportCommand() *TeleportCommand {
	c := &TeleportCommand{VanillaCommand{Command: command.InitCommand("tp", lang.KnownTranslationFactory.PocketmineCommandTpDescription(), lang.KnownTranslationFactory.CommandsTpUsage(), []string{"teleport"})}}
	_ = c.SetPermissions([]string{permission.CommandTeleportSelf, permission.CommandTeleportOther})
	return c
}

func (c *TeleportCommand) findPlayer(sender command.Sender, playerName string) *player.Player {
	subject := srv(sender).GetPlayerByPrefix(playerName)
	if subject == nil {
		sender.SendMessage(lang.KnownTranslationFactory.PocketmineCommandErrorPlayerNotFound(playerName).Prefix(utils.Red))
	}
	return subject
}

func (c *TeleportCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	var subjectName *string
	switch len(args) {
	case 1, 3, 5: // /tp targetPlayer, /tp x y z, /tp x y z yaw pitch - TODO: 5 args could be target x y z yaw :(
		subjectName = nil //self
	case 2, 4, 6: // /tp player1 player2, /tp player1 x y z, /tp player1 x y z yaw pitch
		name := args[0]
		subjectName = &name
		args = args[1:]
	default:
		return syntaxError()
	}

	subject, err := c.fetchPermittedPlayerTarget(sender, subjectName, permission.CommandTeleportSelf, permission.CommandTeleportOther)
	if err != nil || subject == nil {
		return true, err
	}

	switch len(args) {
	case 1:
		targetPlayer := c.findPlayer(sender, args[0])
		if targetPlayer == nil {
			return true, nil
		}
		location := targetPlayer.GetLocation()
		subject.TeleportTo(location.Vector3, location.World, &location.Yaw, &location.Pitch)
		command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.CommandsTpSuccess(subject.GetName(), targetPlayer.GetName()), true)
		return true, nil
	case 3, 5:
		base := subject.GetLocation()
		yaw, pitch := base.Yaw, base.Pitch
		if len(args) == 5 {
			yaw, pitch = phpFloat(args[3]), phpFloat(args[4])
		}
		x := getRelativeDouble(base.X, args[0], MinCoord, MaxCoord)
		y := getRelativeDouble(base.Y, args[1], world.YMin, world.YMax)
		z := getRelativeDouble(base.Z, args[2], MinCoord, MaxCoord)

		subject.TeleportTo(math.NewVector3(x, y, z), base.World, &yaw, &pitch)
		command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.CommandsTpSuccessCoordinates(subject.GetName(), strval(round2(x)), strval(round2(y)), strval(round2(z))), true)
		return true, nil
	}
	panic("This branch should be unreachable (for now)")
}
