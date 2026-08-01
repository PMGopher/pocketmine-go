package block

// FletchingTable is a port of pocketmine\block\FletchingTable.
type FletchingTable struct {
	Opaque
}

func NewFletchingTable(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *FletchingTable {
	f := &FletchingTable{Opaque: Opaque{NewBlock(idInfo, name, typeInfo)}}
	f.Init(f)
	return f
}

func (f *FletchingTable) Clone() Behavior {
	c := *f
	c.rebind(&c)
	return &c
}

func (f *FletchingTable) GetFuelTime() int { return 300 }
