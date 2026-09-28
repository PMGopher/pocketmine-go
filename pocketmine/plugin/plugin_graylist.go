package plugin

import (
	"fmt"
	"sort"
)

// PluginGraylist is a port of pocketmine\plugin\PluginGraylist (plugin_list.yml).
type PluginGraylist struct {
	plugins     map[string]bool
	order       []string
	isWhitelist bool
}

func NewPluginGraylist(plugins []string, whitelist bool) *PluginGraylist {
	g := &PluginGraylist{plugins: map[string]bool{}, isWhitelist: whitelist}
	for _, p := range plugins {
		if !g.plugins[p] {
			g.plugins[p] = true
			g.order = append(g.order, p)
		}
	}
	return g
}

func (g *PluginGraylist) GetPlugins() []string { return append([]string(nil), g.order...) }

func (g *PluginGraylist) IsWhitelist() bool { return g.isWhitelist }

// IsAllowed returns whether the given name is permitted by this graylist.
func (g *PluginGraylist) IsAllowed(name string) bool { return g.isWhitelist == g.plugins[name] }

// PluginGraylistFromArray is a port of PluginGraylist::fromArray. Errors are PHP's
// InvalidArgumentException.
func PluginGraylistFromArray(array map[string]any) (*PluginGraylist, error) {
	mode, _ := array["mode"].(string)
	if mode != "whitelist" && mode != "blacklist" {
		return nil, fmt.Errorf("\"mode\" must be set")
	}
	isWhitelist := mode == "whitelist"

	var plugins []string
	if raw, ok := array["plugins"]; ok && raw != nil {
		var values []any
		switch v := raw.(type) {
		case []any:
			values = v
		case map[string]any:
			keys := make([]string, 0, len(v))
			for k := range v {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				values = append(values, v[k])
			}
		default:
			return nil, fmt.Errorf("\"plugins\" must be an array")
		}
		for k, v := range values {
			switch s := v.(type) {
			case string:
				plugins = append(plugins, s)
			case int:
				plugins = append(plugins, fmt.Sprint(s))
			case float64:
				plugins = append(plugins, fmt.Sprint(s))
			default:
				return nil, fmt.Errorf("\"plugins\" contains invalid element at position %d", k)
			}
		}
	}
	return NewPluginGraylist(plugins, isWhitelist), nil
}

// ToArray is a port of PluginGraylist::toArray.
func (g *PluginGraylist) ToArray() map[string]any {
	mode := "blacklist"
	if g.isWhitelist {
		mode = "whitelist"
	}
	return map[string]any{"mode": mode, "plugins": g.GetPlugins()}
}
