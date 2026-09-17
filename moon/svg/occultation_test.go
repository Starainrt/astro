package svg

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/internal/occultationgeo"
	"github.com/starainrt/astro/internal/svgmap"
	"github.com/starainrt/astro/moon"
)

func TestFindStarOccultationSVGsHR4799(t *testing.T) {
	location := time.FixedZone("CST", 8*3600)
	diagrams, err := FindStarOccultationSVGs(
		time.Date(2025, 6, 5, 0, 0, 0, 0, location),
		time.Date(2025, 6, 6, 0, 0, 0, 0, location),
		hr4799StarCoordinate(),
		moon.OccultationPathOptions{Step: 5 * time.Minute, TargetSpacingKM: 200},
		StarOccultationSVGOptions{Width: 720, Height: 520},
	)
	if err != nil {
		t.Fatalf("FindStarOccultationSVGs() error = %v", err)
	}
	if len(diagrams) != 1 {
		t.Fatalf("FindStarOccultationSVGs() returned %d diagrams, want 1", len(diagrams))
	}
	diagram := diagrams[0]
	for _, want := range []string{
		`<svg`, `width="720"`, `height="520"`, "2025-06-05", "HR 4799",
		"全球掩带", "掩始", "掩甚", "掩终", "掩带宽", "UTC+8",
		`class="occultation-band"`, `class="center-line"`, `class="northern-limit"`,
		`class="southern-limit"`, "掩带范围与边界",
		`class="occultation-rise-set-boundary occultation-start-rise"`, `stroke="#d97706"`, "初掩/掩甚/终掩月升月落线",
		`class="event-marker event-greatest"`, `class="land"`,
	} {
		if !strings.Contains(diagram, want) {
			t.Fatalf("SVG missing %q", want)
		}
	}
	if got := strings.Count(diagram, `class="occultation-band"`); got != 1 {
		t.Fatalf("HR 4799 occultation-band sweep count = %d, want one compound path", got)
	}
	bandIndex := strings.Index(diagram, `class="occultation-band"`)
	curveIndex := strings.Index(diagram, `class="occultation-rise-set-boundary occultation-start-rise"`)
	if bandIndex < 0 || curveIndex < 0 || curveIndex < bandIndex {
		t.Fatalf("stellar SVG draws rise/set curves before the filled band: band=%d curve=%d", bandIndex, curveIndex)
	}
	if err := validateXML(diagram); err != nil {
		t.Fatalf("generated SVG is not valid XML: %v", err)
	}
	curveIndex = strings.Index(diagram, `class="occultation-rise-set-boundary occultation-start-rise"`)
	curveEnd := -1
	if curveIndex >= 0 {
		curveEnd = strings.Index(diagram[curveIndex:], "/>")
	}
	if curveIndex < 0 || curveEnd < 0 || strings.Contains(diagram[curveIndex:curveIndex+curveEnd], "stroke-dasharray") {
		t.Fatal("stellar rise/set phase line is rendered with a gap-producing dash pattern")
	}
}

func TestStarOccultationAutoProjectionIncludesFootprints(t *testing.T) {
	path := sampleStarOccultationPath()
	path.Greatest.Latitude = 70
	for index := range path.CenterLine {
		path.CenterLine[index].Latitude = 65
		path.NorthernLimit[index].Latitude = 70
		path.SouthernLimit[index].Latitude = 60
	}
	path.Start.Latitude = 70
	path.End.Latitude = 70
	footprintTime := path.Start.Time.Add(90 * time.Minute)
	path.Footprints = []moon.OccultationFootprint{{
		Time: footprintTime,
		Polygons: [][]moon.OccultationPathPoint{{
			{Time: footprintTime, Longitude: -10, Latitude: -10},
			{Time: footprintTime, Longitude: 10, Latitude: -10},
			{Time: footprintTime, Longitude: 0, Latitude: 10},
		}},
	}}

	if projection := resolveStarOccultationMapProjection(path, MapProjectionAuto); projection != svgmap.ProjectionEquirectangular {
		t.Fatalf("auto projection = %q, want equirectangular for cross-hemisphere footprint", projection)
	}
	diagram, err := StarOccultationPathSVG(path, StarOccultationSVGOptions{})
	if err != nil {
		t.Fatalf("StarOccultationPathSVG() error = %v", err)
	}
	if !strings.Contains(diagram, "等经纬投影") {
		t.Fatal("auto-projected SVG clipped a cross-hemisphere footprint into a polar map")
	}
}

