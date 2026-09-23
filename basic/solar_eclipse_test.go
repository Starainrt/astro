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
			jde:                JDCalc(2023, 4, 20),
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
			jde:                JDCalc(2024, 4, 8),
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
			jde:                JDCalc(2024, 10, 2),
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
			jde:                JDCalc(2025, 3, 29),
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

func TestSolarEclipseBesselGeometryCacheSeparatesExactAndCandidate(t *testing.T) {
	seed := JDCalc(2024, 4, 8)
	solver := newSolarEclipseSolver(CalcMoonSHByJDE(seed, 0), SolarEclipseModelNASABulletinSplitK)
	tt := solver.newMoonJDE + 0.125
	exactMoon, exactAxis, exactSun := solver.besselGeometryAt(tt)
	exactMoonAgain, exactAxisAgain, exactSunAgain := solver.besselGeometryAt(tt)
	if exactMoon != exactMoonAgain || exactAxis != exactAxisAgain || exactSun != exactSunAgain {
		t.Fatal("exact Bessel cache changed the geometry on a repeated lookup")
	}
	if len(solver.besselGeometryCache) != 1 {
		t.Fatalf("exact Bessel cache size=%d, want 1", len(solver.besselGeometryCache))
	}

	candidateSolver := solver.withLocalEphemeris()
	candidateMoon, candidateAxis, candidateSun, ok := candidateSolver.besselGeometryCandidateAt(tt)
	if !ok {
		t.Fatal("candidate Bessel geometry unavailable inside local ephemeris")
	}
	candidateMoonAgain, candidateAxisAgain, candidateSunAgain, okAgain := candidateSolver.besselGeometryCandidateAt(tt)
	if !okAgain || candidateMoon != candidateMoonAgain || candidateAxis != candidateAxisAgain || candidateSun != candidateSunAgain {
		t.Fatal("candidate Bessel cache changed the geometry on a repeated lookup")
	}
	if len(solver.besselGeometryCache) != 1 {
		t.Fatalf("exact Bessel cache size changed after candidate lookup=%d", len(solver.besselGeometryCache))
	}
	if len(solver.besselCandidateCache) != 1 {
		t.Fatalf("candidate Bessel cache size=%d, want 1", len(solver.besselCandidateCache))
	}
}

func TestSolarEclipseBesselGeometryCacheIsBounded(t *testing.T) {
	cache := make(map[uint64]solarEclipseBesselGeometryCacheEntry)
	entry := solarEclipseBesselGeometryCacheEntry{valid: true}
	for index := 0; index < solarEclipseBesselGeometryCacheMaximumEntries+17; index++ {
		storeSolarEclipseBesselGeometry(cache, uint64(index), entry)
	}
	if len(cache) > solarEclipseBesselGeometryCacheMaximumEntries {
		t.Fatalf("Bessel cache size=%d, want <=%d", len(cache), solarEclipseBesselGeometryCacheMaximumEntries)
	}
	if _, ok := cache[uint64(solarEclipseBesselGeometryCacheMaximumEntries+16)]; !ok {
		t.Fatal("latest Bessel cache entry was evicted")
	}
}

func BenchmarkSolarEclipseBesselGeometryCache(b *testing.B) {
	solver := newSolarEclipseSolver(
		CalcMoonSHByJDE(JDCalc(2024, 4, 8), 0),
		SolarEclipseModelNASABulletinSplitK,
	)
	tt := solver.newMoonJDE + 0.125
	// Prime one exact entry so the benchmark measures the repeated lookup
	// path used by contact and topology refinements.
	solver.besselGeometryAt(tt)
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		solver.besselGeometryAt(tt)
	}
}

func BenchmarkSolarEclipseRepresentativePath(b *testing.B) {
	seed := JDCalc(2024, 4, 8)
	options := SolarEclipsePathOptions{StepDays: 5.0 / 1440.0, TargetSpacingKM: 200}
	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		SolarEclipseCentralPath(seed, options)
	}
}

func TestSolarEclipseDefaultUsesNASABulletinSplitK(t *testing.T) {
	jde := JDCalc(2024, 4, 8)
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
			result := tc.calc(JDCalc(2023, 5, 15))
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

func BenchmarkSolarEclipseGlobal(b *testing.B) {
	seed := JDCalc(2010, 1, 15)
	b.ReportAllocs()
	for iteration := 0; iteration < b.N; iteration++ {
		result := SolarEclipse(seed)
		if !result.HasCentral {
			b.Fatal("expected a central eclipse")
		}
	}
}

func solarEclipseTTJDE(year int, month time.Month, day, hour, minute, second int) float64 {
	return Date2JD(time.Date(year, month, day, hour, minute, second, 0, time.UTC))
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
