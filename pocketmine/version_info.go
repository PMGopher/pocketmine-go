package pocketmine

import (
	"strconv"
	"sync"

	"pocketmine-go/pocketmine/utils"
)

// VersionInfo is a port of pocketmine\VersionInfo.
const (
	Name               = "PocketMine-MP"
	BaseVersion        = "5.44.4"
	IsDevelopmentBuild = true
	BuildChannel       = "stable"
	GithubURL          = "https://github.com/pmmp/PocketMine-MP"

	// WorldDataVersion is PocketMine-MP-specific version ID for world data, used to determine
	// what fixes need to be applied to old world data. This supplements the Minecraft vanilla
	// world version, and should be bumped for any non-Mojang BC-breaking change to save data.
	WorldDataVersion = 1
	// TagWorldDataVersion names the NBT tag (a TAG_Long) used to store WorldDataVersion.
	TagWorldDataVersion = "PMMPDataVersion"
)

// GitHashOverride and BuildNumberOverride let a release build stamp in its provenance at compile
// time, e.g.:
//
//	go build -ldflags "-X 'pocketmine-go/pocketmine.GitHashOverride=<hash>' -X 'pocketmine-go/pocketmine.BuildNumberOverride=42'"
//
// This is the direct equivalent of PHP embedding {"git": ..., "build": ...} into a release Phar's
// metadata: a compiled Go binary has no metadata container to read at runtime the way a Phar
// does, so the info has to be baked in at build time instead. Without an override, GitHash falls
// back to inspecting the git repository at Path (the dev-build case, mirroring PHP's own
// git-checkout fallback), and BuildNumber defaults to 0.
var (
	GitHashOverride     string
	BuildNumberOverride string
)

var (
	gitHashOnce sync.Once
	gitHash     string

	buildNumberOnce sync.Once
	buildNumberVal  int

	fullVersionOnce sync.Once
	fullVersionVal  *utils.VersionString
)

func GitHash() string {
	gitHashOnce.Do(func() {
		if GitHashOverride != "" {
			gitHash = GitHashOverride
			return
		}
		gitHash = utils.RepositoryStatePretty(Path)
	})
	return gitHash
}

func BuildNumber() int {
	buildNumberOnce.Do(func() {
		if n, err := strconv.Atoi(BuildNumberOverride); err == nil {
			buildNumberVal = n
		}
	})
	return buildNumberVal
}

// Version returns the full VersionString for this build.
func Version() *utils.VersionString {
	fullVersionOnce.Do(func() {
		v, err := utils.NewVersionString(BaseVersion, IsDevelopmentBuild, BuildNumber())
		if err != nil {
			// BaseVersion is a compile-time constant in this file, not user input — a failure here
			// means this file itself has an invalid version string, which is a bug, not a runtime condition.
			panic(err)
		}
		fullVersionVal = v
	})
	return fullVersionVal
}
