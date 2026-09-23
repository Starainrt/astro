package basic

import (
	"math"
	"testing"
)

// ΔT 命名模型的契约：已发表锚点、分段连续性、实测段优先的混合语义与模型标记。

func TestDeltaTModelPublishedAnchors(t *testing.T) {
	resetTimeScaleState(t)
	check := func(name string, year, want, tolerance float64) {
		t.Helper()
		if got := DeltaT(year, false); math.Abs(got-want) > tolerance {
			t.Fatalf("%s ΔT(%.5f)=%.6f, want %.6f±%g", name, year, got, want, tolerance)
		}
	}

	if !SetDeltaTModel(DeltaTModelEspenakMeeus2006, false) {
		t.Fatal("Espenak–Meeus 模型应可安装")
	}
	check("E-M", 2017.64, 70.34, 0.01)
	check("E-M", 2024.27, 74.03, 0.01)
	check("E-M", 2100, 202.7, 0.05)

	if !SetDeltaTModel(DeltaTModelMS2004, false) {
		t.Fatal("M&S2004 模型应可安装")
	}
	check("M&S2004", -500, 17203.7, 0.05)
	check("M&S2004", 2100, 230.9, 0.05)

	if !SetDeltaTModel(DeltaTModelEspenakMeeus2006, false) {
		t.Fatal("Espenak–Meeus 模型应可安装")
	}
	// 分段边界：原表各段是独立拟合的，边界处本身留有台阶（实测最大 ≤0.3 s，1600 处约 0.25 s），
	// 这里只守住这个量级，用来抓 Horner 展开之类的结构性错误。
	for _, knot := range []float64{-500, 500, 1600, 1700, 1800, 1860, 1900, 1920, 1941, 1961, 1986, 2005, 2050, 2150} {
		before, after := DeltaT(knot-1e-6, false), DeltaT(knot+1e-6, false)
		if diff := math.Abs(after - before); diff > 0.3 {
			t.Errorf("分段边界 %.0f 台阶过大：%.6f vs %.6f（差 %g）", knot, before, after, diff)
		}
	}
}

func TestDeltaTModelKeepsObservedSpan(t *testing.T) {
	resetTimeScaleState(t)
	jd := JDCalc(2026, 3, 1)
	observed := DeltaT(jd, true)

	if !SetDeltaTModel(DeltaTModelEspenakMeeus2006, true) {
		t.Fatal("Espenak–Meeus 模型应可安装")
	}
	if got := DeltaT(jd, true); math.Abs(got-observed) > 1e-12 {
		t.Fatalf("实测段应优先：got %.9f, want %.9f", got, observed)
	}
	pure := DeltaTModelSeconds(DeltaTModelEspenakMeeus2006, jd, false)
	if math.Abs(pure-observed) < 5 {
		t.Fatalf("纯模型值 %.3f 应与实测 %.3f 明显不同（否则测不到混合语义）", pure, observed)
	}
	if model, keep := GetDeltaTModel(); model != DeltaTModelEspenakMeeus2006 || !keep {
		t.Fatalf("模型标记 = (%q, %v)", model, keep)
	}
	future := JDCalc(2100, 1, 1)
	if got, want := DeltaT(future, true), DeltaTModelSeconds(DeltaTModelEspenakMeeus2006, future, false); math.Abs(got-want) > 1e-9 {
		t.Fatalf("表外应走模型：got %.9f, want %.9f", got, want)
	}

	// SMH2016 混合档必须与内置默认逐位一致。
	if !SetDeltaTModel(DeltaTModelSMH2016, true) {
		t.Fatal("SMH2016 模型应可安装")
	}
	if got := DeltaT(jd, true); got != observed {
		t.Fatalf("SMH2016 混合档 %.9f 与内置默认 %.9f 不一致", got, observed)
	}
	if got := DeltaT(future, true); got != DefaultDeltaTv2(future, true) {
		t.Fatalf("SMH2016 混合档表外 %.9f 与内置默认 %.9f 不一致", got, DefaultDeltaTv2(future, true))
	}

	// 手工注入的任意函数标记为 manual；恢复 nil 回到内置默认。
	SetDeltaTFn(func(float64, bool) float64 { return 0 })
	if model, keep := GetDeltaTModel(); model != DeltaTModelManual || keep {
		t.Fatalf("手工注入标记 = (%q, %v)", model, keep)
	}
	SetDeltaTFn(nil)
	if model, keep := GetDeltaTModel(); model != DeltaTModelDefault || !keep {
		t.Fatalf("恢复默认标记 = (%q, %v)", model, keep)
	}
	if SetDeltaTModel(DeltaTModel("nope"), true) {
		t.Fatal("未知模型应被拒绝")
	}
	if model, _ := GetDeltaTModel(); model != DeltaTModelDefault {
		t.Fatalf("被拒绝后模型不应改动，got %q", model)
	}
}
