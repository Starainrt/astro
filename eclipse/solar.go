package eclipse

import (
	"math"
	"time"

	"github.com/starainrt/astro/basic"
)

const (
	solarEclipseSynodicMonthDays = 29.530588853
	solarEclipseSearchLimit      = 36
	solarEclipseSearchEpsilonDay = 1e-8
	solarEclipseLatitudeLimitDeg = 2.0
)

type solarEclipseCalculator func(float64) basic.SolarEclipseResult

// SolarEclipseRadiusModel 日食月亮半径模型, lunar radius model for solar eclipses.
type SolarEclipseRadiusModel string

const (
	// SolarEclipseModelIAUSingleK IAU 单一 k 值模型, IAU single-k model.
	SolarEclipseModelIAUSingleK SolarEclipseRadiusModel = "iau_single_k"
	// SolarEclipseModelNASABulletinSplitK NASA bulletin 分裂 k 值模型, NASA bulletin split-k model.
	SolarEclipseModelNASABulletinSplitK SolarEclipseRadiusModel = "nasa_bulletin_split_k"
)

// SolarEclipseSunRadiusModel 日食太阳半径口径, solar radius convention for solar eclipses.
type SolarEclipseSunRadiusModel = basic.SolarEclipseSunRadiusModel

const (
	// SolarEclipseSunRadiusStandard 标准档，1 AU 处 959.639″, standard solar radius for catalogue reproduction.
	SolarEclipseSunRadiusStandard = basic.SolarEclipseSunRadiusStandard
	// SolarEclipseSunRadiusMeasured 边缘档，1 AU 处 959.95″, measured eclipse solar radius.
	SolarEclipseSunRadiusMeasured = basic.SolarEclipseSunRadiusMeasured
)

// SolarEclipseOptions 日食搜索的半径口径, radius conventions for a solar eclipse search.
type SolarEclipseOptions struct {
	// RadiusModel 月亮半径模型，非 IAU Single-K 一律按 NASA bulletin Split-K, lunar radius model.
	RadiusModel SolarEclipseRadiusModel
	// SunRadiusModel 太阳半径口径，零值为标准档, solar radius convention.
	SunRadiusModel SolarEclipseSunRadiusModel
}

// SolarEclipseType 全局日食食型, global solar eclipse type.
type SolarEclipseType string

const (
	// SolarEclipseNone 无日食, no solar eclipse.
	SolarEclipseNone SolarEclipseType = "none"
	// SolarEclipsePartial 日偏食, partial solar eclipse.
	SolarEclipsePartial SolarEclipseType = "partial"
	// SolarEclipseAnnular 日环食, annular solar eclipse.
	SolarEclipseAnnular SolarEclipseType = "annular"
	// SolarEclipseTotal 日全食, total solar eclipse.
	SolarEclipseTotal SolarEclipseType = "total"
	// SolarEclipseHybrid 全环食, hybrid solar eclipse.
	SolarEclipseHybrid SolarEclipseType = "hybrid"
)

// SolarEclipseCentrality 中心线进入地球的方式, global eclipse centrality.
type SolarEclipseCentrality string

const (
	// SolarEclipseNonCentral 非中心食, non-central eclipse.
	SolarEclipseNonCentral SolarEclipseCentrality = "non_central"
	// SolarEclipseCentralOneLimit 单界中心食, central eclipse with one limit.
	SolarEclipseCentralOneLimit SolarEclipseCentrality = "central_one_limit"
	// SolarEclipseCentralTwoLimits 双界中心食, central eclipse with two limits.
	SolarEclipseCentralTwoLimits SolarEclipseCentrality = "central_two_limits"
)

