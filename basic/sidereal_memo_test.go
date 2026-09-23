package basic

import (
	"math"
	"sync"
	"testing"
)

// directApparentSiderealTime2006 是记忆化之前的原始表达式，作为逐位对照。
// directApparentSiderealTime2006 is the pre-memo expression, kept as a bit-exact reference.
func directApparentSiderealTime2006(jd float64) float64 {
	return MeanSiderealTime2006(jd) + Nutation2000Bi(jd)*math.Cos(TrueObliquity(jd)*math.Pi/180)/15
}

// directApparentSiderealTime1982 是 1982 模型的原始表达式，同样作为逐位对照。
// directApparentSiderealTime1982 is the pre-refactor 1982 expression, kept as a bit-exact reference.
func directApparentSiderealTime1982(jd float64) float64 {
	return MeanSiderealTime1982(jd) + Nutation2000Bi(jd)*math.Cos(TrueObliquity(jd)*math.Pi/180)/15
}

func siderealMemoTestTimes() []float64 {
	var times []float64
	for _, year := range []int{-720, -100, 0, 1000, 1582, 1900, 2025, 2026, 3000, 5000} {
		times = append(times, JDCalc(year, 3, 7.25), JDCalc(year, 9, 20.5), JDCalc(year, 12, 31.75))
	}
	return times
}

func TestApparentSiderealTimeMemoMatchesDirectSeries(t *testing.T) {
	times := siderealMemoTestTimes()
	firstPass := make([]float64, len(times))
	for index, jd := range times {
		got := ApparentSiderealTime2006(jd)
		if want := directApparentSiderealTime2006(jd); got != want {
			t.Fatalf("jd=%.6f sidereal=%v, want %v", jd, got, want)
		}
		if got1982, want1982 := ApparentSiderealTime1982(jd), directApparentSiderealTime1982(jd); got1982 != want1982 {
			t.Fatalf("jd=%.6f sidereal1982=%v, want %v", jd, got1982, want1982)
		}
		firstPass[index] = got
	}
	// 逆序再查一遍：命中记忆表也必须与首次计算逐位一致。
	// Query again in reverse: memo hits must stay bit-identical to the first evaluation.
	for index := len(times) - 1; index >= 0; index-- {
		if got := ApparentSiderealTime2006(times[index]); got != firstPass[index] {
			t.Errorf("jd=%.6f memoized sidereal=%v, want %v", times[index], got, firstPass[index])
		}
	}
}

func TestApparentSiderealTimeMemoInvalidatedByDeltaTOverride(t *testing.T) {
	original := GetDeltaTFn()
	t.Cleanup(func() { SetDeltaTFn(original) })
	jd := JDCalc(2025, 1, 5.5)

	baseline := ApparentSiderealTime2006(jd)
	if repeat := ApparentSiderealTime2006(jd); repeat != baseline {
		t.Fatalf("memoized sidereal=%v, want stable %v", repeat, baseline)
	}

	SetDeltaTFn(func(date float64, isJulianDay bool) float64 { return 6000 })
	shifted := ApparentSiderealTime2006(jd)
	if shifted == baseline {
		t.Fatalf("ΔT override replayed the memoized value %v", shifted)
	}
	if want := directApparentSiderealTime2006(jd); shifted != want {
		t.Errorf("sidereal after ΔT override=%v, want %v", shifted, want)
	}

	SetDeltaTFn(nil)
	if restored := ApparentSiderealTime2006(jd); restored != baseline {
		t.Errorf("sidereal after restoring ΔT=%v, want %v", restored, baseline)
	}
}

func TestApparentSiderealTimeMemoIsRaceFree(t *testing.T) {
	base := JDCalc(2025, 1, 5.5)
	var wait sync.WaitGroup
	for worker := 0; worker < 4; worker++ {
		wait.Add(1)
		go func(offset int) {
			defer wait.Done()
			for step := 0; step < 256; step++ {
				jd := base + float64(step)*1e-4 + float64(offset)*1e-7
				if got := ApparentSiderealTime2006(jd); got != directApparentSiderealTime2006(jd) {
					t.Errorf("jd=%.9f sidereal=%v, want %v", jd, got, directApparentSiderealTime2006(jd))
					return
				}
			}
		}(worker)
	}
	wait.Wait()
}

// 记忆表只在 ΔT 世代未变时写入：世代必须在求值前采样，否则求值期间发生的 SetDeltaTFn
// 会把旧 ΔT 的结果打上新世代并长期回放。
func TestSiderealMemoRejectsStaleGenerationWrite(t *testing.T) {
	original := GetDeltaTFn()
	defer SetDeltaTFn(original)
	resetSiderealMemo()
	jd := 2460310.5
	generation := siderealMemoGeneration()
	SetDeltaTFn(func(date float64, isJd bool) float64 { return 200 })
	siderealMemoStore(jd, 12345.0, generation)
	if _, ok := siderealMemoLoad(jd); ok {
		t.Fatalf("stale-generation write was accepted")
	}
	// 世代未变时正常写入并命中。
	current := siderealMemoGeneration()
	siderealMemoStore(jd, 12345.0, current)
	value, ok := siderealMemoLoad(jd)
	if !ok || value != 12345.0 {
		t.Fatalf("current-generation write missed: value=%v ok=%v", value, ok)
	}
}

// 同一批瞬时重复求值必须命中记忆表（否则扩表/世代逻辑等于没生效）。
func TestSiderealMemoHitsRepeatedInstants(t *testing.T) {
	resetSiderealMemo()
	const distinct = 8192
	for offset := 0; offset < distinct; offset++ {
		_ = ApparentSiderealTime2006(2460310.5 + float64(offset)*0.25)
	}
	hitsBefore, missesBefore, _ := siderealMemoStats()
	if missesBefore < distinct || hitsBefore != 0 {
		t.Fatalf("first pass should miss every distinct instant: hits=%d misses=%d", hitsBefore, missesBefore)
	}
	for offset := 0; offset < distinct; offset++ {
		_ = ApparentSiderealTime2006(2460310.5 + float64(offset)*0.25)
	}
	hits, misses, _ := siderealMemoStats()
	if hits-hitsBefore < distinct*9/10 {
		t.Fatalf("second pass hit rate too low: hits=%d misses=%d (want >= 90%% of %d)",
			hits-hitsBefore, misses-missesBefore, distinct)
	}
}
