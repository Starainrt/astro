package basic

import "math"

const (
	solarEclipseCentralEnvelopeArcStepDegrees    = 0.4
	solarEclipseCentralEnvelopeMinArcStepDegrees = 0.01
	solarEclipseCentralEnvelopeMaxArcSteps       = 2000
	solarEclipseCentralEnvelopeMaxSpacingKM      = 48.0
	solarEclipseCentralEnvelopeEndDistanceKM     = 50.0
	solarEclipseTotalEnvelopeTargetSpacingKM     = 200.0
)

// centralTwoLimitBandEnvelope follows both continuous local-centrality limits
// between the solved sunrise and sunset closures. Unlike a union of discrete
// shadow footprints, every non-horizon point satisfies C=0 and dC/dt=0.
func (solver solarEclipseSolver) centralTwoLimitBandEnvelope(
	closures [][]SolarEclipsePathPoint,
	referenceJDE float64,
) [][]SolarEclipsePathPoint {
	branches := solver.centralBandVectorBranches(closures, referenceJDE)
	if len(branches) != 2 {
		return nil
	}
	return joinSolarEclipseCentralBandBranches(branches, closures)
}

func (solver solarEclipseSolver) centralBandVectorBranches(
	closures [][]SolarEclipsePathPoint,
	referenceJDE float64,
) [][]SolarEclipsePathPoint {
	if len(closures) != 2 || len(closures[0]) < 2 || len(closures[1]) < 2 {
		return nil
	}
	startRoots := []SolarEclipsePathPoint{closures[0][0], closures[0][len(closures[0])-1]}
	endRoots := []SolarEclipsePathPoint{closures[1][0], closures[1][len(closures[1])-1]}
	branches := make([][]SolarEclipsePathPoint, 2)
	endIndices := [2]int{-1, -1}
	solver = solver.withLocalEphemeris()
	for index, root := range startRoots {
		var transitions []SolarEclipsePathPoint
		branch, endIndex, ok := solver.traceCentralBandVectorEnvelope(root, endRoots, &transitions, referenceJDE)
		if !ok {
			return nil
		}
		branches[index], endIndices[index] = branch, endIndex
	}
	if endIndices[0] == endIndices[1] {
		return nil
	}
	return branches
}

func solarEclipseTotalBandEnvelope(
	segments, closures [][]SolarEclipsePathPoint,
) [][]SolarEclipsePathPoint {
	if len(segments) != 2 || len(segments[0]) < 2 || len(segments[1]) < 2 {
		return nil
	}
	branches := make([][]SolarEclipsePathPoint, 2)
	for index, segment := range segments {
		branch := append([]SolarEclipsePathPoint(nil), segment...)
		if branch[0].JDE > branch[len(branch)-1].JDE {
			for left, right := 0, len(branch)-1; left < right; left, right = left+1, right-1 {
				branch[left], branch[right] = branch[right], branch[left]
			}
		}
		branches[index] = branch
	}
	if len(closures) == 2 && len(closures[0]) > 1 && len(closures[1]) > 1 &&
		solarEclipsePathDistanceKM(closures[0][0], closures[1][0]) < 0.01 &&
		solarEclipsePathDistanceKM(closures[0][len(closures[0])-1], closures[1][len(closures[1])-1]) < 0.01 {
		ring := append([]SolarEclipsePathPoint{}, branches[0]...)
		for index := len(branches[1]) - 1; index >= 0; index-- {
			ring = append(ring, branches[1][index])
		}
		ring = deduplicateSolarEclipsePathPoints(ring)
		if len(ring) >= 4 {
			ring = append(ring, ring[0])
			return [][]SolarEclipsePathPoint{ring}
		}
	}
	return joinSolarEclipseCentralBandBranches(branches, closures)
}

func joinSolarEclipseCentralBandBranches(
	branches, closures [][]SolarEclipsePathPoint,
) [][]SolarEclipsePathPoint {
	if len(branches) != 2 || len(branches[0]) == 0 || len(branches[1]) == 0 ||
		len(closures) != 2 || len(closures[0]) == 0 || len(closures[1]) == 0 {
		return nil
	}
	endClosure, ok := orientSolarEclipsePath(
		closures[1], branches[0][len(branches[0])-1], branches[1][len(branches[1])-1],
	)
	if !ok {
		return nil
	}
	startClosure, ok := orientSolarEclipsePath(
		closures[0], branches[1][0], branches[0][0],
	)
	if !ok {
		return nil
	}

	ring := make([]SolarEclipsePathPoint, 0,
		len(branches[0])+len(branches[1])+len(startClosure)+len(endClosure),
	)
	ring = append(ring, branches[0]...)
	ring = append(ring, endClosure[1:]...)
	for index := len(branches[1]) - 2; index >= 0; index-- {
		ring = append(ring, branches[1][index])
	}
	ring = append(ring, startClosure[1:]...)
	ring = deduplicateSolarEclipsePathPoints(ring)
	if len(ring) < 4 || solarEclipsePathDistanceKM(ring[0], ring[len(ring)-1]) > 0.01 {
		return nil
	}
	ring[len(ring)-1] = ring[0]
	return [][]SolarEclipsePathPoint{ring}
}

