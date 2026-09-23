package basic

import "math"

func uniqueSolarEclipsePathTimes(times []float64) []float64 {
	return movingDiskUniqueTimes(times)
}

func (solver solarEclipseSolver) refineCentralPathSpacing(points []SolarEclipsePathPoint, targetSpacingKM float64) []SolarEclipsePathPoint {
	if len(points) < 2 || targetSpacingKM <= 0 {
		return points
	}

	refined := make([]SolarEclipsePathPoint, 0, len(points))
	refined = append(refined, points[0])
	for i := 1; i < len(points); i++ {
		refined = solver.appendRefinedCentralPathSegment(refined, points[i-1], points[i], targetSpacingKM, 0)
	}
	return refined
}

func (solver solarEclipseSolver) refineCentralPathLimitSpacing(
	northern, southern []SolarEclipsePathPoint,
	centralBeginJDE, centralEndJDE, targetSpacingKM float64,
) ([]SolarEclipsePathPoint, []SolarEclipsePathPoint) {
	if len(northern) < 2 || len(northern) != len(southern) || targetSpacingKM <= 0 {
		return northern, southern
	}
	refinedNorth := make([]SolarEclipsePathPoint, 0, len(northern))
	refinedSouth := make([]SolarEclipsePathPoint, 0, len(southern))
	refinedNorth = append(refinedNorth, northern[0])
	refinedSouth = append(refinedSouth, southern[0])
	for index := 1; index < len(northern); index++ {
		refinedNorth, refinedSouth = solver.appendRefinedCentralPathLimitSegment(
			refinedNorth, refinedSouth,
			northern[index-1], southern[index-1],
			northern[index], southern[index],
			centralBeginJDE, centralEndJDE, targetSpacingKM, 0,
		)
	}
	return normalizeSolarEclipseCentralLimitPairs(refinedNorth, refinedSouth)
}

// normalizeSolarEclipsePathPointSeries removes numerical duplicate samples
// introduced by midpoint refinement. A public path is a time-ordered series;
// retaining a rounded duplicate makes GeoJSON/SVG consumers either reject the
// line or render a zero-length kink.
func normalizeSolarEclipsePathPointSeries(points []SolarEclipsePathPoint) []SolarEclipsePathPoint {
	if len(points) < 2 {
		return points
	}
	result := make([]SolarEclipsePathPoint, 0, len(points))
	for _, point := range points {
		if !finite(point.JDE) || !finite(point.Longitude) || !finite(point.Latitude) {
			continue
		}
		if len(result) == 0 {
			result = append(result, point)
			continue
		}
		last := result[len(result)-1]
		if point.JDE <= last.JDE+solarEclipsePathDuplicateTimeDays {
			if solarEclipsePathDistanceKM(last, point) <= 0.01 {
				// Keep the later evaluation so its derived altitude/width is the
				// one exposed at the surviving timestamp.
				result[len(result)-1] = point
				continue
			}
			// A branch change cannot be represented by a single public series;
			// discard the numerically ambiguous sample rather than emitting a
			// non-monotone line.
			continue
		}
		result = append(result, point)
	}
	return result
}

func normalizeSolarEclipseCentralLimitPairs(
	northern, southern []SolarEclipsePathPoint,
) ([]SolarEclipsePathPoint, []SolarEclipsePathPoint) {
	if len(northern) != len(southern) || len(northern) < 2 {
		return nil, nil
	}
	resultNorth := make([]SolarEclipsePathPoint, 0, len(northern))
	resultSouth := make([]SolarEclipsePathPoint, 0, len(southern))
	for index := range northern {
		north, south := northern[index], southern[index]
		if !finite(north.JDE) || !finite(south.JDE) ||
			!finite(north.Longitude) || !finite(north.Latitude) ||
			!finite(south.Longitude) || !finite(south.Latitude) {
			continue
		}
		if len(resultNorth) == 0 {
			resultNorth = append(resultNorth, north)
			resultSouth = append(resultSouth, south)
			continue
		}
		lastNorth, lastSouth := resultNorth[len(resultNorth)-1], resultSouth[len(resultSouth)-1]
		// Near a grazing polar contact, the limit solver can refine one logical
		// instant through several numerically distinct branch solutions. Their
		// JDEs differ by microseconds or milliseconds, while the longitude/latitude
		// may jump to the opposite polar chart branch. Treat that interval as one
		// sample; retaining both points creates a false edge in map geometry.
		const nearDuplicateTimeDays = 100.0 / 86400000.0
		if north.JDE <= lastNorth.JDE+nearDuplicateTimeDays ||
			south.JDE <= lastSouth.JDE+nearDuplicateTimeDays {
			if solarEclipsePathDistanceKM(lastNorth, north) <= 0.01 &&
				solarEclipsePathDistanceKM(lastSouth, south) <= 0.01 {
				resultNorth[len(resultNorth)-1] = north
				resultSouth[len(resultSouth)-1] = south
			}
			continue
		}
		resultNorth = append(resultNorth, north)
		resultSouth = append(resultSouth, south)
	}
	if len(resultNorth) < 2 || len(resultNorth) != len(resultSouth) {
		return nil, nil
	}
	return resultNorth, resultSouth
}

func (solver solarEclipseSolver) appendRefinedCentralPathLimitSegment(
	northern, southern []SolarEclipsePathPoint,
	startNorth, startSouth, endNorth, endSouth SolarEclipsePathPoint,
	centralBeginJDE, centralEndJDE, targetSpacingKM float64,
	depth int,
) ([]SolarEclipsePathPoint, []SolarEclipsePathPoint) {
	maximumDistance := math.Max(
		solarEclipsePathDistanceKM(startNorth, endNorth),
		solarEclipsePathDistanceKM(startSouth, endSouth),
	)
	middleJDE := (startNorth.JDE + endNorth.JDE) / 2
	if depth >= solarEclipsePathMaxAdaptiveDepth ||
		maximumDistance <= targetSpacingKM ||
		middleJDE <= centralBeginJDE || middleJDE >= centralEndJDE {
		return append(northern, endNorth), append(southern, endSouth)
	}
	middleCenter, centerOK := solver.centralPathPointAt(middleJDE)
	if !centerOK {
		return append(northern, endNorth), append(southern, endSouth)
	}
	before, beforeOK := solver.centralPathPointAt(middleJDE - solarEclipsePathVelocityStepDays)
	after, afterOK := solver.centralPathPointAt(middleJDE + solarEclipsePathVelocityStepDays)
	if !beforeOK {
		before = middleCenter
	}
	if !afterOK {
		after = middleCenter
	}
	first, second, limitsOK := solver.centralPathLimitsAtAlong(middleCenter, before, after)
	if !limitsOK {
		first, second, limitsOK = solver.centralPathLimitsAt(middleCenter)
	}
	if !limitsOK {
		return append(northern, endNorth), append(southern, endSouth)
	}
	keepDistance := solarEclipsePathDistanceKM(startNorth, first) +
		solarEclipsePathDistanceKM(startSouth, second) +
		solarEclipsePathDistanceKM(first, endNorth) +
		solarEclipsePathDistanceKM(second, endSouth)
	swapDistance := solarEclipsePathDistanceKM(startNorth, second) +
		solarEclipsePathDistanceKM(startSouth, first) +
		solarEclipsePathDistanceKM(second, endNorth) +
		solarEclipsePathDistanceKM(first, endSouth)
	if swapDistance < keepDistance {
		first, second = second, first
	}
	northern, southern = solver.appendRefinedCentralPathLimitSegment(
		northern, southern, startNorth, startSouth, first, second,
		centralBeginJDE, centralEndJDE, targetSpacingKM, depth+1,
	)
	return solver.appendRefinedCentralPathLimitSegment(
		northern, southern, first, second, endNorth, endSouth,
		centralBeginJDE, centralEndJDE, targetSpacingKM, depth+1,
	)
}

