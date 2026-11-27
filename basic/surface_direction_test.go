package basic

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/internal/ephem"
)

type surfaceTestEngine struct {
	positions, velocities map[ephem.Body]Vector3
}

func (e surfaceTestEngine) State(target, center ephem.Body, jdTT float64) (p, v [3]float64, err error) {
	targetPosition, ok := e.positions[target]
	centerPosition, centerOK := e.positions[center]
	if !ok || !centerOK {
		return p, v, ephem.ErrUnsupportedBody
	}
	for i := range p {
		v[i] = e.velocities[target][i] - e.velocities[center][i]
		p[i] = targetPosition[i] - centerPosition[i] + v[i]*(jdTT-2451545)
	}
	return p, v, nil
}

func (surfaceTestEngine) Coverage(ephem.Body) (float64, float64, bool) {
	return 2400000, 2500000, true
}

func surfaceIndependentAberration(direction, velocity Vector3) Vector3 {
	const c = 299792.458 * 86400
	speed := math.Sqrt(velocity[0]*velocity[0] + velocity[1]*velocity[1] + velocity[2]*velocity[2])
	if speed == 0 {
		return direction
	}
	beta := speed / c
	gamma := 1 / math.Sqrt(1-beta*beta)
	axis := Vector3{velocity[0] / speed, velocity[1] / speed, velocity[2] / speed}
	parallel := direction[0]*axis[0] + direction[1]*axis[1] + direction[2]*axis[2]
	var result Vector3
	for i := range result {
		result[i] = (direction[i] + ((gamma-1)*parallel+gamma*beta)*axis[i]) / (gamma * (1 + beta*parallel))
	}
	length := math.Sqrt(result[0]*result[0] + result[1]*result[1] + result[2]*result[2])
	for i := range result {
		result[i] /= length
	}
	return result
}

func TestSurfaceDirectionConsistencyInstances(t *testing.T) {
	previous := GetEphemerisModel()
	defer SetEphemerisModel(previous)
	for _, model := range []EphemerisModel{EphemerisLegacy, EphemerisSemianalytic} {
		SetEphemerisModel(model)
		inst := ephem.NewModelInstance(ephem.Model(model))
		for body := SkyMercury; body <= SkyNeptune; body++ {
			for _, target := range []SkyBody{SkySun, SkyEarth, SkyMercury, SkyVenus, SkyMars, SkyJupiter, SkySaturn, SkyUranus, SkyNeptune} {
				if body == target {
					continue
				}
				for _, jd := range []float64{2451545, 2461319.5} {
					observer := BodyObserver{Body: body, EastLongitude: 30, Latitude: -25, Height: 2}
					got, err := surfaceDirectionAt(target, observer, jd, nil)
					want, wantErr := surfaceDirectionAt(target, observer, jd, inst)
					if err != nil || wantErr != nil || got != want {
						t.Fatalf("model=%d observer=%d target=%d jd=%g got=%+v want=%+v err=%v/%v", model, body, target, jd, got, want, err, wantErr)
					}
					length := math.Sqrt(vectorDot(got.direction, got.direction))
					if !isFiniteFloat(length) || math.Abs(length-1) > 1e-15 || !(got.distanceKM > 0) || !(got.tauDays > 0) || !isFiniteFloat(got.distanceKM) || !isFiniteFloat(got.tauDays) {
						t.Fatalf("invalid output: %+v", got)
					}
				}
			}
		}
	}
}

