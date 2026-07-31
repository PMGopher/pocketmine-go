package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// Vine is a port of pocketmine\block\Vine.
//
// PHP represents faces as an assoc array keyed (and valued) by Facing; Go uses a
// map[math.Facing]bool set instead, converting to/from the []Facing slice DataDescriber's
// HorizontalFacingFlags needs at describe time (same "local variable dance" as Lever's enum
// encoding).
type Vine struct {
	Flowable

	Faces map[math.Facing]bool
}

func NewVine(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Vine {
	v := &Vine{Flowable: Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}, Faces: map[math.Facing]bool{}}
	v.Init(v)
	return v
}

// Clone can't use the usual one-line `c := *v; c.rebind(&c); return &c` pattern: Faces is a map,
// a reference type, so a plain struct copy would leave the clone sharing the same underlying map
// as the original — mutating one would mutate the other. The map has to be copied explicitly.
func (v *Vine) Clone() Behavior {
	c := *v
	c.Faces = make(map[math.Facing]bool, len(v.Faces))
	for f := range v.Faces {
		c.Faces[f] = true
	}
	c.rebind(&c)
	return &c
}

func (v *Vine) DescribeBlockOnlyState(w runtime.DataDescriber) {
	faces := make([]math.Facing, 0, len(v.Faces))
	for f := range v.Faces {
		faces = append(faces, f)
	}
	w.HorizontalFacingFlags(&faces)
	v.Faces = make(map[math.Facing]bool, len(faces))
	for _, f := range faces {
		v.Faces[f] = true
	}
}

func (v *Vine) GetFaces() []math.Facing {
	faces := make([]math.Facing, 0, len(v.Faces))
	for f := range v.Faces {
		faces = append(faces, f)
	}
	return faces
}

func (v *Vine) HasFace(face math.Facing) bool { return v.Faces[face] }

func validateHorizontalFace(face math.Facing) {
	if face != math.North && face != math.South && face != math.West && face != math.East {
		panic("Facing can only be north, east, south or west")
	}
}

func (v *Vine) SetFaces(faces []math.Facing) {
	unique := make(map[math.Facing]bool, len(faces))
	for _, f := range faces {
		validateHorizontalFace(f)
		unique[f] = true
	}
	v.Faces = unique
}

func (v *Vine) SetFace(face math.Facing, value bool) {
	validateHorizontalFace(face)
	if value {
		v.Faces[face] = true
	} else {
		delete(v.Faces, face)
	}
}

func (v *Vine) HasEntityCollision() bool { return true }

func (v *Vine) CanClimb() bool { return true }

func (v *Vine) CanBeReplaced() bool { return true }

func (v *Vine) OnEntityInside(entity Entity) bool {
	entity.ResetFallDistance()
	return true
}

func (v *Vine) RecalculateCollisionBoxes() []math.AxisAlignedBB { return nil }

func (v *Vine) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	opposite := math.Opposite(face)
	if !blockReplace.(blockGeometry).GetSide(opposite, 1).(blockGeometry).IsFullCube() || math.FacingAxis(face) == math.AxisY {
		return false
	}

	if replaceVine, ok := blockReplace.(*Vine); ok {
		v.Faces = make(map[math.Facing]bool, len(replaceVine.Faces)+1)
		for f := range replaceVine.Faces {
			v.Faces[f] = true
		}
	} else {
		v.Faces = map[math.Facing]bool{}
	}
	v.Faces[opposite] = true

	return v.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

func (v *Vine) OnNearbyBlockChange() {
	changed := false
	up, upIsVine := v.GetSide(math.Up, 1).(*Vine)

	for face := range v.Faces {
		supported := upIsVine && up.Faces[face]
		if !supported && !v.GetSide(face, 1).IsSolid() {
			delete(v.Faces, face)
			changed = true
		}
	}

	if !changed {
		return
	}
	world, err := v.position.GetWorld()
	if err != nil {
		return
	}
	if len(v.Faces) == 0 {
		world.UseBreakOn(v.position.AsVector3())
	} else if err := world.SetBlock(v.position, v.self); err != nil {
		panic(err)
	}
}

func (v *Vine) TicksRandomly() bool { return true }

// OnRandomTick is a stub: vine growth isn't implemented in the PHP original either (see its
// `//TODO: vine growth`).
func (v *Vine) OnRandomTick() {}

func (v *Vine) GetDrops(item Item) []Item {
	if item.GetBlockToolType()&ToolTypeShears != 0 {
		return v.self.GetDropsForCompatibleTool(item)
	}
	return nil
}

func (v *Vine) GetFlameEncouragement() int { return 15 }

func (v *Vine) GetFlammability() int { return 100 }
