package svg_test

import (
	"fmt"
	"strings"
	"time"

	"github.com/starainrt/astro/satellite"
	"github.com/starainrt/astro/satellite/svg"
)

func exampleISS() satellite.Elements {
	elements, err := satellite.ParseTLEText([]byte(`ISS (ZARYA)
1 25544U 98067A   26317.91612726  .00008668  00000+0  16320-3 0  9995
2 25544  51.6324 279.1461 0007381  22.6178 337.5134 15.49359724594966`))
	if err != nil {
		panic(err)
	}
	return elements[0]
}

func ExampleCoverageSVG() {
	iss := exampleISS()
	start := time.Date(2026, 11, 14, 1, 15, 0, 0, time.UTC)
	sweep, err := satellite.CoverageBetween(iss, start, start.Add(35*time.Minute), time.Minute, satellite.CoverageOptions{MinElevation: 10})
	if err != nil {
		panic(err)
	}
	image, err := svg.CoverageSVG(sweep, svg.CoverageSVGOptions{
		Projection: svg.MapProjectionOrthographic,
		Location:   time.UTC,
		Highlight:  start.Add(20 * time.Minute),
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(len(sweep.Instants), strings.Contains(image, "<svg"))
	// Output:
	// 36 true
}

func coverageStationExample() string {
	iss := exampleISS()
	start := time.Date(2026, 11, 14, 18, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	station := satellite.Observer{Latitude: 40}
	sweep, err := satellite.CoverageBetween(iss, start, end, 10*time.Second, satellite.CoverageOptions{MinElevation: 10})
	if err != nil {
		panic(err)
	}
	standardMagnitude := 0.0
	lookOptions := satellite.ObserveOptions{NoRefraction: true, StandardMagnitude: &standardMagnitude}
	passes, _, err := satellite.Passes(iss, station, start, end, satellite.PassOptions{ObserveOptions: lookOptions, MinElevation: 10})
	if err != nil {
		panic(err)
	}
	details := make([]svg.CoveragePass, len(passes))
	for i, pass := range passes {
		details[i].Pass = pass
		details[i].Samples = append(details[i].Samples, pass.Rise)
		for at := pass.Rise.Time.Add(15 * time.Second); at.Before(pass.Set.Time); at = at.Add(15 * time.Second) {
			look, err := satellite.LookAt(iss, station, at, lookOptions)
			if err != nil {
				panic(err)
			}
			details[i].Samples = append(details[i].Samples, look)
		}
		if pass.Set.Time.After(pass.Rise.Time) {
			details[i].Samples = append(details[i].Samples, pass.Set)
		}
	}
	image, err := svg.CoverageSVG(sweep, svg.CoverageSVGOptions{Projection: svg.MapProjectionOrthographic, Observer: &station, Passes: details, Highlight: start.Add(23 * time.Minute)})
	if err != nil {
		panic(err)
	}
	return image
}

func ExampleCoverageSVG_station() {
	image := coverageStationExample()
	fmt.Println(strings.Count(image, `data-role="coverage-pass"`), strings.Contains(image, `data-role="coverage-magnitude-curve"`))
	// Output:
	// 1 true
}

func ExampleTransitMapSVG() {
	iss := exampleISS()
	start := time.Date(2026, 11, 14, 0, 40, 0, 0, time.UTC)
	options := satellite.TransitSearchOptions{Targets: []satellite.TransitTarget{satellite.TransitMoon}}
	events, err := satellite.TransitEvents(iss, start, start.Add(40*time.Minute), options)
	if err != nil {
		panic(err)
	}
	path, err := satellite.TransitEventPath(iss, events[0], time.Second)
	if err != nil {
		panic(err)
	}
	image, err := svg.TransitMapSVG(path, svg.TransitMapSVGOptions{Projection: svg.MapProjectionOrthographic, Location: time.UTC})
	if err != nil {
		panic(err)
	}
	fmt.Println(events[0].Complete, strings.Contains(image, "<svg"))
	// Output:
	// true true
}

func ExampleTransitLocalSVG() {
	iss := exampleISS()
	station := satellite.Observer{Longitude: 162.43010446888346, Latitude: -42.00800539858516}
	start := time.Date(2026, 11, 14, 0, 48, 0, 0, time.UTC)
	options := satellite.TransitOptions{Targets: []satellite.TransitTarget{satellite.TransitMoon}}
	transits, err := satellite.Transits(iss, station, start, start.Add(4*time.Minute), options)
	if err != nil {
		panic(err)
	}
	track, err := satellite.TransitFrames(iss, station, transits[0], 0)
	if err != nil {
		panic(err)
	}
	image, err := svg.TransitLocalSVG(track, svg.TransitLocalSVGOptions{Location: time.UTC})
	if err != nil {
		panic(err)
	}
	fmt.Println(strings.Contains(image, "<svg"))
	// Output:
	// true
}
