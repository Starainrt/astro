package basic

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/internal/geodata"
)

func TestPlanetOccultationCombinedPositionMatchesSeparateEphemerides(t *testing.T) {
	tt := occultationTimeToTT(time.Date(2025, time.February, 1, 4, 0, 48, 0, time.UTC))
	for _, planet := range []OccultationPlanet{
		OccultationMercury,
		OccultationVenus,
		OccultationMars,
		OccultationJupiter,
		OccultationSaturn,
		OccultationUranus,
		OccultationNeptune,
	} {
		config, ok := planetOccultationConfigFor(planet)
		if !ok {
			t.Fatalf("%s occultation config is unavailable", planet)
		}
		wantRA, wantDec := config.apparentRaDecN(tt, -1)
		wantDistance := config.earthDistanceN(tt, -1)
		gotRA, gotDec, distance := planetOccultationApparentPositionAndDistanceN(tt, config, -1)
		if gotRA != wantRA || gotDec != wantDec || distance != wantDistance {
			t.Fatalf("%s combined position = %.15g %.15g %.15g, want %.15g %.15g %.15g",
				planet, gotRA, gotDec, distance, wantRA, wantDec, wantDistance)
		}
	}
}

func TestPlanetOccultationEventCacheReusesStateAndContactFrames(t *testing.T) {
	config, ok := planetOccultationConfigFor(OccultationSaturn)
	if !ok {
		t.Fatal("Saturn occultation config is unavailable")
	}
	tt := occultationTimeToTT(time.Date(2024, time.August, 21, 2, 41, 36, 0, time.UTC))
	cache := newPlanetOccultationEventCache(config)

	wantOuter, wantOuterOK := planetOccultationPathFrameAt(tt, config)
	wantTotal, wantTotalOK := planetOccultationTotalPathFrameAt(tt, config)
	for iteration := 0; iteration < 2; iteration++ {
		gotOuter, gotOuterOK := cache.outerFrameAt(tt)
		gotTotal, gotTotalOK := cache.totalFrameAt(tt)
		_ = cache.riseSetContextAt(tt)
		if gotOuterOK != wantOuterOK || !occultationPathFrameGeometryEqual(gotOuter, wantOuter) {
			t.Fatalf("cached outer frame differs on iteration %d", iteration)
		}
		if gotTotalOK != wantTotalOK || !occultationPathFrameGeometryEqual(gotTotal, wantTotal) {
			t.Fatalf("cached total frame differs on iteration %d", iteration)
		}
	}
	if len(cache.states) != 1 || len(cache.outerFrames) != 1 || len(cache.totalFrames) != 1 {
		t.Fatalf("cache sizes = states:%d outer:%d total:%d, want one entry each",
			len(cache.states), len(cache.outerFrames), len(cache.totalFrames))
	}
}

func TestOccultationPathFramesReuseMoonDistanceForAngularRadius(t *testing.T) {
	tt := occultationTimeToTT(time.Date(2025, time.February, 1, 4, 0, 48, 0, time.UTC))
	want := MoonSemidiameter(tt) * math.Pi / (180 * 3600)
	starFrame, ok := starOccultationPathFrameAt(tt, StarCoordinate{
		RA: 0, Dec: 0, Epoch: time.Date(2000, time.January, 1, 12, 0, 0, 0, time.UTC), Frame: CoordinateFrameJ2000,
	})
	if !ok {
		t.Fatal("stellar occultation frame is unavailable")
	}
	if difference := math.Abs(starFrame.moonRadius - want); difference > 1e-15 {
		t.Fatalf("stellar cached lunar radius differs by %.15g radians", difference)
	}

	config, _ := planetOccultationConfigFor(OccultationSaturn)
	planetFrame, ok := planetOccultationPathFrameAt(tt, config)
	if !ok {
		t.Fatal("planet occultation frame is unavailable")
	}
	if difference := math.Abs(planetFrame.moonRadius - want); difference > 1e-15 {
		t.Fatalf("planet cached lunar radius differs by %.15g radians", difference)
	}
}

