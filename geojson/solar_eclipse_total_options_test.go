package geojson_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
)

func TestSolarTotalBandIndependentOfMagnitudeLines(t *testing.T) {
	for _, fixture := range []struct {
		date     time.Time
		lon, lat float64
	}{
		{time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC), -5.991755201, 45.070085755},
		{time.Date(2024, 4, 8, 0, 0, 0, 0, time.UTC), -32.007627414, 49.493804579},
	} {
		t.Run(fixture.date.Format("2006-01-02"), func(t *testing.T) {
			central, ok := eclipse.SolarEclipseCentralPath(fixture.date, eclipse.SolarEclipsePathOptions{Step: time.Minute})
			if !ok {
				t.Fatal("missing central path")
			}
			local, ok := eclipse.GeometricLocalSolarEclipseOnDate(fixture.date, fixture.lon, fixture.lat, 0)
			if !ok || local.Type != eclipse.SolarEclipseTotal || !local.VisibleAtGreatest {
				t.Fatal("invalid totality witness")
			}
			var reference [][]eclipse.SolarEclipsePathPoint
			for _, magnitudes := range [][]float64{nil, {1}} {
				partial, ok := eclipse.SolarEclipsePartialFootprints(fixture.date, eclipse.SolarEclipsePartialFootprintOptions{
					Step: 2 * time.Minute, BoundaryPoints: 96, MagnitudeValues: magnitudes,
				})
				if !ok || len(partial.CentralBandSegments) == 0 {
					t.Fatal("missing authoritative totality band")
				}
				if reference == nil {
					reference = partial.CentralBandSegments
				} else if !reflect.DeepEqual(reference, partial.CentralBandSegments) {
					t.Fatal("magnitude display option changes totality geometry")
				}
				if len(magnitudes) == 0 && len(partial.MagnitudeContours) != 0 {
					t.Fatal("unrequested magnitude lines")
				}
				for _, ring := range partial.CentralBandSegments {
					assertSolarPathMaximumEdgeKM(t, ring, 250)
				}
				for _, path := range []*eclipse.SolarEclipsePath{&central, nil} {
					data, err := geojson.MarshalSolarEclipse(partial, path)
					if err != nil {
						t.Fatal(err)
					}
					band := featureWithRole(t, decodeCollection(t, data), "central-band")
					if !geometryContainsPoint(t, band.Geometry, fixture.lon, fixture.lat) {
						t.Fatal("totality band omits visible site")
					}
				}
			}
		})
	}
}