func TestSurfaceDirectionSyntheticSpinAberration(t *testing.T) {
	engine := surfaceTestEngine{positions: map[ephem.Body]Vector3{ephem.SSB: {}, ephem.Sun: {2e8, -3e8, 1e8}}, velocities: map[ephem.Body]Vector3{}}
	maximum := 0.0
	for _, sample := range []struct {
		body SkyBody
		id   ephem.Body
		a    float64
	}{{SkyMercury, ephem.Mercury, 2440.53}, {SkyVenus, ephem.Venus, 6051.8}, {SkyMars, ephem.Mars, 3396.19}, {SkyJupiter, ephem.Jupiter, 71492}, {SkySaturn, ephem.Saturn, 60268}, {SkyUranus, ephem.Uranus, 25559}, {SkyNeptune, ephem.Neptune, 24764}} {
		engine.positions[sample.id] = Vector3{1e8, -2e8, 3e7}
		engine.velocities[sample.id] = Vector3{-5e5, 1e6, 3e4}
		for _, latitude := range []float64{0, 90, -90} {
			observer := BodyObserver{Body: sample.body, EastLongitude: 30, Latitude: latitude, Height: 2}
			frame, err := bodyObserverFrameAt(observer, 2451545)
			if err != nil {
				t.Fatal(err)
			}
			wantVelocity := engine.velocities[sample.id]
			if latitude == 0 {
				for i := range wantVelocity {
					wantVelocity[i] += frame.east[i] * (sample.a + observer.Height) * frame.orientation.RotationRate * math.Pi / 180
				}
			}
			var geometric Vector3
			for i := range geometric {
				geometric[i] = engine.positions[ephem.Sun][i] - engine.positions[sample.id][i] - frame.positionKM[i]
			}
			distance := vectorMagnitude(geometric)
			for i := range geometric {
				geometric[i] /= distance
			}
			want := surfaceIndependentAberration(geometric, wantVelocity)
			got, err := surfaceDirectionAt(SkySun, observer, 2451545, ephem.NewInstance(engine))
			if err != nil || math.Abs(got.distanceKM-distance) > 1e-7 || math.Abs(got.tauDays-distance/(299792.458*86400)) > 1e-15 {
				t.Fatalf("body=%d lat=%g got=%+v distance=%g err=%v", sample.body, latitude, got, distance, err)
			}
			for i := range want {
				maximum = math.Max(maximum, math.Abs(got.direction[i]-want[i]))
				if math.Abs(got.direction[i]-want[i]) > 1e-15 {
					t.Fatalf("body=%d lat=%g direction=%v want=%v", sample.body, latitude, got.direction, want)
				}
			}
		}
	}
	t.Logf("independent spin/aberration maximum component difference = %.12g", maximum)
}

func TestSurfaceDirectionSyntheticLegacySpin(t *testing.T) {
	inst := ephem.NewModelInstance(ephem.Legacy)
	const jd = 2461319.5
	maximum := 0.0
	for _, body := range []SkyBody{SkyVenus, SkyJupiter, SkyUranus} {
		observer := BodyObserver{Body: body, EastLongitude: 70, Height: 3}
		model, _, _, _, _ := bodyOrientationModel(body)
		frame, err := bodyObserverFrameAt(observer, jd)
		if err != nil {
			t.Fatal(err)
		}
		position := bodyHeliocentricPositionJ2000(model.planetIndex, jd, inst)
		before := bodyHeliocentricPositionJ2000(model.planetIndex, jd-0.01, inst)
		after := bodyHeliocentricPositionJ2000(model.planetIndex, jd+0.01, inst)
		var direction, velocity Vector3
		for i := range direction {
			direction[i] = -position[i] - frame.positionKM[i]
			velocity[i] = (after[i]-before[i])/0.02 + frame.east[i]*(model.equatorialRadius+observer.Height)*frame.orientation.RotationRate*math.Pi/180
		}
		distance := vectorMagnitude(direction)
		for i := range direction {
			direction[i] /= distance
		}
		want := surfaceIndependentAberration(direction, velocity)
		got, err := surfaceDirectionAt(SkySun, observer, jd, inst)
		if err != nil || got.distanceKM != distance {
			t.Fatalf("body=%d distance=%g want=%g err=%v", body, got.distanceKM, distance, err)
		}
		for i := range want {
			maximum = math.Max(maximum, math.Abs(got.direction[i]-want[i]))
			if math.Abs(got.direction[i]-want[i]) > 1e-15 {
				t.Fatalf("body=%d got=%v want=%v", body, got.direction, want)
			}
		}
	}
	t.Logf("legacy independent spin/aberration maximum component difference = %.12g", maximum)
}

func TestSurfaceDirectionConsistencyModels(t *testing.T) {
	legacy, vector := ephem.NewModelInstance(ephem.Legacy), ephem.NewModelInstance(ephem.Semianalytic)
	for body := SkyMercury; body <= SkyNeptune; body++ {
		maximum := 0.0
		for _, target := range []SkyBody{SkySun, SkyEarth, SkyJupiter} {
			if body == target {
				continue
			}
			observer := BodyObserver{Body: body, EastLongitude: 30, Latitude: 25}
			a, err := surfaceDirectionAt(target, observer, 2461319.5, legacy)
			b, bErr := surfaceDirectionAt(target, observer, 2461319.5, vector)
			if err != nil || bErr != nil {
				t.Fatalf("%v/%v", err, bErr)
			}
			maximum = math.Max(maximum, math.Atan2(vectorMagnitude(pxp(a.direction, b.direction)), vectorDot(a.direction, b.direction))*180/math.Pi*3600)
		}
		t.Logf("body=%d model direction maximum difference = %.9f arcsec", body, maximum)
	}
}

