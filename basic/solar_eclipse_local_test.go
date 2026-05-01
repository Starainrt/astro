package basic

import (
	"math"
	"testing"
	"time"
)

func TestLocalSolarEclipseAgainstNASABaseline(t *testing.T) {
	testCases := []struct {
		name                string
		seedTT              float64
		lon                 float64
		lat                 float64
		height              float64
		wantType            SolarEclipseType
		wantGreatestUTC     time.Time
		wantPartialStartUTC time.Time
		wantPartialEndUTC   time.Time
		wantMagnitude       float64
		wantObscuration     float64
		wantSunAltitude     float64
		wantSunAzimuth      float64
		wantCentralDuration time.Duration
	}{
		{
			// NASA GSFC local circumstances page:
			// https://eclipse.gsfc.nasa.gov/SEcirc/SEcircNA/ChicagoIL1+21.html
			name:                "2024-04-08 chicago partial",
			seedTT:              solarEclipseUTToTTJDE(time.Date(2024, 4, 8, 0, 0, 0, 0, time.UTC)),
			lon:                 -87.65,
			lat:                 41.85,
			height:              0,
			wantType:            SolarEclipsePartial,
			wantGreatestUTC:     time.Date(2024, 4, 8, 19, 7, 0, 0, time.UTC),
			wantPartialStartUTC: time.Date(2024, 4, 8, 17, 51, 0, 0, time.UTC),
			wantPartialEndUTC:   time.Date(2024, 4, 8, 20, 22, 0, 0, time.UTC),
			wantMagnitude:       0.942,
			wantObscuration:     0.938,
			wantSunAltitude:     52,
			wantSunAzimuth:      211,
		},
		{
			// NASA GSFC eclipse catalog entry:
			// https://eclipse.gsfc.nasa.gov/SEgoogle/SEgoogle2001/SE2024Apr08Tgoogle.html
			name:                "2024-04-08 greatest total",
			seedTT:              solarEclipseUTToTTJDE(time.Date(2024, 4, 8, 0, 0, 0, 0, time.UTC)),
			lon:                 -104.1,
			lat:                 25.3,
			height:              0,
			wantType:            SolarEclipseTotal,
			wantGreatestUTC:     time.Date(2024, 4, 8, 18, 17, 15, 0, time.UTC),
			wantPartialStartUTC: time.Date(2024, 4, 8, 16, 58, 0, 0, time.UTC),
			wantPartialEndUTC:   time.Date(2024, 4, 8, 19, 40, 0, 0, time.UTC),
			wantMagnitude:       1.0566,
			wantObscuration:     1.0,
			wantSunAltitude:     70,
			wantSunAzimuth:      150,
			wantCentralDuration: 4*time.Minute + 28*time.Second,
		},
		{
			// NASA GSFC eclipse catalog entry:
			// https://eclipse.gsfc.nasa.gov/SEgoogle/SEgoogle2001/SE2024Oct02Agoogle.html
			name:                "2024-10-02 greatest annular",
			seedTT:              solarEclipseUTToTTJDE(time.Date(2024, 10, 2, 0, 0, 0, 0, time.UTC)),
			lon:                 -114.5,
			lat:                 -22.0,
			height:              0,
			wantType:            SolarEclipseAnnular,
			wantGreatestUTC:     time.Date(2024, 10, 2, 18, 44, 59, 0, time.UTC),
			wantPartialStartUTC: time.Date(2024, 10, 2, 17, 3, 0, 0, time.UTC),
			wantPartialEndUTC:   time.Date(2024, 10, 2, 20, 33, 0, 0, time.UTC),
			wantMagnitude:       0.9326,
			wantObscuration:     0.871,
			wantSunAltitude:     69,
			wantSunAzimuth:      31,
			wantCentralDuration: 7*time.Minute + 25*time.Second,
		},
	}

	const (
		partialTimeTolerance  = 90 * time.Second
		greatestTimeTolerance = 45 * time.Second
		magnitudeTolerance    = 0.002
		obscurationTolerance  = 0.01
		altitudeTolerance     = 1.0
		azimuthTolerance      = 1.5
		durationTolerance     = 5 * time.Second
	)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := LocalSolarEclipse(tc.seedTT, tc.lon, tc.lat, tc.height)
			if result.Type != tc.wantType {
				t.Fatalf("type mismatch: got %s want %s", result.Type, tc.wantType)
			}

			assertLocalSolarEclipseJDEClose(
				t,
				"GreatestEclipse",
				result.GreatestEclipse,
				solarEclipseUTToTTJDE(tc.wantGreatestUTC),
				greatestTimeTolerance,
			)
			assertLocalSolarEclipseJDEClose(
				t,
				"PartialStart",
				result.PartialStart,
				solarEclipseUTToTTJDE(tc.wantPartialStartUTC),
				partialTimeTolerance,
			)
			assertLocalSolarEclipseJDEClose(
				t,
				"PartialEnd",
				result.PartialEnd,
				solarEclipseUTToTTJDE(tc.wantPartialEndUTC),
				partialTimeTolerance,
			)

			assertSolarEclipseFloatClose(t, "Magnitude", result.Magnitude, tc.wantMagnitude, magnitudeTolerance)
			assertSolarEclipseFloatClose(t, "Obscuration", result.Obscuration, tc.wantObscuration, obscurationTolerance)
			assertSolarEclipseFloatClose(t, "SunAltitude", result.SunAltitude, tc.wantSunAltitude, altitudeTolerance)
			assertSolarEclipseFloatClose(t, "SunAzimuth", result.SunAzimuth, tc.wantSunAzimuth, azimuthTolerance)

			if tc.wantCentralDuration > 0 {
				if result.CentralStart == 0 || result.CentralEnd == 0 {
					t.Fatalf("expected central contact times, got %+v", result)
				}
				duration := jdeDuration(result.CentralStart, result.CentralEnd)
				assertDurationClose(t, "CentralDuration", duration, tc.wantCentralDuration, durationTolerance)
			} else if result.CentralStart != 0 || result.CentralEnd != 0 || result.HasCentral {
				t.Fatalf("expected no central phase, got %+v", result)
			}
		})
	}
}

