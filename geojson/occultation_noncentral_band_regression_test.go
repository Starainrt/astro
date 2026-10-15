package geojson_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/internal/geodata"
	"github.com/starainrt/astro/moon"
)

// TestMarshalPlanetOccultationNonCentralPartialBandContainsTotalBand 是非中心月掩偏掩带
// 截断的回归：默认（密集瞬时足迹）模式下偏掩带曾只由纯足迹扫掠构造，极向部分被截断，
// 使月升可见性边界落在掩带之外，全掩带反而越出偏掩带。两场事件都是月影轴不与地球椭球
// 相交、且相位曲线在极区折点处与零残差相切的非中心事件；全掩带必须完全落在偏掩带内。
// TestMarshalPlanetOccultationNonCentralPartialBandContainsTotalBand covers the truncated
// partial band of non-central occultations: in the default dense-footprint mode the band
// used to be built by a pure footprint sweep, which cut off its poleward part, left the
// moonrise visibility boundary outside the band and let the total band escape it. Both
// events have a shadow axis that misses the ellipsoid and a phase curve that tangents the
// zero residual at a polar fold. The total band must stay inside the partial band.
func TestMarshalPlanetOccultationNonCentralPartialBandContainsTotalBand(t *testing.T) {
	for _, start := range []time.Time{
		time.Date(2025, time.January, 5, 0, 0, 0, 0, time.UTC),
		time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC),
	} {
		t.Run(start.Format("2006-01-02"), func(t *testing.T) {
			paths, err := moon.FindPlanetOccultationPaths(
				start, start.Add(24*time.Hour), moon.OccultationNeptune, moon.OccultationPathOptions{},
			)
			if err != nil || len(paths) != 1 {
				t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
			}
			data, err := geojson.MarshalPlanetOccultation(paths[0])
			if err != nil {
				t.Fatalf("MarshalPlanetOccultation: %v", err)
			}
			collection := decodeCollection(t, data)
			partialBand := featureWithRole(t, collection, "partial-band")
			if authoritative, ok := partialBand.Properties["static_band_authoritative"].(bool); !ok || !authoritative {
				t.Fatalf("partial-band source=%v authoritative=%v, want the analytic authoritative band",
					partialBand.Properties["source"], partialBand.Properties["static_band_authoritative"])
			}
			partialRings := geoJSONMultiPolygonOuterRings(t, partialBand)
			totalRings := geoJSONMultiPolygonOuterRings(t, featureWithRole(t, collection, "total-band"))
			if len(totalRings) == 0 {
				t.Fatal("total-band geometry is missing")
			}
			if miss := geodata.SphericalPolygonsPathMissDistanceKM(partialRings, totalRings, true); miss > 1 {
				t.Fatalf("total-band escapes partial-band by %.1f km", miss)
			}
		})
	}
}

// TestMarshalPlanetOccultationDenseModeBandKeepsCompactBandFlagFalse 固定默认（密集瞬时足迹）
// 模式的掩带来源契约：几何由解析接触/相位网络给出（static_band_authoritative），而
// compact_band 只表达调用方是否请求了紧凑掩带模式，因此默认模式下必须仍为 false。
// TestMarshalPlanetOccultationDenseModeBandKeepsCompactBandFlagFalse pins the dense-footprint
// default: the geometry comes from the analytic contact/phase network, while compact_band only
// reports whether the caller requested compact-band mode and therefore stays false.
func TestMarshalPlanetOccultationDenseModeBandKeepsCompactBandFlagFalse(t *testing.T) {
	start := time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC)
	paths, err := moon.FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), moon.OccultationNeptune, moon.OccultationPathOptions{},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	if len(paths[0].PartialFootprints) == 0 || len(paths[0].PartialBandFootprints) != 0 {
		t.Fatalf("dense mode footprints=%d bandFootprints=%d, want the dense domain only",
			len(paths[0].PartialFootprints), len(paths[0].PartialBandFootprints))
	}
	data, err := geojson.MarshalPlanetOccultation(paths[0])
	if err != nil {
		t.Fatalf("MarshalPlanetOccultation: %v", err)
	}
	collection := decodeCollection(t, data)
	partialBand := featureWithRole(t, collection, "partial-band")
	if compact, ok := partialBand.Properties["compact_band"].(bool); !ok || compact {
		t.Fatalf("partial-band compact_band=%v, want false because compact mode was not requested",
			partialBand.Properties["compact_band"])
	}
	if authoritative, ok := partialBand.Properties["static_band_authoritative"].(bool); !ok || !authoritative {
		t.Fatalf("partial-band source=%v authoritative=%v, want the analytic authoritative band",
			partialBand.Properties["source"], partialBand.Properties["static_band_authoritative"])
	}
}

// TestMarshalPlanetOccultationCentralEventIsStable ensures that the analytic
// fallback selection does not introduce nondeterministic output for an ordinary
// central event. Repeated marshaling of the same computed path must be byte stable.
func TestMarshalPlanetOccultationCentralEventIsStable(t *testing.T) {
	start := time.Date(2024, time.July, 25, 0, 0, 0, 0, time.FixedZone("UTC+8", 8*60*60))
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
	first, err := geojson.MarshalPlanetOccultation(paths[0])
	if err != nil {
		t.Fatalf("first MarshalPlanetOccultation: %v", err)
	}
	second, err := geojson.MarshalPlanetOccultation(paths[0])
	if err != nil {
		t.Fatalf("second MarshalPlanetOccultation: %v", err)
	}
	if string(first) != string(second) {
		t.Fatal("repeated central-event GeoJSON marshaling is not byte stable")
	}
}

