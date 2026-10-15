package sun

import (
	"testing"
	"time"
)

func TestRiseSetWithAtmosphere(t *testing.T) {
	date := time.Date(2025, 1, 1, 0, 0, 0, 0, time.FixedZone("CST", 8*3600))
	base, err := RiseTime(date, 116.4, 39.9, 0, true)
	if err != nil {
		t.Fatal(err)
	}
	same, err := RiseTimeWithAtmosphere(date, 116.4, 39.9, 0, 1010, 10)
	if diff := same.Sub(base); err != nil || diff > 10*time.Millisecond || diff < -10*time.Millisecond {
		t.Fatalf("standard weather %v vs %v (%v)", same, base, err)
	}
	cold, err := RiseTimeWithAtmosphere(date, 116.4, 39.9, 0, 1040, -30)
	if err != nil || !cold.Before(base) {
		t.Fatalf("cold dense air should advance sunrise: %v vs %v (%v)", cold, base, err)
	}
	set, err := SetTimeWithAtmosphere(date, 116.4, 39.9, 0, 1040, -30)
	if err != nil || set.Location() != date.Location() {
		t.Fatalf("sunset: %v %v", set, err)
	}
}