func (solver solarEclipseSolver) appendRefinedCentralPathSegment(
	points []SolarEclipsePathPoint,
	start, end SolarEclipsePathPoint,
	targetSpacingKM float64,
	depth int,
) []SolarEclipsePathPoint {
	if depth >= solarEclipsePathMaxAdaptiveDepth {
		return append(points, end)
	}

	segmentDistanceKM := solarEclipsePathDistanceKM(start, end)
	midJDE := (start.JDE + end.JDE) / 2
	mid, ok := solver.centralPathPointAt(midJDE)
	if !ok {
		return append(points, end)
	}
	// The renderer joins samples with the shorter great-circle arc.  Compare
	// the physical midpoint with that arc's midpoint so a short but sharply
	// turning segment is refined even when its endpoints are close together.
	geodesicMiddle := solarEclipsePathSphericalInterpolate(start, end, 0.5)
	curvatureErrorKM := solarEclipsePathDistanceKM(mid, geodesicMiddle)
	curvatureToleranceKM := math.Max(
		solarEclipsePathMinimumCurvatureKM,
		targetSpacingKM*solarEclipsePathAdaptiveCurvatureFraction,
	)
	if segmentDistanceKM <= targetSpacingKM && curvatureErrorKM <= curvatureToleranceKM {
		return append(points, end)
	}

	points = solver.appendRefinedCentralPathSegment(points, start, mid, targetSpacingKM, depth+1)
	return solver.appendRefinedCentralPathSegment(points, mid, end, targetSpacingKM, depth+1)
}

func (solver solarEclipseSolver) centralPathPointAt(jde float64) (SolarEclipsePathPoint, bool) {
	moon := solver.besselMoonAt(jde)
	axis := solver.besselAxisAt(jde)
	intersection := solarEclipseLineEar2(
		moon[0],
		moon[1],
		2,
		moon[0],
		moon[1],
		0,
		solarEclipseEarthPolarRatio,
		1,
		axis,
	)
	if !intersection.valid {
		return SolarEclipsePathPoint{}, false
	}

	longitude, latitude := solarEclipseIntersectionGeodetic(intersection, axis)
	sunAltitudeRad := solarEclipseSunAltitudeAtGreatest(jde, longitude, latitude, axis.gst)
	radii := solver.shadowRadiiAt(moon[2] - intersection.r2)

	widthKM := 0.0
	if math.Abs(math.Sin(sunAltitudeRad)) > 1e-12 {
		widthKM = math.Abs(2*radii.umbraRadius*solarEclipseEarthEquatorialRadiusKM) / math.Abs(math.Sin(sunAltitudeRad))
	}

	return SolarEclipsePathPoint{
		JDE:         jde,
		Longitude:   longitude,
		Latitude:    latitude,
		SunAltitude: sunAltitudeRad / rad,
		WidthKM:     widthKM,
	}, true
}

func (solver solarEclipseSolver) centralPathLimits(centerLine []SolarEclipsePathPoint) ([]SolarEclipsePathPoint, []SolarEclipsePathPoint) {
	northern, southern, paired := solver.centralPathLimitPairs(centerLine)
	return solarEclipseFilterCentralPathLimits(northern, southern, paired)
}

func solarEclipseFilterCentralPathLimits(
	northern, southern []SolarEclipsePathPoint,
	paired []bool,
) ([]SolarEclipsePathPoint, []SolarEclipsePathPoint) {
	filteredNorth := make([]SolarEclipsePathPoint, 0, len(northern))
	filteredSouth := make([]SolarEclipsePathPoint, 0, len(southern))
	for index := range northern {
		if !paired[index] {
			continue
		}
		filteredNorth = append(filteredNorth, northern[index])
		filteredSouth = append(filteredSouth, southern[index])
	}
	// Limits are optional derived lines.  A grazing or very narrow path may
	// yield one numerically valid cross-section even when its center line is
	// usable; exposing that singleton would make GeoJSON consumers reject the
	// otherwise valid event as a line geometry.
	if len(filteredNorth) < 2 || len(filteredNorth) != len(filteredSouth) {
		return nil, nil
	}
	return filteredNorth, filteredSouth
}

// centralPathLimitPairs solves the paired cross-section at every centerline
// sample and keeps the slice aligned with the center line: paired[index] is
// false where no stable pair exists, which is exactly where the analytic
// 2r/sin(altitude) width diverges and the width contract asks for 0.
func (solver solarEclipseSolver) centralPathLimitPairs(
	centerLine []SolarEclipsePathPoint,
) ([]SolarEclipsePathPoint, []SolarEclipsePathPoint, []bool) {
	northern := make([]SolarEclipsePathPoint, len(centerLine))
	southern := make([]SolarEclipsePathPoint, len(centerLine))
	paired := make([]bool, len(centerLine))
	var previousNorth, previousSouth SolarEclipsePathPoint
	havePrevious := false
	for index, center := range centerLine {
		beforeIndex, afterIndex := index-1, index+1
		if beforeIndex < 0 {
			beforeIndex = 0
		}
		if afterIndex >= len(centerLine) {
			afterIndex = len(centerLine) - 1
		}
		first, second, ok := solver.centralPathLimitsAtAlong(
			center,
			centerLine[beforeIndex],
			centerLine[afterIndex],
		)
		if !ok {
			first, second, ok = solver.centralPathLimitsAt(center)
		}
		if !ok {
			continue
		}
		north, south := first, second
		if !havePrevious {
			if second.Latitude > first.Latitude {
				north, south = second, first
			}
		} else {
			keepDistance := solarEclipsePathDistanceKM(first, previousNorth) + solarEclipsePathDistanceKM(second, previousSouth)
			swapDistance := solarEclipsePathDistanceKM(second, previousNorth) + solarEclipsePathDistanceKM(first, previousSouth)
			if swapDistance < keepDistance {
				north, south = second, first
			}
		}
		northern[index] = north
		southern[index] = south
		paired[index] = true
		previousNorth, previousSouth = north, south
		havePrevious = true
	}
	return northern, southern, paired
}

