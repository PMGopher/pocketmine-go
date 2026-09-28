package crash

import (
	"fmt"
	"io"
	"runtime"
	"sort"
	"strings"
	"time"

	"pocketmine-go/pocketmine/utils"
)

// CrashDumpRenderer is a port of pocketmine\crash\CrashDumpRenderer.
type CrashDumpRenderer struct {
	fp   io.Writer
	data *CrashDumpData
}

func NewCrashDumpRenderer(fp io.Writer, data *CrashDumpData) *CrashDumpRenderer {
	return &CrashDumpRenderer{fp: fp, data: data}
}

// RenderHumanReadable is a port of CrashDumpRenderer::renderHumanReadable. The PHP and Zend
// version lines are the Go version, and the Composer libraries are the Go modules.
func (r *CrashDumpRenderer) RenderHumanReadable() {
	general := r.data.General
	r.AddLine(general.Name + " Crash Dump " + time.Unix(0, int64(r.data.Time*1e9)).Format("Mon Jan 2 15:04:05 MST 2006"))
	r.AddLine("")

	version, err := utils.NewVersionString(general.BaseVersion, general.IsDev, general.Build)
	fullVersion := general.BaseVersion
	if err == nil {
		fullVersion = version.GetFullVersion(true)
	}
	r.AddLine(fmt.Sprintf("%s version: %s [Protocol %d]", general.Name, fullVersion, general.Protocol))
	r.AddLine("Git commit: " + general.Git)
	r.AddLine("Go version: " + general.Go)
	r.AddLine("OS: " + general.GoOS + ", " + general.OS)

	if r.data.PluginInvolvement != PluginInvolvementNone {
		r.AddLine("")
		switch r.data.PluginInvolvement {
		case PluginInvolvementDirect:
			r.AddLine("THIS CRASH WAS CAUSED BY A PLUGIN")
		case PluginInvolvementIndirect:
			r.AddLine("A PLUGIN WAS INVOLVED IN THIS CRASH")
		default:
			r.AddLine("Unknown plugin involvement!")
		}
	}
	if r.data.Plugin != "" {
		r.AddLine("BAD PLUGIN: " + r.data.Plugin)
	}

	r.AddLine("")

	r.AddLine("Thread: " + r.data.Thread)
	r.AddLine("Error: " + r.data.Error.Message)
	r.AddLine("File: " + r.data.Error.File)
	r.AddLine(fmt.Sprintf("Line: %d", r.data.Error.Line))
	r.AddLine("Type: " + r.data.Error.Type)
	r.AddLine("Backtrace:")
	for _, line := range r.data.Trace {
		r.AddLine(line)
	}

	r.AddLine("")
	r.AddLine("Code:")

	lineNumbers := make([]int, 0, len(r.data.Code))
	for n := range r.data.Code {
		lineNumbers = append(lineNumbers, n)
	}
	sort.Ints(lineNumbers)
	for _, n := range lineNumbers {
		r.AddLine(fmt.Sprintf("[%d] %s", n, r.data.Code[n]))
	}

	if len(r.data.Plugins) > 0 {
		r.AddLine("")
		r.AddLine("Loaded plugins:")
		for _, name := range sortedPluginNames(r.data.Plugins) {
			p := r.data.Plugins[name]
			r.AddLine(p.Name + " " + p.Version + " by " + strings.Join(p.Authors, ", ") + " for API(s) " + strings.Join(p.API, ", "))
		}
	}

	r.AddLine("")
	r.AddLine("uname -a: " + general.Uname)
	r.AddLine("Go modules: ")
	modules := make([]string, 0, len(general.GoModules))
	for m := range general.GoModules {
		modules = append(modules, m)
	}
	sort.Strings(modules)
	for _, m := range modules {
		r.AddLine("- " + m + " " + general.GoModules[m])
	}
}

func sortedPluginNames(plugins map[string]CrashDumpDataPluginEntry) []string {
	names := make([]string, 0, len(plugins))
	for n := range plugins {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// AddLine is a port of CrashDumpRenderer::addLine.
func (r *CrashDumpRenderer) AddLine(line string) {
	eol := "\n"
	if runtime.GOOS == "windows" {
		eol = "\r\n" // PHP_EOL
	}
	_, _ = io.WriteString(r.fp, line+eol)
}
