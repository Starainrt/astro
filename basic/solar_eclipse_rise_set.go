package basic

import (
	"math"
	"sort"
)

func solarEclipsePathSampleTimes(startJDE, endJDE, greatestJDE, requestedStepDays float64) ([]float64, float64) {
	return solarEclipseMovingDiskEngine().sampleTimes(
		startJDE, endJDE, greatestJDE, requestedStepDays,
	)
}

type solarEclipseRiseSetCurveKey struct {
	phase     RiseSetPhase
	direction RiseSetDirection
}

type solarEclipseRiseSetTrack struct {
	segments [][]SolarEclipsePathPoint
}

type solarEclipseRiseSetEvaluation struct {
	jd     float64
	center localSolarEclipseStateContext
	before localSolarEclipseStateContext
	after  localSolarEclipseStateContext
}

type solarEclipseRiseSetSample struct {
	point SolarEclipsePathPoint
	key   solarEclipseRiseSetCurveKey
}

type solarEclipseRiseSetPhaseJunction struct {
	point     SolarEclipsePathPoint
	direction RiseSetDirection
}

type solarEclipseRiseSetSamplePair struct {
	first  solarEclipseRiseSetSample
	second solarEclipseRiseSetSample
}

func (solver solarEclipseSolver) riseSetCurves(
	startJDE, endJDE, greatestJDE, requestedStepDays float64,
) []SolarEclipseRiseSetCurve {
	curves, _ := solver.riseSetCurvesWithStatus(startJDE, endJDE, greatestJDE, requestedStepDays)
	return curves
}

// riseSetCurvesWithStatus 额外报告六类边界的拓扑校验结果：段数超过每曲线 16 段或用尽
// 32 段总预算时返回的是截断结果，调用方必须显式标记降级而不是当作完整拓扑。
func (solver solarEclipseSolver) riseSetCurvesWithStatus(
	startJDE, endJDE, greatestJDE, requestedStepDays float64,
) ([]SolarEclipseRiseSetCurve, bool) {
	if startJDE == 0 || endJDE == 0 || endJDE <= startJDE {
		return nil, true
	}
	solver = solver.withLocalEphemeris()
	traceStepDays := requestedStepDays
	curves, phaseJunctions, traceStepDays := solver.sampleRiseSetCurves(
		startJDE, endJDE, greatestJDE, traceStepDays,
	)
	topologyStepDays := math.Max(traceStepDays, solarEclipseRiseSetCriticalStepDays)
	// The sampled horizon already contains the complete six-key topology for
	// ordinary events. Running the continuation tracer in that case can split a
	// polar branch into hundreds of tiny components when its tangent changes
	// sign near a fold. Close the sampled endpoints first; only invoke the
	// expensive continuation path if the bounded raw topology cannot be closed.
	if solarEclipseRiseSetRawTopologyUsable(curves) {
		solver.finalizeRiseSetCurveTopology(curves, topologyStepDays, phaseJunctions)
		if solarEclipseRiseSetCurveTopologyComplete(curves) {
			if requestedStepDays > traceStepDays {
				decimateSolarEclipseRiseSetCurves(curves, requestedStepDays)
			}
			return curves, true
		}
	}
	traced := solver.traceRiseSetCurveTopology(curves, startJDE, endJDE, greatestJDE)
	if len(traced) == 6 {
		solver.finalizeRiseSetCurveTopology(traced, topologyStepDays, phaseJunctions)
	}
	if solarEclipseRiseSetCurveTopologyComplete(traced) {
		curves = traced
	} else {
		for _, retryStepDays := range []float64{20.0 / 86400.0, 5.0 / 86400.0} {
			if retryStepDays >= traceStepDays {
				continue
			}
			curves, phaseJunctions, traceStepDays = solver.sampleRiseSetCurves(
				startJDE, endJDE, greatestJDE, retryStepDays,
			)
			topologyStepDays = math.Max(traceStepDays, solarEclipseRiseSetCriticalStepDays)
			retryTraced := solver.traceRiseSetCurveTopology(curves, startJDE, endJDE, greatestJDE)
			if len(retryTraced) == 6 {
				solver.finalizeRiseSetCurveTopology(retryTraced, topologyStepDays, phaseJunctions)
			}
			if solarEclipseRiseSetCurveTopologyComplete(retryTraced) {
				curves = retryTraced
				break
			}
			solver.finalizeRiseSetCurveTopology(curves, topologyStepDays, phaseJunctions)
			if solarEclipseRiseSetCurveTopologyComplete(curves) {
				break
			}
		}
		if !solarEclipseRiseSetCurveTopologyComplete(curves) {
			solver.finalizeRiseSetCurveTopology(curves, topologyStepDays, phaseJunctions)
		}
	}
	if requestedStepDays > traceStepDays {
		decimateSolarEclipseRiseSetCurves(curves, requestedStepDays)
	}
	return curves, solarEclipseRiseSetCurveTopologyComplete(curves)
}

