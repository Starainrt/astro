package sun

import (
	"time"
)

// Altitude 太阳站心视高度角，不含大气折射 / topocentric apparent solar altitude without refraction.
func Altitude(date time.Time, lon, lat float64) float64 {
	return defaultModule.Altitude(date, lon, lat)
}

// AltitudeN 太阳站心视高度角（n 变体） / n-parameterised topocentric apparent solar altitude.
func AltitudeN(date time.Time, lon, lat float64, n int) float64 {
	return defaultModule.AltitudeN(date, lon, lat, n)
}

// ApparentAltitude 太阳视高度角 / apparent solar altitude.
func ApparentAltitude(date time.Time, lon, lat, pressureHPa, temperatureC float64) float64 {
	return defaultModule.ApparentAltitude(date, lon, lat, pressureHPa, temperatureC)
}

// ApparentAltitudeN 太阳视高度角（n 变体） / n-parameterised apparent solar altitude.
func ApparentAltitudeN(date time.Time, lon, lat, pressureHPa, temperatureC float64, n int) float64 {
	return defaultModule.ApparentAltitudeN(date, lon, lat, pressureHPa, temperatureC, n)
}

// ApparentDec 太阳地心视赤纬 / apparent geocentric declination.
func ApparentDec(date time.Time) float64 { return defaultModule.ApparentDec(date) }

// ApparentLo 太阳视黄经 / apparent ecliptic longitude.
func ApparentLo(date time.Time) float64 { return defaultModule.ApparentLo(date) }

// ApparentRa 太阳地心视赤经 / apparent geocentric right ascension.
func ApparentRa(date time.Time) float64 { return defaultModule.ApparentRa(date) }

// ApparentRaDec 太阳地心视赤经、视赤纬 / apparent geocentric right ascension and declination.
func ApparentRaDec(date time.Time) (float64, float64) { return defaultModule.ApparentRaDec(date) }

// ApparentSolarTime 真太阳时，结果时区为按经度换算的地方平太阳时 / apparent solar time, result is a longitude-derived mean-solar zone.
func ApparentSolarTime(date time.Time, lon float64) time.Time {
	return defaultModule.ApparentSolarTime(date, lon)
}

// ApparentSolarTimeN 真太阳时（n 变体） / n-parameterised apparent solar time.
func ApparentSolarTimeN(date time.Time, lon float64, n int) time.Time {
	return defaultModule.ApparentSolarTimeN(date, lon, n)
}

// ApparentZenith 太阳视天顶距 / apparent solar zenith distance.
func ApparentZenith(date time.Time, lon, lat, pressureHPa, temperatureC float64) float64 {
	return defaultModule.ApparentZenith(date, lon, lat, pressureHPa, temperatureC)
}

// ApparentZenithN 太阳视天顶距（n 变体） / n-parameterised apparent solar zenith distance.
func ApparentZenithN(date time.Time, lon, lat, pressureHPa, temperatureC float64, n int) float64 {
	return defaultModule.ApparentZenithN(date, lon, lat, pressureHPa, temperatureC, n)
}

// Azimuth 太阳站心方位角，自正北向东量 / topocentric solar azimuth, measured from north through east.
func Azimuth(date time.Time, lon, lat float64) float64 { return defaultModule.Azimuth(date, lon, lat) }

// AzimuthN 太阳站心方位角（n 变体） / n-parameterised topocentric solar azimuth.
func AzimuthN(date time.Time, lon, lat float64, n int) float64 {
	return defaultModule.AzimuthN(date, lon, lat, n)
}

// CulminationTime 太阳中天时刻，取 date 所在当地民用日 / culmination time, on the local civil day of date.
func CulminationTime(date time.Time, lon float64) time.Time {
	return defaultModule.CulminationTime(date, lon)
}

// CulminationTimeN 太阳中天时刻（n 变体） / n-parameterised culmination time.
func CulminationTimeN(date time.Time, lon float64, n int) time.Time {
	return defaultModule.CulminationTimeN(date, lon, n)
}

// Diameter 太阳视直径，单位角秒 / apparent solar diameter in arcseconds.
func Diameter(date time.Time) float64 { return defaultModule.Diameter(date) }

// DiameterN 太阳视直径（n 变体），单位角秒 / n-parameterised apparent solar diameter in arcseconds.
func DiameterN(date time.Time, n int) float64 { return defaultModule.DiameterN(date, n) }

// DownTime 日落时刻别名 / deprecated sunset alias.
//
// Deprecated: use SetTime instead.
func DownTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	return defaultModule.DownTime(date, lon, lat, height, aero)
}

// DownTimeN 日落时刻别名（n 变体） / deprecated n-parameterised sunset alias.
//
// Deprecated: use SetTimeN instead.
func DownTimeN(date time.Time, lon, lat, height float64, aero bool, n int) (time.Time, error) {
	return defaultModule.DownTimeN(date, lon, lat, height, aero, n)
}

// EarthDistance 日地距离 / Earth-Sun distance.
func EarthDistance(date time.Time) float64 { return defaultModule.EarthDistance(date) }

// EquationTime 均时差 / equation of time.
func EquationTime(date time.Time) float64 { return defaultModule.EquationTime(date) }

// EveningTwilight 暮光终时，取 date 所在当地民用日 / evening twilight, on the local civil day of date.
func EveningTwilight(date time.Time, lon, lat, angle float64) (time.Time, error) {
	return defaultModule.EveningTwilight(date, lon, lat, angle)
}

