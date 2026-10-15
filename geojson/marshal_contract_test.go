package geojson_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/moon"
)

// 本文件钉住三条契约：全掩可见性轮廓必须校验、退化影区整条缺省、
// 地平闭合必须能区分精确擦地点闭合与采样端点近似闭合。

func copyOccultationContours(contours [][]moon.OccultationPathPoint) [][]moon.OccultationPathPoint {
	copied := make([][]moon.OccultationPathPoint, len(contours))
	for index, contour := range contours {
		copied[index] = append([]moon.OccultationPathPoint(nil), contour...)
	}
	return copied
}

func TestMarshalPlanetOccultationValidatesTotalVisibilityContours(t *testing.T) {
	path := mars20250729TestFixture(t, time.Minute, true).path
	if !path.HasTotalBand || len(path.TotalBandContours) == 0 || len(path.TotalBandContours[0]) < 2 {
		t.Fatal("fixture has no usable total band contour")
	}
	// 复制的样本本身合法，只有注入的可见性轮廓是待测对象。
	valid := append([]moon.OccultationPathPoint(nil),
		path.TotalBandContours[0][0], path.TotalBandContours[0][1])

	outOfOrder := path
	outOfOrder.TotalVisibilityContours = [][]moon.OccultationPathPoint{{valid[0], valid[0]}}
	if _, err := geojson.MarshalPlanetOccultation(outOfOrder); err == nil {
		t.Fatal("out-of-order total visibility contours must be rejected")
	}

	outside := path
	outside.TotalVisibilityContours = [][]moon.OccultationPathPoint{{
		{Time: path.TotalStart.Time.Add(-time.Minute)}, valid[1],
	}}
	if _, err := geojson.MarshalPlanetOccultation(outside); err == nil {
		t.Fatal("total visibility contours outside the total interval must be rejected")
	}

	orphan := path
	orphan.HasTotalBand = false
	orphan.TotalComplete = false
	orphan.TotalStart = moon.OccultationPathPoint{}
	orphan.TotalEnd = moon.OccultationPathPoint{}
	orphan.NorthernTotalLimit = nil
	orphan.SouthernTotalLimit = nil
	orphan.TotalFootprints = nil
	orphan.TotalBandFootprints = nil
	orphan.TotalBandContours = nil
	orphan.TotalRiseSetCurves = nil
	orphan.GreatestTotalWidthKM = 0
	orphan.TotalVisibilityContours = [][]moon.OccultationPathPoint{{valid[0], valid[1]}}
	if _, err := geojson.MarshalPlanetOccultation(orphan); err == nil {
		t.Fatal("total visibility contours without HasTotalBand must be rejected")
	}
	orphan.TotalVisibilityContours = nil
	if _, err := geojson.MarshalPlanetOccultation(orphan); err != nil {
		t.Fatalf("cleared total-band fields must marshal: %v", err)
	}
}

func degenerateShadowBoundary(stamp time.Time) [][]eclipse.SolarEclipsePathPoint {
	return [][]eclipse.SolarEclipsePathPoint{{
		{Time: stamp, Longitude: 0, Latitude: 0},
		{Time: stamp, Longitude: 0.01, Latitude: 0},
		{Time: stamp, Longitude: 0.01, Latitude: 1e-11},
		{Time: stamp, Longitude: 0, Latitude: 1e-11},
	}}
}

func TestMarshalSolarEclipseShadowInstantSkipsDegenerateRegion(t *testing.T) {
	stamp := time.Date(2024, time.April, 8, 18, 0, 0, 0, time.UTC)
	instant := eclipse.SolarEclipseShadowInstant{
		Time: stamp, Closed: true, Boundaries: degenerateShadowBoundary(stamp),
	}
	raw, err := geojson.MarshalSolarEclipseShadowInstant(instant)
	if err != nil {
		t.Fatalf("degenerate footprint must be omitted, not rejected: %v", err)
	}
	var collection struct {
		Features []json.RawMessage `json:"features"`
	}
	if err := json.Unmarshal(raw, &collection); err != nil {
		t.Fatalf("decode GeoJSON: %v", err)
	}
	if len(collection.Features) != 0 {
		t.Fatalf("features=%d, want the degenerate region omitted", len(collection.Features))
	}
}

