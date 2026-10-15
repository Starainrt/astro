package basic

import (
	"math"

	"github.com/starainrt/astro/internal/geodata"
)

// Absolute Julian dates quantize a moving sky position at about 1e-10 radians.
const solarEclipseCentralVectorTolerance = 2e-10

// solarCentralBandSkyOffset uses a signed internal-contact radius and stable
// sky-plane coordinates. Unlike acos(dot) - abs(radius), these remain smooth
// when a hybrid shadow shrinks to zero and changes from annular to total.
func solarCentralBandSkyOffset(context localSolarEclipseStateContext, longitude, latitude float64) [3]float64 {
	observer := localSolarEclipseObserverXYZ(context.gst, longitude*rad, latitude*rad, 0)
	sun, moon := subtractSolarEclipse3(context.sunXYZ, observer), subtractSolarEclipse3(context.moonXYZ, observer)
	sunDistance := math.Sqrt(dotSolarEclipse3(sun, sun))
	moonDistance := math.Sqrt(dotSolarEclipse3(moon, moon))
	for i := range sun {
		sun[i] /= sunDistance
		moon[i] /= moonDistance
	}
	equatorial := math.Hypot(sun[0], sun[1])
	east := [3]float64{-sun[1] / equatorial, sun[0] / equatorial, 0}
	north := [3]float64{-sun[2] * east[1], sun[2] * east[0], equatorial}
	radius := math.Asin(geodata.EarthEquatorialRadiusKM*context.params.umbralK*localSolarMoonRadiusScale/moonDistance) -
		math.Asin(geodata.EarthEquatorialRadiusKM*context.params.sunRadiusRatio/sunDistance)
	return [3]float64{dotSolarEclipse3(moon, east), dotSolarEclipse3(moon, north), math.Sin(radius)}
}

func solarCentralBandVectorResidual(evaluation solarEclipseRiseSetEvaluation, longitude, latitude, side float64) ([2]float64, bool) {
	center := solarCentralBandSkyOffset(evaluation.center, longitude, latitude)
	before := solarCentralBandSkyOffset(evaluation.before, longitude, latitude)
	after := solarCentralBandSkyOffset(evaluation.after, longitude, latitude)
	velocity := subtractSolarEclipse3(after, before)
	v2 := velocity[0]*velocity[0] + velocity[1]*velocity[1]
	discriminant := v2 - velocity[2]*velocity[2]
	if v2 <= 0 || discriminant <= 0 {
		return [2]float64{}, false
	}
	// At contact, offset = signedRadius * normal. The envelope condition is
	// normal dot velocity = radiusVelocity, giving two regular signed branches.
	cross := side * math.Sqrt(discriminant)
	nx := (velocity[2]*velocity[0] - cross*velocity[1]) / v2
	ny := (velocity[2]*velocity[1] + cross*velocity[0]) / v2
	residual := [2]float64{center[0] - center[2]*nx, center[1] - center[2]*ny}
	return residual, finite(residual[0]) && finite(residual[1])
}

func (solver solarEclipseSolver) centralBandVectorJacobian(coordinates [3]float64, referenceJDE, side float64, exact bool) ([2]float64, [2][3]float64, bool) {
	evaluate := solver.magnitudeCandidateEvaluationAt
	if exact {
		evaluate = solver.magnitudeEvaluationAt
	}
	jd := referenceJDE + coordinates[2]/solarEclipseNonCentralBandTimeScale
	evaluation := evaluate(jd)
	residual, ok := solarCentralBandVectorResidual(evaluation, coordinates[0], coordinates[1], side)
	if !ok {
		return residual, [2][3]float64{}, false
	}
	steps := [3]float64{1e-4, 1e-4, 5 * solarEclipseNonCentralBandTimeScale / 86400}
	var jacobian [2][3]float64
	for column := 0; column < 3; column++ {
		shifted := coordinates
		shifted[column] += steps[column]
		shiftedEvaluation := evaluation
		if column == 2 {
			shiftedEvaluation = evaluate(jd + steps[column]/solarEclipseNonCentralBandTimeScale)
		}
		value, valid := solarCentralBandVectorResidual(shiftedEvaluation, shifted[0], shifted[1], side)
		if !valid {
			return residual, jacobian, false
		}
		for row := range residual {
			jacobian[row][column] = (value[row] - residual[row]) / steps[column]
		}
	}
	return residual, jacobian, true
}

