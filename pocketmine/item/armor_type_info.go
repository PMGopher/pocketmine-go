package item

// ArmorTypeInfo is a port of pocketmine\item\ArmorTypeInfo. The PHP constructor's Material
// parameter defaults to VanillaArmorMaterials::LEATHER() when omitted - that registry isn't
// ported, so NewArmorTypeInfo requires an explicit ArmorMaterial instead.
type ArmorTypeInfo struct {
	DefensePoints int
	MaxDurability int
	ArmorSlot     int
	Toughness     int
	FireProof     bool
	Material      ArmorMaterial
}

func NewArmorTypeInfo(defensePoints, maxDurability, armorSlot int, material ArmorMaterial) ArmorTypeInfo {
	return ArmorTypeInfo{DefensePoints: defensePoints, MaxDurability: maxDurability, ArmorSlot: armorSlot, Material: material}
}

func (i ArmorTypeInfo) GetDefensePoints() int { return i.DefensePoints }

func (i ArmorTypeInfo) GetMaxDurability() int { return i.MaxDurability }

func (i ArmorTypeInfo) GetArmorSlot() int { return i.ArmorSlot }

func (i ArmorTypeInfo) GetToughness() int { return i.Toughness }

func (i ArmorTypeInfo) IsFireProof() bool { return i.FireProof }

func (i ArmorTypeInfo) GetMaterial() ArmorMaterial { return i.Material }
