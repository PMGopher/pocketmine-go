package nbt

import (
	"fmt"
	"iter"
	"strings"
)

// CompoundTag is a port of pocketmine\nbt\tag\CompoundTag.
//
// PHP associative arrays preserve insertion order, and PocketMine relies on that for stable
// write() output, so this keeps an explicit order slice alongside the lookup map — the same
// pattern used by utils.ObjectSet, for the same reason (a plain Go map has no defined order).
type CompoundTag struct {
	values map[string]Tag
	order  []string
}

func NewCompoundTag() *CompoundTag {
	return &CompoundTag{values: map[string]Tag{}}
}

func (c *CompoundTag) Type() Type { return TagCompound }
func (c *CompoundTag) Count() int { return len(c.order) }

// GetTag returns the tag with the specified name, or (nil, false) if it does not exist.
func (c *CompoundTag) GetTag(name string) (Tag, bool) {
	t, ok := c.values[name]
	return t, ok
}

// GetListTag returns the ListTag with the specified name, or (nil, false) if it does not exist.
// err is non-nil if a tag exists with that name but isn't a ListTag.
func (c *CompoundTag) GetListTag(name string) (*ListTag, bool, error) {
	tag, ok := c.GetTag(name)
	if !ok {
		return nil, false, nil
	}
	list, ok := tag.(*ListTag)
	if !ok {
		return nil, false, NewUnexpectedTagTypeException(fmt.Sprintf("Expected a tag of type ListTag, got %T", tag))
	}
	return list, true, nil
}

// GetCompoundTag returns the CompoundTag with the specified name, or (nil, false) if it does not
// exist. err is non-nil if a tag exists with that name but isn't a CompoundTag.
func (c *CompoundTag) GetCompoundTag(name string) (*CompoundTag, bool, error) {
	tag, ok := c.GetTag(name)
	if !ok {
		return nil, false, nil
	}
	compound, ok := tag.(*CompoundTag)
	if !ok {
		return nil, false, NewUnexpectedTagTypeException(fmt.Sprintf("Expected a tag of type CompoundTag, got %T", tag))
	}
	return compound, true, nil
}

// SetTag sets the specified Tag as a child tag at the given name, returning c for chaining.
func (c *CompoundTag) SetTag(name string, tag Tag) *CompoundTag {
	if len(name) > 32767 {
		panic(fmt.Sprintf("Tag name must be at most 32767 bytes, but got %d bytes", len(name)))
	}
	if _, exists := c.values[name]; !exists {
		c.order = append(c.order, name)
	}
	c.values[name] = tag
	return c
}

// RemoveTag removes the child tags with the specified names.
func (c *CompoundTag) RemoveTag(names ...string) {
	for _, name := range names {
		if _, exists := c.values[name]; !exists {
			continue
		}
		delete(c.values, name)
		for i, n := range c.order {
			if n == name {
				c.order = append(c.order[:i], c.order[i+1:]...)
				break
			}
		}
	}
}

func getTagValue[T Tag](c *CompoundTag, name string) (T, error) {
	var zero T
	tag, ok := c.GetTag(name)
	if !ok {
		return zero, NewNoSuchTagException(fmt.Sprintf("Tag %q does not exist", name))
	}
	v, ok := tag.(T)
	if !ok {
		return zero, NewUnexpectedTagTypeException(fmt.Sprintf("Expected a tag of type %T, got %T", zero, tag))
	}
	return v, nil
}

func getTagValueOr[T Tag](c *CompoundTag, name string, def T) T {
	v, err := getTagValue[T](c, name)
	if err != nil {
		return def
	}
	return v
}

func (c *CompoundTag) GetByte(name string) (ByteTag, error)       { return getTagValue[ByteTag](c, name) }
func (c *CompoundTag) GetByteOr(name string, def ByteTag) ByteTag { return getTagValueOr(c, name, def) }

func (c *CompoundTag) GetShort(name string) (ShortTag, error) { return getTagValue[ShortTag](c, name) }
func (c *CompoundTag) GetShortOr(name string, def ShortTag) ShortTag {
	return getTagValueOr(c, name, def)
}

func (c *CompoundTag) GetInt(name string) (IntTag, error)      { return getTagValue[IntTag](c, name) }
func (c *CompoundTag) GetIntOr(name string, def IntTag) IntTag { return getTagValueOr(c, name, def) }

func (c *CompoundTag) GetLong(name string) (LongTag, error) { return getTagValue[LongTag](c, name) }
func (c *CompoundTag) GetLongOr(name string, def LongTag) LongTag {
	return getTagValueOr(c, name, def)
}

func (c *CompoundTag) GetFloat(name string) (FloatTag, error) { return getTagValue[FloatTag](c, name) }
func (c *CompoundTag) GetFloatOr(name string, def FloatTag) FloatTag {
	return getTagValueOr(c, name, def)
}

func (c *CompoundTag) GetDouble(name string) (DoubleTag, error) {
	return getTagValue[DoubleTag](c, name)
}
func (c *CompoundTag) GetDoubleOr(name string, def DoubleTag) DoubleTag {
	return getTagValueOr(c, name, def)
}

