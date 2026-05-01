package planet

import (
	. "github.com/starainrt/astro/tools"
)

type lowMoonTerm struct {
	d   int
	m   int
	mp  int
	f   int
	amp float64
}

var lowMoonITerms = []lowMoonTerm{
	{d: 0, m: 0, mp: 1, f: 0, amp: 6288744},
	{d: 2, m: 0, mp: -1, f: 0, amp: 1274027},
	{d: 2, m: 0, mp: 0, f: 0, amp: 658314},
	{d: 0, m: 0, mp: 2, f: 0, amp: 213618},
	{d: 0, m: 1, mp: 0, f: 0, amp: -185116},
	{d: 0, m: 0, mp: 0, f: 2, amp: -114332},
	{d: 2, m: 0, mp: -2, f: 0, amp: 58793},
	{d: 2, m: -1, mp: -1, f: 0, amp: 57066},
	{d: 2, m: 0, mp: 1, f: 0, amp: 53322},
	{d: 2, m: -1, mp: 0, f: 0, amp: 45758},
	{d: 0, m: 1, mp: -1, f: 0, amp: -40923},
	{d: 1, m: 0, mp: 0, f: 0, amp: -34720},
	{d: 0, m: 1, mp: 1, f: 0, amp: -30383},
	{d: 2, m: 0, mp: 0, f: -2, amp: 15327},
	{d: 0, m: 0, mp: 1, f: 2, amp: -12528},
	{d: 0, m: 0, mp: 1, f: -2, amp: 10980},
	{d: 4, m: 0, mp: -1, f: 0, amp: 10675},
	{d: 0, m: 0, mp: 3, f: 0, amp: 10034},
	{d: 4, m: 0, mp: -2, f: 0, amp: 8548},
	{d: 2, m: 1, mp: -1, f: 0, amp: -7888},
	{d: 2, m: 1, mp: 0, f: 0, amp: -6766},
	{d: 1, m: 0, mp: -1, f: 0, amp: -5163},
	{d: 1, m: 1, mp: 0, f: 0, amp: 4987},
	{d: 2, m: -1, mp: 1, f: 0, amp: 4036},
	{d: 2, m: 0, mp: 2, f: 0, amp: 3994},
	{d: 4, m: 0, mp: 0, f: 0, amp: 3861},
	{d: 2, m: 0, mp: -3, f: 0, amp: 3665},
	{d: 0, m: 1, mp: -2, f: 0, amp: -2689},
	{d: 2, m: 0, mp: -1, f: 2, amp: -2602},
	{d: 2, m: -1, mp: -2, f: 0, amp: 2390},
	{d: 1, m: 0, mp: 1, f: 0, amp: -2348},
	{d: 2, m: -2, mp: 0, f: 0, amp: 2236},
	{d: 0, m: 1, mp: 2, f: 0, amp: -2120},
	{d: 0, m: 2, mp: 0, f: 0, amp: -2069},
	{d: 2, m: -2, mp: -1, f: 0, amp: 2048},
	{d: 2, m: 0, mp: 1, f: -2, amp: -1773},
	{d: 2, m: 0, mp: 0, f: 2, amp: -1595},
	{d: 4, m: -1, mp: -1, f: 0, amp: 1215},
	{d: 0, m: 0, mp: 2, f: 2, amp: -1110},
	{d: 3, m: 0, mp: -1, f: 0, amp: -892},
	{d: 2, m: 1, mp: 1, f: 0, amp: -810},
	{d: 4, m: -1, mp: -2, f: 0, amp: 759},
	{d: 0, m: 2, mp: -1, f: 0, amp: -713},
	{d: 2, m: 2, mp: -1, f: 0, amp: -700},
	{d: 2, m: 1, mp: -2, f: 0, amp: 691},
	{d: 2, m: -1, mp: 0, f: -2, amp: 596},
	{d: 4, m: 0, mp: 1, f: 0, amp: 549},
	{d: 0, m: 0, mp: 4, f: 0, amp: 537},
	{d: 4, m: -1, mp: 0, f: 0, amp: 520},
	{d: 1, m: 0, mp: -2, f: 0, amp: -487},
	{d: 2, m: 1, mp: 0, f: -2, amp: -399},
	{d: 0, m: 0, mp: 2, f: -2, amp: -381},
	{d: 1, m: 1, mp: 1, f: 0, amp: 351},
	{d: 3, m: 0, mp: -2, f: 0, amp: -340},
	{d: 4, m: 0, mp: -3, f: 0, amp: 330},
	{d: 2, m: -1, mp: 2, f: 0, amp: 327},
	{d: 0, m: 2, mp: 1, f: 0, amp: -323},
	{d: 1, m: 1, mp: -1, f: 0, amp: 299},
	{d: 2, m: 0, mp: 3, f: 0, amp: 294},
	{d: 2, m: 0, mp: -1, f: -2, amp: 0},
}

