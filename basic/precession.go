package basic

import (
	"math"

	"github.com/starainrt/astro/internal/ltp"
)

// Precess 函数实现了基于 Explanatory Supplement 2nd ed. table 3.211.1 的岁差计算
// ra0, dec0: 初始赤经和赤纬（弧度）
// startJD: 初始儒略日
// targetJD: 目标儒略日
// 返回: 目标历元的赤经和赤纬（弧度）
func PrecessIAU1976(ra0, dec0, startJD, targetJD float64) (float64, float64) {
	// 定义常量
	ra0 *= rad
	dec0 *= rad
	t := (targetJD - startJD) / 36525.0
	t0 := (startJD - 2451545.0) / 36525.0

	// Explanatory Supplement 2nd ed. table 3.211.1
	zeta := ((2306.2181+1.39656*t0-0.000139*t0*t0)*t +
		(0.30188-0.000344*t0)*t*t + 0.017998*t*t*t) / 3600.0 * rad

	z := ((2306.2181+1.39656*t0-0.000139*t0*t0)*t +
		(1.09468+0.000066*t0)*t*t + 0.018203*t*t*t) / 3600.0 * rad

	theta := ((2004.3109-0.85330*t0-0.000217*t0*t0)*t -
		(0.42665+0.000217*t0)*t*t - 0.041833*t*t*t) / 3600.0 * rad

	termA := math.Cos(dec0) * math.Sin(ra0+z)
	termB := math.Cos(theta)*math.Cos(dec0)*math.Cos(ra0+z) - math.Sin(theta)*math.Sin(dec0)
	termC := math.Sin(theta)*math.Cos(dec0)*math.Cos(ra0+z) + math.Cos(theta)*math.Sin(dec0)

	raMinZ := math.Atan2(termA, termB)
	dec := math.Asin(termC)
	ra := raMinZ + zeta

	return ra / rad, dec / rad
}

// 常量定义
const (
	AS2R = 4.848136811095359935899141e-6 // 角秒到弧度的转换因子
	D2PI = 6.283185307179586476925287    // 2π
	EPS0 = 84381.406 * AS2R              // J2000.0历元的黄赤交角(弧度)
)

// 赤道坐标结构体
type Equatorial struct {
	RA  float64 // 赤经(度)
	Dec float64 // 赤纬(度)
}

// 三维向量
type Vector3 [3]float64

// 三维矩阵
type Matrix3 [3][3]float64

func ltpPECL(epj float64) Vector3 { return Vector3(ltp.EclipticPole(epj)) }

func ltpPEQU(epj float64) Vector3 { return Vector3(ltp.EquatorPole(epj)) }

func pxp(a, b Vector3) Vector3 { return Vector3(ltp.Cross(ltp.Vector3(a), ltp.Vector3(b))) }

func pn(p Vector3) (Vector3, float64) {
	u, w := ltp.Normalize(ltp.Vector3(p))
	return Vector3(u), w
}

// 计算长期岁差矩阵
func ltpPMAT(epj float64) Matrix3 {
	// 计算赤道极向量和黄极向量
	peqr := ltpPEQU(epj)
	pecl := ltpPECL(epj)

	// 计算春分点向量
	v := pxp(peqr, pecl)
	eqx, _ := pn(v)

	// 计算中间行向量
	v = pxp(peqr, eqx)

	// 形成旋转矩阵
	return Matrix3{
		{eqx[0], eqx[1], eqx[2]},
		{v[0], v[1], v[2]},
		{peqr[0], peqr[1], peqr[2]},
	}
}

// 计算包含GCRS框架偏差的长期岁差矩阵
func ltpPBMAT(epj float64) Matrix3 {
	// 框架偏差 (IERS Conventions 2010, Eqs. 5.21 and 5.33)
	dx := -0.016617 * AS2R
	de := -0.0068192 * AS2R
	dr := -0.0146 * AS2R

	// 计算基本岁差矩阵
	rp := ltpPMAT(epj)

	// 应用偏差
	return Matrix3{
		{
			rp[0][0] - rp[0][1]*dr + rp[0][2]*dx,
			rp[0][0]*dr + rp[0][1] + rp[0][2]*de,
			-rp[0][0]*dx - rp[0][1]*de + rp[0][2],
		},
		{
			rp[1][0] - rp[1][1]*dr + rp[1][2]*dx,
			rp[1][0]*dr + rp[1][1] + rp[1][2]*de,
			-rp[1][0]*dx - rp[1][1]*de + rp[1][2],
		},
		{
			rp[2][0] - rp[2][1]*dr + rp[2][2]*dx,
			rp[2][0]*dr + rp[2][1] + rp[2][2]*de,
			-rp[2][0]*dx - rp[2][1]*de + rp[2][2],
		},
	}
}

// 将赤道坐标转换为直角坐标向量
func raDecToVector(ra, dec float64) Vector3 {
	raRad := ra * math.Pi / 180
	decRad := dec * math.Pi / 180
	cosDec := math.Cos(decRad)
	return Vector3{
		math.Cos(raRad) * cosDec,
		math.Sin(raRad) * cosDec,
		math.Sin(decRad),
	}
}

// 将直角坐标向量转换为赤道坐标
func vectorToRaDec(v Vector3) (float64, float64) {
	dec := math.Asin(v[2]) * 180 / math.Pi

	ra := math.Atan2(v[1], v[0]) * 180 / math.Pi
	if ra < 0 {
		ra += 360
	}

	return ra, dec
}

// 应用长期岁差转换
func Precess(ra, dec, jdFrom, jdTo float64) (float64, float64) {
	// 将儒略日转换为儒略纪元
	epjFrom := 2000.0 + (jdFrom-2451545.0)/365.25
	epjTo := 2000.0 + (jdTo-2451545.0)/365.25

	// 计算从起始历元到J2000.0的逆矩阵
	rpFrom := ltpPMAT(epjFrom)
	rpFromInv := Matrix3{
		{rpFrom[0][0], rpFrom[1][0], rpFrom[2][0]},
		{rpFrom[0][1], rpFrom[1][1], rpFrom[2][1]},
		{rpFrom[0][2], rpFrom[1][2], rpFrom[2][2]},
	}

	// 计算从J2000.0到目标历元的矩阵
	rpTo := ltpPMAT(epjTo)

	// 计算复合旋转矩阵
	var rpFinal Matrix3
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			for k := 0; k < 3; k++ {
				rpFinal[i][j] += rpTo[i][k] * rpFromInv[k][j]
			}
		}
	}

	// 将赤道坐标转换为直角坐标向量
	vector := raDecToVector(ra, dec)

	// 应用旋转矩阵
	rotatedVector := Vector3{
		rpFinal[0][0]*vector[0] + rpFinal[0][1]*vector[1] + rpFinal[0][2]*vector[2],
		rpFinal[1][0]*vector[0] + rpFinal[1][1]*vector[1] + rpFinal[1][2]*vector[2],
		rpFinal[2][0]*vector[0] + rpFinal[2][1]*vector[1] + rpFinal[2][2]*vector[2],
	}

	// 将直角坐标向量转换回赤道坐标
	return vectorToRaDec(rotatedVector)
}
