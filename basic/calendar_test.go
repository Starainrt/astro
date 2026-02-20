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

func TestGetJQTime(t *testing.T) {
	originalFunc := func(Year, Angle int) float64 { //节气时间
		var j int = 1
		var Day int
		var tp float64
		if Angle%2 == 0 {
			Day = 18
		} else {
			Day = 3
		}
		if Angle%10 != 0 {
			tp = float64(Angle+15.0) / 30.0
		} else {
			tp = float64(Angle) / 30.0
		}
		Month := 3 + tp
		if Month > 12 {
			Month -= 12
		}
		JD1 := JDECalc(int(Year), int(Month), float64(Day))
		if Angle == 0 {
			Angle = 360
		}
		for i := 0; i < j; i++ {
			for {
				JD0 := JD1
				stDegree := JQLospec(JD0, float64(Angle)) - float64(Angle)
				stDegreep := (JQLospec(JD0+0.000005, float64(Angle)) - JQLospec(JD0-0.000005, float64(Angle))) / 0.00001
				JD1 = JD0 - stDegree/stDegreep
				if math.Abs(JD1-JD0) <= 0.00001 {
					break
				}
			}
			JD1 -= 0.001
		}
		JD1 += 0.001
		return TD2UT(JD1, false)
	}

	// 测试数据：年份从1900-2200抽样，角度覆盖关键值
	testCases := []struct {
		year, angle int
	}{
		// 边界年份
		{1900, 0}, {1900, 15}, {1900, 30}, {1900, 45}, {1900, 90},
		{1900, 180}, {1900, 270}, {1900, 360},

		// 中间年份抽样
		{1950, 0}, {1950, 30}, {1950, 90}, {1950, 180}, {1950, 270},
		{2000, 0}, {2000, 15}, {2000, 45}, {2000, 90}, {2000, 360},
		{2023, 0}, {2023, 30}, {2023, 90}, {2023, 180}, {2023, 270},

		// 未来年份抽样
		{2100, 0}, {2100, 15}, {2100, 30}, {2100, 45}, {2100, 90},
		{2100, 180}, {2100, 270}, {2100, 360},
		{2200, 0}, {2200, 30}, {2200, 90}, {2200, 180}, {2200, 270},
	}

	// 执行测试
	allPassed := true
	for _, tc := range testCases {
		originalResult := originalFunc(tc.year, tc.angle)
		optimizedResult := GetJQTime(tc.year, tc.angle)

		diff := math.Abs(originalResult - optimizedResult)

		if diff > 1e-10 {
			t.Errorf("测试失败: year=%d, angle=%d\n原始结果: %.15f\n优化结果: %.15f\n差异: %.15f",
				tc.year, tc.angle, originalResult, optimizedResult, diff)
			allPassed = false
		} else {
			t.Logf("测试通过: year=%d, angle=%d, 结果: %.15f",
				tc.year, tc.angle, optimizedResult)
		}
	}

	if allPassed {
		t.Log("所有测试用例通过！优化函数与原始函数结果完全一致")
	}
}

func TestJQ(t *testing.T) {
	fmt.Println(GetJQTime(-721, 15))
}

func TestCal6402(t *testing.T) {
	var year = 6402
	var tz = 8.00 / 24.00
	winterSolsticeDay := GetJQTime(year, 270) + tz
	firstNewMoonDay := TD2UT(CalcMoonSHByJDE(winterSolsticeDay-15, 0), false) + tz
	nextNewMoonDay := TD2UT(CalcMoonSHByJDE(firstNewMoonDay+28, 0), false) + tz
	fmt.Println(JDE2Date(firstNewMoonDay))
	fmt.Println(JDE2Date(nextNewMoonDay))
	fmt.Println(HSunTrueLo(TD2UT(nextNewMoonDay, false)))
	fmt.Println(HMoonTrueLo(TD2UT(nextNewMoonDay, false)))
	firstNewMoonDay = normalizeTimePoint(firstNewMoonDay)
	nextNewMoonDay = normalizeTimePoint(nextNewMoonDay)
	fmt.Println(JDE2Date(winterSolsticeDay))
	//fmt.Println(JDE2Date(GetSolar(1984, 10, 2, true, 8.0/24.0)))
	fmt.Println(GetLunar(6402, 11, 24, 8.0/24.0))
	fmt.Println(GetLunar(6402, 12, 24, 8.0/24.0))
	for i := 1; i <= 12; i++ {
		fmt.Print("6403", i, "24 ---- ")
		fmt.Println(GetLunar(6403, i, 24, 8.0/24.0))
	}
	fmt.Println("-------")
	for _, v := range GetMoonLoops(float64(2132), 17) {
		fmt.Println(JDE2Date(v))
	}
	fmt.Println("-------")
	for _, v := range GetJieqiLoops(2132, 25) {
		fmt.Println(JDE2Date(v))
	}
}
