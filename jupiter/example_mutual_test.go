package jupiter_test

import (
	"fmt"
	"time"

	"github.com/starainrt/astro"
	"github.com/starainrt/astro/jupiter"
)

func ExampleNextGalileanMutualEvent() {
	date := time.Date(2026, 6, 4, 0, 0, 0, 0, time.UTC)
	event := jupiter.NextGalileanMutualEvent(date, jupiter.GalileanSatelliteIo, jupiter.GalileanSatelliteEuropa, jupiter.GalileanMutualOccultation)
	if !event.Valid {
		fmt.Println("no event")
		return
	}
	fmt.Println(event.Type, event.Active, event.Passive)
	fmt.Println(event.Start.Before(event.Greatest), event.Greatest.Before(event.End))
	fmt.Println(event.Duration == event.End.Sub(event.Start))
	// Output:
	// mutual_occultation 1 2
	// true true
	// true
}

func ExampleNextGalileanMutualEvent_eclipse() {
	date := time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC)
	event := jupiter.NextGalileanMutualEvent(date, jupiter.GalileanSatelliteIo, jupiter.GalileanSatelliteEuropa, jupiter.GalileanMutualEclipse)
	if !event.Valid {
		fmt.Println("no event")
		return
	}
	fmt.Println(event.Type, event.Active, event.Passive)
	fmt.Println(event.ImpactParameter < 1, event.Coverage >= 0 && event.Coverage <= 1)
	passive := event.GreatestState[event.Passive-1]
	fmt.Println(passive.Occultation, passive.Eclipse)
	// Output:
	// mutual_eclipse 1 2
	// true true
	// false false
}

func ExampleJupiter_NextGalileanMutualEvent() {
	context, err := astro.NewContext(astro.ModelSemianalytic)
	if err != nil {
		fmt.Println(err)
		return
	}
	date := time.Date(2026, 6, 4, 0, 0, 0, 0, time.UTC)
	event := context.Jupiter().NextGalileanMutualEvent(date, jupiter.GalileanSatelliteEuropa, jupiter.GalileanSatelliteIo, jupiter.GalileanMutualOccultation)
	fmt.Println(event.Valid, event.Active, event.Passive)
	// Output:
	// true 1 2
}

func ExampleGalileanEventsBetween() {
	start := time.Date(2026, 6, 4, 0, 0, 0, 0, time.UTC)
	options := jupiter.GalileanEventsOptions{
		Satellites:     [4]bool{true, true},
		IncludeContact: true,
		IncludeMutual:  true,
	}
	events, err := jupiter.GalileanEventsBetween(start, start.Add(24*time.Hour), options)
	if err != nil {
		fmt.Println(err)
		return
	}
	counts := make(map[jupiter.GalileanEventFamily]int)
	for _, event := range events {
		counts[event.Family]++
		if event.Family == jupiter.GalileanFamilyMutual {
			fmt.Println(event.Mutual.Type, event.Mutual.Active, event.Mutual.Passive)
		}
	}
	fmt.Println(counts[jupiter.GalileanFamilyPhenomenon] > 0, counts[jupiter.GalileanFamilyContact] > 0)
	// Output:
	// mutual_occultation 1 2
	// true true
}

func Example_earthUTCDate() {
	zone := time.FixedZone("UTC+14", 14*3600)
	instant := time.Date(2026, 6, 4, 1, 30, 0, 0, zone)
	fmt.Println(astro.EarthUTCDate(instant).Format(time.RFC3339))
	fmt.Println(astro.EarthUTCDate(time.Time{}).IsZero())
	// Output:
	// 2026-06-03T00:00:00Z
	// true
}
