package block

import (
	"math/rand"

	"pocketmine-go/pocketmine/block/tile"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/sound"
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

// furnaceTile is `instanceof TileFurnace`: tile.NormalFurnace, BlastFurnace and Smoker.
type furnaceTile interface {
	CanOpenWith(key string) bool
	GetInventory() tile.Inventory
	OnUpdate() bool
	GetFurnaceType() tile.FurnaceType
}

// OnInteract is a port of Furnace::onInteract.
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
	tileFurnace, ok := t.(furnaceTile)
	if !ok {
		return true
	}
	if tileFurnace.CanOpenWith(item.GetCustomName()) {
		openTileWindow(player, tileFurnace.GetInventory())
	}
	return true
}

// OnScheduledUpdate is a port of Furnace::onScheduledUpdate.
func (f *Furnace) OnScheduledUpdate() {
	world, err := f.position.GetWorld()
	if err != nil {
		return
	}
	t, _ := world.GetTile(f.position)
	if furnace, ok := t.(furnaceTile); ok && furnace.OnUpdate() {
		if rand.Intn(60) == 0 { // in vanilla this is between 1 and 5 seconds; try to average about 3
			world.AddSound(f.position.Vector3, furnaceCookSound(furnace.GetFurnaceType()))
		}
		world.ScheduleDelayedBlockUpdate(f.position.Vector3, 1) //TODO: check this
	}
}

// furnaceCookSound is a port of FurnaceType::getCookSound.
func furnaceCookSound(t tile.FurnaceType) sound.Sound {
	switch t {
	case tile.FurnaceTypeBlastFurnace:
		return sound.BlastFurnaceSound{}
	case tile.FurnaceTypeSmoker:
		return sound.SmokerSound{}
	case tile.FurnaceTypeCampfire, tile.FurnaceTypeSoulCampfire:
		return sound.CampfireSound{}
	}
	return sound.FurnaceSound{}
}
