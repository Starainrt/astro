package sun

import (
	"errors"
	"math"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/civiltime"
)

var (
	ERR_SUN_NEVER_RISE = errors.New("ERROR:极夜，太阳在今日永远在地平线下！")
	ERR_SUN_NEVER_SET  = errors.New("ERROR:极昼，太阳在今日永远在地平线上！")
	// ERR_SUN_NEVER_DOWN deprecated: -- use ERR_SUN_NEVER_RISE instead
	ERR_SUN_NEVER_DOWN      = ERR_SUN_NEVER_SET
	ERR_TWILIGHT_NOT_EXISTS = errors.New("ERROR:今日晨昏朦影不存在！")
)

func riseSetResult(date time.Time, jd float64, err error) (time.Time, error) {
	if err != nil {
		switch {
		case errors.Is(err, basic.ErrNeverRise):
			return time.Time{}, ERR_SUN_NEVER_RISE
		case errors.Is(err, basic.ErrNeverSet):
			return time.Time{}, ERR_SUN_NEVER_SET
		default:
			return time.Time{}, err
		}
	}
	_, offset := date.Zone()
	return basic.JD2DateByZone(jd-float64(offset)/86400, date.Location(), false), nil
}

func twilightResult(date time.Time, jd float64, err error) (time.Time, error) {
	if err != nil {
		return time.Time{}, ERR_TWILIGHT_NOT_EXISTS
	}
	_, offset := date.Zone()
	return basic.JD2DateByZone(jd-float64(offset)/86400, date.Location(), false), nil
}

/*
太阳
视星等	−26.74
绝对星等	4.839
光谱类型	G2V
金属量	Z = 0.0122
角直径	31.6′ – 32.7′
*/

// RiseTime 日出时刻 / sunrise time.
func (module *Sun) RiseTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	if result, err, handled := civiltime.Event(date, basic.ErrNotOnThisDate, func(d time.Time) (time.Time, error) {
		return module.RiseTime(d, lon, lat, height, aero)
	}); handled {
		return result, err
	}
	var aeroFloat float64
	if aero {
		aeroFloat = 1
	}
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	// 以 date 的当地日期为锚点，并读取其时区偏移参与地方时计算。
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	// 返回值保持与输入 date 一致的时区。
	riseJD, err := basic.GetSunRiseTimeWithInstance(localJD, lon, lat, timezone, aeroFloat, height, module.inst)
	return riseSetResult(date, riseJD, err)
}

// RiseTimeN 日出时刻（n 变体） / n-parameterised sunrise time.
func (module *Sun) RiseTimeN(date time.Time, lon, lat, height float64, aero bool, n int) (time.Time, error) {
	if result, err, handled := civiltime.Event(date, basic.ErrNotOnThisDate, func(d time.Time) (time.Time, error) {
		return module.RiseTimeN(d, lon, lat, height, aero, n)
	}); handled {
		return result, err
	}
	var aeroFloat float64
	if aero {
		aeroFloat = 1
	}
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	riseJD, err := basic.GetSunRiseTimeNWithInstance(localJD, lon, lat, timezone, aeroFloat, height, n, module.inst)
	return riseSetResult(date, riseJD, err)
}

// DownTime 日落时刻别名 / deprecated sunset alias.
func (module *Sun) DownTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	return module.SetTime(date, lon, lat, height, aero)
}

// DownTimeN 日落时刻别名（n 变体） / deprecated n-parameterised sunset alias.
func (module *Sun) DownTimeN(date time.Time, lon, lat, height float64, aero bool, n int) (time.Time, error) {
	return module.SetTimeN(date, lon, lat, height, aero, n)
}

// SetTime 日落时刻 / sunset time.
func (module *Sun) SetTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	if result, err, handled := civiltime.Event(date, basic.ErrNotOnThisDate, func(d time.Time) (time.Time, error) {
		return module.SetTime(d, lon, lat, height, aero)
	}); handled {
		return result, err
	}
	var aeroFloat float64
	if aero {
		aeroFloat = 1
	}
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	downJD, err := basic.GetSunSetTimeWithInstance(localJD, lon, lat, timezone, aeroFloat, height, module.inst)
	return riseSetResult(date, downJD, err)
}

