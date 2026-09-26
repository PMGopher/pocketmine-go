package blockconvert

import (
	"strings"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/data/bedrock"
)

// BlockSerializerDeserializerRegistrar is a port of
// pocketmine\data\bedrock\block\convert\BlockSerializerDeserializerRegistrar: registers a block's
// serializer and deserializer together from one description.
type BlockSerializerDeserializerRegistrar struct {
	Deserializer *BlockStateToObjectDeserializer
	Serializer   *BlockObjectToStateSerializer
}

func NewBlockSerializerDeserializerRegistrar(deserializer *BlockStateToObjectDeserializer, serializer *BlockObjectToStateSerializer) *BlockSerializerDeserializerRegistrar {
	return &BlockSerializerDeserializerRegistrar{Deserializer: deserializer, Serializer: serializer}
}

// compileFlattenedIdPartMatrix is a port of BlockSerializerDeserializerRegistrar::compileFlattenedIdPartMatrix:
// every combination of the components' possible values.
func compileFlattenedIdPartMatrix(components []any) [][]string {
	var result [][]string
	for _, component := range components {
		var column []string
		switch c := component.(type) {
		case string:
			column = []string{c}
		case StringProperty:
			column = c.GetPossibleValues()
		}
		if len(result) == 0 {
			for _, value := range column {
				result = append(result, []string{value})
			}
		} else {
			var stepResult [][]string
			for _, parts := range result {
				for _, value := range column {
					stepPart := append(append([]string(nil), parts...), value)
					stepResult = append(stepResult, stepPart)
				}
			}
			result = stepResult
		}
	}
	return result
}

// serializeFlattenedID is a port of BlockSerializerDeserializerRegistrar::serializeFlattenedId.
func serializeFlattenedID(blk block.Behavior, idComponents []any) string {
	var id strings.Builder
	for _, infix := range idComponents {
		switch c := infix.(type) {
		case string:
			id.WriteString(c)
		case StringProperty:
			id.WriteString(c.SerializePlain(blk))
		}
	}
	return id.String()
}

// deserializeFlattenedID is a port of BlockSerializerDeserializerRegistrar::deserializeFlattenedId.
func deserializeFlattenedID(baseBlock block.Behavior, idComponents []any, idPropertyValues []string) block.Behavior {
	preparedBlock := baseBlock.Clone()
	for k, component := range idComponents {
		if prop, ok := component.(StringProperty); ok {
			prop.DeserializePlain(preparedBlock, idPropertyValues[k])
		}
	}
	return preparedBlock
}

// MapSimple is a port of BlockSerializerDeserializerRegistrar::mapSimple.
func (r *BlockSerializerDeserializerRegistrar) MapSimple(blk block.Behavior, id string) {
	r.Deserializer.MapSimple(id, func() block.Behavior { return blk.Clone() })
	r.Serializer.MapSimple(blk, id)
}

// MapFlattenedId is a port of BlockSerializerDeserializerRegistrar::mapFlattenedId.
func (r *BlockSerializerDeserializerRegistrar) MapFlattenedId(model *FlattenedIdModel) {
	blk := model.GetBlock()
	idComponents := model.GetIdComponents()
	if len(idComponents) == 0 {
		panic("No ID components provided")
	}
	properties := model.GetProperties()

	//This is a really cursed hack that lets us essentially write flattened properties as blockstate properties, and
	//then pull them out to compile an ID :D
	//This works surprisingly well and is much more elegant than I would've expected

	if len(properties) > 0 {
		r.Serializer.Map(blk, func(b block.Behavior) bedrock.BlockStateData {
			writer := NewBlockStateWriter(serializeFlattenedID(b, idComponents))
			for _, property := range properties {
				property.Serialize(b, writer)
			}
			return writer.GetBlockStateData()
		})
	} else {
		r.Serializer.Map(blk, func(b block.Behavior) bedrock.BlockStateData {
			//fast path for blocks with no state properties
			return NewBlockStateWriter(serializeFlattenedID(b, idComponents)).GetBlockStateData()
		})
	}

	for _, idParts := range compileFlattenedIdPartMatrix(idComponents) {
		//deconstruct the ID into a partial state
		//we can do this at registration time since there will be multiple deserializers
		preparedBlock := deserializeFlattenedID(blk, idComponents, idParts)
		id := strings.Join(idParts, "")

		if len(properties) > 0 {
			r.Deserializer.Map(id, func(reader *BlockStateReader) block.Behavior {
				b := preparedBlock.Clone()
				for _, property := range properties {
					property.Deserialize(b, reader)
				}
				return b
			})
		} else {
			//fast path for blocks with no state properties
			r.Deserializer.Map(id, func(*BlockStateReader) block.Behavior { return preparedBlock.Clone() })
		}
	}
}

// MapColored is a port of BlockSerializerDeserializerRegistrar::mapColored.
func (r *BlockSerializerDeserializerRegistrar) MapColored(blk block.Behavior, idPrefix, idSuffix string) {
	r.MapFlattenedId(NewFlattenedIdModel(blk).IdComponents(
		idPrefix,
		GetCommonProperties().DyeColorIdInfix,
		idSuffix,
	))
}

// MapSlab is a port of BlockSerializerDeserializerRegistrar::mapSlab.
func (r *BlockSerializerDeserializerRegistrar) MapSlab(blk block.Behavior, typ string) {
	c := GetCommonProperties()
	r.MapFlattenedId(NewFlattenedIdModel(blk).
		IdComponents("minecraft:", typ, "_", c.SlabIdInfix, "slab").
		Properties(c.SlabPositionProperty))
}

// MapStairs is a port of BlockSerializerDeserializerRegistrar::mapStairs.
func (r *BlockSerializerDeserializerRegistrar) MapStairs(blk block.Behavior, id string) {
	r.MapModel(NewModel(blk, id).Properties(GetCommonProperties().StairProperties...))
}

// MapModel is a port of BlockSerializerDeserializerRegistrar::mapModel.
func (r *BlockSerializerDeserializerRegistrar) MapModel(model *Model) {
	id := model.GetID()
	blk := model.GetBlock()
	propertyDescriptors := model.GetProperties()

	r.Deserializer.Map(id, func(in *BlockStateReader) block.Behavior {
		newBlock := blk.Clone()
		for _, descriptor := range propertyDescriptors {
			descriptor.Deserialize(newBlock, in)
		}
		return newBlock
	})
	r.Serializer.Map(blk, func(b block.Behavior) bedrock.BlockStateData {
		writer := NewBlockStateWriter(id)
		for _, descriptor := range propertyDescriptors {
			descriptor.Serialize(b, writer)
		}
		return writer.GetBlockStateData()
	})
}