func horizonCutShadowInstant(t *testing.T) eclipse.SolarEclipseShadowInstant {
	t.Helper()
	date := time.Date(2009, time.July, 22, 0, 0, 0, 0, time.UTC)
	info, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
		Step: 2 * time.Minute, BoundaryPoints: 96, CentralShadowStep: 2 * time.Minute,
		DisableRiseSet: true,
	})
	if !ok {
		t.Fatal("expected the 2009-07-22 eclipse")
	}
	for _, footprint := range info.CentralShadowFootprints {
		if footprint.Closed || len(footprint.HorizonEnds) != 2 {
			continue
		}
		return eclipse.SolarEclipseShadowInstant{
			Time: footprint.Time, Kind: eclipse.SolarEclipseShadowUmbra,
			Boundaries: footprint.Boundaries, HorizonEnds: footprint.HorizonEnds,
		}
	}
	t.Fatal("fixture has no horizon-cut central-shadow footprint")
	return eclipse.SolarEclipseShadowInstant{}
}

func shadowClosureExact(t *testing.T, instant eclipse.SolarEclipseShadowInstant) interface{} {
	t.Helper()
	raw, err := geojson.MarshalSolarEclipseShadowInstant(instant)
	if err != nil {
		t.Fatalf("MarshalSolarEclipseShadowInstant: %v", err)
	}
	region := featureWithRole(t, decodeCollection(t, raw), "central-shadow-footprint")
	closure, ok := region.Properties["closure"].(map[string]interface{})
	if !ok {
		t.Fatal("horizon-cut region has no closure property")
	}
	if closure["kind"] != "horizon" {
		t.Fatalf("closure kind=%v, want horizon", closure["kind"])
	}
	return closure["exact"]
}

func TestMarshalSolarEclipseShadowInstantMarksSampledClosure(t *testing.T) {
	instant := horizonCutShadowInstant(t)
	if exact := shadowClosureExact(t, instant); exact != true {
		t.Fatalf("closure exact=%v, want true with two grazing points", exact)
	}
	sampled := instant
	sampled.HorizonEnds = nil
	if exact := shadowClosureExact(t, sampled); exact != false {
		t.Fatalf("closure exact=%v, want false without grazing points", exact)
	}
}

func TestSolarEclipseCentralShadowClosureExactnessMatchesGrazingPoints(t *testing.T) {
	fixtures := centralShadowRegionFixtures()
	exactRegions := 0
	cutRegions := 0
	for _, fixture := range fixtures {
		collection, info := centralShadowMarshalFixture(t, fixture)
		grazing := make(map[string]bool, len(info.CentralShadowFootprints))
		for _, footprint := range info.CentralShadowFootprints {
			grazing[footprint.Time.UTC().Format(time.RFC3339Nano)] = len(footprint.HorizonEnds) == 2
		}
		for _, region := range featuresWithRole(collection, "central-shadow-footprint") {
			closure, ok := region.Properties["closure"].(map[string]interface{})
			if !ok {
				continue
			}
			cutRegions++
			stamp, _ := region.Properties["time"].(string)
			want, present := grazing[stamp]
			if !present {
				t.Fatalf("%s: region %s has no matching footprint", fixture.name, stamp)
			}
			exact, ok := closure["exact"].(bool)
			if !ok {
				t.Fatalf("%s: closure of %s has no exact flag", fixture.name, stamp)
			}
			if exact != want {
				t.Fatalf("%s: closure exact=%v, want %v for %s", fixture.name, exact, want, stamp)
			}
			if exact {
				exactRegions++
			}
		}
	}
	if cutRegions == 0 || exactRegions == 0 {
		t.Fatalf("horizon-cut regions=%d with grazing points=%d; the marker is untested", cutRegions, exactRegions)
	}
}
