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
	assertSarosAnchorTable(t, solarSarosAnchors[:], true)
	assertSarosAnchorTable(t, lunarSarosAnchors[:], false)
	assertSarosHeadOverrides(t, solarSarosHeadOverrides[:], solarSarosAnchors[:])
	assertSarosHeadOverrides(t, lunarSarosHeadOverrides[:], lunarSarosAnchors[:])
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
}

func assertSarosAnchorTable(t *testing.T, anchors []sarosAnchor, solar bool) {
	t.Helper()
	if len(anchors) == 0 {
		t.Fatal("expected non-empty Saros anchor table")
	}
	seenDates := make(map[[3]int]int, len(anchors))
	lastSeries := int(anchors[0].Series) - 1
	for _, anchor := range anchors {
		series := int(anchor.Series)
		if series <= lastSeries {
			t.Fatalf("series not strictly increasing: prev=%d current=%d", lastSeries, series)
		}
		lastSeries = series
		if anchor.Count == 0 || int(anchor.Count) >= sarosWalkLimit {
			t.Fatalf("unexpected anchor count for series %d: %d", series, anchor.Count)
		}
		dateKey := [3]int{int(anchor.Year), int(anchor.Month), int(anchor.Day)}
		if previous, ok := seenDates[dateKey]; ok {
			t.Fatalf("duplicate Saros head date %v for series %d and %d", dateKey, previous, series)
		}
		seenDates[dateKey] = series
	}
	if solar {
		if got := int(anchors[0].Series); got != 0 {
			t.Fatalf("unexpected first solar series: got %d want 0", got)
		}
	} else {
		if got := int(anchors[0].Series); got != 1 {
			t.Fatalf("unexpected first lunar series: got %d want 1", got)
		}
	}
}

func assertSarosHeadOverrides(t *testing.T, overrides []sarosHeadOverride, anchors []sarosAnchor) {
	t.Helper()
	if len(overrides) == 0 {
		return
	}
	seenHeads := make(map[[3]int]int, len(overrides))
	anchorSeries := make(map[int]int, len(anchors))
	for _, anchor := range anchors {
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
