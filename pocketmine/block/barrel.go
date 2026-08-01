package block

import (
	stdmath "math"

	"pocketmine-go/pocketmine/block/tile"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// Barrel is a port of pocketmine\block\Barrel.
type Barrel struct {
	Opaque
	FacingComponent

	Open bool
}

func NewBarrel(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Barrel {
	b := &Barrel{
		Opaque:          Opaque{NewBlock(idInfo, name, typeInfo)},
		FacingComponent: NewFacingComponent(),
	}
	b.Init(b)
	return b
}

func (b *Barrel) Clone() Behavior {
	c := *b
	c.rebind(&c)
	return &c
}

func (b *Barrel) DescribeBlockOnlyState(w runtime.DataDescriber) {
	b.DescribeFacing(w)
	w.Bool(&b.Open)
}

func (b *Barrel) IsOpen() bool { return b.Open }

func (b *Barrel) SetOpen(open bool) { b.Open = open }

// Place is a port of Barrel::place: faces towards/away from the player based on where they're
// standing and looking, rather than just the opposite of their horizontal facing.
func (b *Barrel) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if player != nil {
		playerPos := player.GetPosition()
		if stdmath.Abs(playerPos.X-b.position.X) < 2 && stdmath.Abs(playerPos.Z-b.position.Z) < 2 {
			y := player.GetEyePos().Y
			switch {
			case y-b.position.Y > 2:
				b.Facing = math.Up
			case b.position.Y-y > 0:
				b.Facing = math.Down
			default:
				b.Facing = math.Opposite(player.GetHorizontalFacing())
			}
		} else {
			b.Facing = math.Opposite(player.GetHorizontalFacing())
		}
	}
	return b.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

// OnInteract is a port of Barrel::onInteract, minus actually opening the inventory window
// (player.SetCurrentWindow isn't ported - see block.Chest.OnInteract's doc comment for the same
// gap). The CanOpenWith lock check that would gate it is fully real.
func (b *Barrel) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	if player == nil {
		return true
	}
	world, err := b.position.GetWorld()
	if err != nil {
		return true
	}
	t, ok := world.GetTile(b.position)
	if !ok {
		return true
	}
	tileBarrel, ok := t.(*tile.Barrel)
	if !ok {
		return true
	}
	if !tileBarrel.CanOpenWith(item.GetCustomName()) {
		return true
	}
	// player.SetCurrentWindow(tileBarrel.GetInventory()) - not ported, see doc comment above.
	return true
}

func (b *Barrel) GetFuelTime() int { return 300 }