func (solver solarEclipseSolver) centralPathLimitsAtAlong(
	center, before, after SolarEclipsePathPoint,
) (SolarEclipsePathPoint, SolarEclipsePathPoint, bool) {
	footprint := solver.shadowFootprintAtWithSpacing(
		center.JDE,
		solarEclipseCentralBandBoundaryPoints,
		solarEclipseCentralShadow,
		0,
	)
	if len(footprint.Boundaries) == 0 {
		return SolarEclipsePathPoint{}, SolarEclipsePathPoint{}, false
	}
	latitude := center.Latitude * rad
	tangentX := math.Remainder(after.Longitude-before.Longitude, 360) * math.Cos(latitude)
	tangentY := after.Latitude - before.Latitude
	tangentLength := math.Hypot(tangentX, tangentY)
	if tangentLength <= 1e-12 {
		return SolarEclipsePathPoint{}, SolarEclipsePathPoint{}, false
	}
	normalX, normalY := -tangentY/tangentLength, tangentX/tangentLength
	axis := solver.besselAxisAt(center.JDE)
	var candidates []SolarEclipsePathPoint
	appendCrossing := func(first, second SolarEclipsePathPoint) {
		firstResidual := solarEclipseAlongTrackResidual(first, center, tangentX, tangentY)
		secondResidual := solarEclipseAlongTrackResidual(second, center, tangentX, tangentY)
		if !finite(firstResidual) || !finite(secondResidual) {
			return
		}
		if math.Abs(firstResidual) <= 1e-10 {
			candidates = append(candidates, first)
		}
		if firstResidual*secondResidual > 0 || math.Abs(secondResidual-firstResidual) <= 1e-12 {
			return
		}
		fraction := -firstResidual / (secondResidual - firstResidual)
		if fraction <= 0 || fraction >= 1 {
			return
		}
		longitude := normalizeLongitude(first.Longitude + math.Remainder(second.Longitude-first.Longitude, 360)*fraction)
		latitudeValue := first.Latitude + (second.Latitude-first.Latitude)*fraction
		candidates = append(candidates, SolarEclipsePathPoint{
			JDE:         center.JDE,
			Longitude:   longitude,
			Latitude:    latitudeValue,
			SunAltitude: solarEclipseSunAltitudeAtGreatest(center.JDE, longitude, latitudeValue, axis.gst) / rad,
			WidthKM:     center.WidthKM,
		})
	}
	for _, segment := range footprint.Boundaries {
		for index := 1; index < len(segment); index++ {
			appendCrossing(segment[index-1], segment[index])
		}
	}
	if footprint.Closed && len(footprint.Boundaries) == 1 && len(footprint.Boundaries[0]) > 2 {
		segment := footprint.Boundaries[0]
		appendCrossing(segment[len(segment)-1], segment[0])
	}
	if len(candidates) < 2 {
		return SolarEclipsePathPoint{}, SolarEclipsePathPoint{}, false
	}
	minimum, maximum := math.Inf(1), math.Inf(-1)
	var minimumPoint, maximumPoint SolarEclipsePathPoint
	for _, point := range candidates {
		value := solarEclipseCrossTrackResidual(point, center, normalX, normalY)
		if value < minimum {
			minimum, minimumPoint = value, point
		}
		if value > maximum {
			maximum, maximumPoint = value, point
		}
	}
	if !finite(minimum) || !finite(maximum) || maximum-minimum <= 1e-8 {
		return SolarEclipsePathPoint{}, SolarEclipsePathPoint{}, false
	}
	return maximumPoint, minimumPoint, true
}

func solarEclipseAlongTrackResidual(
	point, center SolarEclipsePathPoint, tangentX, tangentY float64,
) float64 {
	latitude := center.Latitude * rad
	deltaX := math.Remainder(point.Longitude-center.Longitude, 360) * math.Cos(latitude)
	deltaY := point.Latitude - center.Latitude
	return deltaX*tangentX + deltaY*tangentY
}

func solarEclipseCrossTrackResidual(
	point, center SolarEclipsePathPoint, normalX, normalY float64,
) float64 {
	latitude := center.Latitude * rad
	deltaX := math.Remainder(point.Longitude-center.Longitude, 360) * math.Cos(latitude)
	deltaY := point.Latitude - center.Latitude
	return deltaX*normalX + deltaY*normalY
}

func (solver solarEclipseSolver) centralPathLimitsAt(center SolarEclipsePathPoint) (SolarEclipsePathPoint, SolarEclipsePathPoint, bool) {
	moon, axis, sun := solver.besselGeometryAt(center.JDE)

	vx, vy, speed := solver.besselVelocityXYAt(center.JDE)
	if speed <= 0 {
		return SolarEclipsePathPoint{}, SolarEclipsePathPoint{}, false
	}
	perpX := -vy / speed
	perpY := vx / speed
	angle := math.Atan2(perpY, perpX)
	first, okFirst := solver.shadowFootprintPointAt(
		center.JDE, moon, axis, sun, angle, solarEclipseCentralShadow,
	)
	second, okSecond := solver.shadowFootprintPointAt(
		center.JDE, moon, axis, sun, angle+math.Pi, solarEclipseCentralShadow,
	)
	if !okFirst || !okSecond {
		return SolarEclipsePathPoint{}, SolarEclipsePathPoint{}, false
	}
	first.WidthKM = center.WidthKM
	second.WidthKM = center.WidthKM
	if first.Latitude >= second.Latitude {
		return first, second, true
	}
	return second, first, true
}

func (solver solarEclipseSolver) besselVelocityXYAt(jd float64) (float64, float64, float64) {
	before := solver.besselMoonAt(jd - solarEclipsePathVelocityStepDays)
	after := solver.besselMoonAt(jd + solarEclipsePathVelocityStepDays)
	vx := (after[0] - before[0]) / (2 * solarEclipsePathVelocityStepDays)
	vy := (after[1] - before[1]) / (2 * solarEclipsePathVelocityStepDays)
	return vx, vy, math.Hypot(vx, vy)
}

func solarEclipsePathPointFromBesselXY(jde, x, y float64, axis solarEclipseAxis) (SolarEclipsePathPoint, bool) {
	longitude, latitude, ok := solarEclipseBesselXYToGeodetic(x, y, axis, true)
	if !ok {
		return SolarEclipsePathPoint{}, false
	}
	sunAltitudeRad := solarEclipseSunAltitudeAtGreatest(jde, longitude, latitude, axis.gst)
	return SolarEclipsePathPoint{
		JDE:         jde,
		Longitude:   longitude,
		Latitude:    latitude,
		SunAltitude: sunAltitudeRad / rad,
	}, true
}

