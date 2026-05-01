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
func MaximumNorthDeclinationsInMonth(year int, month time.Month) []MaximumDeclinationInfo {
	return convertMaximumDeclinationInfos(basic.MoonMaximumNorthDeclinations(year, month))
}

// MaximumSouthDeclinationsInMonth 指定年月内的所有月球最大南赤纬事件 / all maximum southern lunar declination events in the given Gregorian month.
func MaximumSouthDeclinationsInMonth(year int, month time.Month) []MaximumDeclinationInfo {
	return convertMaximumDeclinationInfos(basic.MoonMaximumSouthDeclinations(year, month))
}

// LastMaximumNorthDeclination 指定时刻之前最近一次月球最大北赤纬 / last maximum northern lunar declination at or before date.
func LastMaximumNorthDeclination(date time.Time) MaximumDeclinationInfo {
	return convertMaximumDeclinationInfo(date, basic.LastMoonMaximumNorthDeclination(timeToUTJDE(date)))
}

// NextMaximumNorthDeclination 指定时刻之后最近一次月球最大北赤纬 / next maximum northern lunar declination after date.
func NextMaximumNorthDeclination(date time.Time) MaximumDeclinationInfo {
	return convertMaximumDeclinationInfo(date, basic.NextMoonMaximumNorthDeclination(timeToUTJDE(date)))
}

// ClosestMaximumNorthDeclination 离指定时刻最近一次月球最大北赤纬 / closest maximum northern lunar declination to date.
func ClosestMaximumNorthDeclination(date time.Time) MaximumDeclinationInfo {
	return convertMaximumDeclinationInfo(date, basic.ClosestMoonMaximumNorthDeclination(timeToUTJDE(date)))
}

// LastMaximumSouthDeclination 指定时刻之前最近一次月球最大南赤纬 / last maximum southern lunar declination at or before date.
func LastMaximumSouthDeclination(date time.Time) MaximumDeclinationInfo {
	return convertMaximumDeclinationInfo(date, basic.LastMoonMaximumSouthDeclination(timeToUTJDE(date)))
}

// NextMaximumSouthDeclination 指定时刻之后最近一次月球最大南赤纬 / next maximum southern lunar declination after date.
func NextMaximumSouthDeclination(date time.Time) MaximumDeclinationInfo {
	return convertMaximumDeclinationInfo(date, basic.NextMoonMaximumSouthDeclination(timeToUTJDE(date)))
}

// ClosestMaximumSouthDeclination 离指定时刻最近一次月球最大南赤纬 / closest maximum southern lunar declination to date.
func ClosestMaximumSouthDeclination(date time.Time) MaximumDeclinationInfo {
	return convertMaximumDeclinationInfo(date, basic.ClosestMoonMaximumSouthDeclination(timeToUTJDE(date)))
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
		Time:        basic.JDE2DateByZone(event.JDE, location, false),
		Declination: event.Declination,
	}
}

func timeToUTJDE(date time.Time) float64 {
	return basic.Date2JDE(date.UTC())
}
