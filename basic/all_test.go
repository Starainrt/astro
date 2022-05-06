package basic

import (
	"fmt"
	"testing"
)

func Test_All(t *testing.T) {
	show()
}

func Benchmark_All(b *testing.B) {
	for i := 0; i < b.N; i++ {
		show()
	}
}

func show() {
	jde := GetNowJDE() - 1
	ra := HSunSeeRa(jde - 8.0/24.0)
	dec := HSunSeeDec(jde - 8.0/24.0)
	fmt.Printf("当前JDE:%.14f\n", jde)
	fmt.Println("当前太阳黄经：", HSunSeeLo(jde-8.0/24.0))
	fmt.Println("当前太阳赤经：", ra)
	fmt.Println("当前太阳赤纬：", dec)
	fmt.Println("当前太阳星座：", WhichCst(ra, dec, jde))
	fmt.Println("当前黄赤交角：", EclipticObliquity(jde-8.0/24.0, true))
	fmt.Println("当前日出：", JDE2Date(GetSunRiseTime(jde, 115, 32, 8, 1, 10)))
	fmt.Println("当前日落：", JDE2Date(GetSunDownTime(jde, 115, 32, 8, 1, 10)))
	fmt.Println("当前晨影 -6：", JDE2Date(GetAsaTime(jde, 115, 32, 8, -6)))
	fmt.Println("当前晨影 -12：", JDE2Date(GetAsaTime(jde, 115, 32, 8, -12)))
	fmt.Println("当前昏影 -6：", JDE2Date(GetBanTime(jde, 115, 32, 8, -6)))
	fmt.Println("当前昏影 -12：", JDE2Date(GetBanTime(jde, 115, 32, 8, -12)))
	fmt.Print("农历：")
	fmt.Println(GetLunar(2019, 10, 23, 8.0/24.0))
	fmt.Println("当前月出：", JDE2Date(GetMoonRiseTime(jde, 115, 32, 8, 1, 10)))
	fmt.Println("当前月落：", JDE2Date(GetMoonDownTime(jde, 115, 32, 8, 1, 10)))
	fmt.Println("月相：", MoonLight(jde-8.0/24.0))
}
