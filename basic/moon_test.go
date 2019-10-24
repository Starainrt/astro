package basic

import (
	"fmt"
	"testing"
	"time"
)

func Test_MoonS(t *testing.T) {
	//fmt.Println(Sita(2451547))
	//fmt.Println(MoonHeight(2451547, 115, 32, 8))
	a := time.Now().UnixNano()
	b := GetMoonRiseTime(GetNowJDE(), 115, 32, 8, 0)
	fmt.Println(time.Now().UnixNano() - a)
	fmt.Println(JDE2Date((b)))
	fmt.Println(time.Now().UnixNano() - a)
	//fmt.Printf("%.14f", GetMoonRiseTime(2451547, 115, 32, 8, 0))
}