func (solver solarEclipseSolver) shadowContactPair(
	greatestJDE float64,
	kind solarEclipseShadowKind,
	internal bool,
) (SolarEclipsePathPoint, SolarEclipsePathPoint, bool) {
	middleResidual, ok := solver.shadowContactResidual(greatestJDE, kind, internal)
	if !ok || middleResidual > 0 {
		return SolarEclipsePathPoint{}, SolarEclipsePathPoint{}, false
	}
	firstJDE, firstOK := solver.shadowContactRoot(greatestJDE, -1, middleResidual, kind, internal)
	lastJDE, lastOK := solver.shadowContactRoot(greatestJDE, 1, middleResidual, kind, internal)
	if !firstOK || !lastOK {
		return SolarEclipsePathPoint{}, SolarEclipsePathPoint{}, false
	}
	first, firstOK := solver.shadowContactPointAt(firstJDE, kind, internal)
	last, lastOK := solver.shadowContactPointAt(lastJDE, kind, internal)
	if !firstOK || !lastOK {
		return SolarEclipsePathPoint{}, SolarEclipsePathPoint{}, false
	}
	return first, last, true
}

func (solver solarEclipseSolver) shadowContactRoot(
	greatestJDE float64,
	direction float64,
	middleResidual float64,
	kind solarEclipseShadowKind,
	internal bool,
) (float64, bool) {
	insideJDE := greatestJDE
	insideResidual := middleResidual
	for span := solarEclipseShadowContactSearchStepDays; span <= solarEclipseShadowContactSearchSpanDays; span += solarEclipseShadowContactSearchStepDays {
		outsideJDE := greatestJDE + direction*span
		outsideResidual, ok := solver.shadowContactResidual(outsideJDE, kind, internal)
		if !ok {
			continue
		}
		if outsideResidual >= 0 {
			leftJDE, rightJDE := outsideJDE, insideJDE
			leftResidual, rightResidual := outsideResidual, insideResidual
			if leftJDE > rightJDE {
				leftJDE, rightJDE = rightJDE, leftJDE
				leftResidual, rightResidual = rightResidual, leftResidual
			}
			for rightJDE-leftJDE > solarEclipseShadowContactToleranceDays {
				middleJDE := (leftJDE + rightJDE) / 2
				residual, valid := solver.shadowContactResidual(middleJDE, kind, internal)
				if !valid {
					return 0, false
				}
				if (residual >= 0) == (leftResidual >= 0) {
					leftJDE, leftResidual = middleJDE, residual
				} else {
					rightJDE, rightResidual = middleJDE, residual
				}
			}
			return (leftJDE + rightJDE) / 2, true
		}
		insideJDE, insideResidual = outsideJDE, outsideResidual
	}
	return 0, false
}

func (solver solarEclipseSolver) shadowContactResidual(
	jde float64,
	kind solarEclipseShadowKind,
	internal bool,
) (float64, bool) {
	if kind == solarEclipseCentralShadow && solver.exactCentralContact && !internal {
		return solver.shadowContactResidualExact(jde)
	}
	moon := solver.besselMoonAt(jde)
	distanceSquared := moon[0]*moon[0] + moon[1]*moon[1]
	if distanceSquared <= 0 {
		return 0, false
	}
	radius := solver.shadowRadiusAt(moon[2], kind)
	if radius <= 0 {
		return 0, false
	}
	earthRadius := 1 - (1/solarEclipseEarthPolarRatioSquared-1)*moon[1]*moon[1]/distanceSquared/2
	limit := earthRadius + radius
	if internal {
		limit = earthRadius - radius
	}
	if limit <= 0 {
		return 0, false
	}
	return math.Sqrt(distanceSquared) - limit, true
}

func (solver solarEclipseSolver) shadowContactResidualExact(jde float64) (float64, bool) {
	_, value, ok := solver.shadowContactMaximum(jde)
	return -value, ok
}

func (solver solarEclipseSolver) shadowContactMaximum(jde float64) (float64, float64, bool) {
	const samples = 32
	step := 2 * math.Pi / samples
	bestIndex := -1
	bestValue := math.Inf(-1)
	for index := 0; index < samples; index++ {
		value, ok := solver.shadowBoundaryDiscriminant(jde, float64(index)*step)
		if !ok {
			continue
		}
		if value > bestValue {
			bestIndex, bestValue = index, value
		}
	}
	if bestIndex < 0 {
		return 0, 0, false
	}
	left := float64(bestIndex)*step - step
	right := float64(bestIndex)*step + step
	valueAt := func(angle float64) float64 {
		value, ok := solver.shadowBoundaryDiscriminant(jde, angle)
		if !ok {
			return math.Inf(-1)
		}
		return value
	}
	// Golden-section maximization removes the polar effective-radius
	// approximation from contact times while retaining a bounded cost.
	golden := (math.Sqrt(5) - 1) / 2
	x1 := right - golden*(right-left)
	x2 := left + golden*(right-left)
	f1, f2 := valueAt(x1), valueAt(x2)
	for iteration := 0; iteration < 32; iteration++ {
		if f1 < f2 {
			left, x1, f1 = x1, x2, f2
			x2 = left + golden*(right-left)
			f2 = valueAt(x2)
		} else {
			right, x2, f2 = x2, x1, f1
			x1 = right - golden*(right-left)
			f1 = valueAt(x1)
		}
	}
	bestAngle := float64(bestIndex) * step
	if f1 > bestValue {
		bestAngle, bestValue = x1, f1
	}
	if f2 > bestValue {
		bestAngle, bestValue = x2, f2
	}
	return bestAngle, bestValue, true
}

func (solver solarEclipseSolver) shadowBoundaryDiscriminant(jde, angle float64) (float64, bool) {
	moon, axis, _ := solver.besselGeometryAt(jde)
	radius := solver.shadowRadiusAt(moon[2], solarEclipseCentralShadow)
	if radius <= 0 {
		return 0, false
	}
	cosAngle, sinAngle := math.Cos(angle), math.Sin(angle)
	best := math.Inf(-1)
	for iteration := 0; iteration < solarEclipsePartialFootprintIterationLimit; iteration++ {
		x, y := moon[0]+radius*cosAngle, moon[1]+radius*sinAngle
		discriminant := solarEclipseLineEllipsoidDiscriminant(
			x, y, 2, x, y, 0, solarEclipseEarthPolarRatio, 1, axis,
		)
		best = discriminant
		if discriminant < 0 {
			return discriminant, true
		}
		intersection := solarEclipseLineEar2(
			x, y, 2, x, y, 0, solarEclipseEarthPolarRatio, 1, axis,
		)
		if !intersection.valid {
			return discriminant, true
		}
		nextRadius := solver.shadowRadiusAt(moon[2]-intersection.r2, solarEclipseCentralShadow)
		if nextRadius <= 0 {
			return discriminant, true
		}
		if math.Abs(nextRadius-radius) <= solarEclipsePartialFootprintPointTolerance {
			break
		}
		radius = nextRadius
	}
	return best, true
}

