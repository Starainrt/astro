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
		jd := basic.Date2JDE(contact.at.UTC())
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
	if altitude := basic.HMoonHeight(basic.Date2JDE(info.PenumbralStart), point.longitude, point.latitude, 0); altitude >= -0.01 {
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
	}{{"visible-at-p1", info.PenumbralStart}, {"visible-at-p4", info.PenumbralEnd}} {
		band := featureWithRole(t, collection, contact.role)
		var polygons [][][][]float64
		if err := json.Unmarshal(band.Geometry.Coordinates, &polygons); err != nil {
			t.Fatal(err)
		}
		jd := basic.Date2JDE(contact.at.UTC())
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
