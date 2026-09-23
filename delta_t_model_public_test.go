// 根包命名 ΔT 模型的契约：实测段优先、表外走模型、标记可回读。
package astro_test

import (
	"math"
	"testing"

	"github.com/starainrt/astro"
	"github.com/starainrt/astro/basic"
)

func TestDeltaTModelFacade(t *testing.T) {
	astro.SetDeltaT(nil)
	defer astro.SetDeltaT(nil)

	if !astro.SetDeltaTModel(astro.DeltaTModelNASACanon2006, true) {
		t.Fatal("NASA canon 模型应可安装")
	}
	if model, keep := astro.GetDeltaTModel(); model != astro.DeltaTModelNASACanon2006 || !keep {
		t.Fatalf("模型标记 = (%q, %v)", model, keep)
	}
	// 实测段（2026.2 年落在逐月表内）必须给实测值，与内置默认一致。
	withModel := astro.DeltaT()(2026.2, false)
	astro.SetDeltaT(nil)
	builtin := astro.DeltaT()(2026.2, false)
	if math.Abs(withModel-builtin) > 1e-9 {
		t.Fatalf("实测段优先失效：模型档 %.6f vs 内置 %.6f", withModel, builtin)
	}

	// 表外（公元 0 年）走模型：canon 比未加配对项的 E–M 小约 49.4 秒。
	canon := astro.DeltaTModelSeconds(astro.DeltaTModelNASACanon2006, 1721060.5, false)
	espenak := astro.DeltaTModelSeconds(astro.DeltaTModelEspenakMeeus2006, 1721060.5, false)
	if diff := espenak - canon; math.Abs(diff-49.43) > 0.05 {
		t.Fatalf("公元 0 年配对修正 = %.3f s, want 49.43 s", diff)
	}
	if !astro.SetDeltaTModel(astro.DeltaTModelMS2004, true) {
		t.Fatal("M&S2004 模型应可安装")
	}
	if got := astro.DeltaT()(0.0, false); math.Abs(got-canon) < 10 {
		t.Fatalf("公元 0 年 M&S2004 = %.3f，应与 canon %.3f 明显不同", got, canon)
	}

	// 手工注入报 manual；未知模型被拒绝且不改动现状。
	astro.SetDeltaT(func(float64, bool) float64 { return 0 })
	if model, keep := astro.GetDeltaTModel(); model != astro.DeltaTModelManual || keep {
		t.Fatalf("手工注入标记 = (%q, %v)", model, keep)
	}
	if astro.SetDeltaTModel(astro.DeltaTModel("nope"), true) {
		t.Fatal("未知模型应被拒绝")
	}
	if model, _ := astro.GetDeltaTModel(); model != astro.DeltaTModelManual {
		t.Fatalf("被拒绝后模型不应改动，got %q", model)
	}
}

func TestDefaultDeltaTConfigurationRoundTrip(t *testing.T) {
	astroResetDefaultModel(t)
	for _, year := range []int{1800, 2026, 2100} {
		jd := basic.JDCalc(year, 3, 1)
		before := astro.DeltaT()(jd, true)
		model, keep := astro.GetDeltaTModel()
		if model != astro.DeltaTModelDefault || !keep {
			t.Fatalf("default model metadata=(%q,%t)", model, keep)
		}
		if !astro.SetDeltaTModel(astro.DeltaTModelMS2004, false) || !astro.SetDeltaTModel(model, keep) {
			t.Fatal("model selection failed")
		}
		if got := astro.DeltaT()(jd, true); got != before {
			t.Errorf("year %d: before %.12f after %.12f", year, before, got)
		}
		if got := astro.DeltaTModelSeconds(model, jd, keep); got != before {
			t.Errorf("year %d: model evaluation %.12f, active %.12f", year, got, before)
		}
	}
}

func astroResetDefaultModel(t *testing.T) {
	t.Helper()
	astro.SetDeltaT(nil)
	t.Cleanup(func() { astro.SetDeltaT(nil) })
}
