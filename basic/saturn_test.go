package basic

import (
	"fmt"
	"testing"
)

func TestSaturn(t *testing.T) {
	jde := GetNowJDE() - 6000
	for i := 0.00; i < 20; i++ {
		fmt.Println(jde+i*365, JDE2Date(jde+i*365), JDE2Date(LastSaturnProgradeToRetrograde(jde+i*365)))
	}
}
