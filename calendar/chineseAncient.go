package calendar

import (
	"fmt"
	"math"
	"time"

	"github.com/starainrt/astro/basic"
)

// AncientCalendarSystem 古六历系统 / ancient calendar system.
//
// 用于显式选择先秦古历或秦汉颛顼历。
// It identifies an explicitly selected pre-Qin or Qin/Early-Han calendar.
type AncientCalendarSystem string

const (
	AncientCalendarDefault AncientCalendarSystem = ""
	AncientCalendarChunqiu AncientCalendarSystem = "chunqiu"
	AncientCalendarZhou    AncientCalendarSystem = "zhou"
	AncientCalendarLu      AncientCalendarSystem = "lu"
	AncientCalendarHuangdi AncientCalendarSystem = "huangdi"
	AncientCalendarYin     AncientCalendarSystem = "yin"
	AncientCalendarXia1    AncientCalendarSystem = "xia1"
	AncientCalendarXia2    AncientCalendarSystem = "xia2"
	AncientCalendarZhuanxu AncientCalendarSystem = "zhuanxu"
	AncientCalendarQinHan  AncientCalendarSystem = "qin_han"
)

const (
	ancientMinYear         = -721
	ancientMaxYear         = -221
	ancientBoundaryMinYear = ancientMinYear - 1
	ancientBoundaryMaxYear = qinHanMinYear
	ancientLunarMonth      = 29.0 + 499.0/940.0
	ancientSolarYear       = 365.25
	chunqiuLunarMonth      = 30328.0 / 1027.0
	chunqiuYearEpoch       = -721
	chunqiuJDEpoch         = 1457727.761054236
	chunqiuLeapYearCount   = 244
	ancientDateEpsilon     = 1e-9
)

type ancientMonth struct {
	lunarYear int
	month     int
	day       int
	leap      bool
	startJDN  int
	endJDN    int
	system    AncientCalendarSystem
	name      string
}

type ancientSixParameters struct {
	yEpoch      int
	jdEpoch     float64
	jdEpochMoon float64
	ziOffset    int
	name        string
}

var chunqiuLeapYearBitmap = []byte{
	82, 73, 82, 164, 8, 155, 72, 201, 160, 138, 162, 144, 37, 73, 162, 73,
	145, 164, 81, 146, 34, 19, 163, 148, 168, 34, 67, 69, 37, 37, 1,
}

// SolarToLunarWithCalendar 公历转农历（显式古历） / solar to lunar calendar with an explicit ancient calendar.
//
// 传入公历日期和古历系统，返回该古历系统下的农历结果。
// Input is a civil date and an ancient calendar system. The result uses that explicit calendar.
func SolarToLunarWithCalendar(date time.Time, system AncientCalendarSystem) (Time, error) {
	if system == AncientCalendarDefault {
		return SolarToLunar(date)
	}
	date = date.In(getCst())
	return innerSolarToLunarByYMDWithCalendar(date.Year(), int(date.Month()), date.Day(), date, system)
}

// SolarToLunarByYMDWithCalendar 公历转农历（按年月日，显式古历） / solar to lunar calendar by YMD with an explicit ancient calendar.
//
// 传入公历年月日和古历系统，返回该古历系统下的农历结果。
// Inputs are civil year, month, day, and an ancient calendar system.
func SolarToLunarByYMDWithCalendar(year, month, day int, system AncientCalendarSystem) (Time, error) {
	if system == AncientCalendarDefault {
		return SolarToLunarByYMD(year, month, day)
	}
	return innerSolarToLunarByYMDWithCalendar(year, month, day, time.Time{}, system)
}

// LunarToSolarWithCalendar 农历描述转公历（显式古历） / lunar description to solar date with an explicit ancient calendar.
//
// 传入农历日期描述和古历系统，返回该古历系统下匹配的公历日期。
// Input is a lunar-date description and an ancient calendar system.
func LunarToSolarWithCalendar(desc string, system AncientCalendarSystem) ([]Time, error) {
	if system == AncientCalendarDefault {
		return LunarToSolar(desc)
	}
	date, err := parseChineseDate(desc)
	if err != nil {
		return nil, err
	}
	if date.year == 0 || date.comment != "" {
		if date.comment != "" {
			if result, known, err := lunarToSolarAncientEra(date, system); known {
				return result, err
			}
		}
		return nil, fmt.Errorf("显式古历暂不支持年号日期")
	}
	if date.houMonth && system != AncientCalendarQinHan && system != AncientCalendarZhuanxu {
		return nil, fmt.Errorf("未找到对应日期")
	}
	result, err := LunarToSolarByYMDWithCalendar(date.year, date.month, date.day, date.leap, system)
	if err != nil {
		return nil, err
	}
	return []Time{result}, nil
}

