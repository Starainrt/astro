package basic

import (
	"math"
	"sort"

	"github.com/starainrt/astro/internal/geodata"
)

func (solver solarEclipseSolver) magnitudeContourSegments(
	startJDE, endJDE, centralStartJDE, centralEndJDE,
	greatestJDE, magnitude, fallbackStepDays float64, hybrid bool,
) [][]SolarEclipsePathPoint {
	return solver.magnitudeContourSegmentsWithSpacing(
		startJDE, endJDE, centralStartJDE, centralEndJDE,
		greatestJDE, magnitude, fallbackStepDays, hybrid,
		solarEclipseMagnitudeContourTargetSpacingKM,
	)
}

func (solver solarEclipseSolver) magnitudeContourSegmentsWithSpacing(
	startJDE, endJDE, centralStartJDE, centralEndJDE,
	greatestJDE, magnitude, fallbackStepDays float64, hybrid bool,
	targetSpacingKM float64,
) [][]SolarEclipsePathPoint {
	if targetSpacingKM <= 0 || math.IsNaN(targetSpacingKM) || math.IsInf(targetSpacingKM, 0) {
		targetSpacingKM = solarEclipseMagnitudeContourTargetSpacingKM
	}
	// 追踪用事件局部插值星历求残差与雅可比，输出点仍由 validMagnitudeArcState 精确确认。
	solver = solver.withLocalEphemeris()
	var transitions []SolarEclipsePathPoint
	if hybrid && math.Abs(magnitude-1) <= 1e-12 {
		transitions = solver.centralMagnitudeOneTransitionsInInterval(
			centralStartJDE, centralEndJDE, greatestJDE, fallbackStepDays,
		)
		if len(transitions) == 2 {
			if segments := solver.hybridMagnitudeOneLimitSegments(transitions, fallbackStepDays); len(segments) == 2 {
				return segments
			}
		}
	}
	seeds := solver.magnitudeContourPointsAt(greatestJDE, magnitude)
	if len(seeds) == 0 {
		times, _ := solarEclipsePathSampleTimes(startJDE, endJDE, greatestJDE, fallbackStepDays)
		seedIndex := sort.SearchFloat64s(times, greatestJDE)
		for offset := 1; offset < len(times); offset++ {
			for _, index := range []int{seedIndex - offset, seedIndex + offset} {
				if index < 0 || index >= len(times) {
					continue
				}
				seeds = solver.magnitudeContourPointsAt(times[index], magnitude)
				if len(seeds) > 0 {
					break
				}
			}
			if len(seeds) > 0 {
				break
			}
		}
	}
	segments := make([][]SolarEclipsePathPoint, 0, len(seeds))
	constrainBranch := hybrid && math.Abs(magnitude-1) <= 1e-12
	for _, seed := range seeds {
		backward := solver.traceMagnitudeContourArc(
			seed, magnitude, -1, greatestJDE, startJDE, endJDE, targetSpacingKM, constrainBranch,
		)
		forward := solver.traceMagnitudeContourArc(
			seed, magnitude, 1, greatestJDE, startJDE, endJDE, targetSpacingKM, constrainBranch,
		)
		segment := make([]SolarEclipsePathPoint, 0, len(backward)+len(forward)-1)
		for index := len(backward) - 1; index >= 0; index-- {
			segment = append(segment, backward[index])
		}
		segment = append(segment, forward[1:]...)
		if len(segment) >= 2 {
			segments = append(segments, segment)
		}
	}
	if len(transitions) > 0 {
		segments = solver.completeMagnitudeOneContourTransitions(segments, transitions)
	}
	// A grazing zero-magnitude solve can converge to a degenerate branch that
	// only touches the horizon at its endpoints and otherwise lies below it.
	// That branch is not part of the visible envelope; retaining it creates a
	// short backward spur when the phase curves are polygonized. Keep branches
	// with a measurable above-horizon portion and let rise/set curves represent
	// the exact tangent contact itself.
	visible := segments[:0]
	for _, segment := range segments {
		maximumAltitude := math.Inf(-1)
		for _, point := range segment {
			maximumAltitude = math.Max(maximumAltitude, point.SunAltitude)
		}
		if maximumAltitude > 1e-6 {
			visible = append(visible, segment)
		}
	}
	segments = visible
	return segments
}

func (solver solarEclipseSolver) hybridMagnitudeOneLimitSegments(
	transitions []SolarEclipsePathPoint,
	fallbackStepDays float64,
) [][]SolarEclipsePathPoint {
	if len(transitions) != 2 {
		return nil
	}
	start, end := transitions[0], transitions[1]
	if start.JDE > end.JDE {
		start, end = end, start
	}
	stepDays := math.Min(fallbackStepDays, (end.JDE-start.JDE)/64)
	centerLine, _ := solver.centralPathPoints(
		start.JDE,
		end.JDE,
		(start.JDE+end.JDE)/2,
		SolarEclipsePathOptions{
			StepDays:        stepDays,
			TargetSpacingKM: solarEclipseMagnitudeContourTargetSpacingKM,
		},
	)
	northern, southern := solver.centralPathLimits(centerLine)
	if len(northern) < 2 || len(northern) != len(southern) {
		return nil
	}
	for _, segment := range [][]SolarEclipsePathPoint{northern, southern} {
		segment[0] = start
		segment[0].WidthKM = 0
		segment[len(segment)-1] = end
		segment[len(segment)-1].WidthKM = 0
	}
	return [][]SolarEclipsePathPoint{northern, southern}
}

func (solver solarEclipseSolver) centralMagnitudeOneTransitionsInInterval(
	startJDE, endJDE, greatestJDE, stepDays float64,
) []SolarEclipsePathPoint {
	if startJDE == 0 || endJDE == 0 || endJDE <= startJDE {
		return nil
	}
	transitionStepDays := math.Min(stepDays, (endJDE-startJDE)/64)
	times, _ := solarEclipsePathSampleTimes(startJDE, endJDE, greatestJDE, transitionStepDays)
	edgeOffset := math.Min(1.0/86400.0, (endJDE-startJDE)/1000)
	if edgeOffset > 0 {
		times = append(times, startJDE+edgeOffset, endJDE-edgeOffset)
		sort.Float64s(times)
		times = uniqueSolarEclipsePathTimes(times)
	}
	points := make([]SolarEclipsePathPoint, 0, len(times))
	for _, jd := range times {
		if point, ok := solver.centralPathPointAt(jd); ok {
			points = append(points, point)
		}
	}
	return solver.centralMagnitudeOneTransitions(points)
}

