package basic

import (
	"math"
	"testing"
)

// 等值线追踪走插值星历，但每个输出点都必须满足精确星历下的食分 / contour tracing uses the
// interpolator, yet every output point must satisfy the target magnitude under exact ephemerides.
func TestSolarEclipseMagnitudeContoursExactAtOutputPoints(t *testing.T) {
	cases := []struct {
		name          string
		year, month   int
		day           float64
		deltaTSeconds float64
	}{
		{"2023-04-20 hybrid", 2023, 4, 20, 0},
		{"2024-04-08 total", 2024, 4, 8, 0},
		{"2022-10-25 partial", 2022, 10, 25, 0},
		{"2027-08-02 total explicit ΔT", 2027, 8, 2, 75},
	}
	magnitudes := []float64{0.2, 0.5, 0.8}
	for _, tc := range cases {
		seed := JDCalc(tc.year, tc.month, tc.day)
		result := SolarEclipsePartialFootprints(seed, SolarEclipsePartialFootprintOptions{
			StepDays: 5.0 / 1440, BoundaryPoints: 48, MagnitudeValues: magnitudes,
			DeltaTSeconds: tc.deltaTSeconds,
		})
		solver := newSolarEclipseSolverWithOptions(solarEclipseAnchorNewMoon(seed), SolarEclipseOptions{
			RadiusModel: SolarEclipseModelNASABulletinSplitK,
		}).withDeltaTSeconds(tc.deltaTSeconds)
		if len(result.MagnitudeContours) != len(magnitudes) {
			t.Fatalf("%s: %d contours, want %d", tc.name, len(result.MagnitudeContours), len(magnitudes))
		}
		for _, contour := range result.MagnitudeContours {
			points := 0
			for _, segment := range contour.Segments {
				for _, point := range segment {
					// 端点可能由地平线或混合食分支精修得到，只检查追踪内点 / endpoints may come
					// from horizon or hybrid refinement; only traced interior points are checked.
					if point.SunAltitude < 0.05 {
						continue
					}
					state := solver.localStateContextAt(point.JDE).stateAt(point.Longitude*rad, point.Latitude*rad, 0)
					got := solarEclipseMagnitudeAtTarget(state, contour.Magnitude)
					if math.IsNaN(got) || math.Abs(got-contour.Magnitude) > 1e-6 {
						t.Fatalf("%s m=%.1f point (%.6f, %.6f) at %.8f: exact magnitude %.9f",
							tc.name, contour.Magnitude, point.Longitude, point.Latitude, point.JDE, got)
					}
					points++
				}
			}
			if points < 20 {
				t.Fatalf("%s m=%.1f: only %d interior points checked", tc.name, contour.Magnitude, points)
			}
		}
	}
}

// 追踪依赖的插值误差在事件 ±4 小时内远低于等值线容差 / interpolation error over ±4 h stays far below contour tolerance.
func TestSolarEclipseLocalEphemerisAngularErrorForContours(t *testing.T) {
	center := CalcMoonSHByJDE(JDCalc(2024, 4, 8), 0)
	ephemeris := newSolarEclipseLocalEphemeris(center)
	maxArcsec := 0.0
	for step := -1440; step <= 1440; step++ {
		jde := center + float64(step)*10.0/86400
		sun, moon, ok := ephemeris.equatorialAt(jde)
		if !ok {
			t.Fatalf("interpolator rejected %+.4f d", jde-center)
		}
		wantSun, wantMoon := solarEclipseSunMoonEquatorial(jde)
		for _, pair := range [][2][3]float64{{sun, wantSun}, {moon, wantMoon}} {
			got := solarEclipseLLRToXYZ(pair[0][0], pair[0][1], pair[0][2])
			want := solarEclipseLLRToXYZ(pair[1][0], pair[1][1], pair[1][2])
			cross := [3]float64{got[1]*want[2] - got[2]*want[1], got[2]*want[0] - got[0]*want[2], got[0]*want[1] - got[1]*want[0]}
			dot := got[0]*want[0] + got[1]*want[1] + got[2]*want[2]
			arcsec := math.Atan2(math.Sqrt(cross[0]*cross[0]+cross[1]*cross[1]+cross[2]*cross[2]), dot) / rad * 3600
			maxArcsec = math.Max(maxArcsec, arcsec)
		}
	}
	if maxArcsec > 1e-3 {
		t.Fatalf("max interpolation error %.3e arcsec, want <= 1e-3", maxArcsec)
	}
}
