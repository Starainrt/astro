package basic

import (
	"math"
	"sort"
	"time"

	"github.com/starainrt/astro/internal/geodata"
)

func appendOccultationRiseSetSamples(
	tracks []*occultationRiseSetTrack,
	points []OccultationPathPoint,
	stepDays float64,
) []*occultationRiseSetTrack {
	sort.SliceStable(points, func(first, second int) bool {
		return points[first].Longitude < points[second].Longitude
	})
	type match struct {
		point, track int
		distance     float64
	}
	var matches []match
	for pointIndex, point := range points {
		for index, track := range tracks {
			if len(track.segments) == 0 || len(track.segments[len(track.segments)-1]) == 0 {
				continue
			}
			last := track.segments[len(track.segments)-1][len(track.segments[len(track.segments)-1])-1]
			deltaDays := point.Time.Sub(last.Time).Hours() / 24
			if deltaDays <= 0 || deltaDays > 2.5*stepDays {
				continue
			}
			distance := occultationPathDistanceKM(last, point)
			if riseSetGeographicBranchChanged(distance, deltaDays) {
				continue
			}
			matches = append(matches, match{point: pointIndex, track: index, distance: distance})
		}
	}
	// Match the closest pairs before creating tracks for newly appearing
	// roots. Longitude order must not let a new root claim an existing branch
	// ahead of that branch's much closer continuation.
	sort.SliceStable(matches, func(i, j int) bool { return matches[i].distance < matches[j].distance })
	usedTracks, usedPoints := make([]bool, len(tracks)), make([]bool, len(points))
	for _, candidate := range matches {
		if usedTracks[candidate.track] || usedPoints[candidate.point] {
			continue
		}
		usedTracks[candidate.track], usedPoints[candidate.point] = true, true
		track := tracks[candidate.track]
		track.segments[len(track.segments)-1] = append(track.segments[len(track.segments)-1], points[candidate.point])
	}
	for index, point := range points {
		if !usedPoints[index] {
			tracks = append(tracks, &occultationRiseSetTrack{segments: [][]OccultationPathPoint{{point}}})
		}
	}
	return tracks
}

func completeOccultationRiseSetCurveEndpoints(
	curves []OccultationRiseSetCurve,
	stepDays float64,
	location *time.Location,
	cache *occultationRiseSetEvaluationCache,
) {
	for index := range curves {
		completeOccultationRiseSetFoldEndpoints(&curves[index], stepDays, location, cache)
	}
	completeOccultationRiseSetPhaseJunctions(curves, stepDays, location, cache)
	completeOccultationRiseSetDirectionJunctions(curves, stepDays, location, cache)
	refinementCache := cache.candidateOnly()
	for index := range curves {
		// Split sampled branch jumps before inserting adaptive midpoints.  If a
		// midpoint root is singular, refinement must not turn that jump back into
		// a long straight segment.
		normalizeOccultationRiseSetCurveSegments(&curves[index])
		refineOccultationRiseSetCurveSpacing(&curves[index], location, refinementCache)
		normalizeOccultationRiseSetCurveSegments(&curves[index])
	}
	refineOccultationRiseSetPhaseJunctionApproaches(curves, location, refinementCache)
	for index := range curves {
		normalizeOccultationRiseSetCurveSegments(&curves[index])
	}
}

// normalizeOccultationRiseSetCurveSegments keeps each rendered branch
// strictly time-ordered. A completed horizon fold is a branch boundary when
// its time reverses; roots at the same location and time are numerical dupes.
func normalizeOccultationRiseSetCurveSegments(curve *OccultationRiseSetCurve) {
	if curve == nil {
		return
	}
	segments := make([][]OccultationPathPoint, 0, len(curve.Segments))
	for _, segment := range curve.Segments {
		if len(segment) == 0 {
			continue
		}
		current := []OccultationPathPoint{segment[0]}
		for _, point := range segment[1:] {
			last := current[len(current)-1]
			pointTT, lastTT := occultationTimeToTT(point.Time), occultationTimeToTT(last.Time)
			if pointTT > lastTT+occultationRiseSetTimeEpsilonDays {
				if occultationRiseSetBranchChanged(
					occultationPathDistanceKM(last, point), pointTT-lastTT,
				) {
					if len(current) >= 2 {
						segments = append(segments, current)
					}
					current = []OccultationPathPoint{point}
					continue
				}
				current = append(current, point)
				continue
			}
			if math.Abs(pointTT-lastTT) <= occultationRiseSetTimeEpsilonDays &&
				occultationPathDistanceKM(point, last) <= 0.01 {
				continue
			}
			if len(current) >= 2 {
				segments = append(segments, current)
			}
			current = []OccultationPathPoint{point}
		}
		if len(current) >= 2 {
			segments = append(segments, current)
		}
	}
	curve.Segments = segments
}

type occultationRiseSetEndpoint struct {
	curveIndex   int
	segmentIndex int
	atStart      bool
	point        OccultationPathPoint
}

type occultationRiseSetPhaseAttachment struct {
	endpoint occultationRiseSetEndpoint
	fold     OccultationPathPoint
	hasFold  bool
}

func completeOccultationRiseSetFoldEndpoints(
	curve *OccultationRiseSetCurve,
	stepDays float64,
	location *time.Location,
	cache *occultationRiseSetEvaluationCache,
) {
	if curve == nil || len(curve.Segments) < 2 {
		return
	}
	endpoints := make([]occultationRiseSetEndpoint, 0, 2*len(curve.Segments))
	for segmentIndex, segment := range curve.Segments {
		if len(segment) == 0 {
			continue
		}
		for _, atStart := range []bool{true, false} {
			endpoints = append(endpoints, occultationRiseSetEndpoint{
				segmentIndex: segmentIndex, atStart: atStart,
				point: occultationRiseSetSegmentEndpoint(segment, atStart),
			})
		}
	}
	used := make(map[[2]int]bool, len(endpoints))
	for firstIndex := 0; firstIndex < len(endpoints); firstIndex++ {
		first := endpoints[firstIndex]
		firstKey := [2]int{first.segmentIndex, boolInt(first.atStart)}
		if used[firstKey] {
			continue
		}
		for secondIndex := firstIndex + 1; secondIndex < len(endpoints); secondIndex++ {
			second := endpoints[secondIndex]
			secondKey := [2]int{second.segmentIndex, boolInt(second.atStart)}
			if first.segmentIndex == second.segmentIndex || first.atStart != second.atStart || used[secondKey] {
				continue
			}
			// Close roots can cease to be resolved on adjacent time samples.
			// Candidate pairing may span one sample; the fold solver and its
			// physical residuals still decide whether the branches really meet.
			if math.Abs(occultationTimeToTT(first.point.Time)-occultationTimeToTT(second.point.Time)) > math.Max(1e-8, 1.5*stepDays) {
				continue
			}
			distance := occultationPathDistanceKM(first.point, second.point)
			if distance <= 0.01 || distance > 3000 {
				continue
			}
			if !occultationRiseSetFoldCandidate(
				curve.Segments[first.segmentIndex], first.atStart,
				curve.Segments[second.segmentIndex], second.atStart,
			) {
				continue
			}
			fold, ok := refineOccultationRiseSetFold(
				first.point, second.point, first.atStart,
				curve.Phase == RiseSetPhaseGreatest, stepDays, location, cache,
			)
			if !ok || math.Abs(occultationTimeToTT(fold.Time)-(occultationTimeToTT(first.point.Time)+occultationTimeToTT(second.point.Time))/2) > 2.5*stepDays ||
				occultationPathDistanceKM(fold, first.point) > 3000 || occultationPathDistanceKM(fold, second.point) > 3000 {
				continue
			}
			curve.Segments[first.segmentIndex] = occultationRiseSetAddEndpoint(curve.Segments[first.segmentIndex], fold, first.atStart)
			curve.Segments[second.segmentIndex] = occultationRiseSetAddEndpoint(curve.Segments[second.segmentIndex], fold, second.atStart)
			used[firstKey], used[secondKey] = true, true
			break
		}
	}
}

