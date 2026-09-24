package leveldb

import (
	"errors"
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"

	"pocketmine-go/pocketmine/nbt"
)

// entityNBTMaxDepth bounds nested NBT depth when reading entity data (PHP's readMultiple uses the
// serializer's default of 0 = unlimited; a bound protects against corrupted data).
const entityNBTMaxDepth = 512

// SaveEntities is the ENTITIES half of LevelDB::writeChunk (writeTags): every entity's NBT
// concatenated as little-endian NBT roots under the chunk's ENTITIES key, or the key deleted if
// there are none.
func SaveEntities(db *leveldb.DB, chunkX, chunkZ int32, entities []*nbt.CompoundTag) error {
	key := taggedKey(chunkX, chunkZ, tagEntities)
	if len(entities) == 0 {
		return db.Delete(key, nil)
	}
	roots := make([]*nbt.TreeRoot, 0, len(entities))
	for _, tag := range entities {
		root, err := nbt.NewTreeRoot(tag, "")
		if err != nil {
			return err
		}
		roots = append(roots, root)
	}
	data, err := nbt.NewLittleEndianSerializer().WriteMultiple(roots)
	if err != nil {
		return fmt.Errorf("leveldb: saving chunk (%d,%d) entities: %w", chunkX, chunkZ, err)
	}
	return db.Put(key, data, nil)
}

// LoadEntities is the ENTITIES half of LevelDB::readChunk: the NBT of every entity saved in the
// chunk (nil if none).
func LoadEntities(db *leveldb.DB, chunkX, chunkZ int32) ([]*nbt.CompoundTag, error) {
	data, err := db.Get(taggedKey(chunkX, chunkZ, tagEntities), nil)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	roots, err := nbt.NewLittleEndianSerializer().ReadMultiple(data, entityNBTMaxDepth)
	if err != nil {
		return nil, fmt.Errorf("leveldb: corrupted entity data in chunk (%d,%d): %w", chunkX, chunkZ, err)
	}
	result := make([]*nbt.CompoundTag, 0, len(roots))
	for _, root := range roots {
		tag, err := root.MustGetCompoundTag()
		if err != nil {
			return nil, fmt.Errorf("leveldb: corrupted entity data in chunk (%d,%d): %w", chunkX, chunkZ, err)
		}
		result = append(result, tag)
	}
	return result, nil
}
