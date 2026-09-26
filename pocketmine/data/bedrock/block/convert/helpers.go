package blockconvert

import (
	"fmt"
	"strconv"

	"pocketmine-go/pocketmine/block"
	blockutils "pocketmine-go/pocketmine/block/utils"
	ids "pocketmine-go/pocketmine/data/bedrock/block"
)

func itoa(i int) string { return strconv.Itoa(i) }

func typeName(v any) string { return fmt.Sprintf("%T", v) }

// The parts of BlockStateSerializerHelper/BlockStateDeserializerHelper the serializer's and
// deserializer's mapSlab/mapStairs/mapLog use.

// encodeSlab is a port of BlockStateSerializerHelper::encodeSlab.
func encodeSlab(blk slabLike, singleID, doubleID string) *BlockStateWriter {
	if blk.GetSlabType() == blockutils.SlabTypeDouble {
		//this is (intentionally) also written for double slabs (as zero) to maintain bug parity with MCPE
		return NewBlockStateWriter(doubleID).WriteSlabPosition(blockutils.SlabTypeBottom)
	}
	return NewBlockStateWriter(singleID).WriteSlabPosition(blk.GetSlabType())
}

// encodeStairs is a port of BlockStateSerializerHelper::encodeStairs.
func encodeStairs(blk block.Behavior, out *BlockStateWriter) *BlockStateWriter {
	return out.
		WriteBool(ids.UPSIDE_DOWN_BIT, as[stairLike](blk).IsUpsideDown()).
		WriteWeirdoHorizontalFacing(as[facingHolder](blk).GetFacing())
}

// encodeLog is a port of BlockStateSerializerHelper::encodeLog.
func encodeLog(blk block.Behavior, unstrippedID, strippedID string) *BlockStateWriter {
	id := unstrippedID
	if as[woodLike](blk).IsStripped() {
		id = strippedID
	}
	return NewBlockStateWriter(id).WritePillarAxis(as[axisHolder](blk).GetAxis())
}

// decodeSingleSlab is a port of BlockStateDeserializerHelper::decodeSingleSlab.
func decodeSingleSlab(blk block.Behavior, in *BlockStateReader) block.Behavior {
	as[slabLike](blk).SetSlabType(in.ReadSlabPosition())
	return blk
}

// decodeDoubleSlab is a port of BlockStateDeserializerHelper::decodeDoubleSlab.
func decodeDoubleSlab(blk block.Behavior, in *BlockStateReader) block.Behavior {
	in.Ignored(ids.MC_VERTICAL_HALF)
	as[slabLike](blk).SetSlabType(blockutils.SlabTypeDouble)
	return blk
}

// decodeStairs is a port of BlockStateDeserializerHelper::decodeStairs.
func decodeStairs(blk block.Behavior, in *BlockStateReader) block.Behavior {
	as[stairLike](blk).SetUpsideDown(in.ReadBool(ids.UPSIDE_DOWN_BIT))
	as[facingHolder](blk).SetFacing(in.ReadWeirdoHorizontalFacing())
	return blk
}

// decodeLog is a port of BlockStateDeserializerHelper::decodeLog.
func decodeLog(blk block.Behavior, stripped bool, in *BlockStateReader) block.Behavior {
	as[axisHolder](blk).SetAxis(in.ReadPillarAxis())
	as[woodLike](blk).SetStripped(stripped)
	return blk
}
