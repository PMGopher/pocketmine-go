package projectile

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/math"
)

// Throwable is a port of the abstract pocketmine\entity\projectile\Throwable.
type Throwable struct {
	Projectile
}

func (t *Throwable) GetInitialSizeInfo() entity.EntitySizeInfo {
	return entity.NewEntitySizeInfo(0.25, 0.25)
}

func (t *Throwable) GetInitialDragMultiplier() float64 { return 0.01 }

func (t *Throwable) GetInitialGravity() float64 { return 0.03 }

// OnHitBlock is a port of Throwable::onHitBlock: throwables break on hitting a block.
func (t *Throwable) OnHitBlock(blockHit block.Behavior, hitResult math.RayTraceResult) {
	t.Projectile.OnHitBlock(blockHit, hitResult)
	t.FlagForDespawn()
}
