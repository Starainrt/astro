package basic

import (
	. "github.com/starainrt/astro/tools"
)

// StarHeight 星体的高度角
// 传入 jde时间、瞬时赤经、瞬时赤纬、经度、纬度、时区，jde时间应为时区时间
// 返回高度角，单位为度
func StarHeight(jde, ra, dec, lon, lat, timezone float64) float64 {
	// 转换为世界时
	utcJde := jde - timezone/24.0
	// 计算视恒星时
	st := Limit360(SeeStarTime(utcJde)*15 + lon)
	// 计算时角
	H := Limit360(st - ra)
	// 高度角、时角与天球座标三角转换公式
	// sin(h)=sin(lat)*sin(dec)+cos(dec)*cos(lat)*cos(H)
	sinHeight := Sin(lat)*Sin(dec) + Cos(dec)*Cos(lat)*Cos(H)
	return ArcSin(sinHeight)
}

//StarAzimuth 星体的方位角
// 传入 jde时间、瞬时赤经、瞬时赤纬、经度、纬度、时区，jde时间应为时区时间
// 返回方位角，单位为度，正北为0，度数顺时针增加，取值范围[0-360)
func StarAzimuth(jde, ra, dec, lon, lat, timezone float64) float64 {
	// 转换为世界时
	utcJde := jde - timezone/24.0
	// 计算视恒星时
	st := Limit360(SeeStarTime(utcJde)*15 + lon)
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

//StarHourAngle 星体的时角
// 传入 jde时间、瞬时赤经、瞬时赤纬、经度、时区，jde时间应为时区时间
// 返回时角
func StarHourAngle(jde, ra, lon, timezone float64) float64 {
	// 转换为世界时
	utcJde := jde - timezone/24.0
	// 计算视恒星时
	st := Limit360(SeeStarTime(utcJde)*15 + lon)
	// 计算时角
	return Limit360(st - ra)
}

// TrueStarTime 不含章动下的恒星时
func TrueStarTime(JD float64) float64 {
	T := (JD - 2451545) / 36525
	return (Limit360(280.46061837+360.98564736629*(JD-2451545.0)+0.000387933*T*T-T*T*T/38710000) / 15)
}

// SeeStarTime 视恒星时，计算章动
func SeeStarTime(JD float64) float64 {
	tmp := TrueStarTime(JD)
	return tmp + HJZD(JD)*Cos(Sita(JD))/15
}
func StarAngle(RA, DEC, JD, Lon, Lat, TZ float64) float64 {
	//JD=JD-8/24+TZ/24;
	calcjd := JD - TZ/24
	st := Limit360(SeeStarTime(calcjd)*15 + Lon)
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

func StarRiseTime(jde, ra, dec, lon, lat, height, timezone float64, aero bool) (float64, error) {

	return 0, nil
}
