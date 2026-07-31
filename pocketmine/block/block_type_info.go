package block

// BlockTypeInfo is a port of pocketmine\block\BlockTypeInfo.
type BlockTypeInfo struct {
	breakInfo       *BlockBreakInfo
	typeTags        map[string]bool
	enchantmentTags []string
}

func NewBlockTypeInfo(breakInfo *BlockBreakInfo, typeTags []string, enchantmentTags []string) *BlockTypeInfo {
	tags := make(map[string]bool, len(typeTags))
	for _, t := range typeTags {
		tags[t] = true
	}
	return &BlockTypeInfo{breakInfo: breakInfo, typeTags: tags, enchantmentTags: enchantmentTags}
}

func (t *BlockTypeInfo) GetBreakInfo() *BlockBreakInfo { return t.breakInfo }

func (t *BlockTypeInfo) GetTypeTags() []string {
	result := make([]string, 0, len(t.typeTags))
	for tag := range t.typeTags {
		result = append(result, tag)
	}
	return result
}

func (t *BlockTypeInfo) HasTypeTag(tag string) bool { return t.typeTags[tag] }

func (t *BlockTypeInfo) GetEnchantmentTags() []string { return t.enchantmentTags }
