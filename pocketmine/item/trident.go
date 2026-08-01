package item

// Trident is a port of pocketmine\item\Trident. OnReleaseUsing/OnAttackEntity/OnDestroyBlock
// aren't ported (throwing/attacking need a real Player/Entity - see the Item interface's doc
// comment). CanStartUsingItem's body doesn't actually touch its Player parameter
// (`$this->damage < $this->getMaxDurability() - DAMAGE_ON_THROW`), but it's skipped anyway for
// consistency with the rest of that method family being excluded from Item here.
type Trident struct {
	Tool
}

func NewTrident(identifier ItemIdentifier, name string) *Trident {
	t := &Trident{}
	t.Init(t, identifier, name)
	return t
}

func (t *Trident) Clone() Item {
	c := *t
	c.rebind(&c)
	return &c
}

func (t *Trident) GetMaxDurability() int { return 251 }

func (t *Trident) GetAttackPoints() int { return 9 }