func TestPlanetOccultationCanDisableInstantaneousFootprints(t *testing.T) {
	start := time.Date(2024, time.August, 21, 0, 0, 0, 0, time.UTC)
	paths, err := FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), OccultationSaturn,
		OccultationPathOptions{Step: 10 * time.Minute, DisableFootprints: true},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	path := paths[0]
	if len(path.PartialFootprints) != 0 || len(path.TotalFootprints) != 0 {
		t.Fatalf("disabled footprint counts partial=%d total=%d, want zero", len(path.PartialFootprints), len(path.TotalFootprints))
	}
	if len(path.PartialBandFootprints) == 0 || len(path.TotalBandFootprints) == 0 {
		t.Fatalf("compact band support counts partial=%d total=%d, want both nonzero", len(path.PartialBandFootprints), len(path.TotalBandFootprints))
	}
	if len(path.RiseSetCurves) != 6 {
		t.Fatalf("rise/set curve count=%d, want six", len(path.RiseSetCurves))
	}
	if len(path.TotalRiseSetCurves) != 6 {
		t.Fatalf("total rise/set curve count=%d, want six", len(path.TotalRiseSetCurves))
	}
	if len(path.CenterLine) == 0 || len(path.NorthernLimit) == 0 || len(path.SouthernLimit) == 0 {
		t.Fatal("disabled footprints removed the center line or outer limits")
	}
	if !path.HasTotalBand || len(path.NorthernTotalLimit) == 0 || len(path.SouthernTotalLimit) == 0 {
		t.Fatal("disabled footprints removed the total-occultation band")
	}
}

func TestPlanetOccultationCompactBandCanIncludeLowFrequencyTimeline(t *testing.T) {
	start := time.Date(2025, time.January, 5, 0, 0, 0, 0, time.FixedZone("UTC+8", 8*60*60))
	paths, err := FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), OccultationSaturn,
		OccultationPathOptions{
			Step: 20 * time.Minute, DisableFootprints: true,
			IncludeFootprintTimeline: true, FootprintTimelineStep: 5 * time.Minute,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("paths=%d err=%v, want one", len(paths), err)
	}
	path := paths[0]
	if len(path.PartialBandFootprints) == 0 || len(path.PartialFootprints) == 0 {
		t.Fatalf("partial compact/timeline counts=%d/%d, want both", len(path.PartialBandFootprints), len(path.PartialFootprints))
	}
	if len(path.TotalBandFootprints) == 0 || len(path.TotalFootprints) == 0 {
		t.Fatalf("total compact/timeline counts=%d/%d, want both", len(path.TotalBandFootprints), len(path.TotalFootprints))
	}
	if len(path.PartialFootprints) > 50 || len(path.TotalFootprints) > 50 {
		t.Fatalf("five-minute timeline is unexpectedly dense: partial=%d total=%d", len(path.PartialFootprints), len(path.TotalFootprints))
	}
	timelinePoints := 0
	for _, footprints := range [][]PlanetOccultationFootprint{path.PartialFootprints, path.TotalFootprints} {
		for _, footprint := range footprints {
			for polygonIndex, polygon := range footprint.Polygons {
				if len(polygon) < 4 || occultationPathDistanceKM(polygon[0], polygon[len(polygon)-1]) > 0.001 {
					t.Fatalf("timeline footprint at %v polygon %d is not closed", footprint.Time, polygonIndex)
				}
				timelinePoints += len(polygon)
				for pointIndex := 1; pointIndex < len(polygon); pointIndex++ {
					if distance := occultationPathDistanceKM(polygon[pointIndex-1], polygon[pointIndex]); distance > 301 {
						t.Fatalf("timeline footprint at %v polygon %d edge %d spans %.1f km, want at most 301 km",
							footprint.Time, polygonIndex, pointIndex-1, distance)
					}
				}
			}
		}
	}
	if timelinePoints > 15000 {
		t.Fatalf("timeline contains %d polygon points, want at most 15000", timelinePoints)
	}
}