// occultationRiseSetFoldCandidate reports whether two nearby endpoints are
// actually separating into different sampled branches. Endpoints that merely
// happen to be close are common on smooth curves and do not need a fold root.
func occultationRiseSetFoldCandidate(
	first []OccultationPathPoint,
	firstAtStart bool,
	second []OccultationPathPoint,
	secondAtStart bool,
) bool {
	firstAdjacent, firstOK := occultationRiseSetAdjacentPoint(first, firstAtStart)
	secondAdjacent, secondOK := occultationRiseSetAdjacentPoint(second, secondAtStart)
	if !firstOK || !secondOK {
		return true
	}
	endpointDistance := occultationPathDistanceKM(
		occultationRiseSetSegmentEndpoint(first, firstAtStart),
		occultationRiseSetSegmentEndpoint(second, secondAtStart),
	)
	adjacentDistance := occultationPathDistanceKM(firstAdjacent, secondAdjacent)
	margin := 250.0
	firstEndpoint := occultationRiseSetSegmentEndpoint(first, firstAtStart)
	secondEndpoint := occultationRiseSetSegmentEndpoint(second, secondAtStart)
	if math.Max(math.Abs(firstEndpoint.Latitude), math.Abs(secondEndpoint.Latitude)) >= 70 {
		// Longitude convergence makes the first post-fold samples unusually
		// close near a pole. Leave numerical room for a 10-second sample to
		// land just inside the ordinary 250 km divergence margin.
		margin = 200
	}
	return adjacentDistance > endpointDistance+margin
}

func occultationRiseSetAdjacentPoint(segment []OccultationPathPoint, atStart bool) (OccultationPathPoint, bool) {
	if len(segment) < 2 {
		return OccultationPathPoint{}, false
	}
	if atStart {
		return segment[1], true
	}
	return segment[len(segment)-2], true
}

func completeOccultationRiseSetPhaseJunctions(
	curves []OccultationRiseSetCurve,
	stepDays float64,
	location *time.Location,
	cache *occultationRiseSetEvaluationCache,
) {
	curveIndices := make(map[occultationRiseSetCurveKey]int, len(curves))
	for index, curve := range curves {
		curveIndices[occultationRiseSetCurveKey{phase: curve.Phase, direction: curve.Direction}] = index
	}
	for _, direction := range []RiseSetDirection{RiseSetDirectionRise, RiseSetDirectionSet} {
		startIndex, haveStart := curveIndices[occultationRiseSetCurveKey{phase: RiseSetPhaseStart, direction: direction}]
		greatestIndex, haveGreatest := curveIndices[occultationRiseSetCurveKey{phase: RiseSetPhaseGreatest, direction: direction}]
		endIndex, haveEnd := curveIndices[occultationRiseSetCurveKey{phase: RiseSetPhaseEnd, direction: direction}]
		if !haveStart || !haveGreatest || !haveEnd {
			continue
		}
		seeds := occultationRiseSetUnsharedEndpoints(startIndex, curves[startIndex].Segments)
		for _, seed := range seeds {
			candidateSeeds := []OccultationPathPoint{seed.point}
			if endEndpoint, endOK := occultationRiseSetClosestEndpoint(seed.point, curves[endIndex].Segments, stepDays); endOK {
				candidateSeeds = append([]OccultationPathPoint{
					occultationRiseSetMidpoint(seed.point, endEndpoint.point),
					endEndpoint.point,
				}, candidateSeeds...)
			}
			junction, ok := OccultationPathPoint{}, false
			for _, candidate := range candidateSeeds {
				junction, ok = refineOccultationRiseSetPhaseJunctionOnHorizon(candidate, location, cache)
				if !ok {
					junction, ok = refineOccultationRiseSetPhaseJunction(candidate, location, cache)
				}
				if ok {
					break
				}
			}
			if !ok || math.Abs(occultationTimeToTT(junction.Time)-occultationTimeToTT(seed.point.Time)) > 3*stepDays ||
				occultationPathDistanceKM(junction, seed.point) > 3000 {
				continue
			}
			matched := make([]occultationRiseSetPhaseAttachment, 0, 3)
			attachmentTargets := make(map[int]OccultationPathPoint, 1)
			recovery := occultationRiseSetNarrowPhaseRecovery{}
			recovered := false
			// Contact extrema include the changing apparent radii; greatest
			// separation need not end at exactly the same time. Complete the
			// start/end pair even when the diagnostic greatest branch differs.
			for _, curveIndex := range []int{startIndex, endIndex, greatestIndex} {
				endpoint, endpointOK := occultationRiseSetClosestEndpoint(
					junction, curves[curveIndex].Segments, stepDays,
				)
				if !endpointOK {
					if curveIndex == greatestIndex {
						continue
					}
					matched = nil
					break
				}
				endpoint.curveIndex = curveIndex
				attachment := occultationRiseSetPhaseAttachment{endpoint: endpoint}
				if curveIndex == greatestIndex && occultationPathDistanceKM(junction, endpoint.point) > occultationPathBoundarySpacingKM {
					if !occultationRiseSetPhaseSegmentIsContinuous(
						junction, endpoint.point, RiseSetPhaseGreatest, direction,
						location, cache,
					) {
						fold, foldOK := refineOccultationRiseSetFold(
							junction, endpoint.point, endpoint.atStart, true,
							stepDays, location, cache,
						)
						if !foldOK || !occultationRiseSetFoldBridgesPhaseJunction(
							junction, endpoint, fold, direction, cache,
						) {
							recovery, recovered = occultationRiseSetRecoverNarrowPhaseJunction(
								junction, endpoint, direction, stepDays, location, cache,
							)
							if !recovered {
								break
							}
							attachmentTargets[curveIndex] = recovery.junction
						} else {
							attachment.fold, attachment.hasFold = fold, true
						}
					}
				}
				matched = append(matched, attachment)
			}
			if len(matched) < 2 {
				continue
			}
			for _, attachment := range matched {
				endpoint := attachment.endpoint
				target := junction
				if recoveredTarget, ok := attachmentTargets[endpoint.curveIndex]; ok {
					target = recoveredTarget
				}
				if attachment.hasFold {
					curves[endpoint.curveIndex].Segments[endpoint.segmentIndex] = occultationRiseSetAddEndpoint(
						curves[endpoint.curveIndex].Segments[endpoint.segmentIndex], attachment.fold, endpoint.atStart,
					)
					shortBranch := []OccultationPathPoint{attachment.fold, target}
					if !endpoint.atStart {
						shortBranch[0], shortBranch[1] = shortBranch[1], shortBranch[0]
					}
					curves[endpoint.curveIndex].Segments = append(curves[endpoint.curveIndex].Segments, shortBranch)
					continue
				}
				curves[endpoint.curveIndex].Segments[endpoint.segmentIndex] = occultationRiseSetAddEndpoint(
					curves[endpoint.curveIndex].Segments[endpoint.segmentIndex], target, endpoint.atStart,
				)
			}
			if recovered {
				if segment := recovery.segments[RiseSetPhaseStart]; len(segment) >= 2 {
					curves[startIndex].Segments = append(curves[startIndex].Segments, segment)
				}
				if segment := recovery.segments[RiseSetPhaseEnd]; len(segment) >= 2 {
					curves[endIndex].Segments = append(curves[endIndex].Segments, segment)
				}
			}
		}
	}
}

