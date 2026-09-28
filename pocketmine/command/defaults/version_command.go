package defaults

import (
	"runtime"
	"strconv"
	"strings"

	"github.com/sandertv/gophertunnel/minecraft/protocol"

	"pocketmine-go/pocketmine"
	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/permission"
	"pocketmine-go/pocketmine/plugin"
	"pocketmine-go/pocketmine/utils"
)

// VersionCommand is a port of pocketmine\command\defaults\VersionCommand. PHP's "PHP version"
// and "PHP JIT" lines are replaced by a "Go version" line (the server runs on Go, which has no
// JIT).
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
		sender.SendMessage("Go version: " + utils.Green + runtime.Version() + utils.Reset)
		sender.SendMessage(lang.KnownTranslationFactory.PocketmineCommandVersionOperatingSystem(utils.Green + utils.GetOS() + utils.Reset))
		return true, nil
	}
	pluginName := strings.Join(args, " ")
	pluginManager := srv(sender).GetPluginManager()
	if exactPlugin := pluginManager.GetPlugin(pluginName); exactPlugin != nil {
		describeToSender(exactPlugin, sender)
		return true, nil
	}

	found := false
	pluginName = strings.ToLower(pluginName)
	for _, p := range pluginManager.GetPlugins() {
		if strings.Contains(strings.ToLower(p.GetName()), pluginName) {
			describeToSender(p, sender)
			found = true
		}
	}

	if !found {
		sender.SendMessage(lang.KnownTranslationFactory.PocketmineCommandVersionNoSuchPlugin())
	}
	return true, nil
}

// describeToSender is a port of VersionCommand::describeToSender.
func describeToSender(p plugin.Plugin, sender command.Sender) {
	desc := p.GetDescription()
	sender.SendMessage(lang.KnownTranslationFactory.PocketmineCommandVersionPluginHeader(
		utils.DarkGreen+desc.GetName()+utils.Reset,
		utils.DarkGreen+desc.GetVersion()+utils.Reset,
	))

	if desc.GetDescription() != "" {
		sender.SendMessage(desc.GetDescription())
	}

	if desc.GetWebsite() != "" {
		sender.SendMessage(lang.KnownTranslationFactory.PocketmineCommandVersionPluginWebsite(desc.GetWebsite()))
	}

	if authors := desc.GetAuthors(); len(authors) > 0 {
		if len(authors) == 1 {
			sender.SendMessage(lang.KnownTranslationFactory.PocketmineCommandVersionPluginAuthor(strings.Join(authors, ", ")))
		} else {
			sender.SendMessage(lang.KnownTranslationFactory.PocketmineCommandVersionPluginAuthors(strings.Join(authors, ", ")))
		}
	}
}
