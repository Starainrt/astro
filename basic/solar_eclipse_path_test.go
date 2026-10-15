package basic

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/internal/ephemtest"
)

func TestSolarEclipseCentralPathMatchesGlobalGreatest(t *testing.T) {
	seedJDE := JDCalc(2024, 4, 8)
	global := SolarEclipse(seedJDE)
	path := SolarEclipseCentralPath(seedJDE, SolarEclipsePathOptions{StepDays: 5.0 / 1440.0})

	if path.Eclipse.Type != global.Type {
		t.Fatalf("type mismatch: got %s want %s", path.Eclipse.Type, global.Type)
	}
	if !path.Eclipse.HasCentral {
		t.Fatalf("expected central eclipse path")
	}
	if len(path.CenterLine) == 0 {
		t.Fatalf("expected center line points")
	}
	if len(path.NorthernLimit) == 0 || len(path.SouthernLimit) == 0 {
		t.Fatalf("expected central path limits: north=%d south=%d", len(path.NorthernLimit), len(path.SouthernLimit))
	}

	assertSolarEclipseJDEClose(t, "Greatest.JDE", path.Greatest.JDE, global.GreatestEclipse, 1e-8)
	assertSolarEclipseFloatClose(t, "Greatest.Longitude", path.Greatest.Longitude, global.GreatestLongitude, 1e-9)
	assertSolarEclipseFloatClose(t, "Greatest.Latitude", path.Greatest.Latitude, global.GreatestLatitude, 1e-9)
	assertSolarEclipseFloatClose(t, "Greatest.WidthKM", path.Greatest.WidthKM, global.PathWidthKM, 1e-9)

	foundGreatest := false
	for _, point := range path.CenterLine {
		if math.Abs(point.JDE-global.GreatestEclipse) <= solarEclipsePathDuplicateTimeDays {
			foundGreatest = true
			break
		}
	}
	if !foundGreatest {
		t.Fatalf("center line should include greatest eclipse JDE %.12f", global.GreatestEclipse)
	}
}

func TestSolarEclipseCentralPathTargetSpacingRefinesSamples(t *testing.T) {
	seedJDE := JDCalc(2024, 4, 8)
	coarse := SolarEclipseCentralPath(seedJDE, SolarEclipsePathOptions{StepDays: 20.0 / 1440.0})
	refined := SolarEclipseCentralPath(seedJDE, SolarEclipsePathOptions{
		StepDays:        20.0 / 1440.0,
		TargetSpacingKM: 120,
	})

	if len(coarse.CenterLine) == 0 || len(refined.CenterLine) == 0 {
		t.Fatalf("expected path points: coarse=%d refined=%d", len(coarse.CenterLine), len(refined.CenterLine))
	}
	if len(refined.CenterLine) <= len(coarse.CenterLine) {
		t.Fatalf("target spacing should refine samples: coarse=%d refined=%d", len(coarse.CenterLine), len(refined.CenterLine))
	}
	for i := 1; i < len(refined.CenterLine); i++ {
		distanceKM := solarEclipsePathDistanceKM(refined.CenterLine[i-1], refined.CenterLine[i])
		if distanceKM > 120.1 {
			t.Fatalf("segment %d too long: got %.6f km want <= 120.1 km", i, distanceKM)
		}
	}
}

func TestSolarEclipseCentralPathAdaptiveCurvature2543(t *testing.T) {
	seed := JDCalc(2543, 10, 29)
	path := SolarEclipseCentralPath(seed, SolarEclipsePathOptions{
		StepDays:        20.0 / 1440.0,
		TargetSpacingKM: 700,
	})
	if len(path.CenterLine) < 2 {
		t.Fatalf("2543-10-29 center line has %d points, want at least two", len(path.CenterLine))
	}
	coarse := SolarEclipseCentralPath(seed, SolarEclipsePathOptions{StepDays: 20.0 / 1440.0})
	if len(path.CenterLine) <= len(coarse.CenterLine) {
		t.Fatalf("adaptive curvature refinement did not add samples: coarse=%d refined=%d",
			len(coarse.CenterLine), len(path.CenterLine))
	}

	solver := newSolarEclipseSolver(CalcMoonSHByJDE(seed, 0), SolarEclipseModelNASABulletinSplitK)
	tolerance := math.Max(solarEclipsePathMinimumCurvatureKM,
		path.TargetSpacingKM*solarEclipsePathAdaptiveCurvatureFraction)
	for index := 1; index < len(path.CenterLine); index++ {
		start, end := path.CenterLine[index-1], path.CenterLine[index]
		middle, ok := solver.centralPathPointAt((start.JDE + end.JDE) / 2)
		if !ok {
			continue
		}
		geodesicMiddle := solarEclipsePathSphericalInterpolate(start, end, 0.5)
		if errorKM := solarEclipsePathDistanceKM(middle, geodesicMiddle); errorKM > tolerance+1e-6 {
			t.Fatalf("segment %d midpoint error=%.6f km, want <= %.6f km", index, errorKM, tolerance)
		}
	}
}

func TestSolarEclipseCentralPathLimitsKeepValidBoundarySamples(t *testing.T) {
	path := SolarEclipseCentralPath(JDCalc(2008, 8, 1), SolarEclipsePathOptions{
		StepDays: 2.0 / 1440.0,
	})
	if len(path.CenterLine) < 3 || len(path.NorthernLimit) < 3 || len(path.NorthernLimit) != len(path.SouthernLimit) {
		t.Fatalf("unexpected 2008 path samples: center=%d north=%d south=%d", len(path.CenterLine), len(path.NorthernLimit), len(path.SouthernLimit))
	}
	if path.NorthernLimit[0].JDE >= path.CenterLine[0].JDE ||
		path.NorthernLimit[len(path.NorthernLimit)-1].JDE <= path.CenterLine[len(path.CenterLine)-1].JDE {
		t.Fatal("central limit samples should include the external contact caps")
	}
	if solarEclipsePathDistanceKM(path.NorthernLimit[0], path.SouthernLimit[0]) > 1 ||
		solarEclipsePathDistanceKM(path.NorthernLimit[len(path.NorthernLimit)-1], path.SouthernLimit[len(path.SouthernLimit)-1]) > 1 {
		t.Fatal("external contact caps should collapse to one contact point")
	}
	for index := 1; index < len(path.NorthernLimit); index++ {
		if solarEclipsePathDistanceKM(path.NorthernLimit[index-1], path.NorthernLimit[index]) > 500 {
			t.Fatalf("northern limit branch jumps at %d", index)
		}
		if solarEclipsePathDistanceKM(path.SouthernLimit[index-1], path.SouthernLimit[index]) > 500 {
			t.Fatalf("southern limit branch jumps at %d", index)
		}
	}
}

func TestSolarEclipseCentralPathLimitsRemainStrictlyOrdered19851101(t *testing.T) {
	path := SolarEclipseCentralPath(JDCalc(1985, 11, 1), SolarEclipsePathOptions{
		StepDays: 10.0 / 1440.0,
	})
	if len(path.CenterLine) < 2 || len(path.NorthernLimit) < 2 ||
		len(path.NorthernLimit) != len(path.SouthernLimit) {
		t.Fatalf("unexpected 1985-11-01 path sizes: center=%d north=%d south=%d",
			len(path.CenterLine), len(path.NorthernLimit), len(path.SouthernLimit))
	}
	for index := 1; index < len(path.CenterLine); index++ {
		if path.CenterLine[index].JDE <= path.CenterLine[index-1].JDE {
			t.Fatalf("center-line times are not strictly increasing at %d", index)
		}
	}
	for index := 1; index < len(path.NorthernLimit); index++ {
		if path.NorthernLimit[index].JDE <= path.NorthernLimit[index-1].JDE ||
			path.SouthernLimit[index].JDE <= path.SouthernLimit[index-1].JDE {
			t.Fatalf("central-limit times are not strictly increasing at %d: north %.12f -> %.12f, south %.12f -> %.12f",
				index, path.NorthernLimit[index-1].JDE, path.NorthernLimit[index].JDE,
				path.SouthernLimit[index-1].JDE, path.SouthernLimit[index].JDE)
		}
		if math.Abs(path.NorthernLimit[index].JDE-path.SouthernLimit[index].JDE) > solarEclipsePathDuplicateTimeDays {
			t.Fatalf("central-limit sample %d times do not match", index)
		}
	}
}

