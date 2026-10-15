package formula

import "math"

const arcsecToRadian = math.Pi / 180 / 3600

// TIOLocatorArcsec 地球自转原点定位角 s′，单位角秒 / TIO locator s′ in arcseconds.
//
// jdeTT 为 TT 儒略日。采用 IERS 2010 的线性近似 s′ = −47 µas·t，t 为 J2000 起的儒略世纪数。
// jdeTT is a TT Julian date. Uses the IERS 2010 linear model s′ = −47 µas·t,
// with t in Julian centuries since J2000.
func TIOLocatorArcsec(jdeTT float64) float64 {
	if !isFinitePhotometry(jdeTT) {
		return math.NaN()
	}
	return -47e-6 * (jdeTT - 2451545.0) / 36525
}

// PolarMotionMatrix 极移矩阵 W，把地球中间参考系 (TIRS) 矢量转到 ITRS / polar-motion matrix from TIRS to ITRS.
//
// xp、yp 为 IERS 公布的极坐标，sPrime 为 TIOLocatorArcsec 的结果，单位都是角秒。
// 返回 W = R1(−yp)·R2(−xp)·R3(s′)，按 v_ITRS = W·v_TIRS 使用。
// xp and yp are IERS pole coordinates and sPrime is the TIO locator, all in
// arcseconds. Returns W = R1(−yp)·R2(−xp)·R3(s′) for v_ITRS = W·v_TIRS.
func PolarMotionMatrix(xp, yp, sPrime float64) [3][3]float64 {
	if !isFinitePhotometry(xp) || !isFinitePhotometry(yp) || !isFinitePhotometry(sPrime) {
		return nanMatrix()
	}
	return multiplyMatrix(rotationX(-yp*arcsecToRadian),
		multiplyMatrix(rotationY(-xp*arcsecToRadian), rotationZ(sPrime*arcsecToRadian)))
}

// CelestialPoleOffsetMatrix 天球极偏差修正矩阵 / celestial pole offset correction matrix.
//
// dX、dY 为 IERS 公布的 CIP 坐标改正，单位毫角秒。返回一阶小角矩阵 Δ，
// 左乘模型给出的 GCRS→CIRS（或 GCRS→真赤道）矩阵即得改正后的矩阵：C′ = Δ·C。
// dX and dY are IERS CIP corrections in milliarcseconds. Returns the
// first-order matrix Δ; left-multiply a model GCRS→CIRS (or GCRS→true-equator)
// matrix to apply the offsets: C′ = Δ·C.
func CelestialPoleOffsetMatrix(dX, dY float64) [3][3]float64 {
	if !isFinitePhotometry(dX) || !isFinitePhotometry(dY) {
		return nanMatrix()
	}
	x := dX * arcsecToRadian / 1000
	y := dY * arcsecToRadian / 1000
	return [3][3]float64{
		{1, 0, -x},
		{0, 1, -y},
		{x, y, 1},
	}
}

// rotationX 等为被动旋转：把矢量所在坐标轴绕 x/y/z 轴转 angle 弧度。
func rotationX(angle float64) [3][3]float64 {
	s, c := math.Sincos(angle)
	return [3][3]float64{{1, 0, 0}, {0, c, s}, {0, -s, c}}
}

func rotationY(angle float64) [3][3]float64 {
	s, c := math.Sincos(angle)
	return [3][3]float64{{c, 0, -s}, {0, 1, 0}, {s, 0, c}}
}

func rotationZ(angle float64) [3][3]float64 {
	s, c := math.Sincos(angle)
	return [3][3]float64{{c, s, 0}, {-s, c, 0}, {0, 0, 1}}
}

func multiplyMatrix(a, b [3][3]float64) [3][3]float64 {
	var out [3][3]float64
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			out[i][j] = a[i][0]*b[0][j] + a[i][1]*b[1][j] + a[i][2]*b[2][j]
		}
	}
	return out
}

func nanMatrix() [3][3]float64 {
	n := math.NaN()
	return [3][3]float64{{n, n, n}, {n, n, n}, {n, n, n}}
}
