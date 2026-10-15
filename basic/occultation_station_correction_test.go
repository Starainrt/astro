package basic

import (
	"fmt"
	"math"
	"testing"
	"time"
)

func TestOccultationStationEnvelopeEqualTimeBranches(t *testing.T) {
	// 同时间戳平台是 JD 量化的产物：平台内的点与边都必须保留，只有真折返才切段。
	for _, test := range []struct {
		name     string
		times    []int
		unit     time.Duration
		segments int
	}{
		{"leading plateau", []int{0, 0, 0, 1, 2}, time.Second, 1},
		{"interior plateau", []int{0, 1, 1, 1, 2}, time.Second, 1},
		{"trailing plateau", []int{0, 1, 1, 1}, time.Second, 1},
		{"entire plateau", []int{0, 0, 0}, time.Second, 0},
		{"reverse branches", []int{2, 1, 1, 1, 0}, time.Second, 1},
		{"ordinary fold", []int{0, 1, 2, 1, 0}, time.Second, 2},
		// Saturn 2025-01-05：时间折返的顶点恰好落进同一量化格。
		{"plateau at fold", []int{0, 1, 2, 2, 1, 0}, time.Second, 2},
		// 平台出口只晚 1.5 µs：亚量化偏移不能越过出口点。
		{"tight plateau exit", []int{0, 4, 4, 4, 5}, 375 * time.Nanosecond, 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			points := make([]OccultationPathPoint, len(test.times))
			for index, value := range test.times {
				points[index] = OccultationPathPoint{
					Time: time.Unix(1000, 0).Add(time.Duration(value) * test.unit), Longitude: float64(index),
				}
			}
			segments := occultationStationSplitEnvelopeAtTimeFolds(points)
			if len(segments) != test.segments {
				t.Fatalf("segments=%d, want %d", len(segments), test.segments)
			}
			if test.segments == 0 {
				return
			}
			kept := make([]bool, len(points))
			edges := map[[2]int]bool{}
			for _, segment := range segments {
				for index, point := range segment {
					// 平台点只允许亚量化（< 1 ms）的时间调整，位置不变。
					original := points[int(point.Longitude)]
					if delta := point.Time.Sub(original.Time); delta < -time.Millisecond || delta > time.Millisecond {
						t.Fatalf("point %v moved by %v", point.Longitude, delta)
					}
					kept[int(point.Longitude)] = true
					if index > 0 {
						if !point.Time.After(segment[index-1].Time) {
							t.Fatalf("non-increasing segment: %+v", segment)
						}
						first, second := int(segment[index-1].Longitude), int(point.Longitude)
						if first > second {
							first, second = second, first
						}
						edges[[2]int{first, second}] = true
					}
				}
			}
			for index := range points {
				if !kept[index] {
					t.Fatalf("point %d dropped: %+v", index, segments)
				}
				if index > 0 && !edges[[2]int{index - 1, index}] {
					t.Fatalf("lost edge %d-%d: %+v", index-1, index, segments)
				}
			}
		})
	}
}

