package coord

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// AtmosphericRefractionFromTrueAltitude 真高度角折射修正 / atmospheric refraction from true altitude.
//
// 输入真高度角，返回应加到真高度角上的大气折射修正量，单位度。
func AtmosphericRefractionFromTrueAltitude(trueAltitude, pressureHPa, temperatureC float64) float64 {
	return basic.RefractionFromTrueAltitude(trueAltitude, pressureHPa, temperatureC)
}

// AtmosphericRefractionFromApparentAltitude 视高度角折射修正 / atmospheric refraction from apparent altitude.
//
// 输入视高度角，返回对应的大气折射修正量，单位度。
func AtmosphericRefractionFromApparentAltitude(apparentAltitude, pressureHPa, temperatureC float64) float64 {
	return basic.RefractionFromApparentAltitude(apparentAltitude, pressureHPa, temperatureC)
}

// ApparentAltitude 真高度角转视高度角 / apparent altitude from true altitude.
//
// 输入真高度角，返回加入标准大气折射后的视高度角，单位度。
func ApparentAltitude(trueAltitude, pressureHPa, temperatureC float64) float64 {
	return basic.ApparentAltitude(trueAltitude, pressureHPa, temperatureC)
}

// TrueAltitude 视高度角转真高度角 / true altitude from apparent altitude.
//
// 输入视高度角，返回去除大气折射后的真高度角，单位度。
func TrueAltitude(apparentAltitude, pressureHPa, temperatureC float64) float64 {
	return basic.TrueAltitude(apparentAltitude, pressureHPa, temperatureC)
}

// EquatorialToApparentHorizontal 赤道坐标转视地平坐标 / converts equatorial coordinates to apparent horizontal coordinates.
func EquatorialToApparentHorizontal(date time.Time, ra, dec, observerLon, observerLat, pressureHPa, temperatureC float64) Horizontal {
	horizontal := EquatorialToHorizontal(date, ra, dec, observerLon, observerLat)
	horizontal.Altitude = basic.ApparentAltitude(horizontal.Altitude, pressureHPa, temperatureC)
	horizontal.Zenith = 90 - horizontal.Altitude
	return horizontal
}
