package moon

import (
	"errors"
	"time"

	"github.com/starainrt/astro/basic"
	lite "github.com/starainrt/astro/lite/internal"
	. "github.com/starainrt/astro/tools"
)

var (
	ERR_MOON_NEVER_RISE = errors.New("ERROR:极夜，月亮在今日永远在地平线下！")
	ERR_MOON_NEVER_SET  = errors.New("ERROR:极昼，月亮在今日永远在地平线上！")
	ERR_NOT_TODAY       = errors.New("ERROR:月亮已在（昨日/明日）（升起/降下）")
)

// TrueLo 轻量真黄经 / lightweight true ecliptic longitude.
func TrueLo(date time.Time) float64 {
	return lite.MoonGeocentric(basic.Date2JDE(date.UTC())).Longitude
}

// TrueBo 轻量真黄纬 / lightweight true ecliptic latitude.
func TrueBo(date time.Time) float64 {
	return lite.MoonGeocentric(basic.Date2JDE(date.UTC())).Latitude
}

// TrueRa 轻量真赤经 / lightweight true right ascension.
func TrueRa(date time.Time) float64 {
	return lite.MoonGeocentric(basic.Date2JDE(date.UTC())).RightAscension
}

// TrueDec 轻量真赤纬 / lightweight true declination.
func TrueDec(date time.Time) float64 {
	return lite.MoonGeocentric(basic.Date2JDE(date.UTC())).Declination
}

// TrueRaDec 轻量真赤经、真赤纬 / lightweight true right ascension and declination.
func TrueRaDec(date time.Time) (float64, float64) {
	state := lite.MoonGeocentric(basic.Date2JDE(date.UTC()))
	return state.RightAscension, state.Declination
}

// ApparentRa 轻量站心视赤经 / lightweight topocentric apparent right ascension.
func ApparentRa(date time.Time, lon, lat float64) float64 {
	state := lite.MoonTopocentric(basic.Date2JDE(date.UTC()), lon, lat, 0)
	return state.RightAscension
}

// ApparentDec 轻量站心视赤纬 / lightweight topocentric apparent declination.
func ApparentDec(date time.Time, lon, lat float64) float64 {
	state := lite.MoonTopocentric(basic.Date2JDE(date.UTC()), lon, lat, 0)
	return state.Declination
}

// ApparentRaDec 轻量站心视赤经、视赤纬 / lightweight topocentric apparent right ascension and declination.
func ApparentRaDec(date time.Time, lon, lat float64) (float64, float64) {
	state := lite.MoonTopocentric(basic.Date2JDE(date.UTC()), lon, lat, 0)
	return state.RightAscension, state.Declination
}

// HourAngle 轻量时角 / lightweight hour angle.
func HourAngle(date time.Time, lon, lat float64) float64 {
	_, _, hourAngle := lite.HorizontalCoordinates(ApparentRa(date, lon, lat), ApparentDec(date, lon, lat), basic.Date2JDE(date.UTC()), lon, lat)
	return hourAngle
}

// Azimuth 轻量方位角 / lightweight azimuth.
func Azimuth(date time.Time, lon, lat float64) float64 {
	_, azimuth, _ := lite.HorizontalCoordinates(ApparentRa(date, lon, lat), ApparentDec(date, lon, lat), basic.Date2JDE(date.UTC()), lon, lat)
	return azimuth
}

// Altitude 轻量高度角 / lightweight altitude.
func Altitude(date time.Time, lon, lat float64) float64 {
	altitude, _, _ := lite.HorizontalCoordinates(ApparentRa(date, lon, lat), ApparentDec(date, lon, lat), basic.Date2JDE(date.UTC()), lon, lat)
	return altitude
}

// Zenith 轻量天顶距 / lightweight zenith distance.
func Zenith(date time.Time, lon, lat float64) float64 {
	return 90 - Altitude(date, lon, lat)
}

// SunMoonLoDiff 轻量日月黄经差 / lightweight Moon-Sun ecliptic-longitude difference.
func SunMoonLoDiff(date time.Time) float64 {
	jd := basic.Date2JDE(date.UTC())
	return Limit360(lite.MoonGeocentric(jd).Longitude - lite.SunApparentLo(jd))
}

// PhaseAge 轻量月龄 / lightweight lunar age in days.
func PhaseAge(date time.Time) float64 {
	return lite.SynodicMonthDays * SunMoonLoDiff(date) / 360.0
}

// Phase 轻量受照比例 / lightweight illuminated fraction.
func Phase(date time.Time) float64 {
	return 0.5 * (1 - Cos(SunMoonLoDiff(date)))
}

// RiseTime 轻量月出时刻 / lightweight moonrise time.
func RiseTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	return riseSetTime(date, lon, lat, height, aero, true)
}

// SetTime 轻量月落时刻 / lightweight moonset time.
func SetTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	return riseSetTime(date, lon, lat, height, aero, false)
}

func riseSetTime(date time.Time, lon, lat, height float64, aero, isRise bool) (time.Time, error) {
	localMidnight := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	localJD := basic.Date2JDE(localMidnight)
	_, offset := localMidnight.Zone()
	timezone := float64(offset) / 3600.0

	targetAltitude := -basic.HeightDegreeByLat(height, lat)
	if aero {
		targetAltitude -= 0.83333
	}

	altitudeFn := func(localJD float64) float64 {
		utJD := localJD - timezone/24.0
		state := lite.MoonTopocentric(utJD, lon, lat, height)
		altitude, _, _ := lite.HorizontalCoordinates(state.RightAscension, state.Declination, utJD, lon, lat)
		return altitude
	}

	eventJD, err := lite.SearchRiseSet(localJD, targetAltitude, 15, isRise, altitudeFn)
	if err != nil {
		switch {
		case errors.Is(err, lite.ErrNeverRise):
			return time.Time{}, ERR_MOON_NEVER_RISE
		case errors.Is(err, lite.ErrNeverSet):
			return time.Time{}, ERR_MOON_NEVER_SET
		case errors.Is(err, lite.ErrNotOnThisDate):
			return time.Time{}, ERR_NOT_TODAY
		default:
			return time.Time{}, err
		}
	}
	return basic.JDE2DateByZone(eventJD, date.Location(), true), nil
}
