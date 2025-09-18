package basic

import (
	"math"

	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

func MarsL(JD float64) float64 {
	return planet.WherePlanet(3, 0, JD)
}

func MarsB(JD float64) float64 {
	return planet.WherePlanet(3, 1, JD)
}
func MarsR(JD float64) float64 {
	return planet.WherePlanet(3, 2, JD)
}
func AMarsX(JD float64) float64 {
	l := MarsL(JD)
	b := MarsB(JD)
	r := MarsR(JD)
	el := planet.WherePlanet(-1, 0, JD)
	eb := planet.WherePlanet(-1, 1, JD)
	er := planet.WherePlanet(-1, 2, JD)
	x := r*Cos(b)*Cos(l) - er*Cos(eb)*Cos(el)
	return x
}

func AMarsY(JD float64) float64 {

	l := MarsL(JD)
	b := MarsB(JD)
	r := MarsR(JD)
	el := planet.WherePlanet(-1, 0, JD)
	eb := planet.WherePlanet(-1, 1, JD)
	er := planet.WherePlanet(-1, 2, JD)
	y := r*Cos(b)*Sin(l) - er*Cos(eb)*Sin(el)
	return y
}
func AMarsZ(JD float64) float64 {
	//l := MarsL(JD)
	b := MarsB(JD)
	r := MarsR(JD)
	//	el := planet.WherePlanet(-1, 0, JD)
	eb := planet.WherePlanet(-1, 1, JD)
	er := planet.WherePlanet(-1, 2, JD)
	z := r*Sin(b) - er*Sin(eb)
	return z
}

func AMarsXYZ(JD float64) (float64, float64, float64) {
	l := MarsL(JD)
	b := MarsB(JD)
	r := MarsR(JD)
	el := planet.WherePlanet(-1, 0, JD)
	eb := planet.WherePlanet(-1, 1, JD)
	er := planet.WherePlanet(-1, 2, JD)
	x := r*Cos(b)*Cos(l) - er*Cos(eb)*Cos(el)
	y := r*Cos(b)*Sin(l) - er*Cos(eb)*Sin(el)
	z := r*Sin(b) - er*Sin(eb)
	return x, y, z
}

func MarsApparentRa(JD float64) float64 {
	lo, bo := MarsApparentLoBo(JD)
	sita := Sita(JD)
	ra := math.Atan2((Sin(lo)*Cos(sita) - Tan(bo)*Sin(sita)), Cos(lo))
	ra = ra * 180 / math.Pi
	return Limit360(ra)
}
func MarsApparentDec(JD float64) float64 {
	lo, bo := MarsApparentLoBo(JD)
	sita := Sita(JD)
	dec := ArcSin(Sin(bo)*Cos(sita) + Cos(bo)*Sin(sita)*Sin(lo))
	return dec
}

func MarsApparentRaDec(JD float64) (float64, float64) {
	lo, bo := MarsApparentLoBo(JD)
	sita := Sita(JD)
	ra := math.Atan2((Sin(lo)*Cos(sita) - Tan(bo)*Sin(sita)), Cos(lo))
	ra = ra * 180 / math.Pi
	dec := ArcSin(Sin(bo)*Cos(sita) + Cos(bo)*Sin(sita)*Sin(lo))
	return Limit360(ra), dec
}

func EarthMarsAway(JD float64) float64 {
	x, y, z := AMarsXYZ(JD)
	to := math.Sqrt(x*x + y*y + z*z)
	return to
}

func MarsApparentLo(JD float64) float64 {
	x, y, z := AMarsXYZ(JD)
	to := 0.0057755183 * math.Sqrt(x*x+y*y+z*z)
	x, y, z = AMarsXYZ(JD - to)
	lo := math.Atan2(y, x)
	//bo := math.Atan2(z, math.Sqrt(x*x+y*y))
	lo = lo * 180.0 / math.Pi
	//bo = bo * 180 / math.Pi
	lo = Limit360(lo) + Nutation2000Bi(JD)
	//lo-=GXCLo(lo,bo,JD)/3600;
	//bo+=GXCBo(lo,bo,JD);
	return lo
}

func MarsApparentBo(JD float64) float64 {
	x, y, z := AMarsXYZ(JD)
	to := 0.0057755183 * math.Sqrt(x*x+y*y+z*z)
	x, y, z = AMarsXYZ(JD - to)
	//lo := math.Atan2(y, x)
	bo := math.Atan2(z, math.Sqrt(x*x+y*y))
	//lo = lo * 180 / math.Pi
	bo = bo * 180 / math.Pi
	//lo+=GXCLo(lo,bo,JD);
	//bo+=GXCBo(lo,bo,JD)/3600;
	//lo+=Nutation2000Bi(JD);
	return bo
}

func MarsApparentLoBo(JD float64) (float64, float64) {
	x, y, z := AMarsXYZ(JD)
	to := 0.0057755183 * math.Sqrt(x*x+y*y+z*z)
	x, y, z = AMarsXYZ(JD - to)
	lo := math.Atan2(y, x)
	bo := math.Atan2(z, math.Sqrt(x*x+y*y))
	lo = lo * 180 / math.Pi
	bo = bo * 180 / math.Pi
	lo = Limit360(lo)
	//lo -= GXCLo(lo, bo, JD) / 3600
	//bo += GXCBo(lo, bo, JD)
	lo += Nutation2000Bi(JD)
	return lo, bo
}

func MarsTrueLoBo(JD float64) (float64, float64) {
	x, y, z := AMarsXYZ(JD)
	to := 0.0057755183 * math.Sqrt(x*x+y*y+z*z)
	x, y, z = AMarsXYZ(JD - to)
	lo := math.Atan2(y, x)
	bo := math.Atan2(z, math.Sqrt(x*x+y*y))
	lo = lo * 180 / math.Pi
	bo = bo * 180 / math.Pi
	lo = Limit360(lo)
	return lo, bo
}

func MarsTrueLo(JD float64) float64 {
	x, y, _ := AMarsXYZ(JD)
	lo := math.Atan2(y, x)
	lo = lo * 180 / math.Pi
	lo = Limit360(lo)
	return lo
}

func MarsMag(JD float64) float64 {
	AwaySun := MarsR(JD)
	AwayEarth := EarthMarsAway(JD)
	Away := planet.WherePlanet(-1, 2, JD)
	i := (AwaySun*AwaySun + AwayEarth*AwayEarth - Away*Away) / (2 * AwaySun * AwayEarth)
	i = ArcCos(i)
	Mag := -1.52 + 5*math.Log10(AwaySun*AwayEarth) + 0.016*i
	return FloatRound(Mag, 2)
}

func MarsHeight(jde, lon, lat, timezone float64) float64 {
	// 转换为世界时
	utcJde := jde - timezone/24.0
	// 计算视恒星时
	ra, dec := MarsApparentRaDec(TD2UT(utcJde, true))
	st := Limit360(ApparentSiderealTime(utcJde)*15 + lon)
	// 计算时角
	H := Limit360(st - ra)
	// 高度角、时角与天球座标三角转换公式
	// sin(h)=sin(lat)*sin(dec)+cos(dec)*cos(lat)*cos(H)
	sinHeight := Sin(lat)*Sin(dec) + Cos(dec)*Cos(lat)*Cos(H)
	return ArcSin(sinHeight)
}

func MarsAzimuth(jde, lon, lat, timezone float64) float64 {
	// 转换为世界时
	utcJde := jde - timezone/24.0
	// 计算视恒星时
	ra, dec := MarsApparentRaDec(TD2UT(utcJde, true))
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

func MarsHourAngle(JD, Lon, TZ float64) float64 {
	startime := Limit360(ApparentSiderealTime(JD-TZ/24)*15 + Lon)
	timeangle := startime - MarsApparentRa(TD2UT(JD-TZ/24.0, true))
	if timeangle < 0 {
		timeangle += 360
	}
	return timeangle
}

func MarsCulminationTime(jde, lon, timezone float64) float64 {
	//jde 世界时，非力学时，当地时区 0时，无需转换力学时
	//ra,dec 瞬时天球座标，非J2000等时间天球坐标
	jde = math.Floor(jde) + 0.5
	JD1 := jde + Limit360(360-MarsHourAngle(jde, lon, timezone))/15.0/24.0*0.99726851851851851851
	limitHA := func(jde, lon, timezone float64) float64 {
		ha := MarsHourAngle(jde, lon, timezone)
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

func MarsRiseTime(JD, Lon, Lat, TZ, ZS, HEI float64) float64 {
	return marsRiseDown(JD, Lon, Lat, TZ, ZS, HEI, true)
}

func MarsDownTime(JD, Lon, Lat, TZ, ZS, HEI float64) float64 {
	return marsRiseDown(JD, Lon, Lat, TZ, ZS, HEI, false)
}

func marsRiseDown(JD, Lon, Lat, TZ, ZS, HEI float64, isRise bool) float64 {
	var An float64
	JD = math.Floor(JD) + 0.5
	ntz := math.Round(Lon / 15)
	if ZS != 0 {
		An = -0.8333
	}
	An = An - HeightDegreeByLat(HEI, Lat)
	tztime := MarsCulminationTime(JD, Lon, ntz)
	if MarsHeight(tztime, Lon, Lat, ntz) < An {
		return -2 //极夜
	}
	if MarsHeight(tztime-0.5, Lon, Lat, ntz) > An {
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
		for MarsHeight(rise, Lon, Lat, ntz) > An {
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
		stDegree := MarsHeight(JD0, Lon, Lat, ntz) - An
		stDegreep := (MarsHeight(JD0+0.000005, Lon, Lat, ntz) - MarsHeight(JD0-0.000005, Lon, Lat, ntz)) / 0.00001
		JD1 = JD0 - stDegree/stDegreep
		if math.Abs(JD1-JD0) <= 0.00001 {
			break
		}
	}
	return JD1 - ntz/24 + TZ/24
}

// Pos

const MARS_S_PERIOD = 1 / ((1 / 365.256363004) - (1 / 686.98))

func marsConjunction(jde, degree float64, next uint8) float64 {
	//0=last 1=next
	decSub := func(jde float64, degree float64, filter bool) float64 {
		sub := Limit360(Limit360(MarsApparentLo(jde)-HSunApparentLo(jde)) - degree)
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
	dayCost := MARS_S_PERIOD / 360
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

func LastMarsConjunction(jde float64) float64 {
	return marsConjunction(jde, 0, 0)
}

func NextMarsConjunction(jde float64) float64 {
	return marsConjunction(jde, 0, 1)
}

func LastMarsOpposition(jde float64) float64 {
	return marsConjunction(jde, 180, 0)
}

func NextMarsOpposition(jde float64) float64 {
	return marsConjunction(jde, 180, 1)
}

func NextMarsEasternQuadrature(jde float64) float64 {
	return marsConjunction(jde, 90, 1)
}

func LastMarsEasternQuadrature(jde float64) float64 {
	return marsConjunction(jde, 90, 0)
}

func NextMarsWesternQuadrature(jde float64) float64 {
	return marsConjunction(jde, 270, 1)
}

func LastMarsWesternQuadrature(jde float64) float64 {
	return marsConjunction(jde, 270, 0)
}

func marsRetrograde(jde float64, isLeft bool) float64 {
	//0=last 1=next
	decSub := func(jde float64, val float64) float64 {
		sub := MarsApparentRa(jde+val) - MarsApparentRa(jde-val)
		if sub > 180 {
			sub -= 360
		}
		if sub < -180 {
			sub += 360
		}
		return sub / (2 * val)
	}
	jde = NextMarsOpposition(jde)
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

func NextMarsRetrogradeToPrograde(jde float64) float64 {
	date := marsRetrograde(jde, false)
	if date < jde {
		op := NextMarsOpposition(jde)
		return marsRetrograde(op+10, false)
	}
	return date
}

func LastMarsRetrogradeToPrograde(jde float64) float64 {
	jde = LastMarsOpposition(jde) - 10
	date := marsRetrograde(jde, false)
	if date > jde {
		op := LastMarsOpposition(jde)
		return marsRetrograde(op-10, false)
	}
	return date
}

func NextMarsProgradeToRetrograde(jde float64) float64 {
	date := marsRetrograde(jde, true)
	if date < jde {
		op := NextMarsOpposition(jde)
		return marsRetrograde(op+10, true)
	}
	return date
}

func LastMarsProgradeToRetrograde(jde float64) float64 {
	jde = LastMarsOpposition(jde) - 10
	date := marsRetrograde(jde, true)
	if date > jde {
		op := LastMarsOpposition(jde)
		return marsRetrograde(op-10, true)
	}
	return date
}
