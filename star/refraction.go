package star

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// ApparentAltitude 恒星视高度角 / apparent stellar altitude.
func ApparentAltitude(date time.Time, ra, dec, lon, lat, pressureHPa, temperatureC float64) float64 {
	return basic.ApparentAltitude(Altitude(date, ra, dec, lon, lat), pressureHPa, temperatureC)
}

// ApparentZenith 恒星视天顶距 / apparent stellar zenith distance.
func ApparentZenith(date time.Time, ra, dec, lon, lat, pressureHPa, temperatureC float64) float64 {
	return 90 - ApparentAltitude(date, ra, dec, lon, lat, pressureHPa, temperatureC)
}
