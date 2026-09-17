package svg

import (
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/internal/occultationgeo"
	"github.com/starainrt/astro/moon"
)

// 本节钉住入口契约：进入几何引擎的每一组轮廓都要先校验；引擎失败属于内部错误，
// 用非哨兵错误返回（不得谎报数据无效），也不能静默改画原始足迹。

func TestStarOccultationPathSVGRejectsMalformedBandContours(t *testing.T) {
	base := sampleStarOccultationPath().Start.Time
	valid := func() [][]moon.OccultationPathPoint {
		return [][]moon.OccultationPathPoint{{
			reviewPoint(base, 10*time.Minute, 110, 10),
			reviewPoint(base, 20*time.Minute, 120, 10),
		}}
	}
	tests := []struct {
		name   string
		mutate func([][]moon.OccultationPathPoint) [][]moon.OccultationPathPoint
	}{
		{name: "nan longitude", mutate: func(contours [][]moon.OccultationPathPoint) [][]moon.OccultationPathPoint {
			contours[0][1].Longitude = math.NaN()
			return contours
		}},
		{name: "duplicate times", mutate: func(contours [][]moon.OccultationPathPoint) [][]moon.OccultationPathPoint {
			contours[0][1].Time = contours[0][0].Time
			return contours
		}},
		{name: "single point", mutate: func([][]moon.OccultationPathPoint) [][]moon.OccultationPathPoint {
			return [][]moon.OccultationPathPoint{{reviewPoint(base, 10*time.Minute, 110, 10)}}
		}},
		{name: "empty contour", mutate: func([][]moon.OccultationPathPoint) [][]moon.OccultationPathPoint {
			return [][]moon.OccultationPathPoint{{}}
		}},
		{name: "outside event time", mutate: func(contours [][]moon.OccultationPathPoint) [][]moon.OccultationPathPoint {
			contours[0][0].Time = base.Add(-time.Minute)
			return contours
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := sampleStarOccultationPath()
			path.BandContours = test.mutate(valid())
			if _, err := StarOccultationPathSVG(path, StarOccultationSVGOptions{}); !errors.Is(err, ErrInvalidStarOccultationPath) {
				t.Fatalf("StarOccultationPathSVG() error = %v, want ErrInvalidStarOccultationPath", err)
			}
		})
	}
}

func TestStarOccultationPathSVGAcceptsValidBandContours(t *testing.T) {
	path := sampleStarOccultationPath()
	base := path.Start.Time
	path.BandContours = [][]moon.OccultationPathPoint{{
		reviewPoint(base, 10*time.Minute, 110, 10),
		reviewPoint(base, 20*time.Minute, 120, 12),
		reviewPoint(base, 30*time.Minute, 130, 14),
	}}
	path.VisibilityContours = [][]moon.OccultationPathPoint{{
		reviewPoint(base, 12*time.Minute, 112, 11),
		reviewPoint(base, 24*time.Minute, 124, 13),
	}}
	if _, err := StarOccultationPathSVG(path, StarOccultationSVGOptions{}); err != nil {
		t.Fatalf("StarOccultationPathSVG() rejected valid contours: %v", err)
	}
}

