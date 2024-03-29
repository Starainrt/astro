package basic

import (
	"testing"
)

func Test_Isxz(t *testing.T) {
	now := GetNowJDE()
	//finish on 30s
	for i := 0.00; i <= 360.00; i += 0.5 {
		for j := -90.00; j <= 90.00; j += 0.5 {
			WhichCst(float64(i), float64(j), now)
		}
	}
}

func Benchmark_IsXZ(b *testing.B) {
	jde := GetNowJDE()
	for i := 0; i < b.N; i++ {
		//GetNowJDE()
		WhichCst(11.11, 12.12, jde)
	}

}
