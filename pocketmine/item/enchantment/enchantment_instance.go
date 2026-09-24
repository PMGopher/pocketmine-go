package enchantment

// EnchantmentInstance is a port of pocketmine\item\enchantment\EnchantmentInstance - a container
// for enchantment data applied to items.
type EnchantmentInstance struct {
	enchantment Enchantment
	level       int
}

// NewEnchantmentInstance is a port of EnchantmentInstance::__construct (PHP's default level is 1).
func NewEnchantmentInstance(enchantment Enchantment, level int) *EnchantmentInstance {
	return &EnchantmentInstance{enchantment: enchantment, level: level}
}

// GetType returns the type of this enchantment.
func (e *EnchantmentInstance) GetType() Enchantment { return e.enchantment }

// GetLevel returns the level of the enchantment.
func (e *EnchantmentInstance) GetLevel() int { return e.level }
