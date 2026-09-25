package defaults

import (
	"strconv"
	"time"

	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/permission"
)

// SaveCommand is a port of pocketmine\command\defaults\SaveCommand.
type SaveCommand struct{ VanillaCommand }

func NewSaveCommand() *SaveCommand {
	return &SaveCommand{newVanillaCommand("save-all", lang.KnownTranslationFactory.PocketmineCommandSaveDescription(), nil, nil, permission.CommandSavePerform)}
}

func (c *SaveCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.PocketmineSaveStart(), true)
	start := time.Now()

	for _, p := range srv(sender).GetOnlinePlayers() {
		p.Save()
	}
	worldManager := srv(sender).GetWorldManager()
	for _, w := range worldManager.GetWorlds() {
		if err := worldManager.SaveWorld(w); err != nil {
			srv(sender).GetLogger().Error(err.Error())
		}
	}

	command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.PocketmineSaveSuccess(strconv.FormatFloat(float64(time.Since(start).Milliseconds())/1000, 'f', -1, 64)), true)
	return true, nil
}

// SaveOffCommand is a port of pocketmine\command\defaults\SaveOffCommand.
type SaveOffCommand struct{ VanillaCommand }

func NewSaveOffCommand() *SaveOffCommand {
	return &SaveOffCommand{newVanillaCommand("save-off", lang.KnownTranslationFactory.PocketmineCommandSaveoffDescription(), nil, nil, permission.CommandSaveDisable)}
}

func (c *SaveOffCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	srv(sender).GetWorldManager().SetAutoSave(false)
	command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.CommandsSaveDisabled(), true)
	return true, nil
}

// SaveOnCommand is a port of pocketmine\command\defaults\SaveOnCommand.
type SaveOnCommand struct{ VanillaCommand }

func NewSaveOnCommand() *SaveOnCommand {
	return &SaveOnCommand{newVanillaCommand("save-on", lang.KnownTranslationFactory.PocketmineCommandSaveonDescription(), nil, nil, permission.CommandSaveEnable)}
}

func (c *SaveOnCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	srv(sender).GetWorldManager().SetAutoSave(true)
	command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.CommandsSaveEnabled(), true)
	return true, nil
}
