package basic

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"
)

func TestSolarEclipseRiseSetCurvesSatisfyLocalPhaseEquations(t *testing.T) {
	result := SolarEclipsePartialFootprints(JDCalc(2024, 4, 8), SolarEclipsePartialFootprintOptions{
		StepDays: 10.0 / 1440.0,
	})
	if len(result.RiseSetCurves) != 6 {
		t.Fatalf("solar rise/set curve count = %d, want 6", len(result.RiseSetCurves))
	}
	solver := newSolarEclipseSolver(CalcMoonSHByJDE(JDCalc(2024, 4, 8), 0), SolarEclipseModelNASABulletinSplitK)
	for _, curve := range result.RiseSetCurves {
		for _, segment := range curve.Segments {
			for _, point := range riseSetSolarTestSamples(segment) {
				evaluation := solarEclipseRiseSetEvaluation{
					jd:     point.JDE,
					center: newLocalSolarEclipseStateContext(point.JDE, solver.params),
					before: newLocalSolarEclipseStateContext(
						point.JDE-solarEclipseRiseSetDerivativeStepDays, solver.params,
					),
					after: newLocalSolarEclipseStateContext(
						point.JDE+solarEclipseRiseSetDerivativeStepDays, solver.params,
					),
				}
				state := evaluation.center.stateAt(point.Longitude*rad, point.Latitude*rad, 0)
				if math.Abs(state.sunAltitudeRad/rad) > 1e-6 {
					t.Fatalf("solar %s/%s altitude = %.9f deg", curve.Phase, curve.Direction, state.sunAltitudeRad/rad)
				}
				assertSolarRiseSetPhase(t, curve, point, state, evaluation)
			}
		}
	}
}

func TestSolarEclipseRiseSetCurvesCloseFoldsAndPhaseJunctions(t *testing.T) {
	result := SolarEclipsePartialFootprints(JDCalc(2031, 5, 21), SolarEclipsePartialFootprintOptions{
		StepDays: 2.0 / 1440.0,
	})
	if len(result.RiseSetCurves) != 6 {
		t.Fatalf("solar rise/set curve count = %d, want 6", len(result.RiseSetCurves))
	}

	curves := make(map[solarEclipseRiseSetCurveKey]SolarEclipseRiseSetCurve, len(result.RiseSetCurves))
	for _, curve := range result.RiseSetCurves {
		key := solarEclipseRiseSetCurveKey{phase: curve.Phase, direction: curve.Direction}
		curves[key] = curve
		if len(curve.Segments) != 2 {
			t.Fatalf("solar %s/%s segments = %d, want 2", curve.Phase, curve.Direction, len(curve.Segments))
		}
		if !solarRiseSetSegmentsShareEndpoint(curve.Segments[0], curve.Segments[1]) {
			t.Fatalf("solar %s/%s branches do not share their fold endpoint", curve.Phase, curve.Direction)
		}
		for segmentIndex, segment := range curve.Segments {
			for pointIndex := 1; pointIndex < len(segment); pointIndex++ {
				if distance := solarEclipsePathDistanceKM(segment[pointIndex-1], segment[pointIndex]); distance > 1.1*solarEclipseRiseSetTargetSpacingKM {
					t.Fatalf("solar %s/%s segment %d interval %d distance = %.3f km",
						curve.Phase, curve.Direction, segmentIndex, pointIndex, distance)
				}
			}
		}
	}

	for _, direction := range []RiseSetDirection{RiseSetDirectionRise, RiseSetDirectionSet} {
		start := curves[solarEclipseRiseSetCurveKey{phase: RiseSetPhaseStart, direction: direction}]
		greatest := curves[solarEclipseRiseSetCurveKey{phase: RiseSetPhaseGreatest, direction: direction}]
		end := curves[solarEclipseRiseSetCurveKey{phase: RiseSetPhaseEnd, direction: direction}]
		junctions := 0
		for _, segment := range start.Segments {
			if solarRiseSetSegmentSharesEndpointWithBoth(segment, greatest.Segments, end.Segments) {
				junctions++
			}
		}
		if junctions != 2 {
			t.Fatalf("solar %s phase junctions = %d, want 2", direction, junctions)
		}
	}
}

func TestSolarEclipseRiseSetCurvesCloseFoldsWithAdditionalBranches(t *testing.T) {
	result := SolarEclipsePartialFootprints(JDCalc(2010, 1, 15), SolarEclipsePartialFootprintOptions{
		StepDays: 10.0 / 1440.0,
	})
	for _, phase := range []RiseSetPhase{RiseSetPhaseGreatest, RiseSetPhaseEnd} {
		var curve SolarEclipseRiseSetCurve
		for _, candidate := range result.RiseSetCurves {
			if candidate.Phase == phase && candidate.Direction == RiseSetDirectionRise {
				curve = candidate
				break
			}
		}
		if len(curve.Segments) != 3 {
			t.Fatalf("solar %s/rise segments = %d, want 3", phase, len(curve.Segments))
		}
		folds := 0
		for first := 0; first < len(curve.Segments); first++ {
			for second := first + 1; second < len(curve.Segments); second++ {
				if solarRiseSetSegmentsShareEndpoint(curve.Segments[first], curve.Segments[second]) {
					folds++
				}
			}
		}
		if folds == 0 {
			t.Fatalf("solar %s/rise additional branches do not share their fold endpoint", phase)
		}
	}
}

func TestSolarEclipseRiseSetCurvesCloseSunsetPhaseJunctions20100115(t *testing.T) {
	result := SolarEclipsePartialFootprints(JDCalc(2010, 1, 15), SolarEclipsePartialFootprintOptions{
		StepDays: 10.0 / 1440.0,
	})
	curves := make(map[solarEclipseRiseSetCurveKey]SolarEclipseRiseSetCurve, len(result.RiseSetCurves))
	for _, curve := range result.RiseSetCurves {
		curves[solarEclipseRiseSetCurveKey{phase: curve.Phase, direction: curve.Direction}] = curve
	}
	start := curves[solarEclipseRiseSetCurveKey{phase: RiseSetPhaseStart, direction: RiseSetDirectionSet}]
	greatest := curves[solarEclipseRiseSetCurveKey{phase: RiseSetPhaseGreatest, direction: RiseSetDirectionSet}]
	end := curves[solarEclipseRiseSetCurveKey{phase: RiseSetPhaseEnd, direction: RiseSetDirectionSet}]
	if got := solarRiseSetSharedPhaseJunctionCount(start.Segments, greatest.Segments, end.Segments); got != 1 {
		t.Fatalf("sunset phase junctions=%d, want one shared endpoint", got)
	}
	if len(greatest.Segments) < 2 {
		t.Fatalf("sunset greatest branches=%d, want a short branch through the phase junction", len(greatest.Segments))
	}
}

