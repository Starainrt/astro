package basic

import (
	"math"
	"testing"
)

func TestDefaultDeltaTFractionalYear(t *testing.T) {
	for _, year := range []float64{-700.5, -0.5, 0, 1000.5, 1582.75, 1800.5, 2000.5, 2026.5} {
		whole := math.Floor(year)
		start := JDCalc(int(whole), 1, 1)
		end := JDCalc(int(whole)+1, 1, 1)
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

// 闰秒只改 TT−UTC，不改 TT−UT1：ΔT 在闰秒两侧连续，TT−UTC 才跳 1 秒。
func TestDeltaTStaysContinuousAcrossLeapSecond(t *testing.T) {
	boundary := 2457754.5
	before := DeltaTv2(boundary - 1e-6)
	after := DeltaTv2(boundary + 1e-6)
	if math.Abs(after-before) > 1e-3 {
		t.Fatalf("ΔT should stay continuous at a leap second: %v → %v", before, after)
	}
	if step := TTMinusUTCSeconds(boundary+1e-6) - TTMinusUTCSeconds(boundary-1e-6); math.Abs(step-1) > 1e-9 {
		t.Fatalf("TT−UTC should step by one second: %v", step)
	}
	if got := DeltaTSplineY(math.NaN()); !math.IsNaN(got) {
		t.Fatalf("DeltaTSplineY(NaN)=%v, want NaN", got)
	}
}

func TestUTC2TTRoundTrip(t *testing.T) {
	for _, year := range []int{-2000, -720, 0, 26, 1426, 2025, 2027, 3627, 4026, 5000} {
		utc := JDCalc(year, 9, 20.123456)
		tt := UTC2TT(utc)
		if got := TT2UTC(tt); math.Abs(got-utc) > math.Nextafter(utc, math.Inf(1))-utc {
			t.Errorf("year=%d UTC round trip differs by %.9f seconds", year, (got-utc)*86400)
		}
		if got := UTC2TT(TT2UTC(tt)); got != tt {
			t.Errorf("year=%d TT round trip differs by %.9f seconds", year, (got-tt)*86400)
		}
	}
	for _, seconds := range []float64{-70, -1, -0.1, 0, 0.1, 1, 70} {
		utc := 2457754.5 + seconds/86400
		if got := TT2UTC(UTC2TT(utc)); got != utc {
			t.Errorf("leap second offset=%g round trip differs by %.9f seconds", seconds, (got-utc)*86400)
		}
	}
}

func TestUTC2TTRoundTripUnderCustomDeltaT(t *testing.T) {
	original := GetDeltaTFn()
	t.Cleanup(func() { SetDeltaTFn(original) })
	for _, slope := range []float64{0, 0.01} {
		SetDeltaTFn(func(jd float64, isJD bool) float64 {
			if !isJD {
				t.Fatal("conversion must request Delta T at a Julian day")
			}
			return 10000 + slope*(jd-2451545)
		})
		utc := 3000000.123456
		if got := TT2UTC(UTC2TT(utc)); got != utc {
			t.Errorf("custom slope=%g round trip differs by %.9f seconds", slope, (got-utc)*86400)
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
