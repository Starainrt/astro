package basic

import (
	"time"
)

// ClosestJupiterGalileanPhenomenonContactEvent 最近一次 IMCCE 风格接触事件 / closest IMCCE-style contact event.
func ClosestJupiterGalileanPhenomenonContactEvent(jd float64, satellite int, phenomenonType JupiterGalileanPhenomenonType) JupiterGalileanPhenomenonContactEvent {
	return ClosestJupiterGalileanPhenomenonContactEventWithInstance(jd, satellite, phenomenonType, nil)
}

// ClosestJupiterGalileanPhenomenonEvent 最近一次伽利略卫星现象，jd 与返回时刻均为民用儒略日 / closest Galilean-satellite event in civil Julian days.
func ClosestJupiterGalileanPhenomenonEvent(jd float64, satellite int, phenomenonType JupiterGalileanPhenomenonType) JupiterGalileanPhenomenonEvent {
	return ClosestJupiterGalileanPhenomenonEventWithInstance(jd, satellite, phenomenonType, nil)
}

// ClosestMercuryTransit 最近一次全球水星凌日 / closest global transit of Mercury.
func ClosestMercuryTransit(jde float64) PlanetTransitResult {
	return ClosestMercuryTransitWithInstance(jde, nil)
}

// ClosestMoonMaximumNorthDeclination 离指定时刻最近一次月球最大北赤纬 / closest maximum northern lunar declination to jd.
func ClosestMoonMaximumNorthDeclination(jd float64) DeclinationEvent {
	return ClosestMoonMaximumNorthDeclinationWithInstance(jd, nil)
}

// ClosestMoonMaximumSouthDeclination 离指定时刻最近一次月球最大南赤纬 / closest maximum southern lunar declination to jd.
func ClosestMoonMaximumSouthDeclination(jd float64) DeclinationEvent {
	return ClosestMoonMaximumSouthDeclinationWithInstance(jd, nil)
}

// ClosestMoonPlanetConjunction 离指定时刻最近一次行星合月（赤经合） / closest Moon-planet conjunction to jde.
func ClosestMoonPlanetConjunction(jde float64, planet MoonPlanetConjunctionPlanet) float64 {
	return ClosestMoonPlanetConjunctionWithInstance(jde, planet, nil)
}

// ClosestVenusTransit 最近一次全球金星凌日 / closest global transit of Venus.
func ClosestVenusTransit(jde float64) PlanetTransitResult {
	return ClosestVenusTransitWithInstance(jde, nil)
}

// EveningTwilight 昏朦影，localJD 为当地 0 时民用日锚点且只取整数日 / evening twilight on the local civil day anchored at localJD.
func EveningTwilight(localJD, lon, lat, tz, targetAltitude float64) (float64, error) {
	return EveningTwilightWithInstance(localJD, lon, lat, tz, targetAltitude, nil)
}

// FindBestPlanetOccultations 窗口内每次行星月掩的全球海平面几何掩甚点，可见性不参与选点 / global sea-level geometric greatest point of each planetary occultation.
func FindBestPlanetOccultations(start, end time.Time, planet OccultationPlanet,
	options OccultationSearchOptions) ([]PlanetOccultationInfo, error) {
	return FindBestPlanetOccultationsWithInstance(start, end, planet, options, nil)
}

// FindBestStarOccultations 窗口内每次恒星月掩的全球海平面几何掩甚点，可见性不参与选点 / global sea-level geometric greatest point of each stellar occultation.
func FindBestStarOccultations(start, end time.Time, star StarCoordinate, options OccultationSearchOptions) ([]StarOccultationInfo, error) {
	return FindBestStarOccultationsWithInstance(start, end, star, options, nil)
}

// FindPlanetOccultationPaths 行星月掩的全球外、内接触掩带，按全球掩甚点选取事件 / global outer- and inner-contact footprints of planetary occultations.
func FindPlanetOccultationPaths(start, end time.Time, planet OccultationPlanet,
	options OccultationPathOptions) ([]PlanetOccultationPath, error) {
	return FindPlanetOccultationPathsWithInstance(start, end, planet, options, nil)
}

// FindPlanetOccultations 固定观测点的有限盘面行星月掩 / planetary occultations at a fixed site.
func FindPlanetOccultations(start, end time.Time, planet OccultationPlanet, longitude, latitude, height float64,
	options OccultationSearchOptions) ([]PlanetOccultationInfo, error) {
	return FindPlanetOccultationsWithInstance(start, end, planet, longitude, latitude, height, options, nil)
}

// FindStarOccultationPaths 点源恒星月掩的全球掩带，不加载内嵌星表 / global footprint of a stellar occultation; no embedded catalog is loaded.
func FindStarOccultationPaths(start, end time.Time, star StarCoordinate, options OccultationPathOptions) ([]StarOccultationPath, error) {
	return FindStarOccultationPathsWithInstance(start, end, star, options, nil)
}

// FindStarOccultations 固定观测点的点源恒星月掩，不加载内嵌星表 / stellar occultations at a fixed site; no embedded catalog is loaded.
func FindStarOccultations(start, end time.Time, star StarCoordinate, longitude, latitude, height float64,
	options OccultationSearchOptions) ([]StarOccultationInfo, error) {
	return FindStarOccultationsWithInstance(start, end, star, longitude, latitude, height, options, nil)
}

// GetSunRiseTime 日出时刻，julianDay 为当地 0 时民用日锚点 / sunrise on the local civil day anchored at julianDay.
func GetSunRiseTime(julianDay, longitude, latitude, timeZone, zenithShift, height float64) (float64, error) {
	return GetSunRiseTimeWithInstance(julianDay, longitude, latitude, timeZone, zenithShift, height, nil)
}

// GetSunSetTime 日落时刻，julianDay 为当地 0 时民用日锚点 / sunset on the local civil day anchored at julianDay.
func GetSunSetTime(julianDay, longitude, latitude, timeZone, zenithShift, height float64) (float64, error) {
	return GetSunSetTimeWithInstance(julianDay, longitude, latitude, timeZone, zenithShift, height, nil)
}

// HMoonGeocentricApparentDec 月亮地心视赤纬 / apparent geocentric declination of the Moon.
func HMoonGeocentricApparentDec(jde float64) float64 {
	return HMoonGeocentricApparentDecWithInstance(jde, nil)
}

// HMoonGeocentricApparentDecN 月亮地心视赤纬（n 变体） / n-parameterised apparent geocentric declination of the Moon.
func HMoonGeocentricApparentDecN(jde float64, n int) float64 {
	return HMoonGeocentricApparentDecNWithInstance(jde, n, nil)
}

// HMoonGeocentricApparentRa 月亮地心视赤经 / apparent geocentric right ascension of the Moon.
func HMoonGeocentricApparentRa(jde float64) float64 {
	return HMoonGeocentricApparentRaWithInstance(jde, nil)
}

// HMoonGeocentricApparentRaN 月亮地心视赤经（n 变体） / n-parameterised apparent geocentric right ascension of the Moon.
func HMoonGeocentricApparentRaN(jde float64, n int) float64 {
	return HMoonGeocentricApparentRaNWithInstance(jde, n, nil)
}

// HMoonGeocentricTrueDec 月亮地心真赤纬 / true geocentric declination of the Moon.
func HMoonGeocentricTrueDec(jde float64) float64 { return HMoonGeocentricTrueDecWithInstance(jde, nil) }

// HMoonGeocentricTrueDecN 月亮地心真赤纬（n 变体） / n-parameterised true geocentric declination of the Moon.
func HMoonGeocentricTrueDecN(jde float64, n int) float64 {
	return HMoonGeocentricTrueDecNWithInstance(jde, n, nil)
}

// HMoonGeocentricTrueRa 月亮地心真赤经 / true geocentric right ascension of the Moon.
func HMoonGeocentricTrueRa(jde float64) float64 { return HMoonGeocentricTrueRaWithInstance(jde, nil) }

// HMoonGeocentricTrueRaDec 月亮地心真赤经、真赤纬 / true geocentric right ascension and declination of the Moon.
func HMoonGeocentricTrueRaDec(jde float64) (float64, float64) {
	return HMoonGeocentricTrueRaDecWithInstance(jde, nil)
}

// HMoonGeocentricTrueRaDecN 月亮地心真赤经、真赤纬（n 变体） / n-parameterised true geocentric right ascension and declination of the Moon.
func HMoonGeocentricTrueRaDecN(jde float64, n int) (float64, float64) {
	return HMoonGeocentricTrueRaDecNWithInstance(jde, n, nil)
}

// HMoonGeocentricTrueRaN 月亮地心真赤经（n 变体） / n-parameterised true geocentric right ascension of the Moon.
func HMoonGeocentricTrueRaN(jde float64, n int) float64 {
	return HMoonGeocentricTrueRaNWithInstance(jde, n, nil)
}

// HMoonHeight 月心几何高度角（度，不含折射），jd 为 tz 时区的当地民用儒略日 / geometric Moon-centre altitude for a local civil Julian day in zone tz.
func HMoonHeight(jd, lon, lat, tz float64) float64 {
	return HMoonHeightWithInstance(jd, lon, lat, tz, nil)
}

// JupiterAscendingNodeN 木星升交点黄经（n 变体） / n-parameterised ascending node longitude of Jupiter.
func JupiterAscendingNodeN(jd float64, n int) float64 {
	return JupiterAscendingNodeNWithInstance(jd, n, nil)
}

// JupiterAzimuthN 木星方位角（n 变体） / n-parameterised azimuth of Jupiter.
func JupiterAzimuthN(localJD, lon, lat, timezone float64, n int) float64 {
	return JupiterAzimuthNWithInstance(localJD, lon, lat, timezone, n, nil)
}

// JupiterBrightLimbPositionAngleN 木星亮面中心位置角（n 变体） / n-parameterised position angle of Jupiter bright limb.
func JupiterBrightLimbPositionAngleN(jde float64, n int) float64 {
	return JupiterBrightLimbPositionAngleNWithInstance(jde, n, nil)
}

// JupiterCentralMeridiansN 木星 System I/II/III 中央经线（n 变体） / n-parameterised Jupiter System I/II/III central meridians.
func JupiterCentralMeridiansN(jde float64, n int) JupiterCentralMeridianInfo {
	return JupiterCentralMeridiansNWithInstance(jde, n, nil)
}

// JupiterCulminationTimeN 木星中天时间（n 变体） / n-parameterised culmination time of Jupiter.
func JupiterCulminationTimeN(localJD, lon, timezone float64, n int) float64 {
	return JupiterCulminationTimeNWithInstance(localJD, lon, timezone, n, nil)
}

// JupiterDSDEN 木星 DS/DE 行星中心赤纬（n 变体） / n-parameterised Jupiter planetocentric declinations of Sun and Earth.
func JupiterDSDEN(jde float64, n int) (ds, de float64) { return JupiterDSDENWithInstance(jde, n, nil) }

// JupiterDescendingNodeN 木星降交点黄经（n 变体） / n-parameterised descending node longitude of Jupiter.
func JupiterDescendingNodeN(jd float64, n int) float64 {
	return JupiterDescendingNodeNWithInstance(jd, n, nil)
}

// JupiterDiameterN 木星视直径（n 变体），单位角秒 / n-parameterised apparent Jupiter diameter in arcseconds.
func JupiterDiameterN(jde float64, n int) float64 { return JupiterDiameterNWithInstance(jde, n, nil) }

// JupiterGalileanSatelliteObservations 四颗伽利略卫星视位置，次序为 Io、Europa、Ganymede、Callisto / apparent geometry of Io, Europa, Ganymede and Callisto in that order.
func JupiterGalileanSatelliteObservations(jde float64) [4]JupiterGalileanObservation {
	return JupiterGalileanSatelliteObservationsWithInstance(jde, nil)
}

// JupiterGalileanSatellitePhenomena 四颗伽利略卫星瞬时现象 / instantaneous phenomena of the four Galilean satellites.
func JupiterGalileanSatellitePhenomena(jde float64) [4]JupiterGalileanPhenomenon {
	return JupiterGalileanSatellitePhenomenaWithInstance(jde, nil)
}

// JupiterHeightN 木星高度角（n 变体） / n-parameterised altitude of Jupiter.
func JupiterHeightN(localJD, lon, lat, timezone float64, n int) float64 {
	return JupiterHeightNWithInstance(localJD, lon, lat, timezone, n, nil)
}

// JupiterHourAngleN 木星时角（n 变体） / n-parameterised hour angle of Jupiter.
func JupiterHourAngleN(jd, lon, timezone float64, n int) float64 {
	return JupiterHourAngleNWithInstance(jd, lon, timezone, n, nil)
}

// JupiterIlluminatedFractionN 木星被照亮比例（n 变体） / n-parameterised illuminated fraction of Jupiter.
func JupiterIlluminatedFractionN(jd float64, n int) float64 {
	return JupiterIlluminatedFractionNWithInstance(jd, n, nil)
}

// JupiterMagN 木星视星等（n 变体） / n-parameterised apparent magnitude of Jupiter.
func JupiterMagN(jde float64, n int) float64 { return JupiterMagNWithInstance(jde, n, nil) }