// LunarToSolarByYMDWithCalendar 农历转公历（按年月日，显式古历） / lunar to solar calendar by YMD with an explicit ancient calendar.
//
// 传入农历年月日、闰月标记和古历系统，返回该古历系统下匹配的公历日期。
// Inputs are lunar year, month, day, leap-month flag, and an ancient calendar system.
func LunarToSolarByYMDWithCalendar(year, month, day int, leap bool, system AncientCalendarSystem) (Time, error) {
	if system == AncientCalendarDefault {
		return LunarToSolarByYMD(year, month, day, leap)
	}
	if system == AncientCalendarQinHan {
		if result, ok := lunarToSolarQinHan(year, month, day, leap); ok {
			return tagCalendar(result, AncientCalendarQinHan, ancientCalendarName(AncientCalendarQinHan)), nil
		}
		return Time{}, fmt.Errorf("未找到对应日期")
	}
	lmonth, ok := ancientMonthByLunar(year, month, leap, system)
	if !ok {
		return Time{}, fmt.Errorf("未找到对应日期")
	}
	if day < 1 || day > lmonth.endJDN-lmonth.startJDN {
		return Time{}, fmt.Errorf("日期超出范围")
	}
	lmonth.day = day
	date := ancientJDNToDate(lmonth.startJDN + day - 1)
	if !ancientSolarYearInRange(date.Year()) {
		return Time{}, fmt.Errorf("未找到对应日期")
	}
	return ancientTime(date, lmonth), nil
}

func calendricalJieQiWithCalendar(year, term int, system AncientCalendarSystem) (time.Time, error) {
	if _, err := calendricalJieQiTermIndex(term); err != nil {
		return time.Time{}, err
	}
	if system == AncientCalendarDefault {
		return defaultCalendricalJieQi(year, term)
	}
	return calendricalJieQiBySystem(year, term, system)
}

func defaultCalendricalJieQi(year, term int) (time.Time, error) {
	if year < ancientMinYear || year > hanQingJieQiMaxYear {
		return time.Time{}, fmt.Errorf("该年份暂不支持历法节气")
	}
	if year < -479 {
		return time.Time{}, fmt.Errorf("历法 %s 暂不支持历法节气", AncientCalendarChunqiu)
	}
	if year < qinHanMinSolarYear {
		return calendricalJieQiBySystem(year, term, AncientCalendarZhou)
	}
	if year == qinHanMinSolarYear {
		qinHanDate, qinHanErr := calendricalJieQiBySystem(year, term, AncientCalendarQinHan)
		if qinHanErr == nil {
			return qinHanDate, nil
		}
		return calendricalJieQiBySystem(year, term, AncientCalendarZhou)
	}
	if year < qinHanMaxYear {
		return calendricalJieQiBySystem(year, term, AncientCalendarQinHan)
	}
	if year == qinHanMaxYear {
		qinHanDate, qinHanErr := calendricalJieQiBySystem(year, term, AncientCalendarQinHan)
		if qinHanErr == nil {
			return qinHanDate, nil
		}
		return hanQingCalendricalJieQiDate(year, term)
	}
	return hanQingCalendricalJieQiDate(year, term)
}

func calendricalJieQiBySystem(year, term int, system AncientCalendarSystem) (time.Time, error) {
	switch system {
	case AncientCalendarQinHan:
		if year < qinHanMinSolarYear || year > qinHanMaxYear {
			return time.Time{}, fmt.Errorf("历法 %s 不支持该年份", system)
		}
		date, err := ancientSixCalendricalJieQiDate(year, term, AncientCalendarZhuanxu)
		if err != nil {
			return time.Time{}, err
		}
		if !qinHanCalendricalDateSupported(date) {
			return time.Time{}, fmt.Errorf("历法 %s 不支持该年份", system)
		}
		return date, nil
	case AncientCalendarChunqiu:
		return time.Time{}, fmt.Errorf("历法 %s 暂不支持历法节气", system)
	case AncientCalendarZhou, AncientCalendarLu, AncientCalendarHuangdi, AncientCalendarYin, AncientCalendarXia1, AncientCalendarXia2, AncientCalendarZhuanxu:
		if !ancientSolarYearInRange(year) {
			return time.Time{}, fmt.Errorf("历法 %s 不支持该年份", system)
		}
		return ancientSixCalendricalJieQiDate(year, term, system)
	default:
		return time.Time{}, fmt.Errorf("不支持的古历系统: %s", system)
	}
}

