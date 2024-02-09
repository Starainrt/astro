package basic

import (
	"github.com/starainrt/astro/tools"
	"fmt"
	"math"
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
func Test_MoonDiff(t *testing.T) {
	n := JDECalc(2000, 1, 1)
	var maxRa, maxDec, maxLo, maxBo float64
	for i := float64(0); i < 365.2422*3; i++ {
		tLo := HMoonApparentLo(n + i)
		tBo := HMoonTrueBo(n + i)
		tRa, tDec := HMoonTrueRaDec(n + i)
		fRa, fDec := MoonTrueRaDec(n + i)
		fLo := MoonApparentLo(n + i)
		fBo := MoonTrueBo(n + i)
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

		tmp = tools.Limit360(math.Abs(tBo - fBo))
		if tmp > 300 {
			tmp = 360 - tmp
		}
		if tmp > maxBo {
			maxBo = tmp
		}
	}
	fmt.Printf("%.15f %.15f %.15f %.15f\n", maxRa*3600, maxDec*3600, maxLo*3600, maxBo*3600)
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

func TestMoonCu(t *testing.T) {
	jde := math.Floor(GetNowJDE() - 20.0/24.0)
	n := MoonCulminationTime(jde, 115, 23, 8)
	fmt.Println(JDE2Date(n))
	fmt.Println(MoonTimeAngle(n, 115, 23, 8))
	fmt.Println(MoonAngle(n, 115, 23, 8))
	//fmt.Println(JDE2Date(jde))
	//ra, dec := HMoonApparentRaDec(jde, 115, 23, 8)
	//fmt.Println(tools.Format(ra/15, 1), tools.Format(dec, 0))
	//fmt.Println(JDE2Date(GetMoonTZTime(jde, 115, 23, 8)))
	//fmt.Println(JDE2Date(GetMoonDownTime(jde+1, 115, 23, 8, 1, 0)))
}
