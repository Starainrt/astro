package geojson_test

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/internal/geodata"
	"github.com/starainrt/astro/internal/occultationgeo"
	"github.com/starainrt/astro/moon"
)

// TestSolarEclipseP2SarosGeoJSONSamples serializes real eclipse paths at
// evenly spaced Saros-family offsets around the current epoch. It is kept
// intentionally smaller than the opt-in millennium diagnostic while still
// exercising antimeridian, polar and non-central topology in normal tests.
func TestSolarEclipseP2SarosGeoJSONSamples(t *testing.T) {
	const sarosDays = 6585.321314
	seed := basic.JDCalc(2024, 4, 8)
	for _, familyIndex := range []int{-28, -21, -14, -7, 0, 7, 14, 21, 28} {
		familyIndex := familyIndex
		t.Run("saros-"+formatP2SignedIndex(familyIndex), func(t *testing.T) {
			date := basic.JD2DateByZone(seed+float64(familyIndex)*sarosDays, time.UTC, false)
			partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
				Step: 10 * time.Minute, BoundaryPoints: 96, CentralShadowStep: 5 * time.Minute,
				DisableRiseSet: true,
			})
			if !ok || !partial.Eclipse.HasPartial {
				t.Fatalf("solar eclipse unavailable at %s: ok=%v type=%s", date.Format("2006-01-02"), ok, partial.Eclipse.Type)
			}
			data, err := geojson.MarshalSolarEclipse(partial, nil)
			if err != nil {
				t.Fatalf("MarshalSolarEclipse(%s): %v", date.Format("2006-01-02"), err)
			}
			collection := decodeCollection(t, data)
			assertCollectionCoordinates(t, collection)
			assertClosedMultiPolygon(t, featureWithRole(t, collection, "partial-band"))
			if partial.Eclipse.Type == eclipse.SolarEclipsePartial {
				if len(featuresWithRole(collection, "central-band")) != 0 {
					t.Fatal("partial-only eclipse unexpectedly contains a central band")
				}
			} else {
				assertClosedMultiPolygon(t, featureWithRole(t, collection, "central-band"))
			}
			if len(featuresWithRole(collection, "greatest")) != 1 {
				t.Fatalf("greatest feature count=%d, want one", len(featuresWithRole(collection, "greatest")))
			}
		})
	}
}

func TestOccultationP2RepresentativeGeoJSONPathsRemainClosed(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	star := moon.StarCoordinate{
		ID: "HR 4799", RA: 189.1975, Dec: -5.831944444444,
		Epoch:                          time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
		Frame:                          moon.CoordinateFrameJ2000,
		ProperMotionRACosDecMasPerYear: -28, ProperMotionDecMasPerYear: -18,
	}
	starCases := []struct {
		name  string
		start time.Time
	}{
		{name: "hr4799-2025", start: time.Date(2025, time.June, 5, 0, 0, 0, 0, zone)},
	}
	for _, test := range starCases {
		t.Run(test.name, func(t *testing.T) {
			paths, err := moon.FindStarOccultationPaths(test.start, test.start.Add(24*time.Hour), star,
				moon.OccultationPathOptions{Step: 20 * time.Minute, TargetSpacingKM: 900, DisableFootprints: true})
			if err != nil || len(paths) != 1 {
				t.Fatalf("FindStarOccultationPaths() paths=%d err=%v, want one", len(paths), err)
			}
			data, err := geojson.MarshalStarOccultation(paths[0])
			if err != nil {
				t.Fatalf("MarshalStarOccultation: %v", err)
			}
			collection := decodeCollection(t, data)
			assertCollectionCoordinates(t, collection)
			assertClosedMultiPolygon(t, featureWithRole(t, collection, "occultation-band"))
			assertClosedMultiLineFeature(t, featureWithRole(t, collection, "band-outline"))
			assertRoles(t, collection, "occultation-band", "band-outline", "visibility-boundary", "center-line", "north-limit", "south-limit")
		})
	}

	planetCases := []struct {
		name   string
		start  time.Time
		planet moon.OccultationPlanet
	}{
		{name: "venus-2025-09-19", start: time.Date(2025, time.September, 19, 0, 0, 0, 0, zone), planet: moon.OccultationVenus},
		{name: "saturn-2025-01-05", start: time.Date(2025, time.January, 5, 0, 0, 0, 0, zone), planet: moon.OccultationSaturn},
		{name: "mars-2025-06-30", start: time.Date(2025, time.June, 30, 0, 0, 0, 0, zone), planet: moon.OccultationMars},
	}
	for _, test := range planetCases {
		t.Run(test.name, func(t *testing.T) {
			paths, err := moon.FindPlanetOccultationPaths(test.start, test.start.Add(24*time.Hour), test.planet,
				moon.OccultationPathOptions{Step: 20 * time.Minute, TargetSpacingKM: 900, DisableFootprints: true})
			if err != nil || len(paths) != 1 {
				t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
			}
			data, err := geojson.MarshalPlanetOccultation(paths[0])
			if err != nil {
				t.Fatalf("MarshalPlanetOccultation: %v", err)
			}
			collection := decodeCollection(t, data)
			assertCollectionCoordinates(t, collection)
			assertClosedMultiPolygon(t, featureWithRole(t, collection, "partial-band"))
			assertClosedMultiLineFeature(t, featureWithRole(t, collection, "band-outline"))
			assertOccultationP0ContactContourSource(t, featureWithRole(t, collection, "partial-band"))
			if paths[0].HasTotalBand {
				assertClosedMultiPolygon(t, featureWithRole(t, collection, "total-band"))
				assertClosedMultiLineFeature(t, featureWithRole(t, collection, "total-band-outline"))
				assertOccultationP0ContactContourSource(t, featureWithRole(t, collection, "total-band"))
			}
			assertRoles(t, collection, "partial-band", "band-outline", "visibility-boundary", "center-line", "north-limit", "south-limit")
		})
	}
}

