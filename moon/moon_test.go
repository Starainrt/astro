package moon

import (
	"fmt"
	"testing"
)

func TestMoonI(t *testing.T) {
	fmt.Println(MoonR(2465445.9755443))
}

func Test_NewCalc(t *testing.T) {
	fmt.Printf("%.14f", MoonCalcNew(2, 2451546.0))
}
