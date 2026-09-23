package basic

import (
	"math"
	"testing"
)

// 根数表口径：经典贝塞尔符号、NASA 式三次拟合，以及与已发布表的 μ 时间自变量换算。

const (
	besselianNASAT0JDE  = 2460409.25
	besselianNASADeltaT = 70.6
	besselianNASASeed   = 2460409.262835
)

func nasa2024BesselianElements(t *testing.T) SolarEclipseBesselianElementsResult {
	t.Helper()
	result, ok := SolarEclipseBesselianElements(besselianNASASeed, SolarEclipseBesselianElementsOptions{
		DeltaTSeconds: besselianNASADeltaT,
		ReferenceJDE:  besselianNASAT0JDE,
	})
	if !ok {
		t.Fatal("2024-04-08 should have a solar eclipse")
	}
	return result
}

func TestSolarEclipseBesselianElementsMatchesNASA2024(t *testing.T) {
	result := nasa2024BesselianElements(t)

	cases := []struct {
		label string
		got   SolarEclipseBesselianPolynomial
		want  SolarEclipseBesselianPolynomial
		tol   [4]float64
	}{
		{"x", result.X, SolarEclipseBesselianPolynomial{-0.318157, 0.5117105, 0.0000326, -0.0000085}, [4]float64{2e-4, 1e-5, 1e-6, 1e-6}},
		{"y", result.Y, SolarEclipseBesselianPolynomial{0.219747, 0.2709586, -0.0000594, -0.0000047}, [4]float64{2e-4, 1e-5, 1e-6, 1e-6}},
		{"d", result.D, SolarEclipseBesselianPolynomial{7.58620, 0.014844, -0.000002, 0}, [4]float64{1e-4, 1e-5, 1e-5, 1e-5}},
		{"l1", result.L1, SolarEclipseBesselianPolynomial{0.535813, 0.0000618, -0.0000128, 0}, [4]float64{1e-4, 1e-5, 1e-5, 1e-5}},
		{"l2", result.L2, SolarEclipseBesselianPolynomial{-0.010274, 0.0000615, -0.0000127, 0}, [4]float64{1e-4, 1e-5, 1e-5, 1e-5}},
	}
	for _, item := range cases {
		for n := range item.want {
			if math.Abs(item.got[n]-item.want[n]) > item.tol[n] {
				t.Errorf("%s[%d] = %.8f, want %.8f (tol %.1e)", item.label, n, item.got[n], item.want[n], item.tol[n])
			}
		}
	}

	published := SolarEclipseBesselianMuForPublishedTable(result.Mu, result.DeltaTSeconds)
	for n, want := range [4]float64{89.59122, 15.004084, 0, 0} {
		if math.Abs(published[n]-want) > 1e-4 {
			t.Errorf("published mu[%d] = %.6f, want %.6f", n, published[n], want)
		}
	}

	if math.Abs(result.TanF1-0.0046683) > 1e-6 || math.Abs(result.TanF2-0.0046450) > 1e-6 {
		t.Errorf("tan f = %.8f / %.8f, want 0.0046683 / 0.0046450", result.TanF1, result.TanF2)
	}
	if math.Abs(result.Gamma-0.3431) > 1e-4 {
		t.Errorf("gamma = %.6f, want 0.3431", result.Gamma)
	}
	if math.Abs(result.Magnitude-1.0566) > 1e-3 {
		t.Errorf("magnitude = %.6f, want 1.0566", result.Magnitude)
	}
	if result.T0JDE != besselianNASAT0JDE || result.ValidHours != 3 {
		t.Errorf("t0 = %.6f validHours = %g, want %.6f / 3", result.T0JDE, result.ValidHours, besselianNASAT0JDE)
	}
	if result.Model != SolarEclipseModelNASABulletinSplitK {
		t.Errorf("model = %q, want %q", result.Model, SolarEclipseModelNASABulletinSplitK)
	}
	if result.PenumbralK != solarEclipsePenumbralK || result.UmbralK != solarEclipseUmbralK {
		t.Errorf("k = %g / %g, want %g / %g", result.PenumbralK, result.UmbralK, solarEclipsePenumbralK, solarEclipseUmbralK)
	}
	if math.Abs(result.DeltaTSeconds-besselianNASADeltaT) > 1e-9 {
		t.Errorf("deltaT = %g, want %g", result.DeltaTSeconds, besselianNASADeltaT)
	}
}

