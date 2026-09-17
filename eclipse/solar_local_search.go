package eclipse

import (
	"math"
	"time"

	"github.com/starainrt/astro/basic"
)

// SolarEclipseLocalSearchKind 站心搜索的目标食型 / which station eclipse to look for.
type SolarEclipseLocalSearchKind int

const (
	// SolarEclipseLocalSearchAny 任意站心日食 / any local eclipse.
	SolarEclipseLocalSearchAny SolarEclipseLocalSearchKind = iota
	// SolarEclipseLocalSearchTotal 站心全食 / local total eclipse.
	SolarEclipseLocalSearchTotal
	// SolarEclipseLocalSearchAnnular 站心环食 / local annular eclipse.
	SolarEclipseLocalSearchAnnular
)

// SolarEclipseLocalSearchOptions 站心搜索配置，零值为向后找任意可见日食 / station search options.
type SolarEclipseLocalSearchOptions struct {
	// Kind 目标食型，零值为 Any / target kind.
	Kind SolarEclipseLocalSearchKind
	// Backward 为 true 时向过去搜索 / search towards the past.
	Backward bool
	// IncludeCurrent 起始时刻那一场也算命中 / accept an eclipse at the start instant.
	IncludeCurrent bool
	// Geometric 为 true 时跳过可见性门槛 / skip the visibility gate.
	Geometric bool
	// Model 月亮半径模型，零值为 NASA bulletin Split-K / lunar radius model.
	Model SolarEclipseRadiusModel
	// MaxYears 搜索跨度（年），<=0 用与旧迭代上限等价的默认跨度 / search horizon in years.
	MaxYears float64
}

// SolarEclipseLocalSearchStatus 搜索结局与实际覆盖范围 / search outcome and horizon covered.
type SolarEclipseLocalSearchStatus struct {
	// Found 是否找到目标事件 / whether the target was found.
	Found bool
	// Exhausted 已走到跨度上界仍未找到 / horizon consumed without a hit.
	Exhausted bool
	// YearsScanned 实际覆盖跨度（年），MonthsScanned 候选朔望月数 / covered horizon and candidates.
	YearsScanned  float64
	MonthsScanned int
	// Capped 被迭代硬上限截停，此时 Exhausted 不再代表跨度内没有 / iteration cap stopped the search.
	Capped bool
	// GlobalChecks 与 LocalChecks 是全局与站心解算次数 / solution counters.
	GlobalChecks int
	LocalChecks  int
}

const (
	// 旧口径为 6000 次候选步进，实测覆盖 992.33~992.47 年；取 992.5 年完整覆盖旧跨度。
	localSolarEclipseDefaultSearchYears = 992.5
	// 时间跨度搜索的迭代硬上限，防止病态输入死循环。
	localSolarEclipseSearchIterationCap = 200000
)

// SearchLocalCentralSolarEclipse 在给定跨度内搜索站心食并回报是否用尽跨度 / station search within a horizon.
func SearchLocalCentralSolarEclipse(
	date time.Time, lon, lat, height float64,
	options SolarEclipseLocalSearchOptions,
) (LocalSolarEclipseInfo, SolarEclipseLocalSearchStatus) {
	calculator := localSolarEclipseCalculatorForModel(options.Model)
	mode := localSolarEclipseQueryVisible
	if options.Geometric {
		mode = localSolarEclipseQueryGeometric
	}
	maxYears := options.MaxYears
	if maxYears <= 0 || math.IsNaN(maxYears) || math.IsInf(maxYears, 0) {
		maxYears = localSolarEclipseDefaultSearchYears
	}
	direction := 1
	if options.Backward {
		direction = -1
	}
	targetTT := solarEclipseTimeToTTJDE(date)
	info, status, ok := searchLocalCentralSolarEclipse(
		targetTT, targetTT+float64(direction)*maxYears*365.25, direction,
		options.IncludeCurrent, options.Kind, lon, lat, height, date.Location(),
		calculator, mode,
	)
	status.Found = ok
	status.Exhausted = !ok && !status.Capped
	return info, status
}

func localSolarEclipseCalculatorForModel(model SolarEclipseRadiusModel) localSolarEclipseCalculator {
	if model == SolarEclipseModelIAUSingleK {
		return localSolarEclipseIAUSingleK
	}
	return localSolarEclipseNASABulletinSplitK
}

func searchLocalCentralSolarEclipse(
	targetTT, horizonTT float64,
	direction int,
	includeCurrent bool,
	kind SolarEclipseLocalSearchKind,
	lon, lat, height float64,
	location *time.Location,
	calculator localSolarEclipseCalculator,
	mode localSolarEclipseQueryMode,
) (LocalSolarEclipseInfo, SolarEclipseLocalSearchStatus, bool) {
	status := SolarEclipseLocalSearchStatus{}
	candidateTT := basic.CalcMoonSHByJDE(targetTT, 0)
	capped := true
	for iteration := 0; iteration < localSolarEclipseSearchIterationCap; iteration++ {
		if direction > 0 && candidateTT > horizonTT {
			capped = false
			break
		}
		if direction < 0 && candidateTT < horizonTT {
			capped = false
			break
		}
		status.MonthsScanned++
		status.YearsScanned = math.Abs(candidateTT-targetTT) / 365.25
		if isPotentialLocalSolarEclipse(candidateTT) {
			globalResult := calculator.global(candidateTT)
			status.GlobalChecks++
			if globalResult.Type != basic.SolarEclipseNone {
				if localSolarEclipseKindMatchesGlobal(kind, globalResult) {
					result := calculator.local(globalResult.GreatestEclipse, lon, lat, height)
					status.LocalChecks++
					if localSolarEclipseKindMatchesLocal(kind, result) {
						info := localSolarEclipseInfoFromBasic(result, lon, lat, height, location)
						if (mode != localSolarEclipseQueryVisible || localSolarEclipseKindVisible(kind, info)) &&
							localSolarEclipseMatchesDirection(result.GreatestEclipse, targetTT, direction, includeCurrent) {
							return info, status, true
						}
					}
				}
			}
		}
		candidateTT = nextEclipseSearchCandidateTT(candidateTT, 0, direction, localSolarEclipseSynodicMonthDays)
	}
	// 循环自然结束说明撞上了迭代硬上限。
	status.Capped = capped
	return LocalSolarEclipseInfo{}, status, false
}

func localSolarEclipseKindMatchesGlobal(kind SolarEclipseLocalSearchKind, result basic.SolarEclipseResult) bool {
	switch kind {
	case SolarEclipseLocalSearchTotal:
		return result.HasTotal || result.HasHybrid
	case SolarEclipseLocalSearchAnnular:
		return result.HasAnnular || result.HasHybrid
	}
	return true
}

func localSolarEclipseKindMatchesLocal(kind SolarEclipseLocalSearchKind, result basic.LocalSolarEclipseResult) bool {
	switch kind {
	case SolarEclipseLocalSearchTotal:
		return result.HasTotal
	case SolarEclipseLocalSearchAnnular:
		return result.HasAnnular && !result.HasTotal
	}
	return result.Type != basic.SolarEclipseNone
}

func localSolarEclipseKindVisible(kind SolarEclipseLocalSearchKind, info LocalSolarEclipseInfo) bool {
	if kind == SolarEclipseLocalSearchAny {
		return localSolarEclipseVisible(info)
	}
	return localCentralSolarEclipseVisible(info)
}
