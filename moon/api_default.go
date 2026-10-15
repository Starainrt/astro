package moon

import (
	"time"
)

// Altitude 月亮高度角，date 的时区参与地方时计算 / lunar altitude; the zone of date enters local time.
func Altitude(date time.Time, lon, lat float64) float64 {
	return defaultModule.Altitude(date, lon, lat)
}

// ApogeesInMonth 指定年月内的所有月球远地点 / all lunar apogees in the given Gregorian month.
func ApogeesInMonth(year int, month time.Month) []ApsisInfo {
	return defaultModule.ApogeesInMonth(year, month)
}

// ApparentAltitude 月亮视高度角 / apparent lunar altitude.
func ApparentAltitude(date time.Time, lon, lat, pressureHPa, temperatureC float64) float64 {
	return defaultModule.ApparentAltitude(date, lon, lat, pressureHPa, temperatureC)
}

// ApparentDec 月亮站心视赤纬 / apparent topocentric declination.
func ApparentDec(date time.Time, lon, lat float64) float64 {
	return defaultModule.ApparentDec(date, lon, lat)
}

// ApparentLo 月亮地心视黄经 / apparent geocentric ecliptic longitude.
func ApparentLo(date time.Time) float64 { return defaultModule.ApparentLo(date) }

// ApparentRa 月亮站心视赤经，date 的时区参与地方时计算 / apparent topocentric right ascension; the zone of date enters local time.
func ApparentRa(date time.Time, lon, lat float64) float64 {
	return defaultModule.ApparentRa(date, lon, lat)
}

// ApparentRaDec 月亮站心视赤经、视赤纬 / apparent topocentric right ascension and declination.
func ApparentRaDec(date time.Time, lon, lat float64) (float64, float64) {
	return defaultModule.ApparentRaDec(date, lon, lat)
}

// ApparentZenith 月亮视天顶距 / apparent lunar zenith distance.
func ApparentZenith(date time.Time, lon, lat, pressureHPa, temperatureC float64) float64 {
	return defaultModule.ApparentZenith(date, lon, lat, pressureHPa, temperatureC)
}

// AscendingNode 月球升交点黄经 / ascending node longitude of the Moon.
func AscendingNode(date time.Time) float64 { return defaultModule.AscendingNode(date) }

// AscendingNodeN 月球升交点黄经（n 变体） / n-parameterised ascending node longitude of the Moon.
func AscendingNodeN(date time.Time, n int) float64 { return defaultModule.AscendingNodeN(date, n) }

// Azimuth 月亮方位角，正北为 0° 向东增加 / lunar azimuth from north through east.
func Azimuth(date time.Time, lon, lat float64) float64 { return defaultModule.Azimuth(date, lon, lat) }

// BrightLimbPositionAngle 月亮明亮边缘位置角，单位度 / position angle of the Moon's bright limb in degrees.
func BrightLimbPositionAngle(date time.Time) float64 {
	return defaultModule.BrightLimbPositionAngle(date)
}

// BrightLimbPositionAngleN 月亮明亮边缘位置角（n 变体），单位度 / n-parameterised position angle of the Moon's bright limb in degrees.
func BrightLimbPositionAngleN(date time.Time, n int) float64 {
	return defaultModule.BrightLimbPositionAngleN(date, n)
}

// ClosestConjunctionWithPlanet 最近一次行星合月（赤经合） / closest Moon-planet conjunction.
func ClosestConjunctionWithPlanet(date time.Time, planet ConjunctionPlanet) time.Time {
	return defaultModule.ClosestConjunctionWithPlanet(date, planet)
}

// ClosestFirstQuarter 最近上弦英文别名 / English alias for ClosestShangXianYue.
func ClosestFirstQuarter(date time.Time) time.Time { return defaultModule.ClosestFirstQuarter(date) }

// ClosestFullMoon 最近望月英文别名 / English alias for ClosestWangYue.
func ClosestFullMoon(date time.Time) time.Time { return defaultModule.ClosestFullMoon(date) }

// ClosestLastQuarter 最近下弦英文别名 / English alias for ClosestXiaXianYue.
func ClosestLastQuarter(date time.Time) time.Time { return defaultModule.ClosestLastQuarter(date) }

