package moon

import (
	"errors"
	"math"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/civiltime"
	"github.com/starainrt/astro/tools"
)

var (
	ERR_MOON_NEVER_RISE = errors.New("ERROR:极夜，月亮在今日永远在地平线下！")
	ERR_MOON_NEVER_SET  = errors.New("ERROR:极昼，月亮在今日永远在地平线上！")
	// ERR_MOON_NEVER_DOWN deprecated: -- use ERR_MOON_NEVER_SET instead
	ERR_MOON_NEVER_DOWN = ERR_MOON_NEVER_SET
	ERR_NOT_TODAY       = errors.New("ERROR:月亮已在（昨日/明日）（升起/降下）")
)

func riseSetResult(date time.Time, jd float64, err error) (time.Time, error) {
	if err != nil {
		switch {
		case errors.Is(err, basic.ErrNotOnThisDate):
			return time.Time{}, ERR_NOT_TODAY
		case errors.Is(err, basic.ErrNeverRise):
			return time.Time{}, ERR_MOON_NEVER_RISE
		case errors.Is(err, basic.ErrNeverSet):
			return time.Time{}, ERR_MOON_NEVER_SET
		default:
			return time.Time{}, err
		}
	}
	_, offset := date.Zone()
	return basic.JD2DateByZone(jd-float64(offset)/86400, date.Location(), false), nil
}

// TrueLo 月亮真黄经 / true ecliptic longitude.
func (module *Moon) TrueLo(date time.Time) float64 {
	jd := basic.Date2JD(date.UTC())
	return basic.HMoonTrueLoWithInstance(basic.UTC2TT(jd), module.inst)
}

// TrueLoN 月亮真黄经（n 变体） / n-parameterised true ecliptic longitude.
func (module *Moon) TrueLoN(date time.Time, n int) float64 {
	jd := basic.Date2JD(date.UTC())
	return basic.HMoonTrueLoNWithInstance(basic.UTC2TT(jd), n, module.inst)
}

// TrueBo 月亮真黄纬 / true ecliptic latitude.
func (module *Moon) TrueBo(date time.Time) float64 {
	jd := basic.Date2JD(date.UTC())
	return basic.HMoonTrueBoWithInstance(basic.UTC2TT(jd), module.inst)
}

// TrueBoN 月亮真黄纬（n 变体） / n-parameterised true ecliptic latitude.
func (module *Moon) TrueBoN(date time.Time, n int) float64 {
	jd := basic.Date2JD(date.UTC())
	return basic.HMoonTrueBoNWithInstance(basic.UTC2TT(jd), n, module.inst)
}

// ApparentLo 月亮地心视黄经 / apparent geocentric ecliptic longitude.
func (module *Moon) ApparentLo(date time.Time) float64 {
	jd := basic.Date2JD(date.UTC())
	return basic.HMoonApparentLoWithInstance(basic.UTC2TT(jd), module.inst)
}

// TrueRa 月亮地心真赤经 / true geocentric right ascension.
func (module *Moon) TrueRa(date time.Time) float64 {
	jd := basic.Date2JD(date.UTC())
	return basic.HMoonGeocentricTrueRaWithInstance(basic.UTC2TT(jd), module.inst)
}

// TrueDec 月亮地心真赤纬 / true geocentric declination.
func (module *Moon) TrueDec(date time.Time) float64 {
	jd := basic.Date2JD(date.UTC())
	return basic.HMoonGeocentricTrueDecWithInstance(basic.UTC2TT(jd), module.inst)
}

// TrueRaDec 月亮地心真赤经、真赤纬 / true geocentric right ascension and declination.
func (module *Moon) TrueRaDec(date time.Time) (float64, float64) {
	jd := basic.Date2JD(date.UTC())
	return basic.HMoonGeocentricTrueRaDecWithInstance(basic.UTC2TT(jd), module.inst)
}

// GeocentricApparentRa 月亮地心视赤经 / apparent geocentric right ascension.
func (module *Moon) GeocentricApparentRa(date time.Time) float64 {
	jd := basic.Date2JD(date.UTC())
	return basic.HMoonGeocentricApparentRaWithInstance(basic.UTC2TT(jd), module.inst)
}

