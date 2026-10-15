package eclipse

import (
	"math"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/ephem"
)

const (
	lunarEclipseSynodicMonthDays = 29.530588853
	lunarEclipseSearchLimit      = 24
	lunarEclipseSearchEpsilonDay = 1e-8
	// 默认口径仍以 Danjon 为主，但对五千年目录边界那类“Danjon 判无食、Chauvenet 判极浅半影食”的个例，
	// 允许在公开默认包装层回退到 Chauvenet，避免把目录中确有记录的边缘半影食整个跳过。
	lunarEclipseDefaultFallbackMaxPenumbralMagnitude = 0.03

	// 当天判断的第一层只做粗筛：
	// 如果本地中午的日月黄经差明显不在满月附近，则这一天不可能发生月食。
	lunarEclipseDayPhaseMin = 120.0
	lunarEclipseDayPhaseMax = 240.0

	// 月食只会发生在月球接近黄道节点时。
	// 这里保守放宽到 2 度，作为“是否值得进入精算”的预筛条件。
	lunarEclipseLatitudeLimitDeg  = 2.0
	lunarEclipseLongitudeLimitDeg = 15.0
)

type lunarEclipseCalculator func(float64) basic.LunarEclipseResult

// LunarEclipseShadowModel 标识月食用的是哪套影半径模型。
// LunarEclipseShadowModel identifies which shadow-radius model produced a result.
type LunarEclipseShadowModel = basic.LunarEclipseShadowModel

const (
	// LunarEclipseShadowModelDanjon 是 Danjon 影半径模型。
	LunarEclipseShadowModelDanjon = basic.LunarEclipseShadowModelDanjon
	// LunarEclipseShadowModelChauvenet 是 Chauvenet 影半径模型。
	LunarEclipseShadowModelChauvenet = basic.LunarEclipseShadowModelChauvenet
)

// LunarEclipseType 月食类型 / lunar eclipse type.
type LunarEclipseType string

const (
	// LunarEclipseNone 无月食, no lunar eclipse.
	LunarEclipseNone LunarEclipseType = "none"
	// LunarEclipsePenumbral 半影月食, penumbral lunar eclipse.
	LunarEclipsePenumbral LunarEclipseType = "penumbral"
	// LunarEclipsePartial 月偏食, partial lunar eclipse.
	LunarEclipsePartial LunarEclipseType = "partial"
	// LunarEclipseTotal 月全食, total lunar eclipse.
	LunarEclipseTotal LunarEclipseType = "total"
)

// LunarEclipseContactPoint 表示月食接触点在月面上的方位。
// LunarEclipseContactPoint describes a lunar eclipse contact point on the Moon limb.
type LunarEclipseContactPoint struct {
	// Label 是接触标签，如 P1/U1/U2/U3/U4/P4。
	// Label is the contact label, such as P1/U1/U2/U3/U4/P4.
	Label string
	// Time 是该接触时刻，保持用户输入时区。
	// Time is the contact time, preserving the input timezone.
	Time time.Time
	// ContactPositionAngle 是月面接触点位置角，从天球北点起向东量，单位度。
	// ContactPositionAngle is the Moon-limb contact position angle from celestial north toward east, in degrees.
	ContactPositionAngle float64
	// ContactClockwiseAngle 是图面上从北点顺时针量到接触点的角度，单位度。
	// ContactClockwiseAngle is the chart clockwise angle from north to the contact point, in degrees.
	ContactClockwiseAngle float64
	// MoonCenterPositionAngle 是月心相对地影中心的位置角，从北点起向东量，单位度。
	// MoonCenterPositionAngle is the Moon-center position angle from the shadow center, in degrees.
	MoonCenterPositionAngle float64
	// ShadowCenterPositionAngle 是地影中心相对月心的位置角，从北点起向东量，单位度。
	// ShadowCenterPositionAngle is the shadow-center position angle from the Moon center, in degrees.
	ShadowCenterPositionAngle float64
}