func TestOccultationStationOracleRefinesPlanetContours(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	cases := []struct {
		name   string
		start  time.Time
		planet OccultationPlanet
	}{
		{name: "Mars-20250729", start: time.Date(2025, time.July, 29, 0, 0, 0, 0, zone), planet: OccultationMars},
		{name: "Saturn-20240725", start: time.Date(2024, time.July, 25, 0, 0, 0, 0, zone), planet: OccultationSaturn},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			paths, err := FindPlanetOccultationPaths(test.start, test.start.Add(24*time.Hour), test.planet, OccultationPathOptions{
				Step: 20 * time.Minute, TargetSpacingKM: 900, DisableFootprints: true, DisableRiseSet: true,
			})
			if err != nil || len(paths) != 1 {
				t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
			}
			config, ok := planetOccultationConfigFor(test.planet)
			if !ok {
				t.Fatalf("%s config unavailable", test.planet)
			}
			cache := newPlanetOccultationEventCache(config)
			cache.prepareLocalEphemeris(occultationTimeToTT(paths[0].Greatest.Time))
			checked := 0
			maxOffset := 0.0
			maxSeedResidual := 0.0
			unsolved := 0
			for _, contourSet := range []struct {
				name     string
				contours [][]OccultationPathPoint
				frameAt  occultationPathFrameFunc
				total    bool
			}{
				{name: "partial", contours: occultationContactBandContoursWithAdditionalTimes(
					paths[0].Start, paths[0].End, centerTimeTT(paths[0].Start.Time), centerTimeTT(paths[0].End.Time),
					centerTimeTT(paths[0].Greatest.Time), cache.outerFrameAt,
					OccultationPathOptions{Step: 20 * time.Minute, DisableRiseSet: true}, time.UTC, nil,
				), frameAt: cache.outerFrameAt},
				{name: "total", contours: occultationContactBandContoursWithAdditionalTimes(
					paths[0].TotalStart, paths[0].TotalEnd, centerTimeTT(paths[0].TotalStart.Time), centerTimeTT(paths[0].TotalEnd.Time),
					centerTimeTT(paths[0].Greatest.Time), cache.totalFrameAt,
					OccultationPathOptions{Step: 20 * time.Minute, DisableRiseSet: true}, time.UTC, nil,
				), frameAt: cache.totalFrameAt, total: true},
			} {
				for _, contour := range contourSet.contours {
					if len(contour) == 0 {
						continue
					}
					stride := int(math.Max(1, math.Ceil(float64(len(contour))/12)))
					for index := stride; index+stride < len(contour); index += stride {
						seed := contour[index]
						sample, solved := occultationStationCorrectBoundaryPoint(
							centerTimeTT(seed.Time), seed, contourSet.frameAt, cache.riseSetContextAt,
							contourSet.total, time.UTC,
						)
						if !solved || !sample.valid {
							unsolved++
							continue
						}
						if math.Abs(sample.contactResidualDeg) > 1e-5 {
							t.Fatalf("%s contour sample %d contact residual=%.9g arcsec", contourSet.name, index, sample.contactResidualDeg)
						}
						maxOffset = math.Max(maxOffset, math.Abs(sample.offsetKM))
						maxSeedResidual = math.Max(maxSeedResidual, math.Abs(sample.seedResidualDeg))
						checked++
					}
				}
			}
			if checked < 8 || unsolved > checked {
				t.Fatalf("checked %d station contour samples, unsolved=%d", checked, unsolved)
			}
			t.Logf("station oracle samples=%d unsolved=%d max offset=%.1f km max geocentric residual=%.6f arcsec", checked, unsolved, maxOffset, maxSeedResidual)
			if maxOffset > occultationStationOracleMaximumOffsetKM {
				t.Fatalf("max station correction offset=%.1f km exceeds oracle bound", maxOffset)
			}
		})
	}
}

func TestOccultationStationCorrectedOpenFootprintEndpointsStayOnHorizon(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2025, time.July, 29, 0, 0, 0, 0, zone)
	paths, err := FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), OccultationMars,
		OccultationPathOptions{
			Step: 20 * time.Minute, TargetSpacingKM: 900,
			RiseSetStep: time.Minute, DisableFootprints: true,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	config, _ := planetOccultationConfigFor(OccultationMars)
	cache := newPlanetOccultationEventCache(config)
	cache.preparePathEphemeris(occultationTimeToTT(paths[0].Greatest.Time), OccultationPathAlgorithmOptimized)
	horizonEndpoints, coneEndpoints := 0, 0
	for _, band := range []struct {
		name       string
		frameAt    occultationPathFrameFunc
		footprints []PlanetOccultationFootprint
	}{
		{name: "partial", frameAt: cache.outerFrameAt, footprints: paths[0].PartialBandFootprints},
		{name: "total", frameAt: cache.totalFrameAt, footprints: paths[0].TotalBandFootprints},
	} {
		for footprintIndex, footprint := range band.footprints {
			if footprint.Closed {
				continue
			}
			geocentric, geocentricOK := planetOccultationFootprintAtWithResolution(
				occultationTimeToTT(footprint.Time), band.frameAt, zone,
				planetOccultationBandBoundaryPoints, planetOccultationBandHorizonPoints,
				planetOccultationBandTargetSpacingKM,
			)
			if !geocentricOK {
				t.Fatalf("%s footprint %d has no geocentric support at %v", band.name, footprintIndex, footprint.Time)
			}
			for boundaryIndex, boundary := range footprint.Boundaries {
				if len(boundary) < 2 {
					continue
				}
				for _, pointIndex := range []int{0, len(boundary) - 1} {
					point := boundary[pointIndex]
					if math.Abs(point.MoonAltitude) <= 1e-5 {
						horizonEndpoints++
						continue
					}
					// 接触锥离开椭球处的尖点端点在地平线以上，地平线求解无法也
					// 不应移动它；此时端点必须与同一时刻的地心边界端点重合。
					if !planetOccultationFootprintHasGeocentricConeEndpoint(geocentric, point) {
						t.Fatalf(
							"%s footprint %d boundary %d endpoint %d MoonAltitude=%.9f deg is neither on the horizon nor a geocentric cone endpoint",
							band.name, footprintIndex, boundaryIndex, pointIndex, point.MoonAltitude,
						)
					}
					coneEndpoints++
				}
			}
		}
	}
	if horizonEndpoints < 20 {
		t.Fatalf("checked only %d horizon endpoints", horizonEndpoints)
	}
	if coneEndpoints == 0 {
		t.Fatal("no cone-cusp endpoint observed; the cone-edge branch is untested")
	}
}

