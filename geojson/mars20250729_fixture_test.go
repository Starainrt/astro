package geojson_test

import (
	"sync"
	"testing"
	"time"

	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/moon"
)

type mars20250729FixtureKey struct {
	riseSetStep     time.Duration
	includeTimeline bool
}

type mars20250729Fixture struct {
	path       moon.PlanetOccultationPath
	collection decodedCollection
}

var mars20250729Fixtures = struct {
	sync.Mutex
	values map[mars20250729FixtureKey]*mars20250729Fixture
}{values: make(map[mars20250729FixtureKey]*mars20250729Fixture)}

// mars20250729TestFixture computes each distinct Mars regression scenario once.
// The returned path and decoded collection are read-only test fixtures; caching
// them avoids repeating the expensive ephemeris and GeoJSON setup in related
// assertions without sharing mutable production state.
func mars20250729TestFixture(
	t *testing.T,
	riseSetStep time.Duration,
	includeTimeline bool,
) *mars20250729Fixture {
	t.Helper()
	key := mars20250729FixtureKey{
		riseSetStep:     riseSetStep,
		includeTimeline: includeTimeline,
	}

	mars20250729Fixtures.Lock()
	defer mars20250729Fixtures.Unlock()
	if fixture := mars20250729Fixtures.values[key]; fixture != nil {
		return fixture
	}

	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2025, time.July, 29, 0, 0, 0, 0, zone)
	options := moon.OccultationPathOptions{
		Step:              20 * time.Minute,
		TargetSpacingKM:   900,
		RiseSetStep:       riseSetStep,
		DisableFootprints: true,
	}
	if includeTimeline {
		options.IncludeFootprintTimeline = true
		options.FootprintTimelineStep = 5 * time.Minute
	}
	paths, err := moon.FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), moon.OccultationMars, options,
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	data, err := geojson.MarshalPlanetOccultation(paths[0])
	if err != nil {
		t.Fatalf("MarshalPlanetOccultation: %v", err)
	}
	fixture := &mars20250729Fixture{
		path:       paths[0],
		collection: decodeCollection(t, data),
	}
	mars20250729Fixtures.values[key] = fixture
	return fixture
}
