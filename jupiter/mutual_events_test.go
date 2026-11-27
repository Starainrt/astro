package jupiter_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro"
	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/jupiter"
)

type mutualPublicSearch func(time.Time, int, int, jupiter.GalileanMutualPhenomenonType) jupiter.GalileanMutualEvent

func TestGalileanMutualEclipseConsistencyFieldContracts(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "mutual_events.go", nil, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	wanted := map[string]string{"ImpactParameter": "半影", "Coverage": "本影", "Annular": "本影"}
	ast.Inspect(file, func(node ast.Node) bool {
		declaration, ok := node.(*ast.TypeSpec)
		if !ok || declaration.Name.Name != "GalileanMutualEvent" {
			return true
		}
		structure := declaration.Type.(*ast.StructType)
		for _, field := range structure.Fields.List {
			for _, name := range field.Names {
				keyword, wantedField := wanted[name.Name]
				if !wantedField {
					continue
				}
				text := field.Comment.Text()
				if !strings.Contains(text, keyword) || !strings.Contains(text, " / ") {
					t.Errorf("%s missing bilingual eclipse contract: %q", name.Name, text)
				}
				delete(wanted, name.Name)
			}
		}
		return false
	})
	if len(wanted) != 0 {
		t.Fatalf("missing public fields: %v", wanted)
	}
}

func TestGalileanMutualEclipseConsistencyPaths(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*3600)
	for _, model := range []astro.Model{astro.ModelLegacy, astro.ModelSemianalytic} {
		t.Run(fmt.Sprint(model), func(t *testing.T) {
			instance, paths := mutualPublicPaths(t, model)
			event := basic.NextJupiterGalileanMutualEventWithInstance(2461466.5, 1, 3, basic.JupiterGalileanMutualEclipse, instance)
			if !event.Valid || event.Active != 1 || event.Passive != 3 {
				t.Fatalf("missing eclipse fixture: %+v", event)
			}
			maximumRoundTrip := 0.0
			for direction, query := range []float64{event.End + 1.0/1440, event.Start - 1.0/1440, event.Greatest} {
				date := basic.JD2DateByZone(basic.TT2UTC(query), zone, false)
				seed := basic.UTC2TT(basic.Date2JD(date.UTC()))
				queries := []func(float64, int, int, basic.JupiterGalileanMutualPhenomenonType, *basic.Instance) basic.JupiterGalileanMutualEvent{basic.LastJupiterGalileanMutualEventWithInstance, basic.NextJupiterGalileanMutualEventWithInstance, basic.ClosestJupiterGalileanMutualEventWithInstance}
				want := queries[direction](seed, 1, 3, basic.JupiterGalileanMutualEclipse, instance)
				var reference jupiter.GalileanMutualEvent
				for index, path := range paths {
					got := path.searches[direction](date, 1, 3, jupiter.GalileanMutualEclipse)
					maximumRoundTrip = math.Max(maximumRoundTrip, assertMutualPublicMatchesBasic(t, got, want, zone))
					if index == 0 {
						reference = got
					}
					assertMutualPublicBits(t, got, reference)
					assertMutualPublicBits(t, path.searches[direction](date, 3, 1, jupiter.GalileanMutualEclipse), reference)
					assertMutualPublicMatchesBasic(t, path.searches[direction](date.UTC(), 1, 3, jupiter.GalileanMutualEclipse), want, time.UTC)
				}
				other := astro.ModelSemianalytic
				if other == model {
					other = astro.ModelLegacy
				}
				if err := astro.SetModel(other); err != nil {
					t.Fatal(err)
				}
				for _, path := range paths[1:] {
					assertMutualPublicBits(t, path.searches[direction](date, 1, 3, jupiter.GalileanMutualEclipse), reference)
				}
				if err := astro.SetModel(model); err != nil {
					t.Fatal(err)
				}
			}
			t.Logf("eclipse model=%v three paths bitwise equal; time.Time round-trip %.9fs", model, maximumRoundTrip)
		})
	}
}