func solarEclipseRiseSetRawTopologyUsable(curves []SolarEclipseRiseSetCurve) bool {
	if len(curves) != 6 {
		return false
	}
	seen := make(map[solarEclipseRiseSetCurveKey]bool, 6)
	segments := 0
	for _, curve := range curves {
		key := solarEclipseRiseSetCurveKey{phase: curve.Phase, direction: curve.Direction}
		if seen[key] || len(curve.Segments) == 0 {
			return false
		}
		seen[key] = true
		segments += len(curve.Segments)
	}
	return segments <= 32
}

func (solver solarEclipseSolver) sampleRiseSetCurves(
	startJDE, endJDE, greatestJDE, traceStepDays float64,
) ([]SolarEclipseRiseSetCurve, []solarEclipseRiseSetPhaseJunction, float64) {
	times, traceStepDays := solarEclipsePathSampleTimes(startJDE, endJDE, greatestJDE, traceStepDays)
	keys := []solarEclipseRiseSetCurveKey{
		{RiseSetPhaseStart, RiseSetDirectionRise},
		{RiseSetPhaseStart, RiseSetDirectionSet},
		{RiseSetPhaseGreatest, RiseSetDirectionRise},
		{RiseSetPhaseGreatest, RiseSetDirectionSet},
		{RiseSetPhaseEnd, RiseSetDirectionRise},
		{RiseSetPhaseEnd, RiseSetDirectionSet},
	}
	tracks := make(map[solarEclipseRiseSetCurveKey][]*solarEclipseRiseSetTrack, len(keys))
	var previousContactSamples []solarEclipseRiseSetSample
	var phaseJunctions []solarEclipseRiseSetPhaseJunction
	for _, jd := range times {
		pointsAt := solver.riseSetCandidatePointsAt(jd, solarEclipseRiseSetBoundaryPoints)
		contactSamples := solarEclipseRiseSetContactSamples(pointsAt)
		phaseJunctions = solver.appendRiseSetPhaseJunctionsAtSamples(
			phaseJunctions, contactSamples, traceStepDays,
		)
		phaseJunctions = solver.appendRiseSetPhaseJunctionsBetweenSamples(
			phaseJunctions, previousContactSamples, contactSamples,
		)
		previousContactSamples = contactSamples
		for _, key := range keys {
			tracks[key] = appendSolarEclipseRiseSetSamples(tracks[key], pointsAt[key], traceStepDays)
		}
	}

	curves := make([]SolarEclipseRiseSetCurve, 0, len(keys))
	for _, key := range keys {
		segments := make([][]SolarEclipsePathPoint, 0, len(tracks[key]))
		for _, track := range tracks[key] {
			for _, segment := range track.segments {
				if len(segment) >= 2 {
					segments = append(segments, segment)
				}
			}
		}
		if len(segments) == 0 {
			continue
		}
		curves = append(curves, SolarEclipseRiseSetCurve{
			Phase: key.phase, Direction: key.direction, Segments: segments,
		})
	}
	return curves, phaseJunctions, traceStepDays
}

