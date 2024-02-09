package basic

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"
)

func TestGenerateMagic(t *testing.T) {
	generateMagicNumber()
}
func generateMagicNumber() {
	//0月份 00000 日期 0000闰月 0000000000000 农历信息
	var tz = 8.0000 / 24.000
	yearMap := make(map[int][][]int) // {1 month,1 leap,2 29/30}
	spYear := make(map[int][]int)
	var upper []uint16
	var lower []uint16
	var full []uint32
	//var info uint32 = 0
	for year := 1899; year <= 2401; year++ {
		fmt.Println(year)
		jieqi := GetJieqiLoops(year, 25)        //一年的节气
		moon := GetMoonLoops(float64(year), 17) //一年朔月日
		winter1 := jieqi[0] - 8.0/24 + tz       //第一年冬至日
		winter2 := jieqi[24] - 8.0/24 + tz      //第二年冬至日
		for idx, v := range moon {
			if tz != 8.0/24 {
				v = v - 8.0/24 + tz
			}
			if v-math.Floor(v) > 0.5 {
				moon[idx] = math.Floor(v) + 0.5
			} else {
				moon[idx] = math.Floor(v) - 0.5
			}
		} //置闰月为0点
		for idx, v := range jieqi {
			if tz != 8.0/24 {
				v = v - 8.0/24 + tz
			}
			if v-math.Floor(v) > 0.5 {
				jieqi[idx] = math.Floor(v) + 0.5
			} else {
				jieqi[idx] = math.Floor(v) - 0.5
			}
		} //置节气为0点
		mooncount := 0           //年内朔望月计数
		var min, max int = 20, 0 //最大最小计数
		for i := 0; i < 15; i++ {
			if moon[i] >= winter1 && moon[i] < winter2 {
				if i <= min {
					min = i
				}
				if i >= max {
					max = i
				}
				mooncount++
			}
		}
		leapmonth := 20
		if mooncount == 13 { //存在闰月
			var j, i = 2, 0
			for i = min; i <= max; i++ {
				if !(moon[i] <= jieqi[j] && moon[i+1] > jieqi[j]) {
					break
				}
				j += 2
			}
			leapmonth = i - min + 1
		}
		month := 11
		for idx := min; idx <= max; idx++ {
			leap := 0
			if idx != leapmonth {
				month++
				if month > 12 {
					month -= 12
				}
			} else {
				leap = 1
			}
			if leap == 0 && month == 1 {
				cp := JDE2Date(moon[idx])
				spYear[year+1] = append(spYear[year+1], []int{int(cp.Month()), cp.Day()}...)
			}
			if idx < 6 && month > 10 {
				yearMap[year] = append(yearMap[year], []int{month, leap, int(moon[idx+1] - moon[idx])})
			} else {
				yearMap[year+1] = append(yearMap[year+1], []int{month, leap, int(moon[idx+1] - moon[idx])})
			}
		}
	}
	for year := 1900; year <= 2400; year++ {
		fmt.Println(year)
		magic := magicNumber(yearMap[year], spYear[year])
		up, low := magicNumberSpilt(magic)
		upper = append(upper, up)
		lower = append(lower, uint16(low))
		full = append(full, uint32(magic))
	}
	res := make(map[string]interface{})
	res["up"] = upper
	res["low"] = lower
	res["full"] = full
	d, _ := json.Marshal(res)
	os.WriteFile("test.json", d, 0644)
}

func magicNumber(y1 [][]int, y2 []int) int32 {
	var res int32
	res = int32(y2[1]) << 18
	if y2[0] == 2 {
		res = res | 0x800000
	}
	for idx, v := range y1 {
		if v[2] == 30 {
			res = res | (1 << (13 - idx))
		}
		if v[1] == 1 {
			res = (res & 0xFC3FFF) | int32((v[0])<<14)
		}
	}
	return res
}

func magicNumberSpilt(magic int32) (uint16, uint8) {
	var upper uint16
	var lower uint8
	lower = uint8(magic)
	upper = uint16(magic >> 8)
	return upper, lower
}