// LunarEclipseInfo 月食信息, lunar eclipse information.
//
// 所有时刻字段都保持用户输入的时区。
// 不存在的阶段使用零值 time.Time。
type LunarEclipseInfo struct {
	// Type 月食类型, eclipse type.
	Type LunarEclipseType
	// LunarLimbApplied 接触时刻已按月缘剖面修正；剖面失败时为 false 且结果保持圆月面 / contacts refined with a lunar limb profile.
	LunarLimbApplied bool
	// HasSaros 存在沙罗序列信息（可能是锚点外推结果）, has Saros series metadata (possibly extrapolated).
	HasSaros bool
	// Saros 是沙罗序列信息，包括系列号、系列内序号和总成员数。
	// Saros is Saros series metadata with the series number, member index, and total member count.
	Saros SarosInfo

	// PenumbralMagnitude 半影食分, penumbral magnitude.
	PenumbralMagnitude float64
	// UmbralMagnitude 本影食分；纯半影月食时可为负值, umbral magnitude; can be negative for purely penumbral eclipses.
	UmbralMagnitude float64

	// PenumbralStart 半影始, penumbral eclipse begins.
	PenumbralStart time.Time
	// PartialStart 初亏, partial eclipse begins.
	PartialStart time.Time
	// TotalStart 食既, total eclipse begins.
	TotalStart time.Time
	// Maximum 食甚, greatest eclipse.
	Maximum time.Time
	// TotalEnd 生光, total eclipse ends.
	TotalEnd time.Time
	// PartialEnd 复圆, partial eclipse ends.
	PartialEnd time.Time
	// PenumbralEnd 半影终, penumbral eclipse ends.
	PenumbralEnd time.Time
	// GreatestJDE 是食甚的力学时儒略日；与 DeltaTSeconds 配对即可复现上列民用时刻。
	// GreatestJDE is the TT Julian ephemeris day of greatest eclipse; with DeltaTSeconds it reproduces the civil times above.
	GreatestJDE float64
	// DeltaTSeconds 是本次计算实际使用的 TT−UT1，单位秒。
	// DeltaTSeconds is the TT−UT1 used by this computation, in seconds.
	DeltaTSeconds float64

	// ContactPoints 是各接触时刻在月面上的接触点方位。
	// ContactPoints are Moon-limb contact position angles at eclipse contacts.
	ContactPoints []LunarEclipseContactPoint

	// ShadowModel 是本次使用的影半径模型，决定下列影几何与食分的具体数值。
	// ShadowModel is the shadow-radius model used; it fixes the shadow geometry and magnitudes below.
	ShadowModel LunarEclipseShadowModel
	// Gamma 是食甚时月心到地影轴的最小距离，单位地球赤道半径。
	// Gamma is the minimum distance from the Moon's centre to the shadow axis at greatest eclipse, in Earth equatorial radii.
	Gamma float64
	// AxisDegrees 是同一个角距的角度制数值，即 NASA 月食图上的 Axis；Gamma 除以月球处的地球视差即得。
	// AxisDegrees is the same distance in degrees, the Axis column of NASA lunar-eclipse charts; it equals Gamma divided by the Earth's parallax at the Moon.
	AxisDegrees float64
	// PenumbralRadiusDegrees 与 UmbralRadiusDegrees 是食甚时的半影、本影半径，单位度。
	// PenumbralRadiusDegrees and UmbralRadiusDegrees are the shadow radii at greatest eclipse, in degrees.
	PenumbralRadiusDegrees float64
	UmbralRadiusDegrees    float64
	// MoonDistanceEarthRadii 是食甚时的地心月距。
	// MoonDistanceEarthRadii is the geocentric Moon distance at greatest eclipse.
	MoonDistanceEarthRadii float64

	// HasPenumbral 有半影阶段, has penumbral phase.
	HasPenumbral bool
	// HasPartial 有偏食阶段, has partial phase.
	HasPartial bool
	// HasTotal 有全食阶段, has total phase.
	HasTotal bool
}