func TestSolarEclipseRiseSetCurvesCloseSunsetPhaseJunctions23090609(t *testing.T) {
	result := SolarEclipsePartialFootprints(JDCalc(2309, 6, 9), SolarEclipsePartialFootprintOptions{
		StepDays: 2.0 / 1440.0,
	})
	curves := make(map[solarEclipseRiseSetCurveKey]SolarEclipseRiseSetCurve, len(result.RiseSetCurves))
	for _, curve := range result.RiseSetCurves {
		curves[solarEclipseRiseSetCurveKey{phase: curve.Phase, direction: curve.Direction}] = curve
	}
	start := curves[solarEclipseRiseSetCurveKey{phase: RiseSetPhaseStart, direction: RiseSetDirectionSet}]
	greatest := curves[solarEclipseRiseSetCurveKey{phase: RiseSetPhaseGreatest, direction: RiseSetDirectionSet}]
	end := curves[solarEclipseRiseSetCurveKey{phase: RiseSetPhaseEnd, direction: RiseSetDirectionSet}]
	if got := solarRiseSetSharedPhaseJunctionCount(start.Segments, greatest.Segments, end.Segments); got != 2 {
		t.Fatalf("2309-06-09 sunset phase junctions=%d, want two shared endpoints", got)
	}
}

func TestSolarEclipseRiseSetCurveEndpointsAcrossEclipseTypes(t *testing.T) {
	tests := []struct {
		year, month, day   int
		directionJunctions int
	}{
		{2008, 8, 1, 3},
		{2010, 1, 15, 3},
		{2014, 4, 29, 3},
		{2024, 4, 8, 0},
		{2025, 3, 29, 3},
	}
	for _, test := range tests {
		result := SolarEclipsePartialFootprints(
			JDCalc(test.year, test.month, float64(test.day)),
			SolarEclipsePartialFootprintOptions{StepDays: 2.0 / 1440.0},
		)
		if len(result.RiseSetCurves) != 6 {
			t.Fatalf("%04d-%02d-%02d rise/set curve count = %d, want 6",
				test.year, test.month, test.day, len(result.RiseSetCurves))
		}
		solver := newSolarEclipseSolver(
			CalcMoonSHByJDE(JDCalc(test.year, test.month, float64(test.day)), 0),
			SolarEclipseModelNASABulletinSplitK,
		)
		curves := make(map[solarEclipseRiseSetCurveKey]SolarEclipseRiseSetCurve, len(result.RiseSetCurves))
		for _, curve := range result.RiseSetCurves {
			key := solarEclipseRiseSetCurveKey{phase: curve.Phase, direction: curve.Direction}
			curves[key] = curve
			if len(curve.Segments) == 2 && !solarRiseSetSegmentsShareEndpoint(curve.Segments[0], curve.Segments[1]) {
				t.Fatalf("%04d-%02d-%02d %s/%s branches do not share their fold endpoint",
					test.year, test.month, test.day, curve.Phase, curve.Direction)
			}
			for segmentIndex, segment := range curve.Segments {
				for _, point := range []SolarEclipsePathPoint{segment[0], segment[len(segment)-1]} {
					evaluation := solver.magnitudeEvaluationAt(point.JDE)
					state := evaluation.center.stateAt(point.Longitude*rad, point.Latitude*rad, 0)
					assertSolarRiseSetPhase(t, curve, point, state, evaluation)
				}
				for pointIndex := 1; pointIndex < len(segment); pointIndex++ {
					if segment[pointIndex].JDE <= segment[pointIndex-1].JDE {
						t.Fatalf("%04d-%02d-%02d %s/%s segment %d time is not increasing at %d",
							test.year, test.month, test.day, curve.Phase, curve.Direction, segmentIndex, pointIndex)
					}
				}
			}
		}
		if count := solarRiseSetDirectionJunctionCount(curves); count != test.directionJunctions {
			t.Fatalf("%04d-%02d-%02d shared direction junctions = %d, want %d",
				test.year, test.month, test.day, count, test.directionJunctions)
		}
	}
}

func TestSolarEclipseNonCentralGreatestHorizonGo116386Smoke(t *testing.T) {
	if !go116386() {
		t.Skip("Go 1.16/386 only")
	}
	seed := JDCalc(1957, 10, 23)
	for _, stepSeconds := range []float64{120, 5} {
		result := SolarEclipsePartialFootprints(seed, SolarEclipsePartialFootprintOptions{
			StepDays: 2.0 / 1440, BoundaryPoints: 24, RiseSetStepDays: stepSeconds / 86400,
		})
		if result.Eclipse.Type != SolarEclipseTotal || result.Eclipse.Centrality != SolarEclipseNonCentral {
			t.Fatalf("type=%s centrality=%s, want non-central total", result.Eclipse.Type, result.Eclipse.Centrality)
		}
		if len(result.RiseSetCurves) == 0 || !solarEclipseRiseSetCurveTopologyComplete(result.RiseSetCurves) {
			t.Fatalf("step %gs: rise/set curves are empty or not time-ordered", stepSeconds)
		}
	}
}

func TestSolarEclipseNonCentralGreatestHorizonWithoutTimeFold(t *testing.T) {
	if go116386() {
		t.Skip("Go 1.16/386 的浮点分支在掠射边界多分出段；精确拓扑由 Go 1.17+ 门禁覆盖")
	}
	seed := JDCalc(1957, 10, 23)
	solver := newSolarEclipseSolver(CalcMoonSHByJDE(seed, 0), SolarEclipseModelNASABulletinSplitK)
	for _, stepSeconds := range []float64{120, 5} {
		t.Run(fmt.Sprintf("step_%gs", stepSeconds), func(t *testing.T) {
			result := SolarEclipsePartialFootprints(seed, SolarEclipsePartialFootprintOptions{
				StepDays: 2.0 / 1440, BoundaryPoints: 24, RiseSetStepDays: stepSeconds / 86400,
			})
			if result.Eclipse.Type != SolarEclipseTotal || result.Eclipse.Centrality != SolarEclipseNonCentral {
				t.Fatalf("type=%s centrality=%s, want non-central total", result.Eclipse.Type, result.Eclipse.Centrality)
			}
			if !solarEclipseRiseSetCurveTopologyComplete(result.RiseSetCurves) {
				t.Fatal("horizon branches must be time-ordered and share their endpoints")
			}
			curves := make(map[solarEclipseRiseSetCurveKey]SolarEclipseRiseSetCurve, len(result.RiseSetCurves))
			for _, curve := range result.RiseSetCurves {
				curves[solarEclipseRiseSetCurveKey{curve.Phase, curve.Direction}] = curve
			}
			// Non-central totality need not produce a time fold. Here the
			// greatest branches meet at a sunrise/sunset transition instead.
			for _, direction := range []RiseSetDirection{RiseSetDirectionRise, RiseSetDirectionSet} {
				start := curves[solarEclipseRiseSetCurveKey{RiseSetPhaseStart, direction}]
				greatest := curves[solarEclipseRiseSetCurveKey{RiseSetPhaseGreatest, direction}]
				end := curves[solarEclipseRiseSetCurveKey{RiseSetPhaseEnd, direction}]
				if len(greatest.Segments) != 1 {
					t.Fatalf("greatest/%s branches=%d, want one", direction, len(greatest.Segments))
				}
				if count := solarRiseSetSharedPhaseJunctionCount(start.Segments, greatest.Segments, end.Segments); count != 1 {
					t.Fatalf("%s phase junctions=%d, want one", direction, count)
				}
			}
			rise := curves[solarEclipseRiseSetCurveKey{RiseSetPhaseGreatest, RiseSetDirectionRise}].Segments[0]
			set := curves[solarEclipseRiseSetCurveKey{RiseSetPhaseGreatest, RiseSetDirectionSet}].Segments[0]
			if !solarRiseSetSameEndpoint(rise[len(rise)-1], set[0]) {
				t.Fatal("greatest sunrise and sunset branches do not share their transition")
			}
			// Solve the horizon roots independently of path sampling and
			// endpoint completion, so a missing branch cannot pass by count.
			checked := 0
			for jd := result.Eclipse.PartialBeginOnEarth; jd <= result.Eclipse.PartialEndOnEarth; jd += 60.0 / 86400 {
				for key, roots := range solver.riseSetPointsAt(jd, 720) {
					if key.phase != RiseSetPhaseGreatest {
						continue
					}
					for _, root := range roots {
						distance := math.Inf(1)
						for _, segment := range curves[key].Segments {
							for index := 1; index < len(segment); index++ {
								distance = math.Min(distance, solarEclipseRiseSetChordDeviationKM(root, segment[index-1], segment[index]))
							}
						}
						if distance > solarEclipseRiseSetChordToleranceKM {
							t.Fatalf("greatest/%s root at %.9f is %.6f km from the exported curve", key.direction, jd, distance)
						}
						checked++
					}
				}
			}
			if checked < 100 {
				t.Fatalf("checked %d exact roots, want coverage across the event", checked)
			}
		})
	}
}

