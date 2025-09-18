package basic

import (
	"math"

	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

func VenusL(JD float64) float64 {
	return planet.WherePlanet(2, 0, JD)
}

func VenusB(JD float64) float64 {
	return planet.WherePlanet(2, 1, JD)
}
func VenusR(JD float64) float64 {
	return planet.WherePlanet(2, 2, JD)
}
func AVenusX(JD float64) float64 {
	l := VenusL(JD)
	b := VenusB(JD)
	r := VenusR(JD)
	el := planet.WherePlanet(-1, 0, JD)
	eb := planet.WherePlanet(-1, 1, JD)
	er := planet.WherePlanet(-1, 2, JD)
	x := r*Cos(b)*Cos(l) - er*Cos(eb)*Cos(el)
	return x
}

func AVenusY(JD float64) float64 {

	l := VenusL(JD)
	b := VenusB(JD)
	r := VenusR(JD)
	el := planet.WherePlanet(-1, 0, JD)
	eb := planet.WherePlanet(-1, 1, JD)
	er := planet.WherePlanet(-1, 2, JD)
	y := r*Cos(b)*Sin(l) - er*Cos(eb)*Sin(el)
	return y
}
func AVenusZ(JD float64) float64 {
	//l := VenusL(JD)
	b := VenusB(JD)
	r := VenusR(JD)
	//	el := planet.WherePlanet(-1, 0, JD)
	eb := planet.WherePlanet(-1, 1, JD)
	er := planet.WherePlanet(-1, 2, JD)
	z := r*Sin(b) - er*Sin(eb)
	return z
}

func AVenusXYZ(JD float64) (float64, float64, float64) {
	l := VenusL(JD)
	b := VenusB(JD)
	r := VenusR(JD)
	el := planet.WherePlanet(-1, 0, JD)
	eb := planet.WherePlanet(-1, 1, JD)
	er := planet.WherePlanet(-1, 2, JD)
	x := r*Cos(b)*Cos(l) - er*Cos(eb)*Cos(el)
	y := r*Cos(b)*Sin(l) - er*Cos(eb)*Sin(el)
	z := r*Sin(b) - er*Sin(eb)
	return x, y, z
}

func VenusApparentRa(JD float64) float64 {
	lo, bo := VenusApparentLoBo(JD)
	sita := Sita(JD)
	ra := math.Atan2((Sin(lo)*Cos(sita) - Tan(bo)*Sin(sita)), Cos(lo))
	ra = ra * 180 / math.Pi
	return Limit360(ra)
}
func VenusApparentDec(JD float64) float64 {
	lo, bo := VenusApparentLoBo(JD)
	sita := Sita(JD)
	dec := ArcSin(Sin(bo)*Cos(sita) + Cos(bo)*Sin(sita)*Sin(lo))
	return dec
}

func VenusApparentRaDec(JD float64) (float64, float64) {
	lo, bo := VenusApparentLoBo(JD)
	sita := Sita(JD)
	ra := math.Atan2((Sin(lo)*Cos(sita) - Tan(bo)*Sin(sita)), Cos(lo))
	ra = ra * 180 / math.Pi
	dec := ArcSin(Sin(bo)*Cos(sita) + Cos(bo)*Sin(sita)*Sin(lo))
	return Limit360(ra), dec
}

func EarthVenusAway(JD float64) float64 {
	x, y, z := AVenusXYZ(JD)
	to := math.Sqrt(x*x + y*y + z*z)
	return to
}

func VenusApparentLo(JD float64) float64 {
	x, y, z := AVenusXYZ(JD)
	to := 0.0057755183 * math.Sqrt(x*x+y*y+z*z)
	x, y, z = AVenusXYZ(JD - to)
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

func VenusApparentBo(JD float64) float64 {
	x, y, z := AVenusXYZ(JD)
	to := 0.0057755183 * math.Sqrt(x*x+y*y+z*z)
	x, y, z = AVenusXYZ(JD - to)
	//lo := math.Atan2(y, x)
	bo := math.Atan2(z, math.Sqrt(x*x+y*y))
	//lo = lo * 180 / math.Pi
	bo = bo * 180 / math.Pi
	//lo+=GXCLo(lo,bo,JD);
	//bo+=GXCBo(lo,bo,JD)/3600;
	//lo+=Nutation2000Bi(JD);
	return bo
}

func VenusApparentLoBo(JD float64) (float64, float64) {
	x, y, z := AVenusXYZ(JD)
	to := 0.0057755183 * math.Sqrt(x*x+y*y+z*z)
	x, y, z = AVenusXYZ(JD - to)
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

func VenusMag(JD float64) float64 {
	AwaySun := VenusR(JD)
	AwayEarth := EarthVenusAway(JD)
	Away := planet.WherePlanet(-1, 2, JD)
	i := (AwaySun*AwaySun + AwayEarth*AwayEarth - Away*Away) / (2 * AwaySun * AwayEarth)
	i = ArcCos(i)
	Mag := -4.40 + 5*math.Log10(AwaySun*AwayEarth) + 0.0009*i + 0.000239*i*i - 0.00000065*i*i*i
	return FloatRound(Mag, 2)
}

func VenusHeight(jde, lon, lat, timezone float64) float64 {
	// 转换为世界时
	utcJde := jde - timezone/24.0
	// 计算视恒星时
	ra, dec := VenusApparentRaDec(TD2UT(utcJde, true))
	st := Limit360(ApparentSiderealTime(utcJde)*15 + lon)
	// 计算时角
	H := Limit360(st - ra)
	// 高度角、时角与天球座标三角转换公式
	// sin(h)=sin(lat)*sin(dec)+cos(dec)*cos(lat)*cos(H)
	sinHeight := Sin(lat)*Sin(dec) + Cos(dec)*Cos(lat)*Cos(H)
	return ArcSin(sinHeight)
}

func VenusAzimuth(jde, lon, lat, timezone float64) float64 {
	// 转换为世界时
	utcJde := jde - timezone/24.0
	// 计算视恒星时
	ra, dec := VenusApparentRaDec(TD2UT(utcJde, true))
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

func VenusHourAngle(JD, Lon, TZ float64) float64 {
	startime := Limit360(ApparentSiderealTime(JD-TZ/24)*15 + Lon)
	timeangle := startime - VenusApparentRa(TD2UT(JD-TZ/24.0, true))
	if timeangle < 0 {
		timeangle += 360
	}
	return timeangle
}

func VenusCulminationTime(jde, lon, timezone float64) float64 {
	//jde 世界时，非力学时，当地时区 0时，无需转换力学时
	//ra,dec 瞬时天球座标，非J2000等时间天球坐标
	jde = math.Floor(jde) + 0.5
	JD1 := jde + Limit360(360-VenusHourAngle(jde, lon, timezone))/15.0/24.0*0.99726851851851851851
	limitHA := func(jde, lon, timezone float64) float64 {
		ha := VenusHourAngle(jde, lon, timezone)
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

func VenusRiseTime(JD, Lon, Lat, TZ, ZS, HEI float64) float64 {
	return venusRiseDown(JD, Lon, Lat, TZ, ZS, HEI, true)
}

func VenusDownTime(JD, Lon, Lat, TZ, ZS, HEI float64) float64 {
	return venusRiseDown(JD, Lon, Lat, TZ, ZS, HEI, false)
}

func venusRiseDown(JD, Lon, Lat, TZ, ZS, HEI float64, isRise bool) float64 {
	var An float64
	JD = math.Floor(JD) + 0.5
	ntz := math.Round(Lon / 15)
	if ZS != 0 {
		An = -0.8333
	}
	An = An - HeightDegreeByLat(HEI, Lat)
	tztime := VenusCulminationTime(JD, Lon, ntz)
	if VenusHeight(tztime, Lon, Lat, ntz) < An {
		return -2 //极夜
	}
	if VenusHeight(tztime-0.5, Lon, Lat, ntz) > An {
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
		for VenusHeight(rise, Lon, Lat, ntz) > An {
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
		stDegree := VenusHeight(JD0, Lon, Lat, ntz) - An
		stDegreep := (VenusHeight(JD0+0.000005, Lon, Lat, ntz) - VenusHeight(JD0-0.000005, Lon, Lat, ntz)) / 0.00001
		JD1 = JD0 - stDegree/stDegreep
		if math.Abs(JD1-JD0) <= 0.00001 {
			break
		}
	}
	return JD1 - ntz/24 + TZ/24
}

// Pos

const VENUS_S_PERIOD = 1 / ((1 / 224.701) - (1 / 365.256363004))

func venusConjunction(jde float64, next uint8) float64 {
	//0=last 1=next
	decSub := func(jde float64) float64 {
		sub := Limit360(VenusApparentLo(jde) - HSunApparentLo(jde))
		if sub > 180 {
			sub -= 360
		}
		if sub < -180 {
			sub += 360
		}
		return sub
	}
	nowSub := decSub(jde)
	pos := math.Abs(decSub(jde+1/86400.0)) - math.Abs(nowSub)
	if pos >= 0 && next == 1 && nowSub > 0 {
		jde += VENUS_S_PERIOD/8.0 + 2
	}
	if pos >= 0 && next == 1 && nowSub < 0 {
		jde += VENUS_S_PERIOD/6.0 + 2
	}
	if pos <= 0 && next == 0 && nowSub < 0 {
		jde -= VENUS_S_PERIOD/8.0 + 2
	}
	if pos <= 0 && next == 0 && nowSub > 0 {
		jde -= VENUS_S_PERIOD/6.0 + 2
	}
	for {
		nowSub := decSub(jde)
		pos := math.Abs(decSub(jde+1/86400.0)) - math.Abs(nowSub)
		if math.Abs(nowSub) > 24 || (pos > 0 && next == 1) || (pos < 0 && next == 0) {
			if next == 1 {
				jde += 8
			} else {
				jde -= 8
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

func LastVenusConjunction(jde float64) float64 {
	return venusConjunction(jde, 0)
}

func NextVenusConjunction(jde float64) float64 {
	return venusConjunction(jde, 1)
}

func NextVenusInferiorConjunction(jde float64) float64 {
	date := NextVenusConjunction(jde)
	if EarthVenusAway(date) > EarthAway(date) {
		return NextVenusConjunction(date + 2)
	}
	return date
}

func NextVenusSuperiorConjunction(jde float64) float64 {
	date := NextVenusConjunction(jde)
	if EarthVenusAway(date) < EarthAway(date) {
		return NextVenusConjunction(date + 2)
	}
	return date
}

func LastVenusInferiorConjunction(jde float64) float64 {
	date := LastVenusConjunction(jde)
	if EarthVenusAway(date) > EarthAway(date) {
		return LastVenusConjunction(date - 2)
	}
	return date
}

func LastVenusSuperiorConjunction(jde float64) float64 {
	date := LastVenusConjunction(jde)
	if EarthVenusAway(date) < EarthAway(date) {
		return LastVenusConjunction(date - 2)
	}
	return date
}

func venusRetrograde(jde float64) float64 {
	//0=last 1=next
	decSunSub := func(jde float64) float64 {
		sub := Limit360(VenusApparentRa(jde) - SunApparentRa(jde))
		if sub > 180 {
			sub -= 360
		}
		if sub < -180 {
			sub += 360
		}
		return sub
	}
	decSub := func(jde float64, val float64) float64 {
		sub := VenusApparentRa(jde+val) - VenusApparentRa(jde-val)
		if sub > 180 {
			sub -= 360
		}
		if sub < -180 {
			sub += 360
		}
		return sub / (2 * val)
	}
	lastHe := LastVenusConjunction(jde)
	nextHe := NextVenusConjunction(jde)
	nowSub := decSunSub(jde)
	if nowSub > 0 {
		jde = lastHe + ((nextHe - lastHe) / 5.0 * 3.5)
	} else {
		jde = lastHe + 10
	}
	for {
		nowSub := decSub(jde, 1.0/86400.0)
		if math.Abs(nowSub) > 0.5 {
			jde += 5
			continue
		}
		break
	}
	JD1 := jde
	for {
		JD0 := JD1
		stDegree := decSub(JD0, 0.5/86400.0)
		stDegreep := (decSub(JD0+10.0/86400.0, 0.5/86400.0) - decSub(JD0-10.0/86400.0, 0.5/86400.0)) / (20.0 / 86400.0)
		JD1 = JD0 - stDegree/stDegreep
		if math.Abs(JD1-JD0) <= 20.0/86400.0 {
			break
		}
	}
	JD1 = JD1 - 10.0/86400.0
	min := JD1
	minRa := 100.0
	for i := 0.0; i < 40.0; i++ {
		tmp := decSub(JD1+i*0.5/86400.0, 0.5/86400.0)
		if math.Abs(tmp) < math.Abs(minRa) {
			minRa = tmp
			min = JD1 + i*0.5/86400.0
		}
	}
	//fmt.Println((min - lastHe) / (nextHe - lastHe))
	return TD2UT(min, false)
}

func NextVenusRetrograde(jde float64) float64 {
	date := venusRetrograde(jde)
	if date < jde {
		nextHe := NextVenusConjunction(jde)
		return venusRetrograde(nextHe + 2)
	}
	return date
}

func LastVenusRetrograde(jde float64) float64 {
	lastHe := LastVenusConjunction(jde)
	date := venusRetrograde(lastHe + 2)
	if date > jde {
		lastLastHe := LastVenusConjunction(lastHe - 2)
		return venusRetrograde(lastLastHe + 2)
	}
	return date
}

func NextVenusProgradeToRetrograde(jde float64) float64 {
	date := NextVenusRetrograde(jde)
	sub := Limit360(VenusApparentRa(date) - SunApparentRa(date))
	if sub > 180 {
		return NextVenusRetrograde(date + VENUS_S_PERIOD/2)
	}
	return date
}

func NextVenusRetrogradeToPrograde(jde float64) float64 {
	date := NextVenusRetrograde(jde)
	sub := Limit360(VenusApparentRa(date) - SunApparentRa(date))
	if sub < 180 {
		return NextVenusRetrograde(date + 12)
	}
	return date
}

func LastVenusProgradeToRetrograde(jde float64) float64 {
	date := LastVenusRetrograde(jde)
	sub := Limit360(VenusApparentRa(date) - SunApparentRa(date))
	if sub > 180 {
		return LastVenusRetrograde(date - 12)
	}
	return date
}

func LastVenusRetrogradeToPrograde(jde float64) float64 {
	date := LastVenusRetrograde(jde)
	sub := Limit360(VenusApparentRa(date) - SunApparentRa(date))
	if sub < 180 {
		return LastVenusRetrograde(date - VENUS_S_PERIOD/2)
	}
	return date
}

func VenusSunElongation(jde float64) float64 {
	lo1, bo1 := VenusApparentLoBo(jde)
	lo2 := SunApparentLo(jde)
	bo2 := HSunTrueBo(jde)
	return StarAngularSeparation(lo1, bo1, lo2, bo2)
}
func venusGreatestElongation(jde float64) float64 {
	decSunSub := func(jde float64) float64 {
		sub := Limit360(VenusApparentRa(jde) - SunApparentRa(jde))
		if sub > 180 {
			sub -= 360
		}
		if sub < -180 {
			sub += 360
		}
		return sub
	}
	decSub := func(jde float64, val float64) float64 {
		sub := VenusSunElongation(jde+val) - VenusSunElongation(jde-val)
		if sub > 180 {
			sub -= 360
		}
		if sub < -180 {
			sub += 360
		}
		return sub / (2 * val)
	}
	lastHe := LastVenusConjunction(jde)
	nextHe := NextVenusConjunction(jde)
	nowSub := decSunSub(jde)
	if nowSub > 0 {
		jde = lastHe + ((nextHe - lastHe) / 5.0 * 2.5)
	} else {
		jde = lastHe + ((nextHe - lastHe) / 5.0)
	}
	for {
		nowSub := decSub(jde, 1.0/86400.0)
		if math.Abs(nowSub) > 0.15 {
			jde += 5
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

func NextVenusGreatestElongation(jde float64) float64 {
	date := venusGreatestElongation(jde)
	if date < jde {
		nextHe := NextVenusConjunction(jde)
		return venusGreatestElongation(nextHe + 2)
	}
	return date
}

func LastVenusGreatestElongation(jde float64) float64 {
	lastHe := LastVenusConjunction(jde)
	date := venusGreatestElongation(lastHe + 2)
	if date > jde {
		lastLastHe := LastVenusConjunction(lastHe - 2)
		return venusGreatestElongation(lastLastHe + 2)
	}
	return date
}

func NextVenusGreatestElongationEast(jde float64) float64 {
	date := NextVenusGreatestElongation(jde)
	sub := Limit360(VenusApparentRa(date) - SunApparentRa(date))
	if sub > 180 {
		return NextVenusGreatestElongation(date + 1)
	}
	return date
}

func NextVenusGreatestElongationWest(jde float64) float64 {
	date := NextVenusGreatestElongation(jde)
	sub := Limit360(VenusApparentRa(date) - SunApparentRa(date))
	if sub < 180 {
		return NextVenusGreatestElongation(date + 1)
	}
	return date
}

func LastVenusGreatestElongationEast(jde float64) float64 {
	date := LastVenusGreatestElongation(jde)
	sub := Limit360(VenusApparentRa(date) - SunApparentRa(date))
	if sub > 180 {
		return LastVenusGreatestElongation(date - 1)
	}
	return date
}

func LastVenusGreatestElongationWest(jde float64) float64 {
	date := LastVenusGreatestElongation(jde)
	sub := Limit360(VenusApparentRa(date) - SunApparentRa(date))
	if sub < 180 {
		return LastVenusGreatestElongation(date - 1)
	}
	return date
}
