package basic

import (
	"errors"
	"math"
	"testing"
	"time"
)

func TestPlanetOccultationHorizonArcStaysBetweenContacts(t *testing.T) {
	config, _ := planetOccultationConfigFor(OccultationSaturn)
	tt := occultationTimeToTT(time.Date(2025, 1, 5, 17, 0, 0, 0, time.UTC))
	frame, ok := planetOccultationPathFrameAt(tt, config)
	if !ok {
		t.Fatal("missing frame")
	}
	fine := planetOccultationHorizonCircle(tt, frame, time.UTC, 3600)
	for _, tc := range []struct {
		from, to int
		want     []int
	}{
		{20, 180, []int{100, 180}},
		{180, 20, []int{100, 20}},
		{3590, 130, []int{0, 100, 130}},
		{130, 3590, []int{100, 0, 3590}},
		{1790, 1930, []int{1800, 1900, 1930}},
		{1930, 1790, []int{1900, 1800, 1790}},
		{20, 30, []int{25, 30}},
		{30, 20, []int{25, 20}},
	} {
		arc := planetOccultationHorizonArc(tt, frame, []OccultationPathPoint{fine[tc.to], fine[tc.from]}, time.UTC, 36)
		if len(arc) != len(tc.want) {
			t.Errorf("arc %d -> %d has %d points, want %d", tc.from, tc.to, len(arc), len(tc.want))
			continue
		}
		for i, index := range tc.want {
			if distance := occultationPathDistanceKM(arc[i], fine[index]); distance > 1e-4 {
				t.Errorf("arc %d -> %d point %d misses circle sample %d by %.6f km", tc.from, tc.to, i, index, distance)
			}
		}
	}
}

func TestPlanetOccultationSaturnFootprintsContainCenterLine(t *testing.T) {
	config, ok := planetOccultationConfigFor(OccultationSaturn)
	if !ok {
		t.Fatal("Saturn occultation config is unavailable")
	}
	seedTT := occultationTimeToTT(time.Date(2025, time.February, 1, 4, 0, 48, 0, time.UTC))
	for _, contact := range []struct {
		name    string
		frameAt occultationPathFrameFunc
	}{
		{name: "partial", frameAt: func(tt float64) (occultationPathFrame, bool) {
			return planetOccultationPathFrameAt(tt, config)
		}},
		{name: "total", frameAt: func(tt float64) (occultationPathFrame, bool) {
			return planetOccultationTotalPathFrameAt(tt, config)
		}},
	} {
		startTT, endTT, found := occultationPathWindowForFrame(
			seedTT, seedTT-occultationPathSearchSpanDays, seedTT+occultationPathSearchSpanDays,
			contact.frameAt, true,
		)
		if !found {
			t.Fatalf("%s center interval is unavailable", contact.name)
		}
		checked := 0
		for tt := startTT + 2.0/1440.0; tt < endTT-2.0/1440.0; tt += 5.0 / 1440.0 {
			frame, frameOK := contact.frameAt(tt)
			center, _, centerOK := occultationEarthLineIntersection(frame.moon, frame.axis)
			footprint, footprintOK := planetOccultationFootprintAt(tt, contact.frameAt, time.UTC)
			if !frameOK || !centerOK || !footprintOK {
				t.Fatalf("%s center or footprint is unavailable at TT %.9f", contact.name, tt)
			}
			longitude, latitude := occultationPathGeodetic(tt, center)
			if !planetOccultationFootprintContains(footprint, longitude, latitude) {
				t.Fatalf("%s footprint does not contain center %.6f, %.6f at TT %.9f",
					contact.name, longitude, latitude, tt)
			}
			checked++
		}
		if checked < 10 {
			t.Fatalf("%s checked only %d center samples", contact.name, checked)
		}
	}
}

