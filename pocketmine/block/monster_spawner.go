package block

import (
	"math/rand"

	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
)

// MonsterSpawner is a port of pocketmine\block\MonsterSpawner.
type MonsterSpawner struct {
	Transparent
}

func NewMonsterSpawner(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *MonsterSpawner {
	m := &MonsterSpawner{Transparent{NewBlock(idInfo, name, typeInfo)}}
	m.Init(m)
	return m
}

func (m *MonsterSpawner) Clone() Behavior {
	c := *m
	c.rebind(&c)
	return &c
}

// GetDropsForCompatibleTool deliberately returns nothing, matching the PHP original's
// `return [];` (this isn't a not-yet-ported gap).
func (m *MonsterSpawner) GetDropsForCompatibleTool(item Item) []Item { return nil }

func (m *MonsterSpawner) GetXpDropAmount() int { return rand.Intn(29) + 15 } // 15-43

// OnScheduledUpdate is a TODO in the PHP original too - spawner tick logic isn't implemented
// upstream either.
func (m *MonsterSpawner) OnScheduledUpdate() {}

func (m *MonsterSpawner) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}
