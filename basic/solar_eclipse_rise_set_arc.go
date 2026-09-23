package basic

import (
	"math"
	"sort"
)

const (
	solarEclipseRiseSetArcTimeScale     = 360.0
	solarEclipseRiseSetArcInitialStep   = 4.0
	solarEclipseRiseSetArcMinimumStep   = 0.0025
	solarEclipseRiseSetArcMaximumSteps  = 8000
	solarEclipseRiseSetArcTargetSpacing = 500.0
	solarEclipseRiseSetArcCloseDistance = 550.0
	solarEclipseRiseSetArcCloseTimeDays = 5.0 / 1440.0
	solarEclipseRiseSetArcSeedDistance  = 550.0
	solarEclipseRiseSetArcSeedTimeDays  = 30.0 / 1440.0
)

type solarEclipseRiseSetArcState struct {
	coordinates [3]float64
	tangent     [3]float64
	point       SolarEclipsePathPoint
}

type solarEclipseRiseSetArcSegment struct {
	key    solarEclipseRiseSetCurveKey
	points []SolarEclipsePathPoint
}

type solarEclipseRiseSetArcTransition uint8

const (
	solarEclipseRiseSetArcNoTransition solarEclipseRiseSetArcTransition = iota
	solarEclipseRiseSetArcPhaseTransition
	solarEclipseRiseSetArcDirectionTransition
	solarEclipseRiseSetArcFoldTransition
)

func solarEclipseRiseSetCurveTopologyComplete(curves []SolarEclipseRiseSetCurve) bool {
	if len(curves) != 6 {
		return false
	}
	keys := make(map[solarEclipseRiseSetCurveKey]bool, 6)
	for curveIndex, curve := range curves {
		key := solarEclipseRiseSetCurveKey{phase: curve.Phase, direction: curve.Direction}
		if keys[key] || len(curve.Segments) == 0 || len(curve.Segments) > 16 {
			return false
		}
		keys[key] = true
		for segmentIndex, segment := range curve.Segments {
			if len(segment) < 2 {
				return false
			}
			for pointIndex := 1; pointIndex < len(segment); pointIndex++ {
				if segment[pointIndex].JDE <= segment[pointIndex-1].JDE+solarEclipseRiseSetTimeEpsilonDays {
					return false
				}
			}
			for _, pointIndex := range []int{0, len(segment) - 1} {
				if !solarEclipseRiseSetArcEndpointShared(curves, curveIndex, segmentIndex, pointIndex) {
					return false
				}
			}
		}
	}
	return len(keys) == 6
}

func (solver solarEclipseSolver) traceRiseSetCurveTopology(
	seedCurves []SolarEclipseRiseSetCurve,
	startJDE, endJDE, referenceJDE float64,
) []SolarEclipseRiseSetCurve {
	if len(seedCurves) == 0 {
		return nil
	}
	type component struct {
		greatest bool
		states   []solarEclipseRiseSetArcState
	}
	var components []component
	for _, curve := range seedCurves {
		greatest := curve.Phase == RiseSetPhaseGreatest
		for _, segment := range curve.Segments {
			if len(segment) < 2 {
				continue
			}
			seed := segment[len(segment)/2]
			covered := false
			for _, existing := range components {
				if existing.greatest != greatest || !solarEclipseRiseSetArcContainsSeed(existing.states, seed) {
					continue
				}
				covered = true
				break
			}
			if covered {
				continue
			}
			states := solver.traceRiseSetArcComponent(seed, greatest, referenceJDE, startJDE, endJDE)
			if len(states) < 3 {
				continue
			}
			components = append(components, component{greatest: greatest, states: states})
		}
	}

	var traced []solarEclipseRiseSetArcSegment
	for _, component := range components {
		traced = append(traced, solver.splitRiseSetArcComponent(component.states, component.greatest)...)
	}
	if len(traced) == 0 {
		return nil
	}
	curveSegments := make(map[solarEclipseRiseSetCurveKey][][]SolarEclipsePathPoint, 6)
	for _, segment := range traced {
		if len(segment.points) < 2 {
			continue
		}
		points := segment.points
		if points[0].JDE > points[len(points)-1].JDE {
			points = append([]SolarEclipsePathPoint(nil), points...)
			for left, right := 0, len(points)-1; left < right; left, right = left+1, right-1 {
				points[left], points[right] = points[right], points[left]
			}
		}
		curveSegments[segment.key] = append(curveSegments[segment.key], points)
	}
	keys := []solarEclipseRiseSetCurveKey{
		{RiseSetPhaseStart, RiseSetDirectionRise},
		{RiseSetPhaseStart, RiseSetDirectionSet},
		{RiseSetPhaseGreatest, RiseSetDirectionRise},
		{RiseSetPhaseGreatest, RiseSetDirectionSet},
		{RiseSetPhaseEnd, RiseSetDirectionRise},
		{RiseSetPhaseEnd, RiseSetDirectionSet},
	}
	curves := make([]SolarEclipseRiseSetCurve, 0, len(keys))
	for _, key := range keys {
		segments := curveSegments[key]
		if len(segments) == 0 {
			continue
		}
		curve := SolarEclipseRiseSetCurve{Phase: key.phase, Direction: key.direction, Segments: segments}
		normalizeSolarEclipseRiseSetCurveSegments(&curve)
		mergeSolarEclipseRiseSetArcContinuations(&curve)
		if len(curve.Segments) > 0 {
			curves = append(curves, curve)
		}
	}
	return curves
}

