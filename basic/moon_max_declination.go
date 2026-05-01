package basic

import (
	"math"
	"sort"
	"time"

	. "github.com/starainrt/astro/tools"
)

const (
	moonMaxDeclinationMeanMonthDays = 27.321582247
	moonMaxDeclinationBaseCycle     = 1336.86
	moonMaxDeclinationSearchSpan    = 3
)

// DeclinationEvent 赤纬极值事件 / declination extremum event.
type DeclinationEvent struct {
	// JDE 是事件发生时刻对应的世界时儒略日 / event time as UTC-based Julian day.
	JDE float64
	// Declination 是该时刻月心地心赤纬，单位度 / geocentric lunar declination at the event, in degrees.
	Declination float64
}

type moonMaxDeclinationCoefficients struct {
	D0   float64
	M0   float64
	MP0  float64
	F0   float64
	JDE0 float64
	sign float64
	tc   [44]float64
	dc   [37]float64
}

var moonMaxDeclinationNorthCoefficients = moonMaxDeclinationCoefficients{
	D0:   152.2029,
	M0:   14.8591,
	MP0:  4.6881,
	F0:   325.8867,
	JDE0: 2451562.5897,
	sign: 1,
	tc: [44]float64{
		0.8975, -0.4726, -0.1030, -0.0976, -0.0462, -0.0461, -0.0438, 0.0162,
		-0.0157, 0.0145, 0.0136, -0.0095, -0.0091, -0.0089, 0.0075, -0.0068,
		0.0061, -0.0047, -0.0043, -0.0040, -0.0037, 0.0031, 0.0030, -0.0029,
		-0.0029, -0.0027, 0.0024, -0.0021, 0.0019, 0.0018, 0.0018, 0.0017,
		0.0017, -0.0014, 0.0013, 0.0013, 0.0012, 0.0011, -0.0011, 0.0010,
		0.0010, -0.0009, 0.0007, -0.0007,
	},
	dc: [37]float64{
		5.1093, 0.2658, 0.1448, -0.0322, 0.0133, 0.0125, -0.0124, -0.0101,
		0.0097, -0.0087, 0.0074, 0.0067, 0.0063, 0.0060, -0.0057, -0.0056,
		0.0052, 0.0041, -0.0040, 0.0038, -0.0034, -0.0029, 0.0029, -0.0028,
		-0.0028, -0.0023, -0.0021, 0.0019, 0.0018, 0.0017, 0.0015, 0.0014,
		-0.0012, -0.0012, -0.0010, -0.0010, 0.0006,
	},
}

var moonMaxDeclinationSouthCoefficients = moonMaxDeclinationCoefficients{
	D0:   345.6676,
	M0:   1.3951,
	MP0:  186.21,
	F0:   145.1633,
	JDE0: 2451548.9289,
	sign: -1,
	tc: [44]float64{
		-0.8975, -0.4726, -0.1030, -0.0976, 0.0541, 0.0516, -0.0438, 0.0112,
		0.0157, 0.0023, -0.0136, 0.0110, 0.0091, 0.0089, 0.0075, -0.0030,
		-0.0061, -0.0047, -0.0043, 0.0040, -0.0037, -0.0031, 0.0030, 0.0029,
		-0.0029, -0.0027, 0.0024, -0.0021, -0.0019, -0.0006, -0.0018, -0.0017,
		0.0017, 0.0014, -0.0013, -0.0013, 0.0012, 0.0011, 0.0011, 0.0010,
		0.0010, -0.0009, -0.0007, -0.0007,
	},
	dc: [37]float64{
		-5.1093, 0.2658, -0.1448, 0.0322, 0.0133, 0.0125, -0.0015, 0.0101,
		-0.0097, 0.0087, 0.0074, 0.0067, -0.0063, -0.0060, 0.0057, -0.0056,
		-0.0052, -0.0041, -0.0040, -0.0038, 0.0034, -0.0029, 0.0029, 0.0028,
		-0.0028, 0.0023, 0.0021, 0.0019, 0.0018, -0.0017, 0.0015, 0.0014,
		0.0012, -0.0012, 0.0010, -0.0010, 0.0037,
	},
}

// MoonMaximumNorthDeclinations 指定年月内的所有月球最大北赤纬事件 / all maximum northern lunar declination events in the given Gregorian month.
func MoonMaximumNorthDeclinations(year int, month time.Month) []DeclinationEvent {
	return moonMaximumDeclinationsInMonth(year, month, moonMaxDeclinationNorthCoefficients)
}

// MoonMaximumSouthDeclinations 指定年月内的所有月球最大南赤纬事件 / all maximum southern lunar declination events in the given Gregorian month.
func MoonMaximumSouthDeclinations(year int, month time.Month) []DeclinationEvent {
	return moonMaximumDeclinationsInMonth(year, month, moonMaxDeclinationSouthCoefficients)
}

