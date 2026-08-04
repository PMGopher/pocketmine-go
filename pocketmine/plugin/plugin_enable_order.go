package plugin

import "strings"

// EnableOrder is a port of pocketmine\plugin\PluginEnableOrder.
type EnableOrder int

const (
	EnableOrderStartup EnableOrder = iota
	EnableOrderPostworld
)

// enableOrderAliases mirrors PluginEnableOrder::getAliases for every case.
var enableOrderAliases = map[EnableOrder][]string{
	EnableOrderStartup:   {"startup"},
	EnableOrderPostworld: {"postworld"},
}

// GetAliases is a port of PluginEnableOrder::getAliases.
func (o EnableOrder) GetAliases() []string { return enableOrderAliases[o] }

// EnableOrderFromString is a port of PluginEnableOrder::fromString. ok is false if name matches
// no known alias (real PHP returns null in that case).
func EnableOrderFromString(name string) (order EnableOrder, ok bool) {
	lower := strings.ToLower(name)
	for o, aliases := range enableOrderAliases {
		for _, alias := range aliases {
			if alias == lower {
				return o, true
			}
		}
	}
	return 0, false
}
