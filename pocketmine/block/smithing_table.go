package block

import "pocketmine-go/pocketmine/math"

// SmithingTable is a port of pocketmine\block\SmithingTable.
type SmithingTable struct {
	Opaque
}

func NewSmithingTable(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *SmithingTable {
	s := &SmithingTable{Opaque: Opaque{NewBlock(idInfo, name, typeInfo)}}
	s.Init(s)
	return s
}

func (s *SmithingTable) Clone() Behavior {
	c := *s
	c.rebind(&c)
	return &c
}

func (s *SmithingTable) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	openWindow(player, WindowSmithingTable, s.position)
	return true
}

func (s *SmithingTable) GetFuelTime() int { return 300 }
