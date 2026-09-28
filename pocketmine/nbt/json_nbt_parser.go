package nbt

import (
	"errors"
	"fmt"
	stdmath "math"
	"regexp"
	"strconv"
	"strings"
)

// jsonNbtStream is the part of pocketmine\utils\BinaryStream JsonNbtParser uses.
type jsonNbtStream struct {
	data   string
	offset int
}

// binaryDataError is BinaryDataException: reading past the end of the stream.
type binaryDataError struct{ message string }

func (e *binaryDataError) Error() string { return e.message }

func (s *jsonNbtStream) feof() bool { return s.offset >= len(s.data) }

func (s *jsonNbtStream) get() (byte, error) {
	if s.offset >= len(s.data) {
		return 0, &binaryDataError{fmt.Sprintf("Not enough bytes left in buffer: need 1, have %d", len(s.data)-s.offset)}
	}
	c := s.data[s.offset]
	s.offset++
	return c, nil
}

// ParseJson is a port of pocketmine\nbt\JsonNbtParser::parseJson: parses JSON-formatted NBT (as
// used by the /give command) into a CompoundTag. Errors are *NbtDataException.
func ParseJson(data string) (*CompoundTag, error) {
	stream := &jsonNbtStream{data: strings.Trim(data, " \r\n\t")}

	ret, err := func() (*CompoundTag, error) {
		b, err := stream.get()
		if err != nil {
			return nil, err
		}
		if b != '{' {
			return nil, NewNbtDataException(fmt.Sprintf("Syntax error: expected compound start but got '%c'", b))
		}
		return jsonParseCompound(stream) //don't return directly, syntax needs to be validated
	}()
	if err != nil {
		var binErr *binaryDataError
		if errors.As(err, &binErr) {
			return nil, NewNbtDataException(fmt.Sprintf("Syntax error: %s at offset %d", binErr.message, stream.offset))
		}
		return nil, NewNbtDataException(fmt.Sprintf("%s at offset %d", err.Error(), stream.offset))
	}
	if !stream.feof() {
		return nil, NewNbtDataException("Syntax error: unexpected trailing characters after end of tag: " + stream.data[stream.offset:])
	}
	return ret, nil
}

func jsonParseList(stream *jsonNbtStream) (*ListTag, error) {
	retval, _ := NewListTag(nil, TagEnd)

	more, err := jsonSkipWhitespace(stream, ']')
	if err != nil {
		return nil, err
	}
	if !more {
		return retval, nil
	}
	for !stream.feof() {
		value, err := jsonReadValue(stream)
		if err != nil {
			var invalid *InvalidTagValueException
			if errors.As(err, &invalid) {
				return nil, NewNbtDataException("Data error: " + invalid.Message)
			}
			return nil, err
		}
		if expectedType := retval.GetTagType(); expectedType != TagEnd && expectedType != value.Type() {
			return nil, NewNbtDataException("Data error: lists can only contain one type of value")
		}
		if err := retval.Push(value); err != nil {
			return nil, NewNbtDataException("Data error: " + err.Error())
		}
		end, err := jsonReadBreak(stream, ']')
		if err != nil {
			return nil, err
		}
		if end {
			return retval, nil
		}
	}
	return nil, NewNbtDataException("Syntax error: unexpected end of stream")
}

func jsonParseCompound(stream *jsonNbtStream) (*CompoundTag, error) {
	retval := NewCompoundTag()

	more, err := jsonSkipWhitespace(stream, '}')
	if err != nil {
		return nil, err
	}
	if !more {
		return retval, nil
	}
	for !stream.feof() {
		k, err := jsonReadKey(stream)
		if err != nil {
			return nil, err
		}
		if _, exists := retval.GetTag(k); exists {
			return nil, NewNbtDataException("Syntax error: duplicate compound leaf node '" + k + "'")
		}
		value, err := jsonReadValue(stream)
		if err != nil {
			var invalid *InvalidTagValueException
			if errors.As(err, &invalid) {
				return nil, NewNbtDataException("Data error: " + invalid.Message)
			}
			return nil, err
		}
		retval.SetTag(k, value)

		end, err := jsonReadBreak(stream, '}')
		if err != nil {
			return nil, err
		}
		if end {
			return retval, nil
		}
	}
	return nil, NewNbtDataException("Syntax error: unexpected end of stream")
}

