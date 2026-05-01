package basic

import "math"

// SolarEclipseRadiusModel 表示日食计算中月亮平均半径 k 的取法。
type SolarEclipseRadiusModel string

const (
	// SolarEclipseModelIAUSingleK 使用 IAU 单一月亮平均半径 k。
	SolarEclipseModelIAUSingleK SolarEclipseRadiusModel = "iau_single_k"
	// SolarEclipseModelNASABulletinSplitK 使用 NASA bulletin 的 Split-K 口径。
	SolarEclipseModelNASABulletinSplitK SolarEclipseRadiusModel = "nasa_bulletin_split_k"
)

// SolarEclipseType 表示整场日食的全局食型。
type SolarEclipseType string

const (
	// SolarEclipseNone 表示该次朔月没有发生日食。
	SolarEclipseNone SolarEclipseType = "none"
	// SolarEclipsePartial 表示日偏食。
	SolarEclipsePartial SolarEclipseType = "partial"
	// SolarEclipseAnnular 表示日环食。
	SolarEclipseAnnular SolarEclipseType = "annular"
	// SolarEclipseTotal 表示日全食。
	SolarEclipseTotal SolarEclipseType = "total"
	// SolarEclipseHybrid 表示全环食/混合食。
	SolarEclipseHybrid SolarEclipseType = "hybrid"
)

// SolarEclipseCentrality 表示中心线进入地球的方式。
type SolarEclipseCentrality string

const (
	// SolarEclipseNonCentral 表示无中心线进入地球。
	SolarEclipseNonCentral SolarEclipseCentrality = "non_central"
	// SolarEclipseCentralOneLimit 表示中心线只形成一侧极限条件。
	SolarEclipseCentralOneLimit SolarEclipseCentrality = "central_one_limit"
	// SolarEclipseCentralTwoLimits 表示中心线完整进入地球，两侧都有界线。
	SolarEclipseCentralTwoLimits SolarEclipseCentrality = "central_two_limits"
)

// SolarEclipseResult 表示一次朔月附近的全局日食几何结果。
//
// 所有时刻字段都使用力学时儒略日（JDE, TT）。
// 输入 seedJDE 只需要落在目标朔月附近，允许相差数天。
type SolarEclipseResult struct {
	Model      SolarEclipseRadiusModel
	Type       SolarEclipseType
	Centrality SolarEclipseCentrality

	// GreatestEclipse 是全局“影轴最接近地心”的时刻。
	GreatestEclipse float64

	// PartialBeginOnEarth / PartialEndOnEarth 是地球范围的偏食开始 / 结束时刻。
	PartialBeginOnEarth float64
	PartialEndOnEarth   float64
	// CentralBeginOnEarth / CentralEndOnEarth 是中心线进入 / 离开地球的时刻。
	CentralBeginOnEarth float64
	CentralEndOnEarth   float64

	// Magnitude 是全局食分。
	Magnitude float64
	// Gamma 是月影轴到地心的有符号最小距离，单位为地球赤道半径。
	Gamma float64
	// PathWidthKM 是食甚点处中心食带宽度。非中心食时为 0。
	PathWidthKM float64

	// GreatestLongitude / GreatestLatitude 是日食食甚点地理坐标，东经为正，西经为负。
	GreatestLongitude float64
	GreatestLatitude  float64

	HasPartial bool
	HasCentral bool
	HasAnnular bool
	HasTotal   bool
	HasHybrid  bool
}

type solarEclipseModelParameters struct {
	penumbralK float64
	umbralK    float64
}

type solarEclipseShadowRadii struct {
	penumbraRadius float64
	umbraRadius    float64
	absUmbraRadius float64
	magnitude      float64
}

type solarEclipseAxis struct {
	rightAscension float64
	tilt           float64
	gst            float64
}

type solarEclipseSolver struct {
	newMoonJDE float64
	model      SolarEclipseRadiusModel
	params     solarEclipseModelParameters

	meanSunMoonDistance float64
	penumbraConeTangent float64
	umbraConeTangent    float64
}

