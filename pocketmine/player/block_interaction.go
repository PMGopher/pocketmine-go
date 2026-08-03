package player

import (
	stdmath "math"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/math"
)

// maxReachDistanceCreative/Survival mirror Player::MAX_REACH_DISTANCE_CREATIVE/SURVIVAL.
const (
	maxReachDistanceCreative = 13
	maxReachDistanceSurvival = 7
)

// GetDirectionVector is a port of Entity::getDirectionVector, using Player's own yaw/pitch (this
// port's Entity doesn't carry a Location with rotation - see Player's own doc comment on why
// yaw/pitch live here instead).
func (p *Player) GetDirectionVector() math.Vector3 {
	pitchRad := p.pitch * stdmath.Pi / 180
	yawRad := p.yaw * stdmath.Pi / 180

	y := -stdmath.Sin(pitchRad)
	xz := stdmath.Cos(pitchRad)
	x := -xz * stdmath.Sin(yawRad)
	z := xz * stdmath.Cos(yawRad)

	return math.NewVector3(x, y, z).Normalize()
}

// CanInteract is a port of Player::canInteract.
func (p *Player) CanInteract(pos math.Vector3, maxDistance float64) bool {
	return p.canInteract(pos, maxDistance, stdmath.Sqrt(3)/2)
}

func (p *Player) canInteract(pos math.Vector3, maxDistance, maxDiff float64) bool {
	eyePos := p.GetEyePos()
	if eyePos.DistanceSquared(pos) > maxDistance*maxDistance {
		return false
	}

	dV := p.GetDirectionVector()
	eyeDot := dV.Dot(eyePos)
	targetDot := dV.Dot(pos)
	return (targetDot - eyeDot) >= -maxDiff
}

// hasTypeTagChecker/sideAccessible are the local surfaces AttackBlock's fire-extinguish check
// needs - HasTypeTag/GetSide are both promoted from *block.Block, not part of block.Behavior
// itself (same "declare the exact promoted method this file needs" convention as world.go's own
// positionable/fullCubeChecker).
type hasTypeTagChecker interface{ HasTypeTag(tag string) bool }
type sideAccessible interface {
	GetSide(side math.Facing, step int) block.Behavior
}

// blockTypeTagsFire mirrors BlockTypeTags::FIRE ("pocketmine:fire").
const blockTypeTagsFire = "pocketmine:fire"

// AttackBlock is a port of a slice of Player::attackBlock.
//
// Not ported: PlayerInteractEvent (no cancellable event bus wired to World/Player yet - matches
// this port's other documented "no event bus" gaps) and the ArmSwingAnimation broadcast to viewers
// (no viewer/broadcast-to-nearby-players plumbing exists - matches World.AddSound's own gap).
// heldItem is the item the player is holding right now (see SurvivalBlockBreakHandler's own doc
// comment on why this port's Player has no "selected hotbar slot" concept yet to read it from
// directly).
func (p *Player) AttackBlock(pos math.Vector3, face math.Facing, heldItem block.Item) bool {
	if pos.DistanceSquared(p.GetPosition()) > 10000 {
		return false
	}

	target := p.world.GetBlockAt(pos.FloorX(), pos.FloorY(), pos.FloorZ())

	if target.OnAttack(heldItem, face, p) {
		return true
	}

	if sa, ok := target.(sideAccessible); ok {
		sideBlock := sa.GetSide(face, 1)
		if tc, ok := sideBlock.(hasTypeTagChecker); ok && tc.HasTypeTag(blockTypeTagsFire) {
			_ = p.world.SetBlock(sideBlock.GetPosition(), block.VanillaAir())
			return true
		}
	}

	if !p.IsCreative() && !target.GetBreakInfo().BreaksInstantly() {
		p.blockBreakHandler = NewSurvivalBlockBreakHandler(p, pos, target, face, 16, heldItem)
	}

	return true
}

// ContinueBreakBlock is a port of Player::continueBreakBlock.
func (p *Player) ContinueBreakBlock(pos math.Vector3, face math.Facing) {
	if p.blockBreakHandler != nil && p.blockBreakHandler.GetBlockPos().DistanceSquared(pos) < 0.0001 {
		p.blockBreakHandler.SetTargetedFace(face)
	}
}

// StopBreakBlock is a port of Player::stopBreakBlock.
func (p *Player) StopBreakBlock(pos math.Vector3) {
	if p.blockBreakHandler != nil && p.blockBreakHandler.GetBlockPos().DistanceSquared(pos) < 0.0001 {
		p.blockBreakHandler.Close()
		p.blockBreakHandler = nil
	}
}

// BreakBlock is a port of a slice of Player::breakBlock.
//
// Not ported: removeCurrentWindow (no inventory-window system exists), the ArmSwingAnimation
// broadcast (see AttackBlock's own doc comment), and returnItemsFromAction/HungerManager::exhaust
// (no drops/hunger system wired to this yet - World.UseBreakOn is already this port's own
// documented simplified form, see its doc comment).
func (p *Player) BreakBlock(pos math.Vector3) bool {
	maxDistance := float64(maxReachDistanceSurvival)
	if p.IsCreative() {
		maxDistance = maxReachDistanceCreative
	}

	if !p.CanInteract(pos.Add(0.5, 0.5, 0.5), maxDistance) {
		return false
	}

	p.StopBreakBlock(pos)
	return p.world.UseBreakOn(pos)
}

// UpdateBreakingBlock is a port of the relevant slice of Player::onUpdate's own
// `$this->blockBreakHandler?->update()` call - drives the active SurvivalBlockBreakHandler (if
// any) for one tick, cancelling the break (matching real PHP's own cancel-on-too-far-away
// behaviour) if it reports the player has moved out of range.
func (p *Player) UpdateBreakingBlock(heldItem block.Item) {
	if p.blockBreakHandler == nil {
		return
	}
	if !p.blockBreakHandler.Update(heldItem) {
		p.blockBreakHandler.Close()
		p.blockBreakHandler = nil
	}
}

// GetBlockBreakHandler is a port of Player::$blockBreakHandler's own implicit getter usage
// elsewhere in real PHP (there's no dedicated getBlockBreakHandler() method in real PHP - callers
// just read the property directly - this port exposes it the same way any other unexported field
// would be, via an accessor).
func (p *Player) GetBlockBreakHandler() *SurvivalBlockBreakHandler { return p.blockBreakHandler }
