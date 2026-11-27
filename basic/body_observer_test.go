package basic

import (
	"errors"
	"math"
	"testing"
)

const (
	bodyObserverAngleTolerance    = 1e-10
	bodyObserverPositionTolerance = 1e-10
	bodyObserverUnitTolerance     = 1e-12
)

func TestBodyObserverSyntheticEllipsoid(t *testing.T) {
	maximum := 0.0
	orientation := BodyOrientation{PoleRA: -90, PoleDec: 90}
	for _, axes := range []struct{ a, b float64 }{{5, 3}, {4, 4}} {
		model := planetPhysicalModel{equatorialRadius: axes.a, polarRadius: axes.b}
		for _, sample := range []struct{ longitude, latitude float64 }{{0, 0}, {90, 0}, {60, 30}, {-42, -53}, {137, 89.9999999}, {15, 90}, {-35, -90}} {
			lon, lat := sample.longitude*math.Pi/180, sample.latitude*math.Pi/180
			up := Vector3{math.Cos(lat) * math.Cos(lon), math.Cos(lat) * math.Sin(lon), math.Sin(lat)}
			denominator := math.Sqrt(axes.a*axes.a*(up[0]*up[0]+up[1]*up[1]) + axes.b*axes.b*up[2]*up[2])
			surface := Vector3{axes.a * axes.a * up[0] / denominator, axes.a * axes.a * up[1] / denominator, axes.b * axes.b * up[2] / denominator}
			gradient := normalizeVector(Vector3{surface[0] / (axes.a * axes.a), surface[1] / (axes.a * axes.a), surface[2] / (axes.b * axes.b)})
			north := Vector3{-math.Sin(lat) * math.Cos(lon), -math.Sin(lat) * math.Sin(lon), math.Cos(lat)}
			if math.Abs(sample.latitude) == 90 {
				north = Vector3{1, 0, 0}
			}
			east := pxp(north, up)
			for _, height := range []float64{-2, 0, 2} {
				observer := BodyObserver{Body: SkyMars, EastLongitude: sample.longitude, Latitude: sample.latitude, Height: height}
				frame := bodyObserverFrameFromModel(observer, model, orientation)
				for i := range surface {
					difference := math.Abs(frame.positionKM[i] - surface[i] - height*up[i])
					maximum = math.Max(maximum, difference)
					if !isFiniteFloat(difference) || difference > bodyObserverPositionTolerance {
						t.Fatalf("axes=%+v observer=%+v position[%d]=%.17g want=%.17g", axes, observer, i, frame.positionKM[i], surface[i]+height*up[i])
					}
					if math.Abs(frame.up[i]-gradient[i]) > bodyObserverUnitTolerance || math.Abs(frame.north[i]-north[i]) > bodyObserverUnitTolerance || math.Abs(frame.east[i]-east[i]) > bodyObserverUnitTolerance {
						t.Fatalf("observer=%+v up=%v north=%v east=%v", observer, frame.up, frame.north, frame.east)
					}
				}
			}
		}
	}
	t.Logf("maximum independent ellipsoid position error = %.12g km", maximum)
}

func TestBodyObserverSyntheticPolarFrame(t *testing.T) {
	model := planetPhysicalModel{equatorialRadius: 5, polarRadius: 3}
	for _, orientation := range []BodyOrientation{{PoleRA: -90, PoleDec: 90}, {PoleRA: 63, PoleDec: 21, PrimeMeridianAngle: 37}} {
		alpha, delta, w := orientation.PoleRA*math.Pi/180, orientation.PoleDec*math.Pi/180, orientation.PrimeMeridianAngle*math.Pi/180
		pole := Vector3{math.Cos(delta) * math.Cos(alpha), math.Cos(delta) * math.Sin(alpha), math.Sin(delta)}
		prime := Vector3{-math.Sin(alpha)*math.Cos(w) - math.Sin(delta)*math.Cos(alpha)*math.Sin(w), math.Cos(alpha)*math.Cos(w) - math.Sin(delta)*math.Sin(alpha)*math.Sin(w), math.Cos(delta) * math.Sin(w)}
		for _, sign := range []float64{-1, 1} {
			for _, longitude := range []float64{0, 43, -137, 360} {
				frame := bodyObserverFrameFromModel(BodyObserver{EastLongitude: longitude, Latitude: sign * 90, Height: 2}, model, orientation)
				for i := range pole {
					if math.Abs(frame.up[i]-sign*pole[i]) > bodyObserverUnitTolerance || math.Abs(frame.north[i]-prime[i]) > bodyObserverUnitTolerance || math.Abs(frame.positionKM[i]-sign*5*pole[i]) > bodyObserverPositionTolerance {
						t.Fatalf("orientation=%+v sign=%g longitude=%g frame=%+v", orientation, sign, longitude, frame)
					}
				}
				if math.Abs(vectorDot(frame.east, frame.north)) > bodyObserverUnitTolerance || math.Abs(vectorDot(pxp(frame.east, frame.north), frame.up)-1) > bodyObserverUnitTolerance {
					t.Fatalf("polar frame is not right-handed: %+v", frame)
				}
			}
		}
	}
}

