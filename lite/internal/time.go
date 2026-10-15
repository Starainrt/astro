package internal

import (
	"errors"
	"math"
	"time"
)

// ErrInvalidCivilDate 日期非法 / invalid civil date.
var ErrInvalidCivilDate = errors.New("invalid civil date")

// Date2JD 日期转儒略日 / Julian day from a civil time.
func Date2JD(date time.Time) float64 {
	day := float64(date.Day()) + float64(date.Hour())/24.0 + float64(date.Minute())/1440.0 + float64(date.Second())/86400.0 + float64(date.Nanosecond())/86400000000000.0
	return JDCalc(date.Year(), int(date.Month()), day)
}

// JDCalc 由公历年月日计算儒略日 / Julian day from a civil date.
func JDCalc(year, month int, day float64) float64 {
	if !validCivilDate(year, month, day) {
		return math.NaN()
	}
	effectiveYear, effectiveMonth, effectiveDay := year, month, int(math.Floor(day))
	if month == 1 || month == 2 {
		year--
		month += 12
	}
	correction := 0
	if effectiveYear > 1582 || (effectiveYear == 1582 && (effectiveMonth > 10 || (effectiveMonth == 10 && effectiveDay >= 15))) {
		century := int(year / 100)
		correction = 2 - century + int(century/4)
	}
	return math.Floor(365.25*(float64(year)+4716.0)) + math.Floor(30.6001*float64(month+1)) + day + float64(correction) - 1524.5
}

func validCivilDate(year, month int, day float64) bool {
	if math.IsNaN(day) || math.IsInf(day, 0) || month < 1 || month > 12 || day < 1 {
		return false
	}
	dayInt := int(math.Floor(day))
	if dayInt > daysInCivilMonth(year, month) {
		return false
	}
	return !(year == 1582 && month == 10 && day >= 5 && day < 15)
}

func daysInCivilMonth(year, month int) int {
	switch month {
	case 1, 3, 5, 7, 8, 10, 12:
		return 31
	case 4, 6, 9, 11:
		return 30
	case 2:
		if isCivilLeapYear(year, month, 1) {
			return 29
		}
		return 28
	default:
		return 0
	}
}

func isCivilLeapYear(year, month int, day float64) bool {
	if year < 1582 || (year == 1582 && (month < 10 || (month == 10 && day <= 4))) {
		return year%4 == 0
	}
	if year%400 == 0 {
		return true
	}
	if year%100 == 0 {
		return false
	}
	return year%4 == 0
}

// JD2DateByZone 儒略日转目标时区日期 / convert a Julian day to a target timezone.
func JD2DateByZone(jd float64, tz *time.Location, byZone bool) time.Time {
	jd += 0.5
	z := float64(int(jd))
	fraction := jd - z
	a := z
	if z >= 2299161.0 {
		alpha := math.Floor((z - 1867216.25) / 36524.25)
		a = z + 1 + alpha - math.Floor(alpha/4)
	}
	b := a + 1524
	c := math.Floor((b - 122.1) / 365.25)
	d := math.Floor(365.25 * c)
	e := math.Floor((b - d) / 30.6001)
	days := b - d - math.Floor(30.6001*e) + fraction
	month := e - 1
	if e >= 14 {
		month = e - 13
	}
	year := c - 4716
	if month <= 2 {
		year = c - 4715
	}
	seconds := (days - math.Floor(days)) * 86400
	day := math.Floor(days)
	date := time.Date(int(year), time.Month(int(month)), int(day), 0, 0, 0, 0, time.UTC).
		Add(time.Duration(int64(1000000000 * seconds)))
	if !byZone {
		return date.In(tz)
	}
	y, m, dayInt := date.Date()
	h, minute, second := date.Clock()
	return time.Date(y, m, dayInt, h, minute, second, date.Nanosecond(), tz)
}

// UTC2UT1 轻量民用时转世界时，忽略 DUT1（|UT1−UTC| < 0.9 s） / lightweight UTC to UT1, DUT1 ignored.
func UTC2UT1(jd float64) float64 {
	return jd
}