func TestOccultationP2Venus20250919HasNoInternalOutline(t *testing.T) {
	start := time.Date(2025, time.September, 19, 0, 0, 0, 0, time.FixedZone("UTC+8", 8*3600))
	for _, algorithm := range []moon.OccultationPathAlgorithm{moon.OccultationPathAlgorithmOptimized, moon.OccultationPathAlgorithmExact} {
		t.Run(string(algorithm), func(t *testing.T) {
			paths, err := moon.FindPlanetOccultationPaths(start, start.Add(24*time.Hour), moon.OccultationVenus,
				moon.OccultationPathOptions{
					Algorithm: algorithm, Step: 20 * time.Minute, TargetSpacingKM: 900,
					DisableFootprints: true, RiseSetStep: time.Minute,
				})
			if err != nil || len(paths) != 1 {
				t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
			}
			path := paths[0]
			data, err := geojson.MarshalPlanetOccultation(path)
			if err != nil {
				t.Fatalf("MarshalPlanetOccultation: %v", err)
			}
			collection := decodeCollection(t, data)
			assertCollectionCoordinates(t, collection)
			for _, role := range []string{"partial-band", "total-band"} {
				feature := featureWithRole(t, collection, role)
				assertClosedMultiPolygon(t, feature)
				assertOccultationP0ContactContourSource(t, feature)
				rings := geoJSONMultiPolygonOuterRings(t, feature)
				if len(rings) != 1 {
					t.Errorf("%s rings=%d, want one continuous band without overlapping slivers", role, len(rings))
				}
				contours, curves, footprints := path.PartialBandContours, path.RiseSetCurves, path.PartialBandFootprints
				if role == "total-band" {
					contours, curves, footprints = path.TotalBandContours, path.TotalRiseSetCurves, path.TotalBandFootprints
				}
				lines := occultationP0AuthoritativeBoundaryLines(contours, curves, footprints)
				for _, ring := range rings {
					for _, point := range ring {
						// Test the physical southern boundary, not the map's pole/dateline closure.
						if point.Latitude < 10 && point.Longitude > 25 && point.Longitude < 40 {
							if miss := geoPointLineDistanceKM(point, lines); miss > 1 {
								t.Errorf("%s southern outline leaves source by %.3f km at %+v", role, miss, point)
							}
						}
					}
				}
			}
			assertOccultationP2PhaseCurvesInsidePartialBand(t, collection, 1)
			partial := geoJSONMultiPolygonOuterRings(t, featureWithRole(t, collection, "partial-band"))
			total := geoJSONMultiPolygonOuterRings(t, featureWithRole(t, collection, "total-band"))
			if miss := geodata.SphericalPolygonsPathMissDistanceKM(partial, total, true); miss > 1 {
				t.Fatalf("total band extends %.3f km beyond partial band", miss)
			}
		})
	}
}

func TestOccultationP2Mars20250114GeoJSONAcceptsFoldedBandContours(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2025, time.January, 14, 0, 0, 0, 0, zone)
	paths, err := moon.FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), moon.OccultationMars,
		moon.OccultationPathOptions{
			Step: 20 * time.Minute, TargetSpacingKM: 900,
			DisableFootprints: true, IncludeFootprintTimeline: true,
			FootprintTimelineStep: 5 * time.Minute, RiseSetStep: time.Minute,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	data, err := geojson.MarshalPlanetOccultationWithTimeMarkers(
		paths[0], geojson.TimeMarkerOptions{Step: 30 * time.Minute, Location: zone},
	)
	if err != nil {
		t.Fatalf("MarshalPlanetOccultationWithTimeMarkers: %v", err)
	}
	collection := decodeCollection(t, data)
	assertClosedMultiPolygon(t, featureWithRole(t, collection, "partial-band"))
	assertClosedMultiLineFeature(t, featureWithRole(t, collection, "band-outline"))
}

