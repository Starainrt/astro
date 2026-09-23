package basic

import (
	"math"
	"testing"
)

// 尺度换算的锚定契约：闰秒表、月度实测 ΔT、恒等式 TT−UTC = ΔT + DUT1、边界与政策、钩子与世代号。

func resetTimeScaleState(t *testing.T) {
	t.Helper()
	SetTTMinusUTCFn(nil)
	SetTimeScaleFuturePolicy(TimeScaleLeapSecond)
	SetDeltaTFn(nil)
	t.Cleanup(func() {
		SetTTMinusUTCFn(nil)
		SetTimeScaleFuturePolicy(TimeScaleLeapSecond)
		SetDeltaTFn(nil)
	})
}

func TestTimeScaleCalendarAnchors(t *testing.T) {
	if got := JDCalc(1972, 1, 1); got != 2441317.5 {
		t.Fatalf("JDCalc(1972,1,1)=%v, want 2441317.5", got)
	}
	if got := JDCalc(2026, 9, 1); got != timeScaleExactEndJDE {
		t.Fatalf("JDCalc(2026,9,1)=%v, want %v", got, timeScaleExactEndJDE)
	}
	if got := deltaTMonthlyJDE[len(deltaTMonthlyJDE)-1]; got != timeScaleExactEndJDE {
		t.Fatalf("月度表末项=%v, want 精确窗口末端 %v", got, timeScaleExactEndJDE)
	}
	if len(deltaTMonthlyJDE) != len(deltaTMonthlyObserved) {
		t.Fatalf("月度表长度不一致: %d vs %d", len(deltaTMonthlyJDE), len(deltaTMonthlyObserved))
	}
}

// 闰时进位量由 ΔT 决定，步长为一小时，校正后 DUT1 必须落回容限内。
func TestTimeScaleLeapHourPolicyCarriesFromDeltaT(t *testing.T) {
	resetTimeScaleState(t)
	SetTimeScaleFuturePolicy(TimeScaleLeapHour)
	base := TTMinusUTCSeconds(timeScaleExactEndJDE)
	if got := TTMinusUTCSeconds(timeScaleExactEndJDE); got != base {
		t.Fatalf("交接处应连续：%v vs %v", got, base)
	}
	prev := utcToTTOffsetSeconds(timeScaleExactEndJDE)
	carries := 0
	firstCarryYear := 0.0
	for year := 2027; year <= 3500; year++ {
		jd := JDCalc(year, 1, 1)
		if dut1 := DUT1Seconds(jd); math.Abs(dut1) >= timeScaleLeapHourSeconds {
			t.Fatalf("%d 年 |DUT1|=%v 越过容限 %v", year, math.Abs(dut1), timeScaleLeapHourSeconds)
		}
		offset := utcToTTOffsetSeconds(jd)
		if offset == prev {
			continue
		}
		step := offset - prev
		if step <= 0 || math.Abs(step-timeScaleLeapHourSeconds) > 1e-9 {
			t.Fatalf("%d 年进位量 %v, want +%v", year, step, timeScaleLeapHourSeconds)
		}
		if carries == 0 {
			firstCarryYear = float64(year)
		}
		carries++
		prev = offset
	}
	if carries < 1 || carries > 3 {
		t.Fatalf("2027..3500 进位次数 %d 不合理", carries)
	}
	if math.Abs(firstCarryYear-2909) > 3 {
		t.Errorf("首次进位应在 2909 年前后, got %.0f", firstCarryYear)
	}
	for _, jd := range []float64{JDCalc(2400, 1, 1), JDCalc(2909, 1, 1), JDCalc(3200, 1, 1)} {
		if got := TT2UTC(UTC2TT(jd)); math.Abs(got-jd) > 1e-9 {
			t.Errorf("jd=%v 往返差 %g 秒", jd, (got-jd)*86400)
		}
	}
}

// 默认与零值均为整数秒外推，首次越限前保持末端偏移。
func TestTimeScaleDefaultFuturePolicyIsLeapSecondRegime(t *testing.T) {
	resetTimeScaleState(t)
	if got := GetTimeScaleFuturePolicy(); got != TimeScaleLeapSecond {
		t.Fatalf("默认未来政策=%v, want TimeScaleLeapSecond", got)
	}
	var zero TimeScaleFuturePolicy
	if zero != TimeScaleLeapSecond {
		t.Fatalf("零值政策=%v, 应即默认", zero)
	}
	base := TTMinusUTCSeconds(timeScaleExactEndJDE)
	for _, jd := range []float64{JDCalc(2027, 1, 1), JDCalc(2029, 7, 1), JDCalc(2035, 7, 1)} {
		if got := utcToTTOffsetSeconds(jd); math.Abs(got-base) > 1e-9 {
			t.Errorf("%v 年默认政策 TT−UTC=%v, want %v（首次触限前应与已宣告偏移相同）",
				jd, got, base)
		}
		if got := DUT1Seconds(jd); math.Abs(got) > utcDUT1ToleranceSeconds+1e-9 {
			t.Errorf("%v 年默认政策 |DUT1|=%v 越过容限", jd, math.Abs(got))
		}
	}
	// 平滑跟随 UT1 的政策必须给出不同的值，否则这条守卫会与它混同。
	SetTimeScaleFuturePolicy(TimeScaleAssumeUT1Tracking)
	if got := utcToTTOffsetSeconds(JDCalc(2035, 7, 1)); math.Abs(got-base) < 1e-3 {
		t.Fatalf("跟随 UT1 政策在 2035 年给出 %v，与已宣告偏移 %v 混同，测试失去区分度", got, base)
	}
	SetTimeScaleFuturePolicy(TimeScaleLeapSecond)
}

