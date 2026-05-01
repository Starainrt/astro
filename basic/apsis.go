package basic

import (
	"math"
	"sort"
	"time"
)

const (
	earthApsisBaseTTJDE        = 2451547.507
	earthApsisMeanYearDays     = 365.2596358
	earthApsisQuadraticTerm    = 0.0000000156
	earthApsisBaseYear         = 2000.01
	earthApsisSeedScale        = 0.99997
	earthApsisBracketHalfWidth = 5.0
	earthApsisSampleStep       = 0.25
	earthApsisDerivativeStep   = 1e-3
	earthApsisToleranceDays    = 1e-8
	earthApsisMaxIterations    = 24
	moonApsisBaseTTJDE         = 2451534.6698
	moonApsisMeanMonthDays     = 27.55454989
	moonApsisBaseCycle         = 1325.55
	moonApsisQuadraticTerm     = -0.0006691
	moonApsisCubicTerm         = -0.000001098
	moonApsisQuarticTerm       = 0.0000000052
	moonApsisBracketHalfWidth  = 2.0
	moonApsisSampleStep        = 0.125
	moonApsisDerivativeStep    = 1e-4
	moonApsisToleranceDays     = 1e-8
	moonApsisMaxIterations     = 24
)

// ApsisEvent 轨道极值事件 / orbital distance extremum event.
type ApsisEvent struct {
	// JDE 是事件发生时刻对应的世界时儒略日 / event time as UTC-based Julian day.
	JDE float64
	// Distance 是极值距离；地球相关事件单位 AU，月球相关事件单位 km / extremum distance.
	Distance float64
}

type apsisSearchConfig struct {
	bracketHalfWidth float64
	sampleStep       float64
	derivativeStep   float64
	toleranceDays    float64
	maxIterations    int
	maximize         bool
}

// EarthPerihelion 地球指定年份的近日点 / Earth perihelion in the given year.
func EarthPerihelion(year int) ApsisEvent {
	return earthApsis(year, false)
}

// EarthAphelion 地球指定年份的远日点 / Earth aphelion in the given year.
func EarthAphelion(year int) ApsisEvent {
	return earthApsis(year, true)
}

// MoonPerigees 指定年月内的所有月球近地点 / all lunar perigees in the given Gregorian month.
func MoonPerigees(year int, month time.Month) []ApsisEvent {
	return moonApsisInMonth(year, month, false)
}

// MoonApogees 指定年月内的所有月球远地点 / all lunar apogees in the given Gregorian month.
func MoonApogees(year int, month time.Month) []ApsisEvent {
	return moonApsisInMonth(year, month, true)
}

func earthApsis(year int, aphelion bool) ApsisEvent {
	seedTT := earthApsisSeedTT(year, aphelion)
	cfg := apsisSearchConfig{
		bracketHalfWidth: earthApsisBracketHalfWidth,
		sampleStep:       earthApsisSampleStep,
		derivativeStep:   earthApsisDerivativeStep,
		toleranceDays:    earthApsisToleranceDays,
		maxIterations:    earthApsisMaxIterations,
		maximize:         aphelion,
	}
	eventTT, distanceAU := refineDistanceExtremum(seedTT, cfg, EarthAway)
	return ApsisEvent{
		JDE:      TD2UT(eventTT, false),
		Distance: distanceAU,
	}
}

func moonApsisInMonth(year int, month time.Month, apogee bool) []ApsisEvent {
	startUTC := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	endUTC := startUTC.AddDate(0, 1, 0)
	startTT := TD2UT(Date2JDE(startUTC), true)
	endTT := TD2UT(Date2JDE(endUTC), true)

	kStart := int(math.Floor((startTT-moonApsisBaseTTJDE)/moonApsisMeanMonthDays)) - 1
	kEnd := int(math.Ceil((endTT-moonApsisBaseTTJDE)/moonApsisMeanMonthDays)) + 1
	phase := 0.0
	if apogee {
		phase = 0.5
	}

	cfg := apsisSearchConfig{
		bracketHalfWidth: moonApsisBracketHalfWidth,
		sampleStep:       moonApsisSampleStep,
		derivativeStep:   moonApsisDerivativeStep,
		toleranceDays:    moonApsisToleranceDays,
		maxIterations:    moonApsisMaxIterations,
		maximize:         apogee,
	}

	events := make([]ApsisEvent, 0, 2)
	for k := kStart; k <= kEnd; k++ {
		seedTT := moonApsisSeedTT(float64(k) + phase)
		eventTT, distanceKM := refineDistanceExtremum(seedTT, cfg, HMoonAway)
		eventUT := TD2UT(eventTT, false)
		eventTimeUTC := JDE2DateByZone(eventUT, time.UTC, false)
		if eventTimeUTC.Before(startUTC) || !eventTimeUTC.Before(endUTC) {
			continue
		}
		events = append(events, ApsisEvent{
			JDE:      eventUT,
			Distance: distanceKM,
		})
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].JDE < events[j].JDE
	})
	return events
}

