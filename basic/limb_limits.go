package basic

import (
	"math"
	"time"
)

const (
	// 月掩南北界：星迹相对月心是直线，偏离切点 δ 后月缘需高出 R(1/cosδ − 1)，8° 时约 17 km，远超月面起伏。
	limbGrazeSpanDeg = 8.0
	limbGrazeStepDeg = 0.1
	// 界线点月缘平移上限；超出视为不收敛并保留圆月面点。
	limbLimitMaxShiftKM = 30.0
	limbLimitProbeKM    = 1.0
	limbLimitIterations = 4
	// 太阳 / 月球高度低于此值的界线点不做月缘平移。
	limbLimitHorizonDeg = 0.5
)

// limbLimitGap 返回站点处的圆月面接触间隙与月缘要求的间隙（同一单位），二者相等即为月缘界线。
type limbLimitGap func(longitude, latitude float64) (gap, target float64, ok bool)

// limbLimitSolve 从圆月面界线点出发，沿间隙梯度做牛顿迭代，使 gap − gap0 = target；失败时返回原点。
// gap0 是原点处的圆月面间隙：界线点与站心间隙零点之间的模型差由它吸收，只平移月缘带来的部分。
func limbLimitSolve(longitude, latitude float64, evaluate limbLimitGap) (float64, float64, bool) {
	kmPerDegLat := math.Pi / 180 * topocentricEarthRadiusKM
	gap0, _, ok := evaluate(longitude, latitude)
	if !ok {
		return longitude, latitude, false
	}
	lon, lat := longitude, latitude
	for iteration := 0; iteration < limbLimitIterations; iteration++ {
		gap, target, ok := evaluate(lon, lat)
		if !ok {
			return longitude, latitude, false
		}
		residual := gap - gap0 - target
		kmPerDegLon := kmPerDegLat * math.Cos(lat*rad)
		if kmPerDegLon < 1 {
			return longitude, latitude, false
		}
		dLon := limbLimitProbeKM / kmPerDegLon
		dLat := limbLimitProbeKM / kmPerDegLat
		east, _, okEast := evaluate(lon+dLon, lat)
		north, _, okNorth := evaluate(lon, lat+dLat)
		if !okEast || !okNorth {
			return longitude, latitude, false
		}
		gx := (east - gap) / limbLimitProbeKM
		gy := (north - gap) / limbLimitProbeKM
		norm2 := gx*gx + gy*gy
		if !finite(norm2) || norm2 <= 0 {
			return longitude, latitude, false
		}
		stepX, stepY := -residual*gx/norm2, -residual*gy/norm2
		lon += stepX / kmPerDegLon
		lat += stepY / kmPerDegLat
		if math.Abs(stepX) < 1e-4 && math.Abs(stepY) < 1e-4 {
			break
		}
	}
	shiftKM := math.Hypot((lon-longitude)*kmPerDegLat*math.Cos(latitude*rad), (lat-latitude)*kmPerDegLat)
	if !finite(shiftKM) || shiftKM > limbLimitMaxShiftKM || lat > 90 || lat < -90 {
		return longitude, latitude, false
	}
	if lon > 180 {
		lon -= 360
	} else if lon < -180 {
		lon += 360
	}
	return lon, lat, true
}

func occultationVectorArray(vector occultationPathVector) [3]float64 {
	return [3]float64{vector.x, vector.y, vector.z}
}