func TestSurfaceDirectionSyntheticInvalid(t *testing.T) {
	inst := ephem.NewModelInstance(ephem.Semianalytic)
	valid := BodyObserver{Body: SkyMars}
	for _, sample := range []struct {
		observer BodyObserver
		target   SkyBody
		jd       float64
		inst     *Instance
	}{
		{valid, SkyMars, 2451545, inst}, {valid, SkyMoon, 2451545, inst}, {valid, SkyBody(99), 2451545, inst},
		{BodyObserver{Body: SkyEarth}, SkySun, 2451545, inst}, {BodyObserver{Body: SkySun}, SkyEarth, 2451545, inst},
		{BodyObserver{Body: SkyMars, Latitude: 91}, SkySun, 2451545, inst}, {BodyObserver{Body: SkyMars, Height: math.NaN()}, SkySun, 2451545, inst},
		{valid, SkySun, math.Inf(1), inst}, {valid, SkySun, math.NaN(), inst},
	} {
		got, err := surfaceDirectionAt(sample.target, sample.observer, sample.jd, sample.inst)
		if !errors.Is(err, ErrInvalidObservationInput) || got != (surfaceDirection{}) {
			t.Fatalf("sample=%+v result=%+v err=%v", sample, got, err)
		}
	}
	for _, custom := range []*Instance{ephem.NewInstance(nil), ephem.NewInstance(surfaceTestEngine{})} {
		got, err := surfaceDirectionAt(SkySun, valid, 2451545, custom)
		if err == nil || got != (surfaceDirection{}) {
			t.Fatalf("custom result=%+v err=%v", got, err)
		}
	}
}

func TestSurfaceDirectionConsistencyEarthCenter(t *testing.T) {
	inst := ephem.NewModelInstance(ephem.Legacy)
	for _, jd := range []float64{2451545, 2461319.5} {
		wantAU := rotateEclipticToEquatorial(earthHeliocentricVectorJ2000(jd, inst), orbitJ2000Obliquity)
		got := surfaceLegacyTarget(SkyEarth, jd, inst)
		for i := range got {
			if math.Abs(got[i]-wantAU[i]*astronomicalUnitKM) > 1e-6 {
				t.Fatalf("jd=%g Earth=%v expected=%v AU", jd, got, wantAU)
			}
		}
	}
}

func surfaceTestICRF(mean Vector3) Vector3 {
	bias := FrameBiasMatrix()
	return Vector3{
		bias[0][0]*mean[0] + bias[1][0]*mean[1] + bias[2][0]*mean[2],
		bias[0][1]*mean[0] + bias[1][1]*mean[1] + bias[2][1]*mean[2],
		bias[0][2]*mean[0] + bias[1][2]*mean[1] + bias[2][2]*mean[2],
	}
}

