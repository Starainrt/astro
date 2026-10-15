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
	// LunarLimb 月缘剖面，nil 沿用模块设置；修正站心接触、食分与遮蔽率，以及全球始末与食甚点中心食时长 / lunar limb profile for local contacts and values, global P1/P4 and central duration.
	LunarLimb LunarLimb
}

// LunarLimb 月缘剖面接口，内置实现见 github.com/starainrt/astro/moon/limb, lunar limb profile interface.
type LunarLimb = basic.LunarLimb

// SolarEclipseLimbContact 月缘修正后的全球偏食始末接触点位置 / limb-corrected global contact location.
type SolarEclipseLimbContact = basic.SolarEclipseLimbContact

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
	// LunarLimbApplied 全球偏食始末与食甚点中心食时长已按月缘剖面修正；剖面失败时为 false 且结果保持圆月面 / contacts refined with a lunar limb profile.
	LunarLimbApplied bool
	// PartialBeginContact / PartialEndContact 月缘修正后的偏食始末接触点位置，未修正时 Defined 为 false / limb-corrected contact locations.
	PartialBeginContact SolarEclipseLimbContact
	PartialEndContact   SolarEclipseLimbContact
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
func (module *Eclipse) SolarEclipseOnDate(date time.Time) (SolarEclipseInfo, bool) {
	return module.SolarEclipseOnDateNASABulletinSplitK(date)
}

// SolarEclipseOnDateNASABulletinSplitK 当地自然日全局日食查询（NASA bulletin Split-K） / local-date global solar eclipse query with NASA bulletin Split-K.
func (module *Eclipse) SolarEclipseOnDateNASABulletinSplitK(date time.Time) (SolarEclipseInfo, bool) {
	return solarEclipseOnDate(date, func(seedJDE float64) basic.SolarEclipseResult {
		return module.solarResult(seedJDE, basic.SolarEclipseModelNASABulletinSplitK)
	}, module)
}

// SolarEclipseOnDateIAUSingleK 当地自然日全局日食查询（IAU Single-K） / local-date global solar eclipse query with IAU Single-K.
func (module *Eclipse) SolarEclipseOnDateIAUSingleK(date time.Time) (SolarEclipseInfo, bool) {
	return solarEclipseOnDate(date, func(seedJDE float64) basic.SolarEclipseResult {
		return module.solarResult(seedJDE, basic.SolarEclipseModelIAUSingleK)
	}, module)
}

// SolarEclipseOnDateWithOptions 当地自然日全局日食查询（自定义半径口径） / local-date global solar eclipse query with custom radius conventions.
func (module *Eclipse) SolarEclipseOnDateWithOptions(date time.Time, options SolarEclipseOptions) (SolarEclipseInfo, bool) {
	return solarEclipseOnDate(date, solarEclipseCalculatorFor(options, module), module)
}

// LastSolarEclipseWithOptions 上次日食（自定义半径口径） / previous solar eclipse with custom radius conventions.
func (module *Eclipse) LastSolarEclipseWithOptions(date time.Time, options SolarEclipseOptions) SolarEclipseInfo {
	info, _ := searchSolarEclipse(date, -1, true, solarEclipseCalculatorFor(options, module), module)
	return info
}

// NextSolarEclipseWithOptions 下次日食（自定义半径口径） / next solar eclipse with custom radius conventions.
func (module *Eclipse) NextSolarEclipseWithOptions(date time.Time, options SolarEclipseOptions) SolarEclipseInfo {
	info, _ := searchSolarEclipse(date, 1, false, solarEclipseCalculatorFor(options, module), module)
	return info
}

// ClosestSolarEclipseWithOptions 最近一次日食（自定义半径口径） / closest solar eclipse with custom radius conventions.
func (module *Eclipse) ClosestSolarEclipseWithOptions(date time.Time, options SolarEclipseOptions) SolarEclipseInfo {
	last, hasLast := searchSolarEclipse(date, -1, true, solarEclipseCalculatorFor(options, module), module)
	next, hasNext := searchSolarEclipse(date, 1, false, solarEclipseCalculatorFor(options, module), module)
	return closestSolarEclipse(date, last, hasLast, next, hasNext)
}

