package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

const (
	snowLayerMinLayers = 1
	snowLayerMaxLayers = 8
)

// SnowLayer is a port of pocketmine\block\SnowLayer.
type SnowLayer struct {
	Flowable
	FallableComponent

	Layers int
}

func NewSnowLayer(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *SnowLayer {
	s := &SnowLayer{Flowable: Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}, Layers: snowLayerMinLayers}
	s.Init(s)
	return s
}

func (s *SnowLayer) Clone() Behavior {
	c := *s
	c.rebind(&c)
	return &c
}

func (s *SnowLayer) DescribeBlockOnlyState(w runtime.DataDescriber) {
	layers := s.Layers
	w.BoundedIntAuto(snowLayerMinLayers, snowLayerMaxLayers, &layers)
	s.Layers = layers
}

func (s *SnowLayer) GetLayers() int { return s.Layers }

func (s *SnowLayer) SetLayers(layers int) {
	if layers < snowLayerMinLayers || layers > snowLayerMaxLayers {
		panic("Layers must be in range 1 ... 8")
	}
	s.Layers = layers
}

func (s *SnowLayer) CanBeReplaced() bool { return s.Layers < snowLayerMaxLayers }

func (s *SnowLayer) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	return []math.AxisAlignedBB{math.OneAABB().TrimmedCopy(math.Up, float64(snowLayerMaxLayers-s.Layers+1)/8)}
}

func (s *SnowLayer) GetSupportType(facing math.Facing) blockutils.SupportType {
	if !s.CanBeReplaced() {
		return blockutils.SupportTypeFull
	}
	return blockutils.SupportTypeNone
}

func (s *SnowLayer) canBeSupportedAt(blk Behavior) bool {
	return blk.(blockGeometry).GetAdjacentSupportType(math.Down) == blockutils.SupportTypeFull
}

func (s *SnowLayer) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if replace, ok := blockReplace.(*SnowLayer); ok {
		if replace.Layers >= snowLayerMaxLayers {
			return false
		}
		s.Layers = replace.Layers + 1
	}
	if s.canBeSupportedAt(blockReplace) {
		return s.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
	}
	return false
}

func (s *SnowLayer) TicksRandomly() bool { return true }

// OnRandomTick is a port of SnowLayer::onRandomTick.
func (s *SnowLayer) OnRandomTick() {
	world, err := s.position.GetWorld()
	if err != nil {
		return
	}
	pos := s.position.AsVector3()
	if world.GetBlockLightAt(pos.FloorX(), pos.FloorY(), pos.FloorZ()) >= 12 {
		Melt(s.self, VanillaAir())
	}
}

// GetDropsForCompatibleTool should return [VanillaItems.SNOWBALL().SetCount(max(1, Layers/2))] —
// needs real Item construction from the unported item package (see
// Block.GetDropsForCompatibleTool's doc comment), so this returns nil for now.
func (s *SnowLayer) GetDropsForCompatibleTool(item Item) []Item { return nil }
