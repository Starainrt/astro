package basic

import "github.com/starainrt/astro/internal/iau2006"

// RotationMatrix 3×3 旋转矩阵，按行存放，作用于列向量 / 3×3 rotation matrix, row-major, acting on column vectors.
type RotationMatrix = iau2006.Matrix

// Nutation2000A IAU 2000A 章动，返回 (黄经章动, 交角章动)，单位度；不进入默认计算链 / IAU 2000A nutation in degrees, opt-in only.
func Nutation2000A(jde float64) (float64, float64) {
	dpsi, deps := iau2006.Nutation(jde)
	return dpsi / 3600, deps / 3600
}

// MeanObliquity2006 IAU 2006 平黄赤交角（度）/ IAU 2006 mean obliquity in degrees.
func MeanObliquity2006(jde float64) float64 { return iau2006.MeanObliquity(jde) / 3600 }

// EquationOfEquinoxes2006 春分点方程主项 Δψ·cos εA（度），按 IAU 2000A 章动 / equation of the equinoxes in degrees, IAU 2000A.
func EquationOfEquinoxes2006(jde float64) float64 { return iau2006.EquationOfEquinoxes(jde) / 3600 }

// Precession2006Matrix J2000 平赤道到历元平赤道的 IAU 2006 岁差矩阵 / IAU 2006 precession matrix.
func Precession2006Matrix(jde float64) RotationMatrix { return iau2006.Precession(jde) }

// Nutation2000AMatrix 历元平赤道到真赤道的 IAU 2000A 章动矩阵 / IAU 2000A nutation matrix.
func Nutation2000AMatrix(jde float64) RotationMatrix { return iau2006.NutationMatrix(jde) }

// FrameBiasMatrix ICRS 到 J2000 平赤道的框架偏差矩阵 / frame bias from ICRS to the J2000 mean equator.
func FrameBiasMatrix() RotationMatrix { return iau2006.FrameBias() }

// BiasPrecessionNutation2006 ICRS 到真赤道真春分点的完整旋转 N·P·B / full ICRS to true-of-date rotation.
func BiasPrecessionNutation2006(jde float64) RotationMatrix {
	return iau2006.BiasPrecessionNutation(jde)
}
