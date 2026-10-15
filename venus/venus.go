package venus

import (
	"errors"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/calendar"
)

var (
	ERR_VENUS_NEVER_RISE = errors.New("ERROR:极夜，金星今日永远在地平线下！")
	ERR_VENUS_NEVER_SET  = errors.New("ERROR:极昼，金星今日永远在地平线上！")
	// ERR_VENUS_NEVER_DOWN deprecated: -- use ERR_VENUS_NEVER_SET instead
	ERR_VENUS_NEVER_DOWN = ERR_VENUS_NEVER_SET
)

func riseSetResult(date time.Time, jd float64, err error) (time.Time, error) {
	if err != nil {
		switch {
		case errors.Is(err, basic.ErrNeverRise):
			return time.Time{}, ERR_VENUS_NEVER_RISE
		case errors.Is(err, basic.ErrNeverSet):
			return time.Time{}, ERR_VENUS_NEVER_SET
		default:
			return time.Time{}, err
		}
	}
	_, offset := date.Zone()
	return basic.JD2DateByZone(jd-float64(offset)/86400, date.Location(), false), nil
}

// ApparentLo 视黄经 / apparent ecliptic longitude.
func (module *Venus) ApparentLo(date time.Time) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.VenusApparentLoWithInstance(basic.UTC2TT(jd), module.inst)
}

// ApparentBo 视黄纬 / apparent ecliptic latitude.
func (module *Venus) ApparentBo(date time.Time) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.VenusApparentBoWithInstance(basic.UTC2TT(jd), module.inst)
}

// ApparentRa 视赤经 / apparent right ascension.
func (module *Venus) ApparentRa(date time.Time) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.VenusApparentRaWithInstance(basic.UTC2TT(jd), module.inst)
}

// ApparentDec 视赤纬 / apparent declination.
func (module *Venus) ApparentDec(date time.Time) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.VenusApparentDecWithInstance(basic.UTC2TT(jd), module.inst)
}

// ApparentRaDec 视赤经、视赤纬 / apparent right ascension and declination.
func (module *Venus) ApparentRaDec(date time.Time) (float64, float64) {
	jd := calendar.Date2JD(date.UTC())
	return basic.VenusApparentRaDecWithInstance(basic.UTC2TT(jd), module.inst)
}

// ApparentMagnitude 视星等 / apparent magnitude.
func (module *Venus) ApparentMagnitude(date time.Time) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.VenusMagWithInstance(basic.UTC2TT(jd), module.inst)
}

// EarthDistance 地心距离 / Earth distance.
func (module *Venus) EarthDistance(date time.Time) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.EarthVenusAwayWithInstance(basic.UTC2TT(jd), module.inst)
}

// SunDistance 日心距离 / Sun distance.
func (module *Venus) SunDistance(date time.Time) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.PlanetCoordinateWithInstance(2, 2, basic.UTC2TT(jd), module.inst)
}

// Altitude 站心视高度角，不含大气折射 / topocentric apparent altitude without refraction.
func (module *Venus) Altitude(date time.Time, lon, lat float64) float64 {
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.VenusHeightWithInstance(localJD, lon, lat, timezone, module.inst)
}

// Zenith 天顶距 / zenith distance.
func (module *Venus) Zenith(date time.Time, lon, lat float64) float64 {
	return 90 - module.Altitude(date, lon, lat)
}

// Azimuth 站心方位角 / topocentric azimuth.
func (module *Venus) Azimuth(date time.Time, lon, lat float64) float64 {
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.VenusAzimuthWithInstance(localJD, lon, lat, timezone, module.inst)
}

// HourAngle 地心视时角，不含站心视差 / geocentric apparent hour angle without topocentric parallax.
func (module *Venus) HourAngle(date time.Time, lon float64) float64 {
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.VenusHourAngleWithInstance(localJD, lon, timezone, module.inst)
}

// CulminationTime 中天时刻 / culmination time.
func (module *Venus) CulminationTime(date time.Time, lon float64) time.Time {
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	calcJD := basic.VenusCulminationTimeWithInstance(localJD, lon, timezone, module.inst) - timezone/24.00
	return basic.JD2DateByZone(calcJD, date.Location(), false)
}

// RiseTime 升起时间 / rise time.
func (module *Venus) RiseTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	var aeroFloat float64
	if aero {
		aeroFloat = 1
	}
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	riseJD, err := basic.VenusRiseTimeWithInstance(localJD, lon, lat, timezone, aeroFloat, height, module.inst)
	return riseSetResult(date, riseJD, err)
}

// DownTime 落下时间别名 / deprecated set-time alias.
func (module *Venus) DownTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	return module.SetTime(date, lon, lat, height, aero)
}

