package sound

// NoteInstrument is a port of pocketmine\world\sound\NoteInstrument.
type NoteInstrument int

const (
	NoteInstrumentPiano NoteInstrument = iota
	NoteInstrumentBassDrum
	NoteInstrumentSnare
	NoteInstrumentClicksAndSticks
	NoteInstrumentDoubleBass
	NoteInstrumentBell
	NoteInstrumentFlute
	NoteInstrumentChime
	NoteInstrumentGuitar
	NoteInstrumentXylophone
	NoteInstrumentIronXylophone
	NoteInstrumentCowBell
	NoteInstrumentDidgeridoo
	NoteInstrumentBit
	NoteInstrumentBanjo
	NoteInstrumentPling
)

// noteInstrumentIDs is pocketmine\data\bedrock\NoteInstrumentIdMap. It lives here because NoteSound
// needs it and data/bedrock can't be imported by this package's importers without a cycle.
var noteInstrumentIDs = map[NoteInstrument]int{
	NoteInstrumentPiano:           0,
	NoteInstrumentBassDrum:        1,
	NoteInstrumentSnare:           2,
	NoteInstrumentClicksAndSticks: 3,
	NoteInstrumentDoubleBass:      4,
	NoteInstrumentFlute:           5,
	NoteInstrumentBell:            6,
	NoteInstrumentGuitar:          7,
	NoteInstrumentChime:           8,
	NoteInstrumentXylophone:       9,
	NoteInstrumentIronXylophone:   10,
	NoteInstrumentCowBell:         11,
	NoteInstrumentDidgeridoo:      12,
	NoteInstrumentBit:             13,
	NoteInstrumentBanjo:           14,
	NoteInstrumentPling:           15,
}

// NoteInstrumentToID is NoteInstrumentIdMap::toId.
func NoteInstrumentToID(instrument NoteInstrument) int { return noteInstrumentIDs[instrument] }

// NoteInstrumentFromID is NoteInstrumentIdMap::fromId.
func NoteInstrumentFromID(id int) (NoteInstrument, bool) {
	for instrument, v := range noteInstrumentIDs {
		if v == id {
			return instrument, true
		}
	}
	return 0, false
}
