package calendar

import (
	"strings"
	"testing"
)

// 年份早于年号表首个条目时，innerEras 曾追加零值 EraDesc，描述退化成空年号的“年十月廿五”。
// 兜底档（≤ −104 送进明清表）与三国/南北朝/南明的补充分支都会命中，这里锁住修复后的不变量：
// 任何农历记录的年号描述都非空，且不会以“年”开头（即不存在空年号）。
// When a year precedes the first entry of an era table, innerEras used to append a zero-valued
// EraDesc and the description degenerated to an empty era name ("年十月廿五"). The catch-all branch
// (years <= -104 sent to the Ming/Qing table) and the Three Kingdoms / Northern-Southern dynasties /
// Southern Ming supplements all hit it. These assertions pin the invariant after the fix: every lunar
// record has a non-empty era description that never starts with "年" alone.
func TestEraDescriptionNeverEmpty(t *testing.T) {
	cases := []struct {
		year, month, day int
		primaryEra       string
	}{
		{-104, 12, 20, "前一零五年"}, // 报告里的样例：太初改历交接年
		{-103, 1, 5, "前一零五年"},   // 同上，跨公历年
		{-104, 1, 1, "元封六年"},    // 同一年里年号仍正常
		{221, 1, 5, "延康元年"},     // 三国补充分支（农历年 220 早于魏表首年）
		{384, 1, 5, "太元八年"},     // 南北朝补充分支
		{397, 1, 5, "太元二十一年"},   // 北凉补充分支
		{534, 1, 5, "中大通五年"},    // 东魏北齐补充分支
		{1647, 1, 5, "顺治三年"},    // 南明补充分支（农历年 1646 早于永历表首年）
		{1648, 1, 5, "顺治四年"},
		{1912, 12, 25, "民国元年"},
	}
	for _, c := range cases {
		fwd, err := SolarToLunarByYMD(c.year, c.month, c.day)
		if err != nil {
			t.Fatalf("SolarToLunarByYMD(%d,%d,%d) error: %v", c.year, c.month, c.day, err)
		}
		if len(fwd.Lunars()) == 0 {
			t.Fatalf("%d-%02d-%02d: no lunar records", c.year, c.month, c.day)
		}
		for i, l := range fwd.Lunars() {
			for _, desc := range l.LunarDescWithDynastyAndEmperor() {
				if strings.TrimSpace(desc) == "" || strings.HasPrefix(strings.TrimSpace(desc), "年") {
					t.Errorf("%d-%02d-%02d record %d: empty-era description %q", c.year, c.month, c.day, i, desc)
				}
			}
			for _, info := range l.LunarInfo() {
				if strings.TrimSpace(info.EraDesc) == "" || strings.TrimSpace(info.EraDesc) == "年" {
					t.Errorf("%d-%02d-%02d record %d: empty era description %q", c.year, c.month, c.day, i, info.EraDesc)
				}
				if strings.TrimSpace(info.LunarWithEraDesc) == "" || strings.HasPrefix(strings.TrimSpace(info.LunarWithEraDesc), "年") {
					t.Errorf("%d-%02d-%02d record %d: empty-era description %q", c.year, c.month, c.day, i, info.LunarWithEraDesc)
				}
			}
		}
		if got := fwd.LunarInfo()[0].EraDesc; got != c.primaryEra {
			t.Errorf("%d-%02d-%02d primary era = %q, want %q", c.year, c.month, c.day, got, c.primaryEra)
		}
	}
}

// 抽样式全库扫描：跨朝代边界的年份最容易踩到零值年号，这里按区间取样确认没有空年号描述。
// A sampled sweep: years at dynasty boundaries are the ones that hit the zero-valued era, so sample
// each band and confirm no record carries an empty era description.
func TestEraDescriptionSampledSweep(t *testing.T) {
	for year := -200; year <= 1920; year += 3 {
		for _, md := range [][2]int{{1, 5}, {5, 20}, {11, 25}} {
			fwd, err := SolarToLunarByYMD(year, md[0], md[1])
			if err != nil {
				continue
			}
			for i, l := range fwd.Lunars() {
				for _, info := range l.LunarInfo() {
					if strings.TrimSpace(info.EraDesc) == "" || strings.TrimSpace(info.EraDesc) == "年" {
						t.Fatalf("%d-%02d-%02d record %d: empty era description (desc %q)", year, md[0], md[1], i, info.LunarWithEraDesc)
					}
				}
			}
		}
	}
}