func planetOccultationFootprintHasGeocentricConeEndpoint(
	footprint PlanetOccultationFootprint,
	point OccultationPathPoint,
) bool {
	for _, boundary := range footprint.Boundaries {
		if len(boundary) < 2 {
			continue
		}
		for _, candidate := range []OccultationPathPoint{boundary[0], boundary[len(boundary)-1]} {
			// 锥体尖点端点在月球地平线以上，因此用“离开地平线”筛选地心对应点。
			if math.Abs(candidate.MoonAltitude) <= 1e-5 {
				continue
			}
			if occultationPathDistanceKM(candidate, point) <= 1 {
				return true
			}
		}
	}
	return false
}

func TestOccultationStationCorrectedContoursStayOnTemporalEnvelope(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2025, time.July, 29, 0, 0, 0, 0, zone)
	paths, err := FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), OccultationMars,
		OccultationPathOptions{
			Step: 20 * time.Minute, TargetSpacingKM: 900,
			RiseSetStep: time.Minute, DisableFootprints: true,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	config, ok := planetOccultationConfigFor(OccultationMars)
	if !ok {
		t.Fatal("Mars config unavailable")
	}
	cache := newPlanetOccultationEventCache(config)
	cache.prepareLocalEphemeris(occultationTimeToTT(paths[0].Greatest.Time))

	checked := 0
	failed := 0
	firstFailure := ""
	for _, band := range []struct {
		name     string
		contours [][]OccultationPathPoint
		total    bool
	}{
		{name: "partial", contours: paths[0].PartialBandContours},
		{name: "total", contours: paths[0].TotalBandContours, total: true},
	} {
		for contourIndex, contour := range band.contours {
			for pointIndex, point := range contour {
				tt := centerTimeTT(point.Time)
				contextAt := cache.riseSetContextAt
				if band.total {
					contextAt = cache.totalRiseSetContextAt
				}
				evaluation := occultationRiseSetEvaluation{
					tt:     tt,
					center: contextAt(tt),
					before: contextAt(tt - occultationRiseSetDerivativeStepDays),
					after:  contextAt(tt + occultationRiseSetDerivativeStepDays),
				}
				state := evaluation.center.stateAt(point.Longitude, point.Latitude)
				derivative := evaluation.contactDerivative(point.Longitude, point.Latitude)
				if !state.valid || !finite(derivative) {
					t.Fatalf("%s contour %d point %d has invalid station state", band.name, contourIndex, pointIndex)
				}
				if math.Abs(state.contactMetric) > 1e-5 || math.Abs(derivative) > occultationRiseSetJunctionDerivativeTolerance {
					failed++
					if firstFailure == "" {
						firstFailure = fmt.Sprintf(
							"%s contour %d point %d contact=%.9g arcsec derivative=%.9g arcsec/day",
							band.name, contourIndex, pointIndex, state.contactMetric, derivative,
						)
					}
				}
				checked++
			}
		}
	}
	if checked < 100 {
		t.Fatalf("checked only %d station envelope points", checked)
	}
	if failed > 0 {
		t.Fatalf("%d/%d points miss the temporal envelope; first: %s", failed, checked, firstFailure)
	}
}

