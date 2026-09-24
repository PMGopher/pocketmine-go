// Package projectile is a port of pocketmine\entity\projectile.
package projectile

import (
	stdmath "math"

	"github.com/go-gl/mathgl/mgl32"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/data"
	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/entity/object"
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/world"
)

// Projectile NBT keys, a port of Projectile's TAG_* constants.
const (
	tagStuckOnBlockPos = "StuckToBlockPos"
	tagDamage          = "damage" //TAG_Double
	tagTileX           = "tileX"  //TAG_Int
	tagTileY           = "tileY"  //TAG_Int
	tagTileZ           = "tileZ"  //TAG_Int
)

// ProjectileSource is a port of pocketmine\entity\projectile\ProjectileSource: a marker for
// entities which can shoot projectiles (Human).
type ProjectileSource interface{}

// projectileShaper is the self-dispatch surface: the Projectile methods Projectile's own bodies
// call on $this that subtypes override.
type projectileShaper interface {
	entity.Hooks

	GetHorizontalFacing() math.Facing
	GetResultDamage() int
	OnHit(event entityevent.ProjectileHit)
	OnHitEntity(entityHit world.Entity, hitResult math.RayTraceResult)
	OnHitBlock(blockHit block.Behavior, hitResult math.RayTraceResult)
	DespawnsOnEntityHit() bool
	CalculateInterceptWithBlock(blk block.Behavior, start, end math.Vector3) (math.RayTraceResult, bool)
}

// Projectile is a port of the abstract pocketmine\entity\projectile\Projectile.
type Projectile struct {
	entity.Entity

	pself projectileShaper

	damage float64

	// blockHit is the position of the block the projectile is stuck in, or nil.
	blockHit *math.Vector3
}

// ConstructProjectile is a port of Projectile::__construct: concrete projectiles set their own
// fields, then call this with themselves as self. shootingEntity may be nil.
func (p *Projectile) ConstructProjectile(self projectileShaper, location entity.Location, shootingEntity world.Entity, tag *nbt.CompoundTag) {
	p.pself = self
	p.Construct(self, location, tag)
	if shootingEntity != nil {
		p.SetOwningEntity(shootingEntity)
	}
}

// Attack is a port of Projectile::attack: only the void can damage a projectile.
func (p *Projectile) Attack(source entityevent.DamageSource) {
	if source.GetCause() == entityevent.CauseVoid {
		p.Entity.Attack(source)
	}
}

// InitEntity is a port of Projectile::initEntity.
func (p *Projectile) InitEntity(tag *nbt.CompoundTag) {
	p.Entity.InitEntity(tag)

	p.SetMaxHealth(1)
	p.SetHealth(1)
	p.damage = float64(tag.GetDoubleOr(tagDamage, nbt.DoubleTag(p.damage)))

	if stuck, ok, _ := tag.GetListTag(tagStuckOnBlockPos); ok && (stuck.GetTagType() == nbt.TagInt || stuck.Count() == 0) {
		values := stuck.Values()
		if len(values) != 3 {
			panic(data.NewSavedDataLoadingError(tagStuckOnBlockPos + " tag should be a list of 3 TAG_Int"))
		}
		blockHit := math.NewVector3(float64(values[0].(nbt.IntTag)), float64(values[1].(nbt.IntTag)), float64(values[2].(nbt.IntTag)))
		p.blockHit = &blockHit
	} else {
		tileX, xOK := tag.GetTag(tagTileX)
		tileY, yOK := tag.GetTag(tagTileY)
		tileZ, zOK := tag.GetTag(tagTileZ)
		if xOK && yOK && zOK {
			x, xInt := tileX.(nbt.IntTag)
			y, yInt := tileY.(nbt.IntTag)
			z, zInt := tileZ.(nbt.IntTag)
			if xInt && yInt && zInt {
				blockHit := math.NewVector3(float64(x), float64(y), float64(z))
				p.blockHit = &blockHit
			}
		}
	}
}

// livingEntity is the surface Projectile::canCollideWith needs to recognise a Living.
type livingEntity interface {
	IsLiving() bool
}

// CanCollideWith is a port of Projectile::canCollideWith: projectiles hit living entities and end
// crystals, while in flight.
func (p *Projectile) CanCollideWith(other world.Entity) bool {
	_, isLiving := other.(livingEntity)
	_, isEndCrystal := other.(*object.EndCrystal)
	return (isLiving || isEndCrystal) && !p.OnGround
}