// SetTimeN 日落时刻（n 变体） / n-parameterised sunset time.
func (module *Sun) SetTimeN(date time.Time, lon, lat, height float64, aero bool, n int) (time.Time, error) {
	if result, err, handled := civiltime.Event(date, basic.ErrNotOnThisDate, func(d time.Time) (time.Time, error) {
		return module.SetTimeN(d, lon, lat, height, aero, n)
	}); handled {
		return result, err
	}
	var aeroFloat float64
	if aero {
		aeroFloat = 1
	}
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	downJD, err := basic.GetSunSetTimeNWithInstance(localJD, lon, lat, timezone, aeroFloat, height, n, module.inst)
	return riseSetResult(date, downJD, err)
}

// MorningTwilight 晨光始时 / morning twilight.
func (module *Sun) MorningTwilight(date time.Time, lon, lat, angle float64) (time.Time, error) {
	if result, err, handled := civiltime.Event(date, basic.ErrNotOnThisDate, func(d time.Time) (time.Time, error) {
		return module.MorningTwilight(d, lon, lat, angle)
	}); handled {
		return result, err
	}
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	calcJD, err := basic.MorningTwilightWithInstance(localJD, lon, lat, timezone, angle, module.inst)
	return twilightResult(date, calcJD, err)
}

// MorningTwilightN 晨光始时（n 变体） / n-parameterised morning twilight.
func (module *Sun) MorningTwilightN(date time.Time, lon, lat, angle float64, n int) (time.Time, error) {
	if result, err, handled := civiltime.Event(date, basic.ErrNotOnThisDate, func(d time.Time) (time.Time, error) {
		return module.MorningTwilightN(d, lon, lat, angle, n)
	}); handled {
		return result, err
	}
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	calcJD, err := basic.MorningTwilightNWithInstance(localJD, lon, lat, timezone, angle, n, module.inst)
	return twilightResult(date, calcJD, err)
}

// EveningTwilight 暮光终时 / evening twilight.
func (module *Sun) EveningTwilight(date time.Time, lon, lat, angle float64) (time.Time, error) {
	if result, err, handled := civiltime.Event(date, basic.ErrNotOnThisDate, func(d time.Time) (time.Time, error) {
		return module.EveningTwilight(d, lon, lat, angle)
	}); handled {
		return result, err
	}
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	//不需要进行力学时转换，会在GetBanTime中转换
	calcJD, err := basic.EveningTwilightWithInstance(localJD, lon, lat, timezone, angle, module.inst)
	return twilightResult(date, calcJD, err)
}

// EveningTwilightN 暮光终时（n 变体） / n-parameterised evening twilight.
func (module *Sun) EveningTwilightN(date time.Time, lon, lat, angle float64, n int) (time.Time, error) {
	if result, err, handled := civiltime.Event(date, basic.ErrNotOnThisDate, func(d time.Time) (time.Time, error) {
		return module.EveningTwilightN(d, lon, lat, angle, n)
	}); handled {
		return result, err
	}
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	calcJD, err := basic.EveningTwilightNWithInstance(localJD, lon, lat, timezone, angle, n, module.inst)
	return twilightResult(date, calcJD, err)
}

// EclipticObliquity 黄赤交角 / ecliptic obliquity.
//
// 返回 date 对应绝对时刻的黄赤交角，单位度；nutation 为 true 时加入交角章动。
// Returns the obliquity of the ecliptic at the instant represented by date, in degrees. When nutation is true, obliquity nutation is included.
func EclipticObliquity(date time.Time, nutation bool) float64 {
	//转换为UTC时间
	jd := basic.Date2JD(date.UTC())
	//进行力学时转换
	jde := basic.UTC2TT(jd)
	//黄赤交角计算
	return basic.EclipticObliquity(jde, nutation)
}

// EclipticNutation 黄经章动（IAU 2000B） / nutation in longitude, IAU 2000B.
//
// 返回 date 对应绝对时刻的黄经章动，单位度。
// Returns nutation in longitude at the instant represented by date, in degrees.
func EclipticNutation(date time.Time) float64 {
	//转换为UTC时间
	jd := basic.Date2JD(date.UTC())
	//进行力学时转换与章动计算
	return basic.Nutation2000Bi(basic.UTC2TT(jd))
}

