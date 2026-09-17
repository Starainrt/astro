package geojson_test

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
)

func TestMarshalSolarEclipse20560713CentralBandHasNoEndFold(t *testing.T) {
	date := time.Date(2056, time.July, 13, 0, 0, 0, 0, time.UTC)
	localDate := time.Date(2056, time.July, 12, 0, 0, 0, 0, time.UTC)
	local, localOK := eclipse.LocalSolarEclipseOnDate(localDate, -64.3737, -5.4978, 0)
	if !localOK || local.Type != eclipse.SolarEclipseAnnular || !local.HasCentral || local.SunAltitude <= 0 {
		t.Fatalf("reference site is not a visible annular eclipse: ok=%v type=%s central=%v altitude=%.6f",
			localOK, local.Type, local.HasCentral, local.SunAltitude)
	}
	partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
		Step:              2 * time.Minute,
		BoundaryPoints:    96,
		CentralShadowStep: 2 * time.Minute,
		MagnitudeValues:   []float64{0.2, 0.4, 0.6, 0.8, 1.0},
	})
	if !ok {
		t.Fatal("expected solar eclipse footprints")
	}
	central, ok := eclipse.SolarEclipseCentralPath(date, eclipse.SolarEclipsePathOptions{
		Step: 2 * time.Minute, TargetSpacingKM: 700,
	})
	if !ok || len(central.CenterLine) < 2 {
		t.Fatal("expected a central eclipse path")
	}
	data, err := geojson.MarshalSolarEclipse(partial, &central)
	if err != nil {
		t.Fatalf("MarshalSolarEclipse: %v", err)
	}
	collection := decodeCollection(t, data)
	band := featureWithRole(t, collection, "central-band")
	assertClosedMultiPolygon(t, band)

	var polygons [][][][]float64
	if err := json.Unmarshal(band.Geometry.Coordinates, &polygons); err != nil {
		t.Fatalf("decode central band: %v", err)
	}
	if len(polygons) == 0 || len(polygons) > 2 {
		t.Fatalf("central band has %d polygons, want one physical band with at most one antimeridian split", len(polygons))
	}
	if len(polygons) == 2 && !solarCentralBandPartsMeetAntimeridian(polygons) {
		t.Fatal("two central-band polygons do not form an antimeridian split")
	}
	for _, polygon := range polygons {
		if len(polygon) == 0 {
			t.Fatal("central-band polygon has no exterior ring")
		}
		assertSolarCentralBandRingSimpleAndSampled(t, polygon[0], 250)
	}
	if !geometryContainsPoint(t, band.Geometry, -64.3737, -5.4978) {
		t.Fatal("central-band omits the locally visible annular greatest point near U4")
	}

	center := featureWithRole(t, collection, "center-line")
	var centerLines [][][]float64
	if err := json.Unmarshal(center.Geometry.Coordinates, &centerLines); err != nil {
		t.Fatalf("decode center line: %v", err)
	}
	for segmentIndex, segment := range centerLines {
		for pointIndex, point := range segment {
			if !geometryContainsPoint(t, band.Geometry, point[0], point[1]) &&
				geoJSONMultiPolygonBoundaryDistanceKM(polygons, point) > 10 {
				t.Fatalf("center-line segment %d point %d lies outside central band", segmentIndex, pointIndex)
			}
		}
	}

	start := central.CenterLine[0]
	end := central.CenterLine[len(central.CenterLine)-1]
	assertSolarCentralBandDoesNotReverseNear(t, polygons, []float64{start.Longitude, start.Latitude}, 1600)
	assertSolarCentralBandDoesNotReverseNear(t, polygons, []float64{end.Longitude, end.Latitude}, 1600)
}

