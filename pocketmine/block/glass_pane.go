package block

import "pocketmine-go/pocketmine/math"

// GlassPane is a port of pocketmine\block\GlassPane.
type GlassPane struct {
	Thin
}

func NewGlassPane(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *GlassPane {
	g := &GlassPane{Thin{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}, Connections: map[math.Facing]bool{}}}
	g.Init(g)
	return g
}

// Clone can't use the usual one-line pattern: Connections is a map, a reference type - see
// Vine.Clone's doc comment for the same reasoning.
func (g *GlassPane) Clone() Behavior {
	c := *g
	c.Connections = make(map[math.Facing]bool, len(g.Connections))
	for k, v := range g.Connections {
		c.Connections[k] = v
	}
	c.rebind(&c)
	return &c
}

// GetDropsForCompatibleTool deliberately returns nothing, matching the PHP original's
// `return [];` (this isn't a not-yet-ported gap).
func (g *GlassPane) GetDropsForCompatibleTool(item Item) []Item { return nil }

func (g *GlassPane) IsAffectedBySilkTouch() bool { return true }
