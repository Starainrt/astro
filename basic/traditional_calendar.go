package basic

import "math"

// GetShiErCi 十二次 / twelve divisions of the ecliptic in traditional Chinese astronomy.
func GetShiErCi(jd float64) string { //十二次
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

// GetWuHouTime 七十二候时刻 / time of a pentad marker in traditional Chinese calendrics.
func GetWuHouTime(Year, Angle int) float64 {
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
		stDegree := JQLospec(JD0, float64(Angle)) - float64(Angle)
		stDegreep := (JQLospec(JD0+0.000005, float64(Angle)) - JQLospec(JD0-0.000005, float64(Angle))) / 0.00001
		JD1 = JD0 - stDegree/stDegreep
		if math.Abs(JD1-JD0) <= 0.00001 {
			break
		}
	}
	return TD2UT(JD1, false)
}

// GetGanZhi 年干支 / ganzhi for a year number.
func GetGanZhi(year int) string {
	tiangan := []string{"庚", "辛", "壬", "癸", "甲", "乙", "丙", "丁", "戊", "己"}
	dizhi := []string{"申", "酉", "戌", "亥", "子", "丑", "寅", "卯", "辰", "巳", "午", "未"}
	t := year - (year / 10 * 10)
	if t < 0 {
		t += 10
	}
	d := year % 12
	if d < 0 {
		d += 12
	}
	return tiangan[t] + dizhi[d]
}
