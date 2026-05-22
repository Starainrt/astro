package basic

import (
	"math"
	"testing"
	"time"
)

func mercuryTTJDJST(year int, month time.Month, day, hour, minute, second int) float64 {
	loc := time.FixedZone("JST", 9*3600)
	return TD2UT(Date2JDE(time.Date(year, month, day, hour, minute, second, 0, loc).UTC()), true)
}

func TestMercuryTypedStationRegression1929(t *testing.T) {
	loc := time.FixedZone("JST", 9*3600)
	const tolerance = 30.0 / 86400.0

	query := mercuryTTJDJST(1929, time.September, 20, 0, 0, 0)
	wantP2R := mercuryTTJDJST(1929, time.September, 26, 1, 58, 0)
	wantR2P := mercuryTTJDJST(1929, time.October, 16, 23, 32, 33)

	nextP2R := NextMercuryProgradeToRetrograde(query)
	nextR2P := NextMercuryRetrogradeToPrograde(query)
	if math.Abs(nextP2R-wantP2R) > tolerance {
		t.Fatalf("next P2R mismatch: got %s want %s", JDE2DateByZone(nextP2R, loc, false), JDE2DateByZone(wantP2R, loc, false))
	}
	if math.Abs(nextR2P-wantR2P) > tolerance {
		t.Fatalf("next R2P mismatch: got %s want %s", JDE2DateByZone(nextR2P, loc, false), JDE2DateByZone(wantR2P, loc, false))
	}

	query = mercuryTTJDJST(1929, time.October, 20, 0, 0, 0)
	lastP2R := LastMercuryProgradeToRetrograde(query)
	lastR2P := LastMercuryRetrogradeToPrograde(query)
	if math.Abs(lastP2R-wantP2R) > tolerance {
		t.Fatalf("last P2R mismatch: got %s want %s", JDE2DateByZone(lastP2R, loc, false), JDE2DateByZone(wantP2R, loc, false))
	}
	if math.Abs(lastR2P-wantR2P) > tolerance {
		t.Fatalf("last R2P mismatch: got %s want %s", JDE2DateByZone(lastR2P, loc, false), JDE2DateByZone(wantR2P, loc, false))
	}
}