func TestSurfaceStarDirectionSyntheticNoDistance(t *testing.T) {
	engine := surfaceTestEngine{positions: map[ephem.Body]Vector3{ephem.SSB: {}, ephem.Sun: {}, ephem.Mars: {2e8, -1e8, 3e7}}, velocities: map[ephem.Body]Vector3{ephem.Mars: {-1e6, 2e6, 1e5}}}
	observer := BodyObserver{Body: SkyMars, EastLongitude: 80, Height: 2}
	epoch := time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC)
	maximum := 0.0
	for _, frameID := range []CoordinateFrame{CoordinateFrameICRS, CoordinateFrameJ2000} {
		for _, jd := range []float64{2451545, 2461319.5} {
			star := StarCoordinate{RA: 73, Dec: -21, Epoch: epoch, Frame: frameID}
			frame, err := bodyObserverFrameAt(observer, jd)
			if err != nil {
				t.Fatal(err)
			}
			velocity := engine.velocities[ephem.Mars]
			for i := range velocity {
				velocity[i] += frame.east[i] * 3398.19 * frame.orientation.RotationRate * math.Pi / 180
			}
			q := raDecToVector(star.RA, star.Dec)
			if frameID == CoordinateFrameICRS {
				bias := FrameBiasMatrix()
				q = normalizeVector(Vector3{
					bias[0][0]*q[0] + bias[0][1]*q[1] + bias[0][2]*q[2],
					bias[1][0]*q[0] + bias[1][1]*q[1] + bias[1][2]*q[2],
					bias[2][0]*q[0] + bias[2][1]*q[1] + bias[2][2]*q[2],
				})
			}
			q = normalizeVector(surfaceTestICRF(q))
			want := surfaceIndependentAberration(q, velocity)
			got, err := surfaceStarDirectionAt(star, observer, jd, ephem.NewInstance(engine))
			if err != nil || got.distanceKM != 0 || got.tauDays != 0 {
				t.Fatalf("frame=%s got=%+v err=%v", frameID, got, err)
			}
			for i := range want {
				maximum = math.Max(maximum, math.Abs(got.direction[i]-want[i]))
				if math.Abs(got.direction[i]-want[i]) > 1e-15 {
					t.Fatalf("frame=%s direction=%v want=%v", frameID, got.direction, want)
				}
			}
		}
	}
	t.Logf("distance-free frame/aberration maximum component difference = %.12g", maximum)
}

func TestSurfaceStarDirectionSyntheticMotion(t *testing.T) {
	epoch := time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC)
	const years = 25.0
	jd := occultationTimeToTT(epoch) + years*365.25
	engine := surfaceTestEngine{positions: map[ephem.Body]Vector3{ephem.SSB: {}, ephem.Sun: {}, ephem.Mars: {}}, velocities: map[ephem.Body]Vector3{}}
	inst := ephem.NewInstance(engine)
	observer := BodyObserver{Body: SkyMars, Latitude: 90}
	star := StarCoordinate{RA: 0, Dec: 0, Epoch: epoch, Frame: CoordinateFrameJ2000, ProperMotionRACosDecMasPerYear: 3600, ProperMotionDecMasPerYear: -1800}
	want := normalizeVector(surfaceTestICRF(raDecToVector(0.025, -0.0125)))
	got, err := surfaceStarDirectionAt(star, observer, jd, inst)
	if err != nil || got.distanceKM != 0 || got.tauDays != 0 {
		t.Fatalf("2D got=%+v err=%v", got, err)
	}
	for i := range want {
		if math.Abs(got.direction[i]-want[i]) > 1e-15 {
			t.Fatalf("2D got=%v want=%v", got.direction, want)
		}
	}
	star.ParallaxMas, star.RadialVelocityKmPerSecond = 1000, 10
	const distanceAU = 206264806.247 / 1000
	position := surfaceTestICRF(Vector3{
		distanceAU + years*10*365.25*86400/149597870.7,
		distanceAU * years * 3600 / 3600000 * math.Pi / 180,
		distanceAU * years * -1800 / 3600000 * math.Pi / 180,
	})
	frame, err := bodyObserverFrameAt(observer, jd)
	if err != nil {
		t.Fatal(err)
	}
	for i := range position {
		position[i] = position[i]*astronomicalUnitKM - frame.positionKM[i]
	}
	want = normalizeVector(position)
	got, err = surfaceStarDirectionAt(star, observer, jd, inst)
	if err != nil || math.Abs(got.distanceKM-vectorMagnitude(position)) > 0.01 || got.tauDays != 0 {
		t.Fatalf("3D got=%+v distance=%g err=%v", got, vectorMagnitude(position), err)
	}
	for i := range want {
		if math.Abs(got.direction[i]-want[i]) > 1e-15 {
			t.Fatalf("3D got=%v want=%v", got.direction, want)
		}
	}
	star.DistanceLightYear = 4.2
	withBoth, err := surfaceStarDirectionAt(star, observer, jd, inst)
	if err != nil || got != withBoth {
		t.Fatalf("explicit parallax precedence: got=%+v both=%+v err=%v", got, withBoth, err)
	}
}

