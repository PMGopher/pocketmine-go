package tile

import (
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

const (
	ChestTagPairX = "pairx"
	ChestTagPairZ = "pairz"
)

// Chest is a port of pocketmine\block\tile\Chest, minus its inventory/Container half entirely -
// see ContainerComponent's doc comment for why the inventory package can't be imported here.
// Everything else - pairing state (IsPaired/GetPair/PairWith/Unpair), the name, and the lock (via
// ContainerComponent.CanOpenWith) - is fully real.
type Chest struct {
	SpawnableBase
	NameableComponent
	ContainerComponent

	PairX, PairZ int
	HasPair      bool
}

func NewChest(world World, pos math.Vector3) *Chest {
	c := &Chest{}
	c.SpawnableBase = SpawnableBase{TileBase: NewTileBase(world, pos)}
	c.Init(c)
	return c
}

func (c *Chest) SaveID() string { return "Chest" }

func (c *Chest) GetDefaultName() string { return "Chest" }

func (c *Chest) GetName() string { return c.NameableComponent.GetName(c) }

// IsPaired is a port of Chest::isPaired.
func (c *Chest) IsPaired() bool { return c.HasPair }

// GetPair is a port of Chest::getPair.
func (c *Chest) GetPair() (*Chest, bool) {
	if !c.HasPair {
		return nil, false
	}
	world, ok := c.position.GetWorld()
	if !ok {
		return nil, false
	}
	t, ok := world.GetTileAt(c.PairX, c.position.FloorY(), c.PairZ)
	if !ok {
		return nil, false
	}
	pair, ok := t.(*Chest)
	return pair, ok
}

func (c *Chest) createPair(other *Chest) {
	c.PairX, c.PairZ, c.HasPair = other.position.FloorX(), other.position.FloorZ(), true
	other.PairX, other.PairZ, other.HasPair = c.position.FloorX(), c.position.FloorZ(), true
}

// PairWith is a port of Chest::pairWith.
func (c *Chest) PairWith(other *Chest) bool {
	if c.HasPair || other.HasPair {
		return false
	}
	c.createPair(other)
	c.ClearSpawnCompoundCache()
	other.ClearSpawnCompoundCache()
	return true
}

// Unpair is a port of Chest::unpair.
func (c *Chest) Unpair() bool {
	if !c.HasPair {
		return false
	}
	pair, hadPair := c.GetPair()
	c.HasPair = false
	c.ClearSpawnCompoundCache()

	if hadPair {
		pair.HasPair = false
		pair.ClearSpawnCompoundCache()
	}
	return true
}

func (c *Chest) ReadSaveData(tag *nbt.CompoundTag) error {
	pairXTag, okX := tag.GetInt(ChestTagPairX)
	pairZTag, okZ := tag.GetInt(ChestTagPairZ)
	if okX == nil && okZ == nil {
		pairX, pairZ := int(pairXTag), int(pairZTag)
		sameXAdjacentZ := c.position.FloorX() == pairX && absInt(c.position.FloorZ()-pairZ) == 1
		sameZAdjacentX := c.position.FloorZ() == pairZ && absInt(c.position.FloorX()-pairX) == 1
		if sameXAdjacentZ || sameZAdjacentX {
			c.PairX, c.PairZ, c.HasPair = pairX, pairZ, true
		} else {
			c.HasPair = false
		}
	}
	c.LoadName(tag)
	return nil
}

func (c *Chest) WriteSaveData(tag *nbt.CompoundTag) {
	if c.HasPair {
		tag.SetInt(ChestTagPairX, nbt.IntTag(c.PairX))
		tag.SetInt(ChestTagPairZ, nbt.IntTag(c.PairZ))
	}
	c.SaveName(tag)
}

func (c *Chest) AddAdditionalSpawnData(tag *nbt.CompoundTag) {
	if c.HasPair {
		tag.SetInt(ChestTagPairX, nbt.IntTag(c.PairX))
		tag.SetInt(ChestTagPairZ, nbt.IntTag(c.PairZ))
	}
	c.NameableComponent.AddAdditionalSpawnData(tag)
}

// CopyDataFromItem must be defined here rather than relying on promotion - see
// NameableComponent.ApplyItemCustomName's doc comment for why.
func (c *Chest) CopyDataFromItem(item Item) {
	c.TileBase.CopyDataFromItem(item)
	c.ApplyItemCustomName(item)
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
