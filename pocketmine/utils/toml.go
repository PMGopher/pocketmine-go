package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// The server's config files (pocketmine.toml, plugin.toml, plugin configs, plugin_list.toml,
// resource_packs.toml) are TOML. Config trees are map[string]any either way, so the code reading
// them doesn't depend on the format.

// ParseTOML decodes a TOML document into a config tree. Integers come back as int (go-toml gives
// int64), like the other formats, so readers can keep asserting .(int).
func ParseTOML(content []byte) (map[string]any, error) {
	data := map[string]any{}
	if err := toml.Unmarshal(content, &data); err != nil {
		return nil, err
	}
	return normalizeTOMLValue(data).(map[string]any), nil
}

func normalizeTOMLValue(v any) any {
	switch v := v.(type) {
	case int64:
		return int(v)
	case map[string]any:
		for k, e := range v {
			v[k] = normalizeTOMLValue(e)
		}
		return v
	case []any:
		for i, e := range v {
			v[i] = normalizeTOMLValue(e)
		}
		return v
	}
	return v
}

// MarshalTOML encodes a config tree as TOML. TOML has no null, so nil values are left out.
func MarshalTOML(data map[string]any) ([]byte, error) {
	return toml.Marshal(withoutNils(data))
}

func withoutNils(v any) any {
	switch v := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for k, e := range v {
			if e != nil {
				out[k] = withoutNils(e)
			}
		}
		return out
	case map[any]any: // from old YAML files
		out := make(map[string]any, len(v))
		for k, e := range v {
			if e != nil {
				out[fmt.Sprint(k)] = withoutNils(e)
			}
		}
		return out
	case []any:
		out := make([]any, 0, len(v))
		for _, e := range v {
			if e != nil {
				out = append(out, withoutNils(e))
			}
		}
		return out
	}
	return v
}

// ConvertYAMLToTOML moves a config file of an older version of this server from YAML to TOML:
// if tomlFile doesn't exist yet and yamlFile does, tomlFile is written with the same settings
// (under header, a comment) and yamlFile is renamed to <yamlFile>.bak. It reports whether it
// converted anything.
func ConvertYAMLToTOML(yamlFile, tomlFile, header string) (bool, error) {
	if _, err := os.Stat(tomlFile); err == nil {
		return false, nil
	}
	content, err := os.ReadFile(yamlFile)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	var data map[string]any
	if err := yaml.Unmarshal([]byte(FixYAMLIndexes(string(content))), &data); err != nil {
		return false, WrapConfigLoadException(yamlFile, err)
	}
	if data == nil {
		data = map[string]any{}
	}
	converted, err := MarshalTOML(data)
	if err != nil {
		return false, fmt.Errorf("converting %s to TOML: %w", yamlFile, err)
	}
	var out strings.Builder
	for _, line := range strings.Split(strings.TrimSpace(header), "\n") {
		out.WriteString("# " + line + "\n")
	}
	out.WriteString("\n")
	out.Write(converted)
	if err := SafeFilePutContents(tomlFile, []byte(out.String())); err != nil {
		return false, err
	}
	if err := os.Rename(yamlFile, yamlFile+".bak"); err != nil {
		return true, err
	}
	return true, nil
}
