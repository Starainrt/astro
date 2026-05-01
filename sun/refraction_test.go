package sun

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

func TestApparentAltitudeWrappers(t *testing.T) {
	date := time.Date(2026, 4, 28, 9, 30, 45, 0, time.UTC)
	pressureHPa := 1010.0
	temperatureC := 10.0

	trueAltitude := Altitude(date, 115, 40)
	assertSunRefractionClose(t, "ApparentAltitude", ApparentAltitude(date, 115, 40, pressureHPa, temperatureC), basic.ApparentAltitude(trueAltitude, pressureHPa, temperatureC), 1e-12)
	assertSunRefractionClose(t, "ApparentZenith", ApparentZenith(date, 115, 40, pressureHPa, temperatureC), 90-ApparentAltitude(date, 115, 40, pressureHPa, temperatureC), 1e-12)

	trueAltitudeN := AltitudeN(date, 115, 40, -1)
	assertSunRefractionClose(t, "ApparentAltitudeN", ApparentAltitudeN(date, 115, 40, pressureHPa, temperatureC, -1), basic.ApparentAltitude(trueAltitudeN, pressureHPa, temperatureC), 1e-12)
	assertSunRefractionClose(t, "ApparentZenithN", ApparentZenithN(date, 115, 40, pressureHPa, temperatureC, -1), 90-ApparentAltitudeN(date, 115, 40, pressureHPa, temperatureC, -1), 1e-12)
}

func assertSunRefractionClose(t *testing.T, name string, got, want, tolerance float64) {
	t.Helper()
	if math.Abs(got-want) > tolerance {
		t.Fatalf("%s mismatch: got %.18f want %.18f", name, got, want)
	}
}