func calendricalJieQiTermIndex(term int) (int, error) {
	if term < 0 || term >= 360 || term%15 != 0 {
		return 0, fmt.Errorf("节气参数超出范围")
	}
	return ((term - JQ_冬至 + 360) % 360) / 15, nil
}

func ancientSixCalendricalJieQiDate(year, term int, system AncientCalendarSystem) (time.Time, error) {
	termIndex, err := calendricalJieQiTermIndex(term)
	if err != nil {
		return time.Time{}, err
	}
	param, ok := ancientSixCalendarParameters(system)
	if !ok {
		return time.Time{}, fmt.Errorf("不支持的古历系统: %s", system)
	}
	dy := year - param.yEpoch - 1
	winterSolstice := param.jdEpoch + float64(dy)*ancientSolarYear
	if termIndex == 0 {
		winterSolstice += ancientSolarYear
	}
	jd := winterSolstice + float64(termIndex)*ancientSolarYear/24
	return calendricalJieQiDateFromJD(jd), nil
}

func calendricalJieQiDateFromJD(jd float64) time.Time {
	jdn := int(math.Floor(jd + 0.5 + ancientDateEpsilon))
	date := ancientJDNToDate(jdn)
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, getCst())
}

func qinHanStartDate() time.Time {
	return qinHanJDNToDate(qinHanMonthStartJDNs(qinHanMinYear)[0])
}

func qinHanEndDate() time.Time {
	months := qinHanMonthsForYear(qinHanMaxYear)
	if len(months) == 0 {
		return time.Time{}
	}
	return qinHanJDNToDate(months[len(months)-1].endJDN)
}

func qinHanCalendricalDateSupported(date time.Time) bool {
	if date.Before(qinHanStartDate()) {
		return false
	}
	end := qinHanEndDate()
	if end.IsZero() {
		return false
	}
	return date.Before(end)
}

func innerSolarToLunarAncientByYMD(year, month, day int, hmi time.Time) (Time, bool) {
	system, ok := defaultAncientCalendarSystemForYear(year)
	if !ok {
		return Time{}, false
	}
	result, err := innerSolarToLunarByYMDWithCalendar(year, month, day, hmi, system)
	if err != nil {
		return Time{}, false
	}
	return result, true
}

func lunarToSolarAncientDefault(year, month, day int, leap bool) (Time, bool) {
	system, ok := defaultAncientCalendarSystemForLunarYear(year)
	if !ok {
		return Time{}, false
	}
	result, err := LunarToSolarByYMDWithCalendar(year, month, day, leap, system)
	if err != nil {
		return Time{}, false
	}
	if !ancientSolarYearInRange(result.Solar().In(getCst()).Year()) {
		return Time{}, false
	}
	return result, true
}

func innerSolarToLunarByYMDWithCalendar(year, month, day int, hmi time.Time, system AncientCalendarSystem) (Time, error) {
	if system == AncientCalendarQinHan {
		if err := validateQinHanCalendarSolarInput(year, month, day); err != nil {
			return Time{}, err
		}
		if year > qinHanMaxYear {
			return Time{}, fmt.Errorf("历法 %s 不支持该年份", system)
		}
		if result, ok := innerSolarToLunarQinHanByYMD(year, month, day); ok {
			if !hmi.IsZero() {
				result.solarTime = hmi
				for i := range result.lunars {
					result.lunars[i].solarDate = hmi
				}
			}
			return tagCalendar(result, AncientCalendarQinHan, ancientCalendarName(AncientCalendarQinHan)), nil
		}
		return Time{}, fmt.Errorf("无法获取农历信息")
	}
	if !isPreQinSystem(system) {
		return Time{}, fmt.Errorf("不支持的古历系统: %s", system)
	}
	if err := validatePreQinCalendarSolarInput(year, month, day, system); err != nil {
		return Time{}, err
	}
	targetJDN := ancientDateJDN(year, month, day)
	for lunarYear := year - 1; lunarYear <= year+1; lunarYear++ {
		months, ok := ancientMonthsForYear(lunarYear, system)
		if !ok {
			continue
		}
		for _, m := range months {
			if targetJDN >= m.startJDN && targetJDN < m.endJDN {
				m.day = targetJDN - m.startJDN + 1
				date := hmi
				if date.IsZero() {
					date = time.Date(year, time.Month(month), day, 0, 0, 0, 0, getCst())
				}
				return ancientTime(date, m), nil
			}
		}
	}
	return Time{}, fmt.Errorf("无法获取农历信息")
}

