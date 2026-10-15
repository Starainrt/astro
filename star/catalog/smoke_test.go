package catalog

import (
	"math"
	"os"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

func TestCatalogCoordinatesFeedSearches(t *testing.T) {
	bsc := readFixtureCatalog(t, "bsc5_sample.dat", func(f *os.File) (*Catalog, error) {
		return ReadBSC5(f, ReadOptions{})
	})
	target, ok := bsc.Lookup("HR 4799")
	if !ok {
		t.Fatal("HR 4799 missing")
	}
	location := time.FixedZone("CST", 8*3600)
	best, err := basic.FindBestStarOccultations(
		time.Date(2025, 6, 5, 0, 0, 0, 0, location),
		time.Date(2025, 6, 6, 0, 0, 0, 0, location),
		target.Coordinate, basic.OccultationSearchOptions{MaxEvents: 1})
	if err != nil || len(best) != 1 {
		t.Fatalf("occultation = %d events, %v", len(best), err)
	}
	if d := best[0].Greatest.Sub(time.Date(2025, 6, 5, 20, 2, 7, 700000000, location)).Seconds(); math.Abs(d) > 5 {
		t.Fatalf("greatest off by %.1f s", d)
	}

	sirius, ok := bsc.Lookup("HR 2491")
	if !ok {
		t.Fatal("HR 2491 missing")
	}
	startJD := basic.Date2JD(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	events := basic.FindStarHeliacalEventsWithInstance(sirius.Coordinate, sirius.Magnitude, startJD, startJD+366, 31.2, 30.0, basic.HeliacalOptions{}, nil)
	if len(events) == 0 {
		t.Fatal("no heliacal events for Sirius at Cairo")
	}
}
