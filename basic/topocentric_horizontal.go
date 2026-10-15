package basic

import (
	"math"

	"github.com/starainrt/astro/internal/ephem"
	. "github.com/starainrt/astro/tools"
)

// topocentricHourAngleDec 由地心视赤经、视赤纬（度）与地心距（AU）给出观测者处的站心时角与站心视赤纬（度）。
// 恒星时只算一次，视差修正与时角共用；观测者取海拔 0，海拔对太阳系天体视差的影响可忽略。
func topocentricHourAngleDec(utcJD, lon, lat, ra, dec, distanceAU float64) (hourAngle, topocentricDec float64) {
	gast := ApparentSiderealTime(UTC2UT1(utcJD)) * 15
	topocentricRA, topocentricDec := topocentricRaDecWithSidereal(ra, dec, lat, lon, gast, distanceAU, 0)
	return Limit360(gast + lon - topocentricRA), topocentricDec
}

// horizontalAltitude 时角、赤纬、纬度（度）转高度角 / altitude from hour angle, declination and latitude.
func horizontalAltitude(hourAngle, dec, lat float64) float64 {
	return ArcSin(Sin(lat)*Sin(dec) + Cos(dec)*Cos(lat)*Cos(hourAngle))
}

// horizontalAzimuth 时角、赤纬、纬度（度）转方位角，北起向东 / azimuth from north through east.
func horizontalAzimuth(hourAngle, dec, lat float64) float64 {
	tanAzimuth := Sin(hourAngle) / (Cos(hourAngle)*Sin(lat) - Tan(dec)*Cos(lat))
	azimuth := ArcTan(tanAzimuth)
	if azimuth < 0 {
		if hourAngle/15 < 12 {
			return azimuth + 360
		}
		return azimuth + 180
	}
	if hourAngle/15 < 12 {
		return azimuth + 180
	}
	return azimuth
}

// planetApparentRaDecDistanceNWithInstance 行星地心视赤经、视赤纬与几何地心距，三者取自同一次求值。
func planetApparentRaDecDistanceNWithInstance(index int, jde float64, n int, inst *Instance) (ra, dec, distanceAU float64) {
	geo, _, distanceAU, ok := planetApparentPositionWithInstance(index, jde, n, inst)
	if !ok || math.IsNaN(geo.lo) || math.IsNaN(geo.bo) {
		return math.NaN(), math.NaN(), math.NaN()
	}
	ra, dec = planetApparentRaDecFromLoBo(jde, geo.lo, geo.bo)
	return ra, dec, distanceAU
}

func planetTopocentricHourAngleDecN(index int, localJD, lon, lat, timezone float64, n int, inst *Instance) (float64, float64) {
	utcJD := localJD - timezone/24.0
	ra, dec, distanceAU := planetApparentRaDecDistanceNWithInstance(index, UTC2TT(utcJD), n, inst)
	return topocentricHourAngleDec(utcJD, lon, lat, ra, dec, distanceAU)
}

// planetHeightN 行星站心视高度角（度，不含折射）/ topocentric planet altitude without refraction.
func planetHeightN(index int, localJD, lon, lat, timezone float64, n int, inst *Instance) float64 {
	hourAngle, dec := planetTopocentricHourAngleDecN(index, localJD, lon, lat, timezone, n, inst)
	return horizontalAltitude(hourAngle, dec, lat)
}

// planetAzimuthN 行星站心方位角（度）/ topocentric planet azimuth.
func planetAzimuthN(index int, localJD, lon, lat, timezone float64, n int, inst *Instance) float64 {
	hourAngle, dec := planetTopocentricHourAngleDecN(index, localJD, lon, lat, timezone, n, inst)
	return horizontalAzimuth(hourAngle, dec, lat)
}

// sunTopocentricHourAngleDecN 太阳站心时角、站心视赤纬与地日距 / topocentric solar hour angle, declination and distance.
func sunTopocentricHourAngleDecN(jd, lon, lat, tz float64, n int, inst *Instance) (hourAngle, dec, distanceAU float64) {
	utcJD := jd - tz/24.0
	ra, dec, distanceAU := hSunApparentRaDecDistanceN(UTC2TT(utcJD), n, inst)
	hourAngle, dec = topocentricHourAngleDec(utcJD, lon, lat, ra, dec, distanceAU)
	return hourAngle, dec, distanceAU
}

// sunAzimuthN 地距只用于视差，矢量视位置可用时直接取其距离，省去一次地距求值。
func sunAzimuthN(jd, lon, lat, tz float64, n int, inst *Instance) float64 {
	utcJD := jd - tz/24.0
	tt := UTC2TT(utcJD)
	var ra, dec, distanceAU float64
	if lo, bo, _, dist, ok, err := vectorApparentWithInstance(ephem.Sun, tt, n, inst); err != nil {
		return math.NaN()
	} else if ok {
		ra, dec = LoBoToRaDec(tt, lo, bo)
		distanceAU = dist
	} else {
		ra, dec, distanceAU = hSunApparentRaDecDistanceN(tt, n, inst)
	}
	hourAngle, topocentricDec := topocentricHourAngleDec(utcJD, lon, lat, ra, dec, distanceAU)
	return horizontalAzimuth(hourAngle, topocentricDec, lat)
}
