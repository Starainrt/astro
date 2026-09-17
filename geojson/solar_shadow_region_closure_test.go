package geojson

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	eclipsecore "github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/internal/geodata"
	"github.com/starainrt/astro/internal/solarclosure"
)

// 一致性契约：GeoJSON 的偏食可见域与 SVG 渲染域同源。并集按 SVG 的精确补口口径复算，
// 逐足迹区域的收口点必须是核心层给出的精确擦地点，而不是采样端点外推的近似弧。
func TestSolarEclipseRegionsUseSharedExactClosure(t *testing.T) {
	for _, date := range closureGeoJSONDates() {
		partial, ok := eclipsecore.SolarEclipsePartialFootprints(date, eclipsecore.SolarEclipsePartialFootprintOptions{
			Step: 2 * time.Minute, BoundaryPoints: 96,
		})
		if !ok {
			t.Fatalf("missing partial footprints for %s", date.Format("2006-01-02"))
		}
		data, err := MarshalSolarEclipse(partial, nil)
		if err != nil {
			t.Fatalf("%s: %v", date.Format("2006-01-02"), err)
		}
		band, footprints := closureDecodeSolarRegions(t, data)
		if len(band) == 0 {
			t.Fatalf("%s: no partial-band region", date.Format("2006-01-02"))
		}

		svgBand, ok := closureSVGBandPolygons(partial)
		if !ok {
			t.Fatalf("%s: shared union is unavailable", date.Format("2006-01-02"))
		}
		bandDeviation := closureRegionDeviationKM(svgBand, band)
		if bandDeviation > closureGeoJSONBandToleranceKM {
			t.Fatalf("%s: band regions differ by %.6f km, tolerance %.6f km",
				date.Format("2006-01-02"), bandDeviation, closureGeoJSONBandToleranceKM)
		}

		compared, skipped := 0, 0
		worst := 0.0
		for _, footprint := range partial.Footprints {
			region := footprints[footprint.Time.UTC().Format(time.RFC3339Nano)]
			if len(region) == 0 {
				continue
			}
			if closureRegionReachesPole(region) {
				skipped++
				continue
			}
			if !footprint.Closed {
				for _, end := range footprint.HorizonEnds {
					if !closureRegionCarriesPoint(region, geodata.GeoPoint{
						Longitude: end.Longitude, Latitude: end.Latitude,
					}) {
						t.Fatalf("%s: the %s region is not closed at the grazing point %v",
							date.Format("2006-01-02"), footprint.Time.Format(time.RFC3339), end)
					}
				}
				for _, end := range footprint.HorizonEnds {
					worst = math.Max(worst, geodata.SphericalPolygonsPathMissDistanceKM(
						region,
						[][]geodata.GeoPoint{{{Longitude: end.Longitude, Latitude: end.Latitude}}},
						false,
					))
				}
			}
			compared++
		}
		if compared == 0 {
			t.Fatalf("%s: no footprint region was checked", date.Format("2006-01-02"))
		}
		if worst > closureGeoJSONGrazingToleranceKM {
			t.Fatalf("%s: grazing points miss the region by %.6f km, tolerance %.6f km",
				date.Format("2006-01-02"), worst, closureGeoJSONGrazingToleranceKM)
		}
		t.Logf("%s band=%.6f km footprints=%d skipped=%d grazing=%.6f km",
			date.Format("2006-01-02"), bandDeviation, compared, skipped, worst)
	}
}

func closureGeoJSONDates() []time.Time {
	return []time.Time{
		time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC),
		time.Date(2010, time.January, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2014, time.April, 29, 0, 0, 0, 0, time.UTC),
		time.Date(2009, time.July, 22, 0, 0, 0, 0, time.UTC),
	}
}

