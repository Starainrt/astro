package basic

import "math"

func GetLunar(year, month, day int, tz float64) (lyear, lmonth, lday int, leap bool, result string) {
	julianDayEpoch := JDECalc(year, month, float64(day))
	// 确定农历年份
	lyear = year
	adjustedYear := year
	if month == 11 || month == 12 {
		winterSolsticeDay := GetJQTime(year, 270) + tz
		//firstNewMoonDay := TD2UT(CalcMoonS(float64(year)+11.0/12.0+5.0/30.0/12.0, 0), true) + tz
		//nextNewMoonDay := TD2UT(CalcMoonS(float64(year)+1.0, 0), true) + tz
		firstNewMoonDay := TD2UT(CalcMoonSHByJDE(winterSolsticeDay-16, 0), false) + tz
		nextNewMoonDay := TD2UT(CalcMoonSHByJDE(firstNewMoonDay+28, 0), false) + tz

		firstNewMoonDay = normalizeTimePoint(firstNewMoonDay)
		nextNewMoonDay = normalizeTimePoint(nextNewMoonDay)

		if winterSolsticeDay >= firstNewMoonDay && winterSolsticeDay < nextNewMoonDay && julianDayEpoch < firstNewMoonDay {
			adjustedYear--
		}
		if winterSolsticeDay >= nextNewMoonDay && julianDayEpoch < nextNewMoonDay {
			adjustedYear--
		}
	} else {
		adjustedYear--
	}

	// 获取节气和朔望月数据
	solarTerms := GetJieqiLoops(adjustedYear, 25)
	newMoonDays := GetMoonLoops(float64(adjustedYear), 17)

	// 计算冬至日期
	winterSolsticeFirst := normalizeTimePoint(solarTerms[0] - 8.0/24 + tz)
	winterSolsticeSecond := normalizeTimePoint(solarTerms[24] - 8.0/24 + tz)

	// 规范化时间点
	normalizeTimeArray(newMoonDays, tz)
	normalizeTimeArray(solarTerms, tz)

	// 计算朔望月范围
	minMoonIndex, maxMoonIndex := 20, 0
	moonCount := 0
	for i := 0; i < len(newMoonDays)-1; i++ {
		if (newMoonDays[i] <= winterSolsticeFirst && newMoonDays[i+1] > winterSolsticeFirst) ||
			(newMoonDays[i] > winterSolsticeFirst && newMoonDays[i] < winterSolsticeSecond && newMoonDays[i+1] <= winterSolsticeSecond) {
			if i <= minMoonIndex {
				minMoonIndex = i
			}
			if i >= maxMoonIndex {
				maxMoonIndex = i
			}
			moonCount++
		}
	}

	// 确定闰月位置
	leapMonthPos := 20
	if moonCount >= 13 {
		solarTermIndex, i := 0, 0
		for i = minMoonIndex; i <= maxMoonIndex; i++ {
			if !(newMoonDays[i] <= solarTerms[solarTermIndex] && newMoonDays[i+1] > solarTerms[solarTermIndex]) {
				break
			}
			solarTermIndex += 2
		}
		leapMonthPos = i - minMoonIndex
	}

	// 找到当前月相索引
	currentMoonIndex := 0
	for currentMoonIndex = minMoonIndex; currentMoonIndex <= maxMoonIndex; currentMoonIndex++ {
		if newMoonDays[currentMoonIndex] > julianDayEpoch {
			break
		}
	}

	// 计算农历月份
	lmonth = currentMoonIndex - minMoonIndex - 1
	shouldAdjustLeap := false
	leap = false

	if lmonth >= leapMonthPos {
		shouldAdjustLeap = true
	}
	if lmonth == leapMonthPos {
		leap = true
	}
	if lmonth < 2 {
		lmonth += 11
	} else {
		lmonth--
	}
	if shouldAdjustLeap {
		lmonth--
	}
	if lmonth <= 0 {
		lmonth += 12
	}

	// 计算农历日期
	lday = int(julianDayEpoch-newMoonDays[currentMoonIndex-1]) + 1

	// 生成农历日期字符串
	result = formatLunarDateString(lmonth, lday, leap)
	if lmonth >= 10 && month < 3 {
		lyear--
	}
	return
}

func GetSolar(year, month, day int, leap bool, tz float64) float64 {
	adjustedYear := year
	if month < 11 {
		adjustedYear--
	}

	// 获取节气和朔望月数据
	solarTerms := GetJieqiLoops(adjustedYear, 25)
	newMoonDays := GetMoonLoops(float64(adjustedYear), 17)

	// 计算冬至日期
	winterSolsticeFirst := normalizeTimePoint(solarTerms[0] - 8.0/24 + tz)
	winterSolsticeSecond := normalizeTimePoint(solarTerms[24] - 8.0/24 + tz)

	// 规范化时间点
	normalizeTimeArray(newMoonDays, tz)
	normalizeTimeArray(solarTerms, tz)

	// 计算朔望月范围
	minMoonIndex, maxMoonIndex := 20, 0
	moonCount := 0
	for i := 0; i < 15; i++ {
		if (newMoonDays[i] <= winterSolsticeFirst && newMoonDays[i+1] > winterSolsticeFirst) ||
			(newMoonDays[i] > winterSolsticeFirst && newMoonDays[i] < winterSolsticeSecond && newMoonDays[i+1] <= winterSolsticeSecond) {
			if i <= minMoonIndex {
				minMoonIndex = i
			}
			if i >= maxMoonIndex {
				maxMoonIndex = i
			}
			moonCount++
		}
	}

	// 确定闰月位置
	leapMonthPos := 20
	if moonCount >= 13 {
		solarTermIndex, i := 0, 0
		for i = minMoonIndex; i <= maxMoonIndex; i++ {
			if !(newMoonDays[i] <= solarTerms[solarTermIndex] && newMoonDays[i+1] > solarTerms[solarTermIndex]) {
				break
			}
			solarTermIndex += 2
		}
		leapMonthPos = i - minMoonIndex
	}
	actualMonth := month
	if actualMonth > 10 {
		actualMonth -= 11
	} else {
		actualMonth++
	}
	// 计算实际月份索引
	if leap {
		actualMonth++
	}

	if actualMonth >= leapMonthPos && !leap {
		actualMonth++
	}

	return newMoonDays[minMoonIndex+actualMonth] + float64(day) - 1
}

func normalizeTimeArray(timeArray []float64, tz float64) {
	for idx, timeValue := range timeArray {
		adjustedTime := timeValue
		if tz != 8.0/24 {
			adjustedTime = timeValue - 8.0/24 + tz
		}
		timeArray[idx] = normalizeTimePoint(adjustedTime)
	}
}

func normalizeTimePoint(timePoint float64) float64 {
	if timePoint-math.Floor(timePoint) > 0.5 {
		return math.Floor(timePoint) + 0.5
	}
	return math.Floor(timePoint) - 0.5
}

func formatLunarDateString(lunarMonth, lunarDay int, isLeap bool) string {
	monthNames := []string{"十", "一", "二", "三", "四", "五", "六", "七", "八", "九", "十", "冬", "腊"}
	dayPrefixes := []string{"初", "十", "廿", "三"}

	var dateString string

	if isLeap {
		dateString += "闰"
	}

	if lunarMonth == 1 {
		dateString += "正月"
	} else {
		dateString += monthNames[lunarMonth] + "月"
	}

	if lunarDay == 20 {
		dateString += "二十"
	} else if lunarDay == 10 {
		dateString += "初十"
	} else {
		dateString += dayPrefixes[lunarDay/10] + monthNames[lunarDay%10]
	}

	return dateString
}
