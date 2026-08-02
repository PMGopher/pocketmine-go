package object

import "pocketmine-go/pocketmine/block"

// OreType is a port of pocketmine\world\generator\object\OreType.
type OreType struct {
	Material     block.Behavior
	Replaces     block.Behavior
	ClusterCount int
	ClusterSize  int
	MinHeight    int
	MaxHeight    int
}

func NewOreType(material, replaces block.Behavior, clusterCount, clusterSize, minHeight, maxHeight int) *OreType {
	return &OreType{Material: material, Replaces: replaces, ClusterCount: clusterCount, ClusterSize: clusterSize, MinHeight: minHeight, MaxHeight: maxHeight}
}
