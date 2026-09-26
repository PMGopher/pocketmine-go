package region

import (
	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/world/format/io"
)

// init registers the region providers with io's WorldProviderManager (PHP's constructor does it
// directly), after LevelDB: anvil, mcregion, pmanvil.
func init() {
	io.RegisterBuiltinProvider("anvil", io.NewReadOnlyWorldProviderManagerEntry(IsValidAnvil, func(path string, logger log.Logger) (io.WorldProvider, error) {
		return NewAnvil(path, logger)
	}), false, 1)
	io.RegisterBuiltinProvider("mcregion", io.NewReadOnlyWorldProviderManagerEntry(IsValidMcRegion, func(path string, logger log.Logger) (io.WorldProvider, error) {
		return NewMcRegion(path, logger)
	}), false, 2)
	io.RegisterBuiltinProvider("pmanvil", io.NewReadOnlyWorldProviderManagerEntry(IsValidPMAnvil, func(path string, logger log.Logger) (io.WorldProvider, error) {
		return NewPMAnvil(path, logger)
	}), false, 3)
}
