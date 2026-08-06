package basic

import (
	"math"
	"testing"
	"time"

	. "github.com/starainrt/astro/tools"
)

func TestTopocentricRaDecUsesUTJulianDateForSiderealTime(t *testing.T) {
	ut := Date2JDE(time.Date(2025, 6, 5, 12, 2, 7, 700000000, time.UTC))
	ra := 189.527817246
	dec := -5.973400893
	lat := 6.79657
	lon := 121.55381
	distanceAU := HMoonAwayN(TD2UT(ut, true), -1) / 149597870.7

	gotRA, gotDec := TopocentricRaDec(ra, dec, lat, lon, ut, distanceAU, 0)
	wantRA, wantDec := independentTopocentricRaDec(ra, dec, lat, lon, ut, distanceAU, 0)
	if delta := angularDistanceArcsec(gotRA, gotDec, wantRA, wantDec); delta > 1e-6 {
		t.Fatalf("TopocentricRaDec differs from independent formula by %.9f arcsec", delta)
	}
}

func TestTopocentricRaAndDecMatchCombinedResult(t *testing.T) {
	ut := Date2JDE(time.Date(2025, 6, 5, 12, 2, 7, 700000000, time.UTC))
	ra := 189.527817246
	dec := -5.973400893
	lat := 6.79657
	lon := 121.55381
	distanceAU := HMoonAwayN(TD2UT(ut, true), -1) / 149597870.7

	wantRA, wantDec := TopocentricRaDec(ra, dec, lat, lon, ut, distanceAU, 0)
	if got := TopocentricRa(ra, dec, lat, lon, ut, distanceAU, 0); got != wantRA {
		t.Fatalf("TopocentricRa = %.12f, want %.12f", got, wantRA)
	}
	if got := TopocentricDec(ra, dec, lat, lon, ut, distanceAU, 0); got != wantDec {
		t.Fatalf("TopocentricDec = %.12f, want %.12f", got, wantDec)
	}
}

func TestHMoonHeightUsesUTForTopocentricCorrection(t *testing.T) {
	ut := Date2JDE(time.Date(2026, 4, 28, 16, 1, 30, 0, time.UTC))
	longitude := 0.0
	latitude := 51.4779
	ra, dec := HMoonApparentRaDecN(ut, longitude, latitude, 0, -1)
	hourAngle := Limit360(ApparentSiderealTime(ut)*15 + longitude - ra)
	want := ArcSin(Sin(latitude)*Sin(dec) + Cos(dec)*Cos(latitude)*Cos(hourAngle))
	got := HMoonHeightN(ut, longitude, latitude, 0, -1)
	if difference := math.Abs(got - want); difference > 1e-10 {
		t.Fatalf("HMoonHeightN differs from the UT topocentric position by %.12f degrees", difference)
	}
}

func independentTopocentricRaDec(ra, dec, lat, lon, ut, distanceAU, height float64) (float64, float64) {
	const (
		equatorialRadiusKM = 6378.14
		polarRadiusKM      = 6356.755
	)
	u := math.Atan(polarRadiusKM / equatorialRadiusKM * Tan(lat))
	rhoCos := math.Cos(u) + height/6378140.0*Cos(lat)
	rhoSin := polarRadiusKM/equatorialRadiusKM*math.Sin(u) + height/6378140.0*Sin(lat)
	sinParallax := Sin(0.0024427777777) / distanceAU
	hourAngle := Limit360(ApparentSiderealTime(ut)*15 + lon - ra)
	deltaRA := math.Atan2(
		-rhoCos*sinParallax*Sin(hourAngle),
		Cos(dec)-rhoCos*sinParallax*Cos(hourAngle),
	)
	topRA := ra + deltaRA*180/math.Pi
	topDec := math.Atan2(
		(Sin(dec)-rhoSin*sinParallax)*math.Cos(deltaRA),
		Cos(dec)-rhoCos*sinParallax*Cos(hourAngle),
	) * 180 / math.Pi
	return topRA, topDec
}

func angularDistanceArcsec(ra1, dec1, ra2, dec2 float64) float64 {
	cosDistance := Sin(dec1)*Sin(dec2) + Cos(dec1)*Cos(dec2)*Cos(ra1-ra2)
	return math.Acos(math.Max(-1, math.Min(1, cosDistance))) * 180 / math.Pi * 3600
}
