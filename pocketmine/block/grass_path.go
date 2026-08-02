package block

import "pocketmine-go/pocketmine/math"

// GrassPath is a port of pocketmine\block\GrassPath.
type GrassPath struct {
	Transparent
}

func NewGrassPath(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *GrassPath {
	g := &GrassPath{Transparent{NewBlock(idInfo, name, typeInfo)}}
	g.Init(g)
	return g
}

func (g *GrassPath) Clone() Behavior {
	c := *g
	c.rebind(&c)
	return &c
}

func (g *GrassPath) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	return []math.AxisAlignedBB{math.OneAABB().TrimmedCopy(math.Up, 1.0/16)}
}

// OnNearbyBlockChange is a port of GrassPath::onNearbyBlockChange.
func (g *GrassPath) OnNearbyBlockChange() {
	if g.self.(blockGeometry).GetSide(math.Up, 1).IsSolid() {
		world, err := g.position.GetWorld()
		if err != nil {
			return
		}
		_ = world.SetBlock(g.position, VanillaDirt())
	}
}

func (g *GrassPath) IsAffectedBySilkTouch() bool { return true }

// GetDropsForCompatibleTool is a port of GrassPath::getDropsForCompatibleTool.
func (g *GrassPath) GetDropsForCompatibleTool(item Item) []Item {
	dropped := asItemOrNil(VanillaDirt())
	if dropped == nil {
		return nil
	}
	return []Item{dropped}
}
