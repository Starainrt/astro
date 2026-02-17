package basic

import (
	"math"

	. "github.com/starainrt/astro/tools"
)

func MoonLo(JD float64) float64 { //'月球平黄经
	T := (JD - 2451545) / 36525
	MoonLo := 218.3164591 + 481267.88134236*T - 0.0013268*T*T + T*T*T/538841 - T*T*T*T/65194000
	return MoonLo
}

func SunMoonAngle(JD float64) float64 { // '月日距角
	// Dim T As Double
	T := (JD - 2451545) / 36525
	SunMoonAngle := 297.8502042 + 445267.1115168*T - 0.00163*T*T + T*T*T/545868 - T*T*T*T/113065000
	return SunMoonAngle
}

func MoonM(JD float64) float64 { // '月平近点角
	//Dim T As Double
	T := (JD - 2451545) / 36525
	MoonM := 134.9634114 + 477198.8676313*T + 0.008997*T*T + T*T*T/69699 - T*T*T*T/14712000
	return MoonM
}

func MoonLonX(JD float64) float64 { // As Double '月球经度参数(到升交点的平角距离)
	//Dim T As Double
	T := (JD - 2451545) / 36525
	MoonLonX := 93.2720993 + 483202.0175273*T - 0.0034029*T*T - T*T*T/3526000 + T*T*T*T/863310000
	return MoonLonX
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
	//A3 := Limit360(313.45 + 481266.484 * T);
	//die(D." ".F." ".E);
	Ifun := make(map[int]map[int]int)
	for i := 1; i <= 60; i++ {
		Ifun[i] = make(map[int]int)
	}
	Ifun[1][1] = 0
	Ifun[1][2] = 0
	Ifun[1][3] = 1
	Ifun[1][4] = 0
	Ifun[1][5] = 6288744
	Ifun[2][1] = 2
	Ifun[2][2] = 0
	Ifun[2][3] = -1
	Ifun[2][4] = 0
	Ifun[2][5] = 1274027
	Ifun[3][1] = 2
	Ifun[3][2] = 0
	Ifun[3][3] = 0
	Ifun[3][4] = 0
	Ifun[3][5] = 658314
	Ifun[4][1] = 0
	Ifun[4][2] = 0
	Ifun[4][3] = 2
	Ifun[4][4] = 0
	Ifun[4][5] = 213618
	Ifun[5][1] = 0
	Ifun[5][2] = 1
	Ifun[5][3] = 0
	Ifun[5][4] = 0
	Ifun[5][5] = -185116
	Ifun[6][1] = 0
	Ifun[6][2] = 0
	Ifun[6][3] = 0
	Ifun[6][4] = 2
	Ifun[6][5] = -114332
	Ifun[7][1] = 2
	Ifun[7][2] = 0
	Ifun[7][3] = -2
	Ifun[7][4] = 0
	Ifun[7][5] = 58793
	Ifun[8][1] = 2
	Ifun[8][2] = -1
	Ifun[8][3] = -1
	Ifun[8][4] = 0
	Ifun[8][5] = 57066
	Ifun[9][1] = 2
	Ifun[9][2] = 0
	Ifun[9][3] = 1
	Ifun[9][4] = 0
	Ifun[9][5] = 53322
	Ifun[10][1] = 2
	Ifun[10][2] = -1
	Ifun[10][3] = 0
	Ifun[10][4] = 0
	Ifun[10][5] = 45758
	Ifun[11][1] = 0
	Ifun[11][2] = 1
	Ifun[11][3] = -1
	Ifun[11][4] = 0
	Ifun[11][5] = -40923
	Ifun[12][1] = 1
	Ifun[12][2] = 0
	Ifun[12][3] = 0
	Ifun[12][4] = 0
	Ifun[12][5] = -34720
	Ifun[13][1] = 0
	Ifun[13][2] = 1
	Ifun[13][3] = 1
	Ifun[13][4] = 0
	Ifun[13][5] = -30383
	Ifun[14][1] = 2
	Ifun[14][2] = 0
	Ifun[14][3] = 0
	Ifun[14][4] = -2
	Ifun[14][5] = 15327
	Ifun[15][1] = 0
	Ifun[15][2] = 0
	Ifun[15][3] = 1
	Ifun[15][4] = 2
	Ifun[15][5] = -12528
	Ifun[16][1] = 0
	Ifun[16][2] = 0
	Ifun[16][3] = 1
	Ifun[16][4] = -2
	Ifun[16][5] = 10980
	Ifun[17][1] = 4
	Ifun[17][2] = 0
	Ifun[17][3] = -1
	Ifun[17][4] = 0
	Ifun[17][5] = 10675
	Ifun[18][1] = 0
	Ifun[18][2] = 0
	Ifun[18][3] = 3
	Ifun[18][4] = 0
	Ifun[18][5] = 10034
	Ifun[19][1] = 4
	Ifun[19][2] = 0
	Ifun[19][3] = -2
	Ifun[19][4] = 0
	Ifun[19][5] = 8548
	Ifun[20][1] = 2
	Ifun[20][2] = 1
	Ifun[20][3] = -1
	Ifun[20][4] = 0
	Ifun[20][5] = -7888
	Ifun[21][1] = 2
	Ifun[21][2] = 1
	Ifun[21][3] = 0
	Ifun[21][4] = 0
	Ifun[21][5] = -6766
	Ifun[22][1] = 1
	Ifun[22][2] = 0
	Ifun[22][3] = -1
	Ifun[22][4] = 0
	Ifun[22][5] = -5163
	Ifun[23][1] = 1
	Ifun[23][2] = 1
	Ifun[23][3] = 0
	Ifun[23][4] = 0
	Ifun[23][5] = 4987
	Ifun[24][1] = 2
	Ifun[24][2] = -1
	Ifun[24][3] = 1
	Ifun[24][4] = 0
	Ifun[24][5] = 4036
	Ifun[25][1] = 2
	Ifun[25][2] = 0
	Ifun[25][3] = 2
	Ifun[25][4] = 0
	Ifun[25][5] = 3994
	Ifun[26][1] = 4
	Ifun[26][2] = 0
	Ifun[26][3] = 0
	Ifun[26][4] = 0
	Ifun[26][5] = 3861
	Ifun[27][1] = 2
	Ifun[27][2] = 0
	Ifun[27][3] = -3
	Ifun[27][4] = 0
	Ifun[27][5] = 3665
	Ifun[28][1] = 0
	Ifun[28][2] = 1
	Ifun[28][3] = -2
	Ifun[28][4] = 0
	Ifun[28][5] = -2689
	Ifun[29][1] = 2
	Ifun[29][2] = 0
	Ifun[29][3] = -1
	Ifun[29][4] = 2
	Ifun[29][5] = -2602
	Ifun[30][1] = 2
	Ifun[30][2] = -1
	Ifun[30][3] = -2
	Ifun[30][4] = 0
	Ifun[30][5] = 2390
	Ifun[31][1] = 1
	Ifun[31][2] = 0
	Ifun[31][3] = 1
	Ifun[31][4] = 0
	Ifun[31][5] = -2348
	Ifun[32][1] = 2
	Ifun[32][2] = -2
	Ifun[32][3] = 0
	Ifun[32][4] = 0
	Ifun[32][5] = 2236
	Ifun[33][1] = 0
	Ifun[33][2] = 1
	Ifun[33][3] = 2
	Ifun[33][4] = 0
	Ifun[33][5] = -2120
	Ifun[34][1] = 0
	Ifun[34][2] = 2
	Ifun[34][3] = 0
	Ifun[34][4] = 0
	Ifun[34][5] = -2069
	Ifun[35][1] = 2
	Ifun[35][2] = -2
	Ifun[35][3] = -1
	Ifun[35][4] = 0
	Ifun[35][5] = 2048
	Ifun[36][1] = 2
	Ifun[36][2] = 0
	Ifun[36][3] = 1
	Ifun[36][4] = -2
	Ifun[36][5] = -1773
	Ifun[37][1] = 2
	Ifun[37][2] = 0
	Ifun[37][3] = 0
	Ifun[37][4] = 2
	Ifun[37][5] = -1595
	Ifun[38][1] = 4
	Ifun[38][2] = -1
	Ifun[38][3] = -1
	Ifun[38][4] = 0
	Ifun[38][5] = 1215
	Ifun[39][1] = 0
	Ifun[39][2] = 0
	Ifun[39][3] = 2
	Ifun[39][4] = 2
	Ifun[39][5] = -1110
	Ifun[40][1] = 3
	Ifun[40][2] = 0
	Ifun[40][3] = -1
	Ifun[40][4] = 0
	Ifun[40][5] = -892
	Ifun[41][1] = 2
	Ifun[41][2] = 1
	Ifun[41][3] = 1
	Ifun[41][4] = 0
	Ifun[41][5] = -810
	Ifun[42][1] = 4
	Ifun[42][2] = -1
	Ifun[42][3] = -2
	Ifun[42][4] = 0
	Ifun[42][5] = 759
	Ifun[43][1] = 0
	Ifun[43][2] = 2
	Ifun[43][3] = -1
	Ifun[43][4] = 0
	Ifun[43][5] = -713
	Ifun[44][1] = 2
	Ifun[44][2] = 2
	Ifun[44][3] = -1
	Ifun[44][4] = 0
	Ifun[44][5] = -700
	Ifun[45][1] = 2
	Ifun[45][2] = 1
	Ifun[45][3] = -2
	Ifun[45][4] = 0
	Ifun[45][5] = 691
	Ifun[46][1] = 2
	Ifun[46][2] = -1
	Ifun[46][3] = 0
	Ifun[46][4] = -2
	Ifun[46][5] = 596
	Ifun[47][1] = 4
	Ifun[47][2] = 0
	Ifun[47][3] = 1
	Ifun[47][4] = 0
	Ifun[47][5] = 549
	Ifun[48][1] = 0
	Ifun[48][2] = 0
	Ifun[48][3] = 4
	Ifun[48][4] = 0
	Ifun[48][5] = 537
	Ifun[49][1] = 4
	Ifun[49][2] = -1
	Ifun[49][3] = 0
	Ifun[49][4] = 0
	Ifun[49][5] = 520
	Ifun[50][1] = 1
	Ifun[50][2] = 0
	Ifun[50][3] = -2
	Ifun[50][4] = 0
	Ifun[50][5] = -487
	Ifun[51][1] = 2
	Ifun[51][2] = 1
	Ifun[51][3] = 0
	Ifun[51][4] = -2
	Ifun[51][5] = -399
	Ifun[52][1] = 0
	Ifun[52][2] = 0
	Ifun[52][3] = 2
	Ifun[52][4] = -2
	Ifun[52][5] = -381
	Ifun[53][1] = 1
	Ifun[53][2] = 1
	Ifun[53][3] = 1
	Ifun[53][4] = 0
	Ifun[53][5] = 351
	Ifun[54][1] = 3
	Ifun[54][2] = 0
	Ifun[54][3] = -2
	Ifun[54][4] = 0
	Ifun[54][5] = -340
	Ifun[55][1] = 4
	Ifun[55][2] = 0
	Ifun[55][3] = -3
	Ifun[55][4] = 0
	Ifun[55][5] = 330
	Ifun[56][1] = 2
	Ifun[56][2] = -1
	Ifun[56][3] = 2
	Ifun[56][4] = 0
	Ifun[56][5] = 327
	Ifun[57][1] = 0
	Ifun[57][2] = 2
	Ifun[57][3] = 1
	Ifun[57][4] = 0
	Ifun[57][5] = -323
	Ifun[58][1] = 1
	Ifun[58][2] = 1
	Ifun[58][3] = -1
	Ifun[58][4] = 0
	Ifun[58][5] = 299
	Ifun[59][1] = 2
	Ifun[59][2] = 0
	Ifun[59][3] = 3
	Ifun[59][4] = 0
	Ifun[59][5] = 294
	Ifun[60][1] = 2
	Ifun[60][2] = 0
	Ifun[60][3] = -1
	Ifun[60][4] = -2
	Ifun[60][5] = 0
	var MoonI float64
	var TEMP float64
	for i := 1; i < 61; i++ {
		if Abs(Ifun[i][2]) == 1 {
			TEMP = Sin(float64(Ifun[i][1])*D+float64(Ifun[i][2])*IsunM+float64(Ifun[i][3])*IMoonM+float64(Ifun[i][4])*F) * float64(Ifun[i][5]) * E
		} else if Abs(Ifun[i][2]) == 2 {
			TEMP = Sin(float64(Ifun[i][1])*D+float64(Ifun[i][2])*IsunM+float64(Ifun[i][3])*IMoonM+float64(Ifun[i][4])*F) * float64(Ifun[i][5]) * E * E
		} else {
			TEMP = Sin(float64(Ifun[i][1])*D+float64(Ifun[i][2])*IsunM+float64(Ifun[i][3])*IMoonM+float64(Ifun[i][4])*F) * float64(Ifun[i][5])
		}
		MoonI = MoonI + TEMP
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
	Ifun := make(map[int]map[int]float64)
	for i := 1; i <= 60; i++ {
		Ifun[i] = make(map[int]float64)
	}
	Ifun[1][1] = 0
	Ifun[1][2] = 0
	Ifun[1][3] = 1
	Ifun[1][4] = 0
	Ifun[1][5] = -20905355
	Ifun[2][1] = 2
	Ifun[2][2] = 0
	Ifun[2][3] = -1
	Ifun[2][4] = 0
	Ifun[2][5] = -3699111
	Ifun[3][1] = 2
	Ifun[3][2] = 0
	Ifun[3][3] = 0
	Ifun[3][4] = 0
	Ifun[3][5] = -2955968
	Ifun[4][1] = 0
	Ifun[4][2] = 0
	Ifun[4][3] = 2
	Ifun[4][4] = 0
	Ifun[4][5] = -569925
	Ifun[5][1] = 0
	Ifun[5][2] = 1
	Ifun[5][3] = 0
	Ifun[5][4] = 0
	Ifun[5][5] = 48888
	Ifun[6][1] = 0
	Ifun[6][2] = 0
	Ifun[6][3] = 0
	Ifun[6][4] = 2
	Ifun[6][5] = -3149
	Ifun[7][1] = 2
	Ifun[7][2] = 0
	Ifun[7][3] = -2
	Ifun[7][4] = 0
	Ifun[7][5] = 246158
	Ifun[8][1] = 2
	Ifun[8][2] = -1
	Ifun[8][3] = -1
	Ifun[8][4] = 0
	Ifun[8][5] = -152138
	Ifun[9][1] = 2
	Ifun[9][2] = 0
	Ifun[9][3] = 1
	Ifun[9][4] = 0
	Ifun[9][5] = -170733
	Ifun[10][1] = 2
	Ifun[10][2] = -1
	Ifun[10][3] = 0
	Ifun[10][4] = 0
	Ifun[10][5] = -204586
	Ifun[11][1] = 0
	Ifun[11][2] = 1
	Ifun[11][3] = -1
	Ifun[11][4] = 0
	Ifun[11][5] = -129620
	Ifun[12][1] = 1
	Ifun[12][2] = 0
	Ifun[12][3] = 0
	Ifun[12][4] = 0
	Ifun[12][5] = 108743
	Ifun[13][1] = 0
	Ifun[13][2] = 1
	Ifun[13][3] = 1
	Ifun[13][4] = 0
	Ifun[13][5] = 104755
	Ifun[14][1] = 2
	Ifun[14][2] = 0
	Ifun[14][3] = 0
	Ifun[14][4] = -2
	Ifun[14][5] = 10321
	Ifun[15][1] = 0
	Ifun[15][2] = 0
	Ifun[15][3] = 1
	Ifun[15][4] = 2
	Ifun[15][5] = 0
	Ifun[16][1] = 0
	Ifun[16][2] = 0
	Ifun[16][3] = 1
	Ifun[16][4] = -2
	Ifun[16][5] = 79661
	Ifun[17][1] = 4
	Ifun[17][2] = 0
	Ifun[17][3] = -1
	Ifun[17][4] = 0
	Ifun[17][5] = -34782
	Ifun[18][1] = 0
	Ifun[18][2] = 0
	Ifun[18][3] = 3
	Ifun[18][4] = 0
	Ifun[18][5] = -23210
	Ifun[19][1] = 4
	Ifun[19][2] = 0
	Ifun[19][3] = -2
	Ifun[19][4] = 0
	Ifun[19][5] = -21636
	Ifun[20][1] = 2
	Ifun[20][2] = 1
	Ifun[20][3] = -1
	Ifun[20][4] = 0
	Ifun[20][5] = 24208
	Ifun[21][1] = 2
	Ifun[21][2] = 1
	Ifun[21][3] = 0
	Ifun[21][4] = 0
	Ifun[21][5] = 30824
	Ifun[22][1] = 1
	Ifun[22][2] = 0
	Ifun[22][3] = -1
	Ifun[22][4] = 0
	Ifun[22][5] = -8379
	Ifun[23][1] = 1
	Ifun[23][2] = 1
	Ifun[23][3] = 0
	Ifun[23][4] = 0
	Ifun[23][5] = -16675
	Ifun[24][1] = 2
	Ifun[24][2] = -1
	Ifun[24][3] = 1
	Ifun[24][4] = 0
	Ifun[24][5] = -12831
	Ifun[25][1] = 2
	Ifun[25][2] = 0
	Ifun[25][3] = 2
	Ifun[25][4] = 0
	Ifun[25][5] = -10445
	Ifun[26][1] = 4
	Ifun[26][2] = 0
	Ifun[26][3] = 0
	Ifun[26][4] = 0
	Ifun[26][5] = -11650
	Ifun[27][1] = 2
	Ifun[27][2] = 0
	Ifun[27][3] = -3
	Ifun[27][4] = 0
	Ifun[27][5] = 14403
	Ifun[28][1] = 0
	Ifun[28][2] = 1
	Ifun[28][3] = -2
	Ifun[28][4] = 0
	Ifun[28][5] = -7003
	Ifun[29][1] = 2
	Ifun[29][2] = 0
	Ifun[29][3] = -1
	Ifun[29][4] = 2
	Ifun[29][5] = 0
	Ifun[30][1] = 2
	Ifun[30][2] = -1
	Ifun[30][3] = -2
	Ifun[30][4] = 0
	Ifun[30][5] = 10056
	Ifun[31][1] = 1
	Ifun[31][2] = 0
	Ifun[31][3] = 1
	Ifun[31][4] = 0
	Ifun[31][5] = 6322
	Ifun[32][1] = 2
	Ifun[32][2] = -2
	Ifun[32][3] = 0
	Ifun[32][4] = 0
	Ifun[32][5] = -9884
	Ifun[33][1] = 0
	Ifun[33][2] = 1
	Ifun[33][3] = 2
	Ifun[33][4] = 0
	Ifun[33][5] = 5751
	Ifun[34][1] = 0
	Ifun[34][2] = 2
	Ifun[34][3] = 0
	Ifun[34][4] = 0
	Ifun[34][5] = 0
	Ifun[35][1] = 2
	Ifun[35][2] = -2
	Ifun[35][3] = -1
	Ifun[35][4] = 0
	Ifun[35][5] = -4950
	Ifun[36][1] = 2
	Ifun[36][2] = 0
	Ifun[36][3] = 1
	Ifun[36][4] = -2
	Ifun[36][5] = 4130
	Ifun[37][1] = 2
	Ifun[37][2] = 0
	Ifun[37][3] = 0
	Ifun[37][4] = 2
	Ifun[37][5] = 0
	Ifun[38][1] = 4
	Ifun[38][2] = -1
	Ifun[38][3] = -1
	Ifun[38][4] = 0
	Ifun[38][5] = -3958
	Ifun[39][1] = 0
	Ifun[39][2] = 0
	Ifun[39][3] = 2
	Ifun[39][4] = 2
	Ifun[39][5] = 0
	Ifun[40][1] = 3
	Ifun[40][2] = 0
	Ifun[40][3] = -1
	Ifun[40][4] = 0
	Ifun[40][5] = 3258
	Ifun[41][1] = 2
	Ifun[41][2] = 1
	Ifun[41][3] = 1
	Ifun[41][4] = 0
	Ifun[41][5] = 2616
	Ifun[42][1] = 4
	Ifun[42][2] = -1
	Ifun[42][3] = -2
	Ifun[42][4] = 0
	Ifun[42][5] = -1897
	Ifun[43][1] = 0
	Ifun[43][2] = 2
	Ifun[43][3] = -1
	Ifun[43][4] = 0
	Ifun[43][5] = -2117
	Ifun[44][1] = 2
	Ifun[44][2] = 2
	Ifun[44][3] = -1
	Ifun[44][4] = 0
	Ifun[44][5] = 2354
	Ifun[45][1] = 2
	Ifun[45][2] = 1
	Ifun[45][3] = -2
	Ifun[45][4] = 0
	Ifun[45][5] = 0
	Ifun[46][1] = 2
	Ifun[46][2] = -1
	Ifun[46][3] = 0
	Ifun[46][4] = -2
	Ifun[46][5] = 0
	Ifun[47][1] = 4
	Ifun[47][2] = 0
	Ifun[47][3] = 1
	Ifun[47][4] = 0
	Ifun[47][5] = -1423
	Ifun[48][1] = 0
	Ifun[48][2] = 0
	Ifun[48][3] = 4
	Ifun[48][4] = 0
	Ifun[48][5] = -1117
	Ifun[49][1] = 4
	Ifun[49][2] = -1
	Ifun[49][3] = 0
	Ifun[49][4] = 0
	Ifun[49][5] = -1571
	Ifun[50][1] = 1
	Ifun[50][2] = 0
	Ifun[50][3] = -2
	Ifun[50][4] = 0
	Ifun[50][5] = -1739
	Ifun[51][1] = 2
	Ifun[51][2] = 1
	Ifun[51][3] = 0
	Ifun[51][4] = -2
	Ifun[51][5] = 0
	Ifun[52][1] = 0
	Ifun[52][2] = 0
	Ifun[52][3] = 2
	Ifun[52][4] = -2
	Ifun[52][5] = -4421
	Ifun[53][1] = 1
	Ifun[53][2] = 1
	Ifun[53][3] = 1
	Ifun[53][4] = 0
	Ifun[53][5] = 0
	Ifun[54][1] = 3
	Ifun[54][2] = 0
	Ifun[54][3] = -2
	Ifun[54][4] = 0
	Ifun[54][5] = 0
	Ifun[55][1] = 4
	Ifun[55][2] = 0
	Ifun[55][3] = -3
	Ifun[55][4] = 0
	Ifun[55][5] = 0
	Ifun[56][1] = 2
	Ifun[56][2] = -1
	Ifun[56][3] = 2
	Ifun[56][4] = 0
	Ifun[56][5] = 0
	Ifun[57][1] = 0
	Ifun[57][2] = 2
	Ifun[57][3] = 1
	Ifun[57][4] = 0
	Ifun[57][5] = 1165
	Ifun[58][1] = 1
	Ifun[58][2] = 1
	Ifun[58][3] = -1
	Ifun[58][4] = 0
	Ifun[58][5] = 0
	Ifun[59][1] = 2
	Ifun[59][2] = 0
	Ifun[59][3] = 3
	Ifun[59][4] = 0
	Ifun[59][5] = 0
	Ifun[60][1] = 2
	Ifun[60][2] = 0
	Ifun[60][3] = -1
	Ifun[60][4] = -2
	Ifun[60][5] = 8752
	var MoonR, TEMP float64 = 0, 0
	for i := 1; i < 61; i++ {
		if math.Abs(Ifun[i][2]) == float64(1) {
			TEMP = Cos(Ifun[i][1]*D+Ifun[i][2]*IsunM+Ifun[i][3]*IMoonM+Ifun[i][4]*F) * Ifun[i][5] * E
		} else if math.Abs(Ifun[i][2]) == float64(2) {
			TEMP = Cos(Ifun[i][1]*D+Ifun[i][2]*IsunM+Ifun[i][3]*IMoonM+Ifun[i][4]*F) * Ifun[i][5] * E * E
		} else {
			TEMP = Cos(Ifun[i][1]*D+Ifun[i][2]*IsunM+Ifun[i][3]*IMoonM+Ifun[i][4]*F) * Ifun[i][5]
		}
		MoonR = MoonR + TEMP
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
	Ifun := make(map[int]map[int]float64)
	for i := 1; i <= 60; i++ {
		Ifun[i] = make(map[int]float64)
	}
	//die(IsunM." ".IMoonM." ".A3);
	Ifun[1][1] = 0
	Ifun[1][2] = 0
	Ifun[1][3] = 0
	Ifun[1][4] = 1
	Ifun[1][5] = 5128122
	Ifun[2][1] = 0
	Ifun[2][2] = 0
	Ifun[2][3] = 1
	Ifun[2][4] = 1
	Ifun[2][5] = 280602
	Ifun[3][1] = 0
	Ifun[3][2] = 0
	Ifun[3][3] = 1
	Ifun[3][4] = -1
	Ifun[3][5] = 277693
	Ifun[4][1] = 2
	Ifun[4][2] = 0
	Ifun[4][3] = 0
	Ifun[4][4] = -1
	Ifun[4][5] = 173237
	Ifun[5][1] = 2
	Ifun[5][2] = 0
	Ifun[5][3] = -1
	Ifun[5][4] = 1
	Ifun[5][5] = 55413
	Ifun[6][1] = 2
	Ifun[6][2] = 0
	Ifun[6][3] = -1
	Ifun[6][4] = -1
	Ifun[6][5] = 46271
	Ifun[7][1] = 2
	Ifun[7][2] = 0
	Ifun[7][3] = 0
	Ifun[7][4] = 1
	Ifun[7][5] = 32573
	Ifun[8][1] = 0
	Ifun[8][2] = 0
	Ifun[8][3] = 2
	Ifun[8][4] = 1
	Ifun[8][5] = 17198
	Ifun[9][1] = 2
	Ifun[9][2] = 0
	Ifun[9][3] = 1
	Ifun[9][4] = -1
	Ifun[9][5] = 9266
	Ifun[10][1] = 0
	Ifun[10][2] = 0
	Ifun[10][3] = 2
	Ifun[10][4] = -1
	Ifun[10][5] = 8822
	Ifun[11][1] = 2
	Ifun[11][2] = -1
	Ifun[11][3] = 0
	Ifun[11][4] = -1
	Ifun[11][5] = 8216
	Ifun[12][1] = 2
	Ifun[12][2] = 0
	Ifun[12][3] = -2
	Ifun[12][4] = -1
	Ifun[12][5] = 4324
	Ifun[13][1] = 2
	Ifun[13][2] = 0
	Ifun[13][3] = 1
	Ifun[13][4] = 1
	Ifun[13][5] = 4200
	Ifun[14][1] = 2
	Ifun[14][2] = 1
	Ifun[14][3] = 0
	Ifun[14][4] = -1
	Ifun[14][5] = -3359
	Ifun[15][1] = 2
	Ifun[15][2] = -1
	Ifun[15][3] = -1
	Ifun[15][4] = 1
	Ifun[15][5] = 2463
	Ifun[16][1] = 2
	Ifun[16][2] = -1
	Ifun[16][3] = 0
	Ifun[16][4] = 1
	Ifun[16][5] = 2211
	Ifun[17][1] = 2
	Ifun[17][2] = -1
	Ifun[17][3] = -1
	Ifun[17][4] = -1
	Ifun[17][5] = 2065
	Ifun[18][1] = 0
	Ifun[18][2] = 1
	Ifun[18][3] = -1
	Ifun[18][4] = -1
	Ifun[18][5] = -1870
	Ifun[19][1] = 4
	Ifun[19][2] = 0
	Ifun[19][3] = -1
	Ifun[19][4] = -1
	Ifun[19][5] = 1828
	Ifun[20][1] = 0
	Ifun[20][2] = 1
	Ifun[20][3] = 0
	Ifun[20][4] = 1
	Ifun[20][5] = -1794
	Ifun[21][1] = 0
	Ifun[21][2] = 0
	Ifun[21][3] = 0
	Ifun[21][4] = 3
	Ifun[21][5] = -1749
	Ifun[22][1] = 0
	Ifun[22][2] = 1
	Ifun[22][3] = -1
	Ifun[22][4] = 1
	Ifun[22][5] = -1565
	Ifun[23][1] = 1
	Ifun[23][2] = 0
	Ifun[23][3] = 0
	Ifun[23][4] = 1
	Ifun[23][5] = -1491
	Ifun[24][1] = 0
	Ifun[24][2] = 1
	Ifun[24][3] = 1
	Ifun[24][4] = 1
	Ifun[24][5] = -1475
	Ifun[25][1] = 0
	Ifun[25][2] = 1
	Ifun[25][3] = 1
	Ifun[25][4] = -1
	Ifun[25][5] = -1410
	Ifun[26][1] = 0
	Ifun[26][2] = 1
	Ifun[26][3] = 0
	Ifun[26][4] = -1
	Ifun[26][5] = -1344
	Ifun[27][1] = 1
	Ifun[27][2] = 0
	Ifun[27][3] = 0
	Ifun[27][4] = -1
	Ifun[27][5] = -1335
	Ifun[28][1] = 0
	Ifun[28][2] = 0
	Ifun[28][3] = 3
	Ifun[28][4] = 1
	Ifun[28][5] = 1107
	Ifun[29][1] = 4
	Ifun[29][2] = 0
	Ifun[29][3] = 0
	Ifun[29][4] = -1
	Ifun[29][5] = 1021
	Ifun[30][1] = 4
	Ifun[30][2] = 0
	Ifun[30][3] = -1
	Ifun[30][4] = 1
	Ifun[30][5] = 833
	Ifun[31][1] = 0
	Ifun[31][2] = 0
	Ifun[31][3] = 1
	Ifun[31][4] = -3
	Ifun[31][5] = 777
	Ifun[32][1] = 4
	Ifun[32][2] = 0
	Ifun[32][3] = -2
	Ifun[32][4] = 1
	Ifun[32][5] = 671
	Ifun[33][1] = 2
	Ifun[33][2] = 0
	Ifun[33][3] = 0
	Ifun[33][4] = -3
	Ifun[33][5] = 607
	Ifun[34][1] = 2
	Ifun[34][2] = 0
	Ifun[34][3] = 2
	Ifun[34][4] = -1
	Ifun[34][5] = 596
	Ifun[35][1] = 2
	Ifun[35][2] = -1
	Ifun[35][3] = 1
	Ifun[35][4] = -1
	Ifun[35][5] = 491
	Ifun[36][1] = 2
	Ifun[36][2] = 0
	Ifun[36][3] = -2
	Ifun[36][4] = 1
	Ifun[36][5] = -451
	Ifun[37][1] = 0
	Ifun[37][2] = 0
	Ifun[37][3] = 3
	Ifun[37][4] = -1
	Ifun[37][5] = 439
	Ifun[38][1] = 2
	Ifun[38][2] = 0
	Ifun[38][3] = 2
	Ifun[38][4] = 1
	Ifun[38][5] = 422
	Ifun[39][1] = 2
	Ifun[39][2] = 0
	Ifun[39][3] = -3
	Ifun[39][4] = -1
	Ifun[39][5] = 421
	Ifun[40][1] = 2
	Ifun[40][2] = 1
	Ifun[40][3] = -1
	Ifun[40][4] = 1
	Ifun[40][5] = -366
	Ifun[41][1] = 2
	Ifun[41][2] = 1
	Ifun[41][3] = 0
	Ifun[41][4] = 1
	Ifun[41][5] = -351
	Ifun[42][1] = 4
	Ifun[42][2] = 0
	Ifun[42][3] = 0
	Ifun[42][4] = 1
	Ifun[42][5] = 331
	Ifun[43][1] = 2
	Ifun[43][2] = -1
	Ifun[43][3] = 1
	Ifun[43][4] = 1
	Ifun[43][5] = 315
	Ifun[44][1] = 2
	Ifun[44][2] = -2
	Ifun[44][3] = 0
	Ifun[44][4] = -1
	Ifun[44][5] = 302
	Ifun[45][1] = 0
	Ifun[45][2] = 0
	Ifun[45][3] = 1
	Ifun[45][4] = 3
	Ifun[45][5] = -283
	Ifun[46][1] = 2
	Ifun[46][2] = 1
	Ifun[46][3] = 1
	Ifun[46][4] = -1
	Ifun[46][5] = -229
	Ifun[47][1] = 1
	Ifun[47][2] = 1
	Ifun[47][3] = 0
	Ifun[47][4] = -1
	Ifun[47][5] = 223
	Ifun[48][1] = 1
	Ifun[48][2] = 1
	Ifun[48][3] = 0
	Ifun[48][4] = 1
	Ifun[48][5] = 223
	Ifun[49][1] = 0
	Ifun[49][2] = 1
	Ifun[49][3] = -2
	Ifun[49][4] = -1
	Ifun[49][5] = -220
	Ifun[50][1] = 2
	Ifun[50][2] = 1
	Ifun[50][3] = -1
	Ifun[50][4] = -1
	Ifun[50][5] = -220
	Ifun[51][1] = 1
	Ifun[51][2] = 0
	Ifun[51][3] = 1
	Ifun[51][4] = 1
	Ifun[51][5] = -185
	Ifun[52][1] = 2
	Ifun[52][2] = -1
	Ifun[52][3] = -2
	Ifun[52][4] = -1
	Ifun[52][5] = 181
	Ifun[53][1] = 0
	Ifun[53][2] = 1
	Ifun[53][3] = 2
	Ifun[53][4] = 1
	Ifun[53][5] = -177
	Ifun[54][1] = 4
	Ifun[54][2] = 0
	Ifun[54][3] = -2
	Ifun[54][4] = -1
	Ifun[54][5] = 176
	Ifun[55][1] = 4
	Ifun[55][2] = -1
	Ifun[55][3] = -1
	Ifun[55][4] = -1
	Ifun[55][5] = 166
	Ifun[56][1] = 1
	Ifun[56][2] = 0
	Ifun[56][3] = 1
	Ifun[56][4] = -1
	Ifun[56][5] = -164
	Ifun[57][1] = 4
	Ifun[57][2] = 0
	Ifun[57][3] = 1
	Ifun[57][4] = -1
	Ifun[57][5] = 132
	Ifun[58][1] = 1
	Ifun[58][2] = 0
	Ifun[58][3] = -1
	Ifun[58][4] = -1
	Ifun[58][5] = -119
	Ifun[59][1] = 4
	Ifun[59][2] = -1
	Ifun[59][3] = 0
	Ifun[59][4] = -1
	Ifun[59][5] = 115
	Ifun[60][1] = 2
	Ifun[60][2] = -2
	Ifun[60][3] = 0
	Ifun[60][4] = 1
	Ifun[60][5] = 107
	var TEMP, MoonB float64 = 0, 0
	for i := 1; i < 61; i++ {
		if math.Abs(Ifun[i][2]) == float64(1) {
			TEMP = Sin(Ifun[i][1]*D+Ifun[i][2]*IsunM+Ifun[i][3]*IMoonM+Ifun[i][4]*F) * Ifun[i][5] * E
		} else if math.Abs(Ifun[i][2]) == float64(2) {
			TEMP = Sin(Ifun[i][1]*D+Ifun[i][2]*IsunM+Ifun[i][3]*IMoonM+Ifun[i][4]*F) * Ifun[i][5] * E * E
		} else {
			TEMP = Sin(Ifun[i][1]*D+Ifun[i][2]*IsunM+Ifun[i][3]*IMoonM+Ifun[i][4]*F) * Ifun[i][5]
		}
		MoonB = MoonB + TEMP
	}
	//MoonB = MoonB + 3958 * Sin(A1) + 1962 * Sin(MoonLo(JD) - F) + 318 * Sin(A2);
	MoonB += -2235*Sin(MoonLo(JD)) + 382*Sin(A3) + 175*Sin(A1-F) + 175*Sin(A1+F) + 127*Sin(MoonLo(JD)-IMoonM) - 115*Sin(MoonLo(JD)+IMoonM)
	return MoonB
}

func MoonTrueLo(JD float64) float64 {
	return (Limit360(MoonLo(JD) + (MoonI(JD) / 1000000)))
}
func MoonTrueBo(JD float64) float64 {
	return (MoonB(JD) / 1000000)
}
func MoonAway(JD float64) float64 { //'月地距离
	MoonAway := 385000.56 + MoonR(JD)/1000
	return MoonAway
}

/*
 * @name 月球视黄经
 */
func MoonApparentLo(JD float64) float64 {
	return MoonTrueLo(JD) + Nutation2000Bi(JD)
}

/*
 * 月球真赤纬
 */
func MoonTrueDec(JD float64) float64 {
	MoonLo := MoonApparentLo(JD)
	MoonBo := MoonTrueBo(JD)
	tmp := Sin(MoonBo)*Cos(Sita(JD)) + Cos(MoonBo)*Sin(Sita(JD))*Sin(MoonLo)
	res := ArcSin(tmp)
	return res
}

/*
 * 月球真赤经
 */
func MoonTrueRa(JD float64) float64 {
	return LoToRa(JD, MoonApparentLo(JD), MoonTrueBo(JD))
}

func MoonTrueRaDec(JD float64) (float64, float64) {
	return LoBoToRaDec(JD, MoonApparentLo(JD), MoonTrueBo(JD))
}

/*
*

	*
	传入世界时
*/
func MoonApparentRa(JD, lon, lat float64, tz int) float64 {
	jde := TD2UT(JD, true)
	ra := MoonTrueRa(jde - float64(tz)/24.000)
	dec := MoonTrueDec(jde - float64(tz)/24.000)
	away := MoonAway(jde-float64(tz)/24.000) / 149597870.7
	nra := ZhanXinRa(ra, dec, lat, lon, JD-float64(tz)/24.000, away, 0)
	return nra
}

func MoonApparentDec(JD, lon, lat, tz float64) float64 {
	jde := TD2UT(JD, true)
	ra := MoonTrueRa(jde - tz/24.0)
	dec := MoonTrueDec(jde - tz/24)
	away := MoonAway(jde-tz/24) / 149597870.7
	ndec := ZhanXinDec(ra, dec, lat, lon, JD-tz/24, away, 0)
	return ndec
}

func MoonPhase(JD float64) float64 {
	MoonBo := HMoonTrueBo(JD)
	SunLo := HSunApparentLo(JD)
	MoonLo := HMoonApparentLo(JD)
	tmp := Cos(MoonBo) * Cos(SunLo-MoonLo)
	R := Distance(JD) * 149597870.691
	i := R * Sin(ArcCos(tmp)) / (HMoonAway(JD) - R*tmp)
	i = ArcTan(i)
	if i < 0 {
		i += 180
	}
	if i > 180 {
		i -= 180
	}
	k := (1 + Cos(i)) / 2
	return k
}

func SunMoonSeek(JDE float64, degree float64) float64 {
	p := HMoonApparentLo(JDE) - (HSunApparentLo(JDE)) - degree
	for p < -180 {
		p += 360
	}
	for p > 180 {
		p -= 360
	}
	return p
}

func CalcMoonSHByJDE(JDE float64, C int) float64 {
	C = C * 180
	JD1 := JDE
	for {
		JD0 := JD1
		stDegree := SunMoonSeek(JD0, float64(C))
		stDegreep := (SunMoonSeek(JD0+0.000005, float64(C)) - SunMoonSeek(JD0-0.000005, float64(C))) / 0.00001
		JD1 = JD0 - stDegree/stDegreep
		if math.Abs(JD1-JD0) <= 0.00001 {
			break
		}
	}
	return JD1
}

func CalcMoonSH(Year float64, C int) float64 {
	JDE := CalcMoonS(Year, C)
	C = C * 180
	JD1 := JDE
	for {
		JD0 := JD1
		stDegree := SunMoonSeek(JD0, float64(C))
		stDegreep := (SunMoonSeek(JD0+0.000005, float64(C)) - SunMoonSeek(JD0-0.000005, float64(C))) / 0.00001
		JD1 = JD0 - stDegree/stDegreep
		if math.Abs(JD1-JD0) <= 0.00001 {
			break
		}
	}
	return JD1
}

/*
 * C=0朔月时刻 =1 望月
 */
func CalcMoonS(Year float64, C int) float64 {
	k := math.Floor((Year - 2000) * 12.36827)
	if C == 1 {
		k += 0.5
	}
	T := k / 1236.85
	JDE := 2451550.09765 + 29.530588853*k + 0.0001337*T*T - 0.000000150*T*T*T + 0.00000000073*T*T*T*T
	//太阳平近点角：
	M := Limit360(2.5534 + 29.10535669*k - 0.0000218*T*T - 0.00000011*T*T*T)
	//月亮的平近点角：
	N := Limit360(201.5643 + 385.81693528*k + 0.0107438*T*T + 0.00001239*T*T*T - 0.000000058*T*T*T*T)
	//月亮的纬度参数：
	F := Limit360(160.7108 + 390.67050274*k - 0.0016341*T*T - 0.00000227*T*T*T + 0.000000011*T*T*T*T)
	//月亮轨道升交点经度：
	O := Limit360(124.7746 - 1.56375580*k + 0.0020691*T*T + 0.00000215*T*T*T)
	E := 1 - 0.002516*T - 0.0000074*T*T
	//die(E." ".M." ".N." ".F." ".O);
	ZQ := []float64{N, M, 2 * N, 2 * F, N - M, N + M, 2 * M, N - 2*F, N + 2*F, 2*N + M, 3 * N, M + 2*F, M - 2*F, 2*N - M, O, N + 2*M, 2*N - 2*F, 3 * M, N + M - 2*F, 2*N + 2*F, N + M + 2*F, N - M + 2*F, N - M - 2*F, 3*N + M, 4 * N}
	var MN []float64
	if C == 0 {
		MN = []float64{-0.40720, 0.17241 * E, 0.01608, 0.01039, 0.00739 * E, -0.00514 * E, 0.00208 * E * E, -0.00111, -0.00057, 0.00056 * E, -0.00042, 0.00042 * E, 0.00038 * E, -0.00024 * E, -0.00017, -0.00007, 0.00004, 0.00004, 0.00003, 0.00003, -0.00003, 0.00003, -0.00002, -0.00002, 0.00002}
	} else {
		MN = []float64{-0.40614, 0.17302 * E, 0.01614, 0.01043, 0.00734 * E, -0.00515 * E, 0.00209 * E * E, -0.00111, -0.00057, 0.00056 * E, -0.00042, 0.00042 * E, 0.00038 * E, -0.00024 * E, -0.00017, -0.00007, 0.00004, 0.00004, 0.00003, 0.00003, -0.00003, 0.00003, -0.00002, -0.00002, 0.00002}
	}
	var tmp float64 = 0
	for k, v := range ZQ {
		tmp += Sin(v) * MN[k]
	}
	//die(tmp);
	A1 := 299.77 + 0.107408*k - 0.009173*T*T
	A2 := 251.88 + 0.016321*k
	A3 := 251.83 + 26.651886*k
	A4 := 349.42 + 36.412478*k
	A5 := 84.66 + 18.206239*k
	A6 := 141.74 + 53.303771*k
	A7 := 207.14 + 2.453732*k
	A8 := 154.84 + 7.306860*k
	A9 := 34.52 + 27.261239*k
	A10 := 207.19 + 0.121824*k
	A11 := 291.34 + 1.844379*k
	A12 := 161.72 + 24.198154*k
	A13 := 239.56 + 25.513099*k
	A14 := 331.55 + 3.592518*k
	tmp2 := 325*Sin(A1) + 165*Sin(A2) + 164*Sin(A3) + 126*Sin(A4) + 110*Sin(A5) + 62*Sin(A6) + 60*Sin(A7) + 56*Sin(A8) + 47*Sin(A9) + 42*Sin(A10) + 40*Sin(A11) + 37*Sin(A12) + 35*Sin(A13) + 23*Sin(A14)
	tmp2 /= 1000000
	JDE = JDE + tmp2 + tmp
	return JDE
}

func CalcMoonXHByJDE(JDE float64, C int) float64 {
	if C == 0 {
		C = 90
	} else {
		C = -90
	}
	JD1 := JDE
	for {
		JD0 := JD1
		stDegree := SunMoonSeek(JD0, float64(C))
		stDegreep := (SunMoonSeek(JD0+0.000005, float64(C)) - SunMoonSeek(JD0-0.000005, float64(C))) / 0.00001
		JD1 = JD0 - stDegree/stDegreep
		if math.Abs(JD1-JD0) <= 0.00001 {
			break
		}
	}
	return JD1
}

func CalcMoonXH(Year float64, C int) float64 {
	JDE := CalcMoonX(Year, C)
	if C == 0 {
		C = 90
	} else {
		C = -90
	}
	JD1 := JDE
	for {
		JD0 := JD1
		stDegree := SunMoonSeek(JD0, float64(C))
		stDegreep := (SunMoonSeek(JD0+0.000005, float64(C)) - SunMoonSeek(JD0-0.000005, float64(C))) / 0.00001
		JD1 = JD0 - stDegree/stDegreep
		if math.Abs(JD1-JD0) <= 0.00001 {
			break
		}
	}
	return JD1
}

func CalcMoonX(Year float64, C int) float64 {
	k := math.Floor((Year-2000)*12.36827) + 0.25
	if C == 1 {
		k += 0.5
	}
	T := k / 1236.85
	JDE := 2451550.09765 + 29.530588853*k + 0.0001337*T*T - 0.000000150*T*T*T + 0.00000000073*T*T*T*T
	//太阳平近点角：
	M := Limit360(2.5534 + 29.10535669*k - 0.0000218*T*T - 0.00000011*T*T*T)
	//月亮的平近点角：
	N := Limit360(201.5643 + 385.81693528*k + 0.0107438*T*T + 0.00001239*T*T*T - 0.000000058*T*T*T*T)
	//月亮的纬度参数：
	F := Limit360(160.7108 + 390.67050274*k - 0.0016341*T*T - 0.00000227*T*T*T + 0.000000011*T*T*T*T)
	//月亮轨道升交点经度：
	O := Limit360(124.7746 - 1.56375580*k + 0.0020691*T*T + 0.00000215*T*T*T)
	E := 1 - 0.002516*T - 0.0000074*T*T
	//die(E." ".M." ".N." ".F." ".O);
	ZQ := []float64{N, M, N + M, 2 * N, 2 * F, N - M, 2 * M, N - 2*F, N + 2*F, 3 * N, 2*N - M, M + 2*F, M - 2*F, N + 2*M, 2*N + M, O, N - M - 2*F, 2*N + 2*F, N + M + 2*F, N - 2*F, N + M - 2*F, 3 * M, 2*N - 2*F, N - M + 2*F, M + 3*N}
	MN := []float64{-0.62801, 0.17172 * E, -0.01183 * E, 0.00862, 0.00804, 0.00454 * E, 0.00204 * E * E, -0.00180, -0.00070, -0.00040, -0.00034 * E, 0.00032 * E, 0.00032 * E, -0.00028 * E * E, 0.00027 * E, -0.00017, -0.00005, 0.00004, -0.00004, 0.00004, 0.00003, 0.00003, 0.00002, 0.00002, -0.00002}
	var tmp float64 = 0
	for k, v := range ZQ {
		tmp += Sin(v) * MN[k]
	}
	W := 0.00306 - 0.00038*E*Cos(M) + 0.00026*Cos(N) - 0.00002*Cos(N-M) + 0.00002*Cos(N+M) + 0.00002*Cos(2*F)
	A1 := 299.77 + 0.107408*k - 0.009173*T*T
	A2 := 251.88 + 0.016321*k
	A3 := 251.83 + 26.651886*k
	A4 := 349.42 + 36.412478*k
	A5 := 84.66 + 18.206239*k
	A6 := 141.74 + 53.303771*k
	A7 := 207.14 + 2.453732*k
	A8 := 154.84 + 7.306860*k
	A9 := 34.52 + 27.261239*k
	A10 := 207.19 + 0.121824*k
	A11 := 291.34 + 1.844379*k
	A12 := 161.72 + 24.198154*k
	A13 := 239.56 + 25.513099*k
	A14 := 331.55 + 3.592518*k
	tmp2 := 325*Sin(A1) + 165*Sin(A2) + 164*Sin(A3) + 126*Sin(A4) + 110*Sin(A5) + 62*Sin(A6) + 60*Sin(A7) + 56*Sin(A8) + 47*Sin(A9) + 42*Sin(A10) + 40*Sin(A11) + 37*Sin(A12) + 35*Sin(A13) + 23*Sin(A14)
	tmp2 /= 1000000
	//die(tmp2);
	//die(JDE." ".tmp." ".tmp2." ".W);
	JDE = JDE + tmp2 + tmp
	if C == 0 {
		JDE += W
	} else {
		JDE -= W
	}
	return JDE
}

/*
 * 月球方位角
 */
func MoonAngle(JD, Lon, Lat, TZ float64) float64 {

	//tmp := (TZ*15 - Lon) * 4 / 60
	calcjd := TD2UT(JD-TZ/24, true)
	ra := MoonTrueRa(calcjd)
	dec := MoonTrueDec(calcjd)
	away := MoonAway(calcjd) / 149597870.7
	ndec := ZhanXinDec(ra, dec, Lat, Lon, JD-TZ/24, away, 0)
	nra := ZhanXinRa(ra, dec, Lat, Lon, JD-TZ/24, away, 0)
	calcjd = JD - TZ/24
	st := Limit360(ApparentSiderealTime(calcjd)*15 + Lon)
	H := Limit360(st - nra)
	tmp2 := Sin(H) / (Cos(H)*Sin(Lat) - Tan(ndec)*Cos(Lat))
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

func MoonHeight(JD, Lon, Lat, TZ float64) float64 {
	//	tmp := (TZ*15 - Lon) * 4 / 60
	//truejd=JD-tmp/24;
	calcjd := TD2UT(JD-TZ/24, true)
	ra := MoonTrueRa(calcjd)
	dec := MoonTrueDec(calcjd)
	away := MoonAway(calcjd) / 149597870.7
	ndec := ZhanXinDec(ra, dec, Lat, Lon, JD-TZ/24, away, 0)
	nra := ZhanXinRa(ra, dec, Lat, Lon, JD-TZ/24, away, 0)
	calcjd = JD - TZ/24
	st := Limit360(ApparentSiderealTime(calcjd)*15 + Lon)
	H := Limit360(st - nra)
	tmp2 := Sin(Lat)*Sin(ndec) + Cos(ndec)*Cos(Lat)*Cos(H)
	return ArcSin(tmp2)
}

func HMoonAngle(JD, Lon, Lat, TZ float64) float64 {
	//tmp := (TZ*15 - Lon) * 4 / 60
	//truejd=JD-tmp/24;
	calcjd := TD2UT(JD-TZ/24, true)
	ra := HMoonTrueRa(calcjd)
	dec := HMoonTrueDec(calcjd)
	away := HMoonAway(calcjd) / 149597870.7
	ndec := ZhanXinDec(ra, dec, Lat, Lon, JD-TZ/24, away, 0)
	nra := ZhanXinRa(ra, dec, Lat, Lon, JD-TZ/24, away, 0)
	calcjd = JD - TZ/24
	st := Limit360(ApparentSiderealTime(calcjd)*15 + Lon)
	H := Limit360(st - nra)
	tmp2 := Sin(H) / (Cos(H)*Sin(Lat) - Tan(ndec)*Cos(Lat))
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
func HMoonHeight(JD, Lon, Lat, TZ float64) float64 {
	//	tmp := (TZ*15 - Lon) * 4 / 60
	//truejd=JD-tmp/24;
	calcjd := TD2UT(JD-TZ/24, true)
	ra, dec := HMoonTrueRaDec(calcjd)
	away := HMoonAway(calcjd) / 149597870.7
	nra, ndec := ZhanXinRaDec(ra, dec, Lat, Lon, calcjd, away, 0)
	calcjd = JD - TZ/24
	st := Limit360(ApparentSiderealTime(calcjd)*15 + Lon)
	H := Limit360(st - nra)
	tmp2 := Sin(Lat)*Sin(ndec) + Cos(ndec)*Cos(Lat)*Cos(H)
	return ArcSin(tmp2)
}

// 废弃
func GetMoonTZTime(JD, Lon, Lat, TZ float64) float64 { //实际中天时间{
	JD = math.Floor(JD) + 0.5
	ttm := MoonTimeAngle(JD, Lon, Lat, TZ)
	if ttm > 0 && ttm < 180 {
		JD += 0.5
	}
	JD1 := JD
	for {
		JD0 := JD1
		stDegree := MoonTimeAngle(JD0, Lon, Lat, TZ) - 359.599
		stDegreep := (MoonTimeAngle(JD0+0.000005, Lon, Lat, TZ) - MoonTimeAngle(JD0-0.000005, Lon, Lat, TZ)) / 0.00001
		JD1 = JD0 - stDegree/stDegreep
		if math.Abs(JD1-JD0) <= 0.00001 {
			break
		}
	}
	return JD1
}

func MoonCulminationTime(jde, lon, lat, timezone float64) float64 {
	//jde 世界时，非力学时，当地时区 0时，无需转换力学时
	//ra,dec 瞬时天球座标，非J2000等时间天球坐标
	jde = math.Floor(jde) + 0.5
	JD1 := jde + Limit360(360-MoonTimeAngle(jde, lon, lat, timezone))/15.0/24.0/0.9
	limitHA := func(jde, lon, timezone float64) float64 {
		ha := MoonTimeAngle(jde, lon, lat, timezone)
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

func MoonTimeAngle(JD, Lon, Lat, TZ float64) float64 {
	startime := Limit360(ApparentSiderealTime(JD-TZ/24)*15 + Lon)
	timeangle := startime - HMoonApparentRa(JD, Lon, Lat, TZ)
	if timeangle < 0 {
		timeangle += 360
	}
	return timeangle
}

func GetMoonRiseTime(julianDay, longitude, latitude, timeZone, zenithShift, height float64) float64 {
	originalTimeZone := timeZone
	timeZone = longitude / 15
	var moonAngle, timeToMeridian float64 = 0, 0
	julianDayZero := math.Floor(julianDay) + 0.5
	//julianDay = math.Floor(julianDay) + 0.5 - originalTimeZone/24 + timeZone/24 // 求0时JDE
	//fix:这里时间分界线应当以传入的时区为准，不应当使用当地时区，否则在0时的判断会出错
	julianDay = math.Floor(julianDay) + 0.5
	estimatedTime := julianDay
	moonHeight := MoonHeight(julianDay, longitude, latitude, originalTimeZone) // 求此时月亮高度

	if zenithShift != 0 {
		moonAngle = -0.83333 // 修正大气折射
	}
	moonAngle = moonAngle - HeightDegreeByLat(height, latitude)

	moonAngleTime := MoonTimeAngle(julianDay, longitude, latitude, originalTimeZone)

	if moonHeight-moonAngle > 0 { // 月亮在地平线上或在落下与下中天之间
		if moonAngleTime > 180 {
			timeToMeridian = (180 + 360 - moonAngleTime) / 15
		} else {
			timeToMeridian = (180 - moonAngleTime) / 15
		}
		estimatedTime += (timeToMeridian/24 + (timeToMeridian/24*12.0)/15.0/24.0)
	}

	if moonHeight-moonAngle < 0 && moonAngleTime > 180 {
		timeToMeridian = (180 - moonAngleTime) / 15
		estimatedTime += (timeToMeridian/24 + (timeToMeridian/24*12.0)/15.0/24.0)
	} else if moonHeight-moonAngle < 0 && moonAngleTime < 180 {
		timeToMeridian = (180 - moonAngleTime) / 15
		estimatedTime += (timeToMeridian/24 + (timeToMeridian/24*12.0)/15.0/24.0)
	}

	currentAngle := MoonTimeAngle(estimatedTime, longitude, latitude, timeZone)
	if math.Abs(currentAngle-180) > 0.5 {
		estimatedTime += (180 - currentAngle) * 4.0 / 60.0 / 24.0
	}

	currentHeight := HMoonHeight(estimatedTime, longitude, latitude, timeZone)
	if !(currentHeight < -10 && math.Abs(latitude) < 60) {
		if currentHeight > moonAngle {
			return -1 // 拱
		}
		checkTime := estimatedTime + 12.0/24.0 + 6.0/15.0/24.0
		checkAngle := MoonTimeAngle(checkTime, longitude, latitude, timeZone)
		if checkAngle < 90 {
			checkAngle += 360
		}
		checkTime += (360 - checkAngle) * 4.0 / 60.0 / 24.0
		if HMoonHeight(checkTime, longitude, latitude, timeZone) < moonAngle {
			return -2 // 沉
		}
	}

	moonDeclination := MoonApparentDec(estimatedTime, longitude, latitude, timeZone)
	tmp := (Sin(moonAngle) - Sin(moonDeclination)*Sin(latitude)) / (Cos(moonDeclination) * Cos(latitude))

	if math.Abs(tmp) <= 1 && latitude < 85 {
		hourAngle := (180 - ArcCos(tmp)) / 15
		estimatedTime += hourAngle/24.00 + hourAngle/33.00/15.00
	} else {
		i := 0
		for MoonHeight(estimatedTime, longitude, latitude, timeZone) < moonAngle {
			i++
			estimatedTime += 15.0 / 60.0 / 24.0
			if i > 48 {
				break
			}
		}
	}

	// 使用牛顿迭代法求精确解
	estimatedTime = moonRiseSetNewtonRaphsonIteration(estimatedTime, longitude, latitude, timeZone, moonAngle, HMoonHeight, 0.00002)

	estimatedTime = estimatedTime - timeZone/24 + originalTimeZone/24

	if estimatedTime > julianDayZero+1 || estimatedTime < julianDayZero {
		return -3 // 明日
	} else {
		return estimatedTime
	}
}

func GetMoonSetTime(julianDay, longitude, latitude, timeZone, zenithShift, height float64) float64 {
	originalTimeZone := timeZone
	timeZone = longitude / 15
	var moonAngle, timeToMeridian float64 = 0, 0
	julianDayZero := math.Floor(julianDay) + 0.5
	//julianDay = math.Floor(julianDay) + 0.5 - originalTimeZone/24 + timeZone/24 // 求0时JDE
	//fix:这里时间分界线应当以传入的时区为准，不应当使用当地时区，否则在0时的判断会出错
	julianDay = math.Floor(julianDay) + 0.5
	estimatedTime := julianDay
	moonHeight := MoonHeight(julianDay, longitude, latitude, originalTimeZone) // 求此时月亮高度

	if zenithShift != 0 {
		moonAngle = -0.83333 // 修正大气折射
	}
	moonAngle = moonAngle - HeightDegreeByLat(height, latitude)

	moonAngleTime := MoonTimeAngle(julianDay, longitude, latitude, originalTimeZone)

	if moonHeight-moonAngle < 0 {
		timeToMeridian = (360 - moonAngleTime) / 15
		estimatedTime += (timeToMeridian/24 + (timeToMeridian/24.0*12.0)/15.0/24.0)
	}

	// 月亮在地平线上或在落下与下中天之间
	if moonHeight-moonAngle > 0 && moonAngleTime < 180 {
		timeToMeridian = (-moonAngleTime) / 15
		estimatedTime += (timeToMeridian/24.0 + (timeToMeridian/24.0*12.0)/15.0/24.0)
	} else if moonHeight-moonAngle > 0 {
		timeToMeridian = (360 - moonAngleTime) / 15
		estimatedTime += (timeToMeridian/24.0 + (timeToMeridian/24.0*12.0)/15.0/24.0)
	}

	currentAngle := MoonTimeAngle(estimatedTime, longitude, latitude, timeZone)
	if currentAngle < 180 {
		currentAngle += 360
	}
	if math.Abs(currentAngle-360) > 0.5 {
		estimatedTime += (360 - currentAngle) * 4.0 / 60.0 / 24.0
	}

	// estimatedTime = 月球中天时间
	currentHeight := HMoonHeight(estimatedTime, longitude, latitude, timeZone)
	if !(currentHeight > 10 && math.Abs(latitude) < 60) {
		if currentHeight < moonAngle {
			return -2 // 沉
		}
		checkTime := estimatedTime + 12.0/24.0 + 6.0/15.0/24.0
		angleSubtraction := 180 - MoonTimeAngle(checkTime, longitude, latitude, timeZone)
		checkTime += angleSubtraction * 4.0 / 60.0 / 24.0
		if HMoonHeight(checkTime, longitude, latitude, timeZone) > moonAngle {
			return -1 // 拱
		}
	}

	moonDeclination := MoonApparentDec(estimatedTime, longitude, latitude, timeZone)
	tmp := (Sin(moonAngle) - Sin(moonDeclination)*Sin(latitude)) / (Cos(moonDeclination) * Cos(latitude))

	if math.Abs(tmp) <= 1 && latitude < 85 {
		hourAngle := (ArcCos(tmp)) / 15.0
		estimatedTime += hourAngle/24 + hourAngle/33.0/15.0
	} else {
		i := 0
		for MoonHeight(estimatedTime, longitude, latitude, timeZone) > moonAngle {
			i++
			estimatedTime += 15.0 / 60.0 / 24.0
			if i > 48 {
				break
			}
		}
	}

	// 使用牛顿迭代法求精确解
	estimatedTime = moonRiseSetNewtonRaphsonIteration(estimatedTime, longitude, latitude, timeZone, moonAngle, HMoonHeight, 0.00002)

	estimatedTime = estimatedTime - timeZone/24 + originalTimeZone/24

	if estimatedTime > julianDayZero+1 || estimatedTime < julianDayZero {
		return -3 // 明日
	} else {
		return estimatedTime
	}
}

// heightFunction 高度函数类型定义，用于牛顿迭代法
type heightFunction func(time, longitude, latitude, timeZone float64) float64

// moonRiseSetNewtonRaphsonIteration 牛顿-拉夫逊迭代法求解天体高度方程
func moonRiseSetNewtonRaphsonIteration(initialTime, longitude, latitude, timeZone, targetAngle float64,
	heightFunc heightFunction, tolerance float64) float64 {
	const derivativeStep = 0.000005

	currentTime := initialTime

	for {
		previousTime := currentTime

		// 计算函数值：f(t) = height(t) - targetAngle
		functionValue := heightFunc(previousTime, longitude, latitude, timeZone) - targetAngle

		// 计算导数：f'(t) ≈ (f(t+h) - f(t-h)) / (2h)
		derivative := (heightFunc(previousTime+derivativeStep, longitude, latitude, timeZone) -
			heightFunc(previousTime-derivativeStep, longitude, latitude, timeZone)) / (2 * derivativeStep)

		// 牛顿-拉夫逊公式：t_new = t_old - f(t) / f'(t)
		currentTime = previousTime - functionValue/derivative

		// 检查收敛
		if math.Abs(currentTime-previousTime) <= tolerance {
			break
		}
	}

	return currentTime
}

func GetMoonCir() [][][]float64 {
	XL1 := [][][]float64{{
		//ML0
		{
			22639.586, 0.78475822, 8328.691424623, 1.5229241, 25.0719, -0.123598, 4586.438, 0.1873974, 7214.06286536, -2.184756, -18.860, 0.08280, 2369.914, 2.5429520, 15542.75428998, -0.661832, 6.212, -0.04080, 769.026, 3.140313, 16657.38284925, 3.04585, 50.144, -0.2472, 666.418, 1.527671, 628.30195521, -0.02664, 0.062, -0.0054, 411.596, 4.826607, 16866.9323150, -1.28012, -1.07, -0.0059, 211.656, 4.115028, -1114.6285593, -3.70768, -43.93, 0.2064, 205.436, 0.230523, 6585.7609101, -2.15812, -18.92, 0.0882, 191.956, 4.898507, 23871.4457146, 0.86109, 31.28, -0.164, 164.729, 2.586078, 14914.4523348, -0.6352, 6.15, -0.035, 147.321, 5.45530, -7700.3894694, -1.5496, -25.01, 0.118, 124.988, 0.48608, 7771.3771450, -0.3309, 3.11, -0.020, 109.380, 3.88323, 8956.9933798, 1.4963, 25.13, -0.129, 55.177, 5.57033, -1324.1780250, 0.6183, 7.3, -0.035, 45.100, 0.89898, 25195.623740, 0.2428, 24.0, -0.129, 39.533, 3.81213, -8538.240890, 2.8030, 26.1, -0.118, 38.430, 4.30115, 22756.817155, -2.8466, -12.6, 0.042, 36.124, 5.49587, 24986.074274, 4.5688, 75.2, -0.371, 30.773, 1.94559, 14428.125731, -4.3695, -37.7, 0.166, 28.397, 3.28586, 7842.364821, -2.2114, -18.8, 0.077, 24.358, 5.64142, 16171.056245, -0.6885, 6.3, -0.046, 18.585, 4.41371, -557.314280, -1.8538, -22.0, 0.10, 17.954, 3.58454, 8399.679100, -0.3576, 3.2, -0.03, 14.530, 4.9416, 23243.143759, 0.888, 31.2, -0.16, 14.380, 0.9709, 32200.137139, 2.384, 56.4, -0.29, 14.251, 5.7641, -2.301200, 1.523, 25.1, -0.12, 13.899, 0.3735, 31085.508580, -1.324, 12.4, -0.08, 13.194, 1.7595, -9443.319984, -5.231, -69.0, 0.33, 9.679, 3.0997, -16029.080894, -3.072, -50.1, 0.24, 9.366, 0.3016, 24080.995180, -3.465, -19.9, 0.08, 8.606, 4.1582, -1742.930514, -3.681, -44.0, 0.21, 8.453, 2.8416, 16100.068570, 1.192, 28.2, -0.14, 8.050, 2.6292, 14286.150380, -0.609, 6.1, -0.03, 7.630, 6.2388, 17285.684804, 3.019, 50.2, -0.25, 7.447, 1.4845, 1256.603910, -0.053, 0.1, -0.01, 7.371, 0.2736, 5957.458955, -2.131, -19.0, 0.09, 7.063, 5.6715, 33.757047, -0.308, -3.6, 0.02, 6.383, 4.7843, 7004.513400, 2.141, 32.4, -0.16, 5.742, 2.6572, 32409.686605, -1.942, 5, -0.05, 4.374, 4.3443, 22128.51520, -2.820, -13, 0.05, 3.998, 3.2545, 33524.31516, 1.766, 49, -0.25, 3.210, 2.2443, 14985.44001, -2.516, -16, 0.06, 2.915, 1.7138, 24499.74767, 0.834, 31, -0.17, 2.732, 1.9887, 13799.82378, -4.343, -38, 0.17, 2.568, 5.4122, -7072.08751, -1.576, -25, 0.11, 2.521, 3.2427, 8470.66678, -2.238, -19, 0.07, 2.489, 4.0719, -486.32660, -3.734, -44, 0.20, 2.146, 5.6135, -1952.47998, 0.645, 7, -0.03, 1.978, 2.7291, 39414.20000, 0.199, 37, -0.21, 1.934, 1.5682, 33314.76570, 6.092, 100, -0.5, 1.871, 0.4166, 30457.20662, -1.297, 12, -0.1, 1.753, 2.0582, -8886.00570, -3.38, -47, 0.2, 1.437, 2.386, -695.87607, 0.59, 7, 0, 1.373, 3.026, -209.54947, 4.33, 51, -0.2, 1.262, 5.940, 16728.37052, 1.17, 28, -0.1, 1.224, 6.172, 6656.74859, -4.04, -41, 0.2, 1.187, 5.873, 6099.43431, -5.89, -63, 0.3, 1.177, 1.014, 31571.83518, 2.41, 56, -0.3, 1.162, 3.840, 9585.29534, 1.47, 25, -0.1, 1.143, 5.639, 8364.73984, -2.18, -19, 0.1, 1.078, 1.229, 70.98768, -1.88, -22, 0.1, 1.059, 3.326, 40528.82856, 3.91, 81, -0.4, 0.990, 5.013, 40738.37803, -0.42, 30, -0.2, 0.948, 5.687, -17772.01141, -6.75, -94, 0.5, 0.876, 0.298, -0.35232, 0, 0, 0, 0.822, 2.994, 393.02097, 0, 0, 0, 0.788, 1.836, 8326.39022, 3.05, 50, -0.2, 0.752, 4.985, 22614.84180, 0.91, 31, -0.2, 0.740, 2.875, 8330.99262, 0, 0, 0, 0.669, 0.744, -24357.77232, -4.60, -75, 0.4, 0.644, 1.314, 8393.12577, -2.18, -19, 0.1, 0.639, 5.888, 575.33849, 0, 0, 0, 0.635, 1.116, 23385.11911, -2.87, -13, 0, 0.584, 5.197, 24428.75999, 2.71, 53, -0.3, 0.583, 3.513, -9095.55517, 0.95, 4, 0, 0.572, 6.059, 29970.88002, -5.03, -32, 0.1, 0.565, 2.960, 0.32863, 1.52, 25, -0.1, 0.561, 4.001, -17981.56087, -2.43, -43, 0.2, 0.557, 0.529, 7143.07519, -0.30, 3, 0, 0.546, 2.311, 25614.37623, 4.54, 75, -0.4, 0.536, 4.229, 15752.30376, -4.99, -45, 0.2, 0.493, 3.316, -8294.9344, -1.83, -29, 0.1, 0.491, 1.744, 8362.4485, 1.21, 21, -0.1, 0.478, 1.803, -10071.6219, -5.20, -69, 0.3, 0.454, 0.857, 15333.2048, 3.66, 57, -0.3, 0.445, 2.071, 8311.7707, -2.18, -19, 0.1, 0.426, 0.345, 23452.6932, -3.44, -20, 0.1, 0.420, 4.941, 33733.8646, -2.56, -2, 0, 0.413, 1.642, 17495.2343, -1.31, -1, 0, 0.404, 1.458, 23314.1314, -0.99, 9, -0.1, 0.395, 2.132, 38299.5714, -3.51, -6, 0, 0.382, 2.700, 31781.3846, -1.92, 5, 0, 0.375, 4.827, 6376.2114, 2.17, 32, -0.2, 0.361, 3.867, 16833.1753, -0.97, 3, 0, 0.358, 5.044, 15056.4277, -4.40, -38, 0.2, 0.350, 5.157, -8257.7037, -3.40, -47, 0.2, 0.344, 4.233, 157.7344, 0, 0, 0, 0.340, 2.672, 13657.8484, -0.58, 6, 0, 0.329, 5.610, 41853.0066, 3.29, 74, -0.4, 0.325, 5.895, -39.8149, 0, 0, 0, 0.309, 4.387, 21500.2132, -2.79, -13, 0.1, 0.302, 1.278, 786.0419, 0, 0, 0, 0.302, 5.341, -24567.3218, -0.27, -24, 0.1, 0.301, 1.045, 5889.8848, -1.57, -12, 0, 0.294, 4.201, -2371.2325, -3.65, -44, 0.2, 0.293, 3.704, 21642.1886, -6.55, -57, 0.2, 0.290, 4.069, 32828.4391, 2.36, 56, -0.3, 0.289, 3.472, 31713.8105, -1.35, 12, -0.1, 0.285, 5.407, -33.7814, 0.31, 4, 0, 0.283, 5.998, -16.9207, -3.71, -44, 0.2, 0.283, 2.772, 38785.8980, 0.23, 37, -0.2, 0.274, 5.343, 15613.7420, -2.54, -16, 0.1, 0.263, 3.997, 25823.9257, 0.22, 24, -0.1, 0.254, 0.600, 24638.3095, -1.61, 2, 0, 0.253, 1.344, 6447.1991, 0.29, 10, -0.1, 0.250, 0.887, 141.9754, -3.76, -44, 0.2, 0.247, 0.317, 5329.1570, -2.10, -19, 0.1, 0.245, 0.141, 36.0484, -3.71, -44, 0.2, 0.231, 2.287, 14357.1381, -2.49, -16, 0.1, 0.227, 5.158, 2.6298, 0, 0, 0, 0.219, 5.085, 47742.8914, 1.72, 63, -0.3, 0.211, 2.145, 6638.7244, -2.18, -19, 0.1, 0.201, 4.415, 39623.7495, -4.13, -14, 0, 0.194, 2.091, 588.4927, 0, 0, 0, 0.193, 3.057, -15400.7789, -3.10, -50, 0, 0.186, 5.598, 16799.3582, -0.72, 6, 0, 0.185, 3.886, 1150.6770, 0, 0, 0, 0.183, 1.619, 7178.0144, 1.52, 25, 0, 0.181, 2.635, 8328.3391, 1.52, 25, 0, 0.181, 2.077, 8329.0437, 1.52, 25, 0, 0.179, 3.215, -9652.8694, -0.90, -18, 0, 0.176, 1.716, -8815.0180, -5.26, -69, 0, 0.175, 5.673, 550.7553, 0, 0, 0, 0.170, 2.060, 31295.0580, -5.6, -39, 0, 0.167, 1.239, 7211.7617, -0.7, 6, 0, 0.165, 4.499, 14967.4158, -0.7, 6, 0, 0.164, 3.595, 15540.4531, 0.9, 31, 0, 0.164, 4.237, 522.3694, 0, 0, 0, 0.163, 4.633, 15545.0555, -2.2, -19, 0, 0.161, 0.478, 6428.0209, -2.2, -19, 0, 0.158, 2.03, 13171.5218, -4.3, -38, 0, 0.157, 2.28, 7216.3641, -3.7, -44, 0, 0.154, 5.65, 7935.6705, 1.5, 25, 0, 0.152, 0.46, 29828.9047, -1.3, 12, 0, 0.151, 1.19, -0.7113, 0, 0, 0, 0.150, 1.42, 23942.4334, -1.0, 9, 0, 0.144, 2.75, 7753.3529, 1.5, 25, 0, 0.137, 2.08, 7213.7105, -2.2, -19, 0, 0.137, 1.44, 7214.4152, -2.2, -19, 0, 0.136, 4.46, -1185.6162, -1.8, -22, 0, 0.136, 3.03, 8000.1048, -2.2, -19, 0, 0.134, 2.83, 14756.7124, -0.7, 6, 0, 0.131, 5.05, 6821.0419, -2.2, -19, 0, 0.128, 5.99, -17214.6971, -4.9, -72, 0, 0.127, 5.35, 8721.7124, 1.5, 25, 0, 0.126, 4.49, 46628.2629, -2.0, 19, 0, 0.125, 5.94, 7149.6285, 1.5, 25, 0, 0.124, 1.09, 49067.0695, 1.1, 55, 0, 0.121, 2.88, 15471.7666, 1.2, 28, 0, 0.111, 3.92, 41643.4571, 7.6, 125, -1, 0.110, 1.96, 8904.0299, 1.5, 25, 0, 0.106, 3.30, -18.0489, -2.2, -19, 0, 0.105, 2.30, -4.9310, 1.5, 25, 0, 0.104, 2.22, -6.5590, -1.9, -22, 0, 0.101, 1.44, 1884.9059, -0.1, 0, 0, 0.100, 5.92, 5471.1324, -5.9, -63, 0, 0.099, 1.12, 15149.7333, -0.7, 6, 0, 0.096, 4.73, 15508.9972, -0.4, 10, 0, 0.095, 5.18, 7230.9835, 1.5, 25, 0, 0.093, 3.37, 39900.5266, 3.9, 81, 0, 0.092, 2.01, 25057.0619, 2.7, 53, 0, 0.092, 1.21, -79.6298, 0, 0, 0, 0.092, 1.65, -26310.2523, -4.0, -68, 0, 0.091, 1.01, 42062.5561, -1.0, 23, 0, 0.090, 6.10, 29342.5781, -5.0, -32, 0, 0.090, 4.43, 15542.4020, -0.7, 6, 0, 0.090, 3.80, 15543.1066, -0.7, 6, 0, 0.089, 4.15, 6063.3859, -2.2, -19, 0, 0.086, 4.03, 52.9691, 0, 0, 0, 0.085, 0.49, 47952.4409, -2.6, 11, 0, 0.085, 1.60, 7632.8154, 2.1, 32, 0, 0.084, 0.22, 14392.0773, -0.7, 6, 0, 0.083, 6.22, 6028.4466, -4.0, -41, 0, 0.083, 0.63, -7909.9389, 2.8, 26, 0, 0.083, 5.20, -77.5523, 0, 0, 0, 0.082, 2.74, 8786.1467, -2.2, -19, 0, 0.080, 2.43, 9166.5428, -2.8, -26, 0, 0.080, 3.70, -25405.1732, 4.1, 27, 0, 0.078, 5.68, 48857.5200, 5.4, 106, -1, 0.077, 1.85, 8315.5735, -2.2, -19, 0, 0.075, 5.46, -18191.1103, 1.9, 8, 0, 0.075, 1.41, -16238.6304, 1.3, 1, 0, 0.074, 5.06, 40110.0761, -0.4, 30, 0, 0.072, 2.10, 64.4343, -3.7, -44, 0, 0.071, 2.17, 37671.2695, -3.5, -6, 0, 0.069, 1.71, 16693.4313, -0.7, 6, 0, 0.069, 3.33, -26100.7028, -8.3, -119, 1, 0.068, 1.09, 8329.4028, 1.5, 25, 0, 0.068, 3.62, 8327.9801, 1.5, 25, 0, 0.068, 2.41, 16833.1509, -1.0, 3, 0, 0.067, 3.40, 24709.2971, -3.5, -20, 0, 0.067, 1.65, 8346.7156, -0.3, 3, 0, 0.066, 2.61, 22547.2677, 1.5, 39, 0, 0.066, 3.50, 15576.5113, -1.0, 3, 0, 0.065, 5.76, 33037.9886, -2.0, 5, 0, 0.065, 4.58, 8322.1325, -0.3, 3, 0, 0.065, 6.20, 17913.9868, 3.0, 50, 0, 0.065, 1.50, 22685.8295, -1.0, 9, 0, 0.065, 2.37, 7180.3058, -1.9, -15, 0, 0.064, 1.06, 30943.5332, 2.4, 56, 0, 0.064, 1.89, 8288.8765, 1.5, 25, 0, 0.064, 4.70, 6.0335, 0.3, 4, 0, 0.063, 2.83, 8368.5063, 1.5, 25, 0, 0.063, 5.66, -2580.7819, 0.7, 7, 0, 0.062, 3.78, 7056.3285, -2.2, -19, 0, 0.061, 1.49, 8294.9100, 1.8, 29, 0, 0.061, 0.12, -10281.1714, -0.9, -18, 0, 0.061, 3.06, -8362.4729, -1.2, -21, 0, 0.061, 4.43, 8170.9571, 1.5, 25, 0, 0.059, 5.78, -13.1179, -3.7, -44, 0, 0.059, 5.97, 6625.5702, -2.2, -19, 0, 0.058, 5.01, -0.5080, -0.3, 0, 0, 0.058, 2.73, 7161.0938, -2.2, -19, 0, 0.057, 0.19, 7214.0629, -2.2, -19, 0, 0.057, 4.00, 22199.5029, -4.7, -35, 0, 0.057, 5.38, 8119.1420, 5.8, 76, 0, 0.056, 1.07, 7542.6495, 1.5, 25, 0, 0.056, 0.28, 8486.4258, 1.5, 25, 0, 0.054, 4.19, 16655.0816, 4.6, 75, 0, 0.053, 0.72, 7267.0320, -2.2, -19, 0, 0.053, 3.12, 12.6192, 0.6, 7, 0, 0.052, 2.99, -32896.013, -1.8, -49, 0, 0.052, 3.46, 1097.708, 0, 0, 0, 0.051, 5.37, -6443.786, -1.6, -25, 0, 0.051, 1.35, 7789.401, -2.2, -19, 0, 0.051, 5.83, 40042.502, 0.2, 38, 0, 0.051, 3.63, 9114.733, 1.5, 25, 0, 0.050, 1.51, 8504.484, -2.5, -22, 0, 0.050, 5.23, 16659.684, 1.5, 25, 0, 0.050, 1.15, 7247.820, -2.5, -23, 0, 0.047, 0.25, -1290.421, 0.3, 0, 0, 0.047, 4.67, -32686.464, -6.1, -100, 0, 0.047, 3.49, 548.678, 0, 0, 0, 0.047, 2.37, 6663.308, -2.2, -19, 0, 0.046, 0.98, 1572.084, 0, 0, 0, 0.046, 2.04, 14954.262, -0.7, 6, 0, 0.046, 3.72, 6691.693, -2.2, -19, 0, 0.045, 6.19, -235.287, 0, 0, 0, 0.044, 2.96, 32967.001, -0.1, 27, 0, 0.044, 3.82, -1671.943, -5.6, -66, 0, 0.043, 5.82, 1179.063, 0, 0, 0, 0.043, 0.07, 34152.617, 1.7, 49, 0, 0.043, 3.71, 6514.773, -0.3, 0, 0, 0.043, 5.62, 15.732, -2.5, -23, 0, 0.043, 5.80, 8351.233, -2.2, -19, 0, 0.042, 0.27, 7740.199, 1.5, 25, 0, 0.042, 6.14, 15385.020, -0.7, 6, 0, 0.042, 6.13, 7285.051, -4.1, -41, 0, 0.041, 1.27, 32757.451, 4.2, 78, 0, 0.041, 4.46, 8275.722, 1.5, 25, 0, 0.040, 0.23, 8381.661, 1.5, 25, 0, 0.040, 5.87, -766.864, 2.5, 29, 0, 0.040, 1.66, 254.431, 0, 0, 0, 0.040, 0.40, 9027.981, -0.4, 0, 0, 0.040, 2.96, 7777.936, 1.5, 25, 0, 0.039, 4.67, 33943.068, 6.1, 100, 0, 0.039, 3.52, 8326.062, 1.5, 25, 0, 0.039, 3.75, 21013.887, -6.5, -57, 0, 0.039, 5.60, 606.978, 0, 0, 0, 0.039, 1.19, 8331.321, 1.5, 25, 0, 0.039, 2.84, 7211.433, -2.2, -19, 0, 0.038, 0.67, 7216.693, -2.2, -19, 0, 0.038, 6.22, 25161.867, 0.6, 28, 0, 0.038, 4.40, 7806.322, 1.5, 25, 0, 0.038, 4.16, 9179.168, -2.2, -19, 0, 0.037, 4.73, 14991.999, -0.7, 6, 0, 0.036, 0.35, 67.514, -0.6, -7, 0, 0.036, 3.70, 25266.611, -1.6, 0, 0, 0.036, 5.39, 16328.796, -0.7, 6, 0, 0.035, 1.44, 7174.248, -2.2, -19, 0, 0.035, 5.00, 15684.730, -4.4, -38, 0, 0.035, 0.39, -15.419, -2.2, -19, 0, 0.035, 6.07, 15020.385, -0.7, 6, 0, 0.034, 6.01, 7371.797, -2.2, -19, 0, 0.034, 0.96, -16623.626, -3.4, -54, 0, 0.033, 6.24, 9479.368, 1.5, 25, 0, 0.033, 3.21, 23661.896, 5.2, 82, 0, 0.033, 4.06, 8311.418, -2.2, -19, 0, 0.033, 2.40, 1965.105, 0, 0, 0, 0.033, 5.17, 15489.785, -0.7, 6, 0, 0.033, 5.03, 21986.540, 0.9, 31, 0, 0.033, 4.10, 16691.140, 2.7, 46, 0, 0.033, 5.13, 47114.589, 1.7, 63, 0, 0.033, 4.45, 8917.184, 1.5, 25, 0, 0.033, 4.23, 2.078, 0, 0, 0, 0.032, 2.33, 75.251, 1.5, 25, 0, 0.032, 2.10, 7253.878, -2.2, -19, 0, 0.032, 3.11, -0.224, 1.5, 25, 0, 0.032, 4.43, 16640.462, -0.7, 6, 0, 0.032, 5.68, 8328.363, 0, 0, 0, 0.031, 5.32, 8329.020, 3.0, 50, 0, 0.031, 3.70, 16118.093, -0.7, 6, 0, 0.030, 3.67, 16721.817, -0.7, 6, 0, 0.030, 5.27, -1881.492, -1.2, -15, 0, 0.030, 5.72, 8157.839, -2.2, -19, 0, 0.029, 5.73, -18400.313, -6.7, -94, 0, 0.029, 2.76, 16.000, -2.2, -19, 0, 0.029, 1.75, 8879.447, 1.5, 25, 0, 0.029, 0.32, 8851.061, 1.5, 25, 0, 0.029, 0.90, 14704.903, 3.7, 57, 0, 0.028, 2.90, 15595.723, -0.7, 6, 0, 0.028, 5.88, 16864.631, 0.2, 24, 0, 0.028, 0.63, 16869.234, -2.8, -26, 0, 0.028, 4.04, -18609.863, -2.4, -43, 0, 0.027, 5.83, 6727.736, -5.9, -63, 0, 0.027, 6.12, 418.752, 4.3, 51, 0, 0.027, 0.14, 41157.131, 3.9, 81, 0, 0.026, 3.80, 15.542, 0, 0, 0, 0.026, 1.68, 50181.698, 4.8, 99, -1, 0.026, 0.32, 315.469, 0, 0, 0, 0.025, 5.67, 19.188, 0.3, 0, 0, 0.025, 3.16, 62.133, -2.2, -19, 0, 0.025, 3.76, 15502.939, -0.7, 6, 0, 0.025, 4.53, 45999.961, -2.0, 19, 0, 0.024, 3.21, 837.851, -4.4, -51, 0, 0.024, 2.82, 38157.596, 0.3, 37, 0, 0.024, 5.21, 15540.124, -0.7, 6, 0, 0.024, 0.26, 14218.576, 0, 13, 0, 0.024, 3.01, 15545.384, -0.7, 6, 0, 0.024, 1.16, -17424.247, -0.6, -21, 0, 0.023, 2.34, -67.574, 0.6, 7, 0, 0.023, 2.44, 18.024, -1.9, -22, 0, 0.023, 3.70, 469.400, 0, 0, 0, 0.023, 0.72, 7136.511, -2.2, -19, 0, 0.023, 4.50, 15582.569, -0.7, 6, 0, 0.023, 2.80, -16586.395, -4.9, -72, 0, 0.023, 1.51, 80.182, 0, 0, 0, 0.023, 1.09, 5261.583, -1.5, -12, 0, 0.023, 0.56, 54956.954, -0.5, 44, 0, 0.023, 4.01, 8550.860, -2.2, -19, 0, 0.023, 4.46, 38995.448, -4.1, -14, 0, 0.023, 3.82, 2358.126, 0, 0, 0, 0.022, 3.77, 32271.125, 0.5, 34, 0, 0.022, 0.82, 15935.775, -0.7, 6, 0, 0.022, 1.07, 24013.421, -2.9, -13, 0, 0.022, 0.40, 8940.078, -2.2, -19, 0, 0.022, 2.06, 15700.489, -0.7, 6, 0, 0.022, 4.27, 15124.002, -5.0, -45, 0, 0.021, 1.16, 56071.583, 3.2, 88, 0, 0.021, 5.58, 9572.189, -2.2, -19, 0, 0.020, 1.70, -17.273, -3.7, -44, 0, 0.020, 3.05, 214.617, 0, 0, 0, 0.020, 4.41, 8391.048, -2.2, -19, 0, 0.020, 5.95, 23869.145, 2.4, 56, 0, 0.020, 0.42, 40947.927, -4.7, -21, 0, 0.019, 1.39, 5818.897, 0.3, 10, 0, 0.019, 0.71, 23873.747, -0.7, 6, 0, 0.019, 2.81, 7291.615, -2.2, -19, 0, 0.019, 5.09, 8428.018, -2.2, -19, 0, 0.019, 4.14, 6518.187, -1.6, -12, 0, 0.019, 3.85, 21.330, 0, 0, 0, 0.018, 0.66, 14445.046, -0.7, 6, 0, 0.018, 1.65, 0.966, -4.0, -48, 0, 0.018, 5.64, -17143.709, -6.8, -94, 0, 0.018, 6.01, 7736.432, -2.2, -19, 0, 0.018, 2.74, 31153.083, -1.9, 5, 0, 0.018, 4.58, 6116.355, -2.2, -19, 0, 0.018, 2.28, 46.401, 0.3, 0, 0, 0.018, 3.80, 10213.597, 1.4, 25, 0, 0.018, 2.84, 56281.132, -1.1, 36, 0, 0.018, 3.53, 8249.062, 1.5, 25, 0, 0.017, 4.43, 20871.911, -3, -13, 0, 0.017, 4.44, 627.596, 0, 0, 0, 0.017, 1.85, 628.308, 0, 0, 0, 0.017, 1.19, 8408.321, 2, 25, 0, 0.017, 1.95, 7214.056, -2, -19, 0, 0.017, 1.57, 7214.070, -2, -19, 0, 0.017, 1.65, 13870.811, -6, -60, 0, 0.017, 0.30, 22.542, -4, -44, 0, 0.017, 2.62, -119.445, 0, 0, 0, 0.016, 4.87, 5747.909, 2, 32, 0, 0.016, 4.45, 14339.108, -1, 6, 0, 0.016, 1.83, 41366.680, 0, 30, 0, 0.016, 4.53, 16309.618, -3, -23, 0, 0.016, 2.54, 15542.754, -1, 6, 0, 0.016, 6.05, 1203.646, 0, 0, 0, 0.015, 5.2, 2751.147, 0, 0, 0, 0.015, 1.8, -10699.924, -5, -69, 0, 0.015, 0.4, 22824.391, -3, -20, 0, 0.015, 2.1, 30666.756, -6, -39, 0, 0.015, 2.1, 6010.417, -2, -19, 0, 0.015, 0.7, -23729.470, -5, -75, 0, 0.015, 1.4, 14363.691, -1, 6, 0, 0.015, 5.8, 16900.689, -2, 0, 0, 0.015, 5.2, 23800.458, 3, 53, 0, 0.015, 5.3, 6035.000, -2, -19, 0, 0.015, 1.2, 8251.139, 2, 25, 0, 0.015, 3.6, -8.860, 0, 0, 0, 0.015, 0.8, 882.739, 0, 0, 0, 0.015, 3.0, 1021.329, 0, 0, 0, 0.015, 0.6, 23296.107, 1, 31, 0, 0.014, 5.4, 7227.181, 2, 25, 0, 0.014, 0.1, 7213.352, -2, -19, 0, 0.014, 4.0, 15506.706, 3, 50, 0, 0.014, 3.4, 7214.774, -2, -19, 0, 0.014, 4.6, 6665.385, -2, -19, 0, 0.014, 0.1, -8.636, -2, -22, 0, 0.014, 3.1, 15465.202, -1, 6, 0, 0.014, 4.9, 508.863, 0, 0, 0, 0.014, 3.5, 8406.244, 2, 25, 0, 0.014, 1.3, 13313.497, -8, -82, 0, 0.014, 2.8, 49276.619, -3, 0, 0, 0.014, 0.1, 30528.194, -3, -10, 0, 0.013, 1.7, 25128.050, 1, 31, 0, 0.013, 2.9, 14128.405, -1, 6, 0, 0.013, 3.4, 57395.761, 3, 80, 0, 0.013, 2.7, 13029.546, -1, 6, 0, 0.013, 3.9, 7802.556, -2, -19, 0, 0.013, 1.6, 8258.802, -2, -19, 0, 0.013, 2.2, 8417.709, -2, -19, 0, 0.013, 0.7, 9965.210, -2, -19, 0, 0.013, 3.4, 50391.247, 0, 48, 0, 0.013, 3.0, 7134.433, -2, -19, 0, 0.013, 2.9, 30599.182, -5, -31, 0, 0.013, 3.6, -9723.857, 1, 0, 0, 0.013, 4.8, 7607.084, -2, -19, 0, 0.012, 0.8, 23837.689, 1, 35, 0, 0.012, 3.6, 4.409, -4, -44, 0, 0.012, 5.0, 16657.031, 3, 50, 0, 0.012, 4.4, 16657.735, 3, 50, 0, 0.012, 1.1, 15578.803, -4, -38, 0, 0.012, 6.0, -11.490, 0, 0, 0, 0.012, 1.9, 8164.398, 0, 0, 0, 0.012, 2.4, 31852.372, -4, -17, 0, 0.012, 2.4, 6607.085, -2, -19, 0, 0.012, 4.2, 8359.870, 0, 0, 0, 0.012, 0.5, 5799.713, -2, -19, 0, 0.012, 2.7, 7220.622, 0, 0, 0, 0.012, 4.3, -139.720, 0, 0, 0, 0.012, 2.3, 13728.836, -2, -16, 0, 0.011, 3.6, 14912.146, 1, 31, 0, 0.011, 4.7, 14916.748, -2, -19, 0},
		//ML1
		{
			1.67680, 4.66926, 628.301955, -0.0266, 0.1, -0.005, 0.51642, 3.3721, 6585.760910, -2.158, -18.9, 0.09, 0.41383, 5.7277, 14914.452335, -0.635, 6.2, -0.04, 0.37115, 3.9695, 7700.389469, 1.550, 25.0, -0.12, 0.27560, 0.7416, 8956.993380, 1.496, 25.1, -0.13, 0.24599, 4.2253, -2.301200, 1.523, 25.1, -0.12, 0.07118, 0.1443, 7842.36482, -2.211, -19, 0.08, 0.06128, 2.4998, 16171.05625, -0.688, 6, 0, 0.04516, 0.443, 8399.67910, -0.36, 3, 0, 0.04048, 5.771, 14286.15038, -0.61, 6, 0, 0.03747, 4.626, 1256.60391, -0.05, 0, 0, 0.03707, 3.415, 5957.45895, -2.13, -19, 0.1, 0.03649, 1.800, 23243.14376, 0.89, 31, -0.2, 0.02438, 0.042, 16029.08089, 3.07, 50, -0.2, 0.02165, 1.017, -1742.93051, -3.68, -44, 0.2, 0.01923, 3.097, 17285.68480, 3.02, 50, -0.3, 0.01692, 1.280, 0.3286, 1.52, 25, -0.1, 0.01361, 0.298, 8326.3902, 3.05, 50, -0.2, 0.01293, 4.013, 7072.0875, 1.58, 25, -0.1, 0.01276, 4.413, 8330.9926, 0, 0, 0, 0.01270, 0.101, 8470.6668, -2.24, -19, 0.1, 0.01097, 1.203, 22128.5152, -2.82, -13, 0, 0.01088, 2.545, 15542.7543, -0.66, 6, 0, 0.00835, 0.190, 7214.0629, -2.18, -19, 0.1, 0.00734, 4.855, 24499.7477, 0.83, 31, -0.2, 0.00686, 5.130, 13799.8238, -4.34, -38, 0.2, 0.00631, 0.930, -486.3266, -3.73, -44, 0, 0.00585, 0.699, 9585.2953, 1.5, 25, 0, 0.00566, 4.073, 8328.3391, 1.5, 25, 0, 0.00566, 0.638, 8329.0437, 1.5, 25, 0, 0.00539, 2.472, -1952.4800, 0.6, 7, 0, 0.00509, 2.88, -0.7113, 0, 0, 0, 0.00469, 3.56, 30457.2066, -1.3, 12, 0, 0.00387, 0.78, -0.3523, 0, 0, 0, 0.00378, 1.84, 22614.8418, 0.9, 31, 0, 0.00362, 5.53, -695.8761, 0.6, 7, 0, 0.00317, 2.80, 16728.3705, 1.2, 28, 0, 0.00303, 6.07, 157.7344, 0, 0, 0, 0.00300, 2.53, 33.7570, -0.3, -4, 0, 0.00295, 4.16, 31571.8352, 2.4, 56, 0, 0.00289, 5.98, 7211.7617, -0.7, 6, 0, 0.00285, 2.06, 15540.4531, 0.9, 31, 0, 0.00283, 2.65, 2.6298, 0, 0, 0, 0.00282, 6.17, 15545.0555, -2.2, -19, 0, 0.00278, 1.23, -39.8149, 0, 0, 0, 0.00272, 3.82, 7216.3641, -3.7, -44, 0, 0.00270, 4.37, 70.9877, -1.9, -22, 0, 0.00256, 5.81, 13657.8484, -0.6, 6, 0, 0.00244, 5.64, -0.2237, 1.5, 25, 0, 0.00240, 2.96, 8311.7707, -2.2, -19, 0, 0.00239, 0.87, -33.7814, 0.3, 4, 0, 0.00216, 2.31, 15.9995, -2.2, -19, 0, 0.00186, 3.46, 5329.1570, -2.1, -19, 0, 0.00169, 2.40, 24357.772, 4.6, 75, 0, 0.00161, 5.80, 8329.403, 1.5, 25, 0, 0.00161, 5.20, 8327.980, 1.5, 25, 0, 0.00160, 4.26, 23385.119, -2.9, -13, 0, 0.00156, 1.26, 550.755, 0, 0, 0, 0.00155, 1.25, 21500.213, -2.8, -13, 0, 0.00152, 0.60, -16.921, -3.7, -44, 0, 0.00150, 2.71, -79.630, 0, 0, 0, 0.00150, 5.29, 15.542, 0, 0, 0, 0.00148, 1.06, -2371.232, -3.7, -44, 0, 0.00141, 0.77, 8328.691, 1.5, 25, 0, 0.00141, 3.67, 7143.075, -0.3, 0, 0, 0.00138, 5.45, 25614.376, 4.5, 75, 0, 0.00129, 4.90, 23871.446, 0.9, 31, 0, 0.00126, 4.03, 141.975, -3.8, -44, 0, 0.00124, 6.01, 522.369, 0, 0, 0, 0.00120, 4.94, -10071.622, -5.2, -69, 0, 0.00118, 5.07, -15.419, -2.2, -19, 0, 0.00107, 3.49, 23452.693, -3.4, -20, 0, 0.00104, 4.78, 17495.234, -1.3, 0, 0, 0.00103, 1.44, -18.049, -2.2, -19, 0, 0.00102, 5.63, 15542.402, -0.7, 6, 0, 0.00102, 2.59, 15543.107, -0.7, 6, 0, 0.00100, 4.11, -6.559, -1.9, -22, 0, 0.00097, 0.08, 15400.779, 3.1, 50, 0, 0.00096, 5.84, 31781.385, -1.9, 5, 0, 0.00094, 1.08, 8328.363, 0, 0, 0, 0.00094, 2.46, 16799.358, -0.7, 6, 0, 0.00094, 1.69, 6376.211, 2.2, 32, 0, 0.00093, 3.64, 8329.020, 3.0, 50, 0, 0.00093, 2.65, 16655.082, 4.6, 75, 0, 0.00090, 1.90, 15056.428, -4.4, -38, 0, 0.00089, 1.59, 52.969, 0, 0, 0, 0.00088, 2.02, -8257.704, -3.4, -47, 0, 0.00088, 3.02, 7213.711, -2.2, -19, 0, 0.00087, 0.50, 7214.415, -2.2, -19, 0, 0.00087, 0.49, 16659.684, 1.5, 25, 0, 0.00082, 5.64, -4.931, 1.5, 25, 0, 0.00079, 5.17, 13171.522, -4.3, -38, 0, 0.00076, 3.60, 29828.905, -1.3, 12, 0, 0.00076, 4.08, 24567.322, 0.3, 24, 0, 0.00076, 4.58, 1884.906, -0.1, 0, 0, 0.00073, 0.33, 31713.811, -1.4, 12, 0, 0.00073, 0.93, 32828.439, 2.4, 56, 0, 0.00071, 5.91, 38785.898, 0.2, 37, 0, 0.00069, 2.20, 15613.742, -2.5, -16, 0, 0.00066, 3.87, 15.732, -2.5, -23, 0, 0.00066, 0.86, 25823.926, 0.2, 24, 0, 0.00065, 2.52, 8170.957, 1.5, 25, 0, 0.00063, 0.18, 8322.132, -0.3, 0, 0, 0.00060, 5.84, 8326.062, 1.5, 25, 0, 0.00060, 5.15, 8331.321, 1.5, 25, 0, 0.00060, 2.18, 8486.426, 1.5, 25, 0, 0.00058, 2.30, -1.731, -4, -44, 0, 0.00058, 5.43, 14357.138, -2, -16, 0, 0.00057, 3.09, 8294.910, 2, 29, 0, 0.00057, 4.67, -8362.473, -1, -21, 0, 0.00056, 4.15, 16833.151, -1, 0, 0, 0.00054, 1.93, 7056.329, -2, -19, 0, 0.00054, 5.27, 8315.574, -2, -19, 0, 0.00052, 5.6, 8311.418, -2, -19, 0, 0.00052, 2.7, -77.552, 0, 0, 0, 0.00051, 4.3, 7230.984, 2, 25, 0, 0.00050, 0.4, -0.508, 0, 0, 0, 0.00049, 5.4, 7211.433, -2, -19, 0, 0.00049, 4.4, 7216.693, -2, -19, 0, 0.00049, 4.3, 16864.631, 0, 24, 0, 0.00049, 2.2, 16869.234, -3, -26, 0, 0.00047, 6.1, 627.596, 0, 0, 0, 0.00047, 5.0, 12.619, 1, 7, 0, 0.00045, 4.9, -8815.018, -5, -69, 0, 0.00044, 1.6, 62.133, -2, -19, 0, 0.00042, 2.9, -13.118, -4, -44, 0, 0.00042, 4.1, -119.445, 0, 0, 0, 0.00041, 4.3, 22756.817, -3, -13, 0, 0.00041, 3.6, 8288.877, 2, 25, 0, 0.00040, 0.5, 6663.308, -2, -19, 0, 0.00040, 1.1, 8368.506, 2, 25, 0, 0.00039, 4.1, 6443.786, 2, 25, 0, 0.00039, 3.1, 16657.383, 3, 50, 0, 0.00038, 0.1, 16657.031, 3, 50, 0, 0.00038, 3.0, 16657.735, 3, 50, 0, 0.00038, 4.6, 23942.433, -1, 9, 0, 0.00037, 4.3, 15385.020, -1, 6, 0, 0.00037, 5.0, 548.678, 0, 0, 0, 0.00036, 1.8, 7213.352, -2, -19, 0, 0.00036, 1.7, 7214.774, -2, -19, 0, 0.00035, 1.1, 7777.936, 2, 25, 0, 0.00035, 1.6, -8.860, 0, 0, 0, 0.00035, 4.4, 23869.145, 2, 56, 0, 0.00035, 2.0, 6691.693, -2, -19, 0, 0.00034, 1.3, -1185.616, -2, -22, 0, 0.00034, 2.2, 23873.747, -1, 6, 0, 0.00033, 2.0, -235.287, 0, 0, 0, 0.00033, 3.1, 17913.987, 3, 50, 0, 0.00033, 1.0, 8351.233, -2, -19, 0},
		//ML2
		{0.004870, 4.6693, 628.30196, -0.027, 0, -0.01, 0.002280, 2.6746, -2.30120, 1.523, 25, -0.12, 0.001500, 3.372, 6585.76091, -2.16, -19, 0.1, 0.001200, 5.728, 14914.45233, -0.64, 6, 0, 0.001080, 3.969, 7700.38947, 1.55, 25, -0.1, 0.000800, 0.742, 8956.99338, 1.50, 25, -0.1, 0.000254, 6.002, 0.3286, 1.52, 25, -0.1, 0.000210, 0.144, 7842.3648, -2.21, -19, 0, 0.000180, 2.500, 16171.0562, -0.7, 6, 0, 0.000130, 0.44, 8399.6791, -0.4, 3, 0, 0.000126, 5.03, 8326.3902, 3.0, 50, 0, 0.000120, 5.77, 14286.1504, -0.6, 6, 0, 0.000118, 5.96, 8330.9926, 0, 0, 0, 0.000110, 1.80, 23243.1438, 0.9, 31, 0, 0.000110, 3.42, 5957.4590, -2.1, -19, 0, 0.000110, 4.63, 1256.6039, -0.1, 0, 0, 0.000099, 4.70, -0.7113, 0, 0, 0, 0.000070, 0.04, 16029.0809, 3.1, 50, 0, 0.000070, 5.14, 8328.3391, 1.5, 25, 0, 0.000070, 5.85, 8329.0437, 1.5, 25, 0, 0.000060, 1.02, -1742.9305, -3.7, -44, 0, 0.000060, 3.10, 17285.6848, 3.0, 50, 0, 0.000054, 5.69, -0.352, 0, 0, 0, 0.000043, 0.52, 15.542, 0, 0, 0, 0.000041, 2.03, 2.630, 0, 0, 0, 0.000040, 0.10, 8470.667, -2.2, -19, 0, 0.000040, 4.01, 7072.088, 1.6, 25, 0, 0.000036, 2.93, -8.860, -0.3, 0, 0, 0.000030, 1.20, 22128.515, -2.8, -13, 0, 0.000030, 2.54, 15542.754, -0.7, 6, 0, 0.000027, 4.43, 7211.762, -0.7, 6, 0, 0.000026, 0.51, 15540.453, 0.9, 31, 0, 0.000026, 1.44, 15545.055, -2.2, -19, 0, 0.000025, 5.37, 7216.364, -3.7, -44, 0},
		//ML3
		{0.00001200, 1.041, -2.3012, 1.52, 25, -0.1, 0.00000170, 0.31, -0.711, 0, 0, 0}},

		{ //精度1角秒
			//MB0
			{18461.240, 0.05710892, 8433.466157492, -0.6400617, -0.5345, -0.00294, 1010.167, 2.412663, 16762.15758211, 0.88286, 24.537, -0.1265, 999.694, 5.440038, -104.77473287, 2.16299, 25.606, -0.1207, 623.652, 0.915047, 7109.28813249, -0.02177, 6.746, -0.0379, 199.484, 1.815303, 15647.5290228, -2.82482, -19.39, 0.0799, 166.574, 4.842677, -1219.4032921, -1.5447, -18.33, 0.086, 117.261, 4.17086, 23976.2204475, -1.3019, 5.68, -0.044, 61.912, 4.76822, 25090.8490067, 2.4058, 49.61, -0.250, 33.357, 3.27060, 15437.979557, 1.5012, 31.8, -0.161, 31.760, 1.51241, 8223.916692, 3.6859, 50.7, -0.244, 29.577, 0.95817, 6480.986177, 0.0049, 6.7, -0.032, 15.566, 2.4871, -9548.094717, -3.068, -43.4, 0.21, 15.122, 0.2432, 32304.911872, 0.221, 30.7, -0.17, 12.094, 4.0135, 7737.590088, -0.048, 6.8, -0.04, 8.868, 1.8584, 15019.227068, -2.798, -19.5, 0.09, 8.045, 5.3812, 8399.709110, -0.332, 3.1, -0.02, 7.959, 4.2140, 23347.918492, -1.275, 5.6, -0.04, 7.435, 4.8858, -1847.705247, -1.518, -18.4, 0.09, 6.731, 3.8274, -16133.855627, -0.910, -24.5, 0.12, 6.580, 2.6732, 14323.350998, -2.207, -12.1, 0.04, 6.460, 3.1556, 9061.768113, -0.667, -0.5, -0.01, 6.296, 0.1713, 25300.398472, -1.920, -1.6, -0.01, 5.632, 0.8000, 733.076688, -2.190, -26, 0.12, 5.368, 2.1140, 16204.843302, -0.971, 3, -0.02, 5.311, 5.5111, 17390.459537, 0.856, 25, -0.13, 5.076, 2.2553, 523.52722, 2.136, 26, -0.13, 4.840, 6.1830, -7805.16420, 0.613, 1, 0, 4.806, 5.1414, -662.08901, 0.309, 4, -0.02, 3.984, 0.8406, 33419.54043, 3.929, 75, -0.37, 3.674, 5.0288, 22652.04242, -0.684, 13, -0.08, 2.998, 5.9291, 31190.28331, -3.487, -13, 0.04, 2.799, 2.1842, -16971.70705, 3.443, 27, -0.11, 2.414, 3.5735, 22861.59189, -5.010, -38, 0.16, 2.186, 3.9424, -9757.64418, 1.258, 8, -0.03, 2.146, 5.6262, 23766.67098, 3.024, 57, -0.29, 1.766, 3.3137, 14809.67760, 1.528, 32, -0.2, 1.624, 2.6013, 7318.83760, -4.35, -44, 0.2, 1.581, 3.8680, 16552.60812, 5.21, 76, -0.4, 1.520, 2.599, 40633.60330, 1.74, 56, -0.3, 1.516, 0.132, -17876.78614, -4.59, -68, 0.3, 1.510, 3.927, 8399.68473, -0.33, 3, 0, 1.318, 4.914, 16275.83098, -2.85, -19, 0.1, 1.264, 0.986, 24604.52240, -1.33, 6, 0, 1.192, 2.001, 39518.97474, -1.96, 12, -0.1, 1.135, 0.286, 31676.60992, 0.25, 31, -0.2, 1.086, 1.001, 5852.68422, 0.03, 7, 0, 1.019, 2.527, 33629.08990, -0.40, 23, -0.1, 0.823, 0.086, 16066.28151, 1.47, 32, -0.2, 0.804, 1.957, -33.78706, 0.28, 4, 0, 0.803, 5.212, 16833.14526, -1.00, 3, 0, 0.793, 1.472, -24462.54705, -2.43, -50, 0.2, 0.791, 1.658, -591.10134, -1.57, -18, 0.1, 0.667, 4.470, 24533.53473, 0.55, 28, -0.1, 0.650, 2.530, -10176.39667, -3.04, -43, 0.2, 0.639, 1.583, 25719.15096, 2.38, 50, -0.3, 0.634, 0.318, 5994.65957, -3.73, -37, 0.2, 0.631, 2.147, 8435.76736, -2.16, -26, 0.1, 0.630, 1.109, 8431.16496, 0.88, 25, -0.1, 0.596, 2.716, 13695.04904, -2.18, -12, 0.1, 0.589, 1.214, 7666.60241, 1.83, 29, -0.1, 0.473, 1.101, 30980.7338, 0.84, 38, -0.2, 0.456, 0.116, -71.0177, 1.85, 22, -0.1, 0.430, 2.786, -8990.7804, -1.21, -21, 0.1, 0.416, 1.454, 16728.4005, 1.19, 28, -0.1, 0.415, 5.072, 22023.7405, -0.66, 13, -0.1, 0.383, 4.257, 22719.6165, -1.25, 6, 0, 0.352, 2.972, 14880.6653, -0.35, 10, -0.1, 0.339, 5.972, 30561.9814, -3.46, -13, 0, 0.329, 1.587, -18086.3356, -0.26, -17, 0.1, 0.326, 1.016, 8467.2232, -0.95, -4, 0, 0.315, 1.902, 14390.9251, -2.77, -20, 0.1, 0.313, 4.611, 8852.2186, 3.66, 51, -0.2, 0.305, 0.616, 6551.9739, -1.88, -15, 0.1, 0.301, 4.728, -7595.6147, -3.71, -51, 0.2, 0.299, 1.874, 7143.0452, -0.33, 3, 0, 0.291, 3.156, -1428.9528, 2.78, 33, -0.2, 0.269, 4.929, -2476.0072, -1.49, -18, 0.1, 0.263, 3.196, 41748.2319, 5.45, 100, -0.5, 0.254, 3.387, -1009.8538, -5.87, -70, 0.3, 0.245, 1.930, 32514.4613, -4.10, -20, 0.1, 0.237, 3.342, 32933.2138, 0.19, 31, -0.2, 0.214, 3.617, 22233.2899, -4.98, -38, 0.2, 0.213, 4.357, 47847.6662, -0.44, 37, -0.2, 0.206, 3.872, 23418.9062, -3.16, -16, 0.1, 0.172, 5.772, 14951.6530, -2.2, -12, 0, 0.158, 2.04, 38890.6728, -1.9, 12, 0, 0.146, 1.70, 32095.3624, 4.5, 82, 0, 0.145, 4.29, 40843.1528, -2.6, 5, 0, 0.139, 2.90, 7876.1519, -2.5, -23, 0, 0.138, 4.95, 48962.2947, 3.3, 81, 0, 0.134, 3.97, 8365.8920, -0.1, 7, 0, 0.134, 4.06, -26205.4776, -6.1, -94, 0, 0.130, 1.40, -8643.0156, 5.0, 52, 0, 0.129, 5.67, 23138.3690, 3.1, 57, 0, 0.124, 2.64, 40005.3013, 1.8, 56, 0, 0.118, 4.88, 41957.7813, 1.1, 49, 0, 0.113, 3.78, -15505.5537, -0.9, -24, 0, 0.113, 4.87, 16904.1329, -2.9, -19, 0, 0.113, 1.84, 23280.3444, -0.7, 13, 0, 0.110, 0.43, -17319.4719, -2.7, -47, 0, 0.105, 1.61, 37.2006, -1.6, -18, 0, 0.102, 1.28, 25161.8367, 0.5, 28, 0, 0.095, 0.76, 1361.3786, -2.2, -25, 0, 0.094, 0.50, 29866.1053, -2.9, -6, 0, 0.092, 6.22, 24881.2995, 6.7, 101, 0, 0.088, 3.99, -10385.9461, 1.3, 8, 0, 0.085, 4.71, 70.9933, -1.9, -22, 0, 0.084, 0.86, 15613.7720, -2.5, -16, 0, 0.081, 4.43, 21537.4139, -4.4, -31, 0, 0.080, 1.86, -8365.9521, 0, -7, 0, 0.080, 0, 16728.3762, 1.2, 28, 0, 0.079, 2.44, -8919.7928, -3.1, -43, 0, 0.078, 3.69, -452.5395, -4.0, -48, 0, 0.075, 5.40, -32791.2385, -4.0, -75, 0, 0.073, 5.80, -1185.6462, -1.9, -22, 0, 0.070, 3.46, 16759.8564, 2.4, 50, 0, 0.069, 3.36, 14181.3756, 1.6, 32, 0, 0.068, 4.50, 16764.4588, -0.6, -1, 0, 0.067, 4.74, 8446.0854, 0, 7, 0, 0.066, 5.86, 24185.7699, -5.6, -46, 0, 0.064, 0.54, 32862.2262, 2.1, 53, 0, 0.063, 2.44, 24394.9729, 3.0, 57, 0, 0.063, 1.77, 5785.1101, 0.6, 14, 0, 0.062, 2.64, 6690.5356, -4.3, -45, 0, 0.062, 2.21, 1151.8292, 2.1, 26, 0, 0.062, 3.94, 34047.8424, 3.9, 75, 0, 0.060, 1.40, 38404.3462, -5.7, -32, 0, 0.058, 0.33, 31048.3080, 0.3, 31, 0, 0.057, 3.11, 9690.0701, -0.7, 0, 0, 0.057, 1.14, 30352.4319, 0.9, 38, 0, 0.056, 6.00, 8504.4538, -2.5, -22, 0, 0.055, 5.47, 18018.7615, 0.8, 25, 0, 0.055, 0.17, -18505.0881, -4.6, -69, 0, 0.055, 0.76, -9129.3422, 1.2, 8, 0, 0.054, 5.70, 7947.1396, -4.4, -44, 0, 0.053, 0.36, 5366.358, -3.7, -37, 0, 0.052, 4.01, -68.726, -1.5, -18, 0, 0.051, 2.74, 31818.585, -3.5, -13, 0, 0.051, 0.98, 16798.206, -2.8, -19, 0, 0.050, 5.94, 8293.747, -0.3, 0, 0, 0.049, 1.52, 15090.215, -4.7, -41, 0, 0.048, 3.46, 39309.425, 2.4, 63, 0, 0.046, 3.21, 23942.463, -1.0, 9, 0, 0.046, 3.35, 7143.070, -0.3, 0, 0, 0.042, 3.76, 46733.038, -4.1, -7, 0, 0.042, 5.18, 8288.351, 0, 7, 0, 0.040, 3.37, 16795.915, 0.6, 21, 0, 0.039, 4.54, -1776.718, -3.4, -40, 0, 0.039, 0.04, 8439.500, -0.3, 0, 0, 0.037, 3.91, 8479.867, -0.3, 0, 0, 0.037, 2.86, 38194.797, -1.3, 19, 0, 0.036, 1.04, 5224.382, 0.1, 7, 0, 0.036, 3.57, 15995.294, 3.4, 54, 0, 0.036, 5.33, 23209.357, 1.2, 35, 0, 0.035, 1.02, 8452.654, -0.3, 0, 0, 0.035, 4.31, 8294.904, 1.8, 29, 0, 0.035, 2.76, 13066.747, -2.2, -12, 0, 0.034, 6.07, 15508.967, -0.4, 10, 0, 0.032, 1.89, -17529.021, 1.6, 0, 0, 0.031, 5.70, 41261.905, 1.7, 56, 0, 0.031, 5.33, 30075.655, -7.2, -57, 0, 0.031, 5.97, -40.340, -1.5, -18, 0, 0.030, 2.87, 6533.950, 0, 7, 0, 0.030, 0.36, 49171.844, -1.1, 30, 0, 0.030, 4.40, 47219.364, -0.4, 37, 0, 0.030, 0.39, 23489.894, -5.0, -38, 0, 0.029, 5.12, 21395.439, -0.6, 13, 0, 0.029, 2.62, 8715.153, -0.3, 0, 0, 0.029, 2.94, 16826.592, -2.8, -19, 0, 0.028, 6.23, 31747.598, -1.6, 9, 0, 0.028, 0.43, 56176.358, 1.1, 62, 0, 0.027, 2.32, 8792.706, -0.3, 0, 0, 0.026, 3.02, 14252.363, -0.3, 10, 0, 0.025, 5.10, 40147.277, -2.0, 12, 0, 0.025, 4.59, 8433.795, 0.9, 25, 0, 0.025, 4.95, 8433.138, -2.2, -26, 0, 0.025, 1.03, -9338.545, -7.4, -95, 0, 0.024, 0.68, 17180.910, 5.2, 76, 0, 0.024, 3.81, 25057.092, 2.7, 53, 0, 0.024, 6.02, 29933.679, -3.4, -13, 0, 0.024, 2.37, -15924.306, -5.2, -76, 0, 0.024, 3.46, 8681.372, 0, 7, 0, 0.023, 2.80, 7108.936, 0, 7, 0, 0.023, 2.17, 7109.640, 0, 7, 0, 0.023, 2.57, -10804.699, -3.0, -44, 0, 0.022, 1.21, 6323.246, 0, 7, 0, 0.022, 3.22, 8259.965, 0, 7, 0, 0.022, 1.97, 7106.987, 1.5, 32, 0, 0.022, 3.00, 7111.589, -1.5, -18, 0, 0.022, 1.22, 14532.900, -6.5, -63, 0, 0.021, 0.66, 5923.672, -1.8, -15, 0, 0.021, 0.69, 24047.208, -3.2, -16, 0, 0.020, 5.51, -26415.027, -1.8, -42, 0, 0.020, 3.70, 16745.237, -2.8, -19, 0, 0.020, 1.26, 7038.300, 1.9, 29, 0, 0.020, 5.78, 6716.267, 0, 7, 0, 0.020, 3.76, 7895.330, 0, 7, 0, 0.019, 0.44, -121.695, -1.5, -18, 0, 0.018, 2.16, 15576.541, -0.9, 0, 0, 0.018, 0.09, -17248.484, -4.6, -68, 0, 0.018, 6.14, -7176.862, 0.6, 0, 0, 0.018, 5.55, 50076.923, 7.0, 125, -1, 0.017, 2.47, 8257.674, 3, 47, 0, 0.017, 4.20, 31609.036, 1, 38, 0, 0.017, 0.50, 175.762, -4, -48, 0, 0.017, 3.20, -2057.255, 3, 33, 0},
			//MB1
			{0.07430, 4.0998, 6480.98618, 0.005, 7, -0.03, 0.03043, 0.872, 7737.59009, -0.05, 7, 0, 0.02229, 5.000, 15019.22707, -2.80, -19, 0.1, 0.01999, 1.072, 23347.91849, -1.28, 6, 0, 0.01869, 1.744, -1847.70525, -1.52, -18, 0.1, 0.01696, 5.597, 16133.8556, 0.91, 24, -0.1, 0.01623, 0.014, 9061.7681, -0.67, 0, 0, 0.01419, 3.942, 733.0767, -2.19, -26, 0.1, 0.01338, 2.370, 17390.4595, 0.86, 25, -0.1, 0.01304, 5.633, 8399.6847, -0.33, 3, 0, 0.01279, 0.886, -523.5272, -2.14, -26, 0.1, 0.01215, 3.242, 7805.1642, -0.61, -1, 0, 0.01088, 3.686, 8435.7674, -2.16, -26, 0.1, 0.01088, 5.853, 8431.1650, 0.88, 25, -0.1, 0.00546, 4.143, 5852.6842, 0, 7, 0, 0.00443, 0.17, 14809.6776, 1.5, 32, 0, 0.00342, 2.24, 8399.7091, -0.3, 3, 0, 0.00330, 1.77, 16275.8310, -2.9, -19, 0, 0.00318, 4.13, 24604.5224, -1.3, 6, 0, 0.00296, 0.90, 7109.2881, 0, 7, 0, 0.00285, 3.43, 31676.6099, 0.2, 31, 0, 0.00207, 3.23, 16066.2815, 1.5, 32, 0, 0.00202, 2.07, 16833.1453, -1.0, 3, 0, 0.00202, 5.10, -33.7871, 0.3, 4, 0, 0.00200, 1.67, 24462.5471, 2.4, 50, 0, 0.00198, 4.80, -591.1013, -1.6, -18, 0, 0.00193, 1.12, 22719.6165, -1.2, 6, 0, 0.00164, 5.67, -10176.397, -3.0, -43, 0, 0.00161, 4.73, 25719.151, 2.4, 50, 0, 0.00158, 5.04, 14390.925, -2.8, -20, 0, 0.00149, 5.86, 13695.049, -2.2, -12, 0, 0.00135, 1.79, -2476.007, -1.5, -18, 0, 0.00121, 1.93, 16759.856, 2.4, 50, 0, 0.00117, 6.04, 16764.459, -0.6, 0, 0, 0.00104, 1.93, 22023.740, -0.7, 13, 0, 0.00085, 2.83, 30561.981, -3.5, -13, 0, 0.00079, 4.81, -8852.219, -3.7, -51, 0, 0.00076, 1.59, -7595.615, -3.7, -51, 0, 0.00075, 2.91, 8433.795, 0.9, 25, 0, 0.00075, 0.35, 8433.138, -2.2, -26, 0, 0.00073, 0.13, 70.993, -1.9, -22, 0, 0.00069, 1.70, 16728.376, 1.2, 28, 0, 0.00068, 0.83, 8365.892, -0.1, 7, 0, 0.00060, 0.20, 32933.214, 0.2, 31, 0, 0.00060, 0.31, 8446.085, 0, 7, 0},
			//MB2
			{0.000220, 4.100, 6480.9862, 0, 7, 0, 0.000101, 5.24, 8435.7674, -2.2, -26, 0, 0.000101, 4.30, 8431.1650, 0.9, 25, 0, 0.000090, 0.87, 7737.5901, 0, 7, 0, 0.000060, 1.07, 23347.9185, -1.3, 6, 0, 0.000060, 5.00, 15019.2271, -2.8, -19, 0, 0.000050, 1.74, -1847.705, -1.5, -18, 0, 0.000050, 5.60, 16133.856, 0.9, 24, 0, 0.000050, 0.01, 9061.768, -0.7, 0, 0, 0.000040, 3.24, 7805.164, -0.6, 0, 0, 0.000040, 3.94, 733.077, -2.2, -26, 0, 0.000040, 0.89, -523.527, -2.1, -26, 0}},

		{ //精度1千米
			//MR0
			{385000.510, 0, 0, 0, 0, 0, 20905.354, 5.4971472, 8328.691424623, 1.522924, 25.0719, -0.12360, 3699.111, 4.8997864, 7214.06286536, -2.184756, -18.860, 0.0828, 2955.967, 0.972156, 15542.75428998, -0.66183, 6.212, -0.0408, 569.925, 1.569516, 16657.3828492, 3.04585, 50.14, -0.2472, 246.158, 5.68582, -1114.6285593, -3.7077, -43.93, 0.206, 204.586, 1.01528, 14914.4523348, -0.6352, 6.15, -0.035, 170.733, 3.32771, 23871.4457146, 0.8611, 31.28, -0.164, 152.138, 4.94291, 6585.7609101, -2.1581, -18.92, 0.088, 129.620, 0.74291, -7700.3894694, -1.5496, -25.01, 0.118, 108.743, 5.19847, 7771.3771450, -0.3309, 3.1, -0.020, 104.755, 2.31243, 8956.993380, 1.4963, 25.1, -0.129, 79.661, 5.38293, -8538.240890, 2.8030, 26.1, -0.118, 48.888, 6.24006, 628.301955, -0.0266, 0.1, -0.005, 34.783, 2.73035, 22756.817155, -2.847, -12.6, 0.04, 30.824, 4.0706, 16171.056245, -0.688, 6.3, -0.05, 24.208, 1.7151, 7842.364821, -2.211, -18.8, 0.08, 23.210, 3.9251, 24986.074274, 4.569, 75.2, -0.37, 21.636, 0.3748, 14428.125731, -4.370, -37.7, 0.17, 16.675, 2.0137, 8399.679100, -0.358, 3.2, -0.03, 14.403, 3.3303, -9443.319984, -5.231, -69.0, 0.33, 12.831, 3.3708, 23243.143759, 0.888, 31.2, -0.16, 11.650, 5.0859, 31085.508580, -1.324, 12, -0.08, 10.445, 5.6833, 32200.13714, 2.384, 56, -0.29, 10.321, 0.8579, -1324.17803, 0.618, 7, -0.03, 10.056, 5.7290, -1742.93051, -3.681, -44, 0.21, 9.884, 1.0584, 14286.15038, -0.609, 6, -0.03, 8.752, 4.7856, -9652.86945, -0.905, -18, 0.09, 8.379, 5.9845, -557.31428, -1.854, -22, 0.10, 7.003, 4.6705, -16029.08089, -3.072, -50, 0.24, 6.322, 1.2708, 16100.06857, 1.192, 28, -0.14, 5.751, 4.6680, 17285.68480, 3.019, 50, -0.25, 4.950, 4.9860, 5957.45895, -2.131, -19, 0.09, 4.421, 4.5969, -209.54947, 4.326, 51, -0.24, 4.131, 3.2135, 7004.51340, 2.141, 32, -0.16, 3.958, 2.7735, 22128.51520, -2.820, -13, 0.05, 3.258, 0.6735, 14985.44001, -2.52, -16, 0.1, 3.148, 0.114, 16866.93231, -1.28, -1, 0, 2.616, 0.143, 24499.74767, 0.83, 31, -0.2, 2.354, 1.672, 8470.66678, -2.24, -19, 0.1, 2.117, 0.700, -7072.08751, -1.58, -25, 0.1, 1.897, 0.418, 13799.82378, -4.34, -38, 0.2, 1.739, 3.629, -8886.00570, -3.38, -47, 0.2, 1.571, 5.129, 30457.20662, -1.30, 12, -0.1, 1.423, 1.158, 39414.20000, 0.20, 37, -0.2, 1.419, 6.171, 23314.13143, -0.99, 9, -0.1, 1.166, 2.269, 9585.29534, 1.47, 25, -0.1, 1.117, 6.281, 33314.76570, 6.09, 100, -0.5, 1.066, 6.197, 1256.60391, -0.05, 0, 0, 1.059, 4.068, 8364.73984, -2.18, -19, 0.1, 0.933, 4.369, 16728.3705, 1.17, 28, -0.1, 0.862, 4.601, 6656.7486, -4.04, -41, 0.2, 0.851, 2.800, 70.9877, -1.88, -22, 0.1, 0.849, 5.726, 31571.8352, 2.41, 56, -0.3, 0.796, 5.084, -9095.5552, 0.95, 4, 0, 0.779, 0.975, -17772.0114, -6.75, -94, 0.5, 0.774, 2.658, 15752.3038, -4.99, -45, 0.2, 0.728, 0.266, 8326.3902, 3.05, 50, -0.2, 0.683, 1.304, 8330.9926, 0, 0, 0, 0.670, 1.756, 40528.8286, 3.91, 81, -0.4, 0.658, 3.414, 22614.8418, 0.91, 31, -0.2, 0.657, 0.901, -1952.4800, 0.64, 7, 0, 0.598, 6.026, 8393.1258, -2.18, -19, 0.1, 0.596, 5.014, 24080.9952, -3.46, -20, 0.1, 0.579, 5.829, 23385.1191, -2.87, -13, 0, 0.514, 4.302, 6099.4343, -5.89, -63, 0.3, 0.508, 1.830, 14218.5763, -0.04, 13, -0.1, 0.498, 5.242, 7143.0752, -0.30, 3, 0, 0.495, 3.373, -10071.6219, -5.20, -69, 0.3, 0.473, 2.430, -17981.5609, -2.43, -43, 0.2, 0.456, 4.887, -8294.9344, -1.83, -29, 0.1, 0.453, 0.173, 8362.4485, 1.21, 21, -0.1, 0.423, 4.489, 29970.8800, -5.03, -32, 0.1, 0.422, 2.315, -24357.7723, -4.60, -75, 0.4, 0.411, 1.102, 13657.8484, -0.58, 6, 0, 0.410, 0.500, 8311.7707, -2.18, -19, 0.1, 0.379, 3.626, 24428.7600, 2.71, 53, 0, 0.355, 0.740, 25614.3762, 4.54, 75, 0, 0.343, 5.772, -2371.2325, -3.7, -44, 0, 0.335, 0.857, 9166.5428, -2.8, -26, 0, 0.332, 0.444, -8257.7037, -3.4, -47, 0, 0.323, 4.829, -10281.1714, -0.9, -18, 0, 0.322, 5.758, 5889.8848, -1.6, -12, 0, 0.287, 0.56, 38299.5714, -3.5, -6, 0, 0.284, 5.57, 15333.2048, 3.7, 57, 0, 0.279, 2.82, 21500.2132, -2.8, -13, 0, 0.256, 0.72, 14357.1381, -2.5, -16, 0, 0.248, 2.20, -7909.9389, 2.8, 26, 0, 0.245, 1.90, 31713.8105, -1.4, 12, 0, 0.237, 3.47, 15056.4277, -4.4, -38, 0, 0.213, 3.77, 15613.7420, -2.5, -16, 0, 0.213, 2.50, 32828.4391, 2.4, 56, 0, 0.209, 3.26, 6376.2114, 2.2, 32, 0, 0.205, 2.93, 14967.4158, -0.7, 6, 0, 0.205, 2.02, 15540.4531, 0.9, 31, 0, 0.204, 3.06, 15545.0555, -2.2, -19, 0, 0.203, 1.20, 38785.8980, 0.2, 37, 0, 0.201, 6.06, 6447.1991, 0.3, 10, 0, 0.186, 6.13, -16238.6304, 1.3, 1, 0, 0.183, 2.13, 21642.1886, -6.6, -57, 0, 0.169, 3.29, -8815.0180, -5.3, -69, 0, 0.167, 1.06, 8328.3391, 1.5, 25, 0, 0.167, 0.51, 8329.0437, 1.5, 25, 0, 0.167, 1.26, 14756.7124, -0.7, 6, 0, 0.158, 0.07, 17495.2343, -1.3, -1, 0, 0.157, 0.57, 6638.7244, -2.2, -19, 0, 0.157, 6.21, 22685.8295, -1.0, 9, 0, 0.148, 5.03, 5329.1570, -2.1, -19, 0, 0.148, 4.03, 16799.3582, -0.7, 6, 0, 0.145, 0.05, 7178.0144, 1.5, 25, 0, 0.144, 5.64, -486.3266, -3.7, -44, 0, 0.139, 3.51, 47742.8914, 1.7, 63, 0, 0.138, 4.07, 7935.6705, 1.5, 25, 0, 0.136, 4.63, -15400.7789, -3.1, -50, 0, 0.136, 3.96, -695.8761, 0.6, 7, 0, 0.135, 5.95, 7211.7617, -0.7, 6, 0, 0.128, 5.17, 29828.9047, -1.3, 12, 0, 0.127, 1.18, 7753.3529, 1.5, 25, 0, 0.127, 0.71, 7216.3641, -3.7, -44, 0, 0.124, 5.83, 15149.7333, -0.7, 6, 0, 0.121, 1.46, 8000.1048, -2.2, -19, 0, 0.120, 3.78, 8721.7124, 1.5, 25, 0, 0.116, 5.19, 6428.0209, -2.2, -19, 0, 0.114, 2.89, -1185.6162, -1.8, -22, 0, 0.112, 2.85, 15542.4020, -0.7, 6, 0, 0.112, 2.23, 15543.1066, -0.7, 6, 0, 0.110, 0.51, 7213.7105, -2.2, -19, 0, 0.110, 6.15, 7214.4152, -2.2, -19, 0, 0.110, 1.31, 15471.7666, 1.2, 28, 0, 0.109, 2.46, 141.9754, -3.8, -44, 0, 0.108, 0.46, 13171.5218, -4.3, -38, 0, 0.108, 6.13, 23942.4334, -1.0, 9, 0, 0.107, 3.15, 15508.9972, -0.4, 10, 0, 0.105, 0.39, 8904.030, 1.5, 25, 0, 0.105, 4.93, 14392.077, -0.7, 6, 0, 0.103, 2.47, 25195.624, 0.2, 24, 0, 0.101, 3.48, 6821.042, -2.2, -19, 0, 0.099, 4.37, 7149.629, 1.5, 25, 0, 0.099, 1.27, -17214.697, -4.9, -72, 0, 0.096, 1.93, 15576.511, -1.0, 0, 0, 0.086, 2.92, 46628.263, -2.0, 19, 0, 0.085, 6.22, 8504.484, -2.5, -22, 0, 0.080, 3.40, -2438.807, -3.1, -37, 0, 0.080, 1.17, 8786.147, -2.2, -19, 0, 0.077, 3.61, 7230.984, 1.5, 25, 0, 0.071, 0.28, 8315.574, -2.2, -19, 0, 0.067, 4.53, 29342.578, -5.0, -32, 0, 0.065, 2.24, 31642.823, 0.5, 34, 0, 0.063, 5.80, 8329.403, 1.5, 25, 0, 0.063, 2.05, 8327.980, 1.5, 25, 0, 0.062, 0.08, 8346.716, -0.3, 0, 0, 0.061, 4.85, 36.048, -3.7, -44, 0, 0.061, 2.58, 6063.386, -2.2, -19, 0, 0.061, 4.30, -766.864, 2.5, 29, 0, 0.060, 3.01, 8322.132, -0.3, 0, 0, 0.059, 0.44, 25057.062, 2.7, 53, 0, 0.059, 0.31, 8288.877, 1.5, 25, 0, 0.059, 2.35, 41643.457, 7.6, 125, -1, 0.059, 1.26, 8368.506, 1.5, 25, 0, 0.058, 1.80, 39900.527, 3.9, 81, 0, 0.058, 1.87, 13590.274, 0, 13, 0, 0.057, 0.47, 14954.262, -0.7, 6, 0, 0.057, 6.20, 8294.910, 1.8, 29, 0, 0.056, 4.63, -8362.473, -1.2, -21, 0, 0.055, 2.86, 8170.957, 1.5, 25, 0, 0.055, 0.03, 7632.815, 2.1, 32, 0, 0.053, 0.80, 7180.306, -1.9, -15, 0, 0.053, 4.64, 6028.447, -4.0, -41, 0, 0.053, 4.57, 15385.020, -0.7, 6, 0, 0.052, 0.60, 37671.269, -3.5, -6, 0, 0.052, 4.99, 8486.426, 1.5, 25, 0, 0.051, 4.62, 17913.987, 3.0, 50, 0, 0.050, 1.64, 837.851, -4.4, -51, 0, 0.049, 5.79, 7542.649, 1.5, 25, 0, 0.049, 2.06, 9114.733, 1.5, 25, 0, 0.049, 2.21, 7056.329, -2.2, -19, 0, 0.049, 4.90, 7214.063, -2.2, -19, 0, 0.048, 5.39, -1671.943, -5.6, -66, 0, 0.047, 4.90, -26100.703, -8.3, -119, 1, 0.047, 1.60, -9024.567, -0.9, -18, 0, 0.046, 1.16, 7161.094, -2.2, -19, 0, 0.046, 5.77, 30943.533, 2.4, 56, 0, 0.046, 2.43, 22199.503, -4.7, -35, 0, 0.046, 3.16, 14991.999, -0.7, 6, 0, 0.044, 4.11, 48857.520, 5.4, 106, -1, 0.044, 4.39, 6625.570, -2.2, -19, 0, 0.044, 6.06, 7789.401, -2.2, -19, 0, 0.043, 0.14, 16693.431, -0.7, 6, 0, 0.043, 4.50, 15020.385, -0.7, 6, 0, 0.043, 4.35, 5471.132, -5.9, -63, 0, 0.043, 4.32, 575.338, 0, 0, 0, 0.043, 5.43, 7267.032, -2.2, -19, 0, 0.043, 3.82, 16328.796, -0.7, 6, 0, 0.042, 2.73, -17424.247, -0.6, -21, 0, 0.041, 3.60, 15489.785, -0.7, 6, 0, 0.040, 2.62, 16655.082, 4.6, 75, 0, 0.040, 4.23, 8351.233, -2.2, -19, 0, 0.039, 0.66, -6443.786, -1.6, -25, 0, 0.039, 2.13, 16118.093, -0.7, 6, 0, 0.039, 5.86, 7247.820, -2.5, -23, 0, 0.038, 4.56, 7285.051, -4.1, -41, 0, 0.038, 2.59, 9179.168, -2.2, -19, 0, 0.038, 1.42, 393.021, 0, 0, 0, 0.038, 4.94, 8381.661, 1.5, 25, 0, 0.037, 5.06, 23452.693, -3.4, -20, 0, 0.037, 5.11, 9027.981, -0.4, 0, 0, 0.037, 4.98, 7740.199, 1.5, 25, 0, 0.037, 3.66, 16659.684, 1.5, 25, 0, 0.037, 2.89, 8275.722, 1.5, 25, 0, 0.037, 4.26, 40042.502, 0.2, 38, 0, 0.036, 1.95, 8326.062, 1.5, 25, 0, 0.036, 5.90, 8331.321, 1.5, 25, 0, 0.035, 1.33, 15595.723, -0.7, 6, 0, 0.035, 1.39, 7777.936, 2, 25, 0, 0.035, 0.80, 6663.308, -2, -19, 0, 0.035, 0.53, 64.434, -4, -44, 0, 0.034, 2.15, 6691.693, -2, -19, 0, 0.034, 1.90, -8467.253, 1, 0, 0, 0.033, 2.83, 7806.322, 2, 25, 0, 0.033, 4.67, 9479.368, 2, 25, 0, 0.033, 1.41, 418.752, 4, 51, 0},
			//MR1
			{0.5139, 4.1569, 14914.452335, -0.635, 6.2, -0.04, 0.3824, 1.8013, 6585.760910, -2.158, -19, 0.09, 0.3265, 2.3987, 7700.38947, 1.550, 25, -0.12, 0.2640, 5.4540, 8956.99338, 1.496, 25, -0.13, 0.1230, 3.0985, 628.30196, -0.027, 0, 0, 0.0775, 0.929, 16171.05625, -0.69, 6, 0, 0.0607, 4.857, 7842.36482, -2.21, -19, 0.1, 0.0497, 4.200, 14286.15038, -0.61, 6, 0, 0.0419, 5.155, 8399.67910, -0.36, 3, 0, 0.0322, 0.229, 23243.1438, 0.89, 31, -0.2, 0.0253, 2.587, -1742.9305, -3.68, -44, 0.2, 0.0249, 1.844, 5957.4590, -2.13, -19, 0.1, 0.0176, 4.754, 16029.0809, 3.07, 50, -0.2, 0.0145, 1.526, 17285.6848, 3.02, 50, -0.3, 0.0137, 1.004, 15542.7543, -0.66, 6, 0, 0.0126, 5.010, 8326.3902, 3.05, 50, 0, 0.0119, 4.814, 8470.6668, -2.24, -19, 0, 0.0118, 2.843, 8330.9926, 0, 0, 0, 0.0107, 2.442, 7072.0875, 1.6, 25, 0, 0.0099, 5.92, 22128.5152, -2.8, -13, 0, 0.0066, 3.28, 24499.7477, 0.8, 31, 0, 0.0065, 4.90, 7214.0629, -2.2, -19, 0, 0.0059, 5.41, 9585.2953, 1.5, 25, 0, 0.0054, 3.06, 1256.6039, -0.1, 0, 0, 0.0052, 2.50, 8328.3391, 1.5, 25, 0, 0.0052, 5.35, 8329.0437, 1.5, 25, 0, 0.0048, 3.56, 13799.8238, -4.3, -38, 0, 0.0039, 1.99, 30457.2066, -1.3, 12, 0, 0.0035, 0.49, 15540.4531, 0.9, 31, 0, 0.0035, 4.60, 15545.0555, -2.2, -19, 0, 0.0033, 0.27, 22614.842, 0.9, 31, 0, 0.0031, 4.24, 13657.848, -0.6, 6, 0, 0.0023, 1.23, 16728.371, 1.2, 28, 0, 0.0023, 5.50, 8328.691, 1.5, 25, 0, 0.0023, 0, 0, 0, 0, 0, 0.0023, 4.41, 7211.762, -0.7, 6, 0, 0.0022, 1.39, 8311.771, -2.2, -19, 0, 0.0022, 2.25, 7216.364, -3.7, -44, 0, 0.0021, 5.94, 70.988, -1.9, -22, 0, 0.0021, 2.58, 31571.835, 2.4, 56, 0, 0.0017, 2.63, -2371.232, -3.7, -44, 0, 0.0016, 4.04, -1952.480, 0.6, 7, 0, 0.0015, 4.23, 8329.403, 1.5, 25, 0, 0.0015, 3.63, 8327.980, 1.5, 25, 0, 0.0015, 2.69, 23385.119, -2.9, -13, 0, 0.0014, 5.96, 21500.213, -2.8, -13, 0, 0.0013, 4.06, 15542.402, -0.7, 6, 0, 0.0013, 1.02, 15543.107, -0.7, 6, 0, 0.0013, 2.10, 7143.075, -0.3, 0, 0, 0.0012, 0.23, -10071.622, -5.2, -69, 0, 0.0011, 3.33, 23871.446, 1, 31, 0, 0.0011, 1.89, 5329.157, -2, -19, 0},
			//MR2
			{0.001490, 4.157, 14914.45233, -0.64, 6, 0, 0.001110, 1.801, 6585.7609, -2.16, -19, 0.1, 0.000950, 2.399, 7700.3895, 1.55, 25, -0.1, 0.000770, 5.454, 8956.9934, 1.50, 25, -0.1, 0.000360, 3.098, 628.3020, 0, 0, 0, 0.000230, 0.93, 16171.0562, -0.7, 6, 0, 0.000180, 4.86, 7842.3648, -2.2, -19, 0, 0.000140, 4.20, 14286.1504, -0.6, 6, 0, 0.000120, 5.16, 8399.6791, -0.4, 0, 0, 0.000116, 3.46, 8326.390, 3.0, 50, 0, 0.000109, 4.39, 8330.993, 0, 0, 0, 0.000090, 0.23, 23243.144, 0.9, 31, 0}}}
	return XL1
}

func MoonCalcNew(zn int, JD float64) float64 {
	rad := 180.0 * 3600.0 / math.Pi
	t := (JD - 2451545.0) / 36525.0
	XL1 := GetMoonCir()
	ob := XL1[zn]
	var v float64
	var tn float64 = 1
	t2 := t * t
	t3 := t2 * t
	t4 := t3 * t
	t5 := t4 * t
	tx := t - 10
	if zn == 0 {
		v += (3.81034409 + 8399.684730072*t - 3.319e-05*t2 + 3.11e-08*t3 - 2.033e-10*t4) * rad //月球平黄经(弧度)
		v += 5028.792262*t + 1.1124406*t2 + 0.00007699*t3 - 0.000023479*t4 - 0.0000000178*t5   //岁差(角秒)
		if tx > 0 {
			v += -0.866 + 1.43*tx + 0.054*tx*tx //对公元3000年至公元5000年的拟合,最大误差小于10角秒
		}
	}
	t2 /= 1e4
	t3 /= 1e8
	t4 /= 1e8
	n := len(ob[0])
	for i := 0; i < len(ob); i++ {
		F := ob[i]
		N := math.Floor(float64(n*len(F))/float64(len(ob[0])) + 0.5)
		if i != 0 {
			N += 6
		}
		if N >= float64(len(F)) {
			N = float64(len(F))
		}
		var c float64 = 0
		for j := 0; float64(j) < N; j += 6 {
			c += F[j] * math.Cos(F[j+1]+t*F[j+2]+t2*F[j+3]+t3*F[j+4]+t4*F[j+5])
		}
		v += c * tn
		tn *= t
	}
	if zn != 2 {
		v /= rad
	}
	return v
}

func HMoonTrueLo(JD float64) float64 { //计算月亮
	v := MoonCalcNew(0, JD) * 180 / math.Pi
	return Limit360(v)
}

func HMoonTrueBo(JD float64) float64 {
	v := MoonCalcNew(1, JD) * 180 / math.Pi
	return v
}

func HMoonAway(JD float64) float64 { //'月地距离
	v := MoonCalcNew(2, JD)
	return v
}

/*
 * @name 月球视黄经
 */
func HMoonApparentLo(JD float64) float64 {
	return HMoonTrueLo(JD) + Nutation2000Bi(JD)
}

func HMoonTrueRaDec(JD float64) (float64, float64) {
	return LoBoToRaDec(JD, HMoonApparentLo(JD), HMoonTrueBo(JD))
}

/*
 * 月球真赤纬
 */
func HMoonTrueDec(JD float64) float64 {
	MoonLo := HMoonApparentLo(JD)
	MoonBo := HMoonTrueBo(JD)
	tmp := Sin(MoonBo)*Cos(Sita(JD)) + Cos(MoonBo)*Sin(Sita(JD))*Sin(MoonLo)
	res := ArcSin(tmp)
	return res
}

/*
 * 月球真赤经
 */
func HMoonTrueRa(JD float64) float64 {
	return LoToRa(JD, HMoonApparentLo(JD), HMoonTrueBo(JD))
}

/*
*
*
传入世界时
*/
func HMoonApparentRaDec(JD, lon, lat, tz float64) (float64, float64) {
	jde := TD2UT(JD, true)
	ra := HMoonTrueRa(jde - tz/24)
	dec := HMoonTrueDec(jde - tz/24)
	away := HMoonAway(jde-tz/24) / 149597870.7
	nra, ndec := ZhanXinRaDec(ra, dec, lat, lon, JD-tz/24, away, 0)
	return nra, ndec
}

func HMoonApparentRa(JD, lon, lat, tz float64) float64 {
	jde := TD2UT(JD, true)
	ra := HMoonTrueRa(jde - tz/24)
	dec := HMoonTrueDec(jde - tz/24)
	away := HMoonAway(jde-tz/24) / 149597870.7
	nra := ZhanXinRa(ra, dec, lat, lon, JD-tz/24, away, 0)
	return nra
}
func HMoonApparentDec(JD, lon, lat, tz float64) float64 {
	jde := TD2UT(JD, true)
	ra := HMoonTrueRa(jde - tz/24)
	dec := HMoonTrueDec(jde - tz/24)
	away := HMoonAway(jde-tz/24) / 149597870.7
	ndec := ZhanXinDec(ra, dec, lat, lon, JD-tz/24, away, 0)
	return ndec
}
