package eclipse

import (
	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/ephem"
	"reflect"
	"testing"
	"time"
)

type contextAPI interface {
	ClosestGeometricLocalLunarEclipse(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo
	ClosestGeometricLocalLunarEclipseChauvenet(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo
	ClosestGeometricLocalLunarEclipseDanjon(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo
	ClosestGeometricLocalSolarEclipse(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo
	ClosestGeometricLocalSolarEclipseIAUSingleK(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo
	ClosestGeometricLocalSolarEclipseNASABulletinSplitK(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo
	ClosestLocalAnnularSolarEclipse(date time.Time, lon, lat, height float64) (LocalSolarEclipseInfo, bool)
	ClosestLocalLunarEclipse(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo
	ClosestLocalLunarEclipseChauvenet(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo
	ClosestLocalLunarEclipseDanjon(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo
	ClosestLocalSolarEclipse(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo
	ClosestLocalSolarEclipseIAUSingleK(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo
	ClosestLocalSolarEclipseNASABulletinSplitK(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo
	ClosestLocalTotalLunarEclipse(date time.Time, lon, lat, height float64) (LocalLunarEclipseInfo, bool)
	ClosestLocalTotalSolarEclipse(date time.Time, lon, lat, height float64) (LocalSolarEclipseInfo, bool)
	ClosestLunarEclipse(date time.Time) LunarEclipseInfo
	ClosestLunarEclipseChauvenet(date time.Time) LunarEclipseInfo
	ClosestLunarEclipseDanjon(date time.Time) LunarEclipseInfo
	ClosestSolarEclipse(date time.Time) SolarEclipseInfo
	ClosestSolarEclipseIAUSingleK(date time.Time) SolarEclipseInfo
	ClosestSolarEclipseNASABulletinSplitK(date time.Time) SolarEclipseInfo
	ClosestSolarEclipseWithOptions(date time.Time, options SolarEclipseOptions) SolarEclipseInfo
	GeometricLocalLunarEclipseOnDate(date time.Time, lon, lat, height float64) (LocalLunarEclipseInfo, bool)
	GeometricLocalLunarEclipseOnDateChauvenet(date time.Time, lon, lat, height float64) (LocalLunarEclipseInfo, bool)
	GeometricLocalLunarEclipseOnDateDanjon(date time.Time, lon, lat, height float64) (LocalLunarEclipseInfo, bool)
	GeometricLocalSolarEclipseOnDate(date time.Time, lon, lat, height float64) (LocalSolarEclipseInfo, bool)
	GeometricLocalSolarEclipseOnDateIAUSingleK(date time.Time, lon, lat, height float64) (LocalSolarEclipseInfo, bool)
	GeometricLocalSolarEclipseOnDateNASABulletinSplitK(date time.Time, lon, lat, height float64) (LocalSolarEclipseInfo, bool)
	LastGeometricLocalLunarEclipse(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo
	LastGeometricLocalLunarEclipseChauvenet(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo
	LastGeometricLocalLunarEclipseDanjon(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo
	LastGeometricLocalSolarEclipse(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo
	LastGeometricLocalSolarEclipseIAUSingleK(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo
	LastGeometricLocalSolarEclipseNASABulletinSplitK(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo
	LastLocalAnnularSolarEclipse(date time.Time, lon, lat, height float64) (LocalSolarEclipseInfo, bool)
	LastLocalLunarEclipse(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo
	LastLocalLunarEclipseChauvenet(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo
	LastLocalLunarEclipseDanjon(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo
	LastLocalSolarEclipse(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo
	LastLocalSolarEclipseIAUSingleK(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo
	LastLocalSolarEclipseNASABulletinSplitK(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo
	LastLocalTotalLunarEclipse(date time.Time, lon, lat, height float64) (LocalLunarEclipseInfo, bool)
	LastLocalTotalSolarEclipse(date time.Time, lon, lat, height float64) (LocalSolarEclipseInfo, bool)
	LastLunarEclipse(date time.Time) LunarEclipseInfo
	LastLunarEclipseChauvenet(date time.Time) LunarEclipseInfo
	LastLunarEclipseDanjon(date time.Time) LunarEclipseInfo
	LastSolarEclipse(date time.Time) SolarEclipseInfo
	LastSolarEclipseIAUSingleK(date time.Time) SolarEclipseInfo
	LastSolarEclipseNASABulletinSplitK(date time.Time) SolarEclipseInfo
	LastSolarEclipseWithOptions(date time.Time, options SolarEclipseOptions) SolarEclipseInfo
	LocalLunarEclipseOnDate(date time.Time, lon, lat, height float64) (LocalLunarEclipseInfo, bool)
	LocalLunarEclipseOnDateChauvenet(date time.Time, lon, lat, height float64) (LocalLunarEclipseInfo, bool)
	LocalLunarEclipseOnDateDanjon(date time.Time, lon, lat, height float64) (LocalLunarEclipseInfo, bool)
	LocalSolarEclipseInfoInUT1(info LocalSolarEclipseInfo) LocalSolarEclipseInfo
	LocalSolarEclipseOnDate(date time.Time, lon, lat, height float64) (LocalSolarEclipseInfo, bool)
	LocalSolarEclipseOnDateIAUSingleK(date time.Time, lon, lat, height float64) (LocalSolarEclipseInfo, bool)
	LocalSolarEclipseOnDateNASABulletinSplitK(date time.Time, lon, lat, height float64) (LocalSolarEclipseInfo, bool)
	LocalSolarEclipseOnDateWithOptions(date time.Time, lon, lat, height float64, options SolarEclipseOptions) (LocalSolarEclipseInfo, bool)
	LunarEclipseGeocentricPanelAt(date time.Time) (LunarEclipseGeocentricPanel, bool)
	LunarEclipseGeocentricPanelChauvenet(date time.Time) (LunarEclipseGeocentricPanel, bool)
	LunarEclipseInfoInUT1(info LunarEclipseInfo) LunarEclipseInfo
	LunarEclipseOnDate(date time.Time) (LunarEclipseInfo, bool)
	LunarEclipseOnDateChauvenet(date time.Time) (LunarEclipseInfo, bool)
	LunarEclipseOnDateDanjon(date time.Time) (LunarEclipseInfo, bool)
	NewSolarEclipseShadowSolver(options SolarEclipseShadowSolverOptions) *SolarEclipseShadowSolver
	NextGeometricLocalLunarEclipse(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo
	NextGeometricLocalLunarEclipseChauvenet(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo
	NextGeometricLocalLunarEclipseDanjon(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo
	NextGeometricLocalSolarEclipse(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo
	NextGeometricLocalSolarEclipseIAUSingleK(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo
	NextGeometricLocalSolarEclipseNASABulletinSplitK(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo
	NextLocalAnnularSolarEclipse(date time.Time, lon, lat, height float64) (LocalSolarEclipseInfo, bool)
	NextLocalLunarEclipse(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo
	NextLocalLunarEclipseChauvenet(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo
	NextLocalLunarEclipseDanjon(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo
	NextLocalSolarEclipse(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo
	NextLocalSolarEclipseIAUSingleK(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo
	NextLocalSolarEclipseNASABulletinSplitK(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo
	NextLocalTotalLunarEclipse(date time.Time, lon, lat, height float64) (LocalLunarEclipseInfo, bool)
	NextLocalTotalSolarEclipse(date time.Time, lon, lat, height float64) (LocalSolarEclipseInfo, bool)
	NextLunarEclipse(date time.Time) LunarEclipseInfo
	NextLunarEclipseChauvenet(date time.Time) LunarEclipseInfo
	NextLunarEclipseDanjon(date time.Time) LunarEclipseInfo
	NextSolarEclipse(date time.Time) SolarEclipseInfo
	NextSolarEclipseIAUSingleK(date time.Time) SolarEclipseInfo
	NextSolarEclipseNASABulletinSplitK(date time.Time) SolarEclipseInfo
	NextSolarEclipseWithOptions(date time.Time, options SolarEclipseOptions) SolarEclipseInfo
	SearchLocalCentralSolarEclipse(
		date time.Time, lon, lat, height float64,
		options SolarEclipseLocalSearchOptions,
	) (LocalSolarEclipseInfo, SolarEclipseLocalSearchStatus)
	SolarEclipseBesselianElements(
		seedJDE float64, options SolarEclipseBesselianElementsOptions,
	) (SolarEclipseBesselianElementsResult, bool)
	SolarEclipseBesselianMuForPublishedTable(
		mu SolarEclipseBesselianPolynomial, deltaTSeconds float64,
	) SolarEclipseBesselianPolynomial
	SolarEclipseCandidates(
		start, end time.Time, options SolarEclipseCandidateOptions,
	) []SolarEclipseCandidate
	SolarEclipseCentralPath(date time.Time, options SolarEclipsePathOptions) (SolarEclipsePath, bool)
	SolarEclipseCentralPathIAUSingleK(date time.Time, options SolarEclipsePathOptions) (SolarEclipsePath, bool)
	SolarEclipseCentralPathNASABulletinSplitK(date time.Time, options SolarEclipsePathOptions) (SolarEclipsePath, bool)
	SolarEclipseGeocentricPanelAt(date time.Time) (SolarEclipseGeocentricPanel, bool)
	SolarEclipseGeocentricPanelIAUSingleK(date time.Time) (SolarEclipseGeocentricPanel, bool)
	SolarEclipseGeocentricPanelInUT1(panel SolarEclipseGeocentricPanel) SolarEclipseGeocentricPanel
	SolarEclipseGeocentricPanelWithOptions(date time.Time, options SolarEclipseOptions) (SolarEclipseGeocentricPanel, bool)
	SolarEclipseInfoInUT1(info SolarEclipseInfo) SolarEclipseInfo
	SolarEclipseOnDate(date time.Time) (SolarEclipseInfo, bool)
	SolarEclipseOnDateIAUSingleK(date time.Time) (SolarEclipseInfo, bool)
	SolarEclipseOnDateNASABulletinSplitK(date time.Time) (SolarEclipseInfo, bool)
	SolarEclipseOnDateWithOptions(date time.Time, options SolarEclipseOptions) (SolarEclipseInfo, bool)
	SolarEclipsePartialArea(date time.Time, options SolarEclipsePartialAreaOptions) (SolarEclipsePartialAreaInfo, bool)
	SolarEclipsePartialAreaIAUSingleK(date time.Time, options SolarEclipsePartialAreaOptions) (SolarEclipsePartialAreaInfo, bool)
	SolarEclipsePartialAreaNASABulletinSplitK(date time.Time, options SolarEclipsePartialAreaOptions) (SolarEclipsePartialAreaInfo, bool)
	SolarEclipsePartialFootprints(date time.Time, options SolarEclipsePartialFootprintOptions) (SolarEclipsePartialFootprintsInfo, bool)
	SolarEclipsePartialFootprintsIAUSingleK(date time.Time, options SolarEclipsePartialFootprintOptions) (SolarEclipsePartialFootprintsInfo, bool)
	SolarEclipsePartialFootprintsInUT1(info SolarEclipsePartialFootprintsInfo) SolarEclipsePartialFootprintsInfo
	SolarEclipsePartialFootprintsNASABulletinSplitK(date time.Time, options SolarEclipsePartialFootprintOptions) (SolarEclipsePartialFootprintsInfo, bool)
	SolarEclipsePathInUT1(path SolarEclipsePath) SolarEclipsePath
	SolarEclipseShadowAt(value time.Time) (SolarEclipseShadowInstant, bool)
	SolarEclipseShadowBetween(
		start, end time.Time, step time.Duration,
	) []SolarEclipseShadowInstant
	SolarEclipseShadowBetweenWithOptions(
		start, end time.Time, step time.Duration, options SolarEclipseShadowSolverOptions,
	) []SolarEclipseShadowInstant
	SolarEclipseStationStateAt(value time.Time, lon, lat, height float64) SolarEclipseStationState
	TimeLabelsInUT1(times []time.Time) []time.Time
}

var _ contextAPI = (*Eclipse)(nil)

func TestContextModuleIsolation(t *testing.T) {
	previous := basic.GetEphemerisModel()
	t.Cleanup(func() { basic.SetEphemerisModel(previous) })
	date := time.Date(2024, 4, 8, 0, 0, 0, 0, time.UTC)
	for _, model := range []ephem.Model{ephem.Legacy, ephem.Semianalytic} {
		basic.SetEphemerisModel(basic.EphemerisModel(model))
		want, wantOK := SolarEclipseOnDate(date)
		module := NewWithInstance(ephem.NewModelInstance(model))
		basic.SetEphemerisModel(basic.EphemerisModel(1 - int(model)))
		got, gotOK := module.SolarEclipseOnDate(date)
		if !wantOK || gotOK != wantOK || got != want {
			t.Fatalf("model %d: solar eclipse escaped instance", model)
		}
	}
}

// 站心日食、候选枚举与中心食搜索必须使用模块实例，切换全局模型不得改变结果。
func TestContextLocalSolarEclipseIsolation(t *testing.T) {
	previous := basic.GetEphemerisModel()
	t.Cleanup(func() { basic.SetEphemerisModel(previous) })
	date := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	for _, model := range []ephem.Model{ephem.Legacy, ephem.Semianalytic} {
		module := NewWithInstance(ephem.NewModelInstance(model))
		run := func() []interface{} {
			next := module.NextLocalSolarEclipse(date, -97, 32, 0)
			iau := module.NextLocalSolarEclipseIAUSingleK(date, -97, 32, 0)
			onDate, ok := module.LocalSolarEclipseOnDate(time.Date(2024, 4, 8, 0, 0, 0, 0, time.UTC), -97, 32, 0)
			candidates := module.SolarEclipseCandidates(date, date.AddDate(1, 0, 0), SolarEclipseCandidateOptions{})
			central, status := module.SearchLocalCentralSolarEclipse(date, -97, 32, 0, SolarEclipseLocalSearchOptions{MaxYears: 2})
			return []interface{}{next, iau, onDate, ok, candidates, central, status}
		}
		basic.SetEphemerisModel(basic.EphemerisModel(model))
		want := run()
		basic.SetEphemerisModel(basic.EphemerisModel(1 - int(model)))
		got := run()
		for i := range want {
			if !reflect.DeepEqual(got[i], want[i]) {
				t.Fatalf("model %d: result %d escaped instance:\n got %+v\nwant %+v", model, i, got[i], want[i])
			}
		}
	}
}
