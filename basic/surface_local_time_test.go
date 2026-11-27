package basic

import (
	"errors"
	"math"
	"testing"

	"github.com/starainrt/astro/internal/ephem"
)

func TestSurfaceLocalTimeSyntheticMeanSun(t *testing.T) {
	const mu = 1.32712440018e11 * 86400 * 86400
	for _, tilt := range []float64{0.1, 25, 98} {
		orientation := BodyOrientation{PoleRA: 270, PoleDec: 90 - tilt}
		for _, e := range []float64{0.01, 0.2, 0.7} {
			for _, anomaly := range []float64{-179, -90, 0, 70, 179} {
				nuRad := anomaly * math.Pi / 180
				axis := 2e8
				parameter := axis * (1 - e*e)
				radius := parameter / (1 + e*math.Cos(nuRad))
				r := Vector3{radius * math.Cos(nuRad), radius * math.Sin(nuRad), 0}
				speed := math.Sqrt(mu / parameter)
				v := Vector3{-speed * math.Sin(nuRad), speed * (e + math.Cos(nuRad)), 0}
				frame, err := bodyEquatorialFrameFromState(orientation, r, v)
				if err != nil {
					t.Fatal(err)
				}
				alpha, longitude, err := surfaceMeanSolarCoordinates(r, v, frame)
				eccentricAnomalyRad := math.Atan2(math.Sqrt(1-e*e)*math.Sin(nuRad), e+math.Cos(nuRad))
				wantMean := 180 + (eccentricAnomalyRad-e*math.Sin(eccentricAnomalyRad))*180/math.Pi
				wantLongitude := anomaly
				if !frame.equinoxDefined {
					wantLongitude -= 180
				}
				if err != nil || math.Abs(angleDifference(alpha, wantMean)) > 1e-10 || math.Abs(angleDifference(longitude, wantLongitude)) > 1e-10 {
					t.Fatalf("tilt=%g e=%g nuRad=%g alpha=%.15g want=%.15g lambda=%.15g want=%.15g err=%v", tilt, e, anomaly, alpha, wantMean, longitude, wantLongitude, err)
				}
			}
		}
	}
}

func TestSurfaceLocalTimeSyntheticCircularOrbit(t *testing.T) {
	const mu = 1.32712440018e11 * 86400 * 86400
	const orbitSpeed = 1 << 21
	const radius = mu / (orbitSpeed * orbitSpeed)
	const epoch = 2451545.0
	engine := surfaceTestEngine{
		positions:  map[ephem.Body]Vector3{ephem.Sun: {}, ephem.Mars: {radius, 0, 0}},
		velocities: map[ephem.Body]Vector3{ephem.Mars: {0, orbitSpeed, 0}},
	}
	inst := ephem.NewInstance(engine)
	observer := BodyObserver{Body: SkyMars}
	clock, err := SurfaceLocalTimeAtWithInstance(observer, epoch, inst)
	if err != nil {
		t.Fatal(err)
	}
	orientation, err := BodyOrientationAt(SkyMars, epoch)
	if err != nil {
		t.Fatal(err)
	}
	wantLongitude := 630 - orientation.PoleRA
	wantMeanTime := 12 + (orientation.PrimeMeridianAngle-(wantLongitude-180))/15
	if !clock.EquinoxDefined || math.Abs(clock.SolarLongitude-wantLongitude) > 1e-10 || math.Abs(clock.LocalMeanSolarTime-wantMeanTime) > 1e-10 {
		t.Fatalf("clock=%+v, want Ls %.12f and LMST %.12f", clock, wantLongitude, wantMeanTime)
	}
	position, err := SurfacePositionAtWithInstance(SkySun, observer, epoch, inst)
	if err != nil {
		t.Fatal(err)
	}
	if want := 12 + position.HourAngle/15; math.Abs(clock.LocalTrueSolarTime-want) > 1e-12 {
		t.Fatalf("LTST %.12f, want %.12f", clock.LocalTrueSolarTime, want)
	}
}

