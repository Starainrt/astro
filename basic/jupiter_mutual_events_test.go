package basic

import (
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"os"
	"sort"
	"strings"
	"testing"
	"time"
)

const (
	mutualPolarSpeedTolerance          = 1.05
	mutualPolarDeclinationLimitDegrees = 3.9
)

func TestJupiterGalileanMutualSyntheticEndpointReuse(t *testing.T) {
	jde := JDCalc(2026, 8, 1)
	memo := jupiterGalileanMutualSplitMemo{}
	first, firstOK := memo.at(jde, 1, 3, nil)
	second, secondOK := memo.at(jde, 1, 3, nil)
	if !firstOK || !secondOK || first != second || memo.next != 1 {
		t.Fatal("repeated occultation endpoint was not reused")
	}
	neighbor := math.Nextafter(jde, math.Inf(1))
	if _, ok := memo.at(neighbor, 1, 3, nil); !ok || memo.next != 2 {
		t.Fatal("distinct endpoint bit patterns aliased")
	}
	search := jupiterGalileanEclipseSearch{first: 1, second: 3, memo: &jupiterGalileanEclipseMemo{}}
	firstEclipse, firstEclipseOK := search.geometry(jde)
	secondEclipse, secondEclipseOK := search.geometry(jde)
	if firstEclipseOK != secondEclipseOK || firstEclipse != secondEclipse || search.memo.next != 1 {
		t.Fatal("repeated eclipse endpoint was not reused")
	}
	if _, ok := search.geometry(neighbor); ok != firstEclipseOK || search.memo.next != 2 {
		t.Fatal("distinct eclipse endpoint bit patterns aliased")
	}
	copySearch := search
	copySearch.geometry(jde)
	if search.memo.next != 2 {
		t.Fatal("search copy lost call-local endpoint reuse")
	}
	for sample := 2; sample < 8; sample++ {
		candidate := jde + float64(sample)/24
		got, gotOK := memo.at(candidate, 1, 3, nil)
		want, wantOK := jupiterGalileanMutualGeometryAt(candidate, 1, 3, nil)
		if gotOK != wantOK || got != want {
			t.Fatalf("occultation replacement sample=%d differs", sample)
		}
		gotEclipse, gotEclipseOK := search.geometry(candidate)
		wantEclipse, wantEclipseOK := jupiterGalileanMutualEclipseGeometryAt(candidate, 1, 3, nil)
		if gotEclipseOK != wantEclipseOK || gotEclipse != wantEclipse {
			t.Fatalf("eclipse replacement sample=%d differs", sample)
		}
	}
	uncached := jupiterGalileanEclipseSearch{first: 1, second: 3}
	uncached.geometry(jde)
	if uncached.memo != nil {
		t.Fatal("standalone evaluation created a persistent cache")
	}
}

func TestJupiterGalileanMutualConsistencyPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("multi-year chain and cross-season timings")
	}
	for _, phenomenon := range []JupiterGalileanMutualPhenomenonType{JupiterGalileanMutualOccultation, JupiterGalileanMutualEclipse} {
		for _, yearEnd := range []int{2028, 2027} {
			totalEvents := 0
			var totalElapsed, totalInWindow time.Duration
			for _, pair := range [][2]int{{1, 2}, {1, 3}, {1, 4}, {2, 3}, {2, 4}, {3, 4}} {
				seed := JDCalc(2026, 1, 1)
				count, chainCount := 0, 0
				var chainTotal, chainMaximum, firstElapsed, terminalElapsed time.Duration
				for {
					started := time.Now()
					event := NextJupiterGalileanMutualEvent(seed, pair[0], pair[1], phenomenon)
					elapsed := time.Since(started)
					totalElapsed += elapsed
					if !event.Valid || event.Start > JDCalc(yearEnd, 1, 1) {
						terminalElapsed = elapsed
						break
					}
					totalInWindow += elapsed
					if count == 0 {
						firstElapsed = elapsed
					} else {
						chainCount++
						chainTotal += elapsed
						if elapsed > chainMaximum {
							chainMaximum = elapsed
						}
					}
					count++
					seed = event.End + 0.1/86400
				}
				average := time.Duration(0)
				if chainCount > 0 {
					average = chainTotal / time.Duration(chainCount)
				}
				totalEvents += count
				t.Logf("timing type=%s years=2026-%d pair=%v events=%d first=%s chain_calls=%d chain_total=%s chain_average=%s chain_max=%s terminal=%s", phenomenon, yearEnd-1, pair, count, firstElapsed, chainCount, chainTotal, average, chainMaximum, terminalElapsed)
			}
			t.Logf("timing total type=%s years=2026-%d events=%d in_window=%s including_six_terminal_calls=%s", phenomenon, yearEnd-1, totalEvents, totalInWindow, totalElapsed)
		}
	}
	for _, test := range []struct {
		year, month   int
		day           float64
		first, second int
		phenomenon    JupiterGalileanMutualPhenomenonType
	}{
		{2026, 12, 6, 1, 2, JupiterGalileanMutualEclipse},
		{2026, 10, 3, 1, 2, JupiterGalileanMutualOccultation},
		{2027, 8, 10, 2, 3, JupiterGalileanMutualEclipse},
		{2027, 8, 10, 2, 3, JupiterGalileanMutualOccultation},
		{2030, 6, 1, 1, 3, JupiterGalileanMutualEclipse},
	} {
		seed := JDCalc(test.year, test.month, test.day)
		started := time.Now()
		event := NextJupiterGalileanMutualEvent(seed, test.first, test.second, test.phenomenon)
		t.Logf("timing cross-season type=%s pair=%d-%d seed=%s greatest=%s elapsed=%s", test.phenomenon, test.first, test.second, JD2Date(seed).Format("2006-01-02T15:04:05"), JD2Date(event.Greatest).Format("2006-01-02T15:04:05"), time.Since(started))
		if event.Valid && event.End < seed+1 {
			tailSeed := event.End + jupiterGalileanMutualEventEpsilonDays
			started = time.Now()
			next := NextJupiterGalileanMutualEvent(tailSeed, test.first, test.second, test.phenomenon)
			t.Logf("timing cross-season-tail type=%s pair=%d-%d seed=%s greatest=%s elapsed=%s", test.phenomenon, test.first, test.second, JD2Date(tailSeed).Format("2006-01-02T15:04:05.000"), JD2Date(next.Greatest).Format("2006-01-02T15:04:05"), time.Since(started))
		}
	}
}

func TestJupiterGalileanMutualSyntheticPolarGuard(t *testing.T) {
	for _, test := range []struct {
		degrees float64
		allowed bool
	}{{0, true}, {3.89, true}, {-3.89, true}, {3.91, false}, {-3.91, false}, {90, false}, {math.NaN(), false}} {
		if got := jupiterGalileanMutualPolarAllows(math.Sin(test.degrees * rad)); got != test.allowed {
			t.Fatalf("polar guard at %.6f degrees = %v", test.degrees, got)
		}
	}
	for _, allowed := range []bool{false, true} {
		got := jupiterGalileanMutualOuterStep(1, 4, 20, 10, allowed)
		want := 0.05
		if allowed {
			want = 0.4
		}
		if math.Float64bits(got) != math.Float64bits(want) {
			t.Fatalf("guarded step %.12f want %.12f", got, want)
		}
		for _, otherAllowed := range []bool{false, true} {
			got := jupiterGalileanMutualPolarExcludes(2, 2, 10, 0.1, allowed, otherAllowed)
			if got != (allowed && otherAllowed) {
				t.Fatalf("endpoint guards %v/%v exclude=%v", allowed, otherAllowed, got)
			}
		}
	}
}

func TestJupiterGalileanMutualConsistencyPolarSpeedBound(t *testing.T) {
	start, end := JDCalc(2020, 1, 1), JDCalc(2034, 1, 1)
	if testing.Short() {
		start, end = JDCalc(2026, 1, 1), JDCalc(2027, 1, 1)
	}
	for _, pair := range [][2]int{{1, 2}, {1, 3}, {1, 4}, {2, 3}, {2, 4}, {3, 4}} {
		maximumOccultation, maximumEclipse, maximumDE, maximumDS := 0.0, 0.0, 0.0, 0.0
		bound := jupiterGalileanMutualPolarSpeedUpperBound(pair[0], pair[1]) / jupiterGalileanMutualSpeedSafetyFactor
		for sample := 0; ; sample++ {
			jde := start + float64(sample)/24
			if jde >= end {
				break
			}
			before, beforeOK := jupiterGalileanMutualGeometryAt(jde-30.0/86400, pair[0], pair[1])
			after, afterOK := jupiterGalileanMutualGeometryAt(jde+30.0/86400, pair[0], pair[1])
			if !beforeOK || !afterOK {
				t.Fatalf("invalid occultation geometry pair=%v jde=%.9f", pair, jde)
			}
			for _, geometry := range []jupiterGalileanMutualGeometry{before, after} {
				if !isFinite(geometry.Polar) || geometry.Polar < 0 || geometry.Polar > geometry.Separation+1e-12 {
					t.Fatalf("invalid polar component pair=%v", pair)
				}
			}
			ratio := math.Abs(after.Polar-before.Polar) * 1440 / bound
			maximumOccultation = math.Max(maximumOccultation, ratio)
			if ratio > mutualPolarSpeedTolerance {
				t.Fatalf("pair=%v jde=%.9f occultation polar speed ratio %.12f > 1.05", pair, jde, ratio)
			}
			eclipseBefore, eclipseBeforeOK := jupiterGalileanMutualEclipseGeometryAt(jde-30.0/86400, pair[0], pair[1], nil)
			eclipseAfter, eclipseAfterOK := jupiterGalileanMutualEclipseGeometryAt(jde+30.0/86400, pair[0], pair[1], nil)
			if eclipseBeforeOK && eclipseAfterOK && eclipseBefore.Active == eclipseAfter.Active {
				for _, geometry := range []jupiterGalileanMutualEclipseGeometry{eclipseBefore, eclipseAfter} {
					if !isFinite(geometry.PolarKM) || geometry.PolarKM < 0 || geometry.PolarKM > geometry.SeparationKM+1e-8 {
						t.Fatalf("invalid eclipse polar component pair=%v", pair)
					}
				}
				ratio = math.Abs(eclipseAfter.PolarKM-eclipseBefore.PolarKM) * 1440 / (bound * jupiterGalileanEquatorialRadiusKM)
				maximumEclipse = math.Max(maximumEclipse, ratio)
				if ratio > mutualPolarSpeedTolerance {
					t.Fatalf("pair=%v jde=%.9f eclipse polar speed ratio %.12f > 1.05", pair, jde, ratio)
				}
			}
			declinationEarth := math.Abs(jupiterGalileanMutualEarthDeclination(jde))
			frame := newJupiterGalileanEclipseFrame(jde, nil)
			declinationSun := math.Abs(math.Asin(-vectorDot(normalizeVector(frame.jupiterHelio), frame.pole)) * deg)
			maximumDE, maximumDS = math.Max(maximumDE, declinationEarth), math.Max(maximumDS, declinationSun)
			if !isFinite(declinationEarth) || !isFinite(declinationSun) || declinationEarth >= mutualPolarDeclinationLimitDegrees || declinationSun >= mutualPolarDeclinationLimitDegrees {
				t.Fatalf("polar guard limit pair=%v jde=%.9f DE/DS=%.9f/%.9f", pair, jde, declinationEarth, declinationSun)
			}
		}
		t.Logf("pair=%v maximum polar speed ratios occultation=%.9f eclipse=%.9f max |DE|=%.9f |DS|=%.9f degrees", pair, maximumOccultation, maximumEclipse, maximumDE, maximumDS)
	}
}

