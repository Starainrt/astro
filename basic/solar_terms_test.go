package basic

import (
	"math"
	"testing"
)

// GetJieqiLoops 的黄经必须归化到 (0,360]：loop>31 时旧实现把 >360° 的角度直接交给 GetJQTime，
// 那里 Newton 迭代静默失败，返回 9 个 NaN；loop<=0 旧实现还会 makeslice panic。
func TestGetJieqiLoopsNormalizesAngleAndGuardsLoop(t *testing.T) {
	short := GetJieqiLoops(2024, 25)
	long := GetJieqiLoops(2024, 40)
	if len(long) != 40 {
		t.Fatalf("loop=40 length = %d", len(long))
	}
	for index, value := range long {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			t.Fatalf("loop=40 index %d is not finite: %v", index, value)
		}
		if index > 0 && value <= long[index-1] {
			t.Fatalf("loop=40 index %d is not increasing: %v <= %v", index, value, long[index-1])
		}
	}
	for index := range short {
		if math.Float64bits(short[index]) != math.Float64bits(long[index]) {
			t.Fatalf("loop=25 index %d changed when loop grows: %v vs %v", index, short[index], long[index])
		}
	}
	if got := GetJieqiLoops(2024, 0); got != nil {
		t.Fatalf("loop=0 = %v, want nil", got)
	}
	if got := GetJieqiLoops(2024, -3); got != nil {
		t.Fatalf("negative loop = %v, want nil (no panic)", got)
	}
}
