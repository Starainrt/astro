package basic

import (
	"github.com/starainrt/astro/tools"
	"fmt"
	"math"
	"os"
	"testing"
	"time"
)

func Test_Jq(t *testing.T) {
	data := GetJieqiLoops(2019, 24)
	for i := 1; i < 25; i++ {
		fmt.Println(JDE2Date(data[i]))
	}
	//fmt.Println(JDE2Date(GetWHTime(2019, 10)))
	//fmt.Println(JDE2Date(GetJQTime(2020, 0)))
	//date := TD2UT(GetJQTime(2020, 0), true)
	//fmt.Println(HSunApparentLo(date))
}

func TestZD(t *testing.T) {
	jde := 2452982.9872345612
	zd := HJZD(jde)
	fmt.Println(zd)
	if zd != -0.003746747950462434 {
		t.Fatal("not equal")
	}
	zd = JJZD(jde)
	fmt.Println(zd)
	if zd != 0.001513453926274198 {
		t.Fatal("not equal")
	}
}

func Test_SunLo(t *testing.T) {
	fmt.Printf("%.14f\n", HSunTrueLo(2458840.0134162))
	fmt.Printf("%.14f", HSunApparentLo(2458840.0134162))
}

func Test_SunDiff(t *testing.T) {
	n := JDECalc(2000, 1, 1)
	var maxRa, maxDec, maxLo float64
	for i := float64(0); i < 365.2422*30; i++ {
		tLo := HSunApparentLo(n + i)
		tRa, tDec := HSunApparentRaDec(n + i)
		fRa, fDec := SunApparentRaDec(n + i)
		fLo := SunApparentLo(n + i)
		tmp := tools.Limit360(math.Abs(tRa - fRa))
		if tmp > 300 {
			tmp = 360 - tmp
		}
		if tmp > maxRa {
			maxRa = tmp
		}
		tmp = tools.Limit360(math.Abs(tDec - fDec))
		if tmp > 300 {
			tmp = 360 - tmp
		}
		if tmp > maxDec {
			maxDec = tmp
		}

		tmp = tools.Limit360(math.Abs(tLo - fLo))
		if tmp > 300 {
			tmp = 360 - tmp
		}
		if tmp > maxLo {
			maxLo = tmp
		}
	}
	fmt.Printf("%.15f %.15f %.15f\n", maxRa*3600, maxDec*3600, maxLo*3600)
}

func Benchmark_SunRise(b *testing.B) {
	jde := GetNowJDE()
	for i := 0; i < b.N; i++ {
		//GetNowJDE()
		GetSunRiseTime(jde, 115, 32, 8, 0, 10)
	}

}

func Benchmark_SunLo(b *testing.B) {
	jde := GetNowJDE()
	for i := 0; i < b.N; i++ {
		//GetNowJDE()
		HSunApparentLo(jde)
	}

}

func Test_Cal(t *testing.T) {
	fmt.Println(JDE2Date(GetSolar(2020, 1, 1, false, 8.0/24.0)))
	fmt.Println(JDE2Date(GetSolar(2020, 4, 1, false, 8.0/24.0)))
	fmt.Println(JDE2Date(GetSolar(2020, 4, 1, true, 8.0/24.0)))
	fmt.Println(JDE2Date(GetSolar(2033, 11, 3, false, 8.0/24.0)))
	fmt.Println(JDE2Date(GetSolar(2033, 11, 3, true, 8.0/24.0)))
	fmt.Println(JDE2Date(GetSolar(2034, 1, 1, false, 8.0/24.0)))
}

