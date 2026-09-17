package eclipse

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

// 地心面板、站心搜索枚举与无状态批量入口的契约：模型标记、计数语义与边界参数。

func TestSolarEclipseGeocentricPanelModelsContract(t *testing.T) {
	date := time.Date(2009, time.July, 22, 0, 0, 0, 0, time.UTC)
	singleK, ok := SolarEclipseGeocentricPanelIAUSingleK(date)
	if !ok {
		t.Fatal("IAU Single-K panel missing for 2009-07-22")
	}
	if !singleK.SingleK || singleK.Ephemeris != "IAU Single-K" {
		t.Fatalf("IAU panel SingleK=%v ephemeris=%q", singleK.SingleK, singleK.Ephemeris)
	}
	if singleK.PenumbralK != singleK.UmbralK || singleK.PenumbralK != basic.SolarEclipsePenumbralK {
		t.Fatalf("IAU panel k1=%.7f k2=%.7f, want the single k %.7f",
			singleK.PenumbralK, singleK.UmbralK, basic.SolarEclipsePenumbralK)
	}
	splitK, ok := SolarEclipseGeocentricPanelAt(date)
	if !ok {
		t.Fatal("Split-K panel missing for 2009-07-22")
	}
	if splitK.SingleK || splitK.Ephemeris != "NASA bulletin Split-K" {
		t.Fatalf("Split-K panel SingleK=%v ephemeris=%q", splitK.SingleK, splitK.Ephemeris)
	}
	if splitK.UmbralK != basic.SolarEclipseUmbralK || splitK.PenumbralK == splitK.UmbralK {
		t.Fatalf("Split-K panel k1=%.7f k2=%.7f, want two distinct radii", splitK.PenumbralK, splitK.UmbralK)
	}
	if singleK.SunSemidiameterArcsec <= 0 || splitK.MoonSemidiameterArcsec <= 0 ||
		math.IsNaN(singleK.MoonParallaxArcsec) || math.IsNaN(splitK.SunParallaxArcsec) {
		t.Fatalf("panel radii are not physical: %+v / %+v", singleK, splitK)
	}
	// 边界：当天没有日食时必须返回零值面板，而不是"取最近一次"的结果。
	eclipseFree := time.Date(2024, time.April, 15, 0, 0, 0, 0, time.UTC)
	panel, ok := SolarEclipseGeocentricPanelAt(eclipseFree)
	if ok || panel != (SolarEclipseGeocentricPanel{}) {
		t.Fatalf("panel for an eclipse-free day = %+v ok=%v, want the zero panel", panel, ok)
	}
	if _, ok := SolarEclipseGeocentricPanelIAUSingleK(eclipseFree); ok {
		t.Fatal("IAU panel accepted an eclipse-free day")
	}
}

func TestLunarEclipseGeocentricPanelShadowModelsContract(t *testing.T) {
	if LunarEclipseShadowModelDanjon == LunarEclipseShadowModelChauvenet {
		t.Fatal("the two shadow models must be distinct constants")
	}
	if LunarEclipseShadowModelDanjon != LunarEclipseShadowModel(0) {
		t.Fatalf("the zero shadow model = %d, want Danjon", int(LunarEclipseShadowModelDanjon))
	}
	date := time.Date(2025, time.March, 14, 0, 0, 0, 0, time.UTC)
	danjon, ok := LunarEclipseGeocentricPanelAt(date)
	if !ok {
		t.Fatal("Danjon panel missing for 2025-03-14")
	}
	chauvenet, ok := LunarEclipseGeocentricPanelChauvenet(date)
	if !ok {
		t.Fatal("Chauvenet panel missing for 2025-03-14")
	}
	if danjon.ShadowModel != LunarEclipseShadowModelDanjon || danjon.Ephemeris != "Danjon" {
		t.Fatalf("default panel model=%d ephemeris=%q, want Danjon", int(danjon.ShadowModel), danjon.Ephemeris)
	}
	if chauvenet.ShadowModel != LunarEclipseShadowModelChauvenet || chauvenet.Ephemeris != "Chauvenet" {
		t.Fatalf("Chauvenet panel model=%d ephemeris=%q", int(chauvenet.ShadowModel), chauvenet.Ephemeris)
	}
	if chauvenet.UmbralRadiusDegrees <= danjon.UmbralRadiusDegrees {
		t.Fatalf("Chauvenet umbral radius %.6f must exceed the Danjon one %.6f",
			chauvenet.UmbralRadiusDegrees, danjon.UmbralRadiusDegrees)
	}
	if math.Abs(danjon.Gamma-chauvenet.Gamma) > 1e-9 ||
		math.Abs(danjon.MoonDistanceEarthRadii-chauvenet.MoonDistanceEarthRadii) > 1e-9 {
		t.Fatalf("shadow models changed the geometry: gamma %.9f/%.9f distance %.9f/%.9f",
			danjon.Gamma, chauvenet.Gamma, danjon.MoonDistanceEarthRadii, chauvenet.MoonDistanceEarthRadii)
	}
	if danjon.PenumbralMagnitude <= danjon.UmbralMagnitude {
		t.Fatalf("penumbral magnitude %.5f must exceed the umbral one %.5f",
			danjon.PenumbralMagnitude, danjon.UmbralMagnitude)
	}
	// 边界：最近一次望没有月食时返回零值面板。
	eclipseFree := time.Date(2025, time.April, 13, 0, 0, 0, 0, time.UTC)
	panel, ok := LunarEclipseGeocentricPanelAt(eclipseFree)
	if ok || panel != (LunarEclipseGeocentricPanel{}) {
		t.Fatalf("panel for an eclipse-free full moon = %+v ok=%v, want the zero panel", panel, ok)
	}
	if _, ok := LunarEclipseGeocentricPanelChauvenet(eclipseFree); ok {
		t.Fatal("Chauvenet panel accepted an eclipse-free full moon")
	}
}