// EveningTwilightN 暮光终时（n 变体），取 date 所在当地民用日 / n-parameterised evening twilight, on the local civil day of date.
func EveningTwilightN(date time.Time, lon, lat, angle float64, n int) (time.Time, error) {
	return defaultModule.EveningTwilightN(date, lon, lat, angle, n)
}

// HourAngle 太阳地心视时角，不含站心视差，lat 不参与计算仅为参数形状一致 / geocentric apparent solar hour angle without topocentric parallax, lat is unused, kept for signature symmetry.
func HourAngle(date time.Time, lon, lat float64) float64 {
	return defaultModule.HourAngle(date, lon, lat)
}

// HourAngleN 太阳时角（n 变体） / n-parameterised hour angle.
func HourAngleN(date time.Time, lon, lat float64, n int) float64 {
	return defaultModule.HourAngleN(date, lon, lat, n)
}

// MorningTwilight 晨光始时，angle 常取 -6/-12/-18 度 / morning twilight, angle typically -6/-12/-18.
func MorningTwilight(date time.Time, lon, lat, angle float64) (time.Time, error) {
	return defaultModule.MorningTwilight(date, lon, lat, angle)
}

// MorningTwilightN 晨光始时（n 变体） / n-parameterised morning twilight.
func MorningTwilightN(date time.Time, lon, lat, angle float64, n int) (time.Time, error) {
	return defaultModule.MorningTwilightN(date, lon, lat, angle, n)
}

// ParallacticAngle 太阳视差角（天顶方向角） / solar parallactic angle.
func ParallacticAngle(date time.Time, lon, lat float64) float64 {
	return defaultModule.ParallacticAngle(date, lon, lat)
}

// ParallacticAngleN 太阳视差角（天顶方向角）（n 变体） / n-parameterised solar parallactic angle.
func ParallacticAngleN(date time.Time, lon, lat float64, n int) float64 {
	return defaultModule.ParallacticAngleN(date, lon, lat, n)
}

// Physical 太阳物理观测参数 / physical observing parameters of the Sun.
func Physical(date time.Time) PhysicalInfo { return defaultModule.Physical(date) }

// PhysicalN 太阳物理观测参数（n 变体） / n-parameterised physical observing parameters of the Sun.
func PhysicalN(date time.Time, n int) PhysicalInfo { return defaultModule.PhysicalN(date, n) }

// RiseTime 日出时刻，取 date 所在当地民用日；aero 为 true 时按上缘过地平线 / sunrise time, on the local civil day of date; aero uses an upper-limb crossing.
func RiseTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	return defaultModule.RiseTime(date, lon, lat, height, aero)
}

// RiseTimeN 日出时刻（n 变体） / n-parameterised sunrise time.
func RiseTimeN(date time.Time, lon, lat, height float64, aero bool, n int) (time.Time, error) {
	return defaultModule.RiseTimeN(date, lon, lat, height, aero, n)
}

// Semidiameter 太阳视半径，单位角秒 / apparent solar semidiameter in arcseconds.
func Semidiameter(date time.Time) float64 { return defaultModule.Semidiameter(date) }

// SemidiameterN 太阳视半径（n 变体），单位角秒 / n-parameterised apparent solar semidiameter in arcseconds.
func SemidiameterN(date time.Time, n int) float64 { return defaultModule.SemidiameterN(date, n) }

// SetTime 日落时刻，取 date 所在当地民用日 / sunset time, on the local civil day of date.
func SetTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	return defaultModule.SetTime(date, lon, lat, height, aero)
}

// SetTimeN 日落时刻（n 变体），取 date 所在当地民用日 / n-parameterised sunset time, on the local civil day of date.
func SetTimeN(date time.Time, lon, lat, height float64, aero bool, n int) (time.Time, error) {
	return defaultModule.SetTimeN(date, lon, lat, height, aero, n)
}

// TrueBo 太阳真黄纬 / true ecliptic latitude.
func TrueBo(date time.Time) float64 { return defaultModule.TrueBo(date) }

// TrueBoN 太阳真黄纬（n 变体） / n-parameterised true ecliptic latitude.
func TrueBoN(date time.Time, n int) float64 { return defaultModule.TrueBoN(date, n) }

// TrueLo 太阳真黄经 / true ecliptic longitude.
func TrueLo(date time.Time) float64 { return defaultModule.TrueLo(date) }

// TrueLoN 太阳真黄经（n 变体） / n-parameterised true ecliptic longitude.
func TrueLoN(date time.Time, n int) float64 { return defaultModule.TrueLoN(date, n) }

// Zenith 太阳天顶距 / solar zenith distance.
func Zenith(date time.Time, lon, lat float64) float64 { return defaultModule.Zenith(date, lon, lat) }

// ZenithN 太阳天顶距（n 变体） / n-parameterised solar zenith distance.
func ZenithN(date time.Time, lon, lat float64, n int) float64 {
	return defaultModule.ZenithN(date, lon, lat, n)
}

// RiseTimeWithAtmosphere 日出时刻，按给定气压（hPa）与气温（°C）计算大气折射 / sunrise time with explicit pressure and temperature.
func RiseTimeWithAtmosphere(date time.Time, lon, lat, height, pressureHPa, temperatureC float64) (time.Time, error) {
	return defaultModule.RiseTimeWithAtmosphere(date, lon, lat, height, pressureHPa, temperatureC)
}

// SetTimeWithAtmosphere 日落时刻，按给定气压（hPa）与气温（°C）计算大气折射 / sunset time with explicit pressure and temperature.
func SetTimeWithAtmosphere(date time.Time, lon, lat, height, pressureHPa, temperatureC float64) (time.Time, error) {
	return defaultModule.SetTimeWithAtmosphere(date, lon, lat, height, pressureHPa, temperatureC)
}
