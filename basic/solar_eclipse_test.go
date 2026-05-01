package basic

import (
	"math"
	"testing"
	"time"
)

type solarEclipseBaseline struct {
	name string
	jde  float64

	expectedType       SolarEclipseType
	expectedCentrality SolarEclipseCentrality
	expectedGreatestTT float64
	expectedGamma      float64
	expectedMagnitude  float64
	expectedLongitude  float64
	expectedLatitude   float64
	expectedPathWidth  float64
}

func TestSolarEclipseAgainstNASABaseline(t *testing.T) {
	// NASA GSFC Solar Eclipse Search Engine, Besselian Elements pages:
	// - 2023 Apr 20 hybrid
	// - 2024 Apr 08 total
	// - 2024 Oct 02 annular
	// - 2025 Mar 29 partial
	testCases := []solarEclipseBaseline{
		{
			name:               "2023-04-20 hybrid",
			jde:                JDECalc(2023, 4, 20),
			expectedType:       SolarEclipseHybrid,
			expectedCentrality: SolarEclipseCentralTwoLimits,
			expectedGreatestTT: solarEclipseTTJDE(2023, time.April, 20, 4, 17, 56),
			expectedGamma:      -0.3952,
			expectedMagnitude:  1.0132,
			expectedLongitude:  125.8,
			expectedLatitude:   -9.6,
			expectedPathWidth:  49.0,
		},
		{
			name:               "2024-04-08 total",
			jde:                JDECalc(2024, 4, 8),
			expectedType:       SolarEclipseTotal,
			expectedCentrality: SolarEclipseCentralTwoLimits,
			expectedGreatestTT: solarEclipseTTJDE(2024, time.April, 8, 18, 18, 29),
			expectedGamma:      0.3431,
			expectedMagnitude:  1.0566,
			expectedLongitude:  -104.1,
			expectedLatitude:   25.3,
			expectedPathWidth:  197.5,
		},
		{
			name:               "2024-10-02 annular",
			jde:                JDECalc(2024, 10, 2),
			expectedType:       SolarEclipseAnnular,
			expectedCentrality: SolarEclipseCentralTwoLimits,
			expectedGreatestTT: solarEclipseTTJDE(2024, time.October, 2, 18, 46, 13),
			expectedGamma:      -0.3509,
			expectedMagnitude:  0.9326,
			expectedLongitude:  -114.5,
			expectedLatitude:   -22.0,
			expectedPathWidth:  266.5,
		},
		{
			name:               "2025-03-29 partial",
			jde:                JDECalc(2025, 3, 29),
			expectedType:       SolarEclipsePartial,
			expectedCentrality: SolarEclipseNonCentral,
			expectedGreatestTT: solarEclipseTTJDE(2025, time.March, 29, 10, 48, 36),
			expectedGamma:      1.0405,
			expectedMagnitude:  0.9376,
			expectedLongitude:  -77.1,
			expectedLatitude:   61.1,
			expectedPathWidth:  0,
		},
	}

	const (
		timeToleranceDays   = 2.0 / 86400.0
		gammaTolerance      = 5e-4
		magnitudeTolerance  = 5e-4
		coordinateTolerance = 0.1
		pathWidthTolerance  = 5.0
	)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SolarEclipse(tc.jde)

			if result.Type != tc.expectedType {
				t.Fatalf("Type mismatch: got %s want %s", result.Type, tc.expectedType)
			}
			if result.Centrality != tc.expectedCentrality {
				t.Fatalf("Centrality mismatch: got %s want %s", result.Centrality, tc.expectedCentrality)
			}

			assertSolarEclipseJDEClose(t, "GreatestEclipse", result.GreatestEclipse, tc.expectedGreatestTT, timeToleranceDays)
			assertSolarEclipseFloatClose(t, "Gamma", result.Gamma, tc.expectedGamma, gammaTolerance)
			assertSolarEclipseFloatClose(t, "Magnitude", result.Magnitude, tc.expectedMagnitude, magnitudeTolerance)
			assertSolarEclipseFloatClose(t, "GreatestLongitude", result.GreatestLongitude, tc.expectedLongitude, coordinateTolerance)
			assertSolarEclipseFloatClose(t, "GreatestLatitude", result.GreatestLatitude, tc.expectedLatitude, coordinateTolerance)
			assertSolarEclipseFloatClose(t, "PathWidthKM", result.PathWidthKM, tc.expectedPathWidth, pathWidthTolerance)

			if result.HasPartial && !(result.PartialBeginOnEarth < result.GreatestEclipse && result.GreatestEclipse < result.PartialEndOnEarth) {
				t.Fatalf(
					"partial contact order invalid: begin=%.12f greatest=%.12f end=%.12f",
					result.PartialBeginOnEarth, result.GreatestEclipse, result.PartialEndOnEarth,
				)
			}
			if result.HasCentral && !(result.CentralBeginOnEarth < result.GreatestEclipse && result.GreatestEclipse < result.CentralEndOnEarth) {
				t.Fatalf(
					"central contact order invalid: begin=%.12f greatest=%.12f end=%.12f",
					result.CentralBeginOnEarth, result.GreatestEclipse, result.CentralEndOnEarth,
				)
			}
		})
	}
}

