package item

// EnchantedBook is a port of pocketmine\item\EnchantedBook.
type EnchantedBook struct {
	ItemBase
}

func NewEnchantedBook(identifier ItemIdentifier, name string, enchantmentTags ...string) *EnchantedBook {
	e := &EnchantedBook{}
	e.Init(e, identifier, name)
	e.enchantmentTags = enchantmentTags
	return e
}

func (e *EnchantedBook) Clone() Item {
	c := *e
	c.rebind(&c)
	return &c
}

func (e *EnchantedBook) GetMaxStackSize() int { return 1 }
