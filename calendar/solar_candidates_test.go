package calendar

import (
	"testing"
	"time"
)

// L3 契约：Solar() 保持库内默认选择，SolarCandidates() 给出全部合法候选（主答案在首），
// 且每个候选都能正向回到该农历日。
func TestSolarCandidatesForReformWindows(t *testing.T) {
	cases := []struct {
		year, month, day int
		leap             bool
		want             []string
	}{
		{762, 4, 1, false, []string{"0762-03-01", "0762-04-29"}},     // 唐肃宗改制
		{700, 11, 1, false, []string{"0700-12-15", "0700-10-17"}},    // 武则天改制
		{700, 12, 1, false, []string{"0700-11-15", "0701-01-14"}},    // 武则天改制
		{23, 12, 1, false, []string{"0023-12-02", "0023-12-31"}},     // 王莽改制
		{239, 12, 1, false, []string{"0239-12-13", "0240-01-12"}},    // 魏明帝改制
		{-104, 10, 1, false, []string{"-0104-11-26", "-0105-11-08"}}, // 太初改历交接
	}
	for _, c := range cases {
		res, err := LunarToSolarByYMD(c.year, c.month, c.day, c.leap)
		if err != nil {
			t.Fatalf("LunarToSolarByYMD(%d,%d,%d,%v) error: %v", c.year, c.month, c.day, c.leap, err)
		}
		candidates := res.SolarCandidates()
		if len(candidates) != len(c.want) {
			t.Errorf("%d/%d/%d: got %d candidates %v, want %d %v",
				c.year, c.month, c.day, len(candidates), formatCandidates(candidates), len(c.want), c.want)
			continue
		}
		if !candidates[0].Equal(res.Solar()) {
			t.Errorf("%d/%d/%d: first candidate %s is not Solar() %s",
				c.year, c.month, c.day, formatAncientSolar(candidates[0]), formatAncientSolar(res.Solar()))
		}
		for i, v := range candidates {
			if got := formatAncientSolar(v); got != c.want[i] {
				t.Errorf("%d/%d/%d: candidate %d = %s, want %s", c.year, c.month, c.day, i, got, c.want[i])
			}
			for j := 0; j < i; j++ {
				if candidates[j].Equal(v) {
					t.Errorf("%d/%d/%d: candidate %d duplicates candidate %d (%s)",
						c.year, c.month, c.day, i, j, formatAncientSolar(v))
				}
			}
			fwd, err := SolarToLunarByYMD(v.Year(), int(v.Month()), v.Day())
			if err != nil {
				t.Errorf("%d/%d/%d: candidate %s does not convert forward: %v",
					c.year, c.month, c.day, formatAncientSolar(v), err)
				continue
			}
			l := fwd.Lunar()
			if l.LunarYear() != c.year || l.LunarMonth() != c.month || l.LunarDay() != c.day || l.IsLeap() != c.leap {
				t.Errorf("%d/%d/%d: candidate %s converts back to lunar %d/%d/%d leap=%v",
					c.year, c.month, c.day, formatAncientSolar(v),
					l.LunarYear(), l.LunarMonth(), l.LunarDay(), l.IsLeap())
			}
		}
	}
}

// 改历窗口外只有一个候选，且 SolarCandidates() 返回的是副本。
func TestSolarCandidatesAreSingleOutsideReformWindows(t *testing.T) {
	cases := []struct{ year, month, day int }{
		{763, 4, 1}, {701, 11, 1}, {24, 12, 1}, {241, 12, 1}, {-103, 10, 1},
		{-105, 10, 1}, {2025, 1, 1}, {1912, 1, 1}, {1000, 7, 7}, {1, 1, 1},
	}
	for _, c := range cases {
		res, err := LunarToSolarByYMD(c.year, c.month, c.day, false)
		if err != nil {
			t.Fatalf("LunarToSolarByYMD(%d,%d,%d) error: %v", c.year, c.month, c.day, err)
		}
		candidates := res.SolarCandidates()
		if len(candidates) != 1 {
			t.Errorf("%d/%d/%d: got %d candidates %v, want a single one",
				c.year, c.month, c.day, len(candidates), formatCandidates(candidates))
			continue
		}
		if !candidates[0].Equal(res.Solar()) {
			t.Errorf("%d/%d/%d: single candidate %s is not Solar() %s",
				c.year, c.month, c.day, formatAncientSolar(candidates[0]), formatAncientSolar(res.Solar()))
		}
		// 返回值必须是副本。
		candidates[0] = candidates[0].AddDate(1, 1, 1)
		if again := res.SolarCandidates(); !again[0].Equal(res.Solar()) {
			t.Errorf("%d/%d/%d: SolarCandidates() is not a copy (%s)", c.year, c.month, c.day, formatAncientSolar(again[0]))
		}
	}
}

// 1582 改历跳过的十天本身不存在，与 L2/L3 无关。
func TestGregorianSkippedDaysHaveNoCandidates(t *testing.T) {
	for day := 5; day <= 14; day++ {
		if _, err := SolarToLunarByYMD(1582, 10, day); err == nil {
			t.Errorf("SolarToLunarByYMD(1582,10,%d) unexpectedly succeeded", day)
		}
	}
	res, err := LunarToSolarByYMD(1582, 9, 18, false)
	if err != nil {
		t.Fatalf("LunarToSolarByYMD(1582,9,18) error: %v", err)
	}
	if got := len(res.SolarCandidates()); got != 1 {
		t.Errorf("lunar 1582/9/18 has %d candidates, want 1", got)
	}
	if _, err := time.Parse("2006-01-02", formatAncientSolar(res.Solar())); err != nil {
		t.Errorf("solar label %q is not a valid date", formatAncientSolar(res.Solar()))
	}
}

func formatCandidates(candidates []time.Time) []string {
	out := make([]string, 0, len(candidates))
	for _, v := range candidates {
		out = append(out, formatAncientSolar(v))
	}
	return out
}
