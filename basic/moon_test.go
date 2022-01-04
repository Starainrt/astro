package basic

import (
	"fmt"
	"testing"
	"time"
)

func Benchmark_MoonRiseBench(b *testing.B) {
	jde := GetNowJDE()
	for i := 0; i < b.N; i++ {
		GetMoonRiseTime(jde, 115, 32, 8, 0, 10)
	}

}

func Test_MoonS(t *testing.T) {
	//fmt.Println(Sita(2451547))
	//fmt.Println(MoonHeight(2451547, 115, 32, 8))
	a := time.Now().UnixNano()
	b := GetMoonRiseTime(GetNowJDE(), 115, 32, 8, 0, 10)
	fmt.Println(HMoonHeight(b, 115, 32, 8))
	fmt.Println(time.Now().UnixNano() - a)
	fmt.Println(JDE2Date((b)))
	fmt.Println(time.Now().UnixNano() - a)
	//fmt.Printf("%.14f", GetMoonRiseTime(2451547, 115, 32, 8, 0))
}
