package astro

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"math"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/ephem"
	"github.com/starainrt/astro/satellite"
)

func TestContextSatelliteTwilightConsistency(t *testing.T) {
	previousModel, previousInstance := GetModel(), ephem.Global()
	t.Cleanup(func() {
		SetModel(previousModel)
		ephem.SetGlobal(previousInstance)
	})
	requireSatelliteResult(t, SetModel(ModelSemianalytic))
	context, err := NewContext(ModelSemianalytic)
	requireSatelliteResult(t, err)
	explicit := satellite.NewWithInstance(ephem.NewModelInstance(ephem.Semianalytic))
	data, err := os.ReadFile("satellite/testdata/skyfield_coverage.json")
	requireSatelliteResult(t, err)
	var fixture struct {
		Groups []struct{ Elements satellite.Elements }
	}
	requireSatelliteResult(t, json.Unmarshal(data, &fixture))
	elements := fixture.Groups[4].Elements
	var sweeps []satellite.CoverageSweep
	for _, test := range []struct {
		hour      int
		threshold *float64
	}{{14, nil}, {14, satelliteTwilight(30)}, {0, satelliteTwilight(89.999)}, {14, satelliteTwilight(90)}, {14, satelliteTwilight(-90)}} {
		at := time.Date(2026, 11, 14, test.hour, 0, 0, 0, time.UTC)
		until := at.Add(10 * time.Second)
		options := satellite.CoverageOptions{TwilightSunAltitude: test.threshold}
		want, err := satellite.CoverageBetween(elements, at, until, 10*time.Second, options)
		requireSatelliteResult(t, err)
		for _, module := range []*satellite.Satellite{explicit, context.Satellite()} {
			got, err := module.CoverageBetween(elements, at, until, 10*time.Second, options)
			requireSatelliteResult(t, err)
			assertSatelliteBits(t, "CoverageBetween", want, got)
			instant, err := module.CoverageAt(elements, at, options)
			requireSatelliteResult(t, err)
			assertSatelliteBits(t, "CoverageAt", want.Instants[0], instant)
		}
		sweeps = append(sweeps, want)
	}
	encoded, err := json.Marshal(sweeps)
	requireSatelliteResult(t, err)
	t.Logf("coverage SHA-256: %x", sha256.Sum256(encoded))
}

func satelliteTwilight(value float64) *float64 { return &value }

