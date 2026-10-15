package saturn

import (
	"errors"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/calendar"
)

var (
	ERR_SATURN_NEVER_RISE = errors.New("ERROR:极夜，土星今日永远在地平线下！")
	ERR_SATURN_NEVER_SET  = errors.New("ERROR:极昼，土星今日永远在地平线上！")
	// ERR_SATURN_NEVER_DOWN deprecated: -- use ERR_SATURN_NEVER_SET instead
	ERR_SATURN_NEVER_DOWN = ERR_SATURN_NEVER_SET
)

func riseSetResult(date time.Time, jd float64, err error) (time.Time, error) {
	if err != nil {
		switch {
		case errors.Is(err, basic.ErrNeverRise):
			return time.Time{}, ERR_SATURN_NEVER_RISE
		case errors.Is(err, basic.ErrNeverSet):
			return time.Time{}, ERR_SATURN_NEVER_SET
		default:
			return time.Time{}, err
		}
	}
	_, offset := date.Zone()
	return basic.JD2DateByZone(jd-float64(offset)/86400, date.Location(), false), nil
}

// ApparentLo 视黄经 / apparent ecliptic longitude.
func (module *Saturn) ApparentLo(date time.Time) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.SaturnApparentLoWithInstance(basic.UTC2TT(jd), module.inst)
}

// ApparentBo 视黄纬 / apparent ecliptic latitude.
func (module *Saturn) ApparentBo(date time.Time) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.SaturnApparentBoWithInstance(basic.UTC2TT(jd), module.inst)
}

// ApparentRa 视赤经 / apparent right ascension.
func (module *Saturn) ApparentRa(date time.Time) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.SaturnApparentRaWithInstance(basic.UTC2TT(jd), module.inst)
}

// ApparentDec 视赤纬 / apparent declination.
func (module *Saturn) ApparentDec(date time.Time) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.SaturnApparentDecWithInstance(basic.UTC2TT(jd), module.inst)
}

// ApparentRaDec 视赤经、视赤纬 / apparent right ascension and declination.
func (module *Saturn) ApparentRaDec(date time.Time) (float64, float64) {
	jd := calendar.Date2JD(date.UTC())
	return basic.SaturnApparentRaDecWithInstance(basic.UTC2TT(jd), module.inst)
}

// ApparentMagnitude 视星等 / apparent magnitude.
func (module *Saturn) ApparentMagnitude(date time.Time) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.SaturnMagWithInstance(basic.UTC2TT(jd), module.inst)
}

// EarthDistance 地心距离 / Earth distance.
func (module *Saturn) EarthDistance(date time.Time) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.EarthSaturnAwayWithInstance(basic.UTC2TT(jd), module.inst)
}

// SunDistance 日心距离 / Sun distance.
func (module *Saturn) SunDistance(date time.Time) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.PlanetCoordinateWithInstance(5, 2, basic.UTC2TT(jd), module.inst)
}

// Altitude 站心视高度角，不含大气折射 / topocentric apparent altitude without refraction.
func (module *Saturn) Altitude(date time.Time, lon, lat float64) float64 {
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.SaturnHeightWithInstance(localJD, lon, lat, timezone, module.inst)
}

// Zenith 天顶距 / zenith distance.
func (module *Saturn) Zenith(date time.Time, lon, lat float64) float64 {
	return 90 - module.Altitude(date, lon, lat)
}

// Azimuth 站心方位角 / topocentric azimuth.
func (module *Saturn) Azimuth(date time.Time, lon, lat float64) float64 {
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.SaturnAzimuthWithInstance(localJD, lon, lat, timezone, module.inst)
}

// HourAngle 地心视时角，不含站心视差 / geocentric apparent hour angle without topocentric parallax.
func (module *Saturn) HourAngle(date time.Time, lon float64) float64 {
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.SaturnHourAngleWithInstance(localJD, lon, timezone, module.inst)
}

// CulminationTime 中天时刻 / culmination time.
func (module *Saturn) CulminationTime(date time.Time, lon float64) time.Time {
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	calcJD := basic.SaturnCulminationTimeWithInstance(localJD, lon, timezone, module.inst) - timezone/24.00
	return basic.JD2DateByZone(calcJD, date.Location(), false)
}

