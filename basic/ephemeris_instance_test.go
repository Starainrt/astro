package basic

import (
	"math"
	"testing"

	"github.com/starainrt/astro/internal/ephem"
)

type p3ProbeEngine struct {
	angle    float64
	distance float64
}

func (engine p3ProbeEngine) State(target, center ephem.Body, jdTDB float64) (pos, vel [3]float64, err error) {
	angle := engine.angle + float64(target-center)*0.01
	distance := engine.distance
	if distance == 0 {
		distance = 1
	}
	pos[0] = math.Cos(angle) * astronomicalUnitKM * distance
	pos[1] = math.Sin(angle) * astronomicalUnitKM * distance
	return pos, vel, nil
}

func (engine p3ProbeEngine) Coverage(target ephem.Body) (start, end float64, ok bool) {
	return 2400000, 2500000, target != ephem.SSB
}

func (engine p3ProbeEngine) Apparent(target ephem.Body, jdTT float64) (dir [3]float64, distKM, tauDays float64, err error) {
	angle := engine.angle + float64(target)*0.01
	dir[0] = math.Cos(angle)
	dir[1] = math.Sin(angle)
	distance := engine.distance
	if distance == 0 {
		distance = 1
	}
	return dir, astronomicalUnitKM * distance, 0.005, nil
}

func TestP3SunMoonFunctionsUseExplicitInstances(t *testing.T) {
	previous := ephem.Active()
	ephem.SetActive(ephem.Legacy)
	t.Cleanup(func() { ephem.SetActive(previous) })

	first := ephem.NewInstance(p3ProbeEngine{angle: 0})
	second := ephem.NewInstance(p3ProbeEngine{angle: math.Pi / 2})
	jd := 2451545.0

	sunFirst := HSunApparentLoNWithInstance(jd, -1, first)
	sunSecond := HSunApparentLoNWithInstance(jd, -1, second)
	if math.Abs(angleDifference(sunFirst, sunSecond)) < 1 {
		t.Fatalf("sun apparent longitude did not follow instance: %.6f vs %.6f", sunFirst, sunSecond)
	}

	moonFirst := HMoonApparentLoNWithInstance(jd, -1, first)
	moonSecond := HMoonApparentLoNWithInstance(jd, -1, second)
	if math.Abs(angleDifference(moonFirst, moonSecond)) < 1 {
		t.Fatalf("moon apparent longitude did not follow instance: %.6f vs %.6f", moonFirst, moonSecond)
	}

	if got, want := HSunApparentLoN(jd, -1), HSunApparentLoNWithInstance(jd, -1, nil); math.Abs(angleDifference(got, want)) > 1e-12 {
		t.Fatalf("nil instance changed legacy-compatible result: %.15f vs %.15f", got, want)
	}
}