// JupiterPhaseAngleN 木星相位角（n 变体） / n-parameterised phase angle of Jupiter.
func JupiterPhaseAngleN(jd float64, n int) float64 { return JupiterPhaseAngleNWithInstance(jd, n, nil) }

// JupiterPhysicalN 木星物理观测参数（n 变体） / n-parameterised physical observing parameters of Jupiter.
func JupiterPhysicalN(jd float64, n int) PlanetPhysicalInfo {
	return JupiterPhysicalNWithInstance(jd, n, nil)
}

// JupiterRiseTimeN 木星升起时间（n 变体） / n-parameterised rise time of Jupiter.
func JupiterRiseTimeN(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int) (float64, error) {
	return JupiterRiseTimeNWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight, n, nil)
}

// JupiterSemidiameterN 木星视半径（n 变体），单位角秒 / n-parameterised apparent Jupiter semidiameter in arcseconds.
func JupiterSemidiameterN(jde float64, n int) float64 {
	return JupiterSemidiameterNWithInstance(jde, n, nil)
}

// JupiterSetTimeN 木星落下时间（n 变体） / n-parameterised set time of Jupiter.
func JupiterSetTimeN(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int) (float64, error) {
	return JupiterSetTimeNWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight, n, nil)
}

// LastJupiterGalileanPhenomenonContactEvent 上一次 IMCCE 风格接触事件 / previous IMCCE-style contact event.
func LastJupiterGalileanPhenomenonContactEvent(jd float64, satellite int, phenomenonType JupiterGalileanPhenomenonType) JupiterGalileanPhenomenonContactEvent {
	return LastJupiterGalileanPhenomenonContactEventWithInstance(jd, satellite, phenomenonType, nil)
}

// LastJupiterGalileanPhenomenonEvent 上一次伽利略卫星现象，jd 与返回时刻均为民用儒略日 / previous Galilean-satellite event in civil Julian days.
func LastJupiterGalileanPhenomenonEvent(jd float64, satellite int, phenomenonType JupiterGalileanPhenomenonType) JupiterGalileanPhenomenonEvent {
	return LastJupiterGalileanPhenomenonEventWithInstance(jd, satellite, phenomenonType, nil)
}

// LastMercuryTransit 上一次全球水星凌日 / previous global transit of Mercury.
func LastMercuryTransit(jde float64) PlanetTransitResult {
	return LastMercuryTransitWithInstance(jde, nil)
}

// LastMoonMaximumNorthDeclination 指定时刻之前最近一次月球最大北赤纬 / last maximum northern lunar declination at or before jd.
func LastMoonMaximumNorthDeclination(jd float64) DeclinationEvent {
	return LastMoonMaximumNorthDeclinationWithInstance(jd, nil)
}

// LastMoonMaximumSouthDeclination 指定时刻之前最近一次月球最大南赤纬 / last maximum southern lunar declination at or before jd.
func LastMoonMaximumSouthDeclination(jd float64) DeclinationEvent {
	return LastMoonMaximumSouthDeclinationWithInstance(jd, nil)
}

// LastMoonPlanetConjunction 指定时刻之前最近一次行星合月（赤经合） / previous Moon-planet conjunction at or before jde.
func LastMoonPlanetConjunction(jde float64, planet MoonPlanetConjunctionPlanet) float64 {
	return LastMoonPlanetConjunctionWithInstance(jde, planet, nil)
}

// LastVenusTransit 上一次全球金星凌日 / previous global transit of Venus.
func LastVenusTransit(jde float64) PlanetTransitResult { return LastVenusTransitWithInstance(jde, nil) }

// LocalSolarEclipseDiagramIAUSingleK 站心日食视圆图几何（IAU Single-K） / local solar eclipse disk diagram, IAU Single-K.
func LocalSolarEclipseDiagramIAUSingleK(seedJDE, lon, lat, height float64, options LocalSolarEclipseDiagramOptions) LocalSolarEclipseDiagramResult {
	return LocalSolarEclipseDiagramIAUSingleKWithInstance(seedJDE, lon, lat, height, options, nil)
}

// LocalSolarEclipseDiagramNASABulletinSplitK 站心日食视圆图几何（NASA bulletin Split-K） / local solar eclipse disk diagram, NASA bulletin Split-K.
func LocalSolarEclipseDiagramNASABulletinSplitK(seedJDE, lon, lat, height float64, options LocalSolarEclipseDiagramOptions) LocalSolarEclipseDiagramResult {
	return LocalSolarEclipseDiagramNASABulletinSplitKWithInstance(seedJDE, lon, lat, height, options, nil)
}

// LocalSolarEclipseWithOptions 计算给定近朔时刻附近的一次站心日食，半径口径由 options 指定 / computes one local solar eclipse with the given radius conventions.
func LocalSolarEclipseWithOptions(seedJDE, lon, lat, height float64, options SolarEclipseOptions) LocalSolarEclipseResult {
	return LocalSolarEclipseWithOptionsWithInstance(seedJDE, lon, lat, height, options, nil)
}

// MarsAscendingNodeN 火星升交点黄经（n 变体） / n-parameterised ascending node longitude of Mars.
func MarsAscendingNodeN(jd float64, n int) float64 { return MarsAscendingNodeNWithInstance(jd, n, nil) }

// MarsAzimuthN 火星方位角（n 变体） / n-parameterised azimuth of Mars.
func MarsAzimuthN(localJD, lon, lat, timezone float64, n int) float64 {
	return MarsAzimuthNWithInstance(localJD, lon, lat, timezone, n, nil)
}

// MarsBrightLimbPositionAngleN 火星亮面中心位置角（n 变体） / n-parameterised position angle of Mars bright limb.
func MarsBrightLimbPositionAngleN(jde float64, n int) float64 {
	return MarsBrightLimbPositionAngleNWithInstance(jde, n, nil)
}

// MarsCulminationTimeN 火星中天时间（n 变体） / n-parameterised culmination time of Mars.
func MarsCulminationTimeN(localJD, lon, timezone float64, n int) float64 {
	return MarsCulminationTimeNWithInstance(localJD, lon, timezone, n, nil)
}

// MarsDescendingNodeN 火星降交点黄经（n 变体） / n-parameterised descending node longitude of Mars.
func MarsDescendingNodeN(jd float64, n int) float64 {
	return MarsDescendingNodeNWithInstance(jd, n, nil)
}

// MarsDiameterN 火星视直径（n 变体），单位角秒 / n-parameterised apparent Mars diameter in arcseconds.
func MarsDiameterN(jde float64, n int) float64 { return MarsDiameterNWithInstance(jde, n, nil) }

// MarsHeightN 火星高度角（n 变体） / n-parameterised altitude of Mars.
func MarsHeightN(localJD, lon, lat, timezone float64, n int) float64 {
	return MarsHeightNWithInstance(localJD, lon, lat, timezone, n, nil)
}

// MarsHourAngleN 火星时角（n 变体） / n-parameterised hour angle of Mars.
func MarsHourAngleN(jd, lon, timezone float64, n int) float64 {
	return MarsHourAngleNWithInstance(jd, lon, timezone, n, nil)
}

// MarsIlluminatedFractionN 火星被照亮比例（n 变体） / n-parameterised illuminated fraction of Mars.
func MarsIlluminatedFractionN(jd float64, n int) float64 {
	return MarsIlluminatedFractionNWithInstance(jd, n, nil)
}

// MarsMagN 火星视星等（n 变体） / n-parameterised apparent magnitude of Mars.
func MarsMagN(jde float64, n int) float64 { return MarsMagNWithInstance(jde, n, nil) }

// MarsPhaseAngleN 火星相位角（n 变体） / n-parameterised phase angle of Mars.
func MarsPhaseAngleN(jd float64, n int) float64 { return MarsPhaseAngleNWithInstance(jd, n, nil) }

// MarsPhysicalN 火星物理观测参数（n 变体） / n-parameterised physical observing parameters of Mars.
func MarsPhysicalN(jd float64, n int) PlanetPhysicalInfo {
	return MarsPhysicalNWithInstance(jd, n, nil)
}

// MarsRiseTimeN 火星升起时间（n 变体） / n-parameterised rise time of Mars.
func MarsRiseTimeN(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int) (float64, error) {
	return MarsRiseTimeNWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight, n, nil)
}

// MarsSemidiameterN 火星视半径（n 变体），单位角秒 / n-parameterised apparent Mars semidiameter in arcseconds.
func MarsSemidiameterN(jde float64, n int) float64 { return MarsSemidiameterNWithInstance(jde, n, nil) }

// MarsSetTimeN 火星落下时间（n 变体） / n-parameterised set time of Mars.
func MarsSetTimeN(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int) (float64, error) {
	return MarsSetTimeNWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight, n, nil)
}

// MercuryAscendingNodeN 水星升交点黄经（n 变体） / n-parameterised ascending node longitude of Mercury.
func MercuryAscendingNodeN(jd float64, n int) float64 {
	return MercuryAscendingNodeNWithInstance(jd, n, nil)
}

// MercuryAzimuthN 水星方位角（n 变体） / n-parameterised azimuth of Mercury.
func MercuryAzimuthN(localJD, lon, lat, timezone float64, n int) float64 {
	return MercuryAzimuthNWithInstance(localJD, lon, lat, timezone, n, nil)
}

// MercuryBrightLimbPositionAngleN 水星亮面中心位置角（n 变体） / n-parameterised position angle of Mercury bright limb.
func MercuryBrightLimbPositionAngleN(jde float64, n int) float64 {
	return MercuryBrightLimbPositionAngleNWithInstance(jde, n, nil)
}

// MercuryCulminationTimeN 水星中天时间（n 变体） / n-parameterised culmination time of Mercury.
func MercuryCulminationTimeN(localJD, lon, timezone float64, n int) float64 {
	return MercuryCulminationTimeNWithInstance(localJD, lon, timezone, n, nil)
}

// MercuryDescendingNodeN 水星降交点黄经（n 变体） / n-parameterised descending node longitude of Mercury.
func MercuryDescendingNodeN(jd float64, n int) float64 {
	return MercuryDescendingNodeNWithInstance(jd, n, nil)
}

// MercuryDiameterN 水星视直径（n 变体），单位角秒 / n-parameterised apparent Mercury diameter in arcseconds.
func MercuryDiameterN(jde float64, n int) float64 { return MercuryDiameterNWithInstance(jde, n, nil) }

// MercuryHeightN 水星高度角（n 变体） / n-parameterised altitude of Mercury.
func MercuryHeightN(localJD, lon, lat, timezone float64, n int) float64 {
	return MercuryHeightNWithInstance(localJD, lon, lat, timezone, n, nil)
}

// MercuryHourAngleN 水星时角（n 变体） / n-parameterised hour angle of Mercury.
func MercuryHourAngleN(jd, lon, timezone float64, n int) float64 {
	return MercuryHourAngleNWithInstance(jd, lon, timezone, n, nil)
}

// MercuryIlluminatedFractionN 水星被照亮比例（n 变体） / n-parameterised illuminated fraction of Mercury.
func MercuryIlluminatedFractionN(jd float64, n int) float64 {
	return MercuryIlluminatedFractionNWithInstance(jd, n, nil)
}

// MercuryMagN 水星视星等（n 变体） / n-parameterised apparent magnitude of Mercury.
func MercuryMagN(jde float64, n int) float64 { return MercuryMagNWithInstance(jde, n, nil) }

// MercuryPhaseAngleN 水星相位角（n 变体） / n-parameterised phase angle of Mercury.
func MercuryPhaseAngleN(jd float64, n int) float64 { return MercuryPhaseAngleNWithInstance(jd, n, nil) }

// MercuryPhysicalN 水星物理观测参数（n 变体） / n-parameterised physical observing parameters of Mercury.
func MercuryPhysicalN(jd float64, n int) PlanetPhysicalInfo {
	return MercuryPhysicalNWithInstance(jd, n, nil)
}

// MercuryRiseTimeN 水星升起时间（n 变体） / n-parameterised rise time of Mercury.
func MercuryRiseTimeN(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int) (float64, error) {
	return MercuryRiseTimeNWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight, n, nil)
}

// MercurySemidiameterN 水星视半径（n 变体），单位角秒 / n-parameterised apparent Mercury semidiameter in arcseconds.
func MercurySemidiameterN(jde float64, n int) float64 {
	return MercurySemidiameterNWithInstance(jde, n, nil)
}

// MercurySetTimeN 水星落下时间（n 变体） / n-parameterised set time of Mercury.
func MercurySetTimeN(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int) (float64, error) {
	return MercurySetTimeNWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight, n, nil)
}

// MoonApogees 指定年月内的所有月球远地点 / all lunar apogees in the given Gregorian month.
func MoonApogees(year int, month time.Month) []ApsisEvent {
	return MoonApogeesWithInstance(year, month, nil)
}

// MoonAscendingNodeN 月球升交点黄经（n 变体） / n-parameterised ascending node longitude of the Moon.
func MoonAscendingNodeN(jd float64, n int) float64 { return MoonAscendingNodeNWithInstance(jd, n, nil) }

// MoonBrightLimbPositionAngleN 月亮明亮边缘位置角（n 变体） / n-parameterised position angle of the Moon's bright limb.
func MoonBrightLimbPositionAngleN(jde float64, n int) float64 {
	return MoonBrightLimbPositionAngleNWithInstance(jde, n, nil)
}

