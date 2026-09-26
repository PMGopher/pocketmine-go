package leveldb

// SubChunkVersion is a port of pocketmine\world\format\io\leveldb\SubChunkVersion: the version
// byte at the start of a subchunk's data.
const (
	SubChunkVersionClassic        = 0
	SubChunkVersionPalettedSingle = 1

	// The following are not used by vanilla, but treated the same as version 0 due to a legacy
	// converter which erroneously used the version byte as subchunk height.
	SubChunkVersionClassicBug2 = 2
	SubChunkVersionClassicBug3 = 3
	SubChunkVersionClassicBug4 = 4
	SubChunkVersionClassicBug5 = 5
	SubChunkVersionClassicBug6 = 6
	SubChunkVersionClassicBug7 = 7

	// SubChunkVersionPalettedMulti is paletted with layers: almost identical to v1, but includes a
	// length prefix and 0 or more storages. First seen in 1.4 Update Aquatic to support water
	// inside other blocks.
	SubChunkVersionPalettedMulti = 8
	// SubChunkVersionPalettedMultiWithOffset is identical to v8 except for a height byte after the
	// layer count byte. First seen in 1.18 for Caves and Cliffs.
	SubChunkVersionPalettedMultiWithOffset = 9
)
