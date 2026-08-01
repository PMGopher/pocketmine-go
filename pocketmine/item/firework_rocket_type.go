package item

import "pocketmine-go/pocketmine/world/sound"

// FireworkRocketType is a port of pocketmine\item\FireworkRocketType.
type FireworkRocketType int

const (
	FireworkRocketTypeSmallBall FireworkRocketType = iota
	FireworkRocketTypeLargeBall
	FireworkRocketTypeStar
	FireworkRocketTypeCreeper
	FireworkRocketTypeBurst
)

// GetExplosionSound is a port of FireworkRocketType::getExplosionSound.
func (t FireworkRocketType) GetExplosionSound() sound.Sound {
	if t == FireworkRocketTypeLargeBall {
		return sound.FireworkLargeExplosionSound{}
	}
	return sound.FireworkExplosionSound{}
}
