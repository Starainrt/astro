package basic

import (
	. "github.com/starainrt/astro/tools"
	"math"
)

// StarHeight 星体的高度角
// 传入 jde时间、瞬时赤经、瞬时赤纬、经度、纬度、时区，jde时间应为时区时间
// 返回高度角，单位为度
func StarHeight(jde, ra, dec, lon, lat, timezone float64) float64 {
	// 转换为世界时
	utcJde := jde - timezone/24.0
	// 计算视恒星时
	st := Limit360(ApparentSiderealTime(utcJde)*15 + lon)
	// 计算时角
	H := Limit360(st - ra)
	// 高度角、时角与天球座标三角转换公式
	// sin(h)=sin(lat)*sin(dec)+cos(dec)*cos(lat)*cos(H)
	sinHeight := Sin(lat)*Sin(dec) + Cos(dec)*Cos(lat)*Cos(H)
	return ArcSin(sinHeight)
}

// StarAzimuth 星体的方位角
// 传入 jde时间、瞬时赤经、瞬时赤纬、经度、纬度、时区，jde时间应为时区时间
// 返回方位角，单位为度，正北为0，度数顺时针增加，取值范围[0-360)
func StarAzimuth(jde, ra, dec, lon, lat, timezone float64) float64 {
	// 转换为世界时
	utcJde := jde - timezone/24.0
	// 计算视恒星时
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

// StarHourAngle 星体的时角
// 传入 jde时间、瞬时赤经、瞬时赤纬、经度、时区，jde时间应为时区时间
// 返回时角
func StarHourAngle(jde, ra, lon, timezone float64) float64 {
	// 转换为世界时
	utcJde := jde - timezone/24.0
	// 计算视恒星时
	st := Limit360(ApparentSiderealTime(utcJde)*15 + lon)
	// 计算时角
	return Limit360(st - ra)
}

// MeanSiderealTime 平恒星时
func MeanSiderealTime(JD float64) float64 {
	return MeanSiderealTime2006(JD)
}

// ApparentSiderealTime 视恒星时，计算章动
func ApparentSiderealTime(JD float64) float64 {
	return ApparentSiderealTime2006(JD)
}

// MeanSiderealTime1982 不含章动下的恒星时
func MeanSiderealTime1982(JD float64) float64 {
	T := (JD - 2451545) / 36525
	return (Limit360(280.46061837+360.98564736629*(JD-2451545.0)+0.000387933*T*T-T*T*T/38710000) / 15)
}

// ApparentSiderealTime1982 视恒星时，计算章动
func ApparentSiderealTime1982(JD float64) float64 {
	tmp := MeanSiderealTime1982(JD)
	return tmp + Nutation2000Bi(JD)*Cos(Sita(JD))/15
}

// EarthRotationAngle 计算地球自转角 (ERA)
// jd_ut1: UT1 时间的儒略日
// 返回值: 地球自转角 (弧度)
func EarthRotationAngle(jd_ut1 float64) float64 {
	t := jd_ut1 - 2451545.0
	frac := math.Mod(jd_ut1, 1.0)

	era := math.Mod(math.Pi*2*(0.7790572732640+0.00273781191135448*t+frac), math.Pi*2)
	if era < 0 {
		era += math.Pi * 2
	}

	return era
}

// MeanSiderealTime2006 计算格林尼治平恒星时 (GMST)
// jd_ut1: UT1 时间的儒略日
// jd_tt: TT 时间的儒略日
// 返回值: 格林尼治平恒星时 (弧度)
func MeanSiderealTime2006(jd_ut1 float64) float64 {
	jd_tt := TD2UT(jd_ut1, true)
	t := (jd_tt - 2451545.0) / 36525.0
	era := EarthRotationAngle(jd_ut1)

	// 公式 2.12
	gmst := math.Mod(era+(0.014506+4612.15739966*t+1.39667721*t*t+
		-0.00009344*t*t*t+0.00001882*t*t*t*t)/60/60*math.Pi/180, math.Pi*2)

	if gmst < 0 {
		gmst += math.Pi * 2
	}

	return gmst * deg / 15
}

// ApparentSiderealTime2006 视恒星时，计算章动
func ApparentSiderealTime2006(JD float64) float64 {
	tmp := MeanSiderealTime2006(JD)
	return tmp + Nutation2000Bi(JD)*Cos(Sita(JD))/15
}

func StarAngle(RA, DEC, JD, Lon, Lat, TZ float64) float64 {
	//JD=JD-8/24+TZ/24;
	calcjd := JD - TZ/24
	st := Limit360(ApparentSiderealTime(calcjd)*15 + Lon)
	H := Limit360(st - RA)
	tmp2 := Sin(H) / (Cos(H)*Sin(Lat) - Tan(DEC)*Cos(Lat))
	Angle := ArcTan(tmp2)
	if Angle < 0 {
		if H/15 < 12 {
			return Angle + 360
		} else {
			return Angle + 180
		}
	} else {
		if H/15 < 12 {
			return Angle + 180
		} else {
			return Angle
		}
	}
}

func StarRiseTime(jde, ra, dec, lon, lat, height, timezone float64, aero bool) float64 {
	return StarRiseDownTime(jde, ra, dec, lon, lat, height, timezone, aero, true)
}

func StarDownTime(jde, ra, dec, lon, lat, height, timezone float64, aero bool) float64 {
	return StarRiseDownTime(jde, ra, dec, lon, lat, height, timezone, aero, false)
}

func StarRiseDownTime(jde, ra, dec, lon, lat, height, timezone float64, aero, isRise bool) float64 {
	//jde 世界时，非力学时，当地时区 0时，无需转换力学时
	//ra,dec 瞬时天球座标，非J2000等时间天球坐标
	jde = math.Floor(jde) + 0.5
	var An float64 = 0
	if aero {
		An = -0.566667
	}
	An = An - HeightDegreeByLat(height, lat)
	sct := StarCulminationTime(jde, ra, lon, timezone)
	tmp := (Sin(An) - Sin(dec)*Sin(lat)) / (Cos(dec) * Cos(lat))
	if math.Abs(tmp) > 1 {
		if StarHeight(sct, ra, dec, lon, lat, timezone) < 0 {
			return -2 //极夜
		} else {
			return -1 //极昼
		}
	}
	var JD1 float64
	if isRise {
		JD1 = sct - ArcCos(tmp)/15.0/24.0
	} else {
		JD1 = sct + ArcCos(tmp)/15.0/24.0
	}
	for {
		JD0 := JD1
		stDegree := StarHeight(JD0, ra, dec, lon, lat, timezone) - An
		stDegreep := (StarHeight(JD0+0.000005, ra, dec, lon, lat, timezone) - StarHeight(JD0-0.000005, ra, dec, lon, lat, timezone)) / 0.00001
		JD1 = JD0 - stDegree/stDegreep
		if math.Abs(JD1-JD0) <= 0.00001 {
			break
		}
	}
	return JD1
}

func StarCulminationTime(jde, ra, lon, timezone float64) float64 {
	//jde 世界时，非力学时，当地时区 0时，无需转换力学时
	//ra,dec 瞬时天球座标，非J2000等时间天球坐标
	jde = math.Floor(jde) + 0.5
	JD1 := jde + Limit360(360-StarHourAngle(jde, ra, lon, timezone))/15.0/24.0*0.99726851851851851851
	limitStarHA := func(jde, ra, lon, timezone float64) float64 {
		ha := StarHourAngle(jde, ra, lon, timezone)
		if ha < 180 {
			ha += 360
		}
		return ha
	}
	for {
		JD0 := JD1
		stDegree := limitStarHA(JD0, ra, lon, timezone) - 360
		stDegreep := (limitStarHA(JD0+0.000005, ra, lon, timezone) - limitStarHA(JD0-0.000005, ra, lon, timezone)) / 0.00001
		JD1 = JD0 - stDegree/stDegreep
		if math.Abs(JD1-JD0) <= 0.00001 {
			break
		}
	}
	return JD1
}

func StarAngularSeparation(ra1, dec1, ra2, dec2 float64) float64 {
	//cos(d)=sinδ1 sinδ2 + cosδ1 cosδ2 cos(α1-α2)
	d := Sin(dec1)*Sin(dec2) + Cos(dec1)*Cos(dec2)*Cos(ra1-ra2)
	if math.Abs(d) >= 0.999999997 {
		//d = √(Δα*cosδ)2+(Δδ)2
		tmp1 := ((ra1 - ra2) * Cos((dec1+dec2)/2))
		tmp2 := (dec1 - dec2)
		return math.Sqrt(tmp1*tmp1 + tmp2*tmp2)
	}
	return ArcCos(d)
}