// MoonDescendingNodeN 月球降交点黄经（n 变体） / n-parameterised descending node longitude of the Moon.
func MoonDescendingNodeN(jd float64, n int) float64 {
	return MoonDescendingNodeNWithInstance(jd, n, nil)
}

// MoonDiameterN 月亮视直径（n 变体），单位角秒 / n-parameterised apparent lunar diameter in arcseconds.
func MoonDiameterN(jde float64, n int) float64 { return MoonDiameterNWithInstance(jde, n, nil) }

// MoonHorizon UTC 儒略日下海平面月心几何地平圈的 [经度, 纬度] 顶点，不含折射；samples<=0 取 360，其余夹到 [12, 1440] / sea-level geometric Moon-centre horizon vertices.
func MoonHorizon(jdUTC float64, samples int) [][2]float64 {
	return MoonHorizonWithInstance(jdUTC, samples, nil)
}

// MoonMaximumNorthDeclinations 指定年月内的所有月球最大北赤纬事件 / all maximum northern lunar declination events in the given Gregorian month.
func MoonMaximumNorthDeclinations(year int, month time.Month) []DeclinationEvent {
	return MoonMaximumNorthDeclinationsWithInstance(year, month, nil)
}

// MoonMaximumSouthDeclinations 指定年月内的所有月球最大南赤纬事件 / all maximum southern lunar declination events in the given Gregorian month.
func MoonMaximumSouthDeclinations(year int, month time.Month) []DeclinationEvent {
	return MoonMaximumSouthDeclinationsWithInstance(year, month, nil)
}

// MoonPerigees 指定年月内的所有月球近地点 / all lunar perigees in the given Gregorian month.
func MoonPerigees(year int, month time.Month) []ApsisEvent {
	return MoonPerigeesWithInstance(year, month, nil)
}

// MoonPhysical 月球物理观测参数 / physical observing parameters of the Moon.
func MoonPhysical(jde float64) MoonPhysicalInfo { return MoonPhysicalWithInstance(jde, nil) }

// MoonPhysicalN 月球物理观测参数（n 变体） / n-parameterised physical observing parameters of the Moon.
func MoonPhysicalN(jde float64, n int) MoonPhysicalInfo {
	return MoonPhysicalNWithInstance(jde, n, nil)
}

// MoonSemidiameter 月亮视半径，单位角秒 / apparent lunar semidiameter in arcseconds.
func MoonSemidiameter(jde float64) float64 { return MoonSemidiameterWithInstance(jde, nil) }

// MoonSemidiameterN 月亮视半径（n 变体），单位角秒 / n-parameterised apparent lunar semidiameter in arcseconds.
func MoonSemidiameterN(jde float64, n int) float64 { return MoonSemidiameterNWithInstance(jde, n, nil) }

// MoonStateAt 由 UTC 儒略日构造该瞬间的可复用月球状态 / builds the reusable state for one UTC Julian day.
func MoonStateAt(utcJD float64) MoonState { return MoonStateAtWithInstance(utcJD, nil) }

// MoonTopocentricBrightLimbPositionAngleN 月亮站心明亮边缘位置角（n 变体） / n-parameterised topocentric position angle of the Moon's bright limb.
func MoonTopocentricBrightLimbPositionAngleN(jde, observerLon, observerLat, height float64, n int) float64 {
	return MoonTopocentricBrightLimbPositionAngleNWithInstance(jde, observerLon, observerLat, height, n, nil)
}

// MoonTopocentricPhysicalN 月球站心物理观测参数（n 变体） / n-parameterised topocentric physical observing parameters of the Moon.
func MoonTopocentricPhysicalN(jde, observerLon, observerLat, height float64, n int) MoonPhysicalInfo {
	return MoonTopocentricPhysicalNWithInstance(jde, observerLon, observerLat, height, n, nil)
}

// NeptuneAscendingNodeN 海王星升交点黄经（n 变体） / n-parameterised ascending node longitude of Neptune.
func NeptuneAscendingNodeN(jd float64, n int) float64 {
	return NeptuneAscendingNodeNWithInstance(jd, n, nil)
}

// NeptuneAzimuthN 海王星方位角（n 变体） / n-parameterised azimuth of Neptune.
func NeptuneAzimuthN(localJD, lon, lat, timezone float64, n int) float64 {
	return NeptuneAzimuthNWithInstance(localJD, lon, lat, timezone, n, nil)
}

// NeptuneBrightLimbPositionAngleN 海王星亮面中心位置角（n 变体） / n-parameterised position angle of Neptune bright limb.
func NeptuneBrightLimbPositionAngleN(jde float64, n int) float64 {
	return NeptuneBrightLimbPositionAngleNWithInstance(jde, n, nil)
}

// NeptuneCulminationTimeN 海王星中天时间（n 变体） / n-parameterised culmination time of Neptune.
func NeptuneCulminationTimeN(localJD, lon, timezone float64, n int) float64 {
	return NeptuneCulminationTimeNWithInstance(localJD, lon, timezone, n, nil)
}

// NeptuneDescendingNodeN 海王星降交点黄经（n 变体） / n-parameterised descending node longitude of Neptune.
func NeptuneDescendingNodeN(jd float64, n int) float64 {
	return NeptuneDescendingNodeNWithInstance(jd, n, nil)
}

// NeptuneDiameterN 海王星视直径（n 变体），单位角秒 / n-parameterised apparent Neptune diameter in arcseconds.
func NeptuneDiameterN(jde float64, n int) float64 { return NeptuneDiameterNWithInstance(jde, n, nil) }

// NeptuneHeightN 海王星高度角（n 变体） / n-parameterised altitude of Neptune.
func NeptuneHeightN(localJD, lon, lat, timezone float64, n int) float64 {
	return NeptuneHeightNWithInstance(localJD, lon, lat, timezone, n, nil)
}

// NeptuneHourAngleN 海王星时角（n 变体） / n-parameterised hour angle of Neptune.
func NeptuneHourAngleN(jd, lon, timezone float64, n int) float64 {
	return NeptuneHourAngleNWithInstance(jd, lon, timezone, n, nil)
}

// NeptuneIlluminatedFractionN 海王星被照亮比例（n 变体） / n-parameterised illuminated fraction of Neptune.
func NeptuneIlluminatedFractionN(jd float64, n int) float64 {
	return NeptuneIlluminatedFractionNWithInstance(jd, n, nil)
}

// NeptuneMagN 海王星视星等（n 变体） / n-parameterised apparent magnitude of Neptune.
func NeptuneMagN(jde float64, n int) float64 { return NeptuneMagNWithInstance(jde, n, nil) }

// NeptunePhaseAngleN 海王星相位角（n 变体） / n-parameterised phase angle of Neptune.
func NeptunePhaseAngleN(jd float64, n int) float64 { return NeptunePhaseAngleNWithInstance(jd, n, nil) }

// NeptunePhysicalN 海王星物理观测参数（n 变体） / n-parameterised physical observing parameters of Neptune.
func NeptunePhysicalN(jd float64, n int) PlanetPhysicalInfo {
	return NeptunePhysicalNWithInstance(jd, n, nil)
}

// NeptuneRiseTimeN 海王星升起时间（n 变体） / n-parameterised rise time of Neptune.
func NeptuneRiseTimeN(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int) (float64, error) {
	return NeptuneRiseTimeNWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight, n, nil)
}

// NeptuneSemidiameterN 海王星视半径（n 变体），单位角秒 / n-parameterised apparent Neptune semidiameter in arcseconds.
func NeptuneSemidiameterN(jde float64, n int) float64 {
	return NeptuneSemidiameterNWithInstance(jde, n, nil)
}

// NeptuneSetTimeN 海王星落下时间（n 变体） / n-parameterised set time of Neptune.
func NeptuneSetTimeN(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int) (float64, error) {
	return NeptuneSetTimeNWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight, n, nil)
}

// NewSolarEclipseShadowSolver 构造单时刻求解器 / builds a single-instant solver.
func NewSolarEclipseShadowSolver(options SolarEclipseShadowSolverOptions) *SolarEclipseShadowSolver {
	return NewSolarEclipseShadowSolverWithInstance(options, nil)
}

// NextJupiterGalileanPhenomenonContactEvent 下一次 IMCCE 风格接触事件 / next IMCCE-style contact event.
func NextJupiterGalileanPhenomenonContactEvent(jd float64, satellite int, phenomenonType JupiterGalileanPhenomenonType) JupiterGalileanPhenomenonContactEvent {
	return NextJupiterGalileanPhenomenonContactEventWithInstance(jd, satellite, phenomenonType, nil)
}

// NextJupiterGalileanPhenomenonEvent 下一次伽利略卫星现象，jd 与返回时刻均为民用儒略日 / next Galilean-satellite event in civil Julian days.
func NextJupiterGalileanPhenomenonEvent(jd float64, satellite int, phenomenonType JupiterGalileanPhenomenonType) JupiterGalileanPhenomenonEvent {
	return NextJupiterGalileanPhenomenonEventWithInstance(jd, satellite, phenomenonType, nil)
}

// NextMercuryTransit 下一次全球水星凌日 / next global transit of Mercury.
func NextMercuryTransit(jde float64) PlanetTransitResult {
	return NextMercuryTransitWithInstance(jde, nil)
}

// NextMoonMaximumNorthDeclination 指定时刻之后最近一次月球最大北赤纬 / next maximum northern lunar declination after jd.
func NextMoonMaximumNorthDeclination(jd float64) DeclinationEvent {
	return NextMoonMaximumNorthDeclinationWithInstance(jd, nil)
}

// NextMoonMaximumSouthDeclination 指定时刻之后最近一次月球最大南赤纬 / next maximum southern lunar declination after jd.
func NextMoonMaximumSouthDeclination(jd float64) DeclinationEvent {
	return NextMoonMaximumSouthDeclinationWithInstance(jd, nil)
}

// NextMoonPlanetConjunction 指定时刻之后最近一次行星合月（赤经合） / next Moon-planet conjunction at or after jde.
func NextMoonPlanetConjunction(jde float64, planet MoonPlanetConjunctionPlanet) float64 {
	return NextMoonPlanetConjunctionWithInstance(jde, planet, nil)
}

// NextVenusTransit 下一次全球金星凌日 / next global transit of Venus.
func NextVenusTransit(jde float64) PlanetTransitResult { return NextVenusTransitWithInstance(jde, nil) }

// PlanetOccultationDiagram 已求解固定地点行星月掩的月心切平面轨迹，事件无效时不含帧 / tangent-plane track of a solved fixed-site planetary occultation.
func PlanetOccultationDiagram(
	info PlanetOccultationInfo,
	options PlanetOccultationDiagramOptions,
) PlanetOccultationDiagramResult {
	return PlanetOccultationDiagramWithInstance(info, options, nil)
}

// PlanetOccultationFootprintsAt 指定时刻行星月掩外、内接触可见足迹 / visible outer- and inner-contact planetary footprints at one instant.
func PlanetOccultationFootprintsAt(at time.Time, planet OccultationPlanet) (PlanetOccultationInstant, error) {
	return PlanetOccultationFootprintsAtWithInstance(at, planet, nil)
}

// PlanetTransitVisibilityColumns 按 2° 经度、1° 纬度探测站心可见区并细化边界 / sample topocentric visibility at 2-degree longitude and 1-degree latitude spacing and refine boundaries.
func PlanetTransitVisibilityColumns(transit PlanetTransitResult) []PlanetTransitVisibilityColumn {
	return PlanetTransitVisibilityColumnsWithInstance(transit, nil)
}

// PlanetTransitVisibilityPolygons 可见区的经度条带多边形 / longitude-strip polygons of the visible region.
func PlanetTransitVisibilityPolygons(transit PlanetTransitResult) [][]PlanetTransitIsochronePoint {
	return PlanetTransitVisibilityPolygonsWithInstance(transit, nil)
}

// SaturnAscendingNodeN 土星升交点黄经（n 变体） / n-parameterised ascending node longitude of Saturn.
func SaturnAscendingNodeN(jd float64, n int) float64 {
	return SaturnAscendingNodeNWithInstance(jd, n, nil)
}

// SaturnAzimuthN 土星方位角（n 变体） / n-parameterised azimuth of Saturn.
func SaturnAzimuthN(localJD, lon, lat, timezone float64, n int) float64 {
	return SaturnAzimuthNWithInstance(localJD, lon, lat, timezone, n, nil)
}

// SaturnBrightLimbPositionAngleN 土星亮面中心位置角（n 变体） / n-parameterised position angle of Saturn bright limb.
func SaturnBrightLimbPositionAngleN(jde float64, n int) float64 {
	return SaturnBrightLimbPositionAngleNWithInstance(jde, n, nil)
}

// SaturnCulminationTimeN 土星中天时间（n 变体） / n-parameterised culmination time of Saturn.
func SaturnCulminationTimeN(localJD, lon, timezone float64, n int) float64 {
	return SaturnCulminationTimeNWithInstance(localJD, lon, timezone, n, nil)
}

// SaturnDescendingNodeN 土星降交点黄经（n 变体） / n-parameterised descending node longitude of Saturn.
func SaturnDescendingNodeN(jd float64, n int) float64 {
	return SaturnDescendingNodeNWithInstance(jd, n, nil)
}

