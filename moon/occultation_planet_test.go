package moon

import (
	"errors"
	"math"
	"testing"
	"time"
)

func TestFindPlanetOccultationsReturnsFiniteDiskContacts(t *testing.T) {
	location := time.FixedZone("CST", 8*3600)
	observer := Observer{Longitude: -30.072, Latitude: 16.21}
	start := time.Date(2024, time.August, 21, 9, 30, 0, 0, location)
	end := time.Date(2024, time.August, 21, 12, 0, 0, 0, location)
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
	if result.Planet != OccultationSaturn || result.TargetID != "Saturn" {
		t.Fatalf("target = %v/%q, want Saturn", result.Planet, result.TargetID)
	}
	if result.Observer != observer {
		t.Fatalf("observer = %+v, want %+v", result.Observer, observer)
	}
	if result.Type != OccultationTotal || !result.HasInternalContacts || !result.ContactsComplete {
		t.Fatalf("unexpected event geometry: type=%q internal=%v complete=%v", result.Type, result.HasInternalContacts, result.ContactsComplete)
	}
	if result.Greatest.Location() != location || result.ExternalImmersion.Location() != location || result.ExternalEmersion.Location() != location {
		t.Fatalf("result times did not preserve the start location: %+v", result)
	}
	wantGreatest := time.Date(2024, time.August, 21, 10, 49, 10, 0, location)
	if delta := math.Abs(result.Greatest.Sub(wantGreatest).Seconds()); delta > 2 {
		t.Fatalf("greatest differs from regression baseline by %.3fs: got %v want %v", delta, result.Greatest, wantGreatest)
	}
	if result.MinimumSeparationArcsec > result.MoonSemidiameterArcsec-result.PlanetSemidiameterArcsec {
		t.Fatalf("total event does not fully cover the planet: separation=%.6f Moon=%.6f planet=%.6f",
			result.MinimumSeparationArcsec, result.MoonSemidiameterArcsec, result.PlanetSemidiameterArcsec)
	}
}

