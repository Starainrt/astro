package eclipse

import (
	"math"
	"testing"
	"time"
)

func solarEclipsePanelFormatRA(degrees float64) (int, int, float64) {
	total := math.Mod(degrees, 360) / 15
	hours := int(total)
	minutes := int((total - float64(hours)) * 60)
	seconds := ((total-float64(hours))*60 - float64(minutes)) * 60
	return hours, minutes, seconds
}

// 地心量面板要与 NASA 星历表同量级：视半径/视差/天平动/月序数逐项对拍 2009-07-22。
func TestSolarEclipseGeocentricPanelMatchesNASABulletin(t *testing.T) {
	panel, ok := SolarEclipseGeocentricPanelAt(time.Date(2009, time.July, 22, 0, 0, 0, 0, time.UTC))
	if !ok {
		t.Fatal("missing eclipse")
	}
	sunHours, sunMinutes, sunSeconds := solarEclipsePanelFormatRA(panel.SunRightAscensionDeg)
	t.Logf("sun  R.A. %02dh%02dm%04.1fs  Dec %+.4f  S.D. %.1f\"  H.P. %.1f\"",
		sunHours, sunMinutes, sunSeconds, panel.SunDeclinationDeg, panel.SunSemidiameterArcsec, panel.SunParallaxArcsec)
	moonHours, moonMinutes, moonSeconds := solarEclipsePanelFormatRA(panel.MoonRightAscensionDeg)
	t.Logf("moon R.A. %02dh%02dm%04.1fs  Dec %+.4f  S.D. %.1f\"  H.P. %.1f\"",
		moonHours, moonMinutes, moonSeconds, panel.MoonDeclinationDeg, panel.MoonSemidiameterArcsec, panel.MoonParallaxArcsec)
	t.Logf("libration l=%+.2f b=%+.2f c=%.2f  Brown=%d  ΔT=%.1fs  k1=%.7f k2=%.7f",
		panel.LibrationLongitudeDeg, panel.LibrationLatitudeDeg, panel.LibrationPositionAngleDeg,
		panel.BrownLunationNumber, panel.DeltaTSeconds, panel.PenumbralK, panel.UmbralK)

	// NASA 2009-07-22：太阳 00°15'44.1" / 00°00'08.7"，月亮 00°16'42.3" / 01°01'19.8"。
	if math.Abs(panel.SunSemidiameterArcsec-944.1) > 1.5 {
		t.Fatalf("sun semidiameter = %.1f\", want 944.1", panel.SunSemidiameterArcsec)
	}
	if math.Abs(panel.SunParallaxArcsec-8.7) > 0.4 {
		t.Fatalf("sun parallax = %.1f\", want 8.7", panel.SunParallaxArcsec)
	}
	if math.Abs(panel.MoonSemidiameterArcsec-1002.3) > 2 {
		t.Fatalf("moon semidiameter = %.1f\", want 1002.3", panel.MoonSemidiameterArcsec)
	}
	if math.Abs(panel.MoonParallaxArcsec-3679.8) > 4 {
		t.Fatalf("moon parallax = %.1f\", want 3679.8", panel.MoonParallaxArcsec)
	}
	if panel.BrownLunationNumber != 1071 {
		t.Fatalf("brown lunation = %d, want 1071", panel.BrownLunationNumber)
	}
	if math.Abs(panel.LibrationLongitudeDeg-0.67) > 0.05 || math.Abs(panel.LibrationPositionAngleDeg-10.52) > 0.05 {
		t.Fatalf("libration l=%.2f c=%.2f, want +0.67 / 10.52",
			panel.LibrationLongitudeDeg, panel.LibrationPositionAngleDeg)
	}
	if panel.Conjunction.IsZero() || panel.Conjunction.After(panel.Conjunction.Add(time.Hour)) {
		t.Fatal("conjunction instant missing")
	}
	t.Logf("conjunction (equal apparent ecliptic longitude) %s", panel.Conjunction.UTC().Format("15:04:05.0"))
	t.Logf("conjunction (equal apparent right ascension)   %s", panel.RightAscensionConjunction.UTC().Format("15:04:05.0"))

	// NASA 全球图的 Geocentric Conjunction 取视赤经相等，2009-07-22 为 02:33:04.4 UT。
	nasa := time.Date(2009, time.July, 22, 2, 33, 4, 0, time.UTC)
	if delta := panel.RightAscensionConjunction.Sub(nasa).Seconds(); math.Abs(delta) > 2 {
		t.Fatalf("right-ascension conjunction %s differs from NASA by %.1f s",
			panel.RightAscensionConjunction.Format("15:04:05.0"), delta)
	}
	// 该时差只约束本事件，不代表两种合时刻的普遍关系。
	if delta := panel.Conjunction.Sub(panel.RightAscensionConjunction).Seconds(); delta < 60 || delta > 120 {
		t.Fatalf("for 2009-07-22 the new moon is %.1f s after the right-ascension conjunction (event-specific, want 60..120 s)", delta)
	}
}

// 两种合时刻的差值随几何变化，符号也会改变。
func TestGeocentricConjunctionGapIsNotConstant(t *testing.T) {
	cases := []struct {
		date string
		want float64
	}{
		{"2004-04-19", -3090.6},
		{"2007-09-11", 3529.6},
		{"2009-07-22", -90.2},
		{"2025-09-21", 3368.4},
	}
	minGap, maxGap := math.Inf(1), math.Inf(-1)
	for _, tc := range cases {
		date, err := time.Parse("2006-01-02", tc.date)
		if err != nil {
			t.Fatal(err)
		}
		panel, ok := SolarEclipseGeocentricPanelAt(date.Add(12 * time.Hour))
		if !ok {
			t.Fatalf("%s 不是日食日", tc.date)
		}
		gap := panel.RightAscensionConjunction.Sub(panel.Conjunction).Seconds()
		if math.Abs(gap-tc.want) > 30 {
			t.Errorf("%s 两者相差 %.1f s, want %.1f±30 s", tc.date, gap, tc.want)
		}
		minGap, maxGap = math.Min(minGap, gap), math.Max(maxGap, gap)
	}
	if minGap > -1200 || maxGap < 1200 {
		t.Errorf("样本应同时出现提前与推后超过 20 分钟的事件，got %.1f … %.1f s", minGap, maxGap)
	}
	if maxGap-minGap < 3000 {
		t.Errorf("样本跨度 %.1f s，不足以说明该差值不是常数", maxGap-minGap)
	}
}
