// Package craftingjson is a port of pocketmine\crafting\json: the models recipe data files (pmmp
// BedrockData's recipes/*.json) are read into.
package craftingjson

import (
	"encoding/json"
	"fmt"
)

// strictUnmarshal is JsonMapper with bExceptionOnUndefinedProperty: unknown fields are errors.
func strictUnmarshal(data []byte, v any) error {
	dec := json.NewDecoder(bytesReader(data))
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

// LoadArray decodes a data file holding a JSON array of T (strictly: unknown fields are errors),
// like CraftingManagerFromDataHelper::loadJsonArrayOfObjectsFile.
func LoadArray[T any](data []byte) ([]T, error) {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("root should be an array: %w", err)
	}
	result := make([]T, 0, len(raw))
	for i, r := range raw {
		var v T
		if err := strictUnmarshal(r, &v); err != nil {
			return nil, fmt.Errorf("Invalid entry at index %d: %w", i, err)
		}
		result = append(result, v)
	}
	return result, nil
}
