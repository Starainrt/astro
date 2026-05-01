package basic

import (
	"math"

	. "github.com/starainrt/astro/tools"
)

// OrbitMeanMotion 返回历元平均角速度，单位度/日。
// 对近日点形式的抛物/双曲轨道返回 NaN。
func OrbitMeanMotion(elements OrbitElements) float64 {
	if elements.usesPerihelionForm() {
		if !elements.validPerihelionForm() || elements.E >= 1 {
			return math.NaN()
		}
		semiMajorAxis := elements.Q / (1 - elements.E)
		if !isFinitePositive(semiMajorAxis) {
			return math.NaN()
		}
		if isFinite(elements.MDot) && elements.MDot != 0 {
			return elements.MDot
		}
		return gaussianGravitationalConstant / math.Pow(semiMajorAxis, 1.5) * deg
	}
	if !elements.validEllipticClassical() {
		return math.NaN()
	}
	if isFinite(elements.MDot) && elements.MDot != 0 {
		return elements.MDot
	}
	return gaussianGravitationalConstant / math.Pow(elements.A, 1.5) * deg
}

// OrbitMeanAnomaly 返回给定 TT/TDB 儒略日的平近点角，单位度。
// 对抛物/双曲轨道返回 NaN。
func OrbitMeanAnomaly(jd float64, elements OrbitElements) float64 {
	meanAnomalyDeg, ok := orbitMeanAnomalyDegAt(jd, elements)
	if !ok {
		return math.NaN()
	}
	return Limit360(meanAnomalyDeg)
}

// OrbitEccentricAnomaly 返回给定 TT/TDB 儒略日的偏近点角，单位度。
// 仅适用于椭圆轨道；对抛物/双曲轨道返回 NaN。
func OrbitEccentricAnomaly(jd float64, elements OrbitElements) float64 {
	resolved := orbitElementsAt(jd, elements)
	meanAnomalyDeg, ok := orbitMeanAnomalyDegAt(jd, elements)
	if !ok || !resolved.validEllipticClassical() && !(resolved.validPerihelionForm() && resolved.E < 1) {
		return math.NaN()
	}
	eccentricAnomaly, ok := orbitEccentricAnomalyRad(meanAnomalyDeg*rad, resolved.E)
	if !ok {
		return math.NaN()
	}
	return Limit360(eccentricAnomaly * deg)
}

// OrbitTrueAnomaly 返回给定 TT/TDB 儒略日的真近点角，单位度。
func OrbitTrueAnomaly(jd float64, elements OrbitElements) float64 {
	trueAnomaly, _, _, ok := orbitTrueAnomalyAndRadius(jd, elements)
	if !ok {
		return math.NaN()
	}
	return Limit360(trueAnomaly * deg)
}

func orbitMeanAnomalyDegAt(jd float64, elements OrbitElements) (float64, bool) {
	if !isFinite(jd) {
		return math.NaN(), false
	}
	if elements.usesPerihelionForm() {
		if !elements.validPerihelionForm() || elements.E >= 1 {
			return math.NaN(), false
		}
		semiMajorAxis := elements.Q / (1 - elements.E)
		if !isFinitePositive(semiMajorAxis) {
			return math.NaN(), false
		}
		meanMotion := elements.MDot
		if !isFinite(meanMotion) || meanMotion == 0 {
			meanMotion = gaussianGravitationalConstant / math.Pow(semiMajorAxis, 1.5) * deg
		}
		return meanMotion * (jd - elements.TpJD), true
	}

	resolved := orbitElementsAt(jd, elements)
	if !resolved.validEllipticClassical() {
		return math.NaN(), false
	}
	if isFinite(elements.MDot) && elements.MDot != 0 {
		return elements.M0 + elements.MDot*(jd-elements.EpochJD), true
	}
	meanMotion := gaussianGravitationalConstant / math.Pow(resolved.A, 1.5) * deg
	return resolved.M0 + meanMotion*(jd-elements.EpochJD), true
}

func orbitEccentricAnomalyRad(meanAnomalyRad, eccentricity float64) (float64, bool) {
	if !isFinite(meanAnomalyRad) || !isFinite(eccentricity) || eccentricity < 0 || eccentricity >= 1 {
		return math.NaN(), false
	}
	if meanAnomalyRad > math.Pi {
		meanAnomalyRad -= 2 * math.Pi
	} else if meanAnomalyRad < -math.Pi {
		meanAnomalyRad += 2 * math.Pi
	}

	eccentricAnomaly := meanAnomalyRad
	if eccentricity >= 0.8 {
		eccentricAnomaly = math.Pi
		if meanAnomalyRad < 0 {
			eccentricAnomaly = -math.Pi
		}
	}

	for i := 0; i < 32; i++ {
		sinE, cosE := math.Sincos(eccentricAnomaly)
		delta := (eccentricAnomaly - eccentricity*sinE - meanAnomalyRad) / (1 - eccentricity*cosE)
		eccentricAnomaly -= delta
		if math.Abs(delta) < 1e-14 {
			return eccentricAnomaly, true
		}
	}
	return eccentricAnomaly, true
}

