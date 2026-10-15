package basic

import "math"

// Uranus 北极 J2000 赤经、赤纬与扁率。
const (
	uranusPoleRA      = 257.311
	uranusPoleDec     = -15.175
	uranusFlattening  = 0.02293
	j2000ObliquityDeg = 23.4392911
)

// planetMagnitude2018 用 Mallama & Hilton 2018 模型求视星等；index 为 1..7（水星..海王星）。
// 超出模型相位角范围返回 NaN。
func planetMagnitude2018(index int, jde float64, inst *Instance) float64 {
	r := planetRWithInstance(index, jde, inst)
	delta := planetEarthAwayWithInstance(index, jde, inst)
	earthR := planetRWithInstance(-1, jde, inst)
	if !(r > 0) || !(delta > 0) || !(earthR > 0) {
		return math.NaN()
	}
	cosPhase := (r*r + delta*delta - earthR*earthR) / (2 * r * delta)
	a := math.Acos(math.Max(-1, math.Min(1, cosPhase))) * 180 / math.Pi
	distance := 5 * math.Log10(r*delta)
	switch index {
	case 1:
		return -0.613 + a*(6.3280e-2+a*(-1.6336e-3+a*(3.3644e-5+a*(-3.4265e-7+a*(1.6893e-9-3.0334e-12*a))))) + distance
	case 2:
		if a <= 163.7 {
			return -4.384 + a*(-1.044e-3+a*(3.687e-4+a*(-2.814e-6+8.938e-9*a))) + distance
		}
		return 236.05828 + a*(-2.81914+8.39034e-3*a) + distance
	case 3:
		phys := MarsPhysicalNWithInstance(jde, -1, inst)
		lon := PlanetCoordinateNWithInstance(3, 0, jde, -1, inst)
		return marsMagnitude2018(r, delta, a, phys.SubEarthLongitude, phys.SubSolarLongitude, lon)
	case 4:
		if a <= 12 {
			return -9.395 + a*(-3.7e-4+6.16e-4*a) + distance
		}
		x := a / 180
		albedo := 1 + x*(-1.507+x*(-0.363+x*(-0.062+x*(2.809-1.876*x))))
		if !(albedo > 0) {
			return math.NaN()
		}
		return -9.428 - 2.5*math.Log10(albedo) + distance
	case 5:
		if a > 6.5 {
			return math.NaN()
		}
		earthLat, sunLat, _, _, _, _ := saturnRingParametersWithInstance(jde, inst)
		be, bs := earthLat*rad, sunLat*rad
		tilt := 0.0
		if be*bs > 0 {
			tilt = math.Abs(math.Sin(math.Sqrt(be * bs)))
		}
		return -8.914 - 1.825*tilt + 0.026*a - 0.378*tilt*math.Exp(-2.25*a) + distance
	case 6:
		phase := 0.0
		if a > 3.1 {
			phase = a * (6.587e-3 + 1.045e-4*a)
		}
		return -7.110 - 0.00084*uranusEffectiveSubLatitude(jde, inst) + phase + distance
	case 7:
		year := 2000 + (jde-2451545)/365.25
		unit := -7.00
		switch {
		case year < 1980:
			unit = -6.89
		case year <= 2000:
			unit = -6.89 - 0.0054*(year-1980)
		}
		if a > 1.9 && year > 2000 {
			unit += a * (7.944e-3 + 9.617e-5*a)
		}
		return unit + distance
	}
	return math.NaN()
}

// uranusEffectiveSubLatitude 天王星星下点与日下点行星面纬度绝对值的平均（度）。
func uranusEffectiveSubLatitude(jde float64, inst *Instance) float64 {
	pole := eclipticUnitFromEquatorial(uranusPoleRA, uranusPoleDec)
	geoLon, geoLat := planetApparentLoBoWithInstance(6, jde, inst)
	helioLon, helioLat := planetLWithInstance(6, jde, inst), planetBWithInstance(6, jde, inst)
	sub := func(lon, lat float64) float64 {
		d := sphericalUnit(lon, lat)
		// 星下点方向自行星指向观测者，与视方向相反。
		centric := math.Asin(math.Max(-1, math.Min(1, -(pole[0]*d[0]+pole[1]*d[1]+pole[2]*d[2]))))
		graphic := math.Atan(math.Tan(centric) / ((1 - uranusFlattening) * (1 - uranusFlattening)))
		return math.Abs(graphic) * 180 / math.Pi
	}
	return (sub(geoLon, geoLat) + sub(helioLon, helioLat)) / 2
}

func sphericalUnit(lonDeg, latDeg float64) [3]float64 {
	lon, lat := lonDeg*rad, latDeg*rad
	return [3]float64{math.Cos(lat) * math.Cos(lon), math.Cos(lat) * math.Sin(lon), math.Sin(lat)}
}

func eclipticUnitFromEquatorial(raDeg, decDeg float64) [3]float64 {
	v := sphericalUnit(raDeg, decDeg)
	s, c := math.Sincos(j2000ObliquityDeg * rad)
	return [3]float64{v[0], v[1]*c + v[2]*s, -v[1]*s + v[2]*c}
}

