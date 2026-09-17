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

// SolarEclipseType 整场日食的全局食型。
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
	// CentralDurationDays 是食甚点的中心食持续时间，单位为日；没有中心食时为 0。
	// 这是日食目录（如 NASA「Central Dur.」）采用的口径：食甚点的中心食时长。
	// CentralDurationDays is the central-phase duration at the greatest eclipse,
	// in days, and 0 when the event has no central phase.
	CentralDurationDays float64
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

	localStateContextCache map[uint64]localSolarEclipseStateContext
	localEphemeris         *solarEclipseLocalEphemeris
	// deltaTSeconds 是调用方显式给出的 ΔT（秒）；NaN 表示未覆盖，用进程级模型。
	// 只影响地球自转相位（轴的 gst），不改变任何 TT 时刻。
	deltaTSeconds        float64
	besselGeometryCache  map[uint64]solarEclipseBesselGeometryCacheEntry
	besselCandidateCache map[uint64]solarEclipseBesselGeometryCacheEntry
	exactCentralContact  bool

	meanSunMoonDistance float64
	penumbraConeTangent float64
	umbraConeTangent    float64
}

const solarEclipseBesselGeometryCacheMaximumEntries = movingDiskEventCacheMaximumEntries

// solarEclipseBesselGeometryCacheEntry keeps exact and candidate geometry in
// separate maps. Candidate geometry is interpolated and is only suitable for
// coarse scans; mixing it with exact geometry would silently reduce contact
// and topology accuracy.
type solarEclipseBesselGeometryCacheEntry struct {
	// generation 记录写入时的 ΔT 世代：轴里的 gst 由 ΔT 决定，ΔT 覆盖后条目必须失效。
	// generation is the ΔT generation at write time: the axis carries a ΔT-dependent
	// gst, so overriding ΔT has to invalidate the entry.
	generation uint64
	moon       [3]float64
	axis       solarEclipseAxis
	sun        [3]float64
	valid      bool
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
	// 赤道自转线速度，用于把 ΔT 误差换算成地面横移（见 DeltaTGroundShiftKM）。
	// Equatorial rotation speed, used to convert a ΔT error into ground displacement.
	solarEclipseEarthEquatorialRotationKMPerSecond = 0.4651
	solarEclipseEarthPolarRatio                    = 0.99664719
	solarEclipseEarthPolarRatioSquared             = solarEclipseEarthPolarRatio * solarEclipseEarthPolarRatio
	solarEclipseAstronomicalUnitKM                 = 1.49597870691e8

	// IAU Single-K 对所有接触统一使用 0.2725076；
	// NASA bulletin Split-K 对半影仍使用 0.2725076，对本影/反本影使用 0.2722810。
	solarEclipseSolarRadiusRatio = 109.1222
	solarEclipsePenumbralK       = 0.2725076
	solarEclipseUmbralK          = 0.2722810
	// SolarEclipsePenumbralK 与 SolarEclipseUmbralK 是月面半径与地球赤道半径之比，
	// 即 NASA 星历表里的 k1（半影）与 k2（本影/反本影）；IAU Single-K 两者都用 k1。
	// SolarEclipsePenumbralK and SolarEclipseUmbralK are the lunar-to-terrestrial radius ratios
	// published as k1 (penumbra) and k2 (umbra/antumbra); IAU Single-K uses k1 for both.
	SolarEclipsePenumbralK = solarEclipsePenumbralK
	SolarEclipseUmbralK    = solarEclipseUmbralK

	solarEclipseNodeCount                  = 7
	solarEclipseNodeStepDays               = 0.04
	solarEclipseMoonLonAberrRad            = -3.4e-6
	solarEclipseAxisContactInitialStepDays = 1.0 / 86400.0
	solarEclipseAxisContactMaximumStepDays = 30.0 / 1440.0
	solarEclipseAxisContactToleranceDays   = 1e-9

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
	return solarEclipseWithDeltaT(seedJDE, model, 0)
}

func solarEclipseWithDeltaT(
	seedJDE float64,
	model SolarEclipseRadiusModel,
	deltaTSeconds float64,
) SolarEclipseResult {
	newMoonJDE := CalcMoonSHByJDE(seedJDE, 0)
	solver := newSolarEclipseSolver(newMoonJDE, model).withDeltaTSeconds(deltaTSeconds)
	return solver.eclipseResult()
}