func jsonSkipWhitespace(stream *jsonNbtStream, terminator byte) (bool, error) {
	for !stream.feof() {
		b, _ := stream.get()
		if b == terminator {
			return false, nil
		}
		if b == ' ' || b == '\n' || b == '\t' || b == '\r' {
			continue
		}
		stream.offset--
		return true, nil
	}
	return false, NewNbtDataException("Syntax error: unexpected end of stream, expected start of key")
}

// jsonReadBreak returns true if the terminator has been found, false if a comma was found.
func jsonReadBreak(stream *jsonNbtStream, terminator byte) (bool, error) {
	if stream.feof() {
		return false, NewNbtDataException(fmt.Sprintf("Syntax error: unexpected end of stream, expected '%c'", terminator))
	}
	offset := stream.offset
	c, _ := stream.get()
	if c == ',' {
		return false, nil
	}
	if c == terminator {
		return true, nil
	}
	return false, NewNbtDataException(fmt.Sprintf("Syntax error: unexpected '%c' end at offset %d", c, offset))
}

func jsonReadValue(stream *jsonNbtStream) (Tag, error) {
	var value strings.Builder
	inQuotes := false
	offset := stream.offset
	foundEnd := false
	var retval Tag

	for !stream.feof() {
		offset = stream.offset
		c, _ := stream.get()

		if inQuotes { //anything is allowed inside quotes, except unescaped quotes
			if c == '"' {
				inQuotes = false
				tag, err := NewStringTag(value.String())
				if err != nil {
					return nil, err
				}
				retval = tag
				foundEnd = true
			} else if c == '\\' {
				escaped, err := stream.get()
				if err != nil {
					return nil, err
				}
				value.WriteByte(escaped)
			} else {
				value.WriteByte(c)
			}
			continue
		}

		if c == ',' || c == '}' || c == ']' { //end of parent tag
			stream.offset-- //the caller needs to be able to read this character
			foundEnd = true
			break
		}

		if value.Len() == 0 || foundEnd {
			if c == '\r' || c == '\n' || c == '\t' || c == ' ' { //leading or trailing whitespace, ignore it
				continue
			}
			if foundEnd { //unexpected non-whitespace character after end of value
				return nil, NewNbtDataException(fmt.Sprintf("Syntax error: unexpected '%c' after end of value at offset %d", c, offset))
			}
		}

		switch c {
		case '"': //start of quoted string
			if value.Len() != 0 {
				return nil, NewNbtDataException(fmt.Sprintf("Syntax error: unexpected quote at offset %d", offset))
			}
			inQuotes = true
		case '{': //start of compound tag
			if value.Len() != 0 {
				return nil, NewNbtDataException(fmt.Sprintf("Syntax error: unexpected compound start at offset %d (enclose in double quotes for literal)", offset))
			}
			compound, err := jsonParseCompound(stream)
			if err != nil {
				return nil, err
			}
			retval = compound
			foundEnd = true
		case '[': //start of list tag - TODO: arrays
			if value.Len() != 0 {
				return nil, NewNbtDataException(fmt.Sprintf("Syntax error: unexpected list start at offset %d (enclose in double quotes for literal)", offset))
			}
			list, err := jsonParseList(stream)
			if err != nil {
				return nil, err
			}
			retval = list
			foundEnd = true
		default: //any other character
			value.WriteByte(c)
		}
	}

	if retval != nil {
		return retval, nil
	}

	raw := value.String()
	if raw == "" {
		return nil, NewNbtDataException(fmt.Sprintf("Syntax error: empty value at offset %d", offset))
	}
	if !foundEnd {
		return nil, NewNbtDataException(fmt.Sprintf("Syntax error: unexpected end of stream at offset %d", offset))
	}

	last := strings.ToLower(raw[len(raw)-1:])
	part := raw[:len(raw)-1]
	if last != "b" && last != "s" && last != "l" && last != "f" && last != "d" {
		part = raw
		last = ""
	}

	if !phpIsNumeric(part) {
		return NewStringTag(raw)
	}
	if last == "f" || last == "d" || strings.Contains(part, ".") || strings.Contains(part, "e") { //e = scientific notation
		f, _ := strconv.ParseFloat(strings.TrimSpace(part), 64)
		if last == "d" {
			return DoubleTag(f), nil
		}
		return FloatTag(float32(f)), nil
	}

	v := phpIntCast(part)
	switch last {
	case "b":
		return integerTag(v, stdmath.MinInt8, stdmath.MaxInt8, func(v int64) Tag { return ByteTag(v) })
	case "s":
		return integerTag(v, stdmath.MinInt16, stdmath.MaxInt16, func(v int64) Tag { return ShortTag(v) })
	case "l":
		return LongTag(v), nil
	default:
		return integerTag(v, stdmath.MinInt32, stdmath.MaxInt32, func(v int64) Tag { return IntTag(v) })
	}
}