// SolarEclipseInfo 全局日食信息, global solar eclipse information.
//
// 所有时刻字段都保持用户输入的时区。
// 不存在的阶段使用零值 time.Time。
type SolarEclipseInfo struct {
	// Model 日食月亮半径模型, eclipse lunar radius model.
	Model SolarEclipseRadiusModel
	// SunRadiusModel 日食太阳半径口径, solar radius convention.
	SunRadiusModel SolarEclipseSunRadiusModel
	// Type 全局食型, global eclipse type.
	Type SolarEclipseType
	// Centrality 中心性, eclipse centrality.
	Centrality SolarEclipseCentrality
	// HasSaros 存在沙罗序列信息（可能是锚点外推结果）, has Saros series metadata (possibly extrapolated).
	HasSaros bool
	// Saros 是沙罗序列信息，包括系列号、系列内序号和总成员数。
	// Saros is Saros series metadata with the series number, member index, and total member count.
	Saros SarosInfo

	// GreatestEclipse 食甚时刻, greatest eclipse.
	GreatestEclipse time.Time
	// PartialBeginOnEarth 地球范围偏食始, partial eclipse begins on Earth.
	PartialBeginOnEarth time.Time
	// PartialEndOnEarth 地球范围偏食终, partial eclipse ends on Earth.
	PartialEndOnEarth time.Time
	// CentralBeginOnEarth 地球范围中心食始, central eclipse begins on Earth.
	CentralBeginOnEarth time.Time
	// CentralEndOnEarth 地球范围中心食终, central eclipse ends on Earth.
	CentralEndOnEarth time.Time
	// GreatestJDE 是食甚的力学时儒略日；与 DeltaTSeconds 配对即可复现上列民用时刻。
	// GreatestJDE is the TT Julian ephemeris day of greatest eclipse; with DeltaTSeconds it reproduces the civil times above.
	GreatestJDE float64
	// DeltaTSeconds 是本次计算实际使用的 TT−UT1，单位秒。
	// DeltaTSeconds is the TT−UT1 used by this computation, in seconds.
	DeltaTSeconds float64

	// Magnitude 全局食分, global eclipse magnitude.
	Magnitude float64
	// Gamma 食甚时影轴到地心的距离, gamma at greatest eclipse.
	Gamma float64
	// CentralDuration 是食甚点的中心食持续时间，即日食目录（如 NASA「Central Dur.」）
	// 的取值口径；沿整条中心线的最大值见 SolarEclipsePath.MaxCentralDuration。
	// CentralDuration is the central-phase duration at the greatest eclipse, the
	// value catalogues publish; SolarEclipsePath.MaxCentralDuration is the maximum
	// along the whole track.
	CentralDuration time.Duration
	// PathWidthKM 是食甚处中心食带宽度；非中心食为 0；单侧极限（中心带仅触及地球边缘）时该解析式失效
	// 并一并置 0，此时 PathWidthDefined 为 false，NASA 目录该栏印 '-'。
	// PathWidthKM is the central path width at greatest eclipse, 0 for a non-central
	// event, and 0 when the analytic formula fails at a single-sided limit where the
	// band only grazes the Earth's limb; PathWidthDefined is false there and
	// catalogues print '-' for this column.
	PathWidthKM float64
	// PathWidthDefined 表示上面的带宽是否有定义：只有南北两限都存在时才为 true。
	// PathWidthDefined reports whether the width above is defined: it is true only
	// when both band limits exist.
	PathWidthDefined bool

	// GreatestLongitude 食甚点经度，东正西负, longitude of greatest eclipse, east positive.
	GreatestLongitude float64
	// GreatestLatitude 食甚点纬度，北正南负, latitude of greatest eclipse, north positive.
	GreatestLatitude float64

	// HasPartial 存在偏食阶段, has partial phase.
	HasPartial bool
	// HasCentral 存在中心食阶段, has central phase.
	HasCentral bool
	// HasAnnular 存在环食阶段, has annular phase.
	HasAnnular bool
	// HasTotal 存在全食阶段, has total phase.
	HasTotal bool
	// HasHybrid 为混合食, is hybrid eclipse.
	HasHybrid bool
}

// SolarEclipseOnDate 当地自然日全局日食查询 / local-date global solar eclipse query.
// Determine whether a global solar eclipse overlaps the local date, using NASA bulletin Split-K by default.
func SolarEclipseOnDate(date time.Time) (SolarEclipseInfo, bool) {
	return SolarEclipseOnDateNASABulletinSplitK(date)
}

// SolarEclipseOnDateNASABulletinSplitK 当地自然日全局日食查询（NASA bulletin Split-K） / local-date global solar eclipse query with NASA bulletin Split-K.
// Determine whether a global solar eclipse overlaps the local date with the NASA bulletin Split-K model.
func SolarEclipseOnDateNASABulletinSplitK(date time.Time) (SolarEclipseInfo, bool) {
	return solarEclipseOnDate(date, basic.SolarEclipseNASABulletinSplitK)
}

// SolarEclipseOnDateIAUSingleK 当地自然日全局日食查询（IAU Single-K） / local-date global solar eclipse query with IAU Single-K.
// Determine whether a global solar eclipse overlaps the local date with the IAU Single-K model.
func SolarEclipseOnDateIAUSingleK(date time.Time) (SolarEclipseInfo, bool) {
	return solarEclipseOnDate(date, basic.SolarEclipseIAUSingleK)
}

