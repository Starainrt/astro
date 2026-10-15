package calendar

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/starainrt/astro/basic"
)

var tiangan = []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
var dizhi = []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}

var cstLocation = time.FixedZone("CST", 8*3600)

func getCst() *time.Location {
	return cstLocation
}

const (
	JQ_春分 = 15 * iota
	JQ_清明
	JQ_谷雨
	JQ_立夏
	JQ_小满
	JQ_芒种
	JQ_夏至
	JQ_小暑
	JQ_大暑
	JQ_立秋
	JQ_处暑
	JQ_白露
	JQ_秋分
	JQ_寒露
	JQ_霜降
	JQ_立冬
	JQ_小雪
	JQ_大雪
	JQ_冬至
	JQ_小寒
	JQ_大寒
	JQ_立春
	JQ_雨水
	JQ_惊蛰
)

// Lunar 公历转农历 / solar to lunar calendar.
// 传入 公历年月日，时区
// 返回 农历月，日，是否闰月以及文字描述
// 按现行农历GB/T 33661-2017算法计算，推荐使用年限为[1929-3000]年
// 古代由于定朔定气误差此处计算会与古时不符
// Inputs are civil year, month, day, and timezone offset in hours.
// Returns the lunar year, month, day, leap-month flag, and text description.
// The current GB/T 33661-2017 lunar-calendar convention is recommended for years 1929-3000.
// For ancient dates, the result may differ from historical practice because computed new-moon and solar-term reconstructions are approximate for ancient dates.
func Lunar(year, month, day int, timezone float64) (int, int, int, bool, string) {
	return basic.GetLunar(year, month, day, timezone/24.0)
}

// Solar 农历转公历 / lunar to solar calendar.
// 传入 农历年份，月，日，是否闰月，时区
// 传出 公历时间
// 农历年份用公历年份代替，但是岁首需要使用农历岁首
// 例：计算己亥猪年腊月三十日对应的公历（即2020年1月24日）
// 由于农历还未到鼠年，故应当传入Solar(2019,12,30,false)
// 按现行农历GB/T 33661-2017算法计算，推荐使用年限为[1929-3000]年
// 古代由于定朔定气误差此处计算会与古时不符
// Inputs are the civil-year proxy of the lunar year, lunar month, lunar day, leap-month flag, and timezone offset in hours.
// Returns the corresponding civil time.
// The lunar year parameter follows the civil year containing the lunar New Year of that cycle.
// For example, the last day of the Ji-Hai year corresponds to 2020-01-24, but should still be passed as `Solar(2019, 12, 30, false, ...)`.
// The current GB/T 33661-2017 lunar-calendar convention is recommended for years 1929-3000.
// For ancient dates, the result may differ from historical practice because computed new-moon and solar-term reconstructions are approximate for ancient dates.
func Solar(year, month, day int, leap bool, timezone float64) time.Time {
	jd := basic.GetSolar(year, month, day, leap, timezone/24.0)
	zone := time.FixedZone("CST", int(timezone*3600))
	return basic.JD2DateByZone(jd, zone, true)
}

// SolarToLunar 公历转农历 / solar to lunar calendar.
// 传入 公历年月日
// 返回 包含农历信息的Time结构体
// 支持年份：[-721,3000]
// [-721,-221] 按默认先秦古历，[-220,-104] 秦汉颛顼历有效日期按复原算法，-104年交接后及[-103,1912]按照古代历法提供的农历信息
// (1912,3000]按现行农历GB/T 33661-2017算法计算
// Input is a civil `time.Time`.
// Returns a `Time` value carrying the lunar-calendar information.
// Supported civil years are [-721, 3000].
// Years [-721, -221] use the default pre-Qin ancient calendars.
// Years [-220, -104] use the reconstructed Qin and early-Han Zhuanxu calendar where that calendar has data.
// Late -104 and years [-103, 1912] use the historical-calendar tables included in this package.
// Years (1912, 3000] use the current GB/T 33661-2017 lunar-calendar convention.
func SolarToLunar(date time.Time) (Time, error) {
	return innerSolarToLunar(date)
}

// SolarToLunarByYMD 公历转农历（按年月日） / solar to lunar calendar by year, month, and day.
// 传入 公历年月日
// 返回 包含农历信息的Time结构体
// 支持年份：[-721,3000]
// [-721,-221] 按默认先秦古历，[-220,-104] 秦汉颛顼历有效日期按复原算法，-104年交接后及[-103,1912]按照古代历法提供的农历信息
// (1912,3000]按现行农历GB/T 33661-2017算法计算
// Inputs are the civil year, month, and day.
// Returns a `Time` value carrying the lunar-calendar information.
// Supported civil years are [-721, 3000].
// Years [-721, -221] use the default pre-Qin ancient calendars.
// Years [-220, -104] use the reconstructed Qin and early-Han Zhuanxu calendar where that calendar has data.
// Late -104 and years [-103, 1912] use the historical-calendar tables included in this package.
// Years (1912, 3000] use the current GB/T 33661-2017 lunar-calendar convention.
func SolarToLunarByYMD(year, month, day int) (Time, error) {
	return innerSolarToLunarByYMD(year, month, day)
}

func innerSolarToLunar(date time.Time) (Time, error) {
	date = date.In(getCst())
	if date.Year() < ancientMinYear || date.Year() > 9999 {
		return Time{}, fmt.Errorf("日期超出范围")
	}
	if err := basic.ValidateCivilDate(date.Year(), int(date.Month()), float64(date.Day())); err != nil {
		return Time{}, fmt.Errorf("公历日期不存在")
	}
	if date.Year() < qinHanMinSolarYear {
		if result, ok := innerSolarToLunarAncientByYMD(date.Year(), int(date.Month()), date.Day(), date); ok {
			return result, nil
		}
		return Time{}, fmt.Errorf("无法获取农历信息")
	}
	if date.Year() <= qinHanMaxYear {
		if result, ok := innerSolarToLunarQinHan(date); ok {
			return tagCalendar(result, AncientCalendarQinHan, ancientCalendarName(AncientCalendarQinHan)), nil
		}
		if date.Year() == qinHanMinSolarYear {
			if result, ok := innerSolarToLunarAncientByYMD(date.Year(), int(date.Month()), date.Day(), date); ok {
				return result, nil
			}
		}
		if date.Year() == qinHanMaxYear {
			return innerSolarToLunarHanQing(date), nil
		}
		return Time{}, fmt.Errorf("无法获取农历信息")
	}
	if date.Year() <= 1912 {
		return innerSolarToLunarHanQing(date), nil
	}
	if date.Year() < 2400 {
		y, m, d, l, desc := rapidLunarModern(date.Year(), int(date.Month()), date.Day())
		if desc == "无法获取农历信息" {
			return Time{}, fmt.Errorf("无法获取农历信息")
		}
		return transformModenLunar2Time(date, y, m, d, l, desc), nil
	}
	y, m, d, l, desc := basic.GetLunar(date.Year(), int(date.Month()), date.Day(), 8.0/24.0)
	return transformModenLunar2Time(date, y, m, d, l, desc), nil
}

func innerSolarToLunarByYMD(year, month, day int) (Time, error) {
	result, err := innerSolarToLunarByYMDRaw(year, month, day)
	if err != nil {
		return result, err
	}
	// 输入为儒略历独有闰日时补记 JulianOnly/JDE。
	if phantom, ok := julianOnlyPhantomJDE(year, month, day); ok {
		result = markJulianOnly(result, phantom)
	}
	return result, nil
}

