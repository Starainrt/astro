package sun

import (
	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/ephem"
	"testing"
	"time"
)

type contextAPI interface {
	Altitude(date time.Time, lon, lat float64) float64
	AltitudeN(date time.Time, lon, lat float64, n int) float64
	ApparentAltitude(date time.Time, lon, lat, pressureHPa, temperatureC float64) float64
	ApparentAltitudeN(date time.Time, lon, lat, pressureHPa, temperatureC float64, n int) float64
	ApparentDec(date time.Time) float64
	ApparentLo(date time.Time) float64
	ApparentRa(date time.Time) float64
	ApparentRaDec(date time.Time) (float64, float64)
	ApparentSolarTime(date time.Time, lon float64) time.Time
	ApparentSolarTimeN(date time.Time, lon float64, n int) time.Time
	ApparentZenith(date time.Time, lon, lat, pressureHPa, temperatureC float64) float64
	ApparentZenithN(date time.Time, lon, lat, pressureHPa, temperatureC float64, n int) float64
	AxialtiltNutation(date time.Time) float64
	AxialtiltNutation1980(date time.Time) float64
	Azimuth(date time.Time, lon, lat float64) float64
	AzimuthN(date time.Time, lon, lat float64, n int) float64
	CulminationTime(date time.Time, lon float64) time.Time
	CulminationTimeN(date time.Time, lon float64, n int) time.Time
	Diameter(date time.Time) float64
	DiameterN(date time.Time, n int) float64
	DownTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error)
	DownTimeN(date time.Time, lon, lat, height float64, aero bool, n int) (time.Time, error)
	EarthDistance(date time.Time) float64
	EclipticNutation(date time.Time) float64
	EclipticNutation1980(date time.Time) float64
	EclipticObliquity(date time.Time, nutation bool) float64
	EquationTime(date time.Time) float64
	EveningTwilight(date time.Time, lon, lat, angle float64) (time.Time, error)
	EveningTwilightN(date time.Time, lon, lat, angle float64, n int) (time.Time, error)
	GeometricLo(date time.Time) float64
	HourAngle(date time.Time, lon, lat float64) float64
	HourAngleN(date time.Time, lon, lat float64, n int) float64
	MidFunc(date time.Time) float64
	MorningTwilight(date time.Time, lon, lat, angle float64) (time.Time, error)
	MorningTwilightN(date time.Time, lon, lat, angle float64, n int) (time.Time, error)
	ParallacticAngle(date time.Time, lon, lat float64) float64
	ParallacticAngleN(date time.Time, lon, lat float64, n int) float64
	Physical(date time.Time) PhysicalInfo
	PhysicalN(date time.Time, n int) PhysicalInfo
	RiseTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error)
	RiseTimeN(date time.Time, lon, lat, height float64, aero bool, n int) (time.Time, error)
	Semidiameter(date time.Time) float64
	SemidiameterN(date time.Time, n int) float64
	SetTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error)
	SetTimeN(date time.Time, lon, lat, height float64, aero bool, n int) (time.Time, error)
	TrueBo(date time.Time) float64
	TrueBoN(date time.Time, n int) float64
	TrueLo(date time.Time) float64
	TrueLoN(date time.Time, n int) float64
	Zenith(date time.Time, lon, lat float64) float64
	ZenithN(date time.Time, lon, lat float64, n int) float64
}

var _ contextAPI = (*Sun)(nil)

func TestContextModuleIsolation(t *testing.T) {
	previous := basic.GetEphemerisModel()
	t.Cleanup(func() { basic.SetEphemerisModel(previous) })
	date := time.Date(2024, 10, 6, 12, 0, 0, 0, time.UTC)
	for _, model := range []ephem.Model{ephem.Legacy, ephem.Semianalytic} {
		basic.SetEphemerisModel(basic.EphemerisModel(model))
		want := ApparentLo(date)
		module := NewWithInstance(ephem.NewModelInstance(model))
		basic.SetEphemerisModel(basic.EphemerisModel(1 - int(model)))
		if got := module.ApparentLo(date); got != want {
			t.Fatalf("model %d: longitude %.12f, want %.12f", model, got, want)
		}
	}
}