// LastMoonMaximumNorthDeclination 指定时刻之前最近一次月球最大北赤纬 / last maximum northern lunar declination at or before jd.
func LastMoonMaximumNorthDeclination(jd float64) DeclinationEvent {
	return moonMaximumDeclinationSearch(jd, moonMaxDeclinationNorthCoefficients, -1, true)
}

// NextMoonMaximumNorthDeclination 指定时刻之后最近一次月球最大北赤纬 / next maximum northern lunar declination after jd.
func NextMoonMaximumNorthDeclination(jd float64) DeclinationEvent {
	return moonMaximumDeclinationSearch(jd, moonMaxDeclinationNorthCoefficients, 1, false)
}

// ClosestMoonMaximumNorthDeclination 离指定时刻最近一次月球最大北赤纬 / closest maximum northern lunar declination to jd.
func ClosestMoonMaximumNorthDeclination(jd float64) DeclinationEvent {
	return moonClosestMaximumDeclination(jd, moonMaxDeclinationNorthCoefficients)
}

// LastMoonMaximumSouthDeclination 指定时刻之前最近一次月球最大南赤纬 / last maximum southern lunar declination at or before jd.
func LastMoonMaximumSouthDeclination(jd float64) DeclinationEvent {
	return moonMaximumDeclinationSearch(jd, moonMaxDeclinationSouthCoefficients, -1, true)
}

// NextMoonMaximumSouthDeclination 指定时刻之后最近一次月球最大南赤纬 / next maximum southern lunar declination after jd.
func NextMoonMaximumSouthDeclination(jd float64) DeclinationEvent {
	return moonMaximumDeclinationSearch(jd, moonMaxDeclinationSouthCoefficients, 1, false)
}

// ClosestMoonMaximumSouthDeclination 离指定时刻最近一次月球最大南赤纬 / closest maximum southern lunar declination to jd.
func ClosestMoonMaximumSouthDeclination(jd float64) DeclinationEvent {
	return moonClosestMaximumDeclination(jd, moonMaxDeclinationSouthCoefficients)
}

func moonMaximumDeclinationsInMonth(year int, month time.Month, coeffs moonMaxDeclinationCoefficients) []DeclinationEvent {
	startUTC := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	endUTC := startUTC.AddDate(0, 1, 0)
	startTT := TD2UT(Date2JDE(startUTC), true)
	endTT := TD2UT(Date2JDE(endUTC), true)

	kStart := int(math.Floor((startTT-coeffs.JDE0)/moonMaxDeclinationMeanMonthDays)) - 1
	kEnd := int(math.Ceil((endTT-coeffs.JDE0)/moonMaxDeclinationMeanMonthDays)) + 1

	cfg := apsisSearchConfig{
		bracketHalfWidth: moonApsisBracketHalfWidth,
		sampleStep:       moonApsisSampleStep,
		derivativeStep:   moonApsisDerivativeStep,
		toleranceDays:    moonApsisToleranceDays,
		maxIterations:    moonApsisMaxIterations,
		maximize:         coeffs.sign > 0,
	}

	events := make([]DeclinationEvent, 0, 2)
	for k := kStart; k <= kEnd; k++ {
		event := moonMaximumDeclinationEvent(k, coeffs, cfg)
		eventTimeUTC := JDE2DateByZone(event.JDE, time.UTC, false)
		if eventTimeUTC.Before(startUTC) || !eventTimeUTC.Before(endUTC) {
			continue
		}
		events = append(events, event)
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].JDE < events[j].JDE
	})
	return events
}

func moonMaximumDeclinationEvent(k int, coeffs moonMaxDeclinationCoefficients, cfg apsisSearchConfig) DeclinationEvent {
	seedTT := moonMaximumDeclinationSeedTT(k, coeffs)
	eventTT, declination := refineDistanceExtremum(seedTT, cfg, func(sampleTT float64) float64 {
		return HMoonTrueDecN(sampleTT, -1)
	})
	return DeclinationEvent{
		JDE:         TD2UT(eventTT, false),
		Declination: declination,
	}
}

func moonMaximumDeclinationSearch(jd float64, coeffs moonMaxDeclinationCoefficients, direction int, includeCurrent bool) DeclinationEvent {
	cfg := apsisSearchConfig{
		bracketHalfWidth: moonApsisBracketHalfWidth,
		sampleStep:       moonApsisSampleStep,
		derivativeStep:   moonApsisDerivativeStep,
		toleranceDays:    moonApsisToleranceDays,
		maxIterations:    moonApsisMaxIterations,
		maximize:         coeffs.sign > 0,
	}
	targetTT := TD2UT(jd, true)
	centerK := int(math.Round((targetTT - coeffs.JDE0) / moonMaxDeclinationMeanMonthDays))

	found := false
	bestDistance := math.Inf(1)
	var best DeclinationEvent
	for offset := -moonMaxDeclinationSearchSpan; offset <= moonMaxDeclinationSearchSpan; offset++ {
		event := moonMaximumDeclinationEvent(centerK+offset, coeffs, cfg)
		delta := event.JDE - jd
		if !moonMaximumDeclinationMatchesDirection(delta, direction, includeCurrent) {
			continue
		}
		distance := math.Abs(delta)
		if !found || distance < bestDistance || (distance == bestDistance && moonMaximumDeclinationEarlier(event, best)) {
			best = event
			bestDistance = distance
			found = true
		}
	}
	return best
}

