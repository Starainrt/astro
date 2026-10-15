package geojson_test

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro"
	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/internal/geodata"
	"github.com/starainrt/astro/moon"
)

type decodedCollection struct {
	Type     string           `json:"type"`
	Features []decodedFeature `json:"features"`
}

type decodedFeature struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Geometry   struct {
		Type        string          `json:"type"`
		Coordinates json.RawMessage `json:"coordinates"`
		Geometries  json.RawMessage `json:"geometries"`
	} `json:"geometry"`
}

func solarGeoJSONMapClosureEdge(first, second []float64) bool {
	return len(first) >= 2 && len(second) >= 2 &&
		(math.Abs(first[1]) >= 89.999999 || math.Abs(second[1]) >= 89.999999 ||
			math.Abs(first[0]) == 180 && first[0] == second[0])
}

// solarGeoJSONScanCollection 是扫描类测试读取导出集合的最小结构：只需要属性与几何坐标。
type solarGeoJSONScanCollection struct {
	Features []struct {
		Properties map[string]interface{} `json:"properties"`
		Geometry   struct {
			Coordinates json.RawMessage `json:"coordinates"`
		} `json:"geometry"`
	} `json:"features"`
}

func TestMarshalSolarEclipseFeatureCollection(t *testing.T) {
	date := time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
		Step:           20 * time.Minute,
		BoundaryPoints: 36,
	})
	if !ok {
		t.Fatal("expected solar partial footprints")
	}
	central, ok := eclipse.SolarEclipseCentralPath(date, eclipse.SolarEclipsePathOptions{Step: 5 * time.Minute})
	if !ok {
		t.Fatal("expected solar central path")
	}

	data, err := geojson.MarshalSolarEclipse(partial, &central)
	if err != nil {
		t.Fatalf("MarshalSolarEclipse: %v", err)
	}
	collection := decodeCollection(t, data)
	assertRoles(t, collection,
		"partial-footprint", "partial-band", "central-band", "center-line", "north-limit", "south-limit", "greatest")
	assertCollectionCoordinates(t, collection)
	partialBand := featureWithRole(t, collection, "partial-band")
	if partialBand.Geometry.Type != "MultiPolygon" {
		t.Fatalf("partial-band geometry=%q, want MultiPolygon", partialBand.Geometry.Type)
	}
	if partialBand.Properties["source"] != "zero-magnitude-envelope+horizon-boundary" {
		t.Fatalf("partial-band source=%v", partialBand.Properties["source"])
	}
	assertClosedMultiPolygon(t, featureWithRole(t, collection, "central-band"))
	assertTimedLineAligned(t, featureWithRole(t, collection, "center-line"))
}

func TestMarshalSolarEclipse20090722BuildsContinuousPartialBand(t *testing.T) {
	date := time.Date(2009, time.July, 22, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
		Step: 2 * time.Minute, BoundaryPoints: 96,
	})
	if !ok {
		t.Fatal("expected 2009-07-22 solar eclipse")
	}
	data, err := geojson.MarshalSolarEclipse(partial, nil)
	if err != nil {
		t.Fatalf("MarshalSolarEclipse: %v", err)
	}
	collection := decodeCollection(t, data)
	overlays := featuresWithRole(collection, "partial-band")
	if len(overlays) != 1 {
		t.Fatalf("partial-band count=%d, want one authoritative visibility region", len(overlays))
	}
	if overlays[0].Properties["source"] != "zero-magnitude-envelope+horizon-boundary" {
		t.Fatalf("partial-band source=%v", overlays[0].Properties["source"])
	}
	assertClosedMultiPolygon(t, overlays[0])
	assertCollectionCoordinates(t, collection)
	var polygons [][][][]float64
	if err := json.Unmarshal(overlays[0].Geometry.Coordinates, &polygons); err != nil {
		t.Fatalf("decode partial-band: %v", err)
	}
	if len(polygons) < 2 {
		t.Fatalf("partial-band fragments=%d, want antimeridian-safe split geometry", len(polygons))
	}
	for footprintIndex, footprint := range partial.Footprints {
		for boundaryIndex, boundary := range footprint.Boundaries {
			if len(boundary) == 0 {
				continue
			}
			step := (len(boundary) + 7) / 8
			for pointIndex := 0; pointIndex < len(boundary); pointIndex += step {
				point := []float64{boundary[pointIndex].Longitude, boundary[pointIndex].Latitude}
				if !geoJSONMultiPolygonContains(polygons, point[0], point[1]) &&
					geoJSONMultiPolygonBoundaryDistanceKM(polygons, point) > 5 {
					t.Fatalf("partial footprint %d boundary %d point %d protrudes outside the continuous band",
						footprintIndex, boundaryIndex, pointIndex)
				}
			}
		}
	}
	authoritativeLines := solarPartialBandAuthoritativeLines(partial)
	for polygonIndex, polygon := range polygons {
		for ringIndex, ring := range polygon {
			for pointIndex, point := range ring {
				if distance := geoJSONPointToLinesDistanceKM(point, authoritativeLines); distance > 5 {
					t.Fatalf("partial-band polygon %d ring %d point %d is %.1f km from the zero-magnitude or horizon boundary",
						polygonIndex, ringIndex, pointIndex, distance)
				}
			}
		}
	}
}

func TestMarshalSolarEclipsePartialBandFallsBackWithoutRiseSetTopology(t *testing.T) {
	date := time.Date(2009, time.July, 22, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
		Step: 2 * time.Minute, BoundaryPoints: 48, DisableRiseSet: true,
	})
	if !ok {
		t.Fatal("expected 2009-07-22 solar eclipse")
	}
	if len(partial.PartialBandContours) != 0 || len(partial.RiseSetCurves) != 0 {
		t.Fatal("disabled rise/set topology unexpectedly produced authoritative boundary lines")
	}
	data, err := geojson.MarshalSolarEclipse(partial, nil)
	if err != nil {
		t.Fatalf("MarshalSolarEclipse: %v", err)
	}
	band := featureWithRole(t, decodeCollection(t, data), "partial-band")
	if source := band.Properties["source"]; source != "open-boundary-endpoint-outlines" {
		t.Fatalf("fallback partial-band source=%v", source)
	}
	assertClosedMultiPolygon(t, band)
}

func solarPartialBandAuthoritativeLines(partial eclipse.SolarEclipsePartialFootprintsInfo) [][][]float64 {
	lines := make([][][]float64, 0, len(partial.PartialBandContours)+12)
	appendSegment := func(segment []eclipse.SolarEclipsePathPoint) {
		line := make([][]float64, len(segment))
		for index, point := range segment {
			line[index] = []float64{point.Longitude, point.Latitude}
		}
		lines = append(lines, line)
	}
	for _, contour := range partial.PartialBandContours {
		appendSegment(contour)
	}
	for _, curve := range partial.RiseSetCurves {
		for _, segment := range curve.Segments {
			appendSegment(segment)
		}
	}
	return lines
}

func geoJSONPointToLinesDistanceKM(point []float64, lines [][][]float64) float64 {
	minimum := math.Inf(1)
	for _, line := range lines {
		for index := 1; index < len(line); index++ {
			minimum = math.Min(minimum, geoJSONPointSegmentDistanceKM(point, line[index-1], line[index]))
		}
	}
	return minimum
}

func TestMarshalSolarEclipse20350902RiseSetCurveOrder(t *testing.T) {
	date := time.Date(2035, time.September, 2, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
		Step: 2 * time.Minute, BoundaryPoints: 96, CentralShadowStep: 2 * time.Minute,
		MagnitudeValues: []float64{0.2, 0.4, 0.6, 0.8, 1.0},
	})
	if !ok {
		t.Fatal("expected solar partial footprints")
	}
	for curveIndex, curve := range partial.RiseSetCurves {
		for segmentIndex, segment := range curve.Segments {
			for pointIndex := 1; pointIndex < len(segment); pointIndex++ {
				if !segment[pointIndex].Time.After(segment[pointIndex-1].Time) {
					t.Fatalf("curve=%d phase=%s direction=%s segment=%d point=%d previous=%s current=%s", curveIndex, curve.Phase, curve.Direction, segmentIndex, pointIndex, segment[pointIndex-1].Time.Format(time.RFC3339Nano), segment[pointIndex].Time.Format(time.RFC3339Nano))
				}
			}
		}
	}
	central, ok := eclipse.SolarEclipseCentralPath(date, eclipse.SolarEclipsePathOptions{Step: 2 * time.Minute})
	if !ok {
		t.Fatal("expected solar central path")
	}
	if _, err := geojson.MarshalSolarEclipse(partial, &central); err != nil {
		t.Fatalf("MarshalSolarEclipse: %v", err)
	}
}

func TestMarshalSolarEclipse19851101CentralLimitsHaveStrictTimes(t *testing.T) {
	date := time.Date(1985, time.November, 1, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
		Step: 10 * time.Minute, BoundaryPoints: 36,
	})
	if !ok {
		t.Fatal("expected 1985-11-01 partial footprints")
	}
	central, ok := eclipse.SolarEclipseCentralPath(date, eclipse.SolarEclipsePathOptions{
		Step: 10 * time.Minute,
	})
	if !ok {
		t.Fatal("expected 1985-11-01 central path")
	}
	if _, err := geojson.MarshalSolarEclipse(partial, &central); err != nil {
		t.Fatalf("MarshalSolarEclipse: %v", err)
	}
}

func TestMarshalSolarEclipseExportsVisibilityAntumbralAndMagnitudeLines(t *testing.T) {
	date := time.Date(2014, time.April, 29, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
		Step:              10 * time.Minute,
		BoundaryPoints:    24,
		CentralShadowStep: 2 * time.Minute,
		MagnitudeValues:   []float64{0.4, 0.8, 1.0},
	})
	if !ok {
		t.Fatal("expected non-central annular partial footprints")
	}
	data, err := geojson.MarshalSolarEclipse(partial, nil)
	if err != nil {
		t.Fatalf("MarshalSolarEclipse: %v", err)
	}
	collection := decodeCollection(t, data)
	assertRoles(t, collection, "visibility-boundary", "central-shadow-footprint", "central-band", "magnitude-line", "greatest")
	if got := len(featuresWithRole(collection, "central-shadow-footprint")); got < 3 {
		t.Fatalf("central-shadow-footprint count=%d, want at least 3", got)
	}
	// 被地平线切断的中心影足迹仍然是区域；物理边界另出 central-shadow-boundary 供描边。
	// A horizon-cut central-shadow footprint is still a region; its physical boundary
	// is exported separately as central-shadow-boundary.
	openFootprints := 0
	for _, footprint := range featuresWithRole(collection, "central-shadow-footprint") {
		if footprint.Geometry.Type != "MultiPolygon" {
			t.Fatalf("central-shadow-footprint geometry=%q, want MultiPolygon", footprint.Geometry.Type)
		}
		if closed, ok := footprint.Properties["source_boundary_closed"].(bool); !ok || !closed {
			openFootprints++
		}
	}
	boundaries := featuresWithRole(collection, "central-shadow-boundary")
	if len(boundaries) != openFootprints {
		t.Fatalf("central-shadow-boundary count=%d, want %d horizon-cut footprints", len(boundaries), openFootprints)
	}
	for _, boundary := range boundaries {
		if boundary.Geometry.Type != "MultiLineString" {
			t.Fatalf("central-shadow-boundary geometry=%q, want MultiLineString", boundary.Geometry.Type)
		}
	}
	centralBand := featureWithRole(t, collection, "central-band")
	assertClosedMultiPolygon(t, centralBand)
	var centralBandPolygons [][][][]float64
	if err := json.Unmarshal(centralBand.Geometry.Coordinates, &centralBandPolygons); err != nil {
		t.Fatalf("decode central-band: %v", err)
	}
	if len(centralBandPolygons) != 1 {
		t.Fatalf("non-central central-band polygon count=%d, want one continuous sweep", len(centralBandPolygons))
	}
	if centralBand.Properties["source"] != "besselian-critical-envelope" {
		t.Fatalf("central-band source=%v, want besselian-critical-envelope", centralBand.Properties["source"])
	}
	lines := featuresWithRole(collection, "magnitude-line")
	if len(lines) != 2 {
		t.Fatalf("magnitude-line count=%d, want one MultiLineString per requested magnitude", len(lines))
	}
	seen := make(map[string]bool)
	for _, line := range lines {
		magnitude, ok := line.Properties["magnitude"].(float64)
		if !ok || magnitude <= 0 || magnitude > 1 {
			t.Fatalf("invalid magnitude property: %#v", line.Properties["magnitude"])
		}
		if line.Geometry.Type != "MultiLineString" {
			t.Fatalf("magnitude %.1f geometry=%q, want MultiLineString", magnitude, line.Geometry.Type)
		}
		assertTimedLineAligned(t, line)
		seen[fmt.Sprintf("%.1f", magnitude)] = true
	}
	for _, key := range []string{"0.4", "0.8"} {
		if !seen[key] {
			t.Fatalf("missing magnitude line %s", key)
		}
	}
	assertRiseSetBoundaryFeatures(t, collection, "sun")
}

func TestMarshalSolarEclipse20140429NonCentralBandHasNoCombTeeth(t *testing.T) {
	partial, ok := eclipse.SolarEclipsePartialFootprints(
		time.Date(2014, time.April, 29, 0, 0, 0, 0, time.UTC),
		eclipse.SolarEclipsePartialFootprintOptions{
			Step: 10 * time.Minute, BoundaryPoints: 96, CentralShadowStep: 2 * time.Minute,
		},
	)
	if !ok {
		t.Fatal("expected 2014 non-central annular eclipse")
	}
	data, err := geojson.MarshalSolarEclipse(partial, nil)
	if err != nil {
		t.Fatalf("MarshalSolarEclipse: %v", err)
	}
	collection := decodeCollection(t, data)
	partialBand := featureWithRole(t, collection, "partial-band")
	source, _ := partialBand.Properties["source"].(string)
	if source != "zero-magnitude-envelope+horizon-boundary" &&
		source != "open-boundary-endpoint-outlines" {
		t.Fatalf("2014 partial-band source=%q", source)
	}
	assertClosedMultiPolygon(t, partialBand)
	if source == "zero-magnitude-envelope+horizon-boundary" {
		var partialPolygons [][][][]float64
		if err := json.Unmarshal(partialBand.Geometry.Coordinates, &partialPolygons); err != nil {
			t.Fatalf("decode partial band: %v", err)
		}
		for footprintIndex, footprint := range partial.Footprints {
			for boundaryIndex, boundary := range footprint.Boundaries {
				for pointIndex, point := range boundary {
					coordinate := []float64{point.Longitude, point.Latitude}
					if !geoJSONMultiPolygonContains(partialPolygons, coordinate[0], coordinate[1]) &&
						geoJSONMultiPolygonBoundaryDistanceKM(partialPolygons, coordinate) > 5 {
						t.Fatalf("authoritative partial-band misses footprint %d boundary %d point %d",
							footprintIndex, boundaryIndex, pointIndex)
					}
				}
			}
		}
	}
	band := featureWithRole(t, collection, "central-band")
	if band.Properties["source"] != "besselian-critical-envelope" {
		t.Fatalf("central-band source=%v, want besselian-critical-envelope", band.Properties["source"])
	}
	var polygons [][][][]float64
	if err := json.Unmarshal(band.Geometry.Coordinates, &polygons); err != nil {
		t.Fatalf("decode central band: %v", err)
	}
	if len(polygons) != 1 || len(polygons[0]) != 1 {
		t.Fatalf("2014 central band polygons=%d rings=%d, want one exterior ring", len(polygons), len(polygons[0]))
	}
	ring := polygons[0][0]
	if len(ring) >= 300 {
		t.Fatalf("2014 central-band ring points=%d, sampled ribbon union likely retained comb teeth", len(ring))
	}
	perimeter := 0.0
	maximumEdge := 0.0
	minimumGreatestDistance := math.Inf(1)
	greatest := []float64{partial.Eclipse.GreatestLongitude, partial.Eclipse.GreatestLatitude}
	for index := 1; index < len(ring); index++ {
		edge := geoJSONCoordinateDistanceKM(ring[index-1], ring[index])
		perimeter += edge
		maximumEdge = math.Max(maximumEdge, edge)
		minimumGreatestDistance = math.Min(minimumGreatestDistance,
			geoJSONPointSegmentDistanceKM(greatest, ring[index-1], ring[index]))
	}
	if perimeter >= 3000 {
		t.Fatalf("2014 central band perimeter=%.1f km, likely contains comb-like retracing", perimeter)
	}
	if maximumEdge > 16 {
		t.Fatalf("2014 central-band maximum edge=%.3f km, want a spatially refined boundary", maximumEdge)
	}
	// NASA rounds the non-central greatest marker independently from the local
	// greatest-at-sunset boundary. They are close, but are not the same root.
	if minimumGreatestDistance > 2 {
		t.Fatalf("2014 greatest marker is %.3f km from the grazing band tip", minimumGreatestDistance)
	}

	var greatestSet decodedFeature
	foundGreatestSet := false
	for _, feature := range featuresWithRole(decodeCollection(t, data), "visibility-boundary") {
		if feature.Properties["phase"] == "greatest" && feature.Properties["horizon"] == "set" {
			greatestSet, foundGreatestSet = feature, true
			break
		}
	}
	if !foundGreatestSet {
		t.Fatal("missing greatest-at-sunset visibility boundary")
	}
	var lines [][][]float64
	if err := json.Unmarshal(greatestSet.Geometry.Coordinates, &lines); err != nil {
		t.Fatalf("decode greatest-at-sunset line: %v", err)
	}
	maximumSharedBoundaryDistance := 0.0
	sharedPoints := 0
	for _, line := range lines {
		for _, point := range line {
			minimumDistance := math.Inf(1)
			for index := 1; index < len(ring); index++ {
				minimumDistance = math.Min(minimumDistance,
					geoJSONPointSegmentDistanceKM(point, ring[index-1], ring[index]))
			}
			if minimumDistance <= 0.01 {
				sharedPoints++
				maximumSharedBoundaryDistance = math.Max(maximumSharedBoundaryDistance, minimumDistance)
			}
		}
	}
	if sharedPoints < 80 || maximumSharedBoundaryDistance > 0.01 {
		t.Fatalf("greatest-at-sunset line shares %d exact band-edge samples (max %.6f km), want a dense coincident edge", sharedPoints, maximumSharedBoundaryDistance)
	}
}

