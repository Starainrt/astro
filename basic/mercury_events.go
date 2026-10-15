package basic

import (
	"math"

	. "github.com/starainrt/astro/tools"
)

const (
	MERCURY_S_PERIOD                    = 1 / ((1 / 87.9691) - (1 / 365.256363004))
	mercuryConjunctionDerivativeStepDay = 2e-5 * 36525.0
	mercuryLightTimeDaysPerAU           = 0.0057755183
	mercuryEventSearchN                 = 16
	mercuryStationWindowDays            = 30.0
	mercuryStationDerivativeStepDay     = 0.01
	mercuryStationCoarseStepDay         = 2.0
	mercuryStationHalfWindowDay         = 2.0
	mercuryStationMotionTolerance       = 1e-3
	// mercuryStationAnchorAttempts 类型化「留」在锚点不满足侧向不变量时，最多推进/回退几个会合周期。
	mercuryStationAnchorAttempts = 4
	// mercuryConjunctionSameInstantDegrees 合搜索「同刻」快速路径的视黄经差触发阈值（度）。
	// 0.05 度约合 30 分钟的时间跨度，远大于截断级数在根附近的误差，
	// 从而保证「查询落在合附近」时一定走精确解 + 侧向判定，而不是被方向扫描跨过去。
	mercuryConjunctionSameInstantDegrees = 5.0e-2
	// mercuryConjunctionScanStepDay / mercuryConjunctionScanMaxSteps 方向性括号扫描参数：
	// 8 天一步 × 80 步 = 640 天，覆盖一个水星会合周期（115.88 天）且有大量余量。
	mercuryConjunctionScanStepDay  = 8.0
	mercuryConjunctionScanMaxSteps = 80
	// mercuryConjunctionPolishToleranceDay 全项抛光的时间容差（0.01 秒）：
	// 割线法从 8 天括号收敛到该量级只需多 1~2 次求值，但把合时刻精度保持在毫秒级
	// （与旧实现牛顿法的 1e-5 天口径一致）。
	mercuryConjunctionPolishToleranceDay = 0.01 / 86400.0
)

type mercuryConjunctionLBR struct {
	lo float64
	bo float64
	r  float64
}

type mercuryConjunctionGeo struct {
	lo   float64
	bo   float64
	dist float64
}

type mercuryConjunctionResult struct {
	diff         float64
	sunLightDays float64
	geoLightDays float64
}

func mercuryHelioN(planetIndex int, jde float64, n int, instances ...*Instance) mercuryConjunctionLBR {
	inst := instanceValue(instances)

	return mercuryConjunctionLBR{
		lo: PlanetCoordinateNWithInstance(planetIndex, 0, jde, n, inst),
		bo: PlanetCoordinateNWithInstance(planetIndex, 1, jde, n, inst),
		r:  PlanetCoordinateNWithInstance(planetIndex, 2, jde, n, inst),
	}
}

func mercuryGeocentric(planetPos, earthPos mercuryConjunctionLBR) mercuryConjunctionGeo {
	x := planetPos.r*Cos(planetPos.bo)*Cos(planetPos.lo) - earthPos.r*Cos(earthPos.bo)*Cos(earthPos.lo)
	y := planetPos.r*Cos(planetPos.bo)*Sin(planetPos.lo) - earthPos.r*Cos(earthPos.bo)*Sin(earthPos.lo)
	z := planetPos.r*Sin(planetPos.bo) - earthPos.r*Sin(earthPos.bo)
	dist := math.Sqrt(x*x + y*y + z*z)
	return mercuryConjunctionGeo{
		lo:   Limit360(math.Atan2(y, x) * 180 / math.Pi),
		bo:   math.Atan2(z, math.Sqrt(x*x+y*y)) * 180 / math.Pi,
		dist: dist,
	}
}

func mercuryConjunctionAngleDelta(diff float64) float64 {
	diff = Limit360(diff)
	if diff > 180 {
		diff -= 360
	}
	if diff < -180 {
		diff += 360
	}
	return diff
}

func mercuryConjunctionHeliocentricDelta(jde, targetDeg float64, n int, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	planetLo := PlanetCoordinateNWithInstance(1, 0, jde, n, inst)
	earthLo := PlanetCoordinateNWithInstance(-1, 0, jde, n, inst)
	return mercuryConjunctionAngleDelta(planetLo - earthLo - targetDeg)
}

func mercuryConjunctionDifference(jd float64, n int, targetDeg, sunLightDays, geoLightDays float64, instances ...*Instance) mercuryConjunctionResult {
	inst := instanceValue(instances)

	earthForSun := mercuryHelioN(-1, jd-sunLightDays, n, inst)
	sunLo := Limit360(earthForSun.lo + 180)
	earth := mercuryHelioN(-1, jd-geoLightDays, n, inst)
	planetPos := mercuryHelioN(1, jd-geoLightDays, n, inst)
	geo := mercuryGeocentric(planetPos, earth)
	return mercuryConjunctionResult{
		diff:         mercuryConjunctionAngleDelta(geo.lo - sunLo - targetDeg),
		sunLightDays: earthForSun.r * mercuryLightTimeDaysPerAU,
		geoLightDays: geo.dist * mercuryLightTimeDaysPerAU,
	}
}

