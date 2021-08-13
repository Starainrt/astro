package basic

import (
	"fmt"
	"testing"
)

func Test_Isxz(t *testing.T) {
	now := GetNowJDE()
	for i := 0.00; i <= 360.00; i+=0.5 {
		for j := -90.00; j <= 90.00; j+=0.5 {
			fmt.Println(i, j)
			fmt.Println(WhichCst(float64(i), float64(j), now))
		}
	}
}
