package eclipse

import (
	"testing"
	"time"
)

func TestSolarSarosInfoAgainstNASAExamples(t *testing.T) {
	t.Run("2024 Apr 08 total", func(t *testing.T) {
		info := ClosestSolarEclipse(time.Date(2024, 4, 8, 12, 0, 0, 0, time.UTC))
		assertSarosInfo(t, info.HasSaros, info.Saros, 139, 30, 71)
	})

	t.Run("1501 May 17 first member", func(t *testing.T) {
		info := ClosestSolarEclipse(time.Date(1501, 5, 17, 12, 0, 0, 0, time.UTC))
		assertSarosInfo(t, info.HasSaros, info.Saros, 139, 1, 71)
	})

	t.Run("2763 Jul 03 last member", func(t *testing.T) {
		info := ClosestSolarEclipse(time.Date(2763, 7, 3, 12, 0, 0, 0, time.UTC))
		assertSarosInfo(t, info.HasSaros, info.Saros, 139, 71, 71)
	})

	t.Run("series 22 edge-range member", func(t *testing.T) {
		info := ClosestSolarEclipse(time.Date(-1994, 9, 13, 12, 0, 0, 0, time.UTC))
		assertSarosInfo(t, info.HasSaros, info.Saros, 22, 11, 71)
	})
}

func TestSolarSarosInfoDerivedOutsideCatalogRange(t *testing.T) {
	for _, tc := range []struct {
		name string
		date time.Time
	}{
		{name: "before catalog", date: time.Date(-3200, time.January, 1, 12, 0, 0, 0, time.UTC)},
		{name: "after catalog", date: time.Date(3288, time.November, 15, 12, 0, 0, 0, time.UTC)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			info := ClosestSolarEclipse(tc.date)
			if !info.HasSaros || info.Saros.Verified {
				t.Fatalf("expected provisional solar Saros: has=%v saros=%+v", info.HasSaros, info.Saros)
			}
			if info.Saros.Member < 1 || info.Saros.Member > info.Saros.Count || info.Saros.Count < 1 {
				t.Fatalf("invalid provisional solar Saros range: %+v", info.Saros)
			}
		})
	}
}

func TestSarosInfoDerivedAtFarEpochs(t *testing.T) {
	for _, tc := range []struct {
		name string
		date time.Time
	}{
		{name: "solar 6000 BCE", date: time.Date(-6000, time.January, 1, 12, 0, 0, 0, time.UTC)},
		{name: "solar 8000 BCE", date: time.Date(-8000, time.January, 1, 12, 0, 0, 0, time.UTC)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			info := ClosestSolarEclipse(tc.date)
			if !info.HasSaros || info.Saros.Verified || info.Saros.Member < 1 || info.Saros.Member > info.Saros.Count {
				t.Fatalf("expected far-epoch provisional solar Saros: has=%v info=%+v", info.HasSaros, info.Saros)
			}
		})
	}
}

func TestLocalSolarSarosMatchesGlobal(t *testing.T) {
	date := time.Date(2009, 7, 22, 12, 0, 0, 0, time.FixedZone("CST", 8*3600))
	global := ClosestSolarEclipse(date)
	local, ok := LocalSolarEclipseOnDate(date, 121.9850, 30.6167, 0)
	if !ok {
		t.Fatal("expected a visible local solar eclipse")
	}
	if !global.HasSaros || !local.HasSaros {
		t.Fatalf("expected both global and local solar eclipses to have Saros info: global=%v local=%v", global.HasSaros, local.HasSaros)
	}
	if global.Saros != local.Saros {
		t.Fatalf("local solar Saros mismatch: got %+v want %+v", local.Saros, global.Saros)
	}
}

