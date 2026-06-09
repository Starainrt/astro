package calendar

import (
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

type solarYMD struct {
	year  int
	month int
	day   int
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

func Test_ChineseCalendarQinHan(t *testing.T) {
	testData := []lunarSolar{
		{Lyear: -130, Lmonth: 10, Lday: 1, Leap: false, Year: -131, Month: 11, Day: 25, Desc: "十月初一", GanZhiDay: "壬申"},
		{Lyear: -130, Lmonth: 11, Lday: 1, Leap: false, Year: -131, Month: 12, Day: 24, Desc: "十一月初一", GanZhiDay: "辛丑"},
		{Lyear: -130, Lmonth: 12, Lday: 1, Leap: false, Year: -130, Month: 1, Day: 23, Desc: "十二月初一", GanZhiDay: "辛未"},
		{Lyear: -130, Lmonth: 1, Lday: 1, Leap: false, Year: -130, Month: 2, Day: 21, Desc: "正月初一", GanZhiDay: "庚子"},
		{Lyear: -130, Lmonth: 9, Lday: 1, Leap: false, Year: -130, Month: 10, Day: 15, Desc: "九月初一", GanZhiDay: "丙申"},
		{Lyear: -201, Lmonth: 10, Lday: 1, Leap: false, Year: -202, Month: 10, Day: 31, Desc: "十月初一", GanZhiDay: "甲午"},
		{Lyear: -201, Lmonth: 1, Lday: 1, Leap: false, Year: -201, Month: 1, Day: 28, Desc: "正月初一", GanZhiDay: "癸亥"},
		{Lyear: -201, Lmonth: 9, Lday: 1, Leap: true, Year: -201, Month: 10, Day: 20, Desc: "后九月初一", GanZhiDay: "戊子"},
		// -104 的秦汉颛顼历日期与后续查表历存在重叠，秦汉语义用显式历法验证。
		{Lyear: -104, Lmonth: 10, Lday: 1, Leap: false, Year: -105, Month: 11, Day: 8, Desc: "十月初一"},
	}
	for _, v := range testData {
		res, err := SolarToLunarByYMD(v.Year, v.Month, v.Day)
		if err != nil {
			t.Fatal(v, err)
		}
		lunar := res.Lunar()
		if lunar.LunarYear() != v.Lyear || lunar.LunarMonth() != v.Lmonth || lunar.LunarDay() != v.Lday || lunar.IsLeap() != v.Leap {
			t.Fatal(v, lunar.LunarYear(), lunar.LunarMonth(), lunar.LunarDay(), lunar.IsLeap())
		}
		if lunar.MonthDay() != v.Desc {
			t.Fatal(v, lunar.MonthDay())
		}
		if v.GanZhiDay != "" && lunar.GanZhiDay() != v.GanZhiDay {
			t.Fatal(v, lunar.GanZhiDay())
		}
		if lunar.GanZhiMonth() != "" {
			t.Fatal(v, lunar.GanZhiMonth())
		}
		if lunar.CalendarSystem() != AncientCalendarQinHan || lunar.CalendarName() != ancientCalendarName(AncientCalendarQinHan) {
			t.Fatal(v, lunar.CalendarSystem(), lunar.CalendarName())
		}
		infos := res.LunarInfo()
		if len(infos) != 1 || infos[0].CalendarSystem != AncientCalendarQinHan || infos[0].CalendarName != ancientCalendarName(AncientCalendarQinHan) {
			t.Fatal(v, infos)
		}

		date, err := LunarToSolarByYMDWithCalendar(v.Lyear, v.Lmonth, v.Lday, v.Leap, AncientCalendarQinHan)
		if err != nil {
			t.Fatal(v, err)
		}
		solar := date.Time()
		if solar.Year() != v.Year || int(solar.Month()) != v.Month || solar.Day() != v.Day {
			t.Fatal(v, solar)
		}
	}
}

func Test_ChineseCalendarQinHanHandoffToHanQing(t *testing.T) {
	lastQinHan, err := SolarToLunarByYMD(-104, 11, 25)
	if err != nil {
		t.Fatal(err)
	}
	lastLunar := lastQinHan.Lunar()
	if lastLunar.LunarYear() != -104 || lastLunar.LunarMonth() != 9 || lastLunar.LunarDay() != 30 || !lastLunar.IsLeap() || lastLunar.CalendarSystem() != AncientCalendarQinHan {
		t.Fatalf("unexpected last QinHan day: y=%d m=%d d=%d leap=%v system=%q",
			lastLunar.LunarYear(), lastLunar.LunarMonth(), lastLunar.LunarDay(), lastLunar.IsLeap(), lastLunar.CalendarSystem())
	}

	after, err := SolarToLunarByYMD(-104, 12, 1)
	if err != nil {
		t.Fatal(err)
	}
	afterLunar := after.Lunar()
	if afterLunar.LunarYear() != -104 || afterLunar.LunarMonth() != 10 || afterLunar.LunarDay() != 6 || afterLunar.IsLeap() || afterLunar.CalendarSystem() == AncientCalendarQinHan {
		t.Fatalf("unexpected HanQing handoff day: y=%d m=%d d=%d leap=%v system=%q",
			afterLunar.LunarYear(), afterLunar.LunarMonth(), afterLunar.LunarDay(), afterLunar.IsLeap(), afterLunar.CalendarSystem())
	}

	roundtrip, err := LunarToSolarByYMD(afterLunar.LunarYear(), afterLunar.LunarMonth(), afterLunar.LunarDay(), afterLunar.IsLeap())
	if err != nil {
		t.Fatal(err)
	}
	if roundtrip.Solar().Year() != -104 || int(roundtrip.Solar().Month()) != 12 || roundtrip.Solar().Day() != 1 {
		t.Fatal(roundtrip.Solar())
	}

	parsed, err := LunarToSolar("-104年十月初六")
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed) != 1 || parsed[0].Solar().Year() != -104 || int(parsed[0].Solar().Month()) != 12 || parsed[0].Solar().Day() != 1 {
		t.Fatal(parsed)
	}
}

func Test_ChineseCalendarQinHanWithCalendarPreservesTime(t *testing.T) {
	input := time.Date(-200, time.January, 17, 13, 14, 15, 123, getCst())
	result, err := SolarToLunarWithCalendar(input, AncientCalendarQinHan)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Solar().Equal(input) {
		t.Fatalf("solar time mismatch: got %s want %s", result.Solar(), input)
	}
	lunar := result.Lunar()
	if lunar.CalendarSystem() != AncientCalendarQinHan {
		t.Fatal(lunar.CalendarSystem())
	}
	infos := result.LunarInfo()
	if len(infos) != 1 || !infos[0].SolarDate.Equal(input) {
		t.Fatalf("lunar info solar date mismatch: %#v", infos)
	}

	byYMD, err := SolarToLunarByYMDWithCalendar(-200, 1, 17, AncientCalendarQinHan)
	if err != nil {
		t.Fatal(err)
	}
	if byYMD.Solar().Hour() != 0 || byYMD.Solar().Minute() != 0 || byYMD.Solar().Second() != 0 || byYMD.Solar().Nanosecond() != 0 {
		t.Fatalf("expected YMD route to keep midnight, got %s", byYMD.Solar())
	}
}

func Test_ChineseCalendarQinHanEveryFiveYears(t *testing.T) {
	monthOrder := []struct {
		month int
		leap  bool
	}{
		{10, false},
		{11, false},
		{12, false},
		{1, false},
		{2, false},
		{3, false},
		{4, false},
		{5, false},
		{6, false},
		{7, false},
		{8, false},
		{9, false},
		{9, true},
	}
	testData := []struct {
		lunarYear int
		starts    []solarYMD
	}{
		{lunarYear: -220, starts: []solarYMD{{-221, 10, 31}, {-221, 11, 30}, {-221, 12, 29}, {-220, 1, 28}, {-220, 2, 27}, {-220, 3, 27}, {-220, 4, 26}, {-220, 5, 25}, {-220, 6, 24}, {-220, 7, 23}, {-220, 8, 22}, {-220, 9, 20}, {-220, 10, 20}}},
		{lunarYear: -215, starts: []solarYMD{{-216, 11, 4}, {-216, 12, 4}, {-215, 1, 2}, {-215, 2, 1}, {-215, 3, 2}, {-215, 4, 1}, {-215, 5, 1}, {-215, 5, 30}, {-215, 6, 29}, {-215, 7, 28}, {-215, 8, 27}, {-215, 9, 25}, {-215, 10, 25}}},
		{lunarYear: -210, starts: []solarYMD{{-211, 11, 9}, {-211, 12, 9}, {-210, 1, 7}, {-210, 2, 6}, {-210, 3, 7}, {-210, 4, 6}, {-210, 5, 5}, {-210, 6, 4}, {-210, 7, 3}, {-210, 8, 2}, {-210, 9, 1}, {-210, 9, 30}}},
		{lunarYear: -205, starts: []solarYMD{{-206, 11, 14}, {-206, 12, 14}, {-205, 1, 12}, {-205, 2, 11}, {-205, 3, 12}, {-205, 4, 11}, {-205, 5, 10}, {-205, 6, 9}, {-205, 7, 8}, {-205, 8, 7}, {-205, 9, 5}, {-205, 10, 5}}},
		{lunarYear: -200, starts: []solarYMD{{-201, 11, 19}, {-201, 12, 18}, {-200, 1, 17}, {-200, 2, 15}, {-200, 3, 16}, {-200, 4, 15}, {-200, 5, 14}, {-200, 6, 13}, {-200, 7, 12}, {-200, 8, 11}, {-200, 9, 9}, {-200, 10, 9}}},
		{lunarYear: -195, starts: []solarYMD{{-196, 11, 23}, {-196, 12, 22}, {-195, 1, 21}, {-195, 2, 19}, {-195, 3, 21}, {-195, 4, 19}, {-195, 5, 19}, {-195, 6, 18}, {-195, 7, 17}, {-195, 8, 16}, {-195, 9, 14}, {-195, 10, 14}}},
		{lunarYear: -190, starts: []solarYMD{{-191, 10, 29}, {-191, 11, 28}, {-191, 12, 27}, {-190, 1, 26}, {-190, 2, 24}, {-190, 3, 26}, {-190, 4, 24}, {-190, 5, 24}, {-190, 6, 22}, {-190, 7, 22}, {-190, 8, 21}, {-190, 9, 19}, {-190, 10, 19}}},
		{lunarYear: -185, starts: []solarYMD{{-186, 11, 3}, {-186, 12, 3}, {-185, 1, 1}, {-185, 1, 31}, {-185, 3, 1}, {-185, 3, 31}, {-185, 4, 29}, {-185, 5, 29}, {-185, 6, 27}, {-185, 7, 27}, {-185, 8, 25}, {-185, 9, 24}, {-185, 10, 23}}},
		{lunarYear: -180, starts: []solarYMD{{-181, 11, 8}, {-181, 12, 8}, {-180, 1, 6}, {-180, 2, 5}, {-180, 3, 5}, {-180, 4, 4}, {-180, 5, 3}, {-180, 6, 2}, {-180, 7, 1}, {-180, 7, 31}, {-180, 8, 29}, {-180, 9, 28}}},
		{lunarYear: -175, starts: []solarYMD{{-176, 11, 12}, {-176, 12, 11}, {-175, 1, 10}, {-175, 2, 9}, {-175, 3, 10}, {-175, 4, 9}, {-175, 5, 8}, {-175, 6, 7}, {-175, 7, 6}, {-175, 8, 5}, {-175, 9, 3}, {-175, 10, 3}}},
		{lunarYear: -170, starts: []solarYMD{{-171, 11, 17}, {-171, 12, 16}, {-170, 1, 15}, {-170, 2, 13}, {-170, 3, 15}, {-170, 4, 14}, {-170, 5, 13}, {-170, 6, 12}, {-170, 7, 11}, {-170, 8, 10}, {-170, 9, 8}, {-170, 10, 8}}},
		{lunarYear: -165, starts: []solarYMD{{-166, 11, 22}, {-166, 12, 21}, {-165, 1, 20}, {-165, 2, 18}, {-165, 3, 20}, {-165, 4, 18}, {-165, 5, 18}, {-165, 6, 16}, {-165, 7, 16}, {-165, 8, 15}, {-165, 9, 13}, {-165, 10, 13}}},
		{lunarYear: -160, starts: []solarYMD{{-161, 11, 27}, {-161, 12, 26}, {-160, 1, 25}, {-160, 2, 23}, {-160, 3, 24}, {-160, 4, 22}, {-160, 5, 22}, {-160, 6, 20}, {-160, 7, 20}, {-160, 8, 18}, {-160, 9, 17}, {-160, 10, 16}}},
		{lunarYear: -155, starts: []solarYMD{{-156, 11, 1}, {-156, 12, 1}, {-156, 12, 30}, {-155, 1, 29}, {-155, 2, 27}, {-155, 3, 29}, {-155, 4, 27}, {-155, 5, 27}, {-155, 6, 25}, {-155, 7, 25}, {-155, 8, 23}, {-155, 9, 22}, {-155, 10, 21}}},
		{lunarYear: -150, starts: []solarYMD{{-151, 11, 6}, {-151, 12, 5}, {-150, 1, 4}, {-150, 2, 3}, {-150, 3, 4}, {-150, 4, 3}, {-150, 5, 2}, {-150, 6, 1}, {-150, 6, 30}, {-150, 7, 30}, {-150, 8, 28}, {-150, 9, 27}, {-150, 10, 26}}},
		{lunarYear: -145, starts: []solarYMD{{-146, 11, 11}, {-146, 12, 10}, {-145, 1, 9}, {-145, 2, 7}, {-145, 3, 9}, {-145, 4, 8}, {-145, 5, 7}, {-145, 6, 6}, {-145, 7, 5}, {-145, 8, 4}, {-145, 9, 2}, {-145, 10, 2}}},
		{lunarYear: -140, starts: []solarYMD{{-141, 11, 16}, {-141, 12, 15}, {-140, 1, 14}, {-140, 2, 12}, {-140, 3, 13}, {-140, 4, 11}, {-140, 5, 11}, {-140, 6, 9}, {-140, 7, 9}, {-140, 8, 8}, {-140, 9, 6}, {-140, 10, 6}}},
		{lunarYear: -135, starts: []solarYMD{{-136, 11, 20}, {-136, 12, 19}, {-135, 1, 18}, {-135, 2, 16}, {-135, 3, 18}, {-135, 4, 16}, {-135, 5, 16}, {-135, 6, 14}, {-135, 7, 14}, {-135, 8, 12}, {-135, 9, 11}, {-135, 10, 11}}},
		{lunarYear: -130, starts: []solarYMD{{-131, 11, 25}, {-131, 12, 24}, {-130, 1, 23}, {-130, 2, 21}, {-130, 3, 23}, {-130, 4, 21}, {-130, 5, 21}, {-130, 6, 19}, {-130, 7, 19}, {-130, 8, 17}, {-130, 9, 16}, {-130, 10, 15}}},
		{lunarYear: -125, starts: []solarYMD{{-126, 10, 31}, {-126, 11, 30}, {-126, 12, 29}, {-125, 1, 28}, {-125, 2, 26}, {-125, 3, 28}, {-125, 4, 26}, {-125, 5, 26}, {-125, 6, 24}, {-125, 7, 24}, {-125, 8, 22}, {-125, 9, 21}, {-125, 10, 20}}},
		{lunarYear: -120, starts: []solarYMD{{-121, 11, 5}, {-121, 12, 4}, {-120, 1, 3}, {-120, 2, 1}, {-120, 3, 2}, {-120, 4, 1}, {-120, 4, 30}, {-120, 5, 30}, {-120, 6, 28}, {-120, 7, 28}, {-120, 8, 26}, {-120, 9, 25}, {-120, 10, 24}}},
		{lunarYear: -115, starts: []solarYMD{{-116, 11, 9}, {-116, 12, 8}, {-115, 1, 7}, {-115, 2, 5}, {-115, 3, 7}, {-115, 4, 5}, {-115, 5, 5}, {-115, 6, 4}, {-115, 7, 3}, {-115, 8, 2}, {-115, 8, 31}, {-115, 9, 30}}},
		{lunarYear: -110, starts: []solarYMD{{-111, 11, 14}, {-111, 12, 13}, {-110, 1, 12}, {-110, 2, 10}, {-110, 3, 12}, {-110, 4, 10}, {-110, 5, 10}, {-110, 6, 8}, {-110, 7, 8}, {-110, 8, 6}, {-110, 9, 5}, {-110, 10, 5}}},
		{lunarYear: -105, starts: []solarYMD{{-106, 11, 19}, {-106, 12, 18}, {-105, 1, 17}, {-105, 2, 15}, {-105, 3, 17}, {-105, 4, 15}, {-105, 5, 15}, {-105, 6, 13}, {-105, 7, 13}, {-105, 8, 11}, {-105, 9, 10}, {-105, 10, 9}}},
		{lunarYear: -104, starts: []solarYMD{{-105, 11, 8}, {-105, 12, 8}, {-104, 1, 6}, {-104, 2, 5}, {-104, 3, 5}, {-104, 4, 4}, {-104, 5, 3}, {-104, 6, 2}, {-104, 7, 1}, {-104, 7, 31}, {-104, 8, 29}, {-104, 9, 28}, {-104, 10, 27}}},
	}
	for _, tc := range testData {
		if len(tc.starts) < 12 || len(tc.starts) > len(monthOrder) {
			t.Fatal(tc.lunarYear, len(tc.starts))
		}
		for i, start := range tc.starts {
			expectedMonth := monthOrder[i]
			res, err := SolarToLunarByYMD(start.year, start.month, start.day)
			if err != nil {
				t.Fatal(tc.lunarYear, start, err)
			}
			lunar := res.Lunar()
			if lunar.LunarYear() != tc.lunarYear || lunar.LunarMonth() != expectedMonth.month || lunar.LunarDay() != 1 || lunar.IsLeap() != expectedMonth.leap {
				t.Fatal(tc.lunarYear, start, lunar.LunarYear(), lunar.LunarMonth(), lunar.LunarDay(), lunar.IsLeap())
			}
			if expectedMonth.leap && (lunar.LunarMonth() != 9 || !lunar.IsLeap() || lunar.MonthDay() != "后九月初一") {
				t.Fatal(tc.lunarYear, start, lunar.LunarMonth(), lunar.IsLeap(), lunar.MonthDay())
			}

			solar, err := LunarToSolarByYMDWithCalendar(tc.lunarYear, expectedMonth.month, 1, expectedMonth.leap, AncientCalendarQinHan)
			if err != nil {
				t.Fatal(tc.lunarYear, expectedMonth, err)
			}
			if solar.Time().Year() != start.year || int(solar.Time().Month()) != start.month || solar.Time().Day() != start.day {
				t.Fatal(tc.lunarYear, expectedMonth, solar.Time(), start)
			}
		}
	}
}

func Test_ChineseCalendarQinHanHouJiuYueParse(t *testing.T) {
	testData := []struct {
		desc  string
		year  int
		month int
		day   int
	}{
		{desc: "-201年后九月初一", year: -201, month: 10, day: 20},
		{desc: "-201年後九月初一", year: -201, month: 10, day: 20},
		{desc: "-104年后九月初一", year: -104, month: 10, day: 27},
		{desc: "-104年後九月初一", year: -104, month: 10, day: 27},
	}
	for _, tc := range testData {
		results, err := LunarToSolar(tc.desc)
		if err != nil {
			t.Fatal(tc.desc, err)
		}
		if len(results) != 1 {
			t.Fatal(tc.desc, len(results))
		}
		solar := results[0].Time()
		lunar := results[0].Lunar()
		if solar.Year() != tc.year || int(solar.Month()) != tc.month || solar.Day() != tc.day {
			t.Fatal(tc.desc, solar)
		}
		if lunar.LunarYear() != tc.year || lunar.LunarMonth() != 9 || lunar.LunarDay() != 1 || !lunar.IsLeap() {
			t.Fatal(tc.desc, lunar.LunarYear(), lunar.LunarMonth(), lunar.LunarDay(), lunar.IsLeap())
		}
		if lunar.MonthDay() != "后九月初一" {
			t.Fatal(tc.desc, lunar.MonthDay())
		}
		if lunar.CalendarSystem() != AncientCalendarQinHan {
			t.Fatal(tc.desc, lunar.CalendarSystem())
		}
	}
	for _, desc := range []string{"2020年后四月初一", "2020年后九月初一", "元丰六年后九月初一"} {
		if _, err := LunarToSolar(desc); err == nil {
			t.Fatal("expected invalid hou month to be rejected:", desc)
		}
	}
	if _, err := LunarToSolarWithCalendar("-250年后九月初一", AncientCalendarZhou); err == nil {
		t.Fatal("expected explicit Zhou calendar to reject hou month")
	}
}

func Test_ChineseCalendarNegativeGanZhiDayIndex(t *testing.T) {
	lunar, err := SolarToLunarByYMD(-201, 1, 28)
	if err != nil {
		t.Fatal(err)
	}
	if got := lunar.Lunar().GanZhiDay(); got != "癸亥" {
		t.Fatalf("unexpected gan zhi day: got %q want %q", got, "癸亥")
	}
	if got := GanZhiOfDay(time.Date(-201, time.January, 28, 0, 0, 0, 0, getCst())); got != "癸亥" {
		t.Fatalf("unexpected direct gan zhi day: got %q want %q", got, "癸亥")
	}
}

func Test_ChineseCalendarCalendricalJieQi(t *testing.T) {
	testData := []struct {
		name   string
		year   int
		term   int
		system AncientCalendarSystem
		want   solarYMD
	}{
		{name: "qin han xiaoxue", year: -202, term: JQ_小雪, system: AncientCalendarQinHan, want: solarYMD{-202, 11, 24}},
		{name: "qin han dongzhi", year: -202, term: JQ_冬至, system: AncientCalendarQinHan, want: solarYMD{-202, 12, 25}},
		{name: "qin han xiazhi", year: -201, term: JQ_夏至, system: AncientCalendarQinHan, want: solarYMD{-201, 6, 25}},
		{name: "zhou dongzhi", year: -387, term: JQ_冬至, system: AncientCalendarZhou, want: solarYMD{-387, 12, 25}},
		{name: "default han qing xiaohan", year: -103, term: JQ_小寒, system: AncientCalendarDefault, want: solarYMD{-103, 1, 9}},
		{name: "default han qing lichun", year: -103, term: JQ_立春, system: AncientCalendarDefault, want: solarYMD{-103, 2, 8}},
		{name: "default han qing dongzhi", year: -103, term: JQ_冬至, system: AncientCalendarDefault, want: solarYMD{-103, 12, 25}},
		{name: "default han qing exception", year: 445, term: JQ_立春, system: AncientCalendarDefault, want: solarYMD{445, 2, 3}},
		{name: "default han qing cross row", year: 1582, term: JQ_小寒, system: AncientCalendarDefault, want: solarYMD{1581, 12, 27}},
		{name: "default han qing gregorian handoff", year: 1582, term: JQ_冬至, system: AncientCalendarDefault, want: solarYMD{1582, 12, 22}},
		{name: "default han qing upper xiaohan", year: 1912, term: JQ_小寒, system: AncientCalendarDefault, want: solarYMD{1912, 1, 7}},
		{name: "default han qing upper dongzhi", year: 1912, term: JQ_冬至, system: AncientCalendarDefault, want: solarYMD{1912, 12, 22}},
	}
	for _, tc := range testData {
		t.Run(tc.name, func(t *testing.T) {
			got, err := CalendricalJieQiWithCalendar(tc.year, tc.term, tc.system)
			if err != nil {
				t.Fatal(err)
			}
			assertCalendricalJieQiDate(t, got, tc.want)
		})
	}

	got, err := CalendricalJieQi(-202, JQ_冬至)
	if err != nil {
		t.Fatal(err)
	}
	assertCalendricalJieQiDate(t, got, solarYMD{-202, 12, 25})

	earlyDefault, err := CalendricalJieQi(-221, JQ_霜降)
	if err != nil {
		t.Fatal(err)
	}
	earlyZhou, err := CalendricalJieQiWithCalendar(-221, JQ_霜降, AncientCalendarZhou)
	if err != nil {
		t.Fatal(err)
	}
	if !earlyDefault.Equal(earlyZhou) || !earlyDefault.Before(qinHanStartDate()) {
		t.Fatalf("unexpected default -221 pre-transition term: default=%s zhou=%s", earlyDefault, earlyZhou)
	}

	lateDefault, err := CalendricalJieQi(-221, JQ_立冬)
	if err != nil {
		t.Fatal(err)
	}
	lateQinHan, err := CalendricalJieQiWithCalendar(-221, JQ_立冬, AncientCalendarQinHan)
	if err != nil {
		t.Fatal(err)
	}
	if !lateDefault.Equal(lateQinHan) || lateDefault.Before(qinHanStartDate()) {
		t.Fatalf("unexpected default -221 post-transition term: default=%s qinHan=%s", lateDefault, lateQinHan)
	}
}

func Test_ChineseCalendarCalendricalJieQiBoundaries(t *testing.T) {
	if _, err := CalendricalJieQiWithCalendar(-104, JQ_春分, AncientCalendarQinHan); err != nil {
		t.Fatal(err)
	}
	if _, err := CalendricalJieQiWithCalendar(-500, JQ_冬至, AncientCalendarChunqiu); err == nil {
		t.Fatal("expected Chunqiu calendrical solar terms to be unsupported")
	}
	if _, err := CalendricalJieQiWithCalendar(-221, JQ_霜降, AncientCalendarQinHan); err == nil {
		t.Fatal("expected explicit QinHan solar terms before adoption to be rejected")
	}
	if _, err := CalendricalJieQiWithCalendar(-103, JQ_冬至, AncientCalendarQinHan); err == nil {
		t.Fatal("expected explicit QinHan solar terms after range to be rejected")
	}
	if _, err := CalendricalJieQiWithCalendar(2026, JQ_冬至, AncientCalendarZhou); err == nil {
		t.Fatal("expected explicit ancient calendar to reject modern solar-term year")
	}
	got, err := CalendricalJieQi(-103, JQ_冬至)
	if err != nil {
		t.Fatal(err)
	}
	assertCalendricalJieQiDate(t, got, solarYMD{-103, 12, 25})
	if _, err := CalendricalJieQi(1913, JQ_冬至); err == nil {
		t.Fatal("expected default calendrical solar terms to reject years after table")
	}
	if _, err := CalendricalJieQi(-202, 7); err == nil {
		t.Fatal("expected invalid solar-term angle to be rejected")
	}
}

func assertCalendricalJieQiDate(t *testing.T, got time.Time, want solarYMD) {
	t.Helper()
	if got.Year() != want.year || int(got.Month()) != want.month || got.Day() != want.day {
		t.Fatalf("date mismatch: got %04d-%02d-%02d want %04d-%02d-%02d",
			got.Year(), got.Month(), got.Day(), want.year, want.month, want.day)
	}
	if got.Hour() != 0 || got.Minute() != 0 || got.Second() != 0 || got.Nanosecond() != 0 {
		t.Fatalf("expected midnight, got %s", got)
	}
	if _, offset := got.Zone(); offset != 8*3600 {
		t.Fatalf("expected UTC+8, got %s", got)
	}
}

func Test_ChineseCalendarAncientNegativeYearDescRoundtrip(t *testing.T) {
	res, err := SolarToLunarByYMD(-251, 11, 30)
	if err != nil {
		t.Fatal(err)
	}
	descs := res.LunarDesc()
	if len(descs) != 1 || descs[0] != "负二五零年正月初一" {
		t.Fatalf("unexpected descs: %v", descs)
	}
	for _, desc := range []string{descs[0], "負二五零年正月初一"} {
		results, err := LunarToSolar(desc)
		if err != nil {
			t.Fatal(desc, err)
		}
		if len(results) != 1 {
			t.Fatal(desc, len(results))
		}
		solar := results[0].Solar()
		if solar.Year() != -251 || int(solar.Month()) != 11 || solar.Day() != 30 {
			t.Fatal(desc, solar)
		}
		lunar := results[0].Lunar()
		if lunar.LunarYear() != -250 || lunar.LunarMonth() != 1 || lunar.LunarDay() != 1 || lunar.IsLeap() {
			t.Fatal(desc, lunar.LunarYear(), lunar.LunarMonth(), lunar.LunarDay(), lunar.IsLeap())
		}
	}
}

func Test_ChineseCalendarAncientPreQin(t *testing.T) {
	testData := []struct {
		name   string
		system AncientCalendarSystem
		lyear  int
		lmonth int
		lday   int
		leap   bool
		year   int
		month  int
		day    int
		desc   string
	}{
		{name: "default zhou", system: AncientCalendarDefault, lyear: -250, lmonth: 1, lday: 1, year: -251, month: 11, day: 30, desc: "正月初一"},
		{name: "zhou", system: AncientCalendarZhou, lyear: -250, lmonth: 1, lday: 1, year: -251, month: 11, day: 30, desc: "正月初一"},
		{name: "lu", system: AncientCalendarLu, lyear: -250, lmonth: 1, lday: 1, year: -251, month: 12, day: 1, desc: "正月初一"},
		{name: "yin", system: AncientCalendarYin, lyear: -250, lmonth: 1, lday: 1, year: -250, month: 1, day: 29, desc: "正月初一"},
		{name: "zhuanxu", system: AncientCalendarZhuanxu, lyear: -250, lmonth: 10, lday: 1, year: -251, month: 11, day: 1, desc: "十月初一"},
		{name: "chunqiu", system: AncientCalendarChunqiu, lyear: -500, lmonth: 1, lday: 1, year: -501, month: 12, day: 5, desc: "正月初一"},
	}
	for _, tc := range testData {
		t.Run(tc.name, func(t *testing.T) {
			var res Time
			var err error
			if tc.system == AncientCalendarDefault {
				res, err = LunarToSolarByYMD(tc.lyear, tc.lmonth, tc.lday, tc.leap)
			} else {
				res, err = LunarToSolarByYMDWithCalendar(tc.lyear, tc.lmonth, tc.lday, tc.leap, tc.system)
			}
			if err != nil {
				t.Fatal(err)
			}
			if res.Solar().Year() != tc.year || int(res.Solar().Month()) != tc.month || res.Solar().Day() != tc.day {
				t.Fatalf("solar mismatch: got %04d-%02d-%02d want %04d-%02d-%02d",
					res.Solar().Year(), res.Solar().Month(), res.Solar().Day(), tc.year, tc.month, tc.day)
			}
			lunar := res.Lunar()
			if lunar.LunarYear() != tc.lyear || lunar.LunarMonth() != tc.lmonth || lunar.LunarDay() != tc.lday || lunar.IsLeap() != tc.leap {
				t.Fatalf("lunar mismatch: got y=%d m=%d d=%d leap=%v", lunar.LunarYear(), lunar.LunarMonth(), lunar.LunarDay(), lunar.IsLeap())
			}
			if lunar.MonthDay() != tc.desc {
				t.Fatalf("desc mismatch: got %q want %q", lunar.MonthDay(), tc.desc)
			}
			if lunar.GanZhiMonth() != "" {
				t.Fatalf("unexpected ancient ganzhi month: %q", lunar.GanZhiMonth())
			}
			if tc.system != AncientCalendarDefault && lunar.CalendarSystem() != tc.system {
				t.Fatalf("system mismatch: got %q want %q", lunar.CalendarSystem(), tc.system)
			}
			infos := res.LunarInfo()
			if len(infos) != 1 || infos[0].CalendarSystem != lunar.CalendarSystem() || infos[0].CalendarName != lunar.CalendarName() {
				t.Fatalf("lunar info calendar mismatch: %#v", infos)
			}

			var back Time
			if tc.system == AncientCalendarDefault {
				back, err = SolarToLunarByYMD(tc.year, tc.month, tc.day)
			} else {
				back, err = SolarToLunarByYMDWithCalendar(tc.year, tc.month, tc.day, tc.system)
			}
			if err != nil {
				t.Fatal(err)
			}
			backLunar := back.Lunar()
			if backLunar.LunarYear() != tc.lyear || backLunar.LunarMonth() != tc.lmonth || backLunar.LunarDay() != tc.lday || backLunar.IsLeap() != tc.leap {
				t.Fatalf("roundtrip lunar mismatch: got y=%d m=%d d=%d leap=%v", backLunar.LunarYear(), backLunar.LunarMonth(), backLunar.LunarDay(), backLunar.IsLeap())
			}
		})
	}
}

func Test_ChineseCalendarAncientWithCalendarBoundaries(t *testing.T) {
	if _, err := SolarToLunarByYMDWithCalendar(2026, 1, 1, AncientCalendarZhou); err == nil {
		t.Fatal("expected explicit ancient calendar to reject modern year")
	}
	if _, err := SolarToLunarByYMD(-722, 1, 1); err == nil {
		t.Fatal("expected default pre-Qin route to reject years before -721")
	}
	lower, err := SolarToLunarByYMD(-721, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	lowerLunar := lower.Lunar()
	if lowerLunar.LunarYear() != -722 || lowerLunar.LunarMonth() != 12 || lowerLunar.LunarDay() != 16 || lowerLunar.CalendarSystem() != AncientCalendarChunqiu {
		t.Fatalf("unexpected -721 lower boundary lunar: y=%d m=%d d=%d system=%q",
			lowerLunar.LunarYear(), lowerLunar.LunarMonth(), lowerLunar.LunarDay(), lowerLunar.CalendarSystem())
	}
	lowerBack, err := LunarToSolarByYMD(-722, 12, 16, false)
	if err != nil {
		t.Fatal(err)
	}
	if lowerBack.Solar().Year() != -721 || int(lowerBack.Solar().Month()) != 1 || lowerBack.Solar().Day() != 1 {
		t.Fatalf("unexpected -722 boundary roundtrip: %v", lowerBack.Solar())
	}
	if _, err := LunarToSolarByYMD(-722, 1, 1, false); err == nil {
		t.Fatal("expected N_-722 dates before supported civil range to be rejected")
	}
	results, err := LunarToSolarWithCalendar("-250年正月初一", AncientCalendarLu)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Solar().Year() != -251 || int(results[0].Solar().Month()) != 12 || results[0].Solar().Day() != 1 {
		t.Fatalf("unexpected LunarToSolarWithCalendar result: %#v", results)
	}
	defaultResults, err := LunarToSolar("-250年正月初一")
	if err != nil {
		t.Fatal(err)
	}
	if len(defaultResults) != 1 || defaultResults[0].Solar().Year() != -251 || int(defaultResults[0].Solar().Month()) != 11 || defaultResults[0].Solar().Day() != 30 {
		t.Fatalf("unexpected default LunarToSolar result: %#v", defaultResults)
	}
	transition, err := SolarToLunarByYMD(-221, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if transition.Lunar().CalendarSystem() != AncientCalendarZhou {
		t.Fatalf("expected -221 early date to use Zhou fallback, got %q", transition.Lunar().CalendarSystem())
	}
	zhouTransition, err := SolarToLunarByYMDWithCalendar(-221, 11, 29, AncientCalendarZhou)
	if err != nil {
		t.Fatal(err)
	}
	zhouLunar := zhouTransition.Lunar()
	if zhouLunar.LunarYear() != -220 || zhouLunar.LunarMonth() != 1 || zhouLunar.LunarDay() != 1 || zhouLunar.CalendarSystem() != AncientCalendarZhou {
		t.Fatalf("unexpected explicit Zhou -221 transition: y=%d m=%d d=%d system=%q",
			zhouLunar.LunarYear(), zhouLunar.LunarMonth(), zhouLunar.LunarDay(), zhouLunar.CalendarSystem())
	}
	zhouBack, err := LunarToSolarByYMDWithCalendar(-220, 1, 1, false, AncientCalendarZhou)
	if err != nil {
		t.Fatal(err)
	}
	if zhouBack.Solar().Year() != -221 || int(zhouBack.Solar().Month()) != 11 || zhouBack.Solar().Day() != 29 {
		t.Fatalf("unexpected explicit Zhou N_-220 roundtrip: %v", zhouBack.Solar())
	}
	if _, err := LunarToSolarByYMDWithCalendar(-220, 3, 1, false, AncientCalendarZhou); err == nil {
		t.Fatal("expected explicit Zhou N_-220 dates after supported civil range to be rejected")
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

func TestLunarToSolarReturnsErrorWhenNoTextCandidate(t *testing.T) {
	testCases := []string{
		"不存在元年正月初一",
		"元丰六年十三月初一",
		"元丰六年十月甲丑日",
	}
	for _, tc := range testCases {
		res, err := LunarToSolar(tc)
		if err == nil {
			t.Fatalf("expected error for %q, got nil", tc)
		}
		if len(res) != 0 {
			t.Fatalf("expected empty result for %q, got %v", tc, res)
		}
	}
}

func TestHistoricalEraRegression(t *testing.T) {
	compareRanges := func(name string, got, want [][]int) {
		t.Helper()
		if len(got) != len(want) {
			t.Fatalf("%s range count mismatch: got=%v want=%v", name, got, want)
		}
		for i := range got {
			if len(got[i]) != len(want[i]) || got[i][0] != want[i][0] || got[i][1] != want[i][1] {
				t.Fatalf("%s range mismatch: got=%v want=%v", name, got, want)
			}
		}
	}

	compareRanges("祥兴", nianHaoMap()["祥兴"], [][]int{{1278, 1279}})
	compareRanges("金天兴", liaoJinYuanEraMap()["天兴"], [][]int{{1232, 1234}})
	compareRanges("贞祐", liaoJinYuanEraMap()["贞祐"], [][]int{{1213, 1217}})
	compareRanges("贞佑 alias", liaoJinYuanEraMap()["贞佑"], [][]int{{1213, 1217}})
	for alias, canonical := range map[string]string{
		"延佑": "延祐",
		"德佑": "德祐",
		"宝佑": "宝祐",
		"淳佑": "淳祐",
		"元佑": "元祐",
		"嘉佑": "嘉祐",
		"皇佑": "皇祐",
		"景佑": "景祐",
		"乾佑": "乾祐",
		"天佑": "天祐",
	} {
		compareRanges(alias+" alias", nianHaoMap()[alias], nianHaoMap()[canonical])
	}

	for _, tc := range []string{
		"祥兴元年正月初一",
		"贞祐元年正月初一",
		"贞佑元年正月初一",
		"嘉佑元年正月初一",
		"元佑元年正月初一",
		"天佑元年正月初一",
	} {
		res, err := LunarToSolar(tc)
		if err != nil {
			t.Fatalf("LunarToSolar(%q) error: %v", tc, err)
		}
		if len(res) == 0 {
			t.Fatalf("LunarToSolar(%q) returned no candidates", tc)
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