func (solver solarEclipseSolver) correctCentralBandVectorBoundary(predictor, tangent [3]float64, referenceJDE, side float64) (solarEclipseNonCentralBandState, bool) {
	coordinates := predictor
	exact := false
	for iteration := 0; iteration < 16; iteration++ {
		residual, jacobian, ok := solver.centralBandVectorJacobian(coordinates, referenceJDE, side, exact)
		if !ok {
			return solarEclipseNonCentralBandState{}, false
		}
		plane := dotSolarEclipse3(subtractSolarEclipse3(coordinates, predictor), tangent)
		if math.Hypot(residual[0], residual[1]) <= solarEclipseCentralVectorTolerance && math.Abs(plane) <= 1e-9 {
			jde := referenceJDE + coordinates[2]/solarEclipseNonCentralBandTimeScale
			evaluation := solver.magnitudeEvaluationAt(jde)
			check, valid := solarCentralBandVectorResidual(evaluation, coordinates[0], coordinates[1], side)
			if !valid || math.Hypot(check[0], check[1]) > solarEclipseCentralVectorTolerance {
				exact = true
				continue
			}
			nextTangent, valid := solarEclipseMagnitudeArcTangent(jacobian)
			state := evaluation.center.stateAt(coordinates[0]*rad, coordinates[1]*rad, 0)
			return solarEclipseNonCentralBandState{
				coordinates: coordinates, tangent: nextTangent,
				point: SolarEclipsePathPoint{JDE: jde, Longitude: geodata.NormalizeLongitude(coordinates[0]), Latitude: coordinates[1], SunAltitude: state.sunAltitudeRad / rad},
			}, valid
		}
		delta, valid := solveSolarEclipse3x3([3][3]float64{jacobian[0], jacobian[1], tangent}, [3]float64{-residual[0], -residual[1], -plane})
		if !valid {
			return solarEclipseNonCentralBandState{}, false
		}
		scale := math.Max(1, math.Sqrt(dotSolarEclipse3(delta, delta))/2)
		for i := range coordinates {
			coordinates[i] += delta[i] / scale
		}
		if math.Abs(coordinates[1]) >= 89.999999 {
			return solarEclipseNonCentralBandState{}, false
		}
	}
	return solarEclipseNonCentralBandState{}, false
}