func TestTimeScaleLeapTableMagnitudes(t *testing.T) {
	resetTimeScaleState(t)
	if len(utcLeapJDEs) != len(utcLeapOffsets) {
		t.Fatalf("闰秒表长度不一致: %d vs %d", len(utcLeapJDEs), len(utcLeapOffsets))
	}
	for i := range utcLeapJDEs {
		if got := TTMinusUTCSeconds(utcLeapJDEs[i]); got != utcLeapOffsets[i] {
			t.Fatalf("第 %d 条生效时刻 %v 起 TT−UTC=%v, want %v", i, utcLeapJDEs[i], got, utcLeapOffsets[i])
		}
		if i > 0 {
			if descending := utcLeapJDEs[i-1] > utcLeapJDEs[i]; !descending {
				t.Fatalf("闰秒表应按时刻降序, 第 %d 条顺序错误", i)
			}
			if step := utcLeapOffsets[i-1] - utcLeapOffsets[i]; step != 1 {
				t.Fatalf("历史闰秒步长应为 1 秒, 第 %d 条为 %v", i, step)
			}
		}
	}
	if got := TTMinusUTCSeconds(utcLeapJDEs[len(utcLeapJDEs)-1] - 1); got != utcLeapBaseSeconds {
		t.Fatalf("1972 年以前应回退基值 %v, got %v", utcLeapBaseSeconds, got)
	}
}

// 若将来采用"闰时"（|UTC−UT1| 容限放宽到 H 秒、积满一次才校正），本库不需要改代码：
// 注入一条规则驱动的偏移即可。这里把这个能力锁死——交接连续、|DUT1| 不越界、往返一致、可还原。
func TestLeapHourScheduleIsExpressibleThroughHook(t *testing.T) {
	resetTimeScaleState(t)
	const hour = 3600.0
	base := TTMinusUTCSeconds(timeScaleExactEndJDE)
	SetTTMinusUTCFn(func(jd float64) float64 {
		drift := base - DeltaT(jd, true) // 未校正的 DUT1
		steps := math.Floor(math.Abs(drift) / hour)
		if drift < 0 {
			return base + hour*steps
		}
		return base - hour*steps
	})
	if got := TTMinusUTCSeconds(timeScaleExactEndJDE); got != base {
		t.Fatalf("交接处应连续：got %v, want %v", got, base)
	}
	for _, year := range []int{2030, 2100, 2500, 2909, 3000, 3500} {
		dut1 := DUT1Seconds(JDCalc(year, 1, 1))
		if math.Abs(dut1) >= hour {
			t.Errorf("%d 年 |DUT1|=%v 越过容限 %v", year, math.Abs(dut1), hour)
		}
	}
	for _, jd := range []float64{JDCalc(2050, 1, 1), JDCalc(2909, 1, 1), JDCalc(3000, 1, 1)} {
		if got := TT2UTC(UTC2TT(jd)); math.Abs(got-jd) > 1e-9 {
			t.Errorf("jd=%v 往返差 %g 秒", jd, (got-jd)*86400)
		}
	}
	SetTTMinusUTCFn(nil)
	if got := TTMinusUTCSeconds(timeScaleExactEndJDE); got != base {
		t.Fatalf("还原默认后 got %v, want %v", got, base)
	}
}

func TestTimeScaleLeapTableAnchors(t *testing.T) {
	resetTimeScaleState(t)
	cases := []struct {
		year, month int
		want        float64
	}{
		{1972, 1, 42.184},
		{1972, 7, 43.184},
		{2017, 1, 69.184},
		{2026, 4, 69.184},
	}
	for _, item := range cases {
		if got := TTMinusUTCSeconds(JDCalc(item.year, item.month, 1)); got != item.want {
			t.Errorf("TT−UTC(%d-%02d)=%v, want %v", item.year, item.month, got, item.want)
		}
	}
}