func solarEclipseLineEllipsoidDiscriminant(
	x1, y1, z1, x2, y2, z2, polarRatio, radius float64,
	axis solarEclipseAxis,
) float64 {
	cosTilt, sinTilt := math.Cos(axis.tilt), math.Sin(axis.tilt)
	x1Rot := x1
	y1Rot := cosTilt*y1 - sinTilt*z1
	z1Rot := sinTilt*y1 + cosTilt*z1
	x2Rot := x2
	y2Rot := cosTilt*y2 - sinTilt*z2
	z2Rot := sinTilt*y2 + cosTilt*z2
	dx, dy, dz := x2Rot-x1Rot, y2Rot-y1Rot, z2Rot-z1Rot
	polarRatioSquared := polarRatio * polarRatio
	a := dx*dx + dy*dy + dz*dz/polarRatioSquared
	b := x1Rot*dx + y1Rot*dy + z1Rot*dz/polarRatioSquared
	c := x1Rot*x1Rot + y1Rot*y1Rot + z1Rot*z1Rot/polarRatioSquared - radius*radius
	return b*b - a*c
}

func (solver solarEclipseSolver) shadowContactPointAt(
	jde float64,
	kind solarEclipseShadowKind,
	internal bool,
) (SolarEclipsePathPoint, bool) {
	if kind == solarEclipseCentralShadow && solver.exactCentralContact && !internal {
		angle, _, ok := solver.shadowContactMaximum(jde)
		if !ok {
			return SolarEclipsePathPoint{}, false
		}
		if point, pointOK := solver.centralShadowPointAt(jde, angle); pointOK {
			return point, true
		}
		for _, offset := range []float64{-1e-8, 1e-8, -1e-7, 1e-7} {
			offsetJDE := jde + offset
			offsetAngle, _, maximumOK := solver.shadowContactMaximum(offsetJDE)
			if !maximumOK {
				continue
			}
			if point, pointOK := solver.centralShadowPointAt(offsetJDE, offsetAngle); pointOK {
				point.JDE = jde
				return point, true
			}
		}
		return SolarEclipsePathPoint{}, false
	}
	moon := solver.besselMoonAt(jde)
	distance := math.Hypot(moon[0], moon[1])
	if distance <= 0 || solver.shadowRadiusAt(moon[2], kind) <= 0 {
		return SolarEclipsePathPoint{}, false
	}
	axis := solver.besselAxisAt(jde)
	unitX, unitY := moon[0]/distance, moon[1]/distance
	insideScale, outsideScale := 0.0, 1.1
	var intersection solarEclipseLineIntersection
	for iteration := 0; iteration < 48; iteration++ {
		scale := (insideScale + outsideScale) / 2
		candidate := solarEclipseLineEar2(
			scale*unitX, scale*unitY, 2,
			scale*unitX, scale*unitY, 0,
			solarEclipseEarthPolarRatio, 1, axis,
		)
		if candidate.valid {
			insideScale = scale
			intersection = candidate
		} else {
			outsideScale = scale
		}
	}
	if !intersection.valid {
		return SolarEclipsePathPoint{}, false
	}
	longitude, latitude := solarEclipseIntersectionGeodetic(intersection, axis)
	sunAltitudeRad := solarEclipseSunAltitudeAtGreatest(jde, longitude, latitude, axis.gst)
	return SolarEclipsePathPoint{
		JDE:         jde,
		Longitude:   longitude,
		Latitude:    latitude,
		SunAltitude: sunAltitudeRad / rad,
	}, true
}

func (solver solarEclipseSolver) shadowRadiusAt(moonBesselZ float64, kind solarEclipseShadowKind) float64 {
	radii := solver.shadowRadiiAt(moonBesselZ)
	if kind == solarEclipseCentralShadow {
		return radii.absUmbraRadius
	}
	return radii.penumbraRadius
}

func (solver solarEclipseSolver) partialFootprintAt(jde float64, boundaryPoints int) SolarEclipsePartialFootprint {
	return solver.shadowFootprintAt(jde, boundaryPoints, solarEclipsePenumbralShadow)
}

func (solver solarEclipseSolver) shadowFootprintAt(
	jde float64,
	boundaryPoints int,
	kind solarEclipseShadowKind,
) SolarEclipsePartialFootprint {
	return solver.shadowFootprintAtWithSpacing(jde, boundaryPoints, kind, 0)
}

func (solver solarEclipseSolver) shadowFootprintAtWithSpacing(
	jde float64,
	boundaryPoints int,
	kind solarEclipseShadowKind,
	targetSpacingKM float64,
) SolarEclipsePartialFootprint {
	moon, axis, sun := solver.besselGeometryAt(jde)
	return solver.shadowFootprintAtWithGeometry(
		jde, moon, axis, sun, boundaryPoints, kind, targetSpacingKM,
	)
}

// shadowFootprintAtWithGeometry 用调用方给定的贝塞尔几何求瞬时足迹。显式 ΔT 的单时刻
// 入口走这条路径：几何由调用方按自己的 ΔT 构造，避免落到进程级 ΔT 上。
// shadowFootprintAtWithGeometry solves one instantaneous footprint from precomputed
// Besselian geometry, so a caller with an explicit ΔT never falls back to the global one.
func (solver solarEclipseSolver) shadowFootprintAtWithGeometry(
	jd float64,
	moon [3]float64,
	axis solarEclipseAxis,
	sun [3]float64,
	boundaryPoints int,
	kind solarEclipseShadowKind,
	targetSpacingKM float64,
) SolarEclipsePartialFootprint {
	samples := make([]solarEclipsePartialBoundarySample, boundaryPoints)
	for i := range samples {
		angle := 2 * math.Pi * float64(i) / float64(boundaryPoints)
		point, ok := solver.shadowFootprintPointAt(jd, moon, axis, sun, angle, kind)
		samples[i] = solarEclipsePartialBoundarySample{
			point: point,
			ok:    ok,
			angle: angle,
		}
	}
	samples = solver.refineShadowFootprintTransitions(jd, moon, axis, sun, samples, kind)
	if targetSpacingKM > 0 {
		samples = solver.refineShadowFootprintSpacing(
			jd, moon, axis, sun, samples, kind, targetSpacingKM,
		)
	}

	boundaries, closed := solarEclipsePartialBoundarySegments(samples)
	footprint := SolarEclipsePartialFootprint{
		JDE:        jd,
		Boundaries: boundaries,
		Closed:     closed,
	}
	if !closed {
		// The sampled rim stops where the shooting gives up, which can be tens of
		// kilometres inside the horizon.  Recover the exact tangency points so the
		// caller can close the region with a horizon arc that really meets the
		// physical boundary; Boundaries itself stays untouched so every existing
		// consumer of the open rim keeps its geometry.
		footprint.HorizonEnds = solver.shadowFootprintHorizonEnds(
			jd, moon, axis, sun, samples, boundaries, kind,
		)
	}
	return footprint
}