func TestSolarEclipseRiseSetRawTopologyAvoidsPolarTraceExplosion(t *testing.T) {
	seed := JDCalc(2309, 6, 9)
	result := SolarEclipse(seed)
	solver := newSolarEclipseSolver(CalcMoonSHByJDE(seed, 0), SolarEclipseModelNASABulletinSplitK)
	curves, junctions, actualStep := solver.sampleRiseSetCurves(
		result.PartialBeginOnEarth, result.PartialEndOnEarth, result.GreatestEclipse,
		2.0/1440.0,
	)
	if !solarEclipseRiseSetRawTopologyUsable(curves) {
		t.Fatalf("2309-06-09 sampled topology is not bounded: curves=%d", len(curves))
	}
	solver.finalizeRiseSetCurveTopology(curves, actualStep, junctions)
	if !solarEclipseRiseSetCurveTopologyComplete(curves) {
		t.Fatal("2309-06-09 sampled topology should close without continuation tracing")
	}
}

func TestSolarEclipseNonCentralGreatestHorizonFoldsRemainTimedSegments(t *testing.T) {
	for _, date := range [][3]int{
		{1656, 7, 21},
		{1928, 5, 19},
		{1967, 11, 2},
	} {
		date := date
		t.Run(fmt.Sprintf("%04d-%02d-%02d", date[0], date[1], date[2]), func(t *testing.T) {
			if date[0] == 1967 {
				// 该日的折返只在旧模型几何下出现，用来覆盖拆段分支。
				ephemtest.PinLegacy(t)
			}
			result := SolarEclipsePartialFootprints(JDCalc(date[0], date[1], float64(date[2])), SolarEclipsePartialFootprintOptions{
				StepDays: 2.0 / 1440.0, BoundaryPoints: 24, RiseSetStepDays: 2.0 / 1440.0,
			})
			if result.Eclipse.Type != SolarEclipseTotal || result.Eclipse.Centrality != SolarEclipseNonCentral {
				t.Fatalf("type=%s centrality=%s, want non-central total", result.Eclipse.Type, result.Eclipse.Centrality)
			}
			greatestSegments := 0
			for index := range result.RiseSetCurves {
				curve := &result.RiseSetCurves[index]
				if curve.Phase == RiseSetPhaseGreatest && curve.Direction == RiseSetDirectionRise {
					if len(curve.Segments) > greatestSegments {
						greatestSegments = len(curve.Segments)
					}
				}
				if curve.Phase == RiseSetPhaseGreatest && curve.Direction == RiseSetDirectionSet {
					if len(curve.Segments) > greatestSegments {
						greatestSegments = len(curve.Segments)
					}
				}
				for segmentIndex, segment := range curve.Segments {
					if len(segment) < 2 {
						t.Fatalf("curve %s/%s segment %d has %d points", curve.Phase, curve.Direction, segmentIndex, len(segment))
					}
					for pointIndex := 1; pointIndex < len(segment); pointIndex++ {
						if segment[pointIndex].JDE <= segment[pointIndex-1].JDE+solarEclipseRiseSetTimeEpsilonDays {
							t.Fatalf("curve %s/%s segment %d folds at %d: %.12f -> %.12f",
								curve.Phase, curve.Direction, segmentIndex, pointIndex,
								segment[pointIndex-1].JDE, segment[pointIndex].JDE)
						}
					}
				}
			}
			if greatestSegments < 2 {
				t.Fatalf("greatest horizon fold was not split into independent timed branches")
			}
		})
	}
}

func TestSolarEclipseCentralPathPartialHasNoCenterLine(t *testing.T) {
	path := SolarEclipseCentralPath(JDCalc(2025, 3, 29), SolarEclipsePathOptions{})

	if path.Eclipse.Type != SolarEclipsePartial {
		t.Fatalf("unexpected eclipse type: got %s want %s", path.Eclipse.Type, SolarEclipsePartial)
	}
	if path.Eclipse.HasCentral {
		t.Fatalf("partial eclipse should not have central path")
	}
	if len(path.CenterLine) != 0 || len(path.NorthernLimit) != 0 || len(path.SouthernLimit) != 0 {
		t.Fatalf(
			"partial eclipse should not return central path points: center=%d north=%d south=%d",
			len(path.CenterLine),
			len(path.NorthernLimit),
			len(path.SouthernLimit),
		)
	}
}

func TestSolarEclipsePartialFootprintsIncludeGreatest(t *testing.T) {
	seedJDE := JDCalc(2024, 4, 8)
	global := SolarEclipse(seedJDE)
	footprints := SolarEclipsePartialFootprints(seedJDE, SolarEclipsePartialFootprintOptions{
		StepDays:       30.0 / 1440.0,
		BoundaryPoints: 72,
	})

	if footprints.Eclipse.Type != SolarEclipseTotal {
		t.Fatalf("unexpected eclipse type: got %s want %s", footprints.Eclipse.Type, SolarEclipseTotal)
	}
	if footprints.BoundaryPoints != 72 {
		t.Fatalf("boundary points mismatch: got %d want 72", footprints.BoundaryPoints)
	}
	if len(footprints.Footprints) == 0 {
		t.Fatalf("expected partial footprints")
	}

	foundGreatest := false
	for _, footprint := range footprints.Footprints {
		if math.Abs(footprint.JDE-global.GreatestEclipse) <= solarEclipsePathDuplicateTimeDays {
			foundGreatest = true
		}
		if len(footprint.Boundaries) == 0 {
			t.Fatalf("footprint at %.12f has no boundaries", footprint.JDE)
		}
		for _, boundary := range footprint.Boundaries {
			if len(boundary) == 0 {
				t.Fatalf("footprint at %.12f has an empty boundary segment", footprint.JDE)
			}
			for _, point := range boundary {
				if math.Abs(point.JDE-footprint.JDE) > 1e-12 {
					t.Fatalf("point JDE mismatch: got %.12f want %.12f", point.JDE, footprint.JDE)
				}
				if point.Longitude < -180 || point.Longitude > 180 {
					t.Fatalf("longitude out of range: %.9f", point.Longitude)
				}
				if point.Latitude < -90 || point.Latitude > 90 {
					t.Fatalf("latitude out of range: %.9f", point.Latitude)
				}
			}
		}
		assertSolarEclipseFootprintClosedFlag(t, footprint)
	}
	if !foundGreatest {
		t.Fatalf("partial footprints should include greatest eclipse JDE %.12f", global.GreatestEclipse)
	}
	if footprints.Footprints[0].Closed {
		t.Fatal("grazing first footprint must remain open for horizon closure")
	}
}

func TestSolarEclipsePartialFootprintBoundarySpacing20431003(t *testing.T) {
	seedJDE := JDCalc(2043, 10, 3)
	global := SolarEclipse(seedJDE)
	solver := newSolarEclipseSolver(
		CalcMoonSHByJDE(seedJDE, 0),
		SolarEclipseModelNASABulletinSplitK,
	)
	footprints, _, _ := solver.partialFootprints(
		global.PartialBeginOnEarth,
		global.PartialEndOnEarth,
		global.GreatestEclipse,
		SolarEclipsePartialFootprintOptions{
			StepDays:       2.0 / 1440.0,
			BoundaryPoints: 96,
		},
	)
	if len(footprints) == 0 {
		t.Fatal("expected 2043-10-03 penumbral footprints")
	}

	maximumDistance := 0.0
	var maximumTime float64
	for _, footprint := range footprints {
		for _, boundary := range footprint.Boundaries {
			for index := 1; index < len(boundary); index++ {
				distance := solarEclipsePathDistanceKM(boundary[index-1], boundary[index])
				if distance > maximumDistance {
					maximumDistance = distance
					maximumTime = footprint.JDE
				}
			}
		}
	}
	if maximumDistance > 250 {
		t.Fatalf("2043-10-03 penumbral footprint has a %.1f km boundary chord at JDE %.12f, want <=250 km",
			maximumDistance, maximumTime)
	}
}