func orbitHyperbolicAnomaly(meanAnomaly, eccentricity float64) (float64, bool) {
	if !isFinite(meanAnomaly) || !isFinite(eccentricity) || eccentricity <= 1 {
		return math.NaN(), false
	}

	hyperbolicAnomaly := math.Asinh(meanAnomaly / eccentricity)
	if hyperbolicAnomaly == 0 && meanAnomaly != 0 {
		hyperbolicAnomaly = math.Log(2*math.Abs(meanAnomaly)/eccentricity + 1.8)
		if meanAnomaly < 0 {
			hyperbolicAnomaly = -hyperbolicAnomaly
		}
	}

	for i := 0; i < 32; i++ {
		sinhH := math.Sinh(hyperbolicAnomaly)
		coshH := math.Cosh(hyperbolicAnomaly)
		delta := (eccentricity*sinhH - hyperbolicAnomaly - meanAnomaly) / (eccentricity*coshH - 1)
		hyperbolicAnomaly -= delta
		if math.Abs(delta) < 1e-14 {
			return hyperbolicAnomaly, true
		}
	}
	return hyperbolicAnomaly, true
}

func orbitTrueAnomalyAndRadius(jd float64, elements OrbitElements) (trueAnomaly, radius float64, resolved OrbitElements, ok bool) {
	resolved = orbitElementsAt(jd, elements)
	if resolved.usesPerihelionForm() {
		if !resolved.validPerihelionForm() {
			return math.NaN(), math.NaN(), resolved, false
		}
		switch {
		case math.Abs(resolved.E-1) <= orbitParabolicTolerance:
			return orbitParabolicTrueAnomalyAndRadius(jd, resolved)
		case resolved.E < 1:
			return orbitEllipticTrueAnomalyAndRadiusFromPerihelion(jd, resolved)
		default:
			return orbitHyperbolicTrueAnomalyAndRadius(jd, resolved)
		}
	}
	if !resolved.validEllipticClassical() {
		return math.NaN(), math.NaN(), resolved, false
	}
	meanAnomalyDeg, ok := orbitMeanAnomalyDegAt(jd, elements)
	if !ok {
		return math.NaN(), math.NaN(), resolved, false
	}
	trueAnomaly, radius, ok = orbitEllipticTrueAnomalyAndRadius(meanAnomalyDeg*rad, resolved.A, resolved.E)
	return trueAnomaly, radius, resolved, ok
}

func orbitEllipticTrueAnomalyAndRadiusFromPerihelion(jd float64, elements OrbitElements) (trueAnomaly, radius float64, resolved OrbitElements, ok bool) {
	semiMajorAxis := elements.Q / (1 - elements.E)
	if !isFinitePositive(semiMajorAxis) {
		return math.NaN(), math.NaN(), elements, false
	}
	meanAnomalyDeg, ok := orbitMeanAnomalyDegAt(jd, elements)
	if !ok {
		return math.NaN(), math.NaN(), elements, false
	}
	trueAnomaly, radius, ok = orbitEllipticTrueAnomalyAndRadius(meanAnomalyDeg*rad, semiMajorAxis, elements.E)
	if !ok {
		return math.NaN(), math.NaN(), elements, false
	}
	resolved = elements
	resolved.A = semiMajorAxis
	return trueAnomaly, radius, resolved, true
}

func orbitEllipticTrueAnomalyAndRadius(meanAnomalyRad, semiMajorAxis, eccentricity float64) (float64, float64, bool) {
	eccentricAnomaly, ok := orbitEccentricAnomalyRad(meanAnomalyRad, eccentricity)
	if !ok {
		return math.NaN(), math.NaN(), false
	}
	sinE, cosE := math.Sincos(eccentricAnomaly)
	radius := semiMajorAxis * (1 - eccentricity*cosE)
	trueAnomaly := math.Atan2(math.Sqrt(1-eccentricity*eccentricity)*sinE, cosE-eccentricity)
	return trueAnomaly, radius, true
}

func orbitParabolicTrueAnomalyAndRadius(jd float64, elements OrbitElements) (trueAnomaly, radius float64, resolved OrbitElements, ok bool) {
	if !isFinitePositive(elements.Q) || !isFinite(elements.TpJD) {
		return math.NaN(), math.NaN(), elements, false
	}
	w := 1.5 * gaussianGravitationalConstant * (jd - elements.TpJD) / (math.Sqrt2 * math.Pow(elements.Q, 1.5))
	y := math.Cbrt(w + math.Sqrt(w*w+1))
	if y == 0 {
		return 0, elements.Q, elements, true
	}
	d := y - 1/y
	trueAnomaly = 2 * math.Atan(d)
	radius = elements.Q * (1 + d*d)
	resolved = elements
	return trueAnomaly, radius, resolved, true
}

func orbitHyperbolicTrueAnomalyAndRadius(jd float64, elements OrbitElements) (trueAnomaly, radius float64, resolved OrbitElements, ok bool) {
	if !isFinitePositive(elements.Q) || !isFinite(elements.TpJD) || !isFinite(elements.E) || elements.E <= 1 {
		return math.NaN(), math.NaN(), elements, false
	}
	semiMajorAxis := elements.Q / (elements.E - 1)
	meanAnomaly := gaussianGravitationalConstant * (jd - elements.TpJD) / math.Pow(semiMajorAxis, 1.5)
	hyperbolicAnomaly, ok := orbitHyperbolicAnomaly(meanAnomaly, elements.E)
	if !ok {
		return math.NaN(), math.NaN(), elements, false
	}
	radius = semiMajorAxis * (elements.E*math.Cosh(hyperbolicAnomaly) - 1)
	trueAnomaly = 2 * math.Atan(math.Sqrt((elements.E+1)/(elements.E-1))*math.Tanh(hyperbolicAnomaly/2))
	resolved = elements
	resolved.A = -semiMajorAxis
	return trueAnomaly, radius, resolved, true
}
