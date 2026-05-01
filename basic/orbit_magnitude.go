package basic

import "math"

// OrbitAsteroidMagnitudeHG 返回小行星 H-G 模型的视星等。
func OrbitAsteroidMagnitudeHG(jd float64, elements OrbitElements, absoluteMagnitude, slopeParameter float64) float64 {
	if !isFinite(jd) || !isFinite(absoluteMagnitude) || !isFinite(slopeParameter) {
		return math.NaN()
	}

	sunDistance := OrbitSunDistance(jd, elements)
	earthDistance := OrbitEarthDistance(jd, elements)
	phaseAngle := OrbitPhaseAngle(jd, elements)
	if !isFinitePositive(sunDistance) || !isFinitePositive(earthDistance) || !isFinite(phaseAngle) {
		return math.NaN()
	}

	phaseBlend := orbitHGSlopeBlend(phaseAngle, slopeParameter)
	if phaseBlend == 0 {
		return math.Inf(1)
	}
	if !isFinitePositive(phaseBlend) {
		return math.NaN()
	}

	return absoluteMagnitude + 5*math.Log10(sunDistance*earthDistance) - 2.5*math.Log10(phaseBlend)
}

func orbitHGSlopeBlend(phaseAngle, slopeParameter float64) float64 {
	phi1 := orbitHGPhaseFunction1(phaseAngle)
	phi2 := orbitHGPhaseFunction2(phaseAngle)
	return (1-slopeParameter)*phi1 + slopeParameter*phi2
}

func orbitHGPhaseFunction1(phaseAngle float64) float64 {
	return math.Exp(-3.33 * math.Pow(orbitHGTanHalfPhaseAngle(phaseAngle), 0.63))
}

func orbitHGPhaseFunction2(phaseAngle float64) float64 {
	return math.Exp(-1.87 * math.Pow(orbitHGTanHalfPhaseAngle(phaseAngle), 1.22))
}

func orbitHGTanHalfPhaseAngle(phaseAngle float64) float64 {
	return math.Tan((phaseAngle * math.Pi / 180) / 2)
}
