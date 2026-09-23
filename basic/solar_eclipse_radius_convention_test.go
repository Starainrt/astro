package basic

import (
	"math"
	"testing"
)

// 半径口径契约：k 常量固定，太阳半径口径选得进、报得出，且只按几何改结果。

func TestSolarEclipseRadiusConstants(t *testing.T) {
	if SolarEclipsePenumbralK != 0.2724880 || SolarEclipseUmbralK != 0.2722810 || SolarEclipseIAUSingleRadiusK != 0.2725076 {
		t.Fatalf("k constants = %.7f/%.7f/%.7f",
			SolarEclipsePenumbralK, SolarEclipseUmbralK, SolarEclipseIAUSingleRadiusK)
	}
	standard := SolarEclipseSunSemidiameter(2451545.0, SolarEclipseSunRadiusStandard)
	measured := SolarEclipseSunSemidiameter(2451545.0, SolarEclipseSunRadiusMeasured)
	if ratio := measured / standard; math.Abs(ratio-959.95/959.639) > 1e-8 {
		t.Fatalf("sun radius ratio = %.12f, want %.12f", ratio, 959.95/959.639)
	}
}

func TestSolarEclipseSunRadiusConventionChangesGeometry(t *testing.T) {
	seed := JDCalc(2024, 4, 8)
	base := SolarEclipseNASABulletinSplitK(seed)
	zero := SolarEclipseWithOptions(seed, SolarEclipseOptions{})
	if base.SunRadiusModel != SolarEclipseSunRadiusStandard || zero.SunRadiusModel != SolarEclipseSunRadiusStandard {
		t.Fatalf("standard provenance = %q / %q", base.SunRadiusModel, zero.SunRadiusModel)
	}
	if zero.PathWidthKM != base.PathWidthKM || zero.CentralDurationDays != base.CentralDurationDays {
		t.Fatalf("zero options changed the geometry: %.4f/%.4f", zero.PathWidthKM, base.PathWidthKM)
	}

	edge := SolarEclipseWithOptions(seed, SolarEclipseOptions{
		RadiusModel:    SolarEclipseModelNASABulletinSplitK,
		SunRadiusModel: SolarEclipseSunRadiusMeasured,
	})
	if edge.SunRadiusModel != SolarEclipseSunRadiusMeasured {
		t.Fatalf("edge provenance = %q", edge.SunRadiusModel)
	}
	if edge.Type != base.Type || math.Abs(edge.Gamma-base.Gamma) > 1e-9 {
		t.Fatalf("type/gamma changed with the solar radius: %s/%s %.9f/%.9f",
			base.Type, edge.Type, base.Gamma, edge.Gamma)
	}
	// 太阳更大 ⇒ 本影更小：全食带变窄、中心食时长变短、食分变小。
	if !(edge.Magnitude < base.Magnitude) {
		t.Fatalf("edge magnitude %.7f >= standard %.7f", edge.Magnitude, base.Magnitude)
	}
	if drop := base.PathWidthKM - edge.PathWidthKM; drop < 0.5 || drop > 2.0 {
		t.Fatalf("path width drop = %.4f km, want 0.5..2.0", drop)
	}
	if drop := (base.CentralDurationDays - edge.CentralDurationDays) * 86400; drop < 1.0 || drop > 2.5 {
		t.Fatalf("central duration drop = %.4f s, want 1.0..2.5", drop)
	}
}

func TestSolarEclipsePenumbralModelChangesPartialMagnitude(t *testing.T) {
	seed := JDCalc(2022, 10, 25)
	split := SolarEclipseNASABulletinSplitK(seed)
	single := SolarEclipseIAUSingleK(seed)
	if split.Type != SolarEclipsePartial || single.Type != SolarEclipsePartial {
		t.Fatalf("types = %s/%s, want two partial eclipses", split.Type, single.Type)
	}
	if !(single.Magnitude > split.Magnitude) || single.Magnitude-split.Magnitude > 5e-4 {
		t.Fatalf("split=%.7f single=%.7f", split.Magnitude, single.Magnitude)
	}
	edge := SolarEclipseWithOptions(seed, SolarEclipseOptions{SunRadiusModel: SolarEclipseSunRadiusMeasured})
	if edge.Model != SolarEclipseModelNASABulletinSplitK {
		t.Fatalf("zero k model resolved to %q", edge.Model)
	}
	if !(edge.Magnitude < split.Magnitude) {
		t.Fatalf("edge magnitude %.7f >= split %.7f", edge.Magnitude, split.Magnitude)
	}
}

func TestLocalSolarEclipseSunRadiusConvention(t *testing.T) {
	seed := JDCalc(2023, 10, 14)
	base := LocalSolarEclipseNASABulletinSplitK(seed, -100, 40, 0)
	edge := LocalSolarEclipseWithOptions(seed, -100, 40, 0, SolarEclipseOptions{
		RadiusModel:    SolarEclipseModelNASABulletinSplitK,
		SunRadiusModel: SolarEclipseSunRadiusMeasured,
	})
	if base.SunRadiusModel != SolarEclipseSunRadiusStandard || edge.SunRadiusModel != SolarEclipseSunRadiusMeasured {
		t.Fatalf("local provenance = %q / %q", base.SunRadiusModel, edge.SunRadiusModel)
	}
	if !(edge.Magnitude < base.Magnitude) {
		t.Fatalf("local edge magnitude %.7f >= standard %.7f", edge.Magnitude, base.Magnitude)
	}
	if shift := math.Abs(edge.PartialStart-base.PartialStart) * 86400; shift <= 0 || shift > 30 {
		t.Fatalf("local partial start shift = %.3f s, want 0..30", shift)
	}
}
func TestLocalSolarEclipseIAUSingleKUsesModelPenumbralK(t *testing.T) {
	seed := JDCalc(2024, 4, 8)
	iau := LocalSolarEclipseIAUSingleK(seed, -87.65, 41.85, 0)
	split := LocalSolarEclipseNASABulletinSplitK(seed, -87.65, 41.85, 0)
	if iau.PartialStart == split.PartialStart || iau.PartialEnd == split.PartialEnd {
		t.Fatalf("partial contacts coincide: %.9f/%.9f and %.9f/%.9f",
			iau.PartialStart, split.PartialStart, iau.PartialEnd, split.PartialEnd)
	}
	// k1 更大的模型半影也更大：偏食必然更长、更宽。
	if !(iau.PartialStart < split.PartialStart && iau.PartialEnd > split.PartialEnd) {
		t.Fatalf("IAU contacts %.9f/%.9f not wider than Split-K %.9f/%.9f",
			iau.PartialStart, iau.PartialEnd, split.PartialStart, split.PartialEnd)
	}
	if !(iau.Magnitude > split.Magnitude && iau.Obscuration > split.Obscuration) {
		t.Fatalf("IAU magnitude/obscuration %.9f/%.9f not above Split-K %.9f/%.9f",
			iau.Magnitude, iau.Obscuration, split.Magnitude, split.Obscuration)
	}
}