func (solver solarEclipseSolver) traceCentralBandEnvelope(
	root SolarEclipsePathPoint,
	endRoots []SolarEclipsePathPoint,
	referenceJDE float64,
) ([]SolarEclipsePathPoint, int, bool) {
	state, ok := solver.nonCentralBandStateAt(root, referenceJDE)
	if !ok {
		return nil, -1, false
	}
	step := solarEclipseCentralEnvelopeArcStepDegrees
	next, ok := solver.centralBandEnvelopeVisibleStart(state, step, referenceJDE)
	if !ok {
		return []SolarEclipsePathPoint{root}, -1, false
	}
	if dotSolarEclipse3(next.tangent, subtractSolarEclipse3(next.coordinates, state.coordinates)) < 0 {
		for index := range next.tangent {
			next.tangent[index] = -next.tangent[index]
		}
	}
	points := []SolarEclipsePathPoint{root, next.point}
	state = next
	for count := 0; count < solarEclipseCentralEnvelopeMaxArcSteps; count++ {
		predictor := state.coordinates
		for index := range predictor {
			predictor[index] += step * state.tangent[index]
		}
		candidate, iterations, candidateOK := solver.correctNonCentralBandBoundary(
			predictor, state.tangent, referenceJDE,
		)
		if !candidateOK {
			step /= 2
			if step < solarEclipseCentralEnvelopeMinArcStepDegrees {
				return points, -1, false
			}
			continue
		}
		if dotSolarEclipse3(candidate.tangent, state.tangent) < 0 {
			for index := range candidate.tangent {
				candidate.tangent[index] = -candidate.tangent[index]
			}
		}
		distance := solarEclipsePathDistanceKM(state.point, candidate.point)
		if distance > solarEclipseCentralEnvelopeMaxSpacingKM {
			step /= 2
			if step < solarEclipseCentralEnvelopeMinArcStepDegrees {
				return points, -1, false
			}
			continue
		}
		if candidate.point.SunAltitude < 0 {
			endIndex, endDistance := nearestSolarEclipsePathPoint(state.point, candidate.point, endRoots)
			if endIndex < 0 || endDistance > solarEclipseCentralEnvelopeEndDistanceKM {
				return points, endIndex, false
			}
			points = append(points, endRoots[endIndex])
			return points, endIndex, true
		}
		points = append(points, candidate.point)
		state = candidate
		if distance < solarEclipseCentralEnvelopeMaxSpacingKM/2 && iterations <= 4 {
			step = math.Min(solarEclipseCentralEnvelopeArcStepDegrees, step*1.5)
		}
	}
	return points, -1, false
}

func (solver solarEclipseSolver) centralBandEnvelopeVisibleStart(
	state solarEclipseNonCentralBandState,
	step, referenceJDE float64,
) (solarEclipseNonCentralBandState, bool) {
	best := solarEclipseNonCentralBandState{}
	found := false
	for trialStep := step; trialStep >= solarEclipseCentralEnvelopeMinArcStepDegrees; trialStep /= 2 {
		for _, direction := range []float64{1, -1} {
			predictor := state.coordinates
			tangent := state.tangent
			for index := range predictor {
				tangent[index] *= direction
				predictor[index] += trialStep * tangent[index]
			}
			candidate, _, ok := solver.correctNonCentralBandBoundary(predictor, tangent, referenceJDE)
			if ok && candidate.point.SunAltitude > 0 &&
				(!found || candidate.point.SunAltitude > best.point.SunAltitude) {
				best, found = candidate, true
			}
		}
		if found {
			break
		}
	}
	return best, found
}

func nearestSolarEclipsePathPoint(
	first, second SolarEclipsePathPoint,
	points []SolarEclipsePathPoint,
) (int, float64) {
	bestIndex, bestDistance := -1, math.Inf(1)
	for index, point := range points {
		distance := math.Min(
			solarEclipsePathDistanceKM(first, point),
			solarEclipsePathDistanceKM(second, point),
		)
		if distance < bestDistance {
			bestIndex, bestDistance = index, distance
		}
	}
	return bestIndex, bestDistance
}

func orientSolarEclipsePath(
	points []SolarEclipsePathPoint,
	start, end SolarEclipsePathPoint,
) ([]SolarEclipsePathPoint, bool) {
	if len(points) < 2 {
		return nil, false
	}
	oriented := append([]SolarEclipsePathPoint(nil), points...)
	direct := solarEclipsePathDistanceKM(start, oriented[0]) +
		solarEclipsePathDistanceKM(end, oriented[len(oriented)-1])
	reverse := solarEclipsePathDistanceKM(start, oriented[len(oriented)-1]) +
		solarEclipsePathDistanceKM(end, oriented[0])
	if reverse < direct {
		for left, right := 0, len(oriented)-1; left < right; left, right = left+1, right-1 {
			oriented[left], oriented[right] = oriented[right], oriented[left]
		}
	}
	if solarEclipsePathDistanceKM(start, oriented[0]) > 0.1 ||
		solarEclipsePathDistanceKM(end, oriented[len(oriented)-1]) > 0.1 {
		return nil, false
	}
	oriented[0], oriented[len(oriented)-1] = start, end
	return oriented, true
}
