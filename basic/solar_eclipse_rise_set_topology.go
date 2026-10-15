package basic

import (
	"math"

	"github.com/starainrt/astro/internal/geodata"
)

func (solver solarEclipseSolver) completeRiseSetCurveEndpoints(
	curves []SolarEclipseRiseSetCurve,
	stepDays float64,
	phaseJunctions []solarEclipseRiseSetPhaseJunction,
) {
	for index := range curves {
		solver.completeRiseSetFoldEndpoints(&curves[index], stepDays)
	}
	solver.completeRiseSetPhaseJunctions(curves, stepDays, phaseJunctions)
	solver.completeRiseSetDirectionJunctions(curves, stepDays)
	for index := range curves {
		solver.refineRiseSetCurveSpacing(&curves[index])
		normalizeSolarEclipseRiseSetCurveSegments(&curves[index])
	}
}

// normalizeSolarEclipseRiseSetCurveSegments keeps each rendered branch
// strictly time-ordered. Endpoint completion can discover a real horizon fold
// after the sampled branch was built; that fold belongs to a separate branch,
// not to a reversed segment. Near-identical roots are numerical duplicates.
func normalizeSolarEclipseRiseSetCurveSegments(curve *SolarEclipseRiseSetCurve) {
	if curve == nil {
		return
	}
	segments := make([][]SolarEclipsePathPoint, 0, len(curve.Segments))
	for _, segment := range curve.Segments {
		segments = append(segments, splitSolarEclipseRiseSetTimeFolds(segment)...)
	}
	curve.Segments = segments
}

// splitSolarEclipseRiseSetTimeFolds preserves spatial branches when a
// horizon curve folds in time. A rendered segment must be strictly increasing
// in JDE, but the physical curve can turn around at a high-latitude horizon
// fold. Each monotonic branch is emitted separately; descending branches are
// reversed so their geometry is retained without violating the API contract.
func splitSolarEclipseRiseSetTimeFolds(segment []SolarEclipsePathPoint) [][]SolarEclipsePathPoint {
	if len(segment) < 2 {
		return nil
	}
	points := make([]SolarEclipsePathPoint, 0, len(segment))
	for _, point := range segment {
		if len(points) > 0 {
			last := points[len(points)-1]
			if math.Abs(point.JDE-last.JDE) <= solarEclipseRiseSetTimeEpsilonDays {
				if solarEclipsePathDistanceKM(point, last) <= 0.01 {
					continue
				}
				// Distinct points at the same instant are a junction, not a
				// valid edge of a timed segment. Keep both as separate runs.
				points = append(points, point)
				continue
			}
		}
		points = append(points, point)
	}
	if len(points) < 2 {
		return nil
	}
	result := make([][]SolarEclipsePathPoint, 0, 2)
	start := 0
	direction := 0
	flush := func(end int, branchDirection int) {
		if end-start < 1 {
			return
		}
		branch := append([]SolarEclipsePathPoint(nil), points[start:end+1]...)
		if branchDirection < 0 {
			for left, right := 0, len(branch)-1; left < right; left, right = left+1, right-1 {
				branch[left], branch[right] = branch[right], branch[left]
			}
		}
		if len(branch) >= 2 && branch[len(branch)-1].JDE > branch[0].JDE+solarEclipseRiseSetTimeEpsilonDays {
			result = append(result, branch)
		}
	}
	for index := 1; index < len(points); index++ {
		delta := points[index].JDE - points[index-1].JDE
		if math.Abs(delta) <= solarEclipseRiseSetTimeEpsilonDays {
			flush(index-1, direction)
			start = index
			direction = 0
			continue
		}
		sign := 1
		if delta < 0 {
			sign = -1
		}
		if direction == 0 {
			direction = sign
			continue
		}
		if sign != direction {
			// Keep the fold vertex in both adjacent branches. This is
			// necessary to retain the actual spatial turn after reversing
			// the descending branch.
			flush(index-1, direction)
			start = index - 1
			direction = sign
		}
	}
	flush(len(points)-1, direction)
	return result
}

