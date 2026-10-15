package moon

import (
	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/ephem"
	"testing"
	"time"
)

type contextAPI interface {
	Altitude(date time.Time, lon, lat float64) float64
	ApogeesInMonth(year int, month time.Month) []ApsisInfo
	ApparentAltitude(date time.Time, lon, lat, pressureHPa, temperatureC float64) float64
	ApparentDec(date time.Time, lon, lat float64) float64
	ApparentLo(date time.Time) float64
	ApparentRa(date time.Time, lon, lat float64) float64
	ApparentRaDec(date time.Time, lon, lat float64) (float64, float64)
	ApparentZenith(date time.Time, lon, lat, pressureHPa, temperatureC float64) float64
	AscendingNode(date time.Time) float64
	AscendingNodeN(date time.Time, n int) float64
	Azimuth(date time.Time, lon, lat float64) float64
	BrightLimbPositionAngle(date time.Time) float64
	BrightLimbPositionAngleN(date time.Time, n int) float64
	ClosestConjunctionWithPlanet(date time.Time, planet ConjunctionPlanet) time.Time
	ClosestFirstQuarter(date time.Time) time.Time
	ClosestFullMoon(date time.Time) time.Time
	ClosestLastQuarter(date time.Time) time.Time
	ClosestMaximumNorthDeclination(date time.Time) MaximumDeclinationInfo
	ClosestMaximumSouthDeclination(date time.Time) MaximumDeclinationInfo
	ClosestNewMoon(date time.Time) time.Time
	ClosestShangXianYue(date time.Time) time.Time
	ClosestShuoYue(date time.Time) time.Time
	ClosestWangYue(date time.Time) time.Time
	ClosestXiaXianYue(date time.Time) time.Time
	CulminationTime(date time.Time, lon, lat float64) time.Time
	DescendingNode(date time.Time) float64
	DescendingNodeN(date time.Time, n int) float64
	Diameter(date time.Time) float64
	DiameterN(date time.Time, n int) float64
	DownTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error)
	EarthDistance(date time.Time) float64
	FindBestPlanetOccultations(start, end time.Time, planet OccultationPlanet,
		options OccultationSearchOptions) ([]PlanetOccultationInfo, error)
	FindBestStarOccultations(start, end time.Time, star StarCoordinate, options OccultationSearchOptions) ([]StarOccultationInfo, error)
	FindPlanetOccultationPaths(start, end time.Time, planet OccultationPlanet,
		options OccultationPathOptions) ([]PlanetOccultationPath, error)
	FindPlanetOccultations(start, end time.Time, planet OccultationPlanet, longitude, latitude, height float64,
		options OccultationSearchOptions) ([]PlanetOccultationInfo, error)
	FindStarOccultationPaths(start, end time.Time, star StarCoordinate, options OccultationPathOptions) ([]StarOccultationPath, error)
	FindStarOccultations(start, end time.Time, star StarCoordinate, longitude, latitude, height float64,
		options OccultationSearchOptions) ([]StarOccultationInfo, error)
	FirstQuarter(year float64) time.Time
	FullMoon(year float64) time.Time
	GeocentricApparentDec(date time.Time) float64
	GeocentricApparentRa(date time.Time) float64
	GeocentricApparentRaDec(date time.Time) (float64, float64)
	HourAngle(date time.Time, lon, lat float64) float64
	LastConjunctionWithPlanet(date time.Time, planet ConjunctionPlanet) time.Time
	LastFirstQuarter(date time.Time) time.Time
	LastFullMoon(date time.Time) time.Time
	LastLastQuarter(date time.Time) time.Time
	LastMaximumNorthDeclination(date time.Time) MaximumDeclinationInfo
	LastMaximumSouthDeclination(date time.Time) MaximumDeclinationInfo
	LastNewMoon(date time.Time) time.Time
	LastQuarter(year float64) time.Time
	LastShangXianYue(date time.Time) time.Time
	LastShuoYue(date time.Time) time.Time
	LastWangYue(date time.Time) time.Time
	LastXiaXianYue(date time.Time) time.Time
	MaximumNorthDeclinationsInMonth(year int, month time.Month) []MaximumDeclinationInfo
	MaximumSouthDeclinationsInMonth(year int, month time.Month) []MaximumDeclinationInfo
	NewMoon(year float64) time.Time
	NextConjunctionWithPlanet(date time.Time, planet ConjunctionPlanet) time.Time
	NextFirstQuarter(date time.Time) time.Time
	NextFullMoon(date time.Time) time.Time
	NextLastQuarter(date time.Time) time.Time
	NextMaximumNorthDeclination(date time.Time) MaximumDeclinationInfo
	NextMaximumSouthDeclination(date time.Time) MaximumDeclinationInfo
	NextNewMoon(date time.Time) time.Time
	NextShangXianYue(date time.Time) time.Time
	NextShuoYue(date time.Time) time.Time
	NextWangYue(date time.Time) time.Time
	NextXiaXianYue(date time.Time) time.Time
	ParallacticAngle(date time.Time, lon, lat float64) float64
	PerigeesInMonth(year int, month time.Month) []ApsisInfo
	Phase(date time.Time) float64
	PhaseDesc(date time.Time) string
	Physical(date time.Time) PhysicalInfo
	PhysicalN(date time.Time, n int) PhysicalInfo
	PlanetOccultationDiagram(
		info PlanetOccultationInfo,
		options PlanetOccultationDiagramOptions,
	) PlanetOccultationDiagramResult
	PlanetOccultationFootprintsAt(at time.Time, planet OccultationPlanet) (PlanetOccultationInstant, error)
	PlanetOccultationInfoInUT1(info basic.PlanetOccultationInfo) basic.PlanetOccultationInfo
	PlanetOccultationPathInUT1(path basic.PlanetOccultationPath) basic.PlanetOccultationPath
	RiseTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error)
	Semidiameter(date time.Time) float64
	SemidiameterN(date time.Time, n int) float64
	SetTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error)
	ShangXianYue(year float64) time.Time
	ShuoYue(year float64) time.Time
	StarCoordinateFromStarData(star StarData) (StarCoordinate, error)
	StarOccultationDiagram(
		info StarOccultationInfo,
		star StarCoordinate,
		options StarOccultationDiagramOptions,
	) StarOccultationDiagramResult
	StarOccultationFootprintAt(at time.Time, star StarCoordinate) (StarOccultationInstant, error)
	StarOccultationInfoInUT1(info basic.StarOccultationInfo) basic.StarOccultationInfo
	StarOccultationPathInUT1(path basic.StarOccultationPath) basic.StarOccultationPath
	SunMoonLoDiff(date time.Time) float64
	TopocentricBrightLimbPositionAngle(date time.Time, observerLon, observerLat, height float64) float64
	TopocentricBrightLimbPositionAngleN(date time.Time, observerLon, observerLat, height float64, n int) float64
	TopocentricPhysical(date time.Time, observerLon, observerLat, height float64) PhysicalInfo
	TopocentricPhysicalN(date time.Time, observerLon, observerLat, height float64, n int) PhysicalInfo
	TrueBo(date time.Time) float64
	TrueBoN(date time.Time, n int) float64
	TrueDec(date time.Time) float64
	TrueLo(date time.Time) float64
	TrueLoN(date time.Time, n int) float64
	TrueRa(date time.Time) float64
	TrueRaDec(date time.Time) (float64, float64)
	WangYue(year float64) time.Time
	XiaXianYue(year float64) time.Time
	Zenith(date time.Time, lon, lat float64) float64
}

var _ contextAPI = (*Moon)(nil)

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
