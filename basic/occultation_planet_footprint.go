package basic

import (
	"math"
	"sort"
	"time"
)

const (
	planetOccultationFootprintBoundaryPoints = 180
	planetOccultationHorizonPoints           = 360
	planetOccultationTimelineBoundaryPoints  = 180
	planetOccultationTimelineHorizonPoints   = 180
	planetOccultationTimelineTargetSpacingKM = 300.0
	planetOccultationFootprintMaxSamples     = 360
	planetOccultationFootprintTargetStep     = time.Minute
	planetOccultationBandBoundaryPoints      = 360
	planetOccultationBandHorizonPoints       = 360
	planetOccultationBandMaxSamples          = 128
	planetOccultationBandTargetStep          = 5 * time.Minute
	planetOccultationBandTargetSpacingKM     = 50.0
	planetOccultationBandAdaptiveMaxDepth    = 12
	planetOccultationCenterCapRadiusKM       = 60.0
	planetOccultationCenterCapPoints         = 16
	// 二分到 1e-9 度（约 0.1 毫米）即停，保证插入的切点位于地平线内侧。
	planetOccultationFootprintHorizonToleranceDeg = 1e-9
	planetOccultationFootprintHorizonSearchSteps  = 48
)

type planetOccultationFootprintSample struct {
	point OccultationPathPoint
	theta float64
	ok    bool
}

func planetOccultationFootprints(
	startTT, endTT, greatestTT float64,
	frameAt occultationPathFrameFunc,
	options OccultationPathOptions,
	location *time.Location,
) []PlanetOccultationFootprint {
	return planetOccultationFootprintsWithSampling(
		startTT, endTT, greatestTT, frameAt, location,
		planetOccultationFootprintSampleStepDays(options), planetOccultationFootprintMaxSamples,
		planetOccultationFootprintBoundaryPoints, planetOccultationHorizonPoints, 0, false,
	)
}

func planetOccultationTimelineFootprints(
	startTT, endTT, greatestTT float64,
	frameAt occultationPathFrameFunc,
	options OccultationPathOptions,
	location *time.Location,
) []PlanetOccultationFootprint {
	return planetOccultationFootprintsWithSampling(
		startTT, endTT, greatestTT, frameAt, location,
		planetOccultationFootprintSampleStepDays(options), planetOccultationFootprintMaxSamples,
		planetOccultationTimelineBoundaryPoints, planetOccultationTimelineHorizonPoints,
		planetOccultationTimelineTargetSpacingKM, false,
	)
}

func planetOccultationBandFootprints(
	startTT, endTT, greatestTT float64,
	frameAt occultationPathFrameFunc,
	location *time.Location,
	additionalTimes []float64,
) []PlanetOccultationFootprint {
	times := planetOccultationBandSampleTimesWithAdditionalTimes(
		startTT, endTT, greatestTT, additionalTimes,
	)
	return planetOccultationFootprintsAtTimes(
		times, frameAt, location,
		planetOccultationBandBoundaryPoints, planetOccultationBandHorizonPoints,
		planetOccultationBandTargetSpacingKM, 150,
	)
}

func planetOccultationBandFootprintsWithContourTimes(
	startTT, endTT, greatestTT float64,
	frameAt occultationPathFrameFunc,
	location *time.Location,
	additionalTimes []float64,
) ([]PlanetOccultationFootprint, []float64) {
	contourTimes := occultationAppendSampleTimes(nil, additionalTimes...)
	footprints := planetOccultationBandFootprints(
		startTT, endTT, greatestTT, frameAt, location, contourTimes,
	)
	transitionTimes := occultationFootprintTransitionTimes(footprints)
	if len(transitionTimes) == 0 {
		return footprints, contourTimes
	}
	contourTimes = occultationAppendSampleTimes(contourTimes, transitionTimes...)
	// The first pass already contains the complete five-minute grid and all
	// rise/set contour times. The second pass used to recompute every one of
	// those expensive finite-disk footprints just to add the handful of
	// visibility-transition instants. Evaluate only those new instants, merge
	// them into the existing grid, then run the same endpoint refinement.
	transitionFootprints := planetOccultationFootprintsAtTimes(
		transitionTimes, frameAt, location,
		planetOccultationBandBoundaryPoints, planetOccultationBandHorizonPoints,
		planetOccultationBandTargetSpacingKM, 150,
	)
	footprints = mergePlanetOccultationFootprintsByExactTime(
		footprints, transitionFootprints,
	)
	footprints = refinePlanetOccultationFootprintTransitions(
		footprints, frameAt, location,
		planetOccultationBandBoundaryPoints, planetOccultationBandHorizonPoints,
		planetOccultationBandTargetSpacingKM, 150,
	)
	return footprints, contourTimes
}