func TestPlanetOccultationPathSVGRejectsMalformedContours(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*moon.PlanetOccultationPath)
	}{
		{name: "partial band contour time order", mutate: func(path *moon.PlanetOccultationPath) {
			base := path.Start.Time
			path.PartialBandContours = [][]moon.OccultationPathPoint{{
				reviewPoint(base, 20*time.Minute, 120, 10),
				reviewPoint(base, 10*time.Minute, 110, 10),
			}}
		}},
		{name: "partial visibility contour nan", mutate: func(path *moon.PlanetOccultationPath) {
			base := path.Start.Time
			path.PartialVisibilityContours = [][]moon.OccultationPathPoint{{
				reviewPoint(base, 10*time.Minute, 110, 10),
				reviewPoint(base, 20*time.Minute, 120, math.NaN()),
			}}
		}},
		{name: "total band contour outside total interval", mutate: func(path *moon.PlanetOccultationPath) {
			base := path.TotalStart.Time
			path.TotalBandContours = [][]moon.OccultationPathPoint{{
				reviewPoint(base, -time.Minute, 100, 8),
				reviewPoint(base, 10*time.Minute, 110, 8),
			}}
		}},
		{name: "total visibility contour point count", mutate: func(path *moon.PlanetOccultationPath) {
			path.TotalVisibilityContours = [][]moon.OccultationPathPoint{
				{reviewPoint(path.TotalStart.Time, time.Minute, 100, 8)},
			}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := samplePlanetOccultationPath()
			test.mutate(&path)
			if _, err := PlanetOccultationPathSVG(path, PlanetOccultationSVGOptions{}); !errors.Is(err, ErrInvalidPlanetOccultationPath) {
				t.Fatalf("PlanetOccultationPathSVG() error = %v, want ErrInvalidPlanetOccultationPath", err)
			}
		})
	}
}

func TestPlanetOccultationPathSVGAcceptsValidContours(t *testing.T) {
	path := samplePlanetOccultationPath()
	start, totalStart := path.Start.Time, path.TotalStart.Time
	path.PartialBandContours = [][]moon.OccultationPathPoint{{
		reviewPoint(start, 10*time.Minute, 110, 10),
		reviewPoint(start, 20*time.Minute, 120, 12),
	}}
	path.PartialVisibilityContours = [][]moon.OccultationPathPoint{{
		reviewPoint(start, 15*time.Minute, 115, 10),
		reviewPoint(start, 25*time.Minute, 125, 12),
	}}
	path.TotalBandContours = [][]moon.OccultationPathPoint{{
		reviewPoint(totalStart, 5*time.Minute, 105, 8),
		reviewPoint(totalStart, 15*time.Minute, 115, 8),
	}}
	path.TotalVisibilityContours = [][]moon.OccultationPathPoint{{
		reviewPoint(totalStart, 6*time.Minute, 106, 8),
		reviewPoint(totalStart, 16*time.Minute, 116, 8),
	}}
	if _, err := PlanetOccultationPathSVG(path, PlanetOccultationSVGOptions{}); err != nil {
		t.Fatalf("PlanetOccultationPathSVG() rejected valid contours: %v", err)
	}
}

func TestStarOccultationPathSVGReportsBandGeometryError(t *testing.T) {
	path := sampleStarOccultationPath()
	path.RiseSetCurves = []moon.OccultationRiseSetCurve{sampleOccultationRiseSetCurve(path.Start.Time)}
	path.BandFootprints = []moon.OccultationFootprint{reviewFootprint(path.Greatest.Time)}
	path.BandContours = [][]moon.OccultationPathPoint{{
		reviewPoint(path.Start.Time, 10*time.Minute, 110, 0),
		reviewPoint(path.Start.Time, 10*time.Minute+time.Second, -80, 0),
	}}
	diagram, err := StarOccultationPathSVG(path, StarOccultationSVGOptions{})
	if err == nil {
		t.Fatal("StarOccultationPathSVG() accepted a band geometry failure")
	}
	if errors.Is(err, ErrInvalidStarOccultationPath) {
		t.Fatalf("geometry engine failure = %v, must not be reported as invalid path data", err)
	}
	if !strings.Contains(err.Error(), "band geometry") {
		t.Fatalf("geometry engine failure = %v, want a render-stage context", err)
	}
	if diagram != "" {
		t.Fatal("stellar renderer returned a diagram after a band geometry failure")
	}
}

