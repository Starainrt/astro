package moon

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// ApsisInfo 轨道极值事件 / orbital distance extremum event.
type ApsisInfo struct {
	// Time 事件时刻，UTC / event time in UTC.
	Time time.Time
	// Distance 极值距离，单位 km / extremum distance in km.
	Distance float64
}

// PerigeesInMonth 指定年月内的所有月球近地点 / all lunar perigees in the given Gregorian month.
func PerigeesInMonth(year int, month time.Month) []ApsisInfo {
	return convertMoonApsisInfos(basic.MoonPerigees(year, month))
}

// ApogeesInMonth 指定年月内的所有月球远地点 / all lunar apogees in the given Gregorian month.
func ApogeesInMonth(year int, month time.Month) []ApsisInfo {
	return convertMoonApsisInfos(basic.MoonApogees(year, month))
}

func convertMoonApsisInfos(events []basic.ApsisEvent) []ApsisInfo {
	result := make([]ApsisInfo, 0, len(events))
	for _, event := range events {
		result = append(result, ApsisInfo{
			Time:     basic.JDE2DateByZone(event.JDE, time.UTC, false),
			Distance: event.Distance,
		})
	}
	return result
}