// EclipticNutation1980 黄经章动（IAU 1980） / nutation in longitude, IAU 1980.
//
// 返回 date 对应绝对时刻的黄经章动，单位度。
// Returns nutation in longitude at the instant represented by date, in degrees.
func EclipticNutation1980(date time.Time) float64 {
	//转换为UTC时间
	jd := basic.Date2JD(date.UTC())
	//进行力学时转换与章动计算
	return basic.Nutation1980i(basic.UTC2TT(jd))
}

// AxialtiltNutation 交角章动（IAU 2000B） / nutation in obliquity, IAU 2000B.
//
// 返回 date 对应绝对时刻的交角章动，单位度。
// Returns nutation in obliquity at the instant represented by date, in degrees.
func AxialtiltNutation(date time.Time) float64 {
	//转换为UTC时间
	jd := basic.Date2JD(date.UTC())
	//进行力学时转换与章动计算
	return basic.Nutation2000Bs(basic.UTC2TT(jd))
}

// AxialtiltNutation1980 交角章动（IAU 1980） / nutation in obliquity, IAU 1980.
//
// 返回 date 对应绝对时刻的交角章动，单位度。
// Returns nutation in obliquity at the instant represented by date, in degrees.
func AxialtiltNutation1980(date time.Time) float64 {
	//转换为UTC时间
	jd := basic.Date2JD(date.UTC())
	//进行力学时转换与章动计算
	return basic.Nutation1980s(basic.UTC2TT(jd))
}

// GeometricLo 太阳几何黄经 / geometric ecliptic longitude.
//
// 返回 date 对应绝对时刻的太阳几何黄经，单位度。
// Returns the Sun's geometric ecliptic longitude at the instant represented by date, in degrees.
func GeometricLo(date time.Time) float64 {
	//转换为UTC时间
	jd := basic.Date2JD(date.UTC())
	return basic.SunLo(basic.UTC2TT(jd))
}

// TrueLo 太阳真黄经 / true ecliptic longitude.
func (module *Sun) TrueLo(date time.Time) float64 {
	//转换为UTC时间
	jd := basic.Date2JD(date.UTC())
	return basic.HSunTrueLoWithInstance(basic.UTC2TT(jd), module.inst)
}

// TrueLoN 太阳真黄经（n 变体） / n-parameterised true ecliptic longitude.
func (module *Sun) TrueLoN(date time.Time, n int) float64 {
	jd := basic.Date2JD(date.UTC())
	return basic.HSunTrueLoNWithInstance(basic.UTC2TT(jd), n, module.inst)
}

// TrueBo 太阳真黄纬 / true ecliptic latitude.
func (module *Sun) TrueBo(date time.Time) float64 {
	//转换为UTC时间
	jd := basic.Date2JD(date.UTC())
	return basic.HSunTrueBoWithInstance(basic.UTC2TT(jd), module.inst)
}

// TrueBoN 太阳真黄纬（n 变体） / n-parameterised true ecliptic latitude.
func (module *Sun) TrueBoN(date time.Time, n int) float64 {
	jd := basic.Date2JD(date.UTC())
	return basic.HSunTrueBoNWithInstance(basic.UTC2TT(jd), n, module.inst)
}

// ApparentLo 太阳视黄经 / apparent ecliptic longitude.
func (module *Sun) ApparentLo(date time.Time) float64 {
	//转换为UTC时间
	jd := basic.Date2JD(date.UTC())
	return basic.HSunApparentLoWithInstance(basic.UTC2TT(jd), module.inst)
}

// ApparentRa 太阳地心视赤经 / apparent geocentric right ascension.
func (module *Sun) ApparentRa(date time.Time) float64 {
	//转换为UTC时间
	jd := basic.Date2JD(date.UTC())
	return basic.HSunApparentRaWithInstance(basic.UTC2TT(jd), module.inst)
}

// ApparentDec 太阳地心视赤纬 / apparent geocentric declination.
func (module *Sun) ApparentDec(date time.Time) float64 {
	//转换为UTC时间
	jd := basic.Date2JD(date.UTC())
	return basic.HSunApparentDecWithInstance(basic.UTC2TT(jd), module.inst)
}

