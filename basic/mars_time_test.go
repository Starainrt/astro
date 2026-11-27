package basic

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"testing"

	"github.com/starainrt/astro/internal/ephem"
)

const (
	marsSolDateTolerance                = 1e-5
	marsCoordinatedTimeToleranceSeconds = 1
	marsMeanTimeToleranceSeconds        = 60
	marsEquationOfTimeToleranceSeconds  = 60
	marsLongitudeToleranceDegrees       = 0.05
)

func marsConventionalSolarTime(jdTT float64) (longitude, equationOfTime float64) {
	d := jdTT - 2451545
	meanRad := (19.3871 + 0.52402073*d) * math.Pi / 180
	meanLongitude := 270.3871 + 0.524038496*d
	perturbation := 0.0
	for _, term := range [][3]float64{{.0071, 2.2353, 49.409}, {.0057, 2.7543, 168.173}, {.0039, 1.1177, 191.837}, {.0037, 15.7866, 21.736}, {.0021, 2.1354, 15.704}, {.0020, 2.4694, 95.528}, {.0018, 32.8493, 49.095}} {
		perturbation += term[0] * math.Cos((.985626*d/term[1]+term[2])*math.Pi/180)
	}
	center := (10.691+3e-7*d)*math.Sin(meanRad) + .623*math.Sin(2*meanRad) + .050*math.Sin(3*meanRad) + .005*math.Sin(4*meanRad) + .0005*math.Sin(5*meanRad) + perturbation
	longitude = math.Mod(meanLongitude+center, 360)
	if longitude < 0 {
		longitude += 360
	}
	lsRad := longitude * math.Pi / 180
	equationOfTime = 2.861*math.Sin(2*lsRad) - .071*math.Sin(4*lsRad) + .002*math.Sin(6*lsRad) - center
	return
}

func TestMarsTimeExternalSamples(t *testing.T) {
	raw, err := os.ReadFile("testdata/mars_time_samples.json")
	if err != nil {
		t.Fatal(err)
	}
	var data struct {
		Samples []struct {
			Label, Source, Version string
			Retrieved              string  `json:"retrieved_at"`
			JDTT                   float64 `json:"jd_tt"`
			East                   float64 `json:"east_longitude_degrees"`
			MSD                    float64 `json:"msd_sol"`
			Derived                bool    `json:"msd_derived"`
			MTC                    float64 `json:"mtc_hours"`
			LMST                   float64 `json:"lmst_hours"`
			LTST                   float64 `json:"ltst_hours"`
			Ls                     float64 `json:"solar_longitude_degrees"`
			EOT                    float64 `json:"eot_degrees"`
		}
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatal(err)
	}
	if len(data.Samples) != 2 {
		t.Fatalf("sample count=%d", len(data.Samples))
	}
	inst := ephem.NewModelInstance(ephem.Semianalytic)
	var maximum [6]float64
	for _, sample := range data.Samples {
		if sample.Label == "" || sample.Source == "" || sample.Version == "" || sample.Retrieved == "" || !sample.Derived {
			t.Fatalf("sample metadata=%+v", sample)
		}
		msd, err := MarsSolDate(sample.JDTT)
		if err != nil {
			t.Fatal(err)
		}
		mtc, err := MarsCoordinatedTime(sample.JDTT)
		if err != nil {
			t.Fatal(err)
		}
		mean, err := MarsLocalMeanSolarTime(sample.JDTT, sample.East)
		if err != nil {
			t.Fatal(err)
		}
		local, err := SurfaceLocalTimeAtWithInstance(BodyObserver{Body: SkyMars, EastLongitude: sample.East}, sample.JDTT, inst)
		if err != nil {
			t.Fatal(err)
		}
		trueTime, err := MarsLocalTrueSolarTimeWithInstance(sample.JDTT, sample.East, inst)
		if err != nil {
			t.Fatal(err)
		}
		longitude, err := MarsSolarLongitudeWithInstance(sample.JDTT, inst)
		if err != nil {
			t.Fatal(err)
		}
		delta := [6]float64{math.Abs(msd - sample.MSD), math.Remainder(mtc-sample.MTC, 24) * 3600, math.Remainder(local.LocalMeanSolarTime-sample.LMST, 24) * 3600, math.Remainder((trueTime-local.LocalMeanSolarTime)-sample.EOT/15, 24) * 3600, math.Remainder(longitude-sample.Ls, 360), math.Remainder(trueTime-sample.LTST, 24) * 3600}
		for i, x := range delta {
			maximum[i] = math.Max(maximum[i], math.Abs(x))
		}
		if math.Abs(delta[0]) > marsSolDateTolerance || math.Abs(delta[1]) > marsCoordinatedTimeToleranceSeconds || math.Abs(delta[2]) > marsMeanTimeToleranceSeconds || math.Abs(delta[3]) > marsEquationOfTimeToleranceSeconds || math.Abs(delta[4]) > marsLongitudeToleranceDegrees || math.Abs(math.Remainder(mean-sample.LMST, 24))*3600 > marsCoordinatedTimeToleranceSeconds {
			t.Errorf("%s deltas MSD=%.12g sol MTC=%+.9g s LMST=%+.9g s EOT=%+.9g s Ls=%+.9g deg LTST=%+.9g s", sample.Label, delta[0], delta[1], delta[2], delta[3], delta[4], delta[5])
		}
		t.Logf("%s LMST=%+.9f s LTST=%+.9f s EOT=%+.9f s Ls=%+.12f deg", sample.Label, delta[2], delta[5], delta[3], delta[4])
	}
	t.Logf("sample maxima MSD=%.12g sol MTC=%.9f s LMST=%.9f s EOT=%.9f s Ls=%.12f deg LTST=%.9f s", maximum[0], maximum[1], maximum[2], maximum[3], maximum[4], maximum[5])
}