func TestP3PlanetFunctionsUseExplicitInstances(t *testing.T) {
	first := ephem.NewInstance(p3ProbeEngine{angle: 0})
	second := ephem.NewInstance(p3ProbeEngine{angle: math.Pi / 2})
	jd := 2451545.0

	for _, sample := range []struct {
		name   string
		first  float64
		second float64
	}{
		{"MercuryL", MercuryLWithInstance(jd, first), MercuryLWithInstance(jd, second)},
		{"VenusL", VenusLWithInstance(jd, first), VenusLWithInstance(jd, second)},
		{"MarsL", MarsLWithInstance(jd, first), MarsLWithInstance(jd, second)},
		{"JupiterL", JupiterLWithInstance(jd, first), JupiterLWithInstance(jd, second)},
		{"SaturnL", SaturnLWithInstance(jd, first), SaturnLWithInstance(jd, second)},
		{"UranusL", UranusLWithInstance(jd, first), UranusLWithInstance(jd, second)},
		{"NeptuneL", NeptuneLWithInstance(jd, first), NeptuneLWithInstance(jd, second)},
		{"MercuryApparentLo", MercuryApparentLoWithInstance(jd, first), MercuryApparentLoWithInstance(jd, second)},
		{"NeptuneApparentLo", NeptuneApparentLoWithInstance(jd, first), NeptuneApparentLoWithInstance(jd, second)},
	} {
		if math.Abs(angleDifference(sample.first, sample.second)) < 1 {
			t.Fatalf("%s did not follow explicit instance: %.6f vs %.6f", sample.name, sample.first, sample.second)
		}
	}

	for _, model := range []ephem.Model{ephem.Legacy, ephem.Semianalytic} {
		inst := ephem.NewModelInstance(model)
		for _, sample := range []struct {
			name  string
			value float64
		}{
			{"MercuryL", MercuryLWithInstance(jd, inst)},
			{"MercuryR", MercuryRWithInstance(jd, inst)},
			{"MercuryApparentRa", MercuryApparentRaWithInstance(jd, inst)},
			{"NeptuneApparentDec", NeptuneApparentDecWithInstance(jd, inst)},
		} {
			if math.IsNaN(sample.value) || math.IsInf(sample.value, 0) {
				t.Fatalf("%s model %d returned non-finite value: %v", sample.name, model, sample.value)
			}
		}
	}

	previous := ephem.Active()
	ephem.SetActive(ephem.Legacy)
	t.Cleanup(func() { ephem.SetActive(previous) })
	if got, want := MercuryLWithInstance(jd, nil), MercuryL(jd); got != want {
		t.Fatalf("nil instance changed MercuryL: %.15f vs %.15f", got, want)
	}
	if got, want := NeptuneApparentRaWithInstance(jd, nil), NeptuneApparentRa(jd); got != want {
		t.Fatalf("nil instance changed NeptuneApparentRa: %.15f vs %.15f", got, want)
	}
}

func TestP3MoonPhaseFunctionsUseExplicitInstances(t *testing.T) {
	first := ephem.NewInstance(p3ProbeEngine{angle: 0})
	second := ephem.NewInstance(p3ProbeEngine{angle: math.Pi / 2})
	jd := 2451545.0

	phaseFirst := MoonPhaseWithInstance(jd, first)
	phaseSecond := MoonPhaseWithInstance(jd, second)
	if math.Abs(phaseFirst-phaseSecond) < 1e-6 {
		t.Fatalf("moon phase did not follow explicit instance: %.12f vs %.12f", phaseFirst, phaseSecond)
	}
	seekFirst := SunMoonSeekWithInstance(jd, 0, first)
	seekSecond := SunMoonSeekWithInstance(jd, 0, second)
	if math.Abs(angleDifference(seekFirst, seekSecond)) < 1 {
		t.Fatalf("sun-moon phase residual did not follow explicit instance: %.6f vs %.6f", seekFirst, seekSecond)
	}

	for _, model := range []ephem.Model{ephem.Legacy, ephem.Semianalytic} {
		inst := ephem.NewModelInstance(model)
		for _, sample := range []struct {
			name  string
			value float64
		}{
			{"CalcMoonSHByJDE", CalcMoonSHByJDEWithInstance(jd, 0, inst)},
			{"CalcMoonXHByJDE", CalcMoonXHByJDEWithInstance(jd, 0, inst)},
			{"MoonPhase", MoonPhaseWithInstance(jd, inst)},
		} {
			if math.IsNaN(sample.value) || math.IsInf(sample.value, 0) {
				t.Fatalf("%s model %d returned non-finite value: %v", sample.name, model, sample.value)
			}
		}
	}

	if got, want := MoonPhaseWithInstance(jd, nil), MoonPhase(jd); got != want {
		t.Fatalf("nil instance changed MoonPhase: %.15f vs %.15f", got, want)
	}
	if got, want := CalcMoonSHByJDEWithInstance(jd, 0, nil), CalcMoonSHByJDE(jd, 0); got != want {
		t.Fatalf("nil instance changed CalcMoonSHByJDE: %.15f vs %.15f", got, want)
	}
}

