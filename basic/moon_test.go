package basic

import (
	"fmt"
	"testing"
	"time"
)

func Benchmark_MoonRiseBench(b *testing.B) {
	jde := GetNowJDE()
	for i := 0; i < b.N; i++ {
		GetMoonRiseTime(jde, 105, 40, 8, 0, 10)
	}

}

func Test_MoonDown(t *testing.T) {
	jde := GetNowJDE()
	for i := 30.0; i < 90.0; i += 0.3 {
		fmt.Println(i, GetMoonDownTime(jde, 115, float64(i), 8, 1, 0))
	}
}

func Test_MoonRise(t *testing.T) {
	//2.459984692085961e+06 113.58880556 87.36833333 8 0 0
	//2.459984692085961e+06 113.58880556 87.36833333 8 0 0
	//2.4599846948519214e+06 113.58880556 87.36833333 8 0 0

	//cst := time.FixedZone("cst", 8*3600)
	//jde := Date2JDE(time.Date(2023, 2, 9, 15, 59, 0, 0, cst))
	fmt.Println(GetMoonRiseTime(2.4599846948519214e+06, 113.58880556, 87.36833333, 8, 0, 0))
	for i := 30.0; i < 90.0; i += 0.3 {
		fmt.Println(i, GetMoonRiseTime(2.459984692085961e+06, 117.76653567, float64(i), 8, 0, 0))
	}
}

func Test_MoonS(t *testing.T) {
	//fmt.Println(Sita(2451547))
	//fmt.Println(MoonHeight(2451547, 115, 32, 8))
	a := time.Now().UnixNano()
	b := GetMoonRiseTime(GetNowJDE(), 123, 40, 8, 0, 10)
	fmt.Println(HMoonHeight(b, 115, 32, 8))
	fmt.Println(time.Now().UnixNano() - a)
	fmt.Println(JDE2Date((b)))
	fmt.Println(time.Now().UnixNano() - a)
	//fmt.Printf("%.14f", GetMoonRiseTime(2451547, 115, 32, 8, 0))
}
