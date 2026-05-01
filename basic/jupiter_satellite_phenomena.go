package basic

import "math"

const solarRadiusAU = 695700.0 / astronomicalUnitKM

// JupiterGalileanPhenomenon 木星伽利略卫星瞬时现象 / instantaneous Galilean-satellite phenomena.
//
// Transit 表示卫星本体在木星盘前；Occultation 表示卫星在木星盘后被掩蔽；Eclipse 表示卫星落入木星本影；ShadowTransit 表示卫星影心落在可见木星盘面上。
// Transit means the satellite itself is in front of Jupiter's disk; Occultation means it is hidden behind the disk; Eclipse means the satellite lies in Jupiter's umbra; ShadowTransit means the center of the satellite shadow falls on the visible Jovian disk.
type JupiterGalileanPhenomenon struct {
	Transit       bool
	Occultation   bool
	Eclipse       bool
	ShadowTransit bool

	ShadowOffsetXArcsec float64
	ShadowOffsetYArcsec float64

	ShadowOffsetXJupiterRadii float64
	ShadowOffsetYJupiterRadii float64
}

// JupiterGalileanSatellitePhenomenon 单颗伽利略卫星瞬时现象 / instantaneous phenomena of one Galilean satellite.
func JupiterGalileanSatellitePhenomenon(jd float64, satellite int) JupiterGalileanPhenomenon {
	if satellite < 1 || satellite > 4 || !isFinite(jd) {
		return invalidJupiterGalileanPhenomenon()
	}
	context := newJupiterGalileanObservationContext(jd)
	return context.phenomenonForSatellite(satellite - 1)
}

// JupiterGalileanSatellitePhenomena 四颗伽利略卫星瞬时现象 / instantaneous phenomena of the four Galilean satellites.
func JupiterGalileanSatellitePhenomena(jd float64) [4]JupiterGalileanPhenomenon {
	var phenomena [4]JupiterGalileanPhenomenon
	context := newJupiterGalileanObservationContext(jd)
	for i := range phenomena {
		phenomena[i] = context.phenomenonForSatellite(i)
	}
	return phenomena
}

func (context jupiterGalileanObservationContext) phenomenonForSatellite(index int) JupiterGalileanPhenomenon {
	if index < 0 || index >= 4 || context.jupiterDistance == 0 {
		return invalidJupiterGalileanPhenomenon()
	}
	observation := context.observationForSatellite(index)
	stateVector := Vector3{observation.State.X, observation.State.Y, observation.State.Z}
	radiusAU := jupiterGalileanEquatorialRadiusKM / astronomicalUnitKM
	xEarth := observation.OffsetXJupiterRadii
	yEarth := observation.OffsetYJupiterRadii
	onEarthDisk := ellipseInside(xEarth, yEarth, 1, context.earthMinorRadius)

	xSunAU := vectorDot(stateVector, context.sunEast)
	ySunAU := vectorDot(stateVector, context.sunNorth)
	zSunAU := vectorDot(stateVector, context.sunLineOfSight)
	xSun := xSunAU / radiusAU
	ySun := ySunAU / radiusAU
	umbraScale := jupiterUmbraScale(zSunAU, context.sunDistanceAU)
	eclipse := false
	if zSunAU > 0 && umbraScale > 0 {
		eclipse = ellipseInside(xSun, ySun, umbraScale, context.sunMinorRadius*umbraScale)
	}

	shadowTransit, shadowXAU, shadowYAU := context.shadowTransitFor(stateVector)
	phenomenon := JupiterGalileanPhenomenon{
		Transit:       onEarthDisk && observation.InFrontOfJupiter,
		Occultation:   onEarthDisk && !observation.InFrontOfJupiter,
		Eclipse:       eclipse,
		ShadowTransit: shadowTransit,

		ShadowOffsetXArcsec:       math.NaN(),
		ShadowOffsetYArcsec:       math.NaN(),
		ShadowOffsetXJupiterRadii: math.NaN(),
		ShadowOffsetYJupiterRadii: math.NaN(),
	}
	if shadowTransit {
		phenomenon.ShadowOffsetXArcsec = math.Atan2(shadowXAU, context.jupiterDistance) * deg * 3600
		phenomenon.ShadowOffsetYArcsec = math.Atan2(shadowYAU, context.jupiterDistance) * deg * 3600
		phenomenon.ShadowOffsetXJupiterRadii = shadowXAU / radiusAU
		phenomenon.ShadowOffsetYJupiterRadii = shadowYAU / radiusAU
	}
	return phenomenon
}