// shadowFootprintHorizonEnds 求未闭合足迹边界两端在地平圈上的擦地点。
// shadowFootprintHorizonEnds returns the horizon tangency point at each end of an
// open footprint boundary.
func (solver solarEclipseSolver) shadowFootprintHorizonEnds(
	jd float64,
	moon [3]float64,
	axis solarEclipseAxis,
	sun [3]float64,
	samples []solarEclipsePartialBoundarySample,
	boundaries [][]SolarEclipsePathPoint,
	kind solarEclipseShadowKind,
) []SolarEclipsePathPoint {
	if len(samples) < 2 || len(boundaries) == 0 || len(boundaries[0]) == 0 {
		return nil
	}
	lastSegment := boundaries[len(boundaries)-1]
	if len(lastSegment) == 0 {
		return nil
	}
	// A footprint can be cut by the horizon along more than one arc (a wide
	// penumbra near the terminator has two), so the two ends of the joined
	// boundary are the first and the last ok/!ok transition rather than an
	// arbitrary pair of them.
	firstIndex, lastIndex := -1, -1
	for index := range samples {
		if samples[index].ok == samples[(index+1)%len(samples)].ok {
			continue
		}
		if firstIndex < 0 {
			firstIndex = index
		}
		lastIndex = index
	}
	if firstIndex < 0 || lastIndex == firstIndex {
		return nil
	}
	ends := make([]SolarEclipsePathPoint, 0, 2)
	for _, index := range []int{firstIndex, lastIndex} {
		sample, next := samples[index], samples[(index+1)%len(samples)]
		// The transition samples are already bisected in azimuth by
		// refineShadowFootprintTransitions, so the valid side of the pair is the
		// azimuth that carries the tangency.
		angle := sample.angle
		if !sample.ok {
			angle = next.angle
		}
		point, ok := solver.shadowFootprintHorizonEndAt(jd, moon, axis, sun, angle, kind)
		if !ok {
			return nil
		}
		ends = append(ends, point)
	}
	start := boundaries[0][0]
	if solarEclipsePathDistanceKM(ends[0], start) > solarEclipsePathDistanceKM(ends[1], start) {
		ends[0], ends[1] = ends[1], ends[0]
	}
	return ends
}

// shadowFootprintHorizonEndAt 在给定方位上求阴影锥面与地表的切点。该方位由采样
// 的 ok/!ok 转移给出，切点就是本影（或半影）边界真正终止的位置：轴线平行线与椭球
// 相切，因此当地太阳高度为 0，也就是落在该时刻的地平圈上。
// shadowFootprintHorizonEndAt solves the cone/ground tangency point on one azimuth.
// The azimuth comes from the sampled ok/!ok transition, and the tangency point is
// where the boundary really ends: the axis-parallel line grazes the ellipsoid, so
// the local solar altitude is zero and the point lies on the horizon.
func (solver solarEclipseSolver) shadowFootprintHorizonEndAt(
	jd float64,
	moon [3]float64,
	axis solarEclipseAxis,
	sun [3]float64,
	angle float64,
	kind solarEclipseShadowKind,
) (SolarEclipsePathPoint, bool) {
	radius := solver.shadowRadiusAt(moon[2], kind)
	if radius <= 0 {
		return SolarEclipsePathPoint{}, false
	}
	cosAngle, sinAngle := math.Cos(angle), math.Sin(angle)
	discriminantAt := func(value float64) float64 {
		x := moon[0] + value*cosAngle
		y := moon[1] + value*sinAngle
		return solarEclipseLineEllipsoidDiscriminant(
			x, y, 2, x, y, 0, solarEclipseEarthPolarRatio, 1, axis,
		)
	}
	if discriminantAt(radius) < 0 {
		return SolarEclipsePathPoint{}, false
	}
	// The transition azimuth is where the sampled radius meets an edge of the
	// silhouette, and that edge can be either the inner or the outer one: a shadow
	// whose axis still crosses the Earth is cut outwards, one whose axis misses the
	// Earth is cut inwards, and a large penumbral circle can meet either.  Search
	// both directions and keep the nearer sign change, which is the edge the
	// sampled radius actually ran into.
	valid, invalid := radius, 0.0
	found := false
	nearest := math.Inf(1)
	for offset := solarEclipseHorizonEndSearchStep; offset <= solarEclipseHorizonEndSearchLimit; offset *= 2 {
		if discriminantAt(radius+offset) < 0 {
			nearest, invalid, found = offset, radius+offset, true
			break
		}
	}
	for offset := solarEclipseHorizonEndSearchStep; offset <= radius; offset *= 2 {
		candidate := radius - offset
		if candidate <= 0 {
			break
		}
		if discriminantAt(candidate) < 0 {
			if offset < nearest {
				invalid, found = candidate, true
			}
			break
		}
	}
	if !found {
		return SolarEclipsePathPoint{}, false
	}
	// valid stays on the disc >= 0 side and invalid on the other one; the bracket
	// order differs between the outward and the inward cut, so the two ends are
	// never sorted against each other.
	inside, outside := valid, invalid
	for round := 0; round < solarEclipseHorizonEndBisectionRounds; round++ {
		middle := (inside + outside) / 2
		if discriminantAt(middle) >= 0 {
			inside = middle
			continue
		}
		outside = middle
	}
	tangent := inside
	intersection := solarEclipseLineEar2(
		moon[0]+tangent*cosAngle, moon[1]+tangent*sinAngle, 2,
		moon[0]+tangent*cosAngle, moon[1]+tangent*sinAngle, 0,
		solarEclipseEarthPolarRatio, 1, axis,
	)
	if !intersection.valid {
		return SolarEclipsePathPoint{}, false
	}
	longitude, latitude := solarEclipseIntersectionGeodetic(intersection, axis)
	sunAltitudeRad := solarEclipseSunAltitudeFromEquatorial(sun, longitude, latitude, axis.gst)
	return SolarEclipsePathPoint{
		JDE:         jd,
		Longitude:   longitude,
		Latitude:    latitude,
		SunAltitude: sunAltitudeRad / rad,
	}, true
}

