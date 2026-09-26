package enchantment

import (
	"fmt"
	"slices"
	"sync"
)

// ItemEnchantmentTagRegistry is a port of pocketmine\item\enchantment\ItemEnchantmentTagRegistry:
// it tracks which tags are nested in other tags (e.g. ARMOR contains HELMET, CHESTPLATE, ...),
// so that an enchantment registered for ARMOR applies to an item tagged HELMET. Every
// registered tag is also nested in the internal TagAll.
type ItemEnchantmentTagRegistry struct {
	mu     sync.RWMutex
	tagMap map[string][]string
}

var (
	itemEnchantmentTagRegistry     *ItemEnchantmentTagRegistry
	itemEnchantmentTagRegistryOnce sync.Once
)

// GetItemEnchantmentTagRegistry is a port of ItemEnchantmentTagRegistry::getInstance.
func GetItemEnchantmentTagRegistry() *ItemEnchantmentTagRegistry {
	itemEnchantmentTagRegistryOnce.Do(func() {
		r := &ItemEnchantmentTagRegistry{tagMap: map[string][]string{}}
		r.Register(TagArmor, TagHelmet, TagChestplate, TagLeggings, TagBoots)
		r.Register(TagShield)
		r.Register(TagSword)
		r.Register(TagTrident)
		r.Register(TagBow)
		r.Register(TagCrossbow)
		r.Register(TagShears)
		r.Register(TagFlintAndSteel)
		r.Register(TagBlockTools, TagAxe, TagPickaxe, TagShovel, TagHoe)
		r.Register(TagFishingRod)
		r.Register(TagCarrotOnStick)
		r.Register(TagCompass)
		r.Register(TagMask)
		r.Register(TagElytra)
		r.Register(TagBrush)
		r.Register(TagWeapons, TagSword, TagTrident, TagBow, TagCrossbow, TagBlockTools)
		itemEnchantmentTagRegistry = r
	})
	return itemEnchantmentTagRegistry
}

// Register is a port of ItemEnchantmentTagRegistry::register: it registers tag (if needed) and
// nests nestedTags in it, registering those too if they aren't yet.
func (r *ItemEnchantmentTagRegistry) Register(tag string, nestedTags ...string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.register(tag, nestedTags)
}

func (r *ItemEnchantmentTagRegistry) register(tag string, nestedTags []string) {
	assertNotInternalTag(tag)

	for _, nestedTag := range nestedTags {
		if _, ok := r.tagMap[nestedTag]; !ok {
			r.register(nestedTag, nil)
		}
		r.tagMap[tag] = append(r.tagMap[tag], nestedTag)
	}

	if _, ok := r.tagMap[tag]; !ok {
		r.tagMap[tag] = []string{}
		r.tagMap[TagAll] = append(r.tagMap[TagAll], tag)
	}
}

// Unregister is a port of ItemEnchantmentTagRegistry::unregister: it removes tag and every
// nesting of it in other tags.
func (r *ItemEnchantmentTagRegistry) Unregister(tag string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.tagMap[tag]; !ok {
		return
	}
	assertNotInternalTag(tag)

	delete(r.tagMap, tag)

	for key, nestedTags := range r.tagMap {
		if i := slices.Index(nestedTags, tag); i != -1 {
			r.tagMap[key] = slices.Delete(slices.Clone(nestedTags), i, i+1)
		}
	}
}

// RemoveNested is a port of ItemEnchantmentTagRegistry::removeNested.
func (r *ItemEnchantmentTagRegistry) RemoveNested(tag string, nestedTags ...string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	assertNotInternalTag(tag)
	var remaining []string
	for _, nested := range r.tagMap[tag] {
		if !slices.Contains(nestedTags, nested) {
			remaining = append(remaining, nested)
		}
	}
	if remaining == nil {
		remaining = []string{}
	}
	r.tagMap[tag] = remaining
}

// GetNested is a port of ItemEnchantmentTagRegistry::getNested: the tags nested directly in tag.
func (r *ItemEnchantmentTagRegistry) GetNested(tag string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return slices.Clone(r.tagMap[tag])
}

// IsTagArrayIntersection is a port of ItemEnchantmentTagRegistry::isTagArrayIntersection: whether
// the two tag lists share at least one leaf tag (a tag with nothing nested in it).
func (r *ItemEnchantmentTagRegistry) IsTagArrayIntersection(firstTags, secondTags []string) bool {
	if len(firstTags) == 0 || len(secondTags) == 0 {
		return false
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	firstLeafTags := r.getLeafTagsForArray(firstTags)
	for tag := range r.getLeafTagsForArray(secondTags) {
		if _, ok := firstLeafTags[tag]; ok {
			return true
		}
	}
	return false
}

func (r *ItemEnchantmentTagRegistry) getLeafTagsForArray(tags []string) map[string]struct{} {
	result := map[string]struct{}{}
	for _, tag := range tags {
		for _, leaf := range r.getLeafTags(tag) {
			result[leaf] = struct{}{}
		}
	}
	return result
}

// getLeafTags is a port of ItemEnchantmentTagRegistry::getLeafTags.
func (r *ItemEnchantmentTagRegistry) getLeafTags(tag string) []string {
	var result []string
	tagsToHandle := []string{tag}

	for len(tagsToHandle) != 0 {
		currentTag := tagsToHandle[0]
		tagsToHandle = tagsToHandle[1:]
		nestedTags := r.tagMap[currentTag]

		if len(nestedTags) == 0 {
			result = append(result, currentTag)
		} else {
			tagsToHandle = append(tagsToHandle, nestedTags...)
		}
	}

	return result
}

func assertNotInternalTag(tag string) {
	if tag == TagAll {
		panic(fmt.Sprintf("Cannot perform any operations on the internal item enchantment tag '%s'", tag))
	}
}
