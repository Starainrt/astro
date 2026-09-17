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
	// 朔（视黄经相等）比它晚约 90 s，两者不是同一个量。
	if delta := panel.Conjunction.Sub(panel.RightAscensionConjunction).Seconds(); delta < 60 || delta > 120 {
		t.Fatalf("new moon is %.1f s after the right-ascension conjunction, want about 90 s", delta)
	}
}