func mergePlanetOccultationFootprintsByExactTime(
	base, extra []PlanetOccultationFootprint,
) []PlanetOccultationFootprint {
	if len(extra) == 0 {
		return base
	}
	result := make([]PlanetOccultationFootprint, 0, len(base)+len(extra))
	result = append(result, base...)
	result = append(result, extra...)
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Time.Before(result[j].Time)
	})
	const tolerance = time.Second
	for index := 1; index < len(result); {
		if result[index].Closed == result[index-1].Closed &&
			result[index].Time.Sub(result[index-1].Time) <= tolerance {
			result = append(result[:index], result[index+1:]...)
			continue
		}
		index++
	}
	return result
}

func planetOccultationBandSampleTimesWithAdditionalTimes(
	startTT, endTT, greatestTT float64,
	additionalTimes []float64,
) []float64 {
	times := planetOccultationBandSampleTimes(
		startTT, endTT, greatestTT, planetOccultationBandSampleStepDays(), planetOccultationBandMaxSamples,
	)
	for _, tt := range additionalTimes {
		if tt >= startTT && tt <= endTT && finite(tt) {
			times = append(times, tt)
		}
	}
	sort.Float64s(times)
	return uniqueOccultationPathTimes(times)
}

func planetOccultationBandSampleStepDays() float64 {
	return float64(planetOccultationBandTargetStep) / float64(24*time.Hour)
}

func planetOccultationFootprintsWithSampling(
	startTT, endTT, greatestTT float64,
	frameAt occultationPathFrameFunc,
	location *time.Location,
	stepDays float64,
	maxSamples, boundaryPoints, horizonPoints int,
	targetSpacingKM float64,
	refineContacts bool,
) []PlanetOccultationFootprint {
	times := occultationPathSampleTimesWithLimit(startTT, endTT, greatestTT, stepDays, maxSamples)
	if refineContacts {
		times = planetOccultationBandSampleTimes(startTT, endTT, greatestTT, stepDays, maxSamples)
	}
	return planetOccultationFootprintsAtTimes(
		times, frameAt, location, boundaryPoints, horizonPoints, targetSpacingKM, 0,
	)
}

func planetOccultationFootprintsAtTimes(
	times []float64,
	frameAt occultationPathFrameFunc,
	location *time.Location,
	boundaryPoints, horizonPoints int,
	targetSpacingKM, maximumEndpointStepKM float64,
) []PlanetOccultationFootprint {
	footprints := make([]PlanetOccultationFootprint, 0, len(times))
	for _, tt := range times {
		footprint, ok := planetOccultationFootprintAtWithResolution(
			tt, frameAt, location, boundaryPoints, horizonPoints, targetSpacingKM,
		)
		if ok {
			footprints = append(footprints, footprint)
		}
	}
	return refinePlanetOccultationFootprintTransitions(
		footprints, frameAt, location, boundaryPoints, horizonPoints, targetSpacingKM, maximumEndpointStepKM,
	)
}

func refinePlanetOccultationFootprintTransitions(
	footprints []PlanetOccultationFootprint,
	frameAt occultationPathFrameFunc,
	location *time.Location,
	boundaryPoints, horizonPoints int,
	targetSpacingKM float64,
	maximumEndpointStepKM float64,
) []PlanetOccultationFootprint {
	if len(footprints) < 2 || maximumEndpointStepKM <= 0 {
		return footprints
	}
	const transitionTolerance = 100 * time.Millisecond
	result := make([]PlanetOccultationFootprint, 0, len(footprints)+4)
	result = append(result, footprints[0])
	for index := 1; index < len(footprints); index++ {
		left, right := footprints[index-1], footprints[index]
		if left.Closed != right.Closed && right.Time.Sub(left.Time) > transitionTolerance {
			leftTT := occultationTimeToTT(left.Time)
			rightTT := occultationTimeToTT(right.Time)
			leftClosed := left.Closed
			for rightTT-leftTT > float64(transitionTolerance)/float64(24*time.Hour) {
				middleTT := (leftTT + rightTT) / 2
				middle, ok := planetOccultationFootprintAtWithResolution(
					middleTT, frameAt, location, boundaryPoints, horizonPoints, targetSpacingKM,
				)
				if !ok {
					break
				}
				if middle.Closed == leftClosed {
					leftTT = middleTT
					left = middle
				} else {
					rightTT = middleTT
					right = middle
				}
			}
			result = appendPlanetOccultationFootprintEndpointRefinement(
				result, left, frameAt, location,
				boundaryPoints, horizonPoints, targetSpacingKM, maximumEndpointStepKM, 0,
			)
			if right.Time.After(result[len(result)-1].Time) {
				result = append(result, right)
			}
			result = appendPlanetOccultationFootprintEndpointRefinement(
				result, footprints[index], frameAt, location,
				boundaryPoints, horizonPoints, targetSpacingKM, maximumEndpointStepKM, 0,
			)
			continue
		}
		result = appendPlanetOccultationFootprintEndpointRefinement(
			result, footprints[index], frameAt, location,
			boundaryPoints, horizonPoints, targetSpacingKM, maximumEndpointStepKM, 0,
		)
	}
	return deduplicateClosedPlanetOccultationFootprints(result, time.Second)
}

