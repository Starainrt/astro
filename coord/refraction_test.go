package coord

import (
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

func TestRefractionWrappers(t *testing.T) {
	assertClose(t, "AtmosphericRefractionFromTrueAltitude", AtmosphericRefractionFromTrueAltitude(10, 1010, 10), basic.RefractionFromTrueAltitude(10, 1010, 10), 1e-12)
	assertClose(t, "AtmosphericRefractionFromApparentAltitude", AtmosphericRefractionFromApparentAltitude(10, 1010, 10), basic.RefractionFromApparentAltitude(10, 1010, 10), 1e-12)
	assertClose(t, "ApparentAltitude", ApparentAltitude(10, 1010, 10), basic.ApparentAltitude(10, 1010, 10), 1e-12)
	assertClose(t, "TrueAltitude", TrueAltitude(10, 1010, 10), basic.TrueAltitude(10, 1010, 10), 1e-12)
}

func TestEquatorialToApparentHorizontal(t *testing.T) {
	date := time.Date(2026, 4, 27, 2, 30, 45, 0, time.UTC)
	ra := 101.28715533
	dec := -16.71611586
	observerLon := 115.0
	observerLat := 40.0
	pressureHPa := 1010.0
	temperatureC := 10.0

	trueHorizontal := EquatorialToHorizontal(date, ra, dec, observerLon, observerLat)
	apparentHorizontal := EquatorialToApparentHorizontal(date, ra, dec, observerLon, observerLat, pressureHPa, temperatureC)

	assertClose(t, "apparent altitude", apparentHorizontal.Altitude, basic.ApparentAltitude(trueHorizontal.Altitude, pressureHPa, temperatureC), 1e-12)
	assertClose(t, "apparent zenith", apparentHorizontal.Zenith, 90-apparentHorizontal.Altitude, 1e-12)
	assertClose(t, "apparent azimuth unchanged", apparentHorizontal.Azimuth, trueHorizontal.Azimuth, 1e-12)
	assertClose(t, "apparent hour angle unchanged", apparentHorizontal.HourAngle, trueHorizontal.HourAngle, 1e-12)
}
