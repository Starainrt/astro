package basic

import (
	"math"
)

/*
黄赤交角、nutation==true时，计算交角章动
*/
func EclipticObliquity(jde float64, nutation bool) float64 {
	eps := Obliquity1980(jde)
	if nutation {
		eps += Nutation2000Bs(jde)
	}
	return eps
}

func Sita(JD float64) float64 {
	return EclipticObliquity(JD, true)
}

// 黄经章动 1980
func Nutation1980i(jd float64) float64 { // '黄经章动
	res, _ := Nutation1980(jd)
	return res
}

// 交角章动1980
func Nutation1980s(jd float64) float64 { //交角章动
	_, res := Nutation1980(jd)
	return res
}

// 黄经章动 2000B
func Nutation2000Bi(jd float64) float64 { // '黄经章动
	res, _ := Nutation2000B(jd)
	return res
}

// 交角章动2000B
func Nutation2000Bs(jd float64) float64 { //交角章动
	_, res := Nutation2000B(jd)
	return res
}

// 定义 IAU 1980 章动系数
var coefficientsNut1980 = [][9]float64{
	{0, 0, 0, 0, 1, -171996.0, -174.2, 92025.0, 8.9},
	{0, 0, 0, 0, 2, 2062.0, 0.2, -895.0, 0.5},
	{-2, 0, 2, 0, 1, 46.0, 0.0, -24.0, 0.0},
	{2, 0, -2, 0, 0, 11.0, 0.0, 0.0, 0.0},
	{-2, 0, 2, 0, 2, -3.0, 0.0, 1.0, 0.0},
	{1, -1, 0, -1, 0, -3.0, 0.0, 0.0, 0.0},
	{0, -2, 2, -2, 1, -2.0, 0.0, 1.0, 0.0},
	{2, 0, -2, 0, 1, 1.0, 0.0, 0.0, 0.0},
	{0, 0, 2, -2, 2, -13187.0, -1.6, 5736.0, -3.1},
	{0, 1, 0, 0, 0, 1426.0, -3.4, 54.0, -0.1},
	{0, 1, 2, -2, 2, -517.0, 1.2, 224.0, -0.6},
	{0, -1, 2, -2, 2, 217.0, -0.5, -95.0, 0.3},
	{0, 0, 2, -2, 1, 129.0, 0.1, -70.0, 0.0},
	{2, 0, 0, -2, 0, 48.0, 0.0, 1.0, 0.0},
	{0, 0, 2, -2, 0, -22.0, 0.0, 0.0, 0.0},
	{0, 2, 0, 0, 0, 17.0, -0.1, 0.0, 0.0},
	{0, 1, 0, 0, 1, -15.0, 0.0, 9.0, 0.0},
	{0, 2, 2, -2, 2, -16.0, 0.1, 7.0, 0.0},
	{0, -1, 0, 0, 1, -12.0, 0.0, 6.0, 0.0},
	{-2, 0, 0, 2, 1, -6.0, 0.0, 3.0, 0.0},
	{0, -1, 2, -2, 1, -5.0, 0.0, 3.0, 0.0},
	{2, 0, 0, -2, 1, 4.0, 0.0, -2.0, 0.0},
	{0, 1, 2, -2, 1, 4.0, 0.0, -2.0, 0.0},
	{1, 0, 0, -1, 0, -4.0, 0.0, 0.0, 0.0},
	{2, 1, 0, -2, 0, 1.0, 0.0, 0.0, 0.0},
	{0, 0, -2, 2, 1, 1.0, 0.0, 0.0, 0.0},
	{0, 1, -2, 2, 0, -1.0, 0.0, 0.0, 0.0},
	{0, 1, 0, 0, 2, 1.0, 0.0, 0.0, 0.0},
	{-1, 0, 0, 1, 1, 1.0, 0.0, 0.0, 0.0},
	{0, 1, 2, -2, 0, -1.0, 0.0, 0.0, 0.0},
	{0, 0, 2, 0, 2, -2274.0, -0.2, 977.0, -0.5},
	{1, 0, 0, 0, 0, 712.0, 0.1, -7.0, 0.0},
	{0, 0, 2, 0, 1, -386.0, -0.4, 200.0, 0.0},
	{1, 0, 2, 0, 2, -301.0, 0.0, 129.0, -0.1},
	{1, 0, 0, -2, 0, -158.0, 0.0, -1.0, 0.0},
	{-1, 0, 2, 0, 2, 123.0, 0.0, -53.0, 0.0},
	{0, 0, 0, 2, 0, 63.0, 0.0, -2.0, 0.0},
	{1, 0, 0, 0, 1, 63.0, 0.1, -33.0, 0.0},
	{-1, 0, 0, 0, 1, -58.0, -0.1, 32.0, 0.0},
	{-1, 0, 2, 2, 2, -59.0, 0.0, 26.0, 0.0},
	{1, 0, 2, 0, 1, -51.0, 0.0, 27.0, 0.0},
	{0, 0, 2, 2, 2, -38.0, 0.0, 16.0, 0.0},
	{2, 0, 0, 0, 0, 29.0, 0.0, -1.0, 0.0},
	{1, 0, 2, -2, 2, 29.0, 0.0, -12.0, 0.0},
	{2, 0, 2, 0, 2, -31.0, 0.0, 13.0, 0.0},
	{0, 0, 2, 0, 0, 26.0, 0.0, -1.0, 0.0},
	{-1, 0, 2, 0, 1, 21.0, 0.0, -10.0, 0.0},
	{-1, 0, 0, 2, 1, 16.0, 0.0, -8.0, 0.0},
	{1, 0, 0, -2, 1, -13.0, 0.0, 7.0, 0.0},
	{-1, 0, 2, 2, 1, -10.0, 0.0, 5.0, 0.0},
	{1, 1, 0, -2, 0, -7.0, 0.0, 0.0, 0.0},
	{0, 1, 2, 0, 2, 7.0, 0.0, -3.0, 0.0},
	{0, -1, 2, 0, 2, -7.0, 0.0, 3.0, 0.0},
	{1, 0, 2, 2, 2, -8.0, 0.0, 3.0, 0.0},
	{1, 0, 0, 2, 0, 6.0, 0.0, 0.0, 0.0},
	{2, 0, 2, -2, 2, 6.0, 0.0, -3.0, 0.0},
	{0, 0, 0, 2, 1, -6.0, 0.0, 3.0, 0.0},
	{0, 0, 2, 2, 1, -7.0, 0.0, 3.0, 0.0},
	{1, 0, 2, -2, 1, 6.0, 0.0, -3.0, 0.0},
	{0, 0, 0, -2, 1, -5.0, 0.0, 3.0, 0.0},
	{1, -1, 0, 0, 0, 5.0, 0.0, 0.0, 0.0},
	{2, 0, 2, 0, 1, -5.0, 0.0, 3.0, 0.0},
	{0, 1, 0, -2, 0, -4.0, 0.0, 0.0, 0.0},
	{1, 0, -2, 0, 0, 4.0, 0.0, 0.0, 0.0},
	{0, 0, 0, 1, 0, -4.0, 0.0, 0.0, 0.0},
	{1, 1, 0, 0, 0, -3.0, 0.0, 0.0, 0.0},
	{1, 0, 2, 0, 0, 3.0, 0.0, 0.0, 0.0},
	{1, -1, 2, 0, 2, -3.0, 0.0, 1.0, 0.0},
	{-1, -1, 2, 2, 2, -3.0, 0.0, 1.0, 0.0},
	{-2, 0, 0, 0, 1, -2.0, 0.0, 1.0, 0.0},
	{3, 0, 2, 0, 2, -3.0, 0.0, 1.0, 0.0},
	{0, -1, 2, 2, 2, -3.0, 0.0, 1.0, 0.0},
	{1, 1, 2, 0, 2, 2.0, 0.0, -1.0, 0.0},
	{-1, 0, 2, -2, 1, -2.0, 0.0, 1.0, 0.0},
	{2, 0, 0, 0, 1, 2.0, 0.0, -1.0, 0.0},
	{1, 0, 0, 0, 2, -2.0, 0.0, 1.0, 0.0},
	{3, 0, 0, 0, 0, 2.0, 0.0, 0.0, 0.0},
	{0, 0, 2, 1, 2, 2.0, 0.0, -1.0, 0.0},
	{-1, 0, 0, 0, 2, 1.0, 0.0, -1.0, 0.0},
	{1, 0, 0, -4, 0, -1.0, 0.0, 0.0, 0.0},
	{-2, 0, 2, 2, 2, 1.0, 0.0, -1.0, 0.0},
	{-1, 0, 2, 4, 2, -2.0, 0.0, 1.0, 0.0},
	{2, 0, 0, -4, 0, -1.0, 0.0, 0.0, 0.0},
	{1, 1, 2, -2, 2, 1.0, 0.0, -1.0, 0.0},
	{1, 0, 2, 2, 1, -1.0, 0.0, 1.0, 0.0},
	{-2, 0, 2, 4, 2, -1.0, 0.0, 1.0, 0.0},
	{-1, 0, 4, 0, 2, 1.0, 0.0, 0.0, 0.0},
	{1, -1, 0, -2, 0, 1.0, 0.0, 0.0, 0.0},
	{2, 0, 2, -2, 1, 1.0, 0.0, -1.0, 0.0},
	{2, 0, 2, 2, 2, -1.0, 0.0, 0.0, 0.0},
	{1, 0, 0, 2, 1, -1.0, 0.0, 0.0, 0.0},
	{0, 0, 4, -2, 2, 1.0, 0.0, 0.0, 0.0},
	{3, 0, 2, -2, 2, 1.0, 0.0, 0.0, 0.0},
	{1, 0, 2, -2, 0, -1.0, 0.0, 0.0, 0.0},
	{0, 1, 2, 0, 1, 1.0, 0.0, 0.0, 0.0},
	{-1, -1, 0, 2, 1, 1.0, 0.0, 0.0, 0.0},
	{0, 0, -2, 0, 1, -1.0, 0.0, 0.0, 0.0},
	{0, 0, 2, -1, 2, -1.0, 0.0, 0.0, 0.0},
	{0, 1, 0, 2, 0, -1.0, 0.0, 0.0, 0.0},
	{1, 0, -2, -2, 0, -1.0, 0.0, 0.0, 0.0},
	{0, -1, 2, 0, 1, -1.0, 0.0, 0.0, 0.0},
	{1, 1, 0, -2, 1, -1.0, 0.0, 0.0, 0.0},
	{1, 0, -2, 2, 0, -1.0, 0.0, 0.0, 0.0},
	{2, 0, 0, 2, 0, 1.0, 0.0, 0.0, 0.0},
	{0, 0, 2, 4, 2, -1.0, 0.0, 0.0, 0.0},
	{0, 1, 0, 1, 0, 1.0, 0.0, 0.0, 0.0},
}