func TestSolarEclipseFootprintBoundaryBudgetIsDeterministic(t *testing.T) {
	if got := solarEclipseEffectiveBoundaryPoints(1000, 1440); got != 1440 {
		t.Fatalf("ordinary event boundary points=%d, want requested 1440", got)
	}
	if got := solarEclipseEffectiveBoundaryPoints(30000, 1440); got != 66 {
		t.Fatalf("dense event boundary points=%d, want point-budget cap 66", got)
	}
	if got := solarEclipseMaximumBoundaryPoints(30000); got != 66 {
		t.Fatalf("dense event maximum boundary points=%d, want 66", got)
	}
	if got := solarEclipseEffectiveBoundaryPoints(1, 4); got != solarEclipsePartialFootprintMinBoundaryPoints {
		t.Fatalf("minimum boundary points=%d, want %d", got, solarEclipsePartialFootprintMinBoundaryPoints)
	}
}

func TestSolarEclipseMagnitudeContourInputHasBoundedCardinality(t *testing.T) {
	values := make([]float64, solarEclipseMagnitudeContourMaxValues+8)
	for index := range values {
		values[index] = float64(index+1) / 100
	}
	options := normalizeSolarEclipsePartialFootprintOptions(
		SolarEclipsePartialFootprintOptions{MagnitudeValues: values},
	)
	if len(options.MagnitudeValues) != solarEclipseMagnitudeContourMaxValues {
		t.Fatalf("magnitude values=%d, want %d", len(options.MagnitudeValues), solarEclipseMagnitudeContourMaxValues)
	}
}

func TestSolarEclipsePartialBandContoursMeetHorizonFootprintTracks(t *testing.T) {
	seed := JDCalc(2009, 7, 22)
	result := SolarEclipsePartialFootprints(seed, SolarEclipsePartialFootprintOptions{
		StepDays: 2.0 / 1440.0, BoundaryPoints: 96,
	})
	if len(result.PartialBandContours) != 2 {
		t.Fatalf("partial-band contours=%d, want two zero-magnitude envelopes", len(result.PartialBandContours))
	}
	solver := newSolarEclipseSolver(CalcMoonSHByJDE(seed, 0), SolarEclipseModelNASABulletinSplitK)
	for contourIndex, contour := range result.PartialBandContours {
		if len(contour) < 2 {
			t.Fatalf("partial-band contour %d has %d points", contourIndex, len(contour))
		}
		for endpointIndex, endpoint := range []SolarEclipsePathPoint{contour[0], contour[len(contour)-1]} {
			if math.Abs(endpoint.SunAltitude) > 1e-4 {
				t.Fatalf("contour %d endpoint %d altitude=%.9f, want horizon", contourIndex, endpointIndex, endpoint.SunAltitude)
			}
			state := solver.localStateContextAt(endpoint.JDE).stateAt(endpoint.Longitude*rad, endpoint.Latitude*rad, 0)
			if magnitude := solarEclipseLocalMagnitude(state); math.Abs(magnitude) > 2e-6 {
				t.Fatalf("contour %d endpoint %d magnitude=%.9f, want zero", contourIndex, endpointIndex, magnitude)
			}
			minimum := math.Inf(1)
			for _, curve := range result.RiseSetCurves {
				if curve.Phase == RiseSetPhaseGreatest {
					continue
				}
				for _, segment := range curve.Segments {
					for _, point := range segment {
						minimum = math.Min(minimum, solarEclipsePathDistanceKM(endpoint, point))
					}
				}
			}
			if minimum > 10 {
				t.Fatalf("contour %d endpoint %d misses the start/end horizon network by %.3f km", contourIndex, endpointIndex, minimum)
			}
		}
	}
}

func TestSolarEclipseDisableRiseSetAlsoSkipsPartialBandTopology(t *testing.T) {
	result := SolarEclipsePartialFootprints(JDCalc(2009, 7, 22), SolarEclipsePartialFootprintOptions{
		StepDays: 10.0 / 1440.0, BoundaryPoints: 24, DisableRiseSetCurves: true,
	})
	if len(result.RiseSetCurves) != 0 || len(result.PartialBandContours) != 0 {
		t.Fatalf("disabled rise/set returned curves=%d partial-band contours=%d",
			len(result.RiseSetCurves), len(result.PartialBandContours))
	}
}

func TestSolarEclipseMagnitudeContoursIncludeNonCentralAnnularBand(t *testing.T) {
	result := SolarEclipsePartialFootprints(JDCalc(2014, 4, 29), SolarEclipsePartialFootprintOptions{
		StepDays:              10.0 / 1440.0,
		BoundaryPoints:        24,
		CentralShadowStepDays: 2.0 / 1440.0,
		MagnitudeValues:       []float64{0.4, 0.8, 1.0},
	})
	if result.Eclipse.Type != SolarEclipseAnnular || result.Eclipse.Centrality != SolarEclipseNonCentral {
		t.Fatalf("unexpected eclipse classification: type=%s centrality=%s", result.Eclipse.Type, result.Eclipse.Centrality)
	}
	if len(result.CentralShadowFootprints) < 3 {
		t.Fatalf("expected non-central antumbral footprints, got %d", len(result.CentralShadowFootprints))
	}
	if len(result.MagnitudeContours) != 2 {
		t.Fatalf("expected two magnitude contours below the annular maximum, got %d", len(result.MagnitudeContours))
	}
	for _, contour := range result.MagnitudeContours {
		if len(contour.Segments) == 0 {
			t.Fatalf("magnitude %.2f has no continuous envelope segments", contour.Magnitude)
		}
		for segmentIndex, segment := range contour.Segments {
			if len(segment) < 2 {
				t.Fatalf("magnitude %.2f segment %d has %d points", contour.Magnitude, segmentIndex, len(segment))
			}
			for index := 1; index < len(segment); index++ {
				if distance := solarEclipsePathDistanceKM(segment[index-1], segment[index]); distance > 1.1*solarEclipseMagnitudeContourTargetSpacingKM {
					t.Fatalf("magnitude %.2f segment %d interval %d distance %.3f km",
						contour.Magnitude, segmentIndex, index, distance)
				}
			}
		}
	}
}

func TestSolarEclipseMagnitudeContoursAllowTotalityValuesAboveOne(t *testing.T) {
	result := SolarEclipsePartialFootprints(JDCalc(2024, 4, 8), SolarEclipsePartialFootprintOptions{
		StepDays: 10.0 / 1440.0, MagnitudeValues: []float64{1.01},
	})
	if len(result.MagnitudeContours) != 1 || result.MagnitudeContours[0].Magnitude != 1.01 {
		t.Fatalf("magnitude contours = %#v, want one contour at 1.01", result.MagnitudeContours)
	}
}