// SolarEclipseOnDateWithOptions 当地自然日全局日食查询（自定义半径口径） / local-date global solar eclipse query with custom radius conventions.
func SolarEclipseOnDateWithOptions(date time.Time, options SolarEclipseOptions) (SolarEclipseInfo, bool) {
	return solarEclipseOnDate(date, solarEclipseCalculatorFor(options))
}

// LastSolarEclipseWithOptions 上次日食（自定义半径口径） / previous solar eclipse with custom radius conventions.
func LastSolarEclipseWithOptions(date time.Time, options SolarEclipseOptions) SolarEclipseInfo {
	info, _ := searchSolarEclipse(date, -1, true, solarEclipseCalculatorFor(options))
	return info
}

// NextSolarEclipseWithOptions 下次日食（自定义半径口径） / next solar eclipse with custom radius conventions.
func NextSolarEclipseWithOptions(date time.Time, options SolarEclipseOptions) SolarEclipseInfo {
	info, _ := searchSolarEclipse(date, 1, false, solarEclipseCalculatorFor(options))
	return info
}

// ClosestSolarEclipseWithOptions 最近一次日食（自定义半径口径） / closest solar eclipse with custom radius conventions.
func ClosestSolarEclipseWithOptions(date time.Time, options SolarEclipseOptions) SolarEclipseInfo {
	last, hasLast := searchSolarEclipse(date, -1, true, solarEclipseCalculatorFor(options))
	next, hasNext := searchSolarEclipse(date, 1, false, solarEclipseCalculatorFor(options))
	return closestSolarEclipse(date, last, hasLast, next, hasNext)
}

func solarEclipseCalculatorFor(options SolarEclipseOptions) solarEclipseCalculator {
	basicOptions := basic.SolarEclipseOptions{
		RadiusModel:    basic.SolarEclipseRadiusModel(options.RadiusModel),
		SunRadiusModel: basic.SolarEclipseSunRadiusModel(options.SunRadiusModel),
	}
	if basicOptions.RadiusModel != basic.SolarEclipseModelIAUSingleK {
		basicOptions.RadiusModel = basic.SolarEclipseModelNASABulletinSplitK
	}
	return func(seed float64) basic.SolarEclipseResult {
		return basic.SolarEclipseWithOptions(seed, basicOptions)
	}
}

func solarEclipseOnDate(date time.Time, calculator solarEclipseCalculator) (SolarEclipseInfo, bool) {
	location := date.Location()
	dayStart, dayMid, dayEnd := solarEclipseLocalDayBounds(date)

	candidateTT := basic.CalcMoonSHByJDE(solarEclipseTimeToTTJDE(dayMid), 0)
	result := calculator(candidateTT)
	if result.Type == basic.SolarEclipseNone {
		return SolarEclipseInfo{}, false
	}

	info := solarEclipseInfoFromBasic(result, location)
	if !solarEclipseOverlapsDate(info, dayStart, dayEnd) {
		return SolarEclipseInfo{}, false
	}
	return info, true
}

// LastSolarEclipse 上次日食 / previous solar eclipse.
// Previous solar eclipse, using NASA bulletin Split-K by default.
func LastSolarEclipse(date time.Time) SolarEclipseInfo {
	return LastSolarEclipseNASABulletinSplitK(date)
}

// LastSolarEclipseNASABulletinSplitK 上次日食（NASA bulletin Split-K） / previous solar eclipse with NASA bulletin Split-K.
// Previous solar eclipse with the NASA bulletin Split-K model.
func LastSolarEclipseNASABulletinSplitK(date time.Time) SolarEclipseInfo {
	info, _ := searchSolarEclipse(date, -1, true, basic.SolarEclipseNASABulletinSplitK)
	return info
}

// LastSolarEclipseIAUSingleK 上次日食（IAU Single-K） / previous solar eclipse with IAU Single-K.
// Previous solar eclipse with the IAU Single-K model.
func LastSolarEclipseIAUSingleK(date time.Time) SolarEclipseInfo {
	info, _ := searchSolarEclipse(date, -1, true, basic.SolarEclipseIAUSingleK)
	return info
}

// NextSolarEclipse 下次日食 / next solar eclipse.
// Next solar eclipse, using NASA bulletin Split-K by default.
func NextSolarEclipse(date time.Time) SolarEclipseInfo {
	return NextSolarEclipseNASABulletinSplitK(date)
}

// NextSolarEclipseNASABulletinSplitK 下次日食（NASA bulletin Split-K） / next solar eclipse with NASA bulletin Split-K.
// Next solar eclipse with the NASA bulletin Split-K model.
func NextSolarEclipseNASABulletinSplitK(date time.Time) SolarEclipseInfo {
	info, _ := searchSolarEclipse(date, 1, false, basic.SolarEclipseNASABulletinSplitK)
	return info
}

