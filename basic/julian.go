package basic

import (
	"errors"
	"math"
	"time"
)

var ErrInvalidCivilDate = errors.New("invalid civil date")
var timeNow = time.Now

// Date2JDE 日期转儒略日
func Date2JDE(date time.Time) float64 {
	day := float64(date.Day()) + float64(date.Hour())/24.0 + float64(date.Minute())/24.0/60.0 + float64(date.Second())/24.0/3600.0 + float64(date.Nanosecond())/1000000000.0/3600.0/24.0
	return JDECalc(date.Year(), int(date.Month()), day)
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

/*
@name: 儒略日计算
@dec: 计算给定时间的儒略日，1582年改力后为格里高利历，之前为儒略历
@     请注意，传入的时间在天文计算中一般为力学时，应当注意和世界时的转化
*/
func JDECalc(year, month int, day float64) float64 {
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

/*
@name: 获得当前儒略日时间：当地世界时，非格林尼治时间
*/
func GetNowJDE() (nowJDE float64) {
	now := timeNow()
	dayFraction := float64(now.Second())/3600.0/24.0 + float64(now.Minute())/60.0/24.0 + float64(now.Hour())/24.0
	nowJDE = JDECalc(now.Year(), int(now.Month()), float64(now.Day())+dayFraction)
	return
}

func JDE2Date(jd float64) time.Time {
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

// JDE2DateByZone JDE（儒略日）转日期
// jd: 儒略日
// tz: 目标时区
// byZone: (true: 传入的儒略日视为目标时区当地时间的儒略日，false: 传入的儒略日视为UTC时间的儒略日)
// 回参：转换后的日期，时区始终为目标时区
func JDE2DateByZone(jd float64, tz *time.Location, byZone bool) time.Time {
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
	var transTz = tz
	if !byZone {
		transTz = time.UTC
	}
	return time.Date(int(years), time.Month(int(months)), int(days), 0, 0, 0, 0, transTz).
		Add(time.Duration(int64(1000000000 * tms))).In(tz)
}
