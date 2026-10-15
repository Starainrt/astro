package calendar

import (
	"fmt"
	"testing"
	"time"
)

// 1582 改历：反向的 +10 天窗口必须与正向的 −10 天修正对称（1582-10-05..1583-01-13）。
// formatAncientSolar 统一负年份的显示。
func formatAncientSolar(t time.Time) string {
	year, sign := t.Year(), ""
	if year < 0 {
		year, sign = -year, "-"
	}
	return fmt.Sprintf("%s%04d-%02d-%02d", sign, year, int(t.Month()), t.Day())
}

func TestGregorianReformReverseMatchesForward(t *testing.T) {
	cases := []struct {
		ly, lm, ld int
		leap       bool
		want       string
	}{
		{1582, 12, 8, false, "1583-01-01"},  // 连续坐标 1582-12-22，原有 1582 规则覆盖
		{1582, 12, 18, false, "1583-01-11"}, // 连续坐标 1583-01-01，修复前返回 1583-01-01
		{1582, 12, 30, false, "1583-01-23"}, // 连续坐标 1583-01-13，修复前返回 1583-01-13
		{1583, 1, 1, false, "1583-01-24"},   // 连续坐标 1583-01-14，本就不该 +10
	}
	for _, tc := range cases {
		got, err := LunarToSolarByYMD(tc.ly, tc.lm, tc.ld, tc.leap)
		if err != nil {
			t.Fatalf("LunarToSolarByYMD(%d,%d,%d) err=%v", tc.ly, tc.lm, tc.ld, err)
		}
		if s := got.Solar().Format("2006-01-02"); s != tc.want {
			t.Fatalf("LunarToSolarByYMD(%d,%d,%d) = %s, want %s", tc.ly, tc.lm, tc.ld, s, tc.want)
		}
	}

	// 改历前后整段往返一致（跳过被删除的 1582-10-05..14，它们在公历里不存在）。
	// The whole span round-trips apart from the ten days the reform deleted.
	for day := time.Date(1582, 10, 15, 0, 0, 0, 0, getCst()); day.Before(time.Date(1583, 3, 1, 0, 0, 0, 0, getCst())); day = day.AddDate(0, 0, 1) {
		y, m, d := day.Year(), int(day.Month()), day.Day()
		lunar, err := SolarToLunarByYMD(y, m, d)
		if err != nil {
			t.Fatalf("SolarToLunarByYMD(%d,%d,%d) err=%v", y, m, d, err)
		}
		l := lunar.Lunar()
		back, err := LunarToSolarByYMD(l.LunarYear(), l.LunarMonth(), l.LunarDay(), l.IsLeap())
		if err != nil {
			t.Fatalf("LunarToSolarByYMD(%d,%d,%d) err=%v", l.LunarYear(), l.LunarMonth(), l.LunarDay(), err)
		}
		if s := back.Solar(); s.Year() != y || int(s.Month()) != m || s.Day() != d {
			t.Fatalf("%d-%02d-%02d → 农历 %d/%d/%d → 反向 %s（改历往返不一致）",
				y, m, d, l.LunarYear(), l.LunarMonth(), l.LunarDay(), s.Format("2006-01-02"))
		}
	}
}

// 先秦岁首跨界：正向按公历年选默认古历、反向按农历年选，交界处必须自洽。
// 农历 -479 年正月落在公历 -480 年 12 月（春秋历），反向不得选成周历。
func TestAncientYearStartCrossingRoundTrips(t *testing.T) {
	cases := []struct {
		ly, lm, ld int
		want       string
	}{
		{-479, 1, 1, "-0480-12-12"},
		{-479, 1, 20, "-0480-12-31"},
		{-479, 1, 22, "-0479-01-01"},
		{-479, 1, 29, "-0479-01-08"},
	}
	for _, tc := range cases {
		got, err := LunarToSolarByYMD(tc.ly, tc.lm, tc.ld, false)
		if err != nil {
			t.Fatalf("LunarToSolarByYMD(%d,%d,%d) err=%v", tc.ly, tc.lm, tc.ld, err)
		}
		if s := formatAncientSolar(got.Solar()); s != tc.want {
			t.Fatalf("LunarToSolarByYMD(%d,%d,%d) = %s, want %s", tc.ly, tc.lm, tc.ld, s, tc.want)
		}
	}
	// 该农历日在正向可达的范围内没有对应公历日（春秋历说是 -479-01-01、周历说是 -480-12-31，
	// 而正向对这两个公历日分别给出周历 1/22 与春秋历 1/20），默认口径必须报错而不是给错日子。
	// That lunar day has no civil date in the default frame (the forward maps -479-01-01 to Zhou
	// 1/22 and -480-12-31 to Chunqiu 1/20), so the default entry point must report an error.
	if _, err := LunarToSolarByYMD(-479, 1, 21, false); err == nil {
		t.Fatal("LunarToSolarByYMD(-479,1,21) 应报错：该农历日在默认历法口径下没有公历日")
	}

	// 交界窗口往返一致。
	// The junction window round-trips.
	for day := time.Date(-480, 12, 1, 0, 0, 0, 0, getCst()); day.Before(time.Date(-479, 2, 1, 0, 0, 0, 0, getCst())); day = day.AddDate(0, 0, 1) {
		y, m, d := day.Year(), int(day.Month()), day.Day()
		lunar, err := SolarToLunarByYMD(y, m, d)
		if err != nil {
			continue
		}
		l := lunar.Lunar()
		back, err := LunarToSolarByYMD(l.LunarYear(), l.LunarMonth(), l.LunarDay(), l.IsLeap())
		if err != nil {
			t.Fatalf("%d-%02d-%02d → 农历 %d/%d/%d → 反向 err=%v", y, m, d, l.LunarYear(), l.LunarMonth(), l.LunarDay(), err)
		}
		if s := back.Solar(); s.Year() != y || int(s.Month()) != m || s.Day() != d {
			t.Fatalf("%d-%02d-%02d → 农历 %d/%d/%d → 反向 %s（岁首跨界往返不一致）",
				y, m, d, l.LunarYear(), l.LunarMonth(), l.LunarDay(), formatAncientSolar(s))
		}
	}
}