const (
	mutualBaselineShadowToleranceKM         = 0.1
	mutualBaselineContactToleranceSeconds   = 0.001
	mutualBaselineGreatestToleranceSeconds  = 0.05
	mutualBaselineSeparationToleranceArcsec = 1e-9
	mutualBaselineFieldTolerance            = 1e-7
)

func mutualPhenomenaExactlyEqual(first, second [4]JupiterGalileanPhenomenon) bool {
	for index := range first {
		firstState, secondState := first[index], second[index]
		for _, coordinates := range [][2]float64{
			{firstState.ShadowOffsetXArcsec, secondState.ShadowOffsetXArcsec},
			{firstState.ShadowOffsetYArcsec, secondState.ShadowOffsetYArcsec},
			{firstState.ShadowOffsetXJupiterRadii, secondState.ShadowOffsetXJupiterRadii},
			{firstState.ShadowOffsetYJupiterRadii, secondState.ShadowOffsetYJupiterRadii},
		} {
			if math.Float64bits(coordinates[0]) != math.Float64bits(coordinates[1]) {
				return false
			}
		}
		if !mutualPhenomenonFlagsEqual(firstState, secondState) {
			return false
		}
	}
	return true
}

func mutualPhenomenonFlagsEqual(first, second JupiterGalileanPhenomenon) bool {
	return first.Transit == second.Transit && first.Occultation == second.Occultation && first.Eclipse == second.Eclipse && first.ShadowTransit == second.ShadowTransit
}

func mutualEclipseBaselineDifference(baseline, current JupiterGalileanMutualEvent) (float64, error) {
	if baseline.Valid != current.Valid || baseline.Type != current.Type || baseline.Active != current.Active || baseline.Passive != current.Passive || baseline.Annular != current.Annular {
		return 0, fmt.Errorf("event identity or flags changed")
	}
	for _, event := range []JupiterGalileanMutualEvent{baseline, current} {
		state := JupiterGalileanSatellitePhenomenaWithInstance(event.Greatest, nil)
		if !mutualPhenomenaExactlyEqual(state, event.GreatestState) {
			return 0, fmt.Errorf("recorded state differs at its own Greatest %.12f", event.Greatest)
		}
	}
	maximumKM := 0.0
	context := newJupiterGalileanObservationContext(baseline.Greatest)
	for index, first := range baseline.GreatestState {
		second := current.GreatestState[index]
		if !mutualPhenomenonFlagsEqual(first, second) {
			return 0, fmt.Errorf("satellite %d phenomenon flags changed", index+1)
		}
		for _, coordinates := range [][5]float64{
			{first.ShadowOffsetXArcsec, first.ShadowOffsetYArcsec, second.ShadowOffsetXArcsec, second.ShadowOffsetYArcsec, context.jupiterDistance * astronomicalUnitKM * rad / 3600},
			{first.ShadowOffsetXJupiterRadii, first.ShadowOffsetYJupiterRadii, second.ShadowOffsetXJupiterRadii, second.ShadowOffsetYJupiterRadii, jupiterGalileanEquatorialRadiusKM},
		} {
			if math.IsNaN(coordinates[0]) && math.IsNaN(coordinates[1]) && math.IsNaN(coordinates[2]) && math.IsNaN(coordinates[3]) {
				continue
			}
			differenceKM := math.Hypot(coordinates[2]-coordinates[0], coordinates[3]-coordinates[1]) * coordinates[4]
			if !isFinite(differenceKM) || differenceKM > mutualBaselineShadowToleranceKM {
				return differenceKM, fmt.Errorf("satellite %d shadow difference %.12g km", index+1, differenceKM)
			}
			maximumKM = math.Max(maximumKM, differenceKM)
		}
	}
	for _, field := range []struct {
		name                     string
		first, second, tolerance float64
	}{
		{"Start", baseline.Start, current.Start, mutualBaselineContactToleranceSeconds / 86400},
		{"End", baseline.End, current.End, mutualBaselineContactToleranceSeconds / 86400},
		{"Duration", baseline.DurationDays, current.DurationDays, mutualBaselineContactToleranceSeconds / 86400},
		{"Greatest", baseline.Greatest, current.Greatest, mutualBaselineGreatestToleranceSeconds / 86400},
		{"separation", baseline.MinimumSeparationArcsec, current.MinimumSeparationArcsec, mutualBaselineSeparationToleranceArcsec},
		{"impact", baseline.ImpactParameter, current.ImpactParameter, mutualBaselineFieldTolerance},
		{"coverage", baseline.Coverage, current.Coverage, mutualBaselineFieldTolerance},
	} {
		if !isFinite(field.first) || !isFinite(field.second) || math.Abs(field.second-field.first) > field.tolerance {
			return maximumKM, fmt.Errorf("%s difference %.12g exceeds %.12g", field.name, math.Abs(field.second-field.first), field.tolerance)
		}
	}
	return maximumKM, nil
}

func TestJupiterGalileanMutualEclipseSyntheticBaselineState(t *testing.T) {
	event := NextJupiterGalileanMutualEvent(JDCalc(2026, 12, 31), 2, 3, JupiterGalileanMutualEclipse)
	if !event.Valid {
		t.Fatal("missing fixture")
	}
	if _, err := mutualEclipseBaselineDifference(event, event); err != nil {
		t.Fatal(err)
	}
	shifted := event
	shifted.Greatest += 0.001 / 86400
	shifted.GreatestState = JupiterGalileanSatellitePhenomenaWithInstance(shifted.Greatest, nil)
	if _, err := mutualEclipseBaselineDifference(event, shifted); err != nil {
		t.Fatal(err)
	}
	corrupted := event
	corrupted.GreatestState[0].ShadowOffsetXArcsec = 0
	if _, err := mutualEclipseBaselineDifference(event, corrupted); err == nil {
		t.Fatal("corrupt recorded state accepted")
	}
	corrupted = event
	corrupted.GreatestState[0].Eclipse = !corrupted.GreatestState[0].Eclipse
	if _, err := mutualEclipseBaselineDifference(event, corrupted); err == nil {
		t.Fatal("changed flag accepted")
	}
	shifted.Greatest = event.Greatest + 1.0/86400
	shifted.GreatestState = JupiterGalileanSatellitePhenomenaWithInstance(shifted.Greatest, nil)
	if _, err := mutualEclipseBaselineDifference(event, shifted); err == nil {
		t.Fatal("excessive shadow or Greatest difference accepted")
	}
	shifted.Greatest = event.Greatest + 0.049/86400
	shifted.GreatestState = JupiterGalileanSatellitePhenomenaWithInstance(shifted.Greatest, nil)
	if _, err := mutualEclipseBaselineDifference(event, shifted); err == nil || !strings.Contains(err.Error(), "shadow difference") {
		t.Fatalf("shadow limit not checked independently of Greatest: %v", err)
	}
}

const (
	mutualExternalGreatestToleranceSeconds  = 30.0
	mutualExternalContactToleranceSeconds   = 60.0
	mutualExternalDurationToleranceSeconds  = 60.0
	mutualExternalDistanceToleranceKM       = 45.0
	mutualExternalSeparationToleranceArcsec = 0.15
	mutualExternalCurveTimeToleranceSeconds = 20.0
)

type mutualExternalRecord struct {
	Label           string  `json:"label"`
	Type            string  `json:"type"`
	Active          int     `json:"active"`
	Passive         int     `json:"passive"`
	StartTT         string  `json:"start_tt"`
	GreatestTT      string  `json:"greatest_tt"`
	EndTT           string  `json:"end_tt"`
	DurationMinutes float64 `json:"duration_minutes"`
	ImpactFactor    float64 `json:"impact_factor"`
	Source          string  `json:"source"`
	TimeScale       string  `json:"time_scale"`
	SourceURL       string  `json:"source_url"`
	ResponseSHA256  string  `json:"response_sha256"`
	Subtype         string  `json:"subtype"`
	OriginalRecord  struct {
		Events map[string]struct {
			Visibility *bool `json:"visibility"`
		} `json:"events"`
	} `json:"original_record"`
}

type mutualCurveSample struct {
	TT                string  `json:"tt"`
	SeparationArcsecs float64 `json:"separation_arcsec"`
	GeometryJD        float64 `json:"geometry_jd_tt"`
}

type mutualSeparationCurve struct {
	Label      string              `json:"label"`
	Type       string              `json:"type"`
	Active     int                 `json:"active"`
	Passive    int                 `json:"passive"`
	Category   string              `json:"category"`
	Samples    []mutualCurveSample `json:"samples"`
	Source     string              `json:"source"`
	Meaning    string              `json:"sample_time_meaning"`
	Correction string              `json:"vector_correction"`
}

func TestJupiterGalileanMutualEclipseExternalPredictions(t *testing.T) {
	mutualEclipseExternalCensus(t)
}

func TestJupiterGalileanMutualEclipseExternalSeparation(t *testing.T) {
	mutualEclipseExternalCurves(t)
}

type mutualEclipseExternalComparison struct {
	mutualExternalComparison
	greatestKM  float64
	impactDelta float64
	penumbral   bool
}

func mutualEclipseExternalVisibility(jde float64, passive int) (bool, string) {
	context := newJupiterGalileanObservationContext(jde)
	observation := context.observationForSatellite(passive - 1)
	reason := ""
	if context.phenomenonForSatellite(passive - 1).Eclipse {
		reason += "Passive.Eclipse;"
	}
	limb := ellipseSignedDistance(observation.OffsetXJupiterRadii, observation.OffsetYJupiterRadii, 1, context.earthMinorRadius)
	if !observation.InFrontOfJupiter && limb < jupiterGalileanSatelliteRadiusJupiterRadii(passive) {
		reason += "Passive.OccultationDisk;"
	}
	return reason == "", reason
}

func mutualEclipseExternalClass(geometry jupiterGalileanMutualEclipseGeometry) (string, float64) {
	if geometry.UmbraKM > 0 {
		if geometry.SeparationKM+geometry.PassiveRadiusKM <= geometry.UmbraKM {
			return "Total", geometry.UmbraKM - geometry.SeparationKM - geometry.PassiveRadiusKM
		}
		if geometry.SeparationKM+geometry.UmbraKM <= geometry.PassiveRadiusKM {
			return "Annular", geometry.PassiveRadiusKM - geometry.SeparationKM - geometry.UmbraKM
		}
		if geometry.SeparationKM < geometry.UmbraKM+geometry.PassiveRadiusKM {
			return "Partial", math.Min(geometry.UmbraKM+geometry.PassiveRadiusKM-geometry.SeparationKM, geometry.SeparationKM-math.Abs(geometry.PassiveRadiusKM-geometry.UmbraKM))
		}
	}
	return "Penumbral", geometry.SeparationKM - geometry.UmbraKM - geometry.PassiveRadiusKM
}

