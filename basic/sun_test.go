package basic

import (
	"fmt"
	"testing"
	"time"
)

func Test_Jq(t *testing.T) {
	data := GetOneYearJQ(2019)
	for i := 1; i < 25; i++ {
		fmt.Println(JDE2Date(data[i]))
	}
	//fmt.Println(JDE2Date(GetWHTime(2019, 10)))
	//fmt.Println(JDE2Date(GetJQTime(2020, 0)))
	//date := TD2UT(GetJQTime(2020, 0), true)
	//fmt.Println(HSunApparentLo(date))
}

func Test_SunLo(t *testing.T) {
	fmt.Printf("%.14f\n", HSunTrueLo(2458840.0134162))
	fmt.Printf("%.14f", HSunApparentLo(2458840.0134162))
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
