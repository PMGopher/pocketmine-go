package tile

import (
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

const (
	ChestTagPairX    = "pairx"
	ChestTagPairZ    = "pairz"
	ChestTagPairLead = "pairlead"
)

// NewDoubleChestInventoryFunc is `new DoubleChestInventory($left, $right)`, set by block/inventory
// (see Inventory).
var NewDoubleChestInventoryFunc func(left, right Inventory) Inventory

// loadedTerrainWorld is what Chest needs from its world to check its pair's chunk.
type loadedTerrainWorld interface {
	IsInLoadedTerrain(pos math.Vector3) bool
	IsChunkLoaded(chunkX, chunkZ int) bool
}

// Chest is a port of pocketmine\block\tile\Chest.
type Chest struct {
	SpawnableBase
	NameableComponent
	ContainerComponent

	doubleInventory Inventory

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

// ReadSaveData is a port of Chest::readSaveData.
func (c *Chest) ReadSaveData(tag *nbt.CompoundTag) error {
	pairXTag, okX := tag.GetTag(ChestTagPairX)
	pairZTag, okZ := tag.GetTag(ChestTagPairZ)
	pairXVal, isIntX := pairXTag.(nbt.IntTag)
	pairZVal, isIntZ := pairZTag.(nbt.IntTag)
	if okX && okZ && isIntX && isIntZ {
		pairX, pairZ := int(pairXVal), int(pairZVal)
		sameXAdjacentZ := c.position.FloorX() == pairX && absInt(c.position.FloorZ()-pairZ) == 1
		sameZAdjacentX := c.position.FloorZ() == pairZ && absInt(c.position.FloorX()-pairX) == 1
		if sameXAdjacentZ || sameZAdjacentX {
			c.PairX, c.PairZ, c.HasPair = pairX, pairZ, true
		} else {
			c.HasPair = false
		}
	}
	c.LoadName(tag)
	c.loadItems(c, tag)
	return nil
}

// WriteSaveData is a port of Chest::writeSaveData.
func (c *Chest) WriteSaveData(tag *nbt.CompoundTag) {
	if c.HasPair {
		tag.SetInt(ChestTagPairX, nbt.IntTag(c.PairX))
		tag.SetInt(ChestTagPairZ, nbt.IntTag(c.PairZ))
	}
	c.SaveName(tag)
	c.saveItems(c, tag)
}

// GetCleanedNBT is a port of Chest::getCleanedNBT: the pairing isn't kept in the item.
func (c *Chest) GetCleanedNBT() *nbt.CompoundTag {
	tag := c.TileBase.GetCleanedNBT()
	if tag != nil {
		//TODO: replace this with a purpose flag on writeSaveData()
		tag.RemoveTag(ChestTagPairX, ChestTagPairZ)
	}
	return tag
}

// CloseHook is Chest::close: the viewers of both its own and the double inventory are removed,
// and the pair forgets the double inventory.
func (c *Chest) CloseHook() {
	c.removeAllViewers()
	if c.doubleInventory != nil {
		world, _ := c.position.GetWorld()
		if lt, ok := world.(loadedTerrainWorld); ok && c.HasPair && lt.IsChunkLoaded(c.PairX>>4, c.PairZ>>4) {
			if RemoveAllViewersFunc != nil {
				RemoveAllViewersFunc(c.doubleInventory)
			}
			if pair, ok := c.GetPair(); ok {
				pair.doubleInventory = nil
			}
		}
		c.doubleInventory = nil
	}
}

// OnBlockDestroyedHook is a port of Chest::onBlockDestroyedHook.
func (c *Chest) OnBlockDestroyedHook() {
	c.Unpair()
	c.dropContents(c)
}

// GetInventory is a port of Chest::getInventory: the double chest inventory when paired,
// otherwise the chest's own.
func (c *Chest) GetInventory() Inventory {
	if c.HasPair && c.doubleInventory == nil {
		c.checkPairing()
	}
	if c.doubleInventory != nil {
		return c.doubleInventory
	}
	return c.realInventory(c)
}

// GetRealInventory is a port of Chest::getRealInventory (the chest's own ChestInventory).
func (c *Chest) GetRealInventory() Inventory { return c.realInventory(c) }

// checkPairing is a port of Chest::checkPairing.
func (c *Chest) checkPairing() {
	world, _ := c.position.GetWorld()
	lt, _ := world.(loadedTerrainWorld)
	if c.HasPair && lt != nil && !lt.IsInLoadedTerrain(math.NewVector3(float64(c.PairX), c.position.Y, float64(c.PairZ))) {
		// paired to a tile in an unloaded chunk
		c.doubleInventory = nil
	} else if pair, ok := c.GetPair(); ok {
		if !pair.HasPair {
			pair.createPair(c)
			pair.checkPairing()
		}
		if c.doubleInventory == nil {
			if pair.doubleInventory != nil {
				c.doubleInventory = pair.doubleInventory
			} else if NewDoubleChestInventoryFunc != nil {
				if pair.position.FloorX()+(pair.position.FloorZ()<<15) > c.position.FloorX()+(c.position.FloorZ()<<15) { // Order them correctly
					c.doubleInventory = NewDoubleChestInventoryFunc(pair.realInventory(pair), c.realInventory(c))
				} else {
					c.doubleInventory = NewDoubleChestInventoryFunc(c.realInventory(c), pair.realInventory(pair))
				}
				pair.doubleInventory = c.doubleInventory
			}
		}
	} else {
		c.doubleInventory = nil
		c.HasPair = false
	}
}

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
	c.checkPairing()
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
		pair.checkPairing()
		pair.ClearSpawnCompoundCache()
	}
	c.checkPairing()
	return true
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
