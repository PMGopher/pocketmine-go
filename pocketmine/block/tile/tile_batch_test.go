package tile

import (
	"testing"

	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

func TestComparatorSaveDataRoundTrip(t *testing.T) {
	w := &fakeWorld{}
	c := NewComparator(w, math.Vector3{})
	c.SetSignalStrength(9)

	saved := c.SaveNBT()
	decoded := NewComparator(w, math.Vector3{})
	if err := decoded.ReadSaveData(saved); err != nil {
		t.Fatalf("ReadSaveData: %v", err)
	}
	if decoded.GetSignalStrength() != 9 {
		t.Errorf("GetSignalStrength() = %d, want 9", decoded.GetSignalStrength())
	}
}

func TestBeaconSaveDataRoundTrip(t *testing.T) {
	w := &fakeWorld{}
	b := NewBeacon(w, math.Vector3{})
	b.SetPrimaryEffect(3)
	b.SetSecondaryEffect(7)

	saved := b.SaveNBT()
	decoded := NewBeacon(w, math.Vector3{})
	if err := decoded.ReadSaveData(saved); err != nil {
		t.Fatalf("ReadSaveData: %v", err)
	}
	if decoded.GetPrimaryEffect() != 3 || decoded.GetSecondaryEffect() != 7 {
		t.Errorf("effects = (%d, %d), want (3, 7)", decoded.GetPrimaryEffect(), decoded.GetSecondaryEffect())
	}
}

func TestMobHeadSaveDataRoundTrip(t *testing.T) {
	w := &fakeWorld{}
	m := NewMobHead(w, math.Vector3{})
	m.SetMobHeadType(blockutils.MobHeadTypeDragon)
	m.SetRotation(5)

	saved := m.SaveNBT()
	decoded := NewMobHead(w, math.Vector3{})
	if err := decoded.ReadSaveData(saved); err != nil {
		t.Fatalf("ReadSaveData: %v", err)
	}
	if decoded.GetMobHeadType() != blockutils.MobHeadTypeDragon {
		t.Errorf("GetMobHeadType() = %v, want Dragon", decoded.GetMobHeadType())
	}
	if decoded.GetRotation() != 5 {
		t.Errorf("GetRotation() = %d, want 5", decoded.GetRotation())
	}
}

func TestMobHeadReadSaveDataRejectsInvalidSkullType(t *testing.T) {
	w := &fakeWorld{}
	m := NewMobHead(w, math.Vector3{})

	tag := nbt.NewCompoundTag()
	tag.SetByte(mobHeadTagSkullType, nbt.ByteTag(99))

	if err := m.ReadSaveData(tag); err == nil {
		t.Error("expected an out-of-range skull type to return an error")
	}
}

func TestMobHeadReadSaveDataFallsBackToLegacyRotTag(t *testing.T) {
	w := &fakeWorld{}
	m := NewMobHead(w, math.Vector3{})

	tag := nbt.NewCompoundTag()
	tag.SetByte(mobHeadTagRot, nbt.ByteTag(11))

	if err := m.ReadSaveData(tag); err != nil {
		t.Fatalf("ReadSaveData: %v", err)
	}
	if m.GetRotation() != 11 {
		t.Errorf("GetRotation() = %d, want 11 (from legacy Rot tag)", m.GetRotation())
	}
}

func TestEnchantTableNameDefaultsAndCustom(t *testing.T) {
	w := &fakeWorld{}
	e := NewEnchantTable(w, math.Vector3{})

	if got := e.GetName(); got != "Enchanting Table" {
		t.Errorf("GetName() = %q, want %q", got, "Enchanting Table")
	}
	e.SetName("Magic Desk")
	if got := e.GetName(); got != "Magic Desk" {
		t.Errorf("GetName() = %q, want %q", got, "Magic Desk")
	}

	saved := e.SaveNBT()
	decoded := NewEnchantTable(w, math.Vector3{})
	if err := decoded.ReadSaveData(saved); err != nil {
		t.Fatalf("ReadSaveData: %v", err)
	}
	if got := decoded.GetName(); got != "Magic Desk" {
		t.Errorf("decoded GetName() = %q, want %q", got, "Magic Desk")
	}
}

func TestBellSaveDataRoundTrip(t *testing.T) {
	w := &fakeWorld{}
	b := NewBell(w, math.Vector3{})
	b.SetRinging(true)
	b.SetFacing(int(math.East))
	b.SetTicks(4)

	saved := b.SaveNBT()
	decoded := NewBell(w, math.Vector3{})
	if err := decoded.ReadSaveData(saved); err != nil {
		t.Fatalf("ReadSaveData: %v", err)
	}
	if !decoded.IsRinging() || decoded.GetFacing() != int(math.East) || decoded.GetTicks() != 4 {
		t.Errorf("decoded = (ringing=%v, facing=%d, ticks=%d), want (true, %d, 4)",
			decoded.IsRinging(), decoded.GetFacing(), decoded.GetTicks(), int(math.East))
	}
}