func mutualEclipseExternalCensus(t *testing.T) {
	t.Helper()
	records := loadMutualExternalRecords(t)
	start, end := math.Inf(1), math.Inf(-1)
	for _, record := range records {
		if record.Type == string(JupiterGalileanMutualEclipse) {
			start = math.Min(start, mutualExternalJD(t, record.StartTT))
			end = math.Max(end, mutualExternalJD(t, record.EndTT))
		}
	}
	models := make(map[[2]int][]JupiterGalileanMutualEvent)
	matched := make(map[[2]int][]bool)
	for _, pair := range [][2]int{{1, 2}, {1, 3}, {1, 4}, {2, 3}, {2, 4}, {3, 4}} {
		models[pair] = collectMutualEclipsesBySearch(start, end, pair[0], pair[1])
		matched[pair] = make([]bool, len(models[pair]))
		t.Logf("eclipse prediction coverage pair=%v start=%.12f end=%.12f events=%d", pair, start, end, len(models[pair]))
	}
	var comparisons []mutualEclipseExternalComparison
	classes := make(map[string]int)
	categories := make(map[string]bool)
	pairsCovered := make(map[[2]int]bool)
	classMatches, visibilityMatches, secondsMatches, predictionOnly := 0, 0, 0, 0
	for _, record := range records {
		if record.Type != string(JupiterGalileanMutualEclipse) {
			continue
		}
		first, second, _ := normalizeJupiterGalileanMutualPair(record.Active, record.Passive)
		pair := [2]int{first, second}
		pairsCovered[pair] = true
		reference := mutualExternalJD(t, record.GreatestTT)
		match := -1
		for index, event := range models[pair] {
			if event.Active == record.Active && event.Passive == record.Passive && math.Abs(event.Greatest-reference) <= 3.0/24 {
				if match >= 0 || matched[pair][index] {
					t.Fatalf("nonunique eclipse match: %s", record.Label)
				}
				match = index
			}
		}
		if match < 0 {
			search := jupiterGalileanEclipseSearch{first: first, second: second}
			center := jupiterGalileanEclipseEmission(reference, record.Passive, nil)
			best := center
			minimum := math.Inf(1)
			for minute := -180; minute <= 180; minute++ {
				candidate := center + float64(minute)/1440
				geometry, ok := search.geometry(candidate)
				if ok && geometry.SeparationKM < minimum {
					best, minimum = candidate, geometry.SeparationKM
				}
			}
			best = search.minimum(best-1.0/1440, best+1.0/1440)
			geometry, ok := search.geometry(best)
			gap := geometry.SeparationKM - geometry.PenumbraKM - geometry.PassiveRadiusKM
			delta := (jupiterGalileanEclipseReception(best, record.Passive, nil) - reference) * 86400
			t.Logf("eclipse prediction_only label=%s gap=%+.6fkm closest_delta=%+.6fs", record.Label, gap, delta)
			if !ok || geometry.Active != record.Active || geometry.Passive != record.Passive || !isFinite(gap) || gap < 0 || gap > mutualExternalDistanceToleranceKM || math.Abs(delta) > mutualExternalGreatestToleranceSeconds {
				t.Fatalf("eclipse prediction-only failed: %s gap=%.6fkm delta=%.6fs", record.Label, gap, delta)
			}
			predictionOnly++
			continue
		}
		matched[pair][match] = true
		event := models[pair][match]
		if visible, reason := mutualEclipseExternalVisibility(event.Greatest, event.Passive); !visible {
			t.Fatalf("matched eclipse invisible: %s reason=%s", record.Label, reason)
		}
		for _, phase := range []struct{ name, epoch string }{{"begin", record.StartTT}, {"greatest", record.GreatestTT}, {"end", record.EndTT}} {
			expected, present := record.OriginalRecord.Events[phase.name]
			if !present || expected.Visibility == nil {
				t.Fatalf("missing eclipse visibility: %s %s", record.Label, phase.name)
			}
			visible, reason := mutualEclipseExternalVisibility(mutualExternalJD(t, phase.epoch), record.Passive)
			t.Logf("eclipse visibility label=%s phase=%s got=%v want=%v reason=%s", record.Label, phase.name, visible, *expected.Visibility, reason)
			if visible != *expected.Visibility {
				t.Fatalf("eclipse visibility mismatch: %s %s", record.Label, phase.name)
			}
			visibilityMatches++
		}
		geometryJD := jupiterGalileanEclipseEmission(event.Greatest, event.Passive, nil)
		geometry, ok := jupiterGalileanMutualEclipseGeometryAt(geometryJD, first, second, nil)
		if !ok {
			t.Fatalf("eclipse greatest geometry invalid: %s", record.Label)
		}
		class, margin := mutualEclipseExternalClass(geometry)
		classes[record.Subtype+"/"+class]++
		if class == record.Subtype {
			classMatches++
		} else {
			allowed := record.Label == "mutual_eclipse_1-2_20261003183601609" || record.Label == "mutual_eclipse_1-3_20261213152855626"
			t.Logf("eclipse classification label=%s reference=%s model=%s margin=%.6fkm", record.Label, record.Subtype, class, margin)
			if !allowed || record.Subtype != "Partial" || class != "Annular" || margin >= 5 {
				t.Fatalf("eclipse classification mismatch: %s %s/%s margin=%.6fkm", record.Label, record.Subtype, class, margin)
			}
		}
		categories["deep"] = categories["deep"] || event.ImpactParameter < 0.5
		categories["grazing"] = categories["grazing"] || event.ImpactParameter >= 0.9
		categories["annular"] = categories["annular"] || class == "Annular"
		categories["ordinary"] = categories["ordinary"] || class == "Partial" && event.ImpactParameter >= 0.5 && event.ImpactParameter < 0.9
		comparison := mutualEclipseExternalCompare(t, record, event, geometry)
		comparison.penumbral = class == "Penumbral"
		comparisons = append(comparisons, comparison)
		if math.Abs(comparison.differences[0]) <= mutualExternalGreatestToleranceSeconds {
			secondsMatches++
		}
	}
	excluded, modelOnly := 0, 0
	for pair, events := range models {
		if !pairsCovered[pair] {
			t.Fatalf("missing eclipse pair %v", pair)
		}
		for index, event := range events {
			if matched[pair][index] {
				continue
			}
			if visible, reason := mutualEclipseExternalVisibility(event.Greatest, event.Passive); !visible {
				t.Logf("eclipse model_only_excluded pair=%v active=%d passive=%d greatest_tt=%s b=%.9f reason=%s", pair, event.Active, event.Passive, JD2Date(event.Greatest).Format("2006-01-02T15:04:05.000"), event.ImpactParameter, reason)
				excluded++
				continue
			}
			geometry, ok := jupiterGalileanMutualEclipseGeometryAt(jupiterGalileanEclipseEmission(event.Greatest, event.Passive, nil), pair[0], pair[1], nil)
			depth := geometry.PenumbraKM + geometry.PassiveRadiusKM - geometry.SeparationKM
			t.Logf("eclipse model_only_visible pair=%v active=%d passive=%d greatest_tt=%s b=%.9f depth=%.6fkm", pair, event.Active, event.Passive, JD2Date(event.Greatest).Format("2006-01-02T15:04:05.000"), event.ImpactParameter, depth)
			if !ok || !isFinite(depth) || depth < 0 || depth > mutualExternalDistanceToleranceKM {
				t.Fatalf("visible model-only eclipse too deep: %v %.6fkm", pair, depth)
			}
			modelOnly++
		}
	}
	for _, category := range []string{"ordinary", "deep", "grazing", "annular"} {
		if !categories[category] {
			t.Fatalf("missing eclipse category %s", category)
		}
	}
	keys := make([]string, 0, len(classes))
	for key := range classes {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		t.Logf("eclipse classification cross=%s count=%d", key, classes[key])
	}
	mutualEclipseExternalSummary(t, comparisons)
	t.Logf("eclipse external matched=%d greatest_seconds=%d greatest_seconds_or_km=%d class_match=%d visibility=%d excluded=%d visible_model_only=%d prediction_only=%d", len(comparisons), secondsMatches, len(comparisons), classMatches, visibilityMatches, excluded, modelOnly, predictionOnly)
	if predictionOnly == 0 {
		t.Log("eclipse prediction_only=[]; no forecast-only fixture exercises the gap rule")
	}
	if len(comparisons) != 173 || classMatches != 171 || visibilityMatches != 519 || excluded != 38 || modelOnly != 3 || predictionOnly != 0 {
		t.Fatal("production eclipse census does not reproduce the frozen reference counts")
	}
}

func mutualEclipseExternalCompare(t *testing.T, record mutualExternalRecord, event JupiterGalileanMutualEvent, geometry jupiterGalileanMutualEclipseGeometry) mutualEclipseExternalComparison {
	t.Helper()
	start, end := mutualExternalJD(t, record.StartTT), mutualExternalJD(t, record.EndTT)
	reference := mutualExternalJD(t, record.GreatestTT)
	refGeometry, ok := jupiterGalileanMutualEclipseGeometryAt(jupiterGalileanEclipseEmission(reference, event.Passive, nil), event.Active, event.Passive, nil)
	if !ok {
		t.Fatalf("reference eclipse geometry invalid: %s", record.Label)
	}
	comparison := mutualEclipseExternalComparison{greatestKM: refGeometry.SeparationKM - geometry.SeparationKM, impactDelta: event.ImpactParameter - record.ImpactFactor}
	comparison.impact = event.ImpactParameter
	comparison.differences = [4]float64{(event.Greatest - reference) * 86400, (event.Start - start) * 86400, (event.End - end) * 86400, ((event.End - event.Start) - (end - start)) * 86400}
	for index, contact := range []float64{event.Start, event.End} {
		epoch := jupiterGalileanEclipseEmission(contact, event.Passive, nil)
		speed := math.Abs(mutualEclipseSolarSeparationForTest(epoch+2.0/86400, event.Active, event.Passive)-mutualEclipseSolarSeparationForTest(epoch-2.0/86400, event.Active, event.Passive)) / 4
		comparison.equivalent[index] = comparison.differences[index+1] * speed
	}
	minimum := jupiterGalileanEclipseEmission(event.Greatest, event.Passive, nil)
	sensitivity := (mutualEclipseExternalDuration(t, minimum, event.Active, event.Passive, 1) - mutualEclipseExternalDuration(t, minimum, event.Active, event.Passive, -1)) / 2
	if !isFinite(sensitivity) || sensitivity <= 0 {
		t.Fatalf("eclipse duration sensitivity invalid: %s %.9f", record.Label, sensitivity)
	}
	comparison.equivalent[2] = comparison.differences[3] / sensitivity
	greatestBasis := "seconds"
	if math.Abs(comparison.differences[0]) > mutualExternalGreatestToleranceSeconds {
		greatestBasis = "km"
		if !isFinite(comparison.greatestKM) || math.Abs(comparison.greatestKM) > mutualExternalDistanceToleranceKM {
			t.Fatalf("eclipse greatest mismatch: %s %.6fs %.6fkm", record.Label, comparison.differences[0], comparison.greatestKM)
		}
	}
	for index, tolerance := range []float64{mutualExternalContactToleranceSeconds, mutualExternalContactToleranceSeconds, mutualExternalDurationToleranceSeconds} {
		comparison.basis[index] = "seconds"
		if math.Abs(comparison.differences[index+1]) > tolerance {
			comparison.basis[index] = "km"
			if math.Abs(comparison.equivalent[index]) > mutualExternalDistanceToleranceKM {
				t.Fatalf("eclipse contact mismatch: %s item=%d %.6fs %.6fkm", record.Label, index, comparison.differences[index+1], comparison.equivalent[index])
			}
		}
		if !isFinite(comparison.differences[index+1]) || !isFinite(comparison.equivalent[index]) {
			t.Fatalf("nonfinite eclipse comparison: %s", record.Label)
		}
	}
	t.Logf("eclipse prediction label=%s b=%.9f impact_delta=%+.9f greatest=%+.6fs start=%+.6fs end=%+.6fs duration=%+.6fs greatest_km=%+.6f start_km=%+.6f end_km=%+.6f duration_km=%+.6f sensitivity=%.6fs/km basis=%s/%s/%s/%s", record.Label, comparison.impact, comparison.impactDelta, comparison.differences[0], comparison.differences[1], comparison.differences[2], comparison.differences[3], comparison.greatestKM, comparison.equivalent[0], comparison.equivalent[1], comparison.equivalent[2], sensitivity, greatestBasis, comparison.basis[0], comparison.basis[1], comparison.basis[2])
	return comparison
}

