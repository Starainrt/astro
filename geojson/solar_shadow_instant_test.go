package geojson_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
)

// 契约：单时刻本影导出只含该时刻的 region 与被切断时的物理边界；同一时刻的几何必须与
// 整包采样的对应 feature 完全一致，前端的"插值→精确替换"才不会跳变。
// Contract: the single-instant export contains only that instant's region (and the
// physical boundary when the horizon cuts it), and its geometry must match the packaged
// sample at the same time exactly.

func TestMarshalSolarEclipseShadowInstantMatchesPackagedFeatures(t *testing.T) {
	date := time.Date(2009, time.July, 22, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
		Step: 2 * time.Minute, BoundaryPoints: 96, CentralShadowStep: 2 * time.Minute,
		DisableRiseSet: true,
	})
	if !ok {
		t.Fatal("expected the 2009-07-22 eclipse")
	}
	packaged := decodeCollection(t, mustMarshalSolarEclipse(t, partial))
	regions := featuresWithRole(packaged, "central-shadow-footprint")
	boundaries := map[string]decodedFeature{}
	for _, feature := range featuresWithRole(packaged, "central-shadow-boundary") {
		stamp, _ := feature.Properties["time"].(string)
		boundaries[stamp] = feature
	}
	if len(regions) == 0 {
		t.Fatal("packaged export has no central-shadow-footprint")
	}
	comparedOpen := 0
	for _, index := range []int{0, len(regions) / 2, len(regions) - 1} {
		feature := regions[index]
		stamp, _ := feature.Properties["time"].(string)
		value, err := time.Parse(time.RFC3339Nano, stamp)
		if err != nil {
			t.Fatalf("parse packaged time %q: %v", stamp, err)
		}
		instant, ok := eclipse.SolarEclipseShadowAt(value)
		if !ok {
			t.Fatalf("%s: single-instant call reported no umbra but the package has a feature", stamp)
		}
		data, err := geojson.MarshalSolarEclipseShadowInstant(instant)
		if err != nil {
			t.Fatalf("%s: MarshalSolarEclipseShadowInstant: %v", stamp, err)
		}
		single := decodeCollection(t, data)
		if len(single.Features) == 0 {
			t.Fatalf("%s: single-instant export has no features", stamp)
		}
		instantRegion := featureWithRole(t, single, "central-shadow-footprint")
		// 时间戳经 time.Time 往返会有约 1e-9 日的舍入，几何容差取 1e-7 度（约 1 厘米）。
		assertSameCoordinates(t, stamp+" region", feature.Geometry.Coordinates, instantRegion.Geometry.Coordinates)
		// 整包 feature 与单时刻导出的插值签名必须由同一口径给出，前端才能用同一个键比较。
		if feature.Properties["interp_signature"] != instantRegion.Properties["interp_signature"] {
			t.Fatalf("%s: packaged interp_signature=%v, instant=%v",
				stamp, feature.Properties["interp_signature"], instantRegion.Properties["interp_signature"])
		}
		if feature.Properties["interp_signature"] != instant.Topology.Signature() {
			t.Fatalf("%s: packaged interp_signature=%v, topology=%q",
				stamp, feature.Properties["interp_signature"], instant.Topology.Signature())
		}
		closed, _ := feature.Properties["source_boundary_closed"].(bool)
		if closed {
			if len(single.Features) != 1 {
				t.Fatalf("%s: self-closed instant exported %d features, want only the region", stamp, len(single.Features))
			}
			continue
		}
		comparedOpen++
		packagedBoundary, present := boundaries[stamp]
		if !present {
			t.Fatalf("%s: packaged export has no boundary feature to compare", stamp)
		}
		instantBoundary := featureWithRole(t, single, "central-shadow-boundary")
		assertSameCoordinates(t, stamp+" boundary", packagedBoundary.Geometry.Coordinates, instantBoundary.Geometry.Coordinates)
		for _, key := range []string{"closure", "interp_signature", "delta_t_seconds", "geometry_role"} {
			if _, present := instantRegion.Properties[key]; !present {
				t.Fatalf("%s: instant region is missing property %q", stamp, key)
			}
		}
	}
	if comparedOpen == 0 {
		t.Fatal("no horizon-cut instant was compared against the packaged boundary")
	}
}

