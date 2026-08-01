package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// ResinClump is a port of pocketmine\block\ResinClump. Same MultiAnySupportTrait algorithm as
// GlowLichen - see glow_lichen.go for the shared reasoning (Go map iteration order vs PHP's
// insertion-order Facing::ALL tie-break in Place, etc).
type ResinClump struct {
	Transparent

	Faces map[math.Facing]bool
}

func NewResinClump(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *ResinClump {
	r := &ResinClump{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}, Faces: map[math.Facing]bool{}}
	r.Init(r)
	return r
}

// Clone can't use the usual one-line pattern: Faces is a map, a reference type - see Vine.Clone's
// doc comment for the same reasoning.
func (r *ResinClump) Clone() Behavior {
	c := *r
	c.Faces = make(map[math.Facing]bool, len(r.Faces))
	for k, v := range r.Faces {
		c.Faces[k] = v
	}
	c.rebind(&c)
	return &c
}

func (r *ResinClump) DescribeBlockOnlyState(w runtime.DataDescriber) {
	faces := make([]math.Facing, 0, len(r.Faces))
	for f := range r.Faces {
		faces = append(faces, f)
	}
	w.FacingFlags(&faces)
	r.Faces = make(map[math.Facing]bool, len(faces))
	for _, f := range faces {
		r.Faces[f] = true
	}
}

func (r *ResinClump) GetFaces() []math.Facing {
	faces := make([]math.Facing, 0, len(r.Faces))
	for f := range r.Faces {
		faces = append(faces, f)
	}
	return faces
}

func (r *ResinClump) HasFace(face math.Facing) bool { return r.Faces[face] }

func (r *ResinClump) SetFaces(faces []math.Facing) {
	unique := make(map[math.Facing]bool, len(faces))
	for _, f := range faces {
		math.ValidateFacing(f)
		unique[f] = true
	}
	r.Faces = unique
}

func (r *ResinClump) SetFace(face math.Facing, value bool) {
	math.ValidateFacing(face)
	if value {
		r.Faces[face] = true
	} else {
		delete(r.Faces, face)
	}
}

func (r *ResinClump) IsSolid() bool { return false }

func (r *ResinClump) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}

func (r *ResinClump) CanBeReplaced() bool { return true }

func (r *ResinClump) RecalculateCollisionBoxes() []math.AxisAlignedBB { return nil }

func (r *ResinClump) getInitialPlaceFaces(blockReplace Behavior) map[math.Facing]bool {
	if replace, ok := blockReplace.(*ResinClump); ok {
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
func (r *ResinClump) getAvailableFaces() []math.Facing {
	var faces []math.Facing
	geo := r.self.(blockGeometry)
	for _, face := range math.AllFacing {
		if !r.Faces[face] && geo.GetAdjacentSupportType(face) == blockutils.SupportTypeFull {
			faces = append(faces, face)
		}
	}
	return faces
}

func (r *ResinClump) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	r.Faces = r.getInitialPlaceFaces(blockReplace)
	available := r.getAvailableFaces()
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
	r.Faces[placedFace] = true

	return r.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

func (r *ResinClump) OnNearbyBlockChange() {
	changed := false
	geo := r.self.(blockGeometry)
	for face := range r.Faces {
		if geo.GetAdjacentSupportType(face) != blockutils.SupportTypeFull {
			delete(r.Faces, face)
			changed = true
		}
	}

	if !changed {
		return
	}
	world, err := r.position.GetWorld()
	if err != nil {
		return
	}
	if len(r.Faces) == 0 {
		world.UseBreakOn(r.position.AsVector3())
	} else if err := world.SetBlock(r.position, r.self); err != nil {
		panic(err)
	}
}
