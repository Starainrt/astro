package basic

import (
	"fmt"
	"math"
	"testing"

	"github.com/starainrt/astro/internal/geodata"
)

func TestSolarEclipseHybridEnvelope21640323(t *testing.T) {
	seed := JDCalc(2164, 3, 23)
	result := SolarEclipsePartialFootprints(seed, SolarEclipsePartialFootprintOptions{StepDays: 2.0 / 1440, BoundaryPoints: 96})
	if len(result.CentralBandHorizonClosures) != 2 {
		t.Fatal("missing hybrid horizon closures")
	}
	if len(result.CentralBandSegments) != 3 {
		t.Fatalf("hybrid envelope segments=%d, want annular/total/annular", len(result.CentralBandSegments))
	}
	solver := newSolarEclipseSolver(CalcMoonSHByJDE(seed, 0), SolarEclipseModelNASABulletinSplitK)
	for _, closure := range result.CentralBandHorizonClosures {
		for _, point := range []SolarEclipsePathPoint{closure[0], closure[len(closure)-1]} {
			evaluation := solver.magnitudeEvaluationAt(point.JDE)
			state := evaluation.center.stateAt(point.Longitude*rad, point.Latitude*rad, 0)
			first, _ := solarCentralBandVectorResidual(evaluation, point.Longitude, point.Latitude, 1)
			second, _ := solarCentralBandVectorResidual(evaluation, point.Longitude, point.Latitude, -1)
			if math.Abs(solarEclipseCentralContactGap(state)) > 1e-7 || math.Abs(state.sunAltitudeRad) > 1e-7 ||
				math.Min(math.Hypot(first[0], first[1]), math.Hypot(second[0], second[1])) > 1.01*solarEclipseCentralVectorTolerance {
				t.Fatalf("hybrid root is not a central-envelope horizon intersection: %+v", point)
			}
		}
	}
}

func TestSolarEclipseHybridSignedEnvelopeEvents(t *testing.T) {
	for _, date := range [][3]int{{1144, 7, 3}, {1827, 10, 20}, {1854, 11, 20}, {1986, 10, 3}, {2013, 11, 3}, {2023, 4, 20}, {2164, 3, 23}, {2172, 10, 17}} {
		t.Run(fmt.Sprintf("%04d-%02d-%02d", date[0], date[1], date[2]), func(t *testing.T) {
			seed := JDCalc(date[0], date[1], float64(date[2]))
			result := SolarEclipsePartialFootprints(seed, SolarEclipsePartialFootprintOptions{StepDays: 2.0 / 1440, BoundaryPoints: 96, CentralShadowStepDays: 2.0 / 1440})
			if result.Eclipse.Type != SolarEclipseHybrid || len(result.CentralBandHorizonClosures) != 2 || len(result.CentralBandSegments) < 2 {
				t.Fatalf("type=%s closures=%d segments=%d", result.Eclipse.Type, len(result.CentralBandHorizonClosures), len(result.CentralBandSegments))
			}
			solver := newSolarEclipseSolver(CalcMoonSHByJDE(seed, 0), SolarEclipseModelNASABulletinSplitK)
			polygons := make([][]geodata.GeoPoint, len(result.CentralBandSegments))
			for i, ring := range result.CentralBandSegments {
				if len(ring) < 4 || solarEclipsePathDistanceKM(ring[0], ring[len(ring)-1]) > 0.001 {
					t.Fatal("open hybrid component")
				}
				for _, point := range ring {
					polygons[i] = append(polygons[i], geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude})
					if solarEclipseCentralEnvelopePointOnHorizonClosure(point, result.CentralBandHorizonClosures) {
						continue
					}
					evaluation := solver.magnitudeEvaluationAt(point.JDE)
					first, firstOK := solarCentralBandVectorResidual(evaluation, point.Longitude, point.Latitude, 1)
					second, secondOK := solarCentralBandVectorResidual(evaluation, point.Longitude, point.Latitude, -1)
					if !firstOK || !secondOK || math.Min(math.Hypot(first[0], first[1]), math.Hypot(second[0], second[1])) > 1.01*solarEclipseCentralVectorTolerance {
						t.Fatalf("non-critical point: %+v residuals=%v,%v", point, first, second)
					}
				}
			}
			var visible []geodata.GeoPoint
			for _, footprint := range result.CentralShadowFootprints {
				context := solver.localStateContextAt(footprint.JDE)
				for _, boundary := range footprint.Boundaries {
					for i, point := range boundary {
						if i%4 != 0 {
							continue
						}
						state := context.stateAt(point.Longitude*rad, point.Latitude*rad, 0)
						if state.sunAltitudeRad > 0 && solarEclipseCentralContactGap(state) < -1e-8 {
							visible = append(visible, geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude})
						}
					}
				}
			}
			contained := geodata.SphericalPolygonsContainPoints(polygons, visible)
			for i, inside := range contained {
				if !inside && !geodata.SphericalPolygonsContainPathsWithinKM(polygons, [][]geodata.GeoPoint{{visible[i], visible[i]}}, false, 0.05) {
					t.Errorf("visible central-shadow point outside envelope: %+v", visible[i])
				}
			}
			t.Logf("segments=%d visible shadow samples=%d", len(polygons), len(visible))
		})
	}
}
