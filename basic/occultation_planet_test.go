package basic

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/internal/geodata"
)

func TestPlanetOccultationSupportsAllPlanetTargets(t *testing.T) {
	tt := UTC2TT(Date2JD(time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)))
	tests := []struct {
		planet OccultationPlanet
		name   string
	}{
		{OccultationMercury, "Mercury"},
		{OccultationVenus, "Venus"},
		{OccultationMars, "Mars"},
		{OccultationJupiter, "Jupiter"},
		{OccultationSaturn, "Saturn"},
		{OccultationUranus, "Uranus"},
		{OccultationNeptune, "Neptune"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.planet.Validate(); err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
			if test.planet.String() != test.name {
				t.Fatalf("String() = %q, want %q", test.planet.String(), test.name)
			}
			config, ok := planetOccultationConfigFor(test.planet)
			if !ok {
				t.Fatal("planet occultation config is unavailable")
			}
			state := planetOccultationStateAt(tt, config, nil, -1)
			if !state.valid || state.planetSemidiameter <= 0 || state.moonSemidiameter <= state.planetSemidiameter {
				t.Fatalf("invalid planet state: %+v", state)
			}
		})
	}
}

func TestPlanetOccultationSaturnContactsSolveDynamicDiskMetrics(t *testing.T) {
	observer := Observer{Longitude: -30.072, Latitude: 16.21}
	start := time.Date(2024, time.August, 21, 1, 30, 0, 0, time.UTC)
	end := time.Date(2024, time.August, 21, 4, 0, 0, 0, time.UTC)
	results, err := FindPlanetOccultations(
		start, end, OccultationSaturn,
		observer.Longitude, observer.Latitude, observer.Height,
		OccultationSearchOptions{},
	)
	if err != nil {
		t.Fatalf("FindPlanetOccultations() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("FindPlanetOccultations() returned %d events, want 1", len(results))
	}
	result := results[0]
	if result.Type != OccultationTotal || !result.HasInternalContacts || !result.ContactsComplete {
		t.Fatalf("unexpected event geometry: type=%q internal=%v complete=%v", result.Type, result.HasInternalContacts, result.ContactsComplete)
	}

	config, ok := planetOccultationConfigFor(OccultationSaturn)
	if !ok {
		t.Fatal("Saturn occultation config is unavailable")
	}
	contacts := []struct {
		name     string
		value    time.Time
		internal bool
	}{
		{"C1", result.ExternalImmersion, false},
		{"C2", result.InternalImmersion, true},
		{"C3", result.InternalEmersion, true},
		{"C4", result.ExternalEmersion, false},
	}
	for _, contact := range contacts {
		state := planetOccultationStateAt(occultationTimeToTT(contact.value), config, &observer, -1)
		metric := state.externalContactMetric
		if contact.internal {
			metric = state.internalContactMetric
		}
		if math.Abs(metric) > 0.1 {
			t.Errorf("%s contact residual = %.6f arcsec, want <= 0.1", contact.name, metric)
		}
	}

	if !(result.ExternalImmersion.Before(result.InternalImmersion) &&
		result.InternalImmersion.Before(result.Greatest) &&
		result.Greatest.Before(result.InternalEmersion) &&
		result.InternalEmersion.Before(result.ExternalEmersion)) {
		t.Fatalf("contact order is invalid: %+v", result)
	}
	if result.PlanetSemidiameterArcsec <= 0 || result.MoonSemidiameterArcsec <= result.PlanetSemidiameterArcsec {
		t.Fatalf("invalid dynamic semidiameters: Moon=%.6f planet=%.6f", result.MoonSemidiameterArcsec, result.PlanetSemidiameterArcsec)
	}
}

func TestPlanetOccultationUsesStationMoonDistanceForRadius(t *testing.T) {
	tt := occultationTimeToTT(time.Date(2024, time.August, 21, 2, 49, 10, 0, time.UTC))
	config, ok := planetOccultationConfigFor(OccultationSaturn)
	if !ok {
		t.Fatal("Saturn occultation config is unavailable")
	}
	moonRA, _ := HMoonGeocentricApparentRaDecN(tt, -1)
	subMoonLongitude := geodata.NormalizeLongitude(moonRA - ApparentSiderealTime(TT2UTC(tt))*15)
	near := Observer{Longitude: subMoonLongitude, Latitude: 0}
	far := Observer{Longitude: geodata.NormalizeLongitude(subMoonLongitude + 180), Latitude: 0}
	nearState := planetOccultationStateAt(tt, config, &near, -1)
	farState := planetOccultationStateAt(tt, config, &far, -1)
	if !nearState.valid || !farState.valid {
		t.Fatalf("invalid station states: near=%+v far=%+v", nearState, farState)
	}
	if nearState.moonSemidiameter <= farState.moonSemidiameter {
		t.Fatalf("station Moon radius did not follow station distance: near=%.6f far=%.6f", nearState.moonSemidiameter, farState.moonSemidiameter)
	}
	if nearState.moonSemidiameter-MoonSemidiameterN(tt, -1) <= 0 ||
		farState.moonSemidiameter-MoonSemidiameterN(tt, -1) >= 0 {
		t.Fatalf("station radius does not straddle geocentric radius: near=%.6f geo=%.6f far=%.6f",
			nearState.moonSemidiameter, MoonSemidiameterN(tt, -1), farState.moonSemidiameter)
	}
}

func TestPlanetOccultationUsesStationPlanetDistanceForRadius(t *testing.T) {
	tt := occultationTimeToTT(time.Date(2024, time.March, 11, 1, 0, 0, 0, time.UTC))
	config, ok := planetOccultationConfigFor(OccultationMercury)
	if !ok {
		t.Fatal("Mercury occultation config is unavailable")
	}
	planetRA, _ := config.apparentRaDecN(tt, -1)
	subPlanetLongitude := geodata.NormalizeLongitude(planetRA - ApparentSiderealTime(TT2UTC(tt))*15)
	near := Observer{Longitude: subPlanetLongitude, Latitude: 0}
	far := Observer{Longitude: geodata.NormalizeLongitude(subPlanetLongitude + 180), Latitude: 0}
	nearState := planetOccultationStateAt(tt, config, &near, -1)
	farState := planetOccultationStateAt(tt, config, &far, -1)
	if !nearState.valid || !farState.valid {
		t.Fatalf("invalid station states: near=%+v far=%+v", nearState, farState)
	}
	if nearState.planetSemidiameter <= farState.planetSemidiameter {
		t.Fatalf("station planet radius did not follow station distance: near=%.12f far=%.12f",
			nearState.planetSemidiameter, farState.planetSemidiameter)
	}
	geocentric := config.semidiameterN(tt, -1)
	if nearState.planetSemidiameter <= geocentric || farState.planetSemidiameter >= geocentric {
		t.Fatalf("station planet radius does not straddle geocentric radius: near=%.12f geo=%.12f far=%.12f",
			nearState.planetSemidiameter, geocentric, farState.planetSemidiameter)
	}
}

func TestPlanetOccultationBestObserverRefinesDynamicMetric(t *testing.T) {
	config, ok := planetOccultationConfigFor(OccultationSaturn)
	if !ok {
		t.Fatal("Saturn occultation config is unavailable")
	}
	startTT := occultationTimeToTT(time.Date(2024, time.August, 21, 1, 30, 0, 0, time.UTC))
	endTT := occultationTimeToTT(time.Date(2024, time.August, 21, 4, 0, 0, 0, time.UTC))
	seedTT := occultationTimeToTT(time.Date(2024, time.August, 21, 2, 49, 10, 0, time.UTC))
	bestTT, observer, _, bestOK := planetOccultationBestObserver(seedTT, startTT, endTT, config)
	if !bestOK {
		t.Fatal("best-observer search failed")
	}
	metric := planetOccultationExternalContactMetric(bestTT, config, &observer, -1)
	for _, deltaSeconds := range []float64{-0.1, 0.1} {
		neighbor := planetOccultationExternalContactMetric(bestTT+deltaSeconds/86400, config, &observer, -1)
		if metric > neighbor+1e-6 {
			t.Fatalf("best time does not minimize the dynamic metric: center=%.12f neighbor(%+.1fs)=%.12f",
				metric, deltaSeconds, neighbor)
		}
	}
}

func TestPlanetOccultationInnerPlanetContactsUseDynamicTargets(t *testing.T) {
	tests := []struct {
		name   string
		planet OccultationPlanet
		start  time.Time
		end    time.Time
	}{
		{"Mercury", OccultationMercury, time.Date(2024, time.March, 10, 0, 0, 0, 0, time.UTC), time.Date(2024, time.March, 12, 0, 0, 0, 0, time.UTC)},
		{"Venus", OccultationVenus, time.Date(2024, time.April, 6, 0, 0, 0, 0, time.UTC), time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			results, err := FindBestPlanetOccultations(test.start, test.end, test.planet, OccultationSearchOptions{MaxEvents: 1})
			if err != nil {
				t.Fatalf("FindBestPlanetOccultations() error = %v", err)
			}
			if len(results) != 1 {
				t.Fatalf("FindBestPlanetOccultations() returned %d events, want 1", len(results))
			}
			result := results[0]
			if result.Planet != test.planet || !result.HasInternalContacts || !result.ContactsComplete {
				t.Fatalf("unexpected event geometry: %+v", result)
			}
			if !(result.ExternalImmersion.Before(result.InternalImmersion) &&
				result.InternalImmersion.Before(result.Greatest) &&
				result.Greatest.Before(result.InternalEmersion) &&
				result.InternalEmersion.Before(result.ExternalEmersion)) {
				t.Fatalf("contact order is invalid: %+v", result)
			}

			config, ok := planetOccultationConfigFor(test.planet)
			if !ok {
				t.Fatalf("%s occultation config is unavailable", test.name)
			}
			contacts := []struct {
				value    time.Time
				internal bool
			}{
				{result.ExternalImmersion, false},
				{result.InternalImmersion, true},
				{result.InternalEmersion, true},
				{result.ExternalEmersion, false},
			}
			for index, contact := range contacts {
				state := planetOccultationStateAt(occultationTimeToTT(contact.value), config, &result.Observer, -1)
				metric := state.externalContactMetric
				if contact.internal {
					metric = state.internalContactMetric
				}
				if math.Abs(metric) > 0.1 {
					t.Errorf("contact %d residual = %.6f arcsec, want <= 0.1", index+1, metric)
				}
			}
		})
	}
}