func TestSolarEclipseDefaultUsesNASABulletinSplitK(t *testing.T) {
	jde := JDECalc(2024, 4, 8)
	defaultResult := SolarEclipse(jde)
	nasaResult := SolarEclipseNASABulletinSplitK(jde)
	iauResult := SolarEclipseIAUSingleK(jde)

	if defaultResult.Model != SolarEclipseModelNASABulletinSplitK {
		t.Fatalf("default model mismatch: got %s want %s", defaultResult.Model, SolarEclipseModelNASABulletinSplitK)
	}

	assertSolarEclipseJDEClose(t, "GreatestEclipse", defaultResult.GreatestEclipse, nasaResult.GreatestEclipse, 1e-12)
	assertSolarEclipseFloatClose(t, "Gamma", defaultResult.Gamma, nasaResult.Gamma, 1e-12)
	assertSolarEclipseFloatClose(t, "Magnitude", defaultResult.Magnitude, nasaResult.Magnitude, 1e-12)
	assertSolarEclipseFloatClose(t, "PathWidthKM", defaultResult.PathWidthKM, nasaResult.PathWidthKM, 1e-12)

	if math.Abs(defaultResult.PathWidthKM-iauResult.PathWidthKM) < 0.5 {
		t.Fatalf(
			"default model should not collapse to IAU Single-K: default=%.6f iau=%.6f",
			defaultResult.PathWidthKM, iauResult.PathWidthKM,
		)
	}
	if !(iauResult.PathWidthKM > defaultResult.PathWidthKM) {
		t.Fatalf(
			"IAU Single-K should produce a wider total path than NASA Split-K here: iau=%.6f default=%.6f",
			iauResult.PathWidthKM, defaultResult.PathWidthKM,
		)
	}
}

func TestSolarEclipseNoEvent(t *testing.T) {
	testCases := []struct {
		name string
		calc func(float64) SolarEclipseResult
	}{
		{name: "default", calc: SolarEclipse},
		{name: "nasa", calc: SolarEclipseNASABulletinSplitK},
		{name: "iau", calc: SolarEclipseIAUSingleK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.calc(JDECalc(2023, 5, 15))
			if result.Type != SolarEclipseNone {
				t.Fatalf("Type mismatch: got %s want %s", result.Type, SolarEclipseNone)
			}
			if result.HasPartial || result.HasCentral || result.HasAnnular || result.HasTotal || result.HasHybrid {
				t.Fatalf("unexpected eclipse flags: %+v", result)
			}
			if result.PartialBeginOnEarth != 0 || result.PartialEndOnEarth != 0 || result.CentralBeginOnEarth != 0 || result.CentralEndOnEarth != 0 {
				t.Fatalf("expected no contact times, got %+v", result)
			}
		})
	}
}

func solarEclipseTTJDE(year int, month time.Month, day, hour, minute, second int) float64 {
	return Date2JDE(time.Date(year, month, day, hour, minute, second, 0, time.UTC))
}

func assertSolarEclipseJDEClose(t *testing.T, name string, got, want, tolerance float64) {
	t.Helper()
	if math.Abs(got-want) > tolerance {
		t.Fatalf("%s mismatch: got %.12f want %.12f", name, got, want)
	}
}

func assertSolarEclipseFloatClose(t *testing.T, name string, got, want, tolerance float64) {
	t.Helper()
	if math.Abs(got-want) > tolerance {
		t.Fatalf("%s mismatch: got %.9f want %.9f", name, got, want)
	}
}
