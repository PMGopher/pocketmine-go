package utils

import "testing"

// Values checked against ExperienceUtils in PocketMine-MP.
func TestGetXpToReachLevel(t *testing.T) {
	for level, want := range map[int]int{0: 0, 1: 7, 16: 352, 17: 394, 30: 1395, 31: 1507, 32: 1628, 40: 2920} {
		if got := GetXpToReachLevel(level); got != want {
			t.Errorf("GetXpToReachLevel(%d) = %d, want %d", level, got, want)
		}
	}
}

func TestGetXpToCompleteLevel(t *testing.T) {
	for level, want := range map[int]int{0: 7, 15: 37, 16: 42, 30: 112, 31: 121} {
		if got := GetXpToCompleteLevel(level); got != want {
			t.Errorf("GetXpToCompleteLevel(%d) = %d, want %d", level, got, want)
		}
	}
}

func TestGetLevelFromXpRoundTrips(t *testing.T) {
	for _, level := range []int{0, 1, 5, 16, 17, 31, 32, 50} {
		got, err := GetLevelFromXp(GetXpToReachLevel(level))
		if err != nil {
			t.Fatal(err)
		}
		if diff := got - float64(level); diff < -1e-9 || diff > 1e-9 {
			t.Errorf("GetLevelFromXp(GetXpToReachLevel(%d)) = %v", level, got)
		}
	}
	half, _ := GetLevelFromXp(GetXpToReachLevel(3) + GetXpToCompleteLevel(3)/2)
	if half <= 3 || half >= 4 {
		t.Errorf("level from mid-level XP = %v, want between 3 and 4", half)
	}
	if _, err := GetLevelFromXp(-1); err == nil {
		t.Error("GetLevelFromXp(-1) returned no error")
	}
}
