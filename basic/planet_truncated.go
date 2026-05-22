package basic

import (
	"math"

	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

// Exported N variants below keep the same jd semantics as the non-N APIs; n < 0 means full series.
type planetDeclinationFuncN func(float64, int) float64

func planetXYZN(planetIndex int, jd float64, n int) (float64, float64, float64) {
	l := planet.WherePlanetN(planetIndex, 0, jd, n)
	b := planet.WherePlanetN(planetIndex, 1, jd, n)
	r := planet.WherePlanetN(planetIndex, 2, jd, n)
	el := planet.WherePlanetN(-1, 0, jd, n)
	eb := planet.WherePlanetN(-1, 1, jd, n)
	er := planet.WherePlanetN(-1, 2, jd, n)
	x := r*Cos(b)*Cos(l) - er*Cos(eb)*Cos(el)
	y := r*Cos(b)*Sin(l) - er*Cos(eb)*Sin(el)
	z := r*Sin(b) - er*Sin(eb)
	return x, y, z
}

func planetApparentLoBoN(planetIndex int, jd float64, n int) (float64, float64) {
	geo, _ := planetApparentGeocentricPositionN(planetIndex, jd, n)
	return geo.lo, geo.bo
}

func planetApparentRaManualN(planetIndex int, jd float64, n int) float64 {
	lo, bo := planetApparentLoBoN(planetIndex, jd, n)
	eps := TrueObliquity(jd)
	ra := math.Atan2((Sin(lo)*Cos(eps) - Tan(bo)*Sin(eps)), Cos(lo))
	return Limit360(ra * 180 / math.Pi)
}

func planetApparentDecManualN(planetIndex int, jd float64, n int) float64 {
	lo, bo := planetApparentLoBoN(planetIndex, jd, n)
	eps := TrueObliquity(jd)
	return ArcSin(Sin(bo)*Cos(eps) + Cos(bo)*Sin(eps)*Sin(lo))
}

func planetApparentRaDecManualN(planetIndex int, jd float64, n int) (float64, float64) {
	lo, bo := planetApparentLoBoN(planetIndex, jd, n)
	eps := TrueObliquity(jd)
	ra := math.Atan2((Sin(lo)*Cos(eps) - Tan(bo)*Sin(eps)), Cos(lo))
	ra = ra * 180 / math.Pi
	dec := ArcSin(Sin(bo)*Cos(eps) + Cos(bo)*Sin(eps)*Sin(lo))
	return Limit360(ra), dec
}

func planetEarthAwayN(planetIndex int, jd float64, n int) float64 {
	return planetEarthAwayExplicitN(planetIndex, jd, n)
}

func planetHeightN(jde, lon, lat, timezone float64, n int, apparentRaDec func(float64, int) (float64, float64)) float64 {
	utcJde := jde - timezone/24.0
	ra, dec := apparentRaDec(TD2UT(utcJde, true), n)
	st := Limit360(ApparentSiderealTime(utcJde)*15 + lon)
	H := Limit360(st - ra)
	sinHeight := Sin(lat)*Sin(dec) + Cos(dec)*Cos(lat)*Cos(H)
	return ArcSin(sinHeight)
}

func planetAzimuthN(jde, lon, lat, timezone float64, n int, apparentRaDec func(float64, int) (float64, float64)) float64 {
	utcJde := jde - timezone/24.0
	ra, dec := apparentRaDec(TD2UT(utcJde, true), n)
	st := Limit360(ApparentSiderealTime(utcJde)*15 + lon)
	H := Limit360(st - ra)
	tanAzimuth := Sin(H) / (Cos(H)*Sin(lat) - Tan(dec)*Cos(lat))
	azimuth := ArcTan(tanAzimuth)
	if azimuth < 0 {
		if H/15 < 12 {
			return azimuth + 360
		}
		return azimuth + 180
	}
	if H/15 < 12 {
		return azimuth + 180
	}
	return azimuth
}

func planetHourAngleN(jd, lon, timezone float64, n int, apparentRa func(float64, int) float64) float64 {
	siderealLongitude := Limit360(ApparentSiderealTime(jd-timezone/24)*15 + lon)
	hourAngle := siderealLongitude - apparentRa(TD2UT(jd-timezone/24.0, true), n)
	if hourAngle < 0 {
		hourAngle += 360
	}
	return hourAngle
}

func planetCulminationTimeN(jde, lon, timezone float64, n int, hourAngle func(float64, float64, float64, int) float64) float64 {
	jde = math.Floor(jde) + 0.5
	estimateJD := jde + Limit360(360-hourAngle(jde, lon, timezone, n))/15.0/24.0*0.99726851851851851851
	normalizedHourAngle := func(jde, lon, timezone float64) float64 {
		currentHourAngle := hourAngle(jde, lon, timezone, n)
		if currentHourAngle < 180 {
			currentHourAngle += 360
		}
		return currentHourAngle
	}
	for {
		prevJD := estimateJD
		hourAngleDelta := normalizedHourAngle(prevJD, lon, timezone) - 360
		hourAngleSlope := (normalizedHourAngle(prevJD+0.000005, lon, timezone) - normalizedHourAngle(prevJD-0.000005, lon, timezone)) / 0.00001
		estimateJD = prevJD - hourAngleDelta/hourAngleSlope
		if math.Abs(estimateJD-prevJD) <= 0.00001 {
			break
		}
	}
	return estimateJD
}

func planetRiseDownN(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, isRise bool, n int, culmination func(float64, float64, float64, int) float64, height func(float64, float64, float64, float64, int) float64, declination planetDeclinationFuncN) (float64, error) {
	jd = math.Floor(jd) + 0.5
	localTimezone := math.Round(lon / 15)
	targetAltitude := StandardAltitudePlanet(aeroCorrection, observerHeight, lat)
	culminationJD := culmination(jd, lon, localTimezone, n)
	if height(culminationJD, lon, lat, localTimezone, n) < targetAltitude {
		return 0, ErrNeverRise
	}
	if height(culminationJD-0.5, lon, lat, localTimezone, n) > targetAltitude {
		return 0, ErrNeverSet
	}
	dec := declination(TD2UT(culminationJD-localTimezone/24, true), n)
	cosHourAngle := (Sin(targetAltitude) - Sin(dec)*Sin(lat)) / (Cos(dec) * Cos(lat))
	var eventJD float64
	if math.Abs(cosHourAngle) <= 1 {
		hourOffset := ArcCos(cosHourAngle) / 15
		if isRise {
			eventJD = culminationJD - hourOffset/24 - 25.0/24.0/60.0
		} else {
			eventJD = culminationJD + hourOffset/24 - 25.0/24.0/60.0
		}
	} else {
		eventJD = culminationJD
		steps := 0
		for height(eventJD, lon, lat, localTimezone, n) > targetAltitude {
			steps++
			if isRise {
				eventJD -= 15.0 / 60.0 / 24.0
			} else {
				eventJD += 15.0 / 60.0 / 24.0
			}
			if steps > 48 {
				break
			}
		}
	}
	estimateJD := eventJD
	for {
		prevJD := estimateJD
		altitudeDelta := height(prevJD, lon, lat, localTimezone, n) - targetAltitude
		altitudeSlope := (height(prevJD+0.000005, lon, lat, localTimezone, n) - height(prevJD-0.000005, lon, lat, localTimezone, n)) / 0.00001
		estimateJD = prevJD - altitudeDelta/altitudeSlope
		if math.Abs(estimateJD-prevJD) <= 0.00001 {
			break
		}
	}
	return estimateJD - localTimezone/24 + timezone/24, nil
}

// MercuryApparentLoN 水星视黄经（截断版） / truncated apparent ecliptic longitude of Mercury.
func MercuryApparentLoN(jd float64, n int) float64 {
	lo, _ := planetApparentLoBoN(1, jd, n)
	return lo
}

// MercuryApparentBoN 水星视黄纬（截断版） / truncated apparent ecliptic latitude of Mercury.
func MercuryApparentBoN(jd float64, n int) float64 {
	_, bo := planetApparentLoBoN(1, jd, n)
	return bo
}

// MercuryApparentLoBoN 水星视黄经黄纬（截断版） / truncated apparent ecliptic longitude and latitude of Mercury.
func MercuryApparentLoBoN(jd float64, n int) (float64, float64) {
	return planetApparentLoBoN(1, jd, n)
}

// MercuryApparentRaN 水星视赤经（截断版） / truncated apparent right ascension of Mercury.
func MercuryApparentRaN(jd float64, n int) float64 {
	lo, bo := MercuryApparentLoBoN(jd, n)
	return LoToRa(jd, lo, bo)
}

// MercuryApparentDecN 水星视赤纬（截断版） / truncated apparent declination of Mercury.
func MercuryApparentDecN(jd float64, n int) float64 {
	lo, bo := MercuryApparentLoBoN(jd, n)
	eps := TrueObliquity(jd)
	return ArcSin(Sin(bo)*Cos(eps) + Cos(bo)*Sin(eps)*Sin(lo))
}

// MercuryApparentRaDecN 水星视赤经赤纬（截断版） / truncated apparent right ascension and declination of Mercury.
func MercuryApparentRaDecN(jd float64, n int) (float64, float64) {
	lo, bo := MercuryApparentLoBoN(jd, n)
	return LoBoToRaDec(jd, lo, bo)
}

// EarthMercuryAwayN 地水距离（截断版） / truncated Earth-Mercury distance.
func EarthMercuryAwayN(jd float64, n int) float64 {
	return planetEarthAwayN(1, jd, n)
}

// MercuryMagN 水星视星等（截断版） / truncated apparent magnitude of Mercury.
func MercuryMagN(jd float64, n int) float64 {
	awaySun := planet.WherePlanetN(1, 2, jd, n)
	awayEarth := EarthMercuryAwayN(jd, n)
	away := planet.WherePlanetN(-1, 2, jd, n)
	i := (awaySun*awaySun + awayEarth*awayEarth - away*away) / (2 * awaySun * awayEarth)
	i = ArcCos(i)
	mag := -0.42 + 5*math.Log10(awaySun*awayEarth) + 0.0380*i - 0.000273*i*i + 0.000002*i*i*i
	return FloatRound(mag, 2)
}

// MercuryHeightN 水星高度角（截断版） / truncated altitude of Mercury.
func MercuryHeightN(jde, lon, lat, timezone float64, n int) float64 {
	return planetHeightN(jde, lon, lat, timezone, n, MercuryApparentRaDecN)
}

// MercuryAzimuthN 水星方位角（截断版） / truncated azimuth of Mercury.
func MercuryAzimuthN(jde, lon, lat, timezone float64, n int) float64 {
	return planetAzimuthN(jde, lon, lat, timezone, n, MercuryApparentRaDecN)
}

// MercuryHourAngleN 水星时角（截断版） / truncated hour angle of Mercury.
func MercuryHourAngleN(jd, lon, timezone float64, n int) float64 {
	return planetHourAngleN(jd, lon, timezone, n, MercuryApparentRaN)
}

// MercuryCulminationTimeN 水星中天时间（截断版） / truncated culmination time of Mercury.
func MercuryCulminationTimeN(jde, lon, timezone float64, n int) float64 {
	return planetCulminationTimeN(jde, lon, timezone, n, MercuryHourAngleN)
}

// MercuryRiseTimeN 水星升起时间（截断版） / truncated rise time of Mercury.
func MercuryRiseTimeN(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int) (float64, error) {
	return planetRiseDownN(jd, lon, lat, timezone, aeroCorrection, observerHeight, true, n, MercuryCulminationTimeN, MercuryHeightN, MercuryApparentDecN)
}

// MercurySetTimeN 水星落下时间（截断版） / truncated set time of Mercury.
func MercurySetTimeN(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int) (float64, error) {
	return planetRiseDownN(jd, lon, lat, timezone, aeroCorrection, observerHeight, false, n, MercuryCulminationTimeN, MercuryHeightN, MercuryApparentDecN)
}

// VenusApparentLoN 金星视黄经（截断版） / truncated apparent ecliptic longitude of Venus.
func VenusApparentLoN(jd float64, n int) float64 {
	lo, _ := planetApparentLoBoN(2, jd, n)
	return lo
}

// VenusApparentBoN 金星视黄纬（截断版） / truncated apparent ecliptic latitude of Venus.
func VenusApparentBoN(jd float64, n int) float64 {
	_, bo := planetApparentLoBoN(2, jd, n)
	return bo
}

// VenusApparentLoBoN 金星视黄经黄纬（截断版） / truncated apparent ecliptic longitude and latitude of Venus.
func VenusApparentLoBoN(jd float64, n int) (float64, float64) {
	return planetApparentLoBoN(2, jd, n)
}

// VenusApparentRaN 金星视赤经（截断版） / truncated apparent right ascension of Venus.
func VenusApparentRaN(jd float64, n int) float64 {
	return planetApparentRaManualN(2, jd, n)
}

// VenusApparentDecN 金星视赤纬（截断版） / truncated apparent declination of Venus.
func VenusApparentDecN(jd float64, n int) float64 {
	return planetApparentDecManualN(2, jd, n)
}

// VenusApparentRaDecN 金星视赤经赤纬（截断版） / truncated apparent right ascension and declination of Venus.
func VenusApparentRaDecN(jd float64, n int) (float64, float64) {
	return planetApparentRaDecManualN(2, jd, n)
}

// EarthVenusAwayN 地金距离（截断版） / truncated Earth-Venus distance.
func EarthVenusAwayN(jd float64, n int) float64 {
	return planetEarthAwayN(2, jd, n)
}

// VenusMagN 金星视星等（截断版） / truncated apparent magnitude of Venus.
func VenusMagN(jd float64, n int) float64 {
	awaySun := planet.WherePlanetN(2, 2, jd, n)
	awayEarth := EarthVenusAwayN(jd, n)
	away := planet.WherePlanetN(-1, 2, jd, n)
	i := (awaySun*awaySun + awayEarth*awayEarth - away*away) / (2 * awaySun * awayEarth)
	i = ArcCos(i)
	mag := -4.40 + 5*math.Log10(awaySun*awayEarth) + 0.0009*i + 0.000239*i*i - 0.00000065*i*i*i
	return FloatRound(mag, 2)
}

// VenusHeightN 金星高度角（截断版） / truncated altitude of Venus.
func VenusHeightN(jde, lon, lat, timezone float64, n int) float64 {
	return planetHeightN(jde, lon, lat, timezone, n, VenusApparentRaDecN)
}

// VenusAzimuthN 金星方位角（截断版） / truncated azimuth of Venus.
func VenusAzimuthN(jde, lon, lat, timezone float64, n int) float64 {
	return planetAzimuthN(jde, lon, lat, timezone, n, VenusApparentRaDecN)
}

// VenusHourAngleN 金星时角（截断版） / truncated hour angle of Venus.
func VenusHourAngleN(jd, lon, timezone float64, n int) float64 {
	return planetHourAngleN(jd, lon, timezone, n, VenusApparentRaN)
}

// VenusCulminationTimeN 金星中天时间（截断版） / truncated culmination time of Venus.
func VenusCulminationTimeN(jde, lon, timezone float64, n int) float64 {
	return planetCulminationTimeN(jde, lon, timezone, n, VenusHourAngleN)
}

// VenusRiseTimeN 金星升起时间（截断版） / truncated rise time of Venus.
func VenusRiseTimeN(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int) (float64, error) {
	return planetRiseDownN(jd, lon, lat, timezone, aeroCorrection, observerHeight, true, n, VenusCulminationTimeN, VenusHeightN, VenusApparentDecN)
}

// VenusSetTimeN 金星落下时间（截断版） / truncated set time of Venus.
func VenusSetTimeN(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int) (float64, error) {
	return planetRiseDownN(jd, lon, lat, timezone, aeroCorrection, observerHeight, false, n, VenusCulminationTimeN, VenusHeightN, VenusApparentDecN)
}

// MarsApparentLoN 火星视黄经（截断版） / truncated apparent ecliptic longitude of Mars.
func MarsApparentLoN(jd float64, n int) float64 {
	lo, _ := planetApparentLoBoN(3, jd, n)
	return lo
}

// MarsApparentBoN 火星视黄纬（截断版） / truncated apparent ecliptic latitude of Mars.
func MarsApparentBoN(jd float64, n int) float64 {
	_, bo := planetApparentLoBoN(3, jd, n)
	return bo
}

// MarsApparentLoBoN 火星视黄经黄纬（截断版） / truncated apparent ecliptic longitude and latitude of Mars.
func MarsApparentLoBoN(jd float64, n int) (float64, float64) {
	return planetApparentLoBoN(3, jd, n)
}

// MarsApparentRaN 火星视赤经（截断版） / truncated apparent right ascension of Mars.
func MarsApparentRaN(jd float64, n int) float64 {
	return planetApparentRaManualN(3, jd, n)
}

// MarsApparentDecN 火星视赤纬（截断版） / truncated apparent declination of Mars.
func MarsApparentDecN(jd float64, n int) float64 {
	return planetApparentDecManualN(3, jd, n)
}

// MarsApparentRaDecN 火星视赤经赤纬（截断版） / truncated apparent right ascension and declination of Mars.
func MarsApparentRaDecN(jd float64, n int) (float64, float64) {
	return planetApparentRaDecManualN(3, jd, n)
}

// EarthMarsAwayN 地火距离（截断版） / truncated Earth-Mars distance.
func EarthMarsAwayN(jd float64, n int) float64 {
	return planetEarthAwayN(3, jd, n)
}

// MarsMagN 火星视星等（截断版） / truncated apparent magnitude of Mars.
func MarsMagN(jd float64, n int) float64 {
	awaySun := planet.WherePlanetN(3, 2, jd, n)
	awayEarth := EarthMarsAwayN(jd, n)
	away := planet.WherePlanetN(-1, 2, jd, n)
	i := (awaySun*awaySun + awayEarth*awayEarth - away*away) / (2 * awaySun * awayEarth)
	i = ArcCos(i)
	mag := -1.52 + 5*math.Log10(awaySun*awayEarth) + 0.016*i
	return FloatRound(mag, 2)
}

// MarsHeightN 火星高度角（截断版） / truncated altitude of Mars.
func MarsHeightN(jde, lon, lat, timezone float64, n int) float64 {
	return planetHeightN(jde, lon, lat, timezone, n, MarsApparentRaDecN)
}

// MarsAzimuthN 火星方位角（截断版） / truncated azimuth of Mars.
func MarsAzimuthN(jde, lon, lat, timezone float64, n int) float64 {
	return planetAzimuthN(jde, lon, lat, timezone, n, MarsApparentRaDecN)
}

// MarsHourAngleN 火星时角（截断版） / truncated hour angle of Mars.
func MarsHourAngleN(jd, lon, timezone float64, n int) float64 {
	return planetHourAngleN(jd, lon, timezone, n, MarsApparentRaN)
}

// MarsCulminationTimeN 火星中天时间（截断版） / truncated culmination time of Mars.
func MarsCulminationTimeN(jde, lon, timezone float64, n int) float64 {
	return planetCulminationTimeN(jde, lon, timezone, n, MarsHourAngleN)
}

// MarsRiseTimeN 火星升起时间（截断版） / truncated rise time of Mars.
func MarsRiseTimeN(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int) (float64, error) {
	return planetRiseDownN(jd, lon, lat, timezone, aeroCorrection, observerHeight, true, n, MarsCulminationTimeN, MarsHeightN, MarsApparentDecN)
}

// MarsSetTimeN 火星落下时间（截断版） / truncated set time of Mars.
func MarsSetTimeN(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int) (float64, error) {
	return planetRiseDownN(jd, lon, lat, timezone, aeroCorrection, observerHeight, false, n, MarsCulminationTimeN, MarsHeightN, MarsApparentDecN)
}

// JupiterApparentLoN 木星视黄经（截断版） / truncated apparent ecliptic longitude of Jupiter.
func JupiterApparentLoN(jd float64, n int) float64 {
	lo, _ := planetApparentLoBoN(4, jd, n)
	return lo
}

// JupiterApparentBoN 木星视黄纬（截断版） / truncated apparent ecliptic latitude of Jupiter.
func JupiterApparentBoN(jd float64, n int) float64 {
	_, bo := planetApparentLoBoN(4, jd, n)
	return bo
}

// JupiterApparentLoBoN 木星视黄经黄纬（截断版） / truncated apparent ecliptic longitude and latitude of Jupiter.
func JupiterApparentLoBoN(jd float64, n int) (float64, float64) {
	return planetApparentLoBoN(4, jd, n)
}

// JupiterApparentRaN 木星视赤经（截断版） / truncated apparent right ascension of Jupiter.
func JupiterApparentRaN(jd float64, n int) float64 {
	return planetApparentRaManualN(4, jd, n)
}

// JupiterApparentDecN 木星视赤纬（截断版） / truncated apparent declination of Jupiter.
func JupiterApparentDecN(jd float64, n int) float64 {
	return planetApparentDecManualN(4, jd, n)
}

// JupiterApparentRaDecN 木星视赤经赤纬（截断版） / truncated apparent right ascension and declination of Jupiter.
func JupiterApparentRaDecN(jd float64, n int) (float64, float64) {
	return planetApparentRaDecManualN(4, jd, n)
}

// EarthJupiterAwayN 地木距离（截断版） / truncated Earth-Jupiter distance.
func EarthJupiterAwayN(jd float64, n int) float64 {
	return planetEarthAwayN(4, jd, n)
}

// JupiterMagN 木星视星等（截断版） / truncated apparent magnitude of Jupiter.
func JupiterMagN(jd float64, n int) float64 {
	awaySun := planet.WherePlanetN(4, 2, jd, n)
	awayEarth := EarthJupiterAwayN(jd, n)
	away := planet.WherePlanetN(-1, 2, jd, n)
	i := (awaySun*awaySun + awayEarth*awayEarth - away*away) / (2 * awaySun * awayEarth)
	i = ArcCos(i)
	mag := -9.40 + 5*math.Log10(awaySun*awayEarth) + 0.0005*i
	return FloatRound(mag, 2)
}

// JupiterHeightN 木星高度角（截断版） / truncated altitude of Jupiter.
func JupiterHeightN(jde, lon, lat, timezone float64, n int) float64 {
	return planetHeightN(jde, lon, lat, timezone, n, JupiterApparentRaDecN)
}

// JupiterAzimuthN 木星方位角（截断版） / truncated azimuth of Jupiter.
func JupiterAzimuthN(jde, lon, lat, timezone float64, n int) float64 {
	return planetAzimuthN(jde, lon, lat, timezone, n, JupiterApparentRaDecN)
}

// JupiterHourAngleN 木星时角（截断版） / truncated hour angle of Jupiter.
func JupiterHourAngleN(jd, lon, timezone float64, n int) float64 {
	return planetHourAngleN(jd, lon, timezone, n, JupiterApparentRaN)
}

// JupiterCulminationTimeN 木星中天时间（截断版） / truncated culmination time of Jupiter.
func JupiterCulminationTimeN(jde, lon, timezone float64, n int) float64 {
	return planetCulminationTimeN(jde, lon, timezone, n, JupiterHourAngleN)
}

// JupiterRiseTimeN 木星升起时间（截断版） / truncated rise time of Jupiter.
func JupiterRiseTimeN(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int) (float64, error) {
	return planetRiseDownN(jd, lon, lat, timezone, aeroCorrection, observerHeight, true, n, JupiterCulminationTimeN, JupiterHeightN, JupiterApparentDecN)
}

// JupiterSetTimeN 木星落下时间（截断版） / truncated set time of Jupiter.
func JupiterSetTimeN(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int) (float64, error) {
	return planetRiseDownN(jd, lon, lat, timezone, aeroCorrection, observerHeight, false, n, JupiterCulminationTimeN, JupiterHeightN, JupiterApparentDecN)
}

// SaturnApparentLoN 土星视黄经（截断版） / truncated apparent ecliptic longitude of Saturn.
func SaturnApparentLoN(jd float64, n int) float64 {
	lo, _ := planetApparentLoBoN(5, jd, n)
	return lo
}

// SaturnApparentBoN 土星视黄纬（截断版） / truncated apparent ecliptic latitude of Saturn.
func SaturnApparentBoN(jd float64, n int) float64 {
	_, bo := planetApparentLoBoN(5, jd, n)
	return bo
}

// SaturnApparentLoBoN 土星视黄经黄纬（截断版） / truncated apparent ecliptic longitude and latitude of Saturn.
func SaturnApparentLoBoN(jd float64, n int) (float64, float64) {
	return planetApparentLoBoN(5, jd, n)
}

// SaturnApparentRaN 土星视赤经（截断版） / truncated apparent right ascension of Saturn.
func SaturnApparentRaN(jd float64, n int) float64 {
	return planetApparentRaManualN(5, jd, n)
}

// SaturnApparentDecN 土星视赤纬（截断版） / truncated apparent declination of Saturn.
func SaturnApparentDecN(jd float64, n int) float64 {
	return planetApparentDecManualN(5, jd, n)
}

// SaturnApparentRaDecN 土星视赤经赤纬（截断版） / truncated apparent right ascension and declination of Saturn.
func SaturnApparentRaDecN(jd float64, n int) (float64, float64) {
	return planetApparentRaDecManualN(5, jd, n)
}

// EarthSaturnAwayN 地土距离（截断版） / truncated Earth-Saturn distance.
func EarthSaturnAwayN(jd float64, n int) float64 {
	return planetEarthAwayN(5, jd, n)
}

// SaturnMagN 土星视星等（截断版） / truncated apparent magnitude of Saturn.
func SaturnMagN(jd float64, n int) float64 {
	awaySun := planet.WherePlanetN(5, 2, jd, n)
	awayEarth := EarthSaturnAwayN(jd, n)
	ringB, _, _, deltaU, _, _ := SaturnRingParametersN(jd, n)
	ringB = math.Abs(ringB)
	mag := -8.68 + 5*math.Log10(awaySun*awayEarth) + 0.044*deltaU - 2.6*Sin(ringB) + 1.25*Sin(ringB)*Sin(ringB)
	return FloatRound(mag, 2)
}

// SaturnHeightN 土星高度角（截断版） / truncated altitude of Saturn.
func SaturnHeightN(jde, lon, lat, timezone float64, n int) float64 {
	return planetHeightN(jde, lon, lat, timezone, n, SaturnApparentRaDecN)
}

// SaturnAzimuthN 土星方位角（截断版） / truncated azimuth of Saturn.
func SaturnAzimuthN(jde, lon, lat, timezone float64, n int) float64 {
	return planetAzimuthN(jde, lon, lat, timezone, n, SaturnApparentRaDecN)
}

// SaturnHourAngleN 土星时角（截断版） / truncated hour angle of Saturn.
func SaturnHourAngleN(jd, lon, timezone float64, n int) float64 {
	return planetHourAngleN(jd, lon, timezone, n, SaturnApparentRaN)
}

// SaturnCulminationTimeN 土星中天时间（截断版） / truncated culmination time of Saturn.
func SaturnCulminationTimeN(jde, lon, timezone float64, n int) float64 {
	return planetCulminationTimeN(jde, lon, timezone, n, SaturnHourAngleN)
}

// SaturnRiseTimeN 土星升起时间（截断版） / truncated rise time of Saturn.
func SaturnRiseTimeN(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int) (float64, error) {
	return planetRiseDownN(jd, lon, lat, timezone, aeroCorrection, observerHeight, true, n, SaturnCulminationTimeN, SaturnHeightN, SaturnApparentDecN)
}

// SaturnSetTimeN 土星落下时间（截断版） / truncated set time of Saturn.
func SaturnSetTimeN(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int) (float64, error) {
	return planetRiseDownN(jd, lon, lat, timezone, aeroCorrection, observerHeight, false, n, SaturnCulminationTimeN, SaturnHeightN, SaturnApparentDecN)
}

// UranusApparentLoN 天王星视黄经（截断版） / truncated apparent ecliptic longitude of Uranus.
func UranusApparentLoN(jd float64, n int) float64 {
	lo, _ := planetApparentLoBoN(6, jd, n)
	return lo
}

// UranusApparentBoN 天王星视黄纬（截断版） / truncated apparent ecliptic latitude of Uranus.
func UranusApparentBoN(jd float64, n int) float64 {
	_, bo := planetApparentLoBoN(6, jd, n)
	return bo
}

// UranusApparentLoBoN 天王星视黄经黄纬（截断版） / truncated apparent ecliptic longitude and latitude of Uranus.
func UranusApparentLoBoN(jd float64, n int) (float64, float64) {
	return planetApparentLoBoN(6, jd, n)
}

// UranusApparentRaN 天王星视赤经（截断版） / truncated apparent right ascension of Uranus.
func UranusApparentRaN(jd float64, n int) float64 {
	return planetApparentRaManualN(6, jd, n)
}

// UranusApparentDecN 天王星视赤纬（截断版） / truncated apparent declination of Uranus.
func UranusApparentDecN(jd float64, n int) float64 {
	return planetApparentDecManualN(6, jd, n)
}

// UranusApparentRaDecN 天王星视赤经赤纬（截断版） / truncated apparent right ascension and declination of Uranus.
func UranusApparentRaDecN(jd float64, n int) (float64, float64) {
	return planetApparentRaDecManualN(6, jd, n)
}

// EarthUranusAwayN 地天距离（截断版） / truncated Earth-Uranus distance.
func EarthUranusAwayN(jd float64, n int) float64 {
	return planetEarthAwayN(6, jd, n)
}

// UranusMagN 天王星视星等（截断版） / truncated apparent magnitude of Uranus.
func UranusMagN(jd float64, n int) float64 {
	awaySun := planet.WherePlanetN(6, 2, jd, n)
	awayEarth := EarthUranusAwayN(jd, n)
	away := planet.WherePlanetN(-1, 2, jd, n)
	i := (awaySun*awaySun + awayEarth*awayEarth - away*away) / (2 * awaySun * awayEarth)
	i = ArcCos(i)
	mag := -7.19 + 5*math.Log10(awaySun*awayEarth) + 0.016*i
	return FloatRound(mag, 2)
}

// UranusHeightN 天王星高度角（截断版） / truncated altitude of Uranus.
func UranusHeightN(jde, lon, lat, timezone float64, n int) float64 {
	return planetHeightN(jde, lon, lat, timezone, n, UranusApparentRaDecN)
}

// UranusAzimuthN 天王星方位角（截断版） / truncated azimuth of Uranus.
func UranusAzimuthN(jde, lon, lat, timezone float64, n int) float64 {
	return planetAzimuthN(jde, lon, lat, timezone, n, UranusApparentRaDecN)
}

// UranusHourAngleN 天王星时角（截断版） / truncated hour angle of Uranus.
func UranusHourAngleN(jd, lon, timezone float64, n int) float64 {
	return planetHourAngleN(jd, lon, timezone, n, UranusApparentRaN)
}

// UranusCulminationTimeN 天王星中天时间（截断版） / truncated culmination time of Uranus.
func UranusCulminationTimeN(jde, lon, timezone float64, n int) float64 {
	return planetCulminationTimeN(jde, lon, timezone, n, UranusHourAngleN)
}

// UranusRiseTimeN 天王星升起时间（截断版） / truncated rise time of Uranus.
func UranusRiseTimeN(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int) (float64, error) {
	return planetRiseDownN(jd, lon, lat, timezone, aeroCorrection, observerHeight, true, n, UranusCulminationTimeN, UranusHeightN, UranusApparentDecN)
}

// UranusSetTimeN 天王星落下时间（截断版） / truncated set time of Uranus.
func UranusSetTimeN(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int) (float64, error) {
	return planetRiseDownN(jd, lon, lat, timezone, aeroCorrection, observerHeight, false, n, UranusCulminationTimeN, UranusHeightN, UranusApparentDecN)
}

// NeptuneApparentLoN 海王星视黄经（截断版） / truncated apparent ecliptic longitude of Neptune.
func NeptuneApparentLoN(jd float64, n int) float64 {
	lo, _ := planetApparentLoBoN(7, jd, n)
	return lo
}

// NeptuneApparentBoN 海王星视黄纬（截断版） / truncated apparent ecliptic latitude of Neptune.
func NeptuneApparentBoN(jd float64, n int) float64 {
	_, bo := planetApparentLoBoN(7, jd, n)
	return bo
}

// NeptuneApparentLoBoN 海王星视黄经黄纬（截断版） / truncated apparent ecliptic longitude and latitude of Neptune.
func NeptuneApparentLoBoN(jd float64, n int) (float64, float64) {
	return planetApparentLoBoN(7, jd, n)
}

// NeptuneApparentRaN 海王星视赤经（截断版） / truncated apparent right ascension of Neptune.
func NeptuneApparentRaN(jd float64, n int) float64 {
	return planetApparentRaManualN(7, jd, n)
}

// NeptuneApparentDecN 海王星视赤纬（截断版） / truncated apparent declination of Neptune.
func NeptuneApparentDecN(jd float64, n int) float64 {
	return planetApparentDecManualN(7, jd, n)
}

// NeptuneApparentRaDecN 海王星视赤经赤纬（截断版） / truncated apparent right ascension and declination of Neptune.
func NeptuneApparentRaDecN(jd float64, n int) (float64, float64) {
	return planetApparentRaDecManualN(7, jd, n)
}

// EarthNeptuneAwayN 地海距离（截断版） / truncated Earth-Neptune distance.
func EarthNeptuneAwayN(jd float64, n int) float64 {
	return planetEarthAwayN(7, jd, n)
}

// NeptuneMagN 海王星视星等（截断版） / truncated apparent magnitude of Neptune.
func NeptuneMagN(jd float64, n int) float64 {
	awaySun := planet.WherePlanetN(7, 2, jd, n)
	awayEarth := EarthNeptuneAwayN(jd, n)
	mag := -6.87 + 5*math.Log10(awaySun*awayEarth)
	return FloatRound(mag, 2)
}

// NeptuneHeightN 海王星高度角（截断版） / truncated altitude of Neptune.
func NeptuneHeightN(jde, lon, lat, timezone float64, n int) float64 {
	return planetHeightN(jde, lon, lat, timezone, n, NeptuneApparentRaDecN)
}

// NeptuneAzimuthN 海王星方位角（截断版） / truncated azimuth of Neptune.
func NeptuneAzimuthN(jde, lon, lat, timezone float64, n int) float64 {
	return planetAzimuthN(jde, lon, lat, timezone, n, NeptuneApparentRaDecN)
}

// NeptuneHourAngleN 海王星时角（截断版） / truncated hour angle of Neptune.
func NeptuneHourAngleN(jd, lon, timezone float64, n int) float64 {
	return planetHourAngleN(jd, lon, timezone, n, NeptuneApparentRaN)
}

// NeptuneCulminationTimeN 海王星中天时间（截断版） / truncated culmination time of Neptune.
func NeptuneCulminationTimeN(jde, lon, timezone float64, n int) float64 {
	return planetCulminationTimeN(jde, lon, timezone, n, NeptuneHourAngleN)
}

// NeptuneRiseTimeN 海王星升起时间（截断版） / truncated rise time of Neptune.
func NeptuneRiseTimeN(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int) (float64, error) {
	return planetRiseDownN(jd, lon, lat, timezone, aeroCorrection, observerHeight, true, n, NeptuneCulminationTimeN, NeptuneHeightN, NeptuneApparentDecN)
}

// NeptuneSetTimeN 海王星落下时间（截断版） / truncated set time of Neptune.
func NeptuneSetTimeN(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int) (float64, error) {
	return planetRiseDownN(jd, lon, lat, timezone, aeroCorrection, observerHeight, false, n, NeptuneCulminationTimeN, NeptuneHeightN, NeptuneApparentDecN)
}