func (c *CompoundTag) GetByteArray(name string) (ByteArrayTag, error) {
	return getTagValue[ByteArrayTag](c, name)
}
func (c *CompoundTag) GetByteArrayOr(name string, def ByteArrayTag) ByteArrayTag {
	return getTagValueOr(c, name, def)
}

func (c *CompoundTag) GetString(name string) (StringTag, error) {
	return getTagValue[StringTag](c, name)
}
func (c *CompoundTag) GetStringOr(name string, def StringTag) StringTag {
	return getTagValueOr(c, name, def)
}

func (c *CompoundTag) GetIntArray(name string) (IntArrayTag, error) {
	return getTagValue[IntArrayTag](c, name)
}
func (c *CompoundTag) GetIntArrayOr(name string, def IntArrayTag) IntArrayTag {
	return getTagValueOr(c, name, def)
}

func (c *CompoundTag) SetByte(name string, v ByteTag) *CompoundTag     { return c.SetTag(name, v) }
func (c *CompoundTag) SetShort(name string, v ShortTag) *CompoundTag   { return c.SetTag(name, v) }
func (c *CompoundTag) SetInt(name string, v IntTag) *CompoundTag       { return c.SetTag(name, v) }
func (c *CompoundTag) SetLong(name string, v LongTag) *CompoundTag     { return c.SetTag(name, v) }
func (c *CompoundTag) SetFloat(name string, v FloatTag) *CompoundTag   { return c.SetTag(name, v) }
func (c *CompoundTag) SetDouble(name string, v DoubleTag) *CompoundTag { return c.SetTag(name, v) }
func (c *CompoundTag) SetByteArray(name string, v ByteArrayTag) *CompoundTag {
	return c.SetTag(name, v)
}
func (c *CompoundTag) SetString(name string, v StringTag) *CompoundTag     { return c.SetTag(name, v) }
func (c *CompoundTag) SetIntArray(name string, v IntArrayTag) *CompoundTag { return c.SetTag(name, v) }

func readCompoundTag(r StreamReader, tracker *ReaderTracker) (*CompoundTag, error) {
	result := NewCompoundTag()
	err := tracker.ProtectDepth(func() error {
		for {
			typeByte, err := r.ReadByte()
			if err != nil {
				return err
			}
			t := Type(typeByte)
			if t == TagEnd {
				return nil
			}
			name, err := r.ReadString()
			if err != nil {
				return err
			}
			tag, err := createTag(t, r, tracker)
			if err != nil {
				return err
			}
			if _, exists := result.GetTag(name); exists {
				// Technically a corruption case, but very common in old worlds (pretty much every
				// furnace in worlds saved before ~2017); we can't recover the borked data from the
				// rest in Anvil/McRegion worlds, so this can't be a hard error.
				continue
			}
			result.SetTag(name, tag)
		}
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *CompoundTag) write(w StreamWriter) error {
	for _, name := range c.order {
		tag := c.values[name]
		w.WriteByte(uint8(tag.Type()))
		if err := w.WriteString(name); err != nil {
			return err
		}
		if err := tag.write(w); err != nil {
			return err
		}
	}
	w.WriteByte(uint8(TagEnd))
	return nil
}

func (c *CompoundTag) stringify(indentation int) string {
	var b strings.Builder
	b.WriteString("{\n")
	for _, name := range c.order {
		b.WriteString(strings.Repeat("  ", indentation+1))
		b.WriteString(fmt.Sprintf("%q => ", name))
		b.WriteString(c.values[name].stringify(indentation + 1))
		b.WriteString("\n")
	}
	b.WriteString(strings.Repeat("  ", indentation))
	b.WriteString("}")
	return b.String()
}

func (c *CompoundTag) String() string { return "TAG_Compound=" + c.stringify(0) }

// Clone returns a deep copy of c. See ListTag.Clone's doc comment.
func (c *CompoundTag) Clone() *CompoundTag {
	clone := &CompoundTag{values: make(map[string]Tag, len(c.values)), order: append([]string(nil), c.order...)}
	for k, v := range c.values {
		clone.values[k] = CloneTag(v)
	}
	return clone
}

func (c *CompoundTag) Equals(that Tag) bool {
	o, ok := that.(*CompoundTag)
	if !ok || len(o.values) != len(c.values) {
		return false
	}
	for k, v := range c.values {
		other, exists := o.values[k]
		if !exists || !v.Equals(other) {
			return false
		}
	}
	return true
}

// Merge returns a copy of c with values from other merged in (deep-cloned). Tags present in
// both are overwritten by other's.
func (c *CompoundTag) Merge(other *CompoundTag) *CompoundTag {
	result := c.Clone()
	for _, name := range other.order {
		result.SetTag(name, CloneTag(other.values[name]))
	}
	return result
}

// All returns a range-over-func iterator over (name, tag) pairs in insertion order — the
// idiomatic Go replacement for PHP's IteratorAggregate.
func (c *CompoundTag) All() iter.Seq2[string, Tag] {
	return func(yield func(string, Tag) bool) {
		for _, name := range c.order {
			if !yield(name, c.values[name]) {
				return
			}
		}
	}
}
