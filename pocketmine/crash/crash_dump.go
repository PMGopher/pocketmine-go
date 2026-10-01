package crash

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/sandertv/gophertunnel/minecraft/protocol"

	"pocketmine-go/pocketmine"
	"pocketmine-go/pocketmine/plugin"
	"pocketmine-go/pocketmine/utils"
)

// formatVersion is CrashDump::FORMAT_VERSION.
const formatVersion = 4

// CrashDump::PLUGIN_INVOLVEMENT_*.
const (
	PluginInvolvementNone     = "none"
	PluginInvolvementDirect   = "direct"
	PluginInvolvementIndirect = "indirect"
)

// Server is what CrashDump needs from pocketmine\Server.
type Server interface {
	GetStartTime() time.Time
	GetDataPath() string
	GetName() string
	// GetPropertyBool is ServerConfigGroup::getPropertyBool (pocketmine.toml).
	GetPropertyBool(variable string, defaultValue bool) bool
}

// Frame is one frame of a crash's stack trace (PHP's ThreadCrashInfoFrame).
type Frame struct {
	Function string
	File     string
	Line     int
}

// PrintableFrame is ThreadCrashInfoFrame::getPrintableFrame's "#i file(line): function()".
func (f Frame) PrintableFrame(i int) string {
	return fmt.Sprintf("#%d %s(%d): %s()", i, utils.CleanPath(f.File), f.Line, f.Function)
}

// Crash is what the server records about the crash (PHP's global $lastExceptionError, set by
// Server::exceptionHandler): the panic's type and message, where it happened and the stack.
type Crash struct {
	Type    string
	Message string
	Trace   []Frame
	Thread  string
}

// FromPanic builds a Crash from a recovered panic value and the stack of the goroutine that
// panicked (runtime.Callers from the deferred function that recovered it), skipping the runtime's
// own panic frames.
func FromPanic(p any, pcs []uintptr, thread string) Crash {
	c := Crash{Type: fmt.Sprintf("%T", p), Thread: thread}
	if err, ok := p.(error); ok {
		c.Message = err.Error()
	} else {
		c.Message = fmt.Sprint(p)
	}
	frames := runtime.CallersFrames(pcs)
	for {
		f, more := frames.Next()
		if f.Function != "" && !strings.HasPrefix(f.Function, "runtime.") {
			c.Trace = append(c.Trace, Frame{Function: f.Function, File: f.File, Line: f.Line})
		}
		if !more {
			break
		}
	}
	return c
}

// CrashDump is a port of pocketmine\crash\CrashDump.
type CrashDump struct {
	server        Server
	pluginManager *plugin.PluginManager
	data          *CrashDumpData
	encodedData   []byte
}

// NewCrashDump is a port of CrashDump::__construct. pluginManager may be nil.
func NewCrashDump(server Server, pluginManager *plugin.PluginManager, crash Crash) (*CrashDump, error) {
	now := time.Now()
	d := &CrashDump{server: server, pluginManager: pluginManager, data: &CrashDumpData{
		FormatVersion: formatVersion,
		Time:          float64(now.UnixNano()) / 1e9,
		Uptime:        now.Sub(server.GetStartTime()).Seconds(),
		Code:          map[int]string{},
		Plugins:       map[string]CrashDumpDataPluginEntry{},
		Parameters:    []string{},
	}}

	d.baseCrash(crash)
	d.generalData()
	d.pluginsData()

	d.extraData()

	jsonData, err := json.Marshal(d.data)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	w, err := zlib.NewWriterLevel(&buf, 9)
	if err != nil {
		return nil, err
	}
	_, _ = w.Write(jsonData)
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("ZLIB compression failed: %w", err)
	}
	d.encodedData = buf.Bytes()
	return d, nil
}

func (d *CrashDump) GetEncodedData() []byte { return d.encodedData }

func (d *CrashDump) GetData() *CrashDumpData { return d.data }

// EncodeData is a port of CrashDump::encodeData.
func (d *CrashDump) EncodeData(renderer *CrashDumpRenderer) {
	renderer.AddLine("")
	renderer.AddLine("----------------------REPORT THE DATA BELOW THIS LINE-----------------------")
	renderer.AddLine("")
	renderer.AddLine("===BEGIN CRASH DUMP===")
	encoded := base64.StdEncoding.EncodeToString(d.encodedData)
	for len(encoded) > 76 {
		renderer.AddLine(encoded[:76])
		encoded = encoded[76:]
	}
	if encoded != "" {
		renderer.AddLine(encoded)
	}
	renderer.AddLine("===END CRASH DUMP===")
}

