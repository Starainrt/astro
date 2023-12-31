package basic

import (
	. "github.com/starainrt/astro/tools"
	"fmt"
	"testing"
)

func Test_ParseStar(t *testing.T) {
	//dat := []byte(`2491  9Alp CMaBD-16 1591  48915151881 257I   5423           064044.6-163444064508.9-164258227.22-08.88-1.46   0.00 -0.05 -0.03   A1Vm               -0.553-1.205 +.375-008SBO    13 10.3  11.2AB   4*`)
	err := LoadStarData()
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range stardat {
		_, err = parseStarData(v)
		if err != nil {
			t.Fatal(err)
		}
	}

}
func TestGetStarByChniese(t *testing.T) {
	err := LoadStarData()
	if err != nil {
		t.Fatal(err)
	}
	sirius, err := StarDataByChinese("天狼星")
	if err != nil {
		t.Fatal(err)
	}
	if sirius.HIP != 32349 || sirius.HR != 2491 {
		t.Fatal("cannot found star")
	}
	fmt.Printf("%+v\n", sirius)
	sirius, err = StarDataByChinese("天狼")
	if err != nil {
		t.Fatal(err)
	}
	if sirius.HIP != 32349 || sirius.HR != 2491 {
		t.Fatal("cannot found star")
	}
}

func TestGetRaDecByDate(t *testing.T) {
	err := LoadStarData()
	if err != nil {
		t.Fatal(err)
	}
	sirius, err := StarDataByHR(2491)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", sirius)
	fmt.Println(Format(sirius.Ra/15, 1), Format(sirius.Dec, 0))
	now := GetNowJDE()
	ra, dec := sirius.RaDecByJde(now)
	fmt.Println(Format(ra/15, 1), Format(dec, 0))
}
