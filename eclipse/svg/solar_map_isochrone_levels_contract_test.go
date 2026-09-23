package svg

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
	eclipsecore "github.com/starainrt/astro/eclipse"
)

// 等时线电平的期望网格在测试内用整数纳秒与当地日起点独立重算，期望时刻写成字面量。

const solarEclipseContractLevelCap = 64

func solarEclipseContractLevelGrid(
	begin, end time.Time, step time.Duration, location *time.Location,
) []time.Time {
	if step <= 0 || begin.IsZero() || end.IsZero() || location == nil {
		return nil
	}
	local := begin.In(location)
	dayStart := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, location)
	ticks := (int64(local.Sub(dayStart)) + int64(step)/2) / int64(step)
	current := dayStart.Add(time.Duration(ticks) * step)
	if current.Before(begin) {
		current = current.Add(step)
	}
	levels := make([]time.Time, 0, solarEclipseContractLevelCap)
	for !current.After(end) && len(levels) < solarEclipseContractLevelCap {
		levels = append(levels, current)
		current = current.Add(step)
	}
	return levels
}

func solarEclipseContractLevelText(levels []time.Time) []string {
	formatted := make([]string, 0, len(levels))
	for _, level := range levels {
		formatted = append(formatted, level.UTC().Format(time.RFC3339))
	}
	return formatted
}

func solarEclipseContractLevelDiff(got, want []string) string {
	if strings.Join(got, ",") == strings.Join(want, ",") {
		return ""
	}
	for index := 0; index < len(got) && index < len(want); index++ {
		if got[index] != want[index] {
			return "level " + strconv.Itoa(index) + " = " + got[index] + ", want " + want[index] +
				" (" + strconv.Itoa(len(got)) + " vs " + strconv.Itoa(len(want)) + " levels)"
		}
	}
	return strconv.Itoa(len(got)) + " levels, want " + strconv.Itoa(len(want))
}

// 导出 TT/UTC 换算把力学时儒略日换回 UTC 时刻，口径与核心写进 PartialBegin/EndOnEarth 的一致。
func solarEclipseContractTimeFromTT(ttJDE float64) time.Time {
	return basic.JD2DateByZone(ttJDE-basic.DeltaT(ttJDE, true)/86400, time.UTC, false)
}

func TestSolarEclipseGreatestTimeLevelsContract(t *testing.T) {
	day := func(hour, minute, second int) time.Time {
		return time.Date(2024, time.January, 1, hour, minute, second, 0, time.UTC)
	}
	utc8 := time.FixedZone("UTC+8", 8*3600)
	cases := []struct {
		name       string
		begin, end time.Time
		step       time.Duration
		location   *time.Location
		want       []string
		wantCount  int
		wantFirst  string
		wantLast   string
	}{
		{
			name: "起点落在刻度之间时从下一个整刻度开始", begin: day(0, 7, 0), end: day(2, 7, 0),
			step: 30 * time.Minute, location: time.UTC,
			want: []string{"2024-01-01T00:30:00Z", "2024-01-01T01:00:00Z", "2024-01-01T01:30:00Z", "2024-01-01T02:00:00Z"},
		},
		{
			name: "起点恰在刻度上时不重复生成该刻度", begin: day(0, 30, 0), end: day(2, 0, 0),
			step: 30 * time.Minute, location: time.UTC,
			want: []string{"2024-01-01T00:30:00Z", "2024-01-01T01:00:00Z", "2024-01-01T01:30:00Z", "2024-01-01T02:00:00Z"},
		},
		{
			name: "终点恰在刻度上时包含终点", begin: day(0, 7, 0), end: day(1, 30, 0),
			step: 30 * time.Minute, location: time.UTC,
			want: []string{"2024-01-01T00:30:00Z", "2024-01-01T01:00:00Z", "2024-01-01T01:30:00Z"},
		},
		{
			name: "窗口内没有整刻度时为空", begin: day(0, 1, 0), end: day(0, 20, 0),
			step: time.Hour, location: time.UTC,
			want: nil,
		},
		{
			name: "零步长关闭等时线", begin: day(0, 7, 0), end: day(2, 7, 0),
			step: 0, location: time.UTC,
			want: nil,
		},
		{
			name: "负步长关闭等时线", begin: day(0, 7, 0), end: day(2, 7, 0),
			step: -30 * time.Minute, location: time.UTC,
			want: nil,
		},
		{
			name: "偏食窗口缺失时为空", begin: time.Time{}, end: day(2, 7, 0),
			step: 30 * time.Minute, location: time.UTC,
			want: nil,
		},
		{
			name: "超过上限时截断到 64 条", begin: day(0, 0, 30), end: day(2, 0, 30),
			step: time.Minute, location: time.UTC,
			wantCount: 64, wantFirst: "2024-01-01T00:01:00Z", wantLast: "2024-01-01T01:04:00Z",
		},
		{
			name: "整刻度按当地日午夜对齐", begin: time.Date(2024, time.January, 1, 17, 37, 0, 0, time.UTC),
			end: time.Date(2024, time.January, 3, 0, 0, 0, 0, time.UTC), step: 5 * time.Hour, location: utc8,
			want: []string{
				"2024-01-01T21:00:00Z", "2024-01-02T02:00:00Z", "2024-01-02T07:00:00Z",
				"2024-01-02T12:00:00Z", "2024-01-02T17:00:00Z", "2024-01-02T22:00:00Z",
			},
		},
	}
	for _, testCase := range cases {
		info := eclipsecore.SolarEclipseInfo{
			PartialBeginOnEarth: testCase.begin,
			PartialEndOnEarth:   testCase.end,
		}
		levels := solarEclipseGreatestTimeLevels(info, SolarEclipseMapSVGOptions{
			Location: testCase.location, GreatestTimeStep: testCase.step,
		})
		got := solarEclipseContractLevelText(levels)
		want := solarEclipseContractLevelText(solarEclipseContractLevelGrid(
			testCase.begin, testCase.end, testCase.step, testCase.location))
		if diff := solarEclipseContractLevelDiff(got, want); diff != "" {
			t.Fatalf("%s: levels disagree with the independent grid: %s", testCase.name, diff)
		}
		if testCase.want != nil || len(got) == 0 {
			if diff := solarEclipseContractLevelDiff(got, testCase.want); diff != "" {
				t.Fatalf("%s: literal expectation: %s", testCase.name, diff)
			}
			continue
		}
		if len(got) != testCase.wantCount || got[0] != testCase.wantFirst || got[len(got)-1] != testCase.wantLast {
			t.Fatalf("%s: %d levels from %s to %s, want %d from %s to %s",
				testCase.name, len(got), got[0], got[len(got)-1], testCase.wantCount, testCase.wantFirst, testCase.wantLast)
		}
	}
}

