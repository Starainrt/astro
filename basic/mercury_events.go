package basic

import (
	"math"

	"github.com/starainrt/astro/planet"
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

func mercuryHelioN(planetIndex int, jde float64, n int) mercuryConjunctionLBR {
	return mercuryConjunctionLBR{
		lo: planet.WherePlanetN(planetIndex, 0, jde, n),
		bo: planet.WherePlanetN(planetIndex, 1, jde, n),
		r:  planet.WherePlanetN(planetIndex, 2, jde, n),
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

func mercuryConjunctionHeliocentricDelta(jde, targetDeg float64, n int) float64 {
	planetLo := planet.WherePlanetN(1, 0, jde, n)
	earthLo := planet.WherePlanetN(-1, 0, jde, n)
	return mercuryConjunctionAngleDelta(planetLo - earthLo - targetDeg)
}

func mercuryConjunctionDifference(jd float64, n int, targetDeg, sunLightDays, geoLightDays float64) mercuryConjunctionResult {
	earthForSun := mercuryHelioN(-1, jd-sunLightDays, n)
	sunLo := Limit360(earthForSun.lo + 180)
	earth := mercuryHelioN(-1, jd-geoLightDays, n)
	planetPos := mercuryHelioN(1, jd-geoLightDays, n)
	geo := mercuryGeocentric(planetPos, earth)
	return mercuryConjunctionResult{
		diff:         mercuryConjunctionAngleDelta(geo.lo - sunLo - targetDeg),
		sunLightDays: earthForSun.r * mercuryLightTimeDaysPerAU,
		geoLightDays: geo.dist * mercuryLightTimeDaysPerAU,
	}
}

func mercuryConjunctionExactDelta(jde float64) float64 {
	return mercuryConjunctionAngleDelta(MercuryApparentLo(jde) - HSunApparentLo(jde))
}

func mercuryConjunctionApproxTT(seed float64, inferior bool) float64 {
	heliocentricTarget := 180.0
	if inferior {
		heliocentricTarget = 0
	}
	jde := seed
	for i := 0; i < 6; i++ {
		jde -= mercuryConjunctionHeliocentricDelta(jde, heliocentricTarget, 8) / (360.0 / MERCURY_S_PERIOD)
	}

	startSample := mercuryConjunctionDifference(jde, 8, 0, 0, 0)
	nextSample := mercuryConjunctionDifference(jde+mercuryConjunctionDerivativeStepDay, 8, 0, 0, 0)
	diffSlope := mercuryConjunctionAngleDelta(nextSample.diff-startSample.diff) / mercuryConjunctionDerivativeStepDay

	refined := mercuryConjunctionDifference(jde, 40, 0, startSample.sunLightDays, startSample.geoLightDays)
	jde -= refined.diff / diffSlope
	final := mercuryConjunctionDifference(jde, -1, 0, refined.sunLightDays, refined.geoLightDays)
	jde -= final.diff / diffSlope
	return jde
}

func mercuryConjunctionExactTT(seed float64, inferior bool) float64 {
	estimateJD := mercuryConjunctionApproxTT(seed, inferior)
	converged := false
	for i := 0; i < eventNewtonMaxIterations; i++ {
		prevJDE := estimateJD
		longitudeDelta := mercuryConjunctionExactDelta(prevJDE)
		longitudeSlope := (mercuryConjunctionExactDelta(prevJDE+0.000005) - mercuryConjunctionExactDelta(prevJDE-0.000005)) / 0.00001
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
func mercuryConjunction(jde float64, next uint8) float64 {
	//0=last 1=next
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	if math.Abs(mercuryConjunctionExactDelta(jde)) <= mercuryConjunctionSameInstantDegrees {
		best := math.NaN()
		consider := func(inferior bool) {
			eventUT := TT2UTC(mercuryConjunctionExactTT(jde, inferior))
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
	leftValue := mercuryConjunctionDeltaN(leftJDE, mercuryEventSearchN)
	if !isFiniteFloat(leftValue) {
		return math.NaN()
	}
	for i := 0; i < mercuryConjunctionScanMaxSteps; i++ {
		rightJDE := jde + direction*mercuryConjunctionScanStepDay*float64(i+1)
		rightValue := mercuryConjunctionDeltaN(rightJDE, mercuryEventSearchN)
		if !isFiniteFloat(rightValue) {
			return math.NaN()
		}
		if leftValue == 0 || rightValue == 0 || leftValue*rightValue < 0 {
			return mercuryConjunctionPolish(jde, leftJDE, rightJDE, direction)
		}
		leftJDE, leftValue = rightJDE, rightValue
	}
	return math.NaN()
}

// mercuryConjunctionDeltaN 截断级数下的水星-太阳视黄经差（度，[-180,180]），用于方向性括号扫描。
func mercuryConjunctionDeltaN(jde float64, n int) float64 {
	return mercuryConjunctionAngleDelta(MercuryApparentLoN(jde, n) - HSunApparentLoN(jde, n))
}

// mercuryConjunctionPolish 用全项级数在截断级数给出的括号内抛光。
// 截断误差可能让括号两端在全项函数上同号（罕见），此时沿扫描方向再扩一两个扫描步；
// 若仍未被确认（典型情形：查询几乎正好落在合上，截断级数在根两侧的符号与全项不一致），
// 退回全项级数的方向扫描，保证有界且不返回 NaN。
func mercuryConjunctionPolish(jde, leftJDE, rightJDE, direction float64) float64 {
	for attempt := 0; attempt < 3; attempt++ {
		leftValue := mercuryConjunctionExactDelta(leftJDE)
		rightValue := mercuryConjunctionExactDelta(rightJDE)
		if !isFiniteFloat(leftValue) || !isFiniteFloat(rightValue) {
			return math.NaN()
		}
		if leftValue == 0 || rightValue == 0 || leftValue*rightValue < 0 {
			root, ok := eventBracketSecantRoot(leftJDE, rightJDE, leftValue, rightValue,
				mercuryConjunctionPolishToleranceDay, mercuryConjunctionExactDelta)
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
	return mercuryConjunctionFullDirectionalScan(jde, direction)
}

// mercuryConjunctionFullDirectionalScan 全项级数的方向扫描（截断括号未被确认时的兜底）。
func mercuryConjunctionFullDirectionalScan(jde, direction float64) float64 {
	leftJDE := jde
	leftValue := mercuryConjunctionExactDelta(leftJDE)
	if !isFiniteFloat(leftValue) {
		return math.NaN()
	}
	for i := 0; i < mercuryConjunctionScanMaxSteps; i++ {
		rightJDE := jde + direction*mercuryConjunctionScanStepDay*float64(i+1)
		rightValue := mercuryConjunctionExactDelta(rightJDE)
		if !isFiniteFloat(rightValue) {
			return math.NaN()
		}
		if leftValue == 0 || rightValue == 0 || leftValue*rightValue < 0 {
			root, ok := eventBracketSecantRoot(leftJDE, rightJDE, leftValue, rightValue,
				mercuryConjunctionPolishToleranceDay, mercuryConjunctionExactDelta)
			if !ok {
				return math.NaN()
			}
			return TT2UTC(root)
		}
		leftJDE, leftValue = rightJDE, rightValue
	}
	return math.NaN()
}

func LastMercuryConjunction(jde float64) float64 {
	return inclusiveLastSimpleEvent(jde, LastMercuryConjunctionStrict, NextMercuryConjunctionStrict)
}

func NextMercuryConjunction(jde float64) float64 {
	return inclusiveNextSimpleEvent(jde, LastMercuryConjunctionStrict, NextMercuryConjunctionStrict)
}

func LastMercuryConjunctionStrict(jde float64) float64 {
	return mercuryConjunction(jde, 0)
}

func NextMercuryConjunctionStrict(jde float64) float64 {
	return mercuryConjunction(jde, 1)
}

func NextMercuryInferiorConjunction(jde float64) float64 {
	date := NextMercuryConjunctionStrict(jde)
	if EarthMercuryAway(date) > EarthAway(date) {
		return NextMercuryConjunctionStrict(date + 2)
	}
	return date
}

func NextMercurySuperiorConjunction(jde float64) float64 {
	date := NextMercuryConjunctionStrict(jde)
	if EarthMercuryAway(date) < EarthAway(date) {
		return NextMercuryConjunctionStrict(date + 2)
	}
	return date
}

func LastMercuryInferiorConjunction(jde float64) float64 {
	date := LastMercuryConjunctionStrict(jde)
	if EarthMercuryAway(date) > EarthAway(date) {
		return LastMercuryConjunctionStrict(date - 2)
	}
	return date
}

func LastMercurySuperiorConjunction(jde float64) float64 {
	date := LastMercuryConjunctionStrict(jde)
	if EarthMercuryAway(date) < EarthAway(date) {
		return LastMercuryConjunctionStrict(date - 2)
	}
	return date
}

func mercuryRADerivative(jde, delta float64) float64 {
	sub := MercuryApparentRa(jde+delta) - MercuryApparentRa(jde-delta)
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub / (2 * delta)
}

func mercuryRADerivativeN(jde, delta float64, n int) float64 {
	sub := MercuryApparentRaN(jde+delta, n) - MercuryApparentRaN(jde-delta, n)
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub / (2 * delta)
}

func mercuryStationInWindow(startTT, endTT float64) float64 {
	bestJDE := zeroEventInWindow(startTT, endTT, mercuryStationCoarseStepDay, mercuryStationHalfWindowDay, 30.0/86400.0, func(jd float64) float64 {
		return mercuryRADerivativeN(jd, mercuryStationDerivativeStepDay, mercuryEventSearchN)
	}, func(jd float64) float64 {
		return mercuryRADerivative(jd, mercuryStationDerivativeStepDay)
	})
	return TT2UTC(bestJDE)
}

func mercuryStationBetween(startTT, endTT float64) bool {
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
	leftValue := mercuryRADerivativeN(left, mercuryStationDerivativeStepDay, mercuryEventSearchN)
	for left < endTT {
		right := left + mercuryStationCoarseStepDay
		if right > endTT {
			right = endTT
		}
		rightValue := mercuryRADerivativeN(right, mercuryStationDerivativeStepDay, mercuryEventSearchN)
		if leftValue == 0 || leftValue*rightValue < 0 || rightValue == 0 {
			return true
		}
		left = right
		leftValue = rightValue
	}
	return false
}

func mercuryProgradeToRetrogradeAroundInferior(inferiorUT float64) float64 {
	inferiorTT := UTC2TT(inferiorUT)
	return mercuryStationInWindow(inferiorTT-mercuryStationWindowDays, inferiorTT)
}

func mercuryRetrogradeToProgradeAroundInferior(inferiorUT float64) float64 {
	inferiorTT := UTC2TT(inferiorUT)
	return mercuryStationInWindow(inferiorTT, inferiorTT+mercuryStationWindowDays)
}

// 类型化「留」统一结构：以下合为锚点求候选，再用侧向不变量（Next 不得早于查询、
// Last 不得晚于查询）校验；不满足则把锚点推进/回退一个会合周期重试，最多
// mercuryStationAnchorAttempts 次，最终仍有界失败为 NaN。
// 这保证 Last* 永远不会返回未来、Next* 永远不会返回过去的事件。
func NextMercuryProgradeToRetrograde(jde float64) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	inferior := NextMercuryInferiorConjunction(jde)
	for i := 0; i < mercuryStationAnchorAttempts && isFiniteFloat(inferior); i++ {
		date := mercuryProgradeToRetrogradeAroundInferior(inferior)
		if isFiniteFloat(date) && stationUTQueryAfterOrEqual(date, jde) {
			return date
		}
		inferior = NextMercuryInferiorConjunction(eventUTNextQueryTT(inferior))
	}
	return math.NaN()
}

func NextMercuryRetrogradeToPrograde(jde float64) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	inferior := LastMercuryInferiorConjunction(jde)
	for i := 0; i < mercuryStationAnchorAttempts && isFiniteFloat(inferior); i++ {
		date := mercuryRetrogradeToProgradeAroundInferior(inferior)
		if isFiniteFloat(date) && stationUTQueryAfterOrEqual(date, jde) {
			return date
		}
		inferior = NextMercuryInferiorConjunction(eventUTNextQueryTT(inferior))
	}
	return math.NaN()
}

func LastMercuryProgradeToRetrograde(jde float64) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	inferior := NextMercuryInferiorConjunction(jde)
	for i := 0; i < mercuryStationAnchorAttempts && isFiniteFloat(inferior); i++ {
		date := mercuryProgradeToRetrogradeAroundInferior(inferior)
		if isFiniteFloat(date) && stationUTQueryBeforeOrEqual(date, jde) {
			return date
		}
		inferior = LastMercuryInferiorConjunction(eventUTLastQueryTT(inferior))
	}
	return math.NaN()
}

func LastMercuryRetrogradeToPrograde(jde float64) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	inferior := LastMercuryInferiorConjunction(jde)
	for i := 0; i < mercuryStationAnchorAttempts && isFiniteFloat(inferior); i++ {
		date := mercuryRetrogradeToProgradeAroundInferior(inferior)
		if isFiniteFloat(date) && stationUTQueryBeforeOrEqual(date, jde) {
			return date
		}
		inferior = LastMercuryInferiorConjunction(eventUTLastQueryTT(inferior))
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

func nextMercuryRetrogradeFromTyped(jde float64) float64 {
	return earliestFiniteEventUT(NextMercuryProgradeToRetrograde(jde), NextMercuryRetrogradeToPrograde(jde))
}

func NextMercuryRetrograde(jde float64) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	motion := mercuryRADerivative(jde, mercuryStationDerivativeStepDay)
	if motion > mercuryStationMotionTolerance {
		p2r := NextMercuryProgradeToRetrograde(jde)
		if isFiniteFloat(p2r) && !mercuryStationBetween(jde, UTC2TT(p2r)) {
			return p2r
		}
		best := earliestFiniteEventUT(p2r, NextMercuryRetrogradeToPrograde(jde))
		if !stationUTQueryAfterOrEqual(best, jde) {
			return math.NaN()
		}
		return best
	}
	if motion < -mercuryStationMotionTolerance {
		r2p := NextMercuryRetrogradeToPrograde(jde)
		if isFiniteFloat(r2p) && !mercuryStationBetween(jde, UTC2TT(r2p)) {
			return r2p
		}
		best := earliestFiniteEventUT(NextMercuryProgradeToRetrograde(jde), r2p)
		if !stationUTQueryAfterOrEqual(best, jde) {
			return math.NaN()
		}
		return best
	}
	return nextMercuryRetrogradeFromTyped(jde)
}

func lastMercuryRetrogradeFromTyped(jde float64) float64 {
	return latestFiniteEventUT(LastMercuryProgradeToRetrograde(jde), LastMercuryRetrogradeToPrograde(jde))
}

func LastMercuryRetrograde(jde float64) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	motion := mercuryRADerivative(jde, mercuryStationDerivativeStepDay)
	if motion > mercuryStationMotionTolerance {
		r2p := LastMercuryRetrogradeToPrograde(jde)
		if isFiniteFloat(r2p) && !mercuryStationBetween(UTC2TT(r2p), jde) {
			return r2p
		}
		best := latestFiniteEventUT(LastMercuryProgradeToRetrograde(jde), r2p)
		if !stationUTQueryBeforeOrEqual(best, jde) {
			return math.NaN()
		}
		return best
	}
	if motion < -mercuryStationMotionTolerance {
		p2r := LastMercuryProgradeToRetrograde(jde)
		if isFiniteFloat(p2r) && !mercuryStationBetween(UTC2TT(p2r), jde) {
			return p2r
		}
		best := latestFiniteEventUT(p2r, LastMercuryRetrogradeToPrograde(jde))
		if !stationUTQueryBeforeOrEqual(best, jde) {
			return math.NaN()
		}
		return best
	}
	return lastMercuryRetrogradeFromTyped(jde)
}

func LastMercuryRetrogradeStrict(jde float64) float64 {
	return LastMercuryRetrograde(jde)
}

func NextMercuryRetrogradeStrict(jde float64) float64 {
	return NextMercuryRetrograde(jde)
}

func MercurySunElongation(jde float64) float64 {
	lo1, bo1 := MercuryApparentLoBo(jde)
	lo2 := HSunApparentLo(jde)
	bo2 := HSunTrueBo(jde)
	return StarAngularSeparation(lo1, bo1, lo2, bo2)
}

func mercurySunElongationN(jde float64, n int) float64 {
	lo1, bo1 := MercuryApparentLoBoN(jde, n)
	lo2 := HSunApparentLoN(jde, n)
	bo2 := HSunTrueBoN(jde, n)
	return StarAngularSeparation(lo1, bo1, lo2, bo2)
}

// mercuryGreatestElongationInWindow 求窗口内大距：目标是公开 MercurySunElongation 的极大（视距角），
// 而不是忽略光行差/视位置修正的真距角，否则返回的时刻不是调用方能量到的那个极值。
// 窗口两端是世界时，目标函数收力学时，因此逐次换算。
func mercuryGreatestElongationInWindow(start, end float64) float64 {
	return maximizeInWindow(start, end, 2.0, func(utJD float64) float64 {
		return mercurySunElongationN(UTC2TT(utJD), mercuryEventSearchN)
	}, func(utJD float64) float64 {
		return MercurySunElongation(UTC2TT(utJD))
	})
}

func mercuryEastElongationWindowEndingAt(inferior float64) (float64, float64) {
	lastSuperior := LastMercurySuperiorConjunction(eventUTLastQueryTT(inferior))
	return lastSuperior + innerEventWindowPadding, inferior - innerEventWindowPadding
}

func mercuryWestElongationWindowEndingAt(superior float64) (float64, float64) {
	lastInferior := LastMercuryInferiorConjunction(eventUTLastQueryTT(superior))
	return lastInferior + innerEventWindowPadding, superior - innerEventWindowPadding
}

func mercuryEastElongationWindowContaining(jde float64) (float64, float64) {
	nextInferior := NextMercuryInferiorConjunction(jde)
	start, end := mercuryEastElongationWindowEndingAt(nextInferior)
	if eventUTQueryBeforeOrEqual(start, jde) {
		return start, end
	}
	currentInferior := LastMercuryInferiorConjunction(jde)
	return mercuryEastElongationWindowEndingAt(currentInferior)
}

func mercuryWestElongationWindowContaining(jde float64) (float64, float64) {
	nextSuperior := NextMercurySuperiorConjunction(jde)
	start, end := mercuryWestElongationWindowEndingAt(nextSuperior)
	if eventUTQueryBeforeOrEqual(start, jde) {
		return start, end
	}
	currentSuperior := LastMercurySuperiorConjunction(jde)
	return mercuryWestElongationWindowEndingAt(currentSuperior)
}

func nextMercuryGreatestElongationTyped(jde float64, east bool) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	if east {
		start, windowEnd := mercuryEastElongationWindowContaining(jde)
		for i := 0; i < eventDirectionalSearchIterations; i++ {
			date := mercuryGreatestElongationInWindow(start, windowEnd)
			if !isFiniteFloat(date) {
				return math.NaN()
			}
			if eventUTQueryAfterOrEqual(date, jde) {
				return date
			}
			nextInferior := NextMercuryInferiorConjunction(eventUTNextQueryTT(windowEnd))
			start, windowEnd = mercuryEastElongationWindowEndingAt(nextInferior)
		}
		return math.NaN()
	}
	start, windowEnd := mercuryWestElongationWindowContaining(jde)
	for i := 0; i < eventDirectionalSearchIterations; i++ {
		date := mercuryGreatestElongationInWindow(start, windowEnd)
		if !isFiniteFloat(date) {
			return math.NaN()
		}
		if eventUTQueryAfterOrEqual(date, jde) {
			return date
		}
		nextSuperior := NextMercurySuperiorConjunction(eventUTNextQueryTT(windowEnd))
		start, windowEnd = mercuryWestElongationWindowEndingAt(nextSuperior)
	}
	return math.NaN()
}

func lastMercuryGreatestElongationTyped(jde float64, east bool) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	if east {
		start, windowEnd := mercuryEastElongationWindowContaining(jde)
		for i := 0; i < eventDirectionalSearchIterations; i++ {
			date := mercuryGreatestElongationInWindow(start, windowEnd)
			if !isFiniteFloat(date) {
				return math.NaN()
			}
			if eventUTQueryBeforeOrEqual(date, jde) {
				return date
			}
			prevInferior := LastMercuryInferiorConjunction(eventUTLastQueryTT(start))
			start, windowEnd = mercuryEastElongationWindowEndingAt(prevInferior)
		}
		return math.NaN()
	}
	start, windowEnd := mercuryWestElongationWindowContaining(jde)
	for i := 0; i < eventDirectionalSearchIterations; i++ {
		date := mercuryGreatestElongationInWindow(start, windowEnd)
		if !isFiniteFloat(date) {
			return math.NaN()
		}
		if eventUTQueryBeforeOrEqual(date, jde) {
			return date
		}
		prevSuperior := LastMercurySuperiorConjunction(eventUTLastQueryTT(start))
		start, windowEnd = mercuryWestElongationWindowEndingAt(prevSuperior)
	}
	return math.NaN()
}

// mercuryElongationWindowAt 返回包含 jde 的该侧大距窗口；查询落在合的 4 秒内边距里时两侧都不包含。
func mercuryElongationWindowAt(jde float64, east bool) (float64, float64, bool) {
	if east {
		start, end := mercuryEastElongationWindowEndingAt(NextMercuryInferiorConjunction(jde))
		return start, end, eventUTQueryBeforeOrEqual(start, jde) && eventUTQueryAfterOrEqual(end, jde)
	}
	start, end := mercuryWestElongationWindowEndingAt(NextMercurySuperiorConjunction(jde))
	return start, end, eventUTQueryBeforeOrEqual(start, jde) && eventUTQueryAfterOrEqual(end, jde)
}

// 无东西侧参数的 Next/Last 先只看查询所在窗口那一侧：该侧大距未过就是答案，已过则另一侧紧接着的
// 下一个才是答案，因此只有在查询贴住合（两侧窗口都不含）时才退化为两侧都算。赤经差在合附近会提前
// 变号，不能用它定窗口。
func NextMercuryGreatestElongation(jde float64) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	for _, east := range [2]bool{true, false} {
		start, end, ok := mercuryElongationWindowAt(jde, east)
		if !ok {
			continue
		}
		if date := mercuryGreatestElongationInWindow(start, end); isFiniteFloat(date) && eventUTQueryAfterOrEqual(date, jde) {
			return date
		}
		if date := nextMercuryGreatestElongationTyped(jde, !east); isFiniteFloat(date) && eventUTQueryAfterOrEqual(date, jde) {
			return date
		}
		break
	}
	return earliestFiniteEventUT(nextMercuryGreatestElongationTyped(jde, true), nextMercuryGreatestElongationTyped(jde, false))
}

func LastMercuryGreatestElongation(jde float64) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	for _, east := range [2]bool{true, false} {
		start, end, ok := mercuryElongationWindowAt(jde, east)
		if !ok {
			continue
		}
		if date := mercuryGreatestElongationInWindow(start, end); isFiniteFloat(date) && eventUTQueryBeforeOrEqual(date, jde) {
			return date
		}
		if date := lastMercuryGreatestElongationTyped(jde, !east); isFiniteFloat(date) && eventUTQueryBeforeOrEqual(date, jde) {
			return date
		}
		break
	}
	return latestFiniteEventUT(lastMercuryGreatestElongationTyped(jde, true), lastMercuryGreatestElongationTyped(jde, false))
}

func LastMercuryInferiorConjunctionInclusive(jde float64) float64 {
	return inclusiveLastSimpleEvent(jde, LastMercuryInferiorConjunction, NextMercuryInferiorConjunction)
}

func NextMercuryInferiorConjunctionInclusive(jde float64) float64 {
	return inclusiveNextSimpleEvent(jde, LastMercuryInferiorConjunction, NextMercuryInferiorConjunction)
}

func LastMercurySuperiorConjunctionInclusive(jde float64) float64 {
	return inclusiveLastSimpleEvent(jde, LastMercurySuperiorConjunction, NextMercurySuperiorConjunction)
}

func NextMercurySuperiorConjunctionInclusive(jde float64) float64 {
	return inclusiveNextSimpleEvent(jde, LastMercurySuperiorConjunction, NextMercurySuperiorConjunction)
}

func LastMercuryRetrogradeInclusive(jde float64) float64 {
	return inclusiveLastSimpleEvent(jde, LastMercuryRetrograde, NextMercuryRetrograde)
}

func NextMercuryRetrogradeInclusive(jde float64) float64 {
	return inclusiveNextSimpleEvent(jde, LastMercuryRetrograde, NextMercuryRetrograde)
}

func LastMercuryProgradeToRetrogradeInclusive(jde float64) float64 {
	return inclusiveLastSimpleEvent(jde, LastMercuryProgradeToRetrograde, NextMercuryProgradeToRetrograde)
}

func NextMercuryProgradeToRetrogradeInclusive(jde float64) float64 {
	return inclusiveNextSimpleEvent(jde, LastMercuryProgradeToRetrograde, NextMercuryProgradeToRetrograde)
}

func LastMercuryRetrogradeToProgradeInclusive(jde float64) float64 {
	return inclusiveLastSimpleEvent(jde, LastMercuryRetrogradeToPrograde, NextMercuryRetrogradeToPrograde)
}

func NextMercuryRetrogradeToProgradeInclusive(jde float64) float64 {
	return inclusiveNextSimpleEvent(jde, LastMercuryRetrogradeToPrograde, NextMercuryRetrogradeToPrograde)
}

func LastMercuryGreatestElongationInclusive(jde float64) float64 {
	return inclusiveLastSimpleEvent(jde, LastMercuryGreatestElongation, NextMercuryGreatestElongation)
}

func NextMercuryGreatestElongationInclusive(jde float64) float64 {
	return inclusiveNextSimpleEvent(jde, LastMercuryGreatestElongation, NextMercuryGreatestElongation)
}

func LastMercuryGreatestElongationEastInclusive(jde float64) float64 {
	return inclusiveLastSimpleEvent(jde, LastMercuryGreatestElongationEast, NextMercuryGreatestElongationEast)
}

func NextMercuryGreatestElongationEastInclusive(jde float64) float64 {
	return inclusiveNextSimpleEvent(jde, LastMercuryGreatestElongationEast, NextMercuryGreatestElongationEast)
}

func LastMercuryGreatestElongationWestInclusive(jde float64) float64 {
	return inclusiveLastSimpleEvent(jde, LastMercuryGreatestElongationWest, NextMercuryGreatestElongationWest)
}

func NextMercuryGreatestElongationWestInclusive(jde float64) float64 {
	return inclusiveNextSimpleEvent(jde, LastMercuryGreatestElongationWest, NextMercuryGreatestElongationWest)
}

func NextMercuryGreatestElongationEast(jde float64) float64 {
	return nextMercuryGreatestElongationTyped(jde, true)
}

func NextMercuryGreatestElongationWest(jde float64) float64 {
	return nextMercuryGreatestElongationTyped(jde, false)
}

func LastMercuryGreatestElongationEast(jde float64) float64 {
	return lastMercuryGreatestElongationTyped(jde, true)
}

func LastMercuryGreatestElongationWest(jde float64) float64 {
	return lastMercuryGreatestElongationTyped(jde, false)
}