func deduplicateClosedPlanetOccultationFootprints(
	footprints []PlanetOccultationFootprint,
	tolerance time.Duration,
) []PlanetOccultationFootprint {
	if len(footprints) < 2 || tolerance <= 0 {
		return footprints
	}
	result := footprints[:1]
	for _, footprint := range footprints[1:] {
		previous := result[len(result)-1]
		if previous.Closed && footprint.Closed && footprint.Time.Sub(previous.Time) <= tolerance {
			continue
		}
		result = append(result, footprint)
	}
	return result
}

func appendPlanetOccultationFootprintEndpointRefinement(
	result []PlanetOccultationFootprint,
	right PlanetOccultationFootprint,
	frameAt occultationPathFrameFunc,
	location *time.Location,
	boundaryPoints, horizonPoints int,
	targetSpacingKM, maximumEndpointStepKM float64,
	depth int,
) []PlanetOccultationFootprint {
	left := result[len(result)-1]
	if left.Closed || maximumEndpointStepKM <= 0 || depth >= 8 ||
		(left.Closed == right.Closed && occultationFootprintEndpointStepKM(left, right) <= maximumEndpointStepKM) {
		if right.Time.After(left.Time) {
			return append(result, right)
		}
		return result
	}
	middleTT := (occultationTimeToTT(left.Time) + occultationTimeToTT(right.Time)) / 2
	middle, ok := planetOccultationFootprintAtWithResolution(
		middleTT, frameAt, location, boundaryPoints, horizonPoints, targetSpacingKM,
	)
	if !ok || middle.Closed != left.Closed || !middle.Time.After(left.Time) || !right.Time.After(middle.Time) {
		return append(result, right)
	}
	result = appendPlanetOccultationFootprintEndpointRefinement(
		result, middle, frameAt, location, boundaryPoints, horizonPoints,
		targetSpacingKM, maximumEndpointStepKM, depth+1,
	)
	return appendPlanetOccultationFootprintEndpointRefinement(
		result, right, frameAt, location, boundaryPoints, horizonPoints,
		targetSpacingKM, maximumEndpointStepKM, depth+1,
	)
}

func occultationFootprintEndpointStepKM(first, second PlanetOccultationFootprint) float64 {
	if len(first.Boundaries) != 1 || len(second.Boundaries) != 1 ||
		len(first.Boundaries[0]) < 2 || len(second.Boundaries[0]) < 2 {
		return math.Inf(1)
	}
	firstBoundary := first.Boundaries[0]
	secondBoundary := second.Boundaries[0]
	keep := math.Max(
		occultationPathDistanceKM(firstBoundary[0], secondBoundary[0]),
		occultationPathDistanceKM(firstBoundary[len(firstBoundary)-1], secondBoundary[len(secondBoundary)-1]),
	)
	reverse := math.Max(
		occultationPathDistanceKM(firstBoundary[0], secondBoundary[len(secondBoundary)-1]),
		occultationPathDistanceKM(firstBoundary[len(firstBoundary)-1], secondBoundary[0]),
	)
	return math.Min(keep, reverse)
}

func occultationRiseSetEndpointTimes(curves []OccultationRiseSetCurve) []float64 {
	times := make([]float64, 0, 2*len(curves))
	for _, curve := range curves {
		for _, segment := range curve.Segments {
			if len(segment) < 2 {
				continue
			}
			times = append(times,
				occultationTimeToTT(segment[0].Time),
				occultationTimeToTT(segment[len(segment)-1].Time),
			)
		}
	}
	sort.Float64s(times)
	const tolerance = float64(time.Second) / float64(24*time.Hour)
	result := times[:0]
	counts := make([]int, 0, len(times))
	for _, tt := range times {
		if len(result) == 0 || tt-result[len(result)-1] > tolerance {
			result = append(result, tt)
			counts = append(counts, 1)
			continue
		}
		index := len(result) - 1
		counts[index]++
		result[index] += (tt - result[index]) / float64(counts[index])
	}
	return result
}