func solarEclipseCalculatorFor(options SolarEclipseOptions, modules ...*Eclipse) solarEclipseCalculator {
	module := &defaultModule
	if len(modules) > 0 {
		module = modules[0]
	}

	basicOptions := basic.SolarEclipseOptions{
		RadiusModel:    basic.SolarEclipseRadiusModel(options.RadiusModel),
		SunRadiusModel: basic.SolarEclipseSunRadiusModel(options.SunRadiusModel),
	}
	if basicOptions.RadiusModel != basic.SolarEclipseModelIAUSingleK {
		basicOptions.RadiusModel = basic.SolarEclipseModelNASABulletinSplitK
	}
	basicOptions.LunarLimb = options.LunarLimb
	if basicOptions.LunarLimb == nil {
		basicOptions.LunarLimb = module.lunarLimb
	}
	return func(seed float64) basic.SolarEclipseResult {
		return basic.SolarEclipseWithOptionsWithInstance(seed, basicOptions, module.inst)
	}
}

// solarResult 按模块星历与月缘计算一次全球日食；未设置月缘时与包级计算相同。
func (module *Eclipse) solarResult(seedJDE float64, model basic.SolarEclipseRadiusModel) basic.SolarEclipseResult {
	if module.lunarLimb == nil {
		if model == basic.SolarEclipseModelIAUSingleK {
			return basic.SolarEclipseIAUSingleKWithInstance(seedJDE, module.inst)
		}
		return basic.SolarEclipseNASABulletinSplitKWithInstance(seedJDE, module.inst)
	}
	return basic.SolarEclipseWithOptionsWithInstance(seedJDE,
		basic.SolarEclipseOptions{RadiusModel: model, LunarLimb: module.lunarLimb}, module.inst)
}

func solarEclipseOnDate(date time.Time, calculator solarEclipseCalculator, modules ...*Eclipse) (SolarEclipseInfo, bool) {
	module := &defaultModule
	if len(modules) > 0 {
		module = modules[0]
	}

	location := date.Location()
	dayStart, dayMid, dayEnd := solarEclipseLocalDayBounds(date)

	candidateTT := basic.CalcMoonSHByJDEWithInstance(solarEclipseTimeToTTJDE(dayMid), 0, module.inst)
	result := calculator(candidateTT)
	if result.Type == basic.SolarEclipseNone {
		return SolarEclipseInfo{}, false
	}

	info := solarEclipseInfoFromBasic(result, location, module)
	if !solarEclipseOverlapsDate(info, dayStart, dayEnd) {
		return SolarEclipseInfo{}, false
	}
	return info, true
}

// LastSolarEclipse 上次日食 / previous solar eclipse.
func (module *Eclipse) LastSolarEclipse(date time.Time) SolarEclipseInfo {
	return module.LastSolarEclipseNASABulletinSplitK(date)
}

// LastSolarEclipseNASABulletinSplitK 上次日食（NASA bulletin Split-K） / previous solar eclipse with NASA bulletin Split-K.
func (module *Eclipse) LastSolarEclipseNASABulletinSplitK(date time.Time) SolarEclipseInfo {
	info, _ := searchSolarEclipse(date, -1, true, func(seedJDE float64) basic.SolarEclipseResult {
		return module.solarResult(seedJDE, basic.SolarEclipseModelNASABulletinSplitK)
	}, module)
	return info
}

// LastSolarEclipseIAUSingleK 上次日食（IAU Single-K） / previous solar eclipse with IAU Single-K.
func (module *Eclipse) LastSolarEclipseIAUSingleK(date time.Time) SolarEclipseInfo {
	info, _ := searchSolarEclipse(date, -1, true, func(seedJDE float64) basic.SolarEclipseResult {
		return module.solarResult(seedJDE, basic.SolarEclipseModelIAUSingleK)
	}, module)
	return info
}

// NextSolarEclipse 下次日食 / next solar eclipse.
func (module *Eclipse) NextSolarEclipse(date time.Time) SolarEclipseInfo {
	return module.NextSolarEclipseNASABulletinSplitK(date)
}

// NextSolarEclipseNASABulletinSplitK 下次日食（NASA bulletin Split-K） / next solar eclipse with NASA bulletin Split-K.
func (module *Eclipse) NextSolarEclipseNASABulletinSplitK(date time.Time) SolarEclipseInfo {
	info, _ := searchSolarEclipse(date, 1, false, func(seedJDE float64) basic.SolarEclipseResult {
		return module.solarResult(seedJDE, basic.SolarEclipseModelNASABulletinSplitK)
	}, module)
	return info
}

