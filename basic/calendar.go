package basic

import (
	"fmt"
	"math"
	"strings"
	"time"
)

/*
 @name: 儒略日计算
 @dec: 计算给定时间的儒略日，1582年改力后为格里高利历，之前为儒略历
 @     请注意，传入的时间在天文计算中一般为力学时，应当注意和世界时的转化
*/
func JDECalc(Year, Month int, Day float64) float64 {
	if Month == 1 || Month == 2 {
		Year--
		Month += 12
	}
	var tmpvarB int
	tmpvar := fmt.Sprintf("%04d-%02d-%2d", Year, Month, int(math.Floor(Day)))
	if strings.Compare(tmpvar, `1582-10-04`) != 1 {
		tmpvarB = 0
	} else {
		tmpvarA := int(Year / 100)
		tmpvarB = 2 - tmpvarA + int(tmpvarA/4)
	}
	return (math.Floor(365.25*(float64(Year)+4716.0)) + math.Floor(30.6001*float64(Month+1)) + Day + float64(tmpvarB) - 1524.5)
}

/*
 @name: 获得当前儒略日时间：当地世界时，非格林尼治时间
*/
func GetNowJDE() (NowJDE float64) {
	Time := float64(time.Now().Second())/3600.0/24.0 + float64(time.Now().Minute())/60.0/24.0 + float64(time.Now().Hour())/24.0
	NowJDE = JDECalc(time.Now().Year(), int(time.Now().Month()), float64(time.Now().Day())+Time)
	return
}

func dt_ext(y, jsd float64) float64 { // 二次曲线外推，用于数值外插
	dy := (y - 1820) / 100.00
	return -20 + jsd*dy*dy
}
func dt_cal(y float64) float64 { //传入年， 返回世界时UT与原子时（力学时 TD）之差, ΔT = TD - UT
	dt_at := []float64{
		-4000, 108371.7, -13036.80, 392.000, 0.0000,
		-500, 17201.0, -627.82, 16.170, -0.3413,
		-150, 12200.6, -346.41, 5.403, -0.1593,
		150, 9113.8, -328.13, -1.647, 0.0377,
		500, 5707.5, -391.41, 0.915, 0.3145,
		900, 2203.4, -283.45, 13.034, -0.1778,
		1300, 490.1, -57.35, 2.085, -0.0072,
		1600, 120.0, -9.81, -1.532, 0.1403,
		1700, 10.2, -0.91, 0.510, -0.0370,
		1800, 13.4, -0.72, 0.202, -0.0193,
		1830, 7.8, -1.81, 0.416, -0.0247,
		1860, 8.3, -0.13, -0.406, 0.0292,
		1880, -5.4, 0.32, -0.183, 0.0173,
		1900, -2.3, 2.06, 0.169, -0.0135,
		1920, 21.2, 1.69, -0.304, 0.0167,
		1940, 24.2, 1.22, -0.064, 0.0031,
		1960, 33.2, 0.51, 0.231, -0.0109,
		1980, 51.0, 1.29, -0.026, 0.0032,
		2000, 63.87, 0.1, 0, 0,
		2005, 64.7, 0.4, 0, 0, //一次项记为x,则 10x=0.4秒/年*(2015-2005),解得x=0.4
		2015, 69,
	}
	y0 := dt_at[len(dt_at)-2] //表中最后一年
	t0 := dt_at[len(dt_at)-1] //表中最后一年的 deltatT
	if y >= y0 {
		jsd := float64(31) // sjd是y1年之后的加速度估计
		// 瑞士星历表jsd=31, NASA网站jsd=32, skmap的jsd=29
		if y > y0+100.00 {
			return dt_ext(y, jsd)
		}
		v := dt_ext(y, jsd)        //二次曲线外推
		dv := dt_ext(y0, jsd) - t0 // ye年的二次外推与te的差
		return (v - dv*(y0+100.00-y)/100.00)
	}
	d := dt_at
	var i int
	for i = 0; i < len(d); i += 5 {
		if float64(y) < d[i+5] {
			break
			// 判断年所在的区间
		}
	}
	t1 := (y - d[i]) / (d[i+5] - d[i]) * 10.00 //////// 三次插值， 保证精确性
	t2 := t1 * t1
	t3 := t2 * t1
	res := d[i+1] + d[i+2]*t1 + d[i+3]*t2 + d[i+4]*t3
	return (res)
}
func DeltaT(Date float64, IsJDE bool) (Result float64) { //传入年或儒略日，传出为秒
	var Year float64
	if IsJDE {
		dates := JDE2Date(Date)
		Year = float64(dates.Year()) + float64(dates.YearDay())/365.0
	} else {
		Year = Date
	}
	if Year < 2010 {
		Result = dt_cal(Year)
		return
	}
	if Year < 2100 && Year >= 2010 {
		var t = (Year - 2000.0)
		Result = 62.92 + 0.32217*t + 0.005589*t*t
		return
	}
	if Year >= 2100 && Year <= 2150 {
		Result = -20 + 32*(((Year-1820.0)/100.0)*((Year-1820.0)/100.0)) - 0.5628*(2150-Year)
		return
	}
	if Year > 2150 {
		//tmp=(Year-1820)/100;
		//Result= -20 + 32 * tmp*tmp;
		Result = dt_cal(Year)
		return
	}
	return
}
func TD2UT(JDE float64, UT2TD bool) float64 { // true 世界时转力学时CC，false 力学时转世界时VV

	Deltat := DeltaT(JDE, true)
	if UT2TD {
		return JDE + Deltat/3600/24
	} else {
		return JDE - Deltat/3600/24
	}
}