// fmod 函数实现浮点数取模
func fmod(f, m float64) float64 {
	return f - m*math.Floor(f/m)
}

// frac 函数获取小数部分
func frac(f float64) float64 {
	return f - math.Floor(f)
}

// Obliquity1980 计算黄赤交角，单位为度
// 公式 3.222-1
func Obliquity1980(jd float64) float64 {
	T := (jd - 2451545.0) / 36525.0
	as2r := ((1.0 / 3600.0) * math.Pi) / 180.0
	eps := 84381.448 - 46.8150*T - 0.00059*T*T + 0.001813*T*T*T // 84381.448 = 23d26'21.448 转换为角秒

	return math.Mod(eps*as2r, 2*math.Pi) * deg
}

// Nutation1980 计算 IAU 1980 章动模型
// 返回黄经章动 (dpsi) 和交角章动 (deps)，单位为度
func Nutation1980(jd float64) (float64, float64) {
	t := (jd - 2451545.0) / 36525.0

	twoPI := 2 * math.Pi
	as2r := ((1.0 / 3600.0) * math.Pi) / 180.0

	// 表 3.222.2 - 计算章动参数
	l := fmod((485866.733+(715922.633+(31.310+0.064*t)*t)*t)*as2r+frac(1325.0*t)*twoPI, twoPI)
	lp := fmod((1287099.804+(1292581.224+(-0.577-0.012*t)*t)*t)*as2r+frac(99.0*t)*twoPI, twoPI)
	F := fmod((335778.877+(295263.137+(-13.257+0.011*t)*t)*t)*as2r+frac(1342.0*t)*twoPI, twoPI)
	D := fmod((1072261.307+(1105601.328+(-6.891+0.019*t)*t)*t)*as2r+frac(1236.0*t)*twoPI, twoPI)
	O := fmod((450160.280+(-482890.539+(7.455+0.008*t)*t)*t)*as2r+frac(-5.0*t)*twoPI, twoPI)

	deps := 0.0
	dpsi := 0.0

	// 公式 3.222-6 - 计算章动
	for i := len(coefficientsNut1980) - 1; i >= 0; i-- {
		sumargs := coefficientsNut1980[i][0]*l +
			coefficientsNut1980[i][1]*lp +
			coefficientsNut1980[i][2]*F +
			coefficientsNut1980[i][3]*D +
			coefficientsNut1980[i][4]*O

		deps += math.Cos(sumargs) * (coefficientsNut1980[i][7] + coefficientsNut1980[i][8]*t)
		dpsi += math.Sin(sumargs) * (coefficientsNut1980[i][5] + coefficientsNut1980[i][6]*t)
	}

	deps = (deps * as2r) / 10000.0
	dpsi = (dpsi * as2r) / 10000.0

	return dpsi * deg, deps * deg
}

