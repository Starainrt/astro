package moon

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/moon/limb"
)

// 月缘只应让接触时刻移动几秒：既要确实移动，又不能超过 15 s；掩甚与食型不变，模块副本不影响原模块。
func TestOccultationLunarLimbShiftsContacts(t *testing.T) {
	location := time.FixedZone("CST", 8*3600)
	start := time.Date(2024, time.August, 21, 9, 30, 0, 0, location)
	end := time.Date(2024, time.August, 21, 12, 0, 0, 0, location)
	plain := NewWithInstance(nil)
	limbed := plain.WithLunarLimb(limb.Default())
	a, err := plain.FindPlanetOccultations(start, end, OccultationSaturn, -30.072, 16.21, 0, OccultationSearchOptions{})
	if err != nil || len(a) != 1 {
		t.Fatalf("circular: %v, %d events", err, len(a))
	}
	b, err := limbed.FindPlanetOccultations(start, end, OccultationSaturn, -30.072, 16.21, 0, OccultationSearchOptions{})
	if err != nil || len(b) != 1 {
		t.Fatalf("limb: %v, %d events", err, len(b))
	}
	if !a[0].Greatest.Equal(b[0].Greatest) || a[0].Type != b[0].Type {
		t.Fatal("greatest or type changed")
	}
	moved := false
	for i, pair := range [][2]time.Time{
		{a[0].ExternalImmersion, b[0].ExternalImmersion}, {a[0].InternalImmersion, b[0].InternalImmersion},
		{a[0].InternalEmersion, b[0].InternalEmersion}, {a[0].ExternalEmersion, b[0].ExternalEmersion},
	} {
		shift := pair[1].Sub(pair[0]).Seconds()
		t.Logf("Saturn contact %d shift %+.3f s", i, shift)
		if math.Abs(shift) > 15 {
			t.Errorf("contact %d shift %.3f s too large", i, shift)
		}
		if math.Abs(shift) > 0.05 {
			moved = true
		}
	}
	if !moved {
		t.Error("lunar limb did not move any contact")
	}
	viaOptions, err := plain.FindPlanetOccultations(start, end, OccultationSaturn, -30.072, 16.21, 0,
		OccultationSearchOptions{LunarLimb: limb.Default()})
	if err != nil || len(viaOptions) != 1 || viaOptions[0] != b[0] {
		t.Error("options limb differs from module limb")
	}
	again, _ := plain.FindPlanetOccultations(start, end, OccultationSaturn, -30.072, 16.21, 0, OccultationSearchOptions{})
	if len(again) != 1 || again[0] != a[0] {
		t.Error("WithLunarLimb changed the original module")
	}
}

func TestStarOccultationLunarLimbShiftsContacts(t *testing.T) {
	observer := Observer{Longitude: 121.4737, Latitude: 31.2304, Height: 4}
	eventTime := time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)
	ra, dec := occultationTestMoonTopocentricRaDec(occultationTestTimeToTT(eventTime), observer, -1)
	star := StarCoordinate{ID: "synthetic", RA: ra + 0.1, Dec: dec + 0.1, Epoch: eventTime, Frame: CoordinateFrameApparentOfDate}
	start, end := eventTime.Add(-3*time.Hour), eventTime.Add(3*time.Hour)
	plain := NewWithInstance(nil)
	a, err := plain.FindStarOccultations(start, end, star, observer.Longitude, observer.Latitude, observer.Height, OccultationSearchOptions{MaxEvents: 1})
	if err != nil || len(a) != 1 {
		t.Fatalf("circular: %v, %d events", err, len(a))
	}
	b, err := plain.WithLunarLimb(limb.Default()).FindStarOccultations(start, end, star, observer.Longitude, observer.Latitude, observer.Height, OccultationSearchOptions{MaxEvents: 1})
	if err != nil || len(b) != 1 {
		t.Fatalf("limb: %v, %d events", err, len(b))
	}
	if !a[0].Greatest.Equal(b[0].Greatest) {
		t.Fatal("greatest changed")
	}
	for i, pair := range [][2]time.Time{{a[0].Immersion, b[0].Immersion}, {a[0].Emersion, b[0].Emersion}} {
		shift := pair[1].Sub(pair[0]).Seconds()
		t.Logf("star contact %d shift %+.3f s", i, shift)
		if math.Abs(shift) > 15 {
			t.Errorf("contact %d shift %.3f s too large", i, shift)
		}
	}
}

func TestStarOccultationPathLunarLimbShiftsGrazeLines(t *testing.T) {
	star := StarCoordinate{
		ID: "HR 4799", RA: 189.1975, Dec: -5.831944444444,
		Epoch: time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC), Frame: CoordinateFrameJ2000,
		ProperMotionRACosDecMasPerYear: -28, ProperMotionDecMasPerYear: -18,
	}
	start := time.Date(2025, 6, 4, 16, 0, 0, 0, time.UTC)
	end := time.Date(2025, 6, 5, 16, 0, 0, 0, time.UTC)
	options := OccultationPathOptions{Step: 5 * time.Minute, TargetSpacingKM: 200, DisableFootprints: true, DisableRiseSet: true}
	plain := NewWithInstance(nil)
	circular, err := plain.FindStarOccultationPaths(start, end, star, options)
	if err != nil || len(circular) != 1 {
		t.Fatalf("circular: %v %d", err, len(circular))
	}
	limbed, err := plain.WithLunarLimb(limb.Default()).FindStarOccultationPaths(start, end, star, options)
	if err != nil || len(limbed) != 1 {
		t.Fatalf("limbed: %v %d", err, len(limbed))
	}
	for side, pair := range [][2][]OccultationPathPoint{{circular[0].NorthernLimit, limbed[0].NorthernLimit}, {circular[0].SouthernLimit, limbed[0].SouthernLimit}} {
		worst, moved := 0.0, 0
		for i := range pair[0] {
			a, b := pair[0][i], pair[1][i]
			d := math.Hypot((a.Longitude-b.Longitude)*math.Cos(a.Latitude*math.Pi/180), a.Latitude-b.Latitude) * 111.2
			if d > 0 {
				moved++
			}
			worst = math.Max(worst, d)
		}
		t.Logf("side %d: %d/%d moved, max %.3f km", side, moved, len(pair[0]), worst)
		if worst > 15 || moved == 0 {
			t.Errorf("side %d: implausible graze shift (max %.3f km, moved %d)", side, worst, moved)
		}
	}
}