func TestStarOccultationLegendOmitsDisabledRiseSetCurves(t *testing.T) {
	path := sampleStarOccultationPath()
	path.RiseSetCurves = nil
	diagram, err := StarOccultationPathSVG(path, StarOccultationSVGOptions{})
	if err != nil {
		t.Fatalf("StarOccultationPathSVG: %v", err)
	}
	if strings.Contains(diagram, "初掩/掩甚/终掩月升月落线") || strings.Contains(diagram, "Rise/set phase lines") {
		t.Fatal("stellar SVG legend claims disabled rise/set curves are present")
	}
}

func TestPolarOccultationLayoutSeparatesLegendAndFooter(t *testing.T) {
	layout := starOccultationSVGLayoutFor(
		StarOccultationSVGOptions{Width: 900, Height: 760},
		110,
		svgmap.ProjectionNorthPolar,
		svgmap.GeoPoint{},
	)
	legendY := layout.mapY + layout.mapHeight + 30
	if gap := layout.footerY - legendY; gap < 24 {
		t.Fatalf("polar legend/footer gap = %.1f px, want at least 24 px", gap)
	}
}

func TestPolarStarOccultationSVGSplitsGrazingBoundaryBranchChanges(t *testing.T) {
	tests := []struct {
		name       string
		date       time.Time
		star       moon.StarCoordinate
		projection MapProjection
	}{
		{
			name: "Antares south polar",
			date: time.Date(2026, time.February, 11, 0, 0, 0, 0, time.UTC),
			star: moon.StarCoordinate{
				ID: "Antares", RA: 247.3516666666667, Dec: -26.431944444444444,
				Epoch: time.Date(2000, time.January, 1, 12, 0, 0, 0, time.UTC), Frame: moon.CoordinateFrameJ2000,
				ProperMotionRACosDecMasPerYear: -10, ProperMotionDecMasPerYear: -20, ParallaxMas: 24,
			},
			projection: MapProjectionSouthPolar,
		},
		{
			name: "Regulus north polar",
			date: time.Date(2025, time.August, 23, 0, 0, 0, 0, time.UTC),
			star: moon.StarCoordinate{
				ID: "Regulus", RA: 152.09291666666667, Dec: 11.967222222222222,
				Epoch: time.Date(2000, time.January, 1, 12, 0, 0, 0, time.UTC), Frame: moon.CoordinateFrameJ2000,
				ProperMotionRACosDecMasPerYear: -248, ProperMotionDecMasPerYear: 6, ParallaxMas: 45,
			},
			projection: MapProjectionNorthPolar,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			paths, err := moon.FindStarOccultationPaths(
				test.date, test.date.Add(24*time.Hour), test.star,
				moon.OccultationPathOptions{Step: 5 * time.Minute, TargetSpacingKM: 200},
			)
			if err != nil || len(paths) != 1 {
				t.Fatalf("FindStarOccultationPaths() paths=%d err=%v, want one", len(paths), err)
			}
			path := paths[0]
			if len(path.Footprints) == 0 {
				t.Fatal("stellar path has no instantaneous footprints")
			}
			if ranges := occultationgeo.ContinuousPairedBoundaryRanges(path.NorthernLimit, path.SouthernLimit); len(ranges) < 2 {
				t.Fatalf("continuous band ranges = %d, want branch change to be split", len(ranges))
			}
			projection := internalMapProjection(test.projection)
			for _, limit := range [][]moon.OccultationPathPoint{path.NorthernLimit, path.SouthernLimit} {
				for _, segment := range starOccultationBoundarySegmentsForProjection(limit, svgmap.ClipView{Projection: projection}) {
					for index := 1; index < len(segment); index++ {
						if distance := occultationgeo.DistanceKM(segment[index-1], segment[index]); distance > occultationgeo.BoundaryBranchJumpKM+1e-6 {
							t.Fatalf("rendered boundary segment still spans %.1f km branch change", distance)
						}
					}
				}
			}
			diagram, err := StarOccultationPathSVG(path, StarOccultationSVGOptions{
				Width: 1200, Height: 800, Location: time.UTC, Projection: test.projection,
			})
			if err != nil {
				t.Fatalf("StarOccultationPathSVG() error = %v", err)
			}
			if strings.Count(diagram, `class="occultation-band"`) != 1 {
				t.Fatal("instantaneous footprints were not rendered as one compound sweep")
			}
			for _, className := range []string{`class="northern-limit"`, `class="southern-limit"`} {
				if !strings.Contains(diagram, className) {
					t.Fatalf("footprint sweep SVG missing split boundary %s", className)
				}
			}
			if err := validateXML(diagram); err != nil {
				t.Fatalf("generated SVG is not valid XML: %v", err)
			}
		})
	}
}