func TestMarshalSolarEclipseShadowInstantEmpty(t *testing.T) {
	instant, ok := eclipse.SolarEclipseShadowAt(time.Date(2009, time.July, 22, 12, 0, 0, 0, time.UTC))
	if ok || !instant.Empty() {
		t.Fatalf("expected an empty instant, got ok=%v empty=%v", ok, instant.Empty())
	}
	data, err := geojson.MarshalSolarEclipseShadowInstant(instant)
	if err != nil {
		t.Fatalf("empty instant returned an error: %v", err)
	}
	var decoded struct {
		Type     string        `json:"type"`
		Features []interface{} `json:"features"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("decode empty export: %v", err)
	}
	if decoded.Type != "FeatureCollection" || len(decoded.Features) != 0 {
		t.Fatalf("empty export shape: type=%q features=%d", decoded.Type, len(decoded.Features))
	}
}

func TestMarshalSolarEclipseShadowInstantProperties(t *testing.T) {
	value := time.Date(2009, time.July, 22, 4, 19, 23, 378387689, time.UTC)
	instant, ok := eclipse.SolarEclipseShadowAt(value)
	if !ok {
		t.Fatal("expected an umbral footprint at the test instant")
	}
	data, err := geojson.MarshalSolarEclipseShadowInstant(instant)
	if err != nil {
		t.Fatalf("MarshalSolarEclipseShadowInstant: %v", err)
	}
	collection := decodeCollection(t, data)
	if len(collection.Features) != 2 {
		t.Fatalf("features=%d, want region + boundary", len(collection.Features))
	}
	region := featureWithRole(t, collection, "central-shadow-footprint")
	if region.Geometry.Type != "MultiPolygon" || region.Properties["geometry_role"] != "horizon-closed-region" {
		t.Fatalf("region shape: type=%v role=%v", region.Geometry.Type, region.Properties["geometry_role"])
	}
	if region.Properties["source_boundary_closed"] != false {
		t.Fatalf("region source_boundary_closed=%v, want false", region.Properties["source_boundary_closed"])
	}
	closure, ok := region.Properties["closure"].(map[string]interface{})
	if !ok || closure["kind"] != "horizon" {
		t.Fatalf("closure=%v, want kind=horizon", region.Properties["closure"])
	}
	if signature, _ := region.Properties["interp_signature"].(string); signature != instant.Topology.Signature() {
		t.Fatalf("interp_signature=%v, want %q", region.Properties["interp_signature"], instant.Topology.Signature())
	}
	boundary := featureWithRole(t, collection, "central-shadow-boundary")
	if boundary.Geometry.Type != "MultiLineString" || boundary.Properties["geometry_role"] != "open-boundary" {
		t.Fatalf("boundary shape: type=%v role=%v", boundary.Geometry.Type, boundary.Properties["geometry_role"])
	}
	if boundary.Properties["interp_signature"] != region.Properties["interp_signature"] {
		t.Fatal("region and boundary must share the interpolation signature")
	}
}

// assertSameCoordinates 比较两份几何坐标：结构必须一致，数值容差 1e-7 度。
// assertSameCoordinates compares two coordinate trees with a 1e-7 degree tolerance.
func assertSameCoordinates(t *testing.T, label string, want, got json.RawMessage) {
	t.Helper()
	var wantTree, gotTree interface{}
	if err := json.Unmarshal(want, &wantTree); err != nil {
		t.Fatalf("%s: decode packaged coordinates: %v", label, err)
	}
	if err := json.Unmarshal(got, &gotTree); err != nil {
		t.Fatalf("%s: decode instant coordinates: %v", label, err)
	}
	var compare func(path string, a, b interface{})
	compare = func(path string, a, b interface{}) {
		aList, aOK := a.([]interface{})
		bList, bOK := b.([]interface{})
		if aOK != bOK {
			t.Fatalf("%s%s: structure differs", label, path)
		}
		if !aOK {
			aNumber, aIsNumber := a.(float64)
			bNumber, bIsNumber := b.(float64)
			if !aIsNumber || !bIsNumber {
				return
			}
			if difference := aNumber - bNumber; difference > 1e-7 || difference < -1e-7 {
				t.Fatalf("%s%s: %.12f vs %.12f", label, path, aNumber, bNumber)
			}
			return
		}
		if len(aList) != len(bList) {
			t.Fatalf("%s%s: length %d vs %d", label, path, len(aList), len(bList))
		}
		for index := range aList {
			compare(fmt.Sprintf("%s[%d]", path, index), aList[index], bList[index])
		}
	}
	compare("", wantTree, gotTree)
}

func TestMarshalSolarEclipseShadowInstantSelfClosedShape(t *testing.T) {
	instant, ok := eclipse.SolarEclipseShadowAt(time.Date(2009, time.July, 22, 0, 55, 0, 0, time.UTC))
	if !ok {
		t.Fatal("expected a self-closed footprint")
	}
	if !instant.Closed {
		t.Fatalf("expected closed at the test instant, signature=%s", instant.Topology.Signature())
	}
	data, err := geojson.MarshalSolarEclipseShadowInstant(instant)
	if err != nil {
		t.Fatalf("MarshalSolarEclipseShadowInstant: %v", err)
	}
	collection := decodeCollection(t, data)
	if len(collection.Features) != 1 {
		t.Fatalf("self-closed instant exported %d features, want only the region", len(collection.Features))
	}
	region := collection.Features[0]
	if region.Properties["source_boundary_closed"] != true {
		t.Fatalf("source_boundary_closed=%v, want true", region.Properties["source_boundary_closed"])
	}
	// 与整包导出同构：自身闭合的区域没有 geometry_role，也没有 closure。
	for _, key := range []string{"geometry_role", "closure"} {
		if _, present := region.Properties[key]; present {
			t.Fatalf("self-closed region must not carry %q (packaged export does not)", key)
		}
	}
}

func TestMarshalSolarEclipseShadowInstantPenumbraRoles(t *testing.T) {
	solver := eclipse.NewSolarEclipseShadowSolver(eclipse.SolarEclipseShadowSolverOptions{
		Kind: eclipse.SolarEclipseShadowPenumbra,
	})
	instant, ok := solver.ShadowAt(time.Date(2009, time.July, 22, 0, 52, 0, 0, time.UTC))
	if !ok {
		t.Fatal("expected a penumbra footprint")
	}
	data, err := geojson.MarshalSolarEclipseShadowInstant(instant)
	if err != nil {
		t.Fatalf("MarshalSolarEclipseShadowInstant: %v", err)
	}
	collection := decodeCollection(t, data)
	region := featureWithRole(t, collection, "partial-footprint")
	if region.Geometry.Type != "MultiPolygon" {
		t.Fatalf("penumbra region geometry=%q, want MultiPolygon", region.Geometry.Type)
	}
	if signature, _ := region.Properties["interp_signature"].(string); len(signature) < 8 || signature[:8] != "penumbra" {
		t.Fatalf("penumbra signature=%v, want a penumbra prefix", region.Properties["interp_signature"])
	}
	if _, present := region.Properties["delta_t_seconds"]; !present {
		t.Fatal("penumbra region is missing delta_t_seconds")
	}
	if instant.Closed {
		if len(collection.Features) != 1 {
			t.Fatalf("self-closed penumbra exported %d features, want only the region", len(collection.Features))
		}
		return
	}
	boundary := featureWithRole(t, collection, "partial-footprint-boundary")
	if boundary.Geometry.Type != "MultiLineString" {
		t.Fatalf("penumbra boundary geometry=%q, want MultiLineString", boundary.Geometry.Type)
	}
	if closure, _ := region.Properties["closure"].(map[string]interface{}); closure["kind"] != "horizon" {
		t.Fatalf("penumbra closure=%v, want kind=horizon", region.Properties["closure"])
	}
}

func TestMarshalSampledPartialFootprintCarriesClosureAndHorizonEndpoints(t *testing.T) {
	date := time.Date(2009, time.July, 22, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
		Step: 2 * time.Minute, BoundaryPoints: 96, CentralShadowStep: 2 * time.Minute,
		DisableRiseSet: true,
	})
	if !ok {
		t.Fatal("expected the 2009-07-22 eclipse")
	}
	collection := decodeCollection(t, mustMarshalSolarEclipse(t, partial))
	openCount, closureCount := 0, 0
	for _, feature := range featuresWithRole(collection, "partial-footprint") {
		if _, present := feature.Properties["interp_signature"]; !present {
			t.Fatal("packaged partial footprint is missing interp_signature")
		}
		closed, _ := feature.Properties["source_boundary_closed"].(bool)
		if closed {
			if _, present := feature.Properties["closure"]; present {
				t.Fatal("self-closed partial footprint must not carry a closure")
			}
			continue
		}
		openCount++
		if _, present := feature.Properties["closure"]; present {
			closureCount++
		}
		if feature.Properties["geometry_role"] != "horizon-closed-region" {
			t.Fatalf("open partial footprint geometry_role=%v, want horizon-closed-region",
				feature.Properties["geometry_role"])
		}
	}
	if openCount == 0 || closureCount != openCount {
		t.Fatalf("open=%d withClosure=%d, want every open footprint annotated", openCount, closureCount)
	}
}
