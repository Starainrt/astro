//go:build astro_lowmem
// +build astro_lowmem

// 被测契约：FlashModel 与同表 Model 的位置、速度逐位一致，覆盖 ±1 万年。
package lunar

import (
	"math"
	"testing"
)

type lunarEvaluator interface {
	PositionJ2000(jdTT float64) [3]float64
	PositionVelocityJ2000(jdTT float64) (pos, vel [3]float64)
}

func assertSameLunar(t *testing.T, name string, want, got lunarEvaluator) {
	n := 0
	for jd := j2000 - 3652500.0; jd <= j2000+3652500.0; jd += 3652.5 * 0.973 {
		wp := want.PositionJ2000(jd)
		gp := got.PositionJ2000(jd)
		wvp, wv := want.PositionVelocityJ2000(jd)
		gvp, gv := got.PositionVelocityJ2000(jd)
		for i := 0; i < 3; i++ {
			if math.Float64bits(wp[i]) != math.Float64bits(gp[i]) ||
				math.Float64bits(wvp[i]) != math.Float64bits(gvp[i]) ||
				math.Float64bits(wv[i]) != math.Float64bits(gv[i]) {
				t.Fatalf("%s jd=%.3f axis %d: pos %v vs %v, vel %v vs %v", name, jd, i, wp[i], gp[i], wv[i], gv[i])
			}
		}
		n++
	}
	if n < 1000 {
		t.Fatalf("%s: only %d samples", name, n)
	}
}

func TestFlashDefaultMatchesDefault(t *testing.T) {
	assertSameLunar(t, "default", Default(), FlashDefault())
}

func TestFlashFastMatchesFast(t *testing.T) {
	assertSameLunar(t, "fast", Fast(), FlashFast())
}

func TestParseFlashRejectsTruncated(t *testing.T) {
	for _, n := range []int{0, 3, 5, 40, len(flashDefaultTable) / 2, len(flashDefaultTable) - 1} {
		if _, err := ParseFlash(flashDefaultTable[:n]); err == nil {
			t.Errorf("accepted table truncated to %d bytes", n)
		}
	}
	if _, err := ParseFlash(flashDefaultTable + "x"); err == nil {
		t.Error("accepted trailing bytes")
	}
}