func mutualEclipseExternalDuration(t *testing.T, minimum float64, first, second int, offset float64) float64 {
	t.Helper()
	search := jupiterGalileanEclipseSearch{first: first, second: second}
	gap, ok := search.gap(minimum)
	if !ok {
		t.Fatal("invalid perturbed eclipse geometry")
	}
	if gap >= offset {
		return 0
	}
	var roots [2]float64
	for index, direction := range []int{-1, 1} {
		inside := minimum
		found := false
		for minute := 1; minute <= 360; minute++ {
			outside := minimum + float64(direction*minute)/1440
			value, valid := search.gap(outside)
			if !valid {
				t.Fatal("invalid perturbed eclipse contact")
			}
			if value > offset {
				for iteration := 0; iteration < 80; iteration++ {
					middle := (inside + outside) / 2
					value, _ = search.gap(middle)
					if value > offset {
						outside = middle
					} else {
						inside = middle
					}
				}
				roots[index], found = (inside+outside)/2, true
				break
			}
			inside = outside
		}
		if !found {
			t.Fatal("no perturbed eclipse contact within six hours")
		}
	}
	geometry, _ := search.geometry(minimum)
	return (jupiterGalileanEclipseReception(roots[1], geometry.Passive, nil) - jupiterGalileanEclipseReception(roots[0], geometry.Passive, nil)) * 86400
}

func mutualEclipseExternalSummary(t *testing.T, comparisons []mutualEclipseExternalComparison) {
	t.Helper()
	for bin, name := range []string{"b<0.5", "0.5<=b<0.8", "0.8<=b<0.9", "b>=0.9", "pure_penumbra", "all"} {
		var values [9][]float64
		count := 0
		for _, row := range comparisons {
			if bin < 4 && (row.penumbral || mutualExternalImpactBin(row.impact) != bin) || bin == 4 && !row.penumbral {
				continue
			}
			count++
			for index, value := range row.differences {
				values[index] = append(values[index], math.Abs(value))
			}
			values[4] = append(values[4], math.Abs(row.greatestKM))
			for index, value := range row.equivalent {
				values[index+5] = append(values[index+5], math.Abs(value))
			}
			values[8] = append(values[8], math.Abs(row.impactDelta))
		}
		t.Logf("eclipse summary bin=%s count=%d", name, count)
		for index, field := range []string{"greatest_s", "start_s", "end_s", "duration_s", "greatest_km", "start_km", "end_km", "duration_km", "impact_delta"} {
			if len(values[index]) == 0 {
				continue
			}
			sort.Float64s(values[index])
			var quantiles [3]float64
			for position, fraction := range []float64{0.5, 0.9, 1} {
				rank := float64(len(values[index])-1) * fraction
				lower, upper := int(math.Floor(rank)), int(math.Ceil(rank))
				quantiles[position] = values[index][lower] + (rank-float64(lower))*(values[index][upper]-values[index][lower])
			}
			t.Logf("eclipse summary bin=%s field=%s abs_median=%.9f abs_p90=%.9f abs_max=%.9f", name, field, quantiles[0], quantiles[1], quantiles[2])
		}
	}
}

func mutualEclipseExternalCurves(t *testing.T) {
	t.Helper()
	content, err := os.ReadFile("../jupiter/testdata/galilean_mutual_horizons.json")
	if err != nil {
		t.Fatal(err)
	}
	var curves []mutualSeparationCurve
	if err := json.Unmarshal(content, &curves); err != nil {
		t.Fatal(err)
	}
	maximumAngle, maximumTime, maximumGrid := 0.0, 0.0, 0.0
	count := 0
	coverage := make(map[string]bool)
	for _, curve := range curves {
		if curve.Type != string(JupiterGalileanMutualEclipse) {
			continue
		}
		if !strings.Contains(curve.Meaning, "Earth reception TT") || !strings.Contains(curve.Correction, "NONE") || len(curve.Samples) != 181 {
			t.Fatalf("invalid eclipse curve time semantics: %s", curve.Label)
		}
		minimumIndex := 0
		for index, sample := range curve.Samples {
			geometry, ok := jupiterGalileanMutualEclipseGeometryAt(sample.GeometryJD, curve.Active, curve.Passive, nil)
			if !ok || sample.GeometryJD >= mutualCurveSampleJD(t, sample.TT) {
				t.Fatalf("invalid eclipse curve geometry epoch: %s", curve.Label)
			}
			difference := math.Abs(geometry.MinimumSeparationArcsec - sample.SeparationArcsecs)
			maximumGrid = math.Max(maximumGrid, difference)
			if difference > mutualExternalSeparationToleranceArcsec {
				t.Fatalf("eclipse curve sample mismatch: %s %.9f arcsec", curve.Label, difference)
			}
			if sample.SeparationArcsecs < curve.Samples[minimumIndex].SeparationArcsecs {
				minimumIndex = index
			}
		}
		if minimumIndex == 0 || minimumIndex == len(curve.Samples)-1 {
			t.Fatalf("eclipse curve minimum at edge: %s", curve.Label)
		}
		referenceTime, referenceAngle := mutualCurveMinimum(t, curve.Samples)
		search := jupiterGalileanEclipseSearch{first: curve.Active, second: curve.Passive}
		minimum := search.minimum(curve.Samples[minimumIndex-1].GeometryJD, curve.Samples[minimumIndex+1].GeometryJD)
		geometry, ok := search.geometry(minimum)
		if !ok || geometry.Active != curve.Active || geometry.Passive != curve.Passive {
			t.Fatalf("eclipse curve direction mismatch: %s", curve.Label)
		}
		angleDelta := geometry.MinimumSeparationArcsec - referenceAngle
		timeDelta := (jupiterGalileanEclipseReception(minimum, curve.Passive, nil) - referenceTime) * 86400
		maximumAngle, maximumTime = math.Max(maximumAngle, math.Abs(angleDelta)), math.Max(maximumTime, math.Abs(timeDelta))
		t.Logf("eclipse curve label=%s category=%s angle_delta=%+.9farcsec greatest_delta=%+.6fs geometry_tt=%.12f", curve.Label, curve.Category, angleDelta, timeDelta, minimum)
		if math.Abs(angleDelta) > mutualExternalSeparationToleranceArcsec || math.Abs(timeDelta) > mutualExternalCurveTimeToleranceSeconds {
			t.Fatalf("eclipse curve minimum mismatch: %s %.9farcsec %.6fs", curve.Label, angleDelta, timeDelta)
		}
		count++
		coverage[curve.Category] = true
	}
	for _, category := range []string{"ordinary", "deep", "grazing", "annular"} {
		if !coverage[category] {
			t.Fatalf("missing eclipse curve category %s", category)
		}
	}
	if count != 9 {
		t.Fatalf("eclipse curves=%d want=9", count)
	}
	t.Logf("eclipse curves=%d maximum_minimum_angle_difference=%.9farcsec / %.2f greatest_difference=%.6fs / %.0f maximum_grid_difference=%.9farcsec", count, maximumAngle, mutualExternalSeparationToleranceArcsec, maximumTime, mutualExternalCurveTimeToleranceSeconds, maximumGrid)
}

func TestJupiterGalileanMutualEclipseConsistencySearch(t *testing.T) {
	pairs := [][2]int{{1, 2}, {1, 3}, {1, 4}, {2, 3}, {2, 4}, {3, 4}}
	if testing.Short() {
		pairs = pairs[1:2]
	}
	for _, pair := range pairs {
		want := NextJupiterGalileanMutualEvent(JDCalc(2026, 1, 1), pair[0], pair[1], JupiterGalileanMutualEclipse)
		if !want.Valid || want.Type != JupiterGalileanMutualEclipse || !(want.Start < want.Greatest && want.Greatest < want.End) {
			t.Fatalf("pair %v invalid eclipse: %+v", pair, want)
		}
		for _, value := range []float64{want.Start, want.Greatest, want.End, want.DurationDays, want.MinimumSeparationArcsec, want.ImpactParameter, want.Coverage} {
			if !isFinite(value) {
				t.Fatalf("pair %v nonfinite eclipse field", pair)
			}
		}
		if want.DurationDays != want.End-want.Start || want.Coverage < 0 || want.Coverage > 1 || want.ImpactParameter >= 1 {
			t.Fatalf("pair %v invalid metrics: %+v", pair, want)
		}
		assertSameJupiterGalileanMutualEvent(t, NextJupiterGalileanMutualEvent(want.Start-1.0/1440, pair[1], pair[0], JupiterGalileanMutualEclipse), want)
		for _, seed := range []float64{want.Start - 1.0/1440, (want.Start + want.Greatest) / 2, want.Greatest, (want.Greatest + want.End) / 2, want.End + 1.0/1440} {
			last := LastJupiterGalileanMutualEvent(seed, pair[0], pair[1], JupiterGalileanMutualEclipse)
			next := NextJupiterGalileanMutualEvent(seed, pair[0], pair[1], JupiterGalileanMutualEclipse)
			assertSameJupiterGalileanMutualEvent(t, ClosestJupiterGalileanMutualEvent(seed, pair[0], pair[1], JupiterGalileanMutualEclipse), want)
			if seed >= want.Greatest {
				assertSameJupiterGalileanMutualEvent(t, last, want)
			} else if !last.Valid || last.Greatest > seed {
				t.Fatalf("last direction at %.12f: %+v", seed, last)
			}
			if seed < want.Greatest {
				assertSameJupiterGalileanMutualEvent(t, next, want)
			} else if !next.Valid || next.Greatest <= seed || next.Start < want.End {
				t.Fatalf("next direction at %.12f: %+v", seed, next)
			}
		}
		withGate, found := searchJupiterGalileanMutualEventWithSeasonGate(want.Start-1, pair[0], pair[1], JupiterGalileanMutualEclipse, 1, false, true, nil)
		withoutGate, foundWithout := searchJupiterGalileanMutualEventWithSeasonGate(want.Start-1, pair[0], pair[1], JupiterGalileanMutualEclipse, 1, false, false, nil)
		if !found || !foundWithout {
			t.Fatalf("pair %v gate search failed", pair)
		}
		assertSameJupiterGalileanMutualEvent(t, withGate, withoutGate)
	}
}

func TestJupiterGalileanMutualEclipseSyntheticBoundaries(t *testing.T) {
	for _, pair := range [][2]int{{0, 2}, {1, 1}, {1, 5}} {
		for _, query := range []func(float64, int, int, JupiterGalileanMutualPhenomenonType) JupiterGalileanMutualEvent{LastJupiterGalileanMutualEvent, NextJupiterGalileanMutualEvent, ClosestJupiterGalileanMutualEvent} {
			if event := query(JDCalc(2026, 1, 1), pair[0], pair[1], JupiterGalileanMutualEclipse); event != (JupiterGalileanMutualEvent{}) {
				t.Fatalf("invalid pair %v did not return zero: %+v", pair, event)
			}
		}
	}
	for _, seed := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if event := NextJupiterGalileanMutualEvent(seed, 1, 2, JupiterGalileanMutualEclipse); event != (JupiterGalileanMutualEvent{}) {
			t.Fatalf("nonfinite seed did not return zero: %+v", event)
		}
	}
	search := jupiterGalileanEclipseSearch{first: 1, second: 3}
	if search.seasonAllows(JDCalc(2024, 1, 1)) || !search.seasonAllows(JDCalc(2027, 3, 2)) {
		t.Fatal("solar season gate does not distinguish seasons")
	}
}

