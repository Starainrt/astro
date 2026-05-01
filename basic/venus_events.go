package basic

import (
	"math"

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

func venusSunElongationN(jde float64, n int) float64 {
	lo1, bo1 := VenusApparentLoBoN(jde, n)
	lo2 := SunApparentLo(jde)
	bo2 := HSunTrueBoN(jde, n)
	return StarAngularSeparation(lo1, bo1, lo2, bo2)
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
	//0=last 1=next
	nowSub := venusSunLongitudeDeltaN(jde, venusEventSearchN)
	pos := math.Abs(venusSunLongitudeDeltaN(jde+1/86400.0, venusEventSearchN)) - math.Abs(nowSub)
	if pos >= 0 && next == 1 && nowSub > 0 {
		jde += VENUS_S_PERIOD/8.0 + 2
	}
	if pos >= 0 && next == 1 && nowSub < 0 {
		jde += VENUS_S_PERIOD/6.0 + 2
	}
	if pos <= 0 && next == 0 && nowSub < 0 {
		jde -= VENUS_S_PERIOD/8.0 + 2
	}
	if pos <= 0 && next == 0 && nowSub > 0 {
		jde -= VENUS_S_PERIOD/6.0 + 2
	}
	for {
		nowSub := venusSunLongitudeDeltaN(jde, venusEventSearchN)
		pos := math.Abs(venusSunLongitudeDeltaN(jde+1/86400.0, venusEventSearchN)) - math.Abs(nowSub)
		if math.Abs(nowSub) > 24 || (pos > 0 && next == 1) || (pos < 0 && next == 0) {
			if next == 1 {
				jde += 8
			} else {
				jde -= 8
			}
			continue
		}
		break
	}
	JD1 := jde
	for {
		JD0 := JD1
		stDegree := venusSunLongitudeDelta(JD0)
		stDegreep := (venusSunLongitudeDelta(JD0+0.000005) - venusSunLongitudeDelta(JD0-0.000005)) / 0.00001
		JD1 = JD0 - stDegree/stDegreep
		if math.Abs(JD1-JD0) <= 0.00001 {
			break
		}
	}
	return TD2UT(JD1, false)
}

func LastVenusConjunction(jde float64) float64 {
	return venusConjunction(jde, 0)
}

func NextVenusConjunction(jde float64) float64 {
	return venusConjunction(jde, 1)
}

func NextVenusInferiorConjunction(jde float64) float64 {
	date := NextVenusConjunction(jde)
	if EarthVenusAway(date) > EarthAway(date) {
		return NextVenusConjunction(date + 2)
	}
	return date
}

func NextVenusSuperiorConjunction(jde float64) float64 {
	date := NextVenusConjunction(jde)
	if EarthVenusAway(date) < EarthAway(date) {
		return NextVenusConjunction(date + 2)
	}
	return date
}

func LastVenusInferiorConjunction(jde float64) float64 {
	date := LastVenusConjunction(jde)
	if EarthVenusAway(date) > EarthAway(date) {
		return LastVenusConjunction(date - 2)
	}
	return date
}

func LastVenusSuperiorConjunction(jde float64) float64 {
	date := LastVenusConjunction(jde)
	if EarthVenusAway(date) < EarthAway(date) {
		return LastVenusConjunction(date - 2)
	}
	return date
}

func venusRetrograde(jde float64) float64 {
	//0=last 1=next
	lastHe := LastVenusConjunction(jde)
	nextHe := NextVenusConjunction(jde)
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
	date := venusRetrograde(jde)
	if date < jde {
		nextHe := NextVenusConjunction(jde)
		return venusRetrograde(nextHe + 2)
	}
	return date
}

func LastVenusRetrograde(jde float64) float64 {
	lastHe := LastVenusConjunction(jde)
	date := venusRetrograde(lastHe + 2)
	if date > jde {
		lastLastHe := LastVenusConjunction(lastHe - 2)
		return venusRetrograde(lastLastHe + 2)
	}
	return date
}

func NextVenusProgradeToRetrograde(jde float64) float64 {
	date := NextVenusRetrograde(jde)
	sub := Limit360(VenusApparentRa(date) - SunApparentRa(date))
	if sub > 180 {
		return NextVenusRetrograde(date + VENUS_S_PERIOD/2)
	}
	return date
}

func NextVenusRetrogradeToPrograde(jde float64) float64 {
	date := NextVenusRetrograde(jde)
	sub := Limit360(VenusApparentRa(date) - SunApparentRa(date))
	if sub < 180 {
		return NextVenusRetrograde(date + 12)
	}
	return date
}

func LastVenusProgradeToRetrograde(jde float64) float64 {
	date := LastVenusRetrograde(jde)
	sub := Limit360(VenusApparentRa(date) - SunApparentRa(date))
	if sub > 180 {
		return LastVenusRetrograde(date - 12)
	}
	return date
}

func LastVenusRetrogradeToPrograde(jde float64) float64 {
	date := LastVenusRetrograde(jde)
	sub := Limit360(VenusApparentRa(date) - SunApparentRa(date))
	if sub < 180 {
		return LastVenusRetrograde(date - VENUS_S_PERIOD/2)
	}
	return date
}

func VenusSunElongation(jde float64) float64 {
	lo1, bo1 := VenusApparentLoBo(jde)
	lo2 := SunApparentLo(jde)
	bo2 := HSunTrueBo(jde)
	return StarAngularSeparation(lo1, bo1, lo2, bo2)
}

func venusGreatestElongation(jde float64) float64 {
	lastHe := LastVenusConjunction(jde)
	nextHe := NextVenusConjunction(jde)
	nowSub := venusSunRADelta(jde)
	if nowSub > 0 {
		jde = lastHe + ((nextHe - lastHe) / 5.0 * 2.5)
	} else {
		jde = lastHe + ((nextHe - lastHe) / 5.0)
	}
	for {
		nowSub := venusElongationDerivativeN(jde, 1.0/86400.0, venusEventSearchN)
		if math.Abs(nowSub) > 0.15 {
			jde += 5
			continue
		}
		break
	}
	JD1 := jde
	for {
		JD0 := JD1
		stDegree := venusElongationDerivative(JD0, 2.0/86400.0)
		stDegreep := (venusElongationDerivative(JD0+15.0/86400.0, 2.0/86400.0) - venusElongationDerivative(JD0-15.0/86400.0, 2.0/86400.0)) / (30.0 / 86400.0)
		JD1 = JD0 - stDegree/stDegreep
		if math.Abs(JD1-JD0) <= 30.0/86400.0 {
			break
		}
	}
	min := eventZeroRefine(JD1, 15.0/86400.0, 0.5/86400.0, func(jd float64) float64 {
		return venusElongationDerivative(jd, 0.5/86400.0)
	})
	//fmt.Println((min - lastHe) / (nextHe - lastHe))
	return TD2UT(min, false)
}

func NextVenusGreatestElongation(jde float64) float64 {
	date := venusGreatestElongation(jde)
	if date < jde {
		nextHe := NextVenusConjunction(jde)
		return venusGreatestElongation(nextHe + 2)
	}
	return date
}

func LastVenusGreatestElongation(jde float64) float64 {
	lastHe := LastVenusConjunction(jde)
	date := venusGreatestElongation(lastHe + 2)
	if date > jde {
		lastLastHe := LastVenusConjunction(lastHe - 2)
		return venusGreatestElongation(lastLastHe + 2)
	}
	return date
}

func NextVenusGreatestElongationEast(jde float64) float64 {
	date := NextVenusGreatestElongation(jde)
	sub := Limit360(VenusApparentRa(date) - SunApparentRa(date))
	if sub > 180 {
		return NextVenusGreatestElongation(date + 1)
	}
	return date
}

func NextVenusGreatestElongationWest(jde float64) float64 {
	date := NextVenusGreatestElongation(jde)
	sub := Limit360(VenusApparentRa(date) - SunApparentRa(date))
	if sub < 180 {
		return NextVenusGreatestElongation(date + 1)
	}
	return date
}

func LastVenusGreatestElongationEast(jde float64) float64 {
	date := LastVenusGreatestElongation(jde)
	sub := Limit360(VenusApparentRa(date) - SunApparentRa(date))
	if sub > 180 {
		return LastVenusGreatestElongation(date - 1)
	}
	return date
}

func LastVenusGreatestElongationWest(jde float64) float64 {
	date := LastVenusGreatestElongation(jde)
	sub := Limit360(VenusApparentRa(date) - SunApparentRa(date))
	if sub < 180 {
		return LastVenusGreatestElongation(date - 1)
	}
	return date
}
