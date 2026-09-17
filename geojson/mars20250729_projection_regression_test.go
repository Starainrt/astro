package geojson_test

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/internal/geodata"
)

func TestMars20250729RenderedLineworkUsesProjectedSpacing(t *testing.T) {
	collection := mars20250729TestFixture(t, time.Minute, true).collection
	for _, role := range []string{"partial-band", "total-band", "band-outline", "total-band-outline"} {
		feature := featureWithRole(t, collection, role)
		if maximum := maxProjectedGeometryEdgeKM(feature.Geometry.Coordinates); maximum > 50 {
			t.Fatalf("%s projected edge=%.1f km, want <=50 km", role, maximum)
		}
	}
	for _, feature := range featuresWithRole(collection, "visibility-boundary") {
		if maximum := maxProjectedGeometryEdgeKM(feature.Geometry.Coordinates); maximum > 40 {
			t.Fatalf("visibility-boundary projected edge=%.1f km, want <=40 km", maximum)
		}
	}
	for _, feature := range featuresWithRole(collection, "horizon-connector") {
		if maximum := maxProjectedGeometryEdgeKM(feature.Geometry.Coordinates); maximum > 40 {
			t.Fatalf("horizon-connector projected edge=%.1f km, want <=40 km", maximum)
		}
	}
}

func TestMars20250729AuthoritativeBandsRejectPolarBacktracks(t *testing.T) {
	collection := mars20250729TestFixture(t, time.Minute, true).collection
	for _, role := range []string{"partial-band", "total-band"} {
		for ringIndex, ring := range geoJSONMultiPolygonOuterRings(t, featureWithRole(t, collection, role)) {
			for pointIndex := 1; pointIndex+1 < len(ring); pointIndex++ {
				if geoPointDistanceKM(ring[pointIndex-1], ring[pointIndex+1]) > 20 {
				} else {
					angle := geoJSONRingTurnDegrees(ring[pointIndex-1], ring[pointIndex], ring[pointIndex+1])
					if angle < 30 {
						t.Fatalf("%s ring %d retains a %.2f degree polar backtrack at point %d", role, ringIndex, angle, pointIndex)
					}
				}

				previous, middle, next := ring[pointIndex-1], ring[pointIndex], ring[pointIndex+1]
				if math.Abs(middle.Latitude) < 60 ||
					(middle.Latitude-previous.Latitude)*(next.Latitude-middle.Latitude) >= 0 ||
					projectedGeoJSONPointDistanceKM(previous, next) > 80 {
					continue
				}
				if angle := projectedGeoJSONRingTurnDegrees(previous, middle, next); angle < 110 {
					t.Fatalf("%s ring %d retains a %.2f degree projected sweep junction at point %d", role, ringIndex, angle, pointIndex)
				}
			}
		}
	}
}

func TestMars20250729OuterPhaseEnvelopeSharesBandOutline(t *testing.T) {
	collection := mars20250729TestFixture(t, time.Minute, true).collection
	outline := featureWithRole(t, collection, "band-outline")
	var outlineLines [][][]float64
	if err := json.Unmarshal(outline.Geometry.Coordinates, &outlineLines); err != nil {
		t.Fatalf("decode band-outline: %v", err)
	}
	if len(outlineLines) != 1 || len(outlineLines[0]) < 1000 {
		t.Fatalf("band-outline has %d lines and %d points, want one retained outer ring", len(outlineLines), len(outlineLines[0]))
	}
	maximumMatchedFraction := 0.0
	maximumSourceArcKM := 0.0
	for _, boundary := range featuresWithRole(collection, "visibility-boundary") {
		phase, _ := boundary.Properties["phase"].(string)
		if phase != "start" && phase != "end" {
			continue
		}
		var lines [][][]float64
		if err := json.Unmarshal(boundary.Geometry.Coordinates, &lines); err != nil {
			t.Fatalf("decode visibility-boundary: %v", err)
		}
		for _, line := range lines {
			if len(line) < 2 {
				continue
			}
			arc := 0.0
			matched := 0
			for index, point := range line {
				if index > 0 {
					arc += geoJSONCoordinateDistanceKM(line[index-1], point)
				}
				if geoPointLineDistanceKM(
					geodata.GeoPoint{Longitude: point[0], Latitude: point[1]}, outlineLines,
				) <= 0.5 {
					matched++
				}
			}
			if arc > maximumSourceArcKM {
				maximumSourceArcKM = arc
				maximumMatchedFraction = float64(matched) / float64(len(line))
			}
		}
	}
	if maximumSourceArcKM < 3000 || maximumMatchedFraction < 0.98 {
		t.Fatalf("outer phase envelope matched fraction=%.3f over %.1f km, want >=.98 over the long exterior arc",
			maximumMatchedFraction, maximumSourceArcKM)
	}
}