func TestBodyObserverConsistencyFrames(t *testing.T) {
	maximum := 0.0
	for body := SkyMercury; body <= SkyNeptune; body++ {
		for _, jdTT := range []float64{2451545, 2461319.5} {
			orientation, err := BodyOrientationAt(body, jdTT)
			if err != nil {
				t.Fatal(err)
			}
			pole := raDecToVector(orientation.PoleRA, orientation.PoleDec)
			for _, latitude := range []float64{-90, -89.9999999, -45, 0, 30, 89.9999999, 90} {
				observer := BodyObserver{Body: body, EastLongitude: -222.6, Latitude: latitude}
				frame, err := bodyObserverFrameAt(observer, jdTT)
				if err != nil || frame.orientation != orientation {
					t.Fatalf("observer=%+v frame=%+v err=%v", observer, frame, err)
				}
				for _, axis := range []Vector3{frame.east, frame.north, frame.up} {
					for _, component := range axis {
						if !isFiniteFloat(component) {
							t.Fatalf("nonfinite frame: %+v", frame)
						}
					}
					if math.Abs(vectorDot(axis, axis)-1) > bodyObserverUnitTolerance {
						t.Fatalf("nonunit frame: %+v", frame)
					}
				}
				if math.Abs(vectorDot(frame.east, frame.north)) > bodyObserverUnitTolerance || math.Abs(vectorDot(frame.north, frame.up)) > bodyObserverUnitTolerance || math.Abs(vectorDot(frame.up, frame.east)) > bodyObserverUnitTolerance || math.Abs(vectorDot(pxp(frame.east, frame.north), frame.up)-1) > bodyObserverUnitTolerance {
					t.Fatalf("nonorthogonal frame: %+v", frame)
				}
				if math.Abs(vectorDot(frame.up, pole)-math.Sin(latitude*math.Pi/180)) > bodyObserverUnitTolerance {
					t.Fatalf("up is not the planetodetic normal: observer=%+v frame=%+v", observer, frame)
				}
				if math.Abs(latitude) != 90 && math.Abs(vectorDot(frame.north, pole)-math.Cos(latitude*math.Pi/180)) > bodyObserverUnitTolerance {
					t.Fatalf("north has wrong sense: observer=%+v frame=%+v", observer, frame)
				}
				observer.Height = 5
				raised, err := bodyObserverFrameAt(observer, jdTT)
				if err != nil || raised.up != frame.up || raised.north != frame.north || raised.east != frame.east {
					t.Fatalf("height changed axes: err=%v", err)
				}
				for i := range frame.positionKM {
					difference := math.Abs(raised.positionKM[i] - frame.positionKM[i] - 5*frame.up[i])
					maximum = math.Max(maximum, difference)
					if !isFiniteFloat(difference) || difference > bodyObserverPositionTolerance {
						t.Fatalf("height shift error=%.12g km", difference)
					}
				}
			}
		}
	}
	t.Logf("maximum height displacement error = %.12g km", maximum)
}