func validatePreQinCalendarSolarInput(year, month, day int, system AncientCalendarSystem) error {
	if !ancientSolarYearInRange(year) {
		return fmt.Errorf("历法 %s 不支持该年份", system)
	}
	return validateAncientCivilDate(year, month, day)
}

func ancientSolarYearInRange(year int) bool {
	return year >= ancientMinYear && year <= qinHanMinSolarYear
}

func validateQinHanCalendarSolarInput(year, month, day int) error {
	if year < qinHanMinSolarYear || year > qinHanMaxYear {
		return fmt.Errorf("历法 %s 不支持该年份", AncientCalendarQinHan)
	}
	return validateAncientCivilDate(year, month, day)
}

func validateAncientCivilDate(year, month, day int) error {
	if month < 1 || month > 12 {
		return fmt.Errorf("月份超出范围")
	}
	if day < 1 || day > 31 {
		return fmt.Errorf("日期超出范围")
	}
	if err := basic.ValidateCivilDate(year, month, float64(day)); err != nil {
		return fmt.Errorf("公历日期不存在")
	}
	return nil
}

func defaultAncientCalendarSystemForYear(year int) (AncientCalendarSystem, bool) {
	if year < ancientMinYear || year > qinHanMinSolarYear {
		return AncientCalendarDefault, false
	}
	if year < -479 {
		return AncientCalendarChunqiu, true
	}
	return AncientCalendarZhou, true
}

func defaultAncientCalendarSystemForLunarYear(year int) (AncientCalendarSystem, bool) {
	if year < ancientBoundaryMinYear || year > ancientMaxYear {
		return AncientCalendarDefault, false
	}
	if year < -479 {
		return AncientCalendarChunqiu, true
	}
	return AncientCalendarZhou, true
}

func ancientMonthsForYear(year int, system AncientCalendarSystem) ([]ancientMonth, bool) {
	if !ancientSystemSupportsTableYear(year, system) {
		return nil, false
	}
	switch system {
	case AncientCalendarChunqiu:
		return chunqiuMonthsForYear(year)
	case AncientCalendarZhou, AncientCalendarLu, AncientCalendarHuangdi, AncientCalendarYin, AncientCalendarXia1, AncientCalendarXia2, AncientCalendarZhuanxu:
		return ancientSixMonthsForYear(year, system)
	default:
		return nil, false
	}
}

func ancientSystemSupportsTableYear(year int, system AncientCalendarSystem) bool {
	if year < ancientBoundaryMinYear {
		return false
	}
	if system == AncientCalendarChunqiu {
		return year <= -479
	}
	if !isPreQinSystem(system) {
		return false
	}
	return year <= ancientBoundaryMaxYear
}

func isPreQinSystem(system AncientCalendarSystem) bool {
	switch system {
	case AncientCalendarChunqiu, AncientCalendarZhou, AncientCalendarLu, AncientCalendarHuangdi, AncientCalendarYin, AncientCalendarXia1, AncientCalendarXia2, AncientCalendarZhuanxu:
		return true
	default:
		return false
	}
}

func chunqiuLeapYear(index int) int {
	if index < 0 || index >= chunqiuLeapYearCount {
		return 0
	}
	if chunqiuLeapYearBitmap[index/8]&(1<<uint(index%8)) != 0 {
		return 1
	}
	return 0
}

func chunqiuAccLeapsBefore(index int) int {
	if index <= 0 {
		return 0
	}
	if index > chunqiuLeapYearCount {
		index = chunqiuLeapYearCount
	}
	count := 0
	fullBytes := index / 8
	for i := 0; i < fullBytes; i++ {
		count += bitCount(chunqiuLeapYearBitmap[i])
	}
	for i := fullBytes * 8; i < index; i++ {
		count += chunqiuLeapYear(i)
	}
	return count
}

