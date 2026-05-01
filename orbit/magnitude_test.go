package orbit

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

func TestAsteroidMagnitudeHGMatchesDirectFormula(t *testing.T) {
	elements := sampleObservationElements()
	date := time.Date(2025, 11, 21, 20, 0, 0, 0, time.UTC)
	absoluteMagnitude := 3.34
	slopeParameter := 0.12

	got := AsteroidMagnitudeHG(date, elements, absoluteMagnitude, slopeParameter)
	if math.IsNaN(got) || math.IsInf(got, 0) {
		t.Fatalf("magnitude should be finite: %.18f", got)
	}

	sunDistance := SunDistance(date, elements)
	earthDistance := EarthDistance(date, elements)
	phaseAngle := PhaseAngle(date, elements)
	want := absoluteMagnitude + 5*math.Log10(sunDistance*earthDistance) - 2.5*math.Log10(hgSlopeBlendTest(phaseAngle, slopeParameter))
	if math.Abs(got-want) > 1e-12 {
		t.Fatalf("H-G magnitude mismatch: got %.15f want %.15f", got, want)
	}
}

func TestAsteroidMagnitudeHGShiftsWithAbsoluteMagnitude(t *testing.T) {
	elements := sampleObservationElements()
	date := time.Date(2025, 11, 21, 20, 0, 0, 0, time.UTC)

	baseMagnitude := AsteroidMagnitudeHG(date, elements, 3.34, 0.12)
	shiftedMagnitude := AsteroidMagnitudeHG(date, elements, 4.34, 0.12)
	if math.Abs((shiftedMagnitude-baseMagnitude)-1) > 1e-12 {
		t.Fatalf("H shift should move magnitude by exactly 1: base=%.15f shifted=%.15f", baseMagnitude, shiftedMagnitude)
	}
}

func TestAsteroidMagnitudeHGInvalidInputReturnsNaN(t *testing.T) {
	elements := sampleObservationElements()
	date := time.Date(2025, 11, 21, 20, 0, 0, 0, time.UTC)

	if !math.IsNaN(AsteroidMagnitudeHG(date, elements, math.NaN(), 0.12)) {
		t.Fatalf("NaN H should produce NaN result")
	}
	if !math.IsNaN(basic.OrbitAsteroidMagnitudeHG(math.NaN(), toBasicElements(elements), 3.34, 0.12)) {
		t.Fatalf("NaN jd should produce NaN result")
	}
}

func hgSlopeBlendTest(phaseAngle, slopeParameter float64) float64 {
	tanHalf := math.Tan((phaseAngle * math.Pi / 180) / 2)
	phi1 := math.Exp(-3.33 * math.Pow(tanHalf, 0.63))
	phi2 := math.Exp(-1.87 * math.Pow(tanHalf, 1.22))
	return (1-slopeParameter)*phi1 + slopeParameter*phi2
}
