package basic

import (
	"math"
	"testing"
)

func TestBarycentricScaleDefinitions(t *testing.T) {
	for _, jd := range []float64{JDCalc(-3000, 1, 1), barycentricT0JDE, JDCalc(2026, 1, 1), JDCalc(6000, 1, 1)} {
		tcgOffset := TCGMinusTTSeconds(jd)
		if diff := tcgOffset*(1-6.969290134e-10) - 6.969290134e-10*(jd-2443144.5003725)*86400; math.Abs(diff) > 1e-12 {
			t.Errorf("jd=%v TCG defining relation residual=%g s", jd, diff)
		}
		tcbOffset := TCBMinusTTSeconds(jd)
		if diff := tcbOffset*(1-1.550519768e-8) - 1.550519768e-8*(jd-2443144.5003725)*86400 - TDBMinusTTSeconds(jd) - 65.5e-6; math.Abs(diff) > 1e-12 {
			t.Errorf("jd=%v TCB defining relation residual=%g s", jd, diff)
		}
	}
	if got := TCGMinusTTSeconds(barycentricT0JDE); got != 0 {
		t.Errorf("TCG-TT at T0=%v, want 0", got)
	}
	if got, want := TCB2TDB(barycentricT0JDE), barycentricT0JDE-65.5e-6/86400; got != want {
		t.Errorf("TDB at TCB T0=%.12f, want %.12f", got, want)
	}
}

func TestBarycentricLinearConversionAnchors(t *testing.T) {
	for _, tc := range []struct {
		name string
		fn   func(float64) float64
		jd   float64
		want float64
	}{
		{"TCB2TDB", TCB2TDB, 2453750.5 + 0.893019599, 2453750.5 + 0.8928551362746343397},
		{"TDB2TCB", TDB2TCB, 2453750.5 + 0.892855137, 2453750.5 + 0.8930195997253656716},
		{"TCG2TT", TCG2TT, 2453750.5 + 0.892862531, 2453750.5 + 0.8928551387488816828},
		{"TT2TCG", TT2TCG, 2453750.5 + 0.892482639, 2453750.5 + 0.8924900312508587113},
	} {
		ulp := math.Nextafter(tc.want, math.Inf(1)) - tc.want
		if got := tc.fn(tc.jd); math.Abs(got-tc.want) > ulp {
			t.Errorf("%s=%.12f, want %.12f", tc.name, got, tc.want)
		}
	}
}

func TestBarycentricPeriodicTermsMatchFullSeries(t *testing.T) {
	for _, tc := range []struct {
		jd   float64
		want float64
	}{
		{625307.5, 0.001558488628712037},
		{766574.5, -0.000924800103328196},
		{1721057.5, 0.0008624527814561311},
		{2433282.5, -7.069829559472634e-05},
		{2443144.5, -6.551401857064118e-05},
		{2451544.5, -0.0001137630988927298},
		{2461041.5, -8.20152430051247e-05},
		{2461222.5, 0.0001186656922342686},
		{2469807.5, -8.01882947792431e-05},
		{3182029.5, -0.0009618304620565844},
		{3912514.5, -0.00138271667055309},
	} {
		if got := TDBMinusTTSeconds(tc.jd); math.Abs(got-tc.want) > 2.1e-6 {
			t.Errorf("jd=%v TDB-TT=%.9f us, want %.9f us", tc.jd, got*1e6, tc.want*1e6)
		}
	}
}

func TestBarycentricScaleRoundTrips(t *testing.T) {
	for _, jd := range []float64{
		JDCalc(-3000, 1, 1), JDCalc(0, 1, 1), JDCalc(1900, 1, 1), JDCalc(1950, 6, 1),
		barycentricT0JDE, JDCalc(2000, 1, 1), JDCalc(2026, 4, 1), JDCalc(2050, 7, 1),
		JDCalc(2100, 1, 1), JDCalc(6000, 1, 1),
	} {
		ulp := math.Nextafter(jd, math.Inf(1)) - jd
		for _, tc := range []struct {
			name string
			fwd  func(float64) float64
			back func(float64) float64
		}{
			{"TT-TCG", TT2TCG, TCG2TT},
			{"TT-TCB", TT2TCB, TCB2TT},
			{"TT-TDB", TT2TDB, TDB2TT},
			{"TCB-TDB", TCB2TDB, TDB2TCB},
		} {
			if got := tc.back(tc.fwd(jd)); math.Abs(got-jd) > ulp {
				t.Errorf("jd=%v %s round-trip error=%g s", jd, tc.name, (got-jd)*86400)
			}
			if got := tc.fwd(tc.back(jd)); math.Abs(got-jd) > ulp {
				t.Errorf("jd=%v %s inverse round-trip error=%g s", jd, tc.name, (got-jd)*86400)
			}
		}
		// 组合路径分别舍入，允许两次 JD 舍入误差。
		if diff := TCB2TDB(TT2TCB(jd)) - TT2TDB(jd); math.Abs(diff) > 2*ulp {
			t.Errorf("jd=%v TT-TCB-TDB differs from TT-TDB by %g s", jd, diff*86400)
		}
		if diff := (TT2TDB(jd)-jd)*86400 - TDBMinusTTSeconds(jd); math.Abs(diff) > ulp*86400 {
			t.Errorf("jd=%v TDB seconds and JD conversions differ by %g s", jd, diff)
		}
	}
}

func TestBarycentricScaleMagnitudes(t *testing.T) {
	jd := JDCalc(2026, 1, 1)
	if got := TCGMinusTTSeconds(jd); math.Abs(got-1.0777) > 5e-3 {
		t.Errorf("TCG-TT=%v s, want about 1.078", got)
	}
	if got := TCBMinusTTSeconds(jd); math.Abs(got-23.9757) > 5e-3 {
		t.Errorf("TCB-TT=%v s, want about 23.976", got)
	}
	if got := TCGMinusTTSeconds(barycentricT0JDE+365.25) - TCGMinusTTSeconds(barycentricT0JDE); math.Abs(got-0.02204) > 5e-5 {
		t.Errorf("TCG-TT annual increase=%v s, want about 0.02204", got)
	}
	if got := TCBMinusTTSeconds(barycentricT0JDE+365.25) - TCBMinusTTSeconds(barycentricT0JDE); math.Abs(got-0.48934) > 5e-4 {
		t.Errorf("TCB-TT annual increase=%v s, want about 0.48934", got)
	}
	if got := TDBMinusTTSeconds(JDCalc(2100, 1, 1)) - TDBMinusTTSeconds(JDCalc(2000, 1, 1)); math.Abs(got) > 2e-4 {
		t.Errorf("TDB-TT century difference=%v s, want within 0.2 ms", got)
	}
	min, max := math.Inf(1), math.Inf(-1)
	for month := 1; month <= 12; month++ {
		value := TDBMinusTTSeconds(JDCalc(2026, month, 1))
		min, max = math.Min(min, value), math.Max(max, value)
	}
	if amplitude := max - min; amplitude < 3.0e-3 || amplitude > 3.6e-3 {
		t.Errorf("TDB-TT annual range=%v s, want 3.0-3.6 ms", amplitude)
	}
}