func TestFindBestPlanetOccultationsReturnsVisibleGroundPoint(t *testing.T) {
	start := time.Date(2024, time.August, 20, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, time.August, 22, 0, 0, 0, 0, time.UTC)
	results, err := FindBestPlanetOccultations(start, end, OccultationSaturn, OccultationSearchOptions{MaxEvents: 1})
	if err != nil {
		t.Fatalf("FindBestPlanetOccultations() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("FindBestPlanetOccultations() returned %d events, want 1", len(results))
	}
	result := results[0]
	if result.Observer.Longitude < -180 || result.Observer.Longitude > 180 || result.Observer.Latitude < -90 || result.Observer.Latitude > 90 {
		t.Fatalf("best observer out of bounds: %+v", result.Observer)
	}
	if result.Observer.Height != 0 {
		t.Fatalf("best observer height = %.1f, want sea level", result.Observer.Height)
	}
	if !result.VisibleAtGreatest || result.MoonAltitudeAtGreatest < 0 {
		t.Fatalf("best event is not visible: altitude=%.6f visible=%v", result.MoonAltitudeAtGreatest, result.VisibleAtGreatest)
	}
	if !result.HasInternalContacts || !result.ContactsComplete {
		t.Fatalf("best event contacts are incomplete: %+v", result)
	}
}

func TestFindPlanetOccultationsReturnsPartialDiskEventWithoutInternalContacts(t *testing.T) {
	start := time.Date(2024, time.August, 21, 1, 30, 0, 0, time.UTC)
	end := time.Date(2024, time.August, 21, 4, 0, 0, 0, time.UTC)
	// 使用站点到月球的距离计算月球视半径后，-6.5 度仍在内接触外侧 / -6.5 degrees remains just outside internal contact after using the
	// 使用站点到月球的距离计算月球视半径 / station-to-Moon distance for the lunar semidiameter.
	results, err := FindPlanetOccultations(start, end, OccultationSaturn, -30.072, -6.5, 0, OccultationSearchOptions{})
	if err != nil {
		t.Fatalf("FindPlanetOccultations() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("FindPlanetOccultations() returned %d events, want 1", len(results))
	}
	result := results[0]
	if result.Type != OccultationPartial || result.HasInternalContacts {
		t.Fatalf("event geometry = %q internal=%v, want partial without internal contacts", result.Type, result.HasInternalContacts)
	}
	if result.InternalImmersion != (time.Time{}) || result.InternalEmersion != (time.Time{}) {
		t.Fatalf("partial event returned internal contacts: C2=%v C3=%v", result.InternalImmersion, result.InternalEmersion)
	}
	if !result.ContactsComplete || result.ExternalImmersion.IsZero() || result.ExternalEmersion.IsZero() {
		t.Fatalf("partial event external contacts are incomplete: %+v", result)
	}
	if !(result.ExternalImmersion.Before(result.Greatest) && result.Greatest.Before(result.ExternalEmersion)) {
		t.Fatalf("partial event contact order is invalid: %+v", result)
	}
}

func TestFindPlanetOccultationsValidatesInputs(t *testing.T) {
	start := time.Date(2024, time.August, 21, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	tests := []struct {
		name      string
		start     time.Time
		end       time.Time
		planet    OccultationPlanet
		longitude float64
		options   OccultationSearchOptions
	}{
		{name: "zero time", planet: OccultationSaturn},
		{name: "unordered time", start: start, end: start, planet: OccultationSaturn},
		{name: "unsupported planet", start: start, end: end, planet: OccultationPlanet("pluto")},
		{name: "invalid observer", start: start, end: end, planet: OccultationSaturn, longitude: math.NaN()},
		{name: "invalid options", start: start, end: end, planet: OccultationSaturn, options: OccultationSearchOptions{MaxEvents: -1}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := FindPlanetOccultations(test.start, test.end, test.planet, test.longitude, 0, 0, test.options)
			if !errors.Is(err, ErrInvalidOccultationInput) {
				t.Fatalf("FindPlanetOccultations() error = %v, want ErrInvalidOccultationInput", err)
			}
		})
	}
}

func TestFindBestPlanetOccultationsKeepsNonCentralEvents(t *testing.T) {
	// 2025-01-05 月掩海王星与 2025-07-28 月掩火星都是月影轴不与地球椭球相交的非中心
	// 事件：最佳站心曾取外接触切点，该点位于掩带边缘、落在掩食之外，整场被事件搜索
	// 丢弃。事件搜索与路径搜索必须报告同一场事件，掩甚站心必须落在掩食之内。
	// The 2025-01-05 Neptune and 2025-07-28 Mars occultations are non-central events
	// whose shadow axis misses the ellipsoid. The best station used to be taken at the
	// outer contact tangent, which sits on the band edge and outside the occultation, so
	// the whole event was dropped from the search. The event and path searches must
	// report the same event with a greatest station inside the occultation.
	tests := []struct {
		name   string
		planet OccultationPlanet
		day    time.Time
	}{
		{name: "2025-01-05 Neptune", planet: OccultationNeptune, day: time.Date(2025, time.January, 5, 0, 0, 0, 0, time.UTC)},
		{name: "2025-07-28 Mars", planet: OccultationMars, day: time.Date(2025, time.July, 28, 0, 0, 0, 0, time.UTC)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			start := test.day.Add(-24 * time.Hour)
			end := test.day.Add(24 * time.Hour)
			events, err := FindBestPlanetOccultations(start, end, test.planet, OccultationSearchOptions{})
			if err != nil {
				t.Fatalf("FindBestPlanetOccultations() error = %v", err)
			}
			paths, err := FindPlanetOccultationPaths(start, end, test.planet, OccultationPathOptions{})
			if err != nil {
				t.Fatalf("FindPlanetOccultationPaths() error = %v", err)
			}
			if len(events) != 1 || len(paths) != 1 {
				t.Fatalf("event search returned %d events, path search %d, want one each", len(events), len(paths))
			}
			event := events[0]
			limit := event.MoonSemidiameterArcsec + event.PlanetSemidiameterArcsec
			if event.MinimumSeparationArcsec > limit {
				t.Fatalf("greatest station lies outside the occultation: separation=%.3f limit=%.3f",
					event.MinimumSeparationArcsec, limit)
			}
			if delta := math.Abs(event.Greatest.Sub(paths[0].Greatest.Time).Seconds()); delta > 60 {
				t.Fatalf("greatest differs from the path search by %.1fs", delta)
			}
		})
	}
}
