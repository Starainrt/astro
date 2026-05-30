package formula

import "math"

const darkAdaptedPupilDiameterMM = 7.0

// LightGatheringPowerRatio 集光力比值 / light-gathering power ratio.
//
//	diameter1MM: 第一个望远镜口径，单位毫米
//	diameter2MM: 第二个望远镜口径，单位毫米
//
// 返回：
//
//	集光力比值，等于 (diameter1MM / diameter2MM)^2
//
// Returns the light-gathering power ratio, equal to (diameter1MM / diameter2MM)^2.
func LightGatheringPowerRatio(diameter1MM, diameter2MM float64) float64 {
	if diameter1MM <= 0 || diameter2MM <= 0 ||
		math.IsNaN(diameter1MM) || math.IsInf(diameter1MM, 0) ||
		math.IsNaN(diameter2MM) || math.IsInf(diameter2MM, 0) {
		return math.NaN()
	}
	return math.Pow(diameter1MM/diameter2MM, 2)
}

// DawesLimitArcsec Dawes 极限 / Dawes limit in arcseconds.
//
//	diameterMM: 望远镜口径，单位毫米
//
// 返回：
//
//	Dawes 极限，单位角秒
//
// Returns the Dawes limit in arcseconds for the supplied aperture.
func DawesLimitArcsec(diameterMM float64) float64 {
	if diameterMM <= 0 || math.IsNaN(diameterMM) || math.IsInf(diameterMM, 0) {
		return math.NaN()
	}
	return 116 / diameterMM
}

// RayleighLimitArcsec Rayleigh 极限 / Rayleigh limit in arcseconds.
//
//	diameterMM: 望远镜口径，单位毫米
//
// 返回：
//
//	Rayleigh 极限，单位角秒
//
// Returns the Rayleigh limit in arcseconds for the supplied aperture.
func RayleighLimitArcsec(diameterMM float64) float64 {
	if diameterMM <= 0 || math.IsNaN(diameterMM) || math.IsInf(diameterMM, 0) {
		return math.NaN()
	}
	return 138.4 / diameterMM
}

// LimitingMagnitudeEmpirical 经验极限星等 / empirical limiting magnitude.
//
//	diameterMM: 望远镜有效口径，单位毫米
//	nakedEyeLimit: 观测地裸眼极限星等，例如乡村暗空可近似取 6
//
// 返回：
//
//	经验极限星等；这是经验值，不包含天空背景、倍率、透过率和观测经验修正
//
// Returns an empirical limiting magnitude estimate. It does not account for sky background, magnification, transmission, or observer skill.
//
// 例：
//
//	// 70mm 小型折射镜，裸眼极限 6 等
//	mag := formula.LimitingMagnitudeEmpirical(70, 6)
func LimitingMagnitudeEmpirical(diameterMM, nakedEyeLimit float64) float64 {
	if diameterMM <= 0 ||
		math.IsNaN(diameterMM) || math.IsInf(diameterMM, 0) ||
		math.IsNaN(nakedEyeLimit) || math.IsInf(nakedEyeLimit, 0) {
		return math.NaN()
	}
	return nakedEyeLimit + 5*math.Log10(diameterMM/darkAdaptedPupilDiameterMM)
}
