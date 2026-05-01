package sun

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// ApparentAltitude 太阳视高度角 / apparent solar altitude.
func ApparentAltitude(date time.Time, lon, lat, pressureHPa, temperatureC float64) float64 {
	return basic.ApparentAltitude(Altitude(date, lon, lat), pressureHPa, temperatureC)
}

// ApparentZenith 太阳视天顶距 / apparent solar zenith distance.
func ApparentZenith(date time.Time, lon, lat, pressureHPa, temperatureC float64) float64 {
	return 90 - ApparentAltitude(date, lon, lat, pressureHPa, temperatureC)
}

// ApparentAltitudeN 太阳视高度角（截断版） / truncated apparent solar altitude.
func ApparentAltitudeN(date time.Time, lon, lat, pressureHPa, temperatureC float64, n int) float64 {
	return basic.ApparentAltitude(AltitudeN(date, lon, lat, n), pressureHPa, temperatureC)
}

// ApparentZenithN 太阳视天顶距（截断版） / truncated apparent solar zenith distance.
func ApparentZenithN(date time.Time, lon, lat, pressureHPa, temperatureC float64, n int) float64 {
	return 90 - ApparentAltitudeN(date, lon, lat, pressureHPa, temperatureC, n)
}
