package geojson_test

import (
	"encoding/json"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
)

func TestLunarGeoJSONUsesTopocentricHorizon(t *testing.T) {
	date := time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC)
	info, ok := eclipse.LunarEclipseOnDate(date)
	if !ok {
		t.Fatal("missing lunar eclipse")
	}
	data, err := geojson.MarshalLunarEclipse(info, 360)
	if err != nil {
		t.Fatal(err)
	}
	collection := decodeCollection(t, data)
	for _, contact := range []struct {
		role, horizon string
		at            time.Time
	}{
		{"visible-at-p1", "p1-horizon", info.PenumbralStart},
		{"visible-at-p4", "p4-horizon", info.PenumbralEnd},
	} {
		band := featureWithRole(t, collection, contact.role)
		line := featureWithRole(t, collection, contact.horizon)
		var segments [][][]float64
		if err := json.Unmarshal(line.Geometry.Coordinates, &segments); err != nil {
			t.Fatal(err)
		}
		jd := basic.Date2JD(contact.at.UTC())
		for _, segment := range segments {
			for _, point := range segment {
				if math.Abs(point[0]) == 180 {
					continue
				}
				if altitude := basic.HMoonHeight(jd, point[0], point[1], 0); math.Abs(altitude) > 1e-8 {
					t.Fatalf("horizon altitude=%g", altitude)
				}
			}
		}
		for lon := -175.; lon < 180; lon += 10 {
			for lat := -85.; lat < 90; lat += 10 {
				altitude := basic.HMoonHeight(jd, lon, lat, 0)
				if math.Abs(altitude) < 0.05 {
					continue
				}
				if inside := geometryContainsPoint(t, band.Geometry, lon, lat); inside != (altitude > 0) {
					t.Fatalf("%s point=(%v,%v) inside=%v altitude=%v", contact.role, lon, lat, inside, altitude)
				}
			}
		}
	}
}

// TestLunarGeoJSONEnvelopeCoversPolarLens 固定时间包络的必要性：见证站点不在任何一块 P1/P4 瞬时半球里。
func TestLunarGeoJSONEnvelopeCoversPolarLens(t *testing.T) {
	date := time.Date(2029, 1, 1, 12, 0, 0, 0, time.FixedZone("CST", 8*3600))
	info, ok := eclipse.LunarEclipseOnDate(date)
	if !ok {
		t.Fatal("missing lunar eclipse")
	}
	data, err := geojson.MarshalLunarEclipse(info, 360)
	if err != nil {
		t.Fatal(err)
	}
	collection := decodeCollection(t, data)
	const longitude, latitude = 108.729001, -59.937452

	during := featureWithRole(t, collection, "visible-during-eclipse")
	if !geometryContainsPoint(t, during.Geometry, longitude, latitude) {
		t.Fatal("visible-during-eclipse does not cover the polar lens witness")
	}
	if geometryContainsPoint(t, featureWithRole(t, collection, "visible-throughout-eclipse").Geometry, longitude, latitude) {
		t.Fatal("visible-throughout-eclipse covers a site that loses the penumbral ends")
	}
	for _, role := range []string{"visible-at-p1", "visible-at-p4"} {
		if geometryContainsPoint(t, featureWithRole(t, collection, role).Geometry, longitude, latitude) {
			t.Fatalf("%s unexpectedly covers the polar lens witness", role)
		}
	}

	local, localOK := eclipse.LocalLunarEclipseOnDate(date, longitude, latitude, 0)
	if !localOK || local.Visibility != eclipse.LocalLunarEclipseRiseAndSet {
		t.Fatalf("local visibility=%q ok=%v, want %q", local.Visibility, localOK, eclipse.LocalLunarEclipseRiseAndSet)
	}
}