func innerSolarToLunarByYMDRaw(year, month, day int) (Time, error) {
	if year < ancientMinYear || year > 9999 {
		return Time{}, fmt.Errorf("日期超出范围")
	}
	if month < 1 || month > 12 {
		return Time{}, fmt.Errorf("月份超出范围")
	}
	if day < 1 || day > 31 {
		return Time{}, fmt.Errorf("日期超出范围")
	}
	if err := basic.ValidateCivilDate(year, month, float64(day)); err != nil {
		return Time{}, fmt.Errorf("公历日期不存在")
	}
	if year < qinHanMinSolarYear {
		if result, ok := innerSolarToLunarAncientByYMD(year, month, day, time.Time{}); ok {
			return result, nil
		}
		return Time{}, fmt.Errorf("无法获取农历信息")
	}
	if year <= qinHanMaxYear {
		if result, ok := innerSolarToLunarQinHanByYMD(year, month, day); ok {
			return tagCalendar(result, AncientCalendarQinHan, ancientCalendarName(AncientCalendarQinHan)), nil
		}
		if year == qinHanMinSolarYear {
			if result, ok := innerSolarToLunarAncientByYMD(year, month, day, time.Time{}); ok {
				return result, nil
			}
		}
		if year == qinHanMaxYear {
			return innerSolarToLunarHanQingByYMD(year, month, day, time.Time{}), nil
		}
		return Time{}, fmt.Errorf("无法获取农历信息")
	}
	if year <= 1912 {
		return innerSolarToLunarHanQingByYMD(year, month, day, time.Time{}), nil
	}
	if year < 2400 {
		y, m, d, l, desc := rapidLunarModern(year, month, day)
		if desc == "无法获取农历信息" {
			return Time{}, fmt.Errorf("无法获取农历信息")
		}
		return transformModenLunar2Time(time.Date(year, time.Month(month), day, 0, 0, 0, 0, getCst()), y, m, d, l, desc), nil
	}
	y, m, d, l, desc := basic.GetLunar(year, month, day, 8.0/24.0)
	return transformModenLunar2Time(time.Date(year, time.Month(month), day, 0, 0, 0, 0, getCst()), y, m, d, l, desc), nil
}

func transformModenLunar2Time(date time.Time, year, month, day int, leap bool, desc string) Time {
	return Time{
		solarTime: date,
		lunars: []LunarTime{
			{
				solarDate:   date,
				year:        year,
				month:       month,
				day:         day,
				leap:        leap,
				desc:        desc,
				comment:     "",
				ganzhiMonth: commonGanZhiOfMonth(year, month),
				eras:        nil,
			},
		},
	}
}

// LunarToSolar 农历转公历 / lunar to solar calendar.
// 传入 农历描述，如"二零二零年正月初一","元丰六年十月十二","元嘉二十七年七月庚午日"
// 传出 包含公里农历信息的Time结构体切片
// 传入参数支持如下结构
// 农历年中文描述+农历月中文描述+农历日中文描述
// 农历年中文描述+农历月中文描述+干支日中文描述
// 年号+农历月中文描述+农历日中文描述
// 年号+农历月中文描述+干支日中文描述
// 支持年份：[-721,3000]
// Input is a lunar-date description such as `二零二零年正月初一`, `元丰六年十月十二`, or `元嘉二十七年七月庚午日`.
// Returns all matching `Time` results with both civil and lunar information.
// The parser accepts these forms:
// lunar year text + lunar month text + lunar day text
// lunar year text + lunar month text + sexagenary day text
// era name + lunar month text + lunar day text
// era name + lunar month text + sexagenary day text
// Supported civil result years are [-721, 3000]. Boundary lunar year -722 may be accepted when the result still falls inside that civil range.
func LunarToSolar(desc string) ([]Time, error) {
	dates, err := innerParseLunar(desc)
	if err != nil {
		return nil, err
	}
	var results []Time
	for _, v := range dates {
		date, err := SolarToLunar(v)
		if err != nil {
			return nil, err
		}
		results = append(results, date)
	}
	// 字符串入口先把描述解析成 time.Time，儒略历独有闰日会偏一天，这里按闰日复核一次。
	if parsed, perr := parseChineseDate(desc); perr == nil && parsed.LunarMonth() != 0 && parsed.LunarDay() != 0 {
		for i := range results {
			if fixed, ok := julianOnlyTextResult(results[i], parsed.LunarMonth(), parsed.LunarDay(), parsed.IsLeap()); ok {
				results[i] = fixed
			}
		}
	}
	return results, nil
}

// LunarToSolarSingle 农历转公历（单一结果） / lunar to solar calendar, single result.
//
// Deprecated: 推荐使用LunarToSolarByYMD
// 传入 农历年月日，是否闰月
// 传出 包含公里农历信息的Time结构体
// 支持年份：公历结果在[-721,3000]范围内，边界农历年可回溯到-722
// [-721,-221] 按默认先秦古历，[-220,-105] 按秦汉颛顼历复原算法，-104年重叠日期按默认公历交接选择，[-103,1912] 按照古代历法提供的农历信息,注意，这里农历月份代表的是以当时的历法推定的农历月与正月的距离，正月为1，二月为2，依次类推，闰月显示所闰月
// (1912,3000]按现行农历GB/T 33661-2017算法计算
// Deprecated: use LunarToSolarByYMD.
// Inputs are lunar year, month, day, and the leap-month flag.
// Returns a `Time` value carrying both civil and lunar information.
// Supported civil result years are [-721, 3000]. Boundary lunar year -722 may be accepted when the result still falls inside that civil range.
// Years [-721, -221] use the default pre-Qin ancient calendars.
// Years [-220, -105] use the reconstructed Qin and early-Han Zhuanxu calendar.
// Ambiguous -104 lunar dates follow the default civil handoff; use LunarToSolarByYMDWithCalendar for a specific ancient calendar.
// For years [-103, 1912], the lunar month index follows the historical calendar in force at that time, counted from the first month of that year.
// Years (1912, 3000] use the current GB/T 33661-2017 lunar-calendar convention.
func LunarToSolarSingle(year, month, day int, leap bool) (Time, error) {
	return LunarToSolarByYMD(year, month, day, leap)
}

// LunarToSolarByYMD 农历转公历（按年月日） / lunar to solar calendar by year, month, and day.
// 传入 农历年月日，是否闰月
// 传出 包含公里农历信息的Time结构体
// 支持年份：公历结果在[-721,3000]范围内，边界农历年可回溯到-722
// [-721,-221] 按默认先秦古历，[-220,-105] 按秦汉颛顼历复原算法，-104年重叠日期按默认公历交接选择，[-103,1912] 按照古代历法提供的农历信息,注意，这里农历月份代表的是以当时的历法推定的农历月与正月的距离，正月为1，二月为2，依次类推，闰月显示所闰月
// (1912,3000]按现行农历GB/T 33661-2017算法计算
// Inputs are lunar year, month, day, and the leap-month flag.
// Returns a `Time` value carrying both civil and lunar information.
// Supported civil result years are [-721, 3000]. Boundary lunar year -722 may be accepted when the result still falls inside that civil range.
// Years [-721, -221] use the default pre-Qin ancient calendars.
// Years [-220, -105] use the reconstructed Qin and early-Han Zhuanxu calendar.
// Ambiguous -104 lunar dates follow the default civil handoff; use LunarToSolarByYMDWithCalendar for a specific ancient calendar.
// For years [-103, 1912], the lunar month index follows the historical calendar in force at that time, counted from the first month of that year.
// Years (1912, 3000] use the current GB/T 33661-2017 lunar-calendar convention.
func LunarToSolarByYMD(year, month, day int, leap bool) (Time, error) {
	result, err := lunarToSolarByYMD(year, month, day, leap)
	if err != nil {
		return result, err
	}
	return finalizeLunarToSolarResult(result, year, month, day, leap), nil
}