func TestMarshalSolarEclipse20430409NonCentralTotalBandUsesCriticalEnvelope(t *testing.T) {
	partial, ok := eclipse.SolarEclipsePartialFootprints(
		time.Date(2043, time.April, 9, 0, 0, 0, 0, time.UTC),
		eclipse.SolarEclipsePartialFootprintOptions{
			Step: 2 * time.Minute, BoundaryPoints: 96, CentralShadowStep: 2 * time.Minute,
			MagnitudeValues: []float64{0.2, 0.4, 0.6, 0.8, 1},
		},
	)
	if !ok || partial.Eclipse.Type != eclipse.SolarEclipseTotal ||
		partial.Eclipse.Centrality != eclipse.SolarEclipseNonCentral {
		t.Fatalf("2043 eclipse type=%s centrality=%s ok=%v, want non-central total",
			partial.Eclipse.Type, partial.Eclipse.Centrality, ok)
	}
	band := featureWithRole(t, decodeCollection(t, mustMarshalSolarEclipse(t, partial)), "central-band")
	if band.Properties["source"] != "besselian-critical-envelope" {
		t.Fatalf("2043 central-band source=%v, want besselian-critical-envelope", band.Properties["source"])
	}
	var polygons [][][][]float64
	if err := json.Unmarshal(band.Geometry.Coordinates, &polygons); err != nil {
		t.Fatalf("decode 2043 central band: %v", err)
	}
	if len(polygons) != 1 || len(polygons[0]) != 1 {
		t.Fatalf("2043 central band polygons=%d rings=%d, want one exterior ring", len(polygons), len(polygons[0]))
	}
	ring := polygons[0][0]
	if len(ring) < 100 || len(ring) > 500 {
		t.Fatalf("2043 central-band ring points=%d, want one compact smooth envelope", len(ring))
	}
	for index := 1; index < len(ring); index++ {
		if edge := geoJSONCoordinateDistanceKM(ring[index-1], ring[index]); edge > 28 {
			t.Fatalf("2043 central-band edge %d=%.3f km, want adaptive spatial sampling", index-1, edge)
		}
	}
	for first := 0; first+1 < len(ring); first++ {
		for second := first + 2; second+1 < len(ring); second++ {
			if first == 0 && second+1 == len(ring)-1 {
				continue
			}
			if geoJSONSegmentsCross(ring[first], ring[first+1], ring[second], ring[second+1]) {
				t.Fatalf("2043 central-band ring self-intersects between edges %d and %d", first, second)
			}
		}
	}
}

func TestMarshalSolarEclipse20430409PartialBandHasNoPolarSeam(t *testing.T) {
	partial, ok := eclipse.SolarEclipsePartialFootprints(
		time.Date(2043, time.April, 9, 0, 0, 0, 0, time.UTC),
		eclipse.SolarEclipsePartialFootprintOptions{
			Step: 2 * time.Minute, BoundaryPoints: 96, CentralShadowStep: 2 * time.Minute,
			MagnitudeValues: []float64{0.2, 0.4, 0.6, 0.8, 1},
		},
	)
	if !ok {
		t.Fatal("expected 2043-04-09 solar eclipse")
	}
	band := featureWithRole(t, decodeCollection(t, mustMarshalSolarEclipse(t, partial)), "partial-band")
	var polygons [][][][]float64
	if err := json.Unmarshal(band.Geometry.Coordinates, &polygons); err != nil {
		t.Fatalf("decode 2043 partial band: %v", err)
	}
	assertClosedMultiPolygon(t, band)
	// A band containing the pole can be one closed map fragment. Check both
	// sides of the antimeridian and independent station visibility instead of
	// requiring a particular number of fragments.
	for _, point := range [][2]float64{
		{149, 60}, {149.5, 60}, {179, 60}, {-179, 60}, {179, 80}, {-179, 80}, {0, 89},
		{0, 40}, {-100, 20}, {110, 40},
	} {
		_, visible := eclipse.LocalSolarEclipseOnDate(time.Date(2043, 4, 9, 0, 0, 0, 0, time.UTC), point[0], point[1], 0)
		if got := geoJSONMultiPolygonContains(polygons, point[0], point[1]); got != visible {
			t.Errorf("2043 partial-band at %v contains=%v, station visibility=%v", point, got, visible)
		}
	}
}

func TestMarshalSolarEclipseNonCentralGreatestHorizonFolds(t *testing.T) {
	for _, date := range []time.Time{
		time.Date(1656, time.July, 21, 0, 0, 0, 0, time.UTC),
		time.Date(1928, time.May, 19, 0, 0, 0, 0, time.UTC),
		time.Date(1957, time.October, 23, 0, 0, 0, 0, time.UTC),
		time.Date(1967, time.November, 2, 0, 0, 0, 0, time.UTC),
	} {
		partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
			Step: 2 * time.Minute, BoundaryPoints: 24, CentralShadowStep: 2 * time.Minute,
			RiseSetStep: 2 * time.Minute,
		})
		if !ok {
			t.Fatalf("%s: expected eclipse", date.Format("2006-01-02"))
		}
		if _, err := geojson.MarshalSolarEclipse(partial, nil); err != nil {
			t.Fatalf("%s: MarshalSolarEclipse: %v", date.Format("2006-01-02"), err)
		}
	}
}

func TestMarshalSolarEclipse19500318UsesValidatedOpenSweepFallback(t *testing.T) {
	partial, ok := eclipse.SolarEclipsePartialFootprints(
		time.Date(1950, time.March, 18, 0, 0, 0, 0, time.UTC),
		eclipse.SolarEclipsePartialFootprintOptions{Step: 5 * time.Minute, BoundaryPoints: 24},
	)
	if !ok {
		t.Fatal("expected 1950 non-central annular eclipse")
	}
	band := featureWithRole(t, decodeCollection(t, mustMarshalSolarEclipse(t, partial)), "central-band")
	if band.Properties["source"] != "central-shadow-sweep" {
		t.Fatalf("1950 central-band source=%v, want validated open sweep fallback", band.Properties["source"])
	}
	assertClosedMultiPolygon(t, band)
}

func mustMarshalSolarEclipse(t *testing.T, partial eclipse.SolarEclipsePartialFootprintsInfo) []byte {
	t.Helper()
	data, err := geojson.MarshalSolarEclipse(partial, nil)
	if err != nil {
		t.Fatalf("MarshalSolarEclipse: %v", err)
	}
	return data
}

func geoJSONSegmentsCross(a, b, c, d []float64) bool {
	orientation := func(first, second, third []float64) float64 {
		return (second[0]-first[0])*(third[1]-first[1]) -
			(second[1]-first[1])*(third[0]-first[0])
	}
	first, second := orientation(a, b, c), orientation(a, b, d)
	third, fourth := orientation(c, d, a), orientation(c, d, b)
	return ((first > 1e-10 && second < -1e-10) || (first < -1e-10 && second > 1e-10)) &&
		((third > 1e-10 && fourth < -1e-10) || (third < -1e-10 && fourth > 1e-10))
}

func TestMarshalSolarEclipseCoarseCentralPathGracefullyRecovers(t *testing.T) {
	for _, date := range []time.Time{
		time.Date(1891, time.June, 6, 0, 0, 0, 0, time.UTC),
		time.Date(2119, time.March, 11, 0, 0, 0, 0, time.UTC),
	} {
		partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
			Step: 60 * time.Minute, BoundaryPoints: 96, CentralShadowStep: 2 * time.Minute, DisableRiseSet: true,
		})
		if !ok {
			t.Fatalf("%s partial footprints unavailable", date.Format("2006-01-02"))
		}
		central, ok := eclipse.SolarEclipseCentralPath(date, eclipse.SolarEclipsePathOptions{Step: 60 * time.Minute})
		if !ok || len(central.CenterLine) < 2 {
			t.Fatalf("%s coarse central path points=%d, want at least two", date.Format("2006-01-02"), len(central.CenterLine))
		}
		if _, err := geojson.MarshalSolarEclipse(partial, &central); err != nil {
			t.Fatalf("%s MarshalSolarEclipse: %v", date.Format("2006-01-02"), err)
		}
	}
}

func TestMarshalSolarEclipse20230420HybridBandFollowsCenterLine(t *testing.T) {
	date := time.Date(2023, time.April, 20, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
		Step: 10 * time.Minute, BoundaryPoints: 96, CentralShadowStep: 2 * time.Minute,
		MagnitudeValues: []float64{1.0, 1.01},
	})
	if !ok || partial.Eclipse.Type != eclipse.SolarEclipseHybrid {
		t.Fatalf("expected hybrid eclipse, got ok=%v type=%s", ok, partial.Eclipse.Type)
	}
	central, ok := eclipse.SolarEclipseCentralPath(date, eclipse.SolarEclipsePathOptions{Step: 2 * time.Minute})
	if !ok {
		t.Fatal("expected hybrid central path")
	}
	data, err := geojson.MarshalSolarEclipse(partial, &central)
	if err != nil {
		t.Fatalf("MarshalSolarEclipse: %v", err)
	}
	collection := decodeCollection(t, data)
	band := featureWithRole(t, collection, "central-band")
	var polygons [][][][]float64
	if err := json.Unmarshal(band.Geometry.Coordinates, &polygons); err != nil {
		t.Fatalf("decode hybrid central band: %v", err)
	}
	if len(polygons) != 4 {
		t.Fatalf("hybrid central band polygon count=%d, want three physical lobes split at the antimeridian", len(polygons))
	}
	center := featureWithRole(t, collection, "center-line")
	var lines [][][]float64
	if err := json.Unmarshal(center.Geometry.Coordinates, &lines); err != nil {
		t.Fatalf("decode hybrid center line: %v", err)
	}
	for lineIndex, line := range lines {
		for pointIndex, point := range line {
			if !geometryContainsPoint(t, band.Geometry, point[0], point[1]) {
				t.Fatalf("center line[%d] point %d lies outside hybrid central band at %.6f, %.6f", lineIndex, pointIndex, point[0], point[1])
			}
		}
	}
	magnitudeLines := featuresWithRole(collection, "magnitude-line")
	if len(magnitudeLines) != 2 {
		t.Fatalf("hybrid magnitude lines=%d, want two (1.0 and 1.01)", len(magnitudeLines))
	}
	foundHighMagnitude := false
	var magnitudeOne decodedFeature
	for _, line := range magnitudeLines {
		if line.Properties["magnitude"] == 1.0 {
			magnitudeOne = line
		}
		if line.Properties["magnitude"] == 1.01 {
			foundHighMagnitude = true
		}
	}
	if !foundHighMagnitude {
		t.Fatal("hybrid GeoJSON is missing the 1.01 magnitude contour")
	}
	var magnitudeOneLines [][][]float64
	if err := json.Unmarshal(magnitudeOne.Geometry.Coordinates, &magnitudeOneLines); err != nil {
		t.Fatalf("decode hybrid 1.0 magnitude line: %v", err)
	}
	for lineIndex, line := range magnitudeOneLines {
		for _, pointIndex := range []int{0, len(line) - 1} {
			minimumDistance := math.Inf(1)
			for _, centerLine := range lines {
				for _, centerPoint := range centerLine {
					minimumDistance = math.Min(minimumDistance, geoJSONCoordinateDistanceKM(line[pointIndex], centerPoint))
				}
			}
			if minimumDistance > 0.1 {
				t.Fatalf("hybrid 1.0 line[%d] endpoint %d misses center line by %.3f km", lineIndex, pointIndex, minimumDistance)
			}
		}
	}
}

func TestMarshalSolarEclipseExportsTotalMagnitudeAboveOne(t *testing.T) {
	partial, ok := eclipse.SolarEclipsePartialFootprints(
		time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC),
		eclipse.SolarEclipsePartialFootprintOptions{
			Step: 10 * time.Minute, BoundaryPoints: 24, MagnitudeValues: []float64{1.01},
		},
	)
	if !ok || len(partial.MagnitudeContours) != 1 {
		t.Fatalf("expected one totality magnitude contour, got ok=%v contours=%d", ok, len(partial.MagnitudeContours))
	}
	data, err := geojson.MarshalSolarEclipse(partial, nil)
	if err != nil {
		t.Fatalf("MarshalSolarEclipse: %v", err)
	}
	collection := decodeCollection(t, data)
	lines := featuresWithRole(collection, "magnitude-line")
	if len(lines) != 1 || lines[0].Properties["magnitude"] != 1.01 {
		t.Fatalf("magnitude lines = %#v, want one line at 1.01", lines)
	}
}

func TestMarshalSolarEclipse20260812CentralBandFollowsTotalityEnvelope(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	date := time.Date(2026, time.August, 12, 0, 0, 0, 0, zone)
	partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
		Step: 2 * time.Minute, BoundaryPoints: 96, CentralShadowStep: 2 * time.Minute,
		MagnitudeValues: []float64{1.0},
	})
	if !ok || partial.Eclipse.Type != eclipse.SolarEclipseTotal {
		t.Fatalf("expected 2026-08-12 total eclipse, got ok=%v type=%s", ok, partial.Eclipse.Type)
	}
	central, ok := eclipse.SolarEclipseCentralPath(date, eclipse.SolarEclipsePathOptions{
		Step: 2 * time.Minute, TargetSpacingKM: 700,
	})
	if !ok {
		t.Fatal("expected central path")
	}
	data, err := geojson.MarshalSolarEclipse(partial, &central)
	if err != nil {
		t.Fatalf("MarshalSolarEclipse: %v", err)
	}
	band := featureWithRole(t, decodeCollection(t, data), "central-band")
	assertClosedMultiPolygon(t, band)
	if source := band.Properties["source"]; source != "magnitude-one-envelope" {
		t.Fatalf("2026 central-band source=%v, want magnitude-one-envelope", source)
	}
	var polygons [][][][]float64
	if err := json.Unmarshal(band.Geometry.Coordinates, &polygons); err != nil {
		t.Fatalf("decode central band: %v", err)
	}
	if len(polygons) == 0 || len(polygons) > 2 {
		t.Fatalf("central band has %d polygons, want one physical band with at most one antimeridian split", len(polygons))
	}
	for _, polygon := range polygons {
		if len(polygon) == 0 {
			t.Fatal("central-band polygon has no exterior ring")
		}
		assertSolarCentralBandRingSimpleAndSampled(t, polygon[0], 250)
	}
	for closureIndex, closure := range partial.CentralBandHorizonClosures {
		for pointIndex, point := range closure {
			coordinate := []float64{point.Longitude, point.Latitude}
			if distance := geoJSONMultiPolygonBoundaryDistanceKM(polygons, coordinate); distance > 0.1 {
				t.Fatalf("horizon closure %d point %d is %.3f km from the magnitude-one boundary", closureIndex, pointIndex, distance)
			}
		}
	}
	// These points are independently classified by the local Split-K solver
	// as total, but the old same-time cross-section band omitted them.
	for _, point := range [][2]float64{{-4, 43.25}, {-4, 43.75}, {-2, 42.25}} {
		if !geometryContainsPoint(t, band.Geometry, point[0], point[1]) {
			t.Fatalf("central band omits independently total point %.2f, %.2f", point[0], point[1])
		}
	}
}

