package eclipse

import (
	"time"
)

// ClosestGeometricLocalLunarEclipse 最近一次几何月食，默认 Danjon 模型 / closest geometric local lunar eclipse, Danjon by default.
func ClosestGeometricLocalLunarEclipse(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo {
	return defaultModule.ClosestGeometricLocalLunarEclipse(date, lon, lat, height)
}

// ClosestGeometricLocalLunarEclipseChauvenet 最近一次几何月食（Chauvenet） / closest geometric local lunar eclipse with Chauvenet model.
func ClosestGeometricLocalLunarEclipseChauvenet(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo {
	return defaultModule.ClosestGeometricLocalLunarEclipseChauvenet(date, lon, lat, height)
}

// ClosestGeometricLocalLunarEclipseDanjon 最近一次几何月食（Danjon） / closest geometric local lunar eclipse with Danjon model.
func ClosestGeometricLocalLunarEclipseDanjon(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo {
	return defaultModule.ClosestGeometricLocalLunarEclipseDanjon(date, lon, lat, height)
}

// ClosestGeometricLocalSolarEclipse 最近一次站心几何日食，默认 NASA bulletin Split-K 模型 / closest geometric local solar eclipse, NASA bulletin Split-K by default.
func ClosestGeometricLocalSolarEclipse(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo {
	return defaultModule.ClosestGeometricLocalSolarEclipse(date, lon, lat, height)
}

// ClosestGeometricLocalSolarEclipseIAUSingleK 最近一次站心几何日食（IAU Single-K） / closest geometric local solar eclipse with IAU Single-K.
func ClosestGeometricLocalSolarEclipseIAUSingleK(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo {
	return defaultModule.ClosestGeometricLocalSolarEclipseIAUSingleK(date, lon, lat, height)
}

// ClosestGeometricLocalSolarEclipseNASABulletinSplitK 最近一次站心几何日食（NASA bulletin Split-K） / closest geometric local solar eclipse with NASA bulletin Split-K.
func ClosestGeometricLocalSolarEclipseNASABulletinSplitK(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo {
	return defaultModule.ClosestGeometricLocalSolarEclipseNASABulletinSplitK(date, lon, lat, height)
}

// ClosestLocalAnnularSolarEclipse 最近一次站心日环食，默认 NASA bulletin Split-K 模型 / closest local annular solar eclipse, NASA bulletin Split-K by default.
func ClosestLocalAnnularSolarEclipse(date time.Time, lon, lat, height float64) (LocalSolarEclipseInfo, bool) {
	return defaultModule.ClosestLocalAnnularSolarEclipse(date, lon, lat, height)
}

// ClosestLocalLunarEclipse 最近一次可见月食，默认 Danjon 模型 / closest visible local lunar eclipse, Danjon by default.
func ClosestLocalLunarEclipse(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo {
	return defaultModule.ClosestLocalLunarEclipse(date, lon, lat, height)
}

// ClosestLocalLunarEclipseChauvenet 最近一次可见月食（Chauvenet） / closest visible local lunar eclipse with Chauvenet model.
func ClosestLocalLunarEclipseChauvenet(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo {
	return defaultModule.ClosestLocalLunarEclipseChauvenet(date, lon, lat, height)
}

// ClosestLocalLunarEclipseDanjon 最近一次可见月食（Danjon） / closest visible local lunar eclipse with Danjon model.
func ClosestLocalLunarEclipseDanjon(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo {
	return defaultModule.ClosestLocalLunarEclipseDanjon(date, lon, lat, height)
}

// ClosestLocalSolarEclipse 最近一次站心日食，默认 NASA bulletin Split-K 模型 / closest local solar eclipse, NASA bulletin Split-K by default.
func ClosestLocalSolarEclipse(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo {
	return defaultModule.ClosestLocalSolarEclipse(date, lon, lat, height)
}

// ClosestLocalSolarEclipseIAUSingleK 最近一次站心日食（IAU Single-K） / closest local solar eclipse with IAU Single-K.
func ClosestLocalSolarEclipseIAUSingleK(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo {
	return defaultModule.ClosestLocalSolarEclipseIAUSingleK(date, lon, lat, height)
}

// ClosestLocalSolarEclipseNASABulletinSplitK 最近一次站心日食（NASA bulletin Split-K） / closest local solar eclipse with NASA bulletin Split-K.
func ClosestLocalSolarEclipseNASABulletinSplitK(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo {
	return defaultModule.ClosestLocalSolarEclipseNASABulletinSplitK(date, lon, lat, height)
}

// ClosestLocalTotalLunarEclipse 最近一次可见月全食，默认 Danjon 模型 / closest visible local total lunar eclipse, Danjon by default.
func ClosestLocalTotalLunarEclipse(date time.Time, lon, lat, height float64) (LocalLunarEclipseInfo, bool) {
	return defaultModule.ClosestLocalTotalLunarEclipse(date, lon, lat, height)
}

// ClosestLocalTotalSolarEclipse 最近一次站心日全食，默认 NASA bulletin Split-K 模型 / closest local total solar eclipse, NASA bulletin Split-K by default.
func ClosestLocalTotalSolarEclipse(date time.Time, lon, lat, height float64) (LocalSolarEclipseInfo, bool) {
	return defaultModule.ClosestLocalTotalSolarEclipse(date, lon, lat, height)
}

// ClosestSolarEclipse 最近一次日食，默认 NASA bulletin Split-K 模型 / closest solar eclipse, NASA bulletin Split-K by default.
func ClosestSolarEclipse(date time.Time) SolarEclipseInfo {
	return defaultModule.ClosestSolarEclipse(date)
}

// ClosestSolarEclipseIAUSingleK 最近一次日食（IAU Single-K） / closest solar eclipse with IAU Single-K.
func ClosestSolarEclipseIAUSingleK(date time.Time) SolarEclipseInfo {
	return defaultModule.ClosestSolarEclipseIAUSingleK(date)
}

// ClosestSolarEclipseNASABulletinSplitK 最近一次日食（NASA bulletin Split-K） / closest solar eclipse with NASA bulletin Split-K.
func ClosestSolarEclipseNASABulletinSplitK(date time.Time) SolarEclipseInfo {
	return defaultModule.ClosestSolarEclipseNASABulletinSplitK(date)
}

// ClosestSolarEclipseWithOptions 最近一次日食（自定义半径口径） / closest solar eclipse with custom radius conventions.
func ClosestSolarEclipseWithOptions(date time.Time, options SolarEclipseOptions) SolarEclipseInfo {
	return defaultModule.ClosestSolarEclipseWithOptions(date, options)
}

// GeometricLocalLunarEclipseOnDate 当地几何月食查询，默认 Danjon 模型 / local geometric lunar eclipse query, Danjon by default.
func GeometricLocalLunarEclipseOnDate(date time.Time, lon, lat, height float64) (LocalLunarEclipseInfo, bool) {
	return defaultModule.GeometricLocalLunarEclipseOnDate(date, lon, lat, height)
}

// GeometricLocalLunarEclipseOnDateChauvenet 当地几何月食查询（Chauvenet） / local geometric lunar eclipse query with Chauvenet model.
func GeometricLocalLunarEclipseOnDateChauvenet(date time.Time, lon, lat, height float64) (LocalLunarEclipseInfo, bool) {
	return defaultModule.GeometricLocalLunarEclipseOnDateChauvenet(date, lon, lat, height)
}

// GeometricLocalLunarEclipseOnDateDanjon 当地几何月食查询（Danjon） / local geometric lunar eclipse query with Danjon model.
func GeometricLocalLunarEclipseOnDateDanjon(date time.Time, lon, lat, height float64) (LocalLunarEclipseInfo, bool) {
	return defaultModule.GeometricLocalLunarEclipseOnDateDanjon(date, lon, lat, height)
}

// GeometricLocalSolarEclipseOnDate 当地站心几何日食查询，默认 NASA bulletin Split-K 模型 / local geometric solar eclipse query, NASA bulletin Split-K by default.
func GeometricLocalSolarEclipseOnDate(date time.Time, lon, lat, height float64) (LocalSolarEclipseInfo, bool) {
	return defaultModule.GeometricLocalSolarEclipseOnDate(date, lon, lat, height)
}

// GeometricLocalSolarEclipseOnDateIAUSingleK 当地站心几何日食查询（IAU Single-K） / local geometric solar eclipse query with IAU Single-K.
func GeometricLocalSolarEclipseOnDateIAUSingleK(date time.Time, lon, lat, height float64) (LocalSolarEclipseInfo, bool) {
	return defaultModule.GeometricLocalSolarEclipseOnDateIAUSingleK(date, lon, lat, height)
}

// GeometricLocalSolarEclipseOnDateNASABulletinSplitK 当地站心几何日食查询（NASA bulletin Split-K） / local geometric solar eclipse query with NASA bulletin Split-K.
func GeometricLocalSolarEclipseOnDateNASABulletinSplitK(date time.Time, lon, lat, height float64) (LocalSolarEclipseInfo, bool) {
	return defaultModule.GeometricLocalSolarEclipseOnDateNASABulletinSplitK(date, lon, lat, height)
}

// LastGeometricLocalLunarEclipse 上次几何月食，默认 Danjon 模型 / previous geometric local lunar eclipse, Danjon by default.
func LastGeometricLocalLunarEclipse(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo {
	return defaultModule.LastGeometricLocalLunarEclipse(date, lon, lat, height)
}

// LastGeometricLocalLunarEclipseChauvenet 上次几何月食（Chauvenet） / previous geometric local lunar eclipse with Chauvenet model.
func LastGeometricLocalLunarEclipseChauvenet(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo {
	return defaultModule.LastGeometricLocalLunarEclipseChauvenet(date, lon, lat, height)
}

// LastGeometricLocalLunarEclipseDanjon 上次几何月食（Danjon） / previous geometric local lunar eclipse with Danjon model.
func LastGeometricLocalLunarEclipseDanjon(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo {
	return defaultModule.LastGeometricLocalLunarEclipseDanjon(date, lon, lat, height)
}

// LastGeometricLocalSolarEclipse 上次站心几何日食，默认 NASA bulletin Split-K 模型 / previous geometric local solar eclipse, NASA bulletin Split-K by default.
func LastGeometricLocalSolarEclipse(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo {
	return defaultModule.LastGeometricLocalSolarEclipse(date, lon, lat, height)
}

// LastGeometricLocalSolarEclipseIAUSingleK 上次站心几何日食（IAU Single-K） / previous geometric local solar eclipse with IAU Single-K.
func LastGeometricLocalSolarEclipseIAUSingleK(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo {
	return defaultModule.LastGeometricLocalSolarEclipseIAUSingleK(date, lon, lat, height)
}

// LastGeometricLocalSolarEclipseNASABulletinSplitK 上次站心几何日食（NASA bulletin Split-K） / previous geometric local solar eclipse with NASA bulletin Split-K.
func LastGeometricLocalSolarEclipseNASABulletinSplitK(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo {
	return defaultModule.LastGeometricLocalSolarEclipseNASABulletinSplitK(date, lon, lat, height)
}

// LastLocalAnnularSolarEclipse 上次站心日环食，默认 NASA bulletin Split-K 模型 / previous local annular solar eclipse, NASA bulletin Split-K by default.
func LastLocalAnnularSolarEclipse(date time.Time, lon, lat, height float64) (LocalSolarEclipseInfo, bool) {
	return defaultModule.LastLocalAnnularSolarEclipse(date, lon, lat, height)
}

// LastLocalLunarEclipse 上次可见月食，默认 Danjon 模型 / previous visible local lunar eclipse, Danjon by default.
func LastLocalLunarEclipse(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo {
	return defaultModule.LastLocalLunarEclipse(date, lon, lat, height)
}

// LastLocalLunarEclipseChauvenet 上次可见月食（Chauvenet） / previous visible local lunar eclipse with Chauvenet model.
func LastLocalLunarEclipseChauvenet(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo {
	return defaultModule.LastLocalLunarEclipseChauvenet(date, lon, lat, height)
}

// LastLocalLunarEclipseDanjon 上次可见月食（Danjon） / previous visible local lunar eclipse with Danjon model.
func LastLocalLunarEclipseDanjon(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo {
	return defaultModule.LastLocalLunarEclipseDanjon(date, lon, lat, height)
}

// LastLocalSolarEclipse 上次站心日食，默认 NASA bulletin Split-K 模型 / previous local solar eclipse, NASA bulletin Split-K by default.
func LastLocalSolarEclipse(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo {
	return defaultModule.LastLocalSolarEclipse(date, lon, lat, height)
}

// LastLocalSolarEclipseIAUSingleK 上次站心日食（IAU Single-K） / previous local solar eclipse with IAU Single-K.
func LastLocalSolarEclipseIAUSingleK(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo {
	return defaultModule.LastLocalSolarEclipseIAUSingleK(date, lon, lat, height)
}

// LastLocalSolarEclipseNASABulletinSplitK 上次站心日食（NASA bulletin Split-K） / previous local solar eclipse with NASA bulletin Split-K.
func LastLocalSolarEclipseNASABulletinSplitK(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo {
	return defaultModule.LastLocalSolarEclipseNASABulletinSplitK(date, lon, lat, height)
}

// LastLocalTotalLunarEclipse 上次可见月全食，默认 Danjon 模型 / previous visible local total lunar eclipse, Danjon by default.
func LastLocalTotalLunarEclipse(date time.Time, lon, lat, height float64) (LocalLunarEclipseInfo, bool) {
	return defaultModule.LastLocalTotalLunarEclipse(date, lon, lat, height)
}

// LastLocalTotalSolarEclipse 上次站心日全食，默认 NASA bulletin Split-K 模型 / previous local total solar eclipse, NASA bulletin Split-K by default.
func LastLocalTotalSolarEclipse(date time.Time, lon, lat, height float64) (LocalSolarEclipseInfo, bool) {
	return defaultModule.LastLocalTotalSolarEclipse(date, lon, lat, height)
}

// LastSolarEclipse 上次日食，默认 NASA bulletin Split-K 模型 / previous solar eclipse, NASA bulletin Split-K by default.
func LastSolarEclipse(date time.Time) SolarEclipseInfo { return defaultModule.LastSolarEclipse(date) }

// LastSolarEclipseIAUSingleK 上次日食（IAU Single-K） / previous solar eclipse with IAU Single-K.
func LastSolarEclipseIAUSingleK(date time.Time) SolarEclipseInfo {
	return defaultModule.LastSolarEclipseIAUSingleK(date)
}

// LastSolarEclipseNASABulletinSplitK 上次日食（NASA bulletin Split-K） / previous solar eclipse with NASA bulletin Split-K.
func LastSolarEclipseNASABulletinSplitK(date time.Time) SolarEclipseInfo {
	return defaultModule.LastSolarEclipseNASABulletinSplitK(date)
}

// LastSolarEclipseWithOptions 上次日食（自定义半径口径） / previous solar eclipse with custom radius conventions.
func LastSolarEclipseWithOptions(date time.Time, options SolarEclipseOptions) SolarEclipseInfo {
	return defaultModule.LastSolarEclipseWithOptions(date, options)
}

// LocalLunarEclipseOnDate 当地可见月食查询，默认 Danjon 模型 / local visible lunar eclipse query, Danjon by default.
func LocalLunarEclipseOnDate(date time.Time, lon, lat, height float64) (LocalLunarEclipseInfo, bool) {
	return defaultModule.LocalLunarEclipseOnDate(date, lon, lat, height)
}

// LocalLunarEclipseOnDateChauvenet 当地可见月食查询（Chauvenet） / local visible lunar eclipse query with Chauvenet model.
func LocalLunarEclipseOnDateChauvenet(date time.Time, lon, lat, height float64) (LocalLunarEclipseInfo, bool) {
	return defaultModule.LocalLunarEclipseOnDateChauvenet(date, lon, lat, height)
}

// LocalLunarEclipseOnDateDanjon 当地可见月食查询（Danjon） / local visible lunar eclipse query with Danjon model.
func LocalLunarEclipseOnDateDanjon(date time.Time, lon, lat, height float64) (LocalLunarEclipseInfo, bool) {
	return defaultModule.LocalLunarEclipseOnDateDanjon(date, lon, lat, height)
}

// LocalSolarEclipseOnDate 当地站心日食查询，默认 NASA bulletin Split-K 模型 / local topocentric solar eclipse query, NASA bulletin Split-K by default.
func LocalSolarEclipseOnDate(date time.Time, lon, lat, height float64) (LocalSolarEclipseInfo, bool) {
	return defaultModule.LocalSolarEclipseOnDate(date, lon, lat, height)
}

// LocalSolarEclipseOnDateIAUSingleK 当地站心日食查询（IAU Single-K） / local topocentric solar eclipse query with IAU Single-K.
func LocalSolarEclipseOnDateIAUSingleK(date time.Time, lon, lat, height float64) (LocalSolarEclipseInfo, bool) {
	return defaultModule.LocalSolarEclipseOnDateIAUSingleK(date, lon, lat, height)
}

// LocalSolarEclipseOnDateNASABulletinSplitK 当地站心日食查询（NASA bulletin Split-K） / local topocentric solar eclipse query with NASA bulletin Split-K.
func LocalSolarEclipseOnDateNASABulletinSplitK(date time.Time, lon, lat, height float64) (LocalSolarEclipseInfo, bool) {
	return defaultModule.LocalSolarEclipseOnDateNASABulletinSplitK(date, lon, lat, height)
}

// LocalSolarEclipseOnDateWithOptions 当地站心日食查询（自定义半径口径） / local topocentric solar eclipse query with custom radius conventions.
func LocalSolarEclipseOnDateWithOptions(date time.Time, lon, lat, height float64, options SolarEclipseOptions) (LocalSolarEclipseInfo, bool) {
	return defaultModule.LocalSolarEclipseOnDateWithOptions(date, lon, lat, height, options)
}

// LunarEclipseGeocentricPanelAt 计算给定日期月食的地心量面板，使用 Danjon 影半径模型 / computes the geocentric panel of one lunar eclipse with the Danjon shadow model.
func LunarEclipseGeocentricPanelAt(date time.Time) (LunarEclipseGeocentricPanel, bool) {
	return defaultModule.LunarEclipseGeocentricPanelAt(date)
}

// LunarEclipseGeocentricPanelChauvenet 计算给定日期月食的地心量面板，使用 Chauvenet 影半径模型 / computes the geocentric panel with the Chauvenet shadow model.
func LunarEclipseGeocentricPanelChauvenet(date time.Time) (LunarEclipseGeocentricPanel, bool) {
	return defaultModule.LunarEclipseGeocentricPanelChauvenet(date)
}

// NewSolarEclipseShadowSolver 构造单时刻求解器 / builds a single-instant solver.
func NewSolarEclipseShadowSolver(options SolarEclipseShadowSolverOptions) *SolarEclipseShadowSolver {
	return defaultModule.NewSolarEclipseShadowSolver(options)
}

// NextGeometricLocalLunarEclipse 下次几何月食，默认 Danjon 模型 / next geometric local lunar eclipse, Danjon by default.
func NextGeometricLocalLunarEclipse(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo {
	return defaultModule.NextGeometricLocalLunarEclipse(date, lon, lat, height)
}

// NextGeometricLocalLunarEclipseChauvenet 下次几何月食（Chauvenet） / next geometric local lunar eclipse with Chauvenet model.
func NextGeometricLocalLunarEclipseChauvenet(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo {
	return defaultModule.NextGeometricLocalLunarEclipseChauvenet(date, lon, lat, height)
}

// NextGeometricLocalLunarEclipseDanjon 下次几何月食（Danjon） / next geometric local lunar eclipse with Danjon model.
func NextGeometricLocalLunarEclipseDanjon(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo {
	return defaultModule.NextGeometricLocalLunarEclipseDanjon(date, lon, lat, height)
}

// NextGeometricLocalSolarEclipse 下次站心几何日食，默认 NASA bulletin Split-K 模型 / next geometric local solar eclipse, NASA bulletin Split-K by default.
func NextGeometricLocalSolarEclipse(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo {
	return defaultModule.NextGeometricLocalSolarEclipse(date, lon, lat, height)
}

// NextGeometricLocalSolarEclipseIAUSingleK 下次站心几何日食（IAU Single-K） / next geometric local solar eclipse with IAU Single-K.
func NextGeometricLocalSolarEclipseIAUSingleK(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo {
	return defaultModule.NextGeometricLocalSolarEclipseIAUSingleK(date, lon, lat, height)
}

// NextGeometricLocalSolarEclipseNASABulletinSplitK 下次站心几何日食（NASA bulletin Split-K） / next geometric local solar eclipse with NASA bulletin Split-K.
func NextGeometricLocalSolarEclipseNASABulletinSplitK(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo {
	return defaultModule.NextGeometricLocalSolarEclipseNASABulletinSplitK(date, lon, lat, height)
}

// NextLocalAnnularSolarEclipse 下次站心日环食，默认 NASA bulletin Split-K 模型 / next local annular solar eclipse, NASA bulletin Split-K by default.
func NextLocalAnnularSolarEclipse(date time.Time, lon, lat, height float64) (LocalSolarEclipseInfo, bool) {
	return defaultModule.NextLocalAnnularSolarEclipse(date, lon, lat, height)
}

// NextLocalLunarEclipse 下次可见月食，默认 Danjon 模型 / next visible local lunar eclipse, Danjon by default.
func NextLocalLunarEclipse(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo {
	return defaultModule.NextLocalLunarEclipse(date, lon, lat, height)
}

// NextLocalLunarEclipseChauvenet 下次可见月食（Chauvenet） / next visible local lunar eclipse with Chauvenet model.
func NextLocalLunarEclipseChauvenet(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo {
	return defaultModule.NextLocalLunarEclipseChauvenet(date, lon, lat, height)
}

// NextLocalLunarEclipseDanjon 下次可见月食（Danjon） / next visible local lunar eclipse with Danjon model.
func NextLocalLunarEclipseDanjon(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo {
	return defaultModule.NextLocalLunarEclipseDanjon(date, lon, lat, height)
}

// NextLocalSolarEclipse 下次站心日食，默认 NASA bulletin Split-K 模型 / next local solar eclipse, NASA bulletin Split-K by default.
func NextLocalSolarEclipse(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo {
	return defaultModule.NextLocalSolarEclipse(date, lon, lat, height)
}

// NextLocalSolarEclipseIAUSingleK 下次站心日食（IAU Single-K） / next local solar eclipse with IAU Single-K.
func NextLocalSolarEclipseIAUSingleK(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo {
	return defaultModule.NextLocalSolarEclipseIAUSingleK(date, lon, lat, height)
}

// NextLocalSolarEclipseNASABulletinSplitK 下次站心日食（NASA bulletin Split-K） / next local solar eclipse with NASA bulletin Split-K.
func NextLocalSolarEclipseNASABulletinSplitK(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo {
	return defaultModule.NextLocalSolarEclipseNASABulletinSplitK(date, lon, lat, height)
}

// NextLocalTotalLunarEclipse 下次可见月全食，默认 Danjon 模型 / next visible local total lunar eclipse, Danjon by default.
func NextLocalTotalLunarEclipse(date time.Time, lon, lat, height float64) (LocalLunarEclipseInfo, bool) {
	return defaultModule.NextLocalTotalLunarEclipse(date, lon, lat, height)
}

// NextLocalTotalSolarEclipse 下次站心日全食，默认 NASA bulletin Split-K 模型 / next local total solar eclipse, NASA bulletin Split-K by default.
func NextLocalTotalSolarEclipse(date time.Time, lon, lat, height float64) (LocalSolarEclipseInfo, bool) {
	return defaultModule.NextLocalTotalSolarEclipse(date, lon, lat, height)
}

// NextSolarEclipse 下次日食，默认 NASA bulletin Split-K 模型 / next solar eclipse, NASA bulletin Split-K by default.
func NextSolarEclipse(date time.Time) SolarEclipseInfo { return defaultModule.NextSolarEclipse(date) }

// NextSolarEclipseIAUSingleK 下次日食（IAU Single-K） / next solar eclipse with IAU Single-K.
func NextSolarEclipseIAUSingleK(date time.Time) SolarEclipseInfo {
	return defaultModule.NextSolarEclipseIAUSingleK(date)
}

// NextSolarEclipseNASABulletinSplitK 下次日食（NASA bulletin Split-K） / next solar eclipse with NASA bulletin Split-K.
func NextSolarEclipseNASABulletinSplitK(date time.Time) SolarEclipseInfo {
	return defaultModule.NextSolarEclipseNASABulletinSplitK(date)
}

// NextSolarEclipseWithOptions 下次日食（自定义半径口径） / next solar eclipse with custom radius conventions.
func NextSolarEclipseWithOptions(date time.Time, options SolarEclipseOptions) SolarEclipseInfo {
	return defaultModule.NextSolarEclipseWithOptions(date, options)
}

// SearchLocalCentralSolarEclipse 在给定跨度内搜索站心食并回报是否用尽跨度 / station search within a horizon.
func SearchLocalCentralSolarEclipse(
	date time.Time, lon, lat, height float64,
	options SolarEclipseLocalSearchOptions,
) (LocalSolarEclipseInfo, SolarEclipseLocalSearchStatus) {
	return defaultModule.SearchLocalCentralSolarEclipse(date, lon, lat, height, options)
}

// SolarEclipseBesselianElements 计算给定近朔时刻附近一次日食的多项式贝塞尔根数，窗口内无日食时返回 false / Polynomial Besselian elements for the solar eclipse near the given new-moon instant.
func SolarEclipseBesselianElements(
	seedJDE float64, options SolarEclipseBesselianElementsOptions,
) (SolarEclipseBesselianElementsResult, bool) {
	return defaultModule.SolarEclipseBesselianElements(seedJDE, options)
}

// SolarEclipseCandidates 枚举区间内的日食时刻表，按时间升序 / eclipses within a range, in time order.
func SolarEclipseCandidates(
	start, end time.Time, options SolarEclipseCandidateOptions,
) []SolarEclipseCandidate {
	return defaultModule.SolarEclipseCandidates(start, end, options)
}

// SolarEclipseCentralPath 计算指定日期附近的日食中心路径，默认使用 NASA bulletin Split-K 模型 / computes the central path near the given date, using NASA bulletin Split-K by default.
func SolarEclipseCentralPath(date time.Time, options SolarEclipsePathOptions) (SolarEclipsePath, bool) {
	return defaultModule.SolarEclipseCentralPath(date, options)
}

// SolarEclipseCentralPathIAUSingleK 使用 IAU Single-K 模型计算日食中心路径 / computes the central path with the IAU Single-K model.
func SolarEclipseCentralPathIAUSingleK(date time.Time, options SolarEclipsePathOptions) (SolarEclipsePath, bool) {
	return defaultModule.SolarEclipseCentralPathIAUSingleK(date, options)
}

// SolarEclipseCentralPathNASABulletinSplitK 使用 NASA bulletin Split-K 模型计算日食中心路径 / computes the central path with the NASA bulletin Split-K model.
func SolarEclipseCentralPathNASABulletinSplitK(date time.Time, options SolarEclipsePathOptions) (SolarEclipsePath, bool) {
	return defaultModule.SolarEclipseCentralPathNASABulletinSplitK(date, options)
}

// SolarEclipseGeocentricPanelAt 计算给定日食在食甚时刻的地心量面板 / computes the geocentric panel of one eclipse at greatest eclipse.
func SolarEclipseGeocentricPanelAt(date time.Time) (SolarEclipseGeocentricPanel, bool) {
	return defaultModule.SolarEclipseGeocentricPanelAt(date)
}

// SolarEclipseGeocentricPanelIAUSingleK 使用 IAU Single-K 计算地心量面板 / computes the geocentric panel with the IAU Single-K model.
func SolarEclipseGeocentricPanelIAUSingleK(date time.Time) (SolarEclipseGeocentricPanel, bool) {
	return defaultModule.SolarEclipseGeocentricPanelIAUSingleK(date)
}

// SolarEclipseGeocentricPanelWithOptions 使用自定义半径口径计算食甚时刻的地心量面板 / computes the geocentric panel at greatest eclipse with custom radius conventions.
func SolarEclipseGeocentricPanelWithOptions(date time.Time, options SolarEclipseOptions) (SolarEclipseGeocentricPanel, bool) {
	return defaultModule.SolarEclipseGeocentricPanelWithOptions(date, options)
}

// SolarEclipseOnDate 当地自然日全局日食查询，默认 NASA bulletin Split-K 模型 / local-date global solar eclipse query, NASA bulletin Split-K by default.
func SolarEclipseOnDate(date time.Time) (SolarEclipseInfo, bool) {
	return defaultModule.SolarEclipseOnDate(date)
}

// SolarEclipseOnDateIAUSingleK 当地自然日全局日食查询（IAU Single-K） / local-date global solar eclipse query with IAU Single-K.
func SolarEclipseOnDateIAUSingleK(date time.Time) (SolarEclipseInfo, bool) {
	return defaultModule.SolarEclipseOnDateIAUSingleK(date)
}

// SolarEclipseOnDateNASABulletinSplitK 当地自然日全局日食查询（NASA bulletin Split-K） / local-date global solar eclipse query with NASA bulletin Split-K.
func SolarEclipseOnDateNASABulletinSplitK(date time.Time) (SolarEclipseInfo, bool) {
	return defaultModule.SolarEclipseOnDateNASABulletinSplitK(date)
}

// SolarEclipseOnDateWithOptions 当地自然日全局日食查询（自定义半径口径） / local-date global solar eclipse query with custom radius conventions.
func SolarEclipseOnDateWithOptions(date time.Time, options SolarEclipseOptions) (SolarEclipseInfo, bool) {
	return defaultModule.SolarEclipseOnDateWithOptions(date, options)
}

// SolarEclipsePartialArea 计算半影足迹，是 SolarEclipsePartialFootprints 的兼容包装 / computes penumbral footprint samples and is a compatibility wrapper for SolarEclipsePartialFootprints.
func SolarEclipsePartialArea(date time.Time, options SolarEclipsePartialAreaOptions) (SolarEclipsePartialAreaInfo, bool) {
	return defaultModule.SolarEclipsePartialArea(date, options)
}

// SolarEclipsePartialAreaIAUSingleK 是 SolarEclipsePartialFootprintsIAUSingleK 的兼容包装 / is a compatibility wrapper for SolarEclipsePartialFootprintsIAUSingleK.
func SolarEclipsePartialAreaIAUSingleK(date time.Time, options SolarEclipsePartialAreaOptions) (SolarEclipsePartialAreaInfo, bool) {
	return defaultModule.SolarEclipsePartialAreaIAUSingleK(date, options)
}

// SolarEclipsePartialAreaNASABulletinSplitK 是 SolarEclipsePartialFootprintsNASABulletinSplitK 的兼容包装 / is a compatibility wrapper for SolarEclipsePartialFootprintsNASABulletinSplitK.
func SolarEclipsePartialAreaNASABulletinSplitK(date time.Time, options SolarEclipsePartialAreaOptions) (SolarEclipsePartialAreaInfo, bool) {
	return defaultModule.SolarEclipsePartialAreaNASABulletinSplitK(date, options)
}

// SolarEclipsePartialFootprints 计算指定日期附近的日食半影足迹，默认使用 NASA bulletin Split-K 模型 / computes penumbral footprint samples near the given date, using NASA bulletin Split-K by default.
func SolarEclipsePartialFootprints(date time.Time, options SolarEclipsePartialFootprintOptions) (SolarEclipsePartialFootprintsInfo, bool) {
	return defaultModule.SolarEclipsePartialFootprints(date, options)
}

// SolarEclipsePartialFootprintsIAUSingleK 使用 IAU Single-K 模型计算日食半影足迹 / computes penumbral footprint samples with the IAU Single-K model.
func SolarEclipsePartialFootprintsIAUSingleK(date time.Time, options SolarEclipsePartialFootprintOptions) (SolarEclipsePartialFootprintsInfo, bool) {
	return defaultModule.SolarEclipsePartialFootprintsIAUSingleK(date, options)
}

// SolarEclipsePartialFootprintsNASABulletinSplitK 使用 NASA bulletin Split-K 模型计算日食半影足迹 / computes penumbral footprint samples with the NASA bulletin Split-K model.
func SolarEclipsePartialFootprintsNASABulletinSplitK(date time.Time, options SolarEclipsePartialFootprintOptions) (SolarEclipsePartialFootprintsInfo, bool) {
	return defaultModule.SolarEclipsePartialFootprintsNASABulletinSplitK(date, options)
}

// SolarEclipseShadowAt 无状态便捷入口，可并发调用 / stateless convenience entry point.
func SolarEclipseShadowAt(value time.Time) (SolarEclipseShadowInstant, bool) {
	return defaultModule.SolarEclipseShadowAt(value)
}

// SolarEclipseShadowBetween 无状态批量入口，使用进程级 ΔT / stateless batch entry point.
func SolarEclipseShadowBetween(
	start, end time.Time, step time.Duration,
) []SolarEclipseShadowInstant {
	return defaultModule.SolarEclipseShadowBetween(start, end, step)
}

// SolarEclipseShadowBetweenWithOptions 可显式指定 ΔT 与模型 / with explicit ΔT and model.
func SolarEclipseShadowBetweenWithOptions(
	start, end time.Time, step time.Duration, options SolarEclipseShadowSolverOptions,
) []SolarEclipseShadowInstant {
	return defaultModule.SolarEclipseShadowBetweenWithOptions(start, end, step, options)
}

// SolarEclipseStationStateAt 无状态便捷入口，可并发调用 / stateless convenience entry point.
func SolarEclipseStationStateAt(value time.Time, lon, lat, height float64) SolarEclipseStationState {
	return defaultModule.SolarEclipseStationStateAt(value, lon, lat, height)
}