// NextSolarEclipseIAUSingleK 下次日食（IAU Single-K） / next solar eclipse with IAU Single-K.
// Next solar eclipse with the IAU Single-K model.
func NextSolarEclipseIAUSingleK(date time.Time) SolarEclipseInfo {
	info, _ := searchSolarEclipse(date, 1, false, basic.SolarEclipseIAUSingleK)
	return info
}

// ClosestSolarEclipse 最近一次日食 / closest solar eclipse.
// Closest solar eclipse, using NASA bulletin Split-K by default.
func ClosestSolarEclipse(date time.Time) SolarEclipseInfo {
	return ClosestSolarEclipseNASABulletinSplitK(date)
}

// ClosestSolarEclipseNASABulletinSplitK 最近一次日食（NASA bulletin Split-K） / closest solar eclipse with NASA bulletin Split-K.
// Closest solar eclipse with the NASA bulletin Split-K model.
func ClosestSolarEclipseNASABulletinSplitK(date time.Time) SolarEclipseInfo {
	last, hasLast := searchSolarEclipse(date, -1, true, basic.SolarEclipseNASABulletinSplitK)
	next, hasNext := searchSolarEclipse(date, 1, false, basic.SolarEclipseNASABulletinSplitK)
	return closestSolarEclipse(date, last, hasLast, next, hasNext)
}

// ClosestSolarEclipseIAUSingleK 最近一次日食（IAU Single-K） / closest solar eclipse with IAU Single-K.
// Closest solar eclipse with the IAU Single-K model.
func ClosestSolarEclipseIAUSingleK(date time.Time) SolarEclipseInfo {
	last, hasLast := searchSolarEclipse(date, -1, true, basic.SolarEclipseIAUSingleK)
	next, hasNext := searchSolarEclipse(date, 1, false, basic.SolarEclipseIAUSingleK)
	return closestSolarEclipse(date, last, hasLast, next, hasNext)
}

func closestSolarEclipse(
	date time.Time,
	last SolarEclipseInfo,
	hasLast bool,
	next SolarEclipseInfo,
	hasNext bool,
) SolarEclipseInfo {
	switch {
	case hasLast && !hasNext:
		return last
	case !hasLast && hasNext:
		return next
	case !hasLast && !hasNext:
		return SolarEclipseInfo{}
	}

	lastDistance := math.Abs(date.Sub(last.GreatestEclipse).Seconds())
	nextDistance := math.Abs(next.GreatestEclipse.Sub(date).Seconds())
	if lastDistance <= nextDistance {
		return last
	}
	return next
}

func searchSolarEclipse(
	date time.Time,
	direction int,
	includeCurrent bool,
	calculator solarEclipseCalculator,
) (SolarEclipseInfo, bool) {
	targetTT := solarEclipseTimeToTTJDE(date)
	candidateTT := basic.CalcMoonSHByJDE(targetTT, 0)

	for i := 0; i < solarEclipseSearchLimit; i++ {
		if isPotentialSolarEclipse(candidateTT) {
			result := calculator(candidateTT)
			if result.Type != basic.SolarEclipseNone && solarEclipseMatchesDirection(result.GreatestEclipse, targetTT, direction, includeCurrent) {
				return solarEclipseInfoFromBasic(result, date.Location()), true
			}
		}
		candidateTT = nextEclipseSearchCandidateTT(candidateTT, 0, direction, solarEclipseSynodicMonthDays)
	}

	return SolarEclipseInfo{}, false
}

func isPotentialSolarEclipse(newMoonTT float64) bool {
	return math.Abs(basic.HMoonTrueBo(newMoonTT)) <= solarEclipseLatitudeLimitDeg
}

func solarEclipseMatchesDirection(greatestTT, targetTT float64, direction int, includeCurrent bool) bool {
	delta := greatestTT - targetTT
	if math.Abs(delta) <= solarEclipseSearchEpsilonDay {
		return direction < 0 && includeCurrent
	}
	if direction > 0 {
		return delta > 0
	}
	return delta < 0
}

func solarEclipseLocalDayBounds(date time.Time) (time.Time, time.Time, time.Time) {
	location := date.Location()
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, location)
	dayMid := time.Date(date.Year(), date.Month(), date.Day(), 12, 0, 0, 0, location)
	dayEnd := time.Date(date.Year(), date.Month(), date.Day()+1, 0, 0, 0, 0, location)
	return dayStart, dayMid, dayEnd
}

