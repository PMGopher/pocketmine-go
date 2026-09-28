package nbt

import (
	"strings"
	"testing"
)

// The error cases are JsonNbtParserTest.php from pmmp/NBT.
func TestParseJsonErrors(t *testing.T) {
	for input, want := range map[string]string{
		"{SomeTag:[]":                  "unexpected end of stream",
		"{SomeTag:[":                   "unexpected end of stream",
		"{SomeTag:[]}}":                "unexpected trailing characters",
		"[1,2,3]":                      "expected compound start",
		"dsfhjfughfuy{string:string} ": "expected compound start",
		"{TestList:[1f, string2, 3b]}": "lists can only contain one type of value",
		"{Test:hi,Test:hi}":            "duplicate compound leaf node",
		"{Test:hi,Test:1}":             "duplicate compound leaf node",
		"{Test:300b}":                  "Data error: Value 300 is outside the allowed range -128 - 127",
		"":                             "Syntax error: Not enough bytes left in buffer",
	} {
		_, err := ParseJson(input)
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%q: error %v, want %q", input, err, want)
		}
	}
}

func TestParseJson(t *testing.T) {
	tag, err := ParseJson("{}")
	if err != nil || tag.Count() != 0 {
		t.Fatalf("empty compound: %v %v", tag, err)
	}

	tag, err = ParseJson("{TestList:[]}")
	if err != nil {
		t.Fatal(err)
	}
	if list, ok := tag.GetTag("TestList"); !ok || list.(*ListTag).Count() != 0 {
		t.Errorf("empty list: %v", list)
	}

	tag, err = ParseJson("{\"String With Spaces\": 1}")
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := tag.GetTag("String With Spaces"); v != IntTag(1) {
		t.Errorf("quoted key: %v", v)
	}

	tag, err = ParseJson("{TestString:\"  TEST  minecraft:stone  \"}")
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := tag.GetTag("TestString"); v != StringTag("  TEST  minecraft:stone  ") {
		t.Errorf("quoted value: %v", v)
	}

	tag, err = ParseJson(`{display:{Name:"Sword",Lore:["a","b"]},b:1b,s:2s,l:3l,f:1.5,d:2d,n:-4,e:1e2}`)
	if err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]Tag{"b": ByteTag(1), "s": ShortTag(2), "l": LongTag(3), "f": FloatTag(1.5), "d": DoubleTag(2), "n": IntTag(-4), "e": FloatTag(100)} {
		if v, _ := tag.GetTag(name); v != want {
			t.Errorf("%s = %v, want %v", name, v, want)
		}
	}
	display, _ := tag.GetTag("display")
	if name, _ := display.(*CompoundTag).GetTag("Name"); name != StringTag("Sword") {
		t.Errorf("display.Name = %v", name)
	}
}