func TestSolarEclipseBesselianElementsFollowsSampledElements(t *testing.T) {
	result := nasa2024BesselianElements(t)
	solver := newSolarEclipseSolver(
		CalcMoonSHByJDE(besselianNASASeed, 0),
		SolarEclipseModelNASABulletinSplitK,
	).withDeltaTSeconds(besselianNASADeltaT)

	for _, hours := range []float64{-2.7, -0.8, 0.4, 2.3} {
		point := solver.besselianElementsAt(result.T0JDE + hours/24)
		if math.Abs(result.X.At(hours)-point.x) > 1e-5 ||
			math.Abs(result.Y.At(hours)-point.y) > 1e-5 ||
			math.Abs(result.L1.At(hours)-point.l1) > 1e-5 ||
			math.Abs(result.L2.At(hours)-point.l2) > 1e-5 ||
			math.Abs(result.D.At(hours)-point.d) > 1e-5 {
			t.Errorf("t=%+.2fh polynomial (%g,%g,%g,%g,%g) does not track elements (%g,%g,%g,%g,%g)",
				hours, result.X.At(hours), result.Y.At(hours), result.D.At(hours), result.L1.At(hours), result.L2.At(hours),
				point.x, point.y, point.d, point.l1, point.l2)
		}
		if math.Abs(math.Mod(result.Mu.At(hours)-point.mu, 360)) > 1e-4 {
			t.Errorf("t=%+.2fh mu = %g does not track %g", hours, result.Mu.At(hours), point.mu)
		}
	}
}

func TestSolarEclipseBesselianElementsMuStaysContinuousAcrossZero(t *testing.T) {
	// 2013-11-03 的食甚点在格林尼治附近，μ 的采样点会跨过 0°，是展开逻辑的边界用例。
	for _, seed := range []float64{besselianNASASeed, JDCalc(2013, 11, 3)} {
		result, ok := SolarEclipseBesselianElements(seed, SolarEclipseBesselianElementsOptions{
			DeltaTSeconds: besselianNASADeltaT,
		})
		if !ok {
			t.Fatalf("seed %.4f should have a solar eclipse", seed)
		}
		if result.Mu.At(0) < 0 || result.Mu.At(0) >= 360 {
			t.Errorf("seed %.4f: mu(0) = %g, want [0,360)", seed, result.Mu.At(0))
		}
		if math.Abs(result.Mu[1]-15.041067) > 0.05 {
			t.Errorf("seed %.4f: mu rate = %.6f, want about 15.041067", seed, result.Mu[1])
		}
		for _, hours := range []float64{-3, -1.5, 1.5, 3} {
			drift := result.Mu.At(hours) - (result.Mu.At(0) + 15.041067*hours)
			if math.Abs(drift) > 0.5 {
				t.Errorf("seed %.4f: mu(%+.1f) is not a continuous continuation, drift %.4f", seed, hours, drift)
			}
		}
	}
}

func TestSolarEclipseBesselianElementsDefaultsToWholeHour(t *testing.T) {
	result, ok := SolarEclipseBesselianElements(besselianNASASeed, SolarEclipseBesselianElementsOptions{
		DeltaTSeconds: besselianNASADeltaT,
	})
	if !ok {
		t.Fatal("2024-04-08 should have a solar eclipse")
	}
	if math.Abs(result.T0JDE*24-math.Round(result.T0JDE*24)) > 1e-9 {
		t.Errorf("t0 = %.9f is not a whole TT hour", result.T0JDE)
	}
	if math.Abs(result.T0JDE*24-besselianNASASeed*24) > 0.5 {
		t.Errorf("t0 = %.6f is not the whole hour nearest greatest eclipse %.6f", result.T0JDE, besselianNASASeed)
	}
	if result.ValidHours != 3 {
		t.Errorf("validHours = %g, want 3", result.ValidHours)
	}
	// 食甚落在半小时之后的事件取下一个整小时：NASA 2009-07-22 表食甚 02:36:24 TDT，t0 记 03:00 TDT。
	forward, ok := SolarEclipseBesselianElements(JDCalc(2009, 7, 22), SolarEclipseBesselianElementsOptions{
		DeltaTSeconds: 65.9,
	})
	if !ok {
		t.Fatal("2009-07-22 should have a solar eclipse")
	}
	if math.Abs(forward.T0JDE-2455034.625) > 1e-9 {
		t.Errorf("t0 = %.9f, want 2455034.625 (03:00 TDT)", forward.T0JDE)
	}
}