// TestLunarGeoJSONEnvelopesMatchAltitudeExtrema 用高度极值独立判据钉住两个时间包络；采样取整点经度，
// 并跳过零高度 0.05° 以内的边界点，1° 经度采样在区域边缘的半格误差不算失配。
func TestLunarGeoJSONEnvelopesMatchAltitudeExtrema(t *testing.T) {
	for _, day := range []string{
		"0275-09-22", "0386-09-24", "1076-09-15", "1904-09-24", "2396-03-25",
		"2779-03-24", "2955-09-23", "3188-09-27", "3738-03-19", "4026-03-16",
	} {
		t.Run(day, func(t *testing.T) {
			date, err := time.Parse("2006-01-02", day)
			if err != nil {
				t.Fatal(err)
			}
			info, ok := eclipse.LunarEclipseOnDate(date)
			if !ok {
				t.Fatal("missing lunar eclipse")
			}
			data, err := geojson.MarshalLunarEclipse(info, 360)
			if err != nil {
				t.Fatal(err)
			}
			collection := decodeCollection(t, data)
			during := featureWithRole(t, collection, "visible-during-eclipse")
			throughout := featureWithRole(t, collection, "visible-throughout-eclipse")
			jdStart := basic.Date2JD(info.PenumbralStart.UTC())
			jdEnd := basic.Date2JD(info.PenumbralEnd.UTC())
			for _, base := range []float64{-88, 87.5} {
				for longitude := -176.0; longitude < 180; longitude += 8 {
					for latitude := base; latitude <= base+2.5; latitude += 0.5 {
						maximum, minimum := lunarAltitudeExtrema(jdStart, jdEnd, longitude, latitude)
						if math.Abs(maximum) > 0.05 &&
							geometryContainsPoint(t, during.Geometry, longitude, latitude) != (maximum > 0) {
							t.Fatalf("visible-during-eclipse (%v,%v) maximum=%g", longitude, latitude, maximum)
						}
						if math.Abs(minimum) > 0.05 &&
							geometryContainsPoint(t, throughout.Geometry, longitude, latitude) != (minimum > 0) {
							t.Fatalf("visible-throughout-eclipse (%v,%v) minimum=%g", longitude, latitude, minimum)
						}
					}
				}
			}
		})
	}
}

func lunarAltitudeExtrema(jdStart, jdEnd, longitude, latitude float64) (float64, float64) {
	const samples = 48
	maximum, minimum := math.Inf(-1), math.Inf(1)
	for index := 0; index <= samples; index++ {
		altitude := basic.HMoonHeight(jdStart+(jdEnd-jdStart)*float64(index)/samples, longitude, latitude, 0)
		maximum = math.Max(maximum, altitude)
		minimum = math.Min(minimum, altitude)
	}
	return maximum, minimum
}

func TestLunarGeoJSONTimeEnvelopesCoverPolarWindow(t *testing.T) {
	date := time.Date(1800, 4, 9, 0, 0, 0, 0, time.UTC)
	info, ok := eclipse.LunarEclipseOnDate(date)
	if !ok {
		t.Fatal("missing lunar eclipse")
	}
	data, err := geojson.MarshalLunarEclipse(info, 360)
	if err != nil {
		t.Fatal(err)
	}
	collection := decodeCollection(t, data)
	during := featureWithRole(t, collection, "visible-during-eclipse")
	throughout := featureWithRole(t, collection, "visible-throughout-eclipse")
	if during.Properties["aggregation"] != "union" || throughout.Properties["aggregation"] != "intersection" {
		t.Fatalf("unexpected aggregations: during=%v throughout=%v", during.Properties["aggregation"], throughout.Properties["aggregation"])
	}
	if !geometryContainsPoint(t, during.Geometry, 121, 82) {
		t.Fatal("visible-during-eclipse misses a short polar visibility interval")
	}
	if geometryContainsPoint(t, throughout.Geometry, 121, 82) {
		t.Fatal("visible-throughout-eclipse contains a rise-and-set site")
	}
	if !geometryContainsPoint(t, during.Geometry, -69, -84) {
		t.Fatal("visible-during-eclipse misses the interrupted-site witness")
	}
	if geometryContainsPoint(t, throughout.Geometry, -69, -84) {
		t.Fatal("visible-throughout-eclipse contains an interrupted site")
	}
}

func TestLunarGeoJSON19040924DoesNotFillFalseSouthPolarCap(t *testing.T) {
	date := time.Date(1904, 9, 24, 0, 0, 0, 0, time.UTC)
	info, ok := eclipse.LunarEclipseOnDate(date)
	if !ok {
		t.Fatal("missing lunar eclipse")
	}
	data, err := geojson.MarshalLunarEclipse(info, 360)
	if err != nil {
		t.Fatal(err)
	}
	collection := decodeCollection(t, data)
	band := featureWithRole(t, collection, "visible-at-p1")
	point := struct {
		longitude float64
		latitude  float64
	}{-115, -89.9}
	if altitude := basic.HMoonHeight(basic.Date2JD(info.PenumbralStart), point.longitude, point.latitude, 0); altitude >= -0.01 {
		t.Fatalf("regression witness altitude=%g, want below horizon", altitude)
	}
	if geometryContainsPoint(t, band.Geometry, point.longitude, point.latitude) {
		t.Fatalf("visible-at-p1 contains below-horizon polar witness %+v", point)
	}
}