func TestStarOccultationPathSVGPreservesEndpointBranchFragments(t *testing.T) {
	start := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	for _, count := range []int{2, 3} {
		t.Run(fmt.Sprintf("%d points", count), func(t *testing.T) {
			path := svgEndpointBranchJumpPath(start, count)
			segments := starOccultationBoundarySegmentsForProjection(
				path.NorthernLimit, svgmap.ClipView{Projection: svgmap.ProjectionEquirectangular},
			)
			if len(segments) != 2 {
				t.Fatalf("north-limit segment count = %d, want two discontinuous fragments", len(segments))
			}
			if !segments[0][0].Time.Equal(path.Start.Time) ||
				!segments[len(segments)-1][len(segments[len(segments)-1])-1].Time.Equal(path.End.Time) {
				t.Fatal("split SVG boundary does not retain start and end samples")
			}
			for _, segment := range segments {
				for index := 1; index < len(segment); index++ {
					if distance := occultationgeo.DistanceKM(segment[index-1], segment[index]); distance > occultationgeo.BoundaryBranchJumpKM+1e-6 {
						t.Fatalf("endpoint fragment spans an impossible %.1f km jump", distance)
					}
				}
			}

			diagram, err := StarOccultationPathSVG(path, StarOccultationSVGOptions{
				Projection: MapProjectionEquirectangular,
			})
			if err != nil {
				t.Fatalf("StarOccultationPathSVG: %v", err)
			}
			if got := strings.Count(diagram, `class="occultation-band"`); got != 2 {
				t.Fatalf("endpoint-band section count = %d, want 2", got)
			}
			if got := strings.Count(diagram, `class="northern-limit"`); got != 2 {
				t.Fatalf("north-limit path count = %d, want 2", got)
			}
			if err := validateXML(diagram); err != nil {
				t.Fatalf("generated SVG is not valid XML: %v", err)
			}
		})
	}
}

func svgEndpointBranchJumpPath(start time.Time, count int) moon.StarOccultationPath {
	north := make([]moon.OccultationPathPoint, count)
	south := make([]moon.OccultationPathPoint, count)
	for index := range north {
		when := start.Add(time.Duration(index) * time.Second)
		longitude := 30.0 + float64(index)/10
		if index == 0 {
			longitude = 0
		}
		north[index] = moon.OccultationPathPoint{
			Time: when, Longitude: longitude, Latitude: 10, MoonAltitude: 20,
		}
		south[index] = moon.OccultationPathPoint{
			Time: when, Longitude: longitude, Latitude: -10, MoonAltitude: 20,
		}
	}
	return moon.StarOccultationPath{
		TargetID: "endpoint-jump", Start: north[0], Greatest: north[0], End: north[count-1],
		Complete: true, NorthernLimit: north, SouthernLimit: south, Step: time.Second,
	}
}

