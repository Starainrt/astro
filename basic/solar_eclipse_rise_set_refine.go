package basic

import "math"

func (solver solarEclipseSolver) refineRiseSetPhaseJunction(
	jd, longitude, latitude float64,
) (SolarEclipsePathPoint, bool) {
	const (
		geographicStep = 1e-4
		timeStep       = 1.0 / 86400.0
	)
	for iteration := 0; iteration < 24; iteration++ {
		evaluation := solver.magnitudeEvaluationAt(jd)
		residual, ok := solarEclipseRiseSetPhaseJunctionResidualAt(evaluation, longitude, latitude)
		if !ok {
			return SolarEclipsePathPoint{}, false
		}
		if math.Abs(residual[0]) <= 1e-10 && math.Abs(residual[1]) <= 1e-10 && math.Abs(residual[2]) <= 1e-10 {
			break
		}
		longitudeResidual, longitudeOK := solarEclipseRiseSetPhaseJunctionResidualAt(
			evaluation, longitude+geographicStep, latitude,
		)
		latitudeResidual, latitudeOK := solarEclipseRiseSetPhaseJunctionResidualAt(
			evaluation, longitude, latitude+geographicStep,
		)
		timeResidual, timeOK := solver.riseSetPhaseJunctionResidual(jd+timeStep, longitude, latitude)
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
		longitude = normalizeLongitude(longitude + delta[0])
		latitude += delta[1]
		jd += delta[2]
		if latitude <= -89.999999 || latitude >= 89.999999 {
			return SolarEclipsePathPoint{}, false
		}
	}
	residual, ok := solver.riseSetPhaseJunctionResidual(jd, longitude, latitude)
	if !ok || math.Abs(residual[0]) > 1e-7 || math.Abs(residual[1]) > 1e-7 || math.Abs(residual[2]) > 1e-8 {
		return SolarEclipsePathPoint{}, false
	}
	evaluation := solver.magnitudeEvaluationAt(jd)
	if evaluation.partialContactSecondDerivative(longitude, latitude) <= 0 {
		return SolarEclipsePathPoint{}, false
	}
	return SolarEclipsePathPoint{
		JDE: jd, Longitude: longitude, Latitude: latitude, SunAltitude: residual[2] / rad,
	}, true
}

