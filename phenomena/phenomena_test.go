// 被测契约：公开包装与 basic 一致，并保留输入时区。
package phenomena

import (
	"testing"
	"time"
)

func TestGreatConjunctionPublic(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	got := ClosestApproaches(Jupiter, Saturn, 1, time.Date(2020, 10, 1, 0, 0, 0, 0, loc), time.Date(2021, 3, 1, 0, 0, 0, 0, loc))
	if len(got) != 1 || got[0].Time.Location() != loc || got[0].Time.Month() != time.December || got[0].Value > 0.11 {
		t.Fatalf("%+v", got)
	}
}

func TestVenusHeliacalPublic(t *testing.T) {
	got := PlanetHeliacalEvents(Venus, time.Date(2025, 2, 15, 0, 0, 0, 0, time.UTC), time.Date(2025, 4, 30, 0, 0, 0, 0, time.UTC), 116.4, 39.9, HeliacalOptions{})
	if len(got) < 2 {
		t.Fatalf("%+v", got)
	}
}

func TestAltitudeCrossingPublic(t *testing.T) {
	date := time.Date(2025, 3, 20, 0, 0, 0, 0, time.FixedZone("CST", 8*3600))
	got, err := AltitudeCrossing(Sun, date, 116.4, 39.9, 0, -18, AltitudeOptions{Limb: LimbCenter, NoRefraction: true})
	if err != nil || !got.HasRise || !got.HasSet || got.State != AltitudeCrosses {
		t.Fatalf("astronomical twilight: %+v %v", got, err)
	}
	if got.Rise.Hour() != 4 || got.Set.Hour() != 19 {
		t.Errorf("unexpected twilight times %v %v", got.Rise, got.Set)
	}
	star, err := StarAltitudeCrossing(date, 101.287, -16.716, 116.4, 39.9, 0, AltitudeOptions{})
	if err != nil || !star.HasRise {
		t.Fatalf("star: %+v %v", star, err)
	}
	ev := Apsides(Earth, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC))
	if len(ev) != 2 || ev[0].Kind != Perihelion || ev[0].Time.Month() != time.January {
		t.Errorf("earth apsides 2026: %+v", ev)
	}
}
