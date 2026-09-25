// Package defaults is a port of pocketmine\command\defaults: the commands every server has.
package defaults

import (
	"strconv"
	"strings"

	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/player"
	"pocketmine-go/pocketmine/server"
	"pocketmine-go/pocketmine/utils"
)

func init() {
	server.RegisterDefaultCommandsFunc = SetDefaultCommands
}

// SetDefaultCommands is a port of SimpleCommandMap::setDefaultCommands.
//
// Not ported yet: give, clear and enchant/effect's item/effect name lookups where they need
// StringToItemParser (not ported), particle, timings (report upload) and dumpmemory (PHP-specific).
func SetDefaultCommands(m *command.SimpleCommandMap) {
	m.RegisterAll("pocketmine", []command.CommandLike{
		NewBanCommand(),
		NewBanIpCommand(),
		NewBanListCommand(),
		NewDefaultGamemodeCommand(),
		NewDeopCommand(),
		NewDifficultyCommand(),
		NewGamemodeCommand(),
		NewGarbageCollectorCommand(),
		NewHelpCommand(),
		NewKickCommand(),
		NewKillCommand(),
		NewListCommand(),
		NewMeCommand(),
		NewOpCommand(),
		NewPardonCommand(),
		NewPardonIpCommand(),
		NewPluginsCommand(),
		NewSaveCommand(),
		NewSaveOffCommand(),
		NewSaveOnCommand(),
		NewSayCommand(),
		NewSeedCommand(),
		NewSetWorldSpawnCommand(),
		NewSpawnpointCommand(),
		NewStatusCommand(),
		NewStopCommand(),
		NewTeleportCommand(),
		NewTellCommand(),
		NewTimeCommand(),
		NewTitleCommand(),
		NewTransferServerCommand(),
		NewVersionCommand(),
		NewWhitelistCommand(),
		NewXpCommand(),
	})
}

// VanillaCommand coordinate limits, VanillaCommand::MAX_COORD/MIN_COORD.
const (
	MaxCoord = 30000000
	MinCoord = -30000000
)

// VanillaCommand is a port of pocketmine\command\defaults\VanillaCommand.
type VanillaCommand struct {
	command.Command
}

func newVanillaCommand(name string, description, usage any, aliases []string, permission string) VanillaCommand {
	c := VanillaCommand{Command: command.InitCommand(name, description, usage, aliases)}
	c.SetPermission(&permission)
	return c
}

// srv is $sender->getServer() as the concrete server.
func srv(sender command.Sender) *server.Server {
	return sender.GetServer().(*server.Server)
}

// syntaxError is `throw new InvalidCommandSyntaxException()`.
func syntaxError() (any, error) { return nil, &command.InvalidCommandSyntaxException{} }

// senderName is `$sender instanceof Player ? $sender->getDisplayName() : $sender->getName()`.
func senderName(sender command.Sender) string {
	if p, ok := sender.(*player.Player); ok {
		return p.GetDisplayName()
	}
	return sender.GetName()
}

// fetchPermittedPlayerTarget is a port of VanillaCommand::fetchPermittedPlayerTarget.
func (c *VanillaCommand) fetchPermittedPlayerTarget(sender command.Sender, target *string, selfPermission, otherPermission string) (*player.Player, error) {
	var p *player.Player
	//TODO: we need proper command selector support, but this one is useful and easy to hack in for now
	if target != nil && *target != "@s" {
		p = srv(sender).GetPlayerByPrefix(*target)
		if p == nil {
			sender.SendMessage(lang.KnownTranslationFactory.PocketmineCommandErrorPlayerNotFound(*target).Prefix(utils.Red))
			return nil, nil
		}
	} else if self, ok := sender.(*player.Player); ok {
		p = self
	} else {
		return nil, &command.InvalidCommandSyntaxException{}
	}

	isSelf := command.Sender(p) == sender
	if (isSelf && c.TestPermission(sender, &selfPermission)) || (!isSelf && c.TestPermission(sender, &otherPermission)) {
		return p, nil
	}
	return nil, nil
}

// phpInt is PHP's (int) cast of a string: the leading integer, or 0.
func phpInt(value string) int {
	value = strings.TrimSpace(value)
	end := 0
	for end < len(value) && (value[end] >= '0' && value[end] <= '9' || end == 0 && (value[end] == '-' || value[end] == '+')) {
		end++
	}
	i, _ := strconv.Atoi(value[:end])
	return i
}

// phpFloat is PHP's (float) cast of a string: the leading number, or 0.
func phpFloat(value string) float64 {
	value = strings.TrimSpace(value)
	for end := len(value); end > 0; end-- {
		if f, err := strconv.ParseFloat(value[:end], 64); err == nil {
			return f
		}
	}
	return 0
}

// isNumeric is PHP's is_numeric.
func isNumeric(value string) bool {
	_, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	return err == nil && strings.TrimRight(value, " \t\n\r\v\f") == value
}

// getInteger is a port of VanillaCommand::getInteger.
func getInteger(value string, minV, maxV int) int {
	return max(minV, min(maxV, phpInt(value)))
}

// getRelativeDouble is a port of VanillaCommand::getRelativeDouble.
func getRelativeDouble(original float64, input string, minV, maxV float64) float64 {
	if strings.HasPrefix(input, "~") {
		return original + getDouble(input[1:], MinCoord, MaxCoord)
	}
	return getDouble(input, minV, maxV)
}

// getDouble is a port of VanillaCommand::getDouble.
func getDouble(value string, minV, maxV float64) float64 {
	return max(minV, min(maxV, phpFloat(value)))
}

// getBoundedInt is a port of VanillaCommand::getBoundedInt: nil (after telling the sender) when
// out of bounds.
func getBoundedInt(sender command.Sender, input string, minV, maxV int) (*int, error) {
	if !isNumeric(input) {
		return nil, &command.InvalidCommandSyntaxException{}
	}
	v := phpInt(input)
	if v > maxV {
		sender.SendMessage(lang.KnownTranslationFactory.CommandsGenericNumTooBig(input, strconv.Itoa(maxV)).Prefix(utils.Red))
		return nil, nil
	}
	if v < minV {
		sender.SendMessage(lang.KnownTranslationFactory.CommandsGenericNumTooSmall(input, strconv.Itoa(minV)).Prefix(utils.Red))
		return nil, nil
	}
	return &v, nil
}
