package geojson_test

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/internal/geodata"
)

// The static central band must describe the region where the eclipse is
// actually annular or total. For grazing events the shadow axis crosses Earth
// over only a fraction of the umbral contact interval, so a band derived from
// the paired limits alone silently drops the flared ends of the real path.
// These fixtures pin the coverage that NASA's path tables list from U1 to U4
// (for example 2003 May 31: limits from 004 35.9W to 060 19.3W).
const grazingBandCoverageToleranceKM = 100.0

type grazingBandCase struct {
	date    string
	options string
	limit   float64 // maximum tolerated footprint distance outside the band
}

var grazingBandCases = []grazingBandCase{
	// One-limit (|gamma| ~ 0.98-0.997) annulars: the reported defect.
	{"2003-05-31", "overview", 25},
	{"2003-05-31", "detail", 25},
	{"1874-10-10", "overview", 60},
	{"1874-10-10", "detail", 60},
	// One-limit total across the antimeridian.
	{"2185-07-26", "overview", 50},
	{"2185-07-26", "detail", 25},
	// Two-limit annulars whose analytic envelope is unavailable and whose
	// paired-limit ribbon used to be accepted without validation.
	{"1552-07-21", "overview", 25},
	{"-1480-12-27", "overview", 60},
	{"4862-09-28", "overview", 60},
	{"1042-06-20", "overview", 80},
	{"5705-06-17", "overview", 80},
	// Already-correct polar one-limit totality: must not regress.
	{"1522-03-27", "overview", 10},
	{"1522-03-27", "detail", 10},
	// Ordinary two-limit totality.
	{"2024-04-08", "overview", 25},
}

func TestSolarEclipseGrazingCentralBandCoversUmbralSweep(t *testing.T) {
	if testing.Short() {
		t.Skip("slow regression")
	}
	for _, testCase := range grazingBandCases {
		t.Run(testCase.date+"-"+testCase.options, func(t *testing.T) {
			date := grazingBandDate(t, testCase.date)
			info, ok := eclipse.SolarEclipseOnDate(date)
			if !ok || !info.HasCentral {
				t.Fatalf("expected a central solar eclipse on %s", testCase.date)
			}
			partialOptions, pathOptions := grazingBandOptions(info, testCase.options)
			partial, ok := eclipse.SolarEclipsePartialFootprints(date, partialOptions)
			if !ok {
				t.Fatal("missing partial footprints")
			}
			central, ok := eclipse.SolarEclipseCentralPath(date, pathOptions)
			if !ok {
				t.Fatal("missing central path")
			}
			raw, err := geojson.MarshalSolarEclipse(partial, &central)
			if err != nil {
				t.Fatal(err)
			}
			rings := grazingBandRings(t, raw)
			if len(rings) == 0 {
				t.Fatal("missing central-band feature")
			}
			paths := grazingFootprintPaths(partial.CentralBandFootprints)
			if len(paths) == 0 {
				t.Fatal("missing central band footprints")
			}
			miss := geodata.SphericalPolygonsPathMissDistanceKM(rings, paths, true)
			if miss > testCase.limit {
				t.Fatalf("central band leaves the umbral sweep %.1f km outside (limit %.1f km)",
					miss, testCase.limit)
			}
			centerPath := make([]geodata.GeoPoint, 0, len(central.CenterLine))
			for _, point := range central.CenterLine {
				centerPath = append(centerPath, geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude})
			}
			if centerMiss := geodata.SphericalPolygonsPathMissDistanceKM(
				rings, [][]geodata.GeoPoint{centerPath}, false,
			); centerMiss > 25 {
				t.Fatalf("central band leaves the center line %.1f km outside", centerMiss)
			}
		})
	}
}