func mergeSolarEclipseRiseSetArcContinuations(curve *SolarEclipseRiseSetCurve) {
	if curve == nil || len(curve.Segments) < 2 {
		return
	}
	for {
		merged := false
		for firstIndex := 0; firstIndex < len(curve.Segments) && !merged; firstIndex++ {
			first := curve.Segments[firstIndex]
			if len(first) < 2 {
				continue
			}
			for secondIndex := 0; secondIndex < len(curve.Segments); secondIndex++ {
				if firstIndex == secondIndex {
					continue
				}
				second := curve.Segments[secondIndex]
				if len(second) < 2 ||
					math.Abs(second[0].JDE-first[len(first)-1].JDE) > solarEclipseRiseSetAttachmentTimeToleranceDays ||
					solarEclipsePathDistanceKM(first[len(first)-1], second[0]) > solarEclipseRiseSetAttachmentDistanceToleranceKM {
					continue
				}
				joined := make([]SolarEclipsePathPoint, 0, len(first)+len(second))
				switch {
				case second[1].JDE > first[len(first)-1].JDE+solarEclipseRiseSetTimeEpsilonDays:
					joined = append(joined, first...)
					joined = append(joined, second[1:]...)
				case second[0].JDE > first[len(first)-2].JDE+solarEclipseRiseSetTimeEpsilonDays:
					joined = append(joined, first[:len(first)-1]...)
					joined = append(joined, second...)
				default:
					continue
				}
				curve.Segments[firstIndex] = joined
				curve.Segments = append(curve.Segments[:secondIndex], curve.Segments[secondIndex+1:]...)
				merged = true
				break
			}
		}
		if !merged {
			return
		}
	}
}

func snapSolarEclipseRiseSetArcPhaseJunctions(
	curves []SolarEclipseRiseSetCurve,
	junctions []solarEclipseRiseSetPhaseJunction,
) {
	const (
		maximumTimeDays = 1.0 / 1440.0
		maximumDistance = 1000.0
	)
	for _, junction := range junctions {
		for curveIndex := range curves {
			curve := &curves[curveIndex]
			if curve.Direction != junction.direction {
				continue
			}
			if solarEclipseRiseSetArcPointInCurve(junction.point, *curve) {
				continue
			}
			bestSegment, bestPoint := -1, -1
			bestMetric := math.Inf(1)
			for segmentIndex, segment := range curve.Segments {
				if len(segment) < 2 {
					continue
				}
				for _, pointIndex := range []int{0, len(segment) - 1} {
					if solarEclipseRiseSetArcEndpointShared(curves, curveIndex, segmentIndex, pointIndex) {
						continue
					}
					point := segment[pointIndex]
					deltaDays := math.Abs(point.JDE - junction.point.JDE)
					distance := solarEclipsePathDistanceKM(point, junction.point)
					if deltaDays > maximumTimeDays || distance > maximumDistance {
						continue
					}
					metric := distance + deltaDays*8640
					if metric < bestMetric {
						bestSegment, bestPoint, bestMetric = segmentIndex, pointIndex, metric
					}
				}
			}
			if bestSegment >= 0 {
				curve.Segments[bestSegment][bestPoint] = junction.point
			}
		}
	}
	for curveIndex := range curves {
		normalizeSolarEclipseRiseSetCurveSegments(&curves[curveIndex])
		mergeSolarEclipseRiseSetArcContinuations(&curves[curveIndex])
	}
}