func solarRiseSetDirectionJunctionCount(
	curves map[solarEclipseRiseSetCurveKey]SolarEclipseRiseSetCurve,
) int {
	count := 0
	for _, phase := range []RiseSetPhase{RiseSetPhaseStart, RiseSetPhaseGreatest, RiseSetPhaseEnd} {
		rise := curves[solarEclipseRiseSetCurveKey{phase: phase, direction: RiseSetDirectionRise}]
		set := curves[solarEclipseRiseSetCurveKey{phase: phase, direction: RiseSetDirectionSet}]
		for _, segment := range rise.Segments {
			for _, point := range []SolarEclipsePathPoint{segment[0], segment[len(segment)-1]} {
				if solarRiseSetEndpointInSegments(point, set.Segments) {
					count++
				}
			}
		}
	}
	return count
}

func solarRiseSetSharedPhaseJunctionCount(
	start, greatest, end [][]SolarEclipsePathPoint,
) int {
	var junctions []SolarEclipsePathPoint
	for _, segment := range start {
		for _, point := range []SolarEclipsePathPoint{segment[0], segment[len(segment)-1]} {
			if !solarRiseSetEndpointInSegments(point, greatest) || !solarRiseSetEndpointInSegments(point, end) {
				continue
			}
			duplicate := false
			for _, junction := range junctions {
				if solarRiseSetSameEndpoint(point, junction) {
					duplicate = true
					break
				}
			}
			if !duplicate {
				junctions = append(junctions, point)
			}
		}
	}
	return len(junctions)
}

func solarRiseSetSegmentsShareEndpoint(first, second []SolarEclipsePathPoint) bool {
	for _, a := range []SolarEclipsePathPoint{first[0], first[len(first)-1]} {
		for _, b := range []SolarEclipsePathPoint{second[0], second[len(second)-1]} {
			if solarRiseSetSameEndpoint(a, b) {
				return true
			}
		}
	}
	return false
}

func solarRiseSetSegmentSharesEndpointWithBoth(
	segment []SolarEclipsePathPoint,
	firstCandidates, secondCandidates [][]SolarEclipsePathPoint,
) bool {
	for _, point := range []SolarEclipsePathPoint{segment[0], segment[len(segment)-1]} {
		if solarRiseSetEndpointInSegments(point, firstCandidates) && solarRiseSetEndpointInSegments(point, secondCandidates) {
			return true
		}
	}
	return false
}

func solarRiseSetEndpointInSegments(point SolarEclipsePathPoint, segments [][]SolarEclipsePathPoint) bool {
	for _, segment := range segments {
		if solarRiseSetSameEndpoint(point, segment[0]) || solarRiseSetSameEndpoint(point, segment[len(segment)-1]) {
			return true
		}
	}
	return false
}

func solarRiseSetSameEndpoint(first, second SolarEclipsePathPoint) bool {
	return math.Abs(first.JDE-second.JDE) <= 1e-8 && solarEclipsePathDistanceKM(first, second) <= 0.01
}

func TestOccultationRiseSetCurvesSatisfyStarAndPlanetEquations(t *testing.T) {
	location := time.FixedZone("CST", 8*3600)
	star := hr4799OccultationCoordinateForTest()
	starPaths, err := FindStarOccultationPaths(
		time.Date(2025, 6, 5, 0, 0, 0, 0, location),
		time.Date(2025, 6, 6, 0, 0, 0, 0, location),
		star, OccultationPathOptions{Step: 10 * time.Minute},
	)
	if err != nil || len(starPaths) != 1 {
		t.Fatalf("star paths=%d err=%v, want one", len(starPaths), err)
	}
	starContextAt := func(tt float64) occultationRiseSetContext {
		moonRA, moonDec := HMoonGeocentricApparentRaDecN(tt, -1)
		targetRA, targetDec := starApparentRaDecGeocentric(tt, star)
		return newOccultationRiseSetContext(
			tt, moonRA, moonDec, HMoonAwayN(tt, -1),
			targetRA, targetDec, 0, 0,
		)
	}
	assertOccultationRiseSetCurves(t, starPaths[0].RiseSetCurves, starContextAt)

	config, _ := planetOccultationConfigFor(OccultationSaturn)
	planetPaths, err := FindPlanetOccultationPaths(
		time.Date(2024, 8, 21, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 8, 22, 0, 0, 0, 0, time.UTC),
		OccultationSaturn, OccultationPathOptions{Step: 10 * time.Minute},
	)
	if err != nil || len(planetPaths) != 1 {
		t.Fatalf("planet paths=%d err=%v, want one", len(planetPaths), err)
	}
	planetContextAt := func(tt float64) occultationRiseSetContext {
		moonRA, moonDec := HMoonGeocentricApparentRaDecN(tt, -1)
		targetRA, targetDec := config.apparentRaDecN(tt, -1)
		return newOccultationRiseSetContext(
			tt, moonRA, moonDec, HMoonAwayN(tt, -1), targetRA, targetDec,
			config.earthDistanceN(tt, -1)*occultationPathAstronomicalUnitKM, config.equatorialRadiusKM,
		)
	}
	assertOccultationRiseSetCurves(t, planetPaths[0].RiseSetCurves, planetContextAt)
	if !planetPaths[0].HasTotalBand || len(planetPaths[0].TotalRiseSetCurves) != 6 {
		t.Fatalf("planet total rise/set curves=%d hasTotal=%v, want six inner-contact curves",
			len(planetPaths[0].TotalRiseSetCurves), planetPaths[0].HasTotalBand)
	}
	assertOccultationRiseSetCurves(t, planetPaths[0].TotalRiseSetCurves, func(tt float64) occultationRiseSetContext {
		return planetContextAt(tt).withInternalContact()
	})
}