func TestMarshalSolarEclipse20100115CentralBandContainsFuyang(t *testing.T) {
	date := time.Date(2010, time.January, 15, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
		Step:           10 * time.Minute,
		BoundaryPoints: 180,
	})
	if !ok {
		t.Fatal("expected solar partial footprints")
	}
	central, ok := eclipse.SolarEclipseCentralPath(date, eclipse.SolarEclipsePathOptions{
		Step: 10 * time.Minute, TargetSpacingKM: 100,
	})
	if !ok {
		t.Fatal("expected solar central path")
	}
	data, err := geojson.MarshalSolarEclipse(partial, &central)
	if err != nil {
		t.Fatalf("MarshalSolarEclipse: %v", err)
	}
	band := featureWithRole(t, decodeCollection(t, data), "central-band")
	if source := band.Properties["source"]; source != "besselian-critical-envelope" {
		t.Fatalf("2010 central-band source=%v, want continuous critical envelope", source)
	}
	if band.Geometry.Type != "MultiPolygon" {
		t.Fatalf("central-band geometry=%q, want one merged MultiPolygon without internal seams", band.Geometry.Type)
	}
	var polygons [][][][]float64
	if err := json.Unmarshal(band.Geometry.Coordinates, &polygons); err != nil {
		t.Fatalf("decode central-band: %v", err)
	}
	if len(polygons) != 1 {
		t.Fatalf("central-band polygon count=%d, want one continuous outline", len(polygons))
	}
	ring := polygons[0][0]
	for first := 0; first+1 < len(ring); first++ {
		for second := first + 2; second+1 < len(ring); second++ {
			if first == 0 && second+1 == len(ring)-1 {
				continue
			}
			if geoJSONSegmentsCross(ring[first], ring[first+1], ring[second], ring[second+1]) {
				t.Fatalf("2010 central-band ring self-intersects between edges %d and %d", first, second)
			}
		}
	}
	for index := 1; index < len(ring); index++ {
		if distance := geoJSONCoordinateDistanceKM(ring[index-1], ring[index]); distance > 250 {
			t.Fatalf("central-band edge %d is %.1f km, want a sampled curved outline", index, distance)
		}
	}
	minimumEndTurn := 180.0
	for index := 1; index+1 < len(ring); index++ {
		point := ring[index]
		if point[0] < 120 || point[0] > 123 || point[1] < 36 || point[1] > 39 {
			continue
		}
		incoming := math.Atan2(point[1]-ring[index-1][1], point[0]-ring[index-1][0])
		outgoing := math.Atan2(ring[index+1][1]-point[1], ring[index+1][0]-point[0])
		minimumEndTurn = math.Min(minimumEndTurn, math.Remainder((outgoing-incoming)*180/math.Pi, 360))
	}
	if minimumEndTurn < -30 {
		t.Fatalf("2010 eastern central-band cap turns inward by %.1f degrees", minimumEndTurn)
	}
	for name, point := range map[string]eclipse.SolarEclipsePathPoint{"U1": partial.U1, "U4": partial.U4} {
		if !geometryContainsPoint(t, band.Geometry, point.Longitude, point.Latitude) &&
			geoJSONMultiPolygonBoundaryDistanceKM(polygons, []float64{point.Longitude, point.Latitude}) > 1 {
			t.Fatalf("%s external shadow contact lies outside the swept central band", name)
		}
	}
	for _, role := range []string{"north-limit", "south-limit"} {
		var lines [][][]float64
		if err := json.Unmarshal(featureWithRole(t, decodeCollection(t, data), role).Geometry.Coordinates, &lines); err != nil {
			t.Fatalf("decode %s: %v", role, err)
		}
		for _, line := range lines {
			for _, coordinate := range line {
				for name, point := range map[string]eclipse.SolarEclipsePathPoint{"U1": partial.U1, "U4": partial.U4} {
					if math.Abs(coordinate[0]-point.Longitude) <= 1e-10 && math.Abs(coordinate[1]-point.Latitude) <= 1e-10 {
						t.Fatalf("%s retains %s external-contact vertex", role, name)
					}
				}
			}
		}
	}
	for first := 0; first+1 < len(ring); first++ {
		for second := first + 2; second+1 < len(ring); second++ {
			if geoJSONCoordinateDistanceKM(ring[first], ring[second]) < 0.001 {
				t.Fatalf("central-band ring repeats non-adjacent vertices %d and %d", first, second)
			}
		}
	}
	if !geometryContainsPoint(t, band.Geometry, 115.4, 32.9) {
		t.Fatal("Fuyang (115.4E, 32.9N) is outside the 2010-01-15 annular central band")
	}
	center := featureWithRole(t, decodeCollection(t, data), "center-line")
	var centerLines [][][]float64
	if err := json.Unmarshal(center.Geometry.Coordinates, &centerLines); err != nil {
		t.Fatalf("decode center-line: %v", err)
	}
	for segmentIndex, segment := range centerLines {
		for pointIndex, point := range segment {
			if !geometryContainsPoint(t, band.Geometry, point[0], point[1]) &&
				geoJSONMultiPolygonBoundaryDistanceKM(polygons, point) > 10 {
				t.Fatalf("center-line segment %d point %d lies outside the central band at %.6f, %.6f",
					segmentIndex, pointIndex, point[0], point[1])
			}
		}
	}
}

func TestMarshalSolarEclipseHistoricalCrossedLimitsUseSimpleRibbonUnion(t *testing.T) {
	for _, date := range []time.Time{
		time.Date(1547, time.November, 12, 0, 0, 0, 0, time.UTC),
		time.Date(1565, time.November, 22, 0, 0, 0, 0, time.UTC),
	} {
		date := date
		t.Run(date.Format("2006-01-02"), func(t *testing.T) {
			partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
				Step: 2 * time.Minute, BoundaryPoints: 96, CentralShadowStep: 2 * time.Minute,
			})
			if !ok {
				t.Fatal("expected historical annular eclipse")
			}
			central, ok := eclipse.SolarEclipseCentralPath(date, eclipse.SolarEclipsePathOptions{
				Step: 2 * time.Minute, TargetSpacingKM: 700,
			})
			if !ok {
				t.Fatal("expected historical central path")
			}
			band := featureWithRole(t, decodeCollection(t, mustMarshalSolarEclipseWithPath(t, partial, &central)), "central-band")
			if source := band.Properties["source"]; source != "paired-limits+central-shadow-ribbon-union" &&
				source != "paired-limits-ribbon-union" && source != "besselian-critical-envelope" {
				t.Fatalf("central-band source=%v, want a validated central-band geometry", source)
			}
			var polygons [][][][]float64
			if err := json.Unmarshal(band.Geometry.Coordinates, &polygons); err != nil {
				t.Fatalf("decode central-band: %v", err)
			}
			if len(polygons) != 1 || len(polygons[0]) != 1 {
				t.Fatalf("central-band polygons=%d rings=%d, want one exterior ring", len(polygons), len(polygons[0]))
			}
			ring := polygons[0][0]
			for first := 0; first+1 < len(ring); first++ {
				for second := first + 2; second+1 < len(ring); second++ {
					if first == 0 && second+1 == len(ring)-1 {
						continue
					}
					if geoJSONSegmentsCross(ring[first], ring[first+1], ring[second], ring[second+1]) {
						t.Fatalf("central-band ring self-intersects between edges %d and %d", first, second)
					}
				}
			}
			for index := 1; index < len(ring); index++ {
				if edge := geoJSONCoordinateDistanceKM(ring[index-1], ring[index]); edge > 210 {
					t.Fatalf("central-band edge %d=%.3f km, want spatial refinement", index-1, edge)
				}
			}
		})
	}
}

func mustMarshalSolarEclipseWithPath(
	t *testing.T,
	partial eclipse.SolarEclipsePartialFootprintsInfo,
	central *eclipse.SolarEclipsePath,
) []byte {
	t.Helper()
	data, err := geojson.MarshalSolarEclipse(partial, central)
	if err != nil {
		t.Fatalf("MarshalSolarEclipse: %v", err)
	}
	return data
}

func TestMarshalSolarEclipse20100115CenterLineEndsAtGreatestSetBoundary(t *testing.T) {
	date := time.Date(2010, time.January, 15, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
		Step: 10 * time.Minute, BoundaryPoints: 24, RiseSetStep: 2 * time.Minute,
	})
	if !ok {
		t.Fatal("expected solar partial footprints")
	}
	central, ok := eclipse.SolarEclipseCentralPath(date, eclipse.SolarEclipsePathOptions{
		Step: 2 * time.Minute, TargetSpacingKM: 100,
	})
	if !ok {
		t.Fatal("expected solar central path")
	}
	data, err := geojson.MarshalSolarEclipse(partial, &central)
	if err != nil {
		t.Fatalf("MarshalSolarEclipse: %v", err)
	}
	collection := decodeCollection(t, data)

	center := featureWithRole(t, collection, "center-line")
	var centerLines [][][]float64
	if err := json.Unmarshal(center.Geometry.Coordinates, &centerLines); err != nil {
		t.Fatalf("decode center-line: %v", err)
	}
	if len(centerLines) == 0 || len(centerLines[len(centerLines)-1]) == 0 {
		t.Fatal("center-line has no coordinates")
	}
	lastLine := centerLines[len(centerLines)-1]
	last := lastLine[len(lastLine)-1]
	// NASA's path-table Limits row is 36 deg 49.6 min N, 121 deg 40.9 min E.
	if math.Abs(last[0]-121.6817) > 0.12 || math.Abs(last[1]-36.8267) > 0.12 {
		t.Fatalf("center-line limit = (%.6f, %.6f), want NASA limit near (121.6817, 36.8267)",
			last[0], last[1])
	}

	foundGreatestSet := false
	for _, boundary := range featuresWithRole(collection, "visibility-boundary") {
		if boundary.Properties["phase"] != "greatest" || boundary.Properties["horizon"] != "set" {
			continue
		}
		var lines [][][]float64
		if err := json.Unmarshal(boundary.Geometry.Coordinates, &lines); err != nil {
			t.Fatalf("decode greatest-at-sunset boundary: %v", err)
		}
		for _, line := range lines {
			for _, point := range line {
				if point[0] == last[0] && point[1] == last[1] {
					foundGreatestSet = true
				}
			}
		}
	}
	if !foundGreatestSet {
		t.Fatal("center-line limit is not a shared GeoJSON vertex of the greatest-at-sunset boundary")
	}
}

func TestMarshalSolarEclipse20080801CentralBandContainsCenterLine(t *testing.T) {
	date := time.Date(2008, time.August, 1, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
		Step: 30 * time.Minute, BoundaryPoints: 180,
		CentralShadowStep: 2 * time.Minute, DisableRiseSet: true,
	})
	if !ok {
		t.Fatal("expected solar partial footprints")
	}
	central, ok := eclipse.SolarEclipseCentralPath(date, eclipse.SolarEclipsePathOptions{
		Step: 2 * time.Minute,
	})
	if !ok {
		t.Fatal("expected solar central path")
	}
	data, err := geojson.MarshalSolarEclipse(partial, &central)
	if err != nil {
		t.Fatalf("MarshalSolarEclipse: %v", err)
	}
	band := featureWithRole(t, decodeCollection(t, data), "central-band")
	center := featureWithRole(t, decodeCollection(t, data), "center-line")
	var centerLines [][][]float64
	if err := json.Unmarshal(center.Geometry.Coordinates, &centerLines); err != nil {
		t.Fatalf("decode center-line: %v", err)
	}
	for segmentIndex, segment := range centerLines {
		for pointIndex, point := range segment {
			if !geometryContainsPoint(t, band.Geometry, point[0], point[1]) {
				t.Fatalf("center-line segment %d point %d lies outside 2008 central band at %.6f, %.6f", segmentIndex, pointIndex, point[0], point[1])
			}
		}
	}
}

func TestMarshalSolarEclipseCentralShadowStepTwoMinutesFallsBackToStableBand(t *testing.T) {
	for _, date := range []time.Time{
		time.Date(2037, time.July, 13, 0, 0, 0, 0, time.UTC),
		time.Date(2038, time.July, 2, 0, 0, 0, 0, time.UTC),
	} {
		t.Run(date.Format("2006-01-02"), func(t *testing.T) {
			partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
				Step: 10 * time.Minute, BoundaryPoints: 180, CentralShadowStep: 2 * time.Minute,
			})
			if !ok {
				t.Fatal("expected solar partial footprints")
			}
			central, ok := eclipse.SolarEclipseCentralPath(date, eclipse.SolarEclipsePathOptions{Step: 2 * time.Minute})
			if !ok {
				t.Fatal("expected solar central path")
			}
			if _, err := geojson.MarshalSolarEclipse(partial, &central); err != nil {
				t.Fatalf("MarshalSolarEclipse: %v", err)
			}
		})
	}
}

func TestMarshalCentralEclipseBandUnionAcrossEvents(t *testing.T) {
	for _, date := range []time.Time{
		time.Date(2009, time.July, 22, 0, 0, 0, 0, time.UTC),
		time.Date(2012, time.May, 20, 0, 0, 0, 0, time.UTC),
		time.Date(2017, time.August, 21, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC),
	} {
		date := date
		t.Run(date.Format("2006-01-02"), func(t *testing.T) {
			partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
				Step:              10 * time.Minute,
				BoundaryPoints:    48,
				CentralShadowStep: 5 * time.Minute,
				DisableRiseSet:    true,
			})
			if !ok {
				t.Fatal("expected solar partial footprints")
			}
			central, ok := eclipse.SolarEclipseCentralPath(date, eclipse.SolarEclipsePathOptions{
				Step: 10 * time.Minute, TargetSpacingKM: 500,
			})
			if !ok {
				t.Fatal("expected solar central path")
			}
			data, err := geojson.MarshalSolarEclipse(partial, &central)
			if err != nil {
				t.Fatalf("MarshalSolarEclipse: %v", err)
			}
			band := featureWithRole(t, decodeCollection(t, data), "central-band")
			if band.Geometry.Type != "MultiPolygon" {
				t.Fatalf("central-band geometry=%q, want merged MultiPolygon", band.Geometry.Type)
			}
			assertClosedMultiPolygon(t, band)
		})
	}
}

func TestMarshalSolarEclipseAllowsFoldedMagnitudeContourTimes(t *testing.T) {
	date := time.Date(2031, time.May, 21, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
		Step: 20 * time.Minute, BoundaryPoints: 24, MagnitudeValues: []float64{0.8},
	})
	if !ok || len(partial.MagnitudeContours) != 1 {
		t.Fatalf("magnitude contours=%d ok=%v, want one", len(partial.MagnitudeContours), ok)
	}
	folded := false
	for _, segment := range partial.MagnitudeContours[0].Segments {
		for index := 1; index < len(segment); index++ {
			if !segment[index].Time.After(segment[index-1].Time) {
				folded = true
				break
			}
		}
	}
	if !folded {
		t.Fatal("2031 magnitude contour did not exercise a folded greatest-time branch")
	}
	data, err := geojson.MarshalSolarEclipse(partial, nil)
	if err != nil {
		t.Fatalf("MarshalSolarEclipse: %v", err)
	}
	line := featureWithRole(t, decodeCollection(t, data), "magnitude-line")
	assertTimedLineAligned(t, line)
}

func TestMarshalSolarEclipseAllowsSingleLimitCentrality(t *testing.T) {
	date := time.Date(2003, time.May, 30, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
		Step: 20 * time.Minute, BoundaryPoints: 24,
	})
	if !ok {
		t.Fatal("expected solar partial footprints")
	}
	central, ok := eclipse.SolarEclipseCentralPath(date, eclipse.SolarEclipsePathOptions{Step: 10 * time.Minute})
	if !ok || central.Eclipse.Centrality != eclipse.SolarEclipseCentralOneLimit {
		t.Fatalf("expected one-limit central eclipse, got ok=%v centrality=%s", ok, central.Eclipse.Centrality)
	}
	data, err := geojson.MarshalSolarEclipse(partial, &central)
	if err != nil {
		t.Fatalf("MarshalSolarEclipse: %v", err)
	}
	collection := decodeCollection(t, data)
	if len(featuresWithRole(collection, "center-line")) != 1 || len(featuresWithRole(collection, "central-band")) != 1 {
		t.Fatal("single-limit central eclipse should export both its center line and central band")
	}
}

func TestMarshalSolarEclipse20330330SeparatesOpenSweepsAcrossClosedPhase(t *testing.T) {
	date := time.Date(2033, time.March, 30, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
		Step: 10 * time.Minute, BoundaryPoints: 180, CentralShadowStep: 2 * time.Minute,
	})
	if !ok {
		t.Fatal("expected 2033-03-30 solar eclipse")
	}
	central, ok := eclipse.SolarEclipseCentralPath(date, eclipse.SolarEclipsePathOptions{Step: 5 * time.Minute})
	if !ok {
		t.Fatal("expected 2033-03-30 central path")
	}
	data, err := geojson.MarshalSolarEclipse(partial, &central)
	if err != nil {
		t.Fatalf("MarshalSolarEclipse: %v", err)
	}
	assertClosedMultiPolygon(t, featureWithRole(t, decodeCollection(t, data), "central-band"))
}

