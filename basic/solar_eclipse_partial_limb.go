package basic

import "math"

// solarEclipsePartialLimbHorizonDeg 以下的零食分包络点视为地平线端点，改用日出日落约束求解，与升落曲线在同一点相接。
const solarEclipsePartialLimbHorizonDeg = 0.05

// limbCorrectPartialBoundary 把偏食可见域外边界（分包络与初亏、复圆的升落曲线）平移到月缘外切界线，食甚曲线保持原样。
func (solver solarEclipseSolver) limbCorrectPartialBoundary(
	result *SolarEclipsePartialFootprintsResult,
	limb LunarLimb,
) bool {
	if limb == nil {
		return false
	}
	applied := false
	contours := make([][]SolarEclipsePathPoint, len(result.PartialBandContours))
	for index, contour := range result.PartialBandContours {
		corrected, changed := solver.limbCorrectPartialPoints(contour, limb, false)
		contours[index] = corrected
		applied = applied || changed
	}
	result.PartialBandContours = contours

	curves := make([]SolarEclipseRiseSetCurve, len(result.RiseSetCurves))
	for index, curve := range result.RiseSetCurves {
		curves[index] = curve
		if curve.Phase == RiseSetPhaseGreatest {
			continue
		}
		curves[index].Segments = make([][]SolarEclipsePathPoint, len(curve.Segments))
		for segmentIndex, segment := range curve.Segments {
			corrected, changed := solver.limbCorrectPartialPoints(segment, limb, true)
			curves[index].Segments[segmentIndex] = corrected
			applied = applied || changed
		}
	}
	result.RiseSetCurves = curves
	return applied
}

// limbCorrectPartialPoints 逐点平移；求解失败的点保留圆月面位置。horizon 为真表示整条线都在地平线上。
func (solver solarEclipseSolver) limbCorrectPartialPoints(
	points []SolarEclipsePathPoint,
	limb LunarLimb,
	horizon bool,
) ([]SolarEclipsePathPoint, bool) {
	corrected := append([]SolarEclipsePathPoint(nil), points...)
	applied := false
	for index, point := range points {
		if !finite(point.JDE) || point.JDE <= 0 {
			continue
		}
		onHorizon := horizon || math.Abs(point.SunAltitude) < solarEclipsePartialLimbHorizonDeg
		// 介于端点与 0.5° 之间的低空点地面梯度过小，平移不可信，保留圆月面点。
		if !onHorizon && math.Abs(point.SunAltitude) < limbLimitHorizonDeg {
			continue
		}
		context := solver.localStateContextAt(point.JDE)
		evaluate, ok := solver.partialLimbGap(context, point.JDE, limb)
		if !ok {
			continue
		}
		var longitude, latitude float64
		if onHorizon {
			longitude, latitude, ok = partialLimbHorizonSolve(point.Longitude, point.Latitude, evaluate, context)
		} else {
			longitude, latitude, ok = limbLimitSolve(point.Longitude, point.Latitude, evaluate)
		}
		if !ok {
			continue
		}
		if !limbGroundMoved(point.Longitude, point.Latitude, longitude, latitude) {
			continue
		}
		corrected[index].Longitude, corrected[index].Latitude = longitude, latitude
		corrected[index].SunAltitude = context.stateAt(longitude*rad, latitude*rad, 0).sunAltitudeRad / rad
		applied = true
	}
	return corrected, applied
}

// partialLimbGap 返回某时刻站点处的圆月面外切间隙与月缘要求的间隙（弧度），月缘间隙与站心同一判据。
func (solver solarEclipseSolver) partialLimbGap(
	context localSolarEclipseStateContext,
	jde float64,
	limb LunarLimb,
) (limbLimitGap, bool) {
	limbContext := solver.withLimbParams(limb).localStateContextAt(jde)
	if limbContext.limbFrame == nil {
		return nil, false
	}
	return func(longitude, latitude float64) (float64, float64, bool) {
		lonRad, latRad := longitude*rad, latitude*rad
		state := context.stateAt(lonRad, latRad, 0)
		limbState := limbContext.stateAt(lonRad, latRad, 0)
		if !limbState.outerLimbApplied {
			return 0, 0, false
		}
		gap := state.movingDiskContactState().externalContactGap()
		return gap, gap - limbState.movingDiskContactState().externalContactGap(), true
	}, true
}

// partialLimbHorizonSolve 在固定时刻同时满足月缘外切与太阳高度为零：沿该时刻的晨昏线移动地平线上的界线点。
func partialLimbHorizonSolve(
	longitude, latitude float64,
	evaluate limbLimitGap,
	context localSolarEclipseStateContext,
) (float64, float64, bool) {
	kmPerDegLat := math.Pi / 180 * topocentricEarthRadiusKM
	altitudeAt := func(lon, lat float64) float64 {
		return context.stateAt(lon*rad, lat*rad, 0).sunAltitudeRad
	}
	gap0, _, ok := evaluate(longitude, latitude)
	if !ok {
		return longitude, latitude, false
	}
	altitude0 := altitudeAt(longitude, latitude)
	lon, lat := longitude, latitude
	converged := false
	for iteration := 0; iteration < limbLimitIterations+2; iteration++ {
		kmPerDegLon := kmPerDegLat * math.Cos(lat*rad)
		if kmPerDegLon < 1 {
			return longitude, latitude, false
		}
		dLon := limbLimitProbeKM / kmPerDegLon
		dLat := limbLimitProbeKM / kmPerDegLat
		gap, target, ok := evaluate(lon, lat)
		east, _, okEast := evaluate(lon+dLon, lat)
		north, _, okNorth := evaluate(lon, lat+dLat)
		if !ok || !okEast || !okNorth {
			return longitude, latitude, false
		}
		// 高度残差以原点为零点，与 gap0 同理吸收圆月面采样点本身的地平线求解误差。
		altitude := altitudeAt(lon, lat) - altitude0
		altitudeEast := altitudeAt(lon+dLon, lat) - altitude0
		altitudeNorth := altitudeAt(lon, lat+dLat) - altitude0
		gx, gy := (east-gap)/limbLimitProbeKM, (north-gap)/limbLimitProbeKM
		hx, hy := (altitudeEast-altitude)/limbLimitProbeKM, (altitudeNorth-altitude)/limbLimitProbeKM
		determinant := gx*hy - gy*hx
		if !finite(determinant) || determinant == 0 {
			return longitude, latitude, false
		}
		residualGap, residualAltitude := -(gap - gap0 - target), -altitude
		stepX := (residualGap*hy - gy*residualAltitude) / determinant
		stepY := (gx*residualAltitude - residualGap*hx) / determinant
		lon += stepX / kmPerDegLon
		lat += stepY / kmPerDegLat
		if math.Abs(stepX) < 1e-4 && math.Abs(stepY) < 1e-4 {
			converged = true
			break
		}
	}
	shiftKM := math.Hypot(signedAngleDifference(lon, longitude)*kmPerDegLat*math.Cos(latitude*rad), (lat-latitude)*kmPerDegLat)
	if !converged || !finite(shiftKM) || shiftKM > limbLimitMaxShiftKM || lat > 90 || lat < -90 {
		return longitude, latitude, false
	}
	if lon > 180 {
		lon -= 360
	} else if lon < -180 {
		lon += 360
	}
	return lon, lat, true
}
