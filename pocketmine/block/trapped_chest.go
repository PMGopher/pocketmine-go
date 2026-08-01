package block

import (
	"pocketmine-go/pocketmine/block/tile"
	"pocketmine-go/pocketmine/math"
)

// TrappedChest is a port of pocketmine\block\TrappedChest. The PHP original is a near-empty
// subclass of Chest (redstone signal emission is itself unimplemented upstream, marked
// "//TODO: Redstone!").
//
// OnPostPlace is re-implemented here (duplicating Chest.OnPostPlace) rather than inherited via
// embedding, because its neighbor-pairing logic does a same-type check via a concrete *Chest type
// assertion - Go's structural embedding wouldn't let that inherited logic recognize *TrappedChest
// neighbors as valid pairing partners, unlike PHP's instanceof which follows the class hierarchy.
type TrappedChest struct {
	Chest
}

func NewTrappedChest(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *TrappedChest {
	t := &TrappedChest{
		Chest: Chest{
			Transparent:               Transparent{NewBlock(idInfo, name, typeInfo)},
			HorizontalFacingComponent: NewHorizontalFacingComponent(),
		},
	}
	t.Init(t)
	return t
}

func (t *TrappedChest) Clone() Behavior {
	c := *t
	c.rebind(&c)
	return &c
}

// OnPostPlace duplicates Chest.OnPostPlace's pairing logic, matched against *TrappedChest
// neighbors instead of *Chest (see type doc comment).
func (t *TrappedChest) OnPostPlace() {
	world, err := t.position.GetWorld()
	if err != nil {
		return
	}
	tt, ok := world.GetTile(t.position)
	if !ok {
		return
	}
	tileChest, ok := tt.(*tile.Chest)
	if !ok {
		return
	}

	for _, clockwise := range [2]bool{false, true} {
		side := math.RotateY(t.Facing, clockwise)
		neighbor := t.self.(blockGeometry).GetSide(side, 1)
		other, ok := neighbor.(*TrappedChest)
		if !ok || !other.HasSameTypeId(t.self) || other.Facing != t.Facing {
			continue
		}
		pairTileRaw, ok := world.GetTile(other.position)
		if !ok {
			continue
		}
		pairTile, ok := pairTileRaw.(*tile.Chest)
		if !ok || pairTile.IsPaired() {
			continue
		}
		pairTile.PairWith(tileChest)
		break
	}
}
