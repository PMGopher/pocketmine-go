package object

import (
	stdmath "math"

	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/entity/animation"
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/world/sound"
)

// fireworkTagFireworkData/fireworkTagExplosions mirror FireworkRocket (item)::TAG_FIREWORK_DATA/
// TAG_EXPLOSIONS.
const (
	fireworkTagFireworkData = "Fireworks"
	fireworkTagExplosions   = "Explosions"
)

// FireworkRocket is a port of pocketmine\entity\object\FireworkRocket (also Explosive and
// NeverSavedWithChunkEntity).
type FireworkRocket struct {
	entity.Entity

	maxFlightTimeTicks int
	explosions         []item.FireworkRocketExplosion
}

// NewFireworkRocket is a port of FireworkRocket::__construct (panicking on negative flight time).
func NewFireworkRocket(location entity.Location, maxFlightTimeTicks int, explosions []item.FireworkRocketExplosion, tag *nbt.CompoundTag) *FireworkRocket {
	if maxFlightTimeTicks < 0 {
		panic("Life ticks cannot be negative")
	}
	f := &FireworkRocket{maxFlightTimeTicks: maxFlightTimeTicks}
	f.SetExplosions(explosions)
	f.Construct(f, location, tag)
	return f
}

// NeverSavedWithChunk marks FireworkRocket as a NeverSavedWithChunkEntity.
func (f *FireworkRocket) NeverSavedWithChunk() {}

func (f *FireworkRocket) GetNetworkTypeID() string { return entity.EntityIDFireworksRocket }

func (f *FireworkRocket) GetInitialSizeInfo() entity.EntitySizeInfo {
	return entity.NewEntitySizeInfo(0.25, 0.25)
}

func (f *FireworkRocket) GetInitialDragMultiplier() float64 { return 0.0 }

func (f *FireworkRocket) GetInitialGravity() float64 { return 0.0 }

// GetMaxFlightTimeTicks returns the maximum number of ticks this will fly for, before exploding.
func (f *FireworkRocket) GetMaxFlightTimeTicks() int { return f.maxFlightTimeTicks }

// SetMaxFlightTimeTicks panics on a negative value.
func (f *FireworkRocket) SetMaxFlightTimeTicks(maxFlightTimeTicks int) *FireworkRocket {
	if maxFlightTimeTicks < 0 {
		panic("Max flight time ticks cannot be negative")
	}
	f.maxFlightTimeTicks = maxFlightTimeTicks
	return f
}

func (f *FireworkRocket) GetExplosions() []item.FireworkRocketExplosion { return f.explosions }

func (f *FireworkRocket) SetExplosions(explosions []item.FireworkRocketExplosion) *FireworkRocket {
	f.explosions = explosions
	return f
}

// OnFirstUpdate is a port of FireworkRocket::onFirstUpdate.
func (f *FireworkRocket) OnFirstUpdate(currentTick int64) {
	f.Entity.OnFirstUpdate(currentTick)
	f.BroadcastSound(sound.FireworkLaunchSound{})
}

// EntityBaseTick is a port of FireworkRocket::entityBaseTick.
func (f *FireworkRocket) EntityBaseTick(tickDiff int) bool {
	hasUpdate := f.Entity.EntityBaseTick(tickDiff)

	if !f.IsFlaggedForDespawn() {
		//Don't keep accelerating long-lived fireworks - this gets very rapidly out of control and makes the server
		//die. Vanilla fireworks will only live for about 52 ticks maximum anyway, so this only makes sure plugin
		//created fireworks don't murder the server
		if f.TicksLived < 60 {
			motion := f.GetMotion()
			f.AddMotion(motion.X*0.15, 0.04, motion.Z*0.15)
		}

		if f.TicksLived >= f.maxFlightTimeTicks {
			f.FlagForDespawn()
			f.Explode()
		}
	}

	return hasUpdate
}

// interceptCalculator is the promoted-from-*block.Block surface explosion raytracing needs.
type interceptCalculator interface {
	CalculateIntercept(pos1, pos2 math.Vector3) (math.RayTraceResult, bool)
}

// Explode is a port of FireworkRocket::explode.
func (f *FireworkRocket) Explode() {
	explosionCount := len(f.explosions)
	if explosionCount == 0 {
		return
	}
	f.BroadcastAnimation(animation.FireworkParticlesAnimation{Entity: f}, nil)
	for _, explosion := range f.explosions {
		f.BroadcastSound(explosion.GetType().GetExplosionSound())
		if explosion.WillTwinkle() {
			f.BroadcastSound(sound.FireworkCrackleSound{})
		}
	}

	force := float64(explosionCount*2) + 5
	w := f.GetWorld()
	origin := f.GetPosition()
	for _, e := range w.GetCollidingEntities(f.GetBoundingBox().ExpandedCopy(5, 5, 5), f) {
		living, ok := e.(livingTarget)
		if !ok {
			continue
		}

		position := living.GetPosition()
		distance := position.DistanceSquared(origin)
		if distance > 25 {
			continue
		}

		//cast two rays - one to the entity's feet and another to halfway up its body (according to Java, anyway)
		//this seems like it'd miss some cases but who am I to argue with vanilla logic :>
		height := living.GetBoundingBox().GetYLength()
	rays:
		for i := 0; i < 2; i++ {
			target := position.Add(0, 0.5*float64(i)*height, 0)
			seq, err := math.BetweenPoints(origin, target)
			if err != nil {
				continue
			}
			for blockPos := range seq {
				if calc, ok := w.GetBlock(blockPos).(interceptCalculator); ok {
					if _, hit := calc.CalculateIntercept(origin, target); hit {
						continue rays //obstruction, try another path
					}
				}
			}

			//no obstruction
			damage := force * stdmath.Sqrt((5-position.Distance(origin))/5)
			ev := entityevent.NewEntityDamageByEntityEvent(f, living, entityevent.CauseEntityExplosion, damage, nil)
			living.Attack(ev)
			break
		}
	}
}

func (f *FireworkRocket) CanBeCollidedWith() bool { return false }

// SyncNetworkData is a port of FireworkRocket::syncNetworkData.
func (f *FireworkRocket) SyncNetworkData(properties *entity.MetadataCollection) {
	f.Entity.SyncNetworkData(properties)

	values := make([]nbt.Tag, 0, len(f.explosions))
	for _, explosion := range f.explosions {
		values = append(values, explosion.ToCompoundTag())
	}
	explosions, err := nbt.NewListTag(values, nbt.TagCompound)
	if err != nil {
		panic(err)
	}
	fireworksData := nbt.NewCompoundTag().
		SetTag(fireworkTagFireworkData, nbt.NewCompoundTag().
			SetTag(fireworkTagExplosions, explosions))

	properties.SetCompoundTag(entity.MetadataFireworkItem, convert.NbtToMap(fireworksData))
}