// LunarEclipseOnDate 当地自然日月食查询 / local-date lunar eclipse query.
// Determine whether a lunar eclipse occurs on the local date.
// The default path uses Danjon and falls back to Chauvenet for ultra-shallow penumbral edge cases.
//
// 只要该自然日内有任意一个接触时刻，或整场月食与该自然日有时间重叠，就返回 true。
func LunarEclipseOnDate(date time.Time) (LunarEclipseInfo, bool) {
	return lunarEclipseOnDateWithFallback(date, basic.LunarEclipseDanjon, true)
}

// LunarEclipseOnDateDanjon 当地自然日月食查询（Danjon） / local-date lunar eclipse query with Danjon model.
// Determine whether a lunar eclipse occurs on the local date with the Danjon model.
func LunarEclipseOnDateDanjon(date time.Time) (LunarEclipseInfo, bool) {
	return lunarEclipseOnDateWithFallback(date, basic.LunarEclipseDanjon, false)
}

// LunarEclipseOnDateChauvenet 当地自然日月食查询（Chauvenet） / local-date lunar eclipse query with Chauvenet model.
// Determine whether a lunar eclipse occurs on the local date with the Chauvenet model.
func LunarEclipseOnDateChauvenet(date time.Time) (LunarEclipseInfo, bool) {
	return lunarEclipseOnDateWithFallback(date, basic.LunarEclipseChauvenet, false)
}

func lunarEclipseOnDateWithFallback(date time.Time, calculator lunarEclipseCalculator, allowDefaultFallback bool) (LunarEclipseInfo, bool) {
	return lunarEclipseOnDateWithInstance(date, calculator, allowDefaultFallback, nil)
}

func lunarEclipseOnDateWithInstance(date time.Time, calculator lunarEclipseCalculator, allowDefaultFallback bool, inst *ephem.Instance) (LunarEclipseInfo, bool) {
	location := date.Location()
	dayStart, dayMid, dayEnd := lunarEclipseLocalDayBounds(date)

	phaseDiff := moonSunLoDiffWithInstance(dayMid, inst)
	if phaseDiff < lunarEclipseDayPhaseMin || phaseDiff > lunarEclipseDayPhaseMax {
		return LunarEclipseInfo{}, false
	}

	candidateTT := basic.CalcMoonSHByJDEWithInstance(timeToTTJDE(dayMid), 1, inst)
	if !isPotentialLunarEclipseWithInstance(candidateTT, inst) {
		return LunarEclipseInfo{}, false
	}

	result := calculator(candidateTT)
	if result.Type == basic.LunarEclipseNone && allowDefaultFallback {
		if fallback, ok := lunarEclipseDefaultFallbackWithInstance(candidateTT, inst); ok {
			result = fallback
		}
	}
	if result.Type == basic.LunarEclipseNone {
		return LunarEclipseInfo{}, false
	}

	info := lunarEclipseInfoFromBasicWithInstance(result, location, inst)
	if !lunarEclipseOverlapsDate(info, dayStart, dayEnd) {
		return LunarEclipseInfo{}, false
	}
	return info, true
}

// LastLunarEclipse 上次月食 / previous lunar eclipse.
// Previous lunar eclipse.
// The default path uses Danjon and falls back to Chauvenet for ultra-shallow penumbral edge cases.
func LastLunarEclipse(date time.Time) LunarEclipseInfo {
	info, _ := searchLunarEclipse(date, -1, true, basic.LunarEclipseDanjon, true)
	return info
}

// LastLunarEclipseDanjon 上次月食（Danjon） / previous lunar eclipse with Danjon model.
// Previous lunar eclipse with the Danjon model.
func LastLunarEclipseDanjon(date time.Time) LunarEclipseInfo {
	info, _ := searchLunarEclipse(date, -1, true, basic.LunarEclipseDanjon, false)
	return info
}