func TestGalileanMutualEclipseConsistencySeedPosition(t *testing.T) {
	instance, paths := mutualPublicPaths(t, astro.ModelSemianalytic)
	event := basic.NextJupiterGalileanMutualEventWithInstance(2461466.5, 1, 3, basic.JupiterGalileanMutualEclipse, instance)
	if !event.Valid {
		t.Fatal("missing connected eclipse fixture")
	}
	want := paths[0].searches[1](basic.JD2DateByZone(basic.TT2UTC(event.Start-1.0/1440), time.UTC, false), 1, 3, jupiter.GalileanMutualEclipse)
	for _, query := range []float64{event.Start - 1.0/1440, (event.Start + event.Greatest) / 2, event.Greatest, (event.Greatest + event.End) / 2, event.End + 1.0/1440} {
		date := basic.JD2DateByZone(basic.TT2UTC(query), time.UTC, false)
		seed := basic.UTC2TT(basic.Date2JD(date.UTC()))
		for _, path := range paths {
			last := path.searches[0](date, 1, 3, jupiter.GalileanMutualEclipse)
			next := path.searches[1](date, 1, 3, jupiter.GalileanMutualEclipse)
			assertMutualPublicBits(t, path.searches[2](date, 1, 3, jupiter.GalileanMutualEclipse), want)
			if event.Greatest <= seed {
				assertMutualPublicBits(t, last, want)
			} else if !last.Valid || !last.Greatest.Before(date) || last.Greatest.Equal(want.Greatest) {
				t.Fatalf("%s wrong Last direction", path.name)
			}
			if event.Greatest > seed {
				assertMutualPublicBits(t, next, want)
			} else if !next.Valid || !next.Greatest.After(date) || next.Start.Before(want.End) {
				t.Fatalf("%s wrong Next direction", path.name)
			}
		}
	}
}

func TestGalileanMutualEclipseConsistencyZero(t *testing.T) {
	_, paths := mutualPublicPaths(t, astro.ModelSemianalytic)
	date := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, pair := range [][2]int{{1, 1}, {0, 2}, {1, 5}, {5, 2}, {1, 0}} {
		for _, path := range paths {
			for _, query := range path.searches {
				assertMutualPublicBits(t, query(date, pair[0], pair[1], jupiter.GalileanMutualEclipse), jupiter.GalileanMutualEvent{})
			}
		}
	}
}

func TestGalileanMutualEclipseConsistencyInvisible(t *testing.T) {
	_, paths := mutualPublicPaths(t, astro.ModelSemianalytic)
	date := time.Date(2026, 12, 17, 3, 0, 0, 0, time.UTC)
	for _, path := range paths {
		event := path.searches[1](date, 2, 3, jupiter.GalileanMutualEclipse)
		if !event.Valid || event.Active != 3 || event.Passive != 2 || !event.GreatestState[event.Passive-1].Eclipse {
			t.Fatalf("%s missing invisible eclipse or state: %+v", path.name, event)
		}
		tt := basic.UTC2TT(basic.Date2JD(event.Greatest.UTC()))
		reference := basic.Date2JD(time.Date(2026, 12, 17, 4, 9, 15, 829000000, time.UTC))
		if math.Abs(tt-reference)*86400 > 0.1 {
			t.Fatalf("%s returned different invisible event: %.9fs", path.name, (tt-reference)*86400)
		}
		t.Logf("%s returns geometric 3->2 at %s Passive.Eclipse=true", path.name, event.Greatest.Format(time.RFC3339Nano))
	}
}

type mutualPublicPath struct {
	name     string
	searches [3]mutualPublicSearch
}

func mutualPublicPaths(t *testing.T, model astro.Model) (*astro.Instance, []mutualPublicPath) {
	t.Helper()
	previous := astro.GetModel()
	t.Cleanup(func() { astro.SetModel(previous) })
	if err := astro.SetModel(model); err != nil {
		t.Fatal(err)
	}
	instance, err := astro.NewModelInstance(model)
	if err != nil {
		t.Fatal(err)
	}
	context, err := astro.NewContext(model)
	if err != nil {
		t.Fatal(err)
	}
	module := jupiter.NewWithInstance(instance)
	bound := context.Jupiter()
	return instance, []mutualPublicPath{
		{"default", [3]mutualPublicSearch{jupiter.LastGalileanMutualEvent, jupiter.NextGalileanMutualEvent, jupiter.ClosestGalileanMutualEvent}},
		{"instance", [3]mutualPublicSearch{module.LastGalileanMutualEvent, module.NextGalileanMutualEvent, module.ClosestGalileanMutualEvent}},
		{"context", [3]mutualPublicSearch{bound.LastGalileanMutualEvent, bound.NextGalileanMutualEvent, bound.ClosestGalileanMutualEvent}},
	}
}

