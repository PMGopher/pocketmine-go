package block

import (
	"math/rand"
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// GlowLichen is a port of pocketmine\block\GlowLichen.
type GlowLichen struct {
	Transparent

	Faces map[math.Facing]bool
}

func NewGlowLichen(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *GlowLichen {
	g := &GlowLichen{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}, Faces: map[math.Facing]bool{}}
	g.Init(g)
	return g
}

// Clone can't use the usual one-line pattern: Faces is a map, a reference type - see Vine.Clone's
// doc comment for the same reasoning.
func (g *GlowLichen) Clone() Behavior {
	c := *g
	c.Faces = make(map[math.Facing]bool, len(g.Faces))
	for k, v := range g.Faces {
		c.Faces[k] = v
	}
	c.rebind(&c)
	return &c
}

func (g *GlowLichen) DescribeBlockOnlyState(w runtime.DataDescriber) {
	faces := make([]math.Facing, 0, len(g.Faces))
	for f := range g.Faces {
		faces = append(faces, f)
	}
	w.FacingFlags(&faces)
	g.Faces = make(map[math.Facing]bool, len(faces))
	for _, f := range faces {
		g.Faces[f] = true
	}
}

func (g *GlowLichen) GetFaces() []math.Facing {
	faces := make([]math.Facing, 0, len(g.Faces))
	for f := range g.Faces {
		faces = append(faces, f)
	}
	return faces
}

func (g *GlowLichen) HasFace(face math.Facing) bool { return g.Faces[face] }

func (g *GlowLichen) SetFaces(faces []math.Facing) {
	unique := make(map[math.Facing]bool, len(faces))
	for _, f := range faces {
		math.ValidateFacing(f)
		unique[f] = true
	}
	g.Faces = unique
}

func (g *GlowLichen) SetFace(face math.Facing, value bool) {
	math.ValidateFacing(face)
	if value {
		g.Faces[face] = true
	} else {
		delete(g.Faces, face)
	}
}

func (g *GlowLichen) GetLightLevel() int { return 7 }

func (g *GlowLichen) IsSolid() bool { return false }

func (g *GlowLichen) RecalculateCollisionBoxes() []math.AxisAlignedBB { return nil }

func (g *GlowLichen) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}

func (g *GlowLichen) CanBeReplaced() bool { return true }

func (g *GlowLichen) getInitialPlaceFaces(blockReplace Behavior) map[math.Facing]bool {
	if replace, ok := blockReplace.(*GlowLichen); ok {
		faces := make(map[math.Facing]bool, len(replace.Faces))
		for k, v := range replace.Faces {
			faces[k] = v
		}
		return faces
	}
	return map[math.Facing]bool{}
}

// getAvailableFaces iterates math.AllFacing in order, matching PHP's Facing::ALL iteration order
// (array_key_first relies on this same deterministic order in Place below).
func (g *GlowLichen) getAvailableFaces() []math.Facing {
	var faces []math.Facing
	geo := g.self.(blockGeometry)
	for _, face := range math.AllFacing {
		if !g.Faces[face] && geo.GetAdjacentSupportType(face) == blockutils.SupportTypeFull {
			faces = append(faces, face)
		}
	}
	return faces
}

func (g *GlowLichen) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	g.Faces = g.getInitialPlaceFaces(blockReplace)
	available := g.getAvailableFaces()
	if len(available) == 0 {
		return false
	}

	opposite := math.Opposite(face)
	placedFace := available[0]
	for _, f := range available {
		if f == opposite {
			placedFace = opposite
			break
		}
	}
	g.Faces[placedFace] = true

	return g.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

func (g *GlowLichen) OnNearbyBlockChange() {
	changed := false
	geo := g.self.(blockGeometry)
	for face := range g.Faces {
		if geo.GetAdjacentSupportType(face) != blockutils.SupportTypeFull {
			delete(g.Faces, face)
			changed = true
		}
	}

	if !changed {
		return
	}
	world, err := g.position.GetWorld()
	if err != nil {
		return
	}
	if len(g.Faces) == 0 {
		world.UseBreakOn(g.position.AsVector3())
	} else if err := world.SetBlock(g.position, g.self); err != nil {
		panic(err)
	}
}

// OnInteract is a port of GlowLichen::onInteract: bone meal spreads it to a random neighbouring
// face.
func (g *GlowLichen) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	if !isFertilizer(item) || len(g.Faces) == 0 {
		return false
	}
	shuffledFaces := g.GetFaces()
	rand.Shuffle(len(shuffledFaces), func(i, j int) { shuffledFaces[i], shuffledFaces[j] = shuffledFaces[j], shuffledFaces[i] })
	spreadMethods := []func(math.Facing) bool{g.spreadAroundSupport, g.spreadAdjacentToSupport, g.spreadWithinSelf}
	rand.Shuffle(len(spreadMethods), func(i, j int) { spreadMethods[i], spreadMethods[j] = spreadMethods[j], spreadMethods[i] })
