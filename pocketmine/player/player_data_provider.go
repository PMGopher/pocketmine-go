package player

import "pocketmine-go/pocketmine/nbt"

// PlayerDataProvider is a port of pocketmine\player\PlayerDataProvider. Implementations must
// treat player names case-insensitively.
type PlayerDataProvider interface {
	// HasData reports whether there is any data associated with the given player name.
	HasData(name string) bool
	// LoadData returns the data associated with the given player name, or (nil, nil) if there is
	// none (matching real PHP's ?CompoundTag null return) - a non-nil error means loading actually
	// failed (real PHP's own PlayerDataLoadException).
	LoadData(name string) (*nbt.CompoundTag, error)
	// SaveData saves data for the given player name.
	SaveData(name string, data *nbt.CompoundTag) error
}
