package basic

import "math"

// JupiterGalileanPhenomenonDiagram 单星现象绘图快照：地球视平面的卫星与影斑、太阳视平面的木影截面，单位木星赤道半径，X 东 Y 北 / single-satellite phenomenon snapshot in Jupiter equatorial radii, X east and Y north.
type JupiterGalileanPhenomenonDiagram struct {
	SatelliteX, SatelliteY    float64
	SatelliteRadius           float64
	InFront                   bool
	ShadowOnDisk              bool
	ShadowX, ShadowY          float64
	ShadowUmbraRadius         float64
	ShadowPenumbraRadius      float64
	MinorRadiusRatio          float64
	PolePositionAngle         float64
	SunX, SunY                float64
	UmbraScale, PenumbraScale float64
	SunMinorRadiusRatio       float64
	SatelliteRadiusSunPlane   float64
	UmbraEdgeDistance         float64 // 卫星中心到木星本影截面边缘的有符号距离，负值在影内；卫星在木星朝阳一侧时为到木心距离 / signed centre-to-umbra-edge distance, negative inside; distance to Jupiter's centre on the sunward side.
	PenumbraEdgeDistance      float64 // 同上，半影截面 / same for the penumbra.
	Phenomenon                JupiterGalileanPhenomenon
}

// JupiterGalileanPhenomenonDiagramWithInstance 接收 TT 时刻单星现象快照；SunX/SunY 为卫星在木心太阳视平面上的位置，木影截面为以木心为中心、长短半轴各乘 UmbraScale/PenumbraScale 的椭圆 / single-satellite snapshot at reception TT; Sun-plane shadow ellipses are Jupiter's limb scaled by UmbraScale and PenumbraScale.
func JupiterGalileanPhenomenonDiagramWithInstance(jde float64, satellite int, inst *Instance) (JupiterGalileanPhenomenonDiagram, bool) {
	if satellite < 1 || satellite > 4 || !isFinite(jde) {
		return JupiterGalileanPhenomenonDiagram{}, false
	}
	context := newJupiterGalileanObservationContext(jde, inst)
	if !isFinite(context.jupiterDistance) || context.jupiterDistance <= 0 {
		return JupiterGalileanPhenomenonDiagram{}, false
	}
	observation := context.observationForSatellite(satellite - 1)
	phenomenon := context.phenomenonForSatellite(satellite - 1)
	if !isFinite(observation.OffsetXJupiterRadii) || !isFinite(observation.OffsetYJupiterRadii) {
		return JupiterGalileanPhenomenonDiagram{}, false
	}
	radiusAU := jupiterGalileanEquatorialRadiusKM / astronomicalUnitKM
	state := Vector3{observation.State.X, observation.State.Y, observation.State.Z}
	behind := vectorDot(state, context.sunLineOfSight)
	result := JupiterGalileanPhenomenonDiagram{
		SatelliteX:              observation.OffsetXJupiterRadii,
		SatelliteY:              observation.OffsetYJupiterRadii,
		SatelliteRadius:         math.Atan2(jupiterGalileanSatelliteRadiusJupiterRadii(satellite)*radiusAU, observation.Distance) / math.Atan2(radiusAU, context.jupiterDistance),
		InFront:                 observation.InFrontOfJupiter,
		ShadowOnDisk:            phenomenon.ShadowTransit,
		MinorRadiusRatio:        context.earthMinorRadius,
		PolePositionAngle:       math.Atan2(vectorDot(context.bodyZ, context.east), vectorDot(context.bodyZ, context.north)) * deg,
		SunX:                    vectorDot(state, context.sunEast) / radiusAU,
		SunY:                    vectorDot(state, context.sunNorth) / radiusAU,
		UmbraScale:              jupiterUmbraScale(behind, context.sunDistanceAU),
		PenumbraScale:           jupiterPenumbraScale(behind, context.sunDistanceAU),
		SunMinorRadiusRatio:     context.sunMinorRadius,
		SatelliteRadiusSunPlane: jupiterGalileanSatelliteRadiusJupiterRadii(satellite),
		Phenomenon:              phenomenon,
	}
	result.ShadowX, result.ShadowY = context.shadowAxisProjection(state)
	result.ShadowUmbraRadius, result.ShadowPenumbraRadius = context.shadowSection(state, satellite)
	result.UmbraEdgeDistance, result.PenumbraEdgeDistance = math.Hypot(result.SunX, result.SunY), math.Hypot(result.SunX, result.SunY)
	if behind > 0 && result.UmbraScale > 0 {
		result.UmbraEdgeDistance = ellipseSignedDistance(result.SunX, result.SunY, result.UmbraScale, context.sunMinorRadius*result.UmbraScale)
	}
	if behind > 0 && result.PenumbraScale > 0 {
		result.PenumbraEdgeDistance = ellipseSignedDistance(result.SunX, result.SunY, result.PenumbraScale, context.sunMinorRadius*result.PenumbraScale)
	}
	if !phenomenon.ShadowTransit {
		result.Phenomenon.ShadowOffsetXArcsec, result.Phenomenon.ShadowOffsetYArcsec = 0, 0
		result.Phenomenon.ShadowOffsetXJupiterRadii, result.Phenomenon.ShadowOffsetYJupiterRadii = 0, 0
	}
	return result, true
}

