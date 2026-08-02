package block

import "testing"

func TestSoulCampfireGetDropsForCompatibleToolReturnsSoulSoil(t *testing.T) {
	withFakeItemBlockFactory(t)
	w := &fakeWorld{}
	s := newTestSoulCampfire(w)

	drops := s.GetDropsForCompatibleTool(fakeItem{})
	if len(drops) != 1 {
		t.Fatalf("expected 1 drop, got %d", len(drops))
	}
	wrapped, ok := drops[0].(*fakeItemBlock)
	if !ok || wrapped.wrapped.GetTypeId() != SOUL_SOIL {
		t.Errorf("expected a Soul Soil drop, got %#v", drops[0])
	}
}