// TestMarshalPlanetOccultationVisibilityBoundaryStaysInsidePartialBand 是可见性边界必须落在掩带
// 内的回归：密集模式下偏掩带曾被纯足迹扫掠截断，月升可见性边界因此越出掩带。两种采样步长都要
// 满足该不变式，否则紫色相位曲线会画在掩带之外。
// TestMarshalPlanetOccultationVisibilityBoundaryStaysInsidePartialBand is the regression that
// keeps the visibility boundary inside the band: in dense mode the partial band used to be
// truncated by a pure footprint sweep and the moonrise boundary escaped it. Both sampling steps
// must satisfy the invariant, otherwise the phase curves are drawn outside the band.
func TestMarshalPlanetOccultationVisibilityBoundaryStaysInsidePartialBand(t *testing.T) {
	const maximumMissKM = 2.0
	for _, test := range []struct {
		name  string
		start time.Time
		opts  moon.OccultationPathOptions
	}{
		{
			name:  "2025-01-05-dense",
			start: time.Date(2025, time.January, 5, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "2025-02-01-dense",
			start: time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "2025-02-01-compact",
			start: time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC),
			opts: moon.OccultationPathOptions{
				Step: 20 * time.Minute, TargetSpacingKM: 900,
				DisableFootprints: true, RiseSetStep: time.Minute,
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			paths, err := moon.FindPlanetOccultationPaths(
				test.start, test.start.Add(24*time.Hour), moon.OccultationNeptune, test.opts,
			)
			if err != nil || len(paths) != 1 {
				t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
			}
			data, err := geojson.MarshalPlanetOccultation(paths[0])
			if err != nil {
				t.Fatalf("MarshalPlanetOccultation: %v", err)
			}
			collection := decodeCollection(t, data)
			partialRings := geoJSONMultiPolygonOuterRings(t, featureWithRole(t, collection, "partial-band"))
			boundaryPaths := make([][]geodata.GeoPoint, 0)
			for _, feature := range featuresWithRole(collection, "visibility-boundary") {
				boundaryPaths = append(boundaryPaths, geoJSONLineStringPaths(t, feature)...)
			}
			if len(boundaryPaths) == 0 {
				t.Fatal("GeoJSON is missing the visibility-boundary phase curves")
			}
			if miss := geodata.SphericalPolygonsPathMissDistanceKM(partialRings, boundaryPaths, false); miss > maximumMissKM {
				t.Fatalf("visibility-boundary escapes partial-band by %.3f km, want <= %.1f km", miss, maximumMissKM)
			}
		})
	}
}

// TestMarshalPlanetOccultationHorizonConnectorsFollowBandFootprintSource 固定连接线与掩带同源的
// 契约：紧凑模式下两者都由紧凑足迹给出，连接线存在；默认（密集瞬时足迹）模式下瞬时足迹没有
// 可与相位端点配对的开放地平边界，因此不产生连接线。
func TestMarshalPlanetOccultationHorizonConnectorsFollowBandFootprintSource(t *testing.T) {
	start := time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		name        string
		opts        moon.OccultationPathOptions
		wantConnect int
	}{
		{
			name: "compact",
			opts: moon.OccultationPathOptions{
				Step: 20 * time.Minute, TargetSpacingKM: 900,
				DisableFootprints: true, RiseSetStep: time.Minute,
			},
			wantConnect: 1,
		},
		{name: "dense", opts: moon.OccultationPathOptions{}, wantConnect: 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			paths, err := moon.FindPlanetOccultationPaths(
				start, start.Add(24*time.Hour), moon.OccultationNeptune, test.opts,
			)
			if err != nil || len(paths) != 1 {
				t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
			}
			data, err := geojson.MarshalPlanetOccultation(paths[0])
			if err != nil {
				t.Fatalf("MarshalPlanetOccultation: %v", err)
			}
			collection := decodeCollection(t, data)
			connectors := featuresWithRole(collection, "horizon-connector")
			if test.wantConnect == 0 && len(connectors) != 0 {
				t.Fatalf("horizon-connector count=%d, want none in dense mode", len(connectors))
			}
			if test.wantConnect > 0 && len(connectors) == 0 {
				t.Fatal("horizon-connector count=0, want the compact mode closures")
			}
		})
	}
}

func geoJSONLineStringPaths(t *testing.T, feature decodedFeature) [][]geodata.GeoPoint {
	t.Helper()
	role := feature.Properties["role"]
	var lines [][][]float64
	switch feature.Geometry.Type {
	case "LineString":
		var line [][]float64
		if err := json.Unmarshal(feature.Geometry.Coordinates, &line); err != nil {
			t.Fatalf("decode %s coordinates: %v", role, err)
		}
		lines = [][][]float64{line}
	case "MultiLineString":
		if err := json.Unmarshal(feature.Geometry.Coordinates, &lines); err != nil {
			t.Fatalf("decode %s coordinates: %v", role, err)
		}
	default:
		t.Fatalf("%s geometry=%q, want LineString or MultiLineString", role, feature.Geometry.Type)
	}
	paths := make([][]geodata.GeoPoint, 0, len(lines))
	for lineIndex, line := range lines {
		path := make([]geodata.GeoPoint, len(line))
		for pointIndex, point := range line {
			if len(point) < 2 {
				t.Fatalf("%s line %d point %d is malformed", role, lineIndex, pointIndex)
			}
			path[pointIndex] = geodata.GeoPoint{Longitude: point[0], Latitude: point[1]}
		}
		paths = append(paths, path)
	}
	return paths
}
