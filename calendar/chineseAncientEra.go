package calendar

import "fmt"

type ancientEraSource struct {
	system AncientCalendarSystem
	min    int
	max    int
	eras   []Era
}

var zhouAncientEras = []Era{
	{Year: -313, Emperor: "周赧王", Nianhao: "周赧王", Dynasty: "周"},
	{Year: -319, Emperor: "周慎靓王", Nianhao: "周慎靓王", Dynasty: "周"},
	{Year: -367, Emperor: "周显王", Nianhao: "周显王", Dynasty: "周"},
	{Year: -374, Emperor: "周烈王", Nianhao: "周烈王", Dynasty: "周"},
	{Year: -400, Emperor: "周安王", Nianhao: "周安王", Dynasty: "周"},
	{Year: -424, Emperor: "周威烈王", Nianhao: "周威烈王", Dynasty: "周"},
	{Year: -439, Emperor: "周考王", Nianhao: "周考王", Dynasty: "周"},
	{Year: -467, Emperor: "周贞定王", Nianhao: "周贞定王", Dynasty: "周"},
	{Year: -475, Emperor: "周元王", Nianhao: "周元王", Dynasty: "周"},
	{Year: -518, Emperor: "周敬王", Nianhao: "周敬王", Dynasty: "周"},
	{Year: -543, Emperor: "周景王", Nianhao: "周景王", Dynasty: "周"},
	{Year: -570, Emperor: "周灵王", Nianhao: "周灵王", Dynasty: "周"},
	{Year: -584, Emperor: "周简王", Nianhao: "周简王", Dynasty: "周"},
	{Year: -605, Emperor: "周定王", Nianhao: "周定王", Dynasty: "周"},
	{Year: -611, Emperor: "周匡王", Nianhao: "周匡王", Dynasty: "周"},
	{Year: -617, Emperor: "周顷王", Nianhao: "周顷王", Dynasty: "周"},
	{Year: -650, Emperor: "周襄王", Nianhao: "周襄王", Dynasty: "周"},
	{Year: -675, Emperor: "周惠王", Nianhao: "周惠王", Dynasty: "周"},
	{Year: -680, Emperor: "周僖王", Nianhao: "周僖王", Dynasty: "周"},
	{Year: -695, Emperor: "周庄王", Nianhao: "周庄王", Dynasty: "周"},
	{Year: -718, Emperor: "周桓王", Nianhao: "周桓王", Dynasty: "周"},
	{Year: -771, Emperor: "周平王", Nianhao: "周平王", Dynasty: "周"},
}

var luAncientEras = []Era{
	{Year: -271, Emperor: "鲁顷公", Nianhao: "鲁顷公", Dynasty: "鲁"},
	{Year: -294, Emperor: "鲁文公", Nianhao: "鲁文公", Dynasty: "鲁"},
	{Year: -313, Emperor: "鲁平公", Nianhao: "鲁平公", Dynasty: "鲁"},
	{Year: -342, Emperor: "鲁景公", Nianhao: "鲁景公", Dynasty: "鲁"},
	{Year: -351, Emperor: "鲁康公", Nianhao: "鲁康公", Dynasty: "鲁"},
	{Year: -375, Emperor: "鲁共公", Nianhao: "鲁共公", Dynasty: "鲁"},
	{Year: -406, Emperor: "鲁穆公", Nianhao: "鲁穆公", Dynasty: "鲁"},
	{Year: -427, Emperor: "鲁元公", Nianhao: "鲁元公", Dynasty: "鲁"},
	{Year: -465, Emperor: "鲁悼公", Nianhao: "鲁悼公", Dynasty: "鲁"},
	{Year: -493, Emperor: "鲁哀公", Nianhao: "鲁哀公", Dynasty: "鲁"},
	{Year: -508, Emperor: "鲁定公", Nianhao: "鲁定公", Dynasty: "鲁"},
	{Year: -540, Emperor: "鲁昭公", Nianhao: "鲁昭公", Dynasty: "鲁"},
	{Year: -571, Emperor: "鲁襄公", Nianhao: "鲁襄公", Dynasty: "鲁"},
	{Year: -589, Emperor: "鲁成公", Nianhao: "鲁成公", Dynasty: "鲁"},
	{Year: -607, Emperor: "鲁宣公", Nianhao: "鲁宣公", Dynasty: "鲁"},
	{Year: -625, Emperor: "鲁文公", Nianhao: "鲁文公", Dynasty: "鲁"},
	{Year: -658, Emperor: "鲁僖公", Nianhao: "鲁僖公", Dynasty: "鲁"},
	{Year: -660, Emperor: "鲁闵公", Nianhao: "鲁闵公", Dynasty: "鲁"},
	{Year: -692, Emperor: "鲁庄公", Nianhao: "鲁庄公", Dynasty: "鲁"},
	{Year: -710, Emperor: "鲁桓公", Nianhao: "鲁桓公", Dynasty: "鲁"},
	{Year: -721, Emperor: "鲁隐公", Nianhao: "鲁隐公", Dynasty: "鲁"},
	{Year: -767, Emperor: "鲁惠公", Nianhao: "鲁惠公", Dynasty: "鲁"},
}