func occultationRiseSetPhaseSegmentIsContinuous(
	start, end OccultationPathPoint,
	phase RiseSetPhase,
	direction RiseSetDirection,
	location *time.Location,
	cache *occultationRiseSetEvaluationCache,
) bool {
	startTT := occultationTimeToTT(start.Time)
	endTT := occultationTimeToTT(end.Time)
	if math.Abs(endTT-startTT)*86400 < 0.1 {
		return false
	}
	totalDistance := occultationPathDistanceKM(start, end)
	continuityToleranceKM := math.Max(100, 0.1*totalDistance)
	candidate := end
	for divisor := 2.0; divisor <= 1024; divisor *= 2 {
		tt := startTT + (endTT-startTT)/divisor
		seedAngle := occultationRiseSetHorizonAngle(
			tt, candidate.Longitude, candidate.Latitude, cache.context,
		)
		next, ok := occultationRiseSetPhasePointOnHorizon(
			tt, seedAngle, phase, direction, location, cache,
		)
		if !ok {
			return occultationPathDistanceKM(start, candidate) <= continuityToleranceKM
		}
		candidate = next
		// At a three-phase junction the root becomes numerically singular.
		// Once the traced branch is already inside the spatial tolerance,
		// continuing toward the exact junction can jump to its sibling root.
		if occultationPathDistanceKM(start, candidate) <= continuityToleranceKM {
			return true
		}
	}
	return occultationPathDistanceKM(start, candidate) <= continuityToleranceKM
}

type occultationRiseSetNarrowPhaseRecovery struct {
	junction OccultationPathPoint
	segments map[RiseSetPhase][]OccultationPathPoint
}

// A pair of phase junctions can be only a few seconds apart. The ordinary
// horizon scan cannot resolve both roots, so continue the already sampled
// greatest-phase branch to its contact crossing and reconstruct the two short
// contact branches only for that local degeneracy.
func occultationRiseSetRecoverNarrowPhaseJunction(
	primary OccultationPathPoint,
	greatestEndpoint occultationRiseSetEndpoint,
	direction RiseSetDirection,
	stepDays float64,
	location *time.Location,
	cache *occultationRiseSetEvaluationCache,
) (occultationRiseSetNarrowPhaseRecovery, bool) {
	fold, ok := refineOccultationRiseSetFold(
		primary, greatestEndpoint.point, greatestEndpoint.atStart,
		true, stepDays, location, cache,
	)
	if !ok {
		return occultationRiseSetNarrowPhaseRecovery{}, false
	}
	primaryTT := occultationTimeToTT(primary.Time)
	endpointTT := occultationTimeToTT(greatestEndpoint.point.Time)
	foldTT := occultationTimeToTT(fold.Time)
	extensionDirection := primaryTT - endpointTT
	if extensionDirection == 0 || (foldTT-primaryTT)*extensionDirection <= 0 {
		return occultationRiseSetNarrowPhaseRecovery{}, false
	}
	seedAngle := occultationRiseSetHorizonAngle(
		primaryTT, greatestEndpoint.point.Longitude, greatestEndpoint.point.Latitude, cache.context,
	)
	branchPoint, branchEvaluation, ok := occultationRiseSetRawPhasePointOnHorizon(
		primaryTT, seedAngle, true, location, cache,
	)
	if !ok || !occultationRiseSetDirectionMatches(
		branchEvaluation, branchPoint, direction,
	) {
		return occultationRiseSetNarrowPhaseRecovery{}, false
	}
	branchContact := branchEvaluation.center.stateAt(
		branchPoint.Longitude, branchPoint.Latitude,
	).contactMetric
	foldEvaluation := cache.evaluation(foldTT)
	foldContact := foldEvaluation.center.stateAt(fold.Longitude, fold.Latitude).contactMetric
	if !finite(branchContact) || !finite(foldContact) || branchContact >= 0 || foldContact <= 0 {
		return occultationRiseSetNarrowPhaseRecovery{}, false
	}

	inside, outside := branchPoint, fold
	insideContact, outsideContact := branchContact, foldContact
	for iteration := 0; iteration < 56; iteration++ {
		insideTT := occultationTimeToTT(inside.Time)
		outsideTT := occultationTimeToTT(outside.Time)
		if math.Abs(outsideTT-insideTT) <= 1e-10 {
			break
		}
		middleTT := (insideTT + outsideTT) / 2
		insideAngle := occultationRiseSetHorizonAngle(
			middleTT, inside.Longitude, inside.Latitude, cache.context,
		)
		outsideAngle := occultationRiseSetHorizonAngle(
			middleTT, outside.Longitude, outside.Latitude, cache.context,
		)
		if !finite(insideAngle) || !finite(outsideAngle) {
			return occultationRiseSetNarrowPhaseRecovery{}, false
		}
		middleAngle := riseSetNormalizeRadians(
			insideAngle + math.Remainder(outsideAngle-insideAngle, 2*math.Pi)/2,
		)
		middle, evaluation, middleOK := occultationRiseSetRawPhasePointOnHorizon(
			middleTT, middleAngle, true, location, cache,
		)
		if !middleOK || !occultationRiseSetDirectionMatches(evaluation, middle, direction) {
			outside = OccultationPathPoint{
				Time: occultationTTToLocation(middleTT, location),
				Longitude: geodata.NormalizeLongitude(
					inside.Longitude + math.Remainder(outside.Longitude-inside.Longitude, 360)/2,
				),
				Latitude: (inside.Latitude + outside.Latitude) / 2,
			}
			continue
		}
		contact := evaluation.center.stateAt(middle.Longitude, middle.Latitude).contactMetric
		if !finite(contact) {
			return occultationRiseSetNarrowPhaseRecovery{}, false
		}
		if contact <= 0 {
			inside, insideContact = middle, contact
		} else {
			outside, outsideContact = middle, contact
		}
	}
	secondary := inside
	if math.Abs(outsideContact) < math.Abs(insideContact) {
		secondary = outside
	}
	// 时间区间收缩后两地平根仍差求根精度，沿短弦插值接触零点而不是取较优端点。
	if denominator := insideContact - outsideContact; denominator != 0 {
		weight := insideContact / denominator
		insideTT := occultationTimeToTT(inside.Time)
		interpolatedTT := insideTT + (occultationTimeToTT(outside.Time)-insideTT)*weight
		interpolated := OccultationPathPoint{
			Time: occultationTTToLocation(interpolatedTT, location),
			Longitude: geodata.NormalizeLongitude(
				inside.Longitude + math.Remainder(outside.Longitude-inside.Longitude, 360)*weight,
			),
			Latitude: inside.Latitude + (outside.Latitude-inside.Latitude)*weight,
		}
		state := cache.evaluation(interpolatedTT).center.stateAt(interpolated.Longitude, interpolated.Latitude)
		if state.valid && finite(state.contactMetric) &&
			math.Abs(state.contactMetric) < math.Min(math.Abs(insideContact), math.Abs(outsideContact)) {
			secondary = interpolated
		}
	}
	secondaryTT := occultationTimeToTT(secondary.Time)
	secondaryEvaluation := cache.evaluation(secondaryTT)
	secondaryState := secondaryEvaluation.center.stateAt(secondary.Longitude, secondary.Latitude)
	if !secondaryState.valid || math.Abs(secondaryState.contactMetric) > 1e-7 ||
		math.Abs(secondaryEvaluation.separationDerivative(secondary.Longitude, secondary.Latitude)) > 1e-7 ||
		math.Abs(secondaryState.moonAltitude) > 1e-7 ||
		!occultationRiseSetDirectionMatches(secondaryEvaluation, secondary, direction) ||
		math.Abs(secondaryTT-primaryTT) > stepDays ||
		occultationPathDistanceKM(primary, secondary) > 500 {
		return occultationRiseSetNarrowPhaseRecovery{}, false
	}

	phaseRoots, ok := occultationRiseSetNarrowContactRoots(
		primary, secondary, direction, location, cache,
	)
	if !ok {
		return occultationRiseSetNarrowPhaseRecovery{}, false
	}
	first, last := primary, secondary
	if last.Time.Before(first.Time) {
		first, last = last, first
	}
	segments := make(map[RiseSetPhase][]OccultationPathPoint, 2)
	for phase, root := range phaseRoots {
		segments[phase] = []OccultationPathPoint{first, root, last}
	}
	return occultationRiseSetNarrowPhaseRecovery{junction: secondary, segments: segments}, true
}