func mercuryConjunctionExactDelta(jde float64, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	return mercuryConjunctionAngleDelta(MercuryApparentLoWithInstance(jde, inst) - HSunApparentLoWithInstance(jde, inst))
}

func mercuryConjunctionApproxTT(seed float64, inferior bool, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	heliocentricTarget := 180.0
	if inferior {
		heliocentricTarget = 0
	}
	jde := seed
	for i := 0; i < 6; i++ {
		jde -= mercuryConjunctionHeliocentricDelta(jde, heliocentricTarget, 8, inst) / (360.0 / MERCURY_S_PERIOD)
	}

	startSample := mercuryConjunctionDifference(jde, 8, 0, 0, 0, inst)
	nextSample := mercuryConjunctionDifference(jde+mercuryConjunctionDerivativeStepDay, 8, 0, 0, 0, inst)
	diffSlope := mercuryConjunctionAngleDelta(nextSample.diff-startSample.diff) / mercuryConjunctionDerivativeStepDay

	refined := mercuryConjunctionDifference(jde, 40, 0, startSample.sunLightDays, startSample.geoLightDays, inst)
	jde -= refined.diff / diffSlope
	final := mercuryConjunctionDifference(jde, -1, 0, refined.sunLightDays, refined.geoLightDays, inst)
	jde -= final.diff / diffSlope
	return jde
}

func mercuryConjunctionExactTT(seed float64, inferior bool, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	estimateJD := mercuryConjunctionApproxTT(seed, inferior, inst)
	converged := false
	for i := 0; i < eventNewtonMaxIterations; i++ {
		prevJDE := estimateJD
		longitudeDelta := mercuryConjunctionExactDelta(prevJDE, inst)
		longitudeSlope := (mercuryConjunctionExactDelta(prevJDE+0.000005, inst) - mercuryConjunctionExactDelta(prevJDE-0.000005, inst)) / 0.00001
		nextJD := prevJDE - longitudeDelta/longitudeSlope
		estimateJD = nextJD
		if math.Abs(nextJD-prevJDE) <= 0.00001 {
			converged = true
			break
		}
	}
	if !converged {
		return math.NaN()
	}
	return estimateJD
}

// mercuryConjunction 在 jde 的指定方向上求最近一次水星合（内合/外合）。
//
// 旧实现用「启发式跳 + 2 天一步走」：跳过头落到合之后时，|Δ| 会持续变大，
// 于是走满一个会合周期、跳过一次合（例如 2008-01-22 查到 2008-06-07 而不是 2008-02-06），
// 并且下游的类型化「留」会因此返回错事件、甚至让 Last 返回未来。
// 新实现改成方向性括号扫描：截断级数逐段找异号区间（顺序扫描 ⇒ 一定取最近的一个合），
// 再用全项级数在括号内抛光；扫满 640 天仍无括号时有界返回 NaN。
func mercuryConjunction(jde float64, next uint8, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	//0=last 1=next
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	if math.Abs(mercuryConjunctionExactDelta(jde, inst)) <= mercuryConjunctionSameInstantDegrees {
		best := math.NaN()
		consider := func(inferior bool) {
			eventUT := TT2UTC(mercuryConjunctionExactTT(jde, inferior, inst))
			if !isFiniteFloat(eventUT) {
				return
			}
			if next == 0 && !eventUTQueryBeforeOrEqual(eventUT, jde) {
				return
			}
			if next == 1 && !eventUTQueryAfterOrEqual(eventUT, jde) {
				return
			}
			if math.IsNaN(best) || math.Abs(eventUTQueryTTDelta(eventUT, jde)) < math.Abs(eventUTQueryTTDelta(best, jde)) {
				best = eventUT
			}
		}
		consider(true)
		consider(false)
		if !math.IsNaN(best) {
			return best
		}
	}

	direction := 1.0
	if next == 0 {
		direction = -1
	}
	leftJDE := jde
	leftValue := mercuryConjunctionDeltaN(leftJDE, mercuryEventSearchN, inst)
	if !isFiniteFloat(leftValue) {
		return math.NaN()
	}
	for i := 0; i < mercuryConjunctionScanMaxSteps; i++ {
		rightJDE := jde + direction*mercuryConjunctionScanStepDay*float64(i+1)
		rightValue := mercuryConjunctionDeltaN(rightJDE, mercuryEventSearchN, inst)
		if !isFiniteFloat(rightValue) {
			return math.NaN()
		}
		if leftValue == 0 || rightValue == 0 || leftValue*rightValue < 0 {
			return mercuryConjunctionPolish(jde, leftJDE, rightJDE, direction, inst)
		}
		leftJDE, leftValue = rightJDE, rightValue
	}
	return math.NaN()
}

// mercuryConjunctionDeltaN 截断级数下的水星-太阳视黄经差（度，[-180,180]），用于方向性括号扫描。
func mercuryConjunctionDeltaN(jde float64, n int, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	return mercuryConjunctionAngleDelta(MercuryApparentLoNWithInstance(jde, n, inst) - HSunApparentLoNWithInstance(jde, n, inst))
}