// ClosestMaximumNorthDeclination 离指定时刻最近一次月球最大北赤纬 / closest maximum northern lunar declination to date.
func ClosestMaximumNorthDeclination(date time.Time) MaximumDeclinationInfo {
	return defaultModule.ClosestMaximumNorthDeclination(date)
}

// ClosestMaximumSouthDeclination 离指定时刻最近一次月球最大南赤纬 / closest maximum southern lunar declination to date.
func ClosestMaximumSouthDeclination(date time.Time) MaximumDeclinationInfo {
	return defaultModule.ClosestMaximumSouthDeclination(date)
}

// ClosestNewMoon 最近朔月英文别名 / English alias for ClosestShuoYue.
func ClosestNewMoon(date time.Time) time.Time { return defaultModule.ClosestNewMoon(date) }

// ClosestShangXianYue 最近上弦，结果保持 date 的时区 / closest first quarter in date's zone.
func ClosestShangXianYue(date time.Time) time.Time { return defaultModule.ClosestShangXianYue(date) }

// ClosestShuoYue 最近朔月，结果保持 date 的时区 / closest new moon in date's zone.
func ClosestShuoYue(date time.Time) time.Time { return defaultModule.ClosestShuoYue(date) }

// ClosestWangYue 最近望月，结果保持 date 的时区 / closest full moon in date's zone.
func ClosestWangYue(date time.Time) time.Time { return defaultModule.ClosestWangYue(date) }

// ClosestXiaXianYue 最近下弦，结果保持 date 的时区 / closest last quarter in date's zone.
func ClosestXiaXianYue(date time.Time) time.Time { return defaultModule.ClosestXiaXianYue(date) }

// CulminationTime 月亮中天时刻，取 date 时区的当地日期 / culmination time on date's local day.
func CulminationTime(date time.Time, lon, lat float64) time.Time {
	return defaultModule.CulminationTime(date, lon, lat)
}

// DescendingNode 月球降交点黄经 / descending node longitude of the Moon.
func DescendingNode(date time.Time) float64 { return defaultModule.DescendingNode(date) }

// DescendingNodeN 月球降交点黄经（n 变体） / n-parameterised descending node longitude of the Moon.
func DescendingNodeN(date time.Time, n int) float64 { return defaultModule.DescendingNodeN(date, n) }

// Diameter 月亮视直径，单位角秒 / apparent lunar diameter in arcseconds.
func Diameter(date time.Time) float64 { return defaultModule.Diameter(date) }

// DiameterN 月亮视直径（n 变体），单位角秒 / n-parameterised apparent lunar diameter in arcseconds.
func DiameterN(date time.Time, n int) float64 { return defaultModule.DiameterN(date, n) }

// DownTime 月落时刻别名，同 SetTime / deprecated alias of SetTime.
func DownTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	return defaultModule.DownTime(date, lon, lat, height, aero)
}

// EarthDistance 地月距离（千米） / Earth-Moon distance in kilometers.
func EarthDistance(date time.Time) float64 {
	return defaultModule.EarthDistance(date)
}

// FindBestPlanetOccultations 窗口内每次行星月掩的全球海平面几何掩甚点，可见性不参与选点 / global sea-level geometric greatest point of each planetary occultation.
func FindBestPlanetOccultations(start, end time.Time, planet OccultationPlanet,
	options OccultationSearchOptions) ([]PlanetOccultationInfo, error) {
	return defaultModule.FindBestPlanetOccultations(start, end, planet, options)
}

// FindBestStarOccultations 窗口内每次恒星月掩的全球海平面几何掩甚点，可见性不参与选点 / global sea-level geometric greatest point of each stellar occultation.
func FindBestStarOccultations(start, end time.Time, star StarCoordinate, options OccultationSearchOptions) ([]StarOccultationInfo, error) {
	return defaultModule.FindBestStarOccultations(start, end, star, options)
}

// FindPlanetOccultationPaths 窗口内每次行星月掩的全球外接触掩带 / global outer-contact footprint of each planetary occultation.
func FindPlanetOccultationPaths(start, end time.Time, planet OccultationPlanet,
	options OccultationPathOptions) ([]PlanetOccultationPath, error) {
	return defaultModule.FindPlanetOccultationPaths(start, end, planet, options)
}