// occultationLimbCorrectLimitSeries 把月掩外接触南北界平移到月缘掠掩线：
// 星迹在切点附近按直线处理，有效月缘半径取 max R(θ0+δ)·cosδ。只用于外接触（掠掩）界线。
func occultationLimbCorrectLimitSeries(
	points []OccultationPathPoint,
	contextAt occultationRiseSetContextFunc,
	limb LunarLimb,
	location *time.Location,
	inst *Instance,
) ([]OccultationPathPoint, bool) {
	if limb == nil || len(points) == 0 || contextAt == nil {
		return points, false
	}
	applied := false
	corrected := append([]OccultationPathPoint(nil), points...)
	for index, point := range points {
		if math.Abs(point.MoonAltitude) < limbLimitHorizonDeg {
			continue
		}
		tt := centerTimeTT(point.Time)
		context := contextAt(tt)
		if !context.valid || context.moon.distanceKM <= 0 {
			continue
		}
		frame := newLunarLimbFrame(limb, tt, occultationVectorArray(context.moon.positionKM), inst)
		if !frame.valid {
			continue
		}
		evaluate := func(longitude, latitude float64) (float64, float64, bool) {
			state := context.stateAt(longitude, latitude)
			if !state.valid || !finite(state.contactMetric) {
				return 0, 0, false
			}
			parallaxKM, distanceKM, _ := occultationRiseSetObserverVectors(context.siderealDegrees, longitude, latitude)
			moonKM := occultationPathSub(context.moon.positionKM, parallaxKM)
			var targetKM occultationPathVector
			if context.target.distanceKM > 0 {
				targetKM = occultationPathSub(context.target.positionKM, parallaxKM)
			} else {
				targetKM = context.target.direction
			}
			view, ok := frame.viewToward(occultationVectorArray(moonKM), occultationVectorArray(targetKM))
			if !ok {
				return 0, 0, false
			}
			effective := 0.0
			for offset := -limbGrazeSpanDeg; offset <= limbGrazeSpanDeg+1e-9; offset += limbGrazeStepDeg {
				radius, ok := frame.radiusAt(view, offset)
				if !ok {
					return 0, 0, false
				}
				if value := radius * math.Cos(offset*rad); value > effective {
					effective = value
				}
			}
			moonDistanceKM := occultationRiseSetTopocentricDistance(context.moon, distanceKM)
			if !finite(moonDistanceKM) || moonDistanceKM <= effective {
				return 0, 0, false
			}
			// 有效半径更大时掠掩线外移：需要的圆月面间隙为两者角半径之差（度）。
			target := (angularSemidiameterArcsec(effective, moonDistanceKM) -
				angularSemidiameterArcsec(moonEquatorialRadiusKM, moonDistanceKM)) / 3600
			return state.contactMetric, target, true
		}
		longitude, latitude, ok := limbLimitSolve(point.Longitude, point.Latitude, evaluate)
		if !ok {
			continue
		}
		if !limbGroundMoved(point.Longitude, point.Latitude, longitude, latitude) {
			continue
		}
		corrected[index].Longitude, corrected[index].Latitude = longitude, latitude
		applied = true
		state := context.stateAt(longitude, latitude)
		if state.valid {
			corrected[index].MoonAltitude = state.moonAltitude
		}
	}
	return corrected, applied
}

// limbCorrectCentralLimits 把全食 / 环食带南北界平移到月缘界线。
func (solver solarEclipseSolver) limbCorrectCentralLimits(points []SolarEclipsePathPoint, limb LunarLimb) ([]SolarEclipsePathPoint, bool) {
	if limb == nil || len(points) == 0 {
		return points, false
	}
	limbSolver := solver.withLimbParams(limb)
	applied := false
	corrected := append([]SolarEclipsePathPoint(nil), points...)
	for index, point := range points {
		// 地平线上的 U1/U4 端点处地面梯度趋于零，平移不可信，保留圆月面点。
		if math.Abs(point.SunAltitude) < limbLimitHorizonDeg || !finite(point.JDE) || point.JDE <= 0 {
			continue
		}
		context := solver.localStateContextAt(point.JDE)
		limbContext := limbSolver.localStateContextAt(point.JDE)
		if limbContext.limbFrame == nil {
			continue
		}
		// 与站心同一判据：月缘内切间隙取自站心状态的等效内半径，目标是让它在界线上为零。
		evaluate := func(longitude, latitude float64) (float64, float64, bool) {
			lonRad, latRad := longitude*rad, latitude*rad
			state := context.stateAt(lonRad, latRad, 0)
			limbState := limbContext.stateAt(lonRad, latRad, 0)
			if !limbState.limbApplied {
				return 0, 0, false
			}
			gap := state.movingDiskContactState().internalContactGap()
			return gap, gap - limbState.movingDiskContactState().internalContactGap(), true
		}
		longitude, latitude, ok := limbLimitSolve(point.Longitude, point.Latitude, evaluate)
		if !ok {
			continue
		}
		corrected[index].Longitude, corrected[index].Latitude = longitude, latitude
		// 牛顿迭代可能收敛回原点（常数月缘即如此），只有位移超过噪声容差才算真正修正。
		if limbGroundMoved(point.Longitude, point.Latitude, corrected[index].Longitude, corrected[index].Latitude) {
			applied = true
		}
	}
	return corrected, applied
}
