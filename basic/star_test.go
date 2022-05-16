package basic

import (
	"fmt"
	"testing"
)

func Test_StarHeight(t *testing.T) {
	date := GetNowJDE() + 6.0/24.0
	fmt.Println(JDE2Date(date))
	fmt.Println("Sirius Height:", StarHeight(date, 101.5, -16.8, 115, 40, 8.0))
	fmt.Println("Sirius Azimuth:", StarAzimuth(date, 101.5, -16.8, 115, 40, 8.0))

}

func Test_Star(t *testing.T) {
	date := GetNowJDE() - 20/24.0
	fmt.Println(JDE2Date(date))
	fmt.Println("Sirius RiseTime:", JDE2Date(StarRiseTime(date, 101.529, -16.8, 113.568, 22.5, 0, 8.0, true)))
	fmt.Println("Sirius CulminationTime:", JDE2Date(StarCulminationTime(date, 101.529, 113.568, 8.0)))
	fmt.Println("Sirius DownTime:", JDE2Date(StarDownTime(date, 101.529, -16.8, 113.568, 22.5, 0, 8.0, true)))
}

func TestZB(t *testing.T) {
	jde := GetNowJDE()
	fmt.Println(LoBoToRaDec(jde, 156, 0))
}