func TestSurfaceStarDirectionSyntheticNeptuneParallax(t *testing.T) {
	engine := surfaceTestEngine{positions: map[ephem.Body]Vector3{ephem.SSB: {}, ephem.Sun: {}, ephem.Neptune: {30 * astronomicalUnitKM, 0, 0}}, velocities: map[ephem.Body]Vector3{}}
	observer := BodyObserver{Body: SkyNeptune, Latitude: 90}
	star := StarCoordinate{RA: 90, Dec: 0, Epoch: time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC), Frame: CoordinateFrameICRS, ParallaxMas: 768.5}
	got, err := surfaceStarDirectionAt(star, observer, 2451545, ephem.NewInstance(engine))
	if err != nil {
		t.Fatal(err)
	}
	shift := math.Atan2(math.Hypot(got.direction[0], got.direction[2]), got.direction[1]) * 180 / math.Pi * 3600
	want := math.Atan(30/(206264806.247/768.5)) * 180 / math.Pi * 3600
	if math.Abs(shift-want) > 0.005 || got.tauDays != 0 || !(got.distanceKM > 0) {
		t.Fatalf("parallax=%g arcsec expected=%g tau=%g distance=%g", shift, want, got.tauDays, got.distanceKM)
	}
	t.Logf("Neptune 30 AU parallax = %.9f arcsec, center-only hand calculation %.9f arcsec, difference=%+.9f", shift, want, shift-want)
}

func TestSurfaceStarDirectionConsistencyInstances(t *testing.T) {
	previous := GetEphemerisModel()
	defer SetEphemerisModel(previous)
	for _, model := range []EphemerisModel{EphemerisLegacy, EphemerisSemianalytic} {
		SetEphemerisModel(model)
		inst := ephem.NewModelInstance(ephem.Model(model))
		for body := SkyMercury; body <= SkyNeptune; body++ {
			for _, distance := range []float64{0, 4.2465} {
				star := StarCoordinate{RA: 217.429, Dec: -62.68, Epoch: time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC), Frame: CoordinateFrameICRS, DistanceLightYear: distance, ProperMotionRACosDecMasPerYear: -3775, ProperMotionDecMasPerYear: 766, RadialVelocityKmPerSecond: -22}
				observer := BodyObserver{Body: body, EastLongitude: 25, Latitude: -10, Height: 1}
				got, err := surfaceStarDirectionAt(star, observer, 2461319.5, nil)
				want, wantErr := surfaceStarDirectionAt(star, observer, 2461319.5, inst)
				if err != nil || wantErr != nil || got != want || !isFiniteFloat(got.distanceKM) || got.tauDays != 0 || !isFiniteFloat(vectorMagnitude(got.direction)) {
					t.Fatalf("model=%d body=%d distance=%g got=%+v want=%+v err=%v/%v", model, body, distance, got, want, err, wantErr)
				}
			}
		}
	}
}

func TestSurfaceStarDirectionSyntheticInvalid(t *testing.T) {
	valid := StarCoordinate{RA: 10, Dec: 20, Epoch: time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC), Frame: CoordinateFrameICRS}
	observer := BodyObserver{Body: SkyMars}
	inst := ephem.NewModelInstance(ephem.Semianalytic)
	badStars := []StarCoordinate{{}, valid, valid, valid, valid, valid, valid, valid, valid}
	badStars[1].Frame = CoordinateFrameApparentOfDate
	badStars[2].Frame = CoordinateFrame("invalid")
	badStars[3].RA = math.NaN()
	badStars[4].Dec = 91
	badStars[5].ParallaxMas = -1
	badStars[6].ProperMotionDecMasPerYear = math.Inf(1)
	badStars[7].DistanceLightYear = math.Inf(1)
	badStars[8].RadialVelocityKmPerSecond = math.NaN()
	for _, star := range badStars {
		got, err := surfaceStarDirectionAt(star, observer, 2451545, inst)
		if !errors.Is(err, ErrInvalidObservationInput) || got != (surfaceDirection{}) {
			t.Fatalf("star=%+v got=%+v err=%v", star, got, err)
		}
	}
	for _, sample := range []struct {
		observer BodyObserver
		jd       float64
		inst     *Instance
	}{{observer, math.NaN(), inst}, {BodyObserver{Body: SkyEarth}, 2451545, inst}, {observer, 2451545, ephem.NewInstance(nil)}, {observer, 2451545, ephem.NewInstance(surfaceTestEngine{})}} {
		got, err := surfaceStarDirectionAt(valid, sample.observer, sample.jd, sample.inst)
		if err == nil || got != (surfaceDirection{}) {
			t.Fatalf("sample=%+v got=%+v err=%v", sample, got, err)
		}
	}
}
