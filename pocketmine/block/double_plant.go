package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// DoublePlant is a port of pocketmine\block\DoublePlant.
type DoublePlant struct {
	Flowable

	Top bool
}

func NewDoublePlant(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *DoublePlant {
	d := &DoublePlant{Flowable: Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}}
	d.Init(d)
	return d
}

func (d *DoublePlant) Clone() Behavior {
	c := *d
	c.rebind(&c)
	return &c
}

func (d *DoublePlant) DescribeBlockOnlyState(w runtime.DataDescriber) { w.Bool(&d.Top) }

func (d *DoublePlant) IsTop() bool { return d.Top }

func (d *DoublePlant) SetTop(top bool) { d.Top = top }

func (d *DoublePlant) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	down := blockReplace.(blockGeometry).GetSide(math.Down, 1)
	up := blockReplace.(blockGeometry).GetSide(math.Up, 1)
	if down.(blockGeometry).HasTypeTag(BlockTypeTagsDirt) && up.CanBeReplaced() {
		top := d.Clone().(*DoublePlant)
		top.Top = true

		replacePos := blockReplace.GetPosition()
		tx.AddBlock(replacePos, d.self)
		tx.AddBlock(replacePos.GetSide(math.Up, 1), top)
		return true
	}
	return false
}

// IsValidHalfPlant returns whether this double-plant has a corresponding other half.
func (d *DoublePlant) IsValidHalfPlant() bool {
	otherSide := math.Up
	if d.Top {
		otherSide = math.Down
	}
	other, ok := d.GetSide(otherSide, 1).(*DoublePlant)
	return ok && other.HasSameTypeId(d.self) && other.Top != d.Top
}

func (d *DoublePlant) OnNearbyBlockChange() {
	down := d.GetSide(math.Down, 1).(blockGeometry)
	if !d.IsValidHalfPlant() || (!d.Top && !down.HasTypeTag(BlockTypeTagsDirt) && !down.HasTypeTag(BlockTypeTagsMud)) {
		if world, err := d.position.GetWorld(); err == nil {
			world.UseBreakOn(d.position.AsVector3())
		}
	}
}

func (d *DoublePlant) GetDrops(item Item) []Item {
	if d.Top {
		return d.Block.GetDrops(item)
	}
	return nil
}

func (d *DoublePlant) GetAffectedBlocks() []Behavior {
	if d.IsValidHalfPlant() {
		otherSide := math.Up
		if d.Top {
			otherSide = math.Down
		}
		return []Behavior{d.self, d.GetSide(otherSide, 1)}
	}
	return d.Block.GetAffectedBlocks()
}

func (d *DoublePlant) GetFlameEncouragement() int { return 60 }

func (d *DoublePlant) GetFlammability() int { return 100 }