// SaturnDiameterN 土星视直径（n 变体），单位角秒 / n-parameterised apparent Saturn diameter in arcseconds.
func SaturnDiameterN(jde float64, n int) float64 { return SaturnDiameterNWithInstance(jde, n, nil) }

// SaturnHeightN 土星高度角（n 变体） / n-parameterised altitude of Saturn.
func SaturnHeightN(localJD, lon, lat, timezone float64, n int) float64 {
	return SaturnHeightNWithInstance(localJD, lon, lat, timezone, n, nil)
}

// SaturnHourAngleN 土星时角（n 变体） / n-parameterised hour angle of Saturn.
func SaturnHourAngleN(jd, lon, timezone float64, n int) float64 {
	return SaturnHourAngleNWithInstance(jd, lon, timezone, n, nil)
}

// SaturnIlluminatedFractionN 土星被照亮比例（n 变体） / n-parameterised illuminated fraction of Saturn.
func SaturnIlluminatedFractionN(jd float64, n int) float64 {
	return SaturnIlluminatedFractionNWithInstance(jd, n, nil)
}

// SaturnMagN 土星视星等（n 变体） / n-parameterised apparent magnitude of Saturn.
func SaturnMagN(jde float64, n int) float64 { return SaturnMagNWithInstance(jde, n, nil) }

// SaturnPhaseAngleN 土星相位角（n 变体） / n-parameterised phase angle of Saturn.
func SaturnPhaseAngleN(jd float64, n int) float64 { return SaturnPhaseAngleNWithInstance(jd, n, nil) }

// SaturnPhysicalN 土星物理观测参数（n 变体） / n-parameterised physical observing parameters of Saturn.
func SaturnPhysicalN(jd float64, n int) PlanetPhysicalInfo {
	return SaturnPhysicalNWithInstance(jd, n, nil)
}

// SaturnRiseTimeN 土星升起时间（n 变体） / n-parameterised rise time of Saturn.
func SaturnRiseTimeN(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int) (float64, error) {
	return SaturnRiseTimeNWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight, n, nil)
}

// SaturnSemidiameterN 土星视半径（n 变体），单位角秒 / n-parameterised apparent Saturn semidiameter in arcseconds.
func SaturnSemidiameterN(jde float64, n int) float64 {
	return SaturnSemidiameterNWithInstance(jde, n, nil)
}

// SaturnSetTimeN 土星落下时间（n 变体） / n-parameterised set time of Saturn.
func SaturnSetTimeN(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int) (float64, error) {
	return SaturnSetTimeNWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight, n, nil)
}

// SolarEclipseBesselianElements 近朔时刻附近日食的多项式贝塞尔根数，无日食时返回 false / polynomial Besselian elements of the solar eclipse near a new moon.
func SolarEclipseBesselianElements(
	seedJDE float64, options SolarEclipseBesselianElementsOptions,
) (SolarEclipseBesselianElementsResult, bool) {
	return SolarEclipseBesselianElementsWithInstance(seedJDE, options, nil)
}

// SolarEclipseCentralPathIAUSingleK 日食中心路径（IAU Single-K） / solar eclipse central path, IAU Single-K.
func SolarEclipseCentralPathIAUSingleK(seedJDE float64, options SolarEclipsePathOptions) SolarEclipsePathResult {
	return SolarEclipseCentralPathIAUSingleKWithInstance(seedJDE, options, nil)
}

// SolarEclipseCentralPathNASABulletinSplitK 日食中心路径（NASA bulletin Split-K） / solar eclipse central path, NASA bulletin Split-K.
func SolarEclipseCentralPathNASABulletinSplitK(seedJDE float64, options SolarEclipsePathOptions) SolarEclipsePathResult {
	return SolarEclipseCentralPathNASABulletinSplitKWithInstance(seedJDE, options, nil)
}

// SolarEclipseIAUSingleK 近朔时刻附近的全局日食（IAU Single-K） / global solar eclipse near a new moon, IAU Single-K.
func SolarEclipseIAUSingleK(seedJDE float64) SolarEclipseResult {
	return SolarEclipseIAUSingleKWithInstance(seedJDE, nil)
}

// SolarEclipseNASABulletinSplitK 近朔时刻附近的全局日食（NASA bulletin Split-K） / global solar eclipse near a new moon, NASA bulletin Split-K.
func SolarEclipseNASABulletinSplitK(seedJDE float64) SolarEclipseResult {
	return SolarEclipseNASABulletinSplitKWithInstance(seedJDE, nil)
}

// SolarEclipsePartialFootprintsIAUSingleK 日食半影足迹序列（IAU Single-K） / penumbral footprint samples, IAU Single-K.
func SolarEclipsePartialFootprintsIAUSingleK(seedJDE float64, options SolarEclipsePartialFootprintOptions) SolarEclipsePartialFootprintsResult {
	return SolarEclipsePartialFootprintsIAUSingleKWithInstance(seedJDE, options, nil)
}

// SolarEclipsePartialFootprintsNASABulletinSplitK 日食半影足迹序列（NASA bulletin Split-K） / penumbral footprint samples, NASA bulletin Split-K.
func SolarEclipsePartialFootprintsNASABulletinSplitK(seedJDE float64, options SolarEclipsePartialFootprintOptions) SolarEclipsePartialFootprintsResult {
	return SolarEclipsePartialFootprintsNASABulletinSplitKWithInstance(seedJDE, options, nil)
}

// SolarEclipseSunSemidiameter 指定太阳半径口径下的视半径，单位角秒 / apparent solar semidiameter in arcseconds under a given eclipse sun radius convention.
func SolarEclipseSunSemidiameter(jde float64, model SolarEclipseSunRadiusModel) float64 {
	return SolarEclipseSunSemidiameterWithInstance(jde, model, nil)
}

// SolarEclipseWithOptions 计算给定近朔时刻附近的一次全局日食，半径口径由 options 指定 / computes one global solar eclipse with the given radius conventions.
func SolarEclipseWithOptions(seedJDE float64, options SolarEclipseOptions) SolarEclipseResult {
	return SolarEclipseWithOptionsWithInstance(seedJDE, options, nil)
}

// StarOccultationLimbProfile 沿星迹按月缘剖面精密求解一次恒星月掩 / precise stellar occultation along the track with a lunar limb profile.
func StarOccultationLimbProfile(info StarOccultationInfo, star StarCoordinate, limb LunarLimb, options StarOccultationLimbProfileOptions) (StarOccultationLimbProfileResult, error) {
	return StarOccultationLimbProfileWithInstance(info, star, limb, options, nil)
}

// StarOccultationDiagram 已求解固定地点恒星月掩的月心切平面轨迹，事件无效时不含帧 / tangent-plane track of a solved fixed-site stellar occultation.
func StarOccultationDiagram(
	info StarOccultationInfo,
	star StarCoordinate,
	options StarOccultationDiagramOptions,
) StarOccultationDiagramResult {
	return StarOccultationDiagramWithInstance(info, star, options, nil)
}

// StarOccultationFootprintAt 指定时刻点源恒星月掩可见足迹 / visible stellar-occultation footprint at one instant.
func StarOccultationFootprintAt(at time.Time, star StarCoordinate) (StarOccultationInstant, error) {
	return StarOccultationFootprintAtWithInstance(at, star, nil)
}

// SunDiameterN 太阳视直径（n 变体），单位角秒 / n-parameterised apparent solar diameter in arcseconds.
func SunDiameterN(jde float64, n int) float64 { return SunDiameterNWithInstance(jde, n, nil) }

// SunPhysicalN 太阳物理观测参数（n 变体） / n-parameterised physical observing parameters of the Sun.
func SunPhysicalN(jde float64, n int) SunPhysicalInfo { return SunPhysicalNWithInstance(jde, n, nil) }

// SunSemidiameterN 太阳视半径（n 变体），单位角秒 / n-parameterised apparent solar semidiameter in arcseconds.
func SunSemidiameterN(jde float64, n int) float64 { return SunSemidiameterNWithInstance(jde, n, nil) }

// UranusAscendingNodeN 天王星升交点黄经（n 变体） / n-parameterised ascending node longitude of Uranus.
func UranusAscendingNodeN(jd float64, n int) float64 {
	return UranusAscendingNodeNWithInstance(jd, n, nil)
}

// UranusAzimuthN 天王星方位角（n 变体） / n-parameterised azimuth of Uranus.
func UranusAzimuthN(localJD, lon, lat, timezone float64, n int) float64 {
	return UranusAzimuthNWithInstance(localJD, lon, lat, timezone, n, nil)
}

// UranusBrightLimbPositionAngleN 天王星亮面中心位置角（n 变体） / n-parameterised position angle of Uranus bright limb.
func UranusBrightLimbPositionAngleN(jde float64, n int) float64 {
	return UranusBrightLimbPositionAngleNWithInstance(jde, n, nil)
}

// UranusCulminationTimeN 天王星中天时间（n 变体） / n-parameterised culmination time of Uranus.
func UranusCulminationTimeN(localJD, lon, timezone float64, n int) float64 {
	return UranusCulminationTimeNWithInstance(localJD, lon, timezone, n, nil)
}

// UranusDescendingNodeN 天王星降交点黄经（n 变体） / n-parameterised descending node longitude of Uranus.
func UranusDescendingNodeN(jd float64, n int) float64 {
	return UranusDescendingNodeNWithInstance(jd, n, nil)
}

// UranusDiameterN 天王星视直径（n 变体），单位角秒 / n-parameterised apparent Uranus diameter in arcseconds.
func UranusDiameterN(jde float64, n int) float64 { return UranusDiameterNWithInstance(jde, n, nil) }

// UranusHeightN 天王星高度角（n 变体） / n-parameterised altitude of Uranus.
func UranusHeightN(localJD, lon, lat, timezone float64, n int) float64 {
	return UranusHeightNWithInstance(localJD, lon, lat, timezone, n, nil)
}

// UranusHourAngleN 天王星时角（n 变体） / n-parameterised hour angle of Uranus.
func UranusHourAngleN(jd, lon, timezone float64, n int) float64 {
	return UranusHourAngleNWithInstance(jd, lon, timezone, n, nil)
}

// UranusIlluminatedFractionN 天王星被照亮比例（n 变体） / n-parameterised illuminated fraction of Uranus.
func UranusIlluminatedFractionN(jd float64, n int) float64 {
	return UranusIlluminatedFractionNWithInstance(jd, n, nil)
}

// UranusMagN 天王星视星等（n 变体） / n-parameterised apparent magnitude of Uranus.
func UranusMagN(jde float64, n int) float64 { return UranusMagNWithInstance(jde, n, nil) }

// UranusPhaseAngleN 天王星相位角（n 变体） / n-parameterised phase angle of Uranus.
func UranusPhaseAngleN(jd float64, n int) float64 { return UranusPhaseAngleNWithInstance(jd, n, nil) }

// UranusPhysicalN 天王星物理观测参数（n 变体） / n-parameterised physical observing parameters of Uranus.
func UranusPhysicalN(jd float64, n int) PlanetPhysicalInfo {
	return UranusPhysicalNWithInstance(jd, n, nil)
}

// UranusRiseTimeN 天王星升起时间（n 变体） / n-parameterised rise time of Uranus.
func UranusRiseTimeN(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int) (float64, error) {
	return UranusRiseTimeNWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight, n, nil)
}

// UranusSemidiameterN 天王星视半径（n 变体），单位角秒 / n-parameterised apparent Uranus semidiameter in arcseconds.
func UranusSemidiameterN(jde float64, n int) float64 {
	return UranusSemidiameterNWithInstance(jde, n, nil)
}

// UranusSetTimeN 天王星落下时间（n 变体） / n-parameterised set time of Uranus.
func UranusSetTimeN(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int) (float64, error) {
	return UranusSetTimeNWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight, n, nil)
}

// VenusAscendingNodeN 金星升交点黄经（n 变体） / n-parameterised ascending node longitude of Venus.
func VenusAscendingNodeN(jd float64, n int) float64 {
	return VenusAscendingNodeNWithInstance(jd, n, nil)
}

// VenusAzimuthN 金星方位角（n 变体） / n-parameterised azimuth of Venus.
func VenusAzimuthN(localJD, lon, lat, timezone float64, n int) float64 {
	return VenusAzimuthNWithInstance(localJD, lon, lat, timezone, n, nil)
}

// VenusBrightLimbPositionAngleN 金星亮面中心位置角（n 变体） / n-parameterised position angle of Venus bright limb.
func VenusBrightLimbPositionAngleN(jde float64, n int) float64 {
	return VenusBrightLimbPositionAngleNWithInstance(jde, n, nil)
}

// VenusCulminationTimeN 金星中天时间（n 变体） / n-parameterised culmination time of Venus.
func VenusCulminationTimeN(localJD, lon, timezone float64, n int) float64 {
	return VenusCulminationTimeNWithInstance(localJD, lon, timezone, n, nil)
}

// VenusDescendingNodeN 金星降交点黄经（n 变体） / n-parameterised descending node longitude of Venus.
func VenusDescendingNodeN(jd float64, n int) float64 {
	return VenusDescendingNodeNWithInstance(jd, n, nil)
}

