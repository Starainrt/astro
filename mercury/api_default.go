package mercury

import (
	"time"
)

// Altitude 站心视高度角，不含大气折射 / topocentric apparent altitude without refraction.
func Altitude(date time.Time, lon, lat float64) float64 {
	return defaultModule.Altitude(date, lon, lat)
}

// AltitudeN 站心视高度角（n 变体） / n-parameterised topocentric apparent altitude.
func AltitudeN(date time.Time, lon, lat float64, n int) float64 {
	return defaultModule.AltitudeN(date, lon, lat, n)
}

// ApparentBo 视黄纬 / apparent ecliptic latitude.
func ApparentBo(date time.Time) float64 { return defaultModule.ApparentBo(date) }

// ApparentBoN 视黄纬（n 变体） / n-parameterised apparent ecliptic latitude.
func ApparentBoN(date time.Time, n int) float64 { return defaultModule.ApparentBoN(date, n) }

// ApparentDec 视赤纬 / apparent declination.
func ApparentDec(date time.Time) float64 { return defaultModule.ApparentDec(date) }

// ApparentDecN 视赤纬（n 变体） / n-parameterised apparent declination.
func ApparentDecN(date time.Time, n int) float64 { return defaultModule.ApparentDecN(date, n) }

// ApparentLo 视黄经 / apparent ecliptic longitude.
func ApparentLo(date time.Time) float64 { return defaultModule.ApparentLo(date) }

// ApparentLoN 视黄经（n 变体） / n-parameterised apparent ecliptic longitude.
func ApparentLoN(date time.Time, n int) float64 { return defaultModule.ApparentLoN(date, n) }

// ApparentMagnitude 视星等 / apparent magnitude.
func ApparentMagnitude(date time.Time) float64 { return defaultModule.ApparentMagnitude(date) }

// ApparentMagnitudeN 视星等（n 变体） / n-parameterised apparent magnitude.
func ApparentMagnitudeN(date time.Time, n int) float64 {
	return defaultModule.ApparentMagnitudeN(date, n)
}

// ApparentRa 视赤经 / apparent right ascension.
func ApparentRa(date time.Time) float64 { return defaultModule.ApparentRa(date) }

// ApparentRaDec 视赤经、视赤纬 / apparent right ascension and declination.
func ApparentRaDec(date time.Time) (float64, float64) { return defaultModule.ApparentRaDec(date) }

// ApparentRaDecN 视赤经赤纬（n 变体） / n-parameterised apparent right ascension and declination.
func ApparentRaDecN(date time.Time, n int) (float64, float64) {
	return defaultModule.ApparentRaDecN(date, n)
}

// ApparentRaN 视赤经（n 变体） / n-parameterised apparent right ascension.
func ApparentRaN(date time.Time, n int) float64 { return defaultModule.ApparentRaN(date, n) }

// AscendingNode 水星升交点黄经 / ascending node longitude of Mercury.
func AscendingNode(date time.Time) float64 { return defaultModule.AscendingNode(date) }

// AscendingNodeN 水星升交点黄经（n 变体） / n-parameterised ascending node longitude of Mercury.
func AscendingNodeN(date time.Time, n int) float64 { return defaultModule.AscendingNodeN(date, n) }

// Azimuth 站心方位角，自正北向东量 / topocentric azimuth, measured from north through east.
func Azimuth(date time.Time, lon, lat float64) float64 { return defaultModule.Azimuth(date, lon, lat) }

// AzimuthN 站心方位角（n 变体） / n-parameterised topocentric azimuth.
func AzimuthN(date time.Time, lon, lat float64, n int) float64 {
	return defaultModule.AzimuthN(date, lon, lat, n)
}

// BrightLimbPositionAngle 亮面中心位置角，单位度 / bright limb position angle in degrees.
func BrightLimbPositionAngle(date time.Time) float64 {
	return defaultModule.BrightLimbPositionAngle(date)
}

// BrightLimbPositionAngleN 亮面中心位置角（n 变体），单位度 / n-parameterised bright limb position angle in degrees.
func BrightLimbPositionAngleN(date time.Time, n int) float64 {
	return defaultModule.BrightLimbPositionAngleN(date, n)
}