func occultationFootprintTransitionTimes(footprints []PlanetOccultationFootprint) []float64 {
	if len(footprints) < 2 {
		return nil
	}
	times := make([]float64, 0, 2)
	for index := 1; index < len(footprints); index++ {
		previous, current := footprints[index-1], footprints[index]
		if previous.Closed == current.Closed {
			continue
		}
		times = append(times, (occultationTimeToTT(previous.Time)+occultationTimeToTT(current.Time))/2)
	}
	sort.Float64s(times)
	return uniqueOccultationPathTimes(times)
}

func occultationAppendSampleTimes(base []float64, extras ...float64) []float64 {
	if len(extras) == 0 {
		return append([]float64(nil), base...)
	}
	times := append(append([]float64(nil), base...), extras...)
	sort.Float64s(times)
	return uniqueOccultationPathTimes(times)
}

func planetOccultationBandSampleTimes(
	startTT, endTT, greatestTT, stepDays float64,
	maxSamples int,
) []float64 {
	contactOffsets := [...]float64{
		1.0 / 60, 1.0 / 20, 0.1, 0.2, 0.4, 0.7, 1.2, 2, 3, 3.5, 4, 4.5,
	}
	contactSamples := len(contactOffsets)
	baseLimit := maxSamples - 2*contactSamples
	if baseLimit < 3 {
		baseLimit = 3
	}
	times := occultationPathSampleTimesWithLimit(startTT, endTT, greatestTT, stepDays, baseLimit)
	duration := endTT - startTT
	if duration <= 0 {
		return times
	}
	contactStep := math.Min(stepDays, duration/2)
	for _, multiplier := range contactOffsets {
		offset := math.Min(contactStep*multiplier, duration/2)
		times = append(times, startTT+offset, endTT-offset)
	}
	sort.Float64s(times)
	return uniqueOccultationPathTimes(times)
}

func planetOccultationFootprintSampleStepDays(options OccultationPathOptions) float64 {
	if options.IncludeFootprintTimeline {
		step := options.FootprintTimelineStep
		if step <= 0 {
			step = 5 * time.Minute
		}
		return float64(step) / float64(24*time.Hour)
	}
	step := options.Step
	if step <= 0 || step > planetOccultationFootprintTargetStep {
		step = planetOccultationFootprintTargetStep
	}
	return float64(step) / float64(24*time.Hour)
}

func planetOccultationFootprintAt(
	tt float64,
	frameAt occultationPathFrameFunc,
	location *time.Location,
) (PlanetOccultationFootprint, bool) {
	return planetOccultationFootprintAtWithResolution(
		tt, frameAt, location,
		planetOccultationFootprintBoundaryPoints, planetOccultationHorizonPoints, 0,
	)
}

func planetOccultationFootprintAtWithResolution(
	tt float64,
	frameAt occultationPathFrameFunc,
	location *time.Location,
	boundaryPoints, horizonPoints int,
	targetSpacingKM float64,
) (PlanetOccultationFootprint, bool) {
	frame, ok := frameAt(tt)
	if !ok {
		return PlanetOccultationFootprint{}, false
	}
	siderealDegrees := ApparentSiderealTime(TT2UT1(tt)) * 15
	samples := make([]planetOccultationFootprintSample, boundaryPoints)
	for index := range samples {
		theta := 2 * math.Pi * float64(index) / float64(len(samples))
		samples[index] = planetOccultationFootprintSampleAt(
			tt, theta, frame, siderealDegrees, location,
		)
	}
	if targetSpacingKM > 0 {
		samples = refinePlanetOccultationFootprintSamples(
			tt, frame, siderealDegrees, location, samples, targetSpacingKM,
		)
	}
	samples = appendPlanetOccultationFootprintHorizonCrossings(
		tt, frame, siderealDegrees, location, samples,
	)

	segments, closed := planetOccultationFootprintSegments(samples)
	polygons := make([][]OccultationPathPoint, 0, len(segments))
	interiorPolygons := make([][]OccultationPathPoint, 0)
	for _, segment := range segments {
		if len(segment) < 2 {
			continue
		}
		polygon := append([]OccultationPathPoint(nil), segment...)
		if closed {
			polygon = append(polygon, polygon[0])
		} else {
			polygon = append(polygon, planetOccultationHorizonArc(
				tt, frame, segment, location, horizonPoints,
			)...)
		}
		if len(polygon) >= 4 {
			polygons = append(polygons, polygon)
		}
		if anchor, anchorOK := planetOccultationFootprintAnchor(tt, frame, segment, location); anchorOK &&
			!planetOccultationPathRingContains(polygon, anchor.Longitude, anchor.Latitude) {
			// A sampled open cone can leave a small numerical gap between the
			// visible boundary ring and the axis point. Keep the physical
			// centerline covered with a small cap instead of allowing the
			// greatest marker to fall into the partial annulus.
			repairs := planetOccultationCenterRepair(
				anchor, polygon, planetOccultationCenterCapRadiusKM,
				planetOccultationCenterCapPoints,
			)
			polygons = append(polygons, repairs...)
			interiorPolygons = append(interiorPolygons, repairs...)
		}
	}
	if len(polygons) == 0 {
		return PlanetOccultationFootprint{}, false
	}
	return PlanetOccultationFootprint{
		Time:             occultationTTToLocation(tt, location),
		Polygons:         polygons,
		InteriorPolygons: interiorPolygons,
		Boundaries:       segments,
		Closed:           closed,
	}, true
}

