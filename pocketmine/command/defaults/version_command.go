package defaults

import (
	"runtime"
	"strconv"

	"github.com/sandertv/gophertunnel/minecraft/protocol"

	"pocketmine-go/pocketmine"
	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/permission"
	"pocketmine-go/pocketmine/utils"
)

// VersionCommand is a port of pocketmine\command\defaults\VersionCommand. The PHP version and JIT
// lines show the Go runtime version instead, and there are no plugins to describe.
type VersionCommand struct{ VanillaCommand }

func NewVersionCommand() *VersionCommand {
	return &VersionCommand{newVanillaCommand("version", lang.KnownTranslationFactory.PocketmineCommandVersionDescription(), lang.KnownTranslationFactory.PocketmineCommandVersionUsage(), []string{"ver", "about"}, permission.CommandVersion)}
}

func (c *VersionCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	if len(args) == 0 {
		sender.SendMessage(lang.KnownTranslationFactory.PocketmineCommandVersionServerSoftwareName(utils.Green + pocketmine.Name + utils.Reset))
		versionColor := utils.Green
		if pocketmine.IsDevelopmentBuild {
			versionColor = utils.Yellow
		}
		sender.SendMessage(lang.KnownTranslationFactory.PocketmineCommandVersionServerSoftwareVersion(
			versionColor+pocketmine.Version().GetFullVersion(false)+utils.Reset,
			utils.Green+pocketmine.GitHash()+utils.Reset,
		))
		sender.SendMessage(lang.KnownTranslationFactory.PocketmineCommandVersionMinecraftVersion(
			utils.Green+protocol.CurrentVersion+utils.Reset,
			utils.Green+strconv.Itoa(protocol.CurrentProtocol)+utils.Reset,
		))
		sender.SendMessage(lang.KnownTranslationFactory.PocketmineCommandVersionPhpVersion(utils.Green + runtime.Version() + utils.Reset))
		sender.SendMessage(lang.KnownTranslationFactory.PocketmineCommandVersionPhpJitStatus(lang.KnownTranslationFactory.PocketmineCommandVersionPhpJitNotSupported().Format(utils.Green, utils.Reset)))
		sender.SendMessage(lang.KnownTranslationFactory.PocketmineCommandVersionOperatingSystem(utils.Green + utils.GetOS() + utils.Reset))
		return true, nil
	}
	// PluginManager::getPlugin: plugins aren't ported, so no plugin matches.
	sender.SendMessage(lang.KnownTranslationFactory.PocketmineCommandVersionNoSuchPlugin())
	return true, nil
}