func TestGalileanMutualEventCoverage(t *testing.T) {
	_, paths := mutualPublicPaths(t, astro.ModelSemianalytic)
	query := time.Date(2021, 8, 22, 12, 0, 0, 0, time.UTC)
	event := paths[0].searches[2](query, 3, 2, jupiter.GalileanMutualOccultation)
	if !event.Valid || event.Active != 3 || event.Passive != 2 {
		t.Fatalf("reference event invalid: %+v", event)
	}
	if math.Abs(event.Coverage-0.988377027615301) > 1e-12 {
		t.Fatalf("coverage = %.16g, want %.16g", event.Coverage, 0.988377027615301)
	}
}

func TestGalileanMutualEventConsistencyPaths(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*3600)
	for _, model := range []astro.Model{astro.ModelLegacy, astro.ModelSemianalytic} {
		t.Run(fmt.Sprint(model), func(t *testing.T) {
			instance, paths := mutualPublicPaths(t, model)
			event := basic.NextJupiterGalileanMutualEventWithInstance(2461185.5, 1, 2, basic.JupiterGalileanMutualOccultation, instance)
			if !event.Valid {
				t.Fatal("reference event invalid")
			}
			queries := []float64{event.End + 1.0/1440, event.Start - 1.0/1440, event.Greatest}
			basicSearches := []func(float64, int, int, basic.JupiterGalileanMutualPhenomenonType, *basic.Instance) basic.JupiterGalileanMutualEvent{
				basic.LastJupiterGalileanMutualEventWithInstance,
				basic.NextJupiterGalileanMutualEventWithInstance,
				basic.ClosestJupiterGalileanMutualEventWithInstance,
			}
			maximumRoundTrip := 0.0
			for direction, query := range queries {
				date := basic.JD2DateByZone(basic.TT2UTC(query), zone, false)
				seed := basic.UTC2TT(basic.Date2JD(date.UTC()))
				want := basicSearches[direction](seed, 1, 2, basic.JupiterGalileanMutualOccultation, instance)
				var reference jupiter.GalileanMutualEvent
				for index, path := range paths {
					got := path.searches[direction](date, 1, 2, jupiter.GalileanMutualOccultation)
					maximumRoundTrip = math.Max(maximumRoundTrip, assertMutualPublicMatchesBasic(t, got, want, zone))
					if index == 0 {
						reference = got
					}
					assertMutualPublicBits(t, got, reference)
					assertMutualPublicBits(t, path.searches[direction](date, 2, 1, jupiter.GalileanMutualOccultation), reference)
					utcEvent := path.searches[direction](date.UTC(), 1, 2, jupiter.GalileanMutualOccultation)
					assertMutualPublicMatchesBasic(t, utcEvent, want, time.UTC)
				}
				otherModel := astro.ModelSemianalytic
				if model == otherModel {
					otherModel = astro.ModelLegacy
				}
				if err := astro.SetModel(otherModel); err != nil {
					t.Fatal(err)
				}
				for _, path := range paths[1:] {
					assertMutualPublicBits(t, path.searches[direction](date, 1, 2, jupiter.GalileanMutualOccultation), reference)
				}
				if err := astro.SetModel(model); err != nil {
					t.Fatal(err)
				}
			}
			t.Logf("three paths bitwise equal; maximum TT/time.Time round-trip %.9f s", maximumRoundTrip)
		})
	}
}

