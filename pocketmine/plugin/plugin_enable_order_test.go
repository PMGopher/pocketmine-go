package plugin

import "testing"

func TestEnableOrderFromStringRecognizesAliases(t *testing.T) {
	tests := []struct {
		in   string
		want EnableOrder
	}{
		{"startup", EnableOrderStartup},
		{"STARTUP", EnableOrderStartup},
		{"postworld", EnableOrderPostworld},
		{"PostWorld", EnableOrderPostworld},
	}
	for _, tt := range tests {
		got, ok := EnableOrderFromString(tt.in)
		if !ok || got != tt.want {
			t.Errorf("EnableOrderFromString(%q) = %v, %v; want %v, true", tt.in, got, ok, tt.want)
		}
	}
}

func TestEnableOrderFromStringRejectsUnknown(t *testing.T) {
	if _, ok := EnableOrderFromString("whenever"); ok {
		t.Error("EnableOrderFromString(\"whenever\") ok = true, want false")
	}
}
