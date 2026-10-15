package mercury

import (
	"testing"
	"time"
)

func TestTransitWrappers(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	info := NextTransit(time.Date(2019, 1, 1, 0, 0, 0, 0, loc))
	if !info.Valid {
		t.Fatal("expected valid transit")
	}
	if info.Greatest.Location() != loc {
		t.Fatalf("timezone mismatch: got %v want %v", info.Greatest.Location(), loc)
	}
	if info.Greatest.Year() != 2019 || info.Greatest.Month() != time.November || info.Greatest.Day() != 11 {
		t.Fatalf("unexpected greatest time: %s", info.Greatest)
	}
	if !info.HasInternal || info.Duration <= 0 || info.InternalDuration <= 0 {
		t.Fatalf("unexpected durations: %+v", info)
	}
}

func TestLocalTransitWrapper(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	info := NextTransit(time.Date(2019, 1, 1, 0, 0, 0, 0, loc))
	local := LocalTransit(info, -77.0, 38.9, 0)
	if !local.Valid || !local.HasInternal || !local.Visible {
		t.Fatalf("unexpected local transit: %+v", local)
	}
	if local.Greatest.Time.Location() != loc {
		t.Fatalf("timezone mismatch: %v", local.Greatest.Time.Location())
	}
	// 站心接触与地心接触只差视差引起的几十秒。
	if shift := local.Start.Time.Sub(info.Start); shift < -2*time.Minute || shift > 2*time.Minute {
		t.Fatalf("C1 shift %v", shift)
	}
	if LocalTransit(TransitInfo{}, 0, 0, 0).Valid {
		t.Fatal("invalid transit must give invalid local result")
	}
}