func TestOccultationP2Mars20250729AddsMoonriseHorizonConnector(t *testing.T) {
	collection := mars20250729TestFixture(t, 5*time.Minute, true).collection
	assertOccultationHorizonConnectorsAreAuxiliary(t, collection)
	connector := horizonConnectorFeature(t, collection, "rise")
	var lines [][][]float64
	if err := json.Unmarshal(connector.Geometry.Coordinates, &lines); err != nil {
		t.Fatalf("decode horizon connector coordinates: %v", err)
	}
	if len(lines) == 0 || len(lines) > 3 {
		t.Fatalf("horizon connector segment count=%d, want one to three selected transition segments", len(lines))
	}
	phaseEndpoints := mars20250729PhaseEndpoints(t, collection, "rise")
	foundOpeningConnector := false
	for _, line := range lines {
		if len(line) < 2 {
			continue
		}
		first, last := line[0], line[len(line)-1]
		chord := geoJSONCoordinateDistanceKM(first, last)
		if chord < 750 || chord > 1000 {
			continue
		}
		total := 0.0
		maximumStep := 0.0
		for index := 1; index < len(line); index++ {
			step := geoJSONCoordinateDistanceKM(line[index-1], line[index])
			total += step
			if step > maximumStep {
				maximumStep = step
			}
		}
		if total > 1.25*chord || maximumStep > 200 {
			t.Fatalf("moonrise horizon connector detours along the contact limb: total=%.1f chord=%.1f max-step=%.1f km", total, chord, maximumStep)
		}
		if !geoJSONCoordinateMatchesAny(line[0], phaseEndpoints) ||
			!geoJSONCoordinateMatchesAny(line[len(line)-1], phaseEndpoints) {
			t.Fatalf("moonrise horizon connector endpoints do not meet physical phase endpoints: first=%v last=%v",
				line[0], line[len(line)-1])
		}
		foundOpeningConnector = true
		break
	}
	if !foundOpeningConnector {
		t.Fatal("2025-07-29 Mars occultation is missing the moonrise horizon connector in the South Pacific opening region")
	}
}

func TestOccultationP2Mars20250729OneMinuteRiseSetAddsOpeningConnector(t *testing.T) {
	collection := mars20250729TestFixture(t, time.Minute, true).collection
	connector := horizonConnectorFeature(t, collection, "rise")
	var lines [][][]float64
	if err := json.Unmarshal(connector.Geometry.Coordinates, &lines); err != nil {
		t.Fatalf("decode horizon connector coordinates: %v", err)
	}
	startRise := riseSetBoundaryFeature(t, collection, "start", "rise")
	startRiseEndpoints := riseSetFeatureEndpoints(t, startRise)
	foundOpeningConnector := false
	for _, line := range lines {
		if len(line) < 2 {
			continue
		}
		total := 0.0
		for pointIndex, point := range line {
			if pointIndex > 0 {
				total += geoJSONCoordinateDistanceKM(line[pointIndex-1], point)
			}
		}
		chord := geoJSONCoordinateDistanceKM(line[0], line[len(line)-1])
		if chord < 250 || chord > 650 || total < 400 || total > 700 {
			continue
		}
		maximumStep := 0.0
		for index := 1; index < len(line); index++ {
			step := geoJSONCoordinateDistanceKM(line[index-1], line[index])
			if step > maximumStep {
				maximumStep = step
			}
		}
		if maximumStep > 120 {
			t.Fatalf("one-minute moonrise opening connector max-step=%.1f km, want smooth closure", maximumStep)
		}
		if !geoJSONCoordinateMatchesAny(line[0], startRiseEndpoints) ||
			!geoJSONCoordinateMatchesAny(line[len(line)-1], startRiseEndpoints) {
			t.Fatalf("one-minute moonrise opening connector endpoints do not meet start/rise phase endpoints: first=%v last=%v",
				line[0], line[len(line)-1])
		}
		foundOpeningConnector = true
		break
	}
	if !foundOpeningConnector {
		t.Fatal("2025-07-29 Mars one-minute rise/set GeoJSON is missing the opening moonrise connector")
	}
}

func TestOccultationP2Mars20250729DoesNotDuplicateSamePhaseFinalFold(t *testing.T) {
	collection := mars20250729TestFixture(t, 5*time.Minute, true).collection
	connector := horizonConnectorFeature(t, collection, "rise")
	var lines [][][]float64
	if err := json.Unmarshal(connector.Geometry.Coordinates, &lines); err != nil {
		t.Fatalf("decode horizon connector coordinates: %v", err)
	}
	if len(lines) != 1 {
		t.Fatalf("moonrise horizon connector segment count=%d, want only the opening closure; same-phase final fold must be part of end/rise",
			len(lines))
	}
}

func TestOccultationP2Mars20250729StaticBandsStayAuthoritativeAndBounded(t *testing.T) {
	collection := mars20250729TestFixture(t, time.Minute, true).collection
	for _, role := range []string{"partial-band", "total-band"} {
		band := featureWithRole(t, collection, role)
		if authoritative, ok := band.Properties["static_band_authoritative"].(bool); !ok || !authoritative {
			t.Fatalf("%s is not authoritative: properties=%v", role, band.Properties)
		}
		outlineRole := "band-outline"
		if role == "total-band" {
			outlineRole = "total-band-outline"
		}
		outline := featureWithRole(t, collection, outlineRole)
		var lines [][][]float64
		if err := json.Unmarshal(outline.Geometry.Coordinates, &lines); err != nil {
			t.Fatalf("decode %s: %v", outlineRole, err)
		}
		for lineIndex, line := range lines {
			if len(line) < 4 {
				t.Fatalf("%s line %d has %d points", outlineRole, lineIndex, len(line))
			}
			maximum := 0.0
			for pointIndex := 1; pointIndex < len(line); pointIndex++ {
				maximum = math.Max(maximum, geoJSONCoordinateDistanceKM(line[pointIndex-1], line[pointIndex]))
			}
			if maximum > 80 {
				t.Fatalf("%s line %d has an artificial long edge %.1f km", outlineRole, lineIndex, maximum)
			}
		}
	}
}

