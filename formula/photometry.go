package formula

import "math"

// DistanceModulus 距离模数 / distance modulus.
//
//	distanceParsec: 天体距离，单位秒差距 pc
//
// 返回：
//
//	距离模数 m-M
//
// Returns the distance modulus m-M for the supplied distance in parsecs.
func DistanceModulus(distanceParsec float64) float64 {
	if distanceParsec <= 0 || math.IsNaN(distanceParsec) || math.IsInf(distanceParsec, 0) {
		return math.NaN()
	}
	return 5 * math.Log10(distanceParsec/10)
}

// ApparentMagnitudeFromAbsolute 由绝对星等求视星等 / apparent magnitude from absolute magnitude.
//
//	absoluteMagnitude: 绝对星等 M
//	distanceParsec: 天体距离，单位秒差距 pc
//
// 返回：
//
//	视星等 m
//
// Returns apparent magnitude from absolute magnitude and distance.
func ApparentMagnitudeFromAbsolute(absoluteMagnitude, distanceParsec float64) float64 {
	modulus := DistanceModulus(distanceParsec)
	if math.IsNaN(modulus) {
		return math.NaN()
	}
	return absoluteMagnitude + modulus
}

// AbsoluteMagnitudeFromApparent 由视星等求绝对星等 / absolute magnitude from apparent magnitude.
//
//	apparentMagnitude: 视星等 m
//	distanceParsec: 天体距离，单位秒差距 pc
//
// 返回：
//
//	绝对星等 M
//
// Returns absolute magnitude from apparent magnitude and distance.
func AbsoluteMagnitudeFromApparent(apparentMagnitude, distanceParsec float64) float64 {
	modulus := DistanceModulus(distanceParsec)
	if math.IsNaN(modulus) {
		return math.NaN()
	}
	return apparentMagnitude - modulus
}
