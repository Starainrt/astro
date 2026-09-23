package astro

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

// time.Time 往返经 JD 舍入，容差为 0.2 ms。
func TestBarycentricTimeFacade(t *testing.T) {
	tt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if got := TCGMinusTT(tt); math.Abs(got-1.0777) > 5e-3 {
		t.Errorf("TCG−TT=%v 秒, want ≈1.078", got)
	}
	if got := TCBMinusTT(tt); math.Abs(got-23.9757) > 5e-3 {
		t.Errorf("TCB−TT=%v 秒, want ≈23.976", got)
	}
	if got := TDBMinusTT(tt); math.Abs(got) > 2e-3 {
		t.Errorf("TDB−TT=%v 秒, want |·| ≤ 1.7 ms", got)
	}
	for _, tc := range []struct {
		name string
		fwd  func(time.Time) time.Time
		back func(time.Time) time.Time
	}{
		{"TCG", TCGFromTT, TTFromTCG},
		{"TCB", TCBFromTT, TTFromTCB},
		{"TDB", TDBFromTT, TTFromTDB},
	} {
		round := tc.back(tc.fwd(tt))
		if delta := math.Abs(round.Sub(tt).Seconds()); delta > 2e-4 {
			t.Errorf("%s 往返差 %.6f 秒", tc.name, delta)
		}
		if forward := tc.fwd(tt); !forward.After(tt.Add(-time.Second)) || !forward.Before(tt.Add(time.Minute)) {
			t.Errorf("%s 结果 %v 离开合理区间", tc.name, forward)
		}
	}
	tcb := TCBFromTT(tt)
	if diff := tcb.Sub(TDBFromTCB(tcb)).Seconds(); math.Abs(diff-23.9757) > 0.1 {
		t.Errorf("TCB−TDB=%v 秒, want ≈23.976", diff)
	}
	if delta := math.Abs(TTFromTCB(tcb).Sub(tt).Seconds()); delta > 2e-4 {
		t.Errorf("TT→TCB 往返差 %.6f 秒", delta)
	}
}

func TestBarycentricTimeFacadePreservesScaleAndInstant(t *testing.T) {
	tt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, date := range []time.Time{tt, tt.In(time.FixedZone("UTC+08", 8*3600))} {
		jd := basic.Date2JD(date.UTC())
		for name, pair := range map[string][2]float64{
			"TCG": {TCGMinusTT(date), basic.TCGMinusTTSeconds(jd)},
			"TCB": {TCBMinusTT(date), basic.TCBMinusTTSeconds(jd)},
			"TDB": {TDBMinusTT(date), basic.TDBMinusTTSeconds(jd)},
		} {
			if pair[0] != pair[1] {
				t.Errorf("%s facade=%v, basic=%v", name, pair[0], pair[1])
			}
		}
		if direct, composed := TDBFromTT(date), TDBFromTCB(TCBFromTT(date)); math.Abs(direct.Sub(composed).Seconds()) > 1e-4 {
			t.Errorf("direct TDB=%v, composed=%v", direct, composed)
		}
		if got := TCGFromTT(date); got.Location() != time.UTC || !got.Equal(TCGFromTT(tt)) {
			t.Errorf("TCG depends on input Location: %v", got)
		}
	}
}