func (solver solarEclipseSolver) traceCentralBandVectorEnvelope(root SolarEclipsePathPoint, endRoots []SolarEclipsePathPoint, transitions *[]SolarEclipsePathPoint, referenceJDE float64) ([]SolarEclipsePathPoint, int, bool) {
	coordinates := [3]float64{root.Longitude, root.Latitude, (root.JDE - referenceJDE) * solarEclipseNonCentralBandTimeScale}
	evaluation := solver.magnitudeEvaluationAt(root.JDE)
	side := 1.0
	positive, _ := solarCentralBandVectorResidual(evaluation, root.Longitude, root.Latitude, 1)
	negative, _ := solarCentralBandVectorResidual(evaluation, root.Longitude, root.Latitude, -1)
	if math.Hypot(negative[0], negative[1]) < math.Hypot(positive[0], positive[1]) {
		side = -1
	}
	_, jacobian, ok := solver.centralBandVectorJacobian(coordinates, referenceJDE, side, false)
	if !ok {
		return nil, -1, false
	}
	tangent, ok := solarEclipseMagnitudeArcTangent(jacobian)
	if !ok {
		return nil, -1, false
	}
	step := solarEclipseCentralEnvelopeArcStepDegrees / 4
	state := solarEclipseNonCentralBandState{coordinates: coordinates, tangent: tangent, point: root}
	var next solarEclipseNonCentralBandState
	found := false
	for _, direction := range []float64{1, -1} {
		predictor, oriented := coordinates, tangent
		for i := range predictor {
			oriented[i] *= direction
			predictor[i] += step * oriented[i]
		}
		candidate, valid := solver.correctCentralBandVectorBoundary(predictor, oriented, referenceJDE, side)
		if valid && candidate.point.SunAltitude > 0 && (!found || candidate.point.SunAltitude > next.point.SunAltitude) {
			if dotSolarEclipse3(candidate.tangent, oriented) < 0 {
				for i := range candidate.tangent {
					candidate.tangent[i] = -candidate.tangent[i]
				}
			}
			next, found = candidate, true
		}
	}
	if !found {
		return nil, -1, false
	}
	points := []SolarEclipsePathPoint{root}
	transitionIndex := 0
	// A hybrid transition can be less than a second from axis contact, where
	// time also folds along a limit. Locate it by signed radius while tracing.
	appendPoint := func(point SolarEclipsePathPoint) bool {
		first := points[len(points)-1]
		before := solarCentralBandSkyOffset(solver.localStateContextAt(first.JDE), first.Longitude, first.Latitude)
		after := solarCentralBandSkyOffset(solver.localStateContextAt(point.JDE), point.Longitude, point.Latitude)
		if before[2]*after[2] < 0 {
			if transitionIndex == len(*transitions) {
				fraction := before[2] / (before[2] - after[2])
				seed := SolarEclipsePathPoint{JDE: first.JDE + fraction*(point.JDE-first.JDE),
					Longitude: first.Longitude + fraction*math.Remainder(point.Longitude-first.Longitude, 360),
					Latitude:  first.Latitude + fraction*(point.Latitude-first.Latitude)}
				transition, ok := solver.hybridCentralBandTransition(seed, referenceJDE)
				if !ok {
					return false
				}
				*transitions = append(*transitions, transition)
			}
			points = append(points, (*transitions)[transitionIndex])
			transitionIndex++
		}
		points = append(points, point)
		return true
	}
	if !appendPoint(next.point) {
		return nil, -1, false
	}
	state = next
	for count := 0; count < solarEclipseCentralEnvelopeMaxArcSteps; count++ {
		predictor := state.coordinates
		for i := range predictor {
			predictor[i] += step * state.tangent[i]
		}
		candidate, valid := solver.correctCentralBandVectorBoundary(predictor, state.tangent, referenceJDE, side)
		if valid && dotSolarEclipse3(candidate.tangent, state.tangent) < 0 {
			for i := range candidate.tangent {
				candidate.tangent[i] = -candidate.tangent[i]
			}
		}
		distance := solarEclipsePathDistanceKM(state.point, candidate.point)
		chordTolerance := 0.02
		if valid {
			context := solver.localStateContextAt(candidate.point.JDE)
			offset := solarCentralBandSkyOffset(context, candidate.point.Longitude, candidate.point.Latitude)
			shadowRadiusKM := math.Abs(offset[2]) * math.Sqrt(dotSolarEclipse3(context.moonXYZ, context.moonXYZ))
			chordTolerance = math.Min(chordTolerance, math.Max(0.001, shadowRadiusKM/8))
		}
		if !valid || distance > solarEclipseCentralEnvelopeMaxSpacingKM ||
			centralBandVectorChordErrorKM(state, candidate, distance) > chordTolerance {
			step /= 2
			if step < solarEclipseCentralEnvelopeMinArcStepDegrees {
				return points, -1, false
			}
			continue
		}
		if candidate.point.SunAltitude < 0 {
			endIndex, bestResidual := -1, math.Inf(1)
			for i, end := range endRoots {
				residual, ok := solarCentralBandVectorResidual(solver.magnitudeEvaluationAt(end.JDE), end.Longitude, end.Latitude, side)
				if norm := math.Hypot(residual[0], residual[1]); ok && norm < bestResidual {
					endIndex, bestResidual = i, norm
				}
			}
			if endIndex < 0 || solarEclipsePathDistanceKM(state.point, endRoots[endIndex]) > solarEclipseCentralEnvelopeEndDistanceKM {
				return points, -1, false
			}
			ok := appendPoint(endRoots[endIndex])
			return points, endIndex, ok && transitionIndex == len(*transitions)
		}
		if !appendPoint(candidate.point) {
			return points, -1, false
		}
		state = candidate
		if distance < solarEclipseCentralEnvelopeMaxSpacingKM/2 {
			step = math.Min(solarEclipseCentralEnvelopeArcStepDegrees, step*1.5)
		}
	}
	return points, -1, false
}

