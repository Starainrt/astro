package eclipse

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

// 三个 info 都带 provenance：GreatestJDE（力学时）经 TT2UTC 复现民用时刻字段；
// DeltaTSeconds 是 TT−UT1，不能用来做这次换算（会差一个 DUT1）。
func TestEclipseInfoCarriesProvenance(t *testing.T) {
	date := time.Date(2024, time.April, 8, 12, 0, 0, 0, time.UTC)
	solar, ok := SolarEclipseOnDate(date)
	if !ok {
		t.Fatal("缺少 2024-04-08 日食")
	}
	if solar.GreatestJDE == 0 || solar.DeltaTSeconds == 0 {
		t.Fatalf("日食 provenance 未填充: JDE=%v ΔT=%v", solar.GreatestJDE, solar.DeltaTSeconds)
	}
	if got := solarEclipseTTJDEToTime(solar.GreatestJDE, solar.GreatestEclipse.Location()); !got.Equal(solar.GreatestEclipse) {
		t.Errorf("GreatestJDE 未经 TT2UTC 复现食甚时刻: %v vs %v", got, solar.GreatestEclipse)
	}
	// DUT1 非零时，用 ΔT 换算出来的 UT1 时刻必须与民用时刻不同：两者的差就是 DUT1。
	ut1 := basic.JD2DateByZone(solar.GreatestJDE-solar.DeltaTSeconds/86400, solar.GreatestEclipse.Location(), false)
	dut1 := (basic.TT2UTC(solar.GreatestJDE)-solar.GreatestJDE)*86400 - solar.DeltaTSeconds
	if math.Abs(dut1) > 1e-3 && ut1.Equal(solar.GreatestEclipse) {
		t.Errorf("民用时刻字段走了 ΔT（TT−UT1）换算: %v vs %v", ut1, solar.GreatestEclipse)
	}

	local := ClosestLocalSolarEclipse(date, -96.8, 32.8, 0)
	if local.GreatestJDE == 0 || local.DeltaTSeconds == 0 {
		t.Fatalf("站心日食 provenance 未填充: JDE=%v ΔT=%v", local.GreatestJDE, local.DeltaTSeconds)
	}
	if got := solarEclipseTTJDEToTime(local.GreatestJDE, local.GreatestEclipse.Location()); !got.Equal(local.GreatestEclipse) {
		t.Errorf("GreatestJDE+ΔT 未复现站心食甚: %v vs %v", got, local.GreatestEclipse)
	}

	lunar, ok := LunarEclipseOnDate(time.Date(2025, time.September, 7, 12, 0, 0, 0, time.UTC))
	if !ok {
		t.Fatal("缺少 2025-09-07 月食")
	}
	if lunar.GreatestJDE == 0 || lunar.DeltaTSeconds == 0 {
		t.Fatalf("月食 provenance 未填充: JDE=%v ΔT=%v", lunar.GreatestJDE, lunar.DeltaTSeconds)
	}
}