func TestP3SolarTermFunctionsUseExplicitInstances(t *testing.T) {
	first := ephem.NewInstance(p3ProbeEngine{angle: 0})
	second := ephem.NewInstance(p3ProbeEngine{angle: math.Pi / 2})
	jd := 2451545.0

	firstResidual := JQLospecWithInstance(jd, 0, first)
	secondResidual := JQLospecWithInstance(jd, 0, second)
	if math.Abs(angleDifference(firstResidual, secondResidual)) < 1 {
		t.Fatalf("solar-term longitude did not follow explicit instance: %.6f vs %.6f", firstResidual, secondResidual)
	}
	if GetShiErCiWithInstance(jd, first) == GetShiErCiWithInstance(jd, second) {
		t.Fatalf("twelve-division label did not follow explicit instance")
	}

	for _, model := range []ephem.Model{ephem.Legacy, ephem.Semianalytic} {
		inst := ephem.NewModelInstance(model)
		for _, sample := range []struct {
			name  string
			value float64
		}{
			{"GetJQTime", GetJQTimeWithInstance(2024, 270, inst)},
			{"GetWuHouTime", GetWuHouTimeWithInstance(2024, 0, inst)},
		} {
			if math.IsNaN(sample.value) || math.IsInf(sample.value, 0) {
				t.Fatalf("%s model %d returned non-finite value: %v", sample.name, model, sample.value)
			}
		}
		if got := GetJieqiLoopsWithInstance(2024, 3, inst); len(got) != 3 {
			t.Fatalf("GetJieqiLoops model %d returned %d values", model, len(got))
		}
	}

	if got, want := JQLospecWithInstance(jd, 0, nil), JQLospec(jd, 0); got != want {
		t.Fatalf("nil instance changed JQLospec: %.15f vs %.15f", got, want)
	}
	if got, want := GetJQTimeWithInstance(2024, 270, nil), GetJQTime(2024, 270); got != want {
		t.Fatalf("nil instance changed GetJQTime: %.15f vs %.15f", got, want)
	}
}

func TestP3SunObservationFunctionsUseExplicitInstances(t *testing.T) {
	first := ephem.NewInstance(p3ProbeEngine{angle: 0})
	second := ephem.NewInstance(p3ProbeEngine{angle: math.Pi / 2})
	jd := 2451545.0

	firstHeight := SunHeightWithInstance(jd, 121.47, 31.23, 8, first)
	secondHeight := SunHeightWithInstance(jd, 121.47, 31.23, 8, second)
	if math.Abs(firstHeight-secondHeight) < 1 {
		t.Fatalf("sun height did not follow explicit instance: %.6f vs %.6f", firstHeight, secondHeight)
	}
	firstAzimuth := SunAzimuthWithInstance(jd, 121.47, 31.23, 8, first)
	secondAzimuth := SunAzimuthWithInstance(jd, 121.47, 31.23, 8, second)
	if math.Abs(angleDifference(firstAzimuth, secondAzimuth)) < 1 {
		t.Fatalf("sun azimuth did not follow explicit instance: %.6f vs %.6f", firstAzimuth, secondAzimuth)
	}

	for _, model := range []ephem.Model{ephem.Legacy, ephem.Semianalytic} {
		inst := ephem.NewModelInstance(model)
		for _, sample := range []struct {
			name  string
			value float64
		}{
			{"SunTime", SunTimeWithInstance(jd, inst)},
			{"CulminationTime", CulminationTimeWithInstance(jd, 121.47, 8, inst)},
			{"SunTimeAngle", SunTimeAngleWithInstance(jd, 121.47, 31.23, 8, inst)},
			{"SunHeight", SunHeightWithInstance(jd, 121.47, 31.23, 8, inst)},
			{"SunAzimuth", SunAzimuthWithInstance(jd, 121.47, 31.23, 8, inst)},
		} {
			if math.IsNaN(sample.value) || math.IsInf(sample.value, 0) {
				t.Fatalf("%s model %d returned non-finite value: %v", sample.name, model, sample.value)
			}
		}
	}

	if got, want := SunHeightWithInstance(jd, 121.47, 31.23, 8, nil), SunHeight(jd, 121.47, 31.23, 8); got != want {
		t.Fatalf("nil instance changed SunHeight: %.15f vs %.15f", got, want)
	}
	if got, want := SunAzimuthWithInstance(jd, 121.47, 31.23, 8, nil), SunAzimuth(jd, 121.47, 31.23, 8); got != want {
		t.Fatalf("nil instance changed SunAzimuth: %.15f vs %.15f", got, want)
	}
}