func TestTimeScaleMonthlyDeltaTAnchors(t *testing.T) {
	resetTimeScaleState(t)
	// 秒级偏移直接锚定；JDE 级取值受 jd≈2.5e6 的浮点抵消限制（约 2e-5 秒）。
	if got := DeltaT(timeScaleExactEndJDE, true); math.Abs(got-69.1816) > 1e-12 {
		t.Errorf("2026-09 TT−UT1=%v, want 69.1816", got)
	}
	if got := DeltaT(JDCalc(1973, 2, 1), true); math.Abs(got-43.4724) > 1e-12 {
		t.Errorf("1973-02 TT−UT1=%v, want 43.4724", got)
	}
	if got := (UT12TT(timeScaleExactEndJDE) - timeScaleExactEndJDE) * 86400; math.Abs(got-69.1816) > 1e-4 {
		t.Errorf("2026-09 UT12TT 偏移=%v, want 69.1816", got)
	}
	mid := (deltaTMonthlyJDE[10] + deltaTMonthlyJDE[11]) / 2
	want := (deltaTMonthlyObserved[10] + deltaTMonthlyObserved[11]) / 2
	if got := DeltaT(mid, true); math.Abs(got-want) > 1e-12 {
		t.Errorf("中点插值=%v, want %v", got, want)
	}
	if _, ok := deltaTMonthlyAt(deltaTMonthlyJDE[0] - 0.5); ok {
		t.Error("月度表不应覆盖首项之前")
	}
}

// 月度表首与表尾一样要做常值锚定：表外外推在表首处必须与实测值连续，
// 否则表首两侧 DUT1 跳 0.227 秒、UT1 往返在 TT 侧丢掉同样多。
func TestTimeScaleMonthlyDeltaTTableHeadAnchored(t *testing.T) {
	resetTimeScaleState(t)
	head := deltaTMonthlyJDE[0]
	if got := deltaTModelSecondsAtUT(head); math.Abs(got-deltaTMonthlyObserved[0]) > 1e-9 {
		t.Errorf("表首 ΔT=%v, want 实测值 %v", got, deltaTMonthlyObserved[0])
	}
	if gap := deltaTModelSecondsAtUT(head-1e-9) - deltaTModelSecondsAtUT(head+1e-9); math.Abs(gap) > 1e-3 {
		t.Errorf("表首两侧 ΔT 跳变 %g 秒", gap)
	}
	// 表首之前的取值是一个常值平移：平移量等于表首实测值与样条之差，量级 0.227 秒。
	shift := deltaTModelSecondsAtUT(head-1) - deltaTSplineAtJDE(head-1)
	if want := deltaTMonthlyObserved[0] - deltaTSplineAtJDE(head); math.Abs(shift-want) > 1e-9 {
		t.Errorf("表首之前的 ΔT 平移量=%v, want %v", shift, want)
	}
	if math.Abs(shift) > 0.227 {
		t.Errorf("表首之前的 ΔT 平移量 %v 超出预期量级 0.227 秒", shift)
	}
	worst, worstAt := 0.0, 0.0
	for x := head - 0.02; x <= head+0.02; x += 2e-7 {
		for _, err := range []float64{
			(UT12TT(TT2UT1(x)) - x) * 86400,
			(UT12UTC(UTC2UT1(x)) - x) * 86400,
		} {
			if e := math.Abs(err); e > worst {
				worst, worstAt = e, x
			}
		}
	}
	if worst > 1e-6 {
		t.Errorf("表首 ±0.02 天 UT1 往返最大误差 %g 秒 @ %v, want ≤1e-6 秒", worst, worstAt)
	}
}

// 容限是整个窗口的历史事实：按 1e-3 天全窗扫描，抽样点会漏掉月度 ΔT 表首之前的越界。
func TestTimeScaleDUT1Identity(t *testing.T) {
	resetTimeScaleState(t)
	if got := DUT1Seconds(timeScaleExactEndJDE); math.Abs(got-0.0024) > 1e-12 {
		t.Errorf("DUT1(2026-09)=%v, want 0.0024", got)
	}
	worst, worstAt := 0.0, 0.0
	for jd := utcEraStartJDE; jd <= timeScaleExactEndJDE; jd += 1e-3 {
		dut1 := DUT1Seconds(jd)
		if want := TTMinusUTCSeconds(jd) - DeltaT(jd, true); dut1 != want {
			t.Fatalf("恒等式不成立 jd=%v: DUT1=%v want=%v", jd, dut1, want)
		}
		if mag := math.Abs(dut1); mag > worst {
			worst, worstAt = mag, jd
		}
	}
	if worst > utcDUT1ToleranceSeconds {
		t.Errorf("窗口内 |DUT1| 最大 %v 秒 @ %v, 超过现行容限 %v", worst, worstAt, utcDUT1ToleranceSeconds)
	}
	if got := math.Abs(DUT1Seconds(JDCalc(1973, 1, 1))); got > utcDUT1ToleranceSeconds {
		t.Errorf("1973-01-01 |DUT1|=%v 超过容限 %v", got, utcDUT1ToleranceSeconds)
	}
}

