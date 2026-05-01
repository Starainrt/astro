package calendar

import (
	"time"

	"github.com/starainrt/astro/basic"
)

type LunarInfo struct {
	// SolarDate 公历日期
	SolarDate time.Time `json:"solarDate"`
	// LunarYear 农历年的公历映射，如2025
	LunarYear int `json:"lunarYear"`
	// LunarYearChn 农历年公历映射中文表示，比如二零二五
	LunarYearChn string `json:"lunarYearChn"`
	// LunarMonth 农历月，表示以当时的历法推定的农历月与正月的距离，正月为1，二月为2，依次类推
	// 武则天改历时期，正月为1, 十二月为2,一月为3，二月为4,以此类推
	LunarMonth int `json:"lunarMonth"`
	// LunarDay 农历日,[1-30]
	LunarDay int `json:"lunarDay"`
	// IsLeap 是否闰月
	IsLeap bool `json:"isLeap"`
	// LunarMonthDayDesc 农历月日描述，如正月初一。此处，十一月表示为冬月，十二月表示为腊月
	LunarMonthDayDesc string `json:"lunarMonthDayDesc"`
	// GanzhiYear 农历年干支
	GanzhiYear string `json:"ganzhiYear"`
	// GanzhiMonth 农历月干支，闰月从上一个月
	GanzhiMonth string `json:"ganzhiMonth"`
	// GanzhiDay 农历日干支
	GanzhiDay string `json:"ganzhiDay"`
	// Dynasty 朝代，如唐、宋、元、明、清等
	Dynasty string `json:"dynasty"`
	// Emperor 皇帝姓名(仅供参考，多个皇帝用同一个年号的场景，此处不准)
	Emperor string `json:"emperor"`
	// Nianhao  年号 如"开元"
	Nianhao string `json:"nianhao"`
	// YearOfNianhao 该年号的第几年
	YearOfNianhao int `json:"yearOfNianhao"`
	// EraDesc 年代描述，如唐玄宗开元二年
	EraDesc string `json:"eraDesc"`
	// LunarWithEraDesc 农历日期加上年代描述，如开元二年正月初一
	LunarWithEraDesc string `json:"lunarWithNianhaoDesc"`
	// ChineseZodiac 生肖
	ChineseZodiac string `json:"chineseZodiac"`
}

type Time struct {
	solarTime time.Time
	lunars    []LunarTime
}

// Solar 公历时间 / solar time.
//
// 返回内部保存的公历 `time.Time`，不做时区或历法再计算。
func (t Time) Solar() time.Time {
	return t.solarTime
}

// Time 公历时间 / solar time.
//
// 是 `Solar` 的同义接口，便于把 `calendar.Time` 当作普通时间对象使用。
func (t Time) Time() time.Time {
	return t.solarTime
}

// Lunars 农历候选结果 / lunar candidates.
//
// 返回全部候选农历结果。
func (t Time) Lunars() []LunarTime {
	return t.lunars
}

// LunarDesc 农历描述 / lunar descriptions.
//
// 返回全部候选结果的农历描述，如开元二年正月初一；若无年号，则返回年份描述，如二零二五年正月初一。
func (t Time) LunarDesc() []string {
	var res []string
	for _, v := range t.lunars {
		res = append(res, v.LunarDesc()...)
	}
	return res
}

// LunarDescWithEmperor 含君主信息的农历描述 / lunar descriptions with emperor.
//
// 返回全部候选结果中含有君主信息的农历描述，如唐玄宗 开元二年正月初一；若无年号，则返回年份描述，如二零二五年正月初一。
func (t Time) LunarDescWithEmperor() []string {
	var res []string
	for _, v := range t.lunars {
		res = append(res, v.LunarDescWithEmperor()...)
	}
	return res
}

// LunarDescWithDynasty 含朝代信息的农历描述 / lunar descriptions with dynasty.
//
// 返回全部候选结果中含有朝代信息的农历描述，如唐 开元二年正月初一；若无年号，则返回年份描述，如二零二五年正月初一。
func (t Time) LunarDescWithDynasty() []string {
	var res []string
	for _, v := range t.lunars {
		res = append(res, v.LunarDescWithDynasty()...)
	}
	return res
}