// LastLunarEclipseChauvenet 上次月食（Chauvenet） / previous lunar eclipse with Chauvenet model.
// Previous lunar eclipse with the Chauvenet model.
func LastLunarEclipseChauvenet(date time.Time) LunarEclipseInfo {
	info, _ := searchLunarEclipse(date, -1, true, basic.LunarEclipseChauvenet, false)
	return info
}

// NextLunarEclipse 下次月食 / next lunar eclipse.
// Next lunar eclipse.
// The default path uses Danjon and falls back to Chauvenet for ultra-shallow penumbral edge cases.
func NextLunarEclipse(date time.Time) LunarEclipseInfo {
	info, _ := searchLunarEclipse(date, 1, false, basic.LunarEclipseDanjon, true)
	return info
}

// NextLunarEclipseDanjon 下次月食（Danjon） / next lunar eclipse with Danjon model.
// Next lunar eclipse with the Danjon model.
func NextLunarEclipseDanjon(date time.Time) LunarEclipseInfo {
	info, _ := searchLunarEclipse(date, 1, false, basic.LunarEclipseDanjon, false)
	return info
}

// NextLunarEclipseChauvenet 下次月食（Chauvenet） / next lunar eclipse with Chauvenet model.
// Next lunar eclipse with the Chauvenet model.
func NextLunarEclipseChauvenet(date time.Time) LunarEclipseInfo {
	info, _ := searchLunarEclipse(date, 1, false, basic.LunarEclipseChauvenet, false)
	return info
}

// ClosestLunarEclipse 最近一次月食 / closest lunar eclipse.
// Closest lunar eclipse.
// The default path uses Danjon and falls back to Chauvenet for ultra-shallow penumbral edge cases.
func ClosestLunarEclipse(date time.Time) LunarEclipseInfo {
	last, hasLast := searchLunarEclipse(date, -1, true, basic.LunarEclipseDanjon, true)
	next, hasNext := searchLunarEclipse(date, 1, false, basic.LunarEclipseDanjon, true)
	return closestLunarEclipse(date, last, hasLast, next, hasNext)
}

// ClosestLunarEclipseDanjon 最近一次月食（Danjon） / closest lunar eclipse with Danjon model.
// Closest lunar eclipse with the Danjon model.
func ClosestLunarEclipseDanjon(date time.Time) LunarEclipseInfo {
	last, hasLast := searchLunarEclipse(date, -1, true, basic.LunarEclipseDanjon, false)
	next, hasNext := searchLunarEclipse(date, 1, false, basic.LunarEclipseDanjon, false)

	return closestLunarEclipse(date, last, hasLast, next, hasNext)
}

// ClosestLunarEclipseChauvenet 最近一次月食（Chauvenet） / closest lunar eclipse with Chauvenet model.
// Closest lunar eclipse with the Chauvenet model.
func ClosestLunarEclipseChauvenet(date time.Time) LunarEclipseInfo {
	last, hasLast := searchLunarEclipse(date, -1, true, basic.LunarEclipseChauvenet, false)
	next, hasNext := searchLunarEclipse(date, 1, false, basic.LunarEclipseChauvenet, false)

	return closestLunarEclipse(date, last, hasLast, next, hasNext)
}

func closestLunarEclipse(
	date time.Time,
	last LunarEclipseInfo,
	hasLast bool,
	next LunarEclipseInfo,
	hasNext bool,
) LunarEclipseInfo {
	switch {
	case hasLast && !hasNext:
		return last
	case !hasLast && hasNext:
		return next
	case !hasLast && !hasNext:
		return LunarEclipseInfo{}
	}

	lastDistance := math.Abs(date.Sub(last.Maximum).Seconds())
	nextDistance := math.Abs(next.Maximum.Sub(date).Seconds())
	if lastDistance <= nextDistance {
		return last
	}
	return next
}

func searchLunarEclipse(
	date time.Time,
	direction int,
	includeCurrent bool,
	calculator lunarEclipseCalculator,
	allowDefaultFallback bool,
) (LunarEclipseInfo, bool) {
	return searchLunarEclipseWithInstance(date, direction, includeCurrent, calculator, allowDefaultFallback, nil)
}