func mustMarshalPlanetOccultation(t *testing.T, path moon.PlanetOccultationPath) []byte {
	t.Helper()
	data, err := geojson.MarshalPlanetOccultation(path)
	if err != nil {
		t.Fatalf("MarshalPlanetOccultation: %v", err)
	}
	return data
}

func TestOccultationP2Mars20250729DrawsPhaseCurvesAboveStaticBandOutlines(t *testing.T) {
	collection := mars20250729TestFixture(t, time.Minute, true).collection
	lastStaticOutline := -1
	firstPhaseCurve := len(collection.Features)
	for index, feature := range collection.Features {
		switch feature.Properties["role"] {
		case "band-outline", "total-band-outline":
			if index > lastStaticOutline {
				lastStaticOutline = index
			}
		case "visibility-boundary":
			if index < firstPhaseCurve {
				firstPhaseCurve = index
			}
		}
	}
	if lastStaticOutline < 0 {
		t.Fatal("Mars occultation GeoJSON is missing static band outlines")
	}
	if firstPhaseCurve >= len(collection.Features) {
		t.Fatal("Mars occultation GeoJSON is missing physical phase curves")
	}
	if firstPhaseCurve <= lastStaticOutline {
		t.Fatalf("physical phase curves start at feature %d, after static outline %d required",
			firstPhaseCurve, lastStaticOutline)
	}
}

func TestOccultationP2Mars20250729TotalBandUsesInnerRiseSetCurves(t *testing.T) {
	fixture := mars20250729TestFixture(t, time.Minute, true)
	path := fixture.path
	if !path.HasTotalBand {
		t.Fatal("Mars path is missing its total band")
	}
	if len(path.TotalRiseSetCurves) == 0 {
		t.Fatal("Mars total band is missing inner-contact rise/set curves")
	}
	totalBand := featureWithRole(t, fixture.collection, "total-band")
	if authoritative, ok := totalBand.Properties["static_band_authoritative"].(bool); !ok || !authoritative {
		t.Fatalf("total-band static_band_authoritative=%v, want true", totalBand.Properties["static_band_authoritative"])
	}
	if totalBand.Properties["source"] != "visible-footprint-sweep" {
		t.Fatalf("total-band source=%v, want visible-footprint-sweep from the full event-time footprint union",
			totalBand.Properties["source"])
	}
}

func TestOccultationP2Mars20250729ExportsOnlyOuterContactPhaseBoundaries(t *testing.T) {
	fixture := mars20250729TestFixture(t, time.Minute, true)
	path, collection := fixture.path, fixture.collection
	for _, feature := range featuresWithRole(collection, "visibility-boundary") {
		band, ok := feature.Properties["band"].(string)
		if !ok || band != "partial" {
			t.Fatalf("visibility-boundary has invalid band=%v: %#v", feature.Properties["band"], feature.Properties)
		}
	}
	if count := len(featuresWithRole(collection, "visibility-boundary")); count != len(path.RiseSetCurves) {
		t.Fatalf("visibility-boundary count=%d, want %d outer-contact phase curves", count, len(path.RiseSetCurves))
	}
}

func TestOccultationP2Mars20250729StaticBandRemainsContinuousAndContained(t *testing.T) {
	fixture := mars20250729TestFixture(t, time.Minute, true)
	path, collection := fixture.path, fixture.collection
	assertOccultationP2StaticBand(t, path, collection)
	assertOccultationP2PhaseCurvesInsidePartialBand(t, collection, 1)
}

func TestOccultationP2Saturn20240725StaticBandRemainsContinuousAndContained(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2024, time.July, 25, 0, 0, 0, 0, zone)
	paths, err := moon.FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), moon.OccultationSaturn,
		moon.OccultationPathOptions{
			Step: 20 * time.Minute, TargetSpacingKM: 900, RiseSetStep: time.Minute,
			DisableFootprints: true, IncludeFootprintTimeline: true,
			FootprintTimelineStep: 5 * time.Minute,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	path := paths[0]
	data, err := geojson.MarshalPlanetOccultation(path)
	if err != nil {
		t.Fatalf("MarshalPlanetOccultation: %v", err)
	}
	collection := decodeCollection(t, data)
	assertOccultationP2StaticBand(t, path, collection)
}