func occultationRiseSetDirectionMatches(
	evaluation occultationRiseSetEvaluation,
	point OccultationPathPoint,
	direction RiseSetDirection,
) bool {
	derivative := evaluation.moonAltitudeDerivative(point.Longitude, point.Latitude)
	return finite(derivative) && ((direction == RiseSetDirectionRise && derivative > 0) ||
		(direction == RiseSetDirectionSet && derivative < 0))
}

func occultationRiseSetNarrowContactRoots(
	first, second OccultationPathPoint,
	direction RiseSetDirection,
	location *time.Location,
	cache *occultationRiseSetEvaluationCache,
) (map[RiseSetPhase]OccultationPathPoint, bool) {
	tt := (occultationTimeToTT(first.Time) + occultationTimeToTT(second.Time)) / 2
	firstAngle := occultationRiseSetHorizonAngle(tt, first.Longitude, first.Latitude, cache.context)
	secondAngle := occultationRiseSetHorizonAngle(tt, second.Longitude, second.Latitude, cache.context)
	if !finite(firstAngle) || !finite(secondAngle) {
		return nil, false
	}
	centerAngle := riseSetNormalizeRadians(
		firstAngle + math.Remainder(secondAngle-firstAngle, 2*math.Pi)/2,
	)
	result := make(map[RiseSetPhase]OccultationPathPoint, 2)
	for _, phase := range []RiseSetPhase{RiseSetPhaseStart, RiseSetPhaseEnd} {
		found := false
		for scale := 0; scale < 22 && !found; scale++ {
			delta := 0.0
			if scale > 0 {
				delta = math.Ldexp(1e-6, scale-1)
			}
			for _, sign := range []float64{-1, 1} {
				if delta == 0 && sign > 0 {
					continue
				}
				point, ok := occultationRiseSetPhasePointOnHorizon(
					tt, centerAngle+sign*delta, phase, direction, location, cache,
				)
				if !ok || math.Min(
					occultationPathDistanceKM(first, point),
					occultationPathDistanceKM(second, point),
				) > 500 {
					continue
				}
				result[phase], found = point, true
				break
			}
		}
		if !found {
			continue
		}
	}
	return result, len(result) > 0
}

func occultationRiseSetFoldBridgesPhaseJunction(
	junction OccultationPathPoint,
	endpoint occultationRiseSetEndpoint,
	fold OccultationPathPoint,
	direction RiseSetDirection,
	cache *occultationRiseSetEvaluationCache,
) bool {
	junctionTT := occultationTimeToTT(junction.Time)
	endpointTT := occultationTimeToTT(endpoint.point.Time)
	foldTT := occultationTimeToTT(fold.Time)
	const timeToleranceDays = 1e-8
	if endpoint.atStart {
		if foldTT > math.Min(junctionTT, endpointTT)+timeToleranceDays {
			return false
		}
	} else if foldTT < math.Max(junctionTT, endpointTT)-timeToleranceDays {
		return false
	}
	if occultationPathDistanceKM(fold, junction) > 3000 ||
		occultationPathDistanceKM(fold, endpoint.point) > 3000 {
		return false
	}
	evaluation := cache.evaluation(foldTT)
	_, key, valid := evaluation.classify(fold.Longitude, fold.Latitude, true, fold.Time.Location())
	return valid && key.phase == RiseSetPhaseGreatest && key.direction == direction
}

func occultationRiseSetMidpoint(first, second OccultationPathPoint) OccultationPathPoint {
	firstTT := occultationTimeToTT(first.Time)
	secondTT := occultationTimeToTT(second.Time)
	return OccultationPathPoint{
		Time: occultationTTToLocation((firstTT+secondTT)/2, first.Time.Location()),
		Longitude: geodata.NormalizeLongitude(
			first.Longitude + math.Remainder(second.Longitude-first.Longitude, 360)/2,
		),
		Latitude:     (first.Latitude + second.Latitude) / 2,
		MoonAltitude: (first.MoonAltitude + second.MoonAltitude) / 2,
	}
}