func TestStarOccultationPathSVGEnglishAndCustomText(t *testing.T) {
	path := sampleStarOccultationPath()
	path.TargetID = "Alpha < Beta & Gamma"
	diagram, err := StarOccultationPathSVG(path, StarOccultationSVGOptions{
		Language:      "en",
		Location:      time.UTC,
		Title:         "Custom <occultation> & title",
		SummaryText:   "Custom summary",
		GreatestText:  "Custom greatest",
		MapTitle:      "Custom map",
		ContactsTitle: "Custom events",
		FooterNote:    "Custom footer",
	})
	if err != nil {
		t.Fatalf("StarOccultationPathSVG() error = %v", err)
	}
	for _, want := range []string{
		"Custom &lt;occultation&gt; &amp; title",
		"Custom summary",
		"Custom greatest",
		"Custom map",
		"Custom events",
		"Custom footer",
		"Start",
		"Greatest",
		"End",
	} {
		if !strings.Contains(diagram, want) {
			t.Fatalf("SVG missing %q", want)
		}
	}
	if strings.Contains(diagram, "Custom <occultation>") {
		t.Fatal("custom title was not XML escaped")
	}
	if err := validateXML(diagram); err != nil {
		t.Fatalf("generated SVG is not valid XML: %v", err)
	}
}

func TestStarOccultationPathSVGIncludesAlignedTimeLabels(t *testing.T) {
	diagram, err := StarOccultationPathSVG(sampleStarOccultationPath(), StarOccultationSVGOptions{
		Location: time.UTC, TimeLabelStep: 30 * time.Minute,
	})
	if err != nil {
		t.Fatalf("StarOccultationPathSVG() error = %v", err)
	}
	for _, want := range []string{`class="occultation-time-marker"`, `>10:30</text>`, `>12:30</text>`} {
		if !strings.Contains(diagram, want) {
			t.Fatalf("stellar occultation SVG missing time marker %q", want)
		}
	}
	markers := occultationTimeMarkerPoints(
		sampleStarOccultationPath().CenterLine,
		30*time.Minute,
		time.UTC,
		[]time.Time{sampleStarOccultationPath().Start.Time, sampleStarOccultationPath().End.Time},
	)
	foundGreatestTime := false
	for _, marker := range markers {
		if marker.Time.Equal(sampleStarOccultationPath().Greatest.Time) {
			foundGreatestTime = true
			break
		}
	}
	if !foundGreatestTime {
		t.Fatal("aligned time at greatest was discarded instead of being placed below the event label")
	}
}

func TestStarOccultationPathSVGCanDisableTimeLabels(t *testing.T) {
	diagram, err := StarOccultationPathSVG(sampleStarOccultationPath(), StarOccultationSVGOptions{TimeLabelStep: -1})
	if err != nil {
		t.Fatalf("StarOccultationPathSVG() error = %v", err)
	}
	if strings.Contains(diagram, `class="occultation-time-marker"`) {
		t.Fatal("disabled occultation time labels were rendered")
	}
}

func TestStarOccultationPathSVGSplitsAntimeridian(t *testing.T) {
	diagram, err := StarOccultationPathSVG(sampleStarOccultationPath(), StarOccultationSVGOptions{})
	if err != nil {
		t.Fatalf("StarOccultationPathSVG() error = %v", err)
	}
	if got := strings.Count(diagram, `class="center-line"`); got != 2 {
		t.Fatalf("center-line segment count = %d, want 2", got)
	}
	if got := strings.Count(diagram, `class="occultation-band"`); got != 2 {
		t.Fatalf("occultation-band segment count = %d, want 2", got)
	}
	segments := starOccultationPathSegments(sampleStarOccultationPath().CenterLine)
	if len(segments) != 2 {
		t.Fatalf("path segment count = %d, want 2", len(segments))
	}
	if segments[0][len(segments[0])-1].Longitude != 180 || segments[1][0].Longitude != -180 {
		t.Fatalf("antimeridian interpolation = %.3f / %.3f, want +180 / -180",
			segments[0][len(segments[0])-1].Longitude, segments[1][0].Longitude)
	}
	if !segments[0][len(segments[0])-1].Time.Equal(segments[1][0].Time) {
		t.Fatal("antimeridian split points do not share the interpolated time")
	}
}

