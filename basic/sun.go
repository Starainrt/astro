package basic

import (
	"math"

	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

/*
黄赤交角、nutation==true时，计算交角章动
*/
func EclipticObliquity(jde float64, nutation bool) float64 {
	U := (jde - 2451545) / 3652500.000
	sita := 23.000 + 26.000/60.000 + 21.448/3600.000 - ((4680.93*U - 1.55*U*U + 1999.25*U*U*U - 51.38*U*U*U*U - 249.67*U*U*U*U*U - 39.05*U*U*U*U*U*U + 7.12*U*U*U*U*U*U*U + 27.87*U*U*U*U*U*U*U*U + 5.79*U*U*U*U*U*U*U*U*U + 2.45*U*U*U*U*U*U*U*U*U*U) / 3600)
	if nutation {
		return sita + JJZD(jde)
	} else {
		return sita
	}
}

func Sita(JD float64) float64 {
	return EclipticObliquity(JD, true)
}

/*
 * @name 黄经章动
 */
func HJZD(jd float64) float64 { // '黄经章动
	t := (jd - 2451545) / 36525.000
	d := 297.8502042 + 445267.1115168*t - 0.0016300*t*t + t*t*t/545868 - t*t*t*t/113065000
	m := SunM(jd)
	n := MoonM(jd)
	f := MoonLonX(jd)
	o := 125.04452 - 1934.136261*t + 0.0020708*t*t + t*t*t/450000
	tp := [][]float64{{0, 0, 0, 0, 1, -171996, -174.2 * t}, {-2, 0, 0, 2, 2, -13187, -1.6 * t}, {0, 0, 0, 2, 2, -2274, -0.2 * t}, {0, 0, 0, 0, 2, 2062, 0.2 * t}, {0, 1, 0, 0, 0, 1426, -3.4 * t}, {0, 0, 1, 0, 0, 712, 0.1 * t}, {-2, 1, 0, 2, 2, -517, 1.2 * t}, {0, 0, 0, 2, 1, -386, -0.4 * t}, {0, 0, 1, 2, 2, -301, 0}, {-2, -1, 0, 2, 2, 217, -0.5 * t}, {-2, 0, 1, 0, 0, -158, 0}, {-2, 0, 0, 2, 1, 129, 0.1 * t}, {0, 0, -1, 2, 2, 123, 0}, {2, 0, 0, 0, 0, 63, 0}, {0, 0, 1, 0, 1, 63, 0.1 * t}, {2, 0, -1, 2, 2, -59, 0}, {0, 0, -1, 0, 1, -58, -0.1 * t}, {0, 0, 1, 2, 1, -51, 0}, {-2, 0, 2, 0, 0, 48, 0}, {0, 0, -2, 2, 1, 46, 0}, {2, 0, 0, 2, 2, -38, 0}, {0, 0, 2, 2, 2, -31, 0}, {0, 0, 2, 0, 0, 29, 0}, {-2, 0, 1, 2, 2, 29, 0}, {0, 0, 0, 2, 0, 26, 0}, {-2, 0, 0, 2, 0, -22, 0}, {0, 0, -1, 2, 1, 21, 0}, {0, 2, 0, 0, 0, 17, -0.1 * t}, {2, 0, -1, 0, 1, 16, 0}, {-2, 2, 0, 2, 2, -16, 0.1 * t}, {0, 1, 0, 0, 1, -15, 0}, {-2, 0, 1, 0, 1, -13, 0}, {0, -1, 0, 0, 1, -12, 0}, {0, 0, 2, -2, 0, 11, 0}, {2, 0, -1, 2, 1, -10, 0}, {2, 0, 1, 2, 2, -8, 0}, {0, 1, 0, 2, 2, 7, 0}, {-2, 1, 1, 0, 0, -7, 0}, {0, -1, 0, 2, 2, -7, 0}, {2, 0, 0, 2, 1, -7, 0}, {2, 0, 1, 0, 0, 6, 0}, {-2, 0, 2, 2, 2, 6, 0}, {-2, 0, 1, 2, 1, 6, 0}, {2, 0, -2, 0, 1, -6, 0}, {2, 0, 0, 0, 1, -6, 0}, {0, -1, 1, 0, 0, 5, 0}, {-2, -1, 0, 2, 1, -5, 0}, {-2, 0, 0, 0, 1, -5, 0}, {0, 0, 2, 2, 1, -5, 0}, {-2, 0, 2, 0, 1, 4, 0}, {-2, 1, 0, 2, 1, 4, 0}, {0, 0, 1, -2, 0, 4, 0}, {-1, 0, 1, 0, 0, -4, 0}, {-2, 1, 0, 0, 0, -4, 0}, {1, 0, 0, 0, 0, -4, 0}, {0, 0, 1, 2, 0, 3, 0}, {0, 0, -2, 2, 2, -3, 0}, {-1, -1, 1, 0, 0, -3, 0}, {0, 1, 1, 0, 0, -3, 0}, {0, -1, 1, 2, 2, -3, 0}, {2, -1, -1, 2, 2, -3, 0}, {0, 0, 3, 2, 2, -3, 0}, {2, -1, 0, 2, 2, -3, 0}}
	var s float64
	for i := 0; i < len(tp); i++ {
		s += (tp[i][5] + tp[i][6]) * Sin(d*tp[i][0]+m*tp[i][1]+n*tp[i][2]+f*tp[i][3]+o*tp[i][4])
	}
	//P=-17.20*Sin(o)-1.32*Sin(2*280.4665 + 36000.7698*t)-0.23*Sin(2*218.3165 + 481267.8813*t )+0.21*Sin(2*o);
	//return P/3600;
	return (s / 10000) / 3600
}

/*
 * 交角章动
 */
func JJZD(jd float64) float64 { //交角章动
	t := (jd - 2451545) / 36525
	//d = 297.85036 +455267.111480*t - 0.0019142*t*t+ t*t*t/189474;
	//m = 357.52772 + 35999.050340*t - 0.0001603*t*t- t*t*t/300000;
	//n= 134.96298 + 477198.867398*t + 0.0086972*t*t + t*t*t/56250;
	//f = 93.27191 + 483202.017538*t - 0.0036825*t*t + t*t*t/327270;
	d := 297.8502042 + 445267.1115168*t - 0.0016300*t*t + t*t*t/545868 - t*t*t*t/113065000
	m := SunM(jd)
	n := MoonM(jd)
	f := MoonLonX(jd)
	o := 125.04452 - 1934.136261*t + 0.0020708*t*t + t*t*t/450000
	tp := [][]float64{{0, 0, 0, 0, 1, 92025, 8.9 * t}, {-2, 0, 0, 2, 2, 5736, -3.1 * t}, {0, 0, 0, 2, 2, 977, -0.5 * t}, {0, 0, 0, 0, 2, -895, 0.5 * t}, {0, 1, 0, 0, 0, 54, -0.1 * t}, {0, 0, 1, 0, 0, -7, 0}, {-2, 1, 0, 2, 2, 224, -0.6 * t}, {0, 0, 0, 2, 1, 200, 0}, {0, 0, 1, 2, 2, 129, -0.1 * t}, {-2, -1, 0, 2, 2, -95, 0.3 * t}, {-2, 0, 0, 2, 1, -70, 0}, {0, 0, -1, 2, 2, -53, 0}, {2, 0, 0, 0, 0, 63, 0}, {0, 0, 1, 0, 1, -33, 0}, {2, 0, -1, 2, 2, 26, 0}, {0, 0, -1, 0, 1, 32, 0}, {0, 0, 1, 2, 1, 27, 0}, {0, 0, -2, 2, 1, -24, 0}, {2, 0, 0, 2, 2, 16, 0}, {0, 0, 2, 2, 2, 13, 0}, {-2, 0, 1, 2, 2, -12, 0}, {0, 0, -1, 2, 1, -10, 0}, {2, 0, -1, 0, 1, -8, 0}, {-2, 2, 0, 2, 2, 7, 0}, {0, 1, 0, 0, 1, 9, 0}, {-2, 0, 1, 0, 1, 7, 0}, {0, -1, 0, 0, 1, 6, 0}, {2, 0, -1, 2, 1, 5, 0}, {2, 0, 1, 2, 2, 3, 0}, {0, 1, 0, 2, 2, -3, 0}, {0, -1, 0, 2, 2, 3, 0}, {2, 0, 0, 2, 1, 3, 0}, {-2, 0, 2, 2, 2, -3, 0}, {-2, 0, 1, 2, 1, -3, 0}, {2, 0, -2, 0, 1, 3, 0}, {2, 0, 0, 0, 1, 3, 0}, {-2, -1, 0, 2, 1, 3, 0}, {-2, 0, 0, 0, 1, 3, 0}, {0, 0, 2, 2, 1, 3, 0}}
	var s float64 = 0
	for i := 0; i < len(tp); i++ {
		s += (tp[i][5] + tp[i][6]) * Cos(d*tp[i][0]+m*tp[i][1]+n*tp[i][2]+f*tp[i][3]+o*tp[i][4])
	}
	return s / 10000 / 3600
}

/*
@name 太阳几何黄经
*/
func SunLo(jd float64) float64 {
	T := (jd - 2451545) / 365250
	SunLo := 280.4664567 + 360007.6982779*T + 0.03032028*T*T + T*T*T/49931 - T*T*T*T/15299 - T*T*T*T*T/1988000
	return Limit360(SunLo)
}

func SunM(JD float64) float64 {
	T := (JD - 2451545) / 36525
	sunM := 357.5291092 + 35999.0502909*T - 0.0001559*T*T - 0.00000048*T*T*T
	return Limit360(sunM)
}

/*
@name 地球偏心率
*/
func Earthe(JD float64) float64 { //'地球偏心率
	T := (JD - 2451545) / 36525
	Earthe := 0.016708617 - 0.000042037*T - 0.0000001236*T*T
	return Earthe
}

func EarthPI(JD float64) float64 { //近日點經度
	T := (JD - 2451545) / 36525
	return 102.93735 + 1.71953*T + 000046*T*T
}
func SunMidFun(JD float64) float64 { //'太阳中间方程
	T := (JD - 2451545) / 36525
	M := SunM(JD)
	SunMidFun := (1.9146-0.004817*T-0.000014*T*T)*Sin(M) + (0.019993-0.000101*T)*Sin(2*M) + 0.00029*Sin(3*M)
	return SunMidFun
}
func SunTrueLo(JD float64) float64 { // '太阳真黄经
	SunTrueLo := SunLo(JD) + SunMidFun(JD)
	return SunTrueLo
}

func SunApparentLo(JD float64) float64 { //'太阳视黄经

	T := (JD - 2451545) / 36525
	SunApparentLo := SunTrueLo(JD) - 0.00569 - 0.00478*Sin(125.04-1934.136*T)
	return SunApparentLo
}

func SunApparentRa(JD float64) float64 { // '太阳视赤经
	return LoToRa(JD, SunApparentLo(JD), 0)
}

func SunApparentRaDec(JD float64) (float64, float64) {
	return LoBoToRaDec(JD, SunApparentLo(JD), 0)
}

func SunTrueRa(JD float64) float64 { //'太阳真赤经

	sitas := Sita(JD)
	SunTrueRa := ArcTan(Cos(sitas) * Sin(SunTrueLo(JD)) / Cos(SunTrueLo(JD)))
	//Select Case SunTrueLo(JD)
	tmp := SunTrueLo(JD)
	if tmp >= 90 && tmp < 180 {
		SunTrueRa = 180 + SunTrueRa
	} else if tmp >= 180 && tmp < 270 {
		SunTrueRa = 180 + SunTrueRa
	} else if tmp >= 270 && tmp <= 360 {
		SunTrueRa = 360 + SunTrueRa
	}
	return SunTrueRa
}

func SunApparentDec(JD float64) float64 { // '太阳视赤纬
	T := (JD - 2451545) / 36525
	sitas := Sita(JD) + 0.00256*Cos(125.04-1934.136*T)
	SunApparentDec := ArcSin(Sin(sitas) * Sin(SunApparentLo(JD)))
	return SunApparentDec
}

func SunTrueDec(JD float64) float64 { // '太阳真赤纬
	sitas := Sita(JD)
	SunTrueDec := ArcSin(Sin(sitas) * Sin(SunTrueLo(JD)))
	return SunTrueDec
}
func SunTime(JD float64) float64 { //均时差

	tm := (SunLo(JD) - 0.0057183 - (HSunApparentRa(JD)) + (HJZD(JD))*Cos(Sita(JD))) / 15
	if tm > 23 {
		tm = -24 + tm
	}
	return tm
}

func SunSC(Lo, JD float64) float64 { //黄道上的岁差，仅黄纬=0时

	t := (JD - 2451545) / 36525
	//n := 47.0029/3600*t - 0.03302/3600*t*t + 0.000060/3600*t*t*t
	//m := 174.876384/3600 - 869.8089/3600*t + 0.03536/3600*t*t
	pk := 5029.0966/3600.00*t + 1.11113/3600.00*t*t - 0.000006/3600.00*t*t*t
	return Lo + pk
}

func HSunTrueLo(JD float64) float64 {
	L := planet.WherePlanet(0, 0, JD)
	return L
}

func HSunTrueBo(JD float64) float64 {
	L := planet.WherePlanet(0, 1, JD)
	return L
}

func HSunApparentLo(JD float64) float64 {
	L := HSunTrueLo(JD)
	/*
		t := (JD - 2451545) / 365250.0
		R := planet.WherePlanet(-1, 2, JD)
		t2 := t * t
		t3 := t2 * t //千年数的各次方
		R += (-0.0020 + 0.0044*t + 0.0213*t2 - 0.0250*t3)
		L = L + HJZD(JD) - 20.4898/R/3600.00
	*/
	L = L + HJZD(JD) + SunLoGXC(JD)
	return L
}

func SunLoGXC(JD float64) float64 {
	R := planet.WherePlanet(0, 2, JD)
	return -20.49552 / R / 3600
}

func EarthAway(JD float64) float64 {
	//t=(JD - 2451545) / 365250;
	//R=Earth_R5(t)+Earth_R4(t)+Earth_R3(t)+Earth_R2(t)+Earth_R1(t)+Earth_R0(t);
	return planet.WherePlanet(0, 2, JD)
}

func HSunApparentRaDec(JD float64) (float64, float64) {
	return LoBoToRaDec(JD, HSunApparentLo(JD), HSunTrueBo(JD))
}

func HSunApparentRa(JD float64) float64 { // '太阳视赤经
	return LoToRa(JD, HSunApparentLo(JD), HSunTrueBo(JD))
}

func HSunTrueRa(JD float64) float64 { //'太阳真赤经
	tmp := HSunTrueLo(JD)
	sitas := Sita(JD)
	HSunTrueRa := ArcTan(Cos(sitas) * Sin(tmp) / Cos(tmp))
	//Select Case SunTrueLo(JD)
	if tmp >= 90 && tmp < 180 {
		HSunTrueRa = 180 + HSunTrueRa
	} else if tmp >= 180 && tmp < 270 {
		HSunTrueRa = 180 + HSunTrueRa
	} else if tmp >= 270 && tmp <= 360 {
		HSunTrueRa = 360 + HSunTrueRa
	}
	return HSunTrueRa
}

func HSunApparentDec(JD float64) float64 { // '太阳视赤纬
	T := (JD - 2451545) / 36525
	sitas := EclipticObliquity(JD, false) + 0.00256*Cos(125.04-1934.136*T)
	HSunApparentDec := ArcSin(Sin(sitas) * Sin(HSunApparentLo(JD)))
	return HSunApparentDec
}

func HSunTrueDec(JD float64) float64 { // '太阳真赤纬
	sitas := EclipticObliquity(JD, false)
	HSunTrueDec := ArcSin(Sin(sitas) * Sin(HSunTrueLo(JD)))
	return HSunTrueDec
}

func RDJL(jd float64) float64 { //ri di ju li
	f := SunMidFun(jd)
	m := SunM(jd)
	e := Earthe(jd)
	return (1.000001018 * (1 - e*e) / (1 + e*Cos(f+m)))
}

func GetMoonLoops(year float64, loop int) []float64 {
	var start float64
	var newMoon, tmp float64
	moon := make([]float64, loop)
	if year < 6000 {
		start = year + 11.00/12.00 + 5.00/30.00/12.00
	} else {
		start = year + 9.00/12.00 + 5.00/30.00/12.00
	}
	i := 1
	for j := 0; j < loop; j++ {
		if year > 3000 {
			newMoon = TD2UT(CalcMoonSH(start+float64(i-1)/12.5, 0)+8.0/24.0, false)
		} else {
			newMoon = TD2UT(CalcMoonS(start+float64(i-1)/12.5, 0)+8.0/24.0, false)
		}
		if i != 1 {
			if newMoon == tmp {
				j--
				i++
				continue
			}
		}
		moon[j] = newMoon
		tmp = moon[j]
		i++
		// echo DateCalc(moon[i])."<br />";
	}
	return moon
}

func GetJieqiLoops(year, loop int) []float64 {
	start := 270
	jq := make([]float64, loop)
	for i := 1; i <= loop; i++ {
		angle := start + 15*(i-1)
		if angle > 360 {
			angle -= 360
		}
		jq[i-1] = GetJQTime(year+int(math.Ceil(float64(i-1)/24.000)), angle) + 8.0/24.0
	}
	return jq
}

func GetJQTime(Year, Angle int) float64 { //节气时间
	var j int = 1
	var Day int
	var tp float64
	if Angle%2 == 0 {
		Day = 18
	} else {
		Day = 3
	}
	if Angle%10 != 0 {
		tp = float64(Angle+15.0) / 30.0
	} else {
		tp = float64(Angle) / 30.0
	}
	Month := 3 + tp
	if Month > 12 {
		Month -= 12
	}
	JD1 := JDECalc(int(Year), int(Month), float64(Day))
	if Angle == 0 {
		Angle = 360
	}
	for i := 0; i < j; i++ {
		for {
			JD0 := JD1
			stDegree := JQLospec(JD0) - float64(Angle)
			stDegreep := (JQLospec(JD0+0.000005) - JQLospec(JD0-0.000005)) / 0.00001
			JD1 = JD0 - stDegree/stDegreep
			if math.Abs(JD1-JD0) <= 0.00001 {
				break
			}
		}
		JD1 -= 0.001
	}
	JD1 += 0.001
	return TD2UT(JD1, false)
}

func JQLospec(JD float64) float64 {
	t := HSunApparentLo(JD)
	if t <= 12 {
		t += 360
	}
	return t
}

func GetXC(jd float64) string { //十二次
	tlo := HSunApparentLo(jd)
	if tlo >= 255 && tlo < 285 {
		return "星纪"
	} else if tlo >= 285 && tlo < 315 {
		return "玄枵"
	} else if tlo >= 315 && tlo < 345 {
		return "娵訾"
	} else if tlo >= 345 || tlo < 15 {
		return "降娄"
	} else if tlo >= 15 && tlo < 45 {
		return "大梁"
	} else if tlo >= 45 && tlo < 75 {
		return "实沈"
	} else if tlo >= 75 && tlo < 105 {
		return "鹑首"
	} else if tlo >= 105 && tlo < 135 {
		return "鹑火"
	} else if tlo >= 135 && tlo < 165 {
		return "鹑尾"
	} else if tlo >= 165 && tlo < 195 {
		return "寿星"
	} else if tlo >= 195 && tlo < 225 {
		return "大火"
	} else if tlo >= 225 && tlo < 255 {
		return "析木"
	}
	return ""
}

func GetWHTime(Year, Angle int) float64 {
	tmp := Angle
	var Day int
	var tp float64
	Angle = int(Angle/15) * 15
	if Angle%2 == 0 {
		Day = 18
	} else {
		Day = 3
	}
	if Angle%10 != 0 {
		tp = float64(Angle+15) / 30.0
	} else {
		tp = float64(Angle) / 30.0
	}
	Month := int(3 + tp)
	if Month > 12 {
		Month -= 12
	}
	JD1 := JDECalc(Year, Month, float64(Day))
	JD1 += float64(tmp - Angle)
	Angle = tmp
	if Angle <= 5 {
		Angle = 360 + Angle
	}
	for {
		JD0 := JD1
		stDegree := JQLospec(JD0) - float64(Angle)
		stDegreep := (JQLospec(JD0+0.000005) - JQLospec(JD0-0.000005)) / 0.00001
		JD1 = JD0 - stDegree/stDegreep
		if math.Abs(JD1-JD0) <= 0.00001 {
			break
		}
	}
	return TD2UT(JD1, false)
}

/*
 * 太阳中天时刻，通过均时差计算
 */
func GetSunTZTime(JD, Lon, TZ float64) float64 { //实际中天时间
	JD = math.Floor(JD)
	tmp := (TZ*15 - Lon) * 4 / 60
	return JD + tmp/24.0 - SunTime(JD)/24.0
}

/*
 * 昏朦影传入 当天0时时刻
 */
func GetBanTime(JD, Lon, Lat, TZ, An float64) float64 {
	JD = math.Floor(JD) + 1.5
	ntz := math.Round(Lon / 15)
	tztime := GetSunTZTime(JD, Lon, ntz)
	if SunHeight(tztime, Lon, Lat, ntz) < An {
		return -2 //极夜
	}
	if SunHeight(tztime+0.5, Lon, Lat, ntz) > An {
		return -1 //极昼
	}
	tmp := (Sin(An) - Sin(HSunApparentDec(tztime))*Sin(Lat)) / (Cos(HSunApparentDec(tztime)) * Cos(Lat))
	var sundown float64
	if math.Abs(tmp) <= 1 && Lat < 85 {
		rzsc := ArcCos(tmp) / 15
		sundown = tztime + rzsc/24.0 + 35.0/24.0/60.0
	} else {
		sundown = tztime
		i := 0
		for LowSunHeight(sundown, Lon, Lat, ntz) > An {
			i++
			sundown += 15.0 / 60.0 / 24.0
			if i > 48 {
				break
			}
		}
	}
	JD1 := sundown - 5.00/24.00/60.00
	for {
		JD0 := JD1
		stDegree := SunHeight(JD0, Lon, Lat, ntz) - An
		stDegreep := (SunHeight(JD0+0.000005, Lon, Lat, ntz) - SunHeight(JD0-0.000005, Lon, Lat, ntz)) / 0.00001
		JD1 = JD0 - stDegree/stDegreep
		if math.Abs(JD1-JD0) < 0.00001 {
			break
		}
	}
	return JD1 - ntz/24 + TZ/24
}

func GetAsaTime(JD, Lon, Lat, TZ, An float64) float64 {
	JD = math.Floor(JD) + 1.5
	ntz := math.Round(Lon / 15)
	tztime := GetSunTZTime(JD, Lon, ntz)
	if SunHeight(tztime, Lon, Lat, ntz) < An {
		return -2 //极夜
	}
	if SunHeight(tztime-0.5, Lon, Lat, ntz) > An {
		return -1 //极昼
	}
	tmp := (Sin(An) - Sin(HSunApparentDec(tztime))*Sin(Lat)) / (Cos(HSunApparentDec(tztime)) * Cos(Lat))
	var sunrise float64
	if math.Abs(tmp) <= 1 && Lat < 85 {
		rzsc := ArcCos(tmp) / 15
		sunrise = tztime - rzsc/24 - 25.0/24.0/60.0
	} else {
		sunrise = tztime
		i := 0
		for LowSunHeight(sunrise, Lon, Lat, ntz) > An {
			i++
			sunrise -= 15.0 / 60.0 / 24.0
			if i > 48 {
				break
			}
		}
	}
	JD1 := sunrise - 5.00/24.00/60.00
	for {
		JD0 := JD1
		stDegree := SunHeight(JD0, Lon, Lat, ntz) - An
		stDegreep := (SunHeight(JD0+0.000005, Lon, Lat, ntz) - SunHeight(JD0-0.000005, Lon, Lat, ntz)) / 0.00001
		JD1 = JD0 - stDegree/stDegreep
		if math.Abs(JD1-JD0) < 0.00001 {
			break
		}
	}
	return JD1 - ntz/24 + TZ/24
}

/*
 * 太阳时角
 */
func SunTimeAngle(JD, Lon, Lat, TZ float64) float64 {
	startime := Limit360(ApparentSiderealTime(JD-TZ/24)*15 + Lon)
	timeangle := startime - HSunApparentRa(TD2UT(JD-TZ/24, true))
	if timeangle < 0 {
		timeangle += 360
	}
	return timeangle
}

/*
 * 精确计算，传入当日0时JDE
 */
func GetSunRiseTime(JD, Lon, Lat, TZ, ZS, HEI float64) float64 {
	var An float64
	JD = math.Floor(JD) + 1.5
	ntz := math.Round(Lon / 15)
	if ZS != 0 {
		An = -0.8333
	}
	An = An - HeightDegreeByLat(HEI, Lat)
	tztime := GetSunTZTime(JD, Lon, ntz)
	if SunHeight(tztime, Lon, Lat, ntz) < An {
		return -2 //极夜
	}
	if SunHeight(tztime-0.5, Lon, Lat, ntz) > An {
		return -1 //极昼
	}
	//(sin(ho)-sin(φ)*sin(δ2))/(cos(φ)*cos(δ2))
	tmp := (Sin(An) - Sin(HSunApparentDec(tztime))*Sin(Lat)) / (Cos(HSunApparentDec(tztime)) * Cos(Lat))
	var sunrise float64
	if math.Abs(tmp) <= 1 && Lat < 85 {
		rzsc := ArcCos(tmp) / 15
		sunrise = tztime - rzsc/24 - 25.0/24.0/60.0
	} else {
		sunrise = tztime
		i := 0
		//TODO:使用二分法计算
		for LowSunHeight(sunrise, Lon, Lat, ntz) > An {
			i++
			sunrise -= 15.0 / 60.0 / 24.0
			if i > 48 {
				break
			}
		}
	}
	JD1 := sunrise
	for {
		JD0 := JD1
		stDegree := SunHeight(JD0, Lon, Lat, ntz) - An
		stDegreep := (SunHeight(JD0+0.000005, Lon, Lat, ntz) - SunHeight(JD0-0.000005, Lon, Lat, ntz)) / 0.00001
		JD1 = JD0 - stDegree/stDegreep
		if math.Abs(JD1-JD0) <= 0.00001 {
			break
		}
	}
	return JD1 - ntz/24 + TZ/24
}
func GetSunSetTime(JD, Lon, Lat, TZ, ZS, HEI float64) float64 {
	var An float64
	JD = math.Floor(JD) + 1.5
	ntz := math.Round(Lon / 15)
	if ZS != 0 {
		An = -0.8333
	}
	An = An - HeightDegreeByLat(HEI, Lat)
	tztime := GetSunTZTime(JD, Lon, ntz)
	if SunHeight(tztime, Lon, Lat, ntz) < An {
		return -2 //极夜
	}
	if SunHeight(tztime+0.5, Lon, Lat, ntz) > An {
		return -1 //极昼
	}
	tmp := (Sin(An) - Sin(HSunApparentDec(tztime))*Sin(Lat)) / (Cos(HSunApparentDec(tztime)) * Cos(Lat))
	var sundown float64
	if math.Abs(tmp) <= 1 && Lat < 85 {
		rzsc := ArcCos(tmp) / 15
		sundown = tztime + rzsc/24.0 + 35.0/24.0/60.0
	} else {
		sundown = tztime
		i := 0
		for LowSunHeight(sundown, Lon, Lat, ntz) > An {
			i++
			sundown += 15.0 / 60.0 / 24.0
			if i > 48 {
				break
			}
		}
	}
	JD1 := sundown
	for {
		JD0 := JD1
		stDegree := SunHeight(JD0, Lon, Lat, ntz) - An
		stDegreep := (SunHeight(JD0+0.000005, Lon, Lat, ntz) - SunHeight(JD0-0.000005, Lon, Lat, ntz)) / 0.00001
		JD1 = JD0 - stDegree/stDegreep
		if math.Abs(JD1-JD0) <= 0.00001 {
			break
		}
	}
	return JD1 - ntz/24 + TZ/24
}

/*
 * 太阳高度角 世界时
 */
func SunHeight(JD, Lon, Lat, TZ float64) float64 {
	//tmp := (TZ*15 - Lon) * 4 / 60
	//truejd := JD - tmp/24
	calcjd := JD - TZ/24.0
	tjde := TD2UT(calcjd, true)
	st := Limit360(ApparentSiderealTime(calcjd)*15 + Lon)
	ra, dec := HSunApparentRaDec(tjde)
	H := Limit360(st - ra)
	tmp2 := Sin(Lat)*Sin(dec) + Cos(dec)*Cos(Lat)*Cos(H)
	return ArcSin(tmp2)
}
func LowSunHeight(JD, Lon, Lat, TZ float64) float64 {
	//tmp := (TZ*15 - Lon) * 4 / 60
	//truejd := JD - tmp/24
	calcjd := JD - TZ/24
	st := Limit360(ApparentSiderealTime(calcjd)*15 + Lon)
	H := Limit360(st - SunApparentRa(TD2UT(calcjd, true)))
	dec := SunApparentDec(TD2UT(calcjd, true))
	tmp2 := Sin(Lat)*Sin(dec) + Cos(dec)*Cos(Lat)*Cos(H)
	return ArcSin(tmp2)
}
func SunAngle(JD, Lon, Lat, TZ float64) float64 {
	//tmp := (TZ*15 - Lon) * 4 / 60
	//truejd := JD - tmp/24
	calcjd := JD - TZ/24
	st := Limit360(ApparentSiderealTime(calcjd)*15 + Lon)
	H := Limit360(st - HSunApparentRa(TD2UT(calcjd, true)))
	tmp2 := Sin(H) / (Cos(H)*Sin(Lat) - Tan(HSunApparentDec(TD2UT(calcjd, true)))*Cos(Lat))
	Angle := ArcTan(tmp2)
	if Angle < 0 {
		if H/15 < 12 {
			return Angle + 360
		}
		return Angle + 180
	}
	if H/15 < 12 {
		return Angle + 180
	}
	return Angle
}

/*
* 干支
 */
func GetGZ(year int) string {
	tiangan := []string{"庚", "辛", "壬", "癸", "甲", "乙", "丙", "丁", "戊", "已"}
	dizhi := []string{"申", "酉", "戌", "亥", "子", "丑", "寅", "卯", "辰", "巳", "午", "未"}
	t := year - (year / 10 * 10)
	d := year % 12
	return tiangan[t] + dizhi[d] + "年"
}
