package plugin

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"pocketmine-go/pocketmine/permission"
)

var pluginNamePattern = regexp.MustCompile(`^[A-Za-z0-9 _.-]+$`)

// Description is a port of pocketmine\plugin\PluginDescription (parses a plugin.yml manifest).
type Description struct {
	name                       string
	version                    string
	main                       string
	srcNamespacePrefix         string
	api                        []string
	compatibleMcpeProtocols    []int
	compatibleOperatingSystems []string
	extensions                 map[string][]string
	depend                     []string
	softDepend                 []string
	loadBefore                 []string
	commands                   map[string]*DescriptionCommandEntry
	description                string
	authors                    []string
	website                    string
	prefix                     string
	order                      EnableOrder
	permissions                map[string][]*permission.Permission
	rawMap                     map[string]any
}

// NewDescriptionFromYAML is a port of the string-argument form of PluginDescription::__construct
// (`new PluginDescription($yamlString)`), parsing raw plugin.yml content.
func NewDescriptionFromYAML(yamlContent string) (*Description, error) {
	var node yaml.Node
	if err := yaml.Unmarshal([]byte(yamlContent), &node); err != nil {
		return nil, &PluginDescriptionParseException{Message: "YAML parsing error in plugin manifest: " + err.Error()}
	}
	if len(node.Content) == 0 {
		return nil, &PluginDescriptionParseException{Message: "Invalid structure of plugin manifest, expected array but have none"}
	}
	root := node.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, &PluginDescriptionParseException{Message: fmt.Sprintf("Invalid structure of plugin manifest, expected array but have %v", root.Kind)}
	}

	var m map[string]any
	if err := root.Decode(&m); err != nil {
		return nil, &PluginDescriptionParseException{Message: "YAML parsing error in plugin manifest: " + err.Error()}
	}

	return newDescriptionFromMap(m, root)
}

// NewDescriptionFromMap is a port of the array-argument form of PluginDescription::__construct
// (`new PluginDescription($map)`), for callers that already have a decoded manifest (e.g. a
// PharPluginLoader reading plugin.yml out of an archive that's already been parsed).
func NewDescriptionFromMap(m map[string]any) (*Description, error) {
	return newDescriptionFromMap(m, nil)
}

func newDescriptionFromMap(m map[string]any, rootNode *yaml.Node) (*Description, error) {
	d := &Description{rawMap: m, extensions: map[string][]string{}, commands: map[string]*DescriptionCommandEntry{}}
	if err := d.loadMap(m, rootNode); err != nil {
		return nil, err
	}
	return d, nil
}

func asString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case nil:
		return ""
	default:
		return fmt.Sprint(t)
	}
}

func asStringSlice(v any) []string {
	switch t := v.(type) {
	case nil:
		return nil
	case []any:
		out := make([]string, len(t))
		for i, e := range t {
			out[i] = asString(e)
		}
		return out
	case []string:
		return t
	default:
		return []string{asString(v)}
	}
}

func asIntSlice(v any) []int {
	strs := asStringSlice(v)
	out := make([]int, 0, len(strs))
	for _, s := range strs {
		n, _ := strconv.Atoi(strings.TrimSpace(s))
		out = append(out, n)
	}
	return out
}

