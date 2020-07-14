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
	//fmt.Println(HSunSeeLo(date))
}

func Test_SunLo(t *testing.T) {
	fmt.Printf("%.14f\n", HSunTrueLo(2458840.0134162))
	fmt.Printf("%.14f", HSunSeeLo(2458840.0134162))
}

func Test_Cal(t *testing.T) {
	fmt.Println(JDE2Date(GetSolar(2020, 1, 1, false)))
	fmt.Println(JDE2Date(GetSolar(2020, 4, 1, false)))
	fmt.Println(JDE2Date(GetSolar(2020, 4, 1, true)))
	fmt.Println(JDE2Date(GetSolar(2033, 11, 3, false)))
	fmt.Println(JDE2Date(GetSolar(2033, 11, 3, true)))
	fmt.Println(JDE2Date(GetSolar(2034, 1, 1, false)))
}

func Test_SunRise(t *testing.T) {
	a := time.Now().UnixNano()
	b := GetSunRiseTime(GetNowJDE(), 115, 32, 8, 0)
	b = GetSunRiseTime(GetNowJDE()+1, 115, 32, 8, 0)
	b = GetSunRiseTime(GetNowJDE()+2, 115, 32, 8, 0)
	b = GetSunRiseTime(GetNowJDE()+3, 115, 32, 8, 0)
	b = GetSunRiseTime(GetNowJDE()+4, 115, 32, 8, 0)
	b = GetSunRiseTime(GetNowJDE()+5, 115, 32, 8, 0)
	b = GetSunRiseTime(GetNowJDE()+6, 115, 32, 8, 0)
	b = GetSunRiseTime(GetNowJDE()+7, 115, 32, 8, 0)
	b = GetSunRiseTime(GetNowJDE()+8, 115, 32, 8, 0)
	b = GetSunRiseTime(GetNowJDE()+9, 115, 32, 8, 0)
	fmt.Println(time.Now().UnixNano() - a)
	fmt.Println(JDE2Date((b)))
	fmt.Println(time.Now().UnixNano() - a)
}