func TestContextSatelliteConsistency(t *testing.T) {
	previousModel, previousInstance := GetModel(), ephem.Global()
	t.Cleanup(func() {
		SetModel(previousModel)
		ephem.SetGlobal(previousInstance)
	})
	if err := SetModel(ModelSemianalytic); err != nil {
		t.Fatal(err)
	}
	context, err := NewContext(ModelSemianalytic)
	if err != nil {
		t.Fatal(err)
	}
	inst := ephem.NewModelInstance(ephem.Semianalytic)
	explicit := satellite.NewWithInstance(inst)
	module := context.Satellite()
	if module == nil || module != context.Satellite() || module == context.WithLunarLimb(nil).Satellite() {
		t.Fatal("Context must lazily cache its own satellite module")
	}

	data, err := os.ReadFile("satellite/testdata/skyfield_transits.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures struct {
		Groups []struct {
			Label      string
			Elements   satellite.Elements
			Observer   satellite.Observer
			Start, End time.Time
			Target     satellite.TransitTarget
			Rows       []json.RawMessage
		}
	}
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}

	var maximumMoonDifference float64
	var transitCases int
	for index, group := range fixtures.Groups {
		if (group.Label != "iss-Sun" && group.Label != "iss-Moon") || len(group.Rows) == 0 {
			continue
		}
		t.Run(group.Label+"/"+strconv.Itoa(index), func(t *testing.T) {
			transitCases++
			elements, observer := group.Elements, group.Observer
			zone := time.FixedZone("UTC+08:00", 8*60*60)
			start, end := group.Start.In(zone), group.End.In(zone)
			targets := []satellite.TransitTarget{group.Target}
			siteOptions := satellite.TransitOptions{Targets: targets}
			searchOptions := satellite.TransitSearchOptions{Targets: targets}
			nearbyOptions := satellite.NearbyTransitOptions{TransitSearchOptions: searchOptions, RadiusKM: 100}
			look, err := satellite.LookAt(elements, observer, start, satellite.ObserveOptions{})
			requireSatelliteResult(t, err)
			passes, state, err := satellite.Passes(elements, observer, start, end, satellite.PassOptions{})
			requireSatelliteResult(t, err)
			transits, err := satellite.Transits(elements, observer, start, end, siteOptions)
			requireSatelliteResult(t, err)
			events, err := satellite.TransitEvents(elements, start, end, searchOptions)
			requireSatelliteResult(t, err)
			nearby, err := satellite.NearbyTransits([]satellite.Elements{elements}, observer, start, end, nearbyOptions)
			requireSatelliteResult(t, err)
			if len(passes) == 0 || len(transits) == 0 || len(events) == 0 || len(nearby) == 0 || nearby[0].Site == nil {
				t.Fatalf("need nonempty results with a site event: passes=%d transits=%d events=%d nearby=%d", len(passes), len(transits), len(events), len(nearby))
			}
			shadow, err := satellite.TransitShadowAt(elements, group.Target, transits[0].Center, searchOptions)
			requireSatelliteResult(t, err)
			if !shadow.Valid {
				t.Fatal("need a valid shadow")
			}
			track, err := satellite.TransitFrames(elements, observer, transits[0], 0)
			requireSatelliteResult(t, err)
			if len(track.Frames) == 0 {
				t.Fatal("need nonempty frames")
			}
			coverageOptions := satellite.CoverageOptions{MinElevation: 10}
			coverageEnd := start.Add(time.Minute)
			coverage, err := satellite.CoverageBetween(elements, start, coverageEnd, 30*time.Second, coverageOptions)
			requireSatelliteResult(t, err)
			if len(coverage.Swept) == 0 || len(coverage.Instants) != 3 {
				t.Fatal("need a nonempty sweep and three output instants")
			}

			jdTT := basic.UTC2TT(basic.Date2JD(transits[0].Center.UTC()))
			defaultPhase := track.MoonIlluminatedFraction
			if group.Target == satellite.TransitMoon {
				assertSatelliteBits(t, "default MoonIlluminatedFraction", basic.MoonPhaseWithInstance(jdTT, nil), defaultPhase)
			}
			for _, path := range []struct {
				name   string
				module *satellite.Satellite
				inst   *ephem.Instance
			}{{"explicit", explicit, inst}, {"context", module, context.inst}} {
				t.Run(path.name, func(t *testing.T) {
					gotLook, err := path.module.LookAt(elements, observer, start, satellite.ObserveOptions{})
					requireSatelliteResult(t, err)
					assertSatelliteBits(t, "LookAt", look, gotLook)
					gotPasses, gotState, err := path.module.Passes(elements, observer, start, end, satellite.PassOptions{})
					requireSatelliteResult(t, err)
					assertSatelliteBits(t, "Passes", passes, gotPasses)
					assertSatelliteBits(t, "Passes state", state, gotState)
					gotTransits, err := path.module.Transits(elements, observer, start, end, siteOptions)
					requireSatelliteResult(t, err)
					assertSatelliteBits(t, "Transits", transits, gotTransits)
					gotEvents, err := path.module.TransitEvents(elements, start, end, searchOptions)
					requireSatelliteResult(t, err)
					assertSatelliteBits(t, "TransitEvents", events, gotEvents)
					gotNearby, err := path.module.NearbyTransits([]satellite.Elements{elements}, observer, start, end, nearbyOptions)
					requireSatelliteResult(t, err)
					assertSatelliteBits(t, "NearbyTransits", nearby, gotNearby)
					gotShadow, err := path.module.TransitShadowAt(elements, group.Target, gotTransits[0].Center, searchOptions)
					requireSatelliteResult(t, err)
					assertSatelliteBits(t, "TransitShadowAt", shadow, gotShadow)
					gotTrack, err := path.module.TransitFrames(elements, observer, gotTransits[0], 0)
					requireSatelliteResult(t, err)
					if group.Target == satellite.TransitMoon {
						phase := gotTrack.MoonIlluminatedFraction
						assertSatelliteBits(t, "MoonIlluminatedFraction", basic.MoonPhaseWithInstance(jdTT, path.inst), phase)
						difference := math.Abs(phase - defaultPhase)
						if !(difference <= 1e-7) {
							t.Fatalf("MoonIlluminatedFraction difference %.12g > 1e-7", difference)
						}
						maximumMoonDifference = math.Max(maximumMoonDifference, difference)
						gotTrack.MoonIlluminatedFraction = defaultPhase
					}
					assertSatelliteBits(t, "TransitFrames", track, gotTrack)
					gotCoverage, err := path.module.CoverageBetween(elements, start, coverageEnd, 30*time.Second, coverageOptions)
					requireSatelliteResult(t, err)
					assertSatelliteBits(t, "CoverageBetween", coverage, gotCoverage)
				})
			}
		})
		if transitCases == 2 {
			break
		}
	}
	if transitCases != 2 {
		t.Fatalf("need both solar and lunar fixtures, got %d cases", transitCases)
	}
	t.Logf("maximum MoonIlluminatedFraction difference = %.12g", maximumMoonDifference)

	var zero Context
	var absent *Context
	for _, empty := range []*Context{&zero, absent} {
		module := empty.Satellite()
		if module == nil || module != empty.Satellite() || module == context.Satellite() {
			t.Fatal("empty Context must cache a module without using the configured Context")
		}
		group := fixtures.Groups[0]
		if _, err := module.LookAt(group.Elements, group.Observer, group.Start, satellite.ObserveOptions{}); !errors.Is(err, satellite.ErrInvalidInput) {
			t.Fatalf("empty Context LookAt error = %v", err)
		}
	}
}

