package calendar

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// NowJDE 当前时刻儒略日 / current Julian day.
func NowJDE() float64 {
	return basic.GetNowJDE()
}

// Date2JDE 日期转儒略日 / date to Julian day.
func Date2JDE(date time.Time) float64 {
	day := float64(date.Day()) + float64(date.Hour())/24.0 + float64(date.Minute())/24.0/60.0 + float64(date.Second())/24.0/3600.0 + float64(date.Nanosecond())/1000000000.0/3600.0/24.0
	return basic.JDECalc(date.Year(), int(date.Month()), day)
}

// JDE2Date 儒略日转日期 / Julian day to date.
func JDE2Date(jde float64) time.Time {
	return basic.JDE2Date(jde)
}
