package jupiter_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/starainrt/astro"
	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/jupiter"
)

func TestGalileanEventsBetweenSyntheticBenchmarkCoverage(t *testing.T) {
	for filename, names := range map[string][]string{
		"events_between_test.go":                {"BenchmarkGalileanEventsBetween2026"},
		"../basic/jupiter_mutual_bench_test.go": {"BenchmarkJupiterGalileanMutualNextSeasonal", "BenchmarkJupiterGalileanMutualNextLongInterval"},
	} {
		file, err := parser.ParseFile(token.NewFileSet(), filename, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		found := map[string]bool{}
		for _, declaration := range file.Decls {
			if function, ok := declaration.(*ast.FuncDecl); ok {
				found[function.Name.Name] = true
			}
		}
		for _, name := range names {
			if !found[name] {
				t.Fatalf("missing required benchmark %s in %s", name, filename)
			}
		}
	}
}

func BenchmarkGalileanEventsBetween2026(benchmark *testing.B) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, component := range []string{"Ordinary", "Mutual", "All"} {
		benchmark.Run(component, func(benchmark *testing.B) {
			benchmark.ReportAllocs()
			count := 0
			for iteration := 0; iteration < benchmark.N; iteration++ {
				if component == "Mutual" {
					count = benchmarkGalileanMutualBetween(start, end)
					continue
				}
				events, err := jupiter.GalileanEventsBetween(start, end, jupiter.GalileanEventsOptions{IncludeMutual: component == "All"})
				if err != nil {
					benchmark.Fatal(err)
				}
				count = len(events)
			}
			benchmark.StopTimer()
			if count == 0 {
				benchmark.Fatal("empty annual event batch")
			}
			benchmark.ReportMetric(float64(count), "events/op")
		})
	}
}

func benchmarkGalileanMutualBetween(start, end time.Time) int {
	endTT := basic.UTC2TT(basic.Date2JD(end))
	count := 0
	for first := 1; first <= 4; first++ {
		for second := first + 1; second <= 4; second++ {
			for _, phenomenon := range []basic.JupiterGalileanMutualPhenomenonType{basic.JupiterGalileanMutualOccultation, basic.JupiterGalileanMutualEclipse} {
				for seed := start; seed.Before(end); {
					event := basic.NextJupiterGalileanMutualEventBeforeWithInstance(basic.UTC2TT(basic.Date2JD(seed)), endTT, first, second, phenomenon, nil)
					if !event.Valid {
						break
					}
					count++
					seed = basic.JD2DateByZone(basic.TT2UTC(event.End), time.UTC, false).Add(100 * time.Millisecond)
				}
			}
		}
	}
	return count
}

func TestGalileanEventsBetweenSyntheticCutoff(t *testing.T) {
	start := basic.UTC2TT(basic.Date2JD(time.Date(2030, 6, 1, 0, 0, 0, 0, time.UTC)))
	for _, phenomenon := range []basic.JupiterGalileanMutualPhenomenonType{basic.JupiterGalileanMutualOccultation, basic.JupiterGalileanMutualEclipse} {
		event := basic.NextJupiterGalileanMutualEventBeforeWithInstance(start, start+1, 1, 2, phenomenon, nil)
		if event.Valid {
			t.Fatalf("event beyond cutoff: %+v", event)
		}
	}
}