// FindPlanetOccultations 固定地点的有限盘面行星月掩，忽略行星环 / planetary occultations at a fixed site, rings excluded.
func FindPlanetOccultations(start, end time.Time, planet OccultationPlanet, longitude, latitude, height float64,
	options OccultationSearchOptions) ([]PlanetOccultationInfo, error) {
	return defaultModule.FindPlanetOccultations(start, end, planet, longitude, latitude, height, options)
}

// FindStarOccultationPaths 点源恒星月掩的全球掩带，不加载内嵌星表 / global footprint of a stellar occultation; no embedded catalog is loaded.
func FindStarOccultationPaths(start, end time.Time, star StarCoordinate, options OccultationPathOptions) ([]StarOccultationPath, error) {
	return defaultModule.FindStarOccultationPaths(start, end, star, options)
}

// FindStarOccultations 固定地点的点源恒星月掩，不加载内嵌星表 / stellar occultations at a fixed site; no embedded catalog is loaded.
func FindStarOccultations(start, end time.Time, star StarCoordinate, longitude, latitude, height float64,
	options OccultationSearchOptions) ([]StarOccultationInfo, error) {
	return defaultModule.FindStarOccultations(start, end, star, longitude, latitude, height, options)
}

// FirstQuarter 上弦英文别名 / English alias for ShangXianYue.
func FirstQuarter(year float64) time.Time { return defaultModule.FirstQuarter(year) }

// FullMoon 望月英文别名 / English alias for WangYue.
func FullMoon(year float64) time.Time { return defaultModule.FullMoon(year) }

// GeocentricApparentDec 月亮地心视赤纬 / apparent geocentric declination.
func GeocentricApparentDec(date time.Time) float64 { return defaultModule.GeocentricApparentDec(date) }

// GeocentricApparentRa 月亮地心视赤经 / apparent geocentric right ascension.
func GeocentricApparentRa(date time.Time) float64 { return defaultModule.GeocentricApparentRa(date) }

// GeocentricApparentRaDec 月亮地心视赤经、视赤纬 / apparent geocentric right ascension and declination.
func GeocentricApparentRaDec(date time.Time) (float64, float64) {
	return defaultModule.GeocentricApparentRaDec(date)
}

// HourAngle 月亮时角，date 的时区参与地方时计算 / lunar hour angle; the zone of date enters local time.
func HourAngle(date time.Time, lon, lat float64) float64 {
	return defaultModule.HourAngle(date, lon, lat)
}

// LastConjunctionWithPlanet 上一次行星合月（赤经合） / previous Moon-planet conjunction.
func LastConjunctionWithPlanet(date time.Time, planet ConjunctionPlanet) time.Time {
	return defaultModule.LastConjunctionWithPlanet(date, planet)
}

// LastFirstQuarter 上一次上弦英文别名 / English alias for LastShangXianYue.
func LastFirstQuarter(date time.Time) time.Time { return defaultModule.LastFirstQuarter(date) }

// LastFullMoon 上一次望月英文别名 / English alias for LastWangYue.
func LastFullMoon(date time.Time) time.Time { return defaultModule.LastFullMoon(date) }

// LastLastQuarter 上一次下弦英文别名 / English alias for LastXiaXianYue.
func LastLastQuarter(date time.Time) time.Time { return defaultModule.LastLastQuarter(date) }

// LastMaximumNorthDeclination 指定时刻之前最近一次月球最大北赤纬 / last maximum northern lunar declination at or before date.
func LastMaximumNorthDeclination(date time.Time) MaximumDeclinationInfo {
	return defaultModule.LastMaximumNorthDeclination(date)
}

// LastMaximumSouthDeclination 指定时刻之前最近一次月球最大南赤纬 / last maximum southern lunar declination at or before date.
func LastMaximumSouthDeclination(date time.Time) MaximumDeclinationInfo {
	return defaultModule.LastMaximumSouthDeclination(date)
}

// LastNewMoon 上一次朔月英文别名 / English alias for LastShuoYue.
func LastNewMoon(date time.Time) time.Time { return defaultModule.LastNewMoon(date) }

// LastQuarter 下弦英文别名 / English alias for XiaXianYue.
func LastQuarter(year float64) time.Time { return defaultModule.LastQuarter(year) }

// LastShangXianYue 上一次上弦，结果保持 date 的时区 / previous first quarter in date's zone.
func LastShangXianYue(date time.Time) time.Time { return defaultModule.LastShangXianYue(date) }