// GeocentricApparentDec 月亮地心视赤纬 / apparent geocentric declination.
func (module *Moon) GeocentricApparentDec(date time.Time) float64 {
	jd := basic.Date2JD(date.UTC())
	return basic.HMoonGeocentricApparentDecWithInstance(basic.UTC2TT(jd), module.inst)
}

// GeocentricApparentRaDec 月亮地心视赤经、视赤纬 / apparent geocentric right ascension and declination.
func (module *Moon) GeocentricApparentRaDec(date time.Time) (float64, float64) {
	jd := basic.Date2JD(date.UTC())
	return basic.HMoonGeocentricApparentRaDecWithInstance(basic.UTC2TT(jd), module.inst)
}

// ApparentRa 月亮站心视赤经 / apparent topocentric right ascension.
func (module *Moon) ApparentRa(date time.Time, lon, lat float64) float64 {
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	return basic.HMoonApparentRaWithInstance(localJD, lon, lat, float64(loc)/3600.0, module.inst)
}

// ApparentDec 月亮站心视赤纬 / apparent topocentric declination.
func (module *Moon) ApparentDec(date time.Time, lon, lat float64) float64 {
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	return basic.HMoonApparentDecWithInstance(localJD, lon, lat, float64(loc)/3600.0, module.inst)
}

// ApparentRaDec 月亮站心视赤经、视赤纬 / apparent topocentric right ascension and declination.
func (module *Moon) ApparentRaDec(date time.Time, lon, lat float64) (float64, float64) {
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	return basic.HMoonApparentRaDecWithInstance(localJD, lon, lat, float64(loc)/3600.0, module.inst)
}

// HourAngle 月亮时角 / hour angle.
func (module *Moon) HourAngle(date time.Time, lon, lat float64) float64 {
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	return basic.MoonTimeAngleWithInstance(localJD, lon, lat, float64(loc)/3600.0, module.inst)
}

// Azimuth 月亮方位角 / azimuth.
func (module *Moon) Azimuth(date time.Time, lon, lat float64) float64 {
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	return basic.HMoonAzimuthWithInstance(localJD, lon, lat, float64(loc)/3600.0, module.inst)
}

// Altitude 月亮高度角 / lunar altitude.
func (module *Moon) Altitude(date time.Time, lon, lat float64) float64 {
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	return basic.HMoonHeightWithInstance(localJD, lon, lat, float64(loc)/3600.0, module.inst)
}

// Zenith 月亮天顶距 / lunar zenith distance.
func (module *Moon) Zenith(date time.Time, lon, lat float64) float64 {
	return 90 - module.Altitude(date, lon, lat)
}

// CulminationTime 月亮中天时刻 / culmination time.
func (module *Moon) CulminationTime(date time.Time, lon, lat float64) time.Time {
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	calcJD := basic.MoonCulminationTimeWithInstance(localJD, lon, lat, float64(loc)/3600.0, module.inst) - float64(loc)/86400
	return basic.JD2DateByZone(calcJD, date.Location(), false)
}

// RiseTime 月出时刻 / moonrise time.
func (module *Moon) RiseTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	if result, err, handled := civiltime.Event(date, ERR_NOT_TODAY, func(d time.Time) (time.Time, error) {
		return module.RiseTime(d, lon, lat, height, aero)
	}); handled {
		return result, err
	}
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	aeroFloat := 0.00
	if aero {
		aeroFloat = 1
	}
	riseJD, err := basic.GetMoonRiseTimeWithInstance(localJD, lon, lat, timezone, aeroFloat, height, module.inst)
	return riseSetResult(date, riseJD, err)
}

// DownTime 月落时刻别名 / deprecated moonset alias.
func (module *Moon) DownTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	return module.SetTime(date, lon, lat, height, aero)
}

// SetTime 月落时刻 / moonset time.
func (module *Moon) SetTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	if result, err, handled := civiltime.Event(date, ERR_NOT_TODAY, func(d time.Time) (time.Time, error) {
		return module.SetTime(d, lon, lat, height, aero)
	}); handled {
		return result, err
	}
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	aeroFloat := 0.00
	if aero {
		aeroFloat = 1
	}
	downJD, err := basic.GetMoonSetTimeWithInstance(localJD, lon, lat, timezone, aeroFloat, height, module.inst)
	return riseSetResult(date, downJD, err)
}

