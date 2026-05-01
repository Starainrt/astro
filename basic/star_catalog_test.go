package basic

import (
	. "github.com/starainrt/astro/tools"
	"math"
	"testing"
)

func Test_ParseStar(t *testing.T) {
	if err := LoadStarData(); err != nil {
		t.Fatal(err)
	}
	for _, v := range stardat {
		if _, err := parseStarData(v); err != nil {
			t.Fatal(err)
		}
	}
}

func TestGetStarByChniese(t *testing.T) {
	if err := LoadStarData(); err != nil {
		t.Fatal(err)
	}

	sirius, err := StarDataByChinese("天狼星")
	if err != nil {
		t.Fatal(err)
	}
	if sirius.HIP != 32349 || sirius.HR != 2491 {
		t.Fatal("cannot found star")
	}

	sirius, err = StarDataByChinese("天狼")
	if err != nil {
		t.Fatal(err)
	}
	if sirius.HIP != 32349 || sirius.HR != 2491 {
		t.Fatal("cannot found star")
	}
}

func TestGetRaDecByDate(t *testing.T) {
	if err := LoadStarData(); err != nil {
		t.Fatal(err)
	}

	sirius, err := StarDataByHR(2491)
	if err != nil {
		t.Fatal(err)
	}

	ra, dec := sirius.RaDecByJde(2451545.0)
	if math.IsNaN(ra) || math.IsInf(ra, 0) || ra < 0 || ra >= 360 {
		t.Fatalf("invalid ra: %.12f", ra)
	}
	if math.IsNaN(dec) || math.IsInf(dec, 0) || dec < -90 || dec > 90 {
		t.Fatalf("invalid dec: %.12f", dec)
	}
	if Format(sirius.Ra/15, 1) == "" || Format(sirius.Dec, 0) == "" {
		t.Fatal("unexpected empty formatted catalog coordinates")
	}
}
