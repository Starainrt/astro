package basic

import (
	"github.com/starainrt/astro/internal/ephem"
	"testing"
)

func BenchmarkSurfacePositionSun(b *testing.B)   { surfaceBenchmark(b, "position-sun") }
func BenchmarkSurfacePositionEarth(b *testing.B) { surfaceBenchmark(b, "position-earth") }
func BenchmarkSurfaceMarsNextRise(b *testing.B)  { surfaceBenchmark(b, "mars-next") }
func BenchmarkSurfaceJupiterYear(b *testing.B)   { surfaceBenchmark(b, "jupiter-year") }
func BenchmarkSurfaceMercuryYear(b *testing.B)   { surfaceBenchmark(b, "mercury-year") }
func BenchmarkSurfaceUranusDecade(b *testing.B)  { surfaceBenchmark(b, "uranus-decade") }

func surfaceBenchmark(b *testing.B, name string) {
	inst := ephem.NewModelInstance(ephem.Semianalytic)
	observer := BodyObserver{Body: SkyMars, EastLongitude: 30, Latitude: 25}
	target := SkySun
	duration := 365.0
	switch name {
	case "position-earth":
		target = SkyEarth
	case "jupiter-year":
		observer.Body = SkyJupiter
	case "mercury-year":
		observer.Body = SkyMercury
		observer.Latitude = 0
	case "uranus-decade":
		observer.Body = SkyUranus
		duration = 10 * 365.25
	}
	if _, err := SurfacePositionAtWithInstance(target, observer, 2461041.5, inst); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		switch name {
		case "position-sun", "position-earth":
			if _, err := SurfacePositionAtWithInstance(target, observer, 2461041.5, inst); err != nil {
				b.Fatal(err)
			}
		case "mars-next":
			if _, err := NextSurfaceEventWithInstance(SkySun, observer, SurfaceRise, 2461041.5, SurfaceOptions{}, inst); err != nil {
				b.Fatal(err)
			}
		default:
			if _, _, err := SurfaceEventsWithInstance(SkySun, observer, 2461041.5, 2461041.5+duration, SurfaceOptions{}, inst); err != nil {
				b.Fatal(err)
			}
		}
	}
}
