package basic

import (
	"testing"
	"time"

	. "github.com/starainrt/astro/tools"
)

func TestMoonTopocentricPhysicalMatchesCorrectionMethod(t *testing.T) {
	jd := TD2UT(Date2JDE(testTime(2026, 4, 28, 9, 30, 45)), true)
	observerLon := 121.4737
	observerLat := 31.2304

	got := MoonTopocentricPhysical(jd, observerLon, observerLat, 0)
	want := moonTopocentricPhysicalByCorrection(jd, observerLon, observerLat)

	assertPlanetPhaseClose(t, "MoonTopocentricPhysical.LibrationLongitude", got.LibrationLongitude, want.LibrationLongitude, 0.1)
	assertPlanetPhaseClose(t, "MoonTopocentricPhysical.LibrationLatitude", got.LibrationLatitude, want.LibrationLatitude, 0.1)
	assertPlanetPhaseClose(t, "MoonTopocentricPhysical.PositionAngle", got.PositionAngle, want.PositionAngle, 0.1)
}

func TestMoonTopocentricPhysicalSampleSweepFiniteAndInRange(t *testing.T) {
	samples := []struct {
		name        string
		jd          float64
		observerLon float64
		observerLat float64
		height      float64
	}{
		{"shanghai", TD2UT(Date2JDE(testTime(2026, 4, 28, 9, 30, 45)), true), 121.4737, 31.2304, 4},
		{"chicago", TD2UT(Date2JDE(testTime(2024, 3, 25, 7, 0, 0)), true), -87.65, 41.85, 180},
	}

	for _, sample := range samples {
		info := MoonTopocentricPhysical(sample.jd, sample.observerLon, sample.observerLat, sample.height)
		prefix := sample.name + "."

		assertFiniteRange(t, prefix+"OpticalLongitude", info.OpticalLongitude, -180, 180, false)
		assertFiniteRange(t, prefix+"OpticalLatitude", info.OpticalLatitude, -90, 90, false)
		assertFiniteRange(t, prefix+"PhysicalLongitude", info.PhysicalLongitude, -180, 180, false)
		assertFiniteRange(t, prefix+"PhysicalLatitude", info.PhysicalLatitude, -90, 90, false)
		assertFiniteRange(t, prefix+"LibrationLongitude", info.LibrationLongitude, -180, 180, false)
		assertFiniteRange(t, prefix+"LibrationLatitude", info.LibrationLatitude, -90, 90, false)
		assertFiniteRange(t, prefix+"PositionAngle", info.PositionAngle, -90, 90, false)
	}
}

func moonTopocentricPhysicalByCorrection(jd, observerLon, observerLat float64) MoonPhysicalInfo {
	geocentric := MoonPhysical(jd)
	moonRA := HMoonTrueRa(jd)
	moonDec := HMoonTrueDec(jd)
	hourAngle := StarHourAngle(TD2UT(jd, false), moonRA, observerLon, 0)
	horizontalParallax := ArcSin(6378.1366 / HMoonAway(jd))

	Q := ArcTan2(
		Cos(moonDec)*Sin(hourAngle),
		Cos(moonDec)*Sin(observerLat)-Sin(moonDec)*Cos(observerLat)*Cos(hourAngle),
	)
	z := ArcCos(Sin(moonDec)*Sin(observerLat) + Cos(moonDec)*Cos(observerLat)*Cos(hourAngle))
	piPrime := horizontalParallax * (Sin(z) + 0.0084*Sin(2*z))

	deltaL := -piPrime * Sin(Q-geocentric.PositionAngle) / Cos(geocentric.LibrationLatitude)
	deltaB := piPrime * Cos(Q-geocentric.PositionAngle)
	deltaP := deltaL*Sin(geocentric.LibrationLatitude+deltaB) - piPrime*Sin(Q)*Tan(moonDec)

	return MoonPhysicalInfo{
		OpticalLongitude:   geocentric.OpticalLongitude,
		OpticalLatitude:    geocentric.OpticalLatitude,
		PhysicalLongitude:  geocentric.PhysicalLongitude,
		PhysicalLatitude:   geocentric.PhysicalLatitude,
		LibrationLongitude: wrapSignedAngle180(geocentric.LibrationLongitude + deltaL),
		LibrationLatitude:  geocentric.LibrationLatitude + deltaB,
		PositionAngle:      geocentric.PositionAngle + deltaP,
	}
}

func testTime(year int, month time.Month, day, hour, minute, second int) time.Time {
	return time.Date(year, month, day, hour, minute, second, 0, time.UTC)
}