// LastShuoYue 上一次朔月，结果保持 date 的时区 / previous new moon in date's zone.
func LastShuoYue(date time.Time) time.Time { return defaultModule.LastShuoYue(date) }

// LastWangYue 上一次望月，结果保持 date 的时区 / previous full moon in date's zone.
func LastWangYue(date time.Time) time.Time { return defaultModule.LastWangYue(date) }

// LastXiaXianYue 上一次下弦，结果保持 date 的时区 / previous last quarter in date's zone.
func LastXiaXianYue(date time.Time) time.Time { return defaultModule.LastXiaXianYue(date) }

// MaximumNorthDeclinationsInMonth 指定年月内的所有月球最大北赤纬事件 / all maximum northern lunar declination events in the given Gregorian month.
func MaximumNorthDeclinationsInMonth(year int, month time.Month) []MaximumDeclinationInfo {
	return defaultModule.MaximumNorthDeclinationsInMonth(year, month)
}

// MaximumSouthDeclinationsInMonth 指定年月内的所有月球最大南赤纬事件 / all maximum southern lunar declination events in the given Gregorian month.
func MaximumSouthDeclinationsInMonth(year int, month time.Month) []MaximumDeclinationInfo {
	return defaultModule.MaximumSouthDeclinationsInMonth(year, month)
}

// NewMoon 朔月英文别名 / English alias for ShuoYue.
func NewMoon(year float64) time.Time { return defaultModule.NewMoon(year) }

// NextConjunctionWithPlanet 下一次行星合月（赤经合） / next Moon-planet conjunction.
func NextConjunctionWithPlanet(date time.Time, planet ConjunctionPlanet) time.Time {
	return defaultModule.NextConjunctionWithPlanet(date, planet)
}

// NextFirstQuarter 下一次上弦英文别名 / English alias for NextShangXianYue.
func NextFirstQuarter(date time.Time) time.Time { return defaultModule.NextFirstQuarter(date) }

// NextFullMoon 下一次望月英文别名 / English alias for NextWangYue.
func NextFullMoon(date time.Time) time.Time { return defaultModule.NextFullMoon(date) }

// NextLastQuarter 下一次下弦英文别名 / English alias for NextXiaXianYue.
func NextLastQuarter(date time.Time) time.Time { return defaultModule.NextLastQuarter(date) }

// NextMaximumNorthDeclination 指定时刻之后最近一次月球最大北赤纬 / next maximum northern lunar declination after date.
func NextMaximumNorthDeclination(date time.Time) MaximumDeclinationInfo {
	return defaultModule.NextMaximumNorthDeclination(date)
}

// NextMaximumSouthDeclination 指定时刻之后最近一次月球最大南赤纬 / next maximum southern lunar declination after date.
func NextMaximumSouthDeclination(date time.Time) MaximumDeclinationInfo {
	return defaultModule.NextMaximumSouthDeclination(date)
}

// NextNewMoon 下一次朔月英文别名 / English alias for NextShuoYue.
func NextNewMoon(date time.Time) time.Time { return defaultModule.NextNewMoon(date) }

// NextShangXianYue 下一次上弦，结果保持 date 的时区 / next first quarter in date's zone.
func NextShangXianYue(date time.Time) time.Time { return defaultModule.NextShangXianYue(date) }

// NextShuoYue 下一次朔月，结果保持 date 的时区 / next new moon in date's zone.
func NextShuoYue(date time.Time) time.Time { return defaultModule.NextShuoYue(date) }

// NextWangYue 下一次望月，结果保持 date 的时区 / next full moon in date's zone.
func NextWangYue(date time.Time) time.Time { return defaultModule.NextWangYue(date) }

// NextXiaXianYue 下一次下弦，结果保持 date 的时区 / next last quarter in date's zone.
func NextXiaXianYue(date time.Time) time.Time { return defaultModule.NextXiaXianYue(date) }

// ParallacticAngle 月亮站心视差角（天顶方向角） / topocentric lunar parallactic angle.
func ParallacticAngle(date time.Time, lon, lat float64) float64 {
	return defaultModule.ParallacticAngle(date, lon, lat)
}

// PerigeesInMonth 指定年月内的所有月球近地点 / all lunar perigees in the given Gregorian month.
func PerigeesInMonth(year int, month time.Month) []ApsisInfo {
	return defaultModule.PerigeesInMonth(year, month)
}