func (p *Projectile) CanBeCollidedWith() bool { return false }

// GetBaseDamage returns the base damage applied on collision. This is multiplied by the
// projectile's speed to give a result damage.
func (p *Projectile) GetBaseDamage() float64 { return p.damage }

// SetBaseDamage sets the base amount of damage applied by the projectile.
func (p *Projectile) SetBaseDamage(damage float64) { p.damage = damage }

// GetResultDamage returns the amount of damage this projectile will deal to the entity it hits.
func (p *Projectile) GetResultDamage() int { return int(stdmath.Ceil(p.damage)) }

// GetBlockHit returns the position of the block the projectile is stuck in, or nil.
func (p *Projectile) GetBlockHit() *math.Vector3 { return p.blockHit }

// SaveNBT is a port of Projectile::saveNBT.
func (p *Projectile) SaveNBT() *nbt.CompoundTag {
	tag := p.Entity.SaveNBT()

	tag.SetDouble(tagDamage, nbt.DoubleTag(p.damage))

	if p.blockHit != nil {
		list, err := nbt.NewListTag([]nbt.Tag{
			nbt.IntTag(p.blockHit.FloorX()),
			nbt.IntTag(p.blockHit.FloorY()),
			nbt.IntTag(p.blockHit.FloorZ()),
		}, nbt.TagInt)
		if err != nil {
			panic(err)
		}
		tag.SetTag(tagStuckOnBlockPos, list)
	}

	return tag
}

func (p *Projectile) ApplyDragBeforeGravity() bool { return true }

// bbCollider is the promoted-from-*block.Block surface projectiles need.
type bbCollider interface {
	CollidesWithBB(bb math.AxisAlignedBB) bool
	CalculateIntercept(pos1, pos2 math.Vector3) (math.RayTraceResult, bool)
}

// OnNearbyBlockChange is a port of Projectile::onNearbyBlockChange: a stuck projectile falls once
// the block it's stuck in no longer holds it.
func (p *Projectile) OnNearbyBlockChange() {
	if p.blockHit != nil && p.GetWorld().IsInLoadedTerrain(*p.blockHit) {
		blockHit := p.GetWorld().GetBlock(*p.blockHit)
		if collider, ok := blockHit.(bbCollider); !ok || !collider.CollidesWithBB(p.GetBoundingBox().ExpandedCopy(0.001, 0.001, 0.001)) {
			p.blockHit = nil
		}
	}

	p.Entity.OnNearbyBlockChange()
}

// HasMovementUpdate is a port of Projectile::hasMovementUpdate: a stuck projectile doesn't move.
func (p *Projectile) HasMovementUpdate() bool {
	return p.blockHit == nil && p.Entity.HasMovementUpdate()
}

