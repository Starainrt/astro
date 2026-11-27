package basic

import (
	"errors"
	"math"
	"testing"

	"github.com/starainrt/astro/internal/ephem"
)

const (
	bodyEquatorialRoundTripTolerance      = 1e-10
	bodyEquatorialSolarLongitudeTolerance = 0.05
)

func TestBodyEquatorialConsistencyDirections(t *testing.T) {
	for _, model := range []ephem.Model{ephem.Legacy, ephem.Semianalytic} {
		inst := ephem.NewModelInstance(model)
		for body := SkyMercury; body <= SkyNeptune; body++ {
			for _, jdTT := range []float64{2451545 - 50*365.25, 2451545, 2451545 + 50*365.25} {
				orientation, err := BodyOrientationAt(body, jdTT)
				if err != nil {
					t.Fatal(err)
				}
				poleRA, poleDec, err := toBodyEquatorial(body, jdTT, orientation.PoleRA, orientation.PoleDec, BodyEquatorOriginIAUNode, inst)
				if err != nil || math.Abs(poleDec-90) > bodyEquatorialRoundTripTolerance {
					t.Fatalf("pole model=%d body=%d RA=%v Dec=%.17g err=%v", model, body, poleRA, poleDec, err)
				}
				nodeRA, nodeDec, err := toBodyEquatorial(body, jdTT, orientation.PoleRA+90, 0, BodyEquatorOriginIAUNode, inst)
				if err != nil || math.Abs(angleDifference(nodeRA, 0)) > bodyEquatorialRoundTripTolerance || math.Abs(nodeDec) > bodyEquatorialRoundTripTolerance {
					t.Fatalf("node body=%d RA=%.17g Dec=%.17g err=%v", body, nodeRA, nodeDec, err)
				}
				frame, err := bodyEquatorialFrameAt(body, jdTT, inst)
				if err != nil {
					t.Fatal(err)
				}
				if frame.equinoxDefined != (body != SkyMercury) {
					t.Fatalf("model=%d body=%d defined=%v", model, body, frame.equinoxDefined)
				}
				if !frame.equinoxDefined {
					_, _, err = toBodyEquatorial(body, jdTT, 10, 20, BodyEquatorOriginEquinox, inst)
					if !errors.Is(err, ErrInvalidObservationInput) {
						t.Fatalf("undefined equinox err=%v", err)
					}
					_, _, err = fromBodyEquatorial(body, jdTT, 10, 20, BodyEquatorOriginEquinox, inst)
					if !errors.Is(err, ErrInvalidObservationInput) {
						t.Fatalf("undefined inverse equinox err=%v", err)
					}
					continue
				}
				equinoxRA := math.Atan2(frame.equinox[1], frame.equinox[0]) * 180 / math.Pi
				equinoxDec := math.Atan2(frame.equinox[2], math.Hypot(frame.equinox[0], frame.equinox[1])) * 180 / math.Pi
				gotRA, gotDec, err := toBodyEquatorial(body, jdTT, equinoxRA, equinoxDec, BodyEquatorOriginEquinox, inst)
				if err != nil || math.Abs(angleDifference(gotRA, 0)) > bodyEquatorialRoundTripTolerance || math.Abs(gotDec) > bodyEquatorialRoundTripTolerance {
					t.Fatalf("equinox body=%d RA=%v Dec=%v err=%v", body, gotRA, gotDec, err)
				}
				qRA, _, err := toBodyEquatorial(body, jdTT, 33, -16, BodyEquatorOriginIAUNode, inst)
				if err != nil {
					t.Fatal(err)
				}
				eqRA, _, err := toBodyEquatorial(body, jdTT, 33, -16, BodyEquatorOriginEquinox, inst)
				if err != nil {
					t.Fatal(err)
				}
				if math.Abs(angleDifference(qRA-eqRA, frame.equinoxAngle)) > bodyEquatorialRoundTripTolerance {
					t.Fatalf("body=%d origins have wrong relative angle", body)
				}
			}
		}
	}
}

