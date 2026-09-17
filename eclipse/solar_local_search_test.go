package eclipse

import (
	"testing"
	"time"
)

func TestSearchLocalCentralSolarEclipseReportsExhaustion(t *testing.T) {
	start := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	// 北京：20 年跨度内必有全食。
	info, status := SearchLocalCentralSolarEclipse(start, 116.40, 39.90, 0,
		SolarEclipseLocalSearchOptions{Kind: SolarEclipseLocalSearchTotal, MaxYears: 20})
	if !status.Found || status.Exhausted {
		t.Fatalf("Beijing 20y: found=%v exhausted=%v", status.Found, status.Exhausted)
	}
	if got := info.GreatestEclipse.UTC().Format("2006-01-02"); got != "2035-09-02" {
		t.Fatalf("Beijing next total=%s, want 2035-09-02", got)
	}
	if !info.HasTotal || !info.HasCentral {
		t.Fatalf("Beijing 2035-09-02 should be a local total eclipse: %+v", info)
	}
	if status.YearsScanned <= 0 || status.MonthsScanned <= 0 || status.GlobalChecks <= 0 {
		t.Fatalf("status did not account for the work done: %+v", status)
	}
	// 广州：318 年跨度内没有可见日全食 —— 必须报告"用尽跨度"而不是含糊的 false。
	_, exhausted := SearchLocalCentralSolarEclipse(start, 113.26, 23.13, 0,
		SolarEclipseLocalSearchOptions{Kind: SolarEclipseLocalSearchTotal, MaxYears: 300})
	if exhausted.Found || !exhausted.Exhausted {
		t.Fatalf("Guangzhou 300y: found=%v exhausted=%v, want none with Exhausted set",
			exhausted.Found, exhausted.Exhausted)
	}
	if exhausted.YearsScanned < 290 || exhausted.YearsScanned > 301 {
		t.Fatalf("Guangzhou 300y covered %.1f years, want about 300", exhausted.YearsScanned)
	}
}

func TestSearchLocalCentralSolarEclipseHonoursWideHorizon(t *testing.T) {
	if testing.Short() {
		t.Skip("wide-horizon search takes about 13 s")
	}
	start := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	info, status := SearchLocalCentralSolarEclipse(start, 113.26, 23.13, 0,
		SolarEclipseLocalSearchOptions{Kind: SolarEclipseLocalSearchTotal, MaxYears: 1400})
	if !status.Found {
		t.Fatalf("Guangzhou 1400y: found=%v exhausted=%v scanned=%.1f",
			status.Found, status.Exhausted, status.YearsScanned)
	}
	if year := info.GreatestEclipse.UTC().Year(); year != 3301 {
		t.Fatalf("Guangzhou next total within 1400y is in %d, want 3301", year)
	}
}

func TestSearchLocalCentralSolarEclipseMatchesLegacyEntryPoint(t *testing.T) {
	start := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	legacy, ok := NextLocalTotalSolarEclipse(start, 121.47, 31.23, 0)
	if !ok {
		t.Fatal("legacy entry point found no Shanghai total eclipse")
	}
	info, status := SearchLocalCentralSolarEclipse(start, 121.47, 31.23, 0,
		SolarEclipseLocalSearchOptions{Kind: SolarEclipseLocalSearchTotal, MaxYears: 300})
	if !status.Found {
		t.Fatalf("new entry point did not find the Shanghai total eclipse: %+v", status)
	}
	if !info.GreatestEclipse.Equal(legacy.GreatestEclipse) || info.Magnitude != legacy.Magnitude {
		t.Fatalf("new entry point disagrees with the legacy one: %v/%.6f vs %v/%.6f",
			info.GreatestEclipse, info.Magnitude, legacy.GreatestEclipse, legacy.Magnitude)
	}
}

func TestSolarEclipseCandidatesEnumeratesTimetable(t *testing.T) {
	start := time.Date(2009, time.January, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2011, time.January, 1, 0, 0, 0, 0, time.UTC)
	candidates := SolarEclipseCandidates(start, end, SolarEclipseCandidateOptions{IncludeSaros: true})
	seen := map[string]SolarEclipseType{}
	for _, candidate := range candidates {
		if candidate.GreatestEclipse.Before(start) || !candidate.GreatestEclipse.Before(end) {
			t.Fatalf("candidate %v outside the requested range", candidate.GreatestEclipse)
		}
		seen[candidate.GreatestEclipse.UTC().Format("2006-01-02")] = candidate.Type
	}
	want := map[string]SolarEclipseType{
		"2009-01-26": SolarEclipseAnnular,
		"2009-07-22": SolarEclipseTotal,
		"2010-01-15": SolarEclipseAnnular,
		"2010-07-11": SolarEclipseTotal,
	}
	for date, eclipseType := range want {
		if seen[date] != eclipseType {
			t.Fatalf("candidate %s type=%v, want %v (all: %v)", date, seen[date], eclipseType, seen)
		}
	}
	for _, candidate := range candidates {
		if candidate.GreatestEclipse.UTC().Format("2006-01-02") == "2009-07-22" {
			if !candidate.HasSaros || candidate.Saros.Series != 136 {
				t.Fatalf("2009-07-22 Saros=%+v, want series 136", candidate.Saros)
			}
		}
	}
	central := SolarEclipseCandidates(start, end, SolarEclipseCandidateOptions{CentralOnly: true})
	if len(central) != len(candidates) {
		t.Fatalf("2009-2010 has no purely partial eclipses: central=%d all=%d", len(central), len(candidates))
	}
	for _, candidate := range central {
		if candidate.Type == SolarEclipsePartial {
			t.Fatalf("CentralOnly kept a partial eclipse: %v", candidate.GreatestEclipse)
		}
	}
}

