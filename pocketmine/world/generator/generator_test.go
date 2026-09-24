package generator

import "testing"

func TestConvertSeed(t *testing.T) {
	cases := []struct {
		in   string
		want int64
		ok   bool
	}{
		{"", 0, false},
		{"0", 0, true},
		{"1234", 1234, true},
		{"-42", -42, true},
		// Not an integer: Utils::javaStringHash, which returns the unsigned 32-bit hash.
		{"404.4", 49504510, true},
		{"hello", 99162322, true},
	}
	for _, c := range cases {
		got, ok := ConvertSeed(c.in)
		if got != c.want || ok != c.ok {
			t.Errorf("ConvertSeed(%q) = (%d, %v), want (%d, %v)", c.in, got, ok, c.want, c.ok)
		}
	}
}
