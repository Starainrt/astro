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

func TestLocalTransitWrapper(t *testing.T) {
	info := NextTransit(time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC))
	local := LocalTransit(info, 151.2093, -33.8688, 0)
	if !local.Valid || !local.HasInternal || !local.Visible {
		t.Fatalf("unexpected local transit: %+v", local)
	}
	// 悉尼 2012 金星凌日：一触约 22:16 UTC，太阳全程在地平上。
	want := time.Date(2012, 6, 5, 22, 16, 0, 0, time.UTC)
	if diff := local.Start.Time.Sub(want); diff < -time.Minute || diff > time.Minute {
		t.Fatalf("C1 %s", local.Start.Time)
	}
	if !local.Sunrise.IsZero() || !local.Sunset.IsZero() || local.End.SunAltitude <= 0 {
		t.Fatalf("unexpected sun events: %+v", local)
	}
}