func (solver solarEclipseSolver) closeSolarEclipseRiseSetArcFolds(
	curves []SolarEclipseRiseSetCurve,
	junctions []solarEclipseRiseSetPhaseJunction,
) {
	const (
		maximumFoldTimeDays = 2.0 / 1440.0
		maximumFoldDistance = 1500.0
		maximumBridgeDays   = 5.0 / 1440.0
		maximumBridgeKM     = 3000.0
	)
	for curveIndex := range curves {
		curve := &curves[curveIndex]
		greatest := curve.Phase == RiseSetPhaseGreatest
		for segmentIndex, segment := range curve.Segments {
			if len(segment) < 2 {
				continue
			}
			for _, pointIndex := range []int{0, len(segment) - 1} {
				endpoint := curve.Segments[segmentIndex][pointIndex]
				if solarEclipseRiseSetArcEndpointShared(curves, curveIndex, segmentIndex, pointIndex) {
					continue
				}
				fold, ok := solver.refineRiseSetFoldPoint(
					endpoint.JDE, endpoint.Longitude, endpoint.Latitude, greatest,
				)
				if !ok {
					for _, junction := range junctions {
						if junction.direction != curve.Direction ||
							math.Abs(junction.point.JDE-endpoint.JDE) > maximumBridgeDays ||
							solarEclipsePathDistanceKM(junction.point, endpoint) > maximumBridgeKM {
							continue
						}
						fold, ok = solver.refineRiseSetFold(junction.point, endpoint, greatest)
						if ok {
							break
						}
					}
				}
				if !ok || math.Abs(fold.JDE-endpoint.JDE) > maximumFoldTimeDays ||
					solarEclipsePathDistanceKM(fold, endpoint) > maximumFoldDistance {
					continue
				}
				curve.Segments[segmentIndex][pointIndex] = fold
				bestJunction := SolarEclipsePathPoint{}
				bestMetric := math.Inf(1)
				for _, junction := range junctions {
					if junction.direction != curve.Direction ||
						solarEclipseRiseSetArcPointInCurve(junction.point, *curve) {
						continue
					}
					deltaDays := math.Abs(junction.point.JDE - fold.JDE)
					distance := solarEclipsePathDistanceKM(junction.point, fold)
					if deltaDays > maximumBridgeDays || distance > maximumBridgeKM {
						continue
					}
					metric := distance + deltaDays*8640
					if metric < bestMetric {
						bestJunction, bestMetric = junction.point, metric
					}
				}
				if bestMetric == math.Inf(1) {
					continue
				}
				start, end := bestJunction, fold
				if start.JDE > end.JDE {
					start, end = end, start
				}
				bridge := solver.appendRefinedRiseSetSegment(
					[]SolarEclipsePathPoint{start}, start, end,
					curve.Phase, curve.Direction, 0,
				)
				if len(bridge) >= 2 {
					curve.Segments = append(curve.Segments, bridge)
				}
			}
		}
	}
	for curveIndex := range curves {
		normalizeSolarEclipseRiseSetCurveSegments(&curves[curveIndex])
		mergeSolarEclipseRiseSetArcContinuations(&curves[curveIndex])
	}
}

func solarEclipseRiseSetArcEndpointShared(
	curves []SolarEclipseRiseSetCurve,
	curveIndex, segmentIndex, pointIndex int,
) bool {
	point := curves[curveIndex].Segments[segmentIndex][pointIndex]
	for otherCurveIndex, curve := range curves {
		for otherSegmentIndex, segment := range curve.Segments {
			if len(segment) < 2 {
				continue
			}
			for _, otherPointIndex := range []int{0, len(segment) - 1} {
				if curveIndex == otherCurveIndex && segmentIndex == otherSegmentIndex && pointIndex == otherPointIndex {
					continue
				}
				other := segment[otherPointIndex]
				if math.Abs(point.JDE-other.JDE) <= solarEclipseRiseSetTimeEpsilonDays &&
					solarEclipsePathDistanceKM(point, other) <= 0.01 {
					return true
				}
			}
		}
	}
	return false
}

func solarEclipseRiseSetArcPointInCurve(
	point SolarEclipsePathPoint,
	curve SolarEclipseRiseSetCurve,
) bool {
	for _, segment := range curve.Segments {
		if len(segment) < 2 {
			continue
		}
		for _, pointIndex := range []int{0, len(segment) - 1} {
			candidate := segment[pointIndex]
			if math.Abs(point.JDE-candidate.JDE) <= solarEclipseRiseSetTimeEpsilonDays &&
				solarEclipsePathDistanceKM(point, candidate) <= 0.01 {
				return true
			}
		}
	}
	return false
}

func (solver solarEclipseSolver) snapNearCoincidentSolarEclipseRiseSetEndpoints(
	curves []SolarEclipseRiseSetCurve,
) {
	const (
		maximumTimeDays = 2.0 / 86400.0
		maximumDistance = 10.0
	)
	for curveIndex := range curves {
		curve := &curves[curveIndex]
		for firstSegmentIndex, firstSegment := range curve.Segments {
			if len(firstSegment) < 2 {
				continue
			}
			for secondSegmentIndex := firstSegmentIndex + 1; secondSegmentIndex < len(curve.Segments); secondSegmentIndex++ {
				secondSegment := curve.Segments[secondSegmentIndex]
				if len(secondSegment) < 2 {
					continue
				}
				for _, atStart := range []bool{true, false} {
					firstPointIndex, secondPointIndex := len(firstSegment)-1, len(secondSegment)-1
					if atStart {
						firstPointIndex, secondPointIndex = 0, 0
					}
					first := curve.Segments[firstSegmentIndex][firstPointIndex]
					second := curve.Segments[secondSegmentIndex][secondPointIndex]
					if math.Abs(first.JDE-second.JDE) > maximumTimeDays ||
						solarEclipsePathDistanceKM(first, second) > maximumDistance {
						continue
					}
					sharedFirst := solarEclipseRiseSetArcEndpointShared(
						curves, curveIndex, firstSegmentIndex, firstPointIndex,
					)
					sharedSecond := solarEclipseRiseSetArcEndpointShared(
						curves, curveIndex, secondSegmentIndex, secondPointIndex,
					)
					common := first
					switch {
					case sharedSecond && !sharedFirst:
						common = second
					case !sharedFirst && !sharedSecond:
						if fold, ok := solver.refineRiseSetFold(
							first, second, curve.Phase == RiseSetPhaseGreatest,
						); ok {
							common = fold
						} else {
							common = solarEclipseRiseSetMidpoint(first, second)
						}
					}
					curve.Segments[firstSegmentIndex][firstPointIndex] = common
					curve.Segments[secondSegmentIndex][secondPointIndex] = common
				}
			}
		}
	}
}

