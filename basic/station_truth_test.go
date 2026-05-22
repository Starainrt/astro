package basic

import (
	"testing"
	"time"
)

type stationEvent struct {
	when time.Time
	kind string
}

type stationTruthCase struct {
	name    string
	events  []stationEvent
	lastR2P func(float64) float64
	nextR2P func(float64) float64
	lastP2R func(float64) float64
	nextP2R func(float64) float64
}

func mustJST(value string) time.Time {
	loc := time.FixedZone("JST", 9*3600)
	t, err := time.ParseInLocation("2006-01-02 15:04", value, loc)
	if err != nil {
		panic(err)
	}
	return t
}

func toUTJD(t time.Time) float64 {
	return TD2UT(Date2JDE(t.UTC()), true)
}

func TestStationTruthAgainstNAOJ(t *testing.T) {
	cases := []stationTruthCase{
		{
			name: "Mars",
			events: []stationEvent{
				{when: mustJST("2024-12-08 05:59"), kind: "P2R"},
				{when: mustJST("2025-01-16 11:39"), kind: "OPP"},
				{when: mustJST("2025-02-24 18:35"), kind: "R2P"},
				{when: mustJST("2027-01-12 01:10"), kind: "P2R"},
				{when: mustJST("2027-02-20 00:51"), kind: "OPP"},
				{when: mustJST("2027-04-03 02:33"), kind: "R2P"},
			},
			lastR2P: LastMarsRetrogradeToPrograde,
			nextR2P: NextMarsRetrogradeToPrograde,
			lastP2R: LastMarsProgradeToRetrograde,
			nextP2R: NextMarsProgradeToRetrograde,
		},
		{
			name: "Jupiter",
			events: []stationEvent{
				{when: mustJST("2024-10-09 16:13"), kind: "P2R"},
				{when: mustJST("2024-12-08 05:58"), kind: "OPP"},
				{when: mustJST("2025-02-04 22:07"), kind: "R2P"},
				{when: mustJST("2025-11-12 04:54"), kind: "P2R"},
				{when: mustJST("2026-01-10 17:42"), kind: "OPP"},
				{when: mustJST("2026-03-11 11:44"), kind: "R2P"},
				{when: mustJST("2026-12-13 21:03"), kind: "P2R"},
				{when: mustJST("2027-02-11 09:29"), kind: "OPP"},
				{when: mustJST("2027-04-13 15:17"), kind: "R2P"},
			},
			lastR2P: LastJupiterRetrogradeToPrograde,
			nextR2P: NextJupiterRetrogradeToPrograde,
			lastP2R: LastJupiterProgradeToRetrograde,
			nextP2R: NextJupiterProgradeToRetrograde,
		},
		{
			name: "Saturn",
			events: []stationEvent{
				{when: mustJST("2024-07-01 06:15"), kind: "P2R"},
				{when: mustJST("2024-09-08 13:35"), kind: "OPP"},
				{when: mustJST("2024-11-16 14:57"), kind: "R2P"},
				{when: mustJST("2025-07-14 16:57"), kind: "P2R"},
				{when: mustJST("2025-09-21 14:46"), kind: "OPP"},
				{when: mustJST("2025-11-29 09:35"), kind: "R2P"},
				{when: mustJST("2026-07-28 08:09"), kind: "P2R"},
				{when: mustJST("2026-10-04 21:29"), kind: "OPP"},
				{when: mustJST("2026-12-12 08:21"), kind: "R2P"},
				{when: mustJST("2027-08-11 02:53"), kind: "P2R"},
				{when: mustJST("2027-10-18 09:36"), kind: "OPP"},
				{when: mustJST("2027-12-25 12:05"), kind: "R2P"},
			},
			lastR2P: LastSaturnRetrogradeToPrograde,
			nextR2P: NextSaturnRetrogradeToPrograde,
			lastP2R: LastSaturnProgradeToRetrograde,
			nextP2R: NextSaturnProgradeToRetrograde,
		},
		{
			name: "Uranus",
			events: []stationEvent{
				{when: mustJST("2024-01-27 19:50"), kind: "R2P"},
				{when: mustJST("2024-09-02 00:44"), kind: "P2R"},
				{when: mustJST("2024-11-17 11:45"), kind: "OPP"},
				{when: mustJST("2025-01-31 04:04"), kind: "R2P"},
				{when: mustJST("2025-09-06 13:55"), kind: "P2R"},
				{when: mustJST("2025-11-21 21:25"), kind: "OPP"},
				{when: mustJST("2026-02-04 13:37"), kind: "R2P"},
				{when: mustJST("2026-09-11 03:19"), kind: "P2R"},
				{when: mustJST("2026-11-26 07:41"), kind: "OPP"},
				{when: mustJST("2027-02-08 23:03"), kind: "R2P"},
				{when: mustJST("2027-09-15 17:50"), kind: "P2R"},
				{when: mustJST("2027-11-30 18:22"), kind: "OPP"},
			},
			lastR2P: LastUranusRetrogradeToPrograde,
			nextR2P: NextUranusRetrogradeToPrograde,
			lastP2R: LastUranusProgradeToRetrograde,
			nextP2R: NextUranusProgradeToRetrograde,
		},
		{
			name: "Neptune",
			events: []stationEvent{
				{when: mustJST("2024-07-03 12:08"), kind: "P2R"},
				{when: mustJST("2024-09-21 09:17"), kind: "OPP"},
				{when: mustJST("2024-12-08 20:05"), kind: "R2P"},
				{when: mustJST("2025-07-05 23:30"), kind: "P2R"},
				{when: mustJST("2025-09-23 21:54"), kind: "OPP"},
				{when: mustJST("2025-12-11 09:21"), kind: "R2P"},
				{when: mustJST("2026-07-08 13:02"), kind: "P2R"},
				{when: mustJST("2026-09-26 10:36"), kind: "OPP"},
				{when: mustJST("2026-12-13 19:47"), kind: "R2P"},
				{when: mustJST("2027-07-11 01:06"), kind: "P2R"},
				{when: mustJST("2027-09-28 23:19"), kind: "OPP"},
				{when: mustJST("2027-12-16 07:16"), kind: "R2P"},
			},
			lastR2P: LastNeptuneRetrogradeToPrograde,
			nextR2P: NextNeptuneRetrogradeToPrograde,
			lastP2R: LastNeptuneProgradeToRetrograde,
			nextP2R: NextNeptuneProgradeToRetrograde,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			for i, event := range tc.events {
				switch event.kind {
				case "P2R":
					before := event.when.Add(-24 * time.Hour)
					after := event.when.Add(24 * time.Hour)
					nextP2R := JDE2DateByZone(tc.nextP2R(toUTJD(before)), event.when.Location(), false)
					lastP2R := JDE2DateByZone(tc.lastP2R(toUTJD(after)), event.when.Location(), false)
					if !sameMinute(nextP2R, event.when) {
						t.Fatalf("%s next P2R mismatch: got %s want %s", tc.name, nextP2R, event.when)
					}
					if !sameMinute(lastP2R, event.when) {
						t.Fatalf("%s last P2R mismatch: got %s want %s", tc.name, lastP2R, event.when)
					}
				case "R2P":
					before := event.when.Add(-24 * time.Hour)
					after := event.when.Add(24 * time.Hour)
					nextR2P := JDE2DateByZone(tc.nextR2P(toUTJD(before)), event.when.Location(), false)
					lastR2P := JDE2DateByZone(tc.lastR2P(toUTJD(after)), event.when.Location(), false)
					if !sameMinute(nextR2P, event.when) {
						t.Fatalf("%s next R2P mismatch: got %s want %s", tc.name, nextR2P, event.when)
					}
					if !sameMinute(lastR2P, event.when) {
						t.Fatalf("%s last R2P mismatch: got %s want %s", tc.name, lastR2P, event.when)
					}
				case "OPP":
					prev := nearestOfKindBefore(tc.events, i, "P2R")
					next := nearestOfKindAfter(tc.events, i, "R2P")
					if prev.IsZero() || next.IsZero() {
						continue
					}
					query := event.when
					lastP2R := JDE2DateByZone(tc.lastP2R(toUTJD(query)), query.Location(), false)
					nextR2P := JDE2DateByZone(tc.nextR2P(toUTJD(query)), query.Location(), false)
					if !sameMinute(lastP2R, prev) {
						t.Fatalf("%s opposition last P2R mismatch: got %s want %s", tc.name, lastP2R, prev)
					}
					if !sameMinute(nextR2P, next) {
						t.Fatalf("%s opposition next R2P mismatch: got %s want %s", tc.name, nextR2P, next)
					}
				}
			}
		})
	}
}

func nearestOfKindBefore(events []stationEvent, idx int, kind string) time.Time {
	for i := idx - 1; i >= 0; i-- {
		if events[i].kind == kind {
			return events[i].when
		}
	}
	return time.Time{}
}

func nearestOfKindAfter(events []stationEvent, idx int, kind string) time.Time {
	for i := idx + 1; i < len(events); i++ {
		if events[i].kind == kind {
			return events[i].when
		}
	}
	return time.Time{}
}

func sameMinute(got, want time.Time) bool {
	diff := got.Sub(want)
	if diff < 0 {
		diff = -diff
	}
	return diff <= 2*time.Minute
}
