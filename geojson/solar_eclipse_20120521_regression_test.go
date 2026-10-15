package geojson_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"

	"github.com/starainrt/astro/internal/ephemtest"
)

func TestSolarEclipse20120521HasCentralBandHorizonClosure(t *testing.T) {
	date := time.Date(2012, time.May, 21, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
		Step: 2 * time.Minute, BoundaryPoints: 96, CentralShadowStep: 2 * time.Minute,
	})
	if !ok {
		t.Fatal("expected solar eclipse footprints")
	}
	if len(partial.CentralBandHorizonClosures) != 2 {
		t.Fatalf("central-limit horizon closures=%d, want start and end", len(partial.CentralBandHorizonClosures))
	}
	expectedRoots := solarEclipse20120521ExpectedRoots(t, partial)
	for closureIndex, closure := range partial.CentralBandHorizonClosures {
		if len(closure) < 2 {
			t.Fatalf("horizon closure %d has %d points", closureIndex, len(closure))
		}
		assertSolarPathMaximumEdgeKM(t, closure, 12)
		roots := [2]eclipse.SolarEclipsePathPoint{closure[0], closure[len(closure)-1]}
		for rootIndex, root := range roots {
			if distance := geoJSONCoordinateDistanceKM(
				[]float64{root.Longitude, root.Latitude}, expectedRoots[closureIndex][rootIndex],
			); distance > 0.1 {
				t.Fatalf("horizon closure %d root %d differs by %.3f km", closureIndex, rootIndex, distance)
			}
		}
	}
	for closureIndex, closure := range partial.CentralBandHorizonClosures {
		direction := eclipse.RiseSetDirectionRise
		if closureIndex == 1 {
			direction = eclipse.RiseSetDirectionSet
		}
		for pointIndex, point := range closure {
			if !solarGreatestCurveContainsPoint(partial.RiseSetCurves, point, direction) {
				t.Fatalf("horizon closure %d point %d is missing from its greatest/%s curve", closureIndex, pointIndex, direction)
			}
		}
	}

	central, ok := eclipse.SolarEclipseCentralPath(date, eclipse.SolarEclipsePathOptions{
		Step: 2 * time.Minute, TargetSpacingKM: 700,
	})
	if !ok || len(central.CenterLine) < 2 {
		t.Fatal("expected a central eclipse path")
	}
	data, err := geojson.MarshalSolarEclipse(partial, &central)
	if err != nil {
		t.Fatalf("MarshalSolarEclipse: %v", err)
	}
	band := featureWithRole(t, decodeCollection(t, data), "central-band")
	if band.Properties["source"] != "besselian-critical-envelope" {
		t.Fatalf("central-band source=%v, want continuous critical envelope", band.Properties["source"])
	}
	assertClosedMultiPolygon(t, band)
	var polygons [][][][]float64
	if err := json.Unmarshal(band.Geometry.Coordinates, &polygons); err != nil {
		t.Fatalf("decode central band: %v", err)
	}
	if len(polygons) == 0 || len(polygons) > 2 {
		t.Fatalf("central band has %d polygons, want one physical band with at most one antimeridian split", len(polygons))
	}
	for _, polygon := range polygons {
		if len(polygon) == 0 {
			t.Fatal("central-band polygon has no exterior ring")
		}
		assertSolarCentralBandRingSimpleAndSampled(t, polygon[0], 250)
	}
	for closureIndex, closure := range partial.CentralBandHorizonClosures {
		for pointIndex, point := range closure {
			coordinate := []float64{point.Longitude, point.Latitude}
			if distance := geoJSONMultiPolygonBoundaryDistanceKM(polygons, coordinate); distance > 0.1 {
				t.Fatalf("horizon closure %d point %d is %.3f km from the central-band boundary", closureIndex, pointIndex, distance)
			}
		}
	}
	for segmentIndex, segment := range central.CenterLine {
		point := []float64{segment.Longitude, segment.Latitude}
		if !geometryContainsPoint(t, band.Geometry, point[0], point[1]) &&
			geoJSONMultiPolygonBoundaryDistanceKM(polygons, point) > 10 {
			t.Fatalf("center-line point %d lies outside central band", segmentIndex)
		}
	}
}

