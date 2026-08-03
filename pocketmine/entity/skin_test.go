package entity

import "testing"

func validSkinData() []byte { return make([]byte, 64*64*4) }

func TestNewSkinAcceptsAllThreeValidSkinDataSizes(t *testing.T) {
	for _, size := range []int{64 * 32 * 4, 64 * 64 * 4, 128 * 128 * 4} {
		if _, err := NewSkin("id", make([]byte, size), nil, "", nil); err != nil {
			t.Errorf("NewSkin with skin data size %d returned an error: %v", size, err)
		}
	}
}

func TestNewSkinRejectsAnInvalidSkinDataSize(t *testing.T) {
	if _, err := NewSkin("id", make([]byte, 123), nil, "", nil); err == nil {
		t.Error("NewSkin with an invalid skin data size = nil error, want an error")
	}
}

func TestNewSkinRejectsAnEmptySkinID(t *testing.T) {
	if _, err := NewSkin("", validSkinData(), nil, "", nil); err == nil {
		t.Error("NewSkin with an empty skin ID = nil error, want an error")
	}
}

func TestNewSkinAcceptsEmptyCapeDataButRejectsWrongSizedCapeData(t *testing.T) {
	if _, err := NewSkin("id", validSkinData(), nil, "", nil); err != nil {
		t.Errorf("NewSkin with no cape data returned an error: %v", err)
	}
	if _, err := NewSkin("id", validSkinData(), make([]byte, 100), "", nil); err == nil {
		t.Error("NewSkin with wrong-sized cape data = nil error, want an error")
	}
	if _, err := NewSkin("id", validSkinData(), make([]byte, 8192), "", nil); err != nil {
		t.Errorf("NewSkin with exactly 8192 bytes of cape data returned an error: %v", err)
	}
}

func TestNewSkinGettersReturnWhatWasPassedIn(t *testing.T) {
	skinData := validSkinData()
	capeData := make([]byte, 8192)
	geometryData := []byte(`{"foo":"bar"}`)

	s, err := NewSkin("my-skin-id", skinData, capeData, "geometry.humanoid.custom", geometryData)
	if err != nil {
		t.Fatal(err)
	}
	if s.GetSkinID() != "my-skin-id" {
		t.Errorf("GetSkinID() = %q, want %q", s.GetSkinID(), "my-skin-id")
	}
	if s.GetGeometryName() != "geometry.humanoid.custom" {
		t.Errorf("GetGeometryName() = %q, want %q", s.GetGeometryName(), "geometry.humanoid.custom")
	}
	if len(s.GetSkinData()) != len(skinData) {
		t.Errorf("len(GetSkinData()) = %d, want %d", len(s.GetSkinData()), len(skinData))
	}
	if len(s.GetCapeData()) != len(capeData) {
		t.Errorf("len(GetCapeData()) = %d, want %d", len(s.GetCapeData()), len(capeData))
	}
	if string(s.GetGeometryData()) != string(geometryData) {
		t.Errorf("GetGeometryData() = %q, want %q", s.GetGeometryData(), geometryData)
	}
}
