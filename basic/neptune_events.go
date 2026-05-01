package basic

import (
	"math"

	. "github.com/starainrt/astro/tools"
)

// Pos

const (
	NEPTUNE_S_PERIOD            = 1 / ((1 / 365.256363004) - (1 / 4332.59))
	neptuneEventSearchN         = 16
	neptunePhaseCoarseTolerance = 30.0 / 86400.0
)

func neptuneSunLongitudeDelta(jde, degree float64, filter bool) float64 {
	sub := Limit360(Limit360(NeptuneApparentLo(jde)-HSunApparentLo(jde)) - degree)
	if filter {
		if sub > 180 {
			sub -= 360
		}
		if sub < -180 {
			sub += 360
		}
	}
	return sub
}

func neptuneSunLongitudeDeltaN(jde, degree float64, filter bool, n int) float64 {
	sub := Limit360(Limit360(NeptuneApparentLoN(jde, n)-HSunApparentLoN(jde, n)) - degree)
	if filter {
		if sub > 180 {
			sub -= 360
		}
		if sub < -180 {
			sub += 360
		}
	}
	return sub
}

func neptuneConjunctionFull(jde, degree float64, next uint8) float64 {
	//0=last 1=next
	daysPerDegree := NEPTUNE_S_PERIOD / 360
	currentDelta := neptuneSunLongitudeDelta(jde, degree, false)
	if next == 0 {
		jde -= (360 - currentDelta) * daysPerDegree
	} else {
		jde += daysPerDegree * currentDelta
	}
	estimateJD := jde
	for {
		prevJD := estimateJD
		longitudeDelta := neptuneSunLongitudeDelta(prevJD, degree, true)
		longitudeSlope := (neptuneSunLongitudeDelta(prevJD+0.000005, degree, true) - neptuneSunLongitudeDelta(prevJD-0.000005, degree, true)) / 0.00001
		estimateJD = prevJD - longitudeDelta/longitudeSlope
		if math.Abs(estimateJD-prevJD) <= 0.00001 {
			break
		}
	}
	return TD2UT(estimateJD, false)
}

func neptuneConjunction(jde, degree float64, next uint8) float64 {
	//0=last 1=next
	daysPerDegree := NEPTUNE_S_PERIOD / 360
	currentDelta := neptuneSunLongitudeDelta(jde, degree, false)
	if next == 0 {
		jde -= (360 - currentDelta) * daysPerDegree
	} else {
		jde += daysPerDegree * currentDelta
	}
	estimateJD := jde
	for {
		prevJD := estimateJD
		longitudeDelta := neptuneSunLongitudeDeltaN(prevJD, degree, true, neptuneEventSearchN)
		longitudeSlope := (neptuneSunLongitudeDeltaN(prevJD+0.000005, degree, true, neptuneEventSearchN) - neptuneSunLongitudeDeltaN(prevJD-0.000005, degree, true, neptuneEventSearchN)) / 0.00001
		estimateJD = prevJD - longitudeDelta/longitudeSlope
		if math.Abs(estimateJD-prevJD) <= neptunePhaseCoarseTolerance {
			break
		}
	}
	for {
		prevJD := estimateJD
		longitudeDelta := neptuneSunLongitudeDelta(prevJD, degree, true)
		longitudeSlope := (neptuneSunLongitudeDelta(prevJD+0.000005, degree, true) - neptuneSunLongitudeDelta(prevJD-0.000005, degree, true)) / 0.00001
		estimateJD = prevJD - longitudeDelta/longitudeSlope
		if math.Abs(estimateJD-prevJD) <= 0.00001 {
			break
		}
	}
	return TD2UT(estimateJD, false)
}

func LastNeptuneConjunction(jde float64) float64 {
	return neptuneConjunction(jde, 0, 0)
}

func NextNeptuneConjunction(jde float64) float64 {
	return neptuneConjunction(jde, 0, 1)
}

func LastNeptuneOpposition(jde float64) float64 {
	return neptuneConjunction(jde, 180, 0)
}

func NextNeptuneOpposition(jde float64) float64 {
	return neptuneConjunction(jde, 180, 1)
}

func NextNeptuneEasternQuadrature(jde float64) float64 {
	return neptuneConjunction(jde, 90, 1)
}

func LastNeptuneEasternQuadrature(jde float64) float64 {
	return neptuneConjunction(jde, 90, 0)
}

func NextNeptuneWesternQuadrature(jde float64) float64 {
	return neptuneConjunction(jde, 270, 1)
}

func LastNeptuneWesternQuadrature(jde float64) float64 {
	return neptuneConjunction(jde, 270, 0)
}

func neptuneRetrograde(jde float64, searchBeforeOpposition bool) float64 {
	//0=last 1=next
	raRate := func(jde float64, delta float64) float64 {
		sub := NeptuneApparentRa(jde+delta) - NeptuneApparentRa(jde-delta)
		if sub > 180 {
			sub -= 360
		}
		if sub < -180 {
			sub += 360
		}
		return sub / (2 * delta)
	}
	jde = neptuneConjunctionFull(jde, 180, 1)
	if searchBeforeOpposition {
		jde -= 60
	} else {
		jde += 60
	}
	for {
		currentRate := raRate(jde, 1.0/86400.0)
		if math.Abs(currentRate) > 0.55 {
			jde += 2
			continue
		}
		break
	}
	estimateJD := jde
	for {
		prevJD := estimateJD
		rateValue := raRate(prevJD, 2.0/86400.0)
		rateSlope := (raRate(prevJD+15.0/86400.0, 2.0/86400.0) - raRate(prevJD-15.0/86400.0, 2.0/86400.0)) / (30.0 / 86400.0)
		estimateJD = prevJD - rateValue/rateSlope
		if math.Abs(estimateJD-prevJD) <= 30.0/86400.0 {
			break
		}
	}
	bestJD := eventZeroRefine(estimateJD, 15.0/86400.0, 0.5/86400.0, func(jd float64) float64 {
		return raRate(jd, 0.5/86400.0)
	})
	return TD2UT(bestJD, false)
}

func NextNeptuneRetrogradeToPrograde(jde float64) float64 {
	date := neptuneRetrograde(jde, false)
	if date < jde {
		oppositionJD := neptuneConjunctionFull(jde, 180, 1)
		return neptuneRetrograde(oppositionJD+10, false)
	}
	return date
}

func LastNeptuneRetrogradeToPrograde(jde float64) float64 {
	jde = neptuneConjunctionFull(jde, 180, 0) - 10
	date := neptuneRetrograde(jde, false)
	if date > jde {
		oppositionJD := neptuneConjunctionFull(jde, 180, 0)
		return neptuneRetrograde(oppositionJD-10, false)
	}
	return date
}

func NextNeptuneProgradeToRetrograde(jde float64) float64 {
	date := neptuneRetrograde(jde, true)
	if date < jde {
		oppositionJD := neptuneConjunctionFull(jde, 180, 1)
		return neptuneRetrograde(oppositionJD+10, true)
	}
	return date
}

func LastNeptuneProgradeToRetrograde(jde float64) float64 {
	jde = neptuneConjunctionFull(jde, 180, 0) - 10
	date := neptuneRetrograde(jde, true)
	if date > jde {
		oppositionJD := neptuneConjunctionFull(jde, 180, 0)
		return neptuneRetrograde(oppositionJD-10, true)
	}
	return date
}