// finalizeLunarToSolarResult 补记儒略历独有闰日与多公历候选。
func finalizeLunarToSolarResult(result Time, year, month, day int, leap bool) Time {
	if lunarDateMatches(result, year, month, day, leap) {
		result.solarCandidates = lunarToSolarCandidates(year, month, day, leap, result)
		return result
	}
	// 主答案对不上请求：儒略历独有的 2 月 29 日（正向序列在那里缺一号），或改历双纪年里
	// 主偏移落在了另一套纪年上（王莽 23/24 年交界）。
	if phantom, ok := julianOnlyPhantomJDE(result.Solar().Year(), 2, 29); ok {
		return markJulianOnlyLunar(result, year, month, day, leap, phantom)
	}
	alternates := lunarToSolarAlternates(year, month, day, leap)
	if len(alternates) == 0 {
		return result
	}
	primary, err := SolarToLunar(alternates[0])
	if err != nil {
		return result
	}
	if len(alternates) > 1 {
		primary.solarCandidates = alternates
	}
	return primary
}

func lunarToSolarByYMD(year, month, day int, leap bool) (Time, error) {
	if year < ancientBoundaryMinYear || year > 9999 {
		return Time{}, fmt.Errorf("年份超出范围")
	}
	if year < qinHanMinYear {
		if result, ok := lunarToSolarAncientDefault(year, month, day, leap); ok {
			return result, nil
		}
		return Time{}, fmt.Errorf("无法获取农历信息")
	}
	if year <= qinHanMaxYear {
		if year == qinHanMaxYear {
			if result, ok := lunarToSolarHanQingDefault(year, month, day, leap); ok {
				return result, nil
			}
		}
		if result, ok := lunarToSolarQinHan(year, month, day, leap); ok {
			return tagCalendar(result, AncientCalendarQinHan, ancientCalendarName(AncientCalendarQinHan)), nil
		}
		return Time{}, fmt.Errorf("无法获取农历信息")
	}
	if year <= 1912 {
		date := rapidSolarHan2Qing(year, month, day, leap, yearDiffLunar(year, month, day), nil)
		return SolarToLunar(date)
	}
	if year < 2400 {
		date := rapidSolarModern(year, month, day, leap)
		return SolarToLunar(date)
	}
	date := Solar(year, month, day, leap, 8.0)
	return SolarToLunar(date)
}

// julianOnlyCivilDay 报告该日期是否只存在于儒略历（ 1582 年前百年非 400 闰年的 2 月 29 日）。
func julianOnlyCivilDay(year, month, day int) bool {
	if month != 2 || day != 29 {
		return false
	}
	if basic.ValidateCivilDate(year, month, float64(day)) != nil {
		return false
	}
	return time.Date(year, time.February, 29, 0, 0, 0, 0, time.UTC).Month() != time.February
}

// julianOnlyPhantomJDE 返回儒略历独有闰日的儒略日。
func julianOnlyPhantomJDE(year, month, day int) (float64, bool) {
	if !julianOnlyCivilDay(year, month, day) {
		return 0, false
	}
	return basic.JDCalc(year, month, float64(day)), true
}

// julianOnlyTextResult 复核字符串入口的结果是否落在儒略历独有闰日上，是则返回该闰日的结果。
func julianOnlyTextResult(result Time, month, day int, leap bool) (Time, bool) {
	solar := result.Solar()
	year := solar.Year()
	if !julianOnlyCivilDay(year, 2, 29) {
		return result, false
	}
	diff := Date2JD(solar) - basic.JDCalc(year, 2, 29)
	switch diff {
	case 1, -1:
		// +1 表示结果是闰日的规范日，-1 表示结果是闰日的前一天。
	default:
		return result, false
	}
	forward, err := innerSolarToLunarByYMD(year, 2, 29)
	if err != nil {
		return result, false
	}
	lunar := forward.Lunar()
	// 只比日号与闰月标志（改历期间描述用年号月序、记录存规范月序），且与结果的农历日号相差一天。
	if lunar.LunarDay() != day || lunar.IsLeap() != leap {
		return result, false
	}
	resultDay := result.Lunar().LunarDay()
	if diff > 0 {
		// 结果是闰日的规范日，农历上比闰日晚一天。
		if resultDay != day+1 && !(resultDay == 1 && day >= 29) {
			return result, false
		}
	} else {
		// 结果是闰日的前一天，农历上比闰日早一天。
		if resultDay != day-1 && !(day == 1 && resultDay >= 29) {
			return result, false
		}
	}
	return forward, true
}

// markJulianOnly 补记儒略历独有闰日：Solar 取后继日作为规范日，精确日期放入 JDE。
func markJulianOnly(result Time, phantomJD float64) Time {
	canonicalDate := basic.JD2DateByZone(phantomJD, getCst(), true)
	result.solarTime = canonicalDate
	for i := range result.lunars {
		result.lunars[i].solarDate = canonicalDate
		result.lunars[i].julianOnly = true
		result.lunars[i].jd = phantomJD
	}
	return result
}

// markJulianOnlyLunar 反向路径用：用正向记录改写停在规范日的农历身份，并补记闰日标志。
func markJulianOnlyLunar(result Time, year, month, day int, leap bool, phantomJD float64) Time {
	canonicalDate := basic.JD2DateByZone(phantomJD, getCst(), true)
	result = markJulianOnly(result, phantomJD)
	forward, err := innerSolarToLunarByYMD(canonicalDate.Year(), 2, 29)
	if err != nil || !lunarDateMatches(forward, year, month, day, leap) {
		return result
	}
	lunars := make([]LunarTime, len(forward.lunars))
	copy(lunars, forward.lunars)
	for i := range lunars {
		lunars[i].julianOnly = true
		lunars[i].jd = phantomJD
	}
	result.lunars = lunars
	return result
}

// markJulianOnlyJDN 在 JDN 层面判定儒略历独有闰日（古历与秦汉反向直接由 JDN 造记录）。
func markJulianOnlyJDN(result Time, jdn int) Time {
	canonicalDate := basic.JD2DateByZone(float64(jdn)-0.5, getCst(), true)
	phantom, ok := julianOnlyPhantomJDE(canonicalDate.Year(), 2, 29)
	if !ok || int(math.Floor(phantom+0.5)) != jdn {
		return result
	}
	return markJulianOnly(result, phantom)
}

// lunarDateMatches 报告结果里是否有记录正是该农历日。
func lunarDateMatches(result Time, year, month, day int, leap bool) bool {
	for _, l := range result.Lunars() {
		if l.LunarYear() == year && l.LunarMonth() == month && l.LunarDay() == day && l.IsLeap() == leap {
			return true
		}
	}
	return false
}

// lunarReformYear 报告该年是否处于改历双纪年窗口（王莽/魏明帝/武则天/唐肃宗）。
func lunarReformYear(year int) bool {
	return (year >= 9 && year <= 23) || (year >= 237 && year <= 240) ||
		(year >= 689 && year <= 700) || (year >= 761 && year <= 762)
}

