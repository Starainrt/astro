package basic

import (
	"math"
	"sort"
)

// 地平闭包根的扫描式枚举。闭包根的定义是三个条件同时成立：站点落在 central limit 上
// （gap=0）、该站点的间隙在时间上取极值（∂gap/∂t=0）、站点落在地平线上（alt=0）。
// 于是可以先把前两个条件在固定时刻化成一维周期求根（「地平线上的食甚点」），再让这些
// 点上的 gap 随时刻穿越零——闭包根就是那次穿越。这样既不需要种子，也不需要三维牛顿的
// 有限差分雅可比；采样分支恰好终止在根上时不会漏根。

const (
	// 地平圈整圈求根的采样数，与升落曲线同口径。
	solarEclipseClosureScanBoundaryPoints = solarEclipseRiseSetBoundaryPoints
	// 时间导数沿地平圈找根时的折点容差，单位与 ∂gap/∂t（每天）一致。
	solarEclipseClosureScanRateFoldTolerance = 1e-3
	// 行数下限与上限；窗口很短时按比例加密，很长时按比例放稀。
	solarEclipseClosureScanMinimumRows = 24
	solarEclipseClosureScanMaximumRows = 720
	// 默认行距 1 min：闭包根之间的间隙可达数十分钟，够给每个符号变化留出样本。
	solarEclipseClosureScanStepDays = 60.0 / 86400.0
	// 相邻行配对的距离上限：同一条闭合弧上的点一行之内不会超过该距离。
	solarEclipseClosureScanMatchKM = 900.0
	// 时间二分上限，1e-9 天约 1e-4 s。
	solarEclipseClosureScanBisectionSteps = 40
	// 牛顿抛光后允许偏离扫描根的上限，超过说明落到了别的分支。
	solarEclipseClosureScanPolishKM = 2.0
)

type solarEclipseClosureScanCandidate struct {
	jde       float64
	longitude float64
	latitude  float64
	gap       float64
}

type solarEclipseClosureScanRow struct {
	jde        float64
	candidates []solarEclipseClosureScanCandidate
}

// centralLimitHorizonRootsByScan 返回一侧窗口内全部地平闭包根，按时间排序。
func (solver solarEclipseSolver) centralLimitHorizonRootsByScan(
	shadowContactJDE, innerContactJDE float64,
) []SolarEclipsePathPoint {
	low, high := math.Min(shadowContactJDE, innerContactJDE), math.Max(shadowContactJDE, innerContactJDE)
	if low <= 0 || high <= low {
		return nil
	}
	// 闭包弧可以把一个根放在接触窗口之外，窗口按既有口径外扩。
	margin := solarEclipseCentralLimitHorizonContactMarginDays
	if span := high - low; span > 0 {
		margin = math.Max(margin, span*0.15)
	}
	start, end := low-margin, high+margin
	rows := solarEclipseClosureScanRowCount(end - start)
	step := (end - start) / float64(rows)
	scan := make([]solarEclipseClosureScanRow, rows+1)
	for index := range scan {
		scan[index] = solarEclipseClosureScanRow{
			jde:        start + float64(index)*step,
			candidates: solver.closureScanCandidatesAt(start + float64(index)*step),
		}
	}
	roots := make([]SolarEclipsePathPoint, 0, 2)
	for index := 1; index < len(scan); index++ {
		for _, bracket := range solarEclipseClosureScanBrackets(scan[index-1], scan[index]) {
			root, ok := solver.refineSolarEclipseClosureScanRoot(bracket)
			if !ok || root.JDE < low-margin || root.JDE > high+margin ||
				solarEclipseRiseSetPointExists(roots, root) {
				continue
			}
			roots = append(roots, root)
		}
	}
	sort.Slice(roots, func(first, second int) bool { return roots[first].JDE < roots[second].JDE })
	return roots
}

func solarEclipseClosureScanRowCount(spanDays float64) int {
	rows := int(spanDays/solarEclipseClosureScanStepDays + 0.5)
	if rows < solarEclipseClosureScanMinimumRows {
		rows = solarEclipseClosureScanMinimumRows
	}
	if rows > solarEclipseClosureScanMaximumRows {
		rows = solarEclipseClosureScanMaximumRows
	}
	return rows
}