func (solver solarEclipseSolver) finalizeRiseSetCurveTopology(
	curves []SolarEclipseRiseSetCurve,
	topologyStepDays float64,
	phaseJunctions []solarEclipseRiseSetPhaseJunction,
) {
	solver.completeRiseSetCurveEndpoints(curves, topologyStepDays, phaseJunctions)
	snapSolarEclipseRiseSetArcPhaseJunctions(curves, phaseJunctions)
	solver.closeSolarEclipseRiseSetArcFolds(curves, phaseJunctions)
	solver.snapNearCoincidentSolarEclipseRiseSetEndpoints(curves)
	solver.closeSolarEclipseRiseSetArcDirectionJunctions(curves, phaseJunctions)
	deduplicateSolarEclipseRiseSetArcSegments(curves)
	sortSolarEclipseRiseSetSegments(curves)
}

func decimateSolarEclipseRiseSetCurves(
	curves []SolarEclipseRiseSetCurve,
	stepDays float64,
) {
	for curveIndex := range curves {
		for segmentIndex, segment := range curves[curveIndex].Segments {
			if len(segment) < 3 {
				continue
			}
			decimated := make([]SolarEclipsePathPoint, 1, len(segment))
			decimated[0] = segment[0]
			lastIndex := 0
			for pointIndex := 1; pointIndex < len(segment)-1; pointIndex++ {
				last := decimated[len(decimated)-1]
				next := segment[pointIndex+1]
				if next.JDE-last.JDE <= stepDays+solarEclipseRiseSetTimeEpsilonDays &&
					solarEclipsePathDistanceKM(last, next) <= solarEclipseRiseSetTargetSpacingKM {
					accurate := true
					for _, point := range segment[lastIndex+1 : pointIndex+1] {
						if solarEclipseRiseSetChordDeviationKM(point, last, next) > solarEclipseRiseSetChordToleranceKM {
							accurate = false
							break
						}
					}
					if accurate {
						continue
					}
				}
				decimated = append(decimated, segment[pointIndex])
				lastIndex = pointIndex
			}
			decimated = append(decimated, segment[len(segment)-1])
			curves[curveIndex].Segments[segmentIndex] = decimated
		}
	}
}

func solarEclipseRiseSetContactSamples(
	pointsAt map[solarEclipseRiseSetCurveKey][]SolarEclipsePathPoint,
) []solarEclipseRiseSetSample {
	var samples []solarEclipseRiseSetSample
	for _, phase := range []RiseSetPhase{RiseSetPhaseStart, RiseSetPhaseEnd} {
		for _, direction := range []RiseSetDirection{RiseSetDirectionRise, RiseSetDirectionSet} {
			key := solarEclipseRiseSetCurveKey{phase: phase, direction: direction}
			for _, point := range pointsAt[key] {
				samples = append(samples, solarEclipseRiseSetSample{point: point, key: key})
			}
		}
	}
	return samples
}

func (solver solarEclipseSolver) appendRiseSetPhaseJunctionsAtSamples(
	junctions []solarEclipseRiseSetPhaseJunction,
	samples []solarEclipseRiseSetSample,
	stepDays float64,
) []solarEclipseRiseSetPhaseJunction {
	for firstIndex, first := range samples {
		for secondIndex := firstIndex + 1; secondIndex < len(samples); secondIndex++ {
			second := samples[secondIndex]
			if solarEclipsePathDistanceKM(first.point, second.point) > 1500 {
				continue
			}
			candidate, ok := SolarEclipsePathPoint{}, false
			for _, seed := range []SolarEclipsePathPoint{
				solarEclipseRiseSetMidpoint(first.point, second.point),
				first.point,
				second.point,
			} {
				candidate, ok = solver.refineRiseSetPhaseJunctionOnHorizon(seed)
				if !ok {
					candidate, ok = solver.refineRiseSetPhaseJunction(
						seed.JDE, seed.Longitude, seed.Latitude,
					)
				}
				if ok {
					break
				}
			}
			if !ok || math.Abs(candidate.JDE-first.point.JDE) > stepDays ||
				solarEclipsePathDistanceKM(candidate, first.point) > 3000 ||
				solarEclipsePathDistanceKM(candidate, second.point) > 3000 {
				continue
			}
			junctions = solver.appendRiseSetPhaseJunction(junctions, candidate)
		}
	}
	return junctions
}