func (solver solarEclipseSolver) refineShadowFootprintSpacing(
	jd float64,
	moon [3]float64,
	axis solarEclipseAxis,
	sun [3]float64,
	samples []solarEclipsePartialBoundarySample,
	kind solarEclipseShadowKind,
	targetSpacingKM float64,
) []solarEclipsePartialBoundarySample {
	if len(samples) < 2 || targetSpacingKM <= 0 {
		return samples
	}
	result := make([]solarEclipsePartialBoundarySample, 0, len(samples))
	for index, left := range samples {
		right := samples[(index+1)%len(samples)]
		if index == len(samples)-1 {
			right.angle += 2 * math.Pi
		} else if right.angle < left.angle {
			right.angle += 2 * math.Pi
		}
		result = append(result, left)
		result = solver.appendRefinedShadowFootprintInterval(
			result, jd, moon, axis, sun, left, right, kind, targetSpacingKM, 0,
		)
	}
	return result
}

func (solver solarEclipseSolver) appendRefinedShadowFootprintInterval(
	result []solarEclipsePartialBoundarySample,
	jd float64,
	moon [3]float64,
	axis solarEclipseAxis,
	sun [3]float64,
	left, right solarEclipsePartialBoundarySample,
	kind solarEclipseShadowKind,
	targetSpacingKM float64,
	depth int,
) []solarEclipsePartialBoundarySample {
	if depth >= solarEclipseShadowFootprintAdaptiveMaxDepth || !left.ok || !right.ok ||
		solarEclipsePathDistanceKM(left.point, right.point) <= targetSpacingKM {
		return result
	}
	angle := (left.angle + right.angle) / 2
	point, ok := solver.shadowFootprintPointAt(
		jd, moon, axis, sun, math.Mod(angle, 2*math.Pi), kind,
	)
	middle := solarEclipsePartialBoundarySample{point: point, ok: ok, angle: angle}
	if !middle.ok {
		return result
	}
	result = solver.appendRefinedShadowFootprintInterval(
		result, jd, moon, axis, sun, left, middle, kind, targetSpacingKM, depth+1,
	)
	result = append(result, middle)
	return solver.appendRefinedShadowFootprintInterval(
		result, jd, moon, axis, sun, middle, right, kind, targetSpacingKM, depth+1,
	)
}

type solarEclipsePartialBoundarySample struct {
	point SolarEclipsePathPoint
	ok    bool
	angle float64
}

func (solver solarEclipseSolver) refineShadowFootprintTransitions(
	jd float64,
	moon [3]float64,
	axis solarEclipseAxis,
	sun [3]float64,
	samples []solarEclipsePartialBoundarySample,
	kind solarEclipseShadowKind,
) []solarEclipsePartialBoundarySample {
	if len(samples) < 2 {
		return samples
	}
	result := make([]solarEclipsePartialBoundarySample, 0, len(samples)+4)
	for index, sample := range samples {
		result = append(result, sample)
		next := samples[(index+1)%len(samples)]
		if sample.ok == next.ok {
			continue
		}
		nextAngle := next.angle
		if index == len(samples)-1 {
			nextAngle += 2 * math.Pi
		}
		refined := solver.refineShadowFootprintTransition(jd, moon, axis, sun, sample, next, nextAngle, kind)
		result = append(result, refined)
	}
	return result
}

func (solver solarEclipseSolver) refineShadowFootprintTransition(
	jd float64,
	moon [3]float64,
	axis solarEclipseAxis,
	sun [3]float64,
	first, second solarEclipsePartialBoundarySample,
	secondAngle float64,
	kind solarEclipseShadowKind,
) solarEclipsePartialBoundarySample {
	leftAngle := first.angle
	rightAngle := secondAngle
	leftOK := first.ok
	best := first
	if second.ok {
		best = second
		best.angle = secondAngle
	}
	for iteration := 0; iteration < solarEclipsePartialFootprintTransitionIterations; iteration++ {
		middleAngle := (leftAngle + rightAngle) / 2
		evaluationAngle := math.Mod(middleAngle, 2*math.Pi)
		point, ok := solver.shadowFootprintPointAt(jd, moon, axis, sun, evaluationAngle, kind)
		middle := solarEclipsePartialBoundarySample{point: point, ok: ok, angle: middleAngle}
		if ok {
			best = middle
		}
		if ok == leftOK {
			leftAngle = middleAngle
		} else {
			rightAngle = middleAngle
		}
		if rightAngle-leftAngle <= solarEclipsePartialFootprintPointTolerance {
			break
		}
	}
	best.angle = math.Mod(best.angle, 2*math.Pi)
	return best
}

func (solver solarEclipseSolver) shadowFootprintPointAt(
	jd float64,
	moon [3]float64,
	axis solarEclipseAxis,
	sun [3]float64,
	angle float64,
	kind solarEclipseShadowKind,
) (SolarEclipsePathPoint, bool) {
	cosAngle := math.Cos(angle)
	sinAngle := math.Sin(angle)
	radius := solver.shadowRadiusAt(moon[2], kind)
	if radius <= 0 {
		return SolarEclipsePathPoint{}, false
	}

	var intersection solarEclipseLineIntersection
	for i := 0; i < solarEclipsePartialFootprintIterationLimit; i++ {
		x := moon[0] + radius*cosAngle
		y := moon[1] + radius*sinAngle
		intersection = solarEclipseLineEar2(
			x,
			y,
			2,
			x,
			y,
			0,
			solarEclipseEarthPolarRatio,
			1,
			axis,
		)
		if !intersection.valid {
			return SolarEclipsePathPoint{}, false
		}

		nextRadius := solver.shadowRadiusAt(moon[2]-intersection.r2, kind)
		if nextRadius <= 0 {
			return SolarEclipsePathPoint{}, false
		}
		if math.Abs(nextRadius-radius) <= solarEclipsePartialFootprintPointTolerance {
			radius = nextRadius
			break
		}
		radius = nextRadius
	}

	x := moon[0] + radius*cosAngle
	y := moon[1] + radius*sinAngle
	intersection = solarEclipseLineEar2(
		x,
		y,
		2,
		x,
		y,
		0,
		solarEclipseEarthPolarRatio,
		1,
		axis,
	)
	if !intersection.valid {
		return SolarEclipsePathPoint{}, false
	}

	longitude, latitude := solarEclipseIntersectionGeodetic(intersection, axis)
	sunAltitudeRad := solarEclipseSunAltitudeFromEquatorial(sun, longitude, latitude, axis.gst)
	return SolarEclipsePathPoint{
		JDE:         jd,
		Longitude:   longitude,
		Latitude:    latitude,
		SunAltitude: sunAltitudeRad / rad,
	}, true
}

