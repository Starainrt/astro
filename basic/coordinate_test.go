package basic

import (
	"fmt"
	"testing"
)

func Test_LoBoRaDec(t *testing.T) {
	jde := 2451545.0
	lo, bo := RaDecToLoBo(jde, 10, 50)
	fmt.Println("LO,BO", lo, bo)
	fmt.Println(LoBoToRaDec(jde, lo, bo))
	lo, bo = RaDecToLoBo(jde, 40, 80)
	fmt.Println("LO,BO", lo, bo)
	fmt.Println(LoBoToRaDec(jde, lo, bo))
	lo, bo = RaDecToLoBo(jde, 90, 50)
	fmt.Println("LO,BO", lo, bo)
	fmt.Println(LoBoToRaDec(jde, lo, bo))
	lo, bo = RaDecToLoBo(jde, 130, 50)
	fmt.Println("LO,BO", lo, bo)
	fmt.Println(LoBoToRaDec(jde, lo, bo))
	lo, bo = RaDecToLoBo(jde, 160, 50)
	fmt.Println("LO,BO", lo, bo)
	fmt.Println(LoBoToRaDec(jde, lo, bo))
	lo, bo = RaDecToLoBo(jde, 180, 50)
	fmt.Println("LO,BO", lo, bo)
	fmt.Println(LoBoToRaDec(jde, lo, bo))
	lo, bo = RaDecToLoBo(jde, 210, 50)
	fmt.Println("LO,BO", lo, bo)
	fmt.Println(LoBoToRaDec(jde, lo, bo))
	lo, bo = RaDecToLoBo(jde, 260, 50)
	fmt.Println("LO,BO", lo, bo)
	fmt.Println(LoBoToRaDec(jde, lo, bo))
	lo, bo = RaDecToLoBo(jde, 270, 50)
	fmt.Println("LO,BO", lo, bo)
	fmt.Println(LoBoToRaDec(jde, lo, bo))
	lo, bo = RaDecToLoBo(jde, 300, 50)
	fmt.Println("LO,BO", lo, bo)
	fmt.Println(LoBoToRaDec(jde, lo, bo))
	lo, bo = RaDecToLoBo(jde, 350, 50)
	fmt.Println("LO,BO", lo, bo)
	fmt.Println(LoBoToRaDec(jde, lo, bo))
	lo, bo = RaDecToLoBo(jde, 0, 50)
	fmt.Println("LO,BO", lo, bo)
	fmt.Println(LoBoToRaDec(jde, lo, bo))
}