// TestSolarEclipseGrazingCentralBandIsNotRejectedAsEnvelope guards the other
// direction: the flared-end band must still be a single closed continuous
// activation per mode, not a fan of open slices.
func TestSolarEclipseGrazingCentralBandIsSingleContinuousBand(t *testing.T) {
	for _, testCase := range grazingBandCases {
		if testCase.options != "overview" {
			continue
		}
		t.Run(testCase.date, func(t *testing.T) {
			date := grazingBandDate(t, testCase.date)
			info, ok := eclipse.SolarEclipseOnDate(date)
			if !ok {
				t.Fatalf("missing eclipse on %s", testCase.date)
			}
			partialOptions, pathOptions := grazingBandOptions(info, testCase.options)
			partial, ok := eclipse.SolarEclipsePartialFootprints(date, partialOptions)
			if !ok {
				t.Fatal("missing partial footprints")
			}
			central, ok := eclipse.SolarEclipseCentralPath(date, pathOptions)
			if !ok {
				t.Fatal("missing central path")
			}
			raw, err := geojson.MarshalSolarEclipse(partial, &central)
			if err != nil {
				t.Fatal(err)
			}
			var collection solarGeoJSONScanCollection
			if err := json.Unmarshal(raw, &collection); err != nil {
				t.Fatal(err)
			}
			bands := 0
			for _, feature := range collection.Features {
				if role, _ := feature.Properties["role"].(string); role == "central-band" {
					bands++
				}
			}
			if bands != 1 {
				t.Fatalf("central-band feature count = %d, want 1", bands)
			}
		})
	}
}

func grazingBandDate(t *testing.T, text string) time.Time {
	t.Helper()
	if strings.HasPrefix(text, "-") {
		parts := strings.Split(strings.TrimPrefix(text, "-"), "-")
		if len(parts) != 3 {
			t.Fatalf("invalid astronomical date %q", text)
		}
		year, err := strconv.Atoi(parts[0])
		if err != nil {
			t.Fatal(err)
		}
		month, err := strconv.Atoi(parts[1])
		if err != nil {
			t.Fatal(err)
		}
		day, err := strconv.Atoi(parts[2])
		if err != nil {
			t.Fatal(err)
		}
		return time.Date(-year, time.Month(month), day, 12, 0, 0, 0, time.UTC)
	}
	date, err := time.Parse("2006-01-02", text)
	if err != nil {
		t.Fatal(err)
	}
	return date
}

func grazingBandOptions(
	info eclipse.SolarEclipseInfo,
	mode string,
) (eclipse.SolarEclipsePartialFootprintOptions, eclipse.SolarEclipsePathOptions) {
	if mode == "detail" {
		return eclipse.SolarEclipsePartialFootprintOptions{
				Step:              2 * time.Minute,
				BoundaryPoints:    96,
				CentralShadowStep: 2 * time.Minute,
				RiseSetStep:       time.Minute,
				MagnitudeValues:   []float64{0.2, 0.4, 0.6, 0.8, 1.0},
			}, eclipse.SolarEclipsePathOptions{
				Step:            2 * time.Minute,
				TargetSpacingKM: 700,
			}
	}
	partial := eclipse.SolarEclipsePartialFootprintOptions{
		Step:           2 * time.Minute,
		BoundaryPoints: 96,
		RiseSetStep:    2 * time.Minute,
	}
	if info.Type == eclipse.SolarEclipseTotal {
		partial.MagnitudeValues = []float64{1}
	}
	return partial, eclipse.SolarEclipsePathOptions{
		Step:            2 * time.Minute,
		TargetSpacingKM: 150,
	}
}

func grazingBandRings(t *testing.T, raw []byte) [][]geodata.GeoPoint {
	t.Helper()
	var collection solarGeoJSONScanCollection
	if err := json.Unmarshal(raw, &collection); err != nil {
		t.Fatal(err)
	}
	var rings [][]geodata.GeoPoint
	for _, feature := range collection.Features {
		if role, _ := feature.Properties["role"].(string); role != "central-band" {
			continue
		}
		var polygons [][][][]float64
		if err := json.Unmarshal(feature.Geometry.Coordinates, &polygons); err != nil {
			t.Fatal(err)
		}
		for _, polygon := range polygons {
			if len(polygon) == 0 {
				continue
			}
			ring := make([]geodata.GeoPoint, 0, len(polygon[0]))
			for _, position := range polygon[0] {
				if len(position) < 2 {
					continue
				}
				ring = append(ring, geodata.GeoPoint{Longitude: position[0], Latitude: position[1]})
			}
			if len(ring) >= 4 {
				rings = append(rings, ring)
			}
		}
	}
	return rings
}

