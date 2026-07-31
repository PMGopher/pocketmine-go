package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

type ConfigType int

const (
	ConfigDetect     ConfigType = -1 // Detect by file extension
	ConfigProperties ConfigType = 0  // .properties
	ConfigCNF                   = ConfigProperties
	ConfigJSON       ConfigType = 1 // .js, .json
	ConfigYAML       ConfigType = 2 // .yml, .yaml
	// ConfigSerialized (PHP's `serialize()` binary format, historically type 4) is deliberately
	// not ported: nothing outside a PHP runtime can produce or consume it, so it has no meaning
	// for a from-scratch Go server. Any config file with that type fails to load with a clear error.
	ConfigEnum ConfigType = 5 // .txt, .list, .enum
)

var configFormatsByExtension = map[string]ConfigType{
	"properties": ConfigProperties,
	"cnf":        ConfigCNF,
	"conf":       ConfigCNF,
	"config":     ConfigCNF,
	"json":       ConfigJSON,
	"js":         ConfigJSON,
	"yml":        ConfigYAML,
	"yaml":       ConfigYAML,
	"txt":        ConfigEnum,
	"list":       ConfigEnum,
	"enum":       ConfigEnum,
}

// Config is a port of pocketmine\utils\Config: simple config file manipulation across multiple
// formats (properties, JSON, YAML, a newline-list "enum" format).
//
// PHP's `mixed[]` config tree is represented here as map[string]any, with nested values as
// map[string]any or basic JSON/YAML-compatible types (string, float64/json.Number, bool, nil,
// []any) — the same dynamic-typing tradeoff PHP made, just made explicit via `any` instead of
// being implicit in the language.
type Config struct {
	mu          sync.Mutex
	data        map[string]any
	nestedCache map[string]any
	file        string
	configType  ConfigType
	changed     bool
}

// NewConfig loads (or creates, with the given defaults) a config file, mirroring the Config
// constructor. Pass ConfigDetect to detect the type from the file extension.
func NewConfig(file string, configType ConfigType, defaults map[string]any) (*Config, error) {
	c := &Config{nestedCache: map[string]any{}}
	if err := c.load(file, configType, defaults); err != nil {
		return nil, err
	}
	return c, nil
}

// Reload discards all in-memory changes and loads the file again.
func (c *Config) Reload() error {
	c.mu.Lock()
	file, configType := c.file, c.configType
	c.mu.Unlock()
	return c.load(file, configType, nil)
}

func (c *Config) HasChanged() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.changed
}

func (c *Config) SetChanged(changed bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.changed = changed
}

var yamlBareKeywordPattern = regexp.MustCompile(`(?m)^( *)(y|Y|yes|Yes|YES|n|N|no|No|NO|true|True|TRUE|false|False|FALSE|on|On|ON|off|Off|OFF)( *):`)

// FixYAMLIndexes quotes YAML mapping keys that look like booleans (y/n/yes/no/on/off/true/false),
// so the YAML parser treats them as string keys rather than coercing them to bool.
func FixYAMLIndexes(s string) string {
	return yamlBareKeywordPattern.ReplaceAllString(s, `$1"$2"$3:`)
}

func (c *Config) load(file string, configType ConfigType, defaults map[string]any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.file = file
	c.configType = configType
	if c.configType == ConfigDetect {
		ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(file), "."))
		t, ok := configFormatsByExtension[ext]
		if !ok {
			return fmt.Errorf("cannot detect config type of %s", file)
		}
		c.configType = t
	}

	if _, err := os.Stat(file); os.IsNotExist(err) {
		c.data = defaults
		if c.data == nil {
			c.data = map[string]any{}
		}
		return c.saveLocked()
	}

	content, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", file, err)
	}

	var data map[string]any
	switch c.configType {
	case ConfigProperties:
		data = ParseProperties(string(content))
	case ConfigJSON:
		dec := json.NewDecoder(bytes.NewReader(content))
		dec.UseNumber()
		if err := dec.Decode(&data); err != nil {
			return WrapConfigLoadException(file, err)
		}
	case ConfigYAML:
		fixed := FixYAMLIndexes(string(content))
		if err := yaml.Unmarshal([]byte(fixed), &data); err != nil {
			return WrapConfigLoadException(file, err)
		}
	case ConfigEnum:
		data = map[string]any{}
		for _, entry := range ParseList(string(content)) {
			data[entry] = true
		}
	default:
		return fmt.Errorf("invalid or unsupported config type specified")
	}
	if data == nil {
		data = map[string]any{}
	}
	c.data = data
	c.nestedCache = map[string]any{}

	if defaults != nil && fillDefaults(defaults, c.data) > 0 {
		return c.saveLocked()
	}
	return nil
}

func (c *Config) GetPath() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.file
}

// Save flushes the config to disk in its configured format.
func (c *Config) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.saveLocked()
}

func (c *Config) saveLocked() error {
	var content []byte
	switch c.configType {
	case ConfigProperties:
		content = []byte(WriteProperties(c.data))
	case ConfigJSON:
		b, err := json.MarshalIndent(c.data, "", "    ")
		if err != nil {
			return err
		}
		content = b
	case ConfigYAML:
		b, err := yaml.Marshal(c.data)
		if err != nil {
			return err
		}
		content = b
	case ConfigEnum:
		keys := make([]string, 0, len(c.data))
		for k := range c.data {
			keys = append(keys, k)
		}
		content = []byte(WriteList(keys))
	default:
		return NewAssumptionFailedError("Config type is unknown, has not been set, or is unsupported")
	}

	if err := SafeFilePutContents(c.file, content); err != nil {
		return err
	}
	c.changed = false
	return nil
}