func TestStarOccultationPathSVGUsesDetailedPhysicalLand(t *testing.T) {
	diagram, err := StarOccultationPathSVG(sampleStarOccultationPath(), StarOccultationSVGOptions{})
	if err != nil {
		t.Fatalf("StarOccultationPathSVG() error = %v", err)
	}
	for _, want := range []string{
		`class="land-layer"`,
		`class="land"`,
		`fill="#d8d9d2"`,
		`stroke="#a6aaa4"`,
		`fill-rule="evenodd"`,
		`vector-effect="non-scaling-stroke"`,
		"Natural Earth 1:50m",
		"不含行政边界",
	} {
		if !strings.Contains(diagram, want) {
			t.Fatalf("SVG missing detailed-land marker %q", want)
		}
	}
	if got := strings.Count(diagram, `class="land"`); got != 1 {
		t.Fatalf("land path count = %d, want one compact path", got)
	}
	if strings.Contains(diagram, `fill="#000`) || strings.Contains(diagram, `fill="black"`) {
		t.Fatal("land path uses a black fill")
	}
}

func TestStarOccultationPathSVGUsesNorthPolarProjection(t *testing.T) {
	path := sampleStarOccultationPath()
	path.Greatest.Latitude = 72
	for index := range path.CenterLine {
		path.CenterLine[index].Latitude = 62 + float64(index)*4
	}
	for index := range path.NorthernLimit {
		path.NorthernLimit[index].Latitude = 68 + float64(index)*3
		path.SouthernLimit[index].Latitude = 58 + float64(index)*3
	}
	path.Start.Latitude = path.NorthernLimit[0].Latitude
	path.End.Latitude = path.NorthernLimit[len(path.NorthernLimit)-1].Latitude

	diagram, err := StarOccultationPathSVG(path, StarOccultationSVGOptions{})
	if err != nil {
		t.Fatalf("StarOccultationPathSVG() error = %v", err)
	}
	for _, want := range []string{`<circle class="map-ocean"`, `<circle class="map-frame"`, "北极方位等距投影"} {
		if !strings.Contains(diagram, want) {
			t.Fatalf("north-polar SVG missing %q", want)
		}
	}
	if err := validateXML(diagram); err != nil {
		t.Fatalf("north-polar SVG is not valid XML: %v", err)
	}
}

func TestStarOccultationBandSegmentsPreserveAsymmetricAntimeridianCrossing(t *testing.T) {
	start := time.Date(2026, 8, 2, 10, 0, 0, 0, time.UTC)
	northern := []moon.OccultationPathPoint{
		{Time: start, Longitude: 170, Latitude: 20},
		{Time: start.Add(time.Hour), Longitude: -170, Latitude: 10},
	}
	southern := []moon.OccultationPathPoint{
		{Time: start, Longitude: 150, Latitude: 0},
		{Time: start.Add(time.Hour), Longitude: 160, Latitude: -10},
	}
	segments := starOccultationBandSegments(northern, southern)
	if len(segments) != 2 {
		t.Fatalf("asymmetric antimeridian band segment count = %d, want 2", len(segments))
	}
	hasWest, hasEast := false, false
	for _, segment := range segments {
		for _, point := range segment {
			hasWest = hasWest || point.longitude < -179
			hasEast = hasEast || point.longitude > 179
		}
	}
	if !hasWest || !hasEast {
		t.Fatalf("split band does not cover both map edges: west=%v east=%v", hasWest, hasEast)
	}
}

func TestFindStarOccultationSVGsNoEvent(t *testing.T) {
	diagrams, err := FindStarOccultationSVGs(
		time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC),
		moon.StarCoordinate{
			ID:    "polar-star",
			RA:    0,
			Dec:   89,
			Epoch: time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
			Frame: moon.CoordinateFrameICRS,
		},
		moon.OccultationPathOptions{},
		StarOccultationSVGOptions{},
	)
	if err != nil {
		t.Fatalf("FindStarOccultationSVGs() error = %v", err)
	}
	if len(diagrams) != 0 {
		t.Fatalf("FindStarOccultationSVGs() returned %d diagrams, want none", len(diagrams))
	}
}