var qinWarringAncientEras = []Era{
	{Year: -245, Emperor: "秦王政", Nianhao: "秦王政", Dynasty: "秦"},
	{Year: -248, Emperor: "秦庄襄王", Nianhao: "秦庄襄王", Dynasty: "秦"},
	{Year: -249, Emperor: "秦孝文王", Nianhao: "秦孝文王", Dynasty: "秦"},
	{Year: -305, Emperor: "秦昭襄王", Nianhao: "秦昭襄王", Dynasty: "秦"},
}

var qinHanAncientEras = []Era{
	{Year: -109, Emperor: "汉武帝", Nianhao: "元封", Dynasty: "西汉"},
	{Year: -115, Emperor: "汉武帝", Nianhao: "元鼎", Dynasty: "西汉"},
	{Year: -121, Emperor: "汉武帝", Nianhao: "元狩", Dynasty: "西汉"},
	{Year: -127, Emperor: "汉武帝", Nianhao: "元朔", Dynasty: "西汉"},
	{Year: -133, Emperor: "汉武帝", Nianhao: "元光", Dynasty: "西汉"},
	{Year: -139, Emperor: "汉武帝", Nianhao: "建元", Dynasty: "西汉"},
	{Year: -142, Emperor: "汉景帝", Nianhao: "后元", Dynasty: "西汉"},
	{Year: -148, Emperor: "汉景帝", Nianhao: "中元", Dynasty: "西汉"},
	{Year: -155, Emperor: "汉景帝", Nianhao: "前元", Dynasty: "西汉"},
	{Year: -162, Emperor: "汉文帝", Nianhao: "后元", Dynasty: "西汉"},
	{Year: -178, Emperor: "汉文帝", Nianhao: "前元", Dynasty: "西汉"},
	{Year: -186, Emperor: "汉高后", Nianhao: "汉高后", Dynasty: "西汉"},
	{Year: -193, Emperor: "汉惠帝", Nianhao: "汉惠帝", Dynasty: "西汉"},
	{Year: -205, Emperor: "汉高祖", Nianhao: "汉高祖", Dynasty: "西汉"},
	{Year: -208, Emperor: "秦二世", Nianhao: "秦二世", Dynasty: "秦"},
	{Year: -245, Emperor: "秦始皇", Nianhao: "秦始皇", Dynasty: "秦"},
}

func ancientErasForLunarYear(year int, system AncientCalendarSystem) []EraDesc {
	switch system {
	case AncientCalendarQinHan:
		return ancientErasInRange(year, qinHanAncientEras, qinHanMinYear, qinHanMaxYear)
	case AncientCalendarChunqiu, AncientCalendarLu:
		return ancientErasInRange(year, luAncientEras, ancientBoundaryMinYear, -248)
	case AncientCalendarZhuanxu:
		if eras := ancientErasInRange(year, qinWarringAncientEras, -305, ancientBoundaryMaxYear); len(eras) > 0 {
			return eras
		}
		return ancientErasInRange(year, zhouAncientEras, ancientBoundaryMinYear, -255)
	case AncientCalendarZhou, AncientCalendarHuangdi, AncientCalendarYin, AncientCalendarXia1, AncientCalendarXia2:
		return ancientErasInRange(year, zhouAncientEras, ancientBoundaryMinYear, -255)
	default:
		return nil
	}
}

