package basic

import (
	"fmt"
	"testing"
)

func TestMars(t *testing.T) {
	jde := GetNowJDE() - 6000
	/*
		fmt.Println(JDE2Date(VenusCulminationTime(jde, 115, 8)))
		fmt.Println(JDE2Date(VenusRiseTime(jde, 115, 23, 8, 0, 0)))
		fmt.Println(JDE2Date(VenusDownTime(jde, 115, 23, 8, 0, 0)))
		fmt.Println("----------------")
	*/
	//LastVenusConjunction(2.4596600340162036e+06)
	//fmt.Println(2.4590359532407406e+06, JDE2Date(2.4590359532407406e+06), JDE2Date(NextVenusRetrograde(2.4590359532407406e+06)))
	for i := 0.00; i < 1; i++ {
		fmt.Println(jde+i*740, JDE2Date(jde+i*740), JDE2Date(LastMarsProgradeToRetrograde(jde+i*740)))
	}
}
