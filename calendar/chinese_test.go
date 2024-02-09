package calendar

import (
	"fmt"
	"testing"
	"time"
)

type lunarSolar struct {
	Lyear  int
	Lmonth int
	Lday   int
	Leap   bool
	Year   int
	Month  int
	Day    int
}

func Test_ChineseCalendar(t *testing.T) {
	var testData = []lunarSolar{
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
		var lyear int = v.Year
		lmonth, lday, leap, desp := SolarToLunar(time.Date(v.Year, time.Month(v.Month), v.Day, 0, 0, 0, 0, time.Local))
		if lmonth > v.Month {
			lyear--
		}
		fmt.Println(lyear, desp, v.Year, v.Month, v.Day)
		if lyear != v.Lyear || lmonth != v.Lmonth || lday != v.Lday || leap != v.Leap {
			t.Fatal(v, lyear, lmonth, lday, leap, desp)
		}

		date := LunarToSolar(v.Lyear, v.Lmonth, v.Lday, v.Leap)
		if date.Year() != v.Year || int(date.Month()) != v.Month || date.Day() != v.Day {
			t.Fatal(v, date)
		}
	}
}