func TestGalileanEventsBetweenConsistencyCutoffPrefix(t *testing.T) {
	start := basic.UTC2TT(basic.Date2JD(time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)))
	for _, phenomenon := range []basic.JupiterGalileanMutualPhenomenonType{basic.JupiterGalileanMutualOccultation, basic.JupiterGalileanMutualEclipse} {
		want := basic.NextJupiterGalileanMutualEvent(start, 1, 3, phenomenon)
		if !want.Valid {
			t.Fatal("missing fixture")
		}
		for _, seed := range []float64{start, want.Start - 1.0/86400, want.Start, (want.Start + want.Greatest) / 2} {
			unbounded := basic.NextJupiterGalileanMutualEvent(seed, 1, 3, phenomenon)
			got := basic.NextJupiterGalileanMutualEventBeforeWithInstance(seed, unbounded.Greatest+0.001/86400, 1, 3, phenomenon, nil)
			if !galileanEntryBitsEqual(reflect.ValueOf(got), reflect.ValueOf(unbounded)) {
				t.Fatalf("%s seed %.12f bounded prefix differs", phenomenon, seed)
			}
			if excluded := basic.NextJupiterGalileanMutualEventBeforeWithInstance(seed, unbounded.Greatest, 1, 3, phenomenon, nil); excluded.Valid {
				t.Fatalf("%s Greatest at cutoff accepted", phenomenon)
			}
		}
		for _, end := range []float64{start, start - 1, math.NaN(), math.Inf(1)} {
			if got := basic.NextJupiterGalileanMutualEventBeforeWithInstance(start, end, 1, 3, phenomenon, nil); got.Valid {
				t.Fatalf("invalid cutoff %g accepted", end)
			}
		}
	}
}

func TestGalileanEventsBetweenConsistencyNext(t *testing.T) {
	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.FixedZone("UTC+8", 8*3600))
	end := start.Add(14 * 24 * time.Hour)
	options := jupiter.GalileanEventsOptions{IncludeContact: true, IncludeMutual: true}
	got, err := jupiter.GalileanEventsBetween(start, end, options)
	if err != nil {
		t.Fatal(err)
	}
	want := collectGalileanEntriesByNext(start, end, options)
	if !galileanEntryBitsEqual(reflect.ValueOf(got), reflect.ValueOf(want)) {
		t.Fatalf("batch differs from individual Next calls: batch=%d individual=%d", len(got), len(want))
	}
	counts := map[jupiter.GalileanEventFamily]int{}
	for _, event := range got {
		counts[event.Family]++
		assertGalileanEntry(t, event, start.Location())
	}
	for _, family := range []jupiter.GalileanEventFamily{jupiter.GalileanFamilyPhenomenon, jupiter.GalileanFamilyContact, jupiter.GalileanFamilyMutual} {
		if counts[family] == 0 {
			t.Fatalf("missing family %s: %v", family, counts)
		}
	}
	t.Logf("all fields match individual Next calls: %d entries, families=%v", len(got), counts)
}

func TestGalileanEventsBetweenRetainsGrazingContacts(t *testing.T) {
	start := time.Date(2025, 4, 8, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)
	options := jupiter.GalileanEventsOptions{
		Satellites:     [4]bool{false, false, false, true},
		IncludeContact: true,
	}
	events, err := jupiter.GalileanEventsBetween(start, end, options)
	if err != nil {
		t.Fatal(err)
	}

	var expected []jupiter.GalileanEventEntry
	for _, phenomenon := range []jupiter.GalileanPhenomenonType{
		jupiter.GalileanPhenomenonTransit,
		jupiter.GalileanPhenomenonOccultation,
		jupiter.GalileanPhenomenonEclipse,
		jupiter.GalileanPhenomenonShadowTransit,
	} {
		for seed := start; seed.Before(end); {
			contact := jupiter.NextGalileanPhenomenonContactEvent(seed, jupiter.GalileanSatelliteCallisto, phenomenon)
			if !contact.Valid || !contact.Greatest.Before(end) {
				break
			}
			expected = append(expected, jupiter.GalileanEventEntry{Greatest: contact.Greatest, Contact: contact})
			seed = contact.Reappearance.End.Add(100 * time.Millisecond)
		}
	}

	var actual []jupiter.GalileanEventEntry
	found := [3]bool{}
	eclipseCount := 0
	for _, event := range events {
		if event.Family == jupiter.GalileanFamilyContact {
			actual = append(actual, event)
			day := event.Greatest.UTC().Format("2006-01-02")
			if event.Contact.Type == jupiter.GalileanPhenomenonEclipse {
				eclipseCount++
				found[0] = found[0] || day == "2025-04-08"
				found[2] = found[2] || day == "2025-06-14"
			}
			found[1] = found[1] || event.Contact.Type == jupiter.GalileanPhenomenonTransit && day == "2025-06-06"
		}
	}
	if len(actual) != len(expected) || found != [3]bool{true, true, true} || eclipseCount < 3 {
		t.Fatalf("contact events=%d, individual=%d, grazing fixtures=%v, eclipses=%d", len(actual), len(expected), found, eclipseCount)
	}
	for _, want := range expected {
		matched := false
		for _, got := range actual {
			if got.Contact.Type == want.Contact.Type && got.Contact.Greatest.Equal(want.Contact.Greatest) {
				matched = true
				break
			}
		}
		if !matched {
			t.Errorf("missing %s contact at %s", want.Contact.Type, want.Contact.Greatest)
		}
	}
}

