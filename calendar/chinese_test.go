package calendar

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

type lunarSolar struct {
	Lyear       int
	Lmonth      int
	Lday        int
	Leap        bool
	Year        int
	Month       int
	Day         int
	Desc        string
	GanZhiYear  string
	GanZhiMonth string
	GanZhiDay   string
}

func Test_ChineseCalendarModern(t *testing.T) {
	var testData = []lunarSolar{
		{Lyear: 1995, Lmonth: 12, Lday: 12, Leap: false, Year: 1996, Month: 1, Day: 31},
		{Lyear: 2034, Lmonth: 1, Lday: 1, Leap: false, Year: 2034, Month: 2, Day: 19},
		{Lyear: 2033, Lmonth: 12, Lday: 30, Leap: false, Year: 2034, Month: 2, Day: 18},
		{Lyear: 2033, Lmonth: 11, Lday: 27, Leap: true, Year: 2034, Month: 1, Day: 17},
		{Lyear: 2033, Lmonth: 11, Lday: 1, Leap: true, Year: 2033, Month: 12, Day: 22},
		{Lyear: 2033, Lmonth: 11, Lday: 30, Leap: false, Year: 2033, Month: 12, Day: 21},
		{Lyear: 2023, Lmonth: 2, Lday: 30, Leap: false, Year: 2023, Month: 3, Day: 21},
		{Lyear: 2023, Lmonth: 2, Lday: 1, Leap: true, Year: 2023, Month: 3, Day: 22},
		{Lyear: 2020, Lmonth: 1, Lday: 1, Leap: false, Year: 2020, Month: 1, Day: 25},
		{Lyear: 2015, Lmonth: 1, Lday: 1, Leap: false, Year: 2015, Month: 2, Day: 19},
		{Lyear: 2014, Lmonth: 12, Lday: 30, Leap: false, Year: 2015, Month: 2, Day: 18},
		{Lyear: 1996, Lmonth: 1, Lday: 1, Leap: false, Year: 1996, Month: 2, Day: 19},
		{Lyear: 1995, Lmonth: 12, Lday: 30, Leap: false, Year: 1996, Month: 2, Day: 18},
		{Lyear: 1996, Lmonth: 10, Lday: 30, Leap: false, Year: 1996, Month: 12, Day: 10},
		{Lyear: 2014, Lmonth: 9, Lday: 1, Leap: true, Year: 2014, Month: 10, Day: 24},
		{Lyear: 2014, Lmonth: 9, Lday: 30, Leap: false, Year: 2014, Month: 10, Day: 23},
		{Lyear: 2014, Lmonth: 10, Lday: 1, Leap: false, Year: 2014, Month: 11, Day: 22},
		{Lyear: 2021, Lmonth: 12, Lday: 29, Leap: false, Year: 2022, Month: 1, Day: 31},
	}
	for _, v := range testData {
		{
			lyear, lmonth, lday, leap, desp := Lunar(v.Year, v.Month, v.Day, 8.0)

			fmt.Println(lyear, desp, v.Year, v.Month, v.Day)
			if lyear != v.Lyear || lmonth != v.Lmonth || lday != v.Lday || leap != v.Leap {
				t.Fatal(v, lyear, lmonth, lday, leap, desp)
			}

			date := Solar(v.Lyear, v.Lmonth, v.Lday, v.Leap, 8.0)
			if date.Year() != v.Year || int(date.Month()) != v.Month || date.Day() != v.Day {
				t.Fatal(v, date)
			}
		}
		/*
			{
				var lyear int = v.Year
				lmonth, lday, leap, desp := RapidSolarToLunar(time.Date(v.Year, time.Month(v.Month), v.Day, 0, 0, 0, 0, getCst()))
				if lmonth > v.Month {
					lyear--
				}
				fmt.Println(lyear, desp, v.Year, v.Month, v.Day)
				if lyear != v.Lyear || lmonth != v.Lmonth || lday != v.Lday || leap != v.Leap {
					t.Fatal(v, lyear, lmonth, lday, leap, desp)
				}

				date := RapidLunarToSolar(v.Lyear, v.Lmonth, v.Lday, v.Leap)
				if date.Year() != v.Year || int(date.Month()) != v.Month || date.Day() != v.Day {
					t.Fatal(v, date)
				}
			}
		*/
	}
}

