package geojson_test

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/eclipse"
	geojson "github.com/starainrt/astro/geojson"
)

func TestMarshalSolarEclipse43290612UsesUnifiedPolarCentralBand(t *testing.T) {
	for _, sample := range []struct{ year, month, day int }{{4329, 6, 12}, {4005, 4, 22}} {
		t.Run(time.Date(sample.year, time.Month(sample.month), sample.day, 0, 0, 0, 0, time.UTC).Format("2006-01-02"), func(t *testing.T) {
			date := time.Date(sample.year, time.Month(sample.month), sample.day, 0, 0, 0, 0, time.UTC)
			partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{Step: 2 * time.Minute, BoundaryPoints: 96})
			if !ok || len(partial.CentralBandSegments) != 1 {
				t.Fatalf("partial central envelope unavailable: ok=%v segments=%d", ok, len(partial.CentralBandSegments))
			}
			central, ok := eclipse.SolarEclipseCentralPath(date, eclipse.SolarEclipsePathOptions{Step: 2 * time.Minute, TargetSpacingKM: 700})
			if !ok || len(central.CentralBandSegments) != 1 {
				t.Fatalf("central path envelope unavailable: ok=%v segments=%d", ok, len(central.CentralBandSegments))
			}
			if got, want := len(central.CentralBandSegments[0]), len(partial.CentralBandSegments[0]); got != want {
				t.Fatalf("path/partial envelope point count=%d/%d", got, want)
			}
			for _, index := range []int{0, len(partial.CentralBandSegments[0]) / 2} {
				got, want := central.CentralBandSegments[0][index], partial.CentralBandSegments[0][index]
				if math.Abs(got.Longitude-want.Longitude) > 1e-9 || math.Abs(got.Latitude-want.Latitude) > 1e-9 {
					t.Fatalf("path/partial envelope diverges at %d: got=(%.9f,%.9f) want=(%.9f,%.9f)", index, got.Longitude, got.Latitude, want.Longitude, want.Latitude)
				}
			}
			data, err := geojson.MarshalSolarEclipse(partial, &central)
			if err != nil {
				t.Fatalf("MarshalSolarEclipse: %v", err)
			}
			band := featureWithRole(t, decodeCollection(t, data), "central-band")
			if source, ok := band.Properties["source"].(string); !ok || source == "" {
				t.Fatalf("central-band source=%v, want a physical envelope source", band.Properties["source"])
			}
		})
	}
}