func TestLunarGeoJSONPolarVisibilityGrid(t *testing.T) {
	for _, day := range []string{
		"0275-09-22", "0386-09-24", "1076-09-15", "1904-09-24", "2396-03-25",
		"2779-03-24", "2955-09-23", "3188-09-27", "3738-03-19", "4026-03-16",
	} {
		t.Run(day, func(t *testing.T) {
			date, err := time.Parse("2006-01-02", day)
			if err != nil {
				t.Fatal(err)
			}
			info, ok := eclipse.LunarEclipseOnDate(date)
			if !ok {
				t.Fatal("missing lunar eclipse")
			}
			counts := []int{360}
			if day == "1904-09-24" {
				counts = []int{12, 96, 360, 1440}
			}
			for _, count := range counts {
				t.Run(fmt.Sprint(count), func(t *testing.T) {
					assertLunarVisibilityGrid(t, info, count)
				})
			}
		})
	}
}

func assertLunarVisibilityGrid(t *testing.T, info eclipse.LunarEclipseInfo, count int) {
	t.Helper()
	data, err := geojson.MarshalLunarEclipse(info, count)
	if err != nil {
		t.Fatal(err)
	}
	collection := decodeCollection(t, data)
	latitudes := []float64{-89.999999, -89.999, -89.99, -89.9, -89.5, 89.5, 89.9, 89.99, 89.999, 89.999999}
	for lat := -89.0; lat <= 89; lat += 2 {
		latitudes = append(latitudes, lat)
	}
	for _, contact := range []struct {
		role string
		at   time.Time
	}{
		{"visible-at-p1", info.PenumbralStart},
		{"visible-at-p4", info.PenumbralEnd},
	} {
		band := featureWithRole(t, collection, contact.role)
		var polygons [][][][]float64
		if err := json.Unmarshal(band.Geometry.Coordinates, &polygons); err != nil {
			t.Fatal(err)
		}
		jd := basic.Date2JD(contact.at.UTC())
		visible, invisible := 0, 0
		for _, lat := range latitudes {
			for lon := -179.5; lon < 180; lon += 5 {
				altitude := basic.HMoonHeight(jd, lon, lat, 0)
				tolerance := 0.003
				if math.Abs(lat) > 89.99 {
					tolerance = 1e-5
				}
				if math.Abs(altitude) <= tolerance {
					continue
				}
				inside := geoJSONMultiPolygonContains(polygons, lon, lat)
				if inside != (altitude > 0) {
					t.Fatalf("%s point=(%g,%g) inside=%v altitude=%g", contact.role, lon, lat, inside, altitude)
				}
				if inside {
					visible++
				} else {
					invisible++
				}
			}
		}
		if visible == 0 || invisible == 0 {
			t.Fatalf("%s grid must exercise both sides: visible=%d invisible=%d", contact.role, visible, invisible)
		}
	}
}