func TestBodyObserverSyntheticConversions(t *testing.T) {
	maximum := 0.0
	for _, sample := range []struct {
		body SkyBody
		east bool
		a, b float64
	}{
		{SkyMercury, false, 2440.53, 2438.26}, {SkyVenus, true, 6051.8, 6051.8}, {SkyMars, false, 3396.19, 3376.20},
		{SkyJupiter, false, 71492, 66854}, {SkySaturn, false, 60268, 54364}, {SkyUranus, true, 25559, 24973}, {SkyNeptune, false, 24764, 24341},
	} {
		for _, longitude := range []struct{ input, east, west float64 }{{0, 0, 0}, {45, 45, 315}, {-45, 315, 45}, {360, 0, 0}, {405, 45, 315}, {-720, 0, 0}} {
			want := longitude.west
			if sample.east {
				want = longitude.east
			}
			for _, convert := range []func(SkyBody, float64) (float64, error){PlanetographicToEastLongitude, EastToPlanetographicLongitude} {
				got, err := convert(sample.body, longitude.input)
				if err != nil || !isFiniteFloat(got) || got != want || got < 0 || got >= 360 {
					t.Fatalf("body=%d longitude=%g got=%g want=%g err=%v", sample.body, longitude.input, got, want, err)
				}
			}
			got, err := PlanetographicToEastLongitude(sample.body, longitude.input)
			if err != nil {
				t.Fatal(err)
			}
			back, err := EastToPlanetographicLongitude(sample.body, got)
			wantBack := math.Mod(longitude.input+720, 360)
			if err != nil || back != wantBack {
				t.Fatalf("longitude roundtrip body=%d got=%g want=%g err=%v", sample.body, back, wantBack, err)
			}
		}
		for _, latitude := range []float64{-90, -89.9999999, -60, -30, math.Copysign(0, -1), 0, 30, 60, 89.9999999, 90} {
			got, err := PlanetocentricToPlanetodeticLatitude(sample.body, latitude)
			if err != nil {
				t.Fatal(err)
			}
			want := math.Atan2(sample.a*sample.a*math.Sin(latitude*math.Pi/180), sample.b*sample.b*math.Cos(latitude*math.Pi/180)) * 180 / math.Pi
			if !isFiniteFloat(got) || math.Abs(got-want) > bodyObserverAngleTolerance || got < -90 || got > 90 {
				t.Fatalf("body=%d latitude=%g got=%.17g want=%.17g", sample.body, latitude, got, want)
			}
			if math.Abs(latitude) == 90 && math.Float64bits(got) != math.Float64bits(latitude) {
				t.Fatalf("body=%d forward pole not preserved", sample.body)
			}
			back, err := PlanetodeticToPlanetocentricLatitude(sample.body, got)
			if err != nil || !isFiniteFloat(back) || math.Abs(back-latitude) > bodyObserverAngleTolerance {
				t.Fatalf("latitude roundtrip body=%d got=%.17g want=%.17g err=%v", sample.body, back, latitude, err)
			}
			maximum = math.Max(maximum, math.Abs(back-latitude))
			got, err = PlanetodeticToPlanetocentricLatitude(sample.body, latitude)
			want = math.Atan2(sample.b*sample.b*math.Sin(latitude*math.Pi/180), sample.a*sample.a*math.Cos(latitude*math.Pi/180)) * 180 / math.Pi
			if err != nil || !isFiniteFloat(got) || math.Abs(got-want) > bodyObserverAngleTolerance {
				t.Fatalf("inverse body=%d latitude=%g got=%.17g want=%.17g err=%v", sample.body, latitude, got, want, err)
			}
			back, err = PlanetocentricToPlanetodeticLatitude(sample.body, got)
			if err != nil || math.Abs(back-latitude) > bodyObserverAngleTolerance {
				t.Fatalf("inverse roundtrip body=%d got=%.17g want=%.17g err=%v", sample.body, back, latitude, err)
			}
			maximum = math.Max(maximum, math.Abs(back-latitude))
			if math.Abs(latitude) == 90 && (math.Float64bits(got) != math.Float64bits(latitude) || math.Float64bits(back) != math.Float64bits(latitude)) {
				t.Fatalf("body=%d pole not preserved", sample.body)
			}
		}
	}
	t.Logf("maximum latitude roundtrip error = %.12g deg", maximum)
}