func TestOccultationP0Mars20250729StaticBandsUseContactContours(t *testing.T) {
	path := mars20250729TestFixture(t, time.Minute, true).path
	if len(path.PartialBandContours) < 2 {
		t.Fatalf("partial contact contours=%d, want north/south continuous envelopes", len(path.PartialBandContours))
	}
	if !path.HasTotalBand || len(path.TotalBandContours) < 2 {
		t.Fatalf("total contact contours=%d hasTotal=%v, want inner-contact envelopes", len(path.TotalBandContours), path.HasTotalBand)
	}
	collection := mars20250729TestFixture(t, time.Minute, true).collection
	partialBand := featureWithRole(t, collection, "partial-band")
	totalBand := featureWithRole(t, collection, "total-band")
	assertOccultationP0ContactContourSource(t, partialBand)
	assertOccultationP0ContactContourSource(t, totalBand)
	if maximumStep := occultationPointSeriesMaximumStepKM(path.PartialBandContours[0]); maximumStep > 120 {
		t.Fatalf("partial contour maximum step=%.1f km, want <=120 km", maximumStep)
	}
	if maximumStep := occultationPointSeriesMaximumStepKM(path.TotalBandContours[0]); maximumStep > 120 {
		t.Fatalf("total contour maximum step=%.1f km, want <=120 km", maximumStep)
	}
	partialRings := geoJSONMultiPolygonOuterRings(t, partialBand)
	totalRings := geoJSONMultiPolygonOuterRings(t, totalBand)
	assertGeoJSONMultiPolygonMaximumEdge(t, "partial-band", partialRings, 80)
	assertGeoJSONMultiPolygonMaximumEdge(t, "total-band", totalRings, 80)
	assertGeoJSONMultiPolygonFollowsLines(t, "partial-band", partialRings,
		occultationP0BoundaryLines(path.PartialBandContours, path.RiseSetCurves,
			path.PartialBandFootprints, path.NorthernLimit, path.SouthernLimit),
		250,
	)
	assertGeoJSONMultiPolygonFollowsLines(t, "partial-band-authoritative", partialRings,
		occultationP0AuthoritativeBoundaryLines(path.PartialBandContours, path.RiseSetCurves,
			path.PartialBandFootprints, path.NorthernLimit, path.SouthernLimit),
		300,
	)
	assertGeoJSONMultiPolygonFollowsLines(t, "total-band", totalRings,
		occultationP0BoundaryLines(path.TotalBandContours, path.TotalRiseSetCurves,
			path.TotalBandFootprints, path.NorthernTotalLimit, path.SouthernTotalLimit),
		250,
	)
	assertGeoJSONMultiPolygonFollowsLines(t, "total-band-authoritative", totalRings,
		occultationP0AuthoritativeBoundaryLines(path.TotalBandContours, path.TotalRiseSetCurves,
			path.TotalBandFootprints, path.NorthernTotalLimit, path.SouthernTotalLimit),
		300,
	)
	greatest := [][]geodata.GeoPoint{{
		{Longitude: path.Greatest.Longitude, Latitude: path.Greatest.Latitude},
	}}
	if miss := geodata.SphericalPolygonsPathMissDistanceKM(totalRings, greatest, false); miss > 10 {
		t.Fatalf("total-band misses greatest point by %.1f km", miss)
	}
	if miss := geodata.SphericalPolygonsPathMissDistanceKM(partialRings, totalRings, true); miss > 10 {
		t.Fatalf("total-band extends %.1f km outside the partial-band", miss)
	}
}

func TestOccultationP2Mars20250729EndRiseExportsRawPhaseSegments(t *testing.T) {
	fixture := mars20250729TestFixture(t, 5*time.Minute, true)
	path, collection := fixture.path, fixture.collection
	endRise := riseSetBoundaryFeature(t, collection, "end", "rise")
	var lines [][][]float64
	if err := json.Unmarshal(endRise.Geometry.Coordinates, &lines); err != nil {
		t.Fatalf("decode end/rise coordinates: %v", err)
	}
	var sourceSegments [][]moon.OccultationPathPoint
	for _, curve := range path.RiseSetCurves {
		if curve.Phase == moon.RiseSetPhaseEnd && curve.Direction == moon.RiseSetDirectionRise {
			sourceSegments = curve.Segments
			break
		}
	}
	if len(sourceSegments) < 2 {
		t.Fatalf("end/rise source segment count=%d, want a folded multi-branch phase curve", len(sourceSegments))
	}
	if len(lines) != len(sourceSegments) {
		t.Fatalf("end/rise segment count=%d, want raw source segment count %d", len(lines), len(sourceSegments))
	}
	for segmentIndex, line := range lines {
		if len(line) < 2 {
			t.Fatalf("end/rise segment %d has %d points", segmentIndex, len(line))
		}
		source := sourceSegments[segmentIndex]
		if !geoJSONCoordinateMatchesPathPoint(line[0], source[0]) ||
			!geoJSONCoordinateMatchesPathPoint(line[len(line)-1], source[len(source)-1]) {
			t.Fatalf("end/rise segment %d endpoints do not match raw source segment", segmentIndex)
		}
		maximumStep := 0.0
		for index := 1; index < len(line); index++ {
			step := geoJSONCoordinateDistanceKM(line[index-1], line[index])
			if step > maximumStep {
				maximumStep = step
			}
		}
		if maximumStep > 200 {
			t.Fatalf("end/rise segment %d has %.1f km maximum step, want smooth visible curvature", segmentIndex, maximumStep)
		}
	}
}

