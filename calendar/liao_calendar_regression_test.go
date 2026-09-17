package calendar

import (
	"testing"
	"time"
)

// 辽/金/元历修正表曾漏掉 9 个年份，其中缺 1093 会让辽历 1094-01-19 走进第 13 个月并越界 panic。
func TestLiaoJinYuanCorrectionTableIsComplete(t *testing.T) {
	want := map[int]uint32{
		993:  1767184896,
		1000: 2774542080,
		1074: 3040878080,
		1093: 1531977216,
		1102: 1523258624,
		1103: 1456482304,
		1175: 633116928,
		1208: 3047822080,
		1252: 3600820736,
	}
	table := liaoJinYuanCals()
	for year, magic := range want {
		got, ok := table[year]
		if !ok {
			t.Fatalf("liaoJinYuanCals() 缺少 %d 年（源表 y%d 有修正）", year, year)
		}
		if got != magic {
			t.Fatalf("liaoJinYuanCals()[%d] = %d, want %d", year, got, magic)
		}
	}
	if len(table) != 72 {
		t.Fatalf("liaoJinYuanCals() 行数 = %d, want 72（与补全后的表行数一致）", len(table))
	}
}

// 缺行年份的岁末/年初边界：既要不再 panic，也要给出源表推出的那一天。
// Year boundaries of the years that used to be missing: they must not panic and must report
// the day the source table implies.
func TestSolarToLunarNorthernCalendarYearBoundaries(t *testing.T) {
	cases := []struct {
		year, month, day int
		want             string
	}{
		{1094, 1, 19, "辽 大安九年腊月三十"},  // 缺 1093 行：旧实现 panic（index out of range [13]）
		{1176, 2, 12, "金 大定十五年腊月三十"}, // 缺 1175 行：旧实现回退"无法获取农历信息"
		{1209, 2, 6, "金 泰和八年腊月三十"},   // 缺 1208 行：同上
		{1252, 2, 11, "元 元宪宗二年正月初一"}, // 缺 1252 行：同上
	}
	for _, tc := range cases {
		date := time.Date(tc.year, time.Month(tc.month), tc.day, 0, 0, 0, 0, getCst())
		got, err := SolarToLunar(date)
		if err != nil {
			t.Fatalf("SolarToLunar(%s) err=%v", date.Format("2006-01-02"), err)
		}
		if !containsLine(got.LunarDescWithDynasty(), tc.want) {
			t.Fatalf("SolarToLunar(%s) = %v, want a line %q",
				date.Format("2006-01-02"), got.LunarDescWithDynasty(), tc.want)
		}
	}
}

// 缺行年份还会让整月错位 30 天（旧表把闰月当普通月）。这里钉住错位段的头一天。
// Missing rows also shifted a whole month by 30 days (the old table treated a leap month as
// an ordinary one). Pin the first day of each shifted stretch.
func TestSolarToLunarNorthernCalendarMonthAlignment(t *testing.T) {
	cases := []struct {
		year, month, day int
		want             string
	}{
		{1102, 11, 11, "辽 乾统二年九月廿九"},
		{1102, 11, 12, "辽 乾统二年九月三十"}, // 旧表：十月初一
		{1102, 11, 13, "辽 乾统二年十月初一"},
		{1103, 9, 2, "辽 乾统三年七月廿九"},
		{1103, 9, 3, "辽 乾统三年七月三十"}, // 旧表：八月初一
		{1103, 9, 4, "辽 乾统三年八月初一"},
		{1252, 2, 10, "元 元宪宗元年腊月廿九"},
		{1252, 2, 11, "元 元宪宗二年正月初一"}, // 旧表：回退串
	}
	for _, tc := range cases {
		date := time.Date(tc.year, time.Month(tc.month), tc.day, 0, 0, 0, 0, getCst())
		got, err := SolarToLunar(date)
		if err != nil {
			t.Fatalf("SolarToLunar(%s) err=%v", date.Format("2006-01-02"), err)
		}
		if !containsLine(got.LunarDescWithDynasty(), tc.want) {
			t.Fatalf("SolarToLunar(%s) = %v, want a line %q",
				date.Format("2006-01-02"), got.LunarDescWithDynasty(), tc.want)
		}
	}
}

func containsLine(lines []string, want string) bool {
	for _, line := range lines {
		if line == want {
			return true
		}
	}
	return false
}
