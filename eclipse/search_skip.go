package eclipse

import (
	"math"

	"github.com/starainrt/astro/basic"
)

const (
	eclipseSeasonNodeDistanceLimitDeg = 35.0
	eclipseSeasonMaxSearchStep        = 4
)

func nextEclipseSearchCandidateTT(candidateTT float64, phaseType, direction int, synodicMonthDays float64) float64 {
	step := eclipseSearchStep(candidateTT, direction, synodicMonthDays)
	return basic.CalcMoonSHByJDE(candidateTT+float64(direction*step)*synodicMonthDays, phaseType)
}

func eclipseSearchStep(candidateTT float64, direction int, synodicMonthDays float64) int {
	step := 1
	for nextStep := 2; nextStep <= eclipseSeasonMaxSearchStep; nextStep++ {
		skippedTT := candidateTT + float64(direction*(nextStep-1))*synodicMonthDays
		if eclipseNodeDistance(skippedTT) < eclipseSeasonNodeDistanceLimitDeg {
			break
		}
		step = nextStep
	}
	return step
}

func eclipseNodeDistance(ttJDE float64) float64 {
	argument := normalizeDegree360(basic.MoonLonX(ttJDE))
	toAscending := math.Min(argument, 360-argument)
	toDescending := math.Abs(argument - 180)
	return math.Min(toAscending, toDescending)
}