func TestMarshalSolarEclipseAllowsLowSampleOpenFootprints(t *testing.T) {
	for _, fixture := range []struct {
		date time.Time
		step time.Duration
	}{
		{time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC), 5 * time.Minute},
		{time.Date(2025, time.March, 29, 0, 0, 0, 0, time.UTC), 5 * time.Minute},
	} {
		partial, ok := eclipse.SolarEclipsePartialFootprints(fixture.date, eclipse.SolarEclipsePartialFootprintOptions{
			Step: fixture.step, BoundaryPoints: 12,
		})
		if !ok {
			t.Fatalf("%s: expected solar partial footprints", fixture.date.Format("2006-01-02"))
		}
		if _, err := geojson.MarshalSolarEclipse(partial, nil); err != nil {
			t.Fatalf("%s low-sample GeoJSON: %v", fixture.date.Format("2006-01-02"), err)
		}
	}
}

func TestMarshalSolarEclipseWithTimeMarkers(t *testing.T) {
	date := time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
		Step:           20 * time.Minute,
		BoundaryPoints: 36,
	})
	if !ok {
		t.Fatal("expected solar partial footprints")
	}
	central, ok := eclipse.SolarEclipseCentralPath(date, eclipse.SolarEclipsePathOptions{Step: 5 * time.Minute})
	if !ok {
		t.Fatal("expected solar central path")
	}

	data, err := geojson.MarshalSolarEclipseWithTimeMarkers(partial, &central, geojson.TimeMarkerOptions{
		Step:     time.Hour,
		Location: time.FixedZone("CST", 8*60*60),
	})
	if err != nil {
		t.Fatalf("MarshalSolarEclipseWithTimeMarkers: %v", err)
	}
	collection := decodeCollection(t, data)
	markers := featuresWithRole(collection, "time-marker")
	if len(markers) == 0 {
		t.Fatal("solar eclipse has no time markers")
	}
	for _, marker := range markers {
		if marker.Properties["source_role"] != "center-line" {
			t.Fatalf("time marker source_role=%v, want center-line", marker.Properties["source_role"])
		}
		label, ok := marker.Properties["label"].(string)
		if !ok || len(label) != len("15:04") || label[2] != ':' {
			t.Fatalf("invalid time marker label %q", label)
		}
	}
}

func TestMarshalLunarEclipseUsesRequestedBoundarySampling(t *testing.T) {
	info, ok := eclipse.LunarEclipseOnDate(time.Date(2026, time.March, 3, 0, 0, 0, 0, time.UTC))
	if !ok {
		t.Fatal("expected lunar eclipse")
	}
	data, err := geojson.MarshalLunarEclipse(info, 24)
	if err != nil {
		t.Fatalf("MarshalLunarEclipse: %v", err)
	}
	collection := decodeCollection(t, data)
	assertRoles(t, collection, "visible-at-p1", "visible-at-p4",
		"p1-horizon", "p4-horizon", "visible-during-eclipse",
		"visible-throughout-eclipse", "greatest")
	assertCollectionCoordinates(t, collection)
	visible := featureWithRole(t, collection, "visible-at-p1")
	if got := visible.Properties["boundary_points"]; got != float64(24) {
		t.Fatalf("boundary_points=%v, want 24", got)
	}
	assertClosedMultiPolygon(t, visible)
	horizon := featureWithRole(t, collection, "p1-horizon")
	var lines [][][]float64
	if err := json.Unmarshal(horizon.Geometry.Coordinates, &lines); err != nil {
		t.Fatalf("decode P1 horizon: %v", err)
	}
	pointCount := 0
	for _, line := range lines {
		pointCount += len(line)
	}
	if pointCount < 24 {
		t.Fatalf("P1 horizon has %d points, want at least 24", pointCount)
	}
}

func TestMarshalLunarEclipseWithOptions(t *testing.T) {
	date := time.Date(2029, 1, 1, 12, 0, 0, 0, time.FixedZone("CST", 8*3600))
	info, ok := eclipse.LunarEclipseOnDate(date)
	if !ok {
		t.Fatal("missing lunar eclipse")
	}
	defaultData, err := geojson.MarshalLunarEclipse(info, 360)
	if err != nil {
		t.Fatal(err)
	}
	skippedData, err := geojson.MarshalLunarEclipseWithOptions(info, 360, geojson.LunarEclipseOptions{
		SkipRoles: []string{"visible-during-eclipse", "visible-throughout-eclipse"},
	})
	if err != nil {
		t.Fatal(err)
	}
	skipped := decodeCollection(t, skippedData)
	assertRoles(t, skipped, "visible-at-p1", "visible-at-p4", "p1-horizon", "p4-horizon", "greatest")
	for _, role := range []string{"visible-during-eclipse", "visible-throughout-eclipse"} {
		if len(featuresWithRole(skipped, role)) != 0 {
			t.Fatalf("skipped role %s is still present", role)
		}
	}
	// 跳过包络不得改变其余要素的几何。
	band := featureWithRole(t, skipped, "visible-at-p1")
	expected := featureWithRole(t, decodeCollection(t, defaultData), "visible-at-p1")
	if got, want := string(band.Geometry.Coordinates), string(expected.Geometry.Coordinates); got != want {
		t.Fatal("skipping the envelopes changed the P1 hemisphere")
	}

	// 只跳过一块包络时另一块照常输出，而且不再为被跳过的那块成环。
	partialData, err := geojson.MarshalLunarEclipseWithOptions(info, 360, geojson.LunarEclipseOptions{
		SkipRoles: []string{"visible-throughout-eclipse"},
	})
	if err != nil {
		t.Fatal(err)
	}
	partial := decodeCollection(t, partialData)
	assertRoles(t, partial, "visible-at-p1", "visible-during-eclipse")
	if len(featuresWithRole(partial, "visible-throughout-eclipse")) != 0 {
		t.Fatal("skipped envelope role is still present")
	}

	coarseData, err := geojson.MarshalLunarEclipseWithOptions(info, 360, geojson.LunarEclipseOptions{
		EnvelopeLongitudePoints: 180,
		EnvelopeSweepSamples:    16,
	})
	if err != nil {
		t.Fatal(err)
	}
	coarse := decodeCollection(t, coarseData)
	for _, role := range []string{"visible-during-eclipse", "visible-throughout-eclipse"} {
		feature := featureWithRole(t, coarse, role)
		if got := feature.Properties["longitude_points"]; got != float64(180) {
			t.Fatalf("%s longitude_points=%v, want 180", role, got)
		}
		assertClosedMultiPolygon(t, feature)
	}

	if _, err := geojson.MarshalLunarEclipseWithOptions(info, 360, geojson.LunarEclipseOptions{
		SkipRoles: []string{
			"visible-at-p1", "visible-at-p4", "visible-during-eclipse", "visible-throughout-eclipse",
			"penumbra-moonset", "penumbra-moonrise",
			"p1-horizon", "p4-horizon", "greatest",
		},
	}); err == nil {
		t.Fatal("dropping every Feature must fail")
	}

	// 采样档位取值：0 或负值等于默认（逐字节相同），正值收敛到上下限。
	for _, samples := range []int{-5, 0} {
		data, err := geojson.MarshalLunarEclipseWithOptions(info, 360, geojson.LunarEclipseOptions{EnvelopeSweepSamples: samples})
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != string(defaultData) {
			t.Fatalf("EnvelopeSweepSamples=%d 应当与默认输出相同", samples)
		}
	}
	for _, points := range []int{-5, 0} {
		data, err := geojson.MarshalLunarEclipseWithOptions(info, 360, geojson.LunarEclipseOptions{EnvelopeLongitudePoints: points})
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != string(defaultData) {
			t.Fatalf("EnvelopeLongitudePoints=%d 应当与默认输出相同", points)
		}
	}
	for _, testCase := range []struct {
		points int
		want   float64
	}{{1, 12}, {100000, 720}} {
		data, err := geojson.MarshalLunarEclipseWithOptions(info, 360, geojson.LunarEclipseOptions{EnvelopeLongitudePoints: testCase.points})
		if err != nil {
			t.Fatal(err)
		}
		collection := decodeCollection(t, data)
		for _, role := range []string{"visible-during-eclipse", "visible-throughout-eclipse"} {
			if got := featureWithRole(t, collection, role).Properties["longitude_points"]; got != testCase.want {
				t.Fatalf("EnvelopeLongitudePoints=%d 的 %s longitude_points=%v，期望 %v", testCase.points, role, got, testCase.want)
			}
		}
	}
	clampedData, err := geojson.MarshalLunarEclipseWithOptions(info, 360, geojson.LunarEclipseOptions{EnvelopeSweepSamples: 1})
	if err != nil {
		t.Fatal(err)
	}
	clamped := decodeCollection(t, clampedData)
	for _, role := range []string{"visible-during-eclipse", "visible-throughout-eclipse"} {
		assertClosedMultiPolygon(t, featureWithRole(t, clamped, role))
	}
}

func TestMarshalLunarEclipseWithTimeMarkers(t *testing.T) {
	info, ok := eclipse.LunarEclipseOnDate(time.Date(2026, time.March, 3, 0, 0, 0, 0, time.UTC))
	if !ok {
		t.Fatal("expected lunar eclipse")
	}
	data, err := geojson.MarshalLunarEclipseWithTimeMarkers(info, 24, geojson.TimeMarkerOptions{Step: time.Hour})
	if err != nil {
		t.Fatalf("MarshalLunarEclipseWithTimeMarkers: %v", err)
	}
	collection := decodeCollection(t, data)
	markers := featuresWithRole(collection, "time-marker")
	if len(markers) == 0 {
		t.Fatal("lunar eclipse has no time markers")
	}
	for _, marker := range markers {
		if marker.Properties["source_role"] != "sublunar-track" {
			t.Fatalf("time marker source_role=%v, want sublunar-track", marker.Properties["source_role"])
		}
	}
	firstLabel, _ := markers[0].Properties["label"].(string)
	lastLabel, _ := markers[len(markers)-1].Properties["label"].(string)
	if firstLabel != "09:00" || lastLabel != "14:00" {
		t.Fatalf("lunar marker endpoints = %q..%q, want 09:00..14:00", firstLabel, lastLabel)
	}
}

func TestMarshalLunarEclipseRejectsInvalidContactOrder(t *testing.T) {
	info, ok := eclipse.LunarEclipseOnDate(time.Date(2026, time.March, 3, 0, 0, 0, 0, time.UTC))
	if !ok {
		t.Fatal("expected lunar eclipse")
	}
	info.Maximum = info.PenumbralStart.Add(-time.Minute)
	if _, err := geojson.MarshalLunarEclipse(info, 24); err == nil {
		t.Fatal("reversed lunar eclipse contacts were accepted")
	}
}

func TestMarshalStarOccultationSplitsAntimeridian(t *testing.T) {
	start := time.Date(2025, time.June, 5, 17, 45, 0, 0, time.UTC)
	center := occultationSamples(start, []float64{160, 175, -175, -160}, []float64{8, 4, 0, -4})
	north := occultationSamples(start, []float64{158, 174, -174, -158}, []float64{18, 14, 10, 6})
	south := occultationSamples(start, []float64{162, 176, -176, -162}, []float64{-2, -6, -10, -14})
	path := moon.StarOccultationPath{
		TargetID:      "HR 4799",
		Start:         north[0],
		Greatest:      center[2],
		End:           north[len(north)-1],
		Complete:      true,
		CenterLine:    center,
		NorthernLimit: north,
		SouthernLimit: south,
		Step:          time.Hour,
	}

	data, err := geojson.MarshalStarOccultation(path)
	if err != nil {
		t.Fatalf("MarshalStarOccultation: %v", err)
	}
	collection := decodeCollection(t, data)
	assertRoles(t, collection, "occultation-band", "center-line", "north-limit", "south-limit", "start", "greatest", "end")
	assertCollectionCoordinates(t, collection)
	centerFeature := featureWithRole(t, collection, "center-line")
	var lines [][][]float64
	if err := json.Unmarshal(centerFeature.Geometry.Coordinates, &lines); err != nil {
		t.Fatalf("decode center line: %v", err)
	}
	if len(lines) != 2 {
		t.Fatalf("center line has %d antimeridian segments, want 2", len(lines))
	}
	for _, line := range lines {
		for index := 1; index < len(line); index++ {
			if math.Abs(line[index][0]-line[index-1][0]) > 180 {
				t.Fatalf("center line still crosses antimeridian: %#v", line)
			}
		}
	}
	assertTimedLineAligned(t, centerFeature)
}