// MercuryMag2018WithInstance 水星视星等（Mallama & Hilton 2018）/ Mercury magnitude, Mallama & Hilton 2018.
func MercuryMag2018WithInstance(jde float64, inst *Instance) float64 {
	return planetMagnitude2018(1, jde, inst)
}

// VenusMag2018WithInstance 金星视星等（Mallama & Hilton 2018）/ Venus magnitude, Mallama & Hilton 2018.
func VenusMag2018WithInstance(jde float64, inst *Instance) float64 {
	return planetMagnitude2018(2, jde, inst)
}

// MarsMag2018WithInstance 火星视星等（Mallama & Hilton 2018，含自转与季节修正）/ Mars magnitude, Mallama & Hilton 2018 with rotation and season terms.
func MarsMag2018WithInstance(jde float64, inst *Instance) float64 {
	return planetMagnitude2018(3, jde, inst)
}

// JupiterMag2018WithInstance 木星视星等（Mallama & Hilton 2018）/ Jupiter magnitude, Mallama & Hilton 2018.
func JupiterMag2018WithInstance(jde float64, inst *Instance) float64 {
	return planetMagnitude2018(4, jde, inst)
}

// SaturnMag2018WithInstance 土星视星等（Mallama & Hilton 2018，含环）/ Saturn magnitude with rings, Mallama & Hilton 2018.
func SaturnMag2018WithInstance(jde float64, inst *Instance) float64 {
	return planetMagnitude2018(5, jde, inst)
}

// UranusMag2018WithInstance 天王星视星等（Mallama & Hilton 2018）/ Uranus magnitude, Mallama & Hilton 2018.
func UranusMag2018WithInstance(jde float64, inst *Instance) float64 {
	return planetMagnitude2018(6, jde, inst)
}

// NeptuneMag2018WithInstance 海王星视星等（Mallama & Hilton 2018）/ Neptune magnitude, Mallama & Hilton 2018.
func NeptuneMag2018WithInstance(jde float64, inst *Instance) float64 {
	return planetMagnitude2018(7, jde, inst)
}

// 火星自转（有效中央经线，0°..390° 每 10°，首两项为 −20°、−10°）与季节（Ls）星等修正表。
var (
	marsRotationCorrection = [40]float64{
		0.024, 0.034, 0.036, 0.045, 0.038, 0.023, 0.015, 0.011,
		0.000, -0.012, -0.018, -0.036, -0.044, -0.059, -0.060, -0.055,
		-0.043, -0.041, -0.041, -0.036, -0.036, -0.018, -0.038, -0.011,
		0.002, 0.004, 0.018, 0.019, 0.035, 0.050, 0.035, 0.027,
		0.037, 0.048, 0.025, 0.022, 0.024, 0.034, 0.036, 0.045,
	}
	marsSeasonCorrection = [40]float64{
		-0.030, -0.017, -0.029, -0.017, -0.014, -0.006, -0.018, -0.020,
		-0.014, -0.030, -0.008, -0.040, -0.024, -0.037, -0.036, -0.032,
		0.010, 0.010, -0.001, 0.044, 0.025, -0.004, -0.016, -0.008,
		0.029, -0.054, -0.033, 0.055, 0.017, 0.052, 0.006, 0.087,
		0.006, 0.064, 0.030, 0.019, -0.030, -0.017, -0.029, -0.017,
	}
)

// marsStirling 四阶 Stirling 插值；angle 取 [0, 360)，表中 angle=10k 对应下标 k+2。
func marsStirling(table *[40]float64, angle float64) float64 {
	angle = math.Mod(angle, 360)
	if angle < 0 {
		angle += 360
	}
	zero := int(angle / 10)
	p := angle/10 - float64(zero)
	var d1 [3]float64
	for i := range d1 {
		d1[i] = table[zero+i+1] - table[zero+i]
	}
	d2 := [2]float64{d1[1] - d1[0], d1[2] - d1[1]}
	// 四阶差分需要第 4 个一阶差分，表首尾各补两项即为此预留。
	d1b := table[zero+4] - table[zero+3]
	d2c := d1b - d1[2]
	d3 := [2]float64{d2[1] - d2[0], d2c - d2[1]}
	d4 := d3[1] - d3[0]
	a4 := d4 / 24
	a3 := (d3[0] + d3[1]) / 12
	a2 := d2[1]/2 - a4
	a1 := (d1[1]+d1[2])/2 - a3
	a0 := table[zero+2]
	return a0 + p*(a1+p*(a2+p*(a3+p*a4)))
}

// marsMagnitude2018 火星视星等：r、delta 为 AU，alpha 为相位角（度），子地、子日经度为西经（度），helioLon 为日心黄经（度）。
func marsMagnitude2018(r, delta, alpha, subEarthLon, subSunLon, helioLon float64) float64 {
	distance := 5 * math.Log10(r*delta)
	base := -0.367 + alpha*(-0.02573+0.0003445*alpha)
	if alpha <= 50 {
		base = -1.601 + alpha*(0.02267-0.0001302*alpha)
	}
	cm := (subEarthLon + subSunLon) / 2
	if math.Abs(subEarthLon-subSunLon) > 180 {
		cm += 180
	}
	// Ls 由日心黄经平移 −85° 得到。
	return base + distance + marsStirling(&marsRotationCorrection, cm) + marsStirling(&marsSeasonCorrection, helioLon-85)
}
