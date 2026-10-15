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

// 本文件把上游消费方要的契约钉死：
//  1. central-shadow-footprint 要么缺省、要么是 Polygon/MultiPolygon，永不出现线类型；
//  2. 被地平线切断的足迹用该时刻地平圈上的擦地点闭合，闭合弧与地平圈在容差内一致；
//  3. 物理边界曲线另出 central-shadow-boundary，顶点与采样曲线一致，供调用方描边；
//  4. 区域外环保持右手定则，且不出现跨图收口边（含绕极环）。
// This file pins the contract the consumers asked for: the footprint role is always a
// region, the horizon closure really lies on the horizon circle of the footprint time,
// the physical boundary is exported separately with unchanged vertices, and every
// outer ring is right-handed without a synthetic seam.

const centralShadowHorizonToleranceDegrees = 0.01

type centralShadowRegionFixture struct {
	name        string
	year        int
	month       time.Month
	day         int
	wantHorizon bool
}

func centralShadowRegionFixtures() []centralShadowRegionFixture {
	return []centralShadowRegionFixture{
		{name: "2009-07-22 total, both limb cuts", year: 2009, month: time.July, day: 22, wantHorizon: true},
		{name: "2014-04-29 non-central annular, every footprint cut", year: 2014, month: time.April, day: 29, wantHorizon: true},
		{name: "2043-04-09 total, every footprint cut", year: 2043, month: time.April, day: 9, wantHorizon: true},
		{name: "2021-12-04 antarctic total", year: 2021, month: time.December, day: 4, wantHorizon: false},
		{name: "2021-06-10 pole-enclosing footprints", year: 2021, month: time.June, day: 10, wantHorizon: false},
		{name: "2061-10-13 annular with many cut footprints", year: 2061, month: time.October, day: 13, wantHorizon: true},
	}
}

func centralShadowMarshalFixture(
	t *testing.T,
	fixture centralShadowRegionFixture,
) (decodedCollection, eclipse.SolarEclipsePartialFootprintsInfo) {
	t.Helper()
	info, ok := eclipse.SolarEclipsePartialFootprints(
		time.Date(fixture.year, fixture.month, fixture.day, 0, 0, 0, 0, time.UTC),
		eclipse.SolarEclipsePartialFootprintOptions{
			Step: 10 * time.Minute, BoundaryPoints: 96, CentralShadowStep: 2 * time.Minute,
			DisableRiseSet: true,
		},
	)
	if !ok {
		t.Fatalf("%s: no solar eclipse", fixture.name)
	}
	data, err := geojson.MarshalSolarEclipse(info, nil)
	if err != nil {
		t.Fatalf("%s: MarshalSolarEclipse: %v", fixture.name, err)
	}
	return decodeCollection(t, data), info
}

func TestSolarEclipseCentralShadowFootprintIsAlwaysARegion(t *testing.T) {
	cutRegions := 0
	for _, fixture := range centralShadowRegionFixtures() {
		collection, info := centralShadowMarshalFixture(t, fixture)
		if len(info.CentralShadowFootprints) == 0 {
			t.Fatalf("%s: sampler produced no central-shadow footprints", fixture.name)
		}
		regions := featuresWithRole(collection, "central-shadow-footprint")
		if len(regions) == 0 {
			t.Fatalf("%s: GeoJSON has no central-shadow-footprint", fixture.name)
		}
		openRegions := 0
		for _, region := range regions {
			switch region.Geometry.Type {
			case "Polygon", "MultiPolygon":
			default:
				t.Fatalf("%s: central-shadow-footprint geometry=%q, want Polygon or MultiPolygon",
					fixture.name, region.Geometry.Type)
			}
			closed, present := region.Properties["source_boundary_closed"].(bool)
			if !present {
				t.Fatalf("%s: central-shadow-footprint has no source_boundary_closed property", fixture.name)
			}
			if !closed {
				openRegions++
				closure, present := region.Properties["closure"].(map[string]interface{})
				if !present {
					t.Fatalf("%s: horizon-cut footprint has no closure property", fixture.name)
				}
				if closure["kind"] != "horizon" {
					t.Fatalf("%s: closure kind=%v, want horizon", fixture.name, closure["kind"])
				}
				if closure["time"] != region.Properties["time"] {
					t.Fatalf("%s: closure time=%v, want the footprint time %v",
						fixture.name, closure["time"], region.Properties["time"])
				}
			} else if _, present := region.Properties["closure"]; present {
				t.Fatalf("%s: self-closed footprint must not carry a closure property", fixture.name)
			}
		}
		boundaries := featuresWithRole(collection, "central-shadow-boundary")
		if len(boundaries) != openRegions {
			t.Fatalf("%s: central-shadow-boundary count=%d, want one per horizon-cut region (%d)",
				fixture.name, len(boundaries), openRegions)
		}
		for _, boundary := range boundaries {
			if boundary.Geometry.Type != "MultiLineString" {
				t.Fatalf("%s: central-shadow-boundary geometry=%q, want MultiLineString",
					fixture.name, boundary.Geometry.Type)
			}
			if closed, _ := boundary.Properties["source_boundary_closed"].(bool); closed {
				t.Fatalf("%s: central-shadow-boundary must stay an open boundary", fixture.name)
			}
		}
		if fixture.wantHorizon && openRegions == 0 {
			t.Fatalf("%s: expected at least one horizon-cut central-shadow footprint", fixture.name)
		}
		cutRegions += openRegions
	}
	if cutRegions == 0 {
		t.Fatal("no horizon-cut central-shadow footprint in the fixture set; the contract is untested")
	}
}

