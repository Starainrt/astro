package orbit

import (
	"math"
	"testing"
	"time"
)

func TestVisualBinaryMatchesMeeusEtaCoronaeBorealisExample(t *testing.T) {
	elements := VisualBinaryElements{
		PeriodYears:        41.623,
		PeriastronYear:     1934.008,
		Eccentricity:       0.2763,
		SemiMajorAxis:      0.907,
		Inclination:        59.025,
		AscendingNode:      23.717,
		PeriastronArgument: 219.907,
	}

	got := VisualBinaryByYear(1980.0, elements)

	if math.Abs(got.MeanAnomaly-37.788) > 0.001 {
		t.Fatalf("mean anomaly mismatch: got %.6f want %.6f", got.MeanAnomaly, 37.788)
	}
	if math.Abs(got.EccentricAnomaly-49.897) > 0.001 {
		t.Fatalf("eccentric anomaly mismatch: got %.6f want %.6f", got.EccentricAnomaly, 49.897)
	}
	if math.Abs(got.Radius-0.74557) > 1e-5 {
		t.Fatalf("radius mismatch: got %.6f want %.6f", got.Radius, 0.74557)
	}
	if math.Abs(got.TrueAnomaly-63.416) > 0.001 {
		t.Fatalf("true anomaly mismatch: got %.6f want %.6f", got.TrueAnomaly, 63.416)
	}
	if angleDiffAbs(got.PositionAngle, 318.4) > 0.05 {
		t.Fatalf("position angle mismatch: got %.6f want %.6f", got.PositionAngle, 318.4)
	}
	if math.Abs(got.Separation-0.411) > 0.002 {
		t.Fatalf("separation mismatch: got %.6f want %.6f", got.Separation, 0.411)
	}
}

func TestVisualBinaryMatchesMeeusGammaVirginisTable(t *testing.T) {
	elements := VisualBinaryElements{
		PeriodYears:        168.68,
		PeriastronYear:     2005.13,
		Eccentricity:       0.885,
		SemiMajorAxis:      3.697,
		Inclination:        148.0,
		AscendingNode:      36.9,
		PeriastronArgument: 256.5,
	}

	cases := []struct {
		year       float64
		angle      float64
		separation float64
	}{
		{1980.0, 296.65, 3.78},
		{1984.0, 293.10, 3.43},
		{1988.0, 288.70, 3.04},
		{1992.0, 282.89, 2.60},
		{1996.0, 274.41, 2.08},
		{2000.0, 259.34, 1.45},
		{2004.0, 208.67, 0.59},
		{2008.0, 35.54, 1.04},
		{2012.0, 12.72, 1.87},
	}

	for _, tc := range cases {
		got := VisualBinaryByYear(tc.year, elements)
		if angleDiffAbs(got.PositionAngle, tc.angle) > 0.12 {
			t.Fatalf("year %.1f position angle mismatch: got %.6f want %.6f", tc.year, got.PositionAngle, tc.angle)
		}
		if math.Abs(got.Separation-tc.separation) > 0.03 {
			t.Fatalf("year %.1f separation mismatch: got %.6f want %.6f", tc.year, got.Separation, tc.separation)
		}
	}
}

func TestVisualBinaryDateWrapperMatchesYearWrapper(t *testing.T) {
	elements := VisualBinaryElements{
		PeriodYears:        41.623,
		PeriastronYear:     1934.008,
		Eccentricity:       0.2763,
		SemiMajorAxis:      0.907,
		Inclination:        59.025,
		AscendingNode:      23.717,
		PeriastronArgument: 219.907,
	}

	date := time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC)
	fromDate := VisualBinary(date, elements)
	fromYear := VisualBinaryByYear(1980.0, elements)

	if angleDiffAbs(fromDate.PositionAngle, fromYear.PositionAngle) > 1e-12 {
		t.Fatalf("position angle wrapper mismatch: got %.12f want %.12f", fromDate.PositionAngle, fromYear.PositionAngle)
	}
	if math.Abs(fromDate.Separation-fromYear.Separation) > 1e-12 {
		t.Fatalf("separation wrapper mismatch: got %.12f want %.12f", fromDate.Separation, fromYear.Separation)
	}
}

func TestVisualBinaryInvalidInputReturnsNaN(t *testing.T) {
	got := VisualBinaryByYear(2000, VisualBinaryElements{
		PeriodYears:        0,
		PeriastronYear:     2005.13,
		Eccentricity:       0.885,
		SemiMajorAxis:      3.697,
		Inclination:        148.0,
		AscendingNode:      36.9,
		PeriastronArgument: 256.5,
	})

	for name, value := range map[string]float64{
		"mean":       got.MeanAnomaly,
		"eccentric":  got.EccentricAnomaly,
		"true":       got.TrueAnomaly,
		"radius":     got.Radius,
		"angle":      got.PositionAngle,
		"separation": got.Separation,
	} {
		if !math.IsNaN(value) {
			t.Fatalf("%s should be NaN for invalid input, got %.12f", name, value)
		}
	}
}
