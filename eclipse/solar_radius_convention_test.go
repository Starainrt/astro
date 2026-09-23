package eclipse

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

// 面板与选项契约：面板回报的太阳视半径必须与它声明的口径、以及求解器实际使用的半径一致。

func TestSolarEclipsePanelSunRadiusMatchesConvention(t *testing.T) {
	date := time.Date(2024, 4, 8, 18, 0, 0, 0, time.UTC)
	info, ok := SolarEclipseOnDate(date)
	if !ok {
		t.Fatal("2024-04-08 eclipse missing")
	}
	panel, ok := SolarEclipseGeocentricPanelAt(date)
	if !ok {
		t.Fatal("panel missing")
	}
	if panel.SunRadiusModel != SolarEclipseSunRadiusStandard {
		t.Fatalf("panel sun radius model = %q", panel.SunRadiusModel)
	}
	tt := solarEclipseTimeToTTJDE(info.GreatestEclipse)
	want := basic.SolarEclipseSunSemidiameter(tt, basic.SolarEclipseSunRadiusStandard)
	if math.Abs(panel.SunSemidiameterArcsec-want) > 1e-9 {
		t.Fatalf("panel sun S.D. = %.6f, want %.6f", panel.SunSemidiameterArcsec, want)
	}
	// 通用名义半径（IAU 2015，695700 km）比食几何标准档小约 0.4″，面板不应再跟着它走。
	if nominal := basic.SunSemidiameter(tt); math.Abs(panel.SunSemidiameterArcsec-nominal) < 0.3 {
		t.Fatalf("panel sun S.D. %.6f still tracks the nominal radius %.6f", panel.SunSemidiameterArcsec, nominal)
	}
}

func TestSolarEclipseOptionsSelectSunRadius(t *testing.T) {
	date := time.Date(2024, 4, 8, 18, 0, 0, 0, time.UTC)
	standard, ok := SolarEclipseOnDateWithOptions(date, SolarEclipseOptions{})
	if !ok {
		t.Fatal("standard eclipse missing")
	}
	measured, ok := SolarEclipseOnDateWithOptions(date, SolarEclipseOptions{SunRadiusModel: SolarEclipseSunRadiusMeasured})
	if !ok {
		t.Fatal("measured eclipse missing")
	}
	if standard.SunRadiusModel != SolarEclipseSunRadiusStandard || measured.SunRadiusModel != SolarEclipseSunRadiusMeasured {
		t.Fatalf("provenance = %q / %q", standard.SunRadiusModel, measured.SunRadiusModel)
	}
	if standard.Model != SolarEclipseModelNASABulletinSplitK {
		t.Fatalf("zero options resolved the k model to %q", standard.Model)
	}
	if !(measured.PathWidthKM < standard.PathWidthKM) {
		t.Fatalf("path width %.3f >= %.3f", measured.PathWidthKM, standard.PathWidthKM)
	}
	panel, ok := SolarEclipseGeocentricPanelWithOptions(date, SolarEclipseOptions{SunRadiusModel: SolarEclipseSunRadiusMeasured})
	if !ok {
		t.Fatal("measured panel missing")
	}
	if panel.SunRadiusModel != SolarEclipseSunRadiusMeasured {
		t.Fatalf("panel sun radius model = %q", panel.SunRadiusModel)
	}
	if panel.PenumbralK != basic.SolarEclipsePenumbralK || panel.UmbralK != basic.SolarEclipseUmbralK {
		t.Fatalf("split panel k = %.7f/%.7f", panel.PenumbralK, panel.UmbralK)
	}
	if tt := solarEclipseTimeToTTJDE(measured.GreatestEclipse); math.Abs(
		panel.SunSemidiameterArcsec-basic.SolarEclipseSunSemidiameter(tt, basic.SolarEclipseSunRadiusMeasured),
	) > 1e-9 {
		t.Fatalf("measured panel sun S.D. = %.6f", panel.SunSemidiameterArcsec)
	}
}
