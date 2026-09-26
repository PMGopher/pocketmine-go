package enchantment

import (
	"slices"
	"sync"
)

// TaggedItem is the part of item.Item that AvailableEnchantmentRegistry and EnchantingHelper use
// (this package can't import item: item imports enchantment).
type TaggedItem interface {
	GetEnchantmentTags() []string
	HasEnchantments() bool
}

// AvailableEnchantmentRegistry is a port of pocketmine\item\enchantment\AvailableEnchantmentRegistry:
// the enchantments that can be obtained in survival (enchanting table, anvil, fishing, ...) and
// the item tags each applies to. Primary tags make the enchantment available from the
// enchanting table; secondary tags only from other sources (anvil, loot).
type AvailableEnchantmentRegistry struct {
	mu sync.RWMutex
	// enchantments keeps PHP's registration order (it decides the enchanting table's weighted
	// random choice).
	enchantments      []Enchantment
	primaryItemTags   map[Enchantment][]string
	secondaryItemTags map[Enchantment][]string
}

var (
	availableEnchantmentRegistry     *AvailableEnchantmentRegistry
	availableEnchantmentRegistryOnce sync.Once
)

// GetAvailableEnchantmentRegistry is a port of AvailableEnchantmentRegistry::getInstance.
func GetAvailableEnchantmentRegistry() *AvailableEnchantmentRegistry {
	availableEnchantmentRegistryOnce.Do(func() {
		r := &AvailableEnchantmentRegistry{primaryItemTags: map[Enchantment][]string{}, secondaryItemTags: map[Enchantment][]string{}}
		r.Register(VanillaProtection(), []string{TagArmor}, nil)
		r.Register(VanillaFireProtection(), []string{TagArmor}, nil)
		r.Register(VanillaFeatherFalling(), []string{TagBoots}, nil)
		r.Register(VanillaBlastProtection(), []string{TagArmor}, nil)
		r.Register(VanillaProjectileProtection(), []string{TagArmor}, nil)
		r.Register(VanillaThorns(), []string{TagChestplate}, []string{TagHelmet, TagLeggings, TagBoots})
		r.Register(VanillaRespiration(), []string{TagHelmet}, nil)
		r.Register(VanillaAquaAffinity(), []string{TagHelmet}, nil)
		r.Register(VanillaFrostWalker(), nil, []string{TagBoots})
		r.Register(VanillaSharpness(), []string{TagSword, TagAxe}, nil)
		r.Register(VanillaKnockback(), []string{TagSword}, nil)
		r.Register(VanillaFireAspect(), []string{TagSword}, nil)
		r.Register(VanillaEfficiency(), []string{TagBlockTools}, []string{TagShears})
		r.Register(VanillaFortune(), []string{TagBlockTools}, nil)
		r.Register(VanillaSilkTouch(), []string{TagBlockTools}, []string{TagShears})
		r.Register(
			VanillaUnbreaking(),
			[]string{TagArmor, TagWeapons, TagFishingRod},
			[]string{TagShears, TagFlintAndSteel, TagShield, TagCarrotOnStick, TagElytra, TagBrush},
		)
		r.Register(VanillaPower(), []string{TagBow}, nil)
		r.Register(VanillaPunch(), []string{TagBow}, nil)
		r.Register(VanillaFlame(), []string{TagBow}, nil)
		r.Register(VanillaInfinity(), []string{TagBow}, nil)
		r.Register(
			VanillaMending(),
			nil,
			[]string{TagArmor, TagWeapons, TagFishingRod,
				TagShears, TagFlintAndSteel, TagShield, TagCarrotOnStick, TagElytra, TagBrush},
		)
		r.Register(VanillaVanishing(), nil, []string{TagAll})
		r.Register(VanillaSwiftSneak(), nil, []string{TagLeggings})
		availableEnchantmentRegistry = r
	})
	return availableEnchantmentRegistry
}

// Register is a port of AvailableEnchantmentRegistry::register.
func (r *AvailableEnchantmentRegistry) Register(enchantment Enchantment, primaryItemTags, secondaryItemTags []string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !slices.Contains(r.enchantments, enchantment) {
		r.enchantments = append(r.enchantments, enchantment)
	}
	r.primaryItemTags[enchantment] = slices.Clone(primaryItemTags)
	r.secondaryItemTags[enchantment] = slices.Clone(secondaryItemTags)
}

