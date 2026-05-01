package earth

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// ApsisInfo 轨道极值事件 / orbital distance extremum event.
type ApsisInfo struct {
	// Time 事件时刻，UTC / event time in UTC.
	Time time.Time
	// Distance 极值距离，单位 AU / extremum distance in AU.
	Distance float64
}

// Perihelion 指定年份的地球近日点 / Earth perihelion in the given year.
func Perihelion(year int) ApsisInfo {
	return convertEarthApsisInfo(basic.EarthPerihelion(year))
}

// Aphelion 指定年份的地球远日点 / Earth aphelion in the given year.
func Aphelion(year int) ApsisInfo {
	return convertEarthApsisInfo(basic.EarthAphelion(year))
}

func convertEarthApsisInfo(event basic.ApsisEvent) ApsisInfo {
	return ApsisInfo{
		Time:     basic.JDE2DateByZone(event.JDE, time.UTC, false),
		Distance: event.Distance,
	}
}
