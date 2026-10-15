package mars

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// RiseTimeWithAtmosphere 火星升起时刻，按给定气压（hPa）与气温（°C）计算大气折射 / rise time with explicit pressure and temperature.
func (module *Mars) RiseTimeWithAtmosphere(date time.Time, lon, lat, height, pressureHPa, temperatureC float64) (time.Time, error) {
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	_, loc := date.Zone()
	jd, err := basic.MarsRiseTimeWithAtmosphere(basic.Date2JD(date), lon, lat, float64(loc)/3600.0, height, pressureHPa, temperatureC, module.inst)
	return riseSetResult(date, jd, err)
}

// SetTimeWithAtmosphere 火星落下时刻，按给定气压（hPa）与气温（°C）计算大气折射 / set time with explicit pressure and temperature.
func (module *Mars) SetTimeWithAtmosphere(date time.Time, lon, lat, height, pressureHPa, temperatureC float64) (time.Time, error) {
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	_, loc := date.Zone()
	jd, err := basic.MarsSetTimeWithAtmosphere(basic.Date2JD(date), lon, lat, float64(loc)/3600.0, height, pressureHPa, temperatureC, module.inst)
	return riseSetResult(date, jd, err)
}

// RiseTimeWithAtmosphere 火星升起时刻，按给定气压（hPa）与气温（°C）计算大气折射 / rise time with explicit pressure and temperature.
func RiseTimeWithAtmosphere(date time.Time, lon, lat, height, pressureHPa, temperatureC float64) (time.Time, error) {
	return defaultModule.RiseTimeWithAtmosphere(date, lon, lat, height, pressureHPa, temperatureC)
}

// SetTimeWithAtmosphere 火星落下时刻，按给定气压（hPa）与气温（°C）计算大气折射 / set time with explicit pressure and temperature.
func SetTimeWithAtmosphere(date time.Time, lon, lat, height, pressureHPa, temperatureC float64) (time.Time, error) {
	return defaultModule.SetTimeWithAtmosphere(date, lon, lat, height, pressureHPa, temperatureC)
}