func planetOccultationFootprintSampleAt(
	tt, theta float64,
	frame occultationPathFrame,
	siderealDegrees float64,
	location *time.Location,
) planetOccultationFootprintSample {
	sample := planetOccultationFootprintSample{theta: theta}
	vector, _, valid := occultationPathBoundaryVector(frame, theta)
	if !valid {
		return sample
	}
	// frame.moon already contains the geocentric lunar vector for tt;
	// avoid recomputing the full lunar ephemeris for every boundary point.
	sample.point = occultationPathPointFromVectorWithMoonSidereal(
		tt, vector, 0, frame.moon, siderealDegrees, location,
	)
	sample.ok = true
	return sample
}

func refinePlanetOccultationFootprintSamples(
	tt float64,
	frame occultationPathFrame,
	siderealDegrees float64,
	location *time.Location,
	samples []planetOccultationFootprintSample,
	targetSpacingKM float64,
) []planetOccultationFootprintSample {
	if len(samples) < 2 || targetSpacingKM <= 0 {
		return samples
	}
	result := make([]planetOccultationFootprintSample, 0, len(samples))
	for index, left := range samples {
		right := samples[(index+1)%len(samples)]
		if index == len(samples)-1 {
			right.theta += 2 * math.Pi
		}
		result = append(result, left)
		result = append(result, refinePlanetOccultationFootprintInterval(
			tt, frame, siderealDegrees, location,
			left, right, targetSpacingKM, 0,
		)...)
	}
	return result
}

func refinePlanetOccultationFootprintInterval(
	tt float64,
	frame occultationPathFrame,
	siderealDegrees float64,
	location *time.Location,
	left, right planetOccultationFootprintSample,
	targetSpacingKM float64,
	depth int,
) []planetOccultationFootprintSample {
	if depth >= planetOccultationBandAdaptiveMaxDepth {
		return nil
	}
	if left.ok && right.ok && occultationPathDistanceKM(left.point, right.point) <= occultationFootprintAdaptiveSpacingKM(left.point, right.point, targetSpacingKM) {
		return nil
	}
	middle := planetOccultationFootprintSampleAt(
		tt, (left.theta+right.theta)/2, frame, siderealDegrees, location,
	)
	if !left.ok && !right.ok && !middle.ok {
		return nil
	}
	result := refinePlanetOccultationFootprintInterval(
		tt, frame, siderealDegrees, location,
		left, middle, targetSpacingKM, depth+1,
	)
	result = append(result, middle)
	return append(result, refinePlanetOccultationFootprintInterval(
		tt, frame, siderealDegrees, location,
		middle, right, targetSpacingKM, depth+1,
	)...)
}

// occultationFootprintAdaptiveSpacingKM tightens the contact-limb sampling at
// high latitude. A fixed spherical spacing looks like a long straight chord
// after Web Mercator multiplies longitude by sec(latitude); the physical limb
// is unchanged, but the rendered edge loses its curvature. The cosine floor
// bounds the extra work near the poles while leaving equatorial events on the
// existing spacing budget.
func occultationFootprintAdaptiveSpacingKM(
	first, second OccultationPathPoint,
	targetSpacingKM float64,
) float64 {
	if targetSpacingKM <= 0 {
		return targetSpacingKM
	}
	latitude := math.Max(math.Abs(first.Latitude), math.Abs(second.Latitude)) * math.Pi / 180
	scale := math.Max(0.5, math.Cos(latitude))
	return targetSpacingKM * scale
}

// planetOccultationFootprintSampleVisible 报告样本是否位于月球地平线以上；没有边界解的样本同样不可见。
func planetOccultationFootprintSampleVisible(sample planetOccultationFootprintSample) bool {
	return sample.ok && sample.point.MoonAltitude >= 0
}

