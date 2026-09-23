package basic

import (
	"math"
	"time"
)

// 等时线（食甚/掩甚时刻等值线）的共享常量与工具。
// 两个等时线实现（日食、月掩）用同一套容差与时刻取值生成规则，避免口径漂移。

const (
	// greatestTimeContourCoverToleranceKM 是"该点是否已被已绘等时线覆盖"的距离门限。
	// 判据必须用点到折线的距离：同一条曲线被两个种子各画一次时该距离约为 0，而折线弦高在
	// 1.5° 步长下最大几 km，所以 10 km 既不会漏判重复，也远小于不同支路的间距。
	greatestTimeContourCoverToleranceKM = 10.0
	// greatestTimeContourCorrectionBacktracking 是牛顿投影跳出定义域时按二分回退的次数。
	greatestTimeContourCorrectionBacktracking = 6
	// greatestTimeContourKMPerDegree 是地面大圆每度的近似长度，用于局部平面投影。
	greatestTimeContourKMPerDegree = 111.195
	// greatestTimeContourMaxLevels 是单个事件允许的等时线条数上限；超出部分按时间截断，
	// 避免极长事件请求出上万条曲线。
	greatestTimeContourMaxLevels = 64
	// greatestTimeContourResidualTolerance 是残差零点判定的收敛容差。
	greatestTimeContourResidualTolerance = 1e-9
	// greatestTimeContourBisectionIterations 是二分求根的迭代上限。
	greatestTimeContourBisectionIterations = 64
)

// greatestTimeContourDifference 优先用中心差分，一侧越界时退化为单侧差分。
func greatestTimeContourDifference(center, positive, negative, step float64, positiveOK, negativeOK bool) (float64, bool) {
	switch {
	case positiveOK && negativeOK:
		return (positive - negative) / (2 * step), true
	case positiveOK:
		return (positive - center) / step, true
	case negativeOK:
		return (center - negative) / step, true
	}
	return 0, false
}

// greatestTimeContourBisect 在 [left,right] 上二分求残差零点；residual 在定义域外返回 NaN。
func greatestTimeContourBisect(
	residual func(longitude, latitude float64) float64,
	left, right, latitude, leftValue float64,
) (float64, bool) {
	for iteration := 0; iteration < greatestTimeContourBisectionIterations; iteration++ {
		middle := (left + right) / 2
		middleValue := residual(middle, latitude)
		if !finite(middleValue) {
			return 0, false
		}
		if math.Abs(middleValue) <= greatestTimeContourResidualTolerance || right-left <= 1e-9 {
			return middle, true
		}
		if leftValue*middleValue <= 0 {
			right = middle
		} else {
			left, leftValue = middle, middleValue
		}
	}
	return (left + right) / 2, true
}

// greatestTimeContourTickTolerance 是对齐网格在窗口边界上的容差，远小于任何合法步长。
const greatestTimeContourTickTolerance = time.Millisecond

// greatestTimeContourLevel 是一条等时线对应的时刻取值：TT 儒略日与它对应的时刻必须成对传递，
// 前者用于求根，后者用于标注（按步长生成时它是原始对齐时刻，避免 JDE 往返把整分截断成前一分钟）。
type greatestTimeContourLevel struct {
	tt float64
	at time.Time
}

// greatestTimeContourAlignedLevels 生成覆盖 [startTT,endTT] 且对齐到 step 整刻度的时刻取值，
// 最多 maxLevels 条；起点或终点恰好落在整刻度上时该刻度只生成一次。
func greatestTimeContourAlignedLevels(startTT, endTT float64, step time.Duration, maxLevels int) []greatestTimeContourLevel {
	if step <= 0 || maxLevels <= 0 || endTT <= startTT {
		return nil
	}
	start := greatestTimeContourTTToUTC(startTT)
	first := start.Truncate(step)
	// TT 儒略日往返有约 40 µs（1 ULP）误差，恰好落在整刻度上的窗口边界会算到刻度外一点点，容差按 1 ms 给。
	if first.Add(greatestTimeContourTickTolerance).Before(start) {
		first = first.Add(step)
	}
	end := greatestTimeContourTTToUTC(endTT)
	levels := make([]greatestTimeContourLevel, 0, maxLevels)
	for current := first; !current.Add(-greatestTimeContourTickTolerance).After(end) && len(levels) < maxLevels; current = current.Add(step) {
		levels = append(levels, greatestTimeContourLevel{tt: greatestTimeContourUTCToTT(current), at: current})
	}
	return levels
}

// greatestTimeContourTTToUTC 把 TT 儒略日换成对应的 UTC 时刻。
func greatestTimeContourTTToUTC(tt float64) time.Time {
	return JD2DateByZone(TT2UTC(tt), time.UTC, false)
}

// greatestTimeContourUTCToTT 把 UTC 时刻换成对应的 TT 儒略日。
func greatestTimeContourUTCToTT(value time.Time) float64 {
	return UTC2TT(Date2JD(value.UTC()))
}

// greatestTimeContourPointSegmentKM 返回点到折线段的距离，单位 km。
// 段长只有 1–2 度，用局部等距圆柱近似即可，误差远小于覆盖容差。
func greatestTimeContourPointSegmentKM(longitude, latitude, startLongitude, startLatitude, endLongitude, endLatitude float64) float64 {
	cosine := math.Cos(latitude * rad)
	toX := func(value float64) float64 {
		return normalizeLongitude(value-startLongitude) * rad * cosine * greatestTimeContourKMPerDegree
	}
	toY := func(value float64) float64 { return (value - startLatitude) * greatestTimeContourKMPerDegree }
	pointX, pointY := toX(longitude), toY(latitude)
	endX, endY := toX(endLongitude), toY(endLatitude)
	lengthSquared := endX*endX + endY*endY
	projection := 0.0
	if lengthSquared > 0 {
		projection = (pointX*endX + pointY*endY) / lengthSquared
		if projection < 0 {
			projection = 0
		} else if projection > 1 {
			projection = 1
		}
	}
	return math.Hypot(pointX-projection*endX, pointY-projection*endY)
}
