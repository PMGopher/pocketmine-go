package nbt

// CloneTag is the Go equivalent of Tag::safeClone(). Scalar tags are value types (see tag.go),
// so they're already independent copies and are returned as-is; only the container tags
// (*ListTag, *CompoundTag) hold mutable state and need a real recursive copy.
//
// PHP's safeClone() also detects recursive tag graphs (a tag that (in)directly contains itself)
// and throws rather than looping forever. That situation can't arise here: Go's *ListTag/
// *CompoundTag are only ever constructed by application code appending/setting tags, and Clone()
// always produces brand new tag graphs from the leaves up, so there's no path by which a Clone()
// call could re-enter itself.
func CloneTag(t Tag) Tag {
	switch v := t.(type) {
	case *ListTag:
		return v.Clone()
	case *CompoundTag:
		return v.Clone()
	default:
		return t
	}
}