func TestLunarSarosInfoAgainstNASAExamples(t *testing.T) {
	t.Run("2025 Mar 14 total", func(t *testing.T) {
		info := ClosestLunarEclipse(time.Date(2025, 3, 14, 12, 0, 0, 0, time.UTC))
		assertSarosInfo(t, info.HasSaros, info.Saros, 123, 53, 72)
	})

	t.Run("1087 Aug 16 first member", func(t *testing.T) {
		info := ClosestLunarEclipse(time.Date(1087, 8, 16, 12, 0, 0, 0, time.UTC))
		assertSarosInfo(t, info.HasSaros, info.Saros, 123, 1, 72)
	})

	t.Run("2367 Oct 08 last member", func(t *testing.T) {
		info := ClosestLunarEclipse(time.Date(2367, 10, 8, 12, 0, 0, 0, time.UTC))
		assertSarosInfo(t, info.HasSaros, info.Saros, 123, 72, 72)
	})

	t.Run("series 4 edge-range member", func(t *testing.T) {
		info := ClosestLunarEclipse(time.Date(-1997, 10, 31, 12, 0, 0, 0, time.UTC))
		assertSarosInfo(t, info.HasSaros, info.Saros, 4, 30, 78)
	})

	t.Run("series 8 edge-range member", func(t *testing.T) {
		info := ClosestLunarEclipse(time.Date(-1989, 6, 6, 12, 0, 0, 0, time.UTC))
		assertSarosInfo(t, info.HasSaros, info.Saros, 8, 29, 86)
	})

	t.Run("series 61 mid-series member", func(t *testing.T) {
		info := ClosestLunarEclipse(time.Date(14, 4, 4, 12, 0, 0, 0, time.UTC))
		assertSarosInfo(t, info.HasSaros, info.Saros, 61, 45, 78)
	})

	t.Run("series 61 shallow first member default", func(t *testing.T) {
		info := ClosestLunarEclipse(time.Date(-780, 12, 13, 12, 0, 0, 0, time.UTC))
		assertSarosInfo(t, info.HasSaros, info.Saros, 61, 1, 78)
	})
}

func TestLunarSarosInfoDerivedOutsideCatalogRange(t *testing.T) {
	for _, tc := range []struct {
		name string
		date time.Time
	}{
		{name: "before catalog", date: time.Date(-3200, time.January, 1, 12, 0, 0, 0, time.UTC)},
		{name: "after catalog", date: time.Date(3300, time.January, 1, 12, 0, 0, 0, time.UTC)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			info := ClosestLunarEclipse(tc.date)
			if !info.HasSaros || info.Saros.Verified {
				t.Fatalf("expected provisional lunar Saros: has=%v saros=%+v", info.HasSaros, info.Saros)
			}
			if info.Saros.Member < 1 || info.Saros.Member > info.Saros.Count || info.Saros.Count < 1 {
				t.Fatalf("invalid provisional lunar Saros range: %+v", info.Saros)
			}
		})
	}
}

func TestLunarSarosShallowFirstMemberChauvenet(t *testing.T) {
	info := ClosestLunarEclipseChauvenet(time.Date(-780, 12, 13, 12, 0, 0, 0, time.UTC))
	assertSarosInfo(t, info.HasSaros, info.Saros, 61, 1, 78)
}

func TestLocalLunarSarosMatchesGlobal(t *testing.T) {
	date := time.Date(2025, 3, 14, 12, 0, 0, 0, time.FixedZone("CDT", -5*3600))
	global := ClosestLunarEclipse(date)
	local, ok := LocalLunarEclipseOnDate(date, -95.3698, 29.7604, 0)
	if !ok {
		t.Fatal("expected a visible local lunar eclipse")
	}
	if !global.HasSaros || !local.HasSaros {
		t.Fatalf("expected both global and local lunar eclipses to have Saros info: global=%v local=%v", global.HasSaros, local.HasSaros)
	}
	if global.Saros != local.Saros {
		t.Fatalf("local lunar Saros mismatch: got %+v want %+v", local.Saros, global.Saros)
	}
}

func TestSolarPathAndFootprintsCarrySaros(t *testing.T) {
	date := time.Date(2024, 4, 8, 12, 0, 0, 0, time.UTC)
	global := ClosestSolarEclipse(date)

	path, ok := SolarEclipseCentralPath(date, SolarEclipsePathOptions{})
	if !ok {
		t.Fatal("expected central path data")
	}
	assertSarosInfo(t, path.Eclipse.HasSaros, path.Eclipse.Saros, global.Saros.Series, global.Saros.Member, global.Saros.Count)

	footprints, ok := SolarEclipsePartialFootprints(date, SolarEclipsePartialFootprintOptions{})
	if !ok {
		t.Fatal("expected partial footprints data")
	}
	assertSarosInfo(t, footprints.Eclipse.HasSaros, footprints.Eclipse.Saros, global.Saros.Series, global.Saros.Member, global.Saros.Count)
}

func TestSarosAnchorSanity(t *testing.T) {
	assertSarosAnchorTable(t, solarSarosAnchors[:], 0)
	assertSarosAnchorTable(t, lunarSarosAnchors[:], 1)
	assertSarosHeadOverrides(t, solarSarosHeadOverrides[:], solarSarosAnchors[:], 0)
	assertSarosHeadOverrides(t, lunarSarosHeadOverrides[:], lunarSarosAnchors[:], 1)
}

