package eclipse

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// SolarEclipseCandidateOptions 批量枚举配置，零值为全部食型的 NASA bulletin Split-K / enumeration options.
type SolarEclipseCandidateOptions struct {
	// Model 月亮半径模型，零值为 NASA bulletin Split-K / lunar radius model.
	Model SolarEclipseRadiusModel
	// CentralOnly 只保留全食、环食与混合食，含中心线未落到地表的 T0/A0 事件。
	// CentralOnly keeps total, annular, and hybrid eclipses, including T0/A0 events
	// whose axis misses the Earth; the kind filter matches SearchLocalCentralSolarEclipse.
	CentralOnly bool
	// IncludeSaros 附带沙罗序列信息 / attach Saros metadata.
	IncludeSaros bool
}

// SolarEclipseCandidate 一场日食的时刻表条目，不含几何 / geometry-free timetable entry.
type SolarEclipseCandidate struct {
	// GreatestEclipse 食甚时刻 / greatest-eclipse instant.
	GreatestEclipse time.Time
	// Type 全局食型，Centrality 中心食类型 / global type and centrality.
	Type       SolarEclipseType
	Centrality SolarEclipseCentrality
	// Magnitude 与 Gamma 是全局食分与伽马 / global magnitude and gamma.
	Magnitude float64
	Gamma     float64
	// Saros 仅在 IncludeSaros 时有效 / valid only with IncludeSaros.
	Saros    SarosInfo
	HasSaros bool
}

// 食甚可能早于朔，枚举上界要多留这么多天。
const solarEclipseCandidateGreatestLeadDays = 0.5

// solarEclipseCandidateIsCentral 按全局食型判定，与站心搜索的 Kind 口径一致。
func solarEclipseCandidateIsCentral(value basic.SolarEclipseType) bool {
	switch value {
	case basic.SolarEclipseTotal, basic.SolarEclipseAnnular, basic.SolarEclipseHybrid:
		return true
	}
	return false
}

// SolarEclipseCandidates 枚举区间内的日食时刻表，按时间升序 / eclipses within a range, in time order.
func SolarEclipseCandidates(
	start, end time.Time, options SolarEclipseCandidateOptions,
) []SolarEclipseCandidate {
	if !end.After(start) {
		return nil
	}
	calculator := localSolarEclipseCalculatorForModel(options.Model)
	startTT := solarEclipseTimeToTTJDE(start)
	endTT := solarEclipseTimeToTTJDE(end)
	candidates := make([]SolarEclipseCandidate, 0, 8)
	candidateTT := basic.CalcMoonSHByJDE(startTT, 0)
	for iteration := 0; iteration < localSolarEclipseSearchIterationCap; iteration++ {
		// 食甚可能略早于朔，上界要多留一点余量，否则窗口末尾那场会被漏掉。
		if candidateTT > endTT+solarEclipseCandidateGreatestLeadDays {
			break
		}
		if isPotentialLocalSolarEclipse(candidateTT) {
			globalResult := calculator.global(candidateTT)
			if globalResult.Type != basic.SolarEclipseNone &&
				(!options.CentralOnly || solarEclipseCandidateIsCentral(globalResult.Type)) {
				greatest := solarEclipseTTJDEToTime(globalResult.GreatestEclipse, start.Location())
				// 候选朔月取的是“最近”的朔，可能落在 start 之前或 end 之后（食甚早于朔），
				// 按食甚时刻双侧过滤；注意不能 continue：候选推进在循环末尾，跳过它会原地打转。
				if !greatest.Before(start) && !greatest.After(end) {
					candidate := SolarEclipseCandidate{
						GreatestEclipse: greatest,
						Type:            mapBasicSolarEclipseType(globalResult.Type),
						Centrality:      mapBasicSolarEclipseCentrality(globalResult.Centrality),
						Magnitude:       globalResult.Magnitude,
						Gamma:           globalResult.Gamma,
					}
					if options.IncludeSaros {
						candidate.Saros, candidate.HasSaros = solarSarosInfo(globalResult.GreatestEclipse)
					}
					candidates = append(candidates, candidate)
				}
			}
		}
		candidateTT = nextEclipseSearchCandidateTT(candidateTT, 0, 1, localSolarEclipseSynodicMonthDays)
	}
	return candidates
}