func TestStarOccultationCompactBandUsesIndependentTimeline(t *testing.T) {
	start := time.Date(2025, time.June, 5, 0, 0, 0, 0, time.UTC)
	paths, err := FindStarOccultationPaths(
		start, start.Add(24*time.Hour), hr4799OccultationCoordinateForTest(),
		OccultationPathOptions{
			Step: 20 * time.Minute, TargetSpacingKM: 900,
			DisableFootprints: true, IncludeFootprintTimeline: true,
			FootprintTimelineStep: 5 * time.Minute,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindStarOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	path := paths[0]
	if len(path.BandFootprints) == 0 || len(path.Footprints) == 0 {
		t.Fatalf("compact/timeline counts=%d/%d, want both", len(path.BandFootprints), len(path.Footprints))
	}
	for _, footprint := range path.Footprints {
		for polygonIndex, polygon := range footprint.Polygons {
			if len(polygon) < 4 || occultationPathDistanceKM(polygon[0], polygon[len(polygon)-1]) > 0.001 {
				t.Fatalf("timeline footprint at %v polygon %d is not closed", footprint.Time, polygonIndex)
			}
			for pointIndex := 1; pointIndex < len(polygon); pointIndex++ {
				if distance := occultationPathDistanceKM(polygon[pointIndex-1], polygon[pointIndex]); distance > 301 {
					t.Fatalf("timeline footprint at %v polygon %d edge %d spans %.1f km, want at most 301 km",
						footprint.Time, polygonIndex, pointIndex-1, distance)
				}
			}
		}
	}
}

func TestOccultationPathPointBudgetIsBounded(t *testing.T) {
	options := normalizeOccultationPathOptions(OccultationPathOptions{})
	got := occultationPathEstimatedPointCount(0, 0.2, 0, 0, false, 0, 0, false, 0.1, options)
	if got <= 0 || got > occultationPathMaxOutputPointCount {
		t.Fatalf("default occultation point estimate=%d, want within positive budget", got)
	}
	dense := normalizeOccultationPathOptions(OccultationPathOptions{Step: time.Second})
	got = occultationPathEstimatedPointCount(0, 2, 0, 2, true, 0, 2, true, 1, dense)
	if got <= occultationPathFootprintPointBudget || got > occultationPathMaxOutputPointCount {
		t.Fatalf("dense occultation point estimate=%d, want footprint-aware value within budget %d",
			got, occultationPathMaxOutputPointCount)
	}
	if overflow := occultationPathAccumulatePointEstimate(occultationPathMaxOutputPointCount-10, 20); overflow <= occultationPathMaxOutputPointCount {
		t.Fatalf("overflow estimate=%d, want sentinel above %d", overflow, occultationPathMaxOutputPointCount)
	}
}

func TestPlanetOccultation19621010TotalBandIsNarrowerThanOuterBand(t *testing.T) {
	start := time.Date(1962, time.October, 10, 0, 0, 0, 0, time.UTC)
	paths, err := FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), OccultationJupiter,
		OccultationPathOptions{Step: 10 * time.Minute, DisableFootprints: true, DisableRiseSet: true},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	path := paths[0]
	if !path.HasTotalBand || path.GreatestTotalWidthKM <= 0 || path.GreatestTotalWidthKM >= path.Greatest.WidthKM {
		t.Fatalf("widths outer=%.3f total=%.3f hasTotal=%v, want a narrower positive total band",
			path.Greatest.WidthKM, path.GreatestTotalWidthKM, path.HasTotalBand)
	}
}

func TestPlanetOccultation20240725CompactBandRefinesContactsAndBoundaryPairing(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2024, time.July, 25, 0, 0, 0, 0, zone)
	paths, err := FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), OccultationSaturn,
		OccultationPathOptions{
			Step: 20 * time.Minute, TargetSpacingKM: 900,
			DisableFootprints: true,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	path := paths[0]
	for _, test := range []struct {
		name       string
		footprints []PlanetOccultationFootprint
		start, end time.Time
	}{
		{name: "partial", footprints: path.PartialBandFootprints, start: path.Start.Time, end: path.End.Time},
		{name: "total", footprints: path.TotalBandFootprints, start: path.TotalStart.Time, end: path.TotalEnd.Time},
	} {
		if len(test.footprints) < 2 {
			t.Fatalf("%s compact support count=%d, want at least two", test.name, len(test.footprints))
		}
		if gap := test.footprints[0].Time.Sub(test.start); gap > 30*time.Second {
			t.Errorf("%s compact support starts %s after contact, want at most 30s", test.name, gap)
		}
		if gap := test.end.Sub(test.footprints[len(test.footprints)-1].Time); gap > 30*time.Second {
			t.Errorf("%s compact support ends %s before contact, want at most 30s", test.name, gap)
		}
		for footprintIndex, footprint := range test.footprints {
			if footprintIndex > 0 {
				previous := test.footprints[footprintIndex-1]
				if previous.Closed && footprint.Closed && footprint.Time.Sub(previous.Time) <= time.Second {
					t.Errorf("%s compact support retains duplicate closed footprints at %s and %s",
						test.name, previous.Time, footprint.Time)
				}
			}
			for polygonIndex, polygon := range footprint.Polygons {
				for pointIndex := 1; pointIndex < len(polygon); pointIndex++ {
					if distance := occultationPathDistanceKM(polygon[pointIndex-1], polygon[pointIndex]); distance > 150 {
						t.Errorf("%s compact support[%d].polygon[%d] edge %d spans %.1f km, want at most 150 km",
							test.name, footprintIndex, polygonIndex, pointIndex-1, distance)
					}
				}
			}
		}
	}
	for _, test := range []struct {
		name          string
		first, second []OccultationPathPoint
	}{
		{name: "partial", first: path.NorthernLimit, second: path.SouthernLimit},
		{name: "total", first: path.NorthernTotalLimit, second: path.SouthernTotalLimit},
	} {
		for index := 1; index < len(test.first); index++ {
			direct := math.Max(
				occultationPathDistanceKM(test.first[index-1], test.first[index]),
				occultationPathDistanceKM(test.second[index-1], test.second[index]),
			)
			swapped := math.Max(
				occultationPathDistanceKM(test.first[index-1], test.second[index]),
				occultationPathDistanceKM(test.second[index-1], test.first[index]),
			)
			if direct > 2000 && swapped < 750 {
				t.Errorf("%s boundary sample %d keeps a %.1f km direct pairing although the swapped pairing is %.1f km",
					test.name, index, direct, swapped)
			}
		}
	}
}