func TestJupiterGalileanMutualEclipseConsistencyDoubleMinimum(t *testing.T) {
	seed := JDCalc(2027, 3, 2)
	event := NextJupiterGalileanMutualEvent(seed, 1, 3, JupiterGalileanMutualEclipse)
	if !event.Valid {
		t.Fatal("double-minimum eclipse missing")
	}
	geometryJD := jupiterGalileanEclipseEmission(event.Greatest, event.Passive, nil)
	geometry, ok := jupiterGalileanMutualEclipseGeometryAt(geometryJD, 1, 3, nil)
	if !ok || math.Abs(geometryJD-2461467.17205)*86400 > 30 || math.Abs(geometry.SeparationKM-3411) > 5 || event.Active != 1 || event.Passive != 3 {
		t.Fatalf("wrong minimum %.12f: %+v", geometryJD, geometry)
	}
	shallow := jupiterGalileanEclipseReception(2461467.2325, 3, nil)
	if shallow <= event.Start || shallow >= event.End {
		t.Fatalf("shallow minimum outside connected eclipse: %+v", event)
	}
	assertSameJupiterGalileanMutualEvent(t, LastJupiterGalileanMutualEvent(event.End+0.1/86400, 1, 3, JupiterGalileanMutualEclipse), event)
	assertSameJupiterGalileanMutualEvent(t, ClosestJupiterGalileanMutualEvent(shallow, 1, 3, JupiterGalileanMutualEclipse), event)
	next := NextJupiterGalileanMutualEvent(event.End+0.1/86400, 1, 3, JupiterGalileanMutualEclipse)
	if !next.Valid || next.Start <= event.End {
		t.Fatalf("same connected eclipse returned twice: %+v", next)
	}
	t.Logf("deep minimum geometry TT=%.12f s=%.6fkm; one connected event %.12f..%.12f", geometryJD, geometry.SeparationKM, event.Start, event.End)
}

func TestJupiterGalileanMutualEclipseConsistencySpeedBound(t *testing.T) {
	if testing.Short() {
		t.Skip("two-year solar-plane speed survey")
	}
	start, end := JDCalc(2026, 1, 1), JDCalc(2028, 1, 1)
	for _, pair := range [][2]int{{1, 2}, {1, 3}, {1, 4}, {2, 3}, {2, 4}, {3, 4}} {
		maximum := 0.0
		bound := jupiterGalileanMutualRelativeSpeedUpperBound(pair[0], pair[1]) * jupiterGalileanEquatorialRadiusKM / jupiterGalileanMutualSpeedSafetyFactor
		for sample := 0; ; sample++ {
			jd := start + float64(sample)/48
			if jd > end {
				break
			}
			frame := newJupiterGalileanEclipseFrame(jd, nil)
			first := frame.satellite(jd, pair[0])
			second := frame.satellite(jd, pair[1])
			active, passive := pair[0], pair[1]
			if vectorMagnitude(first) > vectorMagnitude(second) {
				active, passive = passive, active
			}
			before := mutualEclipseSolarSeparationForTest(jd-30.0/86400, active, passive)
			after := mutualEclipseSolarSeparationForTest(jd+30.0/86400, active, passive)
			maximum = math.Max(maximum, math.Abs(after-before)*1440/bound)
		}
		if maximum > 1.05 {
			t.Fatalf("pair %v speed ratio %.9f > 1.05", pair, maximum)
		}
		t.Logf("solar-plane pair=%v maximum speed ratio=%.9f", pair, maximum)
	}
}

func mutualEclipseSolarSeparationForTest(jde float64, active, passive int) float64 {
	frame := newJupiterGalileanEclipseFrame(jde, nil)
	first := frame.satellite(jde, active)
	second := frame.satellite(jde, passive)
	axis := normalizeVector(first)
	delta := Vector3{second[0] - first[0], second[1] - first[1], second[2] - first[2]}
	axial := vectorDot(delta, axis)
	return vectorMagnitude(Vector3{delta[0] - axial*axis[0], delta[1] - axial*axis[1], delta[2] - axial*axis[2]}) * astronomicalUnitKM
}

func TestJupiterGalileanMutualEclipseConsistencyHourlyDense(t *testing.T) {
	if testing.Short() {
		t.Skip("two-year six-pair dense eclipse scan")
	}
	start, end := JDCalc(2026, 1, 1), JDCalc(2028, 1, 1)
	for _, pair := range [][2]int{{1, 2}, {1, 3}, {1, 4}, {2, 3}, {2, 4}, {3, 4}} {
		hourly := collectMutualEclipsesBySearch(start, end, pair[0], pair[1])
		dense := collectMutualEclipsesByDenseScan(start, end, pair[0], pair[1])
		if len(hourly) != len(dense) {
			t.Fatalf("pair %v hourly=%d dense=%d", pair, len(hourly), len(dense))
		}
		for index := range hourly {
			assertSameJupiterGalileanMutualEvent(t, hourly[index], dense[index])
		}
		t.Logf("eclipse pair=%v hourly=%d dense=%d bitwise equal", pair, len(hourly), len(dense))
	}
}

func collectMutualEclipsesBySearch(start, end float64, first, second int) []JupiterGalileanMutualEvent {
	var events []JupiterGalileanMutualEvent
	seed := start - 0.5
	for count := 0; count < 512; count++ {
		event := NextJupiterGalileanMutualEvent(seed, first, second, JupiterGalileanMutualEclipse)
		if !event.Valid || event.Greatest >= end {
			break
		}
		if event.Greatest >= start {
			events = append(events, event)
		}
		seed = event.End + 0.1/86400
	}
	return events
}

func collectMutualEclipsesByDenseScan(start, end float64, first, second int) []JupiterGalileanMutualEvent {
	var events []JupiterGalileanMutualEvent
	search := jupiterGalileanEclipseSearch{first: first, second: second}
	lastEnd := start - 1
	for sample := 0; ; sample++ {
		left := start + float64(sample)/288
		if left >= end {
			break
		}
		if left < lastEnd {
			continue
		}
		hit, found := search.hit(left, left+1.0/288, 1)
		if !found {
			continue
		}
		event, found := search.eventAround(hit)
		if found && event.Greatest >= start && event.Greatest < end && event.End > lastEnd {
			events = append(events, event)
			lastEnd = jupiterGalileanEclipseEmission(event.End, event.Passive, nil)
		}
	}
	return events
}

func TestJupiterGalileanMutualEventsConsistency(t *testing.T) {
	seed := JDCalc(2026, 1, 1.5)
	pairs := [][2]int{{1, 2}, {1, 3}, {1, 4}, {2, 3}, {2, 4}, {3, 4}}
	for _, pair := range pairs {
		t.Run(testMutualPairName(pair[0], pair[1]), func(t *testing.T) {
			forward := NextJupiterGalileanMutualEvent(seed, pair[0], pair[1], JupiterGalileanMutualOccultation)
			reverse := NextJupiterGalileanMutualEvent(seed, pair[1], pair[0], JupiterGalileanMutualOccultation)
			if !forward.Valid || !reverse.Valid {
				t.Fatalf("next event invalid: forward=%+v reverse=%+v", forward, reverse)
			}
			assertSameJupiterGalileanMutualEvent(t, forward, reverse)
			if !(forward.Start <= forward.Greatest && forward.Greatest <= forward.End) {
				t.Fatalf("event interval invalid: start=%.12f greatest=%.12f end=%.12f", forward.Start, forward.Greatest, forward.End)
			}
			last := LastJupiterGalileanMutualEvent(forward.End+1, pair[0], pair[1], JupiterGalileanMutualOccultation)
			closest := ClosestJupiterGalileanMutualEvent(forward.Greatest, pair[0], pair[1], JupiterGalileanMutualOccultation)
			if !last.Valid || !closest.Valid {
				t.Fatalf("last or closest event invalid: last=%+v closest=%+v", last, closest)
			}
		})
	}
}

func TestJupiterGalileanMutualEventsConsistencySeedPosition(t *testing.T) {
	seed := JDCalc(2026, 1, 1)
	pairs := [][2]int{{1, 2}, {1, 3}, {1, 4}, {2, 3}, {2, 4}, {3, 4}}
	if testing.Short() {
		pairs = pairs[1:2]
	}
	for _, pair := range pairs {
		pair := pair
		t.Run(testMutualPairName(pair[0], pair[1]), func(t *testing.T) {
			want := NextJupiterGalileanMutualEvent(seed, pair[0], pair[1], JupiterGalileanMutualOccultation)
			if !want.Valid {
				t.Fatalf("first event invalid: %+v", want)
			}
			tests := []struct {
				name       string
				query      float64
				expectLast bool
				expectNext bool
			}{
				{name: "before start", query: want.Start - 1.0/1440, expectNext: true},
				{name: "before greatest", query: (want.Start + want.Greatest) / 2, expectNext: true},
				{name: "at greatest", query: want.Greatest, expectLast: true},
				{name: "after greatest", query: (want.Greatest + want.End) / 2, expectLast: true},
				{name: "after end", query: want.End + 1.0/1440, expectLast: true},
			}
			for _, test := range tests {
				t.Run(test.name, func(t *testing.T) {
					last := LastJupiterGalileanMutualEvent(test.query, pair[0], pair[1], JupiterGalileanMutualOccultation)
					next := NextJupiterGalileanMutualEvent(test.query, pair[0], pair[1], JupiterGalileanMutualOccultation)
					closest := ClosestJupiterGalileanMutualEvent(test.query, pair[0], pair[1], JupiterGalileanMutualOccultation)
					if !closest.Valid {
						t.Fatalf("closest event invalid: %+v", closest)
					}
					assertSameJupiterGalileanMutualEvent(t, closest, want)
					if test.expectLast {
						assertSameJupiterGalileanMutualEvent(t, last, want)
					} else if !last.Valid || !(last.Greatest < test.query) || sameMutualFloat(last.Greatest, want.Greatest) {
						t.Fatalf("last event has wrong direction: query=%.12f last=%+v want=%.12f", test.query, last, want.Greatest)
					}
					if test.expectNext {
						assertSameJupiterGalileanMutualEvent(t, next, want)
					} else if !next.Valid || !(next.Greatest > test.query) || sameMutualFloat(next.Greatest, want.Greatest) {
						t.Fatalf("next event has wrong direction: query=%.12f next=%+v want=%.12f", test.query, next, want.Greatest)
					}
				})
			}
		})
	}
}

func TestJupiterGalileanMutualEventsConsistencyDoubleMinimum(t *testing.T) {
	const seed = 2461196.1
	const seedStep = 2.5 / 1440
	const firstGreatest = 2461199.431865
	const secondGreatest = 2461199.521703
	seedCount := 24
	if testing.Short() {
		seedCount = 1
	}
	var first JupiterGalileanMutualEvent
	for index := 0; index < seedCount; index++ {
		seedIndex := index
		if testing.Short() {
			seedIndex = 21
		}
		candidate := NextJupiterGalileanMutualEvent(seed+float64(seedIndex)*seedStep, 1, 2, JupiterGalileanMutualOccultation)
		if !candidate.Valid {
			t.Fatalf("seed k=%d returned invalid event: %+v", seedIndex, candidate)
		}
		if difference := math.Abs(candidate.Greatest-firstGreatest) * 86400; difference > 1 {
			t.Fatalf("seed k=%d first double-minimum event mismatch: greatest=%.12f difference=%.3fs", seedIndex, candidate.Greatest, difference)
		}
		if index == 0 {
			first = candidate
		} else {
			assertSameJupiterGalileanMutualEvent(t, candidate, first)
		}
	}
	second := NextJupiterGalileanMutualEvent(first.End+0.1/86400, 1, 2, JupiterGalileanMutualOccultation)
	if !second.Valid {
		t.Fatalf("second event invalid: %+v", second)
	}
	if difference := math.Abs(second.Greatest-secondGreatest) * 86400; difference > 1 {
		t.Fatalf("second double-minimum event mismatch: greatest=%.12f difference=%.3fs", second.Greatest, difference)
	}
	if first.Active != second.Active || first.Passive != second.Passive {
		t.Fatalf("active/passive mismatch: first=%d/%d second=%d/%d", first.Active, first.Passive, second.Active, second.Passive)
	}
}

