package basic

import (
	"math"
	"testing"
)

// Central durations at the greatest eclipse point, from NASA's five millennium
// catalogue ("Central Dur." column). The library publishes the same figure on
// SolarEclipseResult.CentralDurationDays.
var solarEclipseCatalogueCentralDurations = []struct {
	date         [3]int
	seconds      float64
	toleranceSec float64
}{
	{[3]int{1991, 7, 11}, 6*60 + 53, 1.5},
	{[3]int{2009, 7, 22}, 6*60 + 39, 1.5},
	{[3]int{2024, 4, 8}, 4*60 + 28, 1.5},
	{[3]int{2017, 8, 21}, 2*60 + 40, 1.5},
	{[3]int{2020, 6, 21}, 38, 1.5},
	{[3]int{1136, 6, 1}, 3*60 + 18, 1.5},
}

func TestSolarEclipseCentralDurationMatchesCatalogue(t *testing.T) {
	for _, item := range solarEclipseCatalogueCentralDurations {
		seed := JDCalc(item.date[0], item.date[1], float64(item.date[2]))
		result := SolarEclipseNASABulletinSplitK(seed)
		if !result.HasCentral {
			t.Fatalf("%04d-%02d-%02d is not a central eclipse", item.date[0], item.date[1], item.date[2])
		}
		seconds := result.CentralDurationDays * 86400
		if math.Abs(seconds-item.seconds) > item.toleranceSec {
			t.Fatalf("%04d-%02d-%02d central duration = %.1f s, catalogue %.0f s",
				item.date[0], item.date[1], item.date[2], seconds, item.seconds)
		}
	}
}

// The path width must stay a footprint figure: never wider than the Earth, and
// zero wherever no paired cross-section exists. A one-limit event has no pair at
// all, which is why the catalogues print no width for it (1874-10-10 is listed
// with "-"), and the analytic 2r/sin(altitude) form diverges towards the horizon
// where the pair is the only meaningful measurement.
func TestSolarEclipsePathWidthStaysPhysical(t *testing.T) {
	cases := []struct {
		date     [3]int
		oneLimit bool
	}{
		{[3]int{2009, 7, 22}, false},
		{[3]int{2017, 8, 21}, false},
		{[3]int{2024, 4, 8}, false},
		{[3]int{2020, 6, 21}, false},
		{[3]int{1136, 6, 1}, false},
		{[3]int{1874, 10, 10}, true},
	}
	for _, item := range cases {
		seed := JDCalc(item.date[0], item.date[1], float64(item.date[2]))
		path := SolarEclipseCentralPathNASABulletinSplitK(seed, SolarEclipsePathOptions{})
		if len(path.CenterLine) == 0 {
			t.Fatalf("%04d-%02d-%02d exported no center line", item.date[0], item.date[1], item.date[2])
		}
		maximum := 0.0
		for _, point := range path.CenterLine {
			if point.WidthKM < 0 || point.WidthKM > solarEclipsePathMaxPossibleWidthKM {
				t.Fatalf("%04d-%02d-%02d width %.1f km at (%.3f, %.3f) is not a footprint width",
					item.date[0], item.date[1], item.date[2], point.WidthKM, point.Longitude, point.Latitude)
			}
			maximum = math.Max(maximum, point.WidthKM)
		}
		if item.oneLimit && maximum != 0 {
			t.Fatalf("%04d-%02d-%02d is a one-limit event but reports a %.1f km width",
				item.date[0], item.date[1], item.date[2], maximum)
		}
		if !item.oneLimit && maximum <= 0 {
			t.Fatalf("%04d-%02d-%02d is a two-limit event but reports no width",
				item.date[0], item.date[1], item.date[2])
		}
	}
}