func TestGalileanMutualEventConsistencySeedPosition(t *testing.T) {
	instance, paths := mutualPublicPaths(t, astro.ModelSemianalytic)
	event := basic.NextJupiterGalileanMutualEventWithInstance(2461185.5, 1, 2, basic.JupiterGalileanMutualOccultation, instance)
	if !event.Valid {
		t.Fatal("reference event invalid")
	}
	want := paths[0].searches[1](basic.JD2DateByZone(basic.TT2UTC(event.Start-1.0/1440), time.UTC, false), 1, 2, jupiter.GalileanMutualOccultation)
	queries := []float64{event.Start - 1.0/1440, (event.Start + event.Greatest) / 2, event.Greatest, (event.Greatest + event.End) / 2, event.End + 1.0/1440}
	for index, query := range queries {
		t.Run(fmt.Sprint(index), func(t *testing.T) {
			date := basic.JD2DateByZone(basic.TT2UTC(query), time.UTC, false)
			seed := basic.UTC2TT(basic.Date2JD(date.UTC()))
			for _, path := range paths {
				last := path.searches[0](date, 1, 2, jupiter.GalileanMutualOccultation)
				next := path.searches[1](date, 1, 2, jupiter.GalileanMutualOccultation)
				closest := path.searches[2](date, 1, 2, jupiter.GalileanMutualOccultation)
				assertMutualPublicBits(t, closest, want)
				if event.Greatest <= seed {
					assertMutualPublicBits(t, last, want)
				} else if !last.Valid || !last.Greatest.Before(date) || last.Greatest.Equal(want.Greatest) {
					t.Fatalf("%s: wrong Last direction: %+v", path.name, last)
				}
				if event.Greatest > seed {
					assertMutualPublicBits(t, next, want)
				} else if !next.Valid || !next.Greatest.After(date) || next.Greatest.Equal(want.Greatest) {
					t.Fatalf("%s: wrong Next direction: %+v", path.name, next)
				}
			}
		})
	}
}

func TestGalileanMutualEventConsistencyZero(t *testing.T) {
	_, paths := mutualPublicPaths(t, astro.ModelSemianalytic)
	date := time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		first, second int
		kind          jupiter.GalileanMutualPhenomenonType
	}{
		{1, 1, jupiter.GalileanMutualOccultation},
		{0, 2, jupiter.GalileanMutualOccultation},
		{5, 2, jupiter.GalileanMutualOccultation},
		{1, 0, jupiter.GalileanMutualOccultation},
		{1, 5, jupiter.GalileanMutualOccultation},
		{1, 2, ""},
		{1, 2, "unknown"},
	}
	for _, test := range tests {
		for _, path := range paths {
			for _, search := range path.searches {
				got := search(date, test.first, test.second, test.kind)
				assertMutualPublicBits(t, got, jupiter.GalileanMutualEvent{})
			}
		}
	}
}

func TestGalileanMutualEventConsistencyInvisible(t *testing.T) {
	const greatestTT = 2461185.637388657779
	const toleranceSeconds = 0.1
	_, paths := mutualPublicPaths(t, astro.ModelSemianalytic)
	date := basic.JD2DateByZone(basic.TT2UTC(greatestTT-1.0/24), time.UTC, false)
	for _, path := range paths {
		event := path.searches[1](date, 1, 2, jupiter.GalileanMutualOccultation)
		if !event.Valid || event.Active != 1 || event.Passive != 2 {
			t.Fatalf("%s: invisible fixture missing: %+v", path.name, event)
		}
		difference := (basic.UTC2TT(basic.Date2JD(event.Greatest.UTC())) - greatestTT) * 86400
		if math.Abs(difference) > toleranceSeconds {
			t.Fatalf("%s: different event: Greatest difference %.9f s", path.name, difference)
		}
		active := event.GreatestState[event.Active-1]
		passive := event.GreatestState[event.Passive-1]
		if active.Occultation || passive.Occultation || !passive.Eclipse {
			t.Fatalf("%s: fixture must be invisible only through Passive.Eclipse: %+v", path.name, event.GreatestState)
		}
		t.Logf("%s: returned invisible %d->%d at %s; Passive.Eclipse=true; Greatest difference %.9f s", path.name, event.Active, event.Passive, event.Greatest.Format(time.RFC3339Nano), difference)
	}
}