// lunarToSolarCandidates 收集该农历日的全部合法公历候选，主答案在前；单候选返回 nil。
func lunarToSolarCandidates(year, month, day int, leap bool, primary Time) []time.Time {
	candidates := []time.Time{primary.Solar()}
	for _, date := range lunarToSolarAlternates(year, month, day, leap) {
		duplicated := false
		for _, existing := range candidates {
			if existing.Equal(date) {
				duplicated = true
				break
			}
		}
		if !duplicated {
			candidates = append(candidates, date)
		}
	}
	if len(candidates) == 1 {
		return nil
	}
	return candidates
}

// lunarToSolarAlternates 收集另一套改历月序（或颛顼历）口径下能正向回到该农历日的公历日。
func lunarToSolarAlternates(year, month, day int, leap bool) []time.Time {
	var alternates []time.Time
	appendIfLegal := func(result Time, ok bool) {
		if !ok || !lunarDateMatches(result, year, month, day, leap) {
			return
		}
		date := result.Solar()
		for _, existing := range alternates {
			if existing.Equal(date) {
				return
			}
		}
		alternates = append(alternates, date)
	}
	switch {
	case year <= qinHanMaxYear:
		// 太初改历交接：颛顼历与默认（太初）口径各是一个候选。
		result, ok := lunarToSolarQinHan(year, month, day, leap)
		if ok {
			appendIfLegal(tagCalendar(result, AncientCalendarQinHan, ancientCalendarName(AncientCalendarQinHan)), true)
		}
	case year <= 1912:
		// 改历双纪年：换另一套月序偏移（0/1/2，255 在 rapidSolarHan2Qing 内部等价于 2）再算一次。
		if !lunarReformYear(year) {
			break
		}
		primaryDiff := yearDiffLunar(year, month, day)
		if primaryDiff == 255 {
			primaryDiff = 2
		}
		for _, diff := range []int{0, 1, 2} {
			if diff == primaryDiff {
				continue
			}
			date := rapidSolarHan2Qing(year, month, day, leap, diff, nil)
			if date.IsZero() {
				continue
			}
			result, err := SolarToLunar(date)
			appendIfLegal(result, err == nil)
		}
	}
	return alternates
}

func lunarToSolarHanQingDefault(year, month, day int, leap bool) (Time, bool) {
	date := rapidSolarHan2Qing(year, month, day, leap, yearDiffLunar(year, month, day), nil)
	if date.IsZero() {
		return Time{}, false
	}
	result, err := SolarToLunar(date)
	if err != nil {
		return Time{}, false
	}
	lunar := result.Lunar()
	if lunar.CalendarSystem() == AncientCalendarQinHan {
		return Time{}, false
	}
	if lunar.LunarYear() != year || lunar.LunarMonth() != month || lunar.LunarDay() != day || lunar.IsLeap() != leap {
		return Time{}, false
	}
	return result, true
}

// JieQi 节气时刻（北京时间） / solar term instant in Beijing time.
//
// 返回传入年份、节气对应的北京时间节气时间。
// Returns the Beijing-time instant of the requested solar term in the supplied year.
func JieQi(year, term int) time.Time {
	calcJD := basic.GetJQTime(year, term)
	zone := time.FixedZone("CST", 8*3600)
	return basic.JD2DateByZone(calcJD, zone, false)
}

// CalendricalJieQi 历法相符节气日期（北京时间当天 0 点） / calendrical solar-term date at Beijing midnight.
//
// 返回默认历法下指定公历年、节气落在的日期，时间固定为北京时间当天 0 点。
// 该函数沿用 `JieQi` 的节气编号，但结果是历法日期，不是现代天文学计算出的精确节气时刻。
// Returns the date on which the requested solar term falls in the default calendrical system,
// normalized to 00:00:00 at UTC+08:00. The term numbering is the same as `JieQi`, but the
// result is a calendrical date rather than the exact modern astronomical instant.
func CalendricalJieQi(year, term int) (time.Time, error) {
	return CalendricalJieQiWithCalendar(year, term, AncientCalendarDefault)
}

// CalendricalJieQiWithCalendar 历法相符节气日期（显式历法） / calendrical solar-term date with an explicit calendar.
//
// 返回指定古历系统中某公历年、节气落在的日期，时间固定为北京时间当天 0 点。
// 春秋历及缺少历法节气资料的年份会返回错误。
// Returns the date on which the requested solar term falls in the specified ancient
// calendar system, normalized to 00:00:00 at UTC+08:00. Calendars or years without
// calendrical solar-term data return an error.
func CalendricalJieQiWithCalendar(year, term int, system AncientCalendarSystem) (time.Time, error) {
	return calendricalJieQiWithCalendar(year, term, system)
}

// WuHou 物候时刻（北京时间） / pentad instant in Beijing time.
//
// 返回传入年份、物候对应的北京时间物候时间。
// Returns the Beijing-time instant of the requested pentad in the supplied year.
func WuHou(year, term int) time.Time {
	calcJD := basic.GetWuHouTime(year, term)
	zone := time.FixedZone("CST", 8*3600)
	return basic.JD2DateByZone(calcJD, zone, false)
}