func TestSolarEclipseMagnitudeContoursCoverHybridAndDeepTotalValues(t *testing.T) {
	for _, test := range []struct {
		name              string
		seed              float64
		values            []float64
		minimumSegmentLen int
	}{
		{name: "2023 hybrid", seed: JDCalc(2023, 4, 20), values: []float64{1.005, 1.01, 1.012}, minimumSegmentLen: 2},
		{name: "2035 total", seed: JDCalc(2035, 9, 2), values: []float64{1.001, 1.01, 1.02}, minimumSegmentLen: 2},
	} {
		result := SolarEclipsePartialFootprints(test.seed, SolarEclipsePartialFootprintOptions{
			StepDays: 10.0 / 1440.0, MagnitudeValues: test.values, DisableRiseSetCurves: true,
		})
		solver := newSolarEclipseSolver(CalcMoonSHByJDE(test.seed, 0), SolarEclipseModelNASABulletinSplitK)
		if len(result.MagnitudeContours) != len(test.values) {
			t.Fatalf("%s contours=%d, want %d", test.name, len(result.MagnitudeContours), len(test.values))
		}
		for _, contour := range result.MagnitudeContours {
			if len(contour.Segments) == 0 {
				t.Fatalf("%s magnitude %.3f has no segments", test.name, contour.Magnitude)
			}
			for segmentIndex, segment := range contour.Segments {
				if len(segment) < test.minimumSegmentLen {
					t.Fatalf("%s magnitude %.3f segment %d has %d points", test.name, contour.Magnitude, segmentIndex, len(segment))
				}
				for _, point := range segment {
					evaluation := solarEclipseRiseSetEvaluation{jd: point.JDE, center: newLocalSolarEclipseStateContext(point.JDE, solver.params)}
					state := evaluation.center.stateAt(point.Longitude*rad, point.Latitude*rad, 0)
					if math.Abs(solarEclipseMagnitudeAtTarget(state, contour.Magnitude)-contour.Magnitude) > 2e-6 {
						t.Fatalf("%s magnitude %.3f segment %d point magnitude=%.9f", test.name, contour.Magnitude, segmentIndex, solarEclipseMagnitudeAtTarget(state, contour.Magnitude))
					}
				}
			}
		}
	}
}

func TestSolarEclipseHybridMagnitudeOneContoursMeetCenterLineTransitions(t *testing.T) {
	for _, test := range []struct {
		year, month, day int
		transitions      int
	}{
		{1827, 10, 20, 2},
		{1845, 10, 30, 2},
		{1854, 11, 20, 1},
		{1909, 6, 17, 2},
		{1986, 10, 3, 2},
		{2013, 11, 3, 1},
		{2023, 4, 20, 2},
		{2172, 10, 17, 1},
	} {
		name := fmt.Sprintf("%04d-%02d-%02d", test.year, test.month, test.day)
		t.Run(name, func(t *testing.T) {
			seed := JDCalc(test.year, test.month, float64(test.day))
			partial := SolarEclipsePartialFootprints(seed, SolarEclipsePartialFootprintOptions{
				StepDays: 10.0 / 1440.0, MagnitudeValues: []float64{1}, DisableRiseSetCurves: true,
			})
			central := SolarEclipseCentralPath(seed, SolarEclipsePathOptions{StepDays: 2.0 / 1440.0})
			if partial.Eclipse.Type != SolarEclipseHybrid || len(partial.MagnitudeContours) != 1 {
				t.Fatalf("unexpected hybrid result: type=%s contours=%d", partial.Eclipse.Type, len(partial.MagnitudeContours))
			}
			segments := partial.MagnitudeContours[0].Segments
			if len(segments) != 2 {
				t.Fatalf("magnitude 1 segments=%d, want two", len(segments))
			}
			solver := newSolarEclipseSolver(CalcMoonSHByJDE(seed, 0), SolarEclipseModelNASABulletinSplitK)
			transitions := solver.centralMagnitudeOneTransitions(central.CenterLine)
			if len(transitions) != test.transitions {
				t.Fatalf("center-line transitions=%d, want %d", len(transitions), test.transitions)
			}
			for transitionIndex, transition := range transitions {
				centerMatches := 0
				for _, point := range central.CenterLine {
					if math.Abs(point.JDE-transition.JDE) <= solarEclipseRiseSetTimeEpsilonDays &&
						solarEclipsePathDistanceKM(point, transition) <= 0.01 {
						centerMatches++
					}
				}
				if centerMatches != 1 {
					t.Fatalf("transition %d center-line matches=%d, want one shared vertex", transitionIndex, centerMatches)
				}
				matches := 0
				for _, segment := range segments {
					for _, endpoint := range []SolarEclipsePathPoint{segment[0], segment[len(segment)-1]} {
						if solarEclipsePathDistanceKM(endpoint, transition) <= 0.1 {
							matches++
						}
					}
				}
				if matches != 2 {
					t.Fatalf("transition %d endpoint matches=%d, want two", transitionIndex, matches)
				}
			}
		})
	}
}

func TestSolarEclipseHybridMagnitudeOneContoursKeepTheirCentralPathSide11440703(t *testing.T) {
	seed := JDCalc(1144, 7, 3)
	partial := SolarEclipsePartialFootprints(seed, SolarEclipsePartialFootprintOptions{
		StepDays: 2.0 / 1440.0, MagnitudeValues: []float64{1}, DisableRiseSetCurves: true,
	})
	if partial.Eclipse.Type != SolarEclipseHybrid || len(partial.MagnitudeContours) != 1 {
		t.Fatalf("unexpected hybrid result: type=%s contours=%d", partial.Eclipse.Type, len(partial.MagnitudeContours))
	}
	solver := newSolarEclipseSolver(CalcMoonSHByJDE(seed, 0), SolarEclipseModelNASABulletinSplitK)
	for segmentIndex, segment := range partial.MagnitudeContours[0].Segments {
		var branchSign float64
		for pointIndex, point := range segment {
			sign, ok := solver.magnitudeContourBranchSign(point)
			if !ok {
				continue
			}
			if branchSign != 0 && branchSign*sign < 0 {
				t.Fatalf("segment %d switches central-path side at point %d", segmentIndex, pointIndex)
			}
			branchSign = sign
		}
		if branchSign == 0 {
			t.Fatalf("segment %d has no resolved central-path side", segmentIndex)
		}
	}
}

func TestSolarEclipseMagnitudeContoursMatchLocalMaximumMagnitude(t *testing.T) {
	tests := []struct {
		year, month, day int
		magnitude        float64
	}{
		{2008, 8, 1, 0.2},
		{2014, 4, 29, 0.4},
		{2024, 4, 8, 0.2},
	}
	for _, test := range tests {
		seed := JDCalc(test.year, test.month, float64(test.day))
		result := SolarEclipsePartialFootprints(seed, SolarEclipsePartialFootprintOptions{
			StepDays: 10.0 / 1440.0, MagnitudeValues: []float64{test.magnitude},
		})
		if len(result.MagnitudeContours) != 1 || len(result.MagnitudeContours[0].Segments) == 0 {
			t.Fatalf("%04d-%02d-%02d magnitude %.1f contour is absent", test.year, test.month, test.day, test.magnitude)
		}
		for _, segment := range result.MagnitudeContours[0].Segments {
			stride := len(segment) / 5
			if stride < 1 {
				stride = 1
			}
			for index := 0; index < len(segment); index += stride {
				point := segment[index]
				local := LocalSolarEclipse(seed, point.Longitude, point.Latitude, 0)
				if !local.HasPartial || math.Abs(local.Magnitude-test.magnitude) > 2e-4 {
					t.Fatalf("%04d-%02d-%02d contour %.1f at %.6f, %.6f has local maximum %.9f",
						test.year, test.month, test.day, test.magnitude, point.Longitude, point.Latitude, local.Magnitude)
				}
			}
		}
	}
}

