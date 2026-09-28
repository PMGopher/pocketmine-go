// Package crash is a port of pocketmine\crash: the crash dump written when the server crashes.
package crash

import "encoding/json"

// CrashDumpDataGeneral is a port of pocketmine\crash\CrashDumpDataGeneral. PHP's php, zend and
// composer_libraries fields describe the PHP runtime; here they're go (the Go version) and
// go_modules (the modules compiled into the server).
type CrashDumpDataGeneral struct {
	Name        string            `json:"name"`
	BaseVersion string            `json:"base_version"`
	Build       int               `json:"build"`
	IsDev       bool              `json:"is_dev"`
	Protocol    int               `json:"protocol"`
	Git         string            `json:"git"`
	Uname       string            `json:"uname"`
	Go          string            `json:"go"`
	GoOS        string            `json:"go_os"`
	OS          string            `json:"os"`
	GoModules   map[string]string `json:"go_modules"`
}

// CrashDumpDataPluginEntry is a port of pocketmine\crash\CrashDumpDataPluginEntry.
type CrashDumpDataPluginEntry struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Authors     []string `json:"authors"`
	API         []string `json:"api"`
	Enabled     bool     `json:"enabled"`
	Depends     []string `json:"depends"`
	SoftDepends []string `json:"softDepends"`
	Main        string   `json:"main"`
	Load        string   `json:"load"`
	Website     string   `json:"website"`
}

// ErrorInfo is the error array of CrashDumpData (PHP's $lastError / error_get_last() shape).
type ErrorInfo struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	File    string `json:"file"`
	Line    int    `json:"line"`
	Thread  string `json:"thread"`
}

// CrashDumpData is a port of pocketmine\crash\CrashDumpData. PHP's extensions, jit_mode and
// phpinfo describe the PHP runtime and have no counterpart.
type CrashDumpData struct {
	FormatVersion       int                                 `json:"format_version"`
	Time                float64                             `json:"time"`
	Uptime              float64                             `json:"uptime"`
	LastError           *ErrorInfo                          `json:"lastError,omitempty"`
	Error               ErrorInfo                           `json:"error"`
	Thread              string                              `json:"thread"`
	PluginInvolvement   string                              `json:"plugin_involvement"`
	Plugin              string                              `json:"plugin"`
	Code                map[int]string                      `json:"code"`
	Trace               []string                            `json:"trace"`
	Plugins             map[string]CrashDumpDataPluginEntry `json:"plugins"`
	Parameters          []string                            `json:"parameters"`
	ServerDotProperties string                              `json:"-"`
	PocketmineDotYml    string                              `json:"-"`
	General             CrashDumpDataGeneral                `json:"general"`
}

// MarshalJSON is a port of CrashDumpData::jsonSerialize: the config files go under their file
// names.
func (d *CrashDumpData) MarshalJSON() ([]byte, error) {
	type plain CrashDumpData
	raw, err := json.Marshal((*plain)(d))
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	result["pocketmine.yml"] = d.PocketmineDotYml
	result["server.properties"] = d.ServerDotProperties
	return json.Marshal(result)
}
