package basic

import (
	"math"

	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

func SaturnL(JD float64) float64 {
	return planet.WherePlanet(5, 0, JD)
}

func SaturnB(JD float64) float64 {
	return planet.WherePlanet(5, 1, JD)
}
func SaturnR(JD float64) float64 {
	return planet.WherePlanet(5, 2, JD)
}
func ASaturnX(JD float64) float64 {
	l := SaturnL(JD)
	b := SaturnB(JD)
	r := SaturnR(JD)
	el := planet.WherePlanet(-1, 0, JD)
	eb := planet.WherePlanet(-1, 1, JD)
	er := planet.WherePlanet(-1, 2, JD)
	x := r*Cos(b)*Cos(l) - er*Cos(eb)*Cos(el)
	return x
}

func ASaturnY(JD float64) float64 {

	l := SaturnL(JD)
	b := SaturnB(JD)
	r := SaturnR(JD)
	el := planet.WherePlanet(-1, 0, JD)
	eb := planet.WherePlanet(-1, 1, JD)
	er := planet.WherePlanet(-1, 2, JD)
	y := r*Cos(b)*Sin(l) - er*Cos(eb)*Sin(el)
	return y
}
func ASaturnZ(JD float64) float64 {
	//l := SaturnL(JD)
	b := SaturnB(JD)
	r := SaturnR(JD)
	//	el := planet.WherePlanet(-1, 0, JD)
	eb := planet.WherePlanet(-1, 1, JD)
	er := planet.WherePlanet(-1, 2, JD)
	z := r*Sin(b) - er*Sin(eb)
	return z
}

func ASaturnXYZ(JD float64) (float64, float64, float64) {
	l := SaturnL(JD)
	b := SaturnB(JD)
	r := SaturnR(JD)
	el := planet.WherePlanet(-1, 0, JD)
	eb := planet.WherePlanet(-1, 1, JD)
	er := planet.WherePlanet(-1, 2, JD)
	x := r*Cos(b)*Cos(l) - er*Cos(eb)*Cos(el)
	y := r*Cos(b)*Sin(l) - er*Cos(eb)*Sin(el)
	z := r*Sin(b) - er*Sin(eb)
	return x, y, z
}

func SaturnApparentRa(JD float64) float64 {
	lo, bo := SaturnApparentLoBo(JD)
	sita := Sita(JD)
	ra := math.Atan2((Sin(lo)*Cos(sita) - Tan(bo)*Sin(sita)), Cos(lo))
	ra = ra * 180 / math.Pi
	return Limit360(ra)
}
func SaturnApparentDec(JD float64) float64 {
	lo, bo := SaturnApparentLoBo(JD)
	sita := Sita(JD)
	dec := ArcSin(Sin(bo)*Cos(sita) + Cos(bo)*Sin(sita)*Sin(lo))
	return dec
}

func SaturnApparentRaDec(JD float64) (float64, float64) {
	lo, bo := SaturnApparentLoBo(JD)
	sita := Sita(JD)
	ra := math.Atan2((Sin(lo)*Cos(sita) - Tan(bo)*Sin(sita)), Cos(lo))
	ra = ra * 180 / math.Pi
	dec := ArcSin(Sin(bo)*Cos(sita) + Cos(bo)*Sin(sita)*Sin(lo))
	return Limit360(ra), dec
}

func EarthSaturnAway(JD float64) float64 {
	x, y, z := ASaturnXYZ(JD)
	to := math.Sqrt(x*x + y*y + z*z)
	return to
}

func SaturnApparentLo(JD float64) float64 {
	x, y, z := ASaturnXYZ(JD)
	to := 0.0057755183 * math.Sqrt(x*x+y*y+z*z)
	x, y, z = ASaturnXYZ(JD - to)
	lo := math.Atan2(y, x)
	bo := math.Atan2(z, math.Sqrt(x*x+y*y))
	lo = lo * 180 / math.Pi
	bo = bo * 180 / math.Pi
	lo = Limit360(lo)
	//lo-=GXCLo(lo,bo,JD)/3600;
	//bo+=GXCBo(lo,bo,JD);
	lo += Nutation2000Bi(JD)
	return lo
}

func SaturnApparentBo(JD float64) float64 {
	x, y, z := ASaturnXYZ(JD)
	to := 0.0057755183 * math.Sqrt(x*x+y*y+z*z)
	x, y, z = ASaturnXYZ(JD - to)
	//lo := math.Atan2(y, x)
	bo := math.Atan2(z, math.Sqrt(x*x+y*y))
	//lo = lo * 180 / math.Pi
	bo = bo * 180 / math.Pi
	//lo+=GXCLo(lo,bo,JD);
	//bo+=GXCBo(lo,bo,JD)/3600;
	//lo+=Nutation2000Bi(JD);
	return bo
}

func SaturnApparentLoBo(JD float64) (float64, float64) {
	x, y, z := ASaturnXYZ(JD)
	to := 0.0057755183 * math.Sqrt(x*x+y*y+z*z)
	x, y, z = ASaturnXYZ(JD - to)
	lo := math.Atan2(y, x)
	bo := math.Atan2(z, math.Sqrt(x*x+y*y))
	lo = lo * 180 / math.Pi
	bo = bo * 180 / math.Pi
	lo = Limit360(lo)
	//lo-=GXCLo(lo,bo,JD)/3600;
	//bo+=GXCBo(lo,bo,JD);
	lo += Nutation2000Bi(JD)
	return lo, bo
}

func SaturnMag(JD float64) float64 {
	AwaySun := SaturnR(JD)
	AwayEarth := EarthSaturnAway(JD)
	Away := planet.WherePlanet(-1, 2, JD)
	i := (AwaySun*AwaySun + AwayEarth*AwayEarth - Away*Away) / (2 * AwaySun * AwayEarth)
	i = ArcCos(i)
	Mag := -8.68 + 5*math.Log10(AwaySun*AwayEarth) + 0.044*i - 2.6*Sin(math.Abs(SaturnRingB(JD))) + 1.25*Sin(math.Abs(SaturnRingB(JD)))*Sin(math.Abs(SaturnRingB(JD)))
	return FloatRound(Mag, 2)
}

func SaturnRingB(JD float64) float64 {
	T := (JD - 2451545) / 36525
	i := 28.075216 - 0.012998*T + 0.000004*T*T
	omi := 169.508470 + 1.394681*T + 0.000412*T*T
	lo, bo := SaturnApparentLoBo(JD)
	B := Sin(i)*Cos(bo)*Sin(lo-omi) - Cos(i)*Cos(bo)
	return ArcSin(B)
}

func SaturnHeight(jde, lon, lat, timezone float64) float64 {
	// 转换为世界时
	utcJde := jde - timezone/24.0
	// 计算视恒星时
	ra, dec := SaturnApparentRaDec(TD2UT(utcJde, true))
	st := Limit360(ApparentSiderealTime(utcJde)*15 + lon)
	// 计算时角
	H := Limit360(st - ra)
	// 高度角、时角与天球座标三角转换公式
	// sin(h)=sin(lat)*sin(dec)+cos(dec)*cos(lat)*cos(H)
	sinHeight := Sin(lat)*Sin(dec) + Cos(dec)*Cos(lat)*Cos(H)
	return ArcSin(sinHeight)
}

func SaturnAzimuth(jde, lon, lat, timezone float64) float64 {
	// 转换为世界时
	utcJde := jde - timezone/24.0
	// 计算视恒星时
	ra, dec := SaturnApparentRaDec(TD2UT(utcJde, true))
	st := Limit360(ApparentSiderealTime(utcJde)*15 + lon)
	// 计算时角
	H := Limit360(st - ra)
	// 三角转换公式
	tanAzimuth := Sin(H) / (Cos(H)*Sin(lat) - Tan(dec)*Cos(lat))
	Azimuth := ArcTan(tanAzimuth)
	if Azimuth < 0 {
		if H/15 < 12 {
			return Azimuth + 360
		}
		return Azimuth + 180
	}
	if H/15 < 12 {
		return Azimuth + 180
	}
	return Azimuth
}

func SaturnHourAngle(JD, Lon, TZ float64) float64 {
	startime := Limit360(ApparentSiderealTime(JD-TZ/24)*15 + Lon)
	timeangle := startime - SaturnApparentRa(TD2UT(JD-TZ/24.0, true))
	if timeangle < 0 {
		timeangle += 360
	}
	return timeangle
}

func SaturnCulminationTime(jde, lon, timezone float64) float64 {
	//jde 世界时，非力学时，当地时区 0时，无需转换力学时
	//ra,dec 瞬时天球座标，非J2000等时间天球坐标
	jde = math.Floor(jde) + 0.5
	JD1 := jde + Limit360(360-SaturnHourAngle(jde, lon, timezone))/15.0/24.0*0.99726851851851851851
	limitHA := func(jde, lon, timezone float64) float64 {
		ha := SaturnHourAngle(jde, lon, timezone)
		if ha < 180 {
			ha += 360
		}
		return ha
	}
	for {
		JD0 := JD1
		stDegree := limitHA(JD0, lon, timezone) - 360
		stDegreep := (limitHA(JD0+0.000005, lon, timezone) - limitHA(JD0-0.000005, lon, timezone)) / 0.00001
		JD1 = JD0 - stDegree/stDegreep
		if math.Abs(JD1-JD0) <= 0.00001 {
			break
		}
	}
	return JD1
}

func SaturnRiseTime(JD, Lon, Lat, TZ, ZS, HEI float64) float64 {
	return saturnRiseDown(JD, Lon, Lat, TZ, ZS, HEI, true)
}

func SaturnDownTime(JD, Lon, Lat, TZ, ZS, HEI float64) float64 {
	return saturnRiseDown(JD, Lon, Lat, TZ, ZS, HEI, false)
}

func saturnRiseDown(JD, Lon, Lat, TZ, ZS, HEI float64, isRise bool) float64 {
	var An float64
	JD = math.Floor(JD) + 0.5
	ntz := math.Round(Lon / 15)
	if ZS != 0 {
		An = -0.8333
	}
	An = An - HeightDegreeByLat(HEI, Lat)
	tztime := SaturnCulminationTime(JD, Lon, ntz)
	if SaturnHeight(tztime, Lon, Lat, ntz) < An {
		return -2 //极夜
	}
	if SaturnHeight(tztime-0.5, Lon, Lat, ntz) > An {
		return -1 //极昼
	}
	dec := HSunApparentDec(TD2UT(tztime-ntz/24, true))
	//(sin(ho)-sin(φ)*sin(δ2))/(cos(φ)*cos(δ2))
	tmp := (Sin(An) - Sin(dec)*Sin(Lat)) / (Cos(dec) * Cos(Lat))
	var rise float64
	if math.Abs(tmp) <= 1 {
		rzsc := ArcCos(tmp) / 15
		if isRise {
			rise = tztime - rzsc/24 - 25.0/24.0/60.0
		} else {
			rise = tztime + rzsc/24 - 25.0/24.0/60.0
		}
	} else {
		rise = tztime
		i := 0
		//TODO:使用二分法计算
		for SaturnHeight(rise, Lon, Lat, ntz) > An {
			i++
			if isRise {
				rise -= 15.0 / 60.0 / 24.0
			} else {
				rise += 15.0 / 60.0 / 24.0
			}
			if i > 48 {
				break
			}
		}
	}
	JD1 := rise
	for {
		JD0 := JD1
		stDegree := SaturnHeight(JD0, Lon, Lat, ntz) - An
		stDegreep := (SaturnHeight(JD0+0.000005, Lon, Lat, ntz) - SaturnHeight(JD0-0.000005, Lon, Lat, ntz)) / 0.00001
		JD1 = JD0 - stDegree/stDegreep
		if math.Abs(JD1-JD0) <= 0.00001 {
			break
		}
	}
	return JD1 - ntz/24 + TZ/24
}

// Pos

const SATURN_S_PERIOD = 1 / ((1 / 365.256363004) - (1 / 10759.0))

func saturnConjunction(jde, degree float64, next uint8) float64 {
	//0=last 1=next
	decSub := func(jde float64, degree float64, filter bool) float64 {
		sub := Limit360(Limit360(SaturnApparentLo(jde)-HSunApparentLo(jde)) - degree)
		if filter {
			if sub > 180 {
				sub -= 360
			}
			if sub < -180 {
				sub += 360
			}
		}
		return sub
	}
	dayCost := SATURN_S_PERIOD / 360
	nowSub := decSub(jde, degree, false)
	if next == 0 {
		jde -= (360 - nowSub) * dayCost
	} else {
		jde += dayCost * nowSub
	}
	JD1 := jde
	for {
		JD0 := JD1
		stDegree := decSub(JD0, degree, true)
		stDegreep := (decSub(JD0+0.000005, degree, true) - decSub(JD0-0.000005, degree, true)) / 0.00001
		JD1 = JD0 - stDegree/stDegreep
		if math.Abs(JD1-JD0) <= 0.00001 {
			break
		}
	}
	return TD2UT(JD1, false)
}

func LastSaturnConjunction(jde float64) float64 {
	return saturnConjunction(jde, 0, 0)
}

func NextSaturnConjunction(jde float64) float64 {
	return saturnConjunction(jde, 0, 1)
}

func LastSaturnOpposition(jde float64) float64 {
	return saturnConjunction(jde, 180, 0)
}

func NextSaturnOpposition(jde float64) float64 {
	return saturnConjunction(jde, 180, 1)
}

func NextSaturnEasternQuadrature(jde float64) float64 {
	return saturnConjunction(jde, 90, 1)
}

func LastSaturnEasternQuadrature(jde float64) float64 {
	return saturnConjunction(jde, 90, 0)
}

func NextSaturnWesternQuadrature(jde float64) float64 {
	return saturnConjunction(jde, 270, 1)
}

func LastSaturnWesternQuadrature(jde float64) float64 {
	return saturnConjunction(jde, 270, 0)
}

func saturnRetrograde(jde float64, isLeft bool) float64 {
	//0=last 1=next
	decSub := func(jde float64, val float64) float64 {
		sub := SaturnApparentRa(jde+val) - SaturnApparentRa(jde-val)
		if sub > 180 {
			sub -= 360
		}
		if sub < -180 {
			sub += 360
		}
		return sub / (2 * val)
	}
	jde = NextSaturnOpposition(jde)
	if isLeft {
		jde -= 60
	} else {
		jde += 60
	}
	for {
		nowSub := decSub(jde, 1.0/86400.0)
		if math.Abs(nowSub) > 0.55 {
			jde += 2
			continue
		}
		break
	}
	JD1 := jde
	for {
		JD0 := JD1
		stDegree := decSub(JD0, 2.0/86400.0)
		stDegreep := (decSub(JD0+15.0/86400.0, 2.0/86400.0) - decSub(JD0-15.0/86400.0, 2.0/86400.0)) / (30.0 / 86400.0)
		JD1 = JD0 - stDegree/stDegreep
		if math.Abs(JD1-JD0) <= 30.0/86400.0 {
			break
		}
	}
	JD1 = JD1 - 15.0/86400.0
	min := JD1
	minRa := 100.0
	for i := 0.0; i < 60.0; i++ {
		tmp := decSub(JD1+i*0.5/86400.0, 0.5/86400.0)
		if math.Abs(tmp) < math.Abs(minRa) {
			minRa = tmp
			min = JD1 + i*0.5/86400.0
		}
	}
	return TD2UT(min, false)
}

func NextSaturnRetrogradeToPrograde(jde float64) float64 {
	date := saturnRetrograde(jde, false)
	if date < jde {
		op := NextSaturnOpposition(jde)
		return saturnRetrograde(op+10, false)
	}
	return date
}

func LastSaturnRetrogradeToPrograde(jde float64) float64 {
	jde = LastSaturnOpposition(jde) - 10
	date := saturnRetrograde(jde, false)
	if date > jde {
		op := LastSaturnOpposition(jde)
		return saturnRetrograde(op-10, false)
	}
	return date
}

func NextSaturnProgradeToRetrograde(jde float64) float64 {
	date := saturnRetrograde(jde, true)
	if date < jde {
		op := NextSaturnOpposition(jde)
		return saturnRetrograde(op+10, true)
	}
	return date
}

func LastSaturnProgradeToRetrograde(jde float64) float64 {
	jde = LastSaturnOpposition(jde) - 10
	date := saturnRetrograde(jde, true)
	if date > jde {
		op := LastSaturnOpposition(jde)
		return saturnRetrograde(op-10, true)
	}
	return date
}
