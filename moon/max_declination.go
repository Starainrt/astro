package moon

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// MaximumDeclinationInfo 月球最大赤纬事件 / maximum lunar declination event.
type MaximumDeclinationInfo struct {
	// Time 事件时刻，UTC / event time in UTC.
	Time time.Time
	// Declination 该时刻的地心赤纬，单位度 / geocentric declination at the event, in degrees.
	Declination float64
}

// MaximumNorthDeclinationsInMonth 指定年月内的所有月球最大北赤纬事件 / all maximum northern lunar declination events in the given Gregorian month.
func (module *Moon) MaximumNorthDeclinationsInMonth(year int, month time.Month) []MaximumDeclinationInfo {
	return convertMaximumDeclinationInfos(basic.MoonMaximumNorthDeclinationsWithInstance(year, month, module.inst))
}

// MaximumSouthDeclinationsInMonth 指定年月内的所有月球最大南赤纬事件 / all maximum southern lunar declination events in the given Gregorian month.
func (module *Moon) MaximumSouthDeclinationsInMonth(year int, month time.Month) []MaximumDeclinationInfo {
	return convertMaximumDeclinationInfos(basic.MoonMaximumSouthDeclinationsWithInstance(year, month, module.inst))
}

// LastMaximumNorthDeclination 指定时刻之前最近一次月球最大北赤纬 / last maximum northern lunar declination at or before date.
func (module *Moon) LastMaximumNorthDeclination(date time.Time) MaximumDeclinationInfo {
	return convertMaximumDeclinationInfo(date, basic.LastMoonMaximumNorthDeclinationWithInstance(timeToUTJDE(date), module.inst))
}

// NextMaximumNorthDeclination 指定时刻之后最近一次月球最大北赤纬 / next maximum northern lunar declination after date.
func (module *Moon) NextMaximumNorthDeclination(date time.Time) MaximumDeclinationInfo {
	return convertMaximumDeclinationInfo(date, basic.NextMoonMaximumNorthDeclinationWithInstance(timeToUTJDE(date), module.inst))
}

// ClosestMaximumNorthDeclination 离指定时刻最近一次月球最大北赤纬 / closest maximum northern lunar declination to date.
func (module *Moon) ClosestMaximumNorthDeclination(date time.Time) MaximumDeclinationInfo {
	return convertMaximumDeclinationInfo(date, basic.ClosestMoonMaximumNorthDeclinationWithInstance(timeToUTJDE(date), module.inst))
}

// LastMaximumSouthDeclination 指定时刻之前最近一次月球最大南赤纬 / last maximum southern lunar declination at or before date.
func (module *Moon) LastMaximumSouthDeclination(date time.Time) MaximumDeclinationInfo {
	return convertMaximumDeclinationInfo(date, basic.LastMoonMaximumSouthDeclinationWithInstance(timeToUTJDE(date), module.inst))
}

// NextMaximumSouthDeclination 指定时刻之后最近一次月球最大南赤纬 / next maximum southern lunar declination after date.
func (module *Moon) NextMaximumSouthDeclination(date time.Time) MaximumDeclinationInfo {
	return convertMaximumDeclinationInfo(date, basic.NextMoonMaximumSouthDeclinationWithInstance(timeToUTJDE(date), module.inst))
}

// ClosestMaximumSouthDeclination 离指定时刻最近一次月球最大南赤纬 / closest maximum southern lunar declination to date.
func (module *Moon) ClosestMaximumSouthDeclination(date time.Time) MaximumDeclinationInfo {
	return convertMaximumDeclinationInfo(date, basic.ClosestMoonMaximumSouthDeclinationWithInstance(timeToUTJDE(date), module.inst))
}

func convertMaximumDeclinationInfos(events []basic.DeclinationEvent) []MaximumDeclinationInfo {
	result := make([]MaximumDeclinationInfo, 0, len(events))
	for _, event := range events {
		result = append(result, convertMaximumDeclinationInfo(time.Time{}, event))
	}
	return result
}

func convertMaximumDeclinationInfo(date time.Time, event basic.DeclinationEvent) MaximumDeclinationInfo {
	location := time.UTC
	if date.Location() != nil {
		location = date.Location()
	}
	return MaximumDeclinationInfo{
		Time:        basic.JD2DateByZone(event.JD, location, false),
		Declination: event.Declination,
	}
}

func timeToUTJDE(date time.Time) float64 {
	return basic.Date2JD(date.UTC())
}
