package planet

import "math"

// GetMoonCir 返回旧月球级数系数的独立副本 / an independent copy of the legacy lunar-series coefficients.
func GetMoonCir() [][][]float64 {
	cir := moonCirTable()
	result := make([][][]float64, len(cir))
	for coordinateIndex, coordinate := range cir {
		result[coordinateIndex] = make([][]float64, len(coordinate))
		for seriesIndex, series := range coordinate {
			result[coordinateIndex][seriesIndex] = append([]float64(nil), series...)
		}
	}
	return result
}

// WhereMoon 当前星历模型下的地心月球历元平黄道几何量 / geocentric Moon of date under the active ephemeris model.
func WhereMoon(coordIndex int, jd float64) float64 {
	return WhereMoonN(coordIndex, jd, -1)
}

// WhereMoonN WhereMoon 的 n 变体：coordIndex 0 黄经、1 黄纬（弧度）、2 距离（km）；旧级数黄经连续累加，半解析模型落在 [0, 2π)，n >= 0 用快速表 / n-parameterised WhereMoon.
func WhereMoonN(coordIndex int, jd float64, n int) float64 {
	if semianalyticActive() && coordIndex >= 0 && coordIndex <= 2 {
		return semiMoon(coordIndex, jd, n)
	}
	return elpMoonN(coordIndex, jd, n)
}

// moonPhaseCompensationThreshold 是启用线性相位补偿的最小 |t|（世纪数）倒数尺度：
// 更小的项本身已低于纳角秒级相位舍入，无需补偿。
const moonPhaseCompensationThreshold = 1e6

func elpMoonN(coordIndex int, jd float64, n int) float64 {
	arcsecPerRadian := 180.0 * 3600.0 / math.Pi
	// Preserve the low parts of both the epoch subtraction and century division.
	days := jd - 2451545.0
	origin := days - jd
	daysLow := (jd - (days - origin)) - (2451545 + origin)
	t := days / 36525.0
	tLow := (math.FMA(-t, 36525, days) + daysLow) / 36525
	// Small angular terms have sub-nanoarcsecond phase roundoff already.
	// Compensate only the larger terms; distance needs no angular precision.
	phaseThreshold := moonPhaseCompensationThreshold / math.Abs(t)
	ob := moonCirTable()[coordIndex]
	var v float64
	var tn float64 = 1
	t2 := t * t
	t3 := t2 * t
	t4 := t3 * t
	t5 := t4 * t
	tx := t - 10
	if coordIndex == 0 {
		// 黄经线性项必须始终走补偿版本：MoonCalcNew 与 HMoonTrueLo 是同一个物理量的两个入口，
		// 只在其中一个入口补偿会让两者差约 1e-9 度。
		longitude := moonReducedLinearPhase(3.81034409, 8399.684730072, t, tLow)
		v += (longitude - 3.319e-05*t2 + 3.11e-08*t3 - 2.033e-10*t4) * arcsecPerRadian
		v += 5028.792262*t + 1.1124406*t2 + 0.00007699*t3 - 0.000023479*t4 - 0.0000000178*t5 //岁差(角秒)
		if tx > 0 {
			v += -0.866 + 1.43*tx + 0.054*tx*tx //对公元3000年至公元5000年的拟合,最大误差小于10角秒
		}
	}
	t2 /= 1e4
	t3 /= 1e8
	t4 /= 1e8
	nScalars := n * 6
	if nScalars < 0 {
		nScalars = len(ob[0])
	}
	for i := 0; i < len(ob); i++ {
		F := ob[i]
		seriesPhaseThreshold := phaseThreshold / math.Abs(tn)
		N := math.Floor(float64(nScalars*len(F))/float64(len(ob[0])) + 0.5)
		if i != 0 {
			N += 6
		}
		if N >= float64(len(F)) {
			N = float64(len(F))
		}
		var c float64 = 0
		for j := 0; float64(j) < N; j += 6 {
			phase := F[j+1] + t*F[j+2]
			if coordIndex != 2 && math.Abs(F[j]*F[j+2]) > seriesPhaseThreshold {
				phase = moonReducedLinearPhase(F[j+1], F[j+2], t, tLow)
			}
			c += F[j] * math.Cos(phase+t2*F[j+3]+t3*F[j+4]+t4*F[j+5])
		}
		v += c * tn
		tn *= t
	}
	if coordIndex != 2 {
		v /= arcsecPerRadian
	}
	return v
}

// moonReducedLinearPhase 小增量不进长期项乘积；2π 低位防止整圈处约化跳变。
func moonReducedLinearPhase(offset, rate, t, tLow float64) float64 {
	const twoPiLow = 2.4492935982947064e-16
	product := rate * t
	roundoff := math.FMA(rate, t, -product) + rate*tLow
	turns := math.Round(product / (2 * math.Pi))
	return math.FMA(-turns, 2*math.Pi, product) + (roundoff - turns*twoPiLow) + offset
}