// ClosestTransit 最近一次全球水星凌日 / closest global Mercury transit.
func ClosestTransit(date time.Time) TransitInfo { return defaultModule.ClosestTransit(date) }

// CulminationTime 中天时刻，取 date 所在当地民用日 / culmination time, on the local civil day of date.
func CulminationTime(date time.Time, lon float64) time.Time {
	return defaultModule.CulminationTime(date, lon)
}

// CulminationTimeN 中天时间（n 变体） / n-parameterised culmination time.
func CulminationTimeN(date time.Time, lon float64, n int) time.Time {
	return defaultModule.CulminationTimeN(date, lon, n)
}

// DescendingNode 水星降交点黄经 / descending node longitude of Mercury.
func DescendingNode(date time.Time) float64 { return defaultModule.DescendingNode(date) }

// DescendingNodeN 水星降交点黄经（n 变体） / n-parameterised descending node longitude of Mercury.
func DescendingNodeN(date time.Time, n int) float64 { return defaultModule.DescendingNodeN(date, n) }

// Diameter 水星视直径，单位角秒 / apparent Mercury diameter in arcseconds.
func Diameter(date time.Time) float64 { return defaultModule.Diameter(date) }

// DiameterN 水星视直径（n 变体），单位角秒 / n-parameterised apparent Mercury diameter in arcseconds.
func DiameterN(date time.Time, n int) float64 { return defaultModule.DiameterN(date, n) }

// DownTime 落下时间别名 / deprecated set-time alias.
//
// Deprecated: use SetTime instead.
func DownTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	return defaultModule.DownTime(date, lon, lat, height, aero)
}

// DownTimeN 落下时间别名（n 变体） / n-parameterised down-time alias.
func DownTimeN(date time.Time, lon, lat, height float64, aero bool, n int) (time.Time, error) {
	return defaultModule.DownTimeN(date, lon, lat, height, aero, n)
}

// EarthDistance 地心距离 / Earth distance.
func EarthDistance(date time.Time) float64 { return defaultModule.EarthDistance(date) }

// EarthDistanceN 地球距离（n 变体） / n-parameterised Earth distance.
func EarthDistanceN(date time.Time, n int) float64 { return defaultModule.EarthDistanceN(date, n) }

// HourAngle 地心视时角，不含站心视差 / geocentric apparent hour angle without topocentric parallax.
func HourAngle(date time.Time, lon float64) float64 { return defaultModule.HourAngle(date, lon) }

// HourAngleN 时角（n 变体） / n-parameterised hour angle.
func HourAngleN(date time.Time, lon float64, n int) float64 {
	return defaultModule.HourAngleN(date, lon, n)
}

// IlluminatedFraction 被照亮比例 / illuminated fraction.
func IlluminatedFraction(date time.Time) float64 { return defaultModule.IlluminatedFraction(date) }

// IlluminatedFractionN 被照亮比例（n 变体） / n-parameterised illuminated fraction.
func IlluminatedFractionN(date time.Time, n int) float64 {
	return defaultModule.IlluminatedFractionN(date, n)
}

// LastConjunction 上一次合日，取 date 所在当地民用日 / previous conjunction with the Sun, on the local civil day of date.
func LastConjunction(date time.Time) time.Time { return defaultModule.LastConjunction(date) }

// LastGreatestElongation 上一次大距，取 date 所在当地民用日 / previous greatest elongation, on the local civil day of date.
func LastGreatestElongation(date time.Time) time.Time {
	return defaultModule.LastGreatestElongation(date)
}

// LastGreatestElongationEast 上一次东大距，取 date 所在当地民用日 / previous greatest eastern elongation, on the local civil day of date.
func LastGreatestElongationEast(date time.Time) time.Time {
	return defaultModule.LastGreatestElongationEast(date)
}

// LastGreatestElongationWest 上一次西大距，取 date 所在当地民用日 / previous greatest western elongation, on the local civil day of date.
func LastGreatestElongationWest(date time.Time) time.Time {
	return defaultModule.LastGreatestElongationWest(date)
}