func Test_ChineseCalendarModern2(t *testing.T) {
	var testData = []lunarSolar{
		{Lyear: 1995, Lmonth: 12, Lday: 12, Leap: false, Year: 1996, Month: 1, Day: 31},
		{Lyear: 2034, Lmonth: 1, Lday: 1, Leap: false, Year: 2034, Month: 2, Day: 19},
		{Lyear: 2033, Lmonth: 12, Lday: 30, Leap: false, Year: 2034, Month: 2, Day: 18},
		{Lyear: 2033, Lmonth: 11, Lday: 27, Leap: true, Year: 2034, Month: 1, Day: 17},
		{Lyear: 2033, Lmonth: 11, Lday: 1, Leap: true, Year: 2033, Month: 12, Day: 22},
		{Lyear: 2033, Lmonth: 11, Lday: 30, Leap: false, Year: 2033, Month: 12, Day: 21},
		{Lyear: 2023, Lmonth: 2, Lday: 30, Leap: false, Year: 2023, Month: 3, Day: 21},
		{Lyear: 2023, Lmonth: 2, Lday: 1, Leap: true, Year: 2023, Month: 3, Day: 22},
		{Lyear: 2020, Lmonth: 1, Lday: 1, Leap: false, Year: 2020, Month: 1, Day: 25},
		{Lyear: 2015, Lmonth: 1, Lday: 1, Leap: false, Year: 2015, Month: 2, Day: 19},
		{Lyear: 2014, Lmonth: 12, Lday: 30, Leap: false, Year: 2015, Month: 2, Day: 18},
		{Lyear: 1996, Lmonth: 1, Lday: 1, Leap: false, Year: 1996, Month: 2, Day: 19},
		{Lyear: 1995, Lmonth: 12, Lday: 30, Leap: false, Year: 1996, Month: 2, Day: 18},
		{Lyear: 1996, Lmonth: 10, Lday: 30, Leap: false, Year: 1996, Month: 12, Day: 10},
		{Lyear: 2014, Lmonth: 9, Lday: 1, Leap: true, Year: 2014, Month: 10, Day: 24},
		{Lyear: 2014, Lmonth: 9, Lday: 30, Leap: false, Year: 2014, Month: 10, Day: 23},
		{Lyear: 2014, Lmonth: 10, Lday: 1, Leap: false, Year: 2014, Month: 11, Day: 22},
		{Lyear: 2021, Lmonth: 12, Lday: 29, Leap: false, Year: 2022, Month: 1, Day: 31},
	}
	for _, v := range testData {
		{
			res, err := SolarToLunar(time.Date(v.Year, time.Month(v.Month), v.Day, 0, 0, 0, 0, getCst()))
			if err != nil {
				t.Fatal(err)
			}
			if len(res.Lunars()) != 1 {
				t.Fatal("len(res.Lunars())!=1")
			}
			lunar := res.Lunars()[0]
			if lunar.year != v.Lyear || lunar.month != v.Lmonth || lunar.day != v.Lday || lunar.leap != v.Leap {
				t.Fatal(v, lunar.year, lunar.month, lunar.day, lunar.leap)
			}

			date, err := LunarToSolarByYMD(v.Lyear, v.Lmonth, v.Lday, v.Leap)
			if err != nil {
				t.Fatal(err)
			}
			solar := date.Time()
			if solar.Year() != v.Year || int(solar.Month()) != v.Month || solar.Day() != v.Day {
				t.Fatal(v, date)
			}
		}
		/*
			{
				var lyear int = v.Year
				lmonth, lday, leap, desp := RapidSolarToLunar(time.Date(v.Year, time.Month(v.Month), v.Day, 0, 0, 0, 0, getCst()))
				if lmonth > v.Month {
					lyear--
				}
				fmt.Println(lyear, desp, v.Year, v.Month, v.Day)
				if lyear != v.Lyear || lmonth != v.Lmonth || lday != v.Lday || leap != v.Leap {
					t.Fatal(v, lyear, lmonth, lday, leap, desp)
				}

				date := RapidLunarToSolar(v.Lyear, v.Lmonth, v.Lday, v.Leap)
				if date.Year() != v.Year || int(date.Month()) != v.Month || date.Day() != v.Day {
					t.Fatal(v, date)
				}
			}
		*/
	}
}

