package mars

import (
	"errors"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/calendar"
)

var (
	ERR_MARS_NEVER_RISE = errors.New("ERROR:极夜，火星今日永远在地平线下！")
	ERR_MARS_NEVER_SET  = errors.New("ERROR:极昼，火星今日永远在地平线上！")
	// ERR_MARS_NEVER_DOWN deprecated: -- use ERR_MARS_NEVER_SET instead
	ERR_MARS_NEVER_DOWN = ERR_MARS_NEVER_SET
)

func riseSetResult(date time.Time, jd float64, err error) (time.Time, error) {
	if err != nil {
		switch {
		case errors.Is(err, basic.ErrNeverRise):
			return time.Time{}, ERR_MARS_NEVER_RISE
		case errors.Is(err, basic.ErrNeverSet):
			return time.Time{}, ERR_MARS_NEVER_SET
		default:
			return time.Time{}, err
		}
	}
	_, offset := date.Zone()
	return basic.JD2DateByZone(jd-float64(offset)/86400, date.Location(), false), nil
}

// ApparentLo 视黄经 / apparent ecliptic longitude.
func (module *Mars) ApparentLo(date time.Time) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.MarsApparentLoWithInstance(basic.UTC2TT(jd), module.inst)
}

// ApparentBo 视黄纬 / apparent ecliptic latitude.
func (module *Mars) ApparentBo(date time.Time) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.MarsApparentBoWithInstance(basic.UTC2TT(jd), module.inst)
}

// ApparentRa 视赤经 / apparent right ascension.
func (module *Mars) ApparentRa(date time.Time) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.MarsApparentRaWithInstance(basic.UTC2TT(jd), module.inst)
}

// ApparentDec 视赤纬 / apparent declination.
func (module *Mars) ApparentDec(date time.Time) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.MarsApparentDecWithInstance(basic.UTC2TT(jd), module.inst)
}

// ApparentRaDec 视赤经、视赤纬 / apparent right ascension and declination.
func (module *Mars) ApparentRaDec(date time.Time) (float64, float64) {
	jd := calendar.Date2JD(date.UTC())
	return basic.MarsApparentRaDecWithInstance(basic.UTC2TT(jd), module.inst)
}

// ApparentMagnitude 视星等 / apparent magnitude.
func (module *Mars) ApparentMagnitude(date time.Time) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.MarsMagWithInstance(basic.UTC2TT(jd), module.inst)
}

// EarthDistance 地心距离 / Earth distance.
func (module *Mars) EarthDistance(date time.Time) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.EarthMarsAwayWithInstance(basic.UTC2TT(jd), module.inst)
}

// SunDistance 日心距离 / Sun distance.
func (module *Mars) SunDistance(date time.Time) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.PlanetCoordinateWithInstance(3, 2, basic.UTC2TT(jd), module.inst)
}

// Altitude 站心视高度角，不含大气折射 / topocentric apparent altitude without refraction.
func (module *Mars) Altitude(date time.Time, lon, lat float64) float64 {
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.MarsHeightWithInstance(localJD, lon, lat, timezone, module.inst)
}

// Zenith 天顶距 / zenith distance.
func (module *Mars) Zenith(date time.Time, lon, lat float64) float64 {
	return 90 - module.Altitude(date, lon, lat)
}

// Azimuth 站心方位角 / topocentric azimuth.
func (module *Mars) Azimuth(date time.Time, lon, lat float64) float64 {
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.MarsAzimuthWithInstance(localJD, lon, lat, timezone, module.inst)
}

// HourAngle 地心视时角，不含站心视差 / geocentric apparent hour angle without topocentric parallax.
func (module *Mars) HourAngle(date time.Time, lon float64) float64 {
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.MarsHourAngleWithInstance(localJD, lon, timezone, module.inst)
}

// CulminationTime 中天时刻 / culmination time.
func (module *Mars) CulminationTime(date time.Time, lon float64) time.Time {
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	calcJD := basic.MarsCulminationTimeWithInstance(localJD, lon, timezone, module.inst) - timezone/24.00
	return basic.JD2DateByZone(calcJD, date.Location(), false)
}

