package basic

import (
	"math"

	"github.com/starainrt/astro/internal/ephem"
)

// planetBody 旧模型行星序号 1..7（水星..海王星，不含地球）对应的 NAIF 编号。
var planetBody = [8]ephem.Body{0, ephem.Mercury, ephem.Venus, ephem.Mars, ephem.Jupiter, ephem.Saturn, ephem.Uranus, ephem.Neptune}

// vectorApparent 新模型全精度时的矢量视位置：真黄道经纬（度）、光行时（日）、几何地心距（AU）；不适用时 ok=false。
// 状态非法时与显式实例路径一致：ok=true 且各值为 NaN，不得退回低精度公式掩盖错误。
func vectorApparent(target ephem.Body, jde float64, n int) (lo, bo, tau, distAU float64, ok bool) {
	if n >= 0 || ephem.Active() != ephem.Semianalytic {
		return 0, 0, 0, 0, false
	}
	lo, bo, tau, distAU, ok, err := apparentState(target, jde, ephem.Global())
	if err != nil {
		return math.NaN(), math.NaN(), math.NaN(), math.NaN(), true
	}
	return lo, bo, tau, distAU, ok
}

// icrfToTrueEcliptic ICRF 单位矢量 → 真黄道经纬（度）；黄赤交角与下游 LoBoToRaDec 同口径，往返得到真赤道。
func icrfToTrueEcliptic(d [3]float64, jde float64) (lo, bo float64) {
	m := ltpPBMAT(2000 + (jde-2451545)/365.25)
	var q [3]float64
	for i := range q {
		q[i] = m[i][0]*d[0] + m[i][1]*d[1] + m[i][2]*d[2]
	}
	s, c := math.Sincos(Obliquity1980(jde) * math.Pi / 180)
	x, y, z := q[0], q[1]*c+q[2]*s, -q[1]*s+q[2]*c
	lo = math.Atan2(y, x)*180/math.Pi + Nutation2000Bi(jde)
	bo = math.Atan2(z, math.Hypot(x, y)) * 180 / math.Pi
	return math.Mod(lo+720, 360), bo
}

// sunApparentLoBoN 太阳视黄经与黄纬；有矢量视位置时两者取自同一方向，否则退回视黄经配几何黄纬。
func sunApparentLoBoN(jde float64, n int, inst *Instance) (float64, float64) {
	if lo, bo, _, _, ok, err := vectorApparentWithInstance(ephem.Sun, jde, n, inst); err != nil {
		return math.NaN(), math.NaN()
	} else if ok {
		return lo, bo
	}
	return HSunApparentLoNWithInstance(jde, n, inst), HSunTrueBoNWithInstance(jde, n, inst)
}

// moonApparentLoBoN 月球视黄经与黄纬；有矢量视位置时两者取自同一方向，否则退回视黄经配几何黄纬。
func moonApparentLoBoN(jde float64, n int, inst *Instance) (float64, float64) {
	if lo, bo, _, _, ok, err := vectorApparentWithInstance(ephem.Moon, jde, n, inst); err != nil {
		return math.NaN(), math.NaN()
	} else if ok {
		return lo, bo
	}
	return HMoonApparentLoNWithInstance(jde, n, inst), HMoonTrueBoNWithInstance(jde, n, inst)
}

func sunApparentRaDecN(jde float64, n int, inst *Instance) (float64, float64) {
	lo, bo := sunApparentLoBoN(jde, n, inst)
	return LoBoToRaDec(jde, lo, bo)
}

func sunApparentRaN(jde float64, n int, inst *Instance) float64 {
	ra, _ := sunApparentRaDecN(jde, n, inst)
	return ra
}

func moonApparentRaDecN(jde float64, n int, inst *Instance) (float64, float64) {
	lo, bo := moonApparentLoBoN(jde, n, inst)
	return LoBoToRaDec(jde, lo, bo)
}

func moonApparentRaN(jde float64, n int, inst *Instance) float64 {
	ra, _ := moonApparentRaDecN(jde, n, inst)
	return ra
}
