package pocketmine

import (
	"os"
	"path/filepath"
)

// Path constants are a port of pocketmine\CoreConstants's define()'d globals.
//
// PHP derives these from dirname(__DIR__) — the source checkout's root directory — because a
// PHP install ships as source (or a Phar bundling that source) that gets interpreted in place.
// A compiled Go binary has no such "next to my own source" location; the natural equivalent is
// the directory containing the running executable, which is where a Go server's shipped
// resources (translations, Bedrock data tables, etc.) would live alongside the binary.
//
// TODO: once resource bundling is decided (a plain resources/ folder next to the binary vs.
// go:embed-ing them into the binary itself), these may need to change from filesystem paths to
// an embed.FS-backed lookup.
var (
	Path                          string
	ResourcePath                  string
	BedrockDataPath               string
	LocaleDataPath                string
	BedrockBlockUpgradeSchemaPath string
	BedrockItemUpgradeSchemaPath  string
)

func init() {
	exe, err := os.Executable()
	if err != nil {
		exe = os.Args[0]
	}
	dir, err := filepath.Abs(filepath.Dir(exe))
	if err != nil {
		dir = "."
	}

	Path = dir + string(filepath.Separator)
	ResourcePath = filepath.Join(dir, "resources") + string(filepath.Separator)
	BedrockDataPath = filepath.Join(dir, "data", "bedrock-data") + string(filepath.Separator)
	LocaleDataPath = filepath.Join(dir, "resources", "translations") + string(filepath.Separator)
	BedrockBlockUpgradeSchemaPath = filepath.Join(dir, "data", "bedrock-block-upgrade-schema") + string(filepath.Separator)
	BedrockItemUpgradeSchemaPath = filepath.Join(dir, "data", "bedrock-item-upgrade-schema") + string(filepath.Separator)
}
