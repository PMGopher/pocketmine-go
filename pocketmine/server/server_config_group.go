package server

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"pocketmine-go/pocketmine/utils"
)

// ServerConfigGroup is a port of pocketmine\ServerConfigGroup: server.properties plus pocketmine.yml,
// with command-line --key=value overrides taking precedence (PHP's getopt).
//
// pocketmine.yml isn't shipped with this port yet (resources/pocketmine.yml isn't vendored), so
// pocketmineYml may be nil; GetProperty then only sees command-line overrides and the defaults
// passed by the caller.
type ServerConfigGroup struct {
	pocketmineYml    *utils.Config
	serverProperties *utils.Config
	args             []string

	propertyCache map[string]any
}

// NewServerConfigGroup is a port of ServerConfigGroup::__construct. args are the command-line
// arguments searched for --key=value overrides (os.Args[1:] in production).
func NewServerConfigGroup(pocketmineYml, serverProperties *utils.Config, args []string) *ServerConfigGroup {
	return &ServerConfigGroup{pocketmineYml: pocketmineYml, serverProperties: serverProperties, args: args, propertyCache: map[string]any{}}
}

// getopt is PHP's getopt("", ["$variable::"]) for one long option: --variable=value or --variable.
func (g *ServerConfigGroup) getopt(variable string) (string, bool) {
	for _, arg := range g.args {
		if value, ok := strings.CutPrefix(arg, "--"+variable+"="); ok {
			return value, true
		}
		if arg == "--"+variable {
			return "", true
		}
	}
	return "", false
}

// GetProperty is a port of ServerConfigGroup::getProperty (pocketmine.yml, nested key).
func (g *ServerConfigGroup) GetProperty(variable string, defaultValue any) any {
	if _, ok := g.propertyCache[variable]; !ok {
		if v, ok := g.getopt(variable); ok {
			g.propertyCache[variable] = v
		} else if g.pocketmineYml != nil {
			g.propertyCache[variable] = g.pocketmineYml.GetNested(variable, nil)
		} else {
			g.propertyCache[variable] = nil
		}
	}
	if v := g.propertyCache[variable]; v != nil {
		return v
	}
	return defaultValue
}

func (g *ServerConfigGroup) GetPropertyBool(variable string, defaultValue bool) bool {
	return toBool(g.GetProperty(variable, defaultValue))
}

func (g *ServerConfigGroup) GetPropertyInt(variable string, defaultValue int) int {
	return toInt(g.GetProperty(variable, defaultValue))
}

func (g *ServerConfigGroup) GetPropertyString(variable string, defaultValue string) string {
	return fmt.Sprint(g.GetProperty(variable, defaultValue))
}

// GetConfigString is a port of ServerConfigGroup::getConfigString (server.properties).
func (g *ServerConfigGroup) GetConfigString(variable string, defaultValue string) string {
	if v, ok := g.getopt(variable); ok {
		return v
	}
	if g.serverProperties.Exists(variable, false) {
		return fmt.Sprint(g.serverProperties.Get(variable, nil))
	}
	return defaultValue
}

func (g *ServerConfigGroup) SetConfigString(variable, value string) {
	g.serverProperties.Set(variable, value)
}

// GetConfigInt is a port of ServerConfigGroup::getConfigInt.
func (g *ServerConfigGroup) GetConfigInt(variable string, defaultValue int) int {
	if v, ok := g.getopt(variable); ok {
		return toInt(v)
	}
	if g.serverProperties.Exists(variable, false) {
		return toInt(g.serverProperties.Get(variable, nil))
	}
	return defaultValue
}

func (g *ServerConfigGroup) SetConfigInt(variable string, value int) {
	g.serverProperties.Set(variable, value)
}

// GetConfigBool is a port of ServerConfigGroup::getConfigBool.
func (g *ServerConfigGroup) GetConfigBool(variable string, defaultValue bool) bool {
	var value any = defaultValue
	if v, ok := g.getopt(variable); ok {
		value = v
	} else if g.serverProperties.Exists(variable, false) {
		value = g.serverProperties.Get(variable, nil)
	}
	switch v := value.(type) {
	case bool:
		return v
	case int, int64:
		return toInt(v) != 0
	case string:
		switch strings.ToLower(v) {
		case "on", "true", "1", "yes":
			return true
		}
	}
	return false
}

func (g *ServerConfigGroup) SetConfigBool(variable string, value bool) {
	if value {
		g.serverProperties.Set(variable, "1")
	} else {
		g.serverProperties.Set(variable, "0")
	}
}

// Save is a port of ServerConfigGroup::save.
func (g *ServerConfigGroup) Save() error {
	if g.serverProperties.HasChanged() {
		if err := g.serverProperties.Save(); err != nil {
			return err
		}
	}
	if g.pocketmineYml != nil && g.pocketmineYml.HasChanged() {
		return g.pocketmineYml.Save()
	}
	return nil
}

// toInt is PHP's (int) cast for the scalar types a Config can hold.
func toInt(v any) int {
	switch v := v.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case bool:
		if v {
			return 1
		}
		return 0
	case string:
		// (int) "12abc" === 12 in PHP: parse the leading integer.
		s := strings.TrimSpace(v)
		end := 0
		for end < len(s) && (s[end] >= '0' && s[end] <= '9' || end == 0 && (s[end] == '-' || s[end] == '+')) {
			end++
		}
		i, _ := strconv.Atoi(s[:end])
		return i
	}
	return 0
}

// toBool is PHP's (bool) cast.
func toBool(v any) bool {
	switch v := v.(type) {
	case bool:
		return v
	case string:
		return v != "" && v != "0"
	case nil:
		return false
	}
	return toInt(v) != 0
}

// defaultArgs are the process's command-line arguments.
func defaultArgs() []string { return os.Args[1:] }