func completeOccultationRiseSetDirectionJunctions(
	curves []OccultationRiseSetCurve,
	stepDays float64,
	location *time.Location,
	cache *occultationRiseSetEvaluationCache,
) {
	type candidate struct {
		riseIndex, setIndex int
		junction            OccultationPathPoint
		riseAttachment      occultationRiseSetPhaseAttachment
		setAttachment       occultationRiseSetPhaseAttachment
		metric              float64
	}
	curveIndices := make(map[occultationRiseSetCurveKey]int, len(curves))
	for index, curve := range curves {
		curveIndices[occultationRiseSetCurveKey{phase: curve.Phase, direction: curve.Direction}] = index
	}
	for _, phase := range []RiseSetPhase{RiseSetPhaseStart, RiseSetPhaseGreatest, RiseSetPhaseEnd} {
		riseIndex, haveRise := curveIndices[occultationRiseSetCurveKey{phase: phase, direction: RiseSetDirectionRise}]
		setIndex, haveSet := curveIndices[occultationRiseSetCurveKey{phase: phase, direction: RiseSetDirectionSet}]
		if !haveRise || !haveSet {
			continue
		}
		riseEndpoints := occultationRiseSetAllEndpoints(curves[riseIndex].Segments)
		setEndpoints := occultationRiseSetAllEndpoints(curves[setIndex].Segments)
		candidates := make([]candidate, 0, len(riseEndpoints)*len(setEndpoints))
		for riseEndpointIndex, riseEndpoint := range riseEndpoints {
			for setEndpointIndex, setEndpoint := range setEndpoints {
				deltaDays := math.Abs(occultationTimeToTT(riseEndpoint.point.Time) - occultationTimeToTT(setEndpoint.point.Time))
				if deltaDays > 3*stepDays {
					continue
				}
				distance := occultationPathDistanceKM(riseEndpoint.point, setEndpoint.point)
				if distance > 3000 {
					continue
				}
				junction, ok := refineOccultationRiseSetDirectionJunctionOnHorizon(
					riseEndpoint.point, setEndpoint.point, phase == RiseSetPhaseGreatest, location, cache,
				)
				if !ok {
					longitude := geodata.NormalizeLongitude(riseEndpoint.point.Longitude + math.Remainder(setEndpoint.point.Longitude-riseEndpoint.point.Longitude, 360)/2)
					latitude := (riseEndpoint.point.Latitude + setEndpoint.point.Latitude) / 2
					junction, ok = refineOccultationRiseSetDirectionJunction(
						(occultationTimeToTT(riseEndpoint.point.Time)+occultationTimeToTT(setEndpoint.point.Time))/2,
						longitude, latitude, phase == RiseSetPhaseGreatest, location, cache,
					)
				}
				if !ok {
					continue
				}
				junctionDeltaRise := math.Abs(occultationTimeToTT(junction.Time) - occultationTimeToTT(riseEndpoint.point.Time))
				junctionDeltaSet := math.Abs(occultationTimeToTT(junction.Time) - occultationTimeToTT(setEndpoint.point.Time))
				junctionDistanceRise := occultationPathDistanceKM(junction, riseEndpoint.point)
				junctionDistanceSet := occultationPathDistanceKM(junction, setEndpoint.point)
				if !occultationRiseSetEndpointAcceptsJunction(riseEndpoint, junction) ||
					!occultationRiseSetEndpointAcceptsJunction(setEndpoint, junction) ||
					junctionDeltaRise > 3*stepDays || junctionDeltaSet > 3*stepDays ||
					junctionDistanceRise > 3000 || junctionDistanceSet > 3000 ||
					occultationRiseSetBranchChanged(junctionDistanceRise, junctionDeltaRise) ||
					occultationRiseSetBranchChanged(junctionDistanceSet, junctionDeltaSet) {
					continue
				}
				riseAttachment, riseOK := occultationRiseSetDirectionAttachment(
					riseEndpoint, junction, phase, RiseSetDirectionRise,
					stepDays, location, cache,
				)
				setAttachment, setOK := occultationRiseSetDirectionAttachment(
					setEndpoint, junction, phase, RiseSetDirectionSet,
					stepDays, location, cache,
				)
				if !riseOK || !setOK {
					continue
				}
				metric := occultationPathDistanceKM(junction, riseEndpoint.point) +
					occultationPathDistanceKM(junction, setEndpoint.point)
				candidates = append(candidates, candidate{
					riseIndex: riseEndpointIndex, setIndex: setEndpointIndex,
					junction: junction, riseAttachment: riseAttachment,
					setAttachment: setAttachment, metric: metric,
				})
			}
		}
		sort.SliceStable(candidates, func(first, second int) bool {
			return candidates[first].metric < candidates[second].metric
		})
		usedRise := make([]bool, len(riseEndpoints))
		usedSet := make([]bool, len(setEndpoints))
		for _, candidate := range candidates {
			if usedRise[candidate.riseIndex] || usedSet[candidate.setIndex] {
				continue
			}
			occultationRiseSetAttachDirectionJunction(
				&curves[riseIndex], candidate.riseAttachment, candidate.junction,
			)
			occultationRiseSetAttachDirectionJunction(
				&curves[setIndex], candidate.setAttachment, candidate.junction,
			)
			usedRise[candidate.riseIndex] = true
			usedSet[candidate.setIndex] = true
		}
	}
}

func occultationRiseSetDirectionAttachment(
	endpoint occultationRiseSetEndpoint,
	junction OccultationPathPoint,
	phase RiseSetPhase,
	direction RiseSetDirection,
	stepDays float64,
	location *time.Location,
	cache *occultationRiseSetEvaluationCache,
) (occultationRiseSetPhaseAttachment, bool) {
	attachment := occultationRiseSetPhaseAttachment{endpoint: endpoint}
	if occultationPathDistanceKM(junction, endpoint.point) <= occultationPathBoundarySpacingKM ||
		occultationRiseSetPhaseSegmentIsContinuous(
			junction, endpoint.point, phase, direction, location, cache,
		) {
		return attachment, true
	}
	fold, ok := refineOccultationRiseSetFold(
		junction, endpoint.point, endpoint.atStart,
		phase == RiseSetPhaseGreatest, stepDays, location, cache,
	)
	if !ok || !occultationRiseSetFoldBridgesDirectionJunction(
		junction, endpoint, fold, phase, direction, cache,
	) {
		return occultationRiseSetPhaseAttachment{}, false
	}
	attachment.fold, attachment.hasFold = fold, true
	return attachment, true
}

func occultationRiseSetAttachDirectionJunction(
	curve *OccultationRiseSetCurve,
	attachment occultationRiseSetPhaseAttachment,
	junction OccultationPathPoint,
) {
	endpoint := attachment.endpoint
	if attachment.hasFold {
		curve.Segments[endpoint.segmentIndex] = occultationRiseSetAddEndpoint(
			curve.Segments[endpoint.segmentIndex], attachment.fold, endpoint.atStart,
		)
		shortBranch := []OccultationPathPoint{attachment.fold, junction}
		if !endpoint.atStart {
			shortBranch[0], shortBranch[1] = shortBranch[1], shortBranch[0]
		}
		curve.Segments = append(curve.Segments, shortBranch)
		return
	}
	curve.Segments[endpoint.segmentIndex] = occultationRiseSetAddEndpoint(
		curve.Segments[endpoint.segmentIndex], junction, endpoint.atStart,
	)
}

func occultationRiseSetFoldBridgesDirectionJunction(
	junction OccultationPathPoint,
	endpoint occultationRiseSetEndpoint,
	fold OccultationPathPoint,
	phase RiseSetPhase,
	direction RiseSetDirection,
	cache *occultationRiseSetEvaluationCache,
) bool {
	junctionTT := occultationTimeToTT(junction.Time)
	endpointTT := occultationTimeToTT(endpoint.point.Time)
	foldTT := occultationTimeToTT(fold.Time)
	if endpoint.atStart {
		if foldTT > math.Min(junctionTT, endpointTT)+occultationRiseSetTimeEpsilonDays {
			return false
		}
	} else if foldTT < math.Max(junctionTT, endpointTT)-occultationRiseSetTimeEpsilonDays {
		return false
	}
	if occultationPathDistanceKM(fold, junction) > 3000 ||
		occultationPathDistanceKM(fold, endpoint.point) > 3000 {
		return false
	}
	evaluation := cache.evaluation(foldTT)
	_, key, valid := evaluation.classify(
		fold.Longitude, fold.Latitude, phase == RiseSetPhaseGreatest, fold.Time.Location(),
	)
	return valid && key.phase == phase && key.direction == direction
}