func TestJupiterGalileanMutualEventsConsistencySpeedBound(t *testing.T) {
	start := JDCalc(2026, 1, 1)
	step := 30.0 / 1440.0
	derivativeStep := 30.0 / 86400.0
	for _, pair := range [][2]int{{1, 2}, {1, 3}, {1, 4}, {2, 3}, {2, 4}, {3, 4}} {
		pair := pair
		t.Run(testMutualPairName(pair[0], pair[1]), func(t *testing.T) {
			baseSpeed := jupiterGalileanMutualRelativeSpeedUpperBound(pair[0], pair[1]) / 1.25
			maxRatio := 0.0
			for index := 0; index <= 30*48; index++ {
				jde := start + float64(index)*step
				left, leftOK := jupiterGalileanMutualGeometryAt(jde-derivativeStep, pair[0], pair[1], nil)
				right, rightOK := jupiterGalileanMutualGeometryAt(jde+derivativeStep, pair[0], pair[1], nil)
				if !leftOK || !rightOK {
					t.Fatalf("geometry invalid at %.12f", jde)
				}
				speed := math.Abs(right.Separation-left.Separation) / (2 * derivativeStep)
				ratio := speed / baseSpeed
				if ratio > maxRatio {
					maxRatio = ratio
				}
				if ratio > 1.05 {
					t.Fatalf("speed bound exceeded at %.12f: ratio=%.6f", jde, ratio)
				}
			}
			t.Logf("pair %d-%d maximum |ds/dt|/(v_a+v_b): %.6f", pair[0], pair[1], maxRatio)
		})
	}
}

func TestJupiterGalileanMutualEventsConsistencyPerformance(t *testing.T) {
	start := JDCalc(2026, 1, 1)
	end := JDCalc(2028, 1, 1)
	for _, pair := range [][2]int{{1, 2}, {1, 3}, {1, 4}, {2, 3}, {2, 4}, {3, 4}} {
		started := time.Now()
		first := NextJupiterGalileanMutualEvent(start, pair[0], pair[1], JupiterGalileanMutualOccultation)
		firstElapsed := time.Since(started)
		if !first.Valid {
			t.Fatalf("pair %d-%d first event invalid", pair[0], pair[1])
		}
		events := []JupiterGalileanMutualEvent{first}
		seed := first.End + 0.1/86400
		var chainTotal time.Duration
		var chainMaximum time.Duration
		chainCalls := 0
		for len(events) < 512 {
			started = time.Now()
			event := NextJupiterGalileanMutualEvent(seed, pair[0], pair[1], JupiterGalileanMutualOccultation)
			elapsed := time.Since(started)
			if !event.Valid || event.Greatest >= end {
				break
			}
			events = append(events, event)
			seed = event.End + 0.1/86400
			chainTotal += elapsed
			if elapsed > chainMaximum {
				chainMaximum = elapsed
			}
			chainCalls++
		}
		chainAverage := time.Duration(0)
		if chainCalls > 0 {
			chainAverage = chainTotal / time.Duration(chainCalls)
		}
		t.Logf("pair %d-%d: seasonal chain average=%.3fms maximum=%.3fms calls=%d; from 2026-01-01 single=%.3fms; events_2026_2027=%d", pair[0], pair[1], float64(chainAverage)/float64(time.Millisecond), float64(chainMaximum)/float64(time.Millisecond), chainCalls, float64(firstElapsed)/float64(time.Millisecond), len(events))
	}
}

func TestJupiterGalileanMutualEventsSyntheticBoundaries(t *testing.T) {
	tests := []struct {
		name      string
		jde       float64
		a         int
		b         int
		typeValue JupiterGalileanMutualPhenomenonType
	}{
		{name: "same satellite", jde: JDCalc(2026, 1, 1.5), a: 1, b: 1, typeValue: JupiterGalileanMutualOccultation},
		{name: "zero satellite", jde: JDCalc(2026, 1, 1.5), a: 0, b: 2, typeValue: JupiterGalileanMutualOccultation},
		{name: "high satellite", jde: JDCalc(2026, 1, 1.5), a: 5, b: 2, typeValue: JupiterGalileanMutualOccultation},
		{name: "nonfinite jde", jde: math.NaN(), a: 1, b: 2, typeValue: JupiterGalileanMutualOccultation},
		{name: "unsupported type", jde: JDCalc(2026, 1, 1.5), a: 1, b: 2, typeValue: "invalid"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, query := range []func() JupiterGalileanMutualEvent{
				func() JupiterGalileanMutualEvent {
					return LastJupiterGalileanMutualEvent(test.jde, test.a, test.b, test.typeValue)
				},
				func() JupiterGalileanMutualEvent {
					return NextJupiterGalileanMutualEvent(test.jde, test.a, test.b, test.typeValue)
				},
				func() JupiterGalileanMutualEvent {
					return ClosestJupiterGalileanMutualEvent(test.jde, test.a, test.b, test.typeValue)
				},
			} {
				if event := query(); event.Valid {
					t.Fatalf("invalid input returned valid event: %+v", event)
				}
			}
		})
	}
}

func TestJupiterGalileanMutualEventsConsistencySeasonGate(t *testing.T) {
	seed := JDCalc(2026, 1, 1.5)
	for _, pair := range [][2]int{{1, 2}, {1, 3}, {1, 4}, {2, 3}, {2, 4}, {3, 4}} {
		withGate, withGateOK := searchJupiterGalileanMutualEventWithSeasonGate(seed, pair[0], pair[1], JupiterGalileanMutualOccultation, 1, false, true, nil)
		withoutGate, withoutGateOK := searchJupiterGalileanMutualEventWithSeasonGate(seed, pair[0], pair[1], JupiterGalileanMutualOccultation, 1, false, false, nil)
		if !withGateOK || !withoutGateOK {
			t.Fatalf("pair %d-%d search failed: with gate=%v without gate=%v", pair[0], pair[1], withGateOK, withoutGateOK)
		}
		assertSameJupiterGalileanMutualTimelineWithin(t, withGate, withoutGate, 1)
	}
}

func TestJupiterGalileanMutualEventsConsistencyOneHourVsFiveMinutes(t *testing.T) {
	if testing.Short() {
		t.Skip("two-year six-pair dense scan")
	}
	start := JDCalc(2026, 1, 1)
	end := JDCalc(2028, 1, 1)
	maxDifferenceSeconds := 0.0
	for _, pair := range [][2]int{{1, 2}, {1, 3}, {1, 4}, {2, 3}, {2, 4}, {3, 4}} {
		hourly := collectMutualEventsBySearch(start, end, pair[0], pair[1])
		dense := collectMutualEventsByFiveMinutes(start, end, pair[0], pair[1])
		if len(hourly) != len(dense) {
			t.Fatalf("pair %d-%d event count mismatch: hourly=%d dense=%d", pair[0], pair[1], len(hourly), len(dense))
		}
		for index := range hourly {
			differenceSeconds := math.Abs(hourly[index].Greatest-dense[index].Greatest) * 86400
			if differenceSeconds > maxDifferenceSeconds {
				maxDifferenceSeconds = differenceSeconds
			}
			if differenceSeconds > 30 {
				t.Fatalf("pair %d-%d event %d greatest mismatch: %.3fs", pair[0], pair[1], index, differenceSeconds)
			}
			if hourly[index].Active != dense[index].Active || hourly[index].Passive != dense[index].Passive {
				t.Fatalf("pair %d-%d event %d active/passive mismatch", pair[0], pair[1], index)
			}
		}
		t.Logf("pair %d-%d: hourly=%d dense=%d", pair[0], pair[1], len(hourly), len(dense))
	}
	t.Logf("one-hour versus five-minute maximum Greatest difference: %.3fs", maxDifferenceSeconds)
}