func (solver solarEclipseSolver) centralMagnitudeOneTransitions(
	points []SolarEclipsePathPoint,
) []SolarEclipsePathPoint {
	if len(points) < 2 {
		return nil
	}
	transitions := make([]SolarEclipsePathPoint, 0, 2)
	previousPoint := points[0]
	previousGap, previousOK := solver.centralMagnitudeOneGap(previousPoint)
	for _, point := range points[1:] {
		gap, ok := solver.centralMagnitudeOneGap(point)
		if previousOK && ok && previousGap*gap <= 0 {
			transition, transitionOK := solver.refineCentralMagnitudeOneTransition(
				previousPoint, point, previousGap, gap,
			)
			if transitionOK && !solarEclipseRiseSetPointExists(transitions, transition) {
				transitions = append(transitions, transition)
			}
		}
		previousPoint, previousGap, previousOK = point, gap, ok
	}
	return transitions
}

func (solver solarEclipseSolver) centralMagnitudeOneGap(point SolarEclipsePathPoint) (float64, bool) {
	context := solver.localStateContextAt(point.JDE)
	state := context.stateAt(point.Longitude*rad, point.Latitude*rad, 0)
	gap := state.moonInnerRadiusRad/state.sunRadiusRad - 1
	return gap, finite(gap)
}

func (solver solarEclipseSolver) refineCentralMagnitudeOneTransition(
	left, right SolarEclipsePathPoint,
	leftGap, rightGap float64,
) (SolarEclipsePathPoint, bool) {
	if left.JDE > right.JDE {
		left, right = right, left
		leftGap, rightGap = rightGap, leftGap
	}
	if leftGap*rightGap > 0 {
		return SolarEclipsePathPoint{}, false
	}
	if math.Abs(leftGap) <= 1e-12 {
		return left, true
	}
	if math.Abs(rightGap) <= 1e-12 {
		return right, true
	}
	for iteration := 0; iteration < 64 && right.JDE-left.JDE > solarEclipsePathDuplicateTimeDays; iteration++ {
		middle, ok := solver.centralPathPointAt((left.JDE + right.JDE) / 2)
		if !ok {
			return SolarEclipsePathPoint{}, false
		}
		middleGap, ok := solver.centralMagnitudeOneGap(middle)
		if !ok {
			return SolarEclipsePathPoint{}, false
		}
		if math.Abs(middleGap) <= 1e-12 {
			return middle, true
		}
		if leftGap*middleGap <= 0 {
			right, rightGap = middle, middleGap
		} else {
			left, leftGap = middle, middleGap
		}
	}
	transition, ok := solver.centralPathPointAt((left.JDE + right.JDE) / 2)
	if !ok {
		return SolarEclipsePathPoint{}, false
	}
	gap, ok := solver.centralMagnitudeOneGap(transition)
	return transition, ok && math.Abs(gap) <= 1e-8
}

func (solver solarEclipseSolver) completeMagnitudeOneContourTransitions(
	segments [][]SolarEclipsePathPoint,
	transitions []SolarEclipsePathPoint,
) [][]SolarEclipsePathPoint {
	if len(transitions) == 0 {
		return segments
	}
	const maximumTimeGapDays = 5.0 / 1440.0
	for segmentIndex, segment := range segments {
		if len(segment) < 2 {
			continue
		}
		for _, atStart := range []bool{true, false} {
			endpointIndex := len(segment) - 1
			if atStart {
				endpointIndex = 0
			}
			endpoint := segment[endpointIndex]
			bestIndex := -1
			bestDistance := math.Inf(1)
			for transitionIndex, transition := range transitions {
				distance := solarEclipsePathDistanceKM(endpoint, transition)
				if distance < bestDistance {
					bestIndex, bestDistance = transitionIndex, distance
				}
			}
			if bestIndex < 0 || bestDistance > 3000 ||
				math.Abs(endpoint.JDE-transitions[bestIndex].JDE) > maximumTimeGapDays {
				continue
			}
			transition := transitions[bestIndex]
			transition.WidthKM = 0
			if bestDistance < 0.01 {
				segment[endpointIndex] = transition
				continue
			}
			if bestDistance > solarEclipseMagnitudeContourTargetSpacingKM {
				if bridge := solver.magnitudeOneTransitionBridge(transition, endpoint); len(bridge) >= 2 {
					if atStart {
						segment = append(bridge[:len(bridge)-1], segment...)
					} else {
						for index := len(bridge) - 2; index >= 0; index-- {
							segment = append(segment, bridge[index])
						}
					}
					continue
				}
			}
			if atStart {
				segment = append([]SolarEclipsePathPoint{transition}, segment...)
			} else {
				segment = append(segment, transition)
			}
		}
		segments[segmentIndex] = segment
	}
	return segments
}

func (solver solarEclipseSolver) magnitudeOneTransitionBridge(
	transition, endpoint SolarEclipsePathPoint,
) []SolarEclipsePathPoint {
	start, end := transition, endpoint
	reverse := false
	if start.JDE > end.JDE {
		start, end = end, start
		reverse = true
	}
	spanDays := end.JDE - start.JDE
	if spanDays <= 0 {
		return nil
	}
	centerLine, _ := solver.centralPathPoints(
		start.JDE,
		end.JDE,
		(start.JDE+end.JDE)/2,
		SolarEclipsePathOptions{
			StepDays:        math.Min(1.0/1440.0, spanDays/16),
			TargetSpacingKM: solarEclipseMagnitudeContourTargetSpacingKM,
		},
	)
	first, second := solver.centralPathLimits(centerLine)
	if len(first) < 2 || len(first) != len(second) {
		return nil
	}
	bridge := first
	if reverse {
		if solarEclipsePathDistanceKM(second[0], endpoint) < solarEclipsePathDistanceKM(first[0], endpoint) {
			bridge = second
		}
	} else if solarEclipsePathDistanceKM(second[len(second)-1], endpoint) <
		solarEclipsePathDistanceKM(first[len(first)-1], endpoint) {
		bridge = second
	}
	bridge[0], bridge[len(bridge)-1] = start, end
	for index := range bridge {
		bridge[index].WidthKM = 0
	}
	if reverse {
		for left, right := 0, len(bridge)-1; left < right; left, right = left+1, right-1 {
			bridge[left], bridge[right] = bridge[right], bridge[left]
		}
	}
	return bridge
}

