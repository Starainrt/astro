package basic

import (
	"math"
	"testing"
)

func TestDefaultDeltaTFractionalYear(t *testing.T) {
	for _, year := range []float64{-700.5, -0.5, 0, 1000.5, 1582.75, 1800.5, 2000.5, 2026.5} {
		whole := math.Floor(year)
		start := JDECalc(int(whole), 1, 1)
		end := JDECalc(int(whole)+1, 1, 1)
		want := DefaultDeltaTv2(start+(year-whole)*(end-start), true)
		got := DefaultDeltaTv2(year, false)
		if math.IsNaN(got) || math.Abs(got-want) > 1e-10 {
			t.Errorf("year=%v DeltaT=%v, want %v", year, got, want)
		}
	}
}

func TestDefaultDeltaTInvalidInput(t *testing.T) {
	for _, value := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		for _, isJD := range []bool{false, true} {
			if got := DefaultDeltaTv2(value, isJD); !math.IsNaN(got) {
				t.Errorf("DeltaT(%v, %v)=%v, want NaN", value, isJD, got)
			}
		}
	}
}

func TestDeltaTLeapSecondBoundary(t *testing.T) {
	boundary := 2457754.5
	if got := DeltaTv2(boundary); got != 69.184 {
		t.Fatalf("DeltaT at leap-second boundary=%v, want 69.184", got)
	}
	if got := DeltaTSplineY(math.NaN()); !math.IsNaN(got) {
		t.Fatalf("DeltaTSplineY(NaN)=%v, want NaN", got)
	}
}

func TestTD2UTRoundTrip(t *testing.T) {
	for _, year := range []int{-2000, -720, 0, 26, 1426, 2025, 2027, 3627, 4026, 5000} {
		ut := JDECalc(year, 9, 20.123456)
		tt := TD2UT(ut, true)
		if got := TD2UT(tt, false); math.Abs(got-ut) > math.Nextafter(ut, math.Inf(1))-ut {
			t.Errorf("year=%d UT round trip differs by %.9f seconds", year, (got-ut)*86400)
		}
		if got := TD2UT(TD2UT(tt, false), true); got != tt {
			t.Errorf("year=%d TT round trip differs by %.9f seconds", year, (got-tt)*86400)
		}
	}
	for _, seconds := range []float64{-70, -1, -0.1, 0, 0.1, 1, 70} {
		ut := 2457754.5 + seconds/86400
		if got := TD2UT(TD2UT(ut, true), false); got != ut {
			t.Errorf("leap second offset=%g round trip differs by %.9f seconds", seconds, (got-ut)*86400)
		}
	}
}

func TestTD2UTCustomDeltaT(t *testing.T) {
	original := GetDeltaTFn()
	t.Cleanup(func() { SetDeltaTFn(original) })
	for _, slope := range []float64{0, 0.01} {
		SetDeltaTFn(func(jd float64, isJD bool) float64 {
			if !isJD {
				t.Fatal("conversion must request Delta T at a Julian day")
			}
			return 10000 + slope*(jd-2451545)
		})
		ut := 3000000.123456
		if got := TD2UT(TD2UT(ut, true), false); got != ut {
			t.Errorf("custom slope=%g round trip differs by %.9f seconds", slope, (got-ut)*86400)
		}
	}
}

// DeltaTSecondsAt：有限值（含 0）是显式覆盖，NaN/±Inf 才回退进程级模型；模型按 UT 求值。
func TestDeltaTSecondsAtOverrideContract(t *testing.T) {
	jd := 2460310.5
	if got := DeltaTSecondsAt(jd, 0); got != 0 {
		t.Fatalf("explicit zero override = %v, want 0", got)
	}
	if got := DeltaTSecondsAt(jd, 12.5); got != 12.5 {
		t.Fatalf("explicit override = %v, want 12.5", got)
	}
	model := DeltaTSecondsAt(jd, math.NaN())
	if !finite(model) || model <= 0 {
		t.Fatalf("model fallback = %v, want a positive finite value", model)
	}
	if got := DeltaTSecondsAt(jd, math.Inf(1)); got != model {
		t.Fatalf("+Inf override = %v, want the model %v", got, model)
	}
	if got := deltaTModelSecondsAtTT(jd); got != model {
		t.Fatalf("deltaTModelSecondsAtTT = %v, want %v", got, model)
	}
	// 模型按 UT 键控：与"把 TT 直接当 UT"的朴素求值有微小差别。
	naive := DeltaT(jd, true)
	if diff := math.Abs(naive - model); diff > 1e-3 {
		t.Fatalf("UT-keyed model differs from naive TT evaluation by %v s", diff)
	}
}
