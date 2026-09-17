package svg

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/internal/geodata"
)

// 一致性契约：同一事件下 SVG 与 GeoJSON 的偏食可见域必须落在同一几何上。
// 并集只允许面选择带来的亚千米差异，逐足迹区域必须逐个重合在精确擦地点上。
func TestSolarEclipsePartialRegionMatchesGeoJSON(t *testing.T) {
	for _, date := range closureConsistencyDates() {
		info, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
			Step: 2 * time.Minute, BoundaryPoints: 96,
		})
		if !ok {
			t.Fatalf("missing partial footprints for %s", date.Format("2006-01-02"))
		}
		data, err := geojson.MarshalSolarEclipse(info, nil)
		if err != nil {
			t.Fatalf("%s: %v", date.Format("2006-01-02"), err)
		}
		regions := closureGeoJSONRegions(t, data)
		band := regions["partial-band"]
		if len(band) == 0 {
			t.Fatalf("%s: GeoJSON carries no partial-band region", date.Format("2006-01-02"))
		}
		polygons, ok := solarEclipsePartialBandPolygons(info)
		if !ok {
			t.Fatalf("%s: SVG carries no partial-band union", date.Format("2006-01-02"))
		}
		bandDeviation := closureSetDeviationKM(polygons, band)
		if bandDeviation > closureBandToleranceKM {
			t.Fatalf("%s: band regions differ by %.6f km, tolerance %.6f km",
				date.Format("2006-01-02"), bandDeviation, closureBandToleranceKM)
		}

		footprintDeviation := 0.0
		compared, skipped := 0, 0
		for _, footprint := range info.Footprints {
			reference := closureFootprintRegionByTime(regions, footprint.Time)
			if len(reference) == 0 {
				continue
			}
			// 极区足迹沿地图上下边闭合成环，两侧的拓扑表示不同，只比对非极区足迹。
			if closureRegionsReachPole(reference) {
				skipped++
				continue
			}
			polygon := solarEclipsePartialFootprintPolygon(footprint, geodata.ProjectionEquirectangular)
			if len(polygon) < 3 {
				t.Fatalf("%s: SVG dropped the %s footprint region",
					date.Format("2006-01-02"), footprint.Time.Format(time.RFC3339))
			}
			deviation := closureRingDeviationKM(polygon, reference)
			if !footprint.Closed && !closureRingCarriesGrazingPoints(polygon, footprint) {
				t.Fatalf("%s: the %s footprint region ignores its exact grazing points (deviation %.4f km)",
					date.Format("2006-01-02"), footprint.Time.Format(time.RFC3339), deviation)
			}
			footprintDeviation = math.Max(footprintDeviation, deviation)
			compared++
		}
		if compared == 0 {
			t.Fatalf("%s: no footprint region pair was compared", date.Format("2006-01-02"))
		}
		if footprintDeviation > closureFootprintToleranceKM {
			t.Fatalf("%s: footprint regions differ by %.6f km, tolerance %.6f km",
				date.Format("2006-01-02"), footprintDeviation, closureFootprintToleranceKM)
		}
		t.Logf("%s band=%.6f km footprints=%d skipped=%d max=%.6f km",
			date.Format("2006-01-02"), bandDeviation, compared, skipped, footprintDeviation)
	}
}

// 缺少精确擦地点时必须降级到逐足迹扫掠，并让 data-source 落到采样扫掠词条。
func TestSolarEclipsePartialRegionDegradesWithoutExactHorizon(t *testing.T) {
	date := time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC)
	info, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
		Step: 10 * time.Minute, BoundaryPoints: 96,
	})
	if !ok {
		t.Fatal("missing 2024-04-08 partial footprints")
	}
	options := SolarEclipseMapSVGOptions{Width: 1200, Height: 800, Location: time.UTC}
	frame := solarEclipseMapLayoutFor(options, geodata.ProjectionEquirectangular,
		geodata.GeoPoint{Longitude: info.Eclipse.GreatestLongitude, Latitude: info.Eclipse.GreatestLatitude}).frame
	if _, source := solarEclipsePartialSweepPath(info, frame); source != eclipseMapSourcePartialBandUnion {
		t.Fatalf("source=%q, want %q", source, eclipseMapSourcePartialBandUnion)
	}

	degraded := info
	degraded.Footprints = append([]eclipse.SolarEclipsePartialFootprint(nil), info.Footprints...)
	open := 0
	for index := range degraded.Footprints {
		if degraded.Footprints[index].Closed {
			continue
		}
		degraded.Footprints[index].HorizonEnds = nil
		open++
	}
	if open == 0 {
		t.Fatal("the event has no horizon-cut footprint to degrade")
	}
	if polygons, ok := solarEclipsePartialBandPolygons(degraded); ok || polygons != nil {
		t.Fatalf("union survived without exact grazing points: %d polygons", len(polygons))
	}
	path, source := solarEclipsePartialSweepPath(degraded, frame)
	if source != eclipseMapSourceSampledFootprintSweep {
		t.Fatalf("degraded source=%q, want %q", source, eclipseMapSourceSampledFootprintSweep)
	}
	if path == "" {
		t.Fatal("degraded sweep rendered no geometry")
	}
}

