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
func HJZD(JD float64) float64 { // '黄经章动

	// Dim p As Double, T As Double, Lmoon As Double
	T := (JD - 2451545) / 36525.000
	D := 297.8502042 + 445267.1115168*T - 0.0016300*T*T + T*T*T/545868 - T*T*T*T/113065000
	M := SunM(JD)
	N := MoonM(JD)
	F := MoonLonX(JD)
	O := 125.04452 - 1934.136261*T + 0.0020708*T*T + T*T*T/450000
	//die(T." ".D." ".M." ".N." ".F." ".O);
	tp := make(map[int]map[int]float64)
	for i := 1; i < 64; i++ {
		tp[i] = make(map[int]float64)
	}
	tp[1][1] = 0
	tp[1][2] = 0
	tp[1][3] = 0
	tp[1][4] = 0
	tp[1][5] = 1
	tp[1][6] = -171996
	tp[1][7] = -174.2 * T
	tp[2][1] = -2
	tp[2][2] = 0
	tp[2][3] = 0
	tp[2][4] = 2
	tp[2][5] = 2
	tp[2][6] = -13187
	tp[2][7] = -1.6 * T
	tp[3][1] = 0
	tp[3][2] = 0
	tp[3][3] = 0
	tp[3][4] = 2
	tp[3][5] = 2
	tp[3][6] = -2274
	tp[3][7] = -0.2 * T
	tp[4][1] = 0
	tp[4][2] = 0
	tp[4][3] = 0
	tp[4][4] = 0
	tp[4][5] = 2
	tp[4][6] = 2062
	tp[4][7] = 0.2 * T
	tp[5][1] = 0
	tp[5][2] = 1
	tp[5][3] = 0
	tp[5][4] = 0
	tp[5][5] = 0
	tp[5][6] = 1426
	tp[5][7] = -3.4 * T
	tp[6][1] = 0
	tp[6][2] = 0
	tp[6][3] = 1
	tp[6][4] = 0
	tp[6][5] = 0
	tp[6][6] = 712
	tp[6][7] = 0.1 * T
	tp[7][1] = -2
	tp[7][2] = 1
	tp[7][3] = 0
	tp[7][4] = 2
	tp[7][5] = 2
	tp[7][6] = -517
	tp[7][7] = 1.2 * T
	tp[8][1] = 0
	tp[8][2] = 0
	tp[8][3] = 0
	tp[8][4] = 2
	tp[8][5] = 1
	tp[8][6] = -386
	tp[8][7] = -0.4 * T
	tp[9][1] = 0
	tp[9][2] = 0
	tp[9][3] = 1
	tp[9][4] = 2
	tp[9][5] = 2
	tp[9][6] = -301
	tp[9][7] = 0
	tp[10][1] = -2
	tp[10][2] = -1
	tp[10][3] = 0
	tp[10][4] = 2
	tp[10][5] = 2
	tp[10][6] = 217
	tp[10][7] = -0.5 * T
	tp[11][1] = -2
	tp[11][2] = 0
	tp[11][3] = 1
	tp[11][4] = 0
	tp[11][5] = 0
	tp[11][6] = -158
	tp[11][7] = 0
	tp[12][1] = -2
	tp[12][2] = 0
	tp[12][3] = 0
	tp[12][4] = 2
	tp[12][5] = 1
	tp[12][6] = 129
	tp[12][7] = 0.1 * T
	tp[13][1] = 0
	tp[13][2] = 0
	tp[13][3] = -1
	tp[13][4] = 2
	tp[13][5] = 2
	tp[13][6] = 123
	tp[13][7] = 0
	tp[14][1] = 2
	tp[14][2] = 0
	tp[14][3] = 0
	tp[14][4] = 0
	tp[14][5] = 0
	tp[14][6] = 63
	tp[14][7] = 0
	tp[15][1] = 0
	tp[15][2] = 0
	tp[15][3] = 1
	tp[15][4] = 0
	tp[15][5] = 1
	tp[15][6] = 63
	tp[15][7] = 0.1 * T
	tp[16][1] = 2
	tp[16][2] = 0
	tp[16][3] = -1
	tp[16][4] = 2
	tp[16][5] = 2
	tp[16][6] = -59
	tp[16][7] = 0
	tp[17][1] = 0
	tp[17][2] = 0
	tp[17][3] = -1
	tp[17][4] = 0
	tp[17][5] = 1
	tp[17][6] = -58
	tp[17][7] = -0.1 * T
	tp[18][1] = 0
	tp[18][2] = 0
	tp[18][3] = 1
	tp[18][4] = 2
	tp[18][5] = 1
	tp[18][6] = -51
	tp[18][7] = 0
	tp[19][1] = -2
	tp[19][2] = 0
	tp[19][3] = 2
	tp[19][4] = 0
	tp[19][5] = 0
	tp[19][6] = 48
	tp[19][7] = 0
	tp[20][1] = 0
	tp[20][2] = 0
	tp[20][3] = -2
	tp[20][4] = 2
	tp[20][5] = 1
	tp[20][6] = 46
	tp[20][7] = 0
	tp[21][1] = 2
	tp[21][2] = 0
	tp[21][3] = 0
	tp[21][4] = 2
	tp[21][5] = 2
	tp[21][6] = -38
	tp[21][7] = 0
	tp[22][1] = 0
	tp[22][2] = 0
	tp[22][3] = 2
	tp[22][4] = 2
	tp[22][5] = 2
	tp[22][6] = -31
	tp[22][7] = 0
	tp[23][1] = 0
	tp[23][2] = 0
	tp[23][3] = 2
	tp[23][4] = 0
	tp[23][5] = 0
	tp[23][6] = 29
	tp[23][7] = 0
	tp[24][1] = -2
	tp[24][2] = 0
	tp[24][3] = 1
	tp[24][4] = 2
	tp[24][5] = 2
	tp[24][6] = 29
	tp[24][7] = 0
	tp[25][1] = 0
	tp[25][2] = 0
	tp[25][3] = 0
	tp[25][4] = 2
	tp[25][5] = 0
	tp[25][6] = 26
	tp[25][7] = 0
	tp[26][1] = -2
	tp[26][2] = 0
	tp[26][3] = 0
	tp[26][4] = 2
	tp[26][5] = 0
	tp[26][6] = -22
	tp[26][7] = 0
	tp[27][1] = 0
	tp[27][2] = 0
	tp[27][3] = -1
	tp[27][4] = 2
	tp[27][5] = 1
	tp[27][6] = 21
	tp[27][7] = 0
	tp[28][1] = 0
	tp[28][2] = 2
	tp[28][3] = 0
	tp[28][4] = 0
	tp[28][5] = 0
	tp[28][6] = 17
	tp[28][7] = -0.1 * T
	tp[29][1] = 2
	tp[29][2] = 0
	tp[29][3] = -1
	tp[29][4] = 0
	tp[29][5] = 1
	tp[29][6] = 16
	tp[29][7] = 0
	tp[30][1] = -2
	tp[30][2] = 2
	tp[30][3] = 0
	tp[30][4] = 2
	tp[30][5] = 2
	tp[30][6] = -16
	tp[30][7] = 0.1 * T
	tp[31][1] = 0
	tp[31][2] = 1
	tp[31][3] = 0
	tp[31][4] = 0
	tp[31][5] = 1
	tp[31][6] = -15
	tp[31][7] = 0
	tp[32][1] = -2
	tp[32][2] = 0
	tp[32][3] = 1
	tp[32][4] = 0
	tp[32][5] = 1
	tp[32][6] = -13
	tp[32][7] = 0
	tp[33][1] = 0
	tp[33][2] = -1
	tp[33][3] = 0
	tp[33][4] = 0
	tp[33][5] = 1
	tp[33][6] = -12
	tp[33][7] = 0
	tp[34][1] = 0
	tp[34][2] = 0
	tp[34][3] = 2
	tp[34][4] = -2
	tp[34][5] = 0
	tp[34][6] = 11
	tp[34][7] = 0
	tp[35][1] = 2
	tp[35][2] = 0
	tp[35][3] = -1
	tp[35][4] = 2
	tp[35][5] = 1
	tp[35][6] = -10
	tp[35][7] = 0
	tp[36][1] = 2
	tp[36][2] = 0
	tp[36][3] = 1
	tp[36][4] = 2
	tp[36][5] = 2
	tp[36][6] = -8
	tp[36][7] = 0
	tp[37][1] = 0
	tp[37][2] = 1
	tp[37][3] = 0
	tp[37][4] = 2
	tp[37][5] = 2
	tp[37][6] = 7
	tp[37][7] = 0
	tp[38][1] = -2
	tp[38][2] = 1
	tp[38][3] = 1
	tp[38][4] = 0
	tp[38][5] = 0
	tp[38][6] = -7
	tp[38][7] = 0
	tp[39][1] = 0
	tp[39][2] = -1
	tp[39][3] = 0
	tp[39][4] = 2
	tp[39][5] = 2
	tp[39][6] = -7
	tp[39][7] = 0
	tp[40][1] = 2
	tp[40][2] = 0
	tp[40][3] = 0
	tp[40][4] = 2
	tp[40][5] = 1
	tp[40][6] = -7
	tp[40][7] = 0
	tp[41][1] = 2
	tp[41][2] = 0
	tp[41][3] = 1
	tp[41][4] = 0
	tp[41][5] = 0
	tp[41][6] = 6
	tp[41][7] = 0
	tp[42][1] = -2
	tp[42][2] = 0
	tp[42][3] = 2
	tp[42][4] = 2
	tp[42][5] = 2
	tp[42][6] = 6
	tp[42][7] = 0
	tp[43][1] = -2
	tp[43][2] = 0
	tp[43][3] = 1
	tp[43][4] = 2
	tp[43][5] = 1
	tp[43][6] = 6
	tp[43][7] = 0
	tp[44][1] = 2
	tp[44][2] = 0
	tp[44][3] = -2
	tp[44][4] = 0
	tp[44][5] = 1
	tp[44][6] = -6
	tp[44][7] = 0
	tp[45][1] = 2
	tp[45][2] = 0
	tp[45][3] = 0
	tp[45][4] = 0
	tp[45][5] = 1
	tp[45][6] = -6
	tp[45][7] = 0
	tp[46][1] = 0
	tp[46][2] = -1
	tp[46][3] = 1
	tp[46][4] = 0
	tp[46][5] = 0
	tp[46][6] = 5
	tp[46][7] = 0
	tp[47][1] = -2
	tp[47][2] = -1
	tp[47][3] = 0
	tp[47][4] = 2
	tp[47][5] = 1
	tp[47][6] = -5
	tp[47][7] = 0
	tp[48][1] = -2
	tp[48][2] = 0
	tp[48][3] = 0
	tp[48][4] = 0
	tp[48][5] = 1
	tp[48][6] = -5
	tp[48][7] = 0
	tp[49][1] = 0
	tp[49][2] = 0
	tp[49][3] = 2
	tp[49][4] = 2
	tp[49][5] = 1
	tp[49][6] = -5
	tp[49][7] = 0
	tp[50][1] = -2
	tp[50][2] = 0
	tp[50][3] = 2
	tp[50][4] = 0
	tp[50][5] = 1
	tp[50][6] = 4
	tp[50][7] = 0
	tp[51][1] = -2
	tp[51][2] = 1
	tp[51][3] = 0
	tp[51][4] = 2
	tp[51][5] = 1
	tp[51][6] = 4
	tp[51][7] = 0
	tp[52][1] = 0
	tp[52][2] = 0
	tp[52][3] = 1
	tp[52][4] = -2
	tp[52][5] = 0
	tp[52][6] = 4
	tp[52][7] = 0
	tp[53][1] = -1
	tp[53][2] = 0
	tp[53][3] = 1
	tp[53][4] = 0
	tp[53][5] = 0
	tp[53][6] = -4
	tp[53][7] = 0
	tp[54][1] = -2
	tp[54][2] = 1
	tp[54][3] = 0
	tp[54][4] = 0
	tp[54][5] = 0
	tp[54][6] = -4
	tp[54][7] = 0
	tp[55][1] = 1
	tp[55][2] = 0
	tp[55][3] = 0
	tp[55][4] = 0
	tp[55][5] = 0
	tp[55][6] = -4
	tp[55][7] = 0
	tp[56][1] = 0
	tp[56][2] = 0
	tp[56][3] = 1
	tp[56][4] = 2
	tp[56][5] = 0
	tp[56][6] = 3
	tp[56][7] = 0
	tp[57][1] = 0
	tp[57][2] = 0
	tp[57][3] = -2
	tp[57][4] = 2
	tp[57][5] = 2
	tp[57][6] = -3
	tp[57][7] = 0
	tp[58][1] = -1
	tp[58][2] = -1
	tp[58][3] = 1
	tp[58][4] = 0
	tp[58][5] = 0
	tp[58][6] = -3
	tp[58][7] = 0
	tp[59][1] = 0
	tp[59][2] = 1
	tp[59][3] = 1
	tp[59][4] = 0
	tp[59][5] = 0
	tp[59][6] = -3
	tp[59][7] = 0
	tp[60][1] = 0
	tp[60][2] = -1
	tp[60][3] = 1
	tp[60][4] = 2
	tp[60][5] = 2
	tp[60][6] = -3
	tp[60][7] = 0
	tp[61][1] = 2
	tp[61][2] = -1
	tp[61][3] = -1
	tp[61][4] = 2
	tp[61][5] = 2
	tp[61][6] = -3
	tp[61][7] = 0
	tp[62][1] = 0
	tp[62][2] = 0
	tp[62][3] = 3
	tp[62][4] = 2
	tp[62][5] = 2
	tp[62][6] = -3
	tp[62][7] = 0
	tp[63][1] = 2
	tp[63][2] = -1
	tp[63][3] = 0
	tp[63][4] = 2
	tp[63][5] = 2
	tp[63][6] = -3
	tp[63][7] = 0
	var S float64
	for i := 1; i < 64; i++ {
		S += (tp[i][6] + tp[i][7]) * Sin(D*tp[i][1]+M*tp[i][2]+N*tp[i][3]+F*tp[i][4]+O*tp[i][5])
	}
	//P=-17.20*Sin(O)-1.32*Sin(2*280.4665 + 36000.7698*T)-0.23*Sin(2*218.3165 + 481267.8813*T )+0.21*Sin(2*O);
	//return P/3600;
	return (S / 10000) / 3600
}