func TestPlanetOccultationMars20250729GreatestIsInsideTotalFootprint(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2025, time.July, 29, 0, 0, 0, 0, zone)
	paths, err := FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), OccultationMars,
		OccultationPathOptions{
			Step: 20 * time.Minute, TargetSpacingKM: 900,
			RiseSetStep: time.Minute, DisableFootprints: true,
			IncludeFootprintTimeline: true, FootprintTimelineStep: 5 * time.Minute,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	path := paths[0]
	if !path.HasTotalBand || len(path.TotalBandFootprints) == 0 {
		t.Fatalf("Mars path has no total band support: hasTotal=%v footprints=%d",
			path.HasTotalBand, len(path.TotalBandFootprints))
	}
	greatest := path.Greatest
	foundGreatestFootprint := false
	for _, footprint := range path.TotalBandFootprints {
		if !footprint.Time.Equal(greatest.Time) {
			continue
		}
		foundGreatestFootprint = true
		if !planetOccultationFootprintContains(footprint, greatest.Longitude, greatest.Latitude) {
			t.Fatalf("greatest point %.6f, %.6f is outside total footprint at %v",
				greatest.Longitude, greatest.Latitude, greatest.Time)
		}
	}
	if !foundGreatestFootprint {
		t.Fatalf("total compact support omitted greatest time %v", greatest.Time)
	}
}

func TestPlanetOccultationPathRejectsExcessiveAggregateSampling(t *testing.T) {
	start := time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC)
	_, err := FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), OccultationSaturn,
		OccultationPathOptions{Step: time.Second},
	)
	if !errors.Is(err, ErrOccultationPathSamplingLimit) {
		t.Fatalf("FindPlanetOccultationPaths() error = %v, want ErrOccultationPathSamplingLimit", err)
	}
}

func TestPlanetOccultationFootprintsHaveIndependentSampleBudget(t *testing.T) {
	times := occultationPathSampleTimesWithLimit(
		0, 1, 0.500001, 1.0/86400.0, planetOccultationFootprintMaxSamples,
	)
	if len(times) > planetOccultationFootprintMaxSamples {
		t.Fatalf("footprint sample count = %d, maximum %d", len(times), planetOccultationFootprintMaxSamples)
	}
	foundGreatest := false
	for _, sample := range times {
		if sample == 0.500001 {
			foundGreatest = true
			break
		}
	}
	if !foundGreatest {
		t.Fatal("bounded footprint samples omitted greatest")
	}
}

func TestPlanetOccultationHorizonCircleUsesTopocentricLunarHorizon(t *testing.T) {
	config, ok := planetOccultationConfigFor(OccultationSaturn)
	if !ok {
		t.Fatal("Saturn occultation config is unavailable")
	}
	tt := occultationTimeToTT(time.Date(2024, time.July, 24, 16, 30, 0, 0, time.UTC))
	frame, ok := planetOccultationTotalPathFrameAt(tt, config)
	if !ok {
		t.Fatal("Saturn total-occultation frame is unavailable")
	}
	points := planetOccultationHorizonCircle(tt, frame, time.UTC, 360)
	if len(points) != 360 {
		t.Fatalf("horizon point count = %d, want 360", len(points))
	}
	for index, point := range points {
		if altitude := math.Abs(point.MoonAltitude); altitude > 5e-5 {
			t.Fatalf("horizon point %d lunar altitude = %.12f degrees, want zero", index, point.MoonAltitude)
		}
	}
}

func planetOccultationFootprintContains(
	footprint PlanetOccultationFootprint,
	longitude, latitude float64,
) bool {
	for _, polygon := range footprint.Polygons {
		inside := false
		for current, previous := 0, len(polygon)-1; current < len(polygon); previous, current = current, current+1 {
			currentX := math.Remainder(polygon[current].Longitude-longitude, 360)
			previousX := math.Remainder(polygon[previous].Longitude-longitude, 360)
			currentY := polygon[current].Latitude
			previousY := polygon[previous].Latitude
			if math.Abs(currentX-previousX) > 180 {
				if currentX < previousX {
					currentX += 360
				} else {
					previousX += 360
				}
			}
			if (currentY > latitude) == (previousY > latitude) {
				continue
			}
			intersectionX := previousX + (latitude-previousY)*(currentX-previousX)/(currentY-previousY)
			if intersectionX > 0 {
				inside = !inside
			}
		}
		if inside {
			return true
		}
	}
	return false
}