// LunarDescWithDynastyAndEmperor 含朝代与君主信息的农历描述 / lunar descriptions with dynasty and emperor.
//
// 返回全部候选结果中含有朝代和君主信息的农历描述，如唐 唐玄宗 开元二年正月初一；若无年号，则返回年份描述，如二零二五年正月初一。
func (t Time) LunarDescWithDynastyAndEmperor() []string {
	var res []string
	for _, v := range t.lunars {
		res = append(res, v.LunarDescWithDynastyAndEmperor()...)
	}
	return res
}

// LunarInfo 农历结构化信息 / structured lunar information.
//
// 返回全部候选结果对应的结构化农历信息切片。
func (t Time) LunarInfo() []LunarInfo {
	var res []LunarInfo
	for _, v := range t.lunars {
		res = append(res, v.LunarInfo()...)
	}
	return res
}

// Eras 朝代、皇帝、年号信息 / era information.
//
// 返回全部候选结果对应的朝代、皇帝、年号信息。
func (t Time) Eras() []EraDesc {
	var res []EraDesc
	for _, v := range t.lunars {
		res = append(res, v.eras...)
	}
	return res
}

// Lunar 首个农历结果 / first lunar result.
//
// 若存在多个候选结果，只返回第一个；无结果时返回零值 `LunarTime`。
func (t Time) Lunar() LunarTime {
	if len(t.lunars) > 0 {
		return t.lunars[0]
	}
	return LunarTime{}
}

// Add 时间偏移 / add a duration.
//
// 返回公历时间偏移后的农历结果。
func (t Time) Add(d time.Duration) Time {
	if d < time.Second {
		newT := t.solarTime.Add(d)
		rT, _ := SolarToLunar(newT)
		return rT
	}
	sec := d.Seconds()
	jde := Date2JDE(t.solarTime)
	jde += sec / 86400.0
	newT := basic.JDE2DateByZone(jde, t.solarTime.Location(), true)
	rT, _ := SolarToLunar(newT)
	return rT
}

type LunarTime struct {
	solarDate time.Time
	//农历年
	year int
	//农历月，表示以当时的历法推定的农历月与正月的距离，正月为1，二月为2，依次类推，闰月显示所闰月
	month int
	//农历日
	day int
	//是否闰月
	leap bool
	//农历描述
	desc string
	//备注
	comment string
	//ganzhi of month 月干支
	ganzhiMonth string

	eras []EraDesc
}

// ShengXiao 生肖 / Chinese zodiac.
func (l LunarTime) ShengXiao() string {
	shengxiao := []string{"猴", "鸡", "狗", "猪", "鼠", "牛", "虎", "兔", "龙", "蛇", "马", "羊"}
	diff := l.LunarYear() % 12
	if diff < 0 {
		diff += 12
	}
	return shengxiao[diff]
}

// Zodiac 生肖别名 / zodiac alias.
func (l LunarTime) Zodiac() string {
	return l.ShengXiao()
}

// GanZhiYear 年干支 / sexagenary year name.
func (l LunarTime) GanZhiYear() string {
	return GanZhiOfYear(l.year)
}

// GanZhiMonth 月干支 / sexagenary month name.
func (l LunarTime) GanZhiMonth() string {
	return l.ganzhiMonth
}

// GanZhiDay 日干支 / sexagenary day name.
func (l LunarTime) GanZhiDay() string {
	return GanZhiOfDay(l.solarDate)
}

// LunarYear 农历年 / lunar year.
func (l LunarTime) LunarYear() int {
	return l.year
}

// LunarMonth 农历月 / lunar month.
func (l LunarTime) LunarMonth() int {
	return l.month
}

// LunarDay 农历日 / lunar day.
func (l LunarTime) LunarDay() int {
	return l.day
}

// IsLeap 是否闰月 / whether the month is leap.
func (l LunarTime) IsLeap() bool {
	return l.leap
}

// Eras 朝代、皇帝、年号信息 / era information.
//
// 返回该农历结果对应的朝代、皇帝、年号信息。
func (l LunarTime) Eras() []EraDesc {
	return l.eras
}

// MonthDay 农历月日描述 / lunar month-day description.
//
// 获取农历月日描述，如正月初一。此处，十一月表示为冬月，十二月表示为腊月。
func (l LunarTime) MonthDay() string {
	return l.desc
}

