package basic

import (
	"errors"
	"math"
	"testing"
	"time"

	. "github.com/starainrt/astro/tools"
)

func TestSunRiseSetDynamicResidual(t *testing.T) {
	const (
		longitude = 116.4074
		latitude  = 39.9042
		timeZone  = 8.0
		height    = 0.0
	)
	jd := JDCalc(2025, 6, 5)

	for _, event := range []struct {
		name string
		get  func(float64, float64, float64, float64, float64, float64) (float64, error)
	}{
		{name: "rise", get: GetSunRiseTime},
		{name: "set", get: GetSunSetTime},
	} {
		eventJD, err := event.get(jd, longitude, latitude, timeZone, 1, height)
		if err != nil {
			t.Fatalf("%s: %v", event.name, err)
		}
		naturalTimeZone := math.Round(longitude / 15)
		localJD := eventJD + naturalTimeZone/24 - timeZone/24
		residual := sunRiseSetResidual(localJD, longitude, latitude, naturalTimeZone, 1, height, -1)
		if math.Abs(residual) > 0.001 {
			t.Fatalf("%s dynamic horizon residual = %.9f degrees", event.name, residual)
		}

		fixedResidual := SunHeight(localJD, longitude, latitude, naturalTimeZone) - StandardAltitudeSun(1, height, latitude)
		if math.Abs(fixedResidual) < 0.01 {
			t.Fatalf("%s still matches the legacy fixed-altitude event: residual %.9f degrees", event.name, fixedResidual)
		}
	}
}

func TestSunApparentStateReusesDistanceWithoutChangingCoordinates(t *testing.T) {
	const jd = 2460827.5
	ra, dec, distanceAU := hSunApparentRaDecDistanceN(jd, -1)
	wantRA, wantDec := HSunApparentRaDec(jd)
	assertClose(t, "sun state RA", ra, wantRA, 1e-12)
	assertClose(t, "sun state Dec", dec, wantDec, 1e-12)
	assertClose(t, "sun state distance", distanceAU, EarthAwayN(jd, -1), 1e-15)
}

func TestSunRiseSetDynamicGrazingKeepsDateAndDirection(t *testing.T) {
	date := time.Date(2025, 6, 10, 0, 0, 0, 0, time.UTC)
	jd := Date2JD(date)
	dayStart := math.Floor(jd) + 0.5

	rise, err := GetSunRiseTime(jd, 0, 66, 0, 1, 0)
	if err != nil {
		t.Fatalf("sunrise: %v", err)
	}
	set, err := GetSunSetTime(jd, 0, 66, 0, 1, 0)
	if err != nil {
		t.Fatalf("sunset: %v", err)
	}

	assertRiseSetEvent(t, "sunrise", rise, dayStart, true, func(eventJD float64) float64 {
		return sunRiseSetResidual(eventJD, 0, 66, 0, 1, 0, -1)
	})
	assertRiseSetEvent(t, "sunset", set, dayStart, false, func(eventJD float64) float64 {
		return sunRiseSetResidual(eventJD, 0, 66, 0, 1, 0, -1)
	})
}

func TestMoonSetDynamicGrazingKeepsDirection(t *testing.T) {
	date := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)
	jd := Date2JD(date)
	dayStart := math.Floor(jd) + 0.5
	set, err := GetMoonSetTime(jd, 0, 80, 0, 1, 0)
	if err != nil {
		t.Fatalf("moonset: %v", err)
	}
	assertRiseSetEvent(t, "moonset", set, dayStart, false, func(eventJD float64) float64 {
		return moonRiseSetResidual(eventJD, 0, 80, 0, 1, 0, -1)
	})
}

func TestMoonRiseSetDirectionalFallbackPreservesMissingEventError(t *testing.T) {
	date := time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)
	jd := Date2JD(date)
	_, err := GetMoonRiseTime(jd, -42.6043, 71.7069, -3, 1, 0)
	if !errors.Is(err, ErrNeverRise) {
		t.Fatalf("moonrise error = %v, want %v", err, ErrNeverRise)
	}
}

func TestMoonRiseSetDynamicUsesObserverHeightForParallax(t *testing.T) {
	date := time.Date(2025, 6, 5, 0, 0, 0, 0, time.UTC)
	jd := Date2JD(date)
	const (
		longitude = 116.4074
		latitude  = 39.9042
		height    = 10000.0
	)
	rise, err := GetMoonRiseTime(jd, longitude, latitude, 0, 1, height)
	if err != nil {
		t.Fatalf("moonrise: %v", err)
	}
	residual := moonRiseSetResidualAtObserverHeight(rise, longitude, latitude, 0, 1, height)
	if math.Abs(residual) > 1e-5 {
		t.Fatalf("moonrise observer-height residual = %.12f degrees", residual)
	}
}

func assertRiseSetEvent(t *testing.T, name string, eventJD, dayStart float64, isRise bool, residual func(float64) float64) {
	t.Helper()
	if eventJD < dayStart || eventJD >= dayStart+1 {
		t.Fatalf("%s %.12f is outside civil day [%.12f, %.12f)", name, eventJD, dayStart, dayStart+1)
	}
	const step = 1.0 / 1440
	slope := (residual(eventJD+step) - residual(eventJD-step)) / (2 * step)
	if isRise && slope <= 0 {
		t.Fatalf("%s slope = %.9f degrees/day, want positive", name, slope)
	}
	if !isRise && slope >= 0 {
		t.Fatalf("%s slope = %.9f degrees/day, want negative", name, slope)
	}
	if value := residual(eventJD); math.Abs(value) > 1e-4 {
		t.Fatalf("%s residual = %.12f degrees", name, value)
	}
}

func moonRiseSetResidualAtObserverHeight(jd, longitude, latitude, timeZone, zenithShift, height float64) float64 {
	calculationJD := UTC2TT(jd - timeZone/24)
	ra, dec := HMoonTrueRaDecN(calculationJD, -1)
	distanceKM := HMoonAwayN(calculationJD, -1)
	topocentricRA, topocentricDec := TopocentricRaDec(ra, dec, latitude, longitude,
		jd-timeZone/24, distanceKM/angularDiameterAstronomicalUnitKM, height)
	siderealTime := Limit360(ApparentSiderealTime(UTC2UT1(jd-timeZone/24))*15 + longitude)
	hourAngle := Limit360(siderealTime - topocentricRA)
	altitude := ArcSin(Sin(latitude)*Sin(topocentricDec) + Cos(topocentricDec)*Cos(latitude)*Cos(hourAngle))
	residual := altitude + HeightDegreeByLat(height, latitude)
	if zenithShift != 0 {
		residual += RefractionFromTrueAltitude(altitude, refractionStandardPressureHPa, refractionStandardTemperatureC)
		residual += angularSemidiameterArcsec(moonEquatorialRadiusKM, distanceKM) / 3600
	}
	return residual
}