func (solver solarEclipseSolver) appendRiseSetPhaseJunction(
	junctions []solarEclipseRiseSetPhaseJunction,
	candidate SolarEclipsePathPoint,
) []solarEclipseRiseSetPhaseJunction {
	evaluation := solver.magnitudeEvaluationAt(candidate.JDE)
	altitudeDerivative := evaluation.sunAltitudeDerivative(candidate.Longitude, candidate.Latitude)
	if !finite(altitudeDerivative) || math.Abs(altitudeDerivative) < 1e-10 {
		return junctions
	}
	direction := RiseSetDirectionSet
	if altitudeDerivative > 0 {
		direction = RiseSetDirectionRise
	}
	for _, junction := range junctions {
		if math.Abs(junction.point.JDE-candidate.JDE) <= solarEclipseRiseSetTimeEpsilonDays &&
			solarEclipsePathDistanceKM(junction.point, candidate) <= 0.01 {
			return junctions
		}
	}
	return append(junctions, solarEclipseRiseSetPhaseJunction{
		point: candidate, direction: direction,
	})
}

func (solver solarEclipseSolver) appendRiseSetPhaseJunctionsBetweenSamples(
	junctions []solarEclipseRiseSetPhaseJunction,
	previous, current []solarEclipseRiseSetSample,
) []solarEclipseRiseSetPhaseJunction {
	if len(previous) == 0 || len(current) == 0 {
		return junctions
	}
	for _, pair := range solarEclipseMatchRiseSetSamples(previous, current) {
		if pair.first.key.phase == pair.second.key.phase {
			continue
		}
		candidate, ok := solver.refineRiseSetPhaseJunctionBetweenSamples(pair.first, pair.second)
		if !ok {
			continue
		}
		if candidate.JDE < pair.first.point.JDE-solarEclipseRiseSetTimeEpsilonDays ||
			candidate.JDE > pair.second.point.JDE+solarEclipseRiseSetTimeEpsilonDays ||
			solarEclipsePathDistanceKM(candidate, pair.first.point) > 6000 ||
			solarEclipsePathDistanceKM(candidate, pair.second.point) > 6000 {
			continue
		}
		junctions = solver.appendRiseSetPhaseJunction(junctions, candidate)
	}
	return junctions
}

func solarEclipseMatchRiseSetSamples(
	previous, current []solarEclipseRiseSetSample,
) []solarEclipseRiseSetSamplePair {
	type candidatePair struct {
		firstIndex  int
		secondIndex int
		distance    float64
	}
	var candidates []candidatePair
	for firstIndex, first := range previous {
		for secondIndex, second := range current {
			deltaDays := second.point.JDE - first.point.JDE
			distance := solarEclipsePathDistanceKM(first.point, second.point)
			if deltaDays <= 0 || riseSetGeographicBranchChanged(distance, deltaDays) {
				continue
			}
			candidates = append(candidates, candidatePair{
				firstIndex: firstIndex, secondIndex: secondIndex, distance: distance,
			})
		}
	}
	sort.Slice(candidates, func(first, second int) bool {
		return candidates[first].distance < candidates[second].distance
	})
	usedFirst := make([]bool, len(previous))
	usedSecond := make([]bool, len(current))
	var pairs []solarEclipseRiseSetSamplePair
	for _, candidate := range candidates {
		if usedFirst[candidate.firstIndex] || usedSecond[candidate.secondIndex] {
			continue
		}
		usedFirst[candidate.firstIndex] = true
		usedSecond[candidate.secondIndex] = true
		pairs = append(pairs, solarEclipseRiseSetSamplePair{
			first: previous[candidate.firstIndex], second: current[candidate.secondIndex],
		})
	}
	return pairs
}

