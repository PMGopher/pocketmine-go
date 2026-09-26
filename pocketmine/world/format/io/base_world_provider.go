package io

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"pocketmine-go/pocketmine/block"

	blockconvert "pocketmine-go/pocketmine/data/bedrock/block/convert"
	blockupgrade "pocketmine-go/pocketmine/data/bedrock/block/upgrade"
	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/world/format"
)

// WorldError is a port of pocketmine\world\WorldException, for errors this package raises itself.
type WorldError struct{ Message string }

func (e *WorldError) Error() string { return e.Message }

// BaseWorldProvider is a port of pocketmine\world\format\io\BaseWorldProvider. Concrete providers
// embed it; InitBase is its constructor.
type BaseWorldProvider struct {
	Path      string
	Logger    log.Logger
	worldData WorldData

	BlockStateDeserializer *blockconvert.BlockStateToObjectDeserializer
	BlockDataUpgrader      *blockupgrade.BlockDataUpgrader
	BlockStateSerializer   *blockconvert.BlockObjectToStateSerializer
}

// InitBase is BaseWorldProvider::__construct: loadLevelData is the concrete provider's
// loadLevelData().
func (p *BaseWorldProvider) InitBase(path string, logger log.Logger, loadLevelData func() (WorldData, error)) error {
	if _, err := os.Stat(path); err != nil {
		return &WorldError{"World does not exist"}
	}
	p.Path = path
	p.Logger = logger

	//TODO: this should not rely on singletons
	p.BlockStateDeserializer = GetBlockStateDeserializer()
	p.BlockDataUpgrader = GetBlockDataUpgrader()
	p.BlockStateSerializer = GetBlockStateSerializer()

	worldData, err := loadLevelData()
	if err != nil {
		return err
	}
	p.worldData = worldData
	return nil
}

// unknownStateID is the block every unreadable state becomes (GlobalBlockStateHandlers'
// unknown block state, the "update!" block).
func (p *BaseWorldProvider) unknownStateID() int32 {
	id, err := p.BlockStateDeserializer.Deserialize(GetUnknownBlockStateData())
	if err != nil {
		panic("the unknown block state must always deserialize: " + err.Error())
	}
	return int32(id)
}

// translatePalette is a port of BaseWorldProvider::translatePalette: upgrades a palette of legacy
// (id << 4 | meta) values to block state IDs.
func (p *BaseWorldProvider) translatePalette(blockArray *format.PalettedBlockArray, logger log.Logger) *format.PalettedBlockArray {
	palette := blockArray.GetPalette()
	newPalette := make([]int32, 0, len(palette))
	var blockDecodeErrors []string
	for k, legacyIdMeta := range palette {
		//TODO: remember data for unknown states so we can implement them later
		id := int(legacyIdMeta >> 4)
		meta := int(legacyIdMeta & 0xf)
		newStateData, err := p.BlockDataUpgrader.UpgradeIntIdMeta(id, meta)
		if err != nil {
			blockDecodeErrors = append(blockDecodeErrors, fmt.Sprintf("Palette offset %d / Failed to upgrade legacy ID/meta %d:%d: %v", k, id, meta, err))
			newStateData = GetUnknownBlockStateData()
		}
		stateID, err := p.BlockStateDeserializer.Deserialize(newStateData)
		if err != nil {
			// this should never happen anyway - if the upgrader returned an invalid state, we have bigger problems
			blockDecodeErrors = append(blockDecodeErrors, fmt.Sprintf("Palette offset %d / Failed to deserialize upgraded state %d:%d: %v", k, id, meta, err))
			stateID = int(p.unknownStateID())
		}
		newPalette = append(newPalette, int32(stateID))
	}
	if len(blockDecodeErrors) > 0 {
		logger.Error("Errors decoding/upgrading blocks:\n - " + strings.Join(blockDecodeErrors, "\n - "))
	}
	result, err := format.NewPalettedBlockArrayFromRaw(blockArray.GetBitsPerBlock(), blockArray.GetWordArray(), newPalette)
	if err != nil {
		panic(err)
	}
	return result
}

// PalettizeLegacySubChunkXZY is a port of BaseWorldProvider::palettizeLegacySubChunkXZY.
func (p *BaseWorldProvider) PalettizeLegacySubChunkXZY(idArray, metaArray []byte, logger log.Logger) *format.PalettedBlockArray {
	return p.translatePalette(ConvertSubChunkXZY(idArray, metaArray), logger)
}

// PalettizeLegacySubChunkYZX is a port of BaseWorldProvider::palettizeLegacySubChunkYZX.
func (p *BaseWorldProvider) PalettizeLegacySubChunkYZX(idArray, metaArray []byte, logger log.Logger) *format.PalettedBlockArray {
	return p.translatePalette(ConvertSubChunkYZX(idArray, metaArray), logger)
}

// PalettizeLegacySubChunkFromColumn is a port of BaseWorldProvider::palettizeLegacySubChunkFromColumn.
func (p *BaseWorldProvider) PalettizeLegacySubChunkFromColumn(idArray, metaArray []byte, yOffset int, logger log.Logger) *format.PalettedBlockArray {
	return p.translatePalette(ConvertSubChunkFromLegacyColumn(idArray, metaArray, yOffset), logger)
}

// GetPath is a port of BaseWorldProvider::getPath.
func (p *BaseWorldProvider) GetPath() string { return p.Path }

// GetWorldData is a port of BaseWorldProvider::getWorldData.
func (p *BaseWorldProvider) GetWorldData() WorldData { return p.worldData }

// EmptyStateID is Block::EMPTY_STATE_ID: the state ID of air, which fills empty subchunks.
func EmptyStateID() int32 {
	emptyStateIDOnce.Do(func() { emptyStateID = int32(block.VanillaAir().GetStateId()) })
	return emptyStateID
}

var (
	emptyStateIDOnce sync.Once
	emptyStateID     int32
)
