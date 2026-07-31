package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/sound"
)

// Button is a port of pocketmine\block\Button.
//
// PHP's Button is abstract (getActivationTime is abstract); Go has no abstract methods, so
// ActivationTime is a plain field, set by each concrete button type's constructor (WoodenButton,
// StoneButton, etc. — not yet ported) instead of an overridden method. Like Button itself, this
// struct isn't meant to be instantiated directly: it has no Clone() of its own, so it doesn't
// satisfy Behavior on its own — only a concrete leaf type embedding it (and implementing Clone)
// does, exactly mirroring PHP's abstract class.
type Button struct {
	Flowable
	FacingComponent

	Pressed        bool
	ActivationTime int
}

func (b *Button) DescribeBlockOnlyState(w runtime.DataDescriber) {
	b.DescribeFacing(w)
	w.Bool(&b.Pressed)
}

func (b *Button) IsPressed() bool { return b.Pressed }

func (b *Button) SetPressed(pressed bool) { b.Pressed = pressed }

func (b *Button) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if b.canBeSupportedAt(blockReplace, face) {
		b.Facing = face
		return b.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
	}
	return false
}

func (b *Button) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	if !b.Pressed {
		b.Pressed = true
		world, err := b.position.GetWorld()
		if err != nil {
			panic(err)
		}
		if err := world.SetBlock(b.position, b.self); err != nil {
			panic(err)
		}
		world.ScheduleDelayedBlockUpdate(b.position.AsVector3(), b.ActivationTime)
		world.AddSound(b.position.AsVector3().Add(0.5, 0.5, 0.5), sound.RedstonePowerOnSound{})
	}
	return true
}

func (b *Button) OnScheduledUpdate() {
	if b.Pressed {
		b.Pressed = false
		world, err := b.position.GetWorld()
		if err != nil {
			panic(err)
		}
		if err := world.SetBlock(b.position, b.self); err != nil {
			panic(err)
		}
		world.AddSound(b.position.AsVector3().Add(0.5, 0.5, 0.5), sound.RedstonePowerOffSound{})
	}
}

func (b *Button) OnNearbyBlockChange() {
	if !b.canBeSupportedAt(b.self, b.Facing) {
		world, err := b.position.GetWorld()
		if err != nil {
			panic(err)
		}
		world.UseBreakOn(b.position.AsVector3())
	}
}

func (b *Button) canBeSupportedAt(blk Behavior, face math.Facing) bool {
	return blk.(blockGeometry).GetAdjacentSupportType(math.Opposite(face)).HasCenterSupport()
}
