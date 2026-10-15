package basic

import (
	"time"
)

// TimeScale 导出时间标签的时标，零值为民用 UTC / output-label time scale, defaulting to civil UTC.
type TimeScale int

const (
	// TimeScaleUTC 默认民用时标 / default civil time scale.
	TimeScaleUTC TimeScale = iota
	// TimeScaleUT1 UT1 读数，输出 Location 须为 nil 或 UTC / UT1 reading; output Location must be nil or UTC.
	TimeScaleUT1
)

// LabelIn 把民用时刻换成该时标下的时刻值 / converts a civil instant to a value in one scale.
func LabelIn(scale TimeScale, date time.Time) time.Time {
	if scale == TimeScaleUT1 {
		return UT1FromUTC(date)
	}
	return date
}

// UT1FromUTC 民用时刻转 UT1 读数，以 UTC Location 承载 / converts a civil instant to a UT1 reading in a UTC Location.
func UT1FromUTC(date time.Time) time.Time {
	return JD2DateByZone(UTC2UT1(Date2JD(date.UTC())), time.UTC, false)
}

// DUT1 返回民用时刻的 UT1−UTC，单位秒 / UT1−UTC in seconds.
func DUT1(date time.Time) float64 {
	return DUT1Seconds(Date2JD(date.UTC()))
}