func (solver solarEclipseSolver) closeSolarEclipseRiseSetArcDirectionJunctions(
	curves []SolarEclipseRiseSetCurve,
	phaseJunctions []solarEclipseRiseSetPhaseJunction,
) {
	const (
		maximumTimeDays         = 2.0 / 1440.0
		maximumEndpointDistance = 1500.0
		maximumPairDistance     = 3000.0
	)
	curveIndices := make(map[solarEclipseRiseSetCurveKey]int, len(curves))
	for index, curve := range curves {
		curveIndices[solarEclipseRiseSetCurveKey{phase: curve.Phase, direction: curve.Direction}] = index
	}
	for curveIndex := range curves {
		curve := &curves[curveIndex]
		greatest := curve.Phase == RiseSetPhaseGreatest
		for segmentIndex, segment := range curve.Segments {
			if len(segment) < 2 {
				continue
			}
			for _, pointIndex := range []int{0, len(segment) - 1} {
				if solarEclipseRiseSetArcEndpointShared(curves, curveIndex, segmentIndex, pointIndex) {
					continue
				}
				endpoint := curve.Segments[segmentIndex][pointIndex]
				junction, ok := solver.refineRiseSetDirectionJunction(
					endpoint.JDE, endpoint.Longitude, endpoint.Latitude, greatest,
				)
				if !ok || math.Abs(junction.JDE-endpoint.JDE) > maximumTimeDays ||
					solarEclipsePathDistanceKM(junction, endpoint) > maximumEndpointDistance {
					continue
				}
				oppositeDirection := RiseSetDirectionSet
				if curve.Direction == RiseSetDirectionSet {
					oppositeDirection = RiseSetDirectionRise
				}
				oppositeIndex, haveOpposite := curveIndices[solarEclipseRiseSetCurveKey{
					phase: curve.Phase, direction: oppositeDirection,
				}]
				if !haveOpposite {
					continue
				}
				bestSegment, bestPoint := -1, -1
				bestMetric := math.Inf(1)
				for otherSegmentIndex, otherSegment := range curves[oppositeIndex].Segments {
					if len(otherSegment) < 2 {
						continue
					}
					for _, otherPointIndex := range []int{0, len(otherSegment) - 1} {
						other := otherSegment[otherPointIndex]
						deltaDays := math.Abs(junction.JDE - other.JDE)
						distance := solarEclipsePathDistanceKM(junction, other)
						if deltaDays > maximumTimeDays || distance > maximumPairDistance {
							continue
						}
						metric := distance + deltaDays*8640
						if metric < bestMetric {
							bestSegment, bestPoint, bestMetric = otherSegmentIndex, otherPointIndex, metric
						}
					}
				}
				if bestSegment < 0 {
					bestJunction := SolarEclipsePathPoint{}
					bestJunctionMetric := math.Inf(1)
					for _, phaseJunction := range phaseJunctions {
						if phaseJunction.direction != oppositeDirection ||
							solarEclipseRiseSetArcPointInCurve(phaseJunction.point, curves[oppositeIndex]) {
							continue
						}
						deltaDays := math.Abs(junction.JDE - phaseJunction.point.JDE)
						distance := solarEclipsePathDistanceKM(junction, phaseJunction.point)
						if deltaDays > maximumTimeDays || distance > maximumPairDistance {
							continue
						}
						metric := distance + deltaDays*8640
						if metric < bestJunctionMetric {
							bestJunction, bestJunctionMetric = phaseJunction.point, metric
						}
					}
					if bestJunctionMetric == math.Inf(1) {
						continue
					}
					start, end := junction, bestJunction
					if start.JDE > end.JDE {
						start, end = end, start
					}
					bridge := solver.appendRefinedRiseSetSegment(
						[]SolarEclipsePathPoint{start}, start, end,
						curve.Phase, oppositeDirection, 0,
					)
					if len(bridge) < 2 {
						continue
					}
					curve.Segments[segmentIndex][pointIndex] = junction
					curves[oppositeIndex].Segments = append(curves[oppositeIndex].Segments, bridge)
					continue
				}
				curve.Segments[segmentIndex][pointIndex] = junction
				curves[oppositeIndex].Segments[bestSegment][bestPoint] = junction
			}
		}
	}
	for curveIndex := range curves {
		normalizeSolarEclipseRiseSetCurveSegments(&curves[curveIndex])
		mergeSolarEclipseRiseSetArcContinuations(&curves[curveIndex])
	}
}

func solarEclipseRiseSetArcContainsSeed(states []solarEclipseRiseSetArcState, seed SolarEclipsePathPoint) bool {
	for _, state := range states {
		if math.Abs(state.point.JDE-seed.JDE) <= solarEclipseRiseSetArcSeedTimeDays &&
			solarEclipsePathDistanceKM(state.point, seed) <= solarEclipseRiseSetArcSeedDistance {
			return true
		}
	}
	return false
}

