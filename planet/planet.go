package planet

import (
	. "github.com/starainrt/astro/tools"
	"math"
)

// WherePlanet 天体 xt 在儒略日 jde 的 VSOP 结果 / VSOP result for body xt at Julian day jde.
//
// xt 取 -1 或 0..7：0 为地球，1..7 依次为水星、金星、火星、木星、土星、天王星、海王星；-1 表示地球，zn 取 0 时给日心黄经。
// zn 取 0 黄经、1 黄纬、2 日心距（AU）；xt 或 zn 越界返回 NaN 而不 panic。
// xt is -1 or 0..7 (0 Earth, 1..7 Mercury through Neptune; -1 selects Earth, giving its heliocentric longitude for zn 0).
// zn is 0 longitude, 1 latitude, 2 heliocentric distance in AU; an out-of-range xt or zn yields NaN instead of panicking.
func WherePlanet(xt, zn int, jde float64) float64 {
	return WherePlanetN(xt, zn, jde, -1)
}

// WherePlanetN 同 WherePlanet 的截断版 / truncated form of WherePlanet.
//
// n < 0 时使用全部项；否则保留约 n 个主项并按比例缩短高阶项。取值域与越界行为同 WherePlanet。
// When n < 0 all terms are used; otherwise roughly n principal terms are kept and higher orders scaled proportionally. Domain and out-of-range behavior match WherePlanet.
func WherePlanetN(xt, zn int, jde float64, n int) float64 {
	if xt < -1 || xt > 7 || zn < 0 || zn > 2 {
		return math.NaN()
	}

	sata := 0
	if xt == -1 {
		xt = 0
		sata = 1
	}

	rad := 180.0000 * 3600.0000 / math.Pi
	t := (jde - 2451545) / 36525.0000
	t /= 10 // 转为儒略千年数

	body := planetViews()[xt]
	coord := body.coords[zn]
	baseOrderTerms := len(coord.orders[0])

	tn := float64(1)
	var v float64
	for i, series := range coord.orders {
		seriesLength := len(series)
		if seriesLength == 0 {
			continue
		}

		termLimit := seriesLength
		if n >= 0 {
			termLimit = int(math.Floor(3*float64(n)*float64(seriesLength)/float64(baseOrderTerms) + 0.5))
			if i != 0 {
				termLimit += 3
			}
			if termLimit > seriesLength {
				termLimit = seriesLength
			}
		}

		var c float64
		for j := 0; j < termLimit; j += 3 {
			c += series[j] * math.Cos(series[j+1]+t*series[j+2])
		}
		v += c * tn
		tn *= t
	}
	v /= body.scale

	if xt == 0 { // 地球
		t2 := t * t
		t3 := t2 * t // 千年数的各次方
		if zn == 0 {
			v += (-0.0728 - 2.7702*t - 1.1019*t2 - 0.0996*t3) / rad
		} else if zn == 1 {
			v += (+0.0000 + 0.0004*t + 0.0004*t2 - 0.0026*t3) / rad
		} else if zn == 2 {
			v += (-0.0020 + 0.0044*t + 0.0213*t2 - 0.0250*t3) / 1000000
		}
	} else { // 其它行星
		planetCorrections := []float64{
			// 经(角秒), 纬(角秒), 距(10-6AU)
			-0.08631, +0.00039, -0.00008, // 水星
			-0.07447, +0.00006, +0.00017, // 金星
			-0.07135, -0.00026, -0.00176, // 火星
			-0.20239, +0.00273, -0.00347, // 木星
			-0.25486, +0.00276, +0.42926, // 土星
			+0.24588, +0.00345, -14.46266, // 天王星
			-0.95116, +0.02481, +58.30651, // 海王星
		}
		dv := planetCorrections[(xt-1)*3+zn]
		if zn == 0 {
			v += -3 * t / rad
		}
		if zn == 2 {
			v += dv / 1000000
		} else {
			v += dv / rad
		}
	}

	if zn == 0 && xt == 0 {
		if sata != 0 {
			return Limit360(v * 180 / math.Pi)
		}
		return Limit360(v*180/math.Pi + 180)
	}
	if zn == 1 && xt == 0 {
		if sata != 0 {
			return v * 180 / math.Pi
		}
		return -(v * 180 / math.Pi)
	}
	if xt > 0 && zn == 1 {
		return v * 180 / math.Pi
	}
	if xt > 0 && zn == 0 {
		return Limit360(v * 180 / math.Pi)
	}
	return v
}
