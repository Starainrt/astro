package basic

import (
	"math"
	"testing"
)

func TestMercuryRetrogradeFastPathKeepsTypedCandidateOrder(t *testing.T) {
	// 这些查询覆盖一个普通古代周期和两个长间隔边界 / These queries cover a normal ancient cycle and two long-gap boundaries
	//，类型化驻留搜索必须作为最终权威 / where the typed station search must remain the final authority.
	for _, queryUT := range []float64{
		1026548.810500779,
		1644733.927538287,
		1645082.416782375,
	} {
		queryTT := TD2UT(queryUT, true)
		nextP2R := NextMercuryProgradeToRetrograde(queryTT)
		nextR2P := NextMercuryRetrogradeToPrograde(queryTT)
		wantNext := math.Min(nextP2R, nextR2P)
		if got := NextMercuryRetrograde(queryTT); math.Abs(got-wantNext) > 1e-7 {
			t.Fatalf("next aggregate mismatch at %.9f: got %.12f want %.12f", queryUT, got, wantNext)
		}

		lastP2R := LastMercuryProgradeToRetrograde(queryTT)
		lastR2P := LastMercuryRetrogradeToPrograde(queryTT)
		wantLast := math.Max(lastP2R, lastR2P)
		if got := LastMercuryRetrograde(queryTT); math.Abs(got-wantLast) > 1e-7 {
			t.Fatalf("last aggregate mismatch at %.9f: got %.12f want %.12f", queryUT, got, wantLast)
		}
	}
}