func geoJSONCoordinateMatchesPathPoint(point []float64, source moon.OccultationPathPoint) bool {
	return len(point) >= 2 &&
		geoJSONCoordinateDistanceKM(point, []float64{source.Longitude, source.Latitude}) <= 0.1
}

func horizonConnectorFeature(
	t *testing.T,
	collection decodedCollection,
	horizon string,
) decodedFeature {
	t.Helper()
	for _, feature := range featuresWithRole(collection, "horizon-connector") {
		if feature.Properties["phase"] == "horizon" && feature.Properties["horizon"] == horizon {
			return feature
		}
	}
	t.Fatalf("horizon-connector %s not found", horizon)
	return decodedFeature{}
}

func assertOccultationHorizonConnectorsAreAuxiliary(t *testing.T, collection decodedCollection) {
	t.Helper()
	if len(featuresWithRole(collection, "horizon-connector")) == 0 {
		t.Fatal("GeoJSON is missing auxiliary horizon-connector features")
	}
	for _, feature := range featuresWithRole(collection, "visibility-boundary") {
		if feature.Properties["phase"] == "horizon" {
			t.Fatal("horizon connector was exported as a visibility-boundary")
		}
	}
}

func assertOccultationP0ContactContourSource(t *testing.T, feature decodedFeature) {
	t.Helper()
	role := feature.Properties["role"]
	if authoritative, ok := feature.Properties["static_band_authoritative"].(bool); !ok || !authoritative {
		t.Fatalf("%s static_band_authoritative=%v, want true", role, feature.Properties["static_band_authoritative"])
	}
	if source := feature.Properties["source"]; source != "visible-footprint-sweep" {
		t.Fatalf("%s source=%v, want visible-footprint-sweep", role, source)
	}
	if boundarySource := feature.Properties["boundary_source"]; boundarySource != "footprint-sweep+horizon-visible" {
		t.Fatalf("%s boundary_source=%v, want footprint-sweep+horizon-visible", role, boundarySource)
	}
}

func assertOccultationP2StaticBand(
	t *testing.T,
	path moon.PlanetOccultationPath,
	collection decodedCollection,
) {
	t.Helper()
	partialBand := featureWithRole(t, collection, "partial-band")
	totalBand := featureWithRole(t, collection, "total-band")
	assertOccultationP0ContactContourSource(t, partialBand)
	assertOccultationP0ContactContourSource(t, totalBand)
	if maximumStep := occultationPointSeriesMaximumStepKM(path.PartialBandContours[0]); maximumStep > 120 {
		t.Fatalf("partial contour maximum step=%.1f km, want <=120 km", maximumStep)
	}
	if path.HasTotalBand {
		if len(path.TotalBandContours) < 2 {
			t.Fatalf("total contour count=%d, want inner-contact envelopes", len(path.TotalBandContours))
		}
		if maximumStep := occultationPointSeriesMaximumStepKM(path.TotalBandContours[0]); maximumStep > 120 {
			t.Fatalf("total contour maximum step=%.1f km, want <=120 km", maximumStep)
		}
	}
	partialRings := geoJSONMultiPolygonOuterRings(t, partialBand)
	totalRings := geoJSONMultiPolygonOuterRings(t, totalBand)
	assertGeoJSONMultiPolygonMaximumEdge(t, "partial-band", partialRings, 220)
	assertGeoJSONMultiPolygonMaximumEdge(t, "total-band", totalRings, 220)
	assertGeoJSONMultiPolygonFollowsLines(t, "partial-band", partialRings,
		occultationP0BoundaryLines(path.PartialBandContours, path.RiseSetCurves,
			path.PartialBandFootprints, path.NorthernLimit, path.SouthernLimit),
		250,
	)
	if path.HasTotalBand {
		assertGeoJSONMultiPolygonFollowsLines(t, "total-band", totalRings,
			occultationP0BoundaryLines(path.TotalBandContours, path.TotalRiseSetCurves,
				path.TotalBandFootprints, path.NorthernTotalLimit, path.SouthernTotalLimit),
			// Split finite-disk branches use a sampled endpoint cap at the polar
			// horizon; allow the bounded 265 km closure residual while retaining
			// the stricter 250 km check for the stable partial envelope above.
			270,
		)
	}
	greatest := [][]geodata.GeoPoint{{
		{Longitude: path.Greatest.Longitude, Latitude: path.Greatest.Latitude},
	}}
	if miss := geodata.SphericalPolygonsPathMissDistanceKM(totalRings, greatest, false); miss > 10 {
		t.Fatalf("total-band misses greatest point by %.1f km", miss)
	}
	if miss := geodata.SphericalPolygonsPathMissDistanceKM(partialRings, totalRings, true); miss > 10 {
		t.Fatalf("total-band extends %.1f km outside the partial-band", miss)
	}
}

