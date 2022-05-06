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
	//compare
}