func Test_SunRise(t *testing.T) {
	a := time.Now().UnixNano()
	//b := GetSunRiseTime(GetNowJDE(), 120, 32, 8, 0)
	//b = GetSunRiseTime(GetNowJDE()+1, 145, 32, 8, 0)
	//b = GetSunRiseTime(GetNowJDE()+2, 135, 50, 8, 0)
	//b = GetSunRiseTime(GetNowJDE()+3, 125, 32, 8, 0)
	//b = GetSunRiseTime(GetNowJDE()+4, 75, 32, 8, 0)
	//b = GetSunRiseTime(GetNowJDE()+5, 85, 32, 8, 0)
	//b = GetSunRiseTime(GetNowJDE()+6, 95, 32, 8, 0)
	//b = GetSunRiseTime(GetNowJDE()+7, 105, 32, 8, 0)
	//b = GetSunRiseTime(GetNowJDE()+8, 115, 32, 8, 0)
	b := GetSunRiseTime(GetNowJDE()+9, 125, 32, 8, 0, 10)
	fmt.Println(time.Now().UnixNano() - a)
	fmt.Println(SunHeight(b, 115, 32, 8))
	fmt.Println(JDE2Date((b)))
	fmt.Println(time.Now().UnixNano() - a)
}

func Test_SunTwilightMo(t *testing.T) {
	cst := time.FixedZone("cst", 8*3600)
	jde := Date2JDE(time.Date(2023, 10, 3, 15, 59, 0, 0, cst))
	fmt.Println(GetAsaTime(jde, 113.58880556, 87.66833333, 8, -6))

	for i := 10.0; i < 90.0; i += 0.3 {
		fmt.Println(i, GetAsaTime(jde, 125.45506654, float64(i), 8, -6))
	}
}

func Test_SunTwilightEv(t *testing.T) {
	cst := time.FixedZone("cst", 8*3600)
	jde := Date2JDE(time.Date(2023, 10, 3, 15, 59, 0, 0, cst))
	for i := 10.0; i < 90.0; i += 0.3 {
		fmt.Println(i, GetBanTime(jde, 115, float64(i), 8, -18))
	}
}

func Test_SunRiseRound(t *testing.T) {
	jde := GetNowJDE()
	for i := 10.0; i < 90.0; i += 0.3 {
		fmt.Println(i, GetSunRiseTime(jde, 115, float64(i), 8, 0, 0))
	}
}
func Test_SunDown(t *testing.T) {
	jde := GetNowJDE()
	for i := 10.0; i < 90.0; i += 0.3 {
		fmt.Println(i, GetSunDownTime(jde, 115, float64(i), 8, 0, 0))
	}
}

func Test_SunAz(t *testing.T) {
	cst := time.FixedZone("cst", 8*3600)
	fmt.Println(SunAngle(Date2JDE(time.Date(2022, 5, 30, 11, 55, 0, 0, cst)),
		120, 30, 8.0))
}

func TestJQDate(t *testing.T) {
	trimDay := func(d float64) float64 {
		if d-math.Floor(d) < 0.5 {
			return math.Floor(d) - 0.5
		}
		return math.Floor(d) + 0.5
	}
	c := 0
	var info string
	for year := 1900; year <= 2600; year++ {
		for pos := 0; pos < 360; pos += 15 {
			n := newGetJQTime(year, pos) + 8.0/24.000000
			o := GetJQTime(year, pos) + 8.0/24.0000000
			if trimDay(n) != trimDay(o) {
				c++
				fmt.Printf("\"%d%03d\"=>%v  %v\n", year, pos, JDE2Date(trimDay(o)), JDE2Date(trimDay(n)))
				info += fmt.Sprintf("\"%d%03d\"=>%.0f,", year, pos, trimDay(o)-trimDay(n))
			}
		}
	}
	fmt.Println(c)
	os.WriteFile("test.txt", []byte(info), 0644)
}

func newGetJQTime(Year, Angle int) float64 { //节气时间
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
			stDegree := newJQLospec(JD0) - float64(Angle)
			stDegreep := (newJQLospec(JD0+0.000005) - newJQLospec(JD0-0.000005)) / 0.00001
			JD1 = JD0 - stDegree/stDegreep
			if math.Abs(JD1-JD0) <= 0.00001 {
				break
			}
		}
		JD1 -= 0.001
	}
	JD1 += 0.001
	return JD1 - 0.0046296296296296
}

func newJQLospec(JD float64) float64 {
	t := tools.FloatRound(SunApparentLo(JD), 9)
	if t <= 12 {
		t += 360
	}
	return t
}
