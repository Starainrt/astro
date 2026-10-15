// Package iau2006 IAU 2006 岁差、IAU 2000A 章动与 ICRS 到真赤道的旋转；系数表单独成包，只有调用时才链接。
package iau2006

import "math"

// Matrix 3×3 旋转矩阵，按行存放，作用于列向量 / 3×3 rotation matrix, row-major, acting on column vectors.
type Matrix [3][3]float64

const (
	j2000        = 2451545.0
	arcsecToRad  = math.Pi / 648000
	arcsecCircle = 1296000.0
	obliquityJ2K = 84381.406
)

type lunisolarTerm struct {
	n                       [5]int8
	lonSin, lonSinT, lonCos float64
	oblCos, oblCosT, oblSin float64
}

type planetaryTerm struct {
	n              [14]int8
	lonSin, lonCos float64
	oblSin, oblCos float64
}

func centuries(jdeTT float64) float64 { return (jdeTT - j2000) / 36525 }

// Nutation IAU 2000A 黄经章动 Δψ 与交角章动 Δε（角秒），jdeTT 为 TT 儒略日 / nutation in longitude and obliquity in arcseconds.
func Nutation(jdeTT float64) (deltaPsi, deltaEpsilon float64) {
	t := centuries(jdeTT)
	var fa [5]float64
	for i, p := range fundamentalPolynomials {
		v := (((p[4]*t+p[3])*t+p[2])*t+p[1])*t + p[0]
		fa[i] = math.Mod(v, arcsecCircle) * arcsecToRad
	}
	var dpsi, deps float64
	for _, term := range lunisolarTerms {
		arg := 0.0
		for i, n := range term.n {
			arg += float64(n) * fa[i]
		}
		s, c := math.Sincos(arg)
		dpsi += s*(term.lonSin+term.lonSinT*t) + c*term.lonCos
		deps += c*(term.oblCos+term.oblCosT*t) + s*term.oblSin
	}
	var pa [14]float64
	for i, p := range planetaryArguments {
		pa[i] = p[0] + p[1]*t
	}
	pa[13] *= t
	for _, term := range planetaryTerms {
		arg := 0.0
		for i, n := range term.n {
			if n != 0 {
				arg += float64(n) * pa[i]
			}
		}
		s, c := math.Sincos(arg)
		dpsi += s*term.lonSin + c*term.lonCos
		deps += s*term.oblSin + c*term.oblCos
	}
	// 系数单位为 0.1 微角秒。
	return dpsi * 1e-7, deps * 1e-7
}

// MeanObliquity IAU 2006 平黄赤交角（角秒）/ IAU 2006 mean obliquity in arcseconds.
func MeanObliquity(jdeTT float64) float64 {
	t := centuries(jdeTT)
	return ((((-0.0000000434*t-0.000000576)*t+0.00200340)*t-0.0001831)*t-46.836769)*t + obliquityJ2K
}

// Precession J2000 平赤道到历元平赤道的 IAU 2006 岁差矩阵 / IAU 2006 precession matrix from J2000 to the mean equator of date.
func Precession(jdeTT float64) Matrix {
	t := centuries(jdeTT)
	psi := ((((-0.0000000951*t+0.000132851)*t-0.00114045)*t-1.0790069)*t + 5038.481507) * t
	omega := ((((0.0000003337*t-0.000000467)*t-0.00772503)*t+0.0512623)*t-0.025754)*t + obliquityJ2K
	chi := ((((-0.0000000560*t+0.000170663)*t-0.00121197)*t-2.3814292)*t + 10.556403) * t
	// R3(χ) R1(−ω) R3(−ψ) R1(ε0)
	return multiply(rotZ(chi*arcsecToRad), multiply(rotX(-omega*arcsecToRad), multiply(rotZ(-psi*arcsecToRad), rotX(obliquityJ2K*arcsecToRad))))
}

// NutationMatrix 历元平赤道到真赤道真春分点的章动矩阵 / nutation matrix from mean to true equator of date.
func NutationMatrix(jdeTT float64) Matrix {
	dpsi, deps := Nutation(jdeTT)
	mean := MeanObliquity(jdeTT) * arcsecToRad
	// R1(−εA−Δε) R3(−Δψ) R1(εA)
	return multiply(rotX(-(mean + deps*arcsecToRad)), multiply(rotZ(-dpsi*arcsecToRad), rotX(mean)))
}

// FrameBias ICRS 到 J2000 平赤道的框架偏差矩阵 / frame bias from ICRS to the J2000 mean equator.
func FrameBias() Matrix {
	const (
		xi0  = -0.0166170 * arcsecToRad
		eta0 = -0.0068192 * arcsecToRad
		da0  = -0.01460 * arcsecToRad
	)
	return Matrix{
		{1 - 0.5*(da0*da0+xi0*xi0), da0, -xi0},
		{-da0, 1 - 0.5*(da0*da0+eta0*eta0), -eta0},
		{xi0, eta0, 1 - 0.5*(eta0*eta0+xi0*xi0)},
	}
}

// BiasPrecessionNutation ICRS 到真赤道真春分点的完整旋转 N·P·B / full rotation from ICRS to the true equator and equinox of date.
func BiasPrecessionNutation(jdeTT float64) Matrix {
	return multiply(NutationMatrix(jdeTT), multiply(Precession(jdeTT), FrameBias()))
}

// EquationOfEquinoxes 春分点方程的主项 Δψ·cos εA（角秒），不含补充项 / equation of the equinoxes, principal term only.
func EquationOfEquinoxes(jdeTT float64) float64 {
	dpsi, _ := Nutation(jdeTT)
	return dpsi * math.Cos(MeanObliquity(jdeTT)*arcsecToRad)
}

// Apply 把矩阵作用于向量 / apply the matrix to a vector.
func (m Matrix) Apply(v [3]float64) [3]float64 {
	var out [3]float64
	for i := range out {
		out[i] = m[i][0]*v[0] + m[i][1]*v[1] + m[i][2]*v[2]
	}
	return out
}

// Transpose 逆旋转 / inverse rotation.
func (m Matrix) Transpose() Matrix {
	var out Matrix
	for i := range out {
		for j := range out[i] {
			out[i][j] = m[j][i]
		}
	}
	return out
}

func rotX(a float64) Matrix {
	s, c := math.Sincos(a)
	return Matrix{{1, 0, 0}, {0, c, s}, {0, -s, c}}
}

func rotZ(a float64) Matrix {
	s, c := math.Sincos(a)
	return Matrix{{c, s, 0}, {-s, c, 0}, {0, 0, 1}}
}

func multiply(a, b Matrix) Matrix {
	var out Matrix
	for i := range out {
		for j := range out[i] {
			out[i][j] = a[i][0]*b[0][j] + a[i][1]*b[1][j] + a[i][2]*b[2][j]
		}
	}
	return out
}