func earthApsisSeedTT(year int, aphelion bool) float64 {
	k := math.Round(earthApsisSeedScale * (float64(year) - earthApsisBaseYear))
	if aphelion {
		k += 0.5
	}
	return earthApsisBaseTTJDE + earthApsisMeanYearDays*k + earthApsisQuadraticTerm*k*k
}

func moonApsisSeedTT(k float64) float64 {
	t := k / moonApsisBaseCycle
	return moonApsisBaseTTJDE +
		moonApsisMeanMonthDays*k +
		moonApsisQuadraticTerm*t*t +
		moonApsisCubicTerm*t*t*t +
		moonApsisQuarticTerm*t*t*t*t
}

func refineDistanceExtremum(seed float64, cfg apsisSearchConfig, distanceFn func(float64) float64) (float64, float64) {
	best := seed
	bestDistance := distanceFn(seed)
	for sample := seed - cfg.bracketHalfWidth; sample <= seed+cfg.bracketHalfWidth+1e-12; sample += cfg.sampleStep {
		dist := distanceFn(sample)
		if distanceBetter(dist, bestDistance, cfg.maximize) {
			best = sample
			bestDistance = dist
		}
	}

	left, right, ok := apsisDerivativeBracket(best, seed, cfg, distanceFn)
	if !ok {
		return best, bestDistance
	}

	leftDeriv := apsisDistanceDerivative(distanceFn, left, cfg.derivativeStep)
	rightDeriv := apsisDistanceDerivative(distanceFn, right, cfg.derivativeStep)
	current := best
	for i := 0; i < cfg.maxIterations; i++ {
		first, second := apsisDistanceDerivatives(distanceFn, current, cfg.derivativeStep)
		next := current
		if math.Abs(second) > 0 {
			next = current - first/second
		}
		if !(next > left && next < right) || math.IsNaN(next) || math.IsInf(next, 0) {
			next = (left + right) / 2
		}

		nextDeriv := apsisDistanceDerivative(distanceFn, next, cfg.derivativeStep)
		if leftDeriv == 0 {
			right = next
			rightDeriv = nextDeriv
		} else if leftDeriv*nextDeriv <= 0 {
			right = next
			rightDeriv = nextDeriv
		} else {
			left = next
			leftDeriv = nextDeriv
		}

		if math.Abs(next-current) <= cfg.toleranceDays || math.Abs(right-left) <= cfg.toleranceDays {
			current = next
			break
		}
		current = next
		_ = rightDeriv
	}

	return current, distanceFn(current)
}

func apsisDerivativeBracket(best, seed float64, cfg apsisSearchConfig, distanceFn func(float64) float64) (float64, float64, bool) {
	leftBound := seed - cfg.bracketHalfWidth
	rightBound := seed + cfg.bracketHalfWidth
	left := best - cfg.sampleStep
	right := best + cfg.sampleStep
	if left < leftBound {
		left = leftBound
	}
	if right > rightBound {
		right = rightBound
	}

	leftDeriv := apsisDistanceDerivative(distanceFn, left, cfg.derivativeStep)
	rightDeriv := apsisDistanceDerivative(distanceFn, right, cfg.derivativeStep)
	for i := 0; i < cfg.maxIterations; i++ {
		if leftDeriv == 0 || rightDeriv == 0 || leftDeriv*rightDeriv < 0 {
			return left, right, true
		}
		if left > leftBound {
			left -= cfg.sampleStep
			if left < leftBound {
				left = leftBound
			}
			leftDeriv = apsisDistanceDerivative(distanceFn, left, cfg.derivativeStep)
		}
		if right < rightBound {
			right += cfg.sampleStep
			if right > rightBound {
				right = rightBound
			}
			rightDeriv = apsisDistanceDerivative(distanceFn, right, cfg.derivativeStep)
		}
	}
	return 0, 0, false
}

func apsisDistanceDerivative(distanceFn func(float64) float64, jd, h float64) float64 {
	return (distanceFn(jd+h) - distanceFn(jd-h)) / (2 * h)
}

func apsisDistanceDerivatives(distanceFn func(float64) float64, jd, h float64) (float64, float64) {
	prev := distanceFn(jd - h)
	curr := distanceFn(jd)
	next := distanceFn(jd + h)
	first := (next - prev) / (2 * h)
	second := (next - 2*curr + prev) / (h * h)
	return first, second
}

func distanceBetter(candidate, current float64, maximize bool) bool {
	if maximize {
		return candidate > current
	}
	return candidate < current
}