func TestGalileanEventsBetweenConsistencyPaths(t *testing.T) {
	for _, model := range []astro.Model{astro.ModelLegacy, astro.ModelSemianalytic} {
		instance, _ := mutualPublicPaths(t, model)
		context, err := astro.NewContext(model)
		if err != nil {
			t.Fatal(err)
		}
		start := time.Date(2026, 6, 6, 0, 0, 0, 0, time.UTC)
		end := start.Add(3 * 24 * time.Hour)
		options := jupiter.GalileanEventsOptions{Satellites: [4]bool{true, true}, IncludeContact: true, IncludeMutual: true}
		want, err := jupiter.GalileanEventsBetween(start, end, options)
		if err != nil || len(want) == 0 {
			t.Fatalf("default path: %v", err)
		}
		for _, module := range []*jupiter.Jupiter{jupiter.NewWithInstance(instance), context.Jupiter()} {
			got, err := module.GalileanEventsBetween(start, end, options)
			if err != nil || !galileanEntryBitsEqual(reflect.ValueOf(got), reflect.ValueOf(want)) {
				t.Fatalf("model %v path mismatch: %v", model, err)
			}
		}
		mutualCount := 0
		for _, event := range want {
			if event.Family == jupiter.GalileanFamilyMutual {
				mutualCount++
			}
			if event.Family == jupiter.GalileanFamilyMutual && (event.Mutual.Active > 2 || event.Mutual.Passive > 2) {
				t.Fatalf("unselected mutual satellite: %+v", event.Mutual)
			}
		}
		if mutualCount == 0 {
			t.Fatal("mutual-event path fixture invalid")
		}
		t.Logf("model=%v three paths bit-identical, entries=%d mutual=%d", model, len(want), mutualCount)
	}
}