func occultationRiseSetEndpointAcceptsJunction(
	endpoint occultationRiseSetEndpoint,
	junction OccultationPathPoint,
) bool {
	endpointTT := occultationTimeToTT(endpoint.point.Time)
	junctionTT := occultationTimeToTT(junction.Time)
	if endpoint.atStart {
		return junctionTT <= endpointTT+occultationRiseSetTimeEpsilonDays
	}
	return junctionTT >= endpointTT-occultationRiseSetTimeEpsilonDays
}

func occultationRiseSetUnsharedEndpoints(curveIndex int, segments [][]OccultationPathPoint) []occultationRiseSetEndpoint {
	endpoints := occultationRiseSetAllEndpoints(segments)
	result := make([]occultationRiseSetEndpoint, 0, len(endpoints))
	for index, endpoint := range endpoints {
		shared := false
		for otherIndex, other := range endpoints {
			if index == otherIndex || endpoint.segmentIndex == other.segmentIndex {
				continue
			}
			if math.Abs(occultationTimeToTT(endpoint.point.Time)-occultationTimeToTT(other.point.Time)) <= 1e-8 &&
				occultationPathDistanceKM(endpoint.point, other.point) <= 0.01 {
				shared = true
				break
			}
		}
		if !shared {
			endpoint.curveIndex = curveIndex
			result = append(result, endpoint)
		}
	}
	return result
}

func occultationRiseSetAllEndpoints(segments [][]OccultationPathPoint) []occultationRiseSetEndpoint {
	endpoints := make([]occultationRiseSetEndpoint, 0, 2*len(segments))
	for segmentIndex, segment := range segments {
		if len(segment) == 0 {
			continue
		}
		for _, atStart := range []bool{true, false} {
			endpoints = append(endpoints, occultationRiseSetEndpoint{
				segmentIndex: segmentIndex, atStart: atStart,
				point: occultationRiseSetSegmentEndpoint(segment, atStart),
			})
		}
	}
	return endpoints
}

func occultationRiseSetClosestEndpoint(
	junction OccultationPathPoint,
	segments [][]OccultationPathPoint,
	stepDays float64,
) (occultationRiseSetEndpoint, bool) {
	best := occultationRiseSetEndpoint{}
	bestMetric := math.Inf(1)
	junctionTT := occultationTimeToTT(junction.Time)
	for segmentIndex, segment := range segments {
		if len(segment) == 0 {
			continue
		}
		for _, atStart := range []bool{true, false} {
			point := occultationRiseSetSegmentEndpoint(segment, atStart)
			pointTT := occultationTimeToTT(point.Time)
			if atStart && junctionTT > pointTT+1e-8 || !atStart && junctionTT < pointTT-1e-8 ||
				math.Abs(junctionTT-pointTT) > 3*stepDays || occultationPathDistanceKM(junction, point) > 3000 {
				continue
			}
			metric := occultationPathDistanceKM(junction, point) + math.Abs(junctionTT-pointTT)*8640
			if metric < bestMetric {
				best = occultationRiseSetEndpoint{segmentIndex: segmentIndex, atStart: atStart, point: point}
				bestMetric = metric
			}
		}
	}
	return best, bestMetric < math.Inf(1)
}

func refineOccultationRiseSetPhaseJunction(
	seed OccultationPathPoint,
	location *time.Location,
	cache *occultationRiseSetEvaluationCache,
) (OccultationPathPoint, bool) {
	coordinates := [3]float64{seed.Longitude, seed.Latitude, occultationTimeToTT(seed.Time)}
	for iteration := 0; iteration < 24; iteration++ {
		evaluation := cache.evaluation(coordinates[2])
		residual, ok := occultationRiseSetPhaseJunctionResidualAt(evaluation, coordinates[0], coordinates[1])
		if !ok {
			return OccultationPathPoint{}, false
		}
		if math.Abs(residual[0]) <= 1e-8 && math.Abs(residual[1]) <= occultationRiseSetJunctionDerivativeTolerance && math.Abs(residual[2]) <= 1e-8 {
			break
		}
		matrix, ok := occultationRiseSetPhaseJunctionJacobian(coordinates, residual, cache)
		if !ok {
			return OccultationPathPoint{}, false
		}
		delta, ok := solveSolarEclipse3x3(matrix, [3]float64{-residual[0], -residual[1], -residual[2]})
		if !ok {
			return OccultationPathPoint{}, false
		}
		occultationRiseSetLimitJunctionDelta(&delta)
		coordinates[0] = geodata.NormalizeLongitude(coordinates[0] + delta[0])
		coordinates[1] += delta[1]
		coordinates[2] += delta[2]
		if coordinates[1] <= -89.999999 || coordinates[1] >= 89.999999 {
			return OccultationPathPoint{}, false
		}
	}
	evaluation := cache.evaluation(coordinates[2])
	residual, ok := occultationRiseSetPhaseJunctionResidualAt(evaluation, coordinates[0], coordinates[1])
	if !ok || math.Abs(residual[0]) > 1e-7 || math.Abs(residual[1]) > occultationRiseSetJunctionDerivativeTolerance || math.Abs(residual[2]) > 1e-7 {
		return OccultationPathPoint{}, false
	}
	state := evaluation.center.stateAt(coordinates[0], coordinates[1])
	return OccultationPathPoint{Time: occultationTTToLocation(coordinates[2], location), Longitude: coordinates[0], Latitude: coordinates[1], MoonAltitude: state.moonAltitude}, true
}

