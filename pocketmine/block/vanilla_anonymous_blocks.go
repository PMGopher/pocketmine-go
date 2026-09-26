package block

import (
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/sound"
)

// The anonymous block subclasses VanillaBlocksInputs.php declares inline (`new class(...) extends
// Opaque{...}`). Stone is in stone.go.

// ReinforcedDeepslate is the anonymous Opaque subclass registered for "reinforced_deepslate": it
// drops nothing.
type ReinforcedDeepslate struct {
	Opaque
}

func NewReinforcedDeepslate(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *ReinforcedDeepslate {
	r := &ReinforcedDeepslate{Opaque{NewBlock(idInfo, name, typeInfo)}}
	r.Init(r)
	return r
}

func (r *ReinforcedDeepslate) Clone() Behavior {
	c := *r
	c.rebind(&c)
	return &c
}

func (r *ReinforcedDeepslate) GetDropsForCompatibleTool(item Item) []Item { return nil }

// fireProofOpaque is the anonymous Opaque subclass registered for "ancient_debris" and
// "netherite": its item doesn't burn.
type fireProofOpaque struct {
	Opaque
}

func newFireProofOpaque(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) Behavior {
	f := &fireProofOpaque{Opaque{NewBlock(idInfo, name, typeInfo)}}
	f.Init(f)
	return f
}

func (f *fireProofOpaque) Clone() Behavior {
	c := *f
	c.rebind(&c)
	return &c
}

func (f *fireProofOpaque) IsFireProofAsItem() bool { return true }

// lightOpaque is the anonymous Opaque subclass registered for "shroomlight" (15) and
// "crying_obsidian" (10): it emits light.
type lightOpaque struct {
	Opaque
	lightLevel int
}

func newLightOpaque(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo, lightLevel int) Behavior {
	l := &lightOpaque{Opaque: Opaque{NewBlock(idInfo, name, typeInfo)}, lightLevel: lightLevel}
	l.Init(l)
	return l
}

func (l *lightOpaque) Clone() Behavior {
	c := *l
	c.rebind(&c)
	return &c
}

func (l *lightOpaque) GetLightLevel() int { return l.lightLevel }

// Amethyst is the anonymous Opaque subclass registered for "amethyst" (AmethystTrait).
type Amethyst struct {
	Opaque
}

func NewAmethyst(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Amethyst {
	a := &Amethyst{Opaque{NewBlock(idInfo, name, typeInfo)}}
	a.Init(a)
	return a
}

func (a *Amethyst) Clone() Behavior {
	c := *a
	c.rebind(&c)
	return &c
}

// OnProjectileHit is a port of AmethystTrait::onProjectileHit.
func (a *Amethyst) OnProjectileHit(projectile Projectile, hitResult math.RayTraceResult) {
	world, err := a.position.GetWorld()
	if err != nil {
		return
	}
	world.AddSound(a.position.AsVector3(), sound.AmethystBlockChimeSound{})
	world.AddSound(a.position.AsVector3(), sound.BlockPunchSound{BlockStateID: a.GetStateId()})
}

// Deepslate is the anonymous SimplePillar subclass registered for "deepslate": it drops cobbled
// deepslate unless mined with silk touch.
type Deepslate struct {
	SimplePillar
}

func NewDeepslate(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Deepslate {
	d := &Deepslate{SimplePillar: *NewSimplePillar(idInfo, name, typeInfo)}
	d.Init(d)
	return d
}

func (d *Deepslate) Clone() Behavior {
	c := *d
	c.rebind(&c)
	return &c
}

func (d *Deepslate) GetDropsForCompatibleTool(item Item) []Item {
	dropped := asItemOrNil(VanillaBlock("cobbled_deepslate"))
	if dropped == nil {
		return nil
	}
	return []Item{dropped}
}

func (d *Deepslate) IsAffectedBySilkTouch() bool { return true }
