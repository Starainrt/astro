package moon

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// ApparentAltitude 月亮视高度角 / apparent lunar altitude.
func ApparentAltitude(date time.Time, lon, lat, pressureHPa, temperatureC float64) float64 {
	return basic.ApparentAltitude(Altitude(date, lon, lat), pressureHPa, temperatureC)
}

// ApparentZenith 月亮视天顶距 / apparent lunar zenith distance.
func ApparentZenith(date time.Time, lon, lat, pressureHPa, temperatureC float64) float64 {
	return 90 - ApparentAltitude(date, lon, lat, pressureHPa, temperatureC)
}