func TestSolarEclipseMagnitudeContoursReachGreatestRiseSetBoundary(t *testing.T) {
	seed := JDCalc(2031, 5, 21)
	result := SolarEclipsePartialFootprints(seed, SolarEclipsePartialFootprintOptions{
		StepDays: 10.0 / 1440.0, MagnitudeValues: []float64{0.8},
	})
	if len(result.MagnitudeContours) != 1 {
		t.Fatalf("magnitude contours=%d, want one", len(result.MagnitudeContours))
	}
	if len(result.MagnitudeContours[0].Segments) != 2 {
		t.Fatalf("magnitude contour segments=%d, want two", len(result.MagnitudeContours[0].Segments))
	}
	solver := newSolarEclipseSolver(CalcMoonSHByJDE(seed, 0), SolarEclipseModelNASABulletinSplitK)
	for segmentIndex, segment := range result.MagnitudeContours[0].Segments {
		for pointIndex := 1; pointIndex < len(segment); pointIndex++ {
			if distance := solarEclipsePathDistanceKM(segment[pointIndex-1], segment[pointIndex]); distance > 550 {
				t.Fatalf("segment %d interval %d distance=%.3f km, want at most 550 km",
					segmentIndex, pointIndex, distance)
			}
		}
		for _, pointIndex := range []int{0, len(segment) - 1} {
			point := segment[pointIndex]
			evaluation := solarEclipseRiseSetEvaluation{
				jd:     point.JDE,
				center: newLocalSolarEclipseStateContext(point.JDE, solver.params),
				before: newLocalSolarEclipseStateContext(point.JDE-solarEclipseRiseSetDerivativeStepDays, solver.params),
				after:  newLocalSolarEclipseStateContext(point.JDE+solarEclipseRiseSetDerivativeStepDays, solver.params),
			}
			state := evaluation.center.stateAt(point.Longitude*rad, point.Latitude*rad, 0)
			if math.Abs(state.sunAltitudeRad/rad) > 1e-5 {
				t.Fatalf("segment %d endpoint %d Sun altitude=%.6f deg, want horizon",
					segmentIndex, pointIndex, state.sunAltitudeRad/rad)
			}
			if math.Abs(solarEclipseLocalMagnitude(state)-0.8) > 2e-7 {
				t.Fatalf("segment %d endpoint %d magnitude=%.9f, want 0.8",
					segmentIndex, pointIndex, solarEclipseLocalMagnitude(state))
			}
			if math.Abs(evaluation.separationDerivative(point.Longitude, point.Latitude)) > 1e-8 {
				t.Fatalf("segment %d endpoint %d greatest derivative=%.9g",
					segmentIndex, pointIndex, evaluation.separationDerivative(point.Longitude, point.Latitude))
			}
		}
	}
}

func TestSolarEclipseMagnitudeContourEndpointsAcrossEclipseTypes(t *testing.T) {
	tests := []struct {
		year, month, day int
		magnitude        float64
	}{
		{2008, 8, 1, 0.2},
		{2014, 4, 29, 0.4},
		{2024, 4, 8, 0.2},
		{2025, 3, 29, 0.2},
	}
	for _, test := range tests {
		seed := JDCalc(test.year, test.month, float64(test.day))
		result := SolarEclipsePartialFootprints(seed, SolarEclipsePartialFootprintOptions{
			StepDays: 10.0 / 1440.0, MagnitudeValues: []float64{test.magnitude},
		})
		if len(result.MagnitudeContours) != 1 || len(result.MagnitudeContours[0].Segments) == 0 {
			t.Fatalf("%04d-%02d-%02d contour count=%d", test.year, test.month, test.day, len(result.MagnitudeContours))
		}
		for segmentIndex, segment := range result.MagnitudeContours[0].Segments {
			for _, pointIndex := range []int{0, len(segment) - 1} {
				if altitude := math.Abs(segment[pointIndex].SunAltitude); altitude > 1e-5 {
					t.Fatalf("%04d-%02d-%02d segment %d endpoint %d altitude=%.6f deg",
						test.year, test.month, test.day, segmentIndex, pointIndex, altitude)
				}
			}
		}
	}
}

func TestSolarEclipseTotalMagnitudeOneContoursMeetHorizonClosures20260812(t *testing.T) {
	result := SolarEclipsePartialFootprints(JDCalc(2026, 8, 12), SolarEclipsePartialFootprintOptions{
		StepDays: 2.0 / 1440.0, BoundaryPoints: 96, CentralShadowStepDays: 2.0 / 1440.0,
		MagnitudeValues: []float64{1},
	})
	if len(result.MagnitudeContours) != 1 || len(result.MagnitudeContours[0].Segments) != 2 {
		t.Fatalf("magnitude-one contours=%d, want two segments", len(result.MagnitudeContours))
	}
	if len(result.CentralBandHorizonClosures) != 2 {
		t.Fatalf("horizon closures=%d, want rise and set arcs", len(result.CentralBandHorizonClosures))
	}
	var endpoints []SolarEclipsePathPoint
	for segmentIndex, segment := range result.MagnitudeContours[0].Segments {
		if len(segment) < 2 {
			t.Fatalf("segment %d has %d points", segmentIndex, len(segment))
		}
		for _, point := range []SolarEclipsePathPoint{segment[0], segment[len(segment)-1]} {
			if math.Abs(point.SunAltitude) > 1e-5 {
				t.Fatalf("segment %d endpoint altitude=%.9f degrees, want horizon", segmentIndex, point.SunAltitude)
			}
			endpoints = append(endpoints, point)
		}
	}
	for closureIndex, closure := range result.CentralBandHorizonClosures {
		for rootIndex, root := range []SolarEclipsePathPoint{closure[0], closure[len(closure)-1]} {
			minimumDistance := math.Inf(1)
			for _, endpoint := range endpoints {
				minimumDistance = math.Min(minimumDistance, solarEclipsePathDistanceKM(root, endpoint))
			}
			if minimumDistance > 0.1 {
				t.Fatalf("closure %d root %d is %.3f km from every magnitude-one endpoint", closureIndex, rootIndex, minimumDistance)
			}
		}
	}
}

func TestSolarEclipseRiseSetJunctionsRemainConnected20100115(t *testing.T) {
	result := SolarEclipsePartialFootprints(JDCalc(2010, 1, 15), SolarEclipsePartialFootprintOptions{
		StepDays: 10.0 / 1440.0, BoundaryPoints: 180,
	})
	if len(result.RiseSetCurves) != 6 {
		t.Fatalf("rise-set curve count=%d, want six", len(result.RiseSetCurves))
	}
	curveIndex := make(map[solarEclipseRiseSetCurveKey]int, len(result.RiseSetCurves))
	for index, curve := range result.RiseSetCurves {
		curveIndex[solarEclipseRiseSetCurveKey{phase: curve.Phase, direction: curve.Direction}] = index
	}
	stepDays := 10.0 / 1440.0
	direction := RiseSetDirectionRise
	startIndex := curveIndex[solarEclipseRiseSetCurveKey{phase: RiseSetPhaseStart, direction: direction}]
	greatestIndex := curveIndex[solarEclipseRiseSetCurveKey{phase: RiseSetPhaseGreatest, direction: direction}]
	endIndex := curveIndex[solarEclipseRiseSetCurveKey{phase: RiseSetPhaseEnd, direction: direction}]
	for _, seed := range solarEclipseRiseSetUnsharedEndpoints(startIndex, result.RiseSetCurves[startIndex].Segments) {
		for _, phaseIndex := range []int{greatestIndex, endIndex} {
			if _, ok := solarEclipseClosestRiseSetEndpoint(
				seed.point, phaseIndex, result.RiseSetCurves[phaseIndex].Segments, nil, stepDays,
			); !ok {
				t.Fatalf("%s seed at %.9f (%.4f, %.4f) has no phase junction in curve %d",
					direction, seed.point.JDE, seed.point.Longitude, seed.point.Latitude, phaseIndex)
			}
		}
	}
}

func TestSolarEclipseCentralPathLimitsIncludeExternalContacts20100115(t *testing.T) {
	path := SolarEclipseCentralPath(
		JDCalc(2010, 1, 15),
		SolarEclipsePathOptions{StepDays: 10.0 / 1440.0, TargetSpacingKM: 100},
	)
	if len(path.NorthernLimit) < 2 || len(path.NorthernLimit) != len(path.SouthernLimit) {
		t.Fatalf("central limits north=%d south=%d, want paired samples", len(path.NorthernLimit), len(path.SouthernLimit))
	}
	firstNorth := path.NorthernLimit[0]
	firstSouth := path.SouthernLimit[0]
	lastNorth := path.NorthernLimit[len(path.NorthernLimit)-1]
	lastSouth := path.SouthernLimit[len(path.SouthernLimit)-1]
	if solarEclipsePathDistanceKM(firstNorth, firstSouth) > 1 || solarEclipsePathDistanceKM(lastNorth, lastSouth) > 1 {
		t.Fatalf("external limits should collapse at contacts: first=%.3f km last=%.3f km",
			solarEclipsePathDistanceKM(firstNorth, firstSouth), solarEclipsePathDistanceKM(lastNorth, lastSouth))
	}
	if !(firstNorth.JDE < path.CenterLine[0].JDE && lastNorth.JDE > path.CenterLine[len(path.CenterLine)-1].JDE) {
		t.Fatalf("contact caps must bracket center-line endpoints: first=%.12f center-first=%.12f center-last=%.12f last=%.12f",
			firstNorth.JDE, path.CenterLine[0].JDE, path.CenterLine[len(path.CenterLine)-1].JDE, lastNorth.JDE)
	}
}