type solarEclipseFeature struct {
	greatestEclipseJDE float64
	greatestLongitude  float64
	greatestLatitude   float64
	magnitude          float64
	gamma              float64
	pathWidthKM        float64

	partialBeginJDE float64
	partialEndJDE   float64
	centralBeginJDE float64
	centralEndJDE   float64

	typeCode string
}

type solarEclipseLineIntersection struct {
	valid bool
	x     float64
	y     float64
	z     float64
	r1    float64
	r2    float64
}

const (
	solarEclipseEarthEquatorialRadiusKM = 6378.1366
	solarEclipseEarthPolarRatio         = 0.99664719
	solarEclipseEarthPolarRatioSquared  = solarEclipseEarthPolarRatio * solarEclipseEarthPolarRatio
	solarEclipseAstronomicalUnitKM      = 1.49597870691e8

	// IAU Single-K 对所有接触统一使用 0.2725076；
	// NASA bulletin Split-K 对半影仍使用 0.2725076，对本影/反本影使用 0.2722810。
	solarEclipseSolarRadiusRatio = 109.1222
	solarEclipsePenumbralK       = 0.2725076
	solarEclipseUmbralK          = 0.2722810

	solarEclipseNodeCount       = 7
	solarEclipseNodeStepDays    = 0.04
	solarEclipseMoonLonAberrRad = -3.4e-6

	// 这两个系数沿用经典贝塞尔近似中的极区有效半径经验值。
	solarEclipseNonCentralLimit = 0.9972
	solarEclipseCentralLimit    = 0.9966
)

var solarEclipseArcsecPerRadian = 180.0 * 3600.0 / math.Pi

// SolarEclipse 计算给定近朔时刻附近的一次全局日食，默认使用 NASABulletin Split-K 模型。
func SolarEclipse(seedJDE float64) SolarEclipseResult {
	return SolarEclipseNASABulletinSplitK(seedJDE)
}

// SolarEclipseIAUSingleK 计算给定近朔时刻附近的一次全局日食，使用 IAU Single-K 模型。
func SolarEclipseIAUSingleK(seedJDE float64) SolarEclipseResult {
	return solarEclipse(seedJDE, SolarEclipseModelIAUSingleK)
}

// SolarEclipseNASABulletinSplitK 计算给定近朔时刻附近的一次全局日食，使用 NASA bulletin Split-K 模型。
func SolarEclipseNASABulletinSplitK(seedJDE float64) SolarEclipseResult {
	return solarEclipse(seedJDE, SolarEclipseModelNASABulletinSplitK)
}

func solarEclipse(seedJDE float64, model SolarEclipseRadiusModel) SolarEclipseResult {
	newMoonJDE := CalcMoonSHByJDE(seedJDE, 0)
	solver := newSolarEclipseSolver(newMoonJDE, model)
	feature := solver.feature()

	result := SolarEclipseResult{
		Model:             model,
		Type:              SolarEclipseNone,
		Centrality:        SolarEclipseNonCentral,
		GreatestEclipse:   feature.greatestEclipseJDE,
		Magnitude:         feature.magnitude,
		Gamma:             feature.gamma,
		PathWidthKM:       feature.pathWidthKM,
		GreatestLongitude: feature.greatestLongitude,
		GreatestLatitude:  feature.greatestLatitude,
	}

	switch feature.typeCode {
	case "P":
		result.Type = SolarEclipsePartial
	case "A0", "A1", "A":
		result.Type = SolarEclipseAnnular
	case "T0", "T1", "T":
		result.Type = SolarEclipseTotal
	case "H", "H2", "H3":
		result.Type = SolarEclipseHybrid
	}

	switch feature.typeCode {
	case "A1", "T1":
		result.Centrality = SolarEclipseCentralOneLimit
	case "A", "T", "H", "H2", "H3":
		result.Centrality = SolarEclipseCentralTwoLimits
	}

	if result.Type != SolarEclipseNone {
		result.HasPartial = true
		result.PartialBeginOnEarth = feature.partialBeginJDE
		result.PartialEndOnEarth = feature.partialEndJDE
	}

	if result.Centrality != SolarEclipseNonCentral {
		result.HasCentral = true
		result.CentralBeginOnEarth = feature.centralBeginJDE
		result.CentralEndOnEarth = feature.centralEndJDE
	}

	switch result.Type {
	case SolarEclipseAnnular:
		result.HasAnnular = true
	case SolarEclipseTotal:
		result.HasTotal = true
	case SolarEclipseHybrid:
		result.HasAnnular = true
		result.HasTotal = true
		result.HasHybrid = true
	}

	return result
}

