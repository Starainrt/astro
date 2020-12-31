package calendar

import (
	"time"

	"github.com/starainrt/astro/basic"
)

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

// Lunar 公历转农历
// 传入 公历年月日
// 返回 农历月，日，是否闰月以及文字描述
func Lunar(year, month, day int) (int, int, bool, string) {
	return basic.GetLunar(year, month, day)
}

// ChineseLunar 公历转农历
// 传入 公历年月日
// 返回 农历月，日，是否闰月以及文字描述
// 忽略时区，日期一律按北京时间计算
func ChineseLunar(date time.Time) (int, int, bool, string) {
	return basic.GetLunar(date.Year(), int(date.Month()), date.Day())
}

// Solar 农历转公历
// 传入 农历年份，月，日，是否闰月
// 传出 公历时间
// 农历年份用公历年份代替，但是岁首需要使用农历岁首
// 例：计算己亥猪年腊月三十日对应的公历（即2020年1月24日）
// 由于农历还未到鼠年，故应当传入Solar(2019,12,30,false)
func Solar(year, month, day int, leap bool) time.Time {
	jde := basic.GetSolar(year, month, day, leap) - 8.0/24.0
	zone := time.FixedZone("CST", 8*3600)
	return basic.JDE2DateByZone(jde, zone)
}

// GanZhi 返回传入年份对应的干支
func GanZhi(year int) string {
	return basic.GetGZ(year)
}

// JieQi 返回传入年份、节气对应的北京时间节气时间
func JieQi(year, term int) time.Time {
	calcJde := basic.GetJQTime(year, term)
	zone := time.FixedZone("CST", 8*3600)
	return basic.JDE2DateByZone(calcJde, zone)
}

// WuHou 返回传入年份、物候对应的北京时间物候时间
func WuHou(year, term int) time.Time {
	calcJde := basic.GetWHTime(year, term)
	zone := time.FixedZone("CST", 8*3600)
	return basic.JDE2DateByZone(calcJde, zone)
}