// VenusDiameterN 金星视直径（n 变体），单位角秒 / n-parameterised apparent Venus diameter in arcseconds.
func VenusDiameterN(jde float64, n int) float64 { return VenusDiameterNWithInstance(jde, n, nil) }

// VenusHeightN 金星高度角（n 变体） / n-parameterised altitude of Venus.
func VenusHeightN(localJD, lon, lat, timezone float64, n int) float64 {
	return VenusHeightNWithInstance(localJD, lon, lat, timezone, n, nil)
}

// VenusHourAngleN 金星时角（n 变体） / n-parameterised hour angle of Venus.
func VenusHourAngleN(jd, lon, timezone float64, n int) float64 {
	return VenusHourAngleNWithInstance(jd, lon, timezone, n, nil)
}

// VenusIlluminatedFractionN 金星被照亮比例（n 变体） / n-parameterised illuminated fraction of Venus.
func VenusIlluminatedFractionN(jd float64, n int) float64 {
	return VenusIlluminatedFractionNWithInstance(jd, n, nil)
}

// VenusMagN 金星视星等（n 变体） / n-parameterised apparent magnitude of Venus.
func VenusMagN(jde float64, n int) float64 { return VenusMagNWithInstance(jde, n, nil) }

// VenusPhaseAngleN 金星相位角（n 变体） / n-parameterised phase angle of Venus.
func VenusPhaseAngleN(jd float64, n int) float64 { return VenusPhaseAngleNWithInstance(jd, n, nil) }

// VenusPhysicalN 金星物理观测参数（n 变体） / n-parameterised physical observing parameters of Venus.
func VenusPhysicalN(jd float64, n int) PlanetPhysicalInfo {
	return VenusPhysicalNWithInstance(jd, n, nil)
}

// VenusRiseTimeN 金星升起时间（n 变体） / n-parameterised rise time of Venus.
func VenusRiseTimeN(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int) (float64, error) {
	return VenusRiseTimeNWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight, n, nil)
}

// VenusSemidiameterN 金星视半径（n 变体），单位角秒 / n-parameterised apparent Venus semidiameter in arcseconds.
func VenusSemidiameterN(jde float64, n int) float64 {
	return VenusSemidiameterNWithInstance(jde, n, nil)
}

// VenusSetTimeN 金星落下时间（n 变体） / n-parameterised set time of Venus.
func VenusSetTimeN(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int) (float64, error) {
	return VenusSetTimeNWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight, n, nil)
}

// NextMercuryGreatestElongation 下一次水星大距（东西任一侧） / next greatest elongation of Mercury on either side.
func NextMercuryGreatestElongation(jde float64) float64 {
	return NextMercuryGreatestElongationWithInstance(jde, nil)
}

// NextVenusGreatestElongation 下一次金星大距（东西任一侧） / next greatest elongation of Venus on either side.
func NextVenusGreatestElongation(jde float64) float64 {
	return NextVenusGreatestElongationWithInstance(jde, nil)
}

// NextMercuryProgradeToRetrograde 下一次水星顺行转逆行的留 / next prograde-to-retrograde station of Mercury.
func NextMercuryProgradeToRetrograde(jde float64) float64 {
	return NextMercuryProgradeToRetrogradeWithInstance(jde, nil)
}

// EveningTwilightN 昏朦影（n 变体） / n-parameterised evening twilight.
func EveningTwilightN(localJD, lon, lat, tz, targetAltitude float64, n int) (float64, error) {
	return EveningTwilightNWithInstance(localJD, lon, lat, tz, targetAltitude, n, nil)
}

// GXCBo 黄纬光行差修正（角秒） / ecliptic-latitude aberration correction in arcseconds.
func GXCBo(lo, bo, jde float64) float64 { return GXCBoWithInstance(lo, bo, jde, nil) }

// GXCLo 黄经光行差修正（角秒） / ecliptic-longitude aberration correction in arcseconds.
func GXCLo(lo, bo, jde float64) float64 { return GXCLoWithInstance(lo, bo, jde, nil) }

// GetMoonRiseTime 月出时间 / moonrise time.
func GetMoonRiseTime(julianDay, longitude, latitude, timeZone, zenithShift, height float64) (float64, error) {
	return GetMoonRiseTimeWithInstance(julianDay, longitude, latitude, timeZone, zenithShift, height, nil)
}

// GetMoonSetTime 月落时间 / moonset time.
func GetMoonSetTime(julianDay, longitude, latitude, timeZone, zenithShift, height float64) (float64, error) {
	return GetMoonSetTimeWithInstance(julianDay, longitude, latitude, timeZone, zenithShift, height, nil)
}

// GetSunRiseTimeN 日出时间（n 变体） / n-parameterised sunrise time.
func GetSunRiseTimeN(julianDay, longitude, latitude, timeZone, zenithShift, height float64, n int) (float64, error) {
	return GetSunRiseTimeNWithInstance(julianDay, longitude, latitude, timeZone, zenithShift, height, n, nil)
}

// GetSunSetTimeN 日落时间（n 变体） / n-parameterised sunset time.
func GetSunSetTimeN(julianDay, longitude, latitude, timeZone, zenithShift, height float64, n int) (float64, error) {
	return GetSunSetTimeNWithInstance(julianDay, longitude, latitude, timeZone, zenithShift, height, n, nil)
}

// HMoonApparentDec 当地民用时下的月球视赤纬 / apparent lunar declination for a local civil Julian day.
func HMoonApparentDec(jd, lon, lat, tz float64) float64 {
	return HMoonApparentDecWithInstance(jd, lon, lat, tz, nil)
}

// HMoonApparentDecN 当地民用时下的月球视赤纬（n 变体） / n-parameterised apparent lunar declination.
func HMoonApparentDecN(jd, lon, lat, tz float64, n int) float64 {
	return HMoonApparentDecNWithInstance(jd, lon, lat, tz, n, nil)
}

// HMoonApparentRa 当地民用时下的月球视赤经 / apparent lunar right ascension for a local civil Julian day.
func HMoonApparentRa(jd, lon, lat, tz float64) float64 {
	return HMoonApparentRaWithInstance(jd, lon, lat, tz, nil)
}

// HMoonApparentRaN 当地民用时下的月球视赤经（n 变体） / n-parameterised apparent lunar right ascension.
func HMoonApparentRaN(jd, lon, lat, tz float64, n int) float64 {
	return HMoonApparentRaNWithInstance(jd, lon, lat, tz, n, nil)
}

// HMoonAzimuth 当地民用时下的月球方位角 / lunar azimuth for a local civil Julian day.
func HMoonAzimuth(jd, lon, lat, tz float64) float64 {
	return HMoonAzimuthWithInstance(jd, lon, lat, tz, nil)
}

// HMoonAzimuthN 当地民用时下的月球方位角（n 变体） / n-parameterised lunar azimuth.
func HMoonAzimuthN(jd, lon, lat, tz float64, n int) float64 {
	return HMoonAzimuthNWithInstance(jd, lon, lat, tz, n, nil)
}

// HMoonHeightN 月心几何高度角（n 变体） / n-parameterised geometric Moon-centre altitude.
func HMoonHeightN(jd, lon, lat, tz float64, n int) float64 {
	return HMoonHeightNWithInstance(jd, lon, lat, tz, n, nil)
}

// HMoonTrueDecN 月球真赤纬（n 变体），含章动光行差，与 HMoonGeocentricApparentDecN 同值 / n-parameterised lunar declination, identical to HMoonGeocentricApparentDecN.
func HMoonTrueDecN(jde float64, n int) float64 { return HMoonTrueDecNWithInstance(jde, n, nil) }

// HMoonTrueRaDecN 月球真赤经、真赤纬（n 变体），含章动光行差，与 HMoonGeocentricApparentRaDecN 同值 / n-parameterised lunar RA/Dec, identical to HMoonGeocentricApparentRaDecN.
func HMoonTrueRaDecN(jde float64, n int) (float64, float64) {
	return HMoonTrueRaDecNWithInstance(jde, n, nil)
}

// HMoonTrueRaN 月球真赤经（n 变体），含章动光行差，与 HMoonGeocentricApparentRaN 同值 / n-parameterised lunar right ascension, identical to HMoonGeocentricApparentRaN.
func HMoonTrueRaN(jde float64, n int) float64 { return HMoonTrueRaNWithInstance(jde, n, nil) }

// JupiterAzimuth 木星方位角 / azimuth of Jupiter.
func JupiterAzimuth(localJD, lon, lat, timezone float64) float64 {
	return JupiterAzimuthWithInstance(localJD, lon, lat, timezone, nil)
}

// JupiterCulminationTime 木星中天时间 / culmination time of Jupiter.
func JupiterCulminationTime(localJD, lon, timezone float64) float64 {
	return JupiterCulminationTimeWithInstance(localJD, lon, timezone, nil)
}

// JupiterHeight 木星高度角 / altitude of Jupiter.
func JupiterHeight(localJD, lon, lat, timezone float64) float64 {
	return JupiterHeightWithInstance(localJD, lon, lat, timezone, nil)
}

// JupiterHourAngle 木星时角 / hour angle of Jupiter.
func JupiterHourAngle(jd, lon, timezone float64) float64 {
	return JupiterHourAngleWithInstance(jd, lon, timezone, nil)
}

// JupiterRiseTime 木星升起时间 / rise time of Jupiter.
func JupiterRiseTime(jd, lon, lat, timezone, aeroCorrection, observerHeight float64) (float64, error) {
	return JupiterRiseTimeWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight, nil)
}

// JupiterSetTime 木星落下时间 / set time of Jupiter.
func JupiterSetTime(jd, lon, lat, timezone, aeroCorrection, observerHeight float64) (float64, error) {
	return JupiterSetTimeWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight, nil)
}

// LastJupiterConjunction 上一次木星合 / previous conjunction of Jupiter.
func LastJupiterConjunction(jde float64) float64 { return LastJupiterConjunctionWithInstance(jde, nil) }

// LastJupiterEasternQuadrature 上一次木星东方照 / previous eastern quadrature of Jupiter.
func LastJupiterEasternQuadrature(jde float64) float64 {
	return LastJupiterEasternQuadratureWithInstance(jde, nil)
}

// LastJupiterOpposition 上一次木星冲 / previous opposition of Jupiter.
func LastJupiterOpposition(jde float64) float64 { return LastJupiterOppositionWithInstance(jde, nil) }

// LastJupiterProgradeToRetrograde 上一次木星顺行转逆行的留 / previous prograde-to-retrograde station of Jupiter.
func LastJupiterProgradeToRetrograde(jde float64) float64 {
	return LastJupiterProgradeToRetrogradeWithInstance(jde, nil)
}

// LastJupiterRetrogradeToPrograde 上一次木星逆行转顺行的留 / previous retrograde-to-prograde station of Jupiter.
func LastJupiterRetrogradeToPrograde(jde float64) float64 {
	return LastJupiterRetrogradeToProgradeWithInstance(jde, nil)
}

// LastJupiterWesternQuadrature 上一次木星西方照 / previous western quadrature of Jupiter.
func LastJupiterWesternQuadrature(jde float64) float64 {
	return LastJupiterWesternQuadratureWithInstance(jde, nil)
}

// LastMarsConjunction 上一次火星合 / previous conjunction of Mars.
func LastMarsConjunction(jde float64) float64 { return LastMarsConjunctionWithInstance(jde, nil) }

// LastMarsEasternQuadrature 上一次火星东方照 / previous eastern quadrature of Mars.
func LastMarsEasternQuadrature(jde float64) float64 {
	return LastMarsEasternQuadratureWithInstance(jde, nil)
}

// LastMarsOpposition 上一次火星冲 / previous opposition of Mars.
func LastMarsOpposition(jde float64) float64 { return LastMarsOppositionWithInstance(jde, nil) }

// LastMarsProgradeToRetrograde 上一次火星顺行转逆行的留 / previous prograde-to-retrograde station of Mars.
func LastMarsProgradeToRetrograde(jde float64) float64 {
	return LastMarsProgradeToRetrogradeWithInstance(jde, nil)
}

// LastMarsRetrogradeToPrograde 上一次火星逆行转顺行的留 / previous retrograde-to-prograde station of Mars.
func LastMarsRetrogradeToPrograde(jde float64) float64 {
	return LastMarsRetrogradeToProgradeWithInstance(jde, nil)
}

// LastMarsWesternQuadrature 上一次火星西方照 / previous western quadrature of Mars.
func LastMarsWesternQuadrature(jde float64) float64 {
	return LastMarsWesternQuadratureWithInstance(jde, nil)
}

// LastMercuryConjunction 上一次水星合 / previous conjunction of Mercury.
func LastMercuryConjunction(jde float64) float64 { return LastMercuryConjunctionWithInstance(jde, nil) }

// LastMercuryConjunctionStrict 上一次水星合，不含查询时刻 / previous conjunction of Mercury, excluding jde.
func LastMercuryConjunctionStrict(jde float64) float64 {
	return LastMercuryConjunctionStrictWithInstance(jde, nil)
}

// LastMercuryGreatestElongation 上一次水星大距（东西任一侧） / previous greatest elongation on either side of Mercury.
func LastMercuryGreatestElongation(jde float64) float64 {
	return LastMercuryGreatestElongationWithInstance(jde, nil)
}