func TestOccultationRiseSetSamplingIsIndependentAndOptional(t *testing.T) {
	location := time.FixedZone("CST", 8*3600)
	start := time.Date(2025, 6, 5, 0, 0, 0, 0, location)
	end := start.Add(24 * time.Hour)
	star := hr4799OccultationCoordinateForTest()
	find := func(options OccultationPathOptions) StarOccultationPath {
		t.Helper()
		paths, err := FindStarOccultationPaths(start, end, star, options)
		if err != nil || len(paths) != 1 {
			t.Fatalf("star paths=%d err=%v, want one", len(paths), err)
		}
		return paths[0]
	}

	fine := find(OccultationPathOptions{Step: 10 * time.Second, RiseSetStep: 10 * time.Minute})
	coarse := find(OccultationPathOptions{Step: 2 * time.Minute, RiseSetStep: 10 * time.Minute})
	if !reflect.DeepEqual(fine.RiseSetCurves, coarse.RiseSetCurves) {
		t.Fatal("rise/set curves changed when only the main path step changed")
	}
	disabled := find(OccultationPathOptions{Step: 2 * time.Minute, DisableRiseSet: true})
	if len(disabled.RiseSetCurves) != 0 {
		t.Fatalf("disabled rise/set curve count = %d, want zero", len(disabled.RiseSetCurves))
	}
	withoutFootprints := find(OccultationPathOptions{Step: 2 * time.Minute, DisableFootprints: true})
	if len(withoutFootprints.Footprints) != 0 {
		t.Fatalf("disabled footprint count = %d, want zero", len(withoutFootprints.Footprints))
	}
	if len(withoutFootprints.BandFootprints) == 0 {
		t.Fatal("disabled dense footprints did not retain compact band support")
	}
	if len(withoutFootprints.RiseSetCurves) != 6 {
		t.Fatalf("rise/set curve count with disabled footprints = %d, want six", len(withoutFootprints.RiseSetCurves))
	}
	if err := (OccultationPathOptions{RiseSetStep: -time.Second}).Validate(); err == nil {
		t.Fatal("negative rise/set step was accepted")
	}
	fineRiseSet := find(OccultationPathOptions{Step: 2 * time.Minute, RiseSetStep: 5 * time.Minute})
	coarseRiseSet := find(OccultationPathOptions{Step: 2 * time.Minute, RiseSetStep: 30 * time.Minute})
	if occultationRiseSetPointCount(coarseRiseSet.RiseSetCurves) >= occultationRiseSetPointCount(fineRiseSet.RiseSetCurves) {
		t.Fatal("coarser occultation rise/set step did not reduce curve samples")
	}
}

func TestOccultationRiseSet20240725CompletesMoonsetGreatestJunction(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2024, time.July, 25, 0, 0, 0, 0, zone)
	paths, err := FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), OccultationSaturn,
		OccultationPathOptions{
			Step: 20 * time.Minute, TargetSpacingKM: 900, DisableFootprints: true,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	curves := make(map[occultationRiseSetCurveKey]OccultationRiseSetCurve)
	for _, curve := range paths[0].RiseSetCurves {
		curves[occultationRiseSetCurveKey{phase: curve.Phase, direction: curve.Direction}] = curve
	}
	greatest := curves[occultationRiseSetCurveKey{phase: RiseSetPhaseGreatest, direction: RiseSetDirectionSet}]
	if len(greatest.Segments) != 1 || len(greatest.Segments[0]) < 2 {
		t.Fatalf("moonset greatest segments=%d, want one usable segment", len(greatest.Segments))
	}
	junction := greatest.Segments[0][len(greatest.Segments[0])-1]
	if junction.Latitude > 35 {
		t.Fatalf("moonset greatest stopped at latitude %.3f, want completed southern junction", junction.Latitude)
	}
	for _, phase := range []RiseSetPhase{RiseSetPhaseStart, RiseSetPhaseEnd} {
		curve := curves[occultationRiseSetCurveKey{phase: phase, direction: RiseSetDirectionSet}]
		shared := false
		for _, segment := range curve.Segments {
			for _, endpoint := range []OccultationPathPoint{segment[0], segment[len(segment)-1]} {
				if math.Abs(endpoint.Time.Sub(junction.Time).Seconds()) <= 1 &&
					occultationPathDistanceKM(endpoint, junction) <= 5 {
					shared = true
					break
				}
			}
		}
		if !shared {
			t.Fatalf("moonset %s curve does not share the greatest-phase junction at %s", phase, junction.Time)
		}
	}
}

func TestOccultationRiseSet20250630MarsClosesMoonsetPhaseJunction(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2025, time.June, 30, 0, 0, 0, 0, zone)
	paths, err := FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), OccultationMars,
		OccultationPathOptions{
			Step: 20 * time.Minute, TargetSpacingKM: 900, RiseSetStep: 5 * time.Minute,
			DisableFootprints: true,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	if len(paths[0].RiseSetCurves) != 6 {
		t.Fatalf("rise/set curve count=%d, want six", len(paths[0].RiseSetCurves))
	}
	assertOccultationRiseSetEndpointsClosed(t, paths[0].RiseSetCurves)
	config, _ := planetOccultationConfigFor(OccultationMars)
	assertOccultationRiseSetCurves(t, paths[0].RiseSetCurves, func(tt float64) occultationRiseSetContext {
		state := planetOccultationEphemerisStateAt(tt, config)
		return newOccultationRiseSetContext(
			tt, state.moonRA, state.moonDec, state.moonDistanceKM,
			state.planetRA, state.planetDec, state.planetDistanceKM,
			config.equatorialRadiusKM,
		)
	})
}

func TestOccultationRiseSet20250105SaturnClosesPolarBranchesWithoutJump(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2025, time.January, 5, 0, 0, 0, 0, zone)
	paths, err := FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), OccultationSaturn,
		OccultationPathOptions{
			Step: 20 * time.Minute, TargetSpacingKM: 900, DisableFootprints: true,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	if len(paths[0].RiseSetCurves) != 6 {
		t.Fatalf("rise/set curve count=%d, want six", len(paths[0].RiseSetCurves))
	}
	path := paths[0]
	assertOccultationRiseSetEndpointNetwork(t, path.RiseSetCurves, path.NorthernLimit, path.SouthernLimit)
	assertOccultationRiseSetNoInstantaneousBranchJumps(t, path.RiseSetCurves)
	assertOccultationRiseSetPolarProjectionSpacing(t, path.RiseSetCurves, 230)
	curves := make(map[occultationRiseSetCurveKey]OccultationRiseSetCurve, len(path.RiseSetCurves))
	for _, curve := range path.RiseSetCurves {
		curves[occultationRiseSetCurveKey{phase: curve.Phase, direction: curve.Direction}] = curve
	}
	startRise := curves[occultationRiseSetCurveKey{phase: RiseSetPhaseStart, direction: RiseSetDirectionRise}]
	startSet := curves[occultationRiseSetCurveKey{phase: RiseSetPhaseStart, direction: RiseSetDirectionSet}]
	greatestRise := curves[occultationRiseSetCurveKey{phase: RiseSetPhaseGreatest, direction: RiseSetDirectionRise}]
	endRise := curves[occultationRiseSetCurveKey{phase: RiseSetPhaseEnd, direction: RiseSetDirectionRise}]
	if !occultationRiseSetCurvePairSharesEndpointInRegion(
		startRise, startSet, -60, 60, 70, 90,
	) {
		t.Fatal("start moonrise/moonset curves do not share their polar direction junction")
	}
	if !occultationRiseSetSegmentsShareEndpointInRegion(startSet.Segments, -60, 60, 70, 90) {
		t.Fatal("moonset start-phase branches do not share their polar fold endpoint")
	}
	if !occultationRiseSetCurvesShareEndpointInRegion(
		startRise, greatestRise, endRise, -100, -70, 20, 50,
	) {
		t.Fatal("moonrise start/greatest/end curves do not share the North American phase junction")
	}
	if !occultationRiseSetSegmentsShareEndpointInRegion(endRise.Segments, -100, -70, 20, 50) {
		t.Fatal("moonrise end-phase branches do not share their North American fold endpoint")
	}
}

