package eclipse

import (
	"math"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/ephem"
)

// LunarEclipseOnDateWithInstance 指定星历实例的当地日月食查询，自定义引擎不推断沙罗 / local-date lunar eclipse with an explicit instance; custom engines omit Saros metadata.
func LunarEclipseOnDateWithInstance(date time.Time, inst *ephem.Instance) (LunarEclipseInfo, bool) {
	return lunarEclipseOnDateWithInstance(date, lunarCalculatorWithInstance(false, inst), true, inst)
}

// LunarEclipseOnDateDanjonWithInstance 指定星历实例的 Danjon 当地日月食查询 / local-date Danjon lunar eclipse with an explicit instance.
func LunarEclipseOnDateDanjonWithInstance(date time.Time, inst *ephem.Instance) (LunarEclipseInfo, bool) {
	return lunarEclipseOnDateWithInstance(date, lunarCalculatorWithInstance(false, inst), false, inst)
}

// LunarEclipseOnDateChauvenetWithInstance 指定星历实例的 Chauvenet 当地日月食查询 / local-date Chauvenet lunar eclipse with an explicit instance.
func LunarEclipseOnDateChauvenetWithInstance(date time.Time, inst *ephem.Instance) (LunarEclipseInfo, bool) {
	return lunarEclipseOnDateWithInstance(date, lunarCalculatorWithInstance(true, inst), false, inst)
}

// LastLunarEclipseWithInstance 指定星历实例的上次月食 / previous lunar eclipse with an explicit instance.
func LastLunarEclipseWithInstance(date time.Time, inst *ephem.Instance) LunarEclipseInfo {
	info, _ := searchLunarEclipseWithInstance(date, -1, true, lunarCalculatorWithInstance(false, inst), true, inst)
	return info
}

// LastLunarEclipseDanjonWithInstance 指定星历实例的 Danjon 上次月食 / previous Danjon lunar eclipse with an explicit instance.
func LastLunarEclipseDanjonWithInstance(date time.Time, inst *ephem.Instance) LunarEclipseInfo {
	info, _ := searchLunarEclipseWithInstance(date, -1, true, lunarCalculatorWithInstance(false, inst), false, inst)
	return info
}

// LastLunarEclipseChauvenetWithInstance 指定星历实例的 Chauvenet 上次月食 / previous Chauvenet lunar eclipse with an explicit instance.
func LastLunarEclipseChauvenetWithInstance(date time.Time, inst *ephem.Instance) LunarEclipseInfo {
	info, _ := searchLunarEclipseWithInstance(date, -1, true, lunarCalculatorWithInstance(true, inst), false, inst)
	return info
}

// NextLunarEclipseWithInstance 指定星历实例的下次月食 / next lunar eclipse with an explicit instance.
func NextLunarEclipseWithInstance(date time.Time, inst *ephem.Instance) LunarEclipseInfo {
	info, _ := searchLunarEclipseWithInstance(date, 1, false, lunarCalculatorWithInstance(false, inst), true, inst)
	return info
}

// NextLunarEclipseDanjonWithInstance 指定星历实例的 Danjon 下次月食 / next Danjon lunar eclipse with an explicit instance.
func NextLunarEclipseDanjonWithInstance(date time.Time, inst *ephem.Instance) LunarEclipseInfo {
	info, _ := searchLunarEclipseWithInstance(date, 1, false, lunarCalculatorWithInstance(false, inst), false, inst)
	return info
}

// NextLunarEclipseChauvenetWithInstance 指定星历实例的 Chauvenet 下次月食 / next Chauvenet lunar eclipse with an explicit instance.
func NextLunarEclipseChauvenetWithInstance(date time.Time, inst *ephem.Instance) LunarEclipseInfo {
	info, _ := searchLunarEclipseWithInstance(date, 1, false, lunarCalculatorWithInstance(true, inst), false, inst)
	return info
}

// ClosestLunarEclipseWithInstance 指定星历实例的最近月食 / closest lunar eclipse with an explicit instance.
func ClosestLunarEclipseWithInstance(date time.Time, inst *ephem.Instance) LunarEclipseInfo {
	return closestLunarEclipseWithInstance(date, false, true, inst)
}

// ClosestLunarEclipseDanjonWithInstance 指定星历实例的 Danjon 最近月食 / closest Danjon lunar eclipse with an explicit instance.
func ClosestLunarEclipseDanjonWithInstance(date time.Time, inst *ephem.Instance) LunarEclipseInfo {
	return closestLunarEclipseWithInstance(date, false, false, inst)
}

