package defaults

import (
	"strconv"

	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/permission"
	"pocketmine-go/pocketmine/player"
	"pocketmine-go/pocketmine/world"
)

// TimeCommand is a port of pocketmine\command\defaults\TimeCommand.
type TimeCommand struct{ VanillaCommand }

func NewTimeCommand() *TimeCommand {
	c := &TimeCommand{VanillaCommand{Command: command.InitCommand("time", lang.KnownTranslationFactory.PocketmineCommandTimeDescription(), lang.KnownTranslationFactory.PocketmineCommandTimeUsage(), nil)}}
	_ = c.SetPermissions([]string{permission.CommandTimeAdd, permission.CommandTimeSet, permission.CommandTimeStart, permission.CommandTimeStop, permission.CommandTimeQuery})
	return c
}

func (c *TimeCommand) test(sender command.Sender, perm string) bool {
	return c.TestPermission(sender, &perm)
}

func (c *TimeCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	if len(args) < 1 {
		return syntaxError()
	}
	worlds := srv(sender).GetWorldManager().GetWorlds()

	switch args[0] {
	case "start":
		if !c.test(sender, permission.CommandTimeStart) {
			return true, nil
		}
		for _, w := range worlds {
			w.StartTime()
		}
		command.BroadcastCommandMessage(sender, "Restarted the time", true)
		return true, nil
	case "stop":
		if !c.test(sender, permission.CommandTimeStop) {
			return true, nil
		}
		for _, w := range worlds {
			w.StopTime()
		}
		command.BroadcastCommandMessage(sender, "Stopped the time", true)
		return true, nil
	case "query":
		if !c.test(sender, permission.CommandTimeQuery) {
			return true, nil
		}
		w := srv(sender).GetWorldManager().GetDefaultWorld()
		if p, ok := sender.(*player.Player); ok {
			w = p.GetWorld()
		}
		sender.SendMessage(sender.GetLanguage().Translate(lang.KnownTranslationFactory.CommandsTimeQuery(strconv.FormatInt(w.GetTime(), 10))))
		return true, nil
	}

	if len(args) < 2 {
		return syntaxError()
	}

	switch args[0] {
	case "set":
		if !c.test(sender, permission.CommandTimeSet) {
			return true, nil
		}
		var value int64
		switch args[1] {
		case "day":
			value = world.TimeDay
		case "noon":
			value = world.TimeNoon
		case "sunset":
			value = world.TimeSunset
		case "night":
			value = world.TimeNight
		case "midnight":
			value = world.TimeMidnight
		case "sunrise":
			value = world.TimeSunrise
		default:
			value = int64(getInteger(args[1], 0, MaxCoord))
		}
		for _, w := range worlds {
			w.SetTime(value)
		}
		command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.CommandsTimeSet(strconv.FormatInt(value, 10)), true)
	case "add":
		if !c.test(sender, permission.CommandTimeAdd) {
			return true, nil
		}
		value := int64(getInteger(args[1], 0, MaxCoord))
		for _, w := range worlds {
			w.SetTime(w.GetTime() + value)
		}
		command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.CommandsTimeAdded(strconv.FormatInt(value, 10)), true)
	default:
		return syntaxError()
	}
	return true, nil
}