// Move is a port of Projectile::move: projectiles raytrace their path each tick instead of
// colliding like other entities, firing ProjectileHit*Event on whatever they hit first.
func (p *Projectile) Move(dx, dy, dz float64) {
	p.ClearBlocksAround()

	start := p.GetPosition()
	end := start.Add(dx, dy, dz)

	var hitEntity world.Entity
	var hitBlock block.Behavior
	var hitResult math.RayTraceResult
	hit := false

	w := p.GetWorld()
	if seq, err := math.BetweenPoints(start, end); err == nil {
		for v := range seq {
			blk := w.GetBlockAtIfLoaded(int(v.X), int(v.Y), int(v.Z))

			if blockHitResult, ok := p.pself.CalculateInterceptWithBlock(blk, start, end); ok {
				end = blockHitResult.HitVector
				hitBlock = blk
				hitResult = blockHitResult
				hit = true
				break
			}
		}
	}

	entityDistance := stdmath.MaxFloat64

	newDiff := end.SubtractVector(start)
	searchBB := p.BoundingBox.AddCoord(newDiff.X, newDiff.Y, newDiff.Z)
	searchBB.Expand(1, 1, 1)
	ownerID, hasOwner := p.GetOwningEntityID()
	for _, e := range w.GetCollidingEntities(searchBB, p.pself) {
		if hasOwner && e.GetID() == ownerID && p.TicksLived < 5 {
			continue
		}

		entityBB := e.GetBoundingBox().ExpandedCopy(0.3, 0.3, 0.3)
		entityHitResult, ok := entityBB.CalculateIntercept(start, end)
		if !ok {
			continue
		}

		distance := p.GetPosition().DistanceSquared(entityHitResult.HitVector)

		if distance < entityDistance {
			entityDistance = distance
			hitEntity = e
			hitBlock = nil
			hitResult = entityHitResult
			hit = true
			end = entityHitResult.HitVector
		}
	}

	location := p.GetLocation()
	p.SetLocationDirect(entity.LocationFromObject(end, location.World, location.Yaw, location.Pitch))
	p.RecalculateBoundingBox()

	if hit {
		var ev entityevent.ProjectileHit
		var specificHitFunc func()
		if hitEntity != nil {
			ev = entityevent.NewProjectileHitEntityEvent(p.pself, hitResult, hitEntity)
			entityHit := hitEntity
			specificHitFunc = func() { p.pself.OnHitEntity(entityHit, hitResult) }
		} else {
			ev = entityevent.NewProjectileHitBlockEvent(p.pself, hitResult, hitBlock)
			blockHit := hitBlock
			specificHitFunc = func() { p.pself.OnHitBlock(blockHit, hitResult) }
		}

		motionBeforeOnHit := p.GetMotion()
		ev.Call()
		p.pself.OnHit(ev)
		specificHitFunc()

		p.IsCollided = true
		p.OnGround = true
		if motionBeforeOnHit.Equals(p.GetMotion()) {
			p.SetMotionDirect(math.Vector3Zero())
		}
	} else {
		p.IsCollided = false
		p.OnGround = false
		p.blockHit = nil

		//recompute angles...
		motion := p.GetMotion()
		f := stdmath.Sqrt(motion.X*motion.X + motion.Z*motion.Z)
		p.SetRotation(
			stdmath.Atan2(motion.X, motion.Z)*180/stdmath.Pi,
			stdmath.Atan2(motion.Y, f)*180/stdmath.Pi,
		)
	}

	w.OnEntityMoved(p.pself)
	p.CheckBlockIntersections()
}

// CalculateInterceptWithBlock is called by Move to allow subclasses to override what happens when
// the projectile collides with a block.
func (p *Projectile) CalculateInterceptWithBlock(blk block.Behavior, start, end math.Vector3) (math.RayTraceResult, bool) {
	if collider, ok := blk.(bbCollider); ok {
		return collider.CalculateIntercept(start, end)
	}
	return math.RayTraceResult{}, false
}

// OnHit is called when the projectile hits something. Override this to perform non-target-specific
// effects when the projectile hits something.
func (p *Projectile) OnHit(event entityevent.ProjectileHit) {}

// OnHitEntity is a port of Projectile::onHitEntity: damages the entity hit.
func (p *Projectile) OnHitEntity(entityHit world.Entity, hitResult math.RayTraceResult) {
	damage := p.pself.GetResultDamage()

	if damage >= 0 {
		var ev entityevent.DamageSource
		if owner := p.GetOwningEntity(); owner == nil {
			ev = entityevent.NewEntityDamageByEntityEvent(p.pself, entityHit, entityevent.CauseProjectile, float64(damage), nil)
		} else {
			ev = entityevent.NewEntityDamageByChildEntityEvent(owner, p.pself, entityHit, entityevent.CauseProjectile, float64(damage), nil)
		}

		entityHit.Attack(ev)

		if p.IsOnFire() {
			combust := entityevent.NewEntityCombustByEntityEvent(p.pself, entityHit, 5)
			combust.Call()
			if !combust.IsCancelled() {
				entityHit.SetOnFire(combust.GetDuration())
			}
		}
	}

	if p.pself.DespawnsOnEntityHit() {
		p.FlagForDespawn()
	}
}

// OnHitBlock is a port of Projectile::onHitBlock: the projectile gets stuck in the block.
func (p *Projectile) OnHitBlock(blockHit block.Behavior, hitResult math.RayTraceResult) {
	pos := blockHit.GetPosition().AsVector3()
	p.blockHit = &pos
	blockHit.OnProjectileHit(p.pself, hitResult)
}

func (p *Projectile) DespawnsOnEntityHit() bool { return true }

func vec32(v math.Vector3) mgl32.Vec3 {
	return mgl32.Vec3{float32(v.X), float32(v.Y), float32(v.Z)}
}
