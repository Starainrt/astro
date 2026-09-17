package geojson_test

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/internal/geodata"
	"github.com/starainrt/astro/internal/occultationgeo"
	"github.com/starainrt/astro/moon"
)

func TestMarshalStarOccultationAntares20240303UsesContinuousVisibleEnvelope(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2024, time.March, 3, 0, 0, 0, 0, zone)
	star := moon.StarCoordinate{
		ID: "HR 6134", RA: 247.35166667, Dec: -26.43194444,
		Epoch:                          time.Date(2000, time.January, 1, 12, 0, 0, 0, time.UTC),
		Frame:                          moon.CoordinateFrameJ2000,
		ProperMotionRACosDecMasPerYear: -10,
		ProperMotionDecMasPerYear:      -20,
		ParallaxMas:                    24,
	}
	paths, err := moon.FindStarOccultationPaths(
		start, start.Add(24*time.Hour), star,
		moon.OccultationPathOptions{
			Step: 20 * time.Minute, TargetSpacingKM: 900, RiseSetStep: time.Minute,
			DisableFootprints: true, IncludeFootprintTimeline: true,
			FootprintTimelineStep: 5 * time.Minute,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindStarOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	data, err := geojson.MarshalStarOccultation(paths[0])
	if err != nil {
		t.Fatalf("MarshalStarOccultation: %v", err)
	}
	collection := decodeCollection(t, data)
	band := featureWithRole(t, collection, "occultation-band")
	if band.Properties["static_band_authoritative"] != true {
		t.Fatalf("occultation-band source=%v authoritative=%v, want analytic authoritative boundary",
			band.Properties["source"], band.Properties["static_band_authoritative"])
	}
	for index, point := range paths[0].CenterLine {
		if !geometryContainsPoint(t, band.Geometry, point.Longitude, point.Latitude) {
			t.Fatalf("center-line sample %d lies outside visible band at %.6f, %.6f",
				index, point.Longitude, point.Latitude)
		}
	}
	for _, feature := range featuresWithRole(collection, "visibility-boundary") {
		var lines [][][]float64
		if err := json.Unmarshal(feature.Geometry.Coordinates, &lines); err != nil {
			t.Fatalf("decode visibility-boundary: %v", err)
		}
		for segmentIndex, line := range lines {
			for pointIndex, point := range line {
				if !geometryContainsPointWithinKM(t, band.Geometry, point[0], point[1], 2) {
					t.Fatalf("visibility-boundary segment %d point %d lies outside visible band at %.6f, %.6f",
						segmentIndex, pointIndex, point[0], point[1])
				}
			}
		}
	}
	assertOccultationBandMaximumEdge(t, band, 45)
}

func TestMarshalStarOccultationAntares20240303RetainsNarrowGreatestSetFold(t *testing.T) {
	start := time.Date(2024, time.March, 3, 0, 0, 0, 0, time.UTC)
	paths, err := moon.FindStarOccultationPaths(
		start, start.Add(24*time.Hour), antaresCoordinateForGeoJSONRegression(),
		moon.OccultationPathOptions{
			Step: 20 * time.Minute, TargetSpacingKM: 900, RiseSetStep: time.Minute,
			DisableFootprints: true, IncludeFootprintTimeline: true,
			FootprintTimelineStep: 5 * time.Minute,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindStarOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	var curve moon.OccultationRiseSetCurve
	for _, candidate := range paths[0].RiseSetCurves {
		if candidate.Phase == moon.RiseSetPhaseGreatest && candidate.Direction == moon.RiseSetDirectionSet {
			curve = candidate
			break
		}
	}
	if len(curve.Segments) < 2 {
		t.Fatalf("greatest/set segments=%d, want the narrow fold branch recovered at one-minute sampling", len(curve.Segments))
	}
	shared := false
	for first := 0; first < len(curve.Segments); first++ {
		for second := first + 1; second < len(curve.Segments); second++ {
			for _, left := range []moon.OccultationPathPoint{curve.Segments[first][0], curve.Segments[first][len(curve.Segments[first])-1]} {
				for _, right := range []moon.OccultationPathPoint{curve.Segments[second][0], curve.Segments[second][len(curve.Segments[second])-1]} {
					if left.Time.Sub(right.Time) < -time.Second || left.Time.Sub(right.Time) > time.Second {
						continue
					}
					if geojsonPointDistanceKM(left.Longitude, left.Latitude, right.Longitude, right.Latitude) <= 1 {
						shared = true
					}
				}
			}
		}
	}
	if !shared {
		t.Fatal("greatest/set fold branches do not share a physical endpoint")
	}
	data, err := geojson.MarshalStarOccultation(paths[0])
	if err != nil {
		t.Fatalf("MarshalStarOccultation: %v", err)
	}
	boundary := riseSetBoundaryFeature(t, decodeCollection(t, data), "greatest", "set")
	var lines [][][]float64
	if err := json.Unmarshal(boundary.Geometry.Coordinates, &lines); err != nil {
		t.Fatalf("decode greatest/set boundary: %v", err)
	}
	if len(lines) < 2 {
		t.Fatalf("serialized greatest/set segments=%d, want at least two folded branches", len(lines))
	}
}

func geojsonPointDistanceKM(firstLongitude, firstLatitude, secondLongitude, secondLatitude float64) float64 {
	const earthRadiusKM = 6378.1366
	const degreesToRadians = 3.141592653589793 / 180
	firstLat, secondLat := firstLatitude*degreesToRadians, secondLatitude*degreesToRadians
	deltaLat := (secondLatitude - firstLatitude) * degreesToRadians
	deltaLon := (secondLongitude - firstLongitude) * degreesToRadians
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) + math.Cos(firstLat)*math.Cos(secondLat)*math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	return 2 * earthRadiusKM * math.Asin(math.Sqrt(math.Max(0, math.Min(1, a))))
}

func TestMarshalStarOccultationAntaresRepresentative2022To2026Topology(t *testing.T) {
	star := antaresCoordinateForGeoJSONRegression()
	for _, date := range []string{
		"2023-09-21", // first event in the series; short temporal horizon closure
		"2023-10-18", // complete endpoint network without a temporal connector
		"2024-03-03", // original disconnected footprint-sweep regression
		"2024-06-20", // antimeridian split
		"2025-08-31", // multi-branch southern polar turn and numerical sliver
		"2026-02-11", // pole-enclosing equirectangular output
		"2026-12-08", // returning ordinary-latitude branch
	} {
		date := date
		t.Run(date, func(t *testing.T) {
			start, err := time.Parse("2006-01-02", date)
			if err != nil {
				t.Fatal(err)
			}
			paths, err := moon.FindStarOccultationPaths(
				start, start.Add(24*time.Hour), star,
				moon.OccultationPathOptions{
					Step: 20 * time.Minute, TargetSpacingKM: 900, RiseSetStep: time.Minute,
					DisableFootprints: true, IncludeFootprintTimeline: true,
					FootprintTimelineStep: 5 * time.Minute,
				},
			)
			if err != nil || len(paths) != 1 {
				t.Fatalf("FindStarOccultationPaths() paths=%d err=%v, want one", len(paths), err)
			}
			path := paths[0]
			polygons, authoritative, err := occultationgeo.VisibleStarBandPolygonsFromAnalyticContours(
				path.BandFootprints, path.BandContours, path.VisibilityContours,
				path.NorthernLimit, path.SouthernLimit, path.RiseSetCurves,
			)
			if err != nil || !authoritative || len(polygons) != 1 {
				t.Fatalf("analytic polygons=%d authoritative=%v err=%v, want one physical band", len(polygons), authoritative, err)
			}
			assertOccultationPhysicalPolygonsMaximumEdge(t, polygons, 45)
			assertOccultationPhysicalPolygonsHaveNoShortHairpins(t, polygons, 35, 25, 12)
			if !geodata.SphericalPolygonsContainPathsWithinKM(
				polygons, occultationPathPointLines([][]moon.OccultationPathPoint{path.CenterLine}), false, 0.1,
			) {
				t.Fatalf("analytic band misses center line by %.3f km",
					geodata.SphericalPolygonsPathMissDistanceKM(
						polygons, occultationPathPointLines([][]moon.OccultationPathPoint{path.CenterLine}), false,
					))
			}
			phaseLines := make([][]geodata.GeoPoint, 0, len(path.RiseSetCurves))
			for _, curve := range path.RiseSetCurves {
				phaseLines = append(phaseLines,
					occultationPathPointLines(occultationgeo.StitchedRiseSetCurveSegments(curve))...,
				)
			}
			if !geodata.SphericalPolygonsContainPathsWithinKM(polygons, phaseLines, false, 2) {
				t.Fatalf("analytic band misses a displayed rise/set phase by %.3f km",
					geodata.SphericalPolygonsPathMissDistanceKM(polygons, phaseLines, false))
			}
			data, err := geojson.MarshalStarOccultation(path)
			if err != nil {
				t.Fatalf("MarshalStarOccultation: %v", err)
			}
			collection := decodeCollection(t, data)
			band := featureWithRole(t, collection, "occultation-band")
			if band.Properties["static_band_authoritative"] != true {
				t.Fatalf("occultation-band source=%v authoritative=%v",
					band.Properties["source"], band.Properties["static_band_authoritative"])
			}
			for index, point := range path.CenterLine {
				if !geometryContainsPointWithinKM(t, band.Geometry, point.Longitude, point.Latitude, 0.1) {
					t.Fatalf("GeoJSON band excludes center sample %d at %.6f, %.6f",
						index, point.Longitude, point.Latitude)
				}
			}
		})
	}
}

func antaresCoordinateForGeoJSONRegression() moon.StarCoordinate {
	return moon.StarCoordinate{
		ID: "HR 6134", RA: 247.35166667, Dec: -26.43194444,
		Epoch:                          time.Date(2000, time.January, 1, 12, 0, 0, 0, time.UTC),
		Frame:                          moon.CoordinateFrameJ2000,
		ProperMotionRACosDecMasPerYear: -10,
		ProperMotionDecMasPerYear:      -20,
		ParallaxMas:                    24,
	}
}

func occultationPathPointLines(sources [][]moon.OccultationPathPoint) [][]geodata.GeoPoint {
	result := make([][]geodata.GeoPoint, 0, len(sources))
	for _, source := range sources {
		if len(source) < 2 {
			continue
		}
		line := make([]geodata.GeoPoint, len(source))
		for index, point := range source {
			line[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
		}
		result = append(result, line)
	}
	return result
}

func assertOccultationPhysicalPolygonsMaximumEdge(
	t *testing.T,
	polygons [][]geodata.GeoPoint,
	maximumKM float64,
) {
	t.Helper()
	for polygonIndex, polygon := range polygons {
		for index := range polygon {
			next := (index + 1) % len(polygon)
			distance := geoJSONCoordinateDistanceKM(
				[]float64{polygon[index].Longitude, polygon[index].Latitude},
				[]float64{polygon[next].Longitude, polygon[next].Latitude},
			)
			if distance > maximumKM {
				t.Fatalf("physical polygon %d edge %d is %.1f km, want <=%.1f km",
					polygonIndex, index, distance, maximumKM)
			}
		}
	}
}

func assertOccultationPhysicalPolygonsHaveNoShortHairpins(
	t *testing.T,
	polygons [][]geodata.GeoPoint,
	maximumClosureKM, minimumDetourKM float64,
	maximumSpan int,
) {
	t.Helper()
	for _, polygon := range polygons {
		ring := make([][]float64, len(polygon))
		for index, point := range polygon {
			ring[index] = []float64{point.Longitude, point.Latitude}
		}
		assertGeoJSONRingHasNoShortHairpins(
			t, "physical occultation-band", ring,
			maximumClosureKM, minimumDetourKM, maximumSpan,
		)
	}
}

func geometryContainsPointWithinKM(t *testing.T, value struct {
	Type        string          `json:"type"`
	Coordinates json.RawMessage `json:"coordinates"`
	Geometries  json.RawMessage `json:"geometries"`
}, longitude, latitude, toleranceKM float64) bool {
	if geometryContainsPoint(t, value, longitude, latitude) {
		return true
	}
	var polygons [][][][]float64
	if value.Type == "MultiPolygon" {
		if json.Unmarshal(value.Coordinates, &polygons) != nil {
			return false
		}
	} else if value.Type == "Polygon" {
		var polygon [][][]float64
		if json.Unmarshal(value.Coordinates, &polygon) != nil {
			return false
		}
		polygons = [][][][]float64{polygon}
	} else {
		return false
	}
	point := []float64{longitude, latitude}
	for _, polygon := range polygons {
		for _, ring := range polygon {
			for index := 1; index < len(ring); index++ {
				if geoJSONPointSegmentDistanceKM(point, ring[index-1], ring[index]) <= toleranceKM {
					return true
				}
			}
		}
	}
	return false
}

func assertOccultationBandMaximumEdge(t *testing.T, feature decodedFeature, maximumKM float64) {
	t.Helper()
	var polygons [][][][]float64
	if err := json.Unmarshal(feature.Geometry.Coordinates, &polygons); err != nil {
		t.Fatalf("decode %v polygons: %v", feature.Properties["role"], err)
	}
	for polygonIndex, polygon := range polygons {
		for ringIndex, ring := range polygon {
			for pointIndex := 1; pointIndex < len(ring); pointIndex++ {
				if distance := geoJSONCoordinateDistanceKM(ring[pointIndex-1], ring[pointIndex]); distance > maximumKM {
					t.Fatalf("%v polygon %d ring %d edge %d is %.1f km, want <= %.1f km",
						feature.Properties["role"], polygonIndex, ringIndex, pointIndex, distance, maximumKM)
				}
			}
		}
	}
}