// Nutation2000B 计算 IAU 2000B 章动模型
// 返回交角章动 (de) 和黄经章动 (dp)，单位为度
func Nutation2000B(jd float64) (float64, float64) {
	// 常量定义
	as2r := 4.848136811095359935899141e-6 // 角秒到弧度的转换因子
	twopi := 6.283185307179586476925287   // 2π

	T := (jd - 2451545) / 36525
	L := math.Mod(485868.249036+1717915923.2178*T, 1296000.0) * as2r
	Lp := math.Mod(1287104.79305+129596581.0481*T, 1296000.0) * as2r
	F := math.Mod(335779.526232+1739527262.8478*T, 1296000.0) * as2r
	D := math.Mod(1072260.70369+1602961601.2090*T, 1296000.0) * as2r
	Om := math.Mod(450160.398036-6962890.5431*T, 1296000.0) * as2r

	dp := 0.0
	de := 0.0
	var arg, sinarg, cosarg float64

	// 以下是 IAU 2000B 章动模型的各项计算
	// 每行对应一个章动项
	arg = math.Mod(L+Lp+2*F+-2*D+2*Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += 1290 * sinarg
	de += -556 * cosarg

	arg = math.Mod(-1*L+2*Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += 1405*sinarg + 4*cosarg
	de += -610*cosarg + 2*sinarg

	arg = math.Mod(-2*L+2*F+2*D+2*Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += 1383*sinarg + -2*cosarg
	de += -594*cosarg + -2*sinarg

	arg = math.Mod(L+2*F+2*D+Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += -1331*sinarg + 8*cosarg
	de += 663*cosarg + 4*sinarg

	arg = math.Mod(-2*Lp+2*F+-2*D+Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += -1283 * sinarg
	de += 672 * cosarg

	arg = math.Mod(-1*L+Lp+D+Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += 1314 * sinarg
	de += -700 * cosarg

	arg = math.Mod(-1*L+2*F+4*D+2*Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += -1521*sinarg + 9*cosarg
	de += 647*cosarg + 4*sinarg

	arg = math.Mod(2*F+D+2*Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += 1660*sinarg + -5*cosarg
	de += -710*cosarg + -2*sinarg

	arg = math.Mod(-1*L+D, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += 4026*sinarg + -353*cosarg
	de += -553*cosarg + -139*sinarg

	arg = math.Mod(L+2*Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += -1981 * sinarg
	de += 854 * cosarg

	arg = math.Mod(-1*L+2*F+-2*D+Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += -1987*sinarg + -6*cosarg
	de += 1073*cosarg + -2*sinarg

	arg = math.Mod(L+2*F, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += 3339*sinarg + -13*cosarg
	de += -107*cosarg + 1*sinarg

	arg = math.Mod(L+Lp, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += -3389*sinarg + 5*cosarg
	de += 35*cosarg + -2*sinarg

	arg = math.Mod(-1*L+Lp+D, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += 3276*sinarg + 1*cosarg
	de += -9 * cosarg

	arg = math.Mod(2*L+Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += 2179*sinarg + -2*cosarg
	de += -1129*cosarg + -2*sinarg

	arg = math.Mod(L+Lp+2*F+2*Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += 2481*sinarg + -7*cosarg
	de += -1062*cosarg + -3*sinarg

	arg = math.Mod(-2*L+Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += -2294*sinarg + -10*cosarg
	de += 1266*cosarg + -4*sinarg

	arg = math.Mod(-1*Lp+2*F+2*D+2*Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += -2647*sinarg + 11*cosarg
	de += 1129*cosarg + 5*sinarg

	arg = math.Mod(-1*L+2*F, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += -4056*sinarg + 5*cosarg
	de += 40*cosarg + -2*sinarg

	arg = math.Mod(-1*L+-1*Lp+2*F+2*D+2*Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += -2819*sinarg + 7*cosarg
	de += 1207*cosarg + 3*sinarg

	arg = math.Mod(D, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += -4230*sinarg + 5*cosarg
	de += -20*cosarg + -2*sinarg

	arg = math.Mod(L+-1*Lp+2*F+2*Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += -2878*sinarg + 8*cosarg
	de += 1232*cosarg + 4*sinarg

	arg = math.Mod(-1*Lp+2*D, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += 4348*sinarg + -10*cosarg
	de += -81*cosarg + 2*sinarg

	arg = math.Mod(3*L+2*F+2*Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += -2904*sinarg + 15*cosarg
	de += 1233*cosarg + 7*sinarg

	arg = math.Mod(-2*L+2*F+2*Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += -3075*sinarg + -2*cosarg
	de += 1313*cosarg + -1*sinarg

	arg = math.Mod(L+-1*Lp, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += 4725*sinarg + -6*cosarg
	de += -41*cosarg + 3*sinarg

	arg = math.Mod(Lp+2*F+-2*D+Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += 3579*sinarg + 5*cosarg
	de += -1900*cosarg + 1*sinarg

	arg = math.Mod(L+2*D, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += 6579*sinarg + -24*cosarg
	de += -199*cosarg + 2*sinarg

	arg = math.Mod(2*L+-2*D+Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += 4065*sinarg + 6*cosarg
	de += -2206*cosarg + 1*sinarg

	arg = math.Mod(-1*L+-1*Lp+2*D, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += 7350*sinarg + -8*cosarg
	de += -51*cosarg + 4*sinarg

	arg = math.Mod(-2*D+Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += (-4940+-11*T)*sinarg + -21*cosarg
	de += 2720*cosarg + -9*sinarg

	arg = math.Mod(-1*Lp+2*F+-2*D+Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += (-4752+-11*T)*sinarg + -3*cosarg
	de += 2719*cosarg + -3*sinarg

	arg = math.Mod(2*L+2*F+Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += -5350*sinarg + 21*cosarg
	de += 2695*cosarg + 12*sinarg

	arg = math.Mod(-2*L+2*D+Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += (-5774+-11*T)*sinarg + -15*cosarg
	de += 3041*cosarg + -5*sinarg

	arg = math.Mod(2*L+2*F+-2*D+2*Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += 6443*sinarg + -7*cosarg
	de += -2768*cosarg + -4*sinarg

	arg = math.Mod(L+2*F+-2*D+Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += (5800+10*T)*sinarg + 2*cosarg
	de += -3045*cosarg + -1*sinarg

	arg = math.Mod(2*D+Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += (-6302+-11*T)*sinarg + 2*cosarg
	de += 3272*cosarg + 4*sinarg

	arg = math.Mod(-1*Lp+2*F+2*Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += (-7141+21*T)*sinarg + 8*cosarg
	de += 3070*cosarg + 4*sinarg

	arg = math.Mod(2*F+2*D+Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += (-6637+-11*T)*sinarg + 25*cosarg
	de += 3353*cosarg + 14*sinarg

	arg = math.Mod(Lp+2*F+2*Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += (7566+-21*T)*sinarg + -11*cosarg
	de += -3250*cosarg + -5*sinarg

	arg = math.Mod(-2*L+2*F, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += -11024*sinarg + -14*cosarg
	de += 104*cosarg + 2*sinarg

	arg = math.Mod(L+2*F+2*D+2*Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += -7691*sinarg + 44*cosarg
	de += 3268*cosarg + 19*sinarg

	arg = math.Mod(2*Lp, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += (16707+-85*T)*sinarg + -10*cosarg
	de += (168+-1*T)*cosarg + 10*sinarg

	arg = math.Mod(-1*L+2*F+2*D+Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += -10204*sinarg + 25*cosarg
	de += 5222*cosarg + 15*sinarg

	arg = math.Mod(-1*Lp+Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += (-12654+11*T)*sinarg + 63*cosarg
	de += 6415*cosarg + 26*sinarg

	arg = math.Mod(L+-2*D+Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += (-12873+-10*T)*sinarg + -37*cosarg
	de += 6953*cosarg + -14*sinarg

	arg = math.Mod(-2*F+2*D, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += 21783*sinarg + 13*cosarg
	de += -167*cosarg + 13*sinarg

	arg = math.Mod(2*Lp+2*F+-2*D+2*Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += (-15794+72*T)*sinarg + -16*cosarg
	de += (6850+-42*T)*cosarg + -5*sinarg

	arg = math.Mod(-1*L+2*D+Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += (15164+10*T)*sinarg + 11*cosarg
	de += -8001*cosarg + -1*sinarg

	arg = math.Mod(Lp+Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += (-14053+-25*T)*sinarg + 79*cosarg
	de += (8551+-2*T)*cosarg + -45*sinarg

	arg = math.Mod(2*F, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += 25887*sinarg + -66*cosarg
	de += -550*cosarg + 11*sinarg

	arg = math.Mod(2*L, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += 29243*sinarg + -74*cosarg
	de += -609*cosarg + 13*sinarg

	arg = math.Mod(-1*L+2*F+Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += (20441+21*T)*sinarg + 10*cosarg
	de += -10758*cosarg + -3*sinarg

	arg = math.Mod(L+2*F+-2*D+2*Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += 28593*sinarg + -1*cosarg
	de += (-12338+10*T)*cosarg + -3*sinarg

	arg = math.Mod(2*L+2*F+2*Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += (-31046+-1*T)*sinarg + 131*cosarg
	de += (13238+-11*T)*cosarg + 59*sinarg

	arg = math.Mod(-2*L+2*D, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += -47722*sinarg + -18*cosarg
	de += 477*cosarg + -25*sinarg

	arg = math.Mod(-2*Lp+2*F+-2*D+2*Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += 32481 * sinarg
	de += -13870 * cosarg

	arg = math.Mod(2*F+2*D+2*Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += (-38571+-1*T)*sinarg + 158*cosarg
	de += (16452+-11*T)*cosarg + 68*sinarg

	arg = math.Mod(2*D, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += (63384+11*T)*sinarg + -150*cosarg
	de += -1220*cosarg + 29*sinarg

	arg = math.Mod(-2*L+2*F+Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += (45893+50*T)*sinarg + 31*cosarg
	de += (-24236+-10*T)*cosarg + 20*sinarg

	arg = math.Mod(L+2*F+Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += (-51613+-42*T)*sinarg + 129*cosarg
	de += 26366*cosarg + 78*sinarg

	arg = math.Mod(-1*L+2*F+2*D+2*Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += (-59641+-11*T)*sinarg + 149*cosarg
	de += (25543+-11*T)*cosarg + 66*sinarg

	arg = math.Mod(-1*L+Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += (-57976+-63*T)*sinarg + -189*cosarg
	de += 31429*cosarg + -75*sinarg

	arg = math.Mod(L+Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += (63110+63*T)*sinarg + 27*cosarg
	de += -33228*cosarg + -9*sinarg

	arg = math.Mod(-1*L+2*D, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += (156994+10*T)*sinarg + -168*cosarg
	de += -1235*cosarg + 82*sinarg

	arg = math.Mod(-1*L+2*F+2*Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += (123457+11*T)*sinarg + 19*cosarg
	de += (-53311+32*T)*cosarg + -4*sinarg

	arg = math.Mod(2*F+-2*D+Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += (128227+137*T)*sinarg + 181*cosarg
	de += (-68982+-9*T)*cosarg + 39*sinarg

	arg = math.Mod(-1*Lp+2*F+-2*D+2*Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += (215829+-494*T)*sinarg + 111*cosarg
	de += (-95929+299*T)*cosarg + 132*sinarg

	arg = math.Mod(L+2*F+2*Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += (-301461+-36*T)*sinarg + 816*cosarg
	de += (129025+-63*T)*cosarg + 367*sinarg

	arg = math.Mod(2*F+Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += (-387298+-367*T)*sinarg + 380*cosarg
	de += (200728+18*T)*cosarg + 318*sinarg

	arg = math.Mod(L, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += (711159+73*T)*sinarg + -872*cosarg
	de += -6750*cosarg + 358*sinarg

	arg = math.Mod(Lp+2*F+-2*D+2*Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += (-516821+1226*T)*sinarg + -524*cosarg
	de += (224386+-677*T)*cosarg + -174*sinarg

	arg = math.Mod(Lp, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += (1475877+-3633*T)*sinarg + 11817*cosarg
	de += (73871+-184*T)*cosarg + -1924*sinarg

	arg = math.Mod(2*Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += (2074554+207*T)*sinarg + -698*cosarg
	de += (-897492+470*T)*cosarg + -291*sinarg

	arg = math.Mod(2*F+2*Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += (-2276413+-234*T)*sinarg + 2796*cosarg
	de += (978459+-485*T)*cosarg + 1374*sinarg

	arg = math.Mod(2*F+-2*D+2*Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += (-13170906+-1675*T)*sinarg + -13696*cosarg
	de += (5730336+-3015*T)*cosarg + -4587*sinarg

	arg = math.Mod(Om, twopi)
	sinarg = math.Sin(arg)
	cosarg = math.Cos(arg)
	dp += (-172064161+-174666*T)*sinarg + 33386*cosarg
	de += (92052331+9086*T)*cosarg + 15377*sinarg

	de *= as2r / 1e7
	dp *= as2r / 1e7

	// 行星章动修正
	dp += -0.135 * (as2r / 1e3)
	de += 0.388 * (as2r / 1e3)

	return dp * deg, de * deg
}