// LastMercuryGreatestElongationEast 上一次水星东大距 / previous greatest eastern elongation of Mercury.
func LastMercuryGreatestElongationEast(jde float64) float64 {
	return LastMercuryGreatestElongationEastWithInstance(jde, nil)
}

// LastMercuryGreatestElongationEastInclusive 上一次水星东大距，含查询时刻 / previous greatest eastern elongation of Mercury, including jde.
func LastMercuryGreatestElongationEastInclusive(jde float64) float64 {
	return LastMercuryGreatestElongationEastInclusiveWithInstance(jde, nil)
}

// LastMercuryGreatestElongationInclusive 上一次水星大距（东西任一侧），含查询时刻 / previous greatest elongation on either side of Mercury, including jde.
func LastMercuryGreatestElongationInclusive(jde float64) float64 {
	return LastMercuryGreatestElongationInclusiveWithInstance(jde, nil)
}

// LastMercuryGreatestElongationWest 上一次水星西大距 / previous greatest western elongation of Mercury.
func LastMercuryGreatestElongationWest(jde float64) float64 {
	return LastMercuryGreatestElongationWestWithInstance(jde, nil)
}

// LastMercuryGreatestElongationWestInclusive 上一次水星西大距，含查询时刻 / previous greatest western elongation of Mercury, including jde.
func LastMercuryGreatestElongationWestInclusive(jde float64) float64 {
	return LastMercuryGreatestElongationWestInclusiveWithInstance(jde, nil)
}

// LastMercuryInferiorConjunction 上一次水星下合 / previous inferior conjunction of Mercury.
func LastMercuryInferiorConjunction(jde float64) float64 {
	return LastMercuryInferiorConjunctionWithInstance(jde, nil)
}

// LastMercuryInferiorConjunctionInclusive 上一次水星下合，含查询时刻 / previous inferior conjunction of Mercury, including jde.
func LastMercuryInferiorConjunctionInclusive(jde float64) float64 {
	return LastMercuryInferiorConjunctionInclusiveWithInstance(jde, nil)
}

// LastMercuryProgradeToRetrograde 上一次水星顺行转逆行的留 / previous prograde-to-retrograde station of Mercury.
func LastMercuryProgradeToRetrograde(jde float64) float64 {
	return LastMercuryProgradeToRetrogradeWithInstance(jde, nil)
}

// LastMercuryProgradeToRetrogradeInclusive 上一次水星顺行转逆行的留，含查询时刻 / previous prograde-to-retrograde station of Mercury, including jde.
func LastMercuryProgradeToRetrogradeInclusive(jde float64) float64 {
	return LastMercuryProgradeToRetrogradeInclusiveWithInstance(jde, nil)
}

// LastMercuryRetrograde 上一次水星留（任一方向） / previous station in either direction of Mercury.
func LastMercuryRetrograde(jde float64) float64 { return LastMercuryRetrogradeWithInstance(jde, nil) }

// LastMercuryRetrogradeInclusive 上一次水星留（任一方向），含查询时刻 / previous station in either direction of Mercury, including jde.
func LastMercuryRetrogradeInclusive(jde float64) float64 {
	return LastMercuryRetrogradeInclusiveWithInstance(jde, nil)
}

// LastMercuryRetrogradeToPrograde 上一次水星逆行转顺行的留 / previous retrograde-to-prograde station of Mercury.
func LastMercuryRetrogradeToPrograde(jde float64) float64 {
	return LastMercuryRetrogradeToProgradeWithInstance(jde, nil)
}

// LastMercuryRetrogradeToProgradeInclusive 上一次水星逆行转顺行的留，含查询时刻 / previous retrograde-to-prograde station of Mercury, including jde.
func LastMercuryRetrogradeToProgradeInclusive(jde float64) float64 {
	return LastMercuryRetrogradeToProgradeInclusiveWithInstance(jde, nil)
}

// LastMercurySuperiorConjunction 上一次水星上合 / previous superior conjunction of Mercury.
func LastMercurySuperiorConjunction(jde float64) float64 {
	return LastMercurySuperiorConjunctionWithInstance(jde, nil)
}

// LastMercurySuperiorConjunctionInclusive 上一次水星上合，含查询时刻 / previous superior conjunction of Mercury, including jde.
func LastMercurySuperiorConjunctionInclusive(jde float64) float64 {
	return LastMercurySuperiorConjunctionInclusiveWithInstance(jde, nil)
}

// LastNeptuneConjunction 上一次海王星合 / previous conjunction of Neptune.
func LastNeptuneConjunction(jde float64) float64 { return LastNeptuneConjunctionWithInstance(jde, nil) }

// LastNeptuneEasternQuadrature 上一次海王星东方照 / previous eastern quadrature of Neptune.
func LastNeptuneEasternQuadrature(jde float64) float64 {
	return LastNeptuneEasternQuadratureWithInstance(jde, nil)
}

// LastNeptuneOpposition 上一次海王星冲 / previous opposition of Neptune.
func LastNeptuneOpposition(jde float64) float64 { return LastNeptuneOppositionWithInstance(jde, nil) }

// LastNeptuneProgradeToRetrograde 上一次海王星顺行转逆行的留 / previous prograde-to-retrograde station of Neptune.
func LastNeptuneProgradeToRetrograde(jde float64) float64 {
	return LastNeptuneProgradeToRetrogradeWithInstance(jde, nil)
}

// LastNeptuneRetrogradeToPrograde 上一次海王星逆行转顺行的留 / previous retrograde-to-prograde station of Neptune.
func LastNeptuneRetrogradeToPrograde(jde float64) float64 {
	return LastNeptuneRetrogradeToProgradeWithInstance(jde, nil)
}

// LastNeptuneWesternQuadrature 上一次海王星西方照 / previous western quadrature of Neptune.
func LastNeptuneWesternQuadrature(jde float64) float64 {
	return LastNeptuneWesternQuadratureWithInstance(jde, nil)
}

// LastSaturnConjunction 上一次土星合 / previous conjunction of Saturn.
func LastSaturnConjunction(jde float64) float64 { return LastSaturnConjunctionWithInstance(jde, nil) }

// LastSaturnEasternQuadrature 上一次土星东方照 / previous eastern quadrature of Saturn.
func LastSaturnEasternQuadrature(jde float64) float64 {
	return LastSaturnEasternQuadratureWithInstance(jde, nil)
}

// LastSaturnOpposition 上一次土星冲 / previous opposition of Saturn.
func LastSaturnOpposition(jde float64) float64 { return LastSaturnOppositionWithInstance(jde, nil) }

// LastSaturnProgradeToRetrograde 上一次土星顺行转逆行的留 / previous prograde-to-retrograde station of Saturn.
func LastSaturnProgradeToRetrograde(jde float64) float64 {
	return LastSaturnProgradeToRetrogradeWithInstance(jde, nil)
}

// LastSaturnRetrogradeToPrograde 上一次土星逆行转顺行的留 / previous retrograde-to-prograde station of Saturn.
func LastSaturnRetrogradeToPrograde(jde float64) float64 {
	return LastSaturnRetrogradeToProgradeWithInstance(jde, nil)
}

// LastSaturnWesternQuadrature 上一次土星西方照 / previous western quadrature of Saturn.
func LastSaturnWesternQuadrature(jde float64) float64 {
	return LastSaturnWesternQuadratureWithInstance(jde, nil)
}

// LastUranusConjunction 上一次天王星合 / previous conjunction of Uranus.
func LastUranusConjunction(jde float64) float64 { return LastUranusConjunctionWithInstance(jde, nil) }

// LastUranusEasternQuadrature 上一次天王星东方照 / previous eastern quadrature of Uranus.
func LastUranusEasternQuadrature(jde float64) float64 {
	return LastUranusEasternQuadratureWithInstance(jde, nil)
}

// LastUranusOpposition 上一次天王星冲 / previous opposition of Uranus.
func LastUranusOpposition(jde float64) float64 { return LastUranusOppositionWithInstance(jde, nil) }

// LastUranusProgradeToRetrograde 上一次天王星顺行转逆行的留 / previous prograde-to-retrograde station of Uranus.
func LastUranusProgradeToRetrograde(jde float64) float64 {
	return LastUranusProgradeToRetrogradeWithInstance(jde, nil)
}

// LastUranusRetrogradeToPrograde 上一次天王星逆行转顺行的留 / previous retrograde-to-prograde station of Uranus.
func LastUranusRetrogradeToPrograde(jde float64) float64 {
	return LastUranusRetrogradeToProgradeWithInstance(jde, nil)
}

// LastUranusWesternQuadrature 上一次天王星西方照 / previous western quadrature of Uranus.
func LastUranusWesternQuadrature(jde float64) float64 {
	return LastUranusWesternQuadratureWithInstance(jde, nil)
}

// LastVenusConjunction 上一次金星合 / previous conjunction of Venus.
func LastVenusConjunction(jde float64) float64 { return LastVenusConjunctionWithInstance(jde, nil) }

// LastVenusConjunctionStrict 上一次金星合，不含查询时刻 / previous conjunction of Venus, excluding jde.
func LastVenusConjunctionStrict(jde float64) float64 {
	return LastVenusConjunctionStrictWithInstance(jde, nil)
}

// LastVenusGreatestElongation 上一次金星大距（东西任一侧） / previous greatest elongation on either side of Venus.
func LastVenusGreatestElongation(jde float64) float64 {
	return LastVenusGreatestElongationWithInstance(jde, nil)
}

// LastVenusGreatestElongationEast 上一次金星东大距 / previous greatest eastern elongation of Venus.
func LastVenusGreatestElongationEast(jde float64) float64 {
	return LastVenusGreatestElongationEastWithInstance(jde, nil)
}

// LastVenusGreatestElongationEastInclusive 上一次金星东大距，含查询时刻 / previous greatest eastern elongation of Venus, including jde.
func LastVenusGreatestElongationEastInclusive(jde float64) float64 {
	return LastVenusGreatestElongationEastInclusiveWithInstance(jde, nil)
}

// LastVenusGreatestElongationInclusive 上一次金星大距（东西任一侧），含查询时刻 / previous greatest elongation on either side of Venus, including jde.
func LastVenusGreatestElongationInclusive(jde float64) float64 {
	return LastVenusGreatestElongationInclusiveWithInstance(jde, nil)
}

// LastVenusGreatestElongationWest 上一次金星西大距 / previous greatest western elongation of Venus.
func LastVenusGreatestElongationWest(jde float64) float64 {
	return LastVenusGreatestElongationWestWithInstance(jde, nil)
}

// LastVenusGreatestElongationWestInclusive 上一次金星西大距，含查询时刻 / previous greatest western elongation of Venus, including jde.
func LastVenusGreatestElongationWestInclusive(jde float64) float64 {
	return LastVenusGreatestElongationWestInclusiveWithInstance(jde, nil)
}

// LastVenusInferiorConjunction 上一次金星下合 / previous inferior conjunction of Venus.
func LastVenusInferiorConjunction(jde float64) float64 {
	return LastVenusInferiorConjunctionWithInstance(jde, nil)
}

// LastVenusInferiorConjunctionInclusive 上一次金星下合，含查询时刻 / previous inferior conjunction of Venus, including jde.
func LastVenusInferiorConjunctionInclusive(jde float64) float64 {
	return LastVenusInferiorConjunctionInclusiveWithInstance(jde, nil)
}

// LastVenusProgradeToRetrograde 上一次金星顺行转逆行的留 / previous prograde-to-retrograde station of Venus.
func LastVenusProgradeToRetrograde(jde float64) float64 {
	return LastVenusProgradeToRetrogradeWithInstance(jde, nil)
}

// LastVenusProgradeToRetrogradeInclusive 上一次金星顺行转逆行的留，含查询时刻 / previous prograde-to-retrograde station of Venus, including jde.
func LastVenusProgradeToRetrogradeInclusive(jde float64) float64 {
	return LastVenusProgradeToRetrogradeInclusiveWithInstance(jde, nil)
}

// LastVenusRetrograde 上一次金星留（任一方向） / previous station in either direction of Venus.
func LastVenusRetrograde(jde float64) float64 { return LastVenusRetrogradeWithInstance(jde, nil) }

// LastVenusRetrogradeInclusive 上一次金星留（任一方向），含查询时刻 / previous station in either direction of Venus, including jde.
func LastVenusRetrogradeInclusive(jde float64) float64 {
	return LastVenusRetrogradeInclusiveWithInstance(jde, nil)
}

// LastVenusRetrogradeToPrograde 上一次金星逆行转顺行的留 / previous retrograde-to-prograde station of Venus.
func LastVenusRetrogradeToPrograde(jde float64) float64 {
	return LastVenusRetrogradeToProgradeWithInstance(jde, nil)
}

// LastVenusRetrogradeToProgradeInclusive 上一次金星逆行转顺行的留，含查询时刻 / previous retrograde-to-prograde station of Venus, including jde.
func LastVenusRetrogradeToProgradeInclusive(jde float64) float64 {
	return LastVenusRetrogradeToProgradeInclusiveWithInstance(jde, nil)
}

// LastVenusSuperiorConjunction 上一次金星上合 / previous superior conjunction of Venus.
func LastVenusSuperiorConjunction(jde float64) float64 {
	return LastVenusSuperiorConjunctionWithInstance(jde, nil)
}