func TestBodyEquatorialConsistencyRoundTrip(t *testing.T) {
	maximum := 0.0
	for _, model := range []ephem.Model{ephem.Legacy, ephem.Semianalytic} {
		inst := ephem.NewModelInstance(model)
		for body := SkyMercury; body <= SkyNeptune; body++ {
			for _, origin := range []BodyEquatorOrigin{BodyEquatorOriginIAUNode, BodyEquatorOriginEquinox} {
				if body == SkyMercury && origin == BodyEquatorOriginEquinox {
					continue
				}
				for _, sample := range []struct{ ra, dec float64 }{{0, 0}, {360, 0}, {-720.5, 20}, {10, -89}, {40, 89}, {101.28715533, -16.71611586}, {289.2, 58.1}} {
					ra, dec, err := toBodyEquatorial(body, 2461319.5, sample.ra, sample.dec, origin, inst)
					if err != nil {
						t.Fatal(err)
					}
					if !isFiniteFloat(ra) || !isFiniteFloat(dec) || ra < 0 || ra >= 360 || dec < -90 || dec > 90 {
						t.Fatalf("nonfinite or out of bounds: %v,%v", ra, dec)
					}
					backRA, backDec, err := fromBodyEquatorial(body, 2461319.5, ra, dec, origin, inst)
					if err != nil {
						t.Fatal(err)
					}
					difference := math.Max(math.Abs(angleDifference(backRA, sample.ra)), math.Abs(backDec-sample.dec))
					maximum = math.Max(maximum, difference)
					if difference >= bodyEquatorialRoundTripTolerance {
						t.Fatalf("model=%d body=%d origin=%d roundtrip=%.12g deg", model, body, origin, difference)
					}
					backRA, backDec, err = fromBodyEquatorial(body, 2461319.5, sample.ra, sample.dec, origin, inst)
					if err != nil {
						t.Fatal(err)
					}
					forwardRA, forwardDec, err := toBodyEquatorial(body, 2461319.5, backRA, backDec, origin, inst)
					if err != nil {
						t.Fatal(err)
					}
					difference = math.Max(math.Abs(angleDifference(forwardRA, sample.ra)), math.Abs(forwardDec-sample.dec))
					maximum = math.Max(maximum, difference)
					if difference >= bodyEquatorialRoundTripTolerance {
						t.Fatalf("reverse roundtrip body=%d origin=%d error=%.12g deg", body, origin, difference)
					}
				}
			}
		}
	}
	t.Logf("maximum roundtrip error = %.12g deg", maximum)
}

func TestBodyEquatorialConsistencyDefaultInstance(t *testing.T) {
	previous := GetEphemerisModel()
	defer SetEphemerisModel(previous)
	for _, model := range []EphemerisModel{EphemerisLegacy, EphemerisSemianalytic} {
		SetEphemerisModel(model)
		inst := ephem.NewModelInstance(ephem.Model(model))
		for body := SkyMercury; body <= SkyNeptune; body++ {
			for _, origin := range []BodyEquatorOrigin{BodyEquatorOriginIAUNode, BodyEquatorOriginEquinox} {
				if body == SkyMercury && origin == BodyEquatorOriginEquinox {
					continue
				}
				for _, sample := range []struct {
					public   func(SkyBody, float64, float64, float64, BodyEquatorOrigin) (float64, float64, error)
					explicit func(SkyBody, float64, float64, float64, BodyEquatorOrigin, *Instance) (float64, float64, error)
				}{
					{ToBodyEquatorial, ToBodyEquatorialWithInstance},
					{FromBodyEquatorial, FromBodyEquatorialWithInstance},
				} {
					gotRA, gotDec, err := sample.public(body, 2461319.5, 101.28715533, -16.71611586, origin)
					if err != nil {
						t.Fatal(err)
					}
					wantRA, wantDec, err := sample.explicit(body, 2461319.5, 101.28715533, -16.71611586, origin, inst)
					if err != nil || math.Float64bits(gotRA) != math.Float64bits(wantRA) || math.Float64bits(gotDec) != math.Float64bits(wantDec) {
						t.Fatalf("model=%d body=%d origin=%d default differs from explicit instance: err=%v", model, body, origin, err)
					}
				}
			}
		}
	}
}