// closureScanCandidatesAt 枚举该时刻地平线上全部 ∂gap/∂t=0 的点，并给出各点的 gap。
func (solver solarEclipseSolver) closureScanCandidatesAt(jde float64) []solarEclipseClosureScanCandidate {
	evaluation := solver.closureScanEvaluationAt(jde)
	sun := solarEclipseXYZToLLR(
		evaluation.center.sunXYZ[0], evaluation.center.sunXYZ[1], evaluation.center.sunXYZ[2],
	)
	centerLongitude := normalizeLongitude((sun[0] - evaluation.center.gst) / rad)
	centerLatitude := sun[1] / rad
	valueAt := func(angle float64) (float64, bool) {
		longitude, latitude := riseSetHorizonPoint(centerLongitude, centerLatitude, angle)
		rate := evaluation.centralContactDerivative(longitude, latitude)
		return rate, finite(rate)
	}
	angles := riseSetCyclicRootsWithFoldTolerance(
		solarEclipseClosureScanBoundaryPoints, solarEclipseClosureScanRateFoldTolerance, valueAt,
	)
	candidates := make([]solarEclipseClosureScanCandidate, 0, len(angles))
	for _, angle := range angles {
		longitude, latitude := riseSetHorizonPoint(centerLongitude, centerLatitude, angle)
		candidate, ok := solver.closureScanCandidateAt(evaluation, longitude, latitude)
		if !ok {
			continue
		}
		candidates = append(candidates, candidate)
	}
	return candidates
}

// closureScanCandidateAt 把固定时刻的地理起点修正到 {∂gap/∂t=0, alt=0} 上。
func (solver solarEclipseSolver) closureScanCandidateAt(
	evaluation solarEclipseRiseSetEvaluation,
	longitude, latitude float64,
) (solarEclipseClosureScanCandidate, bool) {
	// 不能沿用 riseSetRefineGeographicRoot：∂gap/∂t 是 5 s 有限差分，残差噪声约 1e-7，
	// 该函数的收尾判据 1e-8 永远达不到；这里的容差与三维牛顿的收敛判据保持一致。
	const stepDegrees = 1e-4
	residualAt := func(longitude, latitude float64) (float64, float64) {
		state := evaluation.center.stateAt(longitude*rad, latitude*rad, 0)
		return evaluation.centralContactDerivative(longitude, latitude), state.sunAltitudeRad
	}
	for iteration := 0; iteration < 8; iteration++ {
		rate, altitude := residualAt(longitude, latitude)
		if !finite(rate) || !finite(altitude) {
			return solarEclipseClosureScanCandidate{}, false
		}
		if math.Abs(rate) <= 1e-7 && math.Abs(altitude) <= 1e-9 {
			break
		}
		shiftedLongitudeRate, shiftedLongitudeAltitude := residualAt(longitude+stepDegrees, latitude)
		shiftedLatitudeRate, shiftedLatitudeAltitude := residualAt(longitude, latitude+stepDegrees)
		a := (shiftedLongitudeRate - rate) / stepDegrees
		b := (shiftedLatitudeRate - rate) / stepDegrees
		c := (shiftedLongitudeAltitude - altitude) / stepDegrees
		d := (shiftedLatitudeAltitude - altitude) / stepDegrees
		determinant := a*d - b*c
		if !finite(determinant) || math.Abs(determinant) < 1e-18 {
			return solarEclipseClosureScanCandidate{}, false
		}
		deltaLongitude := (-rate*d + b*altitude) / determinant
		deltaLatitude := (c*rate - a*altitude) / determinant
		if scale := math.Max(math.Abs(deltaLongitude), math.Abs(deltaLatitude)); scale > 5 {
			deltaLongitude *= 5 / scale
			deltaLatitude *= 5 / scale
		}
		longitude += deltaLongitude
		latitude += deltaLatitude
		if !finite(longitude) || !finite(latitude) || latitude <= -89.999 || latitude >= 89.999 {
			return solarEclipseClosureScanCandidate{}, false
		}
	}
	rate, altitude := residualAt(longitude, latitude)
	if !finite(rate) || !finite(altitude) || math.Abs(rate) > 1e-6 || math.Abs(altitude) > 1e-8 {
		return solarEclipseClosureScanCandidate{}, false
	}
	state := evaluation.center.stateAt(longitude*rad, latitude*rad, 0)
	gap := solarEclipseCentralContactGap(state)
	if !finite(gap) {
		return solarEclipseClosureScanCandidate{}, false
	}
	return solarEclipseClosureScanCandidate{
		jde: evaluation.jd, longitude: longitude, latitude: latitude, gap: gap,
	}, true
}

