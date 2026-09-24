package enchantment

import "sync"

// IncompatibleEnchantmentGroups constants, a port of
// pocketmine\item\enchantment\IncompatibleEnchantmentGroups.
const (
	IncompatibleGroupProtection  = "protection"
	IncompatibleGroupBowInfinite = "bow_infinite"
	IncompatibleGroupBlockDrops  = "block_drops"
)

// IncompatibleEnchantmentRegistry is a port of
// pocketmine\item\enchantment\IncompatibleEnchantmentRegistry - manages which enchantments are
// incompatible with each other. Enchantments belonging to the same incompatibility group cannot be
// applied side-by-side on the same item.
type IncompatibleEnchantmentRegistry struct {
	incompatibilityMap map[*EnchantmentBase]map[string]bool
}

var (
	incompatibleRegistryOnce     sync.Once
	incompatibleRegistryInstance *IncompatibleEnchantmentRegistry
)

// GetIncompatibleEnchantmentRegistry is the port of IncompatibleEnchantmentRegistry::getInstance().
func GetIncompatibleEnchantmentRegistry() *IncompatibleEnchantmentRegistry {
	incompatibleRegistryOnce.Do(func() {
		r := &IncompatibleEnchantmentRegistry{incompatibilityMap: map[*EnchantmentBase]map[string]bool{}}
		incompatibleRegistryInstance = r
		r.Register(IncompatibleGroupProtection, []Enchantment{VanillaProtection(), VanillaFireProtection(), VanillaBlastProtection(), VanillaProjectileProtection()})
		r.Register(IncompatibleGroupBowInfinite, []Enchantment{VanillaInfinity(), VanillaMending()})
		r.Register(IncompatibleGroupBlockDrops, []Enchantment{VanillaFortune(), VanillaSilkTouch()})
	})
	return incompatibleRegistryInstance
}

// Register registers incompatibility for all enchantments with a tag. Enchantments with the same
// tag cannot be applied side-by-side on the same item.
func (r *IncompatibleEnchantmentRegistry) Register(tag string, enchantments []Enchantment) {
	for _, e := range enchantments {
		key := compatibilityKeyOf(e)
		if r.incompatibilityMap[key] == nil {
			r.incompatibilityMap[key] = map[string]bool{}
		}
		r.incompatibilityMap[key][tag] = true
	}
}

// Unregister unregisters incompatibility for some enchantments with a particular tag.
func (r *IncompatibleEnchantmentRegistry) Unregister(tag string, enchantments []Enchantment) {
	for _, e := range enchantments {
		delete(r.incompatibilityMap[compatibilityKeyOf(e)], tag)
	}
}

// UnregisterAll unregisters incompatibility for all enchantments with a particular tag.
func (r *IncompatibleEnchantmentRegistry) UnregisterAll(tag string) {
	for _, tags := range r.incompatibilityMap {
		delete(tags, tag)
	}
}

// AreCompatible returns whether two enchantments can be applied to the same item.
func (r *IncompatibleEnchantmentRegistry) AreCompatible(first, second Enchantment) bool {
	return r.areCompatibleKeys(compatibilityKeyOf(first), compatibilityKeyOf(second))
}

func (r *IncompatibleEnchantmentRegistry) areCompatibleKeys(first, second *EnchantmentBase) bool {
	for tag := range r.incompatibilityMap[first] {
		if r.incompatibilityMap[second][tag] {
			return false
		}
	}
	return true
}