func TestPlanetOccultationOuterPlanetContactsUseDynamicTargets(t *testing.T) {
	tests := []struct {
		name     string
		planet   OccultationPlanet
		start    time.Time
		end      time.Time
		observer Observer
	}{
		{"Mars", OccultationMars, time.Date(2020, 2, 18, 11, 0, 0, 0, time.UTC), time.Date(2020, 2, 18, 16, 0, 0, 0, time.UTC), Observer{Longitude: -76.01, Latitude: 29.918}},
		{"Jupiter", OccultationJupiter, time.Date(2020, 1, 23, 0, 0, 0, 0, time.UTC), time.Date(2020, 1, 23, 5, 0, 0, 0, time.UTC), Observer{Longitude: 120.15, Latitude: -45.552}},
		{"Uranus", OccultationUranus, time.Date(2022, 2, 7, 19, 0, 0, 0, time.UTC), time.Date(2022, 2, 7, 22, 0, 0, 0, time.UTC), Observer{Longitude: 11.186, Latitude: -62.948}},
		{"Neptune", OccultationNeptune, time.Date(2023, 9, 1, 7, 0, 0, 0, time.UTC), time.Date(2023, 9, 1, 10, 0, 0, 0, time.UTC), Observer{Longitude: -13.896, Latitude: -62.038}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			results, err := FindPlanetOccultations(
				test.start, test.end, test.planet,
				test.observer.Longitude, test.observer.Latitude, test.observer.Height,
				OccultationSearchOptions{},
			)
			if err != nil {
				t.Fatalf("FindPlanetOccultations() error = %v", err)
			}
			if len(results) != 1 {
				t.Fatalf("FindPlanetOccultations() returned %d events, want 1", len(results))
			}
			result := results[0]
			if result.Type != OccultationTotal || !result.HasInternalContacts || !result.ContactsComplete {
				t.Fatalf("unexpected event geometry: %+v", result)
			}
			config, ok := planetOccultationConfigFor(test.planet)
			if !ok {
				t.Fatalf("%s occultation config is unavailable", test.name)
			}
			contacts := []struct {
				value    time.Time
				internal bool
			}{
				{result.ExternalImmersion, false},
				{result.InternalImmersion, true},
				{result.InternalEmersion, true},
				{result.ExternalEmersion, false},
			}
			for index, contact := range contacts {
				state := planetOccultationStateAt(occultationTimeToTT(contact.value), config, &test.observer, -1)
				metric := state.externalContactMetric
				if contact.internal {
					metric = state.internalContactMetric
				}
				if math.Abs(metric) > 0.1 {
					t.Errorf("contact %d residual = %.6f arcsec, want <= 0.1", index+1, metric)
				}
			}
		})
	}
}
