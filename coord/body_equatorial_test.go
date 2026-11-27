package coord

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

const bodyEquatorialRoundTripTolerance = 1e-10

func TestBodyEquatorialConsistencyTTWrappers(t *testing.T) {
	previous := basic.GetEphemerisModel()
	defer basic.SetEphemerisModel(previous)
	for _, model := range []basic.EphemerisModel{basic.EphemerisLegacy, basic.EphemerisSemianalytic} {
		basic.SetEphemerisModel(model)
		for body := basic.SkyMercury; body <= basic.SkyNeptune; body++ {
			for _, origin := range []BodyEquatorOrigin{BodyEquatorOriginIAUNode, BodyEquatorOriginEquinox} {
				if body == basic.SkyMercury && origin == BodyEquatorOriginEquinox {
					continue
				}
				for _, date := range []time.Time{time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC), time.Date(2026, 11, 13, 6, 7, 8, 900000000, time.UTC), time.Date(2050, 1, 1, 12, 0, 0, 0, time.UTC)} {
					jdTT := basic.UTC2TT(basic.Date2JD(date.UTC()))
					for _, sample := range []struct {
						public func(basic.SkyBody, time.Time, float64, float64, BodyEquatorOrigin) (Equatorial, error)
						core   func(basic.SkyBody, float64, float64, float64, basic.BodyEquatorOrigin) (float64, float64, error)
					}{{ToBodyEquatorial, basic.ToBodyEquatorial}, {FromBodyEquatorial, basic.FromBodyEquatorial}} {
						got, err := sample.public(body, date, 101.28715533, -16.71611586, origin)
						if err != nil {
							t.Fatal(err)
						}
						wantRA, wantDec, err := sample.core(body, jdTT, 101.28715533, -16.71611586, origin)
						if err != nil || math.Float64bits(got.RA) != math.Float64bits(wantRA) || math.Float64bits(got.Dec) != math.Float64bits(wantDec) {
							t.Fatalf("model=%d body=%d origin=%d got=%+v want=(%.17g,%.17g) err=%v", model, body, origin, got, wantRA, wantDec, err)
						}
						same, err := sample.public(body, date.In(time.FixedZone("east", 8*3600)), 101.28715533, -16.71611586, origin)
						if err != nil || math.Float64bits(same.RA) != math.Float64bits(got.RA) || math.Float64bits(same.Dec) != math.Float64bits(got.Dec) {
							t.Fatalf("timezone changed result: %+v %+v err=%v", got, same, err)
						}
						if math.IsNaN(got.RA) || math.IsNaN(got.Dec) || math.IsInf(got.RA, 0) || math.IsInf(got.Dec, 0) || got.RA < 0 || got.RA >= 360 || got.Dec < -90 || got.Dec > 90 {
							t.Fatalf("nonfinite or out of bounds: %+v", got)
						}
					}
				}
			}
		}
	}
}

func TestBodyEquatorialConsistencyUTCIsNotTT(t *testing.T) {
	date := time.Date(2026, 11, 13, 6, 7, 8, 0, time.UTC)
	for _, origin := range []BodyEquatorOrigin{BodyEquatorOriginIAUNode, BodyEquatorOriginEquinox} {
		for _, sample := range []struct {
			name   string
			public func(basic.SkyBody, time.Time, float64, float64, BodyEquatorOrigin) (Equatorial, error)
			core   func(basic.SkyBody, float64, float64, float64, basic.BodyEquatorOrigin) (float64, float64, error)
		}{{"to", ToBodyEquatorial, basic.ToBodyEquatorial}, {"from", FromBodyEquatorial, basic.FromBodyEquatorial}} {
			got, err := sample.public(basic.SkyJupiter, date, 101.28715533, -16.71611586, origin)
			if err != nil {
				t.Fatal(err)
			}
			wrongRA, wrongDec, err := sample.core(basic.SkyJupiter, jdUTC(date), 101.28715533, -16.71611586, origin)
			if err != nil {
				t.Fatal(err)
			}
			if math.Float64bits(got.RA) == math.Float64bits(wrongRA) && math.Float64bits(got.Dec) == math.Float64bits(wrongDec) {
				t.Fatalf("%s origin=%d cannot distinguish UTC and TT", sample.name, origin)
			}
			t.Logf("%s origin=%d delta TT-UTC RA=%+.12g Dec=%+.12g deg", sample.name, origin, got.RA-wrongRA, got.Dec-wrongDec)
		}
	}
}

func TestBodyEquatorialConsistencyPublicRoundTrip(t *testing.T) {
	maximum := 0.0
	date := time.Date(2026, 11, 13, 6, 7, 8, 0, time.UTC)
	for body := basic.SkyMercury; body <= basic.SkyNeptune; body++ {
		for _, origin := range []BodyEquatorOrigin{BodyEquatorOriginIAUNode, BodyEquatorOriginEquinox} {
			if body == basic.SkyMercury && origin == BodyEquatorOriginEquinox {
				continue
			}
			for _, sample := range []Equatorial{{0, 0}, {101.28715533, -16.71611586}, {300, 89}, {10, -89}} {
				forward, err := ToBodyEquatorial(body, date, sample.RA, sample.Dec, origin)
				if err != nil {
					t.Fatal(err)
				}
				back, err := FromBodyEquatorial(body, date, forward.RA, forward.Dec, origin)
				if err != nil {
					t.Fatal(err)
				}
				raDifference := math.Mod(back.RA-sample.RA+540, 360) - 180
				difference := math.Max(math.Abs(raDifference), math.Abs(back.Dec-sample.Dec))
				maximum = math.Max(maximum, difference)
				if difference >= bodyEquatorialRoundTripTolerance {
					t.Fatalf("body=%d origin=%d delta=%.12g deg", body, origin, difference)
				}
			}
		}
	}
	t.Logf("maximum public roundtrip error = %.12g deg", maximum)
}

func TestBodyEquatorialSyntheticPublicInvalid(t *testing.T) {
	date := time.Date(2026, 11, 13, 6, 7, 8, 0, time.UTC)
	for _, sample := range []struct {
		body    basic.SkyBody
		ra, dec float64
		origin  BodyEquatorOrigin
	}{
		{basic.SkySun, 0, 0, BodyEquatorOriginIAUNode}, {basic.SkyMoon, 0, 0, BodyEquatorOriginIAUNode}, {basic.SkyEarth, 0, 0, BodyEquatorOriginIAUNode},
		{basic.SkyMercury, 0, 0, BodyEquatorOriginEquinox}, {basic.SkyMars, math.NaN(), 0, BodyEquatorOriginIAUNode}, {basic.SkyMars, 0, math.Inf(1), BodyEquatorOriginIAUNode}, {basic.SkyMars, 0, 91, BodyEquatorOriginIAUNode}, {basic.SkyMars, 0, 0, BodyEquatorOrigin(-1)},
	} {
		for _, convert := range []func(basic.SkyBody, time.Time, float64, float64, BodyEquatorOrigin) (Equatorial, error){ToBodyEquatorial, FromBodyEquatorial} {
			got, err := convert(sample.body, date, sample.ra, sample.dec, sample.origin)
			if !errors.Is(err, basic.ErrInvalidObservationInput) || got != (Equatorial{}) {
				t.Errorf("input=%+v got=%+v err=%v", sample, got, err)
			}
		}
	}
}
