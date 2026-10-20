package kml

import (
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/venus"
)

func TestPlanetTransitGlobalPartialConversion(tester *testing.T) {
	info := venus.NextTransit(time.Date(2854, 1, 1, 0, 0, 0, 0, time.UTC))
	data, err := geojson.MarshalVenusTransit(info, geojson.PlanetTransitOptions{IsochroneStep: time.Minute})
	if err != nil {
		tester.Fatal(err)
	}
	root := convert(tester, string(data), Options{})
	if !strings.Contains(root.Document.Name, "2854-12-14") {
		tester.Fatal(root.Document.Name)
	}
	if root.Document.LookAt == nil || root.Document.LookAt.Latitude > -50 {
		tester.Fatal("KML must focus on the actual partial-transit footprint")
	}
	roles := map[string]bool{}
	for _, item := range root.Document.Folders {
		roles[item.Name] = true
	}
	if !roles["部分凌日可见区"] || !roles["一触等时线"] || !roles["四触等时线"] || roles["二触等时线"] || roles["一触视地平线"] {
		tester.Fatalf("invalid partial-transit layers: %v", roles)
	}
	if paletteStyle(eventPlanetTransit, "transit-partial").FillColor == "" {
		tester.Fatal("missing partial-transit fill")
	}
}
