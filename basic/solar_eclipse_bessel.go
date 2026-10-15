package basic

import (
	"math"
)

//下游研究需要，改成直接导出

const (
	solarEclipseBesselianDefaultValidHours = 3.0
	solarEclipseBesselianSampleCount       = 5
	// 恒星时每秒的角度增量，用于把 ΔT 换算成影轴时角的平移量。
	solarEclipseBesselianSiderealDegreesPerSecond = 15.041067 / 3600
)

// SolarEclipseBesselianPolynomial 是三次多项式系数，索引 n 对应 t 的 n 次幂，t 为自 T0 起算的 TT 小时数。
// SolarEclipseBesselianPolynomial holds the cubic coefficients; index n multiplies t^n with t in TT hours from T0.
type SolarEclipseBesselianPolynomial [4]float64

// At 在自 T0 起 t 小时处求值 / evaluates the polynomial at t TT hours from T0.
func (polynomial SolarEclipseBesselianPolynomial) At(hours float64) float64 {
	return polynomial[0] + hours*(polynomial[1]+hours*(polynomial[2]+hours*polynomial[3]))
}

// SolarEclipseBesselianElementsOptions 是贝塞尔根数表的生成选项 / options for a Besselian element table.
type SolarEclipseBesselianElementsOptions struct {
	// Model 月亮半径模型；只有显式取 IAU Single-K 才切换，其余取值一律按 NASA bulletin Split-K。
	// Model is the lunar radius model; only an explicit IAU Single-K switches it.
	Model SolarEclipseRadiusModel
	// SunRadiusModel 太阳半径口径；零值为标准档。
	// SunRadiusModel is the solar radius convention; the zero value is the standard one.
	SunRadiusModel SolarEclipseSunRadiusModel
	// DeltaTSeconds 显式 ΔT（秒），非正值用进程级模型；它只改变地球自转相位，不改变任何 TT 时刻。
	// DeltaTSeconds is an explicit ΔT in seconds, non-positive uses the process model; it only sets Earth rotation.
	DeltaTSeconds float64
	// ReferenceJDE 多项式参考时刻 T0（TT 儒略日），非正值取食甚最近的整 TT 小时（四舍五入），与已发布根数表一致。
	// ReferenceJDE is the TT reference instant T0; non-positive uses the whole TT hour nearest to greatest eclipse.
	ReferenceJDE float64
	// ValidHours 多项式有效窗口半径（小时），非正值取 3；窗口内取 5 个等距时刻做三次最小二乘。
	// ValidHours is the half-width of the validity window in hours, non-positive uses 3.
	ValidHours float64
}

// SolarEclipseBesselianElementsResult 是一次日食的多项式贝塞尔根数及其口径 / polynomial Besselian elements and the conventions behind them.
type SolarEclipseBesselianElementsResult struct {
	// T0JDE 多项式参考时刻（TT 儒略日），t = (jde - T0JDE) * 24。
	// T0JDE is the TT reference instant; t = (jde - T0JDE) * 24.
	T0JDE float64
	// ValidHours 有效窗口半径（小时），超出该窗口不应使用本多项式。
	// ValidHours is the half-width of the validity window in hours.
	ValidHours float64

	// X 与 Y 是月心在基本面内的坐标，单位地球赤道半径。
	// X and Y are the Moon's fundamental-plane coordinates in equatorial Earth radii.
	X, Y SolarEclipseBesselianPolynomial
	// D 是影轴赤纬，单位度。
	// D is the declination of the shadow axis in degrees.
	D SolarEclipseBesselianPolynomial
	// L1 与 L2 是基本面内的半影、本影半径，单位地球赤道半径；本影为负表示月心尚未越过本影锥顶点。
	// L1 and L2 are the penumbral and umbral radii in the fundamental plane, in equatorial Earth radii.
	L1, L2 SolarEclipseBesselianPolynomial
	// Mu 是影轴格林时角，单位度，窗口内连续、不折回 [0,360)。
	//
	// 口径与已发布根数表不同：本库的恒星时取自 UT = TT - ΔT，得到的是真实格林时角；已发布表改用
	// T0 本身的恒星时（不含 ΔT 自转），两者相差 ΔT × 15.041067/3600 度。要对表先用
	// SolarEclipseBesselianMuForPublishedTable 换算。
	// Mu is the Greenwich hour angle of the shadow axis in degrees, continuous and not folded into [0,360).
	Mu SolarEclipseBesselianPolynomial
	// TanF1 与 TanF2 是半影、本影锥半顶角正切，本次日食内为常数。
	// TanF1 and TanF2 are the penumbral and umbral cone half-angle tangents, constant over the eclipse.
	TanF1, TanF2 float64
	// Gamma 是食甚时刻影轴到地心的距离，单位地球赤道半径。
	// Gamma is the shadow-axis distance from the Earth's centre at greatest eclipse, in equatorial Earth radii.
	Gamma float64
	// Magnitude 是食甚时刻的全局食分。
	// Magnitude is the global eclipse magnitude at greatest eclipse.
	Magnitude float64

	// 下列字段是决定上述数值的口径，随结果一起保留。
	// The fields below are the conventions that fix the numbers above.
	Model          SolarEclipseRadiusModel
	SunRadiusModel SolarEclipseSunRadiusModel
	PenumbralK     float64
	UmbralK        float64
	DeltaTSeconds  float64
}