// 真实事件上再走一遍导出 TT/UTC 换算：窗口端点由力学时儒略日换回，网格仍由测试独立重算。
func TestSolarEclipseGreatestTimeLevelsOnRealEclipseContract(t *testing.T) {
	date := time.Date(2009, time.July, 22, 0, 0, 0, 0, time.UTC)
	reported, ok := eclipsecore.SolarEclipseOnDateNASABulletinSplitK(date)
	if !ok {
		t.Fatal("missing 2009-07-22 eclipse")
	}
	// 同一力学时时刻经导出换算回 UTC，不依赖 eclipse 包内部的 TT→UTC 实现。
	beginTT := basic.UTC2TT(basic.Date2JD(reported.PartialBeginOnEarth.UTC()))
	endTT := basic.UTC2TT(basic.Date2JD(reported.PartialEndOnEarth.UTC()))
	info := eclipsecore.SolarEclipseInfo{
		PartialBeginOnEarth: solarEclipseContractTimeFromTT(beginTT),
		PartialEndOnEarth:   solarEclipseContractTimeFromTT(endTT),
	}
	step := 30 * time.Minute
	levels := solarEclipseGreatestTimeLevels(info, SolarEclipseMapSVGOptions{
		Location: time.UTC, GreatestTimeStep: step,
	})
	got := solarEclipseContractLevelText(levels)
	want := []string{
		"2009-07-22T00:00:00Z", "2009-07-22T00:30:00Z", "2009-07-22T01:00:00Z",
		"2009-07-22T01:30:00Z", "2009-07-22T02:00:00Z", "2009-07-22T02:30:00Z",
		"2009-07-22T03:00:00Z", "2009-07-22T03:30:00Z", "2009-07-22T04:00:00Z",
		"2009-07-22T04:30:00Z", "2009-07-22T05:00:00Z",
	}
	if diff := solarEclipseContractLevelDiff(got, want); diff != "" {
		t.Fatalf("2009-07-22 literal expectation: %s", diff)
	}
	if recomputed := solarEclipseContractLevelText(solarEclipseContractLevelGrid(
		info.PartialBeginOnEarth, info.PartialEndOnEarth, step, time.UTC)); solarEclipseContractLevelDiff(got, recomputed) != "" {
		t.Fatalf("2009-07-22 levels disagree with the independent grid: %s",
			solarEclipseContractLevelDiff(got, recomputed))
	}
}
