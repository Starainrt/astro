package basic

import (
	"math"
	"time"

	"github.com/starainrt/astro/internal/geodata"
)

func refineOccultationRiseSetFold(
	first, second OccultationPathPoint,
	atStart bool,
	greatest bool,
	stepDays float64,
	location *time.Location,
	cache *occultationRiseSetEvaluationCache,
) (OccultationPathPoint, bool) {
	newton := func() (OccultationPathPoint, bool) {
		seedTT := (occultationTimeToTT(first.Time) + occultationTimeToTT(second.Time)) / 2
		firstAngle := occultationRiseSetHorizonAngle(seedTT, first.Longitude, first.Latitude, cache.context)
		secondAngle := occultationRiseSetHorizonAngle(seedTT, second.Longitude, second.Latitude, cache.context)
		coordinates := [2]float64{riseSetNormalizeRadians(firstAngle + math.Remainder(secondAngle-firstAngle, 2*math.Pi)/2), 0}
		const angleStep = 1e-4
		for iteration := 0; iteration < 32; iteration++ {
			residual, ok := occultationRiseSetFoldHorizonResidual(seedTT, coordinates, greatest, cache)
			if !ok {
				return OccultationPathPoint{}, false
			}
			if math.Abs(residual[0]) <= 1e-10 && math.Abs(residual[1]) <= 1e-9 {
				break
			}
			plusAngle, plusAngleOK := occultationRiseSetFoldHorizonResidual(seedTT, [2]float64{coordinates[0] + angleStep, coordinates[1]}, greatest, cache)
			minusAngle, minusAngleOK := occultationRiseSetFoldHorizonResidual(seedTT, [2]float64{coordinates[0] - angleStep, coordinates[1]}, greatest, cache)
			timeStep := 1.0 / 60.0
			plusTime, plusTimeOK := occultationRiseSetFoldHorizonResidual(seedTT, [2]float64{coordinates[0], coordinates[1] + timeStep}, greatest, cache)
			minusTime, minusTimeOK := occultationRiseSetFoldHorizonResidual(seedTT, [2]float64{coordinates[0], coordinates[1] - timeStep}, greatest, cache)
			if !plusAngleOK || !minusAngleOK || !plusTimeOK || !minusTimeOK {
				return OccultationPathPoint{}, false
			}
			matrix := [2][2]float64{
				{(plusAngle[0] - minusAngle[0]) / (2 * angleStep), (plusTime[0] - minusTime[0]) / (2 * timeStep)},
				{(plusAngle[1] - minusAngle[1]) / (2 * angleStep), (plusTime[1] - minusTime[1]) / (2 * timeStep)},
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
		jd := seedTT + coordinates[1]/1440
		longitude, latitude, ok := occultationRiseSetHorizonPoint(jd, coordinates[0], cache.context)
		if !ok {
			return OccultationPathPoint{}, false
		}
		return refineOccultationRiseSetFoldPoint(jd, longitude, latitude, greatest, location, cache)
	}
	if point, ok := newton(); ok {
		return point, true
	}
	return refineOccultationRiseSetFoldBracketed(
		first, second, atStart, greatest, stepDays, location, cache,
	)
}

func refineOccultationRiseSetFoldBracketed(
	first, second OccultationPathPoint,
	atStart bool,
	greatest bool,
	stepDays float64,
	location *time.Location,
	cache *occultationRiseSetEvaluationCache,
) (OccultationPathPoint, bool) {
	sampleTT := (occultationTimeToTT(first.Time) + occultationTimeToTT(second.Time)) / 2
	firstAngle := occultationRiseSetHorizonAngle(sampleTT, first.Longitude, first.Latitude, cache.context)
	secondAngle := occultationRiseSetHorizonAngle(sampleTT, second.Longitude, second.Latitude, cache.context)
	sampleAngle := riseSetNormalizeRadians(firstAngle + math.Remainder(secondAngle-firstAngle, 2*math.Pi)/2)
	sampleAngle, sampleValue, ok := occultationRiseSetFoldExtremumAt(sampleTT, sampleAngle, greatest, cache)
	if !ok {
		return OccultationPathPoint{}, false
	}
	direction := 1.0
	if atStart {
		direction = -1
	}
	var otherTT, otherAngle, otherValue float64
	bracketed := false
	for step := 1; step <= 4; step++ {
		otherTT = sampleTT + direction*float64(step)*stepDays
		otherAngle, otherValue, ok = occultationRiseSetFoldExtremumAt(otherTT, sampleAngle, greatest, cache)
		if !ok {
			continue
		}
		if sampleValue == 0 || otherValue == 0 || sampleValue*otherValue < 0 {
			bracketed = true
			break
		}
	}
	if !bracketed {
		return OccultationPathPoint{}, false
	}

	leftTT, leftAngle, leftValue := sampleTT, sampleAngle, sampleValue
	rightTT, rightAngle, rightValue := otherTT, otherAngle, otherValue
	for iteration := 0; iteration < 64 && math.Abs(rightTT-leftTT) > occultationRiseSetFoldRootToleranceDays; iteration++ {
		middleTT := (leftTT + rightTT) / 2
		middleSeed := riseSetNormalizeRadians(leftAngle + math.Remainder(rightAngle-leftAngle, 2*math.Pi)/2)
		middleAngle, middleValue, middleOK := occultationRiseSetFoldExtremumAt(middleTT, middleSeed, greatest, cache)
		if !middleOK {
			return OccultationPathPoint{}, false
		}
		if leftValue == 0 || leftValue*middleValue <= 0 {
			rightTT, rightAngle, rightValue = middleTT, middleAngle, middleValue
		} else {
			leftTT, leftAngle, leftValue = middleTT, middleAngle, middleValue
		}
	}
	_ = rightValue
	rootTT := (leftTT + rightTT) / 2
	rootSeed := riseSetNormalizeRadians(leftAngle + math.Remainder(rightAngle-leftAngle, 2*math.Pi)/2)
	rootAngle, rootValue, ok := occultationRiseSetFoldExtremumAt(rootTT, rootSeed, greatest, cache)
	if !ok {
		return OccultationPathPoint{}, false
	}
	longitude, latitude, ok := occultationRiseSetHorizonPoint(rootTT, rootAngle, cache.context)
	if !ok {
		return OccultationPathPoint{}, false
	}
	if point, refined := refineOccultationRiseSetFoldPoint(rootTT, longitude, latitude, greatest, location, cache); refined {
		return point, true
	}
	state := cache.context(rootTT).stateAt(longitude, latitude)
	if !state.valid || math.Abs(rootValue) > 1e-7 || math.Abs(state.moonAltitude) > 1e-8 {
		return OccultationPathPoint{}, false
	}
	return OccultationPathPoint{
		Time: occultationTTToLocation(rootTT, location), Longitude: longitude,
		Latitude: latitude, MoonAltitude: state.moonAltitude,
	}, true
}

func occultationRiseSetFoldExtremumAt(
	tt, angle float64,
	greatest bool,
	cache *occultationRiseSetEvaluationCache,
) (float64, float64, bool) {
	const angleStep = 1e-4
	angle = riseSetNormalizeRadians(angle)
	for iteration := 0; iteration < 24; iteration++ {
		residual, ok := occultationRiseSetFoldHorizonResidual(tt, [2]float64{angle, 0}, greatest, cache)
		if !ok {
			return 0, 0, false
		}
		if math.Abs(residual[1]) <= 1e-10 {
			return angle, residual[0], true
		}
		before, beforeOK := occultationRiseSetFoldHorizonResidual(tt, [2]float64{angle - angleStep, 0}, greatest, cache)
		after, afterOK := occultationRiseSetFoldHorizonResidual(tt, [2]float64{angle + angleStep, 0}, greatest, cache)
		if !beforeOK || !afterOK {
			return 0, 0, false
		}
		secondDerivative := (after[1] - before[1]) / (2 * angleStep)
		if !finite(secondDerivative) || math.Abs(secondDerivative) < 1e-16 {
			return 0, 0, false
		}
		delta := -residual[1] / secondDerivative
		if math.Abs(delta) > 0.25 {
			delta = math.Copysign(0.25, delta)
		}
		angle = riseSetNormalizeRadians(angle + delta)
	}
	residual, ok := occultationRiseSetFoldHorizonResidual(tt, [2]float64{angle, 0}, greatest, cache)
	return angle, residual[0], ok && finite(residual[0]) && finite(residual[1]) && math.Abs(residual[1]) <= 1e-7
}

func occultationRiseSetHorizonCenter(context occultationRiseSetContext) (float64, float64) {
	return geodata.NormalizeLongitude(context.moonRA - context.siderealDegrees), context.moonDec
}

func occultationRiseSetHorizonPoint(
	tt, angle float64,
	contextAt occultationRiseSetContextFunc,
) (float64, float64, bool) {
	return occultationRiseSetHorizonPointFromContext(contextAt(tt), angle)
}

func occultationRiseSetHorizonPointFromContext(
	context occultationRiseSetContext,
	angle float64,
) (float64, float64, bool) {
	if !context.valid {
		return 0, 0, false
	}
	center, first, second, basisOK := occultationRiseSetHorizonBasis(context)
	if !basisOK {
		return 0, 0, false
	}
	tangent := occultationRiseSetCircleTangent(first, second, angle)
	pointAt := func(radius float64) (float64, float64, float64, bool) {
		longitude, latitude := occultationRiseSetCirclePointFromTangent(center, tangent, radius)
		residual, ok := context.moonHorizonResidual(longitude, latitude)
		return longitude, latitude, residual, ok
	}
	left, right := 80*rad, 100*rad
	_, _, leftValue, leftOK := pointAt(left)
	_, _, rightValue, rightOK := pointAt(right)
	if !leftOK || !rightOK || leftValue*rightValue > 0 {
		return 0, 0, false
	}
	weightedLeft, weightedRight := leftValue, rightValue
	lastSide := 0
	for iteration := 0; iteration < 12; iteration++ {
		denominator := weightedRight - weightedLeft
		middle := (left + right) / 2
		if finite(denominator) && math.Abs(denominator) > 1e-18 {
			candidate := (left*weightedRight - right*weightedLeft) / denominator
			if candidate > left && candidate < right {
				middle = candidate
			}
		}
		longitude, latitude, middleValue, middleOK := pointAt(middle)
		if !middleOK {
			return 0, 0, false
		}
		if math.Abs(middleValue) <= 1e-6 {
			return longitude, latitude, true
		}
		if leftValue*middleValue <= 0 {
			right, rightValue, weightedRight = middle, middleValue, middleValue
			if lastSide < 0 {
				weightedLeft *= 0.5
			} else {
				weightedLeft = leftValue
			}
			lastSide = -1
		} else {
			left, leftValue, weightedLeft = middle, middleValue, middleValue
			if lastSide > 0 {
				weightedRight *= 0.5
			} else {
				weightedRight = rightValue
			}
			lastSide = 1
		}
	}
	for iteration := 0; iteration < 32; iteration++ {
		middle := (left + right) / 2
		_, _, middleValue, middleOK := pointAt(middle)
		if !middleOK {
			return 0, 0, false
		}
		if leftValue*middleValue <= 0 {
			right, rightValue = middle, middleValue
		} else {
			left, leftValue = middle, middleValue
		}
	}
	longitude, latitude, _, ok := pointAt((left + right) / 2)
	return longitude, latitude, ok
}

func occultationRiseSetHorizonBasis(context occultationRiseSetContext) ([3]float64, [3]float64, [3]float64, bool) {
	centerLongitude, centerLatitude := occultationRiseSetHorizonCenter(context)
	longitude := centerLongitude * rad
	latitude := centerLatitude * rad
	center := [3]float64{
		math.Cos(latitude) * math.Cos(longitude),
		math.Cos(latitude) * math.Sin(longitude),
		math.Sin(latitude),
	}
	reference := [3]float64{0, 0, 1}
	if math.Abs(center[2]) > 0.9 {
		reference = [3]float64{1, 0, 0}
	}
	first := riseSetUnitVector(riseSetCross(reference, center))
	second := riseSetUnitVector(riseSetCross(center, first))
	return center, first, second, finite(center[0]) && finite(first[0]) && finite(second[0])
}

func occultationRiseSetCirclePointFromBasis(center, first, second [3]float64, angle, radius float64) (float64, float64) {
	return occultationRiseSetCirclePointFromTangent(
		center, occultationRiseSetCircleTangent(first, second, angle), radius,
	)
}

func occultationRiseSetCircleTangent(first, second [3]float64, angle float64) [3]float64 {
	sinAngle, cosAngle := math.Sincos(angle)
	return [3]float64{
		first[0]*cosAngle + second[0]*sinAngle,
		first[1]*cosAngle + second[1]*sinAngle,
		first[2]*cosAngle + second[2]*sinAngle,
	}
}

func occultationRiseSetCirclePointFromTangent(center, tangent [3]float64, radius float64) (float64, float64) {
	sinRadius, cosRadius := math.Sincos(radius)
	point := [3]float64{
		center[0]*cosRadius + tangent[0]*sinRadius,
		center[1]*cosRadius + tangent[1]*sinRadius,
		center[2]*cosRadius + tangent[2]*sinRadius,
	}
	return geodata.NormalizeLongitude(math.Atan2(point[1], point[0]) / rad),
		math.Asin(math.Max(-1, math.Min(1, point[2]))) / rad
}

func occultationRiseSetHorizonAngle(tt, longitude, latitude float64, contextAt occultationRiseSetContextFunc) float64 {
	_, first, second, ok := occultationRiseSetHorizonBasis(contextAt(tt))
	if !ok {
		return math.NaN()
	}
	pointLon, pointLat := longitude*rad, latitude*rad
	point := [3]float64{math.Cos(pointLat) * math.Cos(pointLon), math.Cos(pointLat) * math.Sin(pointLon), math.Sin(pointLat)}
	return riseSetNormalizeRadians(math.Atan2(point[0]*second[0]+point[1]*second[1]+point[2]*second[2], point[0]*first[0]+point[1]*first[1]+point[2]*first[2]))
}

func occultationRiseSetFoldHorizonResidual(seedTT float64, coordinates [2]float64, greatest bool, cache *occultationRiseSetEvaluationCache) ([2]float64, bool) {
	jd := seedTT + coordinates[1]/1440
	evaluation := cache.evaluation(jd)
	valueAt := func(angle float64) (float64, bool) {
		longitude, latitude, ok := occultationRiseSetHorizonPoint(jd, angle, cache.context)
		if !ok {
			return 0, false
		}
		return occultationRiseSetPhaseResidual(evaluation, longitude, latitude, greatest)
	}
	center, centerOK := valueAt(coordinates[0])
	before, beforeOK := valueAt(coordinates[0] - 1e-4)
	after, afterOK := valueAt(coordinates[0] + 1e-4)
	return [2]float64{center, (after - before) / (2e-4)}, centerOK && beforeOK && afterOK && finite(center) && finite(before) && finite(after)
}

func refineOccultationRiseSetFoldPoint(
	tt, longitude, latitude float64,
	greatest bool,
	location *time.Location,
	cache *occultationRiseSetEvaluationCache,
) (OccultationPathPoint, bool) {
	coordinates := [3]float64{longitude, latitude, tt}
	for iteration := 0; iteration < 32; iteration++ {
		evaluation := cache.evaluation(coordinates[2])
		residual, ok := occultationRiseSetFoldResidualAt(evaluation, coordinates[0], coordinates[1], greatest)
		if !ok {
			return OccultationPathPoint{}, false
		}
		if math.Abs(residual[0]) <= 1e-10 && math.Abs(residual[1]) <= 1e-10 && math.Abs(residual[2]) <= 1e-10 {
			break
		}
		steps := [3]float64{1e-3, 1e-3, 1.0 / 86400.0}
		matrix := [3][3]float64{}
		for column := range steps {
			plus, minus := coordinates, coordinates
			plus[column] += steps[column]
			minus[column] -= steps[column]
			plusEval := cache.evaluation(plus[2])
			minusEval := cache.evaluation(minus[2])
			plusResidual, plusOK := occultationRiseSetFoldResidualAt(plusEval, plus[0], plus[1], greatest)
			minusResidual, minusOK := occultationRiseSetFoldResidualAt(minusEval, minus[0], minus[1], greatest)
			if !plusOK || !minusOK {
				return OccultationPathPoint{}, false
			}
			for row := range matrix {
				matrix[row][column] = (plusResidual[row] - minusResidual[row]) / (2 * steps[column])
			}
		}
		delta, ok := solveSolarEclipse3x3(matrix, [3]float64{-residual[0], -residual[1], -residual[2]})
		if !ok {
			return OccultationPathPoint{}, false
		}
		geographicScale := math.Max(math.Abs(delta[0]), math.Abs(delta[1]))
		if geographicScale > 2 {
			delta[0] *= 2 / geographicScale
			delta[1] *= 2 / geographicScale
		}
		if math.Abs(delta[2]) > 2.0/1440 {
			delta[2] = math.Copysign(2.0/1440, delta[2])
		}
		for index := range coordinates {
			coordinates[index] += delta[index]
		}
		coordinates[0] = geodata.NormalizeLongitude(coordinates[0])
		if coordinates[1] <= -89.999999 || coordinates[1] >= 89.999999 {
			return OccultationPathPoint{}, false
		}
	}
	evaluation := cache.evaluation(coordinates[2])
	residual, ok := occultationRiseSetFoldResidualAt(evaluation, coordinates[0], coordinates[1], greatest)
	if !ok || math.Abs(residual[0]) > 1e-7 || math.Abs(residual[1]) > 1e-8 || math.Abs(residual[2]) > 1e-7 {
		return OccultationPathPoint{}, false
	}
	state := evaluation.center.stateAt(coordinates[0], coordinates[1])
	return OccultationPathPoint{Time: occultationTTToLocation(coordinates[2], location), Longitude: coordinates[0], Latitude: coordinates[1], MoonAltitude: state.moonAltitude}, true
}

func occultationRiseSetFoldResidualAt(evaluation occultationRiseSetEvaluation, longitude, latitude float64, greatest bool) ([3]float64, bool) {
	valueAt := func(lon, lat float64) ([2]float64, bool) {
		first, ok := occultationRiseSetPhaseResidual(evaluation, lon, lat, greatest)
		state := evaluation.center.stateAt(lon, lat)
		return [2]float64{first, state.moonAltitude}, ok && state.valid
	}
	center, centerOK := valueAt(longitude, latitude)
	lonPlus, lonPlusOK := valueAt(longitude+1e-3, latitude)
	lonMinus, lonMinusOK := valueAt(longitude-1e-3, latitude)
	latPlus, latPlusOK := valueAt(longitude, latitude+1e-3)
	latMinus, latMinusOK := valueAt(longitude, latitude-1e-3)
	if !centerOK || !lonPlusOK || !lonMinusOK || !latPlusOK || !latMinusOK {
		return [3]float64{}, false
	}
	firstLon := (lonPlus[0] - lonMinus[0]) / 2e-3
	firstLat := (latPlus[0] - latMinus[0]) / 2e-3
	altitudeLon := (lonPlus[1] - lonMinus[1]) / 2e-3
	altitudeLat := (latPlus[1] - latMinus[1]) / 2e-3
	residual := [3]float64{center[0], center[1], firstLon*altitudeLat - firstLat*altitudeLon}
	return residual, finite(residual[2])
}

func refineOccultationRiseSetCurveSpacing(curve *OccultationRiseSetCurve, location *time.Location, cache *occultationRiseSetEvaluationCache) {
	if curve == nil {
		return
	}
	for index, segment := range curve.Segments {
		if len(segment) < 2 {
			continue
		}
		refined := make([]OccultationPathPoint, 1, len(segment))
		refined[0] = segment[0]
		for pointIndex := 1; pointIndex < len(segment); pointIndex++ {
			refined = appendRefinedOccultationRiseSetSegment(refined, segment[pointIndex-1], segment[pointIndex], curve.Phase, curve.Direction, location, cache, 0)
		}
		curve.Segments[index] = refined
	}
}

func appendRefinedOccultationRiseSetSegment(
	points []OccultationPathPoint,
	start, end OccultationPathPoint,
	phase RiseSetPhase,
	direction RiseSetDirection,
	location *time.Location,
	cache *occultationRiseSetEvaluationCache,
	depth int,
) []OccultationPathPoint {
	if occultationPathDistanceKM(start, end) <= occultationRiseSetProjectedTargetSpacingKM(start, end) || depth >= 12 {
		return append(points, end)
	}
	middle, ok := occultationRiseSetPhaseMidpoint(
		start, end, phase, direction, location, cache,
	)
	if !ok {
		return append(points, end)
	}
	points = appendRefinedOccultationRiseSetSegment(points, start, middle, phase, direction, location, cache, depth+1)
	return appendRefinedOccultationRiseSetSegment(points, middle, end, phase, direction, location, cache, depth+1)
}

func occultationRiseSetPhaseMidpoint(
	start, end OccultationPathPoint,
	phase RiseSetPhase,
	direction RiseSetDirection,
	location *time.Location,
	cache *occultationRiseSetEvaluationCache,
) (OccultationPathPoint, bool) {
	startTT, endTT := occultationTimeToTT(start.Time), occultationTimeToTT(end.Time)
	tt := (startTT + endTT) / 2
	startAngle := occultationRiseSetHorizonAngle(tt, start.Longitude, start.Latitude, cache.context)
	endAngle := occultationRiseSetHorizonAngle(tt, end.Longitude, end.Latitude, cache.context)
	seedAngle := riseSetNormalizeRadians(startAngle + math.Remainder(endAngle-startAngle, 2*math.Pi)/2)
	if middle, ok := occultationRiseSetPhasePointOnHorizon(
		tt, seedAngle, phase, direction, location, cache,
	); ok && occultationRiseSetRefinementPointIsContinuous(start, middle, end) {
		return middle, true
	}
	longitude := geodata.NormalizeLongitude(start.Longitude + math.Remainder(end.Longitude-start.Longitude, 360)/2)
	latitude := (start.Latitude + end.Latitude) / 2
	evaluation := cache.evaluation(tt)
	longitude, latitude, ok := riseSetRefineGeographicRoot(longitude, latitude, func(lon, lat float64) (float64, float64, bool) {
		first, valid := occultationRiseSetPhaseResidual(evaluation, lon, lat, phase == RiseSetPhaseGreatest)
		state := evaluation.center.stateAt(lon, lat)
		return first, state.moonAltitude, valid && state.valid
	})
	if !ok {
		return OccultationPathPoint{}, false
	}
	middle := OccultationPathPoint{Time: occultationTTToLocation(tt, location), Longitude: longitude, Latitude: latitude, MoonAltitude: evaluation.center.stateAt(longitude, latitude).moonAltitude}
	state := evaluation.center.stateAt(longitude, latitude)
	keyPoint, key, valid := evaluation.classify(longitude, latitude, phase == RiseSetPhaseGreatest, location)
	if valid {
		middle = keyPoint
	}
	if !valid || key.phase != phase || key.direction != direction || !state.valid ||
		!occultationRiseSetRefinementPointIsContinuous(start, middle, end) {
		return OccultationPathPoint{}, false
	}
	return middle, true
}

// refineOccultationRiseSetPhaseJunctionApproaches samples the last two edges
// approaching a shared start/greatest/end junction at the exact implicit
// F=0,H=0 solution. A phase curve can turn rapidly there even when the final
// endpoint is only a few kilometres away; ordinary distance-only sampling
// otherwise renders that physical bend as one visible corner.
func refineOccultationRiseSetPhaseJunctionApproaches(
	curves []OccultationRiseSetCurve,
	location *time.Location,
	cache *occultationRiseSetEvaluationCache,
) {
	for curveIndex := range curves {
		curve := &curves[curveIndex]
		for segmentIndex, segment := range curve.Segments {
			if len(segment) < 2 {
				continue
			}
			for _, atStart := range []bool{true, false} {
				endpoint := occultationRiseSetSegmentEndpoint(segment, atStart)
				if !occultationRiseSetEndpointSharesPhaseJunction(curves, curveIndex, endpoint) {
					continue
				}
				segment = refineOccultationRiseSetJunctionSegment(
					segment, atStart, curve.Phase, curve.Direction, location, cache,
				)
			}
			curve.Segments[segmentIndex] = segment
		}
	}
}

func occultationRiseSetEndpointSharesPhaseJunction(
	curves []OccultationRiseSetCurve,
	curveIndex int,
	endpoint OccultationPathPoint,
) bool {
	for otherIndex, curve := range curves {
		if otherIndex == curveIndex || curve.Phase == curves[curveIndex].Phase ||
			curve.Direction != curves[curveIndex].Direction {
			continue
		}
		for _, segment := range curve.Segments {
			if len(segment) == 0 {
				continue
			}
			for _, other := range []OccultationPathPoint{segment[0], segment[len(segment)-1]} {
				if math.Abs(occultationTimeToTT(endpoint.Time)-occultationTimeToTT(other.Time))*86400 <= 1 &&
					occultationPathDistanceKM(endpoint, other) <= 0.01 {
					return true
				}
			}
		}
	}
	return false
}

func refineOccultationRiseSetJunctionSegment(
	segment []OccultationPathPoint,
	atStart bool,
	phase RiseSetPhase,
	direction RiseSetDirection,
	location *time.Location,
	cache *occultationRiseSetEvaluationCache,
) []OccultationPathPoint {
	if len(segment) < 3 {
		return segment
	}
	firstEdge, lastEdge := 0, len(segment)-1
	if atStart {
		if lastEdge > 2 {
			lastEdge = 2
		}
	} else {
		firstEdge = lastEdge - 2
		if firstEdge < 0 {
			firstEdge = 0
		}
	}
	result := make([]OccultationPathPoint, 1, len(segment)+16)
	result[0] = segment[0]
	for edgeIndex := 0; edgeIndex < len(segment)-1; edgeIndex++ {
		if edgeIndex >= firstEdge && edgeIndex < lastEdge {
			result = appendRefinedOccultationRiseSetJunction(
				result, segment[edgeIndex], segment[edgeIndex+1], phase, direction,
				location, cache, 0,
			)
			continue
		}
		result = append(result, segment[edgeIndex+1])
	}
	return result
}

func occultationRiseSetTurnAngleDegrees(
	first, middle, last OccultationPathPoint,
) float64 {
	longitudeScale := math.Cos(middle.Latitude * rad)
	firstX := math.Remainder(first.Longitude-middle.Longitude, 360) * longitudeScale
	firstY := first.Latitude - middle.Latitude
	lastX := math.Remainder(last.Longitude-middle.Longitude, 360) * longitudeScale
	lastY := last.Latitude - middle.Latitude
	firstLength := math.Hypot(firstX, firstY)
	lastLength := math.Hypot(lastX, lastY)
	if firstLength <= 1e-12 || lastLength <= 1e-12 {
		return 180
	}
	cosine := (firstX*lastX + firstY*lastY) / (firstLength * lastLength)
	return math.Acos(math.Max(-1, math.Min(1, cosine))) / rad
}

func appendRefinedOccultationRiseSetJunction(
	points []OccultationPathPoint,
	start, end OccultationPathPoint,
	phase RiseSetPhase,
	direction RiseSetDirection,
	location *time.Location,
	cache *occultationRiseSetEvaluationCache,
	depth int,
) []OccultationPathPoint {
	if occultationPathDistanceKM(start, end) <= 0.05 || depth >= 12 {
		return append(points, end)
	}
	middle, ok := occultationRiseSetPhaseMidpoint(start, end, phase, direction, location, cache)
	if !ok {
		return append(points, end)
	}
	// A short chord can still miss the rapidly turning physical arc at the
	// contact-envelope junction. Refine its measured sagitta, not just length.
	if planetOccultationPointSegmentDistanceKM(middle, start, end) <= 0.025 {
		return append(points, end)
	}
	points = appendRefinedOccultationRiseSetJunction(
		points, start, middle, phase, direction, location, cache, depth+1,
	)
	return appendRefinedOccultationRiseSetJunction(
		points, middle, end, phase, direction, location, cache, depth+1,
	)
}

func occultationRiseSetProjectedTargetSpacingKM(start, end OccultationPathPoint) float64 {
	latitude := math.Max(math.Abs(start.Latitude), math.Abs(end.Latitude))
	projectionScale := math.Cos(math.Min(latitude, 85) * rad)
	return occultationRiseSetTargetSpacingKM * math.Max(0.1, projectionScale)
}

func occultationRiseSetRefinementPointIsContinuous(
	start, middle, end OccultationPathPoint,
) bool {
	span := occultationPathDistanceKM(start, end)
	if !finite(span) {
		return false
	}
	// Newton can converge to the sibling root when the horizon contour folds.
	// A valid midpoint must remain in the local spatial neighbourhood of both
	// endpoints; the generous floor covers coarse samples while rejecting a
	// cross-polar branch jump.
	tolerance := math.Max(500, 2.5*span)
	return occultationPathDistanceKM(start, middle) <= tolerance &&
		occultationPathDistanceKM(middle, end) <= tolerance
}

func occultationRiseSetPhasePointOnHorizon(
	tt, angle float64,
	phase RiseSetPhase,
	direction RiseSetDirection,
	location *time.Location,
	cache *occultationRiseSetEvaluationCache,
) (OccultationPathPoint, bool) {
	point, evaluation, ok := occultationRiseSetRawPhasePointOnHorizon(
		tt, angle, phase == RiseSetPhaseGreatest, location, cache,
	)
	if !ok {
		return OccultationPathPoint{}, false
	}
	classified, key, valid := evaluation.classify(
		point.Longitude, point.Latitude, phase == RiseSetPhaseGreatest, location,
	)
	if !valid || key.phase != phase || key.direction != direction {
		return OccultationPathPoint{}, false
	}
	return classified, true
}

func occultationRiseSetRawPhasePointOnHorizon(
	tt, angle float64,
	greatest bool,
	location *time.Location,
	cache *occultationRiseSetEvaluationCache,
) (OccultationPathPoint, occultationRiseSetEvaluation, bool) {
	evaluation := cache.evaluation(tt)
	valueAt := func(candidateAngle float64) (float64, float64, float64, bool) {
		longitude, latitude, horizonOK := occultationRiseSetHorizonPoint(tt, candidateAngle, cache.context)
		if !horizonOK {
			return 0, 0, 0, false
		}
		value, valueOK := occultationRiseSetPhaseResidual(evaluation, longitude, latitude, greatest)
		return value, longitude, latitude, valueOK && finite(value)
	}
	const angleStep = 1e-4
	angle = riseSetNormalizeRadians(angle)
	for iteration := 0; iteration < 24; iteration++ {
		value, _, _, ok := valueAt(angle)
		if !ok {
			return OccultationPathPoint{}, occultationRiseSetEvaluation{}, false
		}
		if math.Abs(value) <= 1e-10 {
			break
		}
		before, _, _, beforeOK := valueAt(angle - angleStep)
		after, _, _, afterOK := valueAt(angle + angleStep)
		if !beforeOK || !afterOK {
			return OccultationPathPoint{}, occultationRiseSetEvaluation{}, false
		}
		derivative := (after - before) / (2 * angleStep)
		if !finite(derivative) || math.Abs(derivative) < 1e-16 {
			return OccultationPathPoint{}, occultationRiseSetEvaluation{}, false
		}
		delta := -value / derivative
		if math.Abs(delta) > 0.25 {
			delta = math.Copysign(0.25, delta)
		}
		angle = riseSetNormalizeRadians(angle + delta)
	}
	value, longitude, latitude, ok := valueAt(angle)
	if !ok || math.Abs(value) > 1e-7 {
		return OccultationPathPoint{}, occultationRiseSetEvaluation{}, false
	}
	state := evaluation.center.stateAt(longitude, latitude)
	if !state.valid {
		return OccultationPathPoint{}, occultationRiseSetEvaluation{}, false
	}
	return OccultationPathPoint{
		Time: occultationTTToLocation(tt, location), Longitude: longitude,
		Latitude: latitude, MoonAltitude: state.moonAltitude,
	}, evaluation, true
}
