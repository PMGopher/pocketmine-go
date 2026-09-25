package item

// ProjectileItem is a port of the abstract pocketmine\item\ProjectileItem: items thrown as
// projectile entities (snowballs, eggs, ender pearls, ...). Throwing (onClickAir) is in
// item_use.go; the entity is created by the entity/projectile package (ThrowProjectileFunc).
type ProjectileItem interface {
	Item
	GetThrowForce() float64
}

// throwableItem is the shared state of the simple throwable projectile items.
type throwableItem struct {
	ItemBase

	maxStackSize int
	throwForce   float64
	cooldown     int
	cooldownTag  string
}

func (t *throwableItem) GetMaxStackSize() int { return t.maxStackSize }

func (t *throwableItem) GetThrowForce() float64 { return t.throwForce }

func (t *throwableItem) GetCooldownTicks() int { return t.cooldown }

func (t *throwableItem) GetCooldownTag() (string, bool) { return t.cooldownTag, t.cooldownTag != "" }

// Egg is a port of pocketmine\item\Egg.
type Egg struct{ throwableItem }

func NewEgg(identifier ItemIdentifier, name string) *Egg {
	e := &Egg{throwableItem{maxStackSize: 16, throwForce: 1.5}}
	e.Init(e, identifier, name)
	return e
}

func (e *Egg) Clone() Item {
	c := *e
	c.rebind(&c)
	return &c
}

// Snowball is a port of pocketmine\item\Snowball.
type Snowball struct{ throwableItem }

func NewSnowball(identifier ItemIdentifier, name string) *Snowball {
	s := &Snowball{throwableItem{maxStackSize: 16, throwForce: 1.5}}
	s.Init(s, identifier, name)
	return s
}

func (s *Snowball) Clone() Item {
	c := *s
	c.rebind(&c)
	return &c
}

// IceBomb is a port of pocketmine\item\IceBomb.
type IceBomb struct{ throwableItem }

func NewIceBomb(identifier ItemIdentifier, name string) *IceBomb {
	i := &IceBomb{throwableItem{maxStackSize: 16, throwForce: 1.5, cooldown: 10}}
	i.Init(i, identifier, name)
	return i
}

func (i *IceBomb) Clone() Item {
	c := *i
	c.rebind(&c)
	return &c
}

// ItemCooldownTagEnderPearl mirrors ItemCooldownTags::ENDER_PEARL.
const ItemCooldownTagEnderPearl = "ender_pearl"

// EnderPearl is a port of pocketmine\item\EnderPearl.
type EnderPearl struct{ throwableItem }

func NewEnderPearl(identifier ItemIdentifier, name string) *EnderPearl {
	e := &EnderPearl{throwableItem{maxStackSize: 16, throwForce: 1.5, cooldown: 20, cooldownTag: ItemCooldownTagEnderPearl}}
	e.Init(e, identifier, name)
	return e
}

func (e *EnderPearl) Clone() Item {
	c := *e
	c.rebind(&c)
	return &c
}

// ExperienceBottle is a port of pocketmine\item\ExperienceBottle.
type ExperienceBottle struct{ throwableItem }

func NewExperienceBottle(identifier ItemIdentifier, name string) *ExperienceBottle {
	e := &ExperienceBottle{throwableItem{maxStackSize: 64, throwForce: 0.7}}
	e.Init(e, identifier, name)
	return e
}

func (e *ExperienceBottle) Clone() Item {
	c := *e
	c.rebind(&c)
	return &c
}