func TestSolarEclipseBesselianElementsMatchesPublishedHours(t *testing.T) {
	testCases := []struct {
		name  string
		date  [3]int
		delta float64
		t0    float64
		want  [6]SolarEclipseBesselianPolynomial
	}{
		{
			name: "2009-07-22 total", date: [3]int{2009, 7, 22}, delta: 65.9, t0: 2455034.625,
			want: [6]SolarEclipseBesselianPolynomial{
				{0.240059, 0.5563975, -0.0000583, -0.0000100},
				{-0.003283, -0.1774571, -0.0001346, 0.0000032},
				{20.26424, -0.007873, -0.000004, 0},
				{0.530426, 0.0000063, -0.0000128, 0},
				{-0.015633, 0.0000063, -0.0000127, 0},
				{0, 0, 0, 0},
			},
		},
		{
			name: "2013-11-03 hybrid", date: [3]int{2013, 11, 3}, delta: 67.2, t0: 2456600.041666667,
			want: [6]SolarEclipseBesselianPolynomial{
				{0.183190, 0.5469478, 0.0000282, -0.0000083},
				{0.294721, -0.1200753, 0.0000790, 0.0000017},
				{-15.20965, -0.012636, 0.000003, 0},
				{0.546301, -0.0001121, -0.0000120, 0},
				{0.000143, -0.0001116, -0.0000120, 0},
				{0, 0, 0, 0},
			},
		},
	}
	tolerances := [6][4]float64{
		{2e-4, 1e-5, 1e-6, 1e-6},
		{2e-4, 1e-5, 1e-6, 1e-6},
		{1e-4, 1e-5, 1e-5, 1e-5},
		{1e-4, 1e-5, 1e-5, 1e-5},
		{1e-4, 1e-5, 1e-5, 1e-5},
		{1e-4, 1e-5, 1e-5, 1e-5},
	}
	labels := [6]string{"x", "y", "d", "l1", "l2", "mu"}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, ok := SolarEclipseBesselianElements(JDCalc(tc.date[0], tc.date[1], float64(tc.date[2])),
				SolarEclipseBesselianElementsOptions{DeltaTSeconds: tc.delta})
			if !ok {
				t.Fatalf("%s should have a solar eclipse", tc.name)
			}
			if math.Abs(result.T0JDE-tc.t0) > 1e-9 {
				t.Fatalf("t0 = %.9f, want %.9f", result.T0JDE, tc.t0)
			}
			got := [6]SolarEclipseBesselianPolynomial{result.X, result.Y, result.D, result.L1, result.L2, result.Mu}
			for index := range got {
				if index == 5 {
					continue
				}
				for n := range tc.want[index] {
					if math.Abs(got[index][n]-tc.want[index][n]) > tolerances[index][n] {
						t.Errorf("%s[%d] = %.8f, want %.8f (tol %.1e)", labels[index], n,
							got[index][n], tc.want[index][n], tolerances[index][n])
					}
				}
			}
		})
	}
}

func TestSolarEclipseBesselianElementsIAUSingleKUsesOneK(t *testing.T) {
	result, ok := SolarEclipseBesselianElements(besselianNASASeed, SolarEclipseBesselianElementsOptions{
		Model:         SolarEclipseModelIAUSingleK,
		DeltaTSeconds: besselianNASADeltaT,
	})
	if !ok {
		t.Fatal("2024-04-08 should have a solar eclipse")
	}
	if result.Model != SolarEclipseModelIAUSingleK {
		t.Errorf("model = %q, want %q", result.Model, SolarEclipseModelIAUSingleK)
	}
	if result.UmbralK != result.PenumbralK {
		t.Errorf("k2 = %g, want k1 = %g", result.UmbralK, result.PenumbralK)
	}
}

func TestSolarEclipseBesselianElementsWithoutEclipse(t *testing.T) {
	if _, ok := SolarEclipseBesselianElements(JDCalc(2024, 5, 8), SolarEclipseBesselianElementsOptions{}); ok {
		t.Fatal("2024-05-08 has no solar eclipse")
	}
}

func TestSolarEclipseBesselianElementsRejectsShortSampleSet(t *testing.T) {
	if got := solarEclipseFitCubic([]float64{-1, 0, 1}, []float64{1, 2, 3}); got != (SolarEclipseBesselianPolynomial{}) {
		t.Errorf("fit with three samples = %v, want zero", got)
	}
	if got := solarEclipseFitCubic([]float64{-3, -1.5, 0, 1.5, 3}, []float64{0, 0, 0, 0, 0}); got != (SolarEclipseBesselianPolynomial{}) {
		t.Errorf("fit of a zero column = %v, want zero", got)
	}
}
