package basic

import (
	"math"

	. "github.com/starainrt/astro/tools"
)

// Pos

const (
	URANUS_S_PERIOD            = 1 / ((1 / 365.256363004) - (1 / 30799.095))
	uranusEventSearchN         = 16
	uranusPhaseCoarseTolerance = 30.0 / 86400.0
)

func uranusSunLongitudeDelta(jde, degree float64, filter bool) float64 {
	sub := Limit360(Limit360(UranusApparentLo(jde)-HSunApparentLo(jde)) - degree)
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

func uranusSunLongitudeDeltaN(jde, degree float64, filter bool, n int) float64 {
	sub := Limit360(Limit360(UranusApparentLoN(jde, n)-HSunApparentLoN(jde, n)) - degree)
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

func uranusConjunctionFull(jde, degree float64, next uint8) float64 {
	//0=last 1=next
	daysPerDegree := URANUS_S_PERIOD / 360
	currentDelta := uranusSunLongitudeDelta(jde, degree, false)
	if next == 0 {
		jde -= (360 - currentDelta) * daysPerDegree
	} else {
		jde += daysPerDegree * currentDelta
	}
	estimateJD := jde
	for {
		prevJD := estimateJD
		longitudeDelta := uranusSunLongitudeDelta(prevJD, degree, true)
		longitudeSlope := (uranusSunLongitudeDelta(prevJD+0.000005, degree, true) - uranusSunLongitudeDelta(prevJD-0.000005, degree, true)) / 0.00001
		estimateJD = prevJD - longitudeDelta/longitudeSlope
		if math.Abs(estimateJD-prevJD) <= 0.00001 {
			break
		}
	}
	return TD2UT(estimateJD, false)
}

func uranusConjunction(jde, degree float64, next uint8) float64 {
	//0=last 1=next
	daysPerDegree := URANUS_S_PERIOD / 360
	currentDelta := uranusSunLongitudeDelta(jde, degree, false)
	if next == 0 {
		jde -= (360 - currentDelta) * daysPerDegree
	} else {
		jde += daysPerDegree * currentDelta
	}
	estimateJD := jde
	for {
		prevJD := estimateJD
		longitudeDelta := uranusSunLongitudeDeltaN(prevJD, degree, true, uranusEventSearchN)
		longitudeSlope := (uranusSunLongitudeDeltaN(prevJD+0.000005, degree, true, uranusEventSearchN) - uranusSunLongitudeDeltaN(prevJD-0.000005, degree, true, uranusEventSearchN)) / 0.00001
		estimateJD = prevJD - longitudeDelta/longitudeSlope
		if math.Abs(estimateJD-prevJD) <= uranusPhaseCoarseTolerance {
			break
		}
	}
	for {
		prevJD := estimateJD
		longitudeDelta := uranusSunLongitudeDelta(prevJD, degree, true)
		longitudeSlope := (uranusSunLongitudeDelta(prevJD+0.000005, degree, true) - uranusSunLongitudeDelta(prevJD-0.000005, degree, true)) / 0.00001
		estimateJD = prevJD - longitudeDelta/longitudeSlope
		if math.Abs(estimateJD-prevJD) <= 0.00001 {
			break
		}
	}
	return TD2UT(estimateJD, false)
}

func LastUranusConjunction(jde float64) float64 {
	return uranusConjunction(jde, 0, 0)
}

func NextUranusConjunction(jde float64) float64 {
	return uranusConjunction(jde, 0, 1)
}

func LastUranusOpposition(jde float64) float64 {
	return uranusConjunction(jde, 180, 0)
}

func NextUranusOpposition(jde float64) float64 {
	return uranusConjunction(jde, 180, 1)
}

func NextUranusEasternQuadrature(jde float64) float64 {
	return uranusConjunction(jde, 90, 1)
}

func LastUranusEasternQuadrature(jde float64) float64 {
	return uranusConjunction(jde, 90, 0)
}

func NextUranusWesternQuadrature(jde float64) float64 {
	return uranusConjunction(jde, 270, 1)
}

func LastUranusWesternQuadrature(jde float64) float64 {
	return uranusConjunction(jde, 270, 0)
}

func uranusRetrograde(jde float64, searchBeforeOpposition bool) float64 {
	//0=last 1=next
	raRate := func(jde float64, delta float64) float64 {
		sub := UranusApparentRa(jde+delta) - UranusApparentRa(jde-delta)
		if sub > 180 {
			sub -= 360
		}
		if sub < -180 {
			sub += 360
		}
		return sub / (2 * delta)
	}
	jde = uranusConjunctionFull(jde, 180, 1)
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

func NextUranusRetrogradeToPrograde(jde float64) float64 {
	date := uranusRetrograde(jde, false)
	if date < jde {
		oppositionJD := uranusConjunctionFull(jde, 180, 1)
		return uranusRetrograde(oppositionJD+10, false)
	}
	return date
}

func LastUranusRetrogradeToPrograde(jde float64) float64 {
	jde = uranusConjunctionFull(jde, 180, 0) - 10
	date := uranusRetrograde(jde, false)
	if date > jde {
		oppositionJD := uranusConjunctionFull(jde, 180, 0)
		return uranusRetrograde(oppositionJD-10, false)
	}
	return date
}

func NextUranusProgradeToRetrograde(jde float64) float64 {
	date := uranusRetrograde(jde, true)
	if date < jde {
		oppositionJD := uranusConjunctionFull(jde, 180, 1)
		return uranusRetrograde(oppositionJD+10, true)
	}
	return date
}

func LastUranusProgradeToRetrograde(jde float64) float64 {
	jde = uranusConjunctionFull(jde, 180, 0) - 10
	date := uranusRetrograde(jde, true)
	if date > jde {
		oppositionJD := uranusConjunctionFull(jde, 180, 0)
		return uranusRetrograde(oppositionJD-10, true)
	}
	return date
}