func assertSarosInfo(t *testing.T, has bool, got SarosInfo, wantSeries, wantMember, wantCount int) {
	t.Helper()
	if !has {
		t.Fatal("expected Saros info")
	}
	if got.Series != wantSeries || got.Member != wantMember || got.Count != wantCount {
		t.Fatalf(
			"unexpected Saros info: got {Series:%d Member:%d Count:%d} want {Series:%d Member:%d Count:%d}",
			got.Series,
			got.Member,
			got.Count,
			wantSeries,
			wantMember,
			wantCount,
		)
	}
	if !got.Verified {
		t.Fatalf("expected authoritative Saros metadata: got %+v", got)
	}
}

func assertSarosAnchorTable(t *testing.T, anchors []sarosMagic, seriesBase int) {
	t.Helper()
	if len(anchors) == 0 {
		t.Fatal("expected non-empty Saros anchor table")
	}
	seenDates := make(map[[3]int]int, len(anchors))
	lastSeries := seriesBase - 1
	for index, magic := range anchors {
		anchor := decodeSarosMagic(magic, seriesBase+index)
		series := int(anchor.Series)
		if series <= lastSeries {
			t.Fatalf("series not strictly increasing: prev=%d current=%d", lastSeries, series)
		}
		lastSeries = series
		if anchor.Count == 0 || int(anchor.Count) > sarosExtrapolationWindow {
			t.Fatalf("unexpected anchor count for series %d: %d", series, anchor.Count)
		}
		dateKey := [3]int{int(anchor.Year), int(anchor.Month), int(anchor.Day)}
		if previous, ok := seenDates[dateKey]; ok {
			t.Fatalf("duplicate Saros head date %v for series %d and %d", dateKey, previous, series)
		}
		seenDates[dateKey] = series
	}
	if got := int(decodeSarosMagic(anchors[0], seriesBase).Series); got != seriesBase {
		t.Fatalf("unexpected first series: got %d want %d", got, seriesBase)
	}
}

func assertSarosHeadOverrides(t *testing.T, overrides []sarosHeadOverride, anchors []sarosMagic, seriesBase int) {
	t.Helper()
	if len(overrides) == 0 {
		return
	}
	seenHeads := make(map[[3]int]int, len(overrides))
	anchorSeries := make(map[int]int, len(anchors))
	for index, magic := range anchors {
		anchor := decodeSarosMagic(magic, seriesBase+index)
		anchorSeries[int(anchor.Series)] = int(anchor.Count)
	}
	for _, override := range overrides {
		key := [3]int{int(override.HeadYear), int(override.HeadMonth), int(override.HeadDay)}
		if previous, ok := seenHeads[key]; ok {
			t.Fatalf("duplicate Saros override head date %v for series %d and %d", key, previous, override.Series)
		}
		seenHeads[key] = int(override.Series)
		count, ok := anchorSeries[int(override.Series)]
		if !ok {
			t.Fatalf("override references unknown series %d", override.Series)
		}
		if count != int(override.Count) {
			t.Fatalf("override count mismatch for series %d: got %d want %d", override.Series, override.Count, count)
		}
	}
}

func TestLunarSarosOverrideHeadIsAuthoritative(t *testing.T) {
	// 月食 4 号在锚表与 override 里各有一个头，相差 9 个沙罗周期（−2646-10-06 与
	// −2483-01-12 / MemberOffset +2）。两套格点同时生效时，相隔 127 年的两场月食都报
	// Series=4、Member=3、Verified=true（重复成员号）。override 必须是唯一权威：窗口外
	// 的那场按文档只报"没有沙罗信息"，而不是另一套编号。
	// Lunar series 4 has two heads nine cycles apart (-2646-10-06 in the anchor table and
	// -2483-01-12 with MemberOffset +2 as an override). With both lattices live, two
	// eclipses 127 years apart both reported Series=4, Member=3, Verified=true. The
	// override must be the only authority: the out-of-window eclipse reports no Saros
	// metadata instead of a second numbering.
	anchored := ClosestLunarEclipse(time.Date(-2483, time.January, 11, 0, 0, 0, 0, time.UTC))
	if !anchored.HasSaros {
		t.Fatal("the override head eclipse must keep its Saros metadata")
	}
	if anchored.Saros.Series != 4 || anchored.Saros.Member != 3 || !anchored.Saros.Verified {
		t.Fatalf("override head eclipse Saros=%+v, want series 4 member 3 verified", anchored.Saros)
	}

	earlier := ClosestLunarEclipse(time.Date(-2610, time.October, 27, 0, 0, 0, 0, time.UTC))
	if earlier.HasSaros &&
		earlier.Saros.Series == anchored.Saros.Series &&
		earlier.Saros.Member == anchored.Saros.Member {
		t.Fatalf("two eclipses 127 years apart share series %d member %d",
			earlier.Saros.Series, earlier.Saros.Member)
	}
}
