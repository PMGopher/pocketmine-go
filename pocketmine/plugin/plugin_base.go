package plugin

import (
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/scheduler"
	"pocketmine-go/pocketmine/utils"
)

// PluginBase is a port of pocketmine\plugin\PluginBase. A plugin's main type embeds it and
// overrides what it needs:
//
//	type Main struct{ plugin.PluginBase }
//
//	func (m *Main) OnEnable() error { m.GetLogger().Info("Hello"); return nil }
//
// PHP's protected onLoad/onEnable/onDisable are the optional OnLoad(), OnEnable() error and
// OnDisable() methods on the plugin type; OnCommand is overridden the same way.
type PluginBase struct {
	self Plugin

	isEnabled bool

	loader           PluginLoader
	server           Server
	description      *Description
	dataFolder       string
	file             string
	resourceProvider ResourceProvider

	resourceFolder string

	config     *utils.Config
	configFile string

	logger    *PluginLogger
	scheduler *scheduler.TaskScheduler
}

type (
	onLoader   interface{ OnLoad() }
	onEnabler  interface{ OnEnable() error }
	onDisabler interface{ OnDisable() }
)

// InitPlugin is a port of PluginBase::__construct.
func (b *PluginBase) InitPlugin(self Plugin, loader PluginLoader, server Server, description *Description, dataFolder string, file string, resourceProvider ResourceProvider) {
	b.self = self
	b.loader = loader
	b.server = server
	b.description = description
	b.resourceProvider = resourceProvider

	b.dataFolder = strings.TrimRight(dataFolder, "/"+string(filepath.Separator)) + "/"
	b.file = strings.TrimRight(file, "/"+string(filepath.Separator)) + "/"
	b.resourceFolder = path.Join(b.file, "resources") + "/"

	b.configFile = filepath.Join(b.dataFolder, "config.toml")

	prefix := b.description.GetPrefix()
	if prefix == "" {
		prefix = b.GetName()
	}
	b.logger = NewPluginLogger(server.GetLogger(), prefix)
	b.scheduler = scheduler.NewTaskScheduler(b.GetFullName())

	if l, ok := self.(onLoader); ok {
		l.OnLoad()
	}

	b.registerYamlCommands()
}

// IsEnabled is a port of PluginBase::isEnabled.
func (b *PluginBase) IsEnabled() bool { return b.isEnabled }

// OnEnableStateChange is a port of PluginBase::onEnableStateChange.
func (b *PluginBase) OnEnableStateChange(enabled bool) error {
	if b.isEnabled != enabled {
		b.isEnabled = enabled
		if b.isEnabled {
			if e, ok := b.self.(onEnabler); ok {
				return e.OnEnable()
			}
		} else if d, ok := b.self.(onDisabler); ok {
			d.OnDisable()
		}
	}
	return nil
}

// IsDisabled is a port of PluginBase::isDisabled.
func (b *PluginBase) IsDisabled() bool { return !b.isEnabled }

// GetDataFolder is a port of PluginBase::getDataFolder.
func (b *PluginBase) GetDataFolder() string { return b.dataFolder }

// GetDescription is a port of PluginBase::getDescription.
func (b *PluginBase) GetDescription() *Description { return b.description }

// GetLogger is a port of PluginBase::getLogger.
func (b *PluginBase) GetLogger() log.AttachableLogger { return b.logger }

// registerYamlCommands is a port of PluginBase::registerYamlCommands: registers commands declared
// in the plugin manifest.
func (b *PluginBase) registerYamlCommands() {
	var pluginCmds []command.CommandLike

	executor, _ := b.self.(command.Executor)
	if executor == nil {
		executor = b
	}

	for _, key := range sortedKeys(b.description.GetCommands()) {
		data := b.description.GetCommands()[key]
		if strings.Contains(key, ":") {
			b.logger.Error(b.server.GetLanguage().Translate(lang.KnownTranslationFactory.PocketminePluginCommandError(key, b.description.GetFullName(), ":")))
			continue
		}

		newCmd := command.NewPluginCommand(key, b.self, executor)
		if data.Description != nil {
			newCmd.SetDescription(*data.Description)
		}

		if data.UsageMessage != nil {
			newCmd.SetUsage(*data.UsageMessage)
		}

		var aliasList []string
		for _, alias := range data.Aliases {
			if strings.Contains(alias, ":") {
				b.logger.Error(b.server.GetLanguage().Translate(lang.KnownTranslationFactory.PocketminePluginAliasError(alias, b.description.GetFullName(), ":")))
				continue
			}
			aliasList = append(aliasList, alias)
		}

		newCmd.SetAliases(aliasList)

		permission := data.Permission
		newCmd.SetPermission(&permission)

		if data.PermissionDeniedMessage != nil {
			newCmd.SetPermissionMessage(*data.PermissionDeniedMessage)
		}

		pluginCmds = append(pluginCmds, newCmd)
	}

	if len(pluginCmds) > 0 {
		b.server.GetSimpleCommandMap().RegisterAll(b.description.GetName(), pluginCmds)
	}
}

// getPluginCommand is Server::getPluginCommand.
func getPluginCommand(m *command.SimpleCommandMap, name string) command.PluginOwned {
	if c, ok := m.GetCommand(name).(command.PluginOwned); ok {
		return c
	}
	return nil
}

