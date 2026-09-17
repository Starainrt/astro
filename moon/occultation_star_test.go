package moon

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

func TestFindStarOccultationsFindsPointSourceEvent(t *testing.T) {
	observer := Observer{Longitude: 121.4737, Latitude: 31.2304, Height: 4}
	eventTime := time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)
	tt := occultationTestTimeToTT(eventTime)
	ra, dec := occultationTestMoonTopocentricRaDec(tt, observer, -1)
	star := StarCoordinate{
		ID:    "synthetic-point-source",
		RA:    ra,
		Dec:   dec,
		Epoch: eventTime,
		Frame: CoordinateFrameApparentOfDate,
	}

	start := eventTime.In(time.FixedZone("CST", 8*3600)).Add(-3 * time.Hour)
	end := eventTime.In(time.FixedZone("CST", 8*3600)).Add(3 * time.Hour)
	results, err := FindStarOccultations(start, end, star,
		observer.Longitude, observer.Latitude, observer.Height, OccultationSearchOptions{MaxEvents: 1})
	if err != nil {
		t.Fatalf("FindStarOccultations() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("FindStarOccultations() returned %d events, want 1", len(results))
	}
	result := results[0]
	if result.Observer != observer {
		t.Fatalf("observer = %+v, want %+v", result.Observer, observer)
	}
	if result.Greatest.Location().String() != "CST" {
		t.Fatalf("greatest location = %q, want CST", result.Greatest.Location().String())
	}
	if result.Type != OccultationTotal {
		t.Fatalf("event type = %q, want %q", result.Type, OccultationTotal)
	}
	if result.MinimumSeparationArcsec > result.MoonSemidiameterArcsec {
		t.Fatalf("minimum separation %.3f exceeds lunar radius %.3f", result.MinimumSeparationArcsec, result.MoonSemidiameterArcsec)
	}
	if result.Immersion.IsZero() || result.Emersion.IsZero() {
		t.Fatalf("contact times must be populated: immersion=%v emersion=%v", result.Immersion, result.Emersion)
	}
	if !result.ContactsComplete {
		t.Fatal("contacts must be marked complete")
	}
	if !(result.Immersion.Before(result.Greatest) && result.Greatest.Before(result.Emersion)) {
		t.Fatalf("contact ordering invalid: immersion=%v greatest=%v emersion=%v", result.Immersion, result.Greatest, result.Emersion)
	}
	if delta := math.Abs(result.Greatest.Sub(eventTime).Seconds()); delta > 60 {
		t.Fatalf("greatest differs from synthetic event by %.1fs", delta)
	}
	if result.MoonAzimuthAtGreatest < 0 || result.MoonAzimuthAtGreatest >= 360 {
		t.Fatalf("moon azimuth = %.6f, want [0,360)", result.MoonAzimuthAtGreatest)
	}
	if result.PositionAngleDeg < 0 || result.PositionAngleDeg >= 360 {
		t.Fatalf("position angle = %.6f, want [0,360)", result.PositionAngleDeg)
	}
}

func TestFindBestStarOccultationsReturnsTopocentricGroundPoint(t *testing.T) {
	observer := Observer{Longitude: 121.4737, Latitude: 31.2304, Height: 4}
	eventTime := time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)
	ra, dec := occultationTestMoonTopocentricRaDec(occultationTestTimeToTT(eventTime), observer, -1)
	star := StarCoordinate{
		ID:    "synthetic-best-point-source",
		RA:    ra,
		Dec:   dec,
		Epoch: eventTime,
		Frame: CoordinateFrameApparentOfDate,
	}

	results, err := FindBestStarOccultations(
		eventTime.Add(-3*time.Hour), eventTime.Add(3*time.Hour), star, OccultationSearchOptions{MaxEvents: 1})
	if err != nil {
		t.Fatalf("FindBestStarOccultations() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("FindBestStarOccultations() returned %d events, want 1", len(results))
	}
	result := results[0]
	if result.Observer.Longitude < -180 || result.Observer.Longitude > 180 || result.Observer.Latitude < -90 || result.Observer.Latitude > 90 {
		t.Fatalf("best observer out of bounds: %+v", result.Observer)
	}
	if result.Observer.Height != 0 {
		t.Fatalf("best observer height = %.1f, want sea-level default", result.Observer.Height)
	}
	if result.MinimumSeparationArcsec > result.MoonSemidiameterArcsec {
		t.Fatalf("best point is not an occultation: separation=%.3f radius=%.3f", result.MinimumSeparationArcsec, result.MoonSemidiameterArcsec)
	}
	if !result.VisibleAtGreatest || result.MoonAltitudeAtGreatest < 0 {
		t.Fatalf("best observer must see the Moon: altitude=%.6f visible=%v", result.MoonAltitudeAtGreatest, result.VisibleAtGreatest)
	}
	if result.Immersion.IsZero() || result.Emersion.IsZero() || !result.Immersion.Before(result.Greatest) || !result.Greatest.Before(result.Emersion) {
		t.Fatalf("best point contact ordering invalid: immersion=%v greatest=%v emersion=%v", result.Immersion, result.Greatest, result.Emersion)
	}
}