func grazingFootprintPaths(footprints []eclipse.SolarEclipsePartialFootprint) [][]geodata.GeoPoint {
	var paths [][]geodata.GeoPoint
	for _, footprint := range footprints {
		for _, boundary := range footprint.Boundaries {
			path := make([]geodata.GeoPoint, 0, len(boundary))
			for _, point := range boundary {
				path = append(path, geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude})
			}
			if len(path) >= 3 {
				paths = append(paths, path)
			}
		}
	}
	return paths
}

// TestSolarEclipseGrazingLimitsFollowBandBoundary pins the contract the map
// relies on: the dashed north/south limits and the filled central band must
// describe the same region. Ordinary events agree to a few kilometres because
// the band is built from those very limits; a grazing band is rebuilt from the
// umbral sweep, where the instantaneous cross-section limits stop describing
// the boundary at all (1136-06-01 sat 456 km inside its own band).
func TestSolarEclipseGrazingLimitsFollowBandBoundary(t *testing.T) {
	for _, testCase := range grazingBandCases {
		if testCase.options != "overview" {
			continue
		}
		t.Run(testCase.date, func(t *testing.T) {
			date := grazingBandDate(t, testCase.date)
			info, ok := eclipse.SolarEclipseOnDate(date)
			if !ok {
				t.Fatalf("missing eclipse on %s", testCase.date)
			}
			partialOptions, pathOptions := grazingBandOptions(info, testCase.options)
			partial, ok := eclipse.SolarEclipsePartialFootprints(date, partialOptions)
			if !ok {
				t.Fatal("missing partial footprints")
			}
			central, ok := eclipse.SolarEclipseCentralPath(date, pathOptions)
			if !ok {
				t.Fatal("missing central path")
			}
			raw, err := geojson.MarshalSolarEclipse(partial, &central)
			if err != nil {
				t.Fatal(err)
			}
			rings := grazingBandRings(t, raw)
			if len(rings) == 0 {
				t.Fatal("missing central-band feature")
			}
			if len(rings) == 0 {
				t.Fatal("missing central-band feature")
			}
			for _, role := range []string{"north-limit", "south-limit"} {
				path, ok := grazingLimitPath(t, raw, role)
				if !ok {
					t.Fatalf("missing %s feature", role)
				}
				miss := geodata.SphericalPolygonsPathMissDistanceKM(
					rings, [][]geodata.GeoPoint{path}, false,
				)
				if miss > 1.0 {
					t.Fatalf("%s sits %.1f km from the band boundary", role, miss)
				}
			}
		})
	}
}