func newSolarEclipseSolver(newMoonJDE float64, model SolarEclipseRadiusModel) solarEclipseSolver {
	params := solarEclipseModelParameters{
		penumbralK: solarEclipsePenumbralK,
		umbralK:    solarEclipsePenumbralK,
	}
	if model == SolarEclipseModelNASABulletinSplitK {
		params.umbralK = solarEclipseUmbralK
	}

	firstNodeJDE := newMoonJDE + (0-float64(solarEclipseNodeCount)/2+0.5)*solarEclipseNodeStepDays
	lastNodeJDE := newMoonJDE + (float64(solarEclipseNodeCount-1)-float64(solarEclipseNodeCount)/2+0.5)*solarEclipseNodeStepDays
	firstSun, firstMoon := solarEclipseSunMoonEquatorial(firstNodeJDE)
	lastSun, lastMoon := solarEclipseSunMoonEquatorial(lastNodeJDE)

	meanSunMoonDistance := ((firstSun[2] + lastSun[2]) - (firstMoon[2] + lastMoon[2])) / 2 / solarEclipseEarthEquatorialRadiusKM

	return solarEclipseSolver{
		newMoonJDE:          newMoonJDE,
		model:               model,
		params:              params,
		meanSunMoonDistance: meanSunMoonDistance,
		penumbraConeTangent: (solarEclipseSolarRadiusRatio + params.penumbralK) / meanSunMoonDistance,
		umbraConeTangent:    (solarEclipseSolarRadiusRatio - params.umbralK) / meanSunMoonDistance,
	}
}