func TestOccultationStationCorrectedContoursCloseOnVisiblePhaseJunctions(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2025, time.July, 29, 0, 0, 0, 0, zone)
	paths, err := FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), OccultationMars,
		OccultationPathOptions{
			Step: 20 * time.Minute, TargetSpacingKM: 900,
			RiseSetStep: time.Minute, DisableFootprints: true,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}

	for _, band := range []struct {
		name     string
		contours [][]OccultationPathPoint
		curves   []OccultationRiseSetCurve
	}{
		{name: "partial", contours: paths[0].PartialBandContours, curves: paths[0].RiseSetCurves},
		{name: "total", contours: paths[0].TotalBandContours, curves: paths[0].TotalRiseSetCurves},
	} {
		phasePoints := make([]OccultationPathPoint, 0)
		for _, curve := range band.curves {
			if curve.Phase != RiseSetPhaseStart && curve.Phase != RiseSetPhaseEnd {
				continue
			}
			for _, segment := range curve.Segments {
				phasePoints = append(phasePoints, segment...)
			}
		}
		if len(band.contours) == 0 || len(phasePoints) == 0 {
			t.Fatalf("%s has contours=%d phase points=%d, want visible envelope and phase boundary", band.name, len(band.contours), len(phasePoints))
		}
		for contourIndex, contour := range band.contours {
			if len(contour) < 2 {
				t.Fatalf("%s contour %d has %d points", band.name, contourIndex, len(contour))
			}
			for pointIndex, point := range contour {
				if point.MoonAltitude < -1e-7 {
					t.Fatalf("%s contour %d point %d is below the lunar horizon: altitude=%.9g", band.name, contourIndex, pointIndex, point.MoonAltitude)
				}
				if pointIndex > 0 && !point.Time.After(contour[pointIndex-1].Time) {
					t.Fatalf("%s contour %d times are not strictly increasing at point %d", band.name, contourIndex, pointIndex)
				}
			}
			for _, endpoint := range []OccultationPathPoint{contour[0], contour[len(contour)-1]} {
				if math.Abs(endpoint.MoonAltitude) <= 1e-6 {
					nearestKM := math.Inf(1)
					for _, phasePoint := range phasePoints {
						nearestKM = math.Min(nearestKM, occultationPathDistanceKM(endpoint, phasePoint))
					}
					if nearestKM > 0.1 {
						t.Fatalf("%s contour %d horizon endpoint is %.3f km from the start/end phase line, want <=0.1 km", band.name, contourIndex, nearestKM)
					}
					continue
				}
				shared := false
				for otherIndex, other := range band.contours {
					if otherIndex == contourIndex || len(other) < 2 {
						continue
					}
					for _, otherEndpoint := range []OccultationPathPoint{other[0], other[len(other)-1]} {
						if math.Abs(endpoint.Time.Sub(otherEndpoint.Time).Seconds()) <= 0.01 &&
							occultationPathDistanceKM(endpoint, otherEndpoint) <= 0.001 {
							shared = true
						}
					}
				}
				if !shared {
					t.Fatalf("%s contour %d non-horizon endpoint is not a shared temporal fold", band.name, contourIndex)
				}
			}
		}
	}
}

func TestOccultationStationVisibilityContoursStayOnActiveTemporalMaximum(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2025, time.January, 5, 0, 0, 0, 0, zone)
	paths, err := FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), OccultationSaturn,
		OccultationPathOptions{
			Step: 20 * time.Minute, TargetSpacingKM: 900,
			RiseSetStep: time.Minute, DisableFootprints: true,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	config, ok := planetOccultationConfigFor(OccultationSaturn)
	if !ok {
		t.Fatal("Saturn config unavailable")
	}
	cache := newPlanetOccultationEventCache(config)
	cache.prepareLocalEphemeris(occultationTimeToTT(paths[0].Greatest.Time))
	for _, band := range []struct {
		name     string
		contours [][]OccultationPathPoint
		total    bool
	}{
		{name: "partial", contours: paths[0].PartialVisibilityContours},
		{name: "total", contours: paths[0].TotalVisibilityContours, total: true},
	} {
		if len(band.contours) == 0 {
			t.Fatalf("%s has no lunar-visibility temporal contour", band.name)
		}
		for contourIndex, contour := range band.contours {
			if len(contour) < 3 {
				t.Fatalf("%s contour %d has %d points", band.name, contourIndex, len(contour))
			}
			for pointIndex, point := range contour {
				tt := centerTimeTT(point.Time)
				context := cache.riseSetContextAt
				if band.total {
					context = func(value float64) occultationRiseSetContext {
						return cache.riseSetContextAt(value).withInternalContact()
					}
				}
				evaluation := occultationRiseSetEvaluation{
					tt: tt, center: context(tt),
					before: context(tt - occultationRiseSetDerivativeStepDays),
					after:  context(tt + occultationRiseSetDerivativeStepDays),
				}
				state := evaluation.center.stateAt(point.Longitude, point.Latitude)
				altitudeDerivative := evaluation.moonAltitudeDerivative(point.Longitude, point.Latitude)
				altitudeSecondDerivative := evaluation.moonAltitudeSecondDerivative(point.Longitude, point.Latitude)
				if !state.valid || math.Abs(state.moonAltitude) > 1e-5 ||
					math.Abs(altitudeDerivative) > occultationRiseSetJunctionDerivativeTolerance ||
					altitudeSecondDerivative >= 0 || state.contactMetric > 1e-5 {
					t.Fatalf(
						"%s contour %d point %d residuals H=%.9g Ht=%.9g Htt=%.9g F=%.9g",
						band.name, contourIndex, pointIndex, state.moonAltitude,
						altitudeDerivative, altitudeSecondDerivative, state.contactMetric,
					)
				}
			}
			for _, endpoint := range []OccultationPathPoint{contour[0], contour[len(contour)-1]} {
				context := cache.riseSetContextAt(centerTimeTT(endpoint.Time))
				if band.total {
					context = context.withInternalContact()
				}
				state := context.stateAt(endpoint.Longitude, endpoint.Latitude)
				if !state.valid || math.Abs(state.contactMetric) > 1e-5 {
					t.Fatalf("%s contour %d endpoint contact residual=%.9g", band.name, contourIndex, state.contactMetric)
				}
			}
		}
	}
}
