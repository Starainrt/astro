package satellite

import (
	"errors"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/geodata"
)

func TestCoverageNearZenithTwilight(t *testing.T) {
	at := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	jd := basic.Date2JD(at)
	gast := basic.ApparentSiderealTime(basic.UTC2UT1(jd)) * 15 * degreesToRadians
	sun := apparentSunPosition(basic.UTC2TT(jd), gast, nil)
	lon, lat, _ := geographicPosition(sun)
	day := GeoPoint{Longitude: lon, Latitude: lat}
	for _, threshold := range []float64{89.999, -89.999, 89.999999, -89.999999} {
		coverage, err := CoverageAt(coverageGEOFixture(t), at, CoverageOptions{TwilightSunAltitude: &threshold})
		if err != nil || len(coverage.Night) < 4 || coverage.NightIsGlobal {
			t.Fatalf("threshold %g: night boundary has %d vertices: %v", threshold, len(coverage.Night), err)
		}
		if geodata.ClockwiseRegionIsComplement(coverageGeoRing(coverage.Night)) != (threshold > 0) {
			t.Fatalf("threshold %g: night ring selects the wrong side", threshold)
		}
		for _, point := range coverage.Night {
			if delta := math.Abs(coverageAltitude(sun, point) - threshold); delta > 2e-8 {
				t.Fatalf("threshold %g: residual %g at %v", threshold, delta, point)
			}
		}
		if threshold == 89.999 {
			near := GeoPoint{Longitude: lon, Latitude: lat + .002}
			if !coverageReferenceContains(coverage.Night, day) || coverageReferenceContains(coverage.Night, near) || coverageAltitude(sun, day) <= threshold || coverageAltitude(sun, near) >= threshold {
				t.Fatal("near-zenith daylight cap has wrong point membership")
			}
		}
	}
}

func TestCoverageTwilightEndpoints(t *testing.T) {
	at := time.Date(2026, 11, 14, 14, 0, 0, 0, time.UTC)
	elements := coverageGEOFixture(t)
	for _, threshold := range []float64{-90, 90} {
		sweep, err := CoverageBetween(elements, at, at.Add(10*time.Second), 10*time.Second, CoverageOptions{TwilightSunAltitude: &threshold})
		if err != nil {
			t.Fatal(err)
		}
		for _, coverage := range sweep.Instants {
			if len(coverage.Night) != 0 || coverage.NightIsGlobal != (threshold > 0) || !coverage.SunCenterVisible {
				t.Fatalf("threshold %g: unexpected night boundary", threshold)
			}
			if threshold > 0 && (len(coverage.Visible) != 1 || !reflect.DeepEqual(coverage.Visible[0], coverage.Footprint)) {
				t.Fatal("global night must expose the whole sunlit footprint")
			}
		}
		if threshold < 0 && len(sweep.VisibleSwept) > 0 || threshold > 0 && !reflect.DeepEqual(sweep.VisibleSwept, sweep.Swept) {
			t.Fatalf("threshold %g: unexpected visible sweep", threshold)
		}
	}
	for _, threshold := range []float64{-90.0001, 90.0001, math.NaN(), math.Inf(1)} {
		if _, err := CoverageAt(elements, at, CoverageOptions{TwilightSunAltitude: &threshold}); !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("threshold %g: %v", threshold, err)
		}
	}
}