func (solver solarEclipseSolver) feature() solarEclipseFeature {
	const finiteDifferenceStep = 0.04

	jd := solver.newMoonJDE
	before := solver.besselMoonAt(jd - finiteDifferenceStep)
	center := solver.besselMoonAt(jd)
	after := solver.besselMoonAt(jd + finiteDifferenceStep)

	vx := (after[0] - before[0]) / (2 * finiteDifferenceStep)
	vy := (after[1] - before[1]) / (2 * finiteDifferenceStep)
	vz := (after[2] - before[2]) / (2 * finiteDifferenceStep)
	speed := math.Hypot(vx, vy)
	speedSquared := speed * speed

	t0 := -(center[0]*vx + center[1]*vy) / speedSquared
	greatestEclipseJDE := jd + t0
	xc := center[0] + vx*t0
	yc := center[1] + vy*t0
	zc := center[2] + vz*t0 - 1.37*t0*t0
	gamma := (vx*center[1] - vy*center[0]) / speed
	minimumDistance := math.Abs(gamma)
	axis := solver.besselAxisAt(greatestEclipseJDE)

	axisIntersection := solarEclipseLineEar2(xc, yc, 2, xc, yc, 0, solarEclipseEarthPolarRatio, 1, axis)

	midRadii := solver.shadowRadiiAt(zc)
	greatestRadii := midRadii
	if axisIntersection.valid {
		greatestRadii = solver.shadowRadiiAt(zc - axisIntersection.r2)
	}

	var centralStartParam, centralEndParam float64
	if minimumDistance < 1 {
		paramSpan := math.Sqrt(1-minimumDistance*minimumDistance) / speed
		centralStartParam = t0 - paramSpan
		centralEndParam = t0 + paramSpan
	}

	partialLimit := 1 + midRadii.penumbraRadius
	partialSpan := 0.0
	if minimumDistance < partialLimit {
		partialSpan = math.Sqrt(partialLimit*partialLimit-minimumDistance*minimumDistance) / speed
	}
	partialStartParam := t0 - partialSpan
	partialEndParam := t0 + partialSpan

	typeCode := "N"
	greatestLongitude, greatestLatitude := 0.0, 0.0
	magnitude := 0.0
	pathWidthKM := 0.0

	if !axisIntersection.valid {
		greatestLongitude, greatestLatitude = solarEclipseBesselPointToGeodetic(xc, yc, 0, axis, false)
		magnitude = (midRadii.penumbraRadius - (minimumDistance - solarEclipseNonCentralLimit)) / (midRadii.penumbraRadius - midRadii.umbraRadius)
		switch {
		case minimumDistance > solarEclipseNonCentralLimit+midRadii.penumbraRadius:
			typeCode = "N"
		case minimumDistance > solarEclipseNonCentralLimit+midRadii.absUmbraRadius:
			typeCode = "P"
		default:
			if midRadii.magnitude < 1 {
				typeCode = "A0"
			} else {
				typeCode = "T0"
			}
		}
	} else {
		greatestLongitude = axisIntersectionLongitude(axisIntersection, axis)
		greatestLatitude = axisIntersectionLatitude(axisIntersection, axis)
		magnitude = greatestRadii.magnitude

		switch {
		case minimumDistance > solarEclipseCentralLimit-greatestRadii.absUmbraRadius:
			if greatestRadii.magnitude < 1 {
				typeCode = "A1"
			} else {
				typeCode = "T1"
			}
		default:
			if greatestRadii.magnitude >= 1 {
				startRadii := greatestRadii
				endRadii := greatestRadii
				if minimumDistance < 1 {
					startRadii = solver.shadowRadiiAt(centralStartParam*vz + center[2] - 1.37*centralStartParam*centralStartParam)
					endRadii = solver.shadowRadiiAt(centralEndParam*vz + center[2] - 1.37*centralEndParam*centralEndParam)
				}
				typeCode = "H"
				if startRadii.magnitude > 1 {
					typeCode = "H2"
				}
				if endRadii.magnitude > 1 {
					typeCode = "H3"
				}
				if startRadii.magnitude > 1 && endRadii.magnitude > 1 {
					typeCode = "T"
				}
			} else {
				typeCode = "A"
			}
		}

		if typeCode != "N" && typeCode != "P" {
			sunAltitude := solarEclipseSunAltitudeAtGreatest(greatestEclipseJDE, greatestLongitude, greatestLatitude, axis.gst)
			if math.Abs(math.Sin(sunAltitude)) > 1e-12 {
				pathWidthKM = math.Abs(2*greatestRadii.umbraRadius*solarEclipseEarthEquatorialRadiusKM) / math.Abs(math.Sin(sunAltitude))
			}
		}
	}

	feature := solarEclipseFeature{
		greatestEclipseJDE: greatestEclipseJDE,
		greatestLongitude:  greatestLongitude,
		greatestLatitude:   greatestLatitude,
		magnitude:          magnitude,
		gamma:              gamma,
		pathWidthKM:        pathWidthKM,
		typeCode:           typeCode,
	}

	if typeCode != "N" {
		_, _, feature.partialBeginJDE, _ = solver.quickContactAt(partialStartParam+jd, vx, vy, true)
		_, _, feature.partialEndJDE, _ = solver.quickContactAt(partialEndParam+jd, vx, vy, true)
	}

	if typeCode != "N" && typeCode != "P" {
		_, _, feature.centralBeginJDE, _ = solver.quickContactAt(centralStartParam+jd, vx, vy, false)
		_, _, feature.centralEndJDE, _ = solver.quickContactAt(centralEndParam+jd, vx, vy, false)
	}

	return feature
}