func ancientErasInRange(year int, eras []Era, min, max int) []EraDesc {
	if len(eras) == 0 || year < min || year > max {
		return nil
	}
	return innerEras(year, func() []Era { return eras })
}

func ancientEraSourcesForSystem(system AncientCalendarSystem) []ancientEraSource {
	switch system {
	case AncientCalendarDefault:
		return []ancientEraSource{
			{system: AncientCalendarQinHan, min: qinHanMinYear, max: qinHanMaxYear, eras: qinHanAncientEras},
			{system: AncientCalendarChunqiu, min: ancientBoundaryMinYear, max: -480, eras: luAncientEras},
			{system: AncientCalendarZhou, min: -479, max: -255, eras: zhouAncientEras},
		}
	case AncientCalendarQinHan:
		return []ancientEraSource{{system: system, min: qinHanMinYear, max: qinHanMaxYear, eras: qinHanAncientEras}}
	case AncientCalendarChunqiu, AncientCalendarLu:
		return []ancientEraSource{{system: system, min: ancientBoundaryMinYear, max: -248, eras: luAncientEras}}
	case AncientCalendarZhuanxu:
		return []ancientEraSource{
			{system: system, min: -305, max: ancientBoundaryMaxYear, eras: qinWarringAncientEras},
			{system: system, min: ancientBoundaryMinYear, max: -255, eras: zhouAncientEras},
		}
	case AncientCalendarZhou, AncientCalendarHuangdi, AncientCalendarYin, AncientCalendarXia1, AncientCalendarXia2:
		return []ancientEraSource{{system: system, min: ancientBoundaryMinYear, max: -255, eras: zhouAncientEras}}
	default:
		return nil
	}
}

func lunarToSolarAncientEra(data LunarTime, system AncientCalendarSystem) ([]Time, bool, error) {
	sources := ancientEraSourcesForSystem(system)
	if len(sources) == 0 {
		return nil, false, nil
	}
	known := false
	var results []Time
	for _, source := range sources {
		years, matched := ancientEraYears(source, data.comment, data.year)
		if !matched {
			continue
		}
		known = true
		for _, year := range years {
			times, err := lunarToSolarAncientEraYear(data, year, source.system)
			if err == nil {
				results = append(results, times...)
			}
		}
	}
	if !known {
		return nil, false, nil
	}
	if len(results) == 0 {
		return nil, true, fmt.Errorf("未找到对应日期")
	}
	return results, true, nil
}

func ancientEraYears(source ancientEraSource, nianhao string, ordinal int) ([]int, bool) {
	var years []int
	matched := false
	for idx, era := range source.eras {
		if era.Nianhao != nianhao && era.Emperor != nianhao {
			continue
		}
		matched = true
		end := source.max
		if idx > 0 && source.eras[idx-1].Year-1 < end {
			end = source.eras[idx-1].Year - 1
		}
		year := era.Year + ordinal - 1 - era.Offset
		if year >= source.min && year >= era.Year && year <= end {
			years = append(years, year)
		}
	}
	return years, matched
}

func lunarToSolarAncientEraYear(data LunarTime, year int, system AncientCalendarSystem) ([]Time, error) {
	if data.houMonth && system != AncientCalendarQinHan && system != AncientCalendarZhuanxu {
		return nil, fmt.Errorf("未找到对应日期")
	}
	if data.ganzhiMonth == "" || data.day != 0 {
		result, err := LunarToSolarByYMDWithCalendar(year, data.month, data.day, data.leap, system)
		if err != nil {
			return nil, err
		}
		return []Time{result}, nil
	}

	var results []Time
	for day := 1; day <= 30; day++ {
		result, err := LunarToSolarByYMDWithCalendar(year, data.month, day, data.leap, system)
		if err != nil {
			continue
		}
		if GanZhiOfDay(result.Solar()) == data.ganzhiMonth {
			results = append(results, result)
		}
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("未找到对应日期")
	}
	return results, nil
}