func TestSolarEclipseCentralPathMeetsGreatestSetCurveAtExactLimit20100115(t *testing.T) {
	seed := JDCalc(2010, 1, 15)
	path := SolarEclipseCentralPath(seed, SolarEclipsePathOptions{
		StepDays: 2.0 / 1440.0, TargetSpacingKM: 100,
	})
	if len(path.CenterLine) < 2 {
		t.Fatalf("center line has %d points, want at least two", len(path.CenterLine))
	}
	last := path.CenterLine[len(path.CenterLine)-1]
	// NASA's path-table Limits row is 36°49.6'N, 121°40.9'E with Sun altitude 0°.
	if math.Abs(last.Longitude-121.6817) > 0.12 || math.Abs(last.Latitude-36.8267) > 0.12 {
		t.Fatalf("center-line limit = (%.6f, %.6f), want NASA limit near (121.6817, 36.8267)",
			last.Longitude, last.Latitude)
	}
	if math.Abs(last.SunAltitude) > 0.01 {
		t.Fatalf("center-line limit Sun altitude = %.9f degrees, want horizon contact", last.SunAltitude)
	}

	footprints := SolarEclipsePartialFootprints(seed, SolarEclipsePartialFootprintOptions{
		StepDays: 10.0 / 1440.0, BoundaryPoints: 24, RiseSetStepDays: 2.0 / 1440.0,
	})
	found := false
	for _, curve := range footprints.RiseSetCurves {
		if curve.Phase != RiseSetPhaseGreatest || curve.Direction != RiseSetDirectionSet {
			continue
		}
		for _, segment := range curve.Segments {
			for _, point := range segment {
				if math.Abs(point.JDE-last.JDE) <= solarEclipseRiseSetTimeEpsilonDays &&
					solarEclipsePathDistanceKM(point, last) <= 0.01 {
					found = true
				}
			}
		}
	}
	if !found {
		t.Fatal("exact center-line limit is not a shared vertex of the greatest/set curve")
	}
}

func TestSolarEclipseCentralPathContactsShareGreatestRiseSetVerticesAcrossTypes(t *testing.T) {
	for _, date := range [][3]int{
		{2008, 8, 1},  // total
		{2010, 1, 15}, // annular
		{2012, 5, 20}, // polar annular
		{2023, 4, 20}, // hybrid and antimeridian
		{2035, 9, 2},  // total
		{2309, 6, 9},  // far-future sunset regression
	} {
		date := date
		t.Run(fmt.Sprintf("%04d-%02d-%02d", date[0], date[1], date[2]), func(t *testing.T) {
			seed := JDCalc(date[0], date[1], float64(date[2]))
			path := SolarEclipseCentralPath(seed, SolarEclipsePathOptions{
				StepDays: 2.0 / 1440.0, TargetSpacingKM: 200,
			})
			if len(path.CenterLine) < 2 {
				t.Fatalf("center line has %d points, want at least two", len(path.CenterLine))
			}
			footprints := SolarEclipsePartialFootprints(seed, SolarEclipsePartialFootprintOptions{
				StepDays: 30.0 / 1440.0, BoundaryPoints: 24, RiseSetStepDays: 5.0 / 1440.0,
			})
			contacts := []struct {
				name      string
				point     SolarEclipsePathPoint
				direction RiseSetDirection
			}{
				{"begin", path.CenterLine[0], RiseSetDirectionRise},
				{"end", path.CenterLine[len(path.CenterLine)-1], RiseSetDirectionSet},
			}
			for _, contact := range contacts {
				if math.Abs(contact.point.SunAltitude) > 0.02 {
					t.Fatalf("%s Sun altitude = %.9f degrees, want horizon limit", contact.name, contact.point.SunAltitude)
				}
				if !finite(contact.point.WidthKM) || contact.point.WidthKM <= 0 || contact.point.WidthKM > 5000 {
					t.Fatalf("%s width = %.6f km, want a finite path width", contact.name, contact.point.WidthKM)
				}
				if !solarEclipseRiseSetCurvesContainPoint(
					footprints.RiseSetCurves, contact.point, RiseSetPhaseGreatest, contact.direction,
				) {
					t.Fatalf("%s is not a shared greatest/%s vertex", contact.name, contact.direction)
				}
			}
		})
	}
}

func solarEclipseRiseSetCurvesContainPoint(
	curves []SolarEclipseRiseSetCurve,
	want SolarEclipsePathPoint,
	phase RiseSetPhase,
	direction RiseSetDirection,
) bool {
	for _, curve := range curves {
		if curve.Phase != phase || curve.Direction != direction {
			continue
		}
		for _, segment := range curve.Segments {
			for _, point := range segment {
				if math.Abs(point.JDE-want.JDE) <= solarEclipseRiseSetTimeEpsilonDays &&
					solarEclipsePathDistanceKM(point, want) <= 0.01 {
					return true
				}
			}
		}
	}
	return false
}

func TestSolarEclipseExactNonCentralContactsRecoverAtPolarLimb21410108(t *testing.T) {
	seed := JDCalc(2141, 1, 8)
	result := solarEclipse(seed, SolarEclipseModelNASABulletinSplitK)
	if result.Type != SolarEclipseAnnular || result.Centrality != SolarEclipseNonCentral {
		t.Fatalf("unexpected eclipse classification: type=%s centrality=%s", result.Type, result.Centrality)
	}
	solver := newSolarEclipseSolver(CalcMoonSHByJDE(seed, 0), SolarEclipseModelNASABulletinSplitK)
	solver.exactCentralContact = true
	first, last, ok := solver.shadowContactPair(result.GreatestEclipse, solarEclipseCentralShadow, false)
	if !ok || first.JDE == 0 || last.JDE <= first.JDE {
		t.Fatalf("non-central contact pair is unavailable: first=%.12f last=%.12f ok=%v", first.JDE, last.JDE, ok)
	}
	for name, point := range map[string]SolarEclipsePathPoint{"U1": first, "U4": last} {
		if !finite(point.Longitude) || !finite(point.Latitude) ||
			point.Longitude < -180 || point.Longitude > 180 || point.Latitude < -90 || point.Latitude > 90 {
			t.Fatalf("%s polar-limb contact is invalid: %+v", name, point)
		}
	}
}

func BenchmarkSolarEclipseMagnitudeContours(b *testing.B) {
	seed := JDCalc(2031, 5, 21)
	global := solarEclipse(seed, SolarEclipseModelNASABulletinSplitK)
	options := SolarEclipsePartialFootprintOptions{
		StepDays: 2.0 / 1440.0, MagnitudeValues: []float64{0.2, 0.4, 0.6, 0.8, 1.0},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		solver := newSolarEclipseSolver(CalcMoonSHByJDE(seed, 0), SolarEclipseModelNASABulletinSplitK)
		contours := solver.magnitudeContours(
			global.PartialBeginOnEarth,
			global.PartialEndOnEarth,
			global.CentralBeginOnEarth,
			global.CentralEndOnEarth,
			global.GreatestEclipse,
			options,
			global.Magnitude,
			global.Type == SolarEclipseHybrid,
		)
		if len(contours) != 4 {
			b.Fatalf("magnitude contours=%d, want four", len(contours))
		}
	}
}