// 显式 Legacy 实例的均时差与中天时刻必须复现包级 Legacy 结果。
func TestSunTimeWithLegacyInstanceMatchesPackageLevel(t *testing.T) {
	previous := ephem.Active()
	ephem.SetActive(ephem.Legacy)
	t.Cleanup(func() { ephem.SetActive(previous) })
	inst := ephem.NewModelInstance(ephem.Legacy)
	for _, jd := range []float64{2451545.0, 2460384.5, 2460569.5} {
		if got, want := SunTimeWithInstance(jd, inst), SunTime(jd); math.Abs(got-want) > 1e-9 {
			t.Fatalf("SunTime jd %.1f: instance %.9f h, package %.9f h", jd, got, want)
		}
		if got, want := SunTimeNWithInstance(jd, 20, inst), SunTimeN(jd, 20); math.Abs(got-want) > 1e-9 {
			t.Fatalf("SunTimeN jd %.1f: instance %.9f h, package %.9f h", jd, got, want)
		}
		if got, want := CulminationTimeWithInstance(jd, 116.4, 8, inst), CulminationTime(jd, 116.4, 8); math.Abs(got-want)*86400 > 1e-3 {
			t.Fatalf("CulminationTime jd %.1f: differs by %.3f s", jd, (got-want)*86400)
		}
	}
}

func TestP3SunMoonAliasFunctionsUseExplicitInstances(t *testing.T) {
	first := ephem.NewInstance(p3ProbeEngine{angle: 0})
	second := ephem.NewInstance(p3ProbeEngine{angle: math.Pi / 2})
	jd := 2451545.0

	if math.Abs(angleDifference(SunApparentRaWithInstance(jd, first), SunApparentRaWithInstance(jd, second))) < 1 {
		t.Fatal("SunApparentRa alias did not follow explicit instance")
	}
	if math.Abs(angleDifference(MoonGeocentricApparentRaWithInstance(jd, first), MoonGeocentricApparentRaWithInstance(jd, second))) < 1 {
		t.Fatal("MoonGeocentricApparentRa alias did not follow explicit instance")
	}

	for _, model := range []ephem.Model{ephem.Legacy, ephem.Semianalytic} {
		inst := ephem.NewModelInstance(model)
		for _, sample := range []struct {
			name  string
			value float64
		}{
			{"SunApparentRa", SunApparentRaWithInstance(jd, inst)},
			{"SunTrueDec", SunTrueDecWithInstance(jd, inst)},
			{"MoonGeocentricTrueRa", MoonGeocentricTrueRaWithInstance(jd, inst)},
			{"MoonTrueDec", MoonTrueDecWithInstance(jd, inst)},
		} {
			if math.IsNaN(sample.value) || math.IsInf(sample.value, 0) {
				t.Fatalf("%s model %d returned non-finite value: %v", sample.name, model, sample.value)
			}
		}
	}

	if got, want := SunApparentRaWithInstance(jd, nil), SunApparentRa(jd); got != want {
		t.Fatalf("nil instance changed SunApparentRa: %.15f vs %.15f", got, want)
	}
	if got, want := MoonGeocentricTrueRaWithInstance(jd, nil), HMoonGeocentricTrueRa(jd); got != want {
		t.Fatalf("nil instance changed MoonGeocentricTrueRa: %.15f vs %.15f", got, want)
	}
}