func (solver solarEclipseSolver) eclipseResult() SolarEclipseResult {
	model := solver.model
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
		result.CentralDurationDays = solver.greatestCentralDuration(result)
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

// greatestCentralDuration 在食甚点解一次站心中心食并返回中心相时长（日）；没有中心相时为 0。
// greatestCentralDuration solves the local eclipse at the greatest eclipse point and
// returns its central-phase duration in days.
func (solver solarEclipseSolver) greatestCentralDuration(result SolarEclipseResult) float64 {
	if !result.HasCentral || result.GreatestEclipse <= 0 {
		return 0
	}
	return solver.centralPhaseDurationDaysAt(
		result.GreatestEclipse, result.GreatestLongitude, result.GreatestLatitude,
	)
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
		newMoonJDE:             newMoonJDE,
		model:                  model,
		params:                 params,
		deltaTSeconds:          math.NaN(),
		localStateContextCache: make(map[uint64]localSolarEclipseStateContext),
		besselGeometryCache:    make(map[uint64]solarEclipseBesselGeometryCacheEntry),
		besselCandidateCache:   make(map[uint64]solarEclipseBesselGeometryCacheEntry),
		meanSunMoonDistance:    meanSunMoonDistance,
		penumbraConeTangent:    (solarEclipseSolarRadiusRatio + params.penumbralK) / meanSunMoonDistance,
		umbraConeTangent:       (solarEclipseSolarRadiusRatio - params.umbralK) / meanSunMoonDistance,
	}
}

// withDeltaTSeconds 固定本求解器使用的 ΔT（秒），非正值表示回到进程级模型。覆盖会改变
// 轴里的 gst，因此所有按精确 float 位键控的几何缓存必须同时作废。
func (solver solarEclipseSolver) withDeltaTSeconds(deltaTSeconds float64) solarEclipseSolver {
	if deltaTSeconds <= 0 || math.IsNaN(deltaTSeconds) || math.IsInf(deltaTSeconds, 0) {
		deltaTSeconds = math.NaN()
	}
	solver.deltaTSeconds = deltaTSeconds
	solver.besselGeometryCache = make(map[uint64]solarEclipseBesselGeometryCacheEntry)
	solver.besselCandidateCache = make(map[uint64]solarEclipseBesselGeometryCacheEntry)
	solver.localStateContextCache = make(map[uint64]localSolarEclipseStateContext)
	return solver
}

// effectiveDeltaTSeconds 返回本求解器在某 TT 时刻实际使用的 ΔT（秒）。
func (solver solarEclipseSolver) effectiveDeltaTSeconds(jd float64) float64 {
	if math.IsNaN(solver.deltaTSeconds) {
		return DeltaT(jd, true)
	}
	return solver.deltaTSeconds
}

// siderealTimeAt 返回某 TT 时刻的视恒星时（弧度），ΔT 覆盖时同样生效。
func (solver solarEclipseSolver) siderealTimeAt(jd float64) float64 {
	utJDE := TD2UT(jd, false)
	if !math.IsNaN(solver.deltaTSeconds) {
		utJDE = jd - solver.deltaTSeconds/86400
	}
	return ApparentSiderealTime(utJDE) * 15 * rad
}

// withLocalEphemeris prepares the immutable event-local interpolator used by
// coarse candidate scans. The exact ephemeris remains the fallback outside its
// bounded window and is used by all contact and topology refinements.
func (solver solarEclipseSolver) withLocalEphemeris() solarEclipseSolver {
	if solver.localEphemeris == nil {
		solver.localEphemeris = newSolarEclipseLocalEphemeris(solver.newMoonJDE)
	}
	return solver
}

