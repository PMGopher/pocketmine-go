package runtime

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

// describeState is a stand-in for what a real Block.describeBlockOnlyState() would look like:
// one function, called with any DataDescriber, that both reads and writes depending on which
// concrete implementation is passed in — exactly the PHP pattern this package replicates.
func describeState(d DataDescriber, facing *math.Facing, powered *bool, age *int) {
	d.Facing(facing)
	d.Bool(powered)
	d.BoundedIntAuto(0, 15, age)
}

func TestSizeCalculatorMatchesWriterBits(t *testing.T) {
	facing := math.North
	powered := true
	age := 7

	calc := NewSizeCalculator()
	describeState(calc, &facing, &powered, &age)

	w := NewWriter(64)
	describeState(w, &facing, &powered, &age)

	if calc.GetBitsUsed() != w.GetOffset() {
		t.Fatalf("SizeCalculator bits = %d, Writer bits = %d, want equal", calc.GetBitsUsed(), w.GetOffset())
	}
}

func TestWriteThenReadRoundTrips(t *testing.T) {
	facing := math.East
	powered := true
	age := 9

	w := NewWriter(64)
	describeState(w, &facing, &powered, &age)

	r := NewReader(64, w.GetValue())
	var facing2 math.Facing
	var powered2 bool
	var age2 int
	describeState(r, &facing2, &powered2, &age2)

	if facing2 != facing || powered2 != powered || age2 != age {
		t.Fatalf("round trip = (%v, %v, %v), want (%v, %v, %v)", facing2, powered2, age2, facing, powered, age)
	}
	if r.GetOffset() != w.GetOffset() {
		t.Fatalf("Reader consumed %d bits, Writer produced %d bits", r.GetOffset(), w.GetOffset())
	}
}

func TestAxisEncodingOrderIsXZY(t *testing.T) {
	// The runtime encoding's bit order for axes (0=X, 1=Z, 2=Y) is independent of Axis's own
	// underlying iota values — this pins down that specific mapping.
	for _, tc := range []struct {
		axis math.Axis
		bits int
	}{
		{math.AxisX, 0},
		{math.AxisZ, 1},
		{math.AxisY, 2},
	} {
		w := NewWriter(2)
		axis := tc.axis
		w.Axis(&axis)
		if w.GetValue() != tc.bits {
			t.Fatalf("Axis(%v) encoded as %d, want %d", tc.axis, w.GetValue(), tc.bits)
		}
	}
}

func TestBoundedIntAutoRejectsOutOfRange(t *testing.T) {
	w := NewWriter(8)
	defer func() {
		if recover() == nil {
			t.Fatalf("expected a panic writing an out-of-range bounded int")
		}
	}()
	v := 20
	w.BoundedIntAuto(0, 15, &v)
}

func TestReaderRejectsInsufficientBits(t *testing.T) {
	r := NewReader(4, 0)
	if _, err := r.ReadInt(5); err == nil {
		t.Fatalf("expected an error reading more bits than available")
	}
}

func TestFacingFlagsRoundTrip(t *testing.T) {
	faces := []math.Facing{math.Up, math.North, math.East}

	w := NewWriter(len(math.AllFacing))
	w.FacingFlags(&faces)

	r := NewReader(len(math.AllFacing), w.GetValue())
	var got []math.Facing
	r.FacingFlags(&got)

	want := map[math.Facing]bool{math.Up: true, math.North: true, math.East: true}
	if len(got) != len(want) {
		t.Fatalf("got %v, want the set %v", got, want)
	}
	for _, f := range got {
		if !want[f] {
			t.Fatalf("unexpected facing %v in round-tripped set %v", f, got)
		}
	}
}
