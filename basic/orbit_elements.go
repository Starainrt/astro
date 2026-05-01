package basic

import "math"

const (
	orbitReferenceJD              = 2451545.0
	gaussianGravitationalConstant = 0.01720209895 // rad/day
	lightTimeDaysPerAU            = 0.0057755183
	orbitParabolicTolerance       = 1e-12
)

// OrbitElements 日心二体圆锥曲线根数，参考系为 J2000 平黄道/平春分点。
// EpochJD 与 TpJD 使用 TT/TDB 对应的儒略日。
//
// 两种输入形式：
// 1. 椭圆经典根数：A/E/I/Omega/W/M0（原有形式）
// 2. 近日点形式：Q/E/I/Omega/W/TpJD，用于高偏心、抛物或双曲轨道
//
// 线性 rates 仅作用于经典根数形式，单位均为“每天变化量”。
type OrbitElements struct {
	EpochJD float64
	A       float64 // 半长径 / semi-major axis in AU.
	E       float64 // 离心率 / eccentricity.
	I       float64 // 轨道倾角 / inclination in degrees.
	Omega   float64 // 升交点黄经 / longitude of ascending node in degrees.
	W       float64 // 近日点幅角 / argument of perihelion in degrees.
	M0      float64 // 历元平近点角 / mean anomaly at epoch in degrees.
	Q       float64 // 近日点距离 / perihelion distance in AU.
	TpJD    float64 // 近日点通过时刻 / perihelion passage TT/TDB Julian day.

	ADot     float64 // 半长径日变化 / daily rate of A in AU/day.
	EDot     float64 // 离心率日变化 / daily rate of E per day.
	IDot     float64 // 倾角日变化 / daily rate of I in deg/day.
	OmegaDot float64 // 升交点黄经日变化 / daily rate of Omega in deg/day.
	WDot     float64 // 近日点幅角日变化 / daily rate of W in deg/day.
	MDot     float64 // 平近点角日变化 / daily rate of M in deg/day.
}

func (elements OrbitElements) usesPerihelionForm() bool {
	return isFinitePositive(elements.Q) && isFinite(elements.TpJD)
}

func (elements OrbitElements) validOrientation() bool {
	angles := [...]float64{elements.I, elements.Omega, elements.W}
	for _, angle := range angles {
		if !isFinite(angle) {
			return false
		}
	}
	return true
}

func (elements OrbitElements) validEllipticClassical() bool {
	if !isFinite(elements.EpochJD) || !isFinitePositive(elements.A) {
		return false
	}
	if !isFinite(elements.E) || elements.E < 0 || elements.E >= 1 {
		return false
	}
	if !isFinite(elements.M0) {
		return false
	}
	return elements.validOrientation()
}

func (elements OrbitElements) validPerihelionForm() bool {
	if !elements.usesPerihelionForm() {
		return false
	}
	if !isFinite(elements.E) || elements.E < 0 {
		return false
	}
	return elements.validOrientation()
}

func orbitElementsAt(jd float64, elements OrbitElements) OrbitElements {
	if elements.usesPerihelionForm() || !isFinite(jd) || !isFinite(elements.EpochJD) {
		return elements
	}
	deltaDays := jd - elements.EpochJD
	updated := elements
	updated.A += updated.ADot * deltaDays
	updated.E += updated.EDot * deltaDays
	updated.I += updated.IDot * deltaDays
	updated.Omega += updated.OmegaDot * deltaDays
	updated.W += updated.WDot * deltaDays
	return updated
}

func isFinite(value float64) bool {
	return !(math.IsNaN(value) || math.IsInf(value, 0))
}

func isFinitePositive(value float64) bool {
	return isFinite(value) && value > 0
}
