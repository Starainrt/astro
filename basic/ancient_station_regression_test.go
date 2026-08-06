package basic

import (
	"math"
	"testing"
	"time"
)

const ancientStationTolerance = 10.0 / 1440.0

func ancientStationTT(year int, month time.Month, day int) float64 {
	return TD2UT(Date2JDE(time.Date(year, month, day, 0, 0, 0, 0, time.UTC)), true)
}

func ancientStationUT(year int, month time.Month, day, hour, minute int) float64 {
	return Date2JDE(time.Date(year, month, day, hour, minute, 0, 0, time.UTC))
}

func assertAncientNextStation(t *testing.T, queryTT, gotUT, wantUT float64) {
	t.Helper()
	if !eventUTQueryAfterOrEqual(gotUT, queryTT) {
		t.Fatalf("station is before query: query TT %.9f, got UT %.9f", queryTT, gotUT)
	}
	if diff := math.Abs(gotUT - wantUT); diff > ancientStationTolerance {
		t.Fatalf("station differs by %.3f minutes: got %s, want %s", diff*1440,
			JDE2DateByZone(gotUT, time.UTC, false), JDE2DateByZone(wantUT, time.UTC, false))
	}
}

func TestMarsAncientNextStationsStayOnTheirOppositionSides(t *testing.T) {
	queryTT := ancientStationTT(-210, time.April, 1)
	p2r := NextMarsProgradeToRetrograde(queryTT)
	r2p := NextMarsRetrogradeToPrograde(queryTT)

	assertAncientNextStation(t, queryTT, p2r, ancientStationUT(-209, time.March, 25, 20, 20))
	assertAncientNextStation(t, queryTT, r2p, ancientStationUT(-209, time.June, 6, 11, 4))
	if sameEventJD(p2r, r2p) {
		t.Fatalf("typed Mars stations collapsed to one event: %.9f", p2r)
	}
}

func TestMercuryAncientNextStationsDoNotSkipOrLoop(t *testing.T) {
	janQueryTT := ancientStationTT(-210, time.January, 1)
	assertAncientNextStation(t, janQueryTT, NextMercuryProgradeToRetrograde(janQueryTT), ancientStationUT(-210, time.February, 6, 21, 29))
	assertAncientNextStation(t, janQueryTT, NextMercuryRetrogradeToPrograde(janQueryTT), ancientStationUT(-210, time.March, 1, 14, 20))

	marQueryTT := ancientStationTT(-210, time.March, 1)
	assertAncientNextStation(t, marQueryTT, NextMercuryProgradeToRetrograde(marQueryTT), ancientStationUT(-210, time.June, 11, 18, 50))
	assertAncientNextStation(t, marQueryTT, NextMercuryRetrogradeToPrograde(marQueryTT), ancientStationUT(-210, time.March, 1, 14, 20))
}
