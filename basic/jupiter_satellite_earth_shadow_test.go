package basic

import (
	"math"
	"testing"
)

// 地球视向影斑覆盖比例对照：独立在被食星球面上取经纬网格，按面元朝地球的投影面积加权，不复用扫描与锥函数。
func TestJupiterGalileanEarthShadowConsistencySphereIntegral(t *testing.T) {
	cases := []struct {
		name string
		seed float64
		a, b int
	}{
		{"2026-05-31 1-2", JDCalc(2026, 5, 31), 1, 2},
		{"2026-05-11 1-2", JDCalc(2026, 5, 11), 1, 2},
		{"2026-11-15 1-4", JDCalc(2026, 11, 15), 1, 4},
		{"2021-08-22 2-3", JDCalc(2021, 8, 22), 2, 3},
		{"2026-11-16 1-4 antumbra", JDCalc(2026, 11, 15.9), 1, 4},
	}
	if testing.Short() {
		cases = cases[:1]
	}
	worst := 0.0
	for _, test := range cases {
		event := NextJupiterGalileanMutualEventWithInstance(test.seed, test.a, test.b, JupiterGalileanMutualEclipse, nil)
		if !event.Valid {
			t.Fatalf("%s fixture missing", test.name)
		}
		for _, fraction := range []float64{0.2, 0.5, 0.8} {
			jde := event.Start + fraction*(event.End-event.Start)
			got, ok := JupiterGalileanEarthShadowWithInstance(jde, test.a, test.b, nil)
			if !ok {
				t.Fatal("earth shadow missing")
			}
			umbra, antumbra, shadow, phase := independentEarthShadow(t, jde, event.Active, event.Passive)
			if math.Abs(got.PhaseAngle-phase) > 1e-6 {
				t.Fatalf("%s phase %.9f vs %.9f", test.name, got.PhaseAngle, phase)
			}
			for _, pair := range [][2]float64{{got.UmbraCoverage, umbra}, {got.AntumbraCoverage, antumbra}, {got.ShadowCoverage, shadow}} {
				worst = math.Max(worst, math.Abs(pair[0]-pair[1]))
				if math.Abs(pair[0]-pair[1]) > 0.01 {
					t.Fatalf("%s f=%.1f coverage %.4f vs independent %.4f", test.name, fraction, pair[0], pair[1])
				}
			}
		}
	}
	t.Logf("maximum coverage difference %.4f", worst)
}

func independentEarthShadow(t *testing.T, jde float64, active, passive int) (float64, float64, float64, float64) {
	t.Helper()
	receptionFromEmission := func(emission float64) float64 {
		return jupiterGalileanEclipseReception(emission, passive, nil)
	}
	low, high := jde-0.1, jde
	for iteration := 0; iteration < 60; iteration++ {
		middle := (low + high) / 2
		if receptionFromEmission(middle) < jde {
			low = middle
		} else {
			high = middle
		}
	}
	emission := (low + high) / 2
	jupiter := rotateEclipticToEquatorial(jupiterHeliocentricVectorJ2000(emission, nil), orbitJ2000Obliquity)
	stateOf := func(index int) Vector3 {
		state := jupiterGalileanSatelliteCorrectedStateAtET(emission-jupiterGalileanReferenceJD, index-1, jupiterGalileanUseSolarLongPeriod(emission))
		return Vector3{jupiter[0] + state.X, jupiter[1] + state.Y, jupiter[2] + state.Z}
	}
	a, p := stateOf(active), stateOf(passive)
	earth := rotateEclipticToEquatorial(earthHeliocentricVectorJ2000(jde, nil), orbitJ2000Obliquity)
	toEarth := normalizeVector(Vector3{earth[0] - p[0], earth[1] - p[1], earth[2] - p[2]})
	toSun := normalizeVector(Vector3{-p[0], -p[1], -p[2]})
	radii := [...]float64{1829.4, 1562.6, 2631.2, 2410.3}
	ra, rp := radii[active-1]/astronomicalUnitKM, radii[passive-1]/astronomicalUnitKM
	sunRadius := 695700.0 / astronomicalUnitKM
	axis := normalizeVector(a)
	distance := math.Sqrt(a[0]*a[0] + a[1]*a[1] + a[2]*a[2])
	var total, umbra, antumbra, shadow float64
	const latitudes, longitudes = 180, 360
	for i := 0; i < latitudes; i++ {
		theta := (float64(i) + 0.5) / latitudes * math.Pi
		for j := 0; j < longitudes; j++ {
			phi := (float64(j) + 0.5) / longitudes * 2 * math.Pi
			normal := Vector3{math.Sin(theta) * math.Cos(phi), math.Sin(theta) * math.Sin(phi), math.Cos(theta)}
			facing := normal[0]*toEarth[0] + normal[1]*toEarth[1] + normal[2]*toEarth[2]
			if facing <= 0 {
				continue
			}
			weight := facing * math.Sin(theta)
			total += weight
			if normal[0]*toSun[0]+normal[1]*toSun[1]+normal[2]*toSun[2] <= 0 {
				continue
			}
			point := Vector3{p[0] + rp*normal[0], p[1] + rp*normal[1], p[2] + rp*normal[2]}
			delta := Vector3{point[0] - a[0], point[1] - a[1], point[2] - a[2]}
			along := delta[0]*axis[0] + delta[1]*axis[1] + delta[2]*axis[2]
			if along <= 0 {
				continue
			}
			perpendicular := math.Sqrt(math.Max(0, delta[0]*delta[0]+delta[1]*delta[1]+delta[2]*delta[2]-along*along))
			umbraRadius := ra - along*(sunRadius-ra)/distance
			penumbraRadius := ra + along*(sunRadius+ra)/distance
			if umbraRadius > 0 && perpendicular < umbraRadius {
				umbra += weight
				shadow += weight
			} else if umbraRadius < 0 && perpendicular < -umbraRadius {
				antumbra += weight
				shadow += weight
			} else if perpendicular < penumbraRadius {
				shadow += weight
			}
		}
	}
	phase := math.Acos(math.Max(-1, math.Min(1, toEarth[0]*toSun[0]+toEarth[1]*toSun[1]+toEarth[2]*toSun[2]))) * 180 / math.Pi
	return umbra / total, antumbra / total, shadow / total, phase
}

func TestJupiterGalileanEarthShadowSyntheticInvalid(t *testing.T) {
	for _, pair := range [][2]int{{0, 1}, {2, 2}, {1, 5}} {
		if got, ok := JupiterGalileanEarthShadowWithInstance(JDCalc(2026, 1, 1), pair[0], pair[1], nil); ok || got.Runs != nil {
			t.Fatalf("invalid pair %v accepted", pair)
		}
	}
	if _, ok := JupiterGalileanEarthShadowWithInstance(math.NaN(), 1, 2, nil); ok {
		t.Fatal("NaN accepted")
	}
}