// loadMap is a port of PluginDescription::loadMap.
func (d *Description) loadMap(m map[string]any, rootNode *yaml.Node) error {
	nameRaw, ok := m["name"]
	if !ok {
		return &PluginDescriptionParseException{Message: "Missing Plugin name"}
	}
	name := asString(nameRaw)
	if !pluginNamePattern.MatchString(name) {
		return &PluginDescriptionParseException{Message: "Invalid Plugin name"}
	}
	d.name = strings.ReplaceAll(name, " ", "_")

	d.version = asString(m["version"])

	main, ok := m["main"]
	if !ok {
		return &PluginDescriptionParseException{Message: "Missing Plugin main"}
	}
	d.main = asString(main)
	if strings.HasPrefix(strings.ToLower(d.main), "pocketmine\\") {
		return &PluginDescriptionParseException{Message: "Invalid Plugin main, cannot start within the PocketMine namespace"}
	}

	d.srcNamespacePrefix = asString(m["src-namespace-prefix"])

	d.api = asStringSlice(m["api"])
	d.compatibleMcpeProtocols = asIntSlice(m["mcpe-protocol"])
	d.compatibleOperatingSystems = asStringSlice(m["os"])

	if commandsRaw, ok := m["commands"]; ok {
		commandsMap, ok := commandsRaw.(map[string]any)
		if !ok {
			return &PluginDescriptionParseException{Message: "Invalid Plugin commands, expected a mapping"}
		}
		for commandName, commandDataRaw := range commandsMap {
			commandData, ok := commandDataRaw.(map[string]any)
			if !ok {
				return &PluginDescriptionParseException{Message: fmt.Sprintf("Command %s has invalid properties", commandName)}
			}
			permRaw, ok := commandData["permission"]
			permStr, permIsString := permRaw.(string)
			if !ok || !permIsString {
				return &PluginDescriptionParseException{Message: fmt.Sprintf("Command %s does not have a valid permission set", commandName)}
			}
			entry := &DescriptionCommandEntry{
				Aliases:    asStringSlice(commandData["aliases"]),
				Permission: permStr,
			}
			if v, ok := commandData["description"]; ok {
				s := asString(v)
				entry.Description = &s
			}
			if v, ok := commandData["usage"]; ok {
				s := asString(v)
				entry.UsageMessage = &s
			}
			if v, ok := commandData["permission-message"]; ok {
				s := asString(v)
				entry.PermissionDeniedMessage = &s
			}
			d.commands[commandName] = entry
		}
	}

	if v, ok := m["depend"]; ok {
		d.depend = asStringSlice(v)
	}

	if extensionsRaw, ok := m["extensions"]; ok {
		switch ext := extensionsRaw.(type) {
		case []any:
			for _, name := range ext {
				d.extensions[asString(name)] = []string{"*"}
			}
		case map[string]any:
			for name, v := range ext {
				d.extensions[name] = asStringSlice(v)
			}
		}
	}

	d.softDepend = asStringSlice(m["softdepend"])
	d.loadBefore = asStringSlice(m["loadbefore"])
	d.website = asString(m["website"])
	d.description = asString(m["description"])
	d.prefix = asString(m["prefix"])

	if loadRaw, ok := m["load"]; ok {
		order, ok := EnableOrderFromString(asString(loadRaw))
		if !ok {
			return &PluginDescriptionParseException{Message: "Invalid Plugin \"load\""}
		}
		d.order = order
	} else {
		d.order = EnableOrderPostworld
	}

	if authorRaw, ok := m["author"]; ok {
		if list, ok := authorRaw.([]any); ok {
			for _, a := range list {
				d.authors = append(d.authors, asString(a))
			}
		} else {
			d.authors = append(d.authors, asString(authorRaw))
		}
	}
	if authorsRaw, ok := m["authors"]; ok {
		if list, ok := authorsRaw.([]any); ok {
			for _, a := range list {
				d.authors = append(d.authors, asString(a))
			}
		}
	}

	if _, ok := m["permissions"]; ok {
		entries, err := d.permissionEntries(rootNode)
		if err != nil {
			return &PluginDescriptionParseException{Message: "Invalid Plugin \"permissions\": " + err.Error()}
		}
		perms, err := permission.LoadPermissions(entries, permission.DefaultFalse)
		if err != nil {
			return &PluginDescriptionParseException{Message: "Invalid Plugin \"permissions\": " + err.Error()}
		}
		d.permissions = perms
	}

	return nil
}