func refineOccultationRiseSetPhaseJunctionOnHorizon(
	seed OccultationPathPoint,
	location *time.Location,
	cache *occultationRiseSetEvaluationCache,
) (OccultationPathPoint, bool) {
	seedTT := occultationTimeToTT(seed.Time)
	coordinates := [2]float64{
		occultationRiseSetHorizonAngle(seedTT, seed.Longitude, seed.Latitude, cache.context),
		0,
	}
	residualAt := func(value [2]float64) ([2]float64, float64, float64, float64, bool) {
		tt := seedTT + value[1]/1440
		longitude, latitude, horizonOK := occultationRiseSetHorizonPoint(tt, value[0], cache.context)
		if !horizonOK {
			return [2]float64{}, 0, 0, 0, false
		}
		evaluation := cache.evaluation(tt)
		state := evaluation.center.stateAt(longitude, latitude)
		residual := [2]float64{state.contactMetric, evaluation.contactDerivative(longitude, latitude)}
		return residual, tt, longitude, latitude,
			state.valid && finite(residual[0]) && finite(residual[1])
	}
	const angleStep = 1e-4
	const timeStepMinutes = 1.0 / 60.0
	for iteration := 0; iteration < 32; iteration++ {
		residual, _, _, _, ok := residualAt(coordinates)
		if !ok {
			return OccultationPathPoint{}, false
		}
		if math.Abs(residual[0]) <= 1e-8 && math.Abs(residual[1]) <= occultationRiseSetJunctionDerivativeTolerance {
			break
		}
		plusAngle, _, _, _, plusAngleOK := residualAt([2]float64{coordinates[0] + angleStep, coordinates[1]})
		minusAngle, _, _, _, minusAngleOK := residualAt([2]float64{coordinates[0] - angleStep, coordinates[1]})
		plusTime, _, _, _, plusTimeOK := residualAt([2]float64{coordinates[0], coordinates[1] + timeStepMinutes})
		minusTime, _, _, _, minusTimeOK := residualAt([2]float64{coordinates[0], coordinates[1] - timeStepMinutes})
		if !plusAngleOK || !minusAngleOK || !plusTimeOK || !minusTimeOK {
			return OccultationPathPoint{}, false
		}
		matrix := [2][2]float64{
			{(plusAngle[0] - minusAngle[0]) / (2 * angleStep), (plusTime[0] - minusTime[0]) / (2 * timeStepMinutes)},
			{(plusAngle[1] - minusAngle[1]) / (2 * angleStep), (plusTime[1] - minusTime[1]) / (2 * timeStepMinutes)},
		}
		determinant := matrix[0][0]*matrix[1][1] - matrix[0][1]*matrix[1][0]
		if !finite(determinant) || math.Abs(determinant) < 1e-18 {
			return OccultationPathPoint{}, false
		}
		delta := [2]float64{
			(-residual[0]*matrix[1][1] + matrix[0][1]*residual[1]) / determinant,
			(-matrix[0][0]*residual[1] + residual[0]*matrix[1][0]) / determinant,
		}
		if math.Abs(delta[0]) > 0.25 {
			delta[0] = math.Copysign(0.25, delta[0])
		}
		if math.Abs(delta[1]) > 5 {
			delta[1] = math.Copysign(5, delta[1])
		}
		coordinates[0] = riseSetNormalizeRadians(coordinates[0] + delta[0])
		coordinates[1] += delta[1]
	}
	residual, tt, longitude, latitude, ok := residualAt(coordinates)
	if !ok || math.Abs(residual[0]) > 1e-7 || math.Abs(residual[1]) > occultationRiseSetJunctionDerivativeTolerance {
		return OccultationPathPoint{}, false
	}
	state := cache.context(tt).stateAt(longitude, latitude)
	return OccultationPathPoint{
		Time: occultationTTToLocation(tt, location), Longitude: longitude,
		Latitude: latitude, MoonAltitude: state.moonAltitude,
	}, true
}

func refineOccultationRiseSetDirectionJunction(
	tt, longitude, latitude float64,
	greatest bool,
	location *time.Location,
	cache *occultationRiseSetEvaluationCache,
) (OccultationPathPoint, bool) {
	coordinates := [3]float64{longitude, latitude, tt}
	for iteration := 0; iteration < 24; iteration++ {
		evaluation := cache.evaluation(coordinates[2])
		residual, ok := occultationRiseSetDirectionJunctionResidualAt(evaluation, coordinates[0], coordinates[1], greatest)
		if !ok {
			return OccultationPathPoint{}, false
		}
		if math.Abs(residual[0]) <= 1e-8 && math.Abs(residual[1]) <= 1e-9 && math.Abs(residual[2]) <= 1e-7 {
			break
		}
		matrix, ok := occultationRiseSetDirectionJunctionJacobian(coordinates, residual, greatest, cache)
		if !ok {
			return OccultationPathPoint{}, false
		}
		delta, ok := solveSolarEclipse3x3(matrix, [3]float64{-residual[0], -residual[1], -residual[2]})
		if !ok {
			return OccultationPathPoint{}, false
		}
		occultationRiseSetLimitJunctionDelta(&delta)
		coordinates[0] = geodata.NormalizeLongitude(coordinates[0] + delta[0])
		coordinates[1] += delta[1]
		coordinates[2] += delta[2]
		if coordinates[1] <= -89.999999 || coordinates[1] >= 89.999999 {
			return OccultationPathPoint{}, false
		}
	}
	evaluation := cache.evaluation(coordinates[2])
	residual, ok := occultationRiseSetDirectionJunctionResidualAt(evaluation, coordinates[0], coordinates[1], greatest)
	if !ok || math.Abs(residual[0]) > 1e-7 || math.Abs(residual[1]) > 1e-8 || math.Abs(residual[2]) > 1e-7 {
		return OccultationPathPoint{}, false
	}
	state := evaluation.center.stateAt(coordinates[0], coordinates[1])
	return OccultationPathPoint{Time: occultationTTToLocation(coordinates[2], location), Longitude: coordinates[0], Latitude: coordinates[1], MoonAltitude: state.moonAltitude}, true
}

func refineOccultationRiseSetDirectionJunctionOnHorizon(
	first, second OccultationPathPoint,
	greatest bool,
	location *time.Location,
	cache *occultationRiseSetEvaluationCache,
) (OccultationPathPoint, bool) {
	seedTT := (occultationTimeToTT(first.Time) + occultationTimeToTT(second.Time)) / 2
	firstAngle := occultationRiseSetHorizonAngle(seedTT, first.Longitude, first.Latitude, cache.context)
	secondAngle := occultationRiseSetHorizonAngle(seedTT, second.Longitude, second.Latitude, cache.context)
	if !finite(firstAngle) || !finite(secondAngle) {
		return OccultationPathPoint{}, false
	}
	coordinates := [2]float64{
		riseSetNormalizeRadians(firstAngle + math.Remainder(secondAngle-firstAngle, 2*math.Pi)/2),
		0,
	}
	residualAt := func(value [2]float64) ([2]float64, float64, float64, float64, bool) {
		tt := seedTT + value[1]/1440
		longitude, latitude, ok := occultationRiseSetHorizonPoint(tt, value[0], cache.context)
		if !ok {
			return [2]float64{}, 0, 0, 0, false
		}
		evaluation := cache.evaluation(tt)
		phaseResidual, phaseOK := occultationRiseSetPhaseResidual(evaluation, longitude, latitude, greatest)
		altitudeDerivative := evaluation.moonAltitudeDerivative(longitude, latitude)
		return [2]float64{phaseResidual, altitudeDerivative}, tt, longitude, latitude,
			phaseOK && finite(phaseResidual) && finite(altitudeDerivative)
	}
	const (
		angleStep      = 1e-4
		timeStepMinute = 1.0 / 60
	)
	for iteration := 0; iteration < 32; iteration++ {
		residual, _, _, _, ok := residualAt(coordinates)
		if !ok {
			return OccultationPathPoint{}, false
		}
		if math.Abs(residual[0]) <= 1e-8 && math.Abs(residual[1]) <= 1e-7 {
			break
		}
		anglePlus, _, _, _, anglePlusOK := residualAt([2]float64{coordinates[0] + angleStep, coordinates[1]})
		angleMinus, _, _, _, angleMinusOK := residualAt([2]float64{coordinates[0] - angleStep, coordinates[1]})
		timePlus, _, _, _, timePlusOK := residualAt([2]float64{coordinates[0], coordinates[1] + timeStepMinute})
		timeMinus, _, _, _, timeMinusOK := residualAt([2]float64{coordinates[0], coordinates[1] - timeStepMinute})
		if !anglePlusOK || !angleMinusOK || !timePlusOK || !timeMinusOK {
			return OccultationPathPoint{}, false
		}
		matrix := [2][2]float64{
			{(anglePlus[0] - angleMinus[0]) / (2 * angleStep), (timePlus[0] - timeMinus[0]) / (2 * timeStepMinute)},
			{(anglePlus[1] - angleMinus[1]) / (2 * angleStep), (timePlus[1] - timeMinus[1]) / (2 * timeStepMinute)},
		}
		determinant := matrix[0][0]*matrix[1][1] - matrix[0][1]*matrix[1][0]
		if !finite(determinant) || math.Abs(determinant) < 1e-18 {
			return OccultationPathPoint{}, false
		}
		delta := [2]float64{
			(-residual[0]*matrix[1][1] + matrix[0][1]*residual[1]) / determinant,
			(-matrix[0][0]*residual[1] + residual[0]*matrix[1][0]) / determinant,
		}
		if math.Abs(delta[0]) > 0.25 {
			delta[0] = math.Copysign(0.25, delta[0])
		}
		if math.Abs(delta[1]) > 5 {
			delta[1] = math.Copysign(5, delta[1])
		}
		coordinates[0] = riseSetNormalizeRadians(coordinates[0] + delta[0])
		coordinates[1] += delta[1]
	}
	residual, tt, longitude, latitude, ok := residualAt(coordinates)
	if !ok || math.Abs(residual[0]) > 1e-7 || math.Abs(residual[1]) > 1e-7 {
		return OccultationPathPoint{}, false
	}
	state := cache.context(tt).stateAt(longitude, latitude)
	if !state.valid || math.Abs(state.moonAltitude) > 1e-7 {
		return OccultationPathPoint{}, false
	}
	return OccultationPathPoint{
		Time: occultationTTToLocation(tt, location), Longitude: longitude,
		Latitude: latitude, MoonAltitude: state.moonAltitude,
	}, true
}