func bitCount(v byte) int {
	count := 0
	for v != 0 {
		v &= v - 1
		count++
	}
	return count
}

func chunqiuMonthsForYear(year int) ([]ancientMonth, bool) {
	i := year - chunqiuYearEpoch
	if i < -1 || i >= chunqiuLeapYearCount {
		return nil, false
	}
	leap := 0
	accLeaps := 0
	if i >= 0 {
		leap = chunqiuLeapYear(i)
		accLeaps = chunqiuAccLeapsBefore(i)
	}
	accMonths := 12*i + accLeaps
	monthCount := 12 + leap
	m0 := chunqiuJDEpoch + float64(accMonths)*chunqiuLunarMonth
	jd0 := ancientJDAtLocalMidnight(year-1, 12, 31)
	jdn0 := int(math.Floor(jd0 + 0.6))
	starts := make([]int, monthCount+1)
	for idx := 0; idx <= monthCount; idx++ {
		starts[idx] = jdn0 + int(math.Floor(m0+float64(idx)*chunqiuLunarMonth-jd0+ancientDateEpsilon))
	}
	months := make([]ancientMonth, 0, monthCount)
	for idx := 0; idx < monthCount; idx++ {
		month := idx + 1
		isLeap := false
		if monthCount == 13 && idx == 12 {
			month = 12
			isLeap = true
		}
		months = append(months, ancientMonth{
			lunarYear: year,
			month:     month,
			leap:      isLeap,
			startJDN:  starts[idx],
			endJDN:    starts[idx+1],
			system:    AncientCalendarChunqiu,
			name:      ancientCalendarName(AncientCalendarChunqiu),
		})
	}
	return months, true
}

func ancientSixMonthsForYear(year int, system AncientCalendarSystem) ([]ancientMonth, bool) {
	param, ok := ancientSixCalendarParameters(system)
	if !ok {
		return nil, false
	}
	dy := year - param.yEpoch - 1
	w0 := param.jdEpoch + float64(dy)*ancientSolarYear
	w1 := w0 + ancientSolarYear
	i := math.Floor((math.Floor(w0+1.5) - 0.5 - param.jdEpochMoon) / ancientLunarMonth)
	m0 := param.jdEpochMoon + i*ancientLunarMonth
	m1 := m0 + 13*ancientLunarMonth
	monthCount := 12
	if math.Floor(m1+0.5) < math.Floor(w1+0.5)+0.1 {
		monthCount = 13
	}
	monthOffset := param.ziOffset
	if param.ziOffset > 0 {
		if monthCount == 13 {
			monthOffset++
		}
		m1 = m0 + float64(monthCount+13)*ancientLunarMonth
		w2 := w1 + ancientSolarYear
		monthCount = 12
		if math.Floor(m1+0.5) < math.Floor(w2+0.5)+0.1 {
			monthCount = 13
		}
	}
	m0 += float64(monthOffset) * ancientLunarMonth
	jd0 := ancientJDAtLocalMidnight(year-1, 12, 31)
	jdn0 := int(math.Floor(jd0 + 0.6))
	months := make([]ancientMonth, 0, monthCount)
	for idx := 0; idx < monthCount; idx++ {
		m := m0 + float64(idx)*ancientLunarMonth
		startOffset := int(math.Floor(m - jd0 + ancientDateEpsilon))
		endOffset := int(math.Floor(m + ancientLunarMonth - jd0 + ancientDateEpsilon))
		start := jdn0 + startOffset
		end := jdn0 + endOffset
		month, isLeap := ancientSixMonthNumber(system, idx, monthCount)
		months = append(months, ancientMonth{
			lunarYear: year,
			month:     month,
			leap:      isLeap,
			startJDN:  start,
			endJDN:    end,
			system:    system,
			name:      param.name,
		})
	}
	return months, true
}

func ancientSixMonthNumber(system AncientCalendarSystem, index, monthCount int) (int, bool) {
	if monthCount == 13 && index == 12 {
		if system == AncientCalendarZhuanxu {
			return 9, true
		}
		return 12, true
	}
	if system == AncientCalendarZhuanxu {
		return 1 + ((index + 9) % 12), false
	}
	return index + 1, false
}