// permissionEntries builds the order-preserving []permission.PermissionEntry LoadPermissions
// needs (see LoadPermissions' own doc comment on why declaration order is load-bearing: a
// "default" on one entry carries forward to later entries that don't specify their own). If
// rootNode is nil (the caller constructed this Description from an already-decoded map, not raw
// YAML - see NewDescriptionFromMap), falls back to Go map iteration order, which is NOT
// guaranteed to match the original declaration order - a real, documented limitation of that
// entry point specifically, not a bug in the YAML-parsing path.
func (d *Description) permissionEntries(rootNode *yaml.Node) ([]permission.PermissionEntry, error) {
	if rootNode != nil {
		permsNode := findMappingValue(rootNode, "permissions")
		if permsNode != nil && permsNode.Kind == yaml.MappingNode {
			return permissionEntriesFromNode(permsNode)
		}
	}

	permsMap, _ := d.rawMap["permissions"].(map[string]any)
	entries := make([]permission.PermissionEntry, 0, len(permsMap))
	for name, raw := range permsMap {
		entry, err := permissionEntryFromValue(name, raw)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func findMappingValue(mappingNode *yaml.Node, key string) *yaml.Node {
	if mappingNode.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(mappingNode.Content); i += 2 {
		if mappingNode.Content[i].Value == key {
			return mappingNode.Content[i+1]
		}
	}
	return nil
}

func permissionEntriesFromNode(permsNode *yaml.Node) ([]permission.PermissionEntry, error) {
	entries := make([]permission.PermissionEntry, 0, len(permsNode.Content)/2)
	for i := 0; i+1 < len(permsNode.Content); i += 2 {
		name := permsNode.Content[i].Value
		var raw any
		if err := permsNode.Content[i+1].Decode(&raw); err != nil {
			return nil, err
		}
		entry, err := permissionEntryFromValue(name, raw)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func permissionEntryFromValue(name string, raw any) (permission.PermissionEntry, error) {
	entry := permission.PermissionEntry{Name: name}
	data, ok := raw.(map[string]any)
	if !ok {
		return entry, nil
	}
	if v, ok := data["default"]; ok {
		entry.Default = asString(v)
	}
	if _, ok := data["children"]; ok {
		entry.HasChildren = true
	}
	if v, ok := data["description"]; ok {
		entry.Description = v
	}
	return entry, nil
}

// GetFullName is a port of PluginDescription::getFullName.
func (d *Description) GetFullName() string { return d.name + " v" + d.version }

func (d *Description) GetCompatibleApis() []string                         { return d.api }
func (d *Description) GetCompatibleMcpeProtocols() []int                   { return d.compatibleMcpeProtocols }
func (d *Description) GetCompatibleOperatingSystems() []string             { return d.compatibleOperatingSystems }
func (d *Description) GetAuthors() []string                                { return d.authors }
func (d *Description) GetPrefix() string                                   { return d.prefix }
func (d *Description) GetCommands() map[string]*DescriptionCommandEntry    { return d.commands }
func (d *Description) GetRequiredExtensions() map[string][]string          { return d.extensions }
func (d *Description) GetDepend() []string                                 { return d.depend }
func (d *Description) GetDescription() string                              { return d.description }
func (d *Description) GetLoadBefore() []string                             { return d.loadBefore }
func (d *Description) GetMain() string                                     { return d.main }
func (d *Description) GetSrcNamespacePrefix() string                       { return d.srcNamespacePrefix }
func (d *Description) GetName() string                                     { return d.name }
func (d *Description) GetOrder() EnableOrder                               { return d.order }
func (d *Description) GetPermissions() map[string][]*permission.Permission { return d.permissions }
func (d *Description) GetSoftDepend() []string                             { return d.softDepend }
func (d *Description) GetVersion() string                                  { return d.version }
func (d *Description) GetWebsite() string                                  { return d.website }
func (d *Description) GetMap() map[string]any                              { return d.rawMap }