// 窗口内 UTC 方向只做整数秒跳变：ΔT 的小数部分不得混进 TT−UTC。
func TestTimeScaleUTCOffsetJumps(t *testing.T) {
	resetTimeScaleState(t)
	const wantJumps = 27
	jumps, maxFractional := 0, 0.0
	prev := utcToTTOffsetSeconds(utcEraStartJDE)
	for jd := utcEraStartJDE + 1e-3; jd <= timeScaleExactEndJDE; jd += 1e-3 {
		current := utcToTTOffsetSeconds(jd)
		if current == prev {
			continue
		}
		jumps++
		if fractional := math.Abs((current - prev) - math.Round(current-prev)); fractional > maxFractional {
			maxFractional = fractional
		}
		prev = current
	}
	if jumps != wantJumps {
		t.Errorf("窗口内跳变次数=%d, want %d", jumps, wantJumps)
	}
	if maxFractional != 0 {
		t.Errorf("跳变的非整数残差最大 %v, want 0", maxFractional)
	}
}

// 闰秒让 UTC2TT 不连续：每个闰秒在 TT 侧都有一段没有原像的区间，宽度恰好 1 秒。
// 宽度只能测到 jd≈2.4e6 的 1 ULP（约 4e-5 秒），故容差取 1e-4 秒。
func TestTimeScaleLeapSecondUnrepresentableBand(t *testing.T) {
	resetTimeScaleState(t)
	const ulpToleranceSeconds = 1e-4
	inBand := func(tt float64) bool { return math.Abs((UTC2TT(TT2UTC(tt))-tt)*86400) > 1e-6 }
	bands := 0
	for _, leap := range utcLeapJDEs {
		before, after := TTMinusUTCSeconds(leap-1), TTMinusUTCSeconds(leap)
		if before == after {
			continue
		}
		if step := after - before; step != 1 {
			t.Errorf("闰秒 %v 的偏移步长=%v, want 1 秒", leap, step)
			continue
		}
		bands++
		lower, upper := leap+(before-0.001)/86400, leap+(after+0.001)/86400
		if inBand(lower) || inBand(upper) {
			t.Errorf("闰秒 %v：带外 1 毫秒处往返应精确", leap)
			continue
		}
		center := leap + (before+0.5)/86400
		first := bisectFloat(lower, center, inBand)
		last := bisectFloat(center, upper, inBand)
		if width := (last - first) * 86400; math.Abs(width-1) > ulpToleranceSeconds {
			t.Errorf("闰秒 %v 的 TT 不可表示带宽=%v 秒, want 1±%v", leap, width, ulpToleranceSeconds)
		}
	}
	if bands != 27 {
		t.Errorf("可测带宽的闰秒 %d 个, want 27", bands)
	}
}

// bisectFloat 定位 f 在 [lo, hi] 上的单调跳变点：f(lo) 为假时返回首个为真的点，反之返回最后一个为真的点。
func bisectFloat(lo, hi float64, f func(float64) bool) float64 {
	loTrue := f(lo)
	for i := 0; i < 80; i++ {
		mid := (lo + hi) / 2
		if mid <= lo || mid >= hi {
			break
		}
		if f(mid) == loTrue {
			lo = mid
		} else {
			hi = mid
		}
	}
	if loTrue {
		return lo
	}
	return hi
}

func TestTimeScaleRoundTrips(t *testing.T) {
	resetTimeScaleState(t)
	dates := [][3]float64{{-1000, 1, 1}, {1000, 6, 15}, {1582, 10, 15}, {1972, 1, 1}, {2000, 1, 1}, {2026, 4, 1}, {2030, 1, 1}, {2100, 1, 1}}
	for _, item := range dates {
		roundTripTimeScales(t, JDCalc(int(item[0]), int(item[1]), item[2]))
	}
	// 闰秒前后各 0.5 秒：TT 侧不可表示带之外必须可逆，样本要压在跳变旁而不是远离它。
	for _, leap := range []float64{utcLeapJDEs[0], utcLeapJDEs[4], utcLeapJDEs[len(utcLeapJDEs)-2]} {
		roundTripTimeScales(t, leap-0.5/86400)
		roundTripTimeScales(t, leap+0.5/86400)
	}
}

func roundTripTimeScales(t *testing.T, jd float64) {
	t.Helper()
	if got := TT2UT1(UT12TT(jd)); math.Abs(got-jd) > 1e-9 {
		t.Errorf("UT1 往返 %v: 差 %g 秒", jd, (got-jd)*86400)
	}
	if got := TT2UTC(UTC2TT(jd)); math.Abs(got-jd) > 1e-9 {
		t.Errorf("UTC 往返 %v: 差 %g 秒", jd, (got-jd)*86400)
	}
	if got, want := UTC2UT1(jd), TT2UT1(UTC2TT(jd)); got != want {
		t.Errorf("UTC2UT1 与复合式不一致 %v: %v vs %v", jd, got, want)
	}
	if got := UT12UTC(UTC2UT1(jd)); math.Abs(got-jd) > 1e-9 {
		t.Errorf("UTC↔UT1 往返 %v: 差 %g 秒", jd, (got-jd)*86400)
	}
}