func solarEclipsePartialBoundarySegments(samples []solarEclipsePartialBoundarySample) ([][]SolarEclipsePathPoint, bool) {
	segments := make([][]SolarEclipsePathPoint, 0, 2)
	var current []SolarEclipsePathPoint
	allSamplesValid := len(samples) > 0
	for _, sample := range samples {
		if !sample.ok {
			allSamplesValid = false
			segments = appendSolarEclipsePartialSegment(segments, current)
			current = nil
			continue
		}
		if len(current) > 0 && solarEclipsePathCrossesAntimeridian(current[len(current)-1], sample.point) {
			firstCrossing, secondCrossing := solarEclipsePathAntimeridianCrossings(
				current[len(current)-1], sample.point,
			)
			current = append(current, firstCrossing)
			segments = appendSolarEclipsePartialSegment(segments, current)
			current = []SolarEclipsePathPoint{secondCrossing}
		}
		current = append(current, sample.point)
	}
	segments = appendSolarEclipsePartialSegment(segments, current)
	segments = mergeSolarEclipsePartialWrapSegment(segments, samples)

	if !allSamplesValid {
		return segments, false
	}
	totalPoints := 0
	for _, segment := range segments {
		totalPoints += len(segment)
	}
	if totalPoints < 3 {
		return segments, false
	}
	if len(segments) == 1 && !solarEclipsePathCrossesAntimeridian(segments[0][len(segments[0])-1], segments[0][0]) {
		segments[0] = append(segments[0], segments[0][0])
	}
	return segments, true
}

func appendSolarEclipsePartialSegment(
	segments [][]SolarEclipsePathPoint,
	segment []SolarEclipsePathPoint,
) [][]SolarEclipsePathPoint {
	if len(segment) == 0 {
		return segments
	}
	return append(segments, segment)
}

func mergeSolarEclipsePartialWrapSegment(
	segments [][]SolarEclipsePathPoint,
	samples []solarEclipsePartialBoundarySample,
) [][]SolarEclipsePathPoint {
	if len(segments) < 2 || len(samples) == 0 || !samples[0].ok || !samples[len(samples)-1].ok {
		return segments
	}
	first := segments[0]
	last := segments[len(segments)-1]
	if solarEclipsePathCrossesAntimeridian(last[len(last)-1], first[0]) {
		firstCrossing, secondCrossing := solarEclipsePathAntimeridianCrossings(
			last[len(last)-1], first[0],
		)
		segments[len(segments)-1] = append(last, firstCrossing)
		segments[0] = append([]SolarEclipsePathPoint{secondCrossing}, first...)
		return segments
	}

	merged := make([]SolarEclipsePathPoint, 0, len(last)+len(first))
	merged = append(merged, last...)
	merged = append(merged, first...)
	result := make([][]SolarEclipsePathPoint, 0, len(segments)-1)
	result = append(result, merged)
	result = append(result, segments[1:len(segments)-1]...)
	return result
}

func solarEclipsePathCrossesAntimeridian(a, b SolarEclipsePathPoint) bool {
	return math.Abs(a.Longitude-b.Longitude) > 180
}

func solarEclipsePathAntimeridianCrossings(
	first, second SolarEclipsePathPoint,
) (SolarEclipsePathPoint, SolarEclipsePathPoint) {
	boundary := 180.0
	secondLongitude := second.Longitude
	if first.Longitude < 0 {
		boundary = -180
		secondLongitude -= 360
	} else {
		secondLongitude += 360
	}
	left, right := 0.0, 1.0
	for iteration := 0; iteration < 64; iteration++ {
		middle := (left + right) / 2
		point := solarEclipsePathSphericalInterpolate(first, second, middle)
		middleLongitude := point.Longitude
		for middleLongitude-first.Longitude > 180 {
			middleLongitude -= 360
		}
		for middleLongitude-first.Longitude < -180 {
			middleLongitude += 360
		}
		if (first.Longitude-boundary)*(middleLongitude-boundary) <= 0 {
			right = middle
		} else {
			left = middle
		}
		if math.Abs(middleLongitude-boundary) <= 1e-12 || right-left <= 1e-13 {
			break
		}
	}
	fraction := (left + right) / 2
	if secondLongitude != first.Longitude {
		linearFraction := (boundary - first.Longitude) / (secondLongitude - first.Longitude)
		if linearFraction >= 0 && linearFraction <= 1 && math.Abs(right-left) > 1e-6 {
			fraction = linearFraction
		}
	}
	crossing := solarEclipsePathSphericalInterpolate(first, second, fraction)
	crossing.Longitude = boundary
	opposite := crossing
	opposite.Longitude = -boundary
	return crossing, opposite
}

func solarEclipsePathSphericalInterpolate(
	first, second SolarEclipsePathPoint,
	fraction float64,
) SolarEclipsePathPoint {
	firstVector := solarEclipseLLRToXYZ(first.Longitude*rad, first.Latitude*rad, 1)
	secondVector := solarEclipseLLRToXYZ(second.Longitude*rad, second.Latitude*rad, 1)
	dot := firstVector[0]*secondVector[0] + firstVector[1]*secondVector[1] + firstVector[2]*secondVector[2]
	dot = math.Max(-1, math.Min(1, dot))
	firstWeight, secondWeight := 1-fraction, fraction
	if dot < 1-1e-14 && dot > -1+1e-14 {
		angle := math.Acos(dot)
		sine := math.Sin(angle)
		firstWeight = math.Sin((1-fraction)*angle) / sine
		secondWeight = math.Sin(fraction*angle) / sine
	}
	vector := [3]float64{
		firstWeight*firstVector[0] + secondWeight*secondVector[0],
		firstWeight*firstVector[1] + secondWeight*secondVector[1],
		firstWeight*firstVector[2] + secondWeight*secondVector[2],
	}
	coordinates := solarEclipseXYZToLLR(vector[0], vector[1], vector[2])
	longitude := solarEclipseNormalizeSignedRadians(coordinates[0]) / rad
	return SolarEclipsePathPoint{
		JDE:         first.JDE + fraction*(second.JDE-first.JDE),
		Longitude:   longitude,
		Latitude:    coordinates[1] / rad,
		SunAltitude: first.SunAltitude + fraction*(second.SunAltitude-first.SunAltitude),
		WidthKM:     first.WidthKM + fraction*(second.WidthKM-first.WidthKM),
	}
}

func solarEclipsePathDistanceKM(a, b SolarEclipsePathPoint) float64 {
	lat1 := a.Latitude * rad
	lat2 := b.Latitude * rad
	dlat := lat2 - lat1
	dlon := solarEclipseNormalizeSignedRadians((b.Longitude - a.Longitude) * rad)
	h := math.Sin(dlat/2)*math.Sin(dlat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(dlon/2)*math.Sin(dlon/2)
	if h > 1 {
		h = 1
	}
	return 2 * solarEclipseEarthEquatorialRadiusKM * math.Asin(math.Sqrt(h))
}