var lowMoonRTerms = []lowMoonTerm{
	{d: 0, m: 0, mp: 1, f: 0, amp: -20905355},
	{d: 2, m: 0, mp: -1, f: 0, amp: -3699111},
	{d: 2, m: 0, mp: 0, f: 0, amp: -2955968},
	{d: 0, m: 0, mp: 2, f: 0, amp: -569925},
	{d: 0, m: 1, mp: 0, f: 0, amp: 48888},
	{d: 0, m: 0, mp: 0, f: 2, amp: -3149},
	{d: 2, m: 0, mp: -2, f: 0, amp: 246158},
	{d: 2, m: -1, mp: -1, f: 0, amp: -152138},
	{d: 2, m: 0, mp: 1, f: 0, amp: -170733},
	{d: 2, m: -1, mp: 0, f: 0, amp: -204586},
	{d: 0, m: 1, mp: -1, f: 0, amp: -129620},
	{d: 1, m: 0, mp: 0, f: 0, amp: 108743},
	{d: 0, m: 1, mp: 1, f: 0, amp: 104755},
	{d: 2, m: 0, mp: 0, f: -2, amp: 10321},
	{d: 0, m: 0, mp: 1, f: 2, amp: 0},
	{d: 0, m: 0, mp: 1, f: -2, amp: 79661},
	{d: 4, m: 0, mp: -1, f: 0, amp: -34782},
	{d: 0, m: 0, mp: 3, f: 0, amp: -23210},
	{d: 4, m: 0, mp: -2, f: 0, amp: -21636},
	{d: 2, m: 1, mp: -1, f: 0, amp: 24208},
	{d: 2, m: 1, mp: 0, f: 0, amp: 30824},
	{d: 1, m: 0, mp: -1, f: 0, amp: -8379},
	{d: 1, m: 1, mp: 0, f: 0, amp: -16675},
	{d: 2, m: -1, mp: 1, f: 0, amp: -12831},
	{d: 2, m: 0, mp: 2, f: 0, amp: -10445},
	{d: 4, m: 0, mp: 0, f: 0, amp: -11650},
	{d: 2, m: 0, mp: -3, f: 0, amp: 14403},
	{d: 0, m: 1, mp: -2, f: 0, amp: -7003},
	{d: 2, m: 0, mp: -1, f: 2, amp: 0},
	{d: 2, m: -1, mp: -2, f: 0, amp: 10056},
	{d: 1, m: 0, mp: 1, f: 0, amp: 6322},
	{d: 2, m: -2, mp: 0, f: 0, amp: -9884},
	{d: 0, m: 1, mp: 2, f: 0, amp: 5751},
	{d: 0, m: 2, mp: 0, f: 0, amp: 0},
	{d: 2, m: -2, mp: -1, f: 0, amp: -4950},
	{d: 2, m: 0, mp: 1, f: -2, amp: 4130},
	{d: 2, m: 0, mp: 0, f: 2, amp: 0},
	{d: 4, m: -1, mp: -1, f: 0, amp: -3958},
	{d: 0, m: 0, mp: 2, f: 2, amp: 0},
	{d: 3, m: 0, mp: -1, f: 0, amp: 3258},
	{d: 2, m: 1, mp: 1, f: 0, amp: 2616},
	{d: 4, m: -1, mp: -2, f: 0, amp: -1897},
	{d: 0, m: 2, mp: -1, f: 0, amp: -2117},
	{d: 2, m: 2, mp: -1, f: 0, amp: 2354},
	{d: 2, m: 1, mp: -2, f: 0, amp: 0},
	{d: 2, m: -1, mp: 0, f: -2, amp: 0},
	{d: 4, m: 0, mp: 1, f: 0, amp: -1423},
	{d: 0, m: 0, mp: 4, f: 0, amp: -1117},
	{d: 4, m: -1, mp: 0, f: 0, amp: -1571},
	{d: 1, m: 0, mp: -2, f: 0, amp: -1739},
	{d: 2, m: 1, mp: 0, f: -2, amp: 0},
	{d: 0, m: 0, mp: 2, f: -2, amp: -4421},
	{d: 1, m: 1, mp: 1, f: 0, amp: 0},
	{d: 3, m: 0, mp: -2, f: 0, amp: 0},
	{d: 4, m: 0, mp: -3, f: 0, amp: 0},
	{d: 2, m: -1, mp: 2, f: 0, amp: 0},
	{d: 0, m: 2, mp: 1, f: 0, amp: 1165},
	{d: 1, m: 1, mp: -1, f: 0, amp: 0},
	{d: 2, m: 0, mp: 3, f: 0, amp: 0},
	{d: 2, m: 0, mp: -1, f: -2, amp: 8752},
}

