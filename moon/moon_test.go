package moon

import (
	"math"
	"testing"
	"time"
)

func Test_MoonPhaseDate(t *testing.T) {
	assertUnixClose := func(name string, got time.Time, want int64) {
		t.Helper()
		if math.Abs(float64(got.Unix()-want)) > 5 {
			t.Fatalf("%s = %d, want %d", name, got.Unix(), want)
		}
	}

	//指定北京时间2022年1月20日
	tz := time.FixedZone("CST", 8*3600)
	date := time.Date(2022, 01, 20, 00, 00, 00, 00, tz)
	//指定日期后的下一个朔月
	moonPhase01 := NextShuoYue(date)
	assertUnixClose("NextShuoYue", moonPhase01, 1643694356)
	//指定日期后的上一个朔月
	moonPhase01 = LastShuoYue(date)
	assertUnixClose("LastShuoYue", moonPhase01, 1641148406)
	//离指定日期最近的朔月
	moonPhase01 = ClosestShuoYue(date)
	assertUnixClose("ClosestShuoYue", moonPhase01, 1643694356)
	//离指定日期最近的望月时间
	moonPhase01 = ClosestWangYue(date)
	assertUnixClose("ClosestWangYue", moonPhase01, 1642463301)
	//离指定日期最近的上弦月时间
	moonPhase01 = ClosestShangXianYue(date)
	assertUnixClose("ClosestShangXianYue", moonPhase01, 1641751871)
	//离指定日期最近的下弦月时间
	moonPhase01 = ClosestXiaXianYue(date)
	assertUnixClose("ClosestXiaXianYue", moonPhase01, 1643118050)
}

func TestMoonPhaseEnglishAliases(t *testing.T) {
	tz := time.FixedZone("CST", 8*3600)
	date := time.Date(2022, 1, 20, 0, 0, 0, 0, tz)
	year := 2022.25

	assertSameTime := func(name string, got, want time.Time) {
		t.Helper()
		if !got.Equal(want) {
			t.Fatalf("%s = %s, want %s", name, got, want)
		}
	}

	assertSameTime("NewMoon", NewMoon(year), ShuoYue(year))
	assertSameTime("NextNewMoon", NextNewMoon(date), NextShuoYue(date))
	assertSameTime("LastNewMoon", LastNewMoon(date), LastShuoYue(date))
	assertSameTime("ClosestNewMoon", ClosestNewMoon(date), ClosestShuoYue(date))

	assertSameTime("FullMoon", FullMoon(year), WangYue(year))
	assertSameTime("NextFullMoon", NextFullMoon(date), NextWangYue(date))
	assertSameTime("LastFullMoon", LastFullMoon(date), LastWangYue(date))
	assertSameTime("ClosestFullMoon", ClosestFullMoon(date), ClosestWangYue(date))

	assertSameTime("FirstQuarter", FirstQuarter(year), ShangXianYue(year))
	assertSameTime("NextFirstQuarter", NextFirstQuarter(date), NextShangXianYue(date))
	assertSameTime("LastFirstQuarter", LastFirstQuarter(date), LastShangXianYue(date))
	assertSameTime("ClosestFirstQuarter", ClosestFirstQuarter(date), ClosestShangXianYue(date))

	assertSameTime("LastQuarter", LastQuarter(year), XiaXianYue(year))
	assertSameTime("NextLastQuarter", NextLastQuarter(date), NextXiaXianYue(date))
	assertSameTime("LastLastQuarter", LastLastQuarter(date), LastXiaXianYue(date))
	assertSameTime("ClosestLastQuarter", ClosestLastQuarter(date), ClosestXiaXianYue(date))
}
