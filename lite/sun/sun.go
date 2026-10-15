package sun

import (
	"github.com/starainrt/astro/internal/civiltime"
	"errors"
	"math"
	"time"

	lite "github.com/starainrt/astro/lite/internal"
)

var (
	ERR_SUN_NEVER_RISE = errors.New("ERROR:极夜，太阳在今日永远在地平线下！")
	ERR_SUN_NEVER_SET  = errors.New("ERROR:极昼，太阳在今日永远在地平线上！")
)

// TrueLo 轻量真黄经 / lightweight true ecliptic longitude.
func TrueLo(date time.Time) float64 {
	return lite.SunTrueLo(lite.Date2JD(date.UTC()))
}

// ApparentLo 轻量视黄经 / lightweight apparent ecliptic longitude.
func ApparentLo(date time.Time) float64 {
	return lite.SunApparentLo(lite.Date2JD(date.UTC()))
}

// Distance 轻量日地距离 / lightweight Sun-Earth distance in AU.
func Distance(date time.Time) float64 {
	return lite.SunDistanceAU(lite.Date2JD(date.UTC()))
}

// TrueRa 轻量真赤经 / lightweight true right ascension.
func TrueRa(date time.Time) float64 {
	ra, _ := lite.SunTrueRaDec(lite.Date2JD(date.UTC()))
	return ra
}

// TrueDec 轻量真赤纬 / lightweight true declination.
func TrueDec(date time.Time) float64 {
	_, dec := lite.SunTrueRaDec(lite.Date2JD(date.UTC()))
	return dec
}

// TrueRaDec 轻量真赤经、真赤纬 / lightweight true right ascension and declination.
func TrueRaDec(date time.Time) (float64, float64) {
	return lite.SunTrueRaDec(lite.Date2JD(date.UTC()))
}

// ApparentRa 轻量视赤经 / lightweight apparent right ascension.
func ApparentRa(date time.Time) float64 {
	ra, _ := lite.SunApparentRaDec(lite.Date2JD(date.UTC()))
	return ra
}

// ApparentDec 轻量视赤纬 / lightweight apparent declination.
func ApparentDec(date time.Time) float64 {
	_, dec := lite.SunApparentRaDec(lite.Date2JD(date.UTC()))
	return dec
}

// ApparentRaDec 轻量视赤经、视赤纬 / lightweight apparent right ascension and declination.
func ApparentRaDec(date time.Time) (float64, float64) {
	return lite.SunApparentRaDec(lite.Date2JD(date.UTC()))
}

// HourAngle 轻量时角 / lightweight hour angle.
func HourAngle(date time.Time, lon, lat float64) float64 {
	_, _, hourAngle := lite.HorizontalCoordinates(ApparentRa(date), ApparentDec(date), lite.Date2JD(date.UTC()), lon, lat)
	return hourAngle
}

// Azimuth 轻量方位角 / lightweight azimuth.
func Azimuth(date time.Time, lon, lat float64) float64 {
	_, azimuth, _ := lite.HorizontalCoordinates(ApparentRa(date), ApparentDec(date), lite.Date2JD(date.UTC()), lon, lat)
	return azimuth
}

// Altitude 轻量高度角 / lightweight altitude.
func Altitude(date time.Time, lon, lat float64) float64 {
	altitude, _, _ := lite.HorizontalCoordinates(ApparentRa(date), ApparentDec(date), lite.Date2JD(date.UTC()), lon, lat)
	return altitude
}

// Zenith 轻量天顶距 / lightweight zenith distance.
func Zenith(date time.Time, lon, lat float64) float64 {
	return 90 - Altitude(date, lon, lat)
}

// RiseTime 轻量日出时刻 / lightweight sunrise time.
func RiseTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	if result, err, handled := civiltime.Event(date, lite.ErrNotOnThisDate, func(d time.Time) (time.Time, error) {
		return RiseTime(d, lon, lat, height, aero)
	}); handled {
		return result, err
	}
	return riseSetTime(date, lon, lat, height, aero, true)
}

// SetTime 轻量日落时刻 / lightweight sunset time.
func SetTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	if result, err, handled := civiltime.Event(date, lite.ErrNotOnThisDate, func(d time.Time) (time.Time, error) {
		return SetTime(d, lon, lat, height, aero)
	}); handled {
		return result, err
	}
	return riseSetTime(date, lon, lat, height, aero, false)
}

func riseSetTime(date time.Time, lon, lat, height float64, aero, isRise bool) (time.Time, error) {
	localMidnight := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	localJD := lite.Date2JD(localMidnight)
	_, offset := localMidnight.Zone()
	timezone := float64(offset) / 3600.0

	horizonDip := lite.HeightDegreeByLat(height, lat)

	altitudeFn := func(localJD float64) float64 {
		utcJD := localJD - timezone/24.0
		ra, dec := lite.SunApparentRaDec(utcJD)
		altitude, _, _ := lite.HorizontalCoordinates(ra, dec, utcJD, lon, lat)
		residual := altitude + horizonDip
		if aero {
			residual += lite.RefractionFromTrueAltitude(altitude, 1010, 10)
			residual += liteSunSemidiameterDegrees(lite.SunDistanceAU(utcJD))
		}
		return residual
	}

	eventJD, err := lite.SearchRiseSet(localJD, 0, 30, isRise, altitudeFn)
	if err != nil {
		switch {
		case errors.Is(err, lite.ErrNeverRise):
			return time.Time{}, ERR_SUN_NEVER_RISE
		case errors.Is(err, lite.ErrNeverSet):
			return time.Time{}, ERR_SUN_NEVER_SET
		default:
			return time.Time{}, err
		}
	}
	return lite.JD2DateByZone(eventJD-timezone/24, date.Location(), false), nil
}

func liteSunSemidiameterDegrees(distanceAU float64) float64 {
	const (
		sunRadiusKM = 695700.0
		auKM        = 149597870.7
	)
	return math.Asin(sunRadiusKM/(distanceAU*auKM)) * 180 / math.Pi
}