func (solver solarEclipseSolver) traceRiseSetArcComponent(
	seed SolarEclipsePathPoint,
	greatest bool,
	referenceJDE, startJDE, endJDE float64,
) []solarEclipseRiseSetArcState {
	forward, closed := solver.traceRiseSetArc(seed, greatest, 1, referenceJDE, startJDE, endJDE)
	if closed {
		return forward
	}
	backward, _ := solver.traceRiseSetArc(seed, greatest, -1, referenceJDE, startJDE, endJDE)
	states := make([]solarEclipseRiseSetArcState, 0, len(backward)+len(forward)-1)
	for index := len(backward) - 1; index >= 0; index-- {
		state := backward[index]
		for tangentIndex := range state.tangent {
			state.tangent[tangentIndex] = -state.tangent[tangentIndex]
		}
		states = append(states, state)
	}
	return append(states, forward[1:]...)
}

func (solver solarEclipseSolver) traceRiseSetArc(
	seed SolarEclipsePathPoint,
	greatest bool,
	direction int,
	referenceJDE, startJDE, endJDE float64,
) ([]solarEclipseRiseSetArcState, bool) {
	state, ok := solver.riseSetArcStateAt(seed, greatest, referenceJDE)
	if !ok {
		return nil, false
	}
	for index := range state.tangent {
		state.tangent[index] *= float64(direction)
	}
	initial := state
	states := []solarEclipseRiseSetArcState{state}
	step := solarEclipseRiseSetArcInitialStep
	previousSeedPlane := 0.0
	closureArmed := false
	for count := 0; count < solarEclipseRiseSetArcMaximumSteps; count++ {
		predictor := state.coordinates
		for index := range predictor {
			predictor[index] += step * state.tangent[index]
		}
		next, iterations, nextOK := solver.correctRiseSetArc(predictor, state.tangent, greatest, referenceJDE)
		if !nextOK {
			step /= 2
			if step < solarEclipseRiseSetArcMinimumStep {
				break
			}
			continue
		}
		if dotSolarEclipse3(next.tangent, state.tangent) < 0 {
			for index := range next.tangent {
				next.tangent[index] = -next.tangent[index]
			}
		}
		distance := solarEclipsePathDistanceKM(state.point, next.point)
		multipleTransitions := solver.riseSetArcTransitionCount(state, next, greatest) > 1
		if distance > solarEclipseRiseSetArcTargetSpacing ||
			multipleTransitions && step > 2*solarEclipseRiseSetArcMinimumStep {
			step /= 2
			if step < solarEclipseRiseSetArcMinimumStep && !multipleTransitions {
				break
			}
			if step < solarEclipseRiseSetArcMinimumStep {
				step = solarEclipseRiseSetArcMinimumStep
			}
			continue
		}
		if next.point.JDE < startJDE-0.01 || next.point.JDE > endJDE+0.01 {
			break
		}
		states = append(states, next)
		state = next
		seedPlane := solarEclipseRiseSetArcSeedPlane(next, initial)
		if seedPlane < -solarEclipseRiseSetArcMinimumStep {
			closureArmed = true
		}
		if count > 30 && closureArmed && previousSeedPlane < 0 && seedPlane >= 0 &&
			math.Abs(next.point.JDE-initial.point.JDE) <= solarEclipseRiseSetArcCloseTimeDays &&
			solarEclipsePathDistanceKM(next.point, initial.point) <= solarEclipseRiseSetArcCloseDistance &&
			dotSolarEclipse3(next.tangent, initial.tangent) > 0.5 {
			states[len(states)-1] = initial
			return states, true
		}
		previousSeedPlane = seedPlane
		if distance < solarEclipseRiseSetArcTargetSpacing/2 && iterations <= 4 {
			step = math.Min(solarEclipseRiseSetArcInitialStep, step*1.5)
		}
	}
	return states, false
}

func solarEclipseRiseSetArcSeedPlane(
	state, seed solarEclipseRiseSetArcState,
) float64 {
	delta := [3]float64{
		math.Remainder(state.coordinates[0]-seed.coordinates[0], 360),
		state.coordinates[1] - seed.coordinates[1],
		state.coordinates[2] - seed.coordinates[2],
	}
	return dotSolarEclipse3(delta, seed.tangent)
}

func (solver solarEclipseSolver) riseSetArcTransitionCount(
	first, second solarEclipseRiseSetArcState,
	greatest bool,
) int {
	count := 0
	if first.tangent[2]*second.tangent[2] < 0 {
		count++
	}
	firstEvaluation := solver.magnitudeEvaluationAt(first.point.JDE)
	secondEvaluation := solver.magnitudeEvaluationAt(second.point.JDE)
	if firstEvaluation.sunAltitudeDerivative(first.point.Longitude, first.point.Latitude)*
		secondEvaluation.sunAltitudeDerivative(second.point.Longitude, second.point.Latitude) < 0 {
		count++
	}
	if greatest {
		firstGap := solarEclipsePartialContactGap(firstEvaluation.center.stateAt(first.point.Longitude*rad, first.point.Latitude*rad, 0))
		secondGap := solarEclipsePartialContactGap(secondEvaluation.center.stateAt(second.point.Longitude*rad, second.point.Latitude*rad, 0))
		if firstGap*secondGap < 0 {
			count++
		}
	} else if firstEvaluation.partialContactDerivative(first.point.Longitude, first.point.Latitude)*
		secondEvaluation.partialContactDerivative(second.point.Longitude, second.point.Latitude) < 0 {
		count++
	}
	return count
}