func TestStarOccultationPathSVGRejectsInvalidPath(t *testing.T) {
	path := sampleStarOccultationPath()
	path.Greatest.Longitude = math.NaN()
	_, err := StarOccultationPathSVG(path, StarOccultationSVGOptions{})
	if !errors.Is(err, ErrInvalidStarOccultationPath) {
		t.Fatalf("StarOccultationPathSVG() error = %v, want ErrInvalidStarOccultationPath", err)
	}
}

func TestStarOccultationPathSVGRejectsMalformedRiseSetCurves(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*moon.OccultationRiseSetCurve)
	}{
		{name: "phase", mutate: func(curve *moon.OccultationRiseSetCurve) {
			curve.Phase = moon.RiseSetPhase("bogus")
		}},
		{name: "direction", mutate: func(curve *moon.OccultationRiseSetCurve) {
			curve.Direction = moon.RiseSetDirection("bogus")
		}},
		{name: "coordinate", mutate: func(curve *moon.OccultationRiseSetCurve) {
			curve.Segments[0][1].Longitude = math.NaN()
		}},
		{name: "time order", mutate: func(curve *moon.OccultationRiseSetCurve) {
			curve.Segments[0][1].Time = curve.Segments[0][0].Time
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := sampleStarOccultationPath()
			curve := sampleOccultationRiseSetCurve(path.Start.Time)
			test.mutate(&curve)
			path.RiseSetCurves = []moon.OccultationRiseSetCurve{curve}
			_, err := StarOccultationPathSVG(path, StarOccultationSVGOptions{})
			if !errors.Is(err, ErrInvalidStarOccultationPath) {
				t.Fatalf("StarOccultationPathSVG() error = %v, want ErrInvalidStarOccultationPath", err)
			}
		})
	}
}

func TestStarOccultationPathSVGRejectsMalformedFootprint(t *testing.T) {
	path := sampleStarOccultationPath()
	path.Footprints = []moon.OccultationFootprint{{Time: path.Start.Time}}
	_, err := StarOccultationPathSVG(path, StarOccultationSVGOptions{})
	if !errors.Is(err, ErrInvalidStarOccultationPath) {
		t.Fatalf("StarOccultationPathSVG() error = %v, want ErrInvalidStarOccultationPath", err)
	}
}

func TestStarOccultationPathSVGCompactBandUsesClosedOutlineWithoutRawLimits(t *testing.T) {
	path := sampleStarOccultationPath()
	makeFootprint := func(when time.Time, west, east float64) moon.OccultationFootprint {
		return moon.OccultationFootprint{
			Time: when,
			Polygons: [][]moon.OccultationPathPoint{{
				{Time: when, Longitude: west, Latitude: -10},
				{Time: when, Longitude: east, Latitude: -10},
				{Time: when, Longitude: east, Latitude: 10},
				{Time: when, Longitude: west, Latitude: 10},
			}},
		}
	}
	path.BandFootprints = []moon.OccultationFootprint{
		makeFootprint(path.Start.Time.Add(time.Hour), -20, 5),
		makeFootprint(path.Start.Time.Add(2*time.Hour), -5, 20),
	}
	diagram, err := StarOccultationPathSVG(path, StarOccultationSVGOptions{})
	if err != nil {
		t.Fatalf("StarOccultationPathSVG: %v", err)
	}
	if strings.Contains(diagram, `class="northern-limit"`) || strings.Contains(diagram, `class="southern-limit"`) {
		t.Fatal("compact stellar SVG still overlays discontinuous raw limits")
	}
	pathData := planetOccultationSVGPathData(t, diagram, "occultation-band")
	if !strings.Contains(pathData, "Z") {
		t.Fatal("compact stellar SVG band is not explicitly closed")
	}
}