func BenchmarkLunarEclipseGeoJSON(b *testing.B) {
	for _, day := range []string{"1904-09-24", "2026-03-03", "4026-03-16"} {
		date, err := time.Parse("2006-01-02", day)
		if err != nil {
			b.Fatal(err)
		}
		info, ok := eclipse.LunarEclipseOnDate(date)
		if !ok {
			b.Fatal("missing lunar eclipse")
		}
		b.Run(day, func(b *testing.B) {
			b.ReportAllocs()
			for iteration := 0; iteration < b.N; iteration++ {
				if _, err := geojson.MarshalLunarEclipse(info, 360); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// TestLunarGeoJSONEnvelopeBoundariesMatchDenseTimeSweep 用 1000 点密集时间求极值作为连续时间真值，
// 核对两个包络的边界纬度：包络按 48 个时刻离散采样，边界处的极值是掠射型，误差必须远小于经度列距。
func TestLunarGeoJSONEnvelopeBoundariesMatchDenseTimeSweep(t *testing.T) {
	for _, testCase := range []struct {
		day       string
		longitude float64
		starts    []float64
	}{
		{"2029-01-01", -150, []float64{-20, 30}},
		{"1904-09-24", -170, []float64{0, 40}},
	} {
		t.Run(testCase.day, func(t *testing.T) {
			location := time.UTC
			if testCase.day == "2029-01-01" {
				location = time.FixedZone("CST", 8*3600)
			}
			date, err := time.ParseInLocation("2006-01-02", testCase.day, location)
			if err != nil {
				t.Fatal(err)
			}
			info, ok := eclipse.LunarEclipseOnDate(date)
			if !ok {
				t.Fatal("missing lunar eclipse")
			}
			data, err := geojson.MarshalLunarEclipse(info, 360)
			if err != nil {
				t.Fatal(err)
			}
			collection := decodeCollection(t, data)
			jdStart := basic.Date2JD(info.PenumbralStart.UTC())
			jdEnd := basic.Date2JD(info.PenumbralEnd.UTC())
			for _, role := range []struct {
				name    string
				maximum bool
			}{
				{"visible-during-eclipse", true},
				{"visible-throughout-eclipse", false},
			} {
				feature := featureWithRole(t, collection, role.name)
				for _, start := range testCase.starts {
					for _, limit := range []float64{90, -90} {
						got, hasPolygonEdge := lunarEnvelopeBoundaryLatitude(t, feature, testCase.longitude, start, limit)
						want, hasTrueEdge := lunarDenseExtremumBoundaryLatitude(
							t, jdStart, jdEnd, testCase.longitude, start, limit, role.maximum,
						)
						if hasPolygonEdge != hasTrueEdge {
							t.Fatalf("%s 经度 %v 起点 %v 朝 %v：包络有边界=%v，密集时间真值有边界=%v",
								role.name, testCase.longitude, start, limit, hasPolygonEdge, hasTrueEdge)
						}
						if !hasPolygonEdge {
							continue
						}
						if difference := math.Abs(got - want); difference > 0.01 {
							t.Fatalf("%s 经度 %v 起点 %v 朝 %v：包络边界 %.4f，密集时间真值 %.4f，差 %.4f",
								role.name, testCase.longitude, start, limit, got, want, difference)
						}
					}
				}
			}
		})
	}
}

// TestLunarGeoJSONEnvelopesCrossAntimeridian 固定包络在 ±180° 的连续性：两侧同纬度必须同号，
// 且日界线拆分后的碎片仍覆盖该经度。
func TestLunarGeoJSONEnvelopesCrossAntimeridian(t *testing.T) {
	date := time.Date(2029, 1, 1, 12, 0, 0, 0, time.FixedZone("CST", 8*3600))
	info, ok := eclipse.LunarEclipseOnDate(date)
	if !ok {
		t.Fatal("missing lunar eclipse")
	}
	data, err := geojson.MarshalLunarEclipse(info, 360)
	if err != nil {
		t.Fatal(err)
	}
	collection := decodeCollection(t, data)
	for _, role := range []string{"visible-during-eclipse", "visible-throughout-eclipse"} {
		feature := featureWithRole(t, collection, role)
		var polygons [][][][]float64
		if err := json.Unmarshal(feature.Geometry.Coordinates, &polygons); err != nil {
			t.Fatal(err)
		}
		touchesSeam, insideBoth := false, false
		for _, polygon := range polygons {
			for _, ring := range polygon {
				for _, point := range ring {
					if math.Abs(point[0]) == 180 {
						touchesSeam = true
					}
				}
			}
		}
		if !touchesSeam {
			t.Fatalf("%s 没有落在 ±180° 上的碎片", role)
		}
		for latitude := -85.0; latitude <= 85; latitude += 5 {
			left := geometryContainsPoint(t, feature.Geometry, -179.5, latitude)
			right := geometryContainsPoint(t, feature.Geometry, 179.5, latitude)
			if left != right {
				t.Fatalf("%s 在纬 %.0f 跨越日界线不连续：%v / %v", role, latitude, left, right)
			}
			if left {
				insideBoth = true
			}
		}
		if !insideBoth {
			t.Fatalf("%s 在 ±180° 两侧没有任何共同可见纬度", role)
		}
	}
}

func lunarEnvelopeBoundaryLatitude(
	t *testing.T, feature decodedFeature, longitude, start, limit float64,
) (float64, bool) {
	t.Helper()
	if !geometryContainsPoint(t, feature.Geometry, longitude, start) ||
		geometryContainsPoint(t, feature.Geometry, longitude, limit) {
		return 0, false
	}
	low, high := start, limit
	for iteration := 0; iteration < 40; iteration++ {
		middle := (low + high) / 2
		if geometryContainsPoint(t, feature.Geometry, longitude, middle) {
			low = middle
		} else {
			high = middle
		}
	}
	return low, true
}

func lunarDenseExtremumBoundaryLatitude(
	t *testing.T, jdStart, jdEnd, longitude, start, limit float64, maximum bool,
) (float64, bool) {
	t.Helper()
	extreme := func(latitude float64) float64 {
		value := math.Inf(1)
		if maximum {
			value = math.Inf(-1)
		}
		const samples = 1000
		for index := 0; index <= samples; index++ {
			altitude := basic.HMoonHeight(jdStart+(jdEnd-jdStart)*float64(index)/samples, longitude, latitude, 0)
			if maximum {
				value = math.Max(value, altitude)
			} else {
				value = math.Min(value, altitude)
			}
		}
		return value
	}
	if extreme(start) <= 0 || extreme(limit) > 0 {
		return 0, false
	}
	low, high := start, limit
	for iteration := 0; iteration < 30; iteration++ {
		middle := (low + high) / 2
		if extreme(middle) > 0 {
			low = middle
		} else {
			high = middle
		}
	}
	return low, true
}