func (solver solarEclipseSolver) riseSetArcStateAt(
	point SolarEclipsePathPoint,
	greatest bool,
	referenceJDE float64,
) (solarEclipseRiseSetArcState, bool) {
	coordinates := [3]float64{
		point.Longitude,
		point.Latitude,
		(point.JDE - referenceJDE) * solarEclipseRiseSetArcTimeScale,
	}
	_, jacobian, ok := solver.riseSetArcJacobian(coordinates, greatest, referenceJDE)
	if !ok {
		return solarEclipseRiseSetArcState{}, false
	}
	tangent, ok := solarEclipseMagnitudeArcTangent(jacobian)
	return solarEclipseRiseSetArcState{coordinates: coordinates, tangent: tangent, point: point}, ok
}

func (solver solarEclipseSolver) correctRiseSetArc(
	predictor, tangent [3]float64,
	greatest bool,
	referenceJDE float64,
) (solarEclipseRiseSetArcState, int, bool) {
	coordinates := predictor
	for iteration := 0; iteration < 16; iteration++ {
		residual, jacobian, ok := solver.riseSetArcJacobian(coordinates, greatest, referenceJDE)
		if !ok {
			return solarEclipseRiseSetArcState{}, iteration, false
		}
		planeResidual := dotSolarEclipse3(subtractSolarEclipse3(coordinates, predictor), tangent)
		if math.Abs(residual[0]) <= 1e-10 && math.Abs(residual[1]) <= 1e-10 && math.Abs(planeResidual) <= 1e-9 {
			return solver.validRiseSetArcState(coordinates, jacobian, referenceJDE, iteration+1)
		}
		matrix := [3][3]float64{jacobian[0], jacobian[1], tangent}
		delta, ok := solveSolarEclipse3x3(matrix, [3]float64{-residual[0], -residual[1], -planeResidual})
		if !ok {
			return solarEclipseRiseSetArcState{}, iteration, false
		}
		norm := math.Sqrt(dotSolarEclipse3(delta, delta))
		if norm > 2 {
			for index := range delta {
				delta[index] *= 2 / norm
			}
		}
		for index := range coordinates {
			coordinates[index] += delta[index]
		}
		coordinates[0] = normalizeLongitude(coordinates[0])
		if coordinates[1] <= -89.999999 || coordinates[1] >= 89.999999 {
			return solarEclipseRiseSetArcState{}, iteration, false
		}
	}
	residual, jacobian, ok := solver.riseSetArcJacobian(coordinates, greatest, referenceJDE)
	planeResidual := dotSolarEclipse3(subtractSolarEclipse3(coordinates, predictor), tangent)
	if !ok || math.Abs(residual[0]) > 1e-7 || math.Abs(residual[1]) > 1e-7 || math.Abs(planeResidual) > 1e-7 {
		return solarEclipseRiseSetArcState{}, 16, false
	}
	return solver.validRiseSetArcState(coordinates, jacobian, referenceJDE, 16)
}

func (solver solarEclipseSolver) validRiseSetArcState(
	coordinates [3]float64,
	jacobian [2][3]float64,
	referenceJDE float64,
	iterations int,
) (solarEclipseRiseSetArcState, int, bool) {
	jde := referenceJDE + coordinates[2]/solarEclipseRiseSetArcTimeScale
	longitude, latitude := normalizeLongitude(coordinates[0]), coordinates[1]
	evaluation := solver.magnitudeEvaluationAt(jde)
	state := evaluation.center.stateAt(longitude*rad, latitude*rad, 0)
	tangent, ok := solarEclipseMagnitudeArcTangent(jacobian)
	if !ok {
		return solarEclipseRiseSetArcState{}, iterations, false
	}
	return solarEclipseRiseSetArcState{
		coordinates: coordinates,
		tangent:     tangent,
		point: SolarEclipsePathPoint{
			JDE: jde, Longitude: longitude, Latitude: latitude, SunAltitude: state.sunAltitudeRad / rad,
		},
	}, iterations, true
}

func (solver solarEclipseSolver) riseSetArcJacobian(
	coordinates [3]float64,
	greatest bool,
	referenceJDE float64,
) ([2]float64, [2][3]float64, bool) {
	jde := referenceJDE + coordinates[2]/solarEclipseRiseSetArcTimeScale
	longitude, latitude := normalizeLongitude(coordinates[0]), coordinates[1]
	evaluation := solver.magnitudeEvaluationAt(jde)
	residual, ok := solarEclipseRiseSetArcResidualAt(evaluation, longitude, latitude, greatest)
	if !ok {
		return [2]float64{}, [2][3]float64{}, false
	}
	steps := [3]float64{1e-4, 1e-4, 5.0 * solarEclipseRiseSetArcTimeScale / 86400.0}
	jacobian := [2][3]float64{}
	for column, shifted := range [][2]float64{{longitude + steps[0], latitude}, {longitude, latitude + steps[1]}} {
		shiftedResidual, shiftedOK := solarEclipseRiseSetArcResidualAt(evaluation, shifted[0], shifted[1], greatest)
		if !shiftedOK {
			return [2]float64{}, [2][3]float64{}, false
		}
		for row := 0; row < 2; row++ {
			jacobian[row][column] = (shiftedResidual[row] - residual[row]) / steps[column]
		}
	}
	timeEvaluation := solver.magnitudeEvaluationAt(jde + steps[2]/solarEclipseRiseSetArcTimeScale)
	timeResidual, timeOK := solarEclipseRiseSetArcResidualAt(timeEvaluation, longitude, latitude, greatest)
	if !timeOK {
		return [2]float64{}, [2][3]float64{}, false
	}
	for row := 0; row < 2; row++ {
		jacobian[row][2] = (timeResidual[row] - residual[row]) / steps[2]
	}
	return residual, jacobian, true
}