// SolarEclipseBesselianMuForPublishedTable 把本库的 Mu 换算成与已发布根数表直接可比的取值。
// 已发布表用 T0 本身的恒星时，本库用 UT = TT - ΔT，两者只差一个常数，因此只有常数项平移。
// SolarEclipseBesselianMuForPublishedTable shifts Mu onto the argument used by published element tables.
func SolarEclipseBesselianMuForPublishedTable(
	mu SolarEclipseBesselianPolynomial, deltaTSeconds float64,
) SolarEclipseBesselianPolynomial {
	shift := deltaTSeconds * solarEclipseBesselianSiderealDegreesPerSecond
	return SolarEclipseBesselianPolynomial{mu[0] + shift, mu[1], mu[2], mu[3]}
}

// SolarEclipseBesselianElementsWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func SolarEclipseBesselianElementsWithInstance(
	seedJDE float64, options SolarEclipseBesselianElementsOptions, inst *Instance) (SolarEclipseBesselianElementsResult, bool) {
	options.Model = normalizeSolarEclipseRadiusModel(options.Model)
	options.SunRadiusModel = normalizeSolarEclipseSunRadiusModel(options.SunRadiusModel)
	validHours := options.ValidHours
	if !(validHours > 0) {
		validHours = solarEclipseBesselianDefaultValidHours
	}

	solver := newSolarEclipseSolverWithOptions(solarEclipseAnchorNewMoon(seedJDE, inst), SolarEclipseOptions{
		RadiusModel:    options.Model,
		SunRadiusModel: options.SunRadiusModel,
	}, inst).
		withDeltaTSeconds(options.DeltaTSeconds)
	feature := solver.feature()
	if feature.typeCode == "N" {
		return SolarEclipseBesselianElementsResult{}, false
	}

	t0 := options.ReferenceJDE
	if !(t0 > 0) {
		// 已发布表按最近整小时取 T0（食甚 02:36 TDT 的表 T0 是 03:00），不是取整点下界。
		t0 = math.Round(feature.greatestEclipseJDE*24) / 24
	}

	step := 2 * validHours / float64(solarEclipseBesselianSampleCount-1)
	times := make([]float64, solarEclipseBesselianSampleCount)
	columns := [6][]float64{}
	for index := range columns {
		columns[index] = make([]float64, solarEclipseBesselianSampleCount)
	}
	for index := range times {
		hours := -validHours + step*float64(index)
		times[index] = hours
		point := solver.besselianElementsAt(t0 + hours/24)
		columns[0][index] = point.x
		columns[1][index] = point.y
		columns[2][index] = point.d
		columns[3][index] = point.l1
		columns[4][index] = point.l2
		columns[5][index] = point.mu
	}

	// μ 每窗口跨越的角量远小于 180°，可以先展开成连续序列再归一到 [0,360)。
	columns[5] = solarEclipseUnwrapDegrees(columns[5], solarEclipseBesselianSampleCount/2)

	return SolarEclipseBesselianElementsResult{
		T0JDE:          t0,
		ValidHours:     validHours,
		X:              solarEclipseFitCubic(times, columns[0]),
		Y:              solarEclipseFitCubic(times, columns[1]),
		D:              solarEclipseFitCubic(times, columns[2]),
		L1:             solarEclipseFitCubic(times, columns[3]),
		L2:             solarEclipseFitCubic(times, columns[4]),
		Mu:             solarEclipseFitCubic(times, columns[5]),
		TanF1:          solver.penumbraConeTangent,
		TanF2:          solver.umbraConeTangent,
		Gamma:          feature.gamma,
		Magnitude:      feature.magnitude,
		Model:          options.Model,
		SunRadiusModel: options.SunRadiusModel,
		PenumbralK:     solver.params.penumbralK,
		UmbralK:        solver.params.umbralK,
		DeltaTSeconds:  solver.effectiveDeltaTSeconds(t0),
	}, true
}

