package moon

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// ApparentAltitude 月亮视高度角 / apparent lunar altitude.
func (module *Moon) ApparentAltitude(date time.Time, lon, lat, pressureHPa, temperatureC float64) float64 {
	return basic.ApparentAltitude(module.Altitude(date, lon, lat), pressureHPa, temperatureC)
}

// ApparentZenith 月亮视天顶距 / apparent lunar zenith distance.
func (module *Moon) ApparentZenith(date time.Time, lon, lat, pressureHPa, temperatureC float64) float64 {
	return 90 - module.ApparentAltitude(date, lon, lat, pressureHPa, temperatureC)
}
