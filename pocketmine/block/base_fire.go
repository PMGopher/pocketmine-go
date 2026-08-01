package block

// BaseFire is a port of pocketmine\block\BaseFire. Like Crops/Stem, this isn't meant to be
// instantiated directly - a concrete leaf type (Fire, SoulFire) must embed it and implement
// Clone.
type BaseFire struct {
	Flowable
}

func (b *BaseFire) HasEntityCollision() bool { return true }

func (b *BaseFire) CanBeReplaced() bool { return true }

// OnEntityInside is a port of BaseFire::onEntityInside. The damage half (EntityDamageByBlockEvent
// + Entity.Attack) needs machinery not ported yet, so it's skipped; the ignite half runs
// unconditionally (EntityCombustByBlockEvent is treated as always-uncancelled, same convention as
// every other deferred concrete event in this port), including for arrows - the PHP original's
// Arrow-immunity check is dropped since there's no Arrow marker yet, a minor behavioural gap
// versus a full stub.
func (b *BaseFire) OnEntityInside(entity Entity) bool {
	entity.SetOnFire(8)
	return true
}

// GetDropsForCompatibleTool deliberately returns nothing, matching the PHP original's
// `return [];` (this isn't a not-yet-ported gap).
func (b *BaseFire) GetDropsForCompatibleTool(item Item) []Item { return nil }
