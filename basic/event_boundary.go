package basic

import "math"

const (
	exactEventTolerance     = 2.0 / 86400.0
	exactQueryTTToleranceUT = 0.1 / 86400.0
	// stationQueryToleranceUT 站（留）事件的「同刻」容差 / same-instant tolerance for station events.
	//
	// 站的求解本身有数值不确定度（火星 ~0.2 s、金星 ~2 s、外行星改精修步长后 <0.05 s）。
	// 若与普通事件一样取 0.1 s，查询落在站前几十毫秒时会被判成「站仍在未来」，
	// 于是跳到上一个会合周期。取 0.5 s 可覆盖除金星外的全部不确定度。
	// 上限约束：必须明显小于 1 s，否则会破坏
	// TestInnerPlanetNextEventAdvancesPastReturnedEvent 的「查询=事件+1 秒必须前进」语义。
	stationQueryToleranceUT = 0.5 / 86400.0
)

func sameEventJD(a, b float64) bool {
	return math.Abs(a-b) <= exactEventTolerance
}

// stationUTQueryBeforeOrEqual / stationUTQueryAfterOrEqual 站事件专用侧向判定，
// 容差比普通事件宽（见 stationQueryToleranceUT）。
func stationUTQueryBeforeOrEqual(eventUT, queryTT float64) bool {
	return eventUTQueryTTDelta(eventUT, queryTT) <= stationQueryToleranceUT
}

func stationUTQueryAfterOrEqual(eventUT, queryTT float64) bool {
	return eventUTQueryTTDelta(eventUT, queryTT) >= -stationQueryToleranceUT
}

func closestEventUTToQueryTT(queryTT, best float64, candidates ...float64) float64 {
	bestAbs := math.Abs(eventUTQueryTTDelta(best, queryTT))
	for _, candidate := range candidates {
		candidateAbs := math.Abs(eventUTQueryTTDelta(candidate, queryTT))
		if candidateAbs < bestAbs {
			best = candidate
			bestAbs = candidateAbs
		}
	}
	return best
}

type phaseEventSearchFunc func(jde, degree float64, next uint8) float64
type simpleEventSearchFunc func(jde float64) float64

// inclusive* 的唯一作用是把「查询几乎正好落在事件上」判成包含：
// 反向搜索给出的邻接事件落在同刻容差内时优先返回它，否则返回本方向的事件（可能是 NaN）。
func inclusiveLastPhaseEvent(jde, degree float64, fn phaseEventSearchFunc) float64 {
	last := fn(jde, degree, 0)
	if next := fn(jde, degree, 1); eventUTQueryBeforeOrEqual(next, jde) && eventUTQueryAfterOrEqual(next, jde) {
		return next
	}
	return last
}

func inclusiveNextPhaseEvent(jde, degree float64, fn phaseEventSearchFunc) float64 {
	if last := fn(jde, degree, 0); eventUTQueryBeforeOrEqual(last, jde) && eventUTQueryAfterOrEqual(last, jde) {
		return last
	}
	return fn(jde, degree, 1)
}

func inclusiveLastSimpleEvent(jde float64, lastFn, nextFn simpleEventSearchFunc) float64 {
	last := lastFn(jde)
	if next := nextFn(jde); eventUTQueryBeforeOrEqual(next, jde) && eventUTQueryAfterOrEqual(next, jde) {
		return next
	}
	return last
}

func inclusiveNextSimpleEvent(jde float64, lastFn, nextFn simpleEventSearchFunc) float64 {
	if last := lastFn(jde); eventUTQueryBeforeOrEqual(last, jde) && eventUTQueryAfterOrEqual(last, jde) {
		return last
	}
	return nextFn(jde)
}