func TestTimeScaleCivilEqualsUT1Before1972(t *testing.T) {
	resetTimeScaleState(t)
	for _, jd := range []float64{JDCalc(1900, 1, 1), JDCalc(1971, 12, 31)} {
		if UTC2TT(jd) != UT12TT(jd) {
			t.Errorf("1972 前 UTC2TT 应等于 UT12TT, jd=%v", jd)
		}
		if got := DUT1Seconds(jd); got != 0 {
			t.Errorf("1972 前 DUT1 应为 0, jd=%v got=%v", jd, got)
		}
	}
}

// 跟随 UT1 政策：窗口外保留末端实测 DUT1，TT−UTC 随外推 ΔT 平滑变化，交界处连续。
func TestTimeScaleFuturePolicy(t *testing.T) {
	resetTimeScaleState(t)
	SetTimeScaleFuturePolicy(TimeScaleAssumeUT1Tracking)
	beyond := timeScaleExactEndJDE + 1
	frozenDUT1 := dut1AtTimeScaleExactEnd()
	if got := DUT1Seconds(beyond); math.Abs(got-frozenDUT1) > 1e-12 {
		t.Errorf("跟随 UT1 政策下窗口外 DUT1 应冻结在 %v, got=%v", frozenDUT1, got)
	}
	if got, want := utcToTTOffsetSeconds(beyond), ut1ToTTOffsetSeconds(beyond)+frozenDUT1; got != want {
		t.Errorf("跟随 UT1 政策下窗口外 TT−UTC=%v, want %v", got, want)
	}
	before := utcToTTOffsetSeconds(timeScaleExactEndJDE - 1e-9)
	after := utcToTTOffsetSeconds(timeScaleExactEndJDE + 1e-9)
	if math.Abs(after-before) > 1e-6 {
		t.Errorf("跟随 UT1 政策在窗口末端留下 %g 秒跳变", after-before)
	}
	frozen := TTMinusUTCSeconds(timeScaleExactEndJDE)
	SetTimeScaleFuturePolicy(TimeScaleFreezeUTCOffset)
	if got, want := UTC2TT(beyond), beyond+frozen/86400; got != want {
		t.Errorf("冻结政策下 UTC2TT=%v, want %v", got, want)
	}
	if got := utcToTTOffsetSeconds(beyond); math.Abs(got-frozen) > 1e-12 {
		t.Errorf("冻结政策下 TT−UTC=%v, want %v", got, frozen)
	}
	near, far := DUT1Seconds(beyond), DUT1Seconds(beyond+365)
	if !(far < near) {
		t.Errorf("冻结政策下 DUT1 应随 ΔT 增大而下降: %v → %v", near, far)
	}
}

// UT1Civil 在全时轴生效，含历史闰秒附近。
func TestTimeScaleUT1CivilPolicy(t *testing.T) {
	resetTimeScaleState(t)
	SetTimeScaleFuturePolicy(TimeScaleUT1Civil)
	for _, jd := range []float64{
		JDCalc(1950, 1, 1),
		JDCalc(1972, 1, 1),
		JDCalc(1999, 1, 1),
		JDCalc(2017, 1, 1),
		JDCalc(2020, 1, 1),
		timeScaleExactEndJDE,
		JDCalc(2035, 7, 1),
		JDCalc(2200, 1, 1),
	} {
		if got := DUT1Seconds(jd); got != 0 {
			t.Errorf("jd=%v DUT1=%v, want 0", jd, got)
		}
		if got, want := utcToTTOffsetSeconds(jd), ut1ToTTOffsetSeconds(jd); got != want {
			t.Errorf("jd=%v TT−UTC=%v, want ΔT=%v", jd, got, want)
		}
		if got := UTC2UT1(jd); math.Abs(got-jd) > 1e-9 {
			t.Errorf("jd=%v UTC2UT1 差 %g 秒，应为恒等映射", jd, (got-jd)*86400)
		}
		if got := TT2UTC(UTC2TT(jd)); math.Abs(got-jd) > 1e-9 {
			t.Errorf("jd=%v 往返差 %g 秒", jd, (got-jd)*86400)
		}
		if got, want := TT2UT1(UTC2TT(jd)), UTC2UT1(jd); got != want {
			t.Errorf("jd=%v TT2UT1(UTC2TT)=%v, UTC2UT1=%v（该政策下两条路应合流）", jd, got, want)
		}
	}
	// 实测年代必须让位给 ΔT：1999 年闰秒表是 64.184 秒，政策下应等于 ΔT 且与 2017 年闰秒前后连续。
	if got := utcToTTOffsetSeconds(JDCalc(1999, 1, 1)); math.Abs(got-64.184) < 1e-9 {
		t.Errorf("1999 年 TT−UTC=%v，仍是闰秒表值，政策未覆盖实测窗口", got)
	}
	before := utcToTTOffsetSeconds(utcLeapJDEs[0] - 1e-6)
	after := utcToTTOffsetSeconds(utcLeapJDEs[0] + 1e-6)
	if math.Abs(after-before) > 1e-6 {
		t.Errorf("2017-01-01 历史闰秒处出现 %g 秒阶跃，该政策下不应有阶跃", after-before)
	}
	// 注入的 TT−UTC 覆盖仍然优先于政策。
	SetTTMinusUTCFn(func(float64) float64 { return 42.184 })
	if got := utcToTTOffsetSeconds(JDCalc(2035, 7, 1)); math.Abs(got-42.184) > 1e-9 {
		t.Errorf("覆盖后 TT−UTC=%v, want 42.184", got)
	}
}