func TestMarsTimeExternalConventionalGrid(t *testing.T) {
	inst := ephem.NewModelInstance(ephem.Semianalytic)
	var maximum [3]float64
	count := 0
	for jd := 2451545.0; jd <= 2451545+200*365.25; jd += 37.3 {
		conventionalLs, eot := marsConventionalSolarTime(jd)
		local, err := SurfaceLocalTimeAtWithInstance(BodyObserver{Body: SkyMars}, jd, inst)
		if err != nil {
			t.Fatal(err)
		}
		mean, err := MarsLocalMeanSolarTime(jd, 0)
		if err != nil {
			t.Fatal(err)
		}
		delta := [3]float64{math.Abs(math.Remainder(local.LocalMeanSolarTime-mean, 24)) * 3600, math.Abs(math.Remainder((local.LocalTrueSolarTime-local.LocalMeanSolarTime)-eot/15, 24)) * 3600, math.Abs(math.Remainder(local.SolarLongitude-conventionalLs, 360))}
		limits := [3]float64{marsMeanTimeToleranceSeconds, marsEquationOfTimeToleranceSeconds, marsLongitudeToleranceDegrees}
		for i, x := range delta {
			maximum[i] = math.Max(maximum[i], x)
			if x > limits[i] {
				t.Fatalf("jd=%.9f item=%d delta=%.12g limit=%g", jd, i, x, limits[i])
			}
		}
		count++
	}
	t.Logf("2000-2200 step37.3 days samples=%d maximum LMST=%.9f s EOT=%.9f s Ls=%.12f deg", count, maximum[0], maximum[1], maximum[2])
}

func TestMarsTimeConsistencyInstances(t *testing.T) {
	previous := GetEphemerisModel()
	defer SetEphemerisModel(previous)
	for _, model := range []EphemerisModel{EphemerisLegacy, EphemerisSemianalytic} {
		SetEphemerisModel(model)
		inst := ephem.NewModelInstance(ephem.Model(model))
		for _, jd := range []float64{2451545 - 300*365.25, 2451545, 2461319.5, 2451545 + 300*365.25} {
			ls, err := MarsSolarLongitude(jd)
			want, werr := MarsSolarLongitudeWithInstance(jd, inst)
			if err != nil || werr != nil || ls != want || !isFiniteFloat(ls) || ls < 0 || ls >= 360 {
				t.Fatalf("longitude=%g/%g err=%v/%v", ls, want, err, werr)
			}
			for _, east := range []float64{-720, -184.702, 0, 175.298, 720} {
				got, err := MarsLocalTrueSolarTime(jd, east)
				want, werr := MarsLocalTrueSolarTimeWithInstance(jd, east, inst)
				local, lerr := SurfaceLocalTimeAtWithInstance(BodyObserver{Body: SkyMars, EastLongitude: east}, jd, inst)
				mean, merr := MarsLocalMeanSolarTime(jd, east)
				mtc, terr := MarsCoordinatedTime(jd)
				if err != nil || werr != nil || lerr != nil || merr != nil || terr != nil || got != want || got != local.LocalTrueSolarTime || !isFiniteFloat(got) || got < 0 || got >= 24 || !isFiniteFloat(mean) || mean < 0 || mean >= 24 || math.Abs(math.Remainder(mean-mtc-east/15, 24)) > 1e-12 {
					t.Fatalf("jd=%g east=%g got=%g want=%g local=%+v mean=%g err=%v/%v/%v/%v/%v", jd, east, got, want, local, mean, err, werr, lerr, merr, terr)
				}
			}
		}
	}
}

func TestMarsTimeSyntheticInvalid(t *testing.T) {
	for _, x := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		for _, f := range []func(float64) (float64, error){MarsSolDate, MarsCoordinatedTime, MarsSolarLongitude} {
			got, err := f(x)
			if got != 0 || !errors.Is(err, ErrInvalidObservationInput) {
				t.Fatalf("invalid=%g got=%g err=%v", x, got, err)
			}
		}
		for _, f := range []func(float64, float64) (float64, error){MarsLocalMeanSolarTime, MarsLocalTrueSolarTime} {
			for _, args := range [][2]float64{{x, 0}, {2451545, x}} {
				got, err := f(args[0], args[1])
				if got != 0 || !errors.Is(err, ErrInvalidObservationInput) {
					t.Fatalf("args=%v got=%g err=%v", args, got, err)
				}
			}
		}
	}
	engine := ephem.NewInstance(surfaceTestEngine{})
	a, err := MarsLocalTrueSolarTimeWithInstance(2451545, 0, engine)
	if err == nil || a != 0 {
		t.Fatalf("engine: %g %v", a, err)
	}
	b, err := MarsSolarLongitudeWithInstance(2451545, engine)
	if err == nil || b != 0 {
		t.Fatalf("engine: %g %v", b, err)
	}
	zeroMSD := 2451549.5 - (44796.0-0.0009626)*1.0274912517
	for _, jd := range []float64{math.Nextafter(zeroMSD, 0), zeroMSD} {
		mtc, err := MarsCoordinatedTime(jd)
		if err != nil || mtc < 0 || mtc >= 24 {
			t.Fatalf("jd=%.17g mtc=%.17g err=%v", jd, mtc, err)
		}
	}
}