// appendPlanetOccultationFootprintHorizonCrossings 在相邻可见/不可见样本之间插入地平线切点，环绕接缝同样处理。
func appendPlanetOccultationFootprintHorizonCrossings(
	tt float64,
	frame occultationPathFrame,
	siderealDegrees float64,
	location *time.Location,
	samples []planetOccultationFootprintSample,
) []planetOccultationFootprintSample {
	if len(samples) < 2 {
		return samples
	}
	// 不跨越地平线的足迹直接复用样本切片，避免每次装配都复制一份。
	straddles := false
	for index, left := range samples {
		right := samples[(index+1)%len(samples)]
		if (left.point.MoonAltitude >= 0) != (right.point.MoonAltitude >= 0) {
			straddles = true
			break
		}
	}
	if !straddles {
		return samples
	}
	result := make([]planetOccultationFootprintSample, 0, len(samples)+4)
	for index, left := range samples {
		right := samples[(index+1)%len(samples)]
		if index+1 == len(samples) {
			right.theta += 2 * math.Pi
		}
		result = append(result, left)
		if crossing, ok := planetOccultationFootprintHorizonCrossing(
			tt, frame, siderealDegrees, location, left, right,
		); ok {
			result = append(result, crossing)
		}
	}
	return result
}

// planetOccultationFootprintHorizonCrossing 二分相邻样本的月球高度过零点，只返回非负高度的样本。
func planetOccultationFootprintHorizonCrossing(
	tt float64,
	frame occultationPathFrame,
	siderealDegrees float64,
	location *time.Location,
	left, right planetOccultationFootprintSample,
) (planetOccultationFootprintSample, bool) {
	if !left.ok || !right.ok || !finite(left.point.MoonAltitude) || !finite(right.point.MoonAltitude) {
		return planetOccultationFootprintSample{}, false
	}
	if (left.point.MoonAltitude >= 0) == (right.point.MoonAltitude >= 0) {
		return planetOccultationFootprintSample{}, false
	}
	above := left
	if right.point.MoonAltitude >= 0 {
		above = right
	}
	for step := 0; step < planetOccultationFootprintHorizonSearchSteps; step++ {
		middle := planetOccultationFootprintSampleAt(
			tt, (left.theta+right.theta)/2, frame, siderealDegrees, location,
		)
		if !middle.ok {
			break
		}
		if middle.point.MoonAltitude >= 0 {
			above = middle
			left = middle
		} else {
			right = middle
		}
		if math.Abs(above.point.MoonAltitude) <= planetOccultationFootprintHorizonToleranceDeg {
			break
		}
	}
	return above, true
}

func planetOccultationFootprintSegments(
	samples []planetOccultationFootprintSample,
) ([][]OccultationPathPoint, bool) {
	segments := make([][]OccultationPathPoint, 0, 2)
	current := make([]OccultationPathPoint, 0, len(samples))
	allVisible := len(samples) > 0
	for _, sample := range samples {
		if !planetOccultationFootprintSampleVisible(sample) {
			allVisible = false
			if len(current) > 0 {
				segments = append(segments, current)
				current = nil
			}
			continue
		}
		current = append(current, sample.point)
	}
	if len(current) > 0 {
		segments = append(segments, current)
	}
	if len(segments) > 1 &&
		planetOccultationFootprintSampleVisible(samples[0]) &&
		planetOccultationFootprintSampleVisible(samples[len(samples)-1]) {
		first := segments[0]
		last := segments[len(segments)-1]
		merged := make([]OccultationPathPoint, 0, len(last)+len(first))
		merged = append(merged, last...)
		merged = append(merged, first...)
		segments[0] = merged
		segments = segments[:len(segments)-1]
	}
	return segments, allVisible && len(segments) == 1
}