type solarEclipseMagnitudeArcState struct {
	coordinates [3]float64
	tangent     [3]float64
	point       SolarEclipsePathPoint
}

func (solver solarEclipseSolver) traceMagnitudeContourArc(
	seed SolarEclipsePathPoint,
	magnitude float64,
	direction int,
	referenceJDE, startJDE, endJDE float64,
	targetSpacingKM float64,
	constrainBranch bool,
) []SolarEclipsePathPoint {
	state, ok := solver.magnitudeArcStateAt(seed, magnitude, referenceJDE)
	if !ok {
		return []SolarEclipsePathPoint{seed}
	}
	branchSign, branchConstrained := 0.0, false
	if constrainBranch {
		branchSign, branchConstrained = solver.magnitudeContourBranchSign(seed)
	}
	for index := range state.tangent {
		state.tangent[index] *= float64(direction)
	}
	points := []SolarEclipsePathPoint{seed}
	step := solarEclipseMagnitudeContourArcStepDegrees
	// Continue the F(magnitude, greatest-time)=0 curve in longitude, latitude, and scaled time.
	for count := 0; count < solarEclipseMagnitudeContourMaxArcSteps; count++ {
		predictor := state.coordinates
		for index := range predictor {
			predictor[index] += step * state.tangent[index]
		}
		next, iterations, nextOK := solver.correctMagnitudeContourArc(
			predictor, state.tangent, magnitude, referenceJDE,
		)
		if !nextOK {
			step /= 2
			if step < solarEclipseMagnitudeContourMinArcStepDegrees {
				break
			}
			continue
		}
		if dotSolarEclipse3(next.tangent, state.tangent) < 0 {
			for index := range next.tangent {
				next.tangent[index] = -next.tangent[index]
			}
		}
		if nextSign, nextConstrained := solver.magnitudeContourBranchSign(next.point); constrainBranch && nextConstrained {
			if branchConstrained && branchSign*nextSign < 0 {
				step /= 2
				if step < solarEclipseMagnitudeContourMinArcStepDegrees {
					break
				}
				continue
			}
			branchSign, branchConstrained = nextSign, true
		}
		distance := solarEclipsePathDistanceKM(state.point, next.point)
		if distance > targetSpacingKM || solarEclipseMagnitudeChordErrorKM(state, next, distance) > 2 {
			step /= 2
			if step < solarEclipseMagnitudeContourMinArcStepDegrees {
				break
			}
			continue
		}
		if next.point.JDE < startJDE-0.05 || next.point.JDE > endJDE+0.05 {
			break
		}
		if next.point.SunAltitude < 0 {
			if endpoint, endpointOK := solver.refineMagnitudeHorizonCrossing(state.point, next.point, magnitude); endpointOK {
				points = append(points, endpoint)
				break
			}
			step /= 2
			if step < solarEclipseMagnitudeContourMinArcStepDegrees {
				break
			}
			continue
		}
		points = append(points, next.point)
		state = next
		if next.point.SunAltitude <= 1e-7 {
			break
		}
		if distance < targetSpacingKM/2 && iterations <= 4 {
			step = math.Min(solarEclipseMagnitudeContourArcStepDegrees, step*1.5)
		}
	}
	return points
}

// Bound the chord error as well as the spacing near a curved grazing limit.
func solarEclipseMagnitudeChordErrorKM(first, second solarEclipseMagnitudeArcState, distance float64) float64 {
	latitude := (first.point.Latitude + second.point.Latitude) * rad / 2
	ax, ay := first.tangent[0]*math.Cos(latitude), first.tangent[1]
	bx, by := second.tangent[0]*math.Cos(latitude), second.tangent[1]
	norm := math.Hypot(ax, ay) * math.Hypot(bx, by)
	if norm == 0 {
		return math.Inf(1)
	}
	cosine := math.Max(-1, math.Min(1, (ax*bx+ay*by)/norm))
	return distance * math.Sqrt(2*(1-cosine)) / 8
}

// magnitudeContourBranchSign identifies which side of the simultaneous
// central path a contour point occupies. A magnitude contour can have two
// nearby roots near a hybrid transition; keeping this sign prevents Newton
// correction from silently switching to the opposite root.
func (solver solarEclipseSolver) magnitudeContourBranchSign(point SolarEclipsePathPoint) (float64, bool) {
	center, centerOK := solver.centralPathPointAt(point.JDE)
	before, beforeOK := solver.centralPathPointAt(point.JDE - solarEclipsePathVelocityStepDays)
	after, afterOK := solver.centralPathPointAt(point.JDE + solarEclipsePathVelocityStepDays)
	if !centerOK || !beforeOK || !afterOK {
		return 0, false
	}
	cosLatitude := math.Cos(center.Latitude * rad)
	pathX := math.Remainder(after.Longitude-before.Longitude, 360) * cosLatitude
	pathY := after.Latitude - before.Latitude
	offsetX := math.Remainder(point.Longitude-center.Longitude, 360) * cosLatitude
	offsetY := point.Latitude - center.Latitude
	sign := pathX*offsetY - pathY*offsetX
	if !finite(sign) || math.Abs(sign) <= 1e-10 {
		return 0, false
	}
	return sign, true
}

func (solver solarEclipseSolver) magnitudeArcStateAt(
	point SolarEclipsePathPoint,
	magnitude, referenceJDE float64,
) (solarEclipseMagnitudeArcState, bool) {
	coordinates := [3]float64{
		point.Longitude,
		point.Latitude,
		(point.JDE - referenceJDE) * solarEclipseMagnitudeContourTimeScale,
	}
	_, jacobian, ok := solver.magnitudeEnvelopeJacobian(coordinates, magnitude, referenceJDE)
	if !ok {
		return solarEclipseMagnitudeArcState{}, false
	}
	tangent, ok := solarEclipseMagnitudeArcTangent(jacobian)
	return solarEclipseMagnitudeArcState{coordinates: coordinates, tangent: tangent, point: point}, ok
}

