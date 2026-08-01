package item

import "pocketmine-go/pocketmine/nbt"

const (
	fireworkRocketTagFireworkData         = "Fireworks"
	fireworkRocketTagFlightTimeMultiplier = "Flight"
	fireworkRocketTagExplosions           = "Explosions"
)

// FireworkRocket is a port of pocketmine\item\FireworkRocket. OnInteractBlock (spawning the
// firework entity) needs a real Player/Block/World - see the Item interface's doc comment.
type FireworkRocket struct {
	ItemBase

	FlightTimeMultiplier int
	Explosions           []FireworkRocketExplosion
}

func NewFireworkRocket(identifier ItemIdentifier, name string) *FireworkRocket {
	f := &FireworkRocket{FlightTimeMultiplier: 1}
	f.Init(f, identifier, name)
	return f
}

func (f *FireworkRocket) Clone() Item {
	c := *f
	c.rebind(&c)
	return &c
}

func (f *FireworkRocket) GetFlightTimeMultiplier() int { return f.FlightTimeMultiplier }

// SetFlightTimeMultiplier panics if multiplier is out of range, mirroring the PHP original's
// InvalidArgumentException (a programmer error at the call site).
func (f *FireworkRocket) SetFlightTimeMultiplier(multiplier int) {
	if multiplier < 1 || multiplier > 127 {
		panic("Flight time multiplier must be in range 1-127")
	}
	f.FlightTimeMultiplier = multiplier
}

func (f *FireworkRocket) GetExplosions() []FireworkRocketExplosion { return f.Explosions }

func (f *FireworkRocket) SetExplosions(explosions []FireworkRocketExplosion) {
	f.Explosions = explosions
}

// deserializeCompoundTag/serializeCompoundTag extend ItemBase's own pair, the same self-dispatch
// participation described on Durable's.
func (f *FireworkRocket) deserializeCompoundTag(tag *nbt.CompoundTag) {
	f.ItemBase.deserializeCompoundTag(tag)

	fireworkData, ok, _ := tag.GetCompoundTag(fireworkRocketTagFireworkData)
	if !ok {
		return
	}

	f.SetFlightTimeMultiplier(int(fireworkData.GetByteOr(fireworkRocketTagFlightTimeMultiplier, 1)))

	f.Explosions = nil
	if explosions, ok, _ := fireworkData.GetListTag(fireworkRocketTagExplosions); ok {
		for _, v := range explosions.Values() {
			explosionTag, ok := v.(*nbt.CompoundTag)
			if !ok {
				continue
			}
			if explosion, err := FireworkRocketExplosionFromCompoundTag(explosionTag); err == nil {
				f.Explosions = append(f.Explosions, explosion)
			}
		}
	}
}

func (f *FireworkRocket) serializeCompoundTag(tag *nbt.CompoundTag) {
	f.ItemBase.serializeCompoundTag(tag)

	fireworkData := nbt.NewCompoundTag()
	fireworkData.SetByte(fireworkRocketTagFlightTimeMultiplier, nbt.ByteTag(f.FlightTimeMultiplier))

	values := make([]nbt.Tag, len(f.Explosions))
	for i, e := range f.Explosions {
		values[i] = e.ToCompoundTag()
	}
	explosionsTag, err := nbt.NewListTag(values, nbt.TagCompound)
	if err != nil {
		panic(err)
	}
	fireworkData.SetTag(fireworkRocketTagExplosions, explosionsTag)

	tag.SetTag(fireworkRocketTagFireworkData, fireworkData)
}