func TestLocalSolarEclipseDefaultUsesNASABulletinSplitK(t *testing.T) {
	seedTT := solarEclipseUTToTTJDE(time.Date(2024, 4, 8, 0, 0, 0, 0, time.UTC))
	defaultResult := LocalSolarEclipse(seedTT, -104.1, 25.3, 0)
	nasaResult := LocalSolarEclipseNASABulletinSplitK(seedTT, -104.1, 25.3, 0)
	iauResult := LocalSolarEclipseIAUSingleK(seedTT, -104.1, 25.3, 0)

	if defaultResult.Model != SolarEclipseModelNASABulletinSplitK {
		t.Fatalf("default model mismatch: got %s want %s", defaultResult.Model, SolarEclipseModelNASABulletinSplitK)
	}

	assertSolarEclipseJDEClose(t, "GreatestEclipse", defaultResult.GreatestEclipse, nasaResult.GreatestEclipse, 1e-12)
	assertSolarEclipseFloatClose(t, "Magnitude", defaultResult.Magnitude, nasaResult.Magnitude, 1e-12)
	assertSolarEclipseFloatClose(t, "Obscuration", defaultResult.Obscuration, nasaResult.Obscuration, 1e-12)

	defaultDuration := jdeDuration(defaultResult.CentralStart, defaultResult.CentralEnd)
	iauDuration := jdeDuration(iauResult.CentralStart, iauResult.CentralEnd)
	if !(iauDuration > defaultDuration) {
		t.Fatalf("expected IAU central duration > NASA duration: iau=%v nasa=%v", iauDuration, defaultDuration)
	}
	if !(iauResult.Magnitude > defaultResult.Magnitude) {
		t.Fatalf("expected IAU magnitude > NASA magnitude: iau=%.9f nasa=%.9f", iauResult.Magnitude, defaultResult.Magnitude)
	}
}

func TestLocalSolarEclipseNoEvent(t *testing.T) {
	result := LocalSolarEclipse(
		solarEclipseUTToTTJDE(time.Date(2024, 4, 8, 0, 0, 0, 0, time.UTC)),
		18.4241,
		-33.9249,
		0,
	)
	if result.Type != SolarEclipseNone {
		t.Fatalf("type mismatch: got %s want %s", result.Type, SolarEclipseNone)
	}
	if result.HasPartial || result.HasCentral || result.HasAnnular || result.HasTotal {
		t.Fatalf("unexpected eclipse flags: %+v", result)
	}
	if result.PartialStart != 0 || result.PartialEnd != 0 || result.CentralStart != 0 || result.CentralEnd != 0 {
		t.Fatalf("expected no contact times, got %+v", result)
	}
}

func TestLocalSolarEclipseVisibleAtGreatestRespectsHeight(t *testing.T) {
	seedTT := solarEclipseUTToTTJDE(time.Date(2024, 4, 8, 0, 0, 0, 0, time.UTC))

	seaLevel := LocalSolarEclipse(seedTT, -155.0, -35.0, 0)
	highAltitude := LocalSolarEclipse(seedTT, -155.0, -35.0, 10000)

	if seaLevel.Type != SolarEclipsePartial || highAltitude.Type != SolarEclipsePartial {
		t.Fatalf("unexpected eclipse types: sea=%s high=%s", seaLevel.Type, highAltitude.Type)
	}
	if seaLevel.VisibleAtGreatest {
		t.Fatalf("expected sea-level greatest eclipse to be below the geometric horizon: %+v", seaLevel)
	}

	visibleThreshold := -HeightDegreeByLat(10000, -35.0)
	if !(highAltitude.SunAltitude > visibleThreshold) {
		t.Fatalf("sanity check failed: SunAltitude=%.6f threshold=%.6f", highAltitude.SunAltitude, visibleThreshold)
	}
	if !highAltitude.VisibleAtGreatest {
		t.Fatalf("expected high-altitude greatest eclipse to be visible: %+v", highAltitude)
	}
}

func solarEclipseUTToTTJDE(date time.Time) float64 {
	return TD2UT(Date2JDE(date.UTC()), true)
}

func assertLocalSolarEclipseJDEClose(
	t *testing.T,
	name string,
	got float64,
	want float64,
	tolerance time.Duration,
) {
	t.Helper()
	diff := math.Abs(got-want) * 86400
	if diff > tolerance.Seconds() {
		t.Fatalf("%s mismatch: got %.12f want %.12f diff=%.3fs", name, got, want, diff)
	}
}

func jdeDuration(startJDE, endJDE float64) time.Duration {
	return time.Duration((endJDE-startJDE)*86400*float64(time.Second) + 0.5)
}

func assertDurationClose(t *testing.T, name string, got, want, tolerance time.Duration) {
	t.Helper()
	diff := got - want
	if diff < 0 {
		diff = -diff
	}
	if diff > tolerance {
		t.Fatalf("%s mismatch: got %v want %v diff=%v", name, got, want, diff)
	}
}