func TestSolarEclipseCandidatesRespectRequestedRange(t *testing.T) {
	// 候选朔月取的是"最近"的朔，可能落在区间之前；返回值必须按食甚时刻双向过滤。
	inside := time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC)
	after := time.Date(2024, time.April, 9, 0, 0, 0, 0, time.UTC)
	found := SolarEclipseCandidates(inside, after, SolarEclipseCandidateOptions{})
	if len(found) != 1 {
		t.Fatalf("2024-04-08..04-09 candidates=%d, want the 2024-04-08 eclipse", len(found))
	}
	if year, month, day := found[0].GreatestEclipse.UTC().Date(); year != 2024 || month != time.April || day != 8 {
		t.Fatalf("candidate=%v, want 2024-04-08", found[0].GreatestEclipse.UTC())
	}
	// 区间起点晚于该场食甚：不得再把上一场带回来。
	window := SolarEclipseCandidates(after, after.Add(24*time.Hour), SolarEclipseCandidateOptions{})
	if len(window) != 0 {
		t.Fatalf("2024-04-09..04-10 candidates=%d, want none (the 04-08 eclipse is out of range)", len(window))
	}

	// 食甚可以早于朔：1901-05-18 的食甚 05:33:47 早于该次朔 05:37:29。区间上界落在
	// 两者之间时，按朔月判上界的旧实现会在求值前就跳出，把这场整段漏掉。
	// The greatest eclipse can precede the new moon (1901-05-18: greatest 05:33:47 before
	// the 05:37:29 new moon). With the new moon as the upper bound the old implementation
	// broke out before evaluating it and dropped the eclipse entirely.
	greatest := time.Date(1901, time.May, 18, 5, 33, 47, 0, time.UTC)
	early := SolarEclipseCandidates(greatest.AddDate(0, 0, -3), greatest.Add(30*time.Second), SolarEclipseCandidateOptions{})
	if len(early) != 1 {
		t.Fatalf("1901-05-18 window candidates=%d, want the eclipse whose greatest precedes its new moon", len(early))
	}
	if year, month, day := early[0].GreatestEclipse.UTC().Date(); year != 1901 || month != time.May || day != 18 {
		t.Fatalf("candidate=%v, want 1901-05-18", early[0].GreatestEclipse.UTC())
	}
	if early[0].GreatestEclipse.UTC().After(greatest.Add(30 * time.Second)) {
		t.Fatalf("candidate %v is outside the requested window", early[0].GreatestEclipse.UTC())
	}

	// 反向：朔月在区间内但食甚在区间之后，同样不得返回。
	// Conversely, a new moon inside the window whose greatest falls after the window end
	// must not be returned.
	late := SolarEclipseCandidates(
		time.Date(2026, time.February, 12, 12, 2, 0, 0, time.UTC),
		time.Date(2026, time.February, 17, 12, 7, 0, 0, time.UTC),
		SolarEclipseCandidateOptions{},
	)
	if len(late) != 0 {
		t.Fatalf("2026-02 window candidates=%d (%v), want none", len(late), late[0].GreatestEclipse)
	}
}

func TestSolarEclipseCandidatesCentralOnlyUsesGlobalType(t *testing.T) {
	start := time.Date(-911, time.January, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(-908, time.January, 1, 0, 0, 0, 0, time.UTC)
	all := SolarEclipseCandidates(start, end, SolarEclipseCandidateOptions{})
	central := SolarEclipseCandidates(start, end, SolarEclipseCandidateOptions{CentralOnly: true})
	// −0909-11-15 是全食但影轴未落到地表：按食型过滤必须保留它。
	// -0909-11-15 is total while its axis misses the Earth: the type filter keeps it.
	sawMissingAxis := false
	for _, candidate := range all {
		if candidate.Type == SolarEclipseTotal && candidate.Centrality == SolarEclipseNonCentral &&
			candidate.GreatestEclipse.UTC().Format("2006-01-02") == "-0909-11-15" {
			sawMissingAxis = true
		}
	}
	if !sawMissingAxis {
		t.Fatal("fixture no longer contains the total eclipse whose axis misses the Earth")
	}
	want := make([]SolarEclipseCandidate, 0, len(all))
	for _, candidate := range all {
		if candidate.Type != SolarEclipsePartial {
			want = append(want, candidate)
		}
	}
	if len(central) != len(want) {
		t.Fatalf("CentralOnly kept %d candidates, want the %d non-partial ones", len(central), len(want))
	}
	for index := range want {
		if central[index] != want[index] {
			t.Fatalf("candidate %d = %+v, want %+v", index, central[index], want[index])
		}
	}
}