// 现行闰秒政策：偏移只能是末端已宣告值加整数秒；未阶跃 UT1−UTC 越出 ±0.9 秒时补 ±1 秒；
// 首次阶跃之前与冻结政策逐值相同；ΔT 反向漂移时走负闰秒。
func TestTimeScaleLeapSecondPolicy(t *testing.T) {
	resetTimeScaleState(t)
	SetTimeScaleFuturePolicy(TimeScaleLeapSecond)
	base := TTMinusUTCSeconds(timeScaleExactEndJDE)
	unstepped := func(jd float64) float64 { return base - ut1ToTTOffsetSeconds(jd) }

	low, high := JDCalc(2026, 1, 1), JDCalc(2200, 1, 1)
	if unstepped(high) > -utcDUT1ToleranceSeconds {
		t.Fatal("2026..2200 内没有触限时刻，测试前提失效")
	}
	for i := 0; i < 200; i++ {
		mid := (low + high) / 2
		if unstepped(mid) <= -utcDUT1ToleranceSeconds {
			high = mid
		} else {
			low = mid
		}
	}
	crossing := (low + high) / 2
	for _, sample := range []struct {
		jd   float64
		want float64
	}{
		{crossing - 30, base},
		{crossing + 30, base + 1},
		{JDCalc(2029, 7, 1), base},
		{JDCalc(2035, 7, 1), base},
	} {
		got := utcToTTOffsetSeconds(sample.jd)
		if math.Abs(got-sample.want) > 1e-9 {
			t.Errorf("jd=%v TT−UTC=%v, want %v", sample.jd, got, sample.want)
		}
		if steps := got - base; math.Abs(steps-math.Round(steps)) > 1e-9 {
			t.Errorf("jd=%v 相对已宣告偏移 %v 不是整数秒", sample.jd, steps)
		}
		if dut1 := DUT1Seconds(sample.jd); math.Abs(dut1) > utcDUT1ToleranceSeconds+1e-9 {
			t.Errorf("jd=%v |DUT1|=%v 越过容限 %v", sample.jd, math.Abs(dut1), utcDUT1ToleranceSeconds)
		}
	}
	// 逐年扫到 2400：偏移始终是 base 加整数秒，DUT1 全程留在容限内，阶跃只升不降。
	prev := utcToTTOffsetSeconds(JDCalc(2027, 1, 1))
	for year := 2028; year <= 2400; year++ {
		jd := JDCalc(year, 1, 1)
		offset := utcToTTOffsetSeconds(jd)
		if steps := offset - base; math.Abs(steps-math.Round(steps)) > 1e-9 {
			t.Fatalf("%d 年相对偏移 %v 不是整数秒", year, steps)
		}
		if offset < prev {
			t.Fatalf("%d 年偏移从 %v 降到 %v（本库 ΔT 外推下不应出现负闰秒）", year, prev, offset)
		}
		prev = offset
		if dut1 := DUT1Seconds(jd); math.Abs(dut1) > utcDUT1ToleranceSeconds+1e-9 {
			t.Fatalf("%d 年 |DUT1|=%v 越过容限", year, math.Abs(dut1))
		}
	}
	// 注入 ΔT 随时间下降的模型：未阶跃 UT1−UTC 变为上升，必须走负闰秒。
	SetDeltaTFn(func(jd float64, isJd bool) float64 {
		year := jd
		if isJd {
			year = 2026 + (jd-timeScaleExactEndJDE)/365.25
		}
		return 69.1816 - 0.2*(year-2026)
	})
	jd := JDCalc(2032, 1, 1)
	if got := utcToTTOffsetSeconds(jd); math.Abs(got-(base-1)) > 1e-9 {
		t.Errorf("反向漂移下 TT−UTC=%v, want %v（负闰秒）", got, base-1)
	}
	if dut1 := DUT1Seconds(jd); math.Abs(dut1) > utcDUT1ToleranceSeconds+1e-9 {
		t.Errorf("反向漂移下 |DUT1|=%v 越过容限", math.Abs(dut1))
	}
	SetDeltaTFn(nil)
}

