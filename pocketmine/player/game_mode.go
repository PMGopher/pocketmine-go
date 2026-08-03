// Package player is a port of a slice of pocketmine\player - see Player's own doc comment for the
// large parts deliberately not attempted in this first pass.
package player

import "strings"

// GameMode is a port of pocketmine\player\GameMode.
type GameMode int

const (
	GameModeSurvival GameMode = iota
	GameModeCreative
	GameModeAdventure
	GameModeSpectator
)

// gameModeMetadata mirrors GameMode::getMetadata's match expression - the English name and
// command-friendly aliases for each case (the Translatable is left out: this port has no
// translation-key infrastructure for it to return yet, matching lang package's own current scope).
var gameModeMetadata = map[GameMode]struct {
	name    string
	aliases []string
}{
	GameModeSurvival:  {"Survival", []string{"survival", "s", "0"}},
	GameModeCreative:  {"Creative", []string{"creative", "c", "1"}},
	GameModeAdventure: {"Adventure", []string{"adventure", "a", "2"}},
	GameModeSpectator: {"Spectator", []string{"spectator", "v", "view", "3"}},
}

// GetEnglishName is a port of GameMode::getEnglishName.
func (g GameMode) GetEnglishName() string { return gameModeMetadata[g].name }

// GetAliases is a port of GameMode::getAliases.
func (g GameMode) GetAliases() []string { return gameModeMetadata[g].aliases }

// GameModeFromString is a port of GameMode::fromString.
func GameModeFromString(str string) (GameMode, bool) {
	lower := strings.ToLower(str)
	for _, g := range [...]GameMode{GameModeSurvival, GameModeCreative, GameModeAdventure, GameModeSpectator} {
		for _, alias := range gameModeMetadata[g].aliases {
			if alias == lower {
				return g, true
			}
		}
	}
	return 0, false
}