// pluginsData is a port of CrashDump::pluginsData.
func (d *CrashDump) pluginsData() {
	if d.pluginManager == nil {
		return
	}
	for _, p := range d.pluginManager.GetPlugins() {
		desc := p.GetDescription()
		load := "STARTUP"
		if desc.GetOrder() == plugin.EnableOrderPostworld {
			load = "POSTWORLD"
		}
		d.data.Plugins[desc.GetName()] = CrashDumpDataPluginEntry{
			Name:        desc.GetName(),
			Version:     desc.GetVersion(),
			Authors:     desc.GetAuthors(),
			API:         desc.GetCompatibleApis(),
			Enabled:     p.IsEnabled(),
			Depends:     desc.GetDepend(),
			SoftDepends: desc.GetSoftDepend(),
			Main:        desc.GetMain(),
			Load:        load,
			Website:     desc.GetWebsite(),
		}
	}
}

var rconPasswordPattern = regexp.MustCompile(`(?m)^rcon\.password=(.*)$`)

// extraData is a port of CrashDump::extraData.
func (d *CrashDump) extraData() {
	if d.server.GetPropertyBool("auto-report.send-settings", true) {
		d.data.Parameters = append([]string(nil), os.Args...)
		if serverDotProperties, err := os.ReadFile(filepath.Join(d.server.GetDataPath(), "server.properties")); err == nil {
			d.data.ServerDotProperties = rconPasswordPattern.ReplaceAllString(string(serverDotProperties), "rcon.password=******")
		}
		if pocketmineDotYml, err := os.ReadFile(filepath.Join(d.server.GetDataPath(), "pocketmine.toml")); err == nil {
			d.data.PocketmineDotYml = string(pocketmineDotYml)
		}
	}
}

// baseCrash is a port of CrashDump::baseCrash.
func (d *CrashDump) baseCrash(crash Crash) {
	message := crash.Message
	if pos := strings.Index(message, "\n"); pos >= 0 {
		message = message[:pos]
	}
	message = strings.ToValidUTF8(message, "�")

	var fullFile string
	line := 0
	if len(crash.Trace) > 0 {
		fullFile = crash.Trace[0].File
		line = crash.Trace[0].Line
	}
	d.data.Error = ErrorInfo{
		Type:    crash.Type,
		Message: message,
		File:    utils.CleanPath(fullFile),
		Line:    line,
		Thread:  crash.Thread,
	}

	d.data.PluginInvolvement = PluginInvolvementNone
	for i, frame := range crash.Trace {
		if d.determinePluginFromFrame(frame, i == 0) {
			break
		}
	}

	if d.server.GetPropertyBool("auto-report.send-code", true) && fullFile != "" {
		if f, err := os.Open(fullFile); err == nil {
			scanner := bufio.NewScanner(f)
			for n := 1; scanner.Scan(); n++ {
				if n >= line-10+1 && n < line+10+1 {
					d.data.Code[n] = scanner.Text()
				}
			}
			_ = f.Close()
		}
	}

	d.data.Trace = make([]string, len(crash.Trace))
	for i, frame := range crash.Trace {
		d.data.Trace[i] = frame.PrintableFrame(i)
	}
	d.data.Thread = crash.Thread
}

// determinePluginFromFrame is a port of CrashDump::determinePluginFromFile. PHP checks whether the
// frame's file is outside the server's src folder; here a frame belongs to a plugin when its
// function is in the Go package of one of the loaded plugins' main types.
func (d *CrashDump) determinePluginFromFrame(frame Frame, crashFrame bool) bool {
	if d.pluginManager == nil {
		return false
	}
	for _, p := range d.pluginManager.GetPlugins() {
		t := reflect.TypeOf(p)
		for t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		pkg := t.PkgPath()
		if pkg == "" || !strings.HasPrefix(frame.Function, pkg+".") {
			continue
		}
		if crashFrame {
			d.data.PluginInvolvement = PluginInvolvementDirect
		} else {
			d.data.PluginInvolvement = PluginInvolvementIndirect
		}
		d.data.Plugin = p.GetName()
		return true
	}
	return false
}

// generalData is a port of CrashDump::generalData.
func (d *CrashDump) generalData() {
	modules := map[string]string{}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, dep := range info.Deps {
			modules[dep.Path] = dep.Version + "@" + dep.Sum
		}
	}
	hostname, _ := os.Hostname()

	d.data.General = CrashDumpDataGeneral{
		Name:        d.server.GetName(),
		BaseVersion: pocketmine.BaseVersion,
		Build:       pocketmine.BuildNumber(),
		IsDev:       pocketmine.IsDevelopmentBuild,
		Protocol:    protocol.CurrentProtocol,
		Git:         pocketmine.GitHash(),
		Uname:       strings.Join([]string{runtime.GOOS, hostname, runtime.GOARCH}, " "),
		Go:          runtime.Version(),
		GoOS:        runtime.GOOS,
		OS:          utils.GetOS(),
		GoModules:   modules,
	}
}