func rapidLunarModern(year, month, day int) (int, int, int, bool, string) {
	// 1900~2026 为确定闰秒表结果，2027 起按现行时标政策生成。
	var upper = []uint16{32274, 52242, 41001, 30036, 49204, 36918, 25882, 46101, 34854, 22674, 43026, 31145, 51241, 38964, 26997, 47149, 36885, 23717, 44069, 34258, 53266, 41001, 29036, 49178, 37915, 24875, 46090, 34853, 23698, 43026, 31129, 50229, 38970, 26971, 47126, 36874, 24804, 44068, 32242, 52274, 41013, 28086, 48173, 37909, 25898, 46089, 34852, 22706, 43050, 30189, 50203, 38957, 27989, 47123, 35881, 24788, 45076, 32298, 51258, 40986, 29099, 48170, 37906, 25897, 46121, 34836, 21754, 42038, 31190, 50197, 38949, 27986, 48146, 35881, 23860, 44084, 32309, 51245, 39981, 29093, 49189, 37906, 25897, 46121, 35500, 53274, 42011, 30123, 50218, 38949, 27986, 48146, 36889, 23770, 43066, 32282, 52246, 39978, 29028, 49188, 37938, 24885, 45109, 33846, 22678, 42005, 30186, 51209, 39972, 26994, 47146, 35885, 23853, 43051, 32341, 52242, 41001, 29076, 49172, 37930, 25885, 45082, 33835, 22675, 43026, 30121, 50217, 38964, 27002, 46133, 35862, 23770, 44069, 32466, 52242, 41001, 29108, 48180, 36917, 24950, 45101, 33813, 22674, 43026, 31209, 50217, 38954, 26989, 47131, 34859, 23765, 44068, 34322, 52242, 40985, 29082, 48186, 36890, 24874, 45098, 34852, 21746, 42034, 30197, 50229, 37942, 26966, 47125, 35881, 23828, 44052, 32298, 52266, 39981, 28077, 48171, 37909, 24873, 45097, 34836, 22762, 42010, 30172, 50202, 38955, 26963, 47122, 35881, 24852, 43060, 31290, 51253, 39990, 28058, 48149, 37906, 25897, 45097, 33844, 21686, 42037, 30198, 50221, 39957, 29010, 48146, 36905, 24884, 45098, 32365, 52251, 41003, 30101, 49188, 38930, 26905, 47125, 34842, 22747, 43034, 31210, 50218, 39956, 28018, 48170, 35893, 23866, 44086, 34518, 52245, 41001, 30100, 50196, 37930, 25973, 46124, 34861, 22677, 43029, 31209, 51241, 39956, 28010, 48154, 36892, 23853, 44058, 34507, 53266, 41001, 30100, 49204, 37946, 25946, 45110, 34838, 23754, 43018, 31208, 51240, 39988, 27061, 47149, 35893, 24854, 44053, 34442, 53265, 42024, 29098, 49194, 37933, 25965, 45099, 34837, 23753, 44049, 31192, 51220, 39962, 28059, 47126, 35882, 24852, 45074, 31785, 21652, 41012, 29114, 48181, 37910, 25962, 46121, 34834, 22761, 43049, 31220, 50220, 38957, 28053, 48139, 36901, 25874, 46098, 35433, 53273, 42010, 30125, 50202, 38922, 26917, 47140, 36882, 23769, 43065, 31226, 51254, 39958, 29002, 49162, 37924, 24884, 45106, 34421, 53293, 41005, 30165, 50197, 39945, 26980, 47140, 35882, 23797, 43053, 31277, 51243, 40981, 29001, 49161, 37908, 25898, 45082, 34523, 53270, 42026, 30098, 50194, 38953, 27988, 46132, 34874, 23770, 44054, 31210, 51237, 40978, 29097, 48169, 36916, 24950, 45101, 33813, 21605, 42021, 31186, 50194, 38953, 26988, 47130, 34859, 22765, 44042, 32293, 51236, 40978, 29081, 48185, 35898, 24859, 45078, 34826, 21669, 42020, 30130, 50226, 37941, 25974, 46125, 35861, 22762, 44041, 32228, 52260, 39978, 28077, 48157, 36909, 24853, 45075, 33833, 22676, 43028, 31146, 50234, 39962, 26987, 47146, 36882, 24809, 44073, 34260, 52276, 41014, 29082, 49173, 37926, 26898, 46098, 35497, 54313, 43060, 30197, 50221, 38957, 28005, 47141, 36882, 24809, 45097, 32300, 52250, 40987, 29101, 48170, 37925, 26898, 47122, 34841, 22746, 42042, 31195, 50198, 38954, 28004, 48164, 35890, 23865, 44085, 32310, 51245, 40981, 29098, 50185, 37924, 25970, 46122, 34861, 21614, 42029, 31189, 51219, 38953, 27988, 48148, 36906, 23837, 44058, 32299, 52266, 40978, 29097, 49193, 38932, 24954, 45110, 34838, 22682, 42021, 31186, 51218, 39977, 26996, 47156, 35893, 23862, 43053, 32405, 52261, 42002, 29097, 49193, 37932, 25901, 45083, 33835, 22677, 43045, 31122, 51218, 39961, 27994}
	var lower = []uint8{218, 184, 92, 154, 152, 84, 170, 168, 180, 186, 184, 54, 52, 148, 82, 84, 168, 180, 108, 110, 108, 44, 150, 148, 80, 106, 216, 92, 94, 92, 44, 40, 148, 82, 180, 216, 220, 184, 90, 84, 40, 148, 84, 168, 182, 116, 180, 86, 84, 42, 40, 84, 106, 104, 108, 174, 172, 84, 84, 168, 84, 212, 216, 92, 92, 152, 76, 84, 170, 168, 180, 186, 180, 52, 154, 148, 74, 80, 168, 180, 108, 108, 46, 44, 150, 148, 80, 104, 216, 92, 94, 92, 44, 148, 148, 202, 176, 216, 218, 184, 88, 42, 40, 148, 170, 168, 182, 116, 180, 86, 84, 40, 84, 84, 106, 232, 108, 174, 172, 76, 42, 168, 84, 106, 216, 92, 56, 152, 76, 76, 168, 212, 180, 186, 180, 52, 150, 148, 72, 168, 104, 180, 182, 108, 46, 44, 148, 74, 72, 104, 108, 220, 94, 92, 44, 148, 148, 200, 216, 184, 184, 92, 88, 42, 40, 148, 170, 168, 180, 186, 180, 86, 84, 40, 84, 84, 104, 116, 108, 172, 78, 76, 166, 168, 84, 106, 216, 92, 156, 88, 76, 72, 168, 212, 180, 184, 58, 52, 84, 74, 72, 164, 104, 116, 182, 108, 44, 150, 148, 74, 72, 88, 108, 220, 92, 46, 44, 148, 74, 168, 212, 180, 184, 92, 88, 40, 148, 84, 170, 168, 180, 186, 180, 84, 42, 168, 84, 170, 104, 116, 108, 172, 46, 44, 164, 84, 212, 106, 216, 92, 92, 88, 44, 164, 164, 212, 218, 184, 186, 180, 84, 42, 72, 164, 180, 116, 182, 108, 172, 86, 84, 40, 100, 84, 108, 110, 92, 174, 172, 84, 42, 168, 212, 218, 184, 92, 172, 168, 84, 84, 168, 212, 180, 184, 86, 52, 150, 164, 84, 170, 104, 116, 182, 108, 46, 44, 164, 82, 212, 216, 108, 220, 92, 44, 40, 148, 164, 212, 218, 184, 184, 90, 84, 42, 40, 164, 180, 108, 116, 182, 172, 84, 42, 40, 84, 84, 108, 110, 92, 172, 86, 84, 42, 168, 212, 218, 184, 92, 154, 152, 84, 170, 168, 180, 180, 184, 54, 52, 148, 82, 84, 168, 180, 108, 174, 108, 44, 150, 148, 80, 106, 216, 108, 220, 92, 44, 40, 148, 82, 180, 216, 92, 184, 90, 84, 40, 148, 148, 168, 182, 116, 182, 172, 84, 42, 40, 84, 170, 104, 108, 174, 172, 84, 84, 168, 84, 212, 216, 92, 92, 154, 152, 84, 170, 168, 180, 186, 180, 54, 52, 148, 74, 80, 168, 180, 108, 108, 46, 44, 150, 148, 80, 104, 216, 92, 110, 92, 44, 148, 148, 74, 176, 216, 218, 184, 88, 42, 40, 148, 84, 168, 182, 116, 180, 86, 84, 40, 148, 84, 106, 104, 108, 174, 172, 76, 42, 168, 84, 212, 216, 92, 92, 152, 76, 76, 168, 212, 180, 186, 180, 52, 150, 148, 72, 168, 168, 180, 108, 108, 46, 44, 148, 74, 72, 104, 108, 92, 94, 92, 44, 148}
	if year < 1900 || year > 2400 {
		return 0, 0, 0, false, "超过日期限制"
	}
	useGoto := false
recalc:
	idx := year - 1900
	magic := int32(upper[idx])<<8 + int32(lower[idx])
	springMonth := (magic&0x800000)>>23 + 1
	springDay := (magic & 0x7FFFFF) >> 18
	if !useGoto && springMonth == int32(month) && springDay == int32(day) {
		return year, 1, 1, false, "正月初一"
	}
	if !useGoto && (springMonth > int32(month) || (springMonth == int32(month) && springDay > int32(day))) {
		year--
		useGoto = true
		goto recalc
	}
	calcYear := year
	if useGoto {
		calcYear++
	}
	target := time.Date(calcYear, time.Month(month), day, 0, 0, 0, 0, getCst())
	spring := time.Date(year, time.Month(int(springMonth)), int(springDay), 0, 0, 0, 0, getCst())
	diffDay := int(target.Sub(spring).Hours() / 24)
	lunarMonth := 1
	totalDay := 0
	isLeap := false
	leapMonth := int(uint8(magic>>14) & 0xF)
	strmonth := []string{"十", "一", "二", "三", "四", "五", "六", "七", "八", "九", "十", "冬", "腊"}
	strday := []string{"初", "十", "廿", "三"}
	for i := 0; i < 13; i++ {
		var dayofLunar = 29
		if uint8(magic&0x3FFE>>(13-i))&1 == 1 {
			dayofLunar++
		}
		if totalDay+dayofLunar > diffDay {
			var result string
			if isLeap {
				result += "闰"
			}
			if lunarMonth == 1 {
				result += "正月"
			} else {
				result += strmonth[lunarMonth] + "月"
			}
			lday := diffDay - totalDay + 1
			if lday == 20 {
				result += "二十"
			} else if lday == 10 {
				result += "初十"
			} else {
				result += strday[lday/10] + strmonth[lday%10]
			}
			return year, lunarMonth, lday, isLeap, result
		}
		totalDay += dayofLunar
		lunarMonth++
		if lunarMonth-leapMonth == 1 && !isLeap {
			isLeap = true
			lunarMonth--
		} else {
			isLeap = false
		}
	}
	return 0, 0, 0, false, "无法获取农历信息"
}

