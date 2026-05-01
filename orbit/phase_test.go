package orbit

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

func TestDistancePhaseAndElongationHelpers(t *testing.T) {
	elements := sampleObservationElements()
	date := time.Date(2025, 11, 21, 20, 0, 0, 0, time.UTC)

	sunDistance := SunDistance(date, elements)
	earthDistance := EarthDistance(date, elements)
	phaseAngle := PhaseAngle(date, elements)
	illuminatedFraction := IlluminatedFraction(date, elements)
	phase := Phase(date, elements)
	elongation := Elongation(date, elements)

	for name, value := range map[string]float64{
		"sunDistance":         sunDistance,
		"earthDistance":       earthDistance,
		"phaseAngle":          phaseAngle,
		"illuminatedFraction": illuminatedFraction,
		"phase":               phase,
		"elongation":          elongation,
	} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			t.Fatalf("%s is not finite: %.18f", name, value)
		}
	}

	if math.Abs(sunDistance-HeliocentricEclipticJ2000(date, elements).Distance) > 1e-12 {
		t.Fatalf("sun distance mismatch: got %.15f want %.15f", sunDistance, HeliocentricEclipticJ2000(date, elements).Distance)
	}
	if math.Abs(earthDistance-GeocentricEclipticJ2000(date, elements).Distance) > 1e-12 {
		t.Fatalf("earth distance mismatch: got %.15f want %.15f", earthDistance, GeocentricEclipticJ2000(date, elements).Distance)
	}
	if phaseAngle < 0 || phaseAngle > 180 {
		t.Fatalf("phase angle out of range: %.12f", phaseAngle)
	}
	if elongation < 0 || elongation > 180 {
		t.Fatalf("elongation out of range: %.12f", elongation)
	}
	if illuminatedFraction < 0 || illuminatedFraction > 1 {
		t.Fatalf("illuminated fraction out of range: %.12f", illuminatedFraction)
	}
	if math.Abs(phase-illuminatedFraction) > 1e-12 {
		t.Fatalf("phase alias mismatch: phase=%.15f illuminated=%.15f", phase, illuminatedFraction)
	}

	wantIlluminatedFraction := (1 + math.Cos(phaseAngle*math.Pi/180)) / 2
	if math.Abs(illuminatedFraction-wantIlluminatedFraction) > 1e-12 {
		t.Fatalf("illuminated fraction mismatch: got %.15f want %.15f", illuminatedFraction, wantIlluminatedFraction)
	}

	jd := ttJulianDay(date)
	wantPhaseAngle := math.Acos(clampUnitLocal((sunDistance*sunDistance+earthDistance*earthDistance-basic.EarthAway(jd)*basic.EarthAway(jd))/(2*sunDistance*earthDistance))) * 180 / math.Pi
	if math.Abs(phaseAngle-wantPhaseAngle) > 1e-12 {
		t.Fatalf("phase angle mismatch: got %.15f want %.15f", phaseAngle, wantPhaseAngle)
	}

	objectLon, objectLat, _ := basic.OrbitApparentGeocentricEcliptic(jd, toBasicElements(elements))
	wantElongation := basic.StarAngularSeparation(objectLon, objectLat, basic.HSunApparentLo(jd), basic.HSunTrueBo(jd))
	if math.Abs(elongation-wantElongation) > 1e-12 {
		t.Fatalf("elongation mismatch: got %.15f want %.15f", elongation, wantElongation)
	}
}

func clampUnitLocal(value float64) float64 {
	if value > 1 {
		return 1
	}
	if value < -1 {
		return -1
	}
	return value
}