func (solver solarEclipseSolver) feature() solarEclipseFeature {
	const finiteDifferenceStep = 0.04
	candidateSolver := solver.withLocalEphemeris()

	jd := solver.newMoonJDE
	before := candidateSolver.besselMoonCandidateAt(jd - finiteDifferenceStep)
	center := candidateSolver.besselMoonCandidateAt(jd)
	after := candidateSolver.besselMoonCandidateAt(jd + finiteDifferenceStep)

	vx := (after[0] - before[0]) / (2 * finiteDifferenceStep)
	vy := (after[1] - before[1]) / (2 * finiteDifferenceStep)
	vz := (after[2] - before[2]) / (2 * finiteDifferenceStep)
	speed := math.Hypot(vx, vy)
	speedSquared := speed * speed

	t0 := -(center[0]*vx + center[1]*vy) / speedSquared
	greatestEclipseJDE := jd + t0
	// The three-node velocity fit locates greatest eclipse accurately, but its
	// linearly extrapolated coordinates can miss the true Bessel position by
	// tens of kilometres in a grazing non-central event. Re-evaluate the
	// ephemeris at the solved time before deriving surface coordinates and
	// shadow radii so markers and path geometry use the same state.
	greatestMoon := solver.besselMoonAt(greatestEclipseJDE)
	xc := greatestMoon[0]
	yc := greatestMoon[1]
	zc := greatestMoon[2]
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

	if axisIntersection.valid && typeCode != "N" && typeCode != "P" {
		_, _, feature.centralBeginJDE, _ = solver.quickContactAt(centralStartParam+jd, vx, vy, false)
		_, _, feature.centralEndJDE, _ = solver.quickContactAt(centralEndParam+jd, vx, vy, false)
		if refined, ok := solver.centralAxisContactJDE(feature.centralBeginJDE, greatestEclipseJDE, -1); ok {
			feature.centralBeginJDE = refined
		}
		if refined, ok := solver.centralAxisContactJDE(feature.centralEndJDE, greatestEclipseJDE, 1); ok {
			feature.centralEndJDE = refined
		}
	}

	return feature
}

func (solver solarEclipseSolver) centralAxisContactJDE(
	approximateJDE, greatestJDE, direction float64,
) (float64, bool) {
	if !finite(approximateJDE) || !finite(greatestJDE) || direction == 0 {
		return 0, false
	}
	insideJDE, insideResidual := approximateJDE, solver.centralAxisEarthDiscriminant(approximateJDE)
	if !finite(insideResidual) || insideResidual < 0 {
		insideJDE = greatestJDE
		insideResidual = solver.centralAxisEarthDiscriminant(insideJDE)
		if !finite(insideResidual) || insideResidual < 0 {
			return 0, false
		}
	}

	outsideJDE, outsideResidual := 0.0, 0.0
	foundOutside := false
	for step := solarEclipseAxisContactInitialStepDays; step <= solarEclipseAxisContactMaximumStepDays; step *= 2 {
		candidateJDE := approximateJDE + direction*step
		candidateResidual := solver.centralAxisEarthDiscriminant(candidateJDE)
		if !finite(candidateResidual) {
			continue
		}
		if candidateResidual <= 0 {
			outsideJDE, outsideResidual = candidateJDE, candidateResidual
			foundOutside = true
			break
		}
		insideJDE, insideResidual = candidateJDE, candidateResidual
	}
	if !foundOutside {
		return 0, false
	}

	for iteration := 0; iteration < 24 && math.Abs(outsideJDE-insideJDE) > solarEclipseAxisContactToleranceDays; iteration++ {
		candidateJDE := (insideJDE + outsideJDE) / 2
		denominator := insideResidual - outsideResidual
		if denominator != 0 {
			fraction := insideResidual / denominator
			if fraction > 0.1 && fraction < 0.9 {
				candidateJDE = insideJDE + fraction*(outsideJDE-insideJDE)
			}
		}
		candidateResidual := solver.centralAxisEarthDiscriminant(candidateJDE)
		if !finite(candidateResidual) {
			return 0, false
		}
		if candidateResidual >= 0 {
			insideJDE, insideResidual = candidateJDE, candidateResidual
		} else {
			outsideJDE, outsideResidual = candidateJDE, candidateResidual
		}
	}
	return (insideJDE + outsideJDE) / 2, true
}

func (solver solarEclipseSolver) centralAxisEarthDiscriminant(jd float64) float64 {
	moon, axis, _ := solver.besselGeometryAt(jd)
	return solarEclipseLineEllipsoidDiscriminant(
		moon[0], moon[1], 2,
		moon[0], moon[1], 0,
		solarEclipseEarthPolarRatio, 1, axis,
	)
}