func (solver solarEclipseSolver) quickContactAt(jd, dx, dy float64, penumbral bool) (float64, float64, float64, bool) {
	moon := solver.besselMoonAt(jd)
	radii := solver.shadowRadiiAt(moon[2])
	radius := 0.0
	if penumbral {
		radius = radii.penumbraRadius
	}

	denominator := moon[0]*moon[0] + moon[1]*moon[1]
	if denominator == 0 {
		return 0, 0, 0, false
	}

	effectiveRadius := 1 - (1/solarEclipseEarthPolarRatioSquared-1)*moon[1]*moon[1]/denominator/2 + radius
	velocityProjection := dx*moon[0] + dy*moon[1]
	if velocityProjection == 0 {
		return 0, 0, 0, false
	}

	correction := (effectiveRadius*effectiveRadius - moon[0]*moon[0] - moon[1]*moon[1]) / (2 * velocityProjection)
	x := moon[0] + correction*dx
	y := moon[1] + correction*dy
	jd += correction

	curvature := (1 - solarEclipseEarthPolarRatioSquared) * radius * x * y / math.Pow(effectiveRadius, 3)
	x += curvature * y
	y -= curvature * x

	axis := solver.besselAxisAt(jd)
	longitude, latitude, ok := solarEclipseBesselXYToGeodetic(x/effectiveRadius, y/effectiveRadius, axis, true)
	return longitude, latitude, jd, ok
}

func (solver solarEclipseSolver) shadowRadiiAt(moonBesselZ float64) solarEclipseShadowRadii {
	return solarEclipseShadowRadii{
		penumbraRadius: solver.params.penumbralK + solver.penumbraConeTangent*moonBesselZ,
		umbraRadius:    solver.params.umbralK - solver.umbraConeTangent*moonBesselZ,
		absUmbraRadius: math.Abs(solver.params.umbralK - solver.umbraConeTangent*moonBesselZ),
		magnitude:      solver.params.umbralK / moonBesselZ / solarEclipseSolarRadiusRatio * (solver.meanSunMoonDistance + moonBesselZ),
	}
}

func (solver solarEclipseSolver) besselAxisAt(jd float64) solarEclipseAxis {
	sun, moon := solarEclipseSunMoonEquatorial(jd)
	sunXYZ := solarEclipseLLRToXYZ(sun[0], sun[1], sun[2])
	moonXYZ := solarEclipseLLRToXYZ(moon[0], moon[1], moon[2])
	axis := solarEclipseXYZToLLR(sunXYZ[0]-moonXYZ[0], sunXYZ[1]-moonXYZ[1], sunXYZ[2]-moonXYZ[2])

	utJDE := TD2UT(jd, false)
	return solarEclipseAxis{
		rightAscension: solarEclipseNormalizeRadians(math.Pi/2 + axis[0]),
		tilt:           math.Pi/2 - axis[1],
		gst:            solarEclipseNormalizeSignedRadians(ApparentSiderealTime(utJDE) * 15 * rad),
	}
}

func (solver solarEclipseSolver) besselMoonAt(jd float64) [3]float64 {
	_, moon := solarEclipseSunMoonEquatorial(jd)
	axis := solver.besselAxisAt(jd)

	rotated := solarEclipseRotateLLR(
		solarEclipseNormalizeSignedRadians(moon[0]-axis.rightAscension),
		moon[1],
		moon[2],
		-axis.tilt,
	)
	rectangular := solarEclipseLLRToXYZ(rotated[0], rotated[1], rotated[2])

	return [3]float64{
		rectangular[0] / solarEclipseEarthEquatorialRadiusKM,
		rectangular[1] / solarEclipseEarthEquatorialRadiusKM,
		rectangular[2] / solarEclipseEarthEquatorialRadiusKM,
	}
}