// TestSolarEclipseSampledBandEdgeFollowsGreatestHorizonCurve pins the export
// contract the map shows: for a band rebuilt from sampled footprints, the edge
// that is bounded by the greatest-at-horizon condition must lie on that curve,
// otherwise the filled band and the drawn visibility line weave across each
// other at high zoom.
func TestSolarEclipseSampledBandEdgeFollowsGreatestHorizonCurve(t *testing.T) {
	exercised, skipped := 0, 0
	for _, testCase := range grazingBandCases {
		if testCase.options != "overview" || testCase.limit > 60 {
			continue
		}
		t.Run(testCase.date, func(t *testing.T) {
			date := grazingBandDate(t, testCase.date)
			info, ok := eclipse.SolarEclipseOnDate(date)
			if !ok {
				t.Fatalf("missing eclipse on %s", testCase.date)
			}
			partialOptions, pathOptions := grazingBandOptions(info, testCase.options)
			partial, ok := eclipse.SolarEclipsePartialFootprints(date, partialOptions)
			if !ok {
				t.Fatal("missing partial footprints")
			}
			if !partial.CentralBandSampled {
				skipped++
				t.Skip("analytic envelope: the band is already the exact boundary")
			}
			exercised++
			central, ok := eclipse.SolarEclipseCentralPath(date, pathOptions)
			if !ok {
				t.Fatal("missing central path")
			}
			raw, err := geojson.MarshalSolarEclipse(partial, &central)
			if err != nil {
				t.Fatal(err)
			}
			rings := grazingBandRings(t, raw)
			if len(rings) == 0 {
				t.Fatal("missing central-band feature")
			}
			curves := grazingGreatestCurves(t, raw)
			if len(curves) == 0 {
				t.Fatal("missing greatest visibility curves")
			}
			closest := math.Inf(1)
			for _, point := range rings[0] {
				for _, curve := range curves {
					for index := 0; index+1 < len(curve); index++ {
						closest = math.Min(closest, grazingPointSegmentKM(point, curve[index], curve[index+1]))
					}
				}
			}
			if closest > 1.0 {
				t.Fatalf("band edge stays %.1f km away from the greatest-at-horizon curve", closest)
			}
		})
	}
	// 采样带是少数情形：15 个夹具里只剩 3 个走这条断言。若夹具筛选或"权威带"来源变化，
	// 这些用例会退化成一堆 skip 而不是失败，所以钉住覆盖数下限。4862-09-28 原来在这 4 个里，
	// 它的第四个闭包根过去被三种子启发式漏掉；扫描式枚举解出后该事件成为解析带（带端太阳
	// 高度 +0.004°，与 1136-06-01 的 +0.005° 同量级，即带端确实由地平线封口）。
	if exercised < 3 {
		t.Fatalf("sampled-band assertion exercised by %d fixtures (%d analytic skips); the fixture filter or the band source narrowed silently", exercised, skipped)
	}
	t.Logf("sampled-band edge assertion exercised by %d fixtures, %d analytic skips", exercised, skipped)
}

// grazingGreatestCurves returns the exported greatest-at-horizon boundaries.
func grazingGreatestCurves(t *testing.T, raw []byte) [][]geodata.GeoPoint {
	t.Helper()
	var collection solarGeoJSONScanCollection
	if err := json.Unmarshal(raw, &collection); err != nil {
		t.Fatal(err)
	}
	var curves [][]geodata.GeoPoint
	for _, feature := range collection.Features {
		if role, _ := feature.Properties["role"].(string); role != "visibility-boundary" {
			continue
		}
		if phase, _ := feature.Properties["phase"].(string); phase != "greatest" {
			continue
		}
		var lines [][][]float64
		encoded, err := json.Marshal(feature.Geometry.Coordinates)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(encoded, &lines); err != nil {
			// A single LineString is exported as one coordinate array.
			var line [][]float64
			if lineErr := json.Unmarshal(encoded, &line); lineErr != nil {
				t.Fatal(err)
			}
			lines = [][][]float64{line}
		}
		for _, line := range lines {
			curve := make([]geodata.GeoPoint, 0, len(line))
			for _, position := range line {
				if len(position) < 2 {
					continue
				}
				curve = append(curve, geodata.GeoPoint{Longitude: position[0], Latitude: position[1]})
			}
			if len(curve) >= 2 {
				curves = append(curves, curve)
			}
		}
	}
	return curves
}

// grazingPointSegmentKM is the planar distance from a point to one segment.
func grazingPointSegmentKM(point, first, second geodata.GeoPoint) float64 {
	scale := math.Cos(point.Latitude * math.Pi / 180)
	ax := (first.Longitude - point.Longitude) * scale
	ay := first.Latitude - point.Latitude
	bx := (second.Longitude - point.Longitude) * scale
	by := second.Latitude - point.Latitude
	dx, dy := bx-ax, by-ay
	length := dx*dx + dy*dy
	fraction := 0.0
	if length > 0 {
		fraction = math.Max(0, math.Min(1, -(ax*dx+ay*dy)/length))
	}
	return 111.32 * math.Hypot(ax+fraction*dx, ay+fraction*dy)
}

