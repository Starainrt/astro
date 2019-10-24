package basic

import (
	"fmt"
	"testing"
	"time"
)

func Test_JDECALC(t *testing.T) {
	i := 10
	for i > 0 {
		fmt.Printf("%.18f\n", GetNowJDE())
		time.Sleep(time.Second)
		i--
	}
}

func Test_TDUT(t *testing.T) {
	fmt.Printf("%.18f\n", DeltaT(2119.5, false))
}

func Test_JDE2(t *testing.T) {
	fmt.Println(JDE2Date(2458868.500))
}