func planetOccultationHorizonArc(
	tt float64,
	frame occultationPathFrame,
	segment []OccultationPathPoint,
	location *time.Location,
	pointCount int,
) []OccultationPathPoint {
	if len(segment) < 2 {
		return nil
	}
	circle := planetOccultationHorizonCircle(tt, frame, location, pointCount)
	if len(circle) == 0 {
		return []OccultationPathPoint{segment[0]}
	}
	from, to := segment[len(segment)-1], segment[0]
	unitAt := func(point OccultationPathPoint) occultationPathVector {
		fixed := occultationStationSurfaceVector(point.Longitude, point.Latitude)
		return occultationPathVector{x: fixed.x / occultationPathEarthEquatorialRadiusKM,
			y: fixed.y / occultationPathEarthEquatorialRadiusKM,
			z: fixed.z / (occultationPathEarthEquatorialRadiusKM * occultationPathEarthPolarRatio)}
	}
	moonFixed := occultationPathEarthFixedVector(tt, frame.moon)
	normal := occultationPathUnit(occultationPathVector{x: moonFixed.x, y: moonFixed.y, z: moonFixed.z / occultationPathEarthPolarRatio})
	first := unitAt(circle[0])
	first = occultationPathUnit(occultationPathSub(first, occultationPathScale(normal, occultationPathDot(first, normal))))
	second := occultationPathCross(normal, first)
	angleAt := func(point OccultationPathPoint) float64 {
		value := unitAt(point)
		return math.Atan2(occultationPathDot(value, second), occultationPathDot(value, first))
	}
	// Only interior samples belong to the closure. Rounding either endpoint
	// to its nearest circle sample can extend the arc beyond the contact limb.
	startAngle := angleAt(from)
	delta := math.Remainder(angleAt(to)-startAngle, 2*math.Pi)
	spacing := 2 * math.Pi / float64(len(circle))
	direction := 1
	index := int(math.Floor(startAngle/spacing)) + 1
	if delta < 0 {
		direction = -1
		index = int(math.Ceil(startAngle/spacing)) - 1
	}
	arc := make([]OccultationPathPoint, 0, int(math.Abs(delta)/spacing)+1)
	for angle := float64(index) * spacing; math.Abs(angle-startAngle) < math.Abs(delta); angle += float64(direction) * spacing {
		wrapped := (index%len(circle) + len(circle)) % len(circle)
		point := circle[wrapped]
		// 地平线弧顶点按契约取非负高度，解析圆的模型残差不是地平线以下的几何。
		point.MoonAltitude = math.Max(0, point.MoonAltitude)
		arc = append(arc, point)
		index += direction
	}
	if len(arc) == 0 && math.Abs(delta) > 1e-14 {
		// Even a sub-sample arc needs an interior vertex to close a thin
		// visible crescent. Evaluate its midpoint on the same small circle.
		centerDistance := occultationPathDot(unitAt(circle[0]), normal)
		radius := math.Sqrt(math.Max(0, 1-centerDistance*centerDistance))
		angle := startAngle + delta/2
		unit := occultationPathAdd(occultationPathScale(normal, centerDistance),
			occultationPathScale(occultationPathAdd(occultationPathScale(first, math.Cos(angle)),
				occultationPathScale(second, math.Sin(angle))), radius))
		fixed := occultationPathScale(unit, occultationPathEarthEquatorialRadiusKM)
		fixed.z *= occultationPathEarthPolarRatio
		middle := from
		middle.Longitude, middle.Latitude = occultationStationGeodetic(fixed)
		middle.MoonAltitude = 0
		arc = append(arc, middle)
	}
	return append(arc, to)
}

func planetOccultationFootprintAnchor(
	tt float64,
	frame occultationPathFrame,
	segment []OccultationPathPoint,
	location *time.Location,
) (OccultationPathPoint, bool) {
	if center, _, ok := occultationEarthLineIntersection(frame.moon, frame.axis); ok {
		anchor := occultationPathPointFromVectorWithMoon(
			tt, center, 0, frame.moon, location,
		)
		if anchor.MoonAltitude >= 0 {
			return anchor, true
		}
	}
	if len(segment) == 0 {
		return OccultationPathPoint{}, false
	}
	anchor := segment[0]
	for _, point := range segment[1:] {
		if point.MoonAltitude > anchor.MoonAltitude {
			anchor = point
		}
	}
	return anchor, true
}

func planetOccultationCenterCap(
	center OccultationPathPoint,
	radiusKM float64,
	pointCount int,
) []OccultationPathPoint {
	if pointCount < 3 || radiusKM <= 0 {
		return nil
	}
	const degreesPerRadian = 180 / math.Pi
	latitudeRadians := center.Latitude / degreesPerRadian
	cosine := math.Cos(latitudeRadians)
	if math.Abs(cosine) < 1e-6 {
		cosine = 1e-6
	}
	latitudeOffset := radiusKM / occultationPathEarthEquatorialRadiusKM * degreesPerRadian
	longitudeOffset := latitudeOffset / cosine
	cap := make([]OccultationPathPoint, pointCount+1)
	for index := 0; index < pointCount; index++ {
		angle := 2 * math.Pi * float64(index) / float64(pointCount)
		cap[index] = center
		cap[index].Longitude = normalizeLongitude(center.Longitude + longitudeOffset*math.Cos(angle))
		cap[index].Latitude = math.Max(-90, math.Min(90, center.Latitude+latitudeOffset*math.Sin(angle)))
		cap[index].MoonAltitude = center.MoonAltitude
		cap[index].WidthKM = 0
	}
	cap[pointCount] = cap[0]
	return cap
}

func planetOccultationCenterRepair(
	center OccultationPathPoint,
	ring []OccultationPathPoint,
	radiusKM float64,
	pointCount int,
) [][]OccultationPathPoint {
	cap := planetOccultationCenterCap(center, radiusKM, pointCount)
	if len(ring) < 2 {
		return [][]OccultationPathPoint{cap}
	}
	nearestIndex := 0
	nearestDistance := math.Inf(1)
	for index := 1; index < len(ring); index++ {
		distance := planetOccultationPointSegmentDistanceKM(
			center, ring[index-1], ring[index],
		)
		if distance < nearestDistance {
			nearestDistance = distance
			nearestIndex = index - 1
		}
	}
	nextIndex := (nearestIndex + 1) % len(ring)
	bridge := []OccultationPathPoint{
		center,
		ring[nearestIndex],
		ring[nextIndex],
		center,
	}
	return [][]OccultationPathPoint{cap, bridge}
}