func moonClosestMaximumDeclination(jd float64, coeffs moonMaxDeclinationCoefficients) DeclinationEvent {
	last := moonMaximumDeclinationSearch(jd, coeffs, -1, true)
	next := moonMaximumDeclinationSearch(jd, coeffs, 1, false)
	lastDistance := math.Abs(jd - last.JDE)
	nextDistance := math.Abs(next.JDE - jd)
	if lastDistance <= nextDistance {
		return last
	}
	return next
}

func moonMaximumDeclinationMatchesDirection(delta float64, direction int, includeCurrent bool) bool {
	switch direction {
	case -1:
		if includeCurrent {
			return delta <= 0
		}
		return delta < 0
	case 1:
		if includeCurrent {
			return delta >= 0
		}
		return delta > 0
	default:
		return true
	}
}

func moonMaximumDeclinationEarlier(a, b DeclinationEvent) bool {
	return a.JDE < b.JDE
}

func moonMaximumDeclinationSeedTT(k int, coeffs moonMaxDeclinationCoefficients) float64 {
	cycle := float64(k)
	T := cycle / moonMaxDeclinationBaseCycle
	D := Limit360(coeffs.D0 + 333.0705546*cycle - 0.0004214*T*T + 0.00000011*T*T*T)
	M := Limit360(coeffs.M0 + 26.9281592*cycle - 0.0000355*T*T - 0.0000001*T*T*T)
	MP := Limit360(coeffs.MP0 + 356.9562794*cycle + 0.0103066*T*T + 0.00001251*T*T*T)
	F := Limit360(coeffs.F0 + 1.4467807*cycle - 0.0020690*T*T - 0.00000215*T*T*T)
	E := 1 - 0.002516*T - 0.0000074*T*T

	return coeffs.JDE0 +
		moonMaxDeclinationMeanMonthDays*cycle +
		0.000119804*T*T -
		0.000000141*T*T*T +
		coeffs.tc[0]*Cos(F) +
		coeffs.tc[1]*Sin(MP) +
		coeffs.tc[2]*Sin(2*F) +
		coeffs.tc[3]*Sin(2*D-MP) +
		coeffs.tc[4]*Cos(MP-F) +
		coeffs.tc[5]*Cos(MP+F) +
		coeffs.tc[6]*Sin(2*D) +
		coeffs.tc[7]*Sin(M)*E +
		coeffs.tc[8]*Cos(3*F) +
		coeffs.tc[9]*Sin(MP+2*F) +
		coeffs.tc[10]*Cos(2*D-F) +
		coeffs.tc[11]*Cos(2*D-MP-F) +
		coeffs.tc[12]*Cos(2*D-MP+F) +
		coeffs.tc[13]*Cos(2*D+F) +
		coeffs.tc[14]*Sin(2*MP) +
		coeffs.tc[15]*Sin(MP-2*F) +
		coeffs.tc[16]*Cos(2*MP-F) +
		coeffs.tc[17]*Sin(MP+3*F) +
		coeffs.tc[18]*Sin(2*D-M-MP)*E +
		coeffs.tc[19]*Cos(MP-2*F) +
		coeffs.tc[20]*Sin(2*(D-MP)) +
		coeffs.tc[21]*Sin(F) +
		coeffs.tc[22]*Sin(2*D+MP) +
		coeffs.tc[23]*Cos(MP+2*F) +
		coeffs.tc[24]*Sin(2*D-M)*E +
		coeffs.tc[25]*Sin(MP+F) +
		coeffs.tc[26]*Sin(M-MP)*E +
		coeffs.tc[27]*Sin(MP-3*F) +
		coeffs.tc[28]*Sin(2*MP+F) +
		coeffs.tc[29]*Cos(2*(D-MP)-F) +
		coeffs.tc[30]*Sin(3*F) +
		coeffs.tc[31]*Cos(MP+3*F) +
		coeffs.tc[32]*Cos(2*MP) +
		coeffs.tc[33]*Cos(2*D-MP) +
		coeffs.tc[34]*Cos(2*D+MP+F) +
		coeffs.tc[35]*Cos(MP) +
		coeffs.tc[36]*Sin(3*MP+F) +
		coeffs.tc[37]*Sin(2*D-MP+F) +
		coeffs.tc[38]*Cos(2*(D-MP)) +
		coeffs.tc[39]*Cos(D+F) +
		coeffs.tc[40]*Sin(M+MP)*E +
		coeffs.tc[41]*Sin(2*(D-F)) +
		coeffs.tc[42]*Cos(2*MP+F) +
		coeffs.tc[43]*Cos(3*MP+F)
}
