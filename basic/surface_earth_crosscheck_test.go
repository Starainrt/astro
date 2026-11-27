package basic

import (
	"github.com/starainrt/astro/internal/ephem"
	"math"
	"testing"
)

func TestSurfaceEarthConsistencyCrosscheck(t *testing.T) {
	const limit = 3.0
	inst := ephem.NewModelInstance(ephem.Semianalytic)
	maximumSun, maximumStar := 0.0, 0.0
	for _, day := range []float64{2451544.5, 2459000.5, 2461041.5, 2461319.5} {
		for _, site := range [][2]float64{{0, 0}, {116.4, 39.9}, {-70, -35}} {
			longitude, latitude := site[0], site[1]
			for _, stellar := range []bool{false, true} {
				var star *Vector3
				if stellar {
					v := surfaceEarthStarForTest(UTC2TT(day+0.5), 75, 20)
					star = &v
				}
				sample := surfaceEarthSamplerForTest(inst, longitude, latitude, star)
				config := surfaceSearchConfig{rotationDays: 360 / 360.98564736629, solarDays: 1, latitude: latitude, options: SurfaceOptions{Limb: AltitudeLimbCenter}, steps: 48, envelope: true}
				got, state, err := surfaceSearch(sample, UTC2TT(day), UTC2TT(day+1), config)
				if err != nil || state != AltitudeCrosses {
					t.Fatalf("day=%g site=%v star=%v state=%d err=%v", day, site, stellar, state, err)
				}
				options := AltitudeCrossingOptions{Limb: AltitudeLimbCenter, NoRefraction: true}
				var want AltitudeCrossing
				if stellar {
					want, err = StarAltitudeCrossing(day, 75, 20, longitude, latitude, 0, 0, options)
				} else {
					want, err = SkyAltitudeCrossingWithInstance(SkySun, day, longitude, latitude, 0, 0, 0, options, inst)
				}
				if err != nil {
					t.Fatal(err)
				}
				if want.State != AltitudeCrosses || !isFiniteFloat(want.Rise) || !isFiniteFloat(want.Set) {
					t.Fatalf("day=%g site=%v stellar=%v reference=%+v", day, site, stellar, want)
				}
				counts := [2]int{}
				for _, event := range got {
					if event.Kind > SurfaceSet {
						continue
					}
					expected := want.Rise
					if event.Kind == SurfaceSet {
						expected = want.Set
					}
					delta := math.Abs(event.JDE-UTC2TT(expected)) * 86400
					if stellar {
						maximumStar = math.Max(maximumStar, delta)
					} else {
						maximumSun = math.Max(maximumSun, delta)
					}
					if !isFiniteFloat(delta) || delta > limit {
						t.Fatalf("day=%g site=%v stellar=%v kind=%d delta=%g seconds", day, site, stellar, event.Kind, delta)
					}
					counts[event.Kind]++
				}
				if counts != ([2]int{1, 1}) {
					t.Fatalf("day=%g site=%v stellar=%v crossings=%v", day, site, stellar, counts)
				}
			}
		}
	}
	t.Logf("24 solar and 24 stellar rise/set comparisons: maximum solar=%.9f s stellar=%.9f s, limit=%.0f s", maximumSun, maximumStar, limit)
}

func surfaceEarthFrameForTest(jd, longitude, latitude float64) bodyObserverFrame {
	matrix := BiasPrecessionNutation2006(jd)
	transpose := func(v Vector3) Vector3 {
		return Vector3{matrix[0][0]*v[0] + matrix[1][0]*v[1] + matrix[2][0]*v[2], matrix[0][1]*v[0] + matrix[1][1]*v[1] + matrix[2][1]*v[2], matrix[0][2]*v[0] + matrix[1][2]*v[1] + matrix[2][2]*v[2]}
	}
	pole := transpose(Vector3{0, 0, 1})
	ra := math.Atan2(pole[1], pole[0]) * 180 / math.Pi
	dec := math.Atan2(pole[2], math.Hypot(pole[0], pole[1])) * 180 / math.Pi
	orientation := BodyOrientation{PoleRA: ra, PoleDec: dec, RotationRate: 360.98564736629}
	gast := ApparentSiderealTime(UTC2UT1(TT2UTC(jd))) * 15 * math.Pi / 180
	prime := transpose(Vector3{math.Cos(gast), math.Sin(gast), 0})
	q := normalizeVector(Vector3{-pole[1], pole[0], 0})
	orientation.PrimeMeridianAngle = math.Atan2(vectorDot(prime, pxp(pole, q)), vectorDot(prime, q)) * 180 / math.Pi
	model := planetPhysicalModel{equatorialRadius: 6378.137, polarRadius: 6356.752314245}
	return bodyObserverFrameFromModel(BodyObserver{Body: SkyEarth, EastLongitude: longitude, Latitude: latitude}, model, orientation)
}
func surfaceEarthSamplerForTest(inst *Instance, longitude, latitude float64, star *Vector3) surfaceSampler {
	return func(jd float64) (SurfacePosition, error) {
		frame := surfaceEarthFrameForTest(jd, longitude, latitude)
		var direction Vector3
		distance := 0.0
		if star != nil {
			direction = *star
		} else {
			d, dist, _, err := ephem.ApparentFrom(inst.Engine(), ephem.Sun, ephem.Earth, [3]float64(frame.positionKM), [3]float64(surfaceSpinVelocity(frame)), ephem.Sun, jd)
			if err != nil {
				return SurfacePosition{}, err
			}
			direction, distance = Vector3(d), dist
		}
		return surfacePosition(surfaceDirection{direction: direction, distanceKM: distance, frame: frame}, BodyObserver{EastLongitude: longitude}, star == nil), nil
	}
}
func surfaceEarthStarForTest(jd, ra, dec float64) Vector3 {
	m := BiasPrecessionNutation2006(jd)
	v := raDecToVector(ra, dec)
	return Vector3{m[0][0]*v[0] + m[1][0]*v[1] + m[2][0]*v[2], m[0][1]*v[0] + m[1][1]*v[1] + m[2][1]*v[2], m[0][2]*v[0] + m[1][2]*v[1] + m[2][2]*v[2]}
}