// mercuryConjunctionPolish 用全项级数在截断级数给出的括号内抛光。
// 截断误差可能让括号两端在全项函数上同号（罕见），此时沿扫描方向再扩一两个扫描步；
// 若仍未被确认（典型情形：查询几乎正好落在合上，截断级数在根两侧的符号与全项不一致），
// 退回全项级数的方向扫描，保证有界且不返回 NaN。
func mercuryConjunctionPolish(jde, leftJDE, rightJDE, direction float64, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	for attempt := 0; attempt < 3; attempt++ {
		leftValue := mercuryConjunctionExactDelta(leftJDE, inst)
		rightValue := mercuryConjunctionExactDelta(rightJDE, inst)
		if !isFiniteFloat(leftValue) || !isFiniteFloat(rightValue) {
			return math.NaN()
		}
		if leftValue == 0 || rightValue == 0 || leftValue*rightValue < 0 {
			root, ok := eventBracketSecantRoot(leftJDE, rightJDE, leftValue, rightValue,
				mercuryConjunctionPolishToleranceDay, func(jde float64) float64 { return mercuryConjunctionExactDelta(jde, inst) })
			if !ok {
				return math.NaN()
			}
			return TT2UTC(root)
		}
		if direction > 0 {
			rightJDE += mercuryConjunctionScanStepDay
			continue
		}
		leftJDE -= mercuryConjunctionScanStepDay
	}
	return mercuryConjunctionFullDirectionalScan(jde, direction, inst)
}

// mercuryConjunctionFullDirectionalScan 全项级数的方向扫描（截断括号未被确认时的兜底）。
func mercuryConjunctionFullDirectionalScan(jde, direction float64, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	leftJDE := jde
	leftValue := mercuryConjunctionExactDelta(leftJDE, inst)
	if !isFiniteFloat(leftValue) {
		return math.NaN()
	}
	for i := 0; i < mercuryConjunctionScanMaxSteps; i++ {
		rightJDE := jde + direction*mercuryConjunctionScanStepDay*float64(i+1)
		rightValue := mercuryConjunctionExactDelta(rightJDE, inst)
		if !isFiniteFloat(rightValue) {
			return math.NaN()
		}
		if leftValue == 0 || rightValue == 0 || leftValue*rightValue < 0 {
			root, ok := eventBracketSecantRoot(leftJDE, rightJDE, leftValue, rightValue,
				mercuryConjunctionPolishToleranceDay, func(jde float64) float64 { return mercuryConjunctionExactDelta(jde, inst) })
			if !ok {
				return math.NaN()
			}
			return TT2UTC(root)
		}
		leftJDE, leftValue = rightJDE, rightValue
	}
	return math.NaN()
}

// LastMercuryConjunctionWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastMercuryConjunctionWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastSimpleEvent(jde, func(jde float64) float64 { return LastMercuryConjunctionStrictWithInstance(jde, inst) }, func(jde float64) float64 { return NextMercuryConjunctionStrictWithInstance(jde, inst) })
}

// NextMercuryConjunctionWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextMercuryConjunctionWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextSimpleEvent(jde, func(jde float64) float64 { return LastMercuryConjunctionStrictWithInstance(jde, inst) }, func(jde float64) float64 { return NextMercuryConjunctionStrictWithInstance(jde, inst) })
}

// LastMercuryConjunctionStrictWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastMercuryConjunctionStrictWithInstance(jde float64, inst *Instance) float64 {
	return mercuryConjunction(jde, 0, inst)
}

// NextMercuryConjunctionStrictWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextMercuryConjunctionStrictWithInstance(jde float64, inst *Instance) float64 {
	return mercuryConjunction(jde, 1, inst)
}

// NextMercuryInferiorConjunctionWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextMercuryInferiorConjunctionWithInstance(jde float64, inst *Instance) float64 {
	date := NextMercuryConjunctionStrictWithInstance(jde, inst)
	if EarthMercuryAwayWithInstance(date, inst) > EarthAwayWithInstance(date, inst) {
		return NextMercuryConjunctionStrictWithInstance(date+2, inst)
	}
	return date
}

// NextMercurySuperiorConjunctionWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextMercurySuperiorConjunctionWithInstance(jde float64, inst *Instance) float64 {
	date := NextMercuryConjunctionStrictWithInstance(jde, inst)
	if EarthMercuryAwayWithInstance(date, inst) < EarthAwayWithInstance(date, inst) {
		return NextMercuryConjunctionStrictWithInstance(date+2, inst)
	}
	return date
}

// LastMercuryInferiorConjunctionWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastMercuryInferiorConjunctionWithInstance(jde float64, inst *Instance) float64 {
	date := LastMercuryConjunctionStrictWithInstance(jde, inst)
	if EarthMercuryAwayWithInstance(date, inst) > EarthAwayWithInstance(date, inst) {
		return LastMercuryConjunctionStrictWithInstance(date-2, inst)
	}
	return date
}

// LastMercurySuperiorConjunctionWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastMercurySuperiorConjunctionWithInstance(jde float64, inst *Instance) float64 {
	date := LastMercuryConjunctionStrictWithInstance(jde, inst)
	if EarthMercuryAwayWithInstance(date, inst) < EarthAwayWithInstance(date, inst) {
		return LastMercuryConjunctionStrictWithInstance(date-2, inst)
	}
	return date
}

func mercuryRADerivative(jde, delta float64, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	sub := MercuryApparentRaWithInstance(jde+delta, inst) - MercuryApparentRaWithInstance(jde-delta, inst)
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub / (2 * delta)
}

func mercuryRADerivativeN(jde, delta float64, n int, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	sub := MercuryApparentRaNWithInstance(jde+delta, n, inst) - MercuryApparentRaNWithInstance(jde-delta, n, inst)
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub / (2 * delta)
}

func mercuryStationInWindow(startTT, endTT float64, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	bestJDE := zeroEventInWindow(startTT, endTT, mercuryStationCoarseStepDay, mercuryStationHalfWindowDay, 30.0/86400.0, func(jd float64) float64 {
		return mercuryRADerivativeN(jd, mercuryStationDerivativeStepDay, mercuryEventSearchN, inst)
	}, func(jd float64) float64 {
		return mercuryRADerivative(jd, mercuryStationDerivativeStepDay, inst)
	})
	return TT2UTC(bestJDE)
}

func mercuryStationBetween(startTT, endTT float64, instances ...*Instance) bool {
	inst := instanceValue(instances)

	if endTT < startTT {
		startTT, endTT = endTT, startTT
	}
	if endTT-startTT <= 0 {
		return false
	}
	// 跨度超过一个回留窗口时截断扫描已不可靠：返回 false 不再阻断快速路径。
	// 此时候选事件由 mercuryConjunction 的完整方向扫描保证，快速路径无需再复核。
	// A span beyond the station window makes the truncated scan unreliable; report false so the
	// fast path is not blocked. The typed candidate is already guaranteed by mercuryConjunction's
	// exhaustive directional scan.
	if endTT-startTT > mercuryStationWindowDays {
		return false
	}
	// 截断扫描足以判断单候选快速路径是否安全 / A truncated scan is enough to decide whether the one-candidate fast path is safe.
	left := startTT
	leftValue := mercuryRADerivativeN(left, mercuryStationDerivativeStepDay, mercuryEventSearchN, inst)
	for left < endTT {
		right := left + mercuryStationCoarseStepDay
		if right > endTT {
			right = endTT
		}
		rightValue := mercuryRADerivativeN(right, mercuryStationDerivativeStepDay, mercuryEventSearchN, inst)
		if leftValue == 0 || leftValue*rightValue < 0 || rightValue == 0 {
			return true
		}
		left = right
		leftValue = rightValue
	}
	return false
}

func mercuryProgradeToRetrogradeAroundInferior(inferiorUT float64, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	inferiorTT := UTC2TT(inferiorUT)
	return mercuryStationInWindow(inferiorTT-mercuryStationWindowDays, inferiorTT, inst)
}

func mercuryRetrogradeToProgradeAroundInferior(inferiorUT float64, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	inferiorTT := UTC2TT(inferiorUT)
	return mercuryStationInWindow(inferiorTT, inferiorTT+mercuryStationWindowDays, inst)
}

// NextMercuryProgradeToRetrogradeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextMercuryProgradeToRetrogradeWithInstance(jde float64, inst *Instance) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	// 以下合为锚点求候选，不满足 Next 不早于查询时推进一个会合周期，保证不返回过去的事件。
	inferior := NextMercuryInferiorConjunctionWithInstance(jde, inst)
	for i := 0; i < mercuryStationAnchorAttempts && isFiniteFloat(inferior); i++ {
		date := mercuryProgradeToRetrogradeAroundInferior(inferior, inst)
		if isFiniteFloat(date) && stationUTQueryAfterOrEqual(date, jde) {
			return date
		}
		inferior = NextMercuryInferiorConjunctionWithInstance(eventUTNextQueryTT(inferior), inst)
	}
	return math.NaN()
}

// NextMercuryRetrogradeToProgradeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextMercuryRetrogradeToProgradeWithInstance(jde float64, inst *Instance) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	inferior := LastMercuryInferiorConjunctionWithInstance(jde, inst)
	for i := 0; i < mercuryStationAnchorAttempts && isFiniteFloat(inferior); i++ {
		date := mercuryRetrogradeToProgradeAroundInferior(inferior, inst)
		if isFiniteFloat(date) && stationUTQueryAfterOrEqual(date, jde) {
			return date
		}
		inferior = NextMercuryInferiorConjunctionWithInstance(eventUTNextQueryTT(inferior), inst)
	}
	return math.NaN()
}

// LastMercuryProgradeToRetrogradeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastMercuryProgradeToRetrogradeWithInstance(jde float64, inst *Instance) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	inferior := NextMercuryInferiorConjunctionWithInstance(jde, inst)
	for i := 0; i < mercuryStationAnchorAttempts && isFiniteFloat(inferior); i++ {
		date := mercuryProgradeToRetrogradeAroundInferior(inferior, inst)
		if isFiniteFloat(date) && stationUTQueryBeforeOrEqual(date, jde) {
			return date
		}
		inferior = LastMercuryInferiorConjunctionWithInstance(eventUTLastQueryTT(inferior), inst)
	}
	return math.NaN()
}

// LastMercuryRetrogradeToProgradeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastMercuryRetrogradeToProgradeWithInstance(jde float64, inst *Instance) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	inferior := LastMercuryInferiorConjunctionWithInstance(jde, inst)
	for i := 0; i < mercuryStationAnchorAttempts && isFiniteFloat(inferior); i++ {
		date := mercuryRetrogradeToProgradeAroundInferior(inferior, inst)
		if isFiniteFloat(date) && stationUTQueryBeforeOrEqual(date, jde) {
			return date
		}
		inferior = LastMercuryInferiorConjunctionWithInstance(eventUTLastQueryTT(inferior), inst)
	}
	return math.NaN()
}

// earliestFiniteEventUT / latestFiniteEventUT 在候选中取最早/最晚的有限值（全部非有限时返回 NaN）。
func earliestFiniteEventUT(candidates ...float64) float64 {
	best := math.NaN()
	for _, candidate := range candidates {
		if !isFiniteFloat(candidate) {
			continue
		}
		if math.IsNaN(best) || candidate < best {
			best = candidate
		}
	}
	return best
}

func latestFiniteEventUT(candidates ...float64) float64 {
	best := math.NaN()
	for _, candidate := range candidates {
		if !isFiniteFloat(candidate) {
			continue
		}
		if math.IsNaN(best) || candidate > best {
			best = candidate
		}
	}
	return best
}

func nextMercuryRetrogradeFromTyped(jde float64, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	return earliestFiniteEventUT(NextMercuryProgradeToRetrogradeWithInstance(jde, inst), NextMercuryRetrogradeToProgradeWithInstance(jde, inst))
}

// NextMercuryRetrogradeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextMercuryRetrogradeWithInstance(jde float64, inst *Instance) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	motion := mercuryRADerivative(jde, mercuryStationDerivativeStepDay, inst)
	if motion > mercuryStationMotionTolerance {
		p2r := NextMercuryProgradeToRetrogradeWithInstance(jde, inst)
		if isFiniteFloat(p2r) && !mercuryStationBetween(jde, UTC2TT(p2r), inst) {
			return p2r
		}
		best := earliestFiniteEventUT(p2r, NextMercuryRetrogradeToProgradeWithInstance(jde, inst))
		if !stationUTQueryAfterOrEqual(best, jde) {
			return math.NaN()
		}
		return best
	}
	if motion < -mercuryStationMotionTolerance {
		r2p := NextMercuryRetrogradeToProgradeWithInstance(jde, inst)
		if isFiniteFloat(r2p) && !mercuryStationBetween(jde, UTC2TT(r2p), inst) {
			return r2p
		}
		best := earliestFiniteEventUT(NextMercuryProgradeToRetrogradeWithInstance(jde, inst), r2p)
		if !stationUTQueryAfterOrEqual(best, jde) {
			return math.NaN()
		}
		return best
	}
	return nextMercuryRetrogradeFromTyped(jde, inst)
}

func lastMercuryRetrogradeFromTyped(jde float64, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	return latestFiniteEventUT(LastMercuryProgradeToRetrogradeWithInstance(jde, inst), LastMercuryRetrogradeToProgradeWithInstance(jde, inst))
}

// LastMercuryRetrogradeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastMercuryRetrogradeWithInstance(jde float64, inst *Instance) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	motion := mercuryRADerivative(jde, mercuryStationDerivativeStepDay, inst)
	if motion > mercuryStationMotionTolerance {
		r2p := LastMercuryRetrogradeToProgradeWithInstance(jde, inst)
		if isFiniteFloat(r2p) && !mercuryStationBetween(UTC2TT(r2p), jde, inst) {
			return r2p
		}
		best := latestFiniteEventUT(LastMercuryProgradeToRetrogradeWithInstance(jde, inst), r2p)
		if !stationUTQueryBeforeOrEqual(best, jde) {
			return math.NaN()
		}
		return best
	}
	if motion < -mercuryStationMotionTolerance {
		p2r := LastMercuryProgradeToRetrogradeWithInstance(jde, inst)
		if isFiniteFloat(p2r) && !mercuryStationBetween(UTC2TT(p2r), jde, inst) {
			return p2r
		}
		best := latestFiniteEventUT(p2r, LastMercuryRetrogradeToProgradeWithInstance(jde, inst))
		if !stationUTQueryBeforeOrEqual(best, jde) {
			return math.NaN()
		}
		return best
	}
	return lastMercuryRetrogradeFromTyped(jde, inst)
}

func LastMercuryRetrogradeStrict(jde float64) float64 {
	return LastMercuryRetrograde(jde)
}

func NextMercuryRetrogradeStrict(jde float64) float64 {
	return NextMercuryRetrograde(jde)
}

// MercurySunElongationWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MercurySunElongationWithInstance(jde float64, inst *Instance) float64 {
	lo1, bo1 := MercuryApparentLoBoWithInstance(jde, inst)
	lo2 := HSunApparentLoWithInstance(jde, inst)
	bo2 := HSunTrueBoWithInstance(jde, inst)
	return StarAngularSeparation(lo1, bo1, lo2, bo2)
}

func mercurySunElongationN(jde float64, n int, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	lo1, bo1 := MercuryApparentLoBoNWithInstance(jde, n, inst)
	lo2 := HSunApparentLoNWithInstance(jde, n, inst)
	bo2 := HSunTrueBoNWithInstance(jde, n, inst)
	return StarAngularSeparation(lo1, bo1, lo2, bo2)
}

// mercuryGreatestElongationInWindow 求窗口内大距：目标是公开 MercurySunElongation 的极大（视距角），
// 而不是忽略光行差/视位置修正的真距角，否则返回的时刻不是调用方能量到的那个极值。
// 窗口两端是世界时，目标函数收力学时，因此逐次换算。
func mercuryGreatestElongationInWindow(start, end float64, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	return maximizeInWindow(start, end, 2.0, func(utJD float64) float64 {
		return mercurySunElongationN(UTC2TT(utJD), mercuryEventSearchN, inst)
	}, func(utJD float64) float64 {
		return MercurySunElongationWithInstance(UTC2TT(utJD), inst)
	})
}

func mercuryEastElongationWindowEndingAt(inferior float64, instances ...*Instance) (float64, float64) {
	inst := instanceValue(instances)

	lastSuperior := LastMercurySuperiorConjunctionWithInstance(eventUTLastQueryTT(inferior), inst)
	return lastSuperior + innerEventWindowPadding, inferior - innerEventWindowPadding
}

func mercuryWestElongationWindowEndingAt(superior float64, instances ...*Instance) (float64, float64) {
	inst := instanceValue(instances)

	lastInferior := LastMercuryInferiorConjunctionWithInstance(eventUTLastQueryTT(superior), inst)
	return lastInferior + innerEventWindowPadding, superior - innerEventWindowPadding
}

func mercuryEastElongationWindowContaining(jde float64, instances ...*Instance) (float64, float64) {
	inst := instanceValue(instances)

	nextInferior := NextMercuryInferiorConjunctionWithInstance(jde, inst)
	start, end := mercuryEastElongationWindowEndingAt(nextInferior, inst)
	if eventUTQueryBeforeOrEqual(start, jde) {
		return start, end
	}
	currentInferior := LastMercuryInferiorConjunctionWithInstance(jde, inst)
	return mercuryEastElongationWindowEndingAt(currentInferior, inst)
}

func mercuryWestElongationWindowContaining(jde float64, instances ...*Instance) (float64, float64) {
	inst := instanceValue(instances)

	nextSuperior := NextMercurySuperiorConjunctionWithInstance(jde, inst)
	start, end := mercuryWestElongationWindowEndingAt(nextSuperior, inst)
	if eventUTQueryBeforeOrEqual(start, jde) {
		return start, end
	}
	currentSuperior := LastMercurySuperiorConjunctionWithInstance(jde, inst)
	return mercuryWestElongationWindowEndingAt(currentSuperior, inst)
}

func nextMercuryGreatestElongationTyped(jde float64, east bool, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	if east {
		start, windowEnd := mercuryEastElongationWindowContaining(jde, inst)
		for i := 0; i < eventDirectionalSearchIterations; i++ {
			date := mercuryGreatestElongationInWindow(start, windowEnd, inst)
			if !isFiniteFloat(date) {
				return math.NaN()
			}
			if eventUTQueryAfterOrEqual(date, jde) {
				return date
			}
			nextInferior := NextMercuryInferiorConjunctionWithInstance(eventUTNextQueryTT(windowEnd), inst)
			start, windowEnd = mercuryEastElongationWindowEndingAt(nextInferior, inst)
		}
		return math.NaN()
	}
	start, windowEnd := mercuryWestElongationWindowContaining(jde, inst)
	for i := 0; i < eventDirectionalSearchIterations; i++ {
		date := mercuryGreatestElongationInWindow(start, windowEnd, inst)
		if !isFiniteFloat(date) {
			return math.NaN()
		}
		if eventUTQueryAfterOrEqual(date, jde) {
			return date
		}
		nextSuperior := NextMercurySuperiorConjunctionWithInstance(eventUTNextQueryTT(windowEnd), inst)
		start, windowEnd = mercuryWestElongationWindowEndingAt(nextSuperior, inst)
	}
	return math.NaN()
}