func assertOccultationRiseSetPolarProjectionSpacing(
	t *testing.T,
	curves []OccultationRiseSetCurve,
	maximumMercatorKM float64,
) {
	t.Helper()
	for _, curve := range curves {
		for segmentIndex, segment := range curve.Segments {
			for pointIndex := 1; pointIndex < len(segment); pointIndex++ {
				first, second := segment[pointIndex-1], segment[pointIndex]
				if math.Max(math.Abs(first.Latitude), math.Abs(second.Latitude)) < 75 {
					continue
				}
				deltaLongitude := math.Remainder((second.Longitude-first.Longitude)*rad, 2*math.Pi)
				firstY := math.Log(math.Tan(math.Pi/4 + math.Min(85, math.Max(-85, first.Latitude))*rad/2))
				secondY := math.Log(math.Tan(math.Pi/4 + math.Min(85, math.Max(-85, second.Latitude))*rad/2))
				distance := topocentricEarthRadiusKM * math.Hypot(deltaLongitude, secondY-firstY)
				if distance > maximumMercatorKM {
					t.Fatalf("%s/%s segment %d edge %d has %.1f km Web Mercator length, want at most %.1f km",
						curve.Phase, curve.Direction, segmentIndex, pointIndex-1, distance, maximumMercatorKM)
				}
			}
		}
	}
}

func occultationRiseSetCurvePairSharesEndpointInRegion(
	first, second OccultationRiseSetCurve,
	minimumLongitude, maximumLongitude, minimumLatitude, maximumLatitude float64,
) bool {
	for _, segment := range first.Segments {
		for _, point := range []OccultationPathPoint{segment[0], segment[len(segment)-1]} {
			if point.Longitude < minimumLongitude || point.Longitude > maximumLongitude ||
				point.Latitude < minimumLatitude || point.Latitude > maximumLatitude {
				continue
			}
			if occultationRiseSetCurveHasEndpoint(second, point) {
				return true
			}
		}
	}
	return false
}

func assertOccultationRiseSetNoInstantaneousBranchJumps(t *testing.T, curves []OccultationRiseSetCurve) {
	t.Helper()
	for _, curve := range curves {
		for segmentIndex, segment := range curve.Segments {
			for pointIndex := 1; pointIndex < len(segment); pointIndex++ {
				previous, current := segment[pointIndex-1], segment[pointIndex]
				distance := occultationPathDistanceKM(previous, current)
				deltaSeconds := math.Abs(current.Time.Sub(previous.Time).Seconds())
				if distance > 500 && deltaSeconds < 1 {
					t.Fatalf("%s/%s segment %d contains %.1f km jump in %.3f seconds", curve.Phase, curve.Direction, segmentIndex, distance, deltaSeconds)
				}
			}
		}
	}
}

func occultationRiseSetCurvesShareEndpointInRegion(
	first, second, third OccultationRiseSetCurve,
	minimumLongitude, maximumLongitude, minimumLatitude, maximumLatitude float64,
) bool {
	for _, segment := range first.Segments {
		for _, point := range []OccultationPathPoint{segment[0], segment[len(segment)-1]} {
			if point.Longitude < minimumLongitude || point.Longitude > maximumLongitude ||
				point.Latitude < minimumLatitude || point.Latitude > maximumLatitude {
				continue
			}
			if occultationRiseSetCurveHasEndpoint(second, point) &&
				occultationRiseSetCurveHasEndpoint(third, point) {
				return true
			}
		}
	}
	return false
}

func occultationRiseSetSegmentsShareEndpointInRegion(
	segments [][]OccultationPathPoint,
	minimumLongitude, maximumLongitude, minimumLatitude, maximumLatitude float64,
) bool {
	for firstIndex, first := range segments {
		for _, point := range []OccultationPathPoint{first[0], first[len(first)-1]} {
			if point.Longitude < minimumLongitude || point.Longitude > maximumLongitude ||
				point.Latitude < minimumLatitude || point.Latitude > maximumLatitude {
				continue
			}
			for secondIndex, second := range segments {
				if firstIndex == secondIndex {
					continue
				}
				if occultationRiseSetSegmentHasEndpoint(second, point) {
					return true
				}
			}
		}
	}
	return false
}

func occultationRiseSetCurveHasEndpoint(curve OccultationRiseSetCurve, point OccultationPathPoint) bool {
	for _, segment := range curve.Segments {
		if occultationRiseSetSegmentHasEndpoint(segment, point) {
			return true
		}
	}
	return false
}

func occultationRiseSetSegmentHasEndpoint(segment []OccultationPathPoint, point OccultationPathPoint) bool {
	for _, endpoint := range []OccultationPathPoint{segment[0], segment[len(segment)-1]} {
		if math.Abs(endpoint.Time.Sub(point.Time).Seconds()) <= 1 &&
			occultationPathDistanceKM(endpoint, point) <= 5 {
			return true
		}
	}
	return false
}

func assertOccultationRiseSetEndpointNetwork(
	t *testing.T,
	curves []OccultationRiseSetCurve,
	northern, southern []OccultationPathPoint,
) {
	t.Helper()
	type endpoint struct {
		curve, segment, side int
		point                OccultationPathPoint
	}
	var endpoints []endpoint
	for curveIndex, curve := range curves {
		for segmentIndex, segment := range curve.Segments {
			if len(segment) < 2 {
				t.Fatalf("%s/%s segment %d has fewer than two points", curve.Phase, curve.Direction, segmentIndex)
			}
			endpoints = append(endpoints,
				endpoint{curveIndex, segmentIndex, 0, segment[0]},
				endpoint{curveIndex, segmentIndex, 1, segment[len(segment)-1]},
			)
		}
	}
	for index, current := range endpoints {
		closed := false
		for otherIndex, other := range endpoints {
			if index == otherIndex {
				continue
			}
			if math.Abs(current.point.Time.Sub(other.point.Time).Seconds()) <= 1 &&
				occultationPathDistanceKM(current.point, other.point) <= 5 {
				closed = true
				break
			}
		}
		if closed {
			continue
		}
		nearestBoundary := math.Inf(1)
		for _, point := range append(append([]OccultationPathPoint(nil), northern...), southern...) {
			nearestBoundary = math.Min(nearestBoundary, occultationPathDistanceKM(current.point, point))
		}
		if nearestBoundary > 400 {
			curve := curves[current.curve]
			t.Errorf("%s/%s segment %d endpoint is %.1f km from the occultation-band boundary",
				curve.Phase, curve.Direction, current.segment, nearestBoundary)
		}
	}
}

func TestOccultationRiseSetEvaluationCacheReusesContexts(t *testing.T) {
	calls := 0
	cache := newOccultationRiseSetEvaluationCache(func(tt float64) occultationRiseSetContext {
		calls++
		return newOccultationRiseSetContext(tt, 10, 20, 384400, 30, -5, 0, 0)
	})
	const tt = 2451545.25
	cache.evaluation(tt)
	cache.evaluation(tt)
	if calls != 3 {
		t.Fatalf("same evaluation built %d contexts, want center/before/after once", calls)
	}
	cache.evaluation(tt + occultationRiseSetDerivativeStepDays)
	if calls != 4 {
		t.Fatalf("overlapping evaluation built %d contexts, want one additional context", calls)
	}
}