func TestSurfaceLocalTimeConsistencyInstances(t *testing.T) {
	previous := GetEphemerisModel()
	defer SetEphemerisModel(previous)
	for _, model := range []EphemerisModel{EphemerisLegacy, EphemerisSemianalytic} {
		SetEphemerisModel(model)
		inst := ephem.NewModelInstance(ephem.Model(model))
		for body := SkyMercury; body <= SkyNeptune; body++ {
			for _, jd := range []float64{2451545 - 300*365.25, 2451545, 2461319.5, 2451545 + 300*365.25} {
				observer := BodyObserver{Body: body, EastLongitude: 37, Latitude: -25, Height: 2}
				got, err := SurfaceLocalTimeAt(observer, jd)
				want, werr := SurfaceLocalTimeAtWithInstance(observer, jd, inst)
				if err != nil || werr != nil || got != want {
					t.Fatalf("model=%d body=%d jd=%g got=%+v want=%+v err=%v/%v", model, body, jd, got, want, err, werr)
				}
				orientation, _ := BodyOrientationAt(body, jd)
				if got.PrimeMeridianAngle != orientation.PrimeMeridianAngle || got.OutsideValidity != orientation.OutsideValidity || got.EquinoxDefined != (body != SkyMercury) {
					t.Fatalf("flags/orientation: body=%d result=%+v", body, got)
				}
				values := []float64{got.PrimeMeridianAngle, got.SiderealRotationAngle, got.LocalSiderealTime, got.LocalTrueSolarTime, got.LocalMeanSolarTime, got.SolarDayLength, got.SolarLongitude}
				for _, x := range values {
					if !isFiniteFloat(x) {
						t.Fatalf("nonfinite: %+v", got)
					}
				}
				if got.SiderealRotationAngle < 0 || got.SiderealRotationAngle >= 360 || got.LocalSiderealTime < 0 || got.LocalSiderealTime >= 360 || got.SolarLongitude < 0 || got.SolarLongitude >= 360 || got.LocalTrueSolarTime < 0 || got.LocalTrueSolarTime >= 24 || got.LocalMeanSolarTime < 0 || got.LocalMeanSolarTime >= 24 || got.SolarDayLength <= 0 {
					t.Fatalf("range: %+v", got)
				}
				if body == SkyMercury && (got.LocalSiderealTime != 0 || got.SolarLongitude != 0) {
					t.Fatalf("undefined equinox: %+v", got)
				}
				position, err := SurfacePositionAtWithInstance(SkySun, observer, jd, inst)
				if err != nil || math.Abs(math.Remainder(got.LocalTrueSolarTime-(12+position.HourAngle/15), 24)) > 1e-12 {
					t.Fatalf("apparent solar time got=%+v position=%+v err=%v", got, position, err)
				}
				expectedDay := 1 / math.Abs(orientation.RotationRate/360-1/orbitalPeriodDays[body])
				if got.SolarDayLength != expectedDay || math.Abs(angleDifference(got.SiderealRotationAngle, orientation.PrimeMeridianAngle+37)) > 1e-12 {
					t.Fatalf("rotation: %+v", got)
				}
				shifted := observer
				shifted.EastLongitude += 360
				same, err := SurfaceLocalTimeAtWithInstance(shifted, jd, inst)
				if err != nil || math.Abs(math.Remainder(got.LocalMeanSolarTime-same.LocalMeanSolarTime, 24)) > 1e-12 {
					t.Fatalf("longitude wrap: %+v / %+v %v", got, same, err)
				}
			}
		}
	}
}

func TestSurfaceLocalTimeConsistencyRetrograde(t *testing.T) {
	inst := ephem.NewModelInstance(ephem.Semianalytic)
	for _, body := range []SkyBody{SkyVenus, SkyUranus} {
		a, err := SurfaceLocalTimeAtWithInstance(BodyObserver{Body: body}, 2461319.5, inst)
		b, berr := SurfaceLocalTimeAtWithInstance(BodyObserver{Body: body}, 2461319.5001, inst)
		if err != nil || berr != nil || angleDifference(b.SiderealRotationAngle, a.SiderealRotationAngle) >= 0 || angleDifference(b.LocalSiderealTime, a.LocalSiderealTime) >= 0 {
			t.Fatalf("body=%d a=%+v b=%+v err=%v/%v", body, a, b, err, berr)
		}
	}
}

func TestSurfaceLocalTimeSyntheticInvalid(t *testing.T) {
	for _, observer := range []BodyObserver{{}, {Body: SkyEarth}, {Body: SkySun}, {Body: SkyMars, Latitude: 91}, {Body: SkyMars, EastLongitude: math.Inf(1)}, {Body: SkyMars, Height: math.NaN()}} {
		got, err := SurfaceLocalTimeAt(observer, 2451545)
		if !errors.Is(err, ErrInvalidObservationInput) || got != (SurfaceLocalTime{}) {
			t.Fatalf("observer=%+v got=%+v err=%v", observer, got, err)
		}
	}
	for _, jd := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		got, err := SurfaceLocalTimeAt(BodyObserver{Body: SkyMars}, jd)
		if !errors.Is(err, ErrInvalidObservationInput) || got != (SurfaceLocalTime{}) {
			t.Fatalf("jd=%g got=%+v err=%v", jd, got, err)
		}
	}
	inst := ephem.NewInstance(surfaceTestEngine{})
	got, err := SurfaceLocalTimeAtWithInstance(BodyObserver{Body: SkyMars}, 2451545, inst)
	if err == nil || got != (SurfaceLocalTime{}) {
		t.Fatalf("engine error: %+v %v", got, err)
	}
}
