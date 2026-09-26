package block

import (
	"pocketmine-go/pocketmine/block/tile"
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/sound"
)

// Record is a forward-compatible marker for pocketmine\item\Record - same pattern as
// WritableBookBase in lectern.go. GetRecordType returns a plain blockutils.RecordType (not a
// self-referential Item), so unlike ItemFrame's insertion gap this is satisfied automatically by
// the real item.Record.
type Record interface {
	Item
	GetRecordType() blockutils.RecordType
}

// Jukebox is a port of pocketmine\block\Jukebox.
//
// Redstone output from having a record inserted isn't implemented in the PHP original either
// (marked with its own TODO), so that isn't a gap introduced by this port.
type Jukebox struct {
	Opaque

	RecordItem Record
}

func NewJukebox(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Jukebox {
	j := &Jukebox{Opaque: Opaque{NewBlock(idInfo, name, typeInfo)}}
	j.Init(j)
	return j
}

func (j *Jukebox) Clone() Behavior {
	c := *j
	c.rebind(&c)
	return &c
}

func (j *Jukebox) GetFuelTime() int { return 300 }

// OnInteract is a port of Jukebox::onInteract.
func (j *Jukebox) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	if player != nil {
		if j.RecordItem != nil {
			j.EjectRecord()
		} else if record, ok := item.(Record); ok {
			if p, ok := player.(interface{ SendJukeboxPopup(message any) }); ok {
				p.SendJukeboxPopup(lang.KnownTranslationFactory.RecordNowPlaying(record.GetRecordType().GetTranslatableName()))
			}
			if popped, ok := popItem(item).(Record); ok {
				j.InsertRecord(popped)
			}
		}
	}
	j.setSelf()
	return true
}

func (j *Jukebox) GetRecord() Record { return j.RecordItem }

// EjectRecord is a port of Jukebox::ejectRecord.
func (j *Jukebox) EjectRecord() {
	if j.RecordItem != nil {
		if world, err := j.position.GetWorld(); err == nil {
			dropItem(world, j.position.Add(0.5, 1, 0.5), j.RecordItem)
		}
		j.RecordItem = nil
		j.StopSound()
	}
}

// InsertRecord is a port of Jukebox::insertRecord.
func (j *Jukebox) InsertRecord(record Record) {
	if j.RecordItem == nil {
		j.RecordItem = record
		j.StartSound()
	}
}

func (j *Jukebox) StartSound() {
	if j.RecordItem != nil {
		j.addSound(sound.RecordSound{RecordType: j.RecordItem.GetRecordType()})
	}
}

func (j *Jukebox) StopSound() {
	j.addSound(sound.RecordStopSound{})
}

func (j *Jukebox) OnBreak(item Item, player Player, returnedItems *[]Item) bool {
	j.StopSound()
	return j.Block.OnBreak(item, player, returnedItems)
}

func (j *Jukebox) GetDropsForCompatibleTool(item Item) []Item {
	drops := j.Block.GetDropsForCompatibleTool(item)
	if j.RecordItem != nil {
		drops = append(drops, j.RecordItem)
	}
	return drops
}

// ReadStateFromWorld is a port of Jukebox::readStateFromWorld.
func (j *Jukebox) ReadStateFromWorld() Behavior {
	j.Block.ReadStateFromWorld()
	world, err := j.position.GetWorld()
	if err != nil {
		return j.self
	}
	t, ok := world.GetTile(j.position)
	if !ok {
		return j.self
	}
	tileJukebox, ok := t.(*tile.Jukebox)
	if !ok {
		return j.self
	}
	j.RecordItem = nil
	if tileRecord, has := tileJukebox.GetRecord(); has {
		if r, ok := tileRecord.(Record); ok {
			j.RecordItem = r
		}
	}
	return j.self
}

func (j *Jukebox) addSound(s sound.Sound) {
	world, err := j.position.GetWorld()
	if err != nil {
		return
	}
	world.AddSound(j.position.Vector3, s)
}

func (j *Jukebox) setSelf() {
	world, err := j.position.GetWorld()
	if err != nil {
		return
	}
	_ = world.SetBlock(j.position, j.self)
}

// WriteStateToWorld is a port of Jukebox::writeStateToWorld.
func (j *Jukebox) WriteStateToWorld() {
	j.Block.WriteStateToWorld()
	if t, ok := j.tileAt(); ok {
		if tileJukebox, ok := t.(*tile.Jukebox); ok {
			tileJukebox.SetRecord(asTileItem(j.RecordItem))
		}
	}
}

func init() {
	tile.RecordStopSoundFunc = func(t tile.Tile) {
		pos := t.GetPosition()
		if w, ok := pos.GetWorld(); ok {
			if world, ok := w.(World); ok {
				world.AddSound(pos.Vector3, sound.RecordStopSound{})
			}
		}
	}
}