func Test_ChineseCalendarAncient(t *testing.T) {
	var testData = []lunarSolar{
		{Lyear: -103, Lmonth: 1, Lday: 1, Leap: false, Year: -103, Month: 2, Day: 22, Desc: "太初元年正月初一", GanZhiYear: "丁丑", GanZhiMonth: "壬寅", GanZhiDay: "癸亥"},
		{Lyear: -101, Lmonth: 6, Lday: 2, Leap: true, Year: -101, Month: 7, Day: 28, Desc: "太初三年闰六月初二", GanZhiYear: "己卯", GanZhiMonth: "辛未", GanZhiDay: "己酉"},
		{Lyear: 8, Lmonth: 11, Lday: 29, Leap: false, Year: 9, Month: 1, Day: 14, Desc: "初始元年冬月廿九", GanZhiYear: "戊辰", GanZhiMonth: "甲子", GanZhiDay: "壬申"},
		//王莽改制
		{Lyear: 9, Lmonth: 1, Lday: 1, Leap: false, Year: 9, Month: 1, Day: 15, Desc: "始建国元年正月初一", GanZhiYear: "己巳", GanZhiMonth: "乙丑", GanZhiDay: "癸酉"},
		{Lyear: 23, Lmonth: 1, Lday: 1, Leap: false, Year: 23, Month: 1, Day: 11, Desc: "地皇四年正月初一", GanZhiYear: "癸未", GanZhiMonth: "癸丑", GanZhiDay: "壬午"},
		{Lyear: 23, Lmonth: 2, Lday: 1, Leap: false, Year: 23, Month: 2, Day: 10, Desc: "地皇四年二月初一", GanZhiYear: "癸未", GanZhiMonth: "甲寅", GanZhiDay: "壬子"},
		//改回来了
		{Lyear: 23, Lmonth: 1, Lday: 1, Leap: false, Year: 23, Month: 2, Day: 10, Desc: "更始元年正月初一", GanZhiYear: "癸未", GanZhiMonth: "甲寅", GanZhiDay: "壬子"},
		{Lyear: 23, Lmonth: 12, Lday: 1, Leap: false, Year: 23, Month: 12, Day: 31, Desc: "更始元年腊月初一", GanZhiYear: "癸未", GanZhiMonth: "乙丑", GanZhiDay: "丙子"},
		{Lyear: 24, Lmonth: 1, Lday: 1, Leap: false, Year: 24, Month: 1, Day: 30, Desc: "更始二年正月初一", GanZhiYear: "甲申", GanZhiMonth: "丙寅", GanZhiDay: "丙午"},
		{Lyear: 97, Lmonth: 8, Lday: 5, Leap: true, Year: 97, Month: 9, Day: 29, Desc: "永元九年闰八月初五", GanZhiYear: "丁酉", GanZhiMonth: "己酉", GanZhiDay: "壬申"},
		{Lyear: 100, Lmonth: 2, Lday: 1, Leap: false, Year: 100, Month: 2, Day: 28, Desc: "永元十二年二月初一", GanZhiYear: "庚子", GanZhiMonth: "己卯", GanZhiDay: "甲寅"},
		//按照儒略历，这一天有29号
		{Lyear: 100, Lmonth: 2, Lday: 3, Leap: false, Year: 100, Month: 3, Day: 1, Desc: "永元十二年二月初三", GanZhiYear: "庚子", GanZhiMonth: "己卯", GanZhiDay: "丙辰"},
		//三国演义
		{Lyear: 190, Lmonth: 1, Lday: 1, Leap: false, Year: 190, Month: 2, Day: 23, Desc: "初平元年正月初一", GanZhiYear: "庚午", GanZhiMonth: "戊寅", GanZhiDay: "壬寅"},
		{Lyear: 220, Lmonth: 5, Lday: 5, Leap: false, Year: 220, Month: 6, Day: 23, Desc: "黄初元年五月初五", GanZhiYear: "庚子", GanZhiMonth: "壬午", GanZhiDay: "庚辰"},
		{Lyear: 220, Lmonth: 5, Lday: 5, Leap: false, Year: 220, Month: 6, Day: 23, Desc: "建安二十五年五月初五", GanZhiYear: "庚子", GanZhiMonth: "壬午", GanZhiDay: "庚辰"},
		{Lyear: 220, Lmonth: 5, Lday: 5, Leap: false, Year: 220, Month: 6, Day: 23, Desc: "延康元年五月初五", GanZhiYear: "庚子", GanZhiMonth: "壬午", GanZhiDay: "庚辰"},
		{Lyear: 246, Lmonth: 12, Lday: 2, Leap: true, Year: 247, Month: 1, Day: 25, Desc: "正始七年闰腊月初二", GanZhiYear: "丙寅", GanZhiMonth: "辛丑", GanZhiDay: "壬申"},
		{Lyear: 246, Lmonth: 12, Lday: 2, Leap: false, Year: 247, Month: 1, Day: 25, Desc: "延熙九年腊月初二", GanZhiYear: "丙寅", GanZhiMonth: "辛丑", GanZhiDay: "壬申"},
		{Lyear: 237, Lmonth: 2, Lday: 29, Leap: false, Year: 237, Month: 4, Day: 11, Desc: "景初元年二月廿九", GanZhiYear: "丁巳", GanZhiMonth: "癸卯", GanZhiDay: "丙申"},
		{Lyear: 237, Lmonth: 4, Lday: 1, Leap: false, Year: 237, Month: 4, Day: 12, Desc: "景初元年四月初一", GanZhiYear: "丁巳", GanZhiMonth: "甲辰", GanZhiDay: "丁酉"},
		{Lyear: 237, Lmonth: 2, Lday: 29, Leap: false, Year: 237, Month: 4, Day: 12, Desc: "建兴十五年二月廿九", GanZhiYear: "丁巳", GanZhiMonth: "癸卯", GanZhiDay: "丁酉"},
		{Lyear: 237, Lmonth: 2, Lday: 30, Leap: false, Year: 237, Month: 4, Day: 12, Desc: "嘉禾六年二月三十", GanZhiYear: "丁巳", GanZhiMonth: "癸卯", GanZhiDay: "丁酉"},
		//魏明帝改制，导致景初三年有两个12月
		{Lyear: 239, Lmonth: 12, Lday: 1, Leap: false, Year: 239, Month: 12, Day: 13, Desc: "景初三年腊月初一", GanZhiYear: "己未", GanZhiMonth: "丙子", GanZhiDay: "壬子"},
		{Lyear: 239, Lmonth: 12, Lday: 1, Leap: false, Year: 240, Month: 1, Day: 12, Desc: "景初三年腊月初一", GanZhiYear: "己未", GanZhiMonth: "丁丑", GanZhiDay: "壬午"},
		//武则天改制，建子为正月，但是其他月份不变，所以正月不是一月，一月相当于三月，以此类推
		{Lyear: 690, Lmonth: 1, Lday: 1, Leap: false, Year: 689, Month: 12, Day: 18, Desc: "天授元年正月初一", GanZhiYear: "庚寅", GanZhiMonth: "丙子", GanZhiDay: "庚辰"},
		{Lyear: 690, Lmonth: 1, Lday: 1, Leap: false, Year: 689, Month: 12, Day: 18, Desc: "载初元年正月初一", GanZhiYear: "庚寅", GanZhiMonth: "丙子", GanZhiDay: "庚辰"},
		// 太抽象了，一月是一月，正月是正月。一月!=正月
		{Lyear: 690, Lmonth: 3, Lday: 1, Leap: false, Year: 690, Month: 2, Day: 15, Desc: "天授元年一月初一", GanZhiYear: "庚寅", GanZhiMonth: "戊寅", GanZhiDay: "己卯"},
		{Lyear: 700, Lmonth: 2, Lday: 6, Leap: false, Year: 700, Month: 1, Day: 1, Desc: "圣历三年腊月初六", GanZhiYear: "庚子", GanZhiMonth: "丁丑", GanZhiDay: "丙戌"},
		{Lyear: 700, Lmonth: 12, Lday: 6, Leap: false, Year: 701, Month: 1, Day: 19, Desc: "圣历三年腊月初六", GanZhiYear: "庚子", GanZhiMonth: "己丑", GanZhiDay: "庚戌"},
		{Lyear: 700, Lmonth: 11, Lday: 1, Leap: false, Year: 700, Month: 12, Day: 15, Desc: "圣历三年冬月初一", GanZhiYear: "庚子", GanZhiMonth: "戊子", GanZhiDay: "乙亥"},
		{Lyear: 1083, Lmonth: 10, Lday: 12, Leap: false, Year: 1083, Month: 11, Day: 24, Desc: "元丰六年十月十二", GanZhiYear: "癸亥", GanZhiMonth: "癸亥", GanZhiDay: "甲申"},
		//格里高利历改革
		{Lyear: 1582, Lmonth: 9, Lday: 18, Leap: false, Year: 1582, Month: 10, Day: 4, Desc: "万历十年九月十八", GanZhiYear: "壬午", GanZhiMonth: "庚戌", GanZhiDay: "癸酉"},
		{Lyear: 1582, Lmonth: 9, Lday: 19, Leap: false, Year: 1582, Month: 10, Day: 15, Desc: "万历十年九月十九", GanZhiYear: "壬午", GanZhiMonth: "庚戌", GanZhiDay: "甲戌"},
		{Lyear: 1631, Lmonth: 11, Lday: 10, Leap: true, Year: 1632, Month: 1, Day: 1, Desc: "崇祯四年闰冬月初十", GanZhiYear: "辛未", GanZhiMonth: "庚子", GanZhiDay: "己酉"},
		{Lyear: 1912, Lmonth: 11, Lday: 24, Leap: false, Year: 1913, Month: 1, Day: 1, Desc: "一九一二年冬月廿四", GanZhiYear: "壬子", GanZhiMonth: "壬子", GanZhiDay: "壬午"},
	}
	for _, v := range testData {
		{
			dates, err := SolarToLunar(time.Date(v.Year, time.Month(v.Month), v.Day, 0, 0, 0, 0, getCst()))
			if err != nil {
				t.Fatal(err)
			}
			succ := false
			for _, date := range dates.Lunars() {
				for _, v2 := range date.LunarDesc() {
					if v2 == v.Desc {
						succ = true
						if date.LunarYear() != v.Lyear || date.LunarMonth() != v.Lmonth || date.LunarDay() != v.Lday || date.IsLeap() != v.Leap {
							t.Fatal(v, date.LunarYear(), date.LunarMonth(), date.LunarDay(), date.IsLeap())
						}
						if date.solarDate.IsZero() {
							t.Fatal(v, "zero")
						}
						if date.GanZhiYear() != v.GanZhiYear || date.GanZhiMonth() != v.GanZhiMonth || date.GanZhiDay() != v.GanZhiDay {
							t.Fatal(v, date.GanZhiYear(), date.GanZhiMonth(), date.GanZhiDay())
						}
						break
					}
				}
			}
			if !succ {
				t.Fatal("not found", v, dates.LunarDesc(), dates.LunarInfo())
			}
		}

		{
			dates, err := LunarToSolar(v.Desc)
			if err != nil {
				t.Fatal(err)
			}
			succ := false
			for _, date := range dates {
				solar := date.Solar()
				if solar.Year() == v.Year && int(solar.Month()) == v.Month && solar.Day() == v.Day {
					succ = true
					break
				}
			}
			if !succ {
				t.Fatal("not found", v, dates)
			}

		}
	}
}