func assertOccultationP2PhaseCurvesInsidePartialBand(
	t *testing.T,
	collection decodedCollection,
	toleranceKM float64,
) {
	t.Helper()
	partialBand := featureWithRole(t, collection, "partial-band")
	var polygons [][][][]float64
	if err := json.Unmarshal(partialBand.Geometry.Coordinates, &polygons); err != nil {
		t.Fatalf("decode partial-band coordinates: %v", err)
	}
	maximumMissKM := 0.0
	var maximumMissPoint []float64
	for _, boundary := range featuresWithRole(collection, "visibility-boundary") {
		if boundary.Properties["band"] != "partial" {
			continue
		}
		var lines [][][]float64
		if err := json.Unmarshal(boundary.Geometry.Coordinates, &lines); err != nil {
			t.Fatalf("decode partial visibility-boundary coordinates: %v", err)
		}
		for _, line := range lines {
			for _, point := range line {
				if geometryContainsPoint(t, partialBand.Geometry, point[0], point[1]) {
					continue
				}
				missKM := geoJSONMultiPolygonBoundaryDistanceKM(polygons, point)
				if missKM > maximumMissKM {
					maximumMissKM = missKM
					maximumMissPoint = point
				}
			}
		}
	}
	if maximumMissKM > toleranceKM {
		t.Fatalf("partial visibility-boundary extends %.3f km outside rendered partial-band at %.6f, %.6f, want <=%.1f km",
			maximumMissKM, maximumMissPoint[0], maximumMissPoint[1], toleranceKM)
	}
}

func geoJSONMultiPolygonOuterRings(t *testing.T, feature decodedFeature) [][]geodata.GeoPoint {
	t.Helper()
	if feature.Geometry.Type != "MultiPolygon" {
		t.Fatalf("%s geometry=%q, want MultiPolygon", feature.Properties["role"], feature.Geometry.Type)
	}
	var polygons [][][][]float64
	if err := json.Unmarshal(feature.Geometry.Coordinates, &polygons); err != nil {
		t.Fatalf("decode %s coordinates: %v", feature.Properties["role"], err)
	}
	rings := make([][]geodata.GeoPoint, 0, len(polygons))
	for polygonIndex, polygon := range polygons {
		if len(polygon) == 0 {
			t.Fatalf("%s polygon %d has no rings", feature.Properties["role"], polygonIndex)
		}
		ring := make([]geodata.GeoPoint, len(polygon[0]))
		for pointIndex, point := range polygon[0] {
			if len(point) < 2 {
				t.Fatalf("%s polygon %d point %d is malformed", feature.Properties["role"], polygonIndex, pointIndex)
			}
			ring[pointIndex] = geodata.GeoPoint{Longitude: point[0], Latitude: point[1]}
		}
		rings = append(rings, ring)
	}
	return rings
}

func assertGeoJSONMultiPolygonMaximumEdge(
	t *testing.T,
	role string,
	rings [][]geodata.GeoPoint,
	maximumKM float64,
) {
	t.Helper()
	for ringIndex, ring := range rings {
		for pointIndex := 1; pointIndex < len(ring); pointIndex++ {
			step := geoPointDistanceKM(ring[pointIndex-1], ring[pointIndex])
			if step > maximumKM {
				t.Fatalf("%s ring %d has %.1f km edge at %d, want <= %.1f km",
					role, ringIndex, step, pointIndex, maximumKM)
			}
		}
	}
}

func assertGeoJSONMultiPolygonFollowsLines(
	t *testing.T,
	role string,
	rings [][]geodata.GeoPoint,
	lines [][][]float64,
	maximumDistanceKM float64,
) {
	t.Helper()
	for ringIndex, ring := range rings {
		for pointIndex, point := range ring {
			distance := geoPointLineDistanceKM(point, lines)
			if distance > maximumDistanceKM {
				t.Fatalf("%s ring %d point %d is %.1f km from contact/rise-set boundary, want <= %.1f km: %.6f, %.6f",
					role, ringIndex, pointIndex, distance, maximumDistanceKM, point.Longitude, point.Latitude)
			}
		}
	}
}

func occultationP0BoundaryLines(
	contours [][]moon.OccultationPathPoint,
	curves []moon.OccultationRiseSetCurve,
	footprints []moon.PlanetOccultationFootprint,
	limits ...[]moon.OccultationPathPoint,
) [][][]float64 {
	lines := make([][][]float64, 0, len(contours)+len(curves)*2)
	for _, contour := range contours {
		if line := occultationP0PathLine(contour); len(line) >= 2 {
			lines = append(lines, line)
		}
	}
	for _, curve := range curves {
		for _, segment := range curve.Segments {
			if line := occultationP0PathLine(segment); len(line) >= 2 {
				lines = append(lines, line)
			}
		}
	}
	for _, source := range occultationgeo.ContactSweepBoundaryLines(footprints) {
		line := make([][]float64, 0, len(source))
		for _, point := range source {
			line = append(line, []float64{point.Longitude, point.Latitude})
		}
		if len(line) >= 2 {
			lines = append(lines, line)
		}
	}
	for _, connector := range occultationgeo.HorizonConnectorSegments(footprints, curves, limits...) {
		if line := occultationP0PathLine(connector.Points); len(line) >= 2 {
			lines = append(lines, line)
		}
	}
	return lines
}