func TestOccultationRiseSetRegressionCurveCounts(t *testing.T) {
	polar := StarCoordinate{
		ID: "Antares", RA: 247.3516666666667, Dec: -26.431944444444444,
		Epoch: time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC), Frame: CoordinateFrameJ2000,
		ProperMotionRACosDecMasPerYear: -10, ProperMotionDecMasPerYear: -20, ParallaxMas: 24,
	}
	cases := []struct {
		name string
		find func() ([]OccultationRiseSetCurve, error)
		// wantPoints 是当前采样下的实测点数，断言取 ±40% 带宽，抓采样与落段回归。
		wantPoints int
	}{
		{
			name:       "HR4799",
			wantPoints: 591,
			find: func() ([]OccultationRiseSetCurve, error) {
				paths, err := FindStarOccultationPaths(
					time.Date(2025, 6, 5, 0, 0, 0, 0, time.UTC),
					time.Date(2025, 6, 6, 0, 0, 0, 0, time.UTC),
					hr4799OccultationCoordinateForTest(),
					OccultationPathOptions{Step: 10 * time.Minute},
				)
				if err != nil || len(paths) != 1 {
					return nil, fmt.Errorf("paths=%d err=%v", len(paths), err)
				}
				return paths[0].RiseSetCurves, nil
			},
		},
		{
			name:       "Saturn",
			wantPoints: 707,
			find: func() ([]OccultationRiseSetCurve, error) {
				paths, err := FindPlanetOccultationPaths(
					time.Date(2024, 8, 21, 0, 0, 0, 0, time.UTC),
					time.Date(2024, 8, 22, 0, 0, 0, 0, time.UTC),
					OccultationSaturn,
					OccultationPathOptions{Step: 10 * time.Minute},
				)
				if err != nil || len(paths) != 1 {
					return nil, fmt.Errorf("paths=%d err=%v", len(paths), err)
				}
				return paths[0].RiseSetCurves, nil
			},
		},
		{
			name:       "Saturn-2025-02-01",
			wantPoints: 1490,
			find: func() ([]OccultationRiseSetCurve, error) {
				start := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
				paths, err := FindPlanetOccultationPaths(
					start, start.Add(24*time.Hour), OccultationSaturn,
					OccultationPathOptions{
						Step: 20 * time.Minute, TargetSpacingKM: 900,
						RiseSetStep: 5 * time.Minute, DisableFootprints: true,
					},
				)
				if err != nil || len(paths) != 1 {
					return nil, fmt.Errorf("paths=%d err=%v", len(paths), err)
				}
				return paths[0].RiseSetCurves, nil
			},
		},
		{
			name:       "high-latitude",
			wantPoints: 1262,
			find: func() ([]OccultationRiseSetCurve, error) {
				paths, err := FindStarOccultationPaths(
					time.Date(2026, 2, 11, 0, 0, 0, 0, time.UTC),
					time.Date(2026, 2, 12, 0, 0, 0, 0, time.UTC),
					polar,
					OccultationPathOptions{Step: 10 * time.Minute},
				)
				if err != nil || len(paths) != 1 {
					return nil, fmt.Errorf("paths=%d err=%v", len(paths), err)
				}
				return paths[0].RiseSetCurves, nil
			},
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			curves, err := test.find()
			if err != nil {
				t.Fatal(err)
			}
			if len(curves) != 6 {
				t.Fatalf("rise/set curve count = %d, want 6", len(curves))
			}
			points := occultationRiseSetPointCount(curves)
			if points < test.wantPoints*3/5 || points > test.wantPoints*8/5 {
				t.Fatalf("rise/set point count = %d, want %d +/-40%%", points, test.wantPoints)
			}
			for _, curve := range curves {
				if len(curve.Segments) == 0 {
					t.Fatalf("%s/%s has no segments", curve.Phase, curve.Direction)
				}
			}
		})
	}
}

func occultationRiseSetPointCount(curves []OccultationRiseSetCurve) int {
	count := 0
	for _, curve := range curves {
		for _, segment := range curve.Segments {
			count += len(segment)
		}
	}
	return count
}

func BenchmarkOccultationRiseSetRegression(b *testing.B) {
	polar := StarCoordinate{
		ID: "Antares", RA: 247.3516666666667, Dec: -26.431944444444444,
		Epoch: time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC), Frame: CoordinateFrameJ2000,
		ProperMotionRACosDecMasPerYear: -10, ProperMotionDecMasPerYear: -20, ParallaxMas: 24,
	}
	cases := []struct {
		name string
		find func() ([]OccultationRiseSetCurve, error)
	}{
		{
			name: "HR4799",
			find: func() ([]OccultationRiseSetCurve, error) {
				paths, err := FindStarOccultationPaths(
					time.Date(2025, 6, 5, 0, 0, 0, 0, time.UTC),
					time.Date(2025, 6, 6, 0, 0, 0, 0, time.UTC),
					hr4799OccultationCoordinateForTest(), OccultationPathOptions{Step: 10 * time.Minute},
				)
				if err != nil || len(paths) != 1 {
					return nil, fmt.Errorf("paths=%d err=%v", len(paths), err)
				}
				return paths[0].RiseSetCurves, nil
			},
		},
		{
			name: "Saturn",
			find: func() ([]OccultationRiseSetCurve, error) {
				paths, err := FindPlanetOccultationPaths(
					time.Date(2024, 8, 21, 0, 0, 0, 0, time.UTC),
					time.Date(2024, 8, 22, 0, 0, 0, 0, time.UTC),
					OccultationSaturn, OccultationPathOptions{Step: 10 * time.Minute},
				)
				if err != nil || len(paths) != 1 {
					return nil, fmt.Errorf("paths=%d err=%v", len(paths), err)
				}
				return paths[0].RiseSetCurves, nil
			},
		},
		{
			name: "Saturn-2025-02-01",
			find: func() ([]OccultationRiseSetCurve, error) {
				start := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
				paths, err := FindPlanetOccultationPaths(
					start, start.Add(24*time.Hour), OccultationSaturn,
					OccultationPathOptions{
						Step: 20 * time.Minute, TargetSpacingKM: 900,
						RiseSetStep: 5 * time.Minute, DisableFootprints: true,
					},
				)
				if err != nil || len(paths) != 1 {
					return nil, fmt.Errorf("paths=%d err=%v", len(paths), err)
				}
				return paths[0].RiseSetCurves, nil
			},
		},
		{
			name: "high-latitude",
			find: func() ([]OccultationRiseSetCurve, error) {
				paths, err := FindStarOccultationPaths(
					time.Date(2026, 2, 11, 0, 0, 0, 0, time.UTC),
					time.Date(2026, 2, 12, 0, 0, 0, 0, time.UTC),
					polar, OccultationPathOptions{Step: 10 * time.Minute},
				)
				if err != nil || len(paths) != 1 {
					return nil, fmt.Errorf("paths=%d err=%v", len(paths), err)
				}
				return paths[0].RiseSetCurves, nil
			},
		},
	}
	for _, test := range cases {
		b.Run(test.name, func(b *testing.B) {
			b.ReportAllocs()
			for index := 0; index < b.N; index++ {
				curves, err := test.find()
				if err != nil || len(curves) != 6 {
					b.Fatalf("curves=%d err=%v, want six curves", len(curves), err)
				}
				b.ReportMetric(float64(occultationRiseSetPointCount(curves)), "points/op")
			}
		})
	}
}