func TestSolarEclipseLocalSearchKindAndStatusContract(t *testing.T) {
	if SolarEclipseLocalSearchAny != SolarEclipseLocalSearchKind(0) {
		t.Fatalf("zero kind = %d, want Any", int(SolarEclipseLocalSearchAny))
	}
	if SolarEclipseLocalSearchAny == SolarEclipseLocalSearchTotal ||
		SolarEclipseLocalSearchAny == SolarEclipseLocalSearchAnnular ||
		SolarEclipseLocalSearchTotal == SolarEclipseLocalSearchAnnular {
		t.Fatal("the search kinds must be distinct constants")
	}
	start := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	info, status := SearchLocalCentralSolarEclipse(start, 113.26, 23.13, 0,
		SolarEclipseLocalSearchOptions{MaxYears: 40})
	if !status.Found || status.Exhausted || status.Capped {
		t.Fatalf("Any 40y: found=%v exhausted=%v capped=%v", status.Found, status.Exhausted, status.Capped)
	}
	if info.GreatestEclipse.IsZero() || status.MonthsScanned <= 0 ||
		status.GlobalChecks <= 0 || status.LocalChecks <= 0 || status.YearsScanned <= 0 {
		t.Fatalf("Any 40y status did not account for the work: %+v", status)
	}
	// 边界：跨度小于最近一次朔时用尽跨度。
	_, tiny := SearchLocalCentralSolarEclipse(start, 116.40, 39.90, 0,
		SolarEclipseLocalSearchOptions{Kind: SolarEclipseLocalSearchTotal, MaxYears: 0.01})
	if tiny.Found || !tiny.Exhausted || tiny.Capped {
		t.Fatalf("tiny horizon: found=%v exhausted=%v capped=%v", tiny.Found, tiny.Exhausted, tiny.Capped)
	}
	if tiny.YearsScanned <= 0 || tiny.MonthsScanned > 1 {
		t.Fatalf("tiny horizon scanned %d months over %.4f years", tiny.MonthsScanned, tiny.YearsScanned)
	}
	// Annular 只接受环食阶段：2020-06-21 的环食带经过厦门。
	backward := time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC)
	annular, annulus := SearchLocalCentralSolarEclipse(backward, 118.09, 24.48, 0,
		SolarEclipseLocalSearchOptions{Kind: SolarEclipseLocalSearchAnnular, Backward: true, MaxYears: 2})
	if !annulus.Found || !annular.HasAnnular || annular.HasTotal {
		t.Fatalf("Xiamen annular: found=%v hasAnnular=%v hasTotal=%v",
			annulus.Found, annular.HasAnnular, annular.HasTotal)
	}
	if got := annular.GreatestEclipse.UTC().Format("2006-01-02"); got != "2020-06-21" {
		t.Fatalf("Xiamen annular greatest = %s, want 2020-06-21", got)
	}
	total, totalStatus := SearchLocalCentralSolarEclipse(backward, 118.09, 24.48, 0,
		SolarEclipseLocalSearchOptions{Kind: SolarEclipseLocalSearchTotal, Backward: true, MaxYears: 2})
	if totalStatus.Found && total.GreatestEclipse.UTC().Format("2006-01-02") == "2020-06-21" {
		t.Fatal("the Total kind accepted the 2020-06-21 annular-only event")
	}
}

