package geojson_test

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
)

func ExampleMarshalSolarEclipse() {
	date := time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipse.SolarEclipsePartialFootprints(
		date,
		eclipse.SolarEclipsePartialFootprintOptions{
			Step:           10 * time.Minute,
			BoundaryPoints: 180,
		},
	)
	if !ok {
		return
	}
	central, hasCentral := eclipse.SolarEclipseCentralPath(
		date,
		eclipse.SolarEclipsePathOptions{
			Step:            time.Minute,
			TargetSpacingKM: 20,
		},
	)
	var centralPath *eclipse.SolarEclipsePath
	if hasCentral {
		centralPath = &central
	}

	data, err := geojson.MarshalSolarEclipse(partial, centralPath)
	fmt.Println(err == nil, json.Valid(data))
	// Output: true true
}

func ExampleMarshalSolarEclipseWithTimeMarkers() {
	date := time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipse.SolarEclipsePartialFootprints(
		date,
		eclipse.SolarEclipsePartialFootprintOptions{Step: 20 * time.Minute, BoundaryPoints: 36},
	)
	if !ok {
		return
	}
	central, ok := eclipse.SolarEclipseCentralPath(
		date,
		eclipse.SolarEclipsePathOptions{Step: 5 * time.Minute},
	)
	if !ok {
		return
	}
	data, err := geojson.MarshalSolarEclipseWithTimeMarkers(
		partial,
		&central,
		geojson.TimeMarkerOptions{Step: 30 * time.Minute, Location: time.FixedZone("CST", 8*60*60)},
	)
	fmt.Println(err == nil, json.Valid(data))
	// Output: true true
}
