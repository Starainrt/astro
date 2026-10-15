package eclipse

import (
	"github.com/starainrt/astro/basic"
)

// SolarEclipseBesselianPolynomial 是三次多项式系数，索引 n 对应 t 的 n 次幂，t 为自 T0 起算的 TT 小时数。
// SolarEclipseBesselianPolynomial holds the cubic coefficients; index n multiplies t^n with t in TT hours from T0.
type SolarEclipseBesselianPolynomial = basic.SolarEclipseBesselianPolynomial

// SolarEclipseBesselianElementsOptions 是贝塞尔根数表的生成选项 / options for a Besselian element table.
type SolarEclipseBesselianElementsOptions struct {
	// Model 月亮半径模型，零值为 NASA bulletin Split-K / lunar radius model.
	Model SolarEclipseRadiusModel
	// SunRadiusModel 太阳半径口径，零值为标准档 / solar radius convention.
	SunRadiusModel SolarEclipseSunRadiusModel
	// DeltaTSeconds 显式 ΔT（秒），<=0 用进程级模型，只改变地球自转相位 / explicit ΔT in seconds.
	DeltaTSeconds float64
	// ReferenceJDE 多项式参考时刻 T0（TT 儒略日），<=0 取离食甚最近的整 TT 小时 / TT reference instant, the whole hour nearest greatest eclipse.
	ReferenceJDE float64
	// ValidHours 有效窗口半径（小时），<=0 取 3 / validity window half-width in hours.
	ValidHours float64
}

// SolarEclipseBesselianElementsResult 是一次日食的多项式贝塞尔根数及其口径 / polynomial Besselian elements and their conventions.
type SolarEclipseBesselianElementsResult struct {
	// T0JDE 多项式参考时刻（TT 儒略日），t = (jde - T0JDE) * 24 / TT reference instant.
	T0JDE float64
	// ValidHours 有效窗口半径（小时） / validity window half-width.
	ValidHours float64
	// X 与 Y 是月心在基本面内的坐标，单位地球赤道半径 / fundamental-plane coordinates in equatorial Earth radii.
	X, Y SolarEclipseBesselianPolynomial
	// D 是影轴赤纬，单位度 / declination of the shadow axis in degrees.
	D SolarEclipseBesselianPolynomial
	// L1 与 L2 是基本面内的半影、本影半径，单位地球赤道半径；本影为负表示月心尚未越过本影锥顶点。
	// L1 and L2 are the penumbral and umbral radii in the fundamental plane, in equatorial Earth radii.
	L1, L2 SolarEclipseBesselianPolynomial
	// Mu 是影轴格林时角，单位度，窗口内连续、不折回 [0,360)。
	//
	// 口径与已发布根数表不同：本库的恒星时取自 UT = TT - ΔT，得到真实格林时角；已发布表改用 T0 本身
	// 的恒星时，两者相差 ΔT × 15.041067/3600 度。要对表先用 SolarEclipseBesselianMuForPublishedTable。
	// Mu is the shadow axis' Greenwich hour angle in degrees, continuous and not folded into [0,360).
	Mu SolarEclipseBesselianPolynomial
	// TanF1 与 TanF2 是半影、本影锥半顶角正切，本次日食内为常数 / cone half-angle tangents, constant over the eclipse.
	TanF1, TanF2 float64
	// Gamma 是食甚时刻影轴到地心的距离，单位地球赤道半径 / shadow-axis distance from the Earth's centre.
	Gamma float64
	// Magnitude 是食甚时刻的全局食分 / global eclipse magnitude at greatest eclipse.
	Magnitude float64
	// Model、SunRadiusModel、PenumbralK、UmbralK 与 DeltaTSeconds 是决定上述数值的口径 / the conventions behind the numbers above.
	Model          SolarEclipseRadiusModel
	SunRadiusModel SolarEclipseSunRadiusModel
	PenumbralK     float64
	UmbralK        float64
	DeltaTSeconds  float64
}

// SolarEclipseBesselianElements 计算给定近朔时刻附近一次日食的多项式贝塞尔根数，窗口内无日食时返回 false / polynomial Besselian elements near a new moon; false when no eclipse.
func (module *Eclipse) SolarEclipseBesselianElements(
	seedJDE float64, options SolarEclipseBesselianElementsOptions,
) (SolarEclipseBesselianElementsResult, bool) {
	result, ok := basic.SolarEclipseBesselianElementsWithInstance(seedJDE, basic.SolarEclipseBesselianElementsOptions{
		Model:          basic.SolarEclipseRadiusModel(options.Model),
		SunRadiusModel: basic.SolarEclipseSunRadiusModel(options.SunRadiusModel),
		DeltaTSeconds:  options.DeltaTSeconds,
		ReferenceJDE:   options.ReferenceJDE,
		ValidHours:     options.ValidHours,
	}, module.inst)
	if !ok {
		return SolarEclipseBesselianElementsResult{}, false
	}
	return solarEclipseBesselianElementsFromBasic(result), true
}

// SolarEclipseBesselianMuForPublishedTable 把本库的 Mu 换算成与已发布根数表直接可比的取值。
// 已发布表用 T0 本身的恒星时，本库用 UT = TT - ΔT，两者只差一个常数，因此只有常数项平移。
// SolarEclipseBesselianMuForPublishedTable shifts Mu onto the argument used by published element tables.
func SolarEclipseBesselianMuForPublishedTable(
	mu SolarEclipseBesselianPolynomial, deltaTSeconds float64,
) SolarEclipseBesselianPolynomial {
	return basic.SolarEclipseBesselianMuForPublishedTable(mu, deltaTSeconds)
}

func solarEclipseBesselianElementsFromBasic(
	result basic.SolarEclipseBesselianElementsResult,
) SolarEclipseBesselianElementsResult {
	return SolarEclipseBesselianElementsResult{
		T0JDE:          result.T0JDE,
		ValidHours:     result.ValidHours,
		X:              result.X,
		Y:              result.Y,
		D:              result.D,
		L1:             result.L1,
		L2:             result.L2,
		Mu:             result.Mu,
		TanF1:          result.TanF1,
		TanF2:          result.TanF2,
		Gamma:          result.Gamma,
		Magnitude:      result.Magnitude,
		Model:          mapBasicSolarEclipseModel(result.Model),
		SunRadiusModel: result.SunRadiusModel,
		PenumbralK:     result.PenumbralK,
		UmbralK:        result.UmbralK,
		DeltaTSeconds:  result.DeltaTSeconds,
	}
}