// Phase 月面受照比例，取值 [0, 1] / illuminated fraction in [0, 1].
func Phase(date time.Time) float64 { return defaultModule.Phase(date) }

// PhaseDesc 中文月相名称 / Chinese lunar phase name.
func PhaseDesc(date time.Time) string { return defaultModule.PhaseDesc(date) }

// Physical 月球物理观测参数 / physical observing parameters of the Moon.
func Physical(date time.Time) PhysicalInfo { return defaultModule.Physical(date) }

// PhysicalN 月球物理观测参数（n 变体） / n-parameterised physical observing parameters of the Moon.
func PhysicalN(date time.Time, n int) PhysicalInfo { return defaultModule.PhysicalN(date, n) }

// PlanetOccultationDiagram 已求解固定地点行星月掩的图形几何 / diagram geometry of a solved fixed-site planetary occultation.
func PlanetOccultationDiagram(
	info PlanetOccultationInfo,
	options PlanetOccultationDiagramOptions,
) PlanetOccultationDiagramResult {
	return defaultModule.PlanetOccultationDiagram(info, options)
}

// PlanetOccultationFootprintsAt 指定时刻行星偏掩、全掩可见足迹 / visible partial and total planetary footprints at one instant.
func PlanetOccultationFootprintsAt(at time.Time, planet OccultationPlanet) (PlanetOccultationInstant, error) {
	return defaultModule.PlanetOccultationFootprintsAt(at, planet)
}

// RiseTime 月出时刻，取 date 时区的当地日期；height 为椭球高（米），aero 为真时按大气折射与实时视半径算上缘 / moonrise on date's local day.
func RiseTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	return defaultModule.RiseTime(date, lon, lat, height, aero)
}

// Semidiameter 月亮视半径，单位角秒 / apparent lunar semidiameter in arcseconds.
func Semidiameter(date time.Time) float64 { return defaultModule.Semidiameter(date) }

// SemidiameterN 月亮视半径（n 变体），单位角秒 / n-parameterised apparent lunar semidiameter in arcseconds.
func SemidiameterN(date time.Time, n int) float64 { return defaultModule.SemidiameterN(date, n) }

// SetTime 月落时刻，参数同 RiseTime / moonset on date's local day.
func SetTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	return defaultModule.SetTime(date, lon, lat, height, aero)
}

// ShangXianYue 上弦锚点解，year 为公历小数年，结果为 UTC / first quarter near a decimal-year anchor, in UTC.
func ShangXianYue(year float64) time.Time { return defaultModule.ShangXianYue(year) }

// ShuoYue 朔月锚点解，year 为公历小数年，结果为 UTC / new moon near a decimal-year anchor, in UTC.
func ShuoYue(year float64) time.Time { return defaultModule.ShuoYue(year) }

// StarOccultationLimbProfile 精密恒星月掩：给出全部遮挡区间与带内采样，limb 为 nil 时用模块月缘 / precise stellar occultation; nil limb uses the module profile.
func StarOccultationLimbProfile(info StarOccultationInfo, star StarCoordinate, limb LunarLimb, options StarOccultationLimbProfileOptions) (StarOccultationLimbProfileResult, error) {
	return defaultModule.StarOccultationLimbProfile(info, star, limb, options)
}

// StarOccultationDiagram 已求解固定地点恒星月掩的图形几何 / diagram geometry of a solved fixed-site stellar occultation.
func StarOccultationDiagram(
	info StarOccultationInfo,
	star StarCoordinate,
	options StarOccultationDiagramOptions,
) StarOccultationDiagramResult {
	return defaultModule.StarOccultationDiagram(info, star, options)
}

// StarOccultationFootprintAt 指定时刻点源恒星月掩可见足迹 / visible stellar-occultation footprint at one instant.
func StarOccultationFootprintAt(at time.Time, star StarCoordinate) (StarOccultationInstant, error) {
	return defaultModule.StarOccultationFootprintAt(at, star)
}

// SunMoonLoDiff 月日视黄经差，取 [0, 360)，朔近 0°、望近 180° / Moon minus Sun apparent longitude in [0, 360).
func SunMoonLoDiff(date time.Time) float64 { return defaultModule.SunMoonLoDiff(date) }