func centralBandVectorChordErrorKM(first, second solarEclipseNonCentralBandState, distance float64) float64 {
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

func (solver solarEclipseSolver) hybridCentralBandTransition(seed SolarEclipsePathPoint, referenceJDE float64) (SolarEclipsePathPoint, bool) {
	coordinates := [3]float64{seed.Longitude, seed.Latitude, (seed.JDE - referenceJDE) * solarEclipseNonCentralBandTimeScale}
	steps := [3]float64{1e-4, 1e-4, solarEclipseNonCentralBandTimeScale / 86400}
	for iteration := 0; iteration < 12; iteration++ {
		jde := referenceJDE + coordinates[2]/solarEclipseNonCentralBandTimeScale
		context := solver.localStateContextAt(jde)
		residual := solarCentralBandSkyOffset(context, coordinates[0], coordinates[1])
		if math.Hypot(residual[0], residual[1]) < solarEclipseCentralVectorTolerance/2 && math.Abs(residual[2]) < 1e-12 {
			state := context.stateAt(coordinates[0]*rad, coordinates[1]*rad, 0)
			return SolarEclipsePathPoint{JDE: jde, Longitude: geodata.NormalizeLongitude(coordinates[0]), Latitude: coordinates[1], SunAltitude: state.sunAltitudeRad / rad}, true
		}
		var jacobian [3][3]float64
		for column := range coordinates {
			shifted, shiftedContext := coordinates, context
			shifted[column] += steps[column]
			if column == 2 {
				shiftedContext = solver.localStateContextAt(jde + steps[column]/solarEclipseNonCentralBandTimeScale)
			}
			value := solarCentralBandSkyOffset(shiftedContext, shifted[0], shifted[1])
			for row := range residual {
				jacobian[row][column] = (value[row] - residual[row]) / steps[column]
			}
		}
		delta, valid := solveSolarEclipse3x3(jacobian, [3]float64{-residual[0], -residual[1], -residual[2]})
		if !valid {
			return SolarEclipsePathPoint{}, false
		}
		for i := range coordinates {
			coordinates[i] += delta[i]
		}
	}
	return SolarEclipsePathPoint{}, false
}

func (solver solarEclipseSolver) centralBandVectorHorizonRoots(axisContactJDE, direction, firstContactJDE, lastContactJDE float64) (SolarEclipsePathPoint, SolarEclipsePathPoint, bool) {
	startJDE, endJDE := math.Min(firstContactJDE, lastContactJDE), math.Max(firstContactJDE, lastContactJDE)
	if !finite(startJDE) || !finite(endJDE) || startJDE <= 0 || endJDE <= startJDE {
		return SolarEclipsePathPoint{}, SolarEclipsePathPoint{}, false
	}
	seed, ok := solver.centralPathPointAt(axisContactJDE + direction/86400)
	if !ok {
		return SolarEclipsePathPoint{}, SolarEclipsePathPoint{}, false
	}
	var roots [2]SolarEclipsePathPoint
	for i, side := range []float64{1, -1} {
		coordinates := [3]float64{seed.Longitude, seed.Latitude, (seed.JDE - axisContactJDE) * solarEclipseNonCentralBandTimeScale}
		steps := [3]float64{1e-4, 1e-4, 5 * solarEclipseNonCentralBandTimeScale / 86400}
		found := false
		for iteration := 0; iteration < 12; iteration++ {
			residual, jacobian, valid := solver.centralBandVectorJacobian(coordinates, axisContactJDE, side, true)
			if !valid {
				break
			}
			jde := axisContactJDE + coordinates[2]/solarEclipseNonCentralBandTimeScale
			context := solver.localStateContextAt(jde)
			state := context.stateAt(coordinates[0]*rad, coordinates[1]*rad, 0)
			if math.Hypot(residual[0], residual[1]) <= solarEclipseCentralVectorTolerance && math.Abs(state.sunAltitudeRad) < 1e-8 {
				roots[i] = SolarEclipsePathPoint{JDE: jde, Longitude: geodata.NormalizeLongitude(coordinates[0]), Latitude: coordinates[1], SunAltitude: state.sunAltitudeRad / rad}
				// Grazing horizon roots can be many minutes from axis contact.
				// Bound them by the shadow's limb-crossing interval, not a fixed
				// window around the seed.
				found = jde >= startJDE-solarEclipseCentralLimitHorizonContactMarginDays &&
					jde <= endJDE+solarEclipseCentralLimitHorizonContactMarginDays && math.Abs(coordinates[1]) <= 90
				break
			}
			matrix := [3][3]float64{jacobian[0], jacobian[1], {}}
			for column := range coordinates {
				shifted, shiftedContext := coordinates, context
				shifted[column] += steps[column]
				if column == 2 {
					shiftedContext = solver.localStateContextAt(jde + steps[column]/solarEclipseNonCentralBandTimeScale)
				}
				value := shiftedContext.stateAt(shifted[0]*rad, shifted[1]*rad, 0)
				matrix[2][column] = (value.sunAltitudeRad - state.sunAltitudeRad) / steps[column]
			}
			delta, valid := solveSolarEclipse3x3(matrix, [3]float64{-residual[0], -residual[1], -state.sunAltitudeRad})
			if !valid {
				break
			}
			for j := range coordinates {
				coordinates[j] += delta[j]
			}
		}
		if !found {
			return SolarEclipsePathPoint{}, SolarEclipsePathPoint{}, false
		}
	}
	if roots[1].JDE < roots[0].JDE {
		roots[0], roots[1] = roots[1], roots[0]
	}
	return roots[0], roots[1], solarEclipsePathDistanceKM(roots[0], roots[1]) > 0.001
}

func (solver solarEclipseSolver) hybridCentralBandEnvelope(closures [][]SolarEclipsePathPoint, result SolarEclipseResult) [][]SolarEclipsePathPoint {
	if len(closures) != 2 || len(closures[0]) < 2 || len(closures[1]) < 2 {
		return nil
	}
	var transitions []SolarEclipsePathPoint
	startRoots := []SolarEclipsePathPoint{closures[0][0], closures[0][len(closures[0])-1]}
	endRoots := []SolarEclipsePathPoint{closures[1][0], closures[1][len(closures[1])-1]}
	var branches [2][]SolarEclipsePathPoint
	var ends [2]int
	for i, root := range startRoots {
		branch, end, ok := solver.traceCentralBandVectorEnvelope(root, endRoots, &transitions, result.GreatestEclipse)
		if !ok {
			return nil
		}
		branches[i], ends[i] = branch, end
	}
	if ends[0] == ends[1] {
		return nil
	}
	indices := [2]int{}
	polygons := make([][]SolarEclipsePathPoint, 0, len(transitions)+1)
	for segment := 0; segment <= len(transitions); segment++ {
		var parts [2][]SolarEclipsePathPoint
		for side, branch := range branches {
			last := len(branch) - 1
			if segment < len(transitions) {
				last = indices[side]
				for last < len(branch) && branch[last] != transitions[segment] {
					last++
				}
				if last == len(branch) {
					return nil
				}
			}
			parts[side] = branch[indices[side] : last+1]
			indices[side] = last
		}
		ring := append([]SolarEclipsePathPoint(nil), parts[0]...)
		if segment == len(transitions) {
			closure, ok := orientSolarEclipsePath(closures[1], parts[0][len(parts[0])-1], parts[1][len(parts[1])-1])
			if !ok {
				return nil
			}
			ring = append(ring, closure[1:]...)
		}
		for i := len(parts[1]) - 2; i >= 0; i-- {
			ring = append(ring, parts[1][i])
		}
		if segment == 0 {
			closure, ok := orientSolarEclipsePath(closures[0], parts[1][0], parts[0][0])
			if !ok {
				return nil
			}
			ring = append(ring, closure[1:]...)
		} else {
			ring = append(ring, ring[0])
		}
		polygons = append(polygons, deduplicateSolarEclipsePathPoints(ring))
	}
	return polygons
}