func TestMarshalStarOccultationAntaresCenterLineStaysInsideFootprintBand(t *testing.T) {
	start := time.Date(2026, time.February, 11, 0, 0, 0, 0, time.UTC)
	paths, err := moon.FindStarOccultationPaths(
		start,
		start.Add(24*time.Hour),
		moon.StarCoordinate{
			ID: "Antares", RA: 247.3516666666667, Dec: -26.431944444444444,
			Epoch:                          time.Date(2000, time.January, 1, 12, 0, 0, 0, time.UTC),
			Frame:                          moon.CoordinateFrameJ2000,
			ProperMotionRACosDecMasPerYear: -10,
			ProperMotionDecMasPerYear:      -20,
			ParallaxMas:                    24,
		},
		moon.OccultationPathOptions{Step: 5 * time.Minute, TargetSpacingKM: 200},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindStarOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	if len(paths[0].Footprints) == 0 {
		t.Fatal("Antares path has no instantaneous footprints")
	}
	data, err := geojson.MarshalStarOccultation(paths[0])
	if err != nil {
		t.Fatalf("MarshalStarOccultation: %v", err)
	}
	collection := decodeCollection(t, data)
	assertRiseSetBoundaryFeatures(t, collection, "moon")
	band := featureWithRole(t, collection, "occultation-band")
	var polygons [][][][]float64
	if err := json.Unmarshal(band.Geometry.Coordinates, &polygons); err != nil {
		t.Fatalf("decode occultation band: %v", err)
	}
	if len(polygons)*10 >= len(paths[0].Footprints) {
		t.Fatalf("merged occultation sweep retained %d polygons for %d instantaneous footprints",
			len(polygons), len(paths[0].Footprints))
	}
	center := featureWithRole(t, collection, "center-line")
	var lines [][][]float64
	if err := json.Unmarshal(center.Geometry.Coordinates, &lines); err != nil {
		t.Fatalf("decode center line: %v", err)
	}
	for lineIndex, line := range lines {
		for index := 1; index < len(line); index++ {
			for sample := 0; sample <= 10; sample++ {
				fraction := float64(sample) / 10
				longitude := line[index-1][0] + fraction*(line[index][0]-line[index-1][0])
				latitude := line[index-1][1] + fraction*(line[index][1]-line[index-1][1])
				if !geoJSONMultiPolygonContains(polygons, longitude, latitude) {
					t.Fatalf("center line[%d] segment %d sample %d lies outside footprint band at %.6f, %.6f",
						lineIndex, index-1, sample, longitude, latitude)
				}
			}
		}
	}
	for _, role := range []string{"north-limit", "south-limit"} {
		limit := featureWithRole(t, collection, role)
		var segments [][][]float64
		if err := json.Unmarshal(limit.Geometry.Coordinates, &segments); err != nil {
			t.Fatalf("decode %s: %v", role, err)
		}
		for segmentIndex, segment := range segments {
			for index := 1; index < len(segment); index++ {
				if distance := geoJSONCoordinateDistanceKM(segment[index-1], segment[index]); distance > 750+1e-6 {
					t.Fatalf("%s segment %d still spans %.1f km branch change", role, segmentIndex, distance)
				}
			}
		}
		assertTimedLineAligned(t, limit)
	}
}

func TestMarshalStarOccultationAntaresCompactBandContainsCenterLine(t *testing.T) {
	start := time.Date(2026, time.February, 11, 0, 0, 0, 0, time.UTC)
	paths, err := moon.FindStarOccultationPaths(
		start, start.Add(24*time.Hour),
		moon.StarCoordinate{
			ID: "Antares", RA: 247.3516666666667, Dec: -26.431944444444444,
			Epoch:                          time.Date(2000, time.January, 1, 12, 0, 0, 0, time.UTC),
			Frame:                          moon.CoordinateFrameJ2000,
			ProperMotionRACosDecMasPerYear: -10,
			ProperMotionDecMasPerYear:      -20,
			ParallaxMas:                    24,
		},
		moon.OccultationPathOptions{
			Step: 5 * time.Minute, TargetSpacingKM: 200,
			DisableRiseSet: true, DisableFootprints: true,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindStarOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	path := paths[0]
	if len(path.Footprints) != 0 || len(path.BandFootprints) == 0 {
		t.Fatalf("dense/compact footprint counts=%d/%d, want zero/nonzero", len(path.Footprints), len(path.BandFootprints))
	}
	data, err := geojson.MarshalStarOccultation(path)
	if err != nil {
		t.Fatalf("MarshalStarOccultation: %v", err)
	}
	band := featureWithRole(t, decodeCollection(t, data), "occultation-band")
	for index, point := range path.CenterLine {
		if !geometryContainsPoint(t, band.Geometry, point.Longitude, point.Latitude) {
			t.Fatalf("compact stellar band excludes center-line sample %d at %.6f, %.6f", index, point.Longitude, point.Latitude)
		}
	}
}

func TestMarshalStarOccultationHandlesExactAntimeridian(t *testing.T) {
	start := time.Date(2025, time.June, 5, 17, 45, 0, 0, time.UTC)
	center := occultationSamples(start, []float64{-180, 180, 150}, []float64{2, 1, 0})
	north := occultationSamples(start, []float64{-179, 179, 178}, []float64{12, 11, 10})
	south := occultationSamples(start, []float64{-179, 179, 178}, []float64{-8, -9, -10})
	path := moon.StarOccultationPath{
		TargetID: "exact-antimeridian", Start: north[0], Greatest: center[1], End: north[2],
		Complete: true, CenterLine: center, NorthernLimit: north, SouthernLimit: south,
		Step: time.Hour,
	}
	if _, err := geojson.MarshalStarOccultation(path); err != nil {
		t.Fatalf("exact antimeridian path: %v", err)
	}
}

func TestMarshalStarOccultationWithTimeMarkers(t *testing.T) {
	start := time.Date(2025, time.June, 5, 17, 45, 0, 0, time.UTC)
	center := occultationSamples(start, []float64{20, 30, 40, 50}, []float64{2, 1, 0, -1})
	north := occultationSamples(start, []float64{20, 30, 40, 50}, []float64{12, 11, 10, 9})
	south := occultationSamples(start, []float64{20, 30, 40, 50}, []float64{-8, -9, -10, -11})
	path := moon.StarOccultationPath{
		TargetID:      "HR 4799",
		Start:         north[0],
		Greatest:      center[1],
		End:           north[len(north)-1],
		Complete:      true,
		CenterLine:    center,
		NorthernLimit: north,
		SouthernLimit: south,
		Step:          time.Hour,
	}

	data, err := geojson.MarshalStarOccultationWithTimeMarkers(path, geojson.TimeMarkerOptions{Step: time.Hour})
	if err != nil {
		t.Fatalf("MarshalStarOccultationWithTimeMarkers: %v", err)
	}
	collection := decodeCollection(t, data)
	markers := featuresWithRole(collection, "time-marker")
	if len(markers) != 3 {
		t.Fatalf("got %d time markers, want 3", len(markers))
	}
	for _, marker := range markers {
		if marker.Properties["source_role"] != "center-line" {
			t.Fatalf("time marker source_role=%v, want center-line", marker.Properties["source_role"])
		}
		if marker.Geometry.Type != "Point" {
			t.Fatalf("time marker geometry=%q, want Point", marker.Geometry.Type)
		}
	}
}

func TestMarshalStarOccultationSplitsImpossibleBoundaryJump(t *testing.T) {
	start := time.Date(2026, time.February, 11, 10, 0, 0, 0, time.UTC)
	times := []time.Time{
		start,
		start.Add(time.Minute),
		start.Add(time.Minute + time.Second),
		start.Add(2 * time.Minute),
	}
	series := func(longitudes, latitudes []float64) []moon.OccultationPathPoint {
		points := make([]moon.OccultationPathPoint, len(times))
		for index := range times {
			points[index] = moon.OccultationPathPoint{
				Time: times[index], Longitude: longitudes[index], Latitude: latitudes[index], MoonAltitude: 20,
			}
		}
		return points
	}
	center := series([]float64{0, 0.1, 0.2, 0.3}, []float64{0, 0, 0, 0})
	north := series([]float64{0, 0.1, 30, 30.1}, []float64{10, 10, 10, 10})
	south := series([]float64{0, 0.1, 0.2, 0.3}, []float64{-10, -10, -10, -10})
	path := moon.StarOccultationPath{
		TargetID: "branch-jump", Start: north[0], Greatest: center[2], End: north[3],
		Complete: true, CenterLine: center, NorthernLimit: north, SouthernLimit: south, Step: time.Second,
	}
	data, err := geojson.MarshalStarOccultation(path)
	if err != nil {
		t.Fatalf("MarshalStarOccultation: %v", err)
	}
	northFeature := featureWithRole(t, decodeCollection(t, data), "north-limit")
	var lines [][][]float64
	if err := json.Unmarshal(northFeature.Geometry.Coordinates, &lines); err != nil {
		t.Fatalf("decode north limit: %v", err)
	}
	if len(lines) != 2 {
		t.Fatalf("north limit segment count = %d, want 2 around branch jump", len(lines))
	}
	for _, line := range lines {
		for index := 1; index < len(line); index++ {
			if jump := math.Abs(line[index][0] - line[index-1][0]); jump > 5 {
				t.Fatalf("north limit still contains %.1f degree branch jump", jump)
			}
		}
	}
	assertTimedLineAligned(t, northFeature)
	assertOccultationBandPolygonCount(t, decodeCollection(t, data), "occultation-band", 2)
}

func TestMarshalStarOccultationPreservesEndpointBranchFragments(t *testing.T) {
	start := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	for _, count := range []int{2, 3} {
		t.Run(fmt.Sprintf("%d points", count), func(t *testing.T) {
			path := endpointBranchJumpPath(start, count)
			data, err := geojson.MarshalStarOccultation(path)
			if err != nil {
				t.Fatalf("MarshalStarOccultation: %v", err)
			}
			collection := decodeCollection(t, data)
			north := featureWithRole(t, collection, "north-limit")
			assertTimedLineAligned(t, north)
			timeSegments, ok := north.Properties["times"].([]interface{})
			if !ok || len(timeSegments) != 2 {
				t.Fatalf("north-limit time segments = %#v, want two discontinuous fragments", north.Properties["times"])
			}
			first := timeSegments[0].([]interface{})[0]
			lastSegment := timeSegments[len(timeSegments)-1].([]interface{})
			last := lastSegment[len(lastSegment)-1]
			if first != path.Start.Time.Format(time.RFC3339Nano) || last != path.End.Time.Format(time.RFC3339Nano) {
				t.Fatalf("north-limit time span = %v..%v, want %s..%s",
					first, last, path.Start.Time.Format(time.RFC3339Nano), path.End.Time.Format(time.RFC3339Nano))
			}

			band := featureWithRole(t, collection, "occultation-band")
			if count == 2 && band.Geometry.Type != "MultiLineString" {
				t.Fatalf("two-point discontinuous band geometry = %q, want MultiLineString endpoint sections", band.Geometry.Type)
			}
			if count == 3 {
				if band.Geometry.Type != "GeometryCollection" {
					t.Fatalf("partially continuous band geometry = %q, want GeometryCollection", band.Geometry.Type)
				}
				var geometries []struct {
					Type string `json:"type"`
				}
				if err := json.Unmarshal(band.Geometry.Geometries, &geometries); err != nil {
					t.Fatalf("decode band geometries: %v", err)
				}
				if len(geometries) != 2 || geometries[0].Type != "MultiPolygon" || geometries[1].Type != "MultiLineString" {
					t.Fatalf("band geometries = %#v, want polygon sweep plus endpoint sections", geometries)
				}
			}
		})
	}
}

func endpointBranchJumpPath(start time.Time, count int) moon.StarOccultationPath {
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

func TestMarshalPlanetOccultationSplitsImpossibleFallbackBandJump(t *testing.T) {
	start := time.Date(2026, time.February, 11, 10, 0, 0, 0, time.UTC)
	times := []time.Time{
		start,
		start.Add(time.Minute),
		start.Add(time.Minute + time.Second),
		start.Add(2 * time.Minute),
	}
	series := func(longitudes, latitudes []float64) []moon.OccultationPathPoint {
		points := make([]moon.OccultationPathPoint, len(times))
		for index := range times {
			points[index] = moon.OccultationPathPoint{
				Time: times[index], Longitude: longitudes[index], Latitude: latitudes[index], MoonAltitude: 20,
			}
		}
		return points
	}
	center := series([]float64{0, 0.1, 30, 30.1}, []float64{0, 0, 0, 0})
	north := series([]float64{0, 0.1, 30, 30.1}, []float64{10, 10, 10, 10})
	south := series([]float64{0, 0.1, 30, 30.1}, []float64{-10, -10, -10, -10})
	path := moon.PlanetOccultationPath{
		Planet: moon.OccultationSaturn, TargetID: "branch-jump",
		Start: north[0], Greatest: center[2], End: north[3], Complete: true,
		CenterLine: center, NorthernLimit: north, SouthernLimit: south, Step: time.Second,
	}
	data, err := geojson.MarshalPlanetOccultation(path)
	if err != nil {
		t.Fatalf("MarshalPlanetOccultation: %v", err)
	}
	assertOccultationBandPolygonCount(t, decodeCollection(t, data), "partial-band", 2)
}

func TestTimeMarkerInterpolationUsesShortestAntimeridianPath(t *testing.T) {
	start := time.Date(2025, time.June, 5, 17, 45, 0, 0, time.UTC)
	center := occultationSamples(start, []float64{170, -170}, []float64{2, 0})
	north := occultationSamples(start, []float64{168, -168}, []float64{12, 10})
	south := occultationSamples(start, []float64{172, -172}, []float64{-8, -10})
	path := moon.StarOccultationPath{
		TargetID:      "HR 4799",
		Start:         north[0],
		Greatest:      center[0],
		End:           north[len(north)-1],
		Complete:      true,
		CenterLine:    center,
		NorthernLimit: north,
		SouthernLimit: south,
		Step:          time.Hour,
	}

	data, err := geojson.MarshalStarOccultationWithTimeMarkers(path, geojson.TimeMarkerOptions{Step: 15 * time.Minute})
	if err != nil {
		t.Fatalf("MarshalStarOccultationWithTimeMarkers: %v", err)
	}
	markers := featuresWithRole(decodeCollection(t, data), "time-marker")
	if len(markers) != 3 {
		t.Fatalf("got %d time markers, want 3", len(markers))
	}
	for _, marker := range markers {
		var coordinate []float64
		if err := json.Unmarshal(marker.Geometry.Coordinates, &coordinate); err != nil {
			t.Fatalf("decode time marker: %v", err)
		}
		if math.Abs(coordinate[0]) < 170 {
			t.Fatalf("time marker crossed through longitude %.6f instead of the antimeridian", coordinate[0])
		}
	}
}

func TestMarshalPlanetOccultationIncludesPartialAndTotalFootprints(t *testing.T) {
	start := time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC)
	center := occultationSamples(start, []float64{20, 30, 40}, []float64{2, 1, 0})
	north := occultationSamples(start, []float64{20, 30, 40}, []float64{12, 11, 10})
	south := occultationSamples(start, []float64{20, 30, 40}, []float64{-8, -9, -10})
	totalNorth := occultationSamples(start, []float64{22, 30, 38}, []float64{8, 7, 6})
	totalSouth := occultationSamples(start, []float64{22, 30, 38}, []float64{-4, -5, -6})
	for index := range totalNorth {
		at := start.Add(time.Duration(index+1) * 30 * time.Minute)
		totalNorth[index].Time = at
		totalSouth[index].Time = at
	}
	path := moon.PlanetOccultationPath{
		Planet:               moon.OccultationSaturn,
		TargetID:             "Saturn",
		Start:                north[0],
		Greatest:             center[1],
		End:                  north[len(north)-1],
		Complete:             true,
		CenterLine:           center,
		NorthernLimit:        north,
		SouthernLimit:        south,
		PartialFootprints:    []moon.PlanetOccultationFootprint{sampleFootprint(start.Add(time.Hour), 18, -10, 42, 14)},
		HasTotalBand:         true,
		TotalStart:           totalNorth[0],
		TotalEnd:             totalNorth[len(totalNorth)-1],
		TotalComplete:        true,
		NorthernTotalLimit:   totalNorth,
		SouthernTotalLimit:   totalSouth,
		TotalFootprints:      []moon.PlanetOccultationFootprint{sampleFootprint(start.Add(time.Hour), 23, -5, 37, 9)},
		GreatestTotalWidthKM: 2500,
		Step:                 time.Hour,
		TargetSpacingKM:      50,
	}

	data, err := geojson.MarshalPlanetOccultation(path)
	if err != nil {
		t.Fatalf("MarshalPlanetOccultation: %v", err)
	}
	collection := decodeCollection(t, data)
	assertRoles(t, collection,
		"partial-footprint", "total-footprint", "center-line", "north-limit", "south-limit",
		"north-total-limit", "south-total-limit", "start", "total-start", "greatest", "total-end", "end")
	assertCollectionCoordinates(t, collection)
	if got := featureWithRole(t, collection, "greatest").Properties["planet"]; got != "saturn" {
		t.Fatalf("planet=%v, want saturn", got)
	}
}

func TestMarshalPlanetOccultationAllowsMissingCenterLine(t *testing.T) {
	start := time.Date(2024, time.September, 5, 0, 0, 0, 0, time.UTC)
	paths, err := moon.FindPlanetOccultationPaths(
		start, start.AddDate(0, 0, 1), moon.OccultationVenus, moon.OccultationPathOptions{},
	)
	if err != nil {
		t.Fatalf("FindPlanetOccultationPaths: %v", err)
	}
	if len(paths) != 1 || len(paths[0].CenterLine) != 0 {
		t.Fatalf("unexpected Venus path count/center line: paths=%d center=%d", len(paths), len(paths[0].CenterLine))
	}

	data, err := geojson.MarshalPlanetOccultationWithTimeMarkers(
		paths[0], geojson.TimeMarkerOptions{Step: 30 * time.Minute},
	)
	if err != nil {
		t.Fatalf("MarshalPlanetOccultationWithTimeMarkers: %v", err)
	}
	collection := decodeCollection(t, data)
	if len(featuresWithRole(collection, "center-line")) != 0 || len(featuresWithRole(collection, "time-marker")) != 0 {
		t.Fatal("edge-only planetary path contains center-line features")
	}
	assertRoles(t, collection, "north-limit", "south-limit", "start", "greatest", "end")
}

func TestMarshalPlanetOccultationExportsSixRiseSetPhaseBoundaries(t *testing.T) {
	start := time.Date(2024, time.August, 21, 0, 0, 0, 0, time.UTC)
	paths, err := moon.FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), moon.OccultationSaturn,
		moon.OccultationPathOptions{Step: 10 * time.Minute},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	data, err := geojson.MarshalPlanetOccultation(paths[0])
	if err != nil {
		t.Fatalf("MarshalPlanetOccultation: %v", err)
	}
	assertRiseSetBoundaryFeatures(t, decodeCollection(t, data), "moon")
}

func TestMarshalPlanetOccultation20250105PreservesClosedPolarRiseSetBranches(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2025, time.January, 5, 0, 0, 0, 0, zone)
	paths, err := moon.FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), moon.OccultationSaturn,
		moon.OccultationPathOptions{
			Step: 20 * time.Minute, TargetSpacingKM: 900, DisableFootprints: true,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	data, err := geojson.MarshalPlanetOccultation(paths[0])
	if err != nil {
		t.Fatalf("MarshalPlanetOccultation: %v", err)
	}
	collection := decodeCollection(t, data)
	assertRiseSetBoundaryFeatures(t, collection, "moon")
	startRise := riseSetBoundaryFeature(t, collection, "start", "rise")
	startSet := riseSetBoundaryFeature(t, collection, "start", "set")
	if !riseSetFeaturesShareEndpointInRegion(t, startRise, startSet, -60, 60, 70, 90) {
		t.Fatal("serialized start moonrise/moonset curves do not share their polar direction junction")
	}
	if !riseSetFeatureSegmentsShareEndpointInRegion(t, startSet, -60, 60, 70, 90) &&
		!riseSetFeatureHasInteriorVertexInRegion(t, startSet, -60, 60, 70, 90) {
		t.Fatal("serialized moonset start-phase branches do not preserve their polar fold vertex")
	}
	for _, feature := range featuresWithRole(collection, "visibility-boundary") {
		if feature.Properties["phase"] == "horizon" {
			continue
		}
		assertRiseSetFeatureHasNoInstantaneousBranchJump(t, feature)
	}
}

func TestMarshalPlanetOccultation20250630MarsClosesMoonsetAndBand(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2025, time.June, 30, 0, 0, 0, 0, zone)
	paths, err := moon.FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), moon.OccultationMars,
		moon.OccultationPathOptions{
			Step: 20 * time.Minute, TargetSpacingKM: 900, RiseSetStep: 5 * time.Minute,
			DisableFootprints: true, IncludeFootprintTimeline: true,
			FootprintTimelineStep: 5 * time.Minute,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	data, err := geojson.MarshalPlanetOccultation(paths[0])
	if err != nil {
		t.Fatalf("MarshalPlanetOccultation: %v", err)
	}
	collection := decodeCollection(t, data)
	assertRiseSetBoundaryFeatures(t, collection, "moon")
	startSet := riseSetBoundaryFeature(t, collection, "start", "set")
	greatestSet := riseSetBoundaryFeature(t, collection, "greatest", "set")
	endSet := riseSetBoundaryFeature(t, collection, "end", "set")
	if !riseSetFeaturesShareVertexInRegion(t, startSet, endSet, -90, -70, -40, -10) {
		t.Fatal("serialized moonset start/end curves do not share the first narrow phase junction")
	}
	if !riseSetFeaturesShareVertexInRegion(t, greatestSet, endSet, -90, -70, -40, -10) {
		t.Fatal("serialized moonset greatest/end curves do not share the second narrow phase junction")
	}

	band := featureWithRole(t, collection, "partial-band")
	var polygons [][][][]float64
	if err := json.Unmarshal(band.Geometry.Coordinates, &polygons); err != nil {
		t.Fatalf("decode partial-band: %v", err)
	}
	if band.Geometry.Type != "MultiPolygon" || len(polygons) == 0 || len(polygons) > 2 {
		t.Fatalf("partial-band geometry=%s polygons=%d, want one spherical band split at most once by the antimeridian", band.Geometry.Type, len(polygons))
	}
	for _, polygon := range polygons {
		for _, ring := range polygon {
			if len(ring) < 4 || ring[0][0] != ring[len(ring)-1][0] || ring[0][1] != ring[len(ring)-1][1] {
				t.Fatal("partial-band contains an unclosed polygon ring")
			}
			assertGeoJSONRingHasNoShortHairpins(t, "partial-band", ring, 50, 75, 24)
		}
	}
}