// closureSVGBandPolygons 用 SVG 侧的口径复算并集：瞬时足迹按精确擦地点补口。
func closureSVGBandPolygons(
	partial eclipsecore.SolarEclipsePartialFootprintsInfo,
) ([][]geodata.GeoPoint, bool) {
	if len(partial.PartialBandContours) == 0 || len(partial.RiseSetCurves) == 0 {
		return nil, false
	}
	contours := make([][]geodata.GeoPoint, 0, len(partial.PartialBandContours))
	for _, contour := range partial.PartialBandContours {
		contours = append(contours, solarCentralBandGeoPoints(contour))
	}
	phaseLines := make([][]geodata.GeoPoint, 0, len(partial.RiseSetCurves)*2)
	for _, curve := range partial.RiseSetCurves {
		for _, segment := range curve.Segments {
			phaseLines = append(phaseLines, solarCentralBandGeoPoints(segment))
		}
	}
	footprints := make([]solarclosure.Footprint, 0, len(partial.Footprints))
	for _, footprint := range partial.Footprints {
		input, err := solarClosureFootprint(footprint)
		if err != nil {
			return nil, false
		}
		footprints = append(footprints, input)
	}
	return solarclosure.BandPolygons(
		contours, phaseLines, footprints, true, solarclosure.SnapDistanceKM,
	)
}

func closureDecodeSolarRegions(
	t *testing.T,
	data []byte,
) (band [][]geodata.GeoPoint, footprints map[string][][]geodata.GeoPoint) {
	t.Helper()
	var collection struct {
		Features []struct {
			Properties map[string]interface{} `json:"properties"`
			Geometry   struct {
				Type        string          `json:"type"`
				Coordinates json.RawMessage `json:"coordinates"`
			} `json:"geometry"`
		} `json:"features"`
	}
	if err := json.Unmarshal(data, &collection); err != nil {
		t.Fatalf("decode GeoJSON: %v", err)
	}
	footprints = make(map[string][][]geodata.GeoPoint)
	for _, feature := range collection.Features {
		role, _ := feature.Properties["role"].(string)
		if feature.Geometry.Type != "MultiPolygon" {
			continue
		}
		var polygons [][][][]float64
		if err := json.Unmarshal(feature.Geometry.Coordinates, &polygons); err != nil {
			t.Fatalf("decode %s coordinates: %v", role, err)
		}
		rings := make([][]geodata.GeoPoint, 0, len(polygons))
		for _, polygon := range polygons {
			for _, ring := range polygon {
				points := make([]geodata.GeoPoint, len(ring))
				for index, point := range ring {
					points[index] = geodata.GeoPoint{Longitude: point[0], Latitude: point[1]}
				}
				rings = append(rings, points)
			}
		}
		switch role {
		case "partial-band":
			band = append(band, rings...)
		case "partial-footprint":
			value, _ := feature.Properties["time"].(string)
			footprints[value] = append(footprints[value], rings...)
		}
	}
	return band, footprints
}

func closureRegionReachesPole(region [][]geodata.GeoPoint) bool {
	for _, ring := range region {
		for _, point := range ring {
			if math.Abs(point.Latitude) >= 89.9 {
				return true
			}
		}
	}
	return false
}

func closureRegionCarriesPoint(region [][]geodata.GeoPoint, target geodata.GeoPoint) bool {
	for _, ring := range region {
		for _, point := range ring {
			if math.Abs(point.Latitude-target.Latitude) < 1e-6 &&
				math.Abs(math.Remainder(point.Longitude-target.Longitude, 360)) < 1e-6 {
				return true
			}
		}
	}
	return false
}

func closureRegionDeviationKM(first, second [][]geodata.GeoPoint) float64 {
	return math.Max(
		geodata.SphericalPolygonsPathMissDistanceKM(second, first, true),
		geodata.SphericalPolygonsPathMissDistanceKM(first, second, true),
	)
}

const (
	// 两侧并集只差面选择口径，实测在米级。
	closureGeoJSONBandToleranceKM = 1.0
	// 擦地点是区域的收口顶点，序列化往返只允许亚米级偏差。
	closureGeoJSONGrazingToleranceKM = 0.01
)