func occultationRiseSetPhaseJunctionResidualAt(evaluation occultationRiseSetEvaluation, longitude, latitude float64) ([3]float64, bool) {
	state := evaluation.center.stateAt(longitude, latitude)
	residual := [3]float64{state.contactMetric, evaluation.contactDerivative(longitude, latitude), state.moonAltitude}
	return residual, state.valid && finite(residual[1]) && finite(residual[2])
}

func occultationRiseSetDirectionJunctionResidualAt(evaluation occultationRiseSetEvaluation, longitude, latitude float64, greatest bool) ([3]float64, bool) {
	phase, phaseOK := occultationRiseSetPhaseResidual(evaluation, longitude, latitude, greatest)
	state := evaluation.center.stateAt(longitude, latitude)
	residual := [3]float64{phase, state.moonAltitude, evaluation.moonAltitudeDerivative(longitude, latitude)}
	return residual, phaseOK && state.valid && finite(residual[1]) && finite(residual[2])
}

func occultationRiseSetPhaseResidual(evaluation occultationRiseSetEvaluation, longitude, latitude float64, greatest bool) (float64, bool) {
	if greatest {
		value := evaluation.separationDerivative(longitude, latitude)
		return value, finite(value)
	}
	state := evaluation.center.stateAt(longitude, latitude)
	return state.contactMetric, state.valid && finite(state.contactMetric)
}

func occultationRiseSetPhaseJunctionJacobian(
	coordinates [3]float64,
	residual [3]float64,
	cache *occultationRiseSetEvaluationCache,
) ([3][3]float64, bool) {
	const geographicStep = 1e-4
	const timeStep = 1.0 / 86400.0
	evaluation := cache.evaluation(coordinates[2])
	plusLongitude, longitudeOK := occultationRiseSetPhaseJunctionResidualAt(evaluation, coordinates[0]+geographicStep, coordinates[1])
	plusLatitude, latitudeOK := occultationRiseSetPhaseJunctionResidualAt(evaluation, coordinates[0], coordinates[1]+geographicStep)
	plusTimeEval := cache.evaluation(coordinates[2] + timeStep)
	plusTime, timeOK := occultationRiseSetPhaseJunctionResidualAt(plusTimeEval, coordinates[0], coordinates[1])
	if !longitudeOK || !latitudeOK || !timeOK {
		return [3][3]float64{}, false
	}
	var matrix [3][3]float64
	for row := range matrix {
		matrix[row][0] = (plusLongitude[row] - residual[row]) / geographicStep
		matrix[row][1] = (plusLatitude[row] - residual[row]) / geographicStep
		matrix[row][2] = (plusTime[row] - residual[row]) / timeStep
	}
	return matrix, true
}

func occultationRiseSetDirectionJunctionJacobian(
	coordinates [3]float64,
	residual [3]float64,
	greatest bool,
	cache *occultationRiseSetEvaluationCache,
) ([3][3]float64, bool) {
	const geographicStep = 1e-4
	const timeStep = 1.0 / 86400.0
	evaluation := cache.evaluation(coordinates[2])
	plusLongitude, longitudeOK := occultationRiseSetDirectionJunctionResidualAt(
		evaluation, coordinates[0]+geographicStep, coordinates[1], greatest,
	)
	plusLatitude, latitudeOK := occultationRiseSetDirectionJunctionResidualAt(
		evaluation, coordinates[0], coordinates[1]+geographicStep, greatest,
	)
	plusTimeEval := cache.evaluation(coordinates[2] + timeStep)
	plusTime, timeOK := occultationRiseSetDirectionJunctionResidualAt(
		plusTimeEval, coordinates[0], coordinates[1], greatest,
	)
	if !longitudeOK || !latitudeOK || !timeOK {
		return [3][3]float64{}, false
	}
	var matrix [3][3]float64
	for row := range matrix {
		matrix[row][0] = (plusLongitude[row] - residual[row]) / geographicStep
		matrix[row][1] = (plusLatitude[row] - residual[row]) / geographicStep
		matrix[row][2] = (plusTime[row] - residual[row]) / timeStep
	}
	return matrix, true
}

func occultationRiseSetLimitJunctionDelta(delta *[3]float64) {
	geographicScale := math.Max(math.Abs(delta[0]), math.Abs(delta[1]))
	if geographicScale > 2 {
		delta[0] *= 2 / geographicScale
		delta[1] *= 2 / geographicScale
	}
	if math.Abs(delta[2]) > 5.0/1440 {
		delta[2] = math.Copysign(5.0/1440, delta[2])
	}
}

func occultationRiseSetSegmentEndpoint(segment []OccultationPathPoint, atStart bool) OccultationPathPoint {
	if atStart {
		return segment[0]
	}
	return segment[len(segment)-1]
}

func occultationRiseSetAddEndpoint(segment []OccultationPathPoint, point OccultationPathPoint, atStart bool) []OccultationPathPoint {
	current := occultationRiseSetSegmentEndpoint(segment, atStart)
	if math.Abs(occultationTimeToTT(current.Time)-occultationTimeToTT(point.Time)) <= occultationRiseSetTimeEpsilonDays && occultationPathDistanceKM(current, point) <= 0.01 {
		if atStart {
			segment[0] = point
		} else {
			segment[len(segment)-1] = point
		}
		return segment
	}
	if atStart {
		return append([]OccultationPathPoint{point}, segment...)
	}
	return append(segment, point)
}
