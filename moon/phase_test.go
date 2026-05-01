package moon

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

func TestBrightLimbPositionAngleWrapperMatchesBasic(t *testing.T) {
	date := time.Date(2026, 4, 28, 9, 30, 45, 0, time.UTC)
	jd := observationTT(date)

	got := BrightLimbPositionAngle(date)
	want := basic.MoonBrightLimbPositionAngleN(jd, -1)
	if math.Float64bits(got) != math.Float64bits(want) {
		t.Fatalf("BrightLimbPositionAngle mismatch: got %.18f want %.18f", got, want)
	}
}

func TestTopocentricBrightLimbPositionAngleWrapperMatchesBasic(t *testing.T) {
	date := time.Date(2026, 4, 28, 9, 30, 45, 0, time.UTC)
	jd := observationTT(date)
	observerLon := 121.4737
	observerLat := 31.2304
	height := 4.0

	got := TopocentricBrightLimbPositionAngle(date, observerLon, observerLat, height)
	want := basic.MoonTopocentricBrightLimbPositionAngleN(jd, observerLon, observerLat, height, -1)
	if math.Float64bits(got) != math.Float64bits(want) {
		t.Fatalf("TopocentricBrightLimbPositionAngle mismatch: got %.18f want %.18f", got, want)
	}
}

func TestTopocentricBrightLimbPositionAnglePreservesInstantAcrossTimezones(t *testing.T) {
	utc := time.Date(2026, 4, 28, 9, 30, 45, 123000000, time.UTC)
	shanghai := utc.In(time.FixedZone("UTC+8", 8*3600))
	observerLon := 121.4737
	observerLat := 31.2304
	height := 4.0

	got := TopocentricBrightLimbPositionAngle(shanghai, observerLon, observerLat, height)
	want := TopocentricBrightLimbPositionAngle(utc, observerLon, observerLat, height)
	if math.Float64bits(got) != math.Float64bits(want) {
		t.Fatalf("timezone instant mismatch: got %.18f want %.18f", got, want)
	}
}