func searchLunarEclipseWithInstance(date time.Time, direction int, includeCurrent bool, calculator lunarEclipseCalculator, allowDefaultFallback bool, inst *ephem.Instance) (LunarEclipseInfo, bool) {
	targetTT := timeToTTJDE(date)
	candidateTT := basic.CalcMoonSHByJDEWithInstance(targetTT, 1, inst)

	for i := 0; i < lunarEclipseSearchLimit; i++ {
		if math.IsNaN(candidateTT) || math.IsInf(candidateTT, 0) {
			break
		}
		if isPotentialLunarEclipseWithInstance(candidateTT, inst) {
			result := calculator(candidateTT)
			if result.Type == basic.LunarEclipseNone && allowDefaultFallback {
				if fallback, ok := lunarEclipseDefaultFallbackWithInstance(candidateTT, inst); ok {
					result = fallback
				}
			}
			if result.Type != basic.LunarEclipseNone && lunarEclipseMatchesDirection(result.Maximum, targetTT, direction, includeCurrent) {
				return lunarEclipseInfoFromBasicWithInstance(result, date.Location(), inst), true
			}
		}
		candidateTT = nextLunarEclipseSearchCandidateTTWithInstance(candidateTT, direction, inst)
	}

	return LunarEclipseInfo{}, false
}

func lunarEclipseDefaultFallback(candidateTT float64) (basic.LunarEclipseResult, bool) {
	return lunarEclipseDefaultFallbackWithInstance(candidateTT, nil)
}

func lunarEclipseDefaultFallbackWithInstance(candidateTT float64, inst *ephem.Instance) (basic.LunarEclipseResult, bool) {
	result := basic.LunarEclipseChauvenetWithInstance(candidateTT, inst)
	if result.Type != basic.LunarEclipsePenumbral {
		return basic.LunarEclipseResult{}, false
	}
	if result.HasPartial || result.HasTotal {
		return basic.LunarEclipseResult{}, false
	}
	if result.PenumbralMagnitude <= 0 || result.PenumbralMagnitude > lunarEclipseDefaultFallbackMaxPenumbralMagnitude {
		return basic.LunarEclipseResult{}, false
	}
	return result, true
}

func lunarEclipseMatchesDirection(maximumTT, targetTT float64, direction int, includeCurrent bool) bool {
	delta := maximumTT - targetTT
	if math.Abs(delta) <= lunarEclipseSearchEpsilonDay {
		return direction < 0 && includeCurrent
	}
	if direction > 0 {
		return delta > 0
	}
	return delta < 0
}

func isPotentialLunarEclipse(fullMoonTT float64) bool {
	return isPotentialLunarEclipseWithInstance(fullMoonTT, nil)
}

func isPotentialLunarEclipseWithInstance(fullMoonTT float64, inst *ephem.Instance) bool {
	moonLatitude := math.Abs(basic.HMoonTrueBoWithInstance(fullMoonTT, inst))
	if moonLatitude > lunarEclipseLatitudeLimitDeg {
		return false
	}

	phaseDiff := math.Abs(normalizeDegree180(basic.HMoonApparentLoWithInstance(fullMoonTT, inst) - basic.HSunApparentLoWithInstance(fullMoonTT, inst) - 180))
	return phaseDiff <= lunarEclipseLongitudeLimitDeg
}

func lunarEclipseInfoFromBasic(result basic.LunarEclipseResult, location *time.Location) LunarEclipseInfo {
	return lunarEclipseInfoFromBasicWithInstance(result, location, nil)
}

