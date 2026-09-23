package basic

import (
	"errors"
	"math"
	"time"
)

var ErrInvalidCivilDate = errors.New("invalid civil date")
var timeNow = time.Now

// Date2JD 日期转儒略日
func Date2JD(date time.Time) float64 {
	day := float64(date.Day()) + float64(date.Hour())/24.0 + float64(date.Minute())/24.0/60.0 + float64(date.Second())/24.0/3600.0 + float64(date.Nanosecond())/1000000000.0/3600.0/24.0
	return JDCalc(date.Year(), int(date.Month()), day)
}

func ValidateCivilDate(year, month int, day float64) error {
	if math.IsNaN(day) || math.IsInf(day, 0) {
		return ErrInvalidCivilDate
	}
	if month < 1 || month > 12 {
		return ErrInvalidCivilDate
	}
	if day < 1 {
		return ErrInvalidCivilDate
	}
	dayInt := int(math.Floor(day))
	if dayInt < 1 || dayInt > daysInCivilMonth(year, month) {
		return ErrInvalidCivilDate
	}
	if isGregorianReformGap(year, month, day) {
		return ErrInvalidCivilDate
	}
	return nil
}

func isGregorianReformGap(year, month int, day float64) bool {
	return year == 1582 && month == 10 && day >= 5 && day < 15
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

// JDCalc 由公历年月日（day 可含小数）计算儒略日 / Julian day from a civil year, month and fractional day.
// 1582 年 10 月 15 日起按格里高利历，之前按儒略历；日期非法时返回 NaN。
// Gregorian from 1582-10-15 onward and Julian before it; an invalid civil date yields NaN.
func JDCalc(year, month int, day float64) float64 {
	if err := ValidateCivilDate(year, month, day); err != nil {
		return math.NaN()
	}
	effectiveYear, effectiveMonth, effectiveDay := year, month, int(math.Floor(day))
	if month == 1 || month == 2 {
		year--
		month += 12
	}
	var gregorianCorrection int
	if effectiveYear < 1582 || (effectiveYear == 1582 && (effectiveMonth < 10 || (effectiveMonth == 10 && effectiveDay <= 4))) {
		gregorianCorrection = 0
	} else {
		century := int(year / 100)
		gregorianCorrection = 2 - century + int(century/4)
	}
	return (math.Floor(365.25*(float64(year)+4716.0)) + math.Floor(30.6001*float64(month+1)) + day + float64(gregorianCorrection) - 1524.5)
}

// GetNowJD 按当前时区的日历字段取儒略日 / Julian day from the current clock's calendar fields.
// 读的是当前时区的年月日时分秒，不做时区归算。
// It reads the calendar fields in the current location without any zone conversion.
func GetNowJD() (nowJD float64) {
	now := timeNow()
	dayFraction := float64(now.Second())/3600.0/24.0 + float64(now.Minute())/60.0/24.0 + float64(now.Hour())/24.0
	nowJD = JDCalc(now.Year(), int(now.Month()), float64(now.Day())+dayFraction)
	return
}

// JD2Date 儒略日转 Local 时区的时刻 / Local-zone instant from a Julian day.
// 儒略日的日历字段按 Local 时区解释，1582 年 10 月 15 日之前按儒略历。
// The calendar fields are read in the Local zone, and dates before 1582-10-15 are read in the Julian calendar.
func JD2Date(jd float64) time.Time {
	jd = jd + 0.5
	z := float64(int(jd))
	f := jd - z
	var a, b, years, months, days float64
	if z < 2299161.0 {
		a = z
	} else {
		alpha := math.Floor((z - 1867216.25) / 36524.25)
		a = z + 1 + alpha - math.Floor(alpha/4)
	}
	b = a + 1524
	c := math.Floor((b - 122.1) / 365.25)
	d := math.Floor(365.25 * c)
	e := math.Floor((b - d) / 30.6001)
	days = b - d - math.Floor(30.6001*e) + f
	if e < 14 {
		months = e - 1
	}
	if e == 14 || e == 15 {
		months = e - 13
	}
	if months > 2 {
		years = c - 4716
	}
	if months == 1 || months == 2 {
		years = c - 4715
	}
	tms := (days - math.Floor(days)) * 24 * 3600
	days = math.Floor(days)
	tz, _ := time.LoadLocation("Local")
	dates := time.Date(int(years), time.Month(int(months)), int(days), 0, 0, 0, 0, tz)
	return time.Unix(dates.Unix()+int64(tms), int64((tms-math.Floor(tms))*1000000000))
}

// JD2DateByZone 儒略日转日期
// jd: 儒略日
// tz: 目标时区
// byZone: (true: 传入的儒略日视为目标时区当地时间的儒略日，false: 传入的儒略日视为UTC时间的儒略日)
// 回参：转换后的日期，时区始终为目标时区
func JD2DateByZone(jd float64, tz *time.Location, byZone bool) time.Time {
	jd = jd + 0.5
	z := float64(int(jd))
	f := jd - z
	var a, b, years, months, days float64
	if z < 2299161.0 {
		a = z
	} else {
		alpha := math.Floor((z - 1867216.25) / 36524.25)
		a = z + 1 + alpha - math.Floor(alpha/4)
	}
	b = a + 1524
	c := math.Floor((b - 122.1) / 365.25)
	d := math.Floor(365.25 * c)
	e := math.Floor((b - d) / 30.6001)
	days = b - d - math.Floor(30.6001*e) + f
	if e < 14 {
		months = e - 1
	}
	if e == 14 || e == 15 {
		months = e - 13
	}
	if months > 2 {
		years = c - 4716
	}
	if months == 1 || months == 2 {
		years = c - 4715
	}
	tms := (days - math.Floor(days)) * 24 * 3600
	days = math.Floor(days)
	date := time.Date(int(years), time.Month(int(months)), int(days), 0, 0, 0, 0, time.UTC).
		Add(time.Duration(int64(1000000000 * tms)))
	if !byZone {
		return date.In(tz)
	}
	// 当地 JD 的小数部分是钟表读数，不能按夏令时午夜后的实际时长累加。
	year, month, day := date.Date()
	hour, minute, second := date.Clock()
	return time.Date(year, month, day, hour, minute, second, date.Nanosecond(), tz)
}
