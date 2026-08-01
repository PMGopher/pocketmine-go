package block

import "pocketmine-go/pocketmine/entity"

// fireShaper lets BaseFire reach a concrete leaf's GetFireDamage - same self-dispatch shape as
// bannerShaper/signShaper elsewhere in this port.
type fireShaper interface {
	GetFireDamage() int
}

// BaseFire is a port of pocketmine\block\BaseFire. Like Crops/Stem, this isn't meant to be
// instantiated directly - a concrete leaf type (Fire, SoulFire) must embed it and implement
// Clone.
type BaseFire struct {
	Flowable
}

func (b *BaseFire) HasEntityCollision() bool { return true }

func (b *BaseFire) CanBeReplaced() bool { return true }

// OnEntityInside is a port of BaseFire::onEntityInside. The Arrow-immunity check on the combust
// event is dropped since there's no Arrow marker yet - so the entity always combusts, a minor
// behavioural gap versus a full stub.
func (b *BaseFire) OnEntityInside(e Entity) bool {
	damage := b.self.(fireShaper).GetFireDamage()
	dmgEv := entity.NewEntityDamageByBlockEvent(b.self, e, entity.EntityDamageCauseFire, float64(damage), nil)
	e.Attack(dmgEv)

	combustEv := entity.NewEntityCombustByBlockEvent(b.self, e, 8)
	combustEv.Call()
	if !combustEv.IsCancelled() {
		e.SetOnFire(combustEv.GetDuration())
	}
	return true
}

// GetDropsForCompatibleTool deliberately returns nothing, matching the PHP original's
// `return [];` (this isn't a not-yet-ported gap).
func (b *BaseFire) GetDropsForCompatibleTool(item Item) []Item { return nil }