// SetNested sets a dot-separated nested key, creating intermediate maps as needed.
func (c *Config) SetNested(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	parts := strings.Split(key, ".")
	node := c.data
	for i, part := range parts {
		if i == len(parts)-1 {
			node[part] = value
			break
		}
		next, ok := node[part].(map[string]any)
		if !ok {
			next = map[string]any{}
			node[part] = next
		}
		node = next
	}
	c.nestedCache = map[string]any{}
	c.changed = true
}

// GetNested reads a dot-separated nested key, returning defaultValue if any segment is missing.
func (c *Config) GetNested(key string, defaultValue any) any {
	c.mu.Lock()
	defer c.mu.Unlock()

	if v, ok := c.nestedCache[key]; ok {
		return v
	}

	parts := strings.Split(key, ".")
	var current any = c.data
	for _, part := range parts {
		m, ok := current.(map[string]any)
		if !ok {
			return defaultValue
		}
		v, ok := m[part]
		if !ok {
			return defaultValue
		}
		current = v
	}
	c.nestedCache[key] = current
	return current
}

// RemoveNested removes a dot-separated nested key.
func (c *Config) RemoveNested(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.nestedCache = map[string]any{}
	c.changed = true

	parts := strings.Split(key, ".")
	node := c.data
	for i, part := range parts {
		if i == len(parts)-1 {
			delete(node, part)
			return
		}
		next, ok := node[part].(map[string]any)
		if !ok {
			return
		}
		node = next
	}
}

func (c *Config) Get(key string, defaultValue any) any {
	c.mu.Lock()
	defer c.mu.Unlock()
	if v, ok := c.data[key]; ok {
		return v
	}
	return defaultValue
}

func (c *Config) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
	c.changed = true
	prefix := key + "."
	for nestedKey := range c.nestedCache {
		if strings.HasPrefix(nestedKey, prefix) {
			delete(c.nestedCache, nestedKey)
		}
	}
}

func (c *Config) SetAll(data map[string]any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = data
	c.changed = true
}

// Exists reports whether key is set. If lowercase is true, matches case-insensitively.
func (c *Config) Exists(key string, lowercase bool) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !lowercase {
		_, ok := c.data[key]
		return ok
	}
	key = strings.ToLower(key)
	for k := range c.data {
		if strings.ToLower(k) == key {
			return true
		}
	}
	return false
}

func (c *Config) Remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
	c.changed = true
}

// GetAll returns a snapshot of the whole config tree.
//
// PHP's getAll(bool $keys) changes its return shape (values vs. just keys) based on a flag;
// Go can't express that in one return type, so this is split into GetAll and GetKeys.
func (c *Config) GetAll() map[string]any {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make(map[string]any, len(c.data))
	for k, v := range c.data {
		result[k] = v
	}
	return result
}

func (c *Config) GetKeys() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	keys := make([]string, 0, len(c.data))
	for k := range c.data {
		keys = append(keys, k)
	}
	return keys
}

func (c *Config) SetDefaults(defaults map[string]any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if fillDefaults(defaults, c.data) > 0 {
		c.changed = true
	}
}

func fillDefaults(defaults map[string]any, data map[string]any) int {
	changed := 0
	for k, v := range defaults {
		if sub, isMap := v.(map[string]any); isMap {
			existing, ok := data[k].(map[string]any)
			if !ok {
				existing = map[string]any{}
				data[k] = existing
			}
			changed += fillDefaults(sub, existing)
		} else if _, ok := data[k]; !ok {
			data[k] = v
			changed++
		}
	}
	return changed
}

// ParseList parses the "enum" format: one entry per non-empty, trimmed line.
func ParseList(content string) []string {
	var result []string
	content = strings.TrimSpace(strings.ReplaceAll(content, "\r\n", "\n"))
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}
	return result
}

func WriteList(entries []string) string {
	return strings.Join(entries, "\n")
}

// WriteProperties writes a Java-.properties-style file.
func WriteProperties(config map[string]any) string {
	var b strings.Builder
	b.WriteString("#Properties Config file\r\n#")
	b.WriteString(time.Now().Format("Mon Jan 2 15:04:05 MST 2006"))
	b.WriteString("\r\n")
	for k, v := range config {
		var s string
		switch value := v.(type) {
		case bool:
			if value {
				s = "on"
			} else {
				s = "off"
			}
		default:
			s = fmt.Sprintf("%v", value)
		}
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(s)
		b.WriteString("\r\n")
	}
	return b.String()
}

var propertiesLinePattern = regexp.MustCompile(`(?m)^\s*([a-zA-Z0-9\-_.]+)[ \t]*=([^\r\n]*)`)

// ParseProperties parses a Java-.properties-style file, coercing on/true/yes -> true,
// off/false/no -> false, and numeric-looking values to int64/float64.
func ParseProperties(content string) map[string]any {
	result := map[string]any{}
	for _, m := range propertiesLinePattern.FindAllStringSubmatch(content, -1) {
		key := m[1]
		value := strings.TrimSpace(m[2])
		switch strings.ToLower(value) {
		case "on", "true", "yes":
			result[key] = true
		case "off", "false", "no":
			result[key] = false
		default:
			if i, err := strconv.ParseInt(value, 10, 64); err == nil {
				result[key] = i
			} else if f, err := strconv.ParseFloat(value, 64); err == nil {
				result[key] = f
			} else {
				result[key] = value
			}
		}
	}
	return result
}
