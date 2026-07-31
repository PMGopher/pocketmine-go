package permission

import (
	"fmt"
	"strings"
)

// PermissionParserException is thrown by PermissionParser when it encounters data it doesn't like.
type PermissionParserException struct{ Message string }

func (e *PermissionParserException) Error() string { return e.Message }

// PermissionParser is a port of pocketmine\permission\PermissionParser.
const (
	DefaultOp    = "op"
	DefaultNotOp = "notop"
	DefaultTrue  = "true"
	DefaultFalse = "false"
)

var defaultStringMap = map[string]string{
	"op": DefaultOp, "isop": DefaultOp, "operator": DefaultOp, "isoperator": DefaultOp,
	"admin": DefaultOp, "isadmin": DefaultOp,

	"!op": DefaultNotOp, "notop": DefaultNotOp, "!operator": DefaultNotOp,
	"notoperator": DefaultNotOp, "!admin": DefaultNotOp, "notadmin": DefaultNotOp,

	"true": DefaultTrue, "false": DefaultFalse,
}

// DefaultFromBool mirrors PermissionParser::defaultFromString() for the bool overload.
func DefaultFromBool(value bool) string {
	if value {
		return DefaultTrue
	}
	return DefaultFalse
}

// DefaultFromString mirrors PermissionParser::defaultFromString() for the string overload.
func DefaultFromString(value string) (string, error) {
	if v, ok := defaultStringMap[strings.ToLower(value)]; ok {
		return v, nil
	}
	return "", &PermissionParserException{Message: fmt.Sprintf("Unknown permission default name %q", value)}
}

// PermissionEntry mirrors the shape PermissionParser::loadPermissions() reads each entry from
// (a permissions.yml-style map with optional "default"/"description" keys — "children" is
// rejected, matching the PHP original: nested permission declarations are no longer supported).
type PermissionEntry struct {
	Name        string
	Default     string
	Description any
	HasChildren bool // if true, LoadPermissions returns an error for this entry
}

// LoadPermissions is a port of PermissionParser::loadPermissions(): groups permissions by their
// "default" bucket (op/notop/true/false).
//
// PHP iterates a single associative array where an entry's "default" carries forward to every
// later entry that doesn't specify its own — meaning the *order* entries appear in is load-bearing,
// not just their content. A Go map has no defined iteration order, so this takes an explicit
// ordered slice instead of a map (the same fix as PermissionAttachment's insertion-order tracking
// elsewhere in this package, for the same underlying reason).
func LoadPermissions(entries []PermissionEntry, defaultBucket string) (map[string][]*Permission, error) {
	result := map[string][]*Permission{}
	current := defaultBucket
	for _, entry := range entries {
		if entry.HasChildren {
			return nil, &PermissionParserException{Message: "Nested permission declarations are no longer supported. Declare each permission separately."}
		}
		if entry.Default != "" {
			resolved, err := DefaultFromString(entry.Default)
			if err != nil {
				return nil, err
			}
			current = resolved
		}
		result[current] = append(result[current], NewPermission(entry.Name, entry.Description, nil))
	}
	return result, nil
}
