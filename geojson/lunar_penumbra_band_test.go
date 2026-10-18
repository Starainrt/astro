package geojson_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
)

// 两条"仅见半影"带的判据：月落侧 = 食始在地平上、本影阶段整段在地平下、食终也在地平下；
// 月出侧 = 食终在地平上、本影阶段整段在地平下、食始也在地平下。本影可见性取整个区间的最大高度，
// 只比 U1/U4 会漏掉极区掠射时两刻之间的短暂窗口。用高度独立复算，并允许边界附近半格误差。
func TestLunarGeoJSONPenumbraBandsFollowUmbralContacts(t *testing.T) {
	if testing.Short() {
		t.Skip("slow regression")
	}
	date := time.Date(2029, 1, 1, 0, 0, 0, 0, time.FixedZone("CST", 8*3600))
	info, ok := eclipse.LunarEclipseOnDate(date)
	if !ok || !info.HasPartial {
		t.Fatalf("missing umbral lunar eclipse")
	}
	data, err := geojson.MarshalLunarEclipse(info, 360)
	if err != nil {
		t.Fatal(err)
	}
	collection := decodeCollection(t, data)
	moonset := featureWithRole(t, collection, "penumbra-moonset")
	moonrise := featureWithRole(t, collection, "penumbra-moonrise")
	containsMoonset := decodeMultiPolygonContains(t, moonset.Geometry)
	containsMoonrise := decodeMultiPolygonContains(t, moonrise.Geometry)
	states := make(map[float64]basic.MoonState)
	altitude := func(at time.Time, longitude, latitude float64) float64 {
		jd := basic.Date2JD(at.UTC())
		state, ok := states[jd]
		if !ok {
			state = basic.MoonStateAt(jd)
			states[jd] = state
		}
		return state.HMoonHeight(longitude, latitude)
	}
	umbralBelowHorizon := func(longitude, latitude float64) bool {
		step := info.PartialEnd.Sub(info.PartialStart) / 400
		for at := info.PartialStart; !at.After(info.PartialEnd); at = at.Add(step) {
			// 判据只关心本影阶段是否整段在地平下，首次 >=0 即可停止。
			if altitude(at, longitude, latitude) >= 0 {
				return false
			}
		}
		return true
	}
	checked := map[string]int{}
	for longitude := -179.5; longitude < 180; longitude += 3 {
		for latitude := -89.5; latitude < 90; latitude += 3 {
			heights := []float64{
				altitude(info.PenumbralStart, longitude, latitude),
				altitude(info.PartialStart, longitude, latitude),
				altitude(info.PartialEnd, longitude, latitude),
				altitude(info.PenumbralEnd, longitude, latitude),
			}
			// 区域边缘的半格误差：任一条判据高度落在 ±0.5° 内就不作断言。
			margin := 0.5
			for _, height := range heights {
				if height < 0 {
					height = -height
				}
				if height < margin {
					margin = height
				}
			}
			if margin < 0.5 {
				continue
			}
			umbralBelow := umbralBelowHorizon(longitude, latitude)
			for _, role := range []string{"penumbra-moonset", "penumbra-moonrise"} {
				want := false
				if role == "penumbra-moonset" {
					want = heights[0] > 0 && umbralBelow && heights[3] < 0
				} else {
					want = heights[3] > 0 && umbralBelow && heights[0] < 0
				}
				got := containsMoonset(longitude, latitude)
				if role == "penumbra-moonrise" {
					got = containsMoonrise(longitude, latitude)
				}
				if got != want {
					t.Fatalf("%s (%.1f,%.1f): geometry=%v altitude criterion=%v (P1=%.2f U1=%.2f U4=%.2f P4=%.2f)",
						role, longitude, latitude, got, want, heights[0], heights[1], heights[2], heights[3])
				}
				if want {
					checked[role]++
				}
			}
		}
	}
	if checked["penumbra-moonset"] < 20 || checked["penumbra-moonrise"] < 20 {
		t.Fatalf("too few grid points verified: %v", checked)
	}
}

func decodeMultiPolygonContains(t *testing.T, geometry struct {
	Type        string          `json:"type"`
	Coordinates json.RawMessage `json:"coordinates"`
	Geometries  json.RawMessage `json:"geometries"`
}) func(float64, float64) bool {
	t.Helper()
	switch geometry.Type {
	case "MultiPolygon":
		var polygons [][][][]float64
		if err := json.Unmarshal(geometry.Coordinates, &polygons); err != nil {
			t.Fatalf("decode MultiPolygon coordinates: %v", err)
		}
		return func(longitude, latitude float64) bool {
			return geoJSONMultiPolygonContains(polygons, longitude, latitude)
		}
	case "Polygon":
		var polygon [][][]float64
		if err := json.Unmarshal(geometry.Coordinates, &polygon); err != nil {
			t.Fatalf("decode Polygon coordinates: %v", err)
		}
		if len(polygon) == 0 {
			return func(float64, float64) bool { return false }
		}
		return func(longitude, latitude float64) bool {
			return geoJSONRingContains(polygon[0], longitude, latitude)
		}
	default:
		t.Fatalf("geometry type=%q, want Polygon or MultiPolygon", geometry.Type)
	}
	return nil
}

