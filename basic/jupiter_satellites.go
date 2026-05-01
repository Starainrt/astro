package basic

import (
	"math"

	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

const (
	jupiterGalileanReferenceJD        = 2433282.5
	jupiterGalileanLongPeriodShift    = 310910.16
	jupiterGalileanMinSolarLPYear     = 1150.0
	jupiterGalileanMaxSolarLPYear     = 2750.0
	jupiterGalileanEquatorialRadiusKM = 71492.0
	astronomicalUnitKM                = 149597870.691
)

var (
	jupiterGalileanBaseMeanLongitudes = [4]float64{3.55155228618240, 1.76932271112347, 0.878207923589328, 0.376486233433828}
	jupiterGalileanMu                 = [4]float64{2.82489428433814e-07, 2.82483274392893e-07, 2.82498184184723e-07, 2.82492144889909e-07}
)

const (
	jupiterGalileanFrameNode = 6.24950183065715
	jupiterGalileanFrameTilt = 0.445094736497665
)

type jupiterGalileanL1Term struct {
	Amp    float64
	Period float64
	Phase  float64
}

// JupiterGalileanState 木星伽利略卫星原始状态 / raw Galilean-satellite state.
//
// 输入 jd 使用 TT/TDB 对应的儒略日；返回值为 IMCCE L1 理论的木心 J2000 平赤道直角坐标与速度，单位 AU / AU/day。
// The input jd is a TT/TDB Julian day. Returned coordinates are Jovicentric J2000 mean-equatorial position and velocity from the IMCCE L1 theory, in AU and AU/day.
type JupiterGalileanState struct {
	X  float64
	Y  float64
	Z  float64
	VX float64
	VY float64
	VZ float64
}

// JupiterGalileanObservation 木星伽利略卫星视位置 / apparent Galilean-satellite geometry.
//
// 视位置相对木星中心定义：X 向天球东为正，Y 向天球北为正，Z>0 表示比木星更远、位于盘后。
// Apparent offsets are relative to Jupiter's center: X is positive to celestial east, Y to celestial north, and Z>0 means farther than Jupiter and behind the disk.
type JupiterGalileanObservation struct {
	State JupiterGalileanState

	RA       float64
	Dec      float64
	Distance float64

	OffsetXArcsec float64
	OffsetYArcsec float64

	OffsetXJupiterRadii float64
	OffsetYJupiterRadii float64
	OffsetZJupiterRadii float64

	InFrontOfJupiter bool
}

// JupiterGalileanSatelliteState 伽利略卫星木心 J2000 状态 / Jovicentric J2000 state of a Galilean satellite.
//
// satellite 取 1=Io, 2=Europa, 3=Ganymede, 4=Callisto。jd 为 TT/TDB 对应儒略日。
// satellite is 1=Io, 2=Europa, 3=Ganymede, 4=Callisto. jd is a TT/TDB Julian day.
func JupiterGalileanSatelliteState(jd float64, satellite int) JupiterGalileanState {
	if satellite < 1 || satellite > 4 || !isFinite(jd) {
		return invalidJupiterGalileanState()
	}
	et := jd - jupiterGalileanReferenceJD
	includeSolarLongPeriod := jupiterGalileanUseSolarLongPeriod(jd)
	return jupiterGalileanSatelliteStateAtET(et, satellite-1, includeSolarLongPeriod)
}

// JupiterGalileanSatelliteStates 四颗伽利略卫星木心 J2000 状态 / Jovicentric J2000 states of the four Galilean satellites.
//
// 返回次序固定为 Io、Europa、Ganymede、Callisto。
// The returned order is Io, Europa, Ganymede, Callisto.
func JupiterGalileanSatelliteStates(jd float64) [4]JupiterGalileanState {
	var states [4]JupiterGalileanState
	et := jd - jupiterGalileanReferenceJD
	includeSolarLongPeriod := jupiterGalileanUseSolarLongPeriod(jd)
	for i := range states {
		states[i] = jupiterGalileanSatelliteStateAtET(et, i, includeSolarLongPeriod)
	}
	return states
}

// JupiterGalileanSatelliteObservation 伽利略卫星视位置 / apparent geometry of a Galilean satellite.
//
// jd 为 TT/TDB 对应儒略日；返回卫星的天球视赤道坐标，以及相对木星中心的东/北平面偏移。
// jd is a TT/TDB Julian day. The result contains the satellite's astrometric equatorial coordinates and its east/north sky-plane offsets relative to Jupiter's center.
func JupiterGalileanSatelliteObservation(jd float64, satellite int) JupiterGalileanObservation {
	if satellite < 1 || satellite > 4 || !isFinite(jd) {
		return invalidJupiterGalileanObservation()
	}
	context := newJupiterGalileanObservationContext(jd)
	return context.observationForSatellite(satellite - 1)
}

// JupiterGalileanSatelliteObservations 四颗伽利略卫星视位置 / apparent geometry of the four Galilean satellites.
//
// 返回次序固定为 Io、Europa、Ganymede、Callisto。
// The returned order is Io, Europa, Ganymede, Callisto.
func JupiterGalileanSatelliteObservations(jd float64) [4]JupiterGalileanObservation {
	var observations [4]JupiterGalileanObservation
	context := newJupiterGalileanObservationContext(jd)
	for i := range observations {
		observations[i] = context.observationForSatellite(i)
	}
	return observations
}

type jupiterGalileanObservationContext struct {
	jd               float64
	targetJD         float64
	earthHelioJ2000  Vector3
	jupiterGeoJ2000  Vector3
	jupiterDistance  float64
	jupiterLightTime float64
	sunDistanceAU    float64
	east             Vector3
	north            Vector3
	lineOfSight      Vector3
	earthDirection   Vector3
	sunDirection     Vector3
	sunLineOfSight   Vector3
	sunEast          Vector3
	sunNorth         Vector3
	earthMinorRadius float64
	sunMinorRadius   float64
	bodyX            Vector3
	bodyY            Vector3
	bodyZ            Vector3
}

func newJupiterGalileanObservationContext(jd float64) jupiterGalileanObservationContext {
	context := jupiterGalileanObservationContext{jd: jd}
	if !isFinite(jd) {
		return context
	}
	context.earthHelioJ2000 = rotateEclipticToEquatorial(earthHeliocentricVectorJ2000(jd), orbitJ2000Obliquity)
	context.jupiterGeoJ2000, context.jupiterLightTime = jupiterAstrometricGeocentricVectorJ2000(jd, context.earthHelioJ2000)
	context.targetJD = jd - context.jupiterLightTime
	context.jupiterDistance = vectorMagnitude(context.jupiterGeoJ2000)
	if context.jupiterDistance == 0 {
		return context
	}
	context.lineOfSight = normalizeVector(context.jupiterGeoJ2000)
	context.earthDirection = Vector3{-context.lineOfSight[0], -context.lineOfSight[1], -context.lineOfSight[2]}
	ra, dec := vectorToRaDec(context.lineOfSight)
	context.east = Vector3{-Sin(ra), Cos(ra), 0}
	context.north = Vector3{-Cos(ra) * Sin(dec), -Sin(ra) * Sin(dec), Cos(dec)}
	jupiterHelio := rotateEclipticToEquatorial(jupiterHeliocentricVectorJ2000(context.targetJD), orbitJ2000Obliquity)
	context.sunDistanceAU = vectorMagnitude(jupiterHelio)
	context.sunDirection = normalizeVector(Vector3{-jupiterHelio[0], -jupiterHelio[1], -jupiterHelio[2]})
	context.sunLineOfSight = Vector3{-context.sunDirection[0], -context.sunDirection[1], -context.sunDirection[2]}
	sunRA, sunDec := vectorToRaDec(context.sunLineOfSight)
	context.sunEast = Vector3{-Sin(sunRA), Cos(sunRA), 0}
	context.sunNorth = Vector3{-Cos(sunRA) * Sin(sunDec), -Sin(sunRA) * Sin(sunDec), Cos(sunDec)}
	poleRA, poleDec, _ := jupiterPoleRotation(context.targetJD)
	context.bodyZ = raDecToVector(poleRA, poleDec)
	context.bodyX = normalizeVector(Vector3{-math.Sin(poleRA * rad), math.Cos(poleRA * rad), 0})
	context.bodyY = normalizeVector(pxp(context.bodyZ, context.bodyX))
	context.earthMinorRadius = jupiterProjectedMinorRadius(context.earthDirection, context.bodyZ)
	context.sunMinorRadius = jupiterProjectedMinorRadius(context.sunDirection, context.bodyZ)
	return context
}

func (context jupiterGalileanObservationContext) observationForSatellite(index int) JupiterGalileanObservation {
	if index < 0 || index >= 4 || context.jupiterDistance == 0 {
		return invalidJupiterGalileanObservation()
	}
	state, geocentric := jupiterGalileanSatelliteAstrometricGeocentric(index, context.jd, context.jupiterLightTime, context.earthHelioJ2000)
	direction := normalizeVector(geocentric)
	ra, dec := vectorToRaDec(direction)
	distance := vectorMagnitude(geocentric)
	relative := Vector3{
		geocentric[0] - context.jupiterGeoJ2000[0],
		geocentric[1] - context.jupiterGeoJ2000[1],
		geocentric[2] - context.jupiterGeoJ2000[2],
	}
	zAU := vectorDot(relative, context.lineOfSight)
	radiusAU := jupiterGalileanEquatorialRadiusKM / astronomicalUnitKM
	offsetXRad, offsetYRad := tangentPlaneOffsetAngles(direction, context.lineOfSight, context.east, context.north)
	jupiterSemidiameterArcsec := math.Atan2(radiusAU, context.jupiterDistance) * deg * 3600
	return JupiterGalileanObservation{
		State: state,

		RA:       ra,
		Dec:      dec,
		Distance: distance,

		OffsetXArcsec: offsetXRad * deg * 3600,
		OffsetYArcsec: offsetYRad * deg * 3600,

		OffsetXJupiterRadii: offsetXRad * deg * 3600 / jupiterSemidiameterArcsec,
		OffsetYJupiterRadii: offsetYRad * deg * 3600 / jupiterSemidiameterArcsec,
		OffsetZJupiterRadii: zAU / radiusAU,

		InFrontOfJupiter: zAU < 0,
	}
}

func tangentPlaneOffsetAngles(target, center, east, north Vector3) (float64, float64) {
	denominator := vectorDot(target, center)
	return math.Atan2(vectorDot(target, east), denominator), math.Atan2(vectorDot(target, north), denominator)
}

func jupiterGalileanSatelliteAstrometricGeocentric(index int, jd, initialLightTime float64, earthHelioJ2000 Vector3) (JupiterGalileanState, Vector3) {
	lightTime := initialLightTime
	state := JupiterGalileanState{}
	result := Vector3{}
	includeSolarLongPeriod := jupiterGalileanUseSolarLongPeriod(jd)
	for i := 0; i < 8; i++ {
		targetJD := jd - lightTime
		jupiterHelio := rotateEclipticToEquatorial(jupiterHeliocentricVectorJ2000(targetJD), orbitJ2000Obliquity)
		state = jupiterGalileanSatelliteStateAtET(targetJD-jupiterGalileanReferenceJD, index, includeSolarLongPeriod)
		result = Vector3{
			jupiterHelio[0] + state.X - earthHelioJ2000[0],
			jupiterHelio[1] + state.Y - earthHelioJ2000[1],
			jupiterHelio[2] + state.Z - earthHelioJ2000[2],
		}
		nextLightTime := lightTimeDaysPerAU * vectorMagnitude(result)
		if math.Abs(nextLightTime-lightTime) < 1e-12 {
			break
		}
		lightTime = nextLightTime
	}
	return state, result
}

func jupiterAstrometricGeocentricVectorJ2000(jd float64, earthHelioJ2000 Vector3) (Vector3, float64) {
	lightTime := 0.0
	result := Vector3{}
	for i := 0; i < 8; i++ {
		jupiterHelio := rotateEclipticToEquatorial(jupiterHeliocentricVectorJ2000(jd-lightTime), orbitJ2000Obliquity)
		result = Vector3{
			jupiterHelio[0] - earthHelioJ2000[0],
			jupiterHelio[1] - earthHelioJ2000[1],
			jupiterHelio[2] - earthHelioJ2000[2],
		}
		nextLightTime := lightTimeDaysPerAU * vectorMagnitude(result)
		if math.Abs(nextLightTime-lightTime) < 1e-12 {
			return result, nextLightTime
		}
		lightTime = nextLightTime
	}
	return result, lightTime
}

func jupiterHeliocentricVectorJ2000(jd float64) Vector3 {
	return eclipticVectorAtReferenceEpoch(
		eclipticCartesian(
			planet.WherePlanet(4, 0, jd),
			planet.WherePlanet(4, 1, jd),
			planet.WherePlanet(4, 2, jd),
		),
		jd,
		orbitReferenceJD,
	)
}

func jupiterGalileanSatelliteStateAtET(et float64, index int, includeSolarLongPeriod bool) JupiterGalileanState {
	elements := jupiterGalileanElementsAtET(et, index, includeSolarLongPeriod)
	pv := jupiterGalileanElementsToPV(jupiterGalileanMu[index], elements)
	cosNode, sinNode := math.Cos(jupiterGalileanFrameNode), math.Sin(jupiterGalileanFrameNode)
	cosTilt, sinTilt := math.Cos(jupiterGalileanFrameTilt), math.Sin(jupiterGalileanFrameTilt)
	return JupiterGalileanState{
		X:  pv[0]*cosNode - pv[1]*sinNode*cosTilt + pv[2]*sinTilt*sinNode,
		Y:  pv[0]*sinNode + pv[1]*cosNode*cosTilt - pv[2]*sinTilt*cosNode,
		Z:  pv[1]*sinTilt + pv[2]*cosTilt,
		VX: pv[3]*cosNode - pv[4]*sinNode*cosTilt + pv[5]*sinTilt*sinNode,
		VY: pv[3]*sinNode + pv[4]*cosNode*cosTilt - pv[5]*sinTilt*cosNode,
		VZ: pv[4]*sinTilt + pv[5]*cosTilt,
	}
}

type jupiterGalileanElements struct {
	A float64
	L float64
	K float64
	H float64
	Q float64
	P float64
}

func jupiterGalileanElementsAtET(et float64, index int, includeSolarLongPeriod bool) jupiterGalileanElements {
	longPeriod := jupiterGalileanEvaluateSeries(jupiterGalileanL1LongPeriodTerms[index], et+jupiterGalileanLongPeriodShift, et, includeSolarLongPeriod, index)
	crossPeriod := jupiterGalileanEvaluateSeries(jupiterGalileanL1CrossPeriodTerms[index], et, et, false, index)
	combined := jupiterGalileanElements{
		A: longPeriod.A + crossPeriod.A,
		L: longPeriod.L + crossPeriod.L + jupiterGalileanBaseMeanLongitudes[index]*et,
		K: longPeriod.K + crossPeriod.K,
		H: longPeriod.H + crossPeriod.H,
		Q: longPeriod.Q + crossPeriod.Q,
		P: longPeriod.P + crossPeriod.P,
	}
	combined.L = math.Atan2(math.Sin(combined.L), math.Cos(combined.L))
	if combined.L < 0 {
		combined.L += 2 * math.Pi
	}
	return combined
}

func jupiterGalileanEvaluateSeries(blocks [4][]jupiterGalileanL1Term, angleTime, et float64, includeSolarLongPeriod bool, index int) jupiterGalileanElements {
	vals := [5]float64{}
	if includeSolarLongPeriod {
		x := (et/365.25 - 0.5*(812.721806990360-819.727638594856)) / (0.5 * (812.721806990360 - -819.727638594856))
		tn := [9]float64{1, x}
		for i := 2; i < len(tn); i++ {
			tn[i] = 2*x*tn[i-1] - tn[i-2]
		}
		for variable := 0; variable < len(vals); variable++ {
			sum := 0.0
			for term := 0; term < len(tn); term++ {
				sum += jupiterGalileanL1Chebyshev[index][variable][term] * tn[term]
			}
			vals[variable] = sum - 0.5*jupiterGalileanL1Chebyshev[index][variable][0]
		}
	}

	result := jupiterGalileanElements{}
	for blockIndex, terms := range blocks {
		realPart, imagPart := 0.0, 0.0
		for _, term := range terms {
			angle := term.Phase
			if term.Period != 0 {
				angle += 2 * math.Pi * angleTime / term.Period
			}
			realPart += term.Amp * math.Cos(angle)
			imagPart += term.Amp * math.Sin(angle)
		}
		switch blockIndex {
		case 0:
			result.A = realPart
		case 1:
			result.L = realPart + vals[0]
		case 2:
			result.K = realPart + vals[1]
			result.H = imagPart + vals[2]
		case 3:
			result.Q = realPart + vals[3]
			result.P = imagPart + vals[4]
		}
	}
	return result
}

func jupiterGalileanElementsToPV(mu float64, elements jupiterGalileanElements) [6]float64 {
	k := elements.K
	h := elements.H
	q := elements.Q
	p := elements.P
	a := elements.A
	al := elements.L
	an := math.Sqrt(mu / math.Pow(a, 3))
	ee := al + k*math.Sin(al) - h*math.Cos(al)
	for {
		ce := math.Cos(ee)
		se := math.Sin(ee)
		de := (al - ee + k*se - h*ce) / (1 - k*ce - h*se)
		ee += de
		if math.Abs(de) < 1e-12 {
			break
		}
	}
	ce := math.Cos(ee)
	se := math.Sin(ee)
	dle := h*ce - k*se
	rsam1 := -k*ce - h*se
	asr := 1 / (1 + rsam1)
	phi := math.Sqrt(1 - k*k - h*h)
	psi := 1 / (1 + phi)
	x1 := a * (ce - k - psi*h*dle)
	y1 := a * (se - h + psi*k*dle)
	vx1 := an * asr * a * (-se - psi*h*rsam1)
	vy1 := an * asr * a * (ce + psi*k*rsam1)
	f2 := 2 * math.Sqrt(1-q*q-p*p)
	p2 := 1 - 2*p*p
	q2 := 1 - 2*q*q
	pq := 2 * p * q
	return [6]float64{
		x1*p2 + y1*pq,
		x1*pq + y1*q2,
		(q*y1 - x1*p) * f2,
		vx1*p2 + vy1*pq,
		vx1*pq + vy1*q2,
		(q*vy1 - vx1*p) * f2,
	}
}

func jupiterGalileanUseSolarLongPeriod(jd float64) bool {
	year := 2000.0 + (jd-2451545.0)/365.25
	return year >= jupiterGalileanMinSolarLPYear && year <= jupiterGalileanMaxSolarLPYear
}

func invalidJupiterGalileanState() JupiterGalileanState {
	nan := math.NaN()
	return JupiterGalileanState{X: nan, Y: nan, Z: nan, VX: nan, VY: nan, VZ: nan}
}

func invalidJupiterGalileanObservation() JupiterGalileanObservation {
	nan := math.NaN()
	return JupiterGalileanObservation{
		State:               invalidJupiterGalileanState(),
		RA:                  nan,
		Dec:                 nan,
		Distance:            nan,
		OffsetXArcsec:       nan,
		OffsetYArcsec:       nan,
		OffsetXJupiterRadii: nan,
		OffsetYJupiterRadii: nan,
		OffsetZJupiterRadii: nan,
		InFrontOfJupiter:    false,
	}
}