func (solver solarEclipseSolver) refineRiseSetPhaseJunctionBetweenSamples(
	first, second solarEclipseRiseSetSample,
) (SolarEclipsePathPoint, bool) {
	left, right := first, second
	for iteration := 0; iteration < 8; iteration++ {
		midpoint := solarEclipseRiseSetMidpoint(left.point, right.point)
		midpoint.JDE = (left.point.JDE + right.point.JDE) / 2
		current := solarEclipseRiseSetContactSamples(
			solver.riseSetPointsAt(midpoint.JDE, solarEclipseRiseSetBoundaryPoints),
		)
		bestIndex := -1
		bestDistance := math.Inf(1)
		for index, sample := range current {
			distance := solarEclipsePathDistanceKM(midpoint, sample.point)
			if distance < bestDistance {
				bestIndex, bestDistance = index, distance
			}
		}
		if bestIndex < 0 || bestDistance > 1000 {
			break
		}
		if current[bestIndex].key.phase == left.key.phase {
			left = current[bestIndex]
		} else if current[bestIndex].key.phase == right.key.phase {
			right = current[bestIndex]
		} else {
			break
		}
	}
	for _, seed := range []SolarEclipsePathPoint{
		solarEclipseRiseSetMidpoint(left.point, right.point),
		left.point,
		right.point,
	} {
		candidate, ok := solver.refineRiseSetPhaseJunctionOnHorizon(seed)
		if !ok {
			candidate, ok = solver.refineRiseSetPhaseJunction(
				seed.JDE, seed.Longitude, seed.Latitude,
			)
		}
		if ok {
			return candidate, true
		}
	}
	return SolarEclipsePathPoint{}, false
}

func (solver solarEclipseSolver) riseSetPointsAt(
	jde float64,
	boundaryPoints int,
) map[solarEclipseRiseSetCurveKey][]SolarEclipsePathPoint {
	return solver.riseSetPointsAtEvaluation(jde, boundaryPoints, solver.magnitudeEvaluationAt(jde))
}

func (solver solarEclipseSolver) riseSetCandidatePointsAt(
	jde float64,
	boundaryPoints int,
) map[solarEclipseRiseSetCurveKey][]SolarEclipsePathPoint {
	return solver.riseSetPointsAtEvaluation(jde, boundaryPoints, solver.magnitudeCandidateEvaluationAt(jde))
}

func (solver solarEclipseSolver) riseSetPointsAtEvaluation(
	jd float64,
	boundaryPoints int,
	evaluation solarEclipseRiseSetEvaluation,
) map[solarEclipseRiseSetCurveKey][]SolarEclipsePathPoint {
	result := make(map[solarEclipseRiseSetCurveKey][]SolarEclipsePathPoint, 6)
	sun := solarEclipseXYZToLLR(
		evaluation.center.sunXYZ[0], evaluation.center.sunXYZ[1], evaluation.center.sunXYZ[2],
	)
	centerLongitude := normalizeLongitude((sun[0] - evaluation.center.gst) / rad)
	centerLatitude := sun[1] / rad

	appendRoots := func(greatest bool) {
		valueAt := func(angle float64) (float64, bool) {
			longitude, latitude := riseSetHorizonPoint(centerLongitude, centerLatitude, angle)
			state := evaluation.center.stateAt(longitude*rad, latitude*rad, 0)
			if greatest {
				value := evaluation.separationDerivative(longitude, latitude)
				return value, finite(value)
			}
			value := solarEclipsePartialContactGap(state)
			return value, finite(value)
		}
		for _, angle := range riseSetCyclicRoots(boundaryPoints, valueAt) {
			longitude, latitude := riseSetHorizonPoint(centerLongitude, centerLatitude, angle)
			longitude, latitude, ok := riseSetRefineGeographicRoot(
				longitude,
				latitude,
				func(lon, lat float64) (float64, float64, bool) {
					state := evaluation.center.stateAt(lon*rad, lat*rad, 0)
					first := solarEclipsePartialContactGap(state)
					if greatest {
						first = evaluation.separationDerivative(lon, lat)
					}
					return first, state.sunAltitudeRad, finite(first) && finite(state.sunAltitudeRad)
				},
			)
			if !ok {
				continue
			}
			point, key, valid := evaluation.classify(longitude, latitude, greatest)
			if !valid || solarEclipseRiseSetPointExists(result[key], point) {
				continue
			}
			result[key] = append(result[key], point)
		}
	}
	appendRoots(false)
	appendRoots(true)
	return result
}