func collectGalileanEntriesByNext(start, end time.Time, options jupiter.GalileanEventsOptions) []jupiter.GalileanEventEntry {
	var entries []jupiter.GalileanEventEntry
	selected := options.Satellites
	if selected == [4]bool{} {
		selected = [4]bool{true, true, true, true}
	}
	for satellite := 1; satellite <= 4; satellite++ {
		if !selected[satellite-1] {
			continue
		}
		for _, phenomenon := range []jupiter.GalileanPhenomenonType{jupiter.GalileanPhenomenonTransit, jupiter.GalileanPhenomenonOccultation, jupiter.GalileanPhenomenonEclipse, jupiter.GalileanPhenomenonShadowTransit} {
			seed := start
			for seed.Before(end) {
				event := jupiter.NextGalileanPhenomenonEvent(seed, satellite, phenomenon)
				if !event.Valid || !event.Greatest.Before(end) {
					break
				}
				entries = append(entries, jupiter.GalileanEventEntry{Family: jupiter.GalileanFamilyPhenomenon, Start: event.Start, Greatest: event.Greatest, End: event.End, Phenomenon: event})
				seed = event.End.Add(100 * time.Millisecond)
			}
			if options.IncludeContact {
				seed = start
				for seed.Before(end) {
					event := jupiter.NextGalileanPhenomenonContactEvent(seed, satellite, phenomenon)
					if !event.Valid || !event.Greatest.Before(end) {
						break
					}
					entries = append(entries, jupiter.GalileanEventEntry{Family: jupiter.GalileanFamilyContact, Start: event.Disappearance.Start, Greatest: event.Greatest, End: event.Reappearance.End, Contact: event})
					seed = event.Reappearance.End.Add(100 * time.Millisecond)
				}
			}
		}
	}
	if options.IncludeMutual {
		for first := 1; first <= 4; first++ {
			for second := first + 1; second <= 4; second++ {
				if !selected[first-1] || !selected[second-1] {
					continue
				}
				for _, phenomenon := range []jupiter.GalileanMutualPhenomenonType{jupiter.GalileanMutualOccultation, jupiter.GalileanMutualEclipse} {
					for seed := start; seed.Before(end); {
						event := jupiter.NextGalileanMutualEvent(seed, first, second, phenomenon)
						if !event.Valid || !event.Greatest.Before(end) {
							break
						}
						entries = append(entries, jupiter.GalileanEventEntry{Family: jupiter.GalileanFamilyMutual, Start: event.Start, Greatest: event.Greatest, End: event.End, Mutual: event})
						seed = event.End.Add(100 * time.Millisecond)
					}
				}
			}
		}
	}
	sort.SliceStable(entries, func(first, second int) bool {
		if entries[first].Start.Equal(entries[second].Start) {
			return entries[first].Greatest.Before(entries[second].Greatest)
		}
		return entries[first].Start.Before(entries[second].Start)
	})
	return entries
}

func galileanEntryBitsEqual(first, second reflect.Value) bool {
	if first.Type() != second.Type() {
		return false
	}
	if first.Type() == reflect.TypeOf(time.Time{}) {
		return first.Interface().(time.Time) == second.Interface().(time.Time)
	}
	switch first.Kind() {
	case reflect.Float64:
		return math.Float64bits(first.Float()) == math.Float64bits(second.Float())
	case reflect.Struct:
		for index := 0; index < first.NumField(); index++ {
			if !galileanEntryBitsEqual(first.Field(index), second.Field(index)) {
				return false
			}
		}
	case reflect.Array, reflect.Slice:
		if first.Len() != second.Len() {
			return false
		}
		for index := 0; index < first.Len(); index++ {
			if !galileanEntryBitsEqual(first.Index(index), second.Index(index)) {
				return false
			}
		}
	default:
		return reflect.DeepEqual(first.Interface(), second.Interface())
	}
	return true
}

func assertGalileanEntry(t *testing.T, entry jupiter.GalileanEventEntry, zone *time.Location) {
	t.Helper()
	if entry.Start.Location() != zone || entry.Greatest.Location() != zone || entry.End.Location() != zone || entry.End.Before(entry.Start) {
		t.Fatalf("invalid entry time or zone: %+v", entry)
	}
	zeroPhenomenon := jupiter.GalileanPhenomenonEvent{}
	zeroContact := jupiter.GalileanPhenomenonContactEvent{}
	zeroMutual := jupiter.GalileanMutualEvent{}
	switch entry.Family {
	case jupiter.GalileanFamilyPhenomenon:
		if !entry.Phenomenon.Valid || entry.Contact != zeroContact || entry.Mutual != zeroMutual || entry.Start != entry.Phenomenon.Start || entry.End != entry.Phenomenon.End || entry.Greatest != entry.Phenomenon.Greatest {
			t.Fatalf("inconsistent phenomenon payload: %+v", entry)
		}
		assertGalileanEntryState(t, entry.Phenomenon.GreatestState)
	case jupiter.GalileanFamilyContact:
		if !entry.Contact.Valid || entry.Phenomenon != zeroPhenomenon || entry.Mutual != zeroMutual || entry.Start != entry.Contact.Disappearance.Start || entry.End != entry.Contact.Reappearance.End || entry.Greatest != entry.Contact.Greatest {
			t.Fatalf("inconsistent contact payload: %+v", entry)
		}
		assertGalileanEntryState(t, entry.Contact.GreatestState)
	case jupiter.GalileanFamilyMutual:
		if !entry.Mutual.Valid || entry.Phenomenon != zeroPhenomenon || entry.Contact != zeroContact || entry.Start != entry.Mutual.Start || entry.End != entry.Mutual.End || entry.Greatest != entry.Mutual.Greatest {
			t.Fatalf("inconsistent mutual payload: %+v", entry)
		}
		for _, value := range mutualPublicFloats(entry.Mutual) {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				t.Fatalf("non-finite mutual payload: %+v", entry)
			}
		}
	default:
		t.Fatalf("invalid family: %s", entry.Family)
	}
}