func TestBodyEquatorialExternalSolarLongitude(t *testing.T) {
	inst := ephem.NewModelInstance(ephem.Semianalytic)
	maximum := 0.0
	for _, sample := range []struct{ jdTT, ls float64 }{{2451549.50074, 277.18758}, {2453008.07471, 327.32416}} {
		frame, err := bodyEquatorialFrameAt(SkyMars, sample.jdTT, inst)
		if err != nil {
			t.Fatal(err)
		}
		position, _, err := inst.State(ephem.Mars, ephem.Sun, sample.jdTT)
		if err != nil {
			t.Fatal(err)
		}
		sun := Vector3{-position[0], -position[1], -position[2]}
		got := math.Atan2(vectorDot(pxp(frame.equinox, sun), frame.normal), vectorDot(frame.equinox, sun)) * 180 / math.Pi
		difference := angleDifference(got, sample.ls)
		maximum = math.Max(maximum, math.Abs(difference))
		if math.Abs(difference) > bodyEquatorialSolarLongitudeTolerance {
			t.Errorf("jdTT=%.5f Ls=%.9f want=%.5f delta=%.9f deg", sample.jdTT, got, sample.ls, difference)
		}
		t.Logf("jdTT=%.5f Ls=%.9f delta=%+.9f deg", sample.jdTT, math.Mod(got+360, 360), difference)
	}
	t.Logf("maximum solar longitude error = %.9f deg", maximum)
}

func TestBodyEquatorialConsistencySphericalTriangle(t *testing.T) {
	const obliquityTolerance = 0.01
	inst := ephem.NewModelInstance(ephem.Semianalytic)
	maximum := 0.0
	for _, jdTT := range []float64{2451549.50074, 2453008.07471, 2461319.5} {
		position, velocity, err := inst.State(ephem.Mars, ephem.Sun, jdTT)
		if err != nil {
			t.Fatal(err)
		}
		normal := normalizeVector(pxp(Vector3(position), Vector3(velocity)))
		poleRA, poleDec, _ := marsPoleRotation(jdTT)
		pole := raDecToVector(poleRA, poleDec)
		equinox := normalizeVector(pxp(pole, normal))
		sun := normalizeVector(Vector3{-position[0], -position[1], -position[2]})
		ls := math.Atan2(vectorDot(pxp(equinox, sun), normal), vectorDot(equinox, sun))
		obliquity := math.Acos(vectorDot(pole, normal))
		if math.Abs(obliquity*180/math.Pi-25.19) > obliquityTolerance {
			t.Fatalf("jdTT=%.5f obliquity=%.12f deg", jdTT, obliquity*180/math.Pi)
		}
		ra := math.Atan2(sun[1], sun[0]) * 180 / math.Pi
		dec := math.Atan2(sun[2], math.Hypot(sun[0], sun[1])) * 180 / math.Pi
		gotRA, gotDec, err := toBodyEquatorial(SkyMars, jdTT, ra, dec, BodyEquatorOriginEquinox, inst)
		if err != nil {
			t.Fatal(err)
		}
		wantRA := math.Atan2(math.Cos(obliquity)*math.Sin(ls), math.Cos(ls)) * 180 / math.Pi
		wantDec := math.Asin(math.Sin(obliquity)*math.Sin(ls)) * 180 / math.Pi
		raDifference := angleDifference(gotRA, wantRA)
		decDifference := gotDec - wantDec
		difference := math.Max(math.Abs(raDifference), math.Abs(decDifference))
		if !isFiniteFloat(difference) || difference > bodyEquatorialRoundTripTolerance {
			t.Errorf("jdTT=%.5f BodyRA=%.15g want=%.15g delta=%+.12g BodyDec=%.15g want=%.15g delta=%+.12g deg", jdTT, gotRA, wantRA, raDifference, gotDec, wantDec, decDifference)
		}
		maximum = math.Max(maximum, difference)
		t.Logf("jdTT=%.5f obliquity=%.12f deg RA delta=%+.12g Dec delta=%+.12g deg", jdTT, obliquity*180/math.Pi, raDifference, decDifference)
	}
	t.Logf("maximum spherical triangle error = %.12g deg", maximum)
}

func TestBodyEquatorialSyntheticEarthEquinox(t *testing.T) {
	matrix := BiasPrecessionNutation2006(2461319.5)
	pole := Vector3{matrix[2][0], matrix[2][1], matrix[2][2]}
	equinox := Vector3{matrix[0][0], matrix[0][1], matrix[0][2]}
	obliquityRad := 23.4 * math.Pi / 180
	side := pxp(equinox, pole)
	normal := Vector3{}
	for i := range normal {
		normal[i] = math.Cos(obliquityRad)*pole[i] + math.Sin(obliquityRad)*side[i]
	}
	velocity := pxp(normal, equinox)
	orientation := BodyOrientation{PoleRA: math.Atan2(pole[1], pole[0]) * 180 / math.Pi, PoleDec: math.Atan2(pole[2], math.Hypot(pole[0], pole[1])) * 180 / math.Pi}
	frame, err := bodyEquatorialFrameFromState(orientation, equinox, velocity)
	if err != nil {
		t.Fatal(err)
	}
	qAngle := math.Atan2(vectorDot(equinox, pxp(pole, frame.node)), vectorDot(equinox, frame.node)) * 180 / math.Pi
	if !frame.equinoxDefined || math.Abs(angleDifference(frame.equinoxAngle, qAngle)) > bodyEquatorialRoundTripTolerance {
		t.Fatalf("theta=%.15g want=%.15g defined=%v", frame.equinoxAngle, qAngle, frame.equinoxDefined)
	}
}