// SunMoonLoDiff 日月黄经差 / Moon-Sun longitude difference.
func (module *Moon) SunMoonLoDiff(date time.Time) float64 {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	sunLo := basic.HSunApparentLoWithInstance(jde, module.inst)
	moonLo := basic.HMoonApparentLoWithInstance(jde, module.inst)
	return tools.Limit360(moonLo - sunLo)
}

// PhaseDesc 月相文字描述 / textual lunar phase description.
func (module *Moon) PhaseDesc(date time.Time) string {
	moonSunLoDiff := module.SunMoonLoDiff(date)
	if moonSunLoDiff >= 0 && moonSunLoDiff <= 30 {
		return "新月"
	} else if moonSunLoDiff > 30 && moonSunLoDiff <= 75 {
		return "上峨眉月"
	} else if moonSunLoDiff > 75 && moonSunLoDiff <= 135 {
		return "上弦月"
	} else if moonSunLoDiff > 135 && moonSunLoDiff < 170 {
		return "盈凸月"
	} else if moonSunLoDiff >= 170 && moonSunLoDiff <= 190 {
		return "满月"
	} else if moonSunLoDiff > 190 && moonSunLoDiff < 225 {
		return "亏凸月"
	} else if moonSunLoDiff >= 225 && moonSunLoDiff < 285 {
		return "下弦月"
	} else if moonSunLoDiff >= 285 && moonSunLoDiff < 330 {
		return "下峨眉月"
	} else {
		return "残月"
	}
}

// Phase 月面受照比例 / illuminated fraction.
func (module *Moon) Phase(date time.Time) float64 {
	jd := basic.Date2JD(date.UTC())
	return basic.MoonPhaseWithInstance(basic.UTC2TT(jd), module.inst)
}

// ShuoYue 朔月锚点解 / new-moon solution near a decimal year anchor.
func (module *Moon) ShuoYue(year float64) time.Time {
	jd := basic.TT2UTC(basic.CalcMoonSHWithInstance(year, 0, module.inst))
	return basic.JD2DateByZone(jd, time.UTC, false)
}

// NextShuoYue 下一次朔月 / next new moon.
func (module *Moon) NextShuoYue(date time.Time) time.Time {
	return nextMoonPhase(date, 0, module)
}

// LastShuoYue 上一次朔月 / previous new moon.
func (module *Moon) LastShuoYue(date time.Time) time.Time {
	return lastMoonPhase(date, 0, module)
}

// ClosestShuoYue 最近朔月 / closest new moon.
func (module *Moon) ClosestShuoYue(date time.Time) time.Time {
	return closestMoonPhase(date, 0, module)
}

// NewMoon 朔月英文别名 / English alias for ShuoYue.
func (module *Moon) NewMoon(year float64) time.Time {
	return module.ShuoYue(year)
}

// NextNewMoon 下一次朔月英文别名 / English alias for NextShuoYue.
func (module *Moon) NextNewMoon(date time.Time) time.Time {
	return module.NextShuoYue(date)
}

// LastNewMoon 上一次朔月英文别名 / English alias for LastShuoYue.
func (module *Moon) LastNewMoon(date time.Time) time.Time {
	return module.LastShuoYue(date)
}

// ClosestNewMoon 最近朔月英文别名 / English alias for ClosestShuoYue.
func (module *Moon) ClosestNewMoon(date time.Time) time.Time {
	return module.ClosestShuoYue(date)
}

func closestMoonPhase(date time.Time, typed int, modules ...*Moon) time.Time {
	module := &defaultModule
	if len(modules) > 0 {
		module = modules[0]
	}

	//0=shuo 1=wang 2=shangxian 3=xiaxian
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	if typed < 2 {
		return basic.JD2DateByZone(basic.TT2UTC(basic.CalcMoonSHByJDEWithInstance(jde, typed, module.inst)), date.Location(), false)
	}
	return basic.JD2DateByZone(basic.TT2UTC(basic.CalcMoonXHByJDEWithInstance(jde, typed-2, module.inst)), date.Location(), false)
}