func solarEclipseRiseSetArcResidualAt(
	evaluation solarEclipseRiseSetEvaluation,
	longitude, latitude float64,
	greatest bool,
) ([2]float64, bool) {
	// 非 greatest 相位残差就是同一个 stateAt 的中心距盈余，复用该状态而不是再算一次。
	state := evaluation.center.stateAt(longitude*rad, latitude*rad, 0)
	phase := 0.0
	if greatest {
		phase = evaluation.separationDerivative(longitude, latitude)
	} else {
		phase = solarEclipsePartialContactGap(state)
	}
	return [2]float64{phase, state.sunAltitudeRad},
		finite(phase) && finite(state.sunAltitudeRad)
}

func (solver solarEclipseSolver) splitRiseSetArcComponent(
	states []solarEclipseRiseSetArcState,
	greatest bool,
) []solarEclipseRiseSetArcSegment {
	if len(states) < 2 {
		return nil
	}
	key, valid := solver.riseSetArcKey(states[0], greatest)
	var segments []solarEclipseRiseSetArcSegment
	current := solarEclipseRiseSetArcSegment{key: key}
	if valid {
		current.points = append(current.points, states[0].point)
	}
	for index := 0; index < len(states)-1; index++ {
		first, second := states[index], states[index+1]
		transition, point := solver.riseSetArcTransition(first, second, greatest)
		if transition == solarEclipseRiseSetArcNoTransition {
			if valid {
				current.points = append(current.points, second.point)
			}
			continue
		}
		if valid {
			current.points = append(current.points, point)
			if len(current.points) >= 2 {
				segments = append(segments, current)
			}
		}
		switch transition {
		case solarEclipseRiseSetArcPhaseTransition:
			if greatest {
				valid = !valid
			} else if key.phase == RiseSetPhaseStart {
				key.phase = RiseSetPhaseEnd
			} else {
				key.phase = RiseSetPhaseStart
			}
		case solarEclipseRiseSetArcDirectionTransition:
			if key.direction == RiseSetDirectionRise {
				key.direction = RiseSetDirectionSet
			} else {
				key.direction = RiseSetDirectionRise
			}
		case solarEclipseRiseSetArcFoldTransition:
		}
		if classifiedKey, classifiedValid := solver.riseSetArcKey(second, greatest); classifiedValid {
			key, valid = classifiedKey, true
		} else if greatest {
			valid = false
		}
		current = solarEclipseRiseSetArcSegment{key: key}
		if valid {
			current.points = append(current.points, point, second.point)
		}
	}
	if valid && len(current.points) >= 2 {
		segments = append(segments, current)
	}
	if len(segments) >= 2 && states[0].point == states[len(states)-1].point &&
		segments[0].key == segments[len(segments)-1].key {
		last := segments[len(segments)-1]
		first := segments[0]
		last.points = append(last.points[:len(last.points)-1], first.points...)
		segments[0] = last
		segments = segments[:len(segments)-1]
	}
	return segments
}

func (solver solarEclipseSolver) riseSetArcKey(
	state solarEclipseRiseSetArcState,
	greatest bool,
) (solarEclipseRiseSetCurveKey, bool) {
	evaluation := solver.magnitudeEvaluationAt(state.point.JDE)
	if greatest {
		longitude, latitude := state.point.Longitude, state.point.Latitude
		local := evaluation.center.stateAt(longitude*rad, latitude*rad, 0)
		altitudeDerivative := evaluation.sunAltitudeDerivative(longitude, latitude)
		if solarEclipsePartialContactGap(local) > 0 ||
			evaluation.separationSecondDerivative(longitude, latitude) <= 0 ||
			!finite(altitudeDerivative) || math.Abs(altitudeDerivative) < 1e-10 {
			return solarEclipseRiseSetCurveKey{}, false
		}
		direction := RiseSetDirectionSet
		if altitudeDerivative > 0 {
			direction = RiseSetDirectionRise
		}
		return solarEclipseRiseSetCurveKey{phase: RiseSetPhaseGreatest, direction: direction}, true
	}
	point, key, valid := evaluation.classify(state.point.Longitude, state.point.Latitude, greatest)
	_ = point
	return key, valid
}

