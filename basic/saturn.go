package basic

import (
	"math"

	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

func SaturnL(jd float64) float64 {
	return planet.WherePlanet(5, 0, jd)
}

func SaturnB(jd float64) float64 {
	return planet.WherePlanet(5, 1, jd)
}
func SaturnR(jd float64) float64 {
	return planet.WherePlanet(5, 2, jd)
}
func ASaturnX(jd float64) float64 {
	l := SaturnL(jd)
	b := SaturnB(jd)
	r := SaturnR(jd)
	el := planet.WherePlanet(-1, 0, jd)
	eb := planet.WherePlanet(-1, 1, jd)
	er := planet.WherePlanet(-1, 2, jd)
	x := r*Cos(b)*Cos(l) - er*Cos(eb)*Cos(el)
	return x
}

func ASaturnY(jd float64) float64 {

	l := SaturnL(jd)
	b := SaturnB(jd)
	r := SaturnR(jd)
	el := planet.WherePlanet(-1, 0, jd)
	eb := planet.WherePlanet(-1, 1, jd)
	er := planet.WherePlanet(-1, 2, jd)
	y := r*Cos(b)*Sin(l) - er*Cos(eb)*Sin(el)
	return y
}
func ASaturnZ(jd float64) float64 {
	//l := SaturnL(jd)
	b := SaturnB(jd)
	r := SaturnR(jd)
	//	el := planet.WherePlanet(-1, 0, jd)
	eb := planet.WherePlanet(-1, 1, jd)
	er := planet.WherePlanet(-1, 2, jd)
	z := r*Sin(b) - er*Sin(eb)
	return z
}

func ASaturnXYZ(jd float64) (float64, float64, float64) {
	l := SaturnL(jd)
	b := SaturnB(jd)
	r := SaturnR(jd)
	el := planet.WherePlanet(-1, 0, jd)
	eb := planet.WherePlanet(-1, 1, jd)
	er := planet.WherePlanet(-1, 2, jd)
	x := r*Cos(b)*Cos(l) - er*Cos(eb)*Cos(el)
	y := r*Cos(b)*Sin(l) - er*Cos(eb)*Sin(el)
	z := r*Sin(b) - er*Sin(eb)
	return x, y, z
}

func SaturnApparentRa(jd float64) float64 {
	lo, bo := SaturnApparentLoBo(jd)
	eps := TrueObliquity(jd)
	ra := math.Atan2((Sin(lo)*Cos(eps) - Tan(bo)*Sin(eps)), Cos(lo))
	ra = ra * 180 / math.Pi
	return Limit360(ra)
}
func SaturnApparentDec(jd float64) float64 {
	lo, bo := SaturnApparentLoBo(jd)
	eps := TrueObliquity(jd)
	dec := ArcSin(Sin(bo)*Cos(eps) + Cos(bo)*Sin(eps)*Sin(lo))
	return dec
}

func SaturnApparentRaDec(jd float64) (float64, float64) {
	lo, bo := SaturnApparentLoBo(jd)
	eps := TrueObliquity(jd)
	ra := math.Atan2((Sin(lo)*Cos(eps) - Tan(bo)*Sin(eps)), Cos(lo))
	ra = ra * 180 / math.Pi
	dec := ArcSin(Sin(bo)*Cos(eps) + Cos(bo)*Sin(eps)*Sin(lo))
	return Limit360(ra), dec
}

func EarthSaturnAway(jd float64) float64 {
	x, y, z := ASaturnXYZ(jd)
	to := math.Sqrt(x*x + y*y + z*z)
	return to
}

func SaturnApparentLo(jd float64) float64 {
	x, y, z := ASaturnXYZ(jd)
	to := 0.0057755183 * math.Sqrt(x*x+y*y+z*z)
	x, y, z = ASaturnXYZ(jd - to)
	lo := math.Atan2(y, x)
	bo := math.Atan2(z, math.Sqrt(x*x+y*y))
	lo = lo * 180 / math.Pi
	bo = bo * 180 / math.Pi
	lo = Limit360(lo)
	//lo-=GXCLo(lo,bo,jd)/3600;
	//bo+=GXCBo(lo,bo,jd);
	lo += Nutation2000Bi(jd)
	return lo
}

func SaturnApparentBo(jd float64) float64 {
	x, y, z := ASaturnXYZ(jd)
	to := 0.0057755183 * math.Sqrt(x*x+y*y+z*z)
	x, y, z = ASaturnXYZ(jd - to)
	//lo := math.Atan2(y, x)
	bo := math.Atan2(z, math.Sqrt(x*x+y*y))
	//lo = lo * 180 / math.Pi
	bo = bo * 180 / math.Pi
	//lo+=GXCLo(lo,bo,jd);
	//bo+=GXCBo(lo,bo,jd)/3600;
	//lo+=Nutation2000Bi(jd);
	return bo
}

func SaturnApparentLoBo(jd float64) (float64, float64) {
	x, y, z := ASaturnXYZ(jd)
	to := 0.0057755183 * math.Sqrt(x*x+y*y+z*z)
	x, y, z = ASaturnXYZ(jd - to)
	lo := math.Atan2(y, x)
	bo := math.Atan2(z, math.Sqrt(x*x+y*y))
	lo = lo * 180 / math.Pi
	bo = bo * 180 / math.Pi
	lo = Limit360(lo)
	//lo-=GXCLo(lo,bo,jd)/3600;
	//bo+=GXCBo(lo,bo,jd);
	lo += Nutation2000Bi(jd)
	return lo, bo
}

func SaturnMag(jd float64) float64 {
	return SaturnMagN(jd, -1)
}

func SaturnHeight(jde, lon, lat, timezone float64) float64 {
	// 转换为世界时
	utcJde := jde - timezone/24.0
	// 计算视恒星时
	ra, dec := SaturnApparentRaDec(TD2UT(utcJde, true))
	st := Limit360(ApparentSiderealTime(utcJde)*15 + lon)
	// 计算时角
	hourAngle := Limit360(st - ra)
	// 高度角、时角与天球座标三角转换公式
	// sin(h)=sin(lat)*sin(dec)+cos(dec)*cos(lat)*cos(hourAngle)
	sinHeight := Sin(lat)*Sin(dec) + Cos(dec)*Cos(lat)*Cos(hourAngle)
	return ArcSin(sinHeight)
}

func SaturnAzimuth(jde, lon, lat, timezone float64) float64 {
	// 转换为世界时
	utcJde := jde - timezone/24.0
	// 计算视恒星时
	ra, dec := SaturnApparentRaDec(TD2UT(utcJde, true))
	st := Limit360(ApparentSiderealTime(utcJde)*15 + lon)
	// 计算时角
	hourAngle := Limit360(st - ra)
	// 三角转换公式
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

func SaturnHourAngle(jd, lon, timezone float64) float64 {
	siderealLongitude := Limit360(ApparentSiderealTime(jd-timezone/24)*15 + lon)
	hourAngle := siderealLongitude - SaturnApparentRa(TD2UT(jd-timezone/24.0, true))
	if hourAngle < 0 {
		hourAngle += 360
	}
	return hourAngle
}

func SaturnCulminationTime(jde, lon, timezone float64) float64 {
	//jde 世界时，非力学时，当地时区 0时，无需转换力学时
	//ra,dec 瞬时天球座标，非J2000等时间天球坐标
	jde = math.Floor(jde) + 0.5
	estimateJD := jde + Limit360(360-SaturnHourAngle(jde, lon, timezone))/15.0/24.0*0.99726851851851851851
	normalizedHourAngle := func(jde, lon, timezone float64) float64 {
		currentHourAngle := SaturnHourAngle(jde, lon, timezone)
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

func SaturnRiseTime(jd, lon, lat, timezone, aeroCorrection, observerHeight float64) (float64, error) {
	return saturnRiseDown(jd, lon, lat, timezone, aeroCorrection, observerHeight, true)
}

func SaturnSetTime(jd, lon, lat, timezone, aeroCorrection, observerHeight float64) (float64, error) {
	return saturnRiseDown(jd, lon, lat, timezone, aeroCorrection, observerHeight, false)
}

func saturnRiseDown(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, isRise bool) (float64, error) {
	return planetRiseDown(jd, lon, lat, timezone, aeroCorrection, observerHeight, isRise, SaturnCulminationTime, SaturnHeight, SaturnApparentDec)
}
