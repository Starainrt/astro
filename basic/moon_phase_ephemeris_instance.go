package basic

import (
	"math"

	. "github.com/starainrt/astro/tools"
)

func sunMoonSeekWithInstance(jde, degree float64, inst *Instance) float64 {
	if inst == nil {
		return SunMoonSeek(jde, degree)
	}
	p := HMoonApparentLoWithInstance(jde, inst) - HSunApparentLoWithInstance(jde, inst) - degree
	for p < -180 {
		p += 360
	}
	for p > 180 {
		p -= 360
	}
	return p
}

func refineSunMoonPhaseWithInstance(jde, degree float64, inst *Instance) float64 {
	estimateJD, ok := eventNewtonRefine(jde, 0.00001, func(prevJD float64) float64 {
		value := sunMoonSeekWithInstance(prevJD, degree, inst)
		slope := (sunMoonSeekWithInstance(prevJD+0.000005, degree, inst) - sunMoonSeekWithInstance(prevJD-0.000005, degree, inst)) / 0.00001
		return value / slope
	})
	if !ok {
		return math.NaN()
	}
	return estimateJD
}

// MoonPhaseWithInstance 月球照明比例 / illuminated lunar fraction with an explicit ephemeris instance.
func MoonPhaseWithInstance(jde float64, inst *Instance) float64 {
	if inst == nil {
		return MoonPhase(jde)
	}
	moonBo := HMoonTrueBoWithInstance(jde, inst)
	sunLo := HSunApparentLoWithInstance(jde, inst)
	moonLo := HMoonApparentLoWithInstance(jde, inst)
	tmp := Cos(moonBo) * Cos(sunLo-moonLo)
	earthSunDistance := EarthAwayNWithInstance(jde, -1, inst) * astronomicalUnitKM
	i := earthSunDistance * Sin(ArcCos(tmp)) / (HMoonAwayWithInstance(jde, inst) - earthSunDistance*tmp)
	i = ArcTan(i)
	if i < 0 {
		i += 180
	}
	if i > 180 {
		i -= 180
	}
	return (1 + Cos(i)) / 2
}

// SunMoonSeekWithInstance 日月黄经差相位残差 / Sun-Moon longitude phase residual with an explicit ephemeris instance.
func SunMoonSeekWithInstance(jde, degree float64, inst *Instance) float64 {
	return sunMoonSeekWithInstance(jde, degree, inst)
}

// CalcMoonSHByJDEWithInstance 按历元求朔望时刻 / calculate a syzygy from an epoch with an explicit ephemeris instance.
func CalcMoonSHByJDEWithInstance(jde float64, phaseType int, inst *Instance) float64 {
	if inst == nil {
		return CalcMoonSHByJDE(jde, phaseType)
	}
	return refineSunMoonPhaseWithInstance(jde, float64(phaseType*180), inst)
}

// CalcMoonSHWithInstance 按年份求朔望时刻 / calculate a syzygy from a year with an explicit ephemeris instance.
func CalcMoonSHWithInstance(year float64, phaseType int, inst *Instance) float64 {
	if inst == nil {
		return CalcMoonSH(year, phaseType)
	}
	return refineSunMoonPhaseWithInstance(CalcMoonSWithInstance(year, phaseType, inst), float64(phaseType*180), inst)
}

// CalcMoonSWithInstance 按年份估算朔望基准时刻 / estimate a syzygy base time from a year with an explicit ephemeris instance.
func CalcMoonSWithInstance(year float64, phaseType int, inst *Instance) float64 {
	if inst == nil {
		return CalcMoonS(year, phaseType)
	}
	return CalcMoonS(year, phaseType)
}

// CalcMoonXHByJDEWithInstance 按历元求弦月时刻 / calculate a quarter phase from an epoch with an explicit ephemeris instance.
func CalcMoonXHByJDEWithInstance(jde float64, quarterType int, inst *Instance) float64 {
	if inst == nil {
		return CalcMoonXHByJDE(jde, quarterType)
	}
	degree := -90.0
	if quarterType == 0 {
		degree = 90
	}
	return refineSunMoonPhaseWithInstance(jde, degree, inst)
}

// CalcMoonXHWithInstance 按年份求弦月时刻 / calculate a quarter phase from a year with an explicit ephemeris instance.
func CalcMoonXHWithInstance(year float64, quarterType int, inst *Instance) float64 {
	if inst == nil {
		return CalcMoonXH(year, quarterType)
	}
	degree := -90.0
	if quarterType == 0 {
		degree = 90
	}
	return refineSunMoonPhaseWithInstance(CalcMoonXWithInstance(year, quarterType, inst), degree, inst)
}

// CalcMoonXWithInstance 按年份估算弦月基准时刻 / estimate a quarter-phase base time from a year with an explicit ephemeris instance.
func CalcMoonXWithInstance(year float64, quarterType int, inst *Instance) float64 {
	if inst == nil {
		return CalcMoonX(year, quarterType)
	}
	return CalcMoonX(year, quarterType)
}