func planetOccultationPointSegmentDistanceKM(
	point, start, end OccultationPathPoint,
) float64 {
	latitude := point.Latitude * math.Pi / 180
	scaleX := math.Cos(latitude) * occultationPathEarthEquatorialRadiusKM * math.Pi / 180
	scaleY := occultationPathEarthEquatorialRadiusKM * math.Pi / 180
	x := func(value OccultationPathPoint) float64 {
		return math.Remainder(value.Longitude-point.Longitude, 360) * scaleX
	}
	y := func(value OccultationPathPoint) float64 {
		return (value.Latitude - point.Latitude) * scaleY
	}
	startX, startY := x(start), y(start)
	endX, endY := x(end), y(end)
	deltaX, deltaY := endX-startX, endY-startY
	fraction := 0.0
	if lengthSquared := deltaX*deltaX + deltaY*deltaY; lengthSquared > 0 {
		fraction = math.Max(0, math.Min(1,
			-(startX*deltaX+startY*deltaY)/lengthSquared,
		))
	}
	return math.Hypot(startX+fraction*deltaX, startY+fraction*deltaY)
}

func planetOccultationPathRingContains(
	ring []OccultationPathPoint,
	longitude, latitude float64,
) bool {
	inside := false
	for current, previous := 0, len(ring)-1; current < len(ring); previous, current = current, current+1 {
		currentLongitude := math.Remainder(ring[current].Longitude-longitude, 360)
		previousLongitude := math.Remainder(ring[previous].Longitude-longitude, 360)
		if math.Abs(currentLongitude-previousLongitude) > 180 {
			if currentLongitude < previousLongitude {
				currentLongitude += 360
			} else {
				previousLongitude += 360
			}
		}
		currentLatitude := ring[current].Latitude
		previousLatitude := ring[previous].Latitude
		if (currentLatitude > latitude) == (previousLatitude > latitude) {
			continue
		}
		intersection := previousLongitude +
			(latitude-previousLatitude)*(currentLongitude-previousLongitude)/
				(currentLatitude-previousLatitude)
		if intersection > 0 {
			inside = !inside
		}
	}
	return inside
}

// planetOccultationHorizonCircle returns the exact lunar horizon on the
// reference ellipsoid. In unit-ellipsoid coordinates, tangent points satisfy
// both |u|=1 and scaledMoon dot u=1, so the horizon is a small circle.
func planetOccultationHorizonCircle(
	tt float64,
	frame occultationPathFrame,
	location *time.Location,
	count int,
) []OccultationPathPoint {
	if count < 3 {
		return nil
	}
	equatorialRadius := occultationPathEarthEquatorialRadiusKM
	polarRadius := equatorialRadius * occultationPathEarthPolarRatio
	scaledMoon := occultationPathVector{
		x: frame.moon.x / equatorialRadius,
		y: frame.moon.y / equatorialRadius,
		z: frame.moon.z / polarRadius,
	}
	distance := occultationPathNorm(scaledMoon)
	if distance <= 1 || !finite(distance) {
		return nil
	}
	centerDirection := occultationPathScale(scaledMoon, 1/distance)
	reference := occultationPathVector{z: 1}
	if math.Abs(centerDirection.z) > 0.9 {
		reference = occultationPathVector{x: 1}
	}
	firstAxis := occultationPathUnit(occultationPathCross(reference, centerDirection))
	secondAxis := occultationPathCross(centerDirection, firstAxis)
	centerDistance := 1 / distance
	circleRadius := math.Sqrt(math.Max(0, 1-centerDistance*centerDistance))
	siderealDegrees := ApparentSiderealTime(TT2UT1(tt)) * 15
	points := make([]OccultationPathPoint, count)
	for index := range points {
		angle := 2 * math.Pi * float64(index) / float64(count)
		unitPoint := occultationPathAdd(
			occultationPathScale(centerDirection, centerDistance),
			occultationPathScale(
				occultationPathAdd(
					occultationPathScale(firstAxis, math.Cos(angle)),
					occultationPathScale(secondAxis, math.Sin(angle)),
				),
				circleRadius,
			),
		)
		vector := occultationPathVector{
			x: equatorialRadius * unitPoint.x,
			y: equatorialRadius * unitPoint.y,
			z: polarRadius * unitPoint.z,
		}
		points[index] = occultationPathPointFromVectorWithMoonSidereal(
			tt, vector, 0, frame.moon, siderealDegrees, location,
		)
	}
	return points
}