func rapidSolarModern(year, month, day int, isLeap bool) time.Time {
	// 1900~2026 为确定闰秒表结果，2027 起按现行时标政策生成。
	var upper = []uint16{32274, 52242, 41001, 30036, 49204, 36918, 25882, 46101, 34854, 22674, 43026, 31145, 51241, 38964, 26997, 47149, 36885, 23717, 44069, 34258, 53266, 41001, 29036, 49178, 37915, 24875, 46090, 34853, 23698, 43026, 31129, 50229, 38970, 26971, 47126, 36874, 24804, 44068, 32242, 52274, 41013, 28086, 48173, 37909, 25898, 46089, 34852, 22706, 43050, 30189, 50203, 38957, 27989, 47123, 35881, 24788, 45076, 32298, 51258, 40986, 29099, 48170, 37906, 25897, 46121, 34836, 21754, 42038, 31190, 50197, 38949, 27986, 48146, 35881, 23860, 44084, 32309, 51245, 39981, 29093, 49189, 37906, 25897, 46121, 35500, 53274, 42011, 30123, 50218, 38949, 27986, 48146, 36889, 23770, 43066, 32282, 52246, 39978, 29028, 49188, 37938, 24885, 45109, 33846, 22678, 42005, 30186, 51209, 39972, 26994, 47146, 35885, 23853, 43051, 32341, 52242, 41001, 29076, 49172, 37930, 25885, 45082, 33835, 22675, 43026, 30121, 50217, 38964, 27002, 46133, 35862, 23770, 44069, 32466, 52242, 41001, 29108, 48180, 36917, 24950, 45101, 33813, 22674, 43026, 31209, 50217, 38954, 26989, 47131, 34859, 23765, 44068, 34322, 52242, 40985, 29082, 48186, 36890, 24874, 45098, 34852, 21746, 42034, 30197, 50229, 37942, 26966, 47125, 35881, 23828, 44052, 32298, 52266, 39981, 28077, 48171, 37909, 24873, 45097, 34836, 22762, 42010, 30172, 50202, 38955, 26963, 47122, 35881, 24852, 43060, 31290, 51253, 39990, 28058, 48149, 37906, 25897, 45097, 33844, 21686, 42037, 30198, 50221, 39957, 29010, 48146, 36905, 24884, 45098, 32365, 52251, 41003, 30101, 49188, 38930, 26905, 47125, 34842, 22747, 43034, 31210, 50218, 39956, 28018, 48170, 35893, 23866, 44086, 34518, 52245, 41001, 30100, 50196, 37930, 25973, 46124, 34861, 22677, 43029, 31209, 51241, 39956, 28010, 48154, 36892, 23853, 44058, 34507, 53266, 41001, 30100, 49204, 37946, 25946, 45110, 34838, 23754, 43018, 31208, 51240, 39988, 27061, 47149, 35893, 24854, 44053, 34442, 53265, 42024, 29098, 49194, 37933, 25965, 45099, 34837, 23753, 44049, 31192, 51220, 39962, 28059, 47126, 35882, 24852, 45074, 31785, 21652, 41012, 29114, 48181, 37910, 25962, 46121, 34834, 22761, 43049, 31220, 50220, 38957, 28053, 48139, 36901, 25874, 46098, 35433, 53273, 42010, 30125, 50202, 38922, 26917, 47140, 36882, 23769, 43065, 31226, 51254, 39958, 29002, 49162, 37924, 24884, 45106, 34421, 53293, 41005, 30165, 50197, 39945, 26980, 47140, 35882, 23797, 43053, 31277, 51243, 40981, 29001, 49161, 37908, 25898, 45082, 34523, 53270, 42026, 30098, 50194, 38953, 27988, 46132, 34874, 23770, 44054, 31210, 51237, 40978, 29097, 48169, 36916, 24950, 45101, 33813, 21605, 42021, 31186, 50194, 38953, 26988, 47130, 34859, 22765, 44042, 32293, 51236, 40978, 29081, 48185, 35898, 24859, 45078, 34826, 21669, 42020, 30130, 50226, 37941, 25974, 46125, 35861, 22762, 44041, 32228, 52260, 39978, 28077, 48157, 36909, 24853, 45075, 33833, 22676, 43028, 31146, 50234, 39962, 26987, 47146, 36882, 24809, 44073, 34260, 52276, 41014, 29082, 49173, 37926, 26898, 46098, 35497, 54313, 43060, 30197, 50221, 38957, 28005, 47141, 36882, 24809, 45097, 32300, 52250, 40987, 29101, 48170, 37925, 26898, 47122, 34841, 22746, 42042, 31195, 50198, 38954, 28004, 48164, 35890, 23865, 44085, 32310, 51245, 40981, 29098, 50185, 37924, 25970, 46122, 34861, 21614, 42029, 31189, 51219, 38953, 27988, 48148, 36906, 23837, 44058, 32299, 52266, 40978, 29097, 49193, 38932, 24954, 45110, 34838, 22682, 42021, 31186, 51218, 39977, 26996, 47156, 35893, 23862, 43053, 32405, 52261, 42002, 29097, 49193, 37932, 25901, 45083, 33835, 22677, 43045, 31122, 51218, 39961, 27994}
	var lower = []uint8{218, 184, 92, 154, 152, 84, 170, 168, 180, 186, 184, 54, 52, 148, 82, 84, 168, 180, 108, 110, 108, 44, 150, 148, 80, 106, 216, 92, 94, 92, 44, 40, 148, 82, 180, 216, 220, 184, 90, 84, 40, 148, 84, 168, 182, 116, 180, 86, 84, 42, 40, 84, 106, 104, 108, 174, 172, 84, 84, 168, 84, 212, 216, 92, 92, 152, 76, 84, 170, 168, 180, 186, 180, 52, 154, 148, 74, 80, 168, 180, 108, 108, 46, 44, 150, 148, 80, 104, 216, 92, 94, 92, 44, 148, 148, 202, 176, 216, 218, 184, 88, 42, 40, 148, 170, 168, 182, 116, 180, 86, 84, 40, 84, 84, 106, 232, 108, 174, 172, 76, 42, 168, 84, 106, 216, 92, 56, 152, 76, 76, 168, 212, 180, 186, 180, 52, 150, 148, 72, 168, 104, 180, 182, 108, 46, 44, 148, 74, 72, 104, 108, 220, 94, 92, 44, 148, 148, 200, 216, 184, 184, 92, 88, 42, 40, 148, 170, 168, 180, 186, 180, 86, 84, 40, 84, 84, 104, 116, 108, 172, 78, 76, 166, 168, 84, 106, 216, 92, 156, 88, 76, 72, 168, 212, 180, 184, 58, 52, 84, 74, 72, 164, 104, 116, 182, 108, 44, 150, 148, 74, 72, 88, 108, 220, 92, 46, 44, 148, 74, 168, 212, 180, 184, 92, 88, 40, 148, 84, 170, 168, 180, 186, 180, 84, 42, 168, 84, 170, 104, 116, 108, 172, 46, 44, 164, 84, 212, 106, 216, 92, 92, 88, 44, 164, 164, 212, 218, 184, 186, 180, 84, 42, 72, 164, 180, 116, 182, 108, 172, 86, 84, 40, 100, 84, 108, 110, 92, 174, 172, 84, 42, 168, 212, 218, 184, 92, 172, 168, 84, 84, 168, 212, 180, 184, 86, 52, 150, 164, 84, 170, 104, 116, 182, 108, 46, 44, 164, 82, 212, 216, 108, 220, 92, 44, 40, 148, 164, 212, 218, 184, 184, 90, 84, 42, 40, 164, 180, 108, 116, 182, 172, 84, 42, 40, 84, 84, 108, 110, 92, 172, 86, 84, 42, 168, 212, 218, 184, 92, 154, 152, 84, 170, 168, 180, 180, 184, 54, 52, 148, 82, 84, 168, 180, 108, 174, 108, 44, 150, 148, 80, 106, 216, 108, 220, 92, 44, 40, 148, 82, 180, 216, 92, 184, 90, 84, 40, 148, 148, 168, 182, 116, 182, 172, 84, 42, 40, 84, 170, 104, 108, 174, 172, 84, 84, 168, 84, 212, 216, 92, 92, 154, 152, 84, 170, 168, 180, 186, 180, 54, 52, 148, 74, 80, 168, 180, 108, 108, 46, 44, 150, 148, 80, 104, 216, 92, 110, 92, 44, 148, 148, 74, 176, 216, 218, 184, 88, 42, 40, 148, 84, 168, 182, 116, 180, 86, 84, 40, 148, 84, 106, 104, 108, 174, 172, 76, 42, 168, 84, 212, 216, 92, 92, 152, 76, 76, 168, 212, 180, 186, 180, 52, 150, 148, 72, 168, 168, 180, 108, 108, 46, 44, 148, 74, 72, 104, 108, 92, 94, 92, 44, 148}
	if year < 1900 || year > 2400 {
		return time.Time{}
	}
	idx := year - 1900
	magic := int32(upper[idx])<<8 + int32(lower[idx])
	springMonth := (magic&0x800000)>>23 + 1
	springDay := (magic & 0x7FFFFF) >> 18
	spring := time.Date(year, time.Month(int(springMonth)), int(springDay), 0, 0, 0, 0, getCst())
	lunarMonth := 1
	totalDay := 0
	leap := false
	leapMonth := int(uint8(magic>>14) & 0xF)
	for i := 0; i < 13; i++ {
		if lunarMonth == month && isLeap == leap {
			target := spring.AddDate(0, 0, totalDay+day-1)
			return target
		}
		var dayofLunar = 29
		if uint8(magic&0x3FFE>>(13-i))&1 == 1 {
			dayofLunar++
		}
		totalDay += dayofLunar
		lunarMonth++
		if lunarMonth-leapMonth == 1 && !leap {
			leap = true
			lunarMonth--
		} else {
			leap = false
		}
	}
	return time.Time{}
}