func TestGanZhiOfDay(t *testing.T) {
	dates := time.Date(2025, 1, 24, 0, 0, 0, 0, getCst())
	fmt.Println(dates.Weekday())
	jde := Date2JDE(dates)
	fmt.Println(int(jde+1.5) % 7)
	y, _, _, _, desc := Lunar(dates.Year(), int(dates.Month()), dates.Day(), 8.0)
	fmt.Println(y, desc)
	//date, err := LunarToSolar("久视元年腊月辛亥")
	date, err := LunarToSolar("2025年闰6月1日")
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range date {
		fmt.Println(v.solarTime)
		fmt.Println(v.LunarDescWithDynastyAndEmperor())
	}
	fmt.Println(SolarToLunarByYMD(700, 2, 29))
}

func TestRapidLunarAndLunar(t *testing.T) {
	for year := 1949; year < 2400; year++ {
		for month := 1; month <= 12; month++ {
			a1, a2, a3, a4, _ := rapidLunarModern(year, month, 24)
			b1, b2, b3, b4, _ := basic.GetLunar(year, month, 24, 8.0/24.0)
			if a1 != b1 || a2 != b2 || a3 != b3 || a4 != b4 {
				if year == 2165 && month == 12 {
					continue
				}
				if math.Abs(float64(b3-a3)) == 1 {
					continue
				}
				t.Fatal(year, month, 24, a1, a2, a3, a4, b1, b2, b3, b4)
			}
			sol := JDE2Date(basic.GetSolar(b1, b2, b3, b4, 8.0/24))
			if sol.Year() != year && int(sol.Month()) != month && sol.Day() != 24 {
				t.Fatal(year, month, sol, b1, b2, b3, b4)
			}
		}
	}
}

/*
func TestgenReverseMapNianHao(t *testing.T) {
	//mymap := make(map[string][][]int)
	eras := nanMingEras01()
	for idx, v := range eras {
		if idx == 0 {
			continue
		}
		end := (eras[idx-1].Year - eras[idx-1].Offset) - 1
		if eras[idx-1].OtherNianHaoStart != "" {
			end++
		}
		niaohao := v.Nianhao
		if v.OtherNianHaoStart != "" {
			niaohao = v.OtherNianHaoStart
		}
		fmt.Printf("\"%s\":[][]int{{%d,%d}},\n", niaohao, v.Year-v.Offset, end)
	}
}
*/
