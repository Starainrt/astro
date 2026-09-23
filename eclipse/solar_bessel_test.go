package eclipse

import (
	"math"
	"testing"

	"github.com/starainrt/astro/basic"
)

// 公开包装层与被包装的 basic 结果逐字段一致，模型枚举按 eclipse 层取值。

func TestSolarEclipseBesselianElementsWrapsBasic(t *testing.T) {
	options := SolarEclipseBesselianElementsOptions{
		DeltaTSeconds: 70.6,
		ReferenceJDE:  2460409.25,
	}
	result, ok := SolarEclipseBesselianElements(2460409.262835, options)
	if !ok {
		t.Fatal("2024-04-08 should have a solar eclipse")
	}
	if result.Model != SolarEclipseModelNASABulletinSplitK {
		t.Errorf("model = %q, want %q", result.Model, SolarEclipseModelNASABulletinSplitK)
	}
	if math.Abs(result.Gamma-0.3431) > 1e-4 || math.Abs(result.Magnitude-1.0566) > 1e-3 {
		t.Errorf("gamma/magnitude = %g / %g", result.Gamma, result.Magnitude)
	}
	if math.Abs(result.TanF1-0.0046683) > 1e-6 || math.Abs(result.TanF2-0.0046450) > 1e-6 {
		t.Errorf("tan f = %g / %g", result.TanF1, result.TanF2)
	}
	published := SolarEclipseBesselianMuForPublishedTable(result.Mu, result.DeltaTSeconds)
	if math.Abs(published.At(0)-89.59122) > 1e-4 {
		t.Errorf("published mu(0) = %g, want 89.59122", published.At(0))
	}
	if math.Abs(result.X.At(0)-(-0.318157)) > 2e-4 || math.Abs(result.Y.At(0)-0.219747) > 2e-4 {
		t.Errorf("x/y at t0 = %g / %g", result.X.At(0), result.Y.At(0))
	}
	if result.PenumbralK != basic.SolarEclipsePenumbralK || result.UmbralK != basic.SolarEclipseUmbralK {
		t.Errorf("k = %g / %g", result.PenumbralK, result.UmbralK)
	}
}

func TestSolarEclipseBesselianElementsWithoutEclipse(t *testing.T) {
	if _, ok := SolarEclipseBesselianElements(2460486.5, SolarEclipseBesselianElementsOptions{}); ok {
		t.Fatal("this new moon has no solar eclipse")
	}
}

func TestSolarEclipseBesselianElementsIAUSingleK(t *testing.T) {
	result, ok := SolarEclipseBesselianElements(2460409.262835, SolarEclipseBesselianElementsOptions{
		Model:         SolarEclipseModelIAUSingleK,
		DeltaTSeconds: 70.6,
	})
	if !ok {
		t.Fatal("2024-04-08 should have a solar eclipse")
	}
	if result.Model != SolarEclipseModelIAUSingleK {
		t.Errorf("model = %q, want %q", result.Model, SolarEclipseModelIAUSingleK)
	}
	if result.UmbralK != result.PenumbralK {
		t.Errorf("k2 = %g, want k1 = %g", result.UmbralK, result.PenumbralK)
	}
}
