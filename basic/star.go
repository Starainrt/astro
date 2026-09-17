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
	hourAngle := Limit360(st - ra)
	// 高度角、时角与天球座标三角转换公式
	// sin(h)=sin(lat)*sin(dec)+cos(dec)*cos(lat)*cos(hourAngle)
	sinHeight := Sin(lat)*Sin(dec) + Cos(dec)*Cos(lat)*Cos(hourAngle)
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
func MeanSiderealTime(jd float64) float64 {
	return MeanSiderealTime2006(jd)
}

// ApparentSiderealTime 视恒星时，计算章动
func ApparentSiderealTime(jd float64) float64 {
	return ApparentSiderealTime2006(jd)
}

// MeanSiderealTime1982 不含章动下的恒星时
func MeanSiderealTime1982(jd float64) float64 {
	t := (jd - 2451545) / 36525
	return (Limit360(280.46061837+360.98564736629*(jd-2451545.0)+0.000387933*t*t-t*t*t/38710000) / 15)
}

// ApparentSiderealTime1982 视恒星时，计算章动
func ApparentSiderealTime1982(jd float64) float64 {
	tmp := MeanSiderealTime1982(jd)
	dpsi, deps := Nutation2000B(jd)
	return tmp + dpsi*math.Cos((Obliquity1980(jd)+deps)*math.Pi/180)/15
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

// ApparentSiderealTime2006 视恒星时，计算章动。
// 一次求值只要一份 IAU2000B 章动（黄经与交角在同一次展开里同时得到），并走有界记忆表，
// 避免月掩路径里的重复求值；ΔT 覆盖会使记忆表按世代失效。见 sidereal_memo.go。
// ApparentSiderealTime2006 computes apparent sidereal time. One evaluation needs only a single
// IAU2000B nutation expansion (longitude and obliquity come out of the same series) and is
// memoized so repeated occultation-path queries do not rerun it; a ΔT override invalidates the
// memo by generation. See sidereal_memo.go.
func ApparentSiderealTime2006(jd float64) float64 {
	if value, ok := siderealMemoLoad(jd); ok {
		return value
	}
	generation := siderealMemoGeneration()
	tmp := MeanSiderealTime2006(jd)
	dpsi, deps := Nutation2000B(jd)
	value := tmp + dpsi*math.Cos((Obliquity1980(jd)+deps)*math.Pi/180)/15
	siderealMemoStore(jd, value, generation)
	return value
}

func StarRiseTime(jde, ra, dec, lon, lat, height, timezone float64, aero bool) (float64, error) {
	return StarRiseSetTime(jde, ra, dec, lon, lat, height, timezone, aero, true)
}

func StarSetTime(jde, ra, dec, lon, lat, height, timezone float64, aero bool) (float64, error) {
	return StarRiseSetTime(jde, ra, dec, lon, lat, height, timezone, aero, false)
}

func StarRiseSetTime(jde, ra, dec, lon, lat, height, timezone float64, aero, isRise bool) (float64, error) {
	if !isFiniteFloat(jde) || !isFiniteFloat(ra) || !isFiniteFloat(dec) || !isFiniteFloat(lon) || !isFiniteFloat(lat) || !isFiniteFloat(height) || !isFiniteFloat(timezone) {
		return 0, ErrInvalidObservationInput
	}
	//jde 世界时，非力学时，当地时区 0时，无需转换力学时
	//ra,dec 瞬时天球座标，非J2000等时间天球坐标
	jde = math.Floor(jde) + 0.5
	targetAltitude := StandardAltitudeStar(aero, height, lat)
	sct := StarCulminationTime(jde, ra, lon, timezone)
	tmp := (Sin(targetAltitude) - Sin(dec)*Sin(lat)) / (Cos(dec) * Cos(lat))
	if math.Abs(tmp) > 1 {
		if StarHeight(sct, ra, dec, lon, lat, timezone) < 0 {
			return 0, ErrNeverRise
		}
		return 0, ErrNeverSet
	}
	var estimateJD float64
	if isRise {
		estimateJD = sct - ArcCos(tmp)/15.0/24.0
	} else {
		estimateJD = sct + ArcCos(tmp)/15.0/24.0
	}
	var ok bool
	estimateJD, ok = eventNewtonRefine(estimateJD, 0.00001, func(prevJD float64) float64 {
		stDegree := StarHeight(prevJD, ra, dec, lon, lat, timezone) - targetAltitude
		stDegreep := (StarHeight(prevJD+0.000005, ra, dec, lon, lat, timezone) - StarHeight(prevJD-0.000005, ra, dec, lon, lat, timezone)) / 0.00001
		return stDegree / stDegreep
	})
	if !ok {
		return 0, ErrInvalidObservationInput
	}
	return estimateJD, nil
}

func StarCulminationTime(jde, ra, lon, timezone float64) float64 {
	if !isFiniteFloat(jde) || !isFiniteFloat(ra) || !isFiniteFloat(lon) || !isFiniteFloat(timezone) {
		return math.NaN()
	}
	//jde 世界时，非力学时，当地时区 0时，无需转换力学时
	//ra,dec 瞬时天球座标，非J2000等时间天球坐标
	jde = math.Floor(jde) + 0.5
	estimateJD := jde + Limit360(360-StarHourAngle(jde, ra, lon, timezone))/15.0/24.0*0.99726851851851851851
	limitStarHA := func(jde, ra, lon, timezone float64) float64 {
		ha := StarHourAngle(jde, ra, lon, timezone)
		if ha < 180 {
			ha += 360
		}
		return ha
	}
	var ok bool
	estimateJD, ok = eventNewtonRefine(estimateJD, 0.00001, func(prevJD float64) float64 {
		stDegree := limitStarHA(prevJD, ra, lon, timezone) - 360
		stDegreep := (limitStarHA(prevJD+0.000005, ra, lon, timezone) - limitStarHA(prevJD-0.000005, ra, lon, timezone)) / 0.00001
		return stDegree / stDegreep
	})
	if !ok {
		return math.NaN()
	}
	return estimateJD
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
