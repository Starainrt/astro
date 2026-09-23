package eclipse

import (
	"testing"
	"time"
)

// 站心偏食口径：IAU Single-K 的半影外半径必须用本模型的单一 k，与 Split-K 不可互换。

func TestLocalSolarEclipseOnDateCarriesModelPenumbralK(t *testing.T) {
	date := time.Date(2024, 4, 8, 0, 0, 0, 0, time.UTC)
	iau, okIAU := LocalSolarEclipseOnDateIAUSingleK(date, -87.65, 41.85, 0)
	split, okSplit := LocalSolarEclipseOnDateNASABulletinSplitK(date, -87.65, 41.85, 0)
	if !okIAU || !okSplit {
		t.Fatalf("local eclipse missing: %v/%v", okIAU, okSplit)
	}
	if iau.PartialStart.Equal(split.PartialStart) || iau.PartialEnd.Equal(split.PartialEnd) {
		t.Fatalf("partial contacts coincide: %s/%s and %s/%s",
			iau.PartialStart, split.PartialStart, iau.PartialEnd, split.PartialEnd)
	}
	if !(iau.PartialStart.Before(split.PartialStart) && iau.PartialEnd.After(split.PartialEnd)) {
		t.Fatalf("IAU contacts %s/%s not wider than Split-K %s/%s",
			iau.PartialStart, iau.PartialEnd, split.PartialStart, split.PartialEnd)
	}
	if !(iau.Magnitude > split.Magnitude && iau.Obscuration > split.Obscuration) {
		t.Fatalf("IAU magnitude/obscuration %.9f/%.9f not above Split-K %.9f/%.9f",
			iau.Magnitude, iau.Obscuration, split.Magnitude, split.Obscuration)
	}
	if iau.Model != SolarEclipseModelIAUSingleK || split.Model != SolarEclipseModelNASABulletinSplitK {
		t.Fatalf("radius provenance = %v/%v", iau.Model, split.Model)
	}
}