func (context jupiterGalileanObservationContext) shadowTransitFor(stateVector Vector3) (bool, float64, float64) {
	radiusAU := jupiterGalileanEquatorialRadiusKM / astronomicalUnitKM
	satelliteBody := context.toBodyCoordinates(stateVector)
	satelliteBody = Vector3{satelliteBody[0] / radiusAU, satelliteBody[1] / radiusAU, satelliteBody[2] / radiusAU}
	directionBody := context.toBodyCoordinates(context.sunLineOfSight)
	intersectionBody, ok := ellipsoidRayIntersection(satelliteBody, directionBody, jupiterPolarRadiusRatio())
	if !ok {
		return false, 0, 0
	}
	normalBody := Vector3{intersectionBody[0], intersectionBody[1], intersectionBody[2] / (jupiterPolarRadiusRatio() * jupiterPolarRadiusRatio())}
	earthBody := context.toBodyCoordinates(context.earthDirection)
	if vectorDot(normalBody, earthBody) <= 0 {
		return false, 0, 0
	}
	intersection := context.fromBodyCoordinates(Vector3{
		intersectionBody[0] * radiusAU,
		intersectionBody[1] * radiusAU,
		intersectionBody[2] * radiusAU,
	})
	xAU := vectorDot(intersection, context.east)
	yAU := vectorDot(intersection, context.north)
	x := xAU / radiusAU
	y := yAU / radiusAU
	if !ellipseInside(x, y, 1, context.earthMinorRadius) {
		return false, 0, 0
	}
	return true, xAU, yAU
}

func (context jupiterGalileanObservationContext) toBodyCoordinates(vector Vector3) Vector3 {
	return Vector3{
		vectorDot(vector, context.bodyX),
		vectorDot(vector, context.bodyY),
		vectorDot(vector, context.bodyZ),
	}
}

func (context jupiterGalileanObservationContext) fromBodyCoordinates(vector Vector3) Vector3 {
	return Vector3{
		context.bodyX[0]*vector[0] + context.bodyY[0]*vector[1] + context.bodyZ[0]*vector[2],
		context.bodyX[1]*vector[0] + context.bodyY[1]*vector[1] + context.bodyZ[1]*vector[2],
		context.bodyX[2]*vector[0] + context.bodyY[2]*vector[1] + context.bodyZ[2]*vector[2],
	}
}

func jupiterProjectedMinorRadius(direction, pole Vector3) float64 {
	sinBeta := vectorDot(direction, pole)
	cos2Beta := 1 - sinBeta*sinBeta
	if cos2Beta < 0 {
		cos2Beta = 0
	}
	ratio := jupiterPolarRadiusRatio()
	return math.Sqrt(sinBeta*sinBeta + ratio*ratio*cos2Beta)
}

func jupiterPolarRadiusRatio() float64 {
	return jupiterPhysicalModel.polarRadius / jupiterPhysicalModel.equatorialRadius
}

func ellipseInside(x, y, major, minor float64) bool {
	if major <= 0 || minor <= 0 {
		return false
	}
	return (x*x)/(major*major)+(y*y)/(minor*minor) <= 1+1e-12
}

func ellipsoidRayIntersection(origin, direction Vector3, polarRatio float64) (Vector3, bool) {
	invPolar2 := 1 / (polarRatio * polarRatio)
	a := direction[0]*direction[0] + direction[1]*direction[1] + direction[2]*direction[2]*invPolar2
	b := 2 * (origin[0]*direction[0] + origin[1]*direction[1] + origin[2]*direction[2]*invPolar2)
	c := origin[0]*origin[0] + origin[1]*origin[1] + origin[2]*origin[2]*invPolar2 - 1
	discriminant := b*b - 4*a*c
	if discriminant < 0 {
		return Vector3{}, false
	}
	sqrtDiscriminant := math.Sqrt(discriminant)
	t1 := (-b - sqrtDiscriminant) / (2 * a)
	t2 := (-b + sqrtDiscriminant) / (2 * a)
	t := math.Inf(1)
	if t1 > 0 {
		t = t1
	}
	if t2 > 0 && t2 < t {
		t = t2
	}
	if !isFinite(t) {
		return Vector3{}, false
	}
	return Vector3{
		origin[0] + t*direction[0],
		origin[1] + t*direction[1],
		origin[2] + t*direction[2],
	}, true
}

func jupiterUmbraScale(distanceBehindAU, sunDistanceAU float64) float64 {
	if distanceBehindAU <= 0 || sunDistanceAU <= 0 {
		return 0
	}
	jupiterRadiusAU := jupiterGalileanEquatorialRadiusKM / astronomicalUnitKM
	umbraLength := jupiterRadiusAU * sunDistanceAU / (solarRadiusAU - jupiterRadiusAU)
	if umbraLength <= 0 {
		return 0
	}
	return 1 - distanceBehindAU/umbraLength
}

func invalidJupiterGalileanPhenomenon() JupiterGalileanPhenomenon {
	nan := math.NaN()
	return JupiterGalileanPhenomenon{
		ShadowOffsetXArcsec:       nan,
		ShadowOffsetYArcsec:       nan,
		ShadowOffsetXJupiterRadii: nan,
		ShadowOffsetYJupiterRadii: nan,
	}
}