func (solver solarEclipseSolver) centralAxisContactPointAt(jd float64) (SolarEclipsePathPoint, bool) {
	moon, axis, _ := solver.besselGeometryAt(jd)
	cosTilt, sinTilt := math.Cos(axis.tilt), math.Sin(axis.tilt)
	x1 := moon[0]
	y1 := cosTilt*moon[1] - 2*sinTilt
	z1 := sinTilt*moon[1] + 2*cosTilt
	x2 := moon[0]
	y2 := cosTilt * moon[1]
	z2 := sinTilt * moon[1]
	dx, dy, dz := x2-x1, y2-y1, z2-z1
	polarRatioSquared := solarEclipseEarthPolarRatioSquared
	a := dx*dx + dy*dy + dz*dz/polarRatioSquared
	if !finite(a) || a <= 0 {
		return SolarEclipsePathPoint{}, false
	}
	b := x1*dx + y1*dy + z1*dz/polarRatioSquared
	t := -b / a
	intersection := solarEclipseLineIntersection{
		valid: true,
		x:     x1 + dx*t,
		y:     y1 + dy*t,
		z:     z1 + dz*t,
	}
	longitude, latitude := solarEclipseIntersectionGeodetic(intersection, axis)
	if !finite(longitude) || !finite(latitude) {
		return SolarEclipsePathPoint{}, false
	}
	return SolarEclipsePathPoint{
		JDE:         jd,
		Longitude:   longitude,
		Latitude:    latitude,
		SunAltitude: solarEclipseSunAltitudeAtGreatest(jd, longitude, latitude, axis.gst) / rad,
	}, true
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
	return solarEclipseBesselAxisFromEquatorialWithDeltaT(
		jd, sun, moon, solver.effectiveDeltaTSeconds(jd),
	)
}

func solarEclipseBesselAxisFromEquatorial(jd float64, sun, moon [3]float64) solarEclipseAxis {
	return solarEclipseBesselAxisFromEquatorialWithDeltaT(jd, sun, moon, DeltaT(jd, true))
}

// solarEclipseBesselAxisFromEquatorialWithDeltaT 用显式 ΔT 构造贝塞尔轴：TT 时刻保持
// 不变，ΔT 只决定地球自转相位（恒星时），因此同一 TT 在不同 ΔT 下得到的地面足迹会
// 沿经度平移，这正是"ΔT 只影响自转、不影响几何时刻"的实现点。
// solarEclipseBesselAxisFromEquatorialWithDeltaT builds the Besselian axis with an
// explicit ΔT: the TT instant is untouched and ΔT only sets Earth rotation.
func solarEclipseBesselAxisFromEquatorialWithDeltaT(
	jd float64, sun, moon [3]float64, deltaTSeconds float64,
) solarEclipseAxis {
	sunXYZ := solarEclipseLLRToXYZ(sun[0], sun[1], sun[2])
	moonXYZ := solarEclipseLLRToXYZ(moon[0], moon[1], moon[2])
	axis := solarEclipseXYZToLLR(sunXYZ[0]-moonXYZ[0], sunXYZ[1]-moonXYZ[1], sunXYZ[2]-moonXYZ[2])

	utJDE := jd - deltaTSeconds/86400
	return solarEclipseAxis{
		rightAscension: solarEclipseNormalizeRadians(math.Pi/2 + axis[0]),
		tilt:           math.Pi/2 - axis[1],
		gst:            solarEclipseNormalizeSignedRadians(ApparentSiderealTime(utJDE) * 15 * rad),
	}
}

func (solver solarEclipseSolver) besselMoonAt(jd float64) [3]float64 {
	moon, _, _ := solver.besselGeometryAt(jd)
	return moon
}

func (solver solarEclipseSolver) besselMoonCandidateAt(jd float64) [3]float64 {
	moon, _, _, ok := solver.besselGeometryCandidateAt(jd)
	if !ok {
		return solver.besselMoonAt(jd)
	}
	return moon
}