func (evaluation solarEclipseRiseSetEvaluation) classify(
	longitude, latitude float64,
	greatest bool,
) (SolarEclipsePathPoint, solarEclipseRiseSetCurveKey, bool) {
	state := evaluation.center.stateAt(longitude*rad, latitude*rad, 0)
	altitudeDerivative := evaluation.sunAltitudeDerivative(longitude, latitude)
	if !finite(altitudeDerivative) || math.Abs(altitudeDerivative) < 1e-10 {
		return SolarEclipsePathPoint{}, solarEclipseRiseSetCurveKey{}, false
	}
	direction := RiseSetDirectionSet
	if altitudeDerivative > 0 {
		direction = RiseSetDirectionRise
	}
	phase := RiseSetPhaseGreatest
	if greatest {
		if solarEclipsePartialContactGap(state) > 1e-7 ||
			evaluation.separationSecondDerivative(longitude, latitude) <= 0 {
			return SolarEclipsePathPoint{}, solarEclipseRiseSetCurveKey{}, false
		}
	} else {
		contactDerivative := evaluation.partialContactDerivative(longitude, latitude)
		if !finite(contactDerivative) || math.Abs(contactDerivative) < 1e-10 {
			return SolarEclipsePathPoint{}, solarEclipseRiseSetCurveKey{}, false
		}
		phase = RiseSetPhaseEnd
		if contactDerivative < 0 {
			phase = RiseSetPhaseStart
		}
	}
	return SolarEclipsePathPoint{
		JDE: evaluation.jd, Longitude: longitude, Latitude: latitude, SunAltitude: state.sunAltitudeRad / rad,
	}, solarEclipseRiseSetCurveKey{phase: phase, direction: direction}, true
}

func solarEclipsePartialContactGap(state localSolarEclipseState) float64 {
	return state.movingDiskContactState().externalContactGap()
}

func (evaluation solarEclipseRiseSetEvaluation) partialContactDerivative(longitude, latitude float64) float64 {
	before := evaluation.before.stateAt(longitude*rad, latitude*rad, 0)
	after := evaluation.after.stateAt(longitude*rad, latitude*rad, 0)
	return (solarEclipsePartialContactGap(after) - solarEclipsePartialContactGap(before)) /
		(2 * solarEclipseRiseSetDerivativeStepDays)
}

func (evaluation solarEclipseRiseSetEvaluation) partialContactSecondDerivative(longitude, latitude float64) float64 {
	before := evaluation.before.stateAt(longitude*rad, latitude*rad, 0)
	center := evaluation.center.stateAt(longitude*rad, latitude*rad, 0)
	after := evaluation.after.stateAt(longitude*rad, latitude*rad, 0)
	stepSquared := solarEclipseRiseSetDerivativeStepDays * solarEclipseRiseSetDerivativeStepDays
	return (solarEclipsePartialContactGap(after) - 2*solarEclipsePartialContactGap(center) + solarEclipsePartialContactGap(before)) /
		stepSquared
}