spread:
	for _, sourceFace := range shuffledFaces {
		for _, spreadMethod := range spreadMethods {
			if spreadMethod(sourceFace) {
				item.Pop()
				break spread
			}
		}
	}
	return true
}

// getSpreadBlock is a port of GlowLichen::getSpreadBlock.
func (g *GlowLichen) getSpreadBlock(replace Behavior, spreadFace math.Facing) Behavior {
	var result *GlowLichen
	if lichen, ok := replace.(*GlowLichen); ok && lichen.HasSameTypeId(g.self) {
		if lichen.HasFace(spreadFace) {
			return nil
		}
		result = lichen.Clone().(*GlowLichen)
	} else if replace.GetTypeId() == AIR {
		result = VanillaBlock("glow_lichen").(*GlowLichen)
	} else {
		//TODO: if this is a water block, generate a waterlogged block
		return nil
	}
	result.SetFace(spreadFace, true)
	return result
}

// spread is a port of GlowLichen::spread.
func (g *GlowLichen) spread(world World, replacePos math.Vector3, spreadFace math.Facing) bool {
	side := replacePos.GetSide(spreadFace, 1)
	supportBlock := world.GetBlockAt(side.FloorX(), side.FloorY(), side.FloorZ())
	supportFace := math.Opposite(spreadFace)
	if supportBlock.GetSupportType(supportFace) != blockutils.SupportTypeFull {
		return false
	}
	replacedBlock := supportBlock.(blockGeometry).GetSide(supportFace, 1)
	replacementBlock := g.getSpreadBlock(replacedBlock, math.Opposite(supportFace))
	if replacementBlock == nil {
		return false
	}
	return Spread(replacedBlock, replacementBlock, g.self)
}

// shuffledSpreadFaces is a port of GlowLichen::getShuffledSpreadFaces: every face not on
// sourceFace's axis, in random order.
func shuffledSpreadFaces(sourceFace math.Facing) []math.Facing {
	skipAxis := math.FacingAxis(sourceFace)
	faces := append([]math.Facing(nil), math.AllFacing...)
	rand.Shuffle(len(faces), func(i, j int) { faces[i], faces[j] = faces[j], faces[i] })
	result := faces[:0]
	for _, f := range faces {
		if math.FacingAxis(f) != skipAxis {
			result = append(result, f)
		}
	}
	return result
}

// spreadAroundSupport is a port of GlowLichen::spreadAroundSupport.
func (g *GlowLichen) spreadAroundSupport(sourceFace math.Facing) bool {
	world, err := g.position.GetWorld()
	if err != nil {
		return false
	}
	supportPos := g.position.GetSide(sourceFace, 1)
	for _, spreadFace := range shuffledSpreadFaces(sourceFace) {
		if g.spread(world, supportPos.GetSide(spreadFace, 1).Vector3, math.Opposite(spreadFace)) {
			return true
		}
	}
	return false
}

// spreadAdjacentToSupport is a port of GlowLichen::spreadAdjacentToSupport.
func (g *GlowLichen) spreadAdjacentToSupport(sourceFace math.Facing) bool {
	world, err := g.position.GetWorld()
	if err != nil {
		return false
	}
	for _, spreadFace := range shuffledSpreadFaces(sourceFace) {
		if g.spread(world, g.position.GetSide(spreadFace, 1).Vector3, sourceFace) {
			return true
		}
	}
	return false
}

// spreadWithinSelf is a port of GlowLichen::spreadWithinSelf.
func (g *GlowLichen) spreadWithinSelf(sourceFace math.Facing) bool {
	world, err := g.position.GetWorld()
	if err != nil {
		return false
	}
	for _, spreadFace := range shuffledSpreadFaces(sourceFace) {
		if !g.HasFace(spreadFace) && g.spread(world, g.position.Vector3, spreadFace) {
			return true
		}
	}
	return false
}

func (g *GlowLichen) GetDrops(item Item) []Item {
	if item.GetBlockToolType()&ToolTypeShears != 0 {
		return g.self.GetDropsForCompatibleTool(item)
	}
	return nil
}

func (g *GlowLichen) GetFlameEncouragement() int { return 15 }

func (g *GlowLichen) GetFlammability() int { return 100 }
