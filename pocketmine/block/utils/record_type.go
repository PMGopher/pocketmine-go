package blockutils

// RecordType is a port of pocketmine\block\utils\RecordType. GetTranslatableName isn't ported -
// it needs KnownTranslationFactory, a generated translation-string file not part of this port
// yet. GetSoundId is deprecated in the PHP original too and always returns 0.
type RecordType int

const (
	RecordTypeDisk13 RecordType = iota
	RecordTypeDisk5
	RecordTypeDiskCat
	RecordTypeDiskBlocks
	RecordTypeDiskChirp
	RecordTypeDiskCreator
	RecordTypeDiskCreatorMusicBox
	RecordTypeDiskFar
	RecordTypeDiskLavaChicken
	RecordTypeDiskMall
	RecordTypeDiskMellohi
	RecordTypeDiskOtherside
	RecordTypeDiskPigstep
	RecordTypeDiskPrecipice
	RecordTypeDiskRelic
	RecordTypeDiskStal
	RecordTypeDiskStrad
	RecordTypeDiskWard
	RecordTypeDisk11
	RecordTypeDiskWait
)

var recordTypeSoundNames = map[RecordType]string{
	RecordTypeDisk13:              "C418 - 13",
	RecordTypeDisk5:               "Samuel Åberg - 5",
	RecordTypeDiskCat:             "C418 - cat",
	RecordTypeDiskBlocks:          "C418 - blocks",
	RecordTypeDiskChirp:           "C418 - chirp",
	RecordTypeDiskCreator:         "Lena Raine - Creator",
	RecordTypeDiskCreatorMusicBox: "Lena Raine - Creator (Music Box)",
	RecordTypeDiskFar:             "C418 - far",
	RecordTypeDiskLavaChicken:     "Hyper Potions - Lava Chicken",
	RecordTypeDiskMall:            "C418 - mall",
	RecordTypeDiskMellohi:         "C418 - mellohi",
	RecordTypeDiskOtherside:       "Lena Raine - otherside",
	RecordTypeDiskPigstep:         "Lena Raine - Pigstep",
	RecordTypeDiskPrecipice:       "Aaron Cherof - Precipice",
	RecordTypeDiskRelic:           "Aaron Cherof - Relic",
	RecordTypeDiskStal:            "C418 - stal",
	RecordTypeDiskStrad:           "C418 - strad",
	RecordTypeDiskWard:            "C418 - ward",
	RecordTypeDisk11:              "C418 - 11",
	RecordTypeDiskWait:            "C418 - wait",
}

func (t RecordType) GetSoundName() string { return recordTypeSoundNames[t] }

// GetSoundId is deprecated in the PHP original and always returns 0.
func (t RecordType) GetSoundId() int { return 0 }