func TestP3PlanetMagnitudeFunctionsUseExplicitInstances(t *testing.T) {
	first := ephem.NewInstance(p3ProbeEngine{angle: 0, distance: 0.8})
	second := ephem.NewInstance(p3ProbeEngine{angle: math.Pi / 2, distance: 1.2})
	jd := 2451545.0

	for _, sample := range []struct {
		name   string
		first  float64
		second float64
	}{
		{"MercuryMag", MercuryMagWithInstance(jd, first), MercuryMagWithInstance(jd, second)},
		{"VenusMag", VenusMagWithInstance(jd, first), VenusMagWithInstance(jd, second)},
		{"MarsMag", MarsMagWithInstance(jd, first), MarsMagWithInstance(jd, second)},
		{"JupiterMag", JupiterMagWithInstance(jd, first), JupiterMagWithInstance(jd, second)},
		{"UranusMag", UranusMagWithInstance(jd, first), UranusMagWithInstance(jd, second)},
		{"NeptuneMag", NeptuneMagWithInstance(jd, first), NeptuneMagWithInstance(jd, second)},
	} {
		if math.Abs(sample.first-sample.second) < 1e-6 {
			t.Fatalf("%s did not follow explicit instance: %.12f vs %.12f", sample.name, sample.first, sample.second)
		}
	}

	for _, model := range []ephem.Model{ephem.Legacy, ephem.Semianalytic} {
		inst := ephem.NewModelInstance(model)
		for _, sample := range []struct {
			name  string
			value float64
		}{
			{"MercuryMag", MercuryMagWithInstance(jd, inst)},
			{"JupiterMag", JupiterMagWithInstance(jd, inst)},
			{"NeptuneMag", NeptuneMagWithInstance(jd, inst)},
		} {
			if math.IsNaN(sample.value) || math.IsInf(sample.value, 0) {
				t.Fatalf("%s model %d returned non-finite value: %v", sample.name, model, sample.value)
			}
		}
	}

	if got, want := MercuryMagWithInstance(jd, nil), MercuryMag(jd); got != want {
		t.Fatalf("nil instance changed MercuryMag: %.15f vs %.15f", got, want)
	}
}

func TestP3SaturnRingFunctionsUseExplicitInstances(t *testing.T) {
	first := ephem.NewInstance(p3ProbeEngine{angle: 0, distance: 0.8})
	second := ephem.NewInstance(p3ProbeEngine{angle: math.Pi / 2, distance: 1.2})
	jd := 2451545.0

	firstRing := SaturnRingBWithInstance(jd, first)
	secondRing := SaturnRingBWithInstance(jd, second)
	if math.Abs(firstRing-secondRing) < 1e-6 {
		t.Fatalf("Saturn ring opening angle did not follow explicit instance: %.12f vs %.12f", firstRing, secondRing)
	}
	firstMagnitude := SaturnMagWithInstance(jd, first)
	secondMagnitude := SaturnMagWithInstance(jd, second)
	if math.Abs(firstMagnitude-secondMagnitude) < 1e-6 {
		t.Fatalf("Saturn magnitude did not follow explicit instance: %.12f vs %.12f", firstMagnitude, secondMagnitude)
	}

	for _, model := range []ephem.Model{ephem.Legacy, ephem.Semianalytic} {
		inst := ephem.NewModelInstance(model)
		for _, sample := range []struct {
			name  string
			value float64
		}{
			{"SaturnRingB", SaturnRingBWithInstance(jd, inst)},
			{"SaturnRingPositionAngle", SaturnRingPositionAngleWithInstance(jd, inst)},
			{"SaturnMag", SaturnMagWithInstance(jd, inst)},
		} {
			if math.IsNaN(sample.value) || math.IsInf(sample.value, 0) {
				t.Fatalf("%s model %d returned non-finite value: %v", sample.name, model, sample.value)
			}
		}
	}

	if got, want := SaturnMagWithInstance(jd, nil), SaturnMag(jd); got != want {
		t.Fatalf("nil instance changed SaturnMag: %.15f vs %.15f", got, want)
	}
}

func angleDifference(a, b float64) float64 {
	difference := math.Mod(a-b+180, 360)
	if difference < 0 {
		difference += 360
	}
	return difference - 180
}