func TestPlanetOccultation20250105CompactBandRefinesVisibilityTransitions(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2025, time.January, 5, 0, 0, 0, 0, zone)
	paths, err := FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), OccultationSaturn,
		OccultationPathOptions{
			Step: 20 * time.Minute, TargetSpacingKM: 900,
			DisableRiseSet: true, DisableFootprints: true,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	for _, test := range []struct {
		name       string
		footprints []PlanetOccultationFootprint
	}{
		{name: "partial", footprints: paths[0].PartialBandFootprints},
		{name: "total", footprints: paths[0].TotalBandFootprints},
	} {
		transitions := 0
		for index := 1; index < len(test.footprints); index++ {
			previous, current := test.footprints[index-1], test.footprints[index]
			if previous.Closed == current.Closed {
				continue
			}
			transitions++
			if gap := current.Time.Sub(previous.Time); gap > 150*time.Millisecond {
				t.Errorf("%s visibility transition %d spans %s, want at most 150ms", test.name, transitions, gap)
			}
			open := previous
			if open.Closed {
				open = current
			}
			if len(open.Boundaries) != 1 || len(open.Boundaries[0]) < 2 {
				t.Fatalf("%s visibility transition %d has no open boundary", test.name, transitions)
			}
			// Boundaries contain the visible contact arc, not a closed ring. Its
			// endpoints may remain far apart at the geocentric open/closed
			// transition because station parallax changes the topology. The physical
			// contract is that both endpoints lie on the lunar horizon and the
			// separately exported polygon closes them with the horizon arc.
			boundary := open.Boundaries[0]
			for _, endpoint := range []OccultationPathPoint{boundary[0], boundary[len(boundary)-1]} {
				if math.Abs(endpoint.MoonAltitude) > 1e-5 {
					t.Errorf("%s visibility transition %d endpoint altitude=%g deg, want horizon root",
						test.name, transitions, endpoint.MoonAltitude)
				}
			}
			if len(open.Polygons) == 0 || len(open.Polygons[0]) < len(boundary)+2 {
				t.Errorf("%s visibility transition %d has no horizon-closed polygon", test.name, transitions)
			}
		}
		if transitions != 2 {
			t.Errorf("%s visibility transition count=%d, want 2", test.name, transitions)
		}
	}
}