func centralShadowRegionRings(t *testing.T, feature decodedFeature) [][][2]float64 {
	t.Helper()
	var polygons [][][][]float64
	if err := json.Unmarshal(feature.Geometry.Coordinates, &polygons); err != nil {
		t.Fatalf("decode central-shadow-footprint polygon: %v", err)
	}
	rings := make([][][2]float64, 0, len(polygons))
	for _, polygon := range polygons {
		if len(polygon) == 0 {
			continue
		}
		ring := make([][2]float64, 0, len(polygon[0]))
		for _, coordinate := range polygon[0] {
			ring = append(ring, [2]float64{coordinate[0], coordinate[1]})
		}
		rings = append(rings, ring)
	}
	return rings
}

func centralShadowBoundaryLines(t *testing.T, feature decodedFeature) [][][2]float64 {
	t.Helper()
	var lines [][][]float64
	if err := json.Unmarshal(feature.Geometry.Coordinates, &lines); err != nil {
		t.Fatalf("decode central-shadow-boundary line: %v", err)
	}
	result := make([][][2]float64, 0, len(lines))
	for _, line := range lines {
		points := make([][2]float64, 0, len(line))
		for _, coordinate := range line {
			points = append(points, [2]float64{coordinate[0], coordinate[1]})
		}
		result = append(result, points)
	}
	return result
}

func centralShadowSubsolarPoint(value time.Time) (float64, float64) {
	ttJDE := basic.UTC2TT(basic.Date2JD(value.UTC()))
	ra, dec := basic.HSunApparentRaDec(ttJDE)
	utJDE := basic.TT2UT1(ttJDE)
	longitude := ra - basic.ApparentSiderealTime(utJDE)*15
	for longitude > 180 {
		longitude -= 360
	}
	for longitude < -180 {
		longitude += 360
	}
	return longitude, dec
}

func centralShadowHorizonDistanceDegrees(vertex [2]float64, longitude, latitude float64) float64 {
	first, second := vertex[1]*math.Pi/180, latitude*math.Pi/180
	deltaLongitude := (longitude - vertex[0]) * math.Pi / 180
	deltaLatitude := second - first
	h := math.Sin(deltaLatitude/2)*math.Sin(deltaLatitude/2) +
		math.Cos(first)*math.Cos(second)*math.Sin(deltaLongitude/2)*math.Sin(deltaLongitude/2)
	if h > 1 {
		h = 1
	}
	// 90 degrees minus the distance to the subsolar point is the solar altitude.
	return 90 - 2*math.Asin(math.Sqrt(h))*180/math.Pi
}

// centralShadowClosureViolations 返回所有"没有落在地平圈上"的闭合弧顶点描述。
// centralShadowClosureViolations reports every closure vertex that misses the horizon.
func centralShadowClosureViolations(
	t *testing.T,
	collection decodedCollection,
) []string {
	t.Helper()
	return solarShadowClosureViolations(t, collection, "central-shadow-footprint")
}