func TestSolarEclipse20120521CentralBandContainsVisibleAnnularStation(t *testing.T) {
	localDate := time.Date(2012, time.May, 21, 0, 0, 0, 0, time.FixedZone("UTC+8", 8*60*60))
	local, ok := eclipse.LocalSolarEclipseOnDate(localDate, 120.4913, 27.4779, 0)
	if !ok || local.Type != eclipse.SolarEclipseAnnular || !local.HasCentral || local.SunAltitude <= 0 {
		t.Fatalf("reference station is not visibly annular: ok=%v type=%s central=%v altitude=%.6f",
			ok, local.Type, local.HasCentral, local.SunAltitude)
	}
	date := time.Date(2012, time.May, 21, 0, 0, 0, 0, time.UTC)
	central, ok := eclipse.SolarEclipseCentralPath(date, eclipse.SolarEclipsePathOptions{
		Step: 2 * time.Minute, TargetSpacingKM: 700,
	})
	if !ok {
		t.Fatal("expected solar central path")
	}
	for _, shadowStep := range []time.Duration{0, 2 * time.Minute} {
		shadowStep := shadowStep
		t.Run(shadowStep.String(), func(t *testing.T) {
			partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
				Step: 2 * time.Minute, BoundaryPoints: 96, CentralShadowStep: shadowStep,
			})
			if !ok {
				t.Fatal("expected solar eclipse footprints")
			}
			data, err := geojson.MarshalSolarEclipse(partial, &central)
			if err != nil {
				t.Fatalf("MarshalSolarEclipse: %v", err)
			}
			band := featureWithRole(t, decodeCollection(t, data), "central-band")
			if !geometryContainsPoint(t, band.Geometry, 120.4913, 27.4779) {
				t.Fatal("central-band omits a station that visibly sees annularity")
			}
			if source := band.Properties["source"]; source != "besselian-critical-envelope" {
				t.Fatalf("central-band source=%v, want continuous critical envelope", source)
			}
		})
	}
}

func assertSolarPathMaximumEdgeKM(
	t *testing.T,
	points []eclipse.SolarEclipsePathPoint,
	maximumKM float64,
) {
	t.Helper()
	for index := 1; index < len(points); index++ {
		distance := geoJSONCoordinateDistanceKM(
			[]float64{points[index-1].Longitude, points[index-1].Latitude},
			[]float64{points[index].Longitude, points[index].Latitude},
		)
		if distance > maximumKM {
			t.Fatalf("horizon closure edge %d is %.3f km, want at most %.3f km", index-1, distance, maximumKM)
		}
	}
}

func TestSolarEclipse20120521HorizonClosuresAreStableAcrossSampling(t *testing.T) {
	date := time.Date(2012, time.May, 21, 0, 0, 0, 0, time.UTC)
	reference, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
		Step: 2 * time.Minute, BoundaryPoints: 96, CentralShadowStep: 2 * time.Minute,
	})
	if !ok || len(reference.CentralBandHorizonClosures) != 2 {
		t.Fatalf("reference closures=%d ok=%v, want two", len(reference.CentralBandHorizonClosures), ok)
	}
	expectedRoots := solarEclipse20120521ExpectedRoots(t, reference)
	for _, step := range []time.Duration{time.Minute, 5 * time.Minute, 10 * time.Minute} {
		partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
			Step: step, BoundaryPoints: 96, CentralShadowStep: step,
		})
		if !ok || len(partial.CentralBandHorizonClosures) != 2 {
			t.Fatalf("step %s: closures=%d ok=%v, want two", step, len(partial.CentralBandHorizonClosures), ok)
		}
		for closureIndex, closure := range partial.CentralBandHorizonClosures {
			roots := [2]eclipse.SolarEclipsePathPoint{closure[0], closure[len(closure)-1]}
			for rootIndex, root := range roots {
				if distance := geoJSONCoordinateDistanceKM(
					[]float64{root.Longitude, root.Latitude}, expectedRoots[closureIndex][rootIndex],
				); distance > 0.1 {
					t.Fatalf("step %s: closure %d root %d differs by %.3f km", step, closureIndex, rootIndex, distance)
				}
			}
		}
	}
}

func solarGreatestCurveContainsPoint(
	curves []eclipse.SolarEclipseRiseSetCurve,
	want eclipse.SolarEclipsePathPoint,
	direction eclipse.RiseSetDirection,
) bool {
	for _, curve := range curves {
		if curve.Phase != eclipse.RiseSetPhaseGreatest || curve.Direction != direction {
			continue
		}
		for _, segment := range curve.Segments {
			for _, point := range segment {
				if point.Time.Equal(want.Time) && geoJSONCoordinateDistanceKM(
					[]float64{point.Longitude, point.Latitude},
					[]float64{want.Longitude, want.Latitude},
				) <= 0.001 {
					return true
				}
			}
		}
	}
	return false
}

// solarEclipse20120521ExpectedRoots 两条地平闭合线的起止根；新模型的值取自 2 分钟步长的参考结果。
func solarEclipse20120521ExpectedRoots(t *testing.T, reference eclipse.SolarEclipsePartialFootprintsInfo) [2][2][]float64 {
	t.Helper()
	roots := [2][2][]float64{
		{{109.6259942, 19.9359024}, {107.7443439, 22.3910851}},
		{{-100.0855931, 34.1224725}, {-102.2045906, 31.7284072}},
	}
	values := ephemtest.NewValues(t, "solar_eclipse_20120521_roots")
	for c, closure := range reference.CentralBandHorizonClosures {
		ends := [2]eclipse.SolarEclipsePathPoint{closure[0], closure[len(closure)-1]}
		for r, end := range ends {
			key := fmt.Sprintf("closure%d/root%d", c, r)
			roots[c][r] = []float64{
				values.Want(key+"/lon", roots[c][r][0], end.Longitude),
				values.Want(key+"/lat", roots[c][r][1], end.Latitude),
			}
		}
	}
	return roots
}
