package sun

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// ApparentAltitude 太阳视高度角 / apparent solar altitude.
func (module *Sun) ApparentAltitude(date time.Time, lon, lat, pressureHPa, temperatureC float64) float64 {
	return basic.ApparentAltitude(module.Altitude(date, lon, lat), pressureHPa, temperatureC)
}

// ApparentZenith 太阳视天顶距 / apparent solar zenith distance.
func (module *Sun) ApparentZenith(date time.Time, lon, lat, pressureHPa, temperatureC float64) float64 {
	return 90 - module.ApparentAltitude(date, lon, lat, pressureHPa, temperatureC)
}

// ApparentAltitudeN 太阳视高度角（n 变体） / n-parameterised apparent solar altitude.
func (module *Sun) ApparentAltitudeN(date time.Time, lon, lat, pressureHPa, temperatureC float64, n int) float64 {
	return basic.ApparentAltitude(module.AltitudeN(date, lon, lat, n), pressureHPa, temperatureC)
}

// ApparentZenithN 太阳视天顶距（n 变体） / n-parameterised apparent solar zenith distance.
func (module *Sun) ApparentZenithN(date time.Time, lon, lat, pressureHPa, temperatureC float64, n int) float64 {
	return 90 - module.ApparentAltitudeN(date, lon, lat, pressureHPa, temperatureC, n)
}