func TestTimeScaleHooksAndGeneration(t *testing.T) {
	resetTimeScaleState(t)
	before := deltaTGenerationValue()
	SetTTMinusUTCFn(func(float64) float64 { return 100 })
	if deltaTGenerationValue() == before {
		t.Error("SetTTMinusUTCFn 应递增世代号")
	}
	if got := TTMinusUTCSeconds(JDCalc(2000, 1, 1)); got != 100 {
		t.Errorf("覆盖后 TT−UTC=%v, want 100", got)
	}
	SetTTMinusUTCFn(nil)
	if got := TTMinusUTCSeconds(JDCalc(2000, 1, 1)); math.Abs(got-64.184) > 1e-9 {
		t.Errorf("恢复后 TT−UTC=%v, want 64.184", got)
	}

	before = deltaTGenerationValue()
	SetDeltaTFn(func(float64, bool) float64 { return 100 })
	if deltaTGenerationValue() == before {
		t.Error("SetDeltaTFn 应递增世代号")
	}
	jd := JDCalc(2000, 1, 1)
	if got := ut1ToTTOffsetSeconds(jd); math.Abs(got-100) > 1e-12 {
		t.Errorf("注入 ΔT 后 UT1 偏移=%v, want 100", got)
	}
	SetDeltaTFn(nil)
	if got := ut1ToTTOffsetSeconds(jd); math.Abs(got-63.83) > 0.05 {
		t.Errorf("恢复后 UT12TT 偏移=%v, 应回到月度实测值附近", got)
	}

	before = deltaTGenerationValue()
	SetTimeScaleFuturePolicy(TimeScaleFreezeUTCOffset)
	if deltaTGenerationValue() == before {
		t.Error("SetTimeScaleFuturePolicy 应递增世代号")
	}
}

// JD 0（−4713-11-24）是合法时刻，不是零值哨兵：TT2UTC(0) 必须与 TT2UT1(0) 给出同一个有限值并互逆。
func TestTimeScaleZeroInstantIsNotSentinel(t *testing.T) {
	resetTimeScaleState(t)
	utc := TT2UTC(0)
	if math.IsNaN(utc) || math.IsInf(utc, 0) {
		t.Fatalf("TT2UTC(0)=%v, want 有限值", utc)
	}
	if got := TT2UT1(0); math.Abs(utc-got) > 1e-9 {
		t.Errorf("1972 前民用时标即 UT1：TT2UTC(0)=%v, TT2UT1(0)=%v（差 %g 秒）", utc, got, (utc-got)*86400)
	}
	if back := UTC2TT(utc); math.Abs(back) > 1e-9 {
		t.Errorf("UTC2TT(TT2UTC(0))=%v, want 0（差 %g 秒）", back, back*86400)
	}
}

// DeltaT 的第一个参数按 julianDay 解释为 UT 儒略日或十进制年；口径给错时自变量远在模型跨度之外，
// 必须返回 NaN 而不是继续外推成天文数字。
func TestDeltaTArgumentDomain(t *testing.T) {
	resetTimeScaleState(t)
	for _, year := range []float64{-40000, -2000, -500, 0, 1000.5, 2000, 3000, 40000} {
		if got := DeltaT(year, false); math.IsNaN(got) {
			t.Errorf("DeltaT(%v, 十进制年)=NaN, 应给出模型值", year)
		}
	}
	for _, jd := range []float64{-1e7, JDCalc(-2000, 1, 1), JDCalc(0, 1, 1), JDCalc(2000, 1, 1), JDCalc(5000, 1, 1), 1e7} {
		if got := DeltaT(jd, true); math.IsNaN(got) {
			t.Errorf("DeltaT(%v, 儒略日)=NaN, 应给出模型值", jd)
		}
	}
	if got := DeltaT(JDCalc(2026, 3, 1), false); !math.IsNaN(got) {
		t.Errorf("儒略日当十进制年应越界返回 NaN, got %v", got)
	}
	for _, jd := range []float64{1e8, -1e8} {
		if got := DeltaT(jd, true); !math.IsNaN(got) {
			t.Errorf("DeltaT(%v, 儒略日)=%v, want NaN", jd, got)
		}
	}
	for _, value := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		for _, julianDay := range []bool{false, true} {
			if got := DeltaT(value, julianDay); !math.IsNaN(got) {
				t.Errorf("DeltaT(%v, %v)=%v, want NaN", value, julianDay, got)
			}
		}
	}
}