func TestMarshalPlanetOccultation20250729MarsUsesClosedPolarBand(t *testing.T) {
	fixture := mars20250729TestFixture(t, 5*time.Minute, false)
	path, collection := fixture.path, fixture.collection
	if len(path.CenterLine) != 0 || len(path.RiseSetCurves) != 3 || len(path.RiseSetCurves[2].Segments) < 2 {
		t.Fatalf("unexpected polar path topology: center=%d curves=%d end-rise-segments=%d",
			len(path.CenterLine), len(path.RiseSetCurves), len(path.RiseSetCurves[2].Segments))
	}
	band := featureWithRole(t, collection, "partial-band")
	if band.Geometry.Type != "MultiPolygon" {
		t.Fatalf("partial-band geometry=%q, want MultiPolygon", band.Geometry.Type)
	}
	if authoritative, ok := band.Properties["static_band_authoritative"].(bool); !ok || !authoritative {
		t.Fatalf("partial-band static_band_authoritative=%v, want true", band.Properties["static_band_authoritative"])
	}
	if source := band.Properties["source"]; source != "visible-footprint-sweep" {
		t.Fatalf("partial-band source=%v, want visible-footprint-sweep", source)
	}
	var polygons [][][][]float64
	if err := json.Unmarshal(band.Geometry.Coordinates, &polygons); err != nil {
		t.Fatalf("decode partial-band: %v", err)
	}
	if len(polygons) != 1 {
		t.Fatalf("partial-band polygons=%d, want one continuous polar band", len(polygons))
	}
	for polygonIndex, polygon := range polygons {
		for ringIndex, ring := range polygon {
			if len(ring) < 4 || ring[0][0] != ring[len(ring)-1][0] || ring[0][1] != ring[len(ring)-1][1] {
				t.Fatalf("partial-band polygon %d ring %d is not closed", polygonIndex, ringIndex)
			}
		}
	}
	greatest := featureWithRole(t, collection, "greatest")
	var greatestPoint []float64
	if err := json.Unmarshal(greatest.Geometry.Coordinates, &greatestPoint); err != nil {
		t.Fatalf("decode greatest point: %v", err)
	}
	totalBand := featureWithRole(t, collection, "total-band")
	var totalPolygons [][][][]float64
	if err := json.Unmarshal(totalBand.Geometry.Coordinates, &totalPolygons); err != nil {
		t.Fatalf("decode total-band: %v", err)
	}
	for _, testPoint := range []struct {
		name   string
		lon    float64
		lat    float64
		inside bool
	}{
		{name: "selected", lon: -134.2280, lat: -78.0725, inside: true},
		{name: "visible-sweep", lon: -129.8230, lat: -77.4030, inside: true},
	} {
		partialInside := geoJSONMultiPolygonContains(polygons, testPoint.lon, testPoint.lat)
		totalInside := geoJSONMultiPolygonContains(totalPolygons, testPoint.lon, testPoint.lat)
		if partialInside != testPoint.inside || totalInside != testPoint.inside {
			t.Fatalf("%s point %.4f, %.4f partial=%v total=%v, want inside=%v", testPoint.name, testPoint.lon, testPoint.lat, partialInside, totalInside, testPoint.inside)
		}
	}
	totalRings := geoJSONMultiPolygonOuterRings(t, totalBand)
	greatestPath := [][]geodata.GeoPoint{{
		{Longitude: greatestPoint[0], Latitude: greatestPoint[1]},
	}}
	if miss := geodata.SphericalPolygonsPathMissDistanceKM(totalRings, greatestPath, false); miss > 10 {
		t.Fatalf("total-band excludes greatest point %.6f, %.6f by %.1f km", greatestPoint[0], greatestPoint[1], miss)
	}
}

func TestMarshalPlanetOccultationWithoutFootprintsUsesBands(t *testing.T) {
	start := time.Date(2024, time.August, 21, 0, 0, 0, 0, time.UTC)
	paths, err := moon.FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), moon.OccultationSaturn,
		moon.OccultationPathOptions{Step: 10 * time.Minute, DisableFootprints: true},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	data, err := geojson.MarshalPlanetOccultation(paths[0])
	if err != nil {
		t.Fatalf("MarshalPlanetOccultation: %v", err)
	}
	collection := decodeCollection(t, data)
	assertRoles(t, collection, "partial-band", "total-band", "center-line", "north-limit", "south-limit",
		"north-total-limit", "south-total-limit", "visibility-boundary", "start", "total-start", "greatest", "total-end", "end")
	if len(featuresWithRole(collection, "partial-footprint")) != 0 || len(featuresWithRole(collection, "total-footprint")) != 0 {
		t.Fatal("disabled instantaneous footprints were serialized")
	}
	assertRiseSetBoundaryFeatures(t, collection, "moon")
}

func TestMarshalPlanetOccultationCompactBandHandlesTangentBranchConvergence(t *testing.T) {
	if testing.Short() {
		t.Skip("slow regression")
	}
	for _, date := range []time.Time{
		time.Date(1954, time.June, 30, 0, 0, 0, 0, time.UTC),
		time.Date(1962, time.April, 1, 0, 0, 0, 0, time.UTC),
		time.Date(1965, time.June, 27, 0, 0, 0, 0, time.UTC),
	} {
		paths, err := moon.FindPlanetOccultationPaths(
			date, date.Add(24*time.Hour), moon.OccultationJupiter,
			moon.OccultationPathOptions{Step: 10 * time.Minute, DisableFootprints: true, DisableRiseSet: true},
		)
		if err != nil || len(paths) != 1 {
			t.Fatalf("%s: paths=%d err=%v, want one", date.Format("2006-01-02"), len(paths), err)
		}
		data, err := geojson.MarshalPlanetOccultation(paths[0])
		if err != nil {
			t.Fatalf("%s: MarshalPlanetOccultation: %v", date.Format("2006-01-02"), err)
		}
		assertRoles(t, decodeCollection(t, data), "partial-band", "total-band", "center-line")
	}
}

func TestMarshalPlanetOccultation20240725CompactBandsFollowCenterLine(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2024, time.July, 25, 0, 0, 0, 0, zone)
	paths, err := moon.FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), moon.OccultationSaturn,
		moon.OccultationPathOptions{
			Step: 20 * time.Minute, TargetSpacingKM: 900,
			DisableFootprints: true,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	data, err := geojson.MarshalPlanetOccultation(paths[0])
	if err != nil {
		t.Fatalf("MarshalPlanetOccultation: %v", err)
	}
	collection := decodeCollection(t, data)
	for _, test := range []struct {
		role       string
		points     []moon.OccultationPathPoint
		start, end time.Time
	}{
		{role: "partial-band", points: paths[0].CenterLine},
		{
			role: "total-band", points: paths[0].CenterLine,
			start: paths[0].TotalStart.Time, end: paths[0].TotalEnd.Time,
		},
	} {
		band := featureWithRole(t, collection, test.role)
		var polygons [][][][]float64
		if err := json.Unmarshal(band.Geometry.Coordinates, &polygons); err != nil {
			t.Fatalf("decode %s: %v", test.role, err)
		}
		if band.Geometry.Type != "MultiPolygon" || len(polygons) != 1 {
			t.Fatalf("%s geometry=%s polygons=%d, want one continuous MultiPolygon", test.role, band.Geometry.Type, len(polygons))
		}
		if compact, ok := band.Properties["compact_band"].(bool); !ok || !compact {
			t.Fatalf("%s compact_band=%v, want true", test.role, band.Properties["compact_band"])
		}
		maximumEdge := 0.0
		for _, polygon := range polygons {
			for _, ring := range polygon {
				if len(ring) < 4 || ring[0][0] != ring[len(ring)-1][0] || ring[0][1] != ring[len(ring)-1][1] {
					t.Fatalf("%s contains an unclosed polygon ring", test.role)
				}
				for pointIndex := 1; pointIndex < len(ring); pointIndex++ {
					edge := geoJSONCoordinateDistanceKM(ring[pointIndex-1], ring[pointIndex])
					maximumEdge = math.Max(maximumEdge, edge)
				}
			}
		}
		if maximumEdge > 750 {
			t.Fatalf("%s maximum edge=%.1f km, likely contains a disconnected shard", test.role, maximumEdge)
		}
		for index, point := range test.points {
			if (!test.start.IsZero() && point.Time.Before(test.start)) ||
				(!test.end.IsZero() && point.Time.After(test.end)) {
				continue
			}
			if !geometryContainsPoint(t, band.Geometry, point.Longitude, point.Latitude) {
				t.Fatalf("%s excludes center-line sample %d at %.6f, %.6f", test.role, index, point.Longitude, point.Latitude)
			}
		}
	}
}

func TestMarshalPlanetOccultation20250105CompactBandsHaveSmoothContinuousOutline(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2025, time.January, 5, 0, 0, 0, 0, zone)
	paths, err := moon.FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), moon.OccultationSaturn,
		moon.OccultationPathOptions{
			Step: 20 * time.Minute, TargetSpacingKM: 900,
			DisableFootprints: true, IncludeFootprintTimeline: true,
			FootprintTimelineStep: 5 * time.Minute,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	data, err := geojson.MarshalPlanetOccultation(paths[0])
	if err != nil {
		t.Fatalf("MarshalPlanetOccultation: %v", err)
	}
	collection := decodeCollection(t, data)
	for _, role := range []string{"partial-band", "total-band"} {
		band := featureWithRole(t, collection, role)
		var polygons [][][][]float64
		if err := json.Unmarshal(band.Geometry.Coordinates, &polygons); err != nil {
			t.Fatalf("decode %s: %v", role, err)
		}
		if band.Geometry.Type != "MultiPolygon" || len(polygons) != 1 {
			t.Fatalf("%s geometry=%s polygons=%d, want one continuous MultiPolygon", role, band.Geometry.Type, len(polygons))
		}
		maximumEdge := 0.0
		for _, polygon := range polygons {
			for _, ring := range polygon {
				if len(ring) < 4 || ring[0][0] != ring[len(ring)-1][0] || ring[0][1] != ring[len(ring)-1][1] {
					t.Fatalf("%s contains an unclosed polygon ring", role)
				}
				assertGeoJSONRingHasNoShortHairpins(t, role, ring, 35, 25, 12)
				for index := 1; index < len(ring); index++ {
					maximumEdge = math.Max(maximumEdge, geoJSONCoordinateDistanceKM(ring[index-1], ring[index]))
				}
			}
		}
		if maximumEdge > 175 {
			t.Fatalf("%s maximum edge=%.1f km, want a spatially refined outline", role, maximumEdge)
		}
		for index, point := range paths[0].CenterLine {
			if role == "total-band" && (point.Time.Before(paths[0].TotalStart.Time) || point.Time.After(paths[0].TotalEnd.Time)) {
				continue
			}
			if !geometryContainsPoint(t, band.Geometry, point.Longitude, point.Latitude) {
				t.Fatalf("%s excludes center-line sample %d at %.6f, %.6f", role, index, point.Longitude, point.Latitude)
			}
		}
	}
	partial := featureWithRole(t, collection, "partial-band")
	total := featureWithRole(t, collection, "total-band")
	for index, point := range paths[0].CenterLine {
		if point.Time.Before(paths[0].TotalStart.Time) || point.Time.After(paths[0].TotalEnd.Time) {
			continue
		}
		if !geometryContainsPoint(t, partial.Geometry, point.Longitude, point.Latitude) ||
			!geometryContainsPoint(t, total.Geometry, point.Longitude, point.Latitude) {
			t.Fatalf("total-band sample %d is not covered by both bands", index)
		}
	}
}

func TestMarshalSolarEclipseRejectsMisalignedLimits(t *testing.T) {
	date := time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{})
	if !ok {
		t.Fatal("expected solar partial footprints")
	}
	central, ok := eclipse.SolarEclipseCentralPath(date, eclipse.SolarEclipsePathOptions{})
	if !ok {
		t.Fatal("expected solar central path")
	}
	central.SouthernLimit = central.SouthernLimit[:len(central.SouthernLimit)-1]
	if _, err := geojson.MarshalSolarEclipse(partial, &central); err == nil {
		t.Fatal("misaligned solar limits were accepted")
	}
}

func TestMarshalSolarEclipseRejectsMalformedDerivedGeometry(t *testing.T) {
	date := time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC)
	valid, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
		Step: 10 * time.Minute, BoundaryPoints: 24,
	})
	if !ok {
		t.Fatal("expected solar eclipse")
	}
	t.Run("magnitude above event maximum", func(t *testing.T) {
		partial := valid
		partial.MagnitudeContours = []eclipse.SolarEclipseMagnitudeContour{{
			Magnitude: partial.Eclipse.Magnitude + 0.01,
			Segments:  [][]eclipse.SolarEclipsePathPoint{{partial.Footprints[0].Boundaries[0][0], partial.Footprints[0].Boundaries[0][1]}},
		}}
		if _, err := geojson.MarshalSolarEclipse(partial, nil); err == nil {
			t.Fatal("magnitude contour above event maximum was accepted")
		}
	})
	t.Run("invalid rise set phase", func(t *testing.T) {
		partial := valid
		partial.RiseSetCurves = append([]eclipse.SolarEclipseRiseSetCurve(nil), valid.RiseSetCurves...)
		partial.RiseSetCurves[0].Phase = eclipse.RiseSetPhase("bogus")
		if _, err := geojson.MarshalSolarEclipse(partial, nil); err == nil {
			t.Fatal("invalid solar rise/set phase was accepted")
		}
	})
	t.Run("central band footprint outside contacts", func(t *testing.T) {
		partial := valid
		partial.CentralBandFootprints = append([]eclipse.SolarEclipsePartialFootprint(nil), valid.CentralBandFootprints...)
		partial.CentralBandFootprints[0].Time = partial.U1.Time.Add(-time.Hour)
		if _, err := geojson.MarshalSolarEclipse(partial, nil); err == nil {
			t.Fatal("central-band footprint outside U1-U4 was accepted")
		}
	})
	t.Run("partial eclipse cannot contain central band footprints", func(t *testing.T) {
		partial := valid
		partial.Eclipse.Type = eclipse.SolarEclipsePartial
		if len(partial.CentralBandFootprints) == 0 {
			t.Fatal("expected central-band footprints in the fixture")
		}
		if _, err := geojson.MarshalSolarEclipse(partial, nil); err == nil {
			t.Fatal("partial eclipse central-band footprints were accepted")
		}
	})
}