// 导出的瞬时足迹必须停在月球地平线切口上：边界、闭合多边形与修复面都不得含地平线以下顶点。
func TestPlanetOccultationFootprintsClipAtLunarHorizon(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	saturnStart := time.Date(2025, time.January, 5, 0, 0, 0, 0, zone)
	for _, testCase := range []struct {
		name    string
		options OccultationPathOptions
	}{
		{
			name: "DenseFootprints",
			options: OccultationPathOptions{
				Algorithm: OccultationPathAlgorithmExact, Step: 20 * time.Minute,
				TargetSpacingKM: 900, RiseSetStep: time.Minute,
			},
		},
		{
			name: "TimelineFootprints",
			options: OccultationPathOptions{
				Algorithm: OccultationPathAlgorithmExact, Step: 20 * time.Minute,
				TargetSpacingKM: 900, DisableFootprints: true, IncludeFootprintTimeline: true,
				FootprintTimelineStep: 5 * time.Minute, RiseSetStep: time.Minute,
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			paths, err := FindPlanetOccultationPaths(
				saturnStart, saturnStart.Add(24*time.Hour), OccultationSaturn, testCase.options,
			)
			if err != nil || len(paths) != 1 {
				t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
			}
			path := paths[0]
			below, checked := 0, 0
			for _, footprints := range [][]PlanetOccultationFootprint{path.PartialFootprints, path.TotalFootprints} {
				belowPart, checkedPart := planetOccultationFootprintHorizonViolations(footprints)
				below += belowPart
				checked += checkedPart
			}
			if checked < 1000 {
				t.Fatalf("checked only %d footprint vertices", checked)
			}
			if below != 0 {
				t.Fatalf("%d of %d footprint vertices lie below the lunar horizon", below, checked)
			}
		})
	}
}

func planetOccultationFootprintHorizonViolations(
	footprints []PlanetOccultationFootprint,
) (below, checked int) {
	for _, footprint := range footprints {
		for _, rings := range [][][]OccultationPathPoint{
			footprint.Boundaries, footprint.Polygons, footprint.InteriorPolygons,
		} {
			for _, ring := range rings {
				for _, point := range ring {
					checked++
					if point.MoonAltitude < 0 {
						below++
					}
				}
			}
		}
	}
	return below, checked
}

func BenchmarkOccultationPlanetFootprints(b *testing.B) {
	b.Run("Saturn20240821Default", func(b *testing.B) {
		benchmarkPlanetOccultationFootprints(
			b, OccultationSaturn, time.Date(2024, time.August, 21, 0, 0, 0, 0, time.UTC),
			OccultationPathOptions{},
		)
	})
	b.Run("Mars20250729Compact", func(b *testing.B) {
		benchmarkPlanetOccultationFootprints(
			b, OccultationMars, time.Date(2025, time.July, 29, 0, 0, 0, 0, time.FixedZone("UTC+8", 8*60*60)),
			OccultationPathOptions{
				Step: 20 * time.Minute, TargetSpacingKM: 900, RiseSetStep: time.Minute,
				DisableFootprints: true, IncludeFootprintTimeline: true, FootprintTimelineStep: 5 * time.Minute,
			},
		)
	})
}

// benchmarkPlanetOccultationFootprints 只度量足迹装配：窗口与帧函数取自一次真实路径求解，避免混入路径与升落成本。
func benchmarkPlanetOccultationFootprints(
	b *testing.B,
	planet OccultationPlanet,
	start time.Time,
	options OccultationPathOptions,
) {
	paths, err := FindPlanetOccultationPaths(start, start.Add(24*time.Hour), planet, options)
	if err != nil || len(paths) != 1 {
		b.Fatalf("paths=%d err=%v", len(paths), err)
	}
	config, ok := planetOccultationConfigFor(planet)
	if !ok {
		b.Fatal("missing occultation config")
	}
	path := paths[0]
	cache := newPlanetOccultationEventCache(config)
	greatestTT := occultationTimeToTT(path.Greatest.Time)
	cache.preparePathEphemeris(greatestTT, options.Algorithm)
	startTT, endTT := occultationTimeToTT(path.Start.Time), occultationTimeToTT(path.End.Time)
	location := start.Location()
	frameAt := cache.outerFrameAt
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		if options.DisableFootprints {
			band := planetOccultationBandFootprints(startTT, endTT, greatestTT, frameAt, location, nil)
			timeline := planetOccultationTimelineFootprints(startTT, endTT, greatestTT, frameAt, options, location)
			if len(band) == 0 || len(timeline) == 0 {
				b.Fatalf("band=%d timeline=%d", len(band), len(timeline))
			}
			continue
		}
		if footprints := planetOccultationFootprints(startTT, endTT, greatestTT, frameAt, options, location); len(footprints) == 0 {
			b.Fatal("no footprints")
		}
	}
}