func geoJSONRingTurnDegrees(first, middle, last geodata.GeoPoint) float64 {
	latitude := middle.Latitude * math.Pi / 180
	scale := math.Cos(latitude)
	firstX := (first.Longitude - middle.Longitude) * scale
	firstY := first.Latitude - middle.Latitude
	lastX := (last.Longitude - middle.Longitude) * scale
	lastY := last.Latitude - middle.Latitude
	firstLength := math.Hypot(firstX, firstY)
	lastLength := math.Hypot(lastX, lastY)
	if firstLength <= 1e-12 || lastLength <= 1e-12 {
		return 180
	}
	cosine := (firstX*lastX + firstY*lastY) / (firstLength * lastLength)
	cosine = math.Max(-1, math.Min(1, cosine))
	return math.Acos(cosine) * 180 / math.Pi
}

func projectedGeoJSONRingTurnDegrees(first, middle, last geodata.GeoPoint) float64 {
	firstX, firstY := projectedGeoJSONPoint(first)
	middleX, middleY := projectedGeoJSONPoint(middle)
	lastX, lastY := projectedGeoJSONPoint(last)
	firstX, firstY = firstX-middleX, firstY-middleY
	lastX, lastY = lastX-middleX, lastY-middleY
	firstLength := math.Hypot(firstX, firstY)
	lastLength := math.Hypot(lastX, lastY)
	if firstLength <= 1e-12 || lastLength <= 1e-12 {
		return 180
	}
	cosine := (firstX*lastX + firstY*lastY) / (firstLength * lastLength)
	cosine = math.Max(-1, math.Min(1, cosine))
	return math.Acos(cosine) * 180 / math.Pi
}

func projectedGeoJSONPointDistanceKM(first, second geodata.GeoPoint) float64 {
	firstX, firstY := projectedGeoJSONPoint(first)
	secondX, secondY := projectedGeoJSONPoint(second)
	return math.Hypot(secondX-firstX, secondY-firstY)
}

func projectedGeoJSONPoint(point geodata.GeoPoint) (float64, float64) {
	latitude := math.Max(-85.05112878, math.Min(85.05112878, point.Latitude)) * math.Pi / 180
	return 6378.1366 * point.Longitude * math.Pi / 180,
		6378.1366 * math.Log(math.Tan(math.Pi/4+latitude/2))
}

func maxProjectedGeometryEdgeKM(raw json.RawMessage) float64 {
	var value interface{}
	if err := json.Unmarshal(raw, &value); err != nil {
		return math.Inf(1)
	}
	return maxProjectedGeometryValueEdgeKM(value)
}

func maxProjectedGeometryValueEdgeKM(value interface{}) float64 {
	array, ok := value.([]interface{})
	if !ok || len(array) == 0 {
		return 0
	}
	if len(array) >= 2 {
		if _, ok := array[0].(float64); ok {
			return 0
		}
		if _, ok := array[0].([]interface{}); ok {
			if first, ok := array[0].([]interface{}); ok && len(first) >= 2 {
				if _, ok := first[0].(float64); ok {
					maximum := 0.0
					for index := 1; index < len(array); index++ {
						previous := array[index-1].([]interface{})
						current := array[index].([]interface{})
						maximum = math.Max(maximum, projectedCoordinateDistanceKM(previous, current))
					}
					return maximum
				}
			}
		}
	}
	maximum := 0.0
	for _, child := range array {
		maximum = math.Max(maximum, maxProjectedGeometryValueEdgeKM(child))
	}
	return maximum
}

func projectedCoordinateDistanceKM(first, second []interface{}) float64 {
	longitudeFirst := first[0].(float64)
	latitudeFirst := math.Max(-85.05112878, math.Min(85.05112878, first[1].(float64))) * math.Pi / 180
	longitudeSecond := second[0].(float64)
	latitudeSecond := math.Max(-85.05112878, math.Min(85.05112878, second[1].(float64))) * math.Pi / 180
	longitude := math.Remainder(longitudeSecond-longitudeFirst, 360) * math.Pi / 180
	firstY := math.Log(math.Tan(math.Pi/4 + latitudeFirst/2))
	secondY := math.Log(math.Tan(math.Pi/4 + latitudeSecond/2))
	return 6378.1366 * math.Hypot(longitude, secondY-firstY)
}
