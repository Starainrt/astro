package basic

import (
	"math"

	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

const (
	VENUS_S_PERIOD    = 1 / ((1 / 224.701) - (1 / 365.256363004))
	venusEventSearchN = 16
)

func venusSunLongitudeDelta(jde float64) float64 {
	sub := Limit360(VenusApparentLo(jde) - HSunApparentLo(jde))
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub
}

func venusSunLongitudeDeltaN(jde float64, n int) float64 {
	sub := Limit360(VenusApparentLoN(jde, n) - HSunApparentLoN(jde, n))
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub
}

func venusConjunctionAngleDelta(diff float64) float64 {
	diff = Limit360(diff)
	if diff > 180 {
		diff -= 360
	}
	if diff < -180 {
		diff += 360
	}
	return diff
}

func venusConjunctionHeliocentricDelta(jd, targetDeg float64, n int) float64 {
	planetLo := planet.WherePlanetN(2, 0, jd, n)
	earthLo := planet.WherePlanetN(-1, 0, jd, n)
	return venusConjunctionAngleDelta(planetLo - earthLo - targetDeg)
}

func venusSunRADelta(jde float64) float64 {
	sub := Limit360(VenusApparentRa(jde) - SunApparentRa(jde))
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub
}

func venusRADerivative(jde, val float64) float64 {
	sub := VenusApparentRa(jde+val) - VenusApparentRa(jde-val)
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub / (2 * val)
}

func venusRADerivativeN(jde, val float64, n int) float64 {
	sub := VenusApparentRaN(jde+val, n) - VenusApparentRaN(jde-val, n)
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub / (2 * val)
}

func venusRAContinuousForMax(jde float64) float64 {
	ra := VenusApparentRa(jde)
	if ra < 180 {
		return ra + 360
	}
	return ra
}

func venusRAContinuousForMaxN(jde float64, n int) float64 {
	ra := VenusApparentRaN(jde, n)
	if ra < 180 {
		return ra + 360
	}
	return ra
}

func venusRAContinuousForMin(jde float64) float64 {
	ra := VenusApparentRa(jde)
	if ra > 180 {
		return ra - 360
	}
	return ra
}

func venusRAContinuousForMinN(jde float64, n int) float64 {
	ra := VenusApparentRaN(jde, n)
	if ra > 180 {
		return ra - 360
	}
	return ra
}

func venusRAExtremumRefine(seed, start, end, step float64, fn func(float64) float64) float64 {
	centerJD := clampFloat64(seed, start, end)
	halfStep := step
	bestJD := centerJD
	bestVal := fn(centerJD)
	for i := 0; i < 8; i++ {
		leftJD := clampFloat64(centerJD-halfStep, start, end)
		rightJD := clampFloat64(centerJD+halfStep, start, end)
		leftVal := fn(leftJD)
		centerVal := fn(centerJD)
		rightVal := fn(rightJD)
		if leftVal > bestVal {
			bestVal = leftVal
			bestJD = leftJD
		}
		if centerVal > bestVal {
			bestVal = centerVal
			bestJD = centerJD
		}
		if rightVal > bestVal {
			bestVal = rightVal
			bestJD = rightJD
		}
		denominator := leftVal - 2*centerVal + rightVal
		if denominator == 0 {
			centerJD = bestJD
			halfStep /= 2
			continue
		}
		vertexJD := centerJD + 0.5*halfStep*(leftVal-rightVal)/denominator
		vertexJD = clampFloat64(vertexJD, leftJD, rightJD)
		vertexVal := fn(vertexJD)
		if vertexVal > bestVal {
			bestVal = vertexVal
			bestJD = vertexJD
		}
		centerJD = bestJD
		halfStep /= 2
	}
	return bestJD
}

func venusSunElongationN(jde float64, n int) float64 {
	lo1, bo1 := VenusApparentLoBoN(jde, n)
	lo2 := HSunApparentLoN(jde, n)
	bo2 := HSunTrueBoN(jde, n)
	return StarAngularSeparation(lo1, bo1, lo2, bo2)
}

func venusTrueElongationN(jde float64, n int) float64 {
	earth := mercuryHelioN(-1, jde, n)
	planetPos := mercuryHelioN(2, jde, n)
	geo := mercuryGeocentric(planetPos, earth)
	return StarAngularSeparation(geo.lo, geo.bo, HSunTrueLoN(jde, n), HSunTrueBoN(jde, n))
}

func venusElongationDerivative(jde, val float64) float64 {
	sub := VenusSunElongation(jde+val) - VenusSunElongation(jde-val)
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub / (2 * val)
}

func venusElongationDerivativeN(jde, val float64, n int) float64 {
	sub := venusSunElongationN(jde+val, n) - venusSunElongationN(jde-val, n)
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub / (2 * val)
}

func venusConjunction(jde float64, next uint8) float64 {
	queryTT := jde
	direction := -1.0
	if next == 1 {
		direction = 1
	}
	left := queryTT
	leftVal := venusSunLongitudeDeltaN(left, venusEventSearchN)
	if math.Abs(leftVal) <= 30.0/86400.0 {
		exact := eventZeroRefine(left, 1.0, 0.000005, venusSunLongitudeDelta)
		if math.Abs(exact-queryTT) <= 1.0 {
			return TD2UT(exact, false)
		}
	}
	const step = 8.0
	for i := 0; i < 80; i++ {
		right := queryTT + direction*step*float64(i+1)
		rightVal := venusSunLongitudeDeltaN(right, venusEventSearchN)
		if leftVal == 0 || rightVal == 0 || leftVal*rightVal <= 0 {
			center := (left + right) / 2.0
			halfWindow := math.Abs(right-left) / 2.0
			return TD2UT(eventZeroRefine(center, halfWindow, 0.000005, venusSunLongitudeDelta), false)
		}
		left = right
		leftVal = rightVal
	}
	return TD2UT(eventZeroRefine(queryTT, VENUS_S_PERIOD, 0.000005, venusSunLongitudeDelta), false)
}

func venusConjunctionTypeAt(eventUT float64) bool {
	return EarthVenusAway(eventUT) <= EarthAway(eventUT)
}

func nextVenusTypedConjunctionFromEvent(jde float64, inferior bool) float64 {
	date := NextVenusConjunctionStrict(jde)
	if venusConjunctionTypeAt(date) == inferior {
		return date
	}
	return NextVenusConjunctionStrict(eventUTNextQueryTT(date))
}

func lastVenusTypedConjunctionFromEvent(jde float64, inferior bool) float64 {
	date := LastVenusConjunctionStrict(jde)
	if venusConjunctionTypeAt(date) == inferior {
		return date
	}
	return LastVenusConjunctionStrict(eventUTLastQueryTT(date))
}

func LastVenusConjunction(jde float64) float64 {
	return inclusiveLastSimpleEvent(jde, LastVenusConjunctionStrict, NextVenusConjunctionStrict)
}

func NextVenusConjunction(jde float64) float64 {
	return inclusiveNextSimpleEvent(jde, LastVenusConjunctionStrict, NextVenusConjunctionStrict)
}

func LastVenusConjunctionStrict(jde float64) float64 {
	return venusConjunction(jde, 0)
}

func NextVenusConjunctionStrict(jde float64) float64 {
	return venusConjunction(jde, 1)
}

func nextVenusTypedConjunction(jde float64, inferior bool) float64 {
	return nextVenusTypedConjunctionFromEvent(jde, inferior)
}

func lastVenusTypedConjunction(jde float64, inferior bool) float64 {
	return lastVenusTypedConjunctionFromEvent(jde, inferior)
}

func NextVenusInferiorConjunction(jde float64) float64 {
	return nextVenusTypedConjunction(jde, true)
}

func NextVenusSuperiorConjunction(jde float64) float64 {
	return nextVenusTypedConjunction(jde, false)
}

func LastVenusInferiorConjunction(jde float64) float64 {
	return lastVenusTypedConjunction(jde, true)
}

func LastVenusSuperiorConjunction(jde float64) float64 {
	return lastVenusTypedConjunction(jde, false)
}

func venusRetrograde(jde float64) float64 {
	//0=last 1=next
	lastHe := LastVenusConjunctionStrict(jde)
	nextHe := NextVenusConjunctionStrict(jde)
	nowSub := venusSunRADelta(jde)
	if nowSub > 0 {
		jde = lastHe + ((nextHe - lastHe) / 5.0 * 3.5)
	} else {
		jde = lastHe + 10
	}
	for {
		nowSub := venusRADerivativeN(jde, 1.0/86400.0, venusEventSearchN)
		if math.Abs(nowSub) > 0.5 {
			jde += 5
			continue
		}
		break
	}
	JD1 := jde
	for {
		JD0 := JD1
		stDegree := venusRADerivative(JD0, 0.5/86400.0)
		stDegreep := (venusRADerivative(JD0+10.0/86400.0, 0.5/86400.0) - venusRADerivative(JD0-10.0/86400.0, 0.5/86400.0)) / (20.0 / 86400.0)
		JD1 = JD0 - stDegree/stDegreep
		if math.Abs(JD1-JD0) <= 20.0/86400.0 {
			break
		}
	}
	min := eventZeroRefine(JD1, 10.0/86400.0, 0.5/86400.0, func(jd float64) float64 {
		return venusRADerivative(jd, 0.5/86400.0)
	})
	//fmt.Println((min - lastHe) / (nextHe - lastHe))
	return TD2UT(min, false)
}

func NextVenusRetrograde(jde float64) float64 {
	p2r := NextVenusProgradeToRetrograde(jde)
	r2p := NextVenusRetrogradeToPrograde(jde)
	if sameEventJD(p2r, r2p) {
		return p2r
	}
	if p2r < r2p {
		return p2r
	}
	return r2p
}

func LastVenusRetrograde(jde float64) float64 {
	p2r := LastVenusProgradeToRetrograde(jde)
	r2p := LastVenusRetrogradeToPrograde(jde)
	if sameEventJD(p2r, r2p) {
		return p2r
	}
	if p2r > r2p {
		return p2r
	}
	return r2p
}

func venusStationInWindow(start, end float64, progradeToRetrograde bool) float64 {
	var best float64
	if progradeToRetrograde {
		guess := scanWindowForMax(start, end, 2.0, func(jd float64) float64 {
			return venusRAContinuousForMaxN(jd, venusEventSearchN)
		})
		best = venusRAExtremumRefine(guess, start, end, 1.0, func(jd float64) float64 {
			return venusRAContinuousForMax(jd)
		})
	} else {
		guess := scanWindowForMax(start, end, 2.0, func(jd float64) float64 {
			return -venusRAContinuousForMinN(jd, venusEventSearchN)
		})
		best = venusRAExtremumRefine(guess, start, end, 1.0, func(jd float64) float64 {
			return -venusRAContinuousForMin(jd)
		})
	}
	return TD2UT(best, false)
}

func venusProgradeToRetrogradeAroundInferior(inferior float64) float64 {
	return venusStationInWindow(inferior-30.0, inferior-14.0, true)
}

func venusRetrogradeToProgradeAroundInferior(inferior float64) float64 {
	return venusStationInWindow(inferior+14.0, inferior+24.0, false)
}

func NextVenusProgradeToRetrograde(jde float64) float64 {
	inferior := NextVenusInferiorConjunction(jde)
	for {
		date := venusProgradeToRetrogradeAroundInferior(inferior)
		if eventUTQueryAfterOrEqual(date, jde) {
			return date
		}
		inferior = NextVenusInferiorConjunction(eventUTNextQueryTT(inferior))
	}
}

func NextVenusRetrogradeToPrograde(jde float64) float64 {
	inferior := LastVenusInferiorConjunction(jde)
	for {
		date := venusRetrogradeToProgradeAroundInferior(inferior)
		if eventUTQueryAfterOrEqual(date, jde) {
			return date
		}
		inferior = NextVenusInferiorConjunction(eventUTNextQueryTT(inferior))
	}
}

func LastVenusProgradeToRetrograde(jde float64) float64 {
	inferior := NextVenusInferiorConjunction(jde)
	for {
		date := venusProgradeToRetrogradeAroundInferior(inferior)
		if eventUTQueryBeforeOrEqual(date, jde) {
			return date
		}
		inferior = LastVenusInferiorConjunction(eventUTLastQueryTT(inferior))
	}
}

func LastVenusRetrogradeToPrograde(jde float64) float64 {
	inferior := LastVenusInferiorConjunction(jde)
	for {
		date := venusRetrogradeToProgradeAroundInferior(inferior)
		if eventUTQueryBeforeOrEqual(date, jde) {
			return date
		}
		inferior = LastVenusInferiorConjunction(eventUTLastQueryTT(inferior))
	}
}

func VenusSunElongation(jde float64) float64 {
	lo1, bo1 := VenusApparentLoBo(jde)
	lo2 := HSunApparentLo(jde)
	bo2 := HSunTrueBo(jde)
	return StarAngularSeparation(lo1, bo1, lo2, bo2)
}

func venusGreatestElongationInWindow(start, end float64) float64 {
	best := maximizeInWindow(start, end, 5.0, func(jd float64) float64 {
		return venusTrueElongationN(jd, venusEventSearchN)
	}, func(jd float64) float64 {
		return venusTrueElongationN(jd, -1)
	})
	return TD2UT(best, false)
}

func venusEastElongationWindowEndingAt(inferior float64) (float64, float64) {
	lastSuperior := LastVenusSuperiorConjunction(eventUTLastQueryTT(inferior))
	return lastSuperior + innerEventEpsilon, inferior - innerEventEpsilon
}

func venusWestElongationWindowEndingAt(superior float64) (float64, float64) {
	lastInferior := LastVenusInferiorConjunction(eventUTLastQueryTT(superior))
	return lastInferior + innerEventEpsilon, superior - innerEventEpsilon
}

func venusEastElongationWindowContaining(jde float64) (float64, float64) {
	nextInferior := NextVenusInferiorConjunction(jde)
	start, end := venusEastElongationWindowEndingAt(nextInferior)
	if eventUTQueryBeforeOrEqual(start, jde) && eventUTQueryAfterOrEqual(end, jde) {
		return start, end
	}
	currentInferior := LastVenusInferiorConjunction(jde)
	return venusEastElongationWindowEndingAt(currentInferior)
}

func venusWestElongationWindowContaining(jde float64) (float64, float64) {
	nextSuperior := NextVenusSuperiorConjunction(jde)
	start, end := venusWestElongationWindowEndingAt(nextSuperior)
	if eventUTQueryBeforeOrEqual(start, jde) && eventUTQueryAfterOrEqual(end, jde) {
		return start, end
	}
	currentSuperior := LastVenusSuperiorConjunction(jde)
	return venusWestElongationWindowEndingAt(currentSuperior)
}

func nextVenusGreatestElongationTyped(jde float64, east bool) float64 {
	if east {
		start, windowEnd := venusEastElongationWindowContaining(jde)
		for {
			date := venusGreatestElongationInWindow(start, windowEnd)
			if eventUTQueryAfterOrEqual(date, jde) {
				return date
			}
			nextInferior := NextVenusInferiorConjunction(eventUTNextQueryTT(windowEnd))
			start, windowEnd = venusEastElongationWindowEndingAt(nextInferior)
		}
	}
	start, windowEnd := venusWestElongationWindowContaining(jde)
	for {
		date := venusGreatestElongationInWindow(start, windowEnd)
		if eventUTQueryAfterOrEqual(date, jde) {
			return date
		}
		nextSuperior := NextVenusSuperiorConjunction(eventUTNextQueryTT(windowEnd))
		start, windowEnd = venusWestElongationWindowEndingAt(nextSuperior)
	}
}

func lastVenusGreatestElongationTyped(jde float64, east bool) float64 {
	if east {
		start, windowEnd := venusEastElongationWindowContaining(jde)
		for {
			date := venusGreatestElongationInWindow(start, windowEnd)
			if eventUTQueryBeforeOrEqual(date, jde) {
				return date
			}
			prevInferior := LastVenusInferiorConjunction(eventUTLastQueryTT(start))
			start, windowEnd = venusEastElongationWindowEndingAt(prevInferior)
		}
	}
	start, windowEnd := venusWestElongationWindowContaining(jde)
	for {
		date := venusGreatestElongationInWindow(start, windowEnd)
		if eventUTQueryBeforeOrEqual(date, jde) {
			return date
		}
		prevSuperior := LastVenusSuperiorConjunction(eventUTLastQueryTT(start))
		start, windowEnd = venusWestElongationWindowEndingAt(prevSuperior)
	}
}

func venusGreatestElongation(jde float64) float64 {
	east := venusSunRADelta(jde) > 0
	if east {
		return nextVenusGreatestElongationTyped(jde, true)
	}
	return nextVenusGreatestElongationTyped(jde, false)
}

func NextVenusGreatestElongation(jde float64) float64 {
	east := NextVenusGreatestElongationEast(jde)
	west := NextVenusGreatestElongationWest(jde)
	if sameEventJD(east, west) {
		return east
	}
	if east < west {
		return east
	}
	return west
}

func LastVenusGreatestElongation(jde float64) float64 {
	east := LastVenusGreatestElongationEast(jde)
	west := LastVenusGreatestElongationWest(jde)
	if sameEventJD(east, west) {
		return east
	}
	if east > west {
		return east
	}
	return west
}

func LastVenusInferiorConjunctionInclusive(jde float64) float64 {
	date := LastVenusConjunction(jde)
	if venusConjunctionTypeAt(date) {
		return date
	}
	return LastVenusConjunction(eventUTLastQueryTT(date))
}

func NextVenusInferiorConjunctionInclusive(jde float64) float64 {
	date := NextVenusConjunction(jde)
	if venusConjunctionTypeAt(date) {
		return date
	}
	return NextVenusConjunction(eventUTNextQueryTT(date))
}

func LastVenusSuperiorConjunctionInclusive(jde float64) float64 {
	date := LastVenusConjunction(jde)
	if !venusConjunctionTypeAt(date) {
		return date
	}
	return LastVenusConjunction(eventUTLastQueryTT(date))
}

func NextVenusSuperiorConjunctionInclusive(jde float64) float64 {
	date := NextVenusConjunction(jde)
	if !venusConjunctionTypeAt(date) {
		return date
	}
	return NextVenusConjunction(eventUTNextQueryTT(date))
}

func LastVenusRetrogradeInclusive(jde float64) float64 {
	return inclusiveLastSimpleEvent(jde, LastVenusRetrograde, NextVenusRetrograde)
}

func NextVenusRetrogradeInclusive(jde float64) float64 {
	return inclusiveNextSimpleEvent(jde, LastVenusRetrograde, NextVenusRetrograde)
}

func LastVenusProgradeToRetrogradeInclusive(jde float64) float64 {
	return inclusiveLastSimpleEvent(jde, LastVenusProgradeToRetrograde, NextVenusProgradeToRetrograde)
}

func NextVenusProgradeToRetrogradeInclusive(jde float64) float64 {
	return inclusiveNextSimpleEvent(jde, LastVenusProgradeToRetrograde, NextVenusProgradeToRetrograde)
}

func LastVenusRetrogradeToProgradeInclusive(jde float64) float64 {
	return inclusiveLastSimpleEvent(jde, LastVenusRetrogradeToPrograde, NextVenusRetrogradeToPrograde)
}

func NextVenusRetrogradeToProgradeInclusive(jde float64) float64 {
	return inclusiveNextSimpleEvent(jde, LastVenusRetrogradeToPrograde, NextVenusRetrogradeToPrograde)
}

func LastVenusGreatestElongationInclusive(jde float64) float64 {
	return inclusiveLastSimpleEvent(jde, LastVenusGreatestElongation, NextVenusGreatestElongation)
}

func NextVenusGreatestElongationInclusive(jde float64) float64 {
	return inclusiveNextSimpleEvent(jde, LastVenusGreatestElongation, NextVenusGreatestElongation)
}

func LastVenusGreatestElongationEastInclusive(jde float64) float64 {
	return inclusiveLastSimpleEvent(jde, LastVenusGreatestElongationEast, NextVenusGreatestElongationEast)
}

func NextVenusGreatestElongationEastInclusive(jde float64) float64 {
	return inclusiveNextSimpleEvent(jde, LastVenusGreatestElongationEast, NextVenusGreatestElongationEast)
}

func LastVenusGreatestElongationWestInclusive(jde float64) float64 {
	return inclusiveLastSimpleEvent(jde, LastVenusGreatestElongationWest, NextVenusGreatestElongationWest)
}

func NextVenusGreatestElongationWestInclusive(jde float64) float64 {
	return inclusiveNextSimpleEvent(jde, LastVenusGreatestElongationWest, NextVenusGreatestElongationWest)
}

func NextVenusGreatestElongationEast(jde float64) float64 {
	return nextVenusGreatestElongationTyped(jde, true)
}

func NextVenusGreatestElongationWest(jde float64) float64 {
	return nextVenusGreatestElongationTyped(jde, false)
}

func LastVenusGreatestElongationEast(jde float64) float64 {
	return lastVenusGreatestElongationTyped(jde, true)
}

func LastVenusGreatestElongationWest(jde float64) float64 {
	return lastVenusGreatestElongationTyped(jde, false)
}