func TestJupiterGalileanMutualEventsExternalPredictions(t *testing.T) {
	records := loadMutualExternalRecords(t)
	coverageStart, coverageEnd := math.Inf(1), math.Inf(-1)
	for _, record := range records {
		if record.Type != "mutual_occultation" {
			continue
		}
		if record.Source == "" || record.ImpactFactor == 0 || record.Active < 1 || record.Active > 4 || record.Passive < 1 || record.Passive > 4 || record.Active == record.Passive {
			t.Fatalf("invalid record: %s", record.Label)
		}
		coverageStart = math.Min(coverageStart, mutualExternalJD(t, record.StartTT))
		coverageEnd = math.Max(coverageEnd, mutualExternalJD(t, record.EndTT))
	}
	if !isFinite(coverageStart) || coverageEnd <= coverageStart {
		t.Fatal("no mutual records")
	}
	models := make(map[[2]int][]JupiterGalileanMutualEvent)
	matched := make(map[[2]int][]bool)
	coverage := make(map[[2]int]bool)
	for _, pair := range [][2]int{{1, 2}, {1, 3}, {1, 4}, {2, 3}, {2, 4}, {3, 4}} {
		models[pair] = collectMutualEventsBySearch(coverageStart, coverageEnd, pair[0], pair[1])
		matched[pair] = make([]bool, len(models[pair]))
		t.Logf("prediction coverage pair=%d-%d start=%.12f end=%.12f model_events=%d", pair[0], pair[1], coverageStart, coverageEnd, len(models[pair]))
	}
	var comparisons []mutualExternalComparison
	var missingBins [4]int
	predictionOnly, modelOnly, excludedInvisible := 0, 0, 0
	for _, record := range records {
		if record.Type != "mutual_occultation" {
			continue
		}
		first, second, _ := normalizeJupiterGalileanMutualPair(record.Active, record.Passive)
		pair := [2]int{first, second}
		coverage[pair] = true
		greatest := mutualExternalJD(t, record.GreatestTT)
		event := NextJupiterGalileanMutualEvent(greatest-1.0/24, first, second, JupiterGalileanMutualOccultation)
		if !event.Valid || math.Abs(event.Greatest-greatest)*86400 > mutualExternalGreatestToleranceSeconds {
			closest := mutualExternalClosestApproach(t, greatest, first, second)
			geometry, ok := jupiterGalileanMutualGeometryAt(closest, first, second)
			radius := jupiterGalileanMutualContactRadius(first, second)
			gap := (geometry.Separation - radius) * jupiterGalileanEquatorialRadiusKM
			difference := (closest - greatest) * 86400
			impact := geometry.Separation / radius
			t.Logf("prediction_only label=%s b=%.9f gap=%+.6fkm closest_difference=%+.6fs basis=one_sided_gap", record.Label, impact, gap, difference)
			if !ok || geometry.Active != record.Active || geometry.Passive != record.Passive || !isFinite(gap) || gap < 0 || gap > mutualExternalDistanceToleranceKM || math.Abs(difference) > mutualExternalGreatestToleranceSeconds {
				t.Fatalf("prediction-only mismatch: %s gap=%.6fkm closest_difference=%+.6fs geometry=%+v", record.Label, gap, difference, geometry)
			}
			predictionOnly++
			missingBins[mutualExternalImpactBin(impact)]++
			continue
		}
		if event.Active != record.Active || event.Passive != record.Passive {
			t.Fatalf("foreground mismatch: %s got=%d/%d want=%d/%d", record.Label, event.Active, event.Passive, record.Active, record.Passive)
		}
		matchIndex := -1
		for index, candidate := range models[pair] {
			if candidate.Active == record.Active && candidate.Passive == record.Passive && math.Abs(candidate.Greatest-greatest)*86400 <= mutualExternalGreatestToleranceSeconds {
				if matchIndex >= 0 || matched[pair][index] {
					t.Fatalf("nonunique prediction match: %s", record.Label)
				}
				matchIndex = index
			}
		}
		if matchIndex < 0 {
			t.Fatalf("prediction absent from collected production events: %s", record.Label)
		}
		if reason := mutualExternalInvisibilityReason(event); reason != "" {
			t.Fatalf("matched prediction is invisible: %s reason=%s event=%+v", record.Label, reason, event)
		}
		if reason := mutualExternalInvisibilityReason(models[pair][matchIndex]); reason != "" {
			t.Fatalf("matched collected event is invisible: %s reason=%s event=%+v", record.Label, reason, models[pair][matchIndex])
		}
		t.Logf("matched visibility label=%s greatest=%.12f visible=true", record.Label, event.Greatest)
		matched[pair][matchIndex] = true
		comparisons = append(comparisons, mutualExternalComparePrediction(t, record, event))
	}
	for _, pair := range [][2]int{{1, 2}, {1, 3}, {1, 4}, {2, 3}, {2, 4}, {3, 4}} {
		if !coverage[pair] {
			t.Fatalf("missing satellite pair %d-%d", pair[0], pair[1])
		}
		for index, event := range models[pair] {
			if matched[pair][index] {
				continue
			}
			if reason := mutualExternalInvisibilityReason(event); reason != "" {
				t.Logf("model_only_excluded pair=%d-%d active=%d passive=%d greatest=%.12f greatest_tt=%s b=%.9f reason=%s", pair[0], pair[1], event.Active, event.Passive, event.Greatest, JD2Date(event.Greatest).Format("2006-01-02T15:04:05.000"), event.ImpactParameter, reason)
				excludedInvisible++
				continue
			}
			separation, ok := jupiterGalileanMutualSeparationAt(event.Greatest, pair[0], pair[1], nil)
			depth := (jupiterGalileanMutualContactRadius(pair[0], pair[1]) - separation) * jupiterGalileanEquatorialRadiusKM
			t.Logf("model_only pair=%d-%d active=%d passive=%d greatest=%.12f b=%.9f depth=%+.6fkm basis=one_sided_gap", pair[0], pair[1], event.Active, event.Passive, event.Greatest, event.ImpactParameter, depth)
			if !ok || !isFinite(depth) || depth < 0 || depth > mutualExternalDistanceToleranceKM {
				t.Fatalf("model-only mismatch: pair=%d-%d greatest=%.12f depth=%.6fkm", pair[0], pair[1], event.Greatest, depth)
			}
			modelOnly++
		}
	}
	if excludedInvisible != 16 {
		t.Fatalf("invisible model-only count mismatch: got=%d want=16", excludedInvisible)
	}
	mutualExternalLogSummary(t, comparisons, missingBins)
	t.Logf("external predictions=%d matched=%d matched_visible=%d prediction_only=%d model_only_visible=%d excluded_invisible=%d", len(comparisons)+predictionOnly, len(comparisons), len(comparisons), predictionOnly, modelOnly, excludedInvisible)
}

func mutualExternalInvisibilityReason(event JupiterGalileanMutualEvent) string {
	reason := ""
	if event.GreatestState[event.Active-1].Occultation {
		reason += "Active.Occultation;"
	}
	if event.GreatestState[event.Passive-1].Occultation {
		reason += "Passive.Occultation;"
	}
	if event.GreatestState[event.Passive-1].Eclipse {
		reason += "Passive.Eclipse;"
	}
	return reason
}

type mutualExternalComparison struct {
	impact      float64
	differences [4]float64
	equivalent  [3]float64
	basis       [3]string
}

func mutualExternalComparePrediction(t *testing.T, record mutualExternalRecord, event JupiterGalileanMutualEvent) mutualExternalComparison {
	t.Helper()
	start := mutualExternalJD(t, record.StartTT)
	end := mutualExternalJD(t, record.EndTT)
	comparison := mutualExternalComparison{impact: event.ImpactParameter, differences: [4]float64{
		(event.Greatest - mutualExternalJD(t, record.GreatestTT)) * 86400,
		(event.Start - start) * 86400,
		(event.End - end) * 86400,
		((event.End - event.Start) - (end - start)) * 86400,
	}}
	for index, contact := range []float64{event.Start, event.End} {
		before, beforeOK := jupiterGalileanMutualSeparationAt(contact-2.0/86400, record.Active, record.Passive, nil)
		after, afterOK := jupiterGalileanMutualSeparationAt(contact+2.0/86400, record.Active, record.Passive, nil)
		if !beforeOK || !afterOK {
			t.Fatalf("contact velocity invalid: %s", record.Label)
		}
		comparison.equivalent[index] = comparison.differences[index+1] * math.Abs(after-before) * jupiterGalileanEquatorialRadiusKM / 4
	}
	radius := jupiterGalileanMutualContactRadius(record.Active, record.Passive)
	wider := mutualExternalDurationAtRadius(t, event.Greatest, radius+1/jupiterGalileanEquatorialRadiusKM, record.Active, record.Passive)
	narrower := mutualExternalDurationAtRadius(t, event.Greatest, radius-1/jupiterGalileanEquatorialRadiusKM, record.Active, record.Passive)
	sensitivity := (wider - narrower) / 2
	if !isFinite(sensitivity) || sensitivity <= 0 {
		t.Fatalf("duration sensitivity invalid: %s %.9fs/km", record.Label, sensitivity)
	}
	comparison.equivalent[2] = comparison.differences[3] / sensitivity
	for index, tolerance := range []float64{mutualExternalContactToleranceSeconds, mutualExternalContactToleranceSeconds, mutualExternalDurationToleranceSeconds} {
		comparison.basis[index] = "fail"
		if isFinite(comparison.differences[index+1]) && math.Abs(comparison.differences[index+1]) <= tolerance {
			comparison.basis[index] = "seconds"
		} else if isFinite(comparison.equivalent[index]) && math.Abs(comparison.equivalent[index]) <= mutualExternalDistanceToleranceKM {
			comparison.basis[index] = "km"
		}
	}
	t.Logf("prediction label=%s b=%.9f greatest=%+.6fs start=%+.6fs end=%+.6fs duration=%+.6fs start_km=%+.6f end_km=%+.6f duration_km=%+.6f sensitivity=%.6fs/km basis=greatest:seconds,start:%s,end:%s,duration:%s", record.Label, comparison.impact, comparison.differences[0], comparison.differences[1], comparison.differences[2], comparison.differences[3], comparison.equivalent[0], comparison.equivalent[1], comparison.equivalent[2], sensitivity, comparison.basis[0], comparison.basis[1], comparison.basis[2])
	if !isFinite(comparison.differences[0]) || math.Abs(comparison.differences[0]) > mutualExternalGreatestToleranceSeconds {
		t.Fatalf("greatest mismatch: %s %+.6fs", record.Label, comparison.differences[0])
	}
	for index := range comparison.basis {
		if !isFinite(comparison.equivalent[index]) || comparison.basis[index] == "fail" {
			t.Fatalf("prediction mismatch: %s item=%d difference=%+.6fs equivalent=%+.6fkm", record.Label, index, comparison.differences[index+1], comparison.equivalent[index])
		}
	}
	return comparison
}

func mutualExternalDurationAtRadius(t *testing.T, greatest, radius float64, first, second int) float64 {
	t.Helper()
	separation, ok := jupiterGalileanMutualSeparationAt(greatest, first, second, nil)
	if !ok {
		t.Fatal("duration geometry invalid")
	}
	if separation >= radius {
		return 0
	}
	start, startOK := refineJupiterGalileanMutualBoundary(greatest, -1, radius*radius, first, second, nil)
	end, endOK := refineJupiterGalileanMutualBoundary(greatest, 1, radius*radius, first, second, nil)
	if !startOK || !endOK || end <= start {
		t.Fatal("perturbed-radius contacts invalid")
	}
	return (end - start) * 86400
}

func mutualExternalClosestApproach(t *testing.T, greatest float64, first, second int) float64 {
	t.Helper()
	var values [241]float64
	for index := range values {
		separation, ok := jupiterGalileanMutualSeparationAt(greatest+float64(index-120)/1440, first, second, nil)
		if !ok {
			t.Fatal("closest-approach geometry invalid")
		}
		values[index] = separation
	}
	best := -1
	for index := 1; index < len(values)-1; index++ {
		if values[index] <= values[index-1] && values[index] < values[index+1] && (best < 0 || math.Abs(float64(index-120)) < math.Abs(float64(best-120))) {
			best = index
		}
	}
	if best < 0 {
		t.Fatal("no local closest approach within two hours of prediction")
	}
	return refineJupiterGalileanMutualMinimum(greatest+float64(best-121)/1440, greatest+float64(best-119)/1440, first, second, nil)
}

func mutualExternalImpactBin(impact float64) int {
	for index, upper := range []float64{0.5, 0.8, 0.9} {
		if impact < upper {
			return index
		}
	}
	return 3
}

func mutualExternalLogSummary(t *testing.T, comparisons []mutualExternalComparison, missing [4]int) {
	t.Helper()
	for bin, name := range []string{"b<0.5", "0.5<=b<0.8", "0.8<=b<0.9", "b>=0.9", "all"} {
		var values [7][]float64
		var kmCounts [3]int
		count, absent := 0, 0
		for index, number := range missing {
			if bin == 4 || bin == index {
				absent += number
			}
		}
		for _, comparison := range comparisons {
			if bin != 4 && mutualExternalImpactBin(comparison.impact) != bin {
				continue
			}
			count++
			for index, difference := range comparison.differences {
				values[index] = append(values[index], math.Abs(difference))
			}
			for index, equivalent := range comparison.equivalent {
				values[index+4] = append(values[index+4], math.Abs(equivalent))
				if comparison.basis[index] == "km" {
					kmCounts[index]++
				}
			}
		}
		t.Logf("prediction summary bin=%s matched=%d prediction_only=%d km_basis_start_end_duration=%v", name, count, absent, kmCounts)
		for index, field := range []string{"greatest_s", "start_s", "end_s", "duration_s", "start_km", "end_km", "duration_km"} {
			if len(values[index]) == 0 {
				continue
			}
			sort.Float64s(values[index])
			quantiles := [3]float64{}
			for quantileIndex, fraction := range []float64{0.5, 0.9, 1} {
				rank := float64(len(values[index])-1) * fraction
				lower, upper := int(math.Floor(rank)), int(math.Ceil(rank))
				quantiles[quantileIndex] = values[index][lower] + (rank-float64(lower))*(values[index][upper]-values[index][lower])
			}
			t.Logf("prediction summary bin=%s field=%s abs_median=%.6f abs_p90=%.6f abs_max=%.6f", name, field, quantiles[0], quantiles[1], quantiles[2])
		}
	}
}