// TopocentricBrightLimbPositionAngle 月亮站心明亮边缘位置角（度），height 为椭球高（米） / topocentric bright-limb position angle in degrees.
func TopocentricBrightLimbPositionAngle(date time.Time, observerLon, observerLat, height float64) float64 {
	return defaultModule.TopocentricBrightLimbPositionAngle(date, observerLon, observerLat, height)
}

// TopocentricBrightLimbPositionAngleN 月亮站心明亮边缘位置角（n 变体），单位度 / n-parameterised topocentric position angle of the Moon's bright limb in degrees.
func TopocentricBrightLimbPositionAngleN(date time.Time, observerLon, observerLat, height float64, n int) float64 {
	return defaultModule.TopocentricBrightLimbPositionAngleN(date, observerLon, observerLat, height, n)
}

// TopocentricPhysical 月球站心物理观测参数，height 为椭球高（米） / topocentric physical observing parameters of the Moon.
func TopocentricPhysical(date time.Time, observerLon, observerLat, height float64) PhysicalInfo {
	return defaultModule.TopocentricPhysical(date, observerLon, observerLat, height)
}

// TopocentricPhysicalN 月球站心物理观测参数（n 变体） / n-parameterised topocentric physical observing parameters of the Moon.
func TopocentricPhysicalN(date time.Time, observerLon, observerLat, height float64, n int) PhysicalInfo {
	return defaultModule.TopocentricPhysicalN(date, observerLon, observerLat, height, n)
}

// TrueBo 月亮地心真黄纬 / geocentric true ecliptic latitude.
func TrueBo(date time.Time) float64 { return defaultModule.TrueBo(date) }

// TrueBoN 月亮地心真黄纬，n<0 取完整级数，n>=0 时半解析模型用快速表、Legacy 按 n 截断 / geocentric true ecliptic latitude with series selector n.
func TrueBoN(date time.Time, n int) float64 { return defaultModule.TrueBoN(date, n) }

// TrueDec 月亮地心真赤纬 / geocentric true declination.
func TrueDec(date time.Time) float64 { return defaultModule.TrueDec(date) }

// TrueLo 月亮地心真黄经 / geocentric true ecliptic longitude.
func TrueLo(date time.Time) float64 { return defaultModule.TrueLo(date) }

// TrueLoN 月亮地心真黄经，n<0 取完整级数，n>=0 时半解析模型用快速表、Legacy 按 n 截断 / geocentric true ecliptic longitude with series selector n.
func TrueLoN(date time.Time, n int) float64 { return defaultModule.TrueLoN(date, n) }

// TrueRa 月亮地心真赤经 / geocentric true right ascension.
func TrueRa(date time.Time) float64 { return defaultModule.TrueRa(date) }

// TrueRaDec 月亮地心真赤经、真赤纬 / geocentric true right ascension and declination.
func TrueRaDec(date time.Time) (float64, float64) { return defaultModule.TrueRaDec(date) }

// WangYue 望月锚点解，year 为公历小数年，结果为 UTC / full moon near a decimal-year anchor, in UTC.
func WangYue(year float64) time.Time { return defaultModule.WangYue(year) }

// XiaXianYue 下弦锚点解，year 为公历小数年，结果为 UTC / last quarter near a decimal-year anchor, in UTC.
func XiaXianYue(year float64) time.Time { return defaultModule.XiaXianYue(year) }

// Zenith 月亮天顶距 / lunar zenith distance.
func Zenith(date time.Time, lon, lat float64) float64 { return defaultModule.Zenith(date, lon, lat) }

// RiseTimeWithAtmosphere 月出时刻，按给定气压（hPa）与气温（°C）计算大气折射 / moonrise time with explicit pressure and temperature.
func RiseTimeWithAtmosphere(date time.Time, lon, lat, height, pressureHPa, temperatureC float64) (time.Time, error) {
	return defaultModule.RiseTimeWithAtmosphere(date, lon, lat, height, pressureHPa, temperatureC)
}

// SetTimeWithAtmosphere 月落时刻，按给定气压（hPa）与气温（°C）计算大气折射 / moonset time with explicit pressure and temperature.
func SetTimeWithAtmosphere(date time.Time, lon, lat, height, pressureHPa, temperatureC float64) (time.Time, error) {
	return defaultModule.SetTimeWithAtmosphere(date, lon, lat, height, pressureHPa, temperatureC)
}