// 王莽改历的 23/24 年交界：同一套月序偏移会把"腊月三十"反解到腊月初一（规范日碰撞），
// 此时应改用另一套偏移下能正向回到该农历日的那一天，并把主答案换成它。
// At the Wang Mang 23/24 boundary the primary month offset resolves "12th month, day 30" onto the
// 1st of that month (a canonical-day collision); the reverse must switch to the offset whose civil day
// converts forward to the requested lunar date and make that day the primary answer.
func TestWangMangBoundaryReverseIsSelfConsistent(t *testing.T) {
	cases := []struct {
		ly, lm, ld  int
		wantSolar   string
		selfConsist bool
	}{
		{23, 12, 29, "0023-12-30", true},
		{23, 12, 30, "0024-01-29", true},
		{24, 1, 1, "0024-01-30", true},
		// 23/12/31 在两套月序下都没有对应的公历日（该年有两个腊月，第二个腊月只有 30 天），
		// 保持既有行为返回偏移 1 的结果，只断言主答案不变成空值。
		{23, 12, 31, "0024-01-01", false},
	}
	for _, c := range cases {
		res, err := LunarToSolarByYMD(c.ly, c.lm, c.ld, false)
		if err != nil {
			t.Fatalf("LunarToSolarByYMD(%d,%d,%d) error: %v", c.ly, c.lm, c.ld, err)
		}
		if got := formatAncientSolar(res.Solar()); got != c.wantSolar {
			t.Errorf("lunar %d/%d/%d -> %s, want %s", c.ly, c.lm, c.ld, got, c.wantSolar)
		}
		if !c.selfConsist {
			continue
		}
		fwd, err := SolarToLunarByYMD(res.Solar().Year(), int(res.Solar().Month()), res.Solar().Day())
		if err != nil {
			t.Fatalf("SolarToLunarByYMD(%s) error: %v", c.wantSolar, err)
		}
		l := fwd.Lunar()
		if l.LunarYear() != c.ly || l.LunarMonth() != c.lm || l.LunarDay() != c.ld || l.IsLeap() {
			t.Errorf("lunar %d/%d/%d -> %s converts forward to %d/%d/%d",
				c.ly, c.lm, c.ld, c.wantSolar, l.LunarYear(), l.LunarMonth(), l.LunarDay())
		}
		if res.JD() != fwd.JD() {
			t.Errorf("lunar %d/%d/%d: JDE=%.1f, forward gives %.1f", c.ly, c.lm, c.ld, res.JD(), fwd.JD())
		}
		for _, cand := range res.SolarCandidates() {
			candFwd, err := SolarToLunarByYMD(cand.Year(), int(cand.Month()), cand.Day())
			if err != nil {
				t.Errorf("lunar %d/%d/%d candidate %s does not convert: %v", c.ly, c.lm, c.ld, formatAncientSolar(cand), err)
				continue
			}
			cl := candFwd.Lunar()
			if cl.LunarYear() != c.ly || cl.LunarMonth() != c.lm || cl.LunarDay() != c.ld || cl.IsLeap() {
				t.Errorf("lunar %d/%d/%d candidate %s converts to %d/%d/%d",
					c.ly, c.lm, c.ld, formatAncientSolar(cand), cl.LunarYear(), cl.LunarMonth(), cl.LunarDay())
			}
		}
	}
}
