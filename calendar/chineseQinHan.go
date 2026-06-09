package calendar

import (
	"math"
	"time"

	"github.com/starainrt/astro/basic"
)

const (
	qinHanMinSolarYear = -221
	qinHanMinYear      = -220
	qinHanMaxYear      = -104
	qinHanLunarMonth   = 29.0 + 499.0/940.0
)

var qinHanLeapCycle = []int{0, 0, 1, 0, 0, 1, 0, 0, 1, 0, 1, 0, 0, 1, 0, 0, 1, 0, 1}
var qinHanAccMonthCycle = []int{0, 12, 24, 37, 49, 61, 74, 86, 98, 111, 123, 136, 148, 160, 173, 185, 197, 210, 222}
var qinHanMonthNums = []int{10, 11, 12, 1, 2, 3, 4, 5, 6, 7, 8, 9, 9}
var ancientMonthNames = []string{"", "正", "二", "三", "四", "五", "六", "七", "八", "九", "十", "十一", "十二"}

type qinHanMonth struct {
	lunarYear int
	month     int
	day       int
	leap      bool
	startJDN  int
	endJDN    int
}

func innerSolarToLunarQinHan(date time.Time) (Time, bool) {
	date = date.In(getCst())
	month, ok := qinHanMonthBySolar(date.Year(), int(date.Month()), date.Day())
	if !ok {
		return Time{}, false
	}
	month.day = qinHanDateJDN(date.Year(), int(date.Month()), date.Day()) - month.startJDN + 1
	return qinHanTime(date, month), true
}

func innerSolarToLunarQinHanByYMD(year, month, day int) (Time, bool) {
	return innerSolarToLunarQinHan(time.Date(year, time.Month(month), day, 0, 0, 0, 0, getCst()))
}

func lunarToSolarQinHan(year, month, day int, leap bool) (Time, bool) {
	lmonth, ok := qinHanMonthByLunar(year, month, leap)
	if !ok {
		return Time{}, false
	}
	if day < 1 || day > lmonth.endJDN-lmonth.startJDN {
		return Time{}, false
	}
	lmonth.day = day
	date := qinHanJDNToDate(lmonth.startJDN + day - 1)
	return qinHanTime(date, lmonth), true
}

func rapidSolarQinHan(year, month, day int, leap bool) (time.Time, bool) {
	result, ok := lunarToSolarQinHan(year, month, day, leap)
	if !ok {
		return time.Time{}, false
	}
	return result.Solar(), true
}

func qinHanTime(date time.Time, month qinHanMonth) Time {
	return Time{
		solarTime: date,
		lunars: []LunarTime{
			{
				solarDate: date,
				year:      month.lunarYear,
				month:     month.month,
				day:       month.day,
				leap:      month.leap,
				desc:      formatQinHanLunarDateString(month.month, month.day, month.leap),
			},
		},
	}
}

func qinHanMonthBySolar(year, month, day int) (qinHanMonth, bool) {
	targetJDN := qinHanDateJDN(year, month, day)
	for lunarYear := qinHanMaxInt(qinHanMinYear, year-1); lunarYear <= qinHanMinInt(qinHanMaxYear, year+1); lunarYear++ {
		months := qinHanMonthsForYear(lunarYear)
		for _, m := range months {
			if targetJDN >= m.startJDN && targetJDN < m.endJDN {
				return m, true
			}
		}
	}
	return qinHanMonth{}, false
}

func qinHanMonthByLunar(year, month int, leap bool) (qinHanMonth, bool) {
	if year < qinHanMinYear || year > qinHanMaxYear {
		return qinHanMonth{}, false
	}
	for _, m := range qinHanMonthsForYear(year) {
		if m.month == month && m.leap == leap {
			return m, true
		}
	}
	return qinHanMonth{}, false
}

func qinHanMonthsForYear(year int) []qinHanMonth {
	starts := qinHanMonthStartJDNs(year)
	nextStarts := qinHanMonthStartJDNs(year + 1)
	months := make([]qinHanMonth, 0, len(starts))
	for i, start := range starts {
		end := nextStarts[0]
		if i+1 < len(starts) {
			end = starts[i+1]
		}
		leap := i == 12
		months = append(months, qinHanMonth{
			lunarYear: year,
			month:     qinHanMonthNums[i],
			leap:      leap,
			startJDN:  start,
			endJDN:    end,
		})
	}
	return months
}

func qinHanMonthStartJDNs(year int) []int {
	jdEpoch, accMonthEpoch, yearEpochLeap := qinHanEpoch(year)
	cycle := floorDiv(year-yearEpochLeap, 19)
	yearInCycle := year - yearEpochLeap - 19*cycle
	accMonths := accMonthEpoch + 235*cycle + qinHanAccMonthCycle[yearInCycle]
	monthCount := 12 + qinHanLeapCycle[yearInCycle]
	monthZero := jdEpoch + float64(accMonths)*qinHanLunarMonth
	starts := make([]int, monthCount)
	for i := 0; i < monthCount; i++ {
		base := monthZero
		// 高祖五年正月以后按汉初颛顼历新历元推算，前几个月仍沿用秦历续推。
		if year == -201 && i >= 3 {
			base = 1633701.5 + 470*qinHanLunarMonth
		}
		starts[i] = int(math.Floor(base + float64(i)*qinHanLunarMonth + 0.5 + 1e-9))
	}
	return starts
}

func qinHanEpoch(year int) (float64, int, int) {
	// 三段历元分别对应秦历、汉初改元后和太初改历前的颛顼历推算参数。
	if year >= -162 {
		return 1646163.5, 321, -179
	}
	if year > -201 {
		return 1633701.5, 174, -225
	}
	return 1589523.5, 1670, -225
}

func qinHanDateJDN(year, month, day int) int {
	return int(math.Floor(basic.JDECalc(year, month, float64(day)) + 0.5))
}

func qinHanJDNToDate(jdn int) time.Time {
	return basic.JDE2DateByZone(float64(jdn)-0.5, getCst(), true)
}

func floorDiv(a, b int) int {
	q := a / b
	r := a % b
	if r != 0 && ((r < 0) != (b < 0)) {
		q--
	}
	return q
}

func qinHanMinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func qinHanMaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func formatQinHanLunarDateString(lunarMonth, lunarDay int, isLeap bool) string {
	if isLeap {
		return "后九月" + formatLunarDayString(lunarDay)
	}
	return ancientMonthNames[lunarMonth] + "月" + formatLunarDayString(lunarDay)
}

func formatLunarDayString(lunarDay int) string {
	dayNames := []string{"十", "一", "二", "三", "四", "五", "六", "七", "八", "九", "十"}
	dayPrefixes := []string{"初", "十", "廿", "三"}
	if lunarDay == 20 {
		return "二十"
	}
	if lunarDay == 10 {
		return "初十"
	}
	return dayPrefixes[lunarDay/10] + dayNames[lunarDay%10]
}