// LastVenusSuperiorConjunctionInclusive 上一次金星上合，含查询时刻 / previous superior conjunction of Venus, including jde.
func LastVenusSuperiorConjunctionInclusive(jde float64) float64 {
	return LastVenusSuperiorConjunctionInclusiveWithInstance(jde, nil)
}

// MarsAzimuth 火星方位角 / azimuth of Mars.
func MarsAzimuth(localJD, lon, lat, timezone float64) float64 {
	return MarsAzimuthWithInstance(localJD, lon, lat, timezone, nil)
}

// MarsCulminationTime 火星中天时间 / culmination time of Mars.
func MarsCulminationTime(localJD, lon, timezone float64) float64 {
	return MarsCulminationTimeWithInstance(localJD, lon, timezone, nil)
}

// MarsHeight 火星高度角 / altitude of Mars.
func MarsHeight(localJD, lon, lat, timezone float64) float64 {
	return MarsHeightWithInstance(localJD, lon, lat, timezone, nil)
}

// MarsHourAngle 火星时角 / hour angle of Mars.
func MarsHourAngle(jd, lon, timezone float64) float64 {
	return MarsHourAngleWithInstance(jd, lon, timezone, nil)
}

// MarsRiseTime 火星升起时间 / rise time of Mars.
func MarsRiseTime(jd, lon, lat, timezone, aeroCorrection, observerHeight float64) (float64, error) {
	return MarsRiseTimeWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight, nil)
}

// MarsSetTime 火星落下时间 / set time of Mars.
func MarsSetTime(jd, lon, lat, timezone, aeroCorrection, observerHeight float64) (float64, error) {
	return MarsSetTimeWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight, nil)
}

// MercuryAzimuth 水星方位角 / azimuth of Mercury.
func MercuryAzimuth(localJD, lon, lat, timezone float64) float64 {
	return MercuryAzimuthWithInstance(localJD, lon, lat, timezone, nil)
}

// MercuryCulminationTime 水星中天时间 / culmination time of Mercury.
func MercuryCulminationTime(localJD, lon, timezone float64) float64 {
	return MercuryCulminationTimeWithInstance(localJD, lon, timezone, nil)
}

// MercuryHeight 水星高度角 / altitude of Mercury.
func MercuryHeight(localJD, lon, lat, timezone float64) float64 {
	return MercuryHeightWithInstance(localJD, lon, lat, timezone, nil)
}

// MercuryHourAngle 水星时角 / hour angle of Mercury.
func MercuryHourAngle(jd, lon, timezone float64) float64 {
	return MercuryHourAngleWithInstance(jd, lon, timezone, nil)
}

// MercuryRiseTime 水星升起时间 / rise time of Mercury.
func MercuryRiseTime(jd, lon, lat, timezone, aeroCorrection, observerHeight float64) (float64, error) {
	return MercuryRiseTimeWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight, nil)
}

// MercurySetTime 水星落下时间 / set time of Mercury.
func MercurySetTime(jd, lon, lat, timezone, aeroCorrection, observerHeight float64) (float64, error) {
	return MercurySetTimeWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight, nil)
}

// MercurySunElongation 水星与太阳的角距 / elongation from the Sun of Mercury.
func MercurySunElongation(jde float64) float64 { return MercurySunElongationWithInstance(jde, nil) }

// MoonCulminationTime 月球中天时间 / lunar culmination time.
func MoonCulminationTime(localJD, lon, lat, timezone float64) float64 {
	return MoonCulminationTimeWithInstance(localJD, lon, lat, timezone, nil)
}

// MoonTimeAngle 月球时角 / lunar hour angle.
func MoonTimeAngle(jd, lon, lat, tz float64) float64 {
	return MoonTimeAngleWithInstance(jd, lon, lat, tz, nil)
}

// MorningTwilight 晨朦影；localJD 是本地民用日锚点（当地 0 时），只取整数日 / morning twilight on the local civil day.
func MorningTwilight(localJD, lon, lat, tz, targetAltitude float64) (float64, error) {
	return MorningTwilightWithInstance(localJD, lon, lat, tz, targetAltitude, nil)
}

// MorningTwilightN 晨朦影（n 变体） / n-parameterised morning twilight.
func MorningTwilightN(localJD, lon, lat, tz, targetAltitude float64, n int) (float64, error) {
	return MorningTwilightNWithInstance(localJD, lon, lat, tz, targetAltitude, n, nil)
}

// NeptuneAzimuth 海王星方位角 / azimuth of Neptune.
func NeptuneAzimuth(localJD, lon, lat, timezone float64) float64 {
	return NeptuneAzimuthWithInstance(localJD, lon, lat, timezone, nil)
}

// NeptuneCulminationTime 海王星中天时间 / culmination time of Neptune.
func NeptuneCulminationTime(localJD, lon, timezone float64) float64 {
	return NeptuneCulminationTimeWithInstance(localJD, lon, timezone, nil)
}

// NeptuneHeight 海王星高度角 / altitude of Neptune.
func NeptuneHeight(localJD, lon, lat, timezone float64) float64 {
	return NeptuneHeightWithInstance(localJD, lon, lat, timezone, nil)
}

// NeptuneHourAngle 海王星时角 / hour angle of Neptune.
func NeptuneHourAngle(jd, lon, timezone float64) float64 {
	return NeptuneHourAngleWithInstance(jd, lon, timezone, nil)
}

// NeptuneRiseTime 海王星升起时间 / rise time of Neptune.
func NeptuneRiseTime(jd, lon, lat, timezone, aeroCorrection, observerHeight float64) (float64, error) {
	return NeptuneRiseTimeWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight, nil)
}

// NeptuneSetTime 海王星落下时间 / set time of Neptune.
func NeptuneSetTime(jd, lon, lat, timezone, aeroCorrection, observerHeight float64) (float64, error) {
	return NeptuneSetTimeWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight, nil)
}

// NextJupiterConjunction 下一次木星合 / next conjunction of Jupiter.
func NextJupiterConjunction(jde float64) float64 { return NextJupiterConjunctionWithInstance(jde, nil) }

// NextJupiterEasternQuadrature 下一次木星东方照 / next eastern quadrature of Jupiter.
func NextJupiterEasternQuadrature(jde float64) float64 {
	return NextJupiterEasternQuadratureWithInstance(jde, nil)
}

// NextJupiterOpposition 下一次木星冲 / next opposition of Jupiter.
func NextJupiterOpposition(jde float64) float64 { return NextJupiterOppositionWithInstance(jde, nil) }

// NextJupiterProgradeToRetrograde 下一次木星顺行转逆行的留 / next prograde-to-retrograde station of Jupiter.
func NextJupiterProgradeToRetrograde(jde float64) float64 {
	return NextJupiterProgradeToRetrogradeWithInstance(jde, nil)
}

// NextJupiterRetrogradeToPrograde 下一次木星逆行转顺行的留 / next retrograde-to-prograde station of Jupiter.
func NextJupiterRetrogradeToPrograde(jde float64) float64 {
	return NextJupiterRetrogradeToProgradeWithInstance(jde, nil)
}

// NextJupiterWesternQuadrature 下一次木星西方照 / next western quadrature of Jupiter.
func NextJupiterWesternQuadrature(jde float64) float64 {
	return NextJupiterWesternQuadratureWithInstance(jde, nil)
}

// NextMarsConjunction 下一次火星合 / next conjunction of Mars.
func NextMarsConjunction(jde float64) float64 { return NextMarsConjunctionWithInstance(jde, nil) }

// NextMarsEasternQuadrature 下一次火星东方照 / next eastern quadrature of Mars.
func NextMarsEasternQuadrature(jde float64) float64 {
	return NextMarsEasternQuadratureWithInstance(jde, nil)
}

// NextMarsOpposition 下一次火星冲 / next opposition of Mars.
func NextMarsOpposition(jde float64) float64 { return NextMarsOppositionWithInstance(jde, nil) }

// NextMarsProgradeToRetrograde 下一次火星顺行转逆行的留 / next prograde-to-retrograde station of Mars.
func NextMarsProgradeToRetrograde(jde float64) float64 {
	return NextMarsProgradeToRetrogradeWithInstance(jde, nil)
}

// NextMarsRetrogradeToPrograde 下一次火星逆行转顺行的留 / next retrograde-to-prograde station of Mars.
func NextMarsRetrogradeToPrograde(jde float64) float64 {
	return NextMarsRetrogradeToProgradeWithInstance(jde, nil)
}

// NextMarsWesternQuadrature 下一次火星西方照 / next western quadrature of Mars.
func NextMarsWesternQuadrature(jde float64) float64 {
	return NextMarsWesternQuadratureWithInstance(jde, nil)
}

// NextMercuryConjunction 下一次水星合 / next conjunction of Mercury.
func NextMercuryConjunction(jde float64) float64 { return NextMercuryConjunctionWithInstance(jde, nil) }

// NextMercuryConjunctionStrict 下一次水星合，不含查询时刻 / next conjunction of Mercury, excluding jde.
func NextMercuryConjunctionStrict(jde float64) float64 {
	return NextMercuryConjunctionStrictWithInstance(jde, nil)
}

// NextMercuryGreatestElongationEast 下一次水星东大距 / next greatest eastern elongation of Mercury.
func NextMercuryGreatestElongationEast(jde float64) float64 {
	return NextMercuryGreatestElongationEastWithInstance(jde, nil)
}

// NextMercuryGreatestElongationEastInclusive 下一次水星东大距，含查询时刻 / next greatest eastern elongation of Mercury, including jde.
func NextMercuryGreatestElongationEastInclusive(jde float64) float64 {
	return NextMercuryGreatestElongationEastInclusiveWithInstance(jde, nil)
}

// NextMercuryGreatestElongationInclusive 下一次水星大距（东西任一侧），含查询时刻 / next greatest elongation on either side of Mercury, including jde.
func NextMercuryGreatestElongationInclusive(jde float64) float64 {
	return NextMercuryGreatestElongationInclusiveWithInstance(jde, nil)
}

// NextMercuryGreatestElongationWest 下一次水星西大距 / next greatest western elongation of Mercury.
func NextMercuryGreatestElongationWest(jde float64) float64 {
	return NextMercuryGreatestElongationWestWithInstance(jde, nil)
}

// NextMercuryGreatestElongationWestInclusive 下一次水星西大距，含查询时刻 / next greatest western elongation of Mercury, including jde.
func NextMercuryGreatestElongationWestInclusive(jde float64) float64 {
	return NextMercuryGreatestElongationWestInclusiveWithInstance(jde, nil)
}

// NextMercuryInferiorConjunction 下一次水星下合 / next inferior conjunction of Mercury.
func NextMercuryInferiorConjunction(jde float64) float64 {
	return NextMercuryInferiorConjunctionWithInstance(jde, nil)
}

// NextMercuryInferiorConjunctionInclusive 下一次水星下合，含查询时刻 / next inferior conjunction of Mercury, including jde.
func NextMercuryInferiorConjunctionInclusive(jde float64) float64 {
	return NextMercuryInferiorConjunctionInclusiveWithInstance(jde, nil)
}

// NextMercuryProgradeToRetrogradeInclusive 下一次水星顺行转逆行的留，含查询时刻 / next prograde-to-retrograde station of Mercury, including jde.
func NextMercuryProgradeToRetrogradeInclusive(jde float64) float64 {
	return NextMercuryProgradeToRetrogradeInclusiveWithInstance(jde, nil)
}

// NextMercuryRetrograde 下一次水星留（任一方向） / next station in either direction of Mercury.
func NextMercuryRetrograde(jde float64) float64 { return NextMercuryRetrogradeWithInstance(jde, nil) }

// NextMercuryRetrogradeInclusive 下一次水星留（任一方向），含查询时刻 / next station in either direction of Mercury, including jde.
func NextMercuryRetrogradeInclusive(jde float64) float64 {
	return NextMercuryRetrogradeInclusiveWithInstance(jde, nil)
}

// NextMercuryRetrogradeToPrograde 下一次水星逆行转顺行的留 / next retrograde-to-prograde station of Mercury.
func NextMercuryRetrogradeToPrograde(jde float64) float64 {
	return NextMercuryRetrogradeToProgradeWithInstance(jde, nil)
}

// NextMercuryRetrogradeToProgradeInclusive 下一次水星逆行转顺行的留，含查询时刻 / next retrograde-to-prograde station of Mercury, including jde.
func NextMercuryRetrogradeToProgradeInclusive(jde float64) float64 {
	return NextMercuryRetrogradeToProgradeInclusiveWithInstance(jde, nil)
}

// NextMercurySuperiorConjunction 下一次水星上合 / next superior conjunction of Mercury.
func NextMercurySuperiorConjunction(jde float64) float64 {
	return NextMercurySuperiorConjunctionWithInstance(jde, nil)
}

// NextMercurySuperiorConjunctionInclusive 下一次水星上合，含查询时刻 / next superior conjunction of Mercury, including jde.
func NextMercurySuperiorConjunctionInclusive(jde float64) float64 {
	return NextMercurySuperiorConjunctionInclusiveWithInstance(jde, nil)
}

// NextNeptuneConjunction 下一次海王星合 / next conjunction of Neptune.
func NextNeptuneConjunction(jde float64) float64 { return NextNeptuneConjunctionWithInstance(jde, nil) }