/*
 * @name: JDE转日期，输出为数组
 */
func JDE2Date(JD float64) time.Time {
	JD = JD + 0.5
	Z := float64(int(JD))
	F := JD - Z
	var A, B, Years, Months, Days float64
	if Z < 2299161.0 {
		A = Z
	} else {
		alpha := math.Floor((Z - 1867216.25) / 36524.25)
		A = Z + 1 + alpha - math.Floor(alpha/4)
	}
	B = A + 1524
	C := math.Floor((B - 122.1) / 365.25)
	D := math.Floor(365.25 * C)
	E := math.Floor((B - D) / 30.6001)
	Days = B - D - math.Floor(30.6001*E) + F
	if E < 14 {
		Months = E - 1
	}
	if E == 14 || E == 15 {
		Months = E - 13
	}
	if Months > 2 {
		Years = C - 4716
	}
	if Months == 1 || Months == 2 {
		Years = C - 4715
	}
	tms := (Days - math.Floor(Days)) * 24 * 3600
	Days = math.Floor(Days)
	tz, _ := time.LoadLocation("Local")
	dates := time.Date(int(Years), time.Month(int(Months)), int(Days), 0, 0, 0, 0, tz)
	dates = time.Unix(dates.Unix()+int64(tms), int64((tms-math.Floor(tms))*1000000000))
	return dates
}

func JDE2DateByZone(JD float64, tz *time.Location, byZone bool) time.Time {
	JD = JD + 0.5
	Z := float64(int(JD))
	F := JD - Z
	var A, B, Years, Months, Days float64
	if Z < 2299161.0 {
		A = Z
	} else {
		alpha := math.Floor((Z - 1867216.25) / 36524.25)
		A = Z + 1 + alpha - math.Floor(alpha/4)
	}
	B = A + 1524
	C := math.Floor((B - 122.1) / 365.25)
	D := math.Floor(365.25 * C)
	E := math.Floor((B - D) / 30.6001)
	Days = B - D - math.Floor(30.6001*E) + F
	if E < 14 {
		Months = E - 1
	}
	if E == 14 || E == 15 {
		Months = E - 13
	}
	if Months > 2 {
		Years = C - 4716
	}
	if Months == 1 || Months == 2 {
		Years = C - 4715
	}
	tms := (Days - math.Floor(Days)) * 24 * 3600
	Days = math.Floor(Days)
	if !byZone {
		dates := time.Date(int(Years), time.Month(int(Months)), int(Days), 0, 0, 0, 0, time.UTC)
		return time.Unix(dates.Unix()+int64(tms), int64((tms-math.Floor(tms))*1000000000)).In(tz)
	}
	dates := time.Date(int(Years), time.Month(int(Months)), int(Days), 0, 0, 0, 0, tz)
	return time.Unix(dates.Unix()+int64(tms), int64((tms-math.Floor(tms))*1000000000))
}

