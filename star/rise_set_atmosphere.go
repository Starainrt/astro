package star

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// RiseTimeWithAtmosphere 恒星升起时刻，按给定气压（hPa）与气温（°C）计算大气折射；其余参数同 RiseTime / stellar rise time with explicit pressure and temperature.
func RiseTimeWithAtmosphere(date time.Time, ra, dec, lon, lat, height, pressureHPa, temperatureC float64) (time.Time, error) {
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	_, loc := date.Zone()
	jd, err := basic.StarRiseTimeWithAtmosphere(basic.Date2JD(date), ra, dec, lon, lat, float64(loc)/3600.0, height, pressureHPa, temperatureC)
	return riseSetResult(date, jd, err)
}

// SetTimeWithAtmosphere 恒星落下时刻，按给定气压（hPa）与气温（°C）计算大气折射；其余参数同 SetTime / stellar set time with explicit pressure and temperature.
func SetTimeWithAtmosphere(date time.Time, ra, dec, lon, lat, height, pressureHPa, temperatureC float64) (time.Time, error) {
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	_, loc := date.Zone()
	jd, err := basic.StarSetTimeWithAtmosphere(basic.Date2JD(date), ra, dec, lon, lat, float64(loc)/3600.0, height, pressureHPa, temperatureC)
	return riseSetResult(date, jd, err)
}
