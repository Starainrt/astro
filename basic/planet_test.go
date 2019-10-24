package basic

import (
	"fmt"
	"testing"
)

func Test_Ra(t *testing.T) {
	ra, dec := UranusSeeRaDec(2456789.12345)
	fmt.Printf("%.14f\n%.14f\n", ra, dec)
	fmt.Println(UranusMag(2456789.12345))
	ra, dec = NeptuneSeeRaDec(2456789.12345)
	fmt.Printf("%.14f\n%.14f\n", ra, dec)
	fmt.Println(NeptuneMag(2456789.12345))
}
