package basic

import "math"

// Precess 函数实现了基于 Explanatory Supplement 2nd ed. table 3.211.1 的岁差计算
// ra0, dec0: 初始赤经和赤纬（弧度）
// JD0: 初始儒略日
// JD: 目标儒略日
// 返回: 目标历元的赤经和赤纬（弧度）
func PrecessIAU1976(ra0, dec0, JD0, JD float64) (float64, float64) {
	// 定义常量
	ra0 *= rad
	dec0 *= rad
	t := (JD - JD0) / 36525.0
	T := (JD0 - 2451545.0) / 36525.0

	// Explanatory Supplement 2nd ed. table 3.211.1
	zeta := ((2306.2181+1.39656*T-0.000139*T*T)*t +
		(0.30188-0.000344*T)*t*t + 0.017998*t*t*t) / 3600.0 * rad

	z := ((2306.2181+1.39656*T-0.000139*T*T)*t +
		(1.09468+0.000066*T)*t*t + 0.018203*t*t*t) / 3600.0 * rad

	theta := ((2004.3109-0.85330*T-0.000217*T*T)*t -
		(0.42665+0.000217*T)*t*t - 0.041833*t*t*t) / 3600.0 * rad

	A := math.Cos(dec0) * math.Sin(ra0+z)
	B := math.Cos(theta)*math.Cos(dec0)*math.Cos(ra0+z) - math.Sin(theta)*math.Sin(dec0)
	C := math.Sin(theta)*math.Cos(dec0)*math.Cos(ra0+z) + math.Cos(theta)*math.Sin(dec0)

	raMinZ := math.Atan2(A, B)
	dec := math.Asin(C)
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

// 计算长期岁差的黄极向量
func ltpPECL(epj float64) Vector3 {
	// 多项式系数
	pqPol := [2][4]float64{
		{+5851.607687, -0.1189000, -0.00028913, +0.000000101},
		{-1600.886300, +1.1689818, -0.00000020, -0.000000437},
	}

	// 周期项系数
	pqPer := [8][5]float64{
		{708.15, -5486.751211, -684.661560, 667.666730, -5523.863691},
		{2309.00, -17.127623, 2446.283880, -2354.886252, -549.747450},
		{1620.00, -617.517403, 399.671049, -428.152441, -310.998056},
		{492.20, 413.442940, -356.652376, 376.202861, 421.535876},
		{1183.00, 78.614193, -186.387003, 184.778874, -36.776172},
		{622.00, -180.732815, -316.800070, 335.321713, -145.278396},
		{882.00, -87.676083, 198.296071, -185.138669, -34.744450},
		{547.00, 46.140315, 101.135679, -120.972830, 22.885731},
	}

	// 计算从J2000.0起算的儒略世纪数
	t := (epj - 2000.0) / 100.0

	// 初始化P和Q累加器
	p, q := 0.0, 0.0

	// 周期项
	for i := 0; i < 8; i++ {
		w := D2PI * t
		a := w / pqPer[i][0]
		s := math.Sin(a)
		c := math.Cos(a)
		p += c*pqPer[i][1] + s*pqPer[i][3]
		q += c*pqPer[i][2] + s*pqPer[i][4]
	}

	// 多项式项
	w := 1.0
	for i := 0; i < 4; i++ {
		p += pqPol[0][i] * w
		q += pqPol[1][i] * w
		w *= t
	}

	// 转换为弧度
	p *= AS2R
	q *= AS2R

	// 形成黄极向量
	z := math.Sqrt(math.Max(1.0-p*p-q*q, 0.0))
	s := math.Sin(EPS0)
	c := math.Cos(EPS0)

	return Vector3{p, -q*c - z*s, -q*s + z*c}
}

// 计算长期岁差的赤道极向量
func ltpPEQU(epj float64) Vector3 {
	// 多项式系数
	xyPol := [2][4]float64{
		{+5453.282155, +0.4252841, -0.00037173, -0.000000152},
		{-73750.930350, -0.7675452, -0.00018725, +0.000000231},
	}

	// 周期项系数
	xyPer := [14][5]float64{
		{256.75, -819.940624, 75004.344875, 81491.287984, 1558.515853},
		{708.15, -8444.676815, 624.033993, 787.163481, 7774.939698},
		{274.20, 2600.009459, 1251.136893, 1251.296102, -2219.534038},
		{241.45, 2755.175630, -1102.212834, -1257.950837, -2523.969396},
		{2309.00, -167.659835, -2660.664980, -2966.799730, 247.850422},
		{492.20, 871.855056, 699.291817, 639.744522, -846.485643},
		{396.10, 44.769698, 153.167220, 131.600209, -1393.124055},
		{288.90, -512.313065, -950.865637, -445.040117, 368.526116},
		{231.10, -819.415595, 499.754645, 584.522874, 749.045012},
		{1610.00, -538.071099, -145.188210, -89.756563, 444.704518},
		{620.00, -189.793622, 558.116553, 524.429630, 235.934465},
		{157.87, -402.922932, -23.923029, -13.549067, 374.049623},
		{220.30, 179.516345, -165.405086, -210.157124, -171.330180},
		{1200.00, -9.814756, 9.344131, -44.919798, -22.899655},
	}

	// 计算从J2000.0起算的儒略世纪数
	t := (epj - 2000.0) / 100.0

	// 初始化X和Y累加器
	x, y := 0.0, 0.0

	// 周期项
	for i := 0; i < 14; i++ {
		w := D2PI * t
		a := w / xyPer[i][0]
		s := math.Sin(a)
		c := math.Cos(a)
		x += c*xyPer[i][1] + s*xyPer[i][3]
		y += c*xyPer[i][2] + s*xyPer[i][4]
	}

	// 多项式项
	w := 1.0
	for i := 0; i < 4; i++ {
		x += xyPol[0][i] * w
		y += xyPol[1][i] * w
		w *= t
	}

	// 转换为弧度
	x *= AS2R
	y *= AS2R

	// 形成赤道极向量
	w = x*x + y*y
	z := 0.0
	if w < 1.0 {
		z = math.Sqrt(1.0 - w)
	}

	return Vector3{x, y, z}
}

// 向量叉乘
func pxp(a, b Vector3) Vector3 {
	return Vector3{
		a[1]*b[2] - a[2]*b[1],
		a[2]*b[0] - a[0]*b[2],
		a[0]*b[1] - a[1]*b[0],
	}
}

// 向量归一化
func pn(p Vector3) (Vector3, float64) {
	w := math.Sqrt(p[0]*p[0] + p[1]*p[1] + p[2]*p[2])
	if w == 0.0 {
		return Vector3{0, 0, 0}, 0
	}
	return Vector3{p[0] / w, p[1] / w, p[2] / w}, w
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