func TestBodyObserverConsistencyPhysicalConventions(t *testing.T) {
	for body := SkyMercury; body <= SkyNeptune; body++ {
		model, _, _, _, _ := bodyOrientationModel(body)
		for _, longitude := range []float64{0, 45, 137.4, 315} {
			got, err := EastToPlanetographicLongitude(body, longitude)
			want := bodyLongitude(raDecToVector(longitude, 0), Vector3{1, 0, 0}, Vector3{0, 1, 0}, model.positiveEast)
			if err != nil || math.Abs(angleDifference(got, want)) > bodyObserverAngleTolerance {
				t.Fatalf("body=%d longitude=%g got=%g want=%g err=%v", body, longitude, got, want, err)
			}
		}
		for _, latitude := range []float64{-60, -30, 0, 30, 60} {
			got, err := PlanetocentricToPlanetodeticLatitude(body, latitude)
			want := bodyLatitude(raDecToVector(0, latitude), Vector3{0, 0, 1}, model.equatorialRadius, model.polarRadius)
			if err != nil || math.Abs(got-want) > bodyObserverAngleTolerance {
				t.Fatalf("body=%d latitude=%g got=%g want=%g err=%v", body, latitude, got, want, err)
			}
		}
	}
}

func TestBodyObserverSyntheticInvalid(t *testing.T) {
	for _, observer := range []BodyObserver{
		{}, {Body: SkySun}, {Body: SkyMoon}, {Body: SkyEarth}, {Body: SkyBody(-1)}, {Body: SkyBody(100)},
		{Body: SkyMars, Latitude: -90.001}, {Body: SkyMars, Latitude: 90.001}, {Body: SkyMars, Latitude: math.NaN()},
		{Body: SkyMars, EastLongitude: math.Inf(1)}, {Body: SkyMars, EastLongitude: math.NaN()},
		{Body: SkyMars, Height: math.Inf(-1)}, {Body: SkyMars, Height: math.NaN()},
	} {
		if _, err := validateBodyObserver(observer); !errors.Is(err, ErrInvalidObservationInput) {
			t.Fatalf("observer=%+v err=%v", observer, err)
		}
		frame, err := bodyObserverFrameAt(observer, 2451545)
		if !errors.Is(err, ErrInvalidObservationInput) || frame != (bodyObserverFrame{}) {
			t.Fatalf("observer=%+v frame=%+v err=%v", observer, frame, err)
		}
	}
	for _, jdTT := range []float64{math.NaN(), math.Inf(-1), math.Inf(1)} {
		frame, err := bodyObserverFrameAt(BodyObserver{Body: SkyMars}, jdTT)
		if !errors.Is(err, ErrInvalidObservationInput) || frame != (bodyObserverFrame{}) {
			t.Fatalf("jdTT=%g frame=%+v err=%v", jdTT, frame, err)
		}
	}
	for body := SkyMercury; body <= SkyNeptune; body++ {
		if _, err := validateBodyObserver(BodyObserver{Body: body, EastLongitude: 1080.5, Height: -4}); err != nil {
			t.Fatalf("finite observer rejected: body=%d err=%v", body, err)
		}
	}
	for _, convert := range []func(SkyBody, float64) (float64, error){PlanetographicToEastLongitude, EastToPlanetographicLongitude, PlanetocentricToPlanetodeticLatitude, PlanetodeticToPlanetocentricLatitude} {
		for _, body := range []SkyBody{SkySun, SkyMoon, SkyEarth, SkyBody(-1), SkyBody(100)} {
			got, err := convert(body, 30)
			if !errors.Is(err, ErrInvalidObservationInput) || got != 0 {
				t.Fatalf("body=%d got=%g err=%v", body, got, err)
			}
		}
		for _, value := range []float64{math.NaN(), math.Inf(-1), math.Inf(1)} {
			got, err := convert(SkyMars, value)
			if !errors.Is(err, ErrInvalidObservationInput) || got != 0 {
				t.Fatalf("value=%g got=%g err=%v", value, got, err)
			}
		}
	}
	for _, convert := range []func(SkyBody, float64) (float64, error){PlanetocentricToPlanetodeticLatitude, PlanetodeticToPlanetocentricLatitude} {
		for _, latitude := range []float64{-90.001, 90.001} {
			got, err := convert(SkyMars, latitude)
			if !errors.Is(err, ErrInvalidObservationInput) || got != 0 {
				t.Fatalf("latitude=%g got=%g err=%v", latitude, got, err)
			}
		}
	}
}