func solarEclipseSunMoonEquatorial(jd float64) ([3]float64, [3]float64) {
	julianCentury := (jd - 2451545.0) / 36525.0
	obliquity := EclipticObliquity(jd, true) * rad

	sunLongitude := HSunApparentLo(jd) * rad
	sunLatitude := HSunTrueBo(jd) * rad
	sunDistance := EarthAway(jd) * solarEclipseAstronomicalUnitKM

	moonLongitude := solarEclipseNormalizeRadians(HMoonApparentLo(jd)*rad + solarEclipseMoonLonAberrRad)
	moonLatitude := HMoonTrueBo(jd)*rad + moonLatitudeAberrationRad(julianCentury)
	moonDistance := HMoonAway(jd)

	sunEquatorial := solarEclipseRotateLLR(sunLongitude, sunLatitude, sunDistance, obliquity)
	moonEquatorial := solarEclipseRotateLLR(moonLongitude, moonLatitude, moonDistance, obliquity)

	return [3]float64{sunEquatorial[0], sunEquatorial[1], sunEquatorial[2]},
		[3]float64{moonEquatorial[0], moonEquatorial[1], moonEquatorial[2]}
}

func solarEclipseSunAltitudeAtGreatest(jd, lonDeg, latDeg, gst float64) float64 {
	sun, _ := solarEclipseSunMoonEquatorial(jd)
	horizon := solarEclipseEquatorialToHorizontal(sun[0], sun[1], sun[2], lonDeg*rad, latDeg*rad, gst)
	return horizon[1]
}

func solarEclipseEquatorialToHorizontal(ra, dec, distance, lon, lat, gst float64) [3]float64 {
	rotated := solarEclipseRotateLLR(
		solarEclipseNormalizeRadians(ra+math.Pi/2-gst-lon),
		dec,
		distance,
		math.Pi/2-lat,
	)
	return [3]float64{
		solarEclipseNormalizeRadians(math.Pi/2 - rotated[0]),
		rotated[1],
		rotated[2],
	}
}

func solarEclipseLLRToXYZ(longitude, latitude, distance float64) [3]float64 {
	return [3]float64{
		distance * math.Cos(latitude) * math.Cos(longitude),
		distance * math.Cos(latitude) * math.Sin(longitude),
		distance * math.Sin(latitude),
	}
}

func solarEclipseXYZToLLR(x, y, z float64) [3]float64 {
	distance := math.Sqrt(x*x + y*y + z*z)
	return [3]float64{
		solarEclipseNormalizeRadians(math.Atan2(y, x)),
		math.Asin(z / distance),
		distance,
	}
}

func solarEclipseRotateLLR(longitude, latitude, distance, obliquity float64) [3]float64 {
	rotatedLongitude := math.Atan2(
		math.Sin(longitude)*math.Cos(obliquity)-math.Tan(latitude)*math.Sin(obliquity),
		math.Cos(longitude),
	)
	return [3]float64{
		solarEclipseNormalizeRadians(rotatedLongitude),
		math.Asin(math.Cos(obliquity)*math.Sin(latitude) + math.Sin(obliquity)*math.Cos(latitude)*math.Sin(longitude)),
		distance,
	}
}

func solarEclipseLineEar2(x1, y1, z1, x2, y2, z2, polarRatio, radius float64, axis solarEclipseAxis) solarEclipseLineIntersection {
	cosTilt := math.Cos(axis.tilt)
	sinTilt := math.Sin(axis.tilt)
	x1Rot := x1
	y1Rot := cosTilt*y1 - sinTilt*z1
	z1Rot := sinTilt*y1 + cosTilt*z1
	x2Rot := x2
	y2Rot := cosTilt*y2 - sinTilt*z2
	z2Rot := sinTilt*y2 + cosTilt*z2

	intersection := solarEclipseLineEllipsoid(x1Rot, y1Rot, z1Rot, x2Rot, y2Rot, z2Rot, polarRatio, radius)
	if !intersection.valid {
		return intersection
	}

	return intersection
}