// ClosestLunarEclipseChauvenetWithInstance 指定星历实例的 Chauvenet 最近月食 / closest Chauvenet lunar eclipse with an explicit instance.
func ClosestLunarEclipseChauvenetWithInstance(date time.Time, inst *ephem.Instance) LunarEclipseInfo {
	return closestLunarEclipseWithInstance(date, true, false, inst)
}

func lunarCalculatorWithInstance(chauvenet bool, inst *ephem.Instance) lunarEclipseCalculator {
	if chauvenet {
		return func(jde float64) basic.LunarEclipseResult { return basic.LunarEclipseChauvenetWithInstance(jde, inst) }
	}
	return func(jde float64) basic.LunarEclipseResult { return basic.LunarEclipseDanjonWithInstance(jde, inst) }
}

func closestLunarEclipseWithInstance(date time.Time, chauvenet, allowDefaultFallback bool, inst *ephem.Instance) LunarEclipseInfo {
	calculator := lunarCalculatorWithInstance(chauvenet, inst)
	last, hasLast := searchLunarEclipseWithInstance(date, -1, true, calculator, allowDefaultFallback, inst)
	next, hasNext := searchLunarEclipseWithInstance(date, 1, false, calculator, allowDefaultFallback, inst)
	return closestLunarEclipse(date, last, hasLast, next, hasNext)
}

func moonSunLoDiffWithInstance(date time.Time, inst *ephem.Instance) float64 {
	if inst == nil {
		return moonSunLoDiff(date)
	}
	jde := timeToTTJDE(date)
	return normalizeDegree360(basic.HMoonApparentLoWithInstance(jde, inst) - basic.HSunApparentLoWithInstance(jde, inst))
}

func nextLunarEclipseSearchCandidateTTWithInstance(candidateTT float64, direction int, inst *ephem.Instance) float64 {
	if inst == nil {
		return nextEclipseSearchCandidateTT(candidateTT, 1, direction, lunarEclipseSynodicMonthDays)
	}
	step := 1
	if inst.Model() != ephem.Custom {
		step = eclipseSearchStep(candidateTT, direction, lunarEclipseSynodicMonthDays)
	}
	return basic.CalcMoonSHByJDEWithInstance(candidateTT+float64(direction*step)*lunarEclipseSynodicMonthDays, 1, inst)
}

func lunarEclipseMoonCenterPositionAngleWithInstance(ttJDE float64, inst *ephem.Instance) float64 {
	if inst == nil {
		return lunarEclipseMoonCenterPositionAngle(ttJDE)
	}
	sunRA, sunDec := basic.HSunApparentRaDecWithInstance(ttJDE, inst)
	moonRA, moonDec := basic.HMoonTrueRaDecWithInstance(ttJDE, inst)
	return positionAngle(normalizeDegree360(sunRA+180), -sunDec, moonRA, moonDec)
}

func lunarSarosInfoWithInstance(ttJDE float64, inst *ephem.Instance) (SarosInfo, bool) {
	if inst == nil {
		return lunarSarosInfo(ttJDE)
	}
	if inst.Model() == ephem.Custom {
		return SarosInfo{}, false
	}
	if info, ok := matchSarosMagic(lunarSarosAnchors[:], 1, lunarSarosHeadOverrides[:], ttJDE); ok {
		return info, true
	}
	info, ok := extendedSarosInfoForModel(ttJDE, 1, inst.Model(), func(lunation, phase int) (float64, bool, bool) {
		seed := sarosLunationEpoch + (float64(lunation)+float64(phase)/2)*solarEclipseSynodicMonthDays
		result := basic.LunarEclipseDanjonWithInstance(seed, inst)
		if result.Type == basic.LunarEclipseNone {
			result = basic.LunarEclipseChauvenetWithInstance(seed, inst)
		}
		valid := !math.IsNaN(result.Maximum) && !math.IsInf(result.Maximum, 0) && math.Abs(result.Maximum-seed) < sarosEclipseSeedToleranceDays
		return result.Maximum, result.Type != basic.LunarEclipseNone, valid
	})
	if !ok {
		return SarosInfo{}, false
	}
	return reconcileSarosSpan(info, ttJDE, 1)
}
