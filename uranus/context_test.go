package uranus

import (
	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/ephem"
	"testing"
	"time"
)

type contextAPI interface {
	Altitude(date time.Time, lon, lat float64) float64
	AltitudeN(date time.Time, lon, lat float64, n int) float64
	ApparentBo(date time.Time) float64
	ApparentBoN(date time.Time, n int) float64
	ApparentDec(date time.Time) float64
	ApparentDecN(date time.Time, n int) float64
	ApparentLo(date time.Time) float64
	ApparentLoN(date time.Time, n int) float64
	ApparentMagnitude(date time.Time) float64
	ApparentMagnitudeN(date time.Time, n int) float64
	ApparentRa(date time.Time) float64
	ApparentRaDec(date time.Time) (float64, float64)
	ApparentRaDecN(date time.Time, n int) (float64, float64)
	ApparentRaN(date time.Time, n int) float64
	AscendingNode(date time.Time) float64
	AscendingNodeN(date time.Time, n int) float64
	Azimuth(date time.Time, lon, lat float64) float64
	AzimuthN(date time.Time, lon, lat float64, n int) float64
	BrightLimbPositionAngle(date time.Time) float64
	BrightLimbPositionAngleN(date time.Time, n int) float64
	CulminationTime(date time.Time, lon float64) time.Time
	CulminationTimeN(date time.Time, lon float64, n int) time.Time
	DescendingNode(date time.Time) float64
	DescendingNodeN(date time.Time, n int) float64
	Diameter(date time.Time) float64
	DiameterN(date time.Time, n int) float64
	DownTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error)
	DownTimeN(date time.Time, lon, lat, height float64, aero bool, n int) (time.Time, error)
	EarthDistance(date time.Time) float64
	EarthDistanceN(date time.Time, n int) float64
	HourAngle(date time.Time, lon float64) float64
	HourAngleN(date time.Time, lon float64, n int) float64
	IlluminatedFraction(date time.Time) float64
	IlluminatedFractionN(date time.Time, n int) float64
	LastConjunction(date time.Time) time.Time
	LastEasternQuadrature(date time.Time) time.Time
	LastOpposition(date time.Time) time.Time
	LastProgradeToRetrograde(date time.Time) time.Time
	LastRetrogradeToPrograde(date time.Time) time.Time
	LastWesternQuadrature(date time.Time) time.Time
	NextConjunction(date time.Time) time.Time
	NextEasternQuadrature(date time.Time) time.Time
	NextOpposition(date time.Time) time.Time
	NextProgradeToRetrograde(date time.Time) time.Time
	NextRetrogradeToPrograde(date time.Time) time.Time
	NextWesternQuadrature(date time.Time) time.Time
	ParallacticAngle(date time.Time, lon, lat float64) float64
	ParallacticAngleN(date time.Time, lon, lat float64, n int) float64
	Phase(date time.Time) float64
	PhaseAngle(date time.Time) float64
	PhaseAngleN(date time.Time, n int) float64
	PhaseN(date time.Time, n int) float64
	Physical(date time.Time) PhysicalInfo
	PhysicalN(date time.Time, n int) PhysicalInfo
	PhysicalSystemIII(date time.Time) PhysicalInfo
	PhysicalSystemIIIN(date time.Time, n int) PhysicalInfo
	RiseTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error)
	RiseTimeN(date time.Time, lon, lat, height float64, aero bool, n int) (time.Time, error)
	Semidiameter(date time.Time) float64
	SemidiameterN(date time.Time, n int) float64
	SetTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error)
	SetTimeN(date time.Time, lon, lat, height float64, aero bool, n int) (time.Time, error)
	SunDistance(date time.Time) float64
	SunDistanceN(date time.Time, n int) float64
	Zenith(date time.Time, lon, lat float64) float64
	ZenithN(date time.Time, lon, lat float64, n int) float64
}

var _ contextAPI = (*Uranus)(nil)

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
