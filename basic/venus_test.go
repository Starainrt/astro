package basic

import (
	"fmt"
	"testing"
)

func TestVenus(t *testing.T) {
	jde := 2.4597161573032406e+06 - 720
	/*
		fmt.Println(JDE2Date(VenusCulminationTime(jde, 115, 8)))
		fmt.Println(JDE2Date(VenusRiseTime(jde, 115, 23, 8, 0, 0)))
		fmt.Println(JDE2Date(VenusDownTime(jde, 115, 23, 8, 0, 0)))
		fmt.Println("----------------")
	*/
	//LastVenusConjunction(2.4596600340162036e+06)
	//fmt.Println(2.4590359532407406e+06, JDE2Date(2.4590359532407406e+06), JDE2Date(NextVenusRetrograde(2.4590359532407406e+06)))
	//fmt.Println(jde)
	///fmt.Println(MarsTrueLoBo(jde))
	//fmt.Println((jde-2451545)/36525, JDE2Date(0.2293425175054224*36525+2451545))

	decSub := func(jde float64, val float64) float64 {
		sub := VenusSunElongation(jde+val) - VenusSunElongation(jde-val)
		if sub > 180 {
			sub -= 360
		}
		if sub < -180 {
			sub += 360
		}
		return sub / (2 * val)
	}
	_ = decSub
	for i := 0.00; i < 1800.0; i += 50 {
		fmt.Println(jde+i, JDE2Date(jde+i), JDE2Date(LastVenusGreatestElongationWest(jde+i)))
		//fmt.Println(decSub(jde+i, 1.0/86400.0))
	}
}