func occultationP0AuthoritativeBoundaryLines(
	contours [][]moon.OccultationPathPoint,
	curves []moon.OccultationRiseSetCurve,
	footprints []moon.PlanetOccultationFootprint,
	limits ...[]moon.OccultationPathPoint,
) [][][]float64 {
	lines := make([][][]float64, 0, len(contours)+len(curves)*2)
	for _, contour := range contours {
		if line := occultationP0PathLine(contour); len(line) >= 2 {
			lines = append(lines, line)
		}
	}
	for _, curve := range curves {
		for _, segment := range curve.Segments {
			if line := occultationP0PathLine(segment); len(line) >= 2 {
				lines = append(lines, line)
			}
		}
	}
	for _, connector := range occultationgeo.HorizonConnectorSegments(footprints, curves, limits...) {
		if line := occultationP0PathLine(connector.Points); len(line) >= 2 {
			lines = append(lines, line)
		}
	}
	return lines
}

func occultationP0PathLine(points []moon.OccultationPathPoint) [][]float64 {
	line := make([][]float64, 0, len(points))
	for _, point := range points {
		line = append(line, []float64{point.Longitude, point.Latitude})
	}
	return line
}

func geoPointLineDistanceKM(point geodata.GeoPoint, lines [][][]float64) float64 {
	target := []float64{point.Longitude, point.Latitude}
	minimum := math.Inf(1)
	for _, line := range lines {
		for index := 1; index < len(line); index++ {
			minimum = math.Min(minimum, geoJSONPointSegmentDistanceKM(target, line[index-1], line[index]))
		}
	}
	return minimum
}

func geoPointDistanceKM(first, second geodata.GeoPoint) float64 {
	return geoJSONCoordinateDistanceKM(
		[]float64{first.Longitude, first.Latitude},
		[]float64{second.Longitude, second.Latitude},
	)
}

func occultationPointSeriesMaximumStepKM(points []moon.OccultationPathPoint) float64 {
	maximum := 0.0
	for index := 1; index < len(points); index++ {
		step := geoPointDistanceKM(
			geodata.GeoPoint{Longitude: points[index-1].Longitude, Latitude: points[index-1].Latitude},
			geodata.GeoPoint{Longitude: points[index].Longitude, Latitude: points[index].Latitude},
		)
		if step > maximum {
			maximum = step
		}
	}
	return maximum
}

func geoJSONCoordinateMatchesAny(point []float64, endpoints []decodedRiseSetEndpoint) bool {
	for _, endpoint := range endpoints {
		if geoJSONCoordinateDistanceKM(point, endpoint.coordinate) <= 0.1 {
			return true
		}
	}
	return false
}

func mars20250729PhaseEndpoints(
	t *testing.T,
	collection decodedCollection,
	horizon string,
) []decodedRiseSetEndpoint {
	t.Helper()
	var endpoints []decodedRiseSetEndpoint
	for _, phase := range []string{"start", "greatest", "end"} {
		feature := riseSetBoundaryFeature(t, collection, phase, horizon)
		endpoints = append(endpoints, riseSetFeatureEndpoints(t, feature)...)
	}
	return endpoints
}

func assertClosedMultiLineFeature(t *testing.T, feature decodedFeature) {
	t.Helper()
	if feature.Geometry.Type != "MultiLineString" {
		t.Fatalf("%s geometry=%q, want MultiLineString", feature.Properties["role"], feature.Geometry.Type)
	}
	var lines [][][]float64
	if err := json.Unmarshal(feature.Geometry.Coordinates, &lines); err != nil {
		t.Fatalf("decode %s coordinates: %v", feature.Properties["role"], err)
	}
	if len(lines) == 0 {
		t.Fatalf("%s has no outline segments", feature.Properties["role"])
	}
	for index, line := range lines {
		if len(line) < 4 {
			t.Fatalf("%s segment %d has %d points, want a closed ring", feature.Properties["role"], index, len(line))
		}
		first, last := line[0], line[len(line)-1]
		if len(first) != 2 || len(last) != 2 || first[0] != last[0] || first[1] != last[1] {
			t.Fatalf("%s segment %d is not closed: first=%v last=%v", feature.Properties["role"], index, first, last)
		}
	}
}

func formatP2SignedIndex(value int) string {
	if value < 0 {
		return "-" + formatP2Magnitude(-value)
	}
	return "+" + formatP2Magnitude(value)
}

func formatP2Magnitude(value int) string {
	if value < 10 {
		return "0" + string(rune('0'+value))
	}
	return string(rune('0'+value/10)) + string(rune('0'+value%10))
}

// Keep the JSON import in this P2 file tied to the shared decoder contract;
// this catches accidental changes that make a GeoJSON payload un-decodable
// even when the role-level checks still pass.
func TestP2GeoJSONPayloadsDecodeAsFeatureCollections(t *testing.T) {
	date := time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
		Step: 30 * time.Minute, BoundaryPoints: 36, DisableRiseSet: true,
	})
	if !ok {
		t.Fatal("expected 2024-04-08 eclipse")
	}
	data, err := geojson.MarshalSolarEclipse(partial, nil)
	if err != nil {
		t.Fatal(err)
	}
	var raw struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("payload is not JSON: %v", err)
	}
	if raw.Type != "FeatureCollection" {
		t.Fatalf("payload type=%q, want FeatureCollection", raw.Type)
	}
}