func BenchmarkSolarEclipsePartialFootprintsFull(b *testing.B) {
	seed := JDCalc(2031, 5, 21)
	options := SolarEclipsePartialFootprintOptions{
		StepDays:              2.0 / 1440.0,
		BoundaryPoints:        180,
		CentralShadowStepDays: 10.0 / 1440.0,
		MagnitudeValues:       []float64{0.2, 0.4, 0.6, 0.8, 1.0},
	}
	b.ReportAllocs()
	for iteration := 0; iteration < b.N; iteration++ {
		result := SolarEclipsePartialFootprints(seed, options)
		if len(result.Footprints) == 0 || len(result.MagnitudeContours) != 4 || len(result.RiseSetCurves) == 0 {
			b.Fatalf("incomplete full footprint result: footprints=%d contours=%d rise-set=%d",
				len(result.Footprints), len(result.MagnitudeContours), len(result.RiseSetCurves))
		}
	}
}

func BenchmarkSolarEclipseNonCentralPartialFootprints20431003(b *testing.B) {
	seed := JDCalc(2043, 10, 3)
	options := SolarEclipsePartialFootprintOptions{
		StepDays:              2.0 / 1440.0,
		BoundaryPoints:        96,
		CentralShadowStepDays: 2.0 / 1440.0,
		MagnitudeValues:       []float64{0.2, 0.4, 0.6, 0.8, 1.0},
		RiseSetStepDays:       2.0 / 1440.0,
	}
	b.ReportAllocs()
	for iteration := 0; iteration < b.N; iteration++ {
		result := SolarEclipsePartialFootprints(seed, options)
		if result.Eclipse.Centrality != SolarEclipseNonCentral ||
			len(result.CentralBandSegments) != 1 || len(result.Footprints) == 0 {
			b.Fatalf("incomplete 2043 non-central result: centrality=%s bands=%d footprints=%d",
				result.Eclipse.Centrality, len(result.CentralBandSegments), len(result.Footprints))
		}
	}
}

func TestSolarEclipsePartialFootprintsWorkForPartialOnlyEclipse(t *testing.T) {
	footprints := SolarEclipsePartialFootprints(JDCalc(2025, 3, 29), SolarEclipsePartialFootprintOptions{
		StepDays:       30.0 / 1440.0,
		BoundaryPoints: 72,
	})

	if footprints.Eclipse.Type != SolarEclipsePartial {
		t.Fatalf("unexpected eclipse type: got %s want %s", footprints.Eclipse.Type, SolarEclipsePartial)
	}
	if footprints.Eclipse.HasCentral {
		t.Fatalf("partial-only eclipse should not have central path")
	}
	if len(footprints.Footprints) == 0 {
		t.Fatalf("expected partial footprints for partial-only eclipse")
	}
}

func TestSolarEclipseShadowContactsAgainstNASA2012Baseline(t *testing.T) {
	result := SolarEclipsePartialFootprints(
		solarEclipseUTToTTJDE(time.Date(2012, 5, 20, 0, 0, 0, 0, time.UTC)),
		SolarEclipsePartialFootprintOptions{
			StepDays:              30.0 / 1440.0,
			BoundaryPoints:        72,
			CentralShadowStepDays: 10.0 / 1440.0,
		},
	)
	baseline := []struct {
		name  string
		point SolarEclipsePathPoint
		want  time.Time
	}{
		{"P1", result.P1, time.Date(2012, 5, 20, 20, 56, 7, 0, time.UTC)},
		{"P4", result.P4, time.Date(2012, 5, 21, 2, 49, 21, 500000000, time.UTC)},
		{"U1", result.U1, time.Date(2012, 5, 20, 22, 6, 16, 600000000, time.UTC)},
		{"U2", result.U2, time.Date(2012, 5, 20, 22, 11, 46, 400000000, time.UTC)},
		{"U3", result.U3, time.Date(2012, 5, 21, 1, 33, 42, 800000000, time.UTC)},
		{"U4", result.U4, time.Date(2012, 5, 21, 1, 39, 11, 200000000, time.UTC)},
	}
	for _, contact := range baseline {
		if contact.point.JDE == 0 {
			t.Fatalf("%s contact is absent", contact.name)
		}
		assertLocalSolarEclipseJDEClose(
			t,
			contact.name,
			contact.point.JDE,
			solarEclipseUTToTTJDE(contact.want),
			3*time.Second,
		)
		if math.Abs(contact.point.SunAltitude) > 0.01 {
			t.Fatalf("%s Sun altitude = %.9f degrees, want horizon contact", contact.name, contact.point.SunAltitude)
		}
	}
	if result.P2.JDE != 0 || result.P3.JDE != 0 {
		t.Fatalf("2012 eclipse unexpectedly has P2/P3 contacts: P2=%+v P3=%+v", result.P2, result.P3)
	}
	if len(result.CentralShadowFootprints) == 0 {
		t.Fatal("expected sampled central-shadow footprints")
	}
	if math.Abs(result.CentralShadowStepDays-10.0/1440.0) > 1e-12 {
		t.Fatalf("central shadow step = %.12f days, want ten minutes", result.CentralShadowStepDays)
	}
	for _, footprint := range result.CentralShadowFootprints {
		if len(footprint.Boundaries) == 0 {
			t.Fatalf("central-shadow footprint at %.12f has no boundary", footprint.JDE)
		}
	}
	closedBandFootprints := 0
	for _, footprint := range result.CentralBandFootprints {
		if !footprint.Closed {
			continue
		}
		closedBandFootprints++
		if footprint.JDE < result.U2.JDE-solarEclipsePathDuplicateTimeDays ||
			footprint.JDE > result.U3.JDE+solarEclipsePathDuplicateTimeDays {
			t.Fatalf("closed central-band footprint %.12f is outside U2-U3", footprint.JDE)
		}
	}
	if closedBandFootprints == 0 || closedBandFootprints > 17 {
		t.Fatalf("closed central-band footprints=%d, want a bounded non-empty edge sample", closedBandFootprints)
	}
}

func TestSolarEclipseShadowContactsIncludeP2P3WhenPenumbraEntersEarthDisk(t *testing.T) {
	result := SolarEclipsePartialFootprints(
		solarEclipseUTToTTJDE(time.Date(2024, 4, 8, 0, 0, 0, 0, time.UTC)),
		SolarEclipsePartialFootprintOptions{StepDays: 30.0 / 1440.0, BoundaryPoints: 36},
	)
	for name, contacts := range map[string][]SolarEclipsePathPoint{
		"penumbral": {result.P1, result.P2, result.P3, result.P4},
		"central":   {result.U1, result.U2, result.U3, result.U4},
	} {
		for index, contact := range contacts {
			if contact.JDE == 0 {
				t.Fatalf("%s contact %d is absent", name, index)
			}
			if index > 0 && !(contacts[index-1].JDE < contact.JDE) {
				t.Fatalf("%s contacts out of order at %d: %.12f >= %.12f", name, index, contacts[index-1].JDE, contact.JDE)
			}
		}
	}
	if !(result.P1.JDE < result.U1.JDE && result.U4.JDE < result.P4.JDE) {
		t.Fatalf("central shadow contacts must lie inside partial phase: P1=%v U1=%v U4=%v P4=%v",
			result.P1.JDE, result.U1.JDE, result.U4.JDE, result.P4.JDE)
	}
	if result.CentralShadowFootprints != nil || result.CentralShadowStepDays != 0 {
		t.Fatal("central-shadow footprints must remain disabled by default")
	}
	if len(result.CentralBandFootprints) == 0 || result.CentralBandStepDays <= 0 {
		t.Fatal("default result must retain lightweight central-band end footprints")
	}
}

