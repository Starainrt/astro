package basic

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/internal/ephem"
)

const jupiterRotationMeridianTolerance = 0.02

func jupiterRotationObservationForTest(jdTT float64, inst *Instance) (deltaI, deltaII, tauDays float64) {
	heliocentric := func(index int, jd float64) Vector3 {
		longitudeRad := PlanetCoordinateNWithInstance(index, 0, jd, -1, inst) * math.Pi / 180
		latitudeRad := PlanetCoordinateNWithInstance(index, 1, jd, -1, inst) * math.Pi / 180
		radius := PlanetCoordinateNWithInstance(index, 2, jd, -1, inst)
		return Vector3{radius * math.Cos(latitudeRad) * math.Cos(longitudeRad), radius * math.Cos(latitudeRad) * math.Sin(longitudeRad), radius * math.Sin(latitudeRad)}
	}
	earth := heliocentric(-1, jdTT)
	distance := 4.0
	var jupiter, geocentric Vector3
	for k := 0; k < 2; k++ {
		jupiter = heliocentric(4, jdTT-0.0057755183*distance)
		for i := range geocentric {
			geocentric[i] = jupiter[i] - earth[i]
		}
		distance = vectorMagnitude(geocentric)
	}
	phaseRad := (1 - vectorDot(jupiter, geocentric)/(vectorMagnitude(jupiter)*distance)) / 2
	if jupiter[1]*earth[0]-jupiter[0]*earth[1] < 0 {
		phaseRad = -phaseRad
	}
	meridians := JupiterCentralMeridiansNWithInstance(jdTT, -1, inst)
	return math.Remainder(meridians.SystemI-phaseRad*180/math.Pi-meridians.SystemIII, 360), math.Remainder(meridians.SystemII-phaseRad*180/math.Pi-meridians.SystemIII, 360), 0.0057755183 * distance
}

func TestJupiterRotationConsistencyCentralMeridians(t *testing.T) {
	inst := ephem.NewModelInstance(ephem.Semianalytic)
	d1, d2, tau := jupiterRotationObservationForTest(2451545, inst)
	t.Logf("J2000 c_I=%.12f c_II=%.12f tau=%.15f days", d1+(877.90-870.536)*tau, d2+(870.27-870.536)*tau, tau)
	start := Date2JD(time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC))
	end := Date2JD(time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC))
	minimum := [2]float64{math.Inf(1), math.Inf(1)}
	maximum := [2]float64{math.Inf(-1), math.Inf(-1)}
	count := 0
	for jd := start; jd <= end; jd += 10 {
		first, second, tau := jupiterRotationObservationForTest(jd, inst)
		third, err := JupiterRotationAngleWithInstance(jd-tau, JupiterSystemIII, inst)
		if err != nil {
			t.Fatal(err)
		}
		for i, system := range []JupiterRotationSystem{JupiterSystemI, JupiterSystemII} {
			rotation, err := JupiterRotationAngleWithInstance(jd-tau, system, inst)
			if err != nil {
				t.Fatal(err)
			}
			residual := math.Remainder([]float64{first, second}[i]-(rotation-third), 360)
			minimum[i] = math.Min(minimum[i], residual)
			maximum[i] = math.Max(maximum[i], residual)
			if math.Abs(residual) > jupiterRotationMeridianTolerance {
				t.Fatalf("jd=%g system=%d residual=%.12f deg", jd, system, residual)
			}
		}
		count++
	}
	t.Logf("1900-2100 step10 days samples=%d SystemI=[%+.12f,%+.12f] SystemII=[%+.12f,%+.12f] deg", count, minimum[0], maximum[0], minimum[1], maximum[1])
}

func TestJupiterRotationSyntheticAngles(t *testing.T) {
	for _, sample := range []struct {
		system     JupiterRotationSystem
		zero, rate float64
	}{{JupiterSystemI, 67.103988083590, 877.90}, {JupiterSystemII, 43.357000999787, 870.27}, {JupiterSystemIII, 284.95, 870.536}} {
		for _, d := range []float64{-1, -.25, 0, .25, 1} {
			got, err := JupiterRotationAngle(2451545+d, sample.system)
			want := math.Mod(sample.zero+sample.rate*d, 360)
			if want < 0 {
				want += 360
			}
			if err != nil || math.Abs(math.Remainder(got-want, 360)) > 1e-10 {
				t.Fatalf("system=%d day=%g got=%.15g want=%.15g err=%v", sample.system, d, got, want, err)
			}
		}
	}
}

func TestJupiterRotationConsistencyInstancesAndTimeScale(t *testing.T) {
	previous := GetEphemerisModel()
	defer SetEphemerisModel(previous)
	for _, model := range []EphemerisModel{EphemerisLegacy, EphemerisSemianalytic} {
		SetEphemerisModel(model)
		inst := ephem.NewModelInstance(ephem.Model(model))
		for _, jd := range []float64{2451545 - 300*365.25, 2451545, 2461319.5, 2451545 + 300*365.25} {
			for _, system := range []JupiterRotationSystem{JupiterSystemI, JupiterSystemII, JupiterSystemIII} {
				got, err := JupiterRotationAngle(jd, system)
				want, werr := JupiterRotationAngleWithInstance(jd, system, inst)
				if err != nil || werr != nil || got != want || !isFiniteFloat(got) || got < 0 || got >= 360 {
					t.Fatalf("jd=%g system=%d got=%g want=%g err=%v/%v", jd, system, got, want, err, werr)
				}
				if system == JupiterSystemIII {
					orientation, _ := BodyOrientationAt(SkyJupiter, jd)
					if got != orientation.PrimeMeridianAngle {
						t.Fatalf("SystemIII != W: %g/%g", got, orientation.PrimeMeridianAngle)
					}
				}
			}
		}
	}
	jdUTC := Date2JD(time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC))
	jdTT := UTC2TT(jdUTC)
	for _, system := range []JupiterRotationSystem{JupiterSystemI, JupiterSystemII, JupiterSystemIII} {
		correct, err := JupiterRotationAngle(jdTT, system)
		wrong, werr := JupiterRotationAngle(jdUTC, system)
		difference := math.Abs(math.Remainder(correct-wrong, 360))
		if err != nil || werr != nil || difference <= 0.5 {
			t.Fatalf("TT/UTC system=%d difference=%.12g deg err=%v/%v", system, difference, err, werr)
		}
		t.Logf("TT/UTC system=%d difference=%.12f deg TT-UTC=%.6f s", system, difference, (jdTT-jdUTC)*86400)
	}
}

func TestJupiterRotationSyntheticInvalid(t *testing.T) {
	for _, jd := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		got, err := JupiterRotationAngle(jd, JupiterSystemIII)
		if got != 0 || !errors.Is(err, ErrInvalidObservationInput) {
			t.Fatalf("jd=%g got=%g err=%v", jd, got, err)
		}
	}
	for _, system := range []JupiterRotationSystem{-1, 3, 99} {
		got, err := JupiterRotationAngle(2451545, system)
		if got != 0 || !errors.Is(err, ErrInvalidObservationInput) {
			t.Fatalf("system=%d got=%g err=%v", system, got, err)
		}
	}
}