// NextSolarEclipseIAUSingleK 下次日食（IAU Single-K） / next solar eclipse with IAU Single-K.
func (module *Eclipse) NextSolarEclipseIAUSingleK(date time.Time) SolarEclipseInfo {
	info, _ := searchSolarEclipse(date, 1, false, func(seedJDE float64) basic.SolarEclipseResult {
		return module.solarResult(seedJDE, basic.SolarEclipseModelIAUSingleK)
	}, module)
	return info
}

// ClosestSolarEclipse 最近一次日食 / closest solar eclipse.
func (module *Eclipse) ClosestSolarEclipse(date time.Time) SolarEclipseInfo {
	return module.ClosestSolarEclipseNASABulletinSplitK(date)
}

// ClosestSolarEclipseNASABulletinSplitK 最近一次日食（NASA bulletin Split-K） / closest solar eclipse with NASA bulletin Split-K.
func (module *Eclipse) ClosestSolarEclipseNASABulletinSplitK(date time.Time) SolarEclipseInfo {
	last, hasLast := searchSolarEclipse(date, -1, true, func(seedJDE float64) basic.SolarEclipseResult {
		return module.solarResult(seedJDE, basic.SolarEclipseModelNASABulletinSplitK)
	}, module)
	next, hasNext := searchSolarEclipse(date, 1, false, func(seedJDE float64) basic.SolarEclipseResult {
		return module.solarResult(seedJDE, basic.SolarEclipseModelNASABulletinSplitK)
	}, module)
	return closestSolarEclipse(date, last, hasLast, next, hasNext)
}

// ClosestSolarEclipseIAUSingleK 最近一次日食（IAU Single-K） / closest solar eclipse with IAU Single-K.
func (module *Eclipse) ClosestSolarEclipseIAUSingleK(date time.Time) SolarEclipseInfo {
	last, hasLast := searchSolarEclipse(date, -1, true, func(seedJDE float64) basic.SolarEclipseResult {
		return module.solarResult(seedJDE, basic.SolarEclipseModelIAUSingleK)
	}, module)
	next, hasNext := searchSolarEclipse(date, 1, false, func(seedJDE float64) basic.SolarEclipseResult {
		return module.solarResult(seedJDE, basic.SolarEclipseModelIAUSingleK)
	}, module)
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
	calculator solarEclipseCalculator, modules ...*Eclipse) (SolarEclipseInfo, bool) {
	module := &defaultModule
	if len(modules) > 0 {
		module = modules[0]
	}

	targetTT := solarEclipseTimeToTTJDE(date)
	candidateTT := basic.CalcMoonSHByJDEWithInstance(targetTT, 0, module.inst)

	for i := 0; i < solarEclipseSearchLimit; i++ {
		if module.isPotentialSolarEclipse(candidateTT) {
			result := calculator(candidateTT)
			if result.Type != basic.SolarEclipseNone && solarEclipseMatchesDirection(result.GreatestEclipse, targetTT, direction, includeCurrent) {
				return solarEclipseInfoFromBasic(result, date.Location(), module), true
			}
		}
		candidateTT = nextEclipseSearchCandidateTT(candidateTT, 0, direction, solarEclipseSynodicMonthDays, module)
	}

	return SolarEclipseInfo{}, false
}

func isPotentialSolarEclipse(newMoonTT float64) bool {
	return defaultModule.isPotentialSolarEclipse(newMoonTT)
}

func (module *Eclipse) isPotentialSolarEclipse(newMoonTT float64) bool {
	return math.Abs(basic.HMoonTrueBoWithInstance(newMoonTT, module.inst)) <= solarEclipseLatitudeLimitDeg
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

func solarEclipseInfoFromBasic(result basic.SolarEclipseResult, location *time.Location, modules ...*Eclipse) SolarEclipseInfo {
	module := &defaultModule
	if len(modules) > 0 {
		module = modules[0]
	}

	saros, hasSaros := solarSarosInfo(result.GreatestEclipse, module)
	return SolarEclipseInfo{
		Model:               mapBasicSolarEclipseModel(result.Model),
		SunRadiusModel:      result.SunRadiusModel,
		Type:                mapBasicSolarEclipseType(result.Type),
		Centrality:          mapBasicSolarEclipseCentrality(result.Centrality),
		LunarLimbApplied:    result.LunarLimbApplied,
		PartialBeginContact: result.PartialBeginContact,
		PartialEndContact:   result.PartialEndContact,
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