// RiseTime 升起时间 / rise time.
func (module *Saturn) RiseTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	var aeroFloat float64
	if aero {
		aeroFloat = 1
	}
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	riseJD, err := basic.SaturnRiseTimeWithInstance(localJD, lon, lat, timezone, aeroFloat, height, module.inst)
	return riseSetResult(date, riseJD, err)
}

// DownTime 落下时间别名 / deprecated set-time alias.
func (module *Saturn) DownTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	return module.SetTime(date, lon, lat, height, aero)
}

// SetTime 落下时间 / set time.
func (module *Saturn) SetTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	var aeroFloat float64
	if aero {
		aeroFloat = 1
	}
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	riseJD, err := basic.SaturnSetTimeWithInstance(localJD, lon, lat, timezone, aeroFloat, height, module.inst)
	return riseSetResult(date, riseJD, err)
}

// LastConjunction 上一次合日 / previous conjunction with the Sun.
func (module *Saturn) LastConjunction(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.LastSaturnConjunctionWithInstance(jde, module.inst), date.Location(), false)
}

// NextConjunction 下一次合日 / next conjunction with the Sun.
func (module *Saturn) NextConjunction(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.NextSaturnConjunctionWithInstance(jde, module.inst), date.Location(), false)
}

// LastOpposition 上一次冲日 / previous opposition.
func (module *Saturn) LastOpposition(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.LastSaturnOppositionWithInstance(jde, module.inst), date.Location(), false)
}

// NextOpposition 下一次冲日 / next opposition.
func (module *Saturn) NextOpposition(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.NextSaturnOppositionWithInstance(jde, module.inst), date.Location(), false)
}

// LastProgradeToRetrograde 上一次顺行转逆行留 / previous station from prograde to retrograde.
func (module *Saturn) LastProgradeToRetrograde(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.LastSaturnProgradeToRetrogradeWithInstance(jde, module.inst), date.Location(), false)
}

// NextProgradeToRetrograde 下一次顺行转逆行留 / next station from prograde to retrograde.
func (module *Saturn) NextProgradeToRetrograde(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.NextSaturnProgradeToRetrogradeWithInstance(jde, module.inst), date.Location(), false)
}

// LastRetrogradeToPrograde 上一次逆行转顺行留 / previous station from retrograde to prograde.
func (module *Saturn) LastRetrogradeToPrograde(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.LastSaturnRetrogradeToProgradeWithInstance(jde, module.inst), date.Location(), false)
}

// NextRetrogradeToPrograde 下一次逆行转顺行留 / next station from retrograde to prograde.
func (module *Saturn) NextRetrogradeToPrograde(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.NextSaturnRetrogradeToProgradeWithInstance(jde, module.inst), date.Location(), false)
}

// LastEasternQuadrature 上一次东方照 / previous eastern quadrature.
func (module *Saturn) LastEasternQuadrature(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.LastSaturnEasternQuadratureWithInstance(jde, module.inst), date.Location(), false)
}

// NextEasternQuadrature 下一次东方照 / next eastern quadrature.
func (module *Saturn) NextEasternQuadrature(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.NextSaturnEasternQuadratureWithInstance(jde, module.inst), date.Location(), false)
}

// LastWesternQuadrature 上一次西方照 / previous western quadrature.
func (module *Saturn) LastWesternQuadrature(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.LastSaturnWesternQuadratureWithInstance(jde, module.inst), date.Location(), false)
}

// NextWesternQuadrature 下一次西方照 / next western quadrature.
func (module *Saturn) NextWesternQuadrature(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.NextSaturnWesternQuadratureWithInstance(jde, module.inst), date.Location(), false)
}

// ApparentMagnitude2018 视星等（Mallama & Hilton 2018 模型）/ apparent magnitude, Mallama & Hilton 2018 model.
func (module *Saturn) ApparentMagnitude2018(date time.Time) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.SaturnMag2018WithInstance(basic.UTC2TT(jd), module.inst)
}