func nextSolarEclipseLocalDayStart(dayStart time.Time) time.Time {
	location := dayStart.Location()
	return time.Date(dayStart.Year(), dayStart.Month(), dayStart.Day()+1, 0, 0, 0, 0, location)
}

func solarEclipseInfoFromBasic(result basic.SolarEclipseResult, location *time.Location) SolarEclipseInfo {
	saros, hasSaros := solarSarosInfo(result.GreatestEclipse)
	return SolarEclipseInfo{
		Model:               mapBasicSolarEclipseModel(result.Model),
		SunRadiusModel:      result.SunRadiusModel,
		Type:                mapBasicSolarEclipseType(result.Type),
		Centrality:          mapBasicSolarEclipseCentrality(result.Centrality),
		HasSaros:            hasSaros,
		Saros:               saros,
		GreatestEclipse:     solarEclipseTTJDEToTime(result.GreatestEclipse, location),
		PartialBeginOnEarth: solarEclipseTTJDEToTime(result.PartialBeginOnEarth, location),
		PartialEndOnEarth:   solarEclipseTTJDEToTime(result.PartialEndOnEarth, location),
		CentralBeginOnEarth: solarEclipseTTJDEToTime(result.CentralBeginOnEarth, location),
		CentralEndOnEarth:   solarEclipseTTJDEToTime(result.CentralEndOnEarth, location),
		GreatestJDE:         result.GreatestEclipse,
		DeltaTSeconds:       basic.DeltaT(result.GreatestEclipse, true),
		Magnitude:           result.Magnitude,
		Gamma:               result.Gamma,
		CentralDuration:     solarEclipseDurationFromDays(result.CentralDurationDays),
		PathWidthKM:         result.PathWidthKM,
		PathWidthDefined:    result.PathWidthDefined,
		GreatestLongitude:   result.GreatestLongitude,
		GreatestLatitude:    result.GreatestLatitude,
		HasPartial:          result.HasPartial,
		HasCentral:          result.HasCentral,
		HasAnnular:          result.HasAnnular,
		HasTotal:            result.HasTotal,
		HasHybrid:           result.HasHybrid,
	}
}

func mapBasicSolarEclipseModel(model basic.SolarEclipseRadiusModel) SolarEclipseRadiusModel {
	switch model {
	case basic.SolarEclipseModelIAUSingleK:
		return SolarEclipseModelIAUSingleK
	default:
		return SolarEclipseModelNASABulletinSplitK
	}
}

func mapBasicSolarEclipseType(eclipseType basic.SolarEclipseType) SolarEclipseType {
	switch eclipseType {
	case basic.SolarEclipsePartial:
		return SolarEclipsePartial
	case basic.SolarEclipseAnnular:
		return SolarEclipseAnnular
	case basic.SolarEclipseTotal:
		return SolarEclipseTotal
	case basic.SolarEclipseHybrid:
		return SolarEclipseHybrid
	default:
		return SolarEclipseNone
	}
}

func mapBasicSolarEclipseCentrality(centrality basic.SolarEclipseCentrality) SolarEclipseCentrality {
	switch centrality {
	case basic.SolarEclipseCentralOneLimit:
		return SolarEclipseCentralOneLimit
	case basic.SolarEclipseCentralTwoLimits:
		return SolarEclipseCentralTwoLimits
	default:
		return SolarEclipseNonCentral
	}
}

func solarEclipseOverlapsDate(info SolarEclipseInfo, dayStart, dayEnd time.Time) bool {
	eventStart, eventEnd, ok := solarEclipseRange(info)
	if !ok {
		return false
	}
	return !eventEnd.Before(dayStart) && eventStart.Before(dayEnd)
}

func solarEclipseRange(info SolarEclipseInfo) (time.Time, time.Time, bool) {
	if !info.HasPartial {
		return time.Time{}, time.Time{}, false
	}
	return info.PartialBeginOnEarth, info.PartialEndOnEarth, true
}

func solarEclipseTTJDEToTime(ttJDE float64, location *time.Location) time.Time {
	// 回填的是民用时刻，偏移必须与正向的 UTC2TT 成对（窗口内即闰秒表，窗口外按未来政策）。
	// 用 ΔT（TT−UT1）会让时刻整体偏一个 DUT1，并破坏等时线网格的整刻度对齐。
	if ttJDE == 0 || math.IsNaN(ttJDE) || math.IsInf(ttJDE, 0) {
		return time.Time{}
	}
	return basic.JD2DateByZone(basic.TT2UTC(ttJDE), location, false)
}

func solarEclipseTimeToTTJDE(date time.Time) float64 {
	utcJD := basic.Date2JD(date.UTC())
	return basic.UTC2TT(utcJD)
}