func lunarEclipseInfoFromBasicWithInstance(result basic.LunarEclipseResult, location *time.Location, inst *ephem.Instance) LunarEclipseInfo {
	saros, hasSaros := lunarSarosInfoWithInstance(result.Maximum, inst)
	geometry := basic.LunarEclipseShadowGeometryAtModelWithInstance(result.Maximum, result.ShadowModel, inst)
	return LunarEclipseInfo{
		ShadowModel:            LunarEclipseShadowModel(result.ShadowModel),
		Gamma:                  geometry.Gamma,
		AxisDegrees:            geometry.AxisDegrees,
		PenumbralRadiusDegrees: geometry.PenumbralRadiusDegrees,
		UmbralRadiusDegrees:    geometry.UmbralRadiusDegrees,
		MoonDistanceEarthRadii: geometry.MoonDistanceEarthRadii,
		HasSaros:               hasSaros,
		Saros:                  saros,
		Type:                   mapBasicLunarEclipseType(result.Type),
		LunarLimbApplied:       result.LunarLimbApplied,
		PenumbralMagnitude:     result.PenumbralMagnitude,
		UmbralMagnitude:        result.Magnitude,
		PenumbralStart:         ttJDEToTime(result.PenumbralStart, location),
		PartialStart:           ttJDEToTime(result.PartialStart, location),
		TotalStart:             ttJDEToTime(result.TotalStart, location),
		Maximum:                ttJDEToTime(result.Maximum, location),
		TotalEnd:               ttJDEToTime(result.TotalEnd, location),
		PartialEnd:             ttJDEToTime(result.PartialEnd, location),
		PenumbralEnd:           ttJDEToTime(result.PenumbralEnd, location),
		GreatestJDE:            result.Maximum,
		DeltaTSeconds:          basic.DeltaT(result.Maximum, true),
		ContactPoints:          lunarEclipseContactPointsFromBasicWithInstance(result, location, inst),
		HasPenumbral:           result.HasPenumbral,
		HasPartial:             result.HasPartial,
		HasTotal:               result.HasTotal,
	}
}

func lunarEclipseContactPointsFromBasic(
	result basic.LunarEclipseResult,
	location *time.Location,
) []LunarEclipseContactPoint {
	return lunarEclipseContactPointsFromBasicWithInstance(result, location, nil)
}

func lunarEclipseContactPointsFromBasicWithInstance(result basic.LunarEclipseResult, location *time.Location, inst *ephem.Instance) []LunarEclipseContactPoint {
	if !result.HasPenumbral {
		return nil
	}

	contacts := []LunarEclipseContactPoint{
		lunarEclipseContactPointWithInstance("P1", result.PenumbralStart, location, false, inst),
	}
	if result.HasPartial {
		contacts = append(contacts, lunarEclipseContactPointWithInstance("U1", result.PartialStart, location, false, inst))
	}
	if result.HasTotal {
		contacts = append(contacts, lunarEclipseContactPointWithInstance("U2", result.TotalStart, location, true, inst))
	}
	if result.HasTotal {
		contacts = append(contacts, lunarEclipseContactPointWithInstance("U3", result.TotalEnd, location, true, inst))
	}
	if result.HasPartial {
		contacts = append(contacts, lunarEclipseContactPointWithInstance("U4", result.PartialEnd, location, false, inst))
	}
	contacts = append(contacts, lunarEclipseContactPointWithInstance("P4", result.PenumbralEnd, location, false, inst))
	return contacts
}

func lunarEclipseContactPoint(
	label string,
	ttJDE float64,
	location *time.Location,
	internalContact bool,
) LunarEclipseContactPoint {
	return lunarEclipseContactPointWithInstance(label, ttJDE, location, internalContact, nil)
}

func lunarEclipseContactPointWithInstance(label string, ttJDE float64, location *time.Location, internalContact bool, inst *ephem.Instance) LunarEclipseContactPoint {
	moonCenterPA := lunarEclipseMoonCenterPositionAngleWithInstance(ttJDE, inst)
	shadowCenterPA := normalizeDegree360(moonCenterPA + 180)
	contactPA := shadowCenterPA
	if internalContact {
		contactPA = moonCenterPA
	}
	return LunarEclipseContactPoint{
		Label:                     label,
		Time:                      ttJDEToTime(ttJDE, location),
		ContactPositionAngle:      contactPA,
		ContactClockwiseAngle:     normalizeDegree360(360 - contactPA),
		MoonCenterPositionAngle:   moonCenterPA,
		ShadowCenterPositionAngle: shadowCenterPA,
	}
}