func TestMarshalSolarEclipse20560713HorizonClosureIsStableAcrossSampling(t *testing.T) {
	date := time.Date(2056, time.July, 13, 0, 0, 0, 0, time.UTC)
	var referenceRoots [2][2]eclipse.SolarEclipsePathPoint
	for _, step := range []time.Duration{time.Minute, 2 * time.Minute, 5 * time.Minute, 10 * time.Minute} {
		t.Run(step.String(), func(t *testing.T) {
			partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
				Step: step, BoundaryPoints: 96, CentralShadowStep: step,
			})
			if !ok {
				t.Fatal("expected solar eclipse footprints")
			}
			if len(partial.CentralBandHorizonClosures) != 2 {
				t.Fatalf("central-limit horizon closures=%d, want start and end", len(partial.CentralBandHorizonClosures))
			}
			for closureIndex, closure := range partial.CentralBandHorizonClosures {
				if len(closure) < 2 {
					t.Fatalf("horizon closure %d has %d points", closureIndex, len(closure))
				}
				assertSolarPathMaximumEdgeKM(t, closure, 12)
				roots := [2]eclipse.SolarEclipsePathPoint{closure[0], closure[len(closure)-1]}
				if referenceRoots[closureIndex][0].Time.IsZero() {
					referenceRoots[closureIndex] = roots
				} else {
					for rootIndex := range roots {
						if difference := roots[rootIndex].Time.Sub(referenceRoots[closureIndex][rootIndex].Time); difference < -time.Millisecond || difference > time.Millisecond {
							t.Fatalf("horizon closure %d root %d time differs by %s across sampling", closureIndex, rootIndex, difference)
						}
						if distance := geoJSONCoordinateDistanceKM(
							[]float64{roots[rootIndex].Longitude, roots[rootIndex].Latitude},
							[]float64{referenceRoots[closureIndex][rootIndex].Longitude, referenceRoots[closureIndex][rootIndex].Latitude},
						); distance > 0.01 {
							t.Fatalf("horizon closure %d root %d differs by %.3f km across sampling", closureIndex, rootIndex, distance)
						}
					}
				}
			}
			central, ok := eclipse.SolarEclipseCentralPath(date, eclipse.SolarEclipsePathOptions{
				Step: step, TargetSpacingKM: 700,
			})
			if !ok {
				t.Fatal("expected central eclipse path")
			}
			data, err := geojson.MarshalSolarEclipse(partial, &central)
			if err != nil {
				t.Fatalf("MarshalSolarEclipse: %v", err)
			}
			band := featureWithRole(t, decodeCollection(t, data), "central-band")
			if band.Properties["source"] != "besselian-critical-envelope" {
				t.Fatalf("central-band source=%v, want continuous critical envelope", band.Properties["source"])
			}
			if !geometryContainsPoint(t, band.Geometry, -64.3737, -5.4978) {
				t.Fatal("central-band omits the locally visible annular point")
			}
			var polygons [][][][]float64
			if err := json.Unmarshal(band.Geometry.Coordinates, &polygons); err != nil {
				t.Fatalf("decode central band: %v", err)
			}
			if len(polygons) == 0 || len(polygons) > 2 {
				t.Fatalf("central band has %d polygons", len(polygons))
			}
			for _, polygon := range polygons {
				assertSolarCentralBandRingSimpleAndSampled(t, polygon[0], 250)
			}
		})
	}
}

func solarCentralBandPartsMeetAntimeridian(polygons [][][][]float64) bool {
	hasEast, hasWest := false, false
	for _, polygon := range polygons {
		for _, ring := range polygon {
			for _, point := range ring {
				hasEast = hasEast || math.Abs(point[0]-180) <= 1e-9
				hasWest = hasWest || math.Abs(point[0]+180) <= 1e-9
			}
		}
	}
	return hasEast && hasWest
}

func assertSolarCentralBandRingSimpleAndSampled(t *testing.T, ring [][]float64, maximumEdgeKM float64) {
	t.Helper()
	for index := 1; index < len(ring); index++ {
		if distance := geoJSONCoordinateDistanceKM(ring[index-1], ring[index]); distance > maximumEdgeKM {
			t.Fatalf("central-band edge %d is %.1f km, want at most %.1f km", index, distance, maximumEdgeKM)
		}
	}
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
}

func assertSolarCentralBandDoesNotReverseNear(t *testing.T, polygons [][][][]float64, center []float64, radiusKM float64) {
	t.Helper()
	checked := 0
	for _, polygon := range polygons {
		if len(polygon) == 0 {
			continue
		}
		ring := polygon[0]
		for index := 1; index+1 < len(ring); index++ {
			if geoJSONCoordinateDistanceKM(ring[index], center) > radiusKM {
				continue
			}
			checked++
			incoming := solarRegressionProjectedVector(ring[index], ring[index-1])
			outgoing := solarRegressionProjectedVector(ring[index], ring[index+1])
			incomingLength := math.Hypot(incoming[0], incoming[1])
			outgoingLength := math.Hypot(outgoing[0], outgoing[1])
			if incomingLength == 0 || outgoingLength == 0 {
				continue
			}
			cosine := (incoming[0]*outgoing[0] + incoming[1]*outgoing[1]) / (incomingLength * outgoingLength)
			if cosine > 0.985 {
				t.Fatalf("central-band boundary reverses by %.1f degrees at %.5f, %.5f between [%.5f, %.5f] and [%.5f, %.5f]",
					math.Acos(math.Max(-1, math.Min(1, cosine)))*180/math.Pi,
					ring[index][0], ring[index][1],
					ring[index-1][0], ring[index-1][1],
					ring[index+1][0], ring[index+1][1])
			}
		}
	}
	if checked == 0 {
		t.Fatal("central-band boundary has no samples near the event end")
	}
}

func solarRegressionProjectedVector(origin, point []float64) [2]float64 {
	latitude := origin[1] * math.Pi / 180
	deltaLongitude := math.Remainder(point[0]-origin[0], 360)
	return [2]float64{deltaLongitude * math.Cos(latitude), point[1] - origin[1]}
}