func (solver solarEclipseSolver) riseSetArcTransition(
	first, second solarEclipseRiseSetArcState,
	greatest bool,
) (solarEclipseRiseSetArcTransition, SolarEclipsePathPoint) {
	firstEvaluation := solver.magnitudeEvaluationAt(first.point.JDE)
	secondEvaluation := solver.magnitudeEvaluationAt(second.point.JDE)
	if greatest {
		firstGap := solarEclipsePartialContactGap(firstEvaluation.center.stateAt(first.point.Longitude*rad, first.point.Latitude*rad, 0))
		secondGap := solarEclipsePartialContactGap(secondEvaluation.center.stateAt(second.point.Longitude*rad, second.point.Latitude*rad, 0))
		if firstGap*secondGap <= 0 {
			if point, ok := solver.refineRiseSetPhaseJunctionOnHorizon(solarEclipseRiseSetMidpoint(first.point, second.point)); ok {
				return solarEclipseRiseSetArcPhaseTransition, point
			}
		}
	} else {
		firstDerivative := firstEvaluation.partialContactDerivative(first.point.Longitude, first.point.Latitude)
		secondDerivative := secondEvaluation.partialContactDerivative(second.point.Longitude, second.point.Latitude)
		if firstDerivative*secondDerivative <= 0 {
			if point, ok := solver.refineRiseSetPhaseJunctionOnHorizon(solarEclipseRiseSetMidpoint(first.point, second.point)); ok {
				return solarEclipseRiseSetArcPhaseTransition, point
			}
		}
	}
	firstAltitudeDerivative := firstEvaluation.sunAltitudeDerivative(first.point.Longitude, first.point.Latitude)
	secondAltitudeDerivative := secondEvaluation.sunAltitudeDerivative(second.point.Longitude, second.point.Latitude)
	if firstAltitudeDerivative*secondAltitudeDerivative <= 0 {
		if point, ok := solver.refineRiseSetDirectionJunction(
			(first.point.JDE+second.point.JDE)/2,
			normalizeLongitude(first.point.Longitude+math.Remainder(second.point.Longitude-first.point.Longitude, 360)/2),
			(first.point.Latitude+second.point.Latitude)/2,
			greatest,
		); ok {
			return solarEclipseRiseSetArcDirectionTransition, point
		}
	}
	if first.tangent[2]*second.tangent[2] <= 0 {
		if point, ok := solver.refineRiseSetFold(first.point, second.point, greatest); ok {
			return solarEclipseRiseSetArcFoldTransition, point
		}
	}
	return solarEclipseRiseSetArcNoTransition, SolarEclipsePathPoint{}
}

func sortSolarEclipseRiseSetSegments(curves []SolarEclipseRiseSetCurve) {
	for curveIndex := range curves {
		sort.Slice(curves[curveIndex].Segments, func(first, second int) bool {
			return curves[curveIndex].Segments[first][0].JDE < curves[curveIndex].Segments[second][0].JDE
		})
	}
}

func deduplicateSolarEclipseRiseSetArcSegments(curves []SolarEclipseRiseSetCurve) {
	for curveIndex := range curves {
		segments := curves[curveIndex].Segments
		unique := make([][]SolarEclipsePathPoint, 0, len(segments))
		for _, segment := range segments {
			duplicate := false
			for _, existing := range unique {
				if solarEclipseRiseSetArcSegmentsEquivalent(segment, existing) {
					duplicate = true
					break
				}
			}
			if !duplicate {
				unique = append(unique, segment)
			}
		}
		curves[curveIndex].Segments = unique
	}
}

func solarEclipseRiseSetArcSegmentsEquivalent(
	first, second []SolarEclipsePathPoint,
) bool {
	if len(first) < 2 || len(second) < 2 {
		return false
	}
	for _, endpoints := range [][2]SolarEclipsePathPoint{
		{first[0], second[0]},
		{first[len(first)-1], second[len(second)-1]},
	} {
		if math.Abs(endpoints[0].JDE-endpoints[1].JDE) > 1.0/86400.0 ||
			solarEclipsePathDistanceKM(endpoints[0], endpoints[1]) > 1 {
			return false
		}
	}
	for _, fraction := range []float64{0.25, 0.5, 0.75} {
		jd := first[0].JDE + fraction*(first[len(first)-1].JDE-first[0].JDE)
		firstPoint, firstOK := solarEclipseRiseSetArcPointAtTime(first, jd)
		secondPoint, secondOK := solarEclipseRiseSetArcPointAtTime(second, jd)
		if !firstOK || !secondOK || solarEclipsePathDistanceKM(firstPoint, secondPoint) > 250 {
			return false
		}
	}
	return true
}

func solarEclipseRiseSetArcPointAtTime(
	segment []SolarEclipsePathPoint,
	jd float64,
) (SolarEclipsePathPoint, bool) {
	if len(segment) < 2 || jd < segment[0].JDE || jd > segment[len(segment)-1].JDE {
		return SolarEclipsePathPoint{}, false
	}
	index := sort.Search(len(segment), func(index int) bool { return segment[index].JDE >= jd })
	if index == 0 {
		return segment[0], true
	}
	if index >= len(segment) {
		return segment[len(segment)-1], true
	}
	before, after := segment[index-1], segment[index]
	if after.JDE <= before.JDE {
		return SolarEclipsePathPoint{}, false
	}
	fraction := (jd - before.JDE) / (after.JDE - before.JDE)
	return SolarEclipsePathPoint{
		JDE: jd,
		Longitude: normalizeLongitude(
			before.Longitude + fraction*math.Remainder(after.Longitude-before.Longitude, 360),
		),
		Latitude: before.Latitude + fraction*(after.Latitude-before.Latitude),
	}, true
}