func TestSolarEclipseCentralBandFootprintsCoverNonCentralAndOneLimitEvents(t *testing.T) {
	for _, fixture := range []struct {
		year, month, day int
		centrality       SolarEclipseCentrality
	}{
		{2003, 5, 31, SolarEclipseCentralOneLimit},
		{2014, 4, 29, SolarEclipseNonCentral},
		{2043, 4, 9, SolarEclipseNonCentral},
		{2043, 10, 3, SolarEclipseNonCentral},
	} {
		result := SolarEclipsePartialFootprints(
			JDCalc(fixture.year, fixture.month, float64(fixture.day)),
			SolarEclipsePartialFootprintOptions{StepDays: 20.0 / 1440.0, BoundaryPoints: 24},
		)
		if result.Eclipse.Centrality != fixture.centrality {
			t.Fatalf("%04d-%02d-%02d centrality=%s, want %s",
				fixture.year, fixture.month, fixture.day, result.Eclipse.Centrality, fixture.centrality)
		}
		if len(result.CentralBandFootprints) < 2 {
			t.Fatalf("%04d-%02d-%02d central-band footprints=%d, want at least two",
				fixture.year, fixture.month, fixture.day, len(result.CentralBandFootprints))
		}
		if fixture.centrality == SolarEclipseNonCentral {
			if len(result.CentralBandSegments) != 1 {
				t.Fatalf("%04d-%02d-%02d central-band regions=%d, want one horizon-closed region",
					fixture.year, fixture.month, fixture.day, len(result.CentralBandSegments))
			}
			solver := newSolarEclipseSolver(
				CalcMoonSHByJDE(JDCalc(fixture.year, fixture.month, float64(fixture.day)), 0),
				SolarEclipseModelNASABulletinSplitK,
			)
			horizonPoints := 0
			junctions := 0
			for segmentIndex, segment := range result.CentralBandSegments {
				if len(segment) < 4 || solarEclipsePathDistanceKM(segment[0], segment[len(segment)-1]) > 0.01 {
					t.Fatalf("%04d-%02d-%02d central-band sweep cell %d is not closed",
						fixture.year, fixture.month, fixture.day, segmentIndex)
				}
				for pointIndex, point := range segment {
					evaluation := solver.magnitudeEvaluationAt(point.JDE)
					state := evaluation.center.stateAt(point.Longitude*rad, point.Latitude*rad, 0)
					centralGap := state.separationRad - math.Abs(state.sunRadiusRad-state.moonInnerRadiusRad)
					if math.Abs(state.sunAltitudeRad/rad) <= 1e-5 {
						horizonPoints++
						if centralGap > 1e-7 {
							t.Fatalf("%04d-%02d-%02d central-band horizon point %d gap=%g, want <= 0",
								fixture.year, fixture.month, fixture.day, pointIndex, centralGap)
						}
						if math.Abs(centralGap) <= 1e-7 &&
							math.Abs(evaluation.separationDerivative(point.Longitude, point.Latitude)) <= 1e-8 {
							junctions++
						}
					}
					if pointIndex == 0 {
						continue
					}
					maximumEdgeKM := 1.6 * solarEclipseCentralBandTargetSpacingKM
					if result.Eclipse.Type == SolarEclipseTotal {
						maximumEdgeKM = 1.1 * solarEclipseNonCentralTotalBandTargetSpacingKM
					}
					if distance := solarEclipsePathDistanceKM(segment[pointIndex-1], segment[pointIndex]); distance > maximumEdgeKM {
						t.Fatalf("%04d-%02d-%02d central-band sweep cell %d edge %d = %.3f km",
							fixture.year, fixture.month, fixture.day, segmentIndex, pointIndex-1, distance)
					}
				}
			}
			if horizonPoints < 4 {
				t.Fatalf("%04d-%02d-%02d central-band horizon points=%d, want a sampled closing arc",
					fixture.year, fixture.month, fixture.day, horizonPoints)
			}
			if junctions < 2 {
				t.Fatalf("%04d-%02d-%02d central-band horizon/contact junctions=%d, want two",
					fixture.year, fixture.month, fixture.day, junctions)
			}
			if err := auditSolarEclipseBandContainsFootprints(
				result.CentralBandSegments, result.CentralBandFootprints,
			); err != nil {
				t.Fatalf("%04d-%02d-%02d central-band containment: %v",
					fixture.year, fixture.month, fixture.day, err)
			}
		}
	}
}

func TestSolarEclipseNonCentralBandFallsBackWhenCriticalEnvelopeLeaks(t *testing.T) {
	result := SolarEclipsePartialFootprints(
		JDCalc(1950, 3, 18),
		SolarEclipsePartialFootprintOptions{
			StepDays: 60.0 / 1440.0, BoundaryPoints: 24, DisableRiseSetCurves: true,
		},
	)
	if result.Eclipse.Centrality != SolarEclipseNonCentral {
		t.Fatalf("1950-03-18 centrality=%s, want non-central", result.Eclipse.Centrality)
	}
	if len(result.CentralBandSegments) != 0 {
		t.Fatalf("leaking critical envelope was retained with %d segments", len(result.CentralBandSegments))
	}
	if err := auditSolarEclipseOpenBandSweep(result.CentralBandFootprints); err != nil {
		t.Fatalf("open-footprint fallback is unusable: %v", err)
	}
}

func TestSolarEclipsePartialBoundarySegmentsRemainClosedAcrossAntimeridian(t *testing.T) {
	samples := []solarEclipsePartialBoundarySample{
		{point: SolarEclipsePathPoint{Longitude: 170, Latitude: 20}, ok: true},
		{point: SolarEclipsePathPoint{Longitude: -170, Latitude: 25}, ok: true},
		{point: SolarEclipsePathPoint{Longitude: -160, Latitude: 10}, ok: true},
		{point: SolarEclipsePathPoint{Longitude: 160, Latitude: 5}, ok: true},
	}
	boundaries, closed := solarEclipsePartialBoundarySegments(samples)
	if !closed {
		t.Fatal("complete spherical boundary must remain closed after antimeridian splitting")
	}
	if len(boundaries) < 2 {
		t.Fatalf("expected antimeridian split, got %d boundary segment(s)", len(boundaries))
	}
	for index, boundary := range boundaries {
		if len(boundary) < 2 {
			t.Fatalf("antimeridian boundary %d has only %d point(s)", index, len(boundary))
		}
		first := boundary[0].Longitude
		last := boundary[len(boundary)-1].Longitude
		if math.Abs(first) != 180 && math.Abs(last) != 180 {
			t.Fatalf("antimeridian boundary %d has no exact map-edge endpoint", index)
		}
	}
}

func TestSolarEclipsePartialFootprintsNoEvent(t *testing.T) {
	footprints := SolarEclipsePartialFootprints(JDCalc(2023, 5, 15), SolarEclipsePartialFootprintOptions{})

	if footprints.Eclipse.Type != SolarEclipseNone {
		t.Fatalf("unexpected eclipse type: got %s want %s", footprints.Eclipse.Type, SolarEclipseNone)
	}
	if len(footprints.Footprints) != 0 {
		t.Fatalf("no eclipse should not return footprints: got %d", len(footprints.Footprints))
	}
}

func TestSolarEclipsePartialAreaCompatibilityWrapper(t *testing.T) {
	seedJDE := JDCalc(2024, 4, 8)
	options := SolarEclipsePartialAreaOptions{
		StepDays:       30.0 / 1440.0,
		BoundaryPoints: 72,
	}
	compat := SolarEclipsePartialArea(seedJDE, options)
	primary := SolarEclipsePartialFootprints(seedJDE, options)

	if compat.Eclipse.Type != primary.Eclipse.Type {
		t.Fatalf("compat type mismatch: got %s want %s", compat.Eclipse.Type, primary.Eclipse.Type)
	}
	if len(compat.Footprints) != len(primary.Footprints) {
		t.Fatalf("compat footprint count mismatch: got %d want %d", len(compat.Footprints), len(primary.Footprints))
	}
}

func assertSolarEclipseFootprintClosedFlag(t *testing.T, footprint SolarEclipsePartialFootprint) {
	t.Helper()
	if !footprint.Closed {
		return
	}
	if len(footprint.Boundaries) == 0 {
		t.Fatal("closed footprint has no boundaries")
	}
	if len(footprint.Boundaries) > 1 {
		return
	}
	boundary := footprint.Boundaries[0]
	if len(boundary) < 2 {
		t.Fatalf("closed footprint boundary too short: got %d", len(boundary))
	}
	first := boundary[0]
	last := boundary[len(boundary)-1]
	if first.Longitude != last.Longitude || first.Latitude != last.Latitude {
		t.Fatalf("closed footprint boundary should repeat first point at end")
	}
}