type bodyEquatorialStateEngine struct {
	position, velocity [3]float64
	err                error
	target             ephem.Body
}

func (e bodyEquatorialStateEngine) State(target, center ephem.Body, jdTT float64) ([3]float64, [3]float64, error) {
	if target != e.target || center != ephem.Sun {
		return [3]float64{}, [3]float64{}, ErrInvalidObservationInput
	}
	return e.position, e.velocity, e.err
}
func (e bodyEquatorialStateEngine) Coverage(ephem.Body) (float64, float64, bool) {
	return 2400000, 2500000, true
}

func TestBodyEquatorialSyntheticStateAndInvalid(t *testing.T) {
	const jdTT = 2451545.0
	for body := SkyMercury; body <= SkyNeptune; body++ {
		target := ephem.Body(body)
		if body <= SkyVenus {
			target--
		}
		inst := ephem.NewInstance(bodyEquatorialStateEngine{position: [3]float64{2, 0, 0}, velocity: [3]float64{0, 3, 0}, target: target})
		frame, err := bodyEquatorialFrameAt(body, jdTT, inst)
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(frame.normal[2]-1) > bodyEquatorialRoundTripTolerance || math.Abs(frame.normal[0])+math.Abs(frame.normal[1]) > bodyEquatorialRoundTripTolerance {
			t.Fatalf("body=%d normal=%v", body, frame.normal)
		}
	}
	sentinel := errors.New("state unavailable")
	for _, engine := range []bodyEquatorialStateEngine{
		{position: [3]float64{1, 0, 0}, velocity: [3]float64{0, 1, 0}, err: sentinel, target: ephem.Mars},
		{target: ephem.Mars},
		{position: [3]float64{math.NaN(), 0, 0}, velocity: [3]float64{0, 1, 0}, target: ephem.Mars},
		{position: [3]float64{1, 0, 0}, velocity: [3]float64{0, math.Inf(1), 0}, target: ephem.Mars},
		{position: [3]float64{1, 0, 0}, velocity: [3]float64{1, 0, 0}, target: ephem.Mars},
	} {
		_, err := bodyEquatorialFrameAt(SkyMars, jdTT, ephem.NewInstance(engine))
		if err == nil || engine.err != nil && !errors.Is(err, sentinel) {
			t.Fatalf("invalid engine err=%v", err)
		}
	}
	for _, sample := range []struct {
		body        SkyBody
		jd, ra, dec float64
		origin      BodyEquatorOrigin
	}{
		{SkyEarth, jdTT, 0, 0, BodyEquatorOriginIAUNode}, {SkySun, jdTT, 0, 0, BodyEquatorOriginIAUNode}, {SkyMoon, jdTT, 0, 0, BodyEquatorOriginIAUNode},
		{SkyMars, math.NaN(), 0, 0, BodyEquatorOriginIAUNode}, {SkyMars, math.Inf(1), 0, 0, BodyEquatorOriginIAUNode},
		{SkyMars, jdTT, math.NaN(), 0, BodyEquatorOriginIAUNode}, {SkyMars, jdTT, 0, math.Inf(-1), BodyEquatorOriginIAUNode},
		{SkyMars, jdTT, 0, 90.1, BodyEquatorOriginIAUNode}, {SkyMars, jdTT, 0, -90.1, BodyEquatorOriginIAUNode}, {SkyMars, jdTT, 0, 0, BodyEquatorOrigin(10)},
	} {
		for _, convert := range []func(SkyBody, float64, float64, float64, BodyEquatorOrigin) (float64, float64, error){ToBodyEquatorial, FromBodyEquatorial} {
			ra, dec, err := convert(sample.body, sample.jd, sample.ra, sample.dec, sample.origin)
			if !errors.Is(err, ErrInvalidObservationInput) || ra != 0 || dec != 0 {
				t.Errorf("input=%+v output=%v,%v err=%v", sample, ra, dec, err)
			}
		}
	}
}
