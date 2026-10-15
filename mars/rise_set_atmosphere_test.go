package mars

import (
	"math"
	"testing"
	"time"
)

func TestRiseTimeWithAtmosphere(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	date := time.Date(2025, 1, 1, 0, 0, 0, 0, loc)
	plain, err := RiseTime(date, 116.4, 39.9, 0, true)
	if err != nil {
		t.Fatal(err)
	}
	std, err := RiseTimeWithAtmosphere(date, 116.4, 39.9, 0, 1010, 10)
	if err != nil {
		t.Fatal(err)
	}
	dense, err := RiseTimeWithAtmosphere(date, 116.4, 39.9, 0, 1040, -30)
	if err != nil {
		t.Fatal(err)
	}
	if d := math.Abs(std.Sub(plain).Seconds()); d > 5 {
		t.Errorf("standard weather differs by %.2f s", d)
	}
	if !dense.Before(std) {
		t.Errorf("dense air must rise earlier: %v vs %v", dense, std)
	}
	if _, err := RiseTimeWithAtmosphere(date, 116.4, 39.9, 0, -1, 10); err == nil {
		t.Error("invalid pressure must fail")
	}
}