func solarShadowClosureViolations(
	t *testing.T,
	collection decodedCollection,
	role string,
) []string {
	t.Helper()
	var issues []string
	for _, region := range featuresWithRole(collection, role) {
		closed, _ := region.Properties["source_boundary_closed"].(bool)
		if closed {
			continue
		}
		stamp, _ := region.Properties["time"].(string)
		value, err := time.Parse(time.RFC3339Nano, stamp)
		if err != nil {
			t.Fatalf("parse footprint time %q: %v", stamp, err)
		}
		longitude, latitude := centralShadowSubsolarPoint(value)
		onHorizon := 0
		for _, ring := range centralShadowRegionRings(t, region) {
			for _, vertex := range ring {
				altitude := centralShadowHorizonDistanceDegrees(vertex, longitude, latitude)
				if altitude < -centralShadowHorizonToleranceDegrees {
					issues = append(issues, fmt.Sprintf(
						"%s: vertex %.6f,%.6f is %.4f deg below the horizon", stamp, vertex[0], vertex[1], altitude))
					continue
				}
				if math.Abs(altitude) <= centralShadowHorizonToleranceDegrees {
					onHorizon++
				}
			}
		}
		if onHorizon < 2 {
			issues = append(issues, fmt.Sprintf(
				"%s: only %d ring vertices lie on the horizon, want the closing arc", stamp, onHorizon))
		}
	}
	return issues
}

func TestSolarEclipseCentralShadowClosureLiesOnTheHorizon(t *testing.T) {
	for _, fixture := range centralShadowRegionFixtures() {
		collection, _ := centralShadowMarshalFixture(t, fixture)
		if issues := centralShadowClosureViolations(t, collection); len(issues) > 0 {
			t.Fatalf("%s: %s", fixture.name, issues[0])
		}
		// The physical boundary feature must start and end on the horizon as well,
		// otherwise the region would still be cut short of the terminator.
		for _, boundary := range featuresWithRole(collection, "central-shadow-boundary") {
			stamp, _ := boundary.Properties["time"].(string)
			value, err := time.Parse(time.RFC3339Nano, stamp)
			if err != nil {
				t.Fatalf("%s: parse boundary time %q: %v", fixture.name, stamp, err)
			}
			longitude, latitude := centralShadowSubsolarPoint(value)
			for _, line := range centralShadowBoundaryLines(t, boundary) {
				for _, index := range []int{0, len(line) - 1} {
					altitude := centralShadowHorizonDistanceDegrees(line[index], longitude, latitude)
					if math.Abs(altitude) > centralShadowHorizonToleranceDegrees {
						t.Fatalf("%s: %s boundary endpoint %.6f,%.6f has altitude %.4f deg, want on the horizon",
							fixture.name, stamp, line[index][0], line[index][1], altitude)
					}
				}
			}
		}
	}
}

// TestSolarEclipseCentralShadowClosureTestDetectsCutShortBoundaries 证明上面的容差检查
// 真的能抓住"边界提前停止"的几何：把擦地点清空后退回旧的封口方式，检查必须报错。
// TestSolarEclipseCentralShadowClosureTestDetectsCutShortBoundaries proves the
// tolerance check above has teeth: without the grazing points the fallback closure
// stops short of the horizon and the check must report it.
func TestSolarEclipseCentralShadowClosureTestDetectsCutShortBoundaries(t *testing.T) {
	var fixture centralShadowRegionFixture
	for _, candidate := range centralShadowRegionFixtures() {
		if candidate.wantHorizon {
			fixture = candidate
			break
		}
	}
	info, ok := eclipse.SolarEclipsePartialFootprints(
		time.Date(fixture.year, fixture.month, fixture.day, 0, 0, 0, 0, time.UTC),
		eclipse.SolarEclipsePartialFootprintOptions{
			Step: 10 * time.Minute, BoundaryPoints: 96, CentralShadowStep: 2 * time.Minute,
			DisableRiseSet: true,
		},
	)
	if !ok {
		t.Fatalf("%s: no solar eclipse", fixture.name)
	}
	open := 0
	for index := range info.CentralShadowFootprints {
		if info.CentralShadowFootprints[index].Closed {
			continue
		}
		info.CentralShadowFootprints[index].HorizonEnds = nil
		open++
	}
	if open == 0 {
		t.Fatalf("%s: no horizon-cut footprint to degrade", fixture.name)
	}
	data, err := geojson.MarshalSolarEclipse(info, nil)
	if err != nil {
		t.Fatalf("%s: MarshalSolarEclipse: %v", fixture.name, err)
	}
	collection := decodeCollection(t, data)
	if len(featuresWithRole(collection, "central-shadow-boundary")) == 0 {
		t.Fatal("degraded export lost its physical boundary features")
	}
	if issues := centralShadowClosureViolations(t, collection); len(issues) == 0 {
		t.Fatal("the horizon tolerance check passed on a closure that stops short of the horizon")
	}
}