// GetCommand is a port of PluginBase::getCommand: the plugin's command by name (nil if none).
func (b *PluginBase) GetCommand(name string) *command.PluginCommand {
	commandMap := b.server.GetSimpleCommandMap()
	cmd := getPluginCommand(commandMap, name)
	if cmd == nil || cmd.GetOwningPlugin() != b.self {
		cmd = getPluginCommand(commandMap, strings.ToLower(b.description.GetName())+":"+name)
	}

	if pc, ok := cmd.(*command.PluginCommand); ok && pc.GetOwningPlugin() == b.self {
		return pc
	}
	return nil
}

// OnCommand is a port of PluginBase::onCommand (a plugin overrides it to handle its commands).
func (b *PluginBase) OnCommand(sender command.Sender, cmd command.CommandLike, label string, args []string) bool {
	return false
}

// GetResourceFolder is a port of PluginBase::getResourceFolder: where the plugin's embedded
// resource files are located. For a compiled-in Go plugin this is a go-plugin:// path; read the
// files with GetResource.
func (b *PluginBase) GetResourceFolder() string { return b.resourceFolder }

// GetResourcePath is a port of PluginBase::getResourcePath.
func (b *PluginBase) GetResourcePath(filename string) string {
	return path.Join(b.GetResourceFolder(), filename)
}

// GetResource is a port of PluginBase::getResource: an embedded resource (nil if there's none).
// The caller must close it.
func (b *PluginBase) GetResource(filename string) io.ReadCloser {
	return b.resourceProvider.GetResource(filename)
}

// SaveResource is a port of PluginBase::saveResource: saves an embedded resource to its relative
// location in the data folder. PHP copies it from getResourceFolder(); here it's read through the
// resource provider, so it also works for a Go plugin's embedded files.
func (b *PluginBase) SaveResource(filename string, replace bool) bool {
	if strings.TrimSpace(filename) == "" {
		return false
	}

	source := b.resourceProvider.GetResource(filename)
	if source == nil {
		return false
	}
	defer source.Close()

	destination := filepath.Join(b.dataFolder, filename)
	if _, err := os.Stat(destination); err == nil && !replace {
		return false
	}

	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return false
	}

	out, err := os.Create(destination)
	if err != nil {
		return false
	}
	defer out.Close()
	_, err = io.Copy(out, source)
	return err == nil
}

// GetResources is a port of PluginBase::getResources.
func (b *PluginBase) GetResources() map[string]os.FileInfo { return b.resourceProvider.GetResources() }

// GetConfig is a port of PluginBase::getConfig.
func (b *PluginBase) GetConfig() *utils.Config {
	if b.config == nil {
		b.ReloadConfig()
	}
	return b.config
}

// SaveConfig is a port of PluginBase::saveConfig.
func (b *PluginBase) SaveConfig() error { return b.GetConfig().Save() }

// SaveDefaultConfig is a port of PluginBase::saveDefaultConfig: the plugin's config is
// config.toml in its data folder, copied from resources/config.toml. A config.yml there (from an
// older version of the plugin or of this server) is converted to config.toml instead, and so is a
// plugin that still ships resources/config.yml.
func (b *PluginBase) SaveDefaultConfig() bool {
	if _, err := os.Stat(b.configFile); !errors.Is(err, os.ErrNotExist) {
		return false
	}
	yamlFile := filepath.Join(b.dataFolder, "config.yml")
	if _, err := os.Stat(yamlFile); errors.Is(err, os.ErrNotExist) {
		if b.SaveResource("config.toml", false) {
			return true
		}
		if !b.SaveResource("config.yml", false) {
			return false
		}
	}
	converted, err := utils.ConvertYAMLToTOML(yamlFile, b.configFile, "Converted from config.yml (kept as config.yml.bak).")
	if err != nil {
		b.logger.Error("Failed to convert config.yml to config.toml: " + err.Error())
		return false
	}
	if converted {
		b.logger.Notice("Converted config.yml to config.toml (the old file is kept as config.yml.bak)")
	}
	return converted
}

// ReloadConfig is a port of PluginBase::reloadConfig.
func (b *PluginBase) ReloadConfig() {
	b.SaveDefaultConfig()
	config, err := utils.NewConfig(b.configFile, utils.ConfigTOML, nil)
	if err != nil {
		b.logger.Error("Failed to load config.toml: " + err.Error())
		config, _ = utils.NewConfig(b.configFile, utils.ConfigTOML, map[string]any{})
	}
	b.config = config
}

// GetServer is a port of PluginBase::getServer (type-assert it to *server.Server for the rest of
// the server API).
func (b *PluginBase) GetServer() Server { return b.server }

// GetName is a port of PluginBase::getName.
func (b *PluginBase) GetName() string { return b.description.GetName() }

// Name is GetName, for permission.Plugin.
func (b *PluginBase) Name() string { return b.description.GetName() }

// GetFullName is a port of PluginBase::getFullName.
func (b *PluginBase) GetFullName() string { return b.description.GetFullName() }

// GetFile is a port of PluginBase::getFile.
func (b *PluginBase) GetFile() string { return b.file }

// GetPluginLoader is a port of PluginBase::getPluginLoader.
func (b *PluginBase) GetPluginLoader() PluginLoader { return b.loader }

// GetScheduler is a port of PluginBase::getScheduler.
func (b *PluginBase) GetScheduler() *scheduler.TaskScheduler { return b.scheduler }

// sortedKeys is the keys of m in a stable order (the manifest order isn't kept by Description).
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