var lowMoonBTerms = []lowMoonTerm{
	{d: 0, m: 0, mp: 0, f: 1, amp: 5128122},
	{d: 0, m: 0, mp: 1, f: 1, amp: 280602},
	{d: 0, m: 0, mp: 1, f: -1, amp: 277693},
	{d: 2, m: 0, mp: 0, f: -1, amp: 173237},
	{d: 2, m: 0, mp: -1, f: 1, amp: 55413},
	{d: 2, m: 0, mp: -1, f: -1, amp: 46271},
	{d: 2, m: 0, mp: 0, f: 1, amp: 32573},
	{d: 0, m: 0, mp: 2, f: 1, amp: 17198},
	{d: 2, m: 0, mp: 1, f: -1, amp: 9266},
	{d: 0, m: 0, mp: 2, f: -1, amp: 8822},
	{d: 2, m: -1, mp: 0, f: -1, amp: 8216},
	{d: 2, m: 0, mp: -2, f: -1, amp: 4324},
	{d: 2, m: 0, mp: 1, f: 1, amp: 4200},
	{d: 2, m: 1, mp: 0, f: -1, amp: -3359},
	{d: 2, m: -1, mp: -1, f: 1, amp: 2463},
	{d: 2, m: -1, mp: 0, f: 1, amp: 2211},
	{d: 2, m: -1, mp: -1, f: -1, amp: 2065},
	{d: 0, m: 1, mp: -1, f: -1, amp: -1870},
	{d: 4, m: 0, mp: -1, f: -1, amp: 1828},
	{d: 0, m: 1, mp: 0, f: 1, amp: -1794},
	{d: 0, m: 0, mp: 0, f: 3, amp: -1749},
	{d: 0, m: 1, mp: -1, f: 1, amp: -1565},
	{d: 1, m: 0, mp: 0, f: 1, amp: -1491},
	{d: 0, m: 1, mp: 1, f: 1, amp: -1475},
	{d: 0, m: 1, mp: 1, f: -1, amp: -1410},
	{d: 0, m: 1, mp: 0, f: -1, amp: -1344},
	{d: 1, m: 0, mp: 0, f: -1, amp: -1335},
	{d: 0, m: 0, mp: 3, f: 1, amp: 1107},
	{d: 4, m: 0, mp: 0, f: -1, amp: 1021},
	{d: 4, m: 0, mp: -1, f: 1, amp: 833},
	{d: 0, m: 0, mp: 1, f: -3, amp: 777},
	{d: 4, m: 0, mp: -2, f: 1, amp: 671},
	{d: 2, m: 0, mp: 0, f: -3, amp: 607},
	{d: 2, m: 0, mp: 2, f: -1, amp: 596},
	{d: 2, m: -1, mp: 1, f: -1, amp: 491},
	{d: 2, m: 0, mp: -2, f: 1, amp: -451},
	{d: 0, m: 0, mp: 3, f: -1, amp: 439},
	{d: 2, m: 0, mp: 2, f: 1, amp: 422},
	{d: 2, m: 0, mp: -3, f: -1, amp: 421},
	{d: 2, m: 1, mp: -1, f: 1, amp: -366},
	{d: 2, m: 1, mp: 0, f: 1, amp: -351},
	{d: 4, m: 0, mp: 0, f: 1, amp: 331},
	{d: 2, m: -1, mp: 1, f: 1, amp: 315},
	{d: 2, m: -2, mp: 0, f: -1, amp: 302},
	{d: 0, m: 0, mp: 1, f: 3, amp: -283},
	{d: 2, m: 1, mp: 1, f: -1, amp: -229},
	{d: 1, m: 1, mp: 0, f: -1, amp: 223},
	{d: 1, m: 1, mp: 0, f: 1, amp: 223},
	{d: 0, m: 1, mp: -2, f: -1, amp: -220},
	{d: 2, m: 1, mp: -1, f: -1, amp: -220},
	{d: 1, m: 0, mp: 1, f: 1, amp: -185},
	{d: 2, m: -1, mp: -2, f: -1, amp: 181},
	{d: 0, m: 1, mp: 2, f: 1, amp: -177},
	{d: 4, m: 0, mp: -2, f: -1, amp: 176},
	{d: 4, m: -1, mp: -1, f: -1, amp: 166},
	{d: 1, m: 0, mp: 1, f: -1, amp: -164},
	{d: 4, m: 0, mp: 1, f: -1, amp: 132},
	{d: 1, m: 0, mp: -1, f: -1, amp: -119},
	{d: 4, m: -1, mp: 0, f: -1, amp: 115},
	{d: 2, m: -2, mp: 0, f: 1, amp: 107},
}

