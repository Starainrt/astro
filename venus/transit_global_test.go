package venus

import (
	"testing"
	"time"
)

func TestGlobalTransitWrapper(tester *testing.T) {
	zone := time.FixedZone("UTC+8", 8*3600)
	info := NextTransit(time.Date(2854, 1, 1, 0, 0, 0, 0, zone))
	if !info.Valid || info.Greatest.Year() != 2854 || info.HasExternal || !info.Start.IsZero() || !info.End.IsZero() || info.Duration != 0 || info.Series.Number != 4 {
		tester.Fatalf("invalid global transit: %+v", info)
	}
	if info.Greatest.Location() != zone || info.WindowStart.Location() != zone || info.WindowEnd.Location() != zone {
		tester.Fatal("global transit lost the caller's timezone")
	}
	local := LocalTransit(info, -135, -60, 0)
	if !local.Valid || !local.Visible || local.HasInternal || local.Greatest.Time.Location() != zone {
		tester.Fatalf("invalid local partial transit: %+v", local)
	}
	for _, result := range []TransitInfo{ClosestTransit(info.Greatest), LastTransit(info.Greatest.Add(time.Minute)), NextTransit(info.Greatest.Add(-time.Minute))} {
		difference := result.Greatest.Sub(info.Greatest)
		if !result.Valid || difference > time.Second || difference < -time.Second {
			tester.Fatal("global event lost at a query boundary")
		}
	}
}

func TestLocalTransitRequiresContactsOrWindow(tester *testing.T) {
	info := NextTransit(time.Date(2854, 1, 1, 0, 0, 0, 0, time.UTC))
	info.WindowStart, info.WindowEnd = time.Time{}, time.Time{}
	if local := LocalTransit(info, -135, -60, 0); local.Valid {
		tester.Fatal("partial transit without a window must be rejected")
	}
}
