package basic

import (
	"math"
	"sync"
)

type jupiterGalileanCorrectionSpan struct {
	StartJD    float64
	EndJD      float64
	CenterYear float64
	HalfYears  float64
}

type jupiterGalileanCorrectionTerm struct {
	Period float64
	Sin    float64
	Cos    float64
}

type jupiterGalileanElementCorrection struct {
	Polynomial [3]float64
	Terms      []jupiterGalileanCorrectionTerm
}

type jupiterGalileanFoldedCorrection struct {
	Index   int
	RealSin float64
	RealCos float64
	ImagSin float64
	ImagCos float64
}

var jupiterGalileanFoldedCorrections [4]struct {
	once   sync.Once
	series [2][4][]jupiterGalileanFoldedCorrection
}

func jupiterGalileanFoldCorrections(index int) {
	cache := &jupiterGalileanFoldedCorrections[index]
	for block, element := range []int{-1, 0, 1, 3} {
		if element < 0 {
			continue
		}
		models := jupiterGalileanCorrections[index]
		used := make(map[float64]bool)
		for family, blocks := range [2][4][]jupiterGalileanL1Term{jupiterGalileanL1LongPeriodTerms[index], jupiterGalileanL1CrossPeriodTerms[index]} {
			for position, source := range blocks[block] {
				period := math.Abs(source.Period)
				if used[period] {
					continue
				}
				for termIndex, term := range models[element].Terms {
					if term.Period != period {
						continue
					}
					phase := source.Phase
					if family == 0 {
						phase += 2 * math.Pi * jupiterGalileanLongPeriodShift / source.Period
					}
					sine, cosine := math.Sincos(phase)
					sign := math.Copysign(1, source.Period)
					folded := jupiterGalileanFoldedCorrection{
						Index:   position,
						RealSin: sign*term.Sin*cosine + term.Cos*sine,
						RealCos: term.Cos*cosine - sign*term.Sin*sine,
					}
					if element > 0 {
						partner := models[element+1].Terms[termIndex]
						folded.ImagSin = sign*partner.Sin*cosine + partner.Cos*sine
						folded.ImagCos = partner.Cos*cosine - sign*partner.Sin*sine
					}
					cache.series[family][block] = append(cache.series[family][block], folded)
					used[period] = true
					break
				}
			}
		}
	}
}

func jupiterGalileanFoldedElementsAtET(et float64, index int, includeSolarLongPeriod bool) jupiterGalileanElements {
	span := jupiterGalileanCorrectionSpans[index]
	weight := span.window(et)
	if weight == 0 {
		return jupiterGalileanElementsAtET(et, index, includeSolarLongPeriod)
	}
	cache := &jupiterGalileanFoldedCorrections[index]
	cache.once.Do(func() { jupiterGalileanFoldCorrections(index) })
	longPeriod := jupiterGalileanEvaluateFoldedSeries(jupiterGalileanL1LongPeriodTerms[index], cache.series[0], weight, et+jupiterGalileanLongPeriodShift, et, includeSolarLongPeriod, index)
	crossPeriod := jupiterGalileanEvaluateFoldedSeries(jupiterGalileanL1CrossPeriodTerms[index], cache.series[1], weight, et, et, false, index)
	tau := span.normalizedTime(et)
	var polynomial [5]float64
	for element, model := range jupiterGalileanCorrections[index] {
		polynomial[element] = weight * (model.Polynomial[0] + model.Polynomial[1]*tau + model.Polynomial[2]*(2*tau*tau-1))
	}
	elements := jupiterGalileanElements{
		A: longPeriod.A + crossPeriod.A,
		L: longPeriod.L + crossPeriod.L + jupiterGalileanBaseMeanLongitudes[index]*et + polynomial[0],
		K: longPeriod.K + crossPeriod.K + polynomial[1],
		H: longPeriod.H + crossPeriod.H + polynomial[2],
		Q: longPeriod.Q + crossPeriod.Q + polynomial[3],
		P: longPeriod.P + crossPeriod.P + polynomial[4],
	}
	elements.L = math.Atan2(math.Sin(elements.L), math.Cos(elements.L))
	if elements.L < 0 {
		elements.L += 2 * math.Pi
	}
	return elements
}

func jupiterGalileanEvaluateFoldedSeries(blocks [4][]jupiterGalileanL1Term, corrections [4][]jupiterGalileanFoldedCorrection, weight, angleTime, et float64, includeSolarLongPeriod bool, index int) jupiterGalileanElements {
	vals := jupiterGalileanSolarTermsAtET(et, includeSolarLongPeriod, index)
	result := jupiterGalileanElements{}
	for blockIndex, terms := range blocks {
		realPart, imagPart := 0.0, 0.0
		folded := corrections[blockIndex]
		cursor := 0
		for position, term := range terms {
			angle := term.Phase
			if term.Period != 0 {
				angle += 2 * math.Pi * angleTime / term.Period
			}
			sine, cosine := math.Sincos(angle)
			realPart += term.Amp * cosine
			imagPart += term.Amp * sine
			if cursor < len(folded) && folded[cursor].Index == position {
				correction := folded[cursor]
				realPart += weight * (correction.RealSin*sine + correction.RealCos*cosine)
				imagPart += weight * (correction.ImagSin*sine + correction.ImagCos*cosine)
				cursor++
			}
		}
		switch blockIndex {
		case 0:
			result.A = realPart
		case 1:
			result.L = realPart + vals[0]
		case 2:
			result.K = realPart + vals[1]
			result.H = imagPart + vals[2]
		case 3:
			result.Q = realPart + vals[3]
			result.P = imagPart + vals[4]
		}
	}
	return result
}

func (span jupiterGalileanCorrectionSpan) window(et float64) float64 {
	jd := et + jupiterGalileanReferenceJD
	distance := 0.0
	if jd < span.StartJD {
		distance = span.StartJD - jd
	} else if jd > span.EndJD {
		distance = jd - span.EndJD
	}
	if distance == 0 {
		return 1
	}
	if distance >= 36525 {
		return 0
	}
	return 0.5 * (1 + math.Cos(math.Pi*distance/36525))
}

func (span jupiterGalileanCorrectionSpan) normalizedTime(et float64) float64 {
	year := 2000 + (et+jupiterGalileanReferenceJD-2451545)/365.25
	return math.Max(-1, math.Min(1, (year-span.CenterYear)/span.HalfYears))
}