func lastMercuryGreatestElongationTyped(jde float64, east bool, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	if east {
		start, windowEnd := mercuryEastElongationWindowContaining(jde, inst)
		for i := 0; i < eventDirectionalSearchIterations; i++ {
			date := mercuryGreatestElongationInWindow(start, windowEnd, inst)
			if !isFiniteFloat(date) {
				return math.NaN()
			}
			if eventUTQueryBeforeOrEqual(date, jde) {
				return date
			}
			prevInferior := LastMercuryInferiorConjunctionWithInstance(eventUTLastQueryTT(start), inst)
			start, windowEnd = mercuryEastElongationWindowEndingAt(prevInferior, inst)
		}
		return math.NaN()
	}
	start, windowEnd := mercuryWestElongationWindowContaining(jde, inst)
	for i := 0; i < eventDirectionalSearchIterations; i++ {
		date := mercuryGreatestElongationInWindow(start, windowEnd, inst)
		if !isFiniteFloat(date) {
			return math.NaN()
		}
		if eventUTQueryBeforeOrEqual(date, jde) {
			return date
		}
		prevSuperior := LastMercurySuperiorConjunctionWithInstance(eventUTLastQueryTT(start), inst)
		start, windowEnd = mercuryWestElongationWindowEndingAt(prevSuperior, inst)
	}
	return math.NaN()
}

// mercuryElongationWindowAt 返回包含 jde 的该侧大距窗口；查询落在合的 4 秒内边距里时两侧都不包含。
func mercuryElongationWindowAt(jde float64, east bool, instances ...*Instance) (float64, float64, bool) {
	inst := instanceValue(instances)

	if east {
		start, end := mercuryEastElongationWindowEndingAt(NextMercuryInferiorConjunctionWithInstance(jde, inst), inst)
		return start, end, eventUTQueryBeforeOrEqual(start, jde) && eventUTQueryAfterOrEqual(end, jde)
	}
	start, end := mercuryWestElongationWindowEndingAt(NextMercurySuperiorConjunctionWithInstance(jde, inst), inst)
	return start, end, eventUTQueryBeforeOrEqual(start, jde) && eventUTQueryAfterOrEqual(end, jde)
}

// NextMercuryGreatestElongationWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextMercuryGreatestElongationWithInstance(jde float64, inst *Instance) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	// 先只看查询所在一侧的窗口，未过即答案，已过则取另一侧的下一个；赤经差在合附近会提前变号，不能用来定窗口。
	for _, east := range [2]bool{true, false} {
		start, end, ok := mercuryElongationWindowAt(jde, east, inst)
		if !ok {
			continue
		}
		if date := mercuryGreatestElongationInWindow(start, end, inst); isFiniteFloat(date) && eventUTQueryAfterOrEqual(date, jde) {
			return date
		}
		if date := nextMercuryGreatestElongationTyped(jde, !east, inst); isFiniteFloat(date) && eventUTQueryAfterOrEqual(date, jde) {
			return date
		}
		break
	}
	return earliestFiniteEventUT(nextMercuryGreatestElongationTyped(jde, true, inst), nextMercuryGreatestElongationTyped(jde, false, inst))
}

// LastMercuryGreatestElongationWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastMercuryGreatestElongationWithInstance(jde float64, inst *Instance) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	// 先只看查询所在一侧的窗口，未过即答案，已过则取另一侧的下一个；赤经差在合附近会提前变号，不能用来定窗口。
	for _, east := range [2]bool{true, false} {
		start, end, ok := mercuryElongationWindowAt(jde, east, inst)
		if !ok {
			continue
		}
		if date := mercuryGreatestElongationInWindow(start, end, inst); isFiniteFloat(date) && eventUTQueryBeforeOrEqual(date, jde) {
			return date
		}
		if date := lastMercuryGreatestElongationTyped(jde, !east, inst); isFiniteFloat(date) && eventUTQueryBeforeOrEqual(date, jde) {
			return date
		}
		break
	}
	return latestFiniteEventUT(lastMercuryGreatestElongationTyped(jde, true, inst), lastMercuryGreatestElongationTyped(jde, false, inst))
}

// LastMercuryInferiorConjunctionInclusiveWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastMercuryInferiorConjunctionInclusiveWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastSimpleEvent(jde, func(jde float64) float64 { return LastMercuryInferiorConjunctionWithInstance(jde, inst) }, func(jde float64) float64 { return NextMercuryInferiorConjunctionWithInstance(jde, inst) })
}

// NextMercuryInferiorConjunctionInclusiveWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextMercuryInferiorConjunctionInclusiveWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextSimpleEvent(jde, func(jde float64) float64 { return LastMercuryInferiorConjunctionWithInstance(jde, inst) }, func(jde float64) float64 { return NextMercuryInferiorConjunctionWithInstance(jde, inst) })
}

// LastMercurySuperiorConjunctionInclusiveWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastMercurySuperiorConjunctionInclusiveWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastSimpleEvent(jde, func(jde float64) float64 { return LastMercurySuperiorConjunctionWithInstance(jde, inst) }, func(jde float64) float64 { return NextMercurySuperiorConjunctionWithInstance(jde, inst) })
}

// NextMercurySuperiorConjunctionInclusiveWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextMercurySuperiorConjunctionInclusiveWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextSimpleEvent(jde, func(jde float64) float64 { return LastMercurySuperiorConjunctionWithInstance(jde, inst) }, func(jde float64) float64 { return NextMercurySuperiorConjunctionWithInstance(jde, inst) })
}

// LastMercuryRetrogradeInclusiveWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastMercuryRetrogradeInclusiveWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastSimpleEvent(jde, func(jde float64) float64 { return LastMercuryRetrogradeWithInstance(jde, inst) }, func(jde float64) float64 { return NextMercuryRetrogradeWithInstance(jde, inst) })
}

// NextMercuryRetrogradeInclusiveWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextMercuryRetrogradeInclusiveWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextSimpleEvent(jde, func(jde float64) float64 { return LastMercuryRetrogradeWithInstance(jde, inst) }, func(jde float64) float64 { return NextMercuryRetrogradeWithInstance(jde, inst) })
}

// LastMercuryProgradeToRetrogradeInclusiveWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastMercuryProgradeToRetrogradeInclusiveWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastSimpleEvent(jde, func(jde float64) float64 { return LastMercuryProgradeToRetrogradeWithInstance(jde, inst) }, func(jde float64) float64 { return NextMercuryProgradeToRetrogradeWithInstance(jde, inst) })
}

// NextMercuryProgradeToRetrogradeInclusiveWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextMercuryProgradeToRetrogradeInclusiveWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextSimpleEvent(jde, func(jde float64) float64 { return LastMercuryProgradeToRetrogradeWithInstance(jde, inst) }, func(jde float64) float64 { return NextMercuryProgradeToRetrogradeWithInstance(jde, inst) })
}

// LastMercuryRetrogradeToProgradeInclusiveWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastMercuryRetrogradeToProgradeInclusiveWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastSimpleEvent(jde, func(jde float64) float64 { return LastMercuryRetrogradeToProgradeWithInstance(jde, inst) }, func(jde float64) float64 { return NextMercuryRetrogradeToProgradeWithInstance(jde, inst) })
}

// NextMercuryRetrogradeToProgradeInclusiveWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextMercuryRetrogradeToProgradeInclusiveWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextSimpleEvent(jde, func(jde float64) float64 { return LastMercuryRetrogradeToProgradeWithInstance(jde, inst) }, func(jde float64) float64 { return NextMercuryRetrogradeToProgradeWithInstance(jde, inst) })
}

// LastMercuryGreatestElongationInclusiveWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastMercuryGreatestElongationInclusiveWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastSimpleEvent(jde, func(jde float64) float64 { return LastMercuryGreatestElongationWithInstance(jde, inst) }, func(jde float64) float64 { return NextMercuryGreatestElongationWithInstance(jde, inst) })
}

// NextMercuryGreatestElongationInclusiveWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextMercuryGreatestElongationInclusiveWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextSimpleEvent(jde, func(jde float64) float64 { return LastMercuryGreatestElongationWithInstance(jde, inst) }, func(jde float64) float64 { return NextMercuryGreatestElongationWithInstance(jde, inst) })
}

// LastMercuryGreatestElongationEastInclusiveWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastMercuryGreatestElongationEastInclusiveWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastSimpleEvent(jde, func(jde float64) float64 { return LastMercuryGreatestElongationEastWithInstance(jde, inst) }, func(jde float64) float64 { return NextMercuryGreatestElongationEastWithInstance(jde, inst) })
}

// NextMercuryGreatestElongationEastInclusiveWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextMercuryGreatestElongationEastInclusiveWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextSimpleEvent(jde, func(jde float64) float64 { return LastMercuryGreatestElongationEastWithInstance(jde, inst) }, func(jde float64) float64 { return NextMercuryGreatestElongationEastWithInstance(jde, inst) })
}

// LastMercuryGreatestElongationWestInclusiveWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastMercuryGreatestElongationWestInclusiveWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastSimpleEvent(jde, func(jde float64) float64 { return LastMercuryGreatestElongationWestWithInstance(jde, inst) }, func(jde float64) float64 { return NextMercuryGreatestElongationWestWithInstance(jde, inst) })
}

// NextMercuryGreatestElongationWestInclusiveWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextMercuryGreatestElongationWestInclusiveWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextSimpleEvent(jde, func(jde float64) float64 { return LastMercuryGreatestElongationWestWithInstance(jde, inst) }, func(jde float64) float64 { return NextMercuryGreatestElongationWestWithInstance(jde, inst) })
}

// NextMercuryGreatestElongationEastWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextMercuryGreatestElongationEastWithInstance(jde float64, inst *Instance) float64 {
	return nextMercuryGreatestElongationTyped(jde, true, inst)
}

// NextMercuryGreatestElongationWestWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextMercuryGreatestElongationWestWithInstance(jde float64, inst *Instance) float64 {
	return nextMercuryGreatestElongationTyped(jde, false, inst)
}

// LastMercuryGreatestElongationEastWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastMercuryGreatestElongationEastWithInstance(jde float64, inst *Instance) float64 {
	return lastMercuryGreatestElongationTyped(jde, true, inst)
}

// LastMercuryGreatestElongationWestWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastMercuryGreatestElongationWestWithInstance(jde float64, inst *Instance) float64 {
	return lastMercuryGreatestElongationTyped(jde, false, inst)
}