func closureConsistencyDates() []time.Time {
	return []time.Time{
		time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC),
		time.Date(2010, time.January, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2014, time.April, 29, 0, 0, 0, 0, time.UTC),
		time.Date(2009, time.July, 22, 0, 0, 0, 0, time.UTC),
	}
}

type closureGeoJSONFeature struct {
	Properties map[string]interface{} `json:"properties"`
	Geometry   struct {
		Type        string          `json:"type"`
		Coordinates json.RawMessage `json:"coordinates"`
	} `json:"geometry"`
}

type closureGeoJSONCollection struct {
	Features []closureGeoJSONFeature `json:"features"`
}

// closureGeoJSONRegions 按角色汇总多边形区域，时间属性用作逐足迹区域的键。
func closureGeoJSONRegions(t *testing.T, data []byte) map[string][][]geodata.GeoPoint {
	t.Helper()
	var collection closureGeoJSONCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		t.Fatalf("decode GeoJSON: %v", err)
	}
	regions := make(map[string][][]geodata.GeoPoint)
	for _, feature := range collection.Features {
		role, _ := feature.Properties["role"].(string)
		if feature.Geometry.Type != "MultiPolygon" {
			continue
		}
		var polygons [][][][]float64
		if err := json.Unmarshal(feature.Geometry.Coordinates, &polygons); err != nil {
			t.Fatalf("decode %s coordinates: %v", role, err)
		}
		key := role
		if role == "partial-footprint" {
			value, _ := feature.Properties["time"].(string)
			key = role + "@" + value
		}
		for _, polygon := range polygons {
			for _, ring := range polygon {
				points := make([]geodata.GeoPoint, len(ring))
				for index, point := range ring {
					if len(point) < 2 {
						t.Fatalf("%s carries a %d-value coordinate", role, len(point))
					}
					points[index] = geodata.GeoPoint{Longitude: point[0], Latitude: point[1]}
				}
				regions[key] = append(regions[key], points)
			}
		}
	}
	return regions
}

// closureRegionsReachPole 报告区域是否触及极点；极区环沿地图边闭合，不参与几何比对。
func closureRegionsReachPole(regions [][]geodata.GeoPoint) bool {
	for _, ring := range regions {
		for _, point := range ring {
			if math.Abs(point.Latitude) >= 89.9 {
				return true
			}
		}
	}
	return false
}

// closureRingCarriesGrazingPoints 报告填充环是否仍以精确擦地点收口。
func closureRingCarriesGrazingPoints(ring []geodata.GeoPoint, footprint eclipse.SolarEclipsePartialFootprint) bool {
	if len(footprint.HorizonEnds) != 2 {
		return false
	}
	for _, end := range footprint.HorizonEnds {
		found := false
		for _, point := range ring {
			if math.Abs(point.Longitude-end.Longitude) < 1e-9 && math.Abs(point.Latitude-end.Latitude) < 1e-9 {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func closureFootprintRegionByTime(regions map[string][][]geodata.GeoPoint, value time.Time) [][]geodata.GeoPoint {
	return regions["partial-footprint@"+value.UTC().Format(time.RFC3339Nano)]
}

func closureSetDeviationKM(first, second [][]geodata.GeoPoint) float64 {
	return math.Max(
		geodata.SphericalPolygonsPathMissDistanceKM(second, first, true),
		geodata.SphericalPolygonsPathMissDistanceKM(first, second, true),
	)
}

func closureRingDeviationKM(ring []geodata.GeoPoint, reference [][]geodata.GeoPoint) float64 {
	return closureSetDeviationKM([][]geodata.GeoPoint{ring}, reference)
}

const (
	// 并集的面选择由瞬时足迹提示决定，两种闭合口径的实测差异在米级。
	closureBandToleranceKM = 1.0
	// 逐足迹区域两侧都直接调用同一个精确闭合，只允许序列化往返误差。
	closureFootprintToleranceKM = 0.01
)