func TestPlanetOccultationPathSVGReportsBandGeometryError(t *testing.T) {
	path := samplePlanetOccultationPath()
	totalStart := path.TotalStart.Time
	path.TotalRiseSetCurves = []moon.OccultationRiseSetCurve{
		reviewRiseSetCurve(totalStart.Add(8*time.Minute), totalStart.Add(18*time.Minute)),
	}
	path.TotalBandFootprints = []moon.OccultationFootprint{reviewFootprint(totalStart.Add(10 * time.Minute))}
	path.TotalBandContours = [][]moon.OccultationPathPoint{{
		reviewPoint(totalStart, time.Minute, 100, 0),
		reviewPoint(totalStart, time.Minute+time.Second, -80, 0),
	}}
	diagram, err := PlanetOccultationPathSVG(path, PlanetOccultationSVGOptions{})
	if err == nil {
		t.Fatal("PlanetOccultationPathSVG() accepted a band geometry failure")
	}
	if errors.Is(err, ErrInvalidPlanetOccultationPath) {
		t.Fatalf("geometry engine failure = %v, must not be reported as invalid path data", err)
	}
	if !strings.Contains(err.Error(), "band geometry") {
		t.Fatalf("geometry engine failure = %v, want a render-stage context", err)
	}
	if diagram != "" {
		t.Fatal("planetary renderer returned a diagram after a band geometry failure")
	}
}

func TestStarOccultationLegendOmitsBoundaryClaimForSweepFallback(t *testing.T) {
	path := sampleStarOccultationPath()
	path.BandFootprints = []moon.OccultationFootprint{
		reviewFootprint(path.Start.Time.Add(time.Hour)),
		reviewFootprint(path.Start.Time.Add(2 * time.Hour)),
	}
	diagram, err := StarOccultationPathSVG(path, StarOccultationSVGOptions{})
	if err != nil {
		t.Fatalf("StarOccultationPathSVG: %v", err)
	}
	if strings.Contains(diagram, "掩带范围与边界") {
		t.Fatal("non-authoritative sweep fallback still claims the analytic band boundary")
	}
	if !strings.Contains(diagram, "掩带范围") {
		t.Fatal("non-authoritative sweep fallback is missing the band extent legend entry")
	}
}

func TestStarOccultationLegendClaimsBoundaryForFootprintSweep(t *testing.T) {
	path := starOccultationSimplificationPath(t)
	diagram, err := StarOccultationPathSVG(path, StarOccultationSVGOptions{Width: 720, Height: 520})
	if err != nil {
		t.Fatalf("StarOccultationPathSVG: %v", err)
	}
	if !strings.Contains(diagram, "掩带范围与边界") {
		t.Fatal("stellar SVG legend dropped the band-and-limits entry for the default footprint sweep")
	}
}

func TestStarOccultationSVGDrawsHorizonConnectorsFromResolvedFootprints(t *testing.T) {
	path := reviewHorizonConnectorPath()
	connectors := occultationgeo.StarHorizonConnectorSegments(
		path.Footprints, path.RiseSetCurves, path.NorthernLimit, path.SouthernLimit)
	if len(connectors) == 0 {
		t.Fatal("fixture does not yield a horizon connector for the drawn footprints")
	}
	if compact := occultationgeo.StarHorizonConnectorSegments(
		path.BandFootprints, path.RiseSetCurves, path.NorthernLimit, path.SouthernLimit); len(compact) != 0 {
		t.Fatalf("fixture unexpectedly yields %d compact-band connectors", len(compact))
	}
	diagram, err := StarOccultationPathSVG(path, StarOccultationSVGOptions{})
	if err != nil {
		t.Fatalf("StarOccultationPathSVG: %v", err)
	}
	if !strings.Contains(diagram, `class="occultation-rise-set-boundary occultation-horizon-`) {
		t.Fatal("stellar SVG dropped the horizon connectors derived from the drawn footprints")
	}
}

func TestPlanetOccultationSVGDrawsTotalRiseSetCurves(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2025, time.July, 29, 0, 0, 0, 0, zone)
	paths, err := moon.FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), moon.OccultationMars,
		moon.OccultationPathOptions{
			Step: 20 * time.Minute, TargetSpacingKM: 900, RiseSetStep: time.Minute,
			DisableFootprints: true, IncludeFootprintTimeline: true,
			FootprintTimelineStep: 5 * time.Minute,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	if len(paths[0].TotalRiseSetCurves) == 0 {
		t.Fatal("fixture has no inner-contact rise/set curves")
	}
	diagram, err := PlanetOccultationPathSVG(paths[0], PlanetOccultationSVGOptions{
		Width: 1200, Height: 800, Projection: MapProjectionEquirectangular,
	})
	if err != nil {
		t.Fatalf("PlanetOccultationPathSVG: %v", err)
	}
	for _, want := range []string{
		`class="occultation-total-rise-set-boundary occultation-total-start-rise"`,
		`class="occultation-rise-set-boundary occultation-start-rise"`,
	} {
		if !strings.Contains(diagram, want) {
			t.Fatalf("planetary SVG is missing rise/set curve %s", want)
		}
	}
}