// ApparentRaDec 太阳地心视赤经、视赤纬 / apparent geocentric right ascension and declination.
func (module *Sun) ApparentRaDec(date time.Time) (float64, float64) {
	//转换为UTC时间
	jd := basic.Date2JD(date.UTC())
	return basic.HSunApparentRaDecWithInstance(basic.UTC2TT(jd), module.inst)
}

// MidFunc 太阳中心差 / solar equation of center.
//
// 返回 date 对应绝对时刻的太阳中心差，单位度。
// Returns the Sun's equation of center at the instant represented by date, in degrees.
func MidFunc(date time.Time) float64 {
	//转换为UTC时间
	jd := basic.Date2JD(date.UTC())
	return basic.SunMidFun(basic.UTC2TT(jd))
}

// EquationTime 均时差 / equation of time.
func (module *Sun) EquationTime(date time.Time) float64 {
	//转换为UTC时间
	jd := basic.Date2JD(date.UTC())
	return basic.SunTimeWithInstance(basic.UTC2TT(jd), module.inst)
}

// HourAngle 太阳地心视时角，不含站心视差 / geocentric apparent solar hour angle without topocentric parallax.
func (module *Sun) HourAngle(date time.Time, lon, lat float64) float64 {
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.SunTimeAngleWithInstance(localJD, lon, lat, timezone, module.inst)
}

// HourAngleN 太阳时角（n 变体） / n-parameterised hour angle.
func (module *Sun) HourAngleN(date time.Time, lon, lat float64, n int) float64 {
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.SunTimeAngleNWithInstance(localJD, lon, lat, timezone, n, module.inst)
}

// Azimuth 太阳站心方位角 / topocentric solar azimuth.
func (module *Sun) Azimuth(date time.Time, lon, lat float64) float64 {
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.SunAzimuthWithInstance(localJD, lon, lat, timezone, module.inst)
}

// AzimuthN 太阳站心方位角（n 变体） / n-parameterised topocentric solar azimuth.
func (module *Sun) AzimuthN(date time.Time, lon, lat float64, n int) float64 {
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.SunAzimuthNWithInstance(localJD, lon, lat, timezone, n, module.inst)
}

// Altitude 太阳站心视高度角，不含大气折射 / topocentric apparent solar altitude without refraction.
func (module *Sun) Altitude(date time.Time, lon, lat float64) float64 {
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.SunHeightWithInstance(localJD, lon, lat, timezone, module.inst)
}

// Zenith 太阳天顶距 / solar zenith distance.
func (module *Sun) Zenith(date time.Time, lon, lat float64) float64 {
	return 90 - module.Altitude(date, lon, lat)
}

// AltitudeN 太阳站心视高度角（n 变体） / n-parameterised topocentric apparent solar altitude.
func (module *Sun) AltitudeN(date time.Time, lon, lat float64, n int) float64 {
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.SunHeightNWithInstance(localJD, lon, lat, timezone, n, module.inst)
}

// ZenithN 太阳天顶距（n 变体） / n-parameterised solar zenith distance.
func (module *Sun) ZenithN(date time.Time, lon, lat float64, n int) float64 {
	return 90 - module.AltitudeN(date, lon, lat, n)
}

// CulminationTime 太阳中天时刻 / culmination time.
func (module *Sun) CulminationTime(date time.Time, lon float64) time.Time {
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	localJD := basic.Date2JD(date) + 0.5
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	calcJD := basic.CulminationTimeWithInstance(localJD, lon, timezone, module.inst) - timezone/24.00
	return basic.JD2DateByZone(calcJD, date.Location(), false)
}

// CulminationTimeN 太阳中天时刻（n 变体） / n-parameterised culmination time.
func (module *Sun) CulminationTimeN(date time.Time, lon float64, n int) time.Time {
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	localJD := basic.Date2JD(date) + 0.5
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	calcJD := basic.CulminationTimeNWithInstance(localJD, lon, timezone, n, module.inst) - timezone/24.00
	return basic.JD2DateByZone(calcJD, date.Location(), false)
}

