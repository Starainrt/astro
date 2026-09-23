// 根包时标封面的契约：time.Time 进出、按 UTC 时刻取值、往返自洽。
package astro_test

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro"
)

func TestTimeScalePublicFacade(t *testing.T) {
	utc := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	if got := astro.DUT1(utc); math.Abs(got-0.051) > 1e-3 {
		t.Errorf("DUT1=%v 秒, want 0.051", got)
	}
	if got := astro.TTFromUTC(utc).Sub(utc).Seconds(); math.Abs(got-69.184) > 1e-3 {
		t.Errorf("TT−UTC=%v 秒, want 69.184", got)
	}
	for _, date := range []time.Time{
		time.Date(1960, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(1972, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2000, 6, 30, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2035, 1, 1, 0, 0, 0, 0, time.UTC),
	} {
		if got := astro.UTCFromUT1(astro.UT1FromUTC(date)).Sub(date).Seconds(); math.Abs(got) > 1e-3 {
			t.Errorf("UT1 往返 %v: 差 %g 秒", date, got)
		}
		if got := astro.UTCFromTT(astro.TTFromUTC(date)).Sub(date).Seconds(); math.Abs(got) > 1e-3 {
			t.Errorf("TT 往返 %v: 差 %g 秒", date, got)
		}
	}
	before := time.Date(1960, 1, 1, 0, 0, 0, 0, time.UTC)
	if got := astro.UT1FromUTC(before).Sub(before).Seconds(); math.Abs(got) > 1e-3 {
		t.Errorf("1972 前 UT1 应等于民用时刻, 差 %g 秒", got)
	}
	cst := utc.In(time.FixedZone("CST", 8*3600))
	if got, want := astro.UT1FromUTC(cst), astro.UT1FromUTC(utc); !got.Equal(want) {
		t.Errorf("时区不应改变结果: %v vs %v", got, want)
	}
}

func TestTimeScaleLabelSelection(t *testing.T) {
	utc := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	if got := astro.LabelIn(astro.TimeScaleUTC, utc); !got.Equal(utc) {
		t.Errorf("UTC 时刻不应改动: %v", got)
	}
	ut1 := astro.LabelIn(astro.TimeScaleUT1, utc)
	if delta := ut1.Sub(utc).Seconds(); math.Abs(delta-0.051) > 5e-4 {
		t.Errorf("UT1 时刻应领先民用时刻约 DUT1: got %g 秒", delta)
	}
	if ut1.Location() != time.UTC {
		t.Errorf("UT1 时刻应带 UTC 时区（无时区语义）, got %v", ut1.Location())
	}
	if zero := astro.LabelIn(0, utc); !zero.Equal(utc) {
		t.Errorf("零值时标应为 UTC: %v", zero)
	}
}

func TestTimeScaleModelInjectionFacade(t *testing.T) {
	utc := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	if astro.TTMinusUTC() != nil {
		t.Fatalf("默认 TT−UTC 应来自内置闰秒表（覆盖函数为 nil）")
	}
	baseTTMinusUTC := astro.TTFromUTC(utc).Sub(utc).Seconds()
	baseDUT1 := astro.DUT1(utc)
	if math.Abs(baseTTMinusUTC-69.184) > 1e-3 {
		t.Fatalf("默认 TT−UTC = %g 秒, want 69.184", baseTTMinusUTC)
	}

	astro.SetTTMinusUTC(func(float64) float64 { return 42.184 })
	defer astro.SetTTMinusUTC(nil)
	if astro.TTMinusUTC() == nil {
		t.Fatalf("注入后 TTMinusUTC() 不应为 nil")
	}
	if got := astro.TTFromUTC(utc).Sub(utc).Seconds(); math.Abs(got-42.184) > 1e-3 {
		t.Fatalf("注入后 TT−UTC = %g 秒, want 42.184", got)
	}
	if got := astro.UTCFromTT(astro.TTFromUTC(utc)).Sub(utc).Seconds(); math.Abs(got) > 1e-3 {
		t.Fatalf("注入后 TT 往返差 %g 秒", got)
	}
	// DUT1 = (TT−UTC) − ΔT，ΔT 未变，因此只随 TT−UTC 口径平移。
	if got, want := astro.DUT1(utc), baseDUT1+42.184-baseTTMinusUTC; math.Abs(got-want) > 1e-3 {
		t.Fatalf("注入后 DUT1 = %g 秒, want %g", got, want)
	}

	astro.SetTTMinusUTC(nil)
	if astro.TTMinusUTC() != nil {
		t.Fatalf("恢复后 TTMinusUTC() 应为 nil")
	}
	if got := astro.TTFromUTC(utc).Sub(utc).Seconds(); math.Abs(got-baseTTMinusUTC) > 1e-3 {
		t.Fatalf("恢复后 TT−UTC = %g 秒, want %g", got, baseTTMinusUTC)
	}
	if got := astro.DUT1(utc); math.Abs(got-baseDUT1) > 1e-6 {
		t.Fatalf("恢复后 DUT1 = %g 秒, want %g", got, baseDUT1)
	}

	// ΔT 一侧同样对称：默认模型可通过 DefaultDeltaT 取回。
	if astro.DefaultDeltaT() == nil {
		t.Fatalf("DefaultDeltaT() 不应为 nil")
	}
	astro.SetDeltaT(nil)
	if astro.DeltaT() == nil {
		t.Fatalf("传 nil 后 DeltaT() 应回到默认模型")
	}
}

func TestTimeScaleDefaultModelAndFuturePolicyFacade(t *testing.T) {
	astro.SetTTMinusUTC(nil)
	astro.SetTimeScaleFuturePolicy(astro.TimeScaleAssumeUT1Tracking)
	defer func() {
		astro.SetTTMinusUTC(nil)
		astro.SetTimeScaleFuturePolicy(astro.TimeScaleAssumeUT1Tracking)
	}()

	// 注入覆盖后，默认模型仍给出内置闰秒表口径。
	astro.SetTTMinusUTC(func(float64) float64 { return 42.184 })
	defaultFn := astro.DefaultTTMinusUTC()
	if defaultFn == nil {
		t.Fatalf("DefaultTTMinusUTC() 不应为 nil")
	}
	jd := 2460000.5
	if got := defaultFn(jd); math.Abs(got-69.184) > 1e-9 {
		t.Fatalf("DefaultTTMinusUTC()(JD)=%v, want 69.184（忽略覆盖）", got)
	}
	if got := astro.TTMinusUTC()(jd); math.Abs(got-42.184) > 1e-9 {
		t.Fatalf("TTMinusUTC()(JD)=%v, want 42.184", got)
	}
	astro.SetTTMinusUTC(nil)

	// 未来政策：根包与 basic 同一份状态；切换后窗口外的 TT−UTC 取值改变，恢复后回到原值。
	if got := astro.GetTimeScaleFuturePolicy(); got != astro.TimeScaleAssumeUT1Tracking {
		t.Fatalf("默认政策=%v, want TimeScaleAssumeUT1Tracking", got)
	}
	tail := time.Date(2035, 1, 1, 0, 0, 0, 0, time.UTC)
	tracking := astro.TTFromUTC(tail).Sub(tail).Seconds()

	astro.SetTimeScaleFuturePolicy(astro.TimeScaleFreezeUTCOffset)
	if got := astro.GetTimeScaleFuturePolicy(); got != astro.TimeScaleFreezeUTCOffset {
		t.Fatalf("切换后政策=%v, want TimeScaleFreezeUTCOffset", got)
	}
	frozen := astro.TTFromUTC(tail).Sub(tail).Seconds()
	// jd≈2.46e6 上 float64 只能还原到几十微秒，故按毫秒级比较。
	if math.Abs(frozen-69.184) > 1e-3 {
		t.Fatalf("冻结政策下 TT−UTC=%v, want 69.184", frozen)
	}
	if math.Abs(tracking-frozen) < 1e-6 {
		t.Fatalf("两种政策在 2035 应给出不同 TT−UTC：tracking=%v frozen=%v", tracking, frozen)
	}

	astro.SetTimeScaleFuturePolicy(astro.TimeScaleAssumeUT1Tracking)
	if got := astro.TTFromUTC(tail).Sub(tail).Seconds(); math.Abs(got-tracking) > 1e-3 {
		t.Fatalf("恢复政策后 TT−UTC=%v, want %v", got, tracking)
	}
}