// 中文数字到阿拉伯数字的映射
var chineseNumbers = map[string]int{
	"元": 1, "一": 1, "二": 2, "三": 3, "四": 4, "五": 5,
	"六": 6, "七": 7, "八": 8, "九": 9, "十": 10,
	"十一": 11, "十二": 12, "十三": 13, "十四": 14, "十五": 15,
	"十六": 16, "十七": 17, "十八": 18, "十九": 19, "二十": 20,
	"廿一": 21, "廿二": 22, "廿三": 23, "廿四": 24, "廿五": 25,
	"廿六": 26, "廿七": 27, "廿八": 28, "廿九": 29, "三十": 30, "卅一": 31,
}

var chineseDays = map[string]int{
	"初一": 1, "初二": 2, "初三": 3, "初四": 4, "初五": 5,
	"初六": 6, "初七": 7, "初八": 8, "初九": 9, "初十": 10,
	"十一": 11, "十二": 12, "十三": 13, "十四": 14, "十五": 15,
	"十六": 16, "十七": 17, "十八": 18, "十九": 19, "二十": 20,
	"廿一": 21, "廿二": 22, "廿三": 23, "廿四": 24, "廿五": 25,
	"廿六": 26, "廿七": 27, "廿八": 28, "廿九": 29, "三十": 30, "卅一": 31,
}

// 中文月份到数字的映射
var chineseMonths = map[string]int{
	"正": 1, "腊": 12, "冬": 11, "十一": 11, "十二": 12,
	"一": 1, "二": 2, "三": 3, "四": 4, "五": 5, "六": 6,
	"七": 7, "八": 8, "九": 9, "十": 10,
}

// 正则提到包级：文本入口每次解析都会用到，每次重编译会白花几十微秒。
var (
	chineseDatePrefixYearPattern = regexp.MustCompile(`^([-一二三四五六七八九十零〇\d]+?)年`)
	digitPattern                 = regexp.MustCompile(`\d+`)
	chineseDatePattern           = regexp.MustCompile(`^([\p{Han}]+?)([-负負一二三四五六七八九十零〇\d]*?元?)年([\p{Han}\d]+?)月([\p{Han}\d]+?)日?$`)
)

