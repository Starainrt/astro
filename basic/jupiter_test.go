package basic

import (
	"fmt"
	"testing"
)

func TestJupiter(t *testing.T) {
	jde := GetNowJDE() - 6000
	for i := 0.00; i < 20; i++ {
		fmt.Println(jde+i*365, JDE2Date(jde+i*365), JDE2Date(NextJupiterRetrogradeToPrograde(jde+i*365)))
	}
}