// Unregister is a port of AvailableEnchantmentRegistry::unregister.
func (r *AvailableEnchantmentRegistry) Unregister(enchantment Enchantment) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if i := slices.Index(r.enchantments, enchantment); i != -1 {
		r.enchantments = slices.Delete(r.enchantments, i, i+1)
	}
	delete(r.primaryItemTags, enchantment)
	delete(r.secondaryItemTags, enchantment)
}

// UnregisterAll is a port of AvailableEnchantmentRegistry::unregisterAll.
func (r *AvailableEnchantmentRegistry) UnregisterAll() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.enchantments = nil
	r.primaryItemTags = map[Enchantment][]string{}
	r.secondaryItemTags = map[Enchantment][]string{}
}

// IsRegistered is a port of AvailableEnchantmentRegistry::isRegistered.
func (r *AvailableEnchantmentRegistry) IsRegistered(enchantment Enchantment) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return slices.Contains(r.enchantments, enchantment)
}

// GetPrimaryItemTags is a port of AvailableEnchantmentRegistry::getPrimaryItemTags.
func (r *AvailableEnchantmentRegistry) GetPrimaryItemTags(enchantment Enchantment) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return slices.Clone(r.primaryItemTags[enchantment])
}

// SetPrimaryItemTags is a port of AvailableEnchantmentRegistry::setPrimaryItemTags.
func (r *AvailableEnchantmentRegistry) SetPrimaryItemTags(enchantment Enchantment, tags []string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !slices.Contains(r.enchantments, enchantment) {
		panic("Cannot set primary item tags for non-registered enchantment")
	}
	r.primaryItemTags[enchantment] = slices.Clone(tags)
}

// GetSecondaryItemTags is a port of AvailableEnchantmentRegistry::getSecondaryItemTags.
func (r *AvailableEnchantmentRegistry) GetSecondaryItemTags(enchantment Enchantment) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return slices.Clone(r.secondaryItemTags[enchantment])
}

// SetSecondaryItemTags is a port of AvailableEnchantmentRegistry::setSecondaryItemTags.
func (r *AvailableEnchantmentRegistry) SetSecondaryItemTags(enchantment Enchantment, tags []string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !slices.Contains(r.enchantments, enchantment) {
		panic("Cannot set secondary item tags for non-registered enchantment")
	}
	r.secondaryItemTags[enchantment] = slices.Clone(tags)
}

// GetPrimaryEnchantmentsForItem is a port of AvailableEnchantmentRegistry::getPrimaryEnchantmentsForItem:
// the enchantments the enchanting table can put on item.
func (r *AvailableEnchantmentRegistry) GetPrimaryEnchantmentsForItem(item TaggedItem) []Enchantment {
	itemTags := item.GetEnchantmentTags()
	if len(itemTags) == 0 || item.HasEnchantments() {
		return nil
	}

	tagRegistry := GetItemEnchantmentTagRegistry()
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []Enchantment
	for _, e := range r.enchantments {
		if tagRegistry.IsTagArrayIntersection(r.primaryItemTags[e], itemTags) {
			result = append(result, e)
		}
	}
	return result
}

// GetAllEnchantmentsForItem is a port of AvailableEnchantmentRegistry::getAllEnchantmentsForItem.
func (r *AvailableEnchantmentRegistry) GetAllEnchantmentsForItem(item TaggedItem) []Enchantment {
	if len(item.GetEnchantmentTags()) == 0 {
		return nil
	}

	var result []Enchantment
	for _, e := range r.GetAll() {
		if r.IsAvailableForItem(e, item) {
			result = append(result, e)
		}
	}
	return result
}

// IsAvailableForItem is a port of AvailableEnchantmentRegistry::isAvailableForItem.
func (r *AvailableEnchantmentRegistry) IsAvailableForItem(enchantment Enchantment, item TaggedItem) bool {
	itemTags := item.GetEnchantmentTags()
	tagRegistry := GetItemEnchantmentTagRegistry()

	return tagRegistry.IsTagArrayIntersection(r.GetPrimaryItemTags(enchantment), itemTags) ||
		tagRegistry.IsTagArrayIntersection(r.GetSecondaryItemTags(enchantment), itemTags)
}

// GetAll is a port of AvailableEnchantmentRegistry::getAll.
func (r *AvailableEnchantmentRegistry) GetAll() []Enchantment {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return slices.Clone(r.enchantments)
}
