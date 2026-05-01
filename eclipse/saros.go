package eclipse

import (
	"time"

	"github.com/starainrt/astro/basic"
)

const (
	sarosCycleLunations = 223
	sarosCycleDays      = float64(sarosCycleLunations) * solarEclipseSynodicMonthDays
	sarosWalkLimit      = 100
)

// SarosInfo 沙罗序列信息, Saros series metadata.
type SarosInfo struct {
	// Series 是 NASA 沙罗系列编号；太阳食可能出现 0 号系列。
	// Series is the NASA Saros series number; solar eclipses may use series 0.
	Series int
	// Member 是本次食在该系列中的序号，从 1 开始计数。
	// Member is the 1-based index of this eclipse within the series.
	Member int
	// Count 是该系列的总成员数。
	// Count is the total number of eclipses in the series.
	Count int
}

type sarosAnchor struct {
	Series int16
	Count  uint8
	Year   int16
	Month  uint8
	Day    uint8
}

type sarosHeadOverride struct {
	Series       int16
	Count        uint8
	HeadYear     int16
	HeadMonth    uint8
	HeadDay      uint8
	MemberOffset int8
}

var solarSarosHeadOverrides = [...]sarosHeadOverride{
	{Series: 22, Count: 71, HeadYear: -2192, HeadMonth: 5, HeadDay: 17, MemberOffset: -1},
}

var lunarSarosHeadOverrides = [...]sarosHeadOverride{
	{Series: 4, Count: 78, HeadYear: -2483, HeadMonth: 1, HeadDay: 12, MemberOffset: 2},
	{Series: 8, Count: 86, HeadYear: -2494, HeadMonth: 8, HeadDay: 7, MemberOffset: 0},
	{Series: 61, Count: 78, HeadYear: -762, HeadMonth: 12, HeadDay: 24, MemberOffset: 1},
}

func solarSarosInfo(ttJDE float64) (SarosInfo, bool) {
	headTT, member, ok := solarSarosHead(ttJDE)
	if !ok {
		return SarosInfo{}, false
	}
	if info, ok := matchSarosHeadOverride(solarSarosHeadOverrides[:], headTT, member); ok {
		return info, true
	}
	anchor, ok := matchSarosAnchor(solarSarosAnchors[:], headTT)
	if !ok || member > int(anchor.Count) {
		return SarosInfo{}, false
	}
	return SarosInfo{
		Series: int(anchor.Series),
		Member: member,
		Count:  int(anchor.Count),
	}, true
}

func lunarSarosInfo(ttJDE float64) (SarosInfo, bool) {
	headTT, member, ok := lunarSarosHead(ttJDE)
	if !ok {
		return SarosInfo{}, false
	}
	if info, ok := matchSarosHeadOverride(lunarSarosHeadOverrides[:], headTT, member); ok {
		return info, true
	}
	anchor, ok := matchSarosAnchor(lunarSarosAnchors[:], headTT)
	if !ok || member > int(anchor.Count) {
		return SarosInfo{}, false
	}
	return SarosInfo{
		Series: int(anchor.Series),
		Member: member,
		Count:  int(anchor.Count),
	}, true
}

func solarSarosHead(ttJDE float64) (float64, int, bool) {
	currentTT := ttJDE
	member := 1
	for step := 0; step < sarosWalkLimit; step++ {
		previousSeed := basic.CalcMoonSHByJDE(currentTT-sarosCycleDays, 0)
		previous := basic.SolarEclipseNASABulletinSplitK(previousSeed)
		if previous.Type == basic.SolarEclipseNone {
			return currentTT, member, true
		}
		currentTT = previous.GreatestEclipse
		member++
	}
	return 0, 0, false
}

func lunarSarosHead(ttJDE float64) (float64, int, bool) {
	currentTT := ttJDE
	member := 1
	for step := 0; step < sarosWalkLimit; step++ {
		previousSeed := basic.CalcMoonSHByJDE(currentTT-sarosCycleDays, 1)
		previous := basic.LunarEclipseDanjon(previousSeed)
		if previous.Type == basic.LunarEclipseNone {
			return currentTT, member, true
		}
		currentTT = previous.Maximum
		member++
	}
	return 0, 0, false
}

func matchSarosAnchor(anchors []sarosAnchor, headTT float64) (sarosAnchor, bool) {
	headDate := basic.JDE2DateByZone(headTT, time.UTC, true)
	year, month, day := headDate.Date()
	monthNumber := int(month)
	for _, anchor := range anchors {
		if int(anchor.Year) == year && int(anchor.Month) == monthNumber && int(anchor.Day) == day {
			return anchor, true
		}
	}
	return sarosAnchor{}, false
}

func matchSarosHeadOverride(overrides []sarosHeadOverride, headTT float64, member int) (SarosInfo, bool) {
	headDate := basic.JDE2DateByZone(headTT, time.UTC, true)
	year, month, day := headDate.Date()
	monthNumber := int(month)
	for _, override := range overrides {
		if int(override.HeadYear) != year || int(override.HeadMonth) != monthNumber || int(override.HeadDay) != day {
			continue
		}
		adjustedMember := member + int(override.MemberOffset)
		if adjustedMember < 1 || adjustedMember > int(override.Count) {
			return SarosInfo{}, false
		}
		return SarosInfo{
			Series: int(override.Series),
			Member: adjustedMember,
			Count:  int(override.Count),
		}, true
	}
	return SarosInfo{}, false
}