func MoonLo(JD float64) float64 { //'月球平黄经
	T := (JD - 2451545) / 36525
	MoonLo := 218.3164591 + 481267.88134236*T - 0.0013268*T*T + T*T*T/538841 - T*T*T*T/65194000
	return MoonLo
}

func SunMoonAngle(JD float64) float64 { // '月日距角
	T := (JD - 2451545) / 36525
	SunMoonAngle := 297.8502042 + 445267.1115168*T - 0.00163*T*T + T*T*T/545868 - T*T*T*T/113065000
	return SunMoonAngle
}

func MoonM(JD float64) float64 { // '月平近点角
	T := (JD - 2451545) / 36525
	MoonM := 134.9634114 + 477198.8676313*T + 0.008997*T*T + T*T*T/69699 - T*T*T*T/14712000
	return MoonM
}

func MoonLonX(JD float64) float64 { // As Double '月球经度参数(到升交点的平角距离)
	T := (JD - 2451545) / 36525
	MoonLonX := 93.2720993 + 483202.0175273*T - 0.0034029*T*T - T*T*T/3526000 + T*T*T*T/863310000
	return MoonLonX
}

func lowMoonTermValue(term lowMoonTerm, D, IsunM, IMoonM, F, E float64, trig func(float64) float64) float64 {
	arg := float64(term.d)*D + float64(term.m)*IsunM + float64(term.mp)*IMoonM + float64(term.f)*F
	switch Abs(term.m) {
	case 1:
		return trig(arg) * term.amp * E
	case 2:
		return trig(arg) * term.amp * E * E
	default:
		return trig(arg) * term.amp
	}
}

func MoonI(JD float64) float64 {
	T := (JD - 2451545) / 36525
	D := Limit360(SunMoonAngle(JD))
	IsunM := SunM(JD)
	IMoonM := MoonM(JD)
	F := Limit360(MoonLonX(JD))
	E := 1 - 0.002516*T - 0.0000074*T*T
	A1 := 119.75 + 131.849*T
	A2 := Limit360(53.09 + 479264.29*T)
	var MoonI float64
	for _, term := range lowMoonITerms {
		MoonI += lowMoonTermValue(term, D, IsunM, IMoonM, F, E, Sin)
	}
	MoonI = MoonI + 3958*Sin(A1) + 1962*Sin(MoonLo(JD)-F) + 318*Sin(A2)
	return FR(MoonI)
}

func MoonR(JD float64) float64 {
	T := (JD - 2451545) / 36525
	D := SunMoonAngle(JD)
	IsunM := SunM(JD)
	IMoonM := Limit360(MoonM(JD))
	F := Limit360(MoonLonX(JD))
	E := 1 - 0.002516*T - 0.0000074*T*T
	var MoonR float64
	for _, term := range lowMoonRTerms {
		MoonR += lowMoonTermValue(term, D, IsunM, IMoonM, F, E, Cos)
	}
	return MoonR
}

func MoonB(JD float64) float64 {
	T := (JD - 2451545) / 36525
	D := Limit360(SunMoonAngle(JD))
	IsunM := Limit360(SunM(JD))
	IMoonM := Limit360(MoonM(JD))
	F := Limit360(MoonLonX(JD))
	E := 1 - 0.002516*T - 0.0000074*T*T
	A1 := Limit360(119.75 + 131.849*T)
	A3 := Limit360(313.45 + 481266.484*T)
	var MoonB float64
	for _, term := range lowMoonBTerms {
		MoonB += lowMoonTermValue(term, D, IsunM, IMoonM, F, E, Sin)
	}
	MoonB += -2235*Sin(MoonLo(JD)) + 382*Sin(A3) + 175*Sin(A1-F) + 175*Sin(A1+F) + 127*Sin(MoonLo(JD)-IMoonM) - 115*Sin(MoonLo(JD)+IMoonM)
	return MoonB
}

func MoonTrueLo(JD float64) float64 {
	return Limit360(MoonLo(JD) + (MoonI(JD) / 1000000))
}

func MoonTrueBo(JD float64) float64 {
	return MoonB(JD) / 1000000
}

func MoonAway(JD float64) float64 { //'月地距离
	MoonAway := 385000.56 + MoonR(JD)/1000
	return MoonAway
}