func TestJupiterGalileanMutualEventsExternalSeparation(t *testing.T) {
	records := loadMutualExternalRecords(t)
	recordByLabel := make(map[string]mutualExternalRecord, len(records))
	for _, record := range records {
		recordByLabel[record.Label] = record
	}
	contents, err := os.ReadFile("../jupiter/testdata/galilean_mutual_horizons.json")
	if err != nil {
		t.Fatal(err)
	}
	var curves []mutualSeparationCurve
	if err := json.Unmarshal(contents, &curves); err != nil {
		t.Fatal(err)
	}
	categoryCoverage := make(map[string]bool)
	pairCoverage := make(map[string]bool)
	maxSeparationDifference := 0.0
	maxGreatestDifference := 0.0
	count := 0
	for _, curve := range curves {
		if curve.Type != "mutual_occultation" {
			continue
		}
		if curve.Source == "" || len(curve.Samples) != 181 {
			t.Fatalf("invalid curve: %s", curve.Label)
		}
		record, found := recordByLabel[curve.Label]
		if !found {
			t.Fatalf("missing event record: %s", curve.Label)
		}
		seed := mutualExternalJD(t, record.StartTT) - 0.2/86400
		event := NextJupiterGalileanMutualEvent(seed, curve.Active, curve.Passive, JupiterGalileanMutualOccultation)
		if !event.Valid {
			t.Fatalf("event invalid: %s", curve.Label)
		}
		minimumJDE, minimumSeparation := mutualCurveMinimum(t, curve.Samples)
		separationDifference := math.Abs(event.MinimumSeparationArcsec - minimumSeparation)
		greatestDifference := math.Abs(event.Greatest-minimumJDE) * 86400
		maxSeparationDifference = math.Max(maxSeparationDifference, separationDifference)
		maxGreatestDifference = math.Max(maxGreatestDifference, greatestDifference)
		t.Logf("curve signed differences: %s separation=%+.6f arcsec greatest=%+.3fs", curve.Label, event.MinimumSeparationArcsec-minimumSeparation, (event.Greatest-minimumJDE)*86400)
		if separationDifference > mutualExternalSeparationToleranceArcsec || greatestDifference > mutualExternalCurveTimeToleranceSeconds {
			t.Fatalf("curve mismatch: %s separation=%.6f arcsec greatest=%.3fs", curve.Label, separationDifference, greatestDifference)
		}
		categoryCoverage[curve.Category] = true
		if curve.Category == "ordinary" {
			first, second := curve.Active, curve.Passive
			if first > second {
				first, second = second, first
			}
			pairCoverage[testMutualPairName(first, second)] = true
		}
		count++
	}
	for _, category := range []string{"ordinary", "deep", "grazing", "annular"} {
		if !categoryCoverage[category] {
			t.Fatalf("missing curve category %s", category)
		}
	}
	for _, pair := range [][2]int{{1, 2}, {1, 3}, {1, 4}, {2, 3}, {2, 4}, {3, 4}} {
		if !pairCoverage[testMutualPairName(pair[0], pair[1])] {
			t.Fatalf("missing ordinary curve for pair %d-%d", pair[0], pair[1])
		}
	}
	t.Logf("external curves=%d separation max=%.6f arcsec/0.15 arcsec greatest max=%.3fs/20s; all six pairs and four categories covered", count, maxSeparationDifference, maxGreatestDifference)
}

func loadMutualExternalRecords(t *testing.T) []mutualExternalRecord {
	t.Helper()
	contents, err := os.ReadFile("../jupiter/testdata/galilean_mutual_imcce_2026.json")
	if err != nil {
		t.Fatal(err)
	}
	var records []mutualExternalRecord
	if err := json.Unmarshal(contents, &records); err != nil {
		t.Fatal(err)
	}
	if len(records) == 0 {
		t.Fatal("no event records")
	}
	labels := make(map[string]bool, len(records))
	for _, record := range records {
		parsedURL, err := url.Parse(record.SourceURL)
		if err != nil || parsedURL.Scheme != "https" || parsedURL.Host != "opale.imcce.fr" || parsedURL.Query().Get("theory") != "INPOP19A,JUP365" || record.TimeScale != "TT" || len(record.ResponseSHA256) != 64 {
			t.Fatalf("invalid external model or provenance: %s", record.Label)
		}
		if record.Label == "" || labels[record.Label] {
			t.Fatalf("missing or duplicate external label: %s", record.Label)
		}
		labels[record.Label] = true
	}
	return records
}

func mutualExternalJD(t *testing.T, value string) float64 {
	t.Helper()
	instant, err := time.Parse("2006-01-02T15:04:05.000", value)
	if err != nil {
		t.Fatalf("parse event time %q: %v", value, err)
	}
	return Date2JD(instant)
}

func mutualCurveMinimum(t *testing.T, samples []mutualCurveSample) (float64, float64) {
	t.Helper()
	minimumIndex := 0
	for index := 1; index < len(samples); index++ {
		if samples[index].SeparationArcsecs < samples[minimumIndex].SeparationArcsecs {
			minimumIndex = index
		}
	}
	if minimumIndex == 0 || minimumIndex == len(samples)-1 {
		t.Fatalf("curve minimum is at an edge: index=%d", minimumIndex)
	}
	left := samples[minimumIndex-1].SeparationArcsecs
	center := samples[minimumIndex].SeparationArcsecs
	right := samples[minimumIndex+1].SeparationArcsecs
	curvature := left - 2*center + right
	if curvature <= 0 {
		t.Fatalf("invalid minimum curvature: %.12f", curvature)
	}
	offsetMinutes := 0.5 * (left - right) / curvature
	minimum := center - (left-right)*(left-right)/(8*curvature)
	centerJDE := mutualCurveSampleJD(t, samples[minimumIndex].TT)
	return centerJDE + offsetMinutes/1440, minimum
}

func mutualCurveSampleJD(t *testing.T, value string) float64 {
	t.Helper()
	instant, err := time.Parse("2006-01-02T15:04:05", value)
	if err != nil {
		t.Fatalf("parse curve time %q: %v", value, err)
	}
	return Date2JD(instant)
}

func collectMutualEventsBySearch(start, end float64, first, second int) []JupiterGalileanMutualEvent {
	var events []JupiterGalileanMutualEvent
	seed := start - 12.0/24
	for i := 0; i < 512; i++ {
		event := NextJupiterGalileanMutualEvent(seed, first, second, JupiterGalileanMutualOccultation)
		if !event.Valid || event.Greatest >= end {
			break
		}
		if event.Greatest >= start {
			events = append(events, event)
		}
		seed = event.End + 0.1/86400
	}
	return events
}

func collectMutualEventsByFiveMinutes(start, end float64, first, second int) []JupiterGalileanMutualEvent {
	step := 5.0 / 1440
	previousTime := start
	previousMetric, previousOK := jupiterGalileanMutualMetricAt(previousTime, first, second, nil)
	currentTime := previousTime + step
	currentMetric, currentOK := jupiterGalileanMutualMetricAt(currentTime, first, second, nil)
	var events []JupiterGalileanMutualEvent
	for currentTime+step < end {
		nextTime := currentTime + step
		nextMetric, nextOK := jupiterGalileanMutualMetricAt(nextTime, first, second, nil)
		if previousOK && currentOK && nextOK && currentMetric <= previousMetric && currentMetric <= nextMetric {
			greatest := refineJupiterGalileanMutualMinimum(previousTime, nextTime, first, second, nil)
			if event, ok := findJupiterGalileanMutualEventAround(greatest, first, second, JupiterGalileanMutualOccultation, nil); ok && event.Greatest >= start && event.Greatest < end {
				if len(events) == 0 || event.Greatest-events[len(events)-1].Greatest > 120.0/86400 {
					events = append(events, event)
				}
			}
		}
		previousTime, previousMetric, previousOK = currentTime, currentMetric, currentOK
		currentTime, currentMetric, currentOK = nextTime, nextMetric, nextOK
	}
	return events
}

func testMutualPairName(first, second int) string {
	return string(rune('0'+first)) + "-" + string(rune('0'+second))
}

func assertSameJupiterGalileanMutualEvent(t *testing.T, got, want JupiterGalileanMutualEvent) {
	assertSameJupiterGalileanMutualEventWithin(t, got, want, 0)
}

func assertSameJupiterGalileanMutualEventWithin(t *testing.T, got, want JupiterGalileanMutualEvent, toleranceSeconds float64) {
	t.Helper()
	if got.Valid != want.Valid || got.Type != want.Type || got.Active != want.Active || got.Passive != want.Passive ||
		!sameMutualFloatWithin(got.Start, want.Start, toleranceSeconds) || !sameMutualFloatWithin(got.Greatest, want.Greatest, toleranceSeconds) || !sameMutualFloatWithin(got.End, want.End, toleranceSeconds) ||
		!sameMutualFloatWithin(got.DurationDays, want.DurationDays, toleranceSeconds) || !sameMutualFloatWithin(got.MinimumSeparationArcsec, want.MinimumSeparationArcsec, toleranceSeconds) ||
		!sameMutualFloatWithin(got.ImpactParameter, want.ImpactParameter, toleranceSeconds) || !sameMutualFloatWithin(got.Coverage, want.Coverage, toleranceSeconds) || got.Annular != want.Annular {
		t.Fatalf("mutual event mismatch: got=%+v want=%+v", got, want)
	}
	for index := range got.GreatestState {
		gotState := got.GreatestState[index]
		wantState := want.GreatestState[index]
		if gotState.Transit != wantState.Transit || gotState.Occultation != wantState.Occultation || gotState.Eclipse != wantState.Eclipse || gotState.ShadowTransit != wantState.ShadowTransit ||
			!sameMutualFloatWithin(gotState.ShadowOffsetXArcsec, wantState.ShadowOffsetXArcsec, toleranceSeconds) || !sameMutualFloatWithin(gotState.ShadowOffsetYArcsec, wantState.ShadowOffsetYArcsec, toleranceSeconds) ||
			!sameMutualFloatWithin(gotState.ShadowOffsetXJupiterRadii, wantState.ShadowOffsetXJupiterRadii, toleranceSeconds) || !sameMutualFloatWithin(gotState.ShadowOffsetYJupiterRadii, wantState.ShadowOffsetYJupiterRadii, toleranceSeconds) {
			t.Fatalf("mutual greatest state mismatch at %d: got=%+v want=%+v", index, gotState, wantState)
		}
	}
}

func assertSameJupiterGalileanMutualTimelineWithin(t *testing.T, got, want JupiterGalileanMutualEvent, toleranceSeconds float64) {
	t.Helper()
	if got.Valid != want.Valid || got.Type != want.Type || got.Active != want.Active || got.Passive != want.Passive ||
		!sameMutualFloatWithin(got.Start, want.Start, toleranceSeconds) || !sameMutualFloatWithin(got.Greatest, want.Greatest, toleranceSeconds) || !sameMutualFloatWithin(got.End, want.End, toleranceSeconds) {
		t.Fatalf("mutual event timeline mismatch: got=%+v want=%+v", got, want)
	}
	for index := range got.GreatestState {
		gotState := got.GreatestState[index]
		wantState := want.GreatestState[index]
		if gotState.Transit != wantState.Transit || gotState.Occultation != wantState.Occultation || gotState.Eclipse != wantState.Eclipse || gotState.ShadowTransit != wantState.ShadowTransit {
			t.Fatalf("mutual greatest state mismatch at %d: got=%+v want=%+v", index, gotState, wantState)
		}
	}
}

func sameMutualFloat(got, want float64) bool {
	return got == want || (math.IsNaN(got) && math.IsNaN(want))
}

func sameMutualFloatWithin(got, want, toleranceSeconds float64) bool {
	if sameMutualFloat(got, want) {
		return true
	}
	return toleranceSeconds > 0 && math.Abs(got-want)*86400 <= toleranceSeconds
}
