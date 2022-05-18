package basic

import (
	"math"

	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

func MercuryL(JD float64) float64 {
	return planet.WherePlanet(1, 0, JD)
}

func MercuryB(JD float64) float64 {
	return planet.WherePlanet(1, 1, JD)
}
func MercuryR(JD float64) float64 {
	return planet.WherePlanet(1, 2, JD)
}
func AMercuryX(JD float64) float64 {
	l := MercuryL(JD)
	b := MercuryB(JD)
	r := MercuryR(JD)
	el := planet.WherePlanet(-1, 0, JD)
	eb := planet.WherePlanet(-1, 1, JD)
	er := planet.WherePlanet(-1, 2, JD)
	x := r*Cos(b)*Cos(l) - er*Cos(eb)*Cos(el)
	return x
}

func AMercuryY(JD float64) float64 {

	l := MercuryL(JD)
	b := MercuryB(JD)
	r := MercuryR(JD)
	el := planet.WherePlanet(-1, 0, JD)
	eb := planet.WherePlanet(-1, 1, JD)
	er := planet.WherePlanet(-1, 2, JD)
	y := r*Cos(b)*Sin(l) - er*Cos(eb)*Sin(el)
	return y
}
func AMercuryZ(JD float64) float64 {
	//l := MercuryL(JD)
	b := MercuryB(JD)
	r := MercuryR(JD)
	//	el := planet.WherePlanet(-1, 0, JD)
	eb := planet.WherePlanet(-1, 1, JD)
	er := planet.WherePlanet(-1, 2, JD)
	z := r*Sin(b) - er*Sin(eb)
	return z
}

func AMercuryXYZ(JD float64) (float64, float64, float64) {
	l := MercuryL(JD)
	b := MercuryB(JD)
	r := MercuryR(JD)
	el := planet.WherePlanet(-1, 0, JD)
	eb := planet.WherePlanet(-1, 1, JD)
	er := planet.WherePlanet(-1, 2, JD)
	x := r*Cos(b)*Cos(l) - er*Cos(eb)*Cos(el)
	y := r*Cos(b)*Sin(l) - er*Cos(eb)*Sin(el)
	z := r*Sin(b) - er*Sin(eb)
	return x, y, z
}

func MercuryApparentRa(JD float64) float64 {
	lo, bo := MercuryApparentLoBo(JD)
	return LoToRa(JD, lo, bo)
}
func MercuryApparentDec(JD float64) float64 {
	lo, bo := MercuryApparentLoBo(JD)
	sita := Sita(JD)
	dec := ArcSin(Sin(bo)*Cos(sita) + Cos(bo)*Sin(sita)*Sin(lo))
	return dec
}

func MercuryApparentRaDec(JD float64) (float64, float64) {
	lo, bo := MercuryApparentLoBo(JD)
	return LoBoToRaDec(JD, lo, bo)
}

func EarthMercuryAway(JD float64) float64 {
	x, y, z := AMercuryXYZ(JD)
	to := math.Sqrt(x*x + y*y + z*z)
	return to
}

func MercuryApparentLo(JD float64) float64 {
	x, y, z := AMercuryXYZ(JD)
	to := 0.0057755183 * math.Sqrt(x*x+y*y+z*z)
	x, y, z = AMercuryXYZ(JD - to)
	lo := math.Atan2(y, x)
	bo := math.Atan2(z, math.Sqrt(x*x+y*y))
	lo = lo * 180 / math.Pi
	bo = bo * 180 / math.Pi
	lo = Limit360(lo)
	//lo-=GXCLo(lo,bo,JD)/3600;
	//bo+=GXCBo(lo,bo,JD);
	lo += HJZD(JD)
	return lo
}

func MercuryApparentBo(JD float64) float64 {
	x, y, z := AMercuryXYZ(JD)
	to := 0.0057755183 * math.Sqrt(x*x+y*y+z*z)
	x, y, z = AMercuryXYZ(JD - to)
	//lo := math.Atan2(y, x)
	bo := math.Atan2(z, math.Sqrt(x*x+y*y))
	//lo = lo * 180 / math.Pi
	bo = bo * 180 / math.Pi
	//lo+=GXCLo(lo,bo,JD);
	//bo+=GXCBo(lo,bo,JD)/3600;
	//lo+=HJZD(JD);
	return bo
}

func MercuryApparentLoBo(JD float64) (float64, float64) {
	x, y, z := AMercuryXYZ(JD)
	to := 0.0057755183 * math.Sqrt(x*x+y*y+z*z)
	x, y, z = AMercuryXYZ(JD - to)
	lo := math.Atan2(y, x)
	bo := math.Atan2(z, math.Sqrt(x*x+y*y))
	lo = lo * 180 / math.Pi
	bo = bo * 180 / math.Pi
	lo = Limit360(lo) + HJZD(JD)
	//lo-=GXCLo(lo,bo,JD)/3600;
	//bo+=GXCBo(lo,bo,JD);
	return lo, bo
}

func MercuryMag(JD float64) float64 {
	AwaySun := MercuryR(JD)
	AwayEarth := EarthMercuryAway(JD)
	Away := planet.WherePlanet(-1, 2, JD)
	i := (AwaySun*AwaySun + AwayEarth*AwayEarth - Away*Away) / (2 * AwaySun * AwayEarth)
	i = ArcCos(i)
	Mag := -0.42 + 5*math.Log10(AwaySun*AwayEarth) + 0.0380*i - 0.000273*i*i + 0.000002*i*i*i
	return FloatRound(Mag, 2)
}

func MercuryHeight(jde, lon, lat, timezone float64) float64 {
	// 转换为世界时
	utcJde := jde - timezone/24.0
	// 计算视恒星时
	ra, dec := MercuryApparentRaDec(TD2UT(utcJde, true))
	st := Limit360(ApparentSiderealTime(utcJde)*15 + lon)
	// 计算时角
	H := Limit360(st - ra)
	// 高度角、时角与天球座标三角转换公式
	// sin(h)=sin(lat)*sin(dec)+cos(dec)*cos(lat)*cos(H)
	sinHeight := Sin(lat)*Sin(dec) + Cos(dec)*Cos(lat)*Cos(H)
	return ArcSin(sinHeight)
}

func MercuryAzimuth(jde, lon, lat, timezone float64) float64 {
	// 转换为世界时
	utcJde := jde - timezone/24.0
	// 计算视恒星时
	ra, dec := MercuryApparentRaDec(TD2UT(utcJde, true))
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

func MercuryHourAngle(JD, Lon, TZ float64) float64 {
	startime := Limit360(ApparentSiderealTime(JD-TZ/24)*15 + Lon)
	timeangle := startime - MercuryApparentRa(TD2UT(JD-TZ/24.0, true))
	if timeangle < 0 {
		timeangle += 360
	}
	return timeangle
}

func MercuryCulminationTime(jde, lon, timezone float64) float64 {
	//jde 世界时，非力学时，当地时区 0时，无需转换力学时
	//ra,dec 瞬时天球座标，非J2000等时间天球坐标
	jde = math.Floor(jde) + 0.5
	JD1 := jde + Limit360(360-MercuryHourAngle(jde, lon, timezone))/15.0/24.0*0.99726851851851851851
	limitHA := func(jde, lon, timezone float64) float64 {
		ha := MercuryHourAngle(jde, lon, timezone)
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

func MercuryRiseTime(JD, Lon, Lat, TZ, ZS, HEI float64) float64 {
	return mercuryRiseDown(JD, Lon, Lat, TZ, ZS, HEI, true)
}

func MercuryDownTime(JD, Lon, Lat, TZ, ZS, HEI float64) float64 {
	return mercuryRiseDown(JD, Lon, Lat, TZ, ZS, HEI, false)
}

func mercuryRiseDown(JD, Lon, Lat, TZ, ZS, HEI float64, isRise bool) float64 {
	var An float64
	JD = math.Floor(JD) + 0.5
	ntz := math.Round(Lon / 15)
	if ZS != 0 {
		An = -0.8333
	}
	An = An - HeightDegreeByLat(HEI, Lat)
	tztime := MercuryCulminationTime(JD, Lon, ntz)
	if MercuryHeight(tztime, Lon, Lat, ntz) < An {
		return -2 //极夜
	}
	if MercuryHeight(tztime-0.5, Lon, Lat, ntz) > An {
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
		for MercuryHeight(rise, Lon, Lat, ntz) > An {
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
		stDegree := MercuryHeight(JD0, Lon, Lat, ntz) - An
		stDegreep := (MercuryHeight(JD0+0.000005, Lon, Lat, ntz) - MercuryHeight(JD0-0.000005, Lon, Lat, ntz)) / 0.00001
		JD1 = JD0 - stDegree/stDegreep
		if math.Abs(JD1-JD0) <= 0.00001 {
			break
		}
	}
	return JD1 - ntz/24 + TZ/24
}

// Pos

const MERCURY_S_PERIOD = 1 / ((1 / 87.9691) - (1 / 365.256363004))

func mercuryConjunction(jde float64, next uint8) float64 {
	//0=last 1=next
	decSub := func(jde float64) float64 {
		sub := Limit360(MercuryApparentLo(jde) - HSunApparentLo(jde))
		if sub > 180 {
			sub -= 360
		}
		if sub < -180 {
			sub += 360
		}
		return sub
	}
	nowSub := decSub(jde)
	// pos 大于0:远离太阳 小于0:靠近太阳
	pos := math.Abs(decSub(jde+1/86400.0)) - math.Abs(nowSub)
	if pos >= 0 && next == 1 && nowSub > 0 {
		jde += MERCURY_S_PERIOD/8.0 + 2
	}
	if pos >= 0 && next == 1 && nowSub < 0 {
		jde += MERCURY_S_PERIOD/6.0 + 2
	}
	if pos <= 0 && next == 0 && nowSub < 0 {
		jde -= MERCURY_S_PERIOD/8.0 + 2
	}
	if pos <= 0 && next == 0 && nowSub > 0 {
		jde -= MERCURY_S_PERIOD/6.0 + 2
	}
	for {
		nowSub := decSub(jde)
		pos := math.Abs(decSub(jde+1/86400.0)) - math.Abs(nowSub)
		if math.Abs(nowSub) > 12 || (pos > 0 && next == 1) || (pos < 0 && next == 0) {
			if next == 1 {
				jde += 2
			} else {
				jde -= 2
			}
			continue
		}
		break
	}
	JD1 := jde
	for {
		JD0 := JD1
		stDegree := decSub(JD0)
		stDegreep := (decSub(JD0+0.000005) - decSub(JD0-0.000005)) / 0.00001
		JD1 = JD0 - stDegree/stDegreep
		if math.Abs(JD1-JD0) <= 0.00001 {
			break
		}
	}
	return TD2UT(JD1, false)
}

func LastMercuryConjunction(jde float64) float64 {
	return mercuryConjunction(jde, 0)
}

func NextMercuryConjunction(jde float64) float64 {
	return mercuryConjunction(jde, 1)
}

func NextMercuryInferiorConjunction(jde float64) float64 {
	date := NextMercuryConjunction(jde)
	if EarthMercuryAway(date) > EarthAway(date) {
		return NextMercuryConjunction(date + 2)
	}
	return date
}

func NextMercurySuperiorConjunction(jde float64) float64 {
	date := NextMercuryConjunction(jde)
	if EarthMercuryAway(date) < EarthAway(date) {
		return NextMercuryConjunction(date + 2)
	}
	return date
}

func LastMercuryInferiorConjunction(jde float64) float64 {
	date := LastMercuryConjunction(jde)
	if EarthMercuryAway(date) > EarthAway(date) {
		return LastMercuryConjunction(date - 2)
	}
	return date
}

func LastMercurySuperiorConjunction(jde float64) float64 {
	date := LastMercuryConjunction(jde)
	if EarthMercuryAway(date) < EarthAway(date) {
		return LastMercuryConjunction(date - 2)
	}
	return date
}

func mercuryRetrograde(jde float64) float64 {
	//0=last 1=next
	decSunSub := func(jde float64) float64 {
		sub := Limit360(MercuryApparentRa(jde) - SunApparentRa(jde))
		if sub > 180 {
			sub -= 360
		}
		if sub < -180 {
			sub += 360
		}
		return sub
	}
	decSub := func(jde float64, val float64) float64 {
		sub := MercuryApparentRa(jde+val) - MercuryApparentRa(jde-val)
		if sub > 180 {
			sub -= 360
		}
		if sub < -180 {
			sub += 360
		}
		return sub / (2 * val)
	}
	lastHe := LastMercuryConjunction(jde)
	nextHe := NextMercuryConjunction(jde)
	nowSub := decSunSub(jde)
	if nowSub > 0 {
		jde = lastHe + ((nextHe - lastHe) / 5.0 * 3.5)
	} else {
		jde = lastHe + ((nextHe - lastHe) / 5.5)
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
	//fmt.Println((min - lastHe) / (nextHe - lastHe))
	return TD2UT(min, false)
}

func NextMercuryRetrograde(jde float64) float64 {
	date := mercuryRetrograde(jde)
	if date < jde {
		nextHe := NextMercuryConjunction(jde)
		return mercuryRetrograde(nextHe + 2)
	}
	return date
}

func LastMercuryRetrograde(jde float64) float64 {
	lastHe := LastMercuryConjunction(jde)
	date := mercuryRetrograde(lastHe + 2)
	if date > jde {
		lastLastHe := LastMercuryConjunction(lastHe - 2)
		return mercuryRetrograde(lastLastHe + 2)
	}
	return date
}

func NextMercuryProgradeToRetrograde(jde float64) float64 {
	date := NextMercuryRetrograde(jde)
	sub := Limit360(MercuryApparentRa(date) - SunApparentRa(date))
	if sub > 180 {
		return NextMercuryRetrograde(date + MERCURY_S_PERIOD/2)
	}
	return date
}

func NextMercuryRetrogradeToPrograde(jde float64) float64 {
	date := NextMercuryRetrograde(jde)
	sub := Limit360(MercuryApparentRa(date) - SunApparentRa(date))
	if sub < 180 {
		return NextMercuryRetrograde(date + 12)
	}
	return date
}

func LastMercuryProgradeToRetrograde(jde float64) float64 {
	date := LastMercuryRetrograde(jde)
	sub := Limit360(MercuryApparentRa(date) - SunApparentRa(date))
	if sub > 180 {
		return LastMercuryRetrograde(date - 12)
	}
	return date
}

func LastMercuryRetrogradeToPrograde(jde float64) float64 {
	date := LastMercuryRetrograde(jde)
	sub := Limit360(MercuryApparentRa(date) - SunApparentRa(date))
	if sub < 180 {
		return LastMercuryRetrograde(date - MERCURY_S_PERIOD/2)
	}
	return date
}

func MercurySunElongation(jde float64) float64 {
	lo1, bo1 := MercuryApparentLoBo(jde)
	lo2 := SunApparentLo(jde)
	bo2 := HSunTrueBo(jde)
	return StarAngularSeparation(lo1, bo1, lo2, bo2)
}
func mercuryGreatestElongation(jde float64) float64 {
	decSunSub := func(jde float64) float64 {
		sub := Limit360(MercuryApparentRa(jde) - SunApparentRa(jde))
		if sub > 180 {
			sub -= 360
		}
		if sub < -180 {
			sub += 360
		}
		return sub
	}
	decSub := func(jde float64, val float64) float64 {
		sub := MercurySunElongation(jde+val) - MercurySunElongation(jde-val)
		if sub > 180 {
			sub -= 360
		}
		if sub < -180 {
			sub += 360
		}
		return sub / (2 * val)
	}
	lastHe := LastMercuryConjunction(jde)
	nextHe := NextMercuryConjunction(jde)
	nowSub := decSunSub(jde)
	if nowSub > 0 {
		jde = lastHe + ((nextHe - lastHe) / 5.0 * 2.0)
	} else {
		jde = lastHe + ((nextHe - lastHe) / 6.0)
	}
	for {
		nowSub := decSub(jde, 1.0/86400.0)
		if math.Abs(nowSub) > 0.4 {
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
	//fmt.Println((min - lastHe) / (nextHe - lastHe))
	return TD2UT(min, false)
}

func NextMercuryGreatestElongation(jde float64) float64 {
	date := mercuryGreatestElongation(jde)
	if date < jde {
		nextHe := NextMercuryConjunction(jde)
		return mercuryGreatestElongation(nextHe + 2)
	}
	return date
}

func LastMercuryGreatestElongation(jde float64) float64 {
	lastHe := LastMercuryConjunction(jde)
	date := mercuryGreatestElongation(lastHe + 2)
	if date > jde {
		lastLastHe := LastMercuryConjunction(lastHe - 2)
		return mercuryGreatestElongation(lastLastHe + 2)
	}
	return date
}

func NextMercuryGreatestElongationEast(jde float64) float64 {
	date := NextMercuryGreatestElongation(jde)
	sub := Limit360(MercuryApparentRa(date) - SunApparentRa(date))
	if sub > 180 {
		return NextMercuryGreatestElongation(date + 1)
	}
	return date
}

func NextMercuryGreatestElongationWest(jde float64) float64 {
	date := NextMercuryGreatestElongation(jde)
	sub := Limit360(MercuryApparentRa(date) - SunApparentRa(date))
	if sub < 180 {
		return NextMercuryGreatestElongation(date + 1)
	}
	return date
}

func LastMercuryGreatestElongationEast(jde float64) float64 {
	date := LastMercuryGreatestElongation(jde)
	sub := Limit360(MercuryApparentRa(date) - SunApparentRa(date))
	if sub > 180 {
		return LastMercuryGreatestElongation(date - 1)
	}
	return date
}

func LastMercuryGreatestElongationWest(jde float64) float64 {
	date := LastMercuryGreatestElongation(jde)
	sub := Limit360(MercuryApparentRa(date) - SunApparentRa(date))
	if sub < 180 {
		return LastMercuryGreatestElongation(date - 1)
	}
	return date
}