func assertMutualPublicBits(t *testing.T, got, want jupiter.GalileanMutualEvent) {
	t.Helper()
	if got != want {
		t.Fatalf("event mismatch:\ngot  %+v\nwant %+v", got, want)
	}
	gotValues := mutualPublicFloats(got)
	wantValues := mutualPublicFloats(want)
	for index, value := range gotValues {
		if math.IsNaN(value) || math.IsInf(value, 0) || math.Float64bits(value) != math.Float64bits(wantValues[index]) {
			t.Fatalf("field %d: got %.18g want %.18g", index, value, wantValues[index])
		}
	}
}

func mutualPublicFloats(event jupiter.GalileanMutualEvent) []float64 {
	values := []float64{event.MinimumSeparationArcsec, event.ImpactParameter, event.Coverage}
	for _, state := range event.GreatestState {
		values = append(values, state.ShadowOffsetXArcsec, state.ShadowOffsetYArcsec, state.ShadowOffsetXJupiterR, state.ShadowOffsetYJupiterR)
	}
	return values
}

func assertMutualPublicMatchesBasic(t *testing.T, got jupiter.GalileanMutualEvent, want basic.JupiterGalileanMutualEvent, zone *time.Location) float64 {
	t.Helper()
	if !want.Valid || !got.Valid {
		t.Fatalf("invalid event: got %+v want %+v", got, want)
	}
	expected := jupiter.GalileanMutualEvent{
		Valid:                   true,
		Type:                    jupiter.GalileanMutualPhenomenonType(want.Type),
		Active:                  want.Active,
		Passive:                 want.Passive,
		MinimumSeparationArcsec: want.MinimumSeparationArcsec,
		ImpactParameter:         want.ImpactParameter,
		Coverage:                want.Coverage,
		Annular:                 want.Annular,
		Start:                   basic.JD2DateByZone(basic.TT2UTC(want.Start), zone, false),
		Greatest:                basic.JD2DateByZone(basic.TT2UTC(want.Greatest), zone, false),
		End:                     basic.JD2DateByZone(basic.TT2UTC(want.End), zone, false),
	}
	expected.Duration = expected.End.Sub(expected.Start)
	for index, state := range want.GreatestState {
		expected.GreatestState[index] = jupiter.GalileanSatellitePhenomenon{
			Transit:       state.Transit,
			Occultation:   state.Occultation,
			Eclipse:       state.Eclipse,
			ShadowTransit: state.ShadowTransit,
		}
		if state.ShadowTransit {
			expected.GreatestState[index].ShadowOffsetXArcsec = state.ShadowOffsetXArcsec
			expected.GreatestState[index].ShadowOffsetYArcsec = state.ShadowOffsetYArcsec
			expected.GreatestState[index].ShadowOffsetXJupiterR = state.ShadowOffsetXJupiterRadii
			expected.GreatestState[index].ShadowOffsetYJupiterR = state.ShadowOffsetYJupiterRadii
		}
	}
	assertMutualPublicBits(t, got, expected)
	if got.Start.Location() != zone || got.Greatest.Location() != zone || got.End.Location() != zone || !got.Start.Before(got.Greatest) || !got.Greatest.Before(got.End) {
		t.Fatalf("incorrect timezone or time ordering: %+v", got)
	}
	if got.Coverage < 0 || got.Coverage > 1 || got.ImpactParameter < 0 || got.ImpactParameter >= 1 || got.MinimumSeparationArcsec < 0 {
		t.Fatalf("out-of-range geometry: %+v", got)
	}
	maximum := 0.0
	for index, date := range []time.Time{got.Start, got.Greatest, got.End} {
		jde := []float64{want.Start, want.Greatest, want.End}[index]
		difference := math.Abs(basic.UTC2TT(basic.Date2JD(date.UTC()))-jde) * 86400
		if difference > time.Millisecond.Seconds() {
			t.Fatalf("TT/time.Time round-trip %.9f s exceeds 1 ms", difference)
		}
		maximum = math.Max(maximum, difference)
	}
	return maximum
}
