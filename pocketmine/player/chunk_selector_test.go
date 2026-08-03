package player

import "testing"

func collectChunks(radius, centerX, centerZ int) [][2]int {
	var got [][2]int
	for c := range SelectChunks(radius, centerX, centerZ) {
		got = append(got, c)
	}
	return got
}

func TestSelectChunksRadius1YieldsTheFourChunksAroundTheCenterPoint(t *testing.T) {
	got := collectChunks(1, 0, 0)
	want := map[[2]int]bool{
		{0, 0}: true, {-1, 0}: true, {0, -1}: true, {-1, -1}: true,
	}
	if len(got) != len(want) {
		t.Fatalf("collectChunks(1,0,0) = %v, want exactly %d chunks", got, len(want))
	}
	for _, c := range got {
		if !want[c] {
			t.Errorf("unexpected chunk %v in radius-1 selection", c)
		}
	}
}

func TestSelectChunksNeverYieldsTheSameChunkTwice(t *testing.T) {
	seen := map[[2]int]bool{}
	for _, c := range collectChunks(8, 5, -3) {
		if seen[c] {
			t.Fatalf("chunk %v yielded more than once", c)
		}
		seen[c] = true
	}
}

func TestSelectChunksLargerRadiusIsASupersetOfASmallerOne(t *testing.T) {
	small := collectChunks(3, 0, 0)
	large := collectChunks(6, 0, 0)

	if len(large) <= len(small) {
		t.Fatalf("radius-6 selection (%d chunks) is not larger than radius-3 (%d chunks)", len(large), len(small))
	}

	largeSet := map[[2]int]bool{}
	for _, c := range large {
		largeSet[c] = true
	}
	for _, c := range small {
		if !largeSet[c] {
			t.Errorf("chunk %v from the radius-3 selection is missing from the radius-6 selection", c)
		}
	}
}

func TestSelectChunksIsCenteredOnTheGivenCoordinates(t *testing.T) {
	origin := collectChunks(4, 0, 0)
	shifted := collectChunks(4, 100, 200)

	if len(origin) != len(shifted) {
		t.Fatalf("shifting the center changed the chunk count: %d vs %d", len(origin), len(shifted))
	}
	for i, c := range origin {
		want := [2]int{c[0] + 100, c[1] + 200}
		if shifted[i] != want {
			t.Errorf("chunk %d: shifted = %v, want %v (same relative offset)", i, shifted[i], want)
		}
	}
}

func TestSelectChunksStopsEarlyWhenTheConsumerBreaks(t *testing.T) {
	count := 0
	for range SelectChunks(8, 0, 0) {
		count++
		if count == 3 {
			break
		}
	}
	if count != 3 {
		t.Errorf("count = %d, want exactly 3 (the loop should have stopped at the break)", count)
	}
}