func TestPlanetOccultationFiniteDiskExpandsOuterAndContractsTotalPath(t *testing.T) {
	config, ok := planetOccultationConfigFor(OccultationSaturn)
	if !ok {
		t.Fatal("Saturn occultation config is unavailable")
	}
	tt := occultationTimeToTT(time.Date(2024, time.August, 21, 2, 41, 36, 0, time.UTC))
	frameAt := func(tt float64) (occultationPathFrame, bool) {
		return planetOccultationPathFrameAt(tt, config)
	}
	_, _, finiteWidth, finiteOK := occultationPathLimitsAndWidthForFrame(tt, frameAt)
	if !finiteOK {
		t.Fatal("finite-disk path limits are unavailable")
	}
	pointFrameAt := func(tt float64) (occultationPathFrame, bool) {
		frame, valid := planetOccultationPathFrameAt(tt, config)
		frame.targetRadius = 0
		return frame, valid
	}
	_, _, pointWidth, pointOK := occultationPathLimitsAndWidthForFrame(tt, pointFrameAt)
	if !pointOK {
		t.Fatal("point-source comparison limits are unavailable")
	}
	if finiteWidth <= pointWidth {
		t.Fatalf("finite-disk outer width = %.6f km, want greater than point-source width %.6f km", finiteWidth, pointWidth)
	}
	if finiteWidth-pointWidth < 1 {
		t.Fatalf("finite-disk expansion = %.6f km, want a measurable planetary-radius contribution", finiteWidth-pointWidth)
	}
	innerFrameAt := func(tt float64) (occultationPathFrame, bool) {
		return planetOccultationTotalPathFrameAt(tt, config)
	}
	_, _, totalWidth, totalOK := occultationPathLimitsAndWidthForFrame(tt, innerFrameAt)
	if !totalOK {
		t.Fatal("finite-disk total-occultation limits are unavailable")
	}
	if totalWidth >= pointWidth {
		t.Fatalf("finite-disk total width = %.6f km, want less than point-source width %.6f km", totalWidth, pointWidth)
	}
	if pointWidth-totalWidth < 1 {
		t.Fatalf("finite-disk contraction = %.6f km, want a measurable planetary-radius contribution", pointWidth-totalWidth)
	}
}

func TestPlanetOccultationConesUseTwoSphereCommonTangents(t *testing.T) {
	config, ok := planetOccultationConfigFor(OccultationSaturn)
	if !ok {
		t.Fatal("Saturn occultation config is unavailable")
	}
	tt := occultationTimeToTT(time.Date(2025, time.February, 1, 4, 0, 48, 0, time.UTC))
	outer, ok := planetOccultationPathFrameAt(tt, config)
	if !ok {
		t.Fatal("Saturn outer-contact cone is unavailable")
	}
	inner, ok := planetOccultationTotalPathFrameAt(tt, config)
	if !ok {
		t.Fatal("Saturn inner-contact cone is unavailable")
	}

	planetRA, planetDec := config.apparentRaDecN(tt, -1)
	planetDistance := config.earthDistanceN(tt, -1) * occultationPathAstronomicalUnitKM
	target := occultationPathRaDecVector(planetRA, planetDec, planetDistance)
	moonToTargetDistance := occultationPathNorm(occultationPathSub(target, outer.moon))
	moonRadiusKM := occultationPathNorm(outer.moon) * math.Sin(outer.moonRadius)
	wantOuter := math.Asin((moonRadiusKM + config.equatorialRadiusKM) / moonToTargetDistance)
	wantInner := math.Asin((moonRadiusKM - config.equatorialRadiusKM) / moonToTargetDistance)
	if difference := math.Abs(outer.targetRadius - wantOuter); difference > 1e-15 {
		t.Fatalf("outer-contact cone angle = %.15g rad, want %.15g (difference %.3g)", outer.targetRadius, wantOuter, difference)
	}
	if difference := math.Abs(inner.targetRadius - wantInner); difference > 1e-15 {
		t.Fatalf("inner-contact cone angle = %.15g rad, want %.15g (difference %.3g)", inner.targetRadius, wantInner, difference)
	}
	for _, contact := range []struct {
		name  string
		frame occultationPathFrame
	}{
		{name: "outer", frame: outer},
		{name: "inner", frame: inner},
	} {
		origin, direction, rayOK := occultationPathBoundaryRay(contact.frame, 0.73)
		if !rayOK {
			t.Fatalf("%s-contact boundary ray is unavailable", contact.name)
		}
		moonNormal := occultationPathSub(origin, contact.frame.moon)
		if difference := math.Abs(occultationPathNorm(moonNormal) - moonRadiusKM); difference > 1e-6 {
			t.Fatalf("%s-contact lunar tangency radius differs by %.9f km", contact.name, difference)
		}
		if residual := math.Abs(occultationPathDot(moonNormal, direction)); residual > 1e-6 {
			t.Fatalf("%s-contact ray/lunar-radius dot product = %.9f km", contact.name, residual)
		}
		targetParameter := occultationPathDot(occultationPathSub(target, origin), direction)
		targetTangent := occultationPathAdd(origin, occultationPathScale(direction, targetParameter))
		targetNormal := occultationPathSub(targetTangent, target)
		if difference := math.Abs(occultationPathNorm(targetNormal) - config.equatorialRadiusKM); difference > 1e-5 {
			t.Fatalf("%s-contact planetary tangency radius differs by %.9f km", contact.name, difference)
		}
		if residual := math.Abs(occultationPathDot(targetNormal, direction)); residual > 1e-5 {
			t.Fatalf("%s-contact ray/planet-radius dot product = %.9f km", contact.name, residual)
		}
	}
}