// RiseTime 升起时间 / rise time.
func (module *Mars) RiseTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	var aeroFloat float64
	if aero {
		aeroFloat = 1
	}
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	riseJD, err := basic.MarsRiseTimeWithInstance(localJD, lon, lat, timezone, aeroFloat, height, module.inst)
	return riseSetResult(date, riseJD, err)
}

// DownTime 落下时间别名 / deprecated set-time alias.
func (module *Mars) DownTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	return module.SetTime(date, lon, lat, height, aero)
}

// SetTime 落下时间 / set time.
func (module *Mars) SetTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	var aeroFloat float64
	if aero {
		aeroFloat = 1
	}
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	riseJD, err := basic.MarsSetTimeWithInstance(localJD, lon, lat, timezone, aeroFloat, height, module.inst)
	return riseSetResult(date, riseJD, err)
}

// LastConjunction 上一次合日 / previous conjunction with the Sun.
func (module *Mars) LastConjunction(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.LastMarsConjunctionWithInstance(jde, module.inst), date.Location(), false)
}

// NextConjunction 下一次合日 / next conjunction with the Sun.
func (module *Mars) NextConjunction(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.NextMarsConjunctionWithInstance(jde, module.inst), date.Location(), false)
}

// LastOpposition 上一次冲日 / previous opposition.
func (module *Mars) LastOpposition(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.LastMarsOppositionWithInstance(jde, module.inst), date.Location(), false)
}

// NextOpposition 下一次冲日 / next opposition.
func (module *Mars) NextOpposition(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.NextMarsOppositionWithInstance(jde, module.inst), date.Location(), false)
}

// LastProgradeToRetrograde 上一次顺行转逆行留 / previous station from prograde to retrograde.
func (module *Mars) LastProgradeToRetrograde(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.LastMarsProgradeToRetrogradeWithInstance(jde, module.inst), date.Location(), false)
}

// NextProgradeToRetrograde 下一次顺行转逆行留 / next station from prograde to retrograde.
func (module *Mars) NextProgradeToRetrograde(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.NextMarsProgradeToRetrogradeWithInstance(jde, module.inst), date.Location(), false)
}

// LastRetrogradeToPrograde 上一次逆行转顺行留 / previous station from retrograde to prograde.
func (module *Mars) LastRetrogradeToPrograde(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.LastMarsRetrogradeToProgradeWithInstance(jde, module.inst), date.Location(), false)
}

// NextRetrogradeToPrograde 下一次逆行转顺行留 / next station from retrograde to prograde.
func (module *Mars) NextRetrogradeToPrograde(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.NextMarsRetrogradeToProgradeWithInstance(jde, module.inst), date.Location(), false)
}

// LastEasternQuadrature 上一次东方照 / previous eastern quadrature.
func (module *Mars) LastEasternQuadrature(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.LastMarsEasternQuadratureWithInstance(jde, module.inst), date.Location(), false)
}

// NextEasternQuadrature 下一次东方照 / next eastern quadrature.
func (module *Mars) NextEasternQuadrature(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.NextMarsEasternQuadratureWithInstance(jde, module.inst), date.Location(), false)
}

// LastWesternQuadrature 上一次西方照 / previous western quadrature.
func (module *Mars) LastWesternQuadrature(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.LastMarsWesternQuadratureWithInstance(jde, module.inst), date.Location(), false)
}

// NextWesternQuadrature 下一次西方照 / next western quadrature.
func (module *Mars) NextWesternQuadrature(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.NextMarsWesternQuadratureWithInstance(jde, module.inst), date.Location(), false)
}

// ApparentMagnitude2018 视星等（Mallama & Hilton 2018 模型）/ apparent magnitude, Mallama & Hilton 2018 model.
func (module *Mars) ApparentMagnitude2018(date time.Time) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.MarsMag2018WithInstance(basic.UTC2TT(jd), module.inst)
}