func reviewPoint(base time.Time, offset time.Duration, longitude, latitude float64) moon.OccultationPathPoint {
	return moon.OccultationPathPoint{
		Time:         base.Add(offset),
		Longitude:    longitude,
		Latitude:     latitude,
		MoonAltitude: 20,
	}
}

func reviewFootprint(when time.Time) moon.OccultationFootprint {
	return moon.OccultationFootprint{
		Time: when,
		Polygons: [][]moon.OccultationPathPoint{{
			{Time: when, Longitude: -20, Latitude: -10},
			{Time: when, Longitude: 5, Latitude: -10},
			{Time: when, Longitude: 5, Latitude: 10},
			{Time: when, Longitude: -20, Latitude: 10},
		}},
	}
}

func reviewRiseSetCurve(first, second time.Time) moon.OccultationRiseSetCurve {
	return moon.OccultationRiseSetCurve{
		Phase:     moon.RiseSetPhaseStart,
		Direction: moon.RiseSetDirectionRise,
		Segments: [][]moon.OccultationPathPoint{{
			{Time: first, Longitude: 100, Latitude: 8, MoonAltitude: 0},
			{Time: second, Longitude: 110, Latitude: 8, MoonAltitude: 0},
		}},
	}
}

func reviewHorizonConnectorPath() moon.StarOccultationPath {
	path := sampleStarOccultationPath()
	start := path.Start.Time
	curvePoint := func(offset time.Duration, longitude, latitude float64) moon.OccultationPathPoint {
		return moon.OccultationPathPoint{
			Time: start.Add(offset), Longitude: longitude, Latitude: latitude, MoonAltitude: 0,
		}
	}
	path.RiseSetCurves = []moon.OccultationRiseSetCurve{
		{
			Phase: moon.RiseSetPhaseStart, Direction: moon.RiseSetDirectionRise,
			Segments: [][]moon.OccultationPathPoint{{
				curvePoint(20*time.Minute, 169, 25), curvePoint(21*time.Minute, 170, 25),
			}},
		},
		{
			Phase: moon.RiseSetPhaseEnd, Direction: moon.RiseSetDirectionRise,
			Segments: [][]moon.OccultationPathPoint{{
				curvePoint(21*time.Minute+30*time.Second, 171, 25), curvePoint(22*time.Minute, 172, 25),
			}},
		},
	}
	footprint := func(offset time.Duration, longitude, latitude float64) moon.OccultationFootprint {
		when := start.Add(offset)
		return moon.OccultationFootprint{
			Time: when,
			Polygons: [][]moon.OccultationPathPoint{{
				{Time: when, Longitude: longitude - 0.5, Latitude: latitude - 0.5},
				{Time: when, Longitude: longitude + 0.5, Latitude: latitude - 0.5},
				{Time: when, Longitude: longitude, Latitude: latitude + 0.5},
			}},
			Boundaries: [][]moon.OccultationPathPoint{{
				{Time: when, Longitude: longitude, Latitude: latitude},
				{Time: when, Longitude: longitude + 0.05, Latitude: latitude + 0.05},
				{Time: when, Longitude: longitude + 0.1, Latitude: latitude + 0.1},
			}},
		}
	}
	path.Footprints = []moon.OccultationFootprint{
		footprint(21*time.Minute, 170, 25),
		footprint(21*time.Minute+30*time.Second, 171, 25),
	}
	path.BandFootprints = nil
	return path
}