func nextMoonPhase(date time.Time, typed int, modules ...*Moon) time.Time {
	module := &defaultModule
	if len(modules) > 0 {
		module = modules[0]
	}

	//0=shuo 1=wang 2=shangxian 3=xiaxian
	diffCode := 0.00
	switch typed {
	case 1:
		diffCode = 180
	case 2:
		diffCode = 90
	case 3:
		diffCode = 270
	}
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	cost := basic.HMoonApparentLoWithInstance(jde, module.inst) - basic.HSunApparentLoWithInstance(jde, module.inst) - float64(diffCode)
	for cost < 0 {
		cost += 360
	}
	if cost < 0 && math.Floor(math.Abs(cost)*10000) == 0 {
		cost = 0
	}
	if cost < 240 {
		jde += (240 - cost) / 11.19
	}
	if typed < 2 {
		return basic.JD2DateByZone(basic.TT2UTC(basic.CalcMoonSHByJDEWithInstance(jde, typed, module.inst)), date.Location(), false)
	}
	return basic.JD2DateByZone(basic.TT2UTC(basic.CalcMoonXHByJDEWithInstance(jde, typed-2, module.inst)), date.Location(), false)
}

func lastMoonPhase(date time.Time, typed int, modules ...*Moon) time.Time {
	module := &defaultModule
	if len(modules) > 0 {
		module = modules[0]
	}

	//0=shuo 1=wang 2=shangxian 3=xiaxian
	diffCode := 0.00
	switch typed {
	case 1:
		diffCode = 180
	case 2:
		diffCode = 90
	case 3:
		diffCode = 270
	}
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	cost := basic.HMoonApparentLoWithInstance(jde, module.inst) - basic.HSunApparentLoWithInstance(jde, module.inst) - float64(diffCode)
	for cost < 0 {
		cost += 360
	}
	if cost > 0 && math.Floor(math.Abs(cost)*10000) == 0 {
		cost = 360
	}
	if cost > 120 {
		jde -= (cost - 120) / 11.19
	}
	if typed < 2 {
		return basic.JD2DateByZone(basic.TT2UTC(basic.CalcMoonSHByJDEWithInstance(jde, typed, module.inst)), date.Location(), false)
	}
	return basic.JD2DateByZone(basic.TT2UTC(basic.CalcMoonXHByJDEWithInstance(jde, typed-2, module.inst)), date.Location(), false)
}

// WangYue 望月锚点解 / full-moon solution near a decimal year anchor.
func (module *Moon) WangYue(year float64) time.Time {
	jd := basic.TT2UTC(basic.CalcMoonSHWithInstance(year, 1, module.inst))
	return basic.JD2DateByZone(jd, time.UTC, false)
}

// NextWangYue 下一次望月 / next full moon.
func (module *Moon) NextWangYue(date time.Time) time.Time {
	return nextMoonPhase(date, 1, module)
}

// LastWangYue 上一次望月 / previous full moon.
func (module *Moon) LastWangYue(date time.Time) time.Time {
	return lastMoonPhase(date, 1, module)
}

// ClosestWangYue 最近望月 / closest full moon.
func (module *Moon) ClosestWangYue(date time.Time) time.Time {
	return closestMoonPhase(date, 1, module)
}

// FullMoon 望月英文别名 / English alias for WangYue.
func (module *Moon) FullMoon(year float64) time.Time {
	return module.WangYue(year)
}

// NextFullMoon 下一次望月英文别名 / English alias for NextWangYue.
func (module *Moon) NextFullMoon(date time.Time) time.Time {
	return module.NextWangYue(date)
}

// LastFullMoon 上一次望月英文别名 / English alias for LastWangYue.
func (module *Moon) LastFullMoon(date time.Time) time.Time {
	return module.LastWangYue(date)
}

// ClosestFullMoon 最近望月英文别名 / English alias for ClosestWangYue.
func (module *Moon) ClosestFullMoon(date time.Time) time.Time {
	return module.ClosestWangYue(date)
}

// ShangXianYue 上弦锚点解 / first-quarter solution near a decimal year anchor.
func (module *Moon) ShangXianYue(year float64) time.Time {
	jd := basic.TT2UTC(basic.CalcMoonXHWithInstance(year, 0, module.inst))
	return basic.JD2DateByZone(jd, time.UTC, false)
}

// NextShangXianYue 下一次上弦 / next first quarter.
func (module *Moon) NextShangXianYue(date time.Time) time.Time {
	return nextMoonPhase(date, 2, module)
}

// LastShangXianYue 上一次上弦 / previous first quarter.
func (module *Moon) LastShangXianYue(date time.Time) time.Time {
	return lastMoonPhase(date, 2, module)
}

// ClosestShangXianYue 最近上弦 / closest first quarter.
func (module *Moon) ClosestShangXianYue(date time.Time) time.Time {
	return closestMoonPhase(date, 2, module)
}