func TestMarshalStarOccultationRejectsInvalidPathData(t *testing.T) {
	start := time.Date(2025, time.June, 5, 17, 45, 0, 0, time.UTC)
	valid := sampleStarOccultationPath(start)
	tests := []struct {
		name   string
		mutate func(*moon.StarOccultationPath)
	}{
		{name: "misaligned limits", mutate: func(path *moon.StarOccultationPath) {
			path.SouthernLimit = path.SouthernLimit[:len(path.SouthernLimit)-1]
		}},
		{name: "mismatched limit times", mutate: func(path *moon.StarOccultationPath) {
			path.SouthernLimit[1].Time = path.SouthernLimit[1].Time.Add(time.Second)
		}},
		{name: "non-monotonic line", mutate: func(path *moon.StarOccultationPath) {
			path.CenterLine[1].Time = path.CenterLine[0].Time
		}},
		{name: "zero event time", mutate: func(path *moon.StarOccultationPath) {
			path.Start.Time = time.Time{}
		}},
		{name: "footprint outside event", mutate: func(path *moon.StarOccultationPath) {
			path.Footprints = []moon.OccultationFootprint{{
				Time: path.End.Time.Add(time.Second),
				Polygons: [][]moon.OccultationPathPoint{{
					{Longitude: 0, Latitude: 0}, {Longitude: 1, Latitude: 0}, {Longitude: 0, Latitude: 1},
				}},
			}}
		}},
		{name: "compact footprint outside event", mutate: func(path *moon.StarOccultationPath) {
			path.BandFootprints = []moon.OccultationFootprint{sampleFootprint(
				path.End.Time.Add(time.Second), 10, -10, 20, 10,
			)}
		}},
		{name: "empty footprint", mutate: func(path *moon.StarOccultationPath) {
			path.Footprints = []moon.OccultationFootprint{{Time: path.Greatest.Time}}
		}},
		{name: "mismatched footprint point time", mutate: func(path *moon.StarOccultationPath) {
			footprint := sampleFootprint(path.Greatest.Time, 10, -10, 20, 10)
			footprint.Polygons[0][0].Time = footprint.Time.Add(time.Second)
			path.Footprints = []moon.OccultationFootprint{footprint}
		}},
		{name: "non-finite footprint altitude", mutate: func(path *moon.StarOccultationPath) {
			footprint := sampleFootprint(path.Greatest.Time, 10, -10, 20, 10)
			footprint.Polygons[0][0].MoonAltitude = math.NaN()
			path.Footprints = []moon.OccultationFootprint{footprint}
		}},
		{name: "negative footprint width", mutate: func(path *moon.StarOccultationPath) {
			footprint := sampleFootprint(path.Greatest.Time, 10, -10, 20, 10)
			footprint.Polygons[0][0].WidthKM = -1
			path.Footprints = []moon.OccultationFootprint{footprint}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := valid
			path.CenterLine = append([]moon.OccultationPathPoint(nil), valid.CenterLine...)
			path.NorthernLimit = append([]moon.OccultationPathPoint(nil), valid.NorthernLimit...)
			path.SouthernLimit = append([]moon.OccultationPathPoint(nil), valid.SouthernLimit...)
			test.mutate(&path)
			if _, err := geojson.MarshalStarOccultation(path); err == nil {
				t.Fatal("invalid stellar occultation path was accepted")
			}
		})
	}
}

func TestMarshalPlanetOccultationRejectsInvalidFootprintPolygon(t *testing.T) {
	start := time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC)
	path := samplePlanetOccultationPath(start)
	path.PartialFootprints = []moon.PlanetOccultationFootprint{sampleFootprint(start, 10, -10, 20, 10)}
	path.PartialFootprints[0].Polygons = append(path.PartialFootprints[0].Polygons, []moon.OccultationPathPoint{
		{Longitude: 30, Latitude: 0}, {Longitude: 31, Latitude: 0},
	})
	if _, err := geojson.MarshalPlanetOccultation(path); err == nil {
		t.Fatal("invalid footprint polygon was silently dropped")
	}
}

func TestMarshalPlanetOccultationAllowsCompactBandAndTimedFootprints(t *testing.T) {
	start := time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC)
	path := samplePlanetOccultationPath(start)
	footprint := sampleFootprint(start.Add(time.Hour), 10, -10, 20, 10)
	path.PartialFootprints = []moon.PlanetOccultationFootprint{footprint}
	path.PartialBandFootprints = []moon.PlanetOccultationFootprint{footprint}
	data, err := geojson.MarshalPlanetOccultation(path)
	if err != nil {
		t.Fatalf("MarshalPlanetOccultation: %v", err)
	}
	collection := decodeCollection(t, data)
	if len(featuresWithRole(collection, "partial-band")) != 1 || len(featuresWithRole(collection, "partial-footprint")) != 1 {
		t.Fatal("compact band and timed footprint were not both serialized")
	}
}

func TestMarshalOccultationRejectsMalformedRiseSetCurves(t *testing.T) {
	start := time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC)
	curve := moon.OccultationRiseSetCurve{
		Phase:     moon.RiseSetPhaseStart,
		Direction: moon.RiseSetDirectionRise,
		Segments: [][]moon.OccultationPathPoint{{
			{Time: start.Add(20 * time.Minute), Longitude: 10, Latitude: 20},
			{Time: start.Add(40 * time.Minute), Longitude: math.NaN(), Latitude: 21},
		}},
	}
	starPath := sampleStarOccultationPath(start)
	starPath.RiseSetCurves = []moon.OccultationRiseSetCurve{curve}
	if _, err := geojson.MarshalStarOccultation(starPath); err == nil {
		t.Fatal("stellar rise/set curve with a NaN coordinate was accepted")
	}

	planetPath := samplePlanetOccultationPath(start)
	curve.Segments[0][1].Longitude = 12
	curve.Phase = moon.RiseSetPhase("bogus")
	planetPath.RiseSetCurves = []moon.OccultationRiseSetCurve{curve}
	if _, err := geojson.MarshalPlanetOccultation(planetPath); err == nil {
		t.Fatal("planetary rise/set curve with an invalid phase was accepted")
	}
}

func TestMarshalFunctionsRejectIncompleteInput(t *testing.T) {
	if _, err := geojson.MarshalSolarEclipse(eclipse.SolarEclipsePartialFootprintsInfo{}, nil); err == nil {
		t.Fatal("empty solar eclipse input was accepted")
	}
	if _, err := geojson.MarshalLunarEclipse(eclipse.LunarEclipseInfo{}, 360); err == nil {
		t.Fatal("empty lunar eclipse input was accepted")
	}
	if _, err := geojson.MarshalStarOccultation(moon.StarOccultationPath{}); err == nil {
		t.Fatal("empty stellar occultation input was accepted")
	}
	if _, err := geojson.MarshalPlanetOccultation(moon.PlanetOccultationPath{}); err == nil {
		t.Fatal("empty planetary occultation input was accepted")
	}
}

func TestTimeMarkerOptionsRejectNegativeStep(t *testing.T) {
	start := time.Date(2025, time.June, 5, 17, 45, 0, 0, time.UTC)
	center := occultationSamples(start, []float64{20, 30, 40}, []float64{2, 1, 0})
	north := occultationSamples(start, []float64{20, 30, 40}, []float64{12, 11, 10})
	south := occultationSamples(start, []float64{20, 30, 40}, []float64{-8, -9, -10})
	path := moon.StarOccultationPath{
		TargetID:      "HR 4799",
		Start:         north[0],
		Greatest:      center[1],
		End:           north[len(north)-1],
		Complete:      true,
		CenterLine:    center,
		NorthernLimit: north,
		SouthernLimit: south,
	}
	if _, err := geojson.MarshalStarOccultationWithTimeMarkers(path, geojson.TimeMarkerOptions{Step: -time.Minute}); err == nil {
		t.Fatal("negative time-marker step was accepted")
	}
	if _, err := geojson.MarshalStarOccultationWithTimeMarkers(path, geojson.TimeMarkerOptions{Step: time.Nanosecond}); err == nil {
		t.Fatal("sub-minute time-marker step was accepted")
	}

	excessiveEnd := path.CenterLine[0].Time.Add(24*time.Hour + 2*time.Minute)
	path.End.Time = excessiveEnd
	path.CenterLine[len(path.CenterLine)-1].Time = excessiveEnd
	path.NorthernLimit[len(path.NorthernLimit)-1].Time = excessiveEnd
	path.SouthernLimit[len(path.SouthernLimit)-1].Time = excessiveEnd
	if _, err := geojson.MarshalStarOccultationWithTimeMarkers(path, geojson.TimeMarkerOptions{Step: time.Minute}); err == nil {
		t.Fatal("excessive time-marker count was accepted")
	}
}

func TestTimeMarkerOptionsAreValidatedWithoutCenterLine(t *testing.T) {
	start := time.Date(2025, time.June, 5, 17, 45, 0, 0, time.UTC)
	north := occultationSamples(start, []float64{20, 30, 40}, []float64{12, 11, 10})
	south := occultationSamples(start, []float64{20, 30, 40}, []float64{-8, -9, -10})
	path := moon.StarOccultationPath{
		TargetID: "edge-only", Start: north[0], Greatest: north[1], End: north[2], Complete: true,
		NorthernLimit: north, SouthernLimit: south, Step: time.Hour,
	}
	if _, err := geojson.MarshalStarOccultationWithTimeMarkers(path, geojson.TimeMarkerOptions{Step: time.Nanosecond}); err == nil {
		t.Fatal("edge-only stellar path accepted sub-minute markers")
	}
	planet := moon.PlanetOccultationPath{
		Planet: moon.OccultationVenus, TargetID: "Venus", Start: north[0], Greatest: north[1], End: north[2],
		Complete: true, NorthernLimit: north, SouthernLimit: south, Step: time.Hour,
	}
	if _, err := geojson.MarshalPlanetOccultationWithTimeMarkers(planet, geojson.TimeMarkerOptions{Step: time.Nanosecond}); err == nil {
		t.Fatal("edge-only planetary path accepted sub-minute markers")
	}
}

func decodeCollection(t *testing.T, data []byte) decodedCollection {
	t.Helper()
	var collection decodedCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		t.Fatalf("decode GeoJSON: %v", err)
	}
	if collection.Type != "FeatureCollection" {
		t.Fatalf("collection type=%q, want FeatureCollection", collection.Type)
	}
	if len(collection.Features) == 0 {
		t.Fatal("GeoJSON contains no features")
	}
	for _, feature := range collection.Features {
		if feature.Type != "Feature" {
			t.Fatalf("feature type=%q, want Feature", feature.Type)
		}
		if _, ok := feature.Properties["event"]; !ok {
			t.Fatal("feature has no event property")
		}
		if _, ok := feature.Properties["role"]; !ok {
			t.Fatal("feature has no role property")
		}
	}
	return collection
}

func assertRoles(t *testing.T, collection decodedCollection, roles ...string) {
	t.Helper()
	for _, role := range roles {
		featureWithRole(t, collection, role)
	}
}

func featureWithRole(t *testing.T, collection decodedCollection, role string) decodedFeature {
	t.Helper()
	for _, feature := range collection.Features {
		if feature.Properties["role"] == role {
			return feature
		}
	}
	t.Fatalf("GeoJSON is missing role %q", role)
	return decodedFeature{}
}

func featuresWithRole(collection decodedCollection, role string) []decodedFeature {
	result := make([]decodedFeature, 0)
	for _, feature := range collection.Features {
		if feature.Properties["role"] == role {
			result = append(result, feature)
		}
	}
	return result
}

func assertCollectionCoordinates(t *testing.T, collection decodedCollection) {
	t.Helper()
	for _, feature := range collection.Features {
		var coordinates interface{}
		if err := json.Unmarshal(feature.Geometry.Coordinates, &coordinates); err != nil {
			t.Fatalf("decode %v coordinates: %v", feature.Properties["role"], err)
		}
		assertCoordinateTree(t, coordinates)
	}
}

func assertCoordinateTree(t *testing.T, value interface{}) {
	t.Helper()
	items, ok := value.([]interface{})
	if !ok {
		t.Fatalf("coordinate node has type %T", value)
	}
	if len(items) >= 2 {
		longitude, lonOK := items[0].(float64)
		latitude, latOK := items[1].(float64)
		if lonOK && latOK {
			if math.IsNaN(longitude) || math.IsInf(longitude, 0) || longitude < -180 || longitude > 180 {
				t.Fatalf("invalid longitude %.12f", longitude)
			}
			if math.IsNaN(latitude) || math.IsInf(latitude, 0) || latitude < -90 || latitude > 90 {
				t.Fatalf("invalid latitude %.12f", latitude)
			}
			return
		}
	}
	for _, item := range items {
		assertCoordinateTree(t, item)
	}
}

func assertClosedMultiPolygon(t *testing.T, feature decodedFeature) {
	t.Helper()
	if feature.Geometry.Type != "MultiPolygon" {
		t.Fatalf("%v geometry=%q, want MultiPolygon", feature.Properties["role"], feature.Geometry.Type)
	}
	var polygons [][][][]float64
	if err := json.Unmarshal(feature.Geometry.Coordinates, &polygons); err != nil {
		t.Fatalf("decode %v polygon: %v", feature.Properties["role"], err)
	}
	if len(polygons) == 0 {
		t.Fatalf("%v has no polygons", feature.Properties["role"])
	}
	for _, polygon := range polygons {
		if len(polygon) == 0 || len(polygon[0]) < 4 {
			t.Fatalf("%v contains an incomplete ring", feature.Properties["role"])
		}
		ring := polygon[0]
		first, last := ring[0], ring[len(ring)-1]
		if first[0] != last[0] || first[1] != last[1] {
			t.Fatalf("%v ring is not closed", feature.Properties["role"])
		}
	}
}

// geometryContainsPoint 判定点是否落在导出的几何内；解不出坐标时直接失败，避免 wantInside=false 的用例静默通过。
func geometryContainsPoint(t *testing.T, value struct {
	Type        string          `json:"type"`
	Coordinates json.RawMessage `json:"coordinates"`
	Geometries  json.RawMessage `json:"geometries"`
}, longitude, latitude float64) bool {
	t.Helper()
	switch value.Type {
	case "MultiPolygon":
		var polygons [][][][]float64
		if err := json.Unmarshal(value.Coordinates, &polygons); err != nil {
			t.Fatalf("decode MultiPolygon coordinates: %v", err)
		}
		return geoJSONMultiPolygonContains(polygons, longitude, latitude)
	case "Polygon":
		var polygon [][][]float64
		if err := json.Unmarshal(value.Coordinates, &polygon); err != nil {
			t.Fatalf("decode Polygon coordinates: %v", err)
		}
		if len(polygon) == 0 {
			return false
		}
		return geoJSONRingContains(polygon[0], longitude, latitude)
	case "GeometryCollection":
		var geometries []struct {
			Type        string          `json:"type"`
			Coordinates json.RawMessage `json:"coordinates"`
			Geometries  json.RawMessage `json:"geometries"`
		}
		if err := json.Unmarshal(value.Geometries, &geometries); err != nil {
			t.Fatalf("decode GeometryCollection: %v", err)
		}
		for _, geometry := range geometries {
			if geometryContainsPoint(t, geometry, longitude, latitude) {
				return true
			}
		}
	}
	return false
}

func assertTimedLineAligned(t *testing.T, feature decodedFeature) {
	t.Helper()
	var lines [][][]float64
	if err := json.Unmarshal(feature.Geometry.Coordinates, &lines); err != nil {
		t.Fatalf("decode line coordinates: %v", err)
	}
	timeSegments, ok := feature.Properties["times"].([]interface{})
	if !ok || len(timeSegments) != len(lines) {
		t.Fatalf("times do not align with %d line segments: %#v", len(lines), feature.Properties["times"])
	}
	for index, rawSegment := range timeSegments {
		times, ok := rawSegment.([]interface{})
		if !ok || len(times) != len(lines[index]) {
			t.Fatalf("times segment %d does not align with %d coordinates", index, len(lines[index]))
		}
		for _, rawTime := range times {
			value, ok := rawTime.(string)
			if !ok {
				t.Fatalf("time has type %T", rawTime)
			}
			if _, err := time.Parse(time.RFC3339Nano, value); err != nil {
				t.Fatalf("invalid RFC3339 time %q: %v", value, err)
			}
		}
	}
}

func assertRiseSetBoundaryFeatures(t *testing.T, collection decodedCollection, body string) {
	t.Helper()
	features := featuresWithRole(collection, "visibility-boundary")
	seen := make(map[string]map[string]bool, 2)
	phaseCurveCount := make(map[string]int, 2)
	for _, feature := range features {
		phase, phaseOK := feature.Properties["phase"].(string)
		horizon, horizonOK := feature.Properties["horizon"].(string)
		if !phaseOK || !horizonOK || feature.Properties["body"] != body {
			t.Fatalf("invalid %s visibility properties: %#v", body, feature.Properties)
		}
		if feature.Geometry.Type != "MultiLineString" {
			t.Fatalf("%s %s/%s geometry=%q, want MultiLineString", body, phase, horizon, feature.Geometry.Type)
		}
		assertTimedLineAligned(t, feature)
		if phase == "horizon" {
			continue
		}
		band := ""
		if value, ok := feature.Properties["band"].(string); ok {
			band = value
		}
		if seen[band] == nil {
			seen[band] = make(map[string]bool, 6)
		}
		phaseCurveCount[band]++
		seen[band][phase+"/"+horizon] = true
	}
	for band, count := range phaseCurveCount {
		if count != 6 {
			t.Fatalf("%s/%s phase visibility-boundary count=%d, want 6", body, band, count)
		}
		for _, phase := range []string{"start", "greatest", "end"} {
			for _, horizon := range []string{"rise", "set"} {
				if !seen[band][phase+"/"+horizon] {
					t.Fatalf("missing %s/%s visibility boundary %s/%s", body, band, phase, horizon)
				}
			}
		}
	}
}

