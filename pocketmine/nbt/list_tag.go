package nbt

import (
	"fmt"
	"strings"
)

// ListTag is a port of pocketmine\nbt\tag\ListTag.
type ListTag struct {
	tagType Type
	values  []Tag
}

// NewListTag mirrors the ListTag constructor: values must all share the same Type (inferred from
// the first element if tagType is TagEnd).
func NewListTag(values []Tag, tagType Type) (*ListTag, error) {
	l := &ListTag{tagType: tagType}
	for _, v := range values {
		if err := l.Push(v); err != nil {
			return nil, err
		}
	}
	return l, nil
}

func (l *ListTag) Type() Type { return TagList }

// GetTagType returns the type of tag contained in this list.
func (l *ListTag) GetTagType() Type { return l.tagType }

func (l *ListTag) Count() int  { return len(l.values) }
func (l *ListTag) Empty() bool { return len(l.values) == 0 }

// Values returns a snapshot slice of the list's tags, in order.
func (l *ListTag) Values() []Tag {
	result := make([]Tag, len(l.values))
	copy(result, l.values)
	return result
}

func (l *ListTag) checkTagType(tag Tag) error {
	t := tag.Type()
	if t != l.tagType {
		if len(l.values) == 0 {
			l.tagType = t
		} else {
			return fmt.Errorf("invalid tag of type %s assigned to ListTag, expected %s", typeName(t), typeName(l.values[0].Type()))
		}
	}
	return nil
}

func (l *ListTag) Push(tag Tag) error {
	if err := l.checkTagType(tag); err != nil {
		return err
	}
	l.values = append(l.values, tag)
	return nil
}

func (l *ListTag) Pop() (Tag, error) {
	if len(l.values) == 0 {
		return nil, fmt.Errorf("list is empty")
	}
	last := l.values[len(l.values)-1]
	l.values = l.values[:len(l.values)-1]
	return last, nil
}

func (l *ListTag) Unshift(tag Tag) error {
	if err := l.checkTagType(tag); err != nil {
		return err
	}
	l.values = append([]Tag{tag}, l.values...)
	return nil
}

func (l *ListTag) Shift() (Tag, error) {
	if len(l.values) == 0 {
		return nil, fmt.Errorf("list is empty")
	}
	first := l.values[0]
	l.values = l.values[1:]
	return first, nil
}

// Insert inserts tag at offset, moving later values up by 1 position.
func (l *ListTag) Insert(offset int, tag Tag) error {
	if err := l.checkTagType(tag); err != nil {
		return err
	}
	if offset < 0 || offset > len(l.values) {
		return fmt.Errorf("offset cannot be negative or larger than the list's current size")
	}
	l.values = append(l.values[:offset], append([]Tag{tag}, l.values[offset:]...)...)
	return nil
}

// Remove removes the value at offset. Later tags are moved down by 1 position.
func (l *ListTag) Remove(offset int) {
	if offset < 0 || offset >= len(l.values) {
		return
	}
	l.values = append(l.values[:offset], l.values[offset+1:]...)
}

func (l *ListTag) Get(offset int) (Tag, error) {
	if offset < 0 || offset >= len(l.values) {
		return nil, fmt.Errorf("no such tag at offset %d", offset)
	}
	return l.values[offset], nil
}

func (l *ListTag) First() (Tag, error) {
	if len(l.values) == 0 {
		return nil, fmt.Errorf("list is empty")
	}
	return l.values[0], nil
}

func (l *ListTag) Last() (Tag, error) {
	if len(l.values) == 0 {
		return nil, fmt.Errorf("list is empty")
	}
	return l.values[len(l.values)-1], nil
}

func (l *ListTag) Set(offset int, tag Tag) error {
	if err := l.checkTagType(tag); err != nil {
		return err
	}
	if offset < 0 || offset > len(l.values) {
		return fmt.Errorf("offset cannot be negative or larger than the list's current size")
	}
	if offset == len(l.values) {
		l.values = append(l.values, tag)
	} else {
		l.values[offset] = tag
	}
	return nil
}

func (l *ListTag) IsSet(offset int) bool {
	return offset >= 0 && offset < len(l.values)
}

func readListTag(r StreamReader, tracker *ReaderTracker) (*ListTag, error) {
	tagTypeByte, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	tagType := Type(tagTypeByte)
	size, err := r.ReadInt()
	if err != nil {
		return nil, err
	}

	values := []Tag{}
	if size > 0 {
		if tagType == TagEnd {
			return nil, NewNbtDataException("Unexpected non-empty list of TAG_End")
		}
		err = tracker.ProtectDepth(func() error {
			for i := int32(0); i < size; i++ {
				tag, err := createTag(tagType, r, tracker)
				if err != nil {
					return err
				}
				values = append(values, tag)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return &ListTag{tagType: tagType, values: values}, nil
}

func (l *ListTag) write(w StreamWriter) error {
	w.WriteByte(uint8(l.tagType))
	w.WriteInt(int32(len(l.values)))
	for _, tag := range l.values {
		if err := tag.write(w); err != nil {
			return err
		}
	}
	return nil
}

func (l *ListTag) stringify(indentation int) string {
	var b strings.Builder
	b.WriteString("{\n")
	for _, tag := range l.values {
		b.WriteString(strings.Repeat("  ", indentation+1))
		b.WriteString(tag.stringify(indentation + 1))
		b.WriteString("\n")
	}
	b.WriteString(strings.Repeat("  ", indentation))
	b.WriteString("}")
	return b.String()
}

func (l *ListTag) String() string { return "TAG_List=" + l.stringify(0) }

// Clone returns a deep copy of l — the Go equivalent of PHP's __clone(), which recursively
// safeClone()s every child tag. Scalar child tags are value types and copy for free; only
// nested ListTag/CompoundTag children need an explicit recursive Clone.
func (l *ListTag) Clone() *ListTag {
	values := make([]Tag, len(l.values))
	for i, tag := range l.values {
		values[i] = CloneTag(tag)
	}
	return &ListTag{tagType: l.tagType, values: values}
}

func (l *ListTag) Equals(that Tag) bool {
	o, ok := that.(*ListTag)
	if !ok || len(o.values) != len(l.values) {
		return false
	}
	for i, v := range l.values {
		if !v.Equals(o.values[i]) {
			return false
		}
	}
	return true
}
