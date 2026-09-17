package basic

import "testing"

// TestSolarEclipseSarosPathSeriesRemainFiniteWithinBudget exercises actual
// footprint and central-band generation at several points spanning roughly
// five centuries. The P0 state test catches bad ephemeris values; this test
// also walks the sampled path output and therefore covers the topology input
// used by GeoJSON and SVG consumers.
func TestSolarEclipseSarosPathSeriesRemainFiniteWithinBudget(t *testing.T) {
	const sarosDays = 6585.321314
	seed := JDECalc(2024, 4, 8)
	for _, familyIndex := range []int{-28, -21, -14, -7, 0, 7, 14, 21, 28} {
		familyIndex := familyIndex
		t.Run("saros-"+formatSignedRegressionIndex(familyIndex), func(t *testing.T) {
			result := SolarEclipsePartialFootprints(seed+float64(familyIndex)*sarosDays,
				SolarEclipsePartialFootprintOptions{
					StepDays:              20.0 / 1440.0,
					BoundaryPoints:        72,
					CentralShadowStepDays: 5.0 / 1440.0,
					DisableRiseSetCurves:  true,
				})
			if !result.Eclipse.HasPartial || len(result.Footprints) == 0 {
				t.Fatalf("missing partial path: type=%s footprints=%d", result.Eclipse.Type, len(result.Footprints))
			}
			assertSolarEclipseP2Footprints(t, "partial", result.Footprints)
			assertSolarEclipseP2Footprints(t, "central-shadow", result.CentralShadowFootprints)
			assertSolarEclipseP2Footprints(t, "central-band", result.CentralBandFootprints)
			if result.Eclipse.Centrality == SolarEclipseNonCentral && result.Eclipse.Type != SolarEclipsePartial &&
				len(result.CentralBandFootprints) == 0 {
				t.Fatal("non-central annular/total event has no central-band samples")
			}
			if result.Eclipse.HasCentral {
				path := SolarEclipseCentralPath(seed+float64(familyIndex)*sarosDays,
					SolarEclipsePathOptions{StepDays: 20.0 / 1440.0, TargetSpacingKM: 500})
				if len(path.CenterLine) < 2 || len(path.NorthernLimit) < 2 || len(path.SouthernLimit) < 2 {
					t.Fatalf("central path is incomplete: center=%d north=%d south=%d",
						len(path.CenterLine), len(path.NorthernLimit), len(path.SouthernLimit))
				}
				assertSolarEclipseP2PointSeries(t, "center-line", path.CenterLine)
				assertSolarEclipseP2PointSeries(t, "north-limit", path.NorthernLimit)
				assertSolarEclipseP2PointSeries(t, "south-limit", path.SouthernLimit)
			}
		})
	}
}

func TestSolarEclipseHighResolutionSamplingHonorsPointBudget(t *testing.T) {
	for _, test := range []struct {
		name       string
		shadowStep float64
	}{
		{name: "partial-only"},
		{name: "partial-and-central-shadow", shadowStep: 1.0 / 86400.0},
	} {
		t.Run(test.name, func(t *testing.T) {
			result := SolarEclipsePartialFootprints(JDECalc(2024, 4, 8), SolarEclipsePartialFootprintOptions{
				StepDays:              1.0 / 86400.0,
				BoundaryPoints:        solarEclipsePartialFootprintMaxBoundaryPoints,
				CentralShadowStepDays: test.shadowStep,
				DisableRiseSetCurves:  true,
			})
			if len(result.Footprints) == 0 {
				t.Fatal("high-resolution request returned no partial footprints")
			}
			if result.BoundaryPoints < solarEclipsePartialFootprintMinBoundaryPoints ||
				result.BoundaryPoints > solarEclipsePartialFootprintMaxBoundaryPoints {
				t.Fatalf("effective boundary points=%d outside [%d,%d]", result.BoundaryPoints,
					solarEclipsePartialFootprintMinBoundaryPoints, solarEclipsePartialFootprintMaxBoundaryPoints)
			}
			totalPoints := 0
			for _, series := range [][]SolarEclipsePartialFootprint{result.Footprints, result.CentralShadowFootprints, result.CentralBandFootprints} {
				for _, footprint := range series {
					for _, boundary := range footprint.Boundaries {
						totalPoints += len(boundary)
					}
				}
			}
			if totalPoints > solarEclipsePartialFootprintMaxPointCount {
				t.Fatalf("high-resolution output points=%d, want <=%d", totalPoints, solarEclipsePartialFootprintMaxPointCount)
			}
		})
	}
}

func assertSolarEclipseP2Footprints(t *testing.T, name string, footprints []SolarEclipsePartialFootprint) {
	t.Helper()
	for footprintIndex, footprint := range footprints {
		if !finite(footprint.JDE) || len(footprint.Boundaries) == 0 {
			t.Fatalf("%s footprint %d is incomplete: %+v", name, footprintIndex, footprint)
		}
		for boundaryIndex, boundary := range footprint.Boundaries {
			if len(boundary) < 2 {
				t.Fatalf("%s footprint %d boundary %d has %d points", name, footprintIndex, boundaryIndex, len(boundary))
			}
			for pointIndex, point := range boundary {
				if !finite(point.JDE) || !finite(point.Longitude) || !finite(point.Latitude) ||
					!finite(point.SunAltitude) || point.Longitude < -180 || point.Longitude > 180 ||
					point.Latitude < -90 || point.Latitude > 90 {
					t.Fatalf("%s footprint %d boundary %d point %d is invalid: %+v",
						name, footprintIndex, boundaryIndex, pointIndex, point)
				}
			}
		}
	}
}

func assertSolarEclipseP2PointSeries(t *testing.T, name string, points []SolarEclipsePathPoint) {
	t.Helper()
	for index, point := range points {
		if !finite(point.JDE) || !finite(point.Longitude) || !finite(point.Latitude) ||
			!finite(point.SunAltitude) || !finite(point.WidthKM) || point.Longitude < -180 ||
			point.Longitude > 180 || point.Latitude < -90 || point.Latitude > 90 {
			t.Fatalf("%s point %d is invalid: %+v", name, index, point)
		}
		if index > 0 && point.JDE <= points[index-1].JDE {
			t.Fatalf("%s times are not strictly increasing at %d", name, index)
		}
	}
}

func formatSignedRegressionIndex(value int) string {
	if value >= 0 {
		return "+" + formatRegressionIndexMagnitude(value)
	}
	return "-" + formatRegressionIndexMagnitude(-value)
}

func formatRegressionIndexMagnitude(value int) string {
	if value == 0 {
		return "0"
	}
	return string([]byte{'0' + byte(value/10), '0' + byte(value%10)})
}
