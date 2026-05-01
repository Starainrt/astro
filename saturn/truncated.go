package saturn

import (
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/calendar"
	"github.com/starainrt/astro/planet"
)

// N variants keep the same semantics as the non-N APIs; n < 0 means full series.

// ApparentLoN 视黄经（截断版） / truncated apparent ecliptic longitude.
func ApparentLoN(date time.Time, n int) float64 {
	jde := calendar.Date2JDE(date.UTC())
	return basic.SaturnApparentLoN(basic.TD2UT(jde, true), n)
}

// ApparentBoN 视黄纬（截断版） / truncated apparent ecliptic latitude.
func ApparentBoN(date time.Time, n int) float64 {
	jde := calendar.Date2JDE(date.UTC())
	return basic.SaturnApparentBoN(basic.TD2UT(jde, true), n)
}

// ApparentRaN 视赤经（截断版） / truncated apparent right ascension.
func ApparentRaN(date time.Time, n int) float64 {
	jde := calendar.Date2JDE(date.UTC())
	return basic.SaturnApparentRaN(basic.TD2UT(jde, true), n)
}

// ApparentDecN 视赤纬（截断版） / truncated apparent declination.
func ApparentDecN(date time.Time, n int) float64 {
	jde := calendar.Date2JDE(date.UTC())
	return basic.SaturnApparentDecN(basic.TD2UT(jde, true), n)
}

// ApparentRaDecN 视赤经赤纬（截断版） / truncated apparent right ascension and declination.
func ApparentRaDecN(date time.Time, n int) (float64, float64) {
	jde := calendar.Date2JDE(date.UTC())
	return basic.SaturnApparentRaDecN(basic.TD2UT(jde, true), n)
}

// ApparentMagnitudeN 视星等（截断版） / truncated apparent magnitude.
func ApparentMagnitudeN(date time.Time, n int) float64 {
	jde := calendar.Date2JDE(date.UTC())
	return basic.SaturnMagN(basic.TD2UT(jde, true), n)
}

// EarthDistanceN 地球距离（截断版） / truncated Earth distance.
func EarthDistanceN(date time.Time, n int) float64 {
	jde := calendar.Date2JDE(date.UTC())
	return basic.EarthSaturnAwayN(basic.TD2UT(jde, true), n)
}

// SunDistanceN 太阳距离（截断版） / truncated Sun distance.
func SunDistanceN(date time.Time, n int) float64 {
	jde := calendar.Date2JDE(date.UTC())
	return planet.WherePlanetN(5, 2, basic.TD2UT(jde, true), n)
}

// AltitudeN 高度角（截断版） / truncated altitude angle.
func AltitudeN(date time.Time, lon, lat float64, n int) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.SaturnHeightN(jde, lon, lat, timezone, n)
}

// ZenithN 天顶距（截断版） / truncated zenith distance.
func ZenithN(date time.Time, lon, lat float64, n int) float64 {
	return 90 - AltitudeN(date, lon, lat, n)
}

// AzimuthN 方位角（截断版） / truncated azimuth angle.
func AzimuthN(date time.Time, lon, lat float64, n int) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.SaturnAzimuthN(jde, lon, lat, timezone, n)
}

// HourAngleN 时角（截断版） / truncated hour angle.
func HourAngleN(date time.Time, lon float64, n int) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.SaturnHourAngleN(jde, lon, timezone, n)
}

// CulminationTimeN 中天时间（截断版） / truncated culmination time.
func CulminationTimeN(date time.Time, lon float64, n int) time.Time {
	if date.Hour() > 12 {
		date = date.Add(time.Hour * -12)
	}
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	calcJde := basic.SaturnCulminationTimeN(jde, lon, timezone, n) - timezone/24.0
	return basic.JDE2DateByZone(calcJde, date.Location(), false)
}

// RiseTimeN 升起时间（截断版） / truncated rise time.
func RiseTimeN(date time.Time, lon, lat, height float64, aero bool, n int) (time.Time, error) {
	var aeroFloat float64
	if aero {
		aeroFloat = 1
	}
	if date.Hour() > 12 {
		date = date.Add(time.Hour * -12)
	}
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	riseJde, err := basic.SaturnRiseTimeN(jde, lon, lat, timezone, aeroFloat, height, n)
	return riseSetResult(date, riseJde, err)
}

// DownTimeN 落下时间别名（截断版） / truncated down-time alias.
func DownTimeN(date time.Time, lon, lat, height float64, aero bool, n int) (time.Time, error) {
	return SetTimeN(date, lon, lat, height, aero, n)
}

// SetTimeN 落下时间（截断版） / truncated set time.
func SetTimeN(date time.Time, lon, lat, height float64, aero bool, n int) (time.Time, error) {
	var aeroFloat float64
	if aero {
		aeroFloat = 1
	}
	if date.Hour() > 12 {
		date = date.Add(time.Hour * -12)
	}
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	riseJde, err := basic.SaturnSetTimeN(jde, lon, lat, timezone, aeroFloat, height, n)
	return riseSetResult(date, riseJde, err)
}
