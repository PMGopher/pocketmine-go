package item

import "pocketmine-go/pocketmine/block"

// The durability-loss overrides of onDestroyBlock/onAttackEntity on the tools.

func damageUnlessInstant(blk block.Behavior, apply func(int) bool, amount int) bool {
	if !blk.GetBreakInfo().BreaksInstantly() {
		return apply(amount)
	}
	return false
}

// OnDestroyBlock is a port of Axe::onDestroyBlock.
func (a *Axe) OnDestroyBlock(blk block.Behavior, returnedItems *[]Item) bool {
	return damageUnlessInstant(blk, a.ApplyDamage, 1)
}

// OnAttackEntity is a port of Axe::onAttackEntity.
func (a *Axe) OnAttackEntity(victim Entity, returnedItems *[]Item) bool { return a.ApplyDamage(2) }

// OnDestroyBlock is a port of Pickaxe::onDestroyBlock.
func (p *Pickaxe) OnDestroyBlock(blk block.Behavior, returnedItems *[]Item) bool {
	return damageUnlessInstant(blk, p.ApplyDamage, 1)
}

// OnAttackEntity is a port of Pickaxe::onAttackEntity.
func (p *Pickaxe) OnAttackEntity(victim Entity, returnedItems *[]Item) bool {
	return p.ApplyDamage(2)
}

// OnDestroyBlock is a port of Shovel::onDestroyBlock.
func (s *Shovel) OnDestroyBlock(blk block.Behavior, returnedItems *[]Item) bool {
	return damageUnlessInstant(blk, s.ApplyDamage, 1)
}

// OnAttackEntity is a port of Shovel::onAttackEntity.
func (s *Shovel) OnAttackEntity(victim Entity, returnedItems *[]Item) bool {
	return s.ApplyDamage(2)
}

// OnDestroyBlock is a port of Hoe::onDestroyBlock.
func (h *Hoe) OnDestroyBlock(blk block.Behavior, returnedItems *[]Item) bool {
	return damageUnlessInstant(blk, h.ApplyDamage, 1)
}

// OnAttackEntity is a port of Hoe::onAttackEntity.
func (h *Hoe) OnAttackEntity(victim Entity, returnedItems *[]Item) bool { return h.ApplyDamage(1) }

// OnDestroyBlock is a port of Sword::onDestroyBlock.
func (s *Sword) OnDestroyBlock(blk block.Behavior, returnedItems *[]Item) bool {
	return damageUnlessInstant(blk, s.ApplyDamage, 2)
}

// OnAttackEntity is a port of Sword::onAttackEntity.
func (s *Sword) OnAttackEntity(victim Entity, returnedItems *[]Item) bool { return s.ApplyDamage(1) }

// OnDestroyBlock is a port of Shears::onDestroyBlock.
func (s *Shears) OnDestroyBlock(blk block.Behavior, returnedItems *[]Item) bool {
	return s.ApplyDamage(1)
}
