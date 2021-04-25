package x

import (
	"fmt"
	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/calendar"
	"sort"
	"time"
)

//干支信息
type GanZhi struct {
	YGZ string `json:"ygz"`
	MGZ string `json:"mgz"`
	DGZ string `json:"dgz"`
	HGZ string `json:"hgz"`
}

func NewGanZhi(year, month, day, hour int) {
	cust := time.Date(year, time.Month(month), day, hour, 0, 0, 0, time.Local) //精确到时
	lcb := fixLiChun(year, cust)
	fmt.Println(lcb)

}

var (
	//十天干 甲1 ...癸10
	Gan = [11]string{"err", "甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
	//十二地支 子1 ...亥12
	Zhi = [13]string{"err", "子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}
)

//年干支
func GetYGZ(year, month, day, hour int) string {
	cust := time.Date(year, time.Month(month), day, hour, 0, 0, 0, time.Local) //精确到时
	lcb := fixLiChun(year, cust)
	g, z := YearGZ(year, lcb)
	return g + z
}

//传入阳历年 立春布尔值 返回年干 年支 年干支
//年干支
func YearGZ(year int, lcb bool) (string, string) {
	var aliasGan, aliasZhi string
	switch lcb {
	case false: //日期在立春之前
		//干
		g := 1 + (year+6)%10
		if g -= 1; g < 1 {
			g += 10
		}
		aliasGan = Gan[g]
		//支
		z := 1 + (year+8)%12
		if z -= 1; z < 1 {
			z += 12
		}
		aliasZhi = Zhi[z]
	case true: //日期在立春之后
		yearg := 1 + (year+6)%10
		yearz := 1 + (year+8)%12
		aliasGan = Gan[yearg]
		aliasZhi = Zhi[yearz]
	}

	return aliasGan, aliasZhi

}

//立春修正
func fixLiChun(year int, cust time.Time) bool {
	lct, _ := getJie12T(year)
	lct = time.Date(lct.Year(), lct.Month(), lct.Day(), lct.Hour(), 0, 0, 0, time.Local)
	var b bool
	if cust.Equal(lct) || cust.After(lct) {
		b = true //当前时间在立春之后
	} else {
		b = false //当前时间在立春之前
	}
	return b
}

//传入阳历年数字 返回本年立春阳历时间戳 12节时间戳数组
//获取本年立春时间戳
func getJie12T(year int) (time.Time, []time.Time) {

	year -= 1 //k:1-->上一年冬至时间 k:25-->本年冬至时间 k:4--本年立春
	jq := basic.GetOneYearJQ(year)
	var keys []int
	for k := range jq {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	//k:1上一年冬至...k4:立春... k:25本年冬至
	/*
		"冬至", "小寒", "大寒", "立春", "雨水", "惊蛰",
		"春分", "清明", "谷雨", "立夏", "小满", "芒种",
		"夏至", "小暑", "大暑", "立秋", "处暑", "白露",
		"秋分", "寒露", "霜降", "立冬", "小雪", "大雪", "冬至",
	*/
	var jieArr []time.Time //12节
	var lct time.Time
	for _, v := range keys {
		//fmt.Printf("k:%v -本地时区%v\n", v, calendar.JDE2Date(jq[v]))
		if v%2 == 0 {
			jieArr = append(jieArr, calendar.JDE2Date(jq[v]))
		}
		if v == 4 {
			lct = calendar.JDE2Date(jq[v])
		}
	}

	//12节
	// 小寒  立春  惊蛰  清明  立夏  芒种  小暑  立秋  白露  寒露  立冬  大雪
	//排序后对应的k值 2 4 6 8 10 12 14 16 18 20 22 24
	return lct, jieArr
}
