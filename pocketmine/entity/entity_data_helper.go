package entity

import (
	"fmt"
	stdmath "math"

	"pocketmine-go/pocketmine/data"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/world"
)

// NBT keys shared by every entity, a port of Entity's TAG_* constants.
const (
	tagFire              = "Fire"              //TAG_Short
	tagOnGround          = "OnGround"          //TAG_Byte
	tagFallDistance      = "FallDistance"      //TAG_Float
	tagCustomName        = "CustomName"        //TAG_String
	tagCustomNameVisible = "CustomNameVisible" //TAG_Byte
	TagPos               = "Pos"               //TAG_List<TAG_Double>|TAG_List<TAG_Float>
	TagMotion            = "Motion"            //TAG_List<TAG_Double>|TAG_List<TAG_Float>
	TagRotation          = "Rotation"          //TAG_List<TAG_Float>
)

func validateFloat(tagName, component string, value float64) error {
	if stdmath.IsInf(value, 0) {
		return data.NewSavedDataLoadingError(fmt.Sprintf("%s component of '%s' contains invalid infinite value", component, tagName))
	}
	if stdmath.IsNaN(value) {
		return data.NewSavedDataLoadingError(fmt.Sprintf("%s component of '%s' contains invalid NaN value", component, tagName))
	}
	return nil
}

// ParseLocation is a port of EntityDataHelper::parseLocation.
func ParseLocation(tag *nbt.CompoundTag, w *world.World) (Location, error) {
	pos, err := ParseVec3(tag, TagPos, false)
	if err != nil {
		return Location{}, err
	}

	generic, ok := tag.GetTag(TagRotation)
	list, isList := generic.(*nbt.ListTag)
	if !ok || !isList || (list.GetTagType() != nbt.TagFloat && list.Count() > 0) {
		return Location{}, data.NewSavedDataLoadingError(fmt.Sprintf("'%s' should be a List<Float>", TagRotation))
	}
	values := list.Values()
	if len(values) != 2 {
		return Location{}, data.NewSavedDataLoadingError("Expected exactly 2 entries for 'Rotation'")
	}
	yaw := float64(values[0].(nbt.FloatTag))
	pitch := float64(values[1].(nbt.FloatTag))
	if err := validateFloat(TagRotation, "yaw", yaw); err != nil {
		return Location{}, err
	}
	if err := validateFloat(TagRotation, "pitch", pitch); err != nil {
		return Location{}, err
	}

	return LocationFromObject(pos, w, yaw, pitch), nil
}

// ParseVec3 is a port of EntityDataHelper::parseVec3. A missing optional tag parses as zero.
func ParseVec3(tag *nbt.CompoundTag, tagName string, optional bool) (math.Vector3, error) {
	generic, ok := tag.GetTag(tagName)
	if !ok && optional {
		return math.Vector3Zero(), nil
	}
	list, isList := generic.(*nbt.ListTag)
	if !ok || !isList || (list.Count() > 0 && list.GetTagType() != nbt.TagDouble && list.GetTagType() != nbt.TagFloat) {
		return math.Vector3{}, data.NewSavedDataLoadingError(fmt.Sprintf("'%s' should be a List<Double> or List<Float>", tagName))
	}
	values := list.Values()
	if len(values) != 3 {
		return math.Vector3{}, data.NewSavedDataLoadingError(fmt.Sprintf("Expected exactly 3 entries in '%s' tag", tagName))
	}
	components := make([]float64, 3)
	for i, v := range values {
		switch n := v.(type) {
		case nbt.DoubleTag:
			components[i] = float64(n)
		case nbt.FloatTag:
			components[i] = float64(n)
		}
	}
	for i, name := range []string{"x", "y", "z"} {
		if err := validateFloat(tagName, name, components[i]); err != nil {
			return math.Vector3{}, err
		}
	}
	return math.NewVector3(components[0], components[1], components[2]), nil
}
