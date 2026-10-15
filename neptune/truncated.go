package neptune

import (
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/calendar"
)

// N variants keep the same semantics as the non-N APIs; n < 0 means full series.

// ApparentLoN 视黄经（n 变体） / n-parameterised apparent ecliptic longitude.
func (module *Neptune) ApparentLoN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.NeptuneApparentLoNWithInstance(basic.UTC2TT(jd), n, module.inst)
}

// ApparentBoN 视黄纬（n 变体） / n-parameterised apparent ecliptic latitude.
func (module *Neptune) ApparentBoN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.NeptuneApparentBoNWithInstance(basic.UTC2TT(jd), n, module.inst)
}

// ApparentRaN 视赤经（n 变体） / n-parameterised apparent right ascension.
func (module *Neptune) ApparentRaN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.NeptuneApparentRaNWithInstance(basic.UTC2TT(jd), n, module.inst)
}

// ApparentDecN 视赤纬（n 变体） / n-parameterised apparent declination.
func (module *Neptune) ApparentDecN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.NeptuneApparentDecNWithInstance(basic.UTC2TT(jd), n, module.inst)
}

// ApparentRaDecN 视赤经赤纬（n 变体） / n-parameterised apparent right ascension and declination.
func (module *Neptune) ApparentRaDecN(date time.Time, n int) (float64, float64) {
	jd := calendar.Date2JD(date.UTC())
	return basic.NeptuneApparentRaDecNWithInstance(basic.UTC2TT(jd), n, module.inst)
}

// ApparentMagnitudeN 视星等（n 变体） / n-parameterised apparent magnitude.
func (module *Neptune) ApparentMagnitudeN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.NeptuneMagNWithInstance(basic.UTC2TT(jd), n, module.inst)
}

// EarthDistanceN 地球距离（n 变体） / n-parameterised Earth distance.
func (module *Neptune) EarthDistanceN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.EarthNeptuneAwayNWithInstance(basic.UTC2TT(jd), n, module.inst)
}

// SunDistanceN 太阳距离（n 变体） / n-parameterised Sun distance.
func (module *Neptune) SunDistanceN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.PlanetCoordinateNWithInstance(7, 2, basic.UTC2TT(jd), n, module.inst)
}

// AltitudeN 站心视高度角（n 变体） / n-parameterised topocentric apparent altitude.
func (module *Neptune) AltitudeN(date time.Time, lon, lat float64, n int) float64 {
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.NeptuneHeightNWithInstance(localJD, lon, lat, timezone, n, module.inst)
}

// ZenithN 天顶距（n 变体） / n-parameterised zenith distance.
func (module *Neptune) ZenithN(date time.Time, lon, lat float64, n int) float64 {
	return 90 - module.AltitudeN(date, lon, lat, n)
}

// AzimuthN 站心方位角（n 变体） / n-parameterised topocentric azimuth.
func (module *Neptune) AzimuthN(date time.Time, lon, lat float64, n int) float64 {
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.NeptuneAzimuthNWithInstance(localJD, lon, lat, timezone, n, module.inst)
}

// HourAngleN 时角（n 变体） / n-parameterised hour angle.
func (module *Neptune) HourAngleN(date time.Time, lon float64, n int) float64 {
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.NeptuneHourAngleNWithInstance(localJD, lon, timezone, n, module.inst)
}

// CulminationTimeN 中天时间（n 变体） / n-parameterised culmination time.
func (module *Neptune) CulminationTimeN(date time.Time, lon float64, n int) time.Time {
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	calcJD := basic.NeptuneCulminationTimeNWithInstance(localJD, lon, timezone, n, module.inst) - timezone/24.0
	return basic.JD2DateByZone(calcJD, date.Location(), false)
}

// RiseTimeN 升起时间（n 变体） / n-parameterised rise time.
func (module *Neptune) RiseTimeN(date time.Time, lon, lat, height float64, aero bool, n int) (time.Time, error) {
	var aeroFloat float64
	if aero {
		aeroFloat = 1
	}
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	riseJD, err := basic.NeptuneRiseTimeNWithInstance(localJD, lon, lat, timezone, aeroFloat, height, n, module.inst)
	return riseSetResult(date, riseJD, err)
}

// DownTimeN 落下时间别名（n 变体） / n-parameterised down-time alias.
func (module *Neptune) DownTimeN(date time.Time, lon, lat, height float64, aero bool, n int) (time.Time, error) {
	return module.SetTimeN(date, lon, lat, height, aero, n)
}

// SetTimeN 落下时间（n 变体） / n-parameterised set time.
func (module *Neptune) SetTimeN(date time.Time, lon, lat, height float64, aero bool, n int) (time.Time, error) {
	var aeroFloat float64
	if aero {
		aeroFloat = 1
	}
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	riseJD, err := basic.NeptuneSetTimeNWithInstance(localJD, lon, lat, timezone, aeroFloat, height, n, module.inst)
	return riseSetResult(date, riseJD, err)
}