func TestPlanetOccultationInnerConeUsesSignedTargetRadius(t *testing.T) {
	config, ok := planetOccultationConfigFor(OccultationSaturn)
	if !ok {
		t.Fatal("Saturn occultation config is unavailable")
	}
	tt := occultationTimeToTT(time.Date(2025, time.February, 1, 4, 0, 48, 0, time.UTC))
	frame, ok := planetOccultationTotalPathFrameAt(tt, config)
	if !ok {
		t.Fatal("Saturn inner-contact cone is unavailable")
	}

	for index := 0; index < occultationPathBoundaryScanPoints; index++ {
		theta := 2 * math.Pi * float64(index) / float64(occultationPathBoundaryScanPoints)
		want, _, wantOK := occultationPathBoundaryVector(frame, theta)
		if !wantOK {
			continue
		}
		discriminant, _, scale, lineOK := occultationPathBoundaryLine(frame, theta)
		if !lineOK || discriminant < 0 {
			continue
		}
		got, _, gotOK := occultationPathBoundaryIntersection(frame, theta, 1e-12*math.Max(scale, 1))
		if !gotOK {
			t.Fatalf("signed inner-cone intersection is unavailable at theta %.9f", theta)
		}
		if difference := occultationPathNorm(occultationPathSub(got, want)); difference > 1e-6 {
			t.Fatalf("inner-cone intersection differs by %.6f km at theta %.9f", difference, theta)
		}
		return
	}
	t.Fatal("no comparable Saturn inner-cone boundary point found")
}

func TestOccultationPathBoundaryTangentFindsBetweenSamples(t *testing.T) {
	const boundaryRadiusKM = 1737.4
	theta := math.Pi / float64(occultationPathBoundaryScanPoints)
	offset := geodata.EarthEquatorialRadiusKM + boundaryRadiusKM - 0.01
	moon := occultationPathVector{
		x: 384000,
		y: -offset * math.Cos(theta),
		z: -offset * math.Sin(theta),
	}
	frame := occultationPathFrame{
		moon:       moon,
		axis:       occultationPathVector{x: -1},
		first:      occultationPathVector{y: 1},
		second:     occultationPathVector{z: 1},
		moonRadius: math.Asin(boundaryRadiusKM / occultationPathNorm(moon)),
	}
	for _, sampledTheta := range []float64{0, 2 * math.Pi / float64(occultationPathBoundaryScanPoints)} {
		if _, _, ok := occultationPathBoundaryVector(frame, sampledTheta); ok {
			t.Fatalf("fixture is not narrower than the old sample spacing at theta %.9f", sampledTheta)
		}
	}
	point, tangentTheta, ok := occultationPathBoundaryTangent(frame)
	if !ok {
		t.Fatal("continuous boundary tangency was not found between scan points")
	}
	if math.Abs(tangentTheta-theta) > 5e-5 {
		t.Fatalf("tangent theta = %.9f, want %.9f", tangentTheta, theta)
	}
	polarRatioSquared := occultationPathEarthPolarRatio * occultationPathEarthPolarRatio
	ellipsoidResidual := point.x*point.x + point.y*point.y + point.z*point.z/polarRatioSquared -
		geodata.EarthEquatorialRadiusKM*geodata.EarthEquatorialRadiusKM
	if math.Abs(ellipsoidResidual) > 1e-3 {
		t.Fatalf("tangent point ellipsoid residual = %.9f", ellipsoidResidual)
	}
	frameAt := func(float64) (occultationPathFrame, bool) { return frame, true }
	if _, _, centerOK := occultationEarthLineIntersection(frame.moon, frame.axis); centerOK {
		t.Fatal("synthetic center line unexpectedly intersects Earth")
	}
	north, south, width, limitsOK := occultationPathLimitsAndWidthForFrame(2451545, frameAt)
	if !limitsOK {
		t.Fatal("boundary-only event did not produce path limits")
	}
	if separation := occultationPathNorm(occultationPathSub(north, south)); separation <= 1e-6 {
		t.Fatalf("boundary-only path limits collapsed to one point: separation=%.12f km", separation)
	}
	if width <= 0 {
		t.Fatalf("boundary-only path width = %.12f km, want positive", width)
	}
	greatest, greatestOK := occultationPathBoundaryPointForFrame(2451545, frameAt, time.UTC)
	if !greatestOK {
		t.Fatal("boundary-only event did not produce a greatest surface point")
	}
	if greatest.WidthKM <= 0 {
		t.Fatalf("boundary-only greatest width = %.12f km, want positive", greatest.WidthKM)
	}
}