func (solver solarEclipseSolver) refineRiseSetPhaseJunctionOnHorizon(
	seed SolarEclipsePathPoint,
) (SolarEclipsePathPoint, bool) {
	coordinates := [2]float64{
		solver.riseSetHorizonAngle(seed.JDE, seed.Longitude, seed.Latitude),
		0,
	}
	residualAt := func(value [2]float64) ([2]float64, float64, float64, float64, bool) {
		jd := seed.JDE + value[1]/1440
		longitude, latitude := solver.riseSetHorizonPointAt(jd, value[0])
		evaluation := solver.magnitudeEvaluationAt(jd)
		state := evaluation.center.stateAt(longitude*rad, latitude*rad, 0)
		residual := [2]float64{
			solarEclipsePartialContactGap(state),
			evaluation.partialContactDerivative(longitude, latitude),
		}
		return residual, jd, longitude, latitude, finite(residual[0]) && finite(residual[1])
	}
	const (
		angleStep      = 1e-4
		timeStepMinute = 1.0 / 60.0
	)
	for iteration := 0; iteration < 32; iteration++ {
		residual, _, _, _, ok := residualAt(coordinates)
		if !ok {
			return SolarEclipsePathPoint{}, false
		}
		if math.Abs(residual[0]) <= 1e-10 && math.Abs(residual[1]) <= 1e-9 {
			break
		}
		anglePlus, _, _, _, anglePlusOK := residualAt([2]float64{coordinates[0] + angleStep, coordinates[1]})
		angleMinus, _, _, _, angleMinusOK := residualAt([2]float64{coordinates[0] - angleStep, coordinates[1]})
		timePlus, _, _, _, timePlusOK := residualAt([2]float64{coordinates[0], coordinates[1] + timeStepMinute})
		timeMinus, _, _, _, timeMinusOK := residualAt([2]float64{coordinates[0], coordinates[1] - timeStepMinute})
		if !anglePlusOK || !angleMinusOK || !timePlusOK || !timeMinusOK {
			return SolarEclipsePathPoint{}, false
		}
		matrix := [2][2]float64{
			{(anglePlus[0] - angleMinus[0]) / (2 * angleStep), (timePlus[0] - timeMinus[0]) / (2 * timeStepMinute)},
			{(anglePlus[1] - angleMinus[1]) / (2 * angleStep), (timePlus[1] - timeMinus[1]) / (2 * timeStepMinute)},
		}
		determinant := matrix[0][0]*matrix[1][1] - matrix[0][1]*matrix[1][0]
		if !finite(determinant) || math.Abs(determinant) < 1e-18 {
			return SolarEclipsePathPoint{}, false
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
	residual, jd, longitude, latitude, ok := residualAt(coordinates)
	if !ok || math.Abs(residual[0]) > 1e-7 || math.Abs(residual[1]) > 1e-7 {
		return SolarEclipsePathPoint{}, false
	}
	return solver.refineRiseSetPhaseJunction(jd, longitude, latitude)
}

func (solver solarEclipseSolver) riseSetPhaseSegmentIsContinuous(
	start, end SolarEclipsePathPoint,
	phase RiseSetPhase,
	direction RiseSetDirection,
) bool {
	if math.Abs(end.JDE-start.JDE)*86400 < 0.1 {
		return false
	}
	totalDistance := solarEclipsePathDistanceKM(start, end)
	continuityToleranceKM := math.Max(50, 0.05*totalDistance)
	candidate := end
	for divisor := 2.0; divisor <= 1024; divisor *= 2 {
		jd := start.JDE + (end.JDE-start.JDE)/divisor
		seedAngle := solver.riseSetHorizonAngle(jd, candidate.Longitude, candidate.Latitude)
		next, ok := solver.riseSetPhasePointOnHorizon(jd, seedAngle, phase, direction)
		if !ok {
			return solarEclipsePathDistanceKM(start, candidate) <= continuityToleranceKM
		}
		candidate = next
	}
	return solarEclipsePathDistanceKM(start, candidate) <= continuityToleranceKM
}

func (solver solarEclipseSolver) riseSetPhasePointOnHorizon(
	jd, angle float64,
	phase RiseSetPhase,
	direction RiseSetDirection,
) (SolarEclipsePathPoint, bool) {
	evaluation := solver.magnitudeEvaluationAt(jd)
	greatest := phase == RiseSetPhaseGreatest
	valueAt := func(candidateAngle float64) (float64, float64, float64, bool) {
		longitude, latitude := solver.riseSetHorizonPointAt(jd, candidateAngle)
		value, ok := solarEclipseRiseSetPhaseResidual(evaluation, longitude, latitude, greatest)
		return value, longitude, latitude, ok && finite(value)
	}
	const angleStep = 1e-4
	angle = riseSetNormalizeRadians(angle)
	for iteration := 0; iteration < 24; iteration++ {
		value, _, _, ok := valueAt(angle)
		if !ok {
			return SolarEclipsePathPoint{}, false
		}
		if math.Abs(value) <= 1e-10 {
			break
		}
		before, _, _, beforeOK := valueAt(angle - angleStep)
		after, _, _, afterOK := valueAt(angle + angleStep)
		if !beforeOK || !afterOK {
			return SolarEclipsePathPoint{}, false
		}
		derivative := (after - before) / (2 * angleStep)
		if !finite(derivative) || math.Abs(derivative) < 1e-16 {
			return SolarEclipsePathPoint{}, false
		}
		delta := -value / derivative
		if math.Abs(delta) > 0.25 {
			delta = math.Copysign(0.25, delta)
		}
		angle = riseSetNormalizeRadians(angle + delta)
	}
	value, longitude, latitude, ok := valueAt(angle)
	if !ok || math.Abs(value) > 1e-7 {
		return SolarEclipsePathPoint{}, false
	}
	longitude, latitude, ok = riseSetRefineGeographicRoot(
		longitude,
		latitude,
		func(lon, lat float64) (float64, float64, bool) {
			state := evaluation.center.stateAt(lon*rad, lat*rad, 0)
			first, valid := solarEclipseRiseSetPhaseResidual(evaluation, lon, lat, greatest)
			return first, state.sunAltitudeRad, valid && finite(state.sunAltitudeRad)
		},
	)
	if !ok {
		return SolarEclipsePathPoint{}, false
	}
	point, key, valid := evaluation.classify(longitude, latitude, greatest)
	if !valid || key.phase != phase || key.direction != direction {
		return SolarEclipsePathPoint{}, false
	}
	return point, true
}

func (solver solarEclipseSolver) riseSetFoldBridgesPhaseJunction(
	junction SolarEclipsePathPoint,
	endpoint solarEclipseRiseSetEndpointRef,
	fold SolarEclipsePathPoint,
	phase RiseSetPhase,
	direction RiseSetDirection,
) bool {
	const timeToleranceDays = 1e-8
	if endpoint.atStart {
		if fold.JDE > math.Min(junction.JDE, endpoint.point.JDE)+timeToleranceDays {
			return false
		}
	} else if fold.JDE < math.Max(junction.JDE, endpoint.point.JDE)-timeToleranceDays {
		return false
	}
	if solarEclipsePathDistanceKM(fold, junction) > 6000 ||
		solarEclipsePathDistanceKM(fold, endpoint.point) > 6000 {
		return false
	}
	evaluation := solver.magnitudeEvaluationAt(fold.JDE)
	_, key, valid := evaluation.classify(
		fold.Longitude, fold.Latitude, phase == RiseSetPhaseGreatest,
	)
	return valid && key.phase == phase && key.direction == direction
}

func (solver solarEclipseSolver) refineRiseSetFoldNearPhaseJunction(
	junction SolarEclipsePathPoint,
	endpoint solarEclipseRiseSetEndpointRef,
	phase RiseSetPhase,
	direction RiseSetDirection,
	stepDays float64,
) (SolarEclipsePathPoint, bool) {
	key := solarEclipseRiseSetCurveKey{phase: phase, direction: direction}
	greatest := phase == RiseSetPhaseGreatest
	for _, sign := range []float64{-1, 1} {
		for divisor := 1024.0; divisor >= 1; divisor /= 2 {
			fraction := 1 / divisor
			jd := junction.JDE + sign*fraction*stepDays
			roots := solver.riseSetPointsAt(jd, solarEclipseRiseSetBoundaryPoints)[key]
			if len(roots) < 2 {
				continue
			}
			for first := 0; first < len(roots); first++ {
				for second := first + 1; second < len(roots); second++ {
					fold, ok := solver.refineRiseSetFold(roots[first], roots[second], greatest)
					if ok && solver.riseSetFoldBridgesPhaseJunction(
						junction, endpoint, fold, phase, direction,
					) {
						return fold, true
					}
				}
			}
		}
	}
	return SolarEclipsePathPoint{}, false
}

func (solver solarEclipseSolver) riseSetPhaseJunctionResidual(
	jd, longitude, latitude float64,
) ([3]float64, bool) {
	evaluation := solver.magnitudeEvaluationAt(jd)
	return solarEclipseRiseSetPhaseJunctionResidualAt(evaluation, longitude, latitude)
}

func solarEclipseRiseSetPhaseJunctionResidualAt(
	evaluation solarEclipseRiseSetEvaluation,
	longitude, latitude float64,
) ([3]float64, bool) {
	state := evaluation.center.stateAt(longitude*rad, latitude*rad, 0)
	residual := [3]float64{
		solarEclipsePartialContactGap(state),
		evaluation.partialContactDerivative(longitude, latitude),
		state.sunAltitudeRad,
	}
	return residual, finite(residual[0]) && finite(residual[1]) && finite(residual[2])
}

func (solver solarEclipseSolver) refineRiseSetFold(
	first, second SolarEclipsePathPoint,
	greatest bool,
) (SolarEclipsePathPoint, bool) {
	seedJDE := (first.JDE + second.JDE) / 2
	firstAngle := solver.riseSetHorizonAngle(seedJDE, first.Longitude, first.Latitude)
	secondAngle := solver.riseSetHorizonAngle(seedJDE, second.Longitude, second.Latitude)
	deltaAngle := math.Remainder(secondAngle-firstAngle, 2*math.Pi)
	coordinates := [2]float64{riseSetNormalizeRadians(firstAngle + deltaAngle/2), 0}
	const (
		angleStep      = 1e-4
		timeStepMinute = 1.0 / 60.0
	)
	for iteration := 0; iteration < 32; iteration++ {
		residual, ok := solver.riseSetFoldHorizonResidual(seedJDE, coordinates, greatest)
		if !ok {
			return SolarEclipsePathPoint{}, false
		}
		if math.Abs(residual[0]) <= 1e-10 && math.Abs(residual[1]) <= 1e-9 {
			break
		}
		anglePlus, anglePlusOK := solver.riseSetFoldHorizonResidual(
			seedJDE, [2]float64{coordinates[0] + angleStep, coordinates[1]}, greatest,
		)
		angleMinus, angleMinusOK := solver.riseSetFoldHorizonResidual(
			seedJDE, [2]float64{coordinates[0] - angleStep, coordinates[1]}, greatest,
		)
		timePlus, timePlusOK := solver.riseSetFoldHorizonResidual(
			seedJDE, [2]float64{coordinates[0], coordinates[1] + timeStepMinute}, greatest,
		)
		timeMinus, timeMinusOK := solver.riseSetFoldHorizonResidual(
			seedJDE, [2]float64{coordinates[0], coordinates[1] - timeStepMinute}, greatest,
		)
		if !anglePlusOK || !angleMinusOK || !timePlusOK || !timeMinusOK {
			return SolarEclipsePathPoint{}, false
		}
		matrix := [2][2]float64{}
		for row := 0; row < 2; row++ {
			matrix[row][0] = (anglePlus[row] - angleMinus[row]) / (2 * angleStep)
			matrix[row][1] = (timePlus[row] - timeMinus[row]) / (2 * timeStepMinute)
		}
		determinant := matrix[0][0]*matrix[1][1] - matrix[0][1]*matrix[1][0]
		if !finite(determinant) || math.Abs(determinant) < 1e-18 {
			return SolarEclipsePathPoint{}, false
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
	jd := seedJDE + coordinates[1]/1440.0
	longitude, latitude := solver.riseSetHorizonPointAt(jd, coordinates[0])
	return solver.refineRiseSetFoldPoint(jd, longitude, latitude, greatest)
}

func (solver solarEclipseSolver) riseSetFoldHorizonResidual(
	seedJDE float64,
	coordinates [2]float64,
	greatest bool,
) ([2]float64, bool) {
	const derivativeStep = 1e-4
	jd := seedJDE + coordinates[1]/1440.0
	evaluation := solver.magnitudeEvaluationAt(jd)
	centerLongitude, centerLatitude := solarEclipseRiseSetHorizonCenter(evaluation)
	valueAt := func(angle float64) (float64, bool) {
		longitude, latitude := riseSetHorizonPoint(centerLongitude, centerLatitude, angle)
		return solarEclipseRiseSetPhaseResidual(evaluation, longitude, latitude, greatest)
	}
	center, centerOK := valueAt(coordinates[0])
	before, beforeOK := valueAt(coordinates[0] - derivativeStep)
	after, afterOK := valueAt(coordinates[0] + derivativeStep)
	residual := [2]float64{center, (after - before) / (2 * derivativeStep)}
	return residual, centerOK && beforeOK && afterOK && finite(residual[1])
}

func (solver solarEclipseSolver) riseSetHorizonPointAt(jd, angle float64) (float64, float64) {
	evaluation := solver.magnitudeEvaluationAt(jd)
	centerLongitude, centerLatitude := solarEclipseRiseSetHorizonCenter(evaluation)
	return riseSetHorizonPoint(centerLongitude, centerLatitude, angle)
}

func (solver solarEclipseSolver) riseSetHorizonAngle(
	jd, longitude, latitude float64,
) float64 {
	evaluation := solver.magnitudeEvaluationAt(jd)
	centerLongitude, centerLatitude := solarEclipseRiseSetHorizonCenter(evaluation)
	centerLon, centerLat := centerLongitude*rad, centerLatitude*rad
	center := [3]float64{
		math.Cos(centerLat) * math.Cos(centerLon),
		math.Cos(centerLat) * math.Sin(centerLon),
		math.Sin(centerLat),
	}
	reference := [3]float64{0, 0, 1}
	if math.Abs(center[2]) > 0.9 {
		reference = [3]float64{1, 0, 0}
	}
	first := riseSetUnitVector(riseSetCross(reference, center))
	second := riseSetUnitVector(riseSetCross(center, first))
	lon, lat := longitude*rad, latitude*rad
	point := [3]float64{math.Cos(lat) * math.Cos(lon), math.Cos(lat) * math.Sin(lon), math.Sin(lat)}
	return riseSetNormalizeRadians(math.Atan2(dotSolarEclipse3(point, second), dotSolarEclipse3(point, first)))
}

func solarEclipseRiseSetHorizonCenter(evaluation solarEclipseRiseSetEvaluation) (float64, float64) {
	sun := solarEclipseXYZToLLR(
		evaluation.center.sunXYZ[0], evaluation.center.sunXYZ[1], evaluation.center.sunXYZ[2],
	)
	return normalizeLongitude((sun[0] - evaluation.center.gst) / rad), sun[1] / rad
}

func (solver solarEclipseSolver) refineRiseSetFoldPoint(
	jd, longitude, latitude float64,
	greatest bool,
) (SolarEclipsePathPoint, bool) {
	const (
		geographicStep = 1e-3
		timeStep       = 1.0 / 86400.0
	)
	coordinates := [3]float64{longitude, latitude, jd}
	for iteration := 0; iteration < 32; iteration++ {
		evaluation := solver.magnitudeEvaluationAt(coordinates[2])
		residual, ok := solarEclipseRiseSetFoldResidualAt(evaluation, coordinates[0], coordinates[1], greatest)
		if !ok {
			return SolarEclipsePathPoint{}, false
		}
		if math.Abs(residual[0]) <= 1e-10 && math.Abs(residual[1]) <= 1e-10 && math.Abs(residual[2]) <= 1e-10 {
			break
		}
		steps := [3]float64{geographicStep, geographicStep, timeStep}
		matrix := [3][3]float64{}
		for column := 0; column < 3; column++ {
			plus, minus := coordinates, coordinates
			plus[column] += steps[column]
			minus[column] -= steps[column]
			var plusResidual, minusResidual [3]float64
			var plusOK, minusOK bool
			if column < 2 {
				plusResidual, plusOK = solarEclipseRiseSetFoldResidualAt(evaluation, plus[0], plus[1], greatest)
				minusResidual, minusOK = solarEclipseRiseSetFoldResidualAt(evaluation, minus[0], minus[1], greatest)
			} else {
				plusResidual, plusOK = solver.riseSetFoldResidual(plus[2], plus[0], plus[1], greatest)
				minusResidual, minusOK = solver.riseSetFoldResidual(minus[2], minus[0], minus[1], greatest)
			}
			if !plusOK || !minusOK {
				return SolarEclipsePathPoint{}, false
			}
			for row := 0; row < 3; row++ {
				matrix[row][column] = (plusResidual[row] - minusResidual[row]) / (2 * steps[column])
			}
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
		if math.Abs(delta[2]) > 2.0/1440.0 {
			delta[2] = math.Copysign(2.0/1440.0, delta[2])
		}
		for index := range coordinates {
			coordinates[index] += delta[index]
		}
		coordinates[0] = normalizeLongitude(coordinates[0])
		if coordinates[1] <= -89.999999 || coordinates[1] >= 89.999999 {
			return SolarEclipsePathPoint{}, false
		}
	}
	residual, ok := solver.riseSetFoldResidual(coordinates[2], coordinates[0], coordinates[1], greatest)
	if !ok || math.Abs(residual[0]) > 1e-7 || math.Abs(residual[1]) > 1e-8 || math.Abs(residual[2]) > 1e-7 {
		return SolarEclipsePathPoint{}, false
	}
	evaluation := solver.magnitudeEvaluationAt(coordinates[2])
	state := evaluation.center.stateAt(coordinates[0]*rad, coordinates[1]*rad, 0)
	return SolarEclipsePathPoint{
		JDE: coordinates[2], Longitude: coordinates[0], Latitude: coordinates[1], SunAltitude: state.sunAltitudeRad / rad,
	}, true
}

func (solver solarEclipseSolver) riseSetFoldResidual(
	jd, longitude, latitude float64,
	greatest bool,
) ([3]float64, bool) {
	evaluation := solver.magnitudeEvaluationAt(jd)
	return solarEclipseRiseSetFoldResidualAt(evaluation, longitude, latitude, greatest)
}

func solarEclipseRiseSetFoldResidualAt(
	evaluation solarEclipseRiseSetEvaluation,
	longitude, latitude float64,
	greatest bool,
) ([3]float64, bool) {
	const step = 1e-3
	valueAt := func(lon, lat float64) ([2]float64, bool) {
		first, ok := solarEclipseRiseSetPhaseResidual(evaluation, lon, lat, greatest)
		state := evaluation.center.stateAt(lon*rad, lat*rad, 0)
		return [2]float64{first, state.sunAltitudeRad}, ok && finite(state.sunAltitudeRad)
	}
	center, centerOK := valueAt(longitude, latitude)
	lonPlus, lonPlusOK := valueAt(longitude+step, latitude)
	lonMinus, lonMinusOK := valueAt(longitude-step, latitude)
	latPlus, latPlusOK := valueAt(longitude, latitude+step)
	latMinus, latMinusOK := valueAt(longitude, latitude-step)
	if !centerOK || !lonPlusOK || !lonMinusOK || !latPlusOK || !latMinusOK {
		return [3]float64{}, false
	}
	dFirstLon := (lonPlus[0] - lonMinus[0]) / (2 * step)
	dFirstLat := (latPlus[0] - latMinus[0]) / (2 * step)
	dAltitudeLon := (lonPlus[1] - lonMinus[1]) / (2 * step)
	dAltitudeLat := (latPlus[1] - latMinus[1]) / (2 * step)
	residual := [3]float64{
		center[0],
		center[1],
		dFirstLon*dAltitudeLat - dFirstLat*dAltitudeLon,
	}
	return residual, finite(residual[2])
}

func solarEclipseRiseSetPhaseResidual(
	evaluation solarEclipseRiseSetEvaluation,
	longitude, latitude float64,
	greatest bool,
) (float64, bool) {
	if greatest {
		value := evaluation.separationDerivative(longitude, latitude)
		return value, finite(value)
	}
	state := evaluation.center.stateAt(longitude*rad, latitude*rad, 0)
	value := solarEclipsePartialContactGap(state)
	return value, finite(value)
}

func (solver solarEclipseSolver) refineRiseSetCurveSpacing(curve *SolarEclipseRiseSetCurve) {
	if curve == nil {
		return
	}
	solver = solver.withLocalEphemeris()
	for segmentIndex, segment := range curve.Segments {
		if len(segment) < 2 {
			continue
		}
		refined := make([]SolarEclipsePathPoint, 1, len(segment))
		refined[0] = segment[0]
		for pointIndex := 1; pointIndex < len(segment); pointIndex++ {
			refined = solver.appendRefinedRiseSetSegment(
				refined, segment[pointIndex-1], segment[pointIndex], curve.Phase, curve.Direction, 0,
			)
		}
		curve.Segments[segmentIndex] = refined
	}
}

func (solver solarEclipseSolver) appendRefinedRiseSetSegment(
	points []SolarEclipsePathPoint,
	start, end SolarEclipsePathPoint,
	phase RiseSetPhase,
	direction RiseSetDirection,
	depth int,
) []SolarEclipsePathPoint {
	if depth >= 12 || end.JDE-start.JDE <= solarEclipsePathMinStepDays {
		return append(points, end)
	}
	jd := (start.JDE + end.JDE) / 2
	longitude := normalizeLongitude(start.Longitude + math.Remainder(end.Longitude-start.Longitude, 360)/2)
	latitude := (start.Latitude + end.Latitude) / 2
	// A comfortably straight candidate needs no new output vertex. Reserve
	// half the chord budget for prediction error; every inserted vertex still
	// uses the exact ephemeris and the original phase residual checks.
	short := solarEclipsePathDistanceKM(start, end) <= solarEclipseRiseSetTargetSpacingKM
	if short && solver.localEphemeris != nil {
		candidate, key, valid := solarEclipseRefineRiseSetMiddle(solver.magnitudeCandidateEvaluationAt(jd), longitude, latitude, phase)
		if valid && key.phase == phase && key.direction == direction &&
			solarEclipseRiseSetChordDeviationKM(candidate, start, end) <= solarEclipseRiseSetChordToleranceKM/2 {
			return append(points, end)
		}
	}
	middle, key, valid := solarEclipseRefineRiseSetMiddle(solver.magnitudeEvaluationAt(jd), longitude, latitude, phase)
	if !valid || key.phase != phase || key.direction != direction {
		return append(points, end)
	}
	if short &&
		solarEclipseRiseSetChordDeviationKM(middle, start, end) <= solarEclipseRiseSetChordToleranceKM {
		return append(points, end)
	}
	points = solver.appendRefinedRiseSetSegment(points, start, middle, phase, direction, depth+1)
	return solver.appendRefinedRiseSetSegment(points, middle, end, phase, direction, depth+1)
}

func solarEclipseRefineRiseSetMiddle(
	evaluation solarEclipseRiseSetEvaluation,
	longitude, latitude float64,
	phase RiseSetPhase,
) (SolarEclipsePathPoint, solarEclipseRiseSetCurveKey, bool) {
	greatest := phase == RiseSetPhaseGreatest
	longitude, latitude, ok := riseSetRefineGeographicRoot(longitude, latitude,
		func(lon, lat float64) (float64, float64, bool) {
			first, valid := solarEclipseRiseSetPhaseResidual(evaluation, lon, lat, greatest)
			state := evaluation.center.stateAt(lon*rad, lat*rad, 0)
			return first, state.sunAltitudeRad, valid && finite(state.sunAltitudeRad)
		})
	if !ok {
		return SolarEclipsePathPoint{}, solarEclipseRiseSetCurveKey{}, false
	}
	return evaluation.classify(longitude, latitude, greatest)
}

const solarEclipseRiseSetChordToleranceKM = 0.25

func solarEclipseRiseSetChordDeviationKM(point, start, end SolarEclipsePathPoint) float64 {
	first := solarEclipseLLRToXYZ(start.Longitude*rad, start.Latitude*rad, 1)
	last := solarEclipseLLRToXYZ(end.Longitude*rad, end.Latitude*rad, 1)
	middle := solarEclipseLLRToXYZ(point.Longitude*rad, point.Latitude*rad, 1)
	normal := solarEclipseRiseSetCross(first, last)
	norm := math.Sqrt(dotSolarEclipse3(normal, normal))
	if norm < 1e-12 || dotSolarEclipse3(solarEclipseRiseSetCross(first, middle), normal) < 0 ||
		dotSolarEclipse3(solarEclipseRiseSetCross(middle, last), normal) < 0 {
		return math.Min(solarEclipsePathDistanceKM(point, start), solarEclipsePathDistanceKM(point, end))
	}
	return solarEclipseEarthEquatorialRadiusKM * math.Asin(math.Min(1, math.Abs(dotSolarEclipse3(middle, normal))/norm))
}

func solarEclipseRiseSetCross(first, second [3]float64) [3]float64 {
	return [3]float64{
		first[1]*second[2] - first[2]*second[1],
		first[2]*second[0] - first[0]*second[2],
		first[0]*second[1] - first[1]*second[0],
	}
}