func assertGalileanEntryState(t *testing.T, state jupiter.GalileanSatellitePhenomenon) {
	t.Helper()
	if state.ShadowTransit {
		for _, value := range []float64{state.ShadowOffsetXArcsec, state.ShadowOffsetYArcsec, state.ShadowOffsetXJupiterR, state.ShadowOffsetYJupiterR} {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				t.Fatalf("non-finite defined shadow field: %+v", state)
			}
		}
	}
}

func TestGalileanEventsBetweenSyntheticOptions(t *testing.T) {
	start := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	for _, end := range []time.Time{start.Add(-time.Second), start.Add(366*24*time.Hour + time.Nanosecond)} {
		if _, err := jupiter.GalileanEventsBetween(start, end, jupiter.GalileanEventsOptions{}); err == nil {
			t.Fatalf("invalid interval accepted: %s..%s", start, end)
		}
	}
	empty, err := jupiter.GalileanEventsBetween(start, start, jupiter.GalileanEventsOptions{})
	if err != nil || len(empty) != 0 {
		t.Fatalf("empty interval: %v %v", empty, err)
	}
	options := jupiter.GalileanEventsOptions{Satellites: [4]bool{true}, IncludeContact: true, IncludeMutual: true}
	events, err := jupiter.GalileanEventsBetween(start, start.Add(48*time.Hour), options)
	if err != nil || len(events) == 0 {
		t.Fatalf("selected satellite: count=%d err=%v", len(events), err)
	}
	counts := map[jupiter.GalileanEventFamily]int{}
	for index, event := range events {
		counts[event.Family]++
		if index > 0 && (event.Start.Before(events[index-1].Start) || event.Start.Equal(events[index-1].Start) && event.Greatest.Before(events[index-1].Greatest)) {
			t.Fatalf("unsorted event %d", index)
		}
		switch event.Family {
		case jupiter.GalileanFamilyPhenomenon:
			if !event.Phenomenon.Valid || event.Phenomenon.Satellite != 1 || event.Contact.Valid || event.Mutual.Valid {
				t.Fatalf("phenomenon family: %+v", event)
			}
		case jupiter.GalileanFamilyContact:
			if !event.Contact.Valid || event.Contact.Satellite != 1 || event.Phenomenon.Valid || event.Mutual.Valid {
				t.Fatalf("contact family: %+v", event)
			}
		default:
			t.Fatalf("unexpected family: %s", event.Family)
		}
	}
	if counts[jupiter.GalileanFamilyPhenomenon] == 0 || counts[jupiter.GalileanFamilyContact] == 0 {
		t.Fatalf("missing family: %v", counts)
	}
	implicit, err := jupiter.GalileanEventsBetween(start, start.Add(48*time.Hour), jupiter.GalileanEventsOptions{})
	if err != nil {
		t.Fatal(err)
	}
	explicit, err := jupiter.GalileanEventsBetween(start, start.Add(48*time.Hour), jupiter.GalileanEventsOptions{Satellites: [4]bool{true, true, true, true}})
	if err != nil || !galileanEntryBitsEqual(reflect.ValueOf(implicit), reflect.ValueOf(explicit)) {
		t.Fatalf("all-false selection differs from all-true: %v", err)
	}
	for _, event := range implicit {
		if event.Family != jupiter.GalileanFamilyPhenomenon {
			t.Fatalf("default options added family %s", event.Family)
		}
	}
}