func TestPlanetOccultationSaturnLimitsRemainContinuous(t *testing.T) {
	location := time.FixedZone("UTC+8", 8*3600)
	paths, err := FindPlanetOccultationPaths(
		time.Date(2025, time.February, 1, 0, 0, 0, 0, location),
		time.Date(2025, time.February, 2, 0, 0, 0, 0, location),
		OccultationSaturn,
		OccultationPathOptions{Step: 2 * time.Minute},
	)
	if err != nil {
		t.Fatalf("FindPlanetOccultationPaths() error = %v", err)
	}
	if len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() returned %d paths, want 1", len(paths))
	}

	for _, limit := range []struct {
		name   string
		points []OccultationPathPoint
	}{
		{name: "outer northern", points: paths[0].NorthernLimit},
		{name: "outer southern", points: paths[0].SouthernLimit},
		{name: "total northern", points: paths[0].NorthernTotalLimit},
		{name: "total southern", points: paths[0].SouthernTotalLimit},
	} {
		for index := 1; index < len(limit.points); index++ {
			distance := occultationPathDistanceKM(limit.points[index-1], limit.points[index])
			if distance > 1000 {
				t.Fatalf("%s limit jumps %.1f km between %v and %v", limit.name, distance,
					limit.points[index-1].Time, limit.points[index].Time)
			}
		}
	}
}

func TestRefinedPlanetOccultationCenterLineRespectsWidthTolerance(t *testing.T) {
	start := time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC)
	paths, err := FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), OccultationSaturn,
		OccultationPathOptions{Step: 5 * time.Minute, TargetSpacingKM: 50},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	config, ok := planetOccultationConfigFor(OccultationSaturn)
	if !ok {
		t.Fatal("Saturn occultation config is unavailable")
	}
	frameAt := func(tt float64) (occultationPathFrame, bool) {
		return planetOccultationPathFrameAt(tt, config)
	}
	for index, point := range paths[0].CenterLine {
		exact, pointOK := occultationPathCenterPointForFrame(centerTimeTT(point.Time), frameAt, time.UTC)
		if !pointOK {
			t.Fatalf("exact center point %d is unavailable", index)
		}
		if difference := math.Abs(point.WidthKM - exact.WidthKM); difference > occultationPathWidthToleranceKM {
			t.Fatalf("center point %d width differs from exact value by %.9f km: got %.9f want %.9f",
				index, difference, point.WidthKM, exact.WidthKM)
		}
	}
}