func TestSolarEclipseCentralShadowBoundaryKeepsTheSampledCurve(t *testing.T) {
	fixture := centralShadowRegionFixtures()[0]
	collection, info := centralShadowMarshalFixture(t, fixture)
	boundaries := featuresWithRole(collection, "central-shadow-boundary")
	if len(boundaries) == 0 {
		t.Fatalf("%s: no central-shadow-boundary features", fixture.name)
	}
	byTime := map[string]decodedFeature{}
	for _, boundary := range boundaries {
		stamp, _ := boundary.Properties["time"].(string)
		byTime[stamp] = boundary
	}
	checked := 0
	for _, footprint := range info.CentralShadowFootprints {
		if footprint.Closed {
			continue
		}
		feature, present := byTime[footprint.Time.UTC().Format(time.RFC3339Nano)]
		if !present {
			t.Fatalf("%s: horizon-cut footprint %v has no central-shadow-boundary feature",
				fixture.name, footprint.Time.UTC())
		}
		expected := make([][2]float64, 0, len(footprint.HorizonEnds)+2)
		expected = append(expected, [2]float64{footprint.HorizonEnds[0].Longitude, footprint.HorizonEnds[0].Latitude})
		for _, segment := range footprint.Boundaries {
			for _, point := range segment {
				expected = append(expected, [2]float64{point.Longitude, point.Latitude})
			}
		}
		expected = append(expected, [2]float64{footprint.HorizonEnds[1].Longitude, footprint.HorizonEnds[1].Latitude})
		lines := centralShadowBoundaryLines(t, feature)
		if len(lines) != 1 {
			t.Fatalf("%s: %v boundary has %d segments, want the joined physical curve",
				fixture.name, footprint.Time.UTC(), len(lines))
		}
		if len(lines[0]) != len(expected) {
			t.Fatalf("%s: %v boundary has %d vertices, want %d (grazing points plus the sampled curve)",
				fixture.name, footprint.Time.UTC(), len(lines[0]), len(expected))
		}
		for index := range expected {
			if lines[0][index] != expected[index] {
				t.Fatalf("%s: %v boundary vertex %d = %.9f,%.9f, want %.9f,%.9f",
					fixture.name, footprint.Time.UTC(), index,
					lines[0][index][0], lines[0][index][1], expected[index][0], expected[index][1])
			}
		}
		checked++
	}
	if checked == 0 {
		t.Fatalf("%s: no horizon-cut footprint was compared", fixture.name)
	}
}

func TestSolarEclipseCentralShadowRegionRingsAreRightHanded(t *testing.T) {
	poleRings := 0
	for _, fixture := range centralShadowRegionFixtures() {
		collection, _ := centralShadowMarshalFixture(t, fixture)
		for _, region := range featuresWithRole(collection, "central-shadow-footprint") {
			for _, ring := range centralShadowRegionRings(t, region) {
				if len(ring) < 4 {
					t.Fatalf("%s: ring has %d vertices, want a closed ring", fixture.name, len(ring))
				}
				area, jumps, winding := 0.0, 0, 0.0
				for index := range ring {
					next := ring[(index+1)%len(ring)]
					area += ring[index][0]*next[1] - next[0]*ring[index][1]
					delta := math.Remainder(next[0]-ring[index][0], 360)
					winding += delta
					if math.Abs(delta) > 180 {
						jumps++
					}
				}
				if area <= 0 {
					t.Fatalf("%s: %v ring is not counter-clockwise in lon/lat (area %.6f)",
						fixture.name, region.Properties["time"], area/2)
				}
				if jumps > 0 {
					t.Fatalf("%s: %v ring has %d segment(s) jumping across the map edge",
						fixture.name, region.Properties["time"], jumps)
				}
				if math.Abs(winding) >= 180 {
					poleRings++
					if hemisphereArea := 360 * 180; area/2 >= float64(hemisphereArea) {
						t.Fatalf("%s: %v pole ring covers %.1f deg2, want the enclosed cap",
							fixture.name, region.Properties["time"], area/2)
					}
				}
			}
		}
	}
	if poleRings == 0 {
		t.Fatal("no pole-enclosing ring in the fixture set; the polar convention is untested")
	}
}

func TestSolarEclipseSampledPartialClosureLiesOnTheHorizon(t *testing.T) {
	date := time.Date(2009, time.July, 22, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
		Step: 2 * time.Minute, BoundaryPoints: 96, CentralShadowStep: 2 * time.Minute,
		DisableRiseSet: true,
	})
	if !ok {
		t.Fatal("expected the 2009-07-22 eclipse")
	}
	collection := decodeCollection(t, mustMarshalSolarEclipse(t, partial))
	if issues := solarShadowClosureViolations(t, collection, "partial-footprint"); len(issues) > 0 {
		t.Fatalf("sampled partial footprint: %s", issues[0])
	}
}
