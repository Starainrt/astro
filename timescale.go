package astro

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// TimeScale 导出时间标签的时标，零值为民用 UTC / output-label time scale, defaulting to civil UTC.
type TimeScale = basic.TimeScale

const (
	// TimeScaleUTC 默认民用时标 / default civil time scale.
	TimeScaleUTC = basic.TimeScaleUTC
	// TimeScaleUT1 UT1 读数，输出 Location 须为 nil 或 UTC / UT1 reading; output Location must be nil or UTC.
	TimeScaleUT1 = basic.TimeScaleUT1
)

// LabelIn 把民用时刻换成该时标下的时刻值 / converts a civil instant to a value in one scale.
func LabelIn(scale TimeScale, date time.Time) time.Time {
	return basic.LabelIn(scale, date)
}

// UT1FromUTC 民用时刻转 UT1 读数，以 UTC Location 承载 / converts a civil instant to a UT1 reading in a UTC Location.
func UT1FromUTC(date time.Time) time.Time {
	return basic.UT1FromUTC(date)
}

// UTCFromUT1 是 UT1FromUTC 的逆 / inverts UT1FromUTC.
func UTCFromUT1(ut1 time.Time) time.Time {
	return utcLabel(basic.UT12UTC(basic.Date2JD(ut1.UTC())))
}

// TTFromUTC 民用时刻转 TT，按实例的 UTC 时刻取值 / converts a civil instant to TT.
func TTFromUTC(date time.Time) time.Time {
	return utcLabel(basic.UTC2TT(basic.Date2JD(date.UTC())))
}

// UTCFromTT 是 TTFromUTC 的逆 / inverts TTFromUTC.
func UTCFromTT(tt time.Time) time.Time {
	return utcLabel(basic.TT2UTC(basic.Date2JD(tt.UTC())))
}

// DUT1 返回民用时刻的 UT1−UTC，单位秒 / UT1−UTC in seconds.
func DUT1(date time.Time) float64 {
	return basic.DUT1(date)
}

// TCGFromTT 地球时转地心坐标时 / converts TT to TCG.
func TCGFromTT(tt time.Time) time.Time { return utcLabel(basic.TT2TCG(basic.Date2JD(tt.UTC()))) }

// TTFromTCG 是 TCGFromTT 的逆 / inverts TCGFromTT.
func TTFromTCG(tcg time.Time) time.Time { return utcLabel(basic.TCG2TT(basic.Date2JD(tcg.UTC()))) }

// TCBFromTT 地球时转太阳系质心坐标时，采用地心近似 / converts TT to TCB using a geocentric approximation.
func TCBFromTT(tt time.Time) time.Time { return utcLabel(basic.TT2TCB(basic.Date2JD(tt.UTC()))) }

// TTFromTCB 是 TCBFromTT 的逆 / inverts TCBFromTT.
func TTFromTCB(tcb time.Time) time.Time { return utcLabel(basic.TCB2TT(basic.Date2JD(tcb.UTC()))) }

// TDBFromTT 地球时转太阳系质心力学时，采用地心近似 / converts TT to TDB using a geocentric approximation.
func TDBFromTT(tt time.Time) time.Time { return utcLabel(basic.TT2TDB(basic.Date2JD(tt.UTC()))) }

// TTFromTDB 是 TDBFromTT 的逆 / inverts TDBFromTT.
func TTFromTDB(tdb time.Time) time.Time { return utcLabel(basic.TDB2TT(basic.Date2JD(tdb.UTC()))) }

// TCBFromTDB 质心力学时转质心坐标时 / converts TDB to TCB.
func TCBFromTDB(tdb time.Time) time.Time { return utcLabel(basic.TDB2TCB(basic.Date2JD(tdb.UTC()))) }

// TDBFromTCB 是 TCBFromTDB 的逆 / inverts TCBFromTDB.
func TDBFromTCB(tcb time.Time) time.Time { return utcLabel(basic.TCB2TDB(basic.Date2JD(tcb.UTC()))) }

// TCGMinusTT 返回 TT 时刻的 TCG−TT，单位秒 / TCG−TT in seconds for a TT instant.
func TCGMinusTT(tt time.Time) float64 { return basic.TCGMinusTTSeconds(basic.Date2JD(tt.UTC())) }

// TCBMinusTT 返回 TT 时刻的 TCB−TT，单位秒 / TCB−TT in seconds for a TT instant.
func TCBMinusTT(tt time.Time) float64 { return basic.TCBMinusTTSeconds(basic.Date2JD(tt.UTC())) }

// TDBMinusTT 返回 TT 时刻的地心 TDB−TT 近似值（秒）/ geocentric TDB−TT approximation in seconds at a TT instant.
func TDBMinusTT(tt time.Time) float64 { return basic.TDBMinusTTSeconds(basic.Date2JD(tt.UTC())) }

func utcLabel(jd float64) time.Time {
	return basic.JD2DateByZone(jd, time.UTC, false)
}