func GetLunar(year, month, day int, tz float64) (lmonth, lday int, leap bool, result string) {
	jde := JDECalc(year, month, float64(day)) //计算当前JDE时间
	if month == 11 || month == 12 {           //判断当前日期属于前一年周期还是后一年周期
		//判断方法：当前日期与冬至日所在朔望月的关系
		winterday := GetJQTime(year, 270) + tz                                        //冬至日日期（世界时，北京时间）
		Fday := TD2UT(CalcMoonS(float64(year)+11.0/12.0+5.0/30.0/12.0, 0), true) + tz //朔月（世界时，北京时间）
		Yday := TD2UT(CalcMoonS(float64(year)+1.0, 0), true) + tz                     //下一朔月（世界时，北京时间）
		if Fday-math.Floor(Fday) > 0.5 {
			Fday = math.Floor(Fday) + 0.5
		} else {
			Fday = math.Floor(Fday) - 0.5
		}
		if Yday-math.Floor(Yday) > 0.5 {
			Yday = math.Floor(Yday) + 0.5
		} else {
			Yday = math.Floor(Yday) - 0.5
		}
		if winterday >= Fday && winterday < Yday && jde <= Fday {
			year--
		}
		if winterday >= Yday && jde < Yday {
			year--
		}
	} else {
		year--
	}
	jieqi := GetOneYearJQ(year)           //一年的节气
	moon := GetOneYearMoon(float64(year)) //一年朔月日
	winter1 := jieqi[1]                   //第一年冬至日
	winter2 := jieqi[25]                  //第二年冬至日
	for k, v := range moon {
		if tz != 8.0/24 {
			v = v - 8.0/24 + tz
		}
		if v-math.Floor(v) > 0.5 {
			moon[k] = math.Floor(v) + 0.5
		} else {
			moon[k] = math.Floor(v) - 0.5
		}
	} //置闰月为0点
	for k, v := range jieqi {
		if tz != 8.0/24 {
			v = v - 8.0/24 + tz
		}
		if v-math.Floor(v) > 0.5 {
			jieqi[k] = math.Floor(v) + 0.5
		} else {
			jieqi[k] = math.Floor(v) - 0.5
		}
	} //置节气为0点
	mooncount := 0           //年内朔望月计数
	var min, max int = 20, 0 //最大最小计数
	for i := 1; i < 16; i++ {
		if moon[i] >= winter1 && moon[i] < winter2 {
			if i <= min {
				min = i
			}
			if i >= max {
				max = i
			}
			mooncount++
		}
	}
	leapmonth := 20
	if mooncount == 13 { //存在闰月
		j, i := 3, 1
		for i = min; i <= max; i++ {
			if !(moon[i] <= jieqi[j] && moon[i+1] > jieqi[j]) {
				break
			}
			j += 2
		}
		leapmonth = i - min + 1
	}
	i := 0
	for i = min - 1; i <= max; i++ {
		if moon[i] > jde {
			break
		}
	}
	lmonth = i - min
	var sleap bool = false
	leap = false
	if lmonth >= leapmonth {
		sleap = true
	}
	if lmonth == leapmonth {
		leap = true
	}
	if lmonth < 2 {
		lmonth += 11
	} else {
		lmonth--
	}
	if sleap {
		lmonth--
	}
	if lmonth <= 0 {
		lmonth += 12
	}
	lday = int(jde-moon[i-1]) + 1
	strmonth := []string{"十", "一", "二", "三", "四", "五", "六", "七", "八", "九", "十", "冬", "腊"}
	strday := []string{"初", "十", "廿", "三"}
	if leap {
		result += "闰"
	}
	if lmonth == 1 {
		result += "正月"
	} else {
		result += strmonth[lmonth] + "月"
	}
	if lday == 20 {
		result += "二十"
	} else if lday == 10 {
		result += "初十"
	} else {
		result += strday[lday/10] + strmonth[lday%10]
	}
	return
}

func GetSolar(year, month, day int, leap bool, tz float64) float64 {
	if month < 11 {
		year--
	}
	jieqi := GetOneYearJQ(year)           //一年的节气
	moon := GetOneYearMoon(float64(year)) //一年朔月日
	winter1 := jieqi[1]                   //第一年冬至日
	winter2 := jieqi[25]                  //第二年冬至日
	for k, v := range moon {
		if tz != 8.0/24 {
			v = v - 8.0/24 + tz
		}
		if v-math.Floor(v) > 0.5 {
			moon[k] = math.Floor(v) + 0.5
		} else {
			moon[k] = math.Floor(v) - 0.5
		}
	} //置闰月为0点
	for k, v := range jieqi {
		if tz != 8.0/24 {
			v = v - 8.0/24 + tz
		}
		if v-math.Floor(v) > 0.5 {
			jieqi[k] = math.Floor(v) + 0.5
		} else {
			jieqi[k] = math.Floor(v) - 0.5
		}
	} //置节气为0点
	mooncount := 0           //年内朔望月计数
	var min, max int = 20, 0 //最大最小计数
	for i := 1; i < 16; i++ {
		if moon[i] >= winter1 && moon[i] < winter2 {
			if i <= min {
				min = i
			}
			if i >= max {
				max = i
			}
			mooncount++
		}
	}
	leapmonth := 20
	if mooncount == 13 { //存在闰月
		j, i := 3, 1
		for i = min; i <= max; i++ {
			if !(moon[i] <= jieqi[j] && moon[i+1] > jieqi[j]) {
				break
			}
			j += 2
		}
		leapmonth = i - min + 1
	}
	if leap {
		month++
	}
	if month > 10 {
		month -= 11
	} else {
		month++
	}
	if month >= leapmonth && !leap {
		month++
	}
	jde := moon[min-1+month] + float64(day) - 1
	return jde
}

// Date2JDE 日期转儒略日
func Date2JDE(date time.Time) float64 {
	day := float64(date.Day()) + float64(date.Hour())/24.0 + float64(date.Minute())/24.0/60.0 + float64(date.Second())/24.0/3600.0 + float64(date.Nanosecond())/1000000000.0/3600.0/24.0
	return JDECalc(date.Year(), int(date.Month()), day)
}
