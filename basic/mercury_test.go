package basic

import (
	"fmt"
	"testing"
)

func TestMercury(t *testing.T) {
	jde := GetNowJDE()
	fmt.Println(2.459941309513889e+06, JDE2Date(2.459941309513889e+06), JDE2Date(8.0/24.0+LastMercuryGreatestElongation(2.459941309513889e+06)))
	fmt.Println("-------------for------------")
	for i := 0.00; i < 700.0; i += 5 {
		fmt.Println(jde+i, JDE2Date(jde+i), JDE2Date(8.0/24.0+LastMercuryGreatestElongationWest(jde+i)))
		// fmt.Println("")
	}
}