func (evaluation solarEclipseRiseSetEvaluation) sunAltitudeDerivative(longitude, latitude float64) float64 {
	before := evaluation.before.stateAt(longitude*rad, latitude*rad, 0).sunAltitudeRad
	after := evaluation.after.stateAt(longitude*rad, latitude*rad, 0).sunAltitudeRad
	return (after - before) / (2 * solarEclipseRiseSetDerivativeStepDays)
}

func (evaluation solarEclipseRiseSetEvaluation) sunAltitudeSecondDerivative(longitude, latitude float64) float64 {
	before := evaluation.before.stateAt(longitude*rad, latitude*rad, 0).sunAltitudeRad
	center := evaluation.center.stateAt(longitude*rad, latitude*rad, 0).sunAltitudeRad
	after := evaluation.after.stateAt(longitude*rad, latitude*rad, 0).sunAltitudeRad
	stepSquared := solarEclipseRiseSetDerivativeStepDays * solarEclipseRiseSetDerivativeStepDays
	return (after - 2*center + before) / stepSquared
}

func (evaluation solarEclipseRiseSetEvaluation) separationDerivative(longitude, latitude float64) float64 {
	before := evaluation.before.stateAt(longitude*rad, latitude*rad, 0).separationSquared
	after := evaluation.after.stateAt(longitude*rad, latitude*rad, 0).separationSquared
	return (after - before) / (2 * solarEclipseRiseSetDerivativeStepDays)
}

func (evaluation solarEclipseRiseSetEvaluation) separationSecondDerivative(longitude, latitude float64) float64 {
	before := evaluation.before.stateAt(longitude*rad, latitude*rad, 0).separationSquared
	center := evaluation.center.stateAt(longitude*rad, latitude*rad, 0).separationSquared
	after := evaluation.after.stateAt(longitude*rad, latitude*rad, 0).separationSquared
	stepSquared := solarEclipseRiseSetDerivativeStepDays * solarEclipseRiseSetDerivativeStepDays
	return (after - 2*center + before) / stepSquared
}

func solarEclipseRiseSetPointExists(points []SolarEclipsePathPoint, candidate SolarEclipsePathPoint) bool {
	for _, point := range points {
		if solarEclipsePathDistanceKM(point, candidate) < 0.01 {
			return true
		}
	}
	return false
}

func appendSolarEclipseRiseSetSamples(
	tracks []*solarEclipseRiseSetTrack,
	points []SolarEclipsePathPoint,
	stepDays float64,
) []*solarEclipseRiseSetTrack {
	used := make([]bool, len(tracks))
	for _, point := range points {
		bestTrack := -1
		bestDistance := math.Inf(1)
		for index, track := range tracks {
			if used[index] || len(track.segments) == 0 || len(track.segments[len(track.segments)-1]) == 0 {
				continue
			}
			last := track.segments[len(track.segments)-1][len(track.segments[len(track.segments)-1])-1]
			deltaDays := point.JDE - last.JDE
			if deltaDays <= 0 || deltaDays > 2.5*stepDays {
				continue
			}
			distance := solarEclipsePathDistanceKM(last, point)
			if riseSetGeographicBranchChanged(distance, deltaDays) || distance >= bestDistance {
				continue
			}
			bestTrack, bestDistance = index, distance
		}
		if bestTrack < 0 {
			tracks = append(tracks, &solarEclipseRiseSetTrack{segments: [][]SolarEclipsePathPoint{{point}}})
			used = append(used, true)
			continue
		}
		track := tracks[bestTrack]
		track.segments[len(track.segments)-1] = append(track.segments[len(track.segments)-1], point)
		used[bestTrack] = true
	}
	return tracks
}

func riseSetGeographicBranchChanged(distanceKM, deltaDays float64) bool {
	if distanceKM <= 750 {
		return false
	}
	return deltaDays <= 0 || distanceKM/(deltaDays*86400) > 10
}
