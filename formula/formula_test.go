package formula

import (
	"math"
	"testing"
)

func assertFormulaClose(t *testing.T, name string, got, want, tolerance float64) {
	t.Helper()
	if math.Abs(got-want) > tolerance {
		t.Fatalf("%s mismatch: got %.15f want %.15f", name, got, want)
	}
}

func TestBlackbodyFormulas(t *testing.T) {
	assertFormulaClose(t, "WienPeakWavelength", WienPeakWavelength(5772), 5.020394931568954e-07, 1e-15)
	assertFormulaClose(t, "StefanBoltzmannFlux", StefanBoltzmannFlux(5772), 6.293859246828887e+07, 1e-6)
	assertFormulaClose(t, "PlanckRadianceByWavelength", PlanckRadianceByWavelength(500e-9, 5772), 2.6238540568595848e+13, 1e-1)

	if !math.IsNaN(WienPeakWavelength(0)) {
		t.Fatal("expected WienPeakWavelength(0) to be NaN")
	}
}

func TestPhotometryFormulas(t *testing.T) {
	assertFormulaClose(t, "DistanceModulus(10pc)", DistanceModulus(10), 0, 1e-15)
	assertFormulaClose(t, "DistanceModulus(100pc)", DistanceModulus(100), 5, 1e-15)
	assertFormulaClose(t, "ApparentMagnitudeFromAbsolute", ApparentMagnitudeFromAbsolute(4.83, 100), 9.83, 1e-15)
	assertFormulaClose(t, "AbsoluteMagnitudeFromApparent", AbsoluteMagnitudeFromApparent(9.83, 100), 4.83, 1e-12)
}

func TestStellarParameterFormulas(t *testing.T) {
	luminosity := LuminosityFromRadiusTemperature(2.5*solarRadiusM, 9000)
	radius := RadiusFromLuminosityTemperature(luminosity, 9000)
	temperature := EffectiveTemperatureFromLuminosityRadius(luminosity, 2.5*solarRadiusM)
	assertFormulaClose(t, "RadiusFromLuminosityTemperature", radius, 2.5*solarRadiusM, 1e-4)
	assertFormulaClose(t, "EffectiveTemperatureFromLuminosityRadius", temperature, 9000, 1e-9)

	luminositySolar := LuminositySolarFromRadiusTemperature(2.5, 9000)
	radiusSolar := RadiusSolarFromLuminosityTemperature(luminositySolar, 9000)
	temperatureSolar := EffectiveTemperatureFromLuminositySolarRadius(luminositySolar, 2.5)
	assertFormulaClose(t, "RadiusSolarFromLuminosityTemperature", radiusSolar, 2.5, 1e-12)
	assertFormulaClose(t, "EffectiveTemperatureFromLuminositySolarRadius", temperatureSolar, 9000, 1e-9)
	assertFormulaClose(t, "SolarEffectiveTemperature", SolarEffectiveTemperature(), 5772, 1e-12)
}

func TestSynodicPeriod(t *testing.T) {
	assertFormulaClose(t, "Earth-Venus synodic period", SynodicPeriod(365.25636, 224.70069), 583.9206352820089, 1e-9)
	if !math.IsInf(SynodicPeriod(365.25636, 365.25636), 1) {
		t.Fatal("expected equal periods to yield +Inf synodic period")
	}
}

func TestTelescopeFormulas(t *testing.T) {
	assertFormulaClose(t, "LightGatheringPowerRatio", LightGatheringPowerRatio(200, 100), 4, 1e-15)
	assertFormulaClose(t, "DawesLimitArcsec", DawesLimitArcsec(100), 1.16, 1e-15)
	assertFormulaClose(t, "RayleighLimitArcsec", RayleighLimitArcsec(100), 1.384, 1e-15)
	assertFormulaClose(t, "LimitingMagnitudeEmpirical", LimitingMagnitudeEmpirical(70, 6), 11, 1e-15)

	if !math.IsNaN(LightGatheringPowerRatio(0, 100)) {
		t.Fatal("expected invalid aperture to produce NaN")
	}
}