func parseChineseDate(dateStr string) (LunarTime, error) {
	var result LunarTime
	var err error
	result.desc = dateStr
	if strings.HasPrefix(dateStr, "前") {
		originDateStr := dateStr
		dateStr = strings.TrimPrefix(dateStr, "前")
		re := chineseDatePrefixYearPattern
		matches := re.FindStringSubmatch(dateStr)
		if len(matches) == 2 {
			year, err := parseDirectYear(matches[1])
			if err != nil {
				return result, err
			}
			dateStr = "-" + strconv.Itoa(year-1) + strings.TrimPrefix(dateStr, matches[1])
		} else {
			dateStr = originDateStr
		}
	}
	if !strings.HasPrefix(dateStr, "公元") {
		dateStr = "公元" + dateStr
	}
	// 正则表达式匹配日期格式
	re := chineseDatePattern
	matches := re.FindStringSubmatch(dateStr)
	if len(matches) < 5 {
		return result, fmt.Errorf("无效的日期格式: %s", dateStr)
	}
	matches[1] = strings.TrimPrefix(matches[1], "公元")
	// 提取年号
	result.comment = matches[1]

	// 转换年份
	if result.comment != "" {
		result.year, err = convertChineseNumber(matches[2])
		if err != nil {
			return result, fmt.Errorf("无效的年份: %s", matches[2])
		}
	} else {
		// 直接转换年份
		yearStr := matches[2]
		sign := 1
		if strings.HasPrefix(yearStr, "负") || strings.HasPrefix(yearStr, "負") {
			sign = -1
			yearStr = strings.TrimPrefix(strings.TrimPrefix(yearStr, "负"), "負")
		}
		if digitPattern.MatchString(yearStr) {
			result.year, err = strconv.Atoi(yearStr)
			if err != nil {
				return result, fmt.Errorf("无效的年份: %s", matches[2])
			}
		} else {
			result.year = transfer(yearStr, true)
		}
		result.year *= sign
	}

	// 转换月份
	monthStr := matches[3]
	if strings.HasPrefix(monthStr, "闰") {
		result.leap = true
		monthStr = strings.TrimPrefix(monthStr, "闰")
	}
	if strings.HasPrefix(monthStr, "后") {
		result.leap = true
		result.houMonth = true
		monthStr = strings.TrimPrefix(monthStr, "后")
	} else if strings.HasPrefix(monthStr, "後") {
		result.leap = true
		result.houMonth = true
		monthStr = strings.TrimPrefix(monthStr, "後")
	}
	if month, ok := chineseMonths[monthStr]; ok {
		result.month = month
	} else {
		if digitPattern.MatchString(monthStr) {
			result.month, err = strconv.Atoi(monthStr)
		} else {
			// 尝试将月份字符串转换为数字
			result.month, err = convertChineseNumber(monthStr)
		}
		if err != nil {
			return result, fmt.Errorf("无效的月份: %s", monthStr)
		}
	}
	if result.houMonth && result.month != 9 {
		return result, fmt.Errorf("无效的月份: %s", matches[3])
	}

	// 转换日期
	dayStr := matches[4]
	//判断是否是干支日
	if len(dayStr) == 6 && strings.ContainsAny(dayStr, "甲乙丙丁戊己庚辛壬癸子丑寅卯辰巳午未申酉戌亥") {
		// 临时使用干支月代替
		result.ganzhiMonth = dayStr
		result.day = 0
		return result, nil
	}
	if day, ok := chineseDays[dayStr]; ok {
		result.day = day
	} else {
		if digitPattern.MatchString(dayStr) {
			result.day, err = strconv.Atoi(dayStr)
			if err != nil {
				return result, fmt.Errorf("无效的日期: %s", dayStr)
			}
			return result, nil
		}
		// 尝试将日期字符串转换为数字
		result.day, err = convertChineseNumber(dayStr)
		if err != nil {
			return result, fmt.Errorf("无效的日期: %s", dayStr)
		}
	}
	return result, nil
}

func parseDirectYear(yearStr string) (int, error) {
	if digitPattern.MatchString(yearStr) {
		year, err := strconv.Atoi(yearStr)
		if err != nil {
			return 0, fmt.Errorf("无效的年份: %s", yearStr)
		}
		return year, nil
	}
	year := transfer(yearStr, true)
	if year == 0 {
		return 0, fmt.Errorf("无效的年份: %s", yearStr)
	}
	return year, nil
}

// convertChineseNumber 将中文数字转换为阿拉伯数字
func convertChineseNumber(chineseNum string) (int, error) {
	if num, ok := chineseNumbers[chineseNum]; ok {
		return num, nil
	}

	return transfer(chineseNum, false), nil
}

func number2Chinese(num int, isDirectTrans bool) string {
	chs := []string{"零", "一", "二", "三", "四", "五", "六", "七", "八", "九"}
	if isDirectTrans {
		if num < 0 {
			return "负" + number2Chinese(-num, true)
		}
		var res string
		for i := 0; i < 4; i++ {
			tmp := num / (int(math.Pow10(3 - i)))
			if tmp == 0 && i == 0 {
				continue
			}
			res += chs[tmp]
			num = num % (int(math.Pow10(3 - i)))
		}
		return res
	}
	if num < 0 || num > 99 {
		return ""
	}
	if num < 10 {
		return chs[num]
	}
	if num == 10 {
		return "十"
	}
	if num < 20 {
		return "十" + chs[num-10]
	}
	if num%10 == 0 {
		return chs[num/10] + "十"
	}
	return chs[num/10] + "十" + chs[num%10]
}

func transfer(msg string, direct bool) int {
	keyMap := map[rune]int{
		'〇': 0, '零': 0, '一': 1, '二': 2, '三': 3, '四': 4, '五': 5, '六': 6, '七': 7, '八': 8, '九': 9, '十': 10, '百': 100, '千': 1000, '万': 10000, '亿': 100000000, '两': 2, '俩': 2,
	}
	result := 0
	if direct {
		for _, num := range []rune(msg) {
			if val, match := keyMap[num]; match {
				result = result*10 + val
			} else {
				return 0
			}
		}
		return result
	}
	secCache := 0
	thrCache := 0
	fKWord := map[rune]int{'百': 100, '千': 1000, '万': 10000, '亿': 100000000}
	for _, num := range []rune(msg) {
		if _, match := fKWord[num]; !match {
			if num == '十' && thrCache != 0 {
				thrCache *= keyMap[num]
			} else {
				thrCache += keyMap[num]
			}
		} else {
			if fKWord[num] < 10000 {
				secCache += thrCache * fKWord[num]
				thrCache = 0
			} else {
				secCache += thrCache
				thrCache = 0
				if secCache == 0 {
					result *= fKWord[num]
					continue
				}
				result += secCache * fKWord[num]
				secCache = 0
			}
		}
	}

	result += secCache + thrCache
	return result
}

// GanZhiOfYear 年干支 / sexagenary year name.
func GanZhiOfYear(year int) string {
	return basic.GetGanZhi(year)
}

// GanZhiOfDay 日干支 / sexagenary day name.
func GanZhiOfDay(t time.Time) string {
	t = t.In(getCst())
	return ganZhiOfJD(Date2JD(time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, getCst())))
}

// ganZhiOfJD 由儒略日直接求日干支（儒略历独有闰日无法表示为 time.Time）。
func ganZhiOfJD(jd float64) string {
	diff := int(jd - 2451550.5)
	if diff >= 0 {
		return tiangan[diff%10] + dizhi[diff%12]
	}
	return tiangan[(diff%10+10)%10] + dizhi[(diff%12+12)%12]
}

// 获取每年建寅月的天干
func tianGanIndexForFirstMonth(year int) int {
	diff := (year - 1998) * 2
	if diff >= 0 {
		return diff % 10
	}
	return (diff%10 + 10) % 10
}

// commonGanZhiOfMonth 返回常规以建寅为正月时，指定农历月份的干支
func commonGanZhiOfMonth(year, month int) string {
	start := tianGanIndexForFirstMonth(year)
	index := (start + month - 1) % 10
	return tiangan[index] + dizhi[(month+1)%12]
}

func ganZhiOfDayIndex(t time.Time) (int, int) {
	localJD := Date2JD(time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, getCst()))
	diff := int(localJD - 2451550.5)
	if diff >= 0 {
		return diff % 10, diff % 12
	}
	return (diff%10 + 10) % 10, (diff%12 + 12) % 12
}
