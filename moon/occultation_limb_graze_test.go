package moon

// 外部基准：2017-09-18 轩辕十四掠掩（IOTA 实测 5 段）。内置剖面不保证段数，只钉住主遮挡段与实测最后一段相差不超过 1 s。

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/moon/limb"
)

func TestRegulusGraze20170918MatchesMainObservedInterval(t *testing.T) {
	regulus := StarCoordinate{
		ID: "HIP 49669", RA: 152.09358042, Dec: 11.96719519,
		Epoch: time.Date(1991, time.April, 2, 13, 30, 0, 0, time.UTC), Frame: CoordinateFrameICRS,
		ProperMotionRACosDecMasPerYear: -248.73, ProperMotionDecMasPerYear: 5.59,
		ParallaxMas: 41.13, RadialVelocityKmPerSecond: 6,
	}
	const lon, lat, height = 22.966666666666667, 38.98208333333333, 85
	start := time.Date(2017, time.September, 18, 3, 0, 0, 0, time.UTC)
	events, err := FindStarOccultations(start, start.Add(time.Hour), regulus, lon, lat, height,
		OccultationSearchOptions{LunarLimb: limb.Default()})
	if err != nil || len(events) != 1 {
		t.Fatalf("FindStarOccultations() = %d events, %v; want one", len(events), err)
	}
	profile, err := StarOccultationLimbProfile(events[0], regulus, limb.Default(), StarOccultationLimbProfileOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(profile.Hidden) == 0 {
		t.Fatal("graze reports no hidden interval")
	}
	main := profile.Hidden[len(profile.Hidden)-1]
	for _, pair := range []struct {
		name          string
		got, observed time.Time
	}{
		{"disappearance", main.Start, time.Date(2017, time.September, 18, 3, 29, 48, 0, time.UTC)},
		{"reappearance", main.End, time.Date(2017, time.September, 18, 3, 30, 2, 950_000_000, time.UTC)},
	} {
		if d := pair.got.Sub(pair.observed).Seconds(); math.Abs(d) > 1 {
			t.Errorf("%s %v is %.2f s from the observed %v", pair.name, pair.got, d, pair.observed)
		}
	}
}