// LunarDesc 农历描述 / lunar descriptions.
//
// 获取农历描述，如开元二年正月初一，若无年号，则返回年份描述，如二零二五年正月初一。
func (l LunarTime) LunarDesc() []string {
	return l.innerDescWithNianHao(false, false)
}

// LunarDescWithEmperor 含君主信息的农历描述 / lunar descriptions with emperor.
//
// 获取含有君主信息的农历描述，如唐玄宗 开元二年正月初一，若无年号，则返回年份描述，如二零二五年正月初一。
// 君主信息仅供参考，多个皇帝用同一个年号的场景，此处不准
func (l LunarTime) LunarDescWithEmperor() []string {
	return l.innerDescWithNianHao(true, false)
}

// LunarDescWithDynasty 含朝代信息的农历描述 / lunar descriptions with dynasty.
//
// 获取含有朝代信息的农历描述，如唐 开元二年正月初一，若无年号，则返回年份描述，如二零二五年正月初一。
func (l LunarTime) LunarDescWithDynasty() []string {
	return l.innerDescWithNianHao(false, true)
}

// LunarDescWithDynastyAndEmperor 含朝代和君主信息的农历描述 / lunar descriptions with dynasty and emperor.
//
// 获取含有朝代和君主信息的农历描述，如唐 唐玄宗 开元二年正月初一，若无年号，则返回年份描述，如二零二五年正月初一。
// 君主信息仅供参考，多个皇帝用同一个年号的场景，此处不准
func (l LunarTime) LunarDescWithDynastyAndEmperor() []string {
	return l.innerDescWithNianHao(true, true)
}

func (l LunarTime) innerDescWithNianHao(withEmperor bool, withDynasty bool) []string {
	var res []string
	if len(l.eras) > 0 {
		for _, v := range l.eras {
			tmp := v.String() + l.desc
			if withEmperor {
				tmp = v.Emperor + " " + tmp
			}
			if withDynasty {
				tmp = v.Dynasty + " " + tmp
			}
			res = append(res, tmp)
		}
	} else {
		res = append(res, number2Chinese(l.year, true)+"年"+l.desc)
	}
	return res
}

// LunarInfo 农历结构化信息 / structured lunar information.
//
// 返回该农历结果对应的结构化农历信息切片；若存在多个并行年号，则会有多条记录。
func (l LunarTime) LunarInfo() []LunarInfo {
	var res []LunarInfo
	for _, v := range l.eras {
		li := LunarInfo{
			SolarDate:         l.solarDate,
			LunarYear:         l.year,
			LunarYearChn:      number2Chinese(l.year, true),
			LunarMonth:        l.month,
			LunarDay:          l.day,
			IsLeap:            l.leap,
			LunarMonthDayDesc: l.desc,
			GanzhiYear:        GanZhiOfYear(l.year),
			GanzhiMonth:       l.ganzhiMonth,
			GanzhiDay:         GanZhiOfDay(l.solarDate),
			Dynasty:           v.Dynasty,
			Emperor:           v.Emperor,
			Nianhao:           v.Nianhao,
			YearOfNianhao:     v.YearOfNianHao,
			EraDesc:           v.String(),
			LunarWithEraDesc:  v.String() + l.desc,
			ChineseZodiac:     l.ShengXiao(),
		}
		res = append(res, li)
	}
	if len(l.eras) == 0 {
		li := LunarInfo{
			SolarDate:         l.solarDate,
			LunarYear:         l.year,
			LunarYearChn:      number2Chinese(l.year, true),
			LunarMonth:        l.month,
			LunarDay:          l.day,
			IsLeap:            l.leap,
			LunarMonthDayDesc: l.desc,
			GanzhiYear:        GanZhiOfYear(l.year),
			GanzhiMonth:       l.ganzhiMonth,
			GanzhiDay:         GanZhiOfDay(l.solarDate),
			Dynasty:           "",
			Emperor:           "",
			Nianhao:           "",
			YearOfNianhao:     0,
			EraDesc:           number2Chinese(l.year, true) + "年",
			LunarWithEraDesc:  number2Chinese(l.year, true) + "年" + l.desc,
			ChineseZodiac:     l.ShengXiao(),
		}
		res = append(res, li)
	}
	return res
}
