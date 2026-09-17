package basic

import (
	"fmt"
	"math"
	"testing"
)

func TestSolarEclipseTotalEnvelopeHorizonEndpoints(t *testing.T) {
	for _, date := range [][3]int{
		{2003, 11, 23}, {2021, 12, 4}, {2039, 12, 15},
		{2008, 8, 1}, {2024, 4, 8}, {2026, 8, 12},
	} {
		t.Run(fmt.Sprintf("%04d-%02d-%02d", date[0], date[1], date[2]), func(t *testing.T) {
			seed := JDECalc(date[0], date[1], float64(date[2]))
			result := SolarEclipsePartialFootprints(seed, SolarEclipsePartialFootprintOptions{
				StepDays: 2.0 / 1440, BoundaryPoints: 96, MagnitudeValues: []float64{1},
			})
			if result.Eclipse.Type != SolarEclipseTotal || len(result.CentralBandSegments) != 1 ||
				len(result.CentralBandHorizonClosures) != 2 {
				t.Fatalf("type=%s rings=%d closures=%d, want total with a continuous closed envelope",
					result.Eclipse.Type, len(result.CentralBandSegments), len(result.CentralBandHorizonClosures))
			}
			if len(result.MagnitudeContours) != 1 || len(result.MagnitudeContours[0].Segments) != 2 {
				t.Fatal("missing totality branches")
			}
			ring := result.CentralBandSegments[0]
			if len(ring) < 4 || ring[0] != ring[len(ring)-1] {
				t.Fatal("totality envelope is not closed")
			}
			solver := newSolarEclipseSolver(CalcMoonSHByJDE(seed, 0), SolarEclipseModelNASABulletinSplitK)
			for i, point := range ring {
				if i > 0 && solarEclipsePathDistanceKM(ring[i-1], point) > solarEclipseTotalEnvelopeTargetSpacingKM+1e-6 {
					t.Fatalf("edge %d exceeds spacing limit", i)
				}
				if solarEclipseCentralEnvelopePointOnHorizonClosure(point, result.CentralBandHorizonClosures) {
					continue
				}
				residual, ok := solver.magnitudeHorizonResidual(point.JDE, point.Longitude, point.Latitude, 1)
				if !ok || math.Abs(residual[0]) > 1e-7 || math.Abs(residual[1]) > 1e-8 || point.SunAltitude < -1e-5 {
					t.Fatalf("point %d is not on visible totality envelope: %v altitude=%g", i, residual, point.SunAltitude)
				}
			}
			for _, branch := range result.MagnitudeContours[0].Segments {
				for _, endpoint := range []SolarEclipsePathPoint{branch[0], branch[len(branch)-1]} {
					if !solarEclipseCentralEnvelopePointOnHorizonClosure(endpoint, result.CentralBandHorizonClosures) {
						t.Fatal("totality endpoint is not shared with a horizon closure")
					}
				}
			}
		})
	}
}

func BenchmarkSolarEclipsePolarTotalBand(b *testing.B) {
	seed := JDECalc(2003, 11, 23)
	for i := 0; i < b.N; i++ {
		SolarEclipsePartialFootprints(seed, SolarEclipsePartialFootprintOptions{
			StepDays: 2.0 / 1440, BoundaryPoints: 96, DisableRiseSetCurves: true,
		})
	}
}
