package craftingjson

import (
	"encoding/json"
	"fmt"
)

// ItemStackData is a port of pocketmine\crafting\json\ItemStackData. In the data files it's either
// an object or just the item name.
type ItemStackData struct {
	Name        string   `json:"name"`
	Count       *int     `json:"count,omitempty"`
	BlockStates *string  `json:"block_states,omitempty"`
	Meta        *int     `json:"meta,omitempty"`
	NBT         *string  `json:"nbt,omitempty"`
	CanPlaceOn  []string `json:"can_place_on,omitempty"`
	CanDestroy  []string `json:"can_destroy,omitempty"`
}

// UnmarshalJSON accepts the name-only string form (JsonMapper calls the one-argument constructor).
func (d *ItemStackData) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err == nil {
		*d = ItemStackData{Name: name}
		return nil
	}
	type plain ItemStackData
	var p plain
	if err := strictUnmarshal(data, &p); err != nil {
		return err
	}
	if p.Name == "" {
		return fmt.Errorf("ItemStackData: missing required name")
	}
	*d = ItemStackData(p)
	return nil
}

// MarshalJSON is ItemStackData::jsonSerialize: just the name when nothing else is set.
func (d ItemStackData) MarshalJSON() ([]byte, error) {
	if d.Count == nil && d.BlockStates == nil && d.Meta == nil && d.NBT == nil && d.CanPlaceOn == nil && d.CanDestroy == nil {
		return json.Marshal(d.Name)
	}
	type plain ItemStackData
	return json.Marshal(plain(d))
}