type decodedRiseSetEndpoint struct {
	coordinate []float64
	time       string
	segment    int
}

func riseSetBoundaryFeature(
	t *testing.T,
	collection decodedCollection,
	phase, horizon string,
) decodedFeature {
	t.Helper()
	for _, feature := range featuresWithRole(collection, "visibility-boundary") {
		if feature.Properties["phase"] == phase && feature.Properties["horizon"] == horizon {
			return feature
		}
	}
	t.Fatalf("visibility-boundary %s/%s not found", phase, horizon)
	return decodedFeature{}
}

func riseSetFeatureEndpoints(t *testing.T, feature decodedFeature) []decodedRiseSetEndpoint {
	t.Helper()
	var lines [][][]float64
	if err := json.Unmarshal(feature.Geometry.Coordinates, &lines); err != nil {
		t.Fatalf("decode visibility-boundary coordinates: %v", err)
	}
	timeSegments, ok := feature.Properties["times"].([]interface{})
	if !ok || len(timeSegments) != len(lines) {
		t.Fatalf("visibility-boundary times do not align with %d segments", len(lines))
	}
	endpoints := make([]decodedRiseSetEndpoint, 0, len(lines)*2)
	for segmentIndex, line := range lines {
		times, ok := timeSegments[segmentIndex].([]interface{})
		if !ok || len(times) != len(line) || len(line) < 2 {
			t.Fatalf("visibility-boundary segment %d has %d coordinates and invalid times", segmentIndex, len(line))
		}
		for _, pointIndex := range []int{0, len(line) - 1} {
			value, ok := times[pointIndex].(string)
			if !ok {
				t.Fatalf("visibility-boundary segment %d time %d has type %T", segmentIndex, pointIndex, times[pointIndex])
			}
			endpoints = append(endpoints, decodedRiseSetEndpoint{
				coordinate: line[pointIndex], time: value, segment: segmentIndex,
			})
		}
	}
	return endpoints
}

func riseSetFeaturesShareEndpointInRegion(
	t *testing.T,
	first, second decodedFeature,
	minimumLongitude, maximumLongitude, minimumLatitude, maximumLatitude float64,
) bool {
	t.Helper()
	firstEndpoints := riseSetFeatureEndpoints(t, first)
	secondEndpoints := riseSetFeatureEndpoints(t, second)
	for _, current := range firstEndpoints {
		if !riseSetEndpointInRegion(current, minimumLongitude, maximumLongitude, minimumLatitude, maximumLatitude) {
			continue
		}
		for _, other := range secondEndpoints {
			if riseSetEndpointsMatch(current, other) {
				return true
			}
		}
	}
	return false
}

func riseSetFeaturesShareVertexInRegion(
	t *testing.T,
	first, second decodedFeature,
	minimumLongitude, maximumLongitude, minimumLatitude, maximumLatitude float64,
) bool {
	t.Helper()
	firstVertices := riseSetFeatureVertices(t, first)
	secondVertices := riseSetFeatureVertices(t, second)
	for _, current := range firstVertices {
		if !riseSetEndpointInRegion(current, minimumLongitude, maximumLongitude, minimumLatitude, maximumLatitude) {
			continue
		}
		for _, other := range secondVertices {
			if riseSetEndpointsMatch(current, other) {
				return true
			}
		}
	}
	return false
}

func riseSetFeatureVertices(t *testing.T, feature decodedFeature) []decodedRiseSetEndpoint {
	t.Helper()
	var lines [][][]float64
	if err := json.Unmarshal(feature.Geometry.Coordinates, &lines); err != nil {
		t.Fatalf("decode visibility-boundary coordinates: %v", err)
	}
	timeSegments, ok := feature.Properties["times"].([]interface{})
	if !ok || len(timeSegments) != len(lines) {
		t.Fatalf("visibility-boundary times do not align with %d segments", len(lines))
	}
	vertices := make([]decodedRiseSetEndpoint, 0)
	for segmentIndex, line := range lines {
		times, ok := timeSegments[segmentIndex].([]interface{})
		if !ok || len(times) != len(line) {
			t.Fatalf("visibility-boundary segment %d times do not align with coordinates", segmentIndex)
		}
		for pointIndex, point := range line {
			value, ok := times[pointIndex].(string)
			if !ok {
				t.Fatalf("visibility-boundary segment %d time %d has type %T", segmentIndex, pointIndex, times[pointIndex])
			}
			vertices = append(vertices, decodedRiseSetEndpoint{
				coordinate: point, time: value, segment: segmentIndex,
			})
		}
	}
	return vertices
}

func riseSetFeatureSegmentsShareEndpointInRegion(
	t *testing.T,
	feature decodedFeature,
	minimumLongitude, maximumLongitude, minimumLatitude, maximumLatitude float64,
) bool {
	t.Helper()
	endpoints := riseSetFeatureEndpoints(t, feature)
	for index, current := range endpoints {
		if !riseSetEndpointInRegion(current, minimumLongitude, maximumLongitude, minimumLatitude, maximumLatitude) {
			continue
		}
		for _, other := range endpoints[index+1:] {
			if current.segment != other.segment && riseSetEndpointsMatch(current, other) {
				return true
			}
		}
	}
	return false
}

func riseSetFeatureHasInteriorVertexInRegion(
	t *testing.T,
	feature decodedFeature,
	minimumLongitude, maximumLongitude, minimumLatitude, maximumLatitude float64,
) bool {
	t.Helper()
	var lines [][][]float64
	if err := json.Unmarshal(feature.Geometry.Coordinates, &lines); err != nil {
		t.Fatalf("decode visibility-boundary coordinates: %v", err)
	}
	for _, line := range lines {
		for index := 1; index+1 < len(line); index++ {
			point := decodedRiseSetEndpoint{coordinate: line[index]}
			if riseSetEndpointInRegion(point, minimumLongitude, maximumLongitude, minimumLatitude, maximumLatitude) {
				return true
			}
		}
	}
	return false
}

func riseSetEndpointInRegion(
	point decodedRiseSetEndpoint,
	minimumLongitude, maximumLongitude, minimumLatitude, maximumLatitude float64,
) bool {
	return len(point.coordinate) >= 2 &&
		point.coordinate[0] >= minimumLongitude && point.coordinate[0] <= maximumLongitude &&
		point.coordinate[1] >= minimumLatitude && point.coordinate[1] <= maximumLatitude
}

func riseSetEndpointsMatch(first, second decodedRiseSetEndpoint) bool {
	return first.time == second.time &&
		math.Abs(first.coordinate[0]-second.coordinate[0]) <= 1e-9 &&
		math.Abs(first.coordinate[1]-second.coordinate[1]) <= 1e-9
}

func assertRiseSetFeatureHasNoInstantaneousBranchJump(t *testing.T, feature decodedFeature) {
	t.Helper()
	var lines [][][]float64
	if err := json.Unmarshal(feature.Geometry.Coordinates, &lines); err != nil {
		t.Fatalf("decode visibility-boundary coordinates: %v", err)
	}
	timeSegments, ok := feature.Properties["times"].([]interface{})
	if !ok || len(timeSegments) != len(lines) {
		t.Fatalf("visibility-boundary times do not align with %d segments", len(lines))
	}
	for segmentIndex, line := range lines {
		times, ok := timeSegments[segmentIndex].([]interface{})
		if !ok || len(times) != len(line) {
			t.Fatalf("visibility-boundary segment %d times do not align with coordinates", segmentIndex)
		}
		for pointIndex := 1; pointIndex < len(line); pointIndex++ {
			previous, previousOK := times[pointIndex-1].(string)
			current, currentOK := times[pointIndex].(string)
			previousTime, previousErr := time.Parse(time.RFC3339Nano, previous)
			currentTime, currentErr := time.Parse(time.RFC3339Nano, current)
			if !previousOK || !currentOK || previousErr != nil || currentErr != nil {
				t.Fatalf("visibility-boundary segment %d has invalid adjacent times", segmentIndex)
			}
			distance := geoJSONCoordinateDistanceKM(line[pointIndex-1], line[pointIndex])
			if distance > 500 && absoluteDuration(currentTime.Sub(previousTime)) < time.Second {
				t.Fatalf("visibility-boundary segment %d contains %.1f km jump in %s",
					segmentIndex, distance, currentTime.Sub(previousTime))
			}
		}
	}
}

func absoluteDuration(value time.Duration) time.Duration {
	if value < 0 {
		return -value
	}
	return value
}

func assertOccultationBandPolygonCount(t *testing.T, collection decodedCollection, role string, want int) {
	t.Helper()
	band := featureWithRole(t, collection, role)
	var polygons [][][][]float64
	if err := json.Unmarshal(band.Geometry.Coordinates, &polygons); err != nil {
		t.Fatalf("decode %s: %v", role, err)
	}
	if len(polygons) != want {
		t.Fatalf("%s polygon count = %d, want %d continuous bands", role, len(polygons), want)
	}
}

func geoJSONMultiPolygonContains(polygons [][][][]float64, longitude, latitude float64) bool {
	for _, polygon := range polygons {
		if len(polygon) > 0 && geoJSONRingContains(polygon[0], longitude, latitude) {
			return true
		}
	}
	return false
}

func geoJSONMultiPolygonBoundaryDistanceKM(polygons [][][][]float64, point []float64) float64 {
	minimum := math.Inf(1)
	for _, polygon := range polygons {
		for _, ring := range polygon {
			for index := 1; index < len(ring); index++ {
				minimum = math.Min(minimum, geoJSONPointSegmentDistanceKM(point, ring[index-1], ring[index]))
			}
		}
	}
	return minimum
}

func assertGeoJSONRingHasNoShortHairpins(
	t *testing.T,
	role string,
	ring [][]float64,
	maximumClosureKM, minimumDetourKM float64,
	maximumSpan int,
) {
	t.Helper()
	for start := 0; start+3 < len(ring); start++ {
		limit := start + maximumSpan
		if limit >= len(ring) {
			limit = len(ring) - 1
		}
		arcLength := 0.0
		for end := start + 1; end <= limit; end++ {
			arcLength += geoJSONCoordinateDistanceKM(ring[end-1], ring[end])
			if end < start+3 {
				continue
			}
			closure := geoJSONCoordinateDistanceKM(ring[start], ring[end])
			if closure <= maximumClosureKM && arcLength-closure >= minimumDetourKM {
				t.Fatalf(
					"%s ring has a short hairpin at points %d..%d: closure %.1f km, arc %.1f km",
					role, start, end, closure, arcLength,
				)
			}
		}
	}
}

func geoJSONRingContains(ring [][]float64, longitude, latitude float64) bool {
	inside := false
	for current, previous := 0, len(ring)-1; current < len(ring); previous, current = current, current+1 {
		a, b := ring[previous], ring[current]
		cross := (longitude-a[0])*(b[1]-a[1]) - (latitude-a[1])*(b[0]-a[0])
		if math.Abs(cross) <= 1e-9 &&
			longitude >= math.Min(a[0], b[0])-1e-9 && longitude <= math.Max(a[0], b[0])+1e-9 &&
			latitude >= math.Min(a[1], b[1])-1e-9 && latitude <= math.Max(a[1], b[1])+1e-9 {
			return true
		}
		if (a[1] > latitude) == (b[1] > latitude) {
			continue
		}
		intersection := a[0] + (latitude-a[1])*(b[0]-a[0])/(b[1]-a[1])
		if intersection > longitude {
			inside = !inside
		}
	}
	return inside
}

func geoJSONCoordinateDistanceKM(first, second []float64) float64 {
	firstLatitude := first[1] * math.Pi / 180
	secondLatitude := second[1] * math.Pi / 180
	deltaLatitude := secondLatitude - firstLatitude
	deltaLongitude := math.Remainder((second[0]-first[0])*math.Pi/180, 2*math.Pi)
	haversine := math.Sin(deltaLatitude/2)*math.Sin(deltaLatitude/2) +
		math.Cos(firstLatitude)*math.Cos(secondLatitude)*math.Sin(deltaLongitude/2)*math.Sin(deltaLongitude/2)
	return 2 * 6378.1366 * math.Asin(math.Sqrt(math.Min(1, haversine)))
}

func geoJSONPointSegmentDistanceKM(point, start, end []float64) float64 {
	latitude := point[1] * math.Pi / 180
	x := func(value []float64) float64 {
		return math.Remainder(value[0]-point[0], 360) * math.Cos(latitude) * math.Pi / 180 * 6378.1366
	}
	y := func(value []float64) float64 {
		return (value[1] - point[1]) * math.Pi / 180 * 6378.1366
	}
	startX, startY := x(start), y(start)
	endX, endY := x(end), y(end)
	deltaX, deltaY := endX-startX, endY-startY
	fraction := 0.0
	if lengthSquared := deltaX*deltaX + deltaY*deltaY; lengthSquared > 0 {
		fraction = math.Max(0, math.Min(1, -(startX*deltaX+startY*deltaY)/lengthSquared))
	}
	return math.Hypot(startX+fraction*deltaX, startY+fraction*deltaY)
}

func occultationSamples(start time.Time, longitudes, latitudes []float64) []moon.OccultationPathPoint {
	result := make([]moon.OccultationPathPoint, len(longitudes))
	for index := range result {
		result[index] = moon.OccultationPathPoint{
			Time:         start.Add(time.Duration(index) * time.Hour),
			Longitude:    longitudes[index],
			Latitude:     latitudes[index],
			MoonAltitude: 30,
			WidthKM:      3000,
		}
	}
	return result
}

func sampleStarOccultationPath(start time.Time) moon.StarOccultationPath {
	center := occultationSamples(start, []float64{20, 30, 40}, []float64{2, 1, 0})
	north := occultationSamples(start, []float64{20, 30, 40}, []float64{12, 11, 10})
	south := occultationSamples(start, []float64{20, 30, 40}, []float64{-8, -9, -10})
	return moon.StarOccultationPath{
		TargetID: "HR 4799", Start: north[0], Greatest: center[1], End: north[len(north)-1],
		Complete: true, CenterLine: center, NorthernLimit: north, SouthernLimit: south, Step: time.Hour,
	}
}

func samplePlanetOccultationPath(start time.Time) moon.PlanetOccultationPath {
	star := sampleStarOccultationPath(start)
	return moon.PlanetOccultationPath{
		Planet: moon.OccultationSaturn, TargetID: "Saturn",
		Start: star.Start, Greatest: star.Greatest, End: star.End, Complete: true,
		CenterLine: star.CenterLine, NorthernLimit: star.NorthernLimit, SouthernLimit: star.SouthernLimit,
		Step: time.Hour,
	}
}

func sampleFootprint(at time.Time, west, south, east, north float64) moon.PlanetOccultationFootprint {
	return moon.PlanetOccultationFootprint{
		Time: at,
		Polygons: [][]moon.OccultationPathPoint{{
			{Time: at, Longitude: west, Latitude: south, MoonAltitude: 30, WidthKM: 1000},
			{Time: at, Longitude: east, Latitude: south, MoonAltitude: 30, WidthKM: 1000},
			{Time: at, Longitude: east, Latitude: north, MoonAltitude: 30, WidthKM: 1000},
			{Time: at, Longitude: west, Latitude: north, MoonAltitude: 30, WidthKM: 1000},
		}},
	}
}

// UT1 导出：时间字符串写 UT1 时刻、集合上带 time_scale 成员，且拒绝非 UTC 时区。
func TestMarshalSolarEclipseUT1TimeScale(t *testing.T) {
	date := time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipse.SolarEclipsePartialFootprintsNASABulletinSplitK(date, eclipse.SolarEclipsePartialFootprintOptions{
		Step: 10 * time.Minute, BoundaryPoints: 24,
	})
	if !ok {
		t.Fatal("缺少日食足迹")
	}
	utcData, err := geojson.MarshalSolarEclipse(partial, nil)
	if err != nil {
		t.Fatalf("UTC 导出失败: %v", err)
	}
	if strings.Contains(string(utcData), "time_scale") {
		t.Error("UTC 导出不应带 time_scale 成员")
	}
	ut1Data, err := geojson.MarshalSolarEclipseWithTimeMarkers(partial, nil, geojson.TimeMarkerOptions{TimeScale: astro.TimeScaleUT1})
	if err != nil {
		t.Fatalf("UT1 导出失败: %v", err)
	}
	if !strings.Contains(string(ut1Data), `"time_scale":"UT1"`) {
		t.Errorf("UT1 导出应带 time_scale 成员: %s", string(ut1Data)[:200])
	}
	cst := time.FixedZone("CST", 8*3600)
	if _, err := geojson.MarshalSolarEclipseWithTimeMarkers(partial, nil, geojson.TimeMarkerOptions{TimeScale: astro.TimeScaleUT1, Location: cst}); err == nil {
		t.Error("UT1 配非 UTC 时区应报错")
	}
}
