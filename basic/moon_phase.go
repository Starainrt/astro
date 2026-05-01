package basic

import (
	"math"

	. "github.com/starainrt/astro/tools"
)

func MoonPhase(jd float64) float64 {
	moonBo := HMoonTrueBo(jd)
	sunLo := HSunApparentLo(jd)
	moonLo := HMoonApparentLo(jd)
	tmp := Cos(moonBo) * Cos(sunLo-moonLo)
	earthSunDistance := Distance(jd) * 149597870.691
	i := earthSunDistance * Sin(ArcCos(tmp)) / (HMoonAway(jd) - earthSunDistance*tmp)
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

func SunMoonSeek(jde float64, degree float64) float64 {
	p := HMoonApparentLo(jde) - (HSunApparentLo(jde)) - degree
	for p < -180 {
		p += 360
	}
	for p > 180 {
		p -= 360
	}
	return p
}

func CalcMoonSHByJDE(jde float64, phaseType int) float64 {
	phaseType = phaseType * 180
	estimateJD := jde
	for {
		prevJD := estimateJD
		stDegree := SunMoonSeek(prevJD, float64(phaseType))
		stDegreep := (SunMoonSeek(prevJD+0.000005, float64(phaseType)) - SunMoonSeek(prevJD-0.000005, float64(phaseType))) / 0.00001
		estimateJD = prevJD - stDegree/stDegreep
		if math.Abs(estimateJD-prevJD) <= 0.00001 {
			break
		}
	}
	return estimateJD
}

func CalcMoonSH(year float64, phaseType int) float64 {
	jde := CalcMoonS(year, phaseType)
	phaseType = phaseType * 180
	estimateJD := jde
	for {
		prevJD := estimateJD
		stDegree := SunMoonSeek(prevJD, float64(phaseType))
		stDegreep := (SunMoonSeek(prevJD+0.000005, float64(phaseType)) - SunMoonSeek(prevJD-0.000005, float64(phaseType))) / 0.00001
		estimateJD = prevJD - stDegree/stDegreep
		if math.Abs(estimateJD-prevJD) <= 0.00001 {
			break
		}
	}
	return estimateJD
}

/*
 * C=0朔月时刻 =1 望月
 */
func CalcMoonS(year float64, phaseType int) float64 {
	k := math.Floor((year - 2000) * 12.36827)
	if phaseType == 1 {
		k += 0.5
	}
	T := k / 1236.85
	jde := 2451550.09765 + 29.530588853*k + 0.0001337*T*T - 0.000000150*T*T*T + 0.00000000073*T*T*T*T
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
	angles := []float64{N, M, 2 * N, 2 * F, N - M, N + M, 2 * M, N - 2*F, N + 2*F, 2*N + M, 3 * N, M + 2*F, M - 2*F, 2*N - M, O, N + 2*M, 2*N - 2*F, 3 * M, N + M - 2*F, 2*N + 2*F, N + M + 2*F, N - M + 2*F, N - M - 2*F, 3*N + M, 4 * N}
	var coeffs []float64
	if phaseType == 0 {
		coeffs = []float64{-0.40720, 0.17241 * E, 0.01608, 0.01039, 0.00739 * E, -0.00514 * E, 0.00208 * E * E, -0.00111, -0.00057, 0.00056 * E, -0.00042, 0.00042 * E, 0.00038 * E, -0.00024 * E, -0.00017, -0.00007, 0.00004, 0.00004, 0.00003, 0.00003, -0.00003, 0.00003, -0.00002, -0.00002, 0.00002}
	} else {
		coeffs = []float64{-0.40614, 0.17302 * E, 0.01614, 0.01043, 0.00734 * E, -0.00515 * E, 0.00209 * E * E, -0.00111, -0.00057, 0.00056 * E, -0.00042, 0.00042 * E, 0.00038 * E, -0.00024 * E, -0.00017, -0.00007, 0.00004, 0.00004, 0.00003, 0.00003, -0.00003, 0.00003, -0.00002, -0.00002, 0.00002}
	}
	var correction float64
	for idx, angle := range angles {
		correction += Sin(angle) * coeffs[idx]
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
	planetaryCorrection := 325*Sin(A1) + 165*Sin(A2) + 164*Sin(A3) + 126*Sin(A4) + 110*Sin(A5) + 62*Sin(A6) + 60*Sin(A7) + 56*Sin(A8) + 47*Sin(A9) + 42*Sin(A10) + 40*Sin(A11) + 37*Sin(A12) + 35*Sin(A13) + 23*Sin(A14)
	planetaryCorrection /= 1000000
	jde = jde + planetaryCorrection + correction
	return jde
}

func CalcMoonXHByJDE(jde float64, quarterType int) float64 {
	if quarterType == 0 {
		quarterType = 90
	} else {
		quarterType = -90
	}
	estimateJD := jde
	for {
		prevJD := estimateJD
		stDegree := SunMoonSeek(prevJD, float64(quarterType))
		stDegreep := (SunMoonSeek(prevJD+0.000005, float64(quarterType)) - SunMoonSeek(prevJD-0.000005, float64(quarterType))) / 0.00001
		estimateJD = prevJD - stDegree/stDegreep
		if math.Abs(estimateJD-prevJD) <= 0.00001 {
			break
		}
	}
	return estimateJD
}

func CalcMoonXH(year float64, quarterType int) float64 {
	jde := CalcMoonX(year, quarterType)
	if quarterType == 0 {
		quarterType = 90
	} else {
		quarterType = -90
	}
	estimateJD := jde
	for {
		prevJD := estimateJD
		stDegree := SunMoonSeek(prevJD, float64(quarterType))
		stDegreep := (SunMoonSeek(prevJD+0.000005, float64(quarterType)) - SunMoonSeek(prevJD-0.000005, float64(quarterType))) / 0.00001
		estimateJD = prevJD - stDegree/stDegreep
		if math.Abs(estimateJD-prevJD) <= 0.00001 {
			break
		}
	}
	return estimateJD
}

func CalcMoonX(year float64, quarterType int) float64 {
	k := math.Floor((year-2000)*12.36827) + 0.25
	if quarterType == 1 {
		k += 0.5
	}
	T := k / 1236.85
	jde := 2451550.09765 + 29.530588853*k + 0.0001337*T*T - 0.000000150*T*T*T + 0.00000000073*T*T*T*T
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
	var correction float64
	for idx, angle := range ZQ {
		correction += Sin(angle) * MN[idx]
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
	planetaryCorrection := 325*Sin(A1) + 165*Sin(A2) + 164*Sin(A3) + 126*Sin(A4) + 110*Sin(A5) + 62*Sin(A6) + 60*Sin(A7) + 56*Sin(A8) + 47*Sin(A9) + 42*Sin(A10) + 40*Sin(A11) + 37*Sin(A12) + 35*Sin(A13) + 23*Sin(A14)
	planetaryCorrection /= 1000000
	//die(tmp2);
	//die(JDE." ".tmp." ".tmp2." ".W);
	jde = jde + planetaryCorrection + correction
	if quarterType == 0 {
		jde += W
	} else {
		jde -= W
	}
	return jde
}
