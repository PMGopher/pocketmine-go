package plugin

import (
	"fmt"
	"slices"
	"strings"

	"github.com/sandertv/gophertunnel/minecraft/protocol"

	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/utils"
)

// PluginLoadabilityChecker is a port of pocketmine\plugin\PluginLoadabilityChecker.
type PluginLoadabilityChecker struct {
	apiVersion string
}

func NewPluginLoadabilityChecker(apiVersion string) *PluginLoadabilityChecker {
	return &PluginLoadabilityChecker{apiVersion: apiVersion}
}

// Check is a port of PluginLoadabilityChecker::check: why the plugin can't be loaded, or nil.
func (c *PluginLoadabilityChecker) Check(description *Description) *lang.Translatable {
	name := strings.ToLower(description.GetName())
	if strings.Contains(name, "pocketmine") || strings.Contains(name, "minecraft") || strings.Contains(name, "mojang") {
		return lang.KnownTranslationFactory.PocketminePluginRestrictedName()
	}

	for _, api := range description.GetCompatibleApis() {
		if !utils.IsValidBaseVersion(api) {
			return lang.KnownTranslationFactory.PocketminePluginInvalidAPI(api)
		}
	}

	if !IsCompatible(c.apiVersion, description.GetCompatibleApis()) {
		return lang.KnownTranslationFactory.PocketminePluginIncompatibleAPI(strings.Join(description.GetCompatibleApis(), ", "))
	}

	if ambiguousVersions := CheckAmbiguousVersions(description.GetCompatibleApis()); len(ambiguousVersions) > 0 {
		return lang.KnownTranslationFactory.PocketminePluginAmbiguousMinAPI(strings.Join(ambiguousVersions, ", "))
	}

	if oses := description.GetCompatibleOperatingSystems(); len(oses) > 0 && !slices.Contains(oses, utils.GetOS()) {
		return lang.KnownTranslationFactory.PocketminePluginIncompatibleOS(strings.Join(oses, ", "))
	}

	if pluginMcpeProtocols := description.GetCompatibleMcpeProtocols(); len(pluginMcpeProtocols) > 0 {
		if !slices.Contains(pluginMcpeProtocols, protocol.CurrentProtocol) {
			strs := make([]string, len(pluginMcpeProtocols))
			for i, p := range pluginMcpeProtocols {
				strs[i] = fmt.Sprint(p)
			}
			return lang.KnownTranslationFactory.PocketminePluginIncompatibleProtocol(strings.Join(strs, ", "))
		}
	}

	// PHP checks that each required PHP extension is loaded (and its version). A Go server has no
	// PHP extensions, so any plugin requiring one can't be loaded.
	for extensionName := range description.GetRequiredExtensions() {
		return lang.KnownTranslationFactory.PocketminePluginExtensionNotLoaded(extensionName)
	}

	return nil
}