func lunarEclipseMoonCenterPositionAngle(ttJDE float64) float64 {
	shadowRA, shadowDec := lunarEclipseShadowCenterRaDec(ttJDE)
	moonRA, moonDec := basic.HMoonTrueRaDec(ttJDE)
	return positionAngle(shadowRA, shadowDec, moonRA, moonDec)
}

func lunarEclipseShadowCenterRaDec(ttJDE float64) (float64, float64) {
	sunRA, sunDec := basic.HSunApparentRaDec(ttJDE)
	return normalizeDegree360(sunRA + 180), -sunDec
}

func positionAngle(fromRA, fromDec, toRA, toDec float64) float64 {
	dRA := (toRA - fromRA) * math.Pi / 180
	fromDecRad := fromDec * math.Pi / 180
	toDecRad := toDec * math.Pi / 180
	angle := math.Atan2(
		math.Sin(dRA),
		math.Cos(fromDecRad)*math.Tan(toDecRad)-math.Sin(fromDecRad)*math.Cos(dRA),
	) * 180 / math.Pi
	return normalizeDegree360(angle)
}

func mapBasicLunarEclipseType(eclipseType basic.LunarEclipseType) LunarEclipseType {
	switch eclipseType {
	case basic.LunarEclipsePenumbral:
		return LunarEclipsePenumbral
	case basic.LunarEclipsePartial:
		return LunarEclipsePartial
	case basic.LunarEclipseTotal:
		return LunarEclipseTotal
	default:
		return LunarEclipseNone
	}
}

func lunarEclipseOverlapsDate(info LunarEclipseInfo, dayStart, dayEnd time.Time) bool {
	eventStart, eventEnd, ok := lunarEclipseRange(info)
	if !ok {
		return false
	}
	return !eventEnd.Before(dayStart) && eventStart.Before(dayEnd)
}

func lunarEclipseRange(info LunarEclipseInfo) (time.Time, time.Time, bool) {
	if !info.HasPenumbral {
		return time.Time{}, time.Time{}, false
	}
	return info.PenumbralStart, info.PenumbralEnd, true
}

func ttJDEToTime(ttJDE float64, location *time.Location) time.Time {
	// basic 用 NaN 表示"该阶段不发生"（0 是 −4713-11-24 的真实时刻），
	// 这里必须一并当空值处理，否则 NaN 会被换算成 −296 年那种假时刻。
	if ttJDE == 0 || math.IsNaN(ttJDE) || math.IsInf(ttJDE, 0) {
		return time.Time{}
	}
	utcJD := basic.TT2UTC(ttJDE)
	return basic.JD2DateByZone(utcJD, location, false)
}

func timeToTTJDE(date time.Time) float64 {
	utcJD := basic.Date2JD(date.UTC())
	return basic.UTC2TT(utcJD)
}

func normalizeDegree180(angle float64) float64 {
	angle = math.Mod(angle, 360)
	if angle > 180 {
		angle -= 360
	}
	if angle <= -180 {
		angle += 360
	}
	return angle
}

func normalizeDegree360(angle float64) float64 {
	angle = math.Mod(angle, 360)
	if angle < 0 {
		angle += 360
	}
	return angle
}

func lunarEclipseLocalDayBounds(date time.Time) (time.Time, time.Time, time.Time) {
	location := date.Location()
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, location)
	dayMid := time.Date(date.Year(), date.Month(), date.Day(), 12, 0, 0, 0, location)
	dayEnd := time.Date(date.Year(), date.Month(), date.Day()+1, 0, 0, 0, 0, location)
	return dayStart, dayMid, dayEnd
}

func nextLunarEclipseLocalDayStart(dayStart time.Time) time.Time {
	location := dayStart.Location()
	return time.Date(dayStart.Year(), dayStart.Month(), dayStart.Day()+1, 0, 0, 0, 0, location)
}
