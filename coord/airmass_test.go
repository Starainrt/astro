package coord

import (
	"math"
	"testing"

	"github.com/starainrt/astro/formula"
)

func TestAirmassWrappers(t *testing.T) {
	assertClose(t, "plane-parallel wrapper", AirmassPlaneParallelFromTrueAltitude(30), formula.AirmassPlaneParallel(30), 1e-15)
	assertClose(t, "kasten-young apparent wrapper", AirmassKastenYoungFromApparentAltitude(30), formula.AirmassKastenYoung(30), 1e-15)
	assertClose(t, "pickering apparent wrapper", AirmassPickeringFromApparentAltitude(30), formula.AirmassPickering(30), 1e-15)
}

func TestAirmassTrueAltitudeWrappers(t *testing.T) {
	trueAltitude := 5.0
	pressureHPa := 1013.25
	temperatureC := 10.0
	apparentAltitude := ApparentAltitude(trueAltitude, pressureHPa, temperatureC)

	gotKastenYoung := AirmassKastenYoungFromTrueAltitude(trueAltitude, pressureHPa, temperatureC)
	wantKastenYoung := formula.AirmassKastenYoung(apparentAltitude)
	assertClose(t, "kasten-young true-alt wrapper", gotKastenYoung, wantKastenYoung, 1e-12)

	gotPickering := AirmassPickeringFromTrueAltitude(trueAltitude, pressureHPa, temperatureC)
	wantPickering := formula.AirmassPickering(apparentAltitude)
	assertClose(t, "pickering true-alt wrapper", gotPickering, wantPickering, 1e-12)

	if math.IsNaN(gotKastenYoung) || math.IsNaN(gotPickering) {
		t.Fatal("expected finite airmass values for low but above-horizon altitude")
	}
}