/*
 * 交角章动
 */
func JJZD(JD float64) float64 { //交角章动

	T := (JD - 2451545) / 36525
	//D = 297.85036 +455267.111480*T - 0.0019142*T*T+ T*T*T/189474;
	//M = 357.52772 + 35999.050340*T - 0.0001603*T*T- T*T*T/300000;
	//N= 134.96298 + 477198.867398*T + 0.0086972*T*T + T*T*T/56250;
	//F = 93.27191 + 483202.017538*T - 0.0036825*T*T + T*T*T/327270;
	D := 297.8502042 + 445267.1115168*T - 0.0016300*T*T + T*T*T/545868 - T*T*T*T/113065000
	M := SunM(JD)
	N := MoonM(JD)
	F := MoonLonX(JD)
	O := 125.04452 - 1934.136261*T + 0.0020708*T*T + T*T*T/450000
	tp := make(map[int]map[int]float64)
	for i := 1; i < 64; i++ {
		tp[i] = make(map[int]float64)
	}
	tp[1][1] = 0
	tp[1][2] = 0
	tp[1][3] = 0
	tp[1][4] = 0
	tp[1][5] = 1
	tp[1][6] = 92025
	tp[1][7] = 8.9 * T
	tp[2][1] = -2
	tp[2][2] = 0
	tp[2][3] = 0
	tp[2][4] = 2
	tp[2][5] = 2
	tp[2][6] = 5736
	tp[2][7] = -3.1 * T
	tp[3][1] = 0
	tp[3][2] = 0
	tp[3][3] = 0
	tp[3][4] = 2
	tp[3][5] = 2
	tp[3][6] = 977
	tp[3][7] = -0.5 * T
	tp[4][1] = 0
	tp[4][2] = 0
	tp[4][3] = 0
	tp[4][4] = 0
	tp[4][5] = 2
	tp[4][6] = -895
	tp[4][7] = 0.5 * T
	tp[5][1] = 0
	tp[5][2] = 1
	tp[5][3] = 0
	tp[5][4] = 0
	tp[5][5] = 0
	tp[5][6] = 54
	tp[5][7] = -0.1 * T
	tp[6][1] = 0
	tp[6][2] = 0
	tp[6][3] = 1
	tp[6][4] = 0
	tp[6][5] = 0
	tp[6][6] = -7
	tp[6][7] = 0
	tp[7][1] = -2
	tp[7][2] = 1
	tp[7][3] = 0
	tp[7][4] = 2
	tp[7][5] = 2
	tp[7][6] = 224
	tp[7][7] = -0.6 * T
	tp[8][1] = 0
	tp[8][2] = 0
	tp[8][3] = 0
	tp[8][4] = 2
	tp[8][5] = 1
	tp[8][6] = 200
	tp[8][7] = 0
	tp[9][1] = 0
	tp[9][2] = 0
	tp[9][3] = 1
	tp[9][4] = 2
	tp[9][5] = 2
	tp[9][6] = 129
	tp[9][7] = -0.1 * T
	tp[10][1] = -2
	tp[10][2] = -1
	tp[10][3] = 0
	tp[10][4] = 2
	tp[10][5] = 2
	tp[10][6] = -95
	tp[10][7] = 0.3 * T
	tp[11][1] = -2
	tp[11][2] = 0
	tp[11][3] = 0
	tp[11][4] = 2
	tp[11][5] = 1
	tp[11][6] = -70
	tp[11][7] = 0
	tp[12][1] = 0
	tp[12][2] = 0
	tp[12][3] = -1
	tp[12][4] = 2
	tp[12][5] = 2
	tp[12][6] = -53
	tp[12][7] = 0
	tp[13][1] = 2
	tp[13][2] = 0
	tp[13][3] = 0
	tp[13][4] = 0
	tp[13][5] = 0
	tp[13][6] = 63
	tp[13][7] = 0
	tp[14][1] = 0
	tp[14][2] = 0
	tp[14][3] = 1
	tp[14][4] = 0
	tp[14][5] = 1
	tp[14][6] = -33
	tp[14][7] = 0
	tp[15][1] = 2
	tp[15][2] = 0
	tp[15][3] = -1
	tp[15][4] = 2
	tp[15][5] = 2
	tp[15][6] = 26
	tp[15][7] = 0
	tp[16][1] = 0
	tp[16][2] = 0
	tp[16][3] = -1
	tp[16][4] = 0
	tp[16][5] = 1
	tp[16][6] = 32
	tp[16][7] = 0
	tp[17][1] = 0
	tp[17][2] = 0
	tp[17][3] = 1
	tp[17][4] = 2
	tp[17][5] = 1
	tp[17][6] = 27
	tp[17][7] = 0
	tp[18][1] = 0
	tp[18][2] = 0
	tp[18][3] = -2
	tp[18][4] = 2
	tp[18][5] = 1
	tp[18][6] = -24
	tp[18][7] = 0
	tp[19][1] = 2
	tp[19][2] = 0
	tp[19][3] = 0
	tp[19][4] = 2
	tp[19][5] = 2
	tp[19][6] = 16
	tp[19][7] = 0
	tp[20][1] = 0
	tp[20][2] = 0
	tp[20][3] = 2
	tp[20][4] = 2
	tp[20][5] = 2
	tp[20][6] = 13
	tp[20][7] = 0
	tp[21][1] = -2
	tp[21][2] = 0
	tp[21][3] = 1
	tp[21][4] = 2
	tp[21][5] = 2
	tp[21][6] = -12
	tp[21][7] = 0
	tp[22][1] = 0
	tp[22][2] = 0
	tp[22][3] = -1
	tp[22][4] = 2
	tp[22][5] = 1
	tp[22][6] = -10
	tp[22][7] = 0
	tp[23][1] = 2
	tp[23][2] = 0
	tp[23][3] = -1
	tp[23][4] = 0
	tp[23][5] = 1
	tp[23][6] = -8
	tp[23][7] = 0
	tp[24][1] = -2
	tp[24][2] = 2
	tp[24][3] = 0
	tp[24][4] = 2
	tp[24][5] = 2
	tp[24][6] = 7
	tp[24][7] = 0
	tp[25][1] = 0
	tp[25][2] = 1
	tp[25][3] = 0
	tp[25][4] = 0
	tp[25][5] = 1
	tp[25][6] = 9
	tp[25][7] = 0
	tp[26][1] = -2
	tp[26][2] = 0
	tp[26][3] = 1
	tp[26][4] = 0
	tp[26][5] = 1
	tp[26][6] = 7
	tp[26][7] = 0
	tp[27][1] = 0
	tp[27][2] = -1
	tp[27][3] = 0
	tp[27][4] = 0
	tp[27][5] = 1
	tp[27][6] = 6
	tp[27][7] = 0
	tp[28][1] = 2
	tp[28][2] = 0
	tp[28][3] = -1
	tp[28][4] = 2
	tp[28][5] = 1
	tp[28][6] = 5
	tp[28][7] = 0
	tp[29][1] = 2
	tp[29][2] = 0
	tp[29][3] = 1
	tp[29][4] = 2
	tp[29][5] = 2
	tp[29][6] = 3
	tp[29][7] = 0
	tp[30][1] = 0
	tp[30][2] = 1
	tp[30][3] = 0
	tp[30][4] = 2
	tp[30][5] = 2
	tp[30][6] = -3
	tp[30][7] = 0
	tp[31][1] = 0
	tp[31][2] = -1
	tp[31][3] = 0
	tp[31][4] = 2
	tp[31][5] = 2
	tp[31][6] = 3
	tp[31][7] = 0
	tp[32][1] = 2
	tp[32][2] = 0
	tp[32][3] = 0
	tp[32][4] = 2
	tp[32][5] = 1
	tp[32][6] = 3
	tp[32][7] = 0
	tp[33][1] = -2
	tp[33][2] = 0
	tp[33][3] = 2
	tp[33][4] = 2
	tp[33][5] = 2
	tp[33][6] = -3
	tp[33][7] = 0
	tp[34][1] = -2
	tp[34][2] = 0
	tp[34][3] = 1
	tp[34][4] = 2
	tp[34][5] = 1
	tp[34][6] = -3
	tp[34][7] = 0
	tp[35][1] = 2
	tp[35][2] = 0
	tp[35][3] = -2
	tp[35][4] = 0
	tp[35][5] = 1
	tp[35][6] = 3
	tp[35][7] = 0
	tp[36][1] = 2
	tp[36][2] = 0
	tp[36][3] = 0
	tp[36][4] = 0
	tp[36][5] = 1
	tp[36][6] = 3
	tp[36][7] = 0
	tp[37][1] = -2
	tp[37][2] = -1
	tp[37][3] = 0
	tp[37][4] = 2
	tp[37][5] = 1
	tp[37][6] = 3
	tp[37][7] = 0
	tp[38][1] = -2
	tp[38][2] = 0
	tp[38][3] = 0
	tp[38][4] = 0
	tp[38][5] = 1
	tp[38][6] = 3
	tp[38][7] = 0
	tp[39][1] = 0
	tp[39][2] = 0
	tp[39][3] = 2
	tp[39][4] = 2
	tp[39][5] = 1
	tp[39][6] = 3
	tp[39][7] = 0
	var S float64 = 0
	for i := 1; i < 40; i++ {
		S += (tp[i][6] + tp[i][7]) * Cos(D*tp[i][1]+M*tp[i][2]+N*tp[i][3]+F*tp[i][4]+O*tp[i][5])
	}
	return S / 10000 / 3600
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

func SunSeeLo(JD float64) float64 { //'太阳视黄经

	T := (JD - 2451545) / 36525
	SunSeeLo := SunTrueLo(JD) - 0.00569 - 0.00478*Sin(125.04-1934.136*T)
	return SunSeeLo
}

func SunSeeRa(JD float64) float64 { // '太阳视赤经
	T := (JD - 2451545) / 36525
	sitas := Sita(JD) + 0.00256*Cos(125.04-1934.136*T)
	SunSeeRa := ArcTan(Cos(sitas) * Sin(SunSeeLo(JD)) / Cos(SunSeeLo(JD)))
	tmp := SunSeeLo(JD)
	if tmp >= 90 && tmp < 180 {
		SunSeeRa = 180 + SunSeeRa
	} else if tmp >= 180 && tmp < 270 {
		SunSeeRa = 180 + SunSeeRa
	} else if tmp >= 270 && tmp <= 360 {
		SunSeeRa = 360 + SunSeeRa
	}
	return SunSeeRa
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

func SunSeeDec(JD float64) float64 { // '太阳视赤纬
	T := (JD - 2451545) / 36525
	sitas := Sita(JD) + 0.00256*Cos(125.04-1934.136*T)
	SunSeeDec := ArcSin(Sin(sitas) * Sin(SunSeeLo(JD)))
	return SunSeeDec
}

func SunTrueDec(JD float64) float64 { // '太阳真赤纬
	sitas := Sita(JD)
	SunTrueDec := ArcSin(Sin(sitas) * Sin(SunTrueLo(JD)))
	return SunTrueDec
}
func SunTime(JD float64) float64 { //均时差

	tm := (SunLo(JD) - 0.0057183 - (HSunSeeRa(JD)) + (HJZD(JD))*Cos(Sita(JD))) / 15
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

func HSunSeeLo(JD float64) float64 {
	t := (JD - 2451545) / 365250.0
	L := HSunTrueLo(JD)
	R := planet.WherePlanet(-1, 2, JD)
	t2 := t * t
	t3 := t2 * t //千年数的各次方
	R += (-0.0020 + 0.0044*t + 0.0213*t2 - 0.0250*t3)
	L = L + HJZD(JD) - 20.4898/R/3600
	return L
}

func EarthAway(JD float64) float64 {
	//t=(JD - 2451545) / 365250;
	//R=Earth_R5(t)+Earth_R4(t)+Earth_R3(t)+Earth_R2(t)+Earth_R1(t)+Earth_R0(t);
	return planet.WherePlanet(0, 2, 2555555)
}

func HSunSeeRaDec(JD float64) (float64, float64) {
	T := (JD - 2451545) / 36525
	sitas := Sita(JD) + 0.00256*Cos(125.04-1934.136*T)
	sitas2 := EclipticObliquity(JD, false) + 0.00256*Cos(125.04-1934.136*T)
	tmp := HSunSeeLo(JD)
	HSunSeeRa := ArcTan(Cos(sitas) * Sin(tmp) / Cos(tmp))
	HSunSeeDec := ArcSin(Sin(sitas2) * Sin(tmp))
	if tmp >= 90 && tmp < 180 {
		HSunSeeRa = 180 + HSunSeeRa
	} else if tmp >= 180 && tmp < 270 {
		HSunSeeRa = 180 + HSunSeeRa
	} else if tmp >= 270 && tmp <= 360 {
		HSunSeeRa = 360 + HSunSeeRa
	}
	return HSunSeeRa, HSunSeeDec
}

func HSunSeeRa(JD float64) float64 { // '太阳视赤经
	T := (JD - 2451545) / 36525
	sitas := Sita(JD) + 0.00256*Cos(125.04-1934.136*T)
	tmp := HSunSeeLo(JD)
	HSunSeeRa := ArcTan(Cos(sitas) * Sin(tmp) / Cos(tmp))
	if tmp >= 90 && tmp < 180 {
		HSunSeeRa = 180 + HSunSeeRa
	} else if tmp >= 180 && tmp < 270 {
		HSunSeeRa = 180 + HSunSeeRa
	} else if tmp >= 270 && tmp <= 360 {
		HSunSeeRa = 360 + HSunSeeRa
	}
	return HSunSeeRa
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

func HSunSeeDec(JD float64) float64 { // '太阳视赤纬
	T := (JD - 2451545) / 36525
	sitas := EclipticObliquity(JD, false) + 0.00256*Cos(125.04-1934.136*T)
	HSunSeeDec := ArcSin(Sin(sitas) * Sin(HSunSeeLo(JD)))
	return HSunSeeDec
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

func GetOneYearMoon(year float64) map[int]float64 {
	var start float64
	var tmp1, tmp float64
	moon := make(map[int]float64)
	if year < 6000 {
		start = year + 11.00/12.00 + 5.00/30.00/12.00
	} else {
		start = year + 9.00/12.00 + 5.00/30.00/12.00
	}
	i := 1
	for j := 1; j < 17; j++ {
		if year > 3000 {
			tmp1 = TD2UT(CalcMoonSH(start+float64(i-1)/12.5, 0)+8.0/24.0, false)
		} else {
			tmp1 = TD2UT(CalcMoonS(start+float64(i-1)/12.5, 0)+8.0/24.0, false)
		}
		if i != 1 {
			if tmp1 == tmp {
				j--
				i++
				continue
			}
		}
		moon[j] = tmp1
		tmp = moon[j]
		i++
		// echo DateCalc(moon[i])."<br />";
	}
	return moon
}
func GetOneYearJQ(year int) map[int]float64 {
	start := 270
	var years int
	jq := make(map[int]float64)
	for i := 1; i < 26; i++ {
		angle := start + 15*(i-1)
		if angle > 360 {
			angle -= 360
		}
		if i > 1 {
			years = year + 1
		} else {
			years = year
		}
		jq[i] = GetJQTime(years, angle) + 8.0/24.0
		//  echo DateCalc(jq[i])."<br />";
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
			if math.Floor(JD1-JD0) <= 0.000001 {
				break
			}
		}
		JD1 -= 0.001
	}
	JD1 += 0.001
	return TD2UT(JD1, false)
}

func JQLospec(JD float64) float64 {
	t := HSunSeeLo(JD)
	if t <= 12 {
		t += 360
	}
	return t
}

func GetXC(jd float64) string { //十二次
	tlo := HSunSeeLo(jd)
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
		if math.Floor(JD1-JD0) <= 0.000001 {
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
	dec := HSunSeeDec(tztime)
	tmp := -Tan((math.Abs(Lat)+An)*(Lat/math.Abs(Lat))) * Tan(dec)
	if math.Abs(tmp) > 1 {
		if SunHeight(tztime, Lon, Lat, ntz) < An {
			return -2 //极夜
		}
		if SunHeight(tztime-0.5, Lon, Lat, ntz) > An {
			return -1 //极昼
		}
	}
	tmp = -Tan(Lat) * Tan(dec)
	rzsc := ArcCos(tmp) / 15
	sunrise := tztime + rzsc/24.0 + 35.0/24.0/60.0
	i := 0
	for LowSunHeight(sunrise, Lon, Lat, ntz) < An {
		i++
		sunrise -= 15 / 60 / 24
		if i > 12 {
			break
		}
	}
	JD1 := sunrise - 5.00/24.00/60.00
	for {
		JD0 := JD1
		stDegree := SunHeight(JD0, Lon, Lat, ntz) - An
		stDegreep := (SunHeight(JD0+0.000005, Lon, Lat, ntz) - SunHeight(JD0-0.000005, Lon, Lat, ntz)) / 0.00001
		JD1 = JD0 - stDegree/stDegreep
		if math.Floor(JD1-JD0) < 0.000001 {
			break
		}
	}
	return JD1 - ntz/24 + TZ/24
}

func GetAsaTime(JD, Lon, Lat, TZ, An float64) float64 {
	JD = math.Floor(JD) + 1.5
	ntz := math.Round(Lon / 15)
	tztime := GetSunTZTime(JD, Lon, ntz)
	dec := HSunSeeDec(tztime)
	tmp := -Tan((math.Abs(Lat)+An)*(Lat/math.Abs(Lat))) * Tan(dec)
	if math.Abs(tmp) > 1 {
		if SunHeight(tztime, Lon, Lat, ntz) < An {
			return -2 //极夜
		}
		if SunHeight(tztime-0.5, Lon, Lat, ntz) > An {
			return -1 //极昼
		}
	}
	tmp = -Tan(Lat) * Tan(dec)
	rzsc := ArcCos(tmp) / 15
	sunrise := tztime - rzsc/24 - 25.0/24.0/60.0
	i := 0
	for LowSunHeight(sunrise, Lon, Lat, ntz) > An {
		i++
		sunrise -= 15 / 60 / 24
		if i > 12 {
			break
		}
	}
	JD1 := sunrise - 5.00/24.00/60.00
	for {
		JD0 := JD1
		stDegree := SunHeight(JD0, Lon, Lat, ntz) - An
		stDegreep := (SunHeight(JD0+0.000005, Lon, Lat, ntz) - SunHeight(JD0-0.000005, Lon, Lat, ntz)) / 0.00001
		JD1 = JD0 - stDegree/stDegreep
		if math.Floor(JD1-JD0) < 0.000001 {
			break
		}
	}
	return JD1 - ntz/24 + TZ/24
}

/*
 * 太阳时角
 */
func SunTimeAngle(JD, Lon, Lat, TZ float64) float64 {
	startime := Limit360(SeeStarTime(JD-TZ/24)*15 + Lon)
	timeangle := startime - HSunSeeRa(TD2UT(JD-TZ/24, true))
	if timeangle < 0 {
		timeangle += 360
	}
	return timeangle
}

/*
 * 精确计算，传入当日0时JDE
 */
func GetSunRiseTime(JD, Lon, Lat, TZ, ZS float64) float64 {
	var An float64
	JD = math.Floor(JD) + 1.5
	ntz := math.Round(Lon / 15)
	if ZS != 0 {
		An = -0.8333
	}
	tztime := GetSunTZTime(JD, Lon, ntz)
	dec := HSunSeeDec(tztime)
	tmp := -Tan(Lat) * Tan(dec)
	if math.Abs(tmp) > 1 {
		if SunHeight(tztime, Lon, Lat, ntz) < 0 {
			return -2 //极夜
		}
		if SunHeight(tztime-0.5, Lon, Lat, ntz) > 0 {
			return -1 //极昼
		}
	}
	rzsc := ArcCos(tmp) / 15
	sunrise := tztime - rzsc/24 - 5.0/24.0/60.0
	JD1 := sunrise
	for {
		JD0 := JD1
		stDegree := SunHeight(JD0, Lon, Lat, ntz) - An
		stDegreep := (SunHeight(JD0+0.000005, Lon, Lat, ntz) - SunHeight(JD0-0.000005, Lon, Lat, ntz)) / 0.00001
		JD1 = JD0 - stDegree/stDegreep
		if math.Floor(JD1-JD0) <= 0.000001 {
			break
		}
	}
	return JD1 - ntz/24 + TZ/24
}
func GetSunDownTime(JD, Lon, Lat, TZ, ZS float64) float64 {
	var An float64
	JD = math.Floor(JD) + 1.5
	ntz := math.Round(Lon / 15)
	if ZS != 0 {
		An = -0.8333
	}
	tztime := GetSunTZTime(JD, Lon, ntz)
	dec := HSunSeeDec(tztime)
	tmp := -Tan(Lat) * Tan(dec)
	if math.Abs(tmp) > 1 {
		if SunHeight(tztime, Lon, Lat, ntz) < 0 {
			return -2 //极夜
		}
		if SunHeight(tztime+0.5, Lon, Lat, ntz) > 0 {
			return -1 //极昼
		}
	}
	rzsc := ArcCos(tmp) / 15.0
	sunrise := tztime + rzsc/24 - 5.0/24.0/60.0
	JD1 := sunrise
	for {
		JD0 := JD1
		stDegree := SunHeight(JD0, Lon, Lat, ntz) - An
		stDegreep := (SunHeight(JD0+0.000005, Lon, Lat, ntz) - SunHeight(JD0-0.000005, Lon, Lat, ntz)) / 0.00001
		JD1 = JD0 - stDegree/stDegreep
		if math.Floor(JD1-JD0) <= 0.000001 {
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
	st := Limit360(SeeStarTime(calcjd)*15 + Lon)
	ra, dec := HSunSeeRaDec(tjde)
	H := Limit360(st - ra)
	tmp2 := Sin(Lat)*Sin(dec) + Cos(dec)*Cos(Lat)*Cos(H)
	return ArcSin(tmp2)
}
func LowSunHeight(JD, Lon, Lat, TZ float64) float64 {
	//tmp := (TZ*15 - Lon) * 4 / 60
	//truejd := JD - tmp/24
	calcjd := JD - TZ/24
	st := Limit360(SeeStarTime(calcjd)*15 + Lon)
	H := Limit360(st - SunSeeRa(TD2UT(calcjd, true)))
	dec := SunSeeDec(TD2UT(calcjd, true))
	tmp2 := Sin(Lat)*Sin(dec) + Cos(dec)*Cos(Lat)*Cos(H)
	return ArcSin(tmp2)
}
func SunAngle(JD, Lon, Lat, TZ float64) float64 {
	//tmp := (TZ*15 - Lon) * 4 / 60
	//truejd := JD - tmp/24
	calcjd := JD - TZ/24
	st := Limit360(SeeStarTime(calcjd)*15 + Lon)
	H := Limit360(st - HSunSeeRa(TD2UT(calcjd, true)))
	tmp2 := Sin(H) / (Cos(H)*Sin(Lat) - Tan(HSunSeeDec(TD2UT(calcjd, true)))*Cos(Lat))
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

/*
* 干支
 */
func GetGZ(year int) string {
	tiangan := []string{"庚", "辛", "壬", "癸", "甲", "乙", "丙", "丁", "戊", "已"}
	dizhi := []string{"申", "酉", "戌", "亥", "子", "丑", "寅", "卯", "辰", "巳", "午", "未"}
	t := year - (year / 100 * 10)
	d := year % 12
	return tiangan[t] + dizhi[d] + "年"
}
