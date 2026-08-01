package block

import (
	"pocketmine-go/pocketmine/block/tile"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// Furnace is a port of pocketmine\block\Furnace.
type Furnace struct {
	Opaque
	HorizontalFacingComponent
	LightableComponent

	FurnaceType tile.FurnaceType
}

func NewFurnace(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo, furnaceType tile.FurnaceType) *Furnace {
	f := &Furnace{
		Opaque:                    Opaque{NewBlock(idInfo, name, typeInfo)},
		HorizontalFacingComponent: NewHorizontalFacingComponent(),
		FurnaceType:               furnaceType,
	}
	f.Init(f)
	return f
}

func (f *Furnace) Clone() Behavior {
	c := *f
	c.rebind(&c)
	return &c
}

func (f *Furnace) DescribeBlockOnlyState(w runtime.DataDescriber) {
	f.DescribeHorizontalFacing(w)
	f.DescribeLit(w)
}

// Place is a port of pocketmine\block\utils\FacesOppositePlacingPlayerTrait::place.
func (f *Furnace) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if player != nil {
		f.Facing = math.Opposite(player.GetHorizontalFacing())
	}
	return f.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

func (f *Furnace) GetFurnaceType() tile.FurnaceType { return f.FurnaceType }

func (f *Furnace) GetLightLevel() int {
	if f.Lit {
		return 13
	}
	return 0
}

// OnInteract is a port of Furnace::onInteract, minus actually opening the inventory window
// (player.SetCurrentWindow isn't ported - see block.Chest.OnInteract's doc comment for the same
// gap). The CanOpenWith lock check that would gate it is fully real.
func (f *Furnace) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	if player == nil {
		return true
	}
	world, err := f.position.GetWorld()
	if err != nil {
		return true
	}
	t, ok := world.GetTile(f.position)
	if !ok {
		return true
	}
	tileFurnace, ok := t.(*tile.Furnace)
	if !ok {
		return true
	}
	if tileFurnace.CanOpenWith(item.GetCustomName()) {
		// player.SetCurrentWindow(tileFurnace.GetInventory()) - not ported, see doc comment above.
	}
	return true
}

// OnScheduledUpdate should drive the furnace's smelting simulation (tile.Furnace.OnUpdate, not
// ported - needs the fuel/smelting/result inventory slots plus the crafting package, see
// tile.Furnace's doc comment) and occasionally play a cook sound. Neither is ported, so this is a
// no-op.
func (f *Furnace) OnScheduledUpdate() {}