func (context jupiterGalileanObservationContext) shadowSection(state Vector3, satellite int) (float64, float64) {
	shadow := context.shadowPointFor(state)
	distance := shadow.pathLengthAU
	if !shadow.hasIntersection {
		denominator := vectorDot(context.sunLineOfSight, context.lineOfSight)
		if math.Abs(denominator) < 1e-12 {
			return 0, 0
		}
		distance = -vectorDot(state, context.lineOfSight) / denominator
	}
	if distance <= 0 || context.sunDistanceAU <= 0 {
		return 0, 0
	}
	radius := jupiterGalileanSatelliteRadiusJupiterRadii(satellite)
	solarRadius := solarRadiusAU * astronomicalUnitKM / jupiterGalileanEquatorialRadiusKM
	return math.Max(0, radius-distance*(solarRadius-radius)/context.sunDistanceAU), jupiterGalileanPenumbraRadiusJupiterRadii(satellite, distance, context.sunDistanceAU)
}

// shadowAxisProjection 影轴落点的地球视平面位置：与木星相交时取交点，未相交时取卫星沿太阳方向投到过木心、垂直视线的平面上的点，便于画影斑入盘前后的路径。
func (context jupiterGalileanObservationContext) shadowAxisProjection(stateVector Vector3) (float64, float64) {
	radiusAU := jupiterGalileanEquatorialRadiusKM / astronomicalUnitKM
	satelliteBody := context.toBodyCoordinates(stateVector)
	satelliteBody = Vector3{satelliteBody[0] / radiusAU, satelliteBody[1] / radiusAU, satelliteBody[2] / radiusAU}
	directionBody := context.toBodyCoordinates(context.sunLineOfSight)
	if hit, ok := ellipsoidRayIntersection(satelliteBody, directionBody, jupiterPolarRadiusRatio()); ok {
		point := context.fromBodyCoordinates(hit)
		return vectorDot(point, context.east), vectorDot(point, context.north)
	}
	toward := vectorDot(context.sunLineOfSight, context.lineOfSight)
	if math.Abs(toward) < 1e-12 {
		return vectorDot(stateVector, context.east) / radiusAU, vectorDot(stateVector, context.north) / radiusAU
	}
	step := -vectorDot(stateVector, context.lineOfSight) / toward
	point := Vector3{stateVector[0] + step*context.sunLineOfSight[0], stateVector[1] + step*context.sunLineOfSight[1], stateVector[2] + step*context.sunLineOfSight[2]}
	return vectorDot(point, context.east) / radiusAU, vectorDot(point, context.north) / radiusAU
}