func TestStarOccultationPathSVGAllowsDenseAndCompactFootprints(t *testing.T) {
	path := sampleStarOccultationPath()
	when := path.Greatest.Time
	footprint := moon.OccultationFootprint{
		Time: when,
		Polygons: [][]moon.OccultationPathPoint{{
			{Time: when, Longitude: -10, Latitude: -10},
			{Time: when, Longitude: 10, Latitude: -10},
			{Time: when, Longitude: 0, Latitude: 10},
		}},
	}
	path.Footprints = []moon.OccultationFootprint{footprint}
	path.BandFootprints = []moon.OccultationFootprint{footprint}
	if _, err := StarOccultationPathSVG(path, StarOccultationSVGOptions{}); err != nil {
		t.Fatalf("StarOccultationPathSVG() rejected coexisting static and timed footprints: %v", err)
	}
}

func TestStarOccultationPathSVGRejectsMisalignedLimits(t *testing.T) {
	path := sampleStarOccultationPath()
	path.SouthernLimit[1].Time = path.SouthernLimit[1].Time.Add(time.Second)
	_, err := StarOccultationPathSVG(path, StarOccultationSVGOptions{})
	if !errors.Is(err, ErrInvalidStarOccultationPath) {
		t.Fatalf("StarOccultationPathSVG() error = %v, want ErrInvalidStarOccultationPath", err)
	}
}

func TestStarOccultationPathSVGRejectsCanvasTooSmall(t *testing.T) {
	_, err := StarOccultationPathSVG(sampleStarOccultationPath(), StarOccultationSVGOptions{Width: 1, Height: 1})
	if !errors.Is(err, ErrInvalidStarOccultationSVGOptions) {
		t.Fatalf("StarOccultationPathSVG() error = %v, want ErrInvalidStarOccultationSVGOptions", err)
	}
}

func hr4799StarCoordinate() moon.StarCoordinate {
	return moon.StarCoordinate{
		ID:                             "HR 4799",
		RA:                             189.1975,
		Dec:                            -5.831944444444,
		Epoch:                          time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
		Frame:                          moon.CoordinateFrameJ2000,
		ProperMotionRACosDecMasPerYear: -28,
		ProperMotionDecMasPerYear:      -18,
	}
}

func sampleStarOccultationPath() moon.StarOccultationPath {
	start := time.Date(2026, 8, 2, 10, 0, 0, 0, time.UTC)
	center := []moon.OccultationPathPoint{
		{Time: start, Longitude: 160, Latitude: 18, MoonAltitude: 5, WidthKM: 3200},
		{Time: start.Add(time.Hour), Longitude: 175, Latitude: 10, MoonAltitude: 35, WidthKM: 3300},
		{Time: start.Add(2 * time.Hour), Longitude: -175, Latitude: 1, MoonAltitude: 50, WidthKM: 3400},
		{Time: start.Add(3 * time.Hour), Longitude: -160, Latitude: -8, MoonAltitude: 12, WidthKM: 3300},
	}
	northern := make([]moon.OccultationPathPoint, len(center))
	southern := make([]moon.OccultationPathPoint, len(center))
	for index, point := range center {
		northern[index] = point
		northern[index].Latitude += 12
		northern[index].WidthKM = 0
		southern[index] = point
		southern[index].Latitude -= 12
		southern[index].WidthKM = 0
	}
	return moon.StarOccultationPath{
		TargetID:      "synthetic-star",
		Start:         center[0],
		Greatest:      center[1],
		End:           center[3],
		Complete:      true,
		CenterLine:    center,
		NorthernLimit: northern,
		SouthernLimit: southern,
		Step:          time.Hour,
	}
}

func sampleOccultationRiseSetCurve(start time.Time) moon.OccultationRiseSetCurve {
	return moon.OccultationRiseSetCurve{
		Phase:     moon.RiseSetPhaseStart,
		Direction: moon.RiseSetDirectionRise,
		Segments: [][]moon.OccultationPathPoint{{
			{Time: start.Add(20 * time.Minute), Longitude: 10, Latitude: 20, MoonAltitude: 0},
			{Time: start.Add(40 * time.Minute), Longitude: 12, Latitude: 21, MoonAltitude: 0},
		}},
	}
}

func validateXML(value string) error {
	decoder := xml.NewDecoder(strings.NewReader(value))
	for {
		if _, err := decoder.Token(); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
	}
}
