package satellite

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

func benchmarkISS(b *testing.B) Elements {
	data, err := os.ReadFile("testdata/skyfield_passes.json")
	if err != nil {
		b.Fatal(err)
	}
	var fixture struct {
		Groups []struct {
			Label    string
			Elements Elements
		}
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		b.Fatal(err)
	}
	for _, g := range fixture.Groups {
		if g.Elements.CatalogNumber == 25544 {
			return g.Elements
		}
	}
	b.Fatal("missing satellite fixture")
	return Elements{}
}

func BenchmarkPropagateISS(b *testing.B) {
	e := benchmarkISS(b)
	at := e.Epoch.Add(time.Hour)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		state, err := Propagate(e, at)
		if err != nil || state.Time != at {
			b.Fatal(err)
		}
	}
}

func BenchmarkLookAtISS(b *testing.B) {
	e := benchmarkISS(b)
	at := e.Epoch.Add(time.Hour)
	obs := Observer{Latitude: 40}
	opt := ObserveOptions{}
	if _, err := LookAt(e, obs, at, opt); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		when := at.Add(time.Duration(i%86400) * time.Second)
		look, err := LookAt(e, obs, when, opt)
		if err != nil || look.Time != when {
			b.Fatal(err)
		}
	}
}

func BenchmarkPassesISS(b *testing.B) {
	e := benchmarkISS(b)
	start := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	end := start.Add(10 * 24 * time.Hour)
	obs := Observer{Latitude: 40}
	opt := PassOptions{MinElevation: 10}
	expected, _, err := Passes(e, obs, start, end, opt)
	if err != nil || len(expected) == 0 {
		b.Fatalf("count=%d err=%v", len(expected), err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		passes, _, err := Passes(e, obs, start, end, opt)
		if err != nil || len(passes) != len(expected) {
			b.Fatalf("count=%d want=%d err=%v", len(passes), len(expected), err)
		}
	}
	b.ReportMetric(float64(len(expected)), "passes/op")
}

func BenchmarkTransitsISS(b *testing.B) {
	e := benchmarkISS(b)
	start := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	end := start.Add(30 * 24 * time.Hour)
	obs, opt := Observer{Latitude: 40}, TransitOptions{}
	expected, err := Transits(e, obs, start, end, opt)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		got, err := Transits(e, obs, start, end, opt)
		if err != nil || len(got) != len(expected) {
			b.Fatalf("count=%d want=%d err=%v", len(got), len(expected), err)
		}
	}
	b.ReportMetric(float64(len(expected)), "transits/op")
}

func BenchmarkTransitEventsISS(b *testing.B) {
	e := benchmarkISS(b)
	start := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	end, opt := start.Add(24*time.Hour), TransitSearchOptions{}
	expected, err := TransitEvents(e, start, end, opt)
	if err != nil || len(expected) == 0 {
		b.Fatalf("count=%d err=%v", len(expected), err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		got, err := TransitEvents(e, start, end, opt)
		if err != nil || len(got) != len(expected) {
			b.Fatalf("count=%d want=%d err=%v", len(got), len(expected), err)
		}
	}
	b.ReportMetric(float64(len(expected)), "events/op")
}

func benchmarkLEO(b *testing.B) []Elements {
	data, err := os.ReadFile("testdata/celestrak_elements.json")
	if err != nil {
		b.Fatal(err)
	}
	var fixtures []elementFixture
	if err = json.Unmarshal(data, &fixtures); err != nil {
		b.Fatal(err)
	}
	var elements []Elements
	for _, label := range []string{"iss-tle", "css-tle", "hst-tle", "saramago-json", "iss-supgp"} {
		for _, f := range fixtures {
			if f.Label != label {
				continue
			}
			for _, e := range f.Expected {
				elements = append(elements, e)
				if len(elements) == 10 {
					b.Log("4 distinct LEO satellites plus 6 ISS segment snapshots")
					return elements
				}
			}
		}
	}
	b.Fatal("missing ten LEO element sets")
	return nil
}

func BenchmarkNearbyTransitsLEO(b *testing.B) {
	elements := benchmarkLEO(b)
	start := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	end := start.Add(7 * 24 * time.Hour)
	obs, opt := Observer{Latitude: 40}, NearbyTransitOptions{RadiusKM: 50}
	expected, err := NearbyTransits(elements, obs, start, end, opt)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		got, err := NearbyTransits(elements, obs, start, end, opt)
		if err != nil || len(got) != len(expected) {
			b.Fatalf("count=%d want=%d err=%v", len(got), len(expected), err)
		}
	}
	b.ReportMetric(float64(len(elements)), "elements/op")
	b.ReportMetric(float64(len(expected)), "nearby/op")
}

func benchmarkSolarTransitEvent(b *testing.B) (Elements, TransitEvent) {
	e := benchmarkISS(b)
	start := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	got, err := TransitEvents(e, start, start.Add(24*time.Hour), TransitSearchOptions{Targets: []TransitTarget{TransitSun}})
	if err != nil || len(got) == 0 {
		b.Fatal("missing solar transit band", err)
	}
	return e, got[0]
}

func BenchmarkTransitEventPathISS(b *testing.B) {
	e, event := benchmarkSolarTransitEvent(b)
	expected, err := TransitEventPath(e, event, time.Second)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		got, err := TransitEventPath(e, event, time.Second)
		if err != nil || len(got.CenterLine) != len(expected.CenterLine) {
			b.Fatal(err)
		}
	}
	b.ReportMetric(float64(len(expected.CenterLine)), "points/op")
}

func BenchmarkTransitShadowAtISS(b *testing.B) {
	e, event := benchmarkSolarTransitEvent(b)
	opt := TransitSearchOptions{}
	if got, err := TransitShadowAt(e, TransitSun, event.Greatest, opt); err != nil || !got.Valid {
		b.Fatal("invalid centre snapshot", err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		at := event.Greatest.Add(time.Duration(i%60000-30000) * time.Millisecond)
		got, err := TransitShadowAt(e, TransitSun, at, opt)
		if err != nil || !got.Valid || got.Time != at {
			b.Fatal("invalid snapshot", err)
		}
	}
}