func requireSatelliteResult(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func assertSatelliteBits(t *testing.T, field string, want, got interface{}) {
	t.Helper()
	compareSatelliteBits(t, field, reflect.ValueOf(want), reflect.ValueOf(got))
}

func compareSatelliteBits(t *testing.T, field string, want, got reflect.Value) {
	t.Helper()
	if want.Type() != got.Type() {
		t.Fatalf("%s: types differ: %v and %v", field, want.Type(), got.Type())
	}
	switch want.Kind() {
	case reflect.Float64:
		if math.Float64bits(want.Float()) != math.Float64bits(got.Float()) {
			t.Fatalf("%s: got %.17g, want %.17g", field, got.Float(), want.Float())
		}
	case reflect.Ptr:
		if want.IsNil() != got.IsNil() {
			t.Fatalf("%s: one pointer is nil", field)
		}
		if !want.IsNil() {
			compareSatelliteBits(t, field, want.Elem(), got.Elem())
		}
	case reflect.Array, reflect.Slice:
		if want.Len() != got.Len() || want.Kind() == reflect.Slice && want.IsNil() != got.IsNil() {
			t.Fatalf("%s: sequence length or nilness differs", field)
		}
		for index := 0; index < want.Len(); index++ {
			compareSatelliteBits(t, field+"["+strconv.Itoa(index)+"]", want.Index(index), got.Index(index))
		}
	case reflect.Struct:
		if want.Type() == reflect.TypeOf(time.Time{}) {
			if want.Interface().(time.Time) != got.Interface().(time.Time) {
				t.Fatalf("%s: time or time zone differs: %v, %v", field, want, got)
			}
			return
		}
		for index := 0; index < want.NumField(); index++ {
			compareSatelliteBits(t, field+"."+want.Type().Field(index).Name, want.Field(index), got.Field(index))
		}
	default:
		if !reflect.DeepEqual(want.Interface(), got.Interface()) {
			t.Fatalf("%s: got %v, want %v", field, got, want)
		}
	}
}