func TestSolarEclipseRiseSetStepAllowsCoarseSampling(t *testing.T) {
	seed := JDCalc(2024, 4, 8)
	// Five- and thirty-minute inputs can both hit the spatial chord limit.
	// Use a genuinely dense input so the test measures time decimation.
	fine := SolarEclipsePartialFootprints(seed, SolarEclipsePartialFootprintOptions{
		StepDays: 10.0 / 1440.0, RiseSetStepDays: 30.0 / 86400.0,
	})
	coarse := SolarEclipsePartialFootprints(seed, SolarEclipsePartialFootprintOptions{
		StepDays: 10.0 / 1440.0, RiseSetStepDays: 30.0 / 1440.0,
	})
	if !solarEclipseRiseSetCurveTopologyComplete(fine.RiseSetCurves) ||
		!solarEclipseRiseSetCurveTopologyComplete(coarse.RiseSetCurves) {
		t.Fatal("time decimation must preserve complete rise/set topology")
	}
	if solarRiseSetPointCount(coarse.RiseSetCurves) >= solarRiseSetPointCount(fine.RiseSetCurves) {
		t.Fatal("coarser solar rise/set step did not reduce curve samples")
	}
}

func solarRiseSetPointCount(curves []SolarEclipseRiseSetCurve) int {
	count := 0
	for _, curve := range curves {
		for _, segment := range curve.Segments {
			count += len(segment)
		}
	}
	return count
}

func assertOccultationRiseSetCurves(
	t *testing.T,
	curves []OccultationRiseSetCurve,
	contextAt occultationRiseSetContextFunc,
) {
	t.Helper()
	if len(curves) != 6 {
		t.Fatalf("occultation rise/set curve count = %d, want 6", len(curves))
	}
	for _, curve := range curves {
		for _, segment := range curve.Segments {
			for index := 1; index < len(segment); index++ {
				if !segment[index].Time.After(segment[index-1].Time) {
					t.Fatalf("occultation %s/%s segment time is not increasing at %d", curve.Phase, curve.Direction, index)
				}
				if distance := occultationPathDistanceKM(segment[index-1], segment[index]); distance > occultationPathBoundarySpacingKM+1e-6 {
					t.Fatalf("occultation %s/%s segment gap = %.1f km from %v (%.4f, %.4f) to %v (%.4f, %.4f), want at most %.0f km",
						curve.Phase, curve.Direction, distance,
						segment[index-1].Time, segment[index-1].Longitude, segment[index-1].Latitude,
						segment[index].Time, segment[index].Longitude, segment[index].Latitude,
						occultationPathBoundarySpacingKM)
				}
			}
			for _, point := range riseSetOccultationTestSamples(segment) {
				tt := occultationTimeToTT(point.Time)
				evaluation := occultationRiseSetEvaluation{
					tt: tt, center: contextAt(tt),
					before: contextAt(tt - occultationRiseSetDerivativeStepDays),
					after:  contextAt(tt + occultationRiseSetDerivativeStepDays),
				}
				state := evaluation.center.stateAt(point.Longitude, point.Latitude)
				if !state.valid || math.Abs(state.moonAltitude) > 1e-6 {
					t.Fatalf("occultation %s/%s altitude = %.9f deg", curve.Phase, curve.Direction, state.moonAltitude)
				}
				contactDerivative := evaluation.contactDerivative(point.Longitude, point.Latitude)
				phaseJunction := math.Abs(state.contactMetric) <= 1e-7 &&
					math.Abs(contactDerivative) <= occultationRiseSetJunctionDerivativeTolerance
				if phaseJunction {
					if evaluation.contactSecondDerivative(point.Longitude, point.Latitude) <= 0 {
						t.Fatalf("occultation phase junction is not a contact-gap minimum at %.6f, %.6f",
							point.Longitude, point.Latitude)
					}
				} else if curve.Phase == RiseSetPhaseGreatest {
					derivative := evaluation.separationDerivative(point.Longitude, point.Latitude)
					secondDerivative := evaluation.separationSecondDerivative(point.Longitude, point.Latitude)
					if math.Abs(derivative) > 1e-8 || secondDerivative <= 0 || state.contactMetric > 1e-7 {
						t.Fatalf("occultation greatest residual is invalid at %.6f, %.6f: derivative=%.9g second=%.9g contact=%.9g",
							point.Longitude, point.Latitude, derivative, secondDerivative, state.contactMetric)
					}
				} else {
					if math.Abs(state.contactMetric) > 1e-7 {
						t.Fatalf("occultation %s contact residual = %.9g deg", curve.Phase, state.contactMetric)
					}
					if math.Abs(contactDerivative) > 1e-7 &&
						(curve.Phase == RiseSetPhaseStart && contactDerivative >= 0 || curve.Phase == RiseSetPhaseEnd && contactDerivative <= 0) {
						t.Fatalf("occultation %s contact derivative = %.9g", curve.Phase, contactDerivative)
					}
				}
				altitudeDerivative := evaluation.moonAltitudeDerivative(point.Longitude, point.Latitude)
				if math.Abs(altitudeDerivative) > 1e-7 &&
					(curve.Direction == RiseSetDirectionRise && altitudeDerivative <= 0 ||
						curve.Direction == RiseSetDirectionSet && altitudeDerivative >= 0) {
					t.Fatalf("occultation %s altitude derivative = %.9g", curve.Direction, altitudeDerivative)
				}
			}
		}
	}
	assertOccultationRiseSetEndpointsClosed(t, curves)
}

func assertOccultationRiseSetEndpointsClosed(t *testing.T, curves []OccultationRiseSetCurve) {
	t.Helper()
	type endpoint struct {
		curve, segment, side int
		point                OccultationPathPoint
	}
	var endpoints []endpoint
	for curveIndex, curve := range curves {
		for segmentIndex, segment := range curve.Segments {
			if len(segment) == 0 {
				continue
			}
			endpoints = append(endpoints,
				endpoint{curve: curveIndex, segment: segmentIndex, side: 0, point: segment[0]},
				endpoint{curve: curveIndex, segment: segmentIndex, side: 1, point: segment[len(segment)-1]},
			)
		}
	}
	for index, current := range endpoints {
		matched := false
		nearestIndex := -1
		nearestMetric := math.Inf(1)
		for otherIndex, other := range endpoints {
			if index == otherIndex || current.curve == other.curve && current.segment == other.segment && current.side == other.side {
				continue
			}
			timeDifference := math.Abs(occultationTimeToTT(current.point.Time) - occultationTimeToTT(other.point.Time))
			distance := occultationPathDistanceKM(current.point, other.point)
			metric := distance + timeDifference*86400
			if metric < nearestMetric {
				nearestIndex, nearestMetric = otherIndex, metric
			}
			if timeDifference <= 1e-8 && distance <= 0.01 {
				matched = true
				break
			}
		}
		if !matched {
			curve := curves[current.curve]
			nearest := endpoints[nearestIndex]
			nearestCurve := curves[nearest.curve]
			t.Errorf("occultation %s/%s segment %d has an unclosed endpoint at %v (%.6f, %.6f); nearest %s/%s segment %d differs by %.3fs and %.1f km",
				curve.Phase, curve.Direction, current.segment, current.point.Time,
				current.point.Longitude, current.point.Latitude,
				nearestCurve.Phase, nearestCurve.Direction, nearest.segment,
				math.Abs(current.point.Time.Sub(nearest.point.Time).Seconds()),
				occultationPathDistanceKM(current.point, nearest.point))
		}
	}
}