// EarthDistance 日地距离 / Earth-Sun distance.
func (module *Sun) EarthDistance(date time.Time) float64 {
	jd := basic.Date2JD(date.UTC())
	jde := basic.UTC2TT(jd)
	return basic.EarthAwayWithInstance(jde, module.inst)
}

// ApparentSolarTime 真太阳时 / apparent solar time.
func (module *Sun) ApparentSolarTime(date time.Time, lon float64) time.Time {
	//真太阳时=太阳时角+12小时
	trueTime := (module.HourAngle(date, lon, 0) + 180) / 15
	if trueTime > 24 {
		trueTime -= 24
	}
	//真太阳时的分
	minute := (trueTime - math.Floor(trueTime)) * 60
	//真太阳时的秒
	second := (minute - math.Floor(minute)) * 60
	//当地经度下的本地时区
	trueSunTime := date.In(time.FixedZone("LTZ", int(lon*3600.00/15.0)))
	if trueSunTime.Hour()-int(trueTime) > 12 {
		trueSunTime = trueSunTime.Add(time.Hour * 24)
	} else if int(trueTime)-trueSunTime.Hour() > 12 {
		trueSunTime = trueSunTime.Add(-time.Hour * 24)
	}
	return time.Date(trueSunTime.Year(), trueSunTime.Month(), trueSunTime.Day(),
		int(trueTime), int(minute), int(second), int((second-math.Floor(second))*1000000000),
		time.FixedZone("LTZ", int(lon*3600.00/15.0)))
}

// ApparentSolarTimeN 真太阳时（n 变体） / n-parameterised apparent solar time.
func (module *Sun) ApparentSolarTimeN(date time.Time, lon float64, n int) time.Time {
	trueTime := (module.HourAngleN(date, lon, 0, n) + 180) / 15
	if trueTime > 24 {
		trueTime -= 24
	}
	minute := (trueTime - math.Floor(trueTime)) * 60
	second := (minute - math.Floor(minute)) * 60
	trueSunTime := date.In(time.FixedZone("LTZ", int(lon*3600.00/15.0)))
	if trueSunTime.Hour()-int(trueTime) > 12 {
		trueSunTime = trueSunTime.Add(time.Hour * 24)
	} else if int(trueTime)-trueSunTime.Hour() > 12 {
		trueSunTime = trueSunTime.Add(-time.Hour * 24)
	}
	return time.Date(trueSunTime.Year(), trueSunTime.Month(), trueSunTime.Day(),
		int(trueTime), int(minute), int(second), int((second-math.Floor(second))*1000000000),
		time.FixedZone("LTZ", int(lon*3600.00/15.0)))
}

// RiseTimeWithAtmosphere 日出时刻，按给定气压（hPa）与气温（°C）计算大气折射 / sunrise time with explicit pressure and temperature.
func (module *Sun) RiseTimeWithAtmosphere(date time.Time, lon, lat, height, pressureHPa, temperatureC float64) (time.Time, error) {
	if result, err, handled := civiltime.Event(date, basic.ErrNotOnThisDate, func(d time.Time) (time.Time, error) {
		return module.RiseTimeWithAtmosphere(d, lon, lat, height, pressureHPa, temperatureC)
	}); handled {
		return result, err
	}
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	jd, err := basic.GetSunRiseTimeWithAtmosphere(localJD, lon, lat, float64(loc)/3600.0, height, pressureHPa, temperatureC, module.inst)
	return riseSetResult(date, jd, err)
}

// SetTimeWithAtmosphere 日落时刻，按给定气压（hPa）与气温（°C）计算大气折射 / sunset time with explicit pressure and temperature.
func (module *Sun) SetTimeWithAtmosphere(date time.Time, lon, lat, height, pressureHPa, temperatureC float64) (time.Time, error) {
	if result, err, handled := civiltime.Event(date, basic.ErrNotOnThisDate, func(d time.Time) (time.Time, error) {
		return module.SetTimeWithAtmosphere(d, lon, lat, height, pressureHPa, temperatureC)
	}); handled {
		return result, err
	}
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	jd, err := basic.GetSunSetTimeWithAtmosphere(localJD, lon, lat, float64(loc)/3600.0, height, pressureHPa, temperatureC, module.inst)
	return riseSetResult(date, jd, err)
}
