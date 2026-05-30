package formula

import "math"

// SynodicPeriod 会合周期 / synodic period.
//
//	period1: 第一个天体的恒星周期或朔望周期，单位自定，但必须与 period2 一致
//	period2: 第二个天体的周期，单位需与 period1 一致
//
// 返回：
//
//	会合周期，单位与输入相同
//
// Returns the synodic period in the same unit as the two input periods.
//
// 例：
//
//	// 地球与金星的会合周期，单位天
//	s := formula.SynodicPeriod(365.25636, 224.70069)
func SynodicPeriod(period1, period2 float64) float64 {
	if period1 <= 0 || period2 <= 0 ||
		math.IsNaN(period1) || math.IsInf(period1, 0) ||
		math.IsNaN(period2) || math.IsInf(period2, 0) {
		return math.NaN()
	}
	frequencyDiff := math.Abs(1/period1 - 1/period2)
	if frequencyDiff == 0 {
		return math.Inf(1)
	}
	return 1 / frequencyDiff
}