func assertSolarRiseSetPhase(
	t *testing.T,
	curve SolarEclipseRiseSetCurve,
	point SolarEclipsePathPoint,
	state localSolarEclipseState,
	evaluation solarEclipseRiseSetEvaluation,
) {
	t.Helper()
	contactGap := solarEclipsePartialContactGap(state)
	contactDerivative := evaluation.partialContactDerivative(point.Longitude, point.Latitude)
	phaseJunction := math.Abs(contactGap) <= 1e-7 && math.Abs(contactDerivative) <= 1e-7
	if phaseJunction {
		if evaluation.partialContactSecondDerivative(point.Longitude, point.Latitude) <= 0 {
			t.Fatalf("solar phase junction is not a contact-gap minimum at %.6f, %.6f", point.Longitude, point.Latitude)
		}
	} else if curve.Phase == RiseSetPhaseGreatest {
		if math.Abs(evaluation.separationDerivative(point.Longitude, point.Latitude)) > 1e-8 ||
			evaluation.separationSecondDerivative(point.Longitude, point.Latitude) <= 0 ||
			contactGap > 1e-7 {
			t.Fatalf("solar greatest residual is invalid at %.6f, %.6f", point.Longitude, point.Latitude)
		}
	} else {
		if math.Abs(contactGap) > 1e-7 {
			t.Fatalf("solar %s contact residual = %.9g rad", curve.Phase, contactGap)
		}
		if curve.Phase == RiseSetPhaseStart && contactDerivative >= 0 || curve.Phase == RiseSetPhaseEnd && contactDerivative <= 0 {
			t.Fatalf("solar %s contact derivative = %.9g", curve.Phase, contactDerivative)
		}
	}
	altitudeDerivative := evaluation.sunAltitudeDerivative(point.Longitude, point.Latitude)
	directionJunction := math.Abs(altitudeDerivative) <= 1e-7
	if directionJunction {
		if math.Abs(evaluation.sunAltitudeSecondDerivative(point.Longitude, point.Latitude)) <= 1e-7 {
			t.Fatalf("solar direction junction is not a horizon tangency at %.6f, %.6f", point.Longitude, point.Latitude)
		}
	} else if curve.Direction == RiseSetDirectionRise && altitudeDerivative <= 0 ||
		curve.Direction == RiseSetDirectionSet && altitudeDerivative >= 0 {
		t.Fatalf("solar %s altitude derivative = %.9g", curve.Direction, altitudeDerivative)
	}
}

func riseSetSolarTestSamples(segment []SolarEclipsePathPoint) []SolarEclipsePathPoint {
	return []SolarEclipsePathPoint{segment[0], segment[len(segment)/2], segment[len(segment)-1]}
}

func riseSetOccultationTestSamples(segment []OccultationPathPoint) []OccultationPathPoint {
	return []OccultationPathPoint{segment[0], segment[len(segment)/2], segment[len(segment)-1]}
}

// TestRiseSetCyclicRootsWithFoldToleranceCoversFoldOnlyResiduals 固定折点补根的实际契约：
// 折点处残差与零相切而不变号，符号扫描整圈为空，此时补根必须给出折点根（这是极区相位曲线
// 得以闭合的路径）；而当同一相位上同时存在普通变号根时，历史契约只返回符号根，折点根会被
// 丢弃。后者是已知取舍：实测在重建阶段无条件合并两类根会让 2025-01-05 这类极区事件的解析
// 掩带失去权威性（偏掩带回退为 footprint-sweep-fallback，可见性边界越出掩带 784 km），且在
// 去重容差 1e-4 与 1e-2 下都会退化、仅在 1e-3 下勉强通过，因此不能作为修复落地。
// TestRiseSetCyclicRootsWithFoldToleranceCoversFoldOnlyResiduals pins the actual fold-recovery
// contract: a fold holds the residual at zero without changing sign and leaves the sign scan
// empty, and recovery must then return the fold root, which is how polar phase curves close.
// When an ordinary sign-change root shares the phase, the historical contract returns only the
// sign roots and the fold root is dropped. That is a known trade-off: always merging both
// families in the rebuild pass was measured to strip the analytic band of its authority for
// polar events such as 2025-01-05 (the partial band fell back to footprint-sweep-fallback and
// the visibility boundary escaped it by 784 km), degrading at both 1e-4 and 1e-2 deduplication
// tolerances and only passing at 1e-3, so it cannot ship as a fix.
func TestRiseSetCyclicRootsWithFoldToleranceCoversFoldOnlyResiduals(t *testing.T) {
	const samples = 720
	// 1.0 处与零相切不变号：符号扫描为空，折点补根必须命中。
	// The residual tangents zero at 1.0 without changing sign, so the sign scan stays empty
	// and fold recovery must find it.
	foldOnly := func(angle float64) (float64, bool) {
		return 0.5 * riseSetAngularDistance(angle, 1.0), true
	}
	if roots := riseSetSignChangeRoots(samples, foldOnly); len(roots) != 0 {
		t.Fatalf("riseSetSignChangeRoots() = %v, want no crossing root", roots)
	}
	roots := riseSetCyclicRootsWithFoldTolerance(samples, riseSetFoldRootResidualToleranceDeg, foldOnly)
	if len(roots) != 1 || riseSetAngularDistance(roots[0], 1.0) > 1e-3 {
		t.Fatalf("riseSetCyclicRootsWithFoldTolerance() = %v, want the single fold root near 1", roots)
	}
	// 折点与普通变号根共存：只返回符号根，折点根按现状丢弃。
	// A fold coexisting with ordinary crossing roots: only the sign roots are returned and the
	// fold root is dropped, matching current behaviour.
	mixed := func(angle float64) (float64, bool) {
		if riseSetAngularDistance(angle, 4.0) < 1.2 {
			return riseSetAngularDistance(angle, 4.0) - 0.2, true
		}
		return 0.5 * riseSetAngularDistance(angle, 1.0), true
	}
	signRoots := riseSetSignChangeRoots(samples, mixed)
	if len(signRoots) != 2 {
		t.Fatalf("riseSetSignChangeRoots() = %v, want the two crossing roots", signRoots)
	}
	roots = riseSetCyclicRootsWithFoldTolerance(samples, riseSetFoldRootResidualToleranceDeg, mixed)
	if len(roots) != len(signRoots) {
		t.Fatalf("riseSetCyclicRootsWithFoldTolerance() = %v, want only the %d sign roots",
			roots, len(signRoots))
	}
	for index, root := range roots {
		if riseSetAngularDistance(root, signRoots[index]) > 1e-6 {
			t.Fatalf("riseSetCyclicRootsWithFoldTolerance() = %v, want the sign roots %v", roots, signRoots)
		}
	}
}