// LastInferiorConjunction 上一次下合，取 date 所在当地民用日 / previous inferior conjunction, on the local civil day of date.
func LastInferiorConjunction(date time.Time) time.Time {
	return defaultModule.LastInferiorConjunction(date)
}

// LastProgradeToRetrograde 上一次顺行转逆行留，取 date 所在当地民用日 / previous station from prograde to retrograde, on the local civil day of date.
func LastProgradeToRetrograde(date time.Time) time.Time {
	return defaultModule.LastProgradeToRetrograde(date)
}

// LastRetrograde 上一次留，取 date 所在当地民用日 / previous stationary point, on the local civil day of date.
func LastRetrograde(date time.Time) time.Time { return defaultModule.LastRetrograde(date) }

// LastRetrogradeToPrograde 上一次逆行转顺行留，取 date 所在当地民用日 / previous station from retrograde to prograde, on the local civil day of date.
func LastRetrogradeToPrograde(date time.Time) time.Time {
	return defaultModule.LastRetrogradeToPrograde(date)
}

// LastSuperiorConjunction 上一次上合，取 date 所在当地民用日 / previous superior conjunction, on the local civil day of date.
func LastSuperiorConjunction(date time.Time) time.Time {
	return defaultModule.LastSuperiorConjunction(date)
}

// LastTransit 上一次全球水星凌日 / previous global Mercury transit.
func LastTransit(date time.Time) TransitInfo { return defaultModule.LastTransit(date) }

// LocalTransit 凌日在观测地的站心接触与凌甚 / topocentric contacts of a transit seen from the site.
func LocalTransit(transit TransitInfo, lon, lat, height float64) LocalTransitInfo {
	return defaultModule.LocalTransit(transit, lon, lat, height)
}

// NextConjunction 下一次合日，取 date 所在当地民用日 / next conjunction with the Sun, on the local civil day of date.
func NextConjunction(date time.Time) time.Time { return defaultModule.NextConjunction(date) }

// NextGreatestElongation 下一次大距，取 date 所在当地民用日 / next greatest elongation, on the local civil day of date.
func NextGreatestElongation(date time.Time) time.Time {
	return defaultModule.NextGreatestElongation(date)
}

// NextGreatestElongationEast 下一次东大距，取 date 所在当地民用日 / next greatest eastern elongation, on the local civil day of date.
func NextGreatestElongationEast(date time.Time) time.Time {
	return defaultModule.NextGreatestElongationEast(date)
}

// NextGreatestElongationWest 下一次西大距，取 date 所在当地民用日 / next greatest western elongation, on the local civil day of date.
func NextGreatestElongationWest(date time.Time) time.Time {
	return defaultModule.NextGreatestElongationWest(date)
}

// NextInferiorConjunction 下一次下合，取 date 所在当地民用日 / next inferior conjunction, on the local civil day of date.
func NextInferiorConjunction(date time.Time) time.Time {
	return defaultModule.NextInferiorConjunction(date)
}

// NextProgradeToRetrograde 下一次顺行转逆行留，取 date 所在当地民用日 / next station from prograde to retrograde, on the local civil day of date.
func NextProgradeToRetrograde(date time.Time) time.Time {
	return defaultModule.NextProgradeToRetrograde(date)
}

// NextRetrograde 下一次留，取 date 所在当地民用日 / next stationary point, on the local civil day of date.
func NextRetrograde(date time.Time) time.Time { return defaultModule.NextRetrograde(date) }

// NextRetrogradeToPrograde 下一次逆行转顺行留，取 date 所在当地民用日 / next station from retrograde to prograde, on the local civil day of date.
func NextRetrogradeToPrograde(date time.Time) time.Time {
	return defaultModule.NextRetrogradeToPrograde(date)
}

// NextSuperiorConjunction 下一次上合，取 date 所在当地民用日 / next superior conjunction, on the local civil day of date.
func NextSuperiorConjunction(date time.Time) time.Time {
	return defaultModule.NextSuperiorConjunction(date)
}