// NextNeptuneEasternQuadrature 下一次海王星东方照 / next eastern quadrature of Neptune.
func NextNeptuneEasternQuadrature(jde float64) float64 {
	return NextNeptuneEasternQuadratureWithInstance(jde, nil)
}

// NextNeptuneOpposition 下一次海王星冲 / next opposition of Neptune.
func NextNeptuneOpposition(jde float64) float64 { return NextNeptuneOppositionWithInstance(jde, nil) }

// NextNeptuneProgradeToRetrograde 下一次海王星顺行转逆行的留 / next prograde-to-retrograde station of Neptune.
func NextNeptuneProgradeToRetrograde(jde float64) float64 {
	return NextNeptuneProgradeToRetrogradeWithInstance(jde, nil)
}

// NextNeptuneRetrogradeToPrograde 下一次海王星逆行转顺行的留 / next retrograde-to-prograde station of Neptune.
func NextNeptuneRetrogradeToPrograde(jde float64) float64 {
	return NextNeptuneRetrogradeToProgradeWithInstance(jde, nil)
}

// NextNeptuneWesternQuadrature 下一次海王星西方照 / next western quadrature of Neptune.
func NextNeptuneWesternQuadrature(jde float64) float64 {
	return NextNeptuneWesternQuadratureWithInstance(jde, nil)
}

// NextSaturnConjunction 下一次土星合 / next conjunction of Saturn.
func NextSaturnConjunction(jde float64) float64 { return NextSaturnConjunctionWithInstance(jde, nil) }

// NextSaturnEasternQuadrature 下一次土星东方照 / next eastern quadrature of Saturn.
func NextSaturnEasternQuadrature(jde float64) float64 {
	return NextSaturnEasternQuadratureWithInstance(jde, nil)
}

// NextSaturnOpposition 下一次土星冲 / next opposition of Saturn.
func NextSaturnOpposition(jde float64) float64 { return NextSaturnOppositionWithInstance(jde, nil) }

// NextSaturnProgradeToRetrograde 下一次土星顺行转逆行的留 / next prograde-to-retrograde station of Saturn.
func NextSaturnProgradeToRetrograde(jde float64) float64 {
	return NextSaturnProgradeToRetrogradeWithInstance(jde, nil)
}

// NextSaturnRetrogradeToPrograde 下一次土星逆行转顺行的留 / next retrograde-to-prograde station of Saturn.
func NextSaturnRetrogradeToPrograde(jde float64) float64 {
	return NextSaturnRetrogradeToProgradeWithInstance(jde, nil)
}

// NextSaturnWesternQuadrature 下一次土星西方照 / next western quadrature of Saturn.
func NextSaturnWesternQuadrature(jde float64) float64 {
	return NextSaturnWesternQuadratureWithInstance(jde, nil)
}

// NextUranusConjunction 下一次天王星合 / next conjunction of Uranus.
func NextUranusConjunction(jde float64) float64 { return NextUranusConjunctionWithInstance(jde, nil) }

// NextUranusEasternQuadrature 下一次天王星东方照 / next eastern quadrature of Uranus.
func NextUranusEasternQuadrature(jde float64) float64 {
	return NextUranusEasternQuadratureWithInstance(jde, nil)
}

// NextUranusOpposition 下一次天王星冲 / next opposition of Uranus.
func NextUranusOpposition(jde float64) float64 { return NextUranusOppositionWithInstance(jde, nil) }

// NextUranusProgradeToRetrograde 下一次天王星顺行转逆行的留 / next prograde-to-retrograde station of Uranus.
func NextUranusProgradeToRetrograde(jde float64) float64 {
	return NextUranusProgradeToRetrogradeWithInstance(jde, nil)
}

// NextUranusRetrogradeToPrograde 下一次天王星逆行转顺行的留 / next retrograde-to-prograde station of Uranus.
func NextUranusRetrogradeToPrograde(jde float64) float64 {
	return NextUranusRetrogradeToProgradeWithInstance(jde, nil)
}

// NextUranusWesternQuadrature 下一次天王星西方照 / next western quadrature of Uranus.
func NextUranusWesternQuadrature(jde float64) float64 {
	return NextUranusWesternQuadratureWithInstance(jde, nil)
}

// NextVenusConjunction 下一次金星合 / next conjunction of Venus.
func NextVenusConjunction(jde float64) float64 { return NextVenusConjunctionWithInstance(jde, nil) }

// NextVenusConjunctionStrict 下一次金星合，不含查询时刻 / next conjunction of Venus, excluding jde.
func NextVenusConjunctionStrict(jde float64) float64 {
	return NextVenusConjunctionStrictWithInstance(jde, nil)
}

// NextVenusGreatestElongationEast 下一次金星东大距 / next greatest eastern elongation of Venus.
func NextVenusGreatestElongationEast(jde float64) float64 {
	return NextVenusGreatestElongationEastWithInstance(jde, nil)
}

// NextVenusGreatestElongationEastInclusive 下一次金星东大距，含查询时刻 / next greatest eastern elongation of Venus, including jde.
func NextVenusGreatestElongationEastInclusive(jde float64) float64 {
	return NextVenusGreatestElongationEastInclusiveWithInstance(jde, nil)
}

// NextVenusGreatestElongationInclusive 下一次金星大距（东西任一侧），含查询时刻 / next greatest elongation on either side of Venus, including jde.
func NextVenusGreatestElongationInclusive(jde float64) float64 {
	return NextVenusGreatestElongationInclusiveWithInstance(jde, nil)
}

// NextVenusGreatestElongationWest 下一次金星西大距 / next greatest western elongation of Venus.
func NextVenusGreatestElongationWest(jde float64) float64 {
	return NextVenusGreatestElongationWestWithInstance(jde, nil)
}

// NextVenusGreatestElongationWestInclusive 下一次金星西大距，含查询时刻 / next greatest western elongation of Venus, including jde.
func NextVenusGreatestElongationWestInclusive(jde float64) float64 {
	return NextVenusGreatestElongationWestInclusiveWithInstance(jde, nil)
}

// NextVenusInferiorConjunction 下一次金星下合 / next inferior conjunction of Venus.
func NextVenusInferiorConjunction(jde float64) float64 {
	return NextVenusInferiorConjunctionWithInstance(jde, nil)
}

// NextVenusInferiorConjunctionInclusive 下一次金星下合，含查询时刻 / next inferior conjunction of Venus, including jde.
func NextVenusInferiorConjunctionInclusive(jde float64) float64 {
	return NextVenusInferiorConjunctionInclusiveWithInstance(jde, nil)
}

// NextVenusProgradeToRetrograde 下一次金星顺行转逆行的留 / next prograde-to-retrograde station of Venus.
func NextVenusProgradeToRetrograde(jde float64) float64 {
	return NextVenusProgradeToRetrogradeWithInstance(jde, nil)
}

// NextVenusProgradeToRetrogradeInclusive 下一次金星顺行转逆行的留，含查询时刻 / next prograde-to-retrograde station of Venus, including jde.
func NextVenusProgradeToRetrogradeInclusive(jde float64) float64 {
	return NextVenusProgradeToRetrogradeInclusiveWithInstance(jde, nil)
}

// NextVenusRetrograde 下一次金星留（任一方向） / next station in either direction of Venus.
func NextVenusRetrograde(jde float64) float64 { return NextVenusRetrogradeWithInstance(jde, nil) }

// NextVenusRetrogradeInclusive 下一次金星留（任一方向），含查询时刻 / next station in either direction of Venus, including jde.
func NextVenusRetrogradeInclusive(jde float64) float64 {
	return NextVenusRetrogradeInclusiveWithInstance(jde, nil)
}

// NextVenusRetrogradeToPrograde 下一次金星逆行转顺行的留 / next retrograde-to-prograde station of Venus.
func NextVenusRetrogradeToPrograde(jde float64) float64 {
	return NextVenusRetrogradeToProgradeWithInstance(jde, nil)
}

// NextVenusRetrogradeToProgradeInclusive 下一次金星逆行转顺行的留，含查询时刻 / next retrograde-to-prograde station of Venus, including jde.
func NextVenusRetrogradeToProgradeInclusive(jde float64) float64 {
	return NextVenusRetrogradeToProgradeInclusiveWithInstance(jde, nil)
}

// NextVenusSuperiorConjunction 下一次金星上合 / next superior conjunction of Venus.
func NextVenusSuperiorConjunction(jde float64) float64 {
	return NextVenusSuperiorConjunctionWithInstance(jde, nil)
}

// NextVenusSuperiorConjunctionInclusive 下一次金星上合，含查询时刻 / next superior conjunction of Venus, including jde.
func NextVenusSuperiorConjunctionInclusive(jde float64) float64 {
	return NextVenusSuperiorConjunctionInclusiveWithInstance(jde, nil)
}

// SaturnAzimuth 土星方位角 / azimuth of Saturn.
func SaturnAzimuth(localJD, lon, lat, timezone float64) float64 {
	return SaturnAzimuthWithInstance(localJD, lon, lat, timezone, nil)
}

// SaturnCulminationTime 土星中天时间 / culmination time of Saturn.
func SaturnCulminationTime(localJD, lon, timezone float64) float64 {
	return SaturnCulminationTimeWithInstance(localJD, lon, timezone, nil)
}

// SaturnHeight 土星高度角 / altitude of Saturn.
func SaturnHeight(localJD, lon, lat, timezone float64) float64 {
	return SaturnHeightWithInstance(localJD, lon, lat, timezone, nil)
}

// SaturnHourAngle 土星时角 / hour angle of Saturn.
func SaturnHourAngle(jd, lon, timezone float64) float64 {
	return SaturnHourAngleWithInstance(jd, lon, timezone, nil)
}

// SaturnRiseTime 土星升起时间 / rise time of Saturn.
func SaturnRiseTime(jd, lon, lat, timezone, aeroCorrection, observerHeight float64) (float64, error) {
	return SaturnRiseTimeWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight, nil)
}

// SaturnSetTime 土星落下时间 / set time of Saturn.
func SaturnSetTime(jd, lon, lat, timezone, aeroCorrection, observerHeight float64) (float64, error) {
	return SaturnSetTimeWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight, nil)
}

// UranusAzimuth 天王星方位角 / azimuth of Uranus.
func UranusAzimuth(localJD, lon, lat, timezone float64) float64 {
	return UranusAzimuthWithInstance(localJD, lon, lat, timezone, nil)
}

// UranusCulminationTime 天王星中天时间 / culmination time of Uranus.
func UranusCulminationTime(localJD, lon, timezone float64) float64 {
	return UranusCulminationTimeWithInstance(localJD, lon, timezone, nil)
}

// UranusHeight 天王星高度角 / altitude of Uranus.
func UranusHeight(localJD, lon, lat, timezone float64) float64 {
	return UranusHeightWithInstance(localJD, lon, lat, timezone, nil)
}

// UranusHourAngle 天王星时角 / hour angle of Uranus.
func UranusHourAngle(jd, lon, timezone float64) float64 {
	return UranusHourAngleWithInstance(jd, lon, timezone, nil)
}

// UranusRiseTime 天王星升起时间 / rise time of Uranus.
func UranusRiseTime(jd, lon, lat, timezone, aeroCorrection, observerHeight float64) (float64, error) {
	return UranusRiseTimeWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight, nil)
}

// UranusSetTime 天王星落下时间 / set time of Uranus.
func UranusSetTime(jd, lon, lat, timezone, aeroCorrection, observerHeight float64) (float64, error) {
	return UranusSetTimeWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight, nil)
}

// VenusAzimuth 金星方位角 / azimuth of Venus.
func VenusAzimuth(localJD, lon, lat, timezone float64) float64 {
	return VenusAzimuthWithInstance(localJD, lon, lat, timezone, nil)
}

// VenusCulminationTime 金星中天时间 / culmination time of Venus.
func VenusCulminationTime(localJD, lon, timezone float64) float64 {
	return VenusCulminationTimeWithInstance(localJD, lon, timezone, nil)
}

// VenusHeight 金星高度角 / altitude of Venus.
func VenusHeight(localJD, lon, lat, timezone float64) float64 {
	return VenusHeightWithInstance(localJD, lon, lat, timezone, nil)
}

// VenusHourAngle 金星时角 / hour angle of Venus.
func VenusHourAngle(jd, lon, tz float64) float64 { return VenusHourAngleWithInstance(jd, lon, tz, nil) }

// VenusRiseTime 金星升起时间 / rise time of Venus.
func VenusRiseTime(jd, lon, lat, tz, aeroCorrection, observerHeight float64) (float64, error) {
	return VenusRiseTimeWithInstance(jd, lon, lat, tz, aeroCorrection, observerHeight, nil)
}

// VenusSetTime 金星落下时间 / set time of Venus.
func VenusSetTime(jd, lon, lat, tz, aeroCorrection, observerHeight float64) (float64, error) {
	return VenusSetTimeWithInstance(jd, lon, lat, tz, aeroCorrection, observerHeight, nil)
}

// VenusSunElongation 金星与太阳的角距 / elongation from the Sun of Venus.
func VenusSunElongation(jde float64) float64 { return VenusSunElongationWithInstance(jde, nil) }
