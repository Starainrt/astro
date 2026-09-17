package eclipse

import (
	"fmt"
	"math"
	"testing"
	"time"
)

func lunarPanelTestRA(degrees float64) string {
	total := math.Mod(degrees, 360) / 15
	hours := math.Floor(total)
	minutes := math.Floor((total - hours) * 60)
	return fmt.Sprintf("%02.0fh%02.0fm%04.1fs", hours, minutes, ((total-hours)*60-minutes)*60)
}

func lunarPanelTestDec(degrees float64) string {
	sign := "+"
	if degrees < 0 {
		sign, degrees = "-", -degrees
	}
	whole := math.Floor(degrees)
	minutes := math.Floor((degrees - whole) * 60)
	seconds := ((degrees-whole)*60 - minutes) * 60
	return fmt.Sprintf("%s%02.0f\u00b0%02.0f'%04.1f\"", sign, whole, minutes, seconds)
}

// 地心量面板要与 NASA 月食图 LE2025Mar14T 逐项对得上。
func TestLunarEclipseGeocentricPanelMatchesNASABulletin(t *testing.T) {
	panel, ok := LunarEclipseGeocentricPanelAt(time.Date(2025, time.March, 14, 0, 0, 0, 0, time.UTC))
	if !ok {
		t.Fatal("missing eclipse")
	}
	t.Logf("望(黄经相差180°) %s UT", panel.FullMoon.UTC().Format("15:04:05.0"))
	t.Logf("太阳 R.A. %s  Dec. %s  S.D. %.1f\"  H.P. %.1f\"",
		lunarPanelTestRA(panel.SunRightAscensionDeg), lunarPanelTestDec(panel.SunDeclinationDeg),
		panel.SunSemidiameterArcsec, panel.SunParallaxArcsec)
	t.Logf("月亮 R.A. %s  Dec. %s  S.D. %.1f\"  H.P. %.1f\"",
		lunarPanelTestRA(panel.MoonRightAscensionDeg), lunarPanelTestDec(panel.MoonDeclinationDeg),
		panel.MoonSemidiameterArcsec, panel.MoonParallaxArcsec)
	t.Logf("Gamma %.4f  Axis %.4f°  P.R %.4f°  U.R %.4f°  食分 %.4f / %.4f  ΔT %.1f s  %s",
		panel.Gamma, panel.AxisDegrees, panel.PenumbralRadiusDegrees, panel.UmbralRadiusDegrees,
		panel.UmbralMagnitude, panel.PenumbralMagnitude, panel.DeltaTSeconds, panel.Ephemeris)

	// NASA: Ecliptic Conjunction 06:54:33.5 UT。
	conjunction := time.Date(2025, time.March, 14, 6, 54, 33, 0, time.UTC)
	if delta := panel.FullMoon.UTC().Sub(conjunction).Seconds(); math.Abs(delta) > 15 {
		t.Fatalf("full moon differs from NASA's ecliptic conjunction by %.1f s", delta)
	}
	// NASA: Gamma 0.3481、Axis 0.3171、P.R 1.1899、U.R 0.6537、食分 1.1784 / 2.2595。
	for _, check := range []struct {
		name string
		got  float64
		want float64
	}{{"Gamma", panel.Gamma, 0.3481}, {"Axis", panel.AxisDegrees, 0.3171},
		{"P.Radius", panel.PenumbralRadiusDegrees, 1.1899}, {"U.Radius", panel.UmbralRadiusDegrees, 0.6537},
		{"umbral magnitude", panel.UmbralMagnitude, 1.1784}, {"penumbral magnitude", panel.PenumbralMagnitude, 2.2595},
		{"moon S.D.", panel.MoonSemidiameterArcsec, 892.8}, {"moon H.P.", panel.MoonParallaxArcsec, 3276.8},
		{"sun S.D.", panel.SunSemidiameterArcsec, 965.2}, {"sun H.P.", panel.SunParallaxArcsec, 8.9}} {
		if math.Abs(check.got-check.want) > 1.0 {
			t.Errorf("%s = %.4f, want %.4f", check.name, check.got, check.want)
		}
	}
	// Gamma 与 Axis 必须互为两种刻度。
	parallax := panel.MoonParallaxArcsec / 3600
	if delta := panel.Gamma*parallax - panel.AxisDegrees; math.Abs(delta) > 1e-4 {
		t.Fatalf("Gamma×parallax - Axis = %g", delta)
	}
}