// NextTransit 下一次全球水星凌日 / next global Mercury transit.
func NextTransit(date time.Time) TransitInfo { return defaultModule.NextTransit(date) }

// ParallacticAngle 水星视差角（天顶方向角） / Mercury parallactic angle.
func ParallacticAngle(date time.Time, lon, lat float64) float64 {
	return defaultModule.ParallacticAngle(date, lon, lat)
}

// ParallacticAngleN 水星视差角（天顶方向角）（n 变体） / n-parameterised Mercury parallactic angle.
func ParallacticAngleN(date time.Time, lon, lat float64, n int) float64 {
	return defaultModule.ParallacticAngleN(date, lon, lat, n)
}

// Phase 相位，被照亮比例 / phase, illuminated fraction.
func Phase(date time.Time) float64 { return defaultModule.Phase(date) }

// PhaseAngle 相位角，单位度 / phase angle in degrees.
func PhaseAngle(date time.Time) float64 { return defaultModule.PhaseAngle(date) }

// PhaseAngleN 相位角（n 变体），单位度 / n-parameterised phase angle in degrees.
func PhaseAngleN(date time.Time, n int) float64 { return defaultModule.PhaseAngleN(date, n) }

// PhaseN 相位（n 变体），被照亮比例 / n-parameterised phase, illuminated fraction.
func PhaseN(date time.Time, n int) float64 { return defaultModule.PhaseN(date, n) }

// Physical 水星物理观测参数 / physical observing parameters of Mercury.
func Physical(date time.Time) PhysicalInfo { return defaultModule.Physical(date) }

// PhysicalN 水星物理观测参数（n 变体） / n-parameterised physical observing parameters of Mercury.
func PhysicalN(date time.Time, n int) PhysicalInfo { return defaultModule.PhysicalN(date, n) }

// RiseTime 升起时间，取 date 所在当地民用日 / rise time, on the local civil day of date.
func RiseTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	return defaultModule.RiseTime(date, lon, lat, height, aero)
}

// RiseTimeN 升起时间（n 变体） / n-parameterised rise time.
func RiseTimeN(date time.Time, lon, lat, height float64, aero bool, n int) (time.Time, error) {
	return defaultModule.RiseTimeN(date, lon, lat, height, aero, n)
}

// Semidiameter 水星视半径，单位角秒 / apparent Mercury semidiameter in arcseconds.
func Semidiameter(date time.Time) float64 { return defaultModule.Semidiameter(date) }

// SemidiameterN 水星视半径（n 变体），单位角秒 / n-parameterised apparent Mercury semidiameter in arcseconds.
func SemidiameterN(date time.Time, n int) float64 { return defaultModule.SemidiameterN(date, n) }

// SetTime 落下时间，取 date 所在当地民用日 / set time, on the local civil day of date.
func SetTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	return defaultModule.SetTime(date, lon, lat, height, aero)
}

// SetTimeN 落下时间（n 变体） / n-parameterised set time.
func SetTimeN(date time.Time, lon, lat, height float64, aero bool, n int) (time.Time, error) {
	return defaultModule.SetTimeN(date, lon, lat, height, aero, n)
}

// SunDistance 日心距离 / Sun distance.
func SunDistance(date time.Time) float64 { return defaultModule.SunDistance(date) }

// SunDistanceN 太阳距离（n 变体） / n-parameterised Sun distance.
func SunDistanceN(date time.Time, n int) float64 { return defaultModule.SunDistanceN(date, n) }

// Zenith 天顶距 / zenith distance.
func Zenith(date time.Time, lon, lat float64) float64 { return defaultModule.Zenith(date, lon, lat) }

// ZenithN 天顶距（n 变体） / n-parameterised zenith distance.
func ZenithN(date time.Time, lon, lat float64, n int) float64 {
	return defaultModule.ZenithN(date, lon, lat, n)
}

// ApparentMagnitude2018 视星等（Mallama & Hilton 2018 模型）/ apparent magnitude, Mallama & Hilton 2018 model.
func ApparentMagnitude2018(date time.Time) float64 { return defaultModule.ApparentMagnitude2018(date) }