// 两条带必须与本地可见性类别的细分一致，且可被 SkipRoles 关掉；纯半影月食没有本影接触，不画带。
func TestLunarGeoJSONPenumbraBandsMatchVisibilityClasses(t *testing.T) {
	date := time.Date(2029, 1, 1, 0, 0, 0, 0, time.FixedZone("CST", 8*3600))
	info, _ := eclipse.LunarEclipseOnDate(date)
	data, err := geojson.MarshalLunarEclipse(info, 360)
	if err != nil {
		t.Fatal(err)
	}
	collection := decodeCollection(t, data)
	for _, witness := range []struct {
		role     string
		lon, lat float64
		class    eclipse.LocalLunarEclipseVisibility
	}{
		{"penumbra-moonset", -179, -61, eclipse.LocalLunarEclipsePenumbraMoonset},
		{"penumbra-moonrise", -69, 61, eclipse.LocalLunarEclipsePenumbraMoonrise},
	} {
		local, ok := eclipse.GeometricLocalLunarEclipseOnDate(date, witness.lon, witness.lat, 0)
		if !ok || local.Visibility != witness.class {
			t.Fatalf("(%.0f,%.0f) visibility=%q ok=%v, want %q", witness.lon, witness.lat, local.Visibility, ok, witness.class)
		}
		if !geometryContainsPoint(t, featureWithRole(t, collection, witness.role).Geometry, witness.lon, witness.lat) {
			t.Fatalf("%s does not cover its witness point", witness.role)
		}
	}

	skipped, err := geojson.MarshalLunarEclipseWithOptions(info, 360, geojson.LunarEclipseOptions{
		SkipRoles: []string{"penumbra-moonset", "penumbra-moonrise"},
	})
	if err != nil {
		t.Fatal(err)
	}
	skippedCollection := decodeCollection(t, skipped)
	for _, role := range []string{"penumbra-moonset", "penumbra-moonrise"} {
		if len(featuresWithRole(skippedCollection, role)) != 0 {
			t.Fatalf("%s survived SkipRoles", role)
		}
	}
	assertRoles(t, skippedCollection, "visible-at-p1", "visible-at-p4")

	penumbralDate := time.Date(2020, 1, 11, 0, 0, 0, 0, time.FixedZone("CST", 8*3600))
	penumbral, ok := eclipse.LunarEclipseOnDate(penumbralDate)
	if !ok || penumbral.HasPartial {
		t.Fatalf("expected a purely penumbral eclipse, HasPartial=%v ok=%v", penumbral.HasPartial, ok)
	}
	penumbralData, err := geojson.MarshalLunarEclipse(penumbral, 360)
	if err != nil {
		t.Fatal(err)
	}
	penumbralCollection := decodeCollection(t, penumbralData)
	for _, role := range []string{"penumbra-moonset", "penumbra-moonrise"} {
		if len(featuresWithRole(penumbralCollection, role)) != 0 {
			t.Fatalf("purely penumbral eclipse must not export %s", role)
		}
	}
}

// 极区掠射见证：本影区间中途再短暂也算见到本影，不能算进"仅见半影"；本影整段在地平下才进带。
func TestLunarGeoJSONPenumbraBandsExcludeUmbralGrazingWindow(t *testing.T) {
	date := time.Date(1932, 3, 22, 0, 0, 0, 0, time.UTC)
	info, ok := eclipse.LunarEclipseOnDate(date)
	if !ok || !info.HasPartial {
		t.Fatal("missing umbral lunar eclipse")
	}
	data, err := geojson.MarshalLunarEclipse(info, 360)
	if err != nil {
		t.Fatal(err)
	}
	collection := decodeCollection(t, data)
	moonset := featureWithRole(t, collection, "penumbra-moonset")
	moonrise := featureWithRole(t, collection, "penumbra-moonrise")
	for _, witness := range []struct {
		name     string
		lon, lat float64
		want     eclipse.LocalLunarEclipseVisibility
	}{
		{"本影窗口 +0.0006 度", -91.8, -88.7675, eclipse.LocalLunarEclipseMoonrise},
		{"U1 已在地平上", -91.75, -88.75, eclipse.LocalLunarEclipseMoonrise},
		{"U1 高度 +0.21 度（陡边界）", -75.5, 6.5, eclipse.LocalLunarEclipseMoonset},
		{"本影极值 -2.25 度（宽余量）", -68.75, -70.15, eclipse.LocalLunarEclipsePenumbraMoonset},
	} {
		local, found := eclipse.GeometricLocalLunarEclipseOnDate(date, witness.lon, witness.lat, 0)
		if !found || local.Visibility != witness.want {
			t.Fatalf("%s: 站点 API visibility=%q ok=%v, want %q", witness.name, local.Visibility, found, witness.want)
		}
		inMoonset := geometryContainsPoint(t, moonset.Geometry, witness.lon, witness.lat)
		inMoonrise := geometryContainsPoint(t, moonrise.Geometry, witness.lon, witness.lat)
		switch witness.want {
		case eclipse.LocalLunarEclipsePenumbraMoonset:
			if !inMoonset || inMoonrise {
				t.Fatalf("%s: 半影月落带包含=%v 半影月出带包含=%v，want 只在前者", witness.name, inMoonset, inMoonrise)
			}
		default:
			if inMoonset || inMoonrise {
				t.Fatalf("%s: 见了本影却被算进半影带（月落=%v 月出=%v）", witness.name, inMoonset, inMoonrise)
			}
		}
	}
}