// solarEclipseBesselianPoint 是单一 TT 时刻的经典口径贝塞尔根数。
type solarEclipseBesselianPoint struct {
	x, y, z, d, mu, l1, l2 float64
}

// besselianElementsAt 按经典口径取该时刻的根数：d 为影轴赤纬，μ 为真实格林时角，
// L1/L2 用含 1/cos f 的 ES 形式，且本影取负号口径。
func (solver solarEclipseSolver) besselianElementsAt(jde float64) solarEclipseBesselianPoint {
	moon, axis, _ := solver.besselGeometryAt(jde)
	penumbraHalfAngle := math.Atan(solver.penumbraConeTangent)
	umbraHalfAngle := math.Atan(solver.umbraConeTangent)

	return solarEclipseBesselianPoint{
		x:  moon[0],
		y:  moon[1],
		z:  moon[2],
		d:  (math.Pi/2 - axis.tilt) / rad,
		mu: (axis.gst - (axis.rightAscension - math.Pi/2)) / rad,
		l1: moon[2]*solver.penumbraConeTangent + solver.params.penumbralK/math.Cos(penumbraHalfAngle),
		l2: moon[2]*solver.umbraConeTangent - solver.params.umbralK/math.Cos(umbraHalfAngle),
	}
}

// solarEclipseUnwrapDegrees 把按时间升序的角量展开成连续序列，并把 reference 号样本归入 [0,360)。
func solarEclipseUnwrapDegrees(values []float64, reference int) []float64 {
	unwrapped := make([]float64, len(values))
	copy(unwrapped, values)
	for index := 1; index < len(unwrapped); index++ {
		for unwrapped[index]-unwrapped[index-1] > 180 {
			unwrapped[index] -= 360
		}
		for unwrapped[index]-unwrapped[index-1] < -180 {
			unwrapped[index] += 360
		}
	}
	if reference >= 0 && reference < len(unwrapped) {
		shift := 360 * math.Floor(unwrapped[reference]/360)
		for index := range unwrapped {
			unwrapped[index] -= shift
		}
	}
	return unwrapped
}

// solarEclipseFitCubic 用样本做三次最小二乘拟合，样本少于 4 个时返回零值。
func solarEclipseFitCubic(times, values []float64) SolarEclipseBesselianPolynomial {
	if len(times) < 4 || len(times) != len(values) {
		return SolarEclipseBesselianPolynomial{}
	}

	var normal [4][5]float64
	for index := range times {
		powers := [7]float64{1}
		for n := 1; n < len(powers); n++ {
			powers[n] = powers[n-1] * times[index]
		}
		for row := range normal {
			for column := range normal[row][:4] {
				normal[row][column] += powers[row+column]
			}
			normal[row][4] += powers[row] * values[index]
		}
	}

	for column := range normal {
		pivot := column
		for row := column + 1; row < len(normal); row++ {
			if math.Abs(normal[row][column]) > math.Abs(normal[pivot][column]) {
				pivot = row
			}
		}
		normal[column], normal[pivot] = normal[pivot], normal[column]
		if normal[column][column] == 0 {
			return SolarEclipseBesselianPolynomial{}
		}
		for row := range normal {
			if row == column {
				continue
			}
			factor := normal[row][column] / normal[column][column]
			for c := column; c < len(normal[row]); c++ {
				normal[row][c] -= factor * normal[column][c]
			}
		}
	}

	var polynomial SolarEclipseBesselianPolynomial
	for n := range polynomial {
		polynomial[n] = normal[n][4] / normal[n][n]
	}
	return polynomial
}
