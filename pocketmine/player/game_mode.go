// Package player is a port of pocketmine\player: the Player entity and the classes around it
// (game modes, player info, player data storage, chunk selection, survival block breaking).
package player

import (
	"strings"

	"pocketmine-go/pocketmine/lang"
)

// GameMode is a port of pocketmine\player\GameMode.
type GameMode int

const (
	GameModeSurvival GameMode = iota
	GameModeCreative
	GameModeAdventure
	GameModeSpectator
)

// gameModeMetadata mirrors GameMode::getMetadata's match expression: the English name, the
// translatable name and the command-friendly aliases of each case.
var gameModeMetadata = map[GameMode]struct {
	name         string
	translatable func() *lang.Translatable
	aliases      []string
}{
	GameModeSurvival:  {"Survival", lang.KnownTranslationFactory.GameModeSurvival, []string{"survival", "s", "0"}},
	GameModeCreative:  {"Creative", lang.KnownTranslationFactory.GameModeCreative, []string{"creative", "c", "1"}},
	GameModeAdventure: {"Adventure", lang.KnownTranslationFactory.GameModeAdventure, []string{"adventure", "a", "2"}},
	GameModeSpectator: {"Spectator", lang.KnownTranslationFactory.GameModeSpectator, []string{"spectator", "v", "view", "3"}},
}

// GetTranslatableName is a port of GameMode::getTranslatableName.
func (g GameMode) GetTranslatableName() *lang.Translatable { return gameModeMetadata[g].translatable() }

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

// Name is the enum case name (PHP's $gameMode->name), e.g. "SURVIVAL", which server.properties
// stores.
func (g GameMode) Name() string {
	switch g {
	case GameModeCreative:
		return "CREATIVE"
	case GameModeAdventure:
		return "ADVENTURE"
	case GameModeSpectator:
		return "SPECTATOR"
	}
	return "SURVIVAL"
}