// grazingLimitPath returns one exported limit line as a geographic path.
func grazingLimitPath(t *testing.T, raw []byte, role string) ([]geodata.GeoPoint, bool) {
	t.Helper()
	var collection solarGeoJSONScanCollection
	if err := json.Unmarshal(raw, &collection); err != nil {
		t.Fatal(err)
	}
	for _, feature := range collection.Features {
		if value, _ := feature.Properties["role"].(string); value != role {
			continue
		}
		var line [][]float64
		if err := json.Unmarshal(feature.Geometry.Coordinates, &line); err != nil {
			var lines [][][]float64
			if multiErr := json.Unmarshal(feature.Geometry.Coordinates, &lines); multiErr != nil {
				t.Fatal(err)
			}
			if len(lines) == 0 {
				return nil, false
			}
			line = lines[0]
		}
		path := make([]geodata.GeoPoint, 0, len(line))
		for _, position := range line {
			if len(position) < 2 {
				continue
			}
			path = append(path, geodata.GeoPoint{Longitude: position[0], Latitude: position[1]})
		}
		if len(path) >= 2 {
			return path, true
		}
	}
	return nil, false
}

// TestMarshalSolarEclipse11360601KeepsItsCentralBandSimple covers a shallow
// two-limit event whose northern limit runs through a cusp near the apex of a
// high-latitude path. Concatenating the two limits into one ribbon ring used to
// fold the ring onto itself, so the export contained a spike triangle plus
// disconnected end pieces instead of the swept band.
func TestMarshalSolarEclipse11360601KeepsItsCentralBandSimple(t *testing.T) {
	date := time.Date(1136, time.June, 1, 12, 0, 0, 0, time.UTC)
	info, ok := eclipse.SolarEclipseOnDate(date)
	if !ok || !info.HasCentral {
		t.Fatal("expected a central solar eclipse on 1136-06-01")
	}
	partialOptions, pathOptions := grazingBandOptions(info, "overview")
	partial, ok := eclipse.SolarEclipsePartialFootprints(date, partialOptions)
	if !ok {
		t.Fatal("missing partial footprints")
	}
	central, ok := eclipse.SolarEclipseCentralPath(date, pathOptions)
	if !ok {
		t.Fatal("missing central path")
	}
	raw, err := geojson.MarshalSolarEclipse(partial, &central)
	if err != nil {
		t.Fatal(err)
	}
	rings := grazingBandRings(t, raw)
	if len(rings) != 1 {
		t.Fatalf("central band exported as %d polygons, want 1 simple ring", len(rings))
	}
	if len(rings[0]) < 8 {
		t.Fatalf("central band ring has %d vertices", len(rings[0]))
	}
	if i, j, crossed := grazingRingCrossing(rings[0]); crossed {
		t.Fatalf("central band ring crosses itself between vertices %d and %d", i, j)
	}
	centerPath := make([]geodata.GeoPoint, 0, len(central.CenterLine))
	for _, point := range central.CenterLine {
		centerPath = append(centerPath, geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude})
	}
	if miss := geodata.SphericalPolygonsPathMissDistanceKM(
		rings, [][]geodata.GeoPoint{centerPath}, false,
	); miss > 25 {
		t.Fatalf("central band leaves the center line %.1f km outside", miss)
	}
}

// grazingRingCrossing reports the first planar self-intersection of a ring.
func grazingRingCrossing(ring []geodata.GeoPoint) (int, int, bool) {
	for first := 0; first+1 < len(ring); first++ {
		for second := first + 2; second+1 < len(ring); second++ {
			if first == 0 && second+1 == len(ring)-1 {
				continue
			}
			if grazingSegmentsCross(ring[first], ring[first+1], ring[second], ring[second+1]) {
				return first, second, true
			}
		}
	}
	return 0, 0, false
}

func grazingSegmentsCross(a, b, c, d geodata.GeoPoint) bool {
	side := func(p, q, r geodata.GeoPoint) float64 {
		return (q.Longitude-p.Longitude)*(r.Latitude-p.Latitude) - (q.Latitude-p.Latitude)*(r.Longitude-p.Longitude)
	}
	first := side(c, d, a)
	second := side(c, d, b)
	third := side(a, b, c)
	fourth := side(a, b, d)
	return (first > 0) != (second > 0) && (third > 0) != (fourth > 0)
}