func TestPlanetOccultationSaturnLimitsDoNotDependOnStep(t *testing.T) {
	location := time.FixedZone("UTC+8", 8*3600)
	start := time.Date(2024, time.August, 21, 0, 0, 0, 0, location)
	end := time.Date(2024, time.August, 22, 0, 0, 0, 0, location)
	fine := findSinglePlanetOccultationPath(t, start, end, 30*time.Second)
	coarse := findSinglePlanetOccultationPath(t, start, end, 2*time.Minute)

	for _, limits := range []struct {
		name         string
		fine, coarse []OccultationPathPoint
	}{
		{name: "outer northern", fine: fine.NorthernLimit, coarse: coarse.NorthernLimit},
		{name: "outer southern", fine: fine.SouthernLimit, coarse: coarse.SouthernLimit},
		{name: "total northern", fine: fine.NorthernTotalLimit, coarse: coarse.NorthernTotalLimit},
		{name: "total southern", fine: fine.SouthernTotalLimit, coarse: coarse.SouthernTotalLimit},
	} {
		assertOccultationPathCommonSamplesEqual(t, limits.name, limits.fine, limits.coarse)
		for index := 1; index+1 < len(limits.coarse); index++ {
			paired := coarse.SouthernLimit
			if strings.HasPrefix(limits.name, "total") {
				paired = coarse.SouthernTotalLimit
			}
			if strings.HasSuffix(limits.name, "southern") {
				continue
			}
			if distance := occultationPathDistanceKM(limits.coarse[index], paired[index]); distance < 0.001 {
				t.Fatalf("%s and southern limit collapse at %v", limits.name, limits.coarse[index].Time)
			}
		}
	}
}

func findSinglePlanetOccultationPath(t *testing.T, start, end time.Time, step time.Duration) PlanetOccultationPath {
	t.Helper()
	paths, err := FindPlanetOccultationPaths(start, end, OccultationSaturn, OccultationPathOptions{Step: step})
	if err != nil {
		t.Fatalf("FindPlanetOccultationPaths(step=%v) error = %v", step, err)
	}
	if len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths(step=%v) returned %d paths, want 1", step, len(paths))
	}
	if !paths[0].HasTotalBand {
		t.Fatalf("FindPlanetOccultationPaths(step=%v) has no total band", step)
	}
	return paths[0]
}

func assertOccultationPathCommonSamplesEqual(t *testing.T, name string, fine, coarse []OccultationPathPoint) {
	t.Helper()
	matched := 0
	fineIndex := 0
	for _, coarsePoint := range coarse[1 : len(coarse)-1] {
		for fineIndex+1 < len(fine) && fine[fineIndex].Time.Before(coarsePoint.Time.Add(-20*time.Millisecond)) {
			fineIndex++
		}
		nearest := -1
		nearestDelta := math.Inf(1)
		for candidateIndex := fineIndex - 2; candidateIndex <= fineIndex+2; candidateIndex++ {
			if candidateIndex < 0 || candidateIndex >= len(fine) {
				continue
			}
			delta := math.Abs(fine[candidateIndex].Time.Sub(coarsePoint.Time).Seconds())
			if delta < nearestDelta {
				nearest = candidateIndex
				nearestDelta = delta
			}
		}
		if nearest < 0 || nearestDelta > 0.00001 {
			continue
		}
		matched++
		if distance := occultationPathDistanceKM(fine[nearest], coarsePoint); distance > 5 {
			t.Fatalf("%s differs by %.1f km at common time %v (sample delta %.6f s)",
				name, distance, coarsePoint.Time, nearestDelta)
		}
	}
	if matched < 10 {
		t.Fatalf("%s compared only %d common samples, want at least 10", name, matched)
	}
}

func TestPlanetOccultation20250114MarsBandContourTimesIncrease(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2025, time.January, 14, 0, 0, 0, 0, zone)
	paths, err := FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), OccultationMars,
		OccultationPathOptions{
			Step: 20 * time.Minute, TargetSpacingKM: 900,
			DisableFootprints: true, IncludeFootprintTimeline: true,
			FootprintTimelineStep: 5 * time.Minute, RiseSetStep: time.Minute,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	for _, band := range []struct {
		name     string
		contours [][]OccultationPathPoint
	}{
		{name: "partial", contours: paths[0].PartialBandContours},
		{name: "total", contours: paths[0].TotalBandContours},
	} {
		for contourIndex, contour := range band.contours {
			for pointIndex := 1; pointIndex < len(contour); pointIndex++ {
				previous, current := contour[pointIndex-1], contour[pointIndex]
				if !current.Time.After(previous.Time) {
					t.Fatalf(
						"%s contour %d times do not increase at %d: %s then %s (delta=%s, distance=%.6f km)",
						band.name, contourIndex, pointIndex,
						previous.Time.Format(time.RFC3339Nano), current.Time.Format(time.RFC3339Nano),
						current.Time.Sub(previous.Time), occultationPathDistanceKM(previous, current),
					)
				}
			}
		}
	}
}
