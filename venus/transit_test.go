package venus

import (
	"testing"
	"time"
)

func TestTransitWrappers(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	info := NextTransit(time.Date(2012, 1, 1, 0, 0, 0, 0, loc))
	if !info.Valid {
		t.Fatal("expected valid transit")
	}
	if info.Greatest.Location() != loc {
		t.Fatalf("timezone mismatch: got %v want %v", info.Greatest.Location(), loc)
	}
	if info.Greatest.Year() != 2012 || info.Greatest.Month() != time.June || info.Greatest.Day() != 6 {
		t.Fatalf("unexpected greatest time: %s", info.Greatest)
	}
	if !info.HasInternal || info.Duration <= 0 || info.InternalDuration <= 0 {
		t.Fatalf("unexpected durations: %+v", info)
	}
}
