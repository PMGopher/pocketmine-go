package enchantment

// EnchantingOption is a port of pocketmine\item\enchantment\EnchantingOption: one of the three
// options shown in the enchanting table.
type EnchantingOption struct {
	requiredXpLevel int
	displayName     string
	enchantments    []*EnchantmentInstance
}

func NewEnchantingOption(requiredXpLevel int, displayName string, enchantments []*EnchantmentInstance) *EnchantingOption {
	return &EnchantingOption{requiredXpLevel: requiredXpLevel, displayName: displayName, enchantments: enchantments}
}

// GetRequiredXpLevel returns the minimum XP level required to select this enchantment option.
// It's NOT the number of XP levels that will be subtracted after enchanting.
func (o *EnchantingOption) GetRequiredXpLevel() int { return o.requiredXpLevel }

// GetDisplayName returns the name that will be translated to the 'Standard Galactic Alphabet'
// client-side. This can be any arbitrary text string, since the vanilla client cannot read the
// text anyway. Example: 'bless creature range free'.
func (o *EnchantingOption) GetDisplayName() string { return o.displayName }

// GetEnchantments returns the enchantments that will be applied to the item when this option is
// clicked.
func (o *EnchantingOption) GetEnchantments() []*EnchantmentInstance { return o.enchantments }
