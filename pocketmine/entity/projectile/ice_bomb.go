package projectile

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/entity"
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/sound"
)

// IceBomb is a port of pocketmine\entity\projectile\IceBomb.
type IceBomb struct {
	Throwable
}

// NewIceBomb is a port of IceBomb::__construct (shootingEntity may be nil).
func NewIceBomb(location entity.Location, shootingEntity world.Entity, tag *nbt.CompoundTag) *IceBomb {
	i := &IceBomb{}
	i.ConstructProjectile(i, location, shootingEntity, tag)
	return i
}

func (i *IceBomb) GetNetworkTypeID() string { return entity.EntityIDIceBomb }

func (i *IceBomb) GetResultDamage() int { return -1 }

// CalculateInterceptWithBlock is a port of IceBomb::calculateInterceptWithBlock: ice bombs
// collide with water as if it were a full block.
func (i *IceBomb) CalculateInterceptWithBlock(blk block.Behavior, start, end math.Vector3) (math.RayTraceResult, bool) {
	if blk.GetTypeId() == block.WATER {
		pos := blk.GetPosition()

		return math.OneAABB().OffsetCopy(pos.X, pos.Y, pos.Z).CalculateIntercept(start, end)
	}

	return i.Throwable.CalculateInterceptWithBlock(blk, start, end)
}

// OnHit is a port of IceBomb::onHit.
func (i *IceBomb) OnHit(event entityevent.ProjectileHit) {
	w := i.GetWorld()
	pos := i.GetPosition()

	w.AddSound(pos, sound.IceBombHitSound{})
	itemBreak := itemBreakParticle(item.VanillaIceBomb())
	for n := 0; n < 6; n++ {
		w.AddParticle(pos, itemBreak)
	}
}

// OnHitBlock is a port of IceBomb::onHitBlock: water around the impact freezes.
func (i *IceBomb) OnHitBlock(blockHit block.Behavior, hitResult math.RayTraceResult) {
	i.Throwable.OnHitBlock(blockHit, hitResult)

	pos := blockHit.GetPosition()
	blockWorld, err := pos.GetWorld()
	if err != nil {
		return
	}
	w, ok := blockWorld.(*world.World)
	if !ok {
		return
	}
	posX := pos.FloorX()
	posY := pos.FloorY()
	posZ := pos.FloorZ()

	ice := block.VanillaIce()
	for x := posX - 1; x <= posX+1; x++ {
		for y := posY - 1; y <= posY+1; y++ {
			for z := posZ - 1; z <= posZ+1; z++ {
				if w.GetBlockAtIfLoaded(x, y, z).GetTypeId() == block.WATER {
					_ = w.SetBlockAt(x, y, z, ice)
				}
			}
		}
	}
}
