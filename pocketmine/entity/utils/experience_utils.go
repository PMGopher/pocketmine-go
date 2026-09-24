// Package utils is a port of pocketmine\entity\utils.
package utils

import (
	"errors"
	"slices"

	"pocketmine-go/pocketmine/math"
)

// GetXpToReachLevel is a port of ExperienceUtils::getXpToReachLevel: the total amount of XP
// required to reach the specified level.
func GetXpToReachLevel(level int) int {
	if level <= 16 {
		return level*level + level*6
	} else if level <= 31 {
		return int(float64(level*level)*2.5 - 40.5*float64(level) + 360)
	}

	return int(float64(level*level)*4.5 - 162.5*float64(level) + 2220)
}

// GetXpToCompleteLevel is a port of ExperienceUtils::getXpToCompleteLevel: the amount of XP needed
// to get from the given level to the next.
func GetXpToCompleteLevel(level int) int {
	if level <= 15 {
		return 2*level + 7
	} else if level <= 30 {
		return 5*level - 38
	}
	return 9*level - 158
}

// GetLevelFromXp is a port of ExperienceUtils::getLevelFromXp: the level (with fractional progress)
// for a total amount of XP. Errors for negative XP (PHP's InvalidArgumentException).
func GetLevelFromXp(xp int) (float64, error) {
	if xp < 0 {
		return 0, errors.New("XP must be at least 0")
	}

	var a, b, c float64
	if xp <= GetXpToReachLevel(16) {
		a, b, c = 1, 6, 0
	} else if xp <= GetXpToReachLevel(31) {
		a, b, c = 2.5, -40.5, 360
	} else {
		a, b, c = 4.5, -162.5, 2220
	}

	x, err := math.SolveQuadratic(a, b, c-float64(xp))
	if err != nil {
		return 0, err
	}
	if len(x) == 0 {
		panic("Expected at least 1 solution")
	}

	return slices.Max(x), nil //we're only interested in the positive solution
}