// closureScanEvaluationAt 同时给出中心与前后时刻的星历态：时间导数需要前后两点。
func (solver solarEclipseSolver) closureScanEvaluationAt(jde float64) solarEclipseRiseSetEvaluation {
	return solarEclipseRiseSetEvaluation{
		jd:     jde,
		center: solver.localStateContextAt(jde),
		before: solver.localStateContextAt(jde - solarEclipseRiseSetDerivativeStepDays),
		after:  solver.localStateContextAt(jde + solarEclipseRiseSetDerivativeStepDays),
	}
}

// solarEclipseClosureScanBrackets 在相邻两行之间按最近距离配对，返回 gap 变号的分支区间。
func solarEclipseClosureScanBrackets(
	previous, current solarEclipseClosureScanRow,
) [][2]solarEclipseClosureScanCandidate {
	if len(previous.candidates) == 0 || len(current.candidates) == 0 {
		return nil
	}
	type pair struct {
		previous int
		current  int
		distance float64
	}
	pairs := make([]pair, 0, len(previous.candidates)*len(current.candidates))
	for first := range previous.candidates {
		for second := range current.candidates {
			distance := solarEclipsePathDistanceKM(
				SolarEclipsePathPoint{
					Longitude: previous.candidates[first].longitude,
					Latitude:  previous.candidates[first].latitude,
				},
				SolarEclipsePathPoint{
					Longitude: current.candidates[second].longitude,
					Latitude:  current.candidates[second].latitude,
				},
			)
			if distance > solarEclipseClosureScanMatchKM {
				continue
			}
			pairs = append(pairs, pair{first, second, distance})
		}
	}
	sort.Slice(pairs, func(first, second int) bool { return pairs[first].distance < pairs[second].distance })
	usedPrevious := make([]bool, len(previous.candidates))
	usedCurrent := make([]bool, len(current.candidates))
	brackets := make([][2]solarEclipseClosureScanCandidate, 0, 2)
	for _, candidate := range pairs {
		if usedPrevious[candidate.previous] || usedCurrent[candidate.current] {
			continue
		}
		usedPrevious[candidate.previous] = true
		usedCurrent[candidate.current] = true
		left, right := previous.candidates[candidate.previous], current.candidates[candidate.current]
		if (left.gap <= 0) != (right.gap <= 0) {
			brackets = append(brackets, [2]solarEclipseClosureScanCandidate{left, right})
		}
	}
	return brackets
}

// refineSolarEclipseClosureScanRoot 在时间上二分 gap 的零点，再用既有三维牛顿抛光到同一容差。
func (solver solarEclipseSolver) refineSolarEclipseClosureScanRoot(
	bracket [2]solarEclipseClosureScanCandidate,
) (SolarEclipsePathPoint, bool) {
	start, end := bracket[0], bracket[1]
	if start.gap == 0 {
		return solver.closureScanPointAt(start)
	}
	for iteration := 0; iteration < solarEclipseClosureScanBisectionSteps; iteration++ {
		if end.jde-start.jde <= 1e-9 {
			break
		}
		middle := 0.5 * (start.jde + end.jde)
		longitude := 0.5 * (start.longitude + end.longitude)
		latitude := 0.5 * (start.latitude + end.latitude)
		candidate, ok := solver.closureScanCandidateAt(
			solver.closureScanEvaluationAt(middle), longitude, latitude,
		)
		if !ok {
			return SolarEclipsePathPoint{}, false
		}
		if (start.gap <= 0) == (candidate.gap <= 0) {
			start = candidate
		} else {
			end = candidate
		}
	}
	return solver.closureScanPointAt(start)
}

// closureScanPointAt 把扫描候选抛光到三维系统的同一容差；扫描只负责给出种子。
func (solver solarEclipseSolver) closureScanPointAt(
	candidate solarEclipseClosureScanCandidate,
) (SolarEclipsePathPoint, bool) {
	polished, ok := solveSolarEclipseCentralLimitHorizonRoot(
		solver, [3]float64{candidate.longitude, candidate.latitude, candidate.jde},
	)
	if !ok {
		return SolarEclipsePathPoint{}, false
	}
	point := SolarEclipsePathPoint{
		JDE: candidate.jde, Longitude: candidate.longitude, Latitude: candidate.latitude,
	}
	if solarEclipsePathDistanceKM(polished, point) > solarEclipseClosureScanPolishKM {
		return SolarEclipsePathPoint{}, false
	}
	return polished, true
}