// SetTime 落下时间 / set time.
func (module *Venus) SetTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	var aeroFloat float64
	if aero {
		aeroFloat = 1
	}
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	riseJD, err := basic.VenusSetTimeWithInstance(localJD, lon, lat, timezone, aeroFloat, height, module.inst)
	return riseSetResult(date, riseJD, err)
}

// LastConjunction 上一次合日 / previous conjunction with the Sun.
func (module *Venus) LastConjunction(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.LastVenusConjunctionWithInstance(jde, module.inst), date.Location(), false)
}

// NextConjunction 下一次合日 / next conjunction with the Sun.
func (module *Venus) NextConjunction(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.NextVenusConjunctionWithInstance(jde, module.inst), date.Location(), false)
}

// LastInferiorConjunction 上一次下合 / previous inferior conjunction.
func (module *Venus) LastInferiorConjunction(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.LastVenusInferiorConjunctionInclusiveWithInstance(jde, module.inst), date.Location(), false)
}

// NextInferiorConjunction 下一次下合 / next inferior conjunction.
func (module *Venus) NextInferiorConjunction(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.NextVenusInferiorConjunctionInclusiveWithInstance(jde, module.inst), date.Location(), false)
}

// LastSuperiorConjunction 上一次上合 / previous superior conjunction.
func (module *Venus) LastSuperiorConjunction(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.LastVenusSuperiorConjunctionInclusiveWithInstance(jde, module.inst), date.Location(), false)
}

// NextSuperiorConjunction 下一次上合 / next superior conjunction.
func (module *Venus) NextSuperiorConjunction(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.NextVenusSuperiorConjunctionInclusiveWithInstance(jde, module.inst), date.Location(), false)
}

// LastRetrograde 上一次留 / previous stationary point.
func (module *Venus) LastRetrograde(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.LastVenusRetrogradeInclusiveWithInstance(jde, module.inst), date.Location(), false)
}

// NextRetrograde 下一次留 / next stationary point.
func (module *Venus) NextRetrograde(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.NextVenusRetrogradeInclusiveWithInstance(jde, module.inst), date.Location(), false)
}

// LastProgradeToRetrograde 上一次顺行转逆行留 / previous station from prograde to retrograde.
func (module *Venus) LastProgradeToRetrograde(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.LastVenusProgradeToRetrogradeInclusiveWithInstance(jde, module.inst), date.Location(), false)
}

// NextProgradeToRetrograde 下一次顺行转逆行留 / next station from prograde to retrograde.
func (module *Venus) NextProgradeToRetrograde(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.NextVenusProgradeToRetrogradeInclusiveWithInstance(jde, module.inst), date.Location(), false)
}

// LastRetrogradeToPrograde 上一次逆行转顺行留 / previous station from retrograde to prograde.
func (module *Venus) LastRetrogradeToPrograde(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.LastVenusRetrogradeToProgradeInclusiveWithInstance(jde, module.inst), date.Location(), false)
}

// NextRetrogradeToPrograde 下一次逆行转顺行留 / next station from retrograde to prograde.
func (module *Venus) NextRetrogradeToPrograde(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.NextVenusRetrogradeToProgradeInclusiveWithInstance(jde, module.inst), date.Location(), false)
}

// LastGreatestElongation 上一次大距 / previous greatest elongation.
func (module *Venus) LastGreatestElongation(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.LastVenusGreatestElongationInclusiveWithInstance(jde, module.inst), date.Location(), false)
}

// NextGreatestElongation 下一次大距 / next greatest elongation.
func (module *Venus) NextGreatestElongation(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.NextVenusGreatestElongationInclusiveWithInstance(jde, module.inst), date.Location(), false)
}

// LastGreatestElongationEast 上一次东大距 / previous greatest eastern elongation.
func (module *Venus) LastGreatestElongationEast(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.LastVenusGreatestElongationEastInclusiveWithInstance(jde, module.inst), date.Location(), false)
}

// NextGreatestElongationEast 下一次东大距 / next greatest eastern elongation.
func (module *Venus) NextGreatestElongationEast(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.NextVenusGreatestElongationEastInclusiveWithInstance(jde, module.inst), date.Location(), false)
}

// LastGreatestElongationWest 上一次西大距 / previous greatest western elongation.
func (module *Venus) LastGreatestElongationWest(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.LastVenusGreatestElongationWestInclusiveWithInstance(jde, module.inst), date.Location(), false)
}

// NextGreatestElongationWest 下一次西大距 / next greatest western elongation.
func (module *Venus) NextGreatestElongationWest(date time.Time) time.Time {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.NextVenusGreatestElongationWestInclusiveWithInstance(jde, module.inst), date.Location(), false)
}

// ApparentMagnitude2018 视星等（Mallama & Hilton 2018 模型）/ apparent magnitude, Mallama & Hilton 2018 model.
func (module *Venus) ApparentMagnitude2018(date time.Time) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.VenusMag2018WithInstance(basic.UTC2TT(jd), module.inst)
}