// FirstQuarter 上弦英文别名 / English alias for ShangXianYue.
func (module *Moon) FirstQuarter(year float64) time.Time {
	return module.ShangXianYue(year)
}

// NextFirstQuarter 下一次上弦英文别名 / English alias for NextShangXianYue.
func (module *Moon) NextFirstQuarter(date time.Time) time.Time {
	return module.NextShangXianYue(date)
}

// LastFirstQuarter 上一次上弦英文别名 / English alias for LastShangXianYue.
func (module *Moon) LastFirstQuarter(date time.Time) time.Time {
	return module.LastShangXianYue(date)
}

// ClosestFirstQuarter 最近上弦英文别名 / English alias for ClosestShangXianYue.
func (module *Moon) ClosestFirstQuarter(date time.Time) time.Time {
	return module.ClosestShangXianYue(date)
}

// XiaXianYue 下弦锚点解 / last-quarter solution near a decimal year anchor.
func (module *Moon) XiaXianYue(year float64) time.Time {
	jd := basic.TT2UTC(basic.CalcMoonXHWithInstance(year, 1, module.inst))
	return basic.JD2DateByZone(jd, time.UTC, false)
}

// NextXiaXianYue 下一次下弦 / next last quarter.
func (module *Moon) NextXiaXianYue(date time.Time) time.Time {
	return nextMoonPhase(date, 3, module)
}

// LastXiaXianYue 上一次下弦 / previous last quarter.
func (module *Moon) LastXiaXianYue(date time.Time) time.Time {
	return lastMoonPhase(date, 3, module)
}

// ClosestXiaXianYue 最近下弦 / closest last quarter.
func (module *Moon) ClosestXiaXianYue(date time.Time) time.Time {
	return closestMoonPhase(date, 3, module)
}

// LastQuarter 下弦英文别名 / English alias for XiaXianYue.
func (module *Moon) LastQuarter(year float64) time.Time {
	return module.XiaXianYue(year)
}

// NextLastQuarter 下一次下弦英文别名 / English alias for NextXiaXianYue.
func (module *Moon) NextLastQuarter(date time.Time) time.Time {
	return module.NextXiaXianYue(date)
}

// LastLastQuarter 上一次下弦英文别名 / English alias for LastXiaXianYue.
func (module *Moon) LastLastQuarter(date time.Time) time.Time {
	return module.LastXiaXianYue(date)
}

// ClosestLastQuarter 最近下弦英文别名 / English alias for ClosestXiaXianYue.
func (module *Moon) ClosestLastQuarter(date time.Time) time.Time {
	return module.ClosestXiaXianYue(date)
}

// RiseTimeWithAtmosphere 月出时刻，按给定气压（hPa）与气温（°C）计算大气折射 / moonrise time with explicit pressure and temperature.
func (module *Moon) RiseTimeWithAtmosphere(date time.Time, lon, lat, height, pressureHPa, temperatureC float64) (time.Time, error) {
	if result, err, handled := civiltime.Event(date, ERR_NOT_TODAY, func(d time.Time) (time.Time, error) {
		return module.RiseTimeWithAtmosphere(d, lon, lat, height, pressureHPa, temperatureC)
	}); handled {
		return result, err
	}
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	jd, err := basic.GetMoonRiseTimeWithAtmosphere(localJD, lon, lat, float64(loc)/3600.0, height, pressureHPa, temperatureC, module.inst)
	return riseSetResult(date, jd, err)
}

// SetTimeWithAtmosphere 月落时刻，按给定气压（hPa）与气温（°C）计算大气折射 / moonset time with explicit pressure and temperature.
func (module *Moon) SetTimeWithAtmosphere(date time.Time, lon, lat, height, pressureHPa, temperatureC float64) (time.Time, error) {
	if result, err, handled := civiltime.Event(date, ERR_NOT_TODAY, func(d time.Time) (time.Time, error) {
		return module.SetTimeWithAtmosphere(d, lon, lat, height, pressureHPa, temperatureC)
	}); handled {
		return result, err
	}
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	jd, err := basic.GetMoonSetTimeWithAtmosphere(localJD, lon, lat, float64(loc)/3600.0, height, pressureHPa, temperatureC, module.inst)
	return riseSetResult(date, jd, err)
}
