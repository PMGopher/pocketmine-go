package nbt

import "fmt"

// TreeRoot is a port of pocketmine\nbt\TreeRoot: wraps the root Tag of an NBT file/blob together
// with its (often empty) name, since NBT root tags can have names too.
type TreeRoot struct {
	tag  Tag
	name string
}

func NewTreeRoot(tag Tag, name string) (*TreeRoot, error) {
	if len(name) > 32767 {
		return nil, fmt.Errorf("tag name must be at most 32767 bytes, but got %d bytes", len(name))
	}
	return &TreeRoot{tag: tag, name: name}, nil
}

func (r *TreeRoot) GetTag() Tag     { return r.tag }
func (r *TreeRoot) GetName() string { return r.name }

// MustGetCompoundTag is a helper for the common case of an NBT file/blob with a Compound root.
func (r *TreeRoot) MustGetCompoundTag() (*CompoundTag, error) {
	if c, ok := r.tag.(*CompoundTag); ok {
		return c, nil
	}
	return nil, NewNbtDataException("Root is not a TAG_Compound")
}

func (r *TreeRoot) Equals(that *TreeRoot) bool {
	return r.name == that.name && r.tag.Equals(that.tag)
}

func (r *TreeRoot) String() string {
	prefix := ""
	if r.name != "" {
		prefix = fmt.Sprintf("%q => ", r.name)
	}
	return "ROOT {\n  " + prefix + r.tag.stringify(1) + "\n}"
}
