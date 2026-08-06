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

func TestRaDecByJdeUsesProjectedRightAscensionMotion(t *testing.T) {
	star := InnerStarData{Ra: 10, Dec: 60, PmRA: 3600, PmDec: 0}
	jde := 2451545.0 + 365.2422
	ra, dec := star.RaDecByJde(jde)
	// 投影坐标每年 1 度在 Dec=60 度时对应赤经每年 2 度 / 1 deg/year in the projected coordinate is 2 deg/year in RA at Dec=60.
	wantRA, wantDec := Precess(12, 60, 2451545.0, jde)
	if math.Abs(signedAngleDifference(ra, wantRA)) > 1e-8 || math.Abs(dec-wantDec) > 1e-8 {
		t.Fatalf("position = %.12f %.12f, want %.12f %.12f", ra, dec, wantRA, wantDec)
	}
}
