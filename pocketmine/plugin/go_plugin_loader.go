package plugin

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
)

// goPluginFileExtension is the extension of the virtual plugin files a compiled-in Go plugin
// appears as in the plugins folder (see GoPluginLoader).
const goPluginFileExtension = ".goplugin"

// goPluginRegistration is a plugin registered with RegisterGoPlugin.
type goPluginRegistration struct {
	id        string
	fsys      fs.FS
	newPlugin func() Plugin
}

var (
	goPluginsMu sync.Mutex
	goPlugins   []*goPluginRegistration
)

// RegisterGoPlugin registers a plugin compiled into the server. Call it from the plugin package's
// init() and import that package (e.g. with a blank import in cmd/pocketmine-go):
//
//	//go:embed plugin.toml resources
//	var files embed.FS
//
//	func init() { plugin.RegisterGoPlugin(files, func() plugin.Plugin { return &Main{} }) }
//
// fsys is the plugin's root folder, like the root of a PHP plugin: it must contain plugin.toml (or,
// for plugins written before it, plugin.yml), and may contain a resources folder. newPlugin creates
// the plugin's main object, which is what the "main" class in the manifest is to a PHP plugin; it
// must implement Initializer (embedding PluginBase does that).
//
// The plugin is then loaded by the PluginManager like any other: plugin_list.toml, API version,
// dependencies, load order, data folder (plugins/<name>/ or plugin_data/<name>/) and commands
// from plugin.toml all work as for PHP plugins.
func RegisterGoPlugin(fsys fs.FS, newPlugin func() Plugin) {
	goPluginsMu.Lock()
	defer goPluginsMu.Unlock()
	goPlugins = append(goPlugins, &goPluginRegistration{id: fmt.Sprintf("%d", len(goPlugins)), fsys: fsys, newPlugin: newPlugin})
}

// GoPluginLoader is the PluginLoader for compiled-in Go plugins: the counterpart of
// PharPluginLoader and ScriptPluginLoader, which load PHP code and can't exist in a Go server.
// Each registered plugin appears as a virtual file "<plugins folder>/<id>.goplugin" to the
// PluginManager's triage.
type GoPluginLoader struct{}

func NewGoPluginLoader() *GoPluginLoader { return &GoPluginLoader{} }

func (l *GoPluginLoader) registration(path string) *goPluginRegistration {
	base := filepath.Base(path)
	if !strings.HasSuffix(base, goPluginFileExtension) {
		return nil
	}
	id := strings.TrimSuffix(base, goPluginFileExtension)
	goPluginsMu.Lock()
	defer goPluginsMu.Unlock()
	for _, r := range goPlugins {
		if r.id == id {
			return r
		}
	}
	return nil
}

func (l *GoPluginLoader) CanLoadPlugin(path string) bool { return l.registration(path) != nil }

// LoadPlugin does nothing: the plugin's code is already compiled in.
func (l *GoPluginLoader) LoadPlugin(file string) {}

func (l *GoPluginLoader) GetPluginDescription(file string) (*Description, error) {
	r := l.registration(file)
	if r == nil {
		return nil, nil
	}
	if manifest, err := fs.ReadFile(r.fsys, "plugin.toml"); err == nil {
		return NewDescriptionFromTOML(string(manifest))
	}
	// Plugins written before plugin.toml still load.
	if manifest, err := fs.ReadFile(r.fsys, "plugin.yml"); err == nil {
		return NewDescriptionFromYAML(string(manifest))
	}
	return nil, nil
}

func (l *GoPluginLoader) GetAccessProtocol() string { return "" }

// GetPluginFiles is the virtual files of the registered plugins in dir (see pluginSource).
func (l *GoPluginLoader) GetPluginFiles(dir string) []string {
	goPluginsMu.Lock()
	defer goPluginsMu.Unlock()
	files := make([]string, len(goPlugins))
	for i, r := range goPlugins {
		files[i] = filepath.Join(dir, r.id+goPluginFileExtension)
	}
	return files
}

// NewPlugin creates the registered plugin's main object (see pluginFactory).
func (l *GoPluginLoader) NewPlugin(file string, description *Description) (Plugin, bool) {
	r := l.registration(file)
	if r == nil || r.newPlugin == nil {
		return nil, false
	}
	p := r.newPlugin()
	return p, p != nil
}

// GetResourceProvider is the registered plugin's resources folder (see pluginFactory).
func (l *GoPluginLoader) GetResourceProvider(file string) ResourceProvider {
	r := l.registration(file)
	if r == nil {
		return NewFSResourceProvider(nil)
	}
	resources, err := fs.Sub(r.fsys, "resources")
	if err != nil {
		return NewFSResourceProvider(nil)
	}
	return NewFSResourceProvider(resources)
}