func solarEclipseLineEllipsoid(x1, y1, z1, x2, y2, z2, polarRatio, radius float64) solarEclipseLineIntersection {
	dx := x2 - x1
	dy := y2 - y1
	dz := z2 - z1
	polarRatioSquared := polarRatio * polarRatio

	a := dx*dx + dy*dy + dz*dz/polarRatioSquared
	b := x1*dx + y1*dy + z1*dz/polarRatioSquared
	c := x1*x1 + y1*y1 + z1*z1/polarRatioSquared - radius*radius
	discriminant := b*b - a*c
	if discriminant < 0 {
		return solarEclipseLineIntersection{}
	}

	root := math.Sqrt(discriminant)
	if b < 0 {
		root = -root
	}
	t := (-b + root) / a
	x := x1 + dx*t
	y := y1 + dy*t
	z := z1 + dz*t
	distance := math.Sqrt(dx*dx + dy*dy + dz*dz)

	return solarEclipseLineIntersection{
		valid: true,
		x:     x,
		y:     y,
		z:     z,
		r1:    distance * math.Abs(t),
		r2:    distance * math.Abs(t-1),
	}
}

func axisIntersectionLongitude(intersection solarEclipseLineIntersection, axis solarEclipseAxis) float64 {
	longitude, _ := solarEclipseIntersectionGeodetic(intersection, axis)
	return longitude
}

func axisIntersectionLatitude(intersection solarEclipseLineIntersection, axis solarEclipseAxis) float64 {
	_, latitude := solarEclipseIntersectionGeodetic(intersection, axis)
	return latitude
}

func solarEclipseIntersectionGeodetic(intersection solarEclipseLineIntersection, axis solarEclipseAxis) (float64, float64) {
	longitude := solarEclipseNormalizeSignedRadians(math.Atan2(intersection.y, intersection.x) + axis.rightAscension - axis.gst)
	latitude := math.Atan(intersection.z / solarEclipseEarthPolarRatioSquared / math.Sqrt(intersection.x*intersection.x+intersection.y*intersection.y))
	return longitude * deg, latitude * deg
}

func solarEclipseBesselPointToGeodetic(x, y, z float64, axis solarEclipseAxis, ellipsoidal bool) (float64, float64) {
	point := solarEclipseXYZToLLR(x, y, z)
	rotated := solarEclipseRotateLLR(point[0], point[1], point[2], axis.tilt)
	longitude := solarEclipseNormalizeSignedRadians(rotated[0] + axis.rightAscension - axis.gst)
	latitude := rotated[1]
	if ellipsoidal {
		latitude = math.Atan(math.Tan(latitude) / solarEclipseEarthPolarRatioSquared)
	}
	return longitude * deg, latitude * deg
}

func solarEclipseBesselXYToGeodetic(x, y float64, axis solarEclipseAxis, ellipsoidal bool) (float64, float64, bool) {
	polarRatio := 1.0
	if ellipsoidal {
		polarRatio = solarEclipseEarthPolarRatio
	}
	intersection := solarEclipseLineEar2(x, y, 2, x, y, 0, polarRatio, 1, axis)
	if !intersection.valid {
		return 0, 0, false
	}
	longitude, latitude := solarEclipseIntersectionGeodetic(intersection, axis)
	return longitude, latitude, true
}

func solarEclipseNormalizeRadians(angle float64) float64 {
	angle = math.Mod(angle, 2*math.Pi)
	if angle < 0 {
		angle += 2 * math.Pi
	}
	return angle
}

func solarEclipseNormalizeSignedRadians(angle float64) float64 {
	angle = math.Mod(angle, 2*math.Pi)
	if angle <= -math.Pi {
		angle += 2 * math.Pi
	}
	if angle > math.Pi {
		angle -= 2 * math.Pi
	}
	return angle
}