func (solver solarEclipseSolver) completeRiseSetFoldEndpoints(
	curve *SolarEclipseRiseSetCurve,
	stepDays float64,
) {
	if curve == nil || len(curve.Segments) < 2 {
		return
	}
	type endpointRef struct {
		segmentIndex int
		atStart      bool
		point        SolarEclipsePathPoint
	}
	endpoints := make([]endpointRef, 0, 2*len(curve.Segments))
	for segmentIndex, segment := range curve.Segments {
		if len(segment) == 0 {
			continue
		}
		for _, atStart := range []bool{true, false} {
			endpoints = append(endpoints, endpointRef{
				segmentIndex: segmentIndex,
				atStart:      atStart,
				point:        solarEclipseRiseSetSegmentEndpoint(segment, atStart),
			})
		}
	}
	used := make(map[[2]int]bool, len(endpoints))
	for firstIndex := 0; firstIndex < len(endpoints); firstIndex++ {
		first := endpoints[firstIndex]
		if used[[2]int{first.segmentIndex, boolInt(first.atStart)}] {
			continue
		}
		for secondIndex := firstIndex + 1; secondIndex < len(endpoints); secondIndex++ {
			second := endpoints[secondIndex]
			if first.segmentIndex == second.segmentIndex || first.atStart != second.atStart ||
				used[[2]int{second.segmentIndex, boolInt(second.atStart)}] {
				continue
			}
			if math.Abs(first.point.JDE-second.point.JDE) > math.Max(1e-8, stepDays/4) {
				continue
			}
			distance := solarEclipsePathDistanceKM(first.point, second.point)
			if distance <= 0.01 || distance > 6000 {
				continue
			}
			fold, ok := solver.refineRiseSetFold(first.point, second.point, curve.Phase == RiseSetPhaseGreatest)
			if !ok || math.Abs(fold.JDE-(first.point.JDE+second.point.JDE)/2) > 2.5*stepDays ||
				solarEclipsePathDistanceKM(fold, first.point) > 6000 || solarEclipsePathDistanceKM(fold, second.point) > 6000 {
				continue
			}
			evaluation := solver.magnitudeEvaluationAt(fold.JDE)
			_, key, valid := evaluation.classify(fold.Longitude, fold.Latitude, curve.Phase == RiseSetPhaseGreatest)
			if !valid || key.phase != curve.Phase || key.direction != curve.Direction {
				continue
			}
			curve.Segments[first.segmentIndex] = solarEclipseRiseSetAddEndpoint(
				curve.Segments[first.segmentIndex], fold, first.atStart,
			)
			curve.Segments[second.segmentIndex] = solarEclipseRiseSetAddEndpoint(
				curve.Segments[second.segmentIndex], fold, second.atStart,
			)
			used[[2]int{first.segmentIndex, boolInt(first.atStart)}] = true
			used[[2]int{second.segmentIndex, boolInt(second.atStart)}] = true
			break
		}
	}
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func (solver solarEclipseSolver) completeRiseSetPhaseJunctions(
	curves []SolarEclipseRiseSetCurve,
	stepDays float64,
	phaseJunctions []solarEclipseRiseSetPhaseJunction,
) {
	curveIndices := make(map[solarEclipseRiseSetCurveKey]int, len(curves))
	for index, curve := range curves {
		curveIndices[solarEclipseRiseSetCurveKey{phase: curve.Phase, direction: curve.Direction}] = index
	}
	for _, direction := range []RiseSetDirection{RiseSetDirectionRise, RiseSetDirectionSet} {
		startIndex, haveStart := curveIndices[solarEclipseRiseSetCurveKey{phase: RiseSetPhaseStart, direction: direction}]
		greatestIndex, haveGreatest := curveIndices[solarEclipseRiseSetCurveKey{phase: RiseSetPhaseGreatest, direction: direction}]
		endIndex, haveEnd := curveIndices[solarEclipseRiseSetCurveKey{phase: RiseSetPhaseEnd, direction: direction}]
		if !haveStart || !haveGreatest || !haveEnd {
			continue
		}
		for _, candidate := range phaseJunctions {
			if candidate.direction != direction {
				continue
			}
			solver.attachRiseSetPhaseJunction(
				curves, startIndex, greatestIndex, endIndex,
				candidate.point, direction, stepDays,
			)
		}
		seeds := solarEclipseRiseSetUnsharedEndpoints(startIndex, curves[startIndex].Segments)
		for _, seed := range seeds {
			candidateSeeds := []SolarEclipsePathPoint{seed.point}
			if endEndpoint, endOK := solarEclipseClosestRiseSetEndpoint(
				seed.point, endIndex, curves[endIndex].Segments, nil, stepDays,
			); endOK {
				candidateSeeds = append([]SolarEclipsePathPoint{
					solarEclipseRiseSetMidpoint(seed.point, endEndpoint.point),
					endEndpoint.point,
				}, candidateSeeds...)
			}
			junction, ok := SolarEclipsePathPoint{}, false
			for _, candidate := range candidateSeeds {
				junction, ok = solver.refineRiseSetPhaseJunctionOnHorizon(candidate)
				if !ok {
					junction, ok = solver.refineRiseSetPhaseJunction(
						candidate.JDE, candidate.Longitude, candidate.Latitude,
					)
				}
				if ok {
					break
				}
			}
			if !ok || math.Abs(junction.JDE-seed.point.JDE) > 3*stepDays ||
				solarEclipsePathDistanceKM(junction, seed.point) > 3000 {
				continue
			}
			solver.attachRiseSetPhaseJunction(
				curves, startIndex, greatestIndex, endIndex,
				junction, direction, stepDays,
			)
		}
	}
}

func (solver solarEclipseSolver) attachRiseSetPhaseJunction(
	curves []SolarEclipseRiseSetCurve,
	startIndex, greatestIndex, endIndex int,
	junction SolarEclipsePathPoint,
	direction RiseSetDirection,
	stepDays float64,
) bool {
	matched := make([]solarEclipseRiseSetPhaseAttachment, 0, 3)
	for _, curveIndex := range []int{startIndex, greatestIndex, endIndex} {
		endpoint, ok := solarEclipseClosestRiseSetEndpoint(
			junction, curveIndex, curves[curveIndex].Segments, nil, stepDays,
		)
		if !ok {
			endpoint, ok = solarEclipseClosestRiseSetFoldEndpoint(
				junction, curveIndex, curves[curveIndex].Segments, stepDays,
			)
		}
		if !ok {
			return false
		}
		phase := curves[curveIndex].Phase
		attachment := solarEclipseRiseSetPhaseAttachment{endpoint: endpoint}
		if !solarEclipseRiseSetEndpointTimeDirectionValid(junction, endpoint) ||
			(solarEclipsePathDistanceKM(junction, endpoint.point) > solarEclipseRiseSetPhaseConnectionLimitKM &&
				!solver.riseSetPhaseSegmentIsContinuous(junction, endpoint.point, phase, direction)) {
			greatest := phase == RiseSetPhaseGreatest
			fold, foldOK := solver.refineRiseSetFold(junction, endpoint.point, greatest)
			if !foldOK || !solver.riseSetFoldBridgesPhaseJunction(
				junction, endpoint, fold, phase, direction,
			) {
				fold, foldOK = solver.refineRiseSetFoldNearPhaseJunction(
					junction, endpoint, phase, direction, stepDays,
				)
			}
			if !foldOK || !solver.riseSetFoldBridgesPhaseJunction(
				junction, endpoint, fold, phase, direction,
			) {
				return false
			}
			attachment.fold, attachment.hasFold = fold, true
		}
		matched = append(matched, attachment)
	}
	for _, attachment := range matched {
		endpoint := attachment.endpoint
		if attachment.hasFold {
			curves[endpoint.curveIndex].Segments[endpoint.segmentIndex] = solarEclipseRiseSetAddEndpoint(
				curves[endpoint.curveIndex].Segments[endpoint.segmentIndex], attachment.fold, endpoint.atStart,
			)
			shortBranch := []SolarEclipsePathPoint{attachment.fold, junction}
			if !endpoint.atStart {
				shortBranch[0], shortBranch[1] = shortBranch[1], shortBranch[0]
			}
			curves[endpoint.curveIndex].Segments = append(curves[endpoint.curveIndex].Segments, shortBranch)
			continue
		}
		curves[endpoint.curveIndex].Segments[endpoint.segmentIndex] = solarEclipseRiseSetAddEndpoint(
			curves[endpoint.curveIndex].Segments[endpoint.segmentIndex], junction, endpoint.atStart,
		)
	}
	return true
}

type solarEclipseRiseSetEndpointRef struct {
	curveIndex   int
	segmentIndex int
	atStart      bool
	point        SolarEclipsePathPoint
}

type solarEclipseRiseSetPhaseAttachment struct {
	endpoint solarEclipseRiseSetEndpointRef
	fold     SolarEclipsePathPoint
	hasFold  bool
}

func solarEclipseRiseSetMidpoint(first, second SolarEclipsePathPoint) SolarEclipsePathPoint {
	return SolarEclipsePathPoint{
		JDE: (first.JDE + second.JDE) / 2,
		Longitude: geodata.NormalizeLongitude(
			first.Longitude + math.Remainder(second.Longitude-first.Longitude, 360)/2,
		),
		Latitude:    (first.Latitude + second.Latitude) / 2,
		SunAltitude: (first.SunAltitude + second.SunAltitude) / 2,
	}
}

func solarEclipseRiseSetUnsharedEndpoints(
	curveIndex int,
	segments [][]SolarEclipsePathPoint,
) []solarEclipseRiseSetEndpointRef {
	endpoints := make([]solarEclipseRiseSetEndpointRef, 0, 2*len(segments))
	for segmentIndex, segment := range segments {
		if len(segment) == 0 {
			continue
		}
		for _, atStart := range []bool{true, false} {
			endpoints = append(endpoints, solarEclipseRiseSetEndpointRef{
				curveIndex: curveIndex, segmentIndex: segmentIndex, atStart: atStart,
				point: solarEclipseRiseSetSegmentEndpoint(segment, atStart),
			})
		}
	}
	result := make([]solarEclipseRiseSetEndpointRef, 0, len(endpoints))
	for index, endpoint := range endpoints {
		shared := false
		for otherIndex, other := range endpoints {
			if index == otherIndex || endpoint.segmentIndex == other.segmentIndex {
				continue
			}
			if math.Abs(endpoint.point.JDE-other.point.JDE) <= 1e-8 &&
				solarEclipsePathDistanceKM(endpoint.point, other.point) <= 0.01 {
				shared = true
				break
			}
		}
		if !shared {
			result = append(result, endpoint)
		}
	}
	return result
}

func solarEclipseClosestRiseSetEndpoint(
	junction SolarEclipsePathPoint,
	curveIndex int,
	segments [][]SolarEclipsePathPoint,
	used map[[3]int]bool,
	stepDays float64,
) (solarEclipseRiseSetEndpointRef, bool) {
	best := solarEclipseRiseSetEndpointRef{}
	bestMetric := math.Inf(1)
	for segmentIndex, segment := range segments {
		if len(segment) == 0 {
			continue
		}
		for _, atStart := range []bool{true, false} {
			side := 1
			if atStart {
				side = 0
			}
			if used[[3]int{curveIndex, segmentIndex, side}] {
				continue
			}
			point := solarEclipseRiseSetSegmentEndpoint(segment, atStart)
			if atStart && junction.JDE > point.JDE+solarEclipseRiseSetAttachmentTimeToleranceDays ||
				!atStart && junction.JDE < point.JDE-solarEclipseRiseSetAttachmentTimeToleranceDays ||
				math.Abs(junction.JDE-point.JDE) > 3*stepDays {
				continue
			}
			metric := solarEclipsePathDistanceKM(junction, point) + math.Abs(junction.JDE-point.JDE)*8640
			if metric < bestMetric {
				best = solarEclipseRiseSetEndpointRef{
					curveIndex: curveIndex, segmentIndex: segmentIndex, atStart: atStart, point: point,
				}
				bestMetric = metric
			}
		}
	}
	return best, bestMetric < math.Inf(1)
}

func solarEclipseClosestRiseSetFoldEndpoint(
	junction SolarEclipsePathPoint,
	curveIndex int,
	segments [][]SolarEclipsePathPoint,
	stepDays float64,
) (solarEclipseRiseSetEndpointRef, bool) {
	best := solarEclipseRiseSetEndpointRef{}
	bestMetric := math.Inf(1)
	for segmentIndex, segment := range segments {
		if len(segment) == 0 {
			continue
		}
		for _, atStart := range []bool{true, false} {
			point := solarEclipseRiseSetSegmentEndpoint(segment, atStart)
			deltaDays := math.Abs(junction.JDE - point.JDE)
			distance := solarEclipsePathDistanceKM(junction, point)
			if deltaDays > 3*stepDays || distance > 6000 {
				continue
			}
			metric := distance + deltaDays*8640
			if metric < bestMetric {
				best = solarEclipseRiseSetEndpointRef{
					curveIndex: curveIndex, segmentIndex: segmentIndex, atStart: atStart, point: point,
				}
				bestMetric = metric
			}
		}
	}
	return best, bestMetric < math.Inf(1)
}

func (solver solarEclipseSolver) completeRiseSetDirectionJunctions(
	curves []SolarEclipseRiseSetCurve,
	stepDays float64,
) {
	curveIndices := make(map[solarEclipseRiseSetCurveKey]int, len(curves))
	for index, curve := range curves {
		curveIndices[solarEclipseRiseSetCurveKey{phase: curve.Phase, direction: curve.Direction}] = index
	}
	for _, phase := range []RiseSetPhase{RiseSetPhaseStart, RiseSetPhaseGreatest, RiseSetPhaseEnd} {
		riseIndex, haveRise := curveIndices[solarEclipseRiseSetCurveKey{phase: phase, direction: RiseSetDirectionRise}]
		setIndex, haveSet := curveIndices[solarEclipseRiseSetCurveKey{phase: phase, direction: RiseSetDirectionSet}]
		if !haveRise || !haveSet {
			continue
		}
		riseEndpoints := solarEclipseRiseSetAllEndpoints(riseIndex, curves[riseIndex].Segments)
		setEndpoints := solarEclipseRiseSetAllEndpoints(setIndex, curves[setIndex].Segments)
		usedRise := make([]bool, len(riseEndpoints))
		usedSet := make([]bool, len(setEndpoints))
		for riseEndpointIndex, riseEndpoint := range riseEndpoints {
			bestSetIndex := -1
			bestMetric := math.Inf(1)
			for setEndpointIndex, setEndpoint := range setEndpoints {
				if usedSet[setEndpointIndex] || math.Abs(riseEndpoint.point.JDE-setEndpoint.point.JDE) > 3*stepDays {
					continue
				}
				distance := solarEclipsePathDistanceKM(riseEndpoint.point, setEndpoint.point)
				if distance > 2500 {
					continue
				}
				metric := distance + math.Abs(riseEndpoint.point.JDE-setEndpoint.point.JDE)*8640
				if metric < bestMetric {
					bestSetIndex, bestMetric = setEndpointIndex, metric
				}
			}
			if bestSetIndex < 0 {
				continue
			}
			setEndpoint := setEndpoints[bestSetIndex]
			longitude := geodata.NormalizeLongitude(
				riseEndpoint.point.Longitude + math.Remainder(setEndpoint.point.Longitude-riseEndpoint.point.Longitude, 360)/2,
			)
			latitude := (riseEndpoint.point.Latitude + setEndpoint.point.Latitude) / 2
			junction, ok := solver.refineRiseSetDirectionJunction(
				(riseEndpoint.point.JDE+setEndpoint.point.JDE)/2, longitude, latitude, phase == RiseSetPhaseGreatest,
			)
			if !ok || math.Abs(junction.JDE-riseEndpoint.point.JDE) > 3*stepDays ||
				math.Abs(junction.JDE-setEndpoint.point.JDE) > 3*stepDays ||
				solarEclipsePathDistanceKM(junction, riseEndpoint.point) > 2500 ||
				solarEclipsePathDistanceKM(junction, setEndpoint.point) > 2500 {
				continue
			}
			if !solarEclipseRiseSetEndpointTimeDirectionValid(junction, riseEndpoint) ||
				!solarEclipseRiseSetEndpointTimeDirectionValid(junction, setEndpoint) {
				continue
			}
			curves[riseIndex].Segments[riseEndpoint.segmentIndex] = solarEclipseRiseSetAddEndpoint(
				curves[riseIndex].Segments[riseEndpoint.segmentIndex], junction, riseEndpoint.atStart,
			)
			curves[setIndex].Segments[setEndpoint.segmentIndex] = solarEclipseRiseSetAddEndpoint(
				curves[setIndex].Segments[setEndpoint.segmentIndex], junction, setEndpoint.atStart,
			)
			usedRise[riseEndpointIndex] = true
			usedSet[bestSetIndex] = true
		}
	}
}

func solarEclipseRiseSetAllEndpoints(
	curveIndex int,
	segments [][]SolarEclipsePathPoint,
) []solarEclipseRiseSetEndpointRef {
	endpoints := make([]solarEclipseRiseSetEndpointRef, 0, 2*len(segments))
	for segmentIndex, segment := range segments {
		if len(segment) == 0 {
			continue
		}
		for _, atStart := range []bool{true, false} {
			endpoints = append(endpoints, solarEclipseRiseSetEndpointRef{
				curveIndex: curveIndex, segmentIndex: segmentIndex, atStart: atStart,
				point: solarEclipseRiseSetSegmentEndpoint(segment, atStart),
			})
		}
	}
	return endpoints
}

func solarEclipseRiseSetEndpointTimeDirectionValid(
	junction SolarEclipsePathPoint,
	endpoint solarEclipseRiseSetEndpointRef,
) bool {
	if endpoint.atStart {
		return junction.JDE <= endpoint.point.JDE+solarEclipseRiseSetAttachmentTimeToleranceDays
	}
	return junction.JDE >= endpoint.point.JDE-solarEclipseRiseSetAttachmentTimeToleranceDays
}

func (solver solarEclipseSolver) refineRiseSetDirectionJunction(
	jde, longitude, latitude float64,
	greatest bool,
) (SolarEclipsePathPoint, bool) {
	const (
		geographicStep = 1e-4
		timeStep       = 1.0 / 86400.0
	)
	for iteration := 0; iteration < 24; iteration++ {
		evaluation := solver.magnitudeEvaluationAt(jde)
		residual, ok := solarEclipseRiseSetDirectionJunctionResidualAt(evaluation, longitude, latitude, greatest)
		if !ok {
			return SolarEclipsePathPoint{}, false
		}
		if math.Abs(residual[0]) <= 1e-10 && math.Abs(residual[1]) <= 1e-10 && math.Abs(residual[2]) <= 1e-10 {
			break
		}
		longitudeResidual, longitudeOK := solarEclipseRiseSetDirectionJunctionResidualAt(
			evaluation, longitude+geographicStep, latitude, greatest,
		)
		latitudeResidual, latitudeOK := solarEclipseRiseSetDirectionJunctionResidualAt(
			evaluation, longitude, latitude+geographicStep, greatest,
		)
		timeResidual, timeOK := solver.riseSetDirectionJunctionResidual(
			jde+timeStep, longitude, latitude, greatest,
		)
		if !longitudeOK || !latitudeOK || !timeOK {
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
		if math.Abs(delta[2]) > 5.0/1440.0 {
			delta[2] = math.Copysign(5.0/1440.0, delta[2])
		}
		longitude = geodata.NormalizeLongitude(longitude + delta[0])
		latitude += delta[1]
		jde += delta[2]
		if latitude <= -89.999999 || latitude >= 89.999999 {
			return SolarEclipsePathPoint{}, false
		}
	}
	residual, ok := solver.riseSetDirectionJunctionResidual(jde, longitude, latitude, greatest)
	if !ok || math.Abs(residual[0]) > 1e-7 || math.Abs(residual[1]) > 1e-8 || math.Abs(residual[2]) > 1e-7 {
		return SolarEclipsePathPoint{}, false
	}
	evaluation := solver.magnitudeEvaluationAt(jde)
	state := evaluation.center.stateAt(longitude*rad, latitude*rad, 0)
	if greatest {
		if solarEclipsePartialContactGap(state) > 1e-7 || evaluation.separationSecondDerivative(longitude, latitude) <= 0 {
			return SolarEclipsePathPoint{}, false
		}
	} else if math.Abs(evaluation.partialContactDerivative(longitude, latitude)) < 1e-10 {
		return SolarEclipsePathPoint{}, false
	}
	return SolarEclipsePathPoint{
		JDE: jde, Longitude: longitude, Latitude: latitude, SunAltitude: residual[1] / rad,
	}, true
}

func (solver solarEclipseSolver) riseSetDirectionJunctionResidual(
	jde, longitude, latitude float64,
	greatest bool,
) ([3]float64, bool) {
	evaluation := solver.magnitudeEvaluationAt(jde)
	return solarEclipseRiseSetDirectionJunctionResidualAt(evaluation, longitude, latitude, greatest)
}

func solarEclipseRiseSetDirectionJunctionResidualAt(
	evaluation solarEclipseRiseSetEvaluation,
	longitude, latitude float64,
	greatest bool,
) ([3]float64, bool) {
	phaseResidual, phaseOK := solarEclipseRiseSetPhaseResidual(evaluation, longitude, latitude, greatest)
	state := evaluation.center.stateAt(longitude*rad, latitude*rad, 0)
	altitudeDerivative := evaluation.sunAltitudeDerivative(longitude, latitude)
	residual := [3]float64{phaseResidual, state.sunAltitudeRad, altitudeDerivative}
	return residual, phaseOK && finite(residual[1]) && finite(residual[2])
}

func solarEclipseRiseSetSegmentEndpoint(segment []SolarEclipsePathPoint, atStart bool) SolarEclipsePathPoint {
	if atStart {
		return segment[0]
	}
	return segment[len(segment)-1]
}

func solarEclipseRiseSetAddEndpoint(
	segment []SolarEclipsePathPoint,
	point SolarEclipsePathPoint,
	atStart bool,
) []SolarEclipsePathPoint {
	current := solarEclipseRiseSetSegmentEndpoint(segment, atStart)
	if math.Abs(current.JDE-point.JDE) <= solarEclipseRiseSetAttachmentTimeToleranceDays &&
		solarEclipsePathDistanceKM(current, point) <= solarEclipseRiseSetAttachmentDistanceToleranceKM {
		segmentIndex := len(segment) - 1
		if atStart {
			segmentIndex = 0
		}
		segment[segmentIndex] = point
		return segment
	}
	if atStart {
		result := make([]SolarEclipsePathPoint, 0, len(segment)+1)
		result = append(result, point)
		return append(result, segment...)
	}
	return append(segment, point)
}