func (solver solarEclipseSolver) correctMagnitudeContourArc(
	predictor, tangent [3]float64,
	magnitude, referenceJDE float64,
) (solarEclipseMagnitudeArcState, int, bool) {
	coordinates := predictor
	for iteration := 0; iteration < 16; iteration++ {
		if !finite(coordinates[0]) || !finite(coordinates[1]) || !finite(coordinates[2]) {
			return solarEclipseMagnitudeArcState{}, iteration, false
		}
		residual, jacobian, ok := solver.magnitudeEnvelopeJacobian(coordinates, magnitude, referenceJDE)
		if !ok {
			return solarEclipseMagnitudeArcState{}, iteration, false
		}
		planeResidual := dotSolarEclipse3(subtractSolarEclipse3(coordinates, predictor), tangent)
		if math.Abs(residual[0]) <= 1e-10 && math.Abs(residual[1]) <= 1e-10 && math.Abs(planeResidual) <= 1e-9 {
			return solver.validMagnitudeArcState(coordinates, jacobian, magnitude, referenceJDE, iteration+1)
		}
		matrix := [3][3]float64{jacobian[0], jacobian[1], tangent}
		delta, ok := solveSolarEclipse3x3(matrix, [3]float64{-residual[0], -residual[1], -planeResidual})
		if !ok {
			return solarEclipseMagnitudeArcState{}, iteration, false
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
		if coordinates[1] <= -89.999999 || coordinates[1] >= 89.999999 {
			return solarEclipseMagnitudeArcState{}, iteration, false
		}
	}
	residual, jacobian, ok := solver.magnitudeEnvelopeJacobian(coordinates, magnitude, referenceJDE)
	planeResidual := dotSolarEclipse3(subtractSolarEclipse3(coordinates, predictor), tangent)
	if !ok || math.Abs(residual[0]) > 1e-7 || math.Abs(residual[1]) > 1e-8 || math.Abs(planeResidual) > 1e-7 {
		return solarEclipseMagnitudeArcState{}, 16, false
	}
	return solver.validMagnitudeArcState(coordinates, jacobian, magnitude, referenceJDE, 16)
}

func (solver solarEclipseSolver) validMagnitudeArcState(
	coordinates [3]float64,
	jacobian [2][3]float64,
	magnitude, referenceJDE float64,
	iterations int,
) (solarEclipseMagnitudeArcState, int, bool) {
	jde := referenceJDE + coordinates[2]/solarEclipseMagnitudeContourTimeScale
	longitude := geodata.NormalizeLongitude(coordinates[0])
	latitude := coordinates[1]
	evaluation := solver.magnitudeEvaluationAt(jde)
	state := evaluation.center.stateAt(longitude*rad, latitude*rad, 0)
	if math.Abs(solarEclipseMagnitudeAtTarget(state, magnitude)-magnitude) > 1e-7 ||
		evaluation.separationSecondDerivative(longitude, latitude) <= 0 {
		return solarEclipseMagnitudeArcState{}, iterations, false
	}
	tangent, ok := solarEclipseMagnitudeArcTangent(jacobian)
	if !ok {
		return solarEclipseMagnitudeArcState{}, iterations, false
	}
	return solarEclipseMagnitudeArcState{
		coordinates: coordinates,
		tangent:     tangent,
		point: SolarEclipsePathPoint{
			JDE: jde, Longitude: longitude, Latitude: latitude, SunAltitude: state.sunAltitudeRad / rad,
		},
	}, iterations, true
}

func (solver solarEclipseSolver) magnitudeEnvelopeJacobian(
	coordinates [3]float64,
	magnitude, referenceJDE float64,
) ([2]float64, [2][3]float64, bool) {
	jde := referenceJDE + coordinates[2]/solarEclipseMagnitudeContourTimeScale
	longitude := geodata.NormalizeLongitude(coordinates[0])
	latitude := coordinates[1]
	evaluation := solver.magnitudeCandidateEvaluationAt(jde)
	residual, ok := solarEclipseMagnitudeEnvelopeResidualAt(evaluation, longitude, latitude, magnitude)
	if !ok {
		return [2]float64{}, [2][3]float64{}, false
	}
	steps := [3]float64{1e-4, 1e-4, 5.0 * solarEclipseMagnitudeContourTimeScale / 86400.0}
	jacobian := [2][3]float64{}
	spatialCoordinates := [][2]float64{{longitude + steps[0], latitude}, {longitude, latitude + steps[1]}}
	for column, shifted := range spatialCoordinates {
		shiftedResidual, shiftedOK := solarEclipseMagnitudeEnvelopeResidualAt(
			evaluation, shifted[0], shifted[1], magnitude,
		)
		if !shiftedOK {
			return [2]float64{}, [2][3]float64{}, false
		}
		for row := 0; row < 2; row++ {
			jacobian[row][column] = (shiftedResidual[row] - residual[row]) / steps[column]
		}
	}
	timeEvaluation := solver.magnitudeCandidateEvaluationAt(jde + steps[2]/solarEclipseMagnitudeContourTimeScale)
	timeResidual, timeOK := solarEclipseMagnitudeEnvelopeResidualAt(
		timeEvaluation, longitude, latitude, magnitude,
	)
	if !timeOK {
		return [2]float64{}, [2][3]float64{}, false
	}
	beforeEvaluation := solver.magnitudeCandidateEvaluationAt(jde - steps[2]/solarEclipseMagnitudeContourTimeScale)
	beforeResidual, beforeOK := solarEclipseMagnitudeEnvelopeResidualAt(
		beforeEvaluation, longitude, latitude, magnitude,
	)
	if !beforeOK {
		return [2]float64{}, [2][3]float64{}, false
	}
	for row := 0; row < 2; row++ {
		jacobian[row][2] = (timeResidual[row] - beforeResidual[row]) / (2 * steps[2])
	}
	return residual, jacobian, true
}

func solarEclipseMagnitudeEnvelopeResidualAt(
	evaluation solarEclipseRiseSetEvaluation,
	longitude, latitude, magnitude float64,
) ([2]float64, bool) {
	state := evaluation.center.stateAt(longitude*rad, latitude*rad, 0)
	residual := [2]float64{
		solarEclipseMagnitudeAtTarget(state, magnitude) - magnitude,
		evaluation.separationDerivative(longitude, latitude),
	}
	return residual, finite(residual[0]) && finite(residual[1])
}

func solarEclipseMagnitudeArcTangent(jacobian [2][3]float64) ([3]float64, bool) {
	tangent := [3]float64{
		jacobian[0][1]*jacobian[1][2] - jacobian[0][2]*jacobian[1][1],
		jacobian[0][2]*jacobian[1][0] - jacobian[0][0]*jacobian[1][2],
		jacobian[0][0]*jacobian[1][1] - jacobian[0][1]*jacobian[1][0],
	}
	norm := math.Sqrt(dotSolarEclipse3(tangent, tangent))
	if !finite(norm) || norm < 1e-14 {
		return [3]float64{}, false
	}
	for index := range tangent {
		tangent[index] /= norm
	}
	return tangent, true
}

func dotSolarEclipse3(first, second [3]float64) float64 {
	return first[0]*second[0] + first[1]*second[1] + first[2]*second[2]
}

func subtractSolarEclipse3(first, second [3]float64) [3]float64 {
	return [3]float64{first[0] - second[0], first[1] - second[1], first[2] - second[2]}
}

func (solver solarEclipseSolver) refineMagnitudeHorizonCrossing(
	visible, hidden SolarEclipsePathPoint,
	magnitude float64,
) (SolarEclipsePathPoint, bool) {
	fraction := visible.SunAltitude / (visible.SunAltitude - hidden.SunAltitude)
	deltaLongitude := math.Remainder(hidden.Longitude-visible.Longitude, 360)
	return solver.refineMagnitudeHorizonPoint(
		visible.JDE+fraction*(hidden.JDE-visible.JDE),
		geodata.NormalizeLongitude(visible.Longitude+fraction*deltaLongitude),
		visible.Latitude+fraction*(hidden.Latitude-visible.Latitude),
		magnitude,
	)
}

func (solver solarEclipseSolver) refineMagnitudeHorizonPoint(
	jde, longitude, latitude, magnitude float64,
) (SolarEclipsePathPoint, bool) {
	const (
		geographicStep = 1e-4
		timeStep       = 5.0 / 86400.0
	)
	for iteration := 0; iteration < 24; iteration++ {
		evaluation := solver.magnitudeEvaluationAt(jde)
		residual, ok := solarEclipseMagnitudeHorizonResidualAt(evaluation, longitude, latitude, magnitude)
		if !ok {
			return SolarEclipsePathPoint{}, false
		}
		if math.Abs(residual[0]) <= 1e-10 && math.Abs(residual[1]) <= 1e-10 && math.Abs(residual[2]) <= 1e-10 {
			break
		}
		longitudeResidual, lonOK := solarEclipseMagnitudeHorizonResidualAt(
			evaluation, longitude+geographicStep, latitude, magnitude,
		)
		latitudeResidual, latOK := solarEclipseMagnitudeHorizonResidualAt(
			evaluation, longitude, latitude+geographicStep, magnitude,
		)
		timeResidual, timeOK := solver.magnitudeHorizonResidual(jde+timeStep, longitude, latitude, magnitude)
		if !lonOK || !latOK || !timeOK {
			return SolarEclipsePathPoint{}, false
		}
		matrix := [3][3]float64{}
		for row := 0; row < 3; row++ {
			matrix[row][0] = (longitudeResidual[row] - residual[row]) / geographicStep
			matrix[row][1] = (latitudeResidual[row] - residual[row]) / geographicStep
			matrix[row][2] = (timeResidual[row] - residual[row]) / timeStep
		}
		delta, ok := solveSolarEclipse3x3(matrix, [3]float64{-residual[0], -residual[1], -residual[2]})
		if !ok {
			return SolarEclipsePathPoint{}, false
		}
		geographicScale := math.Max(math.Abs(delta[0]), math.Abs(delta[1]))
		if geographicScale > 2 {
			delta[0] *= 2 / geographicScale
			delta[1] *= 2 / geographicScale
		}
		if math.Abs(delta[2]) > 10.0/1440.0 {
			delta[2] = math.Copysign(10.0/1440.0, delta[2])
		}
		longitude = geodata.NormalizeLongitude(longitude + delta[0])
		latitude += delta[1]
		jde += delta[2]
	}
	residual, ok := solver.magnitudeHorizonResidual(jde, longitude, latitude, magnitude)
	if !ok || math.Abs(residual[0]) > 1e-7 || math.Abs(residual[1]) > 1e-8 || math.Abs(residual[2]) > 1e-7 {
		return SolarEclipsePathPoint{}, false
	}
	evaluation := solver.magnitudeEvaluationAt(jde)
	if evaluation.separationSecondDerivative(longitude, latitude) <= 0 {
		return SolarEclipsePathPoint{}, false
	}
	return SolarEclipsePathPoint{
		JDE: jde, Longitude: longitude, Latitude: latitude, SunAltitude: residual[2] / rad,
	}, true
}

func (solver solarEclipseSolver) magnitudeEvaluationAt(jde float64) solarEclipseRiseSetEvaluation {
	return solarEclipseRiseSetEvaluation{
		jd:     jde,
		center: solver.localStateContextAt(jde),
		before: solver.localStateContextAt(jde - solarEclipseRiseSetDerivativeStepDays),
		after:  solver.localStateContextAt(jde + solarEclipseRiseSetDerivativeStepDays),
	}
}

func (solver solarEclipseSolver) magnitudeCandidateEvaluationAt(jde float64) solarEclipseRiseSetEvaluation {
	return solarEclipseRiseSetEvaluation{
		jd:     jde,
		center: solver.localStateContextCandidateAt(jde),
		before: solver.localStateContextCandidateAt(jde - solarEclipseRiseSetDerivativeStepDays),
		after:  solver.localStateContextCandidateAt(jde + solarEclipseRiseSetDerivativeStepDays),
	}
}

func (solver solarEclipseSolver) localStateContextAt(jde float64) localSolarEclipseStateContext {
	if solver.localStateContextCache == nil {
		return newLocalSolarEclipseStateContextWithOverride(jde, solver.deltaTSeconds, solver.params, solver.inst)
	}
	key := math.Float64bits(jde)
	if context, ok := solver.localStateContextCache[key]; ok &&
		context.generation == deltaTGenerationValue() {
		return context
	}
	context := newLocalSolarEclipseStateContextWithOverride(jde, solver.deltaTSeconds, solver.params, solver.inst)
	return storeLocalSolarEclipseStateContext(solver.localStateContextCache, key, context)
}

func storeLocalSolarEclipseStateContext(
	cache map[uint64]localSolarEclipseStateContext,
	key uint64,
	context localSolarEclipseStateContext,
) localSolarEclipseStateContext {
	if cache == nil {
		return context
	}
	// 与贝塞尔几何缓存同口径：事件级缓存不维护淘汰表，超过上限整体作废更可预测。
	if _, exists := cache[key]; !exists && len(cache) >= solarEclipseBesselGeometryCacheMaximumEntries {
		for cachedKey := range cache {
			delete(cache, cachedKey)
		}
	}
	context.generation = deltaTGenerationValue()
	cache[key] = context
	return context
}

func (solver solarEclipseSolver) localStateContextCandidateAt(jde float64) localSolarEclipseStateContext {
	// 有月缘时坐标架也取插值表；插值表缺失或越界时退回精确上下文。
	if solver.localEphemeris == nil || (solver.params.lunarLimb != nil && solver.limbAxes == nil) {
		return solver.localStateContextAt(jde)
	}
	sun, moon, ok := solver.localEphemeris.equatorialAt(jde)
	if !ok {
		return solver.localStateContextAt(jde)
	}
	context := localSolarEclipseStateContext{
		sunXYZ:  solarEclipseLLRToXYZ(sun[0], sun[1], sun[2]),
		moonXYZ: solarEclipseLLRToXYZ(moon[0], moon[1], moon[2]),
		gst:     solver.siderealTimeAt(jde),
		params:  solver.params,
	}
	if solver.params.lunarLimb != nil {
		frame, ok := solver.limbAxes.frameAt(jde)
		if !ok {
			return solver.localStateContextAt(jde)
		}
		context.limbFrame = &frame
	}
	return context
}

func (solver solarEclipseSolver) magnitudeHorizonResidual(
	jde, longitude, latitude, magnitude float64,
) ([3]float64, bool) {
	evaluation := solver.magnitudeEvaluationAt(jde)
	return solarEclipseMagnitudeHorizonResidualAt(evaluation, longitude, latitude, magnitude)
}

func solarEclipseMagnitudeHorizonResidualAt(
	evaluation solarEclipseRiseSetEvaluation,
	longitude, latitude, magnitude float64,
) ([3]float64, bool) {
	state := evaluation.center.stateAt(longitude*rad, latitude*rad, 0)
	value := [3]float64{
		solarEclipseMagnitudeAtTarget(state, magnitude) - magnitude,
		evaluation.separationDerivative(longitude, latitude),
		state.sunAltitudeRad,
	}
	return value, finite(value[0]) && finite(value[1]) && finite(value[2])
}

func solveSolarEclipse3x3(matrix [3][3]float64, right [3]float64) ([3]float64, bool) {
	augmented := [3][4]float64{}
	for row := 0; row < 3; row++ {
		copy(augmented[row][:3], matrix[row][:])
		augmented[row][3] = right[row]
	}
	for column := 0; column < 3; column++ {
		pivot := column
		for row := column + 1; row < 3; row++ {
			if math.Abs(augmented[row][column]) > math.Abs(augmented[pivot][column]) {
				pivot = row
			}
		}
		if !finite(augmented[pivot][column]) || math.Abs(augmented[pivot][column]) < 1e-18 {
			return [3]float64{}, false
		}
		augmented[column], augmented[pivot] = augmented[pivot], augmented[column]
		for row := column + 1; row < 3; row++ {
			factor := augmented[row][column] / augmented[column][column]
			for index := column; index < 4; index++ {
				augmented[row][index] -= factor * augmented[column][index]
			}
		}
	}
	result := [3]float64{}
	for row := 2; row >= 0; row-- {
		value := augmented[row][3]
		for column := row + 1; column < 3; column++ {
			value -= augmented[row][column] * result[column]
		}
		result[row] = value / augmented[row][row]
		if !finite(result[row]) {
			return [3]float64{}, false
		}
	}
	return result, true
}

func (solver solarEclipseSolver) magnitudeContourPointsAt(jde, magnitude float64) []SolarEclipsePathPoint {
	moon := solver.besselMoonAt(jde)
	axis := solver.besselAxisAt(jde)
	evaluation := solver.magnitudeEvaluationAt(jde)
	valueAt := func(angle float64) (float64, bool) {
		point, ok := solver.magnitudeContourPointAt(
			jde, moon, axis, math.Cos(angle), math.Sin(angle), magnitude,
		)
		if !ok {
			return 0, false
		}
		return evaluation.separationDerivative(point.Longitude, point.Latitude), true
	}
	points := make([]SolarEclipsePathPoint, 0, 2)
	for _, angle := range riseSetCyclicRoots(solarEclipseMagnitudeContourBoundaryPoints, valueAt) {
		seed, ok := solver.magnitudeContourPointAt(
			jde, moon, axis, math.Cos(angle), math.Sin(angle), magnitude,
		)
		if !ok {
			continue
		}
		longitude, latitude, ok := solver.refineMagnitudeEnvelopePoint(
			magnitude, seed.Longitude, seed.Latitude, evaluation,
		)
		if !ok {
			continue
		}
		state := evaluation.center.stateAt(longitude*rad, latitude*rad, 0)
		if state.sunAltitudeRad < -1e-7 || evaluation.separationSecondDerivative(longitude, latitude) <= 0 {
			continue
		}
		point := SolarEclipsePathPoint{
			JDE: jde, Longitude: longitude, Latitude: latitude, SunAltitude: state.sunAltitudeRad / rad,
		}
		if !solarEclipseRiseSetPointExists(points, point) {
			points = append(points, point)
		}
	}
	if magnitude == 1 && len(points) < 2 {
		points = solver.appendMagnitudeOneLimitSeeds(points, jde, evaluation)
	}
	if len(points) == 0 {
		points = solver.magnitudeContourGeographicSeedsAt(jde, magnitude)
	}
	return points
}

func (solver solarEclipseSolver) appendMagnitudeOneLimitSeeds(
	points []SolarEclipsePathPoint,
	jde float64,
	evaluation solarEclipseRiseSetEvaluation,
) []SolarEclipsePathPoint {
	center, ok := solver.centralPathPointAt(jde)
	if !ok {
		return points
	}
	first, second, ok := solver.centralPathLimitsAt(center)
	if !ok {
		return points
	}
	for _, seed := range []SolarEclipsePathPoint{first, second} {
		longitude, latitude, refined := solver.refineMagnitudeEnvelopePoint(
			1, seed.Longitude, seed.Latitude, evaluation,
		)
		if !refined {
			continue
		}
		state := evaluation.center.stateAt(longitude*rad, latitude*rad, 0)
		if state.sunAltitudeRad < -1e-7 || evaluation.separationSecondDerivative(longitude, latitude) <= 0 {
			continue
		}
		candidate := SolarEclipsePathPoint{
			JDE: jde, Longitude: longitude, Latitude: latitude, SunAltitude: state.sunAltitudeRad / rad,
		}
		if !solarEclipseRiseSetPointExists(points, candidate) {
			points = append(points, candidate)
		}
	}
	return points
}

// magnitudeContourGeographicSeedsAt supplies seeds for contours inside the
// signed Bessel umbra. The cone-radius interpolation is well-conditioned for
// m<=1, but it can become negative for a legitimate m>1 contour near a
// total/annular transition. Solving the local-magnitude envelope from the
// greatest point keeps those contours available without changing the normal
// Bessel path.
func (solver solarEclipseSolver) magnitudeContourGeographicSeedsAt(jde, magnitude float64) []SolarEclipsePathPoint {
	center, ok := solver.centralPathPointAt(jde)
	if !ok {
		// A non-central eclipse has no Earth-intersecting shadow axis, but its
		// local maximum still has a well-defined geographic stationary point.
		// Reuse the global greatest-eclipse coordinates as the bounded seed for
		// the local envelope solver instead of treating the missing central axis
		// as evidence that every magnitude contour is absent.
		result := solarEclipse(solver.newMoonJDE, solver.model, solver.inst)
		if !result.HasPartial || result.GreatestEclipse == 0 {
			return nil
		}
		center = SolarEclipsePathPoint{
			JDE:       jde,
			Longitude: result.GreatestLongitude,
			Latitude:  result.GreatestLatitude,
		}
	}
	evaluation := solver.magnitudeEvaluationAt(jde)
	centerState := evaluation.center.stateAt(center.Longitude*rad, center.Latitude*rad, 0)
	maximum := solarEclipseMagnitudeAtTarget(centerState, magnitude)
	if !finite(maximum) || maximum <= magnitude+1e-9 {
		return nil
	}

	centerLon, centerLat := center.Longitude*rad, center.Latitude*rad
	centerVector := [3]float64{
		math.Cos(centerLat) * math.Cos(centerLon),
		math.Cos(centerLat) * math.Sin(centerLon),
		math.Sin(centerLat),
	}
	east := [3]float64{-math.Sin(centerLon), math.Cos(centerLon), 0}
	north := [3]float64{
		-math.Sin(centerLat) * math.Cos(centerLon),
		-math.Sin(centerLat) * math.Sin(centerLon),
		math.Cos(centerLat),
	}
	distanceDegrees := []float64{0.02, 0.05, 0.1, 0.2, 0.4, 0.8, 1.6, 3.2, 6.4, 12.8, 25.6, 51.2}
	seeds := make([]SolarEclipsePathPoint, 0, 2)
	for bearingIndex := 0; bearingIndex < solarEclipseMagnitudeContourFallbackBearings; bearingIndex++ {
		bearing := 2 * math.Pi * float64(bearingIndex) / float64(solarEclipseMagnitudeContourFallbackBearings)
		direction := [3]float64{
			math.Cos(bearing)*north[0] + math.Sin(bearing)*east[0],
			math.Cos(bearing)*north[1] + math.Sin(bearing)*east[1],
			math.Cos(bearing)*north[2] + math.Sin(bearing)*east[2],
		}
		previousDistance := 0.0
		previousValue := maximum - magnitude
		for distanceIndex := 1; distanceIndex < solarEclipseMagnitudeContourFallbackDistances; distanceIndex++ {
			distance := distanceDegrees[distanceIndex]
			angle := distance * rad
			pointVector := [3]float64{
				centerVector[0]*math.Cos(angle) + direction[0]*math.Sin(angle),
				centerVector[1]*math.Cos(angle) + direction[1]*math.Sin(angle),
				centerVector[2]*math.Cos(angle) + direction[2]*math.Sin(angle),
			}
			longitude := geodata.NormalizeLongitude(math.Atan2(pointVector[1], pointVector[0]) / rad)
			latitude := math.Asin(math.Max(-1, math.Min(1, pointVector[2]))) / rad
			state := evaluation.center.stateAt(longitude*rad, latitude*rad, 0)
			value := solarEclipseMagnitudeAtTarget(state, magnitude) - magnitude
			if !finite(value) {
				previousDistance = distance
				previousValue = math.NaN()
				continue
			}
			if finite(previousValue) && previousValue >= 0 && value <= 0 {
				left, right := previousDistance, distance
				leftValue := previousValue
				for iteration := 0; iteration < 48 && right-left > 1e-10; iteration++ {
					middle := (left + right) / 2
					middleAngle := middle * rad
					middleVector := [3]float64{
						centerVector[0]*math.Cos(middleAngle) + direction[0]*math.Sin(middleAngle),
						centerVector[1]*math.Cos(middleAngle) + direction[1]*math.Sin(middleAngle),
						centerVector[2]*math.Cos(middleAngle) + direction[2]*math.Sin(middleAngle),
					}
					middleLongitude := geodata.NormalizeLongitude(math.Atan2(middleVector[1], middleVector[0]) / rad)
					middleLatitude := math.Asin(math.Max(-1, math.Min(1, middleVector[2]))) / rad
					middleState := evaluation.center.stateAt(middleLongitude*rad, middleLatitude*rad, 0)
					middleValue := solarEclipseMagnitudeAtTarget(middleState, magnitude) - magnitude
					if !finite(middleValue) {
						break
					}
					if leftValue*middleValue <= 0 {
						right = middle
					} else {
						left, leftValue = middle, middleValue
					}
				}
				seedDistance := (left + right) / 2
				seedAngle := seedDistance * rad
				seedVector := [3]float64{
					centerVector[0]*math.Cos(seedAngle) + direction[0]*math.Sin(seedAngle),
					centerVector[1]*math.Cos(seedAngle) + direction[1]*math.Sin(seedAngle),
					centerVector[2]*math.Cos(seedAngle) + direction[2]*math.Sin(seedAngle),
				}
				seedLongitude := geodata.NormalizeLongitude(math.Atan2(seedVector[1], seedVector[0]) / rad)
				seedLatitude := math.Asin(math.Max(-1, math.Min(1, seedVector[2]))) / rad
				seedLongitude, seedLatitude, refined := solver.refineMagnitudeEnvelopePoint(
					magnitude, seedLongitude, seedLatitude, evaluation,
				)
				if refined {
					state := evaluation.center.stateAt(seedLongitude*rad, seedLatitude*rad, 0)
					candidate := SolarEclipsePathPoint{
						JDE: jde, Longitude: seedLongitude, Latitude: seedLatitude,
						SunAltitude: state.sunAltitudeRad / rad,
					}
					if !solarEclipseRiseSetPointExists(seeds, candidate) {
						seeds = append(seeds, candidate)
					}
				}
				break
			}
			previousDistance, previousValue = distance, value
		}
	}
	return seeds
}

func (solver solarEclipseSolver) refineMagnitudeEnvelopePoint(
	magnitude, longitude, latitude float64,
	evaluation solarEclipseRiseSetEvaluation,
) (float64, float64, bool) {
	return riseSetRefineGeographicRoot(longitude, latitude, func(lon, lat float64) (float64, float64, bool) {
		state := evaluation.center.stateAt(lon*rad, lat*rad, 0)
		instantaneousMagnitude := solarEclipseMagnitudeAtTarget(state, magnitude)
		return instantaneousMagnitude - magnitude, evaluation.separationDerivative(lon, lat),
			finite(instantaneousMagnitude)
	})
}

func solarEclipseMagnitudeContourCompatibilitySides(
	segments [][]SolarEclipsePathPoint,
) ([]SolarEclipsePathPoint, []SolarEclipsePathPoint) {
	if len(segments) < 2 {
		return nil, nil
	}
	indices := []int{0, 1}
	for index := 2; index < len(segments); index++ {
		if len(segments[index]) <= len(segments[indices[1]]) {
			continue
		}
		indices[1] = index
		if len(segments[indices[1]]) > len(segments[indices[0]]) {
			indices[0], indices[1] = indices[1], indices[0]
		}
	}
	first, second := segments[indices[0]], segments[indices[1]]
	firstLatitude := first[len(first)/2].Latitude
	secondLatitude := second[len(second)/2].Latitude
	if secondLatitude > firstLatitude {
		first, second = second, first
	}
	return first, second
}

func solarEclipseLocalMagnitude(state localSolarEclipseState) float64 {
	if state.sunRadiusRad <= 0 {
		return math.NaN()
	}
	return (state.moonOuterRadiusRad + state.sunRadiusRad - state.separationRad) / (2 * state.sunRadiusRad)
}

func solarEclipseMagnitudeAtTarget(state localSolarEclipseState, target float64) float64 {
	if state.sunRadiusRad <= 0 {
		return math.NaN()
	}
	if target > 1 {
		centralGap := state.moonInnerRadiusRad - state.sunRadiusRad
		if centralGap <= 0 {
			return math.NaN()
		}
		// Inside totality, normalize the magnitude from 1 at the inner
		// contact to the apparent-diameter ratio at zero separation. This
		// preserves the usual m=1 boundary while retaining legal values
		// above one for deep total eclipses.
		return 1 + (state.moonInnerRadiusRad/state.sunRadiusRad-1)*
			(1-state.separationRad/centralGap)
	}
	moonRadius := state.moonOuterRadiusRad
	if target == 1 {
		moonRadius = state.moonInnerRadiusRad
	}
	return (moonRadius + state.sunRadiusRad - state.separationRad) / (2 * state.sunRadiusRad)
}

func (solver solarEclipseSolver) magnitudeContourPointAt(
	jde float64,
	moon [3]float64,
	axis solarEclipseAxis,
	directionX, directionY, magnitude float64,
) (SolarEclipsePathPoint, bool) {
	if magnitude == 0 {
		_, _, sun := solver.besselGeometryAt(jde)
		return solver.shadowFootprintPointAt(jde, moon, axis, sun, math.Atan2(directionY, directionX), solarEclipsePenumbralShadow)
	}
	radii := solver.shadowRadiiAt(moon[2])
	radius := solarEclipseMagnitudeContourRadius(radii, magnitude)
	if magnitude > 1 && radius <= 0 {
		// The initial Bessel plane can be on the antumbral side even though
		// the Earth intersection has a valid totality contour. Start from
		// the absolute umbral edge and let the surface iteration converge.
		radius = radii.absUmbraRadius
	}
	if radius <= 0 {
		return SolarEclipsePathPoint{}, false
	}
	var intersection solarEclipseLineIntersection
	for iteration := 0; iteration < solarEclipsePartialFootprintIterationLimit; iteration++ {
		x := moon[0] + radius*directionX
		y := moon[1] + radius*directionY
		intersection = solarEclipseLineEar2(
			x, y, 2,
			x, y, 0,
			solarEclipseEarthPolarRatio, 1, axis,
		)
		if !intersection.valid {
			return SolarEclipsePathPoint{}, false
		}
		nextRadii := solver.shadowRadiiAt(moon[2] - intersection.r2)
		nextRadius := solarEclipseMagnitudeContourRadius(nextRadii, magnitude)
		if magnitude > 1 && nextRadius <= 0 {
			nextRadius = nextRadii.absUmbraRadius
		}
		if nextRadius <= 0 {
			return SolarEclipsePathPoint{}, false
		}
		if math.Abs(nextRadius-radius) <= solarEclipsePartialFootprintPointTolerance {
			radius = nextRadius
			break
		}
		radius = nextRadius
	}
	x := moon[0] + radius*directionX
	y := moon[1] + radius*directionY
	intersection = solarEclipseLineEar2(
		x, y, 2,
		x, y, 0,
		solarEclipseEarthPolarRatio, 1, axis,
	)
	if !intersection.valid {
		return SolarEclipsePathPoint{}, false
	}
	longitude, latitude := solarEclipseIntersectionGeodetic(intersection, axis)
	sunAltitudeRad := solarEclipseSunAltitudeAtGreatest(jde, longitude, latitude, axis.gst, solver.inst)
	return SolarEclipsePathPoint{
		JDE:         jde,
		Longitude:   longitude,
		Latitude:    latitude,
		SunAltitude: sunAltitudeRad / rad,
	}, true
}

func solarEclipseMagnitudeContourRadius(radii solarEclipseShadowRadii, magnitude float64) float64 {
	if magnitude >= 1 && radii.magnitude > 1 {
		// Above totality, interpolate from the umbral edge (m=1) to the
		// Bessel-axis maximum (m=radii.magnitude). Using the penumbra-to-
		// umbra slope here can turn a valid high-magnitude contour negative
		// near hybrid and shallow total eclipses.
		return radii.absUmbraRadius * (radii.magnitude - magnitude) / (radii.magnitude - 1)
	}
	return radii.penumbraRadius - magnitude*(radii.penumbraRadius-radii.absUmbraRadius)
}
