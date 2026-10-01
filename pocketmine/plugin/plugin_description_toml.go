package plugin

import (
	"github.com/pelletier/go-toml/v2/unstable"

	"pocketmine-go/pocketmine/utils"
)

// NewDescriptionFromTOML parses a plugin.toml manifest: the same keys as PocketMine-MP's
// plugin.yml (name, version, main, api, depend, commands, permissions, ...) in TOML.
func NewDescriptionFromTOML(content string) (*Description, error) {
	m, err := utils.ParseTOML([]byte(content))
	if err != nil {
		return nil, &PluginDescriptionParseException{Message: "TOML parsing error in plugin manifest: " + err.Error()}
	}
	d := &Description{
		rawMap:          m,
		extensions:      map[string][]string{},
		commands:        map[string]*DescriptionCommandEntry{},
		permissionOrder: tomlChildKeyOrder([]byte(content), "permissions"),
	}
	if err := d.loadMap(m, nil); err != nil {
		return nil, err
	}
	return d, nil
}

// tomlChildKeyOrder returns the keys of the table called table in the order the document declares
// them. Permissions need it: a "default" carries forward to the permissions declared after it
// (see permissionEntries), and decoded maps have no order.
func tomlChildKeyOrder(content []byte, table string) []string {
	var p unstable.Parser
	p.Reset(content)
	var current, order []string
	seen := map[string]bool{}
	add := func(path []string) {
		if len(path) >= 2 && path[0] == table && !seen[path[1]] {
			seen[path[1]] = true
			order = append(order, path[1])
		}
	}
	for p.NextExpression() {
		e := p.Expression()
		switch e.Kind {
		case unstable.Table, unstable.ArrayTable:
			current = tomlKeyParts(e)
			add(current)
		case unstable.KeyValue:
			add(append(append([]string{}, current...), tomlKeyParts(e)...))
		}
	}
	return order
}

func tomlKeyParts(n *unstable.Node) []string {
	var parts []string
	for it := n.Key(); it.Next(); {
		parts = append(parts, string(it.Node().Data))
	}
	return parts
}