func TestSolarEclipseShadowBetweenStatelessEntryPointsContract(t *testing.T) {
	start := time.Date(2009, time.July, 22, 1, 0, 0, 0, time.UTC)
	step := 30 * time.Minute
	instants := SolarEclipseShadowBetween(start, start.Add(time.Hour), step)
	if len(instants) != 3 {
		t.Fatalf("stateless batch = %d entries, want 3", len(instants))
	}
	solver := NewSolarEclipseShadowSolver(SolarEclipseShadowSolverOptions{})
	for index, instant := range instants {
		want := start.Add(time.Duration(index) * step)
		if !instant.Time.Equal(want) {
			t.Fatalf("entry %d time = %v, want %v", index, instant.Time, want)
		}
		single, _ := solver.ShadowAt(want)
		if instant.Empty() != single.Empty() || instant.Topology.Signature() != single.Topology.Signature() {
			t.Fatalf("entry %d disagrees with the solver: %q vs %q",
				index, instant.Topology.Signature(), single.Topology.Signature())
		}
	}
	override := 1234.5
	withOptions := SolarEclipseShadowBetweenWithOptions(start, start.Add(time.Hour), step,
		SolarEclipseShadowSolverOptions{DeltaTSeconds: override})
	if len(withOptions) != len(instants) {
		t.Fatalf("option batch = %d entries, want %d", len(withOptions), len(instants))
	}
	for index, instant := range withOptions {
		if instant.DeltaTSeconds != override {
			t.Fatalf("entry %d ΔT = %.4f, want %.4f", index, instant.DeltaTSeconds, override)
		}
	}
	// 边界：非正步长、倒置窗口与零长窗口都返回 nil。
	for _, window := range []struct {
		name       string
		begin, end time.Time
		step       time.Duration
	}{
		{name: "零步长", begin: start, end: start.Add(time.Hour), step: 0},
		{name: "负步长", begin: start, end: start.Add(time.Hour), step: -step},
		{name: "倒置窗口", begin: start.Add(time.Hour), end: start, step: step},
		{name: "零长窗口", begin: start, end: start, step: step},
	} {
		if got := SolarEclipseShadowBetween(window.begin, window.end, window.step); got != nil {
			t.Fatalf("%s returned %d entries, want nil", window.name, len(got))
		}
		if got := SolarEclipseShadowBetweenWithOptions(window.begin, window.end, window.step,
			SolarEclipseShadowSolverOptions{}); got != nil {
			t.Fatalf("%s option entry returned %d entries, want nil", window.name, len(got))
		}
	}
}

func TestSolarEclipseStationStateAtUsesGivenInstantContract(t *testing.T) {
	value := time.Date(2009, time.July, 22, 2, 0, 0, 0, time.UTC)
	daySide := SolarEclipseStationStateAt(value, 150, 20, 0)
	nightSide := SolarEclipseStationStateAt(value, -30, 20, 0)
	if !daySide.Visible || nightSide.Visible {
		t.Fatalf("visibility at %s: 150E=%v 30W=%v", value.Format(time.RFC3339), daySide.Visible, nightSide.Visible)
	}
	if !daySide.Time.Equal(value) || daySide.JDE <= 0 || daySide.DeltaTSeconds <= 0 {
		t.Fatalf("state did not carry the requested instant: %+v", daySide)
	}
	if daySide.SunRadiusDeg <= 0 || daySide.MoonOuterRadiusDeg <= 0 || daySide.SeparationArcsec <= 0 {
		t.Fatalf("state radii are not physical: %+v", daySide)
	}
	if daySide.SunAltitudeDeg <= 0 || nightSide.SunAltitudeDeg >= 0 {
		t.Fatalf("altitudes disagree with the visibility flags: %.3f / %.3f",
			daySide.SunAltitudeDeg, nightSide.SunAltitudeDeg)
	}
	if daySide.Magnitude <= 0 || daySide.Magnitude >= 1 {
		t.Fatalf("day-side partial magnitude = %.6f, want a partial value", daySide.Magnitude)
	}
	if nightSide.Magnitude != 0 || nightSide.Obscuration != 0 {
		t.Fatalf("below-horizon state reports magnitude %.6f obscuration %.6f", nightSide.Magnitude, nightSide.Obscuration)
	}
}