func (solver solarEclipseSolver) besselGeometryAt(jd float64) ([3]float64, solarEclipseAxis, [3]float64) {
	key := math.Float64bits(jd)
	// 命中要求 ΔT 世代一致：轴里的 gst 依赖 ΔT，SetDeltaTFn 之后旧条目必须视为未命中。
	// A hit requires the same ΔT generation: the cached axis carries a ΔT-dependent gst,
	// so entries written before a SetDeltaTFn override must count as misses.
	if entry, ok := solver.besselGeometryCache[key]; ok && entry.generation == deltaTGenerationValue() {
		return entry.moon, entry.axis, entry.sun
	}
	sun, moon := solarEclipseSunMoonEquatorial(jd)
	axis := solarEclipseBesselAxisFromEquatorialWithDeltaT(
		jd, sun, moon, solver.effectiveDeltaTSeconds(jd),
	)
	geometry := solarEclipseBesselGeometryCacheEntry{
		moon:  solarEclipseBesselMoonFromEquatorial(moon, axis),
		axis:  axis,
		sun:   sun,
		valid: true,
	}
	storeSolarEclipseBesselGeometry(solver.besselGeometryCache, key, geometry)
	return geometry.moon, geometry.axis, geometry.sun
}

func (solver solarEclipseSolver) besselGeometryCandidateAt(jd float64) ([3]float64, solarEclipseAxis, [3]float64, bool) {
	key := math.Float64bits(jd)
	if entry, ok := solver.besselCandidateCache[key]; ok && entry.generation == deltaTGenerationValue() {
		return entry.moon, entry.axis, entry.sun, entry.valid
	}
	if solver.localEphemeris == nil {
		return [3]float64{}, solarEclipseAxis{}, [3]float64{}, false
	}
	sun, moon, ok := solver.localEphemeris.equatorialAt(jd)
	if !ok {
		return [3]float64{}, solarEclipseAxis{}, [3]float64{}, false
	}
	axis := solarEclipseBesselAxisFromEquatorialWithDeltaT(
		jd, sun, moon, solver.effectiveDeltaTSeconds(jd),
	)
	geometry := solarEclipseBesselGeometryCacheEntry{
		moon:  solarEclipseBesselMoonFromEquatorial(moon, axis),
		axis:  axis,
		sun:   sun,
		valid: true,
	}
	storeSolarEclipseBesselGeometry(solver.besselCandidateCache, key, geometry)
	return geometry.moon, geometry.axis, geometry.sun, geometry.valid
}

func storeSolarEclipseBesselGeometry(
	cache map[uint64]solarEclipseBesselGeometryCacheEntry,
	key uint64,
	entry solarEclipseBesselGeometryCacheEntry,
) {
	if cache == nil {
		return
	}
	if _, exists := cache[key]; !exists && len(cache) >= solarEclipseBesselGeometryCacheMaximumEntries {
		for cachedKey := range cache {
			delete(cache, cachedKey)
		}
	}
	entry.generation = deltaTGenerationValue()
	cache[key] = entry
}

func solarEclipseBesselMoonFromEquatorial(moon [3]float64, axis solarEclipseAxis) [3]float64 {
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
	nutationLongitude, nutationObliquity := Nutation2000B(jd)
	obliquity := (Obliquity1980(jd) + nutationObliquity) * rad

	// Share the full-series distance and nutation for this single TT.
	sunDistanceAU := EarthAway(jd)
	sunLongitude := (HSunTrueLoN(jd, -1) + nutationLongitude - 20.49552/sunDistanceAU/3600) * rad
	sunLatitude := HSunTrueBo(jd) * rad
	sunDistance := sunDistanceAU * solarEclipseAstronomicalUnitKM

	moonLongitude := solarEclipseNormalizeRadians((HMoonTrueLoN(jd, -1)+nutationLongitude)*rad + solarEclipseMoonLonAberrRad)
	moonLatitude := HMoonTrueBo(jd)*rad + moonLatitudeAberrationRad(julianCentury)
	moonDistance := HMoonAway(jd)

	sunEquatorial := solarEclipseRotateLLR(sunLongitude, sunLatitude, sunDistance, obliquity)
	moonEquatorial := solarEclipseRotateLLR(moonLongitude, moonLatitude, moonDistance, obliquity)

	return [3]float64{sunEquatorial[0], sunEquatorial[1], sunEquatorial[2]},
		[3]float64{moonEquatorial[0], moonEquatorial[1], moonEquatorial[2]}
}

func solarEclipseSunAltitudeAtGreatest(jd, lonDeg, latDeg, gst float64) float64 {
	sun, _ := solarEclipseSunMoonEquatorial(jd)
	return solarEclipseSunAltitudeFromEquatorial(sun, lonDeg, latDeg, gst)
}

func solarEclipseSunAltitudeFromEquatorial(sun [3]float64, lonDeg, latDeg, gst float64) float64 {
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