// integerTag is the range check of IntegerishTagTrait's constructor.
func integerTag(v, minV, maxV int64, make func(int64) Tag) (Tag, error) {
	if v < minV || v > maxV {
		return nil, NewInvalidTagValueException(fmt.Sprintf("Value %d is outside the allowed range %d - %d", v, minV, maxV))
	}
	return make(v), nil
}

var phpNumericPattern = regexp.MustCompile(`^[ \t\n\r\v\f]*[+-]?([0-9]+(\.[0-9]*)?|\.[0-9]+)([eE][+-]?[0-9]+)?[ \t\n\r\v\f]*$`)

// phpIsNumeric is PHP 8's is_numeric for strings.
func phpIsNumeric(s string) bool { return phpNumericPattern.MatchString(s) }

// phpIntCast is PHP's (int) cast of a numeric string: an integer string saturates at the int64
// limits, and one in scientific notation (e.g. "1E5", which has no lowercase "e") goes through a
// float.
func phpIntCast(s string) int64 {
	s = strings.TrimSpace(s)
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}
	f, _ := strconv.ParseFloat(s, 64)
	if f >= stdmath.MaxInt64 {
		return stdmath.MaxInt64
	}
	if f <= stdmath.MinInt64 {
		return stdmath.MinInt64
	}
	return int64(f)
}

func jsonReadKey(stream *jsonNbtStream) (string, error) {
	var key strings.Builder
	offset := stream.offset
	inQuotes := false
	foundEnd := false

	for !stream.feof() {
		c, _ := stream.get()

		if inQuotes {
			if c == '"' {
				inQuotes = false
				foundEnd = true
			} else if c == '\\' {
				escaped, err := stream.get()
				if err != nil {
					return "", err
				}
				key.WriteByte(escaped)
			} else {
				key.WriteByte(c)
			}
			continue
		}

		if c == ':' {
			foundEnd = true
			break
		}

		if key.Len() == 0 || foundEnd {
			if c == '\r' || c == '\n' || c == '\t' || c == ' ' { //leading or trailing whitespace, ignore it
				continue
			}
			if foundEnd { //unexpected non-whitespace character after end of value
				return "", NewNbtDataException(fmt.Sprintf("Syntax error: unexpected '%c' after end of value at offset %d", c, offset))
			}
		}

		switch c {
		case '"': //start of quoted string
			if key.Len() != 0 {
				return "", NewNbtDataException(fmt.Sprintf("Syntax error: unexpected quote at offset %d", offset))
			}
			inQuotes = true
		case '{', '}', '[', ']', ',':
			return "", NewNbtDataException(fmt.Sprintf("Syntax error: unexpected '%c' at offset %d (enclose in double quotes for literal)", c, offset))
		default: //any other character
			key.WriteByte(c)
		}
	}

	if key.Len() == 0 {
		return "", NewNbtDataException(fmt.Sprintf("Syntax error: invalid empty key at offset %d", offset))
	}
	if !foundEnd {
		return "", NewNbtDataException(fmt.Sprintf("Syntax error: unexpected end of stream at offset %d", offset))
	}
	return key.String(), nil
}
