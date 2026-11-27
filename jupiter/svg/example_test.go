package svg_test

import (
	"fmt"
	"strings"
	"time"

	"github.com/starainrt/astro/jupiter"
	"github.com/starainrt/astro/jupiter/svg"
)

func ExampleGalileanSystemSVG() {
	date := time.Date(2026, 6, 4, 13, 40, 0, 0, time.UTC)
	document, ok := svg.GalileanSystemSVG(date, svg.GalileanSystemSVGOptions{
		Language:  "en",
		Location:  time.UTC,
		TrueScale: true,
	})
	fmt.Println(ok, strings.Contains(document, "<svg"))
	// Output:
	// true true
}

func ExampleGalileanMutualEventSVG() {
	date := time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC)
	event := jupiter.NextGalileanMutualEvent(date, jupiter.GalileanSatelliteIo, jupiter.GalileanSatelliteEuropa, jupiter.GalileanMutualEclipse)
	document, ok := svg.GalileanMutualEventSVG(event, svg.GalileanMutualEventSVGOptions{
		Language: "en",
		Location: time.UTC,
	})
	fmt.Println(event.Valid, ok, strings.Contains(document, "<svg"))
	// Output:
	// true true true
}

func ExampleGalileanPhenomenonSVG() {
	date := time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC)
	event := jupiter.NextGalileanPhenomenonContactEvent(date, jupiter.GalileanSatelliteIo, jupiter.GalileanPhenomenonEclipse)
	document, ok := svg.GalileanPhenomenonSVG(event, svg.GalileanPhenomenonSVGOptions{})
	fmt.Println(event.Valid, ok, strings.Contains(document, "<svg"))
	// Output:
	// true true true
}