func TestTimeScaleSteppedPoliciesRejectInvalidDeltaT(t *testing.T) {
	resetTimeScaleState(t)
	for _, policy := range []TimeScaleFuturePolicy{TimeScaleLeapSecond, TimeScaleLeapHour} {
		SetTimeScaleFuturePolicy(policy)
		for _, tc := range []struct {
			name  string
			jd    float64
			model func(float64, bool) float64
		}{
			{"nan-model", JDCalc(2030, 1, 1), func(float64, bool) float64 { return math.NaN() }},
			{"positive-infinity-model", JDCalc(2030, 1, 1), func(float64, bool) float64 { return math.Inf(1) }},
			{"negative-infinity-model", JDCalc(2030, 1, 1), func(float64, bool) float64 { return math.Inf(-1) }},
			{"outside-model-range", 1e8, nil},
		} {
			SetDeltaTFn(tc.model)
			for name, got := range map[string]float64{
				"offset": utcToTTOffsetSeconds(tc.jd),
				"UTC2TT": UTC2TT(tc.jd),
				"TT2UTC": TT2UTC(tc.jd),
				"DUT1":   DUT1Seconds(tc.jd),
			} {
				if !math.IsNaN(got) {
					t.Errorf("policy=%v %s %s=%v, want NaN", policy, tc.name, name, got)
				}
			}
		}
	}
}

func TestTimeScalePropagatesNaN(t *testing.T) {
	resetTimeScaleState(t)
	nan := math.NaN()
	for name, got := range map[string]float64{
		"UTC2TT":  UTC2TT(nan),
		"UT12TT":  UT12TT(nan),
		"TT2UTC":  TT2UTC(nan),
		"TT2UT1":  TT2UT1(nan),
		"UTC2UT1": UTC2UT1(nan),
		"UT12UTC": UT12UTC(nan),
		"DUT1":    DUT1Seconds(nan),
	} {
		if !math.IsNaN(got) {
			t.Errorf("%s(NaN)=%v, want NaN", name, got)
		}
	}
}

// 注入的 TT−UTC 覆盖对窗口之后的查询日期同样生效：未来政策只描述未注入时的形状，
// 否则外部时标模型与实际换算在窗口末端之后会各说各话。
func TestTimeScaleOverrideAppliesBeyondExactWindow(t *testing.T) {
	resetTimeScaleState(t)
	jd := JDCalc(2027, 1, 1)
	SetTTMinusUTCFn(func(at float64) float64 {
		if at >= jd {
			return 70.184
		}
		return 69.184
	})
	for _, policy := range []TimeScaleFuturePolicy{
		TimeScaleAssumeUT1Tracking, TimeScaleFreezeUTCOffset, TimeScaleLeapHour,
		TimeScaleUT1Civil, TimeScaleLeapSecond,
	} {
		SetTimeScaleFuturePolicy(policy)
		if got := utcToTTOffsetSeconds(jd); math.Abs(got-70.184) > 1e-12 {
			t.Errorf("policy=%v: UTC2TT 使用 %.6f 秒，want 70.184", policy, got)
		}
		// jd 本身只有约 4e-5 秒的分辨率，往返差按秒比较。
		if got := (TT2UTC(UTC2TT(jd)) - jd) * 86400; math.Abs(got) > 1e-4 {
			t.Errorf("policy=%v: 往返差 %.9f 秒", policy, got)
		}
		if got := DUT1Seconds(jd); math.Abs(got-(70.184-DeltaT(jd, true))) > 1e-9 {
			t.Errorf("policy=%v: DUT1=%.6f 秒，want %.6f", policy, got, 70.184-DeltaT(jd, true))
		}
	}
	SetTTMinusUTCFn(nil)
	SetTimeScaleFuturePolicy(TimeScaleFreezeUTCOffset)
	if got := utcToTTOffsetSeconds(jd); math.Abs(got-69.184) > 1e-12 {
		t.Errorf("未注入时冻结政策应生效：UTC2TT 使用 %.6f 秒", got)
	}
}

// 默认 TT−UTC 函数必须忽略注入的覆盖：它是"取默认口径"的唯一入口。
func TestTTMinusUTCSecondsDefaultIgnoresOverride(t *testing.T) {
	resetTimeScaleState(t)
	jd := timeScaleExactEndJDE
	want := TTMinusUTCSecondsDefault(jd)
	if math.Abs(want-69.184) > 1e-9 {
		t.Fatalf("默认 TT−UTC(窗口末端)=%v, want 69.184", want)
	}
	SetTTMinusUTCFn(func(float64) float64 { return 42.184 })
	if got := TTMinusUTCSeconds(jd); math.Abs(got-42.184) > 1e-9 {
		t.Fatalf("覆盖后 TTMinusUTCSeconds=%v, want 42.184", got)
	}
	if got := TTMinusUTCSecondsDefault(jd); math.Abs(got-want) > 1e-12 {
		t.Fatalf("覆盖后默认 TT−UTC=%v, want %v（应忽略覆盖）", got, want)
	}
}
