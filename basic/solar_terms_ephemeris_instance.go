package basic

import (
	"math"
)

func solarLongitudeAtTargetWithInstance(jde, target float64, inst *Instance) float64 {
	return wrapSolarLongitude(HSunApparentLoWithInstance(jde, inst), target)
}

func getJQTimeWithInstance(year, angle int, inst *Instance) float64 {
	target := solarTermTarget(angle)
	tt, ok := solveSolarTermTT(solarTermSeed(year, angle), target, func(jde float64) float64 {
		return solarLongitudeAtTargetWithInstance(jde, target, inst)
	})
	if !ok {
		return math.NaN()
	}
	return TT2UTC(tt)
}

// JQLospecWithInstance 指定实例的节气黄经残差 / solar longitude residual with an explicit ephemeris instance.
func JQLospecWithInstance(jde, target float64, inst *Instance) float64 {
	return solarLongitudeAtTargetWithInstance(jde, target, inst)
}

// GetJQTimeWithInstance 指定实例的节气时刻 / solar-term instant with an explicit ephemeris instance.
func GetJQTimeWithInstance(year, angle int, inst *Instance) float64 {
	if inst == nil {
		return GetJQTime(year, angle)
	}
	return getJQTimeWithInstance(year, angle, inst)
}

// GetJieqiLoopsWithInstance 指定实例的连续节气时刻 / consecutive solar-term instants with an explicit ephemeris instance.
func GetJieqiLoopsWithInstance(year, loop int, inst *Instance) []float64 {
	if inst == nil {
		return GetJieqiLoops(year, loop)
	}
	if loop <= 0 {
		return nil
	}
	result := make([]float64, loop)
	for i := 1; i <= loop; i++ {
		angle := 270 + 15*(i-1)
		for angle > 360 {
			angle -= 360
		}
		result[i-1] = getJQTimeWithInstance(year+int(math.Ceil(float64(i-1)/24)), angle, inst) + 8.0/24
	}
	return result
}

// GetMoonLoopsWithInstance 指定实例的连续朔时刻 / consecutive new-moon instants with an explicit ephemeris instance.
func GetMoonLoopsWithInstance(year float64, loop int, inst *Instance) []float64 {
	if inst == nil {
		return GetMoonLoops(year, loop)
	}
	if loop <= 0 {
		return nil
	}
	var start float64
	if year < 6000 {
		start = year + 11.00/12.00 + 5.00/30.00/12.00
	} else {
		start = year + 9.00/12.00 + 5.00/30.00/12.00
	}
	result := make([]float64, loop)
	var last float64
	for i, j := 1, 0; j < loop; i++ {
		var newMoon float64
		if year > 3000 {
			newMoon = TT2UTC(CalcMoonSHWithInstance(start+float64(i-1)/12.5, 0, inst) + 8.0/24)
		} else {
			newMoon = TT2UTC(CalcMoonSWithInstance(start+float64(i-1)/12.5, 0, inst) + 8.0/24)
		}
		if i != 1 && newMoon == last {
			continue
		}
		result[j] = newMoon
		last = newMoon
		j++
	}
	return result
}

// GetShiErCiWithInstance 指定实例的十二次 / twelve ecliptic divisions with an explicit ephemeris instance.
func GetShiErCiWithInstance(jde float64, inst *Instance) string {
	if inst == nil {
		return GetShiErCi(jde)
	}
	tlo := HSunApparentLoWithInstance(jde, inst)
	switch {
	case tlo >= 255 && tlo < 285:
		return "星纪"
	case tlo >= 285 && tlo < 315:
		return "玄枵"
	case tlo >= 315 && tlo < 345:
		return "娵訾"
	case tlo >= 345 || tlo < 15:
		return "降娄"
	case tlo >= 15 && tlo < 45:
		return "大梁"
	case tlo >= 45 && tlo < 75:
		return "实沈"
	case tlo >= 75 && tlo < 105:
		return "鹑首"
	case tlo >= 105 && tlo < 135:
		return "鹑火"
	case tlo >= 135 && tlo < 165:
		return "鹑尾"
	case tlo >= 165 && tlo < 195:
		return "寿星"
	case tlo >= 195 && tlo < 225:
		return "大火"
	case tlo >= 225 && tlo < 255:
		return "析木"
	default:
		return ""
	}
}

// GetWuHouTimeWithInstance 指定实例的七十二候时刻 / pentad marker instant with an explicit ephemeris instance.
func GetWuHouTimeWithInstance(year, angle int, inst *Instance) float64 {
	if inst == nil {
		return GetWuHouTime(year, angle)
	}
	tmp := angle
	angle = angle / 15 * 15
	day := 3
	if angle%2 == 0 {
		day = 18
	}
	timePoint := float64(angle+15) / 30
	if angle%10 == 0 {
		timePoint = float64(angle) / 30
	}
	month := int(3 + timePoint)
	if month > 12 {
		month -= 12
	}
	jde := JDCalc(year, month, float64(day)) + float64(tmp-angle)
	if angle <= 5 {
		angle += 360
	}
	return TT2UTC(refineSolarTermWithInstance(jde, float64(angle), inst))
}

func refineSolarTermWithInstance(jde, angle float64, inst *Instance) float64 {
	result, ok := eventNewtonRefine(jde, 0.00001, func(previousJD float64) float64 {
		value := solarLongitudeAtTargetWithInstance(previousJD, angle, inst) - angle
		derivative := (solarLongitudeAtTargetWithInstance(previousJD+0.000005, angle, inst) - solarLongitudeAtTargetWithInstance(previousJD-0.000005, angle, inst)) / 0.00001
		return value / derivative
	})
	if !ok {
		return math.NaN()
	}
	return result
}