func TestFindStarOccultationsLatitudePrefilterRejectsPolarStar(t *testing.T) {
	star := StarCoordinate{
		ID:    "polar-star",
		RA:    0,
		Dec:   89,
		Epoch: time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
		Frame: CoordinateFrameICRS,
	}
	results, err := FindStarOccultations(
		time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		star, 0, 0, 0, OccultationSearchOptions{})
	if err != nil {
		t.Fatalf("FindStarOccultations() error = %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("polar star returned %d events, want none", len(results))
	}
}

func TestFindStarOccultationsReturnsCompleteContactsOutsideQueryWindow(t *testing.T) {
	observer := Observer{Longitude: 121.4737, Latitude: 31.2304, Height: 4}
	eventTime := time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)
	ra, dec := occultationTestMoonTopocentricRaDec(occultationTestTimeToTT(eventTime), observer, -1)
	star := StarCoordinate{ID: "clipped-window", RA: ra, Dec: dec, Epoch: eventTime, Frame: CoordinateFrameApparentOfDate}
	start := eventTime.Add(-time.Minute)
	end := eventTime.Add(time.Minute)

	results, err := FindStarOccultations(start, end, star,
		observer.Longitude, observer.Latitude, observer.Height, OccultationSearchOptions{})
	if err != nil {
		t.Fatalf("FindStarOccultations() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("FindStarOccultations() returned %d events, want 1", len(results))
	}
	result := results[0]
	if !result.ContactsComplete || result.Immersion.IsZero() || result.Emersion.IsZero() {
		t.Fatalf("incomplete contacts: %+v", result)
	}
	if !result.Immersion.Before(start) || !result.Emersion.After(end) {
		t.Fatalf("contacts were clipped to query window: immersion=%v window=%v..%v emersion=%v", result.Immersion, start, end, result.Emersion)
	}
}

func occultationTestTimeToTT(value time.Time) float64 {
	return basic.TD2UT(basic.Date2JDE(value.UTC()), true)
}

func occultationTestMoonTopocentricRaDec(tt float64, observer Observer, n int) (float64, float64) {
	ra, dec := basic.HMoonGeocentricApparentRaDecN(tt, n)
	distanceAU := basic.HMoonAwayN(tt, n) / 149597870.7
	ra, dec = basic.TopocentricRaDec(ra, dec, observer.Latitude, observer.Longitude, basic.TD2UT(tt, false), distanceAU, observer.Height)
	ra = math.Mod(ra, 360)
	if ra < 0 {
		ra += 360
	}
	return ra, dec
}

func hr4799CoordinateForTest() StarCoordinate {
	return StarCoordinate{
		ID:                             "HR 4799",
		RA:                             189.1975,
		Dec:                            -5.831944444444,
		Epoch:                          time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
		Frame:                          CoordinateFrameJ2000,
		ProperMotionRACosDecMasPerYear: -28,
		ProperMotionDecMasPerYear:      -18,
	}
}

func TestFindStarOccultationsValidatesInputs(t *testing.T) {
	star := StarCoordinate{
		RA:    10,
		Dec:   20,
		Epoch: time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
		Frame: CoordinateFrameICRS,
	}
	_, err := FindStarOccultations(time.Time{}, time.Time{}, star, 0, 0, 0, OccultationSearchOptions{})
	if !errors.Is(err, ErrInvalidOccultationInput) {
		t.Fatalf("FindStarOccultations() error = %v, want ErrInvalidOccultationInput", err)
	}
	start := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)
	_, err = FindStarOccultations(start, start, star, 0, 0, 0, OccultationSearchOptions{})
	if !errors.Is(err, ErrInvalidOccultationInput) {
		t.Fatalf("FindStarOccultations() unordered range error = %v, want ErrInvalidOccultationInput", err)
	}
	_, err = FindStarOccultations(start, start.Add(time.Hour), star, math.NaN(), 0, 0, OccultationSearchOptions{})
	if !errors.Is(err, ErrInvalidOccultationInput) {
		t.Fatalf("FindStarOccultations() observer error = %v, want ErrInvalidOccultationInput", err)
	}
}

func TestFindBestStarOccultationsKeepsNonCentralEvents(t *testing.T) {
	// HR 1118 的黄纬为 +5.58°，接近月掩纬度极限，因此 2025-02-06 与 2025-05-26 的月掩
	// 都是月影轴不与地球椭球相交的非中心事件；最佳站心曾取外接触切点而整场丢失。
	// HR 1118 sits at ecliptic latitude +5.58°, close to the occultation latitude limit,
	// so its 2025-02-06 and 2025-05-26 events are non-central: the shadow axis misses the
	// ellipsoid and the best station used to be taken at the outer contact tangent, which
	// dropped the whole event from the search.
	star := StarCoordinate{
		ID:                             "HR 1118",
		RA:                             55.19291667,
		Dec:                            25.32944444,
		Epoch:                          time.Date(2000, time.January, 1, 12, 0, 0, 0, time.UTC),
		Frame:                          CoordinateFrameJ2000,
		ProperMotionRACosDecMasPerYear: 8,
		ProperMotionDecMasPerYear:      -15,
	}
	for _, day := range []time.Time{
		time.Date(2025, time.February, 6, 0, 0, 0, 0, time.UTC),
		time.Date(2025, time.May, 26, 0, 0, 0, 0, time.UTC),
	} {
		t.Run(day.Format("2006-01-02"), func(t *testing.T) {
			start := day.Add(-24 * time.Hour)
			end := day.Add(24 * time.Hour)
			events, err := FindBestStarOccultations(start, end, star, OccultationSearchOptions{})
			if err != nil {
				t.Fatalf("FindBestStarOccultations() error = %v", err)
			}
			paths, err := FindStarOccultationPaths(start, end, star, OccultationPathOptions{})
			if err != nil {
				t.Fatalf("FindStarOccultationPaths() error = %v", err)
			}
			if len(events) != 1 || len(paths) != 1 {
				t.Fatalf("event search returned %d events, path search %d, want one each", len(events), len(paths))
			}
			event := events[0]
			if event.MinimumSeparationArcsec > event.MoonSemidiameterArcsec {
				t.Fatalf("greatest station lies outside the occultation: separation=%.3f moon=%.3f",
					event.MinimumSeparationArcsec, event.MoonSemidiameterArcsec)
			}
		})
	}
}