func ancientSixCalendarParameters(system AncientCalendarSystem) (ancientSixParameters, bool) {
	switch system {
	case AncientCalendarZhou:
		return ancientSixParameters{-104, 1683430.5001, 1683430.5001, 0, ancientCalendarName(system)}, true
	case AncientCalendarHuangdi:
		return ancientSixParameters{170, 1783510.5001, 1783510.5001, 0, ancientCalendarName(system)}, true
	case AncientCalendarYin:
		return ancientSixParameters{-47, 1704250.5001, 1704250.5001, 1, ancientCalendarName(system)}, true
	case AncientCalendarLu:
		jdEpoch := 1545730.5001
		return ancientSixParameters{-481, jdEpoch, jdEpoch - ancientLunarMonth/19.0, 0, ancientCalendarName(system)}, true
	case AncientCalendarZhuanxu:
		jdEpochMoon := 1726575.5001
		return ancientSixParameters{14, jdEpochMoon - ancientSolarYear/8.0, jdEpochMoon, -1, ancientCalendarName(system)}, true
	case AncientCalendarXia1:
		return ancientSixParameters{444, 1883590.5001, 1883590.5001, 2, ancientCalendarName(system)}, true
	case AncientCalendarXia2:
		jdEpochMoon := 1883650.5001
		return ancientSixParameters{444, jdEpochMoon - ancientSolarYear/6.0, jdEpochMoon, 2, ancientCalendarName(system)}, true
	default:
		return ancientSixParameters{}, false
	}
}

func ancientMonthByLunar(year, month int, leap bool, system AncientCalendarSystem) (ancientMonth, bool) {
	if !ancientSystemSupportsTableYear(year, system) {
		return ancientMonth{}, false
	}
	months, ok := ancientMonthsForYear(year, system)
	if !ok {
		return ancientMonth{}, false
	}
	for _, m := range months {
		if m.month == month && m.leap == leap {
			return m, true
		}
	}
	return ancientMonth{}, false
}

func ancientTime(date time.Time, month ancientMonth) Time {
	return Time{
		solarTime: date,
		lunars: []LunarTime{
			{
				solarDate:      date,
				year:           month.lunarYear,
				month:          month.month,
				day:            month.day,
				leap:           month.leap,
				desc:           formatAncientLunarDateString(month.month, month.day, month.leap, month.system),
				calendarSystem: month.system,
				calendarName:   month.name,
				eras:           ancientErasForLunarYear(month.lunarYear, month.system),
			},
		},
	}
}

func tagCalendar(date Time, system AncientCalendarSystem, name string) Time {
	for i := range date.lunars {
		date.lunars[i].calendarSystem = system
		date.lunars[i].calendarName = name
		if len(date.lunars[i].eras) == 0 {
			date.lunars[i].eras = ancientErasForLunarYear(date.lunars[i].year, system)
		}
	}
	return date
}

func formatAncientLunarDateString(month, day int, leap bool, system AncientCalendarSystem) string {
	if leap {
		if system == AncientCalendarZhuanxu {
			return "后九月" + formatLunarDayString(day)
		}
		return "闰" + formatAncientMonthName(month) + "月" + formatLunarDayString(day)
	}
	return formatAncientMonthName(month) + "月" + formatLunarDayString(day)
}

func formatAncientMonthName(month int) string {
	return ancientMonthNames[month]
}

func ancientCalendarName(system AncientCalendarSystem) string {
	switch system {
	case AncientCalendarChunqiu:
		return "春秋历"
	case AncientCalendarZhou:
		return "周历"
	case AncientCalendarLu:
		return "鲁历"
	case AncientCalendarHuangdi:
		return "黄帝历"
	case AncientCalendarYin:
		return "殷历"
	case AncientCalendarXia1:
		return "夏历(冬至版)"
	case AncientCalendarXia2:
		return "夏历(雨水版)"
	case AncientCalendarZhuanxu:
		return "颛顼历"
	case AncientCalendarQinHan:
		return "秦汉颛顼历"
	default:
		return ""
	}
}

func ancientDateJDN(year, month, day int) int {
	return int(math.Floor(basic.JDECalc(year, month, float64(day)) + 0.5))
}

func ancientJDAtLocalMidnight(year, month, day int) float64 {
	return basic.JDECalc(year, month, float64(day))
}

func ancientJDNToDate(jdn int) time.Time {
	return basic.JDE2DateByZone(float64(jdn)-0.5, getCst(), true)
}
